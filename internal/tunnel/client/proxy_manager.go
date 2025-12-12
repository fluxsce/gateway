// Package client 提供代理管理器的完整实现
// 代理管理器负责管理客户端与服务器之间的数据代理连接
//
// # 代理管理器架构
//
// ## 概述
//
// ProxyManager 负责管理客户端的所有数据代理连接，包括：
//   - 服务器数据连接（Client → Server，由服务端控制生命周期）
//   - 本地服务连接（Client → Local Service，与服务器连接绑定）
//   - 双向数据转发和流量统计
//
// ## 连接管理机制
//
// ### 核心原则
//
// 1. **客户端不维护连接池**：连接由服务端主动创建和管理
// 2. **连接生命周期由客户端控制**：客户端主动关闭连接，数据传输阶段无超时限制
// 3. **serverConn 和 localConn 绑定**：两个连接的生命周期一致
//
// ### 工作流程
//
// 1. 服务端发送 `proxy_request` 消息通知客户端
// 2. 客户端建立新的 `serverConn`（数据连接）到服务端
// 3. 客户端建立新的 `localConn` 到本地服务
// 4. 启动双向数据转发（`relayData`）
// 5. 当任一连接关闭时，`relayData` 返回，两个连接都关闭
//
// ### 连接特点
//
//   - 不使用连接池，每次请求建立新连接
//   - 支持长连接（如 SSE），数据传输阶段无超时限制
//   - 自动配置 TCP 选项（KeepAlive、NoDelay）
//   - 连接由客户端主动关闭，支持长时间数据传输
//
// ## 连接生命周期
//
//	服务端发送 proxy_request → 客户端建立 serverConn
//	                            ↓
//	                        建立 localConn
//	                            ↓
//	                        双向数据转发（阻塞）
//	                            ↓
//	                        任一连接关闭 → 两个连接都关闭
//
// ## 适用场景
//
//	✅ HTTP/HTTPS 请求
//	✅ SSE（Server-Sent Events）长连接
//	✅ WebSocket 连接
//	✅ 大文件传输
//	✅ 所有需要双向数据转发的场景
package client

import (
	"context"
	"fmt"
	"io"
	"net"
	"strings"
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
		logger.Debug("Proxy already exists, skipping creation", map[string]interface{}{
			"serviceId":   service.TunnelServiceId,
			"serviceName": service.ServiceName,
		})
		return nil // 已存在则忽略，不报错
	}

	// 创建代理实例
	// 注意：客户端不维护连接池，连接由服务端控制生命周期
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
		"note":        "no timeout for data transmission, client will close connections",
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

	// 注意：不需要清理连接池，因为客户端不维护连接池
	// 所有连接都在 HandleProxyConnection 中管理，当连接关闭时会自动清理

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

// HandleProxyConnection 处理代理连接
//
// 处理单个代理连接，负责在内网服务和外网请求之间建立数据通道。
// 连接由客户端主动关闭，数据传输阶段无超时限制，支持长时间数据传输。
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
//  2. 建立新的本地服务连接（连接建立阶段有超时，避免无限等待）
//  3. 启动双向数据转发（服务器↔本地服务，阻塞直到任一连接关闭）
//  4. 当 relayData 返回时，客户端主动关闭两个连接
//
// 连接管理:
//   - 客户端主动关闭连接，数据传输阶段无超时限制
//   - serverConn 和 localConn 绑定，生命周期一致
//   - 当 relayData 返回时，客户端主动关闭两个连接
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

	// 建立本地服务连接
	// 注意：连接建立阶段有超时限制（避免本地服务不可用时无限等待）
	// 但数据传输阶段无超时限制，支持长时间数据传输
	localAddr := net.JoinHostPort(proxy.service.LocalAddress, fmt.Sprintf("%d", proxy.service.LocalPort))
	localConn, err := net.DialTimeout("tcp", localAddr, 30*time.Second)
	if err != nil {
		// 连接本地服务失败时，优雅关闭服务器连接
		logger.Error("Failed to connect to local service", map[string]interface{}{
			"serviceId":      serviceID,
			"localAddr":      localAddr,
			"error":          err.Error(),
			"serverConnOpen": serverConn != nil,
		})

		// 优雅关闭服务器连接的写入方向
		if serverConn != nil {
			if tcpConn, ok := serverConn.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
				time.Sleep(100 * time.Millisecond)
			}
			serverConn.Close()
		}

		return fmt.Errorf("failed to connect to local service: %w", err)
	}

	logger.Debug("Created local connection", map[string]interface{}{
		"serviceId": serviceID,
		"localAddr": localAddr,
	})

	// 设置 TCP 选项，支持长连接
	// 注意：不设置读取/写入超时，数据传输阶段无超时限制（Go 默认无超时）
	if tcpConn, ok := localConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}
	if tcpConn, ok := serverConn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetNoDelay(true)
	}

	// 启动双向数据转发
	// 注意：relayData 会阻塞直到任一连接关闭或上下文取消
	// 数据传输阶段无超时限制，支持长时间数据传输
	logger.Debug("Starting data relay", map[string]interface{}{
		"serviceId":     serviceID,
		"hasLocalConn":  localConn != nil,
		"hasServerConn": serverConn != nil,
		"note":          "no timeout for data transmission, client will close connections",
	})

	err = pm.relayData(ctx, serverConn, localConn, proxy)

	// relayData 返回时，客户端主动关闭两个连接
	// 注意：copyData 的 defer 已经关闭了写入方向，这里只需要完全关闭连接
	// 这确保了连接由客户端控制生命周期
	if localConn != nil {
		localConn.Close()
	}
	if serverConn != nil {
		serverConn.Close()
	}

	if err != nil {
		logger.Warn("Data relay completed with error", map[string]interface{}{
			"serviceId": serviceID,
			"error":     err.Error(),
		})
	} else {
		logger.Debug("Data relay completed, connections closed", map[string]interface{}{
			"serviceId": serviceID,
		})
	}

	return err
}

