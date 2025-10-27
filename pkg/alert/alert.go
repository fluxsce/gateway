// Package alert 提供统一的告警通知功能
// 支持多种告警渠道：邮件、QQ、企业微信等
//
// 基本使用方法（服务器配置和发送配置分离）：
//
//	// 1. 创建SMTP服务器配置（可公用）
//	smtpServer := &channel.EmailServerConfig{
//	    SMTPHost: "smtp.example.com",
//	    SMTPPort: 587,
//	    Username: "alert@example.com",
//	    Password: "password",
//	    UseTLS:   true,
//	}
//
//	// 2. 创建默认发送配置
//	defaultSend := &channel.EmailSendConfig{
//	    From:     "alert@example.com",
//	    FromName: "系统告警",
//	    To:       []string{"admin@example.com"},
//	}
//
//	// 3. 创建邮件渠道
//	emailChannel, err := channel.NewEmailChannel("email", smtpServer, defaultSend)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// 4. 添加到管理器
//	alert.AddChannel("email", emailChannel)
//	alert.SetDefaultChannel("email")
//
//	// 5. 发送告警（使用默认配置）
//	alert.SendSimple(ctx, "系统告警", "CPU使用率超过90%")
//
//	// 6. 发送到不同收件人（覆盖发送配置）
//	message := &alert.Message{
//	    Title:   "严重告警",
//	    Content: "数据库连接失败",
//	    Extra: map[string]interface{}{
//	        "send_config": &channel.EmailSendConfig{
//	            From: "alert@example.com",
//	            To:   []string{"emergency@example.com"},
//	        },
//	    },
//	}
//	alert.SendToDefault(ctx, message, nil)
package alert

import (
	"context"
	"time"
)

// =============================================================================
// 全局API - 渠道管理
// =============================================================================

// AddChannel 向全局管理器添加渠道
// 这是一个便捷函数，直接向全局管理器添加渠道
// 参数:
//
//	name: 渠道名称
//	channel: 渠道实例
//
// 返回:
//
//	error: 添加失败时返回错误信息
func AddChannel(name string, channel Channel) error {
	return GetGlobalManager().AddChannel(name, channel)
}

// GetChannel 从全局管理器获取渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	Channel: 渠道实例，如果不存在则返回nil
func GetChannel(name string) Channel {
	return GetGlobalManager().GetChannel(name)
}

// RemoveChannel 从全局管理器移除渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 移除失败时返回错误信息
func RemoveChannel(name string) error {
	return GetGlobalManager().RemoveChannel(name)
}

// ListChannels 列出全局管理器中的所有渠道
// 返回:
//
//	[]string: 渠道名称列表
func ListChannels() []string {
	return GetGlobalManager().ListChannels()
}

// SetDefaultChannel 设置全局管理器的默认渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 设置失败时返回错误信息
func SetDefaultChannel(name string) error {
	return GetGlobalManager().SetDefaultChannel(name)
}

// GetDefaultChannel 获取全局管理器的默认渠道
// 返回:
//
//	Channel: 默认渠道实例，如果不存在则返回nil
func GetDefaultChannel() Channel {
	return GetGlobalManager().GetDefaultChannel()
}

// HasChannel 检查全局管理器中是否存在指定渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	bool: 如果存在返回true，否则返回false
func HasChannel(name string) bool {
	return GetGlobalManager().HasChannel(name)
}

// EnableChannel 启用指定渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 启用失败时返回错误信息
func EnableChannel(name string) error {
	return GetGlobalManager().EnableChannel(name)
}

// DisableChannel 禁用指定渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 禁用失败时返回错误信息
func DisableChannel(name string) error {
	return GetGlobalManager().DisableChannel(name)
}

// CloseAll 关闭全局管理器中的所有渠道
// 这是一个便捷函数，用于应用关闭时清理所有渠道
// 返回:
//
//	error: 关闭过程中的错误信息
func CloseAll() error {
	return GetGlobalManager().CloseAll()
}

// =============================================================================
// 全局API - 消息发送
// =============================================================================

// Send 通过指定渠道发送告警
// 参数:
//
//	ctx: 上下文，用于超时控制和取消操作
//	channelName: 渠道名称
//	message: 告警消息
//	options: 发送选项，可以为nil使用默认选项
//
// 返回:
//
//	*SendResult: 发送结果
func Send(ctx context.Context, channelName string, message *Message, options *SendOptions) *SendResult {
	return GetGlobalManager().Send(ctx, channelName, message, options)
}

// SendToDefault 通过默认渠道发送告警
// 参数:
//
//	ctx: 上下文，用于超时控制和取消操作
//	message: 告警消息
//	options: 发送选项，可以为nil使用默认选项
//
// 返回:
//
//	*SendResult: 发送结果
func SendToDefault(ctx context.Context, message *Message, options *SendOptions) *SendResult {
	return GetGlobalManager().SendToDefault(ctx, message, options)
}

