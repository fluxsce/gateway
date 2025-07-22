package filter

import (
	"gateway/internal/gateway/core"
)

// FilterType 过滤器类型
// 定义了不同的过滤器类型，用于处理请求的不同方面
type FilterType string

const (
	// HeaderFilterType 请求头过滤器
	// 用于修改请求头或响应头
	HeaderFilterType FilterType = "header"

	// QueryParamFilterType 查询参数过滤器
	// 用于修改URL查询参数
	QueryParamFilterType FilterType = "query-param"

	// BodyFilterType 请求体过滤器
	// 用于修改请求体内容
	BodyFilterType FilterType = "body"

	// URLFilterType URL过滤器（通用类型）
	// 用于URL路径相关的过滤操作
	URLFilterType FilterType = "url"

	// StripFilterType 前缀剥离过滤器
	// 用于剥离URL路径前缀
	StripFilterType FilterType = "strip"

	// RewriteFilterType 路径重写过滤器
	// 用于重写URL路径
	RewriteFilterType FilterType = "rewrite"

	// MethodFilterType HTTP方法过滤器
	// 用于修改请求方法
	MethodFilterType FilterType = "method"

	// CookieFilterType Cookie过滤器
	// 用于修改请求或响应Cookie
	CookieFilterType FilterType = "cookie"

	// ResponseFilterType 响应过滤器
	// 用于修改响应体内容
	ResponseFilterType FilterType = "response"
)

// FilterAction 过滤器执行时机
// 定义了过滤器执行的阶段
type FilterAction string

const (
	// PreRouting 路由前
	// 在路由匹配之前执行
	PreRouting FilterAction = "pre-routing"

	// PostRouting 路由后
	// 在路由匹配之后，请求转发之前执行
	PostRouting FilterAction = "post-routing"

	// PreResponse 响应前
	// 在接收到后端响应后，返回给客户端之前执行
	PreResponse FilterAction = "pre-response"
)

// FilterConfig 过滤器配置结构
type FilterConfig struct {
	ID      string                 `yaml:"id" json:"id" mapstructure:"id"`
	Name    string                 `yaml:"name" json:"name" mapstructure:"name"`
	Type    string                 `yaml:"type" json:"type" mapstructure:"type"` // 明确的过滤器类型
	Enabled bool                   `yaml:"enabled" json:"enabled" mapstructure:"enabled"`
	Order   int                    `yaml:"order" json:"order" mapstructure:"order"`
	Action  string                 `yaml:"action,omitempty" json:"action,omitempty" mapstructure:"action,omitempty"`
	Config  map[string]interface{} `yaml:"config" json:"config" mapstructure:"config"`
}

// Filter 过滤器接口
// 所有类型的过滤器都实现此接口
type Filter interface {
	// Apply 应用过滤器
	// 参数:
	// - ctx: 请求上下文
	// 返回值:
	// - error: 过滤过程中的错误
	Apply(ctx *core.Context) error

	// GetType 获取过滤器类型
	// 返回值:
	// - FilterType: 过滤器类型
	GetType() FilterType

	// GetAction 获取过滤器执行时机
	// 返回值:
	// - FilterAction: 过滤器执行时机
	GetAction() FilterAction

	// GetPriority 获取过滤器优先级
	// 返回值:
	// - int: 优先级，值越大优先级越高
	GetPriority() int

	// IsEnabled 是否启用
	// 返回值:
	// - bool: 是否启用
	IsEnabled() bool

	// GetName 获取过滤器名称
	// 返回值:
	// - string: 过滤器名称
	GetName() string

	// GetConfig 获取过滤器配置
	// 返回值:
	// - FilterConfig: 过滤器的配置信息
	GetConfig() FilterConfig
}

// BaseFilter 过滤器基础结构
// 包含所有过滤器共有的属性
type BaseFilter struct {
	// 过滤器类型
	Type FilterType

	// 过滤器执行时机
	Action FilterAction

	// 过滤器优先级
	Priority int

	// 是否启用
	Enabled bool

	// 过滤器名称
	Name string

	// 原始配置信息
	originalConfig FilterConfig
}

// GetType 获取过滤器类型
func (f *BaseFilter) GetType() FilterType {
	return f.Type
}

// GetAction 获取过滤器执行时机
func (f *BaseFilter) GetAction() FilterAction {
	return f.Action
}

// GetPriority 获取过滤器优先级
func (f *BaseFilter) GetPriority() int {
	return f.Priority
}

// IsEnabled 是否启用
func (f *BaseFilter) IsEnabled() bool {
	return f.Enabled
}

// GetName 获取过滤器名称
func (f *BaseFilter) GetName() string {
	return f.Name
}

// GetConfig 获取过滤器配置
func (f *BaseFilter) GetConfig() FilterConfig {
	return f.originalConfig
}

// Apply 实现Filter接口的Apply方法
// 这是一个默认实现，不执行任何操作
// 所有继承BaseFilter的具体过滤器应该重写此方法
func (f *BaseFilter) Apply(ctx *core.Context) error {
	// 基类不执行任何操作，总是成功
	return nil
}

// NewBaseFilter 创建基础过滤器
func NewBaseFilter(filterType FilterType, action FilterAction, priority int, enabled bool, name string) *BaseFilter {
	config := FilterConfig{
		ID:      name,
		Name:    name,
		Enabled: enabled,
		Order:   priority,
		Action:  string(action),
		Config:  make(map[string]interface{}),
	}

	return &BaseFilter{
		Type:           filterType,
		Action:         action,
		Priority:       priority,
		Enabled:        enabled,
		Name:           name,
		originalConfig: config,
	}
}
