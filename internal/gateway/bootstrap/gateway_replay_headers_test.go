package bootstrap

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
)

func TestApplyGatewayReplayHeaders(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(constants.HeaderXGatewayReplay, "Y")
	req.Header.Set(constants.HeaderXGatewayReplayTraceID, "orig-trace-123")
	rec := httptest.NewRecorder()
	ctx := core.NewContext(rec, req)

	applyGatewayReplayHeaders(req, ctx)

	if v, ok := ctx.GetString(constants.ContextKeyPresetTraceID); !ok || v != "orig-trace-123" {
		t.Fatalf("preset trace: got %q ok=%v", v, ok)
	}
	if v, ok := ctx.GetString(constants.ContextKeyIsGatewayReplay); !ok || v != "Y" {
		t.Fatalf("replay flag: got %q ok=%v", v, ok)
	}
	if req.Header.Get(constants.HeaderXGatewayReplay) != "" || req.Header.Get(constants.HeaderXGatewayReplayTraceID) != "" {
		t.Fatal("internal headers should be stripped from request")
	}
}

func TestApplyGatewayReplayHeaders_nonYDoesNotSetReplay(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api", nil)
	req.Header.Set(constants.HeaderXGatewayReplay, "1")
	rec := httptest.NewRecorder()
	ctx := core.NewContext(rec, req)
	applyGatewayReplayHeaders(req, ctx)
	if _, ok := ctx.Get(constants.ContextKeyIsGatewayReplay); ok {
		t.Fatal("expected no is_gateway_replay when header is not Y")
	}
}
