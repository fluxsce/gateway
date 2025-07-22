package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

var (
	// global 全局配置实例，只初始化一次
	global *Config
	// 确保global只初始化一次的标志
	initialized bool
)

// Config 系统配置管理器
// 负责加载和管理全局配置
type Config struct {
	// viper 配置实例
	viper *viper.Viper
}

// LoadOptions 配置加载选项
type LoadOptions struct {
	// AllowOverride 是否允许覆盖已有配置
	AllowOverride bool
	// ClearExisting 是否清除已有配置
	ClearExisting bool
}

// DefaultLoadOptions 默认配置加载选项
func DefaultLoadOptions() LoadOptions {
	return LoadOptions{
		AllowOverride: true,
		ClearExisting: false,
	}
}

// 初始化全局配置实例
func init() {
	if !initialized {
		global = New() // 此时并未加载任何配置文件
		initialized = true
	}
}

// New 创建新的配置实例
// 返回: 配置实例
func New() *Config {
	return &Config{
		viper: viper.New(),
	}
}

// LoadConfig 加载指定目录下的所有配置文件
// 参数:
//   - configDir: 配置文件目录
//   - options: 加载选项，可选
//
// 返回:
//   - error: 可能的错误
func LoadConfig(configDir string, options ...LoadOptions) error {
	// 解析选项
	var opts LoadOptions
	if len(options) > 0 {
		opts = options[0]
	} else {
		opts = DefaultLoadOptions()
	}

	// 如果需要清除已有配置，则重新创建实例
	if opts.ClearExisting {
		global.viper = viper.New()
	} else if !opts.AllowOverride {
		// 检查是否已加载配置
		if global.viper.ConfigFileUsed() != "" {
			return fmt.Errorf("配置已加载，不允许覆盖")
		}
	}

	global.viper.SetConfigType("yaml")
	if configDir != "" {
		global.viper.AddConfigPath(configDir)
	}
	global.viper.AddConfigPath("./configs")
	global.viper.AddConfigPath(".")

	// 设置环境变量前缀
	global.viper.SetEnvPrefix("GATEWAY")
	global.viper.AutomaticEnv()

	// 读取配置文件
	global.viper.SetConfigName("app")
	if err := global.viper.ReadInConfig(); err != nil {
		// 允许app.yaml不存在
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return fmt.Errorf("读取app.yaml配置失败: %w", err)
		}
	}

	// 加载其他配置文件
	configs := []string{"logger", "database", "web"}
	for _, config := range configs {
		global.viper.SetConfigName(config)
		if err := global.viper.MergeInConfig(); err != nil {
			// 允许配置文件不存在
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("读取%s.yaml配置失败: %w", config, err)
			}
		}
	}

	return nil
}

// LoadConfigFile 加载指定的单个配置文件
// 参数:
//   - filePath: 配置文件路径
//   - options: 加载选项，可选
//
// 返回:
//   - error: 可能的错误
func LoadConfigFile(filePath string, options ...LoadOptions) error {
	// 解析选项
	var opts LoadOptions
	if len(options) > 0 {
		opts = options[0]
	} else {
		opts = DefaultLoadOptions()
	}

	// 如果需要清除已有配置，则重新创建实例
	if opts.ClearExisting {
		global.viper = viper.New()
	} else if !opts.AllowOverride {
		// 检查是否已加载配置
		if global.viper.ConfigFileUsed() != "" {
			return fmt.Errorf("配置已加载，不允许覆盖")
		}
	}

	// 获取文件扩展名
	ext := filepath.Ext(filePath)
	if ext == "" {
		return fmt.Errorf("配置文件必须有扩展名")
	}

	// 设置配置类型
	configType := ext[1:] // 去掉点号
	global.viper.SetConfigType(configType)

	// 读取配置文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer file.Close()

	if err := global.viper.MergeConfig(file); err != nil {
		return fmt.Errorf("合并配置文件失败: %w", err)
	}

	return nil
}

// Clear 清除全局配置
// 用于重置配置状态
func Clear() {
	global.viper = viper.New()
}

// Load 加载配置文件
// path: 配置文件路径
// 返回: 可能的错误
func (c *Config) Load(path string) error {
	// 获取配置文件目录
	dir := filepath.Dir(path)
	// 获取配置文件名（不含扩展名）
	name := filepath.Base(path)
	ext := filepath.Ext(path)
	if ext != "" {
		name = name[:len(name)-len(ext)]
	}

	// 设置配置文件信息
	c.viper.SetConfigName(name)
	c.viper.SetConfigType(ext[1:]) // 去掉点号
	c.viper.AddConfigPath(dir)
	c.viper.AddConfigPath(".")
	c.viper.AddConfigPath("./configs")

	// 设置环境变量前缀
	c.viper.SetEnvPrefix("GATEWAY")
	c.viper.AutomaticEnv()

	// 读取配置文件
	if err := c.viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	return nil
}

// LoadAndMerge 加载配置文件并合并到现有配置
// path: 配置文件路径
// 返回: 可能的错误
func (c *Config) LoadAndMerge(path string) error {
	// 创建临时配置实例
	tempCfg := New()
	if err := tempCfg.Load(path); err != nil {
		return err
	}

	// 合并配置
	if err := c.viper.MergeConfigMap(tempCfg.viper.AllSettings()); err != nil {
		return fmt.Errorf("合并配置失败: %w", err)
	}

	return nil
}

