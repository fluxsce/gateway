package event

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/registry/core"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// EventDBWriter 事件数据库写入器
// 实现了 EventHandler 接口，用于将服务事件写入数据库
//
// 主要特性:
//   - 实现 EventHandler 接口，可作为事件订阅者
//   - 支持同步和异步事件写入模式
//   - 异步模式下使用内存队列缓存事件
//   - 支持批量写入提高性能
//   - 定时刷新机制确保事件及时写入
//   - 线程安全的并发操作
//   - 优雅关闭确保数据不丢失
//   - 内存优化：处理完成后自动清理事件中的大对象，避免内存占用过大
type EventDBWriter struct {
	// 数据库连接实例
	db database.Database

	// 异步处理相关
	eventQueue  chan *core.ServiceEvent // 异步事件队列
	batchBuffer []*core.ServiceEvent    // 批量写入缓冲区
	flushTicker *time.Ticker            // 定时刷新ticker
	stopChan    chan struct{}           // 停止信号通道
	wg          sync.WaitGroup          // 等待组，用于优雅关闭

	// 互斥锁，保护批量缓冲区
	mutex sync.Mutex

	// 配置参数
	isAsync       bool          // 是否启用异步模式
	batchSize     int           // 批量写入大小
	queueSize     int           // 异步队列大小
	flushInterval time.Duration // 定时刷新间隔

	// 状态标识
	closed bool
}

// EventDBWriterConfig 事件数据库写入器配置
type EventDBWriterConfig struct {
	IsAsync       bool          `json:"isAsync"`       // 是否启用异步模式，默认false
	BatchSize     int           `json:"batchSize"`     // 批量写入大小，默认100
	QueueSize     int           `json:"queueSize"`     // 异步队列大小，默认1000
	FlushInterval time.Duration `json:"flushInterval"` // 定时刷新间隔，默认5秒
}

// DefaultEventDBWriterConfig 默认配置
func DefaultEventDBWriterConfig() *EventDBWriterConfig {
	return &EventDBWriterConfig{
		IsAsync:       true,
		BatchSize:     100,
		QueueSize:     1000,
		FlushInterval: 5 * time.Second,
	}
}

// NewEventDBWriter 创建一个新的事件数据库写入器
//
// 创建过程:
//  1. 获取数据库连接
//  2. 根据配置决定是否启用异步模式
//  3. 启动异步处理goroutine（如果启用异步）
//  4. 启动定时刷新机制
//
// 参数:
//   - config: 事件写入器配置，如果为nil则使用默认配置
//
// 返回:
//   - *EventDBWriter: 事件数据库写入器实例
//   - error: 创建失败时返回错误信息
func NewEventDBWriter(config *EventDBWriterConfig) (*EventDBWriter, error) {
	// 使用默认配置
	if config == nil {
		config = DefaultEventDBWriterConfig()
	}

	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to get default database connection")
	}

	writer := &EventDBWriter{
		db:            db,
		stopChan:      make(chan struct{}),
		batchBuffer:   make([]*core.ServiceEvent, 0, config.BatchSize),
		isAsync:       config.IsAsync,
		batchSize:     config.BatchSize,
		queueSize:     config.QueueSize,
		flushInterval: config.FlushInterval,
	}

	// 如果启用异步模式，初始化异步处理
	if config.IsAsync {
		writer.eventQueue = make(chan *core.ServiceEvent, config.QueueSize)
		writer.startAsyncProcessor()
	}

	// 启动定时刷新（无论同步还是异步模式都需要）
	writer.startFlushTimer()

	logger.Info("EventDBWriter created", "isAsync", config.IsAsync, "batchSize", config.BatchSize)

	return writer, nil
}

// HandleEvent 实现 EventHandler 接口，处理单个事件
// 根据配置决定是同步写入数据库还是放入异步队列
// 同时会根据事件类型同步更新相关的数据库表
// 注意：处理过程中会自动清理事件中的大对象（Service、Instance）以优化内存使用
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - event: 要处理的服务事件
//
// 返回:
//   - error: 处理失败时返回错误信息
func (w *EventDBWriter) HandleEvent(ctx context.Context, event *core.ServiceEvent) error {
	if w.closed {
		return fmt.Errorf("event writer is closed")
	}

	// 首先处理数据同步操作（根据事件类型更新相关表）
	if err := w.syncDataByEvent(ctx, event); err != nil {
		logger.Error("Failed to sync data by event", "error", err, "eventType", event.EventType, "eventId", event.ServiceEventId)
		// 注意：数据同步失败不影响事件记录，继续执行事件写入
	}

	// 如果启用异步模式，将事件放入队列
	if w.isAsync {
		select {
		case w.eventQueue <- event:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 队列满时的处理策略
			logger.Warn("Event queue is full, dropping event", "eventType", event.EventType, "eventId", event.ServiceEventId)
			return fmt.Errorf("event queue is full")
		}
	}

	// 同步模式：添加到批量缓冲区或直接写入
	return w.addToBatch(event)
}

