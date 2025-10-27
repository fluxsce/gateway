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
	HealthCheck(ctx context.Context) error

	// Stats 获取统计信息
	Stats() map[string]interface{}
}

// Statistics 统计信息
type Statistics struct {
	// TotalSent 总发送数
	TotalSent int64
	// TotalSuccess 总成功数
	TotalSuccess int64
	// TotalFailed 总失败数
	TotalFailed int64
	// LastSendTime 最后发送时间
	LastSendTime time.Time
	// LastError 最后错误
	LastError error
	// AverageDuration 平均发送耗时
	AverageDuration time.Duration
}

// BaseChannel 基础渠道实现
// 提供通用功能，可被具体渠道嵌入
type BaseChannel struct {
	// name 渠道名称
	name string
	// channelType 渠道类型
	channelType AlertType
	// enabled 是否启用
	enabled bool
	// stats 统计信息
	stats *Statistics
}

// NewBaseChannel 创建基础渠道
func NewBaseChannel(name string, channelType AlertType) *BaseChannel {
	return &BaseChannel{
		name:        name,
		channelType: channelType,
		enabled:     true,
		stats: &Statistics{
			TotalSent:    0,
			TotalSuccess: 0,
			TotalFailed:  0,
		},
	}
}

// Type 返回渠道类型
func (b *BaseChannel) Type() AlertType {
	return b.channelType
}

// Name 返回渠道名称
func (b *BaseChannel) Name() string {
	return b.name
}

// IsEnabled 检查渠道是否启用
func (b *BaseChannel) IsEnabled() bool {
	return b.enabled
}

// Enable 启用渠道
func (b *BaseChannel) Enable() error {
	b.enabled = true
	return nil
}

// Disable 禁用渠道
func (b *BaseChannel) Disable() error {
	b.enabled = false
	return nil
}

// Stats 获取统计信息
func (b *BaseChannel) Stats() map[string]interface{} {
	return map[string]interface{}{
		"name":             b.name,
		"type":             b.channelType,
		"enabled":          b.enabled,
		"total_sent":       b.stats.TotalSent,
		"total_success":    b.stats.TotalSuccess,
		"total_failed":     b.stats.TotalFailed,
		"last_send_time":   b.stats.LastSendTime,
		"average_duration": b.stats.AverageDuration,
	}
}

// UpdateStats 更新统计信息
func (b *BaseChannel) UpdateStats(result *SendResult) {
	b.stats.TotalSent++
	b.stats.LastSendTime = result.Timestamp

	if result.Success {
		b.stats.TotalSuccess++
	} else {
		b.stats.TotalFailed++
		b.stats.LastError = result.Error
	}

	// 更新平均耗时
	if b.stats.TotalSent > 0 {
		totalDuration := b.stats.AverageDuration*time.Duration(b.stats.TotalSent-1) + result.Duration
		b.stats.AverageDuration = totalDuration / time.Duration(b.stats.TotalSent)
	}
}

// Close 关闭渠道的默认实现
func (b *BaseChannel) Close() error {
	return nil
}

// HealthCheck 健康检查的默认实现
func (b *BaseChannel) HealthCheck(ctx context.Context) error {
	return nil
}
