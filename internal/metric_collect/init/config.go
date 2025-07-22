package init

import (
	"crypto/md5"
	"fmt"
	"os"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// MetricConfig 指标采集配置
// 定义了采集器的各种配置参数，包括采集间隔、存储配置、数据保留策略等
type MetricConfig struct {
	// 基础配置
	Enabled         bool          `yaml:"enabled" mapstructure:"enabled" json:"enabled"`                            // 是否启用指标采集
	CollectInterval time.Duration `yaml:"collect_interval" mapstructure:"collect_interval" json:"collect_interval"` // 采集间隔时间
	AutoStart       bool          `yaml:"auto_start" mapstructure:"auto_start" json:"auto_start"`                   // 是否自动启动
	TenantId        string        `yaml:"tenant_id" mapstructure:"tenant_id" json:"tenant_id"`                      // 租户ID
	ServerId        string        `yaml:"server_id" mapstructure:"server_id" json:"server_id"`                      // 服务器ID
	Operator        string        `yaml:"operator" mapstructure:"operator" json:"operator"`                         // 操作人

	// 采集器配置
	Collectors CollectorConfig `yaml:"collectors" mapstructure:"collectors" json:"collectors"`

	// 存储配置
	Storage StorageConfig `yaml:"storage" mapstructure:"storage" json:"storage"`
}

// CollectorConfig 采集器配置
// 定义了各种指标采集器的启用状态
type CollectorConfig struct {
	CPU         bool `yaml:"cpu" mapstructure:"cpu" json:"cpu"`                         // CPU指标采集
	Memory      bool `yaml:"memory" mapstructure:"memory" json:"memory"`                // 内存指标采集
	Disk        bool `yaml:"disk" mapstructure:"disk" json:"disk"`                      // 磁盘指标采集
	Network     bool `yaml:"network" mapstructure:"network" json:"network"`             // 网络指标采集
	Process     bool `yaml:"process" mapstructure:"process" json:"process"`             // 进程指标采集
	System      bool `yaml:"system" mapstructure:"system" json:"system"`                // 系统指标采集
	Temperature bool `yaml:"temperature" mapstructure:"temperature" json:"temperature"` // 温度指标采集
}

// StorageConfig 存储配置
// 定义了数据存储相关的配置参数
type StorageConfig struct {
	BatchSize     int           `yaml:"batch_size" mapstructure:"batch_size" json:"batch_size"`             // 批量写入大小
	FlushInterval time.Duration `yaml:"flush_interval" mapstructure:"flush_interval" json:"flush_interval"` // 数据刷新间隔
	MaxRetry      int           `yaml:"max_retry" mapstructure:"max_retry" json:"max_retry"`                // 最大重试次数

	// 数据保留策略
	Retention RetentionConfig `yaml:"retention" mapstructure:"retention" json:"retention"`
}

// RetentionConfig 数据保留配置
// 定义了历史数据的保留和清理策略
type RetentionConfig struct {
	Enabled         bool          `yaml:"enabled" mapstructure:"enabled" json:"enabled"`                            // 是否启用数据清理
	KeepDays        int           `yaml:"keep_days" mapstructure:"keep_days" json:"keep_days"`                      // 数据保留天数
	CleanupInterval time.Duration `yaml:"cleanup_interval" mapstructure:"cleanup_interval" json:"cleanup_interval"` // 清理执行间隔
}

// ConfigValidator 配置验证器
// 负责验证和处理指标采集配置
type ConfigValidator struct{}

// NewConfigValidator 创建配置验证器实例
func NewConfigValidator() *ConfigValidator {
	return &ConfigValidator{}
}

// LoadMetricConfig 加载指标采集配置
// 从全局配置中读取指标采集相关配置，如果配置不存在则使用默认配置
func (v *ConfigValidator) LoadMetricConfig() (*MetricConfig, error) {
	cfg := &MetricConfig{}

	// 从全局配置中读取指标采集配置
	if err := config.GetSection("app.metrics", cfg); err != nil {
		// 如果配置不存在，使用默认配置
		logger.Warn("未找到指标采集配置，使用默认配置", "error", err)
		cfg = v.getDefaultMetricConfig()
	}

	// 验证和处理配置
	if err := v.validateAndProcessConfig(cfg); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}

	return cfg, nil
}