// SendToAll 通过所有启用的渠道发送告警
// 参数:
//
//	ctx: 上下文，用于超时控制和取消操作
//	message: 告警消息
//	options: 发送选项，可以为nil使用默认选项
//
// 返回:
//
//	map[string]*SendResult: 各渠道的发送结果
func SendToAll(ctx context.Context, message *Message, options *SendOptions) map[string]*SendResult {
	return GetGlobalManager().SendToAll(ctx, message, options)
}

// SendToMultiple 通过多个指定渠道发送告警
// 参数:
//
//	ctx: 上下文，用于超时控制和取消操作
//	channelNames: 渠道名称列表
//	message: 告警消息
//	options: 发送选项，可以为nil使用默认选项
//
// 返回:
//
//	map[string]*SendResult: 各渠道的发送结果
func SendToMultiple(ctx context.Context, channelNames []string, message *Message, options *SendOptions) map[string]*SendResult {
	return GetGlobalManager().SendToMultiple(ctx, channelNames, message, options)
}

// =============================================================================
// 全局API - 统计和健康检查
// =============================================================================

// Stats 获取所有渠道的统计信息
// 返回:
//
//	map[string]map[string]interface{}: 各渠道的统计信息
func Stats() map[string]map[string]interface{} {
	return GetGlobalManager().Stats()
}

// HealthCheck 对所有渠道进行健康检查
// 参数:
//
//	ctx: 上下文，用于超时控制
//
// 返回:
//
//	map[string]error: 各渠道的健康检查结果
func HealthCheck(ctx context.Context) map[string]error {
	return GetGlobalManager().HealthCheck(ctx)
}

// =============================================================================
// 便捷函数 - 快速发送告警
// =============================================================================

// SendSimple 发送简单告警到默认渠道
// 参数:
//
//	ctx: 上下文
//	title: 标题
//	content: 内容
//
// 返回:
//
//	*SendResult: 发送结果
func SendSimple(ctx context.Context, title, content string) *SendResult {
	return SendToDefault(ctx, &Message{
		Title:     title,
		Content:   content,
		Timestamp: time.Now(),
	}, nil)
}

// SendWithTags 发送带标签的告警到默认渠道
// 参数:
//
//	ctx: 上下文
//	title: 标题
//	content: 内容
//	tags: 标签
//
// 返回:
//
//	*SendResult: 发送结果
func SendWithTags(ctx context.Context, title, content string, tags map[string]string) *SendResult {
	return SendToDefault(ctx, &Message{
		Title:     title,
		Content:   content,
		Timestamp: time.Now(),
		Tags:      tags,
	}, nil)
}

// =============================================================================
// 便捷函数 - 消息构建器
// =============================================================================

// MessageBuilder 消息构建器
type MessageBuilder struct {
	message *Message
}

// NewMessage 创建新的消息构建器
// 返回:
//
//	*MessageBuilder: 消息构建器实例
func NewMessage() *MessageBuilder {
	return &MessageBuilder{
		message: &Message{
			Timestamp: time.Now(),
			Tags:      make(map[string]string),
			Extra:     make(map[string]interface{}),
		},
	}
}

// WithTitle 设置标题
func (b *MessageBuilder) WithTitle(title string) *MessageBuilder {
	b.message.Title = title
	return b
}

// WithContent 设置内容
func (b *MessageBuilder) WithContent(content string) *MessageBuilder {
	b.message.Content = content
	return b
}

// WithTag 添加标签
func (b *MessageBuilder) WithTag(key, value string) *MessageBuilder {
	if b.message.Tags == nil {
		b.message.Tags = make(map[string]string)
	}
	b.message.Tags[key] = value
	return b
}

// WithTags 设置标签
func (b *MessageBuilder) WithTags(tags map[string]string) *MessageBuilder {
	b.message.Tags = tags
	return b
}

// WithExtra 添加额外数据
func (b *MessageBuilder) WithExtra(key string, value interface{}) *MessageBuilder {
	if b.message.Extra == nil {
		b.message.Extra = make(map[string]interface{})
	}
	b.message.Extra[key] = value
	return b
}

// Build 构建消息
func (b *MessageBuilder) Build() *Message {
	return b.message
}

// SendToDefault 发送到默认渠道
func (b *MessageBuilder) SendToDefault(ctx context.Context, options *SendOptions) *SendResult {
	return SendToDefault(ctx, b.message, options)
}

// SendTo 发送到指定渠道
func (b *MessageBuilder) SendTo(ctx context.Context, channelName string, options *SendOptions) *SendResult {
	return Send(ctx, channelName, b.message, options)
}

// SendToAll 发送到所有渠道
func (b *MessageBuilder) SendToAll(ctx context.Context, options *SendOptions) map[string]*SendResult {
	return SendToAll(ctx, b.message, options)
}
