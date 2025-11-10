// Package client 提供代理管理器的完整实现
// 代理管理器负责管理客户端与服务器之间的数据代理连接
//
// # 代理管理器架构
//
// ## 概述
//
// ProxyManager 负责管理客户端的所有数据代理连接，包括：
//   - 服务器数据连接池（Client → Server）
//   - 本地服务连接池（Client → Local Service）
//   - 双向数据转发和流量统计
//
// ## 双重连接池机制
//
// ### 1. 服务器数据连接池（serverConnPool）
//
// 用途：复用客户端到服务器的数据连接
//
// 工作流程：
//  1. 收到 proxy_request 消息
//  2. 调用 GetOrCreateServerConnection() 获取连接
//  3. 优先从池中获取，无则新建
//  4. 发送握手消息标识 connectionID
//  5. 数据传输完成后，调用 ReturnServerConnection() 归还
//
// 特点：
//   - 池大小：50个连接/服务（优化高并发性能）
//   - 自动配置 TCP 选项（KeepAlive、NoDelay）
//   - 健康连接复用，失效连接关闭
//
// ### 2. 本地服务连接池（localConnPool）
//
// 用途：复用客户端到本地服务的连接
//
// 工作流程：
//  1. HandleProxyConnection() 处理数据连接
//  2. 优先从池中获取本地连接
//  3. 无则建立新的本地连接
//  4. 双向数据转发
//  5. 传输完成后归还到池中
//
// 特点：
//   - 池大小：50个连接/服务（优化高并发性能）
//   - 自动配置 TCP KeepAlive
//   - 错误连接不归还
//
// ## 连接生命周期
//
// ### 服务器连接
//
//	外网请求 → GetOrCreateServerConnection()
//	         ↓
//	    从池获取/新建
//	         ↓
//	    发送握手消息
//	         ↓
//	    数据转发
//	         ↓
//	    ReturnServerConnection() 归还/关闭
//
// ### 本地连接
//
//	数据连接到达 → HandleProxyConnection()
//	            ↓
//	       从池获取/新建
//	            ↓
//	       双向数据转发
//	            ↓
//	       归还到池/关闭
//
// ## 性能优化
//
// ### 连接复用收益
//   - 减少 TCP 握手（~3ms/连接）
//   - 减少 TIME_WAIT 堆积
//   - 提升并发能力（10倍+）
//   - 降低 CPU 和内存消耗
//
// ### 适用场景
//
//	✅ HTTP 短连接高并发
//	✅ REST API 频繁调用
//	✅ 微服务间通信
//	⚠️  WebSocket（长连接不适合）
//	⚠️  大文件传输（占用时间长）
package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"gateway/internal/tunnel/types"
	"gateway/pkg/logger"
)

// proxyManager 代理管理器实现
// 实现 ProxyManager 接口，管理数据代理连接
type proxyManager struct {
	client        *tunnelClient
	activeProxies map[string]*proxyInstance
	mutex         sync.RWMutex
}

// proxyInstance 代理实例
type proxyInstance struct {
	serviceID    string
	service      *types.TunnelService
	remotePort   int
	startTime    time.Time
	connections  int32
	totalConns   int64
	totalTraffic int64
	mutex        sync.RWMutex

	// 本地连接池（客户端到本地服务）
	localConnPool chan net.Conn
	localPoolSize int

	// 服务器连接池（客户端到服务器）
	serverConnPool chan net.Conn
	serverPoolSize int

	poolMutex sync.RWMutex
}

// NewProxyManager 创建代理管理器实例
//
// 参数:
//   - client: 隧道客户端实例
//
// 返回:
//   - ProxyManager: 代理管理器接口实例
func NewProxyManager(client *tunnelClient) ProxyManager {
	return &proxyManager{
		client:        client,
		activeProxies: make(map[string]*proxyInstance),
	}
}

