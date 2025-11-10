// Package server 实现隧道服务端核心功能
//
// 本包提供了基于FRP架构的隧道服务器实现，支持：
// - 控制端口和数据端口分离
// - 静态端口映射和动态服务注册
// - 多客户端连接管理
// - 会话跟踪和连接监控
// - 负载均衡和故障转移
// - 实时指标收集和健康检查
package server

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// DefaultTunnelServer 默认隧道服务器实现
//
// DefaultTunnelServer 是 TunnelServer 接口的标准实现，提供完整的隧道服务功能：
//
// 架构组件：
//   - controlServer: 处理客户端控制连接和认证
//   - proxyServer: 管理数据端口和流量转发
//   - sessionManager: 跟踪客户端会话状态
//   - serviceRegistry: 管理动态服务注册
//   - connectionTracker: 监控连接状态和性能
//   - loadBalancer: 提供负载均衡策略
//
// 状态管理：
//   - 线程安全的客户端连接管理
//   - 实时状态统计和指标更新
//   - 优雅启停和资源清理
//
// 使用示例：
//
//	config := &types.TunnelServer{
//		TunnelServerId: "server-001",
//		ControlPort:   7000,
//		AuthToken:     "secret-token",
//	}
//	server := NewTunnelServer(config, storageManager)
//	defer server.Stop(context.Background())
//
//	if err := server.Start(context.Background()); err != nil {
//		log.Fatal(err)
//	}
type DefaultTunnelServer struct {
	config          *types.TunnelServer
	storageManager  storage.RepositoryManager
	controlServer   ControlServer
	proxyServer     ProxyServer        // 反向代理服务器（用于隧道服务）
	staticProxyMgr  StaticProxyManager // 静态代理管理器（用于端口转发和负载均衡）
	serviceRegistry ServiceRegistry
	loadBalancer    LoadBalancer // 负载均衡器（被静态代理管理器使用）

	// 状态管理
	status           *ServerStatus
	connectedClients map[string]*types.TunnelClient
	mutex            sync.RWMutex

	// 控制通道
	stopChan chan struct{}
}

// NewTunnelServer 创建新的隧道服务器实例
//
// 创建一个完全初始化的隧道服务器实例，包含所有必要的组件和配置。
// 服务器创建后处于停止状态，需要调用 Start 方法来启动服务。
//
// 参数:
//   - config: 服务器配置，包含端口、认证等信息
//   - storageManager: 数据存储管理器，用于持久化状态
//
// 返回:
//   - *DefaultTunnelServer: 新创建的服务器实例
//
// 注意：
//   - 返回的服务器实例尚未启动，需要调用 Start() 方法
//   - 配置一旦设置后不可更改
//   - 存储管理器将用于所有数据持久化操作
func NewTunnelServer(config *types.TunnelServer, storageManager storage.RepositoryManager) *DefaultTunnelServer {
	return &DefaultTunnelServer{
		config:           config,
		storageManager:   storageManager,
		connectedClients: make(map[string]*types.TunnelClient),
		stopChan:         make(chan struct{}),
		status: &ServerStatus{
			Status:           types.ServerStatusStopped,
			StartTime:        time.Now(),
			ConnectedClients: 0,
			TotalTraffic:     0,
		},
	}
}

