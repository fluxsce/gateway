package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/logger"
)

// MemoryEventPublisher 内存事件发布器
// 使用内存存储和处理事件，适用于单机部署
type MemoryEventPublisher struct {
	subscribers []subscriber
	mutex       sync.RWMutex
	isRunning   bool
}

// subscriber 订阅者包装器
type subscriber struct {
	eventSubscriber core.EventSubscriber
}

// NewMemoryEventPublisher 创建一个新的内存事件发布器
// 默认包含数据库写入订阅者，使用默认配置
//
// 返回:
//   - *MemoryEventPublisher: 内存事件发布器实例
//   - error: 创建失败时返回错误信息
func NewMemoryEventPublisher() (*MemoryEventPublisher, error) {
	publisher := &MemoryEventPublisher{
		subscribers: make([]subscriber, 0),
	}

	// 默认创建数据库写入器并作为订阅者添加
	writer, err := NewEventDBWriter(nil) // 使用默认配置
	if err != nil {
		return nil, fmt.Errorf("failed to create event db writer: %w", err)
	}

	// 将数据库写入器作为订阅者添加到列表中
	publisher.subscribers = append(publisher.subscribers, subscriber{
		eventSubscriber: writer,
	})

	logger.Info("MemoryEventPublisher created with default database writer")

	return publisher, nil
}

// Publish 发布事件
// 异步发布事件到所有订阅者（包括数据库写入器）
// 使用独立的context避免前端请求结束导致事件处理中断
func (p *MemoryEventPublisher) Publish(ctx context.Context, event *core.ServiceEvent) error {
	if !p.isRunning {
		return fmt.Errorf("事件系统未启动")
	}

	// 设置事件的默认值（确保数据库 NOT NULL 约束字段有值）
	p.setEventDefaults(event)

	// 创建独立的context，避免前端请求结束导致事件处理中断
	// 设置30秒超时，确保事件处理有足够时间完成
	eventCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	// 异步发布事件到所有订阅者，使用独立的context
	go func() {
		defer cancel() // 确保context资源被释放
		p.publishEvent(eventCtx, event)
	}()

	return nil
}

// setEventDefaults 设置事件的默认值
// 确保数据库 NOT NULL 约束字段有值，避免插入时出现约束违反错误
func (p *MemoryEventPublisher) setEventDefaults(event *core.ServiceEvent) {
	now := time.Now()
	if event.TenantId == "" {
		event.TenantId = "default"
	}

	// 设置 EventType 默认值（数据库 NOT NULL 约束）
	if event.EventType == "" {
		event.EventType = "UNKNOWN"
	}

	// 设置 EventTime 默认值（数据库 NOT NULL 约束）
	event.EventTime = now

	// 设置 AddTime 默认值（数据库 NOT NULL 约束）
	if event.AddTime.IsZero() {
		event.AddTime = now
	}

	// 设置 AddWho 默认值（数据库 NOT NULL 约束）
	if event.AddWho == "" {
		event.AddWho = "SYSTEM"
	}

	// 设置 EditTime 默认值（数据库 NOT NULL 约束）
	if event.EditTime.IsZero() {
		event.EditTime = now
	}

	// 设置 EditWho 默认值（数据库 NOT NULL 约束）
	if event.EditWho == "" {
		event.EditWho = "SYSTEM"
	}

	// 设置 OprSeqFlag 默认值（数据库 NOT NULL 约束）
	if event.OprSeqFlag == "" {
		event.OprSeqFlag = event.ServiceEventId
	}

	// 设置 CurrentVersion 默认值（数据库 NOT NULL 约束，默认值为1）
	if event.CurrentVersion == 0 {
		event.CurrentVersion = 1
	}

	// 设置 ActiveFlag 默认值（数据库 NOT NULL 约束，默认值为'Y'）
	if event.ActiveFlag == "" {
		event.ActiveFlag = "Y"
	}
}

