package security

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/security"
)

func TestSecurityConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *security.SecurityConfig
		description string
	}{
		{
			name: "IPAccessConfig",
			config: &security.SecurityConfig{
				ID:      "test-ip-access",
				Enabled: true,
				IPAccess: security.IPAccessConfig{
					Enabled:       true,
					Whitelist:     []string{"192.168.1.0/24", "10.0.0.0/8", "127.0.0.1"},
					WhitelistCIDR: []string{"192.168.1.0/24", "10.0.0.0/8"},
					Blacklist:     []string{"172.16.0.0/16"},
					BlacklistCIDR: []string{"172.16.0.0/16"},
					DefaultPolicy: "allow",
				},
			},
			description: "IP访问控制配置",
		},
		{
			name: "UserAgentConfig",
			config: &security.SecurityConfig{
				ID:      "test-user-agent",
				Enabled: true,
				UserAgentAccess: security.UserAgentAccessConfig{
					Enabled: true,
					Whitelist: []string{
						"Mozilla/*",
						"Chrome/*",
						"Safari/*",
					},
					Blacklist: []string{
						"*bot*",
						"*crawler*",
						"*spider*",
					},
					DefaultPolicy: "allow",
				},
			},
			description: "User-Agent访问控制配置",
		},
		{
			name: "APIAccessConfig",
			config: &security.SecurityConfig{
				ID:      "test-api-access",
				Enabled: true,
				APIAccess: security.APIAccessConfig{
					Enabled: true,
					Whitelist: []string{
						"/api/public/*",
						"/api/v1/users",
					},
					Blacklist: []string{
						"/api/admin/*",
						"/api/internal/*",
					},
					DefaultPolicy: "allow",
				},
			},
			description: "API访问控制配置",
		},
		{
			name: "DomainAccessConfig",
			config: &security.SecurityConfig{
				ID:      "test-domain-access",
				Enabled: true,
				DomainAccess: security.DomainAccessConfig{
					Enabled: true,
					Whitelist: []string{
						"example.com",
						"*.example.com",
						"app.example.org",
					},
					Blacklist: []string{
						"evil.com",
						"*.malicious.org",
					},
					DefaultPolicy:   "deny",
					AllowSubdomains: true,
				},
			},
			description: "域名访问控制配置",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证配置字段
			assert.NotEmpty(t, tt.config.ID, "ID不应该为空")
			assert.True(t, tt.config.Enabled, "应该启用安全检查")
		})
	}
}

func TestSecurityHandler(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-handler",
		Enabled: true,
		IPAccess: security.IPAccessConfig{
			Enabled:       true,
			Whitelist:     []string{"192.168.1.0/24", "127.0.0.1"},
			WhitelistCIDR: []string{"192.168.1.0/24"},
			Blacklist:     []string{"10.0.0.0/8"},
			BlacklistCIDR: []string{"10.0.0.0/8"},
			DefaultPolicy: "allow",
		},
		UserAgentAccess: security.UserAgentAccessConfig{
			Enabled:       true,
			Whitelist:     []string{"Mozilla/*", "Chrome/*"},
			Blacklist:     []string{"*bot*", "*spider*"},
			DefaultPolicy: "allow",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err, "创建安全处理器失败")
	require.NotNil(t, handler, "安全处理器不应该为nil")

	tests := []struct {
		name        string
		remoteAddr  string
		userAgent   string
		expectAllow bool
		description string
	}{
		{
			name:        "WhitelistedIP",
			remoteAddr:  "192.168.1.100:12345",
			userAgent:   "Mozilla/5.0 (compatible)",
			expectAllow: true,
			description: "白名单IP应该被允许",
		},
		{
			name:        "LocalhostIP",
			remoteAddr:  "127.0.0.1:54321",
			userAgent:   "Chrome/91.0.4472.124",
			expectAllow: true,
			description: "本地IP应该被允许",
		},
		{
			name:        "BlacklistedIP",
			remoteAddr:  "10.0.0.50:12345",
			userAgent:   "Mozilla/5.0 (compatible)",
			expectAllow: false,
			description: "黑名单IP应该被拒绝",
		},
		{
			name:        "BlockedUserAgent",
			remoteAddr:  "192.168.1.50:12345",
			userAgent:   "Googlebot/2.1",
			expectAllow: false,
			description: "被阻止的User-Agent应该被拒绝",
		},
		{
			name:        "ValidRequest",
			remoteAddr:  "192.168.1.200:12345",
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectAllow: true,
			description: "有效请求应该被允许",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试请求
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			req.Header.Set("User-Agent", tt.userAgent)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			// 执行安全检查
			result := handler.Handle(ctx)

			// 验证结果
			assert.Equal(t, tt.expectAllow, result, tt.description)
		})
	}
}

