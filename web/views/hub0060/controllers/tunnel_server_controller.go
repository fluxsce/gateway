package controllers

import (
	"encoding/json"
	"strings"

	"gateway/internal/tunnel/server"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"gateway/web/utils/request"
	"gateway/web/utils/response"
	"gateway/web/views/hub0060/dao"
	"gateway/web/views/hub0060/models"

	"github.com/gin-gonic/gin"
)

// TunnelServerController 隧道服务器控制器
type TunnelServerController struct {
	tunnelServerDAO *dao.TunnelServerDAO
}

// NewTunnelServerController 创建隧道服务器控制器实例
func NewTunnelServerController(db database.Database) *TunnelServerController {
	return &TunnelServerController{
		tunnelServerDAO: dao.NewTunnelServerDAO(db),
	}
}

// getCurrentUser 获取当前用户
func (c *TunnelServerController) getCurrentUser(ctx *gin.Context) string {
	// 使用 request 工具类获取用户信息
	if userName := request.GetUserName(ctx); userName != "" {
		return userName
	}
	if userID := request.GetUserID(ctx); userID != "" {
		return userID
	}
	// 如果无法获取用户信息，返回默认用户
	return "admin"
}

// QueryTunnelServers 查询隧道服务器列表
func (c *TunnelServerController) QueryTunnelServers(ctx *gin.Context) {
	// 使用工具类获取分页参数
	page, pageSize := request.GetPaginationParams(ctx)

	// 绑定查询条件（支持 Query / JSON Body / Form 等多种来源）
	var query models.TunnelServerQueryRequest
	if err := request.BindSafely(ctx, &query); err != nil {
		logger.WarnWithTrace(ctx, "绑定隧道服务器查询条件失败，使用默认条件", "error", err.Error())
	}

	// 设置分页参数到查询对象
	query.PageIndex = page
	query.PageSize = pageSize

	// 注意：租户ID由DAO层从上下文或session中自动获取，Controller层无需处理

	servers, total, err := c.tunnelServerDAO.QueryTunnelServers(ctx.Request.Context(), &query)
	if err != nil {
		logger.ErrorWithTrace(ctx, "查询隧道服务器列表失败", err)
		response.ErrorJSON(ctx, "查询失败: "+err.Error(), "QUERY_TUNNEL_SERVERS")
		return
	}

	// 创建分页信息
	pageInfo := response.NewPageInfo(page, pageSize, total)
	pageInfo.MainKey = "tunnelServerId"

	// 使用统一的分页响应
	response.PageJSON(ctx, servers, pageInfo, "QUERY_TUNNEL_SERVERS")
}

// GetTunnelServer 获取隧道服务器详情
func (c *TunnelServerController) GetTunnelServer(ctx *gin.Context) {
	// 直接从请求参数获取
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "GET_TUNNEL_SERVER")
		return
	}

	server, err := c.tunnelServerDAO.GetTunnelServer(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("获取隧道服务器详情失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVER")
		return
	}

	response.SuccessJSON(ctx, server, "GET_TUNNEL_SERVER")
}

// CreateTunnelServer 创建隧道服务器
func (c *TunnelServerController) CreateTunnelServer(ctx *gin.Context) {
	var server types.TunnelServer
	if err := request.Bind(ctx, &server); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "CREATE_TUNNEL_SERVER")
		return
	}

	// 参数验证
	if strings.TrimSpace(server.ServerName) == "" {
		response.ErrorJSON(ctx, "服务器名称不能为空", "CREATE_TUNNEL_SERVER")
		return
	}
	if strings.TrimSpace(server.ControlAddress) == "" {
		response.ErrorJSON(ctx, "控制端口地址不能为空", "CREATE_TUNNEL_SERVER")
		return
	}
	if server.ControlPort <= 0 || server.ControlPort > 65535 {
		response.ErrorJSON(ctx, "控制端口必须在1-65535之间", "CREATE_TUNNEL_SERVER")
		return
	}
	if server.MaxClients <= 0 {
		response.ErrorJSON(ctx, "最大客户端数量必须大于0", "CREATE_TUNNEL_SERVER")
		return
	}
	if server.MaxPortsPerClient != nil && *server.MaxPortsPerClient <= 0 {
		response.ErrorJSON(ctx, "每个客户端最大端口数必须大于0", "CREATE_TUNNEL_SERVER")
		return
	}
	if server.HeartbeatTimeout <= 0 {
		server.HeartbeatTimeout = 60 // 默认60秒
	}

	// 服务器名称重复性检查已经移到DAO层实现

	// 设置租户ID
	tenantId := request.GetTenantID(ctx)
	if tenantId == "" {
		response.ErrorJSON(ctx, "租户ID不能为空", "CREATE_TUNNEL_SERVER")
		return
	}
	server.TenantId = tenantId

	// 生成ID和设置审计字段
	server.TunnelServerId = random.Generate32BitRandomString()
	server.AddWho = c.getCurrentUser(ctx)
	server.EditWho = server.AddWho
	server.OprSeqFlag = random.Generate32BitRandomString()

	// 生成认证令牌
	if strings.TrimSpace(server.AuthToken) == "" {
		server.AuthToken = random.Generate32BitRandomString()
	}

	createdServer, err := c.tunnelServerDAO.CreateTunnelServer(ctx.Request.Context(), &server)
	if err != nil {
		logger.Error("创建隧道服务器失败", "error", err)
		response.ErrorJSON(ctx, "创建失败: "+err.Error(), "CREATE_TUNNEL_SERVER")
		return
	}

	logger.Info("创建隧道服务器成功", "tunnelServerId", createdServer.TunnelServerId, "serverName", createdServer.ServerName)
	response.SuccessJSON(ctx, createdServer, "CREATE_TUNNEL_SERVER")
}

