package filter

import (
	"fmt"
	"gateway/internal/gateway/core"
	"strings"
)

// HeaderModifierType 头部修改类型
type HeaderModifierType string

const (
	// AddHeader 添加头部
	AddHeader HeaderModifierType = "add"

	// SetHeader 设置头部（替换现有值）
	SetHeader HeaderModifierType = "set"

	// RemoveHeader 移除头部
	RemoveHeader HeaderModifierType = "remove"

	// RenameHeader 重命名头部
	RenameHeader HeaderModifierType = "rename"
)

// HeaderFilter 请求头过滤器
// 用于修改请求头或响应头
type HeaderFilter struct {
	BaseFilter

	// 修改类型
	ModifierType HeaderModifierType

	// 头部名称
	HeaderName string

	// 头部值
	HeaderValue string

	// 目标头部名称（用于RenameHeader）
	TargetHeaderName string

	// 是否为请求头
	// true=请求头，false=响应头
	IsRequestHeader bool
}

// HeaderFilterFromConfig 从配置创建头部过滤器
func HeaderFilterFromConfig(config FilterConfig) (Filter, error) {
	// 获取过滤器执行时机
	action := getFilterActionFromConfig(config)

	// 根据阶段确定是请求头还是响应头过滤器
	isRequest := action == PostRouting

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	var headerFilter *HeaderFilter
	if isRequest {
		headerFilter = NewRequestHeaderFilter(config.Name, action, order)
	} else {
		headerFilter = NewResponseHeaderFilter(config.Name, action, order)
	}

	// 存储原始配置
	headerFilter.originalConfig = config

	// 从配置中提取头部操作参数
	if err := configureHeaderFilter(headerFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置头部过滤器失败: %w", err)
	}

	return headerFilter, nil
}

// NewRequestHeaderFilter 创建请求头过滤器
func NewRequestHeaderFilter(name string, action FilterAction, priority int) *HeaderFilter {
	baseFilter := NewBaseFilter(HeaderFilterType, action, priority, true, name)
	return &HeaderFilter{
		BaseFilter:      *baseFilter,
		IsRequestHeader: true,
	}
}

// NewResponseHeaderFilter 创建响应头过滤器
func NewResponseHeaderFilter(name string, action FilterAction, priority int) *HeaderFilter {
	baseFilter := NewBaseFilter(HeaderFilterType, action, priority, true, name)
	return &HeaderFilter{
		BaseFilter:      *baseFilter,
		IsRequestHeader: false,
	}
}

// ConfigureAdd 配置为添加头部
func (f *HeaderFilter) ConfigureAdd(headerName, headerValue string) *HeaderFilter {
	f.ModifierType = AddHeader
	f.HeaderName = headerName
	f.HeaderValue = headerValue
	return f
}

// ConfigureSet 配置为设置头部
func (f *HeaderFilter) ConfigureSet(headerName, headerValue string) *HeaderFilter {
	f.ModifierType = SetHeader
	f.HeaderName = headerName
	f.HeaderValue = headerValue
	return f
}

// ConfigureRemove 配置为移除头部
func (f *HeaderFilter) ConfigureRemove(headerName string) *HeaderFilter {
	f.ModifierType = RemoveHeader
	f.HeaderName = headerName
	return f
}

// ConfigureRename 配置为重命名头部
func (f *HeaderFilter) ConfigureRename(headerName, targetHeaderName string) *HeaderFilter {
	f.ModifierType = RenameHeader
	f.HeaderName = headerName
	f.TargetHeaderName = targetHeaderName
	return f
}

// Apply 实现Filter接口
func (f *HeaderFilter) Apply(ctx *core.Context) error {
	// 根据配置修改请求头或响应头
	if f.IsRequestHeader {
		// 修改请求头
		return f.applyToRequest(ctx)
	} else {
		// 修改响应头
		return f.applyToResponse(ctx)
	}
}

// 应用到请求头
func (f *HeaderFilter) applyToRequest(ctx *core.Context) error {
	req := ctx.Request

	switch f.ModifierType {
	case AddHeader:
		// 添加头部（不替换已有值）
		req.Header.Add(f.HeaderName, f.HeaderValue)
	case SetHeader:
		// 设置头部（替换已有值）
		req.Header.Set(f.HeaderName, f.HeaderValue)
	case RemoveHeader:
		// 移除头部
		req.Header.Del(f.HeaderName)
	case RenameHeader:
		// 重命名头部
		if values, ok := req.Header[f.HeaderName]; ok {
			// 复制值到新的头部
			for _, value := range values {
				req.Header.Add(f.TargetHeaderName, value)
			}
			// 删除旧的头部
			req.Header.Del(f.HeaderName)
		}
	}

	return nil
}

