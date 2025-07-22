package router

import (
	"errors"
	"fmt"
	"gohub/internal/gateway/constants"
	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/filter"
	"sort"
	"sync"
)

var (
	// ErrRouteNotFound 路由未找到错误
	ErrRouteNotFound = errors.New("route not found")

	// ErrInvalidRoutePattern 无效的路由模式错误
	ErrInvalidRoutePattern = errors.New("invalid route pattern")

	// ErrAssertionFailed 断言失败错误
	ErrAssertionFailed = errors.New("request failed to meet assertion requirements")
)

// RouterHandler 路由器处理器接口
type RouterHandler interface {
	// Handle 处理请求
	Handle(ctx *core.Context) bool

	// AddRoute 添加路由
	AddRoute(config RouteConfig) error

	// RemoveRoute 删除路由
	RemoveRoute(routeID string) error

	// GetRoute 根据ID获取路由
	GetRoute(routeID string) (RouteHandler, error)

	// ListRoutes 列出所有路由
	ListRoutes() []RouteHandler

	// IsEnabled 是否启用
	IsEnabled() bool

	// GetName 获取处理器名称
	GetName() string

	// Validate 验证配置
	Validate() error
}


// RouterConfig 路由器配置结构
type RouterConfig struct {
	// 路由器ID
	ID string `json:"id" yaml:"id" mapstructure:"id"`

	// 是否启用
	Enabled bool `json:"enabled" yaml:"enabled" mapstructure:"enabled"`

	// 路由器名称
	Name string `json:"name,omitempty" yaml:"name,omitempty" mapstructure:"name,omitempty"`

	// 路由列表
	Routes []RouteConfig `json:"routes" yaml:"routes" mapstructure:"routes"`

	// 过滤器配置
	FilterConfig []filter.FilterConfig `json:"filter_config,omitempty" yaml:"filter_config,omitempty" mapstructure:"filter_config,omitempty"`

	// 默认优先级
	DefaultPriority int `json:"default_priority,omitempty" yaml:"default_priority,omitempty" mapstructure:"default_priority,omitempty"`

	// 是否启用路由缓存
	EnableRouteCache bool `json:"enable_route_cache,omitempty" yaml:"enable_route_cache,omitempty" mapstructure:"enable_route_cache,omitempty"`

	// 路由缓存TTL(秒)
	RouteCacheTTL int `json:"route_cache_ttl,omitempty" yaml:"route_cache_ttl,omitempty" mapstructure:"route_cache_ttl,omitempty"`
}

// DefaultRouterConfig 默认路由器配置
var DefaultRouterConfig = RouterConfig{
	ID:               "default-router",
	Enabled:          true,
	Name:             "default-router",
	Routes:           []RouteConfig{},
	FilterConfig:     []filter.FilterConfig{},
	DefaultPriority:  100,
	EnableRouteCache: false,
	RouteCacheTTL:    300,
}

// Router 路由处理器
// 管理多个实现了 RouteHandler 接口的路由实例，类似于 filterChain 管理 filter
type Router struct {
	// 配置
	config RouterConfig

	// 是否启用
	enabled bool

	// 处理器名称
	name string

	// 按ID索引的路由处理器
	routes map[string]RouteHandler

	// 根据优先级排序的路由处理器列表
	prioritizedRoutes []RouteHandler

	// 路由是否已排序
	routesSorted bool

	// 全局过滤器链
	routerFilters []filter.Filter

	// 互斥锁
	mu sync.RWMutex
}

// NewRouter 创建新的路由处理器
func NewRouter(config RouterConfig) *Router {
	return &Router{
		config:            config,
		enabled:           config.Enabled,
		name:              config.Name,
		routes:            make(map[string]RouteHandler),
		prioritizedRoutes: make([]RouteHandler, 0),
		routesSorted:      false,
		routerFilters:     make([]filter.Filter, 0),
	}
}