func TestIPAccessControl(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-ip-control",
		Enabled: true,
		IPAccess: security.IPAccessConfig{
			Enabled: true,
			Whitelist: []string{
				"192.168.1.0/24",
				"10.0.0.0/8",
				"172.16.0.1",
			},
			WhitelistCIDR: []string{
				"192.168.1.0/24",
				"10.0.0.0/8",
			},
			Blacklist: []string{
				"192.168.1.100",  // 特定IP黑名单
				"203.0.113.0/24", // 测试网段黑名单
			},
			BlacklistCIDR: []string{
				"203.0.113.0/24",
			},
			DefaultPolicy: "deny",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		remoteAddr  string
		expectAllow bool
		description string
	}{
		{
			name:        "WhitelistNetwork",
			remoteAddr:  "192.168.1.50:12345",
			expectAllow: true,
			description: "白名单网段应该被允许",
		},
		{
			name:        "WhitelistSpecificIP",
			remoteAddr:  "172.16.0.1:54321",
			expectAllow: true,
			description: "白名单特定IP应该被允许",
		},
		{
			name:        "BlacklistSpecificIP",
			remoteAddr:  "192.168.1.100:12345",
			expectAllow: false,
			description: "黑名单特定IP应该被拒绝（即使在白名单网段）",
		},
		{
			name:        "BlacklistNetwork",
			remoteAddr:  "203.0.113.50:12345",
			expectAllow: false,
			description: "黑名单网段应该被拒绝",
		},
		{
			name:        "UnlistedIP",
			remoteAddr:  "8.8.8.8:53",
			expectAllow: false,
			description: "不在白名单的IP应该被拒绝",
		},
		{
			name:        "LargeWhitelistNetwork",
			remoteAddr:  "10.1.2.3:12345",
			expectAllow: true,
			description: "大型白名单网段应该被允许",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result, tt.description)
		})
	}
}

func TestUserAgentFiltering(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-user-agent",
		Enabled: true,
		UserAgentAccess: security.UserAgentAccessConfig{
			Enabled: true,
			Whitelist: []string{
				"Mozilla/*",
				"Chrome/*",
				"Safari/*",
				"Edge/*",
			},
			Blacklist: []string{
				"*bot*",
				"*crawler*",
				"*spider*",
				"*scraper*",
			},
			DefaultPolicy: "deny",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userAgent   string
		expectAllow bool
		description string
	}{
		{
			name:        "ValidMozilla",
			userAgent:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
			expectAllow: true,
			description: "有效的Mozilla User-Agent应该被允许",
		},
		{
			name:        "ValidChrome",
			userAgent:   "Chrome/91.0.4472.124 Safari/537.36",
			expectAllow: true,
			description: "有效的Chrome User-Agent应该被允许",
		},
		{
			name:        "ValidSafari",
			userAgent:   "Safari/14.1.1 (WebKit/605.1.15)",
			expectAllow: true,
			description: "有效的Safari User-Agent应该被允许",
		},
		{
			name:        "BlockedBot",
			userAgent:   "Googlebot/2.1 (+http://www.google.com/bot.html)",
			expectAllow: false,
			description: "机器人User-Agent应该被阻止",
		},
		{
			name:        "BlockedCrawler",
			userAgent:   "Baiduspider/2.0",
			expectAllow: false,
			description: "爬虫User-Agent应该被阻止",
		},
		{
			name:        "BlockedSpider",
			userAgent:   "YandexBot/3.0 spider",
			expectAllow: false,
			description: "蜘蛛User-Agent应该被阻止",
		},
		{
			name:        "EmptyUserAgent",
			userAgent:   "",
			expectAllow: false,
			description: "空User-Agent应该被拒绝",
		},
		{
			name:        "UnknownUserAgent",
			userAgent:   "CustomClient/1.0",
			expectAllow: false,
			description: "未知User-Agent应该被拒绝",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345" // 使用安全的IP
			req.Header.Set("User-Agent", tt.userAgent)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result, tt.description)
		})
	}
}

