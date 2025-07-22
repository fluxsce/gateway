package auth

import (
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"
)

// JWTAuth JWT认证处理器
type JWTAuth struct {
	*BaseAuthenticator

	// JWT配置
	jwtConfig JWTConfig

	// 原始配置
	originalConfig AuthConfig
}

// NewJWTAuth 创建JWT认证处理器
func NewJWTAuth(jwtConfig JWTConfig, enabled bool, name string) *JWTAuth {
	base := NewBaseAuthenticator(StrategyJWT, enabled, name)

	// 如果名称为空，设置默认名称
	if name == "" {
		name = "JWT Authenticator"
		base.SetName(name)
	}

	auth := &JWTAuth{
		BaseAuthenticator: base,
		jwtConfig:         jwtConfig,
		originalConfig: AuthConfig{
			ID:       jwtConfig.ID,
			Strategy: StrategyJWT,
			Name:     name,
			Enabled:  enabled,
			Config: map[string]interface{}{
				"secret":               jwtConfig.Secret,
				"issuer":               jwtConfig.Issuer,
				"expiration":           jwtConfig.Expiration,
				"algorithm":            jwtConfig.Algorithm,
				"verify_expiration":    jwtConfig.VerifyExpiration,
				"verify_issuer":        jwtConfig.VerifyIssuer,
				"refresh_window":       jwtConfig.RefreshWindow,
				"include_in_response":  jwtConfig.IncludeInResponse,
				"response_header_name": jwtConfig.ResponseHeaderName,
			},
		},
	}

	return auth
}

// GetStrategy 获取认证策略
func (j *JWTAuth) GetStrategy() AuthStrategy {
	return j.strategy
}

// IsEnabled 是否启用
func (j *JWTAuth) IsEnabled() bool {
	return j.enabled
}

// GetName 获取认证器名称
func (j *JWTAuth) GetName() string {
	return j.name
}

// SetName 设置认证器名称
func (j *JWTAuth) SetName(name string) {
	if name != "" {
		j.name = name
	}
}

// SetEnabled 设置是否启用
func (j *JWTAuth) SetEnabled(enabled bool) {
	j.enabled = enabled
}

// Handle 实现core.Handler接口
func (j *JWTAuth) Handle(ctx *core.Context) bool {
	// 提取JWT token
	token, err := j.extractToken(ctx)
	if err != nil {
		j.handleError(ctx, err.Error())
		return false
	}

	// 验证JWT token
	claims, err := j.validateToken(token)
	if err != nil {
		j.handleError(ctx, err.Error())
		return false
	}

	// 将认证信息存储到上下文
	j.storeAuthInfo(ctx, token, claims)

	return true
}

// extractToken 从请求中提取JWT token
func (j *JWTAuth) extractToken(ctx *core.Context) (string, error) {
	authHeader := ctx.Request.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("missing authorization header")
	}

	// 支持Bearer token格式
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer "), nil
	}

	// 直接使用Authorization头的值
	if authHeader != "" {
		return authHeader, nil
	}

	return "", fmt.Errorf("invalid authorization header format")
}

// validateToken 验证JWT token
// 注意：这是一个简化的实现，生产环境中应该使用专业的JWT库
func (j *JWTAuth) validateToken(token string) (map[string]interface{}, error) {
	if token == "" {
		return nil, fmt.Errorf("empty token")
	}

	// TODO: 这里应该实现真正的JWT验证逻辑
	// 1. 解析JWT token
	// 2. 验证签名
	// 3. 验证过期时间
	// 4. 验证签发者
	// 5. 提取claims

	// 简化实现：只检查token不为空
	claims := map[string]interface{}{
		"sub":   "user123",
		"iss":   j.jwtConfig.Issuer,
		"exp":   0, // 这里应该是实际的过期时间
		"valid": true,
	}

	return claims, nil
}

