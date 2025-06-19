package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"

	"gohub/pkg/config"
)

// RedisCache Redis缓存实现
// 实现了统一的Cache接口，提供Redis缓存功能
// 支持单机、哨兵、集群三种模式
type RedisCache struct {
	// client 单机或哨兵模式的Redis客户端
	client *redis.Client
	// clusterClient 集群模式的Redis客户端
	clusterClient *redis.ClusterClient
	// config Redis配置
	config *RedisConfig
	// keyPrefix 键前缀
	keyPrefix string
	// isCluster 是否为集群模式
	isCluster bool
}

// NewRedisCache 创建新的Redis缓存实例
// 根据配置模式自动选择单机、哨兵或集群连接
// 该函数会自动测试连接有效性，如果连接失败会自动清理资源
// 参数:
//   - cfg: Redis配置，如果为nil则从database.yaml中加载默认配置
//
// 返回:
//   - *RedisCache: Redis缓存实例，使用完毕后需要调用Close()方法释放资源
//   - error: 创建或连接失败时返回错误信息
func NewRedisCache(cfg *RedisConfig) (*RedisCache, error) {
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

	// 创建Redis缓存实例
	redisCache := &RedisCache{
		config:    cfg,
		keyPrefix: cfg.KeyPrefix,
		isCluster: cfg.IsClusterMode(),
	}

	// 根据模式创建不同的Redis客户端
	var err error
	switch cfg.Mode {
	case ModeSingle:
		err = redisCache.createSingleClient()
	case ModeSentinel:
		err = redisCache.createSentinelClient()
	case ModeCluster:
		err = redisCache.createClusterClient()
	default:
		return nil, fmt.Errorf("unsupported redis mode: %s", cfg.Mode)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create redis client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisCache.Ping(ctx); err != nil {
		redisCache.Close()
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return redisCache, nil
}

// createSingleClient 创建单机模式客户端
func (r *RedisCache) createSingleClient() error {
	opts := &redis.Options{
		Addr:            r.config.GetAddress(),
		Password:        r.config.Password,
		DB:              r.config.DB,
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

	// 配置TLS
	if r.config.TLSEnabled {
		opts.TLSConfig = r.createTLSConfig()
	}

	r.client = redis.NewClient(opts)
	return nil
}

// createSentinelClient 创建哨兵模式客户端
func (r *RedisCache) createSentinelClient() error {
	opts := &redis.FailoverOptions{
		MasterName:      r.config.MasterName,
		SentinelAddrs:   r.config.GetSentinelAddresses(),
		Password:        r.config.Password,
		DB:              r.config.DB,
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

	// 配置哨兵用户名和密码
	if r.config.SentinelUsername != "" {
		opts.SentinelUsername = r.config.SentinelUsername
	}
	if r.config.SentinelPassword != "" {
		opts.SentinelPassword = r.config.SentinelPassword
	}

	// 配置TLS
	if r.config.TLSEnabled {
		opts.TLSConfig = r.createTLSConfig()
	}

	r.client = redis.NewFailoverClient(opts)
	return nil
}

// createClusterClient 创建集群模式客户端
func (r *RedisCache) createClusterClient() error {
	opts := &redis.ClusterOptions{
		Addrs:           r.config.GetClusterAddresses(),
		Password:        r.config.ClusterPassword,
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
		MaxRedirects:    r.config.MaxRedirects,
		ReadOnly:        r.config.ReadOnly,
		RouteByLatency:  r.config.RouteByLatency,
		RouteRandomly:   r.config.RouteRandomly,
	}

	// 配置集群用户名
	if r.config.ClusterUsername != "" {
		opts.Username = r.config.ClusterUsername
	}

	// 配置TLS
	if r.config.TLSEnabled {
		opts.TLSConfig = r.createTLSConfig()
	}

	r.clusterClient = redis.NewClusterClient(opts)
	return nil
}

// createTLSConfig 创建TLS配置
func (r *RedisCache) createTLSConfig() *tls.Config {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: r.config.TLSInsecureSkipVerify,
	}

	// 加载证书文件
	if r.config.TLSCertFile != "" && r.config.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(r.config.TLSCertFile, r.config.TLSKeyFile)
		if err == nil {
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	// 加载CA证书
	if r.config.TLSCACertFile != "" {
		// 这里可以添加CA证书加载逻辑
		// 由于复杂性，这里暂时跳过
	}

	return tlsConfig
}

// getUniversalClient 获取通用客户端接口（用于内部操作）
func (r *RedisCache) getUniversalClient() redis.UniversalClient {
	if r.isCluster {
		return r.clusterClient
	}
	return r.client
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
// 如果配置了前缀，会自动在前缀和键之间添加分隔符
// 参数:
//   - key: 原始缓存键
// 返回:
//   - string: 带前缀的完整缓存键，格式为 "prefix.key" 或直接返回key（无前缀时）
func (r *RedisCache) buildKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	// 如果前缀不以分隔符结尾，自动添加分隔符
	if !strings.HasSuffix(r.keyPrefix, ".") && !strings.HasSuffix(r.keyPrefix, ":") {
		return r.keyPrefix + "." + key
	}
	return r.keyPrefix + key
}

// parseKey 解析缓存键（去掉前缀）
// 从完整的缓存键中提取原始键名，去除前缀部分
// 参数:
//   - key: 完整的缓存键（包含前缀）
// 返回:
//   - string: 去除前缀后的原始键名
func (r *RedisCache) parseKey(key string) string {
	if r.keyPrefix == "" {
		return key
	}
	
	// 构建完整前缀（包含分隔符）
	fullPrefix := r.keyPrefix
	if !strings.HasSuffix(r.keyPrefix, ".") && !strings.HasSuffix(r.keyPrefix, ":") {
		fullPrefix = r.keyPrefix + "."
	}
	
	if strings.HasPrefix(key, fullPrefix) {
		return key[len(fullPrefix):]
	}
	return key
}

// resolveExpiration 解析过期时间
// 处理过期时间逻辑：0表示使用默认过期时间，负数表示永不过期
// 参数:
//   - expiration: 输入的过期时间
// 返回:
//   - time.Duration: 最终使用的过期时间
func (r *RedisCache) resolveExpiration(expiration time.Duration) time.Duration {
	if expiration == 0 {
		// 0表示使用配置的默认过期时间
		return r.config.GetDefaultExpiration()
	} else if expiration < 0 {
		// 负数表示永不过期
		return 0
	}
	// 正数直接使用
	return expiration
}

// Get 获取缓存值
// 从Redis中获取指定键的值
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
// 返回:
//   - []byte: 缓存值的字节数组，键不存在时返回nil
//   - error: 获取失败时返回错误，键不存在不算错误
func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在时返回nil而不是错误
		}
		return nil, fmt.Errorf("redis get error: %w", err)
	}
	return []byte(result), nil
}

// GetString 获取缓存值（字符串）
// 从Redis中获取指定键的字符串值
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
// 返回:
//   - string: 缓存值的字符串，键不存在时返回空字符串
//   - error: 获取失败时返回错误，键不存在不算错误
func (r *RedisCache) GetString(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().Get(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在时返回空字符串而不是错误
		}
		return "", fmt.Errorf("redis get error: %w", err)
	}
	return result, nil
}

