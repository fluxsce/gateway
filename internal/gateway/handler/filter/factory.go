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

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	// 根据过滤器ID推断类型
	filterType := f.inferFilterType(config)

	// 根据过滤器类型委托给具体的实现类创建
	switch filterType {
	case HeaderFilterType:
		return HeaderFilterFromConfig(config)
	case QueryParamFilterType:
		return QueryParamFilterFromConfig(config)
	case URLFilterType:
		return URLFilterFromConfig(config)
	case BodyFilterType:
		return BodyFilterFromConfig(config)
	case MethodFilterType:
		return MethodFilterFromConfig(config)
	case CookieFilterType:
		return CookieFilterFromConfig(config)
	case ResponseFilterType:
		return ResponseFilterFromConfig(config)
	default:
		return nil, fmt.Errorf("不支持的过滤器类型: %s", filterType)
	}
}

// inferFilterType 根据配置推断过滤器类型
func (f *FilterFactory) inferFilterType(config FilterConfig) FilterType {
	name := strings.ToLower(config.Name)
	id := strings.ToLower(config.ID)

	// 根据ID或名称中的关键词推断类型
	if f.containsAny([]string{name, id}, []string{"header", "headers"}) {
		return HeaderFilterType
	}
	if f.containsAny([]string{name, id}, []string{"query", "param", "parameter"}) {
		return QueryParamFilterType
	}
	if f.containsAny([]string{name, id}, []string{"path", "url", "rewrite", "strip", "prefix"}) {
		return URLFilterType
	}
	if f.containsAny([]string{name, id}, []string{"body", "content"}) {
		return BodyFilterType
	}
	if f.containsAny([]string{name, id}, []string{"method", "verb"}) {
		return MethodFilterType
	}
	if f.containsAny([]string{name, id}, []string{"cookie", "cookies"}) {
		return CookieFilterType
	}
	if f.containsAny([]string{name, id}, []string{"response", "resp"}) {
		return ResponseFilterType
	}

	// 默认返回头部过滤器类型
	return HeaderFilterType
}

// containsAny 检查字符串列表中是否包含任意关键词
func (f *FilterFactory) containsAny(texts []string, keywords []string) bool {
	for _, text := range texts {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				return true
			}
		}
	}
	return false
}

// GetSupportedFilterTypes 获取支持的过滤器类型列表
func GetSupportedFilterTypes() []FilterType {
	return []FilterType{
		HeaderFilterType,
		QueryParamFilterType,
		URLFilterType,
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
		URLFilterType:        "URL路径过滤器",
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
