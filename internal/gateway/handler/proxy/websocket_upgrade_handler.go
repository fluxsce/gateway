package proxy

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"

	"github.com/gorilla/websocket"
)

// WebSocketUpgradeHandler WebSocket协议升级处理器
// 负责检测HTTP到WebSocket的协议升级请求，并将连接转发到后端服务
type WebSocketUpgradeHandler struct {
	serviceManager service.ServiceManager
	config         *WebSocketConfig // 使用统一的WebSocket配置

	// 连接管理
	connections sync.Map // map[connectionID]*wsConnection
	connCounter int64    // 连接ID计数器

	// 统计信息
	stats WebSocketStats
}

// wsConnection WebSocket连接实例
type wsConnection struct {
	id           string
	serviceID    string
	clientConn   *websocket.Conn
	targetConn   *websocket.Conn
	ctx          context.Context
	cancel       context.CancelFunc
	createdAt    time.Time
	lastActivity time.Time

	// 统计信息
	clientToTarget int64
	targetToClient int64
}

// WebSocketStats WebSocket统计信息
type WebSocketStats struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalConnections  int64 `json:"total_connections"`
	FailedUpgrades    int64 `json:"failed_upgrades"`
	BytesReceived     int64 `json:"bytes_received"`
	BytesSent         int64 `json:"bytes_sent"`
}

// NewWebSocketUpgradeHandler 创建WebSocket协议升级处理器
//
// 功能说明：
//   - 负责检测HTTP到WebSocket的协议升级请求
//   - 管理WebSocket连接的生命周期，防止资源泄露
//   - 提供连接统计和监控功能
//   - 支持热重载场景下的优雅关闭
//
// 参数：
//   - serviceManager: 服务管理器，用于获取目标服务配置和选择节点
//   - config: WebSocket配置，如果为nil则使用默认配置
//
// 返回值：
//   - *WebSocketUpgradeHandler: 配置完成的WebSocket升级处理器实例
func NewWebSocketUpgradeHandler(serviceManager service.ServiceManager, config *WebSocketConfig) *WebSocketUpgradeHandler {
	if config == nil {
		defaultConfig := DefaultWebSocketConfig
		config = &defaultConfig
	}

	return &WebSocketUpgradeHandler{
		serviceManager: serviceManager,
		config:         config,
		connCounter:    0,
		stats:          WebSocketStats{},
	}
}

// InheritFromHTTPConfig 从HTTP代理配置继承设置
func (h *WebSocketUpgradeHandler) InheritFromHTTPConfig(httpConfig *HTTPProxyConfig) {
	// 基本的继承逻辑保持简单
	// WebSocket配置本身已经足够简单，不需要复杂的继承
}

// IsWebSocketUpgrade 检测HTTP请求是否为WebSocket协议升级请求
//
// WebSocket协议升级检测规则（基于RFC 6455标准）：
//  1. HTTP方法必须为GET
//  2. Connection头部必须包含"Upgrade"
//  3. Upgrade头部必须为"websocket"
//  4. 必须包含Sec-WebSocket-Key头部
//  5. Sec-WebSocket-Version必须为"13"
//
// 参数：
//   - req: HTTP请求对象
//
// 返回值：
//   - bool: true表示是WebSocket升级请求，false表示普通HTTP请求
func (h *WebSocketUpgradeHandler) IsWebSocketUpgrade(req *http.Request) bool {
	// WebSocket升级请求必须使用GET方法
	if req.Method != "GET" {
		return false
	}

	// Connection头部必须包含"Upgrade"指令（大小写不敏感）
	if !strings.Contains(strings.ToLower(req.Header.Get("Connection")), "upgrade") {
		return false
	}

	// Upgrade头部必须指定为"websocket"协议（大小写不敏感）
	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" {
		return false
	}

	// Sec-WebSocket-Key是WebSocket握手的关键，不能为空
	if req.Header.Get("Sec-WebSocket-Key") == "" {
		return false
	}

	// 只支持WebSocket协议版本13（RFC 6455标准）
	if req.Header.Get("Sec-WebSocket-Version") != "13" {
		return false
	}

	return true
}

