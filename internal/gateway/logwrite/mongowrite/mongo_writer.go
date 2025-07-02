package mongowrite

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gohub/internal/gateway/logwrite/types"
	"gohub/pkg/logger"
)

// MongoWriter MongoDB日志写入器
// 支持异步批量写入、连接池管理和自动重连
type MongoWriter struct {
	// 配置
	config      *types.LogConfig
	mongoConfig *types.MongoDBOutputConfig
	
	// 批量写入控制
	buffer      []interface{}
	bufferMutex sync.Mutex
	
	// 异步处理
	logChan   chan *types.AccessLog
	batchChan chan []interface{}
	
	// 控制协程
	wg        sync.WaitGroup
	closeChan chan struct{}
	
	// 连接状态
	connected bool
	connMutex sync.RWMutex
}

// NewMongoWriter 创建新的MongoDB日志写入器
func NewMongoWriter(config *types.LogConfig) (*MongoWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	mongoConfig, err := config.GetMongoConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse mongo config: %w", err)
	}

	// 设置默认值
	if mongoConfig.ConnectTimeout == 0 {
		mongoConfig.ConnectTimeout = 10000 // 10秒
	}
	if mongoConfig.BatchSize == 0 {
		mongoConfig.BatchSize = 100
	}
	if mongoConfig.Database == "" {
		mongoConfig.Database = "gateway_logs"
	}
	if mongoConfig.Collection == "" {
		mongoConfig.Collection = "access_logs"
	}

	writer := &MongoWriter{
		config:      config,
		mongoConfig: mongoConfig,
		buffer:      make([]interface{}, 0, mongoConfig.BatchSize),
		logChan:     make(chan *types.AccessLog, config.AsyncQueueSize),
		batchChan:   make(chan []interface{}, 100),
		closeChan:   make(chan struct{}),
	}

	// 启动异步处理协程
	writer.startWorkers()

	logger.Info("MongoDB writer created (implementation pending)", 
		"database", mongoConfig.Database, 
		"collection", mongoConfig.Collection)

	return writer, nil
}