// BatchWrite 批量写入多个事件
//
// 参数:
//   - ctx: 上下文
//   - events: 要写入的事件数组
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *EventDBWriter) BatchWrite(ctx context.Context, events []*core.ServiceEvent) error {
	if len(events) == 0 {
		return nil
	}

	if w.closed {
		return fmt.Errorf("event writer is closed")
	}

	// 如果启用异步模式，将所有事件放入队列
	if w.isAsync {
		for _, event := range events {
			select {
			case w.eventQueue <- event:
				// 成功放入队列
			case <-ctx.Done():
				return ctx.Err()
			default:
				logger.Warn("Event queue is full, dropping event", "eventType", event.EventType, "eventId", event.ServiceEventId)
			}
		}
		return nil
	}

	// 同步模式：直接批量写入数据库
	return w.batchWriteDirectly(ctx, events)
}

// Flush 刷新缓冲区，将缓存的事件写入数据库
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 刷新失败时返回错误信息
func (w *EventDBWriter) Flush(ctx context.Context) error {
	if w.closed {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.batchBuffer) == 0 {
		return nil
	}

	// 执行批量写入
	err := w.batchWriteDirectly(ctx, w.batchBuffer)
	if err != nil {
		logger.Error("Failed to flush event batch buffer", "error", err, "count", len(w.batchBuffer))
		return err
	}

	// 清空缓冲区
	w.batchBuffer = w.batchBuffer[:0]
	logger.Debug("Flushed event batch buffer", "count", len(w.batchBuffer))

	return nil
}

// Close 关闭写入器，优雅停止异步处理并刷新缓冲区
//
// 返回:
//   - error: 关闭失败时返回错误信息
func (w *EventDBWriter) Close() error {
	if w.closed {
		return nil
	}

	w.closed = true

	// 发送停止信号
	close(w.stopChan)

	// 等待异步处理goroutine结束
	w.wg.Wait()

	// 刷新剩余的缓冲区数据
	ctx := context.Background()
	if err := w.Flush(ctx); err != nil {
		logger.Error("Failed to flush event buffer during close", "error", err)
	}

	// 关闭定时器
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}

	logger.Info("EventDBWriter closed successfully")
	return nil
}

// startAsyncProcessor 启动异步事件处理goroutine
func (w *EventDBWriter) startAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Async event processor stopped")

		logger.Info("Async event processor started")

		for {
			select {
			case event := <-w.eventQueue:
				if err := w.addToBatch(event); err != nil {
					logger.Error("Failed to process event in async mode", "error", err, "eventId", event.ServiceEventId)
				}

			case <-w.stopChan:
				// 处理剩余的队列中的事件
				w.drainQueue()
				return
			}
		}
	}()
}

// startFlushTimer 启动定时刷新机制
func (w *EventDBWriter) startFlushTimer() {
	if w.flushInterval <= 0 {
		return
	}

	w.flushTicker = time.NewTicker(w.flushInterval)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Event flush timer stopped")

		logger.Info("Event flush timer started", "interval", w.flushInterval)

		for {
			select {
			case <-w.flushTicker.C:
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled event flush failed", "error", err)
				}

			case <-w.stopChan:
				return
			}
		}
	}()
}

// drainQueue 排空队列中剩余的事件
func (w *EventDBWriter) drainQueue() {
	logger.Info("Draining event queue")
	count := 0

	for {
		select {
		case event := <-w.eventQueue:
			if err := w.addToBatch(event); err != nil {
				logger.Error("Failed to process event while draining queue", "error", err, "eventId", event.ServiceEventId)
			}
			count++

		default:
			// 队列为空，执行最终刷新确保缓冲区数据写入
			if count > 0 {
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Failed to flush during event queue drain", "error", err)
				}
			}
			logger.Info("Event queue drained", "processedCount", count)
			return
		}
	}
}

