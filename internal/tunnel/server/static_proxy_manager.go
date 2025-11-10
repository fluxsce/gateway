// Package server 提供静态代理管理器的完整实现
// 静态代理管理器负责管理静态代理节点，支持端口转发、负载均衡等功能
package server

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/storage"
	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// StaticProxyManager 静态代理管理器接口
type StaticProxyManager interface {
	// Initialize 初始化静态代理管理器，加载所有静态代理节点
	Initialize(ctx context.Context) error

	// Start 启动所有静态代理
	Start(ctx context.Context) error

	// Stop 停止所有静态代理
	Stop(ctx context.Context) error

	// GetActiveProxies 获取所有活跃的静态代理
	GetActiveProxies() []*StaticProxyInfo

	// ReloadProxies 重新加载静态代理配置
	ReloadProxies(ctx context.Context) error
}

// staticProxyManager 静态代理管理器实现
type staticProxyManager struct {
	serverID       string
	storageManager storage.RepositoryManager
	loadBalancer   LoadBalancer
	proxies        map[string]*staticProxyInstance
	proxyMutex     sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// staticProxyInstance 静态代理实例
type staticProxyInstance struct {
	node         *types.TunnelServerNode
	listener     net.Listener
	backends     []*types.TunnelServerNode // 后端节点列表（用于负载均衡）
	status       string
	startTime    time.Time
	activeConns  int32
	totalConns   int64
	totalTraffic int64
	statsMutex   sync.RWMutex
}

// StaticProxyInfo 静态代理信息
type StaticProxyInfo struct {
	NodeID            string    `json:"nodeId"`
	NodeName          string    `json:"nodeName"`
	ProxyType         string    `json:"proxyType"`
	ListenAddress     string    `json:"listenAddress"`
	ListenPort        int       `json:"listenPort"`
	BackendCount      int       `json:"backendCount"`
	Status            string    `json:"status"`
	StartTime         time.Time `json:"startTime"`
	ActiveConnections int32     `json:"activeConnections"`
	TotalConnections  int64     `json:"totalConnections"`
	TotalTraffic      int64     `json:"totalTraffic"`
}

// NewStaticProxyManager 创建新的静态代理管理器实例
//
// 参数:
//   - serverID: 隧道服务器ID
//   - storageManager: 存储管理器实例
//   - loadBalancer: 负载均衡器实例
//
// 返回:
//   - StaticProxyManager: 静态代理管理器接口实例
//
// 功能:
//   - 初始化静态代理管理器
//   - 设置负载均衡器
//   - 准备代理实例映射表
func NewStaticProxyManager(serverID string, storageManager storage.RepositoryManager, loadBalancer LoadBalancer) StaticProxyManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &staticProxyManager{
		serverID:       serverID,
		storageManager: storageManager,
		loadBalancer:   loadBalancer,
		proxies:        make(map[string]*staticProxyInstance),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Initialize 初始化静态代理管理器，加载所有静态代理节点
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 初始化失败时返回错误
//
// 功能:
//   - 从数据库加载静态代理节点配置
//   - 按照监听地址和端口归类节点
//   - 为每个监听地址创建一个代理实例，其他同地址节点作为后端
func (m *staticProxyManager) Initialize(ctx context.Context) error {
	logger.Info("Initializing static proxy manager", map[string]interface{}{
		"serverID": m.serverID,
	})

	// 从数据库加载静态代理节点
	nodes, err := m.storageManager.GetTunnelServerNodeRepository().GetByServerID(ctx, m.serverID)
	if err != nil {
		return fmt.Errorf("failed to load server nodes: %w", err)
	}

	// 过滤出静态代理节点
	staticNodes := make([]*types.TunnelServerNode, 0)
	for _, node := range nodes {
		if node.NodeType == types.NodeTypeStatic && node.NodeStatus == types.NodeStatusActive {
			staticNodes = append(staticNodes, node)
		}
	}

	logger.Info("Loaded static proxy nodes", map[string]interface{}{
		"serverID":  m.serverID,
		"nodeCount": len(staticNodes),
	})

	// 按照监听地址和端口归类节点
	// key: "listenAddress:listenPort", value: []*types.TunnelServerNode
	groupedNodes := make(map[string][]*types.TunnelServerNode)
	for _, node := range staticNodes {
		key := fmt.Sprintf("%s:%d", node.ListenAddress, node.ListenPort)
		groupedNodes[key] = append(groupedNodes[key], node)
	}

	// 为每个监听地址创建代理实例
	m.proxyMutex.Lock()
	defer m.proxyMutex.Unlock()

	for listenKey, nodeGroup := range groupedNodes {
		if len(nodeGroup) == 0 {
			continue
		}

		// 使用第一个节点作为主节点（创建监听器）
		primaryNode := nodeGroup[0]

		// 其余节点作为后端节点（用于负载均衡）
		backends := nodeGroup[1:]

		proxy := &staticProxyInstance{
			node:      primaryNode,
			backends:  backends,
			status:    "initialized",
			startTime: time.Now(),
		}

		m.proxies[primaryNode.ServerNodeId] = proxy

		logger.Debug("Static proxy instance created", map[string]interface{}{
			"nodeID":       primaryNode.ServerNodeId,
			"nodeName":     primaryNode.NodeName,
			"proxyType":    primaryNode.ProxyType,
			"listenAddr":   listenKey,
			"backendCount": len(backends),
		})
	}

	return nil
}

// Start 启动所有静态代理
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 启动失败时返回错误
//
// 功能:
//   - 为每个静态代理创建监听器
//   - 启动连接接受循环
//   - 更新代理状态
func (m *staticProxyManager) Start(ctx context.Context) error {
	logger.Info("Starting static proxies", map[string]interface{}{
		"serverID": m.serverID,
	})

	m.proxyMutex.RLock()
	proxies := make([]*staticProxyInstance, 0, len(m.proxies))
	for _, proxy := range m.proxies {
		proxies = append(proxies, proxy)
	}
	m.proxyMutex.RUnlock()

	// 启动每个代理
	for _, proxy := range proxies {
		if err := m.startProxy(ctx, proxy); err != nil {
			logger.Error("Failed to start static proxy", err, map[string]interface{}{
				"nodeID":   proxy.node.ServerNodeId,
				"nodeName": proxy.node.NodeName,
			})
			continue
		}

		logger.Info("Static proxy started", map[string]interface{}{
			"nodeID":       proxy.node.ServerNodeId,
			"nodeName":     proxy.node.NodeName,
			"proxyType":    proxy.node.ProxyType,
			"listenAddr":   fmt.Sprintf("%s:%d", proxy.node.ListenAddress, proxy.node.ListenPort),
			"backendCount": len(proxy.backends),
		})
	}

	return nil
}

// Stop 停止所有静态代理
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 停止失败时返回错误
//
// 功能:
//   - 关闭所有监听器
//   - 等待现有连接完成
//   - 清理资源
func (m *staticProxyManager) Stop(ctx context.Context) error {
	logger.Info("Stopping static proxies", map[string]interface{}{
		"serverID": m.serverID,
	})

	// 取消上下文
	m.cancel()

	m.proxyMutex.Lock()
	defer m.proxyMutex.Unlock()

	// 停止所有代理
	for _, proxy := range m.proxies {
		if proxy.listener != nil {
			proxy.listener.Close()
		}
		proxy.status = "stopped"
	}

	// 等待所有连接处理完成
	done := make(chan struct{})
	go func() {
		m.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("All static proxies stopped gracefully", nil)
	case <-time.After(30 * time.Second):
		logger.Warn("Timeout waiting for static proxies to stop", nil)
	}

	return nil
}

// GetActiveProxies 获取所有活跃的静态代理
//
// 返回:
//   - []*StaticProxyInfo: 静态代理信息列表
func (m *staticProxyManager) GetActiveProxies() []*StaticProxyInfo {
	m.proxyMutex.RLock()
	defer m.proxyMutex.RUnlock()

	proxies := make([]*StaticProxyInfo, 0, len(m.proxies))
	for _, proxy := range m.proxies {
		proxy.statsMutex.RLock()
		info := &StaticProxyInfo{
			NodeID:            proxy.node.ServerNodeId,
			NodeName:          proxy.node.NodeName,
			ProxyType:         proxy.node.ProxyType,
			ListenAddress:     proxy.node.ListenAddress,
			ListenPort:        proxy.node.ListenPort,
			BackendCount:      len(proxy.backends),
			Status:            proxy.status,
			StartTime:         proxy.startTime,
			ActiveConnections: proxy.activeConns,
			TotalConnections:  proxy.totalConns,
			TotalTraffic:      proxy.totalTraffic,
		}
		proxy.statsMutex.RUnlock()
		proxies = append(proxies, info)
	}

	return proxies
}

// ReloadProxies 重新加载静态代理配置
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 重新加载失败时返回错误
func (m *staticProxyManager) ReloadProxies(ctx context.Context) error {
	logger.Info("Reloading static proxies", map[string]interface{}{
		"serverID": m.serverID,
	})

	// 停止现有代理
	if err := m.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop existing proxies: %w", err)
	}

	// 重新初始化
	if err := m.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to reinitialize proxies: %w", err)
	}

	// 重新启动
	if err := m.Start(ctx); err != nil {
		return fmt.Errorf("failed to restart proxies: %w", err)
	}

	return nil
}

