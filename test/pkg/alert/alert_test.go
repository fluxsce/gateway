package alert_test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gateway/pkg/alert"
)

// =============================================================================
// Mock Channel Implementation
// =============================================================================

// MockChannel 模拟告警渠道，用于测试
type MockChannel struct {
	*alert.BaseChannel
	sendFunc        func(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult
	healthCheckFunc func(ctx context.Context) error
	closeFunc       func() error
	sendCount       int
	mutex           sync.Mutex
}

// NewMockChannel 创建模拟渠道
func NewMockChannel(name string, channelType alert.AlertType) *MockChannel {
	return &MockChannel{
		BaseChannel: alert.NewBaseChannel(name, channelType),
	}
}

// Send 发送消息
func (m *MockChannel) Send(ctx context.Context, message *alert.Message, options *alert.SendOptions) *alert.SendResult {
	m.mutex.Lock()
	m.sendCount++
	m.mutex.Unlock()

	if m.sendFunc != nil {
		return m.sendFunc(ctx, message, options)
	}

	// 默认成功
	result := &alert.SendResult{
		Success:   true,
		Timestamp: time.Now(),
		Duration:  10 * time.Millisecond,
	}
	m.UpdateStats(result)
	return result
}

// HealthCheck 健康检查
func (m *MockChannel) HealthCheck(ctx context.Context) error {
	if m.healthCheckFunc != nil {
		return m.healthCheckFunc(ctx)
	}
	return nil
}

// Close 关闭渠道
func (m *MockChannel) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

// GetSendCount 获取发送次数
func (m *MockChannel) GetSendCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.sendCount
}

// =============================================================================
// Manager Tests
// =============================================================================

func TestNewManager(t *testing.T) {
	manager := alert.NewManager()
	if manager == nil {
		t.Fatal("NewManager() 返回 nil")
	}

	channels := manager.ListChannels()
	if len(channels) != 0 {
		t.Errorf("新管理器应该没有渠道，实际有 %d 个", len(channels))
	}
}

func TestManager_AddChannel(t *testing.T) {
	manager := alert.NewManager()

	t.Run("添加有效渠道", func(t *testing.T) {
		channel := NewMockChannel("test1", alert.AlertTypeEmail)
		err := manager.AddChannel("test1", channel)
		if err != nil {
			t.Errorf("添加渠道失败: %v", err)
		}

		// 验证渠道已添加
		if !manager.HasChannel("test1") {
			t.Error("渠道未成功添加")
		}

		// 验证自动设置为默认渠道
		defaultCh := manager.GetDefaultChannel()
		if defaultCh == nil || defaultCh.Name() != "test1" {
			t.Error("第一个渠道应该自动设置为默认渠道")
		}
	})

	t.Run("添加nil渠道", func(t *testing.T) {
		err := manager.AddChannel("nil_channel", nil)
		if err == nil {
			t.Error("添加nil渠道应该返回错误")
		}
	})

	t.Run("添加重复渠道", func(t *testing.T) {
		channel := NewMockChannel("test1", alert.AlertTypeEmail)
		err := manager.AddChannel("test1", channel)
		if err == nil {
			t.Error("添加重复渠道应该返回错误")
		}
	})
}

func TestManager_GetChannel(t *testing.T) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	t.Run("获取存在的渠道", func(t *testing.T) {
		ch := manager.GetChannel("test")
		if ch == nil {
			t.Error("应该返回渠道实例")
		}
		if ch.Name() != "test" {
			t.Errorf("渠道名称不匹配，期望 'test'，实际 '%s'", ch.Name())
		}
	})

	t.Run("获取不存在的渠道", func(t *testing.T) {
		ch := manager.GetChannel("nonexistent")
		if ch != nil {
			t.Error("不存在的渠道应该返回 nil")
		}
	})
}