// getDefaultMetricConfig 获取默认配置
// 返回一个包含合理默认值的配置实例
func (v *ConfigValidator) getDefaultMetricConfig() *MetricConfig {
	cfg := &MetricConfig{
		Enabled:         false,            // 默认不启用，需要显式配置
		CollectInterval: 30 * time.Second, // 默认30秒采集一次
		AutoStart:       false,            // 默认不自动启动
		TenantId:        "default",        // 默认租户
		ServerId:        "",               // 服务器ID将自动生成
		Operator:        "system",         // 默认操作人
	}

	// 默认启用所有采集器（除了温度）
	cfg.Collectors = CollectorConfig{
		CPU:         true,
		Memory:      true,
		Disk:        true,
		Network:     true,
		Process:     true,
		System:      true,
		Temperature: false, // 温度采集可能不是所有系统都支持
	}

	// 默认存储配置
	cfg.Storage = StorageConfig{
		BatchSize:     100,              // 批量写入100条记录
		FlushInterval: 60 * time.Second, // 每60秒刷新一次
		MaxRetry:      3,                // 最大重试3次
		Retention: RetentionConfig{
			Enabled:         true,           // 启用数据清理
			KeepDays:        30,             // 保留30天数据
			CleanupInterval: 24 * time.Hour, // 每24小时清理一次
		},
	}

	return cfg
}

// validateAndProcessConfig 验证和处理配置
// 对配置进行合法性检查，并补充缺失的配置项
func (v *ConfigValidator) validateAndProcessConfig(cfg *MetricConfig) error {
	// 验证基础配置
	if err := v.validateBasicConfig(cfg); err != nil {
		return fmt.Errorf("基础配置验证失败: %w", err)
	}

	// 验证采集器配置
	if err := v.validateCollectorConfig(&cfg.Collectors); err != nil {
		return fmt.Errorf("采集器配置验证失败: %w", err)
	}

	// 验证存储配置
	if err := v.validateStorageConfig(&cfg.Storage); err != nil {
		return fmt.Errorf("存储配置验证失败: %w", err)
	}

	return nil
}

// validateBasicConfig 验证基础配置
func (v *ConfigValidator) validateBasicConfig(cfg *MetricConfig) error {
	// 验证采集间隔
	if cfg.CollectInterval < time.Second {
		return fmt.Errorf("采集间隔不能小于1秒，当前值: %v", cfg.CollectInterval)
	}
	if cfg.CollectInterval > 24*time.Hour {
		return fmt.Errorf("采集间隔不能大于24小时，当前值: %v", cfg.CollectInterval)
	}

	// 验证租户ID
	if cfg.TenantId == "" {
		cfg.TenantId = "default"
		logger.Info("租户ID为空，使用默认值: default")
	}

	// 生成服务器ID（如果未配置）
	if cfg.ServerId == "" {
		cfg.ServerId = v.generateServerId()
		logger.Info("服务器ID未配置，自动生成", "server_id", cfg.ServerId)
	}

	// 验证操作人
	if cfg.Operator == "" {
		cfg.Operator = "system"
		logger.Info("操作人未配置，使用默认值: system")
	}

	return nil
}

// validateCollectorConfig 验证采集器配置
func (v *ConfigValidator) validateCollectorConfig(cfg *CollectorConfig) error {
	// 检查是否至少启用了一个采集器
	if !cfg.CPU && !cfg.Memory && !cfg.Disk && !cfg.Network && !cfg.Process && !cfg.System && !cfg.Temperature {
		logger.Warn("未启用任何采集器，指标采集将无法正常工作")
	}

	// 记录启用的采集器
	var enabledCollectors []string
	if cfg.CPU {
		enabledCollectors = append(enabledCollectors, "CPU")
	}
	if cfg.Memory {
		enabledCollectors = append(enabledCollectors, "Memory")
	}
	if cfg.Disk {
		enabledCollectors = append(enabledCollectors, "Disk")
	}
	if cfg.Network {
		enabledCollectors = append(enabledCollectors, "Network")
	}
	if cfg.Process {
		enabledCollectors = append(enabledCollectors, "Process")
	}
	if cfg.System {
		enabledCollectors = append(enabledCollectors, "System")
	}
	if cfg.Temperature {
		enabledCollectors = append(enabledCollectors, "Temperature")
	}

	if len(enabledCollectors) > 0 {
		logger.Info("已启用的采集器", "collectors", enabledCollectors)
	}

	return nil
}

