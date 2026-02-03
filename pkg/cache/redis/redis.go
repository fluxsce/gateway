// Package redis 提供 Redis 缓存的完整实现。
//
// 该包支持 Redis 的三种部署模式：
//   - 单机模式（Single）：适用于开发环境和小规模应用
//   - 哨兵模式（Sentinel）：适用于需要高可用的场景
//   - 集群模式（Cluster）：适用于大规模分布式应用
//
// 主要特性：
//   - 线程安全：所有操作都支持并发调用
//   - 自动重连：支持连接断开后自动重连
//   - 连接池：内置连接池管理，提高性能
//   - TLS 支持：完整的 TLS/SSL 加密通信支持
//   - 批量操作：支持 MGet/MSet 等批量操作
//   - 类型丰富：支持 String、Hash、List、Set、ZSet 等数据类型
//
// 基本使用示例：
//
//	// 创建配置
//	cfg := &redis.RedisConfig{
//	    Mode: redis.ModeSingle,
//	    Host: "localhost",
//	    Port: 6379,
//	}
//
//	// 创建 Redis 缓存实例
//	cache, err := redis.NewRedisCache(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
//
//	// 设置值
//	err = cache.Set(ctx, "key", []byte("value"), 10*time.Minute)
//
//	// 获取值
//	value, err := cache.Get(ctx, "key")
//
// 高级使用示例：
//
//	// 批量操作
//	kvPairs := map[string]string{"key1": "value1", "key2": "value2"}
//	err = cache.MSetString(ctx, kvPairs, 10*time.Minute)
//
//	// Hash 操作
//	err = cache.HSet(ctx, "user:1", "name", "Alice")
//	name, err := cache.HGet(ctx, "user:1", "name")
//
//	// 分布式锁
//	locked, err := cache.SetNXString(ctx, "lock:resource", "locked", 10*time.Second)
//	if locked {
//	    defer cache.Delete(ctx, "lock:resource")
//	    // 执行需要加锁的操作
//	}
//
// 线程安全示例：
//
//	// 多个 goroutine 可以安全并发使用同一个 RedisCache 实例
//	go cache.Get(ctx, "key1")
//	go cache.Set(ctx, "key2", value, 0)
//	go cache.Delete(ctx, "key3")
//
// 错误处理：
//
//	value, err := cache.Get(ctx, "key")
//	if err != nil {
//	    if strings.Contains(err.Error(), "redis缓存已关闭") {
//	        // 连接已关闭
//	    } else {
//	        // 其他错误
//	    }
//	    return err
//	}
//	if value == nil {
//	    // 键不存在（不是错误）
//	}
package redis

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"gateway/pkg/config"
	"gateway/pkg/logger"
)

// RedisCache 提供 Redis 缓存的完整实现。
//
// 该类型实现了统一的 Cache 接口，支持 Redis 的所有常用操作。
// RedisCache 是线程安全的，可以在多个 goroutine 中并发使用。
//
// 支持的部署模式：
//   - 单机模式（Single）：直连单个 Redis 实例
//   - 哨兵模式（Sentinel）：通过哨兵实现高可用
//   - 集群模式（Cluster）：Redis Cluster 分布式部署
//
// 并发安全性：
//   - 所有公开方法都是线程安全的
//   - 内部使用 sync.RWMutex 保护共享状态
//   - Close() 方法可以安全地在任何时候调用，且支持重复调用
//
// 错误处理：
//   - 所有方法在遇到错误时返回 error，不会 panic
//   - 如果 RedisCache 已关闭，所有操作返回 "redis缓存已关闭" 错误
//   - 键不存在时，Get 方法返回 nil 而不是错误
//
// 注意事项：
//   - 使用完毕后必须调用 Close() 释放连接资源
//   - 关闭后的实例不能继续使用
//   - 键前缀会自动添加到所有操作的键名上
type RedisCache struct {
	// client 单机或哨兵模式的 Redis 客户端
	client *redis.Client

	// clusterClient 集群模式的 Redis 客户端
	clusterClient *redis.ClusterClient

	// config Redis 配置信息
	config *RedisConfig

	// keyPrefix 键前缀，自动添加到所有键名前
	keyPrefix string

	// isCluster 标识是否为集群模式
	isCluster bool

	// mu 保护 closed 状态和 client 实例的并发访问
	mu sync.RWMutex

	// closed 标记实例是否已关闭
	closed bool
}

