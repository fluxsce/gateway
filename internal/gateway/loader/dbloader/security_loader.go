package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gateway/internal/gateway/handler/security"
	"gateway/pkg/database"
	"gateway/pkg/logger"
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

// LoadSecurityConfigByInstanceId 加载实例级别安全配置（通过实例ID直接查询各访问控制配置表）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadSecurityConfigByInstanceId(ctx context.Context, instanceId string) (*security.SecurityConfig, error) {
	// 构建安全配置（合并所有访问控制配置）
	securityConfig := &security.SecurityConfig{
		ID:      instanceId, // 使用实例ID作为安全配置ID
		Name:    "Instance Security Config",
		Enabled: true, // 如果存在任何配置，则认为启用
	}

	// 加载并合并IP访问控制配置
	ipAccess, err := loader.LoadIPAccessConfigByInstanceId(ctx, instanceId)
	if err != nil {
		logger.Warn("加载IP访问控制配置失败", "instanceId", instanceId, "error", err)
	} else if ipAccess != nil {
		securityConfig.IPAccess = *ipAccess
	}

	// 加载并合并User-Agent访问控制配置
	userAgentAccess, err := loader.LoadUserAgentAccessConfigByInstanceId(ctx, instanceId)
	if err != nil {
		logger.Warn("加载User-Agent访问控制配置失败", "instanceId", instanceId, "error", err)
	} else if userAgentAccess != nil {
		securityConfig.UserAgentAccess = *userAgentAccess
	}

	// 加载并合并API访问控制配置
	apiAccess, err := loader.LoadAPIAccessConfigByInstanceId(ctx, instanceId)
	if err != nil {
		logger.Warn("加载API访问控制配置失败", "instanceId", instanceId, "error", err)
	} else if apiAccess != nil {
		securityConfig.APIAccess = *apiAccess
	}

	// 加载并合并域名访问控制配置
	domainAccess, err := loader.LoadDomainAccessConfigByInstanceId(ctx, instanceId)
	if err != nil {
		logger.Warn("加载域名访问控制配置失败", "instanceId", instanceId, "error", err)
	} else if domainAccess != nil {
		securityConfig.DomainAccess = *domainAccess
	}

	return securityConfig, nil
}

// LoadRouteSecurityConfig 加载路由级别安全配置（通过路由ID直接查询各访问控制配置表）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadRouteSecurityConfig(ctx context.Context, routeId string) (*security.SecurityConfig, error) {
	// 构建安全配置（合并所有访问控制配置）
	securityConfig := &security.SecurityConfig{
		ID:      routeId, // 使用路由ID作为安全配置ID
		Name:    "Route Security Config",
		Enabled: true, // 如果存在任何配置，则认为启用
	}

	// 加载并合并IP访问控制配置（路由级别通过routeConfigId关联）
	ipAccess, err := loader.LoadIPAccessConfigByRouteId(ctx, routeId)
	if err != nil {
		logger.Warn("加载路由IP访问控制配置失败", "routeId", routeId, "error", err)
	} else if ipAccess != nil {
		securityConfig.IPAccess = *ipAccess
	}

	// 加载并合并User-Agent访问控制配置
	userAgentAccess, err := loader.LoadUserAgentAccessConfigByRouteId(ctx, routeId)
	if err != nil {
		logger.Warn("加载路由User-Agent访问控制配置失败", "routeId", routeId, "error", err)
	} else if userAgentAccess != nil {
		securityConfig.UserAgentAccess = *userAgentAccess
	}

	// 加载并合并API访问控制配置
	apiAccess, err := loader.LoadAPIAccessConfigByRouteId(ctx, routeId)
	if err != nil {
		logger.Warn("加载路由API访问控制配置失败", "routeId", routeId, "error", err)
	} else if apiAccess != nil {
		securityConfig.APIAccess = *apiAccess
	}

	// 加载并合并域名访问控制配置
	domainAccess, err := loader.LoadDomainAccessConfigByRouteId(ctx, routeId)
	if err != nil {
		logger.Warn("加载路由域名访问控制配置失败", "routeId", routeId, "error", err)
	} else if domainAccess != nil {
		securityConfig.DomainAccess = *domainAccess
	}

	return securityConfig, nil
}

// LoadIPAccessConfigByInstanceId 加载IP访问控制配置（通过实例ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadIPAccessConfigByInstanceId(ctx context.Context, instanceId string) (*security.IPAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是实例ID
	query := `
		SELECT tenantId, ipAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistIps, blacklistIps, whitelistCidrs, blacklistCidrs,
		       trustXForwardedFor, trustXRealIp, activeFlag
		FROM HUB_GW_IP_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY ipAccessConfigId ASC
	`

	// 执行查询
	var records []IPAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询IP访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeIPAccessConfigs(records), nil
}

