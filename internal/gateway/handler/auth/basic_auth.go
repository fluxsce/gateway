package auth

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"
)

// BasicAuth HTTP Basic 认证器
type BasicAuth struct {
	*BaseAuthenticator
	username       string
	password       string
	originalConfig AuthConfig
}

// BasicAuthFromConfig 从配置创建 Basic 认证器
func BasicAuthFromConfig(config AuthConfig) (Authenticator, error) {
	username, password, err := parseBasicAuthConfig(config.Config)
	if err != nil {
		return nil, fmt.Errorf("解析Basic认证配置失败: %w", err)
	}

	base := NewBaseAuthenticator(StrategyBasic, config.Enabled, config.Name)
	if config.Name != "" {
		base.SetName(config.Name)
	}

	return &BasicAuth{
		BaseAuthenticator: base,
		username:          username,
		password:          password,
		originalConfig:    config,
	}, nil
}

// GetConfig 获取认证器配置
func (b *BasicAuth) GetConfig() AuthConfig {
	return b.originalConfig
}

// Handle 处理 Basic 认证
func (b *BasicAuth) Handle(ctx *core.Context) bool {
	if !b.enabled {
		return true
	}
	if ctx == nil || ctx.Request == nil {
		b.handleError(ctx, "invalid request context")
		return false
	}

	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		b.handleError(ctx, "Missing Authorization header")
		return false
	}

	// RFC 7235：认证方案名大小写不敏感
	if !strings.HasPrefix(strings.ToLower(authHeader), "basic ") {
		b.handleError(ctx, "Invalid Authorization header format")
		return false
	}

	credentials := strings.TrimSpace(authHeader[len("Basic "):])
	decoded, err := base64.StdEncoding.DecodeString(credentials)
	if err != nil {
		b.handleError(ctx, "Invalid Basic auth encoding")
		return false
	}

	parts := strings.SplitN(string(decoded), ":", 2)
	if len(parts) != 2 {
		b.handleError(ctx, "Invalid Basic auth format")
		return false
	}

	username, password := parts[0], parts[1]
	if !secureCompareAPIKey(username, b.username) || !secureCompareAPIKey(password, b.password) {
		b.handleError(ctx, "Invalid credentials")
		return false
	}

	b.storeAuthInfo(ctx, username)
	return true
}

// Validate 验证 Basic 认证配置
func (b *BasicAuth) Validate() error {
	if strings.TrimSpace(b.username) == "" {
		return fmt.Errorf("Basic认证用户名不能为空")
	}
	if strings.TrimSpace(b.password) == "" {
		return fmt.Errorf("Basic认证密码不能为空")
	}
	return nil
}

// handleError 处理 Basic 认证错误
func (b *BasicAuth) handleError(ctx *core.Context, message string) {
	if ctx != nil {
		ctx.AddError(fmt.Errorf("Basic authentication failed: %s", message))
		if ctx.Writer != nil {
			ctx.Writer.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
		}
		ctx.Abort(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized: " + message,
		})
	}
}

// storeAuthInfo 存储认证信息到上下文
func (b *BasicAuth) storeAuthInfo(ctx *core.Context, username string) {
	ctx.Set("auth_method", "basic")
	ctx.Set("user_id", username)
	ctx.Set("username", username)
}

// parseBasicAuthConfig 解析 Basic 认证配置
func parseBasicAuthConfig(configMap map[string]interface{}) (string, string, error) {
	if configMap == nil {
		return "", "", errors.New("Basic认证配置不能为空")
	}

	username := strings.TrimSpace(configString(configMap, "username"))
	if username == "" {
		return "", "", errors.New("Basic认证用户名不能为空")
	}

	password := strings.TrimSpace(configString(configMap, "password"))
	if password == "" {
		return "", "", errors.New("Basic认证密码不能为空")
	}

	return username, password, nil
}
