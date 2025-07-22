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
	"gopkg.in/natefinch/lumberjack.v2"
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

// Init 初始化日志系统
// 
// 该函数负责设置全局日志实例，支持多级别、多输出的日志配置
// 主要功能包括：
// 1. 解析日志级别和堆栈跟踪级别
// 2. 创建日志目录
// 3. 配置编码器（JSON或Console格式）
// 4. 设置多个输出目标（默认、错误、信息、调试）
// 5. 集成日志轮转功能
//
// 参数:
//   - config: 日志配置对象，如果为nil则使用默认配置
//
// 返回:
//   - error: 初始化过程中的错误
func Init(config *LoggerConfig) error {
	// 使用默认配置当没有提供配置时
	if config == nil {
		config = &LoggerConfig{
			Level:           "info",        // 默认信息级别
			DefaultOutput:   "stdout",      // 默认输出到标准输出
			Encoding:        "json",        // 默认JSON编码
			ShowCaller:      true,          // 显示调用者信息
			StacktraceLevel: "warn",        // 警告级别及以上显示堆栈
			LogPath:         "./logs",      // 默认日志目录
			MaxSize:         100,           // 默认100MB轮转
			MaxBackups:      10,            // 默认保留10个文件
			MaxAge:          30,            // 默认保留30天
			Compress:        true,          // 默认启用压缩
		}
	}

	// 设置日志级别
	// 支持: debug, info, warn, error, fatal, panic
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel // 解析失败时使用信息级别
	}

	// 解析堆栈跟踪级别
	// 只有达到此级别的日志才会包含堆栈跟踪信息
	stacktraceLevel, err := zapcore.ParseLevel(config.StacktraceLevel)
	if err != nil {
		stacktraceLevel = zapcore.WarnLevel // 解析失败时使用警告级别
	}

	// 确保日志路径存在
	// 只有当使用文件输出时才需要创建目录
	if config.LogPath != "" && config.LogPath != "stdout" && config.LogPath != "stderr" {
		if err := ensureDir(config.LogPath); err != nil {
			return fmt.Errorf("创建日志目录失败: %w", err)
		}
	}

	// 设置编码器
	// 配置时间格式和其他输出格式
	var encoder zapcore.Encoder
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 使用ISO8601时间格式
	
	if config.Encoding == "json" {
		encoder = zapcore.NewJSONEncoder(encoderConfig)    // JSON格式，适合生产环境
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig) // 控制台格式，适合开发环境
	}

	// 创建多输出核心
	// 支持同时输出到多个目标，每个目标可以有不同的日志级别过滤
	var cores []zapcore.Core

	// 默认输出核心
	// 处理所有达到最低级别的日志
	if defaultWriter := getWriteSyncer(config.DefaultOutput, config.LogPath, config); defaultWriter != nil {
		defaultCore := zapcore.NewCore(encoder, defaultWriter, level)
		cores = append(cores, defaultCore)
	}

	// 错误日志输出核心
	// 只处理错误级别及以上的日志，避免重复输出
	if config.ErrorOutput != "" && config.ErrorOutput != config.DefaultOutput {
		errorWriter := getWriteSyncer(config.ErrorOutput, config.LogPath, config)
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

	// 信息日志输出核心
	// 只处理信息级别的日志，实现日志分级存储
	if config.InfoOutput != "" && config.InfoOutput != config.DefaultOutput {
		infoWriter := getWriteSyncer(config.InfoOutput, config.LogPath, config)
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

	// 调试日志输出核心
	// 只处理调试级别的日志，用于开发时的详细信息
	if config.DebugOutput != "" && config.DebugOutput != config.DefaultOutput {
		debugWriter := getWriteSyncer(config.DebugOutput, config.LogPath, config)
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

	// 合并多个核心
	// 使用Tee将所有核心组合成一个，实现多目标输出
	core := zapcore.NewTee(cores...)

	// 添加日志选项
	options := []zap.Option{}
	
	// 添加调用者信息（文件名、行号、函数名）
	if config.ShowCaller {
		options = append(options, zap.AddCaller())
	}
	
	// 添加堆栈跟踪，用于错误诊断
	options = append(options, zap.AddStacktrace(stacktraceLevel))

	// 创建全局日志实例
	log = zap.New(core, options...)
	return nil
}

// getWriteSyncer 根据输出路径创建相应的写入器
// 支持以下输出类型：
// - "stdout": 标准输出
// - "stderr": 标准错误输出  
// - 文件路径: 使用 lumberjack 实现日志轮转的文件输出
// 
// 参数:
//   - output: 输出目标路径或特殊值(stdout/stderr)
//   - logPath: 日志文件根目录，用于相对路径拼接
//   - logConfig: 日志配置对象，包含轮转参数
//
// 返回:
//   - zapcore.WriteSyncer: 可用于zap的写入器，失败时返回nil
func getWriteSyncer(output string, logPath string, logConfig *LoggerConfig) zapcore.WriteSyncer {
	// 空输出路径检查
	if output == "" {
		return nil
	}

	// 处理标准输出
	if output == "stdout" {
		return zapcore.AddSync(os.Stdout)
	}

	// 处理标准错误输出
	if output == "stderr" {
		return zapcore.AddSync(os.Stderr)
	}

	// 检查是否使用绝对路径模式
	// 从全局配置中读取绝对路径配置，默认为false使用相对路径
	useAbsolutePath := false
	if config.IsExist("log.use_absolute_path") {
		useAbsolutePath = config.GetBool("log.use_absolute_path", false)
	}

	// 处理相对路径，除非启用了绝对路径模式
	// 当使用相对路径时，将输出路径与日志根目录拼接
	if !useAbsolutePath && logPath != "" && !filepath.IsAbs(output) {
		output = filepath.Join(logPath, output)
	}

	// 确保日志目录存在
	// 如果目录不存在则创建，创建失败时回退到标准输出
	logDir := filepath.Dir(output)
	if err := ensureDir(logDir); err != nil {
		fmt.Printf("创建日志目录 %s 失败: %v，将使用 stdout 代替\n", logDir, err)
		return zapcore.AddSync(os.Stdout)
	}

	// 使用 lumberjack 实现日志轮转
	// lumberjack 提供了基于文件大小、文件数量、文件年龄的自动轮转功能
	// 轮转规则：
	// 1. 当文件大小超过 MaxSize 时触发轮转
	// 2. 保留最多 MaxBackups 个旧文件
	// 3. 删除超过 MaxAge 天的旧文件
	// 4. 根据 Compress 配置决定是否压缩旧文件
	lumberjackLogger := &lumberjack.Logger{
		Filename:   output,                    // 日志文件路径
		MaxSize:    getMaxSize(logConfig),     // 单个日志文件最大尺寸(MB)
		MaxBackups: getMaxBackups(logConfig),  // 保留的旧日志文件最大数量
		MaxAge:     getMaxAge(logConfig),      // 保留的旧日志文件最大天数
		Compress:   getCompress(logConfig),    // 是否压缩旧日志文件
		LocalTime:  true,                      // 使用本地时间命名轮转文件
	}

	return zapcore.AddSync(lumberjackLogger)
}

// getMaxSize 获取日志文件最大尺寸配置
// 
// 日志轮转的触发条件之一，当日志文件达到此大小时会创建新文件
// 
// 参数:
//   - config: 日志配置对象
//
// 返回:
//   - int: 文件最大尺寸(MB)，默认100MB
func getMaxSize(config *LoggerConfig) int {
	if config != nil && config.MaxSize > 0 {
		return config.MaxSize
	}
	return 100 // 默认100MB
}

// getMaxBackups 获取保留的旧日志文件最大数量配置
//
// 控制轮转后保留的历史日志文件数量，超过此数量的最旧文件会被删除
// 设置为0表示不限制数量
//
// 参数:
//   - config: 日志配置对象
//
// 返回:
//   - int: 保留文件数量，默认10个
func getMaxBackups(config *LoggerConfig) int {
	if config != nil && config.MaxBackups > 0 {
		return config.MaxBackups
	}
	return 10 // 默认保留10个文件
}

// getMaxAge 获取保留的旧日志文件最大天数配置
//
// 控制轮转后保留的历史日志文件的时间，超过此天数的文件会被删除
// 设置为0表示不根据时间删除文件
//
// 参数:
//   - config: 日志配置对象
//
// 返回:
//   - int: 保留天数，默认30天
func getMaxAge(config *LoggerConfig) int {
	if config != nil && config.MaxAge > 0 {
		return config.MaxAge
	}
	return 30 // 默认保留30天
}

// getCompress 获取是否压缩旧日志文件配置
//
// 控制轮转后的历史日志文件是否使用gzip压缩以节省磁盘空间
// 压缩文件会以.gz扩展名结尾
//
// 参数:
//   - config: 日志配置对象
//
// 返回:
//   - bool: 是否压缩，默认true
func getCompress(config *LoggerConfig) bool {
	if config != nil {
		return config.Compress
	}
	return true // 默认开启压缩
}

// ensureDir 确保目录存在
//
// 内部工具函数，用于确保指定目录存在
// 如果目录不存在则创建，使用0755权限
//
// 参数:
//   - dirPath: 目录路径
//
// 返回:
//   - error: 创建过程中的错误
func ensureDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// ===== 基础接口 =====

// Info 记录信息级别日志
// 
// 信息级别用于记录应用程序的一般信息，如业务流程、状态变更等
// 这是生产环境中最常用的日志级别，用于追踪应用程序的正常运行状态
//
// 支持多种参数格式：
// 1. Info("消息")                                    - 简单消息
// 2. Info("消息", err)                               - 消息 + 错误对象
// 3. Info("消息", map[string]any{...})               - 消息 + 结构化数据
// 4. Info("消息", "key1", val1, "key2", val2, ...)   - 消息 + 键值对列表
// 5. Info("消息", zap.String(...), zap.Int(...))     - 消息 + zap字段列表
//
// 参数:
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func Info(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Info(msg, parseArgs(args...)...)
}

// InfoWithTrace 记录带跟踪ID的信息级别日志
//
// 在分布式系统中，跟踪ID用于关联同一请求在不同服务间的日志记录
// 除了跟踪ID，还会自动提取用户ID、用户名、租户ID等上下文信息
//
// 参数:
//   - ctx: 上下文对象，包含跟踪ID和用户信息
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func InfoWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Info(msg, fields...)
}

// Debug 记录调试级别日志
//
// 调试级别用于记录详细的程序执行信息，通常只在开发环境启用
// 包含变量值、方法调用、中间结果等细节信息，有助于问题诊断
//
// 参数格式与Info相同，支持多种参数类型组合
//
// 参数:
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func Debug(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Debug(msg, parseArgs(args...)...)
}

// DebugWithTrace 记录带跟踪ID的调试级别日志
//
// 结合上下文跟踪的调试信息，用于分布式系统的详细问题诊断
//
// 参数:
//   - ctx: 上下文对象，包含跟踪ID和用户信息
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func DebugWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Debug(msg, fields...)
}

