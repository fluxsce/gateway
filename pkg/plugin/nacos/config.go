// Package nacos 提供了 Nacos 客户端的配置管理功能。
//
// 该包提供了一个完整且前端友好的配置接口，支持单机和集群部署。
// 配置分为基础配置和高级配置，满足不同场景的需求。
package nacos

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	// Host 服务器地址（必填）
	Host string `json:"host" yaml:"host" validate:"required" label:"服务器地址"`

	// Port 服务器端口（可选）
	// 默认: 8848
	Port int `json:"port,omitempty" yaml:"port,omitempty" validate:"min=1,max=65535" label:"端口" default:"8848"`

	// GrpcPort GRPC端口（可选）
	// 默认: HTTP端口+1000
	GrpcPort int `json:"grpcPort,omitempty" yaml:"grpcPort,omitempty" validate:"min=1,max=65535" label:"GRPC端口"`

	// ContextPath 上下文路径（可选）
	// 默认: "/nacos"
	ContextPath string `json:"contextPath,omitempty" yaml:"contextPath,omitempty" label:"上下文路径" default:"/nacos"`

	// Scheme 协议（可选）
	// 可选值: "http", "https"
	// 默认: "http"
	Scheme string `json:"scheme,omitempty" yaml:"scheme,omitempty" label:"协议" default:"http"`
}