// Set 设置缓存值
// 在Redis中设置指定键的值，支持过期时间
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要缓存的值（字节数组）
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
// 返回:
//   - error: 设置失败时返回错误信息
func (r *RedisCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return fmt.Errorf("缓存值不能为nil")
	}
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)
	
	fullKey := r.buildKey(key)
	err := r.getUniversalClient().Set(ctx, fullKey, value, finalExpiration).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// SetString 设置缓存值（字符串）
// 在Redis中设置指定键的字符串值，支持过期时间
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要缓存的字符串值
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
// 返回:
//   - error: 设置失败时返回错误信息
func (r *RedisCache) SetString(ctx context.Context, key string, value string, expiration time.Duration) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)
	
	fullKey := r.buildKey(key)
	err := r.getUniversalClient().Set(ctx, fullKey, value, finalExpiration).Err()
	if err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}
	return nil
}

// Delete 删除缓存值
// 从Redis中删除指定的键值对
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 要删除的缓存键名（不包含前缀）
// 返回:
//   - error: 删除失败时返回错误信息
// 注意: 删除不存在的键不会返回错误
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	err := r.getUniversalClient().Del(ctx, fullKey).Err()
	if err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// Exists 检查键是否存在
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().Exists(ctx, fullKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return result > 0, nil
}

