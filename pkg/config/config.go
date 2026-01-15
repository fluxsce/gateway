package config

import (
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gateway/pkg/utils/net"
	"gateway/pkg/utils/path"

	"github.com/spf13/viper"
)

var (
	// global 全局配置实例，只初始化一次
	global *Config
	// 确保global只初始化一次的标志
	initialized bool

	// cachedNodeId 缓存的节点ID
	cachedNodeId string
	// nodeIdOnce 确保节点ID只初始化一次
	nodeIdOnce sync.Once
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

// InitializeConfig 初始化配置（加载配置文件并设置全局时区）
// 这是一个高级初始化函数，整合了配置加载和全局时区设置
// 参数:
//   - configDir: 配置文件目录
//   - options: 加载选项，可选
//
// 返回:
//   - error: 可能的错误
func InitializeConfig(configDir string, options ...LoadOptions) error {
	// 加载配置文件
	if err := LoadConfig(configDir, options...); err != nil {
		return fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 设置全局时区
	if err := setupGlobalTimezone(); err != nil {
		// 时区设置失败不应该阻止应用启动，只记录警告
		log.Printf("设置全局时区失败: %v，使用默认时区", err)
	}

	return nil
}

// setupGlobalTimezone 设置全局时区
// 从配置中读取时区设置并应用到time.Local
// 返回:
//   - error: 可能的错误
func setupGlobalTimezone() error {
	// 获取时区配置
	timezone := GetString("app.local_timezone", "UTC")

	// 加载时区
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("加载时区 '%s' 失败: %w", timezone, err)
	}

	// 设置全局时区
	time.Local = location
	log.Printf("已设置全局时区为: %s", timezone)

	return nil
}

// GetGlobalTimezone 获取当前全局时区配置
// 返回:
//   - string: 时区名称
func GetGlobalTimezone() string {
	return GetString("app.local_timezone", "UTC")
}

// GetVersion 获取应用版本号
// 从配置中读取 app.version
// 返回:
//   - string: 版本号，默认为 "unknown"
func GetVersion() string {
	return GetString("app.version", "3.0.4")
}

// GetAppName 获取应用名称
// 从配置中读取 app.name
// 返回:
//   - string: 应用名称，默认为 "Gateway"
func GetAppName() string {
	return GetString("app.name", "Gateway")
}

// GetNodeId 获取全局节点ID
// 支持容器和跨平台运行，每次获取的结果都是一致的
//
// 优先级：
//  1. 配置文件 app.node_id（如果非空）
//  2. 环境变量 GATEWAY_NODE_ID 或 POD_NAME（K8s已有配置可直接使用）
//  3. 持久化文件 .node_id（位于配置目录）
//  4. 基于机器特征自动生成（hostname + MAC地址的SHA256哈希）
//
// 容器环境：
//   - Docker: 设置环境变量 GATEWAY_NODE_ID=node-001
//   - K8s: 已有 POD_NAME 环境变量时自动使用，无需额外配置
//
// 返回:
//   - string: 节点ID
func GetNodeId() string {
	nodeIdOnce.Do(func() {
		// 1. 优先从配置读取
		if nodeId := GetString("app.node_id", ""); nodeId != "" {
			cachedNodeId = nodeId
			return
		}

		// 2. 从环境变量读取（容器环境推荐）
		// 支持 GATEWAY_NODE_ID 或 POD_NAME（K8s常用）
		for _, envKey := range []string{"GATEWAY_NODE_ID", "POD_NAME"} {
			if nodeId := os.Getenv(envKey); nodeId != "" {
				cachedNodeId = nodeId
				return
			}
		}

		// 3. 尝试从持久化文件读取
		nodeIdFile := filepath.Join(GetConfigDir(), ".node_id")
		if nodeId, err := path.ReadFileContent(nodeIdFile); err == nil && len(nodeId) >= 8 {
			cachedNodeId = nodeId
			return
		}

		// 4. 基于机器特征自动生成
		hostname, _ := os.Hostname()
		if hostname == "" {
			hostname = "unknown"
		}
		macData := net.GetAllMACAddresses()
		if macData == "" {
			macData = net.GetFirstIPv4Address()
		}
		hash := sha256.Sum256([]byte(hostname + "|" + macData))
		cachedNodeId = fmt.Sprintf("%x", hash)

		// 5. 尝试持久化（失败不影响使用）
		if err := path.WriteFileContent(nodeIdFile, cachedNodeId); err != nil {
			log.Printf("警告: 无法持久化节点ID: %v", err)
		}
	})
	return cachedNodeId
}

// ResetNodeId 重置节点ID缓存（仅用于测试）
func ResetNodeId() {
	nodeIdOnce = sync.Once{}
	cachedNodeId = ""
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