func TestManager_RemoveChannel(t *testing.T) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	t.Run("移除存在的渠道", func(t *testing.T) {
		err := manager.RemoveChannel("test")
		if err != nil {
			t.Errorf("移除渠道失败: %v", err)
		}

		if manager.HasChannel("test") {
			t.Error("渠道应该已被移除")
		}
	})

	t.Run("移除不存在的渠道", func(t *testing.T) {
		err := manager.RemoveChannel("nonexistent")
		if err == nil {
			t.Error("移除不存在的渠道应该返回错误")
		}
	})

	t.Run("移除默认渠道", func(t *testing.T) {
		// 添加两个渠道
		ch1 := NewMockChannel("ch1", alert.AlertTypeEmail)
		ch2 := NewMockChannel("ch2", alert.AlertTypeQQ)
		_ = manager.AddChannel("ch1", ch1)
		_ = manager.AddChannel("ch2", ch2)

		// 设置 ch1 为默认渠道
		_ = manager.SetDefaultChannel("ch1")

		// 移除默认渠道
		err := manager.RemoveChannel("ch1")
		if err != nil {
			t.Errorf("移除默认渠道失败: %v", err)
		}

		// 应该自动选择另一个渠道作为默认渠道
		defaultCh := manager.GetDefaultChannel()
		if defaultCh == nil {
			t.Error("移除默认渠道后应该自动选择另一个渠道")
		}
	})
}

func TestManager_ListChannels(t *testing.T) {
	manager := alert.NewManager()

	// 添加多个渠道
	channels := []string{"ch1", "ch2", "ch3"}
	for _, name := range channels {
		ch := NewMockChannel(name, alert.AlertTypeEmail)
		_ = manager.AddChannel(name, ch)
	}

	list := manager.ListChannels()
	if len(list) != len(channels) {
		t.Errorf("渠道数量不匹配，期望 %d，实际 %d", len(channels), len(list))
	}

	// 验证所有渠道都在列表中
	channelMap := make(map[string]bool)
	for _, name := range list {
		channelMap[name] = true
	}

	for _, name := range channels {
		if !channelMap[name] {
			t.Errorf("渠道 '%s' 不在列表中", name)
		}
	}
}

func TestManager_DefaultChannel(t *testing.T) {
	manager := alert.NewManager()

	t.Run("无渠道时获取默认渠道", func(t *testing.T) {
		ch := manager.GetDefaultChannel()
		if ch != nil {
			t.Error("无渠道时应该返回 nil")
		}
	})

	t.Run("设置默认渠道", func(t *testing.T) {
		ch1 := NewMockChannel("ch1", alert.AlertTypeEmail)
		ch2 := NewMockChannel("ch2", alert.AlertTypeQQ)
		_ = manager.AddChannel("ch1", ch1)
		_ = manager.AddChannel("ch2", ch2)

		err := manager.SetDefaultChannel("ch2")
		if err != nil {
			t.Errorf("设置默认渠道失败: %v", err)
		}

		defaultCh := manager.GetDefaultChannel()
		if defaultCh == nil || defaultCh.Name() != "ch2" {
			t.Error("默认渠道设置不正确")
		}
	})

	t.Run("设置不存在的默认渠道", func(t *testing.T) {
		err := manager.SetDefaultChannel("nonexistent")
		if err == nil {
			t.Error("设置不存在的渠道为默认渠道应该返回错误")
		}
	})
}

func TestManager_EnableDisableChannel(t *testing.T) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	t.Run("禁用渠道", func(t *testing.T) {
		err := manager.DisableChannel("test")
		if err != nil {
			t.Errorf("禁用渠道失败: %v", err)
		}

		ch := manager.GetChannel("test")
		if ch.IsEnabled() {
			t.Error("渠道应该已被禁用")
		}
	})

	t.Run("启用渠道", func(t *testing.T) {
		err := manager.EnableChannel("test")
		if err != nil {
			t.Errorf("启用渠道失败: %v", err)
		}

		ch := manager.GetChannel("test")
		if !ch.IsEnabled() {
			t.Error("渠道应该已被启用")
		}
	})

	t.Run("操作不存在的渠道", func(t *testing.T) {
		err := manager.EnableChannel("nonexistent")
		if err == nil {
			t.Error("启用不存在的渠道应该返回错误")
		}

		err = manager.DisableChannel("nonexistent")
		if err == nil {
			t.Error("禁用不存在的渠道应该返回错误")
		}
	})
}

