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

// SecurityConfigController 安全配置控制器
type SecurityConfigController struct {
	dao *dao.SecurityConfigDAO
}

// NewSecurityConfigController 创建安全配置控制器
func NewSecurityConfigController(db database.Database) *SecurityConfigController {
	return &SecurityConfigController{
		dao: dao.NewSecurityConfigDAO(db),
	}
}

// AddSecurityConfig 添加安全配置
func (c *SecurityConfigController) AddSecurityConfig(ctx *gin.Context) {
	// 获取上下文用于跟踪
	reqCtx := ctx
	
	logger.InfoWithTrace(reqCtx, "开始添加安全配置", 
		"controller", "SecurityConfigController", 
		"action", "AddSecurityConfig")

	// 绑定请求参数
	var config models.SecurityConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		logger.ErrorWithTrace(reqCtx, "参数绑定失败", err,
			"controller", "SecurityConfigController",
			"action", "AddSecurityConfig")
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId
	config.SecurityConfigId = "" // 清空ID，让DAO自动生成

	// 验证必填字段
	if config.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}

	// 调用DAO层添加安全配置
	securityConfigId, err := c.dao.AddSecurityConfig(reqCtx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(reqCtx, "添加安全配置失败", err,
			"tenantId", tenantId, 
			"operatorId", operatorId)
		response.ErrorJSON(ctx, "添加安全配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	newConfig, err := c.dao.GetSecurityConfigById(reqCtx, securityConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(reqCtx, "添加成功但获取最新数据失败", err,
			"securityConfigId", securityConfigId, "tenantId", tenantId)
		// 添加成功但获取最新数据失败，仍然返回成功
		result := map[string]interface{}{
			"securityConfigId": securityConfigId,
			"tenantId":         tenantId,
			"message":          "安全配置添加成功",
		}
		response.SuccessJSON(ctx, result, constants.SD00003)
		return
	}

	logger.InfoWithTrace(reqCtx, "安全配置添加成功", 
		"securityConfigId", securityConfigId, 
		"tenantId", tenantId, 
		"operatorId", operatorId)
	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// GetSecurityConfig 获取安全配置详情
func (c *SecurityConfigController) GetSecurityConfig(ctx *gin.Context) {
	// 获取上下文用于跟踪
	reqCtx := ctx
	
	logger.InfoWithTrace(reqCtx, "开始获取安全配置详情", 
		"controller", "SecurityConfigController", 
		"action", "GetSecurityConfig")

	// 获取安全配置ID参数（支持多种数据源）
	securityConfigId := request.GetParam(ctx, "securityConfigId")
	if securityConfigId == "" {
		logger.WarnWithTrace(reqCtx, "安全配置ID参数缺失",
			"controller", "SecurityConfigController",
			"action", "GetSecurityConfig")
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		logger.WarnWithTrace(reqCtx, "无法获取租户信息",
			"controller", "SecurityConfigController",
			"action", "GetSecurityConfig")
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层获取安全配置详情
	config, err := c.dao.GetSecurityConfigById(reqCtx, securityConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(reqCtx, "获取安全配置详情失败", err,
			"securityConfigId", securityConfigId, 
			"tenantId", tenantId)
		response.ErrorJSON(ctx, "获取安全配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		logger.WarnWithTrace(reqCtx, "安全配置不存在", 
			"securityConfigId", securityConfigId, 
			"tenantId", tenantId)
		response.ErrorJSON(ctx, "安全配置不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(reqCtx, "获取安全配置详情成功", 
		"securityConfigId", securityConfigId, 
		"tenantId", tenantId)
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// EditSecurityConfig 编辑安全配置
func (c *SecurityConfigController) EditSecurityConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始编辑安全配置", "controller", "SecurityConfigController", "action", "EditSecurityConfig")

	// 绑定请求参数
	var config models.SecurityConfig
	if err := request.BindSafely(ctx, &config); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 强制使用上下文中的租户ID，防止前端恶意修改
	config.TenantId = tenantId

	// 验证必填字段
	if config.SecurityConfigId == "" {
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
		return
	}
	if config.ConfigName == "" {
		response.ErrorJSON(ctx, "配置名称不能为空", constants.ED00007)
		return
	}

	// 调用DAO层更新安全配置
	err := c.dao.UpdateSecurityConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "编辑安全配置失败", err,
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId, "operatorId", operatorId)
		response.ErrorJSON(ctx, "编辑安全配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetSecurityConfigById(ctx, config.SecurityConfigId, tenantId)
	if err != nil {
		logger.WarnWithTrace(ctx, "编辑成功但获取最新数据失败", err,
			"securityConfigId", config.SecurityConfigId, "tenantId", tenantId)
		// 编辑成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "安全配置编辑成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "安全配置编辑成功", "securityConfigId", config.SecurityConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteSecurityConfig 删除安全配置
func (c *SecurityConfigController) DeleteSecurityConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除安全配置", "controller", "SecurityConfigController", "action", "DeleteSecurityConfig")

	// 获取安全配置ID参数（支持多种数据源）
	securityConfigId := request.GetParam(ctx, "securityConfigId")
	if securityConfigId == "" {
		response.ErrorJSON(ctx, "安全配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO层删除安全配置
	err := c.dao.DeleteSecurityConfig(ctx, securityConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除安全配置失败", err,
			"securityConfigId", securityConfigId, "tenantId", tenantId, "operatorId", operatorId)
		response.ErrorJSON(ctx, "删除安全配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "安全配置删除成功", "securityConfigId", securityConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "安全配置删除成功"}, constants.SD00003)
}

// QuerySecurityConfigs 分页查询安全配置列表
func (c *SecurityConfigController) QuerySecurityConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询安全配置列表", "controller", "SecurityConfigController", "action", "QuerySecurityConfigs")

	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	
	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询安全配置列表
	configs, total, err := c.dao.ListSecurityConfigs(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询安全配置列表失败", err, "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询安全配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "securityConfigId"

	logger.InfoWithTrace(ctx, "查询安全配置列表成功", "tenantId", tenantId, 
		"total", total, "page", page, "pageSize", pageSize)
	response.PageJSON(ctx, configs, pageInfo, constants.SD00002)
}

// QuerySecurityConfigsByGatewayInstance 根据网关实例查询安全配置列表
func (c *SecurityConfigController) QuerySecurityConfigsByGatewayInstance(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询网关实例安全配置", "controller", "SecurityConfigController", "action", "QuerySecurityConfigsByGatewayInstance")

	// 获取网关实例ID参数（支持多种数据源）
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询网关实例安全配置列表
	configs, err := c.dao.ListSecurityConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询网关实例安全配置列表失败", err,
			"gatewayInstanceId", gatewayInstanceId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询网关实例安全配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "查询网关实例安全配置列表成功", "gatewayInstanceId", gatewayInstanceId, 
		"tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, configs, constants.SD00002)
}

// QuerySecurityConfigsByRouteConfig 根据路由配置查询安全配置列表
func (c *SecurityConfigController) QuerySecurityConfigsByRouteConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询路由配置安全配置", "controller", "SecurityConfigController", "action", "QuerySecurityConfigsByRouteConfig")

	// 获取路由配置ID参数（支持多种数据源）
	routeConfigId := request.GetParam(ctx, "routeConfigId")
	if routeConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询路由配置安全配置列表
	configs, err := c.dao.ListSecurityConfigsByRouteConfig(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询路由配置安全配置列表失败", err,
			"routeConfigId", routeConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "查询路由配置安全配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "查询路由配置安全配置列表成功", "routeConfigId", routeConfigId, 
		"tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, configs, constants.SD00002)
} 