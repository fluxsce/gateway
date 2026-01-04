package controllers

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/constants"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0022/dao"
	"gateway/web/views/hub0022/models"

	"github.com/gin-gonic/gin"
)

// ServiceNodeController 服务节点控制器
type ServiceNodeController struct {
	db             database.Database
	serviceNodeDAO *dao.ServiceNodeDAO
}

// NewServiceNodeController 创建服务节点控制器
func NewServiceNodeController(db database.Database) *ServiceNodeController {
	return &ServiceNodeController{
		db:             db,
		serviceNodeDAO: dao.NewServiceNodeDAO(db),
	}
}

// QueryServiceNodes 获取服务节点列表
// @Summary 获取服务节点列表
// @Description 分页获取服务节点列表
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param request body QueryServiceNodesRequest true "查询参数"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/queryServiceNodes [post]
func (c *ServiceNodeController) QueryServiceNodes(ctx *gin.Context) {
	var req QueryServiceNodesRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 从上下文获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 构建筛选条件
	filters := make(map[string]interface{})
	if req.ServiceDefinitionId != "" {
		filters["serviceDefinitionId"] = req.ServiceDefinitionId
	}
	if req.NodeHost != "" {
		filters["nodeHost"] = req.NodeHost
	}
	if req.HealthStatus != "" {
		filters["healthStatus"] = req.HealthStatus
	}

	// 调用DAO获取服务节点列表
	nodes, total, err := c.serviceNodeDAO.QueryServiceNodes(ctx, tenantId, page, pageSize, filters)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务节点列表失败", err)
		response.ErrorJSON(ctx, "获取服务节点列表失败: "+err.Error(), constants.ED00009)
		return
	}

	// 转换为响应格式
	nodeList := make([]map[string]interface{}, 0, len(nodes))
	for _, node := range nodes {
		nodeList = append(nodeList, serviceNodeToMap(node))
	}

	// 创建分页信息并返回
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "serviceNodeId"

	// 使用统一的分页响应
	response.PageJSON(ctx, nodeList, pageInfo, constants.SD00002)
}

// AddServiceNode 创建服务节点
// @Summary 创建服务节点
// @Description 创建新的服务节点
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param serviceNode body models.ServiceNodeModel true "服务节点信息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/addServiceNode [post]
func (c *ServiceNodeController) AddServiceNode(ctx *gin.Context) {
	var serviceNode models.ServiceNodeModel
	if err := request.BindSafely(ctx, &serviceNode); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 获取操作人信息和租户信息
	operatorId := request.GetOperatorID(ctx)
	tenantId := request.GetTenantID(ctx)

	// 设置租户ID
	serviceNode.TenantId = tenantId

	// 验证必填字段
	if serviceNode.ServiceDefinitionId == "" {
		response.ErrorJSON(ctx, "服务定义ID不能为空", constants.ED00007)
		return
	}
	if serviceNode.NodeHost == "" || serviceNode.NodePort == 0 {
		response.ErrorJSON(ctx, "节点主机地址和端口不能为空", constants.ED00007)
		return
	}

	// 创建服务节点
	serviceNodeId, err := c.serviceNodeDAO.CreateServiceNode(ctx, &serviceNode, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "创建服务节点失败", err)
		response.ErrorJSON(ctx, "创建服务节点失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询新添加的服务节点信息
	newServiceNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, serviceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取新创建的服务节点信息失败", err)
		// 即使查询失败，也返回成功但只带有服务节点ID
		response.SuccessJSON(ctx, gin.H{
			"serviceNodeId": serviceNodeId,
			"tenantId":      tenantId,
			"message":       "服务节点创建成功，但获取详细信息失败",
		}, constants.SD00003)
		return
	}

	if newServiceNode == nil {
		logger.ErrorWithTrace(ctx, "新创建的服务节点不存在", "serviceNodeId", serviceNodeId)
		response.SuccessJSON(ctx, gin.H{
			"serviceNodeId": serviceNodeId,
			"tenantId":      tenantId,
			"message":       "服务节点创建成功，但查询详细信息为空",
		}, constants.SD00003)
		return
	}

	// 返回完整的服务节点信息
	nodeInfo := serviceNodeToMap(newServiceNode)

	logger.InfoWithTrace(ctx, "服务节点创建成功",
		"serviceNodeId", serviceNodeId,
		"serviceDefinitionId", newServiceNode.ServiceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"nodeHost", newServiceNode.NodeHost,
		"nodePort", newServiceNode.NodePort)

	response.SuccessJSON(ctx, nodeInfo, constants.SD00003)
}