// startProxy 启动单个静态代理
func (m *staticProxyManager) startProxy(ctx context.Context, proxy *staticProxyInstance) error {
	addr := fmt.Sprintf("%s:%d", proxy.node.ListenAddress, proxy.node.ListenPort)

	// 根据代理类型创建监听器
	switch proxy.node.ProxyType {
	case types.ProxyTypeTCP:
		return m.startTCPProxy(ctx, proxy, addr)
	case types.ProxyTypeUDP:
		return m.startUDPProxy(ctx, proxy, addr)
	case types.ProxyTypeHTTP, types.ProxyTypeHTTPS:
		return m.startHTTPProxy(ctx, proxy, addr)
	default:
		return fmt.Errorf("unsupported proxy type: %s", proxy.node.ProxyType)
	}
}

// startTCPProxy 启动TCP代理
func (m *staticProxyManager) startTCPProxy(ctx context.Context, proxy *staticProxyInstance, addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	proxy.listener = listener
	proxy.status = "running"

	// 启动连接接受循环
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.acceptTCPConnections(ctx, proxy)
	}()

	return nil
}

// startUDPProxy 启动UDP代理
func (m *staticProxyManager) startUDPProxy(ctx context.Context, proxy *staticProxyInstance, addr string) error {
	// UDP代理实现
	logger.Warn("UDP proxy not fully implemented yet", map[string]interface{}{
		"nodeID": proxy.node.ServerNodeId,
	})
	return nil
}