// MGet 批量获取缓存值
// 一次性获取多个键的值，比多次调用Get更高效
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - keys: 要获取的键名列表（不包含前缀）
// 返回:
//   - map[string][]byte: 键值映射，只包含存在的键值对
//   - error: 获取失败时返回错误信息
// 注意: 不存在的键不会出现在返回的映射中
func (r *RedisCache) MGet(ctx context.Context, keys []string) (map[string][]byte, error) {
	if len(keys) == 0 {
		return make(map[string][]byte), nil
	}

	// 验证键的有效性
	for i, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("第%d个键不能为空", i+1)
		}
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := r.getUniversalClient().MGet(ctx, fullKeys...).Result()
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

// MGetString 批量获取缓存值（字符串）
// 一次性获取多个键的字符串值，比多次调用GetString更高效
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - keys: 要获取的键名列表（不包含前缀）
// 返回:
//   - map[string]string: 键值映射，只包含存在的键值对
//   - error: 获取失败时返回错误信息
// 注意: 不存在的键不会出现在返回的映射中
func (r *RedisCache) MGetString(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return make(map[string]string), nil
	}

	// 验证键的有效性
	for i, key := range keys {
		if key == "" {
			return nil, fmt.Errorf("第%d个键不能为空", i+1)
		}
	}

	// 构建完整键列表
	fullKeys := make([]string, len(keys))
	for i, key := range keys {
		fullKeys[i] = r.buildKey(key)
	}

	// 批量获取
	results, err := r.getUniversalClient().MGet(ctx, fullKeys...).Result()
	if err != nil {
		return nil, fmt.Errorf("redis mget error: %w", err)
	}

	// 构建结果映射
	resultMap := make(map[string]string)
	for i, result := range results {
		if result != nil {
			if str, ok := result.(string); ok {
				resultMap[keys[i]] = str
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

	client := r.getUniversalClient()
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	// 如果没有过期时间，使用MSET
	if finalExpiration == 0 {
		pairs := make([]interface{}, 0, len(kvPairs)*2)
		for key, value := range kvPairs {
			pairs = append(pairs, r.buildKey(key), value)
		}
		err := client.MSet(ctx, pairs...).Err()
		if err != nil {
			return fmt.Errorf("redis mset error: %w", err)
		}
		return nil
	}

	// 如果有过期时间，使用管道批量SET
	pipe := client.Pipeline()
	for key, value := range kvPairs {
		pipe.Set(ctx, r.buildKey(key), value, finalExpiration)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline mset error: %w", err)
	}

	return nil
}

// MSetString 批量设置缓存值（字符串）
// 一次性设置多个键的字符串值，比多次调用SetString更高效
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - kvPairs: 键值对映射
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
// 返回:
//   - error: 设置失败时返回错误信息
func (r *RedisCache) MSetString(ctx context.Context, kvPairs map[string]string, expiration time.Duration) error {
	if len(kvPairs) == 0 {
		return nil
	}

	client := r.getUniversalClient()
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)

	// 如果没有过期时间，使用MSET
	if finalExpiration == 0 {
		pairs := make([]interface{}, 0, len(kvPairs)*2)
		for key, value := range kvPairs {
			pairs = append(pairs, r.buildKey(key), value)
		}
		err := client.MSet(ctx, pairs...).Err()
		if err != nil {
			return fmt.Errorf("redis mset error: %w", err)
		}
		return nil
	}

	// 如果有过期时间，使用管道批量SET
	pipe := client.Pipeline()
	for key, value := range kvPairs {
		pipe.Set(ctx, r.buildKey(key), value, finalExpiration)
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

	err := r.getUniversalClient().Del(ctx, fullKeys...).Err()
	if err != nil {
		return fmt.Errorf("redis mdelete error: %w", err)
	}

	return nil
}

// Increment 原子递增
func (r *RedisCache) Increment(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().IncrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis increment error: %w", err)
	}
	return result, nil
}

// Decrement 原子递减
func (r *RedisCache) Decrement(ctx context.Context, key string, delta int64) (int64, error) {
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().DecrBy(ctx, fullKey, delta).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decrement error: %w", err)
	}
	return result, nil
}

