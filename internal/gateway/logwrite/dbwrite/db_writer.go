package dbwrite

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

// DBWriter 实现了 LogWriter 接口，用于将网关访问日志写入数据库
// 支持同步/异步写入模式，具备缓存批量提交功能
//
// 主要特性:
//   - 支持同步和异步日志写入模式
//   - 异步模式下使用内存队列缓存日志
//   - 支持批量写入提高性能
//   - 定时刷新机制确保日志及时写入
//   - 线程安全的并发操作
//   - 优雅关闭确保数据不丢失
type DBWriter struct {
	// 日志配置，包含异步和批量写入配置
	config *types.LogConfig

	// 数据库连接实例
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
}

// NewDBWriter 创建一个新的数据库日志写入器
//
// 创建过程:
//  1. 获取数据库连接
//  2. 根据配置决定是否启用异步模式
//  3. 启动异步处理goroutine（如果启用异步）
//  4. 启动定时刷新机制
//
// 参数:
//   - config: 日志配置，包含异步和批量处理参数
//
// 返回:
//   - *DBWriter: 数据库日志写入器实例
//   - error: 创建失败时返回错误信息
func NewDBWriter(config *types.LogConfig) (*DBWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to get default database connection")
	}

	batch := types.BatchLimit(config)
	writer := &DBWriter{
		config:                  config,
		db:                      db,
		stopChan:                make(chan struct{}),
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

	// 启动定时刷新（无论同步还是异步模式都需要）
	writer.startFlushTimer()

	return writer, nil
}

// FlushBackendTrace 刷新后端追踪日志缓冲区，将缓存的日志写入数据库
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 刷新失败时返回错误信息
func (w *DBWriter) FlushBackendTrace(ctx context.Context) error {
	batch := asyncq.Take(&w.backendTraceMutex, &w.backendTraceBatchBuffer, types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}

	err := w.batchWriteBackendTraceDirectly(ctx, batch)
	if err != nil {
		logger.Error("Failed to flush backend trace batch buffer, dumping failed batch data", "error", err, "count", len(batch))
		for i, log := range batch {
			logger.Warn("Failed backend trace batch item",
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

	logger.Debug("Flushed backend trace batch buffer", "count", len(batch))
	return nil
}

// UpdateAccessLog 按租户与 trace 单次 UPDATE：仅刷新重放结果态列，resetCount 库侧 +1，parentTraceId 置空（不重放自增 retryCount）。
// 不做存在性预查；重放路径始终同步执行，不进入异步队列。
// 返回 SQL 受影响行数，0 时由调用方决定是否 Insert。
func (w *DBWriter) UpdateAccessLog(ctx context.Context, log *types.AccessLog) (int64, error) {
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
	resetCountAndParent := "resetCount = COALESCE(resetCount, 0) + 1, parentTraceId = ?"
	var fullSet string
	if setClause == "" {
		fullSet = resetCountAndParent
	} else {
		fullSet = setClause + ", " + resetCountAndParent
	}
	setArgs = append(setArgs, "")

	query := "UPDATE HUB_GW_ACCESS_LOG SET " + fullSet + " WHERE " + where
	execArgs := append(setArgs, whereArgs...)

	rows, err := w.db.Exec(ctx, query, execArgs, true)
	if err != nil {
		return 0, fmt.Errorf("replay update: %w", err)
	}
	return rows, nil
}

// Write 写入单条访问日志
// 根据配置决定是同步写入数据库还是放入异步队列
//
// 参数:
//   - ctx: 上下文，用于控制超时和取消
//   - log: 要写入的访问日志
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *DBWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.logQueue, log) {
			return nil
		}
		logger.Warn("Log queue is full, dropping log entry", "traceId", log.TraceID)
		return fmt.Errorf("log queue is full")
	}

	// 同步模式：直接写入数据库或缓存批量写入
	if w.config.IsBatchProcessing() {
		return w.addToBatch(log)
	}

	// 直接写入数据库
	return w.writeDirectly(ctx, log)
}

// BatchWrite 批量写入多条访问日志
//
// 参数:
//   - ctx: 上下文
//   - logs: 要写入的日志数组
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *DBWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.logQueue, log) {
				logger.Warn("Log queue is full, dropping log entry", "traceId", log.TraceID)
			}
		}
		return nil
	}

	// 同步模式：直接批量写入数据库
	return w.batchWriteDirectly(ctx, logs)
}

