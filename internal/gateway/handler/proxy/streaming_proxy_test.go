package proxy

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"

	"github.com/gorilla/websocket"
)

type canceledStreamBody struct {
	ctx context.Context
}

func (b *canceledStreamBody) Read(_ []byte) (int, error) {
	<-b.ctx.Done()
	return 0, b.ctx.Err()
}

func (b *canceledStreamBody) Close() error {
	return nil
}

func TestSSEStreamerCopiesSafeHeadersAndFlushesBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil)
	recorder := httptest.NewRecorder()
	ctx := core.NewContext(recorder, request)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":                {"text/event-stream"},
			"X-Stream-ID":                 {"stream-1"},
			"Connection":                  {"keep-alive"},
			"Transfer-Encoding":           {"chunked"},
			"Access-Control-Allow-Origin": {"*"},
		},
		Body: io.NopCloser(strings.NewReader("data: one\n\ndata: two\n\n")),
	}

	err := newSSEStreamer(4, time.Second, 0).Stream(ctx, response)
	if err != nil {
		t.Fatalf("SSE转发失败: %v", err)
	}
	if recorder.Body.String() != "data: one\n\ndata: two\n\n" {
		t.Fatalf("SSE响应体不匹配: %q", recorder.Body.String())
	}
	if recorder.Header().Get("X-Stream-ID") != "stream-1" {
		t.Fatal("自定义响应头未转发")
	}
	if recorder.Header().Get("Connection") != "" ||
		recorder.Header().Get("Transfer-Encoding") != "" ||
		recorder.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("SSE转发了应由网关或net/http管理的响应头")
	}
	value, exists := ctx.Get(constants.ContextKeySSEBytesStreamed)
	if !exists || value.(int64) != int64(recorder.Body.Len()) {
		t.Fatalf("SSE字节统计不正确: %v", value)
	}
}

func TestSSEStreamerSamplesResponseBodyPrefixAndSetsResponseSize(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil)
	recorder := httptest.NewRecorder()
	ctx := core.NewContext(recorder, request)
	payload := "data: abcdefghijklmnop\n\n"
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(payload)),
	}

	if err := newSSEStreamer(4, time.Second, 8).Stream(ctx, response); err != nil {
		t.Fatalf("SSE转发失败: %v", err)
	}
	bodyData, exists := ctx.Get("response_body")
	if !exists {
		t.Fatal("未写入SSE响应体采样")
	}
	sample, ok := bodyData.([]byte)
	if !ok || string(sample) != payload[:8] {
		t.Fatalf("SSE响应体采样不正确: %q", sample)
	}
	size, ok := ctx.GetInt(constants.ContextKeyResponseSize)
	if !ok || size != len(payload) {
		t.Fatalf("SSE response_size 不正确: %v", size)
	}
	disconnect, _ := ctx.Get(constants.ContextKeySSEDisconnectType)
	if disconnect != sseDisconnectCompleted {
		t.Fatalf("SSE断开原因不正确: %v", disconnect)
	}
}

func TestSSEStreamerTreatsClientCancellationAsNormalDisconnect(t *testing.T) {
	requestCtx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil).WithContext(requestCtx)
	ctx := core.NewContext(httptest.NewRecorder(), request)
	response := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": {"text/event-stream"}},
		Body:       &canceledStreamBody{ctx: requestCtx},
	}
	cancel()

	if err := newSSEStreamer(16, time.Second, 0).Stream(ctx, response); err != nil {
		t.Fatalf("客户端取消不应作为SSE错误返回: %v", err)
	}
	if ctx.HasErrors() {
		t.Fatalf("客户端正常断开不应写入错误列表: %v", ctx.GetErrors())
	}
	value, _ := ctx.Get(constants.ContextKeySSEDisconnectType)
	if value != sseDisconnectClientClosed {
		t.Fatalf("断开原因不正确: %v", value)
	}
}