// addToBatch 将事件添加到批量缓冲区
func (w *EventDBWriter) addToBatch(event *core.ServiceEvent) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.batchBuffer = append(w.batchBuffer, event)

	// 如果缓冲区满了，立即刷新
	if len(w.batchBuffer) >= w.batchSize {
		ctx := context.Background()
		if err := w.batchWriteDirectly(ctx, w.batchBuffer); err != nil {
			logger.Error("Failed to write full event batch", "error", err, "count", len(w.batchBuffer))
			return err
		}
		w.batchBuffer = w.batchBuffer[:0]
	}

	return nil
}

// writeDirectly 直接写入单个事件到数据库
func (w *EventDBWriter) writeDirectly(ctx context.Context, event *core.ServiceEvent) error {
	_, err := w.db.Insert(ctx, "HUB_REGISTRY_SERVICE_EVENT", event, true)
	if err != nil {
		return fmt.Errorf("failed to write event: %w", err)
	}
	return nil
}

// batchWriteDirectly 直接批量写入事件到数据库
func (w *EventDBWriter) batchWriteDirectly(ctx context.Context, events []*core.ServiceEvent) error {
	if len(events) == 0 {
		return nil
	}

	// 使用数据库的批量插入方法，自动处理SQL构建和事务提交
	_, err := w.db.BatchInsert(ctx, "HUB_REGISTRY_SERVICE_EVENT", events, true)
	if err != nil {
		return fmt.Errorf("failed to write event batch: %w", err)
	}

	logger.Debug("Successfully wrote event batch", "count", len(events))
	return nil
}

// GetEventTypes 实现 EventSubscriber 接口
// 返回nil表示订阅所有事件类型
func (w *EventDBWriter) GetEventTypes() []string {
	return nil // 数据库写入器订阅所有事件类型
}

// GetSubscriberName 实现 EventSubscriber 接口
// 返回订阅者名称，用于日志记录和调试
func (w *EventDBWriter) GetSubscriberName() string {
	return "EventDBWriter"
}

// syncDataByEvent 根据事件类型同步更新相关的数据库表
// 这个方法会根据不同的事件类型，更新对应的业务表数据
//
// 参数:
//   - ctx: 上下文
//   - event: 服务事件
//
// 返回:
//   - error: 同步失败时返回错误信息
func (w *EventDBWriter) syncDataByEvent(ctx context.Context, event *core.ServiceEvent) error {
	switch event.EventType {
	// 服务相关事件 - 使用 types.go 中定义的常量
	case core.EventTypeServiceRegistered, core.EventTypeServiceUpdated:
		return w.syncServiceData(ctx, event)
	case core.EventTypeServiceDeregistered:
		return w.deleteServiceData(ctx, event)

	// 实例相关事件 - 使用 types.go 中定义的常量
	case core.EventTypeInstanceRegistered:
		return w.syncInstanceData(ctx, event)
	case core.EventTypeInstanceDeregistered:
		return w.deleteInstanceData(ctx, event)
	case core.EventTypeInstanceHeartbeatUpdated:
		return w.updateInstanceHeartbeat(ctx, event)
	case core.EventTypeInstanceHealthChange, core.EventTypeInstanceStatusChange:
		return w.syncInstanceData(ctx, event) // 状态变更也需要同步实例数据

	// 分组相关事件 - 不需要同步到数据库，直接跳过
	case core.EventTypeServiceGroupCreated, core.EventTypeServiceGroupUpdated, core.EventTypeServiceGroupDeleted:
		logger.Debug("Service group events do not require data sync", "eventType", event.EventType)
		return nil

	default:
		// 对于不需要数据同步的事件类型，直接返回成功
		logger.Debug("Event type does not require data sync", "eventType", event.EventType)
		return nil
	}
}

