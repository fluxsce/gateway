package logwrite

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
)

func TestShouldSkipStaticAccessLog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		proxy     string
		result    string
		status    int
		setStatus bool
		want      bool
	}{
		{name: "proxy success", proxy: "static", result: "file", status: http.StatusOK, setStatus: true, want: true},
		{name: "spa success", proxy: "static", result: "spa", status: http.StatusOK, setStatus: true, want: true},
		{name: "not modified", proxy: "static", result: "file", status: http.StatusNotModified, setStatus: true, want: true},
		{name: "not found", proxy: "static", result: "404", status: http.StatusNotFound, setStatus: true, want: false},
		{name: "forbidden", proxy: "static", result: "forbidden", status: http.StatusForbidden, setStatus: true, want: false},
		{name: "too large", proxy: "static", result: "too_large", status: http.StatusRequestEntityTooLarge, setStatus: true, want: false},
		{name: "server error", proxy: "static", result: "error", status: http.StatusInternalServerError, setStatus: true, want: false},
		{name: "options", proxy: "static", result: "options", status: http.StatusNoContent, setStatus: true, want: true},
		{name: "dir slash redirect", proxy: "static", result: "redirect", status: http.StatusMovedPermanently, setStatus: true, want: true},
		{name: "http proxy", proxy: "http", status: http.StatusOK, setStatus: true, want: false},
		{name: "no static mark", status: http.StatusOK, setStatus: true, want: false},
		{name: "result only success", result: "index", status: http.StatusOK, setStatus: true, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := newDiagnosticsContext()
			if tt.proxy != "" {
				ctx.Set(constants.ContextKeyProxyType, tt.proxy)
			}
			if tt.result != "" {
				ctx.Set(constants.ContextKeyStaticResult, tt.result)
			}
			if tt.setStatus {
				ctx.Set(constants.GatewayStatusCode, tt.status)
			}
			if got := shouldSkipStaticAccessLog(ctx); got != tt.want {
				t.Fatalf("shouldSkipStaticAccessLog() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShouldAlertStaticAccessLog(t *testing.T) {
	t.Parallel()

	if shouldAlertStaticAccessLog(&types.AccessLog{ProxyType: "static", GatewayStatusCode: http.StatusNotFound}) {
		t.Fatal("static 404 should not alert")
	}
	if shouldAlertStaticAccessLog(&types.AccessLog{ProxyType: "static", GatewayStatusCode: http.StatusOK}) {
		t.Fatal("static 200 should not alert")
	}
	if !shouldAlertStaticAccessLog(&types.AccessLog{ProxyType: "static", GatewayStatusCode: http.StatusInternalServerError}) {
		t.Fatal("static 500 should alert")
	}
	if shouldAlertStaticAccessLog(&types.AccessLog{ProxyType: "http", GatewayStatusCode: http.StatusNotFound}) {
		t.Fatal("non-static helper is only for static logs")
	}
}

func TestHandleGatewayLogWriteSkipsStaticClientErrors(t *testing.T) {
	t.Parallel()

	cfg := &types.LogConfig{}
	cfg.SetAlertConfig(&types.AlertConfig{
		AlertEnabled:       true,
		AlertStatusCodes:   []int{404, 500},
		AlertOnTimeout:     true,
		TimeoutThresholdMs: 1,
	})

	// 无告警服务时 sendAlert 直接返回；此处确认静态 404/超时不会 panic，且 5xx 仍走检查。
	HandleGatewayLogWrite(cfg, &types.AccessLog{
		ProxyType:             "static",
		GatewayStatusCode:     http.StatusNotFound,
		TotalProcessingTimeMs: 100,
		GatewayInstanceID:     "gw1",
	})
	HandleGatewayLogWrite(cfg, &types.AccessLog{
		ProxyType:         "static",
		GatewayStatusCode: http.StatusInternalServerError,
		GatewayInstanceID: "gw1",
	})
}

func TestShouldSkipStaticAccessLogNilContext(t *testing.T) {
	t.Parallel()
	if shouldSkipStaticAccessLog(nil) {
		t.Fatal("nil context should not skip")
	}
	req := httptest.NewRequest(http.MethodGet, "http://gateway/", nil)
	if shouldSkipStaticAccessLog(core.NewContext(httptest.NewRecorder(), req)) {
		t.Fatal("plain context should not skip")
	}
}
