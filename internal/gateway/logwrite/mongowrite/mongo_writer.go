package mongowrite

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/logwrite/asyncq"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	"gateway/pkg/mongo/factory"
	mongotypes "gateway/pkg/mongo/types"
	"gateway/pkg/mongo/utils"
)

// MongoWriter 将访问日志与后端追踪日志写入 MongoDB。
// 异步时：有界队列 + 批量缓冲 + 定时刷新；关停排空队列后再 Flush。
// UpdateAccessLog（端口重放）始终同步 UpdateOne，不进队列。
type MongoWriter struct {
	config      *types.LogConfig
	mongoClient *client.Client

	logQueue    chan *types.AccessLog
	batchBuffer []*types.AccessLog
	mutex       sync.Mutex

	backendTraceLogQueue    chan *types.BackendTraceLog
	backendTraceBatchBuffer []*types.BackendTraceLog
	backendTraceMutex       sync.Mutex

	flushTicker *time.Ticker
	stopChan    chan struct{}
	wg          sync.WaitGroup
	closed      atomic.Bool
}

// NewMongoWriter 创建 MongoDB 日志写入器。
// 仅在 EnableAsyncLogging=Y 时启动队列消费者；定时刷新按 AsyncFlushIntervalMs。
func NewMongoWriter(config *types.LogConfig) (*MongoWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	mongoClient, err := factory.GetDefaultConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get default MongoDB connection: %w", err)
	}

	batch := types.BatchLimit(config)
	writer := &MongoWriter{
		config:                  config,
		mongoClient:             mongoClient,
		stopChan:                make(chan struct{}),
		batchBuffer:             make([]*types.AccessLog, 0, batch),
		backendTraceBatchBuffer: make([]*types.BackendTraceLog, 0, batch),
	}

	if config.IsAsyncLogging() {
		writer.logQueue = make(chan *types.AccessLog, types.QueueSize(config))
		writer.backendTraceLogQueue = make(chan *types.BackendTraceLog, types.QueueSize(config))
		writer.startAsyncProcessor()
		writer.startBackendTraceAsyncProcessor()
	}
	writer.startFlushTimer()

	var accessLog types.AccessLog
	logger.Info("MongoDB writer created successfully",
		"collection", accessLog.TableName(),
		"async", config.IsAsyncLogging(),
		"batchSize", batch,
		"queueSize", types.QueueSize(config))

	return writer, nil
}

// UpdateAccessLog 按租户与 trace 更新主表文档，$inc resetCount，$set 仅重放结果态字段并清空 parentTraceId（不 $inc retryCount）。
// 同步执行，不经过异步通道。无匹配文档时返回 (0, nil)。
func (w *MongoWriter) UpdateAccessLog(ctx context.Context, log *types.AccessLog) (int64, error) {
	if w.closed.Load() {
		return 0, fmt.Errorf("writer is closed")
	}
	if log.TenantID == "" || log.TraceID == "" {
		return 0, fmt.Errorf("tenantId and traceId required for replay update")
	}

	statePatch := types.NewAccessLogReplayStatePatch(log)
	setDoc, err := utils.ConvertToDocument(statePatch)
	if err != nil {
		return 0, fmt.Errorf("convert access log for replay update: %w", err)
	}
	setDoc["parentTraceId"] = ""

	filter := mongotypes.Filter{
		"tenantId": log.TenantID,
		"traceId":  log.TraceID,
	}
	update := mongotypes.Update{
		"$set": setDoc,
		"$inc": mongotypes.Document{"resetCount": 1},
	}

	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return 0, fmt.Errorf("failed to get default database: %w", err)
	}
	var accessLog types.AccessLog
	collection := database.Collection(accessLog.TableName())

	res, err := collection.UpdateOne(ctx, filter, update, nil)
	if err != nil {
		return 0, fmt.Errorf("mongo replay update: %w", err)
	}
	return res.MatchedCount, nil
}

// Write 写入单条访问日志。异步时入队（满则短等再丢），同步时直接插或进批量缓冲。
func (w *MongoWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}
	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.logQueue, log) {
			return nil
		}
		logger.Warn("Mongo log queue is full, dropping log entry", "traceId", log.TraceID)
		return fmt.Errorf("log queue is full")
	}
	if w.config.IsBatchProcessing() {
		return w.addToBatch(log)
	}
	return w.insertOne(ctx, log)
}