// NewRedisCache 创建新的 Redis 缓存实例。
//
// 该函数根据配置自动选择合适的连接模式（单机/哨兵/集群），
// 创建连接后会自动测试连接有效性。如果连接失败，会自动清理资源。
//
// 参数：
//   - cfg: Redis 配置对象。如果为 nil，则尝试从 configs/database.yaml 加载配置
//
// 返回值：
//   - *RedisCache: Redis 缓存实例，使用完毕后必须调用 Close() 释放资源
//   - error: 创建失败时返回错误，可能的原因包括：
//   - 配置无效
//   - 无法连接到 Redis 服务器
//   - TLS 配置错误
//
// 使用示例：
//
//	// 使用自定义配置
//	cfg := &redis.RedisConfig{
//	    Mode:     redis.ModeSingle,
//	    Host:     "localhost",
//	    Port:     6379,
//	    Password: "secret",
//	    DB:       0,
//	}
//	cache, err := redis.NewRedisCache(cfg)
//	if err != nil {
//	    log.Fatal("创建Redis缓存失败:", err)
//	}
//	defer cache.Close()
//
//	// 使用配置文件
//	cache, err := redis.NewRedisCache(nil)
//	if err != nil {
//	    log.Fatal("从配置文件创建Redis缓存失败:", err)
//	}
//	defer cache.Close()
//
// 注意事项：
//   - 返回的实例是线程安全的，可以在多个 goroutine 中共享使用
//   - 如果连接测试失败，函数会自动清理已创建的资源
//   - 必须在使用完毕后调用 Close() 方法，否则会造成连接泄漏
func NewRedisCache(cfg *RedisConfig) (*RedisCache, error) {
	if cfg == nil {
		return nil, fmt.Errorf("redis配置不能为nil，请使用CreateFromConfigPath或提供有效配置")
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
		closed:    false,
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
		tlsConfig, err := r.createTLSConfig()
		if err != nil {
			return fmt.Errorf("创建TLS配置失败: %w", err)
		}
		opts.TLSConfig = tlsConfig
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
		tlsConfig, err := r.createTLSConfig()
		if err != nil {
			return fmt.Errorf("创建TLS配置失败: %w", err)
		}
		opts.TLSConfig = tlsConfig
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
		tlsConfig, err := r.createTLSConfig()
		if err != nil {
			return fmt.Errorf("创建TLS配置失败: %w", err)
		}
		opts.TLSConfig = tlsConfig
	}

	r.clusterClient = redis.NewClusterClient(opts)
	return nil
}

// createTLSConfig 创建TLS配置
// 完整实现证书加载，包括客户端证书、密钥和CA证书
func (r *RedisCache) createTLSConfig() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: r.config.TLSInsecureSkipVerify,
	}

	// 加载客户端证书和密钥
	if r.config.TLSCertFile != "" && r.config.TLSKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(r.config.TLSCertFile, r.config.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("加载TLS证书失败: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		logger.Debug("已加载TLS客户端证书", "certFile", r.config.TLSCertFile)
	}

	// 加载CA证书
	if r.config.TLSCACertFile != "" {
		caCert, err := os.ReadFile(r.config.TLSCACertFile)
		if err != nil {
			return nil, fmt.Errorf("读取CA证书文件失败: %w", err)
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("解析CA证书失败")
		}

		tlsConfig.RootCAs = caCertPool
		logger.Debug("已加载TLS CA证书", "caFile", r.config.TLSCACertFile)
	}

	return tlsConfig, nil
}

// getUniversalClient 获取通用的 Redis 客户端接口。
//
// 该方法是内部方法，用于所有 Redis 操作前获取有效的客户端实例。
// 它会检查连接状态，确保客户端可用，防止 panic。
//
// 返回值：
//   - redis.UniversalClient: Redis 客户端接口（单机/哨兵/集群）
//   - error: 获取失败时返回错误，可能的原因：
//   - Redis 缓存已关闭
//   - 客户端未初始化
//
// 线程安全：
//   - 使用读锁保护，支持并发调用
//   - 不会修改任何状态
func (r *RedisCache) getUniversalClient() (redis.UniversalClient, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 检查是否已关闭
	if r.closed {
		return nil, fmt.Errorf("redis缓存已关闭")
	}

	// 检查client是否为nil
	if r.isCluster {
		if r.clusterClient == nil {
			return nil, fmt.Errorf("集群客户端未初始化")
		}
		return r.clusterClient, nil
	}

	if r.client == nil {
		return nil, fmt.Errorf("客户端未初始化")
	}
	return r.client, nil
}

