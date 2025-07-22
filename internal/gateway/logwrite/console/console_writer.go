package console

import (
	"context"
	"fmt"
	"os"
	"sync"

	"gohub/internal/gateway/logwrite/types"
)

// ConsoleWriter 控制台日志写入器
// 支持彩色输出、格式化输出和并发安全
type ConsoleWriter struct {
	// 配置选项
	config *types.LogConfig
	
	// 输出格式化器
	formatter Formatter
	
	// 并发控制
	mutex sync.Mutex
	
	// 输出目标
	output *os.File
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

	return writer, nil
}

// Write 写入单条日志
func (w *ConsoleWriter) Write(ctx context.Context, log *types.AccessLog) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	formatted := w.formatter.Format(log)
	_, err := fmt.Fprintln(w.output, formatted)
	return err
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

// Close 关闭写入器
func (w *ConsoleWriter) Close() error {
	// 不关闭标准输出
	return nil
}

// GetLogConfig 获取日志配置
func (w *ConsoleWriter) GetLogConfig() *types.LogConfig {
	return w.config
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
		return "\033[0m"  // 默认颜色
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
		return "\033[0m"  // 默认颜色
	}
} 