// Handle 处理请求
func (r *Router) Handle(ctx *core.Context) bool {
	if !r.enabled {
		return false
	}
	
	// 检查全局前置过滤器并保存原始信息
	hasGlobalPreFilters := PreserveOriginalRequestInfoIfNeeded(ctx, r.routerFilters)
	
	// 应用全局前置过滤器
	if len(r.routerFilters) > 0 {
		for _, f := range r.routerFilters {
			if !f.IsEnabled() {
				continue
			}

			// 仅执行前置路由过滤器
			if f.GetAction() == filter.PreRouting {
				if err := f.Apply(ctx); err != nil {
					ctx.AddError(err)
					return false
				}
			}
		}
	}

	// 查找匹配的路由
	route, err := r.findRoute(ctx)
	if err != nil {
		ctx.AddError(err)
		return false
	}

	if route == nil {
		return false
	}

	// 如果全局前置过滤器没有保存原始信息，检查路由过滤器
	if !hasGlobalPreFilters {
		PreserveOriginalRequestInfoIfNeeded(ctx, route.GetRouteFilters())
	}

	// 处理匹配到的路由
	return route.Handle(ctx)
}

// PreserveOriginalRequestInfoIfNeeded 统一的静态方法：检查过滤器并保存原始请求信息
// 返回是否保存了原始信息
func PreserveOriginalRequestInfoIfNeeded(ctx *core.Context, filters []filter.Filter) bool {
	// 检查是否有修改类过滤器
	hasModifiers := HasModificationFilters(filters)
	
	if !hasModifiers {
		return false
	}
	
	// 检查是否需要保存请求头
	needsHeaders := NeedsHeaderPreservation(filters)
	
	// 保存原始信息
	PreserveOriginalRequestInfo(ctx, needsHeaders)
	return true
}

// PreserveOriginalRequestInfo 静态方法：保存原始请求信息到上下文
func PreserveOriginalRequestInfo(ctx *core.Context, needsHeaders bool) {
	req := ctx.Request
	if req == nil {
		return
	}
	
	// 保存原始请求信息到上下文
	ctx.Set(constants.ContextKeyOriginalMethod, req.Method)
	ctx.Set(constants.ContextKeyOriginalURLPath, req.URL.Path)
	ctx.Set(constants.ContextKeyOriginalQueryString, req.URL.RawQuery)
	
	// 保存原始请求头（深拷贝，仅在需要时）
	if needsHeaders {
		originalHeaders := make(map[string][]string)
		for name, values := range req.Header {
			originalHeaders[name] = append([]string{}, values...)
		}
		ctx.Set(constants.ContextKeyOriginalHeaders, originalHeaders)
	}
}

// HasModificationFilters 静态方法：检查过滤器列表是否包含修改类过滤器
func HasModificationFilters(filters []filter.Filter) bool {
	for _, f := range filters {
		if IsModificationFilterType(f.GetType()) {
			return true
		}
	}
	return false
}

// NeedsHeaderPreservation 静态方法：检查是否需要保存请求头
func NeedsHeaderPreservation(filters []filter.Filter) bool {
	for _, f := range filters {
		if f.GetType() == filter.HeaderFilterType {
			return true
		}
	}
	return false
}

// IsModificationFilterType 静态方法：判断过滤器类型是否为修改类
func IsModificationFilterType(filterType filter.FilterType) bool {
	switch filterType {
	case filter.QueryParamFilterType,
		 filter.HeaderFilterType,
		 filter.BodyFilterType,
		 filter.URLFilterType,
		 filter.StripFilterType,
		 filter.RewriteFilterType,
		 filter.MethodFilterType:
		return true
	default:
		return false
	}
}

// findRoute 查找匹配的路由
func (r *Router) findRoute(ctx *core.Context) (RouteHandler, error) {
	if !r.routesSorted {
		r.mu.Lock()
		if !r.routesSorted {
			r.sortRoutes()
		}
		r.mu.Unlock()
	}

	// 遍历路由列表
	for _, route := range r.prioritizedRoutes {
		// 跳过未启用的路由
		if !route.IsEnabled() {
			continue
		}

		// 检查路由是否匹配
		matched, err := route.Match(ctx)
		if err != nil {
			return nil, err
		}

		if matched {
			return route, nil
		}
	}

	return nil, nil
}