// validateStorageConfig 验证存储配置
func (v *ConfigValidator) validateStorageConfig(cfg *StorageConfig) error {
	// 验证批量大小
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 100
		logger.Info("批量大小无效，使用默认值: 100")
	}
	if cfg.BatchSize > 10000 {
		return fmt.Errorf("批量大小不能超过10000，当前值: %d", cfg.BatchSize)
	}

	// 验证刷新间隔
	if cfg.FlushInterval < time.Second {
		cfg.FlushInterval = 60 * time.Second
		logger.Info("刷新间隔无效，使用默认值: 60秒")
	}
	if cfg.FlushInterval > 24*time.Hour {
		return fmt.Errorf("刷新间隔不能超过24小时，当前值: %v", cfg.FlushInterval)
	}

	// 验证最大重试次数
	if cfg.MaxRetry <= 0 {
		cfg.MaxRetry = 3
		logger.Info("最大重试次数无效，使用默认值: 3")
	}
	if cfg.MaxRetry > 10 {
		return fmt.Errorf("最大重试次数不能超过10，当前值: %d", cfg.MaxRetry)
	}

	// 验证数据保留配置
	if err := v.validateRetentionConfig(&cfg.Retention); err != nil {
		return fmt.Errorf("数据保留配置验证失败: %w", err)
	}

	return nil
}

// validateRetentionConfig 验证数据保留配置
func (v *ConfigValidator) validateRetentionConfig(cfg *RetentionConfig) error {
	if !cfg.Enabled {
		logger.Info("数据保留策略已禁用，历史数据将不会自动清理")
		return nil
	}

	// 验证保留天数
	if cfg.KeepDays <= 0 {
		cfg.KeepDays = 30
		logger.Info("数据保留天数无效，使用默认值: 30天")
	}
	if cfg.KeepDays > 3650 { // 不超过10年
		return fmt.Errorf("数据保留天数不能超过3650天，当前值: %d", cfg.KeepDays)
	}

	// 验证清理间隔
	if cfg.CleanupInterval < time.Hour {
		cfg.CleanupInterval = 24 * time.Hour
		logger.Info("清理间隔无效，使用默认值: 24小时")
	}
	if cfg.CleanupInterval > 30*24*time.Hour { // 不超过30天
		return fmt.Errorf("清理间隔不能超过30天，当前值: %v", cfg.CleanupInterval)
	}

	logger.Info("数据保留策略配置",
		"keep_days", cfg.KeepDays,
		"cleanup_interval", cfg.CleanupInterval)

	return nil
}

// generateServerId 生成服务器ID
// 使用主机名+随机字符串+时间戳生成唯一的服务器标识
func (v *ConfigValidator) generateServerId() string {
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "unknown"
	}

	// 使用主机名+随机字符串生成唯一ID
	randomStr := random.GenerateRandomString(8)
	data := fmt.Sprintf("%s-%s-%d", hostname, randomStr, time.Now().Unix())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("server_%x", hash)[:16]
}

// GetEnabledCollectorNames 获取已启用的采集器名称列表
// 返回当前配置中启用的采集器名称，用于日志记录和状态展示
func (cfg *MetricConfig) GetEnabledCollectorNames() []string {
	var enabled []string
	if cfg.Collectors.CPU {
		enabled = append(enabled, "cpu")
	}
	if cfg.Collectors.Memory {
		enabled = append(enabled, "memory")
	}
	if cfg.Collectors.Disk {
		enabled = append(enabled, "disk")
	}
	if cfg.Collectors.Network {
		enabled = append(enabled, "network")
	}
	if cfg.Collectors.Process {
		enabled = append(enabled, "process")
	}
	if cfg.Collectors.System {
		enabled = append(enabled, "system")
	}
	if cfg.Collectors.Temperature {
		enabled = append(enabled, "temperature")
	}
	return enabled
}

// IsValid 检查配置是否有效
// 快速检查配置的基本有效性
func (cfg *MetricConfig) IsValid() bool {
	return cfg.CollectInterval >= time.Second &&
		cfg.TenantId != "" &&
		cfg.ServerId != "" &&
		cfg.Operator != "" &&
		cfg.Storage.BatchSize > 0 &&
		cfg.Storage.FlushInterval >= time.Second &&
		cfg.Storage.MaxRetry > 0
}

// String 返回配置的字符串表示
// 用于日志记录和调试
func (cfg *MetricConfig) String() string {
	return fmt.Sprintf("MetricConfig{Enabled: %t, CollectInterval: %v, TenantId: %s, ServerId: %s, EnabledCollectors: %v}",
		cfg.Enabled, cfg.CollectInterval, cfg.TenantId, cfg.ServerId, cfg.GetEnabledCollectorNames())
}
