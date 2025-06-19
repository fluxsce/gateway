package cache

import (
	"context"
	"fmt"
	"gohub/pkg/cache/redis"
	"gohub/pkg/config"
	"gohub/pkg/logger"
	"time"
)

// CacheRootConfig 缓存配置根结构
type CacheRootConfig struct {
	Default     string                       `mapstructure:"default"`     // 默认连接名称
	Connections map[string]*redis.RedisConfig `mapstructure:"connections"` // 连接配置映射，直接使用redis.RedisConfig
}

// LoadAllCacheConnections 从配置文件加载所有缓存连接
// 解析配置文件中的所有Redis连接配置，创建并注册缓存实例
// 只有enabled为true的连接才会被创建
// 参数:
//   configPath: 数据库配置文件路径（包含缓存配置）
// 返回:
//   map[string]Cache: 连接名称到缓存实例的映射
//   error: 加载失败时返回错误信息
func LoadAllCacheConnections(configPath string) (map[string]Cache, error) {
	// 首先加载配置文件
	if err := config.LoadConfigFile(configPath); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 解析缓存配置
	var cacheConfig struct {
		Default     string                       `mapstructure:"default"`
		Connections map[string]*redis.RedisConfig `mapstructure:"connections"`
	}

	if err := config.GetSection("cache", &cacheConfig); err != nil {
		return nil, fmt.Errorf("解析缓存配置失败: %w", err)
	}

	// 验证配置
	if len(cacheConfig.Connections) == 0 {
		return nil, fmt.Errorf("未找到缓存连接配置")
	}

	// 获取全局缓存管理器
	manager := GetGlobalManager()
	connections := make(map[string]Cache)

	// 遍历所有配置，创建启用的连接
	for name, connConfig := range cacheConfig.Connections {
		// 检查连接是否启用
		if !connConfig.Enabled {
			logger.Info("跳过未启用的缓存连接", "name", name)
			continue
		}

		// 解析字符串配置到Duration类型
		if err := parseStringConfigs(connConfig); err != nil {
			return nil, fmt.Errorf("解析连接配置 '%s' 失败: %w", name, err)
		}

		// 设置默认值
		connConfig.SetDefaults()

		// 验证配置
		if err := connConfig.Validate(); err != nil {
			return nil, fmt.Errorf("验证连接配置 '%s' 失败: %w", name, err)
		}

		// 创建Redis缓存实例
		cache, err := createRedisCache(name, connConfig)
		if err != nil {
			return nil, fmt.Errorf("创建缓存连接 '%s' 失败: %w", name, err)
		}

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := cache.Ping(ctx); err != nil {
			cancel()
			logger.Error("缓存连接测试失败", "name", name, "error", err)
			// 关闭连接
			cache.Close()
			return nil, fmt.Errorf("缓存连接 '%s' 测试失败: %w", name, err)
		}
		cancel()

		// 添加到管理器
		if err := manager.AddCache(name, cache); err != nil {
			cache.Close()
			return nil, fmt.Errorf("注册缓存实例 '%s' 失败: %w", name, err)
		}

		// 存储连接映射
		connections[name] = cache

		// 记录成功日志
		logger.Info("缓存连接创建成功",
			"name", name,
			"mode", string(connConfig.Mode),
			"connection_string", connConfig.GetConnectionString())
	}

	// 设置默认连接
	if cacheConfig.Default != "" {
		if defaultCache, exists := connections[cacheConfig.Default]; exists {
			// 将默认连接添加为"default"名称
			if err := manager.AddCache("default", defaultCache); err != nil {
				logger.Warn("设置默认缓存连接失败", "name", cacheConfig.Default, "error", err)
			} else {
				logger.Info("设置默认缓存连接", "name", cacheConfig.Default)
			}
		} else {
			logger.Warn("指定的默认缓存连接不存在", "name", cacheConfig.Default)
		}
	}

	logger.Info("缓存系统初始化完成",
		"total_connections", len(connections),
		"default_connection", cacheConfig.Default)

	return connections, nil
}

// createRedisCache 创建Redis缓存实例
// 直接使用完整的Redis配置创建实例
// 参数:
//   name: 连接名称
//   config: Redis配置
// 返回:
//   Cache: Redis缓存实例
//   error: 创建失败时返回错误信息
func createRedisCache(name string, config *redis.RedisConfig) (Cache, error) {
	// 记录连接创建信息
	logger.Debug("创建Redis缓存连接",
		"name", name,
		"mode", string(config.Mode),
		"config", config.String())

	// 直接使用Redis配置创建实例
	redisCache, err := redis.NewRedisCache(config)
	if err != nil {
		return nil, fmt.Errorf("创建Redis实例失败: %w", err)
	}

	return redisCache, nil
}

// parseStringConfigs 解析字符串格式的配置
// 将YAML配置中的字符串格式时间转换为time.Duration
func parseStringConfigs(config *redis.RedisConfig) error {
	// 这个函数用于处理YAML配置文件中可能存在的字符串格式时间配置
	// 由于我们现在直接使用redis.RedisConfig，大部分解析已由mapstructure完成
	// 这里只需要处理特殊情况

	// 如果配置中有字符串格式的地址需要解析，在这里处理
	// 目前redis.RedisConfig已经包含了所有需要的字段，所以这里暂时不需要额外处理

	return nil
}

// GetConnectionInfo 获取连接信息
// 返回当前所有活跃连接的详细信息
func GetConnectionInfo() map[string]map[string]interface{} {
	manager := GetGlobalManager()
	return manager.Stats()
}

// HealthCheck 健康检查
// 检查所有缓存连接的健康状态
func HealthCheck(ctx context.Context) map[string]error {
	manager := GetGlobalManager()
	connections := manager.ListCaches()
	results := make(map[string]error)

	for _, name := range connections {
		cache := manager.GetCache(name)
		if cache == nil {
			results[name] = fmt.Errorf("缓存实例不存在")
			continue
		}

		// 执行ping检查
		if err := cache.Ping(ctx); err != nil {
			results[name] = fmt.Errorf("ping失败: %w", err)
		} else {
			results[name] = nil // 健康
		}
	}

	return results
}

// ReloadConnection 重新加载指定连接
// 关闭现有连接并使用新配置重新创建
func ReloadConnection(name string, config *redis.RedisConfig) error {
	manager := GetGlobalManager()

	// 移除现有连接
	if err := manager.RemoveCache(name); err != nil {
		logger.Warn("移除现有缓存连接失败", "name", name, "error", err)
	}

	// 创建新连接
	cache, err := createRedisCache(name, config)
	if err != nil {
		return fmt.Errorf("重新创建缓存连接失败: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cache.Ping(ctx); err != nil {
		cache.Close()
		return fmt.Errorf("新连接测试失败: %w", err)
	}

	// 添加到管理器
	if err := manager.AddCache(name, cache); err != nil {
		cache.Close()
		return fmt.Errorf("注册新缓存实例失败: %w", err)
	}

	logger.Info("缓存连接重新加载成功", "name", name)
	return nil
}

// CloseAllConnections 关闭所有缓存连接
// 应用关闭时调用，清理所有缓存连接资源
// 返回:
//   error: 关闭过程中的第一个错误
func CloseAllConnections() error {
	return CloseAllCaches()
} 