package config

import (
	"time"
)

// Config 注册中心配置
type Config struct {
	// 服务器配置
	Server ServerConfig `yaml:"server" json:"server"`

	// 数据库配置
	Database DatabaseConfig `yaml:"database" json:"database"`

	// 健康检查配置
	HealthCheck HealthCheckConfig `yaml:"healthCheck" json:"healthCheck"`

	// 事件配置
	Event EventConfig `yaml:"event" json:"event"`

	// 外部注册中心配置
	External ExternalConfig `yaml:"external" json:"external"`

	// 日志配置
	Logging LoggingConfig `yaml:"logging" json:"logging"`

	// 监控配置
	Monitoring MonitoringConfig `yaml:"monitoring" json:"monitoring"`

	// 安全配置
	Security SecurityConfig `yaml:"security" json:"security"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// HTTP服务器配置
	HTTP HTTPConfig `yaml:"http" json:"http"`

	// gRPC服务器配置
	GRPC GRPCConfig `yaml:"grpc" json:"grpc"`

	// WebSocket配置
	WebSocket WebSocketConfig `yaml:"websocket" json:"websocket"`

	// 优雅关闭配置
	GracefulShutdown GracefulShutdownConfig `yaml:"gracefulShutdown" json:"gracefulShutdown"`
}

// HTTPConfig HTTP服务器配置
type HTTPConfig struct {
	// 监听地址
	Host string `yaml:"host" json:"host"`

	// 监听端口
	Port int `yaml:"port" json:"port"`

	// 读取超时
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout"`

	// 写入超时
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout"`

	// 空闲超时
	IdleTimeout time.Duration `yaml:"idleTimeout" json:"idleTimeout"`

	// 最大请求头大小
	MaxHeaderBytes int `yaml:"maxHeaderBytes" json:"maxHeaderBytes"`

	// 启用CORS
	EnableCORS bool `yaml:"enableCORS" json:"enableCORS"`

	// CORS配置
	CORS CORSConfig `yaml:"cors" json:"cors"`

	// 启用压缩
	EnableGzip bool `yaml:"enableGzip" json:"enableGzip"`

	// 启用请求日志
	EnableRequestLog bool `yaml:"enableRequestLog" json:"enableRequestLog"`
}

// GRPCConfig gRPC服务器配置
type GRPCConfig struct {
	// 启用gRPC
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 监听地址
	Host string `yaml:"host" json:"host"`

	// 监听端口
	Port int `yaml:"port" json:"port"`

	// 连接超时
	ConnectionTimeout time.Duration `yaml:"connectionTimeout" json:"connectionTimeout"`

	// 最大接收消息大小
	MaxRecvMsgSize int `yaml:"maxRecvMsgSize" json:"maxRecvMsgSize"`

	// 最大发送消息大小
	MaxSendMsgSize int `yaml:"maxSendMsgSize" json:"maxSendMsgSize"`

	// 启用TLS
	EnableTLS bool `yaml:"enableTLS" json:"enableTLS"`

	// TLS配置
	TLS TLSConfig `yaml:"tls" json:"tls"`
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	// 启用WebSocket
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 路径
	Path string `yaml:"path" json:"path"`

	// 读取缓冲区大小
	ReadBufferSize int `yaml:"readBufferSize" json:"readBufferSize"`

	// 写入缓冲区大小
	WriteBufferSize int `yaml:"writeBufferSize" json:"writeBufferSize"`

	// 握手超时
	HandshakeTimeout time.Duration `yaml:"handshakeTimeout" json:"handshakeTimeout"`

	// 心跳间隔
	PingInterval time.Duration `yaml:"pingInterval" json:"pingInterval"`

	// 心跳超时
	PongTimeout time.Duration `yaml:"pongTimeout" json:"pongTimeout"`

	// 最大连接数
	MaxConnections int `yaml:"maxConnections" json:"maxConnections"`
}

