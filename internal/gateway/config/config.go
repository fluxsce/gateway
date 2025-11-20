package config

import (
	"time"

	"gateway/internal/gateway/handler/auth"
	"gateway/internal/gateway/handler/cors"
	"gateway/internal/gateway/handler/limiter"
	"gateway/internal/gateway/handler/proxy"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/security"
	"gateway/internal/gateway/logwrite/types"
)

// GatewayConfig 网关总体配置
type GatewayConfig struct {
	InstanceID string `json:"instance_id" yaml:"instance_id" mapstructure:"instance_id"`
	// 基础配置
	Base BaseConfig `json:"base" yaml:"base" mapstructure:"base"`
	// 路由配置
	Router router.RouterConfig `json:"router" yaml:"router" mapstructure:"router"`
	// 代理配置 - 直接使用proxy模块的配置（包含负载均衡）
	Proxy proxy.ProxyConfig `json:"proxy" yaml:"proxy" mapstructure:"proxy"`
	// 安全配置
	Security security.SecurityConfig `json:"security" yaml:"security" mapstructure:"security"`
	// 认证配置
	Auth auth.AuthConfig `json:"auth" yaml:"auth" mapstructure:"auth"`
	// CORS配置
	CORS cors.CORSConfig `json:"cors" yaml:"cors" mapstructure:"cors"`
	// 限流配置
	RateLimit limiter.RateLimitConfig `json:"rate_limit" yaml:"rate_limit" mapstructure:"rate_limit"`
	// 注意：熔断器配置不在全局级别，而是在路由级别或服务级别进行配置
	Log types.LogConfig `json:"log" yaml:"log" mapstructure:"log"`
}

// BaseConfig 基础配置
type BaseConfig struct {
	// 监听地址
	Listen string `json:"listen" yaml:"listen" mapstructure:"listen"`
	// 服务名称
	Name string `json:"name" yaml:"name" mapstructure:"name"`
	// 读取超时
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout" mapstructure:"read_timeout"`
	// 写入超时
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout" mapstructure:"write_timeout"`
	// 空闲超时
	IdleTimeout time.Duration `json:"idle_timeout" yaml:"idle_timeout" mapstructure:"idle_timeout"`
	// 最大请求体大小
	MaxBodySize int64 `json:"max_body_size" yaml:"max_body_size" mapstructure:"max_body_size"`
	// 是否使用HTTPS
	EnableHTTPS bool `json:"enable_https" yaml:"enable_https" mapstructure:"enable_https"`
	// HTTPS证书
	CertFile string `json:"cert_file" yaml:"cert_file" mapstructure:"cert_file"`
	// HTTPS密钥
	KeyFile string `json:"key_file" yaml:"key_file" mapstructure:"key_file"`
	// 私钥密码（用于解密加密的私钥）
	KeyPassword string `json:"key_password" yaml:"key_password" mapstructure:"key_password"`
	// 是否启用Gin框架
	UseGin bool `json:"use_gin" yaml:"use_gin" mapstructure:"use_gin"`
	// 是否启用访问日志
	EnableAccessLog bool `json:"enable_access_log" yaml:"enable_access_log" mapstructure:"enable_access_log"`
	// 日志格式
	LogFormat string `json:"log_format" yaml:"log_format" mapstructure:"log_format"`
	// 日志级别
	LogLevel string `json:"log_level" yaml:"log_level" mapstructure:"log_level"`
	// 是否启用压缩
	EnableGzip bool `json:"enable_gzip" yaml:"enable_gzip" mapstructure:"enable_gzip"`
}

// DefaultGatewayConfig 默认网关配置
var DefaultGatewayConfig = GatewayConfig{
	Base: BaseConfig{
		Listen:          ":8080",
		Name:            "Gateway Gateway",
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		IdleTimeout:     120 * time.Second,
		MaxBodySize:     10 * 1024 * 1024, // 10MB
		EnableHTTPS:     false,
		UseGin:          true,
		EnableAccessLog: true,
		LogFormat:       "json",
		LogLevel:        "info",
		EnableGzip:      true,
	},
	Router: router.DefaultRouterConfig,
	// 直接使用proxy模块的默认配置
	Proxy: proxy.DefaultProxyConfig,
	// 安全配置
	Security: security.DefaultSecurityConfig,
	Auth: auth.AuthConfig{
		Enabled:  false,
		Strategy: auth.StrategyNoAuth,
	},
	CORS: cors.CORSConfig{
		Enabled:          true,
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{},
		AllowCredentials: false,
		MaxAge:           86400, // 24小时
	},
	RateLimit: limiter.RateLimitConfig{
		Enabled:         false,
		Algorithm:       limiter.AlgorithmTokenBucket,
		Rate:            100,
		Burst:           50,
		ErrorStatusCode: 429,
		ErrorMessage:    "Rate limit exceeded",
	},
}