// 应用到响应头
func (f *HeaderFilter) applyToResponse(ctx *core.Context) error {
	// 对于响应头，需要修改GinContext中的Writer.Header()
	// 但在这里可能无法直接访问，需要通过Gin Context

	// 从上下文获取响应头
	respHeader, exists := ctx.Get("response_headers")
	if !exists || respHeader == nil {
		// 没有响应头可用，可能是尚未设置
		return nil
	}

	// 将响应头转换为http.Header类型
	header, ok := respHeader.(map[string][]string)
	if !ok {
		// 类型断言失败
		return nil
	}

	switch f.ModifierType {
	case AddHeader:
		// 添加头部（不替换已有值）
		header[f.HeaderName] = append(header[f.HeaderName], f.HeaderValue)
	case SetHeader:
		// 设置头部（替换已有值）
		header[f.HeaderName] = []string{f.HeaderValue}
	case RemoveHeader:
		// 移除头部
		delete(header, f.HeaderName)
	case RenameHeader:
		// 重命名头部
		if values, ok := header[f.HeaderName]; ok {
			// 复制值到新的头部
			header[f.TargetHeaderName] = append(header[f.TargetHeaderName], values...)
			// 删除旧的头部
			delete(header, f.HeaderName)
		}
	}

	// 更新响应头
	ctx.Set("response_headers", header)

	return nil
}

// getFilterActionFromConfig 获取过滤器执行时机
func getFilterActionFromConfig(config FilterConfig) FilterAction {
	// 从配置中获取执行时机，如果没有则根据名称推断
	if config.Action != "" {
		switch strings.ToLower(config.Action) {
		case "pre-routing":
			return PreRouting
		case "post-routing":
			return PostRouting
		case "pre-response":
			return PreResponse
		}
	}

	// 根据名称推断执行时机
	name := strings.ToLower(config.Name)
	id := strings.ToLower(config.ID)

	if containsAny([]string{name, id}, []string{"request", "req", "pre", "before"}) {
		return PostRouting // 请求处理阶段
	}
	if containsAny([]string{name, id}, []string{"response", "resp", "post", "after"}) {
		return PreResponse // 响应处理阶段
	}

	// 默认为路由后执行
	return PostRouting
}

// configureHeaderFilter 配置头部过滤器
func configureHeaderFilter(headerFilter *HeaderFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 首先检查是否有嵌套的 headerConfig 配置
	var headerConfig map[string]interface{}
	if nestedConfig, ok := config["headerConfig"].(map[string]interface{}); ok {
		headerConfig = nestedConfig
	} else {
		// 如果没有嵌套配置，直接使用顶级配置
		headerConfig = config
	}

	// 优先支持前端驼峰命名格式
	if modifierType, ok := headerConfig["modifierType"].(string); ok {
		// 驼峰命名配置处理
		headerName, _ := headerConfig["headerName"].(string)
		headerValue, _ := headerConfig["headerValue"].(string)
		targetHeaderName, _ := headerConfig["targetHeaderName"].(string)

		// 处理isRequestHeader
		var isRequestHeader bool
		if irh, ok := headerConfig["isRequestHeader"].(bool); ok {
			isRequestHeader = irh
		} else if irhStr, ok := headerConfig["isRequestHeader"].(string); ok {
			isRequestHeader = strings.ToLower(irhStr) == "true"
		}

		// 更新过滤器标志
		headerFilter.IsRequestHeader = isRequestHeader

		// 参数验证
		if headerName == "" {
			return fmt.Errorf("headerName 不能为空")
		}

		// 根据操作类型配置
		switch strings.ToLower(modifierType) {
		case "add":
			if headerValue == "" {
				return fmt.Errorf("add操作需要headerValue参数")
			}
			headerFilter.ConfigureAdd(headerName, headerValue)
		case "set":
			if headerValue == "" {
				return fmt.Errorf("set操作需要headerValue参数")
			}
			headerFilter.ConfigureSet(headerName, headerValue)
		case "remove":
			headerFilter.ConfigureRemove(headerName)
		case "rename":
			if targetHeaderName == "" {
				return fmt.Errorf("rename操作需要targetHeaderName参数")
			}
			headerFilter.ConfigureRename(headerName, targetHeaderName)
		default:
			return fmt.Errorf("无效的modifierType: %s", modifierType)
		}

		return nil
	}

	// 兼容旧的下划线命名格式
	if headerName, ok := headerConfig["header_name"].(string); ok {
		if headerValue, ok := headerConfig["header_value"].(string); ok {
			operation := "add"
			if op, ok := headerConfig["operation"].(string); ok {
				operation = strings.ToLower(op)
			}

			switch operation {
			case "add":
				headerFilter.ConfigureAdd(headerName, headerValue)
			case "set":
				headerFilter.ConfigureSet(headerName, headerValue)
			case "remove":
				headerFilter.ConfigureRemove(headerName)
			case "rename":
				if newName, ok := headerConfig["new_name"].(string); ok {
					headerFilter.ConfigureRename(headerName, newName)
				}
			default:
				headerFilter.ConfigureAdd(headerName, headerValue)
			}
		}
	}

	// 批量头部操作配置
	if headers, ok := headerConfig["headers"].(map[string]interface{}); ok {
		for name, value := range headers {
			if valueStr, ok := value.(string); ok {
				headerFilter.ConfigureAdd(name, valueStr)
			}
		}
	}

	return nil
}

// containsAny 检查字符串列表中是否包含任意关键词
func containsAny(texts []string, keywords []string) bool {
	for _, text := range texts {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				return true
			}
		}
	}
	return false
}
