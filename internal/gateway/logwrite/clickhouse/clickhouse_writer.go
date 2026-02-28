package clickhouse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ClickHouseWriter 实现了 LogWriter 接口，专门用于将网关访问日志写入ClickHouse
// 针对ClickHouse的特性进行了优化:
// - 使用批量写入提高性能
// - 利用ClickHouse的列式存储特性
// - 优化内存使用和写入性能
type ClickHouseWriter struct {
	// 日志配置
	config *types.LogConfig

	// ClickHouse数据库连接实例
	db database.Database

	// 异步处理相关
	logQueue    chan *types.AccessLog // 异步日志队列
	batchBuffer []*types.AccessLog    // 批量写入缓冲区
	flushTicker *time.Ticker          // 定时刷新ticker
	stopChan    chan struct{}         // 停止信号通道
	wg          sync.WaitGroup        // 等待组，用于优雅关闭

	// 互斥锁，保护批量缓冲区
	mutex sync.Mutex

	// 后端追踪日志异步处理相关（与主表保持一致的处理模式）
	backendTraceLogQueue    chan *types.BackendTraceLog // 异步后端追踪日志队列
	backendTraceBatchBuffer []*types.BackendTraceLog    // 后端追踪日志批量写入缓冲区
	backendTraceMutex       sync.Mutex                  // 保护后端追踪日志批量缓冲区的互斥锁

	// 状态标识
	closed bool

	// ClickHouse特定的计数器
	insertedCount uint64
	batchCount    uint64
}

// NewClickHouseWriter 创建一个新的ClickHouse日志写入器
func NewClickHouseWriter(config *types.LogConfig) (*ClickHouseWriter, error) {
	// 获取ClickHouse数据库连接
	db := database.GetConnection("clickhouse_main")
	if db == nil {
		return nil, fmt.Errorf("failed to get clickhouse_main database connection")
	}

	// 针对ClickHouse优化批处理大小
	if config.BatchSize < 5000 {
		config.BatchSize = 5000 // ClickHouse推荐的最小批处理大小
	}

	writer := &ClickHouseWriter{
		config:                  config,
		db:                      db,
		stopChan:                make(chan struct{}),
		batchBuffer:             make([]*types.AccessLog, 0, config.BatchSize),
		backendTraceBatchBuffer: make([]*types.BackendTraceLog, 0, config.BatchSize),
	}

	// 如果启用异步日志，初始化异步处理
	if config.IsAsyncLogging() {
		writer.logQueue = make(chan *types.AccessLog, config.AsyncQueueSize)
		writer.backendTraceLogQueue = make(chan *types.BackendTraceLog, config.AsyncQueueSize)
		writer.startAsyncProcessor()
		writer.startBackendTraceAsyncProcessor()
	}

	// 启动定时刷新
	writer.startFlushTimer()

	return writer, nil
}

// Write 写入单条访问日志
func (w *ClickHouseWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	// 如果启用异步模式，将日志放入队列
	if w.config.IsAsyncLogging() {
		select {
		case w.logQueue <- log:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 队列满时的处理策略
			logger.Warn("ClickHouse log queue is full, dropping log entry", "traceId", log.TraceID)
			return fmt.Errorf("log queue is full")
		}
	}

	// 同步模式：优先使用批量写入
	return w.addToBatch(log)
}

// BatchWrite 批量写入多条访问日志
func (w *ClickHouseWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	// 如果启用异步模式，将所有日志放入队列
	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			select {
			case w.logQueue <- log:
				// 成功放入队列
			case <-ctx.Done():
				return ctx.Err()
			default:
				logger.Warn("ClickHouse log queue is full, dropping log entry", "traceId", log.TraceID)
			}
		}
		return nil
	}

	// 同步模式：直接批量写入
	return w.batchWriteDirectly(ctx, logs)
}

