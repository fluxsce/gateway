package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// FixedWindowLimiter 固定窗口限流器
//
// 固定窗口算法将时间划分为固定大小的窗口，每个窗口内统计请求数量。
// 当窗口内请求数超过限制时，拒绝后续请求，直到进入下一个时间窗口。
//
// 特点：
//   - 实现简单，内存占用小
//   - 时间窗口边界可能出现流量突刺（临界问题）
//   - 适合对流量控制精度要求不高的场景
//   - 自动清理长时间未使用的计数器，防止内存泄漏
//
// 示例：
//
//	config := &RateLimitConfig{
//	    Rate:       100,        // 每个窗口最多100个请求
//	    WindowSize: 60,         // 窗口大小60秒
//	    KeyStrategy: "ip",      // 按IP限流
//	}
//	limiter, err := NewFixedWindowLimiter(config)
type FixedWindowLimiter struct {
	*BaseLimiterHandler
	counters     map[string]*fixedWindowCounter // 限流键到计数器的映射
	mu           sync.Mutex                     // 保护counters的互斥锁
	keyExtractor KeyExtractorFunc               // 限流键提取函数
}

// fixedWindowCounter 固定窗口计数器
//
// 记录单个限流键在当前时间窗口内的请求统计信息。
type fixedWindowCounter struct {
	count      int       // 当前窗口请求计数
	startTime  time.Time // 窗口开始时间
	lastUpdate time.Time // 上次更新时间（用于清理长时间未使用的计数器）
}

// NewFixedWindowLimiter 创建固定窗口限流器
//
// 参数：
//   - config: 限流配置，如果为nil则使用默认配置
//
// 返回：
//   - LimiterHandler: 限流处理器实例
//   - error: 创建过程中的错误
//
// 配置说明：
//   - Rate: 时间窗口内允许的最大请求数
//   - WindowSize: 时间窗口大小（秒）
//   - KeyStrategy: 限流键策略（ip/user/path等）
//   - ErrorStatusCode: 限流时返回的HTTP状态码
//   - ErrorMessage: 限流时返回的错误消息
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
//
// 对请求执行固定窗口限流检查。如果当前窗口内请求数未超过限制，
// 则允许请求通过并增加计数；否则拒绝请求并返回错误。
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
//   - rate_limit_algorithm: 限流算法（"fixed-window"）
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
//
// 核心限流逻辑：
//  1. 如果限流键不存在或当前窗口已过期，创建新窗口
//  2. 如果当前窗口内请求数已达限制，拒绝请求
//  3. 否则增加计数并允许请求通过
//
// 参数：
//   - key: 限流键（通过KeyExtractor从请求中提取）
//
// 返回：
//   - bool: true表示允许请求，false表示拒绝请求
//
// 注意：此方法是线程安全的，内部使用互斥锁保护共享状态。
func (f *FixedWindowLimiter) checkFixedWindow(key string) bool {
	// 加锁保护共享的 counters map，确保并发安全
	f.mu.Lock()
	defer f.mu.Unlock()

	// 获取限流配置
	config := f.GetConfig()
	now := time.Now()
	// 计算窗口大小（将秒转换为 time.Duration）
	windowSize := time.Duration(config.WindowSize) * time.Second

	// 尝试获取该 key 对应的计数器
	counter, exists := f.counters[key]

	// 情况1: 计数器不存在（首次请求）或窗口已过期
	// 判断条件: counter 不存在 或 当前时间距离窗口开始时间 >= 窗口大小
	if !exists || now.Sub(counter.startTime) >= windowSize {
		// 如果计数器存在但窗口已过期，检查是否需要清理
		if exists {
			// 清理机制：如果窗口已过期且长时间未使用，删除该计数器以防止内存泄漏
			// 清理时间阈值：max(60秒, 2 * WindowSize)
			// 这样可以确保计数器在完全空闲后一段时间才被清理
			elapsed := now.Sub(counter.lastUpdate).Seconds()
			cleanupThreshold := 60.0 // 默认60秒
			windowSizeSeconds := float64(config.WindowSize)
			// 清理阈值 = 2 * 窗口大小，但至少60秒
			if windowSizeSeconds*2 > cleanupThreshold {
				cleanupThreshold = windowSizeSeconds * 2
			}

			// 如果距离上次更新时间超过清理阈值，删除该计数器
			if elapsed > cleanupThreshold {
				delete(f.counters, key)
				// 重新创建计数器，当前请求计入新窗口的第一个请求
				f.counters[key] = &fixedWindowCounter{
					count:      1,   // 当前请求计入新窗口的第一个请求
					startTime:  now, // 记录新窗口的开始时间
					lastUpdate: now, // 记录最后更新时间
				}
				return true
			}
		}

		// 创建新窗口，计数从1开始（因为当前请求算作第一个）
		// 注意: 这里直接返回 true，表示当前请求被允许
		f.counters[key] = &fixedWindowCounter{
			count:      1,   // 当前请求计入新窗口的第一个请求
			startTime:  now, // 记录新窗口的开始时间
			lastUpdate: now, // 记录最后更新时间
		}
		return true
	}

	// 情况2: 计数器存在且窗口未过期
	// 更新最后更新时间
	counter.lastUpdate = now

	// 检查当前窗口内的请求数是否已达到限制
	// 使用 >= 是因为 count 从1开始计数
	// 例如: Rate=100 时，count 可以从 1 到 100，当 count=100 时，下一个请求会被拒绝
	if counter.count >= config.Rate {
		return false // 已达到速率限制，拒绝请求
	}

	// 情况3: 窗口未过期且未达到限制
	// 增加计数并允许请求通过
	counter.count++
	return true
}

// Validate 验证配置
//
// 检查固定窗口限流器的配置是否合法。
//
// 返回：
//   - error: 配置错误信息，nil表示配置有效
//
// 验证规则：
//   - Rate必须大于0
//   - WindowSize必须大于0
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
//
// 固定窗口限流器在响应阶段不需要执行额外操作。
// 实现此方法以满足LimiterHandler接口要求。
//
// 参数：
//   - ctx: 请求上下文
//   - err: 处理过程中的错误
func (f *FixedWindowLimiter) OnResponse(ctx *core.Context, err error) {
	// 固定窗口限流器通常不需要处理响应结果
}
