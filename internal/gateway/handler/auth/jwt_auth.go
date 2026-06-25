package auth

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"

	"github.com/golang-jwt/jwt/v4"
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
				"public_key":           jwtConfig.PublicKey,
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
	if !j.enabled {
		return true
	}

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
		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		if token == "" {
			return "", fmt.Errorf("empty bearer token")
		}
		return token, nil
	}

	return "", fmt.Errorf("invalid authorization header format")
}

// validateToken 验证JWT token：校验签名、过期时间与签发者
func (j *JWTAuth) validateToken(tokenString string) (map[string]interface{}, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("empty token")
	}

	expectedMethod, err := j.resolveSigningMethod()
	if err != nil {
		return nil, err
	}

	claims := jwt.MapClaims{}
	parserOpts := []jwt.ParserOption{
		jwt.WithValidMethods([]string{expectedMethod.Alg()}),
	}
	if !j.jwtConfig.VerifyExpiration {
		parserOpts = append(parserOpts, jwt.WithoutClaimsValidation())
	}

	parser := jwt.NewParser(parserOpts...)
	token, err := parser.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return j.verificationKey(token)
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if j.jwtConfig.VerifyIssuer {
		if err := j.verifyIssuer(claims); err != nil {
			return nil, err
		}
	}

	return mapClaimsToMap(claims), nil
}

// resolveSigningMethod 根据配置解析签名算法
func (j *JWTAuth) resolveSigningMethod() (jwt.SigningMethod, error) {
	switch strings.ToUpper(j.jwtConfig.Algorithm) {
	case "HS256":
		return jwt.SigningMethodHS256, nil
	case "HS384":
		return jwt.SigningMethodHS384, nil
	case "HS512":
		return jwt.SigningMethodHS512, nil
	case "RS256":
		return jwt.SigningMethodRS256, nil
	case "RS384":
		return jwt.SigningMethodRS384, nil
	case "RS512":
		return jwt.SigningMethodRS512, nil
	default:
		return nil, fmt.Errorf("unsupported JWT algorithm: %s", j.jwtConfig.Algorithm)
	}
}

// verificationKey 返回验签所需的密钥
func (j *JWTAuth) verificationKey(token *jwt.Token) (interface{}, error) {
	switch strings.ToUpper(j.jwtConfig.Algorithm) {
	case "HS256", "HS384", "HS512":
		if j.jwtConfig.Secret == "" {
			return nil, fmt.Errorf("JWT secret is not configured")
		}
		return []byte(j.jwtConfig.Secret), nil
	case "RS256", "RS384", "RS512":
		if j.jwtConfig.PublicKey == "" {
			return nil, fmt.Errorf("JWT public key is not configured for %s", j.jwtConfig.Algorithm)
		}
		return parseRSAPublicKey(j.jwtConfig.PublicKey)
	default:
		return nil, fmt.Errorf("unsupported JWT algorithm: %s", j.jwtConfig.Algorithm)
	}
}

// verifyIssuer 校验 token 签发者
func (j *JWTAuth) verifyIssuer(claims jwt.MapClaims) error {
	if j.jwtConfig.Issuer == "" {
		return fmt.Errorf("issuer verification enabled but issuer is not configured")
	}

	iss, _ := claims["iss"].(string)
	if iss != j.jwtConfig.Issuer {
		return fmt.Errorf("invalid issuer")
	}
	return nil
}

// parseRSAPublicKey 解析 PEM 格式的 RSA 公钥
func parseRSAPublicKey(pemData string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemData))
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM public key")
	}

	switch block.Type {
	case "RSA PUBLIC KEY":
		pub, err := x509.ParsePKCS1PublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse RSA public key: %w", err)
		}
		return pub, nil
	case "PUBLIC KEY":
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKIX public key: %w", err)
		}
		pub, ok := key.(*rsa.PublicKey)
		if !ok {
			return nil, fmt.Errorf("public key is not RSA")
		}
		return pub, nil
	default:
		return nil, fmt.Errorf("unsupported public key type: %s", block.Type)
	}
}

// mapClaimsToMap 将 jwt.MapClaims 转为通用 map
func mapClaimsToMap(claims jwt.MapClaims) map[string]interface{} {
	result := make(map[string]interface{}, len(claims))
	for k, v := range claims {
		result[k] = v
	}
	return result
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

	if config.Algorithm == "" {
		return fmt.Errorf("JWT algorithm cannot be empty")
	}

	algorithm := strings.ToUpper(config.Algorithm)
	switch algorithm {
	case "HS256", "HS384", "HS512":
		if config.Secret == "" {
			return fmt.Errorf("JWT secret cannot be empty")
		}
	case "RS256", "RS384", "RS512":
		if config.PublicKey == "" {
			return fmt.Errorf("JWT public key cannot be empty for %s", config.Algorithm)
		}
	default:
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

	jwtConfig.Secret = configString(configMap, "secret")
	jwtConfig.Issuer = configString(configMap, "issuer")
	jwtConfig.Algorithm = configString(configMap, "algorithm")
	jwtConfig.PublicKey = configString(configMap, "public_key", "publicKey")
	if expiration, ok := configInt(configMap, "expiration"); ok {
		jwtConfig.Expiration = expiration
	}
	if verifyExpiration, ok := configBool(configMap, "verify_expiration", "verifyExpiration"); ok {
		jwtConfig.VerifyExpiration = verifyExpiration
	}
	if verifyIssuer, ok := configBool(configMap, "verify_issuer", "verifyIssuer"); ok {
		jwtConfig.VerifyIssuer = verifyIssuer
	}
	if refreshWindow, ok := configInt(configMap, "refresh_window", "refreshWindow"); ok {
		jwtConfig.RefreshWindow = refreshWindow
	}

	applyJWTConfigDefaults(jwtConfig, configMap)

	return jwtConfig, nil
}

// applyJWTConfigDefaults 为未显式配置的字段填充默认值
func applyJWTConfigDefaults(jwtConfig *JWTConfig, configMap map[string]interface{}) {
	if jwtConfig.Algorithm == "" {
		jwtConfig.Algorithm = DefaultJWTConfig.Algorithm
	}
	if jwtConfig.Expiration <= 0 {
		jwtConfig.Expiration = DefaultJWTConfig.Expiration
	}
	if _, ok := configBool(configMap, "verify_expiration", "verifyExpiration"); !ok {
		jwtConfig.VerifyExpiration = DefaultJWTConfig.VerifyExpiration
	}
	if _, ok := configBool(configMap, "verify_issuer", "verifyIssuer"); !ok {
		jwtConfig.VerifyIssuer = DefaultJWTConfig.VerifyIssuer
	}
	if jwtConfig.RefreshWindow <= 0 {
		jwtConfig.RefreshWindow = DefaultJWTConfig.RefreshWindow
	}
}
