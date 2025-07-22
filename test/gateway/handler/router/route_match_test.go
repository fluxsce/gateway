package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/router"
)

func TestRouteMatching(t *testing.T) {
	tests := []struct {
		name        string
		routePath   string
		matchType   int
		requestPath string
		expected    bool
	}{
		// 精确匹配测试
		{
			name:        "Exact match - match",
			routePath:   "/api/users",
			matchType:   router.MatchTypeExact,
			requestPath: "/api/users",
			expected:    true,
		},
		{
			name:        "Exact match - no match",
			routePath:   "/api/users",
			matchType:   router.MatchTypeExact,
			requestPath: "/api/users/123",
			expected:    false,
		},
		
		// 前缀匹配测试
		{
			name:        "Prefix match - match",
			routePath:   "/api/users",
			matchType:   router.MatchTypePrefix,
			requestPath: "/api/users/123",
			expected:    true,
		},
		{
			name:        "Prefix match - no match",
			routePath:   "/api/users",
			matchType:   router.MatchTypePrefix,
			requestPath: "/api/orders",
			expected:    false,
		},
		
		// 正则匹配测试
		{
			name:        "Regex match - match",
			routePath:   `/api/users/\d+`,
			matchType:   router.MatchTypeRegex,
			requestPath: "/api/users/123",
			expected:    true,
		},
		{
			name:        "Regex match - no match",
			routePath:   `/api/users/\d+`,
			matchType:   router.MatchTypeRegex,
			requestPath: "/api/users/abc",
			expected:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建路由配置
			config := router.RouteConfig{
				ID:        "test-route",
				Name:      "Test Route",
				Path:      tt.routePath,
				MatchType: tt.matchType,
				ServiceID: "test-service",
				Enabled:   true,
			}

			// 创建路由实例
			route, err := router.NewRoute(config)
			if err != nil {
				t.Fatalf("Failed to create route: %v", err)
			}

			// 创建请求
			req, err := http.NewRequest("GET", tt.requestPath, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// 创建响应记录器
			w := httptest.NewRecorder()

			// 创建上下文
			ctx := core.NewContext(w, req)

			// 测试匹配
			matched, err := route.Match(ctx)
			if err != nil {
				t.Fatalf("Match failed: %v", err)
			}

			if matched != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, matched)
			}
		})
	}
}

func TestInvalidRegexPattern(t *testing.T) {
	// 测试无效的正则表达式
	config := router.RouteConfig{
		ID:        "test-route",
		Name:      "Test Route",
		Path:      "[invalid-regex",
		MatchType: router.MatchTypeRegex,
		ServiceID: "test-service",
		Enabled:   true,
	}

	// 创建路由实例应该失败
	_, err := router.NewRoute(config)
	if err == nil {
		t.Error("Expected error for invalid regex pattern, but got none")
	}
} 