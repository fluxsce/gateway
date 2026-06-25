package auth

import (
	"fmt"
	"gateway/internal/gateway/core"
	"strings"
)

// OAuth2Auth OAuth2认证器
// 认证服务提供方在网关注册中心之外，配置由控制台界面写入数据库，不在 config.yaml 中维护。
// 远端 Token 内省校验尚未实现，启用后将明确拒绝请求而非放行。
type OAuth2Auth struct {
	*BaseAuthenticator
	tokenEndpoint      string
	clientID           string
	clientSecret       string
	scope              string
	introspectEndpoint string
	originalConfig     AuthConfig
}

// oauth2RuntimeConfig OAuth2 运行时配置
type oauth2RuntimeConfig struct {
	TokenEndpoint      string
	ClientID           string
	ClientSecret       string
	Scope              string
	IntrospectEndpoint string
}

// OAuth2AuthFromConfig 从配置创建OAuth2认证器
func OAuth2AuthFromConfig(config AuthConfig) (Authenticator, error) {
	oauth2Config, err := parseOAuth2Config(config.Config)
	if err != nil {
		return nil, fmt.Errorf("解析OAuth2配置失败: %w", err)
	}

	base := NewBaseAuthenticator(StrategyOAuth2, config.Enabled, config.Name)
	if config.Name != "" {
		base.SetName(config.Name)
	}

	return &OAuth2Auth{
		BaseAuthenticator:  base,
		tokenEndpoint:      oauth2Config.TokenEndpoint,
		clientID:           oauth2Config.ClientID,
		clientSecret:       oauth2Config.ClientSecret,
		scope:              oauth2Config.Scope,
		introspectEndpoint: oauth2Config.IntrospectEndpoint,
		originalConfig:     config,
	}, nil
}

// Handle 处理OAuth2认证
func (o *OAuth2Auth) Handle(ctx *core.Context) bool {
	if !o.enabled {
		return true
	}

	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		o.handleError(ctx, "Missing Authorization header")
		return false
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		o.handleError(ctx, "Invalid Authorization header format")
		return false
	}

	token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
	if token == "" {
		o.handleError(ctx, "Empty bearer token")
		return false
	}

	tokenInfo, err := o.introspectToken(token)
	if err != nil {
		o.handleError(ctx, err.Error())
		return false
	}

	o.storeAuthInfo(ctx, token, tokenInfo)
	return true
}

// GetConfig 获取OAuth2认证器配置
func (o *OAuth2Auth) GetConfig() AuthConfig {
	return o.originalConfig
}

// Validate 验证OAuth2配置（仅校验可存储字段，不要求远端内省端点可用）
func (o *OAuth2Auth) Validate() error {
	if o.clientID == "" {
		return fmt.Errorf("OAuth2 client ID不能为空")
	}
	if o.clientSecret == "" {
		return fmt.Errorf("OAuth2 client secret不能为空")
	}
	return nil
}

// introspectToken 远端 Token 校验（暂未实现）
func (o *OAuth2Auth) introspectToken(token string) (map[string]interface{}, error) {
	if token == "" {
		return nil, fmt.Errorf("invalid token")
	}
	return nil, fmt.Errorf("OAuth2 remote token validation is not implemented yet")
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

// parseOAuth2Config 解析OAuth2配置，兼容前端 camelCase 与 snake_case 字段名
func parseOAuth2Config(configMap map[string]interface{}) (*oauth2RuntimeConfig, error) {
	if configMap == nil {
		return nil, fmt.Errorf("OAuth2配置不能为空")
	}

	clientID := configString(configMap, "clientID", "client_id")
	if clientID == "" {
		return nil, fmt.Errorf("OAuth2 client_id不能为空")
	}

	clientSecret := configString(configMap, "clientSecret", "client_secret")
	if clientSecret == "" {
		return nil, fmt.Errorf("OAuth2 client_secret不能为空")
	}

	scope := configString(configMap, "scope")
	if scope == "" {
		scope = "read"
	}

	return &oauth2RuntimeConfig{
		TokenEndpoint:      configString(configMap, "tokenEndpoint", "token_endpoint"),
		ClientID:           clientID,
		ClientSecret:       clientSecret,
		Scope:              scope,
		IntrospectEndpoint: configString(configMap, "introspectEndpoint", "introspect_endpoint"),
	}, nil
}
