package filter

import (
	"fmt"
	"net/http"
	"strings"

	"gohub/internal/gateway/core"
)

// MethodFilterMode 方法过滤模式
type MethodFilterMode string

const (
	// AllowMode 允许模式 - 只允许指定的方法
	AllowMode MethodFilterMode = "allow"

	// DenyMode 拒绝模式 - 拒绝指定的方法
	DenyMode MethodFilterMode = "deny"
)

// MethodFilter HTTP方法过滤器
// 用于限制或验证HTTP请求方法
type MethodFilter struct {
	BaseFilter

	// 过滤模式
	Mode MethodFilterMode

	// 允许的方法列表（AllowMode时使用）
	AllowedMethods []string

	// 拒绝的方法列表（DenyMode时使用）
	DeniedMethods []string

	// 拒绝时的状态码
	RejectStatusCode int

	// 拒绝时的错误消息
	RejectMessage string

	// 是否区分大小写
	CaseSensitive bool

	// 自定义响应头
	CustomHeaders map[string]string
}

// MethodFilterFromConfig 从配置创建方法过滤器
func MethodFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	methodFilter := NewMethodFilter(config.Name, action, order)
	methodFilter.originalConfig = config

	// 从配置中提取方法过滤参数
	if err := configureMethodFilter(methodFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置方法过滤器失败: %w", err)
	}

	return methodFilter, nil
}

// NewMethodFilter 创建方法过滤器
func NewMethodFilter(name string, action FilterAction, priority int) *MethodFilter {
	baseFilter := NewBaseFilter(MethodFilterType, action, priority, true, name)
	return &MethodFilter{
		BaseFilter:       *baseFilter,
		Mode:             AllowMode,
		RejectStatusCode: http.StatusMethodNotAllowed, // 405
		RejectMessage:    "Method not allowed",
		CaseSensitive:    false, // 默认不区分大小写
		CustomHeaders:    make(map[string]string),
	}
}

// Apply 实现Filter接口
func (f *MethodFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 获取请求方法
	method := ctx.Request.Method
	if !f.CaseSensitive {
		method = strings.ToUpper(method)
	}

	// 根据过滤模式进行验证
	var allowed bool
	switch f.Mode {
	case AllowMode:
		allowed = f.isMethodAllowed(method)
	case DenyMode:
		allowed = !f.isMethodDenied(method)
	default:
		// 未知模式，默认允许
		allowed = true
	}

	if !allowed {
		// 方法不被允许，设置错误信息到上下文
		ctx.Set("method_filter_rejected", true)
		ctx.Set("reject_status_code", f.RejectStatusCode)
		ctx.Set("reject_message", f.RejectMessage)
		ctx.Set("allowed_methods", f.getAllowedMethodsString())

		// 设置自定义响应头
		for key, value := range f.CustomHeaders {
			ctx.Set("response_header_"+key, value)
		}

		// 如果是OPTIONS请求，可能是预检请求，返回允许的方法
		if strings.ToUpper(ctx.Request.Method) == "OPTIONS" {
			ctx.Set("response_header_Allow", f.getAllowedMethodsString())
			ctx.Set("handle_options_request", true)
			return nil
		}

		return fmt.Errorf("method %s not allowed, allowed methods: %s", 
			ctx.Request.Method, f.getAllowedMethodsString())
	}

	// 记录过滤器应用信息
	ctx.Set("method_filter_applied", true)
	ctx.Set("method_filter_name", f.Name)

	return nil
}

// isMethodAllowed 检查方法是否被允许（AllowMode）
func (f *MethodFilter) isMethodAllowed(method string) bool {
	if len(f.AllowedMethods) == 0 {
		return true // 如果没有配置允许的方法，默认允许所有方法
	}

	normalizedMethod := method
	if !f.CaseSensitive {
		normalizedMethod = strings.ToUpper(method)
	}

	for _, allowedMethod := range f.AllowedMethods {
		if !f.CaseSensitive {
			allowedMethod = strings.ToUpper(allowedMethod)
		}
		if normalizedMethod == allowedMethod {
			return true
		}
	}
	return false
}

// isMethodDenied 检查方法是否被拒绝（DenyMode）
func (f *MethodFilter) isMethodDenied(method string) bool {
	if len(f.DeniedMethods) == 0 {
		return false // 如果没有配置拒绝的方法，默认不拒绝任何方法
	}

	normalizedMethod := method
	if !f.CaseSensitive {
		normalizedMethod = strings.ToUpper(method)
	}

	for _, deniedMethod := range f.DeniedMethods {
		if !f.CaseSensitive {
			deniedMethod = strings.ToUpper(deniedMethod)
		}
		if normalizedMethod == deniedMethod {
			return true
		}
	}
	return false
}

// getAllowedMethodsString 获取允许的方法字符串
func (f *MethodFilter) getAllowedMethodsString() string {
	switch f.Mode {
	case AllowMode:
		if len(f.AllowedMethods) > 0 {
			return strings.Join(f.AllowedMethods, ", ")
		}
		return "ALL"
	case DenyMode:
		// 在拒绝模式下，返回所有标准HTTP方法除了被拒绝的
		standardMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
		var allowed []string
		for _, method := range standardMethods {
			if !f.isMethodDenied(method) {
				allowed = append(allowed, method)
			}
		}
		return strings.Join(allowed, ", ")
	default:
		return "ALL"
	}
}

