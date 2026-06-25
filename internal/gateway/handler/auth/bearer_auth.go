package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"
)

// BearerTokenAuth Bearer Token 认证器。
type BearerTokenAuth struct {
	*BaseAuthenticator
	token          string
	errorStatus    int
	originalConfig AuthConfig
}

// BearerTokenAuthFromConfig 从配置创建 Bearer Token 认证器。
func BearerTokenAuthFromConfig(config AuthConfig) (Authenticator, error) {
	bearerConfig, err := parseBearerTokenConfig(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bearer token config: %w", err)
	}
	if bearerConfig.ID == "" {
		bearerConfig.ID = config.ID
	}
	if err := ValidateBearerTokenConfig(bearerConfig); err != nil {
		return nil, fmt.Errorf("invalid bearer token config: %w", err)
	}

	base := NewBaseAuthenticator(StrategyBearerToken, config.Enabled, config.Name)
	if config.Name != "" {
		base.SetName(config.Name)
	}

	return &BearerTokenAuth{
		BaseAuthenticator: base,
		token:             bearerConfig.Token,
		errorStatus:       bearerConfig.ErrorStatusCode,
		originalConfig:    config,
	}, nil
}

// GetConfig 获取认证器配置。
func (b *BearerTokenAuth) GetConfig() AuthConfig {
	return b.originalConfig
}

// Handle 执行 Bearer Token 认证。
func (b *BearerTokenAuth) Handle(ctx *core.Context) bool {
	if !b.enabled {
		return true
	}
	if ctx == nil || ctx.Request == nil {
		b.handleError(ctx, "invalid request context")
		return false
	}

	authHeader := strings.TrimSpace(ctx.Request.Header.Get("Authorization"))
	if authHeader == "" {
		b.handleError(ctx, "missing Authorization header")
		return false
	}
	if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
		b.handleError(ctx, "invalid Authorization header format")
		return false
	}

	providedToken := strings.TrimSpace(authHeader[len("Bearer "):])
	if providedToken == "" {
		b.handleError(ctx, "empty bearer token")
		return false
	}
	if !secureCompareAPIKey(providedToken, b.token) {
		b.handleError(ctx, "invalid bearer token")
		return false
	}

	ctx.Set("auth_method", "bearer-token")
	ctx.Set("bearer_token", providedToken)
	return true
}

// Validate 验证 Bearer Token 认证配置。
func (b *BearerTokenAuth) Validate() error {
	if strings.TrimSpace(b.token) == "" {
		return errors.New("bearer token cannot be empty")
	}
	return nil
}

func (b *BearerTokenAuth) handleError(ctx *core.Context, message string) {
	if ctx == nil {
		return
	}
	ctx.AddError(fmt.Errorf("bearer token authentication failed: %s", message))
	status := b.errorStatus
	if status < http.StatusBadRequest || status > 599 {
		status = http.StatusUnauthorized
	}
	ctx.Abort(status, map[string]string{
		"error": "Unauthorized: " + message,
	})
}

func parseBearerTokenConfig(configMap map[string]interface{}) (*BearerTokenConfig, error) {
	if configMap == nil {
		return nil, errors.New("bearer token config cannot be nil")
	}
	cfg := &BearerTokenConfig{
		Token:           strings.TrimSpace(configString(configMap, "token")),
		ErrorStatusCode: DefaultBearerTokenConfig.ErrorStatusCode,
	}
	if code, ok := configInt(configMap, "error_status_code"); ok {
		cfg.ErrorStatusCode = code
	}
	return cfg, nil
}

// ValidateBearerTokenConfig 校验 Bearer Token 配置。
func ValidateBearerTokenConfig(config *BearerTokenConfig) error {
	if config == nil {
		return errors.New("bearer token config cannot be nil")
	}
	if strings.TrimSpace(config.Token) == "" {
		return errors.New("bearer token cannot be empty")
	}
	if config.ErrorStatusCode != 0 && (config.ErrorStatusCode < 400 || config.ErrorStatusCode > 599) {
		return fmt.Errorf("invalid error_status_code: %d", config.ErrorStatusCode)
	}
	return nil
}
