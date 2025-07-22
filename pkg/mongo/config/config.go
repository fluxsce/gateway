// Package config 提供MongoDB连接配置功能
//
// 此包定义了MongoDB连接的所有配置选项，包括：
// - 基本连接信息（主机、端口、数据库名）
// - 认证信息（用户名、密码、认证数据库）
// - 连接池配置（最大/最小连接数、超时设置）
// - 副本集配置（副本集名称、多主机地址）
// - SSL/TLS配置（证书文件、验证选项）
// - 读写偏好配置（读写关注级别、超时设置）
// - 其他高级配置（应用名称、重试选项等）
package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

// MongoConfig MongoDB连接配置结构体
// 包含连接MongoDB所需的所有配置参数
type MongoConfig struct {
	// === 基本标识信息 ===
	ID      string `yaml:"id" json:"id" mapstructure:"id"`           // 配置ID，用于标识和引用
	Enabled bool   `yaml:"enabled" json:"enabled" mapstructure:"enabled"` // 是否启用该配置
	
	// === 基本连接信息 ===
	Host     string `yaml:"host" json:"host" mapstructure:"host"`         // MongoDB主机地址，如 "localhost"
	Port     int    `yaml:"port" json:"port" mapstructure:"port"`         // MongoDB端口号，默认 27017
	Database string `yaml:"database" json:"database" mapstructure:"database"` // 要连接的数据库名称
	
	// === 认证信息 ===
	Username string `yaml:"username" json:"username" mapstructure:"username"` // 用户名，用于身份验证
	Password string `yaml:"password" json:"password" mapstructure:"password"` // 密码，用于身份验证
	AuthDB   string `yaml:"authdb" json:"authdb" mapstructure:"authdb"`       // 认证数据库名称，通常为 "admin"
	
	// === 连接池配置 ===
	MaxPoolSize      int           `yaml:"maxPoolSize" json:"maxPoolSize" mapstructure:"maxPoolSize"`           // 最大连接池大小，默认 100
	MinPoolSize      int           `yaml:"minPoolSize" json:"minPoolSize" mapstructure:"minPoolSize"`           // 最小连接池大小，默认 5
	MaxIdleTimeMS    time.Duration `yaml:"maxIdleTimeMS" json:"maxIdleTimeMS" mapstructure:"maxIdleTimeMS"`     // 连接最大空闲时间，默认 30分钟
	ConnectTimeoutMS time.Duration `yaml:"connectTimeoutMS" json:"connectTimeoutMS" mapstructure:"connectTimeoutMS"` // 连接超时时间，默认 10秒
	
	// === 副本集配置 ===
	ReplicaSet string   `yaml:"replicaSet" json:"replicaSet" mapstructure:"replicaSet"` // 副本集名称
	Hosts      []string `yaml:"hosts" json:"hosts" mapstructure:"hosts"`               // 副本集中的主机地址列表
	
	// === SSL/TLS 配置 ===
	EnableTLS     bool   `yaml:"enableTLS" json:"enableTLS" mapstructure:"enableTLS"`         // 是否启用TLS加密连接
	TLSCertFile   string `yaml:"tlsCertFile" json:"tlsCertFile" mapstructure:"tlsCertFile"`     // TLS客户端证书文件路径
	TLSKeyFile    string `yaml:"tlsKeyFile" json:"tlsKeyFile" mapstructure:"tlsKeyFile"`       // TLS客户端私钥文件路径
	TLSCAFile     string `yaml:"tlsCAFile" json:"tlsCAFile" mapstructure:"tlsCAFile"`         // TLS CA证书文件路径
	TLSSkipVerify bool   `yaml:"tlsSkipVerify" json:"tlsSkipVerify" mapstructure:"tlsSkipVerify"` // 是否跳过TLS证书验证（仅用于测试）
	
	// === 读写偏好配置 ===
	ReadPreference      string        `yaml:"readPreference" json:"readPreference" mapstructure:"readPreference"`           // 读偏好：primary, secondary, nearest等
	ReadConcern         string        `yaml:"readConcern" json:"readConcern" mapstructure:"readConcern"`                   // 读关注级别：local, majority, linearizable
	WriteConcern        string        `yaml:"writeConcern" json:"writeConcern" mapstructure:"writeConcern"`                 // 写关注级别：majority, 1, 0等
	WriteConcernTimeout time.Duration `yaml:"writeConcernTimeout" json:"writeConcernTimeout" mapstructure:"writeConcernTimeout"` // 写关注超时时间
	
	// === 其他高级配置 ===
	AppName                  string        `yaml:"appName" json:"appName" mapstructure:"appName"`                                     // 应用程序名称，用于日志记录
	RetryWrites              bool          `yaml:"retryWrites" json:"retryWrites" mapstructure:"retryWrites"`                         // 是否自动重试写操作
	RetryReads               bool          `yaml:"retryReads" json:"retryReads" mapstructure:"retryReads"`                           // 是否自动重试读操作
	ServerSelectionTimeoutMS time.Duration `yaml:"serverSelectionTimeoutMS" json:"serverSelectionTimeoutMS" mapstructure:"serverSelectionTimeoutMS"` // 服务器选择超时时间
	SocketTimeoutMS          time.Duration `yaml:"socketTimeoutMS" json:"socketTimeoutMS" mapstructure:"socketTimeoutMS"`           // Socket操作超时时间
	HeartbeatIntervalMS      time.Duration `yaml:"heartbeatIntervalMS" json:"heartbeatIntervalMS" mapstructure:"heartbeatIntervalMS"` // 心跳检查间隔时间
	
	// === 日志配置 ===
	EnableLogging bool   `yaml:"enableLogging" json:"enableLogging" mapstructure:"enableLogging"` // 是否启用MongoDB驱动日志
	LogLevel      string `yaml:"logLevel" json:"logLevel" mapstructure:"logLevel"`                 // 日志级别：debug, info, warn, error
}

