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
	// TitleTemplate 邮件主题模板（可选）
	// 支持占位符：{{title}}, {{timestamp}}, {{tag.key}} 等
	// 如果为空，使用默认格式（消息标题或"系统告警"）
	TitleTemplate string
	// ContentTemplate 邮件正文模板（可选）
	// 支持占位符：{{content}}, {{title}}, {{tags}}, {{timestamp}} 等
	// 如果为空，使用默认HTML格式
	ContentTemplate string
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
	name             string
	channelType      alert.AlertType
	enabled          bool
	serverConfig     *EmailServerConfig
	sendConfig       *EmailSendConfig
	templateReplacer *TemplateReplacer
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
		name:             name,
		channelType:      alert.AlertTypeEmail,
		enabled:          true,
		serverConfig:     serverConfig,
		sendConfig:       sendConfig,
		templateReplacer: NewTemplateReplacer(),
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
			return result
		default:
		}

		// 尝试发送
		err := e.sendEmail(sendCtx, message)
		if err == nil {
			result.Success = true
			result.Duration = time.Since(startTime)
			return result
		}

		lastErr = err

		// 如果不是最后一次重试，等待后重试
		if i < maxRetries-1 && options.RetryInterval > 0 {
			select {
			case <-sendCtx.Done():
				result.Error = fmt.Errorf("重试等待期间被取消: %w", sendCtx.Err())
				result.Duration = time.Since(startTime)
				return result
			case <-time.After(options.RetryInterval):
				// 继续重试
			}
		}
	}

	result.Error = fmt.Errorf("发送失败（重试%d次）: %w", maxRetries, lastErr)
	result.Duration = time.Since(startTime)
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

	// 主题（如果配置了主题模板，使用模板；否则使用默认格式）
	var subject string
	if e.serverConfig.TitleTemplate != "" {
		// 有模板时，直接使用模板内容，不添加图标
		subject = e.templateReplacer.Replace(e.serverConfig.TitleTemplate, message, nil)
	} else {
		// 没有模板时，使用默认格式并添加警报图标
		subject = message.Title
		if subject == "" {
			subject = "系统告警"
		}
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
	// 如果配置了正文模板，使用模板替换
	if e.serverConfig.ContentTemplate != "" {
		return e.templateReplacer.Replace(e.serverConfig.ContentTemplate, message, nil)
	}

	// 否则使用默认HTML格式（美化版）
	var html strings.Builder
	html.WriteString("<!DOCTYPE html>\r\n")
	html.WriteString("<html>\r\n")
	html.WriteString("<head>\r\n")
	html.WriteString("<meta charset=\"UTF-8\">\r\n")
	html.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\r\n")
	html.WriteString("</head>\r\n")
	html.WriteString("<body style=\"margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Helvetica Neue', Arial, sans-serif; background-color: #f5f7fa; line-height: 1.6;\">\r\n")

	// 外层容器
	html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"background-color: #f5f7fa; padding: 0;\">\r\n")
	html.WriteString("<tr>\r\n")
	html.WriteString("<td style=\"padding: 20px;\">\r\n")

	// 主内容卡片（占满父级容器）
	html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"background-color: #ffffff; border-radius: 8px; box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1); overflow: hidden; max-width: 100%;\">\r\n")

	// 标题区域（带背景色）
	// 如果没有配置 ContentTemplate，使用默认格式时添加图标
	if message.Title != "" {
		html.WriteString("<tr>\r\n")
		html.WriteString("<td style=\"background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); padding: 24px 30px; color: #ffffff;\">\r\n")
		// 如果没有配置 ContentTemplate，使用默认格式时添加图标
		if e.serverConfig.ContentTemplate == "" {
			html.WriteString(fmt.Sprintf("<h1 style=\"margin: 0; font-size: 24px; font-weight: 600; line-height: 1.3; color: #ffffff;\">🚨 %s</h1>\r\n", e.escapeHTML(message.Title)))
		} else {
			html.WriteString(fmt.Sprintf("<h1 style=\"margin: 0; font-size: 24px; font-weight: 600; line-height: 1.3; color: #ffffff;\">%s</h1>\r\n", e.escapeHTML(message.Title)))
		}
		html.WriteString("</td>\r\n")
		html.WriteString("</tr>\r\n")
	}

	// 内容区域
	html.WriteString("<tr>\r\n")
	html.WriteString("<td style=\"padding: 10px;\">\r\n")

	// 消息内容
	if message.Content != "" {
		html.WriteString("<div style=\"background-color: #f8f9fa; border-left: 4px solid #667eea; padding: 16px 20px; margin-bottom: 24px; border-radius: 4px;\">\r\n")
		html.WriteString(fmt.Sprintf("<div style=\"color: #2d3748; font-size: 15px; white-space: pre-wrap; word-wrap: break-word; line-height: 1.7;\">%s</div>\r\n", e.escapeHTML(message.Content)))
		html.WriteString("</div>\r\n")
	}

	// 表格数据（如果 TableData 不为空，且 DisplayFormat 为 table 或未设置）
	if len(message.TableData) > 0 && (message.DisplayFormat == alert.DisplayFormatTable || message.DisplayFormat == "") {
		html.WriteString("<div style=\"margin-bottom: 24px;\">\r\n")
		html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\" style=\"border-collapse: collapse; background-color: #ffffff; border: 1px solid #e2e8f0; border-radius: 6px; overflow: hidden;\">\r\n")

		// 表头（可选，如果 TableData 有特定的键可以显示为表头）
		// 这里我们直接显示为键值对表格
		html.WriteString("<thead>\r\n")
		html.WriteString("<tr style=\"background-color: #f7fafc;\">\r\n")
		html.WriteString("<th style=\"padding: 12px 16px; text-align: left; font-weight: 600; color: #2d3748; font-size: 14px; border-bottom: 2px solid #e2e8f0;\">字段</th>\r\n")
		html.WriteString("<th style=\"padding: 12px 16px; text-align: left; font-weight: 600; color: #2d3748; font-size: 14px; border-bottom: 2px solid #e2e8f0;\">值</th>\r\n")
		html.WriteString("</tr>\r\n")
		html.WriteString("</thead>\r\n")

		// 表格内容
		html.WriteString("<tbody>\r\n")
		for key, value := range message.TableData {
			html.WriteString("<tr style=\"border-bottom: 1px solid #e2e8f0;\">\r\n")
			html.WriteString(fmt.Sprintf("<td style=\"padding: 12px 16px; font-weight: 500; color: #4a5568; font-size: 14px; background-color: #f7fafc; width: 30%%;\">%s</td>\r\n", e.escapeHTML(key)))
			html.WriteString(fmt.Sprintf("<td style=\"padding: 12px 16px; color: #2d3748; font-size: 14px; word-wrap: break-word;\">%s</td>\r\n", e.formatTableValue(value)))
			html.WriteString("</tr>\r\n")
		}
		html.WriteString("</tbody>\r\n")
		html.WriteString("</table>\r\n")
		html.WriteString("</div>\r\n")
	}

	// 分隔线
	html.WriteString("<hr style=\"border: none; border-top: 1px solid #e2e8f0; margin: 24px 0;\">\r\n")

	// 元信息区域
	html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\">\r\n")

	// 时间戳
	if !message.Timestamp.IsZero() {
		html.WriteString("<tr>\r\n")
		html.WriteString("<td style=\"padding: 8px 0;\">\r\n")
		html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\">\r\n")
		html.WriteString("<tr>\r\n")
		html.WriteString("<td width=\"80\" style=\"color: #718096; font-size: 14px; font-weight: 500; vertical-align: top;\">告警时间:</td>\r\n")
		html.WriteString(fmt.Sprintf("<td style=\"color: #2d3748; font-size: 14px;\">%s</td>\r\n", message.Timestamp.Format("2006-01-02 15:04:05")))
		html.WriteString("</tr>\r\n")
		html.WriteString("</table>\r\n")
		html.WriteString("</td>\r\n")
		html.WriteString("</tr>\r\n")
	}

	// 标签
	if len(message.Tags) > 0 {
		html.WriteString("<tr>\r\n")
		html.WriteString("<td style=\"padding: 8px 0;\">\r\n")
		html.WriteString("<table role=\"presentation\" width=\"100%\" cellpadding=\"0\" cellspacing=\"0\">\r\n")
		html.WriteString("<tr>\r\n")
		html.WriteString("<td width=\"80\" style=\"color: #718096; font-size: 14px; font-weight: 500; vertical-align: top; padding-top: 4px;\">标签:</td>\r\n")
		html.WriteString("<td>\r\n")
		for k, v := range message.Tags {
			// 根据标签值设置颜色
			tagColor := e.getTagColor(v)
			html.WriteString(fmt.Sprintf("<span style=\"display: inline-block; background-color: %s; color: #ffffff; padding: 6px 12px; margin: 4px 4px 4px 0; border-radius: 16px; font-size: 13px; font-weight: 500; line-height: 1;\"><strong>%s:</strong> %s</span>\r\n", tagColor, e.escapeHTML(k), e.escapeHTML(v)))
		}
		html.WriteString("</td>\r\n")
		html.WriteString("</tr>\r\n")
		html.WriteString("</table>\r\n")
		html.WriteString("</td>\r\n")
		html.WriteString("</tr>\r\n")
	}

	html.WriteString("</table>\r\n") // 元信息表格结束
	html.WriteString("</td>\r\n")
	html.WriteString("</tr>\r\n")

	// 底部
	html.WriteString("<tr>\r\n")
	html.WriteString("<td style=\"background-color: #f8f9fa; padding: 16px 30px; text-align: center; border-top: 1px solid #e2e8f0;\">\r\n")
	html.WriteString("<p style=\"margin: 0; color: #a0aec0; font-size: 12px;\">系统自动发送，请勿回复</p>\r\n")
	html.WriteString("</td>\r\n")
	html.WriteString("</tr>\r\n")

	html.WriteString("</table>\r\n") // 主内容卡片结束
	html.WriteString("</td>\r\n")
	html.WriteString("</tr>\r\n")
	html.WriteString("</table>\r\n") // 外层容器结束

	html.WriteString("</body>\r\n")
	html.WriteString("</html>\r\n")

	return html.String()
}

