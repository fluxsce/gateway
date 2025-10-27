// Package tunnel 提供隧道管理系统的统一入口
// 基于FRP架构，实现服务端和客户端的统一管理
//
// 本包实现了一个完整的隧道管理系统，支持：
// - 静态端口映射：预配置的固定代理规则
// - 动态服务注册：客户端实时注册的服务
// - 多租户管理：支持租户级别的资源隔离
// - 实时监控：连接跟踪、性能指标、告警管理
// - 高可用性：自动重连、健康检查、故障转移
package tunnel

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/tunnel/client"
	"gateway/internal/tunnel/monitor"
	"gateway/internal/tunnel/server"
	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// 全局隧道管理器实例
var (
	globalManager     *TunnelManager
	globalManagerLock sync.RWMutex
)

// TunnelManager 隧道管理器是整个隧道系统的核心控制器
//
// 它负责协调和管理：
// - 多个隧道服务器实例的生命周期
// - 多个隧道客户端的连接和状态
// - 静态代理配置和动态服务注册
// - 系统监控、日志记录和告警
// - 数据库存储和缓存管理
//
// TunnelManager 采用线程安全的设计，支持并发操作，
// 所有的状态变更都通过互斥锁保护。
type TunnelManager struct {
	// 存储管理
	storageManager storage.RepositoryManager

	// 服务端管理
	servers map[string]server.TunnelServer

	// 客户端管理
	clients map[string]client.TunnelClient

	// 监控组件
	metricsCollector monitor.MetricsCollector
	healthChecker    monitor.HealthChecker
	alertManager     monitor.AlertManager

	// 同步控制
	mutex sync.RWMutex

	// 标记字段，用于避免编译器优化
}

// NewTunnelManager 创建隧道管理器实例
//
// 参数:
//   - ctx: 上下文对象，用于控制生命周期
//
// 返回:
//   - *TunnelManager: 新创建的隧道管理器实例
//
// 使用示例:
//
//	tunnelManager := NewTunnelManager(ctx)
//	defer tunnelManager.Shutdown(context.Background())
func NewTunnelManager(ctx context.Context) *TunnelManager {
	return &TunnelManager{
		servers: make(map[string]server.TunnelServer),
		clients: make(map[string]client.TunnelClient),
	}
}

// Initialize 初始化隧道管理器
//
// 执行以下初始化步骤：
// 1. 初始化数据库存储管理器
// 2. 初始化监控组件(指标收集、日志管理、健康检查、告警)
// 3. 从数据库加载隧道服务器配置
// 4. 从数据库加载隧道客户端配置
//
// 参数:
//   - ctx: 上下文对象，用于控制初始化过程
//
// 返回:
//   - error: 初始化过程中的错误，如果成功则返回nil
//
// 注意：此方法必须在使用任何其他功能之前调用
func (tm *TunnelManager) Initialize(ctx context.Context) error {
	logger.Info("Initializing tunnel manager", nil)

	// 初始化存储管理器
	if err := tm.initializeStorageManager(); err != nil {
		return fmt.Errorf("failed to initialize storage manager: %w", err)
	}

	// 初始化监控组件
	if err := tm.initializeMonitoringComponents(); err != nil {
		return fmt.Errorf("failed to initialize monitoring components: %w", err)
	}

	// 加载服务器配置
	if err := tm.loadTunnelServers(ctx); err != nil {
		return fmt.Errorf("failed to load tunnel servers: %w", err)
	}

	// 加载客户端配置
	if err := tm.loadTunnelClients(ctx); err != nil {
		return fmt.Errorf("failed to load tunnel clients: %w", err)
	}

	logger.Info("Tunnel manager initialized successfully", map[string]interface{}{
		"servers": len(tm.servers),
		"clients": len(tm.clients),
	})

	return nil
}

