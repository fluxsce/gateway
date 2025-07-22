package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"

	"github.com/gorilla/websocket"
)

// WebSocketConnection WebSocket连接管理
type WebSocketConnection struct {
	clientConn *websocket.Conn // 客户端连接
	targetConn *websocket.Conn // 目标服务连接
	ctx        context.Context
	cancel     context.CancelFunc
	config     *WebSocketConfig

	// 连接状态
	mu     sync.RWMutex
	closed bool

	// 统计信息
	clientToTarget int64 // 客户端到目标的消息数
	targetToClient int64 // 目标到客户端的消息数
}

// WebSocketProxy WebSocket代理实现
type WebSocketProxy struct {
	*BaseProxyHandler
	serviceManager service.ServiceManager
	config         *WebSocketConfig

	// 连接管理
	connections map[*WebSocketConnection]bool
	connMutex   sync.RWMutex

	// WebSocket升级器
	upgrader websocket.Upgrader
}

// Handle 处理WebSocket代理请求
func (w *WebSocketProxy) Handle(ctx *core.Context) bool {
	if !w.IsEnabled() {
		return true
	}

	// 检查是否为WebSocket升级请求
	if !w.isWebSocketUpgrade(ctx.Request) {
		ctx.AddError(fmt.Errorf("不是WebSocket升级请求"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "不是WebSocket升级请求",
		})
		return false
	}

	// 获取服务ID
	serviceID := ctx.GetServiceID()
	if serviceID == "" {
		ctx.AddError(fmt.Errorf("服务ID不能为空"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "服务ID不能为空",
		})
		return false
	}

	// 从负载均衡器获取目标节点
	node, err := w.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		ctx.AddError(fmt.Errorf("选择目标节点失败: %w", err))
		ctx.Abort(http.StatusServiceUnavailable, map[string]string{
			"error": "服务不可用",
		})
		return false
	}

	// 设置服务名称和代理类型
	ctx.Set(constants.ContextKeyServiceDefinitionName, w.GetName())
	ctx.Set(constants.ContextKeyProxyType, w.GetType())

	// 设置转发开始时间
	ctx.SetForwardStartTime(time.Now())

	// 代理WebSocket请求
	err = w.ProxyRequest(ctx, node.URL)
	if err != nil {
		ctx.AddError(fmt.Errorf("代理WebSocket请求失败: %w", err))
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "代理WebSocket请求失败",
		})
		ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
		return false
	}

	return true
}

// ProxyRequest 代理WebSocket请求到指定URL
func (w *WebSocketProxy) ProxyRequest(ctx *core.Context, targetURL string) error {
	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("解析目标URL失败: %w", err)
	}

	// 构建WebSocket目标URL
	wsScheme := "ws"
	if target.Scheme == "https" {
		wsScheme = "wss"
	}

	targetWSURL := &url.URL{
		Scheme:   wsScheme,
		Host:     target.Host,
		Path:     ctx.Request.URL.Path,
		RawQuery: ctx.Request.URL.RawQuery,
	}

	// 设置目标URL
	ctx.SetTargetURL(targetWSURL.String())

	// 升级客户端连接到WebSocket
	clientConn, err := w.upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		ctx.SetForwardResponseTime(time.Now())
		return fmt.Errorf("升级客户端连接失败: %w", err)
	}

	// 创建连接到目标服务的WebSocket连接
	targetConn, resp, err := w.connectToTarget(targetWSURL, ctx.Request)
	if err != nil {
		clientConn.Close()
		ctx.SetForwardResponseTime(time.Now())
		return fmt.Errorf("连接到目标服务失败: %w", err)
	}

	// 设置后端状态码
	if resp != nil {
		ctx.Set(constants.BackendStatusCode, resp.StatusCode)
		ctx.Set(constants.GatewayStatusCode, resp.StatusCode)
	}

	// 创建WebSocket连接管理对象
	connCtx, cancel := context.WithCancel(context.Background())
	wsConn := &WebSocketConnection{
		clientConn: clientConn,
		targetConn: targetConn,
		ctx:        connCtx,
		cancel:     cancel,
		config:     w.config,
	}

	// 添加到连接池
	w.connMutex.Lock()
	w.connections[wsConn] = true
	w.connMutex.Unlock()

	// 设置连接参数
	w.configureConnection(wsConn)

	// 开始代理消息
	go w.proxyMessages(wsConn)

	// 等待连接结束或上下文取消
	select {
	case <-connCtx.Done():
		// 连接正常结束
	case <-ctx.Request.Context().Done():
		// 请求上下文取消
		wsConn.cancel()
	}

	// 清理连接
	w.closeConnection(wsConn)
	ctx.SetForwardResponseTime(time.Now())

	return nil
}

