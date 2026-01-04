package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// SlidingWindowLimiter 滑动窗口限流器
//
// 滑动窗口算法是一种精确的限流算法，通过维护一个滑动的时间窗口来控制请求速率。
// 算法原理：
//   - 维护一个固定大小的时间窗口（如60秒）
//   - 记录窗口内所有请求的时间戳
//   - 每次请求时，清理窗口外的过期时间戳
//   - 如果窗口内时间戳数量 < Rate，允许请求并记录时间戳
//   - 否则拒绝请求
//
// 特点：
//   - 精确限流：避免了固定窗口的边界突刺问题
//   - 平滑控制：窗口持续滑动，流量控制更平滑
//   - 内存占用：需要存储窗口内所有请求的时间戳
//   - 性能开销：每次请求需要清理过期时间戳（O(n)复杂度）
//
// 示例：
//
//	config := &RateLimitConfig{
//	    Rate:       100,        // 窗口内最多100个请求
//	    WindowSize: 60,         // 窗口大小60秒
//	    KeyStrategy: "ip",      // 按IP限流
//	}
//	limiter, err := NewSlidingWindowLimiter(config)
//
// 注意：
//   - 此实现使用内存存储时间戳列表，并自动清理长时间未使用的窗口
//   - 清理机制：如果窗口的时间戳列表为空且超过清理时间阈值，自动删除该窗口
//   - 清理时间阈值：max(60秒, 2 * WindowSize)，避免频繁创建/删除窗口
//   - 每次请求都需要遍历时间戳列表清理过期项，高并发场景下可能影响性能
//   - 如果 Rate 很大，考虑使用更高效的实现（如分片窗口、近似算法等）
type SlidingWindowLimiter struct {
	*BaseLimiterHandler
	windows      map[string]*slidingWindow // 限流键到滑动窗口的映射
	mu           sync.Mutex                // 保护windows的互斥锁
	keyExtractor KeyExtractorFunc          // 限流键提取函数
}

// slidingWindow 滑动窗口
//
// 记录单个限流键在滑动时间窗口内的请求时间戳列表。
// 时间戳列表用于统计窗口内的请求数量，并自动清理过期的时间戳。
type slidingWindow struct {
	timestamps []time.Time // 窗口内请求时间戳列表（按时间顺序）
	lastUpdate time.Time   // 上次更新时间（用于清理长时间未使用的窗口）
}

