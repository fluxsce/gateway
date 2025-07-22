package security

/*
DefaultPolicy 使用说明和示例：

1. IP访问控制示例：
   - DefaultPolicy: "allow" + Blacklist: ["192.168.1.100"]
     效果：允许所有IP访问，但拒绝192.168.1.100

   - DefaultPolicy: "deny" + Whitelist: ["192.168.1.0/24"]
     效果：只允许192.168.1.0/24网段访问，拒绝其他所有IP

   - DefaultPolicy: "allow" + Whitelist: ["192.168.1.10"] + Blacklist: ["192.168.1.100"]
     效果：优先拒绝192.168.1.100（黑名单优先），其他IP默认允许

2. User-Agent访问控制示例：
   - DefaultPolicy: "allow" + Blacklist: [".*bot.*"]
     效果：允许所有访问，但拒绝包含"bot"的User-Agent

   - DefaultPolicy: "deny" + Whitelist: ["Mozilla.*Firefox.*"]
     效果：只允许Firefox浏览器访问

3. API接口访问控制示例：
   - DefaultPolicy: "allow" + Blacklist: ["/admin/*"]
     效果：允许访问所有API，但拒绝/admin/路径下的请求

   - DefaultPolicy: "deny" + Whitelist: ["/api/v1/*"] + AllowedMethods: ["GET", "POST"]
     效果：只允许访问/api/v1/路径，且只支持GET和POST方法

4. 域名访问控制示例：
   - DefaultPolicy: "allow" + Blacklist: ["malicious.com"]
     效果：允许所有域名访问，但拒绝malicious.com

   - DefaultPolicy: "deny" + Whitelist: ["api.example.com"] + AllowSubdomains: true
     效果：只允许api.example.com及其子域名访问
*/

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"gateway/internal/gateway/core"
)

// SecurityHandler 安全处理器接口
// 所有安全处理器都必须实现此接口
type SecurityHandler interface {
	// Handle 处理安全检查
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - bool: 是否继续处理后续逻辑
	Handle(ctx *core.Context) bool

	// IsEnabled 是否启用
	// 返回值:
	// - bool: 是否启用
	IsEnabled() bool

	// GetName 获取安全处理器名称
	// 返回值:
	// - string: 处理器名称
	GetName() string

	// Validate 验证配置
	// 返回值:
	// - error: 验证错误
	Validate() error

	// GetConfig 获取配置
	// 返回值:
	// - SecurityConfig: 安全配置
	GetConfig() SecurityConfig
}

// SecurityConfig 安全配置
//
// 安全配置提供多层访问控制机制，包括IP访问控制、User-Agent访问控制、
// API接口访问控制和域名访问控制。每种控制都支持白名单和黑名单模式。
//
// DefaultPolicy 工作原理：
// - "allow" 模式：默认允许访问，通过黑名单拒绝特定请求
// - "deny" 模式：默认拒绝访问，只允许白名单中的请求
//
// 优先级规则：
// 1. 黑名单优先级高于白名单（黑名单中的请求总是被拒绝）
// 2. 白名单优先级高于默认策略
// 3. 如果没有匹配任何白名单或黑名单，则使用默认策略
//
// 安全配置ID
type SecurityConfig struct {
	ID string `json:"id" yaml:"id" mapstructure:"id"`
	// 安全配置名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 是否启用安全检查
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// IP访问控制
	IPAccess IPAccessConfig `json:"ip_access" yaml:"ip_access" mapstructure:"ip_access"`
	// User-Agent访问控制
	UserAgentAccess UserAgentAccessConfig `json:"user_agent_access" yaml:"user_agent_access" mapstructure:"user_agent_access"`
	// API接口访问控制
	APIAccess APIAccessConfig `json:"api_access" yaml:"api_access" mapstructure:"api_access"`
	// 域名访问控制
	DomainAccess DomainAccessConfig `json:"domain_access" yaml:"domain_access" mapstructure:"domain_access"`
	// 自定义配置参数
	CustomConfig map[string]interface{} `json:"custom_config,omitempty" yaml:"custom_config,omitempty" mapstructure:"custom_config,omitempty"`
}

