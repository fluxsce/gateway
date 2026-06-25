package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"gateway/internal/gateway/core"
)

// APIKeyAuth API Key 认证器
type APIKeyAuth struct {
	*BaseAuthenticator
	apiKeyConfig   APIKeyConfig
	originalConfig AuthConfig
}

// NewAPIKeyAuth 创建 API Key 认证器
func NewAPIKeyAuth(apiKeyConfig APIKeyConfig, enabled bool, name string) *APIKeyAuth {
	base := NewBaseAuthenticator(StrategyAPIKey, enabled, name)
	if name == "" {
		name = "API Key Authenticator"
		base.SetName(name)
	}

	return &APIKeyAuth{
		BaseAuthenticator: base,
		apiKeyConfig:      apiKeyConfig,
		originalConfig: AuthConfig{
			ID:       apiKeyConfig.ID,
			Strategy: StrategyAPIKey,
			Name:     name,
			Enabled:  enabled,
			Config: map[string]interface{}{
				"param_name":        apiKeyConfig.ParamName,
				"in":                string(apiKeyConfig.In),
				"key":               apiKeyConfig.Key,
				"error_status_code": apiKeyConfig.ErrorStatusCode,
			},
		},
	}
}

// APIKeyAuthFromConfig 从配置创建 API Key 认证器
func APIKeyAuthFromConfig(config AuthConfig) (Authenticator, error) {
	apiKeyConfig, err := parseAPIKeyConfigFromMap(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API Key config: %w", err)
	}
	if apiKeyConfig.ID == "" {
		apiKeyConfig.ID = config.ID
	}
	if err := ValidateAPIKeyConfig(apiKeyConfig); err != nil {
		return nil, fmt.Errorf("invalid API Key config: %w", err)
	}

	apiKeyAuth := NewAPIKeyAuth(*apiKeyConfig, config.Enabled, config.Name)
	if config.Name != "" {
		apiKeyAuth.SetName(config.Name)
	}
	apiKeyAuth.originalConfig = config
	return apiKeyAuth, nil
}

// parseAPIKeyConfigFromMap 从 map 解析 APIKeyConfig
func parseAPIKeyConfigFromMap(configMap map[string]interface{}) (*APIKeyConfig, error) {
	if configMap == nil {
		return nil, fmt.Errorf("API Key配置不能为空")
	}

	cfg := &APIKeyConfig{
		ParamName:       strings.TrimSpace(configString(configMap, "param_name")),
		In:              KeyLocation(strings.ToLower(configString(configMap, "in"))),
		Key:             strings.TrimSpace(configString(configMap, "key")),
		ErrorStatusCode: DefaultAPIKeyConfig.ErrorStatusCode,
	}
	if cfg.ParamName == "" {
		cfg.ParamName = DefaultAPIKeyConfig.ParamName
	}
	if cfg.In == "" {
		cfg.In = DefaultAPIKeyConfig.In
	}
	if code, ok := configInt(configMap, "error_status_code"); ok {
		cfg.ErrorStatusCode = code
	}
	return cfg, nil
}

func (a *APIKeyAuth) GetStrategy() AuthStrategy { return a.strategy }
func (a *APIKeyAuth) IsEnabled() bool           { return a.enabled }
func (a *APIKeyAuth) GetName() string           { return a.name }
func (a *APIKeyAuth) SetName(name string) {
	if name != "" {
		a.name = name
	}
}
func (a *APIKeyAuth) SetEnabled(enabled bool) { a.enabled = enabled }
func (a *APIKeyAuth) GetConfig() AuthConfig   { return a.originalConfig }
func (a *APIKeyAuth) Validate() error         { return ValidateAPIKeyConfig(&a.apiKeyConfig) }

// Handle 处理 API Key 认证
func (a *APIKeyAuth) Handle(ctx *core.Context) bool {
	if !a.enabled {
		return true
	}

	apiKey, err := a.extractAPIKey(ctx)
	if err != nil {
		a.handleError(ctx, err)
		return false
	}
	if err := a.validateAPIKey(apiKey); err != nil {
		a.handleError(ctx, err)
		return false
	}

	a.storeAuthInfo(ctx, apiKey)
	return true
}

// extractAPIKey 从请求中提取 API Key 值
func (a *APIKeyAuth) extractAPIKey(ctx *core.Context) (string, error) {
	if ctx == nil || ctx.Request == nil {
		return "", errors.New("invalid request context")
	}

	paramName := strings.TrimSpace(a.apiKeyConfig.ParamName)
	if paramName == "" {
		return "", errors.New("API Key param_name is not configured")
	}

	var apiKey string
	switch a.apiKeyConfig.In {
	case InHeader:
		apiKey = ctx.Request.Header.Get(paramName)
	case InQuery:
		apiKey = ctx.Request.URL.Query().Get(paramName)
	case InCookie:
		if cookie, err := ctx.Request.Cookie(paramName); err == nil && cookie != nil {
			apiKey = cookie.Value
		}
	default:
		return "", fmt.Errorf("unsupported API Key location: %s", a.apiKeyConfig.In)
	}

	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return "", fmt.Errorf("missing API Key value in %s (param_name: %s)", a.apiKeyConfig.In, paramName)
	}
	return apiKey, nil
}

// validateAPIKey 校验请求中的 API Key 值是否与配置一致
func (a *APIKeyAuth) validateAPIKey(apiKey string) error {
	expected := strings.TrimSpace(a.apiKeyConfig.Key)
	if expected == "" {
		return errors.New("API Key value is not configured")
	}
	if !secureCompareAPIKey(apiKey, expected) {
		return errors.New("invalid API Key value")
	}
	return nil
}

// secureCompareAPIKey 恒定时间比较，降低时序攻击风险
func secureCompareAPIKey(provided, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// storeAuthInfo 将认证结果写入上下文
func (a *APIKeyAuth) storeAuthInfo(ctx *core.Context, apiKey string) {
	ctx.Set("api_key", apiKey)
	ctx.Set("auth_method", "api-key")
}

// handleError 处理认证失败
func (a *APIKeyAuth) handleError(ctx *core.Context, err error) {
	ctx.AddError(fmt.Errorf("API Key authentication failed: %w", err))
	statusCode := a.apiKeyConfig.ErrorStatusCode
	if statusCode < http.StatusBadRequest || statusCode > 599 {
		statusCode = http.StatusUnauthorized
	}
	ctx.Abort(statusCode, map[string]string{
		"error": "Unauthorized: " + err.Error(),
	})
}

// ValidateAPIKeyConfig 校验 APIKeyConfig
func ValidateAPIKeyConfig(config *APIKeyConfig) error {
	if config == nil {
		return fmt.Errorf("API Key config cannot be nil")
	}
	if strings.TrimSpace(config.ParamName) == "" {
		return fmt.Errorf("API Key param_name cannot be empty")
	}

	switch config.In {
	case InHeader, InQuery, InCookie:
	default:
		return fmt.Errorf("invalid API Key in: %s", config.In)
	}

	if strings.TrimSpace(config.Key) == "" {
		return fmt.Errorf("API Key value cannot be empty")
	}

	if config.ErrorStatusCode != 0 && (config.ErrorStatusCode < 400 || config.ErrorStatusCode > 599) {
		return fmt.Errorf("invalid error_status_code: %d", config.ErrorStatusCode)
	}
	return nil
}