// relayData 双向数据转发
//
// 在服务器连接和本地服务连接之间进行双向数据转发。
// 数据传输阶段无超时限制，支持长时间数据传输（如大文件传输、长连接等）。
// 当任一连接关闭或上下文取消时，转发结束。
//
// 参数:
//   - ctx: 上下文，用于取消操作
//   - remoteConn: 服务器连接
//   - localConn: 本地服务连接
//   - proxy: 代理实例
//
// 返回:
//   - error: 转发过程中的错误（如果有）
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
		// 数据转发完成（任一方向连接关闭）
	case <-ctx.Done():
		// 控制连接断开，应该关闭数据连接
		// 优雅关闭：关闭写入方向，通知对方
		if tcpConn, ok := remoteConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		if tcpConn, ok := localConn.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
		return ctx.Err()
	}

	// 更新总流量
	proxy.mutex.Lock()
	proxy.totalTraffic += totalBytes
	proxy.mutex.Unlock()

	return nil
}

// copyData 复制数据
//
// 从源连接复制数据到目标连接。
// 不设置超时限制，支持长时间数据传输。
// 当源连接关闭时，优雅关闭目标连接的写入方向。
//
// 参数:
//   - dst: 目标连接
//   - src: 源连接
//
// 返回:
//   - int64: 复制的字节数
//   - error: 复制过程中的错误
func (pm *proxyManager) copyData(dst, src net.Conn) (int64, error) {
	bytes, err := io.Copy(dst, src)

	// 关键修复：只有在正常结束（io.EOF）时才关闭写入方向
	// 如果是其他错误（如 broken pipe、connection reset），说明目标连接可能已经关闭
	// 此时不应该再调用 CloseWrite，避免在 HTTP chunked encoding 传输过程中提前关闭
	// HTTP chunked encoding 需要发送结束标记（0\r\n\r\n），如果写入方向被提前关闭，会导致 ERR_INCOMPLETE_CHUNKED_ENCODING
	if err == nil || err == io.EOF {
		// 正常结束，优雅关闭写入方向
		if tcpConn, ok := dst.(*net.TCPConn); ok {
			tcpConn.CloseWrite()
		}
	} else {
		// 检查是否是目标连接关闭导致的错误
		errMsg := err.Error()
		if !strings.Contains(errMsg, "broken pipe") &&
			!strings.Contains(errMsg, "connection reset") &&
			!strings.Contains(errMsg, "use of closed network connection") {
			// 不是目标连接关闭的错误，可能是源连接的问题，仍然关闭写入方向
			if tcpConn, ok := dst.(*net.TCPConn); ok {
				tcpConn.CloseWrite()
			}
		}
		// 如果是目标连接关闭的错误，不调用 CloseWrite，避免重复关闭
	}

	return bytes, err
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
