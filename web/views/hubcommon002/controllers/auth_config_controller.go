// 认证配置控制器文件
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

// AuthConfigController 认证配置控制器
type AuthConfigController struct {
	dao *dao.AuthConfigDAO
}

// NewAuthConfigController 创建认证配置控制器
func NewAuthConfigController(db database.Database) *AuthConfigController {
	return &AuthConfigController{
		dao: dao.NewAuthConfigDAO(db),
	}
}

// AddAuthConfig 添加认证配置
func (c *AuthConfigController) AddAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始添加认证配置", "controller", "AuthConfigController", "action", "AddAuthConfig")

	// 绑定请求参数
	var config models.AuthConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.AuthName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID
	config.TenantId = tenantId

	// 调用DAO层添加认证配置
	err := c.dao.AddAuthConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "添加认证配置失败", "error", err.Error(),
			"authConfigId", config.AuthConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "添加认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	newConfig, err := c.dao.GetAuthConfig(config.TenantId, config.AuthConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "添加成功但获取最新数据失败", "error", err.Error(),
			"authConfigId", config.AuthConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "认证配置添加成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "认证配置添加成功", "authConfigId", config.AuthConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetAuthConfig 获取认证配置详情（使用主键 authConfigId）
func (c *AuthConfigController) GetAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取认证配置", "controller", "AuthConfigController", "action", "GetAuthConfig")

	// 获取主键参数（DAO层会校验）
	authConfigId := request.GetParam(ctx, "authConfigId")
	if authConfigId == "" {
		response.ErrorJSON(ctx, "authConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层获取认证配置（使用主键）
	config, err := c.dao.GetAuthConfig(tenantId, authConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取认证配置失败", "error", err.Error(),
			"authConfigId", authConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "获取认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "认证配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "获取认证配置成功", "authConfigId", authConfigId, "tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// UpdateAuthConfig 更新认证配置
func (c *AuthConfigController) UpdateAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新认证配置", "controller", "AuthConfigController", "action", "UpdateAuthConfig")

	// 绑定请求参数
	var config models.AuthConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if config.AuthConfigId == "" {
		response.ErrorJSON(ctx, "authConfigId不能为空", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 调用DAO层更新认证配置（使用主键 authConfigId）
	err := c.dao.UpdateAuthConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新认证配置失败", "error", err.Error(),
			"authConfigId", config.AuthConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端（使用主键）
	updatedConfig, err := c.dao.GetAuthConfig(tenantId, config.AuthConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(),
			"authConfigId", config.AuthConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "认证配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "认证配置更新成功", "authConfigId", config.AuthConfigId,
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteAuthConfig 删除认证配置
func (c *AuthConfigController) DeleteAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除认证配置", "controller", "AuthConfigController", "action", "DeleteAuthConfig")

	// 获取主键参数（DAO层会校验）
	authConfigId := request.GetParam(ctx, "authConfigId")
	if authConfigId == "" {
		response.ErrorJSON(ctx, "authConfigId不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID（前置校验已处理）
	tenantId := request.GetTenantID(ctx)

	// 调用DAO层删除认证配置（使用主键）
	err := c.dao.DeleteAuthConfig(tenantId, authConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除认证配置失败", "error", err.Error(),
			"authConfigId", authConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "认证配置删除成功", "authConfigId", authConfigId,
		"tenantId", tenantId)
	response.SuccessJSON(ctx, gin.H{"message": "认证配置删除成功"}, constants.SD00003)
}

// QueryAuthConfigs 查询认证配置（根据实例ID或路由ID查询单个配置）
// @Summary 获取认证配置
// @Description 根据gatewayInstanceId或routeConfigId查询单个认证配置，不需要分页
// @Tags 认证配置
// @Produce json
// @Param gatewayInstanceId query string false "网关实例ID（实例级认证，与routeConfigId二选一）"
// @Param routeConfigId query string false "路由配置ID（路由级认证，与gatewayInstanceId二选一）"
// @Success 200 {object} response.JsonData
// @Router /api/hubcommon002/auth/query [post]
func (c *AuthConfigController) QueryAuthConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询认证配置", "controller", "AuthConfigController", "action", "QueryAuthConfigs")

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.AuthConfigQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定认证配置查询条件失败，使用默认条件", "error", err.Error())
	}

	// 验证必填条件：gatewayInstanceId 或 routeConfigId（避免关联错误）
	if query.GatewayInstanceId == "" && query.RouteConfigId == "" {
		response.ErrorJSON(ctx, "gatewayInstanceId或routeConfigId不能同时为空", constants.ED00007)
		return
	}

	var config *models.AuthConfig
	var err error

	// 按网关实例ID查询单个配置
	if query.GatewayInstanceId != "" {
		config, err = c.dao.GetAuthConfigByGatewayInstance(tenantId, query.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例认证配置失败", "error", err.Error(),
				"tenantId", tenantId, "gatewayInstanceId", query.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例认证配置失败: "+err.Error(), constants.ED00009)
			return
		}
	} else if query.RouteConfigId != "" {
		// 按路由配置ID查询单个配置
		config, err = c.dao.GetAuthConfigByRouteConfig(tenantId, query.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置认证配置失败", "error", err.Error(),
				"tenantId", tenantId, "routeConfigId", query.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置认证配置失败: "+err.Error(), constants.ED00009)
			return
		}
	}

	// 没有数据返回空，不报错
	if config == nil {
		logger.InfoWithTrace(ctx, "查询认证配置为空", "tenantId", tenantId)
		response.SuccessJSON(ctx, nil, constants.SD00002)
		return
	}

	logger.InfoWithTrace(ctx, "查询认证配置成功", "tenantId", tenantId,
		"authConfigId", config.AuthConfigId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}
