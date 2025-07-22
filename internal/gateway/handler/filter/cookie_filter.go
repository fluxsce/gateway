package filter

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"gohub/internal/gateway/core"
)

// CookieOperation Cookie操作类型
type CookieOperation string

const (
	// AddCookie 添加Cookie
	AddCookie CookieOperation = "add"

	// RemoveCookie 移除Cookie
	RemoveCookie CookieOperation = "remove"

	// ModifyCookie 修改Cookie
	ModifyCookie CookieOperation = "modify"

	// ValidateCookie 验证Cookie
	ValidateCookie CookieOperation = "validate"

	// FilterCookie 过滤Cookie
	FilterCookie CookieOperation = "filter"
)

// CookieFilter Cookie过滤器
// 用于处理HTTP请求和响应中的Cookie
type CookieFilter struct {
	BaseFilter

	// Cookie操作类型
	Operation CookieOperation

	// Cookie名称
	CookieName string

	// Cookie值
	CookieValue string

	// Cookie属性
	CookieAttributes *CookieAttributes

	// 过滤条件
	FilterConditions map[string]interface{}

	// 验证规则
	ValidationRules map[string]interface{}

	// 是否应用到响应阶段
	ApplyToResponse bool
}

// CookieAttributes Cookie属性
type CookieAttributes struct {
	// 域名
	Domain string

	// 路径
	Path string

	// 过期时间
	Expires *time.Time

	// 最大年龄（秒）
	MaxAge int

	// 是否仅HTTPS
	Secure bool

	// 是否仅HTTP
	HttpOnly bool

	// SameSite属性
	SameSite http.SameSite

	// 自定义属性
	Custom map[string]string
}

// CookieFilterFromConfig 从配置创建Cookie过滤器
func CookieFilterFromConfig(config FilterConfig) (Filter, error) {
	action := getFilterActionFromConfig(config)

	// 使用配置中的order字段，如果没有则使用默认值100
	order := config.Order
	if order <= 0 {
		order = 100
	}

	cookieFilter := NewCookieFilter(config.Name, action, order)
	cookieFilter.originalConfig = config

	// 从配置中提取Cookie操作参数
	if err := configureCookieFilter(cookieFilter, config.Config); err != nil {
		return nil, fmt.Errorf("配置Cookie过滤器失败: %w", err)
	}

	return cookieFilter, nil
}

// NewCookieFilter 创建Cookie过滤器
func NewCookieFilter(name string, action FilterAction, priority int) *CookieFilter {
	baseFilter := NewBaseFilter(CookieFilterType, action, priority, true, name)
	return &CookieFilter{
		BaseFilter:       *baseFilter,
		Operation:        AddCookie,
		CookieAttributes: &CookieAttributes{
			Path:     "/",
			MaxAge:   -1, // 默认为会话Cookie
			SameSite: http.SameSiteDefaultMode,
			Custom:   make(map[string]string),
		},
		FilterConditions: make(map[string]interface{}),
		ValidationRules:  make(map[string]interface{}),
		ApplyToResponse:  false,
	}
}

// Apply 实现Filter接口
func (f *CookieFilter) Apply(ctx *core.Context) error {
	if ctx.Request == nil {
		return fmt.Errorf("request is nil")
	}

	// 根据操作类型执行不同的处理
	switch f.Operation {
	case AddCookie:
		return f.addCookie(ctx)
	case RemoveCookie:
		return f.removeCookie(ctx)
	case ModifyCookie:
		return f.modifyCookie(ctx)
	case ValidateCookie:
		return f.validateCookie(ctx)
	case FilterCookie:
		return f.filterCookie(ctx)
	default:
		return fmt.Errorf("未知的Cookie操作类型: %s", f.Operation)
	}
}

// addCookie 添加Cookie
func (f *CookieFilter) addCookie(ctx *core.Context) error {
	if f.CookieName == "" {
		return fmt.Errorf("Cookie名称不能为空")
	}

	// 如果需要在响应阶段添加Cookie，将信息存储在上下文中
	if f.ApplyToResponse {
		f.storeCookieInContext(ctx, "add")
		return nil
	}

	// 在请求阶段添加Cookie到请求头
	cookie := f.buildCookie()
	ctx.Request.AddCookie(cookie)

	// 记录操作
	ctx.Set("cookie_filter_applied", true)
	ctx.Set("cookie_operation", "add")
	ctx.Set("cookie_name", f.CookieName)

	return nil
}

