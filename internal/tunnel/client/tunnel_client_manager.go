// Package client 实现隧道客户端管理器
// 负责管理隧道客户端的生命周期
// 提供启动、停止、重载等公用方法
package client

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// TunnelClientManager 隧道客户端管理器
// 负责管理隧道客户端的生命周期，处理客户端连接和服务注册
type TunnelClientManager struct {
	clients           map[string]*tunnelClient             // 隧道客户端映射: clientID -> TunnelClient
	mutex             sync.RWMutex                         // 并发保护
	clientRepository  *storage.TunnelClientRepositoryImpl  // 隧道客户端存储接口
	serviceRepository *storage.TunnelServiceRepositoryImpl // 隧道服务存储接口

	// 管理器独立的上下文（不依赖外部传入的 context）
	// 使用 context.Background() 确保生命周期独立于 HTTP 请求
	ctx    context.Context
	cancel context.CancelFunc // 用于 StopAll 时取消
	wg     sync.WaitGroup     // 等待所有 goroutine 完成
}

var (
	// 全局唯一的隧道客户端管理器实例（单例模式）
	globalClientManager *TunnelClientManager
	clientManagerMutex  sync.RWMutex
)

// GetTunnelClientManager 获取全局唯一的隧道客户端管理器实例（单例模式）
// 注意：globalClientManager 在 InitializeClientManager 时创建，在此之前返回 nil
//
// 返回:
//   - *TunnelClientManager: 隧道客户端管理器实例，如果未初始化则返回 nil
func GetTunnelClientManager() *TunnelClientManager {
	clientManagerMutex.RLock()
	defer clientManagerMutex.RUnlock()
	return globalClientManager
}

// InitializeClientManager 初始化隧道客户端管理器
// 创建全局唯一的 TunnelClientManager 实例并从数据库加载所有隧道客户端配置
//
// 参数:
//   - ctx: 上下文，仅用于初始化时的数据库查询，不影响管理器生命周期
//   - db: 数据库连接接口
//
// 返回:
//   - *TunnelClientManager: 隧道客户端管理器实例
//   - error: 错误信息
//
// 重要说明:
//   - 管理器使用 context.Background 创建独立的上下文
//   - 确保管理器的生命周期不受外部 context 如 HTTP 请求影响
//   - 传入的 ctx 仅用于初始化时的数据库查询
func InitializeClientManager(ctx context.Context, db database.Database) (*TunnelClientManager, error) {
	clientManagerMutex.Lock()
	defer clientManagerMutex.Unlock()

	// 如果已经初始化，直接返回
	if globalClientManager != nil {
		return globalClientManager, nil
	}

	// 使用 context.Background 创建独立的上下文
	// 确保管理器的生命周期不受外部 context 如 HTTP 请求影响
	managerCtx, cancel := context.WithCancel(context.Background())

	// 创建全局管理器实例
	globalClientManager = &TunnelClientManager{
		clients:           make(map[string]*tunnelClient),
		clientRepository:  storage.NewTunnelClientRepository(db),
		serviceRepository: storage.NewTunnelServiceRepository(db),
		ctx:               managerCtx,
		cancel:            cancel,
	}

	logger.Info("Tunnel client manager initialized", nil)

	// 从数据库加载所有隧道客户端配置
	// 使用传入的 ctx 进行数据库查询（初始化操作）
	if err := globalClientManager.loadAllClientsFromDB(ctx); err != nil {
		logger.Error("Failed to load tunnel clients from database", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("failed to load tunnel clients: %w", err)
	}

	return globalClientManager, nil
}

// loadAllClientsFromDB 从数据库加载所有隧道客户端配置
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *TunnelClientManager) loadAllClientsFromDB(ctx context.Context) error {
	// 查询所有隧道客户端配置
	clientConfigs, err := m.clientRepository.GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to query tunnel clients: %w", err)
	}

	logger.Info("Loading tunnel clients from database", map[string]interface{}{
		"count": len(clientConfigs),
	})

	// 创建客户端实例
	for _, config := range clientConfigs {
		client := NewTunnelClient(config, m.clientRepository, m.serviceRepository)
		m.clients[config.TunnelClientId] = client.(*tunnelClient)

		logger.Info("Tunnel client loaded", map[string]interface{}{
			"clientID":         config.TunnelClientId,
			"clientName":       config.ClientName,
			"serverAddress":    config.ServerAddress,
			"connectionStatus": config.ConnectionStatus,
		})
	}

	logger.Info("All tunnel clients loaded successfully", map[string]interface{}{
		"totalCount": len(m.clients),
	})

	return nil
}

