package alert_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"gateway/pkg/alert"
	"gateway/pkg/alert/channel"
)

// =============================================================================
// EmailServerConfig Tests
// =============================================================================

func TestEmailServerConfig_Validate(t *testing.T) {
	t.Run("有效配置", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
			FromName: "系统告警",
			UseTLS:   true,
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("空SMTP主机", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		err := config.Validate()
		if err == nil {
			t.Error("空SMTP主机应该验证失败")
		}
		if !strings.Contains(err.Error(), "SMTP服务器地址不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("无效端口号 - 负数", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: -1,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		err := config.Validate()
		if err == nil {
			t.Error("无效端口号应该验证失败")
		}
		if !strings.Contains(err.Error(), "SMTP端口号无效") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("无效端口号 - 超出范围", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 70000,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		err := config.Validate()
		if err == nil {
			t.Error("超出范围的端口号应该验证失败")
		}
	})

	t.Run("空用户名", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "",
			Password: "password123",
			From:     "alert@example.com",
		}

		err := config.Validate()
		if err == nil {
			t.Error("空用户名应该验证失败")
		}
		if !strings.Contains(err.Error(), "用户名不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("空密码", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "",
			From:     "alert@example.com",
		}

		err := config.Validate()
		if err == nil {
			t.Error("空密码应该验证失败")
		}
		if !strings.Contains(err.Error(), "密码不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("空发件人地址", func(t *testing.T) {
		config := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "",
		}

		err := config.Validate()
		if err == nil {
			t.Error("空发件人地址应该验证失败")
		}
		if !strings.Contains(err.Error(), "发件人地址不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("各种端口号", func(t *testing.T) {
		validPorts := []int{25, 465, 587, 2525}
		for _, port := range validPorts {
			config := &channel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: port,
				Username: "user@example.com",
				Password: "password123",
				From:     "alert@example.com",
			}

			err := config.Validate()
			if err != nil {
				t.Errorf("端口 %d 应该是有效的，但验证失败: %v", port, err)
			}
		}
	})
}

// =============================================================================
// EmailSendConfig Tests
// =============================================================================

