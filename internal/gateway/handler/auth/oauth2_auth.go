package auth

import (
	"fmt"
	"gohub/internal/gateway/core"
	"strings"
)

// OAuth2Auth OAuth2认证器
// 实现OAuth2 Bearer Token认证
type OAuth2Auth struct {
	*BaseAuthenticator
	tokenEndpoint      string
	clientID           string
	clientSecret       string
	scope              string
	introspectEndpoint string
}

// OAuth2AuthFromConfig 从配置创建OAuth2认证器
func OAuth2AuthFromConfig(config AuthConfig) (Authenticator, error) {
	// 解析OAuth2配置
	oauth2Config, err := parseOAuth2Config(config.Config)
	if err != nil {
		return nil, fmt.Errorf("解析OAuth2配置失败: %w", err)
	}

	// 创建OAuth2认证器
	auth := NewBaseAuthenticator(StrategyOAuth2, config.Enabled, config.Name)
	if config.Name != "" {
		auth.SetName(config.Name)
	}

	// 验证OAuth2配置
	if oauth2Config == nil {
		return nil, fmt.Errorf("OAuth2配置不能为空")
	}

	// 这里可以添加具体的OAuth2认证逻辑
	return auth, nil
}

// Handle 处理OAuth2认证
func (o *OAuth2Auth) Handle(ctx *core.Context) bool {
	if !o.enabled {
		return true
	}

	// 获取Authorization头中的Bearer token
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		o.handleError(ctx, "Missing Authorization header")
		return false
	}

	// 检查Bearer token格式
	if !strings.HasPrefix(authHeader, "Bearer ") {
		o.handleError(ctx, "Invalid Authorization header format")
		return false
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	if token == "" {
		o.handleError(ctx, "Empty bearer token")
		return false
	}

	// 验证token（通过introspection endpoint）
	tokenInfo, err := o.introspectToken(token)
	if err != nil {
		o.handleError(ctx, err.Error())
		return false
	}

	// 存储认证信息到上下文
	o.storeAuthInfo(ctx, token, tokenInfo)
	return true
}

// Validate 验证OAuth2配置
func (o *OAuth2Auth) Validate() error {
	if o.clientID == "" {
		return fmt.Errorf("OAuth2 client ID不能为空")
	}
	if o.clientSecret == "" {
		return fmt.Errorf("OAuth2 client secret不能为空")
	}
	if o.introspectEndpoint == "" {
		return fmt.Errorf("OAuth2 introspect endpoint不能为空")
	}
	return nil
}

// introspectToken 验证OAuth2 token
func (o *OAuth2Auth) introspectToken(token string) (map[string]interface{}, error) {
	// TODO: 实现真正的token introspection
	// 这里应该调用OAuth2服务器的introspection endpoint
	// 简化实现：检查token不为空
	if token == "" {
		return nil, fmt.Errorf("invalid token")
	}

	return map[string]interface{}{
		"active":    true,
		"client_id": o.clientID,
		"scope":     o.scope,
		"sub":       "oauth2_user",
	}, nil
}

// handleError 处理OAuth2认证错误
func (o *OAuth2Auth) handleError(ctx *core.Context, message string) {
	ctx.AddError(fmt.Errorf("OAuth2 authentication failed: %s", message))
	ctx.Abort(401, map[string]string{
		"error": "Unauthorized: " + message,
	})
}

// storeAuthInfo 存储认证信息到上下文
func (o *OAuth2Auth) storeAuthInfo(ctx *core.Context, token string, tokenInfo map[string]interface{}) {
	ctx.Set("auth_method", "oauth2")
	ctx.Set("oauth2_token", token)
	ctx.Set("oauth2_token_info", tokenInfo)

	if sub, ok := tokenInfo["sub"].(string); ok {
		ctx.Set("user_id", sub)
	}
	if clientID, ok := tokenInfo["client_id"].(string); ok {
		ctx.Set("oauth2_client_id", clientID)
	}
	if scope, ok := tokenInfo["scope"].(string); ok {
		ctx.Set("oauth2_scope", scope)
	}
}

// parseOAuth2Config 解析OAuth2配置
func parseOAuth2Config(configMap map[string]interface{}) (map[string]interface{}, error) {
	if configMap == nil {
		return nil, fmt.Errorf("OAuth2配置不能为空")
	}

	config := make(map[string]interface{})

	if clientID, ok := configMap["client_id"].(string); ok {
		config["client_id"] = clientID
	} else {
		return nil, fmt.Errorf("OAuth2 client_id不能为空")
	}

	if clientSecret, ok := configMap["client_secret"].(string); ok {
		config["client_secret"] = clientSecret
	} else {
		return nil, fmt.Errorf("OAuth2 client_secret不能为空")
	}

	if tokenEndpoint, ok := configMap["token_endpoint"].(string); ok {
		config["token_endpoint"] = tokenEndpoint
	} else {
		config["token_endpoint"] = "" // 设置默认值
	}

	if introspectEndpoint, ok := configMap["introspect_endpoint"].(string); ok {
		config["introspect_endpoint"] = introspectEndpoint
	} else {
		return nil, fmt.Errorf("OAuth2 introspect_endpoint不能为空")
	}

	if scope, ok := configMap["scope"].(string); ok {
		config["scope"] = scope
	} else {
		config["scope"] = "read" // 设置默认scope
	}

	return config, nil
}
