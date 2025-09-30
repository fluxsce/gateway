// Package server 提供正向代理服务器的完整实现
// 正向代理服务器负责处理客户端到外部服务器的代理转发，支持多种代理类型（TCP、UDP、HTTP等）
package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"gateway/pkg/logger"
)

// forwardProxyServer 正向代理服务器实现
// 实现 ForwardProxyServer 接口，处理客户端到外部服务器的代理转发
type forwardProxyServer struct {
	tunnelServer TunnelServer
	proxies      map[string]*proxyInstance
	proxyMutex   sync.RWMutex
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
}

// proxyInstance 代理实例
type proxyInstance struct {
	config       *ProxyConfig
	listener     net.Listener
	httpServer   *http.Server
	status       string
	startTime    time.Time
	connections  int32
	totalConns   int64
	totalTraffic int64
	connMutex    sync.RWMutex
}

// NewForwardProxyServerImpl 创建新的正向代理服务器实例
//
// 参数:
//   - tunnelServer: 隧道服务器实例，用于获取配置和状态
//
// 返回:
//   - ForwardProxyServer: 正向代理服务器接口实例
//
// 功能:
//   - 初始化正向代理服务器
//   - 创建代理实例映射表
//   - 设置默认配置
func NewForwardProxyServerImpl(tunnelServer TunnelServer) ForwardProxyServer {
	return &forwardProxyServer{
		tunnelServer: tunnelServer,
		proxies:      make(map[string]*proxyInstance),
	}
}

// StartProxy 启动指定类型的代理服务
//
// 参数:
//   - ctx: 上下文，用于控制代理生命周期
//   - config: 代理配置，包含代理类型、端口等信息
//
// 返回:
//   - error: 启动失败时返回错误
//
// 功能:
//   - 根据代理类型启动相应的代理服务
//   - 支持 TCP、UDP、HTTP、HTTPS 代理
//   - 记录代理状态和统计信息
func (s *forwardProxyServer) StartProxy(ctx context.Context, config *ProxyConfig) error {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()

	// 检查代理是否已存在
	if _, exists := s.proxies[config.ProxyID]; exists {
		return fmt.Errorf("proxy %s already exists", config.ProxyID)
	}

	// 创建代理实例
	proxy := &proxyInstance{
		config:    config,
		status:    "starting",
		startTime: time.Now(),
	}

	// 根据代理类型启动服务
	switch config.ProxyType {
	case ProxyTypeTCP:
		if err := s.startTCPProxy(ctx, proxy); err != nil {
			return fmt.Errorf("failed to start TCP proxy: %w", err)
		}
	case ProxyTypeUDP:
		if err := s.startUDPProxy(ctx, proxy); err != nil {
			return fmt.Errorf("failed to start UDP proxy: %w", err)
		}
	case ProxyTypeHTTP:
		if err := s.startHTTPProxy(ctx, proxy); err != nil {
			return fmt.Errorf("failed to start HTTP proxy: %w", err)
		}
	case ProxyTypeHTTPS:
		if err := s.startHTTPSProxy(ctx, proxy); err != nil {
			return fmt.Errorf("failed to start HTTPS proxy: %w", err)
		}
	default:
		return fmt.Errorf("unsupported proxy type: %s", config.ProxyType)
	}

	proxy.status = "running"
	s.proxies[config.ProxyID] = proxy

	logger.Info("Proxy started", map[string]interface{}{
		"proxyId":   config.ProxyID,
		"proxyType": config.ProxyType,
		"port":      config.ListenPort,
	})

	return nil
}

