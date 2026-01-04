package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hubcommon002/dao"
	"gateway/web/views/hubcommon002/models"

	"github.com/gin-gonic/gin"
)

// RateLimitConfigController 限流配置控制器
type RateLimitConfigController struct {
	dao *dao.RateLimitConfigDAO
}

// NewRateLimitConfigController 创建限流配置控制器
func NewRateLimitConfigController(db database.Database) *RateLimitConfigController {
	return &RateLimitConfigController{
		dao: dao.NewRateLimitConfigDAO(db),
	}
}

// AddRateLimitConfig 添加限流配置
func (c *RateLimitConfigController) AddRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加限流配置", "controller", "RateLimitConfigController", "action", "AddRateLimitConfig")

	// 绑定请求参数
	var config models.RateLimitConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.LimitName == "" {
		response.ErrorJSON(ctx, "限流规则名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层添加限流配置
	err := c.dao.AddRateLimitConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加限流配置失败", "error", err.Error(),
			"rateLimitConfigId", config.RateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetRateLimitConfig(config.TenantId, config.RateLimitConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"rateLimitConfigId", config.RateLimitConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "限流配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "限流配置添加成功", "rateLimitConfigId", config.RateLimitConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetRateLimitConfig 获取限流配置详情（使用主键 rateLimitConfigId）
func (c *RateLimitConfigController) GetRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取限流配置", "controller", "RateLimitConfigController", "action", "GetRateLimitConfig")

	// 获取主键参数（DAO层会校验）
	rateLimitConfigId := request.GetParam(ctx, "rateLimitConfigId")
	if rateLimitConfigId == "" {
		response.ErrorJSON(ctx, "rateLimitConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取限流配置（使用主键）
	config, err := c.dao.GetRateLimitConfig(tenantId, rateLimitConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取限流配置失败", "error", err.Error(),
			"rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "限流配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取限流配置成功", "rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateRateLimitConfig 更新限流配置
func (c *RateLimitConfigController) UpdateRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新限流配置", "controller", "RateLimitConfigController", "action", "UpdateRateLimitConfig")

	// 绑定请求参数
	var config models.RateLimitConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.RateLimitConfigId == "" {
		response.ErrorJSON(ctx, "rateLimitConfigId不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新限流配置（使用主键 rateLimitConfigId）
	err := c.dao.UpdateRateLimitConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新限流配置失败", "error", err.Error(),
			"rateLimitConfigId", config.RateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetRateLimitConfig(tenantId, config.RateLimitConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"rateLimitConfigId", config.RateLimitConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "限流配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "限流配置更新成功", "rateLimitConfigId", config.RateLimitConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteRateLimitConfig 删除限流配置
func (c *RateLimitConfigController) DeleteRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除限流配置", "controller", "RateLimitConfigController", "action", "DeleteRateLimitConfig")

	// 获取主键参数（DAO层会校验）
	rateLimitConfigId := request.GetParam(ctx, "rateLimitConfigId")
	if rateLimitConfigId == "" {
		response.ErrorJSON(ctx, "rateLimitConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层删除限流配置（使用主键）
	err := c.dao.DeleteRateLimitConfig(tenantId, rateLimitConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除限流配置失败", "error", err.Error(),
			"rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "限流配置删除成功", "rateLimitConfigId", rateLimitConfigId,
		"tenantId", tenantId)
	response.SuccessJSON(ctx, gin.H{"message": "限流配置删除成功"}, constants.SD00003)
}

// QueryRateLimitConfigs 查询限流配置（根据实例ID或路由ID查询单个配置）
// @Summary 获取限流配置
// @Description 根据gatewayInstanceId或routeConfigId查询单个限流配置，不需要分页
// @Tags 限流配置
// @Produce json
// @Param gatewayInstanceId query string false "网关实例ID（实例级限流，与routeConfigId二选一）"
// @Param routeConfigId query string false "路由配置ID（路由级限流，与gatewayInstanceId二选一）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/rate-limit/query [post]
func (c *RateLimitConfigController) QueryRateLimitConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询限流配置", "controller", "RateLimitConfigController", "action", "QueryRateLimitConfigs")

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.RateLimitConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定限流配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：gatewayInstanceId 或 routeConfigId（避免关联错误）
	if query.GatewayInstanceId == "" && query.RouteConfigId == "" {
		response.ErrorJSON(ctx, "gatewayInstanceId或routeConfigId不能同时为空", constants.ED00007)
		return
	}

	var config *models.RateLimitConfig
	var err error

	// 按网关实例ID查询单个配置
	if query.GatewayInstanceId != "" {
		config, err = c.dao.GetRateLimitConfigByGatewayInstance(tenantId, query.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例限流配置失败", "error", err.Error(),
				"tenantId", tenantId, "gatewayInstanceId", query.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例限流配置失败: "+err.Error(), constants.ED00009)
			return
		}
	} else if query.RouteConfigId != "" {
		// 按路由配置ID查询单个配置
		config, err = c.dao.GetRateLimitConfigByRouteConfig(tenantId, query.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置限流配置失败", "error", err.Error(),
				"tenantId", tenantId, "routeConfigId", query.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置限流配置失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 没有数据返回空，不报错
	if config == nil {
		logger.InfoWithTrace(ctx, "查询限流配置为空", "tenantId", tenantId)
		response.SuccessJSON(ctx, nil, constants.SD00002)
		return
	}

	logger.InfoWithTrace(ctx, "查询限流配置成功", "tenantId", tenantId,
		"rateLimitConfigId", config.RateLimitConfigId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}
