// Package server 实现隧道服务端核心功能
//
// 本包提供了基于FRP架构的隧道服务器实现，支持：
// - 控制端口和数据端口分离
// - 动态服务注册
// - 多客户端连接管理
// - 会话跟踪和连接监控
// - 实时指标收集和健康检查
package server

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// ServerStatus 服务器状态
type ServerStatus struct {
	Status           string    `json:"status"`
	StartTime        time.Time `json:"startTime"`
	Uptime           int64     `json:"uptime"`
	ConnectedClients int       `json:"connectedClients"`
	TotalTraffic     int64     `json:"totalTraffic"`
}

// DefaultTunnelServer 默认隧道服务器实现
//
// DefaultTunnelServer 是 TunnelServer 接口的标准实现，提供完整的隧道服务功能：
//
// 架构组件：
//   - 内置控制连接处理（原 controlServer）
//   - proxyServer: 管理数据端口和流量转发
//
// 客户端管理：
//   - connectedClients: 已连接的客户端列表
//   - 每个 TunnelClient 维护自己的 Services 列表和控制连接
//
// 使用示例：
//
//	config := &types.TunnelServer{
//		TunnelServerId: "server-001",
//		ControlPort:   7000,
//		AuthToken:     "secret-token",
//	}
//	server := NewTunnelServer(config)
//	defer server.Stop(context.Background())
//
//	if err := server.Start(context.Background()); err != nil {
//		log.Fatal(err)
//	}
//
// clientConnection 客户端连接信息
// 每个客户端的控制连接独立管理
type clientConnection struct {
	conn   net.Conn      // 控制连接
	writer *bufio.Writer // 写入器
	mutex  sync.Mutex    // 写入互斥锁
}

type DefaultTunnelServer struct {
	config      *types.TunnelServer
	proxyServer *proxyServer                        // 反向代理服务器（用于隧道服务）
	repository  *storage.TunnelServerRepositoryImpl // 数据库存储接口

	// 控制连接管理（原 controlServer 的字段）
	controlListener  net.Listener
	heartbeatTimeout time.Duration

	// 客户端管理: clientID -> TunnelClient
	// 每个 TunnelClient 维护自己的 Services 列表
	connectedClients map[string]*types.TunnelClient

	// 连接管理: clientID -> clientConnection
	// 独立管理每个客户端的控制连接（连接、写入器、互斥锁）
	clientConnections map[string]*clientConnection

	mutex sync.RWMutex // 保护 connectedClients 和 clientConnections

	// 状态管理
	running   bool
	startTime time.Time

	// 控制通道
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	stopChan chan struct{}
}

// NewTunnelServer 创建新的隧道服务器实例
//
// 创建一个完全初始化的隧道服务器实例，包含所有必要的组件和配置。
// 服务器创建后处于停止状态，需要调用 Start 方法来启动服务。
//
// 参数:
//   - config: 服务器配置，包含端口、认证等信息
//   - repository: 数据库存储接口（可选，用于状态持久化）
//
// 返回:
//   - *DefaultTunnelServer: 新创建的服务器实例
//
// 注意：
//   - 返回的服务器实例尚未启动，需要调用 Start() 方法
//   - 配置一旦设置后不可更改
func NewTunnelServer(config *types.TunnelServer, repository *storage.TunnelServerRepositoryImpl) *DefaultTunnelServer {
	return &DefaultTunnelServer{
		config:            config,
		repository:        repository,
		connectedClients:  make(map[string]*types.TunnelClient),
		clientConnections: make(map[string]*clientConnection),
		heartbeatTimeout:  180 * time.Second,
		stopChan:          make(chan struct{}),
	}
}

// Start 启动隧道服务器
//
// 执行完整的服务器启动流程，包括：
// 1. 初始化所有核心组件（代理服务器等）
// 2. 启动控制端口监听，准备接收客户端连接
// 3. 启动后台维护任务（心跳检查等）
//
// 参数:
//   - ctx: 控制启动过程的上下文，可用于超时和取消
//
// 返回:
//   - error: 启动过程中的任何错误
func (s *DefaultTunnelServer) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.running {
		return fmt.Errorf("server is already running")
	}

	logger.Info("Starting tunnel server", map[string]interface{}{
		"serverID":    s.config.TunnelServerId,
		"controlPort": s.config.ControlPort,
	})

	// 创建服务器上下文
	s.ctx, s.cancel = context.WithCancel(context.Background())

	// 初始化反向代理服务器（用于隧道服务）
	s.proxyServer = NewProxyServerImpl(s)

	// 启动控制端口监听
	if err := s.startControlListener(); err != nil {
		return fmt.Errorf("failed to start control listener: %w", err)
	}

	// 启动接受连接的协程
	s.wg.Add(1)
	go s.acceptConnections()

	// 启动心跳检查协程
	s.wg.Add(1)
	go s.heartbeatChecker()

	// 更新服务器状态
	s.running = true
	s.startTime = time.Now()
	s.stopChan = make(chan struct{})

	// 更新数据库状态
	if s.repository != nil {
		s.config.ServerStatus = types.ServerStatusRunning
		s.config.StartTime = &s.startTime
		if err := s.repository.Update(ctx, s.config); err != nil {
			logger.Error("Failed to update server status in database", err, map[string]interface{}{
				"serverID": s.config.TunnelServerId,
				"status":   types.ServerStatusRunning,
			})
			// 不返回错误，因为服务器已经启动成功
		}
	}

	logger.Info("Tunnel server started successfully", map[string]interface{}{
		"serverID":    s.config.TunnelServerId,
		"controlPort": s.config.ControlPort,
	})

	return nil
}

