// 认证配置控制器文件
package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hubcommon002/dao"
	"gohub/web/views/hubcommon002/models"

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
	if err := request.Bind(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if config.AuthName == "" || config.AuthType == "" || config.AuthConfig == "" {
		response.ErrorJSON(ctx, "认证配置名称、认证类型和认证配置不能为空", constants.ED00007)
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

	// 查询最新的配置数据返回给前端
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

// GetAuthConfig 获取认证配置详情（支持多种查询方式）
func (c *AuthConfigController) GetAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取认证配置", "controller", "AuthConfigController", "action", "GetAuthConfig")

	// 定义请求参数结构，支持多种查询方式
	var req struct {
		// 按配置ID查询
		AuthConfigId *string `json:"authConfigId" form:"authConfigId"`
		
		// 按网关实例ID查询
		GatewayInstanceId *string `json:"gatewayInstanceId" form:"gatewayInstanceId"`
		
		// 按路由配置ID查询
		RouteConfigId *string `json:"routeConfigId" form:"routeConfigId"`
		
		// 分页参数（用于按实例或路由查询时）
		Page     int `json:"page" form:"page"`
		PageSize int `json:"pageSize" form:"pageSize"`
	}

	// 绑定请求参数 - 支持JSON和表单数据
	if err := ctx.ShouldBind(&req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 调试日志：输出接收到的参数
	logger.InfoWithTrace(ctx, "接收到的查询参数", 
		"authConfigId", func() string {
			if req.AuthConfigId != nil {
				return *req.AuthConfigId
			}
			return "nil"
		}(),
		"gatewayInstanceId", func() string {
			if req.GatewayInstanceId != nil {
				return *req.GatewayInstanceId
			}
			return "nil"
		}(),
		"routeConfigId", func() string {
			if req.RouteConfigId != nil {
				return *req.RouteConfigId
			}
			return "nil"
		}(),
		"page", req.Page,
		"pageSize", req.PageSize)

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 验证查询参数（至少提供一种查询方式）
	hasValidParam := false
	if req.AuthConfigId != nil && *req.AuthConfigId != "" {
		hasValidParam = true
	}
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		hasValidParam = true
	}
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		hasValidParam = true
	}
	
	if !hasValidParam {
		response.ErrorJSON(ctx, "请提供authConfigId、gatewayInstanceId或routeConfigId中的任意一个", constants.ED00007)
		return
	}

	// 按配置ID查询单个配置
	if req.AuthConfigId != nil && *req.AuthConfigId != "" {
		config, err := c.dao.GetAuthConfig(tenantId, *req.AuthConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取认证配置失败", "error", err.Error(), 
				"authConfigId", *req.AuthConfigId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "获取认证配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "认证配置不存在", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "获取认证配置成功", "authConfigId", *req.AuthConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 按网关实例ID查询单个配置
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		config, err := c.dao.GetAuthConfigByGatewayInstance(tenantId, *req.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例认证配置失败", "error", err.Error(), 
				"tenantId", tenantId, "gatewayInstanceId", *req.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例认证配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该网关实例未配置认证", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询网关实例认证配置成功", "tenantId", tenantId, 
			"gatewayInstanceId", *req.GatewayInstanceId, "authConfigId", config.AuthConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 按路由配置ID查询单个配置
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		config, err := c.dao.GetAuthConfigByRouteConfig(tenantId, *req.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置认证配置失败", "error", err.Error(), 
				"tenantId", tenantId, "routeConfigId", *req.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置认证配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该路由未配置认证", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询路由配置认证配置成功", "tenantId", tenantId, 
			"routeConfigId", *req.RouteConfigId, "authConfigId", config.AuthConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 如果到这里说明参数有问题
	response.ErrorJSON(ctx, "查询参数无效", constants.ED00007)
}

// UpdateAuthConfig 更新认证配置
func (c *AuthConfigController) UpdateAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始更新认证配置", "controller", "AuthConfigController", "action", "UpdateAuthConfig")

	// 绑定请求参数
	var config models.AuthConfig
	if err := request.Bind(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从URL路径获取认证配置ID
	authConfigId := request.GetParam(ctx, "authConfigId")
	if authConfigId == "" {
		response.ErrorJSON(ctx, "认证配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 设置从URL获取的认证配置ID和租户ID
	config.AuthConfigId = authConfigId
	config.TenantId = tenantId

	// 调用DAO层更新认证配置
	err := c.dao.UpdateAuthConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新认证配置失败", "error", err.Error(), 
			"authConfigId", authConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetAuthConfig(tenantId, authConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"authConfigId", authConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "认证配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "认证配置更新成功", "authConfigId", authConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteAuthConfig 删除认证配置
func (c *AuthConfigController) DeleteAuthConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除认证配置", "controller", "AuthConfigController", "action", "DeleteAuthConfig")

	// 获取认证配置ID参数
	authConfigId := request.GetParam(ctx, "authConfigId")
	if authConfigId == "" {
		response.ErrorJSON(ctx, "认证配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 调用DAO层删除认证配置
	err := c.dao.DeleteAuthConfig(tenantId, authConfigId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除认证配置失败", "error", err.Error(), 
			"authConfigId", authConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除认证配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "认证配置删除成功", "authConfigId", authConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "认证配置删除成功"}, constants.SD00003)
}

// QueryAuthConfigs 查询认证配置列表
func (c *AuthConfigController) QueryAuthConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询认证配置列表", "controller", "AuthConfigController", "action", "QueryAuthConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询认证配置列表
	configs, total, err := c.dao.ListAuthConfigs(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询认证配置列表失败", "error", err.Error(), 
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询认证配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应
	result := gin.H{
		"configs":  configs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	logger.InfoWithTrace(ctx, "查询认证配置列表成功", "tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, result, constants.SD00002)
} 