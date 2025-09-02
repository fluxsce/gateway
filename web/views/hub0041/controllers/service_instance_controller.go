package controllers

import (
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0041/dao"
	"gateway/web/views/hub0041/models"

	"github.com/gin-gonic/gin"
)

// ServiceInstanceController 服务实例管理控制器
// 用于管理第三方应用注册的服务实例信息（提供查看、创建、编辑、删除功能）
type ServiceInstanceController struct {
	serviceInstanceDAO *dao.ServiceInstanceDAO
}

// NewServiceInstanceController 创建服务实例控制器
func NewServiceInstanceController(db database.Database) *ServiceInstanceController {
	return &ServiceInstanceController{
		serviceInstanceDAO: dao.NewServiceInstanceDAO(db),
	}
}

// QueryServiceInstances 查询服务实例列表
// @Summary 查询服务实例列表
// @Description 分页查询注册的服务实例列表，支持字段过滤
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body object{activeFlag=string,serviceName=string,groupName=string,instanceStatus=string,healthStatus=string,hostAddress=string,pageIndex=int,pageSize=int} false "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.ServiceInstance}
// @Router /gateway/hub0041/queryServiceInstances [post]
func (c *ServiceInstanceController) QueryServiceInstances(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 使用统一参数获取方法
	activeFlag := request.GetParam(ctx, "activeFlag")
	serviceName := request.GetParam(ctx, "serviceName")
	groupName := request.GetParam(ctx, "groupName")
	instanceStatus := request.GetParam(ctx, "instanceStatus")
	healthStatus := request.GetParam(ctx, "healthStatus")
	hostAddress := request.GetParam(ctx, "hostAddress")

	// 分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 调用DAO查询服务实例列表
	instances, total, err := c.serviceInstanceDAO.QueryServiceInstances(ctx, tenantId, activeFlag, serviceName, groupName, instanceStatus, healthStatus, hostAddress, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务实例列表失败", err)
		response.ErrorJSON(ctx, "查询服务实例列表失败", constants.ED00003)
		return
	}

	// 构建分页响应
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceInstanceId"

	// 使用统一的分页响应
	response.PageJSON(ctx, instances, pageInfo, constants.SD00002)
}

