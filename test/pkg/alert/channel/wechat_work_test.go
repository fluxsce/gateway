package channel

import (
	"context"
	"strings"
	"testing"
	"time"

	"gateway/pkg/alert"
	alertchannel "gateway/pkg/alert/channel"
	"gateway/pkg/httpclient"
)

// 测试用的企业微信 Webhook URL
const testWeChatWorkWebhookURL = "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=aaaaaaaaaa"

// TestWeChatWorkServerConfig_Validate 测试企业微信服务器配置验证
func TestWeChatWorkServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *alertchannel.WeChatWorkServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置-markdown",
			config: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				MessageType: "markdown",
			},
			wantErr: false,
		},
		{
			name: "有效配置-text",
			config: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				MessageType: "text",
			},
			wantErr: false,
		},
		{
			name: "有效配置-默认markdown",
			config: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				MessageType: "",
			},
			wantErr: false,
		},
		{
			name: "Webhook地址为空",
			config: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "",
				MessageType: "markdown",
			},
			wantErr: true,
			errMsg:  "企业微信Webhook地址不能为空",
		},
		{
			name: "消息类型无效",
			config: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
				MessageType: "invalid",
			},
			wantErr: true,
			errMsg:  "消息类型必须是text或markdown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

// TestWeChatWorkSendConfig_Validate 测试企业微信发送配置验证
func TestWeChatWorkSendConfig_Validate(t *testing.T) {
	config := &alertchannel.WeChatWorkSendConfig{
		MentionedList:       []string{"user1", "user2"},
		MentionedMobileList: []string{"13800138000"},
	}

	err := config.Validate()
	if err != nil {
		t.Errorf("Validate() should not return error for valid config, got: %v", err)
	}
}

// TestNewWeChatWorkChannel 测试创建企业微信告警渠道
func TestNewWeChatWorkChannel(t *testing.T) {
	validServerConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	validSendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList: []string{"user1"},
	}

	tests := []struct {
		name         string
		channelName  string
		serverConfig *alertchannel.WeChatWorkServerConfig
		sendConfig   *alertchannel.WeChatWorkSendConfig
		httpClient   httpclient.Client
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "成功创建渠道",
			channelName:  "test-wechat-work",
			serverConfig: validServerConfig,
			sendConfig:   validSendConfig,
			httpClient:   nil, // 使用默认客户端
			wantErr:      false,
		},
		{
			name:         "服务器配置为nil",
			channelName:  "test-wechat-work",
			serverConfig: nil,
			sendConfig:   validSendConfig,
			wantErr:      true,
			errMsg:       "服务器配置不能为空",
		},
		{
			name: "服务器配置验证失败",
			serverConfig: &alertchannel.WeChatWorkServerConfig{
				WebhookURL:  "",
				MessageType: "markdown",
			},
			sendConfig: validSendConfig,
			wantErr:    true,
			errMsg:     "服务器配置验证失败",
		},
		{
			name:         "发送配置为nil-使用默认",
			serverConfig: validServerConfig,
			sendConfig:   nil,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := alertchannel.NewWeChatWorkChannel(tt.channelName, tt.serverConfig, tt.sendConfig, tt.httpClient)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWeChatWorkChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewWeChatWorkChannel() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if channel == nil {
					t.Error("NewWeChatWorkChannel() returned nil channel")
					return
				}
				if channel.Name() != tt.channelName {
					t.Errorf("Name() = %v, want %v", channel.Name(), tt.channelName)
				}
				if channel.Type() != alert.AlertTypeWeChatWork {
					t.Errorf("Type() = %v, want %v", channel.Type(), alert.AlertTypeWeChatWork)
				}
				if !channel.IsEnabled() {
					t.Error("New channel should be enabled by default")
				}
			}
		})
	}
}

// TestWeChatWorkChannel_BasicMethods 测试渠道基本方法
func TestWeChatWorkChannel_BasicMethods(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList: []string{"user1"},
	}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 测试 Type
	if channel.Type() != alert.AlertTypeWeChatWork {
		t.Errorf("Type() = %v, want %v", channel.Type(), alert.AlertTypeWeChatWork)
	}

	// 测试 Name
	if channel.Name() != "test-wechat-work" {
		t.Errorf("Name() = %v, want %v", channel.Name(), "test-wechat-work")
	}

	// 测试 IsEnabled (默认应该启用)
	if !channel.IsEnabled() {
		t.Error("IsEnabled() = false, want true")
	}

	// 测试 Disable
	if err := channel.Disable(); err != nil {
		t.Errorf("Disable() error = %v", err)
	}
	if channel.IsEnabled() {
		t.Error("IsEnabled() = true after Disable(), want false")
	}

	// 测试 Enable
	if err := channel.Enable(); err != nil {
		t.Errorf("Enable() error = %v", err)
	}
	if !channel.IsEnabled() {
		t.Error("IsEnabled() = false after Enable(), want true")
	}

	// 测试 Close
	if err := channel.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

