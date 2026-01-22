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
// 全局API - 便捷函数
// =============================================================================
// 注意：渠道管理相关操作请直接使用 GetGlobalManager() 获取管理器实例

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
