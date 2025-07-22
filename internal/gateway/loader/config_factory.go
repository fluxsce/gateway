package loader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"gateway/internal/gateway/config"
)

// ConfigSource 配置源类型
type ConfigSource string

const (
	ConfigSourceYAML ConfigSource = "yaml"
	ConfigSourceJSON ConfigSource = "json"
	ConfigSourceDB   ConfigSource = "database"
)

// GatewayConfigFactory 网关配置工厂
// 负责从不同配置源加载配置，支持viper和导出功能
type GatewayConfigFactory struct {
	source ConfigSource
	viper  *viper.Viper
}

// NewGatewayConfigFactory 创建网关配置工厂
func NewGatewayConfigFactory(source ConfigSource) *GatewayConfigFactory {
	v := viper.New()
	// 设置默认值
	setDefaultValues(v)

	return &GatewayConfigFactory{
		source: source,
		viper:  v,
	}
}

// LoadConfig 根据配置源加载配置（使用viper）
func (f *GatewayConfigFactory) LoadConfig(configPath string) (*config.GatewayConfig, error) {
	switch f.source {
	case ConfigSourceYAML:
		return f.loadConfigWithViper(configPath, "yaml")
	case ConfigSourceJSON:
		return f.loadConfigWithViper(configPath, "json")
	case ConfigSourceDB:
		return f.loadDatabaseConfig(configPath)
	default:
		return nil, fmt.Errorf("不支持的配置源: %s", f.source)
	}
}

// loadConfigWithViper 使用viper加载配置
func (f *GatewayConfigFactory) loadConfigWithViper(configPath, configType string) (*config.GatewayConfig, error) {
	// 如果配置文件不存在，返回默认配置
	if configPath == "" || !f.fileExists(configPath) {
		return f.createDefaultConfig(), nil
	}

	// 验证配置文件
	if err := f.ValidateConfigFile(configPath); err != nil {
		return nil, err
	}

	// 设置viper配置
	f.viper.SetConfigFile(configPath)
	f.viper.SetConfigType(configType)

	// 读取配置文件
	if err := f.viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	// 解析配置到结构体
	cfg := &config.GatewayConfig{}
	if err := f.viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 合并默认配置
	f.mergeDefaultConfig(cfg)

	return cfg, nil
}

// LoadConfigFromBytes 从字节数组加载配置
func (f *GatewayConfigFactory) LoadConfigFromBytes(data []byte, configType string) (*config.GatewayConfig, error) {
	if err := f.viper.ReadConfig(strings.NewReader(string(data))); err != nil {
		return nil, fmt.Errorf("读取配置数据失败: %w", err)
	}

	cfg := &config.GatewayConfig{}
	if err := f.viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("解析配置数据失败: %w", err)
	}

	f.mergeDefaultConfig(cfg)
	return cfg, nil
}

// ExportConfigToYAML 导出配置为YAML格式
func (f *GatewayConfigFactory) ExportConfigToYAML(cfg *config.GatewayConfig) ([]byte, error) {
	if cfg == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("序列化配置为YAML失败: %w", err)
	}

	return data, nil
}

