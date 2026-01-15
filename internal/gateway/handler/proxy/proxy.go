package proxy

import (
	"time"

	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/service"
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
	ID string `yaml:"id" json:"id" mapstructure:"id"` // HTTP代理配置ID

	// 超时配置
	Timeout        time.Duration `yaml:"timeout" json:"timeout" mapstructure:"timeout"`                         // 超时时间
	SendTimeout    time.Duration `yaml:"send_timeout" json:"send_timeout" mapstructure:"send_timeout"`          // 发送超时
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`          // 读取超时
	ConnectTimeout time.Duration `yaml:"connect_timeout" json:"connect_timeout" mapstructure:"connect_timeout"` // 连接超时

	// 连接配置
	FollowRedirects bool          `yaml:"follow_redirects" json:"follow_redirects" mapstructure:"follow_redirects"`    // 是否跟随重定向
	KeepAlive       bool          `yaml:"keep_alive" json:"keep_alive" mapstructure:"keep_alive"`                      // 是否保持连接
	MaxIdleConns    int           `yaml:"max_idle_conns" json:"max_idle_conns" mapstructure:"max_idle_conns"`          // 最大空闲连接数
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout" json:"idle_conn_timeout" mapstructure:"idle_conn_timeout"` // 空闲连接超时

	// 响应处理配置
	CopyResponseBody bool `yaml:"copy_response_body" json:"copy_response_body" mapstructure:"copy_response_body"` // 是否复制响应体
	BufferSize       int  `yaml:"buffer_size" json:"buffer_size" mapstructure:"buffer_size"`                      // 缓冲区大小
	MaxBufferSize    int  `yaml:"max_buffer_size" json:"max_buffer_size" mapstructure:"max_buffer_size"`          // 最大缓冲区大小

	// 重试配置
	RetryCount   int           `yaml:"retry_count" json:"retry_count" mapstructure:"retry_count"`       // 重试次数
	RetryTimeout time.Duration `yaml:"retry_timeout" json:"retry_timeout" mapstructure:"retry_timeout"` // 重试超时

	// 头部配置
	SetHeaders  map[string]string `yaml:"set_headers,omitempty" json:"set_headers,omitempty" mapstructure:"set_headers,omitempty"`    // 设置头部
	PassHeaders []string          `yaml:"pass_headers,omitempty" json:"pass_headers,omitempty" mapstructure:"pass_headers,omitempty"` // 允许传递的头部
	HideHeaders []string          `yaml:"hide_headers,omitempty" json:"hide_headers,omitempty" mapstructure:"hide_headers,omitempty"` // 隐藏的头部

	// 高级选项
	HTTPVersion        string `yaml:"http_version,omitempty" json:"http_version,omitempty" mapstructure:"http_version,omitempty"` // HTTP版本 "1.0" 或 "1.1"
	PreserveHost       bool   `yaml:"preserve_host" json:"preserve_host" mapstructure:"preserve_host"`                            // 是否保留原始Host头部
	AddXForwardedFor   bool   `yaml:"add_x_forwarded_for" json:"add_x_forwarded_for" mapstructure:"add_x_forwarded_for"`          // 是否添加X-Forwarded-For
	AddXRealIP         bool   `yaml:"add_x_real_ip" json:"add_x_real_ip" mapstructure:"add_x_real_ip"`                            // 是否添加X-Real-IP
	AddXForwardedProto bool   `yaml:"add_x_forwarded_proto" json:"add_x_forwarded_proto" mapstructure:"add_x_forwarded_proto"`    // 是否添加X-Forwarded-Proto

	// 新增nginx风格配置项 - 只新增原来没有的
	ProxyBuffering bool `yaml:"proxy_buffering,omitempty" json:"proxy_buffering,omitempty" mapstructure:"proxy_buffering,omitempty"` // 是否启用代理缓冲

	// TLS配置
	TLSInsecureSkipVerify bool   `yaml:"tls_insecure_skip_verify" json:"tls_insecure_skip_verify" mapstructure:"tls_insecure_skip_verify"`    // 是否跳过TLS证书验证
	TLSMinVersion         string `yaml:"tls_min_version,omitempty" json:"tls_min_version,omitempty" mapstructure:"tls_min_version,omitempty"` // 最小TLS版本 (1.0, 1.1, 1.2, 1.3)
	TLSMaxVersion         string `yaml:"tls_max_version,omitempty" json:"tls_max_version,omitempty" mapstructure:"tls_max_version,omitempty"` // 最大TLS版本 (1.0, 1.1, 1.2, 1.3)
	TLSServerName         string `yaml:"tls_server_name,omitempty" json:"tls_server_name,omitempty" mapstructure:"tls_server_name,omitempty"` // TLS服务器名称

}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	ID             string        `yaml:"id" json:"id" mapstructure:"id"`                                           // WebSocket配置ID
	Enabled        bool          `yaml:"enabled" json:"enabled" mapstructure:"enabled"`                            // 是否启用WebSocket
	PingInterval   time.Duration `yaml:"ping_interval" json:"ping_interval" mapstructure:"ping_interval"`          // Ping间隔
	PongTimeout    time.Duration `yaml:"pong_timeout" json:"pong_timeout" mapstructure:"pong_timeout"`             // Pong超时
	WriteTimeout   time.Duration `yaml:"write_timeout" json:"write_timeout" mapstructure:"write_timeout"`          // 写超时
	ReadTimeout    time.Duration `yaml:"read_timeout" json:"read_timeout" mapstructure:"read_timeout"`             // 读超时
	MaxMessageSize int64         `yaml:"max_message_size" json:"max_message_size" mapstructure:"max_message_size"` // 最大消息大小

	// 缓冲区配置
	ReadBufferSize  int `yaml:"read_buffer_size" json:"read_buffer_size" mapstructure:"read_buffer_size"`    // 读缓冲区大小
	WriteBufferSize int `yaml:"write_buffer_size" json:"write_buffer_size" mapstructure:"write_buffer_size"` // 写缓冲区大小

	// 压缩和协议配置
	EnableCompression bool     `yaml:"enable_compression" json:"enable_compression" mapstructure:"enable_compression"`             // 是否启用压缩
	Subprotocols      []string `yaml:"subprotocols,omitempty" json:"subprotocols,omitempty" mapstructure:"subprotocols,omitempty"` // 子协议
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
	GetType() string

	// IsEnabled 是否启用
	IsEnabled() bool

	// GetName 获取代理名称
	GetName() string

	// GetConfig 获取配置
	GetConfig() ProxyConfig

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
func (b *BaseProxyHandler) GetType() string {
	return string(b.proxyType)
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
	Timeout:          60 * time.Second,
	SendTimeout:      60 * time.Second,
	ReadTimeout:      60 * time.Second,
	ConnectTimeout:   60 * time.Second,
	FollowRedirects:  false,
	KeepAlive:        true,
	MaxIdleConns:     100,
	IdleConnTimeout:  90 * time.Second,
	CopyResponseBody: false,
	BufferSize:       32 * 1024,
	MaxBufferSize:    1024 * 1024,      // 1MB
	RetryCount:       0,                // 默认不重试，只有网络异常时才重试
	RetryTimeout:     30 * time.Second, // 重试超时时间

	// 头部配置默认值 - 前端默认为空，后端也保持为空以保持一致
	SetHeaders:         map[string]string{},
	PassHeaders:        []string{},
	HideHeaders:        []string{},
	HTTPVersion:        "1.1",
	PreserveHost:       false,
	AddXForwardedFor:   true, // 与前端保持一致
	AddXRealIP:         true, // 与前端保持一致
	AddXForwardedProto: true, // 与前端保持一致

	// 新增nginx风格配置默认值
	ProxyBuffering: true,

	// TLS配置默认值
	TLSInsecureSkipVerify: false, // 生产环境应该为false
	TLSMinVersion:         "1.2", // 默认最小TLS版本
	TLSMaxVersion:         "1.3", // 与前端保持一致
	TLSServerName:         "",    // 空表示使用目标主机名
}

var DefaultWebSocketConfig = WebSocketConfig{
	ID:                "default-websocket",
	Enabled:           true,
	PingInterval:      30 * time.Second,
	PongTimeout:       10 * time.Second,
	WriteTimeout:      10 * time.Second,
	ReadTimeout:       60 * time.Second,
	MaxMessageSize:    1024 * 1024, // 1MB
	ReadBufferSize:    4 * 1024,    // 4KB
	WriteBufferSize:   4 * 1024,    // 4KB
	EnableCompression: false,
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
