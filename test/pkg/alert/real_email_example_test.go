package alert_test

import (
	"context"
	"testing"
	"time"

	"gateway/pkg/alert"
	"gateway/pkg/alert/channel"
)

// TestRealEmailExample 真实邮件发送示例
// 这是一个完整的端到端测试示例，展示如何使用真实的SMTP配置发送邮件
func TestRealEmailExample(t *testing.T) {
	// 创建SMTP服务器配置
	serverConfig := &channel.EmailServerConfig{
		SMTPHost: "smtp.exmail.qq.com",
		SMTPPort: 465,
		Username: "shangjian@flux.com.cn",
		Password: "datahub2019FLUX",
		From:     "shangjian@flux.com.cn",
		FromName: "Gateway告警系统",
		UseTLS:   true,
	}

	// 创建默认发送配置
	sendConfig := &channel.EmailSendConfig{
		To: []string{"1751255104@qq.com", "luoqn@flux.com.cn"},
	}

	// 创建邮件渠道
	emailChannel, err := channel.NewEmailChannel("email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("创建邮件渠道失败: %v", err)
	}

	// 方式1: 直接使用渠道发送
	t.Run("直接发送简单消息", func(t *testing.T) {
		ctx := context.Background()
		message := &alert.Message{
			Title:     "测试邮件 - 直接发送",
			Content:   "这是通过EmailChannel直接发送的测试邮件。",
			Timestamp: time.Now(),
		}

		result := emailChannel.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 发送成功！耗时: %v", result.Duration)
		}
	})

	// 方式2: 使用Manager管理渠道
	t.Run("通过Manager发送", func(t *testing.T) {
		manager := alert.NewManager()
		err := manager.AddChannel("email", emailChannel)
		if err != nil {
			t.Fatalf("添加渠道失败: %v", err)
		}

		ctx := context.Background()
		message := &alert.Message{
			Title:     "测试邮件 - Manager发送",
			Content:   "这是通过Manager发送的测试邮件。",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"method": "manager",
				"type":   "test",
			},
		}

		result := manager.SendToDefault(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 发送成功！耗时: %v", result.Duration)
		}
	})

	// 方式3: 使用全局API
	t.Run("通过全局API发送", func(t *testing.T) {
		// 清理现有渠道
		for _, name := range alert.ListChannels() {
			_ = alert.RemoveChannel(name)
		}

		// 添加渠道到全局管理器
		err := alert.AddChannel("email", emailChannel)
		if err != nil {
			t.Fatalf("添加全局渠道失败: %v", err)
		}

		// 设置为默认渠道
		err = alert.SetDefaultChannel("email")
		if err != nil {
			t.Fatalf("设置默认渠道失败: %v", err)
		}

		ctx := context.Background()

		// 使用简单发送
		result := alert.SendSimple(ctx, "测试邮件 - 全局API", "这是通过全局API发送的测试邮件。")
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 发送成功！耗时: %v", result.Duration)
		}
	})

	// 方式4: 使用MessageBuilder
	t.Run("使用MessageBuilder构建和发送", func(t *testing.T) {
		ctx := context.Background()

		// 方式4a: 使用MessageBuilder构建，然后通过渠道发送
		message := alert.NewMessage().
			WithTitle("测试邮件 - MessageBuilder").
			WithContent("这是使用MessageBuilder链式调用构建的邮件。\n\n支持多行文本和格式化内容。").
			WithTag("builder", "true").
			WithTag("priority", "normal").
			WithTag("environment", "test").
			WithExtra("timestamp", time.Now().Unix()).
			Build()

		result := emailChannel.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 使用MessageBuilder发送成功！耗时: %v", result.Duration)
		}
	})

	// 方式5: 自定义发送配置（覆盖默认收件人）
	t.Run("使用自定义发送配置", func(t *testing.T) {
		ctx := context.Background()

		// 在消息中设置自定义发送配置
		customSendConfig := &channel.EmailSendConfig{
			To:  []string{"1751255104@qq.com", "luoqn@flux.com.cn"},
			CC:  []string{"luoqn@flux.com.cn"}, // 可以添加抄送
			BCC: []string{"luoqn@flux.com.cn"}, // 可以添加密送
		}

		message := &alert.Message{
			Title:     "测试邮件 - 自定义配置",
			Content:   "这是使用自定义发送配置的测试邮件。",
			Timestamp: time.Now(),
			Extra: map[string]interface{}{
				"send_config": customSendConfig,
			},
		}

		result := emailChannel.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 使用自定义配置发送成功！耗时: %v", result.Duration)
		}
	})

	// 方式6: 带标签的告警消息
	t.Run("发送带标签的告警消息", func(t *testing.T) {
		ctx := context.Background()

		message := &alert.Message{
			Title: "系统告警 - 性能监控",
			Content: "检测到系统异常:\n\n" +
				"- CPU使用率: 85%\n" +
				"- 内存使用率: 78%\n" +
				"- 磁盘IO: 正常\n" +
				"- 网络流量: 正常\n\n" +
				"建议检查CPU密集型进程。",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"severity":   "warning",
				"service":    "gateway",
				"module":     "monitor",
				"alert_type": "performance",
				"host":       "prod-server-01",
				"region":     "cn-beijing",
			},
		}

		result := emailChannel.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 发送告警邮件成功！耗时: %v", result.Duration)
		}
	})

	// 方式7: 测试重试机制
	t.Run("测试发送选项和重试", func(t *testing.T) {
		ctx := context.Background()

		// 自定义发送选项
		options := &alert.SendOptions{
			Timeout:       20 * time.Second,
			Retry:         2,
			RetryInterval: 3 * time.Second,
			Async:         false,
		}

		message := &alert.Message{
			Title:     "测试邮件 - 重试机制",
			Content:   "这是测试发送选项和重试机制的邮件。",
			Timestamp: time.Now(),
		}

		result := emailChannel.Send(ctx, message, options)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 发送成功（使用自定义选项）！耗时: %v", result.Duration)
		}
	})
}
