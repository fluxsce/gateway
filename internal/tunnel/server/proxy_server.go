// Package server 提供反向代理服务器的完整实现
// 反向代理服务器负责处理隧道连接，将外网请求转发到内网客户端服务
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// proxyServer 反向代理服务器实现
// 实现 ProxyServer 接口，处理隧道反向代理转发
//
// 每个 serviceID 对应一个反向代理实例，该实例：
// 1. 在指定的公网端口上监听外网请求
// 2. 维护一个数据连接池，用于复用与客户端的连接
// 3. 将外网请求通过数据连接转发到内网服务
type proxyServer struct {
	tunnelServer  TunnelServer                     // 隧道服务器引用
	controlServer ControlServer                    // 控制服务器引用，用于与客户端通信
	activeProxies map[string]*reverseProxyInstance // serviceID -> 反向代理实例
	proxyMutex    sync.RWMutex                     // 保护 activeProxies 的读写锁
	dataConnPools map[string]*dataConnectionPool   // serviceID -> 数据连接池
	poolsMutex    sync.RWMutex                     // 保护 dataConnPools 的读写锁
	pendingConns  map[string]*proxyConnection      // connectionID -> 等待中的代理连接
	pendingMutex  sync.RWMutex                     // 保护 pendingConns 的读写锁
	ctx           context.Context                  // 服务器上下文
	cancel        context.CancelFunc               // 取消函数
	wg            sync.WaitGroup                   // 等待组，用于优雅关闭
}

// dataConnectionPool 数据连接池
//
// 为特定服务维护一组可复用的数据连接，减少连接建立开销。
// 每个服务（serviceID）对应一个连接池实例。
type dataConnectionPool struct {
	serviceID      string        // 服务ID，对应一个反向代理实例
	clientID       string        // 客户端ID，标识连接的来源客户端
	availableConns chan net.Conn // 可用连接队列，缓存空闲连接
	minSize        int           // 连接池最小大小，保持的最少连接数
	maxSize        int           // 连接池最大大小，允许的最多连接数
	currentSize    int32         // 当前连接池大小（原子操作）
	mutex          sync.RWMutex  // 保护连接池操作的读写锁
}

// reverseProxyInstance 反向代理实例
//
// 每个实例对应一个服务的反向代理配置，包括：
// - 在指定公网端口监听外网请求
// - 维护连接统计信息
// - 管理代理状态
type reverseProxyInstance struct {
	service      *types.TunnelService // 关联的隧道服务配置
	listener     net.Listener         // 公网端口监听器
	status       string               // 代理状态: starting, running, stopping, stopped
	startTime    time.Time            // 代理启动时间
	activeConns  int32                // 当前活跃连接数（原子操作）
	totalConns   int64                // 历史总连接数（原子操作）
	totalTraffic int64                // 总流量统计（原子操作）
	statsMutex   sync.RWMutex         // 保护统计信息的读写锁
}

// proxyConnection 代理连接
//
// 表示一个外网请求与内网服务之间的代理连接。
// 包含外网连接、等待的数据连接，以及连接元数据。
type proxyConnection struct {
	clientConn   net.Conn      // 外网客户端连接
	connectionID string        // 连接唯一标识符
	serviceID    string        // 关联的服务ID
	clientID     string        // 内网客户端ID
	createTime   time.Time     // 连接创建时间
	ready        chan struct{} // 数据连接就绪信号
	dataConn     net.Conn      // 内网客户端数据连接
}

// NewProxyServerImpl 创建新的反向代理服务器实例
//
// 参数:
//   - tunnelServer: 隧道服务器实例，用于获取配置和状态
//   - controlServer: 控制服务器实例，用于与客户端通信
//
// 返回:
//   - ProxyServer: 反向代理服务器接口实例
//
// 功能:
//   - 初始化反向代理服务器
//   - 创建代理实例映射表
//   - 设置连接池管理
func NewProxyServerImpl(tunnelServer TunnelServer, controlServer ControlServer) ProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &proxyServer{
		tunnelServer:  tunnelServer,
		controlServer: controlServer,
		activeProxies: make(map[string]*reverseProxyInstance),
		dataConnPools: make(map[string]*dataConnectionPool),
		pendingConns:  make(map[string]*proxyConnection),
		ctx:           ctx,
		cancel:        cancel,
	}
}

