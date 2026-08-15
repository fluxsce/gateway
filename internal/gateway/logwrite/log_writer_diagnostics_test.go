package logwrite

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

func TestAppendStreamingDiagnosticsWritesNoteText(t *testing.T) {
	ctx := newDiagnosticsContext()
	ctx.Set(constants.ContextKeySSEDisconnectType, "completed")
	ctx.Set(constants.ContextKeySSEBytesStreamed, int64(313))

	accessLog := types.NewAccessLog("t1", "gw1", "127.0.0.1")
	appendStreamingDiagnostics(accessLog, ctx)

	if accessLog.ErrorMessage != "" || accessLog.ErrorCode != "" {
		t.Fatalf("正常SSE诊断不应写入错误字段: code=%q message=%q", accessLog.ErrorCode, accessLog.ErrorMessage)
	}
	want := "sse_disconnect=completed; sse_bytes=313"
	if accessLog.NoteText != want {
		t.Fatalf("noteText = %q, want %q", accessLog.NoteText, want)
	}
}

func TestAppendStreamingDiagnosticsKeepsExistingErrorMessage(t *testing.T) {
	ctx := newDiagnosticsContext()
	ctx.Set(constants.ContextKeySSEDisconnectType, "upstream_error")
	ctx.Set(constants.ContextKeySSEBytesStreamed, int64(0))

	accessLog := types.NewAccessLog("t1", "gw1", "127.0.0.1")
	accessLog.SetErrorInfo("GATEWAY_ERROR", "Errors: [读取SSE上游响应失败]")
	appendStreamingDiagnostics(accessLog, ctx)

	if accessLog.ErrorMessage != "Errors: [读取SSE上游响应失败]" {
		t.Fatalf("真实错误信息被覆盖: %q", accessLog.ErrorMessage)
	}
	if !strings.Contains(accessLog.NoteText, "sse_disconnect=upstream_error") {
		t.Fatalf("诊断未写入 noteText: %q", accessLog.NoteText)
	}
}

func TestAppendStreamingDiagnosticsIgnoresEmptyContext(t *testing.T) {
	accessLog := types.NewAccessLog("t1", "gw1", "127.0.0.1")
	appendStreamingDiagnostics(accessLog, newDiagnosticsContext())
	if accessLog.NoteText != "" || accessLog.ErrorMessage != "" {
		t.Fatalf("无流式诊断时不应写入备注或错误: note=%q err=%q", accessLog.NoteText, accessLog.ErrorMessage)
	}
}

func newDiagnosticsContext() *core.Context {
	req := httptest.NewRequest(http.MethodGet, "http://gateway/events", nil)
	return core.NewContext(httptest.NewRecorder(), req)
}
