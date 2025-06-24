package logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gohub/pkg/config"
	huberrors "gohub/pkg/utils/huberrors"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// global logger instance
	log *zap.Logger
)

const (
	// TraceIDKey 跟踪ID在上下文中的键名
	TraceIDKey = "trace_id"
	// UserIdKey 用户ID在上下文中的键名
	UserIdKey  = "userId"
	// UserNameKey 用户名在上下文中的键名
	UserNameKey = "userName"
	// TenantIdKey 租户ID在上下文中的键名
	TenantIdKey = "tenantId"
)

// LoggerConfig 日志配置结构体
type LoggerConfig struct {
	// Level 日志级别
	Level string `mapstructure:"level"`
	// Encoding 编码格式
	Encoding string `mapstructure:"encoding"`
	// ShowCaller 是否显示调用者信息
	ShowCaller bool `mapstructure:"show_caller"`
	// StacktraceLevel 显示堆栈跟踪的最小级别
	StacktraceLevel string `mapstructure:"stacktrace_level"`

	// DefaultOutput 默认输出路径
	DefaultOutput string `mapstructure:"default_output"`
	// ErrorOutput 错误日志输出路径
	ErrorOutput string `mapstructure:"error_output"`
	// InfoOutput 信息日志输出路径
	InfoOutput string `mapstructure:"info_output"`
	// DebugOutput 调试日志输出路径
	DebugOutput string `mapstructure:"debug_output"`

	// LogPath 日志文件的根目录，当使用相对路径时会与此路径结合
	LogPath string `mapstructure:"log_path"`
	// MaxSize 单个日志文件最大尺寸(MB)
	MaxSize int `mapstructure:"max_size"`
	// MaxBackups 保留的旧日志文件最大数量
	MaxBackups int `mapstructure:"max_backups"`
	// MaxAge 保留的旧日志文件最大天数
	MaxAge int `mapstructure:"max_age"`
	// Compress 是否压缩旧日志文件
	Compress bool `mapstructure:"compress"`
}

// Setup 设置日志，从配置文件加载
// 返回: 可能的错误
func Setup() error {
	// 尝试从配置中读取日志配置
	var logConfig LoggerConfig
	if config.IsExist("log") {
		err := config.GetSection("log", &logConfig)
		if err != nil {
			// 配置读取失败，使用默认配置
			return Init(nil)
		}
		// 配置读取成功，使用配置初始化
		return Init(&logConfig)
	}

	// 没有配置，使用默认值
	return Init(nil)
}

// Init 初始化日志
// config: 日志配置
// 返回: 可能的错误
func Init(config *LoggerConfig) error {
	if config == nil {
		config = &LoggerConfig{
			Level:           "info",
			DefaultOutput:   "stdout",
			Encoding:        "json",
			ShowCaller:      true,
			StacktraceLevel: "warn",
			LogPath:         "./logs",
			MaxSize:         100,
			MaxBackups:      10,
			MaxAge:          30,
			Compress:        true,
		}
	}

	// 设置日志级别
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// 解析堆栈跟踪级别
	stacktraceLevel, err := zapcore.ParseLevel(config.StacktraceLevel)
	if err != nil {
		stacktraceLevel = zapcore.WarnLevel
	}

	// 确保日志路径存在
	if config.LogPath != "" && config.LogPath != "stdout" && config.LogPath != "stderr" {
		if err := ensureDir(config.LogPath); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	// 设置编码器
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 创建多输出
	var cores []zapcore.Core

	// 默认输出
	if defaultWriter := getWriteSyncer(config.DefaultOutput, config.LogPath); defaultWriter != nil {
		defaultCore := zapcore.NewCore(encoder, defaultWriter, level)
		cores = append(cores, defaultCore)
	}

	// 错误日志输出
	if config.ErrorOutput != "" && config.ErrorOutput != config.DefaultOutput {
		errorWriter := getWriteSyncer(config.ErrorOutput, config.LogPath)
		if errorWriter != nil {
			errorCore := zapcore.NewCore(
				encoder,
				errorWriter,
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl >= zapcore.ErrorLevel && lvl >= level
				}),
			)
			cores = append(cores, errorCore)
		}
	}

	// 信息日志输出
	if config.InfoOutput != "" && config.InfoOutput != config.DefaultOutput {
		infoWriter := getWriteSyncer(config.InfoOutput, config.LogPath)
		if infoWriter != nil {
			infoCore := zapcore.NewCore(
				encoder,
				infoWriter,
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl == zapcore.InfoLevel && lvl >= level
				}),
			)
			cores = append(cores, infoCore)
		}
	}

	// 调试日志输出
	if config.DebugOutput != "" && config.DebugOutput != config.DefaultOutput {
		debugWriter := getWriteSyncer(config.DebugOutput, config.LogPath)
		if debugWriter != nil {
			debugCore := zapcore.NewCore(
				encoder,
				debugWriter,
				zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
					return lvl == zapcore.DebugLevel && lvl >= level
				}),
			)
			cores = append(cores, debugCore)
		}
	}

	// 合并多个core
	core := zapcore.NewTee(cores...)

	// 添加选项
	options := []zap.Option{}
	if config.ShowCaller {
		options = append(options, zap.AddCaller())
	}
	options = append(options, zap.AddStacktrace(stacktraceLevel))

	// 创建日志实例
	log = zap.New(core, options...)
	return nil
}

