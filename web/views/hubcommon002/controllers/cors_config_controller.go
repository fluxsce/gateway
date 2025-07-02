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

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必要信息
	if tenantId == "" || operatorId == "" {
		response.ErrorJSON(ctx, "无法获取租户或操作人信息", constants.ED00007)
		return
	}

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

	// 查询最新的配置数据返回给前端
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

// GetCorsConfig 获取CORS配置详情（支持多种查询方式）
func (c *CorsConfigController) GetCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始获取CORS配置", "controller", "CorsConfigController", "action", "GetCorsConfig")

	// 定义请求参数结构，支持多种查询方式
	var req struct {
		// 按配置ID查询
		CorsConfigId *string `json:"corsConfigId" form:"corsConfigId"`
		
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
		"corsConfigId", func() string {
			if req.CorsConfigId != nil {
				return *req.CorsConfigId
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
	if req.CorsConfigId != nil && *req.CorsConfigId != "" {
		hasValidParam = true
	}
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		hasValidParam = true
	}
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		hasValidParam = true
	}
	
	if !hasValidParam {
		response.ErrorJSON(ctx, "请提供corsConfigId、gatewayInstanceId或routeConfigId中的任意一个", constants.ED00007)
		return
	}

	// 按配置ID查询单个配置
	if req.CorsConfigId != nil && *req.CorsConfigId != "" {
		config, err := c.dao.GetCorsConfig(tenantId, *req.CorsConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "获取CORS配置失败", "error", err.Error(), 
				"corsConfigId", *req.CorsConfigId, "tenantId", tenantId)
			response.ErrorJSON(ctx, "获取CORS配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "CORS配置不存在", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "获取CORS配置成功", "corsConfigId", *req.CorsConfigId, "tenantId", tenantId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 设置分页默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 按网关实例ID查询单个配置
	if req.GatewayInstanceId != nil && *req.GatewayInstanceId != "" {
		config, err := c.dao.GetCorsConfigByGatewayInstance(tenantId, *req.GatewayInstanceId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询网关实例CORS配置失败", "error", err.Error(), 
				"tenantId", tenantId, "gatewayInstanceId", *req.GatewayInstanceId)
			response.ErrorJSON(ctx, "查询网关实例CORS配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该网关实例未配置CORS", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询网关实例CORS配置成功", "tenantId", tenantId, 
			"gatewayInstanceId", *req.GatewayInstanceId, "corsConfigId", config.CorsConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 按路由配置ID查询单个配置
	if req.RouteConfigId != nil && *req.RouteConfigId != "" {
		config, err := c.dao.GetCorsConfigByRouteConfig(tenantId, *req.RouteConfigId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "查询路由配置CORS配置失败", "error", err.Error(), 
				"tenantId", tenantId, "routeConfigId", *req.RouteConfigId)
			response.ErrorJSON(ctx, "查询路由配置CORS配置失败: "+err.Error(), constants.ED00009)
			return
		}

		if config == nil {
			response.ErrorJSON(ctx, "该路由未配置CORS", constants.ED00008)
			return
		}

		logger.InfoWithTrace(ctx, "查询路由配置CORS配置成功", "tenantId", tenantId, 
			"routeConfigId", *req.RouteConfigId, "corsConfigId", config.CorsConfigId)
		response.SuccessJSON(ctx, config, constants.SD00002)
		return
	}

	// 如果到这里说明参数有问题
	response.ErrorJSON(ctx, "查询参数无效", constants.ED00007)
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

	// 从URL路径获取CORS配置ID
	corsConfigId := request.GetParam(ctx, "corsConfigId")
	if corsConfigId == "" {
		response.ErrorJSON(ctx, "CORS配置ID不能为空", constants.ED00007)
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

	// 设置从URL获取的CORS配置ID和租户ID
	config.CorsConfigId = corsConfigId
	config.TenantId = tenantId

	// 调用DAO层更新CORS配置
	err := c.dao.UpdateCorsConfig(ctx, &config, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新CORS配置失败", "error", err.Error(), 
			"corsConfigId", corsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "更新CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询最新的配置数据返回给前端
	updatedConfig, err := c.dao.GetCorsConfig(tenantId, corsConfigId)
	if err != nil {
		logger.WarnWithTrace(ctx, "更新成功但获取最新数据失败", "error", err.Error(), 
			"corsConfigId", corsConfigId, "tenantId", tenantId)
		// 更新成功但获取最新数据失败，仍然返回成功
		response.SuccessJSON(ctx, gin.H{"message": "CORS配置更新成功"}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "CORS配置更新成功", "corsConfigId", corsConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, updatedConfig, constants.SD00003)
}

// DeleteCorsConfig 删除CORS配置
func (c *CorsConfigController) DeleteCorsConfig(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始删除CORS配置", "controller", "CorsConfigController", "action", "DeleteCorsConfig")

	// 获取CORS配置ID参数
	corsConfigId := request.GetParam(ctx, "corsConfigId")
	if corsConfigId == "" {
		response.ErrorJSON(ctx, "CORS配置ID不能为空", constants.ED00007)
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

	// 调用DAO层删除CORS配置
	err := c.dao.DeleteCorsConfig(tenantId, corsConfigId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除CORS配置失败", "error", err.Error(), 
			"corsConfigId", corsConfigId, "tenantId", tenantId)
		response.ErrorJSON(ctx, "删除CORS配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "CORS配置删除成功", "corsConfigId", corsConfigId, 
		"tenantId", tenantId, "operatorId", operatorId)
	response.SuccessJSON(ctx, gin.H{"message": "CORS配置删除成功"}, constants.SD00003)
}

// QueryCorsConfigs 查询CORS配置列表
func (c *CorsConfigController) QueryCorsConfigs(ctx *gin.Context) {
	logger.InfoWithTrace(ctx, "开始查询CORS配置列表", "controller", "CorsConfigController", "action", "QueryCorsConfigs")

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO层查询CORS配置列表
	configs, total, err := c.dao.ListCorsConfigs(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询CORS配置列表失败", "error", err.Error(), 
			"tenantId", tenantId, "page", page, "pageSize", pageSize)
		response.ErrorJSON(ctx, "查询CORS配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应
	result := gin.H{
		"configs":  configs,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	logger.InfoWithTrace(ctx, "查询CORS配置列表成功", "tenantId", tenantId, "count", len(configs))
	response.SuccessJSON(ctx, result, constants.SD00002)
} 