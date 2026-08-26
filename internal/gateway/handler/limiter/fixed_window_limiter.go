package limiter

import (
	"fmt"
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
//   - 计数器写入公共 cache，闲置键靠 TTL 过期
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
	store        *rateLimitCacheStore // 限流键状态，走 pkg/cache
	keyExtractor KeyExtractorFunc     // 限流键提取函数
}

// fixedWindowState 固定窗口在 cache 中的序列化状态。
type fixedWindowState struct {
	Count     int   `json:"c"` // 当前窗口请求计数，从 1 开始
	StartNano int64 `json:"s"` // 窗口开始时间（UnixNano）
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

	config.Algorithm = AlgorithmFixedWindow
	store, err := newRateLimitCacheStore(config)
	if err != nil {
		return nil, err
	}

	return &FixedWindowLimiter{
		BaseLimiterHandler: NewBaseLimiterHandler(config),
		store:              store,
		keyExtractor:       GetKeyExtractor(config.KeyStrategy),
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
// 注意：
//   - 此方法是线程安全的，内部使用存储互斥锁保护读改写
//   - 计数器写入公共 cache；闲置键靠 TTL 过期，不再依赖「同一 key 再次访问」才删除
//   - 缓存读写失败时放行，避免限流存储故障拖垮数据面
func (f *FixedWindowLimiter) checkFixedWindow(key string) bool {
	// 加锁保护同一限流器实例内的 Get-改-Set，避免并发把计数写乱
	f.store.mu.Lock()
	defer f.store.mu.Unlock()

	config := f.GetConfig()
	now := time.Now()
	// 计算窗口大小（将秒转换为 time.Duration）
	windowSize := time.Duration(config.WindowSize) * time.Second

	var counter fixedWindowState
	exists, err := f.store.load(key, &counter)
	if err != nil {
		// 缓存故障时放行：限流是保护能力，存储不可用不应该变成全站 429
		return true
	}

	// 情况1: 计数器不存在（首次请求）或窗口已过期
	// 判断条件: 状态不存在 或 当前时间距离窗口开始时间 >= 窗口大小
	if !exists || now.Sub(time.Unix(0, counter.StartNano)) >= windowSize {
		// 创建新窗口，计数从1开始（因为当前请求算作第一个）
		counter.Count = 1
		counter.StartNano = now.UnixNano()
		_ = f.store.save(key, &counter)
		return true
	}

	// 情况2: 计数器存在且窗口未过期
	// 使用 >= 是因为 count 从1开始计数
	// 例如: Rate=100 时，count 可以从 1 到 100，当 count=100 时，下一个请求会被拒绝
	if counter.Count >= config.Rate {
		return false
	}

	// 情况3: 窗口未过期且未达到限制，增加计数并允许请求通过
	counter.Count++
	_ = f.store.save(key, &counter)
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
