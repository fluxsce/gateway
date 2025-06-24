package router

import (
	"fmt"
	"net/url"
	"strings"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/assertion"
	"gohub/internal/gateway/handler/auth"
	"gohub/internal/gateway/handler/cors"
	"gohub/internal/gateway/handler/filter"
	"gohub/internal/gateway/handler/limiter"
	"gohub/internal/gateway/handler/security"
)

// RouteHandler 路由处理器接口
// 所有路由处理器都必须实现此接口
type RouteHandler interface {
	// Handle 处理请求
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - bool: 是否继续处理后续逻辑
	Handle(ctx *core.Context) bool

	// IsEnabled 是否启用
	// 返回值:
	// - bool: 是否启用
	IsEnabled() bool

	// GetName 获取处理器名称
	// 返回值:
	// - string: 处理器名称
	GetName() string

	// Validate 验证配置
	// 返回值:
	// - error: 验证错误
	Validate() error

	// GetConfig 获取配置
	// 返回值:
	// - Config: 路由配置
	GetConfig() RouteConfig

	// Match 检查是否匹配当前请求
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - bool: 是否匹配
	// - error: 匹配过程中的错误
	Match(ctx *core.Context) (bool, error)
}

// Config 路由配置结构
// 仅包含可序列化的配置信息，不包含运行时实例
type RouteConfig struct {
	// ========== 基础路由配置 ==========

	// 路由ID - 路由的唯一标识符，用于引用和管理路由
	ID string `json:"id" yaml:"id" mapstructure:"id"`

	// 路由名称 - 路由的可读名称，用于显示和识别
	Name string `json:"name" yaml:"name" mapstructure:"name"`

	// 服务ID - 匹配此路由的请求将被转发到的目标服务ID
	ServiceID string `json:"service_id" yaml:"service_id" mapstructure:"service_id"`

	// 路由路径 - 用于创建断言，根据需要生成不同类型的路径断言
	Path string `json:"path" yaml:"path" mapstructure:"path"`

	// 允许的HTTP方法，为空表示允许所有方法
	// 例如: ["GET", "POST"]、["*"]
	Methods []string `json:"methods,omitempty" yaml:"methods,omitempty" mapstructure:"methods,omitempty"`

	// 是否启用 - 控制路由是否参与匹配过程，可用于临时禁用路由
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`

	// 优先级 - 数值越小优先级越高，影响路由的匹配顺序
	// 同等条件下，高优先级的路由将优先匹配
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty" mapstructure:"priority,omitempty"`

	// 路由元数据 - 存储与路由相关的额外信息，可用于自定义处理逻辑
	Metadata map[string]interface{} `json:"metadata,omitempty" yaml:"metadata,omitempty" mapstructure:"metadata,omitempty"`

	// ========== 断言配置 ==========

	// 断言组配置 - 用于序列化和配置文件存储
	AssertionGroupConfig *assertion.AssertionGroupConfig `json:"assertion_group_config,omitempty" yaml:"assertion_group_config,omitempty" mapstructure:"assertion_group_config,omitempty"`

	// ========== 过滤器配置 ==========

	// 过滤器配置 - 用于序列化和配置文件存储
	FilterConfig []filter.FilterConfig `json:"filter_config,omitempty" yaml:"filter_config,omitempty" mapstructure:"filter_config,omitempty"`

	// ========== 功能模块配置 ==========
	// 这些字段存储路由级别的配置信息，无论功能是否启用都会保存

	// CORS配置
	CorsConfig *cors.CORSConfig `json:"cors_config,omitempty" yaml:"cors_config,omitempty" mapstructure:"cors_config,omitempty"`

	// 限流器配置
	LimiterConfig *limiter.RateLimitConfig `json:"limiter_config,omitempty" yaml:"limiter_config,omitempty" mapstructure:"limiter_config,omitempty"`

	// 认证配置
	AuthConfig *auth.AuthConfig `json:"auth_config,omitempty" yaml:"auth_config,omitempty" mapstructure:"auth_config,omitempty"`

	// 安全配置
	SecurityConfig *security.SecurityConfig `json:"security_config,omitempty" yaml:"security_config,omitempty" mapstructure:"security_config,omitempty"`
}

// DefaultRouteConfig 默认路由配置
var DefaultRouteConfig = RouteConfig{
	Enabled:  true,
	Priority: 100,
	Methods:  []string{},
	Metadata: make(map[string]interface{}),
	Name:     "Default Route",
}

// Route 实现 RouteHandler 接口的具体路由结构
type Route struct {
	// 配置
	config RouteConfig

	// 是否启用
	enabled bool

	// 处理器名称
	name string

	// 断言组
	assertionGroup *assertion.AssertionGroup

	// 路由级别过滤器
	routeFilters []filter.Filter

	// 功能模块处理器
	corsHandler     cors.CORSHandler
	limiterHandler  limiter.LimiterHandler
	authHandler     auth.Authenticator
	securityHandler security.SecurityHandler
}