// TestWeChatWorkChannel_SetSendConfig 测试设置发送配置
func TestWeChatWorkChannel_SetSendConfig(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	initialSendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList: []string{"initial"},
	}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, initialSendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 测试设置有效配置
	newSendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList:       []string{"new"},
		MentionedMobileList: []string{"13800138000"},
	}
	if err := channel.SetSendConfig(newSendConfig); err != nil {
		t.Errorf("SetSendConfig() error = %v", err)
	}

	// 测试设置nil配置
	if err := channel.SetSendConfig(nil); err == nil {
		t.Error("SetSendConfig(nil) should return error")
	}
}

// TestWeChatWorkChannel_Send_ChannelDisabled 测试禁用渠道时发送
func TestWeChatWorkChannel_Send_ChannelDisabled(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 禁用渠道
	channel.Disable()

	// 尝试发送
	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容")

	ctx := context.Background()
	result := channel.Send(ctx, message, nil)

	if result.Success {
		t.Error("Send() should fail when channel is disabled")
	}
	if result.Error == nil {
		t.Error("Send() should return error when channel is disabled")
	}
	if !strings.Contains(result.Error.Error(), "未启用") {
		t.Errorf("Send() error = %v, want error containing '未启用'", result.Error)
	}
}

// TestWeChatWorkChannel_Send_WithRetry 测试带重试的发送
func TestWeChatWorkChannel_Send_WithRetry(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容")

	options := &alert.SendOptions{
		Timeout:       5 * time.Second,
		Retry:         2,
		RetryInterval: 100 * time.Millisecond,
	}

	ctx := context.Background()
	// 注意：这个测试会尝试真实连接，如果Webhook不可用会失败
	result := channel.Send(ctx, message, options)

	// 检查结果结构是否正确
	if result == nil {
		t.Error("Send() returned nil result")
		return
	}
	if result.Timestamp.IsZero() {
		t.Error("Send() result should have timestamp")
	}
	if result.Duration == 0 {
		t.Error("Send() result should have duration")
	}
}