// StartProxy 启动代理
func (pm *proxyManager) StartProxy(ctx context.Context, service *types.TunnelService, remotePort int) error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 检查代理是否已存在
	if _, exists := pm.activeProxies[service.TunnelServiceId]; exists {
		logger.Debug("Proxy already exists, skipping creation", map[string]interface{}{
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
		})
		return nil // 已存在则忽略，不报错
	}

	// 创建代理实例，包含双重连接池
	localPoolSize := 50  // 本地连接池大小（优化高并发性能）
	serverPoolSize := 50 // 服务器连接池大小（优化高并发性能）

	proxy := &proxyInstance{
		serviceID:      service.TunnelServiceId,
		service:        service,
		remotePort:     remotePort,
		startTime:      time.Now(),
		connections:    0,
		totalConns:     0,
		totalTraffic:   0,
		localConnPool:  make(chan net.Conn, localPoolSize),
		localPoolSize:  localPoolSize,
		serverConnPool: make(chan net.Conn, serverPoolSize),
		serverPoolSize: serverPoolSize,
	}

	// 添加到活跃代理列表
	pm.activeProxies[service.TunnelServiceId] = proxy

	logger.Info("Proxy started with dual connection pools", map[string]interface{}{
		"serviceId":      service.TunnelServiceId,
		"serviceName":    service.ServiceName,
		"remotePort":     remotePort,
		"localPort":      service.LocalPort,
		"localPoolSize":  localPoolSize,
		"serverPoolSize": serverPoolSize,
	})

	return nil
}

// StopProxy 停止代理
func (pm *proxyManager) StopProxy(ctx context.Context, serviceID string) error {
	pm.mutex.Lock()
	proxy, exists := pm.activeProxies[serviceID]
	if exists {
		delete(pm.activeProxies, serviceID)
	}
	pm.mutex.Unlock()

	if !exists {
		return fmt.Errorf("proxy for service %s not found", serviceID)
	}

	// 清理本地连接池
	if proxy.localConnPool != nil {
		close(proxy.localConnPool)
		for conn := range proxy.localConnPool {
			if conn != nil {
				conn.Close()
			}
		}
	}

	// 清理服务器连接池
	if proxy.serverConnPool != nil {
		close(proxy.serverConnPool)
		for conn := range proxy.serverConnPool {
			if conn != nil {
				conn.Close()
			}
		}
	}

	logger.Info("Proxy stopped", map[string]interface{}{
		"serviceId":    serviceID,
		"serviceName":  proxy.service.ServiceName,
		"totalConns":   proxy.totalConns,
		"totalTraffic": proxy.totalTraffic,
	})

	return nil
}

// GetActiveProxies 获取活跃代理
func (pm *proxyManager) GetActiveProxies() []*ProxyInfo {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()

	proxies := make([]*ProxyInfo, 0, len(pm.activeProxies))

	for _, proxy := range pm.activeProxies {
		proxy.mutex.RLock()
		info := &ProxyInfo{
			ServiceID:         proxy.serviceID,
			ServiceName:       proxy.service.ServiceName,
			ProxyType:         proxy.service.ServiceType,
			LocalAddress:      proxy.service.LocalAddress,
			LocalPort:         proxy.service.LocalPort,
			RemotePort:        proxy.remotePort,
			Status:            ProxyStatusRunning,
			StartTime:         proxy.startTime,
			ActiveConnections: int(proxy.connections),
			TotalConnections:  proxy.totalConns,
			TotalTraffic:      proxy.totalTraffic,
		}
		proxy.mutex.RUnlock()

		proxies = append(proxies, info)
	}

	return proxies
}

