package redis

import (
	"fmt"
	"gohub/pkg/logger"
	"time"
)

// CreateFromConfig 从配置数据创建Redis缓存实例
// 这是Redis模块对外提供的工厂方法，用于从通用配置创建Redis缓存
// 参数:
//   name: 连接名称
//   configData: 配置数据，通常来自YAML配置文件
// 返回:
//   *RedisCache: Redis缓存实例
//   error: 创建失败时返回错误信息
func CreateFromConfig(name string, configData interface{}) (*RedisCache, error) {
	// 将interface{}转换为Redis配置
	redisConfig := &RedisConfig{}
	
	// 使用类型断言来转换配置
	if configMap, ok := configData.(map[string]interface{}); ok {
		if err := mapConfigToRedisConfig(configMap, redisConfig); err != nil {
			return nil, fmt.Errorf("映射配置到Redis配置失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("无效的配置数据类型，期望 map[string]interface{}")
	}

	// 检查是否启用，如果未启用则跳过
	if !redisConfig.Enabled {
		logger.Warn("跳过未启用的Redis缓存连接", "name", name)
		return nil, nil // 返回nil表示跳过此连接
	}

	// 解析字符串配置到Duration类型
	if err := parseRedisStringConfigs(redisConfig); err != nil {
		return nil, fmt.Errorf("解析连接配置失败: %w", err)
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

// mapConfigToRedisConfig 将通用配置映射到Redis配置
func mapConfigToRedisConfig(configMap map[string]interface{}, redisConfig *RedisConfig) error {
	// 基础配置
	if enabled, ok := configMap["enabled"].(bool); ok {
		redisConfig.Enabled = enabled
	}
	
	if mode, ok := configMap["mode"].(string); ok {
		redisConfig.Mode = ConnectionMode(mode)
	}
	
	if host, ok := configMap["host"].(string); ok {
		redisConfig.Host = host
	}
	
	if port, ok := configMap["port"].(int); ok {
		redisConfig.Port = port
	}
	
	if password, ok := configMap["password"].(string); ok {
		redisConfig.Password = password
	}
	
	if database, ok := configMap["database"]; ok {
		switch v := database.(type) {
		case int:
			redisConfig.DB = v
		case float64:
			redisConfig.DB = int(v)
		}
	}

	// 连接池配置
	if poolSize, ok := configMap["pool_size"]; ok {
		switch v := poolSize.(type) {
		case int:
			redisConfig.PoolSize = v
		case float64:
			redisConfig.PoolSize = int(v)
		}
	}
	
	if minIdleConns, ok := configMap["min_idle_connections"]; ok {
		switch v := minIdleConns.(type) {
		case int:
			redisConfig.MinIdleConns = v
		case float64:
			redisConfig.MinIdleConns = int(v)
		}
	}
	
	if maxRetries, ok := configMap["max_retries"]; ok {
		switch v := maxRetries.(type) {
		case int:
			redisConfig.MaxRetries = v
		case float64:
			redisConfig.MaxRetries = int(v)
		}
	}

	// 时间配置（字符串格式）
	if dialTimeout, ok := configMap["dial_timeout"].(string); ok {
		if duration, err := time.ParseDuration(dialTimeout); err == nil {
			redisConfig.DialTimeout = duration
		}
	}
	
	if readTimeout, ok := configMap["read_timeout"].(string); ok {
		if duration, err := time.ParseDuration(readTimeout); err == nil {
			redisConfig.ReadTimeout = duration
		}
	}
	
	if writeTimeout, ok := configMap["write_timeout"].(string); ok {
		if duration, err := time.ParseDuration(writeTimeout); err == nil {
			redisConfig.WriteTimeout = duration
		}
	}
	
	if poolTimeout, ok := configMap["pool_timeout"].(string); ok {
		if duration, err := time.ParseDuration(poolTimeout); err == nil {
			redisConfig.PoolTimeout = duration
		}
	}
	
	if idleTimeout, ok := configMap["idle_timeout"].(string); ok {
		if duration, err := time.ParseDuration(idleTimeout); err == nil {
			redisConfig.IdleTimeout = int64(duration.Milliseconds())
		}
	}

	// 集群配置
	if clusterAddrs, ok := configMap["cluster_addrs"].([]interface{}); ok {
		addrs := make([]string, len(clusterAddrs))
		for i, addr := range clusterAddrs {
			if addrStr, ok := addr.(string); ok {
				addrs[i] = addrStr
			}
		}
		redisConfig.ClusterAddrs = addrs
	}

	// 哨兵配置
	if sentinelAddrs, ok := configMap["sentinel_addrs"].([]interface{}); ok {
		addrs := make([]string, len(sentinelAddrs))
		for i, addr := range sentinelAddrs {
			if addrStr, ok := addr.(string); ok {
				addrs[i] = addrStr
			}
		}
		redisConfig.SentinelAddrs = addrs
	}
	
	if masterName, ok := configMap["master_name"].(string); ok {
		redisConfig.MasterName = masterName
	}

	// TLS配置
	if enableTLS, ok := configMap["enable_tls"].(bool); ok {
		redisConfig.TLSEnabled = enableTLS
	}
	
	if insecureSkipVerify, ok := configMap["insecure_skip_verify"].(bool); ok {
		redisConfig.TLSInsecureSkipVerify = insecureSkipVerify
	}
	
	if certFile, ok := configMap["cert_file"].(string); ok {
		redisConfig.TLSCertFile = certFile
	}
	
	if keyFile, ok := configMap["key_file"].(string); ok {
		redisConfig.TLSKeyFile = keyFile
	}
	
	if caFile, ok := configMap["ca_file"].(string); ok {
		redisConfig.TLSCACertFile = caFile
	}

	return nil
}

// parseRedisStringConfigs 解析Redis字符串格式的配置
// 处理YAML配置文件中可能存在的特殊字符串格式配置
func parseRedisStringConfigs(config *RedisConfig) error {
	// 目前redis.RedisConfig已经包含了所有需要的字段
	// 大部分解析已在mapConfigToRedisConfig中完成
	// 这里处理特殊情况或验证

	// 验证集群地址格式
	if config.Mode == ModeCluster && len(config.ClusterAddrs) == 0 {
		return fmt.Errorf("集群模式需要至少一个集群地址")
	}

	// 验证哨兵配置
	if config.Mode == ModeSentinel {
		if len(config.SentinelAddrs) == 0 {
			return fmt.Errorf("哨兵模式需要至少一个哨兵地址")
		}
		if config.MasterName == "" {
			return fmt.Errorf("哨兵模式需要指定主节点名称")
		}
	}

	return nil
}

// ValidateConfig 验证Redis配置的有效性
// 这是一个额外的验证函数，用于在创建实例前进行深度验证
func ValidateConfig(configData interface{}) error {
	configMap, ok := configData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的配置数据类型")
	}

	// 检查必需字段
	if enabled, exists := configMap["enabled"]; !exists || enabled != true {
		return fmt.Errorf("Redis缓存未启用或配置缺失")
	}

	// 验证连接模式
	if mode, ok := configMap["mode"].(string); ok {
		switch ConnectionMode(mode) {
		case ModeSingle, ModeCluster, ModeSentinel:
			// 有效模式
		default:
			return fmt.Errorf("不支持的Redis模式: %s", mode)
		}
	}

	// 验证基础连接信息
	if mode, ok := configMap["mode"].(string); ok && mode == string(ModeSingle) {
		if host, ok := configMap["host"].(string); !ok || host == "" {
			return fmt.Errorf("单机模式需要指定host")
		}
		if port, ok := configMap["port"]; ok {
			if portInt, ok := port.(int); !ok || portInt <= 0 || portInt > 65535 {
				return fmt.Errorf("无效的端口号")
			}
		}
	}

	return nil
}

// GetDefaultConfig 获取默认的Redis缓存配置
// 用于提供配置模板或作为配置参考
func GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":                true,
		"mode":                   string(ModeSingle),
		"host":                   "localhost",
		"port":                   6379,
		"password":               "",
		"database":               0,
		"pool_size":              10,
		"min_idle_connections":   5,
		"max_retries":            3,
		"dial_timeout":           "5s",
		"read_timeout":           "3s",
		"write_timeout":          "3s",
		"pool_timeout":           "4s",
		"idle_timeout":           "300s",
		"cluster_addrs":          []string{},
		"sentinel_addrs":         []string{},
		"master_name":            "",
		"enable_tls":             false,
		"insecure_skip_verify":   false,
		"cert_file":              "",
		"key_file":               "",
		"ca_file":                "",
	}
} 