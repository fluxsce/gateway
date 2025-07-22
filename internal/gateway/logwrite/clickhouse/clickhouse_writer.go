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
		config:      config,
		db:          db,
		stopChan:    make(chan struct{}),
		batchBuffer: make([]*types.AccessLog, 0, config.BatchSize),
	}

	// 如果启用异步日志，初始化异步处理
	if config.IsAsyncLogging() {
		writer.logQueue = make(chan *types.AccessLog, config.AsyncQueueSize)
		writer.startAsyncProcessor()
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

	// 执行批量写入
	err := w.batchWriteDirectly(ctx, w.batchBuffer)
	if err != nil {
		logger.Error("Failed to flush ClickHouse batch buffer", "error", err, "count", len(w.batchBuffer))
		return err
	}

	// 清空缓冲区
	w.batchBuffer = w.batchBuffer[:0]
	logger.Debug("Flushed ClickHouse batch buffer", "count", len(w.batchBuffer))

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
