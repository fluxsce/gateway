package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"
	"gateway/web/views/hub0021/models"

	"github.com/gin-gonic/gin"
)

// RouterConfigController Router配置控制器
type RouterConfigController struct {
	db              database.Database
	routerConfigDAO *dao.RouterConfigDAO
}

// NewRouterConfigController 创建Router配置控制器
func NewRouterConfigController(db database.Database) *RouterConfigController {
	return &RouterConfigController{
		db:              db,
		routerConfigDAO: dao.NewRouterConfigDAO(db),
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
// @Router /gateway/hub0021/queryRouterConfigs [post]
func (c *RouterConfigController) QueryRouterConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)
	// 获取网关实例ID参数
	gatewayInstanceId := ctx.Query("gatewayInstanceId")

	// 调用DAO获取Router配置列表
	routerConfigs, total, err := c.routerConfigDAO.ListRouterConfigs(ctx, tenantId, gatewayInstanceId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取Router配置列表失败", err)
		response.ErrorJSON(ctx, "获取Router配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 直接返回Router配置列表
	configList := make([]*models.RouterConfig, 0, len(routerConfigs))
	for _, config := range routerConfigs {
		configList = append(configList, config)
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
// @Router /gateway/hub0021/addRouterConfig [post]
func (c *RouterConfigController) AddRouterConfig(ctx *gin.Context) {
	var req models.RouterConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取操作人信息和租户信息
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	req.TenantId = tenantId

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

	// 直接返回Router配置对象
	logger.InfoWithTrace(ctx, "Router配置创建成功",
		"routerConfigId", routerConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routerName", newConfig.RouterName)

	response.SuccessJSON(ctx, newConfig, constants.SD00003)
}

// EditRouterConfig 更新Router配置
// @Summary 更新Router配置
// @Description 更新Router配置信息
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param routerConfig body models.RouterConfig true "Router配置信息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/editRouterConfig [post]
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

	// 设置租户ID和操作人信息
	updateData.TenantId = tenantId

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
		response.SuccessJSON(ctx, gin.H{
			"routerConfigId": updateData.RouterConfigId,
			"message":        "Router配置更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 直接返回更新后的Router配置对象
	logger.InfoWithTrace(ctx, "Router配置更新成功",
		"routerConfigId", updateData.RouterConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, updatedConfig, constants.SD00004)
}

// DeleteRouterConfig 删除Router配置
// @Summary 删除Router配置
// @Description 删除Router配置
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param routerConfigId body string true "Router配置ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/deleteRouterConfig [post]
func (c *RouterConfigController) DeleteRouterConfig(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	routerConfigId := request.GetParam(ctx, "routerConfigId")
	if routerConfigId == "" {
		response.ErrorJSON(ctx, "Router配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 先查询Router配置是否存在
	existingConfig, err := c.routerConfigDAO.GetRouterConfigById(ctx, routerConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询Router配置失败", err)
		response.ErrorJSON(ctx, "查询Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	if existingConfig == nil {
		response.ErrorJSON(ctx, "Router配置不存在", constants.ED00008)
		return
	}

	// 调用DAO删除Router配置
	err = c.routerConfigDAO.DeleteRouterConfig(ctx, routerConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除Router配置失败", err)
		response.ErrorJSON(ctx, "删除Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "Router配置删除成功",
		"routerConfigId", routerConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routerName", existingConfig.RouterName)

	response.SuccessJSON(ctx, gin.H{
		"routerConfigId": routerConfigId,
		"message":        "Router配置删除成功",
	}, constants.SD00005)
}

// GetRouterConfig 获取单个Router配置详情
// @Summary 获取Router配置详情
// @Description 根据ID获取Router配置详细信息
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param routerConfigId body string true "Router配置ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/routerConfig [post]
func (c *RouterConfigController) GetRouterConfig(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	routerConfigId := request.GetParam(ctx, "routerConfigId")
	if routerConfigId == "" {
		response.ErrorJSON(ctx, "Router配置ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取Router配置详情
	config, err := c.routerConfigDAO.GetRouterConfigById(ctx, routerConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取Router配置详情失败", err)
		response.ErrorJSON(ctx, "获取Router配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if config == nil {
		response.ErrorJSON(ctx, "Router配置不存在", constants.ED00008)
		return
	}

	// 直接返回Router配置对象
	response.SuccessJSON(ctx, config, constants.SD00002)
}

// GetRouterConfigsByInstance 根据网关实例获取Router配置
// @Summary 根据网关实例获取Router配置
// @Description 根据网关实例ID获取Router配置（返回单条数据）
// @Tags Router配置管理
// @Accept json
// @Produce json
// @Param gatewayInstanceId body string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/routerConfigs/byInstance [post]
func (c *RouterConfigController) GetRouterConfigsByInstance(ctx *gin.Context) {
	// 使用 request.GetParam 获取参数
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取Router配置（返回单条数据）
	routerConfig, err := c.routerConfigDAO.GetRouterConfigByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例Router配置失败", err)
		response.ErrorJSON(ctx, "获取网关实例Router配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 如果配置不存在，返回 null
	if routerConfig == nil {
		response.SuccessJSON(ctx, nil, constants.SD00002)
		return
	}

	// 直接返回 RouterConfig 对象
	response.SuccessJSON(ctx, routerConfig, constants.SD00002)
}
