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
	"gateway/internal/gateway/logwrite"

	"github.com/gorilla/websocket"
)

// WebSocketStats 汇总当前代际WebSocket会话的连接和流量统计。
type WebSocketStats struct {
	ActiveConnections int64 `json:"active_connections"`
	TotalConnections  int64 `json:"total_connections"`
	FailedUpgrades    int64 `json:"failed_upgrades"`
	BytesReceived     int64 `json:"bytes_received"`
	BytesSent         int64 `json:"bytes_sent"`
	NormalClosed      int64 `json:"normal_closed"`
	ShutdownClosed    int64 `json:"shutdown_closed"`
	ForcedClosed      int64 `json:"forced_closed"`
	ErrorClosed       int64 `json:"error_closed"`
}

// WebSocketBridge 是HTTP Upgrade和专项WebSocket代理共用的会话核心。
type WebSocketBridge struct {
	serviceManager  service.ServiceManager
	baseConfig      *WebSocketConfig
	configOverrides map[string]interface{}
	sessions        sync.Map
	connCounter     atomic.Int64
	active          atomic.Int64
	total           atomic.Int64
	failed          atomic.Int64
	bytesReceived   atomic.Int64
	bytesSent       atomic.Int64
	normalClosed    atomic.Int64
	shutdownClosed  atomic.Int64
	forcedClosed    atomic.Int64
	errorClosed     atomic.Int64
}

type wsOutboundFrame struct {
	messageType int
	payload     []byte
}

type wsEndpoint struct {
	conn     *websocket.Conn
	writes   chan wsOutboundFrame
	lastPing atomic.Int64
	lastPong atomic.Int64
}

// wsBridgeSession 管理一对WebSocket连接及其全部读写协程。
type wsBridgeSession struct {
	id                string
	serviceID         string
	serviceName       string
	client            wsEndpoint
	target            wsEndpoint
	config            WebSocketConfig
	ctx               context.Context
	cancel            context.CancelFunc
	createdAt         time.Time
	lastActivity      atomic.Int64
	clientCount       atomic.Int64
	targetCount       atomic.Int64
	bytesFromClient   atomic.Int64
	bytesFromTarget   atomic.Int64
	requestSampleLim  int
	responseSampleLim int
	sampleMu          sync.Mutex
	requestSample     []byte
	responseSample    []byte
	cleanupOnce       sync.Once
	requestedClose    atomic.Bool
	forcedClose       atomic.Bool
	wg                sync.WaitGroup
	errCh             chan error
}

// NewWebSocketBridge 创建共享WebSocket会话核心。
// baseConfig非空时代表代理级完整配置，其优先级高于服务元数据。
func NewWebSocketBridge(serviceManager service.ServiceManager, baseConfig *WebSocketConfig) *WebSocketBridge {
	var copied *WebSocketConfig
	if baseConfig != nil {
		value := *baseConfig
		copied = &value
	}
	return &WebSocketBridge{
		serviceManager: serviceManager,
		baseConfig:     copied,
	}
}

// NewWebSocketBridgeWithOverrides 创建按字段覆盖服务元数据的Bridge。
func NewWebSocketBridgeWithOverrides(serviceManager service.ServiceManager, overrides map[string]interface{}) *WebSocketBridge {
	copied := make(map[string]interface{}, len(overrides))
	for key, value := range overrides {
		copied[key] = value
	}
	return &WebSocketBridge{
		serviceManager:  serviceManager,
		configOverrides: copied,
	}
}

// IsUpgradeRequest 判断请求是否符合RFC 6455升级握手的基础条件。
func (b *WebSocketBridge) IsUpgradeRequest(req *http.Request) bool {
	return req.Method == http.MethodGet &&
		headerContainsToken(req.Header, "Connection", "upgrade") &&
		strings.EqualFold(req.Header.Get("Upgrade"), "websocket") &&
		req.Header.Get("Sec-WebSocket-Key") != "" &&
		req.Header.Get("Sec-WebSocket-Version") == "13"
}

