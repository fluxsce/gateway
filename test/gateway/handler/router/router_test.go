package router

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/router"
)

func TestRouteConfig(t *testing.T) {
	// 创建基本路由配置
	config := &router.RouteConfig{
		ID:        "test-route",
		ServiceID: "test-service",
		Path:      "/api/v1/users/**",
		Methods:   []string{"GET", "POST"},
		Enabled:   true,
		Priority:  1,
		Metadata: map[string]interface{}{
			"description": "用户API路由",
			"version":     "v1.0",
		},
	}

	// 验证配置字段
	assert.Equal(t, "test-route", config.ID)
	assert.Equal(t, "test-service", config.ServiceID)
	assert.Equal(t, "/api/v1/users/**", config.Path)
	assert.Equal(t, []string{"GET", "POST"}, config.Methods)
	assert.True(t, config.Enabled)
	assert.Equal(t, 1, config.Priority)
	assert.NotNil(t, config.Metadata)
}

func TestRouterConfig(t *testing.T) {
	// 创建路由器配置
	routerConfig := &router.RouterConfig{
		ID:      "test-router",
		Name:    "测试路由器",
		Enabled: true,
		Routes: []router.RouteConfig{
			{
				ID:        "route-1",
				ServiceID: "service-1",
				Path:      "/api/v1/users",
				Methods:   []string{"GET"},
				Enabled:   true,
				Priority:  1,
			},
			{
				ID:        "route-2",
				ServiceID: "service-2",
				Path:      "/api/v1/orders",
				Methods:   []string{"GET", "POST"},
				Enabled:   true,
				Priority:  2,
			},
		},
		EnableRouteCache: true,
		RouteCacheTTL:    60,
		DefaultPriority:  100,
	}

	// 验证路由器配置
	assert.Equal(t, "test-router", routerConfig.ID)
	assert.Equal(t, "测试路由器", routerConfig.Name)
	assert.True(t, routerConfig.Enabled)
	assert.Len(t, routerConfig.Routes, 2)
	assert.True(t, routerConfig.EnableRouteCache)
	assert.Equal(t, 60, routerConfig.RouteCacheTTL)
	assert.Equal(t, 100, routerConfig.DefaultPriority)
}

func TestDefaultRouterConfig(t *testing.T) {
	defaultConfig := router.DefaultRouterConfig

	// 验证默认配置
	assert.Equal(t, "default-router", defaultConfig.ID)
	assert.Equal(t, "default-router", defaultConfig.Name)
	assert.True(t, defaultConfig.Enabled)
	assert.Empty(t, defaultConfig.Routes, "默认路由列表应该为空")
	assert.False(t, defaultConfig.EnableRouteCache, "默认应该禁用路由缓存")
	assert.Equal(t, 300, defaultConfig.RouteCacheTTL, "默认缓存TTL应该是300秒")
	assert.Equal(t, 100, defaultConfig.DefaultPriority, "默认优先级应该是100")
}

func TestRouterHandler(t *testing.T) {
	// 创建路由配置
	routes := []router.RouteConfig{
		{
			ID:        "users-route",
			ServiceID: "user-service",
			Path:      "/api/v1/users/**",
			Methods:   []string{"GET", "POST", "PUT", "DELETE"},
			Enabled:   true,
			Priority:  1,
		},
		{
			ID:        "orders-route",
			ServiceID: "order-service",
			Path:      "/api/v1/orders/**",
			Methods:   []string{"GET", "POST"},
			Enabled:   true,
			Priority:  2,
		},
		{
			ID:        "admin-route",
			ServiceID: "admin-service",
			Path:      "/api/admin/**",
			Methods:   []string{"GET", "POST", "PUT", "DELETE"},
			Enabled:   true,
			Priority:  3,
		},
	}

	routerConfig := &router.RouterConfig{
		ID:               "test-router",
		Name:             "测试路由器",
		Enabled:          true,
		Routes:           routes,
		EnableRouteCache: false, // 测试时禁用缓存
		DefaultPriority:  100,
	}

	// 创建路由器处理器
	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(t, err, "创建路由器处理器失败")
	require.NotNil(t, handler, "路由器处理器不应该为nil")

	// 测试路由匹配
	tests := []struct {
		name          string
		method        string
		path          string
		expectedRoute string
		expectedMatch bool
		description   string
	}{
		{
			name:          "MatchUsersGET",
			method:        "GET",
			path:          "/api/v1/users/123",
			expectedRoute: "users-route",
			expectedMatch: true,
			description:   "GET /api/v1/users/123 应该匹配用户路由",
		},
		{
			name:          "MatchUsersPOST",
			method:        "POST",
			path:          "/api/v1/users",
			expectedRoute: "users-route",
			expectedMatch: true,
			description:   "POST /api/v1/users 应该匹配用户路由",
		},
		{
			name:          "MatchOrdersGET",
			method:        "GET",
			path:          "/api/v1/orders/456",
			expectedRoute: "orders-route",
			expectedMatch: true,
			description:   "GET /api/v1/orders/456 应该匹配订单路由",
		},
		{
			name:          "MatchAdminPUT",
			method:        "PUT",
			path:          "/api/admin/users/123",
			expectedRoute: "admin-route",
			expectedMatch: true,
			description:   "PUT /api/admin/users/123 应该匹配管理路由",
		},
		{
			name:          "NoMatchInvalidMethod",
			method:        "PATCH",
			path:          "/api/v1/orders/456",
			expectedRoute: "",
			expectedMatch: false,
			description:   "PATCH 方法不应该匹配订单路由",
		},
		{
			name:          "NoMatchInvalidPath",
			method:        "GET",
			path:          "/api/v2/users/123",
			expectedRoute: "",
			expectedMatch: false,
			description:   "不存在的路径不应该匹配任何路由",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest(tt.method, tt.path, nil)
			writer := httptest.NewRecorder()

			// 创建测试上下文
			ctx := core.NewContext(writer, req)

			// 执行路由匹配
			result := handler.Handle(ctx)

			if tt.expectedMatch {
				assert.True(t, result, tt.description)

				// 验证上下文中设置的路由信息
				routeID := ctx.GetRouteID()
				assert.Equal(t, tt.expectedRoute, routeID, "路由ID应该匹配")

				// 验证服务ID也被设置
				serviceID := ctx.GetServiceID()
				assert.NotEmpty(t, serviceID, "服务ID应该被设置")
			} else {
				assert.False(t, result, tt.description)
			}
		})
	}
}

