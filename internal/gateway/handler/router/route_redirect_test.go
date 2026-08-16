package router

import (
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gateway/internal/gateway/core"
)

func TestNormalizeRedirect(t *testing.T) {
	status, loc, err := NormalizeRedirect(0, "/#/datahublogin")
	if err != nil || status != http.StatusMovedPermanently || loc != "/#/datahublogin" {
		t.Fatalf("相对路径应规范为 301，实际 status=%d loc=%q err=%v", status, loc, err)
	}
	if _, _, err = NormalizeRedirect(302, "https://example.com/app"); err != nil {
		t.Fatalf("https 目标应合法: %v", err)
	}
	if _, _, err = NormalizeRedirect(301, "https://www.example.com/#/datahublogin"); err != nil {
		t.Fatalf("带片段的绝对地址应合法: %v", err)
	}
	if _, _, err = NormalizeRedirect(302, "HTTP://192.168.1.1:8080/#/datahublogin"); err != nil {
		t.Fatalf("带端口的绝对地址应合法: %v", err)
	}
	if _, _, err = NormalizeRedirect(302, "{scheme}://{host}/#/datahublogin"); err != nil {
		t.Fatalf("绝对地址占位应合法: %v", err)
	}
	if status, _, err = NormalizeRedirect(307, "/next"); err != nil || status != http.StatusTemporaryRedirect {
		t.Fatalf("307 应保留，实际 status=%d err=%v", status, err)
	}
	if status, _, err = NormalizeRedirect(308, "/next"); err != nil || status != http.StatusPermanentRedirect {
		t.Fatalf("308 应保留，实际 status=%d err=%v", status, err)
	}
	if status, _, err = NormalizeRedirect(303, "/next"); err != nil || status != http.StatusMovedPermanently {
		t.Fatalf("不支持的状态码应回落 301，实际 status=%d err=%v", status, err)
	}
	if _, _, err = NormalizeRedirect(301, "https://user:pass@example.com/app"); err == nil {
		t.Fatal("含用户信息的绝对地址应拒绝")
	}
	if _, _, err = NormalizeRedirect(301, "//evil.example"); err == nil {
		t.Fatal("协议相对地址应拒绝")
	}
	if _, _, err = NormalizeRedirect(301, "javascript:alert(1)"); err == nil {
		t.Fatal("非 http(s) 协议应拒绝")
	}
	if _, _, err = NormalizeRedirect(301, "ok\r\nLocation: https://evil"); err == nil {
		t.Fatal("含换行的 Location 应拒绝")
	}
}

func TestRouteWriteRedirect(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:               "r-redirect",
		Name:             "logincenter-datahublogin",
		Path:             "/logincenter/datahublogin",
		MatchType:        MatchTypePrefix,
		Enabled:          true,
		BackendType:      BackendTypeRedirect,
		RedirectStatus:   http.StatusFound,
		RedirectLocation: "/#/datahublogin",
	})
	if err != nil {
		t.Fatalf("重定向路由应能创建: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/logincenter/datahublogin", nil)
	ctx := core.NewContext(rec, req)
	if route.Handle(ctx) {
		t.Fatal("重定向应终止链路")
	}
	if rec.Code != http.StatusFound {
		t.Fatalf("状态码 %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/#/datahublogin" {
		t.Fatalf("Location %s", rec.Header().Get("Location"))
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("302 应 no-store，实际 %s", rec.Header().Get("Cache-Control"))
	}
	if !strings.Contains(rec.Header().Get("Content-Type"), "text/html") {
		t.Fatalf("GET 3xx 应为 HTML，实际 %s", rec.Header().Get("Content-Type"))
	}
	body, _ := io.ReadAll(rec.Body)
	if !strings.Contains(string(body), `href="/#/datahublogin"`) {
		t.Fatalf("GET 3xx 应带链接正文: %s", body)
	}
}

func TestRouteWriteRedirectPermanentOmitsNoStore(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:               "r-redirect-301",
		Path:             "/old",
		MatchType:        MatchTypePrefix,
		Enabled:          true,
		BackendType:      BackendTypeRedirect,
		RedirectStatus:   http.StatusMovedPermanently,
		RedirectLocation: "/#/datahublogin",
	})
	if err != nil {
		t.Fatalf("重定向路由应能创建: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	ctx := core.NewContext(rec, req)
	route.Handle(ctx)
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("状态码 %d", rec.Code)
	}
	if rec.Header().Get("Cache-Control") != "" {
		t.Fatalf("301 不应强制禁缓存，实际 %s", rec.Header().Get("Cache-Control"))
	}
}

func TestRouteWriteRedirectHeadHasNoBody(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:               "r-redirect-head",
		Path:             "/old",
		MatchType:        MatchTypePrefix,
		Enabled:          true,
		BackendType:      BackendTypeRedirect,
		RedirectStatus:   http.StatusTemporaryRedirect,
		RedirectLocation: "/next",
	})
	if err != nil {
		t.Fatalf("重定向路由应能创建: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/old", nil)
	ctx := core.NewContext(rec, req)
	route.Handle(ctx)
	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("状态码 %d", rec.Code)
	}
	if rec.Header().Get("Location") != "/next" {
		t.Fatalf("Location %s", rec.Header().Get("Location"))
	}
	body, _ := io.ReadAll(rec.Body)
	if len(body) != 0 {
		t.Fatalf("HEAD 不应带响应体: %s", body)
	}
}

