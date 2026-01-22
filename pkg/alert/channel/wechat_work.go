package channel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gateway/pkg/alert"
	"gateway/pkg/httpclient"
)

// WeChatWorkServerConfig 企业微信服务器配置（可公用）
type WeChatWorkServerConfig struct {
	// WebhookURL 企业微信机器人Webhook地址
	WebhookURL string
	// Secret 签名密钥（可选，用于安全验证）
	Secret string
	// MessageType 消息类型：text, markdown
	MessageType string
	// Timeout 超时时间（秒）
	Timeout int
	// TitleTemplate 标题模板（可选）
	// 支持占位符：{{title}}, {{timestamp}}, {{tag.key}} 等
	// 如果为空，使用默认格式
	TitleTemplate string
	// ContentTemplate 内容模板（可选）
	// 支持占位符：{{content}}, {{title}}, {{tags}}, {{timestamp}} 等
	// 如果为空，使用默认格式
	ContentTemplate string
}

// Validate 验证服务器配置
func (c *WeChatWorkServerConfig) Validate() error {
	if c.WebhookURL == "" {
		return fmt.Errorf("企业微信Webhook地址不能为空")
	}
	if c.MessageType == "" {
		c.MessageType = "markdown" // 默认使用markdown格式
	}
	if c.MessageType != "text" && c.MessageType != "markdown" {
		return fmt.Errorf("消息类型必须是text或markdown")
	}
	return nil
}

// WeChatWorkSendConfig 企业微信发送配置（每次发送可不同）
type WeChatWorkSendConfig struct {
	// MentionedList @指定成员列表（userid）
	MentionedList []string
	// MentionedMobileList @指定成员手机号列表
	MentionedMobileList []string
}

// Validate 验证发送配置
func (c *WeChatWorkSendConfig) Validate() error {
	// 企业微信发送配置无需验证
	return nil
}

// WeChatWorkChannel 企业微信告警渠道
type WeChatWorkChannel struct {
	name             string
	channelType      alert.AlertType
	enabled          bool
	serverConfig     *WeChatWorkServerConfig
	sendConfig       *WeChatWorkSendConfig
	httpClient       httpclient.Client
	templateReplacer *TemplateReplacer
}

// NewWeChatWorkChannel 创建企业微信告警渠道
// 参数:
//
//	name: 渠道名称
//	serverConfig: 企业微信服务器配置（可公用）
//	sendConfig: 默认发送配置（可在发送时覆盖）
//	httpClient: HTTP客户端（可选，如果为nil则创建默认客户端）
func NewWeChatWorkChannel(name string, serverConfig *WeChatWorkServerConfig, sendConfig *WeChatWorkSendConfig, httpClient httpclient.Client) (*WeChatWorkChannel, error) {
	if serverConfig == nil {
		return nil, fmt.Errorf("服务器配置不能为空")
	}

	if err := serverConfig.Validate(); err != nil {
		return nil, fmt.Errorf("服务器配置验证失败: %w", err)
	}

	if sendConfig == nil {
		sendConfig = &WeChatWorkSendConfig{} // 使用默认配置
	}

	// 如果没有提供HTTP客户端，创建默认客户端
	if httpClient == nil {
		timeout := 30 * time.Second
		if serverConfig.Timeout > 0 {
			timeout = time.Duration(serverConfig.Timeout) * time.Second
		}

		client, err := httpclient.NewClient(&httpclient.ClientConfig{
			Timeout: timeout,
		})
		if err != nil {
			return nil, fmt.Errorf("创建HTTP客户端失败: %w", err)
		}
		httpClient = client
	}

	channel := &WeChatWorkChannel{
		name:             name,
		channelType:      alert.AlertTypeWeChatWork,
		enabled:          true,
		serverConfig:     serverConfig,
		sendConfig:       sendConfig,
		httpClient:       httpClient,
		templateReplacer: NewTemplateReplacer(),
	}

	return channel, nil
}

// SetSendConfig 设置或更新发送配置
func (w *WeChatWorkChannel) SetSendConfig(config *WeChatWorkSendConfig) error {
	if config == nil {
		return fmt.Errorf("发送配置不能为空")
	}
	w.sendConfig = config
	return nil
}

// wechatMessage 企业微信消息结构
type wechatMessage struct {
	MsgType  string                 `json:"msgtype"`
	Text     *wechatTextContent     `json:"text,omitempty"`
	Markdown *wechatMarkdownContent `json:"markdown,omitempty"`
}

