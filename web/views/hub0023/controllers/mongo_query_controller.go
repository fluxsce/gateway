package controllers

import (
	"fmt"
	"time"

	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	"gateway/pkg/utils/ctime"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0023/dao"
	"gateway/web/views/hub0023/models"

	"github.com/gin-gonic/gin"
)

// MongoQueryController MongoDB查询控制器
type MongoQueryController struct {
	mongoQueryDAO      *dao.MongoQueryDAO
	mongoMonitoringDAO *dao.MongoMonitoringDAO
}

// NewMongoQueryController 创建MongoDB查询控制器
func NewMongoQueryController(mongoClient *client.Client) *MongoQueryController {
	return &MongoQueryController{
		mongoQueryDAO:      dao.NewMongoQueryDAO(mongoClient),
		mongoMonitoringDAO: dao.NewMongoMonitoringDAO(mongoClient),
	}
}

// GetGatewayMonitoringOverview 获取网关监控概览数据
// @Summary 获取网关监控概览数据
// @Description 获取网关监控概览数据，包括总请求数、成功失败数、平均响应时间等
// @Tags MongoDB网关监控
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayMonitoringQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/mongo-gateway-monitoring/overview [post]
func (c *MongoQueryController) GetGatewayMonitoringOverview(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayMonitoringQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "网关监控概览查询参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID，不使用前端传递的值
	req.TenantId = request.GetTenantID(ctx)

	// 校验时间范围
	if err := c.validateTimeRange(&req); err != nil {
		logger.ErrorWithTrace(ctx, "网关监控概览查询时间范围校验失败", "error", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00007)
		return
	}

	// 调用DAO查询
	overview, err := c.mongoMonitoringDAO.GetGatewayMonitoringOverview(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "网关监控概览查询失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, overview, constants.SD00002)
}

// GetGatewayMonitoringChartData 获取网关监控图表数据
// @Summary 获取网关监控图表数据
// @Description 获取网关监控图表数据，包括请求趋势、响应时间趋势、状态码分布、热点路由等
// @Tags MongoDB网关监控
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayMonitoringQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/mongo-gateway-monitoring/chart-data [post]
func (c *MongoQueryController) GetGatewayMonitoringChartData(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayMonitoringQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "网关监控图表数据查询参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID，不使用前端传递的值
	req.TenantId = request.GetTenantID(ctx)

	// 校验时间范围
	if err := c.validateTimeRange(&req); err != nil {
		logger.ErrorWithTrace(ctx, "网关监控图表数据查询时间范围校验失败", "error", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00007)
		return
	}

	// 设置默认热点路由限制
	if req.HotRouteLimit <= 0 {
		req.HotRouteLimit = 10
	}

	// 并发查询各种监控数据
	requestTrend, err := c.mongoMonitoringDAO.GetRequestMetricsTrend(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取请求趋势数据失败", "error", err)
		response.ErrorJSON(ctx, "获取请求趋势数据失败: "+err.Error(), constants.ED00009)
		return
	}

	responseTimeTrend, err := c.mongoMonitoringDAO.GetResponseTimeMetricsTrend(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取响应时间趋势数据失败", "error", err)
		response.ErrorJSON(ctx, "获取响应时间趋势数据失败: "+err.Error(), constants.ED00009)
		return
	}

	statusCodeDistribution, err := c.mongoMonitoringDAO.GetStatusCodeDistribution(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取状态码分布数据失败", "error", err)
		response.ErrorJSON(ctx, "获取状态码分布数据失败: "+err.Error(), constants.ED00009)
		return
	}

	hotRoutes, err := c.mongoMonitoringDAO.GetHotRoutes(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取热点路由数据失败", "error", err)
		response.ErrorJSON(ctx, "获取热点路由数据失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建图表数据
	chartData := &models.GatewayMonitoringChartData{
		RequestTrend:           requestTrend,
		ResponseTimeTrend:      responseTimeTrend,
		StatusCodeDistribution: statusCodeDistribution,
		HotRoutes:              hotRoutes,
	}

	response.SuccessJSON(ctx, chartData, constants.SD00002)
}

// validateTimeRange 校验时间范围
// 要求：开始时间和结束时间必填，时间范围不能超过24小时
func (c *MongoQueryController) validateTimeRange(req *models.GatewayMonitoringQueryRequest) error {
	// 校验开始时间和结束时间是否为空
	if req.StartTime == "" {
		return fmt.Errorf("开始时间不能为空")
	}
	if req.EndTime == "" {
		return fmt.Errorf("结束时间不能为空")
	}

	// 解析时间
	startTime, err := ctime.ParseTimeString(req.StartTime)
	if err != nil {
		return fmt.Errorf("开始时间格式错误: %v", err)
	}

	endTime, err := ctime.ParseTimeString(req.EndTime)
	if err != nil {
		return fmt.Errorf("结束时间格式错误: %v", err)
	}

	// 校验时间范围
	if endTime.Before(startTime) {
		return fmt.Errorf("结束时间不能早于开始时间")
	}

	// 校验时间范围不能超过24小时
	duration := endTime.Sub(startTime)
	if duration > 24*time.Hour {
		return fmt.Errorf("查询时间范围不能超过24小时")
	}

	// 校验时间粒度
	// if err := c.validateTimeGranularity(req.TimeGranularity, duration); err != nil {
	// 	return err
	// }

	return nil
}

// validateTimeGranularity 校验时间粒度
func (c *MongoQueryController) validateTimeGranularity(granularity models.TimeGranularity, duration time.Duration) error {
	switch granularity {
	case models.TimeGranularityMinute:
		// 分钟粒度：适用于1小时以内的查询
		if duration > 1*time.Hour {
			return fmt.Errorf("分钟粒度查询时间范围不能超过1小时")
		}
	case models.TimeGranularityHour:
		// 小时粒度：适用于24小时以内的查询
		if duration > 24*time.Hour {
			return fmt.Errorf("小时粒度查询时间范围不能超过24小时")
		}
	case models.TimeGranularityDay:
		// 天粒度：适用于长时间查询，但当前限制为24小时
		if duration > 24*time.Hour {
			return fmt.Errorf("天粒度查询时间范围不能超过24小时")
		}
	default:
		return fmt.Errorf("不支持的时间粒度，支持的值: MINUTE, HOUR, DAY")
	}

	return nil
}

// QueryGatewayLogs 查询网关日志列表（MongoDB版本）
// @Summary 查询网关日志列表（MongoDB版本）
// @Description 支持分页查询和多条件过滤的MongoDB网关日志列表，不返回大字段以提高查询性能
// @Tags MongoDB网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayAccessLogQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/mongo-gateway-log/query [post]
func (c *MongoQueryController) QueryGatewayLogs(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志查询参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 使用公共方法获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	req.PageIndex = page
	req.PageSize = pageSize

	// 从上下文获取租户ID，不使用前端传递的值
	req.TenantId = request.GetTenantID(ctx)

	// 调用DAO查询
	logs, total, err := c.mongoQueryDAO.QueryGatewayLogs(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志查询失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, int(total))

	// 使用统一的分页响应
	response.PageJSON(ctx, logs, pageInfo, constants.SD00002)
}

// GetGatewayLog 获取网关日志详情（MongoDB版本）
// @Summary 获取网关日志详情（MongoDB版本）
// @Description 通过租户ID和链路追踪ID组合主键获取MongoDB网关日志详情
// @Tags MongoDB网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param get body models.GatewayAccessLogGetRequest true "获取参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/mongo-gateway-log/get [post]
func (c *MongoQueryController) GetGatewayLog(ctx *gin.Context) {
	// 解析获取参数
	var req models.GatewayAccessLogGetRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志获取参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID，不使用前端传递的值
	req.TenantId = request.GetTenantID(ctx)

	// 参数验证 - 需要链路追踪ID
	if req.TraceId == "" {
		response.ErrorJSON(ctx, "请提供链路追踪ID", constants.ED00007)
		return
	}

	// 调用DAO获取详情
	log, err := c.mongoQueryDAO.GetGatewayLogByKey(ctx, req.TenantId, req.TraceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志获取失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00008)
		return
	}

	// 查询关联的后端追踪日志（从表）
	backendTraces, err := c.mongoQueryDAO.GetBackendTracesByTraceID(ctx, req.TenantId, req.TraceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询MongoDB后端追踪日志失败", "error", err)
		// 后端追踪日志查询失败不影响主表返回，记录错误后继续
		backendTraces = []models.BackendTraceLog{}
	}

	// 填充后端追踪日志到主表对象
	log.BackendTraces = backendTraces

	response.SuccessJSON(ctx, log, constants.SD00002)
}

// CountGatewayLogs 统计网关日志数量（MongoDB版本）
// @Summary 统计网关日志数量（MongoDB版本）
// @Description 根据查询条件统计MongoDB网关日志数量
// @Tags MongoDB网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param count body models.GatewayAccessLogQueryRequest true "统计参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0023/mongo-gateway-log/count [post]
func (c *MongoQueryController) CountGatewayLogs(ctx *gin.Context) {
	// 解析统计参数
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志统计参数解析失败", "error", err)
		response.ErrorJSON(ctx, "参数解析错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID，不使用前端传递的值
	req.TenantId = request.GetTenantID(ctx)

	// 调用DAO统计
	count, err := c.mongoQueryDAO.CountGatewayLogs(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "MongoDB网关日志统计失败", "error", err)
		response.ErrorJSON(ctx, "统计失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"count":      count,
		"collection": models.GatewayAccessLog{}.TableName(),
	}, constants.SD00002)
}
