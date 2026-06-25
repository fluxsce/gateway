package auth

import (
	"fmt"
	"gateway/internal/gateway/core"
)

// AuthStrategy 认证策略类型
type AuthStrategy string

const (
	// StrategyNoAuth 不需要认证
	StrategyNoAuth AuthStrategy = "none"
	// StrategyJWT 使用JWT认证
	StrategyJWT AuthStrategy = "jwt"
	// StrategyAPIKey 使用API Key认证
	StrategyAPIKey AuthStrategy = "api-key"
	// StrategyBasic 使用Basic认证
	StrategyBasic AuthStrategy = "basic"
	// StrategyOAuth2 使用OAuth2认证
	StrategyOAuth2 AuthStrategy = "oauth2"
	// StrategyBearerToken 使用 Bearer Token 认证
	StrategyBearerToken AuthStrategy = "bearer-token"
	// StrategyJWTAndAPIKey 同时需要JWT和API Key认证（都需要通过）
	StrategyJWTAndAPIKey AuthStrategy = "jwt-and-api-key"
	// StrategyJWTOrAPIKey 使用JWT或API Key认证（任一通过即可）
	StrategyJWTOrAPIKey AuthStrategy = "jwt-or-api-key"
)

// Authenticator 认证器接口
// 所有认证器都必须实现此接口
type Authenticator interface {
	// Handle 处理认证
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - bool: 认证是否成功
	Handle(ctx *core.Context) bool

	// GetStrategy 获取认证策略
	// 返回值:
	// - AuthStrategy: 认证策略类型
	GetStrategy() AuthStrategy

	// IsEnabled 是否启用
	// 返回值:
	// - bool: 是否启用
	IsEnabled() bool

	// GetName 获取认证器名称
	// 返回值:
	// - string: 认证器名称
	GetName() string

	// Validate 验证配置
	// 返回值:
	// - error: 验证错误
	Validate() error

	// GetConfig 获取认证器配置
	// 返回值:
	// - AuthConfig: 认证器的配置信息
	GetConfig() AuthConfig
}

// AuthConfig 认证配置
type AuthConfig struct {
	// 认证配置ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`
	// 是否启用认证
	Enabled bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	// 认证策略
	Strategy AuthStrategy `yaml:"strategy" json:"strategy" mapstructure:"strategy"`
	// 认证器名称
	Name string `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`
	// 不需要认证的路径
	ExcludedPaths []string `yaml:"excluded_paths,omitempty" json:"excluded_paths,omitempty" mapstructure:"excluded_paths,omitempty"`
	// 认证器配置（具体内容由各认证器子模块定义）
	Config map[string]interface{} `yaml:"config,omitempty" json:"config,omitempty" mapstructure:"config,omitempty"`
}

// KeyLocation API Key位置
type KeyLocation string

const (
	// InHeader API Key在请求头中
	InHeader KeyLocation = "header"
	// InQuery API Key在查询参数中
	InQuery KeyLocation = "query"
	// InCookie API Key在Cookie中
	InCookie KeyLocation = "cookie"
)

// APIKeyConfig API Key 认证配置（param_name + in + key）
type APIKeyConfig struct {
	// API Key配置ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`
	// API Key 参数名（Header 名 / Query 名 / Cookie 名）
	ParamName string `yaml:"param_name" json:"param_name" mapstructure:"param_name"`
	// API Key 位置：header, query, cookie
	In KeyLocation `yaml:"in" json:"in" mapstructure:"in"`
	// 期望的 API Key 值
	Key string `yaml:"key" json:"key" mapstructure:"key"`
	// 认证失败 HTTP 状态码
	ErrorStatusCode int `yaml:"error_status_code,omitempty" json:"error_status_code,omitempty" mapstructure:"error_status_code,omitempty"`
}

// BearerTokenConfig Bearer Token 认证配置（Authorization: Bearer <token>）。
type BearerTokenConfig struct {
	// 认证配置ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`
	// 期望的 Bearer Token 值
	Token string `yaml:"token" json:"token" mapstructure:"token"`
	// 认证失败 HTTP 状态码
	ErrorStatusCode int `yaml:"error_status_code,omitempty" json:"error_status_code,omitempty" mapstructure:"error_status_code,omitempty"`
}

