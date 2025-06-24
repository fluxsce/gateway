package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0021/dao"
	"gohub/web/views/hub0021/models"
	"time"

	"github.com/gin-gonic/gin"
)

// RouterConfigController Router配置控制器
type RouterConfigController struct {
	db               database.Database
	routerConfigDAO  *dao.RouterConfigDAO
}

// NewRouterConfigController 创建Router配置控制器
func NewRouterConfigController(db database.Database) *RouterConfigController {
	return &RouterConfigController{
		db:               db,
		routerConfigDAO:  dao.NewRouterConfigDAO(db),
	}
}

// QueryRouterConfigs 获取Router配置列表
// @Summary 获取Router配置列表
// @Description 分页获取Router配置列表
// @Tags Router配置管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param gatewayInstanceId query string false "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/queryRouterConfigs [post]
func (c *RouterConfigController) QueryRouterConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取可选的网关实例ID参数
	gatewayInstanceId := ctx.Query("gatewayInstanceId")

	// 调用DAO获取Router配置列表
	routerConfigs, total, err := c.routerConfigDAO.ListRouterConfigs(ctx, tenantId, gatewayInstanceId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取Router配置列表失败", err)
		response.ErrorJSON(ctx, "获取Router配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	configList := make([]map[string]interface{}, 0, len(routerConfigs))
	for _, config := range routerConfigs {
		configList = append(configList, routerConfigToMap(config))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "routerConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, configList, pageInfo, constants.SD00002)
}

// AddRouterConfig 创建Router配置
// @Summary 创建Router配置
// @Description 创建新的Router配置
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param routerConfig body models.RouterConfig true "Router配置信息"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/addRouterConfig [post]
func (c *RouterConfigController) AddRouterConfig(ctx *gin.Context) {
	var req models.RouterConfig
	if err := request.BindSafely(ctx, &req); err != nil {
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

	// 设置从上下文获取的租户ID和操作人信息
	req.TenantId = tenantId
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.AddTime = time.Now()
	req.EditTime = time.Now()

	// 清空Router配置ID，让DAO自动生成
	req.RouterConfigId = ""

	// 调用DAO添加Router配置
	routerConfigId, err := c.routerConfigDAO.AddRouterConfig(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建Router配置失败", err)
		response.ErrorJSON(ctx, "创建Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的Router配置信息
	newConfig, err := c.routerConfigDAO.GetRouterConfigById(ctx, routerConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的Router配置信息失败", err)
		// 即使查询失败，也返回成功但只带有Router配置ID
		response.SuccessJSON(ctx, gin.H{
			"routerConfigId": routerConfigId,
			"tenantId":       tenantId,
			"message":        "Router配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newConfig == nil {
		logger.ErrorWithTrace(ctx, "新创建的Router配置不存在", "routerConfigId", routerConfigId)
		response.SuccessJSON(ctx, gin.H{
			"routerConfigId": routerConfigId,
			"tenantId":       tenantId,
			"message":        "Router配置创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的Router配置信息
	configInfo := routerConfigToMap(newConfig)

	logger.InfoWithTrace(ctx, "Router配置创建成功", 
		"routerConfigId", routerConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routerName", newConfig.RouterName)

	response.SuccessJSON(ctx, configInfo, constants.SD00003)
}

// EditRouterConfig 更新Router配置
// @Summary 更新Router配置
// @Description 更新Router配置信息
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param routerConfig body models.RouterConfig true "Router配置信息"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/editRouterConfig [post]
func (c *RouterConfigController) EditRouterConfig(ctx *gin.Context) {
	var updateData models.RouterConfig
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.RouterConfigId == "" {
		response.ErrorJSON(ctx, "Router配置ID不能为空", constants.ED00007)
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

	// 获取现有Router配置信息
	currentConfig, err := c.routerConfigDAO.GetRouterConfigById(ctx, updateData.RouterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取Router配置信息失败", err)
		response.ErrorJSON(ctx, "获取Router配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentConfig == nil {
		response.ErrorJSON(ctx, "Router配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	routerConfigId := currentConfig.RouterConfigId
	tenantIdValue := currentConfig.TenantId
	addTime := currentConfig.AddTime
	addWho := currentConfig.AddWho

	// 设置更新时间和操作人（从上下文获取）
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 强制恢复不可修改的字段，防止前端恶意修改
	updateData.RouterConfigId = routerConfigId
	updateData.TenantId = tenantIdValue  // 强制使用数据库中的租户ID
	updateData.AddTime = addTime
	updateData.AddWho = addWho

	// 调用DAO更新Router配置
	err = c.routerConfigDAO.UpdateRouterConfig(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新Router配置失败", err)
		response.ErrorJSON(ctx, "更新Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的Router配置信息
	updatedConfig, err := c.routerConfigDAO.GetRouterConfigById(ctx, updateData.RouterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的Router配置信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回完整的Router配置信息
	configInfo := routerConfigToMap(updatedConfig)

	response.SuccessJSON(ctx, configInfo, constants.SD00004)
}

// DeleteRouterConfig 删除Router配置
// @Summary 删除Router配置
// @Description 删除Router配置
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param request body DeleteRouterConfigRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/deleteRouterConfig [post]
func (c *RouterConfigController) DeleteRouterConfig(ctx *gin.Context) {
	var req DeleteRouterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.RouterConfigId == "" {
		response.ErrorJSON(ctx, "Router配置ID不能为空", constants.ED00007)
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

	// 调用DAO删除Router配置
	err := c.routerConfigDAO.DeleteRouterConfig(ctx, req.RouterConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除Router配置失败", err)
		response.ErrorJSON(ctx, "删除Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"routerConfigId": req.RouterConfigId,
		"message":        "Router配置删除成功",
	}, constants.SD00005)
}

// GetRouterConfig 获取单个Router配置详情
// @Summary 获取Router配置详情
// @Description 根据ID获取Router配置详细信息
// @Tags Router配置管理
// @Produce json
// @Param routerConfigId query string true "Router配置ID"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/routerConfig [post]
func (c *RouterConfigController) GetRouterConfig(ctx *gin.Context) {
	var req GetRouterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	if req.RouterConfigId == "" {
		response.ErrorJSON(ctx, "Router配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取Router配置信息
	config, err := c.routerConfigDAO.GetRouterConfigById(ctx, req.RouterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取Router配置信息失败", err)
		response.ErrorJSON(ctx, "获取Router配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "Router配置不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	configInfo := routerConfigToMap(config)

	response.SuccessJSON(ctx, configInfo, constants.SD00001)
}

// GetRouterConfigsByInstance 根据网关实例获取Router配置列表
// @Summary 根据网关实例获取Router配置列表
// @Description 根据网关实例ID获取所有关联的Router配置
// @Tags Router配置管理
// @Produce json
// @Param request body GetRouterConfigsByInstanceRequest true "查询请求"
// @Success 200 {object} response.JsonData
// @Router /gohub/hub0021/routerConfigs/byInstance [post]
func (c *RouterConfigController) GetRouterConfigsByInstance(ctx *gin.Context) {
	var req GetRouterConfigsByInstanceRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	if req.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证上下文中的必要信息
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取Router配置列表
	configs, err := c.routerConfigDAO.GetRouterConfigsByGatewayInstance(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例Router配置列表失败", err)
		response.ErrorJSON(ctx, "获取Router配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	configList := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		configList = append(configList, routerConfigToMap(config))
	}

	response.SuccessJSON(ctx, gin.H{
		"routerConfigs":     configList,
		"gatewayInstanceId": req.GatewayInstanceId,
		"total":             len(configList),
	}, constants.SD00001)
}

// 请求结构体定义
type DeleteRouterConfigRequest struct {
	RouterConfigId string `json:"routerConfigId" form:"routerConfigId" binding:"required"` // Router配置ID
}

type GetRouterConfigRequest struct {
	RouterConfigId string `json:"routerConfigId" form:"routerConfigId" binding:"required"` // Router配置ID
}

type GetRouterConfigsByInstanceRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" binding:"required"` // 网关实例ID
}

// routerConfigToMap 将Router配置对象转换为Map
func routerConfigToMap(config *models.RouterConfig) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":               config.TenantId,
		"routerConfigId":         config.RouterConfigId,
		"gatewayInstanceId":      config.GatewayInstanceId,
		"routerName":             config.RouterName,
		"routerDesc":             config.RouterDesc,
		"defaultPriority":        config.DefaultPriority,
		"enableRouteCache":       config.EnableRouteCache,
		"routeCacheTtlSeconds":   config.RouteCacheTtlSeconds,
		"maxRoutes":              config.MaxRoutes,
		"routeMatchTimeout":      config.RouteMatchTimeout,
		"enableStrictMode":       config.EnableStrictMode,
		"enableMetrics":          config.EnableMetrics,
		"enableTracing":          config.EnableTracing,
		"caseSensitive":          config.CaseSensitive,
		"removeTrailingSlash":    config.RemoveTrailingSlash,
		"enableGlobalFilters":    config.EnableGlobalFilters,
		"filterExecutionMode":    config.FilterExecutionMode,
		"maxFilterChainDepth":    config.MaxFilterChainDepth,
		"enableRoutePooling":     config.EnableRoutePooling,
		"routePoolSize":          config.RoutePoolSize,
		"enableAsyncProcessing":  config.EnableAsyncProcessing,
		"enableFallback":         config.EnableFallback,
		"fallbackRoute":          config.FallbackRoute,
		"notFoundStatusCode":     config.NotFoundStatusCode,
		"notFoundMessage":        config.NotFoundMessage,
		"routerMetadata":         config.RouterMetadata,
		"customConfig":           config.CustomConfig,
		"reserved1":              config.Reserved1,
		"reserved2":              config.Reserved2,
		"reserved3":              config.Reserved3,
		"reserved4":              config.Reserved4,
		"reserved5":              config.Reserved5,
		"extProperty":            config.ExtProperty,
		"addTime":                config.AddTime,
		"addWho":                 config.AddWho,
		"editTime":               config.EditTime,
		"editWho":                config.EditWho,
		"oprSeqFlag":             config.OprSeqFlag,
		"currentVersion":         config.CurrentVersion,
		"activeFlag":             config.ActiveFlag,
		"noteText":               config.NoteText,
	}
} 