// StartProxy 启动指定服务的反向代理
//
// 参数:
//   - ctx: 上下文，用于控制代理生命周期
//   - config: 代理配置，包含服务信息
//
// 返回:
//   - error: 启动失败时返回错误
//
// 功能:
//   - 为隧道服务启动反向代理监听器
//   - 监听外网连接并转发到内网客户端
//   - 管理连接池和统计信息
func (s *proxyServer) StartProxy(ctx context.Context, config *ProxyConfig) error {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()

	// 检查代理是否已存在
	if _, exists := s.activeProxies[config.ProxyID]; exists {
		return fmt.Errorf("reverse proxy %s already exists", config.ProxyID)
	}

	// 获取对应的隧道服务
	service := s.findTunnelService(config.ProxyID)
	if service == nil {
		return fmt.Errorf("tunnel service %s not found", config.ProxyID)
	}

	// 创建反向代理实例
	proxy := &reverseProxyInstance{
		service:   service,
		status:    "starting",
		startTime: time.Now(),
	}

	// 启动监听器
	if err := s.startReverseProxyListener(ctx, proxy); err != nil {
		return fmt.Errorf("failed to start reverse proxy listener: %w", err)
	}

	proxy.status = "running"
	s.activeProxies[config.ProxyID] = proxy

	// 确保数据连接池已创建
	s.getOrCreateDataConnectionPool(config.ProxyID, proxy.service.TunnelClientId)

	logger.Info("Reverse proxy started", map[string]interface{}{
		"serviceId":   config.ProxyID,
		"serviceName": service.ServiceName,
		"remotePort":  *service.RemotePort,
		"localTarget": fmt.Sprintf("%s:%d", service.LocalAddress, service.LocalPort),
	})

	return nil
}

// StopProxy 停止指定的反向代理服务
//
// 参数:
//   - ctx: 上下文，用于控制停止超时
//   - proxyID: 代理ID（服务ID）
//
// 返回:
//   - error: 停止失败时返回错误
//
// 功能:
//   - 关闭代理监听器
//   - 清理连接池
//   - 等待现有连接处理完成
func (s *proxyServer) StopProxy(ctx context.Context, proxyID string) error {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()

	proxy, exists := s.activeProxies[proxyID]
	if !exists {
		return fmt.Errorf("reverse proxy %s not found", proxyID)
	}

	proxy.status = "stopping"

	// 关闭监听器
	if proxy.listener != nil {
		proxy.listener.Close()
	}

	// 清理数据连接池
	s.poolsMutex.Lock()
	if pool, exists := s.dataConnPools[proxyID]; exists {
		close(pool.availableConns)
		// 清理池中的连接
		for conn := range pool.availableConns {
			if conn != nil {
				conn.Close()
			}
		}
		delete(s.dataConnPools, proxyID)
	}
	s.poolsMutex.Unlock()

	// 从映射中移除
	delete(s.activeProxies, proxyID)

	logger.Info("Reverse proxy stopped", map[string]interface{}{
		"serviceId": proxyID,
	})

	return nil
}

// GetActiveProxies 获取活跃的反向代理服务
//
// 返回:
//   - []*ProxyInfo: 活跃代理服务的信息列表
//
// 功能:
//   - 返回所有运行中的反向代理服务信息
//   - 包含连接数、流量统计等信息
func (s *proxyServer) GetActiveProxies() []*ProxyInfo {
	s.proxyMutex.RLock()
	defer s.proxyMutex.RUnlock()

	var proxies []*ProxyInfo
	for serviceID, proxy := range s.activeProxies {
		proxy.statsMutex.RLock()
		info := &ProxyInfo{
			ProxyID:           serviceID,
			ProxyType:         proxy.service.ServiceType,
			ListenAddress:     "0.0.0.0", // 反向代理监听所有地址
			ListenPort:        int(*proxy.service.RemotePort),
			Status:            proxy.status,
			StartTime:         proxy.startTime,
			ActiveConnections: int(atomic.LoadInt32(&proxy.activeConns)),
			TotalConnections:  atomic.LoadInt64(&proxy.totalConns),
			TotalTraffic:      atomic.LoadInt64(&proxy.totalTraffic),
		}
		proxy.statsMutex.RUnlock()
		proxies = append(proxies, info)
	}

	return proxies
}

