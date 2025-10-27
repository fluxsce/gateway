package alert

import (
	"context"
	"fmt"
	"sync"
)

// Manager 告警管理器
// 负责管理多个告警渠道，提供统一的告警发送接口
type Manager struct {
	// channels 渠道映射，key为渠道名称
	channels map[string]Channel
	// mutex 读写互斥锁，保证并发安全
	mutex sync.RWMutex
	// defaultChannel 默认渠道名称
	defaultChannel string
}

// NewManager 创建新的告警管理器
func NewManager() *Manager {
	return &Manager{
		channels: make(map[string]Channel),
	}
}

// AddChannel 添加告警渠道
// 参数:
//
//	name: 渠道名称
//	channel: 渠道实例
//
// 返回:
//
//	error: 如果渠道名称已存在则返回错误
func (m *Manager) AddChannel(name string, channel Channel) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 防止添加nil渠道
	if channel == nil {
		return fmt.Errorf("不能添加nil渠道 '%s'", name)
	}

	// 检查是否已存在同名渠道
	if _, exists := m.channels[name]; exists {
		return fmt.Errorf("渠道 '%s' 已存在", name)
	}

	// 添加渠道
	m.channels[name] = channel

	// 如果是第一个渠道，设置为默认渠道
	if len(m.channels) == 1 {
		m.defaultChannel = name
	}

	return nil
}

// GetChannel 获取告警渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	Channel: 渠道实例，如果不存在则返回nil
func (m *Manager) GetChannel(name string) Channel {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.channels[name]
}

// RemoveChannel 移除告警渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 移除过程中的错误
func (m *Manager) RemoveChannel(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 检查渠道是否存在
	channel, exists := m.channels[name]
	if !exists {
		return fmt.Errorf("渠道 '%s' 不存在", name)
	}

	// 关闭渠道
	if err := channel.Close(); err != nil {
		return fmt.Errorf("关闭渠道 '%s' 失败: %w", name, err)
	}

	// 从映射中删除
	delete(m.channels, name)

	// 如果删除的是默认渠道，清空默认渠道
	if m.defaultChannel == name {
		m.defaultChannel = ""
		// 如果还有其他渠道，选择第一个作为默认渠道
		for n := range m.channels {
			m.defaultChannel = n
			break
		}
	}

	return nil
}

// ListChannels 列出所有渠道名称
// 返回:
//
//	[]string: 渠道名称列表
func (m *Manager) ListChannels() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.channels))
	for name := range m.channels {
		names = append(names, name)
	}
	return names
}

// SetDefaultChannel 设置默认渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 如果渠道不存在则返回错误
func (m *Manager) SetDefaultChannel(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.channels[name]; !exists {
		return fmt.Errorf("渠道 '%s' 不存在", name)
	}

	m.defaultChannel = name
	return nil
}

// GetDefaultChannel 获取默认渠道
// 返回:
//
//	Channel: 默认渠道实例，如果不存在则返回nil
func (m *Manager) GetDefaultChannel() Channel {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.defaultChannel == "" {
		return nil
	}
	return m.channels[m.defaultChannel]
}

// Send 通过指定渠道发送告警
// 参数:
//
//	ctx: 上下文
//	channelName: 渠道名称
//	message: 告警消息
//	options: 发送选项
//
// 返回:
//
//	*SendResult: 发送结果
func (m *Manager) Send(ctx context.Context, channelName string, message *Message, options *SendOptions) *SendResult {
	channel := m.GetChannel(channelName)
	if channel == nil {
		return &SendResult{
			Success: false,
			Error:   fmt.Errorf("渠道 '%s' 不存在", channelName),
		}
	}

	return channel.Send(ctx, message, options)
}

// SendToDefault 通过默认渠道发送告警
// 参数:
//
//	ctx: 上下文
//	message: 告警消息
//	options: 发送选项
//
// 返回:
//
//	*SendResult: 发送结果
func (m *Manager) SendToDefault(ctx context.Context, message *Message, options *SendOptions) *SendResult {
	channel := m.GetDefaultChannel()
	if channel == nil {
		return &SendResult{
			Success: false,
			Error:   fmt.Errorf("没有可用的默认渠道"),
		}
	}

	return channel.Send(ctx, message, options)
}

