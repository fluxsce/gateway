package dao

import (
	"context"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0023/models"
)

// ClickHouseMonitoringDAO ClickHouse监控数据访问对象
// 专门用于从 HUB_GW_ACCESS_LOG 表中抽取各种监控统计数据
//
// 重要提示：
// 1. 为了优化查询性能，建议在ClickHouse中创建以下索引：
//   - PRIMARY KEY (gatewayStartProcessingTime, tenantId) 基于时间的主键
//   - ORDER BY (gatewayStartProcessingTime, tenantId, gatewayInstanceId) 排序键
//
// 2. 所有聚合查询都经过优化，充分利用ClickHouse的列式存储特性
// 3. 查询时间范围已在控制器层限制为24小时内，防止大数据量查询
// 4. 使用ClickHouse的聚合函数进行高效统计
type ClickHouseMonitoringDAO struct {
	db database.Database
}

// NewClickHouseMonitoringDAO 创建ClickHouse监控数据DAO
func NewClickHouseMonitoringDAO(db database.Database) *ClickHouseMonitoringDAO {
	return &ClickHouseMonitoringDAO{
		db: db,
	}
}

// GetGatewayMonitoringOverview 获取网关监控概览数据
func (dao *ClickHouseMonitoringDAO) GetGatewayMonitoringOverview(ctx context.Context, req *models.GatewayMonitoringQueryRequest) (*models.GatewayMonitoringOverview, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 构建概览查询SQL - 使用ClickHouse的聚合函数一次性计算所有指标
	sql := fmt.Sprintf(`
		SELECT 
			COUNT(*) as totalRequests,
			countIf(gatewayStatusCode >= 200 AND gatewayStatusCode < 300) as successRequests,
			countIf(gatewayStatusCode >= 400 OR gatewayStatusCode < 200) as failedRequests,
			avgIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as avgResponseTime,
			minIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as minResponseTime,
			maxIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as maxResponseTime
		FROM HUB_GW_ACCESS_LOG
		%s
	`, whereClause)

	// 执行查询
	var result struct {
		TotalRequests   int64   `db:"totalRequests"`
		SuccessRequests int64   `db:"successRequests"`
		FailedRequests  int64   `db:"failedRequests"`
		AvgResponseTime float64 `db:"avgResponseTime"`
		MinResponseTime float64 `db:"minResponseTime"`
		MaxResponseTime float64 `db:"maxResponseTime"`
	}

	err = dao.db.QueryOne(ctx, &result, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse监控概览查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse监控概览查询失败")
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
func (dao *ClickHouseMonitoringDAO) GetRequestMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.RequestMetrics, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取时间分组格式和时间分组函数
	timeFormat := dao.getTimeGroupFormat(req.TimeGranularity)
	timeGroupFunc := dao.getTimeGroupFunction(req.TimeGranularity)
	granularitySeconds := dao.getTimeGranularitySeconds(req.TimeGranularity)

	// 构建趋势查询SQL - 使用ClickHouse的时间分组函数
	sql := fmt.Sprintf(`
		SELECT 
			formatDateTime(gatewayStartProcessingTime, '%s') as timeGroup,
			COUNT(*) as totalRequests,
			countIf(gatewayStatusCode >= 200 AND gatewayStatusCode < 300) as successRequests,
			countIf(gatewayStatusCode >= 400 OR gatewayStatusCode < 200) as failedRequests,
			toUnixTimestamp(%s(gatewayStartProcessingTime)) * 1000 as timestamp
		FROM HUB_GW_ACCESS_LOG
		%s
		GROUP BY timeGroup, timestamp
		ORDER BY timestamp
	`, timeFormat, timeGroupFunc, whereClause)

	// 执行查询
	var results []struct {
		TimeGroup       string `db:"timeGroup"`
		TotalRequests   int64  `db:"totalRequests"`
		SuccessRequests int64  `db:"successRequests"`
		FailedRequests  int64  `db:"failedRequests"`
		Timestamp       int64  `db:"timestamp"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse请求趋势查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse请求趋势查询失败")
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
			Timestamp:         result.Timestamp,
			TotalRequests:     result.TotalRequests,
			SuccessRequests:   result.SuccessRequests,
			FailedRequests:    result.FailedRequests,
			RequestsPerSecond: qps,
		})
	}

	return metrics, nil
}

// GetResponseTimeMetricsTrend 获取响应时间指标趋势数据
func (dao *ClickHouseMonitoringDAO) GetResponseTimeMetricsTrend(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.ResponseTimeMetrics, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildMonitoringFilter(req)
	if err != nil {
		return nil, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 获取时间分组格式和时间分组函数
	timeFormat := dao.getTimeGroupFormat(req.TimeGranularity)
	timeGroupFunc := dao.getTimeGroupFunction(req.TimeGranularity)

	// 构建响应时间趋势查询SQL - 使用ClickHouse的百分位数函数
	sql := fmt.Sprintf(`
		SELECT 
			formatDateTime(gatewayStartProcessingTime, '%s') as timeGroup,
			COUNT(*) as requestCount,
			avgIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as avgResponseTime,
			minIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as minResponseTime,
			maxIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as maxResponseTime,
			quantileIf(0.5)(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as p50ResponseTime,
			quantileIf(0.9)(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as p90ResponseTime,
			quantileIf(0.99)(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as p99ResponseTime,
			toUnixTimestamp(%s(gatewayStartProcessingTime)) * 1000 as timestamp
		FROM HUB_GW_ACCESS_LOG
		%s AND totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0
		GROUP BY timeGroup, timestamp
		ORDER BY timestamp
	`, timeFormat, timeGroupFunc, whereClause)

	// 执行查询
	var results []struct {
		TimeGroup       string  `db:"timeGroup"`
		RequestCount    int64   `db:"requestCount"`
		AvgResponseTime float64 `db:"avgResponseTime"`
		MinResponseTime float64 `db:"minResponseTime"`
		MaxResponseTime float64 `db:"maxResponseTime"`
		P50ResponseTime float64 `db:"p50ResponseTime"`
		P90ResponseTime float64 `db:"p90ResponseTime"`
		P99ResponseTime float64 `db:"p99ResponseTime"`
		Timestamp       int64   `db:"timestamp"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse响应时间趋势查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse响应时间趋势查询失败")
	}

	// 检查结果是否为空，如果为空直接返回空切片
	if len(results) == 0 {
		return []models.ResponseTimeMetrics{}, nil
	}

	// 转换为响应格式
	metrics := make([]models.ResponseTimeMetrics, 0, len(results))
	for _, result := range results {
		// 平均响应时间最多保留两位小数
		avgResponseTime := roundToTwoDecimalPlaces(result.AvgResponseTime)

		metrics = append(metrics, models.ResponseTimeMetrics{
			Timestamp:         result.Timestamp,
			RequestCount:      result.RequestCount,
			AvgResponseTimeMs: avgResponseTime,
			MinResponseTimeMs: int(result.MinResponseTime),
			MaxResponseTimeMs: int(result.MaxResponseTime),
			P50ResponseTimeMs: int(result.P50ResponseTime),
			P90ResponseTimeMs: int(result.P90ResponseTime),
			P99ResponseTimeMs: int(result.P99ResponseTime),
		})
	}

	return metrics, nil
}

// GetStatusCodeDistribution 获取状态码分布数据
func (dao *ClickHouseMonitoringDAO) GetStatusCodeDistribution(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringStatusCodeData, error) {
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
		logger.ErrorWithTrace(ctx, "ClickHouse状态码分布总数查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse状态码分布总数查询失败")
	}

	if totalResult.Total == 0 {
		return []models.GatewayMonitoringStatusCodeData{}, nil
	}

	// 构建状态码分布查询SQL
	sql := fmt.Sprintf(`
		SELECT 
			toString(gatewayStatusCode) as statusCode,
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
		logger.ErrorWithTrace(ctx, "ClickHouse状态码分布查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse状态码分布查询失败")
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
func (dao *ClickHouseMonitoringDAO) GetHotRoutes(ctx context.Context, req *models.GatewayMonitoringQueryRequest) ([]models.GatewayMonitoringHotRouteData, error) {
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
			countIf(gatewayStatusCode >= 400 OR gatewayStatusCode < 200) as errorCount,
			avgIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as avgResponseTime,
			minIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as minResponseTime,
			maxIf(totalProcessingTimeMs, totalProcessingTimeMs IS NOT NULL AND totalProcessingTimeMs > 0) as maxResponseTime
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
		MinResponseTime float64 `db:"minResponseTime"`
		MaxResponseTime float64 `db:"maxResponseTime"`
	}

	err = dao.db.Query(ctx, &results, sql, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse热点路由查询失败", "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse热点路由查询失败")
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
func (dao *ClickHouseMonitoringDAO) buildMonitoringFilter(req *models.GatewayMonitoringQueryRequest) (string, []interface{}, error) {
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
func (dao *ClickHouseMonitoringDAO) parseTimeString(timeStr string) (time.Time, error) {
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
// 根据时间粒度返回相应的ClickHouse日期格式字符串
func (dao *ClickHouseMonitoringDAO) getTimeGroupFormat(granularity models.TimeGranularity) string {
	switch granularity {
	case models.TimeGranularityMinute:
		return "%Y-%m-%d %H:%i" // 精确到分钟
	case models.TimeGranularityHour:
		return "%Y-%m-%d %H" // 精确到小时
	case models.TimeGranularityDay:
		return "%Y-%m-%d" // 精确到天
	default:
		return "%Y-%m-%d %H:%i" // 默认按分钟分组
	}
}

// getTimeGroupFunction 获取时间分组函数
// 根据时间粒度返回相应的ClickHouse时间分组函数名
func (dao *ClickHouseMonitoringDAO) getTimeGroupFunction(granularity models.TimeGranularity) string {
	switch granularity {
	case models.TimeGranularityMinute:
		return "toStartOfMinute" // 精确到分钟
	case models.TimeGranularityHour:
		return "toStartOfHour" // 精确到小时
	case models.TimeGranularityDay:
		return "toStartOfDay" // 精确到天
	default:
		return "toStartOfMinute" // 默认按分钟分组
	}
}

// getTimeGranularitySeconds 获取时间粒度对应的秒数
// 用于QPS计算
func (dao *ClickHouseMonitoringDAO) getTimeGranularitySeconds(granularity models.TimeGranularity) int {
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
func (dao *ClickHouseMonitoringDAO) getStatusCodeCategory(statusCode string) string {
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
func (dao *ClickHouseMonitoringDAO) getStatusCodeDescription(statusCode string) string {
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