// Warn 记录警告级别日志
//
// 警告级别用于记录可能的问题或异常情况，但不会影响程序正常运行
// 如配置项缺失使用默认值、API调用较慢、资源使用率较高等
//
// 参数格式与Info相同，支持多种参数类型组合
//
// 参数:
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func Warn(msg string, args ...any) {
	if log == nil {
		return
	}
	log.Warn(msg, parseArgs(args...)...)
}

// WarnWithTrace 记录带跟踪ID的警告级别日志
//
// 结合上下文跟踪的警告信息，便于追踪特定请求的潜在问题
//
// 参数:
//   - ctx: 上下文对象，包含跟踪ID和用户信息
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func WarnWithTrace(ctx context.Context, msg string, args ...any) {
	if log == nil {
		return
	}
	fields := parseArgs(args...)
	fields = appendTraceID(ctx, fields)
	log.Warn(msg, fields...)
}

// Error 记录错误级别日志
//
// 错误级别用于记录程序运行中的错误情况，但程序仍能继续运行
// 自动提供增强的错误处理：
// 1. 对于 Error(msg, err) 格式，自动提取完整的错误栈信息
// 2. 对于其他格式，添加调用栈信息以便调试
//
// 参数格式与Info相同，当第一个参数是error类型时会特殊处理
//
// 参数:
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func Error(msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Error(msg, err)模式
	// 当第二个参数是error类型时，提供增强的错误信息
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用huberrors获取完整错误栈信息
			// 这包括错误链和调用栈的详细信息
			errorStack := huberrors.ErrorStack(err)
			log.Error(msg, zap.Error(err), zap.String("error_stack", errorStack))
			return
		}
	}

	// 对于其他格式的参数，添加调用栈信息
	fields := parseArgs(args...)
	stack := captureStack(2) // 跳过当前函数和调用者
	fields = append(fields, zap.String("error_stack", stack))
	log.Error(msg, fields...)
}

