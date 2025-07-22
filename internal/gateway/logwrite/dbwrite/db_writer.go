package dbwrite

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/database"
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

	// 状态标识
	closed bool
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
	// 获取默认数据库连接
	db := database.GetDefaultConnection()
	if db == nil {
		return nil, fmt.Errorf("failed to get default database connection")
	}

	writer := &DBWriter{
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

	// 启动定时刷新（无论同步还是异步模式都需要）
	writer.startFlushTimer()

	return writer, nil
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
			logger.Warn("Log queue is full, dropping log entry", "traceId", log.TraceID)
			return fmt.Errorf("log queue is full")
		}
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
		logger.Error("Failed to flush batch buffer", "error", err, "count", len(w.batchBuffer))
		return err
	}

	// 清空缓冲区
	w.batchBuffer = w.batchBuffer[:0]
	logger.Debug("Flushed batch buffer", "count", len(w.batchBuffer))

	return nil
}

// Close 关闭写入器，优雅停止异步处理并刷新缓冲区
//
// 返回:
//   - error: 关闭失败时返回错误信息
func (w *DBWriter) Close() error {
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
		logger.Error("Failed to flush buffer during close", "error", err)
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
	if w.config.AsyncFlushIntervalMs <= 0 {
		return
	}

	w.flushTicker = time.NewTicker(time.Duration(w.config.AsyncFlushIntervalMs) * time.Millisecond)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		defer logger.Info("Flush timer stopped")

		logger.Info("Flush timer started", "intervalMs", w.config.AsyncFlushIntervalMs)

		for {
			select {
			case <-w.flushTicker.C:
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled flush failed", "error", err)
				}

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
			// 队列为空，执行最终刷新确保缓冲区数据写入
			if count > 0 {
				ctx := context.Background()
				if err := w.Flush(ctx); err != nil {
					logger.Error("Failed to flush during queue drain", "error", err)
				}
			}
			logger.Info("Queue drained", "processedCount", count)
			return
		}
	}
}

// addToBatch 将日志添加到批量缓冲区
func (w *DBWriter) addToBatch(log *types.AccessLog) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.batchBuffer = append(w.batchBuffer, log)

	// 如果缓冲区满了，立即刷新
	if len(w.batchBuffer) >= w.config.BatchSize {
		ctx := context.Background()
		if err := w.batchWriteDirectly(ctx, w.batchBuffer); err != nil {
			logger.Error("Failed to write full batch", "error", err, "count", len(w.batchBuffer))
			return err
		}
		w.batchBuffer = w.batchBuffer[:0]
	}

	return nil
}

// writeDirectly 直接写入单条日志到数据库
func (w *DBWriter) writeDirectly(ctx context.Context, log *types.AccessLog) error {
	_, err := w.db.Insert(ctx, "HUB_GW_ACCESS_LOG", log, true)
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

	// 使用数据库的批量插入方法，自动处理SQL构建和事务提交
	_, err := w.db.BatchInsert(ctx, "HUB_GW_ACCESS_LOG", logs, true)
	if err != nil {
		return fmt.Errorf("failed to write log batch: %w", err)
	}

	return nil
}