func TestEmailSendConfig_Validate(t *testing.T) {
	t.Run("有效配置 - 只有收件人", func(t *testing.T) {
		config := &channel.EmailSendConfig{
			To: []string{"user1@example.com", "user2@example.com"},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("有效配置 - 包含抄送和密送", func(t *testing.T) {
		config := &channel.EmailSendConfig{
			To:  []string{"user1@example.com"},
			CC:  []string{"cc@example.com"},
			BCC: []string{"bcc@example.com"},
		}

		err := config.Validate()
		if err != nil {
			t.Errorf("有效配置验证失败: %v", err)
		}
	})

	t.Run("空收件人列表", func(t *testing.T) {
		config := &channel.EmailSendConfig{
			To: []string{},
		}

		err := config.Validate()
		if err == nil {
			t.Error("空收件人列表应该验证失败")
		}
		if !strings.Contains(err.Error(), "收件人列表不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("nil收件人列表", func(t *testing.T) {
		config := &channel.EmailSendConfig{
			To: nil,
		}

		err := config.Validate()
		if err == nil {
			t.Error("nil收件人列表应该验证失败")
		}
	})
}

// =============================================================================
// EmailChannel Tests
// =============================================================================

func TestNewEmailChannel(t *testing.T) {
	t.Run("创建有效的邮件渠道", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
			FromName: "系统告警",
			UseTLS:   true,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		ch, err := channel.NewEmailChannel("test_email", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建邮件渠道失败: %v", err)
		}

		if ch == nil {
			t.Fatal("返回的渠道为 nil")
		}

		if ch.Name() != "test_email" {
			t.Errorf("渠道名称不正确，期望 'test_email'，实际 '%s'", ch.Name())
		}

		if ch.Type() != alert.AlertTypeEmail {
			t.Errorf("渠道类型不正确，期望 '%s'，实际 '%s'", alert.AlertTypeEmail, ch.Type())
		}

		if !ch.IsEnabled() {
			t.Error("新创建的渠道应该是启用状态")
		}
	})

	t.Run("nil服务器配置", func(t *testing.T) {
		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		_, err := channel.NewEmailChannel("test", nil, sendConfig)
		if err == nil {
			t.Error("nil服务器配置应该返回错误")
		}
		if !strings.Contains(err.Error(), "服务器配置不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("无效的服务器配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		_, err := channel.NewEmailChannel("test", serverConfig, sendConfig)
		if err == nil {
			t.Error("无效的服务器配置应该返回错误")
		}
		if !strings.Contains(err.Error(), "服务器配置验证失败") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("nil发送配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		_, err := channel.NewEmailChannel("test", serverConfig, nil)
		if err == nil {
			t.Error("nil发送配置应该返回错误")
		}
		if !strings.Contains(err.Error(), "发送配置不能为空") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})

	t.Run("无效的发送配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{},
		}

		_, err := channel.NewEmailChannel("test", serverConfig, sendConfig)
		if err == nil {
			t.Error("无效的发送配置应该返回错误")
		}
		if !strings.Contains(err.Error(), "发送配置验证失败") {
			t.Errorf("错误消息不正确: %v", err)
		}
	})
}

func TestEmailChannel_SetSendConfig(t *testing.T) {
	serverConfig := &channel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password123",
		From:     "alert@example.com",
	}

	sendConfig := &channel.EmailSendConfig{
		To: []string{"admin@example.com"},
	}

	ch, _ := channel.NewEmailChannel("test", serverConfig, sendConfig)

	t.Run("设置有效的发送配置", func(t *testing.T) {
		newConfig := &channel.EmailSendConfig{
			To: []string{"new@example.com", "admin@example.com"},
			CC: []string{"cc@example.com"},
		}

		err := ch.SetSendConfig(newConfig)
		if err != nil {
			t.Errorf("设置有效配置失败: %v", err)
		}
	})

	t.Run("设置nil配置", func(t *testing.T) {
		err := ch.SetSendConfig(nil)
		if err == nil {
			t.Error("设置nil配置应该返回错误")
		}
	})

	t.Run("设置无效配置", func(t *testing.T) {
		invalidConfig := &channel.EmailSendConfig{
			To: []string{},
		}

		err := ch.SetSendConfig(invalidConfig)
		if err == nil {
			t.Error("设置无效配置应该返回错误")
		}
	})
}

func TestEmailChannel_Send_Disabled(t *testing.T) {
	serverConfig := &channel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password123",
		From:     "alert@example.com",
	}

	sendConfig := &channel.EmailSendConfig{
		To: []string{"admin@example.com"},
	}

	ch, _ := channel.NewEmailChannel("test", serverConfig, sendConfig)

	// 禁用渠道
	_ = ch.Disable()

	ctx := context.Background()
	message := &alert.Message{
		Title:     "测试告警",
		Content:   "这是测试内容",
		Timestamp: time.Now(),
	}

	result := ch.Send(ctx, message, nil)

	if result.Success {
		t.Error("禁用的渠道发送应该失败")
	}

	if result.Error == nil {
		t.Error("应该返回错误信息")
	}

	if !strings.Contains(result.Error.Error(), "未启用") {
		t.Errorf("错误消息不正确: %v", result.Error)
	}
}

func TestEmailChannel_Close(t *testing.T) {
	serverConfig := &channel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password123",
		From:     "alert@example.com",
	}

	sendConfig := &channel.EmailSendConfig{
		To: []string{"admin@example.com"},
	}

	ch, _ := channel.NewEmailChannel("test", serverConfig, sendConfig)

	err := ch.Close()
	if err != nil {
		t.Errorf("关闭渠道失败: %v", err)
	}
}

func TestEmailChannel_HealthCheck(t *testing.T) {
	t.Run("健康检查接口测试", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
			Timeout:  1,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		ch, err := channel.NewEmailChannel("test", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建渠道失败: %v", err)
		}

		// 测试健康检查方法存在且可调用
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		// 调用健康检查（预期会失败，因为是假的服务器地址）
		err = ch.HealthCheck(ctx)
		// 不验证结果，只测试接口可用性
		t.Logf("健康检查执行完成，结果: %v", err)
	})
}