// NewDefaultConfig 创建默认的MongoDB配置
// 返回一个包含合理默认值的配置实例
func NewDefaultConfig() *MongoConfig {
	return &MongoConfig{
		// 基本标识信息默认值
		ID:      "default",
		Enabled: true,
		
		// 基本连接信息默认值
		Host:     "localhost",
		Port:     27017,
		Database: "test",
		
		// 连接池默认配置
		MaxPoolSize:      100,                // 最大连接数
		MinPoolSize:      5,                  // 最小连接数
		MaxIdleTimeMS:    30 * time.Minute,   // 连接最大空闲时间
		ConnectTimeoutMS: 10 * time.Second,   // 连接超时时间
		
		// 读写偏好默认配置
		ReadPreference:      "primary",       // 优先读取主节点
		ReadConcern:         "local",         // 本地读关注级别
		WriteConcern:        "majority",      // 大多数写关注级别
		WriteConcernTimeout: 5 * time.Second, // 写关注超时时间
		
		// 其他默认配置
		RetryWrites:              true,                // 启用写重试
		RetryReads:               true,                // 启用读重试
		ServerSelectionTimeoutMS: 30 * time.Second,   // 服务器选择超时
		SocketTimeoutMS:          0,                   // Socket超时（0表示无限制）
		HeartbeatIntervalMS:      10 * time.Second,   // 心跳间隔
		
		// 日志默认配置
		EnableLogging: true,  // 启用日志
		LogLevel:      "info", // 默认日志级别
	}
}

// DefaultConfig 获取默认配置的别名函数
// 保持向后兼容性
func DefaultConfig() *MongoConfig {
	return NewDefaultConfig()
}

// 注意：不再需要URI构建方法，因为我们直接使用结构体参数进行连接
// 这样可以避免字符串拼接/解析的开销，提高性能和类型安全性

