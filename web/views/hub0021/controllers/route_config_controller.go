package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0021/dao"
	"gateway/web/views/hub0021/models"
	"time"

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

	// 转换为响应格式，过滤敏感字段
	routeConfigList := make([]map[string]interface{}, 0, len(routeConfigs))
	for _, routeConfig := range routeConfigs {
		routeConfigList = append(routeConfigList, routeConfigWithServiceToMap(routeConfig))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "routeConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, routeConfigList, pageInfo, constants.SD00002)
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

	// 清空路由配置ID，让DAO自动生成
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
		// 即使查询失败，也返回成功但只带有路由配置ID
		response.SuccessJSON(ctx, gin.H{
			"routeConfigId": routeConfigId,
			"tenantId":      tenantId,
			"message":       "路由配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newRouteConfig == nil {
		logger.ErrorWithTrace(ctx, "新创建的路由配置不存在", "routeConfigId", routeConfigId)
		response.SuccessJSON(ctx, gin.H{
			"routeConfigId": routeConfigId,
			"tenantId":      tenantId,
			"message":       "路由配置创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的路由配置信息，排除敏感字段
	routeConfigInfo := routeConfigToMap(newRouteConfig)

	logger.InfoWithTrace(ctx, "路由配置创建成功",
		"routeConfigId", routeConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routeName", newRouteConfig.RouteName)

	response.SuccessJSON(ctx, routeConfigInfo, constants.SD00003)
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

	// 验证必填字段
	if updateData.RouteConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
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

	// 获取现有路由配置信息
	currentRouteConfig, err := c.routeConfigDAO.GetRouteConfigById(ctx, updateData.RouteConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由配置信息失败", err)
		response.ErrorJSON(ctx, "获取路由配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentRouteConfig == nil {
		response.ErrorJSON(ctx, "路由配置不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	updateData.TenantId = tenantId
	updateData.RouteConfigId = currentRouteConfig.RouteConfigId
	updateData.AddTime = currentRouteConfig.AddTime
	updateData.AddWho = currentRouteConfig.AddWho
	updateData.OprSeqFlag = currentRouteConfig.OprSeqFlag

	// 设置修改信息
	updateData.EditTime = time.Now()
	updateData.EditWho = operatorId

	// 调用DAO更新路由配置
	err = c.routeConfigDAO.UpdateRouteConfig(ctx, &updateData, operatorId)
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
			"message":       "路由配置更新成功，但获取最新信息失败",
		}, constants.SD00004)
		return
	}

	// 返回更新后的路由配置信息
	routeConfigInfo := routeConfigToMap(updatedRouteConfig)

	logger.InfoWithTrace(ctx, "路由配置更新成功",
		"routeConfigId", updateData.RouteConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routeName", updatedRouteConfig.RouteName)

	response.SuccessJSON(ctx, routeConfigInfo, constants.SD00004)
}

// DeleteRouteConfig 删除路由配置
// @Summary 删除路由配置
// @Description 删除路由配置（软删除）
// @Tags 路由配置管理
// @Accept json
// @Produce json
// @Param request body DeleteRouteConfigRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0021/route-configs [delete]
func (c *RouteConfigController) DeleteRouteConfig(ctx *gin.Context) {
	var req DeleteRouteConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.RouteConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
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

	// 检查路由配置是否存在
	existingRouteConfig, err := c.routeConfigDAO.GetRouteConfigById(ctx, req.RouteConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由配置信息失败", err)
		response.ErrorJSON(ctx, "获取路由配置信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if existingRouteConfig == nil {
		response.ErrorJSON(ctx, "路由配置不存在", constants.ED00008)
		return
	}

	// 调用DAO删除路由配置
	err = c.routeConfigDAO.DeleteRouteConfig(ctx, req.RouteConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除路由配置失败", err)
		response.ErrorJSON(ctx, "删除路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "路由配置删除成功",
		"routeConfigId", req.RouteConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"routeName", existingRouteConfig.RouteName)

	response.SuccessJSON(ctx, gin.H{
		"routeConfigId": req.RouteConfigId,
		"routeName":     existingRouteConfig.RouteName,
		"message":       "路由配置删除成功",
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
	if routeConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

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

	// 获取路由断言信息
	assertions, err := c.routeAssertionDAO.GetRouteAssertionsByRouteId(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由断言信息失败", err)
		// 路由断言获取失败不影响主要信息返回
		assertions = []*models.RouteAssertion{}
	}

	// 转换为响应格式
	routeConfigInfo := routeConfigToMap(routeConfig)

	// 添加断言信息
	assertionList := make([]map[string]interface{}, 0, len(assertions))
	for _, assertion := range assertions {
		assertionList = append(assertionList, routeAssertionToMap(assertion))
	}
	routeConfigInfo["assertions"] = assertionList

	response.SuccessJSON(ctx, routeConfigInfo, constants.SD00002)
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
	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取路由配置列表
	routeConfigs, err := c.routeConfigDAO.GetRouteConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例路由配置失败", err)
		response.ErrorJSON(ctx, "获取网关实例路由配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	routeConfigList := make([]map[string]interface{}, 0, len(routeConfigs))
	for _, routeConfig := range routeConfigs {
		routeConfigList = append(routeConfigList, routeConfigWithServiceToMap(routeConfig))
	}

	response.SuccessJSON(ctx, gin.H{
		"gatewayInstanceId": gatewayInstanceId,
		"routeConfigs":      routeConfigList,
		"total":             len(routeConfigList),
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
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 获取网关实例ID参数
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")

	// 调用DAO获取路由统计信息
	statistics, err := c.routeConfigDAO.GetRouteStatistics(ctx, tenantId, gatewayInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由统计信息失败", err)
		response.ErrorJSON(ctx, "获取路由统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建响应数据，确保字段名与前端接口一致
	responseData := gin.H{
		"totalRoutes":       statistics["totalRoutes"],
		"activeRoutes":      statistics["activeRoutes"],
		"inactiveRoutes":    statistics["inactiveRoutes"],
		"exactMatchRoutes":  statistics["exactMatchRoutes"],
		"prefixMatchRoutes": statistics["prefixMatchRoutes"],
		"regexMatchRoutes":  statistics["regexMatchRoutes"],
	}

	logger.InfoWithTrace(ctx, "路由统计信息查询成功",
		"tenantId", tenantId,
		"gatewayInstanceId", gatewayInstanceId,
		"totalRoutes", statistics["totalRoutes"])

	response.SuccessJSON(ctx, responseData, constants.SD00002)
}

// DeleteRouteConfigRequest 删除路由配置请求结构
type DeleteRouteConfigRequest struct {
	RouteConfigId string `json:"routeConfigId" form:"routeConfigId" binding:"required"` // 路由配置ID
}

// routeConfigToMap 将路由配置模型转换为Map，用于API响应
func routeConfigToMap(routeConfig *models.RouteConfig) map[string]interface{} {
	if routeConfig == nil {
		return nil
	}

	return map[string]interface{}{
		"tenantId":            routeConfig.TenantId,
		"routeConfigId":       routeConfig.RouteConfigId,
		"gatewayInstanceId":   routeConfig.GatewayInstanceId,
		"routeName":           routeConfig.RouteName,
		"routePath":           routeConfig.RoutePath,
		"allowedMethods":      routeConfig.AllowedMethods,
		"allowedHosts":        routeConfig.AllowedHosts,
		"matchType":           routeConfig.MatchType,
		"routePriority":       routeConfig.RoutePriority,
		"stripPathPrefix":     routeConfig.StripPathPrefix,
		"rewritePath":         routeConfig.RewritePath,
		"enableWebsocket":     routeConfig.EnableWebsocket,
		"timeoutMs":           routeConfig.TimeoutMs,
		"retryCount":          routeConfig.RetryCount,
		"retryIntervalMs":     routeConfig.RetryIntervalMs,
		"serviceDefinitionId": routeConfig.ServiceDefinitionId,
		"logConfigId":         routeConfig.LogConfigId,
		"routeMetadata":       routeConfig.RouteMetadata,
		"extProperty":         routeConfig.ExtProperty,
		"addTime":             routeConfig.AddTime,
		"addWho":              routeConfig.AddWho,
		"editTime":            routeConfig.EditTime,
		"editWho":             routeConfig.EditWho,
		"currentVersion":      routeConfig.CurrentVersion,
		"activeFlag":          routeConfig.ActiveFlag,
		"noteText":            routeConfig.NoteText,
	}
}

// routeAssertionToMap 将路由断言模型转换为Map，用于API响应
func routeAssertionToMap(assertion *models.RouteAssertion) map[string]interface{} {
	if assertion == nil {
		return nil
	}

	return map[string]interface{}{
		"tenantId":          assertion.TenantId,
		"routeAssertionId":  assertion.RouteAssertionId,
		"routeConfigId":     assertion.RouteConfigId,
		"assertionName":     assertion.AssertionName,
		"assertionType":     assertion.AssertionType,
		"assertionOperator": assertion.AssertionOperator,
		"fieldName":         assertion.FieldName,
		"expectedValue":     assertion.ExpectedValue,
		"patternValue":      assertion.PatternValue,
		"caseSensitive":     assertion.CaseSensitive,
		"assertionOrder":    assertion.AssertionOrder,
		"isRequired":        assertion.IsRequired,
		"assertionDesc":     assertion.AssertionDesc,
		"addTime":           assertion.AddTime,
		"addWho":            assertion.AddWho,
		"editTime":          assertion.EditTime,
		"editWho":           assertion.EditWho,
		"currentVersion":    assertion.CurrentVersion,
		"activeFlag":        assertion.ActiveFlag,
		"noteText":          assertion.NoteText,
	}
}

// routeConfigWithServiceToMap 将带服务定义信息的路由配置模型转换为Map，用于API响应
func routeConfigWithServiceToMap(routeConfig *models.RouteConfigWithService) map[string]interface{} {
	if routeConfig == nil {
		return nil
	}

	result := map[string]interface{}{
		"tenantId":            routeConfig.TenantId,
		"routeConfigId":       routeConfig.RouteConfigId,
		"gatewayInstanceId":   routeConfig.GatewayInstanceId,
		"routeName":           routeConfig.RouteName,
		"routePath":           routeConfig.RoutePath,
		"allowedMethods":      routeConfig.AllowedMethods,
		"allowedHosts":        routeConfig.AllowedHosts,
		"matchType":           routeConfig.MatchType,
		"routePriority":       routeConfig.RoutePriority,
		"stripPathPrefix":     routeConfig.StripPathPrefix,
		"rewritePath":         routeConfig.RewritePath,
		"enableWebsocket":     routeConfig.EnableWebsocket,
		"timeoutMs":           routeConfig.TimeoutMs,
		"retryCount":          routeConfig.RetryCount,
		"retryIntervalMs":     routeConfig.RetryIntervalMs,
		"serviceDefinitionId": routeConfig.ServiceDefinitionId,
		"logConfigId":         routeConfig.LogConfigId,
		"routeMetadata":       routeConfig.RouteMetadata,
		"reserved1":           routeConfig.Reserved1,
		"reserved2":           routeConfig.Reserved2,
		"reserved3":           routeConfig.Reserved3,
		"reserved4":           routeConfig.Reserved4,
		"reserved5":           routeConfig.Reserved5,
		"extProperty":         routeConfig.ExtProperty,
		"addTime":             routeConfig.AddTime,
		"addWho":              routeConfig.AddWho,
		"editTime":            routeConfig.EditTime,
		"editWho":             routeConfig.EditWho,
		"oprSeqFlag":          routeConfig.OprSeqFlag,
		"currentVersion":      routeConfig.CurrentVersion,
		"activeFlag":          routeConfig.ActiveFlag,
		"noteText":            routeConfig.NoteText,
	}

	// 添加服务定义相关信息（可能为空）
	if routeConfig.ServiceName != nil {
		result["serviceName"] = *routeConfig.ServiceName
	} else {
		result["serviceName"] = nil
	}

	if routeConfig.ServiceDesc != nil {
		result["serviceDesc"] = *routeConfig.ServiceDesc
	} else {
		result["serviceDesc"] = nil
	}

	if routeConfig.ServiceType != nil {
		result["serviceType"] = *routeConfig.ServiceType
	} else {
		result["serviceType"] = nil
	}

	if routeConfig.LoadBalanceStrategy != nil {
		result["loadBalanceStrategy"] = *routeConfig.LoadBalanceStrategy
	} else {
		result["loadBalanceStrategy"] = nil
	}

	return result
}