// storeAuthInfo 存储认证信息到上下文
func (j *JWTAuth) storeAuthInfo(ctx *core.Context, token string, claims map[string]interface{}) {
	ctx.Set("jwt_token", token)
	ctx.Set("jwt_claims", claims)

	// 提取常用的claim信息
	if sub, ok := claims["sub"].(string); ok {
		ctx.Set("jwt_subject", sub)
		ctx.Set("user_id", sub) // 通用的用户ID
	}
	if iss, ok := claims["iss"].(string); ok {
		ctx.Set("jwt_issuer", iss)
	}
	if jti, ok := claims["jti"].(string); ok {
		ctx.Set("jwt_id", jti)
	}

	// 标记认证方式
	ctx.Set("auth_method", "jwt")
}

// handleError 处理JWT认证错误
func (j *JWTAuth) handleError(ctx *core.Context, message string) {
	ctx.AddError(fmt.Errorf("JWT authentication failed: %s", message))
	ctx.Abort(http.StatusUnauthorized, map[string]string{
		"error": "Unauthorized: " + message,
	})
}

// GetConfig 获取JWT配置
func (j *JWTAuth) GetConfig() AuthConfig {
	return j.originalConfig
}

// Validate 验证JWT配置
func (j *JWTAuth) Validate() error {
	return ValidateJWTConfig(&j.jwtConfig)
}

// ValidateJWTConfig 验证JWT配置
func ValidateJWTConfig(config *JWTConfig) error {
	if config == nil {
		return fmt.Errorf("JWT config cannot be nil")
	}

	if config.Secret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}

	if config.Algorithm == "" {
		return fmt.Errorf("JWT algorithm cannot be empty")
	}

	// 验证支持的算法
	supportedAlgorithms := []string{"HS256", "HS384", "HS512", "RS256"}
	algorithmValid := false
	for _, alg := range supportedAlgorithms {
		if config.Algorithm == alg {
			algorithmValid = true
			break
		}
	}
	if !algorithmValid {
		return fmt.Errorf("unsupported JWT algorithm: %s", config.Algorithm)
	}

	if config.Expiration <= 0 {
		return fmt.Errorf("JWT expiration must be positive")
	}

	return nil
}

// JWTAuthFromConfig 从配置创建JWT认证器
func JWTAuthFromConfig(config AuthConfig) (Authenticator, error) {
	jwtConfig, err := parseJWTConfigFromMap(config.Config)
	if err != nil {
		return nil, fmt.Errorf("解析JWT配置失败: %w", err)
	}

	// 设置ID如果没有设置
	if jwtConfig.ID == "" {
		jwtConfig.ID = config.ID
	}

	jwtAuth := NewJWTAuth(*jwtConfig, config.Enabled, config.Name)
	if config.Name != "" {
		jwtAuth.SetName(config.Name)
	}
	jwtAuth.originalConfig = config

	return jwtAuth, nil
}

// parseJWTConfigFromMap 从map解析JWT配置
func parseJWTConfigFromMap(configMap map[string]interface{}) (*JWTConfig, error) {
	if configMap == nil {
		return nil, fmt.Errorf("JWT配置不能为空")
	}

	jwtConfig := &JWTConfig{}

	if secret, ok := configMap["secret"].(string); ok {
		jwtConfig.Secret = secret
	}
	if issuer, ok := configMap["issuer"].(string); ok {
		jwtConfig.Issuer = issuer
	}
	if algorithm, ok := configMap["algorithm"].(string); ok {
		jwtConfig.Algorithm = algorithm
	}
	if expiration, ok := configMap["expiration"].(int); ok {
		jwtConfig.Expiration = expiration
	}
	if expiration, ok := configMap["expiration"].(float64); ok {
		jwtConfig.Expiration = int(expiration)
	}
	if verifyExpiration, ok := configMap["verify_expiration"].(bool); ok {
		jwtConfig.VerifyExpiration = verifyExpiration
	}
	if verifyIssuer, ok := configMap["verify_issuer"].(bool); ok {
		jwtConfig.VerifyIssuer = verifyIssuer
	}

	return jwtConfig, nil
}