// BatchWrite 批量写入访问日志。异步时逐条入队，同步时一次 InsertMany。
func (w *MongoWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}
	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.logQueue, log) {
				logger.Warn("Mongo log queue is full, dropping log entry", "traceId", log.TraceID)
			}
		}
		return nil
	}
	return w.insertMany(ctx, logs)
}

// Flush 拿走主表缓冲后在锁外 InsertMany，避免慢写堵住入队。
func (w *MongoWriter) Flush(ctx context.Context) error {
	write, cancel := asyncq.EnsureWriteCtxTimeout(ctx, types.BatchTimeout(w.config))
	defer cancel()
	batch := asyncq.Take(&w.mutex, &w.batchBuffer, types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}
	if err := w.insertMany(write, batch); err != nil {
		logger.Error("Failed to flush Mongo access log batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

// Close 拒绝新写入，排空队列并刷新剩余缓冲。
func (w *MongoWriter) Close() error {
	if !w.closed.CompareAndSwap(false, true) {
		return nil
	}
	close(w.stopChan)
	w.wg.Wait()

	ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
	defer cancel()
	if err := w.Flush(ctx); err != nil {
		logger.Error("Failed to flush buffer on close", "error", err)
	}
	if err := w.FlushBackendTrace(ctx); err != nil {
		logger.Error("Failed to flush backend trace buffer on close", "error", err)
	}
	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}
	logger.Info("MongoDB writer closed")
	return nil
}

// GetLogConfig 获取日志配置。
func (w *MongoWriter) GetLogConfig() *types.LogConfig {
	return w.config
}

// WriteBackendTraceLog 写入单条后端追踪日志，背压规则与 Write 相同。
func (w *MongoWriter) WriteBackendTraceLog(ctx context.Context, log *types.BackendTraceLog) error {
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}
	if w.config.IsAsyncLogging() {
		if asyncq.Offer(w.backendTraceLogQueue, log) {
			return nil
		}
		logger.Warn("Mongo backend trace log queue is full, dropping log entry",
			"traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
		return fmt.Errorf("backend trace log queue is full")
	}
	if w.config.IsBatchProcessing() {
		return w.addBackendTraceToBatch(log)
	}
	return w.insertBackendTraceLogOne(ctx, log)
}

// BatchWriteBackendTraceLog 批量写入后端追踪日志。
func (w *MongoWriter) BatchWriteBackendTraceLog(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}
	if w.closed.Load() {
		return fmt.Errorf("writer is closed")
	}
	if w.config.IsAsyncLogging() {
		for _, log := range logs {
			if !asyncq.Offer(w.backendTraceLogQueue, log) {
				logger.Warn("Mongo backend trace log queue is full, dropping log entry",
					"traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
			}
		}
		return nil
	}
	return w.insertBackendTraceLogMany(ctx, logs)
}

// FlushBackendTrace 拿走从表缓冲后在锁外 InsertMany。
func (w *MongoWriter) FlushBackendTrace(ctx context.Context) error {
	write, cancel := asyncq.EnsureWriteCtxTimeout(ctx, types.BatchTimeout(w.config))
	defer cancel()
	batch := asyncq.Take(&w.backendTraceMutex, &w.backendTraceBatchBuffer, types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}
	if err := w.insertBackendTraceLogMany(write, batch); err != nil {
		logger.Error("Failed to flush Mongo backend trace batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

func (w *MongoWriter) addToBatch(log *types.AccessLog) error {
	batch := asyncq.AppendTakeIfFull(&w.mutex, &w.batchBuffer, log, types.BatchLimit(w.config), types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}
	ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
	defer cancel()
	if err := w.insertMany(ctx, batch); err != nil {
		logger.Error("Failed to write full Mongo batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

func (w *MongoWriter) addBackendTraceToBatch(log *types.BackendTraceLog) error {
	batch := asyncq.AppendTakeIfFull(&w.backendTraceMutex, &w.backendTraceBatchBuffer, log, types.BatchLimit(w.config), types.BatchLimit(w.config))
	if len(batch) == 0 {
		return nil
	}
	ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
	defer cancel()
	if err := w.insertBackendTraceLogMany(ctx, batch); err != nil {
		logger.Error("Failed to write full Mongo backend trace batch", "error", err, "count", len(batch))
		return err
	}
	return nil
}

func (w *MongoWriter) startAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case log := <-w.logQueue:
				if w.config.IsBatchProcessing() {
					_ = w.addToBatch(log)
				} else {
					ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
					if err := w.insertOne(ctx, log); err != nil {
						logger.Error("Failed to write Mongo log in async mode", "error", err, "traceId", log.TraceID)
					}
					cancel()
				}
			case <-w.stopChan:
				asyncq.Drain(w.logQueue, func(log *types.AccessLog) {
					if w.config.IsBatchProcessing() {
						_ = w.addToBatch(log)
					} else {
						ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
						if err := w.insertOne(ctx, log); err != nil {
							logger.Error("Failed to write Mongo log while draining", "error", err, "traceId", log.TraceID)
						}
						cancel()
					}
				})
				return
			}
		}
	}()
}

func (w *MongoWriter) startBackendTraceAsyncProcessor() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case log := <-w.backendTraceLogQueue:
				if w.config.IsBatchProcessing() {
					_ = w.addBackendTraceToBatch(log)
				} else {
					ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
					if err := w.insertBackendTraceLogOne(ctx, log); err != nil {
						logger.Error("Failed to write Mongo backend trace in async mode",
							"error", err, "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
					}
					cancel()
				}
			case <-w.stopChan:
				asyncq.Drain(w.backendTraceLogQueue, func(log *types.BackendTraceLog) {
					if w.config.IsBatchProcessing() {
						_ = w.addBackendTraceToBatch(log)
					} else {
						ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
						if err := w.insertBackendTraceLogOne(ctx, log); err != nil {
							logger.Error("Failed to write Mongo backend trace while draining",
								"error", err, "traceId", log.TraceID, "backendTraceId", log.BackendTraceID)
						}
						cancel()
					}
				})
				return
			}
		}
	}()
}

func (w *MongoWriter) startFlushTimer() {
	w.flushTicker = time.NewTicker(time.Duration(types.FlushIntervalMs(w.config)) * time.Millisecond)
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for {
			select {
			case <-w.flushTicker.C:
				ctx, cancel := asyncq.WriteContext(types.BatchTimeout(w.config))
				if err := w.Flush(ctx); err != nil {
					logger.Error("Scheduled Mongo flush failed", "error", err)
				}
				if err := w.FlushBackendTrace(ctx); err != nil {
					logger.Error("Scheduled Mongo backend trace flush failed", "error", err)
				}
				cancel()
			case <-w.stopChan:
				return
			}
		}
	}()
}

