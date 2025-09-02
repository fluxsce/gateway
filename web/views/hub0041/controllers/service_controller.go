package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0041/dao"
	"gateway/web/views/hub0041/models"

	"github.com/gin-gonic/gin"
)

// ServiceController 服务管理控制器
// 用于管理第三方应用注册的服务信息（提供查看、创建、编辑、删除功能）
type ServiceController struct {
	serviceDAO      *dao.ServiceDAO
	serviceGroupDAO *dao.ServiceGroupDAO
}

// NewServiceController 创建服务控制器
func NewServiceController(db database.Database) *ServiceController {
	return &ServiceController{
		serviceDAO:      dao.NewServiceDAO(db),
		serviceGroupDAO: dao.NewServiceGroupDAO(db),
	}
}

// QueryServices 查询服务列表
// @Summary 查询服务列表
// @Description 分页查询注册的服务列表，支持字段过滤和关键字搜索
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body models.ServiceQueryRequest false "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.Service}
// @Router /gateway/hub0041/queryServices [post]
func (c *ServiceController) QueryServices(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 构建查询请求
	req := &models.ServiceQueryRequest{
		TenantId:     tenantId,
		ActiveFlag:   request.GetParam(ctx, "activeFlag"),
		GroupName:    request.GetParam(ctx, "groupName"),
		ServiceName:  request.GetParam(ctx, "serviceName"),
		ProtocolType: request.GetParam(ctx, "protocolType"),
		Keyword:      request.GetParam(ctx, "keyword"),
	}

	// 分页参数
	req.PageIndex, req.PageSize = request.GetPaginationParams(ctx)

	// 调用DAO查询服务列表
	services, total, err := c.serviceDAO.QueryServices(ctx, req)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务列表失败", err)
		response.ErrorJSON(ctx, "查询服务列表失败", constants.ED00003)
		return
	}

	// 构建分页响应
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	pageInfo.MainKey = "serviceName"

	// 使用统一的分页响应
	response.PageJSON(ctx, services, pageInfo, constants.SD00002)
}

// GetService 获取服务详情
// @Summary 获取服务详情
// @Description 根据服务名称获取服务详细信息
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body object{serviceName=string} true "获取请求"
// @Success 200 {object} response.JsonData{data=models.Service}
// @Router /gateway/hub0041/getService [post]
func (c *ServiceController) GetService(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceName := request.GetParam(ctx, "serviceName")
	if strings.TrimSpace(serviceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务信息
	service, err := c.serviceDAO.GetService(ctx, tenantId, serviceName, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "获取服务信息失败", err)
		response.ErrorJSON(ctx, "获取服务信息失败", constants.ED00003)
		return
	}

	response.SuccessJSON(ctx, service, constants.SD00002)
}