// startHTTPProxy 启动HTTP/HTTPS代理
func (m *staticProxyManager) startHTTPProxy(ctx context.Context, proxy *staticProxyInstance, addr string) error {
	// HTTP代理实现
	logger.Warn("HTTP proxy not fully implemented yet", map[string]interface{}{
		"nodeID": proxy.node.ServerNodeId,
	})
	return nil
}

// acceptTCPConnections 接受TCP连接
func (m *staticProxyManager) acceptTCPConnections(ctx context.Context, proxy *staticProxyInstance) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-m.ctx.Done():
			return
		default:
			conn, err := proxy.listener.Accept()
			if err != nil {
				if m.ctx.Err() != nil {
					return
				}
				logger.Error("Failed to accept connection", err, map[string]interface{}{
					"nodeID": proxy.node.ServerNodeId,
				})
				continue
			}

			// 为每个连接启动处理协程
			m.wg.Add(1)
			go func(conn net.Conn) {
				defer m.wg.Done()
				m.handleTCPConnection(ctx, proxy, conn)
			}(conn)
		}
	}
}

// handleTCPConnection 处理TCP连接
func (m *staticProxyManager) handleTCPConnection(ctx context.Context, proxy *staticProxyInstance, clientConn net.Conn) {
	defer clientConn.Close()

	// 更新连接统计
	proxy.statsMutex.Lock()
	proxy.activeConns++
	proxy.totalConns++
	proxy.statsMutex.Unlock()

	defer func() {
		proxy.statsMutex.Lock()
		proxy.activeConns--
		proxy.statsMutex.Unlock()
	}()

	// 使用负载均衡器选择后端节点
	backend, err := m.selectBackend(ctx, proxy)
	if err != nil {
		logger.Error("Failed to select backend", err, map[string]interface{}{
			"nodeID": proxy.node.ServerNodeId,
		})
		return
	}

	// 连接到后端
	targetAddr := net.JoinHostPort(backend.TargetAddress, fmt.Sprintf("%d", backend.TargetPort))
	backendConn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		logger.Error("Failed to connect to backend", err, map[string]interface{}{
			"nodeID":  proxy.node.ServerNodeId,
			"backend": targetAddr,
		})
		return
	}
	defer backendConn.Close()

	// 双向数据转发
	m.relayData(clientConn, backendConn, proxy)
}

