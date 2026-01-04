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

// RouteConfigController 路由配置控制器
type RouteConfigController struct {
	db                database.Database
	routeConfigDAO    *dao.RouteConfigDAO
	routeAssertionDAO *dao.RouteAssertionDAO
}

// NewRouteConfigController 创建路由配置控制器
func NewRouteConfigController(db database.Database) *RouteConfigController {
	return &RouteConfigController{
		db:                db,
		routeConfigDAO:    dao.NewRouteConfigDAO(db),
		routeAssertionDAO: dao.NewRouteAssertionDAO(db),
	}
}

// QueryRouteConfigs 获取路由配置列表
// @Summary 获取路由配置列表
// @Description 分页获取路由配置列表
// @Tags 路由配置管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param gatewayInstanceId query string false "网关实例ID"
// @Param routeName query string false "路由名称(支持模糊匹配)"
// @Param routePath query string false "路由路径(支持模糊匹配)"
// @Param matchType query int false "匹配类型(0:精确匹配,1:前缀匹配,2:正则匹配)"
// @Param activeFlag query string false "激活状态(Y:激活,N:未激活)"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs [get]
func (c *RouteConfigController) QueryRouteConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取查询参数
	queryParams := &dao.RouteConfigQueryParams{
		TenantId:          tenantId,
		GatewayInstanceId: request.GetParam(ctx, "gatewayInstanceId"),
		RouteName:         request.GetParam(ctx, "routeName"),
		RoutePath:         request.GetParam(ctx, "routePath"),
		MatchType:         request.GetParamInt(ctx, "matchType", 0),
		ActiveFlag:        request.GetParam(ctx, "activeFlag"),
		Page:              page,
		PageSize:          pageSize,
	}

	// 调用DAO获取路由配置列表（关联服务定义）
	routeConfigs, total, err := c.routeConfigDAO.ListRouteConfigs(ctx, queryParams)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由配置列表失败", err)
		// 使用统一的错误响应
		response.ErrorJSON(ctx, "获取路由配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "routeConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, routeConfigs, pageInfo, constants.SD00002)
}

// AddRouteConfig 创建路由配置
// @Summary 创建路由配置
// @Description 创建新的路由配置
// @Tags 路由配置管理
// @Accept json
// @Produce json
// @Param routeConfig body models.RouteConfig true "路由配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs [post]
func (c *RouteConfigController) AddRouteConfig(ctx *gin.Context) {
	var req models.RouteConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 设置租户ID，清空路由配置ID让DAO自动生成
	req.TenantId = tenantId
	req.RouteConfigId = ""

	// 调用DAO添加路由配置
	routeConfigId, err := c.routeConfigDAO.AddRouteConfig(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建路由配置失败", err)
		response.ErrorJSON(ctx, "创建路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的路由配置信息
	newRouteConfig, err := c.routeConfigDAO.GetRouteConfigById(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的路由配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"routeConfigId": routeConfigId,
		}, constants.SD00003)
		return
	}

	// 返回完整的路由配置信息
	response.SuccessJSON(ctx, newRouteConfig, constants.SD00003)
}

// EditRouteConfig 更新路由配置
// @Summary 更新路由配置
// @Description 更新路由配置信息
// @Tags 路由配置管理
// @Accept json
// @Produce json
// @Param routeConfig body models.RouteConfig true "路由配置信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs [put]
func (c *RouteConfigController) EditRouteConfig(ctx *gin.Context) {
	var updateData models.RouteConfig
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 设置租户ID
	updateData.TenantId = tenantId

	// 调用DAO更新路由配置
	err := c.routeConfigDAO.UpdateRouteConfig(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新路由配置失败", err)
		response.ErrorJSON(ctx, "更新路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的路由配置信息
	updatedRouteConfig, err := c.routeConfigDAO.GetRouteConfigById(ctx, updateData.RouteConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的路由配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"routeConfigId": updateData.RouteConfigId,
		}, constants.SD00004)
		return
	}

	// 返回更新后的路由配置信息
	response.SuccessJSON(ctx, updatedRouteConfig, constants.SD00004)
}

// DeleteRouteConfig 删除路由配置
// @Summary 删除路由配置
// @Description 删除路由配置
// @Tags 路由配置管理
// @Accept json
// @Produce json
// @Param routeConfigId query string true "路由配置ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs [delete]
func (c *RouteConfigController) DeleteRouteConfig(ctx *gin.Context) {
	routeConfigId := request.GetParam(ctx, "routeConfigId")
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO删除路由配置
	err := c.routeConfigDAO.DeleteRouteConfig(ctx, routeConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除路由配置失败", err)
		response.ErrorJSON(ctx, "删除路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"routeConfigId": routeConfigId,
	}, constants.SD00005)
}

// GetRouteConfig 获取路由配置详情
// @Summary 获取路由配置详情
// @Description 根据路由配置ID获取详细信息
// @Tags 路由配置管理
// @Produce json
// @Param routeConfigId query string true "路由配置ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-config [get]
func (c *RouteConfigController) GetRouteConfig(ctx *gin.Context) {
	routeConfigId := request.GetParam(ctx, "routeConfigId")
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取路由配置信息
	routeConfig, err := c.routeConfigDAO.GetRouteConfigById(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由配置详情失败", err)
		response.ErrorJSON(ctx, "获取路由配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if routeConfig == nil {
		response.ErrorJSON(ctx, "路由配置不存在", constants.ED00008)
		return
	}

	// 直接返回路由配置信息
	response.SuccessJSON(ctx, routeConfig, constants.SD00002)
}

// GetRouteConfigsByInstance 根据网关实例获取路由配置列表
// @Summary 根据网关实例获取路由配置列表
// @Description 获取指定网关实例下的所有路由配置
// @Tags 路由配置管理
// @Produce json
// @Param gatewayInstanceId query string true "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs/by-instance [get]
func (c *RouteConfigController) GetRouteConfigsByInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	tenantId := request.GetTenantID(ctx)
	activeFlag := request.GetParam(ctx, "activeFlag")

	// 调用DAO获取路由配置列表
	routeConfigs, err := c.routeConfigDAO.GetRouteConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId, activeFlag)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例路由配置失败", err)
		response.ErrorJSON(ctx, "获取网关实例路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"routeConfigs":      routeConfigs,
		"total":             len(routeConfigs),
	}, constants.SD00002)
}

// GetRouteStatistics 获取路由统计信息
// @Summary 获取路由统计信息
// @Description 获取指定租户和网关实例的路由统计信息
// @Tags 路由配置管理
// @Produce json
// @Param gatewayInstanceId query string false "网关实例ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-statistics [get]
func (c *RouteConfigController) GetRouteStatistics(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 调用DAO获取路由统计信息
	statistics, err := c.routeConfigDAO.GetRouteStatistics(ctx, tenantId, gatewayInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由统计信息失败", err)
		response.ErrorJSON(ctx, "获取路由统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	response.SuccessJSON(ctx, statistics, constants.SD00002)
}
