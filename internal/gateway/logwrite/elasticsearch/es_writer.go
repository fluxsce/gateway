package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gohub/internal/gateway/logwrite/types"
	"gohub/pkg/logger"
)

// ESWriter Elasticsearch日志写入器
// 支持批量索引、索引模板管理和连接池
type ESWriter struct {
	// 配置
	config   *types.LogConfig
	esConfig *types.ElasticsearchOutputConfig
	
	// 批量写入控制
	buffer      []ESDocument
	bufferMutex sync.Mutex
	
	// 异步处理
	logChan   chan *types.AccessLog
	batchChan chan []ESDocument
	
	// 控制协程
	wg        sync.WaitGroup
	closeChan chan struct{}
	
	// 索引管理
	currentIndex string
	indexMutex   sync.RWMutex
}

// ESDocument Elasticsearch文档结构
type ESDocument struct {
	Index  string      `json:"_index"`
	Type   string      `json:"_type,omitempty"`
	ID     string      `json:"_id,omitempty"`
	Source interface{} `json:"_source"`
}

// BulkRequest 批量请求结构
type BulkRequest struct {
	Index ESIndexMeta `json:"index"`
}

// ESIndexMeta 索引元数据
type ESIndexMeta struct {
	Index string `json:"_index"`
	Type  string `json:"_type,omitempty"`
}

// NewESWriter 创建新的Elasticsearch日志写入器
func NewESWriter(config *types.LogConfig) (*ESWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	esConfig, err := config.GetElasticsearchConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse elasticsearch config: %w", err)
	}

	// 设置默认值
	if len(esConfig.Addresses) == 0 {
		esConfig.Addresses = []string{"http://localhost:9200"}
	}
	if esConfig.BatchSize == 0 {
		esConfig.BatchSize = 100
	}
	if esConfig.IndexName == "" {
		esConfig.IndexName = "gateway-logs"
	}
	if esConfig.IndexPattern == "" {
		esConfig.IndexPattern = "daily"
	}

	writer := &ESWriter{
		config:    config,
		esConfig:  esConfig,
		buffer:    make([]ESDocument, 0, esConfig.BatchSize),
		logChan:   make(chan *types.AccessLog, config.AsyncQueueSize),
		batchChan: make(chan []ESDocument, 100),
		closeChan: make(chan struct{}),
	}

	// 初始化当前索引
	writer.updateCurrentIndex()

	// 启动异步处理协程
	writer.startWorkers()

	return writer, nil
}