// Flush 刷新缓冲区，将缓存的日志写入数据库
//
// 参数:
//   - ctx: 上下文
//
// 返回:
//   - error: 刷新失败时返回错误信息
func (w *DBWriter) Flush(ctx context.Context) error {
	batch := asyncq.Take(&w.mutex, &w.batchBuffer, types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}

	err := w.batchWriteDirectly(ctx, batch)
	if err != nil {
		logger.Error("Failed to flush batch buffer, dumping failed batch data", "error", err, "count", len(batch))
		for i, log := range batch {
			logger.Warn("Failed batch item",
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

	logger.Debug("Flushed batch buffer", "count", len(batch))
	return nil
}

// Close 关闭写入器，优雅停止异步处理并刷新缓冲区
//
// 返回:
//   - error: 关闭失败时返回错误信息
func (w *DBWriter) Close() error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}

	// 发送停止信号
	close(w.stopChan)

	// 等待异步处理goroutine结束
	w.wg.Wait()

	// 刷新剩余的缓冲区数据
	ctx, cancel := asyncq.WriteContext(0)
	defer cancel()
	if err := w.Flush(ctx); err != nil {
		logger.Error("Failed to flush buffer during close", "error", err)
	}
	if err := w.FlushBackendTrace(ctx); err != nil {
		logger.Error("Failed to flush backend trace buffer during close", "error", err)
	}

	// 关闭定时器
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}

	logger.Info("DBWriter closed successfully")
	return nil
}

// GetLogConfig 获取日志配置
func (w *DBWriter) GetLogConfig() *types.LogConfig {
	return w.config
}

