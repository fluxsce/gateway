package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0061/dao"
	"gateway/web/views/hub0061/models"

	"github.com/gin-gonic/gin"
)

// ServerNodeController 服务器节点控制器
type ServerNodeController struct {
	serverNodeDAO *dao.ServerNodeDAO
}

// NewServerNodeController 创建服务器节点控制器实例
func NewServerNodeController(db database.Database) *ServerNodeController {
	return &ServerNodeController{
		serverNodeDAO: dao.NewServerNodeDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *ServerNodeController) getCurrentUser(ctx *gin.Context) string {
	if userName := request.GetUserName(ctx); userName != "" {
		return userName
	}
	if userID := request.GetUserID(ctx); userID != "" {
		return userID
	}
	return "admin"
}

// QueryServerNodes 查询服务器节点列表
func (c *ServerNodeController) QueryServerNodes(ctx *gin.Context) {
	var req models.ServerNodeQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定查询参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_SERVER_NODES")
		return
	}

	// 参数验证
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	nodes, total, err := c.serverNodeDAO.QueryServerNodes(&req)
	if err != nil {
		logger.Error("查询节点列表失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "QUERY_SERVER_NODES")
		return
	}

	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	response.PageJSON(ctx, nodes, pageInfo, "QUERY_SERVER_NODES")
}

// GetServerNode 获取服务器节点详情
func (c *ServerNodeController) GetServerNode(ctx *gin.Context) {
	serverNodeId := request.GetParam(ctx, "serverNodeId")
	if serverNodeId == "" {
		response.ErrorJSON(ctx, "参数格式错误: serverNodeId不能为空", "GET_SERVER_NODE")
		return
	}

	node, err := c.serverNodeDAO.GetServerNode(serverNodeId)
	if err != nil {
		logger.Error("获取节点详情失败", "serverNodeId", serverNodeId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_SERVER_NODE")
		return
	}

	response.SuccessJSON(ctx, node, "GET_SERVER_NODE")
}

// CreateServerNode 创建服务器节点
func (c *ServerNodeController) CreateServerNode(ctx *gin.Context) {
	var node models.TunnelServerNode
	if err := request.Bind(ctx, &node); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_SERVER_NODE")
		return
	}

	// 参数验证
	if strings.TrimSpace(node.NodeName) == "" {
		response.ErrorJSON(ctx, "节点名称不能为空", "CREATE_SERVER_NODE")
		return
	}
	if strings.TrimSpace(node.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "服务器ID不能为空", "CREATE_SERVER_NODE")
		return
	}
	if strings.TrimSpace(node.ProxyType) == "" {
		response.ErrorJSON(ctx, "代理类型不能为空", "CREATE_SERVER_NODE")
		return
	}
	if node.ListenPort <= 0 || node.ListenPort > 65535 {
		response.ErrorJSON(ctx, "监听端口必须在1-65535之间", "CREATE_SERVER_NODE")
		return
	}
	if strings.TrimSpace(node.TargetAddress) == "" {
		response.ErrorJSON(ctx, "目标地址不能为空", "CREATE_SERVER_NODE")
		return
	}
	if node.TargetPort <= 0 || node.TargetPort > 65535 {
		response.ErrorJSON(ctx, "目标端口必须在1-65535之间", "CREATE_SERVER_NODE")
		return
	}

	// 设置租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", "CREATE_SERVER_NODE")
		return
	}
	node.TenantId = tenantId

	// 生成ID和审计字段
	node.ServerNodeId = random.Generate32BitRandomString()
	node.AddWho = c.getCurrentUser(ctx)
	node.EditWho = node.AddWho
	node.OprSeqFlag = random.Generate32BitRandomString()

	createdNode, err := c.serverNodeDAO.CreateServerNode(&node)
	if err != nil {
		logger.Error("创建节点失败", "error", err)
		response.ErrorJSON(ctx, "创建失败: "+err.Error(), "CREATE_SERVER_NODE")
		return
	}

	logger.Info("创建服务器节点成功", "serverNodeId", createdNode.ServerNodeId, "nodeName", createdNode.NodeName)
	response.SuccessJSON(ctx, createdNode, "CREATE_SERVER_NODE")
}