// CORSConfig CORS配置
type CORSConfig struct {
	// 允许的源
	AllowedOrigins []string `yaml:"allowedOrigins" json:"allowedOrigins"`

	// 允许的方法
	AllowedMethods []string `yaml:"allowedMethods" json:"allowedMethods"`

	// 允许的头部
	AllowedHeaders []string `yaml:"allowedHeaders" json:"allowedHeaders"`

	// 暴露的头部
	ExposedHeaders []string `yaml:"exposedHeaders" json:"exposedHeaders"`

	// 允许凭据
	AllowCredentials bool `yaml:"allowCredentials" json:"allowCredentials"`

	// 预检请求缓存时间
	MaxAge time.Duration `yaml:"maxAge" json:"maxAge"`
}

// TLSConfig TLS配置
type TLSConfig struct {
	// 证书文件
	CertFile string `yaml:"certFile" json:"certFile"`

	// 私钥文件
	KeyFile string `yaml:"keyFile" json:"keyFile"`

	// CA证书文件
	CAFile string `yaml:"caFile" json:"caFile"`

	// 客户端认证
	ClientAuth string `yaml:"clientAuth" json:"clientAuth"`
}

// GracefulShutdownConfig 优雅关闭配置
type GracefulShutdownConfig struct {
	// 超时时间
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// 等待连接关闭
	WaitForConnections bool `yaml:"waitForConnections" json:"waitForConnections"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	// 数据库类型
	Type string `yaml:"type" json:"type"`

	// 连接字符串
	DSN string `yaml:"dsn" json:"dsn"`

	// 最大打开连接数
	MaxOpenConns int `yaml:"maxOpenConns" json:"maxOpenConns"`

	// 最大空闲连接数
	MaxIdleConns int `yaml:"maxIdleConns" json:"maxIdleConns"`

	// 连接最大生存时间
	ConnMaxLifetime time.Duration `yaml:"connMaxLifetime" json:"connMaxLifetime"`

	// 连接最大空闲时间
	ConnMaxIdleTime time.Duration `yaml:"connMaxIdleTime" json:"connMaxIdleTime"`

	// 启用日志
	EnableLog bool `yaml:"enableLog" json:"enableLog"`

	// 日志级别
	LogLevel string `yaml:"logLevel" json:"logLevel"`

	// 慢查询阈值
	SlowThreshold time.Duration `yaml:"slowThreshold" json:"slowThreshold"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	// 启用健康检查
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 检查间隔
	Interval time.Duration `yaml:"interval" json:"interval"`

	// 超时时间
	Timeout time.Duration `yaml:"timeout" json:"timeout"`

	// 重试次数
	MaxRetries int `yaml:"maxRetries" json:"maxRetries"`

	// 重试间隔
	RetryInterval time.Duration `yaml:"retryInterval" json:"retryInterval"`

	// 并发检查数
	ConcurrentChecks int `yaml:"concurrentChecks" json:"concurrentChecks"`

	// 失败阈值
	FailureThreshold int `yaml:"failureThreshold" json:"failureThreshold"`

	// 成功阈值
	SuccessThreshold int `yaml:"successThreshold" json:"successThreshold"`

	// 默认健康检查路径
	DefaultPath string `yaml:"defaultPath" json:"defaultPath"`

	// 启用TCP检查
	EnableTCPCheck bool `yaml:"enableTCPCheck" json:"enableTCPCheck"`

	// 启用HTTP检查
	EnableHTTPCheck bool `yaml:"enableHTTPCheck" json:"enableHTTPCheck"`
}

