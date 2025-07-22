package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0023/dao"
	"gateway/web/views/hub0023/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GatewayLogController 网关日志控制器
type GatewayLogController struct {
	gatewayLogDAO *dao.GatewayLogDAO
}

// NewGatewayLogController 创建网关日志控制器
func NewGatewayLogController(db database.Database) *GatewayLogController {
	return &GatewayLogController{
		gatewayLogDAO: dao.NewGatewayLogDAO(db),
	}
}

// Query 查询网关日志列表
// @Summary 查询网关日志列表
// @Description 支持分页查询和多条件过滤的网关日志列表。为了提高查询性能，列表查询不返回大字段（如请求头、请求体、响应头、响应体、错误堆栈等），这些详细信息可通过详情接口获取。
// @Tags 网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayAccessLogQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/gateway-log/query [post]
func (c *GatewayLogController) Query(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "网关日志查询参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.PageIndex = page
	req.PageSize = pageSize

	// 强制从上下文获取租户ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 调用DAO查询
	gatewayLogs, total, err := c.gatewayLogDAO.Query(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网关日志失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	gatewayLogList := make([]map[string]interface{}, 0, len(gatewayLogs))
	for _, gatewayLog := range gatewayLogs {
		gatewayLogList = append(gatewayLogList, gatewayLogSummaryToMap(&gatewayLog))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "traceId"

	// 使用统一的分页响应
	response.PageJSON(ctx, gatewayLogList, pageInfo, constants.SD00002)
}

// Get 获取网关日志详情
// @Summary 获取网关日志详情
// @Description 通过租户ID和链路追踪ID组合主键获取网关日志详情
// @Tags 网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param get body models.GatewayAccessLogGetRequest true "获取参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/gateway-log/get [post]
func (c *GatewayLogController) Get(ctx *gin.Context) {
	// 解析获取参数
	var req models.GatewayAccessLogGetRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "网关日志获取参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	req.TenantId = tenantId

	// 参数验证 - 需要链路追踪ID
	if req.TraceId == "" {
		response.ErrorJSON(ctx, "请提供链路追踪ID", constants.ED00007)
		return
	}

	// 根据组合主键查询
	log, err := c.gatewayLogDAO.GetByKey(ctx, req.TenantId, req.TraceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "根据组合主键查询失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00008)
		return
	}

	// 转换为响应格式
	gatewayLogInfo := gatewayLogToMap(log)

	response.SuccessJSON(ctx, gatewayLogInfo, constants.SD00002)
}

// Reset 重置网关日志（支持批量重置）
// @Summary 重置网关日志（支持批量重置）
// @Description 通过租户ID和链路追踪ID组合主键重置指定的网关日志记录
// @Tags 网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param reset body models.GatewayAccessLogResetRequest true "重置参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/gateway-log/reset [post]
func (c *GatewayLogController) Reset(ctx *gin.Context) {
	// 解析重置参数
	var req models.GatewayAccessLogResetRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "网关日志重置参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 参数验证 - 需要日志项列表
	if len(req.LogItems) == 0 {
		response.ErrorJSON(ctx, "请提供日志项列表", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 为所有日志项设置租户ID，确保数据安全
	for i := range req.LogItems {
		req.LogItems[i].TenantId = tenantId
	}

	// 获取操作者ID
	operatorId := request.GetOperatorID(ctx)
	if operatorId == "" {
		operatorId = "SYSTEM"
	}

	// 调用DAO重置
	affectedRows, err := c.gatewayLogDAO.Reset(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "重置网关日志失败", "error", err)
		response.ErrorJSON(ctx, "重置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"affectedRows": affectedRows,
		"message":      "重置成功，影响 " + strconv.FormatInt(affectedRows, 10) + " 条记录",
	}, constants.SD00003)
}

// gatewayLogSummaryToMap 将网关日志摘要转换为Map格式，过滤敏感字段
func gatewayLogSummaryToMap(gatewayLog *models.GatewayAccessLogSummary) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                      gatewayLog.TenantId,
		"traceId":                       gatewayLog.TraceId,
		"gatewayInstanceId":             gatewayLog.GatewayInstanceId,
		"gatewayInstanceName":           gatewayLog.GatewayInstanceName,
		"gatewayNodeIp":                 gatewayLog.GatewayNodeIp,
		"routeConfigId":                 gatewayLog.RouteConfigId,
		"routeName":                     gatewayLog.RouteName,
		"serviceDefinitionId":           gatewayLog.ServiceDefinitionId,
		"serviceName":                   gatewayLog.ServiceName,
		"proxyType":                     gatewayLog.ProxyType,
		"requestMethod":                 gatewayLog.RequestMethod,
		"requestPath":                   gatewayLog.RequestPath,
		"requestQuery":                  gatewayLog.RequestQuery,
		"requestSize":                   gatewayLog.RequestSize,
		"clientIpAddress":               gatewayLog.ClientIpAddress,
		"clientPort":                    gatewayLog.ClientPort,
		"userAgent":                     gatewayLog.UserAgent,
		"userIdentifier":                gatewayLog.UserIdentifier,
		"gatewayStartProcessingTime":    gatewayLog.GatewayStartProcessingTime,
		"gatewayFinishedProcessingTime": gatewayLog.GatewayFinishedProcessingTime,
		"totalProcessingTimeMs":         gatewayLog.TotalProcessingTimeMs,
		"gatewayProcessingTimeMs":       gatewayLog.GatewayProcessingTimeMs,
		"backendResponseTimeMs":         gatewayLog.BackendResponseTimeMs,
		"gatewayStatusCode":             gatewayLog.GatewayStatusCode,
		"backendStatusCode":             gatewayLog.BackendStatusCode,
		"responseSize":                  gatewayLog.ResponseSize,
		"matchedRoute":                  gatewayLog.MatchedRoute,
		"forwardAddress":                gatewayLog.ForwardAddress,
		"forwardMethod":                 gatewayLog.ForwardMethod,
		"loadBalancerDecision":          gatewayLog.LoadBalancerDecision,
		"errorMessage":                  gatewayLog.ErrorMessage,
		"errorCode":                     gatewayLog.ErrorCode,
		"resetFlag":                     gatewayLog.ResetFlag,
		"retryCount":                    gatewayLog.RetryCount,
		"resetCount":                    gatewayLog.ResetCount,
		"logLevel":                      gatewayLog.LogLevel,
		"logType":                       gatewayLog.LogType,
		"addTime":                       gatewayLog.AddTime,
		"addWho":                        gatewayLog.AddWho,
		"editTime":                      gatewayLog.EditTime,
		"editWho":                       gatewayLog.EditWho,
		"oprSeqFlag":                    gatewayLog.OprSeqFlag,
		"currentVersion":                gatewayLog.CurrentVersion,
		"activeFlag":                    gatewayLog.ActiveFlag,
		"noteText":                      gatewayLog.NoteText,
	}
}

// gatewayLogToMap 将网关日志转换为Map格式（用于详情查询，包含所有字段）
func gatewayLogToMap(gatewayLog *models.GatewayAccessLog) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                      gatewayLog.TenantId,
		"traceId":                       gatewayLog.TraceId,
		"gatewayInstanceId":             gatewayLog.GatewayInstanceId,
		"gatewayInstanceName":           gatewayLog.GatewayInstanceName,
		"gatewayNodeIp":                 gatewayLog.GatewayNodeIp,
		"routeConfigId":                 gatewayLog.RouteConfigId,
		"routeName":                     gatewayLog.RouteName,
		"serviceDefinitionId":           gatewayLog.ServiceDefinitionId,
		"serviceName":                   gatewayLog.ServiceName,
		"proxyType":                     gatewayLog.ProxyType,
		"logConfigId":                   gatewayLog.LogConfigId,
		"requestMethod":                 gatewayLog.RequestMethod,
		"requestPath":                   gatewayLog.RequestPath,
		"requestQuery":                  gatewayLog.RequestQuery,
		"requestSize":                   gatewayLog.RequestSize,
		"requestHeaders":                gatewayLog.RequestHeaders,
		"requestBody":                   gatewayLog.RequestBody,
		"clientIpAddress":               gatewayLog.ClientIpAddress,
		"clientPort":                    gatewayLog.ClientPort,
		"userAgent":                     gatewayLog.UserAgent,
		"referer":                       gatewayLog.Referer,
		"userIdentifier":                gatewayLog.UserIdentifier,
		"gatewayStartProcessingTime":    gatewayLog.GatewayStartProcessingTime,
		"backendRequestStartTime":       gatewayLog.BackendRequestStartTime,
		"backendResponseReceivedTime":   gatewayLog.BackendResponseReceivedTime,
		"gatewayFinishedProcessingTime": gatewayLog.GatewayFinishedProcessingTime,
		"totalProcessingTimeMs":         gatewayLog.TotalProcessingTimeMs,
		"gatewayProcessingTimeMs":       gatewayLog.GatewayProcessingTimeMs,
		"backendResponseTimeMs":         gatewayLog.BackendResponseTimeMs,
		"gatewayStatusCode":             gatewayLog.GatewayStatusCode,
		"backendStatusCode":             gatewayLog.BackendStatusCode,
		"responseSize":                  gatewayLog.ResponseSize,
		"responseHeaders":               gatewayLog.ResponseHeaders,
		"responseBody":                  gatewayLog.ResponseBody,
		"matchedRoute":                  gatewayLog.MatchedRoute,
		"forwardAddress":                gatewayLog.ForwardAddress,
		"forwardMethod":                 gatewayLog.ForwardMethod,
		"forwardParams":                 gatewayLog.ForwardParams,
		"forwardHeaders":                gatewayLog.ForwardHeaders,
		"forwardBody":                   gatewayLog.ForwardBody,
		"loadBalancerDecision":          gatewayLog.LoadBalancerDecision,
		"errorMessage":                  gatewayLog.ErrorMessage,
		"errorCode":                     gatewayLog.ErrorCode,
		"parentTraceId":                 gatewayLog.ParentTraceId,
		"resetFlag":                     gatewayLog.ResetFlag,
		"retryCount":                    gatewayLog.RetryCount,
		"resetCount":                    gatewayLog.ResetCount,
		"logLevel":                      gatewayLog.LogLevel,
		"logType":                       gatewayLog.LogType,
		"reserved1":                     gatewayLog.Reserved1,
		"reserved2":                     gatewayLog.Reserved2,
		"reserved3":                     gatewayLog.Reserved3,
		"reserved4":                     gatewayLog.Reserved4,
		"reserved5":                     gatewayLog.Reserved5,
		"extProperty":                   gatewayLog.ExtProperty,
		"addTime":                       gatewayLog.AddTime,
		"addWho":                        gatewayLog.AddWho,
		"editTime":                      gatewayLog.EditTime,
		"editWho":                       gatewayLog.EditWho,
		"oprSeqFlag":                    gatewayLog.OprSeqFlag,
		"currentVersion":                gatewayLog.CurrentVersion,
		"activeFlag":                    gatewayLog.ActiveFlag,
		"noteText":                      gatewayLog.NoteText,
	}
}
