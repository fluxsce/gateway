package controllers

import (
	"strings"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0062/dao"
	"gateway/web/views/hub0062/models"

	"github.com/gin-gonic/gin"
)

type TunnelClientController struct {
	clientDAO *dao.TunnelClientDAO
}

func NewTunnelClientController(db database.Database) *TunnelClientController {
	return &TunnelClientController{
		clientDAO: dao.NewTunnelClientDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelClientController) getCurrentUser(ctx *gin.Context) string {
	if userName := request.GetUserName(ctx); userName != "" {
		return userName
	}
	if userID := request.GetUserID(ctx); userID != "" {
		return userID
	}
	return "admin"
}

// QueryTunnelClients 查询客户端列表
func (c *TunnelClientController) QueryTunnelClients(ctx *gin.Context) {
	var req models.TunnelClientQueryRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_TUNNEL_CLIENTS")
		return
	}

	// 参数验证和默认值
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageIndex <= 0 {
		req.PageIndex = 1
	}

	clients, total, err := c.clientDAO.QueryTunnelClients(ctx, &req)
	if err != nil {
		response.ErrorJSON(ctx, "查询客户端列表失败: "+err.Error(), "QUERY_TUNNEL_CLIENTS")
		return
	}

	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	response.PageJSON(ctx, clients, pageInfo, "QUERY_TUNNEL_CLIENTS")
}

// GetTunnelClient 获取客户端详情
func (c *TunnelClientController) GetTunnelClient(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "GET_TUNNEL_CLIENT")
		return
	}

	client, err := c.clientDAO.GetTunnelClient(ctx, tunnelClientId)
	if err != nil {
		response.ErrorJSON(ctx, "获取客户端详情失败: "+err.Error(), "GET_TUNNEL_CLIENT")
		return
	}

	response.SuccessJSON(ctx, client, "GET_TUNNEL_CLIENT")
}

// CreateTunnelClient 创建客户端
func (c *TunnelClientController) CreateTunnelClient(ctx *gin.Context) {
	var client models.TunnelClient
	if err := request.Bind(ctx, &client); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_CLIENT")
		return
	}

	// 参数验证
	if strings.TrimSpace(client.ClientName) == "" {
		response.ErrorJSON(ctx, "客户端名称不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ServerAddress) == "" {
		response.ErrorJSON(ctx, "服务器地址不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	if client.ServerPort <= 0 || client.ServerPort > 65535 {
		response.ErrorJSON(ctx, "服务器端口必须在1-65535之间", "CREATE_TUNNEL_CLIENT")
		return
	}

	// 设置租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	client.TenantId = tenantId

	// 设置创建人
	currentUser := c.getCurrentUser(ctx)
	client.AddWho = currentUser
	client.EditWho = currentUser

	// 创建客户端
	createdClient, err := c.clientDAO.CreateTunnelClient(ctx, &client)
	if err != nil {
		response.ErrorJSON(ctx, "创建客户端失败: "+err.Error(), "CREATE_TUNNEL_CLIENT")
		return
	}

	logger.Info("创建客户端成功", "tunnelClientId", createdClient.TunnelClientId, "clientName", createdClient.ClientName, "user", currentUser)
	response.SuccessJSON(ctx, createdClient, "CREATE_TUNNEL_CLIENT")
}

// UpdateTunnelClient 更新客户端
func (c *TunnelClientController) UpdateTunnelClient(ctx *gin.Context) {
	var client models.TunnelClient
	if err := request.Bind(ctx, &client); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}

	// 参数验证
	if strings.TrimSpace(client.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ClientName) == "" {
		response.ErrorJSON(ctx, "客户端名称不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ServerAddress) == "" {
		response.ErrorJSON(ctx, "服务器地址不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if client.ServerPort <= 0 || client.ServerPort > 65535 {
		response.ErrorJSON(ctx, "服务器端口必须在1-65535之间", "UPDATE_TUNNEL_CLIENT")
		return
	}

	// 设置修改人
	currentUser := c.getCurrentUser(ctx)
	client.EditWho = currentUser

	// 更新客户端
	updatedClient, err := c.clientDAO.UpdateTunnelClient(ctx, &client)
	if err != nil {
		response.ErrorJSON(ctx, "更新客户端失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}

	logger.Info("更新客户端成功", "tunnelClientId", updatedClient.TunnelClientId, "clientName", updatedClient.ClientName, "user", currentUser)
	response.SuccessJSON(ctx, updatedClient, "UPDATE_TUNNEL_CLIENT")
}

// DeleteTunnelClient 删除客户端
func (c *TunnelClientController) DeleteTunnelClient(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "DELETE_TUNNEL_CLIENT")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 删除客户端
	deletedClient, err := c.clientDAO.DeleteTunnelClient(ctx, tunnelClientId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "删除客户端失败: "+err.Error(), "DELETE_TUNNEL_CLIENT")
		return
	}

	logger.Info("删除客户端成功", "tunnelClientId", tunnelClientId, "user", currentUser)
	response.SuccessJSON(ctx, deletedClient, "DELETE_TUNNEL_CLIENT")
}

// GetClientStats 获取客户端统计信息
func (c *TunnelClientController) GetClientStats(ctx *gin.Context) {
	stats, err := c.clientDAO.GetClientStats(ctx)
	if err != nil {
		response.ErrorJSON(ctx, "获取客户端统计信息失败: "+err.Error(), "GET_CLIENT_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_CLIENT_STATS")
}

// StartClient 启动客户端（连接到服务器）
func (c *TunnelClientController) StartClient(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "START_CLIENT")
		return
	}

	// 调用DAO层启动客户端
	err := c.clientDAO.StartClient(ctx, tunnelClientId)
	if err != nil {
		logger.Error("启动客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		response.ErrorJSON(ctx, "启动失败: "+err.Error(), "START_CLIENT")
		return
	}

	logger.Info("启动客户端成功", "tunnelClientId", tunnelClientId)
	response.SuccessJSON(ctx, gin.H{
		"tunnelClientId": tunnelClientId,
		"message":        "客户端启动成功",
	}, "START_CLIENT")
}

// StopClient 停止客户端（断开连接）
func (c *TunnelClientController) StopClient(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "STOP_CLIENT")
		return
	}

	// 调用DAO层停止客户端
	err := c.clientDAO.StopClient(ctx, tunnelClientId)
	if err != nil {
		logger.Error("停止客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		response.ErrorJSON(ctx, "停止失败: "+err.Error(), "STOP_CLIENT")
		return
	}

	logger.Info("停止客户端成功", "tunnelClientId", tunnelClientId)
	response.SuccessJSON(ctx, gin.H{
		"tunnelClientId": tunnelClientId,
		"message":        "客户端停止成功",
	}, "STOP_CLIENT")
}

// RestartClient 重启客户端（重新连接）
func (c *TunnelClientController) RestartClient(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "RESTART_CLIENT")
		return
	}

	// 调用DAO层重启客户端
	err := c.clientDAO.RestartClient(ctx, tunnelClientId)
	if err != nil {
		logger.Error("重启客户端失败", "tunnelClientId", tunnelClientId, "error", err)
		response.ErrorJSON(ctx, "重启失败: "+err.Error(), "RESTART_CLIENT")
		return
	}

	logger.Info("重启客户端成功", "tunnelClientId", tunnelClientId)
	response.SuccessJSON(ctx, gin.H{
		"tunnelClientId": tunnelClientId,
		"message":        "客户端重启成功",
	}, "RESTART_CLIENT")
}

