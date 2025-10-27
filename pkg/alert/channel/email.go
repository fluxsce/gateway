package channel

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"gateway/pkg/alert"
)

// EmailServerConfig SMTP服务器配置（可公用）
type EmailServerConfig struct {
	// SMTPHost SMTP服务器地址
	SMTPHost string
	// SMTPPort SMTP服务器端口
	SMTPPort int
	// Username 用户名
	Username string
	// Password 密码
	Password string
	// From 默认发件人地址（固定，属于服务器配置）
	From string
	// FromName 默认发件人名称
	FromName string
	// UseTLS 是否使用TLS
	UseTLS bool
	// SkipVerify 是否跳过证书验证
	SkipVerify bool
	// Timeout 超时时间（秒）
	Timeout int
}

// Validate 验证服务器配置
func (c *EmailServerConfig) Validate() error {
	if c.SMTPHost == "" {
		return fmt.Errorf("SMTP服务器地址不能为空")
	}
	if c.SMTPPort <= 0 || c.SMTPPort > 65535 {
		return fmt.Errorf("SMTP端口号无效: %d", c.SMTPPort)
	}
	if c.Username == "" {
		return fmt.Errorf("用户名不能为空")
	}
	if c.Password == "" {
		return fmt.Errorf("密码不能为空")
	}
	if c.From == "" {
		return fmt.Errorf("发件人地址不能为空")
	}
	return nil
}

// EmailSendConfig 邮件发送配置（每次发送可不同）
type EmailSendConfig struct {
	// To 收件人列表
	To []string
	// CC 抄送列表
	CC []string
	// BCC 密送列表
	BCC []string
}

// Validate 验证发送配置
func (c *EmailSendConfig) Validate() error {
	if len(c.To) == 0 {
		return fmt.Errorf("收件人列表不能为空")
	}
	return nil
}

// EmailChannel 邮件告警渠道
type EmailChannel struct {
	*alert.BaseChannel
	serverConfig *EmailServerConfig
	sendConfig   *EmailSendConfig
}

// NewEmailChannel 创建邮件告警渠道
// 参数:
//
//	name: 渠道名称
//	serverConfig: SMTP服务器配置（可公用）
//	sendConfig: 默认发送配置（可在发送时覆盖）
func NewEmailChannel(name string, serverConfig *EmailServerConfig, sendConfig *EmailSendConfig) (*EmailChannel, error) {
	if serverConfig == nil {
		return nil, fmt.Errorf("服务器配置不能为空")
	}

	if err := serverConfig.Validate(); err != nil {
		return nil, fmt.Errorf("服务器配置验证失败: %w", err)
	}

	if sendConfig == nil {
		return nil, fmt.Errorf("发送配置不能为空")
	}

	if err := sendConfig.Validate(); err != nil {
		return nil, fmt.Errorf("发送配置验证失败: %w", err)
	}

	channel := &EmailChannel{
		BaseChannel:  alert.NewBaseChannel(name, alert.AlertTypeEmail),
		serverConfig: serverConfig,
		sendConfig:   sendConfig,
	}

	return channel, nil
}

// SetSendConfig 设置或更新发送配置
func (e *EmailChannel) SetSendConfig(config *EmailSendConfig) error {
	if config == nil {
		return fmt.Errorf("发送配置不能为空")
	}
	if err := config.Validate(); err != nil {
		return fmt.Errorf("发送配置验证失败: %w", err)
	}
	e.sendConfig = config
	return nil
}