// NacosConfig Nacos 完整配置结构，支持单机和集群部署。
//
// 该结构体提供了完整的配置选项，既适合前端使用，也支持高级配置。
// 配置分为服务器配置、认证配置、网络配置、缓存配置等模块。
//
// 使用方式：
//   - 单机部署: 设置 Servers 数组包含一个服务器
//   - 集群部署: 设置 Servers 数组包含多个服务器
//   - 简单配置: 只设置必要字段，其他使用默认值
//   - 高级配置: 根据需要调整各种参数
type NacosConfig struct {
	// === 服务器配置 ===

	// Servers 服务器列表（必填）
	// 单机部署时包含一个服务器，集群部署时包含多个服务器
	// 示例:
	//   单机: [{Host: "127.0.0.1", Port: 8848}]
	//   集群: [{Host: "nacos1.com", Port: 8848}, {Host: "nacos2.com", Port: 8848}]
	Servers []ServerConfig `json:"servers" yaml:"servers" validate:"required,min=1" label:"服务器列表"`

	// === 命名空间和分组 ===

	// Namespace 命名空间（可选）
	// 用于环境隔离，如: "dev", "test", "prod"
	// 默认: "public"
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty" label:"命名空间" placeholder:"public"`

	// Group 默认分组（可选）
	// 默认: "DEFAULT_GROUP"
	Group string `json:"group,omitempty" yaml:"group,omitempty" label:"默认分组" placeholder:"DEFAULT_GROUP"`

	// === 认证配置 ===

	// Username 用户名（可选）
	// 当 Nacos 启用认证时需要
	Username string `json:"username,omitempty" yaml:"username,omitempty" label:"用户名"`

	// Password 密码（可选）
	// 当 Nacos 启用认证时需要
	Password string `json:"password,omitempty" yaml:"password,omitempty" label:"密码" type:"password"`

	// AccessKey 访问密钥（可选）
	// 用于阿里云MSE等云服务
	AccessKey string `json:"accessKey,omitempty" yaml:"accessKey,omitempty" label:"访问密钥"`

	// SecretKey 密钥（可选）
	// 用于阿里云MSE等云服务
	SecretKey string `json:"secretKey,omitempty" yaml:"secretKey,omitempty" label:"密钥" type:"password"`

	// === 网络配置 ===

	// Timeout 超时时间（秒）（可选）
	// 默认: 5秒，范围: 3-30秒
	Timeout int `json:"timeout,omitempty" yaml:"timeout,omitempty" validate:"min=3,max=30" label:"超时时间(秒)" default:"5"`

	// BeatInterval 心跳间隔（秒）（可选）
	// 默认: 5秒，范围: 1-30秒
	BeatInterval int `json:"beatInterval,omitempty" yaml:"beatInterval,omitempty" validate:"min=1,max=30" label:"心跳间隔(秒)" default:"5"`

	// === 缓存配置 ===

	// CacheDir 本地缓存目录（可选）
	// 默认: "/tmp/nacos/cache"
	CacheDir string `json:"cacheDir,omitempty" yaml:"cacheDir,omitempty" label:"缓存目录" placeholder:"/tmp/nacos/cache"`

	// NotLoadCacheAtStart 启动时不加载缓存（可选）
	// 默认: true（提高启动速度）
	NotLoadCacheAtStart bool `json:"notLoadCacheAtStart,omitempty" yaml:"notLoadCacheAtStart,omitempty" label:"启动时不加载缓存" default:"true"`

	// DisableUseSnapShot 禁用快照缓存（可选）
	// 默认: false（启用快照提高可用性）
	DisableUseSnapShot bool `json:"disableUseSnapShot,omitempty" yaml:"disableUseSnapShot,omitempty" label:"禁用快照缓存"`

	// UpdateCacheWhenEmpty 空结果时更新缓存（可选）
	// 默认: false
	UpdateCacheWhenEmpty bool `json:"updateCacheWhenEmpty,omitempty" yaml:"updateCacheWhenEmpty,omitempty" label:"空结果时更新缓存"`

	// === 日志配置 ===

	// LogDir 日志目录（可选）
	// 默认: "/tmp/nacos/log"
	LogDir string `json:"logDir,omitempty" yaml:"logDir,omitempty" label:"日志目录" placeholder:"/tmp/nacos/log"`

	// LogLevel 日志级别（可选）
	// 可选值: "debug", "info", "warn", "error"
	// 默认: "info"
	LogLevel string `json:"logLevel,omitempty" yaml:"logLevel,omitempty" label:"日志级别" default:"info"`

	// AppendToStdout 输出到控制台（可选）
	// 默认: false
	AppendToStdout bool `json:"appendToStdout,omitempty" yaml:"appendToStdout,omitempty" label:"输出到控制台"`

	// === 性能配置 ===

	// UpdateThreadNum 更新线程数（可选）
	// 默认: 20，范围: 1-100
	UpdateThreadNum int `json:"updateThreadNum,omitempty" yaml:"updateThreadNum,omitempty" validate:"min=1,max=100" label:"更新线程数" default:"20"`

	// === TLS配置 ===

	// EnableTLS 启用TLS（可选）
	// 默认: false
	EnableTLS bool `json:"enableTLS,omitempty" yaml:"enableTLS,omitempty" label:"启用TLS"`

	// TrustAll 信任所有证书（可选）
	// 仅用于测试环境
	TrustAll bool `json:"trustAll,omitempty" yaml:"trustAll,omitempty" label:"信任所有证书"`

	// CaFile CA证书文件路径（可选）
	CaFile string `json:"caFile,omitempty" yaml:"caFile,omitempty" label:"CA证书文件"`

	// CertFile 客户端证书文件路径（可选）
	CertFile string `json:"certFile,omitempty" yaml:"certFile,omitempty" label:"客户端证书文件"`

	// KeyFile 客户端私钥文件路径（可选）
	KeyFile string `json:"keyFile,omitempty" yaml:"keyFile,omitempty" label:"客户端私钥文件"`

	// === 应用信息 ===

	// AppName 应用名称（可选）
	AppName string `json:"appName,omitempty" yaml:"appName,omitempty" label:"应用名称"`

	// AppKey 应用标识（可选）
	AppKey string `json:"appKey,omitempty" yaml:"appKey,omitempty" label:"应用标识"`

	// === 高级配置 ===

	// OpenKMS 启用KMS加密（可选）
	// 用于阿里云等云服务
	OpenKMS bool `json:"openKMS,omitempty" yaml:"openKMS,omitempty" label:"启用KMS加密"`

	// RegionId 地域ID（可选）
	// 用于阿里云等云服务
	RegionId string `json:"regionId,omitempty" yaml:"regionId,omitempty" label:"地域ID"`
}

// Config 内部使用的完整配置结构。
//
// 这是 Nacos SDK 所需的完整配置，由 NacosConfig 转换而来。
// 包含所有必要的服务器配置和客户端配置。
type Config struct {
	// ServerConfigs 服务端配置列表
	ServerConfigs []constant.ServerConfig `json:"serverConfigs"`

	// ClientConfig 客户端配置
	ClientConfig constant.ClientConfig `json:"clientConfig"`
}

