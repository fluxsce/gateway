package limiter

import (
	"fmt"
	"sync"
	"time"

	"gohub/internal/gateway/core"
)

// LeakyBucketLimiter 漏桶限流器
type LeakyBucketLimiter struct {
	*BaseLimiterHandler
	buckets      map[string]*leakyBucket
	mu           sync.Mutex
	keyExtractor KeyExtractorFunc
}

// leakyBucket 漏桶
type leakyBucket struct {
	capacity   int       // 桶容量
	water      int       // 当前水量
	lastUpdate time.Time // 上次更新时间
	rate       float64   // 漏出速率（每秒）
}

// NewLeakyBucketLimiter 创建漏桶限流器
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
func (l *LeakyBucketLimiter) checkLeakyBucket(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	config := l.GetConfig()
	now := time.Now()
	rate := float64(config.Rate)
	capacity := config.Burst

	bucket, exists := l.buckets[key]
	if !exists {
		// 创建新漏桶
		l.buckets[key] = &leakyBucket{
			capacity:   capacity,
			water:      1, // 第一个请求直接加入
			lastUpdate: now,
			rate:       rate,
		}
		return true
	}

	// 计算泄漏量
	elapsed := now.Sub(bucket.lastUpdate).Seconds()
	leaked := elapsed * bucket.rate
	bucket.water = maxInt(0, bucket.water-int(leaked))
	bucket.lastUpdate = now

	// 检查加入新请求后是否溢出
	if bucket.water+1 > capacity {
		return false // 桶满，拒绝请求
	}

	// 加入新请求（增加1单位水量）
	bucket.water++
	return true
}

// Validate 验证配置
func (l *LeakyBucketLimiter) Validate() error {
	config := l.GetConfig()
	if config.Rate <= 0 {
		return fmt.Errorf("漏桶限流速率必须大于0")
	}

	if config.Burst <= 0 {
		return fmt.Errorf("漏桶容量必须大于0")
	}

	return nil
}

// OnResponse 处理响应结果
func (l *LeakyBucketLimiter) OnResponse(ctx *core.Context, err error) {
	// 漏桶限流器通常不需要处理响应结果
}

// maxInt 返回两个int中的较大者
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