// BatchEnableClients 批量启用客户端
func (c *TunnelClientController) BatchEnableClients(ctx *gin.Context) {
	var req models.BatchOperationRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "BATCH_ENABLE_CLIENTS")
		return
	}

	if len(req.ClientIds) == 0 {
		response.ErrorJSON(ctx, "客户端ID列表不能为空", "BATCH_ENABLE_CLIENTS")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 批量启用客户端
	result, err := c.clientDAO.BatchEnableClients(ctx, req.ClientIds, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "批量启用客户端失败: "+err.Error(), "BATCH_ENABLE_CLIENTS")
		return
	}

	logger.Info("批量启用客户端完成", "successCount", result.SuccessCount, "failedCount", result.FailedCount, "user", currentUser)
	response.SuccessJSON(ctx, result, "BATCH_ENABLE_CLIENTS")
}

// BatchDisableClients 批量禁用客户端
func (c *TunnelClientController) BatchDisableClients(ctx *gin.Context) {
	var req models.BatchOperationRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "BATCH_DISABLE_CLIENTS")
		return
	}

	if len(req.ClientIds) == 0 {
		response.ErrorJSON(ctx, "客户端ID列表不能为空", "BATCH_DISABLE_CLIENTS")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 批量禁用客户端
	result, err := c.clientDAO.BatchDisableClients(ctx, req.ClientIds, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "批量禁用客户端失败: "+err.Error(), "BATCH_DISABLE_CLIENTS")
		return
	}

	logger.Info("批量禁用客户端完成", "successCount", result.SuccessCount, "failedCount", result.FailedCount, "user", currentUser)
	response.SuccessJSON(ctx, result, "BATCH_DISABLE_CLIENTS")
}

// GetClientServices 获取客户端注册的服务列表
func (c *TunnelClientController) GetClientServices(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "GET_CLIENT_SERVICES")
		return
	}

	// TODO: 实现从 hub0063 模块查询服务列表
	// 这里暂时返回空列表，等 hub0063 模块开发完成后再实现
	response.SuccessJSON(ctx, []interface{}{}, "GET_CLIENT_SERVICES")
}

// GetClientSessions 获取客户端会话列表
func (c *TunnelClientController) GetClientSessions(ctx *gin.Context) {
	tunnelClientId := request.GetParam(ctx, "tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "参数格式错误: tunnelClientId不能为空", "GET_CLIENT_SESSIONS")
		return
	}

	// TODO: 实现从 hub0064 模块查询会话列表
	// 这里暂时返回空列表，等 hub0064 模块开发完成后再实现
	response.SuccessJSON(ctx, []interface{}{}, "GET_CLIENT_SESSIONS")
}
