package cors

import (
	"net/http"
	"strconv"
	"strings"

	"gohub/internal/gateway/core"
)

// DefaultCORSHandler 默认CORS处理器
type DefaultCORSHandler struct {
	BaseCORSHandler
	// 允许的域名映射(用于快速查找)
	allowOriginsMap map[string]bool
	// 通配符域名
	wildcardOrigins []string
}

// NewDefaultCORSHandler 创建默认CORS处理器
func NewDefaultCORSHandler(config *CORSConfig) *DefaultCORSHandler {
	cfg := DefaultCORSConfig
	if config != nil {
		cfg = *config

		// 应用默认值
		if cfg.Strategy == "" {
			cfg.Strategy = StrategyDefault
		}
		if len(cfg.AllowOrigins) == 0 {
			cfg.AllowOrigins = DefaultCORSConfig.AllowOrigins
		}
		if len(cfg.AllowMethods) == 0 {
			cfg.AllowMethods = DefaultCORSConfig.AllowMethods
		}
		if len(cfg.AllowHeaders) == 0 {
			cfg.AllowHeaders = DefaultCORSConfig.AllowHeaders
		}
		if cfg.MaxAge == 0 {
			cfg.MaxAge = DefaultCORSConfig.MaxAge
		}
	}

	// 处理允许的方法
	for i, method := range cfg.AllowMethods {
		cfg.AllowMethods[i] = strings.ToUpper(method)
	}

	// 创建域名映射
	allowOriginsMap := make(map[string]bool)
	var wildcardOrigins []string

	for _, origin := range cfg.AllowOrigins {
		// 处理带通配符的域名
		if strings.Contains(origin, "*") {
			wildcardOrigins = append(wildcardOrigins, origin)
		} else {
			allowOriginsMap[origin] = true
		}
	}

	return &DefaultCORSHandler{
		BaseCORSHandler: BaseCORSHandler{
			Strategy: cfg.Strategy,
			Enabled:  cfg.Enabled,
			Name:     "Default CORS Handler",
			Config:   cfg,
		},
		allowOriginsMap: allowOriginsMap,
		wildcardOrigins: wildcardOrigins,
	}
}

// Handle 处理CORS请求
func (d *DefaultCORSHandler) Handle(ctx *core.Context) bool {
	origin := ctx.Request.Header.Get("Origin")
	if origin == "" {
		// 不是跨域请求，继续处理
		return true
	}

	// 判断是否为预检请求
	isPreflight := ctx.Request.Method == http.MethodOptions &&
		ctx.Request.Header.Get("Access-Control-Request-Method") != ""

	// 设置CORS响应头
	if err := d.setCORSHeaders(ctx, origin, isPreflight); err != nil {
		// 域名不被允许
		ctx.Abort(http.StatusForbidden, map[string]string{
			"error": "CORS not allowed",
		})
		return false
	}

	// 如果是预检请求，直接返回200
	if isPreflight {
		ctx.Writer.WriteHeader(http.StatusOK)
		return false
	}

	// 继续处理请求
	return true
}

// setCORSHeaders 设置CORS响应头
func (d *DefaultCORSHandler) setCORSHeaders(ctx *core.Context, origin string, isPreflight bool) error {
	// 检查Origin是否被允许
	if !d.isOriginAllowed(origin) {
		return http.ErrNotSupported
	}

	headers := ctx.Writer.Header()

	// 设置允许的Origin
	headers.Set("Access-Control-Allow-Origin", origin)

	// 设置Vary头
	headers.Add("Vary", "Origin")

	// 如果允许携带凭证
	if d.Config.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}

	// 暴露的响应头
	if len(d.Config.ExposeHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(d.Config.ExposeHeaders, ", "))
	}

	// 如果是预检请求，设置更多头信息
	if isPreflight {
		// 设置允许的方法
		headers.Set("Access-Control-Allow-Methods", strings.Join(d.Config.AllowMethods, ", "))

		// 设置允许的头
		if len(d.Config.AllowHeaders) > 0 {
			headers.Set("Access-Control-Allow-Headers", strings.Join(d.Config.AllowHeaders, ", "))
		} else {
			// 如果没有指定，使用请求的头
			requestHeaders := ctx.Request.Header.Get("Access-Control-Request-Headers")
			if requestHeaders != "" {
				headers.Set("Access-Control-Allow-Headers", requestHeaders)
				headers.Add("Vary", "Access-Control-Request-Headers")
			}
		}

		// 设置预检结果缓存时间
		if d.Config.MaxAge > 0 {
			headers.Set("Access-Control-Max-Age", strconv.Itoa(d.Config.MaxAge))
		}
	}

	return nil
}

// isOriginAllowed 检查Origin是否被允许
func (d *DefaultCORSHandler) isOriginAllowed(origin string) bool {
	// 如果允许所有域
	if len(d.Config.AllowOrigins) > 0 && d.Config.AllowOrigins[0] == "*" {
		return true
	}

	// 检查精确匹配
	if d.allowOriginsMap[origin] {
		return true
	}

	// 检查通配符匹配
	for _, wildcardOrigin := range d.wildcardOrigins {
		if matchWildcardOrigin(origin, wildcardOrigin) {
			return true
		}
	}

	return false
}

// Validate 验证配置
func (d *DefaultCORSHandler) Validate() error {
	return ValidateCORSConfig(&d.Config)
}

// ValidateCORSConfig 验证CORS配置
func ValidateCORSConfig(config *CORSConfig) error {
	if config == nil {
		return http.ErrNotSupported
	}

	// 验证允许的域名
	if len(config.AllowOrigins) == 0 {
		return http.ErrNotSupported
	}

	// 验证允许的方法
	if len(config.AllowMethods) == 0 {
		return http.ErrNotSupported
	}

	// 验证MaxAge
	if config.MaxAge < 0 {
		return http.ErrNotSupported
	}

	return nil
}

// matchWildcardOrigin 匹配带通配符的域名
func matchWildcardOrigin(origin, pattern string) bool {
	// 替换通配符为正则表达式
	if strings.HasPrefix(pattern, "*.") {
		// *.example.com 模式
		domain := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(origin, domain)
	}

	return false
}
