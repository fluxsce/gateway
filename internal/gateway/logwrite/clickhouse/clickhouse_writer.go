package clickhouse

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/logwrite/asyncq"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
	"gateway/pkg/database/sqlutils"
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

	// closed 仅拒绝新的 Write；Flush 在关停排空后仍可执行。
	closed atomic.Bool

	// batchSize 实际批量条数。ClickHouse 写入宜偏大，但不写回共享 LogConfig。
	batchSize int

	// ClickHouse特定的计数器
	insertedCount atomic.Uint64
	batchCount    atomic.Uint64
}

const minClickHouseBatchSize = 5000

// NewClickHouseWriter 创建一个新的ClickHouse日志写入器
func NewClickHouseWriter(config *types.LogConfig) (*ClickHouseWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 获取ClickHouse数据库连接
	db := database.GetConnection("clickhouse_main")
	if db == nil {
		return nil, fmt.Errorf("failed to get clickhouse_main database connection")
	}

	batch := effectiveBatchSize(config)
	writer := &ClickHouseWriter{
		config:                  config,
		db:                      db,
		stopChan:                make(chan struct{}),
		batchSize:               batch,
		batchBuffer:             make([]*types.AccessLog, 0, batch),
		backendTraceBatchBuffer: make([]*types.BackendTraceLog, 0, batch),
	}

	// 如果启用异步日志，初始化异步处理
	if config.IsAsyncLogging() {
		size := types.QueueSize(config)
		writer.logQueue = make(chan *types.AccessLog, size)
		writer.backendTraceLogQueue = make(chan *types.BackendTraceLog, size)
		writer.startAsyncProcessor()
		writer.startBackendTraceAsyncProcessor()
	}

	// 启动定时刷新
	writer.startFlushTimer()

	return writer, nil
}

// UpdateAccessLog 按租户与 trace 单次 ALTER UPDATE：仅刷新重放结果态列，resetCount +1，parentTraceId 置空（不重放自增 retryCount）。
// 端口重放使用预设 trace，约定主表已有对应行；成功返回 1（不依赖 RowsAffected）。
func (w *ClickHouseWriter) UpdateAccessLog(ctx context.Context, log *types.AccessLog) (int64, error) {
	if w.closed.Load() {
		return 0, fmt.Errorf("writer is closed")
	}
	if log.TenantID == "" || log.TraceID == "" {
		return 0, fmt.Errorf("tenantId and traceId required for replay update")
	}

	where := "tenantId = ? AND traceId = ?"
	whereArgs := []interface{}{log.TenantID, log.TraceID}

	statePatch := types.NewAccessLogReplayStatePatch(log)
	setClause, setArgs, err := sqlutils.BuildUpdateQuery("HUB_GW_ACCESS_LOG", &statePatch, true)
	if err != nil {
		return 0, fmt.Errorf("replay update build set: %w", err)
	}
	resetCountAndParent := "resetCount = coalesce(resetCount, 0) + 1, parentTraceId = ''"
	var fullSet string
	if setClause == "" {
		fullSet = resetCountAndParent
	} else {
		fullSet = setClause + ", " + resetCountAndParent
	}

	query := "ALTER TABLE HUB_GW_ACCESS_LOG UPDATE " + fullSet + " WHERE " + where
	execArgs := append(setArgs, whereArgs...)

	if _, err := w.db.Exec(ctx, query, execArgs, true); err != nil {
		return 0, fmt.Errorf("replay update: %w", err)
	}
	// 变更常为异步，驱动多返回 0 行；成功即视为已更新，避免误判走 Insert
	return 1, nil
}

// Write 写入单条访问日志
func (w *ClickHouseWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.logQueue, log) {
			return nil
		}
		logger.Warn("ClickHouse log queue is full, dropping log entry", "traceId", log.TraceID)
		return fmt.Errorf("log queue is full")
	}

	if w.config.IsBatchProcessing() {
		return w.addToBatch(log)
	}
	return w.writeDirectly(ctx, log)
}

// BatchWrite 批量写入多条访问日志
func (w *ClickHouseWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.logQueue, log) {
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
	batch := asyncq.Take(&w.mutex, &w.batchBuffer, w.batchSize)
	if len(batch) == 0 {
		return nil
	}

	err := w.batchWriteDirectly(ctx, batch)
	if err != nil {
		logger.Error("Failed to flush ClickHouse batch buffer, dumping failed batch data", "error", err, "count", len(batch))
		for i, log := range batch {
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
		return err
	}

	logger.Debug("Flushed ClickHouse batch buffer", "count", len(batch))
	return nil
}

// Close 关闭写入器
func (w *ClickHouseWriter) Close() error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}

	// 发送停止信号
	close(w.stopChan)

	// 等待异步处理goroutine结束
	w.wg.Wait()

	ctx, cancel := asyncq.WriteContext(0)
	defer cancel()
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
		"totalInserted", w.insertedCount.Load(),
		"totalBatches", w.batchCount.Load())
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
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.backendTraceLogQueue, log) {
			return nil
		}
		logger.Warn("ClickHouse backend trace log queue is full, dropping log entry", "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
		return fmt.Errorf("backend trace log queue is full")
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

	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.backendTraceLogQueue, log) {
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
	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	_, err := w.db.Insert(write, log.TableName(), log, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log: %w", err)
	}

	w.insertedCount.Add(1)
	return nil
}