// Stop 优雅停止隧道服务器
//
// 执行完整的服务器停止流程，确保所有资源正确清理：
// 1. 停止接收新的客户端连接
// 2. 关闭所有活跃的代理端口
// 3. 断开所有客户端连接并清理会话
// 4. 停止后台任务和组件
//
// 参数:
//   - ctx: 控制停止过程的上下文，可设置超时
//
// 返回:
//   - error: 停止过程中的错误（通常为nil）
func (s *DefaultTunnelServer) Stop(ctx context.Context) error {
	s.mutex.Lock()
	if !s.running {
		s.mutex.Unlock()
		return nil
	}

	logger.Info("Stopping tunnel server", map[string]interface{}{
		"serverID": s.config.TunnelServerId,
	})

	// 1. 发送停止信号
	if s.cancel != nil {
		s.cancel()
	}
	close(s.stopChan)

	// 2. 关闭控制监听器（停止接收新连接）
	if s.controlListener != nil {
		s.controlListener.Close()
	}

	// 3. 停止反向代理服务器（在关闭连接前停止，避免新的代理请求）
	if s.proxyServer != nil {
		if err := s.proxyServer.Stop(ctx); err != nil {
			logger.Error("Failed to stop proxy server", err)
		}
	}

	// 4. 收集并关闭所有客户端连接
	connsToClose := make([]*clientConnection, 0, len(s.clientConnections))
	for _, conn := range s.clientConnections {
		connsToClose = append(connsToClose, conn)
	}
	s.mutex.Unlock()

	// 关闭所有客户端连接
	for _, clientConn := range connsToClose {
		if clientConn.conn != nil {
			clientConn.conn.Close()
		}
	}

	// 5. 等待所有 goroutine 退出（带超时保护）
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All goroutines stopped gracefully", map[string]interface{}{
			"serverID": s.config.TunnelServerId,
		})
	case <-ctx.Done():
		logger.Warn("Timeout waiting for goroutines to stop, forcing shutdown", map[string]interface{}{
			"serverID": s.config.TunnelServerId,
		})
	}

	// 6. 清理所有客户端数据
	s.mutex.Lock()
	// 清理已连接客户端
	for clientID := range s.connectedClients {
		delete(s.connectedClients, clientID)
	}
	// 清理客户端连接（避免泄露）
	for clientID := range s.clientConnections {
		delete(s.clientConnections, clientID)
	}
	s.running = false
	s.mutex.Unlock()

	// 7. 更新数据库状态（使用独立的超时上下文）
	if s.repository != nil {
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dbCancel()

		s.config.ServerStatus = types.ServerStatusStopped
		s.config.StartTime = nil
		if err := s.repository.Update(dbCtx, s.config); err != nil {
			logger.Error("Failed to update server status in database", err, map[string]interface{}{
				"serverID": s.config.TunnelServerId,
				"status":   types.ServerStatusStopped,
			})
			// 不返回错误，因为服务器已经停止成功
		}
	}

	logger.Info("Tunnel server stopped successfully", map[string]interface{}{
		"serverID": s.config.TunnelServerId,
	})

	return nil
}

// IsRunning 检查服务器是否正在运行
func (s *DefaultTunnelServer) IsRunning() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.running
}

// GetStatus 获取服务器当前运行状态
//
// 返回服务器的实时状态信息，包括运行状态、连接统计、性能指标等。
// 此方法线程安全，可以被多个 goroutine 并发调用。
//
// 返回:
//   - ServerStatus: 服务器状态的快照
func (s *DefaultTunnelServer) GetStatus() ServerStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	status := ServerStatus{
		Status:           types.ServerStatusStopped,
		StartTime:        s.startTime,
		ConnectedClients: len(s.connectedClients),
	}

	if s.running {
		status.Status = types.ServerStatusRunning
		status.Uptime = time.Since(s.startTime).Milliseconds()
	}

	// 计算总流量
	var totalTraffic int64
	for _, client := range s.connectedClients {
		if client.Services != nil {
			for _, service := range client.Services {
				totalTraffic += service.TotalTraffic
			}
		}
	}
	status.TotalTraffic = totalTraffic

	return status
}