// NewSlidingWindowLimiter 创建滑动窗口限流器
//
// 参数：
//   - config: 限流配置，如果为nil则使用默认配置
//
// 返回：
//   - LimiterHandler: 限流处理器实例
//   - error: 创建过程中的错误
//
// 配置说明：
//   - Rate: 时间窗口内允许的最大请求数（必须 > 0）
//   - WindowSize: 时间窗口大小（秒，必须 > 0）
//   - KeyStrategy: 限流键策略（ip/user/path等）
//   - ErrorStatusCode: 限流时返回的HTTP状态码
//   - ErrorMessage: 限流时返回的错误消息
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
//
// 对请求执行滑动窗口限流检查。如果窗口内请求数未超过限制，
// 则允许请求通过并记录时间戳；否则拒绝请求并返回错误。
//
// 参数：
//   - ctx: 请求上下文
//
// 返回：
//   - bool: true表示请求通过限流检查，false表示被限流
//
// 上下文设置：
//   - rate_limited: 是否被限流（false）
//   - rate_limit_key: 限流键
//   - rate_limit_algorithm: 限流算法（"sliding-window"）
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
//
// 核心限流逻辑：
//  1. 如果限流键不存在，创建新窗口并记录当前时间戳
//  2. 清理窗口外的过期时间戳（时间 < now - windowSize）
//  3. 如果窗口内时间戳数量 >= Rate，拒绝请求
//  4. 否则添加当前时间戳并允许请求通过
//
// 参数：
//   - key: 限流键（通过KeyExtractor从请求中提取）
//
// 返回：
//   - bool: true表示允许请求，false表示拒绝请求
//
// 注意：
//   - 此方法是线程安全的，内部使用互斥锁保护共享状态
//   - 每次请求都需要遍历时间戳列表清理过期项，时间复杂度 O(n)
//   - 如果 Rate 很大，时间戳列表会很长，可能影响性能
//   - 时间戳列表按时间顺序存储，清理时只需要保留窗口内的时间戳
//
// 性能优化建议：
//   - 如果时间戳列表很长，可以考虑使用二分查找优化清理过程
//   - 或者使用分片窗口、近似算法等更高效的实现
func (s *SlidingWindowLimiter) checkSlidingWindow(key string) bool {
	// 加锁保护共享的 windows map，确保并发安全
	s.mu.Lock()
	defer s.mu.Unlock()

	config := s.GetConfig()
	now := time.Now()
	windowSize := time.Duration(config.WindowSize) * time.Second

	window, exists := s.windows[key]
	if !exists {
		// 情况1: 首次请求该限流键
		// 创建新窗口，记录当前请求的时间戳
		// 当前请求是窗口内的第一个请求，允许通过
		s.windows[key] = &slidingWindow{
			timestamps: []time.Time{now},
			lastUpdate: now,
		}
		return true
	}

	// 情况2: 限流键已存在，需要清理过期时间戳并检查限制
	// 计算窗口的截止时间：当前时间 - 窗口大小
	// 例如：now = 100s, windowSize = 60s, cutoff = 40s
	// 只有时间戳 > 40s 的请求才在窗口内
	cutoff := now.Add(-windowSize)

	// 清理过期时间戳：只保留窗口内的时间戳（时间戳 > cutoff）
	// 使用预分配容量避免多次内存分配
	validTimestamps := make([]time.Time, 0, len(window.timestamps))
	for _, ts := range window.timestamps {
		// 只保留在窗口内的时间戳（时间戳在 cutoff 之后）
		if ts.After(cutoff) {
			validTimestamps = append(validTimestamps, ts)
		}
		// 注意：如果时间戳列表是按时间顺序的，可以使用二分查找优化
		// 但当前实现是简单遍历，对于小到中等规模的列表已经足够
	}
	window.timestamps = validTimestamps

	// 清理机制：如果窗口的时间戳列表为空且长时间未使用，删除该窗口以防止内存泄漏
	// 清理时间阈值：max(60秒, 2 * WindowSize)
	// 这样可以确保窗口在完全空闲后一段时间才被清理
	if len(window.timestamps) == 0 {
		// 计算清理时间阈值（秒）
		// 至少60秒，或者2倍的窗口大小
		cleanupThreshold := 60.0 // 默认60秒
		windowSizeSeconds := float64(config.WindowSize)
		// 清理阈值 = 2 * 窗口大小，但至少60秒
		if windowSizeSeconds*2 > cleanupThreshold {
			cleanupThreshold = windowSizeSeconds * 2
		}

		// 计算距离上次更新的时间（使用旧的 lastUpdate，因为当前时间戳列表为空）
		elapsed := now.Sub(window.lastUpdate).Seconds()

		// 如果距离上次更新时间超过清理阈值，删除该窗口
		if elapsed > cleanupThreshold {
			delete(s.windows, key)
			// 重新创建窗口，当前请求加入
			s.windows[key] = &slidingWindow{
				timestamps: []time.Time{now},
				lastUpdate: now,
			}
			return true
		}
		// 如果未达到清理阈值，保留窗口但更新 lastUpdate（虽然时间戳列表为空，但窗口仍在使用）
		window.lastUpdate = now
	} else {
		// 时间戳列表不为空，更新 lastUpdate
		window.lastUpdate = now
	}

	// 检查窗口内的时间戳数量是否已达到限制
	// 使用 >= 是因为如果已经有 Rate 个时间戳，当前请求是第 (Rate+1) 个，应该被拒绝
	// 例如：Rate = 100，窗口内有 100 个时间戳，当前请求会被拒绝
	if len(window.timestamps) >= config.Rate {
		// 已达到速率限制，拒绝请求
		// 注意：这里不添加当前时间戳，因为请求被拒绝了
		return false
	}

	// 窗口内时间戳数量 < Rate，允许请求通过
	// 添加当前请求的时间戳到窗口内
	window.timestamps = append(window.timestamps, now)
	return true
}

// Validate 验证配置
//
// 检查滑动窗口限流器的配置是否合法。
//
// 返回：
//   - error: 配置错误信息，nil表示配置有效
//
// 验证规则：
//   - Rate必须大于0（窗口内允许的最大请求数必须为正）
//   - WindowSize必须大于0（时间窗口大小必须为正）
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
//
// 滑动窗口限流器在响应阶段不需要执行额外操作。
// 时间戳在请求处理前就已经记录，无论请求成功或失败都不会移除时间戳。
// 实现此方法以满足LimiterHandler接口要求。
//
// 参数：
//   - ctx: 请求上下文
//   - err: 处理过程中的错误
func (s *SlidingWindowLimiter) OnResponse(ctx *core.Context, err error) {
	// 滑动窗口限流器通常不需要处理响应结果
	// 时间戳在请求处理前已记录，不会因为请求失败而移除
}