// Flush 刷新缓冲区，将缓存的日志写入ClickHouse
func (w *ClickHouseWriter) Flush(ctx context.Context) error {
	if w.closed {
		return nil
	}

	w.mutex.Lock()
	defer w.mutex.Unlock()

	if len(w.batchBuffer) == 0 {
		return nil
	}

	// 保存计数用于日志
	count := len(w.batchBuffer)

	// 执行批量写入
	err := w.batchWriteDirectly(ctx, w.batchBuffer)

	if err != nil {
		// 打印失败批次的关键信息，便于排查问题
		logger.Error("Failed to flush ClickHouse batch buffer, dumping failed batch data", "error", err, "count", count)
		for i, log := range w.batchBuffer {
			logger.Warn("Failed ClickHouse batch item",
				"index", i,
				"traceId", log.TraceID,
				"requestMethod", log.RequestMethod,
				"requestMethodLen", len(log.RequestMethod),
				"requestPath", log.RequestPath,
				"requestPathLen", len(log.RequestPath),
				"forwardMethod", log.ForwardMethod,
				"forwardMethodLen", len(log.ForwardMethod),
				"clientIp", log.ClientIPAddress)
		}
	}

	// 无论成功或失败都清空缓冲区，避免失败数据重复写入导致死循环
	w.batchBuffer = w.batchBuffer[:0]

	if err != nil {
		return err
	}

	logger.Debug("Flushed ClickHouse batch buffer", "count", count)
	return nil
}

// Close 关闭写入器
func (w *ClickHouseWriter) Close() error {
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
		logger.Error("Failed to flush ClickHouse buffer during close", "error", err)
	}
	if err := w.FlushBackendTrace(ctx); err != nil {
		logger.Error("Failed to flush ClickHouse backend trace buffer during close", "error", err)
	}

	// 关闭定时器
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}

	logger.Info("ClickHouseWriter closed successfully",
		"totalInserted", w.insertedCount,
		"totalBatches", w.batchCount)
	return nil
}

// GetLogConfig 获取日志配置
func (w *ClickHouseWriter) GetLogConfig() *types.LogConfig {
	return w.config
}

// WriteBackendTraceLog 写入单条后端追踪日志（从表）
// 根据配置决定是同步写入数据库还是放入异步队列
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - log: 要写入的后端追踪日志
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *ClickHouseWriter) WriteBackendTraceLog(ctx context.Context, log *types.BackendTraceLog) error {
	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	// 如果启用异步模式，将日志放入队列
	if w.config.IsAsyncLogging() {
		select {
		case w.backendTraceLogQueue <- log:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 队列满时的处理策略
			logger.Warn("ClickHouse backend trace log queue is full, dropping log entry", "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
			return fmt.Errorf("backend trace log queue is full")
		}
	}

	// 同步模式：直接写入数据库或缓存批量写入
	if w.config.IsBatchProcessing() {
		return w.addBackendTraceToBatch(log)
	}

	// 直接写入数据库
	return w.writeBackendTraceLogDirectly(ctx, log)
}

// BatchWriteBackendTraceLog 批量写入后端追踪日志（从表）
//
// 参数:
//   - ctx: 上下文
//   - logs: 要写入的日志数组
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *ClickHouseWriter) BatchWriteBackendTraceLog(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	if w.closed {
		return fmt.Errorf("writer is closed")
	}

	// 如果启用异步模式，将所有日志放入队列
	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			select {
			case w.backendTraceLogQueue <- log:
				// 成功放入队列
			case <-ctx.Done():
				return ctx.Err()
			default:
				logger.Warn("ClickHouse backend trace log queue is full, dropping log entry", "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
			}
		}
		return nil
	}

	// 同步模式：直接批量写入数据库
	return w.batchWriteBackendTraceLogDirectly(ctx, logs)
}

// writeBackendTraceLogDirectly 直接写入单条后端追踪日志到ClickHouse
func (w *ClickHouseWriter) writeBackendTraceLogDirectly(ctx context.Context, log *types.BackendTraceLog) error {
	_, err := w.db.Insert(ctx, log.TableName(), log, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log: %w", err)
	}

	w.insertedCount++
	return nil
}