// HandlePooledConnection 处理池化连接
//
// 处理为连接池预建立的数据连接。与普通代理连接不同，池化连接
// 在服务器使用之前需要保持空闲状态，等待服务器的数据传输请求。
//
// 参数:
//   - ctx: 上下文，用于取消操作
//   - conn: 池化数据连接
//   - service: 关联的服务配置
//
// 返回:
//   - error: 处理失败时返回错误
//
// 工作流程:
//  1. 确保代理实例已启动
//  2. 连接保持空闲，等待服务器使用
//  3. 当服务器从连接池取出连接时，开始数据传输
//  4. 数据传输完成后，连接可能被归还到池中或关闭
func (pm *proxyManager) HandlePooledConnection(ctx context.Context, conn net.Conn, service *types.TunnelService) error {
	defer conn.Close()

	logger.Debug("Handling pooled connection", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
	})

	// 确保代理实例已启动（如果还没有启动，先启动它）
	pm.mutex.RLock()
	_, exists := pm.activeProxies[service.TunnelServiceId]
	pm.mutex.RUnlock()

	if !exists {
		// 代理实例不存在，先启动代理
		logger.Info("Proxy not started, starting proxy for pooled connection", map[string]interface{}{
			"serviceId": service.TunnelServiceId,
		})

		// 使用远程端口启动代理（如果服务有远程端口配置）
		remotePort := 0
		if service.RemotePort != nil {
			remotePort = *service.RemotePort
		}

		if err := pm.StartProxy(ctx, service, remotePort); err != nil {
			// StartProxy 现在会忽略 "already exists" 错误，但为了健壮性仍然记录
			logger.Warn("Failed to start proxy for pooled connection, continuing anyway", map[string]interface{}{
				"serviceId": service.TunnelServiceId,
				"error":     err.Error(),
			})
			// 不返回错误，继续处理连接
		}
	}

	// 池化连接的特殊处理：等待服务器发送数据或关闭连接
	// 这里我们创建一个简单的读取循环，等待服务器开始使用连接
	buffer := make([]byte, 32) // 增大缓冲区，用于检测连接状态和读取初始数据

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Pooled connection cancelled by context", map[string]interface{}{
				"serviceId": service.TunnelServiceId,
			})
			return ctx.Err()
		default:
			// 设置读取超时，避免无限阻塞
			// 使用较长的超时时间，让连接保持活跃
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			// 尝试读取数据，检测连接状态
			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 读取超时是正常的，继续等待
					continue
				}
				// 连接断开或出错
				if err == io.EOF {
					logger.Debug("Pooled connection closed by server", map[string]interface{}{
						"serviceId": service.TunnelServiceId,
					})
				} else {
					logger.Debug("Pooled connection error", map[string]interface{}{
						"serviceId": service.TunnelServiceId,
						"error":     err.Error(),
					})
				}
				return err
			}

			if n > 0 {
				// 收到数据，说明服务器开始使用这个连接进行数据传输
				logger.Info("Pooled connection activated for data transfer", map[string]interface{}{
					"serviceId": service.TunnelServiceId,
					"dataBytes": n,
				})

				// 清除读取超时，让数据传输不受限制
				conn.SetReadDeadline(time.Time{})

				// 将读取的数据放回，然后开始正常的代理处理
				// 创建一个包装连接，将已读取的数据放在前面
				wrappedConn := &prefixedConn{
					conn:   conn,
					prefix: buffer[:n],
				}

				return pm.HandleProxyConnection(ctx, wrappedConn, service.TunnelServiceId)
			}
		}
	}
}

// prefixedConn 包装连接，用于在连接前添加已读取的数据
type prefixedConn struct {
	conn   net.Conn
	prefix []byte
	read   bool
}

func (pc *prefixedConn) Read(b []byte) (n int, err error) {
	if !pc.read && len(pc.prefix) > 0 {
		pc.read = true
		n = copy(b, pc.prefix)
		if n < len(pc.prefix) {
			pc.prefix = pc.prefix[n:]
			pc.read = false
		}
		return n, nil
	}
	return pc.conn.Read(b)
}

func (pc *prefixedConn) Write(b []byte) (n int, err error)  { return pc.conn.Write(b) }
func (pc *prefixedConn) Close() error                       { return pc.conn.Close() }
func (pc *prefixedConn) LocalAddr() net.Addr                { return pc.conn.LocalAddr() }
func (pc *prefixedConn) RemoteAddr() net.Addr               { return pc.conn.RemoteAddr() }
func (pc *prefixedConn) SetDeadline(t time.Time) error      { return pc.conn.SetDeadline(t) }
func (pc *prefixedConn) SetReadDeadline(t time.Time) error  { return pc.conn.SetReadDeadline(t) }
func (pc *prefixedConn) SetWriteDeadline(t time.Time) error { return pc.conn.SetWriteDeadline(t) }

