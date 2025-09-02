package controllers

import (
	"gateway/internal/registry/core"
	"gateway/pkg/logger"
	"gateway/web/sdkservice/regist/dao"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegistryController 注册服务控制器
// 处理服务注册、发现、心跳等HTTP请求
type RegistryController struct {
	registryDAO *dao.RegistryDAO
}

// NewRegistryController 创建注册服务控制器实例
func NewRegistryController() (*RegistryController, error) {
	// 创建DAO实例
	registryDAO, err := dao.NewRegistryDAO()
	if err != nil {
		return nil, err
	}

	return &RegistryController{
		registryDAO: registryDAO,
	}, nil
}

// =============================================================================
// 服务管理
// =============================================================================

// RegisterService 注册服务
func (ctrl *RegistryController) RegisterService(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")
	groupName := c.GetString("groupName")

	// 构建服务对象
	service := &core.Service{
		TenantId:                   tenantId,
		ServiceName:                request.GetParam(c, "serviceName"),
		ServiceGroupId:             serviceGroupId,
		GroupName:                  groupName,
		ServiceDescription:         request.GetParam(c, "serviceDescription"),
		ProtocolType:               request.GetParam(c, "protocolType"),
		ContextPath:                request.GetParam(c, "contextPath"),
		LoadBalanceStrategy:        request.GetParam(c, "loadBalanceStrategy"),
		HealthCheckUrl:             request.GetParam(c, "healthCheckUrl"),
		HealthCheckIntervalSeconds: request.GetParamInt(c, "healthCheckIntervalSeconds", 30),
		HealthCheckTimeoutSeconds:  request.GetParamInt(c, "healthCheckTimeoutSeconds", 5),
		HealthCheckType:            request.GetParam(c, "healthCheckType"),
		HealthCheckMode:            request.GetParam(c, "healthCheckMode"),
		MetadataJson:               request.GetParam(c, "metadataJson"),
		TagsJson:                   request.GetParam(c, "tagsJson"),
		NoteText:                   request.GetParam(c, "noteText"),
	}

	// 验证必填参数
	if service.ServiceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if service.ProtocolType == "" {
		service.ProtocolType = core.ProtocolTypeHTTP
	}
	if service.HealthCheckType == "" {
		service.HealthCheckType = "HTTP"
	}
	if service.HealthCheckMode == "" {
		service.HealthCheckMode = "PASSIVE"
	}

	// 调用DAO层注册服务
	err := ctrl.registryDAO.RegisterService(c.Request.Context(), service)
	if err != nil {
		logger.Error("注册服务失败", "error", err, "serviceName", service.ServiceName)
		response.ErrorJSON(c, "注册服务失败: "+err.Error(), "REGISTER_SERVICE_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, service, "REGISTER_SERVICE_SUCCESS")
}

// DeregisterService 注销服务
func (ctrl *RegistryController) DeregisterService(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")
	serviceName := request.GetParam(c, "serviceName")

	if serviceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	// 调用DAO层注销服务
	err := ctrl.registryDAO.DeregisterService(c.Request.Context(), tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.Error("注销服务失败", "error", err, "serviceName", serviceName)
		response.ErrorJSON(c, "注销服务失败: "+err.Error(), "DEREGISTER_SERVICE_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, nil, "DEREGISTER_SERVICE_SUCCESS")
}

// =============================================================================
// 服务实例管理
// =============================================================================

// RegisterInstance 注册服务实例
func (ctrl *RegistryController) RegisterInstance(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")
	groupName := c.GetString("groupName")

	// 构建服务实例对象
	instance := &core.ServiceInstance{
		TenantId:          tenantId,
		ServiceInstanceId: request.GetParam(c, "serviceInstanceId"),
		ServiceGroupId:    serviceGroupId,
		ServiceName:       request.GetParam(c, "serviceName"),
		GroupName:         groupName,
		HostAddress:       request.GetParam(c, "hostAddress"),
		PortNumber:        request.GetParamInt(c, "portNumber", 0),
		ContextPath:       request.GetParam(c, "contextPath"),
		InstanceStatus:    request.GetParam(c, "instanceStatus"),
		HealthStatus:      request.GetParam(c, "healthStatus"),
		WeightValue:       request.GetParamInt(c, "weightValue", 100),
		ClientId:          request.GetParam(c, "clientId"),
		ClientVersion:     request.GetParam(c, "clientVersion"),
		ClientType:        request.GetParam(c, "clientType"),
		TempInstanceFlag:  request.GetParam(c, "tempInstanceFlag"),
		MetadataJson:      request.GetParam(c, "metadataJson"),
		TagsJson:          request.GetParam(c, "tagsJson"),
		NoteText:          request.GetParam(c, "noteText"),
	}

	// 验证必填参数
	if instance.ServiceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	if instance.ServiceInstanceId == "" {
		response.ErrorJSON(c, "服务实例ID不能为空", "MISSING_SERVICE_INSTANCE_ID", http.StatusBadRequest)
		return
	}

	if instance.HostAddress == "" {
		response.ErrorJSON(c, "主机地址不能为空", "MISSING_HOST_ADDRESS", http.StatusBadRequest)
		return
	}

	if instance.PortNumber <= 0 {
		response.ErrorJSON(c, "端口号必须大于0", "INVALID_PORT_NUMBER", http.StatusBadRequest)
		return
	}

	// 设置默认值
	if instance.InstanceStatus == "" {
		instance.InstanceStatus = core.InstanceStatusUp
	}
	if instance.HealthStatus == "" {
		instance.HealthStatus = core.HealthStatusHealthy
	}
	if instance.TempInstanceFlag == "" {
		instance.TempInstanceFlag = core.TempInstanceFlagNo
	}
	if instance.ClientType == "" {
		instance.ClientType = core.ClientTypeService
	}

	// 调用DAO层注册服务实例
	instanceInfo, err := ctrl.registryDAO.RegisterInstance(c.Request.Context(), instance)
	if err != nil {
		logger.Error("注册服务实例失败", "error", err, "serviceName", instance.ServiceName, "instanceId", instance.ServiceInstanceId)
		response.ErrorJSON(c, "注册服务实例失败: "+err.Error(), "REGISTER_INSTANCE_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, instanceInfo, "REGISTER_INSTANCE_SUCCESS")
}

// UpdateInstance 更新服务实例
func (ctrl *RegistryController) UpdateInstance(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")
	groupName := c.GetString("groupName")

	// 构建服务实例对象
	instance := &core.ServiceInstance{
		TenantId:          tenantId,
		ServiceInstanceId: request.GetParam(c, "serviceInstanceId"),
		ServiceGroupId:    serviceGroupId,
		ServiceName:       request.GetParam(c, "serviceName"),
		GroupName:         groupName,
		HostAddress:       request.GetParam(c, "hostAddress"),
		PortNumber:        request.GetParamInt(c, "portNumber", 0),
		ContextPath:       request.GetParam(c, "contextPath"),
		InstanceStatus:    request.GetParam(c, "instanceStatus"),
		HealthStatus:      request.GetParam(c, "healthStatus"),
		WeightValue:       request.GetParamInt(c, "weightValue", 100),
		ClientId:          request.GetParam(c, "clientId"),
		ClientVersion:     request.GetParam(c, "clientVersion"),
		ClientType:        request.GetParam(c, "clientType"),
		TempInstanceFlag:  request.GetParam(c, "tempInstanceFlag"),
		MetadataJson:      request.GetParam(c, "metadataJson"),
		TagsJson:          request.GetParam(c, "tagsJson"),
		NoteText:          request.GetParam(c, "noteText"),
	}

	// 验证必填参数
	if instance.ServiceInstanceId == "" {
		response.ErrorJSON(c, "服务实例ID不能为空", "MISSING_SERVICE_INSTANCE_ID", http.StatusBadRequest)
		return
	}

	// 调用DAO层更新服务实例
	err := ctrl.registryDAO.UpdateInstance(c.Request.Context(), instance)
	if err != nil {
		logger.Error("更新服务实例失败", "error", err, "instanceId", instance.ServiceInstanceId)
		response.ErrorJSON(c, "更新服务实例失败: "+err.Error(), "UPDATE_INSTANCE_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, instance, "UPDATE_INSTANCE_SUCCESS")
}

// DeregisterInstance 注销服务实例
func (ctrl *RegistryController) DeregisterInstance(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceInstanceId := request.GetParam(c, "serviceInstanceId")

	if serviceInstanceId == "" {
		response.ErrorJSON(c, "服务实例ID不能为空", "MISSING_SERVICE_INSTANCE_ID", http.StatusBadRequest)
		return
	}

	// 调用DAO层注销服务实例
	err := ctrl.registryDAO.DeregisterInstance(c.Request.Context(), tenantId, serviceInstanceId)
	if err != nil {
		logger.Error("注销服务实例失败", "error", err, "instanceId", serviceInstanceId)
		response.ErrorJSON(c, "注销服务实例失败: "+err.Error(), "DEREGISTER_INSTANCE_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, nil, "DEREGISTER_INSTANCE_SUCCESS")
}

// SendHeartbeat 发送心跳
func (ctrl *RegistryController) SendHeartbeat(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")

	serviceInstanceId := request.GetParam(c, "serviceInstanceId")
	if serviceInstanceId == "" {
		response.ErrorJSON(c, "服务实例ID不能为空", "MISSING_SERVICE_INSTANCE_ID", http.StatusBadRequest)
		return
	}

	// 调用DAO层发送心跳
	err := ctrl.registryDAO.SendHeartbeat(c.Request.Context(), tenantId, serviceInstanceId)
	if err != nil {
		logger.Error("发送心跳失败", "error", err, "instanceId", serviceInstanceId)
		response.ErrorJSON(c, "发送心跳失败: "+err.Error(), "HEARTBEAT_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, nil, "HEARTBEAT_SUCCESS")
}

// UpdateInstanceStatus 更新实例状态
func (ctrl *RegistryController) UpdateInstanceStatus(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")

	serviceInstanceId := request.GetParam(c, "serviceInstanceId")
	instanceStatus := request.GetParam(c, "instanceStatus")
	healthStatus := request.GetParam(c, "healthStatus")
	weightValue := request.GetParamInt(c, "weightValue", 0)

	if serviceInstanceId == "" {
		response.ErrorJSON(c, "服务实例ID不能为空", "MISSING_SERVICE_INSTANCE_ID", http.StatusBadRequest)
		return
	}

	if instanceStatus == "" {
		response.ErrorJSON(c, "实例状态不能为空", "MISSING_INSTANCE_STATUS", http.StatusBadRequest)
		return
	}

	// 调用DAO层更新实例状态
	err := ctrl.registryDAO.UpdateInstanceStatus(c.Request.Context(), tenantId, serviceInstanceId, instanceStatus, healthStatus, weightValue)
	if err != nil {
		logger.Error("更新实例状态失败", "error", err, "instanceId", serviceInstanceId)
		response.ErrorJSON(c, "更新实例状态失败: "+err.Error(), "UPDATE_STATUS_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, nil, "UPDATE_STATUS_SUCCESS")
}

// =============================================================================
// 服务发现
// =============================================================================

// DiscoverInstances 发现服务实例
func (ctrl *RegistryController) DiscoverInstances(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")

	serviceName := request.GetParam(c, "serviceName")
	if serviceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	// 调用DAO层发现服务实例
	instanceInfo, err := ctrl.registryDAO.DiscoverInstance(c.Request.Context(), tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.Error("发现服务实例失败", "error", err, "serviceName", serviceName)
		response.ErrorJSON(c, "发现服务实例失败: "+err.Error(), "DISCOVERY_FAILED", http.StatusNotFound)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, instanceInfo, "DISCOVERY_SUCCESS")
}

// ListInstances 获取服务的所有实例列表
func (ctrl *RegistryController) ListInstances(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")

	serviceName := request.GetParam(c, "serviceName")
	if serviceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	// 调用DAO层获取服务实例列表
	instances, err := ctrl.registryDAO.ListInstances(c.Request.Context(), tenantId, serviceGroupId, serviceName)
	if err != nil {
		logger.Error("获取服务实例列表失败", "error", err, "serviceName", serviceName)
		response.ErrorJSON(c, "获取服务实例列表失败: "+err.Error(), "LIST_INSTANCES_FAILED", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, instances, "LIST_INSTANCES_SUCCESS")
}

// DiscoverService 发现服务
func (ctrl *RegistryController) DiscoverService(c *gin.Context) {
	// 从上下文获取认证信息
	tenantId := c.GetString("tenantId")
	serviceGroupId := c.GetString("serviceGroupId")
	groupName := c.GetString("groupName")

	serviceName := request.GetParam(c, "serviceName")
	if serviceName == "" {
		response.ErrorJSON(c, "服务名称不能为空", "MISSING_SERVICE_NAME", http.StatusBadRequest)
		return
	}

	// 调用DAO层发现服务
	service, err := ctrl.registryDAO.DiscoverService(c.Request.Context(), tenantId, serviceGroupId, groupName, serviceName)
	if err != nil {
		logger.Error("发现服务失败", "error", err, "serviceName", serviceName)
		response.ErrorJSON(c, "发现服务失败: "+err.Error(), "DISCOVER_SERVICE_FAILED", http.StatusNotFound)
		return
	}

	// 返回成功响应
	response.SuccessJSON(c, service, "DISCOVER_SERVICE_SUCCESS")
}