func TestManager_Send(t *testing.T) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	ctx := context.Background()
	message := &alert.Message{
		Title:     "测试告警",
		Content:   "这是一条测试消息",
		Timestamp: time.Now(),
	}

	t.Run("发送到存在的渠道", func(t *testing.T) {
		result := manager.Send(ctx, "test", message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		}

		if channel.GetSendCount() != 1 {
			t.Errorf("发送次数不正确，期望 1，实际 %d", channel.GetSendCount())
		}
	})

	t.Run("发送到不存在的渠道", func(t *testing.T) {
		result := manager.Send(ctx, "nonexistent", message, nil)
		if result.Success {
			t.Error("发送到不存在的渠道应该失败")
		}
		if result.Error == nil {
			t.Error("应该返回错误信息")
		}
	})
}

func TestManager_SendToDefault(t *testing.T) {
	manager := alert.NewManager()

	ctx := context.Background()
	message := &alert.Message{
		Title:     "测试告警",
		Content:   "这是一条测试消息",
		Timestamp: time.Now(),
	}

	t.Run("无默认渠道时发送", func(t *testing.T) {
		result := manager.SendToDefault(ctx, message, nil)
		if result.Success {
			t.Error("无默认渠道时发送应该失败")
		}
	})

	t.Run("有默认渠道时发送", func(t *testing.T) {
		channel := NewMockChannel("test", alert.AlertTypeEmail)
		_ = manager.AddChannel("test", channel)

		result := manager.SendToDefault(ctx, message, nil)
		if !result.Success {
			t.Errorf("发送失败: %v", result.Error)
		}
	})
}

func TestManager_SendToAll(t *testing.T) {
	manager := alert.NewManager()

	// 添加多个渠道
	ch1 := NewMockChannel("ch1", alert.AlertTypeEmail)
	ch2 := NewMockChannel("ch2", alert.AlertTypeQQ)
	ch3 := NewMockChannel("ch3", alert.AlertTypeWeChatWork)
	_ = manager.AddChannel("ch1", ch1)
	_ = manager.AddChannel("ch2", ch2)
	_ = manager.AddChannel("ch3", ch3)

	// 禁用其中一个渠道
	_ = manager.DisableChannel("ch2")

	ctx := context.Background()
	message := &alert.Message{
		Title:     "测试告警",
		Content:   "这是一条测试消息",
		Timestamp: time.Now(),
	}

	results := manager.SendToAll(ctx, message, nil)

	// 应该只发送到启用的渠道
	if len(results) != 2 {
		t.Errorf("结果数量不正确，期望 2（只发送到启用的渠道），实际 %d", len(results))
	}

	// 验证发送到了正确的渠道
	if _, ok := results["ch1"]; !ok {
		t.Error("应该发送到 ch1")
	}
	if _, ok := results["ch3"]; !ok {
		t.Error("应该发送到 ch3")
	}
	if _, ok := results["ch2"]; ok {
		t.Error("不应该发送到已禁用的 ch2")
	}
}