// LoadIPAccessConfigByRouteId 加载IP访问控制配置（通过路由ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadIPAccessConfigByRouteId(ctx context.Context, routeId string) (*security.IPAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是路由ID
	query := `
		SELECT tenantId, ipAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistIps, blacklistIps, whitelistCidrs, blacklistCidrs,
		       trustXForwardedFor, trustXRealIp, activeFlag
		FROM HUB_GW_IP_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY ipAccessConfigId ASC
	`

	// 执行查询
	var records []IPAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由IP访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeIPAccessConfigs(records), nil
}

// LoadUserAgentAccessConfigByRouteId 加载User-Agent访问控制配置（通过路由ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadUserAgentAccessConfigByRouteId(ctx context.Context, routeId string) (*security.UserAgentAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是路由ID
	query := `
		SELECT tenantId, useragentAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPatterns, blacklistPatterns, blockEmptyUserAgent, activeFlag
		FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY useragentAccessConfigId ASC
	`

	// 执行查询
	var records []UserAgentAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由User-Agent访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeUserAgentAccessConfigs(records), nil
}

// LoadAPIAccessConfigByRouteId 加载API访问控制配置（通过路由ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadAPIAccessConfigByRouteId(ctx context.Context, routeId string) (*security.APIAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是路由ID
	query := `
		SELECT tenantId, apiAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPaths, blacklistPaths, allowedMethods, blockedMethods, activeFlag
		FROM HUB_GW_API_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY apiAccessConfigId ASC
	`

	// 执行查询
	var records []APIAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由API访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeAPIAccessConfigs(records), nil
}

// LoadDomainAccessConfigByRouteId 加载域名访问控制配置（通过路由ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadDomainAccessConfigByRouteId(ctx context.Context, routeId string) (*security.DomainAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是路由ID
	query := `
		SELECT tenantId, domainAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistDomains, blacklistDomains, allowSubdomains, activeFlag
		FROM HUB_GW_DOMAIN_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY domainAccessConfigId ASC
	`

	// 执行查询
	var records []DomainAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询路由域名访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeDomainAccessConfigs(records), nil
}

// LoadUserAgentAccessConfigByInstanceId 加载User-Agent访问控制配置（通过实例ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadUserAgentAccessConfigByInstanceId(ctx context.Context, instanceId string) (*security.UserAgentAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是实例ID
	query := `
		SELECT tenantId, useragentAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPatterns, blacklistPatterns, blockEmptyUserAgent, activeFlag
		FROM HUB_GW_UA_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY useragentAccessConfigId ASC
	`

	// 执行查询
	var records []UserAgentAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询User-Agent访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeUserAgentAccessConfigs(records), nil
}

// LoadAPIAccessConfigByInstanceId 加载API访问控制配置（通过实例ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadAPIAccessConfigByInstanceId(ctx context.Context, instanceId string) (*security.APIAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是实例ID
	query := `
		SELECT tenantId, apiAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPaths, blacklistPaths, allowedMethods, blockedMethods, activeFlag
		FROM HUB_GW_API_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY apiAccessConfigId ASC
	`

	// 执行查询
	var records []APIAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询API访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeAPIAccessConfigs(records), nil
}