// publishEvent 发布事件到订阅者
func (p *MemoryEventPublisher) publishEvent(ctx context.Context, event *core.ServiceEvent) {
	p.mutex.RLock()
	defer p.mutex.RUnlock()

	for _, sub := range p.subscribers {
		// 检查是否订阅了该事件类型
		if p.matchEventType(sub.eventSubscriber.GetEventTypes(), event.EventType) {
			// 对于数据库写入器，使用同步处理以避免并发冲突
			if sub.eventSubscriber.GetSubscriberName() == "EventDBWriter" {
				// 同步处理数据库写入，避免SQLite并发锁冲突
				err := sub.eventSubscriber.HandleEvent(ctx, event)
				if err != nil {
					logger.WarnWithTrace(ctx, "事件处理失败",
						"eventType", event.EventType,
						"subscriber", sub.eventSubscriber.GetSubscriberName(),
						"error", err)
				}
			} else {
				// 其他订阅者仍然使用异步处理
				go func(subscriber core.EventSubscriber) {
					err := subscriber.HandleEvent(ctx, event)
					if err != nil {
						logger.WarnWithTrace(ctx, "事件处理失败",
							"eventType", event.EventType,
							"subscriber", subscriber.GetSubscriberName(),
							"error", err)
					}
				}(sub.eventSubscriber)
			}
		}
	}
}

// matchEventType 检查事件类型是否匹配
func (p *MemoryEventPublisher) matchEventType(subscribed []string, eventType string) bool {
	// 如果订阅列表为空，表示订阅所有事件
	if len(subscribed) == 0 {
		return true
	}

	// 检查是否匹配订阅的事件类型
	for _, t := range subscribed {
		if t == eventType {
			return true
		}
	}

	return false
}

// Subscribe 订阅事件
// 注册事件订阅者，订阅者通过接口方法处理事件
func (p *MemoryEventPublisher) Subscribe(eventSubscriber core.EventSubscriber) error {
	if eventSubscriber == nil {
		return fmt.Errorf("订阅者不能为空")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 添加订阅者
	p.subscribers = append(p.subscribers, subscriber{
		eventSubscriber: eventSubscriber,
	})

	logger.Debug("添加事件订阅者", "subscriberName", eventSubscriber.GetSubscriberName())
	return nil
}

// Unsubscribe 取消订阅
// 取消事件订阅者
func (p *MemoryEventPublisher) Unsubscribe(eventSubscriber core.EventSubscriber) error {
	if eventSubscriber == nil {
		return fmt.Errorf("订阅者不能为空")
	}

	p.mutex.Lock()
	defer p.mutex.Unlock()

	// 查找并移除订阅者
	for i, sub := range p.subscribers {
		if sub.eventSubscriber == eventSubscriber {
			// 调用订阅者的关闭方法
			if err := sub.eventSubscriber.Close(); err != nil {
				logger.Warn("关闭订阅者失败", "subscriberName", eventSubscriber.GetSubscriberName(), "error", err)
			}
			// 移除订阅者
			p.subscribers = append(p.subscribers[:i], p.subscribers[i+1:]...)
			logger.Debug("移除事件订阅者", "subscriberName", eventSubscriber.GetSubscriberName())
			return nil
		}
	}

	return fmt.Errorf("未找到订阅者: %s", eventSubscriber.GetSubscriberName())
}

// Start 启动事件系统
func (p *MemoryEventPublisher) Start(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.isRunning {
		return nil
	}

	p.isRunning = true
	logger.InfoWithTrace(ctx, "内存事件发布器已启动")

	return nil
}

// Stop 停止事件系统
func (p *MemoryEventPublisher) Stop(ctx context.Context) error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.isRunning {
		return nil
	}

	p.isRunning = false

	// 关闭所有订阅者
	for _, sub := range p.subscribers {
		if err := sub.eventSubscriber.Close(); err != nil {
			logger.WarnWithTrace(ctx, "关闭订阅者失败",
				"subscriberName", sub.eventSubscriber.GetSubscriberName(),
				"error", err)
		}
	}

	logger.InfoWithTrace(ctx, "内存事件发布器已停止")

	return nil
}
