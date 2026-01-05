package dao

import (
	"context"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/huberrors"
	"gateway/web/views/hub0023/models"
)

// BackendTraceLogDAO 后端追踪日志数据访问对象
type BackendTraceLogDAO struct {
	db database.Database
}

// NewBackendTraceLogDAO 创建后端追踪日志DAO
func NewBackendTraceLogDAO(db database.Database) *BackendTraceLogDAO {
	return &BackendTraceLogDAO{
		db: db,
	}
}

// GetByTraceID 根据租户ID和链路追踪ID获取该请求的所有后端追踪日志
// 一个主请求可能会转发到多个后端服务，因此返回多条记录
func (dao *BackendTraceLogDAO) GetByTraceID(ctx context.Context, tenantID, traceID string) ([]models.BackendTraceLog, error) {
	sql := `
		SELECT tenantId, traceId, backendTraceId, serviceDefinitionId, serviceName,
			   forwardAddress, forwardMethod, forwardPath, forwardQuery, 
			   forwardHeaders, forwardBody, requestSize,
			   loadBalancerStrategy, loadBalancerDecision,
			   requestStartTime, responseReceivedTime, requestDurationMs,
			   statusCode, responseSize, responseHeaders, responseBody,
			   errorCode, errorMessage, successFlag, traceStatus, retryCount,
			   extProperty, addTime, addWho, editTime, editWho, 
			   oprSeqFlag, currentVersion, activeFlag, noteText
		FROM HUB_GW_BACKEND_TRACE_LOG
		WHERE tenantId = ? AND traceId = ? AND activeFlag = 'Y'
		ORDER BY requestStartTime ASC
	`

	var logs []models.BackendTraceLog
	err := dao.db.Query(ctx, &logs, sql, []interface{}{tenantID, traceID}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询后端追踪日志失败", "tenantId", tenantID, "traceId", traceID, "error", err)
		return nil, huberrors.WrapError(err, "后端追踪日志查询失败")
	}

	// 即使没有记录也返回空列表，不返回错误
	if logs == nil {
		logs = []models.BackendTraceLog{}
	}

	return logs, nil
}

// GetByBackendTraceID 根据租户ID、链路追踪ID和后端追踪ID获取具体的后端追踪日志
// 如果查询不到数据，返回 nil, nil（不返回错误）
func (dao *BackendTraceLogDAO) GetByBackendTraceID(ctx context.Context, tenantID, traceID, backendTraceID string) (*models.BackendTraceLog, error) {
	sql := `
		SELECT tenantId, traceId, backendTraceId, serviceDefinitionId, serviceName,
			   forwardAddress, forwardMethod, forwardPath, forwardQuery, 
			   forwardHeaders, forwardBody, requestSize,
			   loadBalancerStrategy, loadBalancerDecision,
			   requestStartTime, responseReceivedTime, requestDurationMs,
			   statusCode, responseSize, responseHeaders, responseBody,
			   errorCode, errorMessage, successFlag, traceStatus, retryCount,
			   extProperty, addTime, addWho, editTime, editWho, 
			   oprSeqFlag, currentVersion, activeFlag, noteText
		FROM HUB_GW_BACKEND_TRACE_LOG
		WHERE tenantId = ? AND traceId = ? AND backendTraceId = ? AND activeFlag = 'Y'
	`

	var logs []models.BackendTraceLog
	err := dao.db.Query(ctx, &logs, sql, []interface{}{tenantID, traceID, backendTraceID}, true)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询后端追踪日志失败", "tenantId", tenantID, "traceId", traceID, "backendTraceId", backendTraceID, "error", err)
		return nil, huberrors.WrapError(err, "后端追踪日志查询失败")
	}

	// 查询不到数据是正常情况，返回 nil, nil
	if len(logs) == 0 {
		return nil, nil
	}

	return &logs[0], nil
}
