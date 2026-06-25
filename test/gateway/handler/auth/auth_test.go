package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/auth"
)

func signTestJWT(secret, issuer string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub": "user123",
		"exp": exp.Unix(),
	}
	if issuer != "" {
		claims["iss"] = issuer
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		panic(err)
	}
	return tokenString
}

func newJWTAuthenticator(t *testing.T, config map[string]interface{}) auth.Authenticator {
	t.Helper()
	authenticator, err := auth.JWTAuthFromConfig(auth.AuthConfig{
		ID:       "test-jwt",
		Enabled:  true,
		Strategy: auth.StrategyJWT,
		Name:     "JWT认证测试",
		Config:   config,
	})
	require.NoError(t, err)
	return authenticator
}

func TestJWTAuthHandle(t *testing.T) {
	secret := "test-secret"
	issuer := "gateway-test"

	validToken := signTestJWT(secret, issuer, time.Now().Add(time.Hour))
	expiredToken := signTestJWT(secret, issuer, time.Now().Add(-time.Hour))
	wrongSecretToken := signTestJWT("wrong-secret", issuer, time.Now().Add(time.Hour))

	baseConfig := map[string]interface{}{
		"secret":            secret,
		"algorithm":         "HS256",
		"issuer":            issuer,
		"verify_issuer":     true,
		"verify_expiration": true,
		"expiration":        3600,
	}

	tests := []struct {
		name        string
		token       string
		expectAllow bool
	}{
		{
			name:        "ValidToken",
			token:       validToken,
			expectAllow: true,
		},
		{
			name:        "ExpiredToken",
			token:       expiredToken,
			expectAllow: false,
		},
		{
			name:        "InvalidSignature",
			token:       wrongSecretToken,
			expectAllow: false,
		},
		{
			name:        "MalformedToken",
			token:       "not-a-jwt",
			expectAllow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := newJWTAuthenticator(t, baseConfig)

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := authenticator.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result)

			if tt.expectAllow {
				claims, ok := ctx.Get("jwt_claims")
				assert.True(t, ok)
				claimsMap, ok := claims.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "user123", claimsMap["sub"])
			}
		})
	}
}

func TestJWTAuthMissingAuthorizationHeader(t *testing.T) {
	authenticator := newJWTAuthenticator(t, map[string]interface{}{
		"secret":     "test-secret",
		"algorithm":  "HS256",
		"expiration": 3600,
	})

	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	assert.False(t, authenticator.Handle(ctx))
}

