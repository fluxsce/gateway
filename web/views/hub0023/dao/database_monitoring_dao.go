package dao

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0023/models"
)

// DatabaseMonitoringDAO 关系数据库监控数据访问对象
// 专门用于从 HUB_GW_ACCESS_LOG 表中抽取各种监控统计数据
//
// 重要提示：
// 1. 为了优化查询性能，建议在数据库中创建以下索引：
//   - INDEX idx_gateway_log_time_tenant (gatewayStartProcessingTime, tenantId)
//   - INDEX idx_gateway_log_status (gatewayStatusCode)
//   - INDEX idx_gateway_log_route (routeConfigId, requestPath)
//
// 2. 所有聚合查询都经过优化，使用数据库的聚合函数
// 3. 查询时间范围已在控制器层限制为24小时内，防止大数据量查询
// 4. 兼容多种关系数据库(MySQL、PostgreSQL、SQLite等)
type DatabaseMonitoringDAO struct {
	db database.Database
}

// NewDatabaseMonitoringDAO 创建关系数据库监控数据DAO
func NewDatabaseMonitoringDAO(db database.Database) *DatabaseMonitoringDAO {
	return &DatabaseMonitoringDAO{
		db: db,
	}
}

// GetGatewayMonitoringOverview 获取网关监控概览数据
// 基于查询条件统计总体监控指标
func (dao *DatabaseMonitoringDAO) GetGatewayMonitoringOverview(ctx context.Context, req *models.GatewayMonitoringQueryRequest) (*models.GatewayMonitoringOverview, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 构建概览查询SQL - 使用数据库的聚合函数一次性计算所有指标
	sql := fmt.Sprintf(`
		SELECT 
			COUNT(*) as totalRequests,
			SUM(CASE WHEN gatewayStatusCode >= 200 AND gatewayStatusCode < 300 THEN 1 ELSE 0 END) as successRequests,
			SUM(CASE WHEN gatewayStatusCode >= 400 OR gatewayStatusCode < 200 THEN 1 ELSE 0 END) as failedRequests,
			AVG(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as avgResponseTime,
			MIN(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as minResponseTime,
			MAX(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as maxResponseTime
		FROM HUB_GW_ACCESS_LOG
		%s
	`, whereClause)

	// 执行查询
	// 使用sql.NullFloat64和sql.NullInt64来处理可能的NULL值
	var result struct {
		TotalRequests   int64   `db:"totalRequests"`
		SuccessRequests int64   `db:"successRequests"`
		FailedRequests  int64   `db:"failedRequests"`
		AvgResponseTime float64 `db:"avgResponseTime"`
		MinResponseTime int64   `db:"minResponseTime"`
		MaxResponseTime int64   `db:"maxResponseTime"`
	}

	err = dao.db.QueryOne(ctx, &result, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库监控概览查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库监控概览查询失败")
	}

	// 构建响应对象
	overview := &models.GatewayMonitoringOverview{
		TotalRequests:     result.TotalRequests,
		SuccessRequests:   result.SuccessRequests,
		FailedRequests:    result.FailedRequests,
		AvgResponseTimeMs: roundToTwoDecimalPlaces(result.AvgResponseTime),
		MinResponseTimeMs: int(result.MinResponseTime),
		MaxResponseTimeMs: int(result.MaxResponseTime),
	}

	return overview, nil
}

