package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// SlidingWindowLimiter 滑动窗口限流器
type SlidingWindowLimiter struct {
	*BaseLimiterHandler
	windows      map[string]*slidingWindow
	mu           sync.Mutex
	keyExtractor KeyExtractorFunc
}

// slidingWindow 滑动窗口
type slidingWindow struct {
	timestamps []time.Time // 请求时间戳列表
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
func NewSlidingWindowLimiter(config *RateLimitConfig) (LimiterHandler, error) {
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

	config.Algorithm = AlgorithmSlidingWindow
	keyExtractor := GetKeyExtractor(config.KeyStrategy)

	return &SlidingWindowLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		windows:            make(map[string]*slidingWindow),
		keyExtractor:       keyExtractor,
	}, nil
}

// Handle 处理滑动窗口限流
func (s *SlidingWindowLimiter) Handle(ctx *core.Context) bool {
	if !s.IsEnabled() {
		return true
	}

	key := s.keyExtractor(ctx)

	if !s.checkSlidingWindow(key) {
		config := s.GetConfig()
		ctx.AddError(fmt.Errorf("sliding window rate limit exceeded for key: %s", key))
		ctx.Abort(config.ErrorStatusCode, map[string]string{
			"error": config.ErrorMessage,
		})
		return false
	}

	ctx.Set("rate_limited", false)
	ctx.Set("rate_limit_key", key)
	ctx.Set("rate_limit_algorithm", "sliding-window")

	return true
}

// checkSlidingWindow 检查滑动窗口限流
func (s *SlidingWindowLimiter) checkSlidingWindow(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.GetConfig()
	now := time.Now()
	windowSize := time.Duration(config.WindowSize) * time.Second

	window, exists := s.windows[key]
	if !exists {
		// 创建新窗口
		s.windows[key] = &slidingWindow{
			timestamps: []time.Time{now},
		}
		return true
	}

	// 清理过期时间戳
	cutoff := now.Add(-windowSize)
	validTimestamps := make([]time.Time, 0, len(window.timestamps))
	for _, ts := range window.timestamps {
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
	}
	window.timestamps = validTimestamps

	// 检查是否超过速率限制
	if len(window.timestamps) >= config.Rate {
		return false
	}

	// 添加当前时间戳
	window.timestamps = append(window.timestamps, now)
	return true
}

// Validate 验证配置
func (s *SlidingWindowLimiter) Validate() error {
	config := s.GetConfig()
	if config.Rate <= 0 {
		return fmt.Errorf("滑动窗口限流速率必须大于0")
	}

	if config.WindowSize <= 0 {
		return fmt.Errorf("滑动窗口时间窗口大小必须大于0")
	}

	return nil
}

// OnResponse 处理响应结果
func (s *SlidingWindowLimiter) OnResponse(ctx *core.Context, err error) {
	// 滑动窗口限流器通常不需要处理响应结果
}
