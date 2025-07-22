package auth

import (
	"encoding/base64"
	"fmt"
	"gateway/internal/gateway/core"
	"strings"
)

// BasicAuth HTTP Basic认证器
// 实现HTTP Basic Authentication
type BasicAuth struct {
	*BaseAuthenticator
	username string
	password string
}

// BasicAuthFromConfig 从配置创建Basic认证器
func BasicAuthFromConfig(config AuthConfig) (Authenticator, error) {
	// 解析Basic认证配置
	username, password, err := parseBasicAuthConfig(config.Config)
	if err != nil {
		return nil, fmt.Errorf("解析Basic认证配置失败: %w", err)
	}

	// 创建Basic认证器
	auth := NewBaseAuthenticator(StrategyBasic, config.Enabled, config.Name)
	if config.Name != "" {
		auth.SetName(config.Name)
	}

	// 这里可以添加具体的Basic认证逻辑
	// 验证用户名和密码不为空
	if username == "" || password == "" {
		return nil, fmt.Errorf("Basic认证需要用户名和密码")
	}

	return auth, nil
}

// Handle 处理Basic认证
func (b *BasicAuth) Handle(ctx *core.Context) bool {
	if !b.enabled {
		return true
	}

	// 获取Authorization头
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		b.handleError(ctx, "Missing Authorization header")
		return false
	}

	// 检查Basic认证格式
	if !strings.HasPrefix(authHeader, "Basic ") {
		b.handleError(ctx, "Invalid Authorization header format")
		return false
	}

	// 解码Basic认证信息
	credentials := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64Decode(credentials)
	if err != nil {
		b.handleError(ctx, "Invalid Basic auth encoding")
		return false
	}

	// 分离用户名和密码
	parts := strings.SplitN(decoded, ":", 2)
	if len(parts) != 2 {
		b.handleError(ctx, "Invalid Basic auth format")
		return false
	}

	username, password := parts[0], parts[1]

	// 验证用户名和密码
	if username != b.username || password != b.password {
		b.handleError(ctx, "Invalid credentials")
		return false
	}

	// 存储认证信息到上下文
	b.storeAuthInfo(ctx, username)
	return true
}

// Validate 验证Basic认证配置
func (b *BasicAuth) Validate() error {
	if b.username == "" {
		return fmt.Errorf("Basic认证用户名不能为空")
	}
	if b.password == "" {
		return fmt.Errorf("Basic认证密码不能为空")
	}
	return nil
}

// handleError 处理Basic认证错误
func (b *BasicAuth) handleError(ctx *core.Context, message string) {
	ctx.AddError(fmt.Errorf("Basic authentication failed: %s", message))
	// 设置WWW-Authenticate头，要求客户端进行Basic认证
	ctx.Writer.Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")
	ctx.Abort(401, map[string]string{
		"error": "Unauthorized: " + message,
	})
}

// storeAuthInfo 存储认证信息到上下文
func (b *BasicAuth) storeAuthInfo(ctx *core.Context, username string) {
	ctx.Set("auth_method", "basic")
	ctx.Set("user_id", username)
	ctx.Set("username", username)
}

// parseBasicAuthConfig 解析Basic认证配置
func parseBasicAuthConfig(configMap map[string]interface{}) (string, string, error) {
	if configMap == nil {
		return "", "", fmt.Errorf("Basic认证配置不能为空")
	}

	username, ok := configMap["username"].(string)
	if !ok || username == "" {
		return "", "", fmt.Errorf("Basic认证用户名不能为空")
	}

	password, ok := configMap["password"].(string)
	if !ok || password == "" {
		return "", "", fmt.Errorf("Basic认证密码不能为空")
	}

	return username, password, nil
}

// base64Decode 解码base64字符串
func base64Decode(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
