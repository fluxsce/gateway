package subscriber

import (
	"context"
	"sync"
	"time"

	pb "gateway/internal/servicecenter/server/proto"
	"gateway/internal/servicecenter/types"
)

// ConfigWatcher 配置监听管理器
// 管理所有监听配置变更的客户端连接
//
// 功能说明：
//   - 统一使用批量监听机制，支持监听单个或多个配置
//   - 单个配置监听可以看作是批量监听的特例（configDataIds 数组只有一个元素）
//   - 所有配置共用同一个 channel，减少连接数
//
// 数据结构：
//
//	watchers[watcherID][configKey] = channel
//	例如：
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config1"] = ch
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config2"] = ch
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config3"] = ch
//	表示：同一个客户端监听多个配置，所有配置共用同一个 channel
type ConfigWatcher struct {
	mu sync.RWMutex
	// 统一监听：一个 watcherID 可以监听多个配置，所有配置共用同一个 channel
	watchers map[string]map[string]chan *pb.ConfigChangeEvent // key: watcherID -> configKey -> channel (同一个 channel)
}

// NewConfigWatcher 创建配置监听管理器
func NewConfigWatcher() *ConfigWatcher {
	return &ConfigWatcher{
		watchers: make(map[string]map[string]chan *pb.ConfigChangeEvent),
	}
}

// Watch 监听配置变更（统一接口，支持监听单个或多个配置）
//
// 处理流程：
//  1. 创建一个共享的 channel（所有配置共用）
//  2. 为每个配置注册监听，都使用同一个 channel
//  3. 注册到 watchers[watcherID][configKey] = ch
//  4. 返回共享 channel 供 Handler 读取
//
// 数据结构：
//
//	watchers[watcherID][configKey] = channel
//	例如：
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config1"] = ch
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config2"] = ch
//	  watchers["watcher_abc123"]["default:public:DEFAULT:config3"] = ch
//	表示：同一个客户端监听多个配置，所有配置共用同一个 channel
//
// 参数：
//   - ctx: 上下文（用于取消监听）
//   - tenantId, namespaceId, groupName: 配置标识
//   - configDataIds: 要监听的配置标识列表（支持单个或多个）
//   - watcherID: 监听者唯一标识（由 Handler 生成）
//
// 返回：
//   - channel: 用于接收所有监听配置变更事件的共享通道
func (w *ConfigWatcher) Watch(
	ctx context.Context,
	tenantId, namespaceId, groupName string,
	configDataIds []string,
	watcherID string,
) <-chan *pb.ConfigChangeEvent {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 创建共享通道（所有配置共用）
	ch := make(chan *pb.ConfigChangeEvent, 100)

	// 初始化监听记录
	if w.watchers[watcherID] == nil {
		w.watchers[watcherID] = make(map[string]chan *pb.ConfigChangeEvent)
	}

	// 为每个配置注册监听（使用同一个 channel）
	for _, configDataId := range configDataIds {
		configKey := w.makeConfigKey(tenantId, namespaceId, groupName, configDataId)
		w.watchers[watcherID][configKey] = ch
	}

	return ch
}

// Unwatch 取消监听
//
// 处理流程：
//  1. 查找监听记录
//  2. 关闭共享 channel（所有配置共用）
//  3. 删除监听记录
func (w *ConfigWatcher) Unwatch(watcherID string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if configs, ok := w.watchers[watcherID]; ok {
		// 获取通道并关闭（所有配置共用同一个通道）
		var ch chan *pb.ConfigChangeEvent
		for _, c := range configs {
			ch = c
			break // 只需要关闭一次
		}
		if ch != nil {
			close(ch)
		}

		// 删除监听记录
		delete(w.watchers, watcherID)
	}
}

// SendToWatcher 向特定订阅者的 channel 发送事件（用于初始推送）
//
// 处理流程：
//  1. 查找订阅者的 channel
//  2. 非阻塞发送事件到该 channel
//
// 用途：
//   - 用于订阅成功后的初始配置推送
//   - 只发送给当前订阅者，不影响其他订阅者
func (w *ConfigWatcher) SendToWatcher(watcherID string, event *pb.ConfigChangeEvent) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if configs, ok := w.watchers[watcherID]; ok {
		// 获取共享 channel（所有配置共用同一个 channel）
		var ch chan *pb.ConfigChangeEvent
		for _, c := range configs {
			ch = c
			break // 只需要获取一次
		}
		if ch != nil {
			// 非阻塞发送
			select {
			case ch <- event:
				// 发送成功
			default:
				// 通道已满，丢弃事件（避免阻塞）
			}
		}
	}
}