// StopProxy 停止指定的代理服务
//
// 参数:
//   - ctx: 上下文，用于控制停止超时
//   - proxyID: 代理ID
//
// 返回:
//   - error: 停止失败时返回错误
//
// 功能:
//   - 关闭代理监听器
//   - 等待现有连接处理完成
//   - 清理代理实例
func (s *forwardProxyServer) StopProxy(ctx context.Context, proxyID string) error {
	s.proxyMutex.Lock()
	defer s.proxyMutex.Unlock()

	proxy, exists := s.proxies[proxyID]
	if !exists {
		return fmt.Errorf("proxy %s not found", proxyID)
	}

	proxy.status = "stopping"

	// 关闭监听器
	if proxy.listener != nil {
		proxy.listener.Close()
	}

	// 关闭HTTP服务器
	if proxy.httpServer != nil {
		shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		proxy.httpServer.Shutdown(shutdownCtx)
	}

	// 从映射中移除
	delete(s.proxies, proxyID)

	logger.Info("Proxy stopped", map[string]interface{}{
		"proxyId": proxyID,
	})

	return nil
}

// GetActiveProxies 获取活跃的代理服务
//
// 返回:
//   - []*ProxyInfo: 活跃代理服务的信息列表
//
// 功能:
//   - 返回所有运行中的代理服务信息
//   - 包含连接数、流量统计等信息
func (s *forwardProxyServer) GetActiveProxies() []*ProxyInfo {
	s.proxyMutex.RLock()
	defer s.proxyMutex.RUnlock()

	var proxies []*ProxyInfo
	for _, proxy := range s.proxies {
		proxy.connMutex.RLock()
		info := &ProxyInfo{
			ProxyID:           proxy.config.ProxyID,
			ProxyType:         proxy.config.ProxyType,
			ListenAddress:     proxy.config.ListenAddress,
			ListenPort:        proxy.config.ListenPort,
			Status:            proxy.status,
			StartTime:         proxy.startTime,
			ActiveConnections: int(proxy.connections),
			TotalConnections:  proxy.totalConns,
			TotalTraffic:      proxy.totalTraffic,
		}
		proxy.connMutex.RUnlock()
		proxies = append(proxies, info)
	}

	return proxies
}

// HandleProxyConnection 处理代理连接
//
// 参数:
//   - ctx: 上下文
//   - conn: 网络连接
//   - proxyID: 代理ID
//
// 返回:
//   - error: 处理失败时返回错误
//
// 功能:
//   - 处理特定代理的连接
//   - 更新连接统计
//   - 转发数据到目标地址
func (s *forwardProxyServer) HandleProxyConnection(ctx context.Context, conn net.Conn, proxyID string) error {
	s.proxyMutex.RLock()
	proxy, exists := s.proxies[proxyID]
	s.proxyMutex.RUnlock()

	if !exists {
		return fmt.Errorf("proxy %s not found", proxyID)
	}

	defer conn.Close()

	// 更新连接统计
	proxy.connMutex.Lock()
	proxy.connections++
	proxy.totalConns++
	proxy.connMutex.Unlock()

	defer func() {
		proxy.connMutex.Lock()
		proxy.connections--
		proxy.connMutex.Unlock()
	}()

	// 连接到目标地址
	targetAddr := fmt.Sprintf("%s:%d", proxy.config.TargetAddress, proxy.config.TargetPort)
	targetConn, err := net.Dial("tcp", targetAddr)
	if err != nil {
		return fmt.Errorf("failed to connect to target %s: %w", targetAddr, err)
	}
	defer targetConn.Close()

	// 双向数据转发
	return s.relayData(ctx, conn, targetConn, proxy)
}

// startTCPProxy 启动TCP代理
func (s *forwardProxyServer) startTCPProxy(ctx context.Context, proxy *proxyInstance) error {
	addr := fmt.Sprintf("%s:%d", proxy.config.ListenAddress, proxy.config.ListenPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	proxy.listener = listener

	// 启动连接接受循环
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptTCPConnections(ctx, proxy)
	}()

	return nil
}

// startUDPProxy 启动UDP代理
func (s *forwardProxyServer) startUDPProxy(ctx context.Context, proxy *proxyInstance) error {
	addr := fmt.Sprintf("%s:%d", proxy.config.ListenAddress, proxy.config.ListenPort)
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address %s: %w", addr, err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on UDP %s: %w", addr, err)
	}

	// 启动UDP数据处理循环
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		defer conn.Close()
		s.handleUDPProxy(ctx, proxy, conn)
	}()

	return nil
}