func TestManager_SendToMultiple(t *testing.T) {
	manager := alert.NewManager()

	// 添加多个渠道
	ch1 := NewMockChannel("ch1", alert.AlertTypeEmail)
	ch2 := NewMockChannel("ch2", alert.AlertTypeQQ)
	ch3 := NewMockChannel("ch3", alert.AlertTypeWeChatWork)
	_ = manager.AddChannel("ch1", ch1)
	_ = manager.AddChannel("ch2", ch2)
	_ = manager.AddChannel("ch3", ch3)

	ctx := context.Background()
	message := &alert.Message{
		Title:     "测试告警",
		Content:   "这是一条测试消息",
		Timestamp: time.Now(),
	}

	t.Run("发送到指定的多个渠道", func(t *testing.T) {
		results := manager.SendToMultiple(ctx, []string{"ch1", "ch3"}, message, nil)

		if len(results) != 2 {
			t.Errorf("结果数量不正确，期望 2，实际 %d", len(results))
		}

		if !results["ch1"].Success {
			t.Error("ch1 发送应该成功")
		}
		if !results["ch3"].Success {
			t.Error("ch3 发送应该成功")
		}
	})

	t.Run("发送到包含不存在的渠道", func(t *testing.T) {
		results := manager.SendToMultiple(ctx, []string{"ch1", "nonexistent"}, message, nil)

		if len(results) != 2 {
			t.Errorf("结果数量不正确，期望 2，实际 %d", len(results))
		}

		if !results["ch1"].Success {
			t.Error("ch1 发送应该成功")
		}
		if results["nonexistent"].Success {
			t.Error("不存在的渠道发送应该失败")
		}
	})
}

func TestManager_Stats(t *testing.T) {
	manager := alert.NewManager()

	ch1 := NewMockChannel("ch1", alert.AlertTypeEmail)
	ch2 := NewMockChannel("ch2", alert.AlertTypeQQ)
	_ = manager.AddChannel("ch1", ch1)
	_ = manager.AddChannel("ch2", ch2)

	stats := manager.Stats()

	if len(stats) != 2 {
		t.Errorf("统计信息数量不正确，期望 2，实际 %d", len(stats))
	}

	if _, ok := stats["ch1"]; !ok {
		t.Error("应该包含 ch1 的统计信息")
	}
	if _, ok := stats["ch2"]; !ok {
		t.Error("应该包含 ch2 的统计信息")
	}
}

func TestManager_HealthCheck(t *testing.T) {
	manager := alert.NewManager()

	// 创建一个健康的渠道和一个不健康的渠道
	healthyChannel := NewMockChannel("healthy", alert.AlertTypeEmail)
	unhealthyChannel := NewMockChannel("unhealthy", alert.AlertTypeQQ)
	unhealthyChannel.healthCheckFunc = func(ctx context.Context) error {
		return errors.New("健康检查失败")
	}

	_ = manager.AddChannel("healthy", healthyChannel)
	_ = manager.AddChannel("unhealthy", unhealthyChannel)

	ctx := context.Background()
	results := manager.HealthCheck(ctx)

	if len(results) != 2 {
		t.Errorf("健康检查结果数量不正确，期望 2，实际 %d", len(results))
	}

	if results["healthy"] != nil {
		t.Errorf("健康渠道的检查应该成功，实际错误: %v", results["healthy"])
	}

	if results["unhealthy"] == nil {
		t.Error("不健康渠道的检查应该失败")
	}
}

func TestManager_CloseAll(t *testing.T) {
	manager := alert.NewManager()

	closeCount := 0
	var mutex sync.Mutex

	// 添加多个渠道
	for i := 0; i < 3; i++ {
		ch := NewMockChannel(fmt.Sprintf("ch%d", i), alert.AlertTypeEmail)
		ch.closeFunc = func() error {
			mutex.Lock()
			closeCount++
			mutex.Unlock()
			return nil
		}
		_ = manager.AddChannel(fmt.Sprintf("ch%d", i), ch)
	}

	err := manager.CloseAll()
	if err != nil {
		t.Errorf("关闭所有渠道失败: %v", err)
	}

	if closeCount != 3 {
		t.Errorf("关闭次数不正确，期望 3，实际 %d", closeCount)
	}

	// 验证所有渠道已被移除
	channels := manager.ListChannels()
	if len(channels) != 0 {
		t.Errorf("关闭后应该没有渠道，实际有 %d 个", len(channels))
	}
}