func TestEmailChannel_Stats(t *testing.T) {
	serverConfig := &channel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password123",
		From:     "alert@example.com",
	}

	sendConfig := &channel.EmailSendConfig{
		To: []string{"admin@example.com"},
	}

	ch, _ := channel.NewEmailChannel("test", serverConfig, sendConfig)

	stats := ch.Stats()

	if stats == nil {
		t.Fatal("Stats() 返回 nil")
	}

	// 验证基本统计字段
	expectedFields := []string{"name", "type", "enabled", "total_sent", "total_success", "total_failed"}
	for _, field := range expectedFields {
		if _, ok := stats[field]; !ok {
			t.Errorf("统计信息缺少字段: %s", field)
		}
	}

	if stats["name"] != "test" {
		t.Errorf("名称不正确，期望 'test'，实际 '%v'", stats["name"])
	}

	if stats["type"] != alert.AlertTypeEmail {
		t.Errorf("类型不正确，期望 '%s'，实际 '%v'", alert.AlertTypeEmail, stats["type"])
	}

	if stats["enabled"] != true {
		t.Error("应该是启用状态")
	}
}

// =============================================================================
// 消息格式测试
// =============================================================================

func TestEmailChannel_MessageFormat(t *testing.T) {
	t.Run("基本消息", func(t *testing.T) {
		message := &alert.Message{
			Title:     "测试告警",
			Content:   "这是测试内容",
			Timestamp: time.Now(),
		}

		if message.Title != "测试告警" {
			t.Error("标题设置不正确")
		}
		if message.Content != "这是测试内容" {
			t.Error("内容设置不正确")
		}
	})

	t.Run("带标签的消息", func(t *testing.T) {
		message := &alert.Message{
			Title:     "测试告警",
			Content:   "这是测试内容",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"severity": "high",
				"service":  "api",
			},
		}

		if len(message.Tags) != 2 {
			t.Errorf("标签数量不正确，期望 2，实际 %d", len(message.Tags))
		}
	})

	t.Run("带自定义发送配置的消息", func(t *testing.T) {
		customSendConfig := &channel.EmailSendConfig{
			To:  []string{"emergency@example.com"},
			CC:  []string{"manager@example.com"},
			BCC: []string{"archive@example.com"},
		}

		message := &alert.Message{
			Title:     "紧急告警",
			Content:   "数据库连接失败",
			Timestamp: time.Now(),
			Extra: map[string]interface{}{
				"send_config": customSendConfig,
			},
		}

		// 验证自定义配置已设置
		if config, ok := message.Extra["send_config"].(*channel.EmailSendConfig); ok {
			if len(config.To) != 1 || config.To[0] != "emergency@example.com" {
				t.Error("自定义收件人配置不正确")
			}
			if len(config.CC) != 1 {
				t.Error("抄送配置不正确")
			}
			if len(config.BCC) != 1 {
				t.Error("密送配置不正确")
			}
		} else {
			t.Error("无法获取自定义发送配置")
		}
	})
}

// =============================================================================
// 配置选项测试
// =============================================================================

func TestEmailChannel_Options(t *testing.T) {
	t.Run("TLS配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost:   "smtp.example.com",
			SMTPPort:   465,
			Username:   "user@example.com",
			Password:   "password123",
			From:       "alert@example.com",
			UseTLS:     true,
			SkipVerify: true,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		ch, err := channel.NewEmailChannel("test_tls", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建TLS渠道失败: %v", err)
		}

		if ch == nil {
			t.Fatal("渠道为 nil")
		}
	})

	t.Run("超时配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
			Timeout:  30,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		ch, err := channel.NewEmailChannel("test_timeout", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建带超时配置的渠道失败: %v", err)
		}

		if ch == nil {
			t.Fatal("渠道为 nil")
		}
	})

	t.Run("发件人名称配置", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.example.com",
			SMTPPort: 587,
			Username: "user@example.com",
			Password: "password123",
			From:     "alert@example.com",
			FromName: "系统监控告警",
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"admin@example.com"},
		}

		ch, err := channel.NewEmailChannel("test_from_name", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建带发件人名称的渠道失败: %v", err)
		}

		if ch == nil {
			t.Fatal("渠道为 nil")
		}
	})
}

// =============================================================================
// 集成测试（使用真实SMTP服务器）
// =============================================================================