// Validate 验证配置的有效性
// 检查配置参数是否符合MongoDB连接要求
func (c *MongoConfig) Validate() error {
	// 检查配置ID
	if c.ID == "" {
		return fmt.Errorf("MongoDB配置错误：必须指定配置ID")
	}
	
	// 如果配置未启用，跳过其他验证
	if !c.Enabled {
		return nil
	}
	
	// 检查主机配置
	if c.Host == "" && len(c.Hosts) == 0 {
		return fmt.Errorf("MongoDB配置错误：必须指定主机地址(host)或主机列表(hosts)")
	}
	
	// 检查端口配置
	if c.Port <= 0 && c.Host != "" {
		return fmt.Errorf("MongoDB配置错误：端口号必须大于0，当前值: %d", c.Port)
	}
	
	// 检查数据库名称
	if c.Database == "" {
		return fmt.Errorf("MongoDB配置错误：必须指定数据库名称")
	}
	
	// 检查连接池配置
	if c.MaxPoolSize <= 0 {
		c.MaxPoolSize = 100 // 设置默认值
	}
	if c.MinPoolSize <= 0 {
		c.MinPoolSize = 5 // 设置默认值
	}
	if c.MaxPoolSize < c.MinPoolSize {
		return fmt.Errorf("MongoDB配置错误：最大连接数(%d)必须大于等于最小连接数(%d)", c.MaxPoolSize, c.MinPoolSize)
	}
	
	// 检查超时配置
	if c.ConnectTimeoutMS < 0 {
		return fmt.Errorf("MongoDB配置错误：连接超时时间不能为负数")
	}
	if c.ServerSelectionTimeoutMS < 0 {
		return fmt.Errorf("MongoDB配置错误：服务器选择超时时间不能为负数")
	}
	
	// 检查读写偏好
	validReadPreferences := []string{"primary", "primaryPreferred", "secondary", "secondaryPreferred", "nearest"}
	if c.ReadPreference != "" {
		valid := false
		for _, pref := range validReadPreferences {
			if c.ReadPreference == pref {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("MongoDB配置错误：无效的读偏好设置 '%s'，有效值为: %v", c.ReadPreference, validReadPreferences)
		}
	}
	
	// 检查读关注级别
	validReadConcerns := []string{"local", "available", "majority", "linearizable", "snapshot"}
	if c.ReadConcern != "" {
		valid := false
		for _, concern := range validReadConcerns {
			if c.ReadConcern == concern {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("MongoDB配置错误：无效的读关注级别 '%s'，有效值为: %v", c.ReadConcern, validReadConcerns)
		}
	}
	
	// 检查日志级别
	validLogLevels := []string{"debug", "info", "warn", "error"}
	if c.LogLevel != "" {
		valid := false
		for _, level := range validLogLevels {
			if c.LogLevel == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("MongoDB配置错误：无效的日志级别 '%s'，有效值为: %v", c.LogLevel, validLogLevels)
		}
	}
	
	return nil
}

// Clone 深度复制配置
// 创建当前配置的完整副本
func (c *MongoConfig) Clone() *MongoConfig {
	clone := *c
	// 复制切片
	if len(c.Hosts) > 0 {
		clone.Hosts = make([]string, len(c.Hosts))
		copy(clone.Hosts, c.Hosts)
	}
	return &clone
}

// String 返回配置的字符串表示
// 实现fmt.Stringer接口，用于日志记录和调试
func (c *MongoConfig) String() string {
	// 隐藏密码信息
	maskedPassword := ""
	if c.Password != "" {
		maskedPassword = "***"
	}
	
	return fmt.Sprintf("MongoConfig{ID: %s, Enabled: %t, Host: %s, Port: %d, Database: %s, Username: %s, Password: %s, MaxPoolSize: %d, MinPoolSize: %d}",
		c.ID, c.Enabled, c.Host, c.Port, c.Database, c.Username, maskedPassword, c.MaxPoolSize, c.MinPoolSize)
}

// IsReplicaSet 检查是否为副本集配置
func (c *MongoConfig) IsReplicaSet() bool {
	return c.ReplicaSet != "" || len(c.Hosts) > 1
}

// IsSSLEnabled 检查是否启用SSL
func (c *MongoConfig) IsSSLEnabled() bool {
	return c.EnableTLS
}

// HasAuthentication 检查是否配置了认证
func (c *MongoConfig) HasAuthentication() bool {
	return c.Username != "" && c.Password != ""
}

// ToClientOptions 将配置转换为MongoDB客户端选项
// 根据当前配置生成完整的MongoDB客户端连接选项
func (c *MongoConfig) ToClientOptions() (*options.ClientOptions, error) {
	// 创建客户端选项
	clientOptions := options.Client()
	
	// 设置主机信息
	if len(c.Hosts) > 0 {
		// 副本集模式
		clientOptions.SetHosts(c.Hosts)
		if c.ReplicaSet != "" {
			clientOptions.SetReplicaSet(c.ReplicaSet)
		}
	} else {
		// 单机模式
		host := c.Host
		if c.Port > 0 {
			host = fmt.Sprintf("%s:%d", c.Host, c.Port)
		}
		clientOptions.SetHosts([]string{host})
	}
	
	// 设置认证信息
	if c.Username != "" && c.Password != "" {
		credential := options.Credential{
			Username: c.Username,
			Password: c.Password,
		}
		if c.AuthDB != "" {
			credential.AuthSource = c.AuthDB
		}
		clientOptions.SetAuth(credential)
	}
	
	// 应用连接池配置
	if c.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(uint64(c.MaxPoolSize))
	}
	if c.MinPoolSize > 0 {
		clientOptions.SetMinPoolSize(uint64(c.MinPoolSize))
	}
	if c.MaxIdleTimeMS > 0 {
		clientOptions.SetMaxConnIdleTime(c.MaxIdleTimeMS)
	}
	
	// 应用超时配置
	if c.ConnectTimeoutMS > 0 {
		clientOptions.SetConnectTimeout(c.ConnectTimeoutMS)
	}
	if c.ServerSelectionTimeoutMS > 0 {
		clientOptions.SetServerSelectionTimeout(c.ServerSelectionTimeoutMS)
	}
	if c.SocketTimeoutMS > 0 {
		clientOptions.SetSocketTimeout(c.SocketTimeoutMS)
	}
	if c.HeartbeatIntervalMS > 0 {
		clientOptions.SetHeartbeatInterval(c.HeartbeatIntervalMS)
	}
	
	// 应用读写偏好配置
	if c.ReadPreference != "" {
		if readPref, err := parseReadPreference(c.ReadPreference); err == nil {
			clientOptions.SetReadPreference(readPref)
		}
	}
	if c.WriteConcern != "" {
		if writeConcern, err := parseWriteConcern(c.WriteConcern, c.WriteConcernTimeout); err == nil {
			clientOptions.SetWriteConcern(writeConcern)
		}
	}
	
	// 应用其他配置
	if c.AppName != "" {
		clientOptions.SetAppName(c.AppName)
	}
	
	// 应用重试配置
	clientOptions.SetRetryWrites(c.RetryWrites)
	clientOptions.SetRetryReads(c.RetryReads)
	
	// 应用TLS配置
	if c.EnableTLS {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: c.TLSSkipVerify,
		}
		if c.TLSCertFile != "" && c.TLSKeyFile != "" {
			cert, err := tls.LoadX509KeyPair(c.TLSCertFile, c.TLSKeyFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load TLS certificate: %w", err)
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
		if c.TLSCAFile != "" {
			caCert, err := os.ReadFile(c.TLSCAFile)
			if err != nil {
				return nil, fmt.Errorf("failed to read CA certificate: %w", err)
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			tlsConfig.RootCAs = caCertPool
		}
		clientOptions.SetTLSConfig(tlsConfig)
	}
	
	return clientOptions, nil
}

// parseReadPreference 解析读偏好字符串
func parseReadPreference(pref string) (*readpref.ReadPref, error) {
	switch pref {
	case "primary":
		return readpref.Primary(), nil
	case "primaryPreferred":
		return readpref.PrimaryPreferred(), nil
	case "secondary":
		return readpref.Secondary(), nil
	case "secondaryPreferred":
		return readpref.SecondaryPreferred(), nil
	case "nearest":
		return readpref.Nearest(), nil
	default:
		return nil, fmt.Errorf("invalid read preference: %s", pref)
	}
}

// parseWriteConcern 解析写关注字符串
func parseWriteConcern(concern string, timeout time.Duration) (*writeconcern.WriteConcern, error) {
	var opts []writeconcern.Option
	
	switch concern {
	case "majority":
		opts = append(opts, writeconcern.WMajority())
	case "1":
		opts = append(opts, writeconcern.W(1))
	case "0":
		opts = append(opts, writeconcern.W(0))
	default:
		return nil, fmt.Errorf("invalid write concern: %s", concern)
	}
	
	if timeout > 0 {
		opts = append(opts, writeconcern.WTimeout(timeout))
	}
	
	return writeconcern.New(opts...), nil
} 