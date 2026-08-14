// Package static 实现静态代理管理器
// 负责管理静态隧道服务器和节点的生命周期
// 提供启动、停止、重载等公用方法
package static

import (
	"context"
	"fmt"
	"sync"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// StaticProxyManager 静态代理管理器
// 负责管理静态隧道服务器和节点的生命周期
type StaticProxyManager struct {
	staticServerRepo *storage.TunnelStaticServerRepositoryImpl // 静态服务器存储接口
	staticNodeRepo   *storage.TunnelStaticNodeRepositoryImpl   // 静态节点存储接口

	servers map[string]*staticProxy // 静态服务器映射: serverID -> staticProxy
	mutex   sync.RWMutex            // 并发保护

	cancel context.CancelFunc // 用于 StopAll 时取消
	wg     sync.WaitGroup     // 等待所有 goroutine 完成
}

var (
	// 全局唯一的静态代理管理器实例（单例模式）
	globalManager *StaticProxyManager
	managerMutex  sync.RWMutex
)

// GetStaticProxyManager 获取全局唯一的静态代理管理器实例（单例模式）
// 注意：globalManager 在 Initialize 时创建，在此之前返回 nil
//
// 返回:
//   - *StaticProxyManager: 静态代理管理器实例，如果未初始化则返回 nil
func GetStaticProxyManager() *StaticProxyManager {
	managerMutex.RLock()
	defer managerMutex.RUnlock()
	return globalManager
}

// EnsureInitialized 返回已初始化的静态代理管理器；尚未初始化时用 db 完成初始化。
// 供管理接口在隧道总开关关闭或启动阶段初始化失败时按需创建管理器。
func EnsureInitialized(ctx context.Context, db database.Database) (*StaticProxyManager, error) {
	if mgr := GetStaticProxyManager(); mgr != nil {
		return mgr, nil
	}
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}
	return Initialize(ctx, db)
}

// Initialize 初始化静态代理管理器
// 从数据库加载所有静态服务器和节点配置，并创建代理实例缓存
// 此方法会创建全局唯一的 StaticProxyManager 实例
//
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接接口
//
// 返回:
//   - *StaticProxyManager: 静态代理管理器实例
//   - error: 错误信息
func Initialize(ctx context.Context, db database.Database) (*StaticProxyManager, error) {
	managerMutex.Lock()
	defer managerMutex.Unlock()

	// 如果已经初始化，直接返回
	if globalManager != nil {
		return globalManager, nil
	}

	// 从父 context 派生取消函数（用于 StopAll）
	// 当父 context 取消时，管理器也会收到取消信号
	_, cancel := context.WithCancel(ctx)

	// 直接创建 repository，不通过 storageManager
	staticServerRepo := storage.NewTunnelStaticServerRepository(db)
	staticNodeRepo := storage.NewTunnelStaticNodeRepository(db)

	// 创建全局管理器实例
	globalManager = &StaticProxyManager{
		staticServerRepo: staticServerRepo,
		staticNodeRepo:   staticNodeRepo,
		servers:          make(map[string]*staticProxy),
		cancel:           cancel,
	}

	logger.Info("Initializing static proxy manager", nil)

	// 预加载失败不阻断管理器创建，启动单个服务时会按需从数据库加载
	servers, err := staticServerRepo.GetAll(ctx)
	if err != nil {
		logger.Error("Failed to preload static servers, will load on demand", map[string]interface{}{
			"error": err.Error(),
		})
		return globalManager, nil
	}

	logger.Info("Loading static servers", map[string]interface{}{
		"count": len(servers),
	})

	// 为每个服务器加载节点并创建代理实例
	for _, server := range servers {
		// 加载该服务器的所有节点
		nodes, err := staticNodeRepo.GetByServerID(ctx, server.TunnelStaticServerId)
		if err != nil {
			logger.Error("Failed to load nodes for server", map[string]interface{}{
				"serverID": server.TunnelStaticServerId,
				"error":    err.Error(),
			})
			continue
		}

		// 将节点列表设置到服务器对象中
		server.Nodes = nodes

		// 确保节点列表已初始化
		if server.Nodes == nil {
			server.Nodes = make([]*types.TunnelStaticNode, 0)
		}

		// 创建静态代理实例并缓存
		proxy := newStaticProxy(server, staticServerRepo, staticNodeRepo)
		globalManager.servers[server.TunnelStaticServerId] = proxy

		logger.Info("Loaded static server", map[string]interface{}{
			"serverID":   server.TunnelStaticServerId,
			"serverName": server.ServerName,
			"nodeCount":  len(nodes),
		})
	}

	logger.Info("Static proxy manager initialized", map[string]interface{}{
		"serverCount": len(globalManager.servers),
	})

	return globalManager, nil
}

