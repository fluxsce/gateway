package auth

import (
	"fmt"
	"strings"
)

// AuthenticatorFactory 认证器工厂
// 用于根据配置创建各种类型的认证器
type AuthenticatorFactory struct{}

// NewAuthenticatorFactory 创建认证器工厂
func NewAuthenticatorFactory() *AuthenticatorFactory {
	return &AuthenticatorFactory{}
}

// CreateAuthenticator 根据配置创建认证器
func (f *AuthenticatorFactory) CreateAuthenticator(config AuthConfig) (Authenticator, error) {
	if config.Strategy == "" {
		return nil, fmt.Errorf("认证策略不能为空")
	}

	// 标准化认证策略
	strategy := f.normalizeStrategy(config.Strategy)

	// 根据认证策略委托给具体的实现类创建
	switch strategy {
	case StrategyNoAuth:
		return f.createNoAuth(config)
	case StrategyJWT:
		return f.createJWTAuth(config)
	case StrategyAPIKey:
		return f.createAPIKeyAuth(config)
	case StrategyBasic:
		return f.createBasicAuth(config)
	case StrategyOAuth2:
		return f.createOAuth2Auth(config)
	case StrategyJWTAndAPIKey:
		return f.createCompositeAuth(config, []AuthStrategy{StrategyJWT, StrategyAPIKey}, true)
	case StrategyJWTOrAPIKey:
		return f.createCompositeAuth(config, []AuthStrategy{StrategyJWT, StrategyAPIKey}, false)
	default:
		return nil, fmt.Errorf("不支持的认证策略: %s", config.Strategy)
	}
}

// createNoAuth 创建无认证器
func (f *AuthenticatorFactory) createNoAuth(config AuthConfig) (Authenticator, error) {
	// 直接使用配置创建无认证器
	authConfig := AuthConfig{
		ID:       config.ID,
		Strategy: StrategyNoAuth,
		Name:     config.Name,
		Enabled:  config.Enabled,
		Config:   config.Config,
	}
	return NoAuthFromConfig(authConfig)
}

// createJWTAuth 创建JWT认证器
func (f *AuthenticatorFactory) createJWTAuth(config AuthConfig) (Authenticator, error) {
	authConfig := AuthConfig{
		ID:       config.ID,
		Strategy: StrategyJWT,
		Name:     config.Name,
		Enabled:  config.Enabled,
		Config:   config.Config,
	}
	return JWTAuthFromConfig(authConfig)
}

// createAPIKeyAuth 创建API Key认证器
func (f *AuthenticatorFactory) createAPIKeyAuth(config AuthConfig) (Authenticator, error) {
	authConfig := AuthConfig{
		ID:       config.ID,
		Strategy: StrategyAPIKey,
		Name:     config.Name,
		Enabled:  config.Enabled,
		Config:   config.Config,
	}
	return APIKeyAuthFromConfig(authConfig)
}

// createBasicAuth 创建Basic认证器
func (f *AuthenticatorFactory) createBasicAuth(config AuthConfig) (Authenticator, error) {
	authConfig := AuthConfig{
		ID:       config.ID,
		Strategy: StrategyBasic,
		Name:     config.Name,
		Enabled:  config.Enabled,
		Config:   config.Config,
	}
	return BasicAuthFromConfig(authConfig)
}

// createOAuth2Auth 创建OAuth2认证器
func (f *AuthenticatorFactory) createOAuth2Auth(config AuthConfig) (Authenticator, error) {
	authConfig := AuthConfig{
		ID:       config.ID,
		Strategy: StrategyOAuth2,
		Name:     config.Name,
		Enabled:  config.Enabled,
		Config:   config.Config,
	}
	return OAuth2AuthFromConfig(authConfig)
}

// createCompositeAuth 创建复合认证器
func (f *AuthenticatorFactory) createCompositeAuth(config AuthConfig, strategies []AuthStrategy, allRequired bool) (Authenticator, error) {
	// 创建复合认证器的逻辑
	// 这里可以根据需要实现复合认证器
	// 暂时返回第一个策略的认证器
	if len(strategies) > 0 {
		firstConfig := config
		firstConfig.Strategy = strategies[0]
		return f.CreateAuthenticator(firstConfig)
	}
	return nil, fmt.Errorf("复合认证策略为空")
}

