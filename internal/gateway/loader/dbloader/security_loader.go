package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gohub/internal/gateway/handler/security"
	"gohub/pkg/database"
	"gohub/pkg/database/sqlutils"
	"gohub/pkg/logger"
)

// SecurityConfigLoader 安全配置加载器
type SecurityConfigLoader struct {
	db       database.Database
	tenantId string
}

// NewSecurityConfigLoader 创建安全配置加载器
func NewSecurityConfigLoader(db database.Database, tenantId string) *SecurityConfigLoader {
	return &SecurityConfigLoader{
		db:       db,
		tenantId: tenantId,
	}
}

// LoadSecurityConfig 加载实例级别安全配置
func (loader *SecurityConfigLoader) LoadSecurityConfig(ctx context.Context, instanceId string) (*security.SecurityConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, securityConfigId, gatewayInstanceId, routeConfigId, configName,
		       configDesc, configPriority, customConfigJson, activeFlag
		FROM HUB_GW_SECURITY_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? 
		AND (routeConfigId IS NULL OR routeConfigId = '') 
		AND activeFlag = 'Y'
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
	var records []SecurityConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询安全配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建安全配置
	securityConfig := &security.SecurityConfig{
		ID:      record.SecurityConfigId,
		Name:    record.ConfigName,
		Enabled: record.ActiveFlag == "Y",
	}

	// 加载IP访问控制配置
	ipAccess, err := loader.LoadIPAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载IP访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if ipAccess != nil {
		securityConfig.IPAccess = *ipAccess
	}

	// 加载User-Agent访问控制配置
	userAgentAccess, err := loader.LoadUserAgentAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载User-Agent访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if userAgentAccess != nil {
		securityConfig.UserAgentAccess = *userAgentAccess
	}

	// 加载API访问控制配置
	apiAccess, err := loader.LoadAPIAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载API访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if apiAccess != nil {
		securityConfig.APIAccess = *apiAccess
	}

	// 加载域名访问控制配置
	domainAccess, err := loader.LoadDomainAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载域名访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if domainAccess != nil {
		securityConfig.DomainAccess = *domainAccess
	}

	// 解析自定义配置
	if record.CustomConfigJson != nil {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(*record.CustomConfigJson), &customConfig); err == nil {
			securityConfig.CustomConfig = customConfig
		}
	}

	return securityConfig, nil
}

// LoadRouteSecurityConfig 加载路由级别安全配置
func (loader *SecurityConfigLoader) LoadRouteSecurityConfig(ctx context.Context, routeId string) (*security.SecurityConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, securityConfigId, gatewayInstanceId, routeConfigId, configName,
		       configDesc, configPriority, customConfigJson, activeFlag
		FROM HUB_GW_SECURITY_CONFIG 
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
	var records []SecurityConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由安全配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建安全配置
	securityConfig := &security.SecurityConfig{
		ID:      record.SecurityConfigId,
		Name:    record.ConfigName,
		Enabled: record.ActiveFlag == "Y",
	}

	// 加载IP访问控制配置
	ipAccess, err := loader.LoadIPAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载IP访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if ipAccess != nil {
		securityConfig.IPAccess = *ipAccess
	}

	// 加载User-Agent访问控制配置
	userAgentAccess, err := loader.LoadUserAgentAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载User-Agent访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if userAgentAccess != nil {
		securityConfig.UserAgentAccess = *userAgentAccess
	}

	// 加载API访问控制配置
	apiAccess, err := loader.LoadAPIAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载API访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if apiAccess != nil {
		securityConfig.APIAccess = *apiAccess
	}

	// 加载域名访问控制配置
	domainAccess, err := loader.LoadDomainAccessConfig(ctx, record.SecurityConfigId)
	if err != nil {
		logger.Warn("加载域名访问控制配置失败", "securityConfigId", record.SecurityConfigId, "error", err)
	} else if domainAccess != nil {
		securityConfig.DomainAccess = *domainAccess
	}

	// 解析自定义配置
	if record.CustomConfigJson != nil {
		var customConfig map[string]interface{}
		if err := json.Unmarshal([]byte(*record.CustomConfigJson), &customConfig); err == nil {
			securityConfig.CustomConfig = customConfig
		}
	}

	return securityConfig, nil
}

// LoadIPAccessConfig 加载IP访问控制配置
func (loader *SecurityConfigLoader) LoadIPAccessConfig(ctx context.Context, securityConfigId string) (*security.IPAccessConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, ipAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistIps, blacklistIps, whitelistCidrs, blacklistCidrs,
		       trustXForwardedFor, trustXRealIp, activeFlag
		FROM HUB_GW_IP_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
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
	allArgs := append([]interface{}{loader.tenantId, securityConfigId}, paginationArgs...)

	// 执行查询
	var records []IPAccessConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询IP访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建IP访问控制配置
	ipAccessConfig := &security.IPAccessConfig{
		ID:                 record.IpAccessConfigId,
		Name:               record.ConfigName,
		Enabled:            record.ActiveFlag == "Y",
		DefaultPolicy:      record.DefaultPolicy,
		TrustXForwardedFor: record.TrustXForwardedFor == "Y",
		TrustXRealIP:       record.TrustXRealIp == "Y",
	}

	// 解析白名单和黑名单
	if record.WhitelistIps != nil && *record.WhitelistIps != "" {
		ipAccessConfig.Whitelist = parseCommaSeparatedString(*record.WhitelistIps)
	}

	if record.BlacklistIps != nil && *record.BlacklistIps != "" {
		ipAccessConfig.Blacklist = parseCommaSeparatedString(*record.BlacklistIps)
	}

	if record.WhitelistCidrs != nil && *record.WhitelistCidrs != "" {
		ipAccessConfig.WhitelistCIDR = parseCommaSeparatedString(*record.WhitelistCidrs)
	}

	if record.BlacklistCidrs != nil && *record.BlacklistCidrs != "" {
		ipAccessConfig.BlacklistCIDR = parseCommaSeparatedString(*record.BlacklistCidrs)
	}

	return ipAccessConfig, nil
}