// InitializeAndStart 初始化并启动所有静态代理
// 从数据库加载所有静态服务器和节点配置，然后启动所有代理
//
// 参数:
//   - ctx: 上下文
//   - db: 数据库连接接口
//
// 返回:
//   - error: 错误信息
func InitializeAndStart(ctx context.Context, db database.Database) error {
	// 先初始化
	manager, err := Initialize(ctx, db)
	if err != nil {
		return fmt.Errorf("failed to initialize: %w", err)
	}

	// 再启动所有代理
	return manager.StartAll(ctx)
}

// Start 启动指定静态代理。
// 每次启动都会从数据库重新加载服务器配置和节点列表，避免进程初始化后新增的节点仍停留在空缓存里。
// 如果代理实例不存在，会自动创建并加入缓存。
//
// 参数:
//   - ctx: 上下文
//   - serverID: 静态服务器ID
//
// 返回:
//   - error: 错误信息
func (m *StaticProxyManager) Start(ctx context.Context, serverID string) error {
	server, err := m.loadServerWithNodes(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to load static proxy %s: %w", serverID, err)
	}
	if len(server.Nodes) == 0 {
		return fmt.Errorf("该静态服务没有可用节点，请先添加节点后再启动")
	}

	m.mutex.Lock()
	proxy, exists := m.servers[serverID]
	if !exists {
		proxy = newStaticProxy(server, m.staticServerRepo, m.staticNodeRepo)
		m.servers[serverID] = proxy
	}
	m.mutex.Unlock()

	if exists {
		proxy.applyServerConfig(server)
	}

	if err := proxy.Start(ctx); err != nil {
		return fmt.Errorf("failed to start static proxy %s: %w", serverID, err)
	}

	logger.Info("Static proxy started", map[string]interface{}{
		"serverID":  serverID,
		"nodeCount": len(server.Nodes),
	})

	return nil
}

// StartAll 启动所有静态代理
// 启动所有已加载的静态代理实例
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *StaticProxyManager) StartAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Starting all static proxies", map[string]interface{}{
		"count": len(m.servers),
	})

	// 启动所有静态代理
	for serverID, proxy := range m.servers {
		if err := proxy.Start(ctx); err != nil {
			logger.Error("Failed to start static proxy", map[string]interface{}{
				"serverID": serverID,
				"error":    err.Error(),
			})
			// 继续启动其他代理，不中断
		}
	}

	logger.Info("All static proxies started", map[string]interface{}{
		"count": len(m.servers),
	})

	return nil
}