// UpdateTunnelServer 更新隧道服务器
func (c *TunnelServerController) UpdateTunnelServer(ctx *gin.Context) {
	var server types.TunnelServer
	if err := request.Bind(ctx, &server); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVER")
		return
	}

	// 参数验证
	if strings.TrimSpace(server.TunnelServerId) == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "UPDATE_TUNNEL_SERVER")
		return
	}
	if strings.TrimSpace(server.ServerName) == "" {
		response.ErrorJSON(ctx, "服务器名称不能为空", "UPDATE_TUNNEL_SERVER")
		return
	}
	if strings.TrimSpace(server.ControlAddress) == "" {
		response.ErrorJSON(ctx, "控制端口地址不能为空", "UPDATE_TUNNEL_SERVER")
		return
	}
	if server.ControlPort <= 0 || server.ControlPort > 65535 {
		response.ErrorJSON(ctx, "控制端口必须在1-65535之间", "UPDATE_TUNNEL_SERVER")
		return
	}
	if server.MaxClients <= 0 {
		response.ErrorJSON(ctx, "最大客户端数量必须大于0", "UPDATE_TUNNEL_SERVER")
		return
	}
	if server.MaxPortsPerClient != nil && *server.MaxPortsPerClient <= 0 {
		response.ErrorJSON(ctx, "每个客户端最大端口数必须大于0", "UPDATE_TUNNEL_SERVER")
		return
	}
	if server.HeartbeatTimeout <= 0 {
		server.HeartbeatTimeout = 60 // 默认60秒
	}

	// 服务器名称重复性检查已经移到DAO层实现

	// 设置审计字段
	server.EditWho = c.getCurrentUser(ctx)

	updatedServer, err := c.tunnelServerDAO.UpdateTunnelServer(ctx.Request.Context(), &server)
	if err != nil {
		logger.Error("更新隧道服务器失败", "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_SERVER")
		return
	}

	logger.Info("更新隧道服务器成功", "tunnelServerId", updatedServer.TunnelServerId, "serverName", updatedServer.ServerName)
	response.SuccessJSON(ctx, updatedServer, "UPDATE_TUNNEL_SERVER")
}

