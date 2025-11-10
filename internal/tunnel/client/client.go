// Package client 提供隧道客户端的完整实现
// 基于FRP架构，实现客户端连接、服务注册和数据转发功能
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// tunnelClient 隧道客户端实现
// 实现 TunnelClient 接口，协调各个子组件的工作
type tunnelClient struct {
	config      *types.TunnelClient
	status      *ClientStatus
	statusMutex sync.RWMutex

	// 存储管理器（用于数据库操作）
	storageManager storage.RepositoryManager

	// 子组件
	controlConn      ControlConnection
	serviceManager   ServiceManager
	proxyManager     ProxyManager
	heartbeatManager HeartbeatManager
	reconnectManager ReconnectManager

	// 控制状态
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	running      bool
	runningMutex sync.RWMutex
}

// NewTunnelClient 创建隧道客户端实例
//
// 参数:
//   - config: 客户端配置对象
//   - storageManager: 存储管理器，用于数据库操作
//
// 返回:
//   - TunnelClient: 隧道客户端接口实例
//
// 功能:
//   - 创建客户端实例并初始化各个子组件
//   - 设置初始状态和配置参数
//   - 建立组件间的协调关系
func NewTunnelClient(config *types.TunnelClient, storageManager storage.RepositoryManager) TunnelClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &tunnelClient{
		config:         config,
		storageManager: storageManager,
		ctx:            ctx,
		cancel:         cancel,
		running:        false,
		status: &ClientStatus{
			Status:             StatusDisconnected,
			ServerAddress:      config.ServerAddress,
			ServerPort:         config.ServerPort,
			Connected:          false,
			ReconnectCount:     0,
			RegisteredServices: 0,
			ActiveProxies:      0,
			TotalTraffic:       0,
			Errors:             []string{},
		},
	}

	// 初始化子组件
	client.controlConn = NewControlConnection(client)
	client.serviceManager = NewServiceManager(client)
	client.proxyManager = NewProxyManager(client)
	client.heartbeatManager = NewHeartbeatManager(client)
	client.reconnectManager = NewReconnectManager(client)

	logger.Info("Tunnel client created", map[string]interface{}{
		"clientId":      config.TunnelClientId,
		"clientName":    config.ClientName,
		"serverAddress": config.ServerAddress,
		"serverPort":    config.ServerPort,
	})

	return client
}

