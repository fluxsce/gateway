package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gateway/internal/gateway/handler/auth"
	"gateway/internal/gateway/handler/cors"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
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
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, authConfigId, authName, authType, authStrategy, authConfig,
		       exemptPaths, exemptHeaders, failureStatusCode, failureMessage
		FROM HUB_GW_AUTH_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, instanceId}, paginationArgs...)

	// 执行查询
	var records []AuthConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询认证配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

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
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, authConfigId, authName, authType, authStrategy, authConfig,
		       exemptPaths, exemptHeaders, failureStatusCode, failureMessage
		FROM HUB_GW_AUTH_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, routeId}, paginationArgs...)

	// 执行查询
	var records []AuthConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由认证配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

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
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, corsConfigId, configName, allowOrigins, allowMethods,
		       allowHeaders, exposeHeaders, allowCredentials, maxAgeSeconds
		FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, instanceId}, paginationArgs...)

	// 执行查询
	var records []CORSConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询CORS配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

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
	corsConf.AllowMethods = parseArrayString(record.AllowMethods)

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
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, corsConfigId, configName, allowOrigins, allowMethods,
		       allowHeaders, exposeHeaders, allowCredentials, maxAgeSeconds
		FROM HUB_GW_CORS_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
	`

	// 创建分页信息（只取第一条记录）
	pagination := sqlutils.NewPaginationInfo(1, 1)

	// 获取数据库类型
	dbType := sqlutils.GetDatabaseType(loader.db)

	// 构建分页查询
	paginatedQuery, paginationArgs, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
	if err != nil {
		return nil, fmt.Errorf("构建分页查询失败: %w", err)
	}

	// 合并查询参数
	allArgs := append([]interface{}{loader.tenantId, routeId}, paginationArgs...)

	// 执行查询
	var records []CORSConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由CORS配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

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
	corsConf.AllowMethods = parseArrayString(record.AllowMethods)

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

// parseArrayString 解析逗号分隔的字符串或JSON数组，处理空白字符和空字符串
// 功能特性：
// - 优先尝试解析JSON数组格式（如 ["a","b","c"]）
// - 如果JSON解析失败，则按逗号分割字符串
// - 去除每个元素的前后空白字符
// - 过滤掉空字符串元素
// - 返回清理后的字符串切片
//
// 参数:
//
//	str: 要解析的字符串（JSON数组或逗号分隔字符串）
//
// 返回:
//
//	[]string: 解析后的字符串切片
//
// 示例:
//
//	parseArrayString(`["a","b","c"]`)
//	// 返回: ["a", "b", "c"]
//	parseArrayString("a, b , c,, d ")
//	// 返回: ["a", "b", "c", "d"]
func parseArrayString(str string) []string {
	if str == "" {
		return []string{}
	}

	// 去除前后空白字符
	str = strings.TrimSpace(str)

	// 优先尝试解析JSON数组
	if strings.HasPrefix(str, "[") && strings.HasSuffix(str, "]") {
		var jsonArray []string
		if err := json.Unmarshal([]byte(str), &jsonArray); err == nil {
			// JSON解析成功，过滤空字符串并去除空白字符
			var result []string
			for _, item := range jsonArray {
				trimmed := strings.TrimSpace(item)
				if trimmed != "" {
					result = append(result, trimmed)
				}
			}
			return result
		}
	}

	// JSON解析失败，按逗号分割
	parts := strings.Split(str, ",")

	// 清理和过滤
	var result []string
	for _, part := range parts {
		// 去除前后空白字符
		trimmed := strings.TrimSpace(part)
		// 过滤掉空字符串
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}
