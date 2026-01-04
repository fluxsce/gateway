// Package server 实现隧道服务端管理器
// 负责管理隧道服务器的生命周期
// 提供启动、停止、重载等公用方法
package server

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// TunnelServerManager 隧道服务端管理器
// 负责管理隧道服务器的生命周期，处理控制端口和客户端注册连接
// 实现内网和外网的交互
type TunnelServerManager struct {
	servers    map[string]*DefaultTunnelServer     // 隧道服务器映射: serverID -> TunnelServer
	mutex      sync.RWMutex                        // 并发保护
	repository *storage.TunnelServerRepositoryImpl // 隧道服务器存储接口

	cancel context.CancelFunc // 用于 StopAll 时取消
	wg     sync.WaitGroup     // 等待所有 goroutine 完成
}

var (
	// 全局唯一的隧道服务端管理器实例（单例模式）
	globalTunnelManager *TunnelServerManager
	tunnelManagerMutex  sync.RWMutex
)

// GetTunnelServerManager 获取全局唯一的隧道服务端管理器实例（单例模式）
// 注意：globalTunnelManager 在 InitializeTunnelManager 时创建，在此之前返回 nil
//
// 返回:
//   - *TunnelServerManager: 隧道服务端管理器实例，如果未初始化则返回 nil
func GetTunnelServerManager() *TunnelServerManager {
	tunnelManagerMutex.RLock()
	defer tunnelManagerMutex.RUnlock()
	return globalTunnelManager
}

// InitializeTunnelManager 初始化隧道服务端管理器
// 创建全局唯一的 TunnelServerManager 实例并从数据库加载所有隧道服务器配置
//
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接接口
//
// 返回:
//   - *TunnelServerManager: 隧道服务端管理器实例
//   - error: 错误信息
func InitializeTunnelManager(ctx context.Context, db database.Database) (*TunnelServerManager, error) {
	tunnelManagerMutex.Lock()
	defer tunnelManagerMutex.Unlock()

	// 如果已经初始化，直接返回
	if globalTunnelManager != nil {
		return globalTunnelManager, nil
	}

	// 从父 context 派生取消函数（用于 StopAll）
	// 当父 context 取消时，管理器也会收到取消信号
	_, cancel := context.WithCancel(ctx)

	// 创建全局管理器实例
	globalTunnelManager = &TunnelServerManager{
		servers:    make(map[string]*DefaultTunnelServer),
		repository: storage.NewTunnelServerRepository(db),
		cancel:     cancel,
	}

	logger.Info("Tunnel server manager initialized", nil)

	// 从数据库加载所有隧道服务器配置
	if err := globalTunnelManager.loadAllServersFromDB(ctx); err != nil {
		logger.Error("Failed to load tunnel servers from database", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to load tunnel servers: %w", err)
	}

	return globalTunnelManager, nil
}

// loadAllServersFromDB 从数据库加载所有隧道服务器配置
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) loadAllServersFromDB(ctx context.Context) error {
	// 查询所有隧道服务器配置
	serverConfigs, err := m.repository.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to query tunnel servers: %w", err)
	}

	logger.Info("Loading tunnel servers from database", map[string]interface{}{
		"count": len(serverConfigs),
	})

	// 创建服务器实例
	for _, config := range serverConfigs {
		server := NewTunnelServer(config, m.repository)
		m.servers[config.TunnelServerId] = server

		logger.Info("Tunnel server loaded", map[string]interface{}{
			"serverID":     config.TunnelServerId,
			"serverName":   config.ServerName,
			"controlPort":  config.ControlPort,
			"serverStatus": config.ServerStatus,
		})
	}

	logger.Info("All tunnel servers loaded successfully", map[string]interface{}{
		"totalCount": len(m.servers),
	})

	return nil
}

