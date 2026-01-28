package interceptor

import (
	"context"
	"time"

	"gateway/pkg/logger"

	"google.golang.org/grpc"
)

// LoggingInterceptor 日志拦截器
// 记录请求开始、结束时间、处理时长、错误信息
type LoggingInterceptor struct{}

// NewLoggingInterceptor 创建日志拦截器
func NewLoggingInterceptor() *LoggingInterceptor {
	return &LoggingInterceptor{}
}

// UnaryServerInterceptor 返回 Unary 日志拦截器
// 记录请求开始、结束时间、处理时长、错误信息
func (l *LoggingInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		startTime := time.Now()

		// 获取客户端 IP
		clientIP, _ := getClientIP(ctx)

		// 获取认证信息（如果存在）
		authenticated := ctx.Value("authenticated")
		authToken := ctx.Value("auth_token")

		// 记录请求开始
		logger.Debug("RPC 请求开始",
			"method", info.FullMethod,
			"clientIP", clientIP,
			"authenticated", authenticated)

		// 执行实际的 RPC 处理
		resp, err := handler(ctx, req)

		// 计算处理时长
		duration := time.Since(startTime)

		// 记录日志
		if err != nil {
			logger.Warn("RPC 请求失败",
				"method", info.FullMethod,
				"clientIP", clientIP,
				"authenticated", authenticated,
				"authToken", authToken,
				"duration", duration,
				"error", err)
		} else {
			logger.Debug("RPC 请求成功",
				"method", info.FullMethod,
				"clientIP", clientIP,
				"authenticated", authenticated,
				"duration", duration)
		}

		return resp, err
	}
}

// StreamServerInterceptor 返回 Stream 日志拦截器
func (l *LoggingInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		startTime := time.Now()

		clientIP, _ := getClientIP(ss.Context())
		authenticated := ss.Context().Value("authenticated")

		logger.Debug("Stream RPC 请求开始",
			"method", info.FullMethod,
			"clientIP", clientIP,
			"authenticated", authenticated,
			"isClientStream", info.IsClientStream,
			"isServerStream", info.IsServerStream)

		err := handler(srv, ss)

		duration := time.Since(startTime)

		if err != nil {
			logger.Warn("Stream RPC 请求失败",
				"method", info.FullMethod,
				"clientIP", clientIP,
				"authenticated", authenticated,
				"duration", duration,
				"error", err)
		} else {
			logger.Debug("Stream RPC 请求成功",
				"method", info.FullMethod,
				"clientIP", clientIP,
				"authenticated", authenticated,
				"duration", duration)
		}

		return err
	}
}

// ================================================================================
// TODO: 扩展日志功能
// ================================================================================

// MetricsCollector 指标收集器（待实现）
type MetricsCollector struct {
	requestCount   map[string]int64 // 方法 -> 请求次数
	requestLatency map[string]int64 // 方法 -> 平均延迟（毫秒）
	errorCount     map[string]int64 // 方法 -> 错误次数
}

// RecordRequest 记录请求指标（待实现）
func (m *MetricsCollector) RecordRequest(method string, duration time.Duration, err error) {
	// TODO: 实现指标收集
	// 1. 记录请求次数
	// 2. 记录请求延迟
	// 3. 记录错误次数
	// 4. 支持导出到 Prometheus、Grafana 等
	logger.Debug("指标收集待实现", "method", method, "duration", duration)
}

// AccessLogger 访问日志记录器（待实现）
type AccessLogger struct {
	logFile string
}

// LogAccess 记录访问日志（待实现）
func (a *AccessLogger) LogAccess(method, clientIP, userAgent string, duration time.Duration, statusCode int) {
	// TODO: 实现访问日志
	// 1. 记录到专门的访问日志文件
	// 2. 支持日志轮转
	// 3. 支持自定义日志格式
	logger.Debug("访问日志待实现",
		"method", method,
		"clientIP", clientIP,
		"duration", duration)
}

// AuditLogger 审计日志记录器（待实现）
type AuditLogger struct {
	logFile string
}

// LogAudit 记录审计日志（待实现）
func (a *AuditLogger) LogAudit(method, user, action, resource string, success bool) {
	// TODO: 实现审计日志
	// 1. 记录敏感操作（注册、注销、配置变更等）
	// 2. 记录操作用户、时间、结果
	// 3. 支持导出到审计系统
	logger.Debug("审计日志待实现",
		"method", method,
		"user", user,
		"action", action,
		"resource", resource,
		"success", success)
}
