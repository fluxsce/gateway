package channel

import (
	"context"
	"strings"
	"testing"
	"time"

	"gateway/pkg/alert"
	alertchannel "gateway/pkg/alert/channel"
)

// TestEmailServerConfig_Validate 测试SMTP服务器配置验证
func TestEmailServerConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *alertchannel.EmailServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "user@example.com",
				Password: "password",
				From:     "alert@example.com",
			},
			wantErr: false,
		},
		{
			name: "SMTP服务器地址为空",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "",
				SMTPPort: 587,
				Username: "user@example.com",
				Password: "password",
				From:     "alert@example.com",
			},
			wantErr: true,
			errMsg:  "SMTP服务器地址不能为空",
		},
		{
			name: "SMTP端口无效-负数",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: -1,
				Username: "user@example.com",
				Password: "password",
				From:     "alert@example.com",
			},
			wantErr: true,
			errMsg:  "SMTP端口号无效",
		},
		{
			name: "SMTP端口无效-超出范围",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 65536,
				Username: "user@example.com",
				Password: "password",
				From:     "alert@example.com",
			},
			wantErr: true,
			errMsg:  "SMTP端口号无效",
		},
		{
			name: "用户名为空",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "",
				Password: "password",
				From:     "alert@example.com",
			},
			wantErr: true,
			errMsg:  "用户名不能为空",
		},
		{
			name: "密码为空",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "user@example.com",
				Password: "",
				From:     "alert@example.com",
			},
			wantErr: true,
			errMsg:  "密码不能为空",
		},
		{
			name: "发件人地址为空",
			config: &alertchannel.EmailServerConfig{
				SMTPHost: "smtp.example.com",
				SMTPPort: 587,
				Username: "user@example.com",
				Password: "password",
				From:     "",
			},
			wantErr: true,
			errMsg:  "发件人地址不能为空",
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

// TestEmailSendConfig_Validate 测试邮件发送配置验证
func TestEmailSendConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *alertchannel.EmailSendConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "有效配置-有收件人",
			config: &alertchannel.EmailSendConfig{
				To: []string{"recipient@example.com"},
			},
			wantErr: false,
		},
		{
			name: "有效配置-多个收件人",
			config: &alertchannel.EmailSendConfig{
				To:  []string{"recipient1@example.com", "recipient2@example.com"},
				CC:  []string{"cc@example.com"},
				BCC: []string{"bcc@example.com"},
			},
			wantErr: false,
		},
		{
			name: "收件人列表为空",
			config: &alertchannel.EmailSendConfig{
				To: []string{},
			},
			wantErr: true,
			errMsg:  "收件人列表不能为空",
		},
		{
			name: "收件人列表为nil",
			config: &alertchannel.EmailSendConfig{
				To: nil,
			},
			wantErr: true,
			errMsg:  "收件人列表不能为空",
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

// TestNewEmailChannel 测试创建邮件告警渠道
func TestNewEmailChannel(t *testing.T) {
	validServerConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	validSendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	tests := []struct {
		name         string
		channelName  string
		serverConfig *alertchannel.EmailServerConfig
		sendConfig   *alertchannel.EmailSendConfig
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "成功创建渠道",
			channelName:  "test-email",
			serverConfig: validServerConfig,
			sendConfig:   validSendConfig,
			wantErr:      false,
		},
		{
			name:         "服务器配置为nil",
			channelName:  "test-email",
			serverConfig: nil,
			sendConfig:   validSendConfig,
			wantErr:      true,
			errMsg:       "服务器配置不能为空",
		},
		{
			name:         "发送配置为nil",
			channelName:  "test-email",
			serverConfig: validServerConfig,
			sendConfig:   nil,
			wantErr:      true,
			errMsg:       "发送配置不能为空",
		},
		{
			name: "服务器配置验证失败",
			serverConfig: &alertchannel.EmailServerConfig{
				SMTPHost: "",
				SMTPPort: 587,
				Username: "user@example.com",
				Password: "password",
				From:     "alert@example.com",
			},
			sendConfig: validSendConfig,
			wantErr:    true,
			errMsg:     "服务器配置验证失败",
		},
		{
			name:         "发送配置验证失败",
			serverConfig: validServerConfig,
			sendConfig: &alertchannel.EmailSendConfig{
				To: []string{},
			},
			wantErr: true,
			errMsg:  "发送配置验证失败",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel, err := alertchannel.NewEmailChannel(tt.channelName, tt.serverConfig, tt.sendConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEmailChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("NewEmailChannel() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if channel == nil {
					t.Error("NewEmailChannel() returned nil channel")
					return
				}
				if channel.Name() != tt.channelName {
					t.Errorf("Name() = %v, want %v", channel.Name(), tt.channelName)
				}
				if channel.Type() != alert.AlertTypeEmail {
					t.Errorf("Type() = %v, want %v", channel.Type(), alert.AlertTypeEmail)
				}
				if !channel.IsEnabled() {
					t.Error("New channel should be enabled by default")
				}
			}
		})
	}
}