// EventConfig 事件配置
type EventConfig struct {
	// 启用事件系统
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 事件缓冲区大小
	BufferSize int `yaml:"bufferSize" json:"bufferSize"`

	// 工作协程数
	WorkerCount int `yaml:"workerCount" json:"workerCount"`

	// 批处理大小
	BatchSize int `yaml:"batchSize" json:"batchSize"`

	// 批处理超时
	BatchTimeout time.Duration `yaml:"batchTimeout" json:"batchTimeout"`

	// 最大订阅者数
	MaxSubscribers int `yaml:"maxSubscribers" json:"maxSubscribers"`

	// 订阅者超时
	SubscriberTimeout time.Duration `yaml:"subscriberTimeout" json:"subscriberTimeout"`

	// 启用事件持久化
	EnablePersistence bool `yaml:"enablePersistence" json:"enablePersistence"`

	// 事件保留时间
	RetentionPeriod time.Duration `yaml:"retentionPeriod" json:"retentionPeriod"`

	// 清理间隔
	CleanupInterval time.Duration `yaml:"cleanupInterval" json:"cleanupInterval"`
}

// ExternalConfig 外部注册中心配置
type ExternalConfig struct {
	// 启用外部注册中心
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 配置刷新间隔
	RefreshInterval time.Duration `yaml:"refreshInterval" json:"refreshInterval"`

	// 连接池大小
	PoolSize int `yaml:"poolSize" json:"poolSize"`

	// 连接超时
	ConnectTimeout time.Duration `yaml:"connectTimeout" json:"connectTimeout"`

	// 读取超时
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout"`

	// 写入超时
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout"`

	// 最大重试次数
	MaxRetries int `yaml:"maxRetries" json:"maxRetries"`

	// 重试间隔
	RetryInterval time.Duration `yaml:"retryInterval" json:"retryInterval"`

	// 启用故障转移
	EnableFailover bool `yaml:"enableFailover" json:"enableFailover"`

	// 故障转移检查间隔
	FailoverCheckInterval time.Duration `yaml:"failoverCheckInterval" json:"failoverCheckInterval"`

	// 启用数据同步
	EnableSync bool `yaml:"enableSync" json:"enableSync"`

	// 同步间隔
	SyncInterval time.Duration `yaml:"syncInterval" json:"syncInterval"`

	// 同步批大小
	SyncBatchSize int `yaml:"syncBatchSize" json:"syncBatchSize"`

	// 冲突解决策略
	ConflictResolution string `yaml:"conflictResolution" json:"conflictResolution"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	// 日志级别
	Level string `yaml:"level" json:"level"`

	// 日志格式
	Format string `yaml:"format" json:"format"`

	// 输出目标
	Output []string `yaml:"output" json:"output"`

	// 文件配置
	File FileLogConfig `yaml:"file" json:"file"`

	// 启用结构化日志
	EnableStructured bool `yaml:"enableStructured" json:"enableStructured"`

	// 启用调用者信息
	EnableCaller bool `yaml:"enableCaller" json:"enableCaller"`

	// 启用堆栈跟踪
	EnableStacktrace bool `yaml:"enableStacktrace" json:"enableStacktrace"`
}

// FileLogConfig 文件日志配置
type FileLogConfig struct {
	// 文件名
	Filename string `yaml:"filename" json:"filename"`

	// 最大文件大小(MB)
	MaxSize int `yaml:"maxSize" json:"maxSize"`

	// 最大备份数
	MaxBackups int `yaml:"maxBackups" json:"maxBackups"`

	// 最大保留天数
	MaxAge int `yaml:"maxAge" json:"maxAge"`

	// 启用压缩
	Compress bool `yaml:"compress" json:"compress"`

	// 本地时间
	LocalTime bool `yaml:"localTime" json:"localTime"`
}

// MonitoringConfig 监控配置
type MonitoringConfig struct {
	// 启用监控
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Prometheus配置
	Prometheus PrometheusConfig `yaml:"prometheus" json:"prometheus"`

	// 指标收集间隔
	MetricsInterval time.Duration `yaml:"metricsInterval" json:"metricsInterval"`

	// 启用性能分析
	EnableProfiling bool `yaml:"enableProfiling" json:"enableProfiling"`

	// 性能分析端口
	ProfilingPort int `yaml:"profilingPort" json:"profilingPort"`
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	// 启用Prometheus
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 监听地址
	Host string `yaml:"host" json:"host"`

	// 监听端口
	Port int `yaml:"port" json:"port"`

	// 指标路径
	Path string `yaml:"path" json:"path"`

	// 命名空间
	Namespace string `yaml:"namespace" json:"namespace"`

	// 子系统
	Subsystem string `yaml:"subsystem" json:"subsystem"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	// 启用认证
	EnableAuth bool `yaml:"enableAuth" json:"enableAuth"`

	// JWT配置
	JWT JWTConfig `yaml:"jwt" json:"jwt"`

	// API密钥配置
	APIKey APIKeyConfig `yaml:"apiKey" json:"apiKey"`

	// 限流配置
	RateLimit RateLimitConfig `yaml:"rateLimit" json:"rateLimit"`

	// IP白名单
	IPWhitelist []string `yaml:"ipWhitelist" json:"ipWhitelist"`

	// IP黑名单
	IPBlacklist []string `yaml:"ipBlacklist" json:"ipBlacklist"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	// 启用JWT
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 签名密钥
	Secret string `yaml:"secret" json:"secret"`

	// 过期时间
	ExpirationTime time.Duration `yaml:"expirationTime" json:"expirationTime"`

	// 刷新时间
	RefreshTime time.Duration `yaml:"refreshTime" json:"refreshTime"`

	// 签名算法
	Algorithm string `yaml:"algorithm" json:"algorithm"`

	// 发行者
	Issuer string `yaml:"issuer" json:"issuer"`
}

// APIKeyConfig API密钥配置
type APIKeyConfig struct {
	// 启用API密钥
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 头部名称
	HeaderName string `yaml:"headerName" json:"headerName"`

	// 查询参数名称
	QueryName string `yaml:"queryName" json:"queryName"`

	// 有效密钥列表
	ValidKeys []string `yaml:"validKeys" json:"validKeys"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// 启用限流
	Enabled bool `yaml:"enabled" json:"enabled"`

	// 每秒请求数
	RequestsPerSecond int `yaml:"requestsPerSecond" json:"requestsPerSecond"`

	// 突发请求数
	Burst int `yaml:"burst" json:"burst"`

	// 限流策略
	Strategy string `yaml:"strategy" json:"strategy"`

	// 限流存储
	Storage string `yaml:"storage" json:"storage"`

	// Redis配置（如果使用Redis存储）
	Redis RedisConfig `yaml:"redis" json:"redis"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	// 地址
	Addr string `yaml:"addr" json:"addr"`

	// 密码
	Password string `yaml:"password" json:"password"`

	// 数据库
	DB int `yaml:"db" json:"db"`

	// 连接池大小
	PoolSize int `yaml:"poolSize" json:"poolSize"`

	// 最小空闲连接数
	MinIdleConns int `yaml:"minIdleConns" json:"minIdleConns"`

	// 连接超时
	DialTimeout time.Duration `yaml:"dialTimeout" json:"dialTimeout"`

	// 读取超时
	ReadTimeout time.Duration `yaml:"readTimeout" json:"readTimeout"`

	// 写入超时
	WriteTimeout time.Duration `yaml:"writeTimeout" json:"writeTimeout"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			HTTP: HTTPConfig{
				Host:           "0.0.0.0",
				Port:           8080,
				ReadTimeout:    30 * time.Second,
				WriteTimeout:   30 * time.Second,
				IdleTimeout:    60 * time.Second,
				MaxHeaderBytes: 1 << 20, // 1MB
				EnableCORS:     true,
				CORS: CORSConfig{
					AllowedOrigins:   []string{"*"},
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
					AllowedHeaders:   []string{"*"},
					AllowCredentials: false,
					MaxAge:           12 * time.Hour,
				},
				EnableGzip:       true,
				EnableRequestLog: true,
			},
			GRPC: GRPCConfig{
				Enabled:           false,
				Host:              "0.0.0.0",
				Port:              9090,
				ConnectionTimeout: 30 * time.Second,
				MaxRecvMsgSize:    4 * 1024 * 1024, // 4MB
				MaxSendMsgSize:    4 * 1024 * 1024, // 4MB
				EnableTLS:         false,
			},
			WebSocket: WebSocketConfig{
				Enabled:          true,
				Path:             "/ws",
				ReadBufferSize:   1024,
				WriteBufferSize:  1024,
				HandshakeTimeout: 10 * time.Second,
				PingInterval:     30 * time.Second,
				PongTimeout:      60 * time.Second,
				MaxConnections:   1000,
			},
			GracefulShutdown: GracefulShutdownConfig{
				Timeout:            30 * time.Second,
				WaitForConnections: true,
			},
		},
		Database: DatabaseConfig{
			Type:            "mysql",
			MaxOpenConns:    100,
			MaxIdleConns:    10,
			ConnMaxLifetime: time.Hour,
			ConnMaxIdleTime: 10 * time.Minute,
			EnableLog:       true,
			LogLevel:        "warn",
			SlowThreshold:   200 * time.Millisecond,
		},
		HealthCheck: HealthCheckConfig{
			Enabled:          true,
			Interval:         30 * time.Second,
			Timeout:          5 * time.Second,
			MaxRetries:       3,
			RetryInterval:    5 * time.Second,
			ConcurrentChecks: 10,
			FailureThreshold: 3,
			SuccessThreshold: 1,
			DefaultPath:      "/health",
			EnableTCPCheck:   true,
			EnableHTTPCheck:  true,
		},
		Event: EventConfig{
			Enabled:           true,
			BufferSize:        1000,
			WorkerCount:       5,
			BatchSize:         100,
			BatchTimeout:      1 * time.Second,
			MaxSubscribers:    100,
			SubscriberTimeout: 30 * time.Second,
			EnablePersistence: true,
			RetentionPeriod:   7 * 24 * time.Hour, // 7天
			CleanupInterval:   1 * time.Hour,
		},
		External: ExternalConfig{
			Enabled:               false,
			RefreshInterval:       30 * time.Second,
			PoolSize:              10,
			ConnectTimeout:        10 * time.Second,
			ReadTimeout:           30 * time.Second,
			WriteTimeout:          30 * time.Second,
			MaxRetries:            3,
			RetryInterval:         5 * time.Second,
			EnableFailover:        true,
			FailoverCheckInterval: 30 * time.Second,
			EnableSync:            false,
			SyncInterval:          5 * time.Minute,
			SyncBatchSize:         100,
			ConflictResolution:    "primary_wins",
		},
		Logging: LoggingConfig{
			Level:  "info",
			Format: "json",
			Output: []string{"stdout"},
			File: FileLogConfig{
				Filename:   "logs/registry.log",
				MaxSize:    100, // MB
				MaxBackups: 10,
				MaxAge:     30, // days
				Compress:   true,
				LocalTime:  true,
			},
			EnableStructured: true,
			EnableCaller:     true,
			EnableStacktrace: false,
		},
		Monitoring: MonitoringConfig{
			Enabled: true,
			Prometheus: PrometheusConfig{
				Enabled:   true,
				Host:      "0.0.0.0",
				Port:      9091,
				Path:      "/metrics",
				Namespace: "registry",
				Subsystem: "server",
			},
			MetricsInterval: 15 * time.Second,
			EnableProfiling: false,
			ProfilingPort:   6060,
		},
		Security: SecurityConfig{
			EnableAuth: false,
			JWT: JWTConfig{
				Enabled:        false,
				ExpirationTime: 24 * time.Hour,
				RefreshTime:    7 * 24 * time.Hour,
				Algorithm:      "HS256",
				Issuer:         "registry-server",
			},
			APIKey: APIKeyConfig{
				Enabled:    false,
				HeaderName: "X-API-Key",
				QueryName:  "api_key",
			},
			RateLimit: RateLimitConfig{
				Enabled:           false,
				RequestsPerSecond: 100,
				Burst:             200,
				Strategy:          "fixed_window",
				Storage:           "memory",
			},
		},
	}
}
