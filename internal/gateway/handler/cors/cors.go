package cors

import (
	"gohub/internal/gateway/core"
)

// CORSStrategy CORS策略类型
type CORSStrategy string

const (
	// StrategyDefault 默认CORS策略
	StrategyDefault CORSStrategy = "default"
	// StrategyStrict 严格CORS策略
	StrategyStrict CORSStrategy = "strict"
	// StrategyPermissive 宽松CORS策略
	StrategyPermissive CORSStrategy = "permissive"
	// StrategyCustom 自定义CORS策略
	StrategyCustom CORSStrategy = "custom"
)

// CORSHandler CORS处理器接口
// 所有CORS处理器都必须实现此接口
type CORSHandler interface {
	// Handle 处理CORS
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - bool: 是否继续处理后续逻辑
	Handle(ctx *core.Context) bool

	// GetStrategy 获取CORS策略
	// 返回值:
	// - CORSStrategy: CORS策略类型
	GetStrategy() CORSStrategy

	// IsEnabled 是否启用
	// 返回值:
	// - bool: 是否启用
	IsEnabled() bool

	// GetName 获取CORS处理器名称
	// 返回值:
	// - string: 处理器名称
	GetName() string

	// Validate 验证配置
	// 返回值:
	// - error: 验证错误
	Validate() error

	// GetConfig 获取配置
	// 返回值:
	// - CORSConfig: CORS配置
	GetConfig() CORSConfig
}

// BaseCORSHandler CORS处理器基础结构
// 包含所有CORS处理器共有的属性
type BaseCORSHandler struct {
	// CORS策略类型
	Strategy CORSStrategy

	// 是否启用
	Enabled bool

	// 处理器名称
	Name string

	// 配置
	Config CORSConfig
}

// GetStrategy 获取CORS策略
func (c *BaseCORSHandler) GetStrategy() CORSStrategy {
	return c.Strategy
}

// IsEnabled 是否启用
func (c *BaseCORSHandler) IsEnabled() bool {
	return c.Enabled
}

// GetName 获取CORS处理器名称
func (c *BaseCORSHandler) GetName() string {
	return c.Name
}

// GetConfig 获取配置
func (c *BaseCORSHandler) GetConfig() CORSConfig {
	return c.Config
}

// Handle 实现CORSHandler接口的Handle方法
// 这是一个默认实现，总是返回true（继续处理）
// 所有继承BaseCORSHandler的具体处理器应该重写此方法
func (c *BaseCORSHandler) Handle(ctx *core.Context) bool {
	// 基类默认继续处理
	return true
}

// Validate 实现CORSHandler接口的Validate方法
// 这是一个默认实现，总是返回nil（验证通过）
// 所有继承BaseCORSHandler的具体处理器应该重写此方法
func (c *BaseCORSHandler) Validate() error {
	// 基类默认验证通过
	return nil
}

// CORSConfig CORS配置
type CORSConfig struct {
	// CORS配置ID
	ID string `yaml:"id" json:"id" mapstructure:"id"`
	// CORS配置名称
	Name string `yaml:"name" json:"name" mapstructure:"name"`
	// 是否启用CORS
	Enabled bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	// CORS策略
	Strategy CORSStrategy `yaml:"strategy,omitempty" json:"strategy,omitempty" mapstructure:"strategy,omitempty"`
	// 允许的域名列表，*表示允许所有域
	AllowOrigins []string `yaml:"allow_origins" json:"allow_origins" mapstructure:"allow_origins"`
	// 允许的HTTP方法列表
	AllowMethods []string `yaml:"allow_methods" json:"allow_methods" mapstructure:"allow_methods"`
	// 允许的HTTP头列表
	AllowHeaders []string `yaml:"allow_headers" json:"allow_headers" mapstructure:"allow_headers"`
	// 允许客户端访问的响应头列表
	ExposeHeaders []string `yaml:"expose_headers,omitempty" json:"expose_headers,omitempty" mapstructure:"expose_headers,omitempty"`
	// 是否允许携带凭证(Cookie等)
	AllowCredentials bool `yaml:"allow_credentials,omitempty" json:"allow_credentials,omitempty" mapstructure:"allow_credentials,omitempty"`
	// 预检请求结果缓存时间(秒)
	MaxAge int `yaml:"max_age,omitempty" json:"max_age,omitempty" mapstructure:"max_age,omitempty"`
	// 自定义配置参数
	CustomConfig map[string]interface{} `yaml:"custom_config,omitempty" json:"custom_config,omitempty" mapstructure:"custom_config,omitempty"`
}

// DefaultCORSConfig 默认CORS配置
var DefaultCORSConfig = CORSConfig{
	ID:               "default-cors",
	Name:             "Default CORS Configuration",
	Enabled:          true,
	Strategy:         StrategyDefault,
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	ExposeHeaders:    []string{},
	AllowCredentials: false,
	MaxAge:           86400, // 24小时
}

// StrictCORSConfig 严格CORS配置
var StrictCORSConfig = CORSConfig{
	ID:               "strict-cors",
	Name:             "Strict CORS Configuration",
	Enabled:          true,
	Strategy:         StrategyStrict,
	AllowOrigins:     []string{}, // 需要明确指定
	AllowMethods:     []string{"GET", "POST"},
	AllowHeaders:     []string{"Content-Type", "Authorization"},
	ExposeHeaders:    []string{},
	AllowCredentials: false,
	MaxAge:           3600, // 1小时
}

// PermissiveCORSConfig 宽松CORS配置
var PermissiveCORSConfig = CORSConfig{
	ID:               "permissive-cors",
	Name:             "Permissive CORS Configuration",
	Enabled:          true,
	Strategy:         StrategyPermissive,
	AllowOrigins:     []string{"*"},
	AllowMethods:     []string{"*"},
	AllowHeaders:     []string{"*"},
	ExposeHeaders:    []string{"*"},
	AllowCredentials: true,
	MaxAge:           86400,
}

// CORS 主CORS处理器
// 管理不同的CORS策略和处理器
type CORS struct {
	config  CORSConfig
	handler CORSHandler
}

// NewCORS 创建CORS处理器
func NewCORS(config *CORSConfig) *CORS {
	cors := &CORS{
		config: DefaultCORSConfig,
	}

	if config != nil {
		cors.config = *config
		
		// 创建默认处理器
		handler := &BaseCORSHandler{
			Strategy: config.Strategy,
			Enabled:  config.Enabled,
			Config:   *config,
		}
		
		// 设置处理器名称
		if config.Name != "" {
			handler.Name = config.Name
		} else if config.ID != "" {
			handler.Name = config.ID
		} else {
			handler.Name = "CORS Handler"
		}
		
		cors.handler = handler
	}

	return cors
}

// Handle 实现core.Handler接口
func (c *CORS) Handle(ctx *core.Context) bool {
	// 如果未启用CORS，继续处理
	if !c.config.Enabled {
		return true
	}

	// 如果没有配置具体的处理器，使用默认行为
	if c.handler == nil {
		return true
	}

	return c.handler.Handle(ctx)
}

// SetHandler 设置CORS处理器
func (c *CORS) SetHandler(handler CORSHandler) {
	if handler != nil && handler.IsEnabled() {
		c.handler = handler
	}
}

// GetHandler 获取CORS处理器
func (c *CORS) GetHandler() CORSHandler {
	return c.handler
}

// GetConfig 获取CORS配置
func (c *CORS) GetConfig() CORSConfig {
	return c.config
}

// SetConfig 设置CORS配置
func (c *CORS) SetConfig(config CORSConfig) {
	c.config = config
}