// startHTTPProxy 启动HTTP代理
func (s *forwardProxyServer) startHTTPProxy(ctx context.Context, proxy *proxyInstance) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.handleHTTPRequest(w, r, proxy)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", proxy.config.ListenAddress, proxy.config.ListenPort),
		Handler: mux,
	}

	proxy.httpServer = server

	// 启动HTTP服务器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTP proxy server error", map[string]interface{}{
				"error":   err.Error(),
				"proxyId": proxy.config.ProxyID,
			})
		}
	}()

	return nil
}

// startHTTPSProxy 启动HTTPS代理
func (s *forwardProxyServer) startHTTPSProxy(ctx context.Context, proxy *proxyInstance) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		s.handleHTTPRequest(w, r, proxy)
	})

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", proxy.config.ListenAddress, proxy.config.ListenPort),
		Handler: mux,
	}

	proxy.httpServer = server

	// 获取TLS配置
	tlsCertFile := proxy.config.Options["tlsCertFile"]
	tlsKeyFile := proxy.config.Options["tlsKeyFile"]

	if tlsCertFile == nil || tlsKeyFile == nil {
		return fmt.Errorf("TLS certificate and key files are required for HTTPS proxy")
	}

	// 启动HTTPS服务器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		if err := server.ListenAndServeTLS(tlsCertFile.(string), tlsKeyFile.(string)); err != nil && err != http.ErrServerClosed {
			logger.Error("HTTPS proxy server error", map[string]interface{}{
				"error":   err.Error(),
				"proxyId": proxy.config.ProxyID,
			})
		}
	}()

	return nil
}

// acceptTCPConnections 接受TCP连接
func (s *forwardProxyServer) acceptTCPConnections(ctx context.Context, proxy *proxyInstance) {
	for {
		conn, err := proxy.listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return
			default:
				logger.Error("Failed to accept connection", map[string]interface{}{
					"error":   err.Error(),
					"proxyId": proxy.config.ProxyID,
				})
				continue
			}
		}

		// 为每个连接启动处理协程
		s.wg.Add(1)
		go func(conn net.Conn) {
			defer s.wg.Done()
			if err := s.HandleProxyConnection(ctx, conn, proxy.config.ProxyID); err != nil {
				logger.Error("Proxy connection handling failed", map[string]interface{}{
					"error":   err.Error(),
					"proxyId": proxy.config.ProxyID,
				})
			}
		}(conn)
	}
}

// handleUDPProxy 处理UDP代理
func (s *forwardProxyServer) handleUDPProxy(ctx context.Context, proxy *proxyInstance, conn *net.UDPConn) {
	buffer := make([]byte, 64*1024) // 64KB缓冲区
	clientMap := make(map[string]*net.UDPConn)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, clientAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				logger.Error("UDP read error", map[string]interface{}{
					"error":   err.Error(),
					"proxyId": proxy.config.ProxyID,
				})
				continue
			}

			clientKey := clientAddr.String()

			// 获取或创建到目标的连接
			targetConn, exists := clientMap[clientKey]
			if !exists {
				targetAddr := fmt.Sprintf("%s:%d", proxy.config.TargetAddress, proxy.config.TargetPort)
				targetUDPAddr, err := net.ResolveUDPAddr("udp", targetAddr)
				if err != nil {
					logger.Error("Failed to resolve target UDP address", map[string]interface{}{
						"error":   err.Error(),
						"target":  targetAddr,
						"proxyId": proxy.config.ProxyID,
					})
					continue
				}

				targetConn, err = net.DialUDP("udp", nil, targetUDPAddr)
				if err != nil {
					logger.Error("Failed to connect to target UDP", map[string]interface{}{
						"error":   err.Error(),
						"target":  targetAddr,
						"proxyId": proxy.config.ProxyID,
					})
					continue
				}

				clientMap[clientKey] = targetConn

				// 启动回程数据处理
				s.wg.Add(1)
				go func(client *net.UDPAddr, target *net.UDPConn) {
					defer s.wg.Done()
					defer target.Close()
					s.handleUDPReturn(conn, client, target, proxy)
				}(clientAddr, targetConn)
			}

			// 转发数据到目标
			if _, err := targetConn.Write(buffer[:n]); err != nil {
				logger.Error("Failed to forward UDP data", map[string]interface{}{
					"error":   err.Error(),
					"proxyId": proxy.config.ProxyID,
				})
				targetConn.Close()
				delete(clientMap, clientKey)
			}
		}
	}
}