func TestAPIAccessControl(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-api-access",
		Enabled: true,
		APIAccess: security.APIAccessConfig{
			Enabled: true,
			Whitelist: []string{
				"/api/public/*",
				"/api/v1/users",
				"/api/v1/orders/*",
			},
			Blacklist: []string{
				"/api/admin/*",
				"/api/internal/*",
				"/api/debug/*",
			},
			DefaultPolicy: "deny",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		path        string
		expectAllow bool
		description string
	}{
		{
			name:        "AllowedPublicAPI",
			path:        "/api/public/info",
			expectAllow: true,
			description: "公开API应该被允许",
		},
		{
			name:        "AllowedUsersAPI",
			path:        "/api/v1/users",
			expectAllow: true,
			description: "用户API应该被允许",
		},
		{
			name:        "AllowedOrdersAPI",
			path:        "/api/v1/orders/123",
			expectAllow: true,
			description: "订单API应该被允许",
		},
		{
			name:        "BlockedAdminAPI",
			path:        "/api/admin/users",
			expectAllow: false,
			description: "管理员API应该被阻止",
		},
		{
			name:        "BlockedInternalAPI",
			path:        "/api/internal/config",
			expectAllow: false,
			description: "内部API应该被阻止",
		},
		{
			name:        "BlockedDebugAPI",
			path:        "/api/debug/logs",
			expectAllow: false,
			description: "调试API应该被阻止",
		},
		{
			name:        "UnlistedAPI",
			path:        "/api/v2/products",
			expectAllow: false,
			description: "未列出的API应该被拒绝",
		},
		{
			name:        "NonAPIPath",
			path:        "/static/css/style.css",
			expectAllow: false,
			description: "非API路径应该被拒绝（因为默认策略是deny）",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			req.RemoteAddr = "192.168.1.100:12345"
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result, tt.description)
		})
	}
}

func TestDomainAccessControl(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-domain-access",
		Enabled: true,
		DomainAccess: security.DomainAccessConfig{
			Enabled: true,
			Whitelist: []string{
				"example.com",
				"*.example.com",
				"app.mysite.org",
			},
			Blacklist: []string{
				"evil.com",
				"*.malicious.org",
			},
			DefaultPolicy:   "deny",
			AllowSubdomains: true,
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	tests := []struct {
		name        string
		host        string
		expectAllow bool
		description string
	}{
		{
			name:        "AllowedMainDomain",
			host:        "example.com",
			expectAllow: true,
			description: "主域名应该被允许",
		},
		{
			name:        "AllowedSubdomain",
			host:        "api.example.com",
			expectAllow: true,
			description: "允许的子域名应该被允许",
		},
		{
			name:        "AllowedSpecificDomain",
			host:        "app.mysite.org",
			expectAllow: true,
			description: "特定允许的域名应该被允许",
		},
		{
			name:        "BlockedDomain",
			host:        "evil.com",
			expectAllow: false,
			description: "被阻止的域名应该被拒绝",
		},
		{
			name:        "BlockedSubdomain",
			host:        "api.malicious.org",
			expectAllow: false,
			description: "被阻止的子域名应该被拒绝",
		},
		{
			name:        "UnlistedDomain",
			host:        "unknown.com",
			expectAllow: false,
			description: "未列出的域名应该被拒绝",
		},
		{
			name:        "EmptyHost",
			host:        "",
			expectAllow: false,
			description: "空主机名应该被拒绝",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.100:12345"
			req.Host = tt.host
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := handler.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result, tt.description)
		})
	}
}