// Start 启动指定隧道服务器
// 启动指定 serverID 的隧道服务器实例
// 如果服务器实例不存在，会从数据库重新加载并创建
//
// 参数:
//   - ctx: 上下文
//   - serverID: 隧道服务器ID
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) Start(ctx context.Context, serverID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, exists := m.servers[serverID]
	if !exists {
		// 服务器实例不存在，从数据库重新加载
		logger.Info("Tunnel server not found in cache, loading from database", map[string]interface{}{
			"serverID": serverID,
		})

		config, err := m.repository.GetByID(ctx, serverID)
		if err != nil {
			return fmt.Errorf("failed to load tunnel server config: %w", err)
		}
		if config == nil {
			return fmt.Errorf("tunnel server not found in database: %s", serverID)
		}

		// 创建服务器实例
		server = NewTunnelServer(config, m.repository)
		m.servers[serverID] = server

		logger.Info("Tunnel server loaded from database", map[string]interface{}{
			"serverID":    serverID,
			"serverName":  config.ServerName,
			"controlPort": config.ControlPort,
		})
	}

	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("failed to start tunnel server %s: %w", serverID, err)
	}

	logger.Info("Tunnel server started", map[string]interface{}{
		"serverID":    serverID,
		"serverName":  server.config.ServerName,
		"controlPort": server.config.ControlPort,
	})

	return nil
}

// StartAll 启动所有隧道服务器
// 启动所有已加载的隧道服务器实例
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) StartAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Starting all tunnel servers", map[string]interface{}{
		"count": len(m.servers),
	})

	// 启动所有隧道服务器
	for serverID, server := range m.servers {
		if err := server.Start(ctx); err != nil {
			logger.Error("Failed to start tunnel server", map[string]interface{}{
				"serverID": serverID,
				"error":    err.Error(),
			})
			// 继续启动其他服务器，不中断
		}
	}

	logger.Info("All tunnel servers started", map[string]interface{}{
		"count": len(m.servers),
	})

	return nil
}

// Stop 停止指定隧道服务器
// 停止指定 serverID 的隧道服务器实例并从缓存中移除
//
// 参数:
//   - ctx: 上下文
//   - serverID: 隧道服务器ID
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 停止后会从缓存中移除，下次启动时会从数据库重新加载
func (m *TunnelServerManager) Stop(ctx context.Context, serverID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, exists := m.servers[serverID]
	if !exists {
		return fmt.Errorf("tunnel server not found: %s", serverID)
	}

	if err := server.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop tunnel server %s: %w", serverID, err)
	}

	// 从缓存中移除已停止的服务器
	delete(m.servers, serverID)

	logger.Info("Tunnel server stopped and removed from cache", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// StopAll 停止所有隧道服务器
// 停止所有正在运行的隧道服务器实例
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) StopAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Stopping all tunnel servers", map[string]interface{}{
		"count": len(m.servers),
	})

	// 取消上下文
	if m.cancel != nil {
		m.cancel()
	}

	// 停止所有隧道服务器
	for serverID, server := range m.servers {
		if err := server.Stop(ctx); err != nil {
			logger.Error("Failed to stop tunnel server", map[string]interface{}{
				"serverID": serverID,
				"error":    err.Error(),
			})
		}
	}

	// 等待所有 goroutine 完成
	m.wg.Wait()

	logger.Info("All tunnel servers stopped", map[string]interface{}{
		"count": len(m.servers),
	})

	return nil
}

// Reload 重载指定隧道服务器配置
// 更新服务器配置并重启
//
// 参数:
//   - ctx: 上下文
//   - config: 新的服务器配置
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) Reload(ctx context.Context, config *types.TunnelServer) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	serverID := config.TunnelServerId

	// 检查是否存在运行中的服务器
	existingServer, exists := m.servers[serverID]
	wasRunning := false
	if exists {
		wasRunning = existingServer.IsRunning()
		// 停止现有服务器
		if err := existingServer.Stop(ctx); err != nil {
			logger.Error("Failed to stop existing server for reload", map[string]interface{}{
				"serverID": serverID,
				"error":    err.Error(),
			})
		}
	}

	// 创建新的服务器实例
	newServer := NewTunnelServer(config, m.repository)
	m.servers[serverID] = newServer

	// 如果之前是运行状态，重新启动
	if wasRunning {
		if err := newServer.Start(ctx); err != nil {
			return fmt.Errorf("failed to restart server after reload: %w", err)
		}
	}

	logger.Info("Tunnel server reloaded", map[string]interface{}{
		"serverID":   serverID,
		"serverName": config.ServerName,
	})

	return nil
}