// HandleWebSocketUpgrade 处理WebSocket协议升级的完整流程
//
// 核心处理步骤：
//  1. 验证服务ID的有效性
//  2. 获取服务配置和WebSocket特定配置
//  3. 选择健康的目标节点
//  4. 执行HTTP到WebSocket的协议升级
//  5. 建立客户端与目标服务的连接代理
//
// 资源管理特点：
//   - 所有连接资源都有完整的生命周期管理
//   - 异常情况下确保资源正确清理，防止泄露
//   - 支持优雅关闭和热重载场景
//
// 参数：
//   - ctx: 网关核心上下文，包含请求信息和响应管理
//   - proxyName: 代理配置名称，用于日志和监控
//   - proxyType: 代理类型，通常为"websocket"
//
// 返回值：
//   - error: 升级过程中的任何错误，nil表示成功
func (h *WebSocketUpgradeHandler) HandleWebSocketUpgrade(ctx *core.Context, proxyName, proxyType string) error {
	serviceID := ctx.GetServiceID()
	if serviceID == "" {
		atomic.AddInt64(&h.stats.FailedUpgrades, 1)
		return fmt.Errorf("服务ID不能为空")
	}

	// 获取服务配置
	serviceConfig, exists := h.serviceManager.GetService(serviceID)
	if !exists || serviceConfig == nil {
		atomic.AddInt64(&h.stats.FailedUpgrades, 1)
		return fmt.Errorf("服务配置不存在: %s", serviceID)
	}

	// 从服务metadata中解析WebSocket配置
	wsConfig := h.parseServiceWebSocketConfig(serviceConfig)

	// 获取目标节点
	node, err := h.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		atomic.AddInt64(&h.stats.FailedUpgrades, 1)
		return fmt.Errorf("选择目标节点失败: %w", err)
	}

	// 设置上下文信息
	ctx.Set(constants.ContextKeyServiceDefinitionName, proxyName)
	ctx.Set(constants.ContextKeyProxyType, proxyType)
	ctx.SetForwardStartTime(time.Now())

	// 执行WebSocket升级
	return h.proxyWebSocketUpgrade(ctx, node, wsConfig)
}

// parseServiceWebSocketConfig 从服务的ServiceMetadata中解析WebSocket配置
// 使用统一的配置解析器，支持驼峰命名和下划线命名方式
// 确保配置解析的一致性和可扩展性
func (h *WebSocketUpgradeHandler) parseServiceWebSocketConfig(serviceConfig *service.ServiceConfig) *WebSocketConfig {
	// 从默认配置开始，避免nil指针引用
	config := *h.config // 深拷贝默认配置

	// 如果服务没有ServiceMetadata，直接返回默认配置
	if serviceConfig.ServiceMetadata == nil || len(serviceConfig.ServiceMetadata) == 0 {
		return &config
	}

	// 将string类型的metadata转换为interface{}类型以供解析器使用
	configMap := make(map[string]interface{})
	for key, value := range serviceConfig.ServiceMetadata {
		configMap[key] = value
	}

	// 使用专门的WebSocket配置解析器，支持多种命名方式
	parser := NewWebSocketConfigParser()
	parser.ParseConfig(configMap, &config)

	return &config
}