// JWTConfig JWT认证配置
type JWTConfig struct {
	// JWT配置ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`
	// JWT 对称密钥（HMAC 算法使用，任意非空字符串）
	Secret string `yaml:"secret" json:"secret" mapstructure:"secret"`
	// 签发者
	Issuer string `yaml:"issuer" json:"issuer" mapstructure:"issuer"`
	// 过期时间（秒）
	Expiration int `yaml:"expiration" json:"expiration" mapstructure:"expiration"`
	// 签名算法：HS256/HS384/HS512 使用 secret；RS256/RS384/RS512 使用 publicKey
	Algorithm string `yaml:"algorithm" json:"algorithm" mapstructure:"algorithm"`
	// RSA 等非对称算法使用的公钥（PEM 格式）
	PublicKey string `yaml:"public_key" json:"public_key" mapstructure:"public_key"`
	// 是否验证过期时间
	VerifyExpiration bool `yaml:"verify_expiration" json:"verify_expiration" mapstructure:"verify_expiration"`
	// 是否验证签发者
	VerifyIssuer bool `yaml:"verify_issuer" json:"verify_issuer" mapstructure:"verify_issuer"`
	// 强制刷新时间窗口（秒），token过期前多少秒可以刷新
	RefreshWindow int `yaml:"refresh_window" json:"refresh_window" mapstructure:"refresh_window"`
	// 是否在响应中包含token信息
	IncludeInResponse bool `yaml:"include_in_response" json:"include_in_response" mapstructure:"include_in_response"`
	// token在响应中的头部名称
	ResponseHeaderName string `yaml:"response_header_name" json:"response_header_name" mapstructure:"response_header_name"`
}

// DefaultAPIKeyConfig 默认API Key配置
var DefaultAPIKeyConfig = APIKeyConfig{
	ID:              "default-apikey",
	ParamName:       "api-key",
	In:              InHeader,
	Key:             "",
	ErrorStatusCode: 401,
}

// DefaultBearerTokenConfig 默认 Bearer Token 配置
var DefaultBearerTokenConfig = BearerTokenConfig{
	ID:              "default-bearer-token",
	Token:           "",
	ErrorStatusCode: 401,
}

// DefaultJWTConfig 默认JWT配置
var DefaultJWTConfig = JWTConfig{
	ID:                 "default-jwt",
	Algorithm:          "HS256",
	VerifyExpiration:   true,
	VerifyIssuer:       false,
	Expiration:         3600,
	RefreshWindow:      300,
	IncludeInResponse:  false,
	ResponseHeaderName: "X-JWT-Token",
}

// ValidateAuthConfig 验证认证配置
func ValidateAuthConfig(config *AuthConfig) error {
	if config == nil {
		return fmt.Errorf("auth config cannot be nil")
	}

	// 验证策略是否有效
	validStrategies := []AuthStrategy{
		StrategyNoAuth, StrategyJWT, StrategyAPIKey, StrategyBasic,
		StrategyOAuth2, StrategyBearerToken, StrategyJWTAndAPIKey, StrategyJWTOrAPIKey,
	}

	strategyValid := false
	for _, strategy := range validStrategies {
		if config.Strategy == strategy {
			strategyValid = true
			break
		}
	}

	if !strategyValid {
		return fmt.Errorf("invalid auth strategy: %s", config.Strategy)
	}

	return nil
}

// BaseAuthenticator 基础认证器
// 提供认证器的基础实现和通用功能
type BaseAuthenticator struct {
	strategy       AuthStrategy
	enabled        bool
	name           string
	originalConfig AuthConfig
}

// NewBaseAuthenticator 创建基础认证器
func NewBaseAuthenticator(strategy AuthStrategy, enabled bool, name string) *BaseAuthenticator {
	config := AuthConfig{
		ID:       name, // 使用name作为默认ID
		Strategy: strategy,
		Name:     name,
		Enabled:  enabled,
		Config:   make(map[string]interface{}),
	}

	return &BaseAuthenticator{
		strategy:       strategy,
		enabled:        enabled,
		name:           name,
		originalConfig: config,
	}
}

// GetStrategy 获取认证策略
func (b *BaseAuthenticator) GetStrategy() AuthStrategy {
	return b.strategy
}

// IsEnabled 是否启用
func (b *BaseAuthenticator) IsEnabled() bool {
	return b.enabled
}

// GetName 获取认证器名称
func (b *BaseAuthenticator) GetName() string {
	return b.name
}

// GetConfig 获取认证器配置
func (b *BaseAuthenticator) GetConfig() AuthConfig {
	return b.originalConfig
}

// SetName 设置认证器名称
func (b *BaseAuthenticator) SetName(name string) {
	if name != "" {
		b.name = name
		b.originalConfig.Name = name
	}
}

// SetEnabled 设置是否启用
func (b *BaseAuthenticator) SetEnabled(enabled bool) {
	b.enabled = enabled
	b.originalConfig.Enabled = enabled
}

// Handle 处理认证（基础实现：总是允许通过）
func (b *BaseAuthenticator) Handle(ctx *core.Context) bool {
	return true
}

// Validate 验证配置（基础实现：总是通过验证）
func (b *BaseAuthenticator) Validate() error {
	return nil
}

// NoAuthFromConfig 从配置创建无认证器
func NoAuthFromConfig(config AuthConfig) (Authenticator, error) {
	auth := NewBaseAuthenticator(StrategyNoAuth, config.Enabled, config.Name)
	auth.originalConfig = config
	return auth, nil
}