// CreateJWTAuthFromConfig 从配置创建JWT认证器
func (f *AuthenticatorFactory) CreateJWTAuthFromConfig(jwtConfig JWTConfig, name string, enabled bool) (Authenticator, error) {
	// 将JWTConfig转换为通用配置
	config := make(map[string]interface{})
	config["secret"] = jwtConfig.Secret
	config["issuer"] = jwtConfig.Issuer
	config["expiration"] = jwtConfig.Expiration
	config["algorithm"] = jwtConfig.Algorithm
	config["verify_expiration"] = jwtConfig.VerifyExpiration
	config["verify_issuer"] = jwtConfig.VerifyIssuer

	jwtAuthConfig := AuthConfig{
		ID:       jwtConfig.ID,
		Strategy: StrategyJWT,
		Name:     name,
		Enabled:  enabled,
		Config:   config,
	}
	return JWTAuthFromConfig(jwtAuthConfig)
}

// CreateAPIKeyAuthFromConfig 从配置创建API Key认证器
func (f *AuthenticatorFactory) CreateAPIKeyAuthFromConfig(apiKeyConfig APIKeyConfig, name string, enabled bool) (Authenticator, error) {
	// 将APIKeyConfig转换为通用配置
	config := make(map[string]interface{})
	config["param_name"] = apiKeyConfig.ParamName
	config["in"] = string(apiKeyConfig.In)
	config["keys"] = apiKeyConfig.Keys
	config["is_prefix_match"] = apiKeyConfig.IsPrefixMatch

	apiKeyAuthConfig := AuthConfig{
		ID:       apiKeyConfig.ID,
		Strategy: StrategyAPIKey,
		Name:     name,
		Enabled:  enabled,
		Config:   config,
	}
	return APIKeyAuthFromConfig(apiKeyAuthConfig)
}

// normalizeStrategy 标准化认证策略
func (f *AuthenticatorFactory) normalizeStrategy(strategy AuthStrategy) AuthStrategy {
	switch strings.ToLower(string(strategy)) {
	case "none", "no-auth", "noauth":
		return StrategyNoAuth
	case "jwt":
		return StrategyJWT
	case "api-key", "apikey", "api_key":
		return StrategyAPIKey
	case "basic":
		return StrategyBasic
	case "oauth2", "oauth":
		return StrategyOAuth2
	case "jwt-and-api-key", "jwt_and_api_key", "jwt-and-apikey":
		return StrategyJWTAndAPIKey
	case "jwt-or-api-key", "jwt_or_api_key", "jwt-or-apikey":
		return StrategyJWTOrAPIKey
	default:
		return strategy
	}
}

// createCompositeHandler 创建复合认证处理器
func (f *AuthenticatorFactory) createCompositeHandler(config AuthConfig) (Authenticator, error) {
	// TODO: 这里可以创建专门的复合认证器来处理组合策略
	// 现在暂时使用第一个可用的认证器
	if config.Strategy == StrategyJWTAndAPIKey || config.Strategy == StrategyJWTOrAPIKey {
		// 尝试JWT配置
		jwtAuthConfig := AuthConfig{
			ID:       config.ID,
			Strategy: StrategyJWT,
			Name:     "JWT Composite Auth",
			Enabled:  config.Enabled,
			Config:   config.Config,
		}
		if handler, err := JWTAuthFromConfig(jwtAuthConfig); err == nil {
			return handler, nil
		}

		// 尝试API Key配置
		apiKeyAuthConfig := AuthConfig{
			ID:       config.ID,
			Strategy: StrategyAPIKey,
			Name:     "API Key Composite Auth",
			Enabled:  config.Enabled,
			Config:   config.Config,
		}
		if handler, err := APIKeyAuthFromConfig(apiKeyAuthConfig); err == nil {
			return handler, nil
		}
	}
	return nil, fmt.Errorf("复合认证策略需要至少一个有效的认证配置")
}

// GetSupportedStrategies 获取支持的认证策略列表
func GetSupportedStrategies() []AuthStrategy {
	return []AuthStrategy{
		StrategyNoAuth,
		StrategyJWT,
		StrategyAPIKey,
		StrategyBasic,
		StrategyOAuth2,
		StrategyJWTAndAPIKey,
		StrategyJWTOrAPIKey,
	}
}

// GetStrategyDescription 获取策略描述
func GetStrategyDescription(strategy AuthStrategy) string {
	descriptions := map[AuthStrategy]string{
		StrategyNoAuth:       "无需认证",
		StrategyJWT:          "JSON Web Token认证",
		StrategyAPIKey:       "API密钥认证",
		StrategyBasic:        "HTTP Basic认证",
		StrategyOAuth2:       "OAuth 2.0认证",
		StrategyJWTAndAPIKey: "JWT和API密钥同时认证",
		StrategyJWTOrAPIKey:  "JWT或API密钥任一认证",
	}

	if desc, exists := descriptions[strategy]; exists {
		return desc
	}
	return "未知认证策略"
}