// NewRoute 创建新的路由实例
func NewRoute(config RouteConfig) (*Route, error) {
	route := &Route{
		config:       config,
		enabled:      config.Enabled,
		name:         config.Name,
		routeFilters: make([]filter.Filter, 0),
	}

	// 如果名称为空，使用ID作为名称
	if route.name == "" {
		route.name = config.ID
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("route config validation failed: %w", err)
	}

	// 初始化断言组
	if config.AssertionGroupConfig != nil {
		// 使用直接从配置创建断言组的方法
		assertionGroup, err := assertion.NewAssertionGroupFromConfig(config.AssertionGroupConfig)
		if err != nil {
			return nil, fmt.Errorf("create assertion group failed: %w", err)
		}
		route.assertionGroup = assertionGroup
	} else {
		route.assertionGroup = assertion.NewAssertionGroup(true)
	}

	// 初始化路由级别过滤器
	if len(config.FilterConfig) > 0 {
		filterFactory := filter.NewFilterFactory()
		for _, filterConfig := range config.FilterConfig {
			// 创建过滤器实例
			filterInstance, err := filterFactory.CreateFilter(filterConfig)
			if err != nil {
				return nil, fmt.Errorf("create filter failed: %w", err)
			}
			// 将过滤器添加到切片中
			route.routeFilters = append(route.routeFilters, filterInstance)
		}
	}

	// 初始化功能模块处理器
	if err := route.initHandlers(); err != nil {
		return nil, fmt.Errorf("init handlers failed: %w", err)
	}

	return route, nil
}

// initHandlers 初始化处理器
func (r *Route) initHandlers() error {
	// 初始化CORS处理器
	if r.config.CorsConfig != nil && r.config.CorsConfig.Enabled {
		factory := cors.NewCORSHandlerFactory()
		corsHandler, err := factory.CreateCORSHandler(*r.config.CorsConfig)
		if err != nil {
			return fmt.Errorf("create CORS handler failed: %w", err)
		}
		r.corsHandler = corsHandler
	}

	// 初始化限流处理器
	if r.config.LimiterConfig != nil && r.config.LimiterConfig.Enabled {
		factory := limiter.NewLimiterFactory()
		limiterHandler, err := factory.CreateLimiter(r.config.LimiterConfig)
		if err != nil {
			return fmt.Errorf("create limiter handler failed: %w", err)
		}
		r.limiterHandler = limiterHandler
	}

	// 初始化认证处理器
	if r.config.AuthConfig != nil && r.config.AuthConfig.Strategy != auth.StrategyNoAuth {
		// 使用工厂创建认证处理器
		factory := auth.NewAuthenticatorFactory()
		authHandler, err := factory.CreateAuthenticator(*r.config.AuthConfig)
		if err != nil {
			return fmt.Errorf("create auth handler failed: %w", err)
		}
		r.authHandler = authHandler
	}

	// 初始化安全处理器
	if r.config.SecurityConfig != nil && r.config.SecurityConfig.Enabled {
		// 使用工厂创建安全处理器
		factory := security.NewSecurityHandlerFactory()
		securityHandler, err := factory.CreateSecurityHandler(*r.config.SecurityConfig)
		if err != nil {
			return fmt.Errorf("create security handler failed: %w", err)
		}
		r.securityHandler = securityHandler
	}

	return nil
}

// Handle 处理请求
func (r *Route) Handle(ctx *core.Context) bool {
	if !r.enabled {
		return true
	}

	// 设置路由信息到上下文
	ctx.SetRouteID(r.config.ID)
	ctx.SetServiceID(r.config.ServiceID)
	ctx.SetMatchedPath(r.config.Path)

	// 按预定义顺序执行所有处理器
	// 执行顺序: Security → CORS → Auth → Limiter → FilterChain
	// 1. Security 处理
	if r.securityHandler != nil {
		if !r.securityHandler.Handle(ctx) {
			return false
		}
	}
	// 2. CORS 处理
	if r.corsHandler != nil {
		if !r.corsHandler.Handle(ctx) {
			return false
		}
	}

	// 3. Auth 处理
	if r.authHandler != nil {
		if !r.authHandler.Handle(ctx) {
			return false
		}
	}

	// 4. Limiter 处理
	if r.limiterHandler != nil {
		if !r.limiterHandler.Handle(ctx) {
			return false
		}
	}

	// 5. 执行路由级别过滤器
	if len(r.routeFilters) > 0 {
		// 按照过滤器的定义执行不同阶段的过滤器
		for _, f := range r.routeFilters {
			if !f.IsEnabled() {
				continue
			}
			if err := f.Apply(ctx); err != nil {
				ctx.AddError(err)
				return false
			}
		}
	}

	// 标记请求为已路由
	ctx.Set("routed", true)

	return true
}