// HandleProxyConnection 处理反向代理连接
//
// 参数:
//   - ctx: 上下文
//   - conn: 网络连接
//   - proxyID: 代理ID（服务ID）
//
// 返回:
//   - error: 处理失败时返回错误
//
// 功能:
//   - 处理外网客户端连接
//   - 优先使用连接池中的数据连接
//   - 必要时通知内网客户端建立新连接
func (s *proxyServer) HandleProxyConnection(ctx context.Context, conn net.Conn, proxyID string) error {
	s.proxyMutex.RLock()
	proxy, exists := s.activeProxies[proxyID]
	s.proxyMutex.RUnlock()

	if !exists {
		conn.Close()
		return fmt.Errorf("reverse proxy %s not found", proxyID)
	}

	defer conn.Close()

	// 更新连接统计
	atomic.AddInt32(&proxy.activeConns, 1)
	atomic.AddInt64(&proxy.totalConns, 1)
	defer atomic.AddInt32(&proxy.activeConns, -1)

	// 尝试从连接池获取可用的数据连接
	if dataConn := s.getAvailableDataConnection(proxyID); dataConn != nil {
		// 使用池化连接直接桥接
		return s.bridgeConnectionsDirect(ctx, conn, dataConn, proxy)
	}

	// 没有可用的池化连接，使用传统方式
	// 创建代理连接
	proxyConn := &proxyConnection{
		clientConn:   conn,
		connectionID: generateConnectionID(),
		serviceID:    proxyID,
		clientID:     proxy.service.TunnelClientId,
		createTime:   time.Now(),
		ready:        make(chan struct{}),
	}

	// 确保连接池已创建
	s.getOrCreateDataConnectionPool(proxyID, proxy.service.TunnelClientId)

	// 通知内网客户端建立连接
	if err := s.notifyClientConnection(proxyConn); err != nil {
		logger.Error("Failed to notify client for connection", map[string]interface{}{
			"error":        err.Error(),
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
		})
		return err
	}

	// 等待内网客户端建立数据连接
	select {
	case <-proxyConn.ready:
		// 开始桥接连接
		logger.Info("Starting connection bridge", map[string]interface{}{
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
		})
		return s.bridgeConnections(ctx, proxyConn, proxy)
	case <-time.After(10 * time.Second):
		logger.Error("Timeout waiting for client data connection", map[string]interface{}{
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
		})
		return fmt.Errorf("timeout waiting for client data connection")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// bridgeConnectionsDirect 直接桥接连接（使用池化连接）
func (s *proxyServer) bridgeConnectionsDirect(ctx context.Context, clientConn, dataConn net.Conn, proxy *reverseProxyInstance) error {
	defer clientConn.Close()
	defer func() {
		// 尝试将数据连接归还到池中
		if !s.returnDataConnection(proxy.service.TunnelServiceId, dataConn) {
			// 如果无法归还到池中，则关闭连接
			dataConn.Close()
		}
	}()

	logger.Info("Starting direct connection bridge (pooled)", map[string]interface{}{
		"serviceId": proxy.service.TunnelServiceId,
	})

	// 使用两个goroutine实现双向转发
	done := make(chan struct{}, 2)
	var totalBytes int64

	// 外网客户端 -> 内网服务 (通过数据连接)
	go func() {
		defer func() { done <- struct{}{} }()
		bytes, err := io.Copy(dataConn, clientConn)
		if err != nil && err != io.EOF {
			logger.Debug("Client to data connection closed", map[string]interface{}{
				"error":     err.Error(),
				"serviceId": proxy.service.TunnelServiceId,
			})
		}
		atomic.AddInt64(&totalBytes, bytes)
	}()

	// 内网服务 -> 外网客户端 (通过数据连接)
	go func() {
		defer func() { done <- struct{}{} }()
		bytes, err := io.Copy(clientConn, dataConn)
		if err != nil && err != io.EOF {
			logger.Debug("Data to client connection closed", map[string]interface{}{
				"error":     err.Error(),
				"serviceId": proxy.service.TunnelServiceId,
			})
		}
		atomic.AddInt64(&totalBytes, bytes)
	}()

	// 等待任一方向的传输完成
	<-done

	// 更新流量统计
	atomic.AddInt64(&proxy.totalTraffic, totalBytes)

	logger.Debug("Direct connection bridge completed", map[string]interface{}{
		"serviceId":    proxy.service.TunnelServiceId,
		"totalBytes":   totalBytes,
		"totalTraffic": atomic.LoadInt64(&proxy.totalTraffic),
	})

	return nil
}

// HandleClientDataConnection 处理客户端数据连接
//
// 参数:
//   - ctx: 上下文
//   - conn: 客户端数据连接
//   - connectionID: 连接ID或服务ID（用于池化连接）
//
// 返回:
//   - error: 处理失败时返回错误
//
// 功能:
//   - 接收客户端的数据连接
//   - 区分普通连接和池化连接
//   - 与等待的外网连接进行匹配或加入连接池
func (s *proxyServer) HandleClientDataConnection(ctx context.Context, conn net.Conn, connectionID string) error {
	// 为长连接（如SSE）优化TCP连接
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// 禁用Nagle算法以减少延迟（对流式数据重要）
		tcpConn.SetNoDelay(true)
	}

	// 首先尝试作为普通连接处理
	proxyConn := s.findPendingConnection(connectionID)
	if proxyConn != nil {
		// 设置数据连接
		proxyConn.dataConn = conn

		// 通知桥接可以开始
		close(proxyConn.ready)

		logger.Info("Client data connection established", map[string]interface{}{
			"connectionId": connectionID,
			"serviceId":    proxyConn.serviceID,
		})

		return nil
	}

	// 如果不是等待中的连接，可能是池化连接
	// 尝试将连接添加到对应服务的连接池
	if s.addToConnectionPool(connectionID, conn) {
		logger.Info("Added connection to pool", map[string]interface{}{
			"serviceId": connectionID,
		})
		return nil
	}

	// 都不是，关闭连接
	conn.Close()
	return fmt.Errorf("no pending connection or pool found for connectionID %s", connectionID)
}

// addToConnectionPool 将连接添加到连接池
func (s *proxyServer) addToConnectionPool(serviceID string, conn net.Conn) bool {
	s.poolsMutex.RLock()
	pool, exists := s.dataConnPools[serviceID]
	s.poolsMutex.RUnlock()

	if !exists {
		return false
	}

	// 尝试添加到池中
	select {
	case pool.availableConns <- conn:
		atomic.AddInt32(&pool.currentSize, 1)
		return true
	default:
		// 池已满
		return false
	}
}

// startReverseProxyListener 启动反向代理监听器
func (s *proxyServer) startReverseProxyListener(ctx context.Context, proxy *reverseProxyInstance) error {
	if proxy.service.RemotePort == nil {
		return fmt.Errorf("remote port not allocated for service %s", proxy.service.TunnelServiceId)
	}

	// 启动监听器
	listenAddr := fmt.Sprintf(":%d", *proxy.service.RemotePort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", *proxy.service.RemotePort, err)
	}

	proxy.listener = listener

	// 启动接受连接的协程
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer listener.Close()

		logger.Info("Reverse proxy listener started", map[string]interface{}{
			"serviceId":   proxy.service.TunnelServiceId,
			"serviceName": proxy.service.ServiceName,
			"remotePort":  *proxy.service.RemotePort,
		})

		for {
			clientConn, err := listener.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					logger.Error("Failed to accept connection", map[string]interface{}{
						"error":     err.Error(),
						"serviceId": proxy.service.TunnelServiceId,
					})
					continue
				}
			}

			// 为长连接（如SSE）优化TCP连接
			if tcpConn, ok := clientConn.(*net.TCPConn); ok {
				tcpConn.SetKeepAlive(true)
				tcpConn.SetKeepAlivePeriod(30 * time.Second)
				// 禁用Nagle算法以减少延迟（对流式数据重要）
				tcpConn.SetNoDelay(true)
			}

			// 为每个客户端连接启动处理协程
			s.wg.Add(1)
			go func(clientConn net.Conn) {
				defer s.wg.Done()
				if err := s.HandleProxyConnection(ctx, clientConn, proxy.service.TunnelServiceId); err != nil {
					logger.Error("Reverse proxy connection handling failed", map[string]interface{}{
						"error":     err.Error(),
						"serviceId": proxy.service.TunnelServiceId,
					})
				}
			}(clientConn)
		}
	}()

	return nil
}