// SetNX 仅当键不存在时设置值
// 原子操作：只有当键不存在时才设置值，常用于分布式锁
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要设置的值（字节数组）
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
// 返回:
//   - bool: true表示设置成功（键不存在），false表示键已存在
//   - error: 操作失败时返回错误信息
func (r *RedisCache) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return false, fmt.Errorf("缓存值不能为nil")
	}
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().SetNX(ctx, fullKey, value, finalExpiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return result, nil
}

// SetNXString 仅当键不存在时设置字符串值
// 原子操作：只有当键不存在时才设置字符串值，常用于分布式锁
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - key: 缓存键名（不包含前缀）
//   - value: 要设置的字符串值
//   - expiration: 过期时间，0表示使用配置的默认过期时间，负数表示永不过期
// 返回:
//   - bool: true表示设置成功（键不存在），false表示键已存在
//   - error: 操作失败时返回错误信息
func (r *RedisCache) SetNXString(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}
	
	// 处理过期时间：0表示使用默认过期时间，负数表示永不过期
	finalExpiration := r.resolveExpiration(expiration)
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().SetNX(ctx, fullKey, value, finalExpiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx error: %w", err)
	}
	return result, nil
}

// TTL 获取键的剩余生存时间
func (r *RedisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().TTL(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return result, nil
}

// Expire 设置键的过期时间
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().Expire(ctx, fullKey, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis expire error: %w", err)
	}
	return result, nil
}

// Ping 测试连接
func (r *RedisCache) Ping(ctx context.Context) error {
	err := r.getUniversalClient().Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}

// Close 关闭缓存连接
// 安全关闭Redis连接，释放所有相关资源
// 该方法是幂等的，多次调用不会产生副作用
// 返回:
//   - error: 关闭连接时的错误信息
func (r *RedisCache) Close() error {
	var closeErr error
	
	// 关闭集群客户端
	if r.isCluster && r.clusterClient != nil {
		if err := r.clusterClient.Close(); err != nil {
			closeErr = fmt.Errorf("关闭集群客户端失败: %w", err)
		}
		r.clusterClient = nil // 防止重复关闭
	}
	
	// 关闭单机/哨兵客户端
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			if closeErr != nil {
				closeErr = fmt.Errorf("%v; 关闭客户端失败: %w", closeErr, err)
			} else {
				closeErr = fmt.Errorf("关闭客户端失败: %w", err)
			}
		}
		r.client = nil // 防止重复关闭
	}
	
	return closeErr
}

// Stats 获取缓存统计信息
func (r *RedisCache) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	var poolStats *redis.PoolStats
	if r.isCluster && r.clusterClient != nil {
		poolStats = r.clusterClient.PoolStats()
	} else if r.client != nil {
		poolStats = r.client.PoolStats()
	}

	if poolStats != nil {
		stats["pool"] = map[string]interface{}{
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
		}
	}

	// 获取配置信息
	stats["config"] = map[string]interface{}{
		"mode":             string(r.config.Mode),
		"connection_string": r.config.GetConnectionString(),
		"pool_size":        r.config.PoolSize,
		"min_idle_conns":   r.config.MinIdleConns,
		"max_idle_conns":   r.config.MaxIdleConns,
		"max_active_conns": r.config.MaxActiveConns,
		"key_prefix":       r.keyPrefix,
		"is_cluster":       r.isCluster,
	}

	return stats
}