// =============================================================================
// BaseChannel Tests
// =============================================================================

func TestBaseChannel(t *testing.T) {
	t.Run("创建基础渠道", func(t *testing.T) {
		base := alert.NewBaseChannel("test", alert.AlertTypeEmail)
		if base == nil {
			t.Fatal("NewBaseChannel() 返回 nil")
		}

		if base.Name() != "test" {
			t.Errorf("名称不匹配，期望 'test'，实际 '%s'", base.Name())
		}

		if base.Type() != alert.AlertTypeEmail {
			t.Errorf("类型不匹配，期望 '%s'，实际 '%s'", alert.AlertTypeEmail, base.Type())
		}

		if !base.IsEnabled() {
			t.Error("新创建的渠道应该是启用状态")
		}
	})

	t.Run("启用和禁用", func(t *testing.T) {
		base := alert.NewBaseChannel("test", alert.AlertTypeEmail)

		err := base.Disable()
		if err != nil {
			t.Errorf("禁用失败: %v", err)
		}
		if base.IsEnabled() {
			t.Error("渠道应该已被禁用")
		}

		err = base.Enable()
		if err != nil {
			t.Errorf("启用失败: %v", err)
		}
		if !base.IsEnabled() {
			t.Error("渠道应该已被启用")
		}
	})

	t.Run("统计信息", func(t *testing.T) {
		base := alert.NewBaseChannel("test", alert.AlertTypeEmail)

		// 初始统计
		stats := base.Stats()
		if stats["total_sent"].(int64) != 0 {
			t.Error("初始发送数应该为 0")
		}

		// 更新统计
		result := &alert.SendResult{
			Success:   true,
			Timestamp: time.Now(),
			Duration:  100 * time.Millisecond,
		}
		base.UpdateStats(result)

		stats = base.Stats()
		if stats["total_sent"].(int64) != 1 {
			t.Errorf("发送数不正确，期望 1，实际 %d", stats["total_sent"])
		}
		if stats["total_success"].(int64) != 1 {
			t.Errorf("成功数不正确，期望 1，实际 %d", stats["total_success"])
		}
	})
}

// =============================================================================
// Message Tests
// =============================================================================

func TestMessage(t *testing.T) {
	t.Run("创建基本消息", func(t *testing.T) {
		msg := &alert.Message{
			Title:     "测试标题",
			Content:   "测试内容",
			Timestamp: time.Now(),
		}

		if msg.Title != "测试标题" {
			t.Error("标题设置不正确")
		}
		if msg.Content != "测试内容" {
			t.Error("内容设置不正确")
		}
	})

	t.Run("带标签的消息", func(t *testing.T) {
		msg := &alert.Message{
			Title:     "测试标题",
			Content:   "测试内容",
			Timestamp: time.Now(),
			Tags: map[string]string{
				"severity": "high",
				"service":  "api",
			},
		}

		if len(msg.Tags) != 2 {
			t.Errorf("标签数量不正确，期望 2，实际 %d", len(msg.Tags))
		}
		if msg.Tags["severity"] != "high" {
			t.Error("标签值不正确")
		}
	})

	t.Run("带额外数据的消息", func(t *testing.T) {
		msg := &alert.Message{
			Title:     "测试标题",
			Content:   "测试内容",
			Timestamp: time.Now(),
			Extra: map[string]interface{}{
				"custom_field": "custom_value",
				"count":        42,
			},
		}

		if len(msg.Extra) != 2 {
			t.Errorf("额外数据数量不正确，期望 2，实际 %d", len(msg.Extra))
		}
	})
}

// =============================================================================
// SendOptions Tests
// =============================================================================

