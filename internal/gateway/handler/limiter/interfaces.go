package limiter

import (
	"gateway/internal/gateway/core"
)

// RateLimitAlgorithm 限流算法类型
type RateLimitAlgorithm string

const (
	// AlgorithmFixedWindow 固定窗口算法
	AlgorithmFixedWindow RateLimitAlgorithm = "fixed-window"
	// AlgorithmSlidingWindow 滑动窗口算法
	AlgorithmSlidingWindow RateLimitAlgorithm = "sliding-window"
	// AlgorithmTokenBucket 令牌桶算法
	AlgorithmTokenBucket RateLimitAlgorithm = "token-bucket"
	// AlgorithmLeakyBucket 漏桶算法
	AlgorithmLeakyBucket RateLimitAlgorithm = "leaky-bucket"
	// AlgorithmNone 无限制
	AlgorithmNone RateLimitAlgorithm = "none"
)

// LimiterHandler 限流器处理器接口
// 所有限流器处理器都必须实现此接口
type LimiterHandler interface {
	// Handle 处理限流
	Handle(ctx *core.Context) bool

	// GetAlgorithm 获取限流算法
	GetAlgorithm() RateLimitAlgorithm

	// IsEnabled 是否启用
	IsEnabled() bool

	// GetName 获取处理器名称
	GetName() string

	// Validate 验证配置
	Validate() error

	// GetConfig 获取配置
	GetConfig() *RateLimitConfig

	// OnResponse 处理响应结果（用于更新限流状态）
	OnResponse(ctx *core.Context, err error)
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 基础信息
	ID        string             `yaml:"id" json:"id" mapstructure:"id"`                                                    // 限流配置ID
	Name      string             `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`                // 处理器名称
	Enabled   bool               `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                                     // 是否启用
	Algorithm RateLimitAlgorithm `yaml:"algorithm,omitempty" json:"algorithm,omitempty" mapstructure:"algorithm,omitempty"` // 限流算法

	// 限流参数
	Rate        int    `yaml:"rate" json:"rate" mapstructure:"rate"`                                                       // 限流速率（每秒请求数）
	Burst       int    `yaml:"burst,omitempty" json:"burst,omitempty" mapstructure:"burst,omitempty"`                      // 突发容量
	WindowSize  int    `yaml:"window_size,omitempty" json:"window_size,omitempty" mapstructure:"window_size,omitempty"`    // 时间窗口大小（秒）
	KeyStrategy string `yaml:"key_strategy,omitempty" json:"key_strategy,omitempty" mapstructure:"key_strategy,omitempty"` // 限流Key策略

	// 错误处理
	ErrorStatusCode int    `yaml:"error_status_code,omitempty" json:"error_status_code,omitempty" mapstructure:"error_status_code,omitempty"` // 限流时返回的HTTP状态码
	ErrorMessage    string `yaml:"error_message,omitempty" json:"error_message,omitempty" mapstructure:"error_message,omitempty"`             // 限流时返回的错误信息

	// 扩展配置
	CustomConfig map[string]interface{} `yaml:"custom_config,omitempty" json:"custom_config,omitempty" mapstructure:"custom_config,omitempty"` // 自定义配置
}

// BaseLimiterHandler 限流器处理器基础结构
type BaseLimiterHandler struct {
	config *RateLimitConfig
}

// NewBaseLimiterHandler 创建基础限流器处理器
func NewBaseLimiterHandler(config *RateLimitConfig) *BaseLimiterHandler {
	if config == nil {
		config = &DefaultRateLimitConfig
	}
	return &BaseLimiterHandler{
		config: config,
	}
}

// GetAlgorithm 获取限流算法
func (b *BaseLimiterHandler) GetAlgorithm() RateLimitAlgorithm {
	return b.config.Algorithm
}

// IsEnabled 是否启用
func (b *BaseLimiterHandler) IsEnabled() bool {
	return b.config.Enabled
}

// GetName 获取处理器名称
func (b *BaseLimiterHandler) GetName() string {
	return b.config.Name
}

// GetConfig 获取配置
func (b *BaseLimiterHandler) GetConfig() *RateLimitConfig {
	return b.config
}

// Handle 处理限流（基类默认实现）
func (b *BaseLimiterHandler) Handle(ctx *core.Context) bool {
	return true
}

// Validate 验证配置（基类默认实现）
func (b *BaseLimiterHandler) Validate() error {
	return nil
}

// OnResponse 处理响应结果（基类默认实现）
func (b *BaseLimiterHandler) OnResponse(ctx *core.Context, err error) {
	// 基类默认不执行任何操作
}

// KeyExtractorFunc 键提取函数类型
type KeyExtractorFunc func(*core.Context) string

// GetKeyExtractor 根据策略获取键提取器
func GetKeyExtractor(strategy string) KeyExtractorFunc {
	switch strategy {
	case "ip":
		return ExtractIPKey
	case "user":
		return ExtractUserKey
	case "path":
		return ExtractPathKey
	case "service":
		return ExtractServiceKey
	case "route":
		return ExtractRouteKey
	default:
		return ExtractIPKey
	}
}

// ExtractIPKey 提取IP键
func ExtractIPKey(ctx *core.Context) string {
	ip := ctx.Request.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = ctx.Request.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = ctx.Request.RemoteAddr
	}
	return "ip:" + ip
}

// ExtractUserKey 提取用户键
func ExtractUserKey(ctx *core.Context) string {
	if userID, exists := ctx.Get("user_id"); exists {
		if userIDStr, ok := userID.(string); ok {
			return "user:" + userIDStr
		}
	}
	return ExtractIPKey(ctx)
}

// ExtractPathKey 提取路径键
func ExtractPathKey(ctx *core.Context) string {
	return "path:" + ctx.Request.URL.Path
}

// ExtractServiceKey 提取服务键
func ExtractServiceKey(ctx *core.Context) string {
	serviceID := ctx.GetServiceID()
	if serviceID == "" {
		return ExtractRouteKey(ctx)
	}
	return "service:" + serviceID
}

// ExtractRouteKey 提取路由键
func ExtractRouteKey(ctx *core.Context) string {
	routeID := ctx.GetRouteID()
	if routeID == "" {
		return ExtractPathKey(ctx)
	}
	return "route:" + routeID
}

// 预设配置
var DefaultRateLimitConfig = RateLimitConfig{
	ID:              "default-ratelimit",
	Name:            "Default Rate Limiter",
	Enabled:         true,
	Algorithm:       AlgorithmTokenBucket,
	Rate:            100,
	Burst:           50,
	WindowSize:      1,
	KeyStrategy:     "ip",
	ErrorStatusCode: 429,
	ErrorMessage:    "Rate limit exceeded",
}
