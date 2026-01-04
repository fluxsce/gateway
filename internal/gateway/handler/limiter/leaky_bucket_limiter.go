package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// LeakyBucketLimiter 漏桶限流器
//
// 漏桶算法是一种流量整形算法，通过维护一个漏桶来控制请求速率。
// 算法原理：
//   - 请求像水一样流入桶中（water 增加）
//   - 桶以固定速率（rate）持续漏水（处理请求，water 减少）
//   - 如果桶满了（water >= capacity），新请求会被拒绝
//   - 桶中的水会持续以固定速率漏出，不会累积
//
// 特点：
//   - 平滑输出：无论输入如何，输出速率都是固定的（rate）
//   - 不允许突发：即使桶是空的，输出速率也不会超过 rate
//   - 适合需要严格控制输出速率的场景（如保护下游服务）
//
// 与令牌桶的区别：
//   - 令牌桶：允许突发，桶满时可以快速处理多个请求
//   - 漏桶：不允许突发，输出速率严格限制为 rate
//
// 示例：
//
//	config := &RateLimitConfig{
//	    Rate:        10,        // 每秒处理10个请求（漏出速率）
//	    Burst:       20,        // 桶容量20，最多可排队20个请求
//	    KeyStrategy: "ip",      // 按IP限流
//	}
//	limiter, err := NewLeakyBucketLimiter(config)
//
// 注意：
//   - 此实现使用内存存储漏桶，并自动清理长时间未使用的桶
//   - 清理机制：如果桶的水量为0且超过清理时间阈值，自动删除该桶
//   - 清理时间阈值：max(60秒, 2 * (capacity / rate))，避免频繁创建/删除
type LeakyBucketLimiter struct {
	*BaseLimiterHandler
	buckets      map[string]*leakyBucket // 限流键到漏桶的映射
	mu           sync.Mutex              // 保护buckets的互斥锁
	keyExtractor KeyExtractorFunc        // 限流键提取函数
}

// leakyBucket 漏桶
//
// 记录单个限流键的漏桶状态信息。
// 漏桶以固定速率漏水（处理请求），如果桶满则拒绝新请求。
type leakyBucket struct {
	capacity   int       // 桶容量（最大可容纳的请求数，等于burst）
	water      int       // 当前水量（待处理的请求数，0 <= water <= capacity）
	lastUpdate time.Time // 上次更新时间（用于计算应该漏出的水量）
	rate       float64   // 漏出速率（每秒处理的请求数）
}

// NewLeakyBucketLimiter 创建漏桶限流器
//
// 参数：
//   - config: 限流配置，如果为nil则使用默认配置
//
// 返回：
//   - LimiterHandler: 限流处理器实例
//   - error: 创建过程中的错误
//
// 配置说明：
//   - Rate: 漏出速率，每秒处理的请求数（必须 > 0）
//   - Burst: 桶容量，最大可容纳的请求数（必须 > 0）
//   - KeyStrategy: 限流键策略（ip/user/path等）
//   - ErrorStatusCode: 限流时返回的HTTP状态码
//   - ErrorMessage: 限流时返回的错误消息
//
// 注意：
//   - Burst 必须 >= Rate，否则桶容量太小，无法正常工作
//   - 如果 Burst < Rate，建议在 Validate 方法中检查并报错
func NewLeakyBucketLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	// 应用默认值
	if config.Rate <= 0 {
		config.Rate = DefaultRateLimitConfig.Rate
	}
	if config.Burst <= 0 {
		config.Burst = DefaultRateLimitConfig.Burst
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

	config.Algorithm = AlgorithmLeakyBucket
	keyExtractor := GetKeyExtractor(config.KeyStrategy)

	return &LeakyBucketLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		buckets:            make(map[string]*leakyBucket),
		keyExtractor:       keyExtractor,
	}, nil
}

