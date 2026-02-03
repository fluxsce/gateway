package memory

import (
	"fmt"
	"time"
)

// MemoryConfig 内存缓存配置结构体
// 包含内存缓存的基础配置信息
type MemoryConfig struct {
	// === 基础配置 ===
	Enabled bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"` // 是否启用内存缓存

	// === 容量配置 ===
	MaxSize int64 `yaml:"max_size" json:"max_size" mapstructure:"max_size"` // 最大存储条目数，0表示无限制，默认: 10000

	// === 过期配置 ===
	DefaultExpiration time.Duration `yaml:"default_expiration" json:"default_expiration" mapstructure:"default_expiration"`    // 默认过期时间，0表示永不过期，默认: 1小时
	CleanupInterval   time.Duration `yaml:"cleanup_interval" json:"cleanup_interval" mapstructure:"cleanup_interval"`          // 清理过期数据的间隔，默认: 10分钟
	EnableLazyCleanup bool          `yaml:"enable_lazy_cleanup" json:"enable_lazy_cleanup" mapstructure:"enable_lazy_cleanup"` // 是否启用懒惰清理（访问时清理），默认: true

	// === 淘汰策略 ===
	EvictionPolicy EvictionPolicy `yaml:"eviction_policy" json:"eviction_policy" mapstructure:"eviction_policy"` // 淘汰策略，默认: ttl（过期时间淘汰）

	// === 键配置 ===
	KeyPrefix string `yaml:"key_prefix" json:"key_prefix" mapstructure:"key_prefix"` // 缓存键前缀，用于区分不同应用或环境

	// === 监控配置 ===
	EnableMetrics    bool   `yaml:"enable_metrics" json:"enable_metrics" mapstructure:"enable_metrics"`          // 是否启用基础指标收集，默认: false
	MetricsNamespace string `yaml:"metrics_namespace" json:"metrics_namespace" mapstructure:"metrics_namespace"` // 指标命名空间，默认: memory_cache
}

// EvictionPolicy 淘汰策略枚举
type EvictionPolicy string

const (
	// EvictionTTL 基于过期时间淘汰（已实现，默认策略）
	EvictionTTL EvictionPolicy = "ttl"
	// EvictionLRU 最近最少使用（预留，未实现）
	EvictionLRU EvictionPolicy = "lru"
	// EvictionRandom 随机淘汰（预留，未实现）
	EvictionRandom EvictionPolicy = "random"
	// EvictionFIFO 先进先出（预留，未实现）
	EvictionFIFO EvictionPolicy = "fifo"
)

// GetType 返回缓存类型标识
func (m *MemoryConfig) GetType() string {
	return "memory"
}

// Validate 验证配置的有效性
func (m *MemoryConfig) Validate() error {
	// 验证容量配置
	if m.MaxSize < 0 {
		return fmt.Errorf("最大存储条目数不能为负数，当前值: %d", m.MaxSize)
	}

	// 验证过期配置
	if m.DefaultExpiration < 0 {
		return fmt.Errorf("默认过期时间不能为负数，当前值: %v", m.DefaultExpiration)
	}

	if m.CleanupInterval <= 0 {
		return fmt.Errorf("清理间隔必须大于0，当前值: %v", m.CleanupInterval)
	}

	// 验证淘汰策略
	switch m.EvictionPolicy {
	case EvictionTTL:
		// 已实现的策略
	case EvictionLRU, EvictionRandom, EvictionFIFO:
		// 预留策略，暂未实现，但配置有效
	case "":
		// 空值，将在SetDefaults中设置默认值
	default:
		return fmt.Errorf("不支持的淘汰策略: %s，支持的策略: ttl(已实现), lru(预留), random(预留), fifo(预留)", m.EvictionPolicy)
	}

	return nil
}

// SetDefaults 设置默认值
func (m *MemoryConfig) SetDefaults() {
	// 基础配置默认值
	// Enabled字段不设置默认值，保持原始值

	// 容量配置默认值
	if m.MaxSize == 0 {
		m.MaxSize = 10000 // 默认最大10000个条目
	}

	// 过期配置默认值
	if m.DefaultExpiration == 0 {
		m.DefaultExpiration = time.Hour // 默认1小时过期
	}
	if m.CleanupInterval == 0 {
		m.CleanupInterval = 10 * time.Minute // 默认10分钟清理一次
	}
	if !m.EnableLazyCleanup {
		m.EnableLazyCleanup = true // 默认启用懒惰清理
	}

	// 淘汰策略默认值
	if m.EvictionPolicy == "" {
		m.EvictionPolicy = EvictionTTL // 默认使用过期时间淘汰策略
	}

	// 监控配置默认值
	if m.MetricsNamespace == "" {
		m.MetricsNamespace = "memory_cache"
	}
}

// GetMaxSize 获取最大存储条目数
func (m *MemoryConfig) GetMaxSize() int64 {
	return m.MaxSize
}

// GetDefaultExpiration 获取默认过期时间
func (m *MemoryConfig) GetDefaultExpiration() time.Duration {
	return m.DefaultExpiration
}

// GetCleanupInterval 获取清理间隔
func (m *MemoryConfig) GetCleanupInterval() time.Duration {
	return m.CleanupInterval
}

// IsEvictionEnabled 是否启用淘汰（基于容量限制）
func (m *MemoryConfig) IsEvictionEnabled() bool {
	return m.MaxSize > 0
}

// IsExpirationEnabled 是否启用过期淘汰
func (m *MemoryConfig) IsExpirationEnabled() bool {
	return m.DefaultExpiration > 0
}

// IsImplementedPolicy 检查策略是否已实现
func (m *MemoryConfig) IsImplementedPolicy() bool {
	return m.EvictionPolicy == EvictionTTL
}

// GetEvictionPolicy 获取淘汰策略
func (m *MemoryConfig) GetEvictionPolicy() EvictionPolicy {
	return m.EvictionPolicy
}

// IsMetricsEnabled 是否启用指标收集
func (m *MemoryConfig) IsMetricsEnabled() bool {
	return m.EnableMetrics
}

// String 返回配置的字符串表示
func (m *MemoryConfig) String() string {
	return fmt.Sprintf("MemoryConfig{Enabled: %v, MaxSize: %d, DefaultExpiration: %v, EvictionPolicy: %s, EnableMetrics: %v}",
		m.Enabled, m.MaxSize, m.DefaultExpiration, m.EvictionPolicy, m.EnableMetrics)
}

// GetDefaultConfig 获取默认的内存缓存配置。
//
// 返回值：
//   - map[string]interface{}: 默认配置映射，用于生成配置模板
func GetDefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"enabled":             true,
		"max_size":            10000,
		"key_prefix":          "",
		"eviction_policy":     string(EvictionTTL),
		"default_expiration":  "1h",
		"cleanup_interval":    "10m",
		"enable_lazy_cleanup": true,
		"enable_metrics":      false,
		"metrics_namespace":   "memory_cache",
	}
}