// FlushAll 清空所有缓存（谨慎使用）
func (r *RedisCache) FlushAll(ctx context.Context) error {
	err := r.getUniversalClient().FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flushall error: %w", err)
	}
	return nil
}

// IsConnected 检查连接是否有效
// 快速检查Redis连接状态，不发送网络请求
// 返回:
//   - bool: true表示连接对象存在，false表示连接已关闭或不存在
// 注意: 该方法只检查连接对象是否存在，不检查网络连通性
func (r *RedisCache) IsConnected() bool {
	if r.isCluster {
		return r.clusterClient != nil
	}
	return r.client != nil
}

// CheckResourceLeak 检查是否存在资源泄漏
// 检查连接池状态，识别可能的资源泄漏问题
// 返回:
//   - bool: true表示可能存在资源泄漏
//   - map[string]interface{}: 详细的资源状态信息
func (r *RedisCache) CheckResourceLeak() (bool, map[string]interface{}) {
	if !r.IsConnected() {
		return false, map[string]interface{}{
			"status": "disconnected",
			"leak_detected": false,
		}
	}

	var poolStats *redis.PoolStats
	if r.isCluster && r.clusterClient != nil {
		poolStats = r.clusterClient.PoolStats()
	} else if r.client != nil {
		poolStats = r.client.PoolStats()
	}

	if poolStats == nil {
		return false, map[string]interface{}{
			"status": "no_pool_stats",
			"leak_detected": false,
		}
	}

	// 检查资源泄漏的指标
	leakDetected := false
	warnings := make([]string, 0)

	// 1. 检查超时连接比例
	if poolStats.TotalConns > 0 {
		timeoutRatio := float64(poolStats.Timeouts) / float64(poolStats.TotalConns)
		if timeoutRatio > 0.1 { // 超过10%的连接超时
			leakDetected = true
			warnings = append(warnings, "连接超时比例过高")
		}
	}

	// 2. 检查连接池使用率
	if r.config.PoolSize > 0 {
		usageRatio := float64(poolStats.TotalConns) / float64(r.config.PoolSize)
		if usageRatio > 0.9 { // 连接池使用率超过90%
			warnings = append(warnings, "连接池使用率过高")
		}
	}

	// 3. 检查空闲连接异常
	if int(poolStats.IdleConns) < r.config.MinIdleConns {
		warnings = append(warnings, "空闲连接数低于最小值")
	}

	// 4. 检查失效连接
	if poolStats.StaleConns > poolStats.TotalConns/2 {
		leakDetected = true
		warnings = append(warnings, "失效连接数过多")
	}

	return leakDetected, map[string]interface{}{
		"status":           "connected",
		"leak_detected":    leakDetected,
		"warnings":         warnings,
		"pool_stats":       poolStats,
		"config_pool_size": r.config.PoolSize,
		"config_min_idle":  r.config.MinIdleConns,
		"config_max_idle":  r.config.MaxIdleConns,
	}
}

// SelectDB 选择数据库（仅单机和哨兵模式支持）
// 使用Redis的SELECT命令切换到指定数据库，无需重新创建连接
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - db: 目标数据库编号（0-15）
// 返回:
//   - error: 切换失败时返回错误信息
// 注意: 集群模式不支持数据库选择，因为集群中的数据是分片的
func (r *RedisCache) SelectDB(ctx context.Context, db int) error {
	if r.isCluster {
		return fmt.Errorf("集群模式不支持数据库选择")
	}

	if db < 0 || db > 15 {
		return fmt.Errorf("数据库编号必须在0-15之间，当前值: %d", db)
	}

	// 如果已经是目标数据库，无需切换
	if r.config.DB == db {
		return nil
	}

	// 使用SELECT命令切换数据库
	err := r.client.Do(ctx, "SELECT", db).Err()
	if err != nil {
		return fmt.Errorf("切换到数据库%d失败: %w", db, err)
	}

	// 更新配置中的数据库编号
	r.config.DB = db

	return nil
}