func TestExpandRedirectLocationUsesForwardedProto(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://gw.local/old", nil)
	req.Host = "gw.local:8080"
	req.Header.Set("X-Forwarded-Proto", "https")
	got := expandRedirectLocation("{scheme}://{host}/#/datahublogin", req)
	if got != "https://gw.local:8080/#/datahublogin" {
		t.Fatalf("展开结果 %s", got)
	}

	tlsReq := httptest.NewRequest(http.MethodGet, "https://gw.local/old", nil)
	tlsReq.TLS = &tls.ConnectionState{}
	tlsReq.Host = "gw.local"
	got = expandRedirectLocation("{scheme}://{host}/app", tlsReq)
	if got != "https://gw.local/app" {
		t.Fatalf("TLS 展开结果 %s", got)
	}
}

func TestExpandRedirectLocationRejectsSpoofedForwarded(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://gw.local/old", nil)
	req.Host = "gw.local"
	req.Header.Set("X-Forwarded-Proto", "https, http")
	got := expandRedirectLocation("{scheme}://{host}/app", req)
	if got != "http://gw.local/app" {
		t.Fatalf("多值 X-Forwarded-Proto 应忽略，实际 %s", got)
	}

	req.Host = "evil.example/steal"
	got = expandRedirectLocation("{scheme}://{host}/app", req)
	if got != "http://{host}/app" {
		t.Fatalf("非法 Host 不应展开，实际 %s", got)
	}
}

func TestRouteWriteRedirectInvalidHost(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:               "r-redirect-host",
		Path:             "/old",
		MatchType:        MatchTypePrefix,
		Enabled:          true,
		BackendType:      BackendTypeRedirect,
		RedirectStatus:   http.StatusFound,
		RedirectLocation: "{scheme}://{host}/#/datahublogin",
	})
	if err != nil {
		t.Fatalf("重定向路由应能创建: %v", err)
	}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/old", nil)
	req.Host = "evil.example/steal"
	ctx := core.NewContext(rec, req)
	if route.Handle(ctx) {
		t.Fatal("非法 Host 应终止链路")
	}
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("非法 Host 应为 500，实际 %d", rec.Code)
	}
}

func TestRouteConfigValidateRedirect(t *testing.T) {
	cfg := RouteConfig{
		ID:               "r-redirect",
		Path:             "/logincenter/datahublogin",
		BackendType:      BackendTypeRedirect,
		RedirectLocation: "/#/datahublogin",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("合法重定向不应要求服务: %v", err)
	}
	cfg.RedirectLocation = ""
	if err := cfg.Validate(); err == nil {
		t.Fatal("重定向缺少目标应失败")
	}
}