// GetRequestMetricsTrend 获取请求指标趋势数据
// 按时间粒度分组统计请求量数据
func (dao *DatabaseMonitoringDAO) GetRequestMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.RequestMetrics, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取时间分组格式
	timeFormat := dao.getTimeGroupFormat(req.TimeGranularity)
	granularitySeconds := dao.getTimeGranularitySeconds(req.TimeGranularity)

	// 构建趋势查询SQL - 使用数据库的时间格式化函数
	// 注意：这里使用兼容性较好的SQL语法
	sql := fmt.Sprintf(`
		SELECT 
			%s as timeGroup,
			COUNT(*) as totalRequests,
			SUM(CASE WHEN gatewayStatusCode >= 200 AND gatewayStatusCode < 300 THEN 1 ELSE 0 END) as successRequests,
			SUM(CASE WHEN gatewayStatusCode >= 400 OR gatewayStatusCode < 200 THEN 1 ELSE 0 END) as failedRequests,
			MIN(gatewayStartProcessingTime) as minTimestamp
		FROM HUB_GW_ACCESS_LOG
		%s
		GROUP BY timeGroup
		ORDER BY minTimestamp
	`, timeFormat, whereClause)

	// 执行查询
	var results []struct {
		TimeGroup       string    `db:"timeGroup"`
		TotalRequests   int64     `db:"totalRequests"`
		SuccessRequests int64     `db:"successRequests"`
		FailedRequests  int64     `db:"failedRequests"`
		MinTimestamp    time.Time `db:"minTimestamp"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库请求趋势查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库请求趋势查询失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.RequestMetrics{}, nil
	}

	// 转换为响应格式
	metrics := make([]models.RequestMetrics, 0, len(results))
	for _, result := range results {
		qps := float64(result.TotalRequests) / float64(granularitySeconds)
		metrics = append(metrics, models.RequestMetrics{
			Timestamp:         result.MinTimestamp.UnixMilli(),
			TotalRequests:     result.TotalRequests,
			SuccessRequests:   result.SuccessRequests,
			FailedRequests:    result.FailedRequests,
			RequestsPerSecond: qps,
		})
	}

	return metrics, nil
}

// GetResponseTimeMetricsTrend 获取响应时间指标趋势数据
// 按时间粒度分组统计响应时间数据
func (dao *DatabaseMonitoringDAO) GetResponseTimeMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.ResponseTimeMetrics, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取时间分组格式
	timeFormat := dao.getTimeGroupFormat(req.TimeGranularity)

	// 构建响应时间趋势查询SQL
	// 注意：由于标准SQL不支持百分位数聚合函数，这里需要在应用层计算
	sql := fmt.Sprintf(`
		SELECT 
			%s as timeGroup,
			COUNT(*) as requestCount,
			AVG(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as avgResponseTime,
			MIN(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as minResponseTime,
			MAX(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as maxResponseTime,
			MIN(gatewayStartProcessingTime) as minTimestamp
		FROM HUB_GW_ACCESS_LOG
		%s AND totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0
		GROUP BY timeGroup
		ORDER BY minTimestamp
	`, timeFormat, whereClause)

	// 执行查询
	var results []struct {
		TimeGroup       string    `db:"timeGroup"`
		RequestCount    int64     `db:"requestCount"`
		AvgResponseTime float64   `db:"avgResponseTime"`
		MinResponseTime int64     `db:"minResponseTime"`
		MaxResponseTime int64     `db:"maxResponseTime"`
		MinTimestamp    time.Time `db:"minTimestamp"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库响应时间趋势查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库响应时间趋势查询失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.ResponseTimeMetrics{}, nil
	}

	// 获取每个时间段的详细响应时间数据用于计算百分位数
	// 这是一个补充查询，获取原始数据用于计算p50、p90、p99
	detailedSQL := fmt.Sprintf(`
		SELECT 
			%s as timeGroup,
			totalProcessingTimeMs
		FROM HUB_GW_ACCESS_LOG
		%s AND totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0
		ORDER BY timeGroup, totalProcessingTimeMs
	`, timeFormat, whereClause)

	var detailedResults []struct {
		TimeGroup           string `db:"timeGroup"`
		TotalProcessingTime int    `db:"totalProcessingTimeMs"`
	}

	err = dao.db.Query(ctx, &detailedResults, detailedSQL, params, true)
	if err != nil {
		logger.WarnWithTrace(ctx, "数据库响应时间详细数据查询失败，将不包含百分位数", "error", err)
		// 即使详细数据查询失败，也继续返回基本的响应时间趋势
	}

	// 按时间分组组织响应时间数据
	timeGroupedData := make(map[string][]int)
	for _, detail := range detailedResults {
		timeGroupedData[detail.TimeGroup] = append(timeGroupedData[detail.TimeGroup], detail.TotalProcessingTime)
	}

	// 转换为响应格式
	metrics := make([]models.ResponseTimeMetrics, 0, len(results))
	for _, result := range results {
		// 计算百分位数
		p50, p90, p99 := 0, 0, 0
		if responseTimes, exists := timeGroupedData[result.TimeGroup]; exists && len(responseTimes) > 0 {
			sort.Ints(responseTimes)
			p50 = calculatePercentile(responseTimes, 0.5)
			p90 = calculatePercentile(responseTimes, 0.9)
			p99 = calculatePercentile(responseTimes, 0.99)
		}

		metrics = append(metrics, models.ResponseTimeMetrics{
			Timestamp:         result.MinTimestamp.UnixMilli(),
			RequestCount:      result.RequestCount,
			AvgResponseTimeMs: roundToTwoDecimalPlaces(result.AvgResponseTime),
			MinResponseTimeMs: int(result.MinResponseTime),
			MaxResponseTimeMs: int(result.MaxResponseTime),
			P50ResponseTimeMs: p50,
			P90ResponseTimeMs: p90,
			P99ResponseTimeMs: p99,
		})
	}

	return metrics, nil
}

// GetStatusCodeDistribution 获取状态码分布数据
func (dao *DatabaseMonitoringDAO) GetStatusCodeDistribution(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringStatusCodeData, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 首先获取总数
	totalCountSQL := fmt.Sprintf(`
		SELECT COUNT(*) as total
		FROM HUB_GW_ACCESS_LOG
		%s
	`, whereClause)

	var totalResult struct {
		Total int64 `db:"total"`
	}
	err = dao.db.QueryOne(ctx, &totalResult, totalCountSQL, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库状态码分布总数查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库状态码分布总数查询失败")
	}

	if totalResult.Total == 0 {
		return []models.GatewayMonitoringStatusCodeData{}, nil
	}

	// 构建状态码分布查询SQL
	sql := fmt.Sprintf(`
		SELECT 
			CAST(gatewayStatusCode AS CHAR) as statusCode,
			COUNT(*) as count
		FROM HUB_GW_ACCESS_LOG
		%s
		GROUP BY gatewayStatusCode
		ORDER BY count DESC
	`, whereClause)

	// 执行查询
	var results []struct {
		StatusCode string `db:"statusCode"`
		Count      int64  `db:"count"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库状态码分布查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库状态码分布查询失败")
	}

	// 转换为响应格式
	distribution := make([]models.GatewayMonitoringStatusCodeData, 0, len(results))
	for _, result := range results {
		percentage := float64(result.Count) / float64(totalResult.Total) * 100
		category := dao.getStatusCodeCategory(result.StatusCode)
		description := dao.getStatusCodeDescription(result.StatusCode)

		distribution = append(distribution, models.GatewayMonitoringStatusCodeData{
			StatusCode:  result.StatusCode,
			Count:       result.Count,
			Percentage:  percentage,
			Category:    category,
			Description: description,
		})
	}

	return distribution, nil
}

// GetHotRoutes 获取热点路由数据
func (dao *DatabaseMonitoringDAO) GetHotRoutes(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringHotRouteData, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 设置合理的默认限制，防止返回过多数据
	limit := req.HotRouteLimit
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 { // 限制最大值，防止大数据查询
		limit = 50
		logger.WarnWithTrace(ctx, "热点路由查询数量被限制", "requestedLimit", req.HotRouteLimit, "actualLimit", limit)
	}

	// 计算查询时间范围的秒数用于QPS计算
	var timeRangeSeconds float64
	startTime, _ := dao.parseTimeString(req.StartTime)
	endTime, _ := dao.parseTimeString(req.EndTime)
	if !startTime.IsZero() && !endTime.IsZero() {
		timeRangeSeconds = endTime.Sub(startTime).Seconds()
	}
	if timeRangeSeconds <= 0 {
		timeRangeSeconds = 1
	}

	// 构建热点路由查询SQL
	sql := fmt.Sprintf(`
		SELECT 
			requestPath as routePath,
			routeConfigId,
			routeName,
			serviceName,
			COUNT(*) as requestCount,
			SUM(CASE WHEN gatewayStatusCode >= 400 OR gatewayStatusCode < 200 THEN 1 ELSE 0 END) as errorCount,
			AVG(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as avgResponseTime,
			MIN(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as minResponseTime,
			MAX(CASE WHEN totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0 THEN totalProcessingTimeMs ELSE NULL END) as maxResponseTime
		FROM HUB_GW_ACCESS_LOG
		%s
		GROUP BY requestPath, routeConfigId, routeName, serviceName
		ORDER BY requestCount DESC
		LIMIT %d
	`, whereClause, limit)

	// 执行查询
	var results []struct {
		RoutePath       string  `db:"routePath"`
		RouteConfigId   string  `db:"routeConfigId"`
		RouteName       string  `db:"routeName"`
		ServiceName     string  `db:"serviceName"`
		RequestCount    int64   `db:"requestCount"`
		ErrorCount      int64   `db:"errorCount"`
		AvgResponseTime float64 `db:"avgResponseTime"`
		MinResponseTime int64   `db:"minResponseTime"`
		MaxResponseTime int64   `db:"maxResponseTime"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库热点路由查询失败", "error", err)
		return nil, huberrors.WrapError(err, "数据库热点路由查询失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.GatewayMonitoringHotRouteData{}, nil
	}

	// 转换为响应格式
	hotRoutes := make([]models.GatewayMonitoringHotRouteData, 0, len(results))
	for _, result := range results {
		// 计算错误率
		errorRate := float64(0)
		if result.RequestCount > 0 {
			errorRate = float64(result.ErrorCount) / float64(result.RequestCount) * 100
		}

		// 计算QPS
		qps := float64(result.RequestCount) / timeRangeSeconds

		hotRoutes = append(hotRoutes, models.GatewayMonitoringHotRouteData{
			RoutePath:         result.RoutePath,
			RouteConfigId:     result.RouteConfigId,
			RouteName:         result.RouteName,
			ServiceName:       result.ServiceName,
			RequestCount:      result.RequestCount,
			ErrorRate:         errorRate,
			QPS:               qps,
			MaxResponseTimeMs: int(result.MaxResponseTime),
			MinResponseTimeMs: int(result.MinResponseTime),
		})
	}

	return hotRoutes, nil
}

// buildMonitoringFilter 构建监控查询条件
func (dao *DatabaseMonitoringDAO) buildMonitoringFilter(req *models.GatewayMonitoringQueryRequest) (string, []interface{}, error) {
	whereClause := "WHERE activeFlag = 'Y'"
	var params []interface{}

	// 时间范围查询（必须字段，优先设置以利用时间索引）
	if req.StartTime != "" {
		startTime, err := dao.parseTimeString(req.StartTime)
		if err != nil {
			return "", nil, huberrors.WrapError(err, "开始时间格式错误")
		}
		whereClause += " AND gatewayStartProcessingTime >= ?"
		params = append(params, startTime)
	}

	if req.EndTime != "" {
		endTime, err := dao.parseTimeString(req.EndTime)
		if err != nil {
			return "", nil, huberrors.WrapError(err, "结束时间格式错误")
		}
		whereClause += " AND gatewayStartProcessingTime <= ?"
		params = append(params, endTime)
	}

	// 基础查询条件（精确匹配，利于索引）
	if req.TenantId != "" {
		whereClause += " AND tenantId = ?"
		params = append(params, req.TenantId)
	}
	if req.GatewayInstanceId != "" {
		whereClause += " AND gatewayInstanceId = ?"
		params = append(params, req.GatewayInstanceId)
	}
	if req.RouteConfigId != "" {
		whereClause += " AND routeConfigId = ?"
		params = append(params, req.RouteConfigId)
	}
	if req.RouteName != "" {
		whereClause += " AND routeName = ?"
		params = append(params, req.RouteName)
	}
	if req.ServiceDefinitionId != "" {
		whereClause += " AND serviceDefinitionId = ?"
		params = append(params, req.ServiceDefinitionId)
	}
	if req.ServiceName != "" {
		whereClause += " AND serviceName = ?"
		params = append(params, req.ServiceName)
	}

	// 模糊查询字段（使用LIKE，性能相对较低，放在最后）
	if req.RequestPath != "" {
		whereClause += " AND requestPath LIKE ?"
		params = append(params, "%"+req.RequestPath+"%")
	}

	return whereClause, params, nil
}

// parseTimeString 解析时间字符串为time.Time
func (dao *DatabaseMonitoringDAO) parseTimeString(timeStr string) (time.Time, error) {
	if timeStr == "" {
		return time.Time{}, nil
	}

	// 使用ctime包解析时间字符串，支持多种格式
	parsedTime, err := ctime.ParseTimeString(timeStr)
	if err != nil {
		return time.Time{}, huberrors.WrapError(err, "时间格式解析失败")
	}

	return parsedTime, nil
}

// getTimeGroupFormat 获取时间分组格式
// 根据时间粒度返回相应的SQL日期格式字符串
func (dao *DatabaseMonitoringDAO) getTimeGroupFormat(granularity models.TimeGranularity) string {
	// 使用兼容性较好的DATE_FORMAT函数（MySQL/MariaDB）
	// 注意：对于PostgreSQL，可能需要使用TO_CHAR函数
	switch granularity {
	case models.TimeGranularityMinute:
		return "DATE_FORMAT(gatewayStartProcessingTime, '%Y-%m-%d %H:%i')" // 精确到分钟
	case models.TimeGranularityHour:
		return "DATE_FORMAT(gatewayStartProcessingTime, '%Y-%m-%d %H')" // 精确到小时
	case models.TimeGranularityDay:
		return "DATE_FORMAT(gatewayStartProcessingTime, '%Y-%m-%d')" // 精确到天
	default:
		return "DATE_FORMAT(gatewayStartProcessingTime, '%Y-%m-%d %H:%i')" // 默认按分钟分组
	}
}

// getTimeGranularitySeconds 获取时间粒度对应的秒数
// 用于QPS计算
func (dao *DatabaseMonitoringDAO) getTimeGranularitySeconds(granularity models.TimeGranularity) int {
	switch granularity {
	case models.TimeGranularityMinute:
		return 60 // 1分钟 = 60秒
	case models.TimeGranularityHour:
		return 3600 // 1小时 = 3600秒
	case models.TimeGranularityDay:
		return 86400 // 1天 = 86400秒
	default:
		return 60 // 默认1分钟
	}
}

// getStatusCodeCategory 获取状态码分类
func (dao *DatabaseMonitoringDAO) getStatusCodeCategory(statusCode string) string {
	if len(statusCode) == 0 {
		return "未知"
	}

	switch statusCode[0] {
	case '2':
		return "成功"
	case '3':
		return "重定向"
	case '4':
		return "客户端错误"
	case '5':
		return "服务端错误"
	default:
		return "其他"
	}
}

// getStatusCodeDescription 获取状态码描述
func (dao *DatabaseMonitoringDAO) getStatusCodeDescription(statusCode string) string {
	statusCodeMap := map[string]string{
		"200": "OK",
		"201": "Created",
		"202": "Accepted",
		"204": "No Content",
		"301": "Moved Permanently",
		"302": "Found",
		"304": "Not Modified",
		"400": "Bad Request",
		"401": "Unauthorized",
		"403": "Forbidden",
		"404": "Not Found",
		"405": "Method Not Allowed",
		"408": "Request Timeout",
		"429": "Too Many Requests",
		"500": "Internal Server Error",
		"502": "Bad Gateway",
		"503": "Service Unavailable",
		"504": "Gateway Timeout",
	}

	if desc, exists := statusCodeMap[statusCode]; exists {
		return desc
	}
	return "Unknown"
}