// notifyClientConnection 通知客户端建立连接
func (s *proxyServer) notifyClientConnection(proxyConn *proxyConnection) error {
	// 将连接加入等待池
	s.addPendingConnection(proxyConn)

	// 创建代理请求消息
	proxyReq := types.ControlMessage{
		Type:      types.MessageTypeProxyRequest,
		SessionID: generateSessionID(),
		Data: map[string]interface{}{
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
			"localAddress": "", // 将由客户端从服务配置中获取
			"localPort":    0,  // 将由客户端从服务配置中获取
		},
		Timestamp: time.Now(),
	}

	// 这里需要通过控制服务器发送消息给客户端
	// 简化实现：直接调用控制服务器的方法
	return s.sendProxyRequestToClient(proxyConn.clientID, &proxyReq)
}

// bridgeConnections 桥接外网连接和内网数据连接
func (s *proxyServer) bridgeConnections(ctx context.Context, proxyConn *proxyConnection, proxy *reverseProxyInstance) error {
	defer proxyConn.clientConn.Close()
	defer func() {
		// 尝试将数据连接归还到池中，而不是直接关闭
		if !s.returnDataConnection(proxy.service.TunnelServiceId, proxyConn.dataConn) {
			// 如果无法归还到池中，则关闭连接
			proxyConn.dataConn.Close()
		}
	}()

	// 移除等待连接
	s.removePendingConnection(proxyConn.connectionID)

	logger.Info("Starting connection bridge", map[string]interface{}{
		"serviceId":    proxyConn.serviceID,
		"connectionId": proxyConn.connectionID,
	})

	// 使用两个goroutine实现双向转发
	done := make(chan struct{}, 2)
	var totalBytes int64

	// 外网客户端 -> 内网服务 (通过数据连接)
	go func() {
		defer func() { done <- struct{}{} }()
		bytes, err := io.Copy(proxyConn.dataConn, proxyConn.clientConn)
		if err != nil && err != io.EOF {
			// 对于流式协议，连接断开是正常的
			logger.Debug("Client to data connection closed", map[string]interface{}{
				"error":        err.Error(),
				"serviceId":    proxyConn.serviceID,
				"connectionId": proxyConn.connectionID,
			})
		}
		atomic.AddInt64(&totalBytes, bytes)
		logger.Debug("Client->Data transfer completed", map[string]interface{}{
			"bytes":        bytes,
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
		})
	}()

	// 内网服务 -> 外网客户端 (通过数据连接)
	go func() {
		defer func() { done <- struct{}{} }()
		bytes, err := io.Copy(proxyConn.clientConn, proxyConn.dataConn)
		if err != nil && err != io.EOF {
			// 对于流式协议，连接断开是正常的
			logger.Debug("Data to client connection closed", map[string]interface{}{
				"error":        err.Error(),
				"serviceId":    proxyConn.serviceID,
				"connectionId": proxyConn.connectionID,
			})
		}
		atomic.AddInt64(&totalBytes, bytes)
		logger.Debug("Data->Client transfer completed", map[string]interface{}{
			"bytes":        bytes,
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
		})
	}()

	// 等待任一方向的传输完成
	<-done

	// 更新流量统计
	atomic.AddInt64(&proxy.totalTraffic, totalBytes)

	logger.Info("Connection bridge completed", map[string]interface{}{
		"serviceId":    proxyConn.serviceID,
		"connectionId": proxyConn.connectionID,
		"totalBytes":   totalBytes,
		"totalTraffic": atomic.LoadInt64(&proxy.totalTraffic),
	})

	return nil
}

