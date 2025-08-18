package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
)

// EventConfig 事件配置
type EventConfig struct {
	BufferSize        int           `yaml:"bufferSize" json:"bufferSize"`
	WorkerCount       int           `yaml:"workerCount" json:"workerCount"`
	BatchSize         int           `yaml:"batchSize" json:"batchSize"`
	BatchTimeout      time.Duration `yaml:"batchTimeout" json:"batchTimeout"`
	MaxSubscribers    int           `yaml:"maxSubscribers" json:"maxSubscribers"`
	SubscriberTimeout time.Duration `yaml:"subscriberTimeout" json:"subscriberTimeout"`
	EnablePersistence bool          `yaml:"enablePersistence" json:"enablePersistence"`
	RetentionPeriod   time.Duration `yaml:"retentionPeriod" json:"retentionPeriod"`
	CleanupInterval   time.Duration `yaml:"cleanupInterval" json:"cleanupInterval"`
}

// DefaultEventConfig 默认事件配置
func DefaultEventConfig() *EventConfig {
	return &EventConfig{
		BufferSize:        1000,
		WorkerCount:       5,
		BatchSize:         100,
		BatchTimeout:      1 * time.Second,
		MaxSubscribers:    100,
		SubscriberTimeout: 30 * time.Second,
		EnablePersistence: true,
		RetentionPeriod:   7 * 24 * time.Hour, // 7天
		CleanupInterval:   1 * time.Hour,
	}
}

// Publisher 事件发布器实现
type Publisher struct {
	config      *EventConfig
	storage     core.Storage
	eventChan   chan *core.ServiceEvent
	subscribers map[string][]chan *core.ServiceEvent
	subMutex    sync.RWMutex
	workers     []*worker
	running     bool
	mutex       sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// worker 工作协程
type worker struct {
	id        int
	publisher *Publisher
	eventChan chan *core.ServiceEvent
	batch     []*core.ServiceEvent
	timer     *time.Timer
}

// NewPublisher 创建事件发布器
func NewPublisher(config *EventConfig, storage core.Storage) *Publisher {
	if config == nil {
		config = DefaultEventConfig()
	}

	ctx, cancel := context.WithCancel(context.Background())

	p := &Publisher{
		config:      config,
		storage:     storage,
		eventChan:   make(chan *core.ServiceEvent, config.BufferSize),
		subscribers: make(map[string][]chan *core.ServiceEvent),
		ctx:         ctx,
		cancel:      cancel,
	}

	// 创建工作协程
	p.workers = make([]*worker, config.WorkerCount)
	for i := 0; i < config.WorkerCount; i++ {
		p.workers[i] = &worker{
			id:        i,
			publisher: p,
			eventChan: make(chan *core.ServiceEvent, config.BufferSize/config.WorkerCount),
			batch:     make([]*core.ServiceEvent, 0, config.BatchSize),
			timer:     time.NewTimer(config.BatchTimeout),
		}
	}

	return p
}

// Start 启动事件发布器
func (p *Publisher) Start() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return nil
	}

	p.running = true

	// 启动事件分发协程
	go p.dispatchEvents()

	// 启动工作协程
	for _, worker := range p.workers {
		go worker.start()
	}

	// 启动清理协程
	if p.config.EnablePersistence {
		go p.cleanupEvents()
	}

	return nil
}

// Publish 发布事件
func (p *Publisher) Publish(ctx context.Context, event *core.ServiceEvent) error {
	p.mutex.RLock()
	if !p.running {
		p.mutex.RUnlock()
		return fmt.Errorf("event publisher not running")
	}
	p.mutex.RUnlock()

	if event == nil {
		return fmt.Errorf("event is nil")
	}

	// 设置事件时间
	if event.EventTime.IsZero() {
		event.EventTime = time.Now()
	}

	// 发送到事件通道
	select {
	case p.eventChan <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event channel full")
	}
}