// AddRoute 添加路由
func (r *Router) AddRoute(config RouteConfig) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.routes[config.ID]; exists {
		return fmt.Errorf("route with ID %s already exists", config.ID)
	}

	// 设置默认优先级
	if config.Priority == 0 {
		config.Priority = r.config.DefaultPriority
	}

	// 创建路由实例
	route, err := RouteFromConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create route: %w", err)
	}

	// 添加到路由映射
	r.routes[config.ID] = route

	// 添加到优先级路由列表
	r.prioritizedRoutes = append(r.prioritizedRoutes, route)
	r.routesSorted = false

	return nil
}

// RemoveRoute 删除路由
func (r *Router) RemoveRoute(routeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.routes[routeID]; !exists {
		return ErrRouteNotFound
	}

	// 从路由映射中移除
	delete(r.routes, routeID)

	// 从优先级路由列表中移除
	for i, route := range r.prioritizedRoutes {
		if route.GetName() == routeID {
			r.prioritizedRoutes = append(r.prioritizedRoutes[:i], r.prioritizedRoutes[i+1:]...)
			break
		}
	}

	return nil
}

// GetRoute 根据ID获取路由
func (r *Router) GetRoute(routeID string) (RouteHandler, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	route, exists := r.routes[routeID]
	if !exists {
		return nil, ErrRouteNotFound
	}
	return route, nil
}

// ListRoutes 列出所有路由
func (r *Router) ListRoutes() []RouteHandler {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]RouteHandler, 0, len(r.routes))
	for _, route := range r.routes {
		routes = append(routes, route)
	}
	return routes
}

// IsEnabled 是否启用
func (r *Router) IsEnabled() bool {
	return r.enabled
}

// GetName 获取处理器名称
func (r *Router) GetName() string {
	return r.name
}

// Validate 验证配置
func (r *Router) Validate() error {
	if r.config.Name == "" {
		return errors.New("router name cannot be empty")
	}
	return nil
}

// sortRoutes 根据优先级对路由进行排序
func (r *Router) sortRoutes() {
	if r.routesSorted {
		return
	}

	sort.Slice(r.prioritizedRoutes, func(i, j int) bool {
		configI := r.prioritizedRoutes[i].GetConfig()
		configJ := r.prioritizedRoutes[j].GetConfig()
		return configI.Priority < configJ.Priority
	})

	r.routesSorted = true
}

// NewRouterHandler 创建RouterHandler实例
func NewRouterHandler(config RouterConfig) (RouterHandler, error) {
	if config.Name == "" {
		config.Name = "default-router"
	}

	router := NewRouter(config)

	if err := router.Validate(); err != nil {
		return nil, fmt.Errorf("router config validation failed: %w", err)
	}

	// 添加配置的路由
	for _, routeConfig := range config.Routes {
		if err := router.AddRoute(routeConfig); err != nil {
			return nil, fmt.Errorf("failed to add route %s: %w", routeConfig.ID, err)
		}
	}

	// 初始化全局过滤器
	if len(config.FilterConfig) > 0 {
		filterFactory := filter.NewFilterFactory()
		for _, filterConfig := range config.FilterConfig {
			// 创建过滤器实例
			filterInstance, err := filterFactory.CreateFilter(filterConfig)
			if err != nil {
				return nil, fmt.Errorf("create global filter failed: %w", err)
			}
			// 将过滤器添加到切片中
			router.routerFilters = append(router.routerFilters, filterInstance)
		}
	}

	return router, nil
}

// RouterFromConfig 从RouterConfig创建RouterHandler
func RouterFromConfig(config RouterConfig) (RouterHandler, error) {
	return NewRouterHandler(config)
}