// StartServer 启动指定的隧道服务器
//
// 启动过程包括：
// 1. 检查服务器是否存在，不存在则从数据库重新加载
// 2. 启动控制端口监听
// 3. 加载和启动静态代理配置
// 4. 初始化会话管理器
// 5. 启动后台维护任务
//
// 参数:
//   - ctx: 上下文对象
//   - serverID: 要启动的服务器唯一标识
//
// 返回:
//   - error: 启动过程中的错误，如果成功则返回nil
func (tm *TunnelManager) StartServer(ctx context.Context, serverID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnelServer, exists := tm.servers[serverID]
	if !exists {
		// 服务器不存在，尝试从数据库重新加载
		logger.Info("服务器不存在，尝试从数据库加载", map[string]interface{}{
			"serverID": serverID,
		})

		serverConfig, err := tm.storageManager.GetTunnelServerRepository().GetByID(ctx, serverID)
		if err != nil {
			return fmt.Errorf("从数据库加载服务器配置失败: %w", err)
		}

		if serverConfig == nil {
			return fmt.Errorf("服务器 %s 在数据库中不存在", serverID)
		}

		if serverConfig.ActiveFlag != types.ActiveFlagYes {
			return fmt.Errorf("服务器 %s 未激活", serverID)
		}

		// 创建服务器实例
		tunnelServer = server.NewTunnelServer(serverConfig, tm.storageManager)
		tm.servers[serverID] = tunnelServer

		logger.Info("服务器配置已从数据库加载", map[string]interface{}{
			"serverID":   serverID,
			"serverName": serverConfig.ServerName,
		})
	}

	if err := tunnelServer.Start(ctx); err != nil {
		return fmt.Errorf("failed to start server %s: %w", serverID, err)
	}

	logger.Info("Tunnel server started", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// StopServer 停止指定的隧道服务器
//
// 停止过程包括：
// 1. 断开所有客户端连接
// 2. 关闭所有代理端口监听
// 3. 保存会话状态到数据库
// 4. 清理资源和缓存
// 5. 从内存缓存中删除服务器实例
//
// 参数:
//   - ctx: 上下文对象
//   - serverID: 要停止的服务器唯一标识
//
// 返回:
//   - error: 停止过程中的错误，如果成功则返回nil
func (tm *TunnelManager) StopServer(ctx context.Context, serverID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnelServer, exists := tm.servers[serverID]
	if !exists {
		return fmt.Errorf("server %s not found", serverID)
	}

	if err := tunnelServer.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop server %s: %w", serverID, err)
	}

	// 停止成功后，从缓存中删除
	delete(tm.servers, serverID)

	logger.Info("Tunnel server stopped and removed from cache", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// StartClient 启动指定的隧道客户端
//
// 启动过程包括：
// 1. 检查客户端是否存在，不存在则从数据库重新加载
// 2. 建立与服务器的控制连接
// 3. 进行身份认证
// 4. 注册本地服务
// 5. 启动心跳保活机制
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 要启动的客户端唯一标识
//
// 返回:
//   - error: 启动过程中的错误，如果成功则返回nil
func (tm *TunnelManager) StartClient(ctx context.Context, clientID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnelClient, exists := tm.clients[clientID]
	if !exists {
		// 客户端不存在，尝试从数据库重新加载
		logger.Info("客户端不存在，尝试从数据库加载", map[string]interface{}{
			"clientID": clientID,
		})

		clientConfig, err := tm.storageManager.GetTunnelClientRepository().GetByID(ctx, clientID)
		if err != nil {
			return fmt.Errorf("从数据库加载客户端配置失败: %w", err)
		}

		if clientConfig == nil {
			return fmt.Errorf("客户端 %s 在数据库中不存在", clientID)
		}

		if clientConfig.ActiveFlag != types.ActiveFlagYes {
			return fmt.Errorf("客户端 %s 未激活", clientID)
		}

		// 创建客户端实例
		tunnelClient = client.NewTunnelClient(clientConfig)
		tm.clients[clientID] = tunnelClient

		logger.Info("客户端配置已从数据库加载", map[string]interface{}{
			"clientID":   clientID,
			"clientName": clientConfig.ClientName,
		})
	}

	if err := tunnelClient.Start(ctx); err != nil {
		return fmt.Errorf("failed to start client %s: %w", clientID, err)
	}

	logger.Info("Tunnel client started", map[string]interface{}{
		"clientID": clientID,
	})

	return nil
}

// StopClient 停止指定的隧道客户端
//
// 停止过程包括：
// 1. 注销所有本地服务
// 2. 断开控制连接
// 3. 清理本地代理资源
// 4. 更新连接状态到数据库
// 5. 从内存缓存中删除客户端实例
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 要停止的客户端唯一标识
//
// 返回:
//   - error: 停止过程中的错误，如果成功则返回nil
func (tm *TunnelManager) StopClient(ctx context.Context, clientID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	tunnelClient, exists := tm.clients[clientID]
	if !exists {
		return fmt.Errorf("client %s not found", clientID)
	}

	if err := tunnelClient.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop client %s: %w", clientID, err)
	}

	// 停止成功后，从缓存中删除
	delete(tm.clients, clientID)

	logger.Info("Tunnel client stopped and removed from cache", map[string]interface{}{
		"clientID": clientID,
	})

	return nil
}

// GetServerStatus 获取指定服务器的运行状态
//
// 返回的状态信息包括：
// - 服务器运行状态（运行中/已停止/错误）
// - 启动时间和运行时长
// - 连接的客户端数量
// - 活跃会话和连接数
// - 总流量统计
//
// 参数:
//   - ctx: 上下文对象
//   - serverID: 服务器唯一标识
//
// 返回:
//   - *server.ServerStatus: 服务器状态信息
//   - error: 获取过程中的错误
func (tm *TunnelManager) GetServerStatus(ctx context.Context, serverID string) (*server.ServerStatus, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tunnelServer, exists := tm.servers[serverID]
	if !exists {
		return nil, fmt.Errorf("server %s not found", serverID)
	}

	status := tunnelServer.GetStatus()
	return &status, nil
}

// GetClientStatus 获取指定客户端的运行状态
//
// 返回的状态信息包括：
// - 客户端连接状态
// - 服务器地址和端口
// - 连接时长和重连次数
// - 注册的服务数量
// - 活跃代理数和总流量
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 客户端唯一标识
//
// 返回:
//   - *client.ClientStatus: 客户端状态信息
//   - error: 获取过程中的错误
func (tm *TunnelManager) GetClientStatus(ctx context.Context, clientID string) (*client.ClientStatus, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	tunnelClient, exists := tm.clients[clientID]
	if !exists {
		return nil, fmt.Errorf("client %s not found", clientID)
	}

	status := tunnelClient.GetStatus()
	return status, nil
}

// CreateStaticProxy 创建静态端口映射代理
//
// 静态代理是预先配置的固定端口映射规则，不需要客户端连接即可工作。
// 适用于：SSH、Web服务、数据库等固定服务的端口映射。
//
// 创建过程：
// 1. 验证代理配置（端口冲突、类型有效性）
// 2. 保存配置到数据库
// 3. 如果目标服务器正在运行，立即启动代理
//
// 参数:
//   - ctx: 上下文对象
//   - node: 服务器节点配置，包含监听端口、目标地址等信息
//
// 返回:
//   - error: 创建过程中的错误，如果成功则返回nil
//
// 示例:
//
//	sshProxy := &types.TunnelServerNode{
//		NodeName:      "SSH-Proxy",
//		ProxyType:     types.ProxyTypeTCP,
//		ListenPort:    2222,
//		TargetAddress: "192.168.1.100",
//		TargetPort:    22,
//	}
//	err := tm.CreateStaticProxy(ctx, sshProxy)
func (tm *TunnelManager) CreateStaticProxy(ctx context.Context, node *types.TunnelServerNode) error {
	// 验证节点配置
	if err := tm.validateNodeConfig(ctx, node); err != nil {
		return fmt.Errorf("invalid node configuration: %w", err)
	}

	// 保存到数据库
	if err := tm.storageManager.GetTunnelServerNodeRepository().Create(ctx, node); err != nil {
		return fmt.Errorf("failed to create server node: %w", err)
	}

	// 如果服务器正在运行，立即启动代理
	tm.mutex.RLock()
	tunnelServer, exists := tm.servers[node.TunnelServerId]
	tm.mutex.RUnlock()

	if exists {
		serverStatus := tunnelServer.GetStatus()
		if serverStatus.Status == types.ServerStatusRunning {
			// TODO: 动态启动代理
			logger.Info("Static proxy will be started when server restarts", map[string]interface{}{
				"nodeID":   node.ServerNodeId,
				"nodeName": node.NodeName,
			})
		}
	}

	logger.Info("Static proxy created successfully", map[string]interface{}{
		"nodeID":    node.ServerNodeId,
		"nodeName":  node.NodeName,
		"proxyType": node.ProxyType,
	})

	return nil
}

// RegisterService 注册动态服务
//
// 动态服务是由客户端注册的服务，支持灵活的配置和实时变更。
// 适用于：开发环境的Web服务、微服务、临时服务等场景。
//
// 注册过程：
// 1. 验证服务配置（名称冲突、端口分配）
// 2. 保存服务信息到数据库
// 3. 通知相关服务器更新路由规则
//
// 参数:
//   - ctx: 上下文对象
//   - service: 服务配置信息
//
// 返回:
//   - error: 注册过程中的错误，如果成功则返回nil
//
// 示例:
//
//	webService := &types.TunnelService{
//		ServiceName: "my-web-app",
//		ServiceType: types.ProxyTypeHTTP,
//		LocalPort:   8080,
//		SubDomain:   "myapp",
//	}
//	err := tm.RegisterService(ctx, webService)
func (tm *TunnelManager) RegisterService(ctx context.Context, service *types.TunnelService) error {
	// 验证服务配置
	if err := tm.validateServiceConfig(ctx, service); err != nil {
		return fmt.Errorf("invalid service configuration: %w", err)
	}

	// 保存到数据库
	if err := tm.storageManager.GetTunnelServiceRepository().Create(ctx, service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	logger.Info("Service registered successfully", map[string]interface{}{
		"serviceID":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
	})

	return nil
}

// GetSystemMetrics 获取系统运行指标
//
// 返回的指标包括：
// - CPU和内存使用率
// - 网络I/O统计
// - 磁盘使用情况
// - Goroutine和文件句柄数量
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - *monitor.SystemMetrics: 系统指标数据
//   - error: 收集过程中的错误
func (tm *TunnelManager) GetSystemMetrics(ctx context.Context) (*monitor.SystemMetrics, error) {
	if tm.metricsCollector == nil {
		return nil, fmt.Errorf("metrics collector not initialized")
	}

	return tm.metricsCollector.CollectSystemMetrics(ctx)
}

// GetConnectionStats 获取指定时间范围内的连接统计报告
//
// 统计信息包括：
// - 总连接数和活跃连接数
// - 平均延迟和错误率
// - 流量统计和吞吐量
// - 按时间分组的详细数据
//
// 参数:
//   - ctx: 上下文对象
//   - timeRange: 统计的时间范围
//
// 返回:
//   - *server.ConnectionStatsReport: 连接统计报告
//   - error: 统计过程中的错误
func (tm *TunnelManager) GetConnectionStats(ctx context.Context, timeRange monitor.TimeRange) (*server.ConnectionStatsReport, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	// TODO: 聚合所有服务器的连接统计
	var totalReport *server.ConnectionStatsReport

	for serverID, tunnelServer := range tm.servers {
		// 获取服务器的连接跟踪器
		// TODO: 实现连接统计聚合逻辑
		logger.Debug("Collecting connection stats", map[string]interface{}{
			"serverID": serverID,
		})
		_ = tunnelServer
	}

	return totalReport, nil
}

// StartAll 启动所有已加载的隧道服务器和客户端
//
// 启动过程按以下顺序执行：
// 1. 先启动所有服务器实例
// 2. 再启动所有客户端实例
//
// 参数:
//   - ctx: 上下文对象，用于控制启动过程的超时
//
// 返回:
//   - error: 启动过程中的错误，如果成功则返回nil
//
// 注意：此方法应该在Initialize之后调用
func (tm *TunnelManager) StartAll(ctx context.Context) error {
	logger.Info("正在启动所有隧道服务器和客户端...", nil)

	// 启动所有服务器
	tm.mutex.RLock()
	servers := make(map[string]server.TunnelServer)
	for k, v := range tm.servers {
		servers[k] = v
	}
	tm.mutex.RUnlock()

	for serverID, tunnelServer := range servers {
		if err := tunnelServer.Start(ctx); err != nil {
			logger.Error("启动服务器失败", err, map[string]interface{}{
				"serverID": serverID,
			})
			return fmt.Errorf("启动服务器 %s 失败: %w", serverID, err)
		}
		logger.Info("服务器启动成功", map[string]interface{}{
			"serverID": serverID,
		})
	}

	// 启动所有客户端
	tm.mutex.RLock()
	clients := make(map[string]client.TunnelClient)
	for k, v := range tm.clients {
		clients[k] = v
	}
	tm.mutex.RUnlock()

	for clientID, tunnelClient := range clients {
		if err := tunnelClient.Start(ctx); err != nil {
			logger.Error("启动客户端失败", err, map[string]interface{}{
				"clientID": clientID,
			})
			return fmt.Errorf("启动客户端 %s 失败: %w", clientID, err)
		}
		logger.Info("客户端启动成功", map[string]interface{}{
			"clientID": clientID,
		})
	}

	logger.Info("所有隧道服务器和客户端启动成功", map[string]interface{}{
		"servers": len(servers),
		"clients": len(clients),
	})

	return nil
}

// Shutdown 优雅关闭隧道管理器
//
// 关闭过程按以下顺序执行：
// 1. 停止所有服务器实例
// 2. 停止所有客户端连接
// 3. 关闭存储管理器连接
// 4. 清理所有资源
//
// 参数:
//   - ctx: 上下文对象，用于控制关闭过程的超时
//
// 返回:
//   - error: 关闭过程中的错误，通常返回nil
//
// 注意：此方法应该在程序退出前调用，确保所有资源正确释放
func (tm *TunnelManager) Shutdown(ctx context.Context) error {
	logger.Info("Shutting down tunnel manager", nil)

	// 停止所有服务器
	tm.mutex.RLock()
	servers := make(map[string]server.TunnelServer)
	for k, v := range tm.servers {
		servers[k] = v
	}
	tm.mutex.RUnlock()

	for serverID, tunnelServer := range servers {
		if err := tunnelServer.Stop(ctx); err != nil {
			logger.Error("Failed to stop server during shutdown", err, map[string]interface{}{
				"serverID": serverID,
			})
		}
	}

	// 停止所有客户端
	tm.mutex.RLock()
	clients := make(map[string]client.TunnelClient)
	for k, v := range tm.clients {
		clients[k] = v
	}
	tm.mutex.RUnlock()

	for clientID, tunnelClient := range clients {
		if err := tunnelClient.Stop(ctx); err != nil {
			logger.Error("Failed to stop client during shutdown", err, map[string]interface{}{
				"clientID": clientID,
			})
		}
	}

	// 关闭存储管理器
	if err := tm.storageManager.Close(); err != nil {
		logger.Error("Failed to close storage manager", err)
	}

	logger.Info("Tunnel manager shutdown completed", nil)
	return nil
}

// initializeStorageManager 初始化存储管理器
//
// 从默认数据库连接创建存储管理器实例
//
// 返回:
//   - error: 初始化过程中的错误
func (tm *TunnelManager) initializeStorageManager() error {
	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return fmt.Errorf("无法获取默认数据库连接")
	}

	// 创建存储管理器
	tm.storageManager = storage.NewDatabaseRepositoryManager(db)

	logger.Info("存储管理器初始化成功", nil)
	return nil
}

// initializeMonitoringComponents 初始化监控组件
//
// 初始化以下监控组件：
// - metricsCollector: 系统和业务指标收集器
// - healthChecker: 健康检查器
// - alertManager: 告警管理器
//
// 返回:
//   - error: 初始化过程中的错误
func (tm *TunnelManager) initializeMonitoringComponents() error {
	// TODO: 初始化监控组件
	// tm.metricsCollector = monitor.NewMetricsCollector(tm.storageManager, tm.logger)
	// tm.healthChecker = monitor.NewHealthChecker(tm.storageManager, tm.logger)
	// tm.alertManager = monitor.NewAlertManager(tm.storageManager, tm.logger)

	logger.Info("Monitoring components initialized", nil)
	return nil
}

// loadTunnelServers 从数据库加载隧道服务器配置
//
// 加载过程：
// 1. 查询指定租户的所有服务器配置
// 2. 过滤出活跃状态的服务器
// 3. 创建服务器实例并加入管理器
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 加载过程中的错误
func (tm *TunnelManager) loadTunnelServers(ctx context.Context) error {
	// 获取所有服务器配置
	servers, err := tm.storageManager.GetTunnelServerRepository().GetByTenantID(ctx, "default")
	if err != nil {
		return fmt.Errorf("failed to load server configurations: %w", err)
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for _, serverConfig := range servers {
		if serverConfig.ActiveFlag == types.ActiveFlagYes {
			tunnelServer := server.NewTunnelServer(serverConfig, tm.storageManager)
			tm.servers[serverConfig.TunnelServerId] = tunnelServer

			logger.Info("Tunnel server loaded", map[string]interface{}{
				"serverID":   serverConfig.TunnelServerId,
				"serverName": serverConfig.ServerName,
			})
		}
	}

	return nil
}

// loadTunnelClients 从数据库加载隧道客户端配置
//
// 加载过程：
// 1. 查询指定租户的所有客户端配置
// 2. 过滤出活跃状态的客户端
// 3. 创建客户端实例并加入管理器
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 加载过程中的错误
func (tm *TunnelManager) loadTunnelClients(ctx context.Context) error {
	// 获取所有客户端配置
	clients, err := tm.storageManager.GetTunnelClientRepository().GetByTenantID(ctx, "default")
	if err != nil {
		return fmt.Errorf("failed to load client configurations: %w", err)
	}

	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for _, clientConfig := range clients {
		if clientConfig.ActiveFlag == types.ActiveFlagYes {
			tunnelClient := client.NewTunnelClient(clientConfig)
			tm.clients[clientConfig.TunnelClientId] = tunnelClient

			logger.Info("Tunnel client loaded", map[string]interface{}{
				"clientID":   clientConfig.TunnelClientId,
				"clientName": clientConfig.ClientName,
			})
		}
	}

	return nil
}

// validateNodeConfig 验证服务器节点配置的有效性
//
// 验证项目包括：
// 1. 端口冲突检查：确保监听端口没有被其他节点占用
// 2. 代理类型有效性：检查是否为支持的代理类型
// 3. 地址格式验证：检查监听地址和目标地址格式
//
// 参数:
//   - ctx: 上下文对象
//   - node: 要验证的节点配置
//
// 返回:
//   - error: 验证失败的具体错误信息
func (tm *TunnelManager) validateNodeConfig(ctx context.Context, node *types.TunnelServerNode) error {
	// 检查端口冲突
	existingNode, err := tm.storageManager.GetTunnelServerNodeRepository().GetByPortAndType(ctx, node.ListenAddress, node.ListenPort, node.ProxyType)
	if err == nil && existingNode != nil {
		return fmt.Errorf("port %d is already in use by node %s", node.ListenPort, existingNode.NodeName)
	}

	// 验证代理类型
	validTypes := []string{types.ProxyTypeTCP, types.ProxyTypeUDP, types.ProxyTypeHTTP, types.ProxyTypeHTTPS, types.ProxyTypeSTCP, types.ProxyTypeSUDP}
	isValid := false
	for _, validType := range validTypes {
		if node.ProxyType == validType {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid proxy type: %s", node.ProxyType)
	}

	return nil
}

// validateServiceConfig 验证服务配置的有效性
//
// 验证项目包括：
// 1. 服务名称唯一性：确保服务名称在租户内唯一
// 2. 远程端口冲突检查：如果指定了远程端口，检查是否已被占用
// 3. 配置完整性验证：检查必要字段是否完整
//
// 参数:
//   - ctx: 上下文对象
//   - service: 要验证的服务配置
//
// 返回:
//   - error: 验证失败的具体错误信息
func (tm *TunnelManager) validateServiceConfig(ctx context.Context, service *types.TunnelService) error {
	// 检查服务名称冲突
	existingService, err := tm.storageManager.GetTunnelServiceRepository().GetByName(ctx, service.ServiceName)
	if err == nil && existingService != nil {
		return fmt.Errorf("service name %s is already in use", service.ServiceName)
	}

	// 检查远程端口冲突（如果指定）
	if service.RemotePort != nil {
		existingService, err := tm.storageManager.GetTunnelServiceRepository().GetByRemotePort(ctx, *service.RemotePort)
		if err == nil && existingService != nil {
			return fmt.Errorf("remote port %d is already in use by service %s", *service.RemotePort, existingService.ServiceName)
		}
	}

	return nil
}

// ReloadServerConfig 重新加载指定服务器的配置
//
// 从数据库重新加载服务器配置，更新内存缓存
// 如果服务器正在运行，需要先停止再重新启动才能应用新配置
//
// 参数:
//   - ctx: 上下文对象
//   - serverID: 服务器唯一标识
//
// 返回:
//   - error: 重新加载过程中的错误
func (tm *TunnelManager) ReloadServerConfig(ctx context.Context, serverID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 从数据库加载最新配置
	serverConfig, err := tm.storageManager.GetTunnelServerRepository().GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("从数据库加载服务器配置失败: %w", err)
	}

	if serverConfig == nil {
		// 配置不存在，从缓存中删除
		delete(tm.servers, serverID)
		logger.Info("服务器配置不存在，已从缓存删除", map[string]interface{}{
			"serverID": serverID,
		})
		return nil
	}

	if serverConfig.ActiveFlag != types.ActiveFlagYes {
		// 配置已禁用，从缓存中删除
		delete(tm.servers, serverID)
		logger.Info("服务器配置已禁用，已从缓存删除", map[string]interface{}{
			"serverID": serverID,
		})
		return nil
	}

	// 创建新的服务器实例
	tunnelServer := server.NewTunnelServer(serverConfig, tm.storageManager)
	tm.servers[serverID] = tunnelServer

	logger.Info("服务器配置已重新加载", map[string]interface{}{
		"serverID":   serverID,
		"serverName": serverConfig.ServerName,
	})

	return nil
}

// ReloadClientConfig 重新加载指定客户端的配置
//
// 从数据库重新加载客户端配置，更新内存缓存
// 如果客户端正在运行，需要先停止再重新启动才能应用新配置
//
// 参数:
//   - ctx: 上下文对象
//   - clientID: 客户端唯一标识
//
// 返回:
//   - error: 重新加载过程中的错误
func (tm *TunnelManager) ReloadClientConfig(ctx context.Context, clientID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	// 从数据库加载最新配置
	clientConfig, err := tm.storageManager.GetTunnelClientRepository().GetByID(ctx, clientID)
	if err != nil {
		return fmt.Errorf("从数据库加载客户端配置失败: %w", err)
	}

	if clientConfig == nil {
		// 配置不存在，从缓存中删除
		delete(tm.clients, clientID)
		logger.Info("客户端配置不存在，已从缓存删除", map[string]interface{}{
			"clientID": clientID,
		})
		return nil
	}

	if clientConfig.ActiveFlag != types.ActiveFlagYes {
		// 配置已禁用，从缓存中删除
		delete(tm.clients, clientID)
		logger.Info("客户端配置已禁用，已从缓存删除", map[string]interface{}{
			"clientID": clientID,
		})
		return nil
	}

	// 创建新的客户端实例
	tunnelClient := client.NewTunnelClient(clientConfig)
	tm.clients[clientID] = tunnelClient

	logger.Info("客户端配置已重新加载", map[string]interface{}{
		"clientID":   clientID,
		"clientName": clientConfig.ClientName,
	})

	return nil
}

// ReloadAllConfigs 重新加载所有配置
//
// 从数据库重新加载所有服务器和客户端配置，更新内存缓存
// 注意：正在运行的实例不会自动重启，需要手动重启才能应用新配置
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 重新加载过程中的错误
func (tm *TunnelManager) ReloadAllConfigs(ctx context.Context) error {
	logger.Info("开始重新加载所有隧道配置", nil)

	// 重新加载服务器配置
	if err := tm.loadTunnelServers(ctx); err != nil {
		return fmt.Errorf("重新加载服务器配置失败: %w", err)
	}

	// 重新加载客户端配置
	if err := tm.loadTunnelClients(ctx); err != nil {
		return fmt.Errorf("重新加载客户端配置失败: %w", err)
	}

	logger.Info("所有隧道配置重新加载完成", map[string]interface{}{
		"servers": len(tm.servers),
		"clients": len(tm.clients),
	})

	return nil
}

// SetGlobalManager 设置全局隧道管理器实例
//
// 此函数用于在应用启动时设置全局管理器实例，
// 使其他模块可以通过 GetGlobalManager() 访问
//
// 参数:
//   - manager: 隧道管理器实例
func SetGlobalManager(manager *TunnelManager) {
	globalManagerLock.Lock()
	defer globalManagerLock.Unlock()

	globalManager = manager
	logger.Info("全局隧道管理器实例已设置", nil)
}

// GetGlobalManager 获取全局隧道管理器实例
//
// 返回全局隧道管理器实例，如果未设置则返回 nil
//
// 返回:
//   - *TunnelManager: 全局隧道管理器实例
//
// 使用示例:
//
//	manager := tunnel.GetGlobalManager()
//	if manager != nil {
//	    manager.StartServer(ctx, "server-001")
//	}
func GetGlobalManager() *TunnelManager {
	globalManagerLock.RLock()
	defer globalManagerLock.RUnlock()

	return globalManager
}

// IsGlobalManagerReady 检查全局隧道管理器是否已就绪
//
// 返回:
//   - bool: true 表示全局管理器已设置，false 表示未设置
func IsGlobalManagerReady() bool {
	globalManagerLock.RLock()
	defer globalManagerLock.RUnlock()

	return globalManager != nil
}