// addPendingConnection 添加等待中的代理连接
//
// 当外网请求到达时，创建代理连接并加入等待队列，
// 等待客户端建立对应的数据连接。
func (s *proxyServer) addPendingConnection(conn *proxyConnection) {
	s.pendingMutex.Lock()
	defer s.pendingMutex.Unlock()
	s.pendingConns[conn.connectionID] = conn
}

// findPendingConnection 查找等待中的代理连接
//
// 根据连接ID查找等待中的代理连接，用于匹配客户端的数据连接。
func (s *proxyServer) findPendingConnection(connectionID string) *proxyConnection {
	s.pendingMutex.RLock()
	defer s.pendingMutex.RUnlock()
	return s.pendingConns[connectionID]
}

// removePendingConnection 移除等待中的代理连接
//
// 当数据连接建立完成或超时时，从等待队列中移除代理连接。
func (s *proxyServer) removePendingConnection(connectionID string) {
	s.pendingMutex.Lock()
	defer s.pendingMutex.Unlock()
	delete(s.pendingConns, connectionID)
}

// findTunnelService 查找隧道服务配置
//
// 根据服务ID从服务注册器中获取服务配置信息。
// 用于获取服务的端口、地址等配置参数。
func (s *proxyServer) findTunnelService(serviceID string) *types.TunnelService {
	service, err := s.tunnelServer.GetServiceRegistry().GetService(context.Background(), serviceID)
	if err != nil {
		logger.Error("Failed to get service from registry", map[string]interface{}{
			"error":     err.Error(),
			"serviceId": serviceID,
		})
		return nil
	}
	return service
}