func TestJWTAuthInvalidIssuer(t *testing.T) {
	secret := "test-secret"
	token := signTestJWT(secret, "other-issuer", time.Now().Add(time.Hour))

	authenticator := newJWTAuthenticator(t, map[string]interface{}{
		"secret":        secret,
		"algorithm":     "HS256",
		"issuer":        "gateway-test",
		"verify_issuer": true,
		"expiration":    3600,
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	assert.False(t, authenticator.Handle(ctx))
}

func TestOAuth2AuthNotImplemented(t *testing.T) {
	authenticator, err := auth.OAuth2AuthFromConfig(auth.AuthConfig{
		ID:       "test-oauth2",
		Enabled:  true,
		Strategy: auth.StrategyOAuth2,
		Name:     "OAuth2测试",
		Config: map[string]interface{}{
			"clientID":      "test-client",
			"clientSecret":  "test-secret",
			"tokenEndpoint": "https://auth.example.com/oauth/token",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer some-token")
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)

	assert.False(t, authenticator.Handle(ctx))
}

func newBearerAuthenticator(t *testing.T, config map[string]interface{}) auth.Authenticator {
	t.Helper()
	authenticator, err := auth.BearerTokenAuthFromConfig(auth.AuthConfig{
		ID:       "test-bearer",
		Enabled:  true,
		Strategy: auth.StrategyBearerToken,
		Name:     "Bearer认证测试",
		Config:   config,
	})
	require.NoError(t, err)
	return authenticator
}

func TestBearerTokenAuthHandle(t *testing.T) {
	authenticator := newBearerAuthenticator(t, map[string]interface{}{
		"token": "demo-bearer-token",
	})

	tests := []struct {
		name        string
		authHeader  string
		expectAllow bool
	}{
		{name: "ValidToken", authHeader: "Bearer demo-bearer-token", expectAllow: true},
		{name: "InvalidToken", authHeader: "Bearer wrong-token", expectAllow: false},
		{name: "LowerCaseBearer", authHeader: "bearer demo-bearer-token", expectAllow: true},
		{name: "MissingHeader", authHeader: "", expectAllow: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			assert.Equal(t, tt.expectAllow, authenticator.Handle(ctx))
		})
	}
}

func TestBearerTokenAuthFromConfigRequiresToken(t *testing.T) {
	_, err := auth.BearerTokenAuthFromConfig(auth.AuthConfig{
		ID:       "test-bearer",
		Enabled:  true,
		Strategy: auth.StrategyBearerToken,
		Config: map[string]interface{}{
			"token": "",
		},
	})
	assert.Error(t, err)
}

func newAPIKeyAuthenticator(t *testing.T, config map[string]interface{}) auth.Authenticator {
	t.Helper()
	authenticator, err := auth.APIKeyAuthFromConfig(auth.AuthConfig{
		ID:       "test-apikey",
		Enabled:  true,
		Strategy: auth.StrategyAPIKey,
		Name:     "API Key认证测试",
		Config:   config,
	})
	require.NoError(t, err)
	return authenticator
}

func TestAPIKeyAuthFromConfigRequiresKey(t *testing.T) {
	_, err := auth.APIKeyAuthFromConfig(auth.AuthConfig{
		ID:       "test-apikey",
		Enabled:  true,
		Strategy: auth.StrategyAPIKey,
		Name:     "API Key认证测试",
		Config: map[string]interface{}{
			"param_name": "X-API-Key",
			"in":         "header",
			"key":        "",
		},
	})
	assert.Error(t, err)
}

func TestAPIKeyAuthHandle(t *testing.T) {
	baseConfig := map[string]interface{}{
		"param_name": "X-API-Key",
		"in":         "header",
		"key":        "secret-key",
	}

	tests := []struct {
		name        string
		headerKey   string
		headerValue string
		expectAllow bool
	}{
		{name: "ValidKey", headerKey: "X-API-Key", headerValue: "secret-key", expectAllow: true},
		{name: "InvalidKey", headerKey: "X-API-Key", headerValue: "wrong-key", expectAllow: false},
		{name: "MissingKey", headerKey: "", headerValue: "", expectAllow: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authenticator := newAPIKeyAuthenticator(t, baseConfig)
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.headerKey != "" {
				req.Header.Set(tt.headerKey, tt.headerValue)
			}
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)

			result := authenticator.Handle(ctx)
			assert.Equal(t, tt.expectAllow, result)
		})
	}
}

func TestAPIKeyAuthQueryParam(t *testing.T) {
	authenticator := newAPIKeyAuthenticator(t, map[string]interface{}{
		"param_name": "api_key",
		"in":         "query",
		"key":        "query-secret",
	})

	req := httptest.NewRequest("GET", "/test?api_key=query-secret", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	assert.True(t, authenticator.Handle(ctx))
}

func TestAPIKeyAuthDisabledSkipsValidation(t *testing.T) {
	authenticator, err := auth.APIKeyAuthFromConfig(auth.AuthConfig{
		ID:       "test-apikey",
		Enabled:  false,
		Strategy: auth.StrategyAPIKey,
		Name:     "API Key认证测试",
		Config: map[string]interface{}{
			"param_name": "X-API-Key",
			"in":         "header",
			"key":        "secret-key",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	assert.True(t, authenticator.Handle(ctx))
}

func basicAuthHeader(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

func newBasicAuthenticator(t *testing.T, username, password string) auth.Authenticator {
	t.Helper()
	authenticator, err := auth.BasicAuthFromConfig(auth.AuthConfig{
		ID:       "test-basic",
		Enabled:  true,
		Strategy: auth.StrategyBasic,
		Name:     "Basic认证测试",
		Config: map[string]interface{}{
			"username": username,
			"password": password,
		},
	})
	require.NoError(t, err)
	return authenticator
}

func TestBasicAuthFromConfigRequiresCredentials(t *testing.T) {
	_, err := auth.BasicAuthFromConfig(auth.AuthConfig{
		Enabled:  true,
		Strategy: auth.StrategyBasic,
		Config: map[string]interface{}{
			"username": "admin",
			"password": "",
		},
	})
	assert.Error(t, err)
}

func TestBasicAuthHandle(t *testing.T) {
	authenticator := newBasicAuthenticator(t, "admin", "secret")

	tests := []struct {
		name        string
		setupReq    func(*http.Request)
		expectAllow bool
	}{
		{
			name: "ValidCredentials",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", basicAuthHeader("admin", "secret"))
			},
			expectAllow: true,
		},
		{
			name: "InvalidPassword",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", basicAuthHeader("admin", "wrong"))
			},
			expectAllow: false,
		},
		{
			name:        "MissingHeader",
			setupReq:    func(req *http.Request) {},
			expectAllow: false,
		},
		{
			name: "CaseInsensitiveScheme",
			setupReq: func(req *http.Request) {
				req.Header.Set("Authorization", "basic "+strings.TrimPrefix(basicAuthHeader("admin", "secret"), "Basic "))
			},
			expectAllow: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			tt.setupReq(req)
			writer := httptest.NewRecorder()
			ctx := core.NewContext(writer, req)
			assert.Equal(t, tt.expectAllow, authenticator.Handle(ctx))
		})
	}
}

func TestBasicAuthDisabledSkipsValidation(t *testing.T) {
	authenticator, err := auth.BasicAuthFromConfig(auth.AuthConfig{
		Enabled:  false,
		Strategy: auth.StrategyBasic,
		Config: map[string]interface{}{
			"username": "admin",
			"password": "secret",
		},
	})
	require.NoError(t, err)

	req := httptest.NewRequest("GET", "/test", nil)
	writer := httptest.NewRecorder()
	ctx := core.NewContext(writer, req)
	assert.True(t, authenticator.Handle(ctx))
}

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
					"key":        "test-key",
				},
			},
			expectError: false,
			description: "有效的API Key认证配置",
		},
		{
			name: "ValidBearerTokenConfig",
			config: &auth.AuthConfig{
				ID:       "test-bearer",
				Enabled:  true,
				Strategy: auth.StrategyBearerToken,
				Name:     "Bearer认证测试",
				Config: map[string]interface{}{
					"token": "demo-bearer-token",
				},
			},
			expectError: false,
			description: "有效的Bearer Token认证配置",
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
		auth.StrategyBearerToken,
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
