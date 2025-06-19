package redis

import (
	"fmt"
	"strings"
	"time"
)

// ConnectionMode Redis连接模式
type ConnectionMode string

const (
	// ModeSingle 单机模式
	ModeSingle ConnectionMode = "single"
	// ModeSentinel 哨兵模式
	ModeSentinel ConnectionMode = "sentinel"
	// ModeCluster 集群模式
	ModeCluster ConnectionMode = "cluster"
)

// RedisConfig Redis缓存配置结构体
// 包含Redis连接、连接池、超时、重试等所有配置信息，支持单机、哨兵、集群模式
type RedisConfig struct {
	// === 基础配置 ===
	Enabled bool           `yaml:"enabled" json:"enabled" mapstructure:"enabled"`       // 是否启用该连接
	Mode    ConnectionMode `yaml:"mode" json:"mode" mapstructure:"mode"`               // 连接模式: single, sentinel, cluster

	// === 单机模式连接配置 ===
	Host     string `yaml:"host" json:"host" mapstructure:"host"`         // Redis服务器地址，例如: localhost, 192.168.1.100
	Port     int    `yaml:"port" json:"port" mapstructure:"port"`         // Redis服务器端口，默认: 6379
	Password string `yaml:"password" json:"password" mapstructure:"password"` // Redis认证密码，如果Redis没有设置密码则留空
	DB       int    `yaml:"db" json:"db" mapstructure:"db"`               // Redis数据库编号，范围0-15，默认: 0

	// === 哨兵模式配置 ===
	SentinelAddrs    []string `yaml:"sentinel_addrs" json:"sentinel_addrs" mapstructure:"sentinel_addrs"`       // 哨兵地址列表，格式: ["host1:port1", "host2:port2"]
	MasterName       string   `yaml:"master_name" json:"master_name" mapstructure:"master_name"`               // 主节点名称，哨兵模式必需
	SentinelPassword string   `yaml:"sentinel_password" json:"sentinel_password" mapstructure:"sentinel_password"` // 哨兵认证密码
	SentinelUsername string   `yaml:"sentinel_username" json:"sentinel_username" mapstructure:"sentinel_username"` // 哨兵认证用户名

	// === 集群模式配置 ===
	ClusterAddrs           []string `yaml:"cluster_addrs" json:"cluster_addrs" mapstructure:"cluster_addrs"`                         // 集群节点地址列表，格式: ["host1:port1", "host2:port2"]
	ClusterUsername        string   `yaml:"cluster_username" json:"cluster_username" mapstructure:"cluster_username"`               // 集群认证用户名
	ClusterPassword        string   `yaml:"cluster_password" json:"cluster_password" mapstructure:"cluster_password"`               // 集群认证密码
	MaxRedirects           int      `yaml:"max_redirects" json:"max_redirects" mapstructure:"max_redirects"`                         // 最大重定向次数，默认: 3
	ReadOnly               bool     `yaml:"read_only" json:"read_only" mapstructure:"read_only"`                                     // 是否只读模式
	RouteByLatency         bool     `yaml:"route_by_latency" json:"route_by_latency" mapstructure:"route_by_latency"`               // 是否根据延迟路由
	RouteRandomly          bool     `yaml:"route_randomly" json:"route_randomly" mapstructure:"route_randomly"`                     // 是否随机路由

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
	KeyPrefix         string        `yaml:"key_prefix" json:"key_prefix" mapstructure:"key_prefix"`                 // 缓存键前缀，用于区分不同应用或环境，例如: "myapp:prod"
	KeyExpire         int           `yaml:"key_expire" json:"key_expire" mapstructure:"key_expire"`                 // 默认过期时间（秒），0表示永不过期
	DefaultExpiration time.Duration `yaml:"-" json:"-" mapstructure:"-"`                                            // 默认过期时间的Duration格式（内部使用）
	EnablePipelining  bool          `yaml:"enable_pipelining" json:"enable_pipelining" mapstructure:"enable_pipelining"`   // 是否启用管道模式提高批量操作性能，默认: true
	EnableCompression bool          `yaml:"enable_compression" json:"enable_compression" mapstructure:"enable_compression"` // 是否启用数据压缩，适用于大数据量场景，默认: false

	// === TLS/SSL配置 ===
	TLSEnabled            bool   `yaml:"tls_enabled" json:"tls_enabled" mapstructure:"tls_enabled"`                         // 是否启用TLS
	TLSCertFile           string `yaml:"tls_cert_file" json:"tls_cert_file" mapstructure:"tls_cert_file"`                   // TLS证书文件路径
	TLSKeyFile            string `yaml:"tls_key_file" json:"tls_key_file" mapstructure:"tls_key_file"`                      // TLS私钥文件路径
	TLSCACertFile         string `yaml:"tls_ca_cert_file" json:"tls_ca_cert_file" mapstructure:"tls_ca_cert_file"`          // TLS CA证书文件路径
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify" json:"tls_insecure_skip_verify" mapstructure:"tls_insecure_skip_verify"` // 是否跳过TLS证书验证

	// === 监控配置 ===
	EnableMetrics    bool `yaml:"enable_metrics" json:"enable_metrics" mapstructure:"enable_metrics"`       // 是否启用指标收集
	MetricsNamespace string `yaml:"metrics_namespace" json:"metrics_namespace" mapstructure:"metrics_namespace"` // 指标命名空间
}