// NewConfig 从用户配置创建完整的 Nacos 配置。
//
// 该方法将前端友好的 NacosConfig 转换为 Nacos SDK 所需的完整配置。
// 自动设置所有最佳实践的默认值，确保配置的可靠性和性能。
//
// 参数:
//   - config: 用户提供的配置
//
// 返回:
//   - *Config: 完整的 Nacos 配置
//   - error: 配置验证错误
//
// 支持的配置模式:
//   - 单机模式: 设置 ServerAddr 和 Port
//   - 集群模式: 设置 Servers 数组
//   - 混合模式: 同时设置会优先使用 Servers
//
// 示例:
//
//	// 单机配置
//	config := &NacosConfig{
//		ServerAddr: "127.0.0.1",
//		Port:       8848,
//		Namespace:  "dev",
//	}
//
//	// 集群配置
//	config := &NacosConfig{
//		Servers: []ServerConfig{
//			{Host: "nacos1.example.com", Port: 8848},
//			{Host: "nacos2.example.com", Port: 8848},
//			{Host: "nacos3.example.com", Port: 8848},
//		},
//		Namespace: "prod",
//		Username:  "admin",
//		Password:  "password",
//	}
func NewConfig(userConfig *NacosConfig) (*Config, error) {
	if userConfig == nil {
		return nil, fmt.Errorf("配置不能为空")
	}

	// 验证配置
	if err := Validate(userConfig); err != nil {
		return nil, err
	}

	// 构建服务器配置列表
	var serverConfigs []constant.ServerConfig

	for _, server := range userConfig.Servers {
		port := server.Port
		if port == 0 {
			port = 8848
		}

		grpcPort := server.GrpcPort
		if grpcPort == 0 {
			grpcPort = port + 1000
		}

		contextPath := server.ContextPath
		if contextPath == "" {
			contextPath = "/nacos"
		}

		scheme := server.Scheme
		if scheme == "" {
			scheme = "http"
		}

		serverConfigs = append(serverConfigs, constant.ServerConfig{
			IpAddr:      server.Host,
			Port:        uint64(port),
			GrpcPort:    uint64(grpcPort),
			ContextPath: contextPath,
			Scheme:      scheme,
		})
	}

	// 设置默认值
	namespace := userConfig.Namespace
	if namespace == "" {
		namespace = "public"
	}

	group := userConfig.Group
	if group == "" {
		group = "DEFAULT_GROUP"
	}

	timeout := userConfig.Timeout
	if timeout == 0 {
		timeout = 5
	}

	beatInterval := userConfig.BeatInterval
	if beatInterval == 0 {
		beatInterval = 5
	}

	updateThreadNum := userConfig.UpdateThreadNum
	if updateThreadNum == 0 {
		updateThreadNum = 20
	}

	cacheDir := userConfig.CacheDir
	if cacheDir == "" {
		cacheDir = "/tmp/nacos/cache"
	}

	logDir := userConfig.LogDir
	if logDir == "" {
		logDir = "/tmp/nacos/log"
	}

	logLevel := userConfig.LogLevel
	if logLevel == "" {
		logLevel = "info"
	}

	// 构建客户端配置
	clientConfig := constant.ClientConfig{
		NamespaceId:          namespace,
		TimeoutMs:            uint64(timeout * 1000),     // 转换为毫秒
		BeatInterval:         int64(beatInterval * 1000), // 转换为毫秒
		UpdateThreadNum:      updateThreadNum,
		CacheDir:             cacheDir,
		LogDir:               logDir,
		LogLevel:             logLevel,
		NotLoadCacheAtStart:  userConfig.NotLoadCacheAtStart,
		DisableUseSnapShot:   userConfig.DisableUseSnapShot,
		UpdateCacheWhenEmpty: userConfig.UpdateCacheWhenEmpty,
		AppendToStdout:       userConfig.AppendToStdout,
		AppName:              userConfig.AppName,
		AppKey:               userConfig.AppKey,
		OpenKMS:              userConfig.OpenKMS,
		RegionId:             userConfig.RegionId,
	}

	// 设置认证信息
	if userConfig.Username != "" && userConfig.Password != "" {
		clientConfig.Username = userConfig.Username
		clientConfig.Password = userConfig.Password
	}

	if userConfig.AccessKey != "" && userConfig.SecretKey != "" {
		clientConfig.AccessKey = userConfig.AccessKey
		clientConfig.SecretKey = userConfig.SecretKey
	}

	// 设置TLS配置
	if userConfig.EnableTLS {
		clientConfig.TLSCfg = constant.TLSConfig{
			Enable:   true,
			TrustAll: userConfig.TrustAll,
			CaFile:   userConfig.CaFile,
			CertFile: userConfig.CertFile,
			KeyFile:  userConfig.KeyFile,
		}
	}

	return &Config{
		ServerConfigs: serverConfigs,
		ClientConfig:  clientConfig,
	}, nil
}

