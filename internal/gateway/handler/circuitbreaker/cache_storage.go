package circuitbreaker

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gateway/pkg/cache"
	"gateway/pkg/cache/memory"
	"gateway/pkg/logger"
)

const (
	// circuitBreakerCachePrefix 熔断状态键前缀，便于按模式扫描与后续 redis 隔离。
	circuitBreakerCachePrefix = "gw:cb:"
	// circuitBreakerFallbackCacheName 未配置 cache 时注册到 Manager 的进程内实例名。
	// 启动时用同名 redis 实例替换即可切换存储。
	circuitBreakerFallbackCacheName = "circuit_breaker"
	// minCircuitBreakerCacheTTL 状态键最短存活时间，避免窗口尚未结束就被淘汰。
	minCircuitBreakerCacheTTL = 5 * time.Minute
)

// cacheCircuitBreakerStorage 用 pkg/cache.Cache 保存熔断状态。
// 默认走 Manager 中的实例；未配置时回退到进程内 memory，后续把同名实例换成 redis 即可。
type cacheCircuitBreakerStorage struct {
	c   cache.Cache   // Manager 中的 memory 或 redis 实例
	ttl time.Duration // 单条状态 TTL，按窗口与开闸时间计算
	mu  sync.Mutex    // 保护 Increment* 的读改写；Record* 由熔断器 mu 串行
}

// newCacheCircuitBreakerStorage 按配置解析缓存实例并创建存储。
func newCacheCircuitBreakerStorage(config *CircuitBreakerConfig) (CircuitBreakerStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("熔断配置不能为空")
	}
	c, err := resolveCircuitBreakerCache(config)
	if err != nil {
		return nil, err
	}
	ttl := circuitBreakerCacheTTL(config)
	return &cacheCircuitBreakerStorage{c: c, ttl: ttl}, nil
}

// resolveCircuitBreakerCache 解析要使用的 Cache：先按 cache_name / default，再回退进程内 memory。
func resolveCircuitBreakerCache(config *CircuitBreakerConfig) (cache.Cache, error) {
	name := ""
	if config.StorageConfig != nil {
		name = config.StorageConfig["cache_name"]
	}
	// 优先用配置指定的实例，便于按环境挂 redis
	if name != "" {
		if inst := cache.GetCache(name); inst != nil {
			return inst, nil
		}
		logger.Warn("熔断指定的缓存实例不存在，将回退", "cache_name", name)
	}
	if inst := cache.GetDefaultCache(); inst != nil {
		return inst, nil
	}
	if inst := cache.GetCache(circuitBreakerFallbackCacheName); inst != nil {
		return inst, nil
	}

	// 启动阶段尚未注册 cache 时，创建进程内 memory 并挂到 Manager，供后续替换
	mem, err := memory.NewMemoryCache(&memory.MemoryConfig{
		Enabled:           true,
		MaxSize:           10000,
		CleanupInterval:   time.Minute,
		EnableLazyCleanup: true,
		KeyPrefix:         "",
	})
	if err != nil {
		return nil, fmt.Errorf("创建熔断回退内存缓存失败: %w", err)
	}
	if err := cache.AddCache(circuitBreakerFallbackCacheName, mem); err != nil {
		// 并发创建时另一路已注册成功，关闭本实例后复用已有实例
		if inst := cache.GetCache(circuitBreakerFallbackCacheName); inst != nil {
			_ = mem.Close()
			return inst, nil
		}
		_ = mem.Close()
		return nil, fmt.Errorf("注册熔断回退缓存失败: %w", err)
	}
	logger.Info("熔断使用进程内 memory 缓存，后续可用 redis 实例替换",
		"cache_name", circuitBreakerFallbackCacheName)
	return mem, nil
}

// circuitBreakerCacheTTL 按统计窗口与开闸时长计算状态键 TTL，并保证不低于最短存活时间。
func circuitBreakerCacheTTL(config *CircuitBreakerConfig) time.Duration {
	seconds := config.WindowSizeSeconds + config.OpenTimeoutSeconds
	if seconds < 1 {
		seconds = 60
	}
	ttl := time.Duration(seconds*2) * time.Second
	if ttl < minCircuitBreakerCacheTTL {
		return minCircuitBreakerCacheTTL
	}
	return ttl
}

// circuitBreakerCacheKey 为熔断逻辑键加上统一前缀，与业务 cache 键隔离。
func circuitBreakerCacheKey(key string) string {
	return circuitBreakerCachePrefix + key
}

// GetInfo 读取指定 key 的熔断状态；不存在时返回 nil, nil。
func (s *cacheCircuitBreakerStorage) GetInfo(key string) (*CircuitBreakerInfo, error) {
	raw, err := s.c.Get(context.Background(), circuitBreakerCacheKey(key))
	if err != nil {
		return nil, err
	}
	if len(raw) == 0 {
		return nil, nil
	}
	info := &CircuitBreakerInfo{}
	if err := json.Unmarshal(raw, info); err != nil {
		return nil, fmt.Errorf("解析熔断状态失败: %w", err)
	}
	return info, nil
}

// SetInfo 写入熔断状态。
func (s *cacheCircuitBreakerStorage) SetInfo(key string, info *CircuitBreakerInfo) error {
	if info == nil {
		return s.Reset(key)
	}
	raw, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化熔断状态失败: %w", err)
	}
	return s.c.Set(context.Background(), circuitBreakerCacheKey(key), raw, s.ttl)
}

// IncrementSuccess 增加成功计数并回写。
func (s *cacheCircuitBreakerStorage) IncrementSuccess(key string, responseTime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, err := s.GetInfo(key)
	if err != nil {
		return err
	}
	if info == nil {
		info = newClosedCircuit()
	}
	info.TotalRequests++
	info.SuccessRequests++
	info.LastRequestTime = time.Now().Unix()
	// 慢调用判定在 RecordSuccess 中按阈值处理，这里只维护计数
	_ = responseTime
	return s.SetInfo(key, info)
}

// IncrementFailure 增加失败计数并回写。
func (s *cacheCircuitBreakerStorage) IncrementFailure(key string, responseTime int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	info, err := s.GetInfo(key)
	if err != nil {
		return err
	}
	if info == nil {
		info = newClosedCircuit()
	}
	now := time.Now().Unix()
	info.TotalRequests++
	info.FailureRequests++
	info.LastFailureTime = now
	info.LastRequestTime = now
	_ = responseTime
	return s.SetInfo(key, info)
}

// Reset 删除指定 key 的状态。
func (s *cacheCircuitBreakerStorage) Reset(key string) error {
	return s.c.Delete(context.Background(), circuitBreakerCacheKey(key))
}

// Cleanup 删除全部熔断缓存键。
func (s *cacheCircuitBreakerStorage) Cleanup() error {
	keys, err := s.c.Keys(context.Background(), circuitBreakerCachePrefix+"*")
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	return s.c.MDelete(context.Background(), keys)
}

// Close 不关闭共享 Cache，避免误关 Manager 中的 redis/memory 实例。
func (s *cacheCircuitBreakerStorage) Close() error {
	return nil
}

// newClosedCircuit 创建关闭态的空统计。
func newClosedCircuit() *CircuitBreakerInfo {
	now := time.Now().Unix()
	return &CircuitBreakerInfo{
		State:       StateClosed,
		WindowStart: now,
	}
}