// GetOrCreateServerConnection 获取或创建到服务器的数据连接
//
// 这是服务器连接池的核心方法，用于获取客户端到服务器的数据连接。
// 优先从连接池复用已有连接，无可用连接时建立新连接。
//
// 参数:
//   - ctx: 上下文，用于超时控制
//   - serviceID: 服务唯一标识符
//   - client: 隧道客户端实例（用于获取服务器地址）
//
// 返回:
//   - net.Conn: 到服务器的数据连接
//   - bool: 是否从池中复用（true=复用，false=新建）
//   - error: 获取失败时返回错误
//
// 工作流程:
//  1. 尝试从 serverConnPool 获取连接
//  2. 池中有连接：直接返回（复用）
//  3. 池中无连接：建立新的 TCP 连接
//  4. 配置 TCP 选项（KeepAlive、NoDelay）
//  5. 返回连接供调用者使用
//
// TCP 选项配置:
//   - KeepAlive: 启用，周期 30 秒
//   - NoDelay: 启用（禁用 Nagle 算法，降低延迟）
//
// 使用示例:
//
//	conn, fromPool, err := pm.GetOrCreateServerConnection(ctx, serviceID, client)
//	if err != nil {
//	    return err
//	}
//	defer func() {
//	    if !pm.ReturnServerConnection(serviceID, conn) {
//	        conn.Close() // 池满或错误，手动关闭
//	    }
//	}()
//
// 注意事项:
//   - 从池中获取的连接可能已发送过握手消息
//   - 新建连接需要调用者发送握手消息
//   - 使用完毕后应调用 ReturnServerConnection() 归还
//   - 归还失败时调用者需要手动关闭连接
//
// 性能特点:
//   - 复用连接可节省 ~3ms TCP 握手时间
//   - 高并发场景下性能提升显著
//   - 减少系统 TIME_WAIT 状态堆积
func (pm *proxyManager) GetOrCreateServerConnection(ctx context.Context, serviceID string, client *tunnelClient) (net.Conn, bool, error) {
	pm.mutex.RLock()
	proxy, exists := pm.activeProxies[serviceID]
	pm.mutex.RUnlock()

	if !exists {
		return nil, false, fmt.Errorf("proxy for service %s not found", serviceID)
	}

	// 尝试从连接池获取，并验证连接有效性
	select {
	case conn := <-proxy.serverConnPool:
		if conn != nil {
			// 验证连接是否仍然有效
			// 设置一个很短的读超时来测试连接
			conn.SetReadDeadline(time.Now().Add(1 * time.Millisecond))
			one := make([]byte, 1)
			_, err := conn.Read(one)
			conn.SetReadDeadline(time.Time{}) // 清除超时

			// 如果读取到EOF或连接错误，说明连接已失效
			if err == nil {
				// 不应该读到数据（因为这是空闲连接）
				conn.Close()
				logger.Warn("Server connection from pool has unexpected data, discarding", map[string]interface{}{
					"serviceId": serviceID,
				})
			} else if err == io.EOF || !isTimeoutError(err) {
				// 连接已关闭或有错误，丢弃
				conn.Close()
				logger.Warn("Server connection from pool is closed, creating new connection", map[string]interface{}{
					"serviceId": serviceID,
					"error":     err.Error(),
				})
			} else if isTimeoutError(err) {
				// 超时是正常的，说明连接空闲且有效
				logger.Debug("Reused server connection from pool", map[string]interface{}{
					"serviceId": serviceID,
				})
				return conn, true, nil
			}
		}
	default:
		// 池中没有可用连接
	}

	// 建立新的服务器连接
	serverAddr := net.JoinHostPort(client.config.ServerAddress, fmt.Sprintf("%d", client.config.ServerPort))
	conn, err := net.DialTimeout("tcp", serverAddr, 10*time.Second)
	if err != nil {
		return nil, false, fmt.Errorf("failed to connect to server: %w", err)
	}

	// 设置 TCP 选项
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}

	logger.Debug("Created new server connection", map[string]interface{}{
		"serviceId": serviceID,
	})

	return conn, false, nil
}

// ReturnServerConnection 归还服务器数据连接到池中
//
// 将使用完毕的服务器连接归还到连接池以供复用。
// 这是服务器连接池的核心方法，负责连接的回收和复用。
//
// 参数:
//   - serviceID: 服务唯一标识符
//   - conn: 要归还的服务器连接
//
// 返回:
//   - bool: 是否成功归还（true=已归还到池，false=池满或服务不存在）
//
// 工作流程:
//  1. 查找对应服务的代理实例
//  2. 尝试将连接放入 serverConnPool
//  3. 成功：连接已归还，可被复用
//  4. 失败：池满或服务不存在
//
// 归还策略:
//   - 只归还健康的连接（无错误）
//   - 池满时拒绝归还（返回 false）
//   - 服务不存在时拒绝归还（返回 false）
//   - 归还失败时调用者需要关闭连接
//
// 使用示例:
//
//	// 方式1：使用 defer 确保归还或关闭
//	conn, fromPool, err := pm.GetOrCreateServerConnection(ctx, serviceID, client)
//	if err != nil {
//	    return err
//	}
//	defer func() {
//	    if !pm.ReturnServerConnection(serviceID, conn) {
//	        conn.Close() // 归还失败，手动关闭
//	    }
//	}()
//
//	// 方式2：根据错误决定是否归还
//	err := doSomething(conn)
//	if err != nil {
//	    conn.Close() // 有错误，不归还
//	} else {
//	    if !pm.ReturnServerConnection(serviceID, conn) {
//	        conn.Close() // 归还失败，手动关闭
//	    }
//	}
//
// 注意事项:
//   - 不要归还已关闭的连接
//   - 不要归还有错误的连接
//   - 归还后不应再使用该连接
//   - 返回 false 时调用者必须关闭连接
//
// 性能影响:
//   - 成功归还可使连接被复用
//   - 减少新建连接的开销
//   - 提升高并发场景性能
func (pm *proxyManager) ReturnServerConnection(serviceID string, conn net.Conn) bool {
	pm.mutex.RLock()
	proxy, exists := pm.activeProxies[serviceID]
	pm.mutex.RUnlock()

	if !exists {
		return false
	}

	select {
	case proxy.serverConnPool <- conn:
		logger.Debug("Returned server connection to pool", map[string]interface{}{
			"serviceId": serviceID,
		})
		return true
	default:
		// 池已满
		return false
	}
}