func (w *MongoWriter) insertBackendTraceLogOne(ctx context.Context, log *types.BackendTraceLog) error {
	doc, err := utils.ConvertToDocument(log)
	if err != nil {
		return fmt.Errorf("failed to convert backend trace log to document: %w", err)
	}
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(log.TableName())
	if _, err = collection.InsertOne(ctx, doc, nil); err != nil {
		return fmt.Errorf("failed to insert backend trace log: %w", err)
	}
	return nil
}

func (w *MongoWriter) insertBackendTraceLogMany(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}
	documents, err := utils.ConvertToDocuments(logs)
	if err != nil {
		return fmt.Errorf("failed to convert backend trace logs to documents: %w", err)
	}
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(logs[0].TableName())
	if _, err = collection.InsertMany(ctx, documents, nil); err != nil {
		return fmt.Errorf("failed to insert backend trace logs: %w", err)
	}
	return nil
}

func (w *MongoWriter) insertOne(ctx context.Context, log *types.AccessLog) error {
	doc, err := utils.ConvertToDocument(log)
	if err != nil {
		return fmt.Errorf("failed to convert log to document: %w", err)
	}
	var accessLog types.AccessLog
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(accessLog.TableName())
	if _, err = collection.InsertOne(ctx, doc, nil); err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}
	return nil
}

func (w *MongoWriter) insertMany(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}
	documents, err := utils.ConvertToDocuments(logs)
	if err != nil {
		return fmt.Errorf("failed to convert logs to documents: %w", err)
	}
	var accessLog types.AccessLog
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(accessLog.TableName())
	if _, err = collection.InsertMany(ctx, documents, nil); err != nil {
		return fmt.Errorf("failed to insert documents: %w", err)
	}
	return nil
}