// removeCookie 移除Cookie
func (f *CookieFilter) removeCookie(ctx *core.Context) error {
	if f.CookieName == "" {
		return fmt.Errorf("Cookie名称不能为空")
	}

	// 如果需要在响应阶段移除Cookie，将信息存储在上下文中
	if f.ApplyToResponse {
		f.storeCookieInContext(ctx, "remove")
		return nil
	}

	// 从请求中移除指定的Cookie
	f.removeCookieFromRequest(ctx)

	// 记录操作
	ctx.Set("cookie_filter_applied", true)
	ctx.Set("cookie_operation", "remove")
	ctx.Set("cookie_name", f.CookieName)

	return nil
}

// modifyCookie 修改Cookie
func (f *CookieFilter) modifyCookie(ctx *core.Context) error {
	if f.CookieName == "" {
		return fmt.Errorf("Cookie名称不能为空")
	}

	// 如果需要在响应阶段修改Cookie，将信息存储在上下文中
	if f.ApplyToResponse {
		f.storeCookieInContext(ctx, "modify")
		return nil
	}

	// 在请求阶段修改Cookie
	f.removeCookieFromRequest(ctx)
	if f.CookieValue != "" {
		cookie := f.buildCookie()
		ctx.Request.AddCookie(cookie)
	}

	// 记录操作
	ctx.Set("cookie_filter_applied", true)
	ctx.Set("cookie_operation", "modify")
	ctx.Set("cookie_name", f.CookieName)

	return nil
}

// validateCookie 验证Cookie
func (f *CookieFilter) validateCookie(ctx *core.Context) error {
	if f.CookieName == "" {
		return fmt.Errorf("Cookie名称不能为空")
	}

	// 获取Cookie
	cookie, err := ctx.Request.Cookie(f.CookieName)
	if err != nil {
		if err == http.ErrNoCookie {
			// 检查是否要求Cookie必须存在
			if required, ok := f.ValidationRules["required"].(bool); ok && required {
				return fmt.Errorf("必需的Cookie '%s' 不存在", f.CookieName)
			}
			return nil // Cookie不存在但不是必需的
		}
		return fmt.Errorf("获取Cookie失败: %w", err)
	}

	// 执行验证规则
	if err := f.applyCookieValidationRules(cookie); err != nil {
		ctx.Set("cookie_validation_failed", true)
		ctx.Set("cookie_validation_error", err.Error())
		return err
	}

	// 记录验证成功
	ctx.Set("cookie_filter_applied", true)
	ctx.Set("cookie_operation", "validate")
	ctx.Set("cookie_name", f.CookieName)
	ctx.Set("cookie_validation_passed", true)

	return nil
}

// filterCookie 过滤Cookie
func (f *CookieFilter) filterCookie(ctx *core.Context) error {
	// 获取所有Cookie
	cookies := ctx.Request.Cookies()
	var filteredCookies []*http.Cookie

	// 应用过滤条件
	for _, cookie := range cookies {
		if f.shouldKeepCookie(cookie) {
			filteredCookies = append(filteredCookies, cookie)
		}
	}

	// 重新设置Cookie
	ctx.Request.Header.Del("Cookie")
	for _, cookie := range filteredCookies {
		ctx.Request.AddCookie(cookie)
	}

	// 记录操作
	ctx.Set("cookie_filter_applied", true)
	ctx.Set("cookie_operation", "filter")
	ctx.Set("cookies_filtered_count", len(cookies)-len(filteredCookies))

	return nil
}

// 辅助方法

// buildCookie 构建Cookie对象
func (f *CookieFilter) buildCookie() *http.Cookie {
	cookie := &http.Cookie{
		Name:  f.CookieName,
		Value: f.CookieValue,
	}

	if f.CookieAttributes != nil {
		cookie.Domain = f.CookieAttributes.Domain
		cookie.Path = f.CookieAttributes.Path
		cookie.MaxAge = f.CookieAttributes.MaxAge
		cookie.Secure = f.CookieAttributes.Secure
		cookie.HttpOnly = f.CookieAttributes.HttpOnly
		cookie.SameSite = f.CookieAttributes.SameSite

		if f.CookieAttributes.Expires != nil {
			cookie.Expires = *f.CookieAttributes.Expires
		}
	}

	return cookie
}