// GetConfig 获取服务器配置信息
//
// 返回:
//   - *types.TunnelServer: 服务器配置对象的引用
func (s *DefaultTunnelServer) GetConfig() *types.TunnelServer {
	return s.config
}

// RegisterClient 注册新的隧道客户端
//
// 当客户端与服务器建立控制连接后，通过此方法完成客户端注册。
//
// 参数:
//   - ctx: 请求上下文
//   - client: 要注册的客户端信息
//
// 返回:
//   - error: 注册过程中的错误
func (s *DefaultTunnelServer) RegisterClient(ctx context.Context, client *types.TunnelClient) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证客户端认证
	if client.AuthToken != s.config.AuthToken {
		return fmt.Errorf("invalid auth token for client %s", client.ClientName)
	}

	// 检查客户端是否已存在
	if existingClient, exists := s.connectedClients[client.TunnelClientId]; exists {
		logger.Warn("Client already registered, updating", map[string]interface{}{
			"clientID":   client.TunnelClientId,
			"clientName": client.ClientName,
			"oldIP":      existingClient.ClientIpAddress,
			"newIP":      client.ClientIpAddress,
		})
		// 关闭旧连接
		if oldConn, exists := s.clientConnections[client.TunnelClientId]; exists {
			if oldConn.conn != nil {
				go oldConn.conn.Close()
			}
			delete(s.clientConnections, client.TunnelClientId)
		}
	}

	// 更新客户端连接状态
	client.ConnectionStatus = types.ConnectionStatusConnected
	connectTime := time.Now()
	client.LastConnectTime = &connectTime

	// 初始化服务列表
	if client.Services == nil {
		client.Services = make(map[string]*types.TunnelService)
	}

	// 注册客户端
	s.connectedClients[client.TunnelClientId] = client

	logger.Info("Client registered successfully", map[string]interface{}{
		"clientID":     client.TunnelClientId,
		"clientName":   client.ClientName,
		"clientIP":     client.ClientIpAddress,
		"totalClients": len(s.connectedClients),
	})

	return nil
}

// UnregisterClient 注销指定的隧道客户端
//
// 当客户端断开连接或需要强制下线时，执行完整的客户端注销流程。
// 会自动注销客户端的所有服务。
//
// 参数:
//   - ctx: 请求上下文
//   - clientID: 要注销的客户端唯一标识
//
// 返回:
//   - error: 注销过程中的错误
func (s *DefaultTunnelServer) UnregisterClient(ctx context.Context, clientID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.connectedClients[clientID]
	if !exists {
		return fmt.Errorf("client %s is not registered", clientID)
	}

	// 注销客户端的所有服务对应的代理
	if s.proxyServer != nil && client.Services != nil {
		for serviceID := range client.Services {
			if err := s.proxyServer.StopProxy(ctx, serviceID); err != nil {
				logger.Error("Failed to stop proxy for service", err, map[string]interface{}{
					"serviceID": serviceID,
				})
			}
		}
	}

	// 移除客户端注册
	delete(s.connectedClients, clientID)

	logger.Info("Client unregistered successfully", map[string]interface{}{
		"clientID":     clientID,
		"clientName":   client.ClientName,
		"totalClients": len(s.connectedClients),
	})

	return nil
}

// RegisterService 为客户端注册服务
//
// 参数:
//   - ctx: 请求上下文
//   - clientID: 客户端ID
//   - service: 要注册的服务
//
// 返回:
//   - error: 注册过程中的错误
func (s *DefaultTunnelServer) RegisterService(ctx context.Context, clientID string, service *types.TunnelService) error {
	s.mutex.Lock()
	client, exists := s.connectedClients[clientID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("client %s is not registered", clientID)
	}

	// 初始化服务列表（如果需要）
	if client.Services == nil {
		client.Services = make(map[string]*types.TunnelService)
	}

	// 注册服务到客户端
	client.Services[service.TunnelServiceId] = service
	client.ServiceCount = len(client.Services)
	s.mutex.Unlock()

	// 启动代理（如果需要）
	if s.proxyServer != nil && service.RemotePort != nil {
		if err := s.proxyServer.StartProxy(ctx, service); err != nil {
			return fmt.Errorf("failed to start proxy for service: %w", err)
		}
	}

	logger.Info("Service registered successfully", map[string]interface{}{
		"clientID":    clientID,
		"serviceID":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
	})

	return nil
}

