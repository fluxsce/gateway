package dao

import (
	"context"
	"fmt"

	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/utils/ctime"
	"gohub/pkg/utils/huberrors"
	"gohub/web/views/hub0023/models"
)

// ClickHouseQueryDAO ClickHouse查询数据访问对象
type ClickHouseQueryDAO struct {
	db database.Database
}

// NewClickHouseQueryDAO 创建ClickHouse查询DAO
func NewClickHouseQueryDAO(db database.Database) *ClickHouseQueryDAO {
	return &ClickHouseQueryDAO{
		db: db,
	}
}

// QueryGatewayLogs 查询网关日志列表（ClickHouse版本）
func (dao *ClickHouseQueryDAO) QueryGatewayLogs(ctx context.Context, req *models.GatewayAccessLogQueryRequest) ([]models.GatewayAccessLogSummary, int, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildGatewayLogFilter(req)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 构建基础查询语句 - 列表查询不返回大字段，优化ClickHouse性能
	baseQuery := fmt.Sprintf(`
		SELECT tenantId, traceId, gatewayInstanceId, gatewayInstanceName, gatewayNodeIp,
			   routeConfigId, routeName, serviceDefinitionId, serviceName, proxyType,
			   requestMethod, requestPath, requestQuery, requestSize, clientIpAddress,
			   clientPort, userAgent, userIdentifier, gatewayStartProcessingTime,
			   gatewayFinishedProcessingTime, totalProcessingTimeMs, gatewayProcessingTimeMs,
			   backendResponseTimeMs, gatewayStatusCode, backendStatusCode, responseSize,
			   matchedRoute, forwardAddress, forwardMethod, loadBalancerDecision, errorMessage,
			   errorCode, resetFlag, retryCount, resetCount, logLevel, logType,
			   addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
		FROM HUB_GW_ACCESS_LOG
		%s
		ORDER BY gatewayStartProcessingTime DESC
	`, whereClause)

	// 构建统计查询 - ClickHouse的COUNT优化
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as count
		FROM HUB_GW_ACCESS_LOG
		%s
	`, whereClause)

	// 执行统计查询
	var countResult struct {
		Count int64 `db:"count"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志统计失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "ClickHouse网关日志统计失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []models.GatewayAccessLogSummary{}, 0, nil
	}

	// 构建分页查询 - ClickHouse使用LIMIT OFFSET语法
	var paginatedQuery string
	var allArgs []interface{}
	
	if req.PageSize > 0 {
		offset := (req.PageIndex - 1) * req.PageSize
		paginatedQuery = baseQuery + fmt.Sprintf(" LIMIT %d OFFSET %d", req.PageSize, offset)
		allArgs = params
	} else {
		paginatedQuery = baseQuery
		allArgs = params
	}

	// 执行分页查询
	var logs []models.GatewayAccessLogSummary
	err = dao.db.Query(ctx, &logs, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志查询失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "ClickHouse网关日志查询失败")
	}

	return logs, int(countResult.Count), nil
}