// Stop 停止指定静态代理
// 停止指定 serverID 的静态代理实例并从缓存中移除
//
// 参数:
//   - ctx: 上下文
//   - serverID: 静态服务器ID
//
// 返回:
//   - error: 错误信息
//
// 重要说明:
//   - 停止后会从缓存中移除，下次启动时会从数据库重新加载
func (m *StaticProxyManager) Stop(ctx context.Context, serverID string) error {
	m.mutex.Lock()
	proxy, exists := m.servers[serverID]
	if !exists {
		m.mutex.Unlock()
		return fmt.Errorf("static proxy not found: %s", serverID)
	}
	m.mutex.Unlock()

	// 不持有管理器锁时调用 proxy.Stop，避免长时间持有锁
	if err := proxy.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop static proxy %s: %w", serverID, err)
	}

	// 从缓存中移除已停止的代理
	m.mutex.Lock()
	delete(m.servers, serverID)
	m.mutex.Unlock()

	logger.Info("Static proxy stopped and removed from cache", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// StopAll 停止所有静态代理
// 停止所有正在运行的静态代理实例
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 错误信息
func (m *StaticProxyManager) StopAll(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	logger.Info("Stopping all static proxies", map[string]interface{}{
		"count": len(m.servers),
	})

	// 取消上下文
	m.cancel()

	// 停止所有静态代理
	for serverID, proxy := range m.servers {
		if err := proxy.Stop(ctx); err != nil {
			logger.Error("Failed to stop static proxy", map[string]interface{}{
				"serverID": serverID,
				"error":    err.Error(),
			})
		}
	}

	// 等待所有goroutine完成
	m.wg.Wait()

	logger.Info("All static proxies stopped", nil)

	return nil
}

// Reload 重载指定静态代理配置
// 从数据库重新加载指定 serverID 的配置并更新运行中的代理
//
// 参数:
//   - ctx: 上下文
//   - serverID: 静态服务器ID
//
// 返回:
//   - error: 错误信息
func (m *StaticProxyManager) Reload(ctx context.Context, serverID string) error {
	// 从数据库获取最新配置（不持有锁）
	server, err := m.staticServerRepo.GetByID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to get server config: %w", err)
	}
	if server == nil {
		return fmt.Errorf("server %s not found", serverID)
	}

	// 加载节点列表
	nodes, err := m.staticNodeRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return fmt.Errorf("failed to load nodes: %w", err)
	}
	server.Nodes = nodes

	// 获取或创建代理
	m.mutex.Lock()
	proxy, exists := m.servers[serverID]
	if !exists {
		// 创建新的代理实例
		if server.Nodes == nil {
			server.Nodes = make([]*types.TunnelStaticNode, 0)
		}
		proxy = newStaticProxy(server, m.staticServerRepo, m.staticNodeRepo)
		m.servers[serverID] = proxy
	}
	m.mutex.Unlock()

	// 重新加载配置（不持有管理器锁）
	if exists {
		if err := proxy.ReloadConfig(ctx, server); err != nil {
			return fmt.Errorf("failed to reload proxy config: %w", err)
		}
	}

	logger.Info("Static proxy reloaded", map[string]interface{}{
		"serverID": serverID,
	})

	return nil
}

// GetStaticProxy 获取指定静态代理实例
//
// 参数:
//   - serverID: 静态服务器ID
//
// 返回:
//   - *staticProxy: 静态代理实例，如果不存在则返回 nil
func (m *StaticProxyManager) GetStaticProxy(serverID string) *staticProxy {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.servers[serverID]
}

// loadServerWithNodes 从数据库加载服务器配置及其节点列表。
func (m *StaticProxyManager) loadServerWithNodes(ctx context.Context, serverID string) (*types.TunnelStaticServer, error) {
	server, err := m.staticServerRepo.GetByID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to get server config: %w", err)
	}
	if server == nil {
		return nil, fmt.Errorf("server %s not found in database", serverID)
	}

	nodes, err := m.staticNodeRepo.GetByServerID(ctx, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to load nodes: %w", err)
	}
	if nodes == nil {
		nodes = make([]*types.TunnelStaticNode, 0)
	}
	server.Nodes = nodes
	return server, nil
}

// loadAndCreateProxy 从数据库加载服务器配置并创建代理实例。
func (m *StaticProxyManager) loadAndCreateProxy(ctx context.Context, serverID string) (*staticProxy, error) {
	server, err := m.loadServerWithNodes(ctx, serverID)
	if err != nil {
		return nil, err
	}

	proxy := newStaticProxy(server, m.staticServerRepo, m.staticNodeRepo)

	logger.Info("Loaded static server from database", map[string]interface{}{
		"serverID":   serverID,
		"serverName": server.ServerName,
		"nodeCount":  len(server.Nodes),
	})

	return proxy, nil
}