// === 扩展操作实现 ===

// Keys 获取匹配模式的所有键（谨慎使用）
func (r *RedisCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	if pattern == "" {
		return nil, fmt.Errorf("匹配模式不能为空")
	}
	
	// 如果有前缀，需要在模式前加上前缀
	fullPattern := pattern
	if r.keyPrefix != "" {
		fullPattern = r.buildKey(pattern)
	}
	
	keys, err := r.getUniversalClient().Keys(ctx, fullPattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys error: %w", err)
	}
	
	// 如果有前缀，需要去掉前缀
	if r.keyPrefix != "" {
		for i, key := range keys {
			keys[i] = r.parseKey(key)
		}
	}
	
	return keys, nil
}

// Size 获取缓存中键的数量
func (r *RedisCache) Size(ctx context.Context) (int64, error) {
	size, err := r.getUniversalClient().DBSize(ctx).Result()
	if err != nil {
		return 0, fmt.Errorf("redis dbsize error: %w", err)
	}
	return size, nil
}

// GetSet 设置新值并返回旧值
func (r *RedisCache) GetSet(ctx context.Context, key string, value []byte) ([]byte, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	if value == nil {
		return nil, fmt.Errorf("缓存值不能为nil")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().GetSet(ctx, fullKey, value).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // 键不存在时返回nil
		}
		return nil, fmt.Errorf("redis getset error: %w", err)
	}
	return []byte(result), nil
}

// GetSetString 设置新字符串值并返回旧字符串值
func (r *RedisCache) GetSetString(ctx context.Context, key string, value string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().GetSet(ctx, fullKey, value).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 键不存在时返回空字符串
		}
		return "", fmt.Errorf("redis getset error: %w", err)
	}
	return result, nil
}

// Append 向字符串值追加内容
func (r *RedisCache) Append(ctx context.Context, key string, value string) (int, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().Append(ctx, fullKey, value).Result()
	if err != nil {
		return 0, fmt.Errorf("redis append error: %w", err)
	}
	return int(result), nil
}

// HSet 设置哈希字段
func (r *RedisCache) HSet(ctx context.Context, key, field, value string) error {
	if key == "" {
		return fmt.Errorf("缓存键不能为空")
	}
	if field == "" {
		return fmt.Errorf("字段名不能为空")
	}
	
	fullKey := r.buildKey(key)
	err := r.getUniversalClient().HSet(ctx, fullKey, field, value).Err()
	if err != nil {
		return fmt.Errorf("redis hset error: %w", err)
	}
	return nil
}

// HGet 获取哈希字段值
func (r *RedisCache) HGet(ctx context.Context, key, field string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	if field == "" {
		return "", fmt.Errorf("字段名不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().HGet(ctx, fullKey, field).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 字段不存在时返回空字符串
		}
		return "", fmt.Errorf("redis hget error: %w", err)
	}
	return result, nil
}

// HGetAll 获取哈希的所有字段和值
func (r *RedisCache) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().HGetAll(ctx, fullKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis hgetall error: %w", err)
	}
	return result, nil
}

// HDel 删除哈希字段
func (r *RedisCache) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(fields) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().HDel(ctx, fullKey, fields...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis hdel error: %w", err)
	}
	return result, nil
}

// LPush 向列表左侧推入元素
func (r *RedisCache) LPush(ctx context.Context, key string, values ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(values) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}
	
	result, err := r.getUniversalClient().LPush(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis lpush error: %w", err)
	}
	return result, nil
}

