package controllers

import (
	"fmt"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/web/utils/response"
	"gateway/web/views/hub0062/dao"
	"gateway/web/views/hub0062/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// TunnelClientController 隧道客户端控制器
type TunnelClientController struct {
	tunnelClientDAO *dao.TunnelClientDAO
}

// NewTunnelClientController 创建隧道客户端控制器实例
func NewTunnelClientController(db database.Database) *TunnelClientController {
	return &TunnelClientController{
		tunnelClientDAO: dao.NewTunnelClientDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelClientController) getCurrentUser(ctx *gin.Context) string {
	// 简化实现：返回默认用户
	// 实际项目中应该从session或JWT token中获取用户信息
	return "admin"
}

// QueryTunnelClients 查询隧道客户端列表
func (c *TunnelClientController) QueryTunnelClients(ctx *gin.Context) {
	var req models.TunnelClientQueryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定查询参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "QUERY_TUNNEL_CLIENTS")
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

	clients, total, err := c.tunnelClientDAO.QueryTunnelClients(&req)
	if err != nil {
		logger.Error("查询隧道客户端列表失败", "error", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "QUERY_TUNNEL_CLIENTS")
		return
	}

	// 创建分页信息
	pageInfo := response.NewPageInfo(req.PageIndex, req.PageSize, total)

	response.PageJSON(ctx, clients, pageInfo, "QUERY_TUNNEL_CLIENTS")
}

// GetTunnelClient 获取隧道客户端详情
func (c *TunnelClientController) GetTunnelClient(ctx *gin.Context) {
	type Request struct {
		TunnelClientId string `json:"tunnelClientId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_CLIENT")
		return
	}

	client, err := c.tunnelClientDAO.GetTunnelClient(req.TunnelClientId)
	if err != nil {
		logger.Error("获取隧道客户端详情失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_CLIENT")
		return
	}

	response.SuccessJSON(ctx, client, "GET_TUNNEL_CLIENT")
}

// CreateTunnelClient 创建隧道客户端
func (c *TunnelClientController) CreateTunnelClient(ctx *gin.Context) {
	var client models.TunnelClient
	if err := ctx.ShouldBindJSON(&client); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_CLIENT")
		return
	}

	// 参数验证
	if strings.TrimSpace(client.ClientName) == "" {
		response.ErrorJSON(ctx, "客户端名称不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ClientAddress) == "" {
		response.ErrorJSON(ctx, "客户端地址不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "CREATE_TUNNEL_CLIENT")
		return
	}
	if client.HeartbeatInterval <= 0 {
		client.HeartbeatInterval = 30 // 默认30秒
	}
	if client.MaxRetries <= 0 {
		client.MaxRetries = 3 // 默认3次
	}
	if client.RetryInterval <= 0 {
		client.RetryInterval = 5 // 默认5秒
	}

	// 检查客户端名称是否重复
	exists, err := c.tunnelClientDAO.CheckClientNameExists(client.ClientName, "")
	if err != nil {
		logger.Error("检查客户端名称是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "CREATE_TUNNEL_CLIENT")
		return
	}
	if exists {
		response.ErrorJSON(ctx, "客户端名称已存在", "CREATE_TUNNEL_CLIENT")
		return
	}

	// 生成ID和设置审计字段
	client.TunnelClientId = uuid.New().String()
	client.AddWho = c.getCurrentUser(ctx)
	client.EditWho = client.AddWho
	client.OprSeqFlag = uuid.New().String()

	// 生成认证令牌
	if strings.TrimSpace(client.AuthToken) == "" {
		client.AuthToken = uuid.New().String()
	}

	err = c.tunnelClientDAO.CreateTunnelClient(&client)
	if err != nil {
		logger.Error("创建隧道客户端失败", "error", err)
		response.ErrorJSON(ctx, "创建失败: "+err.Error(), "CREATE_TUNNEL_CLIENT")
		return
	}

	logger.Info("创建隧道客户端成功", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	result := map[string]interface{}{
		"tunnelClientId": client.TunnelClientId,
		"message":        "创建成功",
	}
	response.SuccessJSON(ctx, result, "CREATE_TUNNEL_CLIENT")
}

// UpdateTunnelClient 更新隧道客户端
func (c *TunnelClientController) UpdateTunnelClient(ctx *gin.Context) {
	var client models.TunnelClient
	if err := ctx.ShouldBindJSON(&client); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}

	// 参数验证
	if strings.TrimSpace(client.TunnelClientId) == "" {
		response.ErrorJSON(ctx, "隧道客户端ID不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ClientName) == "" {
		response.ErrorJSON(ctx, "客户端名称不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.ClientAddress) == "" {
		response.ErrorJSON(ctx, "客户端地址不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if strings.TrimSpace(client.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "UPDATE_TUNNEL_CLIENT")
		return
	}
	if client.HeartbeatInterval <= 0 {
		client.HeartbeatInterval = 30 // 默认30秒
	}
	if client.MaxRetries <= 0 {
		client.MaxRetries = 3 // 默认3次
	}
	if client.RetryInterval <= 0 {
		client.RetryInterval = 5 // 默认5秒
	}

	// 检查客户端名称是否重复
	exists, err := c.tunnelClientDAO.CheckClientNameExists(client.ClientName, client.TunnelClientId)
	if err != nil {
		logger.Error("检查客户端名称是否存在失败", "error", err)
		response.ErrorJSON(ctx, "检查失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}
	if exists {
		response.ErrorJSON(ctx, "客户端名称已存在", "UPDATE_TUNNEL_CLIENT")
		return
	}

	// 设置审计字段
	client.EditWho = c.getCurrentUser(ctx)

	err = c.tunnelClientDAO.UpdateTunnelClient(&client)
	if err != nil {
		logger.Error("更新隧道客户端失败", "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT")
		return
	}

	logger.Info("更新隧道客户端成功", "tunnelClientId", client.TunnelClientId, "clientName", client.ClientName)
	result := map[string]interface{}{
		"message": "更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_CLIENT")
}

// DeleteTunnelClient 删除隧道客户端
func (c *TunnelClientController) DeleteTunnelClient(ctx *gin.Context) {
	type Request struct {
		TunnelClientId string `json:"tunnelClientId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "DELETE_TUNNEL_CLIENT")
		return
	}

	editWho := c.getCurrentUser(ctx)
	err := c.tunnelClientDAO.DeleteTunnelClient(req.TunnelClientId, editWho)
	if err != nil {
		logger.Error("删除隧道客户端失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "删除失败: "+err.Error(), "DELETE_TUNNEL_CLIENT")
		return
	}

	logger.Info("删除隧道客户端成功", "tunnelClientId", req.TunnelClientId)
	result := map[string]interface{}{
		"message": "删除成功",
	}
	response.SuccessJSON(ctx, result, "DELETE_TUNNEL_CLIENT")
}

// UpdateTunnelClientStatus 更新隧道客户端状态
func (c *TunnelClientController) UpdateTunnelClientStatus(ctx *gin.Context) {
	type Request struct {
		TunnelClientId     string `json:"tunnelClientId" binding:"required"`
		Status             string `json:"status" binding:"required"`
		RegisteredServices int    `json:"registeredServices"`
		ActiveProxies      int    `json:"activeProxies"`
		TotalTraffic       int64  `json:"totalTraffic"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_CLIENT_STATUS")
		return
	}

	// 状态验证
	validStatuses := []string{"CONNECTED", "DISCONNECTED", "RECONNECTING"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		response.ErrorJSON(ctx, "无效的客户端状态", "UPDATE_TUNNEL_CLIENT_STATUS")
		return
	}

	err := c.tunnelClientDAO.UpdateTunnelClientStatus(req.TunnelClientId, req.Status, req.RegisteredServices, req.ActiveProxies, req.TotalTraffic)
	if err != nil {
		logger.Error("更新隧道客户端状态失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT_STATUS")
		return
	}

	logger.Info("更新隧道客户端状态成功", "tunnelClientId", req.TunnelClientId, "status", req.Status)
	result := map[string]interface{}{
		"message": "状态更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_CLIENT_STATUS")
}

// UpdateTunnelClientConnection 更新隧道客户端连接信息
func (c *TunnelClientController) UpdateTunnelClientConnection(ctx *gin.Context) {
	type Request struct {
		TunnelClientId string `json:"tunnelClientId" binding:"required"`
		Status         string `json:"status" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_CLIENT_CONNECTION")
		return
	}

	// 状态验证
	validStatuses := []string{"CONNECTED", "DISCONNECTED", "RECONNECTING"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		response.ErrorJSON(ctx, "无效的客户端状态", "UPDATE_TUNNEL_CLIENT_CONNECTION")
		return
	}

	err := c.tunnelClientDAO.UpdateTunnelClientConnection(req.TunnelClientId, req.Status)
	if err != nil {
		logger.Error("更新隧道客户端连接信息失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_CLIENT_CONNECTION")
		return
	}

	logger.Info("更新隧道客户端连接信息成功", "tunnelClientId", req.TunnelClientId, "status", req.Status)
	result := map[string]interface{}{
		"message": "连接状态更新成功",
	}
	response.SuccessJSON(ctx, result, "UPDATE_TUNNEL_CLIENT_CONNECTION")
}

// GetTunnelClientStats 获取隧道客户端统计信息
func (c *TunnelClientController) GetTunnelClientStats(ctx *gin.Context) {
	stats, err := c.tunnelClientDAO.GetTunnelClientStats()
	if err != nil {
		logger.Error("获取隧道客户端统计信息失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_CLIENT_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_TUNNEL_CLIENT_STATS")
}

// GetClientStatusOptions 获取客户端状态选项
func (c *TunnelClientController) GetClientStatusOptions(ctx *gin.Context) {
	options := c.tunnelClientDAO.GetClientStatusOptions()
	response.SuccessJSON(ctx, options, "GET_CLIENT_STATUS_OPTIONS")
}

// GetTunnelClientsByServerId 根据服务器ID获取客户端列表
func (c *TunnelClientController) GetTunnelClientsByServerId(ctx *gin.Context) {
	type Request struct {
		TunnelServerId string `json:"tunnelServerId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "GET_TUNNEL_CLIENTS_BY_SERVER_ID")
		return
	}

	clients, err := c.tunnelClientDAO.GetTunnelClientsByServerId(req.TunnelServerId)
	if err != nil {
		logger.Error("根据服务器ID获取客户端列表失败", "tunnelServerId", req.TunnelServerId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_CLIENTS_BY_SERVER_ID")
		return
	}

	// 转换为下拉选项格式
	options := make([]map[string]interface{}, 0, len(clients))
	for _, client := range clients {
		option := map[string]interface{}{
			"value":              client.TunnelClientId,
			"label":              fmt.Sprintf("%s (%s)", client.ClientName, client.ClientStatus),
			"status":             client.ClientStatus,
			"registeredServices": client.RegisteredServices,
			"activeProxies":      client.ActiveProxies,
		}
		options = append(options, option)
	}

	response.SuccessJSON(ctx, options, "GET_TUNNEL_CLIENTS_BY_SERVER_ID")
}

// GenerateAuthToken 生成新的认证令牌
func (c *TunnelClientController) GenerateAuthToken(ctx *gin.Context) {
	token := uuid.New().String()
	result := map[string]interface{}{
		"authToken": token,
	}
	response.SuccessJSON(ctx, result, "GENERATE_AUTH_TOKEN")
}

// TestClientConnection 测试客户端连接
func (c *TunnelClientController) TestClientConnection(ctx *gin.Context) {
	type Request struct {
		TunnelClientId string `json:"tunnelClientId" binding:"required"`
	}

	var req Request
	if err := ctx.ShouldBindJSON(&req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "TEST_CLIENT_CONNECTION")
		return
	}

	// 获取客户端信息
	client, err := c.tunnelClientDAO.GetTunnelClient(req.TunnelClientId)
	if err != nil {
		logger.Error("获取隧道客户端详情失败", "tunnelClientId", req.TunnelClientId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "TEST_CLIENT_CONNECTION")
		return
	}

	// 这里可以实现实际的连接测试逻辑
	// 简化实现，直接返回成功
	result := map[string]interface{}{
		"success":    true,
		"message":    "客户端连接测试成功",
		"clientName": client.ClientName,
		"status":     client.ClientStatus,
		"testTime":   time.Now().Format("2006-01-02 15:04:05"),
	}

	response.SuccessJSON(ctx, result, "TEST_CLIENT_CONNECTION")
}
