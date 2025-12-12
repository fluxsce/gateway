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
// 2. 将外网请求通过数据连接转发到内网服务
type proxyServer struct {
	tunnelServer  TunnelServer                     // 隧道服务器引用
	controlServer ControlServer                    // 控制服务器引用，用于与客户端通信
	activeProxies map[string]*reverseProxyInstance // serviceID -> 反向代理实例
	proxyMutex    sync.RWMutex                     // 保护 activeProxies 的读写锁
	pendingConns  map[string]*proxyConnection      // connectionID -> 等待中的代理连接
	pendingMutex  sync.RWMutex                     // 保护 pendingConns 的读写锁
	ctx           context.Context                  // 服务器上下文
	cancel        context.CancelFunc               // 取消函数
	wg            sync.WaitGroup                   // 等待组，用于优雅关闭
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
func NewProxyServerImpl(tunnelServer TunnelServer, controlServer ControlServer) ProxyServer {
	ctx, cancel := context.WithCancel(context.Background())
	return &proxyServer{
		tunnelServer:  tunnelServer,
		controlServer: controlServer,
		activeProxies: make(map[string]*reverseProxyInstance),
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
//   - 管理统计信息
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
//   - 通知内网客户端建立新连接
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

	// 创建代理连接
	proxyConn := &proxyConnection{
		clientConn:   conn,
		connectionID: generateConnectionID(),
		serviceID:    proxyID,
		clientID:     proxy.service.TunnelClientId,
		createTime:   time.Now(),
		ready:        make(chan struct{}),
	}

	// 等待内网客户端建立数据连接
	// 关键修复：等待时间应该与客户端连接超时时间一致（60秒）
	// 客户端建立连接需要：TCP握手（最多60秒）+ 发送握手消息 + 服务器处理
	// 如果超时太短，会导致连接建立失败但客户端仍在尝试建立连接
	waitStartTime := time.Now()
	waitTimeout := 65 * time.Second // 比客户端连接超时（60秒）稍长，留出处理时间

	// 通知内网客户端建立连接
	if err := s.notifyClientConnection(proxyConn); err != nil {
		logger.Warn("Failed to notify client for connection", map[string]interface{}{
			"error":        err.Error(),
			"serviceId":    proxyID,
			"connectionId": proxyConn.connectionID,
			"clientId":     proxyConn.clientID,
		})
		return fmt.Errorf("failed to notify client: %w", err)
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

		return fmt.Errorf("timeout waiting for client data connection after %v (client may have failed to establish connection)", waitDuration)
	case <-ctx.Done():
		s.removePendingConnection(proxyConn.connectionID)
		return ctx.Err()
	}
}

// HandleClientDataConnection 处理客户端数据连接
//
// 参数:
//   - ctx: 上下文
//   - conn: 客户端数据连接
//   - connectionID: 连接ID
//
// 返回:
//   - error: 处理失败时返回错误
//
// 功能:
//   - 接收客户端的数据连接
//   - 与等待的外网连接进行匹配
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

	// 查找等待中的连接
	proxyConn := s.findPendingConnection(connectionID)
	if proxyConn == nil {
		// 没有找到等待中的连接，记录详细信息并关闭连接
		s.pendingMutex.RLock()
		pendingCount := len(s.pendingConns)
		pendingIDs := make([]string, 0, pendingCount)
		for id := range s.pendingConns {
			pendingIDs = append(pendingIDs, id)
		}
		s.pendingMutex.RUnlock()

		logger.Warn("No pending connection found for connectionID", map[string]interface{}{
			"connectionId": connectionID,
			"clientId":     clientID,
			"pendingCount": pendingCount,
			"pendingIDs":   pendingIDs,
		})

		conn.Close()
		return fmt.Errorf("no pending connection found for connectionID %s (pending count: %d)", connectionID, pendingCount)
	}

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

	// 通过 proxy_request 新建立的连接，使用完毕后直接关闭
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

	// 关键修复：清除所有连接的读取/写入超时设置，确保长连接（如SSE）可以持续工作
	// 虽然数据连接在 handleDataConnection 中已经清除了超时，但 splice 系统调用可能仍然受到连接状态影响
	// splice 是 Linux 的零拷贝系统调用，io.Copy 会在两个 TCP 连接之间使用 splice
	// 如果任一连接有超时设置或连接状态异常，splice 会返回 "splice: connection timed out"
	// 因此需要确保两个连接都没有超时设置
	// 注意：外网客户端连接（clientConn）通过 listener.Accept() 接受，理论上没有设置超时
	// 但为了确保安全，也清除其超时设置
	if tcpConn, ok := proxyConn.clientConn.(*net.TCPConn); ok {
		tcpConn.SetReadDeadline(time.Time{})  // 清除读取超时
		tcpConn.SetWriteDeadline(time.Time{}) // 清除写入超时
	}
	if tcpConn, ok := proxyConn.dataConn.(*net.TCPConn); ok {
		tcpConn.SetReadDeadline(time.Time{})  // 清除读取超时
		tcpConn.SetWriteDeadline(time.Time{}) // 清除写入超时
	}
	// TCP KeepAlive 会自动检测连接状态，但不会导致 splice 超时

	// 使用两个goroutine实现双向转发
	done := make(chan struct{}, 2)
	var totalBytes int64
	var copyErrors sync.Map

	// 关键修复：使用 io.CopyBuffer 替代 io.Copy，避免 splice 系统调用的问题
	// splice 是 Linux 的零拷贝系统调用，但在某些情况下（如网络中间设备超时）会返回 "splice: connection timed out"
	// io.CopyBuffer 使用标准的 read/write 系统调用，更稳定可靠
	// 缓冲区大小：32KB，平衡性能和内存使用
	bufferSize := 32 * 1024
	buf1 := make([]byte, bufferSize)
	buf2 := make([]byte, bufferSize)

	// 外网客户端 -> 内网服务 (通过数据连接)
	go func() {
		defer func() {
			// 当一方完成时，优雅关闭写入方向，通知对方
			if tcpConn, ok := proxyConn.dataConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
			done <- struct{}{}
		}()
		// 使用 io.CopyBuffer 替代 io.Copy，避免 splice 系统调用
		bytes, err := io.CopyBuffer(proxyConn.dataConn, proxyConn.clientConn, buf1)
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
			done <- struct{}{}
		}()
		// 使用 io.CopyBuffer 替代 io.Copy，避免 splice 系统调用
		bytes, err := io.CopyBuffer(proxyConn.clientConn, proxyConn.dataConn, buf2)
		atomic.AddInt64(&totalBytes, bytes)

		// 关键修复：只有在正常结束（io.EOF）时才关闭写入方向
		// 如果是其他错误（如 broken pipe、connection reset），说明目标连接可能已经关闭
		// 此时不应该再调用 CloseWrite，避免在 HTTP chunked encoding 传输过程中提前关闭
		// HTTP chunked encoding 需要发送结束标记（0\r\n\r\n），如果写入方向被提前关闭，会导致 ERR_INCOMPLETE_CHUNKED_ENCODING
		if err == nil || err == io.EOF {
			// 正常结束，优雅关闭写入方向
			if tcpConn, ok := proxyConn.clientConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
		} else {
			// 检查是否是目标连接关闭导致的错误
			errMsg := err.Error()
			if !strings.Contains(errMsg, "broken pipe") &&
				!strings.Contains(errMsg, "connection reset") &&
				!strings.Contains(errMsg, "use of closed network connection") {
				// 不是目标连接关闭的错误，可能是源连接的问题，仍然关闭写入方向
				if tcpConn, ok := proxyConn.clientConn.(*net.TCPConn); ok {
					tcpConn.CloseWrite()
				}
			}
			// 如果是目标连接关闭的错误，不调用 CloseWrite，避免重复关闭
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
				logger.Error("Connection bridge recoverable error", map[string]interface{}{
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
		})
	}

	if !hasError && totalBytes > 0 {
		logger.Info("Connection bridge completed successfully", map[string]interface{}{
			"serviceId":    proxyConn.serviceID,
			"connectionId": proxyConn.connectionID,
			"totalBytes":   totalBytes,
			"totalTraffic": atomic.LoadInt64(&proxy.totalTraffic),
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
// 当外网请求到达时，通过控制服务器向指定客户端发送代理请求消息。
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
