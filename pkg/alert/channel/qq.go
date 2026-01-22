package channel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"gateway/pkg/alert"
)

// QQServerConfig QQ服务器配置（可公用）
type QQServerConfig struct {
	// WebhookURL QQ机器人Webhook地址
	WebhookURL string
	// Secret 签名密钥（可选）
	Secret string
	// Timeout 超时时间（秒）
	Timeout int
}

// Validate 验证服务器配置
func (c *QQServerConfig) Validate() error {
	if c.WebhookURL == "" {
		return fmt.Errorf("QQ机器人Webhook地址不能为空")
	}
	return nil
}

// QQSendConfig QQ发送配置（每次发送可不同）
type QQSendConfig struct {
	// AtAll 是否@所有人
	AtAll bool
	// AtUsers @指定用户列表（QQ号）
	AtUsers []string
}

// Validate 验证发送配置
func (c *QQSendConfig) Validate() error {
	// QQ发送配置无需验证
	return nil
}

// QQChannel QQ告警渠道
type QQChannel struct {
	name         string
	channelType  alert.AlertType
	enabled      bool
	serverConfig *QQServerConfig
	sendConfig   *QQSendConfig
	httpClient   *http.Client
}

// NewQQChannel 创建QQ告警渠道
// 参数:
//
//	name: 渠道名称
//	serverConfig: QQ服务器配置（可公用）
//	sendConfig: 默认发送配置（可在发送时覆盖）
func NewQQChannel(name string, serverConfig *QQServerConfig, sendConfig *QQSendConfig) (*QQChannel, error) {
	if serverConfig == nil {
		return nil, fmt.Errorf("服务器配置不能为空")
	}

	if err := serverConfig.Validate(); err != nil {
		return nil, fmt.Errorf("服务器配置验证失败: %w", err)
	}

	if sendConfig == nil {
		sendConfig = &QQSendConfig{} // 使用默认配置
	}

	timeout := 30 * time.Second
	if serverConfig.Timeout > 0 {
		timeout = time.Duration(serverConfig.Timeout) * time.Second
	}

	channel := &QQChannel{
		name:         name,
		channelType:  alert.AlertTypeQQ,
		enabled:      true,
		serverConfig: serverConfig,
		sendConfig:   sendConfig,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}

	return channel, nil
}

// SetSendConfig 设置或更新发送配置
func (q *QQChannel) SetSendConfig(config *QQSendConfig) error {
	if config == nil {
		return fmt.Errorf("发送配置不能为空")
	}
	q.sendConfig = config
	return nil
}

// qqMessage QQ消息结构
type qqMessage struct {
	MsgType string    `json:"msg_type"`
	Content qqContent `json:"content"`
	At      *qqAtInfo `json:"at,omitempty"`
}

// qqContent QQ消息内容
type qqContent struct {
	Text string `json:"text"`
}

// qqAtInfo @信息
type qqAtInfo struct {
	AtAll   bool     `json:"at_all"`
	AtUsers []string `json:"at_users,omitempty"`
}

// qqResponse QQ响应
type qqResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Send 发送QQ告警
// 可以在 message.Extra 中传入 "send_config" 来覆盖默认的发送配置
func (q *QQChannel) Send(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	startTime := time.Now()
	result := &alert.SendResult{
		Success:   false,
		Timestamp: startTime,
		Extra:     make(map[string]interface{}),
	}

	// 检查渠道是否启用
	if !q.IsEnabled() {
		result.Error = fmt.Errorf("QQ告警渠道未启用")
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
		err := q.sendMessage(sendCtx, message)
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

// sendMessage 实际发送消息的内部方法
func (q *QQChannel) sendMessage(ctx context.Context, message *alert.Message) error {
	// 获取发送配置（支持从消息中覆盖）
	sendConfig := q.sendConfig
	if customConfig, ok := message.Extra["send_config"].(*QQSendConfig); ok && customConfig != nil {
		sendConfig = customConfig
	}

	// 构建消息内容
	content := q.buildMessageContent(message)

	// 构建QQ消息
	qqMsg := qqMessage{
		MsgType: "text",
		Content: qqContent{
			Text: content,
		},
	}

	// 添加@信息
	if sendConfig.AtAll || len(sendConfig.AtUsers) > 0 {
		qqMsg.At = &qqAtInfo{
			AtAll:   sendConfig.AtAll,
			AtUsers: sendConfig.AtUsers,
		}
	}

	// 序列化消息
	jsonData, err := json.Marshal(qqMsg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %w", err)
	}

	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", q.serverConfig.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := q.httpClient.Do(req)
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
	var qqResp qqResponse
	if err := json.Unmarshal(body, &qqResp); err != nil {
		// 如果无法解析响应，但HTTP状态码是200，认为发送成功
		return nil
	}

	// 检查业务状态码
	if qqResp.Code != 0 {
		return fmt.Errorf("QQ机器人返回错误，代码: %d, 消息: %s", qqResp.Code, qqResp.Message)
	}

	return nil
}

// buildMessageContent 构建消息内容
func (q *QQChannel) buildMessageContent(message *alert.Message) string {
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

// Type 返回渠道类型
func (q *QQChannel) Type() alert.AlertType {
	return q.channelType
}

// Name 返回渠道名称
func (q *QQChannel) Name() string {
	return q.name
}

// IsEnabled 检查渠道是否启用
func (q *QQChannel) IsEnabled() bool {
	return q.enabled
}

// Enable 启用渠道
func (q *QQChannel) Enable() error {
	q.enabled = true
	return nil
}

// Disable 禁用渠道
func (q *QQChannel) Disable() error {
	q.enabled = false
	return nil
}

// Close 关闭渠道
func (q *QQChannel) Close() error {
	// 关闭HTTP客户端
	if q.httpClient != nil {
		q.httpClient.CloseIdleConnections()
	}
	return nil
}

// HealthCheck 健康检查
func (q *QQChannel) HealthCheck(ctx context.Context) *alert.HealthCheckResult {
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

	err := q.sendMessage(ctx, testMsg)
	result.Duration = time.Since(startTime)

	if err != nil {
		result.Success = false
		result.Error = fmt.Errorf("健康检查失败: %w", err)
		result.Message = fmt.Sprintf("QQ渠道健康检查失败: %s", err.Error())
		return result
	}

	result.Success = true
	result.Message = "QQ渠道健康检查通过"
	return result
}
