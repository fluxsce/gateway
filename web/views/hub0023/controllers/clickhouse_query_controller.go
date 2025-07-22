package controllers

import (
	"fmt"
	"time"

	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/utils/ctime"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0023/dao"
	"gohub/web/views/hub0023/models"

	"github.com/gin-gonic/gin"
)

// ClickHouseQueryController ClickHouse查询控制器
type ClickHouseQueryController struct {
	clickhouseQueryDAO      *dao.ClickHouseQueryDAO
	clickhouseMonitoringDAO *dao.ClickHouseMonitoringDAO
}

// NewClickHouseQueryController 创建ClickHouse查询控制器
func NewClickHouseQueryController(db database.Database) *ClickHouseQueryController {
	return &ClickHouseQueryController{
		clickhouseQueryDAO:      dao.NewClickHouseQueryDAO(db),
		clickhouseMonitoringDAO: dao.NewClickHouseMonitoringDAO(db),
	}
}

// GetGatewayMonitoringOverview 获取网关监控概览数据（ClickHouse版本）
// @Summary 获取网关监控概览数据（ClickHouse版本）
// @Description 获取网关监控概览数据，包括总请求数、成功失败数、平均响应时间等
// @Tags ClickHouse网关监控
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayMonitoringQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0023/clickhouse-gateway-monitoring/overview [post]
func (c *ClickHouseQueryController) GetGatewayMonitoringOverview(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayMonitoringQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关监控概览查询参数解析失败", "error", err)
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

	// 校验时间范围
	if err := c.validateTimeRange(&req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关监控概览查询时间范围校验失败", "error", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00007)
		return
	}

	// 调用DAO查询
	overview, err := c.clickhouseMonitoringDAO.GetGatewayMonitoringOverview(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关监控概览查询失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, overview, constants.SD00002)
}

// GetGatewayMonitoringChartData 获取网关监控图表数据（ClickHouse版本）
// @Summary 获取网关监控图表数据（ClickHouse版本）
// @Description 获取网关监控图表数据，包括请求趋势、响应时间趋势、状态码分布、热点路由等
// @Tags ClickHouse网关监控
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayMonitoringQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0023/clickhouse-gateway-monitoring/chart-data [post]
func (c *ClickHouseQueryController) GetGatewayMonitoringChartData(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayMonitoringQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关监控图表数据查询参数解析失败", "error", err)
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

	// 校验时间范围
	if err := c.validateTimeRange(&req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关监控图表数据查询时间范围校验失败", "error", err)
		response.ErrorJSON(ctx, err.Error(), constants.ED00007)
		return
	}

	// 设置默认热点路由限制
	if req.HotRouteLimit <= 0 {
		req.HotRouteLimit = 10
	}

	// 并发查询各种监控数据
	requestTrend, err := c.clickhouseMonitoringDAO.GetRequestMetricsTrend(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取ClickHouse请求趋势数据失败", "error", err)
		response.ErrorJSON(ctx, "获取请求趋势数据失败: "+err.Error(), constants.ED00009)
		return
	}

	responseTimeTrend, err := c.clickhouseMonitoringDAO.GetResponseTimeMetricsTrend(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取ClickHouse响应时间趋势数据失败", "error", err)
		response.ErrorJSON(ctx, "获取响应时间趋势数据失败: "+err.Error(), constants.ED00009)
		return
	}

	statusCodeDistribution, err := c.clickhouseMonitoringDAO.GetStatusCodeDistribution(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取ClickHouse状态码分布数据失败", "error", err)
		response.ErrorJSON(ctx, "获取状态码分布数据失败: "+err.Error(), constants.ED00009)
		return
	}

	hotRoutes, err := c.clickhouseMonitoringDAO.GetHotRoutes(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取ClickHouse热点路由数据失败", "error", err)
		response.ErrorJSON(ctx, "获取热点路由数据失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建图表数据
	chartData := &models.GatewayMonitoringChartData{
		RequestTrend:            requestTrend,
		ResponseTimeTrend:       responseTimeTrend,
		StatusCodeDistribution: statusCodeDistribution,
		HotRoutes:               hotRoutes,
	}

	response.SuccessJSON(ctx, chartData, constants.SD00002)
}

// QueryGatewayLogs 查询网关日志列表（ClickHouse版本）
// @Summary 查询网关日志列表（ClickHouse版本）
// @Description 支持分页查询和多条件过滤的ClickHouse网关日志列表，不返回大字段以提高查询性能
// @Tags ClickHouse网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param query body models.GatewayAccessLogQueryRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0023/clickhouse-gateway-log/query [post]
func (c *ClickHouseQueryController) QueryGatewayLogs(ctx *gin.Context) {
	// 解析查询参数
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志查询参数解析失败", "error", err)
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
	logs, total, err := c.clickhouseQueryDAO.QueryGatewayLogs(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志查询失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建分页信息
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, int(total))

	// 使用统一的分页响应
	response.PageJSON(ctx, logs, pageInfo, constants.SD00002)
}

// GetGatewayLog 获取网关日志详情（ClickHouse版本）
// @Summary 获取网关日志详情（ClickHouse版本）
// @Description 通过租户ID和链路追踪ID组合主键获取ClickHouse网关日志详情
// @Tags ClickHouse网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param get body models.GatewayAccessLogGetRequest true "获取参数"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0023/clickhouse-gateway-log/get [post]
func (c *ClickHouseQueryController) GetGatewayLog(ctx *gin.Context) {
	// 解析获取参数
	var req models.GatewayAccessLogGetRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志获取参数解析失败", "error", err)
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

	// 调用DAO获取详情
	log, err := c.clickhouseQueryDAO.GetGatewayLogByKey(ctx, req.TenantId, req.TraceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志获取失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), constants.ED00008)
		return
	}

	response.SuccessJSON(ctx, log, constants.SD00002)
}

// CountGatewayLogs 统计网关日志数量（ClickHouse版本）
// @Summary 统计网关日志数量（ClickHouse版本）
// @Description 根据查询条件统计ClickHouse网关日志数量
// @Tags ClickHouse网关日志
// @Accept json
// @Accept x-www-form-urlencoded
// @Produce json
// @Param count body models.GatewayAccessLogQueryRequest true "统计参数"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0023/clickhouse-gateway-log/count [post]
func (c *ClickHouseQueryController) CountGatewayLogs(ctx *gin.Context) {
	// 解析统计参数
	var req models.GatewayAccessLogQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志统计参数解析失败", "error", err)
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

	// 调用DAO统计
	count, err := c.clickhouseQueryDAO.CountGatewayLogs(ctx, &req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "ClickHouse网关日志统计失败", "error", err)
		response.ErrorJSON(ctx, "统计失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"count": count,
		"table": models.GatewayAccessLog{}.TableName(),
	}, constants.SD00002)
}

// validateTimeRange 校验时间范围
// 要求：开始时间和结束时间必填，时间范围不能超过24小时
func (c *ClickHouseQueryController) validateTimeRange(req *models.GatewayMonitoringQueryRequest) error {
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

	return nil
} 