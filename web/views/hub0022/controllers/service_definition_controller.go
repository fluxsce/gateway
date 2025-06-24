package controllers

import (
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/web/utils/constants"
	"gohub/web/utils/request"
	"gohub/web/utils/response"
	"gohub/web/views/hub0022/dao"
	"gohub/web/views/hub0022/models"
	"time"

	"github.com/gin-gonic/gin"
)

// ServiceDefinitionController 服务定义控制器
type ServiceDefinitionController struct {
	db                     database.Database
	serviceDefinitionDAO   *dao.ServiceDefinitionDAO
	serviceNodeDAO         *dao.ServiceNodeDAO
}

// NewServiceDefinitionController 创建服务定义控制器
func NewServiceDefinitionController(db database.Database) *ServiceDefinitionController {
	return &ServiceDefinitionController{
		db:                   db,
		serviceDefinitionDAO: dao.NewServiceDefinitionDAO(db),
		serviceNodeDAO:       dao.NewServiceNodeDAO(db),
	}
}

// QueryServiceDefinitions 获取服务定义列表
// @Summary 获取服务定义列表
// @Description 分页获取服务定义列表
// @Tags 服务定义管理
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.JsonData
// @Router /api/hub0022/service-definitions [get]
func (c *ServiceDefinitionController) QueryServiceDefinitions(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务定义列表
	serviceDefinitions, total, err := c.serviceDefinitionDAO.ListServiceDefinitions(ctx, tenantId, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义列表失败", err)
		response.ErrorJSON(ctx, "获取服务定义列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	serviceDefinitionList := make([]map[string]interface{}, 0, len(serviceDefinitions))
	for _, serviceDefinition := range serviceDefinitions {
		serviceDefinitionList = append(serviceDefinitionList, serviceDefinitionToMap(serviceDefinition))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceDefinitionId"

	// 使用统一的分页响应
	response.PageJSON(ctx, serviceDefinitionList, pageInfo, constants.SD00002)
}

// AddServiceDefinition 创建服务定义
// @Summary 创建服务定义
// @Description 创建新的服务定义
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param serviceDefinition body models.ServiceDefinition true "服务定义信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0022/service-definitions [post]
func (c *ServiceDefinitionController) CreateServiceDefinition(ctx *gin.Context) {
	var req models.ServiceDefinition
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

	// 清空服务定义ID，让DAO自动生成
	req.ServiceDefinitionId = ""

	// 调用DAO添加服务定义
	serviceDefinitionId, err := c.serviceDefinitionDAO.CreateServiceDefinition(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建服务定义失败", err)
		response.ErrorJSON(ctx, "创建服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的服务定义信息
	newServiceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, serviceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务定义信息失败", err)
		// 即使查询失败，也返回成功但只带有服务定义ID
		response.SuccessJSON(ctx, gin.H{
			"serviceDefinitionId": serviceDefinitionId,
			"tenantId":            tenantId,
			"message":             "服务定义创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newServiceDefinition == nil {
		logger.ErrorWithTrace(ctx, "新创建的服务定义不存在", "serviceDefinitionId", serviceDefinitionId)
		response.SuccessJSON(ctx, gin.H{
			"serviceDefinitionId": serviceDefinitionId,
			"tenantId":            tenantId,
			"message":             "服务定义创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的服务定义信息，排除敏感字段
	serviceDefinitionInfo := serviceDefinitionToMap(newServiceDefinition)

	logger.InfoWithTrace(ctx, "服务定义创建成功",
		"serviceDefinitionId", serviceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"serviceName", newServiceDefinition.ServiceName)

	response.SuccessJSON(ctx, serviceDefinitionInfo, constants.SD00003)
}

// EditServiceDefinition 更新服务定义
// @Summary 更新服务定义
// @Description 更新服务定义信息
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param serviceDefinition body models.ServiceDefinition true "服务定义信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0022/service-definitions [put]
func (c *ServiceDefinitionController) EditServiceDefinition(ctx *gin.Context) {
	var updateData models.ServiceDefinition
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.ServiceDefinitionId == "" {
		response.ErrorJSON(ctx, "服务定义ID不能为空", constants.ED00007)
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

	// 获取现有服务定义信息
	currentServiceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, updateData.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义信息失败", err)
		response.ErrorJSON(ctx, "获取服务定义信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentServiceDefinition == nil {
		response.ErrorJSON(ctx, "服务定义不存在", constants.ED00008)
		return
	}

	// 设置租户ID和操作人信息
	updateData.TenantId = tenantId

	// 调用DAO更新服务定义
	err = c.serviceDefinitionDAO.UpdateServiceDefinition(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务定义失败", err)
		response.ErrorJSON(ctx, "更新服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的服务定义信息
	updatedServiceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, updateData.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务定义信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"serviceDefinitionId": updateData.ServiceDefinitionId,
			"message":             "服务定义更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回更新后的服务定义信息
	serviceDefinitionInfo := serviceDefinitionToMap(updatedServiceDefinition)

	logger.InfoWithTrace(ctx, "服务定义更新成功",
		"serviceDefinitionId", updateData.ServiceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, serviceDefinitionInfo, constants.SD00004)
}

// DeleteServiceDefinition 删除服务定义
// @Summary 删除服务定义
// @Description 删除服务定义
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param request body DeleteServiceDefinitionRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0022/service-definitions [delete]
func (c *ServiceDefinitionController) DeleteServiceDefinition(ctx *gin.Context) {
	var req DeleteServiceDefinitionRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ServiceDefinitionId == "" {
		response.ErrorJSON(ctx, "服务定义ID不能为空", constants.ED00007)
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

	// 先查询服务定义是否存在
	existingServiceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, req.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务定义失败", err)
		response.ErrorJSON(ctx, "查询服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	if existingServiceDefinition == nil {
		response.ErrorJSON(ctx, "服务定义不存在", constants.ED00008)
		return
	}

	// 检查是否存在关联的服务节点
	serviceNodes, err := c.serviceNodeDAO.GetServiceNodesByService(ctx, req.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询关联服务节点失败", err)
		response.ErrorJSON(ctx, "查询关联服务节点失败: "+err.Error(), constants.ED00009)
		return
	}

	if len(serviceNodes) > 0 {
		response.ErrorJSON(ctx, "存在关联的服务节点，请先删除服务节点", constants.ED00009)
		return
	}

	// 调用DAO删除服务定义
	err = c.serviceDefinitionDAO.DeleteServiceDefinition(ctx, req.ServiceDefinitionId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务定义失败", err)
		response.ErrorJSON(ctx, "删除服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "服务定义删除成功",
		"serviceDefinitionId", req.ServiceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"serviceName", existingServiceDefinition.ServiceName)

	response.SuccessJSON(ctx, gin.H{
		"serviceDefinitionId": req.ServiceDefinitionId,
		"message":             "服务定义删除成功",
	}, constants.SD00005)
}

// GetServiceDefinition 获取服务定义详情
// @Summary 获取服务定义详情
// @Description 根据ID获取服务定义详情
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param request body GetServiceDefinitionRequest true "查询请求"
// @Success 200 {object} response.JsonData
// @Router /api/hub0022/service-definition [post]
func (c *ServiceDefinitionController) GetServiceDefinition(ctx *gin.Context) {
	var req GetServiceDefinitionRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ServiceDefinitionId == "" {
		response.ErrorJSON(ctx, "服务定义ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取服务定义详情
	serviceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, req.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义详情失败", err)
		response.ErrorJSON(ctx, "获取服务定义详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if serviceDefinition == nil {
		response.ErrorJSON(ctx, "服务定义不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	serviceDefinitionInfo := serviceDefinitionToMap(serviceDefinition)

	response.SuccessJSON(ctx, serviceDefinitionInfo, constants.SD00002)
}

// 请求结构体定义

// DeleteServiceDefinitionRequest 删除服务定义请求
type DeleteServiceDefinitionRequest struct {
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" binding:"required"` // 服务定义ID
}

// GetServiceDefinitionRequest 获取服务定义请求
type GetServiceDefinitionRequest struct {
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" binding:"required"` // 服务定义ID
}

// serviceDefinitionToMap 将服务定义转换为Map格式，过滤敏感字段
func serviceDefinitionToMap(serviceDefinition *models.ServiceDefinition) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                     serviceDefinition.TenantId,
		"serviceDefinitionId":          serviceDefinition.ServiceDefinitionId,
		"serviceName":                  serviceDefinition.ServiceName,
		"serviceDesc":                  serviceDefinition.ServiceDesc,
		"serviceType":                  serviceDefinition.ServiceType,
		"loadBalanceStrategy":          serviceDefinition.LoadBalanceStrategy,
		"discoveryType":                serviceDefinition.DiscoveryType,
		"discoveryConfig":              serviceDefinition.DiscoveryConfig,
		"sessionAffinity":              serviceDefinition.SessionAffinity,
		"stickySession":                serviceDefinition.StickySession,
		"maxRetries":                   serviceDefinition.MaxRetries,
		"retryTimeoutMs":               serviceDefinition.RetryTimeoutMs,
		"enableCircuitBreaker":         serviceDefinition.EnableCircuitBreaker,
		"healthCheckEnabled":           serviceDefinition.HealthCheckEnabled,
		"healthCheckPath":              serviceDefinition.HealthCheckPath,
		"healthCheckMethod":            serviceDefinition.HealthCheckMethod,
		"healthCheckIntervalSeconds":   serviceDefinition.HealthCheckIntervalSeconds,
		"healthCheckTimeoutMs":         serviceDefinition.HealthCheckTimeoutMs,
		"healthyThreshold":             serviceDefinition.HealthyThreshold,
		"unhealthyThreshold":           serviceDefinition.UnhealthyThreshold,
		"expectedStatusCodes":          serviceDefinition.ExpectedStatusCodes,
		"healthCheckHeaders":           serviceDefinition.HealthCheckHeaders,
		"loadBalancerConfig":           serviceDefinition.LoadBalancerConfig,
		"serviceMetadata":              serviceDefinition.ServiceMetadata,
		"activeFlag":                   serviceDefinition.ActiveFlag,
		"addTime":                      serviceDefinition.AddTime,
		"addWho":                       serviceDefinition.AddWho,
		"editTime":                     serviceDefinition.EditTime,
		"editWho":                      serviceDefinition.EditWho,
		"currentVersion":               serviceDefinition.CurrentVersion,
		"noteText":                     serviceDefinition.NoteText,
	}
} 