// Package client 提供代理管理器的完整实现
// 代理管理器负责管理客户端与服务器之间的数据代理连接
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
		return fmt.Errorf("proxy for service %s already exists", service.TunnelServiceId)
	}

	// 创建代理实例
	proxy := &proxyInstance{
		serviceID:    service.TunnelServiceId,
		service:      service,
		remotePort:   remotePort,
		startTime:    time.Now(),
		connections:  0,
		totalConns:   0,
		totalTraffic: 0,
	}

	// 添加到活跃代理列表
	pm.activeProxies[service.TunnelServiceId] = proxy

	logger.Info("Proxy started", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
		"remotePort":  remotePort,
		"localPort":   service.LocalPort,
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
//  1. 连接保持空闲，等待服务器使用
//  2. 当服务器从连接池取出连接时，开始数据传输
//  3. 数据传输完成后，连接可能被归还到池中或关闭
func (pm *proxyManager) HandlePooledConnection(ctx context.Context, conn net.Conn, service *types.TunnelService) error {
	defer conn.Close()

	logger.Debug("Handling pooled connection", map[string]interface{}{
		"serviceId":   service.TunnelServiceId,
		"serviceName": service.ServiceName,
	})

	// 池化连接的特殊处理：等待服务器发送数据或关闭连接
	// 这里我们创建一个简单的读取循环，等待服务器开始使用连接
	buffer := make([]byte, 1) // 用于检测连接状态

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Pooled connection cancelled by context", map[string]interface{}{
				"serviceId": service.TunnelServiceId,
			})
			return ctx.Err()
		default:
			// 设置读取超时，避免无限阻塞
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			// 尝试读取数据，检测连接状态
			n, err := conn.Read(buffer)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 读取超时是正常的，继续等待
					continue
				}
				// 连接断开或出错
				logger.Debug("Pooled connection closed or error", map[string]interface{}{
					"serviceId": service.TunnelServiceId,
					"error":     err.Error(),
				})
				return err
			}

			if n > 0 {
				// 收到数据，说明服务器开始使用这个连接进行数据传输
				logger.Info("Pooled connection activated for data transfer", map[string]interface{}{
					"serviceId": service.TunnelServiceId,
				})

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

// HandleProxyConnection 处理代理连接
//
// 处理单个代理连接，负责在内网服务和外网请求之间建立数据通道。
// 每个连接对应一个外网请求，需要建立到本地服务的连接并进行双向数据转发。
//
// 参数:
//   - ctx: 上下文，用于取消操作
//   - conn: 数据连接（来自服务器）
//   - serviceID: 服务唯一标识符
//
// 返回:
//   - error: 处理失败时返回错误
//
// 工作流程:
//  1. 查找对应的代理实例
//  2. 建立到本地服务的连接
//  3. 启动双向数据转发
//  4. 更新连接统计信息
func (pm *proxyManager) HandleProxyConnection(ctx context.Context, conn net.Conn, serviceID string) error {
	// 查找代理实例
	pm.mutex.RLock()
	proxy, exists := pm.activeProxies[serviceID]
	pm.mutex.RUnlock()

	if !exists {
		conn.Close()
		return fmt.Errorf("proxy for service %s not found", serviceID)
	}

	defer conn.Close()

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

	// 连接到本地服务
	localAddr := net.JoinHostPort(proxy.service.LocalAddress, fmt.Sprintf("%d", proxy.service.LocalPort))
	localConn, err := net.DialTimeout("tcp", localAddr, 10*time.Second)
	if err != nil {
		logger.Error("Failed to connect to local service", map[string]interface{}{
			"serviceId": serviceID,
			"localAddr": localAddr,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to connect to local service: %w", err)
	}
	defer localConn.Close()

	// 为长连接（如SSE）启用TCP KeepAlive
	if tcpConn, ok := localConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	// 启动双向数据转发
	return pm.relayData(ctx, conn, localConn, proxy)
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