// selectBackend 选择后端节点
//
// 根据负载均衡策略选择最优的后端节点。
//
// 选择逻辑：
//  1. 如果没有后端节点（单节点模式），直接使用主节点的目标地址
//  2. 如果有多个后端节点（负载均衡模式），使用负载均衡器选择最优节点
//
// 负载均衡支持多种算法：
//   - round_robin: 轮询
//   - least_connections: 最少连接
//   - least_latency: 最低延迟
//   - weighted_random: 加权随机
//   - health_based: 基于健康状况
func (m *staticProxyManager) selectBackend(ctx context.Context, proxy *staticProxyInstance) (*types.TunnelServerNode, error) {
	if len(proxy.backends) == 0 {
		// 单节点模式：直接使用主节点的目标地址
		return proxy.node, nil
	}

	// 负载均衡模式：将主节点也加入候选列表
	allBackends := make([]*types.TunnelServerNode, 0, len(proxy.backends)+1)
	allBackends = append(allBackends, proxy.node)
	allBackends = append(allBackends, proxy.backends...)

	// 使用负载均衡器选择后端
	backend, err := m.loadBalancer.SelectNode(ctx, allBackends)
	if err != nil {
		return nil, fmt.Errorf("failed to select backend node: %w", err)
	}

	return backend, nil
}

// relayData 双向数据转发
func (m *staticProxyManager) relayData(clientConn, backendConn net.Conn, proxy *staticProxyInstance) {
	var wg sync.WaitGroup
	var totalBytes int64

	// 客户端 -> 后端
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 32*1024)
		for {
			n, err := clientConn.Read(buffer)
			if err != nil {
				return
			}
			if n > 0 {
				_, err := backendConn.Write(buffer[:n])
				if err != nil {
					return
				}
				totalBytes += int64(n)
			}
		}
	}()

	// 后端 -> 客户端
	wg.Add(1)
	go func() {
		defer wg.Done()
		buffer := make([]byte, 32*1024)
		for {
			n, err := backendConn.Read(buffer)
			if err != nil {
				return
			}
			if n > 0 {
				_, err := clientConn.Write(buffer[:n])
				if err != nil {
					return
				}
				totalBytes += int64(n)
			}
		}
	}()

	wg.Wait()

	// 更新流量统计
	proxy.statsMutex.Lock()
	proxy.totalTraffic += totalBytes
	proxy.statsMutex.Unlock()
}
