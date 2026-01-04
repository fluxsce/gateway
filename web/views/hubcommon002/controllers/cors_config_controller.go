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

// CorsConfigController CORS配置控制器
type CorsConfigController struct {
	dao *dao.CorsConfigDAO
}

// NewCorsConfigController 创建CORS配置控制器
func NewCorsConfigController(db database.Database) *CorsConfigController {
	return &CorsConfigController{
		dao: dao.NewCorsConfigDAO(db),
	}
}

// AddCorsConfig 添加CORS配置
func (c *CorsConfigController) AddCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加CORS配置", "controller", "CorsConfigController", "action", "AddCorsConfig")

	// 绑定请求参数
	var config models.CorsConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层添加CORS配置
	err := c.dao.AddCorsConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加CORS配置失败", "error", err.Error(),
			"corsConfigId", config.CorsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetCorsConfig(config.TenantId, config.CorsConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"corsConfigId", config.CorsConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "CORS配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "CORS配置添加成功", "corsConfigId", config.CorsConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetCorsConfig 获取CORS配置详情（使用主键 corsConfigId）
func (c *CorsConfigController) GetCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取CORS配置", "controller", "CorsConfigController", "action", "GetCorsConfig")

	// 获取主键参数（DAO层会校验）
	corsConfigId := request.GetParam(ctx, "corsConfigId")
	if corsConfigId == "" {
		response.ErrorJSON(ctx, "corsConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取CORS配置（使用主键）
	config, err := c.dao.GetCorsConfig(tenantId, corsConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取CORS配置失败", "error", err.Error(),
			"corsConfigId", corsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "CORS配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取CORS配置成功", "corsConfigId", corsConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateCorsConfig 更新CORS配置
func (c *CorsConfigController) UpdateCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新CORS配置", "controller", "CorsConfigController", "action", "UpdateCorsConfig")

	// 绑定请求参数
	var config models.CorsConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.CorsConfigId == "" {
		response.ErrorJSON(ctx, "corsConfigId不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新CORS配置（使用主键 corsConfigId）
	err := c.dao.UpdateCorsConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新CORS配置失败", "error", err.Error(),
			"corsConfigId", config.CorsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetCorsConfig(tenantId, config.CorsConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"corsConfigId", config.CorsConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "CORS配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "CORS配置更新成功", "corsConfigId", config.CorsConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteCorsConfig 删除CORS配置
func (c *CorsConfigController) DeleteCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除CORS配置", "controller", "CorsConfigController", "action", "DeleteCorsConfig")

	// 获取主键参数（DAO层会校验）
	corsConfigId := request.GetParam(ctx, "corsConfigId")
	if corsConfigId == "" {
		response.ErrorJSON(ctx, "corsConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层删除CORS配置（使用主键）
	err := c.dao.DeleteCorsConfig(tenantId, corsConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除CORS配置失败", "error", err.Error(),
			"corsConfigId", corsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "CORS配置删除成功", "corsConfigId", corsConfigId,
		"tenantId", tenantId)
	response.SuccessJSON(ctx, gin.H{"message": "CORS配置删除成功"}, constants.SD00003)
}

// QueryCorsConfigs 查询CORS配置（根据实例ID或路由ID查询单个配置）
// @Summary 获取CORS配置
// @Description 根据gatewayInstanceId或routeConfigId查询单个CORS配置，不需要分页
// @Tags CORS配置
// @Produce json
// @Param gatewayInstanceId query string false "网关实例ID（实例级CORS，与routeConfigId二选一）"
// @Param routeConfigId query string false "路由配置ID（路由级CORS，与gatewayInstanceId二选一）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/cors/query [post]
func (c *CorsConfigController) QueryCorsConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询CORS配置", "controller", "CorsConfigController", "action", "QueryCorsConfigs")

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.CorsConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定CORS配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：gatewayInstanceId 或 routeConfigId（避免关联错误）
	if query.GatewayInstanceId == "" && query.RouteConfigId == "" {
		response.ErrorJSON(ctx, "gatewayInstanceId或routeConfigId不能同时为空", constants.ED00007)
		return
	}

	var config *models.CorsConfig
	var err error

	// 按网关实例ID查询单个配置
	if query.GatewayInstanceId != "" {
		config, err = c.dao.GetCorsConfigByGatewayInstance(tenantId, query.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例CORS配置失败", "error", err.Error(),
				"tenantId", tenantId, "gatewayInstanceId", query.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例CORS配置失败: "+err.Error(), constants.ED00009)
			return
		}
	} else if query.RouteConfigId != "" {
		// 按路由配置ID查询单个配置
		config, err = c.dao.GetCorsConfigByRouteConfig(tenantId, query.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置CORS配置失败", "error", err.Error(),
				"tenantId", tenantId, "routeConfigId", query.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置CORS配置失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 没有数据返回空，不报错
	if config == nil {
		logger.InfoWithTrace(ctx, "查询CORS配置为空", "tenantId", tenantId)
		response.SuccessJSON(ctx, nil, constants.SD00002)
		return
	}

	logger.InfoWithTrace(ctx, "查询CORS配置成功", "tenantId", tenantId,
		"corsConfigId", config.CorsConfigId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}