// connectToTarget 连接到目标WebSocket服务
func (w *WebSocketProxy) connectToTarget(targetURL *url.URL, req *http.Request) (*websocket.Conn, *http.Response, error) {
	// 创建WebSocket拨号器
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    w.config.ReadBufferSize,
		WriteBufferSize:   w.config.WriteBufferSize,
		Subprotocols:      w.config.Subprotocols,
		EnableCompression: w.config.EnableCompression,
	}

	// 复制客户端的头部到目标连接
	headers := make(http.Header)
	for name, values := range req.Header {
		// 跳过WebSocket特定头部，由库自动处理
		if isWebSocketHeader(name) {
			continue
		}
		// 跳过hop-by-hop头部
		if isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			headers.Add(name, value)
		}
	}

	// 设置代理头部
	w.setProxyHeaders(req, headers, targetURL.Host)

	// 连接到目标服务
	conn, resp, err := dialer.Dial(targetURL.String(), headers)
	if err != nil {
		return nil, resp, err
	}

	return conn, resp, nil
}

// configureConnection 配置WebSocket连接参数
func (w *WebSocketProxy) configureConnection(wsConn *WebSocketConnection) {
	// 设置客户端连接参数
	if w.config.ReadTimeout > 0 {
		wsConn.clientConn.SetReadDeadline(time.Now().Add(w.config.ReadTimeout))
	}
	if w.config.WriteTimeout > 0 {
		wsConn.clientConn.SetWriteDeadline(time.Now().Add(w.config.WriteTimeout))
	}
	if w.config.MaxMessageSize > 0 {
		wsConn.clientConn.SetReadLimit(w.config.MaxMessageSize)
	}

	// 设置目标连接参数
	if w.config.ReadTimeout > 0 {
		wsConn.targetConn.SetReadDeadline(time.Now().Add(w.config.ReadTimeout))
	}
	if w.config.WriteTimeout > 0 {
		wsConn.targetConn.SetWriteDeadline(time.Now().Add(w.config.WriteTimeout))
	}
	if w.config.MaxMessageSize > 0 {
		wsConn.targetConn.SetReadLimit(w.config.MaxMessageSize)
	}

	// 设置心跳处理
	if w.config.PingInterval > 0 {
		w.setupPingPong(wsConn)
	}
}

// setupPingPong 设置心跳检测
func (w *WebSocketProxy) setupPingPong(wsConn *WebSocketConnection) {
	// 设置pong处理器
	wsConn.clientConn.SetPongHandler(func(string) error {
		if w.config.PongTimeout > 0 {
			wsConn.clientConn.SetReadDeadline(time.Now().Add(w.config.PongTimeout))
		}
		return nil
	})

	wsConn.targetConn.SetPongHandler(func(string) error {
		if w.config.PongTimeout > 0 {
			wsConn.targetConn.SetReadDeadline(time.Now().Add(w.config.PongTimeout))
		}
		return nil
	})

	// 启动ping定时器
	go func() {
		ticker := time.NewTicker(w.config.PingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				wsConn.mu.RLock()
				if wsConn.closed {
					wsConn.mu.RUnlock()
					return
				}
				wsConn.mu.RUnlock()

				// 向客户端发送ping
				if err := wsConn.clientConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					wsConn.cancel()
					return
				}

				// 向目标发送ping
				if err := wsConn.targetConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					wsConn.cancel()
					return
				}

			case <-wsConn.ctx.Done():
				return
			}
		}
	}()
}