// batchWriteBackendTraceLogDirectly 直接批量写入后端追踪日志到ClickHouse
func (w *ClickHouseWriter) batchWriteBackendTraceLogDirectly(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	startTime := time.Now()

	// 使用数据库的批量插入方法
	tableName := logs[0].TableName()
	_, err := w.db.BatchInsert(ctx, tableName, logs, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log batch: %w", err)
	}

	// 更新计数器
	w.insertedCount += uint64(len(logs))
	w.batchCount++

	// 记录性能指标
	duration := time.Since(startTime)
	recordsPerSecond := float64(len(logs)) / duration.Seconds()

	logger.Debug("ClickHouse backend trace log batch write completed",
		"count", len(logs),
		"duration", duration,
		"recordsPerSecond", recordsPerSecond,
		"totalInserted", w.insertedCount,
		"totalBatches", w.batchCount)

	return nil
}

// startAsyncProcessor 启动异步日志处理goroutine
func (w *ClickHouseWriter) startAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("ClickHouse async log processor stopped")

		logger.Info("ClickHouse async log processor started")

		for {
			select {
			case log := <-w.logQueue:
				w.addToBatch(log)

			case <-w.stopChan:
				// 处理剩余的队列中的日志
				w.drainQueue()
				return
			}
		}
	}()
}

// startFlushTimer 启动定时刷新机制
func (w *ClickHouseWriter) startFlushTimer() {
	if w.config.AsyncFlushIntervalMs <= 0 {
		return
	}

	w.flushTicker = time.NewTicker(time.Duration(w.config.AsyncFlushIntervalMs) * time.Millisecond)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("ClickHouse flush timer stopped")

		logger.Info("ClickHouse flush timer started", "intervalMs", w.config.AsyncFlushIntervalMs)

		for {
			select {
			case <-w.flushTicker.C:
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled ClickHouse flush failed", "error", err)
				}
				if err := w.FlushBackendTrace(ctx); err != nil {
					logger.Error("Scheduled ClickHouse backend trace flush failed", "error", err)
				}

			case <-w.stopChan:
				return
			}
		}
	}()
}

// drainQueue 排空队列中剩余的日志
func (w *ClickHouseWriter) drainQueue() {
	logger.Info("Draining ClickHouse log queue")
	count := 0

	for {
		select {
		case log := <-w.logQueue:
			w.addToBatch(log)
			count++

		default:
			// 队列为空，执行最终刷新
			if count > 0 {
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Failed to flush ClickHouse during queue drain", "error", err)
				}
			}
			logger.Info("ClickHouse queue drained", "processedCount", count)
			return
		}
	}
}

// FlushBackendTrace 刷新后端追踪日志缓冲区，将缓存的日志写入ClickHouse
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 刷新失败时返回错误信息
func (w *ClickHouseWriter) FlushBackendTrace(ctx context.Context) error {
	if w.closed {
		return nil
	}

	w.backendTraceMutex.Lock()
	defer w.backendTraceMutex.Unlock()

	if len(w.backendTraceBatchBuffer) == 0 {
		return nil
	}

	// 保存计数用于日志
	count := len(w.backendTraceBatchBuffer)

	// 执行批量写入
	err := w.batchWriteBackendTraceLogDirectly(ctx, w.backendTraceBatchBuffer)

	if err != nil {
		// 打印失败批次的关键信息，便于排查问题
		logger.Error("Failed to flush ClickHouse backend trace batch buffer, dumping failed batch data", "error", err, "count", count)
		for i, log := range w.backendTraceBatchBuffer {
			logger.Warn("Failed ClickHouse backend trace batch item",
				"index", i,
				"traceId", log.TraceID,
				"backendTraceId", log.BackendTraceID,
				"forwardMethod", log.ForwardMethod,
				"forwardMethodLen", len(log.ForwardMethod),
				"forwardPath", log.ForwardPath,
				"forwardPathLen", len(log.ForwardPath),
				"serviceId", log.ServiceDefinitionID,
				"serviceName", log.ServiceName)
		}
	}

	// 无论成功或失败都清空缓冲区，避免失败数据重复写入导致死循环
	w.backendTraceBatchBuffer = w.backendTraceBatchBuffer[:0]

	if err != nil {
		return err
	}

	logger.Debug("Flushed ClickHouse backend trace batch buffer", "count", count)
	return nil
}

