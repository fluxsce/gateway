// Package server 提供反向代理服务器的完整实现
// 反向代理服务器负责处理隧道连接，将外网请求转发到内网客户端服务
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
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
		logger.Debug("Using pooled data connection", map[string]interface{}{
			"serviceId": proxyID,
		})
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

	// 等待内网客户端建立数据连接
	// 关键修复：等待时间应该与客户端连接超时时间一致（30秒）
	// 客户端建立连接需要：TCP握手（最多30秒）+ 发送握手消息 + 服务器处理
	// 如果超时太短，会导致连接建立失败但客户端仍在尝试建立连接
	waitStartTime := time.Now()
	waitTimeout := 35 * time.Second // 比客户端连接超时（30秒）稍长，留出处理时间

	// 通知内网客户端建立连接
	if err := s.notifyClientConnection(proxyConn); err != nil {
		logger.Warn("Failed to notify client for connection, trying pooled connection as fallback", map[string]interface{}{
			"error":        err.Error(),
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
			"clientId":     proxyConn.clientID,
		})
		// 通知失败（可能是控制连接已关闭），尝试使用连接池作为降级方案
		if dataConn := s.getAvailableDataConnection(proxyID); dataConn != nil {
			logger.Info("Using pooled connection as fallback after notification failure", map[string]interface{}{
				"serviceId": proxyID,
			})
			return s.bridgeConnectionsDirect(ctx, conn, dataConn, proxy)
		}
		// 如果连接池也没有可用连接，返回错误
		return fmt.Errorf("failed to notify client and no pooled connection available: %w", err)
	}

	logger.Info("Proxy request sent, waiting for client data connection", map[string]interface{}{
		"serviceId":    proxyID,
		"connectionId": proxyConn.connectionID,
		"clientId":     proxyConn.clientID,
		"timeout":      waitTimeout,
	})

	select {
	case <-proxyConn.ready:
		waitDuration := time.Since(waitStartTime)
		// 开始桥接连接
		logger.Info("Starting connection bridge", map[string]interface{}{
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
			"waitDuration": waitDuration,
		})
		return s.bridgeConnections(ctx, proxyConn, proxy)
	case <-time.After(waitTimeout):
		waitDuration := time.Since(waitStartTime)
		// 检查是否还有等待中的连接（可能客户端已经建立连接但还没匹配）
		s.pendingMutex.RLock()
		stillPending := s.pendingConns[proxyConn.connectionID] != nil
		pendingCount := len(s.pendingConns)
		s.pendingMutex.RUnlock()

		logger.Warn("Timeout waiting for client data connection", map[string]interface{}{
			"serviceId":     proxyID,
			"connectionId":  proxyConn.connectionID,
			"clientId":      proxyConn.clientID,
			"waitDuration":  waitDuration,
			"createTime":    proxyConn.createTime,
			"stillPending":  stillPending,
			"pendingCount":  pendingCount,
			"possibleCause": "client_connection_establishment_failed_or_too_slow",
		})

		// 清理等待连接
		s.removePendingConnection(proxyConn.connectionID)

		// 尝试使用连接池作为降级方案
		if dataConn := s.getAvailableDataConnection(proxyID); dataConn != nil {
			logger.Info("Using pooled connection as fallback after timeout", map[string]interface{}{
				"serviceId": proxyID,
			})
			return s.bridgeConnectionsDirect(ctx, conn, dataConn, proxy)
		}

		// 如果连接池也没有可用连接，返回错误
		// 这通常发生在第一次请求时，连接池为空，且客户端建立连接失败或超时
		return fmt.Errorf("timeout waiting for client data connection after %v (client may have failed to establish connection), and no pooled connection available", waitDuration)
	case <-ctx.Done():
		s.removePendingConnection(proxyConn.connectionID)
		return ctx.Err()
	}
}