// EditServiceNode 更新服务节点
// @Summary 更新服务节点
// @Description 更新服务节点信息
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param serviceNode body models.ServiceNodeModel true "服务节点信息"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/updateServiceNode [post]
func (c *ServiceNodeController) EditServiceNode(ctx *gin.Context) {
	var updateData models.ServiceNodeModel
	if err := request.BindSafely(ctx, &updateData); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if updateData.ServiceNodeId == "" {
		response.ErrorJSON(ctx, "服务节点ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 获取现有服务节点信息
	currentServiceNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, updateData.ServiceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务节点信息失败", err)
		response.ErrorJSON(ctx, "获取服务节点信息失败: "+err.Error(), constants.ED00009)
		return
	}

	if currentServiceNode == nil {
		response.ErrorJSON(ctx, "服务节点不存在", constants.ED00008)
		return
	}

	// 设置租户ID
	updateData.TenantId = tenantId

	// 调用DAO更新服务节点
	err = c.serviceNodeDAO.UpdateServiceNode(ctx, &updateData, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新服务节点失败", err)
		response.ErrorJSON(ctx, "更新服务节点失败: "+err.Error(), constants.ED00009)
		return
	}

	// 查询更新后的服务节点信息
	updatedServiceNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, updateData.ServiceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的服务节点信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"serviceNodeId": updateData.ServiceNodeId,
			"message":       "服务节点更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回更新后的服务节点信息
	nodeInfo := serviceNodeToMap(updatedServiceNode)

	logger.InfoWithTrace(ctx, "服务节点更新成功",
		"serviceNodeId", updateData.ServiceNodeId,
		"serviceDefinitionId", updatedServiceNode.ServiceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId)

	response.SuccessJSON(ctx, nodeInfo, constants.SD00004)
}

// DeleteServiceNode 删除服务节点
// @Summary 删除服务节点
// @Description 删除服务节点
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param request body DeleteServiceNodeRequest true "删除请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/deleteServiceNode [post]
func (c *ServiceNodeController) DeleteServiceNode(ctx *gin.Context) {
	var req DeleteServiceNodeRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ServiceNodeId == "" {
		response.ErrorJSON(ctx, "服务节点ID不能为空", constants.ED00007)
		return
	}

	// 强制从上下文获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 先查询服务节点是否存在
	existingServiceNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, req.ServiceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询服务节点失败", err)
		response.ErrorJSON(ctx, "查询服务节点失败: "+err.Error(), constants.ED00009)
		return
	}

	if existingServiceNode == nil {
		response.ErrorJSON(ctx, "服务节点不存在", constants.ED00008)
		return
	}

	// 调用DAO删除服务节点
	err = c.serviceNodeDAO.DeleteServiceNode(ctx, req.ServiceNodeId, tenantId, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "删除服务节点失败", err)
		response.ErrorJSON(ctx, "删除服务节点失败: "+err.Error(), constants.ED00009)
		return
	}

	logger.InfoWithTrace(ctx, "服务节点删除成功",
		"serviceNodeId", req.ServiceNodeId,
		"serviceDefinitionId", existingServiceNode.ServiceDefinitionId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"nodeHost", existingServiceNode.NodeHost,
		"nodePort", existingServiceNode.NodePort)

	response.SuccessJSON(ctx, gin.H{
		"serviceNodeId": req.ServiceNodeId,
		"message":       "服务节点删除成功",
	}, constants.SD00005)
}

// GetServiceNode 获取服务节点详情
// @Summary 获取服务节点详情
// @Description 根据ID获取服务节点详情
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param request body GetServiceNodeRequest true "查询请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/getServiceNode [post]
func (c *ServiceNodeController) GetServiceNode(ctx *gin.Context) {
	var req GetServiceNodeRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ServiceNodeId == "" {
		response.ErrorJSON(ctx, "服务节点ID不能为空", constants.ED00007)
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)

	// 调用DAO获取服务节点详情
	serviceNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, req.ServiceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取服务节点详情失败", err)
		response.ErrorJSON(ctx, "获取服务节点详情失败: "+err.Error(), constants.ED00009)
		return
	}

	if serviceNode == nil {
		response.ErrorJSON(ctx, "服务节点不存在", constants.ED00008)
		return
	}

	// 转换为响应格式
	nodeInfo := serviceNodeToMap(serviceNode)

	response.SuccessJSON(ctx, nodeInfo, constants.SD00002)
}