// DefaultConfig 返回默认配置。
//
// 创建一个连接到本地 Nacos 服务器的默认配置，适合开发和测试环境。
//
// 返回:
//   - *NacosConfig: 默认的用户配置
//
// 示例:
//
//	config := DefaultConfig()
//	// 可以进一步修改
//	config.Namespace = "dev"
//	config.Username = "nacos"
//	config.Password = "nacos"
func DefaultConfig() *NacosConfig {
	return &NacosConfig{
		Servers: []ServerConfig{
			{Host: "127.0.0.1", Port: 8848},
		},
		Namespace:           "public",
		Group:               "DEFAULT_GROUP",
		Timeout:             5,
		BeatInterval:        5,
		UpdateThreadNum:     20,
		NotLoadCacheAtStart: true,
		LogLevel:            "info",
	}
}

// FromJSON 从 JSON 字符串创建配置。
//
// 该方法支持从 JSON 格式的配置字符串创建 NacosConfig，
// 适用于从配置文件或 API 加载配置。
//
// 参数:
//   - jsonStr: JSON 格式的配置字符串
//
// 返回:
//   - *NacosConfig: 解析后的配置
//   - error: JSON 解析错误
//
// 示例:
//
//	jsonStr := `{
//	  "serverAddr": "nacos.example.com",
//	  "port": 8848,
//	  "namespace": "prod",
//	  "username": "admin",
//	  "password": "password"
//	}`
//	config, err := FromJSON(jsonStr)
func FromJSON(jsonStr string) (*NacosConfig, error) {
	var config NacosConfig
	if err := json.Unmarshal([]byte(jsonStr), &config); err != nil {
		return nil, fmt.Errorf("解析 JSON 配置失败: %w", err)
	}
	return &config, nil
}

// ToJSON 将配置转换为 JSON 字符串。
//
// 该方法将 NacosConfig 序列化为 JSON 字符串，
// 便于配置的存储、传输和调试。
//
// 参数:
//   - config: 要序列化的配置
//
// 返回:
//   - string: JSON 格式的配置字符串
//   - error: 序列化错误
//
// 示例:
//
//	config := &NacosConfig{ServerAddr: "127.0.0.1", Namespace: "dev"}
//	jsonStr, err := ToJSON(config)
func ToJSON(config *NacosConfig) (string, error) {
	if config == nil {
		return "", fmt.Errorf("配置不能为空")
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化配置失败: %w", err)
	}
	return string(data), nil
}