// Proxy 建立上游连接后升级客户端，并阻塞到双向会话完整退出。
// 会话结束后写入访问日志所需的字节统计、关闭原因，以及按配置采样的首帧报文体；并写入简短后端追踪。
func (b *WebSocketBridge) Proxy(ctx *core.Context, proxyName, proxyType string) error {
	requestStartTime := time.Now()
	serviceIDs := ctx.GetServiceIDs()
	if len(serviceIDs) == 0 {
		b.failed.Add(1)
		return fmt.Errorf("服务ID不能为空")
	}
	serviceID := serviceIDs[0]
	serviceConfig, exists := b.serviceManager.GetService(serviceID)
	if !exists || serviceConfig == nil {
		b.failed.Add(1)
		return fmt.Errorf("服务配置不存在: %s", serviceID)
	}
	serviceName := serviceConfig.Name
	node, err := b.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		b.failed.Add(1)
		return fmt.Errorf("选择目标节点失败: %w", err)
	}
	config := b.resolveConfig(serviceConfig)
	if len(config.Subprotocols) == 0 {
		config.Subprotocols = websocket.Subprotocols(ctx.Request)
	}
	targetURL, err := b.buildTargetURL(ctx, node.URL)
	if err != nil {
		b.failed.Add(1)
		return err
	}
	targetURLStr := targetURL.String()
	ctx.SetTargetURL(targetURLStr)
	ctx.Set(constants.ContextKeyServiceDefinitionName, proxyName)
	ctx.Set(constants.ContextKeyProxyType, proxyType)

	var responseStatusCode int
	var responseHeaders map[string][]string
	var responseErr error

	targetConn, response, err := b.connectTarget(targetURL, ctx.Request, &config)
	if err != nil {
		b.failed.Add(1)
		responseErr = err
		if response != nil {
			responseStatusCode = response.StatusCode
			responseHeaders = cloneHeaderMap(response.Header)
			if response.Body != nil {
				_ = response.Body.Close()
			}
		}
		b.writeWebSocketBackendTrace(ctx, serviceID, serviceName, targetURLStr, requestStartTime,
			responseStatusCode, responseHeaders, nil, 0, 0, responseErr)
		return fmt.Errorf("连接WebSocket上游失败: %w", err)
	}
	if response != nil {
		responseStatusCode = response.StatusCode
		responseHeaders = cloneHeaderMap(response.Header)
	}
	upstreamSubprotocol := targetConn.Subprotocol()
	upgrader := websocket.Upgrader{
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		EnableCompression: config.EnableCompression,
		Subprotocols:      config.Subprotocols,
		// 请求已经通过网关CORS处理器；此处不重复维护第二套Origin策略。
		CheckOrigin: func(*http.Request) bool { return true },
	}
	if upstreamSubprotocol != "" {
		upgrader.Subprotocols = []string{upstreamSubprotocol}
	}
	clientConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		b.failed.Add(1)
		_ = targetConn.Close()
		responseErr = err
		b.writeWebSocketBackendTrace(ctx, serviceID, serviceName, targetURLStr, requestStartTime,
			responseStatusCode, responseHeaders, nil, 0, 0, responseErr)
		return fmt.Errorf("升级客户端WebSocket连接失败: %w", err)
	}

	ctx.SetResponded()
	ctx.Set(constants.GatewayStatusCode, http.StatusSwitchingProtocols)
	if responseStatusCode > 0 {
		ctx.Set(constants.BackendStatusCode, responseStatusCode)
	}

	sessionCtx, cancel := context.WithCancel(context.Background())
	session := &wsBridgeSession{
		id:                fmt.Sprintf("ws_%d_%d", time.Now().UnixNano(), b.connCounter.Add(1)),
		serviceID:         serviceID,
		serviceName:       serviceName,
		client:            wsEndpoint{conn: clientConn, writes: make(chan wsOutboundFrame, 64)},
		target:            wsEndpoint{conn: targetConn, writes: make(chan wsOutboundFrame, 64)},
		config:            config,
		ctx:               sessionCtx,
		cancel:            cancel,
		createdAt:         time.Now(),
		requestSampleLim:  resolveBodySampleLimit(ctx, false),
		responseSampleLim: resolveBodySampleLimit(ctx, true),
		errCh:             make(chan error, 8),
	}
	session.lastActivity.Store(time.Now().UnixNano())
	b.sessions.Store(session.id, session)
	b.active.Add(1)
	b.total.Add(1)
	defer b.cleanupSession(session)

	err = b.runSession(session)
	closeReason := session.closeReason(err)
	b.applyWebSocketSessionLog(ctx, session, closeReason)
	switch closeReason {
	case "normal", "connection_closed":
		b.normalClosed.Add(1)
	case "server_shutdown":
		b.shutdownClosed.Add(1)
	case "force_closed":
		b.forcedClosed.Add(1)
	default:
		b.errorClosed.Add(1)
	}

	bytesRx := session.bytesFromClient.Load()
	bytesTx := session.bytesFromTarget.Load()
	var sessionErr error
	if closeReason == "error" {
		sessionErr = err
		responseErr = err
	}
	b.writeWebSocketBackendTrace(ctx, serviceID, serviceName, targetURLStr, requestStartTime,
		responseStatusCode, responseHeaders, session.responseSampleSnapshot(),
		clampInt64ToInt(bytesRx), clampInt64ToInt(bytesTx), responseErr)
	return sessionErr
}

