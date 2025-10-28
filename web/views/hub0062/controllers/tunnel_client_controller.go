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

	clients, total, err := c.clientDAO.QueryTunnelClients(&req)
	if err != nil {
		response.ErrorJSON(ctx, "查询客户端列表失败: "+err.Error(), "QUERY_TUNNEL_CLIENTS")
		return
	}

	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)
	response.PageJSON(ctx, clients, pageInfo, "QUERY_TUNNEL_CLIENTS")
}

// GetTunnelClient 获取客户端详情
func (c *TunnelClientController) GetTunnelClient(ctx *gin.Context) {
	tunnelClientId := ctx.PostForm("tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "GET_TUNNEL_CLIENT")
		return
	}

	client, err := c.clientDAO.GetTunnelClient(tunnelClientId)
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

	// 设置创建人
	currentUser := c.getCurrentUser(ctx)
	client.AddWho = currentUser
	client.EditWho = currentUser

	// 创建客户端
	createdClient, err := c.clientDAO.CreateTunnelClient(&client)
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
	updatedClient, err := c.clientDAO.UpdateTunnelClient(&client)
	if err != nil {
		response.ErrorJSON(ctx, "更新客户端失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}

	logger.Info("更新客户端成功", "tunnelClientId", updatedClient.TunnelClientId, "clientName", updatedClient.ClientName, "user", currentUser)
	response.SuccessJSON(ctx, updatedClient, "UPDATE_TUNNEL_CLIENT")
}

// DeleteTunnelClient 删除客户端
func (c *TunnelClientController) DeleteTunnelClient(ctx *gin.Context) {
	tunnelClientId := ctx.PostForm("tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "DELETE_TUNNEL_CLIENT")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 删除客户端
	deletedClient, err := c.clientDAO.DeleteTunnelClient(tunnelClientId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "删除客户端失败: "+err.Error(), "DELETE_TUNNEL_CLIENT")
		return
	}

	logger.Info("删除客户端成功", "tunnelClientId", tunnelClientId, "user", currentUser)
	response.SuccessJSON(ctx, deletedClient, "DELETE_TUNNEL_CLIENT")
}

// GetClientStats 获取客户端统计信息
func (c *TunnelClientController) GetClientStats(ctx *gin.Context) {
	stats, err := c.clientDAO.GetClientStats()
	if err != nil {
		response.ErrorJSON(ctx, "获取客户端统计信息失败: "+err.Error(), "GET_CLIENT_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_CLIENT_STATS")
}

// GetClientStatus 获取客户端实时状态
func (c *TunnelClientController) GetClientStatus(ctx *gin.Context) {
	tunnelClientId := ctx.PostForm("tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "GET_CLIENT_STATUS")
		return
	}

	status, err := c.clientDAO.GetClientStatus(tunnelClientId)
	if err != nil {
		response.ErrorJSON(ctx, "获取客户端状态失败: "+err.Error(), "GET_CLIENT_STATUS")
		return
	}

	response.SuccessJSON(ctx, status, "GET_CLIENT_STATUS")
}

// ResetAuthToken 重置客户端认证令牌
func (c *TunnelClientController) ResetAuthToken(ctx *gin.Context) {
	var req models.ResetAuthTokenRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "RESET_AUTH_TOKEN")
		return
	}

	if strings.TrimSpace(req.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "RESET_AUTH_TOKEN")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 重置认证令牌
	result, err := c.clientDAO.ResetAuthToken(req.TunnelClientId, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "重置认证令牌失败: "+err.Error(), "RESET_AUTH_TOKEN")
		return
	}

	logger.Info("重置认证令牌成功", "tunnelClientId", req.TunnelClientId, "user", currentUser)
	response.SuccessJSON(ctx, result, "RESET_AUTH_TOKEN")
}

// DisconnectClient 强制断开客户端连接
func (c *TunnelClientController) DisconnectClient(ctx *gin.Context) {
	var req models.DisconnectClientRequest
	if err := request.Bind(ctx, &req); err != nil {
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "DISCONNECT_CLIENT")
		return
	}

	if strings.TrimSpace(req.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "DISCONNECT_CLIENT")
		return
	}

	currentUser := c.getCurrentUser(ctx)

	// 断开客户端连接
	err := c.clientDAO.DisconnectClient(req.TunnelClientId, req.Reason, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "断开客户端连接失败: "+err.Error(), "DISCONNECT_CLIENT")
		return
	}

	logger.Info("断开客户端连接成功", "tunnelClientId", req.TunnelClientId, "reason", req.Reason, "user", currentUser)
	response.SuccessJSON(ctx, gin.H{
		"tunnelClientId": req.TunnelClientId,
		"message":        "客户端连接已断开",
	}, "DISCONNECT_CLIENT")
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
	result, err := c.clientDAO.BatchEnableClients(req.ClientIds, currentUser)
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
	result, err := c.clientDAO.BatchDisableClients(req.ClientIds, currentUser)
	if err != nil {
		response.ErrorJSON(ctx, "批量禁用客户端失败: "+err.Error(), "BATCH_DISABLE_CLIENTS")
		return
	}

	logger.Info("批量禁用客户端完成", "successCount", result.SuccessCount, "failedCount", result.FailedCount, "user", currentUser)
	response.SuccessJSON(ctx, result, "BATCH_DISABLE_CLIENTS")
}

// GetConnectionStatusOptions 获取连接状态选项
func (c *TunnelClientController) GetConnectionStatusOptions(ctx *gin.Context) {
	options := []gin.H{
		{"value": "connected", "label": "已连接"},
		{"value": "disconnected", "label": "已断开"},
		{"value": "connecting", "label": "连接中"},
		{"value": "error", "label": "错误"},
	}

	response.SuccessJSON(ctx, options, "GET_CONNECTION_STATUS_OPTIONS")
}

// GetClientServices 获取客户端注册的服务列表
func (c *TunnelClientController) GetClientServices(ctx *gin.Context) {
	tunnelClientId := ctx.PostForm("tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "GET_CLIENT_SERVICES")
		return
	}

	// TODO: 实现从 hub0063 模块查询服务列表
	// 这里暂时返回空列表，等 hub0063 模块开发完成后再实现
	response.SuccessJSON(ctx, []interface{}{}, "GET_CLIENT_SERVICES")
}

// GetClientSessions 获取客户端会话列表
func (c *TunnelClientController) GetClientSessions(ctx *gin.Context) {
	tunnelClientId := ctx.PostForm("tunnelClientId")
	if tunnelClientId == "" {
		response.ErrorJSON(ctx, "客户端ID不能为空", "GET_CLIENT_SESSIONS")
		return
	}

	// TODO: 实现从 hub0064 模块查询会话列表
	// 这里暂时返回空列表，等 hub0064 模块开发完成后再实现
	response.SuccessJSON(ctx, []interface{}{}, "GET_CLIENT_SESSIONS")
}