// Write 写入单条日志
func (w *ESWriter) Write(ctx context.Context, log *types.AccessLog) error {
	if !w.esConfig.AsyncInsert {
		// 同步写入
		return w.indexOne(ctx, log)
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
func (w *ESWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	if !w.esConfig.AsyncInsert {
		// 同步批量写入
		return w.bulkIndex(ctx, logs)
	}

	// 异步批量写入
	documents := make([]ESDocument, len(logs))
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
func (w *ESWriter) Flush(ctx context.Context) error {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	if len(w.buffer) == 0 {
		return nil
	}

	// 复制缓冲区数据
	documents := make([]ESDocument, len(w.buffer))
	copy(documents, w.buffer)
	w.buffer = w.buffer[:0]

	// 执行批量索引
	return w.bulkIndexDocuments(ctx, documents)
}

// Close 关闭写入器
func (w *ESWriter) Close() error {
	close(w.closeChan)
	w.wg.Wait()

	// 刷新剩余缓冲区
	if err := w.Flush(context.Background()); err != nil {
		logger.Error("Failed to flush buffer on close", "error", err)
	}

	return nil
}

// indexOne 索引单条文档
func (w *ESWriter) indexOne(ctx context.Context, log *types.AccessLog) error {
	// TODO: 实现Elasticsearch单条索引
	logger.Info("Elasticsearch single index not implemented yet")
	return nil
}

// bulkIndex 批量索引文档
func (w *ESWriter) bulkIndex(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	documents := make([]ESDocument, len(logs))
	for i, log := range logs {
		documents[i] = w.logToDocument(log)
	}

	return w.bulkIndexDocuments(ctx, documents)
}

// bulkIndexDocuments 批量索引文档
func (w *ESWriter) bulkIndexDocuments(ctx context.Context, documents []ESDocument) error {
	if len(documents) == 0 {
		return nil
	}

	// 构建批量请求体
	var buf bytes.Buffer
	for _, doc := range documents {
		// 索引操作行
		indexOp := BulkRequest{
			Index: ESIndexMeta{
				Index: doc.Index,
				Type:  "_doc",
			},
		}
		
		indexOpBytes, err := json.Marshal(indexOp)
		if err != nil {
			return fmt.Errorf("failed to marshal index operation: %w", err)
		}
		
		buf.Write(indexOpBytes)
		buf.WriteByte('\n')
		
		// 文档数据行
		docBytes, err := json.Marshal(doc.Source)
		if err != nil {
			return fmt.Errorf("failed to marshal document: %w", err)
		}
		
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	// TODO: 发送到Elasticsearch
	logger.Info("Elasticsearch bulk index not implemented yet", 
		"documents", len(documents), 
		"size", buf.Len())
	
	return nil
}

// logToDocument 将访问日志转换为ES文档
func (w *ESWriter) logToDocument(log *types.AccessLog) ESDocument {
	// 更新当前索引（如果需要）
	w.updateCurrentIndex()
	
	// 创建文档数据
	source := map[string]interface{}{
		"@timestamp":                    log.GatewayReceivedTime,
		"tenant_id":                     log.TenantID,
		"trace_id":                      log.TraceID,
		"gateway_instance_id":           log.GatewayInstanceID,
		"gateway_node_ip":               log.GatewayNodeIP,
		"route_config_id":               log.RouteConfigID,
		"service_definition_id":         log.ServiceDefinitionID,
		"log_config_id":                 log.LogConfigID,
		"request": map[string]interface{}{
			"method":  log.RequestMethod,
			"path":    log.RequestPath,
			"query":   log.RequestQuery,
			"size":    log.RequestSize,
			"headers": log.RequestHeaders,
			"body":    log.RequestBody,
		},
		"client": map[string]interface{}{
			"ip":         log.ClientIPAddress,
			"port":       log.ClientPort,
			"user_agent": log.UserAgent,
			"referer":    log.Referer,
			"user_id":    log.UserIdentifier,
		},
		"timing": map[string]interface{}{
			"gateway_received":              log.GatewayReceivedTime,
			"gateway_start_processing":      log.GatewayStartProcessingTime,
			"backend_request_start":         log.BackendRequestStartTime,
			"backend_response_received":     log.BackendResponseReceivedTime,
			"gateway_finished_processing":   log.GatewayFinishedProcessingTime,
			"gateway_response_sent":         log.GatewayResponseSentTime,
			"total_processing_time_ms":      log.TotalProcessingTimeMs,
			"gateway_processing_time_ms":    log.GatewayProcessingTimeMs,
			"backend_response_time_ms":      log.BackendResponseTimeMs,
			"network_latency_ms":            log.NetworkLatencyMs,
		},
		"response": map[string]interface{}{
			"gateway_status": log.GatewayStatusCode,
			"backend_status": log.BackendStatusCode,
			"size":           log.ResponseSize,
			"headers":        log.ResponseHeaders,
			"body":           log.ResponseBody,
		},
		"forwarding": map[string]interface{}{
			"matched_route":          log.MatchedRoute,
			"forward_address":        log.ForwardAddress,
			"forward_method":         log.ForwardMethod,
			"forward_params":         log.ForwardParams,
			"forward_headers":        log.ForwardHeaders,
			"forward_body":           log.ForwardBody,
			"load_balancer_decision": log.LoadBalancerDecision,
		},
		"error": map[string]interface{}{
			"message": log.ErrorMessage,
			"code":    log.ErrorCode,
		},
		"trace": map[string]interface{}{
			"parent_trace_id": log.ParentTraceID,
		},
		"flags": map[string]interface{}{
			"reset_flag":   log.ResetFlag,
			"retry_count":  log.RetryCount,
			"reset_count":  log.ResetCount,
			"active_flag":  log.ActiveFlag,
		},
		"meta": map[string]interface{}{
			"log_level":       log.LogLevel,
			"log_type":        log.LogType,
			"ext_property":    log.ExtProperty,
			"add_time":        log.AddTime,
			"add_who":         log.AddWho,
			"edit_time":       log.EditTime,
			"edit_who":        log.EditWho,
			"opr_seq_flag":    log.OprSeqFlag,
			"current_version": log.CurrentVersion,
			"note_text":       log.NoteText,
		},
	}

	// 添加分类字段用于聚合分析
	source["status_class"] = getStatusClass(log.GatewayStatusCode)
	source["performance_level"] = log.GetPerformanceLevel()
	source["is_error"] = log.IsServerError() || log.IsClientError()
	source["date"] = log.GatewayReceivedTime.Format("2006-01-02")
	source["hour"] = log.GatewayReceivedTime.Format("2006-01-02-15")

	return ESDocument{
		Index:  w.getCurrentIndex(),
		Source: source,
	}
}

// getCurrentIndex 获取当前索引名称
func (w *ESWriter) getCurrentIndex() string {
	w.indexMutex.RLock()
	defer w.indexMutex.RUnlock()
	return w.currentIndex
}

// updateCurrentIndex 更新当前索引名称
func (w *ESWriter) updateCurrentIndex() {
	var newIndex string
	
	switch w.esConfig.IndexPattern {
	case "daily":
		newIndex = fmt.Sprintf("%s-%s", w.esConfig.IndexName, time.Now().Format("2006.01.02"))
	case "weekly":
		year, week := time.Now().ISOWeek()
		newIndex = fmt.Sprintf("%s-%d.%02d", w.esConfig.IndexName, year, week)
	case "monthly":
		newIndex = fmt.Sprintf("%s-%s", w.esConfig.IndexName, time.Now().Format("2006.01"))
	case "yearly":
		newIndex = fmt.Sprintf("%s-%s", w.esConfig.IndexName, time.Now().Format("2006"))
	default:
		newIndex = w.esConfig.IndexName
	}

	w.indexMutex.Lock()
	if w.currentIndex != newIndex {
		w.currentIndex = newIndex
		logger.Info("Updated Elasticsearch index", "new_index", newIndex)
	}
	w.indexMutex.Unlock()
}

// startWorkers 启动工作协程
func (w *ESWriter) startWorkers() {
	// 启动单条日志处理协程
	w.wg.Add(1)
	go w.singleLogWorker()

	// 启动批量日志处理协程
	w.wg.Add(1)
	go w.batchLogWorker()

	// 启动缓冲区刷新协程
	w.wg.Add(1)
	go w.bufferFlushWorker()

	// 启动索引更新协程
	w.wg.Add(1)
	go w.indexUpdateWorker()
}

// singleLogWorker 单条日志处理协程
func (w *ESWriter) singleLogWorker() {
	defer w.wg.Done()

	for {
		select {
		case log := <-w.logChan:
			w.bufferMutex.Lock()
			w.buffer = append(w.buffer, w.logToDocument(log))
			shouldFlush := len(w.buffer) >= w.esConfig.BatchSize
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
func (w *ESWriter) batchLogWorker() {
	defer w.wg.Done()

	for {
		select {
		case documents := <-w.batchChan:
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := w.bulkIndexDocuments(ctx, documents); err != nil {
				logger.Error("Failed to bulk index documents", "error", err, "count", len(documents))
			}
			cancel()

		case <-w.closeChan:
			return
		}
	}
}

// bufferFlushWorker 缓冲区刷新协程
func (w *ESWriter) bufferFlushWorker() {
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

// indexUpdateWorker 索引更新协程
func (w *ESWriter) indexUpdateWorker() {
	defer w.wg.Done()

	ticker := time.NewTicker(time.Minute) // 每分钟检查一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.updateCurrentIndex()

		case <-w.closeChan:
			return
		}
	}
}

// getStatusClass 获取状态码分类
func getStatusClass(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "success"
	case statusCode >= 300 && statusCode < 400:
		return "redirect"
	case statusCode >= 400 && statusCode < 500:
		return "client_error"
	case statusCode >= 500:
		return "server_error"
	default:
		return "unknown"
	}
} 