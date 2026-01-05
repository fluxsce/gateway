// Package client 提供隧道客户端的完整实现
// 基于FRP架构，实现客户端连接、服务注册和数据转发功能
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// tunnelClient 隧道客户端实现
// 实现 TunnelClient 接口，协调各个子组件的工作
type tunnelClient struct {
	config *types.TunnelClient // config.Services 直接维护服务列表

	// 存储接口（用于数据库操作）
	clientRepository  *storage.TunnelClientRepositoryImpl
	serviceRepository *storage.TunnelServiceRepositoryImpl

	// 服务管理互斥锁（保护 config.Services）
	servicesMutex sync.RWMutex

	// 子组件（简化）
	controlConn      ControlConnection // 包含代理处理逻辑
	heartbeatManager HeartbeatManager  // 包含重连逻辑

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
//   - clientRepository: 隧道客户端存储接口
//   - serviceRepository: 隧道服务存储接口
//
// 返回:
//   - TunnelClient: 隧道客户端接口实例
//
// 功能:
//   - 创建客户端实例并初始化各个子组件
//   - 设置初始状态和配置参数
//   - 建立组件间的协调关系
func NewTunnelClient(config *types.TunnelClient, clientRepository *storage.TunnelClientRepositoryImpl, serviceRepository *storage.TunnelServiceRepositoryImpl) TunnelClient {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化 config.Services（如果为 nil）
	if config.Services == nil {
		config.Services = make(map[string]*types.TunnelService)
	}

	client := &tunnelClient{
		config:            config,
		clientRepository:  clientRepository,
		serviceRepository: serviceRepository,
		ctx:               ctx,
		cancel:            cancel,
		running:           false,
	}

	// 初始化子组件（简化）
	client.controlConn = NewControlConnection(client)
	client.heartbeatManager = NewHeartbeatManager(client)

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

	logger.Info("Starting tunnel client", map[string]interface{}{
		"clientId":   c.config.TunnelClientId,
		"clientName": c.config.ClientName,
	})

	// 1. 先建立控制连接（这是基础，其他组件都依赖它）
	if err := c.controlConn.Connect(c.ctx, c.config.ServerAddress, c.config.ServerPort); err != nil {
		// 连接失败，更新状态为错误
		c.config.ConnectionStatus = types.ConnectionStatusError
		if c.clientRepository != nil {
			disconnectTime := time.Now()
			if updateErr := c.clientRepository.UpdateConnectionStatus(
				ctx,
				c.config.TunnelClientId,
				types.ConnectionStatusError,
				&disconnectTime,
			); updateErr != nil {
				logger.Error("Failed to update connection status to error in database", map[string]interface{}{
					"clientId": c.config.TunnelClientId,
					"error":    updateErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to connect to server: %w", err)
	}

	// 2. 启动心跳管理器（需要连接已建立，包含重连逻辑）
	heartbeatInterval := time.Duration(c.config.HeartbeatInterval) * time.Second
	if err := c.heartbeatManager.Start(c.ctx, heartbeatInterval); err != nil {
		// 心跳管理器启动失败，更新状态为断开
		c.config.ConnectionStatus = types.ConnectionStatusDisconnected
		if c.clientRepository != nil {
			disconnectTime := time.Now()
			if updateErr := c.clientRepository.UpdateConnectionStatus(
				ctx,
				c.config.TunnelClientId,
				types.ConnectionStatusDisconnected,
				&disconnectTime,
			); updateErr != nil {
				logger.Error("Failed to update connection status to disconnected in database", map[string]interface{}{
					"clientId": c.config.TunnelClientId,
					"error":    updateErr.Error(),
				})
			}
		}
		return fmt.Errorf("failed to start heartbeat manager: %w", err)
	}

	// 更新连接状态到 config
	connectTime := time.Now()
	c.config.ConnectionStatus = types.ConnectionStatusConnected
	c.config.LastConnectTime = &connectTime

	// 更新数据库连接状态
	if c.clientRepository != nil {
		if err := c.clientRepository.UpdateConnectionStatus(
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

	// 1. 先取消上下文，通知所有子组件停止
	c.cancel()

	// 2. 注销所有服务
	c.servicesMutex.RLock()
	services := make([]*types.TunnelService, 0, len(c.config.Services))
	for _, service := range c.config.Services {
		services = append(services, service)
	}
	c.servicesMutex.RUnlock()

	for _, service := range services {
		if err := c.UnregisterService(ctx, service.TunnelServiceId); err != nil {
			logger.Error("Failed to unregister service during shutdown", map[string]interface{}{
				"serviceId": service.TunnelServiceId,
				"error":     err.Error(),
			})
		}
	}

	// 3. 停止心跳管理器（包含重连逻辑）
	if err := c.heartbeatManager.Stop(ctx); err != nil {
		logger.Error("Failed to stop heartbeat manager", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 4. 断开控制连接
	if err := c.controlConn.Disconnect(ctx); err != nil {
		logger.Error("Failed to disconnect control connection", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// 5. 等待所有 goroutine 退出（带超时保护）
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	// 6. 更新连接状态到 config（无论是否超时都要更新）
	disconnectTime := time.Now()
	c.config.ConnectionStatus = types.ConnectionStatusDisconnected
	c.config.LastDisconnectTime = &disconnectTime

	select {
	case <-done:
		// 正常退出
		logger.Info("All goroutines stopped successfully", map[string]interface{}{
			"clientId": c.config.TunnelClientId,
		})
	case <-ctx.Done():
		// 超时，但仍然继续清理
		logger.Warn("Timeout waiting for goroutines to stop, forcing shutdown", map[string]interface{}{
			"clientId": c.config.TunnelClientId,
		})
	}

	// 7. 更新数据库连接状态（使用 context.Background 避免超时影响）
	if c.clientRepository != nil {
		dbCtx, dbCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer dbCancel()

		if err := c.clientRepository.UpdateConnectionStatus(
			dbCtx,
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
}

// IsConnected 检查是否已连接
func (c *tunnelClient) IsConnected() bool {
	return c.config.ConnectionStatus == types.ConnectionStatusConnected
}

// GetConnectTime 获取连接时间
func (c *tunnelClient) GetConnectTime() time.Time {
	if c.config.LastConnectTime != nil {
		return *c.config.LastConnectTime
	}
	return time.Time{}
}

// GetReconnectCount 获取重连次数
func (c *tunnelClient) GetReconnectCount() int {
	return c.config.ReconnectCount
}

// GetConfig 获取客户端配置
func (c *tunnelClient) GetConfig() *types.TunnelClient {
	return c.config
}

// RegisterService 注册服务
func (c *tunnelClient) RegisterService(ctx context.Context, service *types.TunnelService) error {
	if !c.IsConnected() {
		return fmt.Errorf("client is not connected")
	}

	// 验证服务ID
	if service.TunnelServiceId == "" {
		return fmt.Errorf("service ID is required")
	}

	// 设置客户端ID（确保一致性）
	service.TunnelClientId = c.config.TunnelClientId

	// 向服务端发送注册请求
	sessionID := random.GenerateUniqueStringWithPrefix("reg_", 32)
	registerMsg := types.NewRegisterServiceMessage(sessionID, service)

	// 等待服务端响应（超时30秒）
	response, err := c.controlConn.SendMessage(ctx, registerMsg, true, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to send register service message: %w", err)
	}

	// 解析响应
	var resp types.RegisterServiceResponse
	if err := parseResponseData(response.Data, &resp); err != nil {
		return fmt.Errorf("failed to parse register service response: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("service registration failed: %s", resp.Message)
	}

	// 如果服务端分配了远程端口，更新到服务配置
	remotePort := 0
	if resp.RemotePort != nil {
		service.RemotePort = resp.RemotePort
		remotePort = *resp.RemotePort
	} else if service.RemotePort != nil {
		// 如果响应中没有远程端口，使用服务配置中的端口
		remotePort = *service.RemotePort
	}

	// 添加到本地服务列表（直接使用 config.Services）
	c.servicesMutex.Lock()
	c.config.Services[service.TunnelServiceId] = service
	serviceCount := len(c.config.Services)
	c.servicesMutex.Unlock()

	// 关键修复：启动代理，将服务添加到 activeProxies 中
	// 这样当服务端发送 proxy_request 时，HandleProxyConnection 才能找到对应的代理
	if remotePort > 0 {
		if err := c.controlConn.StartProxy(ctx, service, remotePort); err != nil {
			logger.Error("Failed to start proxy after service registration", map[string]interface{}{
				"serviceId":   service.TunnelServiceId,
				"serviceName": service.ServiceName,
				"remotePort":  remotePort,
				"error":       err.Error(),
			})
			// 不中断流程，继续执行，但记录错误
		}
	} else {
		logger.Warn("Remote port not available, skipping proxy start", map[string]interface{}{
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
		})
	}

	// 更新数据库服务状态为 active
	now := time.Now()
	if c.serviceRepository != nil {
		if err := c.serviceRepository.UpdateStatus(ctx, service.TunnelServiceId, types.ServiceStatusActive, &now); err != nil {
			logger.Error("Failed to update service status to active in database", map[string]interface{}{
				"serviceId": service.TunnelServiceId,
				"error":     err.Error(),
			})
			// 不中断流程，继续执行
		}
	}

	// 更新数据库服务数量
	if c.clientRepository != nil {
		currentClient, err := c.clientRepository.GetByID(ctx, c.config.TunnelClientId)
		if err != nil {
			logger.Error("Failed to get client info from database", map[string]interface{}{
				"clientId": c.config.TunnelClientId,
				"error":    err.Error(),
			})
		} else if currentClient != nil {
			currentClient.ServiceCount = serviceCount
			if err := c.clientRepository.Update(ctx, currentClient); err != nil {
				logger.Error("Failed to update service count in database", map[string]interface{}{
					"clientId":     c.config.TunnelClientId,
					"serviceCount": serviceCount,
					"error":        err.Error(),
				})
			}
		}
	}

	logger.Info("Service registered successfully", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"serviceType": service.ServiceType,
		"localPort":   service.LocalPort,
		"remotePort":  service.RemotePort,
	})

	return nil
}

// UnregisterService 注销服务
func (c *tunnelClient) UnregisterService(ctx context.Context, serviceID string) error {
	// 从本地服务列表获取服务信息（直接使用 config.Services）
	c.servicesMutex.Lock()
	service, exists := c.config.Services[serviceID]
	if !exists {
		c.servicesMutex.Unlock()
		return fmt.Errorf("service %s not found", serviceID)
	}
	c.servicesMutex.Unlock()

	// 向服务端发送注销请求
	if c.IsConnected() {
		sessionID := random.GenerateUniqueStringWithPrefix("unreg_", 32)
		unregisterMsg := types.NewUnregisterServiceMessage(sessionID, serviceID, service.ServiceName)

		// 等待服务端响应（超时30秒）
		response, err := c.controlConn.SendMessage(ctx, unregisterMsg, true, 30*time.Second)
		if err != nil {
			logger.Error("Failed to send unregister service message", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
			// 继续执行本地注销，即使服务端通知失败
		} else {
			// 解析响应
			var resp types.CommonResponse
			if err := parseResponseData(response.Data, &resp); err != nil {
				logger.Error("Failed to parse unregister service response", map[string]interface{}{
					"serviceId": serviceID,
					"error":     err.Error(),
				})
			} else if !resp.Success {
				logger.Warn("Service unregistration failed on server", map[string]interface{}{
					"serviceId": serviceID,
					"message":   resp.Message,
				})
			}
		}
	}

	// 关键修复：停止代理，从 activeProxies 中移除服务
	// 这样当服务注销后，不会再处理该服务的 proxy_request
	if err := c.controlConn.StopProxy(ctx, serviceID); err != nil {
		logger.Warn("Failed to stop proxy during service unregistration", map[string]interface{}{
			"serviceId": serviceID,
			"error":     err.Error(),
		})
		// 不中断流程，继续执行
	}

	// 从本地服务列表删除
	c.servicesMutex.Lock()
	delete(c.config.Services, serviceID)
	serviceCount := len(c.config.Services)
	c.servicesMutex.Unlock()

	// 更新数据库服务状态为 inactive
	if c.serviceRepository != nil {
		if err := c.serviceRepository.UpdateStatus(ctx, serviceID, types.ServiceStatusInactive, nil); err != nil {
			logger.Error("Failed to update service status to inactive in database", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
			// 不中断流程，继续执行
		}
	}

	// 更新数据库服务数量
	if c.clientRepository != nil {
		currentClient, err := c.clientRepository.GetByID(ctx, c.config.TunnelClientId)
		if err != nil {
			logger.Error("Failed to get client info from database", map[string]interface{}{
				"clientId": c.config.TunnelClientId,
				"error":    err.Error(),
			})
		} else if currentClient != nil {
			currentClient.ServiceCount = serviceCount
			if err := c.clientRepository.Update(ctx, currentClient); err != nil {
				logger.Error("Failed to update service count in database", map[string]interface{}{
					"clientId":     c.config.TunnelClientId,
					"serviceCount": serviceCount,
					"error":        err.Error(),
				})
			}
		}
	}

	logger.Info("Service unregistered successfully", map[string]interface{}{
		"serviceId":   serviceID,
		"serviceName": service.ServiceName,
	})

	return nil
}

// GetRegisteredServices 获取已注册的服务
func (c *tunnelClient) GetRegisteredServices() []*types.TunnelService {
	c.servicesMutex.RLock()
	defer c.servicesMutex.RUnlock()

	services := make([]*types.TunnelService, 0, len(c.config.Services))
	for _, service := range c.config.Services {
		services = append(services, service)
	}
	return services
}

// getRegisteredService 获取已注册的服务（供内部使用）
func (c *tunnelClient) getRegisteredService(serviceID string) *types.TunnelService {
	c.servicesMutex.RLock()
	defer c.servicesMutex.RUnlock()
	return c.config.Services[serviceID]
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
	if c.serviceRepository == nil {
		logger.Debug("Service repository not available, skipping service loading", nil)
		return nil
	}

	// 从数据库查询该客户端的所有服务
	services, err := c.serviceRepository.GetByClientID(ctx, c.config.TunnelClientId)
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
			if err := c.UnregisterService(ctx, service.TunnelServiceId); err != nil {
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

// parseResponseData 解析响应数据到指定的结构体
// 功能说明: 统一的响应数据解析函数，用于将 map[string]interface{} 转换为具体的响应结构体
// 实现方式: 使用 JSON 序列化/反序列化来处理类型转换
// 使用场景: 所有响应处理都应该使用此函数来解析 response.Data
func parseResponseData(data map[string]interface{}, target interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal response data: %w", err)
	}

	if err := json.Unmarshal(dataJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal response data: %w", err)
	}

	return nil
}
