package limiter

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gateway/pkg/cache"
	"gateway/pkg/logger"
)

const (
	// rateLimitCachePrefix 限流状态键前缀，与业务 cache 键隔离。
	rateLimitCachePrefix = "gw:rl:"
	// minRateLimitCacheTTL 闲置键最短存活时间，避免窗口尚未结束就被淘汰。
	minRateLimitCacheTTL = 60 * time.Second
)

var limiterAnonSeq uint64

// rateLimitCacheStore 用 pkg/cache 全局 Manager 里已注册的实例保存限流键状态。
// 只复用，不在限流模块内 NewMemoryCache / AddCache，避免另起一份配置扰乱全局实例。
// 闲置键靠 TTL 过期，不再依赖「同一 key 再次访问」才删除。
type rateLimitCacheStore struct {
	c    cache.Cache   // Manager 中已有的 memory 或 redis 实例
	ttl  time.Duration // 单条状态 TTL，按窗口与桶填满时间计算
	ns   string        // 限流器命名空间，避免不同配置互相覆盖
	algo string        // 算法名，同一配置切换算法时键隔离
	mu   sync.Mutex    // 保护单进程内 Get-改-Set；多实例共享 redis 时仍可能并发覆盖
}

// newRateLimitCacheStore 按配置解析缓存实例并创建存储。
func newRateLimitCacheStore(config *RateLimitConfig) (*rateLimitCacheStore, error) {
	if config == nil {
		return nil, fmt.Errorf("限流配置不能为空")
	}
	c, err := resolveRateLimitCache(config)
	if err != nil {
		return nil, err
	}
	return &rateLimitCacheStore{
		c:    c,
		ttl:  rateLimitCacheTTL(config),
		ns:   rateLimitStoreNamespace(config),
		algo: string(config.Algorithm),
	}, nil
}

// resolveRateLimitCache 从全局 Manager 解析要使用的 Cache。
// 顺序：custom_config.cache_name → default。找不到则报错，不新建、不 AddCache。
func resolveRateLimitCache(config *RateLimitConfig) (cache.Cache, error) {
	name := rateLimitCacheName(config)
	if name != "" {
		if inst := cache.GetCache(name); inst != nil {
			return inst, nil
		}
		logger.Warn("限流指定的缓存实例不存在，将改用 default", "cache_name", name)
	}
	if inst := cache.GetDefaultCache(); inst != nil {
		return inst, nil
	}
	return nil, fmt.Errorf("限流未找到可用缓存实例，请在 cache 配置中注册 default，或在 custom_config.cache_name 指定已有实例")
}

// rateLimitCacheName 从 CustomConfig 读取 cache_name。
func rateLimitCacheName(config *RateLimitConfig) string {
	if config == nil || config.CustomConfig == nil {
		return ""
	}
	raw, ok := config.CustomConfig["cache_name"]
	if !ok || raw == nil {
		return ""
	}
	name, ok := raw.(string)
	if !ok {
		return ""
	}
	return name
}

// rateLimitStoreNamespace 生成限流器缓存命名空间。
// 有配置 ID 时用 ID，便于同实例热更新后继续累计计数；无 ID 时分配一次性序号，避免测试互相污染。
func rateLimitStoreNamespace(config *RateLimitConfig) string {
	if config != nil && config.ID != "" {
		return config.ID
	}
	return fmt.Sprintf("anon-%d", atomic.AddUint64(&limiterAnonSeq, 1))
}

// rateLimitCacheTTL 按窗口与桶填满时长计算状态键 TTL，并保证不低于最短存活时间。
func rateLimitCacheTTL(config *RateLimitConfig) time.Duration {
	windowSeconds := config.WindowSize
	if windowSeconds < 1 {
		windowSeconds = 1
	}
	// 默认 TTL = 2 * 窗口，保证当前窗口结束前键还在
	ttl := time.Duration(windowSeconds*2) * time.Second
	if config.Rate > 0 && config.Burst > 0 {
		// 令牌桶/漏桶还要覆盖「从空到满/从满到空」的时间，避免补令牌/漏水被截断
		fill := time.Duration(float64(config.Burst)/float64(config.Rate)*2) * time.Second
		if fill > ttl {
			ttl = fill
		}
	}
	if ttl < minRateLimitCacheTTL {
		return minRateLimitCacheTTL
	}
	return ttl
}

// cacheKey 为逻辑限流键加上统一前缀与命名空间。
func (s *rateLimitCacheStore) cacheKey(logicalKey string) string {
	return rateLimitCachePrefix + s.ns + ":" + s.algo + ":" + logicalKey
}

// load 读取指定逻辑键的 JSON 状态；不存在时返回 false, nil。
func (s *rateLimitCacheStore) load(logicalKey string, dst interface{}) (bool, error) {
	raw, err := s.c.Get(context.Background(), s.cacheKey(logicalKey))
	if err != nil {
		return false, err
	}
	if len(raw) == 0 {
		return false, nil
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return false, fmt.Errorf("解析限流状态失败: %w", err)
	}
	return true, nil
}

// save 写入限流状态，并刷新 TTL，使活跃键持续存活、闲置键自动过期。
func (s *rateLimitCacheStore) save(logicalKey string, src interface{}) error {
	raw, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("序列化限流状态失败: %w", err)
	}
	return s.c.Set(context.Background(), s.cacheKey(logicalKey), raw, s.ttl)
}