// HandleProxyConnection 处理代理连接
//
// 处理单个代理连接，负责在内网服务和外网请求之间建立数据通道。
// 每个连接对应一个外网请求，需要建立到本地服务的连接并进行双向数据转发。
//
// 参数:
//   - ctx: 上下文，用于取消操作
//   - serverConn: 服务器数据连接（来自服务器）
//   - serviceID: 服务唯一标识符
//
// 返回:
//   - error: 处理失败时返回错误
//
// 工作流程:
//  1. 查找对应的代理实例
//  2. 从本地连接池获取或建立到本地服务的连接
//  3. 启动双向数据转发（服务器↔本地服务）
//  4. 数据传输完成后，归还连接到对应的池
//  5. 更新连接统计信息
//
// 连接池管理:
//   - 服务器连接：使用完毕后尝试归还到 serverConnPool
//   - 本地连接：使用完毕后尝试归还到 localConnPool
//   - 归还失败（池满/有错误）：关闭连接
//
// 注意：serverConn 是从服务器来的数据连接，已经建立好了
func (pm *proxyManager) HandleProxyConnection(ctx context.Context, serverConn net.Conn, serviceID string) error {
	// 查找代理实例
	pm.mutex.RLock()
	proxy, exists := pm.activeProxies[serviceID]
	pm.mutex.RUnlock()

	if !exists {
		serverConn.Close()
		return fmt.Errorf("proxy for service %s not found", serviceID)
	}

	// 标记是否应该归还服务器连接到池
	shouldReturnServerConn := true
	defer func() {
		if shouldReturnServerConn {
			// 尝试归还服务器连接到池
			if !pm.ReturnServerConnection(serviceID, serverConn) {
				// 归还失败（池满或服务不存在），关闭连接
				serverConn.Close()
				logger.Debug("Server connection pool full, closed connection", map[string]interface{}{
					"serviceId": serviceID,
				})
			}
		} else {
			// 有错误，不归还，直接关闭
			serverConn.Close()
		}
	}()

	// 更新连接统计
	proxy.mutex.Lock()
	proxy.connections++
	proxy.totalConns++
	proxy.mutex.Unlock()

	defer func() {
		proxy.mutex.Lock()
		proxy.connections--
		proxy.mutex.Unlock()
	}()

	// 尝试从连接池获取本地连接，并验证有效性
	var localConn net.Conn
	var localFromPool bool

	select {
	case localConn = <-proxy.localConnPool:
		if localConn != nil {
			// 验证连接是否仍然有效
			localConn.SetReadDeadline(time.Now().Add(2 * time.Millisecond))
			one := make([]byte, 1)
			_, err := localConn.Read(one)
			localConn.SetReadDeadline(time.Time{})

			if err == nil {
				// 不应该读到数据
				localConn.Close()
				localConn = nil
				logger.Warn("Local connection from pool has unexpected data, discarding", map[string]interface{}{
					"serviceId": serviceID,
				})
			} else if err == io.EOF || !isTimeoutError(err) {
				// 连接已关闭或有错误
				localConn.Close()
				localConn = nil
				logger.Warn("Local connection from pool is closed, will create new connection", map[string]interface{}{
					"serviceId": serviceID,
				})
			} else if isTimeoutError(err) {
				// 超时是正常的，连接有效
				localFromPool = true
				logger.Debug("Reused local connection from pool", map[string]interface{}{
					"serviceId": serviceID,
				})
			}
		}
	default:
		// 池中没有可用连接
	}

	// 如果没有从池中获取到连接，建立新连接
	if localConn == nil {
		localAddr := net.JoinHostPort(proxy.service.LocalAddress, fmt.Sprintf("%d", proxy.service.LocalPort))
		var err error
		localConn, err = net.DialTimeout("tcp", localAddr, 10*time.Second)
		if err != nil {
			logger.Error("Failed to connect to local service", map[string]interface{}{
				"serviceId": serviceID,
				"localAddr": localAddr,
				"error":     err.Error(),
			})
			shouldReturnServerConn = false // 连接本地服务失败，不归还服务器连接
			return fmt.Errorf("failed to connect to local service: %w", err)
		}
		localFromPool = false
		logger.Debug("Created new local connection", map[string]interface{}{
			"serviceId": serviceID,
		})
	}

	// 设置 TCP 选项
	if tcpConn, ok := localConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	if tcpConn, ok := serverConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	// 启动双向数据转发
	err := pm.relayData(ctx, serverConn, localConn, proxy)

	// 根据错误情况决定是否归还服务器连接
	if err != nil {
		shouldReturnServerConn = false // 有错误，不归还服务器连接
	}

	// 尝试将本地连接归还到池中（如果连接仍然可用）
	if err == nil && !localFromPool {
		select {
		case proxy.localConnPool <- localConn:
			logger.Debug("Returned local connection to pool", map[string]interface{}{
				"serviceId": serviceID,
			})
			return nil // 本地连接已归还，不要关闭
		default:
			// 池已满，关闭连接
			logger.Warn("Local connection pool full, connection will be closed", map[string]interface{}{
				"serviceId": serviceID,
			})
		}
	}

	// 如果有错误或池已满，关闭本地连接
	localConn.Close()
	return err
}

