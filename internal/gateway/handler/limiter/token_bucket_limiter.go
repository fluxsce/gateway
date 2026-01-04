package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// TokenBucketLimiter 令牌桶限流器
//
// 令牌桶算法是一种经典的限流算法，通过维护一个令牌桶来控制请求速率。
// 算法原理：
//   - 令牌以固定速率（rate）持续添加到桶中
//   - 桶有最大容量（capacity = burst），令牌数不会超过容量
//   - 每个请求需要消耗一个令牌才能通过
//   - 如果桶中没有令牌，请求会被拒绝
//
// 特点：
//   - 允许突发流量：桶满时可以处理 burst 个请求
//   - 平滑限流：令牌持续添加，不会出现窗口边界突刺
//   - 适合需要允许短期突发的场景
//
// 示例：
//
//	config := &RateLimitConfig{
//	    Rate:        10,        // 每秒填充10个令牌
//	    Burst:       20,        // 桶容量20，允许突发20个请求
//	    KeyStrategy: "ip",      // 按IP限流
//	}
//	limiter, err := NewTokenBucketLimiter(config)
//
// 注意：
//   - 此实现使用内存存储令牌桶，并自动清理长时间未使用的桶
//   - 清理机制：如果桶的令牌数达到容量（已满）且超过清理时间阈值，自动删除该桶
//   - 清理时间阈值：max(60秒, 2 * (capacity / rate))，避免频繁创建/删除桶
type TokenBucketLimiter struct {
	*BaseLimiterHandler
	buckets      map[string]*tokenBucket // 限流键到令牌桶的映射
	mu           sync.Mutex              // 保护buckets的互斥锁
	keyExtractor KeyExtractorFunc        // 限流键提取函数
}

// tokenBucket 令牌桶
//
// 记录单个限流键的令牌桶状态信息。
type tokenBucket struct {
	rate       float64   // 每秒填充速率（令牌/秒）
	capacity   float64   // 桶容量（最大令牌数，等于burst）
	tokens     float64   // 当前令牌数（0 <= tokens <= capacity）
	lastUpdate time.Time // 上次更新时间（用于计算应该添加的令牌数）
}

// NewTokenBucketLimiter 创建令牌桶限流器
//
// 参数：
//   - config: 限流配置，如果为nil则使用默认配置
//
// 返回：
//   - LimiterHandler: 限流处理器实例
//   - error: 创建过程中的错误
//
// 配置说明：
//   - Rate: 每秒填充的令牌数（必须 > 0）
//   - Burst: 桶容量，允许的突发请求数（默认值为 Rate/2，必须 >= 0）
//   - KeyStrategy: 限流键策略（ip/user/path等）
//   - ErrorStatusCode: 限流时返回的HTTP状态码
//   - ErrorMessage: 限流时返回的错误消息
//
// 默认值处理：
//   - 如果 Burst <= 0，自动设置为 Rate/2（如果仍 <= 0，则使用默认值）
//   - 这样确保桶至少能容纳一些突发请求
func NewTokenBucketLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	// 应用默认值
	if config.Rate <= 0 {
		config.Rate = DefaultRateLimitConfig.Rate
	}
	// Burst 默认值处理：如果未设置或 <= 0，使用 Rate/2
	// 这确保桶至少能容纳一些突发请求，但不会过大
	if config.Burst <= 0 {
		config.Burst = config.Rate / 2
		if config.Burst <= 0 {
			config.Burst = DefaultRateLimitConfig.Burst
		}
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

	config.Algorithm = AlgorithmTokenBucket
	keyExtractor := GetKeyExtractor(config.KeyStrategy)

	return &TokenBucketLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		buckets:            make(map[string]*tokenBucket),
		keyExtractor:       keyExtractor,
	}, nil
}

// Handle 处理令牌桶限流
//
// 对请求执行令牌桶限流检查。如果桶中有可用令牌，则消耗一个令牌并允许请求通过；
// 否则拒绝请求并返回错误。
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
//   - rate_limit_algorithm: 限流算法（"token-bucket"）
func (t *TokenBucketLimiter) Handle(ctx *core.Context) bool {
	if !t.IsEnabled() {
		return true
	}

	key := t.keyExtractor(ctx)

	if !t.checkTokenBucket(key) {
		config := t.GetConfig()
		ctx.AddError(fmt.Errorf("token bucket rate limit exceeded for key: %s", key))
		ctx.Abort(config.ErrorStatusCode, map[string]string{
			"error": config.ErrorMessage,
		})
		return false
	}

	ctx.Set("rate_limited", false)
	ctx.Set("rate_limit_key", key)
	ctx.Set("rate_limit_algorithm", "token-bucket")

	return true
}

