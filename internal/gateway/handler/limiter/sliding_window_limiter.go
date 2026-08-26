package limiter

import (
	"fmt"
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
//   - 窗口状态写入公共 cache，闲置键靠 TTL 过期，避免本地 map 只在再次访问时才删除
//   - 每次请求都需要遍历时间戳列表清理过期项，高并发场景下可能影响性能
//   - 如果 Rate 很大，考虑使用更高效的实现（如分片窗口、近似算法等）
type SlidingWindowLimiter struct {
	*BaseLimiterHandler
	store        *rateLimitCacheStore // 限流键状态，走 pkg/cache
	keyExtractor KeyExtractorFunc     // 限流键提取函数
}

// slidingWindowState 滑动窗口在 cache 中的序列化状态。
// 时间戳用 UnixNano 存储，减少 JSON 体积；列表按时间顺序追加，检查时丢掉早于 cutoff 的项。
type slidingWindowState struct {
	Timestamps []int64 `json:"ts"` // 窗口内请求时间戳（UnixNano）
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
	config = cloneRateLimitConfig(config)

	// 应用默认值（在副本上修改，不会改写调用方或全局 DefaultRateLimitConfig）
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
	store, err := newRateLimitCacheStore(config)
	if err != nil {
		return nil, err
	}

	return &SlidingWindowLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		store:              store,
		keyExtractor:       GetKeyExtractor(config.KeyStrategy),
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
//   - 此方法是线程安全的，内部使用存储互斥锁保护读改写
//   - 缓存读写失败时放行，避免限流存储故障拖垮数据面
//   - 每次请求都需要遍历时间戳列表清理过期项，时间复杂度 O(n)
//   - 如果 Rate 很大，时间戳列表会很长，可能影响性能
//   - 窗口状态写入公共 cache；闲置键靠 TTL 过期，避免本地 map 只在再次访问时才删除
func (s *SlidingWindowLimiter) checkSlidingWindow(key string) bool {
	// 加锁保护同一限流器实例内的 Get-改-Set，避免并发把时间戳列表写乱
	s.store.mu.Lock()
	defer s.store.mu.Unlock()

	config := s.GetConfig()
	now := time.Now()
	nowNano := now.UnixNano()
	// 计算窗口截止时间：当前时间 - 窗口大小
	// 例如：now = 100s, windowSize = 60s, cutoff = 40s，只有时间戳 > 40s 的请求才在窗口内
	cutoff := now.Add(-time.Duration(config.WindowSize) * time.Second).UnixNano()

	var window slidingWindowState
	exists, err := s.store.load(key, &window)
	if err != nil {
		// 缓存故障时放行：限流是保护能力，存储不可用不应该变成全站 429
		return true
	}
	if !exists {
		// 情况1: 首次请求该限流键（或 TTL 过期后重建）
		// 创建新窗口，记录当前请求的时间戳；当前请求是窗口内的第一个请求，允许通过
		window.Timestamps = []int64{nowNano}
		_ = s.store.save(key, &window)
		return true
	}

	// 情况2: 限流键已存在，先清理窗口外的过期时间戳
	old := window.Timestamps
	valid := old[:0]
	for _, ts := range old {
		// 只保留在窗口内的时间戳（时间戳在 cutoff 之后）
		if ts > cutoff {
			valid = append(valid, ts)
		}
	}
	window.Timestamps = valid

	// 使用 >= 是因为如果已经有 Rate 个时间戳，当前请求是第 (Rate+1) 个，应该被拒绝
	// 例如：Rate = 100，窗口内有 100 个时间戳，当前请求会被拒绝
	if len(window.Timestamps) >= config.Rate {
		// 已达到速率限制，拒绝请求；仍回写清理后的列表，避免过期时间戳一直占着配额
		_ = s.store.save(key, &window)
		return false
	}

	// 窗口内时间戳数量 < Rate，允许请求通过并记录当前时间戳
	window.Timestamps = append(window.Timestamps, nowNano)
	_ = s.store.save(key, &window)
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
