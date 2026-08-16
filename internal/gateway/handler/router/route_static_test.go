package router

import (
	"testing"

	"gateway/internal/gateway/handler/statichost"
)

func TestRouteConfigValidateAllowsStaticHostWithoutService(t *testing.T) {
	cfg := RouteConfig{
		ID:   "r-static",
		Path: "/",
		StaticHostConfig: &statichost.StaticHostConfig{
			Enabled:       true,
			RootDirectory: "/var/www",
		},
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("静态托管路由不应要求服务: %v", err)
	}
}

func TestNewRouteCompilesStaticHostSnapshot(t *testing.T) {
	route, err := NewRoute(RouteConfig{
		ID:   "r-static-compile",
		Path: "/",
		StaticHostConfig: &statichost.StaticHostConfig{
			Enabled:       true,
			RootDirectory: t.TempDir(),
			RewriteRules: []statichost.RewriteRule{
				{Mode: statichost.RewriteModeRegex, From: `^/v/(.*)$`, To: "/$1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("合法静态配置应能创建路由: %v", err)
	}
	if !route.staticHostSnapshot.IsActive() || len(route.staticHostSnapshot.Rules) != 1 {
		t.Fatal("创建路由时应编译静态托管快照")
	}

	_, err = NewRoute(RouteConfig{
		ID:   "r-static-bad-regex",
		Path: "/",
		StaticHostConfig: &statichost.StaticHostConfig{
			Enabled:       true,
			RootDirectory: t.TempDir(),
			RewriteRules: []statichost.RewriteRule{
				{Mode: statichost.RewriteModeRegex, From: "(unclosed", To: "/$1"},
			},
		},
	})
	if err == nil {
		t.Fatal("非法正则应在创建路由时失败")
	}
}

func TestRouteConfigValidateProxyRequiresService(t *testing.T) {
	cfg := RouteConfig{
		ID:          "r-proxy",
		Path:        "/api",
		BackendType: BackendTypeProxy,
		StaticHostConfig: &statichost.StaticHostConfig{
			Enabled:       true,
			RootDirectory: "/var/www",
		},
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("服务代理即使有静态配置也应要求服务")
	}
}

func TestRouteConfigValidateStillRequiresBackend(t *testing.T) {
	cfg := RouteConfig{
		ID:   "r-empty",
		Path: "/api",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("既无服务也无静态托管时应校验失败")
	}
}