func TestStreamingTargetPathUsesRouteRewriteAndStripPrefix(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "http://gateway/api/users?id=1", nil)
	ctx := core.NewContext(httptest.NewRecorder(), request)
	ctx.SetMatchedPath("/api")
	ctx.Set(constants.ContextKeyRouteStripPathPrefix, true)
	if actual := buildTargetPath(ctx, "/backend"); actual != "/backend/users" {
		t.Fatalf("移除路由前缀后的目标路径不正确: %s", actual)
	}

	// /api 不得误剥 /apix
	apixRequest := httptest.NewRequest(http.MethodGet, "http://gateway/apix/users", nil)
	apixCtx := core.NewContext(httptest.NewRecorder(), apixRequest)
	apixCtx.SetMatchedPath("/api")
	apixCtx.Set(constants.ContextKeyRouteStripPathPrefix, true)
	if actual := buildTargetPath(apixCtx, "/backend"); actual != "/backend" {
		t.Fatalf("非路径段边界前缀不应剥离: %s", actual)
	}

	// 精确匹配路由前缀时剩余为根路径
	exactRequest := httptest.NewRequest(http.MethodGet, "http://gateway/api", nil)
	exactCtx := core.NewContext(httptest.NewRecorder(), exactRequest)
	exactCtx.SetMatchedPath("/api/")
	exactCtx.Set(constants.ContextKeyRouteStripPathPrefix, true)
	if actual := buildTargetPath(exactCtx, "/backend"); actual != "/backend" {
		t.Fatalf("精确匹配剥前缀后路径不正确: %s", actual)
	}

	ctx.Set(constants.ContextKeyRouteRewritePath, "/stream/events")
	if actual := buildTargetPath(ctx, "/backend"); actual != "/stream/events" {
		t.Fatalf("路由重写路径不正确: %s", actual)
	}
	if actual := buildTargetQuery("token=node", "token=client&id=1"); actual != "token=node&id=1" {
		t.Fatalf("目标查询参数优先级不正确: %s", actual)
	}
}

func TestWebSocketConfigParserAcceptsExplicitZeroTimeouts(t *testing.T) {
	config := DefaultWebSocketConfig
	NewWebSocketConfigParser().ParseConfig(map[string]interface{}{
		"pingInterval": 0,
		"pongTimeout":  0,
		"readTimeout":  0,
		"writeTimeout": 0,
	}, &config)
	if config.PingInterval != 0 || config.PongTimeout != 0 ||
		config.ReadTimeout != 0 || config.WriteTimeout != 0 {
		t.Fatalf("显式零值未关闭WebSocket超时: %+v", config)
	}
}

func TestHTTPProxyStopsAbsoluteTimeoutAfterSSEHeaders(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		writer.WriteHeader(http.StatusOK)
		writer.(http.Flusher).Flush()
		time.Sleep(50 * time.Millisecond)
		_, _ = writer.Write([]byte("data: delayed\n\n"))
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	if err := manager.AddService(&service.ServiceConfig{
		ID:       "sse-service",
		Name:     "sse-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID: "sse-node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
		}},
	}); err != nil {
		t.Fatalf("创建SSE测试服务失败: %v", err)
	}
	httpProxy, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "sse-proxy",
		Config: map[string]interface{}{
			"timeout":        "10ms",
			"readTimeout":    "1s",
			"proxyBuffering": false,
		},
	}, manager)
	if err != nil {
		t.Fatalf("创建SSE代理失败: %v", err)
	}
	defer httpProxy.Close()

	request := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil)
	recorder := httptest.NewRecorder()
	ctx := core.NewContext(recorder, request)
	ctx.SetServiceIDs([]string{"sse-service"})
	startedAt := time.Now()
	if !httpProxy.Handle(ctx) {
		t.Fatalf("SSE代理请求失败: %v", ctx.GetErrors())
	}
	if time.Since(startedAt) < 40*time.Millisecond {
		t.Fatal("SSE仍受普通HTTP绝对总超时限制")
	}
	if recorder.Body.String() != "data: delayed\n\n" {
		t.Fatalf("延迟SSE事件未完整转发: %q", recorder.Body.String())
	}
}

