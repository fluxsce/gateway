package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/filter"
)

func TestGlobalFilterSwitchOffSkipsFilters(t *testing.T) {
	off := false
	handler, err := NewRouterHandler(RouterConfig{
		Name:                "router",
		Enabled:             true,
		EnableGlobalFilters: &off,
		FilterConfig:        []filter.FilterConfig{stripGlobalFilter()},
	})
	if err != nil {
		t.Fatalf("NewRouterHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/aa/users", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	handler.Handle(ctx)
	if req.URL.Path != "/aa/users" {
		t.Fatalf("开关关闭时路径应保持原样，实际 %s", req.URL.Path)
	}
}

func TestGlobalFilterSwitchDefaultsOn(t *testing.T) {
	handler, err := NewRouterHandler(RouterConfig{
		Name:         "router",
		Enabled:      true,
		FilterConfig: []filter.FilterConfig{stripGlobalFilter()},
	})
	if err != nil {
		t.Fatalf("NewRouterHandler: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/aa/users", nil)
	ctx := core.NewContext(httptest.NewRecorder(), req)
	handler.Handle(ctx)
	if req.URL.Path != "/users" {
		t.Fatalf("未配置开关时应执行全局前缀剥离，实际 %s", req.URL.Path)
	}
}

func stripGlobalFilter() filter.FilterConfig {
	return filter.FilterConfig{
		ID:      "strip",
		Name:    "strip",
		Type:    string(filter.StripFilterType),
		Enabled: true,
		Order:   1,
		Action:  string(filter.PreRouting),
		Config: map[string]interface{}{
			"stripConfig": map[string]interface{}{"prefix": "/aa"},
		},
	}
}