// proxyWebSocketUpgrade 执行WebSocket协议升级和连接代理
// 这是HTTP协议升级为WebSocket协议的核心逻辑
// 确保所有资源在异常情况下都能正确清理，避免资源泄露
func (h *WebSocketUpgradeHandler) proxyWebSocketUpgrade(ctx *core.Context, node *service.NodeConfig, wsConfig *WebSocketConfig) error {
	// 生成唯一连接ID，用于连接跟踪和管理
	connID := fmt.Sprintf("ws_%d_%d", time.Now().UnixNano(), atomic.AddInt64(&h.connCounter, 1))

	// 创建WebSocket升级器，配置缓冲区和压缩等选项
	upgrader := websocket.Upgrader{
		ReadBufferSize:    wsConfig.ReadBufferSize,
		WriteBufferSize:   wsConfig.WriteBufferSize,
		EnableCompression: wsConfig.EnableCompression,
		Subprotocols:      wsConfig.Subprotocols,
		CheckOrigin: func(r *http.Request) bool {
			return true // 允许所有来源，生产环境建议配置具体的Origin检查
		},
	}

	// 第一步：将HTTP连接升级为WebSocket连接（客户端侧）
	clientConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		// HTTP协议升级失败，记录响应时间并返回错误
		ctx.SetForwardResponseTime(time.Now())
		atomic.AddInt64(&h.stats.FailedUpgrades, 1)
		return fmt.Errorf("HTTP协议升级为WebSocket失败: %w", err)
	}
	// 重要：连接已被 hijack，立即标记为已响应，防止后续错误处理尝试写入 HTTP 响应
	// 一旦连接被 hijack，就不能再使用标准的 HTTP 响应方法（WriteHeader、Write 等）
	ctx.SetResponded()
	// 确保客户端连接在异常情况下能正确关闭，防止资源泄露
	defer func() {
		if clientConn != nil {
			clientConn.Close()
		}
	}()

	// 第二步：构建目标WebSocket服务的URL
	target, err := url.Parse(node.URL)
	if err != nil {
		return fmt.Errorf("解析目标服务URL失败: %w", err)
	}

	// 根据目标服务的协议选择WebSocket协议类型
	wsScheme := "ws"
	if target.Scheme == "https" {
		wsScheme = "wss"
	}

	// 构建完整的WebSocket目标URL，保持路径和查询参数
	targetWSURL := &url.URL{
		Scheme:   wsScheme,
		Host:     target.Host,
		Path:     ctx.Request.URL.Path,
		RawQuery: ctx.Request.URL.RawQuery,
	}

	ctx.SetTargetURL(targetWSURL.String())

	// 第三步：连接到目标WebSocket服务
	targetConn, resp, err := h.connectToTarget(targetWSURL, ctx.Request, wsConfig)
	if err != nil {
		ctx.SetForwardResponseTime(time.Now())
		atomic.AddInt64(&h.stats.FailedUpgrades, 1)
		return fmt.Errorf("连接到目标WebSocket服务失败: %w", err)
	}
	// 确保目标连接在异常情况下能正确关闭，防止资源泄露
	defer func() {
		if targetConn != nil {
			targetConn.Close()
		}
	}()

	// 记录响应状态码
	if resp != nil {
		ctx.Set(constants.BackendStatusCode, resp.StatusCode)
		ctx.Set(constants.GatewayStatusCode, resp.StatusCode)
	}

	// 注意：响应已在第244行（升级成功后）标记为已处理
	// 这里不需要再次设置，因为连接已被 hijack，不能再使用标准 HTTP 响应方法

	// 第四步：创建连接管理对象，用于生命周期管理
	connCtx, cancel := context.WithCancel(context.Background())
	conn := &wsConnection{
		id:           connID,
		serviceID:    ctx.GetServiceID(),
		clientConn:   clientConn,
		targetConn:   targetConn,
		ctx:          connCtx,
		cancel:       cancel,
		createdAt:    time.Now(),
		lastActivity: time.Now(),
	}

	// 注册连接到管理器，更新统计信息
	h.connections.Store(connID, conn)
	atomic.AddInt64(&h.stats.ActiveConnections, 1)
	atomic.AddInt64(&h.stats.TotalConnections, 1)

	// 使用defer确保连接最终会被清理，这是防止资源泄露的关键
	defer h.cleanupConnection(connID, conn)

	// 第五步：配置连接参数（超时、消息大小限制、心跳等）
	h.configureConnection(conn, wsConfig)

	// 第六步：开始双向消息代理，此时连接资源的生命周期由defer管理
	// 取消defer的资源清理，因为连接将由proxyMessages方法管理
	clientConn = nil // 防止defer关闭连接
	targetConn = nil // 防止defer关闭连接

	// 开始消息代理，这将阻塞直到连接关闭
	return h.proxyMessages(conn, wsConfig, ctx)
}