// NotifyConfigChange 通知配置变更
//
// 处理流程：
//  1. 生成配置唯一键（configKey）
//  2. 遍历所有监听者，检查是否监听了该配置
//  3. 如果监听了，通过对应的 channel 发送事件（非阻塞）
//
// 如何确定发送给哪些客户端？
//   - 遍历 watchers[watcherID]，检查每个 watcherID 是否监听了该 configKey
//   - 如果监听了，通过对应的 channel 发送事件
//   - 每个 channel 对应一个客户端的 gRPC stream（在 Handler 的 goroutine 中）
//
// 事件传递链路：
//
//	NotifyConfigChange() -> channel -> Handler goroutine -> gRPC stream -> 客户端
//
// 并发安全：
//   - 使用 RLock 读取监听者映射（允许多个 goroutine 并发读取）
//   - 事件发送使用非阻塞方式，避免长时间持有锁
//
// 注意：
//   - 调用者需要确保 event 中的 namespaceId、groupName、configDataId 已设置
//   - NotifyConfigUpdate 和 NotifyConfigDelete 已经设置了这些字段
func (w *ConfigWatcher) NotifyConfigChange(tenantId, namespaceId, groupName, configDataId string, event *pb.ConfigChangeEvent) {
	configKey := w.makeConfigKey(tenantId, namespaceId, groupName, configDataId)

	w.mu.RLock()
	defer w.mu.RUnlock()

	// 通知所有监听者（检查是否监听了该配置）
	for watcherID, configs := range w.watchers {
		if ch, ok := configs[configKey]; ok {
			// 该监听者监听了此配置，发送事件
			select {
			case ch <- event:
				// 发送成功：事件已放入 channel，Handler 的 goroutine 会读取并推送给客户端
			default:
				// 通道已满，丢弃事件（避免阻塞）
				// 说明：客户端处理慢，channel 缓冲区（100）已满
				// 此时丢弃事件，避免阻塞其他监听者的通知流程
			}
			_ = watcherID // 用于调试
		}
	}
}

// GetWatcherCount 获取监听者数量
//
// 处理流程：
//  1. 生成配置唯一键（configKey）
//  2. 遍历所有监听者，统计监听了该配置的数量
//  3. 返回监听者数量
func (w *ConfigWatcher) GetWatcherCount(tenantId, namespaceId, groupName, configDataId string) int {
	configKey := w.makeConfigKey(tenantId, namespaceId, groupName, configDataId)

	w.mu.RLock()
	defer w.mu.RUnlock()

	// 统计监听者数量
	count := 0
	for _, configs := range w.watchers {
		if _, ok := configs[configKey]; ok {
			count++
		}
	}
	return count
}

// NotifyConfigUpdate 通知配置更新（辅助方法）
//
// 处理流程：
//  1. 构建配置更新事件（包含配置数据和 MD5）
//  2. 设置配置标识（namespaceId, groupName, configDataId）
//  3. 调用 NotifyConfigChange 通知所有监听者
func (w *ConfigWatcher) NotifyConfigUpdate(config *types.ConfigData) {
	event := &pb.ConfigChangeEvent{
		EventType:    "CONFIG_UPDATED",
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId:  config.NamespaceId,
		GroupName:    config.GroupName,
		ConfigDataId: config.ConfigDataId,
		Config:       convertConfigToProto(config),
		ContentMd5:   config.Md5Value,
	}

	w.NotifyConfigChange(config.TenantId, config.NamespaceId, config.GroupName, config.ConfigDataId, event)
}

// NotifyConfigDelete 通知配置删除（辅助方法）
//
// 处理流程：
//  1. 构建配置删除事件（不包含配置数据）
//  2. 设置配置标识（namespaceId, groupName, configDataId）
//  3. 调用 NotifyConfigChange 通知所有监听者
func (w *ConfigWatcher) NotifyConfigDelete(tenantId, namespaceId, groupName, configDataId string) {
	event := &pb.ConfigChangeEvent{
		EventType:    "CONFIG_DELETED",
		Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
		NamespaceId:  namespaceId,
		GroupName:    groupName,
		ConfigDataId: configDataId,
	}

	w.NotifyConfigChange(tenantId, namespaceId, groupName, configDataId, event)
}

// makeConfigKey 生成配置唯一键
func (w *ConfigWatcher) makeConfigKey(tenantId, namespaceId, groupName, configDataId string) string {
	return tenantId + ":" + namespaceId + ":" + groupName + ":" + configDataId
}

// ========== 辅助转换方法 ==========

func convertConfigToProto(config *types.ConfigData) *pb.ConfigData {
	if config == nil {
		return nil
	}
	return &pb.ConfigData{
		NamespaceId:   config.NamespaceId,
		GroupName:     config.GroupName,
		ConfigDataId:  config.ConfigDataId,
		ContentType:   config.ContentType,
		ConfigContent: config.ConfigContent,
		ContentMd5:    config.Md5Value,
		ConfigDesc:    config.ConfigDescription,
		ConfigVersion: config.Version,
	}
}