// Send 发送邮件告警
// 可以在 message.Extra 中传入 "send_config" 来覆盖默认的发送配置
func (e *EmailChannel) Send(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	startTime := time.Now()
	result := &alert.SendResult{
		Success:   false,
		Timestamp: startTime,
		Extra:     make(map[string]interface{}),
	}

	// 检查渠道是否启用
	if !e.IsEnabled() {
		result.Error = fmt.Errorf("邮件告警渠道未启用")
		result.Duration = time.Since(startTime)
		e.UpdateStats(result)
		return result
	}

	// 使用默认选项
	if options == nil {
		options = alert.DefaultSendOptions()
	}

	// 设置超时
	sendCtx := ctx
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		sendCtx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}

	// 执行发送（带重试）
	var lastErr error
	maxRetries := options.Retry
	if maxRetries < 1 {
		maxRetries = 1
	}

	for i := 0; i < maxRetries; i++ {
		// 检查上下文是否已取消
		select {
		case <-sendCtx.Done():
			result.Error = fmt.Errorf("发送超时或被取消: %w", sendCtx.Err())
			result.Duration = time.Since(startTime)
			e.UpdateStats(result)
			return result
		default:
		}

		// 尝试发送
		err := e.sendEmail(sendCtx, message)
		if err == nil {
			result.Success = true
			result.Duration = time.Since(startTime)
			e.UpdateStats(result)
			return result
		}

		lastErr = err

		// 如果不是最后一次重试，等待后重试
		if i < maxRetries-1 && options.RetryInterval > 0 {
			select {
			case <-sendCtx.Done():
				result.Error = fmt.Errorf("重试等待期间被取消: %w", sendCtx.Err())
				result.Duration = time.Since(startTime)
				e.UpdateStats(result)
				return result
			case <-time.After(options.RetryInterval):
				// 继续重试
			}
		}
	}

	result.Error = fmt.Errorf("发送失败（重试%d次）: %w", maxRetries, lastErr)
	result.Duration = time.Since(startTime)
	e.UpdateStats(result)
	return result
}

// sendEmail 实际发送邮件的内部方法
func (e *EmailChannel) sendEmail(ctx context.Context, message *alert.Message) error {
	// 获取发送配置（支持从消息中覆盖）
	sendConfig := e.sendConfig
	if customConfig, ok := message.Extra["send_config"].(*EmailSendConfig); ok && customConfig != nil {
		if err := customConfig.Validate(); err == nil {
			sendConfig = customConfig
		}
	}

	// 构建邮件内容
	emailContent := e.buildEmailContent(message, sendConfig)

	// 连接SMTP服务器
	addr := fmt.Sprintf("%s:%d", e.serverConfig.SMTPHost, e.serverConfig.SMTPPort)

	// 根据配置选择连接方式
	if e.serverConfig.UseTLS {
		return e.sendWithTLS(addr, emailContent, sendConfig)
	}
	return e.sendWithPlain(addr, emailContent, sendConfig)
}

// sendWithTLS 使用TLS发送邮件
func (e *EmailChannel) sendWithTLS(addr string, content []byte, sendConfig *EmailSendConfig) error {
	// 配置TLS
	tlsConfig := &tls.Config{
		ServerName:         e.serverConfig.SMTPHost,
		InsecureSkipVerify: e.serverConfig.SkipVerify,
	}

	// 建立TLS连接
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("TLS连接失败: %w", err)
	}
	defer conn.Close()

	// 创建SMTP客户端
	client, err := smtp.NewClient(conn, e.serverConfig.SMTPHost)
	if err != nil {
		return fmt.Errorf("创建SMTP客户端失败: %w", err)
	}
	defer client.Close()

	// 认证
	auth := smtp.PlainAuth("", e.serverConfig.Username, e.serverConfig.Password, e.serverConfig.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP认证失败: %w", err)
	}

	// 设置发件人（使用服务器配置中的固定发件人）
	if err := client.Mail(e.serverConfig.From); err != nil {
		return fmt.Errorf("设置发件人失败: %w", err)
	}

	// 设置收件人
	recipients := append([]string{}, sendConfig.To...)
	recipients = append(recipients, sendConfig.CC...)
	recipients = append(recipients, sendConfig.BCC...)

	for _, to := range recipients {
		if err := client.Rcpt(to); err != nil {
			return fmt.Errorf("设置收件人失败: %w", err)
		}
	}

	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("开始发送数据失败: %w", err)
	}

	_, err = w.Write(content)
	if err != nil {
		return fmt.Errorf("写入邮件内容失败: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("完成邮件发送失败: %w", err)
	}

	return client.Quit()
}

