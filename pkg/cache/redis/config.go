package redis

import (
	"fmt"
	"time"
)

// RedisConfig Redis缓存配置结构体
// 包含Redis连接、连接池、超时、重试等所有配置信息
type RedisConfig struct {
	// === 连接配置 ===
	Host     string `yaml:"host" json:"host" mapstructure:"host"`         // Redis服务器地址，例如: localhost, 192.168.1.100
	Port     int    `yaml:"port" json:"port" mapstructure:"port"`         // Redis服务器端口，默认: 6379
	Password string `yaml:"password" json:"password" mapstructure:"password"` // Redis认证密码，如果Redis没有设置密码则留空
	DB       int    `yaml:"db" json:"db" mapstructure:"db"`               // Redis数据库编号，范围0-15，默认: 0

	// === 连接池配置 ===
	PoolSize       int   `yaml:"pool_size" json:"pool_size" mapstructure:"pool_size"`             // 连接池最大连接数，建议值: 100
	MinIdleConns   int   `yaml:"min_idle_conns" json:"min_idle_conns" mapstructure:"min_idle_conns"`   // 最小空闲连接数，保持一定数量的连接避免频繁创建，建议值: 10
	MaxIdleConns   int   `yaml:"max_idle_conns" json:"max_idle_conns" mapstructure:"max_idle_conns"`   // 最大空闲连接数，控制空闲连接上限，建议值: 100
	MaxActiveConns int   `yaml:"max_active_conns" json:"max_active_conns" mapstructure:"max_active_conns"` // 最大活跃连接数，控制同时工作的连接数，建议值: 100
	IdleTimeout    int64 `yaml:"idle_timeout" json:"idle_timeout" mapstructure:"idle_timeout"`       // 空闲连接超时时间（毫秒），超时后连接会被关闭，默认: 1800000（30分钟）

	// === 超时配置 ===
	DialTimeout  time.Duration `yaml:"dial_timeout" json:"dial_timeout" mapstructure:"dial_timeout"`   // 建立连接的超时时间，默认: 5s
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`   // 读取数据的超时时间，默认: 3s
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"` // 写入数据的超时时间，默认: 3s
	PoolTimeout  time.Duration `yaml:"pool_timeout" json:"pool_timeout" mapstructure:"pool_timeout"`   // 从连接池获取连接的超时时间，默认: 4s

	// === 重试配置 ===
	MaxRetries      int           `yaml:"max_retries" json:"max_retries" mapstructure:"max_retries"`           // 操作失败时的最大重试次数，默认: 3
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff" json:"min_retry_backoff" mapstructure:"min_retry_backoff"` // 重试间隔的最小值，默认: 8ms
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff" json:"max_retry_backoff" mapstructure:"max_retry_backoff"` // 重试间隔的最大值，默认: 512ms

	// === 其他配置 ===
	KeyPrefix         string `yaml:"key_prefix" json:"key_prefix" mapstructure:"key_prefix"`                 // 缓存键前缀，用于区分不同应用或环境，例如: "myapp:prod"
	EnablePipelining  bool   `yaml:"enable_pipelining" json:"enable_pipelining" mapstructure:"enable_pipelining"`   // 是否启用管道模式提高批量操作性能，默认: true
	EnableCompression bool   `yaml:"enable_compression" json:"enable_compression" mapstructure:"enable_compression"` // 是否启用数据压缩，适用于大数据量场景，默认: false
}

// GetType 实现CacheConfig接口，返回缓存类型
func (r *RedisConfig) GetType() string {
	return "redis"
}

// Validate 实现CacheConfig接口，验证配置的有效性
func (r *RedisConfig) Validate() error {
	if r.Host == "" {
		return fmt.Errorf("redis host is required")
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("redis port must be between 1 and 65535, got %d", r.Port)
	}

	if r.DB < 0 || r.DB > 15 {
		return fmt.Errorf("redis db must be between 0 and 15, got %d", r.DB)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *RedisConfig) SetDefaults() {
	// 连接配置默认值
	if r.Port == 0 {
		r.Port = 6379
	}

	// 连接池配置默认值
	if r.PoolSize == 0 {
		r.PoolSize = 100
	}
	if r.MinIdleConns == 0 {
		r.MinIdleConns = 10
	}
	if r.MaxIdleConns == 0 {
		r.MaxIdleConns = 100
	}
	if r.MaxActiveConns == 0 {
		r.MaxActiveConns = 100
	}
	if r.IdleTimeout == 0 {
		r.IdleTimeout = 1800000 // 30分钟
	}

	// 超时配置默认值
	if r.DialTimeout == 0 {
		r.DialTimeout = 5 * time.Second
	}
	if r.ReadTimeout == 0 {
		r.ReadTimeout = 3 * time.Second
	}
	if r.WriteTimeout == 0 {
		r.WriteTimeout = 3 * time.Second
	}
	if r.PoolTimeout == 0 {
		r.PoolTimeout = 4 * time.Second
	}

	// 重试配置默认值
	if r.MaxRetries == 0 {
		r.MaxRetries = 3
	}
	if r.MinRetryBackoff == 0 {
		r.MinRetryBackoff = 8 * time.Millisecond
	}
	if r.MaxRetryBackoff == 0 {
		r.MaxRetryBackoff = 512 * time.Millisecond
	}

	// 其他配置默认值
	if !r.EnablePipelining {
		r.EnablePipelining = true
	}
}

// GetAddress 获取Redis服务器地址
func (r *RedisConfig) GetAddress() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

// GetIdleTimeoutDuration 获取空闲超时持续时间
func (r *RedisConfig) GetIdleTimeoutDuration() time.Duration {
	return time.Duration(r.IdleTimeout) * time.Millisecond
}

// String 返回配置的字符串表示（隐藏敏感信息）
func (r *RedisConfig) String() string {
	password := r.Password
	if password != "" {
		password = "***"
	}
	return fmt.Sprintf("Redis{Host:%s, Port:%d, Password:%s, DB:%d, PoolSize:%d}",
		r.Host, r.Port, password, r.DB, r.PoolSize)
}