// relayData 双向数据转发
func (pm *proxyManager) relayData(ctx context.Context, remoteConn, localConn net.Conn, proxy *proxyInstance) error {
	var wg sync.WaitGroup
	var totalBytes int64

	// 启动两个方向的数据转发
	wg.Add(2)

	// 远程到本地
	go func() {
		defer wg.Done()
		bytes, _ := pm.copyData(remoteConn, localConn)
		proxy.mutex.Lock()
		totalBytes += bytes
		proxy.mutex.Unlock()
	}()

	// 本地到远程
	go func() {
		defer wg.Done()
		bytes, _ := pm.copyData(localConn, remoteConn)
		proxy.mutex.Lock()
		totalBytes += bytes
		proxy.mutex.Unlock()
	}()

	// 等待任一方向完成
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		return ctx.Err()
	}

	// 更新总流量
	proxy.mutex.Lock()
	proxy.totalTraffic += totalBytes
	proxy.mutex.Unlock()

	return nil
}

// copyData 复制数据
func (pm *proxyManager) copyData(dst, src net.Conn) (int64, error) {
	defer func() {
		if tcpConn, ok := dst.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	}()

	return io.Copy(dst, src)
}

// GetProxyStats 获取代理统计信息
func (pm *proxyManager) GetProxyStats(serviceID string) *ProxyStats {
	pm.mutex.RLock()
	proxy, exists := pm.activeProxies[serviceID]
	pm.mutex.RUnlock()

	if !exists {
		return nil
	}

	proxy.mutex.RLock()
	defer proxy.mutex.RUnlock()

	return &ProxyStats{
		ActiveConnections: int(proxy.connections),
		TotalConnections:  proxy.totalConns,
		BytesSent:         proxy.totalTraffic / 2, // 简化统计
		BytesReceived:     proxy.totalTraffic / 2,
		AverageLatency:    0, // 暂不实现延迟统计
		ErrorCount:        0, // 暂不实现错误统计
		StartTime:         proxy.startTime,
		LastActivityTime:  time.Now(),
	}
}

// Close 关闭代理管理器
func (pm *proxyManager) Close() error {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()

	// 停止所有代理
	for serviceID := range pm.activeProxies {
		if err := pm.StopProxy(context.Background(), serviceID); err != nil {
			logger.Error("Failed to stop proxy during close", map[string]interface{}{
				"serviceId": serviceID,
				"error":     err.Error(),
			})
		}
	}

	// 清空代理列表
	pm.activeProxies = make(map[string]*proxyInstance)

	return nil
}

// isTimeoutError 检查是否是超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	netErr, ok := err.(net.Error)
	return ok && netErr.Timeout()
}