// ExportConfigToYAMLFile 将配置导出到YAML文件
func (f *GatewayConfigFactory) ExportConfigToYAMLFile(cfg *config.GatewayConfig, filePath string) error {
	if cfg == nil {
		return fmt.Errorf("配置不能为空")
	}

	if filePath == "" {
		return fmt.Errorf("文件路径不能为空")
	}

	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	// 导出为YAML
	data, err := f.ExportConfigToYAML(cfg)
	if err != nil {
		return err
	}

	// 写入文件
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// ExportDefaultConfigToYAML 导出默认配置为YAML文件
func (f *GatewayConfigFactory) ExportDefaultConfigToYAML(filePath string) error {
	defaultCfg := f.createDefaultConfig()
	return f.ExportConfigToYAMLFile(defaultCfg, filePath)
}

// GetConfigAsMap 获取配置的Map表示（用于调试和检查）
func (f *GatewayConfigFactory) GetConfigAsMap() map[string]interface{} {
	return f.viper.AllSettings()
}

// SetConfigValue 设置配置值（用于动态配置）
func (f *GatewayConfigFactory) SetConfigValue(key string, value interface{}) {
	f.viper.Set(key, value)
}

// GetConfigValue 获取配置值
func (f *GatewayConfigFactory) GetConfigValue(key string) interface{} {
	return f.viper.Get(key)
}

// WatchConfig 监听配置文件变化（用于热重载）
func (f *GatewayConfigFactory) WatchConfig(callback func(*config.GatewayConfig)) error {
	f.viper.WatchConfig()
	f.viper.OnConfigChange(func(e fsnotify.Event) {
		cfg := &config.GatewayConfig{}
		if err := f.viper.Unmarshal(cfg); err != nil {
			return
		}
		f.mergeDefaultConfig(cfg)
		if callback != nil {
			callback(cfg)
		}
	})
	return nil
}

// loadDatabaseConfig 从数据库加载配置（预留）
func (f *GatewayConfigFactory) loadDatabaseConfig(configID string) (*config.GatewayConfig, error) {
	// TODO: 实现数据库配置加载逻辑
	// 这里可以连接数据库，根据 configID 查询配置

	// 目前返回默认配置
	return f.createDefaultConfig(), nil
}

// createDefaultConfig 创建默认配置的副本
func (f *GatewayConfigFactory) createDefaultConfig() *config.GatewayConfig {
	defaultCfg := config.DefaultGatewayConfig
	return &defaultCfg
}

// setDefaultValues 设置viper的默认值
func setDefaultValues(v *viper.Viper) {
	// 基础配置默认值
	v.SetDefault("base.listen", ":8080")
	v.SetDefault("base.name", "Gateway Gateway")
	v.SetDefault("base.read_timeout", "30s")
	v.SetDefault("base.write_timeout", "30s")
	v.SetDefault("base.idle_timeout", "120s")
	v.SetDefault("base.max_body_size", 10*1024*1024) // 10MB
	v.SetDefault("base.enable_https", false)
	v.SetDefault("base.use_gin", true)
	v.SetDefault("base.enable_access_log", true)
	v.SetDefault("base.log_format", "json")
	v.SetDefault("base.log_level", "info")
	v.SetDefault("base.enable_gzip", true)

	// 认证配置默认值
	v.SetDefault("auth.enabled", false)
	v.SetDefault("auth.strategy", "no_auth")

	// CORS配置默认值
	v.SetDefault("cors.enabled", true)
	v.SetDefault("cors.allow_origins", []string{"*"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"})
	v.SetDefault("cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	v.SetDefault("cors.max_age", 86400)

	// 限流配置默认值
	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.algorithm", "token_bucket")
	v.SetDefault("rate_limit.rate", 100)
	v.SetDefault("rate_limit.burst", 50)
	v.SetDefault("rate_limit.error_status_code", 429)
	v.SetDefault("rate_limit.error_message", "Rate limit exceeded")
}

// mergeDefaultConfig 合并默认配置
func (f *GatewayConfigFactory) mergeDefaultConfig(cfg *config.GatewayConfig) {
	defaultCfg := f.createDefaultConfig()

	if cfg.Base.Listen == "" {
		cfg.Base.Listen = defaultCfg.Base.Listen
	}
	if cfg.Base.Name == "" {
		cfg.Base.Name = defaultCfg.Base.Name
	}
	if cfg.Base.ReadTimeout == 0 {
		cfg.Base.ReadTimeout = defaultCfg.Base.ReadTimeout
	}
	if cfg.Base.WriteTimeout == 0 {
		cfg.Base.WriteTimeout = defaultCfg.Base.WriteTimeout
	}
	if cfg.Base.IdleTimeout == 0 {
		cfg.Base.IdleTimeout = defaultCfg.Base.IdleTimeout
	}
	if cfg.Base.MaxBodySize == 0 {
		cfg.Base.MaxBodySize = defaultCfg.Base.MaxBodySize
	}
	if cfg.Base.LogFormat == "" {
		cfg.Base.LogFormat = defaultCfg.Base.LogFormat
	}
	if cfg.Base.LogLevel == "" {
		cfg.Base.LogLevel = defaultCfg.Base.LogLevel
	}

	// 合并各模块默认配置
	if cfg.Router.ID == "" {
		cfg.Router = defaultCfg.Router
	}
	if cfg.Proxy.ID == "" {
		cfg.Proxy = defaultCfg.Proxy
	}
	if cfg.Security.ID == "" {
		cfg.Security = defaultCfg.Security
	}
	if cfg.Auth.Strategy == "" {
		cfg.Auth = defaultCfg.Auth
	}
	if len(cfg.CORS.AllowOrigins) == 0 {
		cfg.CORS = defaultCfg.CORS
	}
	if cfg.RateLimit.Rate == 0 {
		cfg.RateLimit = defaultCfg.RateLimit
	}
}

// fileExists 检查文件是否存在
func (f *GatewayConfigFactory) fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// GetSupportedConfigSources 获取支持的配置源
func GetSupportedConfigSources() []ConfigSource {
	return []ConfigSource{
		ConfigSourceYAML,
		ConfigSourceJSON,
		ConfigSourceDB,
	}
}

// GetConfigSourceDescription 获取配置源描述
func GetConfigSourceDescription(source ConfigSource) string {
	descriptions := map[ConfigSource]string{
		ConfigSourceYAML: "YAML配置文件，支持复杂的层级结构和注释",
		ConfigSourceJSON: "JSON配置文件，标准的JSON格式配置",
		ConfigSourceDB:   "数据库配置，从关系型数据库动态加载配置",
	}

	if desc, exists := descriptions[source]; exists {
		return desc
	}
	return "未知配置源"
}

// ValidateConfigFile 验证配置文件格式
func (f *GatewayConfigFactory) ValidateConfigFile(configPath string) error {
	if configPath == "" {
		return fmt.Errorf("配置文件路径不能为空")
	}

	if !f.fileExists(configPath) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	ext := strings.ToLower(filepath.Ext(configPath))

	switch f.source {
	case ConfigSourceYAML:
		if ext != ".yaml" && ext != ".yml" {
			return fmt.Errorf("YAML配置源需要.yaml或.yml文件，当前文件: %s", configPath)
		}
	case ConfigSourceJSON:
		if ext != ".json" {
			return fmt.Errorf("JSON配置源需要.json文件，当前文件: %s", configPath)
		}
	}

	return nil
}