// handleUDPReturn 处理UDP回程数据
func (s *forwardProxyServer) handleUDPReturn(conn *net.UDPConn, clientAddr *net.UDPAddr, targetConn *net.UDPConn, proxy *proxyInstance) {
	buffer := make([]byte, 64*1024)

	for {
		n, err := targetConn.Read(buffer)
		if err != nil {
			return
		}

		// 转发数据回客户端
		if _, err := conn.WriteToUDP(buffer[:n], clientAddr); err != nil {
			logger.Error("Failed to forward UDP return data", map[string]interface{}{
				"error":   err.Error(),
				"proxyId": proxy.config.ProxyID,
			})
			return
		}

		// 更新流量统计
		proxy.connMutex.Lock()
		proxy.totalTraffic += int64(n)
		proxy.connMutex.Unlock()
	}
}

// handleHTTPRequest 处理HTTP请求
func (s *forwardProxyServer) handleHTTPRequest(w http.ResponseWriter, r *http.Request, proxy *proxyInstance) {
	// 更新连接统计
	proxy.connMutex.Lock()
	proxy.connections++
	proxy.totalConns++
	proxy.connMutex.Unlock()

	defer func() {
		proxy.connMutex.Lock()
		proxy.connections--
		proxy.connMutex.Unlock()
	}()

	// 构建目标URL
	targetURL := fmt.Sprintf("http://%s:%d%s", proxy.config.TargetAddress, proxy.config.TargetPort, r.URL.Path)
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	// 创建代理请求
	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}

	// 复制请求头
	for name, values := range r.Header {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}

	// 发送请求
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to proxy request", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// 设置状态码
	w.WriteHeader(resp.StatusCode)

	// 复制响应体
	written, err := io.Copy(w, resp.Body)
	if err != nil {
		logger.Error("Failed to copy response body", map[string]interface{}{
			"error":   err.Error(),
			"proxyId": proxy.config.ProxyID,
		})
		return
	}

	// 更新流量统计
	proxy.connMutex.Lock()
	proxy.totalTraffic += written
	proxy.connMutex.Unlock()
}

// relayData 双向数据转发
func (s *forwardProxyServer) relayData(ctx context.Context, conn1, conn2 net.Conn, proxy *proxyInstance) error {
	var wg sync.WaitGroup
	var totalBytes int64

	// 启动两个方向的数据转发
	wg.Add(2)

	go func() {
		defer wg.Done()
		bytes, _ := s.copyData(conn1, conn2)
		proxy.connMutex.Lock()
		totalBytes += bytes
		proxy.connMutex.Unlock()
	}()

	go func() {
		defer wg.Done()
		bytes, _ := s.copyData(conn2, conn1)
		proxy.connMutex.Lock()
		totalBytes += bytes
		proxy.connMutex.Unlock()
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
	proxy.connMutex.Lock()
	proxy.totalTraffic += totalBytes
	proxy.connMutex.Unlock()

	return nil
}

// copyData 复制数据
func (s *forwardProxyServer) copyData(dst, src net.Conn) (int64, error) {
	defer dst.Close()
	defer src.Close()

	return io.Copy(dst, src)
}