// syncServiceData 同步服务数据到数据库
// 处理服务注册和更新事件，将服务信息同步到 HUB_REGISTRY_SERVICE 表
// 要求事件必须携带完整的服务对象，否则直接跳过处理
// 会先检查数据是否存在，根据检查结果决定插入或更新操作
// 处理完成后会将事件中的 Service 对象置空，避免内存占用过大
func (w *EventDBWriter) syncServiceData(ctx context.Context, event *core.ServiceEvent) error {
	// 只处理携带完整服务对象的事件
	if event.Service == nil {
		logger.Debug("Event does not carry service object, skipping sync", "eventType", event.EventType, "eventId", event.ServiceEventId)
		return nil
	}

	service := event.Service
	// 确保基本字段正确设置
	service.EditTime = event.EventTime
	service.EditWho = "SYSTEM"
	service.OprSeqFlag = event.ServiceEventId

	// 检查服务是否已存在
	var existingService core.Service
	query := "SELECT tenantId, serviceName FROM HUB_REGISTRY_SERVICE WHERE tenantId = ? AND serviceName = ?"
	queryArgs := []interface{}{event.TenantId, event.ServiceName}
	err := w.db.QueryOne(ctx, &existingService, query, queryArgs, true)

	var isServiceExists bool
	if err != nil {
		if err == database.ErrRecordNotFound {
			// 服务不存在，需要插入
			isServiceExists = false
			logger.Debug("Service does not exist, will insert", "serviceName", event.ServiceName, "tenantId", event.TenantId)
		} else {
			// 查询出错
			return fmt.Errorf("failed to check service existence: %w", err)
		}
	} else {
		// 服务已存在，需要更新
		isServiceExists = true
		logger.Debug("Service already exists, will update", "serviceName", event.ServiceName, "tenantId", event.TenantId)
	}

	// 根据存在性和事件类型执行相应操作
	if !isServiceExists {
		// 服务不存在，执行插入操作
		_, err := w.db.Insert(ctx, "HUB_REGISTRY_SERVICE", service, true)
		if err != nil {
			return fmt.Errorf("failed to insert service: %w", err)
		}
		logger.Debug("Service data inserted to database", "serviceName", event.ServiceName, "eventType", event.EventType)
	} else {
		// 服务已存在，执行更新操作
		whereClause := "tenantId = ? AND serviceName = ?"
		whereArgs := []interface{}{event.TenantId, event.ServiceName}
		_, err := w.db.Update(ctx, "HUB_REGISTRY_SERVICE", service, whereClause, whereArgs, true)
		if err != nil {
			return fmt.Errorf("failed to update service: %w", err)
		}
		logger.Debug("Service data updated in database", "serviceName", event.ServiceName, "eventType", event.EventType)
	}

	// 内存优化：数据同步完成后将 Service 对象置空，避免在事件队列中占用过多内存
	event.Service = nil
	logger.Debug("Service object cleared from event for memory optimization", "eventId", event.ServiceEventId)

	return nil
}