// getWriteSyncer 根据输出路径创建相应的写入器
func getWriteSyncer(output string, logPath string) zapcore.WriteSyncer {
	if output == "" {
		return nil
	}

	if output == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	if output == "stderr" {
		return zapcore.AddSync(os.Stderr)
	}

	// 检查是否使用绝对路径
	useAbsolutePath := false
	if config.IsExist("log.use_absolute_path") {
		useAbsolutePath = config.GetBool("log.use_absolute_path", false)
	}

	// 处理相对路径，除非启用了绝对路径模式
	if !useAbsolutePath && logPath != "" && !filepath.IsAbs(output) {
		output = filepath.Join(logPath, output)
	}

	// 确保日志目录存在
	logDir := filepath.Dir(output)
	if err := ensureDir(logDir); err != nil {
		fmt.Printf("创建日志目录 %s 失败: %v，将使用 stdout 代替\n", logDir, err)
		return zapcore.AddSync(os.Stdout)
	}

	// 尝试打开文件
	file, err := os.OpenFile(output, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		// 文件打开失败，使用stdout
		fmt.Printf("打开日志文件 %s 失败: %v，将使用 stdout 代替\n", output, err)
		return zapcore.AddSync(os.Stdout)
	}

	return zapcore.AddSync(file)
}

// ensureDir 确保目录存在
func ensureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// ===== 基础接口 =====

// Info 记录信息级别日志
// 支持多种参数格式：
// 1. Info("消息")
// 2. Info("消息", err) - 当第二个参数是error类型时
// 3. Info("消息", map[string]any{...}) - 当第二个参数是map时
// 4. Info("消息", "key1", val1, "key2", val2, ...) - 键值对列表
// 5. Info("消息", zap.String(...), zap.Int(...), ...) - zap.Field列表
func Info(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Info(msg, parseArgs(args...)...)
}

// InfoWithTrace 记录带跟踪ID的信息级别日志
func InfoWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Info(msg, fields...)
}

// Debug 记录调试级别日志
// 参数格式与Info相同
func Debug(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Debug(msg, parseArgs(args...)...)
}

// DebugWithTrace 记录带跟踪ID的调试级别日志
func DebugWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Debug(msg, fields...)
}

// Warn 记录警告级别日志
// 参数格式与Info相同
func Warn(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Warn(msg, parseArgs(args...)...)
}

// WarnWithTrace 记录带跟踪ID的警告级别日志
func WarnWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Warn(msg, fields...)
}

// Error 记录错误级别日志
// 参数格式与Info相同，当第一个参数是error类型时自动设置error字段
func Error(msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Error(msg, err)模式
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用huberrors获取完整错误栈信息
			errorStack := huberrors.ErrorStack(err)
			log.Error(msg, zap.Error(err), zap.String("error_stack", errorStack))
			return
		}
	}

	// 添加堆栈信息
	fields := parseArgs(args...)
	stack := captureStack(2)
	fields = append(fields, zap.String("error_stack", stack))
	log.Error(msg, fields...)
}

// ErrorWithTrace 记录带跟踪ID的错误级别日志
func ErrorWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Error(msg, err)模式
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用huberrors获取完整错误栈信息
			errorStack := huberrors.ErrorStack(err)
			fields := []zap.Field{zap.Error(err), zap.String("error_stack", errorStack)}
			fields = appendTraceID(ctx, fields)
			log.Error(msg, fields...)
			return
		}
	}

	// 添加堆栈信息
	fields := parseArgs(args...)
	stack := captureStack(2)
	fields = append(fields, zap.String("error_stack", stack))
	fields = appendTraceID(ctx, fields)
	log.Error(msg, fields...)
}

// Fatal 记录致命错误日志并终止程序
// 参数格式与Error相同
func Fatal(msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Fatal(msg, err)模式
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用增强的堆栈跟踪
			stack := captureErrorStack(err)
			log.Fatal(msg, zap.Error(err), zap.String("error_stack", stack))
			return
		}
	}

	// 添加堆栈信息
	fields := parseArgs(args...)
	stack := captureStack(2)
	fields = append(fields, zap.String("error_stack", stack))
	log.Fatal(msg, fields...)
}