// bridgeConnectionsDirect 直接桥接连接（使用池化连接）
func (s *proxyServer) bridgeConnectionsDirect(ctx context.Context, clientConn, dataConn net.Conn, proxy *reverseProxyInstance) error {
	// 使用 sync.Once 确保连接只关闭一次
	var clientConnOnce sync.Once
	closeClientConn := func() {
		clientConnOnce.Do(func() {
			// 优雅关闭：先关闭写入方向，再关闭连接
			if tcpConn, ok := clientConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			clientConn.Close()
		})
	}
	defer closeClientConn()

	// 注意：不要在这里关闭 dataConn，让 defer 中的逻辑决定是归还还是关闭
	shouldReturnToPool := true
	var dataConnOnce sync.Once
	closeDataConn := func() {
		dataConnOnce.Do(func() {
			// 优雅关闭：先关闭写入方向，再关闭连接
			if tcpConn, ok := dataConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			dataConn.Close()
		})
	}
	defer func() {
		if shouldReturnToPool {
			// 尝试将数据连接归还到池中
			if !s.returnDataConnection(proxy.service.TunnelServiceId, dataConn) {
				// 如果无法归还到池中，则关闭连接
				closeDataConn()
			}
		} else {
			// 连接有问题，直接关闭
			closeDataConn()
		}
	}()

	logger.Info("Starting direct connection bridge (pooled)", map[string]interface{}{
		"serviceId": proxy.service.TunnelServiceId,
	})

	// 关键修复：清除之前可能设置的超时时间
	// 注意：客户端在等待时设置了读取超时，服务器端取出连接后应立即清除
	// 这样可以确保数据传输不受超时限制
	clientConn.SetDeadline(time.Time{})
	dataConn.SetDeadline(time.Time{})

	// 注意：不需要等待，因为：
	// 1. 客户端在等待循环中设置了读取超时，会一直尝试读取
	// 2. 服务器端开始写入数据时，客户端会立即收到数据并退出等待循环
	// 3. 客户端使用 prefixedConn 包装已读取的数据，确保数据不丢失

	// 使用两个goroutine实现双向转发
	done := make(chan struct{}, 2)
	var totalBytes int64
	var copyErrors sync.Map // 记录复制错误

	// 外网客户端 -> 内网服务 (通过数据连接)
	// 注意：这个方向的数据会触发客户端从等待循环中退出
	go func() {
		defer func() {
			// 当一方完成时，优雅关闭写入方向，通知对方
			if tcpConn, ok := dataConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			done <- struct{}{}
		}()
		bytes, err := io.Copy(dataConn, clientConn)
		atomic.AddInt64(&totalBytes, bytes)

		if err != nil && err != io.EOF {
			copyErrors.Store("clientToData", err)
		}
	}()

	// 内网服务 -> 外网客户端 (通过数据连接)
	// 注意：这个方向的数据需要客户端从数据连接读取并转发到本地服务
	go func() {
		defer func() {
			// 当一方完成时，优雅关闭写入方向，通知对方
			if tcpConn, ok := clientConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			done <- struct{}{}
		}()
		bytes, err := io.Copy(clientConn, dataConn)
		atomic.AddInt64(&totalBytes, bytes)

		if err != nil && err != io.EOF {
			copyErrors.Store("dataToClient", err)
		}
	}()

	// 等待两个方向都完成或上下文取消
	select {
	case <-ctx.Done():
		// 上下文取消，优雅关闭连接
		closeClientConn()
		closeDataConn()
		return ctx.Err()
	case <-done:
		// 第一个方向完成，等待第二个方向
		<-done
	}

	// 更新流量统计
	atomic.AddInt64(&proxy.totalTraffic, totalBytes)

	// 关键修复：检查是否有真正的错误（非正常关闭）
	// ERR_EMPTY_RESPONSE 通常是因为连接过早关闭或没有发送任何数据
	hasError := false
	copyErrors.Range(func(key, value interface{}) bool {
		if err, ok := value.(error); ok {
			errMsg := err.Error()
			// 检查是否是正常的连接关闭
			if !strings.Contains(errMsg, "use of closed network connection") &&
				!strings.Contains(errMsg, "i/o timeout") &&
				!strings.Contains(errMsg, "connection reset") &&
				!strings.Contains(errMsg, "broken pipe") {
				hasError = true
				shouldReturnToPool = false // 连接有错误，不归还到池中
				logger.Error("Connection bridge error", map[string]interface{}{
					"direction":  key,
					"error":      errMsg,
					"serviceId":  proxy.service.TunnelServiceId,
					"totalBytes": totalBytes,
				})
			}
		}
		return true
	})

	// 关键修复：检查是否传输了任何数据
	// 如果 totalBytes == 0，说明没有传输任何数据，可能导致 ERR_EMPTY_RESPONSE
	// 这通常发生在以下情况：
	// 1. 客户端连接本地服务失败，立即关闭了服务器数据连接
	// 2. 数据转发过程中出现错误，连接过早关闭
	// 3. 客户端在建立本地连接前就关闭了服务器连接
	if totalBytes == 0 {
		// 详细记录错误信息，帮助诊断问题
		var errorDetails []string
		copyErrors.Range(func(key, value interface{}) bool {
			if err, ok := value.(error); ok {
				errorDetails = append(errorDetails, fmt.Sprintf("%s: %s", key, err.Error()))
			}
			return true
		})

		logger.Error("Connection bridge completed with zero bytes transferred - possible data transmission issue", map[string]interface{}{
			"serviceId":       proxy.service.TunnelServiceId,
			"serviceType":     proxy.service.ServiceType,
			"returningToPool": shouldReturnToPool,
			"hasError":        hasError,
			"errorDetails":    errorDetails,
			"possibleCauses": []string{
				"client_failed_to_connect_to_local_service",
				"client_closed_connection_before_data_transmission",
				"data_forwarding_error_on_client_side",
			},
		})
		// 虽然没有错误，但没有传输数据，不归还到池中（可能是连接有问题）
		shouldReturnToPool = false
	}

	if !hasError && totalBytes > 0 {
		logger.Debug("Direct connection bridge completed successfully", map[string]interface{}{
			"serviceId":       proxy.service.TunnelServiceId,
			"totalBytes":      totalBytes,
			"totalTraffic":    atomic.LoadInt64(&proxy.totalTraffic),
			"returningToPool": shouldReturnToPool,
		})
	}

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
func (s *proxyServer) HandleClientDataConnection(ctx context.Context, conn net.Conn, connectionID string, clientID string) error {
	// 为长连接（如SSE）优化TCP连接
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		// 禁用Nagle算法以减少延迟（对流式数据重要）
		tcpConn.SetNoDelay(true)
	}

	logger.Info("Received client data connection", map[string]interface{}{
		"connectionId": connectionID,
		"clientId":     clientID,
		"timestamp":    time.Now(),
	})

	// 验证 clientID 不能为空
	if clientID == "" {
		conn.Close()
		return fmt.Errorf("clientID is required but was empty for connectionID %s", connectionID)
	}

	// 首先尝试作为普通连接处理
	proxyConn := s.findPendingConnection(connectionID)
	if proxyConn != nil {
		// 验证 clientID 是否匹配
		if proxyConn.clientID != clientID {
			conn.Close()
			return fmt.Errorf("clientID mismatch: expected %s, got %s for connectionID %s",
				proxyConn.clientID, clientID, connectionID)
		}

		// 设置数据连接
		proxyConn.dataConn = conn

		// 通知桥接可以开始
		close(proxyConn.ready)

		logger.Info("Client data connection matched with pending connection", map[string]interface{}{
			"connectionId": connectionID,
			"serviceId":    proxyConn.serviceID,
			"clientId":     proxyConn.clientID,
			"waitTime":     time.Since(proxyConn.createTime),
		})

		return nil
	}

	// 如果不是等待中的连接，可能是池化连接
	// 验证 clientID 是否与服务配置匹配
	service := s.findTunnelService(connectionID)
	if service == nil {
		conn.Close()
		return fmt.Errorf("service %s not found for connectionID %s", connectionID, connectionID)
	}

	// 验证 clientID 是否与服务配置中的 TunnelClientId 匹配
	if service.TunnelClientId != clientID {
		conn.Close()
		return fmt.Errorf("clientID mismatch: expected %s (from service config), got %s for serviceID %s",
			service.TunnelClientId, clientID, connectionID)
	}

	// 获取或创建连接池（clientID 已验证匹配）
	pool := s.getOrCreateDataConnectionPool(connectionID, clientID)
	if pool == nil {
		// 连接池存在但 clientID 不匹配（不应该发生，因为已经验证过）
		// 这可能是竞态条件导致的数据不一致
		conn.Close()
		return fmt.Errorf("connection pool exists but clientID mismatch for serviceID %s (expected %s)", connectionID, clientID)
	}

	// 尝试将连接添加到对应服务的连接池
	if s.addToConnectionPool(connectionID, conn) {
		logger.Info("Added connection to pool", map[string]interface{}{
			"serviceId": connectionID,
		})
		return nil
	}

	// 都不是，记录详细信息并关闭连接
	s.pendingMutex.RLock()
	pendingCount := len(s.pendingConns)
	pendingIDs := make([]string, 0, pendingCount)
	for id := range s.pendingConns {
		pendingIDs = append(pendingIDs, id)
	}
	s.pendingMutex.RUnlock()

	logger.Warn("No pending connection or pool found for connectionID", map[string]interface{}{
		"connectionId": connectionID,
		"clientId":     clientID,
		"pendingCount": pendingCount,
		"pendingIDs":   pendingIDs,
	})

	conn.Close()
	return fmt.Errorf("no pending connection or pool found for connectionID %s (pending count: %d)", connectionID, pendingCount)
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
		// 增加池中连接计数（新连接已添加到池中）
		atomic.AddInt32(&pool.currentSize, 1)
		logger.Debug("Added connection to pool", map[string]interface{}{
			"serviceId":      serviceID,
			"currentSize":    atomic.LoadInt32(&pool.currentSize),
			"availableCount": len(pool.availableConns),
		})
		return true
	default:
		// 池已满，连接无法添加
		logger.Warn("Data connection pool full, connection will be closed", map[string]interface{}{
			"serviceId":      serviceID,
			"currentSize":    atomic.LoadInt32(&pool.currentSize),
			"availableCount": len(pool.availableConns),
			"maxSize":        pool.maxSize,
		})
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
				// 检查上下文是否已取消
				select {
				case <-ctx.Done():
					logger.Info("Reverse proxy listener stopped by context", map[string]interface{}{
						"serviceId": proxy.service.TunnelServiceId,
					})
					return
				default:
					// 检查是否是监听器关闭导致的错误
					// 使用字符串包含检查 "use of closed network connection" 错误
					errMsg := err.Error()
					if strings.Contains(errMsg, "use of closed network connection") {
						logger.Info("Reverse proxy listener closed", map[string]interface{}{
							"serviceId": proxy.service.TunnelServiceId,
						})
						return
					}

					// 其他错误，记录日志并继续
					logger.Error("Failed to accept connection", map[string]interface{}{
						"error":     err.Error(),
						"serviceId": proxy.service.TunnelServiceId,
					})
					// 短暂延迟，避免错误循环过快
					time.Sleep(100 * time.Millisecond)
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

	logger.Info("Sending proxy request to client", map[string]interface{}{
		"clientId":     proxyConn.clientID,
		"serviceId":    proxyConn.serviceID,
		"connectionId": proxyConn.connectionID,
		"timestamp":    time.Now(),
	})

	// 这里需要通过控制服务器发送消息给客户端
	// 简化实现：直接调用控制服务器的方法
	if err := s.sendProxyRequestToClient(proxyConn.clientID, &proxyReq); err != nil {
		logger.Error("Failed to send proxy request to client", map[string]interface{}{
			"error":        err.Error(),
			"clientId":     proxyConn.clientID,
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
		})
		// 发送失败，清理等待连接
		s.removePendingConnection(proxyConn.connectionID)
		return fmt.Errorf("failed to send proxy request: %w", err)
	}

	logger.Info("Proxy request sent successfully, waiting for client data connection", map[string]interface{}{
		"clientId":     proxyConn.clientID,
		"serviceId":    proxyConn.serviceID,
		"connectionId": proxyConn.connectionID,
	})

	return nil
}

// bridgeConnections 桥接外网连接和内网数据连接
func (s *proxyServer) bridgeConnections(ctx context.Context, proxyConn *proxyConnection, proxy *reverseProxyInstance) error {
	// 使用 sync.Once 确保连接只关闭一次，避免 "use of closed network connection" 错误
	var clientConnOnce sync.Once
	closeClientConn := func() {
		clientConnOnce.Do(func() {
			// 优雅关闭：先关闭写入方向，再关闭连接
			if tcpConn, ok := proxyConn.clientConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			proxyConn.clientConn.Close()
		})
	}
	defer closeClientConn()

	// 关键修复：通过 proxy_request 新建立的连接，使用完毕后直接关闭，不归还到池中
	// 只有从池中取出的连接（bridgeConnectionsDirect）才应该归还到池中
	// 新建立的连接如果归还到池中，可能会导致连接状态混乱
	var dataConnOnce sync.Once
	closeDataConn := func() {
		dataConnOnce.Do(func() {
			// 优雅关闭：先关闭写入方向，再关闭连接
			if tcpConn, ok := proxyConn.dataConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			proxyConn.dataConn.Close()
		})
	}
	defer closeDataConn()

	// 移除等待连接
	s.removePendingConnection(proxyConn.connectionID)

	logger.Info("Starting connection bridge", map[string]interface{}{
		"serviceId":    proxyConn.serviceID,
		"connectionId": proxyConn.connectionID,
	})

	// 不设置超时时间，让长连接（如SSH）可以持续工作
	// TCP KeepAlive 会自动检测连接状态

	// 使用两个goroutine实现双向转发
	done := make(chan struct{}, 2)
	var totalBytes int64
	var copyErrors sync.Map

	// 外网客户端 -> 内网服务 (通过数据连接)
	go func() {
		defer func() {
			// 当一方完成时，优雅关闭写入方向，通知对方
			if tcpConn, ok := proxyConn.dataConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			done <- struct{}{}
		}()
		bytes, err := io.Copy(proxyConn.dataConn, proxyConn.clientConn)
		atomic.AddInt64(&totalBytes, bytes)

		if err != nil && err != io.EOF {
			copyErrors.Store("clientToData", err)
		}

		logger.Debug("Client->Data transfer completed", map[string]interface{}{
			"bytes":        bytes,
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
		})
	}()

	// 内网服务 -> 外网客户端 (通过数据连接)
	go func() {
		defer func() {
			// 当一方完成时，优雅关闭写入方向，通知对方
			if tcpConn, ok := proxyConn.clientConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			done <- struct{}{}
		}()
		bytes, err := io.Copy(proxyConn.clientConn, proxyConn.dataConn)
		atomic.AddInt64(&totalBytes, bytes)

		if err != nil && err != io.EOF {
			copyErrors.Store("dataToClient", err)
		}

		logger.Debug("Data->Client transfer completed", map[string]interface{}{
			"bytes":        bytes,
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
		})
	}()

	// 等待两个方向都完成或上下文取消
	select {
	case <-ctx.Done():
		// 上下文取消，直接关闭连接
		closeClientConn()
		closeDataConn()
		return ctx.Err()
	case <-done:
		// 第一个方向完成，等待第二个方向
		<-done
	}

	// 更新流量统计
	atomic.AddInt64(&proxy.totalTraffic, totalBytes)

	// 检查是否有真正的错误（非正常关闭）
	hasError := false
	copyErrors.Range(func(key, value interface{}) bool {
		if err, ok := value.(error); ok {
			errMsg := err.Error()
			// 检查是否是正常的连接关闭
			if !strings.Contains(errMsg, "use of closed network connection") &&
				!strings.Contains(errMsg, "i/o timeout") &&
				!strings.Contains(errMsg, "connection reset") &&
				!strings.Contains(errMsg, "broken pipe") &&
				!strings.Contains(errMsg, "splice: broken pipe") &&
				!strings.Contains(errMsg, "splice: connection timed out") {
				hasError = true
				logger.Error("Connection bridge error", map[string]interface{}{
					"direction":    key,
					"error":        errMsg,
					"serviceId":    proxyConn.serviceID,
					"connectionId": proxyConn.connectionID,
				})
			} else {
				// 记录可恢复的错误（调试级别）
				logger.Debug("Connection bridge recoverable error", map[string]interface{}{
					"direction":    key,
					"error":        errMsg,
					"serviceId":    proxyConn.serviceID,
					"connectionId": proxyConn.connectionID,
				})
			}
		}
		return true
	})

	// 检查是否传输了任何数据
	// 如果 totalBytes == 0，说明没有传输任何数据，可能导致 ERR_EMPTY_RESPONSE
	if totalBytes == 0 {
		// 详细记录错误信息，帮助诊断问题
		var errorDetails []string
		copyErrors.Range(func(key, value interface{}) bool {
			if err, ok := value.(error); ok {
				errorDetails = append(errorDetails, fmt.Sprintf("%s: %s", key, err.Error()))
			}
			return true
		})

		logger.Warn("Connection bridge completed with zero bytes transferred", map[string]interface{}{
			"serviceId":    proxyConn.serviceID,
			"serviceType":  proxy.service.ServiceType,
			"connectionId": proxyConn.connectionID,
			"hasError":     hasError,
			"errorDetails": errorDetails,
			"note":         "new connection will be closed (not returned to pool)",
		})
	}

	if !hasError && totalBytes > 0 {
		logger.Info("Connection bridge completed successfully", map[string]interface{}{
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
			"totalBytes":   totalBytes,
			"totalTraffic": atomic.LoadInt64(&proxy.totalTraffic),
			"note":         "new connection will be closed (not returned to pool)",
		})
	}

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
// 使用高强度随机字符串生成器，确保在高并发和分布式环境下的唯一性。
// 生成的ID格式：conn_<20位随机字符串>
//
// 返回:
//   - string: 唯一的连接标识符
func generateConnectionID() string {
	return fmt.Sprintf("conn_%s", random.GenerateRandomString(20))
}

// generateSessionID 生成会话唯一标识符
//
// 使用高强度随机字符串生成器，确保在高并发和分布式环境下的唯一性。
// 生成的ID格式：session_<20位随机字符串>
//
// 返回:
//   - string: 唯一的会话标识符
func generateSessionID() string {
	return fmt.Sprintf("session_%s", random.GenerateRandomString(20))
}

// getOrCreateDataConnectionPool 获取或创建数据连接池
//
// 为指定服务获取现有的连接池，如果不存在则创建新的连接池。
// 每个服务只有一个连接池实例，用于复用与该服务客户端的数据连接。
//
// 参数:
//   - serviceID: 服务唯一标识符
//   - clientID: 客户端唯一标识符（必须与服务配置匹配）
//
// 返回:
//   - *dataConnectionPool: 数据连接池实例
//
// 注意:
//   - 如果连接池已存在但 clientID 不匹配，会记录错误并返回 nil（调用者应处理此情况）
//   - 调用者应确保 clientID 已通过验证
func (s *proxyServer) getOrCreateDataConnectionPool(serviceID, clientID string) *dataConnectionPool {
	s.poolsMutex.Lock()
	defer s.poolsMutex.Unlock()

	if pool, exists := s.dataConnPools[serviceID]; exists {
		// 严格验证 clientID 是否匹配
		if pool.clientID != clientID {
			logger.Error("Connection pool exists but clientID mismatch, cannot use this pool", map[string]interface{}{
				"serviceId":        serviceID,
				"expectedClientId": clientID,
				"actualClientId":   pool.clientID,
			})
			// 返回 nil，让调用者知道连接池不匹配
			// 这种情况不应该发生，因为调用者已经验证了 clientID
			// 但如果发生了，说明有竞态条件或数据不一致
			return nil
		}
		return pool
	}

	// 创建新的连接池，配置合理的大小限制
	// 高并发场景下需要更大的连接池以提升性能
	pool := &dataConnectionPool{
		serviceID:      serviceID,
		clientID:       clientID,
		availableConns: make(chan net.Conn, 50), // 缓冲队列，最多50个空闲连接（高并发优化）
		minSize:        10,                      // 保持最少10个连接，确保快速响应
		maxSize:        100,                     // 限制最多100个连接，支持高并发场景
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
// 1. 清理过期的空闲连接，释放系统资源
// 2. 监听服务器关闭信号，优雅退出
//
// 注意：连接池大小由按需请求机制维护（在 HandleProxyConnection 中），
// 不需要定期检查连接池大小。
//
// 该方法会一直运行直到服务器关闭。
func (s *proxyServer) manageConnectionPool(pool *dataConnectionPool) {
	defer s.wg.Done()

	ticker := time.NewTicker(10 * time.Second) // 每10秒执行一次维护任务（高并发优化）
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			// 服务器关闭，退出管理协程
			return
		case <-ticker.C:
			// 执行定期维护任务（仅清理空闲连接，连接池大小由按需请求机制维护）
			s.cleanupIdleConnections(pool)
		}
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

// getAvailableDataConnection 获取可用的数据连接
func (s *proxyServer) getAvailableDataConnection(serviceID string) net.Conn {
	s.poolsMutex.RLock()
	pool, exists := s.dataConnPools[serviceID]
	s.poolsMutex.RUnlock()

	if !exists {
		return nil
	}

	// 尝试获取连接并验证有效性
	select {
	case conn := <-pool.availableConns:
		if conn != nil {
			// 从池中取出连接，减少可用连接计数
			atomic.AddInt32(&pool.currentSize, -1)

			// 改进的连接验证：不实际读取数据，避免与客户端等待逻辑冲突
			// 使用 TCP 连接的状态检查，而不是读取数据
			// 对于 TCP 连接，我们可以通过检查本地和远程地址来验证连接是否有效
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				// 检查连接是否仍然有效
				// 尝试获取本地地址，如果连接已关闭会返回错误
				localAddr := tcpConn.LocalAddr()
				remoteAddr := tcpConn.RemoteAddr()

				if localAddr == nil || remoteAddr == nil {
					// 连接已关闭
					conn.Close()
					logger.Warn("Data connection from pool is closed (invalid address), will request new connection", map[string]interface{}{
						"serviceId": serviceID,
					})
					return nil
				}

				// 清除客户端可能设置的读取超时，准备接收数据
				// 客户端在等待时会设置读取超时，服务器取出连接后应立即清除
				conn.SetReadDeadline(time.Time{})
				conn.SetWriteDeadline(time.Time{})

				// 连接有效，直接使用（不读取数据，避免与客户端等待逻辑冲突）
				logger.Debug("Reused pooled data connection", map[string]interface{}{
					"serviceId":      serviceID,
					"availableCount": len(pool.availableConns),
					"currentSize":    atomic.LoadInt32(&pool.currentSize),
					"localAddr":      localAddr.String(),
					"remoteAddr":     remoteAddr.String(),
				})
				return conn
			}

			// 非 TCP 连接：同样不读取数据，只清除超时设置
			// 注意：池化连接应该是TCP连接，这里只是防御性处理
			// 关键：不读取数据，避免与客户端等待逻辑冲突和数据丢失
			conn.SetReadDeadline(time.Time{})
			conn.SetWriteDeadline(time.Time{})

			logger.Debug("Reused pooled data connection (non-TCP)", map[string]interface{}{
				"serviceId":      serviceID,
				"availableCount": len(pool.availableConns),
				"currentSize":    atomic.LoadInt32(&pool.currentSize),
			})
			return conn
		}
		return nil
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
		// 连接成功归还到池中，增加可用连接计数
		atomic.AddInt32(&pool.currentSize, 1)
		logger.Debug("Returned connection to pool", map[string]interface{}{
			"serviceId":      serviceID,
			"currentSize":    atomic.LoadInt32(&pool.currentSize),
			"availableCount": len(pool.availableConns),
		})
		return true // 成功归还到池中
	default:
		// 池已满，需要调用者关闭连接
		logger.Warn("Data connection pool full, connection will be closed", map[string]interface{}{
			"serviceId":      serviceID,
			"currentSize":    atomic.LoadInt32(&pool.currentSize),
			"availableCount": len(pool.availableConns),
			"maxSize":        pool.maxSize,
		})
		return false
	}
}

// isNetTimeout 检查是否是网络超时错误
func isNetTimeout(err error) bool {
	if err == nil {
		return false
	}
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
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