func TestRouteMatching1(t *testing.T) {
	tests := []struct {
		name        string
		routePath   string
		requestPath string
		shouldMatch bool
		description string
	}{
		{
			name:        "ExactMatch",
			routePath:   "/api/users",
			requestPath: "/api/users",
			shouldMatch: true,
			description: "精确匹配",
		},
		{
			name:        "WildcardMatch",
			routePath:   "/api/users/**",
			requestPath: "/api/users/123",
			shouldMatch: true,
			description: "通配符匹配",
		},
		{
			name:        "WildcardDeepMatch",
			routePath:   "/api/users/**",
			requestPath: "/api/users/123/profile",
			shouldMatch: true,
			description: "深层通配符匹配",
		},
		{
			name:        "PrefixNoMatch",
			routePath:   "/api/users",
			requestPath: "/api/users/123",
			shouldMatch: false,
			description: "非通配符路径不应该匹配子路径",
		},
		{
			name:        "DifferentPath",
			routePath:   "/api/users",
			requestPath: "/api/orders",
			shouldMatch: false,
			description: "不同路径不应该匹配",
		},
		{
			name:        "CaseSensitive",
			routePath:   "/api/users",
			requestPath: "/API/USERS",
			shouldMatch: false,
			description: "路径匹配应该区分大小写",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建简单的路由配置
			routerConfig := &router.RouterConfig{
				ID:      "test-router",
				Name:    "路径匹配测试",
				Enabled: true,
				Routes: []router.RouteConfig{
					{
						ID:        "test-route",
						ServiceID: "test-service",
						Path:      tt.routePath,
						Methods:   []string{"GET"},
						Enabled:   true,
						Priority:  1,
					},
				},
				EnableRouteCache: false,
				DefaultPriority:  100,
			}

			// 创建路由器处理器
			handler, err := router.NewRouterHandler(*routerConfig)
			require.NoError(t, err)

			// 创建测试请求
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			// 执行路由匹配
			result := handler.Handle(ctx)

			// 验证匹配结果
			assert.Equal(t, tt.shouldMatch, result, tt.description)
		})
	}
}

func TestMethodMatching(t *testing.T) {
	// 创建支持多种方法的路由
	routerConfig := &router.RouterConfig{
		ID:      "method-test-router",
		Name:    "方法匹配测试",
		Enabled: true,
		Routes: []router.RouteConfig{
			{
				ID:        "multi-method-route",
				ServiceID: "test-service",
				Path:      "/api/test",
				Methods:   []string{"GET", "POST", "PUT"},
				Enabled:   true,
				Priority:  1,
			},
		},
		EnableRouteCache: false,
		DefaultPriority:  100,
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(t, err)

	// 测试允许的方法
	allowedMethods := []string{"GET", "POST", "PUT"}
	for _, method := range allowedMethods {
		t.Run("Allow"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/test", nil)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.True(t, result, method+" 方法应该被允许")
		})
	}

	// 测试不允许的方法
	disallowedMethods := []string{"DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range disallowedMethods {
		t.Run("Disallow"+method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/api/test", nil)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.False(t, result, method+" 方法不应该被允许")
		})
	}
}