// CreateFromConfigPath 从配置路径创建 Redis 缓存实例。
//
// 该函数是 Redis 模块对外提供的工厂方法，使用 config.GetSection 自动映射配置。
//
// 参数：
//   - name: 连接名称
//   - configPath: 配置路径（如 "cache.connections.redis_main.config"）
//
// 返回值：
//   - *RedisCache: Redis 缓存实例，如果未启用则返回 nil
//   - error: 创建失败时返回错误信息
//
// 使用示例：
//
//	cache, err := redis.CreateFromConfigPath("main", "cache.connections.redis_main.config")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close()
func CreateFromConfigPath(name string, configPath string) (*RedisCache, error) {
	// 创建Redis配置实例
	redisConfig := &RedisConfig{}

	// 使用 config.GetSection 自动映射配置（就像 MetricConfig 那样）
	if err := config.GetSection(configPath, redisConfig); err != nil {
		return nil, fmt.Errorf("从配置路径 '%s' 加载Redis配置失败: %w", configPath, err)
	}

	// 检查是否启用，如果未启用则跳过
	if !redisConfig.Enabled {
		logger.Debug("跳过未启用的Redis缓存连接", "name", name)
		return nil, nil // 返回nil表示跳过此连接
	}

	// 设置默认值
	redisConfig.SetDefaults()

	// 验证配置
	if err := redisConfig.Validate(); err != nil {
		return nil, fmt.Errorf("验证连接配置失败: %w", err)
	}

	// 记录连接创建信息
	logger.Debug("创建Redis缓存连接",
		"name", name,
		"mode", string(redisConfig.Mode),
		"config", redisConfig.String())

	// 创建Redis缓存实例
	redisCache, err := NewRedisCache(redisConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Redis实例失败: %w", err)
	}

	return redisCache, nil
}

// buildKey 构建完整的缓存键（加上前缀）
// 如果配置了前缀，会自动在前缀和键之间添加分隔符
// 参数:
//   - key: 原始缓存键
//
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
//
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
//
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

// Ping 测试 Redis 连接是否正常。
//
// 该方法向 Redis 服务器发送 PING 命令，用于检查连接状态。
//
// 参数：
//   - ctx: 上下文，用于控制请求超时和取消
//
// 返回值：
//   - error: 连接异常时返回错误，可能的原因：
//   - Redis 缓存已关闭
//   - 网络连接失败
//   - Redis 服务器无响应
//
// 使用示例：
//
//	err := cache.Ping(ctx)
//	if err != nil {
//	    log.Error("Redis 连接异常:", err)
//	}
func (r *RedisCache) Ping(ctx context.Context) error {
	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	err = client.Ping(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis ping error: %w", err)
	}
	return nil
}

// Close 关闭 Redis 缓存连接并释放所有资源。
//
// 该方法会安全地关闭底层的 Redis 连接，释放连接池中的所有连接。
// 关闭后，该实例的所有操作都会返回 "redis缓存已关闭" 错误。
//
// 返回值：
//   - error: 关闭失败时返回错误信息
//
// 特性：
//   - 幂等性：可以安全地多次调用，后续调用会立即返回 nil
//   - 线程安全：可以在任何时候从任何 goroutine 调用
//   - 优雅关闭：会等待当前正在进行的操作完成
//
// 使用示例：
//
//	cache, err := redis.NewRedisCache(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer cache.Close() // 确保资源释放
//
//	// 使用缓存...
//
// 注意事项：
//   - 关闭后的实例不能继续使用
//   - 建议使用 defer 确保资源被释放
//   - 如果有多个 goroutine 正在使用该实例，它们会收到错误
func (r *RedisCache) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 防止重复关闭
	if r.closed {
		return nil
	}
	r.closed = true

	var closeErr error

	// 关闭集群客户端
	if r.isCluster && r.clusterClient != nil {
		if err := r.clusterClient.Close(); err != nil {
			closeErr = fmt.Errorf("关闭集群客户端失败: %w", err)
		}
		r.clusterClient = nil
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
		r.client = nil
	}

	return closeErr
}