// IPAccessConfig IP访问控制配置
type IPAccessConfig struct {
	// IP访问控制配置ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`
	// IP访问控制配置名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 是否启用IP访问控制
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 默认策略: "allow" 或 "deny"
	// - "allow": 默认允许访问，只有在黑名单中的IP会被拒绝
	// - "deny": 默认拒绝访问，只有在白名单中的IP才被允许
	// 注意: 黑名单优先级高于白名单，无论默认策略如何，黑名单中的IP都会被拒绝
	DefaultPolicy string `json:"default_policy" yaml:"default_policy" mapstructure:"default_policy"`
	// IP白名单
	Whitelist []string `json:"whitelist" yaml:"whitelist" mapstructure:"whitelist"`
	// IP黑名单
	Blacklist []string `json:"blacklist" yaml:"blacklist" mapstructure:"blacklist"`
	// CIDR白名单
	WhitelistCIDR []string `json:"whitelist_cidr" yaml:"whitelist_cidr" mapstructure:"whitelist_cidr"`
	// CIDR黑名单
	BlacklistCIDR []string `json:"blacklist_cidr" yaml:"blacklist_cidr" mapstructure:"blacklist_cidr"`
	// 是否信任X-Forwarded-For头
	TrustXForwardedFor bool `json:"trust_x_forwarded_for" yaml:"trust_x_forwarded_for" mapstructure:"trust_x_forwarded_for"`
	// 是否信任X-Real-IP头
	TrustXRealIP bool `json:"trust_x_real_ip" yaml:"trust_x_real_ip" mapstructure:"trust_x_real_ip"`
}

// UserAgentAccessConfig User-Agent访问控制配置
type UserAgentAccessConfig struct {
	// User-Agent访问控制配置ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`
	// User-Agent访问控制配置名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 是否启用User-Agent访问控制
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 默认策略: "allow" 或 "deny"
	// - "allow": 默认允许访问，只有在黑名单中的User-Agent会被拒绝
	// - "deny": 默认拒绝访问，只有在白名单中的User-Agent才被允许
	// 注意: 黑名单优先级高于白名单，无论默认策略如何，黑名单中的User-Agent都会被拒绝
	DefaultPolicy string `json:"default_policy" yaml:"default_policy" mapstructure:"default_policy"`
	// User-Agent白名单（支持正则表达式）
	Whitelist []string `json:"whitelist" yaml:"whitelist" mapstructure:"whitelist"`
	// User-Agent黑名单（支持正则表达式）
	Blacklist []string `json:"blacklist" yaml:"blacklist" mapstructure:"blacklist"`
	// 是否阻止空User-Agent
	BlockEmpty bool `json:"block_empty" yaml:"block_empty" mapstructure:"block_empty"`
}

// APIAccessConfig API接口访问控制配置
type APIAccessConfig struct {
	// API访问控制配置ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`
	// API访问控制配置名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 是否启用API访问控制
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 默认策略: "allow" 或 "deny"
	// - "allow": 默认允许访问，只有在黑名单中的API路径会被拒绝
	// - "deny": 默认拒绝访问，只有在白名单中的API路径才被允许
	// 注意: 黑名单优先级高于白名单，无论默认策略如何，黑名单中的API路径都会被拒绝
	DefaultPolicy string `json:"default_policy" yaml:"default_policy" mapstructure:"default_policy"`
	// API路径白名单（支持通配符）
	Whitelist []string `json:"whitelist" yaml:"whitelist" mapstructure:"whitelist"`
	// API路径黑名单（支持通配符）
	Blacklist []string `json:"blacklist" yaml:"blacklist" mapstructure:"blacklist"`
	// HTTP方法限制
	AllowedMethods []string `json:"allowed_methods" yaml:"allowed_methods" mapstructure:"allowed_methods"`
	// 受限的HTTP方法
	BlockedMethods []string `json:"blocked_methods" yaml:"blocked_methods" mapstructure:"blocked_methods"`
}