// DeleteTunnelServer 删除隧道服务器
func (c *TunnelServerController) DeleteTunnelServer(ctx *gin.Context) {
	// 直接从请求参数获取
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "隧道服务器ID不能为空", "DELETE_TUNNEL_SERVER")
		return
	}

	// 先检查服务器状态
	server, err := c.tunnelServerDAO.GetTunnelServer(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("获取隧道服务器信息失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "获取服务器信息失败: "+err.Error(), "DELETE_TUNNEL_SERVER")
		return
	}

	// 如果服务器正在运行，先停止服务
	if server.ServerStatus == "running" {
		logger.Info("服务器正在运行，先停止服务", "tunnelServerId", tunnelServerId)
		if _, err := c.tunnelServerDAO.StopTunnelServer(ctx.Request.Context(), tunnelServerId); err != nil {
			logger.Error("停止隧道服务器失败", "tunnelServerId", tunnelServerId, "error", err)
			response.ErrorJSON(ctx, "删除前停止服务失败: "+err.Error(), "DELETE_TUNNEL_SERVER")
			return
		}
		logger.Info("服务器已停止", "tunnelServerId", tunnelServerId)
	}

	// 执行删除操作
	editWho := c.getCurrentUser(ctx)
	deletedServer, err := c.tunnelServerDAO.DeleteTunnelServer(ctx.Request.Context(), tunnelServerId, editWho)
	if err != nil {
		logger.Error("删除隧道服务器失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "删除失败: "+err.Error(), "DELETE_TUNNEL_SERVER")
		return
	}

	logger.Info("删除隧道服务器成功", "tunnelServerId", tunnelServerId)
	response.SuccessJSON(ctx, deletedServer, "DELETE_TUNNEL_SERVER")
}

// UpdateTunnelServerStatus 更新隧道服务器状态
func (c *TunnelServerController) UpdateTunnelServerStatus(ctx *gin.Context) {
	type Request struct {
		TunnelServerId string `json:"tunnelServerId" form:"tunnelServerId" query:"tunnelServerId" binding:"required"`
		Status         string `json:"status" form:"status" query:"status" binding:"required"`
	}

	var req Request
	if err := request.Bind(ctx, &req); err != nil {
		logger.Error("绑定参数失败", "error", err)
		response.ErrorJSON(ctx, "参数格式错误: "+err.Error(), "UPDATE_TUNNEL_SERVER_STATUS")
		return
	}

	// 状态验证
	validStatuses := []string{"running", "stopped", "error"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}
	if !isValid {
		response.ErrorJSON(ctx, "无效的服务器状态", "UPDATE_TUNNEL_SERVER_STATUS")
		return
	}

	updatedServer, err := c.tunnelServerDAO.UpdateTunnelServerStatus(ctx.Request.Context(), req.TunnelServerId, req.Status)
	if err != nil {
		logger.Error("更新隧道服务器状态失败", "tunnelServerId", req.TunnelServerId, "error", err)
		response.ErrorJSON(ctx, "更新失败: "+err.Error(), "UPDATE_TUNNEL_SERVER_STATUS")
		return
	}

	logger.Info("更新隧道服务器状态成功", "tunnelServerId", req.TunnelServerId, "status", req.Status)
	response.SuccessJSON(ctx, updatedServer, "UPDATE_TUNNEL_SERVER_STATUS")
}

// GetTunnelServerStats 获取隧道服务器统计信息
func (c *TunnelServerController) GetTunnelServerStats(ctx *gin.Context) {
	stats, err := c.tunnelServerDAO.GetTunnelServerStats(ctx.Request.Context())
	if err != nil {
		logger.Error("获取隧道服务器统计信息失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVER_STATS")
		return
	}

	response.SuccessJSON(ctx, stats, "GET_TUNNEL_SERVER_STATS")
}

// GetServerStatusOptions 获取服务器状态选项
func (c *TunnelServerController) GetServerStatusOptions(ctx *gin.Context) {
	options := c.tunnelServerDAO.GetServerStatusOptions()
	response.SuccessJSON(ctx, options, "GET_SERVER_STATUS_OPTIONS")
}

// GetTunnelServerList 获取隧道服务器列表（用于下拉选择）
func (c *TunnelServerController) GetTunnelServerList(ctx *gin.Context) {
	servers, err := c.tunnelServerDAO.GetTunnelServerList(ctx.Request.Context())
	if err != nil {
		logger.Error("获取隧道服务器列表失败", "error", err)
		response.ErrorJSON(ctx, "获取失败: "+err.Error(), "GET_TUNNEL_SERVER_LIST")
		return
	}

	response.SuccessJSON(ctx, servers, "GET_TUNNEL_SERVER_LIST")
}

// GenerateAuthToken 生成新的认证令牌
func (c *TunnelServerController) GenerateAuthToken(ctx *gin.Context) {
	token := random.Generate32BitRandomString()
	response.SuccessJSON(ctx, gin.H{"authToken": token}, "GENERATE_AUTH_TOKEN")
}

// StartTunnelServer 启动隧道服务器
func (c *TunnelServerController) StartTunnelServer(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "参数格式错误: "+tunnelServerId, "START_TUNNEL_SERVER")
		return
	}

	// 调用DAO层启动服务器
	server, err := c.tunnelServerDAO.StartTunnelServer(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("启动隧道服务器失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "启动失败: "+err.Error(), "START_TUNNEL_SERVER")
		return
	}

	logger.Info("启动隧道服务器成功", "tunnelServerId", tunnelServerId)
	response.SuccessJSON(ctx, server, "START_TUNNEL_SERVER")
}