// sendWithPlain 使用普通连接发送邮件
func (e *EmailChannel) sendWithPlain(addr string, content []byte, sendConfig *EmailSendConfig) error {
	// 认证
	auth := smtp.PlainAuth("", e.serverConfig.Username, e.serverConfig.Password, e.serverConfig.SMTPHost)

	// 收件人列表
	recipients := append([]string{}, sendConfig.To...)
	recipients = append(recipients, sendConfig.CC...)
	recipients = append(recipients, sendConfig.BCC...)

	// 发送邮件（使用服务器配置中的固定发件人）
	err := smtp.SendMail(addr, auth, e.serverConfig.From, recipients, content)
	if err != nil {
		return fmt.Errorf("发送邮件失败: %w", err)
	}

	return nil
}

// buildEmailContent 构建邮件内容
func (e *EmailChannel) buildEmailContent(message *alert.Message, sendConfig *EmailSendConfig) []byte {
	var content strings.Builder

	// 邮件头 - 发件人使用服务器配置中的固定值
	from := e.serverConfig.From
	if e.serverConfig.FromName != "" {
		from = fmt.Sprintf("%s <%s>", e.serverConfig.FromName, e.serverConfig.From)
	}
	content.WriteString(fmt.Sprintf("From: %s\r\n", from))
	content.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(sendConfig.To, ",")))

	if len(sendConfig.CC) > 0 {
		content.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(sendConfig.CC, ",")))
	}

	// 主题
	subject := message.Title
	if subject == "" {
		subject = "系统告警"
	}
	content.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	// MIME类型
	content.WriteString("MIME-Version: 1.0\r\n")
	content.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	content.WriteString("\r\n")

	// 邮件正文
	content.WriteString(e.buildHTMLBody(message))

	return []byte(content.String())
}

// buildHTMLBody 构建HTML格式的邮件正文
func (e *EmailChannel) buildHTMLBody(message *alert.Message) string {
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\r\n")
	html.WriteString("<html>\r\n")
	html.WriteString("<head><meta charset=\"UTF-8\"></head>\r\n")
	html.WriteString("<body style=\"font-family: Arial, sans-serif; padding: 20px;\">\r\n")

	// 标题
	if message.Title != "" {
		html.WriteString(fmt.Sprintf("<h2 style=\"color: #333;\">%s</h2>\r\n", message.Title))
	}

	// 内容
	html.WriteString("<div style=\"background-color: #f5f5f5; padding: 15px; border-radius: 5px; margin-bottom: 20px;\">\r\n")
	html.WriteString(fmt.Sprintf("<pre style=\"white-space: pre-wrap; word-wrap: break-word;\">%s</pre>\r\n", message.Content))
	html.WriteString("</div>\r\n")

	// 时间戳
	html.WriteString(fmt.Sprintf("<p style=\"color: #666;\">告警时间: %s</p>\r\n", message.Timestamp.Format("2006-01-02 15:04:05")))

	// 标签
	if len(message.Tags) > 0 {
		html.WriteString("<div style=\"margin-top: 20px;\">\r\n")
		html.WriteString("<strong>标签:</strong><br>\r\n")
		for k, v := range message.Tags {
			html.WriteString(fmt.Sprintf("<span style=\"background-color: #e0e0e0; padding: 5px 10px; margin: 5px; border-radius: 3px; display: inline-block;\">%s: %s</span>\r\n", k, v))
		}
		html.WriteString("</div>\r\n")
	}

	html.WriteString("</body>\r\n")
	html.WriteString("</html>\r\n")

	return html.String()
}

// Close 关闭渠道
func (e *EmailChannel) Close() error {
	// 邮件渠道无需特殊清理
	return nil
}

// HealthCheck 健康检查
func (e *EmailChannel) HealthCheck(ctx context.Context) error {
	// 尝试连接SMTP服务器
	addr := net.JoinHostPort(e.serverConfig.SMTPHost, fmt.Sprintf("%d", e.serverConfig.SMTPPort))

	timeout := 10 * time.Second
	if e.serverConfig.Timeout > 0 {
		timeout = time.Duration(e.serverConfig.Timeout) * time.Second
	}

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("无法连接到SMTP服务器: %w", err)
	}
	defer conn.Close()

	return nil
}