// Delete 从管理器中删除服务器
// 会先停止服务器再删除
//
// 参数:
//   - ctx: 上下文
//   - serverID: 隧道服务器ID
//
// 返回:
//   - error: 错误信息
func (m *TunnelServerManager) Delete(ctx context.Context, serverID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	server, exists := m.servers[serverID]
	if !exists {
		return nil // 不存在视为删除成功
	}

	// 停止服务器
	if err := server.Stop(ctx); err != nil {
		logger.Error("Failed to stop server before delete", map[string]interface{}{
			"serverID": serverID,
			"error":    err.Error(),
		})
	}

	delete(m.servers, serverID)

	logger.Info("Tunnel server deleted", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// GetServer 获取指定隧道服务器实例
//
// 参数:
//   - serverID: 隧道服务器ID
//
// 返回:
//   - *DefaultTunnelServer: 隧道服务器实例，如果不存在则返回 nil
func (m *TunnelServerManager) GetServer(serverID string) *DefaultTunnelServer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.servers[serverID]
}

// GetAllServers 获取所有隧道服务器实例
//
// 返回:
//   - []*DefaultTunnelServer: 所有隧道服务器实例列表
func (m *TunnelServerManager) GetAllServers() []*DefaultTunnelServer {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	servers := make([]*DefaultTunnelServer, 0, len(m.servers))
	for _, server := range m.servers {
		servers = append(servers, server)
	}
	return servers
}

// GetServerStatus 获取指定服务器状态
//
// 参数:
//   - serverID: 隧道服务器ID
//
// 返回:
//   - *ServerStatus: 服务器状态，如果不存在则返回 nil
func (m *TunnelServerManager) GetServerStatus(serverID string) *ServerStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[serverID]
	if !exists {
		return nil
	}

	status := server.GetStatus()
	return &status
}

// GetAllServerStatuses 获取所有服务器状态
//
// 返回:
//   - map[string]ServerStatus: 服务器ID到状态的映射
func (m *TunnelServerManager) GetAllServerStatuses() map[string]ServerStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	statuses := make(map[string]ServerStatus)
	for serverID, server := range m.servers {
		statuses[serverID] = server.GetStatus()
	}
	return statuses
}

// GetServerCount 获取服务器数量
//
// 返回:
//   - int: 服务器数量
func (m *TunnelServerManager) GetServerCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.servers)
}

// GetRegisteredClients 获取指定服务器上已注册的客户端列表
//
// 参数:
//   - serverID: 隧道服务器ID
//
// 返回:
//   - []*types.TunnelClient: 已注册客户端列表，如果服务器不存在则返回 nil
func (m *TunnelServerManager) GetRegisteredClients(serverID string) []*types.TunnelClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	server, exists := m.servers[serverID]
	if !exists {
		return nil
	}

	return server.GetConnectedClients()
}

// GetRegisteredServices 获取指定服务器上已注册的服务列表
// 如果 serverID 为空，则获取所有服务器的服务列表
//
// 参数:
//   - serverID: 隧道服务器ID（可选，为空时获取所有服务器的服务）
//
// 返回:
//   - []*types.TunnelService: 已注册服务列表
func (m *TunnelServerManager) GetRegisteredServices(serverID string) []*types.TunnelService {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var allServices []*types.TunnelService

	if serverID == "" {
		// 获取所有服务器的服务列表
		for _, server := range m.servers {
			if server == nil {
				continue
			}
			clients := server.GetConnectedClients()
			for _, client := range clients {
				if client != nil && client.Services != nil {
					for _, service := range client.Services {
						if service != nil {
							allServices = append(allServices, service)
						}
					}
				}
			}
		}
	} else {
		// 获取指定服务器的服务列表
		server, exists := m.servers[serverID]
		if !exists {
			return nil
		}
		clients := server.GetConnectedClients()
		for _, client := range clients {
			if client != nil && client.Services != nil {
				for _, service := range client.Services {
					if service != nil {
						allServices = append(allServices, service)
					}
				}
			}
		}
	}

	return allServices
}
