package router

import (
	"fmt"
	"gateway/internal/gateway/handler/filter"
)

// RouterHandlerFactory 路由处理器工厂
type RouterHandlerFactory struct{}

// NewRouterHandlerFactory 创建路由处理器工厂
func NewRouterHandlerFactory() *RouterHandlerFactory {
	return &RouterHandlerFactory{}
}

// CreateRouter 根据配置创建路由处理器
func (f *RouterHandlerFactory) CreateRouter(config RouterConfig) (RouterHandler, error) {
	// 验证配置
	if err := f.validateConfig(config); err != nil {
		return nil, fmt.Errorf("路由器配置验证失败: %w", err)
	}

	// 应用默认配置
	f.applyDefaults(&config)

	// 使用现有的 RouterFromConfig 方法创建路由处理器
	return RouterFromConfig(config)
}

// CreateDefaultRouter 创建默认路由处理器
func (f *RouterHandlerFactory) CreateDefaultRouter() (RouterHandler, error) {
	return f.CreateRouter(DefaultRouterConfig)
}

// CreateRouterWithID 创建指定ID的路由处理器
func (f *RouterHandlerFactory) CreateRouterWithID(id string, enabled bool) (RouterHandler, error) {
	config := DefaultRouterConfig
	config.ID = id
	config.Enabled = enabled
	if config.Name == "" || config.Name == DefaultRouterConfig.Name {
		config.Name = id
	}
	return f.CreateRouter(config)
}

// CreateEnabledRouter 创建启用的路由处理器
func (f *RouterHandlerFactory) CreateEnabledRouter(config RouterConfig) (RouterHandler, error) {
	config.Enabled = true
	return f.CreateRouter(config)
}

// CreateDisabledRouter 创建禁用的路由处理器
func (f *RouterHandlerFactory) CreateDisabledRouter(config RouterConfig) (RouterHandler, error) {
	config.Enabled = false
	return f.CreateRouter(config)
}

// validateConfig 验证路由器配置
func (f *RouterHandlerFactory) validateConfig(config RouterConfig) error {
	if config.ID == "" {
		return fmt.Errorf("路由器ID不能为空")
	}

	if config.Name == "" {
		return fmt.Errorf("路由器名称不能为空")
	}

	// 验证优先级设置
	if config.DefaultPriority <= 0 {
		return fmt.Errorf("默认优先级必须大于0")
	}

	// 验证路由缓存TTL
	if config.EnableRouteCache && config.RouteCacheTTL <= 0 {
		return fmt.Errorf("启用路由缓存时，缓存TTL必须大于0")
	}

	return nil
}

// applyDefaults 应用默认配置
func (f *RouterHandlerFactory) applyDefaults(config *RouterConfig) {
	if config.ID == "" {
		config.ID = DefaultRouterConfig.ID
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	if config.DefaultPriority <= 0 {
		config.DefaultPriority = DefaultRouterConfig.DefaultPriority
	}

	if config.RouteCacheTTL <= 0 {
		config.RouteCacheTTL = DefaultRouterConfig.RouteCacheTTL
	}

	if config.Routes == nil {
		config.Routes = []RouteConfig{}
	}

	if config.FilterConfig == nil {
		config.FilterConfig = []filter.FilterConfig{}
	}
}

// GetSupportedFeatures 获取支持的路由特性
func (f *RouterHandlerFactory) GetSupportedFeatures() []string {
	return []string{
		"路由匹配",
		"优先级排序",
		"动态路由管理",
		"过滤器链",
		"路由缓存",
	}
}

// GetFeatureDescription 获取特性描述
func (f *RouterHandlerFactory) GetFeatureDescription(feature string) string {
	descriptions := map[string]string{
		"路由匹配":   "根据请求路径、方法等条件匹配路由",
		"优先级排序":  "根据优先级对路由进行排序，优先处理高优先级路由",
		"动态路由管理": "支持动态添加、删除路由",
		"过滤器链":   "支持全局和路由级别的过滤器",
		"路由缓存":   "支持路由匹配结果缓存，提高性能",
	}

	if desc, exists := descriptions[feature]; exists {
		return desc
	}
	return "未知路由特性"
}
