package console

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
)

const (
	minConsoleQueueSize = 100
	consoleDrainWait    = 3 * time.Second
)

// ConsoleWriter 控制台访问日志写入器。
// 有界队列 + 单消费者写 stdout：Write 只投递，队列满丢弃，不在请求路径堵 stdout。
type ConsoleWriter struct {
	config    *types.LogConfig
	formatter Formatter
	output    io.Writer

	queue     chan string
	stop      chan struct{}
	done      chan struct{}
	dropped   atomic.Uint64
	async     bool
	writeMu   sync.Mutex
	closeOnce sync.Once
}

// Formatter 定义格式化器接口
type Formatter interface {
	Format(log *types.AccessLog) string
}

// NewConsoleWriter 创建新的控制台日志写入器
func NewConsoleWriter(config *types.LogConfig) (*ConsoleWriter, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	writer := &ConsoleWriter{
		config: config,
		output: os.Stdout,
		async:  config.IsAsyncLogging(),
	}

	// 根据配置选择格式化器
	switch types.LogFormat(config.LogFormat) {
	case types.LogFormatJSON:
		writer.formatter = &JSONFormatter{config: config}
	case types.LogFormatText:
		writer.formatter = &TextFormatter{config: config}
	case types.LogFormatCSV:
		writer.formatter = &CSVFormatter{config: config}
	default:
		writer.formatter = &TextFormatter{config: config}
	}

	if writer.async {
		size := config.AsyncQueueSize
		if size < minConsoleQueueSize {
			size = minConsoleQueueSize
		}
		writer.queue = make(chan string, size)
		writer.stop = make(chan struct{})
		writer.done = make(chan struct{})
		go writer.loop()
	}

	return writer, nil
}

// UpdateAccessLog 控制台为只追加输出，无法按行更新；返回 0 表示未更新，由上层决定是否 Insert。
func (w *ConsoleWriter) UpdateAccessLog(ctx context.Context, log *types.AccessLog) (int64, error) {
	_ = ctx
	_ = log
	return 0, nil
}

// Write 投递一条访问日志。异步时不阻塞调用方；队列满返回 nil 并计数丢弃。
func (w *ConsoleWriter) Write(ctx context.Context, log *types.AccessLog) error {
	_ = ctx
	if log == nil {
		return nil
	}
	return w.offer(w.formatter.Format(log))
}

// BatchWrite 批量写入日志（简化实现，逐条写入）
func (w *ConsoleWriter) BatchWrite(ctx context.Context, logs []*types.AccessLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 控制台输出不需要复杂的批量优化，直接逐条写入
	for _, log := range logs {
		if err := w.Write(ctx, log); err != nil {
			return err
		}
	}

	return nil
}

// Flush 刷新输出缓冲区
func (w *ConsoleWriter) Flush(ctx context.Context) error {
	// 控制台输出通常是立即的，无需特殊刷新
	return nil
}

// Close 停止投递协程，不关闭 stdout。
func (w *ConsoleWriter) Close() error {
	w.closeOnce.Do(func() {
		if !w.async {
			return
		}
		close(w.stop)
		select {
		case <-w.done:
		case <-time.After(consoleDrainWait):
			logger.Warn("等待控制台访问日志队列退出超时")
		}
	})
	return nil
}

// GetLogConfig 获取日志配置
func (w *ConsoleWriter) GetLogConfig() *types.LogConfig {
	return w.config
}

func (w *ConsoleWriter) offer(line string) error {
	if !w.async {
		return w.writeLine(line)
	}
	select {
	case <-w.stop:
		w.recordDrop()
		return nil
	default:
	}
	select {
	case w.queue <- line:
		return nil
	case <-w.stop:
		w.recordDrop()
		return nil
	default:
		w.recordDrop()
		return nil
	}
}

func (w *ConsoleWriter) recordDrop() {
	n := w.dropped.Add(1)
	if n == 1 || n%1000 == 0 {
		logger.Warn("控制台访问日志队列已满，丢弃新日志", "dropped", n, "queueCap", cap(w.queue))
	}
}

func (w *ConsoleWriter) loop() {
	defer close(w.done)
	for {
		select {
		case line := <-w.queue:
			_ = w.writeLine(line)
		case <-w.stop:
			w.drain()
			return
		}
	}
}