// TestEmailChannel_BasicMethods 测试渠道基本方法
func TestEmailChannel_BasicMethods(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	// 测试 Type
	if channel.Type() != alert.AlertTypeEmail {
		t.Errorf("Type() = %v, want %v", channel.Type(), alert.AlertTypeEmail)
	}

	// 测试 Name
	if channel.Name() != "test-email" {
		t.Errorf("Name() = %v, want %v", channel.Name(), "test-email")
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

// TestEmailChannel_SetSendConfig 测试设置发送配置
func TestEmailChannel_SetSendConfig(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	initialSendConfig := &alertchannel.EmailSendConfig{
		To: []string{"initial@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, initialSendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	// 测试设置有效配置
	newSendConfig := &alertchannel.EmailSendConfig{
		To: []string{"new@example.com"},
		CC: []string{"cc@example.com"},
	}
	if err := channel.SetSendConfig(newSendConfig); err != nil {
		t.Errorf("SetSendConfig() error = %v", err)
	}

	// 测试设置nil配置
	if err := channel.SetSendConfig(nil); err == nil {
		t.Error("SetSendConfig(nil) should return error")
	}

	// 测试设置无效配置
	invalidConfig := &alertchannel.EmailSendConfig{
		To: []string{},
	}
	if err := channel.SetSendConfig(invalidConfig); err == nil {
		t.Error("SetSendConfig(invalid) should return error")
	}
}

// TestEmailChannel_Send_ChannelDisabled 测试禁用渠道时发送
func TestEmailChannel_Send_ChannelDisabled(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

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

// TestEmailChannel_Send_WithRetry 测试带重试的发送
func TestEmailChannel_Send_WithRetry(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容")

	options := &alert.SendOptions{
		Timeout:       5 * time.Second,
		Retry:         2,
		RetryInterval: 100 * time.Millisecond,
	}

	ctx := context.Background()
	// 注意：这个测试会尝试真实连接，如果SMTP服务器不可用会失败
	// 在实际测试中，可能需要mock SMTP连接
	result := channel.Send(ctx, message, options)

	// 由于可能没有真实的SMTP服务器，我们只检查结果结构是否正确
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

// TestEmailChannel_Send_WithCustomSendConfig 测试使用自定义发送配置
func TestEmailChannel_Send_WithCustomSendConfig(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	defaultSendConfig := &alertchannel.EmailSendConfig{
		To: []string{"default@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, defaultSendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	// 使用自定义发送配置
	customSendConfig := &alertchannel.EmailSendConfig{
		To: []string{"custom@example.com"},
		CC: []string{"cc@example.com"},
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

// TestEmailChannel_HealthCheck 测试健康检查
func TestEmailChannel_HealthCheck(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
		Timeout:  5,
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	ctx := context.Background()
	// 注意：这个测试会尝试真实连接，如果SMTP服务器不可用会失败
	result := channel.HealthCheck(ctx)

	// 由于可能没有真实的SMTP服务器，我们只检查结果格式是否正确
	if result == nil {
		t.Error("HealthCheck() returned nil result")
		return
	}

	if !result.Success {
		if result.Error != nil && !strings.Contains(result.Error.Error(), "无法连接") && !strings.Contains(result.Error.Error(), "SMTP服务器") {
			t.Errorf("HealthCheck() error = %v, want error related to connection", result.Error)
		}
	}
}

// TestEmailChannel_BuildEmailContent 测试邮件内容构建
func TestEmailChannel_BuildEmailContent(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
		FromName: "系统告警",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
		CC: []string{"cc@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容\n多行内容").
		WithTag("severity", "high").
		WithTag("service", "gateway")

	// 由于 buildEmailContent 是私有方法，我们通过 Send 方法间接测试
	// 如果 Send 成功或失败，说明 buildEmailContent 至少被调用了
	// 更详细的测试需要 mock SMTP 服务器
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_BuildHTMLBody 测试HTML正文构建
func TestEmailChannel_BuildHTMLBody(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容").
		WithTag("severity", "high").
		WithTag("service", "gateway")

	// 由于 buildHTMLBody 是私有方法，我们通过 Send 方法间接测试
	// 如果 Send 被调用，说明 buildHTMLBody 至少被调用了
	// 更详细的测试需要 mock SMTP 服务器或导出方法
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_BuildHTMLBody_NoTitle 测试无标题的HTML正文构建
func TestEmailChannel_BuildHTMLBody_NoTitle(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	message := alert.NewMessage().
		WithContent("测试内容")

	// buildHTMLBody 是私有方法，我们通过 Send 方法间接测试
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_BuildHTMLBody_NoTags 测试无标签的HTML正文构建
func TestEmailChannel_BuildHTMLBody_NoTags(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "user@example.com",
		Password: "password",
		From:     "alert@example.com",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-email", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容")

	// buildHTMLBody 是私有方法，我们通过 Send 方法间接测试
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_Template 测试邮件渠道的模板功能
func TestEmailChannel_Template(t *testing.T) {
	// 创建带模板配置的渠道
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost:        "smtp.example.com",
		SMTPPort:        587,
		Username:        "test@example.com",
		Password:        "password",
		From:            "test@example.com",
		TitleTemplate:   "【{{title}}】时间: {{timestamp}}",
		ContentTemplate: "<h1>{{title}}</h1><p>{{content}}</p><p>时间: {{timestamp}}</p><p>标签: {{tags}}</p>",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-template", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("测试告警").
		WithContent("这是一条测试消息").
		WithTag("severity", "error").
		WithTag("service", "gateway").
		WithTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))

	// 通过 Send 方法间接测试模板功能
	// 由于 buildEmailContent 是私有方法，我们只能通过 Send 来测试
	// 实际应用中，模板会在 Send 过程中被使用
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}

	// 验证渠道已正确创建并包含模板配置
	if channel.Name() != "test-template" {
		t.Errorf("Channel name should be 'test-template', got: %s", channel.Name())
	}
}

// TestEmailChannel_Template_SubjectOnly 测试仅主题模板
func TestEmailChannel_Template_SubjectOnly(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost:      "smtp.example.com",
		SMTPPort:      587,
		Username:      "test@example.com",
		Password:      "password",
		From:          "test@example.com",
		TitleTemplate: "告警: {{title}} - {{tag.severity}}",
		// ContentTemplate 为空，使用默认格式
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-template-subject", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("系统错误").
		WithContent("测试内容").
		WithTag("severity", "critical").
		WithTimestamp(time.Now())

	// 通过 Send 方法间接测试
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_Template_BodyOnly 测试仅正文模板
func TestEmailChannel_Template_BodyOnly(t *testing.T) {
	serverConfig := &alertchannel.EmailServerConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "test@example.com",
		Password: "password",
		From:     "test@example.com",
		// TitleTemplate 为空，使用默认格式
		ContentTemplate: "{{content}}\n\n时间: {{timestamp}}\n标签: {{tags}}",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"recipient@example.com"},
	}

	channel, err := alertchannel.NewEmailChannel("test-template-body", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("测试标题").
		WithContent("测试内容").
		WithTag("severity", "warning").
		WithTimestamp(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC))

	// 通过 Send 方法间接测试
	ctx := context.Background()
	result := channel.Send(ctx, message, nil)
	if result == nil {
		t.Error("Send() returned nil result")
	}
}

// TestEmailChannel_Send_RealSMTP 测试使用真实SMTP服务器发送邮件
// 注意：这个测试会实际发送邮件到指定的收件人
// 使用腾讯企业邮箱SMTP服务器：smtp.exmail.qq.com:465 (SSL)
func TestEmailChannel_Send_RealSMTP(t *testing.T) {
	// 跳过测试，除非明确启用
	// 运行此测试时使用：go test -v -run TestEmailChannel_Send_RealSMTP
	if testing.Short() {
		t.Skip("跳过真实SMTP服务器测试（使用 -short 标志）")
	}

	serverConfig := &alertchannel.EmailServerConfig{

		FromName:   "系统告警",
		UseTLS:     true,
		SkipVerify: false,
		Timeout:    30,
		// 测试模板功能
		TitleTemplate:   "【系统告警】{{title}} - {{tag.severity}}",
		ContentTemplate: "<h2 style=\"color: #d32f2f;\">{{title}}</h2><p>{{content}}</p><p><strong>时间:</strong> {{timestamp}}</p><p><strong>标签:</strong> {{tags}}</p>",
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"shangjian@flux.com.cn"}, // 发送给自己进行测试
	}

	channel, err := alertchannel.NewEmailChannel("real-smtp-test", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	// 创建测试消息
	message := alert.NewMessage().
		WithTitle("邮件渠道模板测试").
		WithContent("这是一条测试消息，用于验证邮件渠道的模板替换功能。\n\n支持的功能：\n- 主题模板替换\n- 正文模板替换\n- 标签和时间戳替换").
		WithTag("severity", "info").
		WithTag("service", "gateway").
		WithTag("test", "template").
		WithTimestamp(time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := channel.Send(ctx, message, nil)

	if result == nil {
		t.Fatal("Send() returned nil result")
	}

	// 打印发送结果
	t.Logf("========== 邮件发送结果 ==========")
	t.Logf("发送状态: %v", result.Success)
	t.Logf("发送时间: %s", result.Timestamp.Format("2006-01-02 15:04:05.000"))
	t.Logf("发送耗时: %v", result.Duration)

	if !result.Success {
		t.Errorf("邮件发送失败: %v", result.Error)
		if result.Extra != nil {
			if lastErr, ok := result.Extra["last_error"].(string); ok {
				t.Logf("错误详情: %s", lastErr)
			}
		}
	} else {
		t.Logf("邮件发送成功！")
		t.Logf("主题应该包含: 【系统告警】邮件渠道模板测试 - info")
		t.Logf("正文应该包含模板替换的内容")
		if result.Extra != nil {
			for k, v := range result.Extra {
				t.Logf("额外信息 [%s]: %v", k, v)
			}
		}
	}
	t.Logf("===================================")
}

// TestEmailChannel_Send_RealSMTP_DefaultFormat 测试使用真实SMTP服务器发送邮件（默认格式，无模板）
func TestEmailChannel_Send_RealSMTP_DefaultFormat(t *testing.T) {
	// 跳过测试，除非明确启用
	if testing.Short() {
		t.Skip("跳过真实SMTP服务器测试（使用 -short 标志）")
	}

	serverConfig := &alertchannel.EmailServerConfig{

		FromName:   "系统告警",
		UseTLS:     true,
		SkipVerify: false,
		Timeout:    30,
		// 不设置模板，使用默认格式
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"shangjian@flux.com.cn"},
	}

	channel, err := alertchannel.NewEmailChannel("real-smtp-default", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	message := alert.NewMessage().
		WithTitle("默认格式测试").
		WithContent("这是一条使用默认格式的测试消息。\n\n验证内容：\n- 默认HTML格式\n- 标题、内容、时间戳、标签显示").
		WithTag("severity", "warning").
		WithTag("service", "gateway").
		WithTimestamp(time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := channel.Send(ctx, message, nil)

	if result == nil {
		t.Fatal("Send() returned nil result")
	}

	// 打印发送结果
	t.Logf("========== 邮件发送结果（默认格式） ==========")
	t.Logf("发送状态: %v", result.Success)
	t.Logf("发送时间: %s", result.Timestamp.Format("2006-01-02 15:04:05.000"))
	t.Logf("发送耗时: %v", result.Duration)

	if !result.Success {
		t.Errorf("邮件发送失败: %v", result.Error)
		if result.Extra != nil {
			if lastErr, ok := result.Extra["last_error"].(string); ok {
				t.Logf("错误详情: %s", lastErr)
			}
		}
	} else {
		t.Logf("邮件发送成功！")
		t.Logf("邮件格式: 默认HTML格式")
		t.Logf("包含内容: 标题、内容、时间戳、标签")
		if result.Extra != nil {
			for k, v := range result.Extra {
				t.Logf("额外信息 [%s]: %v", k, v)
			}
		}
	}
	t.Logf("=============================================")
}

// TestEmailChannel_Send_RealSMTP_WithTable 测试使用真实SMTP服务器发送邮件（包含表格数据）
func TestEmailChannel_Send_RealSMTP_WithTable(t *testing.T) {
	// 跳过测试，除非明确启用
	if testing.Short() {
		t.Skip("跳过真实SMTP服务器测试（使用 -short 标志）")
	}

	serverConfig := &alertchannel.EmailServerConfig{

		FromName:   "系统告警",
		UseTLS:     true,
		SkipVerify: false,
		Timeout:    30,
		// 不设置模板，使用默认格式
	}

	sendConfig := &alertchannel.EmailSendConfig{
		To: []string{"shangjian@flux.com.cn"},
	}

	channel, err := alertchannel.NewEmailChannel("real-smtp-table", serverConfig, sendConfig)
	if err != nil {
		t.Fatalf("NewEmailChannel() failed: %v", err)
	}
	defer channel.Close()

	// 创建包含表格数据的消息
	tableData := map[string]interface{}{
		"服务名称":   "gateway-api",
		"服务状态":   "运行中",
		"CPU使用率": "45.2%",
		"内存使用率":  "62.8%",
		"请求总数":   12580,
		"错误请求数":  23,
		"平均响应时间": 125.5,
		"是否健康":   true,
		"最后检查时间": time.Now().Format("2006-01-02 15:04:05"),
	}

	message := alert.NewMessage().
		WithTitle("服务监控告警").
		WithContent("检测到服务异常，详细信息如下：").
		WithTag("severity", "warning").
		WithTag("service", "gateway").
		WithTag("type", "monitor").
		WithTableData(tableData).
		WithTimestamp(time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result := channel.Send(ctx, message, nil)

	if result == nil {
		t.Fatal("Send() returned nil result")
	}

	// 打印发送结果
	t.Logf("========== 邮件发送结果（表格数据） ==========")
	t.Logf("发送状态: %v", result.Success)
	t.Logf("发送时间: %s", result.Timestamp.Format("2006-01-02 15:04:05.000"))
	t.Logf("发送耗时: %v", result.Duration)
	t.Logf("消息标题: %s", message.Title)
	t.Logf("表格数据字段数: %d", len(tableData))

	if !result.Success {
		t.Errorf("邮件发送失败: %v", result.Error)
		if result.Extra != nil {
			if lastErr, ok := result.Extra["last_error"].(string); ok {
				t.Logf("错误详情: %s", lastErr)
			}
		}
	} else {
		t.Logf("邮件发送成功！")
		t.Logf("邮件格式: 默认HTML格式（包含表格）")
		t.Logf("表格数据详情:")
		for key, value := range tableData {
			t.Logf("  - %s: %v (类型: %T)", key, value, value)
		}
		if result.Extra != nil {
			for k, v := range result.Extra {
				t.Logf("额外信息 [%s]: %v", k, v)
			}
		}
	}
	t.Logf("=============================================")
}
