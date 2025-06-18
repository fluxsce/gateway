package security

import (
	"fmt"
)

// SecurityHandlerFactory 安全处理器工厂
type SecurityHandlerFactory struct{}

// NewSecurityHandlerFactory 创建安全处理器工厂
func NewSecurityHandlerFactory() *SecurityHandlerFactory {
	return &SecurityHandlerFactory{}
}

// CreateSecurityHandler 根据配置创建安全处理器
func (f *SecurityHandlerFactory) CreateSecurityHandler(config SecurityConfig) (SecurityHandler, error) {
	// 创建安全处理器
	security := NewSecurity(&config)

	// 验证配置
	if err := security.Validate(); err != nil {
		return nil, fmt.Errorf("安全处理器配置验证失败: %w", err)
	}

	return security, nil
}

// CreateDefaultSecurityHandler 创建默认安全处理器
func (f *SecurityHandlerFactory) CreateDefaultSecurityHandler() (SecurityHandler, error) {
	return f.CreateSecurityHandler(DefaultSecurityConfig)
}

// CreateDisabledSecurityHandler 创建禁用的安全处理器
func (f *SecurityHandlerFactory) CreateDisabledSecurityHandler() (SecurityHandler, error) {
	config := DefaultSecurityConfig
	config.Enabled = false
	config.ID = "disabled-security"
	return f.CreateSecurityHandler(config)
}

// GetSupportedFeatures 获取支持的安全特性
func (f *SecurityHandlerFactory) GetSupportedFeatures() []string {
	return []string{
		"IP访问控制",
		"User-Agent访问控制",
		"API接口访问控制",
		"域名访问控制",
	}
}

// GetFeatureDescription 获取特性描述
func (f *SecurityHandlerFactory) GetFeatureDescription(feature string) string {
	descriptions := map[string]string{
		"IP访问控制":         "基于客户端IP地址的访问控制，支持白名单、黑名单和CIDR",
		"User-Agent访问控制": "基于HTTP User-Agent头的访问控制，支持正则表达式",
		"API接口访问控制":      "基于API路径和HTTP方法的访问控制，支持通配符",
		"域名访问控制":         "基于请求域名的访问控制，支持子域名匹配",
	}

	if desc, exists := descriptions[feature]; exists {
		return desc
	}
	return "未知安全特性"
}