// removeCookieFromRequest 从请求中移除Cookie
func (f *CookieFilter) removeCookieFromRequest(ctx *core.Context) {
	cookies := ctx.Request.Cookies()
	var newCookies []*http.Cookie

	// 过滤掉指定名称的Cookie
	for _, cookie := range cookies {
		if cookie.Name != f.CookieName {
			newCookies = append(newCookies, cookie)
		}
	}

	// 重新设置Cookie头
	ctx.Request.Header.Del("Cookie")
	for _, cookie := range newCookies {
		ctx.Request.AddCookie(cookie)
	}
}

// storeCookieInContext 将Cookie操作信息存储在上下文中供响应阶段使用
func (f *CookieFilter) storeCookieInContext(ctx *core.Context, operation string) {
	cookieData := map[string]interface{}{
		"operation":   operation,
		"name":        f.CookieName,
		"value":       f.CookieValue,
		"attributes":  f.CookieAttributes,
	}

	// 获取现有的响应Cookie列表
	var responseCookies []map[string]interface{}
	existingVal, exists := ctx.Get("response_cookies");
	if exists && existingVal != nil {
		if existing, ok := existingVal.([]map[string]interface{}); ok {
			responseCookies = existing
		}
	}

	// 添加新的Cookie操作
	responseCookies = append(responseCookies, cookieData)
	ctx.Set("response_cookies", responseCookies)
}