// GetGatewayLogByKey 根据主键获取网关日志详情（ClickHouse版本）
func (dao *ClickHouseQueryDAO) GetGatewayLogByKey(ctx context.Context, tenantId, traceId string) (*models.GatewayAccessLog, error) {
	// 验证参数
	if tenantId == "" {
		return nil, huberrors.NewError("租户ID不能为空")
	}
	if traceId == "" {
		return nil, huberrors.NewError("链路追踪ID不能为空")
	}

	// 构建查询SQL - 获取完整字段
	sql := `
		SELECT tenantId, traceId, gatewayInstanceId, gatewayInstanceName, gatewayNodeIp,
			   routeConfigId, routeName, serviceDefinitionId, serviceName, proxyType, logConfigId,
			   requestMethod, requestPath, requestQuery, requestSize, requestHeaders, requestBody,
			   clientIpAddress, clientPort, userAgent, referer, userIdentifier,
			   gatewayStartProcessingTime, backendRequestStartTime, backendResponseReceivedTime,
			   gatewayFinishedProcessingTime, totalProcessingTimeMs, gatewayProcessingTimeMs,
			   backendResponseTimeMs, gatewayStatusCode, backendStatusCode, responseSize,
			   responseHeaders, responseBody, matchedRoute, forwardAddress, forwardMethod,
			   forwardParams, forwardHeaders, forwardBody, loadBalancerDecision, errorMessage,
			   errorCode, parentTraceId, resetFlag, retryCount, resetCount, logLevel, logType,
			   reserved1, reserved2, reserved3, reserved4, reserved5, extProperty,
			   addTime, addWho, editTime, editWho, oprSeqFlag, currentVersion, activeFlag, noteText
		FROM HUB_GW_ACCESS_LOG
		WHERE tenantId = ? AND traceId = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var logs []models.GatewayAccessLog
	err := dao.db.Query(ctx, &logs, sql, []interface{}{tenantId, traceId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志获取失败", "tenantId", tenantId, "traceId", traceId, "error", err)
		return nil, huberrors.WrapError(err, "ClickHouse网关日志获取失败")
	}

	if len(logs) == 0 {
		return nil, huberrors.NewError("网关日志不存在")
	}

	return &logs[0], nil
}

// CountGatewayLogs 统计网关日志数量（ClickHouse版本）
func (dao *ClickHouseQueryDAO) CountGatewayLogs(ctx context.Context, req *models.GatewayAccessLogQueryRequest) (int64, error) {
	// 构建查询条件
	whereClause, params, err := dao.buildGatewayLogFilter(req)
	if err != nil {
		return 0, huberrors.WrapError(err, "构建查询条件失败")
	}

	// 构建统计查询
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) as count
		FROM HUB_GW_ACCESS_LOG
		%s
	`, whereClause)

	// 执行统计查询
	var countResult struct {
		Count int64 `db:"count"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志统计失败", "error", err)
		return 0, huberrors.WrapError(err, "ClickHouse网关日志统计失败")
	}

	return countResult.Count, nil
}

// buildGatewayLogFilter 构建网关日志查询条件
func (dao *ClickHouseQueryDAO) buildGatewayLogFilter(req *models.GatewayAccessLogQueryRequest) (string, []interface{}, error) {
	whereClause := "WHERE activeFlag = 'Y'"
	var params []interface{}

	// 基础查询条件
	if req.TenantId != "" {
		whereClause += " AND tenantId = ?"
		params = append(params, req.TenantId)
	}

	if req.TraceId != "" {
		whereClause += " AND traceId = ?"
		params = append(params, req.TraceId)
	}

	if req.GatewayInstanceId != "" {
		whereClause += " AND gatewayInstanceId = ?"
		params = append(params, req.GatewayInstanceId)
	}

	if req.GatewayInstanceName != "" {
		whereClause += " AND gatewayInstanceName LIKE ?"
		params = append(params, "%"+req.GatewayInstanceName+"%")
	}

	if req.RouteConfigId != "" {
		whereClause += " AND routeConfigId = ?"
		params = append(params, req.RouteConfigId)
	}

	if req.RouteName != "" {
		whereClause += " AND routeName LIKE ?"
		params = append(params, "%"+req.RouteName+"%")
	}

	if req.ServiceDefinitionId != "" {
		whereClause += " AND serviceDefinitionId = ?"
		params = append(params, req.ServiceDefinitionId)
	}

	if req.ServiceName != "" {
		whereClause += " AND serviceName LIKE ?"
		params = append(params, "%"+req.ServiceName+"%")
	}

	if req.ProxyType != "" {
		whereClause += " AND proxyType = ?"
		params = append(params, req.ProxyType)
	}

	// 请求信息查询条件
	if req.RequestMethod != "" {
		whereClause += " AND requestMethod = ?"
		params = append(params, req.RequestMethod)
	}

	if req.RequestPath != "" {
		whereClause += " AND requestPath LIKE ?"
		params = append(params, "%"+req.RequestPath+"%")
	}

	if req.ClientIpAddress != "" {
		whereClause += " AND clientIpAddress = ?"
		params = append(params, req.ClientIpAddress)
	}

	if req.UserAgent != "" {
		whereClause += " AND userAgent LIKE ?"
		params = append(params, "%"+req.UserAgent+"%")
	}

	if req.UserIdentifier != "" {
		whereClause += " AND userIdentifier = ?"
		params = append(params, req.UserIdentifier)
	}

	// 响应信息查询条件
	if req.GatewayStatusCode != 0 {
		whereClause += " AND gatewayStatusCode = ?"
		params = append(params, req.GatewayStatusCode)
	}

	if req.BackendStatusCode != 0 {
		whereClause += " AND backendStatusCode = ?"
		params = append(params, req.BackendStatusCode)
	}

	// 错误信息查询条件
	if req.ErrorCode != "" {
		whereClause += " AND errorCode = ?"
		params = append(params, req.ErrorCode)
	}

	if req.ErrorMessage != "" {
		whereClause += " AND errorMessage LIKE ?"
		params = append(params, "%"+req.ErrorMessage+"%")
	}

	// 时间条件处理 - 使用ctime包正确解析时间字符串
	if req.StartTime != "" {
		startTime, err := ctime.ParseTimeString(req.StartTime)
		if err != nil {
			return "", nil, huberrors.WrapError(err, "开始时间格式不正确: %s", req.StartTime)
		}
		whereClause += " AND gatewayStartProcessingTime >= ?"
		params = append(params, startTime)
	}

	if req.EndTime != "" {
		endTime, err := ctime.ParseTimeString(req.EndTime)
		if err != nil {
			return "", nil, huberrors.WrapError(err, "结束时间格式不正确: %s", req.EndTime)
		}
		whereClause += " AND gatewayStartProcessingTime <= ?"
		params = append(params, endTime)
	}

	// 性能查询
	if req.MinProcessingTime > 0 {
		whereClause += " AND totalProcessingTimeMs >= ?"
		params = append(params, req.MinProcessingTime)
	}

	if req.MaxProcessingTime > 0 {
		whereClause += " AND totalProcessingTimeMs <= ?"
		params = append(params, req.MaxProcessingTime)
	}

	// 日志级别和类型
	if req.LogLevel != "" {
		whereClause += " AND logLevel = ?"
		params = append(params, req.LogLevel)
	}

	if req.LogType != "" {
		whereClause += " AND logType = ?"
		params = append(params, req.LogType)
	}

	// 重置标记查询
	if req.ResetFlag != "" {
		whereClause += " AND resetFlag = ?"
		params = append(params, req.ResetFlag)
	}

	// 关键词搜索
	if req.Keyword != "" {
		whereClause += " AND (requestPath LIKE ? OR errorMessage LIKE ? OR routeName LIKE ? OR serviceName LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword, keyword, keyword)
	}

	return whereClause, params, nil
} 