func TestRoutePriority(t *testing.T) {
	// 创建具有不同优先级的路由
	routerConfig := &router.RouterConfig{
		ID:      "priority-test-router",
		Name:    "优先级测试",
		Enabled: true,
		Routes: []router.RouteConfig{
			{
				ID:        "low-priority-route",
				ServiceID: "service-low",
				Path:      "/api/**",
				Methods:   []string{"GET"},
				Enabled:   true,
				Priority:  10, // 低优先级
			},
			{
				ID:        "high-priority-route",
				ServiceID: "service-high",
				Path:      "/api/users/**",
				Methods:   []string{"GET"},
				Enabled:   true,
				Priority:  1, // 高优先级
			},
		},
		EnableRouteCache: false,
		DefaultPriority:  100,
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(t, err)

	// 测试高优先级路由应该被优先匹配
	req := httptest.NewRequest("GET", "/api/users/123", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result := handler.Handle(ctx)
	assert.True(t, result, "应该匹配到路由")

	// 验证匹配的是高优先级路由
	routeID := ctx.GetRouteID()
	assert.Equal(t, "high-priority-route", routeID, "应该匹配高优先级路由")

	serviceID := ctx.GetServiceID()
	assert.Equal(t, "service-high", serviceID, "应该设置高优先级路由的服务ID")
}

func TestDisabledRoute(t *testing.T) {
	// 创建包含禁用路由的配置
	routerConfig := &router.RouterConfig{
		ID:      "disabled-test-router",
		Name:    "禁用路由测试",
		Enabled: true,
		Routes: []router.RouteConfig{
			{
				ID:        "disabled-route",
				ServiceID: "disabled-service",
				Path:      "/api/disabled",
				Methods:   []string{"GET"},
				Enabled:   false, // 禁用的路由
				Priority:  1,
			},
			{
				ID:        "enabled-route",
				ServiceID: "enabled-service",
				Path:      "/api/enabled",
				Methods:   []string{"GET"},
				Enabled:   true, // 启用的路由
				Priority:  1,
			},
		},
		EnableRouteCache: false,
		DefaultPriority:  100,
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(t, err)

	// 测试禁用的路由不应该匹配
	req := httptest.NewRequest("GET", "/api/disabled", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result := handler.Handle(ctx)
	assert.False(t, result, "禁用的路由不应该匹配")

	// 测试启用的路由应该匹配
	req = httptest.NewRequest("GET", "/api/enabled", nil)
	writer = httptest.NewRecorder()
	ctx = core.NewContext(writer, req)

	result = handler.Handle(ctx)
	assert.True(t, result, "启用的路由应该匹配")
}

func TestRouterInterface(t *testing.T) {
	routerConfig := &router.RouterConfig{
		ID:      "interface-test",
		Name:    "接口测试",
		Enabled: true,
		Routes:  []router.RouteConfig{},
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(t, err)

	// 验证处理器实现了相应接口
	assert.NotNil(t, handler.Handle, "Handle方法不应该为nil")
}

// 基准测试
func BenchmarkRouterMatching(b *testing.B) {
	// 创建包含多个路由的配置
	routes := make([]router.RouteConfig, 100)
	for i := 0; i < 100; i++ {
		routes[i] = router.RouteConfig{
			ID:        fmt.Sprintf("route-%d", i),
			ServiceID: fmt.Sprintf("service-%d", i),
			Path:      fmt.Sprintf("/api/v1/resource%d/**", i),
			Methods:   []string{"GET", "POST"},
			Enabled:   true,
			Priority:  i + 1,
		}
	}

	routerConfig := &router.RouterConfig{
		ID:               "bench-router",
		Name:             "基准测试路由器",
		Enabled:          true,
		Routes:           routes,
		EnableRouteCache: false, // 禁用缓存测试原始性能
		DefaultPriority:  1000,
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(b, err)

	req := httptest.NewRequest("GET", "/api/v1/resource50/item/123", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx)
	}
}

func BenchmarkRouterMatchingWithCache(b *testing.B) {
	// 创建启用缓存的路由配置
	routes := make([]router.RouteConfig, 100)
	for i := 0; i < 100; i++ {
		routes[i] = router.RouteConfig{
			ID:        fmt.Sprintf("cached-route-%d", i),
			ServiceID: fmt.Sprintf("service-%d", i),
			Path:      fmt.Sprintf("/api/v1/cached%d/**", i),
			Methods:   []string{"GET", "POST"},
			Enabled:   true,
			Priority:  i + 1,
		}
	}

	routerConfig := &router.RouterConfig{
		ID:               "bench-cached-router",
		Name:             "缓存基准测试路由器",
		Enabled:          true,
		Routes:           routes,
		EnableRouteCache: true, // 启用缓存
		RouteCacheTTL:    300,
		DefaultPriority:  1000,
	}

	handler, err := router.NewRouterHandler(*routerConfig)
	require.NoError(b, err)

	req := httptest.NewRequest("GET", "/api/v1/cached50/item/123", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.Handle(ctx)
	}
}