// ConfigureAllowMethods 配置允许的方法
func (f *MethodFilter) ConfigureAllowMethods(methods []string) *MethodFilter {
	f.Mode = AllowMode
	f.AllowedMethods = make([]string, len(methods))
	for i, method := range methods {
		if f.CaseSensitive {
			f.AllowedMethods[i] = method
		} else {
			f.AllowedMethods[i] = strings.ToUpper(method)
		}
	}
	return f
}

// ConfigureDenyMethods 配置拒绝的方法
func (f *MethodFilter) ConfigureDenyMethods(methods []string) *MethodFilter {
	f.Mode = DenyMode
	f.DeniedMethods = make([]string, len(methods))
	for i, method := range methods {
		if f.CaseSensitive {
			f.DeniedMethods[i] = method
		} else {
			f.DeniedMethods[i] = strings.ToUpper(method)
		}
	}
	return f
}

// SetRejectResponse 设置拒绝时的响应
func (f *MethodFilter) SetRejectResponse(statusCode int, message string) *MethodFilter {
	f.RejectStatusCode = statusCode
	f.RejectMessage = message
	return f
}

// SetCaseSensitive 设置是否区分大小写
func (f *MethodFilter) SetCaseSensitive(caseSensitive bool) *MethodFilter {
	f.CaseSensitive = caseSensitive
	// 如果改变了大小写敏感性，需要重新规范化已有的方法列表
	if !caseSensitive {
		f.normalizeMethodLists()
	}
	return f
}

// AddCustomHeader 添加自定义响应头
func (f *MethodFilter) AddCustomHeader(key, value string) *MethodFilter {
	if f.CustomHeaders == nil {
		f.CustomHeaders = make(map[string]string)
	}
	f.CustomHeaders[key] = value
	return f
}

// normalizeMethodLists 规范化方法列表（转为大写）
func (f *MethodFilter) normalizeMethodLists() {
	for i, method := range f.AllowedMethods {
		f.AllowedMethods[i] = strings.ToUpper(method)
	}
	for i, method := range f.DeniedMethods {
		f.DeniedMethods[i] = strings.ToUpper(method)
	}
}

// GetAllowedMethods 获取允许的方法列表
func (f *MethodFilter) GetAllowedMethods() []string {
	return f.AllowedMethods
}

// GetDeniedMethods 获取拒绝的方法列表
func (f *MethodFilter) GetDeniedMethods() []string {
	return f.DeniedMethods
}

// IsMethodSupported 检查方法是否被支持
func (f *MethodFilter) IsMethodSupported(method string) bool {
	switch f.Mode {
	case AllowMode:
		return f.isMethodAllowed(method)
	case DenyMode:
		return !f.isMethodDenied(method)
	default:
		return true
	}
}

// configureMethodFilter 配置方法过滤器
func configureMethodFilter(methodFilter *MethodFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 设置过滤模式
	if mode, ok := config["mode"].(string); ok {
		switch strings.ToLower(mode) {
		case "allow":
			methodFilter.Mode = AllowMode
		case "deny":
			methodFilter.Mode = DenyMode
		default:
			methodFilter.Mode = AllowMode
		}
	}

	// 设置允许的方法
	if methods, ok := config["allowed_methods"].([]interface{}); ok {
		methodFilter.AllowedMethods = make([]string, len(methods))
		for i, method := range methods {
			if methodStr, ok := method.(string); ok {
				if methodFilter.CaseSensitive {
					methodFilter.AllowedMethods[i] = methodStr
				} else {
					methodFilter.AllowedMethods[i] = strings.ToUpper(methodStr)
				}
			}
		}
	} else if methodsStr, ok := config["allowed_methods"].([]string); ok {
		methodFilter.AllowedMethods = make([]string, len(methodsStr))
		for i, method := range methodsStr {
			if methodFilter.CaseSensitive {
				methodFilter.AllowedMethods[i] = method
			} else {
				methodFilter.AllowedMethods[i] = strings.ToUpper(method)
			}
		}
	}

	// 设置拒绝的方法
	if methods, ok := config["denied_methods"].([]interface{}); ok {
		methodFilter.DeniedMethods = make([]string, len(methods))
		for i, method := range methods {
			if methodStr, ok := method.(string); ok {
				if methodFilter.CaseSensitive {
					methodFilter.DeniedMethods[i] = methodStr
				} else {
					methodFilter.DeniedMethods[i] = strings.ToUpper(methodStr)
				}
			}
		}
	} else if methodsStr, ok := config["denied_methods"].([]string); ok {
		methodFilter.DeniedMethods = make([]string, len(methodsStr))
		for i, method := range methodsStr {
			if methodFilter.CaseSensitive {
				methodFilter.DeniedMethods[i] = method
			} else {
				methodFilter.DeniedMethods[i] = strings.ToUpper(method)
			}
		}
	}

	// 设置拒绝状态码
	if status, ok := config["reject_status"].(int); ok {
		methodFilter.RejectStatusCode = status
	} else if statusFloat, ok := config["reject_status"].(float64); ok {
		methodFilter.RejectStatusCode = int(statusFloat)
	}

	// 设置拒绝消息
	if message, ok := config["reject_message"].(string); ok {
		methodFilter.RejectMessage = message
	}

	// 设置大小写敏感性
	if caseSensitive, ok := config["case_sensitive"].(bool); ok {
		methodFilter.SetCaseSensitive(caseSensitive)
	}

	// 设置自定义响应头
	if headers, ok := config["custom_headers"].(map[string]interface{}); ok {
		for key, value := range headers {
			if valueStr, ok := value.(string); ok {
				methodFilter.AddCustomHeader(key, valueStr)
			}
		}
	}

	return nil
} 