// proxyMessages 代理消息传输
func (w *WebSocketProxy) proxyMessages(wsConn *WebSocketConnection) {
	// 启动两个goroutine分别处理两个方向的消息
	var wg sync.WaitGroup
	wg.Add(2)

	// 客户端到目标服务
	go func() {
		defer wg.Done()
		w.copyMessages(wsConn.clientConn, wsConn.targetConn, wsConn, true)
	}()

	// 目标服务到客户端
	go func() {
		defer wg.Done()
		w.copyMessages(wsConn.targetConn, wsConn.clientConn, wsConn, false)
	}()

	// 等待任一方向的消息传输结束
	wg.Wait()

	// 取消连接上下文
	wsConn.cancel()
}

// copyMessages 复制消息从源连接到目标连接
func (w *WebSocketProxy) copyMessages(src, dst *websocket.Conn, wsConn *WebSocketConnection, isClientToTarget bool) {
	defer func() {
		// 发生错误时关闭连接
		wsConn.cancel()
	}()

	for {
		select {
		case <-wsConn.ctx.Done():
			return
		default:
		}

		// 读取消息
		messageType, message, err := src.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// 记录异常关闭错误
			}
			return
		}

		// 更新读取超时
		if w.config.ReadTimeout > 0 {
			src.SetReadDeadline(time.Now().Add(w.config.ReadTimeout))
		}

		// 写入消息到目标连接
		if w.config.WriteTimeout > 0 {
			dst.SetWriteDeadline(time.Now().Add(w.config.WriteTimeout))
		}

		err = dst.WriteMessage(messageType, message)
		if err != nil {
			return
		}

		// 更新统计信息
		wsConn.mu.Lock()
		if isClientToTarget {
			wsConn.clientToTarget++
		} else {
			wsConn.targetToClient++
		}
		wsConn.mu.Unlock()
	}
}

// isWebSocketUpgrade 检查是否为WebSocket升级请求
func (w *WebSocketProxy) isWebSocketUpgrade(req *http.Request) bool {
	// 检查必要的头部
	if strings.ToLower(req.Header.Get("Connection")) != "upgrade" {
		return false
	}
	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" {
		return false
	}
	if req.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}
	if req.Header.Get("Sec-WebSocket-Version") != "13" {
		return false
	}
	return true
}

// isWebSocketHeader 检查是否为WebSocket特定头部
func isWebSocketHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection",
		"upgrade",
		"sec-websocket-key",
		"sec-websocket-version",
		"sec-websocket-protocol",
		"sec-websocket-extensions":
		return true
	default:
		return false
	}
}

// setProxyHeaders 设置代理头部
func (w *WebSocketProxy) setProxyHeaders(req *http.Request, headers http.Header, targetHost string) {
	// 设置X-Forwarded-* 头部
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		headers.Set("X-Forwarded-For", xff+", "+w.getClientIP(req))
	} else {
		headers.Set("X-Forwarded-For", w.getClientIP(req))
	}

	headers.Set("X-Real-IP", w.getClientIP(req))

	scheme := "ws"
	if req.TLS != nil {
		scheme = "wss"
	}
	headers.Set("X-Forwarded-Proto", scheme)
	headers.Set("X-Forwarded-Host", req.Host)

	// 设置User-Agent
	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "Gateway-Gateway/1.0")
	}
}

// getClientIP 获取客户端真实IP
func (w *WebSocketProxy) getClientIP(req *http.Request) string {
	// 优先级：X-Forwarded-For > X-Real-IP > RemoteAddr
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		return ip
	}

	return req.RemoteAddr
}

// GetWebSocketConfig 获取WebSocket代理配置
func (w *WebSocketProxy) GetWebSocketConfig() *WebSocketConfig {
	return w.config
}