// UnregisterService 注销客户端的服务
//
// 参数:
//   - ctx: 请求上下文
//   - clientID: 客户端ID
//   - serviceID: 服务ID
//
// 返回:
//   - error: 注销过程中的错误
func (s *DefaultTunnelServer) UnregisterService(ctx context.Context, clientID string, serviceID string) error {
	s.mutex.Lock()
	client, exists := s.connectedClients[clientID]
	if !exists {
		s.mutex.Unlock()
		return fmt.Errorf("client %s is not registered", clientID)
	}

	// 从客户端注销服务
	if client.Services != nil {
		delete(client.Services, serviceID)
		client.ServiceCount = len(client.Services)
	}
	s.mutex.Unlock()

	// 停止代理
	if s.proxyServer != nil {
		if err := s.proxyServer.StopProxy(ctx, serviceID); err != nil {
			logger.Error("Failed to stop proxy for service", err, map[string]interface{}{
				"serviceID": serviceID,
			})
		}
	}

	logger.Info("Service unregistered successfully", map[string]interface{}{
		"clientID":  clientID,
		"serviceID": serviceID,
	})

	return nil
}

// GetConnectedClients 获取所有已连接的客户端列表
//
// 返回:
//   - []*types.TunnelClient: 已连接客户端的列表
func (s *DefaultTunnelServer) GetConnectedClients() []*types.TunnelClient {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := make([]*types.TunnelClient, 0, len(s.connectedClients))
	for _, client := range s.connectedClients {
		clients = append(clients, client)
	}

	return clients
}

// GetConnectedClient 获取指定的已连接客户端
//
// 参数:
//   - clientID: 客户端ID
//
// 返回:
//   - *types.TunnelClient: 已连接客户端，如果不存在则返回 nil
func (s *DefaultTunnelServer) GetConnectedClient(clientID string) *types.TunnelClient {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.connectedClients[clientID]
}

// GetClientServices 获取客户端的所有服务
//
// 参数:
//   - clientID: 客户端ID
//
// 返回:
//   - []*types.TunnelService: 服务列表
//   - error: 错误信息
func (s *DefaultTunnelServer) GetClientServices(clientID string) ([]*types.TunnelService, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	client, exists := s.connectedClients[clientID]
	if !exists {
		return nil, fmt.Errorf("client %s is not registered", clientID)
	}

	services := make([]*types.TunnelService, 0, len(client.Services))
	for _, service := range client.Services {
		services = append(services, service)
	}

	return services, nil
}

// GetClientService 获取客户端的指定服务
//
// 参数:
//   - clientID: 客户端ID
//   - serviceID: 服务ID
//
// 返回:
//   - *types.TunnelService: 服务，如果不存在则返回 nil
func (s *DefaultTunnelServer) GetClientService(clientID string, serviceID string) *types.TunnelService {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	client, exists := s.connectedClients[clientID]
	if !exists || client.Services == nil {
		return nil
	}

	return client.Services[serviceID]
}

// GetProxyServer 获取代理服务器
func (s *DefaultTunnelServer) GetProxyServer() *proxyServer {
	return s.proxyServer
}

// GetActiveConnections 获取活跃连接数
//
// 返回:
//   - int: 当前活跃的控制连接数量
func (s *DefaultTunnelServer) GetActiveConnections() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	for _, client := range s.connectedClients {
		if client.Authenticated {
			count++
		}
	}
	return count
}

// SendMessageToClient 向指定客户端发送消息
//
// 参数:
//   - clientID: 目标客户端ID
//   - message: 要发送的控制消息
//
// 返回:
//   - error: 发送失败时返回错误
func (s *DefaultTunnelServer) SendMessageToClient(clientID string, message *types.ControlMessage) error {
	s.mutex.RLock()
	client, clientExists := s.connectedClients[clientID]
	clientConn, connExists := s.clientConnections[clientID]
	s.mutex.RUnlock()

	if !clientExists || client == nil {
		return fmt.Errorf("client %s not found or not connected", clientID)
	}

	if !client.Authenticated {
		return fmt.Errorf("client %s not authenticated", clientID)
	}

	if !connExists || clientConn == nil || clientConn.conn == nil {
		return fmt.Errorf("client %s connection is nil", clientID)
	}

	// 使用连接的互斥锁保护写入操作
	clientConn.mutex.Lock()
	defer clientConn.mutex.Unlock()

	// 发送消息
	return s.sendControlMessageToClient(clientConn.conn, clientConn.writer, message)
}

// startControlListener 启动控制端口监听
func (s *DefaultTunnelServer) startControlListener() error {
	listenAddr := fmt.Sprintf("%s:%d", s.config.ControlAddress, s.config.ControlPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	s.controlListener = listener

	logger.Info("Control listener started", map[string]interface{}{
		"address": s.config.ControlAddress,
		"port":    s.config.ControlPort,
	})

	return nil
}