// Validate 验证配置的有效性。
//
// 该方法对配置进行全面验证，包括必填字段检查、
// 参数范围验证和逻辑一致性检查。
//
// 参数:
//   - config: 要验证的配置
//
// 返回:
//   - error: 验证错误，nil 表示配置有效
//
// 验证规则:
//   - 必须设置 ServerAddr 或 Servers
//   - Port 范围: 1-65535
//   - Timeout 范围: 3-30 秒
//   - BeatInterval 范围: 1-30 秒
//   - UpdateThreadNum 范围: 1-100
//   - LogLevel: debug/info/warn/error
//   - TLS 配置一致性检查
//
// 示例:
//
//	config := &NacosConfig{ServerAddr: "127.0.0.1"}
//	if err := Validate(config); err != nil {
//		log.Fatalf("配置验证失败: %v", err)
//	}
func Validate(config *NacosConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证服务器配置
	if len(config.Servers) == 0 {
		return fmt.Errorf("servers 不能为空，至少需要一个服务器")
	}

	// 验证服务器列表
	for i, server := range config.Servers {
		if server.Host == "" {
			return fmt.Errorf("servers[%d].host 不能为空", i)
		}
		if server.Port != 0 && (server.Port < 1 || server.Port > 65535) {
			return fmt.Errorf("servers[%d].port 必须在 1-65535 范围内，当前值: %d", i, server.Port)
		}
		if server.GrpcPort != 0 && (server.GrpcPort < 1 || server.GrpcPort > 65535) {
			return fmt.Errorf("servers[%d].grpcPort 必须在 1-65535 范围内，当前值: %d", i, server.GrpcPort)
		}
		if server.Scheme != "" && server.Scheme != "http" && server.Scheme != "https" {
			return fmt.Errorf("servers[%d].scheme 必须是 http 或 https，当前值: %s", i, server.Scheme)
		}
	}

	// 验证超时时间
	if config.Timeout != 0 && (config.Timeout < 3 || config.Timeout > 30) {
		return fmt.Errorf("timeout 必须在 3-30 秒范围内，当前值: %d", config.Timeout)
	}

	// 验证心跳间隔
	if config.BeatInterval != 0 && (config.BeatInterval < 1 || config.BeatInterval > 30) {
		return fmt.Errorf("beatInterval 必须在 1-30 秒范围内，当前值: %d", config.BeatInterval)
	}

	// 验证线程数
	if config.UpdateThreadNum != 0 && (config.UpdateThreadNum < 1 || config.UpdateThreadNum > 100) {
		return fmt.Errorf("updateThreadNum 必须在 1-100 范围内，当前值: %d", config.UpdateThreadNum)
	}

	// 验证日志级别
	if config.LogLevel != "" {
		validLogLevels := map[string]bool{
			"debug": true, "info": true, "warn": true, "error": true,
		}
		if !validLogLevels[strings.ToLower(config.LogLevel)] {
			return fmt.Errorf("logLevel 必须是 debug、info、warn、error 之一，当前值: %s", config.LogLevel)
		}
	}

	// 验证认证配置一致性
	if (config.Username != "" && config.Password == "") ||
		(config.Username == "" && config.Password != "") {
		return fmt.Errorf("username 和 password 必须同时设置或同时为空")
	}

	if (config.AccessKey != "" && config.SecretKey == "") ||
		(config.AccessKey == "" && config.SecretKey != "") {
		return fmt.Errorf("accessKey 和 secretKey 必须同时设置或同时为空")
	}

	// 验证TLS配置
	if config.EnableTLS {
		if !config.TrustAll && config.CaFile == "" {
			return fmt.Errorf("启用 TLS 且不信任所有证书时，caFile 不能为空")
		}
		if (config.CertFile != "" && config.KeyFile == "") ||
			(config.CertFile == "" && config.KeyFile != "") {
			return fmt.Errorf("certFile 和 keyFile 必须同时设置或同时为空")
		}
	}

	return nil
}