// Validate 验证WebSocket代理配置
func (w *WebSocketProxy) Validate() error {
	wsConfig := w.GetWebSocketConfig()
	if wsConfig.PingInterval < 0 {
		return fmt.Errorf("Ping间隔不能为负数")
	}
	if wsConfig.PongTimeout < 0 {
		return fmt.Errorf("Pong超时不能为负数")
	}
	if wsConfig.WriteTimeout < 0 {
		return fmt.Errorf("写超时不能为负数")
	}
	if wsConfig.ReadTimeout < 0 {
		return fmt.Errorf("读超时不能为负数")
	}
	if wsConfig.MaxMessageSize < 0 {
		return fmt.Errorf("最大消息大小不能为负数")
	}
	if wsConfig.ReadBufferSize < 0 {
		return fmt.Errorf("读缓冲区大小不能为负数")
	}
	if wsConfig.WriteBufferSize < 0 {
		return fmt.Errorf("写缓冲区大小不能为负数")
	}

	return nil
}

// closeConnection 关闭WebSocket连接
func (w *WebSocketProxy) closeConnection(wsConn *WebSocketConnection) {
	wsConn.mu.Lock()
	if wsConn.closed {
		wsConn.mu.Unlock()
		return
	}
	wsConn.closed = true
	wsConn.mu.Unlock()

	// 取消上下文
	wsConn.cancel()

	// 关闭WebSocket连接
	if wsConn.clientConn != nil {
		wsConn.clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		wsConn.clientConn.Close()
	}

	if wsConn.targetConn != nil {
		wsConn.targetConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		wsConn.targetConn.Close()
	}

	// 从连接池中移除
	w.connMutex.Lock()
	delete(w.connections, wsConn)
	w.connMutex.Unlock()
}

// Close 关闭WebSocket代理
func (w *WebSocketProxy) Close() error {
	var lastErr error

	// 关闭所有活跃的WebSocket连接
	w.connMutex.Lock()
	connections := make([]*WebSocketConnection, 0, len(w.connections))
	for conn := range w.connections {
		connections = append(connections, conn)
	}
	w.connMutex.Unlock()

	// 依次关闭所有连接
	for _, conn := range connections {
		w.closeConnection(conn)
	}

	// 关闭服务管理器
	if w.serviceManager != nil {
		if closer, ok := w.serviceManager.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				lastErr = err
			}
		}
	}

	return lastErr
}

// GetConnectionCount 获取当前连接数
func (w *WebSocketProxy) GetConnectionCount() int {
	w.connMutex.RLock()
	defer w.connMutex.RUnlock()
	return len(w.connections)
}

// GetConnectionStats 获取连接统计信息
func (w *WebSocketProxy) GetConnectionStats() map[string]interface{} {
	w.connMutex.RLock()
	defer w.connMutex.RUnlock()

	stats := map[string]interface{}{
		"total_connections": len(w.connections),
		"connections":       make([]map[string]interface{}, 0, len(w.connections)),
	}

	for conn := range w.connections {
		conn.mu.RLock()
		connStats := map[string]interface{}{
			"client_to_target": conn.clientToTarget,
			"target_to_client": conn.targetToClient,
			"closed":           conn.closed,
		}
		conn.mu.RUnlock()
		stats["connections"] = append(stats["connections"].([]map[string]interface{}), connStats)
	}

	return stats
}

// NewWebSocketProxy 创建WebSocket代理
func NewWebSocketProxy(config ProxyConfig, serviceManager service.ServiceManager) (*WebSocketProxy, error) {
	// 解析WebSocket特定配置
	wsConfig := DefaultWebSocketConfig
	if config.Config != nil {
		// 创建WebSocket配置解析器
		parser := NewWebSocketConfigParser()
		parser.ParseConfig(config.Config, &wsConfig)
	}

	// 创建WebSocket升级器
	upgrader := websocket.Upgrader{
		ReadBufferSize:    wsConfig.ReadBufferSize,
		WriteBufferSize:   wsConfig.WriteBufferSize,
		Subprotocols:      wsConfig.Subprotocols,
		EnableCompression: wsConfig.EnableCompression,
		CheckOrigin: func(r *http.Request) bool {
			// 默认允许所有来源，生产环境应该根据需要限制
			return true
		},
	}

	return &WebSocketProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
		config:           &wsConfig,
		connections:      make(map[*WebSocketConnection]bool),
		upgrader:         upgrader,
	}, nil
}
