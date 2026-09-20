package logwrite

import (
	"context"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

func newBackendTraceTestContext(instanceID, traceID, tenantID string) *core.Context {
	ctx := newSubmitTestContext(traceID, tenantID)
	ctx.Set(constants.ContextKeyGatewayInstanceID, instanceID)
	return ctx
}

func waitBackendTraceCount(t *testing.T, w *captureAccessWriter, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if w.backendCount() >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("后端追踪条数 = %d, want >= %d", w.backendCount(), want)
}

func TestWriteBackendTraceLogSyncDoesNotBlockCaller(t *testing.T) {
	const instanceID = "backend-trace-async-instance"
	hold := &holdBackendWriter{
		captureAccessWriter: newCaptureAccessWriter(),
		started:             make(chan struct{}),
		unblock:             make(chan struct{}),
	}
	hold.cfg.RecordRequestBody = "Y"
	hold.cfg.RecordResponseBody = "Y"
	hold.cfg.RecordHeaders = "Y"
	if err := RegisterLogWriter(instanceID, hold); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		close(hold.unblock)
		_ = UnregisterLogWriter(instanceID)
	})

	ctx := newBackendTraceTestContext(instanceID, "TRACE-BACKEND-ASYNC", "tenant-backend")
	start := time.Now()
	returned := make(chan error, 1)
	go func() {
		returned <- WriteBackendTraceLogSync(
			"",
			ctx,
			"svc-1",
			"",
			"GET",
			"http://backend.example/api/ping?x=1",
			12,
			start,
			start.Add(3*time.Millisecond),
			200,
			map[string][]string{"X-Resp": {"ok"}},
			[]byte("resp-body"),
			map[string][]string{"X-Fwd": {"1"}},
			[]byte("req-body"),
			nil,
			"demo-svc",
			1,
			"round-robin",
			"STATIC node=n1 url=http://backend.example candidates=1",
		)
	}()
	select {
	case err := <-returned:
		if err != nil {
			t.Fatalf("WriteBackendTraceLogSync: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("快照入队不应卡住调用方")
	}
	select {
	case <-hold.started:
	case <-time.After(2 * time.Second):
		t.Fatal("worker 未开始写入")
	}
	if hold.backendCount() != 0 {
		t.Fatal("调用方返回时写入器仍应被卡住")
	}
}

func TestWriteBackendTraceLogSyncWritesExpectedFields(t *testing.T) {
	const instanceID = "backend-trace-fields-instance"
	writer := newCaptureAccessWriter()
	writer.cfg.RecordRequestBody = "Y"
	writer.cfg.RecordResponseBody = "Y"
	writer.cfg.RecordHeaders = "Y"
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = UnregisterLogWriter(instanceID)
	})

	start := time.Now().Add(-5 * time.Millisecond)
	ctx := newBackendTraceTestContext(instanceID, "TRACE-BACKEND-FIELDS", "tenant-backend")
	ctx.Set(constants.ContextKeyResponseSize, 42)
	if err := WriteBackendTraceLogSync(
		"",
		ctx,
		"svc-fields",
		"",
		"POST",
		"http://backend.example/v1/echo?q=1",
		9,
		start,
		time.Now(),
		201,
		map[string][]string{"Content-Type": {"text/plain"}},
		[]byte("hello-resp"),
		map[string][]string{"Content-Type": {"application/json"}},
		[]byte("hello-req"),
		nil,
		"echo",
		2,
		"round-robin",
		"INTERNAL node=n1 url=http://backend.example/v1/echo candidates=2",
	); err != nil {
		t.Fatalf("WriteBackendTraceLogSync: %v", err)
	}
	waitBackendTraceCount(t, writer, 1)
	if got := writer.lastBackendTraceID(); got != "TRACE-BACKEND-FIELDS" {
		t.Fatalf("traceId = %q", got)
	}
	writer.mu.Lock()
	log := writer.traces[0]
	writer.mu.Unlock()
	if log.ServiceDefinitionID != "svc-fields" || log.ServiceName != "echo" {
		t.Fatalf("service = %s/%s", log.ServiceDefinitionID, log.ServiceName)
	}
	if log.ForwardPath != "/v1/echo" || log.ForwardQuery != "q=1" {
		t.Fatalf("forward = %s?%s", log.ForwardPath, log.ForwardQuery)
	}
	if log.ForwardBody != "hello-req" || log.ResponseBody != "hello-resp" {
		t.Fatalf("body forward=%q response=%q", log.ForwardBody, log.ResponseBody)
	}
	if log.ResponseSize != 42 || log.StatusCode != 201 || log.RetryCount != 2 {
		t.Fatalf("size=%d status=%d retry=%d", log.ResponseSize, log.StatusCode, log.RetryCount)
	}
	if log.SuccessFlag != "Y" {
		t.Fatalf("successFlag = %q", log.SuccessFlag)
	}
	if log.LoadBalancerStrategy != "round-robin" {
		t.Fatalf("strategy = %q", log.LoadBalancerStrategy)
	}
	if log.LoadBalancerDecision != "INTERNAL node=n1 url=http://backend.example/v1/echo candidates=2" {
		t.Fatalf("decision = %q", log.LoadBalancerDecision)
	}
}

func TestWriteBackendTraceLogSyncSkipsMissingServiceID(t *testing.T) {
	const instanceID = "backend-trace-skip-instance"
	writer := newCaptureAccessWriter()
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = UnregisterLogWriter(instanceID)
	})

	ctx := newBackendTraceTestContext(instanceID, "TRACE-BACKEND-SKIP", "tenant-backend")
	_ = WriteBackendTraceLogSync("", ctx, "", "", "GET", "http://x", 0, time.Now(), time.Now(), 200, nil, nil, nil, nil, nil, "", 0, "", "")
	time.Sleep(50 * time.Millisecond)
	if writer.backendCount() != 0 {
		t.Fatalf("缺 serviceID 不应入队, got %d", writer.backendCount())
	}
}

func TestCloseLogWriterWaitsBackendTrace(t *testing.T) {
	const instanceID = "close-wait-backend-trace"
	writer := newCaptureAccessWriter()
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}

	ctx := newBackendTraceTestContext(instanceID, "TRACE-BACKEND-CLOSE", "tenant-backend")
	_ = WriteBackendTraceLogSync(instanceID, ctx, "svc-close", "", "GET", "http://backend.example/", 0, time.Now(), time.Now(), 200, nil, nil, nil, nil, nil, "svc", 0, "", "")
	if err := CloseLogWriter(instanceID); err != nil {
		t.Fatalf("CloseLogWriter: %v", err)
	}
	if writer.backendCount() != 1 {
		t.Fatalf("关写入器前应写完已提交后端追踪, got %d", writer.backendCount())
	}
}

type holdBackendWriter struct {
	*captureAccessWriter
	started chan struct{}
	unblock chan struct{}
	once    sync.Once
}

func (w *holdBackendWriter) WriteBackendTraceLog(ctx context.Context, log *types.BackendTraceLog) error {
	w.once.Do(func() { close(w.started) })
	<-w.unblock
	return w.captureAccessWriter.WriteBackendTraceLog(ctx, log)
}
