package bootstrap

import (
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gateway/internal/gateway/config"
	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/helper"
)

func TestRequestAdmissionLimiterEnforcesDynamicLimit(t *testing.T) {
	var limiter requestAdmissionLimiter
	limiter.setLimit(1)
	if !limiter.tryAcquire() {
		t.Fatal("first request was rejected")
	}
	if limiter.tryAcquire() {
		t.Fatal("second request exceeded the configured limit")
	}
	if limiter.activeCount() != 1 || limiter.rejectedCount() != 1 {
		t.Fatalf("active=%d rejected=%d, want active=1 rejected=1",
			limiter.activeCount(), limiter.rejectedCount())
	}

	limiter.setLimit(2)
	if !limiter.tryAcquire() {
		t.Fatal("request was rejected after increasing the limit")
	}
	limiter.release()
	limiter.release()
	if limiter.activeCount() != 0 {
		t.Fatalf("active requests = %d, want 0", limiter.activeCount())
	}
}

func TestGatewayWorkerLimitReturnsUnifiedResponse(t *testing.T) {
	cfg := config.DefaultGatewayConfig
	cfg.InstanceID = "capacity-response-test"
	cfg.Base.EnableAccessLog = false
	gateway := &Gateway{}
	gateway.requestLimiter.setLimit(1)
	if !gateway.requestLimiter.tryAcquire() {
		t.Fatal("failed to occupy request capacity")
	}
	defer gateway.requestLimiter.release()

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://gateway.test/orders", nil)
	gateway.serveHTTPWithRuntime(&cfg, core.NewEngine(), recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	if recorder.Header().Get("Retry-After") != "1" {
		t.Fatalf("Retry-After = %q, want 1", recorder.Header().Get("Retry-After"))
	}
	var response helper.GatewayResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Code != constants.ErrorCodeGatewayOverloaded ||
		response.Path != "/orders" ||
		response.TraceID == "" {
		t.Fatalf("unexpected overload response: %+v", response)
	}
	if gateway.requestLimiter.rejectedCount() != 1 {
		t.Fatalf("rejected requests = %d, want 1", gateway.requestLimiter.rejectedCount())
	}
}

func TestEnginePreservesRequestArrivalTime(t *testing.T) {
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://gateway.test/orders", nil)
	ctx := core.NewContext(recorder, request)
	defer ctx.Cancel()
	arrivalTime := ctx.GetStartTime()
	time.Sleep(5 * time.Millisecond)

	core.NewEngine().HandleWithContext(ctx, recorder, request)
	if !ctx.GetStartTime().Equal(arrivalTime) {
		t.Fatalf("request start time changed from %s to %s", arrivalTime, ctx.GetStartTime())
	}
	if ctx.Elapsed() < 5*time.Millisecond {
		t.Fatalf("elapsed time = %s, want at least 5ms", ctx.Elapsed())
	}
}

func TestListenerDispatcherConnectionLimit(t *testing.T) {
	dispatcher := newListenerDispatcher(&stubListener{addr: &net.TCPAddr{}})
	dispatcher.setMaxConnections(1)
	if !dispatcher.tryAcquireConnection() {
		t.Fatal("first connection was rejected")
	}
	if dispatcher.tryAcquireConnection() {
		t.Fatal("second connection exceeded the configured limit")
	}

	client, server := net.Pipe()
	conn := &trackedConn{Conn: server, release: dispatcher.releaseConnection}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
	_ = client.Close()
	if dispatcher.activeConnections.Load() != 0 {
		t.Fatalf("active connections = %d, want 0", dispatcher.activeConnections.Load())
	}
	if dispatcher.rejectedConnections.Load() != 1 {
		t.Fatalf("rejected connections = %d, want 1", dispatcher.rejectedConnections.Load())
	}
}

// stubListener 仅为容量限制单元测试提供监听地址。
type stubListener struct {
	addr net.Addr
}

func (l *stubListener) Accept() (net.Conn, error) {
	return nil, net.ErrClosed
}

func (l *stubListener) Close() error {
	return nil
}

func (l *stubListener) Addr() net.Addr {
	return l.addr
}
