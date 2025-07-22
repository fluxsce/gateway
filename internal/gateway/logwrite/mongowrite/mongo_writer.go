package mongowrite

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/mongo/client"
	"gateway/pkg/mongo/factory"
	"gateway/pkg/mongo/utils"
)

// MongoWriter MongoDB日志写入器
// 支持异步批量写入、连接池管理和自动重连
type MongoWriter struct {
	// 配置
	config *types.LogConfig

	// MongoDB连接
	mongoClient *client.Client

	// 批量写入控制
	buffer      []*types.AccessLog
	bufferMutex sync.Mutex

	// 异步处理
	logChan   chan *types.AccessLog
	batchChan chan []*types.AccessLog

	// 控制协程
	wg        sync.WaitGroup
	closeChan chan struct{}
}

// NewMongoWriter 创建新的MongoDB日志写入器
func NewMongoWriter(config *types.LogConfig) (*MongoWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	// 获取默认MongoDB连接
	mongoClient, err := factory.GetDefaultConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get default MongoDB connection: %w", err)
	}

	writer := &MongoWriter{
		config:      config,
		mongoClient: mongoClient,
		buffer:      make([]*types.AccessLog, 0, 100), // 默认批量大小为100
		logChan:     make(chan *types.AccessLog, config.AsyncQueueSize),
		batchChan:   make(chan []*types.AccessLog, 100),
		closeChan:   make(chan struct{}),
	}

	// 启动异步处理协程
	writer.startWorkers()

	// 创建临时 AccessLog 实例以获取表名
	var accessLog types.AccessLog
	logger.Info("MongoDB writer created successfully",
		"collection", accessLog.TableName())

	return writer, nil
}

// Write 写入单条日志
func (w *MongoWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if !w.config.IsAsyncLogging() {
		// 同步写入
		return w.insertOne(ctx, log)
	}

	// 异步写入
	select {
	case w.logChan <- log:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("log channel is full")
	}
}

// BatchWrite 批量写入日志
func (w *MongoWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	if !w.config.IsAsyncLogging() {
		// 同步批量写入
		return w.insertMany(ctx, logs)
	}

	// 异步批量写入
	select {
	case w.batchChan <- logs:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("batch channel is full")
	}
}

// Flush 刷新缓冲区
func (w *MongoWriter) Flush(ctx context.Context) error {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	if len(w.buffer) == 0 {
		return nil
	}

	// 复制缓冲区数据
	documents := make([]*types.AccessLog, len(w.buffer))
	copy(documents, w.buffer)
	w.buffer = w.buffer[:0]

	// 执行批量插入
	return w.insertMany(ctx, documents)
}

// Close 关闭写入器
func (w *MongoWriter) Close() error {
	close(w.closeChan)
	w.wg.Wait()

	// 刷新剩余缓冲区
	if err := w.Flush(context.Background()); err != nil {
		logger.Error("Failed to flush buffer on close", "error", err)
	}

	logger.Info("MongoDB writer closed")
	return nil
}

// GetLogConfig 获取日志配置
func (w *MongoWriter) GetLogConfig() *types.LogConfig {
	return w.config
}

// insertOne 插入单条文档
func (w *MongoWriter) insertOne(ctx context.Context, log *types.AccessLog) error {
	// 使用公共转换方法将结构体转换为 Document
	doc, err := utils.ConvertToDocument(log)
	if err != nil {
		return fmt.Errorf("failed to convert log to document: %w", err)
	}

	// 获取默认数据库和集合 - 使用 AccessLog 的表名作为集合名称
	var accessLog types.AccessLog
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(accessLog.TableName())

	// 直接插入转换后的文档
	result, err := collection.InsertOne(ctx, doc, nil)
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	logger.Debug("MongoDB single document inserted",
		"trace_id", log.TraceID,
		"inserted_id", result.InsertedID)

	return nil
}

// insertMany 批量插入文档
func (w *MongoWriter) insertMany(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 使用公共转换方法批量转换结构体为 Document
	documents, err := utils.ConvertToDocuments(logs)
	if err != nil {
		return fmt.Errorf("failed to convert logs to documents: %w", err)
	}

	// 获取默认数据库和集合 - 使用 AccessLog 的表名作为集合名称
	var accessLog types.AccessLog
	database, err := w.mongoClient.DefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}
	collection := database.Collection(accessLog.TableName())

	// 批量插入文档
	result, err := collection.InsertMany(ctx, documents, nil)
	if err != nil {
		return fmt.Errorf("failed to insert documents: %w", err)
	}

	logger.Debug("MongoDB batch documents inserted",
		"count", len(documents),
		"inserted_ids_count", len(result.InsertedIDs))

	return nil
}

// startWorkers 启动工作协程
func (w *MongoWriter) startWorkers() {
	// 启动单条日志处理协程
	w.wg.Add(1)
	go w.singleLogWorker()

	// 启动批量日志处理协程
	w.wg.Add(1)
	go w.batchLogWorker()

	// 启动缓冲区刷新协程
	w.wg.Add(1)
	go w.bufferFlushWorker()
}

// singleLogWorker 单条日志处理协程
func (w *MongoWriter) singleLogWorker() {
	defer w.wg.Done()

	for {
		select {
		case log := <-w.logChan:
			w.bufferMutex.Lock()
			w.buffer = append(w.buffer, log)
			shouldFlush := len(w.buffer) >= 100 // 默认批量大小
			w.bufferMutex.Unlock()

			if shouldFlush {
				if err := w.Flush(context.Background()); err != nil {
					logger.Error("Failed to flush buffer", "error", err)
				}
			}

		case <-w.closeChan:
			return
		}
	}
}

// batchLogWorker 批量日志处理协程
func (w *MongoWriter) batchLogWorker() {
	defer w.wg.Done()

	for {
		select {
		case documents := <-w.batchChan:
			ctx, cancel := context.WithTimeout(context.Background(), w.mongoClient.GetConfig().SocketTimeoutMS)
			if err := w.insertMany(ctx, documents); err != nil {
				logger.Error("Failed to insert batch documents", "error", err, "count", len(documents))
			}
			cancel()

		case <-w.closeChan:
			return
		}
	}
}

// bufferFlushWorker 缓冲区刷新协程
func (w *MongoWriter) bufferFlushWorker() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Second) // 每5秒刷新一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := w.Flush(context.Background()); err != nil {
				logger.Error("Failed to flush buffer on timer", "error", err)
			}

		case <-w.closeChan:
			return
		}
	}
}
