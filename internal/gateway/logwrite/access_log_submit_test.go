package logwrite

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

// captureAccessWriter 记录 Write 调用，供提交队列测试使用。
type captureAccessWriter struct {
	cfg  *types.LogConfig
	mu   sync.Mutex
	logs []*types.AccessLog
}

func newCaptureAccessWriter() *captureAccessWriter {
	cfg := &types.LogConfig{}
	cfg.SetDefaults()
	return &captureAccessWriter{cfg: cfg}
}

func (w *captureAccessWriter) Write(_ context.Context, log *types.AccessLog) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.logs = append(w.logs, log)
	return nil
}

func (w *captureAccessWriter) UpdateAccessLog(context.Context, *types.AccessLog) (int64, error) {
	return 0, nil
}

func (w *captureAccessWriter) BatchWrite(context.Context, []*types.AccessLog) error {
	return nil
}

func (w *captureAccessWriter) WriteBackendTraceLog(context.Context, *types.BackendTraceLog) error {
	return nil
}

func (w *captureAccessWriter) BatchWriteBackendTraceLog(context.Context, []*types.BackendTraceLog) error {
	return nil
}

func (w *captureAccessWriter) Flush(context.Context) error { return nil }

func (w *captureAccessWriter) Close() error { return nil }

func (w *captureAccessWriter) GetLogConfig() *types.LogConfig { return w.cfg }

func (w *captureAccessWriter) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.logs)
}

func (w *captureAccessWriter) lastTraceID() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.logs) == 0 {
		return ""
	}
	return w.logs[len(w.logs)-1].TraceID
}

func newSubmitTestContext(traceID, tenantID string) *core.Context {
	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	ctx.Set(constants.ContextKeyTraceID, traceID)
	ctx.Set(constants.ContextKeyTenantID, tenantID)
	ctx.Set(constants.ContextKeySnapshotRequestProto, "HTTP/1.1")
	ctx.Request = nil
	ctx.Writer = nil
	return ctx
}

func waitAccessLogCount(t *testing.T, w *captureAccessWriter, want int) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if w.count() >= want {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("访问日志条数 = %d, want >= %d", w.count(), want)
}

func TestSubmitWriteLogUsesWriteLog(t *testing.T) {
	const instanceID = "submit-write-log-instance"
	writer := newCaptureAccessWriter()
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = UnregisterLogWriter(instanceID)
	})

	ctx := newSubmitTestContext("TRACE-SUBMIT-1", "tenant-submit")
	SubmitWriteLog(instanceID, ctx)
	waitAccessLogCount(t, writer, 1)
	if got := writer.lastTraceID(); got != "TRACE-SUBMIT-1" {
		t.Fatalf("traceId = %q, want TRACE-SUBMIT-1", got)
	}
}

func TestSubmitWriteLogSkipsStaticSuccess(t *testing.T) {
	const instanceID = "submit-skip-static-instance"
	writer := newCaptureAccessWriter()
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = UnregisterLogWriter(instanceID)
	})

	ctx := newSubmitTestContext("TRACE-STATIC-SKIP", "tenant-submit")
	ctx.Set(constants.ContextKeyProxyType, staticProxyType)
	ctx.Set(constants.GatewayStatusCode, http.StatusOK)
	SubmitWriteLog(instanceID, ctx)
	time.Sleep(50 * time.Millisecond)
	if writer.count() != 0 {
		t.Fatalf("静态成功命中不应入队写主表, got %d", writer.count())
	}
}

func TestSubmitWriteLogEmptyArgsNoPanic(t *testing.T) {
	SubmitWriteLog("", nil)
	SubmitWriteLog("x", nil)
}

func TestAccessLogSubmitWorkerCount(t *testing.T) {
	n := accessLogSubmitWorkerCount()
	if n < minAccessLogWorkers || n > maxAccessLogWorkers {
		t.Fatalf("workers = %d, want in [%d, %d]", n, minAccessLogWorkers, maxAccessLogWorkers)
	}
}

func TestGetAccessLogSubmitStats(t *testing.T) {
	stats := GetAccessLogSubmitStats()
	if stats.QueueCap != accessLogSubmitQueueSize {
		t.Fatalf("QueueCap = %d, want %d", stats.QueueCap, accessLogSubmitQueueSize)
	}
	if stats.Workers < minAccessLogWorkers || stats.Workers > maxAccessLogWorkers {
		t.Fatalf("Workers = %d, want in [%d, %d]", stats.Workers, minAccessLogWorkers, maxAccessLogWorkers)
	}
	if stats.EnqueueWait == "" {
		t.Fatal("EnqueueWait 不应为空")
	}
}

func TestCloseLogWriterWaitsSubmittedLogs(t *testing.T) {
	const instanceID = "close-wait-submit"
	writer := newCaptureAccessWriter()
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}

	ctx := newSubmitTestContext("TRACE-CLOSE-WAIT", "tenant-submit")
	SubmitWriteLog(instanceID, ctx)
	if err := CloseLogWriter(instanceID); err != nil {
		t.Fatalf("CloseLogWriter: %v", err)
	}
	if writer.count() != 1 {
		t.Fatalf("关写入器前应写完已提交日志, got %d", writer.count())
	}
}

func TestUpdateLogWriterWaitsInflightWrite(t *testing.T) {
	const instanceID = "update-inflight-write"
	hold := &holdAccessWriter{
		captureAccessWriter: newCaptureAccessWriter(),
		started:             make(chan struct{}),
		unblock:             make(chan struct{}),
	}
	if err := RegisterLogWriter(instanceID, hold); err != nil {
		t.Fatalf("RegisterLogWriter: %v", err)
	}
	t.Cleanup(func() {
		_ = CloseLogWriter(instanceID)
	})

	done := make(chan error, 1)
	go func() {
		done <- WriteLog(instanceID, newSubmitTestContext("TRACE-INFLIGHT", "tenant-submit"))
	}()
	select {
	case <-hold.started:
	case <-time.After(2 * time.Second):
		t.Fatal("WriteLog 未进入旧写入器")
	}

	cfg := &types.LogConfig{
		OutputTargets:      string(types.LogOutputConsole),
		LogFormat:          string(types.LogFormatJSON),
		EnableAsyncLogging: "Y",
	}
	cfg.SetDefaults()
	cfg.OutputTargets = string(types.LogOutputConsole)
	if err := UpdateLogWriter(instanceID, cfg); err != nil {
		t.Fatalf("UpdateLogWriter: %v", err)
	}

	select {
	case err := <-done:
		t.Fatalf("热更新不应在旧 Write 完成前关掉写入器: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(hold.unblock)
	if err := <-done; err != nil {
		t.Fatalf("在途 WriteLog: %v", err)
	}
	if hold.count() != 1 {
		t.Fatalf("在途日志应写入旧 writer, got %d", hold.count())
	}
}

// holdAccessWriter 用于验证热更新会等在途 Write 结束再 Close。
type holdAccessWriter struct {
	*captureAccessWriter
	started chan struct{}
	unblock chan struct{}
	once    sync.Once
}

func (w *holdAccessWriter) Write(ctx context.Context, log *types.AccessLog) error {
	w.once.Do(func() { close(w.started) })
	<-w.unblock
	return w.captureAccessWriter.Write(ctx, log)
}