// UpdateServerNode 更新服务器节点
func (c *ServerNodeController) UpdateServerNode(ctx *gin.Context) {
	var node models.TunnelServerNode
	if err := request.Bind(ctx, &node); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_SERVER_NODE")
		return
	}

	// 参数验证
	if strings.TrimSpace(node.ServerNodeId) == "" {
		response.ErrorJSON(ctx, "节点ID不能为空", "UPDATE_SERVER_NODE")
		return
	}
	if strings.TrimSpace(node.NodeName) == "" {
		response.ErrorJSON(ctx, "节点名称不能为空", "UPDATE_SERVER_NODE")
		return
	}
	if node.ListenPort <= 0 || node.ListenPort > 65535 {
		response.ErrorJSON(ctx, "监听端口必须在1-65535之间", "UPDATE_SERVER_NODE")
		return
	}
	if node.TargetPort <= 0 || node.TargetPort > 65535 {
		response.ErrorJSON(ctx, "目标端口必须在1-65535之间", "UPDATE_SERVER_NODE")
		return
	}

	node.EditWho = c.getCurrentUser(ctx)

	updatedNode, err := c.serverNodeDAO.UpdateServerNode(&node)
	if err != nil {
		logger.Error("更新节点失败", "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_SERVER_NODE")
		return
	}

	logger.Info("更新服务器节点成功", "serverNodeId", updatedNode.ServerNodeId, "nodeName", updatedNode.NodeName)
	response.SuccessJSON(ctx, updatedNode, "UPDATE_SERVER_NODE")
}

// DeleteServerNode 删除服务器节点
func (c *ServerNodeController) DeleteServerNode(ctx *gin.Context) {
	serverNodeId := request.GetParam(ctx, "serverNodeId")
	if serverNodeId == "" {
		response.ErrorJSON(ctx, "参数格式错误: serverNodeId不能为空", "DELETE_SERVER_NODE")
		return
	}

	editWho := c.getCurrentUser(ctx)
	deletedNode, err := c.serverNodeDAO.DeleteServerNode(serverNodeId, editWho)
	if err != nil {
		logger.Error("删除节点失败", "serverNodeId", serverNodeId, "error", err)
		response.ErrorJSON(ctx, "删除失败: "+err.Error(), "DELETE_SERVER_NODE")
		return
	}

	logger.Info("删除服务器节点成功", "serverNodeId", serverNodeId)
	response.SuccessJSON(ctx, deletedNode, "DELETE_SERVER_NODE")
}