// connectToTarget 连接到目标WebSocket服务
func (h *WebSocketUpgradeHandler) connectToTarget(targetURL *url.URL, req *http.Request, config *WebSocketConfig) (*websocket.Conn, *http.Response, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		EnableCompression: config.EnableCompression,
		Subprotocols:      config.Subprotocols,
	}

	// 复制请求头部
	headers := make(http.Header)
	for name, values := range req.Header {
		if h.isWebSocketSpecificHeader(name) || isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			headers.Add(name, value)
		}
	}

	// 设置代理头部
	h.setProxyHeaders(req, headers, targetURL.Host)

	return dialer.Dial(targetURL.String(), headers)
}

// configureConnection 配置WebSocket连接参数
// 设置超时、消息大小限制和心跳检测，确保连接的稳定性和安全性
func (h *WebSocketUpgradeHandler) configureConnection(conn *wsConnection, config *WebSocketConfig) {
	// 设置消息大小限制，防止恶意大消息攻击
	if config.MaxMessageSize > 0 {
		conn.clientConn.SetReadLimit(config.MaxMessageSize)
		conn.targetConn.SetReadLimit(config.MaxMessageSize)
	}

	// 设置初始读取超时时间
	if config.ReadTimeout > 0 {
		conn.clientConn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
		conn.targetConn.SetReadDeadline(time.Now().Add(config.ReadTimeout))
	}

	// 启用心跳检测机制，保持连接活跃
	if config.PingInterval > 0 && config.PongTimeout > 0 {
		h.setupPingPong(conn, config)
	}
}

// setupPingPong 设置WebSocket心跳检测机制
// 心跳检测用于保持连接活跃，及时发现断开的连接，防止僵尸连接占用资源
func (h *WebSocketUpgradeHandler) setupPingPong(conn *wsConnection, config *WebSocketConfig) {
	// 为客户端连接设置pong消息处理器
	// 当收到pong响应时，更新读取超时时间
	conn.clientConn.SetPongHandler(func(string) error {
		if config.PongTimeout > 0 {
			conn.clientConn.SetReadDeadline(time.Now().Add(config.PongTimeout))
		}
		conn.lastActivity = time.Now() // 更新最后活跃时间
		return nil
	})

	// 为目标服务连接设置pong消息处理器
	conn.targetConn.SetPongHandler(func(string) error {
		if config.PongTimeout > 0 {
			conn.targetConn.SetReadDeadline(time.Now().Add(config.PongTimeout))
		}
		conn.lastActivity = time.Now() // 更新最后活跃时间
		return nil
	})

	// 启动心跳定时器协程，定期发送ping消息
	go func() {
		ticker := time.NewTicker(config.PingInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 向客户端发送ping消息
				if err := conn.clientConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					// ping失败说明连接已断开，触发连接清理
					conn.cancel()
					return
				}
				// 向目标服务发送ping消息
				if err := conn.targetConn.WriteMessage(websocket.PingMessage, nil); err != nil {
					// ping失败说明连接已断开，触发连接清理
					conn.cancel()
					return
				}
			case <-conn.ctx.Done():
				// 连接上下文被取消，退出心跳协程
				return
			}
		}
	}()
}