// GetFormSchema 获取前端表单配置架构。
//
// 该方法返回用于前端动态生成配置表单的架构信息，
// 包括字段类型、验证规则、默认值等。
//
// 返回:
//   - map[string]interface{}: 表单架构
//
// 示例:
//
//	schema := GetFormSchema()
//	// 前端可以根据 schema 动态生成表单
func GetFormSchema() map[string]interface{} {
	return map[string]interface{}{
		"basic": map[string]interface{}{
			"title": "基础配置",
			"fields": map[string]interface{}{
				"servers": map[string]interface{}{
					"type":        "array",
					"label":       "服务器列表",
					"required":    true,
					"description": "Nacos服务器列表，单机时包含一个服务器，集群时包含多个服务器",
					"minItems":    1,
					"itemSchema": map[string]interface{}{
						"host": map[string]interface{}{
							"type":        "text",
							"label":       "主机地址",
							"required":    true,
							"placeholder": "127.0.0.1",
							"description": "服务器IP地址或域名",
						},
						"port": map[string]interface{}{
							"type":        "number",
							"label":       "端口",
							"required":    false,
							"default":     8848,
							"min":         1,
							"max":         65535,
							"description": "HTTP端口，默认8848",
						},
						"grpcPort": map[string]interface{}{
							"type":        "number",
							"label":       "GRPC端口",
							"required":    false,
							"description": "GRPC端口，默认为HTTP端口+1000",
							"min":         1,
							"max":         65535,
						},
						"scheme": map[string]interface{}{
							"type":        "select",
							"label":       "协议",
							"required":    false,
							"default":     "http",
							"options":     []string{"http", "https"},
							"description": "连接协议",
						},
						"contextPath": map[string]interface{}{
							"type":        "text",
							"label":       "上下文路径",
							"required":    false,
							"default":     "/nacos",
							"description": "服务器上下文路径",
						},
					},
				},
				"namespace": map[string]interface{}{
					"type":        "text",
					"label":       "命名空间",
					"required":    false,
					"placeholder": "public",
					"description": "用于环境隔离",
					"options": []string{
						"public", "dev", "test", "staging", "prod",
					},
				},
				"group": map[string]interface{}{
					"type":        "text",
					"label":       "默认分组",
					"required":    false,
					"placeholder": "DEFAULT_GROUP",
					"description": "服务注册时的默认分组",
				},
			},
		},
		"auth": map[string]interface{}{
			"title": "认证配置",
			"fields": map[string]interface{}{
				"username": map[string]interface{}{
					"type":        "text",
					"label":       "用户名",
					"required":    false,
					"description": "当 Nacos 启用认证时需要",
				},
				"password": map[string]interface{}{
					"type":        "password",
					"label":       "密码",
					"required":    false,
					"description": "当 Nacos 启用认证时需要",
				},
				"accessKey": map[string]interface{}{
					"type":        "text",
					"label":       "访问密钥",
					"required":    false,
					"description": "用于阿里云MSE等云服务",
				},
				"secretKey": map[string]interface{}{
					"type":        "password",
					"label":       "密钥",
					"required":    false,
					"description": "用于阿里云MSE等云服务",
				},
			},
		},
		"network": map[string]interface{}{
			"title": "网络配置",
			"fields": map[string]interface{}{
				"timeout": map[string]interface{}{
					"type":        "number",
					"label":       "超时时间(秒)",
					"required":    false,
					"default":     5,
					"min":         3,
					"max":         30,
					"description": "网络请求超时时间",
				},
				"beatInterval": map[string]interface{}{
					"type":        "number",
					"label":       "心跳间隔(秒)",
					"required":    false,
					"default":     5,
					"min":         1,
					"max":         30,
					"description": "服务实例心跳间隔",
				},
			},
		},
		"advanced": map[string]interface{}{
			"title": "高级配置",
			"fields": map[string]interface{}{
				"cacheDir": map[string]interface{}{
					"type":        "text",
					"label":       "缓存目录",
					"required":    false,
					"placeholder": "/tmp/nacos/cache",
					"description": "本地缓存存储目录",
				},
				"logDir": map[string]interface{}{
					"type":        "text",
					"label":       "日志目录",
					"required":    false,
					"placeholder": "/tmp/nacos/log",
					"description": "日志文件存储目录",
				},
				"logLevel": map[string]interface{}{
					"type":        "select",
					"label":       "日志级别",
					"required":    false,
					"default":     "info",
					"options":     []string{"debug", "info", "warn", "error"},
					"description": "日志输出级别",
				},
				"updateThreadNum": map[string]interface{}{
					"type":        "number",
					"label":       "更新线程数",
					"required":    false,
					"default":     20,
					"min":         1,
					"max":         100,
					"description": "服务更新线程池大小",
				},
				"enableTLS": map[string]interface{}{
					"type":        "boolean",
					"label":       "启用TLS",
					"required":    false,
					"default":     false,
					"description": "启用TLS加密连接",
				},
			},
		},
	}
}

// GetPresets 获取预设配置模板。
//
// 该方法返回常用的配置模板，方便用户快速选择。
//
// 返回:
//   - map[string]*NacosConfig: 预设配置映射
//
// 示例:
//
//	presets := GetPresets()
//	devConfig := presets["development"]
func GetPresets() map[string]*NacosConfig {
	return map[string]*NacosConfig{
		"local": {
			Servers: []ServerConfig{
				{Host: "127.0.0.1", Port: 8848},
			},
			Namespace: "public",
			Timeout:   5,
		},
		"development": {
			Servers: []ServerConfig{
				{Host: "nacos-dev.example.com", Port: 8848},
			},
			Namespace: "dev",
			Timeout:   5,
		},
		"testing": {
			Servers: []ServerConfig{
				{Host: "nacos-test.example.com", Port: 8848},
			},
			Namespace: "test",
			Username:  "nacos",
			Password:  "nacos",
			Timeout:   8,
		},
		"production": {
			Servers: []ServerConfig{
				{Host: "nacos.example.com", Port: 8848},
			},
			Namespace: "prod",
			Username:  "admin",
			Password:  "***请设置密码***",
			Timeout:   10,
		},
		"cluster": {
			Servers: []ServerConfig{
				{Host: "nacos1.example.com", Port: 8848},
				{Host: "nacos2.example.com", Port: 8848},
				{Host: "nacos3.example.com", Port: 8848},
			},
			Namespace: "prod",
			Username:  "admin",
			Password:  "***请设置密码***",
			Timeout:   10,
		},
		"aliyun_mse": {
			Servers: []ServerConfig{
				{Host: "mse-xxx-nacos-ans.mse.aliyuncs.com", Port: 8848},
			},
			Namespace: "prod",
			AccessKey: "***请设置AccessKey***",
			SecretKey: "***请设置SecretKey***",
			Timeout:   10,
		},
	}
}