func TestHTTPProxyPropagatesClientCancellation(t *testing.T) {
	upstreamStarted := make(chan struct{})
	upstreamCanceled := make(chan struct{})
	upstream := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		close(upstreamStarted)
		select {
		case <-request.Context().Done():
			close(upstreamCanceled)
		case <-time.After(2 * time.Second):
		}
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	if err := manager.AddService(&service.ServiceConfig{
		ID:       "cancel-service",
		Name:     "cancel-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID: "cancel-node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
		}},
	}); err != nil {
		t.Fatalf("创建取消传播测试服务失败: %v", err)
	}
	httpProxy, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "cancel-proxy",
	}, manager)
	if err != nil {
		t.Fatalf("创建取消传播代理失败: %v", err)
	}
	defer httpProxy.Close()

	requestCtx, cancel := context.WithCancel(context.Background())
	request := httptest.NewRequest(http.MethodGet, "http://gateway/cancel", nil).WithContext(requestCtx)
	ctx := core.NewContext(httptest.NewRecorder(), request)
	ctx.SetServiceIDs([]string{"cancel-service"})
	handleDone := make(chan struct{})
	go func() {
		defer close(handleDone)
		httpProxy.Handle(ctx)
	}()
	<-upstreamStarted
	cancel()

	select {
	case <-upstreamCanceled:
	case <-time.After(time.Second):
		t.Fatal("客户端取消未传播到上游请求")
	}
	select {
	case <-handleDone:
	case <-time.After(time.Second):
		t.Fatal("客户端取消后代理处理未退出")
	}
}

func TestHTTPMultiServiceProxyRejectsSSEResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.Header().Set("Content-Type", "text/event-stream")
		_, _ = writer.Write([]byte("data: unsupported\n\n"))
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	for _, serviceID := range []string{"sse-a", "sse-b"} {
		if err := manager.AddService(&service.ServiceConfig{
			ID:       serviceID,
			Name:     serviceID,
			Strategy: service.RoundRobin,
			Nodes: []*service.NodeConfig{{
				ID: serviceID + "-node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
			}},
		}); err != nil {
			t.Fatalf("创建多服务SSE测试服务失败: %v", err)
		}
	}
	httpProxy, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "multi-sse-proxy",
	}, manager)
	if err != nil {
		t.Fatalf("创建多服务SSE代理失败: %v", err)
	}
	defer httpProxy.Close()

	recorder := httptest.NewRecorder()
	ctx := core.NewContext(recorder, httptest.NewRequest(http.MethodGet, "http://gateway/events", nil))
	ctx.SetServiceIDs([]string{"sse-a", "sse-b"})
	if httpProxy.Handle(ctx) {
		t.Fatal("多服务SSE不应进入响应聚合")
	}
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("多服务SSE状态码不正确: %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), `"domain"`) {
		t.Fatalf("多服务SSE未返回统一网关响应: %s", recorder.Body.String())
	}
}

func TestWebSocketEntrypointsShareBridgeBehavior(t *testing.T) {
	upgrader := websocket.Upgrader{
		CheckOrigin:  func(*http.Request) bool { return true },
		Subprotocols: []string{"gateway.v1"},
	}
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			messageType, payload, readErr := conn.ReadMessage()
			if readErr != nil {
				return
			}
			if writeErr := conn.WriteMessage(messageType, payload); writeErr != nil {
				return
			}
		}
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	err := manager.AddService(&service.ServiceConfig{
		ID:       "echo-service",
		Name:     "echo-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID:      "echo-node",
			URL:     upstream.URL,
			Weight:  1,
			Health:  true,
			Enabled: true,
		}},
	})
	if err != nil {
		t.Fatalf("创建测试服务失败: %v", err)
	}
	if closer, ok := manager.(interface{ Close() error }); ok {
		defer closer.Close()
	}

	httpEntry, err := NewHTTPProxy(ProxyConfig{
		Type:    ProxyTypeHTTP,
		Enabled: true,
		Name:    "http-entry",
	}, manager)
	if err != nil {
		t.Fatalf("创建HTTP入口失败: %v", err)
	}
	websocketEntry, err := NewWebSocketProxy(ProxyConfig{
		Type:    ProxyTypeWebSocket,
		Enabled: true,
		Name:    "websocket-entry",
	}, manager)
	if err != nil {
		t.Fatalf("创建WebSocket入口失败: %v", err)
	}

	entries := []struct {
		name    string
		handler ProxyHandler
	}{
		{name: "http-upgrade", handler: httpEntry},
		{name: "websocket-proxy", handler: websocketEntry},
	}
	for _, entry := range entries {
		t.Run(entry.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
				ctx := core.NewContext(writer, request)
				ctx.SetServiceIDs([]string{"echo-service"})
				// 显式模拟历史路由 enableWebsocket=N，升级仍应成功。
				ctx.Set(constants.ContextKeyRouteEnableWebSocket, false)
				entry.handler.Handle(ctx)
			}))
			defer server.Close()

			target := "ws" + strings.TrimPrefix(server.URL, "http")
			dialer := *websocket.DefaultDialer
			dialer.Subprotocols = []string{"gateway.v1"}
			conn, _, dialErr := dialer.Dial(target, nil)
			if dialErr != nil {
				t.Fatalf("连接入口失败: %v", dialErr)
			}
			defer conn.Close()
			if conn.Subprotocol() != "gateway.v1" {
				t.Fatalf("子协议未同步: %q", conn.Subprotocol())
			}
			payload := []byte("shared-bridge")
			if writeErr := conn.WriteMessage(websocket.TextMessage, payload); writeErr != nil {
				t.Fatalf("写入消息失败: %v", writeErr)
			}
			messageType, received, readErr := conn.ReadMessage()
			if readErr != nil {
				t.Fatalf("读取消息失败: %v", readErr)
			}
			if messageType != websocket.TextMessage || string(received) != string(payload) {
				t.Fatalf("回显结果不一致: type=%d payload=%q", messageType, received)
			}
		})
	}
}