// Start 启动指定隧道客户端
// 启动指定 clientID 的隧道客户端实例
// 如果客户端实例不存在，会从数据库重新加载并创建
//
// 参数:
//   - ctx: 上下文，仅用于数据库查询，不影响客户端生命周期
//   - clientID: 隧道客户端ID
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 传入的 ctx 仅用于数据库查询操作
//   - 客户端使用管理器的独立 context m.ctx 运行
//   - 确保客户端不会因为 HTTP 请求超时而被终止
func (m *TunnelClientManager) Start(ctx context.Context, clientID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, exists := m.clients[clientID]
	if !exists {
		// 客户端实例不存在，从数据库重新加载
		logger.Info("Tunnel client not found in cache, loading from database", map[string]interface{}{
			"clientID": clientID,
		})

		// 使用传入的 ctx 进行数据库查询，短期操作
		config, err := m.clientRepository.GetByID(ctx, clientID)
		if err != nil {
			return fmt.Errorf("failed to load tunnel client config: %w", err)
		}
		if config == nil {
			return fmt.Errorf("tunnel client not found in database: %s", clientID)
		}

		// 创建客户端实例
		client = NewTunnelClient(config, m.clientRepository, m.serviceRepository).(*tunnelClient)
		m.clients[clientID] = client

		logger.Info("Tunnel client loaded from database", map[string]interface{}{
			"clientID":      clientID,
			"clientName":    config.ClientName,
			"serverAddress": config.ServerAddress,
		})
	}

	// 使用管理器的独立 context 启动客户端，长期运行
	// 确保客户端不会因为 HTTP 请求超时而被终止
	if err := client.Start(m.ctx); err != nil {
		return fmt.Errorf("failed to start tunnel client %s: %w", clientID, err)
	}

	logger.Info("Tunnel client started", map[string]interface{}{
		"clientID":      clientID,
		"clientName":    client.config.ClientName,
		"serverAddress": client.config.ServerAddress,
	})

	return nil
}

// StartAll 启动所有隧道客户端
// 启动所有已加载的隧道客户端实例
//
// 参数:
//   - ctx: 上下文，仅用于数据库操作，不影响客户端生命周期
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 客户端使用管理器的独立 context m.ctx 运行
//   - 确保客户端不会因为调用方的 context 取消而被终止
func (m *TunnelClientManager) StartAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Starting all tunnel clients", map[string]interface{}{
		"count": len(m.clients),
	})

	// 使用管理器的独立 context 启动所有客户端
	for clientID, client := range m.clients {
		if err := client.Start(m.ctx); err != nil {
			logger.Error("Failed to start tunnel client", map[string]interface{}{
				"clientID": clientID,
				"error":    err.Error(),
			})
			// 继续启动其他客户端，不中断
		}
	}

	logger.Info("All tunnel clients started", map[string]interface{}{
		"count": len(m.clients),
	})

	return nil
}

// Stop 停止指定隧道客户端
// 停止指定 clientID 的隧道客户端实例并从缓存中移除
//
// 参数:
//   - ctx: 上下文，用于数据库更新操作，设置合理的超时时间
//   - clientID: 隧道客户端ID
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 传入的 ctx 用于数据库更新操作如更新连接状态
//   - 建议传入带超时的 context 避免停止操作阻塞过久
//   - 停止后会从缓存中移除，下次启动时会从数据库重新加载
func (m *TunnelClientManager) Stop(ctx context.Context, clientID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, exists := m.clients[clientID]
	if !exists {
		return fmt.Errorf("tunnel client not found: %s", clientID)
	}

	// 使用传入的 ctx 进行停止操作，包括数据库更新
	if err := client.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop tunnel client %s: %w", clientID, err)
	}

	// 从缓存中移除已停止的客户端
	delete(m.clients, clientID)

	logger.Info("Tunnel client stopped and removed from cache", map[string]interface{}{
		"clientID": clientID,
	})

	return nil
}

