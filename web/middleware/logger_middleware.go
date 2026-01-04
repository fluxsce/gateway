package middleware

import (
	"context"
	"fmt"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// TraceIDKey 跟踪ID在上下文中的键名
	TraceIDKey = "trace_id"
	// TraceIDHeader 跟踪ID的HTTP头名称
	TraceIDHeader = "X-Trace-ID"
	// RequestIDHeader 请求ID的HTTP头名称
	RequestIDHeader = "X-Request-ID"
)

// LoggerMiddleware 统一的日志中间件
// 功能：
// 1. 生成或获取跟踪ID
// 2. 设置跟踪ID到上下文
// 3. 记录请求开始和结束日志
// 4. 在响应头中返回跟踪ID
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// 1. 生成或获取跟踪ID
		traceID := getOrGenerateTraceID(c)

		// 2. 设置跟踪ID到上下文
		setTraceIDToContext(c, traceID)

		// 3. 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.GetHeader("User-Agent")

		// 4. 记录请求开始日志
		logger.InfoWithTrace(c.Request.Context(), "请求开始",
			"method", method,
			"path", path,
			"client_ip", clientIP,
			"user_agent", userAgent)

		// 5. 处理请求
		c.Next()

		// 6. 计算处理时间和获取响应状态
		duration := time.Since(start)
		status := c.Writer.Status()
		responseSize := c.Writer.Size()

		// 7. 记录请求结束日志
		logLevel := getLogLevel(status)
		logMessage := fmt.Sprintf("请求完成 - %s %s", method, path)

		switch logLevel {
		case "error":
			logger.ErrorWithTrace(c.Request.Context(), logMessage,
				"method", method,
				"path", path,
				"status", status,
				"duration", duration,
				"response_size", responseSize,
				"client_ip", clientIP)
		case "warn":
			logger.WarnWithTrace(c.Request.Context(), logMessage,
				"method", method,
				"path", path,
				"status", status,
				"duration", duration,
				"response_size", responseSize,
				"client_ip", clientIP)
		default:
			logger.InfoWithTrace(c.Request.Context(), logMessage,
				"method", method,
				"path", path,
				"status", status,
				"duration", duration,
				"response_size", responseSize,
				"client_ip", clientIP)
		}
	}
}

// getOrGenerateTraceID 获取或生成跟踪ID
func getOrGenerateTraceID(c *gin.Context) string {
	// 尝试从请求头获取跟踪ID
	traceID := c.GetHeader(TraceIDHeader)
	if traceID == "" {
		// 尝试从X-Request-ID获取
		traceID = c.GetHeader(RequestIDHeader)
	}

	// 如果没有跟踪ID，生成新的
	if traceID == "" {
		traceID = random.GenerateUniqueStringWithPrefix("TRACE-", 32)
	}

	return traceID
}

// setTraceIDToContext 设置跟踪ID到上下文
func setTraceIDToContext(c *gin.Context, traceID string) {
	// 设置到Gin上下文
	c.Set(TraceIDKey, traceID)

	// 设置到Go标准上下文
	ctx := context.WithValue(c.Request.Context(), TraceIDKey, traceID)
	c.Request = c.Request.WithContext(ctx)

	// 在响应头中返回跟踪ID
	c.Header(TraceIDHeader, traceID)
	c.Header(RequestIDHeader, traceID)
}

// getLogLevel 根据HTTP状态码确定日志级别
func getLogLevel(status int) string {
	if status >= 500 {
		return "error"
	} else if status >= 400 {
		return "warn"
	}
	return "info"
}

// GetTraceIDFromGin 从Gin上下文中获取跟踪ID
func GetTraceIDFromGin(c *gin.Context) string {
	if traceID, exists := c.Get(TraceIDKey); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceIDFromContext 从上下文中获取跟踪ID
func GetTraceIDFromContext(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

// WithTraceID 为上下文添加跟踪ID
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, TraceIDKey, traceID)
}