// GetType 实现CacheConfig接口，返回缓存类型
func (r *RedisConfig) GetType() string {
	return "redis"
}

// Validate 验证配置的有效性
func (r *RedisConfig) Validate() error {
	// 验证连接模式
	switch r.Mode {
	case ModeSingle:
		return r.validateSingleMode()
	case ModeSentinel:
		return r.validateSentinelMode()
	case ModeCluster:
		return r.validateClusterMode()
	default:
		return fmt.Errorf("不支持的连接模式: %s，支持的模式: single, sentinel, cluster", r.Mode)
	}
}

// validateSingleMode 验证单机模式配置
func (r *RedisConfig) validateSingleMode() error {
	if r.Host == "" {
		return fmt.Errorf("单机模式下host不能为空")
	}

	if r.Port <= 0 || r.Port > 65535 {
		return fmt.Errorf("单机模式下port必须在1-65535之间，当前值: %d", r.Port)
	}

	if r.DB < 0 || r.DB > 15 {
		return fmt.Errorf("数据库编号必须在0-15之间，当前值: %d", r.DB)
	}

	return nil
}

// validateSentinelMode 验证哨兵模式配置
func (r *RedisConfig) validateSentinelMode() error {
	if len(r.SentinelAddrs) == 0 {
		return fmt.Errorf("哨兵模式下sentinel_addrs不能为空")
	}

	if r.MasterName == "" {
		return fmt.Errorf("哨兵模式下master_name不能为空")
	}

	// 验证哨兵地址格式
	for i, addr := range r.SentinelAddrs {
		if !strings.Contains(addr, ":") {
			return fmt.Errorf("哨兵地址格式错误 [%d]: %s，应为 host:port 格式", i, addr)
		}
	}

	if r.DB < 0 || r.DB > 15 {
		return fmt.Errorf("数据库编号必须在0-15之间，当前值: %d", r.DB)
	}

	return nil
}

// validateClusterMode 验证集群模式配置
func (r *RedisConfig) validateClusterMode() error {
	if len(r.ClusterAddrs) == 0 {
		return fmt.Errorf("集群模式下cluster_addrs不能为空")
	}

	// 集群模式不支持数据库选择
	if r.DB != 0 {
		return fmt.Errorf("集群模式不支持数据库选择，DB必须为0，当前值: %d", r.DB)
	}

	// 验证集群地址格式
	for i, addr := range r.ClusterAddrs {
		if !strings.Contains(addr, ":") {
			return fmt.Errorf("集群地址格式错误 [%d]: %s，应为 host:port 格式", i, addr)
		}
	}

	if r.MaxRedirects < 0 {
		return fmt.Errorf("最大重定向次数不能为负数，当前值: %d", r.MaxRedirects)
	}

	return nil
}

