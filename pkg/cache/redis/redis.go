package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"gohub/pkg/cache"
	"gohub/pkg/config"
)

// init 函数自动注册Redis缓存工厂
func init() {
	// 注册Redis缓存工厂到全局管理器
	cache.GetGlobalManager().RegisterFactory(cache.TypeRedis, func(config interface{}) (cache.Cache, error) {
		var redisConfig *RedisConfig
		if config != nil {
			if cfg, ok := config.(*RedisConfig); ok {
				redisConfig = cfg
			} else {
				return nil, fmt.Errorf("invalid redis config type")
			}
		}
		return NewRedisCache(redisConfig)
	})
}

// RedisCache Redis缓存实现
// 实现了统一的Cache接口，提供Redis缓存功能
type RedisCache struct {
	// client Redis客户端
	client *redis.Client
	// config Redis配置
	config *RedisConfig
	// keyPrefix 键前缀
	keyPrefix string
}

// NewRedisCache 创建新的Redis缓存实例
// 参数:
//   - cfg: Redis配置，如果为nil则从database.yaml中加载
//
// 返回:
//   - cache.Cache: 缓存接口实例
//   - error: 可能的错误
func NewRedisCache(cfg *RedisConfig) (cache.Cache, error) {
	if cfg == nil {
		// 从配置文件加载Redis配置
		loadedCfg, err := LoadRedisConfigFromFile()
		if err != nil {
			return nil, fmt.Errorf("failed to load redis config: %w", err)
		}
		cfg = loadedCfg
	}

	// 设置默认值
	cfg.SetDefaults()

	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid redis config: %w", err)
	}

	// 创建Redis客户端
	client := redis.NewClient(&redis.Options{
		Addr:            cfg.GetAddress(),
		Password:        cfg.Password,
		DB:              cfg.DB,
		PoolSize:        cfg.PoolSize,
		MinIdleConns:    cfg.MinIdleConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		MaxActiveConns:  cfg.MaxActiveConns,
		ConnMaxIdleTime: cfg.GetIdleTimeoutDuration(),
		DialTimeout:     cfg.DialTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		PoolTimeout:     cfg.PoolTimeout,
		MaxRetries:      cfg.MaxRetries,
		MinRetryBackoff: cfg.MinRetryBackoff,
		MaxRetryBackoff: cfg.MaxRetryBackoff,
	})

	// 获取键前缀
	keyPrefix := cfg.KeyPrefix

	// 创建Redis缓存实例
	redisCache := &RedisCache{
		client:    client,
		config:    cfg,
		keyPrefix: keyPrefix,
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisCache.Ping(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return redisCache, nil
}

// LoadRedisConfigFromFile 从配置文件加载Redis配置
func LoadRedisConfigFromFile() (*RedisConfig, error) {
	// 尝试加载database.yaml配置文件
	err := config.LoadConfigFile("configs/database.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load database config: %w", err)
	}

	// 从cache.redis路径读取配置
	var redisConfig RedisConfig
	if err := config.GetSection("cache.redis", &redisConfig); err != nil {
		return nil, fmt.Errorf("failed to parse redis config from cache.redis section: %w", err)
	}

	return &redisConfig, nil
}

// buildKey 构建完整的缓存键（加上前缀）
func (r *RedisCache) buildKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	return r.keyPrefix + ":" + key
}

// parseKey 解析缓存键（去掉前缀）
func (r *RedisCache) parseKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	prefix := r.keyPrefix + ":"
	if strings.HasPrefix(key, prefix) {
		return key[len(prefix):]
	}
	return key
}

// Get 获取缓存值
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, cache.ErrCacheKeyNotFound
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return []byte(result), nil
}

// Set 设置缓存值
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	fullKey := r.buildKey(key)
	err := r.client.Set(ctx, fullKey, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// Delete 删除缓存值
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	fullKey := r.buildKey(key)
	err := r.client.Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return result > 0, nil
}

// MGet 批量获取缓存值
func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := r.client.MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget error: %w", err)
	}

	// 构建结果映射
	resultMap := make(map[string][]byte)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				resultMap[keys[i]] = []byte(str)
			}
		}
	}

	return resultMap, nil
}