func TestWebSocketBridgeShutdownReleasesSessions(t *testing.T) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	upstream := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		conn, err := upgrader.Upgrade(writer, request, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		for {
			if _, _, err = conn.ReadMessage(); err != nil {
				return
			}
		}
	}))
	defer upstream.Close()

	manager := service.NewServiceManager()
	if closer, ok := manager.(interface{ Close() error }); ok {
		defer closer.Close()
	}
	if err := manager.AddService(&service.ServiceConfig{
		ID:       "shutdown-service",
		Name:     "shutdown-service",
		Strategy: service.RoundRobin,
		Nodes: []*service.NodeConfig{{
			ID: "node", URL: upstream.URL, Weight: 1, Health: true, Enabled: true,
		}},
	}); err != nil {
		t.Fatalf("创建测试服务失败: %v", err)
	}
	config := DefaultWebSocketConfig
	config.PingInterval = 10 * time.Millisecond
	config.PongTimeout = 50 * time.Millisecond
	bridge := NewWebSocketBridge(manager, &config)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := core.NewContext(writer, request)
		ctx.SetServiceIDs([]string{"shutdown-service"})
		_ = bridge.Proxy(ctx, "shutdown-test", string(ProxyTypeWebSocket))
	}))
	defer server.Close()

	target := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(target, nil)
	if err != nil {
		t.Fatalf("连接Bridge失败: %v", err)
	}
	defer conn.Close()
	activeDeadline := time.Now().Add(time.Second)
	for bridge.GetStats().ActiveConnections != 1 && time.Now().Before(activeDeadline) {
		time.Sleep(time.Millisecond)
	}
	if bridge.GetStats().ActiveConnections != 1 {
		t.Fatalf("活跃连接计数不正确: %+v", bridge.GetStats())
	}
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		for {
			if _, _, readErr := conn.ReadMessage(); readErr != nil {
				return
			}
		}
	}()
	for index := 0; index < 20; index++ {
		if err = conn.WriteMessage(websocket.TextMessage, []byte("load")); err != nil {
			t.Fatalf("心跳并发期间写入消息失败: %v", err)
		}
		time.Sleep(2 * time.Millisecond)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	shutdownDone := make(chan error, 1)
	go func() {
		shutdownDone <- bridge.Shutdown(shutdownCtx)
	}()
	if err = <-shutdownDone; err != nil {
		t.Fatalf("Bridge优雅关闭失败: %v", err)
	}
	<-readDone
	if bridge.GetStats().ActiveConnections != 0 {
		t.Fatalf("关闭后连接计数未归零: %+v", bridge.GetStats())
	}
}

func BenchmarkSSEStreamer(b *testing.B) {
	payload := strings.Repeat("data: benchmark\n\n", 1024)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for index := 0; index < b.N; index++ {
		request := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil)
		recorder := httptest.NewRecorder()
		ctx := core.NewContext(recorder, request)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": {"text/event-stream"}},
			Body:       io.NopCloser(strings.NewReader(payload)),
		}
		if err := newSSEStreamer(4096, 0, 0).Stream(ctx, response); err != nil {
			b.Fatal(err)
		}
	}
}