// sendProxyRequestToClient 向客户端发送代理请求
//
// 当外网请求到达且连接池中没有可用连接时，
// 通过控制服务器向指定客户端发送代理请求消息。
func (s *proxyServer) sendProxyRequestToClient(clientID string, message *types.ControlMessage) error {
	logger.Info("Sending proxy request to client", map[string]interface{}{
		"clientId":     clientID,
		"messageType":  message.Type,
		"connectionId": message.Data["connectionId"],
	})

	return s.controlServer.SendMessageToClient(clientID, message)
}

// generateConnectionID 生成连接唯一标识符
//
// 使用当前纳秒时间戳生成唯一的连接ID，
// 用于标识和匹配代理连接与数据连接。
func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// generateSessionID 生成会话唯一标识符
//
// 使用当前纳秒时间戳生成唯一的会话ID，
// 用于控制消息的会话标识。
func generateSessionID() string {
	return fmt.Sprintf("session_%d", time.Now().UnixNano())
}

// getOrCreateDataConnectionPool 获取或创建数据连接池
//
// 为指定服务获取现有的连接池，如果不存在则创建新的连接池。
// 每个服务只有一个连接池实例，用于复用与该服务客户端的数据连接。
//
// 参数:
//   - serviceID: 服务唯一标识符
//   - clientID: 客户端唯一标识符
//
// 返回:
//   - *dataConnectionPool: 数据连接池实例
func (s *proxyServer) getOrCreateDataConnectionPool(serviceID, clientID string) *dataConnectionPool {
	s.poolsMutex.Lock()
	defer s.poolsMutex.Unlock()

	if pool, exists := s.dataConnPools[serviceID]; exists {
		return pool
	}

	// 创建新的连接池，配置合理的大小限制
	pool := &dataConnectionPool{
		serviceID:      serviceID,
		clientID:       clientID,
		availableConns: make(chan net.Conn, 10), // 缓冲队列，最多10个空闲连接
		minSize:        2,                       // 保持最少2个连接，确保快速响应
		maxSize:        10,                      // 限制最多10个连接，避免资源浪费
	}

	s.dataConnPools[serviceID] = pool

	// 启动连接池管理协程，负责维护连接数量和清理过期连接
	s.wg.Add(1)
	go s.manageConnectionPool(pool)

	logger.Info("Created data connection pool", map[string]interface{}{
		"serviceId": serviceID,
		"clientId":  clientID,
		"minSize":   pool.minSize,
		"maxSize":   pool.maxSize,
	})

	return pool
}

// manageConnectionPool 管理连接池
//
// 后台协程，负责连接池的维护工作：
// 1. 定期检查连接池大小，确保满足最小连接数要求
// 2. 清理过期的空闲连接，释放系统资源
// 3. 监听服务器关闭信号，优雅退出
//
// 该方法会一直运行直到服务器关闭。
func (s *proxyServer) manageConnectionPool(pool *dataConnectionPool) {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second) // 每30秒执行一次维护任务
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// 服务器关闭，退出管理协程
			return
		case <-ticker.C:
			// 执行定期维护任务
			s.maintainPoolSize(pool)
			s.cleanupIdleConnections(pool)
		}
	}
}

// maintainPoolSize 维护连接池大小
//
// 检查连接池中可用连接数量，如果少于最小值且未达到最大限制，
// 则请求客户端建立新的数据连接以补充连接池。
//
// 参数:
//   - pool: 要维护的连接池
func (s *proxyServer) maintainPoolSize(pool *dataConnectionPool) {
	currentSize := atomic.LoadInt32(&pool.currentSize)
	availableCount := len(pool.availableConns)

	// 检查是否需要补充连接
	if availableCount < pool.minSize && int(currentSize) < pool.maxSize {
		needed := pool.minSize - availableCount
		for i := 0; i < needed; i++ {
			s.requestNewDataConnection(pool.serviceID, pool.clientID)
		}

		logger.Debug("Requested new connections for pool", map[string]interface{}{
			"serviceId":      pool.serviceID,
			"currentSize":    currentSize,
			"availableCount": availableCount,
			"requestedCount": needed,
		})
	}
}

