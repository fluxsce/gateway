package dbloader

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gohub/internal/gateway/handler/security"
	"gohub/pkg/database"
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
	query := `
		SELECT tenantId, securityConfigId, gatewayInstanceId, routeConfigId, configName,
		       configDesc, configPriority, customConfigJson, activeFlag
		FROM HUB_GATEWAY_SECURITY_CONFIG 
		WHERE tenantId = ? AND gatewayInstanceId = ? 
		AND (routeConfigId IS NULL OR routeConfigId = '') 
		AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record SecurityConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, instanceId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询安全配置失败: %w", err)
	}

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
	query := `
		SELECT tenantId, securityConfigId, gatewayInstanceId, routeConfigId, configName,
		       configDesc, configPriority, customConfigJson, activeFlag
		FROM HUB_GATEWAY_SECURITY_CONFIG 
		WHERE tenantId = ? AND routeConfigId = ? AND activeFlag = 'Y'
		ORDER BY configPriority ASC
		LIMIT 1
	`

	var record SecurityConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, routeId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询路由安全配置失败: %w", err)
	}

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
	query := `
		SELECT tenantId, ipAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistIps, blacklistIps, whitelistCidrs, blacklistCidrs,
		       trustXForwardedFor, trustXRealIp, activeFlag
		FROM HUB_GATEWAY_IP_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var record IPAccessConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, securityConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询IP访问控制配置失败: %w", err)
	}

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
		ipAccessConfig.Whitelist = strings.Split(*record.WhitelistIps, ",")
	}

	if record.BlacklistIps != nil && *record.BlacklistIps != "" {
		ipAccessConfig.Blacklist = strings.Split(*record.BlacklistIps, ",")
	}

	if record.WhitelistCidrs != nil && *record.WhitelistCidrs != "" {
		ipAccessConfig.WhitelistCIDR = strings.Split(*record.WhitelistCidrs, ",")
	}

	if record.BlacklistCidrs != nil && *record.BlacklistCidrs != "" {
		ipAccessConfig.BlacklistCIDR = strings.Split(*record.BlacklistCidrs, ",")
	}

	return ipAccessConfig, nil
}

// LoadUserAgentAccessConfig 加载User-Agent访问控制配置
func (loader *SecurityConfigLoader) LoadUserAgentAccessConfig(ctx context.Context, securityConfigId string) (*security.UserAgentAccessConfig, error) {
	query := `
		SELECT tenantId, useragentAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPatterns, blacklistPatterns, blockEmptyUserAgent, activeFlag
		FROM HUB_GATEWAY_USERAGENT_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var record UserAgentAccessConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, securityConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询User-Agent访问控制配置失败: %w", err)
	}

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
		userAgentConfig.Whitelist = strings.Split(*record.WhitelistPatterns, ",")
	}

	if record.BlacklistPatterns != nil && *record.BlacklistPatterns != "" {
		userAgentConfig.Blacklist = strings.Split(*record.BlacklistPatterns, ",")
	}

	return userAgentConfig, nil
}

// LoadAPIAccessConfig 加载API访问控制配置
func (loader *SecurityConfigLoader) LoadAPIAccessConfig(ctx context.Context, securityConfigId string) (*security.APIAccessConfig, error) {
	query := `
		SELECT tenantId, apiAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistPaths, blacklistPaths, allowedMethods, blockedMethods, activeFlag
		FROM HUB_GATEWAY_API_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var record APIAccessConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, securityConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询API访问控制配置失败: %w", err)
	}

	// 构建API访问控制配置
	apiAccessConfig := &security.APIAccessConfig{
		ID:            record.ApiAccessConfigId,
		Name:          record.ConfigName,
		Enabled:       record.ActiveFlag == "Y",
		DefaultPolicy: record.DefaultPolicy,
	}

	// 解析白名单和黑名单
	if record.WhitelistPaths != nil && *record.WhitelistPaths != "" {
		apiAccessConfig.Whitelist = strings.Split(*record.WhitelistPaths, ",")
	}

	if record.BlacklistPaths != nil && *record.BlacklistPaths != "" {
		apiAccessConfig.Blacklist = strings.Split(*record.BlacklistPaths, ",")
	}

	// 解析允许和阻止的HTTP方法
	if record.AllowedMethods != "" {
		apiAccessConfig.AllowedMethods = strings.Split(record.AllowedMethods, ",")
	}

	if record.BlockedMethods != nil && *record.BlockedMethods != "" {
		apiAccessConfig.BlockedMethods = strings.Split(*record.BlockedMethods, ",")
	}

	return apiAccessConfig, nil
}

// LoadDomainAccessConfig 加载域名访问控制配置
func (loader *SecurityConfigLoader) LoadDomainAccessConfig(ctx context.Context, securityConfigId string) (*security.DomainAccessConfig, error) {
	query := `
		SELECT tenantId, domainAccessConfigId, securityConfigId, configName, defaultPolicy,
		       whitelistDomains, blacklistDomains, allowSubdomains, activeFlag
		FROM HUB_GATEWAY_DOMAIN_ACCESS_CONFIG 
		WHERE tenantId = ? AND securityConfigId = ? AND activeFlag = 'Y'
		LIMIT 1
	`

	var record DomainAccessConfigRecord
	err := loader.db.QueryOne(ctx, &record, query, []interface{}{loader.tenantId, securityConfigId}, true)
	if err != nil {
		if err == database.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("查询域名访问控制配置失败: %w", err)
	}

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
		domainAccessConfig.Whitelist = strings.Split(*record.WhitelistDomains, ",")
	}

	if record.BlacklistDomains != nil && *record.BlacklistDomains != "" {
		domainAccessConfig.Blacklist = strings.Split(*record.BlacklistDomains, ",")
	}

	return domainAccessConfig, nil
} 