// Start 启动客户端
func (c *tunnelClient) Start(ctx context.Context) error {
	c.runningMutex.Lock()
	if c.running {
		c.runningMutex.Unlock()
		return fmt.Errorf("client is already running")
	}
	c.runningMutex.Unlock()

	c.updateStatus(StatusConnecting, false)

	logger.Info("Starting tunnel client", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	// 启动重连管理器
	if err := c.reconnectManager.Start(c.ctx); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to start reconnect manager: %v", err))
		return fmt.Errorf("failed to start reconnect manager: %w", err)
	}

	// 建立控制连接
	if err := c.controlConn.Connect(c.ctx, c.config.ServerAddress, c.config.ServerPort); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to connect to server: %v", err))
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// 启动心跳管理器
	heartbeatInterval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if err := c.heartbeatManager.Start(c.ctx, heartbeatInterval); err != nil {
		c.updateStatus(StatusError, false)
		c.addError(fmt.Sprintf("Failed to start heartbeat manager: %v", err))
		return fmt.Errorf("failed to start heartbeat manager: %w", err)
	}

	c.updateStatus(StatusConnected, true)
	c.updateConnectTime()

	// 更新数据库连接状态
	if c.storageManager != nil {
		connectTime := time.Now()
		if err := c.storageManager.GetTunnelClientRepository().UpdateConnectionStatus(
			ctx,
			c.config.TunnelClientId,
			types.ConnectionStatusConnected,
			&connectTime,
		); err != nil {
			logger.Error("Failed to update connection status in database", map[string]interface{}{
				"clientId": c.config.TunnelClientId,
				"error":    err.Error(),
			})
		}
	}

	// 自动加载并注册数据库中已有的服务（应用重启恢复）
	if err := c.loadAndRegisterServices(ctx); err != nil {
		logger.Error("Failed to load and register services from database", map[string]interface{}{
			"clientId": c.config.TunnelClientId,
			"error":    err.Error(),
		})
		// 不中断启动流程，只记录错误
	}

	// 启动成功后才更新 running 状态
	c.runningMutex.Lock()
	c.running = true
	c.runningMutex.Unlock()

	logger.Info("Tunnel client started successfully", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	return nil
}

// Stop 停止客户端
func (c *tunnelClient) Stop(ctx context.Context) error {
	c.runningMutex.Lock()
	if !c.running {
		c.runningMutex.Unlock()
		return nil
	}
	c.running = false
	c.runningMutex.Unlock()

	logger.Info("Stopping tunnel client", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	// 注销所有服务
	services, err := c.serviceManager.GetAllServices(ctx)
	if err != nil {
		logger.Error("Failed to get services for shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	} else {
		for _, service := range services {
			if err := c.UnregisterService(ctx, service.TunnelServiceId); err != nil {
				logger.Error("Failed to unregister service during shutdown", map[string]interface{}{
					"serviceId": service.TunnelServiceId,
					"error":     err.Error(),
				})
			}
		}
	}

	// 停止心跳管理器
	if err := c.heartbeatManager.Stop(ctx); err != nil {
		logger.Error("Failed to stop heartbeat manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 停止重连管理器
	if err := c.reconnectManager.Stop(ctx); err != nil {
		logger.Error("Failed to stop reconnect manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 断开控制连接
	if err := c.controlConn.Disconnect(ctx); err != nil {
		logger.Error("Failed to disconnect control connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 取消上下文并等待协程退出
	c.cancel()

	// 等待消息循环退出
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		c.updateStatus(StatusStopped, false)

		// 更新数据库连接状态
		if c.storageManager != nil {
			disconnectTime := time.Now()
			if err := c.storageManager.GetTunnelClientRepository().UpdateConnectionStatus(
				ctx,
				c.config.TunnelClientId,
				types.ConnectionStatusDisconnected,
				&disconnectTime,
			); err != nil {
				logger.Error("Failed to update connection status in database", map[string]interface{}{
					"clientId": c.config.TunnelClientId,
					"error":    err.Error(),
				})
			}
		}

		logger.Info("Tunnel client stopped successfully", map[string]interface{}{
			"clientId":   c.config.TunnelClientId,
			"clientName": c.config.ClientName,
		})
		return nil
	case <-ctx.Done():
		c.updateStatus(StatusError, false)
		c.addError("Client stop timeout")
		return fmt.Errorf("client stop timeout")
	}
}

// GetStatus 获取客户端状态
func (c *tunnelClient) GetStatus() *ClientStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()

	// 创建状态副本
	status := &ClientStatus{
		Status:             c.status.Status,
		ServerAddress:      c.status.ServerAddress,
		ServerPort:         c.status.ServerPort,
		Connected:          c.status.Connected,
		LastConnectTime:    c.status.LastConnectTime,
		ConnectionDuration: c.status.ConnectionDuration,
		ReconnectCount:     c.status.ReconnectCount,
		RegisteredServices: c.status.RegisteredServices,
		ActiveProxies:      c.status.ActiveProxies,
		TotalTraffic:       c.status.TotalTraffic,
		LastHeartbeat:      c.status.LastHeartbeat,
		Errors:             make([]string, len(c.status.Errors)),
	}
	copy(status.Errors, c.status.Errors)

	// 更新连接持续时间
	if status.Connected && !status.LastConnectTime.IsZero() {
		status.ConnectionDuration = time.Since(status.LastConnectTime).Milliseconds()
	}

	// 更新心跳时间
	status.LastHeartbeat = c.heartbeatManager.GetLastHeartbeatTime()

	// 更新代理数量
	status.ActiveProxies = len(c.proxyManager.GetActiveProxies())

	return status
}

// GetConfig 获取客户端配置
func (c *tunnelClient) GetConfig() *types.TunnelClient {
	return c.config
}

// RegisterService 注册服务
func (c *tunnelClient) RegisterService(ctx context.Context, service *types.TunnelService) error {
	if !c.isConnected() {
		return fmt.Errorf("client is not connected")
	}

	// 验证服务ID
	if service.TunnelServiceId == "" {
		return fmt.Errorf("service ID is required")
	}

	// 验证服务配置
	if err := c.serviceManager.ValidateService(ctx, service); err != nil {
		return fmt.Errorf("service validation failed: %w", err)
	}

	// 设置客户端ID（确保一致性）
	service.TunnelClientId = c.config.TunnelClientId

	// 注册服务
	if err := c.serviceManager.RegisterService(ctx, service); err != nil {
		return fmt.Errorf("failed to register service: %w", err)
	}

	// 更新状态
	c.statusMutex.Lock()
	c.status.RegisteredServices++
	serviceCount := c.status.RegisteredServices
	c.statusMutex.Unlock()

	// 更新数据库服务数量
	if c.storageManager != nil {
		// 获取当前客户端信息
		currentClient, err := c.storageManager.GetTunnelClientRepository().GetByID(ctx, c.config.TunnelClientId)
		if err != nil {
			logger.Error("Failed to get client info from database", map[string]interface{}{
				"clientId": c.config.TunnelClientId,
				"error":    err.Error(),
			})
		} else if currentClient != nil {
			currentClient.ServiceCount = serviceCount
			if err := c.storageManager.GetTunnelClientRepository().Update(ctx, currentClient); err != nil {
				logger.Error("Failed to update service count in database", map[string]interface{}{
					"clientId":     c.config.TunnelClientId,
					"serviceCount": serviceCount,
					"error":        err.Error(),
				})
			}
		}
	}

	logger.Info("Service registered", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
		"localPort":   service.LocalPort,
	})

	return nil
}

// UnregisterService 注销服务
func (c *tunnelClient) UnregisterService(ctx context.Context, serviceID string) error {
	// 从 serviceManager 获取服务信息
	service, err := c.serviceManager.GetService(ctx, serviceID)
	if err != nil {
		return fmt.Errorf("service %s not found: %w", serviceID, err)
	}

	// 注销服务
	if err := c.serviceManager.UnregisterService(ctx, serviceID); err != nil {
		return fmt.Errorf("failed to unregister service: %w", err)
	}

	// 更新状态
	c.statusMutex.Lock()
	if c.status.RegisteredServices > 0 {
		c.status.RegisteredServices--
	}
	serviceCount := c.status.RegisteredServices
	c.statusMutex.Unlock()

	// 更新数据库服务数量
	if c.storageManager != nil {
		// 获取当前客户端信息
		currentClient, err := c.storageManager.GetTunnelClientRepository().GetByID(ctx, c.config.TunnelClientId)
		if err != nil {
			logger.Error("Failed to get client info from database", map[string]interface{}{
				"clientId": c.config.TunnelClientId,
				"error":    err.Error(),
			})
		} else if currentClient != nil {
			currentClient.ServiceCount = serviceCount
			if err := c.storageManager.GetTunnelClientRepository().Update(ctx, currentClient); err != nil {
				logger.Error("Failed to update service count in database", map[string]interface{}{
					"clientId":     c.config.TunnelClientId,
					"serviceCount": serviceCount,
					"error":        err.Error(),
				})
			}
		}
	}

	logger.Info("Service unregistered", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": service.ServiceName,
	})

	return nil
}

// GetRegisteredServices 获取已注册的服务
func (c *tunnelClient) GetRegisteredServices() []*types.TunnelService {
	services, err := c.serviceManager.GetAllServices(context.Background())
	if err != nil {
		logger.Error("Failed to get registered services", map[string]interface{}{
			"error": err.Error(),
		})
		return []*types.TunnelService{}
	}
	return services
}

// getRegisteredService 获取已注册的服务（供内部使用）
func (c *tunnelClient) getRegisteredService(serviceID string) *types.TunnelService {
	service, err := c.serviceManager.GetService(context.Background(), serviceID)
	if err != nil {
		return nil
	}
	return service
}

// updateStatus 更新客户端状态
func (c *tunnelClient) updateStatus(status string, connected bool) {
	c.statusMutex.Lock()
	c.status.Status = status
	c.status.Connected = connected
	c.statusMutex.Unlock()
}

// updateConnectTime 更新连接时间
func (c *tunnelClient) updateConnectTime() {
	c.statusMutex.Lock()
	c.status.LastConnectTime = time.Now()
	c.statusMutex.Unlock()
}

// addError 添加错误信息
func (c *tunnelClient) addError(errorMsg string) {
	c.statusMutex.Lock()
	c.status.Errors = append(c.status.Errors, errorMsg)
	// 限制错误数量，只保留最近的10个
	if len(c.status.Errors) > 10 {
		c.status.Errors = c.status.Errors[len(c.status.Errors)-10:]
	}
	c.statusMutex.Unlock()
}

// isConnected 检查是否已连接
func (c *tunnelClient) isConnected() bool {
	return c.controlConn.IsConnected()
}

// GetControlConnection 获取控制连接（供子组件使用）
func (c *tunnelClient) GetControlConnection() ControlConnection {
	return c.controlConn
}

// GetProxyManager 获取代理管理器（供子组件使用）
func (c *tunnelClient) GetProxyManager() ProxyManager {
	return c.proxyManager
}

// loadAndRegisterServices 从数据库加载并注册服务
//
// 在客户端启动时调用，用于恢复应用重启前已注册的服务。
// 这确保了服务注册的持久性，即使应用重启也能自动恢复。
//
// 参数:
//   - ctx: 上下文对象
//
// 返回:
//   - error: 加载或注册失败时的错误
//
// 工作流程:
//  1. 从数据库查询该客户端的所有活跃服务
//  2. 过滤出状态为 active 的服务
//  3. 逐个调用 RegisterService 重新注册到服务器
//  4. 记录成功和失败的服务数量
//
// 注意:
//   - 注册失败不会中断整个流程
//   - 只加载 activeFlag='Y' 的服务
//   - 服务注册失败会记录错误但继续处理下一个
func (c *tunnelClient) loadAndRegisterServices(ctx context.Context) error {
	if c.storageManager == nil {
		logger.Debug("Storage manager not available, skipping service loading", nil)
		return nil
	}

	// 从数据库查询该客户端的所有服务
	services, err := c.storageManager.GetTunnelServiceRepository().GetByClientID(ctx, c.config.TunnelClientId)
	if err != nil {
		return fmt.Errorf("failed to query services from database: %w", err)
	}

	if len(services) == 0 {
		logger.Info("No services found in database for this client", map[string]interface{}{
			"clientId": c.config.TunnelClientId,
		})
		return nil
	}

	logger.Info("Loading services from database", map[string]interface{}{
		"clientId":     c.config.TunnelClientId,
		"serviceCount": len(services),
	})

	// 统计注册结果
	successCount := 0
	failureCount := 0

	// 逐个注册服务
	for _, service := range services {
		// 跳过非活跃状态的服务
		if service.ActiveFlag != types.ActiveFlagYes {
			logger.Debug("Skipping inactive service", map[string]interface{}{
				"serviceId":   service.TunnelServiceId,
				"serviceName": service.ServiceName,
			})
			continue
		}

		// 检查服务是否已经在本地注册（可能是之前的残留）
		existingService := c.getRegisteredService(service.TunnelServiceId)
		if existingService != nil {
			logger.Info("Service already registered locally, unregistering before re-registration", map[string]interface{}{
				"serviceId":   service.TunnelServiceId,
				"serviceName": service.ServiceName,
			})

			// 先卸载已存在的服务
			if err := c.serviceManager.UnregisterService(ctx, service.TunnelServiceId); err != nil {
				logger.Error("Failed to unregister existing service", map[string]interface{}{
					"serviceId":   service.TunnelServiceId,
					"serviceName": service.ServiceName,
					"error":       err.Error(),
				})
				// 即使卸载失败，也尝试重新注册
			}
		}

		// 注册服务到服务器
		if err := c.RegisterService(ctx, service); err != nil {
			logger.Error("Failed to register service during startup", map[string]interface{}{
				"serviceId":   service.TunnelServiceId,
				"serviceName": service.ServiceName,
				"error":       err.Error(),
			})
			failureCount++
			// 继续处理下一个服务，不中断整个流程
			continue
		}

		logger.Info("Service registered successfully during startup", map[string]interface{}{
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
			"serviceType": service.ServiceType,
		})
		successCount++
	}

	logger.Info("Service loading completed", map[string]interface{}{
		"clientId":     c.config.TunnelClientId,
		"totalCount":   len(services),
		"successCount": successCount,
		"failureCount": failureCount,
	})

	// 如果所有服务都失败了，返回错误
	if failureCount > 0 && successCount == 0 {
		return fmt.Errorf("all %d services failed to register", failureCount)
	}

	return nil
}