// SetDefaults 设置默认值
func (r *RedisConfig) SetDefaults() {
	// 基础配置默认值
	if r.Mode == "" {
		r.Mode = ModeSingle
	}

	// 单机模式默认值
	if r.Mode == ModeSingle && r.Port == 0 {
		r.Port = 6379
	}

	// 集群模式默认值
	if r.Mode == ModeCluster {
		if r.MaxRedirects == 0 {
			r.MaxRedirects = 3
		}
		// 集群模式强制DB为0
		r.DB = 0
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
	r.EnablePipelining = true // 默认启用管道

	// 默认过期时间处理
	if r.KeyExpire > 0 {
		r.DefaultExpiration = time.Duration(r.KeyExpire) * time.Second
	} else {
		r.DefaultExpiration = 0 // 永不过期
	}

	// 监控配置默认值
	if r.MetricsNamespace == "" {
		r.MetricsNamespace = "redis"
	}
}

// GetAddress 获取Redis服务器地址（仅单机模式）
func (r *RedisConfig) GetAddress() string {
	if r.Mode == ModeSingle {
		return fmt.Sprintf("%s:%d", r.Host, r.Port)
	}
	return ""
}

// GetSentinelAddresses 获取哨兵地址列表（仅哨兵模式）
func (r *RedisConfig) GetSentinelAddresses() []string {
	if r.Mode == ModeSentinel {
		return r.SentinelAddrs
	}
	return nil
}

// GetClusterAddresses 获取集群地址列表（仅集群模式）
func (r *RedisConfig) GetClusterAddresses() []string {
	if r.Mode == ModeCluster {
		return r.ClusterAddrs
	}
	return nil
}

// GetIdleTimeoutDuration 获取空闲超时持续时间
func (r *RedisConfig) GetIdleTimeoutDuration() time.Duration {
	return time.Duration(r.IdleTimeout) * time.Millisecond
}

// GetDefaultExpiration 获取默认过期时间
// 返回配置的默认过期时间，如果为0则表示永不过期
func (r *RedisConfig) GetDefaultExpiration() time.Duration {
	return r.DefaultExpiration
}

// IsClusterMode 检查是否为集群模式
func (r *RedisConfig) IsClusterMode() bool {
	return r.Mode == ModeCluster
}

// IsSentinelMode 检查是否为哨兵模式
func (r *RedisConfig) IsSentinelMode() bool {
	return r.Mode == ModeSentinel
}

// IsSingleMode 检查是否为单机模式
func (r *RedisConfig) IsSingleMode() bool {
	return r.Mode == ModeSingle
}

// GetConnectionString 获取连接字符串描述（用于日志）
func (r *RedisConfig) GetConnectionString() string {
	switch r.Mode {
	case ModeSingle:
		return fmt.Sprintf("single://%s:%d/%d", r.Host, r.Port, r.DB)
	case ModeSentinel:
		return fmt.Sprintf("sentinel://%s@%s/%d", r.MasterName, strings.Join(r.SentinelAddrs, ","), r.DB)
	case ModeCluster:
		return fmt.Sprintf("cluster://%s", strings.Join(r.ClusterAddrs, ","))
	default:
		return "unknown"
	}
}

// String 返回配置的字符串表示（隐藏敏感信息）
func (r *RedisConfig) String() string {
	password := r.Password
	if password != "" {
		password = "***"
	}

	switch r.Mode {
	case ModeSingle:
		return fmt.Sprintf("Redis{Mode:%s, Host:%s, Port:%d, Password:%s, DB:%d, PoolSize:%d}",
			r.Mode, r.Host, r.Port, password, r.DB, r.PoolSize)
	case ModeSentinel:
		return fmt.Sprintf("Redis{Mode:%s, MasterName:%s, Sentinels:%d, Password:%s, DB:%d, PoolSize:%d}",
			r.Mode, r.MasterName, len(r.SentinelAddrs), password, r.DB, r.PoolSize)
	case ModeCluster:
		clusterPassword := r.ClusterPassword
		if clusterPassword != "" {
			clusterPassword = "***"
		}
		return fmt.Sprintf("Redis{Mode:%s, Nodes:%d, Password:%s, PoolSize:%d}",
			r.Mode, len(r.ClusterAddrs), clusterPassword, r.PoolSize)
	default:
		return fmt.Sprintf("Redis{Mode:%s, PoolSize:%d}", r.Mode, r.PoolSize)
	}
}
