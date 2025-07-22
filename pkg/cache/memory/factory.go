package memory

import (
	"fmt"
	"gateway/pkg/logger"
	"time"
)

// CreateFromConfig 从配置数据创建内存缓存实例
// 这是内存缓存模块对外提供的工厂方法，用于从通用配置创建内存缓存
// 参数:
//
//	name: 连接名称
//	configData: 配置数据，通常来自YAML配置文件
//
// 返回:
//
//	*MemoryCache: 内存缓存实例
//	error: 创建失败时返回错误信息
func CreateFromConfig(name string, configData interface{}) (*MemoryCache, error) {
	// 将interface{}转换为内存缓存配置
	memoryConfig := &MemoryConfig{}

	// 使用类型断言来转换配置
	if configMap, ok := configData.(map[string]interface{}); ok {
		if err := mapConfigToMemoryConfig(configMap, memoryConfig); err != nil {
			return nil, fmt.Errorf("映射配置到内存缓存配置失败: %w", err)
		}
	} else {
		return nil, fmt.Errorf("无效的配置数据类型，期望 map[string]interface{}")
	}

	// 检查是否启用，如果未启用则跳过
	if !memoryConfig.Enabled {
		logger.Warn("跳过未启用的内存缓存连接", "name", name)
		return nil, nil // 返回nil表示跳过此连接
	}

	// 设置默认值
	memoryConfig.SetDefaults()

	// 验证配置
	if err := memoryConfig.Validate(); err != nil {
		return nil, fmt.Errorf("验证内存缓存配置失败: %w", err)
	}

	// 记录连接创建信息
	logger.Debug("创建内存缓存连接",
		"name", name,
		"config", memoryConfig.String())

	// 创建内存缓存实例
	memoryCache, err := NewMemoryCache(memoryConfig)
	if err != nil {
		return nil, fmt.Errorf("创建内存缓存实例失败: %w", err)
	}

	return memoryCache, nil
}

// mapConfigToMemoryConfig 将通用配置映射到内存缓存配置
func mapConfigToMemoryConfig(configMap map[string]interface{}, memoryConfig *MemoryConfig) error {
	// 基础配置
	if enabled, ok := configMap["enabled"].(bool); ok {
		memoryConfig.Enabled = enabled
	}

	// 容量配置
	if maxSize, ok := configMap["max_size"]; ok {
		switch v := maxSize.(type) {
		case int:
			memoryConfig.MaxSize = int64(v)
		case int64:
			memoryConfig.MaxSize = v
		case float64: // YAML数字可能被解析为float64
			memoryConfig.MaxSize = int64(v)
		}
	}

	// 键配置
	if keyPrefix, ok := configMap["key_prefix"].(string); ok {
		memoryConfig.KeyPrefix = keyPrefix
	}

	// 淘汰策略
	if evictionPolicy, ok := configMap["eviction_policy"].(string); ok {
		memoryConfig.EvictionPolicy = EvictionPolicy(evictionPolicy)
	}

	// 清理配置
	if enableLazyCleanup, ok := configMap["enable_lazy_cleanup"].(bool); ok {
		memoryConfig.EnableLazyCleanup = enableLazyCleanup
	}

	// 监控配置
	if enableMetrics, ok := configMap["enable_metrics"].(bool); ok {
		memoryConfig.EnableMetrics = enableMetrics
	}

	if metricsNamespace, ok := configMap["metrics_namespace"].(string); ok {
		memoryConfig.MetricsNamespace = metricsNamespace
	}

	// 时间配置（字符串格式）
	if defaultExpStr, ok := configMap["default_expiration"].(string); ok {
		if duration, err := time.ParseDuration(defaultExpStr); err == nil {
			memoryConfig.DefaultExpiration = duration
		} else {
			logger.Warn("解析default_expiration失败，使用默认值", "value", defaultExpStr, "error", err)
		}
	}

	if cleanupIntStr, ok := configMap["cleanup_interval"].(string); ok {
		if duration, err := time.ParseDuration(cleanupIntStr); err == nil {
			memoryConfig.CleanupInterval = duration
		} else {
			logger.Warn("解析cleanup_interval失败，使用默认值", "value", cleanupIntStr, "error", err)
		}
	}

	return nil
}

// ValidateConfig 验证内存缓存配置的有效性
// 这是一个额外的验证函数，用于在创建实例前进行深度验证
func ValidateConfig(configData interface{}) error {
	configMap, ok := configData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("无效的配置数据类型")
	}

	// 检查必需字段
	if enabled, exists := configMap["enabled"]; !exists || enabled != true {
		return fmt.Errorf("内存缓存未启用或配置缺失")
	}

	// 验证淘汰策略
	if policy, ok := configMap["eviction_policy"].(string); ok {
		switch EvictionPolicy(policy) {
		case EvictionTTL:
			// 已实现的策略
		case EvictionLRU, EvictionRandom, EvictionFIFO:
			// 预留策略，配置有效但提示未实现
			logger.Warn("使用了预留策略，当前仅TTL策略已实现", "policy", policy)
		default:
			return fmt.Errorf("不支持的淘汰策略: %s", policy)
		}
	}

	// 验证容量配置
	if maxSize, ok := configMap["max_size"]; ok {
		var size int64
		switch v := maxSize.(type) {
		case int:
			size = int64(v)
		case int64:
			size = v
		case float64:
			size = int64(v)
		default:
			return fmt.Errorf("无效的max_size类型")
		}
		if size <= 0 {
			return fmt.Errorf("max_size必须大于0")
		}
	}

	// 验证时间配置
	timeFields := []string{"default_expiration", "cleanup_interval"}
	for _, field := range timeFields {
		if timeStr, ok := configMap[field].(string); ok && timeStr != "" {
			if _, err := time.ParseDuration(timeStr); err != nil {
				return fmt.Errorf("无效的%s格式: %s", field, timeStr)
			}
		}
	}

	return nil
}

// GetDefaultConfig 获取默认的内存缓存配置
// 用于提供配置模板或作为配置参考
func GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":             true,
		"max_size":            10000,
		"key_prefix":          "",
		"eviction_policy":     string(EvictionTTL), // 默认使用TTL策略
		"default_expiration":  "1h",
		"cleanup_interval":    "10m",
		"enable_lazy_cleanup": true,
		"enable_metrics":      false,
		"metrics_namespace":   "memory_cache",
	}
}