// IsExist 检查配置键是否存在
// 参数:
//   - key: 配置键
//
// 返回:
//   - bool: 键是否存在
func IsExist(key string) bool {
	if global == nil || global.viper == nil {
		return false
	}
	return global.viper.IsSet(key)
}

// Get 获取全局配置的值
// 参数:
//   - key: 配置键
//   - defaultValue: 默认值
//
// 返回:
//   - interface{}: 配置值
func Get(key string, defaultValue interface{}) interface{} {
	if !IsExist(key) {
		return defaultValue
	}
	return global.viper.Get(key)
}

// GetString 获取全局配置的字符串值
// 参数:
//   - key: 配置键
//   - defaultValue: 默认值
//
// 返回:
//   - string: 配置值
func GetString(key string, defaultValue string) string {
	if !IsExist(key) {
		return defaultValue
	}
	return global.viper.GetString(key)
}

// GetInt 获取全局配置的整数值
// 参数:
//   - key: 配置键
//   - defaultValue: 默认值
//
// 返回:
//   - int: 配置值
func GetInt(key string, defaultValue int) int {
	if !IsExist(key) {
		return defaultValue
	}
	return global.viper.GetInt(key)
}

// GetBool 获取全局配置的布尔值
// 参数:
//   - key: 配置键
//   - defaultValue: 默认值
//
// 返回:
//   - bool: 配置值
func GetBool(key string, defaultValue bool) bool {
	if !IsExist(key) {
		return defaultValue
	}
	return global.viper.GetBool(key)
}

// GetStringSlice 获取全局配置的字符串切片值
// 参数:
//   - key: 配置键
//   - defaultValue: 默认值
//
// 返回:
//   - []string: 配置值
func GetStringSlice(key string, defaultValue []string) []string {
	if !IsExist(key) {
		return defaultValue
	}
	return global.viper.GetStringSlice(key)
}

// GetSection 获取配置中的指定部分并解析到结构体
// 参数:
//   - key: 配置键，如"log"
//   - v: 目标结构体的指针
//
// 返回:
//   - error: 可能的错误
func GetSection(key string, v interface{}) error {
	if !IsExist(key) {
		return fmt.Errorf("配置项 %s 不存在", key)
	}
	return global.viper.UnmarshalKey(key, v)
}

// Get 获取配置值
// key: 配置键
// defaultValue: 默认值
// 返回: 配置值
func (c *Config) Get(key string, defaultValue interface{}) interface{} {
	if c == nil || c.viper == nil || !c.viper.IsSet(key) {
		return defaultValue
	}
	return c.viper.Get(key)
}

// GetString 获取字符串配置值
// key: 配置键
// defaultValue: 默认值
// 返回: 配置值
func (c *Config) GetString(key string, defaultValue string) string {
	if c == nil || c.viper == nil || !c.viper.IsSet(key) {
		return defaultValue
	}
	return c.viper.GetString(key)
}

// GetInt 获取整数配置值
// key: 配置键
// defaultValue: 默认值
// 返回: 配置值
func (c *Config) GetInt(key string, defaultValue int) int {
	if c == nil || c.viper == nil || !c.viper.IsSet(key) {
		return defaultValue
	}
	return c.viper.GetInt(key)
}

// GetBool 获取布尔配置值
// key: 配置键
// defaultValue: 默认值
// 返回: 配置值
func (c *Config) GetBool(key string, defaultValue bool) bool {
	if c == nil || c.viper == nil || !c.viper.IsSet(key) {
		return defaultValue
	}
	return c.viper.GetBool(key)
}

// GetStringSlice 获取字符串切片配置值
// key: 配置键
// defaultValue: 默认值
// 返回: 配置值
func (c *Config) GetStringSlice(key string, defaultValue []string) []string {
	if c == nil || c.viper == nil || !c.viper.IsSet(key) {
		return defaultValue
	}
	return c.viper.GetStringSlice(key)
}

// Unmarshal 将配置解析到结构体
// v: 目标结构体
// 返回: 可能的错误
func (c *Config) Unmarshal(v interface{}) error {
	if c == nil || c.viper == nil {
		return fmt.Errorf("配置实例未初始化")
	}
	return c.viper.Unmarshal(v)
}

// Set 设置配置值
// key: 配置键
// value: 配置值
func (c *Config) Set(key string, value interface{}) {
	if c == nil || c.viper == nil {
		return
	}
	c.viper.Set(key, value)
}

// Save 保存配置到文件
// path: 配置文件路径
// 返回: 可能的错误
func (c *Config) Save(path string) error {
	if c == nil || c.viper == nil {
		return fmt.Errorf("配置实例未初始化")
	}

	// 确保目录存在
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}

	// 保存配置
	if err := c.viper.WriteConfigAs(path); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// UnmarshalKey 将指定键的配置解析到结构体
// key: 配置键
// v: 目标结构体
// 返回: 可能的错误
func (c *Config) UnmarshalKey(key string, v interface{}) error {
	if c == nil || c.viper == nil {
		return fmt.Errorf("配置实例未初始化")
	}
	return c.viper.UnmarshalKey(key, v)
}
