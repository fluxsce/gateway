package controllers

import (
	"encoding/json"
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

// FilterConfigController 过滤器配置控制器
type FilterConfigController struct {
	db               database.Database
	filterConfigDAO  *dao.FilterConfigDAO
}

// NewFilterConfigController 创建过滤器配置控制器
func NewFilterConfigController(db database.Database) *FilterConfigController {
	return &FilterConfigController{
		db:              db,
		filterConfigDAO: dao.NewFilterConfigDAO(db),
	}
}

// QueryFilterConfigs 获取过滤器配置列表
func (c *FilterConfigController) QueryFilterConfigs(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)
	// 获取可选的查询参数
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	// 调用DAO获取过滤器配置列表
	filterConfigs, total, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, gatewayInstanceId, routeConfigId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	filterConfigList := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, filterConfig := range filterConfigs {
		filterConfigList = append(filterConfigList, filterConfigToMap(filterConfig))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "filterConfigId"

	// 使用统一的分页响应
	response.PageJSON(ctx, filterConfigList, pageInfo, constants.SD00002)
}

// AddFilterConfig 创建过滤器配置
func (c *FilterConfigController) AddFilterConfig(ctx *gin.Context) {
	var req models.FilterConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
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

	// 设置从上下文获取的租户ID和操作人信息
	req.TenantId = tenantId
	req.AddWho = operatorId
	req.EditWho = operatorId
	req.AddTime = time.Now()
	req.EditTime = time.Now()

	// 清空过滤器配置ID，让DAO自动生成
	req.FilterConfigId = ""

	// 调用DAO添加过滤器配置
	filterConfigId, err := c.filterConfigDAO.AddFilterConfig(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建过滤器配置失败", err)
		response.ErrorJSON(ctx, "创建过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的过滤器配置信息
	newFilterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的过滤器配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": filterConfigId,
			"tenantId":       tenantId,
			"message":        "过滤器配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newFilterConfig == nil {
		logger.ErrorWithTrace(ctx, "新创建的过滤器配置不存在", "filterConfigId", filterConfigId)
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": filterConfigId,
			"tenantId":       tenantId,
			"message":        "过滤器配置创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的过滤器配置信息
	filterConfigInfo := filterConfigToMap(newFilterConfig)

	logger.InfoWithTrace(ctx, "过滤器配置创建成功",
		"filterConfigId", filterConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"filterName", newFilterConfig.FilterName)

	response.SuccessJSON(ctx, filterConfigInfo, constants.SD00003)
}

// GetFilterConfig 获取过滤器配置详情
func (c *FilterConfigController) GetFilterConfig(ctx *gin.Context) {
	filterConfigId := request.GetParam(ctx, "filterConfigId")
	tenantId := request.GetTenantID(ctx)

	if filterConfigId == "" {
		response.ErrorJSON(ctx, "过滤器配置ID不能为空", constants.ED00007)
		return
	}

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置信息
	filterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置详情失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if filterConfig == nil {
		response.ErrorJSON(ctx, "过滤器配置不存在", constants.ED00008)
		return
	}

	// 返回过滤器配置信息
	filterConfigInfo := filterConfigToMap(filterConfig)
	response.SuccessJSON(ctx, filterConfigInfo, constants.SD00002)
}

// EditFilterConfig 更新过滤器配置
func (c *FilterConfigController) EditFilterConfig(ctx *gin.Context) {
	var updateData models.FilterConfig
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.FilterConfigId == "" {
		response.ErrorJSON(ctx, "过滤器配置ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 设置租户ID和操作人信息
	updateData.TenantId = tenantId
	updateData.EditWho = operatorId
	updateData.EditTime = time.Now()

	// 调用DAO更新过滤器配置
	err := c.filterConfigDAO.UpdateFilterConfig(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新过滤器配置失败", err)
		response.ErrorJSON(ctx, "更新过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的过滤器配置信息
	updatedFilterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, updateData.FilterConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的过滤器配置信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": updateData.FilterConfigId,
			"message":        "过滤器配置更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回更新后的过滤器配置信息
	filterConfigInfo := filterConfigToMap(updatedFilterConfig)

	logger.InfoWithTrace(ctx, "过滤器配置更新成功",
		"filterConfigId", updateData.FilterConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, filterConfigInfo, constants.SD00004)
}

// DeleteFilterConfig 删除过滤器配置
func (c *FilterConfigController) DeleteFilterConfig(ctx *gin.Context) {
	var req DeleteFilterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO删除过滤器配置
	err := c.filterConfigDAO.DeleteFilterConfig(ctx, req.FilterConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除过滤器配置失败", err)
		response.ErrorJSON(ctx, "删除过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "过滤器配置删除成功",
		"filterConfigId", req.FilterConfigId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
		"message":        "过滤器配置删除成功",
	}, constants.SD00005)
}

// GetFilterConfigsByInstance 根据网关实例获取过滤器配置列表
func (c *FilterConfigController) GetFilterConfigsByInstance(ctx *gin.Context) {
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	tenantId := request.GetTenantID(ctx)

	if gatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置列表
	filterConfigs, err := c.filterConfigDAO.GetFilterConfigsByGatewayInstance(ctx, gatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "获取网关实例过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	filterConfigList := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, filterConfig := range filterConfigs {
		filterConfigList = append(filterConfigList, filterConfigToMap(filterConfig))
	}

	response.SuccessJSON(ctx, filterConfigList, constants.SD00002)
}

// GetFilterConfigsByRoute 根据路由获取过滤器配置列表
func (c *FilterConfigController) GetFilterConfigsByRoute(ctx *gin.Context) {
	routeConfigId := request.GetParam(ctx, "routeConfigId")
	tenantId := request.GetTenantID(ctx)

	if routeConfigId == "" {
		response.ErrorJSON(ctx, "路由配置ID不能为空", constants.ED00007)
		return
	}
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置列表
	filterConfigs, err := c.filterConfigDAO.GetFilterConfigsByRoute(ctx, routeConfigId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取路由过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "获取路由过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	filterConfigList := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, filterConfig := range filterConfigs {
		filterConfigList = append(filterConfigList, filterConfigToMap(filterConfig))
	}

	response.SuccessJSON(ctx, filterConfigList, constants.SD00002)
}

// GetFilterConfigsByType 根据过滤器类型查询配置
func (c *FilterConfigController) GetFilterConfigsByType(ctx *gin.Context) {
	filterType := request.GetParam(ctx, "filterType")
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	if filterType == "" {
		response.ErrorJSON(ctx, "过滤器类型不能为空", constants.ED00007)
		return
	}
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置列表
	filterConfigs, err := c.filterConfigDAO.GetFilterConfigsByType(ctx, filterType, tenantId, gatewayInstanceId, routeConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "根据类型获取过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "根据类型获取过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	filterConfigList := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, filterConfig := range filterConfigs {
		filterConfigList = append(filterConfigList, filterConfigToMap(filterConfig))
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigs": filterConfigList,
		"filterType":    filterType,
		"total":         len(filterConfigList),
	}, constants.SD00002)
}

// GetFilterConfigsByAction 根据执行时机查询配置
func (c *FilterConfigController) GetFilterConfigsByAction(ctx *gin.Context) {
	filterAction := request.GetParam(ctx, "filterAction")
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	if filterAction == "" {
		response.ErrorJSON(ctx, "过滤器执行时机不能为空", constants.ED00007)
		return
	}
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器配置列表
	filterConfigs, err := c.filterConfigDAO.GetFilterConfigsByAction(ctx, filterAction, tenantId, gatewayInstanceId, routeConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "根据执行时机获取过滤器配置列表失败", err)
		response.ErrorJSON(ctx, "根据执行时机获取过滤器配置列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	filterConfigList := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, filterConfig := range filterConfigs {
		filterConfigList = append(filterConfigList, filterConfigToMap(filterConfig))
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigs": filterConfigList,
		"filterAction":  filterAction,
		"total":         len(filterConfigList),
	}, constants.SD00002)
}

// GetFilterExecutionChain 获取过滤器执行链
func (c *FilterConfigController) GetFilterExecutionChain(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取过滤器执行链
	filterConfigs, err := c.filterConfigDAO.GetFilterExecutionChain(ctx, tenantId, gatewayInstanceId, routeConfigId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器执行链失败", err)
		response.ErrorJSON(ctx, "获取过滤器执行链失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式并按执行时机分组
	executionChain := make(map[string][]map[string]interface{})
	executionChain["pre-routing"] = make([]map[string]interface{}, 0)
	executionChain["post-routing"] = make([]map[string]interface{}, 0)
	executionChain["pre-response"] = make([]map[string]interface{}, 0)

	for _, filterConfig := range filterConfigs {
		filterConfigMap := filterConfigToMap(filterConfig)
		executionChain[filterConfig.FilterAction] = append(executionChain[filterConfig.FilterAction], filterConfigMap)
	}

	response.SuccessJSON(ctx, gin.H{
		"executionChain": executionChain,
		"total":          len(filterConfigs),
	}, constants.SD00002)
}

// UpdateFilterOrder 调整过滤器执行顺序
func (c *FilterConfigController) UpdateFilterOrder(ctx *gin.Context) {
	var req UpdateFilterOrderRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO更新过滤器执行顺序
	err := c.filterConfigDAO.UpdateFilterOrder(ctx, req.FilterConfigId, tenantId, req.NewOrder, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新过滤器执行顺序失败", err)
		response.ErrorJSON(ctx, "更新过滤器执行顺序失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "过滤器执行顺序更新成功",
		"filterConfigId", req.FilterConfigId,
		"newOrder", req.NewOrder,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
		"newOrder":       req.NewOrder,
		"message":        "过滤器执行顺序更新成功",
	}, constants.SD00004)
}

// EnableFilterConfig 启用过滤器配置
func (c *FilterConfigController) EnableFilterConfig(ctx *gin.Context) {
	var req EnableDisableFilterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO启用过滤器配置
	err := c.filterConfigDAO.EnableFilterConfig(ctx, req.FilterConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "启用过滤器配置失败", err)
		response.ErrorJSON(ctx, "启用过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "过滤器配置启用成功",
		"filterConfigId", req.FilterConfigId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
		"status":         "enabled",
		"message":        "过滤器配置启用成功",
	}, constants.SD00004)
}

// DisableFilterConfig 禁用过滤器配置
func (c *FilterConfigController) DisableFilterConfig(ctx *gin.Context) {
	var req EnableDisableFilterConfigRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO禁用过滤器配置
	err := c.filterConfigDAO.DisableFilterConfig(ctx, req.FilterConfigId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "禁用过滤器配置失败", err)
		response.ErrorJSON(ctx, "禁用过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "过滤器配置禁用成功",
		"filterConfigId", req.FilterConfigId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"filterConfigId": req.FilterConfigId,
		"status":         "disabled",
		"message":        "过滤器配置禁用成功",
	}, constants.SD00004)
}

// ValidateFilterConfig 验证过滤器配置
func (c *FilterConfigController) ValidateFilterConfig(ctx *gin.Context) {
	var req models.FilterConfig
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证过滤器类型
	if !models.IsValidFilterType(req.FilterType) {
		response.ErrorJSON(ctx, "无效的过滤器类型: "+req.FilterType, constants.ED00007)
		return
	}

	// 验证过滤器执行时机
	if !models.IsValidFilterAction(req.FilterAction) {
		response.ErrorJSON(ctx, "无效的过滤器执行时机: "+req.FilterAction, constants.ED00007)
		return
	}

	// 验证过滤器配置JSON格式
	if req.FilterConfig != "" {
		var configTest interface{}
		if err := json.Unmarshal([]byte(req.FilterConfig), &configTest); err != nil {
			response.ErrorJSON(ctx, "过滤器配置不是有效的JSON格式: "+err.Error(), constants.ED00007)
			return
		}
	}

	// 验证实例级或路由级配置
	if req.GatewayInstanceId == "" && req.RouteConfigId == "" {
		response.ErrorJSON(ctx, "必须指定网关实例ID或路由配置ID", constants.ED00007)
		return
	}
	if req.GatewayInstanceId != "" && req.RouteConfigId != "" {
		response.ErrorJSON(ctx, "不能同时指定网关实例ID和路由配置ID", constants.ED00007)
		return
	}

	response.SuccessJSON(ctx, gin.H{
		"message": "过滤器配置验证通过",
		"valid":   true,
	}, constants.SD00002)
}

// GetFilterConfigTemplates 获取过滤器配置模板
func (c *FilterConfigController) GetFilterConfigTemplates(ctx *gin.Context) {
	templates := models.GetFilterConfigTemplates()

	response.SuccessJSON(ctx, gin.H{
		"templates": templates,
		"total":     len(templates),
	}, constants.SD00002)
}

// CreateFilterConfigFromTemplate 从模板创建过滤器配置
func (c *FilterConfigController) CreateFilterConfigFromTemplate(ctx *gin.Context) {
	var req CreateFromTemplateRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取模板
	templates := models.GetFilterConfigTemplates()
	var selectedTemplate *models.FilterConfigTemplate
	for _, template := range templates {
		if template.Name == req.TemplateName {
			selectedTemplate = &template
			break
		}
	}

	if selectedTemplate == nil {
		response.ErrorJSON(ctx, "未找到指定的模板: "+req.TemplateName, constants.ED00008)
		return
	}

	// 从模板创建过滤器配置
	filterConfig := models.FilterConfig{
		FilterName:        req.FilterName,
		FilterType:        selectedTemplate.FilterType,
		FilterAction:      selectedTemplate.FilterAction,
		FilterOrder:       selectedTemplate.DefaultOrder,
		FilterDesc:        selectedTemplate.Description,
		GatewayInstanceId: req.GatewayInstanceId,
		RouteConfigId:     req.RouteConfigId,
	}

	// 序列化模板配置
	configBytes, err := json.Marshal(selectedTemplate.ConfigSchema)
	if err != nil {
		response.ErrorJSON(ctx, "序列化模板配置失败: "+err.Error(), constants.ED00009)
		return
	}
	filterConfig.FilterConfig = string(configBytes)

	// 获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	filterConfig.TenantId = tenantId

	// 调用DAO添加过滤器配置
	filterConfigId, err := c.filterConfigDAO.AddFilterConfig(ctx, &filterConfig, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "从模板创建过滤器配置失败", err)
		response.ErrorJSON(ctx, "从模板创建过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新创建的过滤器配置
	newFilterConfig, err := c.filterConfigDAO.GetFilterConfigById(ctx, filterConfigId, tenantId)
	if err != nil {
		response.SuccessJSON(ctx, gin.H{
			"filterConfigId": filterConfigId,
			"message":        "过滤器配置创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	filterConfigInfo := filterConfigToMap(newFilterConfig)

	logger.InfoWithTrace(ctx, "从模板创建过滤器配置成功",
		"filterConfigId", filterConfigId,
		"templateName", req.TemplateName,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, filterConfigInfo, constants.SD00003)
}

// GetFilterConfigStats 获取过滤器配置统计信息
func (c *FilterConfigController) GetFilterConfigStats(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 获取所有过滤器配置
	filterConfigs, _, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, gatewayInstanceId, routeConfigId, 1, 10000)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取过滤器配置统计信息失败", err)
		response.ErrorJSON(ctx, "获取过滤器配置统计信息失败: "+err.Error(), constants.ED00009)
		return
	}

	// 统计信息
	stats := map[string]interface{}{
		"total": len(filterConfigs),
		"byType": make(map[string]int),
		"byAction": make(map[string]int),
		"byStatus": map[string]int{
			"active":   0,
			"inactive": 0,
		},
	}

	byType := stats["byType"].(map[string]int)
	byAction := stats["byAction"].(map[string]int)
	byStatus := stats["byStatus"].(map[string]int)

	for _, config := range filterConfigs {
		// 按类型统计
		byType[config.FilterType]++
		
		// 按执行时机统计
		byAction[config.FilterAction]++
		
		// 按状态统计
		if config.ActiveFlag == "Y" {
			byStatus["active"]++
		} else {
			byStatus["inactive"]++
		}
	}

	response.SuccessJSON(ctx, stats, constants.SD00002)
}

// ExportFilterConfigs 导出过滤器配置
func (c *FilterConfigController) ExportFilterConfigs(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	gatewayInstanceId := request.GetParam(ctx, "gatewayInstanceId")
	routeConfigId := request.GetParam(ctx, "routeConfigId")

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 获取过滤器配置列表
	filterConfigs, _, err := c.filterConfigDAO.ListFilterConfigs(ctx, tenantId, gatewayInstanceId, routeConfigId, 1, 10000)
	if err != nil {
		logger.ErrorWithTrace(ctx, "导出过滤器配置失败", err)
		response.ErrorJSON(ctx, "导出过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为导出格式
	exportData := make([]map[string]interface{}, 0, len(filterConfigs))
	for _, config := range filterConfigs {
		exportData = append(exportData, filterConfigToMap(config))
	}

	response.SuccessJSON(ctx, gin.H{
		"filterConfigs": exportData,
		"exportTime":    time.Now(),
		"total":         len(exportData),
	}, constants.SD00002)
}

// ImportFilterConfigs 导入过滤器配置
func (c *FilterConfigController) ImportFilterConfigs(ctx *gin.Context) {
	var req ImportFilterConfigsRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	successCount := 0
	failedCount := 0
	errors := make([]string, 0)

	for _, configData := range req.FilterConfigs {
		configData.TenantId = tenantId
		configData.FilterConfigId = "" // 让DAO自动生成

		_, err := c.filterConfigDAO.AddFilterConfig(ctx, &configData, operatorId)
		if err != nil {
			failedCount++
			errors = append(errors, "导入过滤器配置失败: "+configData.FilterName+" - "+err.Error())
		} else {
			successCount++
		}
	}

	logger.InfoWithTrace(ctx, "过滤器配置导入完成",
		"successCount", successCount,
		"failedCount", failedCount,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"successCount": successCount,
		"failedCount":  failedCount,
		"errors":       errors,
		"message":      "过滤器配置导入完成",
	}, constants.SD00003)
}

// 请求结构体定义

// DeleteFilterConfigRequest 删除过滤器配置请求
type DeleteFilterConfigRequest struct {
	FilterConfigId string `json:"filterConfigId" form:"filterConfigId" binding:"required"` // 过滤器配置ID
}

// UpdateFilterOrderRequest 更新过滤器执行顺序请求
type UpdateFilterOrderRequest struct {
	FilterConfigId string `json:"filterConfigId" form:"filterConfigId" binding:"required"` // 过滤器配置ID
	NewOrder       int    `json:"newOrder" form:"newOrder" binding:"required"`             // 新的执行顺序
}

// EnableDisableFilterConfigRequest 启用/禁用过滤器配置请求
type EnableDisableFilterConfigRequest struct {
	FilterConfigId string `json:"filterConfigId" form:"filterConfigId" binding:"required"` // 过滤器配置ID
}

// CreateFromTemplateRequest 从模板创建过滤器配置请求
type CreateFromTemplateRequest struct {
	TemplateName      string `json:"templateName" form:"templateName" binding:"required"`           // 模板名称
	FilterName        string `json:"filterName" form:"filterName" binding:"required"`               // 过滤器名称
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId"`                    // 网关实例ID
	RouteConfigId     string `json:"routeConfigId" form:"routeConfigId"`                            // 路由配置ID
}

// ImportFilterConfigsRequest 导入过滤器配置请求
type ImportFilterConfigsRequest struct {
	FilterConfigs []models.FilterConfig `json:"filterConfigs" form:"filterConfigs" binding:"required"` // 过滤器配置列表
}

// BatchUpdateFilterConfigs 批量更新过滤器配置
func (c *FilterConfigController) BatchUpdateFilterConfigs(ctx *gin.Context) {
	var req struct {
		FilterConfigIds []string               `json:"filterConfigIds" form:"filterConfigIds" binding:"required"` // 过滤器配置ID列表
		Updates         map[string]interface{} `json:"updates" form:"updates" binding:"required"`                  // 更新字段
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.FilterConfigIds) == 0 {
		response.ErrorJSON(ctx, "过滤器配置ID列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO批量更新
	err := c.filterConfigDAO.BatchUpdateFilterConfigs(ctx, req.FilterConfigIds, tenantId, req.Updates, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "批量更新过滤器配置失败", err)
		response.ErrorJSON(ctx, "批量更新过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "批量更新过滤器配置成功",
		"filterConfigIds", req.FilterConfigIds,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"message":        "批量更新过滤器配置成功",
		"updatedCount":   len(req.FilterConfigIds),
		"filterConfigIds": req.FilterConfigIds,
	}, constants.SD00004)
}

// BatchDeleteFilterConfigs 批量删除过滤器配置
func (c *FilterConfigController) BatchDeleteFilterConfigs(ctx *gin.Context) {
	var req struct {
		FilterConfigIds []string `json:"filterConfigIds" form:"filterConfigIds" binding:"required"` // 过滤器配置ID列表
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.FilterConfigIds) == 0 {
		response.ErrorJSON(ctx, "过滤器配置ID列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 调用DAO批量删除
	err := c.filterConfigDAO.BatchDeleteFilterConfigs(ctx, req.FilterConfigIds, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "批量删除过滤器配置失败", err)
		response.ErrorJSON(ctx, "批量删除过滤器配置失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "批量删除过滤器配置成功",
		"filterConfigIds", req.FilterConfigIds,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"message":        "批量删除过滤器配置成功",
		"deletedCount":   len(req.FilterConfigIds),
		"filterConfigIds": req.FilterConfigIds,
	}, constants.SD00005)
}

// BatchUpdateFilterOrder 批量更新过滤器配置执行顺序
func (c *FilterConfigController) BatchUpdateFilterOrder(ctx *gin.Context) {
	var req struct {
		Orders []struct {
			FilterConfigId string `json:"filterConfigId" binding:"required"` // 过滤器配置ID
			NewOrder       int    `json:"newOrder" binding:"required"`       // 新的执行顺序
		} `json:"orders" form:"orders" binding:"required"` // 顺序配置列表
	}

	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证参数
	if len(req.Orders) == 0 {
		response.ErrorJSON(ctx, "顺序配置列表不能为空", constants.ED00007)
		return
	}

	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}
	if operatorId == "" {
		response.ErrorJSON(ctx, "无法获取操作人信息", constants.ED00007)
		return
	}

	// 批量更新执行顺序
	var updatedIds []string
	for _, order := range req.Orders {
		err := c.filterConfigDAO.UpdateFilterOrder(ctx, order.FilterConfigId, tenantId, order.NewOrder, operatorId)
		if err != nil {
			logger.ErrorWithTrace(ctx, "更新过滤器配置执行顺序失败", err,
				"filterConfigId", order.FilterConfigId,
				"newOrder", order.NewOrder)
			response.ErrorJSON(ctx, "更新过滤器配置执行顺序失败: "+err.Error(), constants.ED00009)
			return
		}
		updatedIds = append(updatedIds, order.FilterConfigId)
	}

	logger.InfoWithTrace(ctx, "批量更新过滤器配置执行顺序成功",
		"updatedIds", updatedIds,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, gin.H{
		"message":        "批量更新过滤器配置执行顺序成功",
		"updatedCount":   len(updatedIds),
		"filterConfigIds": updatedIds,
	}, constants.SD00004)
}

// GetFilterConfigUsage 获取过滤器配置使用情况
func (c *FilterConfigController) GetFilterConfigUsage(ctx *gin.Context) {
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 获取可选的过滤器配置ID
	filterConfigId := request.GetParam(ctx, "filterConfigId")

	// 这里可以根据实际需求返回过滤器的使用情况
	// 例如：被多少个网关实例使用、被多少个路由使用等
	usageInfo := map[string]interface{}{
		"filterConfigId": filterConfigId,
		"tenantId":       tenantId,
		"gatewayInstances": []map[string]interface{}{
			{
				"gatewayInstanceId": "gateway-001",
				"gatewayName":       "主网关",
				"status":            "active",
			},
		},
		"routes": []map[string]interface{}{
			{
				"routeConfigId": "route-001",
				"routeName":     "用户服务路由",
				"routePath":     "/api/user/*",
				"status":        "active",
			},
		},
		"totalUsage": 2,
		"lastUsed":   time.Now(),
		"statistics": map[string]interface{}{
			"requestCount":  1000,
			"errorCount":    5,
			"avgResponseTime": "120ms",
		},
	}

	response.SuccessJSON(ctx, usageInfo, constants.SD00002)
}

// filterConfigToMap 将过滤器配置转换为Map格式，过滤敏感字段
func filterConfigToMap(filterConfig *models.FilterConfig) map[string]interface{} {
	if filterConfig == nil {
		return nil
	}

	return map[string]interface{}{
		"tenantId":          filterConfig.TenantId,
		"filterConfigId":    filterConfig.FilterConfigId,
		"gatewayInstanceId": filterConfig.GatewayInstanceId,
		"routeConfigId":     filterConfig.RouteConfigId,
		"filterName":        filterConfig.FilterName,
		"filterType":        filterConfig.FilterType,
		"filterAction":      filterConfig.FilterAction,
		"filterOrder":       filterConfig.FilterOrder,
		"filterConfig":      filterConfig.FilterConfig,
		"filterDesc":        filterConfig.FilterDesc,
		"configId":          filterConfig.ConfigId,
		"extProperty":       filterConfig.ExtProperty,
		"addTime":           filterConfig.AddTime,
		"addWho":            filterConfig.AddWho,
		"editTime":          filterConfig.EditTime,
		"editWho":           filterConfig.EditWho,
		"currentVersion":    filterConfig.CurrentVersion,
		"activeFlag":        filterConfig.ActiveFlag,
		"noteText":          filterConfig.NoteText,
	}
} 