// startAsyncProcessor 启动异步日志处理goroutine
func (w *DBWriter) startAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Async log processor stopped")

		logger.Info("Async log processor started")

		for {
			select {
			case log := <-w.logQueue:
				if w.config.IsBatchProcessing() {
					w.addToBatch(log)
				} else {
					ctx := context.Background()
					if err := w.writeDirectly(ctx, log); err != nil {
						logger.Error("Failed to write log in async mode", "error", err, "traceId", log.TraceID)
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
func (w *DBWriter) startFlushTimer() {
	intervalMs := types.FlushIntervalMs(w.config)
	w.flushTicker = time.NewTicker(time.Duration(intervalMs) * time.Millisecond)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Flush timer stopped")

		logger.Info("Flush timer started", "intervalMs", intervalMs)

		for {
			select {
			case <-w.flushTicker.C:
				ctx, cancel := asyncq.WriteContext(0)
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled flush failed", "error", err)
				}
				if err := w.FlushBackendTrace(ctx); err != nil {
					logger.Error("Scheduled backend trace flush failed", "error", err)
				}
				cancel()

			case <-w.stopChan:
				return
			}
		}
	}()
}

// drainQueue 排空队列中剩余的日志
func (w *DBWriter) drainQueue() {
	logger.Info("Draining log queue")
	count := 0

	for {
		select {
		case log := <-w.logQueue:
			if w.config.IsBatchProcessing() {
				w.addToBatch(log)
			} else {
				ctx := context.Background()
				if err := w.writeDirectly(ctx, log); err != nil {
					logger.Error("Failed to write log while draining queue", "error", err, "traceId", log.TraceID)
				}
			}
			count++

		default:
			ctx := context.Background()
			if err := w.Flush(ctx); err != nil {
				logger.Error("Failed to flush during queue drain", "error", err)
			}
			logger.Info("Queue drained", "processedCount", count)
			return
		}
	}
}

// startBackendTraceAsyncProcessor 启动后端追踪日志异步处理goroutine
func (w *DBWriter) startBackendTraceAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Async backend trace log processor stopped")

		logger.Info("Async backend trace log processor started")

		for {
			select {
			case log := <-w.backendTraceLogQueue:
				if w.config.IsBatchProcessing() {
					w.addBackendTraceToBatch(log)
				} else {
					ctx := context.Background()
					if err := w.writeBackendTraceDirectly(ctx, log); err != nil {
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
func (w *DBWriter) drainBackendTraceQueue() {
	logger.Info("Draining backend trace log queue")
	count := 0

	for {
		select {
		case log := <-w.backendTraceLogQueue:
			if w.config.IsBatchProcessing() {
				w.addBackendTraceToBatch(log)
			} else {
				ctx := context.Background()
				if err := w.writeBackendTraceDirectly(ctx, log); err != nil {
					logger.Error("Failed to write backend trace log while draining queue", "error", err, "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
				}
			}
			count++

		default:
			ctx := context.Background()
			if err := w.FlushBackendTrace(ctx); err != nil {
				logger.Error("Failed to flush backend trace during queue drain", "error", err)
			}
			logger.Info("Backend trace queue drained", "processedCount", count)
			return
		}
	}
}

// addBackendTraceToBatch 将后端追踪日志添加到批量缓冲区
func (w *DBWriter) addBackendTraceToBatch(log *types.BackendTraceLog) error {
	limit := types.BatchLimit(w.config)
	batch := asyncq.AppendTakeIfFull(&w.backendTraceMutex, &w.backendTraceBatchBuffer, log, limit, limit)
	if len(batch) == 0 {
		return nil
	}
	ctx := context.Background()
	if err := w.batchWriteBackendTraceDirectly(ctx, batch); err != nil {
		logger.Error("Failed to write full backend trace batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

// addToBatch 将日志添加到批量缓冲区
func (w *DBWriter) addToBatch(log *types.AccessLog) error {
	limit := types.BatchLimit(w.config)
	batch := asyncq.AppendTakeIfFull(&w.mutex, &w.batchBuffer, log, limit, limit)
	if len(batch) == 0 {
		return nil
	}
	ctx := context.Background()
	if err := w.batchWriteDirectly(ctx, batch); err != nil {
		logger.Error("Failed to write full batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

// writeDirectly 直接写入单条日志到数据库
func (w *DBWriter) writeDirectly(ctx context.Context, log *types.AccessLog) error {
	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	_, err := w.db.Insert(write, "HUB_GW_ACCESS_LOG", log, true)
	if err != nil {
		return fmt.Errorf("failed to write log: %w", err)
	}
	return nil
}

// batchWriteDirectly 直接批量写入日志到数据库
func (w *DBWriter) batchWriteDirectly(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	// 使用数据库的批量插入方法，自动处理SQL构建和事务提交
	_, err := w.db.BatchInsert(write, "HUB_GW_ACCESS_LOG", logs, true)
	if err != nil {
		return fmt.Errorf("failed to write log batch: %w", err)
	}

	return nil
}

// writeBackendTraceDirectly 直接写入单条后端追踪日志到数据库
func (w *DBWriter) writeBackendTraceDirectly(ctx context.Context, log *types.BackendTraceLog) error {
	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	_, err := w.db.Insert(write, log.TableName(), log, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log: %w", err)
	}
	return nil
}

// batchWriteBackendTraceDirectly 直接批量写入后端追踪日志到数据库
func (w *DBWriter) batchWriteBackendTraceDirectly(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 使用数据库的批量插入方法
	tableName := types.BackendTraceLogTableName
	if len(logs) > 0 {
		tableName = logs[0].TableName()
	}
	write, cancel := asyncq.EnsureWriteCtx(ctx)
	defer cancel()
	_, err := w.db.BatchInsert(write, tableName, logs, true)
	if err != nil {
		return fmt.Errorf("failed to write backend trace log batch: %w", err)
	}

	return nil
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
func (w *DBWriter) WriteBackendTraceLog(ctx context.Context, log *types.BackendTraceLog) error {
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.backendTraceLogQueue, log) {
			return nil
		}
		logger.Warn("Backend trace log queue is full, dropping log entry", "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
		return fmt.Errorf("backend trace log queue is full")
	}

	// 同步模式：直接写入数据库或缓存批量写入
	if w.config.IsBatchProcessing() {
		return w.addBackendTraceToBatch(log)
	}

	// 直接写入数据库
	return w.writeBackendTraceDirectly(ctx, log)
}

// BatchWriteBackendTraceLog 批量写入后端追踪日志（从表）
//
// 参数:
//   - ctx: 上下文
//   - logs: 要写入的日志数组
//
// 返回:
//   - error: 写入失败时返回错误信息
func (w *DBWriter) BatchWriteBackendTraceLog(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}

	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.backendTraceLogQueue, log) {
				logger.Warn("Backend trace log queue is full, dropping log entry", "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
			}
		}
		return nil
	}

	// 同步模式：直接批量写入数据库
	return w.batchWriteBackendTraceDirectly(ctx, logs)
}
