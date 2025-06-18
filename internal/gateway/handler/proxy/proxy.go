package proxy

import (
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/service"
)

// ProxyType 代理类型
type ProxyType string

const (
	// ProxyTypeHTTP HTTP代理类型
	ProxyTypeHTTP ProxyType = "http"
	// ProxyTypeWebSocket WebSocket代理类型
	ProxyTypeWebSocket ProxyType = "websocket"
	// ProxyTypeTCP TCP代理类型
	ProxyTypeTCP ProxyType = "tcp"
	// ProxyTypeUDP UDP代理类型
	ProxyTypeUDP ProxyType = "udp"
)

// ProxyConfig 基础代理配置
type ProxyConfig struct {
	ID           string                   `yaml:"id" json:"id" mapstructure:"id"`                                                                // 代理ID
	Enabled      bool                     `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                                                 // 是否启用代理
	Type         ProxyType                `yaml:"type" json:"type" mapstructure:"type"`                                                          // 代理类型
	Name         string                   `yaml:"name,omitempty" json:"name,omitempty" mapstructure:"name,omitempty"`                            // 代理名称
	Service      []*service.ServiceConfig `yaml:"service,omitempty" json:"service" mapstructure:"service"`                                       // 服务配置
	Config       map[string]interface{}   `yaml:"config,omitempty" json:"config,omitempty" mapstructure:"config,omitempty"`                      // 具体配置
	CustomConfig map[string]interface{}   `yaml:"custom_config,omitempty" json:"custom_config,omitempty" mapstructure:"custom_config,omitempty"` // 自定义配置
}

// HTTPProxyConfig HTTP代理具体配置
type HTTPProxyConfig struct {
	ID               string        `yaml:"id" json:"id" mapstructure:"id"`                                                 // HTTP代理配置ID
	Timeout          time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout"`                                  // 超时时间
	FollowRedirects  bool          `yaml:"follow_redirects" json:"follow_redirects" mapstructure:"follow_redirects"`       // 是否跟随重定向
	KeepAlive        bool          `yaml:"keep_alive" json:"keep_alive" mapstructure:"keep_alive"`                         // 是否保持连接
	MaxIdleConns     int           `yaml:"max_idle_conns" json:"max_idle_conns" mapstructure:"max_idle_conns"`             // 最大空闲连接数
	IdleConnTimeout  time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout" mapstructure:"idle_conn_timeout"`    // 空闲连接超时
	CopyResponseBody bool          `yaml:"copy_response_body" json:"copy_response_body" mapstructure:"copy_response_body"` // 是否复制响应体
	BufferSize       int           `yaml:"buffer_size" json:"buffer_size" mapstructure:"buffer_size"`                      // 缓冲区大小
	MaxBufferSize    int           `yaml:"max_buffer_size" json:"max_buffer_size" mapstructure:"max_buffer_size"`          // 最大缓冲区大小
	RetryCount       int           `yaml:"retry_count" json:"retry_count" mapstructure:"retry_count"`                      // 重试次数
	RetryTimeout     time.Duration `yaml:"retry_timeout" json:"retry_timeout" mapstructure:"retry_timeout"`                // 重试超时
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	ID             string        `yaml:"id" json:"id" mapstructure:"id"`                                                             // WebSocket配置ID
	Enabled        bool          `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                                              // 是否启用WebSocket
	PingInterval   time.Duration `yaml:"ping_interval" json:"ping_interval" mapstructure:"ping_interval"`                            // Ping间隔
	PongTimeout    time.Duration `yaml:"pong_timeout" json:"pong_timeout" mapstructure:"pong_timeout"`                               // Pong超时
	WriteTimeout   time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"`                            // 写超时
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`                               // 读超时
	MaxMessageSize int64         `yaml:"max_message_size" json:"max_message_size" mapstructure:"max_message_size"`                   // 最大消息大小
	Subprotocols   []string      `yaml:"subprotocols,omitempty" json:"subprotocols,omitempty" mapstructure:"subprotocols,omitempty"` // 子协议
}

// TCPConfig TCP配置
type TCPConfig struct {
	ID           string        `yaml:"id" json:"id" mapstructure:"id"`                                  // TCP配置ID
	Enabled      bool          `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                   // 是否启用TCP
	KeepAlive    bool          `yaml:"keep_alive" json:"keep_alive" mapstructure:"keep_alive"`          // 是否保持连接
	NoDelay      bool          `yaml:"no_delay" json:"no_delay" mapstructure:"no_delay"`                // 是否禁用Nagle算法
	BufferSize   int           `yaml:"buffer_size" json:"buffer_size" mapstructure:"buffer_size"`       // 缓冲区大小
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`    // 读超时
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"` // 写超时
}