func TestSendOptions(t *testing.T) {
	t.Run("默认选项", func(t *testing.T) {
		opts := alert.DefaultSendOptions()

		if opts.Timeout != 30*time.Second {
			t.Errorf("默认超时时间不正确，期望 30s，实际 %v", opts.Timeout)
		}
		if opts.Retry != 3 {
			t.Errorf("默认重试次数不正确，期望 3，实际 %d", opts.Retry)
		}
		if opts.RetryInterval != 5*time.Second {
			t.Errorf("默认重试间隔不正确，期望 5s，实际 %v", opts.RetryInterval)
		}
		if opts.Async {
			t.Error("默认应该是同步发送")
		}
	})

	t.Run("自定义选项", func(t *testing.T) {
		opts := &alert.SendOptions{
			Timeout:       10 * time.Second,
			Retry:         5,
			RetryInterval: 2 * time.Second,
			Async:         true,
		}

		if opts.Timeout != 10*time.Second {
			t.Error("自定义超时时间设置不正确")
		}
		if opts.Retry != 5 {
			t.Error("自定义重试次数设置不正确")
		}
		if !opts.Async {
			t.Error("自定义异步选项设置不正确")
		}
	})
}

// =============================================================================
// MessageBuilder Tests
// =============================================================================

func TestMessageBuilder(t *testing.T) {
	t.Run("构建基本消息", func(t *testing.T) {
		msg := alert.NewMessage().
			WithTitle("测试标题").
			WithContent("测试内容").
			Build()

		if msg.Title != "测试标题" {
			t.Error("标题设置不正确")
		}
		if msg.Content != "测试内容" {
			t.Error("内容设置不正确")
		}
	})

	t.Run("构建带标签的消息", func(t *testing.T) {
		msg := alert.NewMessage().
			WithTitle("测试标题").
			WithContent("测试内容").
			WithTag("severity", "high").
			WithTag("service", "api").
			Build()

		if len(msg.Tags) != 2 {
			t.Errorf("标签数量不正确，期望 2，实际 %d", len(msg.Tags))
		}
		if msg.Tags["severity"] != "high" {
			t.Error("标签值不正确")
		}
	})

	t.Run("构建带多个标签的消息", func(t *testing.T) {
		tags := map[string]string{
			"severity": "high",
			"service":  "api",
			"env":      "prod",
		}

		msg := alert.NewMessage().
			WithTitle("测试标题").
			WithContent("测试内容").
			WithTags(tags).
			Build()

		if len(msg.Tags) != 3 {
			t.Errorf("标签数量不正确，期望 3，实际 %d", len(msg.Tags))
		}
	})

	t.Run("构建带额外数据的消息", func(t *testing.T) {
		msg := alert.NewMessage().
			WithTitle("测试标题").
			WithContent("测试内容").
			WithExtra("custom_field", "custom_value").
			WithExtra("count", 42).
			Build()

		if len(msg.Extra) != 2 {
			t.Errorf("额外数据数量不正确，期望 2，实际 %d", len(msg.Extra))
		}
		if msg.Extra["count"] != 42 {
			t.Error("额外数据值不正确")
		}
	})

	t.Run("链式调用", func(t *testing.T) {
		msg := alert.NewMessage().
			WithTitle("测试标题").
			WithContent("测试内容").
			WithTag("severity", "high").
			WithExtra("count", 42).
			Build()

		if msg.Title != "测试标题" {
			t.Error("标题设置不正确")
		}
		if msg.Content != "测试内容" {
			t.Error("内容设置不正确")
		}
		if len(msg.Tags) != 1 {
			t.Error("标签数量不正确")
		}
		if len(msg.Extra) != 1 {
			t.Error("额外数据数量不正确")
		}
	})
}

// =============================================================================
// Global API Tests
// =============================================================================