// applyWebSocketSessionLog 将会话流量、关闭原因和首帧采样写入网关上下文，供访问日志异步读取。
func (b *WebSocketBridge) applyWebSocketSessionLog(ctx *core.Context, session *wsBridgeSession, closeReason string) {
	bytesRx := session.bytesFromClient.Load()
	bytesTx := session.bytesFromTarget.Load()
	ctx.Set(constants.ContextKeyWebSocketCloseReason, closeReason)
	ctx.Set(constants.ContextKeyWebSocketBytesReceived, bytesRx)
	ctx.Set(constants.ContextKeyWebSocketBytesSent, bytesTx)
	// 访问日志 responseSize 对应“下游可见响应流量”，取上游发往客户端字节。
	ctx.Set(constants.ContextKeyResponseSize, clampInt64ToInt(bytesTx))
	ctx.Set(constants.ContextKeySnapshotRequestSize, clampInt64ToInt(bytesRx))

	session.sampleMu.Lock()
	reqSample := append([]byte(nil), session.requestSample...)
	respSample := append([]byte(nil), session.responseSample...)
	session.sampleMu.Unlock()
	if len(reqSample) > 0 {
		ctx.Set("request_body", reqSample)
	}
	if len(respSample) > 0 {
		ctx.Set("response_body", respSample)
	}
}

// writeWebSocketBackendTrace 写入一次WebSocket握手/会话的后端追踪。
// 请求报文取客户端首帧采样，响应报文取上游首帧采样；字节数字段取会话累计流量。
func (b *WebSocketBridge) writeWebSocketBackendTrace(
	ctx *core.Context,
	serviceID, serviceName, targetURL string,
	requestStartTime time.Time,
	statusCode int,
	responseHeaders map[string][]string,
	responseBody []byte,
	requestSize, responseSize int,
	responseErr error,
) {
	var forwardBody []byte
	if bodyData, exists := ctx.Get("request_body"); exists {
		if bodyBytes, ok := bodyData.([]byte); ok {
			forwardBody = bodyBytes
		}
	}
	// 后端追踪响应大小优先使用会话累计下行字节，便于与访问日志对齐。
	if responseSize > 0 {
		ctx.Set(constants.ContextKeyResponseSize, responseSize)
	}
	_ = logwrite.WriteBackendTraceLogSync(
		"",
		ctx,
		serviceID,
		"",
		http.MethodGet,
		targetURL,
		requestSize,
		requestStartTime,
		time.Now(),
		statusCode,
		responseHeaders,
		responseBody,
		nil,
		forwardBody,
		responseErr,
		serviceName,
		0,
	)
}

func (b *WebSocketBridge) resolveConfig(serviceConfig *service.ServiceConfig) WebSocketConfig {
	config := DefaultWebSocketConfig
	if serviceConfig != nil && len(serviceConfig.ServiceMetadata) > 0 {
		values := make(map[string]interface{}, len(serviceConfig.ServiceMetadata))
		for key, value := range serviceConfig.ServiceMetadata {
			values[key] = value
		}
		NewWebSocketConfigParser().ParseConfig(values, &config)
	}
	if b.baseConfig != nil {
		config = *b.baseConfig
	}
	if len(b.configOverrides) > 0 {
		NewWebSocketConfigParser().ParseConfig(b.configOverrides, &config)
	}
	return config
}

func (b *WebSocketBridge) buildTargetURL(ctx *core.Context, targetValue string) (*url.URL, error) {
	target, err := url.Parse(targetValue)
	if err != nil {
		return nil, fmt.Errorf("解析目标服务URL失败: %w", err)
	}
	scheme := "ws"
	if target.Scheme == "https" || target.Scheme == "wss" {
		scheme = "wss"
	}
	return &url.URL{
		Scheme:   scheme,
		Host:     target.Host,
		Path:     buildTargetPath(ctx, target.Path),
		RawQuery: buildTargetQuery(target.RawQuery, ctx.Request.URL.RawQuery),
	}, nil
}