// Write 写入单条日志
func (w *MongoWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if !w.mongoConfig.AsyncInsert {
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

	if !w.mongoConfig.AsyncInsert {
		// 同步批量写入
		return w.insertMany(ctx, logs)
	}

	// 异步批量写入
	documents := make([]interface{}, len(logs))
	for i, log := range logs {
		documents[i] = w.logToDocument(log)
	}

	select {
	case w.batchChan <- documents:
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
	documents := make([]interface{}, len(w.buffer))
	copy(documents, w.buffer)
	w.buffer = w.buffer[:0]

	// 执行批量插入
	return w.insertManyDocuments(ctx, documents)
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

// insertOne 插入单条文档
func (w *MongoWriter) insertOne(ctx context.Context, log *types.AccessLog) error {
	// TODO: 实现MongoDB单条插入
	logger.Debug("MongoDB single insert (placeholder)", "trace_id", log.TraceID)
	return nil
}

// insertMany 批量插入文档
func (w *MongoWriter) insertMany(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	documents := make([]interface{}, len(logs))
	for i, log := range logs {
		documents[i] = w.logToDocument(log)
	}

	return w.insertManyDocuments(ctx, documents)
}

// insertManyDocuments 批量插入文档
func (w *MongoWriter) insertManyDocuments(ctx context.Context, documents []interface{}) error {
	if len(documents) == 0 {
		return nil
	}

	// TODO: 实现MongoDB批量插入
	logger.Debug("MongoDB batch insert (placeholder)", "count", len(documents))
	return nil
}

// logToDocument 将访问日志转换为MongoDB文档
func (w *MongoWriter) logToDocument(log *types.AccessLog) interface{} {
	// 创建MongoDB文档
	doc := map[string]interface{}{
		"tenant_id":                     log.TenantID,
		"trace_id":                      log.TraceID,
		"gateway_instance_id":           log.GatewayInstanceID,
		"gateway_node_ip":               log.GatewayNodeIP,
		"route_config_id":               log.RouteConfigID,
		"service_definition_id":         log.ServiceDefinitionID,
		"log_config_id":                 log.LogConfigID,
		"request_method":                log.RequestMethod,
		"request_path":                  log.RequestPath,
		"request_query":                 log.RequestQuery,
		"request_size":                  log.RequestSize,
		"request_headers":               log.RequestHeaders,
		"request_body":                  log.RequestBody,
		"client_ip_address":             log.ClientIPAddress,
		"client_port":                   log.ClientPort,
		"user_agent":                    log.UserAgent,
		"referer":                       log.Referer,
		"user_identifier":               log.UserIdentifier,
		"gateway_received_time":         log.GatewayReceivedTime,
		"gateway_start_processing_time": log.GatewayStartProcessingTime,
		"backend_request_start_time":    log.BackendRequestStartTime,
		"backend_response_received_time": log.BackendResponseReceivedTime,
		"gateway_finished_processing_time": log.GatewayFinishedProcessingTime,
		"gateway_response_sent_time":    log.GatewayResponseSentTime,
		"total_processing_time_ms":      log.TotalProcessingTimeMs,
		"gateway_processing_time_ms":    log.GatewayProcessingTimeMs,
		"backend_response_time_ms":      log.BackendResponseTimeMs,
		"network_latency_ms":            log.NetworkLatencyMs,
		"gateway_status_code":           log.GatewayStatusCode,
		"backend_status_code":           log.BackendStatusCode,
		"response_size":                 log.ResponseSize,
		"response_headers":              log.ResponseHeaders,
		"response_body":                 log.ResponseBody,
		"matched_route":                 log.MatchedRoute,
		"forward_address":               log.ForwardAddress,
		"forward_method":                log.ForwardMethod,
		"forward_params":                log.ForwardParams,
		"forward_headers":               log.ForwardHeaders,
		"forward_body":                  log.ForwardBody,
		"load_balancer_decision":        log.LoadBalancerDecision,
		"error_message":                 log.ErrorMessage,
		"error_code":                    log.ErrorCode,
		"parent_trace_id":               log.ParentTraceID,
		"reset_flag":                    log.ResetFlag,
		"retry_count":                   log.RetryCount,
		"reset_count":                   log.ResetCount,
		"log_level":                     log.LogLevel,
		"log_type":                      log.LogType,
		"ext_property":                  log.ExtProperty,
		"add_time":                      log.AddTime,
		"add_who":                       log.AddWho,
		"edit_time":                     log.EditTime,
		"edit_who":                      log.EditWho,
		"opr_seq_flag":                  log.OprSeqFlag,
		"current_version":               log.CurrentVersion,
		"active_flag":                   log.ActiveFlag,
		"note_text":                     log.NoteText,
	}

	// 添加索引字段
	doc["created_at"] = time.Now()
	doc["date"] = time.Now().Format("2006-01-02")
	doc["hour"] = time.Now().Format("2006-01-02-15")

	return doc
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
			w.buffer = append(w.buffer, w.logToDocument(log))
			shouldFlush := len(w.buffer) >= w.mongoConfig.BatchSize
			w.bufferMutex.Unlock()

			if shouldFlush {
				w.Flush(context.Background())
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
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := w.insertManyDocuments(ctx, documents); err != nil {
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
			w.Flush(context.Background())

		case <-w.closeChan:
			return
		}
	}
}

// isConnectionError 判断是否为连接错误
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	
	// 检查常见的连接错误
	errStr := err.Error()
	connectionErrors := []string{
		"connection refused",
		"no reachable servers",
		"server selection timeout",
		"context deadline exceeded",
		"connection reset by peer",
		"broken pipe",
	}

	for _, connErr := range connectionErrors {
		if strings.Contains(errStr, connErr) {
			return true
		}
	}

	return false
} 