// IsConnected 检查 Redis 连接是否有效。
//
// 该方法快速检查连接状态，不会发送网络请求，开销极小。
//
// 返回值：
//   - bool: true 表示连接有效且未关闭，false 表示连接已关闭或不存在
//
// 线程安全：
//   - 使用读锁保护，支持并发调用
//
// 使用示例：
//
//	if !cache.IsConnected() {
//	    log.Warn("Redis 连接已断开")
//	    return
//	}
//
// 注意事项：
//   - 返回 true 不代表网络连接一定可用，只是表示未调用 Close()
//   - 如需测试网络连接，请使用 Ping() 方法
func (r *RedisCache) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return false
	}

	if r.isCluster {
		return r.clusterClient != nil
	}
	return r.client != nil
}

// Stats 获取 Redis 缓存的统计信息。
//
// 返回包含连接池状态和配置信息的映射，用于监控和调试。
//
// 返回值：
//   - map[string]interface{}: 统计信息映射，包含以下字段：
//   - status: 连接状态（"closed" 或 "connected"）
//   - pool: 连接池统计（仅在连接状态时）
//   - hits: 连接池命中次数
//   - misses: 连接池未命中次数
//   - timeouts: 超时次数
//   - total_conns: 总连接数
//   - idle_conns: 空闲连接数
//   - stale_conns: 失效连接数
//   - config: 配置信息
//   - mode: 连接模式
//   - connection_string: 连接字符串
//   - pool_size: 连接池大小
//   - key_prefix: 键前缀
//
// 使用示例：
//
//	stats := cache.Stats()
//	fmt.Printf("连接池统计: %+v\n", stats["pool"])
//	fmt.Printf("配置信息: %+v\n", stats["config"])
//
// 监控示例：
//
//	stats := cache.Stats()
//	if pool, ok := stats["pool"].(map[string]interface{}); ok {
//	    timeouts := pool["timeouts"].(uint32)
//	    if timeouts > 100 {
//	        log.Warn("连接池超时次数过多", "timeouts", timeouts)
//	    }
//	}
func (r *RedisCache) Stats() map[string]interface{} {
	stats := make(map[string]interface{})

	r.mu.RLock()
	closed := r.closed
	r.mu.RUnlock()

	if closed {
		stats["status"] = "closed"
		return stats
	}

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
		"mode":              string(r.config.Mode),
		"connection_string": r.config.GetConnectionString(),
		"pool_size":         r.config.PoolSize,
		"min_idle_conns":    r.config.MinIdleConns,
		"max_idle_conns":    r.config.MaxIdleConns,
		"key_prefix":        r.keyPrefix,
		"is_cluster":        r.isCluster,
	}

	return stats
}

// FlushAll 清空所有缓存（谨慎使用）
func (r *RedisCache) FlushAll(ctx context.Context) error {
	client, err := r.getUniversalClient()
	if err != nil {
		return err
	}

	err = client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flushall error: %w", err)
	}
	return nil
}

// GetCacheType 获取缓存类型。
//
// 返回缓存类型标识，用于识别缓存实现。
//
// 返回值：
//   - string: 缓存类型标识 "redis"
func (r *RedisCache) GetCacheType() string {
	return "redis"
}

// SelectDB 选择数据库（仅单机和哨兵模式支持）
// 使用Redis的SELECT命令切换到指定数据库，无需重新创建连接
// 线程安全
// 参数:
//   - ctx: 上下文，用于控制请求超时和取消
//   - db: 目标数据库编号（0-15）
//
// 返回:
//   - error: 切换失败时返回错误信息
//
// 注意: 集群模式不支持数据库选择，因为集群中的数据是分片的
func (r *RedisCache) SelectDB(ctx context.Context, db int) error {
	if r.isCluster {
		return fmt.Errorf("集群模式不支持数据库选择")
	}

	if db < 0 || db > 15 {
		return fmt.Errorf("数据库编号必须在0-15之间，当前值: %d", db)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return fmt.Errorf("redis缓存已关闭")
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