// ProxyHandler 代理处理器接口
// 所有代理处理器都必须实现此接口
type ProxyHandler interface {
	// Handle 处理代理请求
	Handle(ctx *core.Context) bool

	// GetType 获取代理类型
	GetType() ProxyType

	// IsEnabled 是否启用
	IsEnabled() bool

	// GetName 获取代理名称
	GetName() string

	// GetConfig 获取配置
	GetConfig() ProxyConfig

	// ProxyRequest 代理请求到指定URL
	ProxyRequest(ctx *core.Context, targetURL string) error

	// Validate 验证配置
	Validate() error

	// Close 关闭代理处理器
	Close() error
}

// BaseProxyHandler 代理处理器基础结构
// 提供代理处理器的基础实现和通用功能
type BaseProxyHandler struct {
	enabled        bool
	proxyType      ProxyType
	name           string
	originalConfig ProxyConfig
}

// NewBaseProxyHandler 创建基础代理处理器
func NewBaseProxyHandler(proxyType ProxyType, enabled bool, name string) *BaseProxyHandler {
	config := ProxyConfig{
		Type:    proxyType,
		Enabled: enabled,
		Name:    name,
		Config:  make(map[string]interface{}),
	}

	return &BaseProxyHandler{
		enabled:        enabled,
		proxyType:      proxyType,
		name:           name,
		originalConfig: config,
	}
}

// GetType 获取代理类型
func (b *BaseProxyHandler) GetType() ProxyType {
	return b.proxyType
}

// IsEnabled 是否启用
func (b *BaseProxyHandler) IsEnabled() bool {
	return b.enabled
}

// GetName 获取代理名称
func (b *BaseProxyHandler) GetName() string {
	return b.name
}

// GetConfig 获取配置
func (b *BaseProxyHandler) GetConfig() ProxyConfig {
	return b.originalConfig
}

// SetName 设置代理名称
func (b *BaseProxyHandler) SetName(name string) {
	if name != "" {
		b.name = name
		b.originalConfig.Name = name
	}
}

// SetEnabled 设置是否启用
func (b *BaseProxyHandler) SetEnabled(enabled bool) {
	b.enabled = enabled
	b.originalConfig.Enabled = enabled
}

// Handle 处理代理请求（基础实现：总是允许通过）
func (b *BaseProxyHandler) Handle(ctx *core.Context) bool {
	return true
}

// ProxyRequest 代理请求到指定URL（基础实现）
func (b *BaseProxyHandler) ProxyRequest(ctx *core.Context, targetURL string) error {
	return nil
}

// Validate 验证配置（基础实现：总是通过验证）
func (b *BaseProxyHandler) Validate() error {
	return nil
}

// Close 关闭代理处理器（基础实现）
func (b *BaseProxyHandler) Close() error {
	return nil
}

// 默认配置
var DefaultProxyConfig = ProxyConfig{
	ID:      "default-proxy",
	Enabled: true,
	Type:    ProxyTypeHTTP,
	Name:    "Default Proxy",
	Config:  make(map[string]interface{}),
}

var DefaultHTTPProxyConfig = HTTPProxyConfig{
	ID:               "default-http-proxy",
	Timeout:          10 * time.Second,
	FollowRedirects:  true,
	KeepAlive:        true,
	MaxIdleConns:     100,
	IdleConnTimeout:  90 * time.Second,
	CopyResponseBody: false,
	BufferSize:       32 * 1024,   // 32KB
	MaxBufferSize:    1024 * 1024, // 1MB
	RetryCount:       3,
	RetryTimeout:     5 * time.Second,
}

var DefaultWebSocketConfig = WebSocketConfig{
	ID:             "default-websocket",
	Enabled:        true,
	PingInterval:   30 * time.Second,
	PongTimeout:    10 * time.Second,
	WriteTimeout:   10 * time.Second,
	ReadTimeout:    60 * time.Second,
	MaxMessageSize: 1024 * 1024, // 1MB
}

var DefaultTCPConfig = TCPConfig{
	ID:           "default-tcp",
	Enabled:      true,
	KeepAlive:    true,
	NoDelay:      true,
	BufferSize:   32 * 1024, // 32KB
	ReadTimeout:  30 * time.Second,
	WriteTimeout: 30 * time.Second,
}