// wechatTextContent 文本消息内容
type wechatTextContent struct {
	Content             string   `json:"content"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// wechatMarkdownContent Markdown消息内容
type wechatMarkdownContent struct {
	Content string `json:"content"`
}

// wechatResponse 企业微信响应
type wechatResponse struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

// Send 发送企业微信告警
// 可以在 message.Extra 中传入 "send_config" 来覆盖默认的发送配置
func (w *WeChatWorkChannel) Send(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	startTime := time.Now()
	result := &alert.SendResult{
		Success:   false,
		Timestamp: startTime,
		Extra:     make(map[string]interface{}),
	}

	// 检查渠道是否启用
	if !w.IsEnabled() {
		result.Error = fmt.Errorf("企业微信告警渠道未启用")
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
		responseBody, err := w.sendMessage(sendCtx, message)
		if err == nil {
			result.Success = true
			result.Duration = time.Since(startTime)
			// 保存原始响应体到 Extra 中
			if result.Extra == nil {
				result.Extra = make(map[string]interface{})
			}
			result.Extra["response_body"] = responseBody
			return result
		}

		// 保存响应消息（即使失败也保存）
		if result.Extra == nil {
			result.Extra = make(map[string]interface{})
		}
		if responseBody != "" {
			result.Extra["response_body"] = responseBody
		}
		result.Extra["last_error"] = err.Error()

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

// sendMessage 实际发送消息的内部方法
// 参考企业微信官方文档：https://developer.work.weixin.qq.com/document/path/99110
// 返回响应消息（如果成功）和错误（如果失败）
func (w *WeChatWorkChannel) sendMessage(ctx context.Context, message *alert.Message) (string, error) {
	// 获取发送配置（支持从消息中覆盖）
	sendConfig := w.sendConfig
	if customConfig, ok := message.Extra["send_config"].(*WeChatWorkSendConfig); ok && customConfig != nil {
		sendConfig = customConfig
	}

	// 构建企业微信消息
	var wechatMsg wechatMessage

	if w.serverConfig.MessageType == "markdown" {
		wechatMsg.MsgType = "markdown"
		wechatMsg.Markdown = &wechatMarkdownContent{
			Content: w.buildMarkdownContent(message),
		}
	} else {
		wechatMsg.MsgType = "text"
		wechatMsg.Text = &wechatTextContent{
			Content:             w.buildTextContent(message),
			MentionedList:       sendConfig.MentionedList,
			MentionedMobileList: sendConfig.MentionedMobileList,
		}
	}

	// 构建Webhook URL
	webhookURL := w.serverConfig.WebhookURL

	// 如果配置了密钥，添加签名（企业微信官方要求）
	// 参考文档：https://developer.work.weixin.qq.com/document/path/99110
	// 签名算法：HMAC-SHA256，然后Base64编码
	// 签名字符串：timestamp + "\n" + secret
	if w.serverConfig.Secret != "" {
		timestamp := time.Now().Unix()
		sign := w.generateSign(timestamp)
		// 根据企业微信文档，签名参数通过URL查询参数传递
		// 检查URL是否已有查询参数
		separator := "?"
		if bytes.IndexByte([]byte(webhookURL), '?') != -1 {
			separator = "&"
		}
		webhookURL = fmt.Sprintf("%s%stimestamp=%d&sign=%s", webhookURL, separator, timestamp, sign)
	}

	// 使用httpclient发送POST请求
	resp, err := w.httpClient.Post(ctx, webhookURL, wechatMsg,
		httpclient.WithContentType("application/json"),
	)
	if err != nil {
		return "", fmt.Errorf("发送请求失败: %w", err)
	}

	// 检查HTTP状态码
	if !resp.IsSuccess() {
		responseBody := resp.BodyString()
		return responseBody, fmt.Errorf("HTTP请求失败，状态码: %d, 响应: %s", resp.StatusCode, responseBody)
	}

	// 保存原始响应体
	responseBody := resp.BodyString()

	// 解析响应
	var wechatResp wechatResponse
	if err := json.Unmarshal(resp.Body, &wechatResp); err != nil {
		return responseBody, fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查业务状态码（企业微信返回errcode=0表示成功）
	if wechatResp.ErrCode != 0 {
		return responseBody, fmt.Errorf("企业微信返回错误，代码: %d, 消息: %s", wechatResp.ErrCode, wechatResp.ErrMsg)
	}

	return responseBody, nil
}

// buildTextContent 构建文本消息内容
func (w *WeChatWorkChannel) buildTextContent(message *alert.Message) string {
	// 如果配置了内容模板，使用模板替换
	if w.serverConfig.ContentTemplate != "" {
		return w.templateReplacer.Replace(w.serverConfig.ContentTemplate, message, nil)
	}

	// 否则使用默认格式
	var content bytes.Buffer

	// 标题（如果配置了标题模板，使用模板；否则使用默认格式）
	if message.Title != "" {
		if w.serverConfig.TitleTemplate != "" {
			// 有模板时，直接使用模板内容，不添加图标
			content.WriteString(w.templateReplacer.Replace(w.serverConfig.TitleTemplate, message, nil))
		} else {
			// 没有模板时，使用默认格式并添加图标
			content.WriteString(fmt.Sprintf("🚨 【%s】\n", message.Title))
		}
		content.WriteString("\n")
	}

	// 内容
	content.WriteString(message.Content)
	content.WriteString("\n")

	// 表格数据（如果 TableData 不为空，且 DisplayFormat 为 table 或未设置）
	if len(message.TableData) > 0 && (message.DisplayFormat == alert.DisplayFormatTable || message.DisplayFormat == "") {
		content.WriteString("\n")
		content.WriteString("─────────────────\n")
		content.WriteString("数据详情:\n")
		for key, value := range message.TableData {
			content.WriteString(fmt.Sprintf("  • %s: %s\n", key, w.formatTableValue(value)))
		}
	}

	// 标签和时间戳
	if len(message.Tags) > 0 || !message.Timestamp.IsZero() {
		content.WriteString("\n")
		content.WriteString("─────────────────\n")

		// 标签
		if len(message.Tags) > 0 {
			tagParts := make([]string, 0, len(message.Tags))
			for k, v := range message.Tags {
				tagParts = append(tagParts, fmt.Sprintf("%s: %s", k, v))
			}
			content.WriteString(fmt.Sprintf("标签: %s\n", strings.Join(tagParts, " | ")))
		}

		// 时间戳
		if !message.Timestamp.IsZero() {
			content.WriteString(fmt.Sprintf("时间: %s", message.Timestamp.Format("2006-01-02 15:04:05")))
		}
	}

	return content.String()
}

// buildMarkdownContent 构建Markdown消息内容
// 企业微信Markdown支持：标题、加粗、代码块、引用、列表等
func (w *WeChatWorkChannel) buildMarkdownContent(message *alert.Message) string {
	// 如果配置了内容模板，使用模板替换
	if w.serverConfig.ContentTemplate != "" {
		return w.templateReplacer.Replace(w.serverConfig.ContentTemplate, message, nil)
	}

	// 否则使用默认格式
	var content bytes.Buffer

	// 标题（如果配置了标题模板，使用模板；否则使用默认格式）
	if message.Title != "" {
		if w.serverConfig.TitleTemplate != "" {
			// 有模板时，直接使用模板内容，不添加图标
			content.WriteString(w.templateReplacer.Replace(w.serverConfig.TitleTemplate, message, nil))
		} else {
			// 没有模板时，使用默认格式并添加图标
			content.WriteString(fmt.Sprintf("# 🚨 %s\n\n", message.Title))
		}
	}

	// 内容（使用引用块突出显示）
	if message.Content != "" {
		// 处理多行内容，每行都加上引用符号
		lines := strings.Split(message.Content, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				content.WriteString(fmt.Sprintf("> %s\n", line))
			} else {
				content.WriteString(">\n")
			}
		}
		content.WriteString("\n")
	}

	// 表格数据（如果 TableData 不为空，且 DisplayFormat 为 table 或未设置）
	if len(message.TableData) > 0 && (message.DisplayFormat == alert.DisplayFormatTable || message.DisplayFormat == "") {
		content.WriteString("---\n\n")
		content.WriteString("**数据详情**\n\n")
		// 使用 Markdown 表格格式
		content.WriteString("| 字段 | 值 |\n")
		content.WriteString("| :--- | :--- |\n")
		for key, value := range message.TableData {
			content.WriteString(fmt.Sprintf("| **%s** | %s |\n", key, w.formatTableValueForMarkdown(value)))
		}
		content.WriteString("\n")
	}

	// 标签和时间戳（使用表格或列表格式）
	if len(message.Tags) > 0 || !message.Timestamp.IsZero() {
		// 使用分隔线
		content.WriteString("---\n\n")

		// 标签（使用列表格式，更清晰）
		if len(message.Tags) > 0 {
			content.WriteString("**标签信息**\n\n")
			for k, v := range message.Tags {
				// 根据标签值设置颜色提示
				colorTag := ""
				if strings.ToLower(v) == "error" || strings.ToLower(v) == "critical" || strings.ToLower(v) == "fatal" {
					colorTag = `<font color="warning">`
				} else if strings.ToLower(v) == "warning" || strings.ToLower(v) == "warn" {
					colorTag = `<font color="comment">`
				} else if strings.ToLower(v) == "info" || strings.ToLower(v) == "success" {
					colorTag = `<font color="info">`
				}

				if colorTag != "" {
					content.WriteString(fmt.Sprintf("- **%s**: %s%s</font>\n", k, colorTag, v))
				} else {
					content.WriteString(fmt.Sprintf("- **%s**: `%s`\n", k, v))
				}
			}
			content.WriteString("\n")
		}

		// 时间戳
		if !message.Timestamp.IsZero() {
			content.WriteString(fmt.Sprintf("**告警时间**: <font color=\"comment\">%s</font>", message.Timestamp.Format("2006-01-02 15:04:05")))
		}
	}

	return content.String()
}

// formatTableValue 格式化表格值（用于文本格式）
func (w *WeChatWorkChannel) formatTableValue(value interface{}) string {
	if value == nil {
		return "(空)"
	}

	switch v := value.(type) {
	case string:
		return v
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "是"
		}
		return "否"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// formatTableValueForMarkdown 格式化表格值（用于Markdown格式）
func (w *WeChatWorkChannel) formatTableValueForMarkdown(value interface{}) string {
	if value == nil {
		return "<font color=\"comment\">(空)</font>"
	}

	switch v := value.(type) {
	case string:
		// 转义 Markdown 特殊字符
		escaped := strings.ReplaceAll(v, "|", "\\|")
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		return escaped
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v)
	case float32, float64:
		return fmt.Sprintf("%.2f", v)
	case bool:
		if v {
			return "<font color=\"info\">是</font>"
		}
		return "<font color=\"warning\">否</font>"
	default:
		// 转义 Markdown 特殊字符
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, "|", "\\|")
		escaped = strings.ReplaceAll(escaped, "\n", "<br>")
		return escaped
	}
}

// generateSign 生成签名
// 根据企业微信官方文档：https://developer.work.weixin.qq.com/document/path/99110
// 签名算法：
// 1. 将timestamp、secret拼接成字符串：timestamp + "\n" + secret
// 2. 使用HMAC-SHA256算法计算签名
// 3. 对签名进行Base64编码
func (w *WeChatWorkChannel) generateSign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, w.serverConfig.Secret)
	h := hmac.New(sha256.New, []byte(w.serverConfig.Secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}

// Type 返回渠道类型
func (w *WeChatWorkChannel) Type() alert.AlertType {
	return w.channelType
}

// Name 返回渠道名称
func (w *WeChatWorkChannel) Name() string {
	return w.name
}

// IsEnabled 检查渠道是否启用
func (w *WeChatWorkChannel) IsEnabled() bool {
	return w.enabled
}

// Enable 启用渠道
func (w *WeChatWorkChannel) Enable() error {
	w.enabled = true
	return nil
}

// Disable 禁用渠道
func (w *WeChatWorkChannel) Disable() error {
	w.enabled = false
	return nil
}

// Close 关闭渠道
func (w *WeChatWorkChannel) Close() error {
	// 关闭HTTP客户端
	if w.httpClient != nil {
		return w.httpClient.Close()
	}
	return nil
}

// HealthCheck 健康检查
func (w *WeChatWorkChannel) HealthCheck(ctx context.Context) *alert.HealthCheckResult {
	startTime := time.Now()
	result := &alert.HealthCheckResult{
		Timestamp: startTime,
		Extra:     make(map[string]interface{}),
	}

	// 发送测试消息
	testMsg := &alert.Message{
		Title:     "健康检查",
		Content:   "这是一条健康检查消息",
		Timestamp: time.Now(),
	}

	responseBody, err := w.sendMessage(ctx, testMsg)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("健康检查失败: %w", err)
		result.Message = fmt.Sprintf("企业微信健康检查失败: %s", err.Error())
		if responseBody != "" {
			result.Extra["response_body"] = responseBody
		}
		return result
	}

	result.Success = true
	result.Message = "企业微信渠道健康检查通过"
	if responseBody != "" {
		result.Extra["response_body"] = responseBody
	}
	return result
}