// LoadDomainAccessConfigByInstanceId 加载域名访问控制配置（通过实例ID，使用SecurityConfigId字段关联）
// 全量查询所有配置并合并，不分页
func (loader *SecurityConfigLoader) LoadDomainAccessConfigByInstanceId(ctx context.Context, instanceId string) (*security.DomainAccessConfig, error) {
	// 构建查询语句（全量查询，不分页）
	// SecurityConfigId 字段的值就是实例ID
	query := `
		SELECT tenantId, domainAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistDomains, blacklistDomains, allowSubdomains, activeFlag
		FROM HUB_GW_DOMAIN_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		ORDER BY domainAccessConfigId ASC
	`

	// 执行查询
	var records []DomainAccessConfigRecord
	err := loader.db.Query(ctx, &records, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		return nil, fmt.Errorf("查询域名访问控制配置失败: %w", err)
	}

	// 如果没有找到记录，返回nil
	if len(records) == 0 {
		return nil, nil
	}

	// 合并所有配置
	return mergeDomainAccessConfigs(records), nil
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
//
//	str: 要解析的字符串（JSON数组或逗号分隔字符串）
//
// 返回:
//
//	[]string: 解析后的字符串切片
//
// 示例:
//
//	parseCommaSeparatedString(`["a","b","c"]`)
//	// 返回: ["a", "b", "c"]
//	parseCommaSeparatedString("a, b , c,, d ")
//	// 返回: ["a", "b", "c", "d"]
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

// mergeIPAccessConfigs 合并多个IP访问控制配置
func mergeIPAccessConfigs(records []IPAccessConfigRecord) *security.IPAccessConfig {
	if len(records) == 0 {
		return nil
	}

	// 使用第一个记录作为基础
	first := records[0]
	merged := &security.IPAccessConfig{
		ID:                 first.IpAccessConfigId,
		Name:               first.ConfigName,
		Enabled:            first.ActiveFlag == "Y",
		DefaultPolicy:      first.DefaultPolicy,
		TrustXForwardedFor: first.TrustXForwardedFor == "Y",
		TrustXRealIP:       first.TrustXRealIp == "Y",
		Whitelist:          []string{},
		Blacklist:          []string{},
		WhitelistCIDR:      []string{},
		BlacklistCIDR:      []string{},
	}

	// 使用 map 去重
	whitelistMap := make(map[string]bool)
	blacklistMap := make(map[string]bool)
	whitelistCIDRMap := make(map[string]bool)
	blacklistCIDRMap := make(map[string]bool)

	// 合并所有记录
	for _, record := range records {
		// 合并 Enabled：任何一个启用则启用
		if record.ActiveFlag == "Y" {
			merged.Enabled = true
		}

		// 合并 TrustXForwardedFor：任何一个启用则启用
		if record.TrustXForwardedFor == "Y" {
			merged.TrustXForwardedFor = true
		}

		// 合并 TrustXRealIP：任何一个启用则启用
		if record.TrustXRealIp == "Y" {
			merged.TrustXRealIP = true
		}

		// 合并白名单和黑名单
		if record.WhitelistIps != nil && *record.WhitelistIps != "" {
			for _, ip := range parseCommaSeparatedString(*record.WhitelistIps) {
				whitelistMap[ip] = true
			}
		}

		if record.BlacklistIps != nil && *record.BlacklistIps != "" {
			for _, ip := range parseCommaSeparatedString(*record.BlacklistIps) {
				blacklistMap[ip] = true
			}
		}

		if record.WhitelistCidrs != nil && *record.WhitelistCidrs != "" {
			for _, cidr := range parseCommaSeparatedString(*record.WhitelistCidrs) {
				whitelistCIDRMap[cidr] = true
			}
		}

		if record.BlacklistCidrs != nil && *record.BlacklistCidrs != "" {
			for _, cidr := range parseCommaSeparatedString(*record.BlacklistCidrs) {
				blacklistCIDRMap[cidr] = true
			}
		}
	}

	// 转换为切片
	for ip := range whitelistMap {
		merged.Whitelist = append(merged.Whitelist, ip)
	}
	for ip := range blacklistMap {
		merged.Blacklist = append(merged.Blacklist, ip)
	}
	for cidr := range whitelistCIDRMap {
		merged.WhitelistCIDR = append(merged.WhitelistCIDR, cidr)
	}
	for cidr := range blacklistCIDRMap {
		merged.BlacklistCIDR = append(merged.BlacklistCIDR, cidr)
	}

	return merged
}

// mergeUserAgentAccessConfigs 合并多个User-Agent访问控制配置
func mergeUserAgentAccessConfigs(records []UserAgentAccessConfigRecord) *security.UserAgentAccessConfig {
	if len(records) == 0 {
		return nil
	}

	// 使用第一个记录作为基础
	first := records[0]
	merged := &security.UserAgentAccessConfig{
		ID:            first.UseragentAccessConfigId,
		Name:          first.ConfigName,
		Enabled:       first.ActiveFlag == "Y",
		DefaultPolicy: first.DefaultPolicy,
		BlockEmpty:    first.BlockEmptyUserAgent == "Y",
		Whitelist:     []string{},
		Blacklist:     []string{},
	}

	// 使用 map 去重
	whitelistMap := make(map[string]bool)
	blacklistMap := make(map[string]bool)

	// 合并所有记录
	for _, record := range records {
		// 合并 Enabled：任何一个启用则启用
		if record.ActiveFlag == "Y" {
			merged.Enabled = true
		}

		// 合并 BlockEmpty：任何一个启用则启用
		if record.BlockEmptyUserAgent == "Y" {
			merged.BlockEmpty = true
		}

		// 合并白名单和黑名单
		if record.WhitelistPatterns != nil && *record.WhitelistPatterns != "" {
			for _, pattern := range parseCommaSeparatedString(*record.WhitelistPatterns) {
				whitelistMap[pattern] = true
			}
		}

		if record.BlacklistPatterns != nil && *record.BlacklistPatterns != "" {
			for _, pattern := range parseCommaSeparatedString(*record.BlacklistPatterns) {
				blacklistMap[pattern] = true
			}
		}
	}

	// 转换为切片
	for pattern := range whitelistMap {
		merged.Whitelist = append(merged.Whitelist, pattern)
	}
	for pattern := range blacklistMap {
		merged.Blacklist = append(merged.Blacklist, pattern)
	}

	return merged
}

// mergeAPIAccessConfigs 合并多个API访问控制配置
func mergeAPIAccessConfigs(records []APIAccessConfigRecord) *security.APIAccessConfig {
	if len(records) == 0 {
		return nil
	}

	// 使用第一个记录作为基础
	first := records[0]
	merged := &security.APIAccessConfig{
		ID:             first.ApiAccessConfigId,
		Name:           first.ConfigName,
		Enabled:        first.ActiveFlag == "Y",
		DefaultPolicy:  first.DefaultPolicy,
		Whitelist:      []string{},
		Blacklist:      []string{},
		AllowedMethods: []string{},
		BlockedMethods: []string{},
	}

	// 使用 map 去重
	whitelistMap := make(map[string]bool)
	blacklistMap := make(map[string]bool)
	allowedMethodsMap := make(map[string]bool)
	blockedMethodsMap := make(map[string]bool)

	// 合并所有记录
	for _, record := range records {
		// 合并 Enabled：任何一个启用则启用
		if record.ActiveFlag == "Y" {
			merged.Enabled = true
		}

		// 合并白名单和黑名单
		if record.WhitelistPaths != nil && *record.WhitelistPaths != "" {
			for _, path := range parseCommaSeparatedString(*record.WhitelistPaths) {
				whitelistMap[path] = true
			}
		}

		if record.BlacklistPaths != nil && *record.BlacklistPaths != "" {
			for _, path := range parseCommaSeparatedString(*record.BlacklistPaths) {
				blacklistMap[path] = true
			}
		}

		// 合并允许和阻止的HTTP方法
		if record.AllowedMethods != "" {
			for _, method := range parseCommaSeparatedString(record.AllowedMethods) {
				allowedMethodsMap[method] = true
			}
		}

		if record.BlockedMethods != nil && *record.BlockedMethods != "" {
			for _, method := range parseCommaSeparatedString(*record.BlockedMethods) {
				blockedMethodsMap[method] = true
			}
		}
	}

	// 转换为切片
	for path := range whitelistMap {
		merged.Whitelist = append(merged.Whitelist, path)
	}
	for path := range blacklistMap {
		merged.Blacklist = append(merged.Blacklist, path)
	}
	for method := range allowedMethodsMap {
		merged.AllowedMethods = append(merged.AllowedMethods, method)
	}
	for method := range blockedMethodsMap {
		merged.BlockedMethods = append(merged.BlockedMethods, method)
	}

	return merged
}

// mergeDomainAccessConfigs 合并多个域名访问控制配置
func mergeDomainAccessConfigs(records []DomainAccessConfigRecord) *security.DomainAccessConfig {
	if len(records) == 0 {
		return nil
	}

	// 使用第一个记录作为基础
	first := records[0]
	merged := &security.DomainAccessConfig{
		ID:              first.DomainAccessConfigId,
		Name:            first.ConfigName,
		Enabled:         first.ActiveFlag == "Y",
		DefaultPolicy:   first.DefaultPolicy,
		AllowSubdomains: first.AllowSubdomains == "Y",
		Whitelist:       []string{},
		Blacklist:       []string{},
	}

	// 使用 map 去重
	whitelistMap := make(map[string]bool)
	blacklistMap := make(map[string]bool)

	// 合并所有记录
	for _, record := range records {
		// 合并 Enabled：任何一个启用则启用
		if record.ActiveFlag == "Y" {
			merged.Enabled = true
		}

		// 合并 AllowSubdomains：任何一个启用则启用
		if record.AllowSubdomains == "Y" {
			merged.AllowSubdomains = true
		}

		// 合并白名单和黑名单
		if record.WhitelistDomains != nil && *record.WhitelistDomains != "" {
			for _, domain := range parseCommaSeparatedString(*record.WhitelistDomains) {
				whitelistMap[domain] = true
			}
		}

		if record.BlacklistDomains != nil && *record.BlacklistDomains != "" {
			for _, domain := range parseCommaSeparatedString(*record.BlacklistDomains) {
				blacklistMap[domain] = true
			}
		}
	}

	// 转换为切片
	for domain := range whitelistMap {
		merged.Whitelist = append(merged.Whitelist, domain)
	}
	for domain := range blacklistMap {
		merged.Blacklist = append(merged.Blacklist, domain)
	}

	return merged
}