// escapeHTML 转义HTML特殊字符
func (e *EmailChannel) escapeHTML(text string) string {
	text = strings.ReplaceAll(text, "&", "&amp;")
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	text = strings.ReplaceAll(text, "\"", "&quot;")
	text = strings.ReplaceAll(text, "'", "&#39;")
	return text
}

// getTagColor 根据标签值获取颜色
func (e *EmailChannel) getTagColor(value string) string {
	valueLower := strings.ToLower(value)
	switch valueLower {
	case "error", "critical", "fatal", "danger":
		return "#e53e3e" // 红色
	case "warning", "warn":
		return "#ed8936" // 橙色
	case "info", "information":
		return "#3182ce" // 蓝色
	case "success", "ok", "healthy":
		return "#38a169" // 绿色
	default:
		return "#718096" // 灰色（默认）
	}
}

// formatTableValue 格式化表格值
func (e *EmailChannel) formatTableValue(value interface{}) string {
	if value == nil {
		return "<span style=\"color: #a0aec0; font-style: italic;\">(空)</span>"
	}

	switch v := value.(type) {
	case string:
		return e.escapeHTML(v)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "<span style=\"color: #38a169; font-weight: 500;\">是</span>"
		}
		return "<span style=\"color: #e53e3e; font-weight: 500;\">否</span>"
	default:
		return e.escapeHTML(fmt.Sprintf("%v", v))
	}
}