// proxyMessages 双向代理WebSocket消息传输
// 启动两个协程分别处理客户端->目标服务和目标服务->客户端的消息转发
// 任一方向连接断开时，整个代理连接都会终止，确保资源及时清理
func (h *WebSocketUpgradeHandler) proxyMessages(conn *wsConnection, config *WebSocketConfig, ctx *core.Context) error {
	var wg sync.WaitGroup
	wg.Add(2)

	// 使用通道接收错误，便于处理
	errChan := make(chan error, 2)

	// 启动客户端到目标服务的消息转发协程
	go func() {
		defer wg.Done()
		err := h.copyMessages(conn.clientConn, conn.targetConn, conn, true, config)
		errChan <- err
	}()

	// 启动目标服务到客户端的消息转发协程
	go func() {
		defer wg.Done()
		err := h.copyMessages(conn.targetConn, conn.clientConn, conn, false, config)
		errChan <- err
	}()

	// 等待任一方向的连接结束
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// 处理第一个返回的错误或正常结束
	// 正常的连接关闭不视为错误，异常关闭才记录为错误
	err := <-errChan
	ctx.SetForwardResponseTime(time.Now())

	if err != nil && !websocket.IsUnexpectedCloseError(err,
		websocket.CloseGoingAway,
		websocket.CloseAbnormalClosure,
		websocket.CloseNormalClosure) {
		return fmt.Errorf("WebSocket消息代理异常: %w", err)
	}

	return nil
}

// copyMessages 复制消息
func (h *WebSocketUpgradeHandler) copyMessages(src, dst *websocket.Conn, conn *wsConnection, isClientToTarget bool, config *WebSocketConfig) error {
	defer conn.cancel()

	for {
		select {
		case <-conn.ctx.Done():
			return nil
		default:
		}

		messageType, message, err := src.ReadMessage()
		if err != nil {
			return err
		}

		// 更新活动时间
		conn.lastActivity = time.Now()

		// 更新读取超时
		if config.ReadTimeout > 0 {
			src.SetReadDeadline(time.Now().Add(config.ReadTimeout))
		}

		// 写入消息
		if config.WriteTimeout > 0 {
			dst.SetWriteDeadline(time.Now().Add(config.WriteTimeout))
		}

		err = dst.WriteMessage(messageType, message)
		if err != nil {
			return err
		}

		// 更新统计
		messageSize := int64(len(message))
		if isClientToTarget {
			atomic.AddInt64(&conn.clientToTarget, 1)
			atomic.AddInt64(&h.stats.BytesReceived, messageSize)
		} else {
			atomic.AddInt64(&conn.targetToClient, 1)
			atomic.AddInt64(&h.stats.BytesSent, messageSize)
		}
	}
}

// cleanupConnection 清理WebSocket连接资源
// 这是防止资源泄露的关键方法，确保所有连接资源都被正确释放
// 包括从连接管理器中移除、取消上下文、关闭连接等
func (h *WebSocketUpgradeHandler) cleanupConnection(connID string, conn *wsConnection) {
	// 从连接管理器中移除连接，避免内存泄露
	h.connections.Delete(connID)
	// 更新活跃连接计数
	atomic.AddInt64(&h.stats.ActiveConnections, -1)

	// 取消连接上下文，通知所有相关的goroutine退出
	if conn.cancel != nil {
		conn.cancel()
	}

	// 优雅关闭客户端WebSocket连接
	if conn.clientConn != nil {
		// 发送关闭帧，通知客户端连接即将关闭
		_ = conn.clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// 关闭连接，释放网络资源
		_ = conn.clientConn.Close()
	}

	// 优雅关闭目标服务WebSocket连接
	if conn.targetConn != nil {
		// 发送关闭帧，通知目标服务连接即将关闭
		_ = conn.targetConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		// 关闭连接，释放网络资源
		_ = conn.targetConn.Close()
	}
}