// LoadUserAgentAccessConfig 加载User-Agent访问控制配置
func (loader *SecurityConfigLoader) LoadUserAgentAccessConfig(ctx context.Context, securityConfigId string) (*security.UserAgentAccessConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, useragentAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPatterns, blacklistPatterns, blockEmptyUserAgent, activeFlag
		FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
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
	allArgs := append([]interface{}{loader.tenantId, securityConfigId}, paginationArgs...)

	// 执行查询
	var records []UserAgentAccessConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询User-Agent访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建User-Agent访问控制配置
	userAgentConfig := &security.UserAgentAccessConfig{
		ID:            record.UseragentAccessConfigId,
		Name:          record.ConfigName,
		Enabled:       record.ActiveFlag == "Y",
		DefaultPolicy: record.DefaultPolicy,
		BlockEmpty:    record.BlockEmptyUserAgent == "Y",
	}

	// 解析白名单和黑名单
	if record.WhitelistPatterns != nil && *record.WhitelistPatterns != "" {
		userAgentConfig.Whitelist = parseCommaSeparatedString(*record.WhitelistPatterns)
	}

	if record.BlacklistPatterns != nil && *record.BlacklistPatterns != "" {
		userAgentConfig.Blacklist = parseCommaSeparatedString(*record.BlacklistPatterns)
	}

	return userAgentConfig, nil
}

// LoadAPIAccessConfig 加载API访问控制配置
func (loader *SecurityConfigLoader) LoadAPIAccessConfig(ctx context.Context, securityConfigId string) (*security.APIAccessConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, apiAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPaths, blacklistPaths, allowedMethods, blockedMethods, activeFlag
		FROM HUB_GW_API_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
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
	allArgs := append([]interface{}{loader.tenantId, securityConfigId}, paginationArgs...)

	// 执行查询
	var records []APIAccessConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询API访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建API访问控制配置
	apiAccessConfig := &security.APIAccessConfig{
		ID:            record.ApiAccessConfigId,
		Name:          record.ConfigName,
		Enabled:       record.ActiveFlag == "Y",
		DefaultPolicy: record.DefaultPolicy,
	}

	// 解析白名单和黑名单
	if record.WhitelistPaths != nil && *record.WhitelistPaths != "" {
		apiAccessConfig.Whitelist = parseCommaSeparatedString(*record.WhitelistPaths)
	}

	if record.BlacklistPaths != nil && *record.BlacklistPaths != "" {
		apiAccessConfig.Blacklist = parseCommaSeparatedString(*record.BlacklistPaths)
	}

	// 解析允许和阻止的HTTP方法
	if record.AllowedMethods != "" {
		apiAccessConfig.AllowedMethods = parseCommaSeparatedString(record.AllowedMethods)
	}

	if record.BlockedMethods != nil && *record.BlockedMethods != "" {
		apiAccessConfig.BlockedMethods = parseCommaSeparatedString(*record.BlockedMethods)
	}

	return apiAccessConfig, nil
}

// LoadDomainAccessConfig 加载域名访问控制配置
func (loader *SecurityConfigLoader) LoadDomainAccessConfig(ctx context.Context, securityConfigId string) (*security.DomainAccessConfig, error) {
	// 构建基础查询语句
	baseQuery := `
		SELECT tenantId, domainAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistDomains, blacklistDomains, allowSubdomains, activeFlag
		FROM HUB_GW_DOMAIN_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
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
	allArgs := append([]interface{}{loader.tenantId, securityConfigId}, paginationArgs...)

	// 执行查询
	var records []DomainAccessConfigRecord
	err = loader.db.Query(ctx, &records, paginatedQuery, allArgs, true)
	if err != nil {
		return nil, fmt.Errorf("查询域名访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	record := records[0]

	// 构建域名访问控制配置
	domainAccessConfig := &security.DomainAccessConfig{
		ID:              record.DomainAccessConfigId,
		Name:            record.ConfigName,
		Enabled:         record.ActiveFlag == "Y",
		DefaultPolicy:   record.DefaultPolicy,
		AllowSubdomains: record.AllowSubdomains == "Y",
	}

	// 解析白名单和黑名单
	if record.WhitelistDomains != nil && *record.WhitelistDomains != "" {
		domainAccessConfig.Whitelist = parseCommaSeparatedString(*record.WhitelistDomains)
	}

	if record.BlacklistDomains != nil && *record.BlacklistDomains != "" {
		domainAccessConfig.Blacklist = parseCommaSeparatedString(*record.BlacklistDomains)
	}

	return domainAccessConfig, nil
}

// parseCommaSeparatedString 解析逗号分隔的字符串或JSON数组，处理空白字符和空字符串
// 功能特性：
// - 优先尝试解析JSON数组格式（如 ["a","b","c"]）
// - 如果JSON解析失败，则按逗号分割字符串
// - 去除每个元素的前后空白字符
// - 过滤掉空字符串元素
// - 返回清理后的字符串切片
//
// 参数:
//   str: 要解析的字符串（JSON数组或逗号分隔字符串）
// 返回:
//   []string: 解析后的字符串切片
//
// 示例:
//   parseCommaSeparatedString(`["a","b","c"]`) 
//   // 返回: ["a", "b", "c"]
//   parseCommaSeparatedString("a, b , c,, d ") 
//   // 返回: ["a", "b", "c", "d"]
func parseCommaSeparatedString(str string) []string {
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