// StopTunnelServer 停止隧道服务器
func (c *TunnelServerController) StopTunnelServer(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "参数格式错误: "+tunnelServerId, "START_TUNNEL_SERVER")
		return
	}

	// 调用DAO层停止服务器
	server, err := c.tunnelServerDAO.StopTunnelServer(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("停止隧道服务器失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "停止失败: "+err.Error(), "STOP_TUNNEL_SERVER")
		return
	}

	logger.Info("停止隧道服务器成功", "tunnelServerId", tunnelServerId)
	response.SuccessJSON(ctx, server, "STOP_TUNNEL_SERVER")
}

// RestartTunnelServer 重启隧道服务器
func (c *TunnelServerController) RestartTunnelServer(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "参数格式错误: "+tunnelServerId, "START_TUNNEL_SERVER")
		return
	}
	// 调用DAO层重启服务器
	server, err := c.tunnelServerDAO.RestartTunnelServer(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("重启隧道服务器失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "重启失败: "+err.Error(), "RESTART_TUNNEL_SERVER")
		return
	}

	logger.Info("重启隧道服务器成功", "tunnelServerId", tunnelServerId)
	response.SuccessJSON(ctx, server, "RESTART_TUNNEL_SERVER")
}

// ReloadTunnelServerConfig 重新加载隧道服务器配置
func (c *TunnelServerController) ReloadTunnelServerConfig(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")
	if tunnelServerId == "" {
		response.ErrorJSON(ctx, "参数格式错误: "+tunnelServerId, "START_TUNNEL_SERVER")
		return
	}

	// 调用DAO层重新加载配置
	err := c.tunnelServerDAO.ReloadTunnelServerConfig(ctx.Request.Context(), tunnelServerId)
	if err != nil {
		logger.Error("重新加载隧道服务器配置失败", "tunnelServerId", tunnelServerId, "error", err)
		response.ErrorJSON(ctx, "重新加载配置失败: "+err.Error(), "RELOAD_TUNNEL_SERVER_CONFIG")
		return
	}

	logger.Info("重新加载隧道服务器配置成功", "tunnelServerId", tunnelServerId)
	response.SuccessJSON(ctx, gin.H{
		"tunnelServerId": tunnelServerId,
		"message":        "配置重新加载成功",
	}, "RELOAD_TUNNEL_SERVER_CONFIG")
}

// GetRegisteredClients 获取指定服务器上已注册的客户端列表
// 如果 tunnelServerId 为空，则获取所有服务器的客户端列表
func (c *TunnelServerController) GetRegisteredClients(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")

	// 获取隧道服务器管理器
	manager := server.GetTunnelServerManager()
	if manager == nil {
		logger.Error("隧道服务器管理器未初始化")
		response.ErrorJSON(ctx, "隧道服务器管理器未初始化", "GET_REGISTERED_CLIENTS")
		return
	}

	var allClients []map[string]interface{}

	if tunnelServerId == "" {
		// 如果 tunnelServerId 为空，获取所有服务器的客户端列表
		servers := manager.GetAllServers()
		for _, s := range servers {
			if s == nil {
				continue
			}
			// 获取server的配置以获取serverID
			serverConfig := s.GetConfig()
			if serverConfig == nil {
				continue
			}
			serverID := serverConfig.TunnelServerId
			clients := s.GetConnectedClients()
			// 为每个客户端添加服务器ID字段
			for _, client := range clients {
				if client != nil {
					// 将TunnelClient转换为map，添加tunnelServerId字段
					clientMap := make(map[string]interface{})
					// 使用JSON序列化/反序列化来复制所有字段
					clientJSON, err := json.Marshal(client)
					if err != nil {
						logger.Warn("Failed to marshal client", "error", err.Error())
						continue
					}
					if err := json.Unmarshal(clientJSON, &clientMap); err != nil {
						logger.Warn("Failed to unmarshal client", "error", err.Error())
						continue
					}
					// 添加tunnelServerId字段
					clientMap["tunnelServerId"] = serverID
					allClients = append(allClients, clientMap)
				}
			}
		}
		logger.Info("获取所有服务器的已注册客户端列表成功", "serverCount", len(servers), "clientCount", len(allClients))
	} else {
		// 获取指定服务器的客户端列表
		clients := manager.GetRegisteredClients(tunnelServerId)
		if clients == nil {
			// 如果返回 nil，可能是服务器不存在，返回空列表
			clients = []*types.TunnelClient{}
		}
		// 为每个客户端添加服务器ID字段
		for _, client := range clients {
			if client != nil {
				// 将TunnelClient转换为map，添加tunnelServerId字段
				clientMap := make(map[string]interface{})
				// 使用JSON序列化/反序列化来复制所有字段
				clientJSON, err := json.Marshal(client)
				if err != nil {
					logger.Warn("Failed to marshal client", "error", err.Error())
					continue
				}
				if err := json.Unmarshal(clientJSON, &clientMap); err != nil {
					logger.Warn("Failed to unmarshal client", "error", err.Error())
					continue
				}
				// 添加tunnelServerId字段
				clientMap["tunnelServerId"] = tunnelServerId
				allClients = append(allClients, clientMap)
			}
		}
		logger.Info("获取已注册客户端列表成功", "tunnelServerId", tunnelServerId, "count", len(allClients))
	}

	response.SuccessJSON(ctx, allClients, "GET_REGISTERED_CLIENTS")
}