// StopAll 停止所有隧道客户端
// 停止所有正在运行的隧道客户端实例
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *TunnelClientManager) StopAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Stopping all tunnel clients", map[string]interface{}{
		"count": len(m.clients),
	})

	// 取消上下文
	if m.cancel != nil {
		m.cancel()
	}

	// 停止所有隧道客户端
	for clientID, client := range m.clients {
		if err := client.Stop(ctx); err != nil {
			logger.Error("Failed to stop tunnel client", map[string]interface{}{
				"clientID": clientID,
				"error":    err.Error(),
			})
		}
	}

	// 等待所有 goroutine 完成
	m.wg.Wait()

	logger.Info("All tunnel clients stopped", map[string]interface{}{
		"count": len(m.clients),
	})

	return nil
}

// Reload 重载指定隧道客户端配置
// 更新客户端配置并重启
//
// 参数:
//   - ctx: 上下文，用于数据库操作，设置合理的超时时间
//   - config: 新的客户端配置
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 传入的 ctx 用于停止时的数据库更新
//   - 重新启动时使用管理器的独立 context m.ctx
func (m *TunnelClientManager) Reload(ctx context.Context, config *types.TunnelClient) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	clientID := config.TunnelClientId

	// 检查是否存在运行中的客户端
	existingClient, exists := m.clients[clientID]
	wasRunning := false
	if exists {
		wasRunning = existingClient.IsConnected()
		// 使用传入的 ctx 停止现有客户端，包括数据库更新
		if err := existingClient.Stop(ctx); err != nil {
			logger.Error("Failed to stop existing client for reload", map[string]interface{}{
				"clientID": clientID,
				"error":    err.Error(),
			})
		}
	}

	// 创建新的客户端实例
	newClient := NewTunnelClient(config, m.clientRepository, m.serviceRepository).(*tunnelClient)
	m.clients[clientID] = newClient

	// 如果之前是运行状态，重新启动
	if wasRunning {
		// 使用管理器的独立 context 启动客户端
		if err := newClient.Start(m.ctx); err != nil {
			return fmt.Errorf("failed to restart client after reload: %w", err)
		}
	}

	logger.Info("Tunnel client reloaded", map[string]interface{}{
		"clientID":   clientID,
		"clientName": config.ClientName,
	})

	return nil
}

// Delete 从管理器中删除客户端
// 会先停止客户端再删除
//
// 参数:
//   - ctx: 上下文
//   - clientID: 隧道客户端ID
//
// 返回:
//   - error: 错误信息
func (m *TunnelClientManager) Delete(ctx context.Context, clientID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	client, exists := m.clients[clientID]
	if !exists {
		return nil // 不存在视为删除成功
	}

	// 停止客户端
	if err := client.Stop(ctx); err != nil {
		logger.Error("Failed to stop client before delete", map[string]interface{}{
			"clientID": clientID,
			"error":    err.Error(),
		})
	}

	delete(m.clients, clientID)

	logger.Info("Tunnel client deleted", map[string]interface{}{
		"clientID": clientID,
	})

	return nil
}

// GetClient 获取指定隧道客户端实例
//
// 参数:
//   - clientID: 隧道客户端ID
//
// 返回:
//   - TunnelClient: 隧道客户端实例，如果不存在则返回 nil
func (m *TunnelClientManager) GetClient(clientID string) TunnelClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.clients[clientID]
}

// GetAllClients 获取所有隧道客户端实例
//
// 返回:
//   - []TunnelClient: 所有隧道客户端实例列表
func (m *TunnelClientManager) GetAllClients() []TunnelClient {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	clients := make([]TunnelClient, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	return clients
}

// GetClientCount 获取客户端数量
//
// 返回:
//   - int: 客户端数量
func (m *TunnelClientManager) GetClientCount() int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return len(m.clients)
}
