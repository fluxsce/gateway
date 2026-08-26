package limiter

import (
	"fmt"
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
//   - 桶状态写入公共 cache，闲置键靠 TTL 过期
//   - 令牌填充速率与容量以当前配置为准，不写入 cache
type TokenBucketLimiter struct {
	*BaseLimiterHandler
	store        *rateLimitCacheStore // 限流键状态，走 pkg/cache
	keyExtractor KeyExtractorFunc     // 限流键提取函数
}

// tokenBucketState 令牌桶在 cache 中的序列化状态。
// 填充速率与桶容量以当前配置为准，不写入 cache，避免热更新后仍按旧参数补令牌。
type tokenBucketState struct {
	Tokens     float64 `json:"t"` // 当前令牌数（0 <= tokens <= burst）
	LastUpdate int64   `json:"u"` // 上次更新时间（UnixNano），用于按时间差补令牌
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
	config = cloneRateLimitConfig(config)

	// 应用默认值（在副本上修改，不会改写调用方或全局 DefaultRateLimitConfig）
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
	store, err := newRateLimitCacheStore(config)
	if err != nil {
		return nil, err
	}

	return &TokenBucketLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		store:              store,
		keyExtractor:       GetKeyExtractor(config.KeyStrategy),
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
//   - 此方法是线程安全的，内部使用存储互斥锁保护读改写
//   - 缓存读写失败时放行，避免限流存储故障拖垮数据面
//   - 令牌计算基于时间差，即使长时间无请求，令牌也会持续累积（最多到容量）
//   - 新桶初始填满令牌，允许立即处理突发请求
//   - 桶状态写入公共 cache；闲置键靠 TTL 过期（过期后下次请求会重新按满桶创建）
func (t *TokenBucketLimiter) checkTokenBucket(key string) bool {
	// 加锁保护同一限流器实例内的 Get-改-Set，避免并发把令牌数写乱
	t.store.mu.Lock()
	defer t.store.mu.Unlock()

	config := t.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	burst := float64(config.Burst)

	var bucket tokenBucketState
	exists, err := t.store.load(key, &bucket)
	if err != nil {
		// 缓存故障时放行：限流是保护能力，存储不可用不应该变成全站 429
		return true
	}
	if !exists {
		// 情况1: 首次请求该限流键（或 TTL 过期后重建）
		// 创建新令牌桶，初始填满令牌（等于burst容量），然后立即消耗一个令牌用于当前请求
		// 这样允许立即处理突发请求，符合令牌桶算法的设计
		bucket.Tokens = burst - 1
		bucket.LastUpdate = now.UnixNano()
		_ = t.store.save(key, &bucket)
		return true
	}

	// 情况2: 限流键已存在，需要按时间差补令牌
	// 例如：经过 0.5 秒，速率 10 令牌/秒，应添加 5 个令牌
	elapsed := now.Sub(time.Unix(0, bucket.LastUpdate)).Seconds()
	// 使用 minFloat64 确保令牌数不超过桶容量（防止溢出）
	bucket.Tokens = minFloat64(burst, bucket.Tokens+elapsed*rate)
	bucket.LastUpdate = now.UnixNano()

	// 检查是否有可用令牌（至少需要1个令牌才能处理请求）
	// 使用 < 1 而不是 <= 0，是为了处理浮点数精度问题
	if bucket.Tokens < 1 {
		// 令牌不足，拒绝请求；仍回写补令牌后的状态，避免下次重复按旧时间补发
		_ = t.store.save(key, &bucket)
		return false
	}

	// 令牌充足，消耗一个令牌并允许请求通过
	bucket.Tokens--
	_ = t.store.save(key, &bucket)
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
