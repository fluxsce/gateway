package proxy

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"gateway/internal/gateway/core"
)

func TestSimplePathHandling(t *testing.T) {
	proxy := &HTTPProxy{}

	tests := []struct {
		name        string
		targetPath  string
		requestPath string
		expected    string
		description string
	}{
		{
			name:        "目标路径为空",
			targetPath:  "",
			requestPath: "/users/123",
			expected:    "/users/123",
			description: "目标路径为空时使用请求地址",
		},
		{
			name:        "目标路径只有斜杠",
			targetPath:  "/",
			requestPath: "/users/123",
			expected:    "/users/123",
			description: "目标路径只有斜杠时使用请求地址",
		},
		{
			name:        "路径完全相同",
			targetPath:  "/users/123",
			requestPath: "/users/123",
			expected:    "/users/123",
			description: "当目标路径和请求路径完全相同时，直接返回目标路径",
		},
		{
			name:        "路径完全相同-带斜杠",
			targetPath:  "/users/123/",
			requestPath: "/users/123",
			expected:    "/users/123/",
			description: "当目标路径有斜杠且与请求路径相同时，保留斜杠",
		},
		{
			name:        "前缀不一样-直接使用目标地址",
			targetPath:  "/api",
			requestPath: "/users/123",
			expected:    "/api",
			description: "前缀不一样时直接使用目标地址",
		},
		{
			name:        "前缀一样-有斜杠避免重复",
			targetPath:  "/api/v1/",
			requestPath: "/api/v1/users/123",
			expected:    "/api/v1/users/123",
			description: "前缀一样且目标路径有斜杠时避免重复",
		},
		{
			name:        "用户报告的问题-前缀一样有斜杠",
			targetPath:  "/api/v1/",
			requestPath: "/api/v1/users/123",
			expected:    "/api/v1/users/123",
			description: "用户报告：期望 /api/v1/users/123, 实际 /api/v1/api/v1/use",
		},
		{
			name:        "nginx示例-前缀不一样",
			targetPath:  "/openPlateform/expressApi/expressWeixinReceiveTicket",
			requestPath: "/weixin/component_verify_ticket/test",
			expected:    "/openPlateform/expressApi/expressWeixinReceiveTicket",
			description: "nginx风格：前缀不一样时直接使用目标地址",
		},
		{
			name:        "前缀一样-请求路径包含目标路径",
			targetPath:  "/api/v1",
			requestPath: "/api/v1/users/123",
			expected:    "/api/v1/users/123",
			description: "当请求路径以目标路径为前缀时，直接返回请求路径避免重复",
		},

		{
			name:        "请求根路径-前缀不一样",
			targetPath:  "/api/",
			requestPath: "/",
			expected:    "/api/",
			description: "请求根路径且前缀不一样时返回目标路径",
		},
		{
			name:        "前缀不一样-目标有斜杠",
			targetPath:  "/backend/",
			requestPath: "/users/123",
			expected:    "/backend/",
			description: "前缀不一样时直接使用目标地址，即使有斜杠",
		},
		{
			name:        "路径边界检查",
			targetPath:  "/ap",
			requestPath: "/api/v1/users",
			expected:    "/ap",
			description: "确保/ap不会错误匹配/api/v1/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试上下文
			req := httptest.NewRequest("GET", tt.requestPath, nil)
			w := httptest.NewRecorder()
			ctx := core.NewContext(w, req)

			// 调用方法
			result := proxy.buildProxyPath(ctx, tt.targetPath)

			// 验证结果
			if result != tt.expected {
				t.Errorf("%s: 期望 %s, 实际 %s", tt.description, tt.expected, result)
				// 添加调试信息
				fmt.Printf("DEBUG: targetPath=%s, requestPath=%s, expected=%s, actual=%s\n",
					tt.targetPath, tt.requestPath, tt.expected, result)
			} else {
				t.Logf("✅ %s: %s", tt.name, tt.description)
			}
		})
	}
}