// GetServiceInstance 获取服务实例详情
// @Summary 获取服务实例详情
// @Description 根据服务实例ID获取服务实例详细信息
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body object{serviceInstanceId=string} true "获取请求"
// @Success 200 {object} response.JsonData{data=models.ServiceInstance}
// @Router /gateway/hub0041/getServiceInstance [post]
func (c *ServiceInstanceController) GetServiceInstance(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceInstanceId := request.GetParam(ctx, "serviceInstanceId")
	if strings.TrimSpace(serviceInstanceId) == "" {
		response.ErrorJSON(ctx, "服务实例ID不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务实例信息
	instance, err := c.serviceInstanceDAO.GetServiceInstance(ctx, tenantId, serviceInstanceId, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务实例不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "获取服务实例信息失败", err)
		response.ErrorJSON(ctx, "获取服务实例信息失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, instance, constants.SD00002)
}

// UpdateServiceInstance 更新服务实例信息
// @Summary 更新服务实例信息
// @Description 更新现有服务实例的配置信息
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body models.ServiceInstance true "更新请求"
// @Success 200 {object} response.JsonData{data=models.ServiceInstance}
// @Router /gateway/hub0041/updateServiceInstance [post]
func (c *ServiceInstanceController) UpdateServiceInstance(ctx *gin.Context) {
	// 解析请求参数
	var instance models.ServiceInstance
	if err := request.Bind(ctx, &instance); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 设置租户ID
	instance.TenantId = request.GetTenantID(ctx)

	// 验证必需参数
	if strings.TrimSpace(instance.ServiceInstanceId) == "" {
		response.ErrorJSON(ctx, "服务实例ID不能为空", constants.ED00006)
		return
	}

	// 获取操作人员ID
	operatorId := request.GetUserID(ctx)

	// 验证服务实例是否存在
	existingInstance, err := c.serviceInstanceDAO.GetServiceInstance(ctx, instance.TenantId, instance.ServiceInstanceId, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务实例不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "验证服务实例存在性失败", err)
		response.ErrorJSON(ctx, "验证服务实例存在性失败", constants.ED00003)
		return
	}

	// 保留不可修改的字段
	instance.ServiceGroupId = existingInstance.ServiceGroupId
	instance.ServiceName = existingInstance.ServiceName
	instance.GroupName = existingInstance.GroupName
	instance.HostAddress = existingInstance.HostAddress
	instance.PortNumber = existingInstance.PortNumber
	instance.RegisterTime = existingInstance.RegisterTime
	instance.AddTime = existingInstance.AddTime
	instance.AddWho = existingInstance.AddWho

	// 调用DAO更新服务实例
	updatedInstance, err := c.serviceInstanceDAO.UpdateServiceInstance(ctx, &instance, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务实例失败", err)
		response.ErrorJSON(ctx, "更新服务实例失败", constants.ED00003)
		return
	}

	if updatedInstance == nil {
		logger.ErrorWithTrace(ctx, "更新服务实例成功但返回记录为空", "serviceInstanceId", instance.ServiceInstanceId)
		response.ErrorJSON(ctx, "更新服务实例失败，返回记录为空", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务实例更新成功",
		"serviceInstanceId", updatedInstance.ServiceInstanceId,
		"serviceName", updatedInstance.ServiceName,
		"tenantId", updatedInstance.TenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, updatedInstance, constants.SD00001)
}

// DeleteServiceInstance 删除服务实例
// @Summary 删除服务实例
// @Description 删除指定的服务实例（物理删除）
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body object{serviceInstanceId=string} true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0041/deleteServiceInstance [post]
func (c *ServiceInstanceController) DeleteServiceInstance(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceInstanceId := request.GetParam(ctx, "serviceInstanceId")
	if strings.TrimSpace(serviceInstanceId) == "" {
		response.ErrorJSON(ctx, "服务实例ID不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 检查服务实例是否存在
	instance, err := c.serviceInstanceDAO.GetServiceInstance(ctx, tenantId, serviceInstanceId, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务实例不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "查询服务实例失败", err)
		response.ErrorJSON(ctx, "查询服务实例失败", constants.ED00003)
		return
	}

	// 调用DAO删除服务实例
	err = c.serviceInstanceDAO.DeleteServiceInstance(ctx, tenantId, serviceInstanceId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务实例失败", err)
		response.ErrorJSON(ctx, "删除服务实例失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务实例删除成功",
		"serviceInstanceId", serviceInstanceId,
		"serviceName", instance.ServiceName,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"serviceInstanceId": serviceInstanceId,
		"serviceName":       instance.ServiceName,
		"message":           "服务实例删除成功",
	}, constants.SD00001)
}

// UpdateInstanceHeartbeat 更新服务实例心跳
// @Summary 更新服务实例心跳
// @Description 更新服务实例的心跳时间
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body object{serviceInstanceId=string} true "心跳请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0041/updateInstanceHeartbeat [post]
func (c *ServiceInstanceController) UpdateInstanceHeartbeat(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceInstanceId := request.GetParam(ctx, "serviceInstanceId")
	if strings.TrimSpace(serviceInstanceId) == "" {
		response.ErrorJSON(ctx, "服务实例ID不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO更新心跳时间
	err := c.serviceInstanceDAO.UpdateInstanceHeartbeat(ctx, tenantId, serviceInstanceId, time.Now())
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新实例心跳失败", err)
		response.ErrorJSON(ctx, "更新实例心跳失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务实例心跳更新成功",
		"serviceInstanceId", serviceInstanceId,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"serviceInstanceId": serviceInstanceId,
		"heartbeatTime":     time.Now(),
		"message":           "心跳更新成功",
	}, constants.SD00001)
}

// UpdateInstanceHealthStatus 更新服务实例健康状态
// @Summary 更新服务实例健康状态
// @Description 更新服务实例的健康状态
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body object{serviceInstanceId=string,healthStatus=string} true "健康状态更新请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0041/updateInstanceHealthStatus [post]
func (c *ServiceInstanceController) UpdateInstanceHealthStatus(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceInstanceId := request.GetParam(ctx, "serviceInstanceId")
	healthStatus := request.GetParam(ctx, "healthStatus")

	if strings.TrimSpace(serviceInstanceId) == "" {
		response.ErrorJSON(ctx, "服务实例ID不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(healthStatus) == "" {
		response.ErrorJSON(ctx, "健康状态不能为空", constants.ED00006)
		return
	}

	// 验证健康状态值是否合法
	validHealthStatus := c.serviceInstanceDAO.GetHealthStatusOptions()
	isValid := false
	for _, status := range validHealthStatus {
		if status == healthStatus {
			isValid = true
			break
		}
	}

	if !isValid {
		response.ErrorJSON(ctx, "无效的健康状态值", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO更新健康状态
	err := c.serviceInstanceDAO.UpdateInstanceHealthStatus(ctx, tenantId, serviceInstanceId, healthStatus, time.Now())
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新实例健康状态失败", err)
		response.ErrorJSON(ctx, "更新实例健康状态失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务实例健康状态更新成功",
		"serviceInstanceId", serviceInstanceId,
		"healthStatus", healthStatus,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"serviceInstanceId": serviceInstanceId,
		"healthStatus":      healthStatus,
		"healthCheckTime":   time.Now(),
		"message":           "健康状态更新成功",
	}, constants.SD00001)
}

// GetInstanceStatusOptions 获取实例状态选项
// @Summary 获取实例状态选项
// @Description 获取系统支持的所有实例状态选项
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getInstanceStatusOptions [post]
func (c *ServiceInstanceController) GetInstanceStatusOptions(ctx *gin.Context) {
	options := c.serviceInstanceDAO.GetInstanceStatusOptions()
	response.SuccessJSON(ctx, options, constants.SD00002)
}

// GetHealthStatusOptions 获取健康状态选项
// @Summary 获取健康状态选项
// @Description 获取系统支持的所有健康状态选项
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getHealthStatusOptions [post]
func (c *ServiceInstanceController) GetHealthStatusOptions(ctx *gin.Context) {
	options := c.serviceInstanceDAO.GetHealthStatusOptions()
	response.SuccessJSON(ctx, options, constants.SD00002)
}

// CreateServiceInstance 创建服务实例
// @Summary 创建服务实例
// @Description 创建新的服务实例注册信息
// @Tags 服务实例管理
// @Accept json
// @Produce json
// @Param request body models.ServiceInstance true "创建请求"
// @Success 200 {object} response.JsonData{data=models.ServiceInstance}
// @Router /gateway/hub0041/createServiceInstance [post]
func (c *ServiceInstanceController) CreateServiceInstance(ctx *gin.Context) {
	// 解析请求参数
	var instance models.ServiceInstance
	if err := request.Bind(ctx, &instance); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 实例ID可以为空，DAO层会自动生成

	if strings.TrimSpace(instance.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(instance.ServiceGroupId) == "" {
		response.ErrorJSON(ctx, "服务分组ID不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(instance.GroupName) == "" {
		response.ErrorJSON(ctx, "分组名称不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(instance.HostAddress) == "" {
		response.ErrorJSON(ctx, "主机地址不能为空", constants.ED00006)
		return
	}

	if instance.PortNumber <= 0 {
		response.ErrorJSON(ctx, "端口号必须大于0", constants.ED00006)
		return
	}

	// 获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetUserID(ctx)
	instance.TenantId = tenantId

	// 调用DAO创建服务实例
	createdInstance, err := c.serviceInstanceDAO.CreateServiceInstance(ctx, &instance, operatorId)
	if err != nil {
		if strings.Contains(err.Error(), "实例ID重复") {
			response.ErrorJSON(ctx, "服务实例ID已存在", constants.ED00007)
			return
		}
		if strings.Contains(err.Error(), "服务不存在") {
			response.ErrorJSON(ctx, "指定的服务不存在或已禁用", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "创建服务实例失败", err)
		response.ErrorJSON(ctx, "创建服务实例失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务实例创建成功",
		"serviceInstanceId", createdInstance.ServiceInstanceId,
		"serviceName", createdInstance.ServiceName,
		"hostAddress", createdInstance.HostAddress,
		"portNumber", createdInstance.PortNumber,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, createdInstance, constants.SD00001)
}