// TestWeChatWorkChannel_Send_WithCustomSendConfig 测试使用自定义发送配置
func TestWeChatWorkChannel_Send_WithCustomSendConfig(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "text",
	}

	defaultSendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList: []string{"default"},
	}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, defaultSendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 使用自定义发送配置
	customSendConfig := &alertchannel.WeChatWorkSendConfig{
		MentionedList:       []string{"custom"},
		MentionedMobileList: []string{"13800138000"},
	}

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容").
		WithExtra("send_config", customSendConfig)

	ctx := context.Background()
	result := channel.Send(ctx, message, nil)

	// 检查结果结构
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestWeChatWorkChannel_Send_RealWebhook 测试使用真实Webhook地址发送
// 注意：这个测试会实际发送消息到企业微信，需要有效的Webhook URL
func TestWeChatWorkChannel_Send_RealWebhook(t *testing.T) {
	// 跳过测试，除非显式启用
	if testing.Short() {
		t.Skip("Skipping real webhook test in short mode")
	}

	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  testWeChatWorkWebhookURL,
		MessageType: "markdown",
		Timeout:     10,
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	// 创建HTTP客户端
	httpClient, err := httpclient.NewClient(&httpclient.ClientConfig{
		Timeout: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("Failed to create HTTP client: %v", err)
	}
	defer httpClient.Close()

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, httpClient)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 测试发送Markdown消息
	t.Run("发送Markdown消息", func(t *testing.T) {
		message := alert.NewMessage().
			WithTitle("测试告警").
			WithContent("这是一条测试消息\n\n用于验证企业微信推送功能").
			WithTag("severity", "info").
			WithTag("service", "gateway")

		ctx := context.Background()
		result := channel.Send(ctx, message, nil)

		if result == nil {
			t.Error("Send() returned nil result")
			return
		}

		if !result.Success {
			t.Logf("Send failed: %v", result.Error)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
			// 如果失败，检查是否是JSON格式错误
			if result.Error != nil && strings.Contains(result.Error.Error(), "93017") {
				t.Logf("企业微信返回JSON格式错误，可能需要检查消息格式")
			}
		} else {
			t.Logf("✓ Markdown消息发送成功！耗时: %v", result.Duration)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
		}
	})

	// 测试发送Text消息
	t.Run("发送Text消息", func(t *testing.T) {
		// 创建text类型的渠道
		textServerConfig := &alertchannel.WeChatWorkServerConfig{
			WebhookURL:  testWeChatWorkWebhookURL,
			MessageType: "text",
			Timeout:     10,
		}

		textChannel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work-text", textServerConfig, sendConfig, httpClient)
		if err != nil {
			t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
		}
		defer textChannel.Close()

		message := alert.NewMessage().
			WithTitle("测试告警").
			WithContent("这是一条文本测试消息").
			WithTag("severity", "warning")

		ctx := context.Background()
		result := textChannel.Send(ctx, message, nil)

		if result == nil {
			t.Error("Send() returned nil result")
			return
		}

		if !result.Success {
			t.Logf("Send failed: %v", result.Error)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
		} else {
			t.Logf("✓ Text消息发送成功！耗时: %v", result.Duration)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
		}
	})

	// 测试带@成员的消息
	// 注意：企业微信只有text消息类型支持@成员功能，markdown类型不支持
	t.Run("发送带@成员的消息", func(t *testing.T) {
		// 创建text类型的渠道（@成员功能只在text类型中支持）
		textServerConfig := &alertchannel.WeChatWorkServerConfig{
			WebhookURL:  testWeChatWorkWebhookURL,
			MessageType: "text", // 必须使用text类型才能@成员
			Timeout:     10,
		}

		textChannel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work-mention", textServerConfig, sendConfig, httpClient)
		if err != nil {
			t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
		}
		defer textChannel.Close()

		// 使用真实的成员信息：@秦曦
		// 方式1：使用userid
		mentionConfig := &alertchannel.WeChatWorkSendConfig{
			MentionedList: []string{"QinXi"}, // userid = "秦曦"
		}

		message := alert.NewMessage().
			WithTitle("测试@成员").
			WithContent("这是一条@秦曦的消息\n\n用于测试企业微信@成员功能").
			WithExtra("send_config", mentionConfig)

		ctx := context.Background()
		result := textChannel.Send(ctx, message, nil)

		if result == nil {
			t.Error("Send() returned nil result")
			return
		}

		if !result.Success {
			t.Logf("Send failed: %v", result.Error)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
			// 如果失败，可能是 userid 不正确或成员不在群聊中
			if result.Error != nil && strings.Contains(result.Error.Error(), "企业微信返回错误") {
				t.Logf("提示：请确认 @秦曦 的 userid 是否正确，或该成员是否在群聊中")
			}
		} else {
			t.Logf("✓ @成员消息发送成功！已@秦曦（userid），耗时: %v", result.Duration)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
		}
	})

	// 测试使用手机号@成员
	t.Run("发送使用手机号@成员的消息", func(t *testing.T) {
		// 创建text类型的渠道
		textServerConfig := &alertchannel.WeChatWorkServerConfig{
			WebhookURL:  testWeChatWorkWebhookURL,
			MessageType: "text", // 必须使用text类型才能@成员
			Timeout:     10,
		}

		textChannel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work-mention-mobile", textServerConfig, sendConfig, httpClient)
		if err != nil {
			t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
		}
		defer textChannel.Close()

		// 使用手机号@秦曦（手机号：18772542347）
		mentionConfig := &alertchannel.WeChatWorkSendConfig{
			MentionedMobileList: []string{"18772542347"}, // 秦曦的真实手机号
		}

		message := alert.NewMessage().
			WithTitle("测试@成员(手机号)").
			WithContent("这是一条使用手机号@秦曦的消息").
			WithExtra("send_config", mentionConfig)

		ctx := context.Background()
		result := textChannel.Send(ctx, message, nil)

		if result == nil {
			t.Error("Send() returned nil result")
			return
		}

		if !result.Success {
			t.Logf("Send failed: %v", result.Error)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
			// 如果失败，可能是手机号不正确或成员不在群聊中
			if result.Error != nil && strings.Contains(result.Error.Error(), "企业微信返回错误") {
				t.Logf("提示：请确认手机号 18772542347 是否正确，或该成员是否在群聊中")
			}
		} else {
			t.Logf("✓ @成员消息发送成功！已@秦曦（手机号），耗时: %v", result.Duration)
			// 打印原始响应
			if result.Extra != nil {
				if respBody, ok := result.Extra["response_body"].(string); ok {
					t.Logf("响应体: %s", respBody)
				}
			}
		}
	})
}