// RPush 向列表右侧推入元素
func (r *RedisCache) RPush(ctx context.Context, key string, values ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(values) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(values))
	for i, v := range values {
		vals[i] = v
	}
	
	result, err := r.getUniversalClient().RPush(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis rpush error: %w", err)
	}
	return result, nil
}

// LPop 从列表左侧弹出元素
func (r *RedisCache) LPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().LPop(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 列表为空时返回空字符串
		}
		return "", fmt.Errorf("redis lpop error: %w", err)
	}
	return result, nil
}

// RPop 从列表右侧弹出元素
func (r *RedisCache) RPop(ctx context.Context, key string) (string, error) {
	if key == "" {
		return "", fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().RPop(ctx, fullKey).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // 列表为空时返回空字符串
		}
		return "", fmt.Errorf("redis rpop error: %w", err)
	}
	return result, nil
}

// LLen 获取列表长度
func (r *RedisCache) LLen(ctx context.Context, key string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().LLen(ctx, fullKey).Result()
	if err != nil {
		return 0, fmt.Errorf("redis llen error: %w", err)
	}
	return result, nil
}

// SAdd 向集合添加元素
func (r *RedisCache) SAdd(ctx context.Context, key string, members ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(members) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(members))
	for i, v := range members {
		vals[i] = v
	}
	
	result, err := r.getUniversalClient().SAdd(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis sadd error: %w", err)
	}
	return result, nil
}

// SRem 从集合移除元素
func (r *RedisCache) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(members) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(members))
	for i, v := range members {
		vals[i] = v
	}
	
	result, err := r.getUniversalClient().SRem(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis srem error: %w", err)
	}
	return result, nil
}

// SMembers 获取集合所有成员
func (r *RedisCache) SMembers(ctx context.Context, key string) ([]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().SMembers(ctx, fullKey).Result()
	if err != nil {
		return nil, fmt.Errorf("redis smembers error: %w", err)
	}
	return result, nil
}

// SIsMember 检查元素是否在集合中
func (r *RedisCache) SIsMember(ctx context.Context, key string, member string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().SIsMember(ctx, fullKey, member).Result()
	if err != nil {
		return false, fmt.Errorf("redis sismember error: %w", err)
	}
	return result, nil
}

// ZAdd 向有序集合添加元素
func (r *RedisCache) ZAdd(ctx context.Context, key string, score float64, member string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	z := redis.Z{Score: score, Member: member}
	result, err := r.getUniversalClient().ZAdd(ctx, fullKey, z).Result()
	if err != nil {
		return 0, fmt.Errorf("redis zadd error: %w", err)
	}
	return result, nil
}

// ZRem 从有序集合移除元素
func (r *RedisCache) ZRem(ctx context.Context, key string, members ...string) (int64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	if len(members) == 0 {
		return 0, nil
	}
	
	fullKey := r.buildKey(key)
	// 转换为interface{}切片
	vals := make([]interface{}, len(members))
	for i, v := range members {
		vals[i] = v
	}
	
	result, err := r.getUniversalClient().ZRem(ctx, fullKey, vals...).Result()
	if err != nil {
		return 0, fmt.Errorf("redis zrem error: %w", err)
	}
	return result, nil
}

// ZScore 获取有序集合成员的分数
func (r *RedisCache) ZScore(ctx context.Context, key string, member string) (float64, error) {
	if key == "" {
		return 0, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().ZScore(ctx, fullKey, member).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // 成员不存在时返回0
		}
		return 0, fmt.Errorf("redis zscore error: %w", err)
	}
	return result, nil
}

// ZRange 获取有序集合指定范围的成员
func (r *RedisCache) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if key == "" {
		return nil, fmt.Errorf("缓存键不能为空")
	}
	
	fullKey := r.buildKey(key)
	result, err := r.getUniversalClient().ZRange(ctx, fullKey, start, stop).Result()
	if err != nil {
		return nil, fmt.Errorf("redis zrange error: %w", err)
	}
	return result, nil
}
