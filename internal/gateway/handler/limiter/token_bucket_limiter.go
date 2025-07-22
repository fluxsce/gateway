package limiter

import (
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/core"
)

// TokenBucketLimiter 令牌桶限流器
type TokenBucketLimiter struct {
	*BaseLimiterHandler
	buckets      map[string]*tokenBucket
	mu           sync.Mutex
	keyExtractor KeyExtractorFunc
}

// tokenBucket 令牌桶
type tokenBucket struct {
	rate       float64   // 每秒填充速率
	capacity   float64   // 桶容量
	tokens     float64   // 当前令牌数
	lastUpdate time.Time // 上次更新时间
}

// NewTokenBucketLimiter 创建令牌桶限流器
func NewTokenBucketLimiter(config *RateLimitConfig) (LimiterHandler, error) {
	if config == nil {
		config = &DefaultRateLimitConfig
	}

	// 应用默认值
	if config.Rate <= 0 {
		config.Rate = DefaultRateLimitConfig.Rate
	}
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
func (t *TokenBucketLimiter) checkTokenBucket(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	config := t.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	burst := float64(config.Burst)

	bucket, exists := t.buckets[key]
	if !exists {
		// 创建新令牌桶，初始填满令牌
		t.buckets[key] = &tokenBucket{
			rate:       rate,
			capacity:   burst,
			tokens:     burst,
			lastUpdate: now,
		}
		return true
	}

	// 计算从上次更新到现在应该添加的令牌数
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	bucket.tokens = minFloat64(bucket.capacity, bucket.tokens+elapsed*bucket.rate)
	bucket.lastUpdate = now

	// 如果令牌不足，拒绝请求
	if bucket.tokens < 1 {
		return false
	}

	// 消耗一个令牌
	bucket.tokens--
	return true
}

// Validate 验证配置
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
func (t *TokenBucketLimiter) OnResponse(ctx *core.Context, err error) {
	// 令牌桶限流器通常不需要处理响应结果
}

// minFloat64 返回两个float64值中较小的一个
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
