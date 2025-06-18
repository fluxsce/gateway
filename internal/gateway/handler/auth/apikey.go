package auth

import (
	"errors"
	"fmt"
	"strings"

	"gohub/internal/gateway/core"
)

// APIKeyAuth API Key认证器
type APIKeyAuth struct {
	*BaseAuthenticator

	// API Key配置
	apiKeyConfig APIKeyConfig

	// 原始配置
	originalConfig AuthConfig
}

// NewAPIKeyAuth 创建API Key认证器
func NewAPIKeyAuth(apiKeyConfig APIKeyConfig, enabled bool, name string) *APIKeyAuth {
	base := NewBaseAuthenticator(StrategyAPIKey, enabled, name)

	// 如果名称为空，设置默认名称
	if name == "" {
		name = "API Key Authenticator"
		base.SetName(name)
	}

	auth := &APIKeyAuth{
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
				"keys":              apiKeyConfig.Keys,
				"is_prefix_match":   apiKeyConfig.IsPrefixMatch,
				"error_status_code": apiKeyConfig.ErrorStatusCode,
			},
		},
	}

	return auth
}

// APIKeyAuthFromConfig 从配置创建API Key认证器
func APIKeyAuthFromConfig(config AuthConfig) (Authenticator, error) {
	apiKeyConfig, err := parseAPIKeyConfigFromMap(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse API Key config: %w", err)
	}

	// 设置ID如果没有设置
	if apiKeyConfig.ID == "" {
		apiKeyConfig.ID = config.ID
	}

	apiKeyAuth := NewAPIKeyAuth(*apiKeyConfig, config.Enabled, config.Name)
	if config.Name != "" {
		apiKeyAuth.SetName(config.Name)
	}
	apiKeyAuth.originalConfig = config

	return apiKeyAuth, nil
}

// parseAPIKeyConfigFromMap 从map解析API Key配置
func parseAPIKeyConfigFromMap(configMap map[string]interface{}) (*APIKeyConfig, error) {
	if configMap == nil {
		return nil, fmt.Errorf("API Key配置不能为空")
	}

	apiKeyConfig := &APIKeyConfig{}

	if paramName, ok := configMap["param_name"].(string); ok {
		apiKeyConfig.ParamName = paramName
	}
	if location, ok := configMap["in"].(string); ok {
		apiKeyConfig.In = KeyLocation(location)
	}
	if isPrefixMatch, ok := configMap["is_prefix_match"].(bool); ok {
		apiKeyConfig.IsPrefixMatch = isPrefixMatch
	}

	// 解析keys列表
	if keys, ok := configMap["keys"].([]interface{}); ok {
		for _, keyItem := range keys {
			if keyMap, ok := keyItem.(map[string]interface{}); ok {
				apiKeyItem := APIKeyItem{}
				if name, ok := keyMap["name"].(string); ok {
					apiKeyItem.Name = name
				}
				if value, ok := keyMap["value"].(string); ok {
					apiKeyItem.Value = value
				}
				if roles, ok := keyMap["roles"].([]interface{}); ok {
					for _, role := range roles {
						if roleStr, ok := role.(string); ok {
							apiKeyItem.Roles = append(apiKeyItem.Roles, roleStr)
						}
					}
				}
				apiKeyConfig.Keys = append(apiKeyConfig.Keys, apiKeyItem)
			}
		}
	}

	return apiKeyConfig, nil
}

// GetStrategy 获取认证策略
func (a *APIKeyAuth) GetStrategy() AuthStrategy {
	return a.strategy
}

// IsEnabled 是否启用
func (a *APIKeyAuth) IsEnabled() bool {
	return a.enabled
}

// GetName 获取认证器名称
func (a *APIKeyAuth) GetName() string {
	return a.name
}

// SetName 设置认证器名称
func (a *APIKeyAuth) SetName(name string) {
	if name != "" {
		a.name = name
	}
}

// SetEnabled 设置是否启用
func (a *APIKeyAuth) SetEnabled(enabled bool) {
	a.enabled = enabled
}

