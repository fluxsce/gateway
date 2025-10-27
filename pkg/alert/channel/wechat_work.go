package channel

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/pkg/alert"
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
	*alert.BaseChannel
	serverConfig *WeChatWorkServerConfig
	sendConfig   *WeChatWorkSendConfig
	httpClient   *http.Client
}

// NewWeChatWorkChannel 创建企业微信告警渠道
// 参数:
//
//	name: 渠道名称
//	serverConfig: 企业微信服务器配置（可公用）
//	sendConfig: 默认发送配置（可在发送时覆盖）
func NewWeChatWorkChannel(name string, serverConfig *WeChatWorkServerConfig, sendConfig *WeChatWorkSendConfig) (*WeChatWorkChannel, error) {
	if serverConfig == nil {
		return nil, fmt.Errorf("服务器配置不能为空")
	}

	if err := serverConfig.Validate(); err != nil {
		return nil, fmt.Errorf("服务器配置验证失败: %w", err)
	}

	if sendConfig == nil {
		sendConfig = &WeChatWorkSendConfig{} // 使用默认配置
	}

	timeout := 30 * time.Second
	if serverConfig.Timeout > 0 {
		timeout = time.Duration(serverConfig.Timeout) * time.Second
	}

	channel := &WeChatWorkChannel{
		BaseChannel:  alert.NewBaseChannel(name, alert.AlertTypeWeChatWork),
		serverConfig: serverConfig,
		sendConfig:   sendConfig,
		httpClient: &http.Client{
			Timeout: timeout,
		},
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
		w.UpdateStats(result)
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
			w.UpdateStats(result)
			return result
		default:
		}

		// 尝试发送
		err := w.sendMessage(sendCtx, message)
		if err == nil {
			result.Success = true
			result.Duration = time.Since(startTime)
			w.UpdateStats(result)
			return result
		}

		lastErr = err

		// 如果不是最后一次重试，等待后重试
		if i < maxRetries-1 && options.RetryInterval > 0 {
			select {
			case <-sendCtx.Done():
				result.Error = fmt.Errorf("重试等待期间被取消: %w", sendCtx.Err())
				result.Duration = time.Since(startTime)
				w.UpdateStats(result)
				return result
			case <-time.After(options.RetryInterval):
				// 继续重试
			}
		}
	}

	result.Error = fmt.Errorf("发送失败（重试%d次）: %w", maxRetries, lastErr)
	result.Duration = time.Since(startTime)
	w.UpdateStats(result)
	return result
}

// sendMessage 实际发送消息的内部方法
func (w *WeChatWorkChannel) sendMessage(ctx context.Context, message *alert.Message) error {
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

	// 序列化消息
	jsonData, err := json.Marshal(wechatMsg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 创建HTTP请求
	webhookURL := w.serverConfig.WebhookURL

	// 如果配置了密钥，添加签名
	if w.serverConfig.Secret != "" {
		timestamp := time.Now().Unix()
		sign := w.generateSign(timestamp)
		webhookURL = fmt.Sprintf("%s&timestamp=%d&sign=%s", webhookURL, timestamp, sign)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP请求失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var wechatResp wechatResponse
	if err := json.Unmarshal(body, &wechatResp); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	// 检查业务状态码
	if wechatResp.ErrCode != 0 {
		return fmt.Errorf("企业微信返回错误，代码: %d, 消息: %s", wechatResp.ErrCode, wechatResp.ErrMsg)
	}

	return nil
}

// buildTextContent 构建文本消息内容
func (w *WeChatWorkChannel) buildTextContent(message *alert.Message) string {
	var content bytes.Buffer

	// 标题
	if message.Title != "" {
		content.WriteString(fmt.Sprintf("📢 %s\n", message.Title))
	}

	// 分隔线
	content.WriteString("━━━━━━━━━━━━━━━━\n")

	// 内容
	content.WriteString(message.Content)
	content.WriteString("\n")

	// 标签
	if len(message.Tags) > 0 {
		content.WriteString("━━━━━━━━━━━━━━━━\n")
		content.WriteString("标签:\n")
		for k, v := range message.Tags {
			content.WriteString(fmt.Sprintf("  • %s: %s\n", k, v))
		}
	}

	// 时间戳
	content.WriteString("━━━━━━━━━━━━━━━━\n")
	content.WriteString(fmt.Sprintf("时间: %s", message.Timestamp.Format("2006-01-02 15:04:05")))

	return content.String()
}

// buildMarkdownContent 构建Markdown消息内容
func (w *WeChatWorkChannel) buildMarkdownContent(message *alert.Message) string {
	var content bytes.Buffer

	// 标题
	if message.Title != "" {
		content.WriteString(fmt.Sprintf("# 📢 %s\n\n", message.Title))
	}

	// 内容
	content.WriteString(fmt.Sprintf("> %s\n\n", message.Content))

	// 标签
	if len(message.Tags) > 0 {
		content.WriteString("**标签:**\n")
		for k, v := range message.Tags {
			content.WriteString(fmt.Sprintf("- %s: `%s`\n", k, v))
		}
		content.WriteString("\n")
	}

	// 时间戳
	content.WriteString(fmt.Sprintf("**告警时间:** %s", message.Timestamp.Format("2006-01-02 15:04:05")))

	return content.String()
}

// generateSign 生成签名
func (w *WeChatWorkChannel) generateSign(timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, w.serverConfig.Secret)
	h := hmac.New(sha256.New, []byte(w.serverConfig.Secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature
}

// Close 关闭渠道
func (w *WeChatWorkChannel) Close() error {
	// 关闭HTTP客户端
	if w.httpClient != nil {
		w.httpClient.CloseIdleConnections()
	}
	return nil
}

// HealthCheck 健康检查
func (w *WeChatWorkChannel) HealthCheck(ctx context.Context) error {
	// 发送测试消息
	testMsg := &alert.Message{
		Title:     "健康检查",
		Content:   "这是一条健康检查消息",
		Timestamp: time.Now(),
	}

	err := w.sendMessage(ctx, testMsg)
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}

	return nil
}
