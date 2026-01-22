// Package alert 提供统一的告警通知功能
// 支持多种告警渠道：邮件、QQ、企业微信等
package alert

import (
	"context"
	"time"
)

// AlertType 告警类型
type AlertType string

const (
	// AlertTypeEmail 邮件告警
	AlertTypeEmail AlertType = "email"
	// AlertTypeQQ QQ告警
	AlertTypeQQ AlertType = "qq"
	// AlertTypeWeChatWork 企业微信告警
	AlertTypeWeChatWork AlertType = "wechat_work"
	// AlertTypeDingTalk 钉钉告警
	AlertTypeDingTalk AlertType = "dingtalk"
	// AlertTypeWebhook Webhook告警
	AlertTypeWebhook AlertType = "webhook"
	// AlertTypeSMS 短信告警
	AlertTypeSMS AlertType = "sms"
)

// DisplayFormat 显示格式
type DisplayFormat string

const (
	// DisplayFormatTable 表格格式（默认）
	DisplayFormatTable DisplayFormat = "table"
	// DisplayFormatText 文本格式
	DisplayFormatText DisplayFormat = "text"
)

// Message 告警消息
type Message struct {
	// Title 消息标题
	Title string
	// Content 消息内容
	Content string
	// Timestamp 消息时间戳
	Timestamp time.Time
	// Tags 消息标签，用于分类和过滤
	Tags map[string]string
	// Extra 额外数据，用于特定渠道的扩展信息
	Extra map[string]interface{}
	// DisplayFormat 显示格式：table（表格，默认）或 text（文本）
	DisplayFormat DisplayFormat
	// TableData 表格数据，当 DisplayFormat 为 table 时使用
	// 是一个 map[string]interface{}，键为列名或行标识，值为对应的数据
	TableData map[string]interface{}
}

// NewMessage 创建新的告警消息
// 返回:
//
//	*Message: 新创建的告警消息实例
func NewMessage() *Message {
	return &Message{
		Timestamp:     time.Now(),
		Tags:          make(map[string]string),
		Extra:         make(map[string]interface{}),
		DisplayFormat: DisplayFormatTable, // 默认为表格格式
		TableData:     make(map[string]interface{}),
	}
}

// WithTitle 设置消息标题（链式调用）
// 参数:
//
//	title: 消息标题
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithTitle(title string) *Message {
	m.Title = title
	return m
}

// WithContent 设置消息内容（链式调用）
// 参数:
//
//	content: 消息内容
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithContent(content string) *Message {
	m.Content = content
	return m
}

// WithTag 添加标签（链式调用）
// 参数:
//
//	key: 标签键
//	value: 标签值
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithTag(key, value string) *Message {
	if m.Tags == nil {
		m.Tags = make(map[string]string)
	}
	m.Tags[key] = value
	return m
}

// WithTags 设置标签（链式调用）
// 参数:
//
//	tags: 标签映射
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithTags(tags map[string]string) *Message {
	m.Tags = tags
	return m
}

// WithExtra 添加额外数据（链式调用）
// 参数:
//
//	key: 额外数据的键
//	value: 额外数据的值
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithExtra(key string, value interface{}) *Message {
	if m.Extra == nil {
		m.Extra = make(map[string]interface{})
	}
	m.Extra[key] = value
	return m
}

// WithTimestamp 设置时间戳（链式调用）
// 参数:
//
//	timestamp: 时间戳
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithTimestamp(timestamp time.Time) *Message {
	m.Timestamp = timestamp
	return m
}

// WithDisplayFormat 设置显示格式（链式调用）
// 参数:
//
//	format: 显示格式，table（表格）或 text（文本）
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithDisplayFormat(format DisplayFormat) *Message {
	m.DisplayFormat = format
	return m
}

// WithTableData 设置表格数据（链式调用）
// 参数:
//
//	tableData: 表格数据，是一个 map[string]interface{}
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) WithTableData(tableData map[string]interface{}) *Message {
	m.TableData = tableData
	// 如果设置了表格数据，自动设置为表格格式
	if len(tableData) > 0 {
		m.DisplayFormat = DisplayFormatTable
	}
	return m
}

// AddTableField 添加表格字段（链式调用）
// 参数:
//
//	key: 字段键
//	value: 字段值
//
// 返回:
//
//	*Message: 消息实例，支持链式调用
func (m *Message) AddTableField(key string, value interface{}) *Message {
	if m.TableData == nil {
		m.TableData = make(map[string]interface{})
	}
	m.TableData[key] = value
	// 如果添加了表格数据，自动设置为表格格式
	m.DisplayFormat = DisplayFormatTable
	return m
}

// SendOptions 发送选项
type SendOptions struct {
	// Timeout 发送超时时间
	Timeout time.Duration
	// Retry 重试次数
	Retry int
	// RetryInterval 重试间隔
	RetryInterval time.Duration
	// Async 是否异步发送
	Async bool
}

// DefaultSendOptions 默认发送选项
func DefaultSendOptions() *SendOptions {
	return &SendOptions{
		Timeout:       30 * time.Second,
		Retry:         3,
		RetryInterval: 5 * time.Second,
		Async:         false,
	}
}

// SendResult 发送结果
type SendResult struct {
	// Success 是否成功
	Success bool
	// Error 错误信息
	Error error
	// MessageID 消息ID（如果渠道支持）
	MessageID string
	// Timestamp 发送时间
	Timestamp time.Time
	// Duration 发送耗时
	Duration time.Duration
	// Extra 额外信息
	Extra map[string]interface{}
}

// HealthCheckResult 健康检查结果
type HealthCheckResult struct {
	// Success 是否健康
	Success bool
	// Error 错误信息（如果检查失败）
	Error error
	// Timestamp 检查时间
	Timestamp time.Time
	// Duration 检查耗时
	Duration time.Duration
	// Message 检查消息
	Message string
	// Extra 额外信息
	Extra map[string]interface{}
}

// Channel 告警渠道接口
// 所有告警渠道都需要实现此接口
type Channel interface {
	// Send 发送告警消息
	// 参数:
	//   ctx: 上下文，用于超时控制和取消操作
	//   message: 告警消息
	//   options: 发送选项
	// 返回:
	//   *SendResult: 发送结果
	Send(ctx context.Context, message *Message, options *SendOptions) *SendResult

	// Type 返回渠道类型
	Type() AlertType

	// Name 返回渠道名称
	Name() string

	// IsEnabled 检查渠道是否启用
	IsEnabled() bool

	// Enable 启用渠道
	Enable() error

	// Disable 禁用渠道
	Disable() error

	// Close 关闭渠道，释放资源
	Close() error

	// HealthCheck 健康检查
	// 参数:
	//   ctx: 上下文，用于超时控制和取消操作
	// 返回:
	//   *HealthCheckResult: 健康检查结果
	HealthCheck(ctx context.Context) *HealthCheckResult
}