func TestEmailChannel_Integration(t *testing.T) {
	t.Run("实际发送邮件到QQ邮箱", func(t *testing.T) {
		// 使用腾讯企业邮箱SMTP服务器
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.exmail.qq.com",
			SMTPPort: 465,
			Username: "shangjian@flux.com.cn",
			Password: "datahub2019FLUX",
			From:     "shangjian@flux.com.cn",
			FromName: "Gateway告警测试系统",
			UseTLS:   true,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"1751255104@qq.com"},
		}

		ch, err := channel.NewEmailChannel("test_real_qq", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建邮件渠道失败: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		message := &alert.Message{
			Title:     "Gateway告警系统集成测试",
			Content:   "这是一条来自Gateway告警系统的真实测试消息。\n\n测试时间: " + time.Now().Format("2006-01-02 15:04:05"),
			Timestamp: time.Now(),
			Tags: map[string]string{
				"env":      "test",
				"severity": "info",
				"service":  "gateway",
				"module":   "alert",
			},
		}

		result := ch.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送邮件失败: %v", result.Error)
		} else {
			t.Logf("发送邮件成功！耗时: %v", result.Duration)
		}
	})

	t.Run("发送带多个标签的告警邮件", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.exmail.qq.com",
			SMTPPort: 465,
			Username: "shangjian@flux.com.cn",
			Password: "datahub2019FLUX",
			From:     "shangjian@flux.com.cn",
			FromName: "Gateway监控告警",
			UseTLS:   true,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"1751255104@qq.com"},
		}

		ch, err := channel.NewEmailChannel("test_tags", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建邮件渠道失败: %v", err)
		}

		ctx := context.Background()
		message := &alert.Message{
			Title:     "系统性能告警",
			Content:   "检测到以下异常:\n\n- CPU使用率: 92%\n- 内存使用率: 87%\n- 磁盘IO: 高负载\n\n请及时处理！",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"severity":   "high",
				"service":    "gateway",
				"module":     "monitor",
				"alert_type": "performance",
				"host":       "server-01",
			},
		}

		result := ch.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送邮件失败: %v", result.Error)
		} else {
			t.Logf("发送带标签的邮件成功！耗时: %v", result.Duration)
		}
	})

	t.Run("测试健康检查", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.exmail.qq.com",
			SMTPPort: 465,
			Username: "shangjian@flux.com.cn",
			Password: "datahub2019FLUX",
			From:     "shangjian@flux.com.cn",
			UseTLS:   true,
			Timeout:  10,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"1751255104@qq.com"},
		}

		ch, err := channel.NewEmailChannel("test_health", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建邮件渠道失败: %v", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		err = ch.HealthCheck(ctx)
		if err != nil {
			t.Errorf("健康检查失败: %v", err)
		} else {
			t.Log("健康检查成功！SMTP服务器连接正常")
		}
	})

	t.Run("测试使用MessageBuilder发送", func(t *testing.T) {
		serverConfig := &channel.EmailServerConfig{
			SMTPHost: "smtp.exmail.qq.com",
			SMTPPort: 465,
			Username: "shangjian@flux.com.cn",
			Password: "datahub2019FLUX",
			From:     "shangjian@flux.com.cn",
			FromName: "Gateway告警",
			UseTLS:   true,
		}

		sendConfig := &channel.EmailSendConfig{
			To: []string{"1751255104@qq.com"},
		}

		ch, err := channel.NewEmailChannel("test_builder", serverConfig, sendConfig)
		if err != nil {
			t.Fatalf("创建邮件渠道失败: %v", err)
		}

		ctx := context.Background()

		// 使用MessageBuilder构建消息，然后通过渠道发送
		message := alert.NewMessage().
			WithTitle("使用MessageBuilder的测试").
			WithContent("这是使用MessageBuilder链式调用构建的消息。\n\n功能测试正常！").
			WithTag("method", "builder").
			WithTag("version", "v1.0").
			WithExtra("test_id", "builder_001").
			Build()

		result := ch.Send(ctx, message, nil)
		if !result.Success {
			t.Errorf("使用MessageBuilder发送失败: %v", result.Error)
		} else {
			t.Logf("✓ 使用MessageBuilder发送成功！耗时: %v", result.Duration)
		}
	})
}