// ErrorWithTrace 记录带跟踪ID的错误级别日志
//
// 结合上下文跟踪的错误信息，便于在分布式系统中追踪错误的传播路径
//
// 参数:
//   - ctx: 上下文对象，包含跟踪ID和用户信息
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
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
//
// 致命错误级别用于记录导致程序无法继续运行的严重错误
// 记录日志后程序会立即退出，因此应谨慎使用
// 适用场景：配置错误、关键资源无法获取、系统初始化失败等
//
// 错误信息处理方式与Error函数相同，但会终止程序执行
//
// 参数:
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
func Fatal(msg string, args ...any) {
	if log == nil {
		return
	}

	// 处理常见的Fatal(msg, err)模式
	if len(args) == 1 {
		if err, ok := args[0].(error); ok {
			// 使用增强的堆栈跟踪获取完整错误信息
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
//
// 结合上下文跟踪的致命错误记录，便于在分布式系统中定位问题根源
//
// 参数:
//   - ctx: 上下文对象，包含跟踪ID和用户信息
//   - msg: 日志消息内容
//   - args: 可变参数，支持多种格式的附加信息
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
//
// 这是日志系统的核心参数解析函数，负责将各种格式的参数转换为zap.Field切片
// 支持的参数格式包括：
// 1. 单个map[string]any        - 结构化数据
// 2. 单个[]zap.Field          - 预构建的字段列表
// 3. 单个zap.Field            - 单个字段
// 4. 键值对列表               - key1, val1, key2, val2, ...
// 5. zap.Field列表           - field1, field2, field3, ...
// 6. 混合格式                 - 无法识别时打包为单个字段
//
// 参数:
//   - args: 可变参数列表，支持多种格式
//
// 返回:
//   - []zap.Field: 转换后的zap字段切片
func parseArgs(args ...any) []zap.Field {
	if len(args) == 0 {
		return nil
	}

	// 检查是否为单个map参数
	// map[string]any格式适合传递结构化数据
	if len(args) == 1 {
		if m, ok := args[0].(map[string]any); ok {
			return mapToFields(m)
		}

		// 检查是否为zap.Field切片
		// 这种格式适合预先构建字段列表的场景
		if fields, ok := args[0].([]zap.Field); ok {
			return fields
		}

		// 检查是否为单个zap.Field
		// 单个字段的简单情况
		if field, ok := args[0].(zap.Field); ok {
			return []zap.Field{field}
		}
	}

	// 检查是否为键值对列表 (key1, val1, key2, val2, ...)
	// 这种格式要求参数数量为偶数，且奇数位置必须为字符串
	if len(args) > 1 && len(args)%2 == 0 {
		allStringKeys := true
		for i := 0; i < len(args); i += 2 {
			if _, ok := args[i].(string); !ok {
				allStringKeys = false
				break
			}
		}

		// 如果所有键都是字符串，则按键值对处理
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
	// 所有参数都必须是zap.Field类型
	allFields := true
	for _, arg := range args {
		if _, ok := arg.(zap.Field); !ok {
			allFields = false
			break
		}
	}

	// 如果所有参数都是zap.Field，则直接转换
	if allFields {
		fields := make([]zap.Field, len(args))
		for i, arg := range args {
			fields[i] = arg.(zap.Field)
		}
		return fields
	}

	// 无法识别的格式，打包为一个字段
	// 这是兜底处理，确保不会丢失任何信息
	return []zap.Field{zap.Any("args", args)}
}

// mapToFields 将map转换为zap字段列表
//
// 将键值对格式的数据转换为zap可以处理的字段列表
// 每个键值对会转换为一个zap.Field，使用zap.Any进行类型适配
//
// 参数:
//   - data: 键值对数据
//
// 返回:
//   - []zap.Field: 转换后的字段列表
func mapToFields(data map[string]any) []zap.Field {
	fields := make([]zap.Field, 0, len(data))
	for k, v := range data {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}

// captureStack 获取调用堆栈的详细信息
//
// 这个函数用于捕获当前的函数调用栈，主要用于错误日志的诊断
// 会过滤掉标准库和zap库的栈帧，只保留应用代码的调用信息
//
// 调用栈格式：
// /path/to/file.go:123 package.FunctionName
// /path/to/file.go:456 package.AnotherFunction
//
// 参数:
//   - skip: 要跳过的调用帧数，通常用于跳过日志函数本身
//
// 返回:
//   - string: 格式化的堆栈信息
func captureStack(skip int) string {
	var buffer strings.Builder

	// 获取最多32层堆栈信息
	// 这个深度对于大多数应用来说已经足够
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	// 遍历栈帧
	for {
		frame, more := frames.Next()

		// 跳过标准库和zap库的帧
		// 只保留应用代码的调用信息，使输出更加清晰
		if !strings.Contains(frame.File, "runtime/") &&
			!strings.Contains(frame.File, "zap/") &&
			!strings.Contains(frame.File, "zapcore/") {
			// 构建文件位置信息
			file := frame.File
			line := frame.Line
			function := frame.Function

			// 提取短函数名，去掉包路径前缀
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
//
// 这个函数专门用于处理错误对象，能够提取错误的完整调用链
// 使用了huberrors包的增强功能，支持错误包装和堆栈跟踪
//
// 相比captureStack，这个函数能够：
// 1. 处理错误包装链（如使用fmt.Errorf("wrap: %w", err)）
// 2. 提供更详细的错误上下文信息
// 3. 保留原始错误的堆栈跟踪
//
// 参数:
//   - err: 错误对象
//
// 返回:
//   - string: 格式化的错误堆栈信息
func captureErrorStack(err error) string {
	// 使用huberrors包的ErrorStack函数获取完整的错误链和调用栈
	// 它能够处理嵌套错误并保留原始错误信息
	return huberrors.ErrorStack(err)
}

// CreateLogDirectory 创建日志目录
//
// 用于确保日志目录存在，如果目录不存在则创建
// 这是一个公开的工具函数，可以在应用启动时主动创建日志目录
//
// 参数:
//   - dir: 要创建的目录路径
//
// 返回:
//   - error: 创建过程中的错误
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