// Type 返回渠道类型
func (e *EmailChannel) Type() alert.AlertType {
	return e.channelType
}

// Name 返回渠道名称
func (e *EmailChannel) Name() string {
	return e.name
}

// IsEnabled 检查渠道是否启用
func (e *EmailChannel) IsEnabled() bool {
	return e.enabled
}

// Enable 启用渠道
func (e *EmailChannel) Enable() error {
	e.enabled = true
	return nil
}

// Disable 禁用渠道
func (e *EmailChannel) Disable() error {
	e.enabled = false
	return nil
}

// Close 关闭渠道
func (e *EmailChannel) Close() error {
	// 邮件渠道无需特殊清理
	return nil
}

// HealthCheck 健康检查
func (e *EmailChannel) HealthCheck(ctx context.Context) *alert.HealthCheckResult {
	startTime := time.Now()
	result := &alert.HealthCheckResult{
		Timestamp: startTime,
		Extra:     make(map[string]interface{}),
	}

	// 尝试连接SMTP服务器
	addr := net.JoinHostPort(e.serverConfig.SMTPHost, fmt.Sprintf("%d", e.serverConfig.SMTPPort))

	timeout := 10 * time.Second
	if e.serverConfig.Timeout > 0 {
		timeout = time.Duration(e.serverConfig.Timeout) * time.Second
	}

	conn, err := net.DialTimeout("tcp", addr, timeout)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("无法连接到SMTP服务器: %w", err)
		result.Message = fmt.Sprintf("SMTP服务器连接失败: %s", err.Error())
		return result
	}
	defer conn.Close()

	result.Success = true
	result.Message = "SMTP服务器连接正常"
	return result
}