// Handle 处理漏桶限流
//
// 对请求执行漏桶限流检查。如果桶未满，则允许请求加入桶中（water 增加）；
// 否则拒绝请求并返回错误。桶中的水会持续以固定速率漏出。
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
//   - rate_limit_algorithm: 限流算法（"leaky-bucket"）
func (l *LeakyBucketLimiter) Handle(ctx *core.Context) bool {
	if !l.IsEnabled() {
		return true
	}

	key := l.keyExtractor(ctx)

	if !l.checkLeakyBucket(key) {
		config := l.GetConfig()
		ctx.AddError(fmt.Errorf("leaky bucket rate limit exceeded for key: %s", key))
		ctx.Abort(config.ErrorStatusCode, map[string]string{
			"error": config.ErrorMessage,
		})
		return false
	}

	ctx.Set("rate_limited", false)
	ctx.Set("rate_limit_key", key)
	ctx.Set("rate_limit_algorithm", "leaky-bucket")

	return true
}

// checkLeakyBucket 检查漏桶限流
//
// 核心限流逻辑：
//  1. 如果限流键不存在，创建新漏桶并加入当前请求（water = 1）
//  2. 计算从上次更新到现在应该漏出的水量（基于时间差和漏出速率）
//  3. 更新水量（减少漏出的水量，但不能小于0）
//  4. 如果水量为0且长时间未使用，清理该桶（防止内存泄漏）
//  5. 如果加入新请求后水量 > capacity，拒绝请求
//  6. 否则加入新请求（water++）并允许通过
//
// 参数：
//   - key: 限流键（通过KeyExtractor从请求中提取）
//
// 返回：
//   - bool: true表示允许请求，false表示拒绝请求
//
// 注意：
//   - 此方法是线程安全的，内部使用互斥锁保护共享状态
//   - 漏出计算基于时间差，即使长时间无请求，水量也会持续减少（最多到0）
//   - 新桶初始水量为1，表示第一个请求已加入桶中
//   - 自动清理机制：如果桶的水量为0且超过清理时间阈值，删除该桶
//
// 清理策略：
//   - 清理时间阈值：max(60秒, 2 * (capacity / rate))
//   - 如果水量为0且距离上次更新时间超过阈值，删除该桶
//   - 这样可以防止内存泄漏，同时避免频繁创建/删除桶
func (l *LeakyBucketLimiter) checkLeakyBucket(key string) bool {
	// 加锁保护共享的 buckets map，确保并发安全
	l.mu.Lock()
	defer l.mu.Unlock()

	config := l.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	capacity := config.Burst

	bucket, exists := l.buckets[key]
	if !exists {
		// 情况1: 首次请求该限流键
		// 创建新漏桶，当前请求加入桶中（water = 1）
		// 注意：第一个请求已经占用桶容量（water = 1），直接返回，不执行后续的 water++
		// 桶开始以固定速率漏水（处理请求）
		l.buckets[key] = &leakyBucket{
			capacity:   capacity,
			water:      1, // 第一个请求加入桶中，占用1单位容量
			lastUpdate: now,
			rate:       rate,
		}
		// 当前请求已加入桶中，允许通过
		// 注意：这里直接返回，不执行 water++，因为第一个请求已经在创建桶时加入了
		return true
	}

	// 情况2: 限流键已存在，需要计算漏出量并检查是否可加入新请求
	// 计算从上次更新到现在经过的时间（秒）
	elapsed := now.Sub(bucket.lastUpdate).Seconds()

	// 计算应该漏出的水量：时间差 * 漏出速率
	// 例如：经过 0.5 秒，速率 10 请求/秒，应漏出 5 个单位的水
	leaked := elapsed * bucket.rate

	// 更新水量：当前水量 - 漏出的水量
	// 使用 maxInt(0, ...) 确保水量不会小于0（因为漏出量可能大于当前水量）
	// 例如：water = 3, leaked = 5, 结果 water = 0（不会为负数）
	bucket.water = maxInt(0, bucket.water-int(leaked))

	// 更新最后更新时间，用于下次计算
	bucket.lastUpdate = now

	// 清理机制：如果桶的水量为0且长时间未使用，删除该桶以防止内存泄漏
	// 清理时间阈值：max(60秒, 2 * (capacity / rate))
	// 例如：capacity=20, rate=10, 阈值 = max(60, 2*2) = 60秒
	// 这样可以确保桶在完全空闲后一段时间才被清理
	if bucket.water == 0 {
		// 计算清理时间阈值（秒）
		// 至少60秒，或者2倍的桶清空时间（capacity / rate）
		cleanupThreshold := 60.0 // 默认60秒
		if rate > 0 {
			// 桶清空时间 = capacity / rate（秒）
			emptyTime := float64(capacity) / rate
			// 清理阈值 = 2 * 桶清空时间，但至少60秒
			if emptyTime*2 > cleanupThreshold {
				cleanupThreshold = emptyTime * 2
			}
		}

		// 如果距离上次更新时间超过清理阈值，删除该桶
		if elapsed > cleanupThreshold {
			delete(l.buckets, key)
			// 重新创建桶，当前请求加入
			l.buckets[key] = &leakyBucket{
				capacity:   capacity,
				water:      1,
				lastUpdate: now,
				rate:       rate,
			}
			return true
		}
	}

	// 检查加入新请求后是否会导致桶溢出
	// 使用 > 而不是 >=，因为如果 water = capacity，加入新请求后 water = capacity+1，会溢出
	// 例如：capacity = 10, water = 10, 加入新请求后 water = 11 > 10，应该拒绝
	if bucket.water+1 > capacity {
		// 桶满，拒绝请求
		// 注意：这里不增加水量，因为请求被拒绝了
		return false
	}

	// 桶未满，允许请求通过
	// 加入新请求（增加1单位水量）
	// 注意：第一个请求在创建桶时已经加入（water = 1），后续请求在这里加入（water++）
	// 例如：第一个请求后 water = 1，第二个请求通过后 water = 2，第三个请求通过后 water = 3
	bucket.water++
	return true
}