// Start 启动隧道服务器
//
// 执行完整的服务器启动流程，包括：
// 1. 初始化所有核心组件（控制服务器、代理服务器等）
// 2. 启动控制端口监听，准备接收客户端连接
// 3. 加载和启动所有静态代理配置
// 4. 更新服务器状态并持久化到数据库
// 5. 启动后台维护任务（健康检查、指标更新等）
//
// 参数:
//   - ctx: 控制启动过程的上下文，可用于超时和取消
//
// 返回:
//   - error: 启动过程中的任何错误
//
// 错误情况：
//   - 服务器已经在运行中
//   - 组件初始化失败
//   - 控制端口被占用
//   - 静态代理配置错误
//   - 数据库连接问题
//
// 注意：
//   - 此方法是幂等的，重复调用已启动的服务器会返回错误
//   - 启动失败后服务器保持停止状态
//   - 成功启动后，服务器将在后台持续运行
func (s *DefaultTunnelServer) Start(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.status.Status == types.ServerStatusRunning {
		return fmt.Errorf("server is already running")
	}

	logger.Info("Starting tunnel server", map[string]interface{}{
		"serverID":    s.config.TunnelServerId,
		"controlPort": s.config.ControlPort,
	})

	// 初始化组件
	if err := s.initializeComponents(); err != nil {
		return fmt.Errorf("failed to initialize components: %w", err)
	}

	// 启动控制服务器
	if err := s.controlServer.Start(ctx, s.config.ControlAddress, s.config.ControlPort); err != nil {
		return fmt.Errorf("failed to start control server: %w", err)
	}

	// 加载并启动静态代理节点
	if err := s.loadStaticProxies(ctx); err != nil {
		return fmt.Errorf("failed to load static proxies: %w", err)
	}

	// 更新服务器状态
	s.status.Status = types.ServerStatusRunning
	s.status.StartTime = time.Now()

	// 更新数据库状态
	startTime := time.Now()
	if err := s.storageManager.GetTunnelServerRepository().UpdateStatus(ctx, s.config.TunnelServerId, types.ServerStatusRunning, &startTime); err != nil {
		logger.Error("Failed to update server status in database", err)
	}

	// 启动后台任务
	go s.backgroundTasks(ctx)

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
// 4. 更新服务器状态并持久化
// 5. 停止后台任务和组件
//
// 参数:
//   - ctx: 控制停止过程的上下文，可设置超时
//
// 返回:
//   - error: 停止过程中的错误（通常为nil）
//
// 行为特性：
//   - 优雅停止：给予连接和会话充分的清理时间
//   - 幂等操作：多次调用停止的服务器不会报错
//   - 状态保存：停止前会保存所有重要状态到数据库
//   - 资源清理：确保端口、内存、文件句柄等资源完全释放
//
// 注意：
//   - 停止过程可能需要几秒钟来完成所有清理工作
//   - 使用 context 可以强制中断停止过程
//   - 停止后的服务器可以重新启动
func (s *DefaultTunnelServer) Stop(ctx context.Context) error {
	s.mutex.Lock()
	if s.status.Status == types.ServerStatusStopped {
		s.mutex.Unlock()
		return nil
	}

	logger.Info("Stopping tunnel server", map[string]interface{}{
		"serverID": s.config.TunnelServerId,
	})

	// 发送停止信号
	close(s.stopChan)

	// 收集需要注销的客户端列表（不持有锁）
	clientsToUnregister := make(map[string]*types.TunnelClient)
	for clientID, client := range s.connectedClients {
		clientsToUnregister[clientID] = client
	}
	s.mutex.Unlock()

	// 停止控制服务器
	if s.controlServer != nil {
		if err := s.controlServer.Stop(ctx); err != nil {
			logger.Error("Failed to stop control server", err)
		}
	}

	// 停止反向代理服务器
	if s.proxyServer != nil {
		if err := s.proxyServer.Stop(ctx); err != nil {
			logger.Error("Failed to stop proxy server", err)
		}
	}

	// 停止静态代理管理器
	if s.staticProxyMgr != nil {
		if err := s.staticProxyMgr.Stop(ctx); err != nil {
			logger.Error("Failed to stop static proxy manager", err)
		}
	}

	// 在不持有锁的情况下注销所有客户端，避免死锁
	for clientID, client := range clientsToUnregister {
		if err := s.UnregisterClient(ctx, clientID); err != nil {
			logger.Error("Failed to unregister client", err, map[string]interface{}{
				"clientID":   clientID,
				"clientName": client.ClientName,
			})
		}
	}

	// 更新服务器状态
	s.mutex.Lock()
	s.status.Status = types.ServerStatusStopped
	s.status.Uptime = time.Since(s.status.StartTime).Milliseconds()
	s.mutex.Unlock()

	// 更新数据库状态
	if err := s.storageManager.GetTunnelServerRepository().UpdateStatus(ctx, s.config.TunnelServerId, types.ServerStatusStopped, nil); err != nil {
		logger.Error("Failed to update server status in database", err)
	}

	logger.Info("Tunnel server stopped successfully", map[string]interface{}{
		"serverID": s.config.TunnelServerId,
		"uptime":   s.status.Uptime,
	})

	return nil
}

// GetStatus 获取服务器当前运行状态
//
// 返回服务器的实时状态信息，包括运行状态、连接统计、性能指标等。
// 此方法线程安全，可以被多个 goroutine 并发调用。
//
// 返回的状态信息：
//   - Status: 服务器运行状态（运行中/已停止/错误）
//   - StartTime: 服务器启动时间
//   - Uptime: 运行时长（毫秒）
//   - ConnectedClients: 已连接的客户端数量
//   - ActiveSessions: 活跃会话数
//   - ActiveConnections: 活跃连接数
//   - TotalTraffic: 总流量统计
//
// 返回:
//   - ServerStatus: 服务器状态的快照
//
// 注意：
//   - 返回的是状态的副本，修改不会影响服务器实际状态
//   - 统计数据是实时计算的，反映当前准确状况
//   - 频繁调用此方法不会影响服务器性能
func (s *DefaultTunnelServer) GetStatus() ServerStatus {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	status := *s.status
	if s.status.Status == types.ServerStatusRunning {
		status.Uptime = time.Since(s.status.StartTime).Milliseconds()
	}

	status.ConnectedClients = len(s.connectedClients)

	return status
}

// GetConfig 获取服务器配置信息
//
// 返回服务器的完整配置，包括端口、认证、域名等设置。
//
// 返回:
//   - *types.TunnelServer: 服务器配置对象的引用
//
// 注意：
//   - 返回的是配置对象的引用，不应修改其内容
//   - 配置在服务器生命周期内保持不变
//   - 用于调试、监控和配置检查
func (s *DefaultTunnelServer) GetConfig() *types.TunnelServer {
	return s.config
}

// GetServiceRegistry 获取服务注册器
//
// 返回服务注册器实例，用于管理动态服务注册和端口分配。
//
// 返回:
//   - ServiceRegistry: 服务注册器接口实例
func (s *DefaultTunnelServer) GetServiceRegistry() ServiceRegistry {
	return s.serviceRegistry
}

// RegisterClient 注册新的隧道客户端
//
// 当客户端与服务器建立控制连接后，通过此方法完成客户端注册。
// 注册过程包括身份验证、状态更新和数据库记录。
//
// 注册流程：
// 1. 验证客户端认证令牌
// 2. 检查并处理重复注册
// 3. 更新客户端连接状态
// 4. 持久化客户端信息到数据库
// 5. 添加到活跃客户端列表
//
// 参数:
//   - ctx: 请求上下文
//   - client: 要注册的客户端信息
//
// 返回:
//   - error: 注册过程中的错误
//
// 错误情况：
//   - 认证令牌无效
//   - 数据库操作失败
//   - 客户端信息不完整
//
// 注意：
//   - 重复注册会更新现有客户端信息
//   - 注册成功后客户端可以开始提供服务
//   - 注册信息会持久化到数据库
func (s *DefaultTunnelServer) RegisterClient(ctx context.Context, client *types.TunnelClient) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证客户端认证
	if client.AuthToken != s.config.AuthToken {
		return fmt.Errorf("invalid auth token for client %s", client.ClientName)
	}

	// 检查客户端名称是否已存在
	if existingClient, exists := s.connectedClients[client.TunnelClientId]; exists {
		logger.Warn("Client already registered, updating", map[string]interface{}{
			"clientID":   client.TunnelClientId,
			"clientName": client.ClientName,
			"oldIP":      existingClient.ClientIpAddress,
			"newIP":      client.ClientIpAddress,
		})
	}

	// 更新客户端连接状态
	client.ConnectionStatus = types.ConnectionStatusConnected
	connectTime := time.Now()
	client.LastConnectTime = &connectTime

	// 保存到数据库
	if err := s.storageManager.GetTunnelClientRepository().UpdateConnectionStatus(ctx, client.TunnelClientId, types.ConnectionStatusConnected, &connectTime); err != nil {
		return fmt.Errorf("failed to update client connection status: %w", err)
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
// 确保客户端相关的所有资源和服务都得到正确清理。
//
// 注销流程：
// 1. 关闭客户端的所有活跃会话
// 2. 注销客户端提供的所有服务
// 3. 更新客户端连接状态为已断开
// 4. 从活跃客户端列表中移除
// 5. 记录注销事件
//
// 参数:
//   - ctx: 请求上下文
//   - clientID: 要注销的客户端唯一标识
//
// 返回:
//   - error: 注销过程中的错误
//
// 错误情况：
//   - 客户端不存在
//   - 会话或服务清理失败
//   - 数据库操作失败
//
// 注意：
//   - 注销操作不可逆，客户端需重新连接和注册
//   - 会自动清理客户端相关的所有资源
//   - 注销过程中的错误不会阻止客户端从列表中移除
func (s *DefaultTunnelServer) UnregisterClient(ctx context.Context, clientID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	client, exists := s.connectedClients[clientID]
	if !exists {
		return fmt.Errorf("client %s is not registered", clientID)
	}

	// 注销客户端的所有服务
	// 注意：会话管理由 control_server 负责，客户端断开时会自动清理
	if s.serviceRegistry != nil {
		services, err := s.serviceRegistry.GetServicesByClient(ctx, clientID)
		if err == nil {
			for _, service := range services {
				if err := s.serviceRegistry.UnregisterService(ctx, service.TunnelServiceId); err != nil {
					logger.Error("Failed to unregister client service", err, map[string]interface{}{
						"serviceID": service.TunnelServiceId,
					})
				}
			}
		}
	}

	// 更新客户端连接状态
	disconnectTime := time.Now()
	if err := s.storageManager.GetTunnelClientRepository().UpdateConnectionStatus(ctx, clientID, types.ConnectionStatusDisconnected, &disconnectTime); err != nil {
		logger.Error("Failed to update client disconnection status", err)
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

// GetConnectedClients 获取所有已连接的客户端列表
//
// 返回当前与服务器保持活跃连接的所有客户端信息。
// 此方法线程安全，可以被多个 goroutine 并发调用。
//
// 返回:
//   - []*types.TunnelClient: 已连接客户端的副本列表
//
// 用途：
//   - 监控和管理界面显示客户端状态
//   - 批量操作（如广播消息）
//   - 统计和报告生成
//   - 调试和故障排除
//
// 注意：
//   - 返回的是客户端列表的副本，修改不会影响实际状态
//   - 列表包含完整的客户端信息（IP、连接时间等）
//   - 顺序不保证，不应依赖特定的排序
func (s *DefaultTunnelServer) GetConnectedClients() []*types.TunnelClient {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	clients := make([]*types.TunnelClient, 0, len(s.connectedClients))
	for _, client := range s.connectedClients {
		clients = append(clients, client)
	}

	return clients
}

// BroadcastMessage 向所有已连接客户端广播消息
//
// 向当前所有活跃的客户端发送相同的消息。常用于服务器配置更新、
// 系统通知或协调多个客户端的行为。
//
// 广播流程：
// 1. 获取当前所有已连接客户端的快照
// 2. 遍历客户端列表，逐一发送消息
// 3. 记录发送结果和任何错误
// 4. 返回整体发送结果
//
// 参数:
//   - ctx: 控制广播过程的上下文
//   - message: 要广播的消息内容（字节数组）
//
// 返回:
//   - error: 广播过程中的错误（当前实现总是返回 nil）
//
// 使用场景：
//   - 服务器配置变更通知
//   - 系统维护通知
//   - 集群状态同步
//   - 紧急事件通知
//
// 注意：
//   - 广播是尽力而为的，部分客户端发送失败不会影响其他客户端
//   - 消息内容和格式由调用方定义
//   - 当前是存根实现，需要具体的消息传输机制
func (s *DefaultTunnelServer) BroadcastMessage(ctx context.Context, message []byte) error {
	s.mutex.RLock()
	clients := make([]*types.TunnelClient, 0, len(s.connectedClients))
	for _, client := range s.connectedClients {
		clients = append(clients, client)
	}
	s.mutex.RUnlock()

	for _, client := range clients {
		// TODO: 实现消息发送逻辑
		logger.Debug("Broadcasting message to client", map[string]interface{}{
			"clientID":    client.TunnelClientId,
			"messageSize": len(message),
		})
	}

	return nil
}

// initializeComponents 初始化组件
func (s *DefaultTunnelServer) initializeComponents() error {
	// 创建组件工厂
	factory := NewComponentFactory()

	// 初始化负载均衡器（需要先初始化，因为静态代理管理器依赖它）
	s.loadBalancer = factory.CreateLoadBalancer("round_robin")

	// 初始化控制服务器
	s.controlServer = factory.CreateControlServer(s)

	// 初始化反向代理服务器（用于隧道服务）
	s.proxyServer = factory.CreateProxyServer(s, s.controlServer)

	// 设置代理服务器的控制服务器引用
	s.controlServer.SetProxyServer(s.proxyServer)

	// 初始化静态代理管理器（用于端口转发和负载均衡）
	s.staticProxyMgr = NewStaticProxyManager(s.config.TunnelServerId, s.storageManager, s.loadBalancer)

	// 初始化服务注册器
	s.serviceRegistry = factory.CreateServiceRegistry(s.storageManager)

	// 设置服务注册器的代理服务器引用
	if sr, ok := s.serviceRegistry.(*serviceRegistry); ok {
		sr.SetProxyServer(s.proxyServer)
	}

	return nil
}

// loadStaticProxies 加载并启动静态代理节点
//
// 静态代理节点用于端口转发和负载均衡场景，由独立的静态代理管理器管理。
// 该方法在服务器启动时调用，负责初始化和启动所有静态代理。
//
// 使用场景：
//   - 端口转发：将监听端口的请求转发到目标地址
//   - 负载均衡：使用负载均衡器在多个后端节点间分配请求
//   - 高可用：支持后端节点的健康检查和故障转移
func (s *DefaultTunnelServer) loadStaticProxies(ctx context.Context) error {
	logger.Info("Loading static proxies", map[string]interface{}{
		"serverID": s.config.TunnelServerId,
	})

	// 初始化静态代理管理器（加载配置）
	if err := s.staticProxyMgr.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to initialize static proxy manager: %w", err)
	}

	// 启动所有静态代理
	if err := s.staticProxyMgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start static proxies: %w", err)
	}

	// 获取活跃代理数量
	activeProxies := s.staticProxyMgr.GetActiveProxies()
	logger.Info("Static proxies loaded successfully", map[string]interface{}{
		"serverID":   s.config.TunnelServerId,
		"proxyCount": len(activeProxies),
	})

	return nil
}

// backgroundTasks 后台任务
// 注意：各组件（control_server, proxy_server, static_proxy_manager）都有自己的维护逻辑
// 这里只做服务器级别的简单统计更新
func (s *DefaultTunnelServer) backgroundTasks(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		case <-ticker.C:
			// 更新服务器级别的统计信息
			s.updateMetrics(ctx)
		}
	}
}

// updateMetrics 更新指标
func (s *DefaultTunnelServer) updateMetrics(ctx context.Context) {
	s.mutex.RLock()
	clientCount := len(s.connectedClients)
	s.mutex.RUnlock()

	s.status.ConnectedClients = clientCount
}
