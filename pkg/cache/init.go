package cache

import (
	"context"
	"fmt"
	"gohub/pkg/cache/memory"
	"gohub/pkg/cache/redis"
	"gohub/pkg/config"
	"gohub/pkg/logger"
	"time"
)

// CacheConnectionConfig 缓存连接配置的通用结构
type CacheConnectionConfig struct {
	Type   string      `mapstructure:"type"`   // 缓存类型: "redis" 或 "memory"
	Config interface{} `mapstructure:"config"` // 具体的配置，根据类型解析
}

// CacheRootConfig 缓存配置根结构
type CacheRootConfig struct {
	Default     string                        `mapstructure:"default"`     // 默认连接名称
	Connections map[string]*CacheConnectionConfig `mapstructure:"connections"` // 连接配置映射
}

// LoadAllCacheConnections 从配置文件加载所有缓存连接
// 解析配置文件中的所有缓存连接配置，支持Redis和内存缓存
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
	var cacheConfig CacheRootConfig

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
		// 使用工厂方法创建缓存实例
		cache, err := createCacheFromConfig(name, connConfig)
		if err != nil {
			return nil, fmt.Errorf("创建缓存连接 '%s' 失败: %w", name, err)
		}

		// 如果返回nil，说明连接未启用，跳过处理
		if cache == nil {
			logger.Info("跳过未启用的缓存连接", "name", name, "type", connConfig.Type)
			continue
		}

		// 测试连接
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := cache.Ping(ctx); err != nil {
			cancel()
			logger.Error("缓存连接测试失败", "name", name, "type", connConfig.Type, "error", err)
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
			"type", connConfig.Type,
			"stats", cache.Stats())
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

	// 检查是否有有效的连接
	if len(connections) == 0 {
		logger.Warn("没有启用的缓存连接")
		return connections, nil // 返回空映射，但不报错
	}

	logger.Info("缓存系统初始化完成",
		"active_connections", len(connections),
		"default_connection", cacheConfig.Default)

	return connections, nil
}

// createCacheFromConfig 根据配置类型创建缓存实例
// 这是统一的入口点，负责分发到各个模块的工厂方法
// 
// 注意：此函数解决了Go语言中类型化nil的问题
// 当工厂方法返回(*ConcreteType)(nil)时，赋值给接口变量后 != nil，但调用方法会空指针异常
// 通过在此处统一检查并转换为真正的nil，避免了接口中包含类型化nil的问题
func createCacheFromConfig(name string, connConfig *CacheConnectionConfig) (Cache, error) {
	// 验证配置类型
	if connConfig.Type == "" {
		return nil, fmt.Errorf("缓存类型不能为空")
	}

	if connConfig.Config == nil {
		return nil, fmt.Errorf("缓存配置不能为空")
	}

	// 根据类型调用对应的工厂方法
	switch connConfig.Type {
	case "redis":
		// 使用Redis模块的工厂方法
		redisCache, err := redis.CreateFromConfig(name, connConfig.Config)
		if err != nil {
			return nil, fmt.Errorf("Redis工厂方法创建失败: %w", err)
		}
		// 检查类型化nil：当redisCache为(*RedisCache)(nil)时，转换为真正的nil
		if redisCache == nil {
			return nil, nil
		}
		return redisCache, nil

	case "memory":
		// 使用内存缓存模块的工厂方法
		memoryCache, err := memory.CreateFromConfig(name, connConfig.Config)
		if err != nil {
			return nil, fmt.Errorf("内存缓存工厂方法创建失败: %w", err)
		}
		// 检查类型化nil：当memoryCache为(*MemoryCache)(nil)时，转换为真正的nil
		if memoryCache == nil {
			return nil, nil
		}
		return memoryCache, nil

	default:
		return nil, fmt.Errorf("不支持的缓存类型 '%s'", connConfig.Type)
	}
}

// ValidateConnectionConfig 验证连接配置的有效性
// 在创建连接前进行配置验证，提前发现配置问题
// 对于未启用的连接，跳过验证
func ValidateConnectionConfig(name string, connConfig *CacheConnectionConfig) error {
	if connConfig.Type == "" {
		return fmt.Errorf("连接 '%s' 的缓存类型不能为空", name)
	}

	if connConfig.Config == nil {
		return fmt.Errorf("连接 '%s' 的配置不能为空", name)
	}

	// 检查是否启用，如果未启用则跳过验证
	if configMap, ok := connConfig.Config.(map[string]interface{}); ok {
		if enabled, exists := configMap["enabled"]; exists {
			if enabledBool, ok := enabled.(bool); ok && !enabledBool {
				logger.Debug("跳过未启用连接的验证", "name", name)
				return nil // 未启用的连接跳过验证
			}
		}
	}

	// 根据类型进行特定验证
	switch connConfig.Type {
	case "redis":
		// 使用Redis模块的验证方法
		return redis.ValidateConfig(connConfig.Config)

	case "memory":
		// 使用内存缓存模块的验证方法
		return memory.ValidateConfig(connConfig.Config)

	default:
		return fmt.Errorf("连接 '%s' 使用了不支持的缓存类型 '%s'", name, connConfig.Type)
	}
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
func ReloadConnection(name string, connConfig *CacheConnectionConfig) error {
	manager := GetGlobalManager()

	// 验证新配置
	if err := ValidateConnectionConfig(name, connConfig); err != nil {
		return fmt.Errorf("新配置验证失败: %w", err)
	}

	// 创建新连接
	newCache, err := createCacheFromConfig(name, connConfig)
	if err != nil {
		return fmt.Errorf("创建新连接失败: %w", err)
	}

	// 测试新连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := newCache.Ping(ctx); err != nil {
		newCache.Close()
		return fmt.Errorf("新连接测试失败: %w", err)
	}

	// 移除旧连接（这会自动关闭旧连接）
	if err := manager.RemoveCache(name); err != nil {
		logger.Warn("移除旧缓存连接失败", "name", name, "error", err)
	}

	// 添加新连接
	if err := manager.AddCache(name, newCache); err != nil {
		newCache.Close()
		return fmt.Errorf("注册新连接失败: %w", err)
	}

	logger.Info("缓存连接重新加载成功", "name", name, "type", connConfig.Type)
	return nil
}

// GetSupportedCacheTypes 获取支持的缓存类型列表
// 用于配置验证和文档生成
func GetSupportedCacheTypes() []string {
	return []string{"redis", "memory"}
}

// GetDefaultConfigs 获取所有缓存类型的默认配置
// 用于生成配置模板
func GetDefaultConfigs() map[string]interface{} {
	return map[string]interface{}{
		"redis":  redis.GetDefaultConfig(),
		"memory": memory.GetDefaultConfig(),
	}
}

// CloseAllConnections 关闭所有缓存连接
// 应用关闭时调用，清理所有缓存连接资源
// 返回:
//   error: 关闭过程中的第一个错误
func CloseAllConnections() error {
	return CloseAllCaches()
}