// Handle 实现core.Handler接口
func (a *APIKeyAuth) Handle(ctx *core.Context) bool {
	// 获取API Key
	apiKey, err := a.extractAPIKey(ctx)
	if err != nil {
		a.handleError(ctx, err)
		return false
	}

	// 验证API Key
	keyItem, valid := a.validateAPIKey(apiKey)
	if !valid {
		a.handleError(ctx, errors.New("invalid API Key"))
		return false
	}

	// 存储认证信息到上下文
	a.storeAuthInfo(ctx, apiKey, keyItem)

	return true
}

// extractAPIKey 从请求中提取API Key
func (a *APIKeyAuth) extractAPIKey(ctx *core.Context) (string, error) {
	var apiKey string

	switch a.apiKeyConfig.In {
	case InHeader:
		// 从请求头提取
		apiKey = ctx.Request.Header.Get(a.apiKeyConfig.ParamName)
	case InQuery:
		// 从查询参数提取
		apiKey = ctx.Request.URL.Query().Get(a.apiKeyConfig.ParamName)
	case InCookie:
		// 从Cookie提取
		cookie, err := ctx.Request.Cookie(a.apiKeyConfig.ParamName)
		if err == nil {
			apiKey = cookie.Value
		}
	}

	if apiKey == "" {
		return "", fmt.Errorf("no API Key found in %s", a.apiKeyConfig.In)
	}

	return apiKey, nil
}

// validateAPIKey 验证API Key
func (a *APIKeyAuth) validateAPIKey(apiKey string) (*APIKeyItem, bool) {
	// 如果没有配置Key列表，则不验证
	if len(a.apiKeyConfig.Keys) == 0 {
		return &APIKeyItem{Value: apiKey}, true
	}

	// 前缀匹配
	if a.apiKeyConfig.IsPrefixMatch {
		for _, keyItem := range a.apiKeyConfig.Keys {
			if strings.HasPrefix(apiKey, keyItem.Value) {
				return &keyItem, true
			}
		}
		return nil, false
	}

	// 完全匹配
	for _, keyItem := range a.apiKeyConfig.Keys {
		if apiKey == keyItem.Value {
			return &keyItem, true
		}
	}
	return nil, false
}

// storeAuthInfo 存储认证信息到上下文
func (a *APIKeyAuth) storeAuthInfo(ctx *core.Context, apiKey string, keyItem *APIKeyItem) {
	ctx.Set("api_key", apiKey)
	ctx.Set("auth_method", "api-key")

	if keyItem != nil {
		ctx.Set("api_key_name", keyItem.Name)
		ctx.Set("api_key_roles", keyItem.Roles)
		ctx.Set("user_roles", keyItem.Roles) // 通用的用户角色

		// 如果有名称，也设置为用户ID
		if keyItem.Name != "" {
			ctx.Set("user_id", keyItem.Name)
		}
	}
}

// handleError 处理认证错误
func (a *APIKeyAuth) handleError(ctx *core.Context, err error) {
	ctx.AddError(fmt.Errorf("API Key authentication failed: %w", err))
	ctx.Abort(a.apiKeyConfig.ErrorStatusCode, map[string]string{
		"error": "Unauthorized: " + err.Error(),
	})
}

// GetConfig 获取API Key配置
func (a *APIKeyAuth) GetConfig() AuthConfig {
	return a.originalConfig
}

// Validate 验证API Key配置
func (a *APIKeyAuth) Validate() error {
	return ValidateAPIKeyConfig(&a.apiKeyConfig)
}

// ValidateAPIKeyConfig 验证API Key配置
func ValidateAPIKeyConfig(config *APIKeyConfig) error {
	if config == nil {
		return fmt.Errorf("API Key config cannot be nil")
	}

	if config.ParamName == "" {
		return fmt.Errorf("API Key param name cannot be empty")
	}

	// 验证位置是否有效
	validLocations := []KeyLocation{InHeader, InQuery, InCookie}
	locationValid := false
	for _, loc := range validLocations {
		if config.In == loc {
			locationValid = true
			break
		}
	}
	if !locationValid {
		return fmt.Errorf("invalid API Key location: %s", config.In)
	}

	// 验证错误状态码
	if config.ErrorStatusCode < 400 || config.ErrorStatusCode > 599 {
		return fmt.Errorf("invalid error status code: %d", config.ErrorStatusCode)
	}

	return nil
}