// syncInstanceData 同步实例数据到数据库
// 处理实例注册和更新事件，将实例信息同步到 HUB_REGISTRY_SERVICE_INSTANCE 表
// 要求事件必须携带完整的实例对象，否则直接跳过处理
// 会先检查数据是否存在，根据检查结果决定插入或更新操作
// 处理完成后会将事件中的 Instance 对象置空，避免内存占用过大
func (w *EventDBWriter) syncInstanceData(ctx context.Context, event *core.ServiceEvent) error {
	// 只处理携带完整实例对象的事件
	if event.Instance == nil {
		logger.Debug("Event does not carry instance object, skipping sync", "eventType", event.EventType, "eventId", event.ServiceEventId)
		return nil
	}

	instance := event.Instance
	// 确保基本字段正确设置
	instance.EditTime = event.EventTime
	instance.EditWho = "SYSTEM"
	instance.OprSeqFlag = event.ServiceEventId
	// 更新时间戳信息
	if event.EventType == core.EventTypeInstanceRegistered || event.EventType == "INSTANCE_REGISTERED" {
		instance.RegisterTime = event.EventTime
	}
	if instance.LastHeartbeatTime == nil {
		instance.LastHeartbeatTime = &event.EventTime
	}
	if instance.LastHealthCheckTime == nil {
		instance.LastHealthCheckTime = &event.EventTime
	}

	// 检查实例是否已存在
	var existingInstance core.ServiceInstance
	query := "SELECT serviceInstanceId, tenantId FROM HUB_REGISTRY_SERVICE_INSTANCE WHERE serviceInstanceId = ? AND tenantId = ?"
	queryArgs := []interface{}{event.ServiceInstanceId, event.TenantId}
	err := w.db.QueryOne(ctx, &existingInstance, query, queryArgs, true)

	var isInstanceExists bool
	if err != nil {
		if err == database.ErrRecordNotFound {
			// 实例不存在，需要插入
			isInstanceExists = false
			logger.Debug("Instance does not exist, will insert", "instanceId", event.ServiceInstanceId, "tenantId", event.TenantId)
		} else {
			// 查询出错
			return fmt.Errorf("failed to check instance existence: %w", err)
		}
	} else {
		// 实例已存在，需要更新
		isInstanceExists = true
		logger.Debug("Instance already exists, will update", "instanceId", event.ServiceInstanceId, "tenantId", event.TenantId)
	}

	// 根据存在性执行相应操作
	if !isInstanceExists {
		// 实例不存在，执行插入操作
		_, err := w.db.Insert(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", instance, true)
		if err != nil {
			return fmt.Errorf("failed to insert instance: %w", err)
		}
		logger.Debug("Instance data inserted to database", "instanceId", event.ServiceInstanceId, "eventType", event.EventType)
	} else {
		// 实例已存在，执行更新操作
		whereClause := "serviceInstanceId = ? AND tenantId = ?"
		whereArgs := []interface{}{event.ServiceInstanceId, event.TenantId}
		_, err := w.db.Update(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", instance, whereClause, whereArgs, true)
		if err != nil {
			return fmt.Errorf("failed to update instance: %w", err)
		}
		logger.Debug("Instance data updated in database", "instanceId", event.ServiceInstanceId, "eventType", event.EventType)
	}

	// 内存优化：数据同步完成后将 Instance 对象置空，避免在事件队列中占用过多内存
	event.Instance = nil
	logger.Debug("Instance object cleared from event for memory optimization", "eventId", event.ServiceEventId)

	return nil
}

// deleteServiceData 删除服务数据
// 处理服务注销事件，从 HUB_REGISTRY_SERVICE 表中删除服务数据
func (w *EventDBWriter) deleteServiceData(ctx context.Context, event *core.ServiceEvent) error {
	if event.ServiceName == "" || event.TenantId == "" {
		return fmt.Errorf("service deletion requires serviceName and tenantId")
	}

	// 执行数据库删除
	whereClause := "tenantId = ? AND serviceName = ?"
	whereArgs := []interface{}{event.TenantId, event.ServiceName}
	_, err := w.db.Delete(ctx, "HUB_REGISTRY_SERVICE", whereClause, whereArgs, true)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	logger.Debug("Service data deleted from database", "serviceName", event.ServiceName, "eventType", event.EventType)
	return nil
}

// deleteInstanceData 删除实例数据
// 处理实例注销事件，从 HUB_REGISTRY_SERVICE_INSTANCE 表中删除实例数据
func (w *EventDBWriter) deleteInstanceData(ctx context.Context, event *core.ServiceEvent) error {
	if event.ServiceInstanceId == "" || event.TenantId == "" {
		return fmt.Errorf("instance deletion requires serviceInstanceId and tenantId")
	}

	// 执行数据库删除
	whereClause := "serviceInstanceId = ? AND tenantId = ?"
	whereArgs := []interface{}{event.ServiceInstanceId, event.TenantId}
	_, err := w.db.Delete(ctx, "HUB_REGISTRY_SERVICE_INSTANCE", whereClause, whereArgs, true)
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	logger.Debug("Instance data deleted from database", "instanceId", event.ServiceInstanceId, "eventType", event.EventType)
	return nil
}

// updateInstanceHeartbeat 更新实例心跳时间
// 处理实例心跳事件，更新实例的最后心跳时间和健康状态
// 要求事件必须携带完整的实例对象，否则直接跳过处理
// 处理完成后会将事件中的 Instance 对象置空，避免内存占用过大
func (w *EventDBWriter) updateInstanceHeartbeat(ctx context.Context, event *core.ServiceEvent) error {
	// 只处理携带完整实例对象的事件
	if event.Instance == nil {
		logger.Debug("Event does not carry instance object, skipping heartbeat update", "eventType", event.EventType, "eventId", event.ServiceEventId)
		return nil
	}

	// 更新心跳相关字段
	//event.Instance.LastHeartbeatTime = &event.EventTime
	//event.Instance.HealthStatus = core.HealthStatusHealthy
	//event.Instance.EditTime = time.Now()
	//event.Instance.EditWho = "SYSTEM"
	//event.Instance.OprSeqFlag = event.ServiceEventId

	// 执行完整实例更新（注意：syncInstanceData 内部会置空 Instance 对象）
	return w.syncInstanceData(ctx, event)
}