func (b *WebSocketBridge) connectTarget(targetURL *url.URL, req *http.Request, config *WebSocketConfig) (*websocket.Conn, *http.Response, error) {
	dialer := websocket.Dialer{
		HandshakeTimeout:  10 * time.Second,
		ReadBufferSize:    config.ReadBufferSize,
		WriteBufferSize:   config.WriteBufferSize,
		EnableCompression: config.EnableCompression,
		Subprotocols:      config.Subprotocols,
	}
	headers := make(http.Header)
	for name, values := range req.Header {
		if isWebSocketHeader(name) || isHopByHopHeader(name) {
			continue
		}
		for _, value := range values {
			headers.Add(name, value)
		}
	}
	setWebSocketProxyHeaders(req, headers)
	return dialer.DialContext(req.Context(), targetURL.String(), headers)
}

func (b *WebSocketBridge) runSession(session *wsBridgeSession) error {
	b.configureSession(session)
	session.wg.Add(4)
	go b.writePump(session, &session.client)
	go b.writePump(session, &session.target)
	go b.readPump(session, &session.client, &session.target, true)
	go b.readPump(session, &session.target, &session.client, false)
	if session.config.PingInterval > 0 && session.config.PongTimeout > 0 {
		session.wg.Add(1)
		go b.heartbeatPump(session)
	}

	err := <-session.errCh
	session.cancel()
	// 关闭底层连接以解除另一方向可能阻塞的ReadMessage。
	_ = session.client.conn.Close()
	_ = session.target.conn.Close()
	session.wg.Wait()
	return err
}

func (b *WebSocketBridge) configureSession(session *wsBridgeSession) {
	for _, endpoint := range []*wsEndpoint{&session.client, &session.target} {
		b.configureEndpoint(session, endpoint)
	}
}

func (b *WebSocketBridge) configureEndpoint(session *wsBridgeSession, endpoint *wsEndpoint) {
	now := time.Now().UnixNano()
	endpoint.lastPing.Store(now)
	endpoint.lastPong.Store(now)
	if session.config.MaxMessageSize > 0 {
		endpoint.conn.SetReadLimit(session.config.MaxMessageSize)
	}
	if window := session.readDeadlineWindow(); window > 0 {
		_ = endpoint.conn.SetReadDeadline(time.Now().Add(window))
	}
	endpoint.conn.SetPongHandler(func(string) error {
		now := time.Now()
		session.lastActivity.Store(now.UnixNano())
		endpoint.lastPong.Store(now.UnixNano())
		if window := session.readDeadlineWindow(); window > 0 {
			return endpoint.conn.SetReadDeadline(time.Now().Add(window))
		}
		return nil
	})
}

func (s *wsBridgeSession) readDeadlineWindow() time.Duration {
	window := s.config.ReadTimeout
	heartbeatWindow := s.config.PingInterval + s.config.PongTimeout
	if s.config.PingInterval > 0 && s.config.PongTimeout > 0 && heartbeatWindow > window {
		window = heartbeatWindow
	}
	return window
}

func (b *WebSocketBridge) writePump(session *wsBridgeSession, endpoint *wsEndpoint) {
	defer session.wg.Done()
	for {
		select {
		case <-session.ctx.Done():
			return
		case frame := <-endpoint.writes:
			if session.config.WriteTimeout > 0 {
				_ = endpoint.conn.SetWriteDeadline(time.Now().Add(session.config.WriteTimeout))
			}
			if err := endpoint.conn.WriteMessage(frame.messageType, frame.payload); err != nil {
				session.reportError(err)
				return
			}
			session.lastActivity.Store(time.Now().UnixNano())
		}
	}
}

func (b *WebSocketBridge) readPump(session *wsBridgeSession, src, dst *wsEndpoint, clientToTarget bool) {
	defer session.wg.Done()
	for {
		messageType, payload, err := src.conn.ReadMessage()
		if err != nil {
			session.reportError(err)
			return
		}
		session.lastActivity.Store(time.Now().UnixNano())
		if window := session.readDeadlineWindow(); window > 0 {
			_ = src.conn.SetReadDeadline(time.Now().Add(window))
		}
		// 仅采样首条文本/二进制业务帧，不缓存整会话流量。
		session.maybeSampleFrame(clientToTarget, messageType, payload)
		frame := wsOutboundFrame{messageType: messageType, payload: payload}
		select {
		case <-session.ctx.Done():
			return
		case dst.writes <- frame:
		}
		payloadLen := int64(len(payload))
		if clientToTarget {
			session.clientCount.Add(1)
			session.bytesFromClient.Add(payloadLen)
			b.bytesReceived.Add(payloadLen)
		} else {
			session.targetCount.Add(1)
			session.bytesFromTarget.Add(payloadLen)
			b.bytesSent.Add(payloadLen)
		}
	}
}