// DomainAccessConfig 域名访问控制配置
type DomainAccessConfig struct {
	// 域名访问控制配置ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`
	// 域名访问控制配置名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 是否启用域名访问控制
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`
	// 默认策略: "allow" 或 "deny"
	// - "allow": 默认允许访问，只有在黑名单中的域名会被拒绝
	// - "deny": 默认拒绝访问，只有在白名单中的域名才被允许
	// 注意: 黑名单优先级高于白名单，无论默认策略如何，黑名单中的域名都会被拒绝
	DefaultPolicy string `json:"default_policy" yaml:"default_policy" mapstructure:"default_policy"`
	// 允许的域名列表
	Whitelist []string `json:"whitelist" yaml:"whitelist" mapstructure:"whitelist"`
	// 禁止的域名列表
	Blacklist []string `json:"blacklist" yaml:"blacklist" mapstructure:"blacklist"`
	// 是否允许子域名
	AllowSubdomains bool `json:"allow_subdomains" yaml:"allow_subdomains" mapstructure:"allow_subdomains"`
}

// DefaultSecurityConfig 默认安全配置
var DefaultSecurityConfig = SecurityConfig{
	ID:      "default-security",
	Name:    "Default Security Configuration",
	Enabled: false,
	IPAccess: IPAccessConfig{
		ID:                 "default-ip-access",
		Name:               "Default IP Access Control",
		Enabled:            false,
		DefaultPolicy:      "allow",
		Whitelist:          []string{},
		Blacklist:          []string{},
		WhitelistCIDR:      []string{},
		BlacklistCIDR:      []string{},
		TrustXForwardedFor: true,
		TrustXRealIP:       true,
	},
	UserAgentAccess: UserAgentAccessConfig{
		ID:            "default-useragent-access",
		Name:          "Default User-Agent Access Control",
		Enabled:       false,
		DefaultPolicy: "allow",
		Whitelist:     []string{},
		Blacklist:     []string{},
		BlockEmpty:    false,
	},
	APIAccess: APIAccessConfig{
		ID:             "default-api-access",
		Name:           "Default API Access Control",
		Enabled:        false,
		DefaultPolicy:  "allow",
		Whitelist:      []string{},
		Blacklist:      []string{},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		BlockedMethods: []string{},
	},
	DomainAccess: DomainAccessConfig{
		ID:              "default-domain-access",
		Name:            "Default Domain Access Control",
		Enabled:         false,
		DefaultPolicy:   "allow",
		Whitelist:       []string{},
		Blacklist:       []string{},
		AllowSubdomains: true,
	},
}

// Security 主安全处理器
// 实现 SecurityHandler 接口
type Security struct {
	config  SecurityConfig
	name    string
	enabled bool
}

// NewSecurity 创建安全处理器
func NewSecurity(config *SecurityConfig) *Security {
	security := &Security{
		config:  DefaultSecurityConfig,
		name:    "Security Handler",
		enabled: false,
	}

	if config != nil {
		security.config = *config
		security.enabled = config.Enabled
		if config.Name != "" {
			security.name = config.Name
		} else if config.ID != "" {
			security.name = config.ID
		}
	}

	return security
}

// Handle 实现 SecurityHandler 接口
func (s *Security) Handle(ctx *core.Context) bool {
	// 如果未启用安全检查，继续处理
	if !s.enabled {
		return true
	}

	// 使用内置的安全检查逻辑
	return s.defaultSecurityCheck(ctx)
}

// defaultSecurityCheck 默认安全检查逻辑
func (s *Security) defaultSecurityCheck(ctx *core.Context) bool {
	// IP访问控制检查
	if !s.checkIPAccess(ctx) {
		ctx.Abort(http.StatusForbidden, map[string]string{
			"error": "IP access denied",
		})
		return false
	}

	// User-Agent访问控制检查
	if !s.checkUserAgentAccess(ctx) {
		ctx.Abort(http.StatusForbidden, map[string]string{
			"error": "User-Agent access denied",
		})
		return false
	}

	// API接口访问控制检查
	if !s.checkAPIAccess(ctx) {
		ctx.Abort(http.StatusForbidden, map[string]string{
			"error": "API access denied",
		})
		return false
	}

	// 域名访问控制检查
	if !s.checkDomainAccess(ctx) {
		ctx.Abort(http.StatusForbidden, map[string]string{
			"error": "Domain access denied",
		})
		return false
	}

	return true
}

// IsEnabled 实现 SecurityHandler 接口
func (s *Security) IsEnabled() bool {
	return s.enabled
}

// GetName 实现 SecurityHandler 接口
func (s *Security) GetName() string {
	return s.name
}

// Validate 实现 SecurityHandler 接口
func (s *Security) Validate() error {
	// 基础验证
	if s.config.ID == "" {
		return fmt.Errorf("安全配置ID不能为空")
	}

	// 验证 IP 访问控制配置
	if s.config.IPAccess.Enabled {
		if s.config.IPAccess.DefaultPolicy != "allow" && s.config.IPAccess.DefaultPolicy != "deny" {
			return fmt.Errorf("IP访问控制默认策略必须是 'allow' 或 'deny'")
		}
	}

	// 验证 User-Agent 访问控制配置
	if s.config.UserAgentAccess.Enabled {
		if s.config.UserAgentAccess.DefaultPolicy != "allow" && s.config.UserAgentAccess.DefaultPolicy != "deny" {
			return fmt.Errorf("User-Agent访问控制默认策略必须是 'allow' 或 'deny'")
		}
	}

	// 验证 API 访问控制配置
	if s.config.APIAccess.Enabled {
		if s.config.APIAccess.DefaultPolicy != "allow" && s.config.APIAccess.DefaultPolicy != "deny" {
			return fmt.Errorf("API访问控制默认策略必须是 'allow' 或 'deny'")
		}
	}

	// 验证域名访问控制配置
	if s.config.DomainAccess.Enabled {
		if s.config.DomainAccess.DefaultPolicy != "allow" && s.config.DomainAccess.DefaultPolicy != "deny" {
			return fmt.Errorf("域名访问控制默认策略必须是 'allow' 或 'deny'")
		}
	}

	return nil
}

// GetConfig 获取安全配置
func (s *Security) GetConfig() SecurityConfig {
	return s.config
}

// SetConfig 设置安全配置
func (s *Security) SetConfig(config SecurityConfig) {
	s.config = config
}

// checkIPAccess 检查IP访问权限
//
// DefaultPolicy 实现逻辑：
// 1. 优先检查黑名单：如果IP在黑名单中，直接拒绝
// 2. 检查白名单：如果IP在白名单中，直接允许
// 3. 应用默认策略：
//   - "allow": 默认允许（黑名单之外的IP都允许）
//   - "deny": 默认拒绝（只有白名单中的IP才允许）
func (s *Security) checkIPAccess(ctx *core.Context) bool {
	if !s.config.IPAccess.Enabled {
		return true
	}

	clientIP := s.getClientIP(ctx)
	if clientIP == "" {
		return false
	}

	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}

	// 1. 优先检查黑名单（黑名单优先级最高）
	if s.isIPInList(clientIP, s.config.IPAccess.Blacklist) ||
		s.isIPInCIDRList(ip, s.config.IPAccess.BlacklistCIDR) {
		return false
	}

	// 2. 检查白名单（白名单优先级高于默认策略）
	isInWhitelist := s.isIPInList(clientIP, s.config.IPAccess.Whitelist) ||
		s.isIPInCIDRList(ip, s.config.IPAccess.WhitelistCIDR)

	if isInWhitelist {
		return true
	}

	// 3. 应用默认策略
	if s.config.IPAccess.DefaultPolicy == "deny" {
		// 默认拒绝：只有白名单中的IP才允许（已在上面检查过）
		return false
	}

	// 默认允许：除了黑名单之外的IP都允许（黑名单已在上面检查过）
	return true
}

// checkUserAgentAccess 检查User-Agent访问权限
//
// DefaultPolicy 实现逻辑：
// 1. 检查是否阻止空User-Agent
// 2. 优先检查黑名单：如果User-Agent在黑名单中，直接拒绝
// 3. 检查白名单：如果User-Agent在白名单中，直接允许
// 4. 应用默认策略：
//   - "allow": 默认允许（黑名单之外的User-Agent都允许）
//   - "deny": 默认拒绝（只有白名单中的User-Agent才允许）
func (s *Security) checkUserAgentAccess(ctx *core.Context) bool {
	if !s.config.UserAgentAccess.Enabled {
		return true
	}

	userAgent := ctx.Request.Header.Get("User-Agent")

	// 1. 检查是否阻止空User-Agent（优先级最高）
	if s.config.UserAgentAccess.BlockEmpty && userAgent == "" {
		return false
	}

	// 2. 优先检查黑名单（黑名单优先级最高）
	if s.isUserAgentInList(userAgent, s.config.UserAgentAccess.Blacklist) {
		return false
	}

	// 3. 检查白名单（白名单优先级高于默认策略）
	isInWhitelist := s.isUserAgentInList(userAgent, s.config.UserAgentAccess.Whitelist)

	if isInWhitelist {
		return true
	}

	// 4. 应用默认策略
	if s.config.UserAgentAccess.DefaultPolicy == "deny" {
		// 默认拒绝：只有白名单中的User-Agent才允许（已在上面检查过）
		return false
	}

	// 默认允许：除了黑名单之外的User-Agent都允许（黑名单已在上面检查过）
	return true
}

// checkAPIAccess 检查API接口访问权限
//
// DefaultPolicy 实现逻辑：
// 1. 检查HTTP方法限制
// 2. 优先检查API路径黑名单：如果路径在黑名单中，直接拒绝
// 3. 检查API路径白名单：如果路径在白名单中，直接允许
// 4. 应用默认策略：
//   - "allow": 默认允许（黑名单之外的API路径都允许）
//   - "deny": 默认拒绝（只有白名单中的API路径才允许）
func (s *Security) checkAPIAccess(ctx *core.Context) bool {
	if !s.config.APIAccess.Enabled {
		return true
	}

	path := ctx.Request.URL.Path
	method := ctx.Request.Method

	// 1. 检查HTTP方法限制（优先级最高）
	if len(s.config.APIAccess.BlockedMethods) > 0 {
		if s.isStringInList(method, s.config.APIAccess.BlockedMethods) {
			return false
		}
	}

	if len(s.config.APIAccess.AllowedMethods) > 0 {
		if !s.isStringInList(method, s.config.APIAccess.AllowedMethods) {
			return false
		}
	}

	// 2. 优先检查API路径黑名单（黑名单优先级最高）
	if s.isPathInList(path, s.config.APIAccess.Blacklist) {
		return false
	}

	// 3. 检查API路径白名单（白名单优先级高于默认策略）
	isInWhitelist := s.isPathInList(path, s.config.APIAccess.Whitelist)

	if isInWhitelist {
		return true
	}

	// 4. 应用默认策略
	if s.config.APIAccess.DefaultPolicy == "deny" {
		// 默认拒绝：只有白名单中的API路径才允许（已在上面检查过）
		return false
	}

	// 默认允许：除了黑名单之外的API路径都允许（黑名单已在上面检查过）
	return true
}

// checkDomainAccess 检查域名访问权限
//
// DefaultPolicy 实现逻辑：
// 1. 获取并清理域名（移除端口号）
// 2. 优先检查域名黑名单：如果域名在黑名单中，直接拒绝
// 3. 检查域名白名单：如果域名在白名单中，直接允许
// 4. 应用默认策略：
//   - "allow": 默认允许（黑名单之外的域名都允许）
//   - "deny": 默认拒绝（只有白名单中的域名才允许）
func (s *Security) checkDomainAccess(ctx *core.Context) bool {
	if !s.config.DomainAccess.Enabled {
		return true
	}

	host := ctx.Request.Host
	if host == "" {
		return false
	}

	// 1. 移除端口号，获取纯域名
	if colonIndex := strings.Index(host, ":"); colonIndex != -1 {
		host = host[:colonIndex]
	}

	// 2. 优先检查域名黑名单（黑名单优先级最高）
	if s.isDomainInList(host, s.config.DomainAccess.Blacklist) {
		return false
	}

	// 3. 检查域名白名单（白名单优先级高于默认策略）
	isInWhitelist := s.isDomainInList(host, s.config.DomainAccess.Whitelist)

	if isInWhitelist {
		return true
	}

	// 4. 应用默认策略
	if s.config.DomainAccess.DefaultPolicy == "deny" {
		// 默认拒绝：只有白名单中的域名才允许（已在上面检查过）
		return false
	}

	// 默认允许：除了黑名单之外的域名都允许（黑名单已在上面检查过）
	return true
}

// getClientIP 获取客户端IP
func (s *Security) getClientIP(ctx *core.Context) string {
	// 优先从X-Forwarded-For获取
	if s.config.IPAccess.TrustXForwardedFor {
		if xff := ctx.Request.Header.Get("X-Forwarded-For"); xff != "" {
			// X-Forwarded-For可能包含多个IP，取第一个
			if commaIndex := strings.Index(xff, ","); commaIndex != -1 {
				return strings.TrimSpace(xff[:commaIndex])
			}
			return strings.TrimSpace(xff)
		}
	}

	// 从X-Real-IP获取
	if s.config.IPAccess.TrustXRealIP {
		if xrip := ctx.Request.Header.Get("X-Real-IP"); xrip != "" {
			return strings.TrimSpace(xrip)
		}
	}

	// 从RemoteAddr获取
	if colonIndex := strings.LastIndex(ctx.Request.RemoteAddr, ":"); colonIndex != -1 {
		return ctx.Request.RemoteAddr[:colonIndex]
	}
	return ctx.Request.RemoteAddr
}

// isIPInList 检查IP是否在列表中
func (s *Security) isIPInList(ip string, list []string) bool {
	for _, item := range list {
		if ip == item {
			return true
		}
	}
	return false
}

// isIPInCIDRList 检查IP是否在CIDR列表中
func (s *Security) isIPInCIDRList(ip net.IP, cidrList []string) bool {
	for _, cidr := range cidrList {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// isUserAgentInList 检查User-Agent是否在列表中（支持正则表达式）
func (s *Security) isUserAgentInList(userAgent string, list []string) bool {
	for _, pattern := range list {
		// 尝试作为正则表达式匹配
		if matched, err := regexp.MatchString(pattern, userAgent); err == nil && matched {
			return true
		}
		// 如果正则表达式失败，尝试精确匹配
		if userAgent == pattern {
			return true
		}
	}
	return false
}

// isStringInList 检查字符串是否在列表中
func (s *Security) isStringInList(str string, list []string) bool {
	for _, item := range list {
		if str == item {
			return true
		}
	}
	return false
}

// isPathInList 检查路径是否在列表中（支持通配符）
func (s *Security) isPathInList(path string, list []string) bool {
	for _, pattern := range list {
		if matched, _ := s.matchPath(pattern, path); matched {
			return true
		}
	}
	return false
}

// isDomainInList 检查域名是否在列表中
func (s *Security) isDomainInList(domain string, list []string) bool {
	for _, item := range list {
		if domain == item {
			return true
		}
		// 如果允许子域名，检查是否为子域名
		if s.config.DomainAccess.AllowSubdomains {
			if strings.HasSuffix(domain, "."+item) {
				return true
			}
		}
	}
	return false
}

// matchPath 路径通配符匹配
func (s *Security) matchPath(pattern, path string) (bool, error) {
	// 简单的通配符匹配实现
	// * 匹配任意字符序列
	// ? 匹配单个字符

	// 将通配符模式转换为正则表达式
	regexPattern := strings.ReplaceAll(pattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")
	regexPattern = "^" + regexPattern + "$"

	return regexp.MatchString(regexPattern, path)
}
