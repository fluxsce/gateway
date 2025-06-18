package filter

import (
	"fmt"
	"strings"
)

// URLFilterFromConfig 从配置创建URL过滤器
func URLFilterFromConfig(config FilterConfig) (Filter, error) {
	name := strings.ToLower(config.Name)
	id := strings.ToLower(config.ID)

	// 判断是前缀剥离还是路径重写
	if containsAny([]string{name, id}, []string{"strip", "prefix"}) {
		return StripPrefixFilterFromConfig(config)
	} else {
		return PathRewriteFilterFromConfig(config)
	}
}

// StripPrefixFilterFromConfig 从配置创建前缀剥离过滤器
func StripPrefixFilterFromConfig(config FilterConfig) (Filter, error) {
	prefix := ""
	if config.Config != nil {
		if p, ok := config.Config["prefix"].(string); ok {
			prefix = p
		}
	}

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	filter := NewStripPrefixFilter(config.Name, prefix, order)

	// 存储原始配置
	filter.BaseFilter.originalConfig = config

	return filter, nil
}

// PathRewriteFilterFromConfig 从配置创建路径重写过滤器
func PathRewriteFilterFromConfig(config FilterConfig) (Filter, error) {
	from := ""
	to := ""
	mode := "simple"

	if config.Config != nil {
		if f, ok := config.Config["from"].(string); ok {
			from = f
		}
		if t, ok := config.Config["to"].(string); ok {
			to = t
		}
		if m, ok := config.Config["mode"].(string); ok {
			mode = m
		}
	}

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	var filter *PathRewriteFilter
	var err error

	if mode == "regex" {
		filter, err = NewRegexPathRewriteFilter(config.Name, from, to, order)
		if err != nil {
			return nil, fmt.Errorf("创建正则路径重写过滤器失败: %w", err)
		}
	} else {
		filter = NewSimplePathRewriteFilter(config.Name, from, to, order)
	}

	// 存储原始配置
	filter.BaseFilter.originalConfig = config

	return filter, nil
}

// BodyFilterFromConfig 从配置创建请求体过滤器
func BodyFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	baseFilter := NewBaseFilter(BodyFilterType, action, order, config.Enabled, config.Name)
	baseFilter.originalConfig = config

	return baseFilter, nil
}

// MethodFilterFromConfig 从配置创建方法过滤器
func MethodFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	baseFilter := NewBaseFilter(MethodFilterType, action, order, config.Enabled, config.Name)
	baseFilter.originalConfig = config

	return baseFilter, nil
}

// CookieFilterFromConfig 从配置创建Cookie过滤器
func CookieFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	baseFilter := NewBaseFilter(CookieFilterType, action, order, config.Enabled, config.Name)
	baseFilter.originalConfig = config

	return baseFilter, nil
}

// ResponseFilterFromConfig 从配置创建响应过滤器
func ResponseFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	baseFilter := NewBaseFilter(ResponseFilterType, action, order, config.Enabled, config.Name)
	baseFilter.originalConfig = config

	return baseFilter, nil
}