// Match 检查是否匹配当前请求
func (r *Route) Match(ctx *core.Context) (bool, error) {
	if !r.enabled {
		return false, nil
	}

	req := ctx.Request
	if req == nil {
		return false, nil
	}

	// 1. 优先检查路径前缀匹配
	if !r.isPathMatched(req.URL.Path) {
		return false, nil
	}

	// 2. 检查HTTP方法
	if !r.isMethodAllowed(req.Method) {
		return false, nil
	}

	// 3. 执行其他断言组匹配
	if r.assertionGroup != nil {
		matches, err := r.assertionGroup.Evaluate(ctx)
		if err != nil {
			return false, fmt.Errorf("assertion group evaluation failed: %w", err)
		}
		return matches, nil
	}

	return true, nil
}

// isPathMatched 检查请求路径是否匹配路由路径前缀
func (r *Route) isPathMatched(requestPath string) bool {
	routePath := r.config.Path

	// 如果路由路径为空或为根路径，匹配所有请求
	if routePath == "" || routePath == "/" {
		return true
	}

	// 确保路径以 / 开头
	if !strings.HasPrefix(routePath, "/") {
		routePath = "/" + routePath
	}

	// 处理通配符匹配
	if strings.Contains(routePath, "*") {
		return r.matchWithWildcard(requestPath, routePath)
	}

	// 前缀匹配
	return strings.HasPrefix(requestPath, routePath)
}

// matchWithWildcard 处理带通配符的路径匹配
func (r *Route) matchWithWildcard(requestPath, routePath string) bool {
	// 处理结尾的多个 * 号，将连续的 * 替换为单个 *
	routePath = r.normalizeWildcards(routePath)

	// 如果路径以 * 结尾，进行前缀匹配
	if strings.HasSuffix(routePath, "*") {
		prefix := strings.TrimSuffix(routePath, "*")
		// 去掉可能的尾部斜杠
		prefix = strings.TrimSuffix(prefix, "/")
		if prefix == "" {
			return true // 只有 * 或 /* 的情况，匹配所有
		}
		return strings.HasPrefix(requestPath, prefix)
	}

	// 处理中间包含通配符的情况
	// 将路径按 * 分割，逐段匹配
	parts := strings.Split(routePath, "*")
	if len(parts) == 1 {
		// 没有通配符，直接前缀匹配
		return strings.HasPrefix(requestPath, routePath)
	}

	// 检查第一部分是否匹配
	if !strings.HasPrefix(requestPath, parts[0]) {
		return false
	}

	// 检查最后一部分是否匹配（如果不为空）
	lastPart := parts[len(parts)-1]
	if lastPart != "" && !strings.HasSuffix(requestPath, lastPart) {
		return false
	}

	// 检查中间部分
	currentPos := len(parts[0])
	for i := 1; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" {
			continue
		}

		index := strings.Index(requestPath[currentPos:], part)
		if index == -1 {
			return false
		}
		currentPos += index + len(part)
	}

	return true
}

// normalizeWildcards 标准化通配符，将连续的 * 替换为单个 *
func (r *Route) normalizeWildcards(path string) string {
	// 使用正则表达式替换连续的 * 为单个 *
	for strings.Contains(path, "**") {
		path = strings.ReplaceAll(path, "**", "*")
	}
	return path
}

// isMethodAllowed 检查请求方法是否被允许
func (r *Route) isMethodAllowed(method string) bool {
	if len(r.config.Methods) == 0 {
		return true // 允许所有方法
	}

	for _, allowedMethod := range r.config.Methods {
		if allowedMethod == "*" || strings.EqualFold(allowedMethod, method) {
			return true
		}
	}
	return false
}

// IsEnabled 返回路由是否启用
func (r *Route) IsEnabled() bool {
	return r.enabled
}

// GetName 返回路由名称
func (r *Route) GetName() string {
	return r.name
}

// GetConfig 返回路由配置
func (r *Route) GetConfig() RouteConfig {
	return r.config
}

// Validate 验证路由配置
func (r *Route) Validate() error {
	return r.config.Validate()
}

// Validate 验证配置
func (config *RouteConfig) Validate() error {
	if config.ID == "" {
		return fmt.Errorf("route ID cannot be empty")
	}

	if config.ServiceID == "" {
		return fmt.Errorf("service ID cannot be empty")
	}

	if config.Path == "" {
		return fmt.Errorf("route path cannot be empty")
	}

	// 验证路径格式
	if _, err := url.Parse(config.Path); err != nil {
		return fmt.Errorf("invalid route path format: %w", err)
	}

	return nil
}

// Clone 克隆配置
func (config *RouteConfig) Clone() RouteConfig {
	clone := *config

	// 深拷贝Methods切片
	if config.Methods != nil {
		clone.Methods = make([]string, len(config.Methods))
		copy(clone.Methods, config.Methods)
	}

	// 深拷贝Metadata映射
	if config.Metadata != nil {
		clone.Metadata = make(map[string]interface{})
		for k, v := range config.Metadata {
			clone.Metadata[k] = v
		}
	}

	return clone
}

// RouteFromConfig 从配置创建路由
func RouteFromConfig(config RouteConfig) (RouteHandler, error) {
	return NewRoute(config)
}