// checkTokenBucket 检查令牌桶限流
//
// 核心限流逻辑：
//  1. 如果限流键不存在，创建新令牌桶并初始填满令牌（允许突发）
//  2. 计算从上次更新到现在应该添加的令牌数（基于时间差和填充速率）
//  3. 更新令牌数（不超过桶容量）
//  4. 如果令牌数 < 1，拒绝请求
//  5. 否则消耗一个令牌并允许请求通过
//
// 参数：
//   - key: 限流键（通过KeyExtractor从请求中提取）
//
// 返回：
//   - bool: true表示允许请求，false表示拒绝请求
//
// 注意：
//   - 此方法是线程安全的，内部使用互斥锁保护共享状态
//   - 令牌计算基于时间差，即使长时间无请求，令牌也会持续累积（最多到容量）
//   - 新桶初始填满令牌，允许立即处理突发请求
//
// 清理机制：
//   - 如果桶的令牌数达到容量（已满）且距离上次更新时间超过清理阈值，自动删除该桶
//   - 清理阈值：max(60秒, 2 * (capacity / rate))，确保桶在完全空闲后一段时间才被清理
//   - 这样不会影响限流准确性，因为重新创建桶时会初始填满令牌
func (t *TokenBucketLimiter) checkTokenBucket(key string) bool {
	// 加锁保护共享的 buckets map，确保并发安全
	t.mu.Lock()
	defer t.mu.Unlock()

	config := t.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	burst := float64(config.Burst)

	bucket, exists := t.buckets[key]
	if !exists {
		// 情况1: 首次请求该限流键
		// 创建新令牌桶，初始填满令牌（等于burst容量）
		// 然后立即消耗一个令牌用于当前请求
		// 这样允许立即处理突发请求，符合令牌桶算法的设计
		bucket = &tokenBucket{
			rate:       rate,
			capacity:   burst,
			tokens:     burst - 1, // 初始填满，但当前请求消耗一个令牌
			lastUpdate: now,
		}
		t.buckets[key] = bucket
		// 当前请求已消耗一个令牌，允许通过
		return true
	}

	// 情况2: 限流键已存在，需要更新令牌数
	// 计算从上次更新到现在经过的时间（秒）
	elapsed := now.Sub(bucket.lastUpdate).Seconds()

	// 计算应该添加的令牌数：时间差 * 填充速率
	// 例如：经过 0.5 秒，速率 10 令牌/秒，应添加 5 个令牌
	// 使用 minFloat64 确保令牌数不超过桶容量（防止溢出）
	bucket.tokens = minFloat64(bucket.capacity, bucket.tokens+elapsed*bucket.rate)

	// 更新最后更新时间，用于下次计算
	bucket.lastUpdate = now

	// 清理机制：如果桶的令牌数达到容量（已满）且长时间未使用，删除该桶以防止内存泄漏
	// 清理时间阈值：max(60秒, 2 * (capacity / rate))
	// 例如：capacity=20, rate=10, 阈值 = max(60, 2*2) = 60秒
	// 这样可以确保桶在完全空闲（令牌已满，说明长时间未使用）后一段时间才被清理
	if bucket.tokens >= bucket.capacity {
		// 计算清理时间阈值（秒）
		// 至少60秒，或者2倍的桶填满时间（capacity / rate）
		cleanupThreshold := 60.0 // 默认60秒
		if bucket.rate > 0 {
			// 桶填满时间 = capacity / rate（秒）
			// 即：从空桶到满桶需要的时间
			fillTime := bucket.capacity / bucket.rate
			// 清理阈值 = 2 * 桶填满时间，但至少60秒
			if fillTime*2 > cleanupThreshold {
				cleanupThreshold = fillTime * 2
			}
		}

		// 如果距离上次更新时间超过清理阈值，删除该桶
		if elapsed > cleanupThreshold {
			delete(t.buckets, key)
			// 重新创建桶，初始填满令牌，当前请求消耗一个令牌
			t.buckets[key] = &tokenBucket{
				rate:       rate,
				capacity:   burst,
				tokens:     burst - 1, // 初始填满，但当前请求消耗一个令牌
				lastUpdate: now,
			}
			return true
		}
	}

	// 检查是否有可用令牌（至少需要1个令牌才能处理请求）
	// 注意：这里使用 < 1 而不是 <= 0，是为了处理浮点数精度问题
	if bucket.tokens < 1 {
		// 令牌不足，拒绝请求
		return false
	}

	// 令牌充足，消耗一个令牌并允许请求通过
	bucket.tokens--
	return true
}

// Validate 验证配置
//
// 检查令牌桶限流器的配置是否合法。
//
// 返回：
//   - error: 配置错误信息，nil表示配置有效
//
// 验证规则：
//   - Rate必须大于0（令牌填充速率必须为正）
//   - Burst必须大于等于0（桶容量不能为负，0表示不允许突发）
func (t *TokenBucketLimiter) Validate() error {
	config := t.GetConfig()
	if config.Rate <= 0 {
		return fmt.Errorf("令牌桶限流速率必须大于0")
	}

	if config.Burst < 0 {
		return fmt.Errorf("令牌桶突发流量不能为负数")
	}

	return nil
}

// OnResponse 处理响应结果
//
// 令牌桶限流器在响应阶段不需要执行额外操作。
// 令牌在请求处理前就已经消耗，无论请求成功或失败都不会返还令牌。
// 实现此方法以满足LimiterHandler接口要求。
//
// 参数：
//   - ctx: 请求上下文
//   - err: 处理过程中的错误
func (t *TokenBucketLimiter) OnResponse(ctx *core.Context, err error) {
	// 令牌桶限流器通常不需要处理响应结果
	// 令牌在请求处理前已消耗，不会因为请求失败而返还
}

// minFloat64 返回两个float64值中较小的一个
//
// 用于确保令牌数不超过桶容量，防止令牌溢出。
//
// 参数：
//   - a, b: 要比较的两个浮点数
//
// 返回：
//   - float64: 较小的值
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
