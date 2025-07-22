package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// FixedWindowLimiter 固定窗口限流器
type FixedWindowLimiter struct {
	*BaseLimiterHandler
	counters     map[string]*fixedWindowCounter
	mu           sync.Mutex
	keyExtractor KeyExtractorFunc
}

// fixedWindowCounter 固定窗口计数器
type fixedWindowCounter struct {
	count     int       // 当前窗口请求计数
	startTime time.Time // 窗口开始时间
}

// NewFixedWindowLimiter 创建固定窗口限流器
func NewFixedWindowLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	// 应用默认值
	if config.Rate <= 0 {
		config.Rate = DefaultRateLimitConfig.Rate
	}
	if config.WindowSize <= 0 {
		config.WindowSize = DefaultRateLimitConfig.WindowSize
	}
	if config.KeyStrategy == "" {
		config.KeyStrategy = DefaultRateLimitConfig.KeyStrategy
	}
	if config.ErrorStatusCode == 0 {
		config.ErrorStatusCode = DefaultRateLimitConfig.ErrorStatusCode
	}
	if config.ErrorMessage == "" {
		config.ErrorMessage = DefaultRateLimitConfig.ErrorMessage
	}

	config.Algorithm = AlgorithmFixedWindow
	keyExtractor := GetKeyExtractor(config.KeyStrategy)

	return &FixedWindowLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		counters:           make(map[string]*fixedWindowCounter),
		keyExtractor:       keyExtractor,
	}, nil
}

// Handle 处理固定窗口限流
func (f *FixedWindowLimiter) Handle(ctx *core.Context) bool {
	if !f.IsEnabled() {
		return true
	}

	key := f.keyExtractor(ctx)

	if !f.checkFixedWindow(key) {
		config := f.GetConfig()
		ctx.AddError(fmt.Errorf("fixed window rate limit exceeded for key: %s", key))
		ctx.Abort(config.ErrorStatusCode, map[string]string{
			"error": config.ErrorMessage,
		})
		return false
	}

	ctx.Set("rate_limited", false)
	ctx.Set("rate_limit_key", key)
	ctx.Set("rate_limit_algorithm", "fixed-window")

	return true
}

// checkFixedWindow 检查固定窗口限流
func (f *FixedWindowLimiter) checkFixedWindow(key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	config := f.GetConfig()
	now := time.Now()
	windowSize := time.Duration(config.WindowSize) * time.Second

	counter, exists := f.counters[key]
	if !exists || now.Sub(counter.startTime) >= windowSize {
		// 创建新窗口或窗口已过期，重置计数器
		f.counters[key] = &fixedWindowCounter{
			count:     1,
			startTime: now,
		}
		return true
	}

	// 检查是否超过速率限制
	if counter.count >= config.Rate {
		return false
	}

	// 增加计数
	counter.count++
	return true
}

// Validate 验证配置
func (f *FixedWindowLimiter) Validate() error {
	config := f.GetConfig()
	if config.Rate <= 0 {
		return fmt.Errorf("固定窗口限流速率必须大于0")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("固定窗口时间窗口大小必须大于0")
	}

	return nil
}

// OnResponse 处理响应结果
func (f *FixedWindowLimiter) OnResponse(ctx *core.Context, err error) {
	// 固定窗口限流器通常不需要处理响应结果
}