func (b *WebSocketBridge) heartbeatPump(session *wsBridgeSession) {
	defer session.wg.Done()
	checkInterval := session.config.PingInterval
	if session.config.PongTimeout < checkInterval {
		checkInterval = session.config.PongTimeout
	}
	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()
	for {
		select {
		case <-session.ctx.Done():
			return
		case <-ticker.C:
			if session.requestedClose.Load() {
				return
			}
			if err := session.checkAndSendPing(&session.client); err != nil {
				session.reportError(err)
				return
			}
			if err := session.checkAndSendPing(&session.target); err != nil {
				session.reportError(err)
				return
			}
		}
	}
}

func (s *wsBridgeSession) checkAndSendPing(endpoint *wsEndpoint) error {
	now := time.Now()
	lastPing := time.Unix(0, endpoint.lastPing.Load())
	lastPong := time.Unix(0, endpoint.lastPong.Load())
	if lastPing.After(lastPong) && now.Sub(lastPing) >= s.config.PongTimeout {
		return fmt.Errorf("WebSocket Pong响应超时")
	}
	if now.Sub(lastPing) < s.config.PingInterval {
		return nil
	}
	if !s.enqueue(endpoint, wsOutboundFrame{messageType: websocket.PingMessage}) {
		return fmt.Errorf("WebSocket心跳写队列已满")
	}
	endpoint.lastPing.Store(now.UnixNano())
	return nil
}

func (s *wsBridgeSession) enqueue(endpoint *wsEndpoint, frame wsOutboundFrame) bool {
	select {
	case <-s.ctx.Done():
		return false
	case endpoint.writes <- frame:
		return true
	default:
		return false
	}
}

func (s *wsBridgeSession) reportError(err error) {
	select {
	case s.errCh <- err:
	default:
	}
}

func (s *wsBridgeSession) requestClose(code int, reason string) {
	if !s.requestedClose.CompareAndSwap(false, true) {
		return
	}
	frame := wsOutboundFrame{
		messageType: websocket.CloseMessage,
		payload:     websocket.FormatCloseMessage(code, reason),
	}
	if !s.enqueue(&s.client, frame) || !s.enqueue(&s.target, frame) {
		s.forceClose()
	}
}

func (s *wsBridgeSession) forceClose() {
	s.forcedClose.Store(true)
	s.cancel()
	_ = s.client.conn.Close()
	_ = s.target.conn.Close()
}

func (b *WebSocketBridge) cleanupSession(session *wsBridgeSession) {
	session.cleanupOnce.Do(func() {
		session.forceClose()
		b.sessions.Delete(session.id)
		b.active.Add(-1)
	})
}