func TestDisabledSecurity(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-disabled",
		Enabled: false,
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	// 禁用的安全处理器应该允许所有请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "0.0.0.0:12345"                  // 可疑IP
	req.Header.Set("User-Agent", "Malicious-Bot/1.0") // 可疑User-Agent
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	result := handler.Handle(ctx)
	assert.True(t, result, "禁用的安全处理器应该允许所有请求")
}

func TestDefaultSecurityConfig(t *testing.T) {
	defaultConfig := security.DefaultSecurityConfig

	// 验证默认配置
	assert.Equal(t, "default-security", defaultConfig.ID)
	assert.False(t, defaultConfig.Enabled, "默认应该禁用安全检查")
}

func TestSecurityInterface(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-interface",
		Enabled: true,
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	// 验证处理器实现了SecurityHandler接口
	var _ security.SecurityHandler = handler

	// 测试接口方法
	assert.True(t, handler.IsEnabled())
	assert.Equal(t, *config, handler.GetConfig())
	assert.NoError(t, handler.Validate())
}

func TestNilConfigHandling(t *testing.T) {
	// 使用工厂创建默认处理器
	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateDefaultSecurityHandler()
	require.NoError(t, err)
	assert.NotNil(t, handler, "处理器不应该为nil")
	config := handler.GetConfig()
	assert.NotNil(t, config, "配置不应该为nil")
}

// 并发测试
func TestSecurityConcurrency(t *testing.T) {
	config := &security.SecurityConfig{
		ID:      "test-concurrent",
		Enabled: true,
		IPAccess: security.IPAccessConfig{
			Enabled:       true,
			WhitelistCIDR: []string{"192.168.0.0/16"},
			BlacklistCIDR: []string{"10.0.0.0/8"},
			DefaultPolicy: "deny",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(t, err)

	// 并发执行安全检查
	const numGoroutines = 100
	const requestsPerGoroutine = 10

	results := make(chan bool, numGoroutines*requestsPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			for j := 0; j < requestsPerGoroutine; j++ {
				req := httptest.NewRequest("GET", "/test", nil)
				req.RemoteAddr = "192.168.1.100:12345" // 安全IP
				req.Header.Set("User-Agent", "Mozilla/5.0")
				writer := httptest.NewRecorder()
				ctx := core.NewContext(writer, req)

				result := handler.Handle(ctx)
				results <- result
			}
		}(i)
	}

	// 收集结果
	successCount := 0
	for i := 0; i < numGoroutines*requestsPerGoroutine; i++ {
		if <-results {
			successCount++
		}
	}

	// 验证所有安全请求都成功
	assert.Equal(t, numGoroutines*requestsPerGoroutine, successCount, "所有安全请求都应该成功")
}

// 基准测试
func BenchmarkSecurityHandler(b *testing.B) {
	config := &security.SecurityConfig{
		ID:      "bench-security",
		Enabled: true,
		IPAccess: security.IPAccessConfig{
			Enabled:       true,
			WhitelistCIDR: []string{"192.168.0.0/16", "10.0.0.0/8"},
			BlacklistCIDR: []string{"172.16.0.0/16"},
			DefaultPolicy: "allow",
		},
		UserAgentAccess: security.UserAgentAccessConfig{
			Enabled:       true,
			Whitelist:     []string{"Mozilla/*", "Chrome/*", "Safari/*"},
			Blacklist:     []string{"*bot*", "*crawler*"},
			DefaultPolicy: "allow",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(b, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible)")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		handler.Handle(ctx)
	}
}

func BenchmarkIPValidation(b *testing.B) {
	config := &security.SecurityConfig{
		ID:      "bench-ip",
		Enabled: true,
		IPAccess: security.IPAccessConfig{
			Enabled: true,
			WhitelistCIDR: []string{
				"192.168.0.0/16",
				"10.0.0.0/8",
				"172.16.0.0/12",
				"127.0.0.0/8",
			},
			BlacklistCIDR: []string{
				"203.0.113.0/24",
				"198.51.100.0/24",
			},
			DefaultPolicy: "allow",
		},
	}

	factory := security.NewSecurityHandlerFactory()
	handler, err := factory.CreateSecurityHandler(*config)
	require.NoError(b, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.100:12345"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := httptest.NewRecorder()
		ctx := core.NewContext(writer, req)
		handler.Handle(ctx)
	}
}