// applyCookieValidationRules 应用Cookie验证规则
func (f *CookieFilter) applyCookieValidationRules(cookie *http.Cookie) error {
	// 验证值长度
	if maxLen, ok := f.ValidationRules["max_length"].(int); ok {
		if len(cookie.Value) > maxLen {
			return fmt.Errorf("Cookie值长度超过限制: %d", maxLen)
		}
	}

	// 验证值模式
	if pattern, ok := f.ValidationRules["pattern"].(string); ok {
		// 这里可以添加正则表达式验证
		if !strings.Contains(cookie.Value, pattern) {
			return fmt.Errorf("Cookie值不匹配模式: %s", pattern)
		}
	}

	// 验证值范围
	if allowedValues, ok := f.ValidationRules["allowed_values"].([]interface{}); ok {
		allowed := false
		for _, allowedValue := range allowedValues {
			if allowedValueStr, ok := allowedValue.(string); ok && cookie.Value == allowedValueStr {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("Cookie值不在允许范围内")
		}
	}

	// 验证安全属性
	if requireSecure, ok := f.ValidationRules["require_secure"].(bool); ok && requireSecure {
		if !cookie.Secure {
			return fmt.Errorf("Cookie必须设置Secure属性")
		}
	}

	if requireHttpOnly, ok := f.ValidationRules["require_httponly"].(bool); ok && requireHttpOnly {
		if !cookie.HttpOnly {
			return fmt.Errorf("Cookie必须设置HttpOnly属性")
		}
	}

	return nil
}

// shouldKeepCookie 判断是否应该保留Cookie
func (f *CookieFilter) shouldKeepCookie(cookie *http.Cookie) bool {
	// 检查名称过滤
	if allowedNames, ok := f.FilterConditions["allowed_names"].([]interface{}); ok {
		nameAllowed := false
		for _, name := range allowedNames {
			if nameStr, ok := name.(string); ok && cookie.Name == nameStr {
				nameAllowed = true
				break
			}
		}
		if !nameAllowed {
			return false
		}
	}

	// 检查名称前缀过滤
	if allowedPrefixes, ok := f.FilterConditions["allowed_prefixes"].([]interface{}); ok {
		prefixAllowed := false
		for _, prefix := range allowedPrefixes {
			if prefixStr, ok := prefix.(string); ok && strings.HasPrefix(cookie.Name, prefixStr) {
				prefixAllowed = true
				break
			}
		}
		if !prefixAllowed {
			return false
		}
	}

	// 检查拒绝名称
	if deniedNames, ok := f.FilterConditions["denied_names"].([]interface{}); ok {
		for _, name := range deniedNames {
			if nameStr, ok := name.(string); ok && cookie.Name == nameStr {
				return false
			}
		}
	}

	return true
}

// SetOperation 设置操作类型
func (f *CookieFilter) SetOperation(operation CookieOperation) *CookieFilter {
	f.Operation = operation
	return f
}

// SetCookie 设置Cookie名称和值
func (f *CookieFilter) SetCookie(name, value string) *CookieFilter {
	f.CookieName = name
	f.CookieValue = value
	return f
}

// SetAttributes 设置Cookie属性
func (f *CookieFilter) SetAttributes(attrs *CookieAttributes) *CookieFilter {
	f.CookieAttributes = attrs
	return f
}

// SetApplyToResponse 设置是否应用到响应阶段
func (f *CookieFilter) SetApplyToResponse(apply bool) *CookieFilter {
	f.ApplyToResponse = apply
	return f
}

// AddValidationRule 添加验证规则
func (f *CookieFilter) AddValidationRule(key string, value interface{}) *CookieFilter {
	if f.ValidationRules == nil {
		f.ValidationRules = make(map[string]interface{})
	}
	f.ValidationRules[key] = value
	return f
}

// AddFilterCondition 添加过滤条件
func (f *CookieFilter) AddFilterCondition(key string, value interface{}) *CookieFilter {
	if f.FilterConditions == nil {
		f.FilterConditions = make(map[string]interface{})
	}
	f.FilterConditions[key] = value
	return f
}

// configureCookieFilter 配置Cookie过滤器
func configureCookieFilter(cookieFilter *CookieFilter, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 首先检查是否有嵌套的 cookieConfig 配置
	var cookieConfig map[string]interface{}
	if nestedConfig, ok := config["cookieConfig"].(map[string]interface{}); ok {
		cookieConfig = nestedConfig
	} else {
		// 如果没有嵌套配置，直接使用顶级配置
		cookieConfig = config
	}

	// 优先支持前端驼峰命名格式
	if operation, ok := cookieConfig["operation"].(string); ok {
		// 驼峰命名配置处理
		cookieName, _ := cookieConfig["cookieName"].(string)
		cookieValue, _ := cookieConfig["cookieValue"].(string)
		
		// 参数验证
		if operation == "" {
			return fmt.Errorf("operation 不能为空")
		}
		if cookieName == "" {
			return fmt.Errorf("cookieName 不能为空")
		}
		
		// 设置操作类型
		switch strings.ToLower(operation) {
		case "add", "remove", "modify", "validate", "filter":
			cookieFilter.Operation = CookieOperation(strings.ToLower(operation))
		default:
			return fmt.Errorf("无效的operation: %s，支持的类型: add, remove, modify, validate, filter", operation)
		}
		
		// 设置Cookie名称
		cookieFilter.CookieName = cookieName
		
		// 设置Cookie值（可选）
		if cookieValue != "" {
			cookieFilter.CookieValue = cookieValue
		}
		
		// 设置是否应用到响应
		if applyToResponse, ok := cookieConfig["applyToResponse"].(bool); ok {
			cookieFilter.ApplyToResponse = applyToResponse
		}
		
		// 设置Cookie属性
		if attrs, ok := cookieConfig["cookieAttributes"].(map[string]interface{}); ok {
			if err := configureCookieAttributes(cookieFilter.CookieAttributes, attrs); err != nil {
				return fmt.Errorf("配置Cookie属性失败: %w", err)
			}
		}
		
		// 设置验证规则
		if rules, ok := cookieConfig["validationRules"].(map[string]interface{}); ok {
			cookieFilter.ValidationRules = rules
		}
		
		// 设置过滤条件
		if conditions, ok := cookieConfig["filterConditions"].(map[string]interface{}); ok {
			cookieFilter.FilterConditions = conditions
		}
		
		return nil
	}

	// 兼容旧的下划线命名格式
	// 设置操作类型
	if operation, ok := cookieConfig["operation"].(string); ok {
		cookieFilter.Operation = CookieOperation(operation)
	}

	// 设置Cookie名称
	if name, ok := cookieConfig["cookie_name"].(string); ok {
		cookieFilter.CookieName = name
	}

	// 设置Cookie值
	if value, ok := cookieConfig["cookie_value"].(string); ok {
		cookieFilter.CookieValue = value
	}

	// 设置是否应用到响应
	if applyToResponse, ok := cookieConfig["apply_to_response"].(bool); ok {
		cookieFilter.ApplyToResponse = applyToResponse
	}

	// 设置Cookie属性
	if attrs, ok := cookieConfig["cookie_attributes"].(map[string]interface{}); ok {
		if err := configureCookieAttributes(cookieFilter.CookieAttributes, attrs); err != nil {
			return fmt.Errorf("配置Cookie属性失败: %w", err)
		}
	}

	// 设置验证规则
	if rules, ok := cookieConfig["validation_rules"].(map[string]interface{}); ok {
		cookieFilter.ValidationRules = rules
	}

	// 设置过滤条件
	if conditions, ok := cookieConfig["filter_conditions"].(map[string]interface{}); ok {
		cookieFilter.FilterConditions = conditions
	}

	return nil
}

// configureCookieAttributes 配置Cookie属性
func configureCookieAttributes(attrs *CookieAttributes, config map[string]interface{}) error {
	if config == nil {
		return nil
	}

	// 设置域名 - 支持驼峰命名
	if domain, ok := config["domain"].(string); ok {
		attrs.Domain = domain
	}

	// 设置路径 - 支持驼峰命名
	if path, ok := config["path"].(string); ok {
		attrs.Path = path
	}

	// 设置最大年龄 - 支持驼峰命名 (maxAge) 和下划线 (max_age)
	if maxAge, ok := config["maxAge"].(int); ok {
		attrs.MaxAge = maxAge
	} else if maxAgeFloat, ok := config["maxAge"].(float64); ok {
		attrs.MaxAge = int(maxAgeFloat)
	} else if maxAge, ok := config["max_age"].(int); ok {
		attrs.MaxAge = maxAge
	} else if maxAgeFloat, ok := config["max_age"].(float64); ok {
		attrs.MaxAge = int(maxAgeFloat)
	}

	// 设置过期时间 - 支持驼峰命名
	if expires, ok := config["expires"].(string); ok {
		if expireTime, err := time.Parse(time.RFC3339, expires); err == nil {
			attrs.Expires = &expireTime
		}
	}

	// 设置Secure属性 - 支持驼峰命名
	if secure, ok := config["secure"].(bool); ok {
		attrs.Secure = secure
	}

	// 设置HttpOnly属性 - 支持驼峰命名 (httpOnly) 和下划线 (http_only)
	if httpOnly, ok := config["httpOnly"].(bool); ok {
		attrs.HttpOnly = httpOnly
	} else if httpOnly, ok := config["http_only"].(bool); ok {
		attrs.HttpOnly = httpOnly
	}

	// 设置SameSite属性 - 支持驼峰命名 (sameSite) 和下划线 (same_site)
	if sameSite, ok := config["sameSite"].(string); ok {
		switch strings.ToLower(sameSite) {
		case "strict":
			attrs.SameSite = http.SameSiteStrictMode
		case "lax":
			attrs.SameSite = http.SameSiteLaxMode
		case "none":
			attrs.SameSite = http.SameSiteNoneMode
		default:
			attrs.SameSite = http.SameSiteDefaultMode
		}
	} else if sameSiteInt, ok := config["sameSite"].(int); ok {
		attrs.SameSite = http.SameSite(sameSiteInt)
	} else if sameSite, ok := config["same_site"].(string); ok {
		switch strings.ToLower(sameSite) {
		case "strict":
			attrs.SameSite = http.SameSiteStrictMode
		case "lax":
			attrs.SameSite = http.SameSiteLaxMode
		case "none":
			attrs.SameSite = http.SameSiteNoneMode
		default:
			attrs.SameSite = http.SameSiteDefaultMode
		}
	} else if sameSiteInt, ok := config["same_site"].(int); ok {
		attrs.SameSite = http.SameSite(sameSiteInt)
	}

	// 设置自定义属性 - 支持驼峰命名
	if custom, ok := config["custom"].(map[string]interface{}); ok {
		for key, value := range custom {
			if valueStr, ok := value.(string); ok {
				attrs.Custom[key] = valueStr
			} else {
				attrs.Custom[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return nil
} 