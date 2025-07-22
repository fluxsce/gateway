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

// ServiceDefinitionController 服务定义控制器
type ServiceDefinitionController struct {
	db                   database.Database
	serviceDefinitionDAO *dao.ServiceDefinitionDAO
}

// NewServiceDefinitionController 创建服务定义控制器
func NewServiceDefinitionController(db database.Database) *ServiceDefinitionController {
	return &ServiceDefinitionController{
		db:                   db,
		serviceDefinitionDAO: dao.NewServiceDefinitionDAO(db),
	}
}

// GetServiceDefinitionsByInstance 根据网关实例ID获取服务定义列表
// @Summary 根据网关实例ID获取服务定义列表
// @Description 获取指定网关实例关联的所有服务定义，包含代理配置信息
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param request body GetServiceDefinitionsByInstanceRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/serviceDefinitions/byInstance [post]
func (c *ServiceDefinitionController) GetServiceDefinitionsByInstance(ctx *gin.Context) {
	var req GetServiceDefinitionsByInstanceRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.GatewayInstanceId == "" {
		response.ErrorJSON(ctx, "网关实例ID不能为空", constants.ED00007)
		return
	}

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取服务定义列表
	serviceDefinitions, err := c.serviceDefinitionDAO.GetServiceDefinitionsByInstance(ctx, req.GatewayInstanceId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取网关实例关联的服务定义列表失败", err)
		response.ErrorJSON(ctx, "获取服务定义列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式，过滤敏感字段
	serviceDefinitionList := make([]map[string]interface{}, 0, len(serviceDefinitions))
	for _, serviceDefinition := range serviceDefinitions {
		serviceDefinitionList = append(serviceDefinitionList, serviceDefinitionWithProxyToMap(serviceDefinition))
	}

	// 统计总数
	total := len(serviceDefinitions)

	logger.InfoWithTrace(ctx, "获取网关实例关联的服务定义列表成功",
		"gatewayInstanceId", req.GatewayInstanceId,
		"tenantId", tenantId,
		"total", total)

	// 返回结果（不使用分页，因为通常一个实例关联的服务定义数量不会太多）
	response.SuccessJSON(ctx, serviceDefinitionList, constants.SD00002)
}

// GetServiceDefinitionById 根据ID获取服务定义详情
// @Summary 根据ID获取服务定义详情
// @Description 获取指定服务定义的详细信息
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param request body GetServiceDefinitionRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/serviceDefinition [post]
func (c *ServiceDefinitionController) GetServiceDefinitionById(ctx *gin.Context) {
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

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 调用DAO获取服务定义
	serviceDefinition, err := c.serviceDefinitionDAO.GetServiceDefinitionById(ctx, req.ServiceDefinitionId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务定义失败", err)
		response.ErrorJSON(ctx, "获取服务定义失败: "+err.Error(), constants.ED00009)
		return
	}

	if serviceDefinition == nil {
		response.ErrorJSON(ctx, "服务定义不存在", constants.ED00008)
		return
	}

	// 转换为响应格式，过滤敏感字段
	serviceDefinitionInfo := serviceDefinitionToMap(serviceDefinition)

	logger.InfoWithTrace(ctx, "获取服务定义成功",
		"serviceDefinitionId", req.ServiceDefinitionId,
		"tenantId", tenantId,
		"serviceName", serviceDefinition.ServiceName)

	response.SuccessJSON(ctx, serviceDefinitionInfo, constants.SD00002)
}

// QueryServiceDefinitions 分页查询服务定义列表
// @Summary 分页查询服务定义列表
// @Description 分页获取服务定义列表，支持筛选
// @Tags 服务定义管理
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param serviceName query string false "服务名称（模糊查询）"
// @Param serviceType query int false "服务类型"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0021/queryServiceDefinitions [post]
func (c *ServiceDefinitionController) QueryServiceDefinitions(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	if tenantId == "" {
		response.ErrorJSON(ctx, "无法获取租户信息", constants.ED00007)
		return
	}

	// 构建筛选条件
	filters := make(map[string]interface{})

	// 服务名称模糊查询
	if serviceName := ctx.Query("serviceName"); serviceName != "" {
		filters["serviceName"] = serviceName
	}

	// 服务类型
	if serviceType := ctx.Query("serviceType"); serviceType != "" {
		filters["serviceType"] = serviceType
	}

	// 代理配置ID
	if proxyConfigId := ctx.Query("proxyConfigId"); proxyConfigId != "" {
		filters["proxyConfigId"] = proxyConfigId
	}

	// 负载均衡策略
	if loadBalanceStrategy := ctx.Query("loadBalanceStrategy"); loadBalanceStrategy != "" {
		filters["loadBalanceStrategy"] = loadBalanceStrategy
	}

	// 调用DAO获取服务定义列表
	serviceDefinitions, total, err := c.serviceDefinitionDAO.ListServiceDefinitions(ctx, tenantId, page, pageSize, filters)
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

// GetServiceDefinitionsByInstanceRequest 根据网关实例ID获取服务定义列表的请求参数
type GetServiceDefinitionsByInstanceRequest struct {
	GatewayInstanceId string `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" binding:"required"` // 网关实例ID
}

// GetServiceDefinitionRequest 获取服务定义的请求参数
type GetServiceDefinitionRequest struct {
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId" binding:"required"` // 服务定义ID
}

// serviceDefinitionWithProxyToMap 将服务定义和代理配置的组合对象转换为Map格式，过滤敏感字段
func serviceDefinitionWithProxyToMap(serviceDefinition *models.ServiceDefinitionWithProxy) map[string]interface{} {
	return map[string]interface{}{
		// 服务定义信息
		"tenantId":                   serviceDefinition.TenantId,
		"serviceDefinitionId":        serviceDefinition.ServiceDefinitionId,
		"serviceName":                serviceDefinition.ServiceName,
		"serviceDesc":                serviceDefinition.ServiceDesc,
		"serviceType":                serviceDefinition.ServiceType,
		"loadBalanceStrategy":        serviceDefinition.LoadBalanceStrategy,
		"discoveryType":              serviceDefinition.DiscoveryType,
		"discoveryConfig":            serviceDefinition.DiscoveryConfig,
		"sessionAffinity":            serviceDefinition.SessionAffinity,
		"stickySession":              serviceDefinition.StickySession,
		"maxRetries":                 serviceDefinition.MaxRetries,
		"retryTimeoutMs":             serviceDefinition.RetryTimeoutMs,
		"enableCircuitBreaker":       serviceDefinition.EnableCircuitBreaker,
		"healthCheckEnabled":         serviceDefinition.HealthCheckEnabled,
		"healthCheckPath":            serviceDefinition.HealthCheckPath,
		"healthCheckMethod":          serviceDefinition.HealthCheckMethod,
		"healthCheckIntervalSeconds": serviceDefinition.HealthCheckIntervalSeconds,
		"healthCheckTimeoutMs":       serviceDefinition.HealthCheckTimeoutMs,
		"healthyThreshold":           serviceDefinition.HealthyThreshold,
		"unhealthyThreshold":         serviceDefinition.UnhealthyThreshold,
		"expectedStatusCodes":        serviceDefinition.ExpectedStatusCodes,
		"healthCheckHeaders":         serviceDefinition.HealthCheckHeaders,
		"loadBalancerConfig":         serviceDefinition.LoadBalancerConfig,
		"serviceMetadata":            serviceDefinition.ServiceMetadata,

		// 代理配置信息
		"proxyConfigId":     serviceDefinition.ProxyConfigId,
		"proxyName":         serviceDefinition.ProxyName,
		"proxyType":         serviceDefinition.ProxyType,
		"proxyId":           serviceDefinition.ProxyId,
		"configPriority":    serviceDefinition.ConfigPriority,
		"proxyConfig":       serviceDefinition.ProxyConfig,
		"proxyCustomConfig": serviceDefinition.ProxyCustomConfig,

		// 网关实例信息
		"gatewayInstanceId": serviceDefinition.GatewayInstanceId,
		"instanceName":      serviceDefinition.InstanceName,
		"instanceDesc":      serviceDefinition.InstanceDesc,

		// 标准字段
		"addTime":        serviceDefinition.AddTime,
		"addWho":         serviceDefinition.AddWho,
		"editTime":       serviceDefinition.EditTime,
		"editWho":        serviceDefinition.EditWho,
		"currentVersion": serviceDefinition.CurrentVersion,
		"activeFlag":     serviceDefinition.ActiveFlag,
		"noteText":       serviceDefinition.NoteText,
	}
}

// serviceDefinitionToMap 将服务定义对象转换为Map格式，过滤敏感字段
func serviceDefinitionToMap(serviceDefinition *models.ServiceDefinition) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":                   serviceDefinition.TenantId,
		"serviceDefinitionId":        serviceDefinition.ServiceDefinitionId,
		"serviceName":                serviceDefinition.ServiceName,
		"serviceDesc":                serviceDefinition.ServiceDesc,
		"serviceType":                serviceDefinition.ServiceType,
		"proxyConfigId":              serviceDefinition.ProxyConfigId,
		"loadBalanceStrategy":        serviceDefinition.LoadBalanceStrategy,
		"discoveryType":              serviceDefinition.DiscoveryType,
		"discoveryConfig":            serviceDefinition.DiscoveryConfig,
		"sessionAffinity":            serviceDefinition.SessionAffinity,
		"stickySession":              serviceDefinition.StickySession,
		"maxRetries":                 serviceDefinition.MaxRetries,
		"retryTimeoutMs":             serviceDefinition.RetryTimeoutMs,
		"enableCircuitBreaker":       serviceDefinition.EnableCircuitBreaker,
		"healthCheckEnabled":         serviceDefinition.HealthCheckEnabled,
		"healthCheckPath":            serviceDefinition.HealthCheckPath,
		"healthCheckMethod":          serviceDefinition.HealthCheckMethod,
		"healthCheckIntervalSeconds": serviceDefinition.HealthCheckIntervalSeconds,
		"healthCheckTimeoutMs":       serviceDefinition.HealthCheckTimeoutMs,
		"healthyThreshold":           serviceDefinition.HealthyThreshold,
		"unhealthyThreshold":         serviceDefinition.UnhealthyThreshold,
		"expectedStatusCodes":        serviceDefinition.ExpectedStatusCodes,
		"healthCheckHeaders":         serviceDefinition.HealthCheckHeaders,
		"loadBalancerConfig":         serviceDefinition.LoadBalancerConfig,
		"serviceMetadata":            serviceDefinition.ServiceMetadata,
		"reserved1":                  serviceDefinition.Reserved1,
		"reserved2":                  serviceDefinition.Reserved2,
		"reserved3":                  serviceDefinition.Reserved3,
		"reserved4":                  serviceDefinition.Reserved4,
		"reserved5":                  serviceDefinition.Reserved5,
		"extProperty":                serviceDefinition.ExtProperty,
		"addTime":                    serviceDefinition.AddTime,
		"addWho":                     serviceDefinition.AddWho,
		"editTime":                   serviceDefinition.EditTime,
		"editWho":                    serviceDefinition.EditWho,
		"oprSeqFlag":                 serviceDefinition.OprSeqFlag,
		"currentVersion":             serviceDefinition.CurrentVersion,
		"activeFlag":                 serviceDefinition.ActiveFlag,
		"noteText":                   serviceDefinition.NoteText,
	}
}