// FatalWithTrace 记录带跟踪ID的致命错误日志并终止程序
func FatalWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Fatal(msg, err)模式
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用增强的堆栈跟踪
			stack := captureErrorStack(err)
			fields := []zap.Field{zap.Error(err), zap.String("error_stack", stack)}
			fields = appendTraceID(ctx, fields)
			log.Fatal(msg, fields...)
			return
		}
	}

	// 添加堆栈信息
	fields := parseArgs(args...)
	stack := captureStack(2)
	fields = append(fields, zap.String("error_stack", stack))
	fields = appendTraceID(ctx, fields)
	log.Fatal(msg, fields...)
}

// ===== 工具函数 =====

// parseArgs 解析日志参数，支持多种格式
func parseArgs(args ...any) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	// 检查是否为单个map参数
	if len(args) == 1 {
		if m, ok := args[0].(map[string]any); ok {
			return mapToFields(m)
		}

		// 检查是否为zap.Field切片
		if fields, ok := args[0].([]zap.Field); ok {
			return fields
		}

		// 检查是否为单个zap.Field
		if field, ok := args[0].(zap.Field); ok {
			return []zap.Field{field}
		}
	}

	// 检查是否为键值对列表 (key1, val1, key2, val2, ...)
	if len(args) > 1 && len(args)%2 == 0 {
		allStringKeys := true
		for i := 0; i < len(args); i += 2 {
			if _, ok := args[i].(string); !ok {
				allStringKeys = false
				break
			}
		}

		if allStringKeys {
			fields := make([]zap.Field, 0, len(args)/2)
			for i := 0; i < len(args); i += 2 {
				key := args[i].(string)
				fields = append(fields, zap.Any(key, args[i+1]))
			}
			return fields
		}
	}

	// 检查是否为zap.Field列表
	allFields := true
	for _, arg := range args {
		if _, ok := arg.(zap.Field); !ok {
			allFields = false
			break
		}
	}

	if allFields {
		fields := make([]zap.Field, len(args))
		for i, arg := range args {
			fields[i] = arg.(zap.Field)
		}
		return fields
	}

	// 无法识别的格式，打包为一个字段
	return []zap.Field{zap.Any("args", args)}
}

// mapToFields 将map转换为zap字段列表
func mapToFields(data map[string]any) []zap.Field {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// captureStack 获取调用堆栈的详细信息
// 参数:
//   - skip: 要跳过的调用帧数
//
// 返回:
//   - string: 格式化的堆栈信息
func captureStack(skip int) string {
	var buffer strings.Builder

	// 获取最多32层堆栈信息
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	// 遍历栈帧
	for {
		frame, more := frames.Next()

		// 跳过标准库和zap库的帧
		if !strings.Contains(frame.File, "runtime/") &&
			!strings.Contains(frame.File, "zap/") &&
			!strings.Contains(frame.File, "zapcore/") {
			// 构建文件位置信息
			file := frame.File
			line := frame.Line
			function := frame.Function

			// 提取短函数名
			if idx := strings.LastIndex(function, "/"); idx >= 0 {
				function = function[idx+1:]
			}

			// 将信息追加到缓冲区
			fmt.Fprintf(&buffer, "%s:%d %s\n", file, line, function)
		}

		if !more {
			break
		}
	}

	return buffer.String()
}

// captureErrorStack 获取错误对象的完整堆栈信息
// 参数:
//   - err: 错误对象
//
// 返回:
//   - string: 格式化的堆栈信息
func captureErrorStack(err error) string {
	// 使用huberrors包的ErrorStack函数获取完整的错误链和调用栈
	// 它能够处理嵌套错误并保留原始错误信息
	return huberrors.ErrorStack(err)
}

// CreateLogDirectory 创建日志目录
func CreateLogDirectory(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
	}
	return nil
}

// ===== 跟踪ID相关函数 =====

// appendTraceID 为日志字段添加跟踪ID
func appendTraceID(ctx context.Context, fields []zap.Field) []zap.Field {
	if ctx == nil {
		return fields
	}
	
	if traceID := getTraceIDFromContext(ctx); traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}
	
	// 添加用户ID
	if userID, ok := ctx.Value(UserIdKey).(string); ok && userID != "" {
		fields = append(fields, zap.String("userId", userID))
	}
	
	// 添加用户名
	if userName, ok := ctx.Value(UserNameKey).(string); ok && userName != "" {
		fields = append(fields, zap.String("userName", userName))
	}
	
	// 添加租户ID
	if tenantID, ok := ctx.Value(TenantIdKey).(string); ok && tenantID != "" {
		fields = append(fields, zap.String("tenantId", tenantID))
	}
	
	return fields
}

// getTraceIDFromContext 从上下文中获取跟踪ID
func getTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTraceID 为上下文添加跟踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}

// GetTraceID 从上下文中获取跟踪ID（公开接口）
func GetTraceID(ctx context.Context) string {
	return getTraceIDFromContext(ctx)
}