// Shutdown 优雅关闭WebSocket升级处理器
// 在网关重启或关闭时调用，确保所有WebSocket连接都被正确关闭
// 防止连接资源泄露，这对于热重载场景尤其重要
func (h *WebSocketUpgradeHandler) Shutdown(timeout time.Duration) error {
	// 创建带超时的上下文，防止关闭过程无限阻塞
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// 收集所有当前活跃的连接
	var connections []*wsConnection
	h.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*wsConnection); ok {
			connections = append(connections, conn)
		}
		return true
	})

	// 如果没有活跃连接，直接返回
	if len(connections) == 0 {
		return nil
	}

	// 并发发送关闭消息给所有连接，提高关闭效率
	for _, conn := range connections {
		go func(c *wsConnection) {
			// 向客户端发送服务器重启的关闭消息
			if c.clientConn != nil {
				_ = c.clientConn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, "服务器重启"))
			}
			// 向目标服务发送关闭消息
			if c.targetConn != nil {
				_ = c.targetConn.WriteMessage(websocket.CloseMessage,
					websocket.FormatCloseMessage(websocket.CloseGoingAway, "服务器重启"))
			}
			// 取消连接上下文，触发消息代理协程退出
			if c.cancel != nil {
				c.cancel()
			}
		}(conn)
	}

	// 等待所有连接优雅关闭，定期检查连接状态
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// 超时后强制清理剩余连接，防止资源泄露
			h.connections.Range(func(key, value interface{}) bool {
				if conn, ok := value.(*wsConnection); ok {
					h.cleanupConnection(key.(string), conn)
				}
				return true
			})
			return fmt.Errorf("WebSocket连接关闭超时，已强制清理剩余连接")
		case <-ticker.C:
			// 检查是否所有连接都已关闭
			if atomic.LoadInt64(&h.stats.ActiveConnections) == 0 {
				return nil
			}
		}
	}
}

// GetStats 获取统计信息
func (h *WebSocketUpgradeHandler) GetStats() WebSocketStats {
	return WebSocketStats{
		ActiveConnections: atomic.LoadInt64(&h.stats.ActiveConnections),
		TotalConnections:  atomic.LoadInt64(&h.stats.TotalConnections),
		FailedUpgrades:    atomic.LoadInt64(&h.stats.FailedUpgrades),
		BytesReceived:     atomic.LoadInt64(&h.stats.BytesReceived),
		BytesSent:         atomic.LoadInt64(&h.stats.BytesSent),
	}
}

// GetConnectionInfo 获取连接详细信息
func (h *WebSocketUpgradeHandler) GetConnectionInfo() []map[string]interface{} {
	var connections []map[string]interface{}

	h.connections.Range(func(key, value interface{}) bool {
		if conn, ok := value.(*wsConnection); ok {
			info := map[string]interface{}{
				"id":               conn.id,
				"service_id":       conn.serviceID,
				"created_at":       conn.createdAt,
				"last_activity":    conn.lastActivity,
				"duration":         time.Since(conn.createdAt).Seconds(),
				"idle_duration":    time.Since(conn.lastActivity).Seconds(),
				"client_to_target": atomic.LoadInt64(&conn.clientToTarget),
				"target_to_client": atomic.LoadInt64(&conn.targetToClient),
			}
			connections = append(connections, info)
		}
		return true
	})

	return connections
}

// 辅助方法
func (h *WebSocketUpgradeHandler) isWebSocketSpecificHeader(name string) bool {
	switch strings.ToLower(name) {
	case "connection", "upgrade", "sec-websocket-key", "sec-websocket-version",
		"sec-websocket-protocol", "sec-websocket-extensions":
		return true
	default:
		return false
	}
}

func (h *WebSocketUpgradeHandler) setProxyHeaders(req *http.Request, headers http.Header, targetHost string) {
	// 设置X-Forwarded-* 头部
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		headers.Set("X-Forwarded-For", xff+", "+h.getClientIP(req))
	} else {
		headers.Set("X-Forwarded-For", h.getClientIP(req))
	}

	headers.Set("X-Real-IP", h.getClientIP(req))

	scheme := "ws"
	if req.TLS != nil {
		scheme = "wss"
	}
	headers.Set("X-Forwarded-Proto", scheme)
	headers.Set("X-Forwarded-Host", req.Host)

	if headers.Get("User-Agent") == "" {
		headers.Set("User-Agent", "Gateway-Gateway/1.0")
	}
}

func (h *WebSocketUpgradeHandler) getClientIP(req *http.Request) string {
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
