package router

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/internal/gateway/core"
)

func TestNormalizeMock(t *testing.T) {
	status, contentType, body, err := NormalizeMock(0, "", `{"ok":true}`)
	if err != nil || status != http.StatusOK || contentType != defaultMockContentType || body != `{"ok":true}` {
		t.Fatalf("缺省应规范为 200 JSON，实际 status=%d type=%q body=%q err=%v", status, contentType, body, err)
	}
	if status, _, _, err = NormalizeMock(201, "text/plain; charset=utf-8", "created"); err != nil || status != http.StatusCreated {
		t.Fatalf("201 与自定义类型应保留，实际 status=%d err=%v", status, err)
	}
	if _, _, _, err = NormalizeMock(99, "", ""); err == nil {
		t.Fatal("小于 100 的状态码应拒绝")
	}
	if _, _, _, err = NormalizeMock(600, "", ""); err == nil {
		t.Fatal("大于 599 的状态码应拒绝")
	}
	if _, _, _, err = NormalizeMock(200, "json", "{}"); err == nil {
		t.Fatal("缺少斜杠的 Content-Type 应拒绝")
	}
	if _, _, _, err = NormalizeMock(200, "application/json\r\nX-Test: 1", "{}"); err == nil {
		t.Fatal("含换行的 Content-Type 应拒绝")
	}
	if _, _, _, err = NormalizeMock(http.StatusNoContent, "", "nope"); err == nil {
		t.Fatal("204 带正文应拒绝")
	}
	if _, _, body, err = NormalizeMock(http.StatusNoContent, "", ""); err != nil || body != "" {
		t.Fatalf("204 空正文应合法，body=%q err=%v", body, err)
	}
	if _, _, _, err = NormalizeMock(200, "", strings.Repeat("a", maxMockBodyBytes+1)); err == nil {
		t.Fatal("超长正文应拒绝")
	}
}

func TestRouteWriteMock(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:              "r-mock",
		Name:            "health-mock",
		Path:            "/api/mock/health",
		MatchType:       MatchTypePrefix,
		Enabled:         true,
		BackendType:     BackendTypeMock,
		MockStatus:      http.StatusOK,
		MockContentType: "application/json; charset=utf-8",
		MockBody:        `{"status":"up"}`,
	})
	if err != nil {
		t.Fatalf("Mock 路由应能创建: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/mock/health", nil)
	ctx := core.NewContext(rec, req)
	if route.Handle(ctx) {
		t.Fatal("Mock 应终止链路")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("状态码 %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type %s", rec.Header().Get("Content-Type"))
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("Cache-Control %s", rec.Header().Get("Cache-Control"))
	}
	payload, err := io.ReadAll(rec.Body)
	if err != nil {
		t.Fatalf("读取正文失败: %v", err)
	}
	if string(payload) != `{"status":"up"}` {
		t.Fatalf("正文 %s", payload)
	}
}

func TestRouteWriteMockHeadOmitsBody(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:          "r-mock-head",
		Path:        "/api/mock/health",
		MatchType:   MatchTypeExact,
		Enabled:     true,
		BackendType: BackendTypeMock,
		MockStatus:  http.StatusOK,
		MockBody:    `{"status":"up"}`,
	})
	if err != nil {
		t.Fatalf("Mock 路由应能创建: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/api/mock/health", nil)
	ctx := core.NewContext(rec, req)
	if route.Handle(ctx) {
		t.Fatal("Mock 应终止链路")
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("HEAD 不应写出正文，长度 %d", rec.Body.Len())
	}
	if rec.Header().Get("Content-Length") != "15" {
		t.Fatalf("Content-Length %s", rec.Header().Get("Content-Length"))
	}
}