// TestWeChatWorkChannel_HealthCheck 测试健康检查
func TestWeChatWorkChannel_HealthCheck(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
		Timeout:     5,
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	ctx := context.Background()
	// 注意：这个测试会尝试真实连接，如果Webhook不可用会失败
	result := channel.HealthCheck(ctx)

	// 由于可能没有真实的Webhook，我们只检查结果格式是否正确
	if result == nil {
		t.Error("HealthCheck() returned nil result")
		return
	}

	if !result.Success {
		if result.Error != nil && !strings.Contains(result.Error.Error(), "健康检查失败") && !strings.Contains(result.Error.Error(), "企业微信返回错误") {
			t.Errorf("HealthCheck() error = %v, want error related to health check or wechat error", result.Error)
		}
	}
}

// TestWeChatWorkChannel_BuildContent 测试内容构建
func TestWeChatWorkChannel_BuildContent(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容\n多行内容").
		WithTag("severity", "high").
		WithTag("service", "gateway")

	// 测试Markdown内容构建（通过Send间接测试）
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}

	// 测试Text内容构建
	textServerConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "text",
	}

	textChannel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work-text", textServerConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer textChannel.Close()

	result = textChannel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestWeChatWorkChannel_GenerateSign 测试签名生成
func TestWeChatWorkChannel_GenerateSign(t *testing.T) {
	serverConfig := &alertchannel.WeChatWorkServerConfig{
		WebhookURL:  "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=test",
		MessageType: "markdown",
		Secret:      "test-secret-key",
	}

	sendConfig := &alertchannel.WeChatWorkSendConfig{}

	channel, err := alertchannel.NewWeChatWorkChannel("test-wechat-work", serverConfig, sendConfig, nil)
	if err != nil {
		t.Fatalf("NewWeChatWorkChannel() failed: %v", err)
	}
	defer channel.Close()

	// 测试签名生成（通过Send间接测试，因为generateSign是私有方法）
	message := alert.NewMessage().
		WithTitle("测试签名").
		WithContent("测试签名功能")

	ctx := context.Background()
	result := channel.Send(ctx, message, nil)

	// 检查结果结构
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestTemplateReplacer_Replace 测试模板替换功能
func TestTemplateReplacer_Replace(t *testing.T) {
	replacer := alertchannel.NewTemplateReplacer()

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("这是测试内容").
		WithTag("severity", "error").
		WithTag("service", "gateway").
		WithTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))

	tests := []struct {
		name       string
		template   string
		message    *alert.Message
		customData map[string]interface{}
		want       string
	}{
		{
			name:     "基本字段替换",
			template: "标题: {{title}}, 内容: {{content}}",
			message:  message,
			want:     "标题: 测试标题, 内容: 这是测试内容",
		},
		{
			name:     "时间戳替换",
			template: "时间: {{timestamp}}",
			message:  message,
			want:     "时间: 2024-01-15 10:30:00",
		},
		{
			name:     "标签替换",
			template: "严重程度: {{tag.severity}}, 服务: {{tag.service}}",
			message:  message,
			want:     "严重程度: error, 服务: gateway",
		},
		{
			name:     "所有标签",
			template: "标签: {{tags}}",
			message:  message,
			want:     "标签: severity: error | service: gateway",
		},
		{
			name:     "自定义数据",
			template: "自定义: {{custom_field}}",
			message:  message,
			customData: map[string]interface{}{
				"custom_field": "custom_value",
			},
			want: "自定义: custom_value",
		},
		{
			name:     "混合替换",
			template: "【{{title}}】\n{{content}}\n时间: {{timestamp}}\n{{tags}}",
			message:  message,
			want:     "【测试标题】\n这是测试内容\n时间: 2024-01-15 10:30:00\nseverity: error | service: gateway",
		},
		{
			name:     "未定义的占位符保持原样",
			template: "{{title}} - {{unknown_field}}",
			message:  message,
			want:     "测试标题 - {{unknown_field}}",
		},
		{
			name:     "空模板",
			template: "",
			message:  message,
			want:     "",
		},
	}

	// 添加表格数据测试
	messageWithTable := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("这是测试内容").
		WithTag("severity", "error").
		WithTag("service", "gateway").
		WithTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)).
		WithTableData(map[string]interface{}{
			"服务名称":   "gateway-api",
			"服务状态":   "运行中",
			"CPU使用率": "45.2%",
			"内存使用率":  "62.8%",
			"请求总数":   12580,
			"错误请求数":  23,
			"平均响应时间": 125.5,
			"是否健康":   true,
		})

	tableTests := []struct {
		name       string
		template   string
		message    *alert.Message
		customData map[string]interface{}
		want       string
	}{
		{
			name:     "表格数据单个字段",
			template: "服务名称: {{table.服务名称}}, 服务状态: {{table.服务状态}}",
			message:  messageWithTable,
			want:     "服务名称: gateway-api, 服务状态: 运行中",
		},
		{
			name:     "表格数据所有字段",
			template: "表格数据: {{table}}",
			message:  messageWithTable,
			want:     "表格数据: 服务名称: gateway-api | 服务状态: 运行中 | CPU使用率: 45.2% | 内存使用率: 62.8% | 请求总数: 12580 | 错误请求数: 23 | 平均响应时间: 125.5 | 是否健康: true",
		},
		{
			name:     "表格数据混合其他字段",
			template: "【{{title}}】\n{{content}}\n表格: {{table.服务名称}} | {{table.CPU使用率}}\n时间: {{timestamp}}",
			message:  messageWithTable,
			want:     "【测试标题】\n这是测试内容\n表格: gateway-api | 45.2%\n时间: 2024-01-15 10:30:00",
		},
		{
			name:     "表格数据数字和布尔值",
			template: "请求总数: {{table.请求总数}}, 错误数: {{table.错误请求数}}, 健康: {{table.是否健康}}",
			message:  messageWithTable,
			want:     "请求总数: 12580, 错误数: 23, 健康: true",
		},
		{
			name:     "表格数据为空",
			template: "表格: {{table}}",
			message:  message,
			want:     "表格: ",
		},
	}

	// 合并测试用例
	allTests := append(tests, tableTests...)

	for _, tt := range allTests {
		t.Run(tt.name, func(t *testing.T) {
			got := replacer.Replace(tt.template, tt.message, tt.customData)
			if got != tt.want {
				t.Errorf("Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTemplateReplacer_GetPlaceholders 测试获取占位符
func TestTemplateReplacer_GetPlaceholders(t *testing.T) {
	replacer := alertchannel.NewTemplateReplacer()

	tests := []struct {
		name        string
		template    string
		wantCount   int
		wantContain []string
	}{
		{
			name:        "单个占位符",
			template:    "{{title}}",
			wantCount:   1,
			wantContain: []string{"title"},
		},
		{
			name:        "多个占位符",
			template:    "{{title}} - {{content}} - {{timestamp}}",
			wantCount:   3,
			wantContain: []string{"title", "content", "timestamp"},
		},
		{
			name:        "标签占位符",
			template:    "{{tag.severity}} - {{tag.service}}",
			wantCount:   2,
			wantContain: []string{"tag.severity", "tag.service"},
		},
		{
			name:      "无占位符",
			template:  "普通文本",
			wantCount: 0,
		},
		{
			name:      "空模板",
			template:  "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			placeholders := replacer.GetPlaceholders(tt.template)
			if len(placeholders) != tt.wantCount {
				t.Errorf("GetPlaceholders() count = %v, want %v", len(placeholders), tt.wantCount)
			}
			for _, want := range tt.wantContain {
				found := false
				for _, p := range placeholders {
					if p == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetPlaceholders() should contain %v, got %v", want, placeholders)
				}
			}
		})
	}
}

// TestTemplateReplacer_ValidateTemplate 测试模板验证
func TestTemplateReplacer_ValidateTemplate(t *testing.T) {
	replacer := alertchannel.NewTemplateReplacer()

	tests := []struct {
		name      string
		template  string
		wantValid bool
		wantErr   []string
	}{
		{
			name:      "有效模板",
			template:  "{{title}} - {{content}}",
			wantValid: true,
		},
		{
			name:      "未闭合的占位符",
			template:  "{{title}} - {{content",
			wantValid: false,
			wantErr:   []string{"未闭合的占位符"},
		},
		{
			name:      "空模板",
			template:  "",
			wantValid: true,
		},
		{
			name:      "普通文本",
			template:  "普通文本",
			wantValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, errors := replacer.ValidateTemplate(tt.template)
			if valid != tt.wantValid {
				t.Errorf("ValidateTemplate() valid = %v, want %v", valid, tt.wantValid)
			}
			if len(errors) != len(tt.wantErr) {
				t.Errorf("ValidateTemplate() errors count = %v, want %v", len(errors), len(tt.wantErr))
			}
		})
	}
}