// Validate 验证配置
//
// 检查漏桶限流器的配置是否合法。
//
// 返回：
//   - error: 配置错误信息，nil表示配置有效
//
// 验证规则：
//   - Rate必须大于0（漏出速率必须为正）
//   - Burst必须大于0（桶容量必须为正）
//
// 注意：
//   - 建议 Burst >= Rate，否则桶容量太小，可能无法正常工作
//   - 如果 Burst < Rate，桶可能很快被填满，导致频繁拒绝请求
func (l *LeakyBucketLimiter) Validate() error {
	config := l.GetConfig()
	if config.Rate <= 0 {
		return fmt.Errorf("漏桶限流速率必须大于0")
	}

	if config.Burst <= 0 {
		return fmt.Errorf("漏桶容量必须大于0")
	}

	// 可选：检查 Burst 是否 >= Rate
	// 如果 Burst < Rate，桶容量太小，可能无法正常工作
	// 但这不是必须的，因为有些场景可能需要更严格的限制
	// if config.Burst < config.Rate {
	//     return fmt.Errorf("漏桶容量应该大于等于漏出速率，建议 Burst >= Rate")
	// }

	return nil
}

// OnResponse 处理响应结果
//
// 漏桶限流器在响应阶段不需要执行额外操作。
// 请求在加入桶时就已经被记录（water 增加），无论请求成功或失败都不会改变水量。
// 桶中的水会持续以固定速率漏出，与请求结果无关。
// 实现此方法以满足LimiterHandler接口要求。
//
// 参数：
//   - ctx: 请求上下文
//   - err: 处理过程中的错误
func (l *LeakyBucketLimiter) OnResponse(ctx *core.Context, err error) {
	// 漏桶限流器通常不需要处理响应结果
	// 请求在加入桶时已记录，不会因为请求失败而减少水量
	// 桶中的水会持续以固定速率漏出，与请求结果无关
}

// maxInt 返回两个int中的较大者
//
// 用于确保水量不会小于0，防止水量为负数。
//
// 参数：
//   - a, b: 要比较的两个整数
//
// 返回：
//   - int: 较大的值
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