func TestGlobalAPI(t *testing.T) {
	// 注意：全局API使用单例模式，测试之间可能相互影响
	// 在实际项目中，可能需要提供重置全局管理器的方法用于测试

	t.Run("获取全局管理器", func(t *testing.T) {
		manager1 := alert.GetGlobalManager()
		manager2 := alert.GetGlobalManager()

		if manager1 != manager2 {
			t.Error("全局管理器应该是单例")
		}
	})

	t.Run("全局渠道管理", func(t *testing.T) {
		// 清理可能存在的渠道
		for _, name := range alert.ListChannels() {
			_ = alert.RemoveChannel(name)
		}

		// 添加渠道
		channel := NewMockChannel("global_test", alert.AlertTypeEmail)
		err := alert.AddChannel("global_test", channel)
		if err != nil {
			t.Errorf("添加全局渠道失败: %v", err)
		}

		// 检查渠道
		if !alert.HasChannel("global_test") {
			t.Error("全局渠道未成功添加")
		}

		// 获取渠道
		ch := alert.GetChannel("global_test")
		if ch == nil {
			t.Error("应该能获取全局渠道")
		}

		// 清理
		_ = alert.RemoveChannel("global_test")
	})
}

// =============================================================================
// Concurrent Tests
// =============================================================================

func TestConcurrency(t *testing.T) {
	t.Run("并发添加渠道", func(t *testing.T) {
		manager := alert.NewManager()
		var wg sync.WaitGroup
		errorCount := 0
		var errorMutex sync.Mutex

		// 并发添加渠道
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()
				ch := NewMockChannel(fmt.Sprintf("ch%d", index), alert.AlertTypeEmail)
				err := manager.AddChannel(fmt.Sprintf("ch%d", index), ch)
				if err != nil {
					errorMutex.Lock()
					errorCount++
					errorMutex.Unlock()
				}
			}(i)
		}

		wg.Wait()

		if errorCount > 0 {
			t.Errorf("并发添加渠道出现 %d 个错误", errorCount)
		}

		channels := manager.ListChannels()
		if len(channels) != 10 {
			t.Errorf("渠道数量不正确，期望 10，实际 %d", len(channels))
		}
	})

	t.Run("并发发送消息", func(t *testing.T) {
		manager := alert.NewManager()
		channel := NewMockChannel("test", alert.AlertTypeEmail)
		_ = manager.AddChannel("test", channel)

		var wg sync.WaitGroup
		successCount := 0
		var successMutex sync.Mutex

		ctx := context.Background()
		message := &alert.Message{
			Title:     "并发测试",
			Content:   "测试内容",
			Timestamp: time.Now(),
		}

		// 并发发送
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				result := manager.Send(ctx, "test", message, nil)
				if result.Success {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}()
		}

		wg.Wait()

		if successCount != 100 {
			t.Errorf("成功发送数量不正确，期望 100，实际 %d", successCount)
		}

		if channel.GetSendCount() != 100 {
			t.Errorf("渠道发送次数不正确，期望 100，实际 %d", channel.GetSendCount())
		}
	})
}

// =============================================================================
// Benchmark Tests
// =============================================================================

func BenchmarkManager_Send(b *testing.B) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	ctx := context.Background()
	message := &alert.Message{
		Title:     "基准测试",
		Content:   "测试内容",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.Send(ctx, "test", message, nil)
	}
}

func BenchmarkMessageBuilder(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = alert.NewMessage().
			WithTitle("基准测试").
			WithContent("测试内容").
			WithTag("severity", "high").
			WithExtra("count", i).
			Build()
	}
}

func BenchmarkConcurrentSend(b *testing.B) {
	manager := alert.NewManager()
	channel := NewMockChannel("test", alert.AlertTypeEmail)
	_ = manager.AddChannel("test", channel)

	ctx := context.Background()
	message := &alert.Message{
		Title:     "并发基准测试",
		Content:   "测试内容",
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = manager.Send(ctx, "test", message, nil)
		}
	})
}
