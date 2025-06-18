package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/auth"
)

func TestAuthConfig(t *testing.T) {
	tests := []struct {
		name        string
		config      *auth.AuthConfig
		expectError bool
		description string
	}{
		{
			name: "ValidNoAuthConfig",
			config: &auth.AuthConfig{
				ID:       "test-no-auth",
				Enabled:  true,
				Strategy: auth.StrategyNoAuth,
				Name:     "无认证测试",
			},
			expectError: false,
			description: "有效的无认证配置",
		},
		{
			name: "ValidJWTConfig",
			config: &auth.AuthConfig{
				ID:       "test-jwt",
				Enabled:  true,
				Strategy: auth.StrategyJWT,
				Name:     "JWT认证测试",
				Config: map[string]interface{}{
					"secret":    "test-secret",
					"algorithm": "HS256",
				},
			},
			expectError: false,
			description: "有效的JWT认证配置",
		},
		{
			name: "ValidAPIKeyConfig",
			config: &auth.AuthConfig{
				ID:       "test-apikey",
				Enabled:  true,
				Strategy: auth.StrategyAPIKey,
				Name:     "API Key认证测试",
				Config: map[string]interface{}{
					"param_name": "X-API-Key",
					"in":         "header",
				},
			},
			expectError: false,
			description: "有效的API Key认证配置",
		},
		{
			name: "InvalidStrategy",
			config: &auth.AuthConfig{
				ID:       "test-invalid",
				Enabled:  true,
				Strategy: "invalid-strategy",
				Name:     "无效策略测试",
			},
			expectError: true,
			description: "无效的认证策略应该返回错误",
		},
		{
			name:        "NilConfig",
			config:      nil,
			expectError: true,
			description: "空配置应该返回错误",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := auth.ValidateAuthConfig(tt.config)

			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestNoAuthHandler(t *testing.T) {
	// 创建无认证配置
	config := auth.AuthConfig{
		ID:       "test-no-auth",
		Enabled:  true,
		Strategy: auth.StrategyNoAuth,
		Name:     "无认证测试",
	}

	// 创建无认证处理器
	authenticator, err := auth.NoAuthFromConfig(config)
	require.NoError(t, err, "创建无认证处理器失败")
	require.NotNil(t, authenticator, "认证器不应该为nil")

	// 验证配置
	assert.Equal(t, auth.StrategyNoAuth, authenticator.GetStrategy())
	assert.True(t, authenticator.IsEnabled())
	assert.Equal(t, "无认证测试", authenticator.GetName())

	// 验证配置
	assert.NoError(t, authenticator.Validate())

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()

	// 创建测试上下文
	ctx := core.NewContext(writer, req)

	// 执行认证
	result := authenticator.Handle(ctx)

	// 无认证处理器应该总是返回true
	assert.True(t, result, "无认证处理器应该总是允许访问")
}

func TestAuthStrategy(t *testing.T) {
	// 测试所有认证策略常量
	strategies := []auth.AuthStrategy{
		auth.StrategyNoAuth,
		auth.StrategyJWT,
		auth.StrategyAPIKey,
		auth.StrategyBasic,
		auth.StrategyOAuth2,
		auth.StrategyJWTAndAPIKey,
		auth.StrategyJWTOrAPIKey,
	}

	for _, strategy := range strategies {
		t.Run(string(strategy), func(t *testing.T) {
			config := &auth.AuthConfig{
				ID:       "test-" + string(strategy),
				Enabled:  true,
				Strategy: strategy,
				Name:     "策略测试",
			}

			// 验证策略有效性
			err := auth.ValidateAuthConfig(config)
			assert.NoError(t, err, "有效的认证策略不应该返回错误")
		})
	}
}

func TestDefaultConfigs(t *testing.T) {
	// 测试默认API Key配置
	defaultAPIKey := auth.DefaultAPIKeyConfig
	assert.Equal(t, "default-apikey", defaultAPIKey.ID)
	assert.Equal(t, "api-key", defaultAPIKey.ParamName)
	assert.Equal(t, auth.InHeader, defaultAPIKey.In)
	assert.Equal(t, 401, defaultAPIKey.ErrorStatusCode)

	// 测试默认JWT配置
	defaultJWT := auth.DefaultJWTConfig
	assert.Equal(t, "default-jwt", defaultJWT.ID)
	assert.Equal(t, "HS256", defaultJWT.Algorithm)
	assert.True(t, defaultJWT.VerifyExpiration)
	assert.False(t, defaultJWT.VerifyIssuer)
	assert.Equal(t, 3600, defaultJWT.Expiration)
	assert.Equal(t, 300, defaultJWT.RefreshWindow)
}

func TestKeyLocation(t *testing.T) {
	// 测试API Key位置常量
	locations := []auth.KeyLocation{
		auth.InHeader,
		auth.InQuery,
		auth.InCookie,
	}

	for _, location := range locations {
		t.Run(string(location), func(t *testing.T) {
			// 验证常量值
			assert.NotEmpty(t, string(location), "位置常量不应该为空")
		})
	}
}

func TestAPIKeyItem(t *testing.T) {
	// 创建API Key项目
	item := auth.APIKeyItem{
		Name:  "test-key",
		Value: "secret-value",
		Roles: []string{"admin", "user"},
	}

	// 验证字段
	assert.Equal(t, "test-key", item.Name)
	assert.Equal(t, "secret-value", item.Value)
	assert.Equal(t, []string{"admin", "user"}, item.Roles)
}

func TestBaseAuthenticator(t *testing.T) {
	// 测试基础认证器的公共功能
	config := auth.AuthConfig{
		ID:       "test-base",
		Enabled:  true,
		Strategy: auth.StrategyNoAuth,
		Name:     "基础认证器测试",
	}

	// 创建无认证处理器来测试基础功能
	authenticator, err := auth.NoAuthFromConfig(config)
	require.NoError(t, err)

	// 测试基础方法
	assert.Equal(t, auth.StrategyNoAuth, authenticator.GetStrategy())
	assert.True(t, authenticator.IsEnabled())
	assert.Equal(t, "基础认证器测试", authenticator.GetName())

	// 测试获取配置
	resultConfig := authenticator.GetConfig()
	assert.Equal(t, config.ID, resultConfig.ID)
	assert.Equal(t, config.Strategy, resultConfig.Strategy)
	assert.Equal(t, config.Name, resultConfig.Name)
}

// 基准测试
func BenchmarkNoAuthHandler(b *testing.B) {
	// 创建认证配置
	config := auth.AuthConfig{
		ID:       "bench-no-auth",
		Enabled:  true,
		Strategy: auth.StrategyNoAuth,
		Name:     "无认证基准测试",
	}

	// 创建认证处理器
	authenticator, err := auth.NoAuthFromConfig(config)
	require.NoError(b, err)

	// 创建测试请求
	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()

	// 创建测试上下文
	ctx := core.NewContext(writer, req)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		authenticator.Handle(ctx)
	}
}

func BenchmarkValidateAuthConfig(b *testing.B) {
	config := &auth.AuthConfig{
		ID:       "bench-config",
		Enabled:  true,
		Strategy: auth.StrategyNoAuth,
		Name:     "配置验证基准测试",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		auth.ValidateAuthConfig(config)
	}
}
