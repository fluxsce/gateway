package limiter

import (
	"fmt"
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
//   - 桶状态写入公共 cache，闲置键靠 TTL 过期
//   - 漏出速率与容量以当前配置为准，不写入 cache
type LeakyBucketLimiter struct {
	*BaseLimiterHandler
	store        *rateLimitCacheStore // 限流键状态，走 pkg/cache
	keyExtractor KeyExtractorFunc     // 限流键提取函数
}

// leakyBucketState 漏桶在 cache 中的序列化状态。
// 漏出速率与桶容量以当前配置为准，不写入 cache。
type leakyBucketState struct {
	Water      int   `json:"w"` // 当前水量（待处理请求数，0 <= water <= burst）
	LastUpdate int64 `json:"u"` // 上次更新时间（UnixNano），用于按时间差漏水
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
	config = cloneRateLimitConfig(config)

	// 应用默认值（在副本上修改，不会改写调用方或全局 DefaultRateLimitConfig）
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
	store, err := newRateLimitCacheStore(config)
	if err != nil {
		return nil, err
	}

	return &LeakyBucketLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		store:              store,
		keyExtractor:       GetKeyExtractor(config.KeyStrategy),
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
//  4. 闲置桶由 cache TTL 过期清理，不再在访问路径上 delete
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
//   - 此方法是线程安全的，内部使用存储互斥锁保护读改写
//   - 缓存读写失败时放行，避免限流存储故障拖垮数据面
//   - 漏出计算基于时间差，即使长时间无请求，水量也会持续减少（最多到0）
func (l *LeakyBucketLimiter) checkLeakyBucket(key string) bool {
	// 加锁保护同一限流器实例内的 Get-改-Set，避免并发把水量写乱
	l.store.mu.Lock()
	defer l.store.mu.Unlock()

	config := l.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	capacity := config.Burst

	var bucket leakyBucketState
	exists, err := l.store.load(key, &bucket)
	if err != nil {
		// 缓存故障时放行：限流是保护能力，存储不可用不应该变成全站 429
		return true
	}
	if !exists {
		// 情况1: 首次请求该限流键（或 TTL 过期后重建）
		// 创建新漏桶，当前请求加入桶中（water = 1）
		// 第一个请求已经占用桶容量，直接返回，不执行后续的 water++
		bucket.Water = 1
		bucket.LastUpdate = now.UnixNano()
		_ = l.store.save(key, &bucket)
		return true
	}

	// 情况2: 限流键已存在，先按时间差漏水再决定能否加入新请求
	elapsed := now.Sub(time.Unix(0, bucket.LastUpdate)).Seconds()
	// 漏出水量 = 时间差 * 漏出速率；例如经过 0.5 秒、速率 10 请求/秒，应漏出 5
	// 使用 maxInt(0, ...) 确保水量不会小于0（漏出量可能大于当前水量）
	bucket.Water = maxInt(0, bucket.Water-int(elapsed*rate))
	bucket.LastUpdate = now.UnixNano()

	// 使用 > 而不是 >=：water = capacity 时再加入会变成 capacity+1，应当拒绝
	if bucket.Water+1 > capacity {
		// 桶满，拒绝请求；仍回写漏水后的状态，避免下次按过旧时间重复漏水
		_ = l.store.save(key, &bucket)
		return false
	}

	// 桶未满，允许请求通过并加入新请求（water++）
	bucket.Water++
	_ = l.store.save(key, &bucket)
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