// GetNodeStats 获取节点统计信息
func (c *ServerNodeController) GetNodeStats(ctx *gin.Context) {
	stats, err := c.serverNodeDAO.GetNodeStats()
	if err != nil {
		logger.Error("获取节点统计信息失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_NODE_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_NODE_STATS")
}

// CheckPortConflict 检查端口冲突
func (c *ServerNodeController) CheckPortConflict(ctx *gin.Context) {
	type Request struct {
		ListenAddress string `json:"listenAddress" binding:"required"`
		ListenPort    int    `json:"listenPort" binding:"required"`
		ProxyType     string `json:"proxyType" binding:"required"`
		ExcludeId     string `json:"excludeId"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CHECK_PORT_CONFLICT")
		return
	}

	conflict, err := c.serverNodeDAO.CheckPortConflict(req.ListenAddress, req.ListenPort, req.ProxyType, req.ExcludeId)
	if err != nil {
		logger.Error("检查端口冲突失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CHECK_PORT_CONFLICT")
		return
	}

	response.SuccessJSON(ctx, gin.H{"conflict": conflict}, "CHECK_PORT_CONFLICT")
}

// GetNodesByServer 按服务器查询节点列表
func (c *ServerNodeController) GetNodesByServer(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelServerId不能为空", "GET_NODES_BY_SERVER")
		return
	}

	nodes, err := c.serverNodeDAO.GetNodesByServer(tunnelServerId)
	if err != nil {
		logger.Error("查询服务器节点列表失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "GET_NODES_BY_SERVER")
		return
	}

	response.SuccessJSON(ctx, nodes, "GET_NODES_BY_SERVER")
}

// GetProxyTypeOptions 获取代理类型选项
func (c *ServerNodeController) GetProxyTypeOptions(ctx *gin.Context) {
	options := c.serverNodeDAO.GetProxyTypeOptions()
	response.SuccessJSON(ctx, options, "GET_PROXY_TYPE_OPTIONS")
}

// EnableServerNode 启用节点
func (c *ServerNodeController) EnableServerNode(ctx *gin.Context) {
	serverNodeId := request.GetParam(ctx, "serverNodeId")
	if serverNodeId == "" {
		response.ErrorJSON(ctx, "参数格式错误: serverNodeId不能为空", "ENABLE_SERVER_NODE")
		return
	}

	err := c.serverNodeDAO.EnableServerNode(serverNodeId)
	if err != nil {
		logger.Error("启用节点失败", "serverNodeId", serverNodeId, "error", err)
		response.ErrorJSON(ctx, "启用失败: "+err.Error(), "ENABLE_SERVER_NODE")
		return
	}

	logger.Info("启用服务器节点成功", "serverNodeId", serverNodeId)
	response.SuccessJSON(ctx, gin.H{"serverNodeId": serverNodeId, "message": "启用成功"}, "ENABLE_SERVER_NODE")
}

// DisableServerNode 禁用节点
func (c *ServerNodeController) DisableServerNode(ctx *gin.Context) {
	serverNodeId := request.GetParam(ctx, "serverNodeId")
	if serverNodeId == "" {
		response.ErrorJSON(ctx, "参数格式错误: serverNodeId不能为空", "DISABLE_SERVER_NODE")
		return
	}

	err := c.serverNodeDAO.DisableServerNode(serverNodeId)
	if err != nil {
		logger.Error("禁用节点失败", "serverNodeId", serverNodeId, "error", err)
		response.ErrorJSON(ctx, "禁用失败: "+err.Error(), "DISABLE_SERVER_NODE")
		return
	}

	logger.Info("禁用服务器节点成功", "serverNodeId", serverNodeId)
	response.SuccessJSON(ctx, gin.H{"serverNodeId": serverNodeId, "message": "禁用成功"}, "DISABLE_SERVER_NODE")
}

// BatchCreateNodes 批量创建节点
func (c *ServerNodeController) BatchCreateNodes(ctx *gin.Context) {
	type Request struct {
		Nodes []models.TunnelServerNode `json:"nodes" binding:"required"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "BATCH_CREATE_NODES")
		return
	}

	if len(req.Nodes) == 0 {
		response.ErrorJSON(ctx, "节点列表不能为空", "BATCH_CREATE_NODES")
		return
	}

	// 获取租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", "BATCH_CREATE_NODES")
		return
	}

	currentUser := c.getCurrentUser(ctx)
	var createdNodes []*models.TunnelServerNode
	var errors []map[string]interface{}

	for i, node := range req.Nodes {
		// 设置租户ID
		node.TenantId = tenantId

		// 生成ID和审计字段
		node.ServerNodeId = random.Generate32BitRandomString()
		node.AddWho = currentUser
		node.EditWho = currentUser
		node.OprSeqFlag = random.Generate32BitRandomString()

		createdNode, err := c.serverNodeDAO.CreateServerNode(&node)
		if err != nil {
			errors = append(errors, map[string]interface{}{
				"index":   i + 1,
				"message": err.Error(),
			})
			continue
		}
		createdNodes = append(createdNodes, createdNode)
	}

	result := gin.H{
		"successCount": len(createdNodes),
		"failCount":    len(errors),
		"nodes":        createdNodes,
		"errors":       errors,
	}

	if len(errors) > 0 {
		logger.Warn("批量创建节点部分失败", "successCount", len(createdNodes), "failCount", len(errors))
	} else {
		logger.Info("批量创建节点成功", "count", len(createdNodes))
	}

	response.SuccessJSON(ctx, result, "BATCH_CREATE_NODES")
}