// Shutdown 向当前所有会话发送GoingAway，并等待其在ctx期限内退出。
func (b *WebSocketBridge) Shutdown(ctx context.Context) error {
	b.sessions.Range(func(_, value interface{}) bool {
		if session, ok := value.(*wsBridgeSession); ok {
			session.requestClose(websocket.CloseGoingAway, "服务器正在重载")
		}
		return true
	})
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if b.active.Load() == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			b.ForceClose()
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

// ForceClose 立即关闭全部会话，用于优雅排空超时后的最终资源回收。
func (b *WebSocketBridge) ForceClose() {
	b.sessions.Range(func(_, value interface{}) bool {
		if session, ok := value.(*wsBridgeSession); ok {
			session.forceClose()
		}
		return true
	})
}

// GetStats 返回原子快照。
func (b *WebSocketBridge) GetStats() WebSocketStats {
	return WebSocketStats{
		ActiveConnections: b.active.Load(),
		TotalConnections:  b.total.Load(),
		FailedUpgrades:    b.failed.Load(),
		BytesReceived:     b.bytesReceived.Load(),
		BytesSent:         b.bytesSent.Load(),
		NormalClosed:      b.normalClosed.Load(),
		ShutdownClosed:    b.shutdownClosed.Load(),
		ForcedClosed:      b.forcedClosed.Load(),
		ErrorClosed:       b.errorClosed.Load(),
	}
}

// GetConnectionInfo 返回当前会话的轻量运行时信息。
func (b *WebSocketBridge) GetConnectionInfo() []map[string]interface{} {
	connections := make([]map[string]interface{}, 0)
	now := time.Now()
	b.sessions.Range(func(_, value interface{}) bool {
		if session, ok := value.(*wsBridgeSession); ok {
			lastActivity := time.Unix(0, session.lastActivity.Load())
			connections = append(connections, map[string]interface{}{
				"id":               session.id,
				"service_id":       session.serviceID,
				"created_at":       session.createdAt,
				"last_activity":    lastActivity,
				"duration":         now.Sub(session.createdAt).Seconds(),
				"idle_duration":    now.Sub(lastActivity).Seconds(),
				"client_to_target": session.clientCount.Load(),
				"target_to_client": session.targetCount.Load(),
			})
		}
		return true
	})
	return connections
}

func headerContainsToken(header http.Header, name, token string) bool {
	for _, value := range header.Values(name) {
		for _, candidate := range strings.Split(value, ",") {
			if strings.EqualFold(strings.TrimSpace(candidate), token) {
				return true
			}
		}
	}
	return false
}

func setWebSocketProxyHeaders(req *http.Request, headers http.Header) {
	clientIP := req.RemoteAddr
	if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		clientIP = ip
	}
	if forwarded := req.Header.Get("X-Forwarded-For"); forwarded != "" {
		headers.Set("X-Forwarded-For", forwarded+", "+clientIP)
	} else {
		headers.Set("X-Forwarded-For", clientIP)
	}
	headers.Set("X-Real-IP", clientIP)
	scheme := "ws"
	if req.TLS != nil {
		scheme = "wss"
	}
	headers.Set("X-Forwarded-Proto", scheme)
	headers.Set("X-Forwarded-Host", req.Host)
}

func isWebSocketHeader(name string) bool {
	switch strings.ToLower(name) {
	case "upgrade", "connection", "sec-websocket-key", "sec-websocket-version",
		"sec-websocket-protocol", "sec-websocket-extensions":
		return true
	default:
		return false
	}
}

func (s *wsBridgeSession) closeReason(err error) string {
	if s.forcedClose.Load() {
		return "force_closed"
	}
	if s.requestedClose.Load() {
		return "server_shutdown"
	}
	if err == nil || websocket.IsCloseError(err,
		websocket.CloseNormalClosure,
		websocket.CloseGoingAway) {
		return "normal"
	}
	if websocket.IsCloseError(err,
		websocket.CloseAbnormalClosure,
		websocket.CloseNoStatusReceived) {
		return "connection_closed"
	}
	return "error"
}

// maybeSampleFrame 在配置允许时缓存方向上的首条业务帧前缀，供日志记录报文体。
func (s *wsBridgeSession) maybeSampleFrame(clientToTarget bool, messageType int, payload []byte) {
	if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
		return
	}
	limit := s.responseSampleLim
	if clientToTarget {
		limit = s.requestSampleLim
	}
	if limit <= 0 || len(payload) == 0 {
		return
	}
	s.sampleMu.Lock()
	defer s.sampleMu.Unlock()
	if clientToTarget {
		if s.requestSample != nil {
			return
		}
		s.requestSample = copyBodySample(payload, limit)
		return
	}
	if s.responseSample != nil {
		return
	}
	s.responseSample = copyBodySample(payload, limit)
}

// responseSampleSnapshot 返回响应首帧采样副本，供后端追踪日志使用。
func (s *wsBridgeSession) responseSampleSnapshot() []byte {
	s.sampleMu.Lock()
	defer s.sampleMu.Unlock()
	if len(s.responseSample) == 0 {
		return nil
	}
	return append([]byte(nil), s.responseSample...)
}

// copyBodySample 复制不超过 limit 字节的报文体前缀。
func copyBodySample(payload []byte, limit int) []byte {
	if limit <= 0 || len(payload) == 0 {
		return nil
	}
	if len(payload) > limit {
		payload = payload[:limit]
	}
	return append([]byte(nil), payload...)
}

// cloneHeaderMap 深拷贝HTTP响应头，避免异步日志读取到被回收的原始map。
func cloneHeaderMap(header http.Header) map[string][]string {
	if len(header) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(header))
	for name, values := range header {
		cloned[name] = append([]string(nil), values...)
	}
	return cloned
}