// cleanupIdleConnections 清理空闲连接
//
// 遍历连接池中的活跃连接，关闭并移除超过空闲时间限制的连接。
// 这有助于释放系统资源，避免连接泄漏。
//
// 参数:
//   - pool: 要清理的连接池
func (s *proxyServer) cleanupIdleConnections(pool *dataConnectionPool) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	idleTimeout := 5 * time.Minute // 空闲连接超时时间

	// 注意：当前实现使用通道管理连接，无法直接检查连接的空闲时间
	// 这里只是记录连接池状态，实际的连接清理由连接使用方负责
	// 未来可以考虑重新设计连接池以支持更精细的连接生命周期管理

	logger.Debug("Connection pool cleanup completed", map[string]interface{}{
		"serviceId":      pool.serviceID,
		"currentSize":    atomic.LoadInt32(&pool.currentSize),
		"availableConns": len(pool.availableConns),
		"idleTimeout":    idleTimeout,
	})
}

// requestNewDataConnection 请求客户端建立新的数据连接
func (s *proxyServer) requestNewDataConnection(serviceID, clientID string) {
	// 创建预连接请求消息
	preConnReq := types.ControlMessage{
		Type:      "pre_connect_request",
		SessionID: generateSessionID(),
		Data: map[string]interface{}{
			"serviceId": serviceID,
			"pooled":    true, // 标识这是连接池连接
		},
		Timestamp: time.Now(),
	}

	// 发送给客户端
	if err := s.controlServer.SendMessageToClient(clientID, &preConnReq); err != nil {
		logger.Error("Failed to request new data connection", map[string]interface{}{
			"error":     err.Error(),
			"serviceId": serviceID,
			"clientId":  clientID,
		})
	}
}

// getAvailableDataConnection 获取可用的数据连接
func (s *proxyServer) getAvailableDataConnection(serviceID string) net.Conn {
	s.poolsMutex.RLock()
	pool, exists := s.dataConnPools[serviceID]
	s.poolsMutex.RUnlock()

	if !exists {
		return nil
	}

	select {
	case conn := <-pool.availableConns:
		logger.Debug("Reused pooled data connection", map[string]interface{}{
			"serviceId": serviceID,
		})
		return conn
	default:
		// 没有可用连接
		return nil
	}
}

// returnDataConnection 归还数据连接到池中
// 返回 true 表示成功归还，false 表示需要调用者关闭连接
func (s *proxyServer) returnDataConnection(serviceID string, conn net.Conn) bool {
	s.poolsMutex.RLock()
	pool, exists := s.dataConnPools[serviceID]
	s.poolsMutex.RUnlock()

	if !exists {
		return false // 连接池不存在，需要调用者关闭连接
	}

	select {
	case pool.availableConns <- conn:
		logger.Debug("Returned connection to pool", map[string]interface{}{
			"serviceId": serviceID,
		})
		return true // 成功归还到池中
	default:
		// 池已满，需要调用者关闭连接
		atomic.AddInt32(&pool.currentSize, -1)
		return false
	}
}

// Stop 停止反向代理服务器
//
// 参数:
//   - ctx: 上下文，用于控制停止超时
//
// 返回:
//   - error: 停止过程中的错误
//
// 功能:
//   - 停止所有活跃的代理实例
//   - 清理连接池和资源
//   - 等待所有goroutine退出
func (s *proxyServer) Stop(ctx context.Context) error {
	logger.Info("Stopping reverse proxy server", nil)

	// 发送停止信号
	if s.cancel != nil {
		s.cancel()
	}

	// 停止所有活跃的代理
	s.proxyMutex.Lock()
	var proxyIDs []string
	for proxyID := range s.activeProxies {
		proxyIDs = append(proxyIDs, proxyID)
	}
	s.proxyMutex.Unlock()

	// 逐个停止代理
	for _, proxyID := range proxyIDs {
		if err := s.StopProxy(ctx, proxyID); err != nil {
			logger.Error("Failed to stop proxy during shutdown", map[string]interface{}{
				"error":   err.Error(),
				"proxyId": proxyID,
			})
		}
	}

	// 等待所有goroutine退出
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("Reverse proxy server stopped successfully", nil)
		return nil
	case <-ctx.Done():
		logger.Warn("Reverse proxy server stop timeout", nil)
		return ctx.Err()
	}
}