// MSet 批量设置缓存值
func (r *RedisCache) MSet(ctx context.Context, kvPairs map[string][]byte, expiration time.Duration) error {
	if len(kvPairs) == 0 {
		return nil
	}

	// 如果没有过期时间，使用MSET
	if expiration == 0 {
		pairs := make([]interface{}, 0, len(kvPairs)*2)
		for key, value := range kvPairs {
			pairs = append(pairs, r.buildKey(key), value)
		}
		err := r.client.MSet(ctx, pairs...).Err()
		if err != nil {
			return fmt.Errorf("redis mset error: %w", err)
		}
		return nil
	}

	// 如果有过期时间，使用管道批量SET
	pipe := r.client.Pipeline()
	for key, value := range kvPairs {
		pipe.Set(ctx, r.buildKey(key), value, expiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline mset error: %w", err)
	}

	return nil
}

// MDelete 批量删除缓存值
func (r *RedisCache) MDelete(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	err := r.client.Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("redis mdelete error: %w", err)
	}

	return nil
}

// Increment 原子递增
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment error: %w", err)
	}
	return result, nil
}

// Decrement 原子递减
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.DecrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decrement error: %w", err)
	}
	return result, nil
}

// SetNX 仅当键不存在时设置值
func (r *RedisCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.SetNX(ctx, fullKey, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return result, nil
}

// TTL 获取键的剩余生存时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return result, nil
}

// Expire 设置键的过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	fullKey := r.buildKey(key)
	result, err := r.client.Expire(ctx, fullKey, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis expire error: %w", err)
	}
	return result, nil
}

// Ping 测试连接
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}

// Close 关闭缓存连接
func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Stats 获取缓存统计信息
func (r *RedisCache) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	// 获取连接池统计信息
	poolStats := r.client.PoolStats()
	stats["pool"] = map[string]interface{}{
		"hits":        poolStats.Hits,
		"misses":      poolStats.Misses,
		"timeouts":    poolStats.Timeouts,
		"total_conns": poolStats.TotalConns,
		"idle_conns":  poolStats.IdleConns,
		"stale_conns": poolStats.StaleConns,
	}

	// 获取配置信息
	stats["config"] = map[string]interface{}{
		"addr":             r.config.GetAddress(),
		"db":               r.config.DB,
		"pool_size":        r.config.PoolSize,
		"min_idle_conns":   r.config.MinIdleConns,
		"max_idle_conns":   r.config.MaxIdleConns,
		"max_active_conns": r.config.MaxActiveConns,
		"key_prefix":       r.keyPrefix,
	}

	return stats
}

// FlushAll 清空所有缓存（谨慎使用）
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flushall error: %w", err)
	}
	return nil
}

// SelectDB 选择数据库
func (r *RedisCache) SelectDB(ctx context.Context, db int) error {
	// 对于go-redis客户端，我们需要创建一个新的连接到指定数据库
	// 但这会改变当前实例的数据库，所以我们更新配置并重新连接
	if db < 0 || db > 15 {
		return fmt.Errorf("redis select db error: database number must be between 0 and 15, got %d", db)
	}
	
	// 更新配置中的数据库编号
	r.config.DB = db
	
	// 关闭当前连接
	r.client.Close()
	
	// 创建新的Redis客户端连接到指定数据库
	options := &redis.Options{
		Addr:            r.config.GetAddress(),
		Password:        r.config.Password,
		DB:              db,
		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.MinIdleConns,
		MaxIdleConns:    r.config.MaxIdleConns,
		MaxActiveConns:  r.config.MaxActiveConns,
		ConnMaxIdleTime: r.config.GetIdleTimeoutDuration(),
		DialTimeout:     r.config.DialTimeout,
		ReadTimeout:     r.config.ReadTimeout,
		WriteTimeout:    r.config.WriteTimeout,
		PoolTimeout:     r.config.PoolTimeout,
		MaxRetries:      r.config.MaxRetries,
		MinRetryBackoff: r.config.MinRetryBackoff,
		MaxRetryBackoff: r.config.MaxRetryBackoff,
	}
	
	r.client = redis.NewClient(options)
	
	// 测试新连接
	err := r.client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis select db error: failed to connect to database %d: %w", db, err)
	}
	
	return nil
}