// startBackendTraceAsyncProcessor 启动后端追踪日志异步处理goroutine
func (w *ClickHouseWriter) startBackendTraceAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("ClickHouse async backend trace log processor stopped")

		logger.Info("ClickHouse async backend trace log processor started")

		for {
			select {
			case log := <-w.backendTraceLogQueue:
				if w.config.IsBatchProcessing() {
					w.addBackendTraceToBatch(log)
				} else {
					ctx := context.Background()
					if err := w.writeBackendTraceLogDirectly(ctx, log); err != nil {
						logger.Error("Failed to write backend trace log in async mode", "error", err, "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
					}
				}

			case <-w.stopChan:
				// 处理剩余的队列中的日志
				w.drainBackendTraceQueue()
				return
			}
		}
	}()
}

// drainBackendTraceQueue 排空后端追踪日志队列中剩余的日志
func (w *ClickHouseWriter) drainBackendTraceQueue() {
	logger.Info("Draining ClickHouse backend trace log queue")
	count := 0

	for {
		select {
		case log := <-w.backendTraceLogQueue:
			if w.config.IsBatchProcessing() {
				w.addBackendTraceToBatch(log)
			} else {
				ctx := context.Background()
				if err := w.writeBackendTraceLogDirectly(ctx, log); err != nil {
					logger.Error("Failed to write backend trace log while draining queue", "error", err, "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
				}
			}
			count++

		default:
			// 队列为空，执行最终刷新确保缓冲区数据写入
			if count > 0 {
				ctx := context.Background()
				if err := w.FlushBackendTrace(ctx); err != nil {
					logger.Error("Failed to flush ClickHouse backend trace during queue drain", "error", err)
				}
			}
			logger.Info("ClickHouse backend trace queue drained", "processedCount", count)
			return
		}
	}
}

// addBackendTraceToBatch 将后端追踪日志添加到批量缓冲区
func (w *ClickHouseWriter) addBackendTraceToBatch(log *types.BackendTraceLog) error {
	w.backendTraceMutex.Lock()
	defer w.backendTraceMutex.Unlock()

	w.backendTraceBatchBuffer = append(w.backendTraceBatchBuffer, log)

	// 如果缓冲区满了，立即刷新
	if len(w.backendTraceBatchBuffer) >= w.config.BatchSize {
		ctx := context.Background()
		if err := w.batchWriteBackendTraceLogDirectly(ctx, w.backendTraceBatchBuffer); err != nil {
			logger.Error("Failed to write ClickHouse full backend trace batch", "error", err, "count", len(w.backendTraceBatchBuffer))
			return err
		}
		w.backendTraceBatchBuffer = w.backendTraceBatchBuffer[:0]
	}

	return nil
}

// addToBatch 将日志添加到批量缓冲区
func (w *ClickHouseWriter) addToBatch(log *types.AccessLog) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.batchBuffer = append(w.batchBuffer, log)

	// 如果缓冲区满了，立即刷新
	if len(w.batchBuffer) >= w.config.BatchSize {
		ctx := context.Background()
		if err := w.batchWriteDirectly(ctx, w.batchBuffer); err != nil {
			logger.Error("Failed to write ClickHouse full batch", "error", err, "count", len(w.batchBuffer))
			return err
		}
		w.batchBuffer = w.batchBuffer[:0]
	}

	return nil
}

// batchWriteDirectly 直接批量写入日志到ClickHouse
func (w *ClickHouseWriter) batchWriteDirectly(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	startTime := time.Now()

	// 使用数据库的批量插入方法
	_, err := w.db.BatchInsert(ctx, "HUB_GW_ACCESS_LOG", logs, true)
	if err != nil {
		return fmt.Errorf("failed to write ClickHouse log batch: %w", err)
	}

	// 更新计数器
	w.insertedCount += uint64(len(logs))
	w.batchCount++

	// 记录性能指标
	duration := time.Since(startTime)
	recordsPerSecond := float64(len(logs)) / duration.Seconds()

	logger.Debug("ClickHouse batch write completed",
		"count", len(logs),
		"duration", duration,
		"recordsPerSecond", recordsPerSecond,
		"totalInserted", w.insertedCount,
		"totalBatches", w.batchCount)

	return nil
}