// SendToAll 通过所有启用的渠道发送告警
// 参数:
//
//	ctx: 上下文
//	message: 告警消息
//	options: 发送选项
//
// 返回:
//
//	map[string]*SendResult: 各渠道的发送结果
func (m *Manager) SendToAll(ctx context.Context, message *Message, options *SendOptions) map[string]*SendResult {
	m.mutex.RLock()
	channels := make(map[string]Channel, len(m.channels))
	for name, channel := range m.channels {
		channels[name] = channel
	}
	m.mutex.RUnlock()

	results := make(map[string]*SendResult)
	var wg sync.WaitGroup

	for name, channel := range channels {
		// 跳过未启用的渠道
		if !channel.IsEnabled() {
			continue
		}

		wg.Add(1)
		go func(n string, ch Channel) {
			defer wg.Done()
			results[n] = ch.Send(ctx, message, options)
		}(name, channel)
	}

	wg.Wait()
	return results
}

// SendToMultiple 通过多个指定渠道发送告警
// 参数:
//
//	ctx: 上下文
//	channelNames: 渠道名称列表
//	message: 告警消息
//	options: 发送选项
//
// 返回:
//
//	map[string]*SendResult: 各渠道的发送结果
func (m *Manager) SendToMultiple(ctx context.Context, channelNames []string, message *Message, options *SendOptions) map[string]*SendResult {
	results := make(map[string]*SendResult)
	var wg sync.WaitGroup

	for _, name := range channelNames {
		channel := m.GetChannel(name)
		if channel == nil {
			results[name] = &SendResult{
				Success: false,
				Error:   fmt.Errorf("渠道 '%s' 不存在", name),
			}
			continue
		}

		wg.Add(1)
		go func(n string, ch Channel) {
			defer wg.Done()
			results[n] = ch.Send(ctx, message, options)
		}(name, channel)
	}

	wg.Wait()
	return results
}

// CloseAll 关闭所有渠道
// 返回:
//
//	error: 第一个遇到的错误，如果没有错误则返回nil
func (m *Manager) CloseAll() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var firstErr error
	for name, channel := range m.channels {
		if err := channel.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("关闭渠道 '%s' 失败: %w", name, err)
		}
	}

	// 清空渠道映射
	m.channels = make(map[string]Channel)
	m.defaultChannel = ""
	return firstErr
}

// Stats 获取所有渠道的统计信息
// 返回:
//
//	map[string]map[string]interface{}: 各渠道的统计信息
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]map[string]interface{})
	for name, channel := range m.channels {
		stats[name] = channel.Stats()
	}
	return stats
}

// HealthCheck 对所有渠道进行健康检查
// 参数:
//
//	ctx: 上下文
//
// 返回:
//
//	map[string]error: 各渠道的健康检查结果
func (m *Manager) HealthCheck(ctx context.Context) map[string]error {
	m.mutex.RLock()
	channels := make(map[string]Channel, len(m.channels))
	for name, channel := range m.channels {
		channels[name] = channel
	}
	m.mutex.RUnlock()

	results := make(map[string]error)
	var wg sync.WaitGroup
	var resultMutex sync.Mutex

	for name, channel := range channels {
		wg.Add(1)
		go func(n string, ch Channel) {
			defer wg.Done()
			err := ch.HealthCheck(ctx)
			resultMutex.Lock()
			results[n] = err
			resultMutex.Unlock()
		}(name, channel)
	}

	wg.Wait()
	return results
}

// HasChannel 检查渠道是否存在
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	bool: 如果存在返回true，否则返回false
func (m *Manager) HasChannel(name string) bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	_, exists := m.channels[name]
	return exists
}

// EnableChannel 启用渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 启用过程中的错误
func (m *Manager) EnableChannel(name string) error {
	channel := m.GetChannel(name)
	if channel == nil {
		return fmt.Errorf("渠道 '%s' 不存在", name)
	}
	return channel.Enable()
}

// DisableChannel 禁用渠道
// 参数:
//
//	name: 渠道名称
//
// 返回:
//
//	error: 禁用过程中的错误
func (m *Manager) DisableChannel(name string) error {
	channel := m.GetChannel(name)
	if channel == nil {
		return fmt.Errorf("渠道 '%s' 不存在", name)
	}
	return channel.Disable()
}

// 全局告警管理器实例
var (
	globalManager *Manager
	once          sync.Once
)

// GetGlobalManager 获取全局告警管理器实例
// 使用单例模式确保全局只有一个告警管理器
// 返回:
//
//	*Manager: 全局告警管理器实例
func GetGlobalManager() *Manager {
	once.Do(func() {
		globalManager = NewManager()
	})
	return globalManager
}
