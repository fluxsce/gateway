package controllers

import (
	"time"

	"gateway/internal/servicecenter"
	"gateway/internal/servicecenter/cache"
	"gateway/internal/servicecenter/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0042/dao"
	"gateway/web/views/hub0042/models"

	"github.com/gin-gonic/gin"
)

// ServiceController 服务控制器
type ServiceController struct {
	db         database.Database
	serviceDAO *dao.ServiceDAO
}

// NewServiceController 创建服务控制器
func NewServiceController(db database.Database) *ServiceController {
	return &ServiceController{
		db:         db,
		serviceDAO: dao.NewServiceDAO(db),
	}
}

// QueryServices 获取服务列表
// @Summary 获取服务列表
// @Description 分页获取服务列表，支持条件查询
// @Tags 服务监控
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param serviceName query string false "服务名称（模糊查询）"
// @Param namespaceId query string false "命名空间ID"
// @Param groupName query string false "分组名称"
// @Param serviceType query string false "服务类型（INTERNAL, NACOS, CONSUL, EUREKA, ETCD, ZOOKEEPER）"
// @Param instanceName query string false "服务中心实例名称"
// @Param environment query string false "部署环境（DEVELOPMENT, STAGING, PRODUCTION）"
// @Param activeFlag query string false "活动状态（Y/N）"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [get]
func (c *ServiceController) QueryServices(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)
	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.ServiceQuery
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定服务查询条件失败，使用默认条件", "error", err.Error())
	}

	// 调用DAO获取服务列表
	services, total, err := c.serviceDAO.ListServices(ctx, tenantId, &query, page, pageSize)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务列表失败", err)
		response.ErrorJSON(ctx, "获取服务列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 构建服务列表
	serviceList := make([]map[string]interface{}, 0, len(services))

	for _, service := range services {
		serviceInfo := map[string]interface{}{
			"tenantId":           service.TenantId,
			"namespaceId":        service.NamespaceId,
			"groupName":          service.GroupName,
			"serviceName":        service.ServiceName,
			"serviceType":        service.ServiceType,
			"serviceVersion":     service.ServiceVersion,
			"serviceDescription": service.ServiceDescription,
			"protectThreshold":   service.ProtectThreshold,
			"activeFlag":         service.ActiveFlag,
			"addTime":            service.AddTime,
			"addWho":             service.AddWho,
			"editTime":           service.EditTime,
			"editWho":            service.EditWho,
			"currentVersion":     service.CurrentVersion,
			"noteText":           service.NoteText,
			"nodeCount":          0,
			"healthyNodeCount":   0,
			"unhealthyNodeCount": 0,
		}

		// 从缓存中获取服务节点信息
		globalCache := cache.GetGlobalCache()
		if globalCache != nil {
			// 获取服务节点列表
			nodes, found := globalCache.GetNodes(ctx, tenantId, service.NamespaceId, service.GroupName, service.ServiceName)
			if found && nodes != nil {
				serviceInfo["nodeCount"] = len(nodes)
				healthyCount := 0
				unhealthyCount := 0
				for _, node := range nodes {
					if node.HealthyStatus == "HEALTHY" {
						healthyCount++
					} else if node.HealthyStatus == "UNHEALTHY" {
						unhealthyCount++
					}
				}
				serviceInfo["healthyNodeCount"] = healthyCount
				serviceInfo["unhealthyNodeCount"] = unhealthyCount
			}
		}

		serviceList = append(serviceList, serviceInfo)
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceName"

	// 使用统一的分页响应
	response.PageJSON(ctx, serviceList, pageInfo, constants.SD00002)
}

// GetService 获取单个服务详情
// @Summary 获取服务详情
// @Description 根据服务主键获取服务详细信息，包括节点列表
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param serviceName query string true "服务名称"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services/detail [get]
func (c *ServiceController) GetService(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	serviceName := request.GetParam(ctx, "serviceName")

	// 使用工具类获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 验证必填字段
	if namespaceId == "" || groupName == "" || serviceName == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和serviceName不能为空", constants.ED00006)
		return
	}

	// 调用DAO获取服务详情
	service, err := c.serviceDAO.GetServiceById(ctx, tenantId, namespaceId, groupName, serviceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务详情失败", err)
		response.ErrorJSON(ctx, "获取服务详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if service == nil {
		response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
		return
	}

	// 构建服务详情响应
	serviceInfo := map[string]interface{}{
		"tenantId":              service.TenantId,
		"namespaceId":           service.NamespaceId,
		"groupName":             service.GroupName,
		"serviceName":           service.ServiceName,
		"serviceType":           service.ServiceType,
		"serviceVersion":        service.ServiceVersion,
		"serviceDescription":    service.ServiceDescription,
		"externalServiceConfig": service.ExternalServiceConfig,
		"metadataJson":          service.MetadataJson,
		"tagsJson":              service.TagsJson,
		"protectThreshold":      service.ProtectThreshold,
		"selectorJson":          service.SelectorJson,
		"activeFlag":            service.ActiveFlag,
		"addTime":               service.AddTime,
		"addWho":                service.AddWho,
		"editTime":              service.EditTime,
		"editWho":               service.EditWho,
		"currentVersion":        service.CurrentVersion,
		"noteText":              service.NoteText,
		"extProperty":           service.ExtProperty,
		"nodes":                 []interface{}{},
		"nodeCount":             0,
		"healthyNodeCount":      0,
		"unhealthyNodeCount":    0,
	}

	// 从缓存中获取服务节点信息
	globalCache := cache.GetGlobalCache()
	if globalCache != nil {
		// 获取服务节点列表
		nodes, found := globalCache.GetNodes(ctx, tenantId, namespaceId, groupName, serviceName)
		if found && nodes != nil {
			healthyCount := 0
			unhealthyCount := 0
			for _, node := range nodes {
				if node.HealthyStatus == "HEALTHY" {
					healthyCount++
				} else if node.HealthyStatus == "UNHEALTHY" {
					unhealthyCount++
				}
			}
			// 直接使用节点列表（结构体有完整的 JSON tag）
			serviceInfo["nodes"] = nodes
			serviceInfo["nodeCount"] = len(nodes)
			serviceInfo["healthyNodeCount"] = healthyCount
			serviceInfo["unhealthyNodeCount"] = unhealthyCount
		}
	}

	// 直接返回服务对象
	response.SuccessJSON(ctx, serviceInfo, constants.SD00001)
}

// AddService 创建服务
// @Summary 创建服务
// @Description 创建新的服务
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param service body types.Service true "服务信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [post]
func (c *ServiceController) AddService(ctx *gin.Context) {
	var req types.Service
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID，不使用前端传递的值（前置校验已保证非空）
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceId == "" {
		response.ErrorJSON(ctx, "命名空间ID不能为空", constants.ED00006)
		return
	}
	if req.ServiceName == "" {
		response.ErrorJSON(ctx, "服务名称不能为空", constants.ED00006)
		return
	}

	// 设置默认值
	if req.GroupName == "" {
		req.GroupName = "DEFAULT_GROUP"
	}
	if req.ServiceType == "" {
		req.ServiceType = "INTERNAL"
	}

	// 检查服务是否已存在
	existingService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "检查服务是否存在时出错", err)
		response.ErrorJSON(ctx, "检查服务是否存在失败: "+err.Error(), constants.ED00009)
		return
	}
	if existingService != nil {
		response.ErrorJSON(ctx, "服务已存在，服务名称: "+req.ServiceName, constants.ED00008)
		return
	}

	// 只设置租户ID，其他默认参数（新增人、新增时间等）由DAO处理
	req.TenantId = tenantId

	// 调用DAO添加服务
	err = c.serviceDAO.AddService(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建服务失败", err)
		response.ErrorJSON(ctx, "创建服务失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的服务信息
	newService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务信息失败", err)
		// 即使查询失败，也返回成功但只带有基本信息
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"groupName":   req.GroupName,
			"serviceName": req.ServiceName,
			"message":     "服务创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newService == nil {
		logger.ErrorWithTrace(ctx, "新创建的服务不存在", "serviceName", req.ServiceName)
		response.SuccessJSON(ctx, gin.H{
			"namespaceId": req.NamespaceId,
			"groupName":   req.GroupName,
			"serviceName": req.ServiceName,
			"message":     "服务创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	logger.InfoWithTrace(ctx, "服务创建成功",
		"namespaceId", newService.NamespaceId,
		"groupName", newService.GroupName,
		"serviceName", newService.ServiceName,
		"tenantId", tenantId,
		"operatorId", operatorId)

	// 同步到缓存（AddServiceToCache 会自动处理节点）
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.AddServiceToCache(ctx, newService); err != nil {
			logger.WarnWithTrace(ctx, "添加服务到缓存失败", "error", err,
				"namespaceId", newService.NamespaceId,
				"groupName", newService.GroupName,
				"serviceName", newService.ServiceName)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	// 直接返回服务对象
	response.SuccessJSON(ctx, newService, constants.SD00003)
}

// EditService 更新服务
// @Summary 更新服务
// @Description 更新服务信息
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param service body types.Service true "服务信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [put]
func (c *ServiceController) EditService(ctx *gin.Context) {
	var req types.Service
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NamespaceId == "" || req.GroupName == "" || req.ServiceName == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和serviceName不能为空", constants.ED00006)
		return
	}

	// 获取现有服务信息进行校验
	currentService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务信息失败", err)
		response.ErrorJSON(ctx, "获取服务信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentService == nil {
		response.ErrorJSON(ctx, "服务不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段，确保关键字段不被前端覆盖
	req.TenantId = currentService.TenantId
	req.NamespaceId = currentService.NamespaceId
	req.GroupName = currentService.GroupName
	req.ServiceName = currentService.ServiceName

	// 调用DAO更新服务（DAO会处理EditTime和EditWho）
	err = c.serviceDAO.UpdateService(ctx, &req, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务失败", err)
		response.ErrorJSON(ctx, "更新服务失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的服务信息
	updatedService, err := c.serviceDAO.GetServiceById(ctx, tenantId, req.NamespaceId, req.GroupName, req.ServiceName)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务信息失败", err)
		// 即使查询失败，也返回成功但只带有简单消息
		response.SuccessJSON(ctx, gin.H{
			"message": "更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 同步到缓存（UpdateServiceInCache 会自动处理节点）
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.UpdateServiceInCache(ctx, updatedService); err != nil {
			logger.WarnWithTrace(ctx, "更新服务缓存失败", "error", err,
				"namespaceId", req.NamespaceId,
				"groupName", req.GroupName,
				"serviceName", req.ServiceName)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	// 直接返回服务对象
	response.SuccessJSON(ctx, updatedService, constants.SD00004)
}

// DeleteService 删除服务
// @Summary 删除服务
// @Description 删除服务
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param namespaceId query string true "命名空间ID"
// @Param groupName query string true "分组名称"
// @Param serviceName query string true "服务名称"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/services [delete]
func (c *ServiceController) DeleteService(ctx *gin.Context) {
	namespaceId := request.GetParam(ctx, "namespaceId")
	groupName := request.GetParam(ctx, "groupName")
	serviceName := request.GetParam(ctx, "serviceName")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if namespaceId == "" || groupName == "" || serviceName == "" {
		response.ErrorJSON(ctx, "namespaceId、groupName和serviceName不能为空", constants.ED00006)
		return
	}

	// 调用DAO删除服务
	err := c.serviceDAO.DeleteService(ctx, tenantId, namespaceId, groupName, serviceName, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务失败", err)
		response.ErrorJSON(ctx, "删除服务失败: "+err.Error(), constants.ED00009)
		return
	}

	// 同步删除缓存
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager != nil {
		if err := serviceCenterManager.DeleteServiceFromCache(ctx, tenantId, namespaceId, groupName, serviceName); err != nil {
			logger.WarnWithTrace(ctx, "删除服务缓存失败", "error", err,
				"namespaceId", namespaceId,
				"groupName", groupName,
				"serviceName", serviceName)
			// 缓存失败不影响主流程，只记录警告
		}
	}

	response.SuccessJSON(ctx, gin.H{
		"namespaceId": namespaceId,
		"groupName":   groupName,
		"serviceName": serviceName,
		"message":     "服务删除成功",
	}, constants.SD00005)
}

// EditNode 编辑节点
// @Summary 编辑节点
// @Description 更新服务节点信息（如IP、端口、权重、元数据等），直接操作缓存，不操作数据库
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param node body types.ServiceNode true "节点信息"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/nodes [put]
func (c *ServiceController) EditNode(ctx *gin.Context) {
	var req types.ServiceNode
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if req.NodeId == "" {
		response.ErrorJSON(ctx, "nodeId不能为空", constants.ED00006)
		return
	}

	// 获取 ServiceCenterManager
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager == nil {
		response.ErrorJSON(ctx, "服务中心管理器未初始化", constants.ED00009)
		return
	}

	// 从缓存获取节点信息（用于验证和合并字段）
	globalCache := cache.GetGlobalCache()
	if globalCache == nil {
		response.ErrorJSON(ctx, "缓存未初始化", constants.ED00009)
		return
	}

	// 通过 nodeId 获取节点
	currentNode, found := globalCache.GetNode(ctx, tenantId, req.NodeId)
	if !found || currentNode == nil {
		response.ErrorJSON(ctx, "节点不存在", constants.ED00008)
		return
	}

	// 保留不可修改的字段
	req.TenantId = currentNode.TenantId
	req.NamespaceId = currentNode.NamespaceId
	req.GroupName = currentNode.GroupName
	req.ServiceName = currentNode.ServiceName
	req.NodeId = currentNode.NodeId
	req.RegisterTime = currentNode.RegisterTime
	req.AddTime = currentNode.AddTime
	req.AddWho = currentNode.AddWho

	// 设置编辑信息
	req.EditTime = time.Now()
	req.EditWho = operatorId

	// 如果未提供某些字段，使用原值
	if req.IpAddress == "" {
		req.IpAddress = currentNode.IpAddress
	}
	if req.PortNumber == 0 {
		req.PortNumber = currentNode.PortNumber
	}
	if req.Weight == 0 {
		req.Weight = currentNode.Weight
	}
	if req.InstanceStatus == "" {
		req.InstanceStatus = currentNode.InstanceStatus
	}
	if req.HealthyStatus == "" {
		req.HealthyStatus = currentNode.HealthyStatus
	}
	if req.Ephemeral == "" {
		req.Ephemeral = currentNode.Ephemeral
	}
	if req.MetadataJson == "" {
		req.MetadataJson = currentNode.MetadataJson
	}
	if req.ActiveFlag == "" {
		req.ActiveFlag = currentNode.ActiveFlag
	}

	// 通过 ServiceCenterManager 更新节点缓存（不操作数据库，由外部异步同步服务负责持久化）
	if err := serviceCenterManager.UpdateNodeInCache(ctx, &req); err != nil {
		logger.ErrorWithTrace(ctx, "更新节点缓存失败", err, "nodeId", req.NodeId)
		response.ErrorJSON(ctx, "更新节点失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "节点编辑成功（仅更新缓存）",
		"nodeId", req.NodeId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	// 直接返回节点对象（结构体有完整的 JSON tag）
	response.SuccessJSON(ctx, req, constants.SD00004)
}

// OfflineNode 下线节点
// @Summary 下线节点
// @Description 将服务节点下线（设置状态为DOWN），直接操作缓存，不操作数据库
// @Tags 服务监控
// @Accept json
// @Produce json
// @Param nodeId query string true "节点ID"
// @Success 200 {object} response.JsonData
// @Router /api/hub0042/nodes/offline [post]
func (c *ServiceController) OfflineNode(ctx *gin.Context) {
	nodeId := request.GetParam(ctx, "nodeId")

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 验证必填字段
	if nodeId == "" {
		response.ErrorJSON(ctx, "nodeId不能为空", constants.ED00006)
		return
	}

	// 获取 ServiceCenterManager
	serviceCenterManager := servicecenter.GetManager()
	if serviceCenterManager == nil {
		response.ErrorJSON(ctx, "服务中心管理器未初始化", constants.ED00009)
		return
	}

	// 通过 ServiceCenterManager 下线节点（不操作数据库，由外部异步同步服务负责持久化）
	if err := serviceCenterManager.OfflineNodeInCache(ctx, tenantId, nodeId, operatorId); err != nil {
		logger.ErrorWithTrace(ctx, "下线节点失败", err, "nodeId", nodeId)
		response.ErrorJSON(ctx, "下线节点失败: "+err.Error(), constants.ED00009)
		return
	}

	// 获取更新后的节点信息用于返回
	globalCache := cache.GetGlobalCache()
	if globalCache == nil {
		response.ErrorJSON(ctx, "缓存未初始化", constants.ED00009)
		return
	}

	currentNode, found := globalCache.GetNode(ctx, tenantId, nodeId)
	if !found || currentNode == nil {
		response.ErrorJSON(ctx, "节点不存在", constants.ED00008)
		return
	}

	logger.InfoWithTrace(ctx, "节点下线成功（仅更新缓存）",
		"nodeId", nodeId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"namespaceId", currentNode.NamespaceId,
		"groupName", currentNode.GroupName,
		"serviceName", currentNode.ServiceName)

	// 直接返回节点对象（结构体有完整的 JSON tag）
	response.SuccessJSON(ctx, currentNode, constants.SD00004)
}