// GetRegisteredServices 获取指定服务器上已注册的服务列表
// 如果 tunnelServerId 为空，则获取所有服务器的服务列表
func (c *TunnelServerController) GetRegisteredServices(ctx *gin.Context) {
	tunnelServerId := request.GetParam(ctx, "tunnelServerId")

	// 获取隧道服务器管理器
	manager := server.GetTunnelServerManager()
	if manager == nil {
		logger.Error("隧道服务器管理器未初始化")
		response.ErrorJSON(ctx, "隧道服务器管理器未初始化", "GET_REGISTERED_SERVICES")
		return
	}

	var allServices []map[string]interface{}

	if tunnelServerId == "" {
		// 如果 tunnelServerId 为空，获取所有服务器的服务列表
		servers := manager.GetAllServers()
		for _, s := range servers {
			if s == nil {
				continue
			}
			serverConfig := s.GetConfig()
			if serverConfig == nil {
				continue
			}
			serverID := serverConfig.TunnelServerId
			clients := s.GetConnectedClients()
			for _, client := range clients {
				if client != nil && client.Services != nil {
					for _, service := range client.Services {
						if service != nil {
							// 将TunnelService转换为map，添加tunnelServerId和tunnelClientId字段
							serviceMap := make(map[string]interface{})
							serviceJSON, err := json.Marshal(service)
							if err != nil {
								logger.Warn("Failed to marshal service", "error", err.Error())
								continue
							}
							if err := json.Unmarshal(serviceJSON, &serviceMap); err != nil {
								logger.Warn("Failed to unmarshal service", "error", err.Error())
								continue
							}
							// 添加tunnelServerId和tunnelClientId字段
							serviceMap["tunnelServerId"] = serverID
							serviceMap["tunnelClientId"] = client.TunnelClientId
							allServices = append(allServices, serviceMap)
						}
					}
				}
			}
		}
		logger.Info("获取所有服务器的已注册服务列表成功", "serviceCount", len(allServices))
	} else {
		// 获取指定服务器的服务列表
		services := manager.GetRegisteredServices(tunnelServerId)
		if services == nil {
			services = []*types.TunnelService{}
		}
		// 获取服务器以获取客户端信息
		server := manager.GetServer(tunnelServerId)
		if server != nil {
			clients := server.GetConnectedClients()
			clientMap := make(map[string]*types.TunnelClient)
			for _, client := range clients {
				if client != nil {
					clientMap[client.TunnelClientId] = client
				}
			}
			// 为每个服务添加服务器ID和客户端ID字段
			for _, service := range services {
				if service != nil {
					serviceMap := make(map[string]interface{})
					serviceJSON, err := json.Marshal(service)
					if err != nil {
						logger.Warn("Failed to marshal service", "error", err.Error())
						continue
					}
					if err := json.Unmarshal(serviceJSON, &serviceMap); err != nil {
						logger.Warn("Failed to unmarshal service", "error", err.Error())
						continue
					}
					// 添加tunnelServerId和tunnelClientId字段
					serviceMap["tunnelServerId"] = tunnelServerId
					serviceMap["tunnelClientId"] = service.TunnelClientId
					allServices = append(allServices, serviceMap)
				}
			}
		}
		logger.Info("获取已注册服务列表成功", "tunnelServerId", tunnelServerId, "count", len(allServices))
	}

	response.SuccessJSON(ctx, allServices, "GET_REGISTERED_SERVICES")
}
