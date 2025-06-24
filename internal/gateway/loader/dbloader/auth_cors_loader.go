package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gohub/internal/gateway/handler/auth"
	"gohub/internal/gateway/handler/cors"
	"gohub/pkg/database"
)

// AuthCORSConfigLoader 认证和CORS配置加载器
type AuthCORSConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewAuthCORSConfigLoader 创建认证和CORS配置加载器
func NewAuthCORSConfigLoader(db database.Database, tenantId string) *AuthCORSConfigLoader {
	return &AuthCORSConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadAuthConfig 加载实例级别认证配置
func (loader *AuthCORSConfigLoader) LoadAuthConfig(ctx context.Context, instanceId string) (*auth.AuthConfig, error) {
	query := `
		SELECT tenantId, authConfigId, authName, authType, authStrategy, authConfig,
		       exemptPaths, exemptHeaders, failureStatusCode, failureMessage
		FROM HUB_GATEWAY_AUTH_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record AuthConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询认证配置失败: %w", err)
	}

	// 构建认证配置
	authConf := &auth.AuthConfig{
		ID:      record.AuthConfigId,
		Name:    record.AuthName,
		Enabled: true,
	}

	// 解析认证策略
	switch record.AuthStrategy {
	case "REQUIRED":
		// 根据认证类型选择具体策略
		switch record.AuthType {
		case "JWT":
			authConf.Strategy = auth.StrategyJWT
		case "API_KEY":
			authConf.Strategy = auth.StrategyAPIKey
		case "OAUTH2":
			authConf.Strategy = auth.StrategyOAuth2
		case "BASIC":
			authConf.Strategy = auth.StrategyBasic
		default:
			authConf.Strategy = auth.StrategyJWT
		}
	case "OPTIONAL":
		authConf.Strategy = auth.StrategyJWT
	case "DISABLED":
		authConf.Strategy = auth.StrategyNoAuth
		authConf.Enabled = false
	default:
		authConf.Strategy = auth.StrategyNoAuth
	}

	// 解析豁免路径
	if record.ExemptPaths != nil {
		var paths []string
		if err := json.Unmarshal([]byte(*record.ExemptPaths), &paths); err == nil {
			authConf.ExcludedPaths = paths
		}
	}

	// 解析认证配置
	if record.AuthConfig != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.AuthConfig), &configMap); err == nil {
			authConf.Config = configMap
		}
	}

	return authConf, nil
}

// LoadRouteAuthConfig 加载路由级别认证配置
func (loader *AuthCORSConfigLoader) LoadRouteAuthConfig(ctx context.Context, routeId string) (*auth.AuthConfig, error) {
	query := `
		SELECT tenantId, authConfigId, authName, authType, authStrategy, authConfig,
		       exemptPaths, exemptHeaders, failureStatusCode, failureMessage
		FROM HUB_GATEWAY_AUTH_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record AuthConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询路由认证配置失败: %w", err)
	}

	// 构建认证配置
	authConf := &auth.AuthConfig{
		ID:      record.AuthConfigId,
		Name:    record.AuthName,
		Enabled: true,
	}

	// 解析认证策略
	switch record.AuthStrategy {
	case "REQUIRED":
		// 根据认证类型选择具体策略
		switch record.AuthType {
		case "JWT":
			authConf.Strategy = auth.StrategyJWT
		case "API_KEY":
			authConf.Strategy = auth.StrategyAPIKey
		case "OAUTH2":
			authConf.Strategy = auth.StrategyOAuth2
		case "BASIC":
			authConf.Strategy = auth.StrategyBasic
		default:
			authConf.Strategy = auth.StrategyJWT
		}
	case "OPTIONAL":
		authConf.Strategy = auth.StrategyJWT
	case "DISABLED":
		authConf.Strategy = auth.StrategyNoAuth
		authConf.Enabled = false
	default:
		authConf.Strategy = auth.StrategyNoAuth
	}

	// 解析豁免路径
	if record.ExemptPaths != nil {
		var paths []string
		if err := json.Unmarshal([]byte(*record.ExemptPaths), &paths); err == nil {
			authConf.ExcludedPaths = paths
		}
	}

	// 解析认证配置
	if record.AuthConfig != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(record.AuthConfig), &configMap); err == nil {
			authConf.Config = configMap
		}
	}

	return authConf, nil
}

// LoadCORSConfig 加载实例级别CORS配置
func (loader *AuthCORSConfigLoader) LoadCORSConfig(ctx context.Context, instanceId string) (*cors.CORSConfig, error) {
	query := `
		SELECT tenantId, corsConfigId, configName, allowOrigins, allowMethods,
		       allowHeaders, exposeHeaders, allowCredentials, maxAgeSeconds
		FROM HUB_GATEWAY_CORS_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record CORSConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询CORS配置失败: %w", err)
	}

	// 构建CORS配置
	corsConf := &cors.CORSConfig{
		ID:              record.CorsConfigId,
		Name:            record.ConfigName,
		Enabled:         true,
		AllowCredentials: record.AllowCredentials == "Y",
		MaxAge:           record.MaxAgeSeconds,
	}

	// 解析允许的源
	var origins []string
	if err := json.Unmarshal([]byte(record.AllowOrigins), &origins); err == nil {
		corsConf.AllowOrigins = origins
	}

	// 解析允许的方法
	corsConf.AllowMethods = strings.Split(record.AllowMethods, ",")

	// 解析允许的头
	if record.AllowHeaders != nil {
		var headers []string
		if err := json.Unmarshal([]byte(*record.AllowHeaders), &headers); err == nil {
			corsConf.AllowHeaders = headers
		}
	}

	// 解析暴露的头
	if record.ExposeHeaders != nil {
		var headers []string
		if err := json.Unmarshal([]byte(*record.ExposeHeaders), &headers); err == nil {
			corsConf.ExposeHeaders = headers
		}
	}

	return corsConf, nil
}

// LoadRouteCORSConfig 加载路由级别CORS配置
func (loader *AuthCORSConfigLoader) LoadRouteCORSConfig(ctx context.Context, routeId string) (*cors.CORSConfig, error) {
	query := `
		SELECT tenantId, corsConfigId, configName, allowOrigins, allowMethods,
		       allowHeaders, exposeHeaders, allowCredentials, maxAgeSeconds
		FROM HUB_GATEWAY_CORS_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record CORSConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询路由CORS配置失败: %w", err)
	}

	// 构建CORS配置
	corsConf := &cors.CORSConfig{
		ID:               record.CorsConfigId,
		Name:             record.ConfigName,
		Enabled:          true,
		AllowCredentials: record.AllowCredentials == "Y",
		MaxAge:           record.MaxAgeSeconds,
	}

	// 解析允许的源
	var origins []string
	if err := json.Unmarshal([]byte(record.AllowOrigins), &origins); err == nil {
		corsConf.AllowOrigins = origins
	}

	// 解析允许的方法
	corsConf.AllowMethods = strings.Split(record.AllowMethods, ",")

	// 解析允许的头
	if record.AllowHeaders != nil {
		var headers []string
		if err := json.Unmarshal([]byte(*record.AllowHeaders), &headers); err == nil {
			corsConf.AllowHeaders = headers
		}
	}

	// 解析暴露的头
	if record.ExposeHeaders != nil {
		var headers []string
		if err := json.Unmarshal([]byte(*record.ExposeHeaders), &headers); err == nil {
			corsConf.ExposeHeaders = headers
		}
	}

	return corsConf, nil
} 