// UpdateNodeHealth 更新节点健康状态
// @Summary 更新节点健康状态
// @Description 更新节点的健康状态
// @Tags 服务节点管理
// @Accept json
// @Produce json
// @Param request body UpdateNodeHealthRequest true "更新请求"
// @Success 200 {object} response.JsonData
// @Router /gateway/hub0022/updateNodeHealth [post]
func (c *ServiceNodeController) UpdateNodeHealth(ctx *gin.Context) {
	var req UpdateNodeHealthRequest
	if err := request.BindSafely(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数错误: "+err.Error(), constants.ED00006)
		return
	}

	// 验证必填字段
	if req.ServiceNodeId == "" {
		response.ErrorJSON(ctx, "服务节点ID不能为空", constants.ED00007)
		return
	}
	if req.HealthStatus == "" {
		response.ErrorJSON(ctx, "健康状态不能为空", constants.ED00007)
		return
	}

	// 获取租户ID和操作人ID
	tenantId := request.GetTenantID(ctx)
	operatorId := request.GetOperatorID(ctx)

	// 调用DAO更新节点健康状态
	err := c.serviceNodeDAO.UpdateNodeHealth(ctx, req.ServiceNodeId, tenantId, req.HealthStatus, req.HealthCheckResult, operatorId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "更新节点健康状态失败", err)
		response.ErrorJSON(ctx, "更新节点健康状态失败: "+err.Error(), constants.ED00009)
		return
	}

	// 获取更新后的节点信息
	updatedNode, err := c.serviceNodeDAO.GetServiceNodeById(ctx, req.ServiceNodeId, tenantId)
	if err != nil {
		logger.ErrorWithTrace(ctx, "获取更新后的节点信息失败", err)
		response.SuccessJSON(ctx, gin.H{
			"serviceNodeId": req.ServiceNodeId,
			"message":       "节点健康状态更新成功，但获取详细信息失败",
		}, constants.SD00004)
		return
	}

	// 返回更新后的节点信息
	nodeInfo := serviceNodeToMap(updatedNode)

	logger.InfoWithTrace(ctx, "节点健康状态更新成功",
		"serviceNodeId", req.ServiceNodeId,
		"tenantId", tenantId,
		"operatorId", operatorId,
		"healthStatus", req.HealthStatus)

	response.SuccessJSON(ctx, nodeInfo, constants.SD00004)
}

// 请求结构体定义

// QueryServiceNodesRequest 查询服务节点列表请求
type QueryServiceNodesRequest struct {
	ServiceDefinitionId string `json:"serviceDefinitionId" form:"serviceDefinitionId" query:"serviceDefinitionId"` // 服务定义ID
	NodeHost            string `json:"nodeHost" form:"nodeHost" query:"nodeHost"`                                  // 节点主机地址
	HealthStatus        string `json:"healthStatus" form:"healthStatus" query:"healthStatus"`                      // 健康状态
}

// DeleteServiceNodeRequest 删除服务节点请求
type DeleteServiceNodeRequest struct {
	ServiceNodeId string `json:"serviceNodeId" form:"serviceNodeId" query:"serviceNodeId" binding:"required"` // 服务节点ID
}

// GetServiceNodeRequest 获取服务节点请求
type GetServiceNodeRequest struct {
	ServiceNodeId string `json:"serviceNodeId" form:"serviceNodeId" query:"serviceNodeId" binding:"required"` // 服务节点ID
}

// UpdateNodeHealthRequest 更新节点健康状态请求
type UpdateNodeHealthRequest struct {
	ServiceNodeId     string `json:"serviceNodeId" form:"serviceNodeId" query:"serviceNodeId" binding:"required"` // 服务节点ID
	HealthStatus      string `json:"healthStatus" form:"healthStatus" query:"healthStatus" binding:"required"`    // 健康状态(N不健康,Y健康)
	HealthCheckResult string `json:"healthCheckResult" form:"healthCheckResult" query:"healthCheckResult"`        // 健康检查结果详情
}

// serviceNodeToMap 将服务节点转换为Map格式
func serviceNodeToMap(serviceNode *models.ServiceNodeModel) map[string]interface{} {
	return map[string]interface{}{
		"tenantId":            serviceNode.TenantId,
		"serviceNodeId":       serviceNode.ServiceNodeId,
		"serviceDefinitionId": serviceNode.ServiceDefinitionId,
		"nodeId":              serviceNode.NodeId,
		"nodeUrl":             serviceNode.NodeUrl,
		"nodeHost":            serviceNode.NodeHost,
		"nodePort":            serviceNode.NodePort,
		"nodeProtocol":        serviceNode.NodeProtocol,
		"nodeWeight":          serviceNode.NodeWeight,
		"healthStatus":        serviceNode.HealthStatus,
		"nodeMetadata":        serviceNode.NodeMetadata,
		"nodeStatus":          serviceNode.NodeStatus,
		"lastHealthCheckTime": serviceNode.LastHealthCheckTime,
		"healthCheckResult":   serviceNode.HealthCheckResult,
		"activeFlag":          serviceNode.ActiveFlag,
		"addTime":             serviceNode.AddTime,
		"addWho":              serviceNode.AddWho,
		"editTime":            serviceNode.EditTime,
		"editWho":             serviceNode.EditWho,
		"currentVersion":      serviceNode.CurrentVersion,
		"noteText":            serviceNode.NoteText,
	}
}
