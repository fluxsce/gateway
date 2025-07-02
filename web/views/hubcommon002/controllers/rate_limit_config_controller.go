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

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

	// 验证必填字段
	if config.LimitName == "" {
		response.ErrorJSON(ctx, "限流规则名称不能为空", constants.ED00007)
		return
	}
	if config.LimitRate <= 0 {
		response.ErrorJSON(ctx, "限流速率必须大于0", constants.ED00007)
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

	// 查询最新的配置数据返回给前端
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

// GetRateLimitConfig 获取限流配置详情（支持多种查询方式）
func (c *RateLimitConfigController) GetRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取限流配置", "controller", "RateLimitConfigController", "action", "GetRateLimitConfig")

	// 定义请求参数结构，支持多种查询方式
	var req struct {
		// 按配置ID查询
		RateLimitConfigId *string `json:"rateLimitConfigId" form:"rateLimitConfigId"`
		
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
		"rateLimitConfigId", func() string {
			if req.RateLimitConfigId != nil {
				return *req.RateLimitConfigId
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
	if req.RateLimitConfigId != nil && *req.RateLimitConfigId != "" {
		hasValidParam = true
	}
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		hasValidParam = true
	}
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		hasValidParam = true
	}
	
	if !hasValidParam {
		response.ErrorJSON(ctx, "请提供rateLimitConfigId、gatewayInstanceId或routeConfigId中的任意一个", constants.ED00007)
		return
	}

	// 按配置ID查询单个配置
	if req.RateLimitConfigId != nil && *req.RateLimitConfigId != "" {
		config, err := c.dao.GetRateLimitConfig(tenantId, *req.RateLimitConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取限流配置失败", "error", err.Error(), 
				"rateLimitConfigId", *req.RateLimitConfigId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "获取限流配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "限流配置不存在", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "获取限流配置成功", "rateLimitConfigId", *req.RateLimitConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 按网关实例ID查询单个配置
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		config, err := c.dao.GetRateLimitConfigByGatewayInstance(tenantId, *req.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例限流配置失败", "error", err.Error(), 
				"tenantId", tenantId, "gatewayInstanceId", *req.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例限流配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该网关实例未配置限流", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询网关实例限流配置成功", "tenantId", tenantId, 
			"gatewayInstanceId", *req.GatewayInstanceId, "rateLimitConfigId", config.RateLimitConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 按路由配置ID查询单个配置
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		config, err := c.dao.GetRateLimitConfigByRouteConfig(tenantId, *req.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置限流配置失败", "error", err.Error(), 
				"tenantId", tenantId, "routeConfigId", *req.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置限流配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该路由未配置限流", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询路由配置限流配置成功", "tenantId", tenantId, 
			"routeConfigId", *req.RouteConfigId, "rateLimitConfigId", config.RateLimitConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 如果到这里说明参数有问题
	response.ErrorJSON(ctx, "查询参数无效", constants.ED00007)
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

	// 从URL路径获取限流配置ID
	rateLimitConfigId := request.GetParam(ctx, "rateLimitConfigId")
	if rateLimitConfigId == "" {
		response.ErrorJSON(ctx, "限流配置ID不能为空", constants.ED00007)
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

	// 设置从URL获取的限流配置ID和租户ID
	config.RateLimitConfigId = rateLimitConfigId
	config.TenantId = tenantId

	// 调用DAO层更新限流配置
	err := c.dao.UpdateRateLimitConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新限流配置失败", "error", err.Error(), 
			"rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetRateLimitConfig(tenantId, rateLimitConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "限流配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "限流配置更新成功", "rateLimitConfigId", rateLimitConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteRateLimitConfig 删除限流配置
func (c *RateLimitConfigController) DeleteRateLimitConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除限流配置", "controller", "RateLimitConfigController", "action", "DeleteRateLimitConfig")

	// 获取限流配置ID参数
	rateLimitConfigId := request.GetParam(ctx, "rateLimitConfigId")
	if rateLimitConfigId == "" {
		response.ErrorJSON(ctx, "限流配置ID不能为空", constants.ED00007)
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

	// 调用DAO层删除限流配置
	err := c.dao.DeleteRateLimitConfig(tenantId, rateLimitConfigId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除限流配置失败", "error", err.Error(), 
			"rateLimitConfigId", rateLimitConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除限流配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "限流配置删除成功", "rateLimitConfigId", rateLimitConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "限流配置删除成功"}, constants.SD00003)
}

// QueryRateLimitConfigs 查询限流配置列表
func (c *RateLimitConfigController) QueryRateLimitConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询限流配置列表", "controller", "RateLimitConfigController", "action", "QueryRateLimitConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询限流配置列表
	configs, total, err := c.dao.ListRateLimitConfigs(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询限流配置列表失败", "error", err.Error(), 
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询限流配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应
	result := gin.H{
		"configs":  configs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	logger.InfoWithTrace(ctx, "查询限流配置列表成功", "tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, result, constants.SD00002)
} 