// batchWriteBackendTraceLogDirectly 直接批量写入后端追踪日志到ClickHouse
func (w *ClickHouseWriter) batchWriteBackendTraceLogDirectly(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	startTime := time.Now()

	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	tableName := logs[0].TableName()
	_, err := w.db.BatchInsert(write, tableName, logs, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log batch: %w", err)
	}

	w.insertedCount.Add(uint64(len(logs)))
	w.batchCount.Add(1)

	duration := time.Since(startTime)
	recordsPerSecond := float64(len(logs)) / duration.Seconds()

	logger.Debug("ClickHouse backend trace log batch write completed",
		"count", len(logs),
		"duration", duration,
		"recordsPerSecond", recordsPerSecond,
		"totalInserted", w.insertedCount.Load(),
		"totalBatches", w.batchCount.Load())

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
				if w.config.IsBatchProcessing() {
					w.addToBatch(log)
				} else {
					ctx := context.Background()
					if err := w.writeDirectly(ctx, log); err != nil {
						logger.Error("Failed to write ClickHouse log in async mode", "error", err, "traceId", log.TraceID)
					}
				}

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
	intervalMs := types.FlushIntervalMs(w.config)
	w.flushTicker = time.NewTicker(time.Duration(intervalMs) * time.Millisecond)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("ClickHouse flush timer stopped")

		logger.Info("ClickHouse flush timer started", "intervalMs", intervalMs)

		for {
			select {
			case <-w.flushTicker.C:
				ctx, cancel := asyncq.WriteContext(0)
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled ClickHouse flush failed", "error", err)
				}
				if err := w.FlushBackendTrace(ctx); err != nil {
					logger.Error("Scheduled ClickHouse backend trace flush failed", "error", err)
				}
				cancel()

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
			if w.config.IsBatchProcessing() {
				w.addToBatch(log)
			} else {
				ctx := context.Background()
				if err := w.writeDirectly(ctx, log); err != nil {
					logger.Error("Failed to write ClickHouse log while draining queue", "error", err, "traceId", log.TraceID)
				}
			}
			count++

		default:
			ctx := context.Background()
			if err := w.Flush(ctx); err != nil {
				logger.Error("Failed to flush ClickHouse during queue drain", "error", err)
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
	batch := asyncq.Take(&w.backendTraceMutex, &w.backendTraceBatchBuffer, w.batchSize)
	if len(batch) == 0 {
		return nil
	}

	err := w.batchWriteBackendTraceLogDirectly(ctx, batch)
	if err != nil {
		logger.Error("Failed to flush ClickHouse backend trace batch buffer, dumping failed batch data", "error", err, "count", len(batch))
		for i, log := range batch {
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
		return err
	}

	logger.Debug("Flushed ClickHouse backend trace batch buffer", "count", len(batch))
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
			ctx := context.Background()
			if err := w.FlushBackendTrace(ctx); err != nil {
				logger.Error("Failed to flush ClickHouse backend trace during queue drain", "error", err)
			}
			logger.Info("ClickHouse backend trace queue drained", "processedCount", count)
			return
		}
	}
}

// addBackendTraceToBatch 将后端追踪日志添加到批量缓冲区
func (w *ClickHouseWriter) addBackendTraceToBatch(log *types.BackendTraceLog) error {
	limit := w.batchSize
	batch := asyncq.AppendTakeIfFull(&w.backendTraceMutex, &w.backendTraceBatchBuffer, log, limit, limit)
	if len(batch) == 0 {
		return nil
	}
	ctx := context.Background()
	if err := w.batchWriteBackendTraceLogDirectly(ctx, batch); err != nil {
		logger.Error("Failed to write ClickHouse full backend trace batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

// addToBatch 将日志添加到批量缓冲区
func (w *ClickHouseWriter) addToBatch(log *types.AccessLog) error {
	limit := w.batchSize
	batch := asyncq.AppendTakeIfFull(&w.mutex, &w.batchBuffer, log, limit, limit)
	if len(batch) == 0 {
		return nil
	}
	ctx := context.Background()
	if err := w.batchWriteDirectly(ctx, batch); err != nil {
		logger.Error("Failed to write ClickHouse full batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

// batchWriteDirectly 直接批量写入日志到ClickHouse
func (w *ClickHouseWriter) batchWriteDirectly(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	startTime := time.Now()

	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	// 使用数据库的批量插入方法
	_, err := w.db.BatchInsert(write, "HUB_GW_ACCESS_LOG", logs, true)
	if err != nil {
		return fmt.Errorf("failed to write ClickHouse log batch: %w", err)
	}

	w.insertedCount.Add(uint64(len(logs)))
	w.batchCount.Add(1)

	duration := time.Since(startTime)
	recordsPerSecond := float64(len(logs)) / duration.Seconds()

	logger.Debug("ClickHouse batch write completed",
		"count", len(logs),
		"duration", duration,
		"recordsPerSecond", recordsPerSecond,
		"totalInserted", w.insertedCount.Load(),
		"totalBatches", w.batchCount.Load())

	return nil
}

func (w *ClickHouseWriter) writeDirectly(ctx context.Context, log *types.AccessLog) error {
	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	_, err := w.db.Insert(write, "HUB_GW_ACCESS_LOG", log, true)
	if err != nil {
		return fmt.Errorf("failed to write ClickHouse log: %w", err)
	}
	w.insertedCount.Add(1)
	return nil
}

// effectiveBatchSize 计算 ClickHouse 实际批量条数，不修改传入配置。
func effectiveBatchSize(config *types.LogConfig) int {
	n := types.BatchLimit(config)
	if n < minClickHouseBatchSize {
		return minClickHouseBatchSize
	}
	return n
}