func (w *ConsoleWriter) drain() {
	deadline := time.Now().Add(consoleDrainWait)
	for {
		if time.Now().After(deadline) {
			return
		}
		select {
		case line := <-w.queue:
			_ = w.writeLine(line)
		default:
			return
		}
	}
}

func (w *ConsoleWriter) writeLine(line string) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	_, err := fmt.Fprintln(w.output, line)
	return err
}

// WriteBackendTraceLog 投递一条后端追踪日志，背压规则与 Write 相同。
func (w *ConsoleWriter) WriteBackendTraceLog(ctx context.Context, log *types.BackendTraceLog) error {
	_ = ctx
	if log == nil {
		return nil
	}
	return w.offer(w.formatBackendTraceLog(log))
}

// BatchWriteBackendTraceLog 批量写入后端追踪日志（从表）
func (w *ConsoleWriter) BatchWriteBackendTraceLog(ctx context.Context, logs []*types.BackendTraceLog) error {
	if len(logs) == 0 {
		return nil
	}

	// 控制台输出不需要复杂的批量优化，直接逐条写入
	for _, log := range logs {
		if err := w.WriteBackendTraceLog(ctx, log); err != nil {
			return err
		}
	}

	return nil
}

// formatBackendTraceLog 格式化后端追踪日志
func (w *ConsoleWriter) formatBackendTraceLog(log *types.BackendTraceLog) string {
	// 使用JSON格式输出后端追踪日志
	jsonStr, err := log.ToJSON()
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to format backend trace log: %s"}`, err.Error())
	}
	return jsonStr
}

// JSONFormatter JSON格式化器
type JSONFormatter struct {
	config *types.LogConfig
}

func (f *JSONFormatter) Format(log *types.AccessLog) string {
	jsonStr, err := log.ToJSON(f.config)
	if err != nil {
		return fmt.Sprintf(`{"error": "failed to format log: %s"}`, err.Error())
	}
	return jsonStr
}

// TextFormatter 文本格式化器 - 类似Nginx访问日志格式
type TextFormatter struct {
	config *types.LogConfig
}

func (f *TextFormatter) Format(log *types.AccessLog) string {
	// 添加颜色支持
	statusColor := getStatusColor(log.GatewayStatusCode)
	methodColor := getMethodColor(log.RequestMethod)
	resetColor := "\033[0m"

	return fmt.Sprintf("%s[%s]%s %s%s %s%s %s%d%s %dB %dms %s",
		"\033[36m", // 时间颜色 - 青色
		log.GatewayStartProcessingTime.Format("2006-01-02 15:04:05.000"),
		resetColor,
		methodColor,
		log.RequestMethod,
		resetColor,
		log.RequestPath,
		statusColor,
		log.GatewayStatusCode,
		resetColor,
		log.ResponseSize,
		log.TotalProcessingTimeMs,
		log.ClientIPAddress,
	)
}

// CSVFormatter CSV格式化器
type CSVFormatter struct {
	config *types.LogConfig
}

func (f *CSVFormatter) Format(log *types.AccessLog) string {
	return log.ToCSV(f.config)
}

// getStatusColor 根据HTTP状态码返回颜色代码
func getStatusColor(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "\033[32m" // 绿色 - 成功
	case statusCode >= 300 && statusCode < 400:
		return "\033[33m" // 黄色 - 重定向
	case statusCode >= 400 && statusCode < 500:
		return "\033[31m" // 红色 - 客户端错误
	case statusCode >= 500:
		return "\033[35m" // 紫色 - 服务器错误
	default:
		return "\033[0m" // 默认颜色
	}
}

// getMethodColor 根据HTTP方法返回颜色代码
func getMethodColor(method string) string {
	switch method {
	case "GET":
		return "\033[34m" // 蓝色
	case "POST":
		return "\033[32m" // 绿色
	case "PUT":
		return "\033[33m" // 黄色
	case "DELETE":
		return "\033[31m" // 红色
	case "PATCH":
		return "\033[36m" // 青色
	default:
		return "\033[0m" // 默认颜色
	}
}