// Subscribe 订阅事件
func (p *Publisher) Subscribe(ctx context.Context, tenantId, serviceName, groupName string) (<-chan *core.ServiceEvent, error) {
	p.mutex.RLock()
	if !p.running {
		p.mutex.RUnlock()
		return nil, fmt.Errorf("event publisher not running")
	}
	p.mutex.RUnlock()

	// 检查订阅者数量限制
	p.subMutex.RLock()
	totalSubscribers := 0
	for _, subs := range p.subscribers {
		totalSubscribers += len(subs)
	}
	p.subMutex.RUnlock()

	if totalSubscribers >= p.config.MaxSubscribers {
		return nil, core.ErrTooManySubscribers
	}

	// 创建订阅通道
	eventChan := make(chan *core.ServiceEvent, 100)

	// 构建订阅键
	key := p.buildSubscriptionKey(tenantId, serviceName, groupName)

	p.subMutex.Lock()
	p.subscribers[key] = append(p.subscribers[key], eventChan)
	p.subMutex.Unlock()

	// 启动超时清理协程
	go p.cleanupSubscriber(key, eventChan)

	return eventChan, nil
}

// Unsubscribe 取消订阅
func (p *Publisher) Unsubscribe(ctx context.Context, tenantId, serviceName, groupName string) error {
	key := p.buildSubscriptionKey(tenantId, serviceName, groupName)

	p.subMutex.Lock()
	defer p.subMutex.Unlock()

	// 关闭所有订阅通道
	if channels, exists := p.subscribers[key]; exists {
		for _, ch := range channels {
			close(ch)
		}
		delete(p.subscribers, key)
	}

	return nil
}

// Close 关闭事件发布器
func (p *Publisher) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if !p.running {
		return nil
	}

	p.running = false

	// 取消上下文
	p.cancel()

	// 关闭事件通道
	close(p.eventChan)

	// 等待工作协程结束
	for _, worker := range p.workers {
		worker.stop()
	}

	// 关闭所有订阅通道
	p.subMutex.Lock()
	for key, channels := range p.subscribers {
		for _, ch := range channels {
			close(ch)
		}
		delete(p.subscribers, key)
	}
	p.subMutex.Unlock()

	return nil
}

// ================== 内部方法 ==================

// dispatchEvents 分发事件
func (p *Publisher) dispatchEvents() {
	for event := range p.eventChan {
		// 选择工作协程（简单的轮询）
		workerIndex := int(time.Now().UnixNano()) % len(p.workers)
		worker := p.workers[workerIndex]

		select {
		case worker.eventChan <- event:
		default:
			// 如果工作协程忙，尝试其他工作协程
			for i := 0; i < len(p.workers); i++ {
				idx := (workerIndex + i) % len(p.workers)
				select {
				case p.workers[idx].eventChan <- event:
					goto sent
				default:
					continue
				}
			}
			// 所有工作协程都忙，丢弃事件
			fmt.Printf("all workers busy, dropping event: %s\n", event.EventType)
		sent:
		}
	}
}

// buildSubscriptionKey 构建订阅键
func (p *Publisher) buildSubscriptionKey(tenantId, serviceName, groupName string) string {
	if serviceName == "" && groupName == "" {
		return fmt.Sprintf("%s:*:*", tenantId)
	} else if serviceName == "" {
		return fmt.Sprintf("%s:%s:*", tenantId, groupName)
	} else if groupName == "" {
		return fmt.Sprintf("%s:*:%s", tenantId, serviceName)
	}
	return fmt.Sprintf("%s:%s:%s", tenantId, groupName, serviceName)
}

// matchSubscription 匹配订阅
func (p *Publisher) matchSubscription(event *core.ServiceEvent, key string) bool {
	parts := []string{event.TenantId, event.GroupName, event.ServiceName}
	keyParts := []string{event.TenantId, event.GroupName, event.ServiceName}

	// 解析订阅键
	if key != "" {
		keyParts = []string{}
		for _, part := range []string{event.TenantId, event.GroupName, event.ServiceName} {
			keyParts = append(keyParts, part)
		}
	}

	// 简单匹配逻辑
	eventKey := fmt.Sprintf("%s:%s:%s", parts[0], parts[1], parts[2])
	return key == eventKey || key == fmt.Sprintf("%s:*:*", parts[0]) ||
		key == fmt.Sprintf("%s:%s:*", parts[0], parts[1]) ||
		key == fmt.Sprintf("%s:*:%s", parts[0], parts[2])
}