// UpdateService 更新服务信息
// @Summary 更新服务信息
// @Description 更新现有服务的配置信息
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body models.Service true "更新请求"
// @Success 200 {object} response.JsonData{data=models.Service}
// @Router /gateway/hub0041/updateService [post]
func (c *ServiceController) UpdateService(ctx *gin.Context) {
	// 解析请求参数
	var service models.Service
	if err := request.Bind(ctx, &service); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 设置租户ID
	service.TenantId = request.GetTenantID(ctx)

	// 验证必需参数
	if strings.TrimSpace(service.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	// 获取操作人员ID
	operatorId := request.GetUserID(ctx)

	// 验证服务是否存在
	existingService, err := c.serviceDAO.GetService(ctx, service.TenantId, service.ServiceName, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "验证服务存在性失败", err)
		response.ErrorJSON(ctx, "验证服务存在性失败", constants.ED00003)
		return
	}

	// 保留不可修改的字段
	service.ServiceGroupId = existingService.ServiceGroupId
	service.GroupName = existingService.GroupName
	service.AddTime = existingService.AddTime
	service.AddWho = existingService.AddWho

	// 调用DAO更新服务
	updatedService, err := c.serviceDAO.UpdateService(ctx, &service, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务失败", err)
		response.ErrorJSON(ctx, "更新服务失败", constants.ED00003)
		return
	}

	if updatedService == nil {
		logger.ErrorWithTrace(ctx, "更新服务成功但返回记录为空", "serviceName", service.ServiceName)
		response.ErrorJSON(ctx, "更新服务失败，返回记录为空", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务更新成功",
		"serviceName", updatedService.ServiceName,
		"tenantId", updatedService.TenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, updatedService, constants.SD00001)
}

// DeleteService 删除服务
// @Summary 删除服务
// @Description 删除指定的服务（物理删除）
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body object{serviceName=string} true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0041/deleteService [post]
func (c *ServiceController) DeleteService(ctx *gin.Context) {
	// 使用统一参数获取方法
	serviceName := request.GetParam(ctx, "serviceName")
	if strings.TrimSpace(serviceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 检查服务是否存在
	_, err := c.serviceDAO.GetService(ctx, tenantId, serviceName, "")
	if err != nil {
		if strings.Contains(err.Error(), "不存在") {
			response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "查询服务失败", err)
		response.ErrorJSON(ctx, "查询服务失败", constants.ED00003)
		return
	}

	// TODO: 检查服务下是否还有实例
	// 这里应该先检查该服务下是否还有关联的服务实例，如果有应该阻止删除或提示用户

	// 调用DAO删除服务
	err = c.serviceDAO.DeleteService(ctx, tenantId, serviceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务失败", err)
		response.ErrorJSON(ctx, "删除服务失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务删除成功",
		"serviceName", serviceName,
		"tenantId", tenantId)

	response.SuccessJSON(ctx, gin.H{
		"serviceName": serviceName,
		"message":     "服务删除成功",
	}, constants.SD00001)
}

// GetServiceProtocolTypes 获取服务协议类型列表
// @Summary 获取服务协议类型列表
// @Description 获取系统支持的所有服务协议类型
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getServiceProtocolTypes [post]
func (c *ServiceController) GetServiceProtocolTypes(ctx *gin.Context) {
	types := c.serviceDAO.GetServiceProtocolTypes()
	response.SuccessJSON(ctx, types, constants.SD00002)
}

// GetLoadBalanceStrategies 获取负载均衡策略列表
// @Summary 获取负载均衡策略列表
// @Description 获取系统支持的所有负载均衡策略
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Success 200 {object} response.JsonData{data=[]string}
// @Router /gateway/hub0041/getLoadBalanceStrategies [post]
func (c *ServiceController) GetLoadBalanceStrategies(ctx *gin.Context) {
	strategies := c.serviceDAO.GetLoadBalanceStrategies()
	response.SuccessJSON(ctx, strategies, constants.SD00002)
}

// CreateService 创建服务
// @Summary 创建服务
// @Description 创建新的服务注册信息
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body models.Service true "创建请求"
// @Success 200 {object} response.JsonData{data=models.Service}
// @Router /gateway/hub0041/createService [post]
func (c *ServiceController) CreateService(ctx *gin.Context) {
	// 解析请求参数
	var service models.Service
	if err := request.Bind(ctx, &service); err != nil {
		logger.ErrorWithTrace(ctx, "参数解析失败", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if strings.TrimSpace(service.ServiceName) == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(service.ServiceGroupId) == "" {
		response.ErrorJSON(ctx, "服务分组ID不能为空", constants.ED00006)
		return
	}

	if strings.TrimSpace(service.GroupName) == "" {
		response.ErrorJSON(ctx, "分组名称不能为空", constants.ED00006)
		return
	}

	// 获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetUserID(ctx)
	service.TenantId = tenantId

	// 调用DAO创建服务
	createdService, err := c.serviceDAO.CreateService(ctx, &service, operatorId)
	if err != nil {
		if strings.Contains(err.Error(), "服务名称重复") {
			response.ErrorJSON(ctx, "服务名称已存在", constants.ED00007)
			return
		}
		if strings.Contains(err.Error(), "服务分组不存在") {
			response.ErrorJSON(ctx, "指定的服务分组不存在或已禁用", constants.ED00008)
			return
		}
		logger.ErrorWithTrace(ctx, "创建服务失败", err)
		response.ErrorJSON(ctx, "创建服务失败", constants.ED00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务创建成功",
		"serviceName", createdService.ServiceName,
		"serviceGroupId", createdService.ServiceGroupId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, createdService, constants.SD00001)
}

// GetServiceGroups 获取服务分组列表（命名空间列表）
// @Summary 获取服务分组列表
// @Description 获取租户下的所有服务分组（命名空间）列表
// @Tags 服务注册管理
// @Accept json
// @Produce json
// @Param request body object{activeFlag=string} false "查询请求"
// @Success 200 {object} response.JsonData{data=[]models.ServiceGroup}
// @Router /gateway/hub0041/getServiceGroups [post]
func (c *ServiceController) GetServiceGroups(ctx *gin.Context) {
	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 获取过滤参数
	activeFlag := request.GetParam(ctx, "activeFlag")
	if activeFlag == "" {
		activeFlag = "Y" // 默认只显示活动的分组
	}

	// 调用DAO查询服务分组列表
	serviceGroups, err := c.serviceGroupDAO.GetServiceGroups(ctx, tenantId, activeFlag)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务分组列表失败", err)
		response.ErrorJSON(ctx, "查询服务分组列表失败", constants.ED00003)
		return
	}

	logger.DebugWithTrace(ctx, "获取服务分组列表成功",
		"tenantId", tenantId,
		"activeFlag", activeFlag,
		"count", len(serviceGroups))

	response.SuccessJSON(ctx, serviceGroups, constants.SD00002)
}
