package dao

import (
	"context"
	"fmt"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
	"gateway/pkg/logger"
	"gateway/pkg/utils/ctime"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0023/models"
	"strings"
	"time"
)

// GatewayLogDAO 网关日志数据访问对象
type GatewayLogDAO struct {
	db database.Database
}

// NewGatewayLogDAO 创建网关日志DAO
func NewGatewayLogDAO(db database.Database) *GatewayLogDAO {
	return &GatewayLogDAO{
		db: db,
	}
}

// GetByKey 根据租户ID和链路追踪ID获取网关日志
func (dao *GatewayLogDAO) GetByKey(ctx context.Context, tenantId, traceId string) (*models.GatewayAccessLog, error) {
	var log models.GatewayAccessLog
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
	`

	var logs []models.GatewayAccessLog
	err := dao.db.Query(ctx, &logs, sql, []interface{}{tenantId, traceId}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网关日志失败", "error", err)
		return nil, fmt.Errorf("网关日志查询失败: %v", err)
	}

	if len(logs) == 0 {
		return nil, fmt.Errorf("网关日志不存在")
	}

	log = logs[0]

	return &log, nil
}

// Query 查询网关日志列表
// 性能优化：列表查询不返回大字段（requestHeaders, requestBody, responseHeaders, responseBody, forwardParams, forwardHeaders, forwardBody, extProperty）
// 这些字段可能包含大量数据，在列表展示时不需要，只在详情查询时获取
// 这样可以显著提高查询性能，减少网络传输量和内存使用
func (dao *GatewayLogDAO) Query(ctx context.Context, req *models.GatewayAccessLogQueryRequest) ([]models.GatewayAccessLogSummary, int, error) {
	// 构建查询条件
	whereClause := "WHERE activeFlag = 'Y'"
	var params []interface{}

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

	if req.GatewayStatusCode != 0 {
		whereClause += " AND gatewayStatusCode = ?"
		params = append(params, req.GatewayStatusCode)
	}

	if req.BackendStatusCode != 0 {
		whereClause += " AND backendStatusCode = ?"
		params = append(params, req.BackendStatusCode)
	}

	if req.ErrorCode != "" {
		whereClause += " AND errorCode = ?"
		params = append(params, req.ErrorCode)
	}

	if req.ErrorMessage != "" {
		whereClause += " AND errorMessage LIKE ?"
		params = append(params, "%"+req.ErrorMessage+"%")
	}

	// 时间条件处理 - 使用ctime包正确解析时间字符串
	// Oracle数据库需要将字符串转换为time.Time类型才能正确进行时间比较
	if req.StartTime != "" {
		startTime, err := ctime.ParseTimeString(req.StartTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "开始时间格式不正确", "startTime", req.StartTime, "error", err)
			return nil, 0, huberrors.WrapError(err, "开始时间格式不正确: %s", req.StartTime)
		}
		whereClause += " AND gatewayStartProcessingTime >= ?"
		params = append(params, startTime)
	}

	if req.EndTime != "" {
		endTime, err := ctime.ParseTimeString(req.EndTime)
		if err != nil {
			logger.ErrorWithTrace(ctx, "结束时间格式不正确", "endTime", req.EndTime, "error", err)
			return nil, 0, huberrors.WrapError(err, "结束时间格式不正确: %s", req.EndTime)
		}
		whereClause += " AND gatewayStartProcessingTime <= ?"
		params = append(params, endTime)
	}

	if req.MinProcessingTime > 0 {
		whereClause += " AND totalProcessingTimeMs >= ?"
		params = append(params, req.MinProcessingTime)
	}

	if req.MaxProcessingTime > 0 {
		whereClause += " AND totalProcessingTimeMs <= ?"
		params = append(params, req.MaxProcessingTime)
	}

	if req.LogLevel != "" {
		whereClause += " AND logLevel = ?"
		params = append(params, req.LogLevel)
	}

	if req.LogType != "" {
		whereClause += " AND logType = ?"
		params = append(params, req.LogType)
	}

	if req.ResetFlag != "" {
		whereClause += " AND resetFlag = ?"
		params = append(params, req.ResetFlag)
	}

	if req.Keyword != "" {
		whereClause += " AND (requestPath LIKE ? OR errorMessage LIKE ? OR routeName LIKE ? OR serviceName LIKE ?)"
		keyword := "%" + req.Keyword + "%"
		params = append(params, keyword, keyword, keyword, keyword)
	}

	// 构建基础查询语句 - 列表查询不返回大字段
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
		FROM HUB_GW_ACCESS_LOG %s
		ORDER BY gatewayStartProcessingTime DESC
	`, whereClause)

	// 构建统计查询
	countQuery, err := sqlutils.BuildCountQuery(baseQuery)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建统计查询失败")
	}

	// 执行统计查询
	var countResult struct {
		Count int `db:"COUNT(*)"`
	}
	err = dao.db.QueryOne(ctx, &countResult, countQuery, params, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网关日志总数失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询网关日志总数失败")
	}

	// 如果没有记录，直接返回空列表
	if countResult.Count == 0 {
		return []models.GatewayAccessLogSummary{}, 0, nil
	}

	// 创建分页信息
	pagination := sqlutils.NewPaginationInfo(req.PageIndex, req.PageSize)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(dao.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, 0, huberrors.WrapError(err, "构建分页查询失败")
	}

	// 合并查询参数
	allArgs := append(params, paginationArgs...)

	// 执行分页查询
	var logs []models.GatewayAccessLogSummary
	err = dao.db.Query(ctx, &logs, paginatedQuery, allArgs, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网关日志数据失败", "error", err)
		return nil, 0, huberrors.WrapError(err, "查询网关日志数据失败")
	}

	return logs, countResult.Count, nil
}

// Reset 重置网关日志（支持批量重置）
func (dao *GatewayLogDAO) Reset(ctx context.Context, req *models.GatewayAccessLogResetRequest, operatorId string) (int64, error) {
	now := time.Now()

	// 构建重置条件 - 支持按组合主键列表重置
	var whereConditions []string
	var params []interface{}

	for _, item := range req.LogItems {
		whereConditions = append(whereConditions, "(tenantId = ? AND traceId = ?)")
		params = append(params, item.TenantId, item.TraceId)
	}

	whereClause := "WHERE activeFlag = 'Y' AND (" + strings.Join(whereConditions, " OR ") + ")"

	// 重置操作：清空响应相关字段，将日志状态重置为初始状态
	sql := fmt.Sprintf(`
		UPDATE HUB_GW_ACCESS_LOG SET
			gatewayFinishedProcessingTime = NULL,
			totalProcessingTimeMs = 0,
			gatewayProcessingTimeMs = 0,
			backendResponseTimeMs = 0,
			gatewayStatusCode = 0,
			backendStatusCode = 0,
			responseHeaders = '',
			responseBody = '',
			errorMessage = '',
			errorCode = '',
			resetFlag = 'Y',
			resetCount = resetCount + 1,
			editTime = ?,
			editWho = ?,
			currentVersion = currentVersion + 1
		%s
	`, whereClause)

	// 添加更新参数
	updateParams := append([]interface{}{now, operatorId}, params...)

	result, err := dao.db.Exec(ctx, sql, updateParams, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "重置网关日志失败", "error", err)
		return 0, err
	}

	return result, nil
}