// cleanupSubscriber 清理订阅者
func (p *Publisher) cleanupSubscriber(key string, eventChan chan *core.ServiceEvent) {
	timer := time.NewTimer(p.config.SubscriberTimeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		// 超时，清理订阅者
		p.subMutex.Lock()
		if channels, exists := p.subscribers[key]; exists {
			for i, ch := range channels {
				if ch == eventChan {
					// 移除这个通道
					p.subscribers[key] = append(channels[:i], channels[i+1:]...)
					close(eventChan)
					break
				}
			}
			// 如果没有订阅者了，删除键
			if len(p.subscribers[key]) == 0 {
				delete(p.subscribers, key)
			}
		}
		p.subMutex.Unlock()
	case <-p.ctx.Done():
		return
	}
}

// cleanupEvents 清理过期事件
func (p *Publisher) cleanupEvents() {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 执行清理逻辑
			// 这里应该调用存储层的清理方法
			// 暂时跳过具体实现
		case <-p.ctx.Done():
			return
		}
	}
}

// ================== 工作协程方法 ==================

// start 启动工作协程
func (w *worker) start() {
	w.timer.Stop()

	for {
		select {
		case event, ok := <-w.eventChan:
			if !ok {
				// 通道关闭，处理剩余批次
				w.processBatch()
				return
			}

			// 添加到批次
			w.batch = append(w.batch, event)

			// 检查是否需要处理批次
			if len(w.batch) >= w.publisher.config.BatchSize {
				w.processBatch()
				w.resetTimer()
			} else if len(w.batch) == 1 {
				// 第一个事件，启动定时器
				w.resetTimer()
			}

		case <-w.timer.C:
			// 超时，处理批次
			if len(w.batch) > 0 {
				w.processBatch()
			}
			w.resetTimer()

		case <-w.publisher.ctx.Done():
			// 上下文取消，处理剩余批次
			w.processBatch()
			return
		}
	}
}

// stop 停止工作协程
func (w *worker) stop() {
	close(w.eventChan)
	w.timer.Stop()
}

// processBatch 处理批次
func (w *worker) processBatch() {
	if len(w.batch) == 0 {
		return
	}

	// 持久化事件
	if w.publisher.config.EnablePersistence && w.publisher.storage != nil {
		for _, event := range w.batch {
			if err := w.publisher.storage.LogEvent(context.Background(), event); err != nil {
				fmt.Printf("log event failed: %v\n", err)
			}
		}
	}

	// 分发给订阅者
	w.publisher.subMutex.RLock()
	for key, channels := range w.publisher.subscribers {
		for _, event := range w.batch {
			if w.publisher.matchSubscription(event, key) {
				for _, ch := range channels {
					select {
					case ch <- event:
					default:
						// 订阅者通道满了，跳过
						fmt.Printf("subscriber channel full, dropping event: %s\n", event.EventType)
					}
				}
			}
		}
	}
	w.publisher.subMutex.RUnlock()

	// 清空批次
	w.batch = w.batch[:0]
}

// resetTimer 重置定时器
func (w *worker) resetTimer() {
	if !w.timer.Stop() {
		select {
		case <-w.timer.C:
		default:
		}
	}
	w.timer.Reset(w.publisher.config.BatchTimeout)
}

// IsRunning 检查是否运行中
func (p *Publisher) IsRunning() bool {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.running
}

// GetStats 获取统计信息
func (p *Publisher) GetStats() map[string]interface{} {
	p.subMutex.RLock()
	defer p.subMutex.RUnlock()

	totalSubscribers := 0
	subscriptionKeys := make([]string, 0, len(p.subscribers))

	for key, channels := range p.subscribers {
		totalSubscribers += len(channels)
		subscriptionKeys = append(subscriptionKeys, key)
	}

	return map[string]interface{}{
		"running":          p.IsRunning(),
		"totalSubscribers": totalSubscribers,
		"subscriptionKeys": subscriptionKeys,
		"workerCount":      len(p.workers),
		"eventChannelSize": len(p.eventChan),
		"eventChannelCap":  cap(p.eventChan),
	}
}
