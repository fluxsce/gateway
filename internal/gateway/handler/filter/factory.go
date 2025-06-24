package filter

import (
	"fmt"
	"strings"
)

// FilterFactory 过滤器工厂
type FilterFactory struct{}

// NewFilterFactory 创建过滤器工厂
func NewFilterFactory() *FilterFactory {
	return &FilterFactory{}
}

// CreateFilter 根据配置创建过滤器
func (f *FilterFactory) CreateFilter(config FilterConfig) (Filter, error) {
	if config.ID == "" {
		return nil, fmt.Errorf("过滤器ID不能为空")
	}

	if config.Name == "" {
		config.Name = config.ID
	}

	// 验证过滤器类型是否明确指定
	if config.Type == "" {
		return nil, fmt.Errorf("过滤器类型不能为空，必须明确指定")
	}

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	// 根据明确指定的过滤器类型创建对应的过滤器
	switch FilterType(config.Type) {
	case HeaderFilterType:
		return HeaderFilterFromConfig(config)
	case QueryParamFilterType:
		return QueryParamFilterFromConfig(config)
	case URLFilterType:
		// URL过滤器根据子类型分发到具体的过滤器
		return createURLFilter(config)
	case StripFilterType:
		return StripPrefixFilterFromConfig(config)
	case RewriteFilterType:
		return PathRewriteFilterFromConfig(config)
	case BodyFilterType:
		return BodyFilterFromConfig(config)
	case MethodFilterType:
		return MethodFilterFromConfig(config)
	case CookieFilterType:
		return CookieFilterFromConfig(config)
	case ResponseFilterType:
		return ResponseFilterFromConfig(config)
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", config.Type)
	}
}

// createURLFilter 创建URL过滤器（根据子类型分发）
func createURLFilter(config FilterConfig) (Filter, error) {
	// 从配置中获取URL过滤器的子类型
	subType := "rewrite" // 默认为路径重写
	if config.Config != nil {
		if st, ok := config.Config["sub_type"].(string); ok {
			subType = strings.ToLower(st)
		}
	}

	// 根据子类型创建对应的过滤器
	switch subType {
	case "strip", "strip-prefix":
		return StripPrefixFilterFromConfig(config)
	case "rewrite", "path-rewrite":
		return PathRewriteFilterFromConfig(config)
	default:
		// 默认创建路径重写过滤器
		return PathRewriteFilterFromConfig(config)
	}
}

// GetSupportedFilterTypes 获取支持的过滤器类型列表
func GetSupportedFilterTypes() []FilterType {
	return []FilterType{
		HeaderFilterType,
		QueryParamFilterType,
		URLFilterType,
		StripFilterType,
		RewriteFilterType,
		BodyFilterType,
		MethodFilterType,
		CookieFilterType,
		ResponseFilterType,
	}
}

// GetFilterTypeDescription 获取过滤器类型描述
func GetFilterTypeDescription(filterType FilterType) string {
	descriptions := map[FilterType]string{
		HeaderFilterType:     "请求头/响应头过滤器",
		QueryParamFilterType: "查询参数过滤器",
		URLFilterType:        "URL路径过滤器（通用）",
		StripFilterType:      "前缀剥离过滤器",
		RewriteFilterType:    "路径重写过滤器",
		BodyFilterType:       "请求体过滤器",
		MethodFilterType:     "HTTP方法过滤器",
		CookieFilterType:     "Cookie过滤器",
		ResponseFilterType:   "响应过滤器",
	}

	if desc, exists := descriptions[filterType]; exists {
		return desc
	}
	return "未知过滤器类型"
}
