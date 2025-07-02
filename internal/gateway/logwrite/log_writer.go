package logwrite

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gohub/internal/gateway/constants"
	"gohub/internal/gateway/core"
	"gohub/internal/gateway/logwrite/types"
	"gohub/pkg/logger"
	"gohub/pkg/utils/random"
)

// LogWriter 定义日志写入器接口
type LogWriter interface {
	// Write 写入单条日志
	Write(ctx context.Context, log *types.AccessLog) error
	
	// BatchWrite 批量写入日志
	BatchWrite(ctx context.Context, logs []*types.AccessLog) error
	
	// Flush 刷新缓冲区
	Flush(ctx context.Context) error
	
	// Close 关闭写入器
	Close() error
}

var (
	// 全局写入器缓存 - 按实例ID直接缓存LogWriter
	writerCache = make(map[string]LogWriter)
	// 保护写入器缓存的互斥锁
	cacheMutex sync.RWMutex
)

// RegisterLogWriter 注册日志写入器到指定实例
func RegisterLogWriter(instanceID string, writer LogWriter) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	// 如果已存在写入器，先关闭它
	if existingWriter, exists := writerCache[instanceID]; exists {
		if err := existingWriter.Close(); err != nil {
			logger.Error("Failed to close existing writer", "instanceID", instanceID, "error", err)
		}
	}
	
	// 注册新写入器
	writerCache[instanceID] = writer
	logger.Info("Writer registered", "instanceID", instanceID)
	return nil
}

// UnregisterLogWriter 注销指定实例的写入器
func UnregisterLogWriter(instanceID string) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	writer, exists := writerCache[instanceID]
	if !exists {
		return fmt.Errorf("writer not found for instance: %s", instanceID)
	}
	
	// 关闭写入器
	if err := writer.Close(); err != nil {
		logger.Error("Failed to close writer", "instanceID", instanceID, "error", err)
	}
	
	// 从缓存中删除
	delete(writerCache, instanceID)
	
	logger.Info("Writer unregistered", "instanceID", instanceID)
	return nil
}

// InitLogManager 初始化指定实例的日志写入器
func InitLogManager(instanceID string, config *types.LogConfig) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	if config == nil {
		return fmt.Errorf("log config cannot be nil")
	}
	
	// 使用静态工厂方法创建写入器（每个实例只有一种输出类型）
	writer, err := CreateWriter(config)
	if err != nil {
		return fmt.Errorf("failed to create writer: %v", err)
	}
	
	// 注册写入器
	if err := RegisterLogWriter(instanceID, writer); err != nil {
		return err
	}
	
	logger.Info("Log manager initialized", 
		"instanceID", instanceID,
		"targets", config.GetOutputTargets())
	
	return nil
}

// GetLogWriter 获取指定实例的写入器
func GetLogWriter(instanceID string) (LogWriter, error) {
	if instanceID == "" {
		return nil, fmt.Errorf("instanceID cannot be empty")
	}
	
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	
	writer, exists := writerCache[instanceID]
	if !exists {
		return nil, fmt.Errorf("writer not found for instance: %s", instanceID)
	}
	
	return writer, nil
}

// HasLogWriter 检查指定实例是否存在写入器
func HasLogWriter(instanceID string) bool {
	if instanceID == "" {
		return false
	}
	
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	
	_, exists := writerCache[instanceID]
	return exists
}

// GetInstanceIDs 获取所有实例ID列表
func GetInstanceIDs() []string {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	
	ids := make([]string, 0, len(writerCache))
	for id := range writerCache {
		ids = append(ids, id)
	}
	
	return ids
}

// WriteLog 写入单条日志到指定实例
func WriteLog(instanceID string, gatewayCtx *core.Context) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	writer, err := GetLogWriter(instanceID)
	if err != nil {
		return err
	}
	
	// 从网关上下文构建访问日志
	accessLog := buildAccessLogFromContext(gatewayCtx)
	
	return writer.Write(gatewayCtx.Ctx, accessLog)
}

// FlushLogWriter 刷新指定实例写入器的缓冲区
func FlushLogWriter(instanceID string) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	writer, err := GetLogWriter(instanceID)
	if err != nil {
		return err
	}
	
	return writer.Flush(context.Background())
}

// CloseAllLogWriters 关闭所有写入器
func CloseAllLogWriters() error {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	
	var lastErr error
	for instanceID, writer := range writerCache {
		if err := writer.Close(); err != nil {
			logger.Error("Failed to close writer", "instanceID", instanceID, "error", err)
			lastErr = err
		}
	}
	
	// 清空缓存
	writerCache = make(map[string]LogWriter)
	logger.Info("All log writers closed")
	return lastErr
}

// CloseLogWriter 关闭指定实例的写入器
func CloseLogWriter(instanceID string) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	return UnregisterLogWriter(instanceID)
}

// UpdateLogWriter 动态更新指定实例的日志写入器配置
// 这个方法会关闭旧的写入器，创建新的写入器，并更新缓存
// 
// 安全特性：
// - 原子性更新：要么完全成功，要么完全失败
// - 优雅降级：更新失败时保持旧配置
// - 资源清理：确保旧写入器正确关闭
// - 并发安全：使用互斥锁保护更新过程
// - 最小锁持有时间：在锁外创建新写入器，锁内仅进行原子替换
func UpdateLogWriter(instanceID string, newConfig *types.LogConfig) error {
	if instanceID == "" {
		return fmt.Errorf("instanceID cannot be empty")
	}
	
	if newConfig == nil {
		return fmt.Errorf("log config cannot be nil")
	}

	// 在锁外创建新的写入器，减少锁持有时间
	newWriter, err := CreateWriter(newConfig)
	if err != nil {
		return fmt.Errorf("failed to create new writer: %w", err)
	}

	// 获取锁，进行原子性替换
	cacheMutex.Lock()
	
	// 获取旧的写入器实例
	oldWriter, exists := writerCache[instanceID]
	
	// 原子性地替换写入器
	writerCache[instanceID] = newWriter
	
	// 立即释放锁，避免影响其他并发操作
	cacheMutex.Unlock()

	// 如果存在旧写入器，异步进行优雅关闭
	if exists {
		go func(writer LogWriter, id string) {
			// 给旧写入器一些时间完成正在进行的操作
			time.Sleep(100 * time.Millisecond)
			
			// 刷新旧写入器的缓冲区
			if err := writer.Flush(context.Background()); err != nil {
				logger.Warn("Failed to flush old writer before update", 
					"instanceID", id, "error", err)
			}
			
			// 关闭旧写入器
			if err := writer.Close(); err != nil {
				logger.Warn("Failed to close old writer during update", 
					"instanceID", id, "error", err)
			}
			
			logger.Debug("Old writer closed successfully", "instanceID", id)
		}(oldWriter, instanceID)
	}
	
	logger.Info("Log writer updated successfully", 
		"instanceID", instanceID,
		"targets", newConfig.GetOutputTargets(),
		"hadOldWriter", exists)
	
	return nil
}

// buildAccessLogFromContext 从网关上下文构建访问日志
func buildAccessLogFromContext(gatewayCtx *core.Context) *types.AccessLog {
	// 提取请求信息
	req := gatewayCtx.Request
	now := time.Now()
	
	// 从上下文中获取trace_id，如果没有则生成新的
	var traceID string
	if id, exists := gatewayCtx.Get(constants.ContextKeyTraceID); exists {
		if idStr, ok := id.(string); ok {
			traceID = idStr
		}
	}
	if traceID == "" {
		traceID = random.Generate32BitRandomString()
	}
	
	// 创建访问日志
	accessLog := &types.AccessLog{
		// 基础标识信息
		TenantID:                 "default", // 可以从上下文中获取
		TraceID:                  traceID,   // 使用上下文中的trace_id
		GatewayInstanceID:        "gateway-001", // 可以从配置中获取
		GatewayNodeIP:            getLocalIP(),
		
		// 请求基本信息
		RequestMethod:            req.Method,
		RequestPath:              req.URL.Path,
		RequestQuery:             stringPtr(req.URL.RawQuery),
		ClientIPAddress:          getClientIP(req),
		UserAgent:                stringPtr(req.UserAgent()),
		
		// 网关处理信息（通过扩展属性存储）
		MatchedRoute:             stringPtr(gatewayCtx.GetMatchedPath()),
		ForwardAddress:           stringPtr(gatewayCtx.GetTargetURL()),
		
		// 时间信息
		GatewayReceivedTime:      now,
		GatewayStartProcessingTime: now,
		GatewayFinishedProcessingTime: now,
		GatewayResponseSentTime:  now,
		TotalProcessingTimeMs:    int(gatewayCtx.Elapsed().Milliseconds()),
		GatewayProcessingTimeMs:  int(gatewayCtx.Elapsed().Milliseconds()),
		
		// 响应信息
		GatewayStatusCode:        200, // 默认值，实际应该从响应中获取
		ResponseSize:             0,   // 默认值，实际应该从响应中获取
		
		// 标准数据库字段
		LogLevel:                 "INFO",
		LogType:                  types.LogTypeAccess,
		AddTime:                  now,
		EditTime:                 now,
		AddWho:                   types.DefaultAddWho,
		EditWho:                  types.DefaultEditWho,
		OprSeqFlag:               generateOprSeqFlag(),
		CurrentVersion:           types.DefaultVersion,
		ActiveFlag:               types.DefaultActiveFlag,
		ResetFlag:                types.DefaultResetFlag,
		RetryCount:               0,
		ResetCount:               0,
	}
	
	// 如果有错误，添加错误信息
	if gatewayCtx.HasErrors() {
		errors := gatewayCtx.GetErrors()
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		accessLog.ErrorMessage = stringPtr(fmt.Sprintf("Errors: %v", errorMessages))
		accessLog.LogLevel = "ERROR"
	}
	
	// 通过扩展属性存储额外信息
	extProperty := map[string]interface{}{
		"routeID":   gatewayCtx.GetRouteID(),
		"serviceID": gatewayCtx.GetServiceID(),
	}
	if extData, err := jsonMarshal(extProperty); err == nil {
		accessLog.ExtProperty = stringPtr(extData)
	}
	
	return accessLog
}

// getClientIP 获取客户端真实IP
func getClientIP(req *http.Request) string {
	// 尝试从各种头部获取真实IP
	clientIP := req.Header.Get("X-Forwarded-For")
	if clientIP == "" {
		clientIP = req.Header.Get("X-Real-IP")
	}
	if clientIP == "" {
		clientIP = req.Header.Get("X-Client-IP")
	}
	if clientIP == "" {
		clientIP = req.RemoteAddr
	}
	
	// 如果是X-Forwarded-For，取第一个IP
	if clientIP != "" {
		if idx := strings.Index(clientIP, ","); idx > 0 {
			clientIP = strings.TrimSpace(clientIP[:idx])
		}
	}
	
	return clientIP
}

// getLocalIP 获取本地IP（简化实现）
func getLocalIP() string {
	return "127.0.0.1" // 实际应该获取真实的本地IP
}

// generateTraceID 生成链路追踪ID（已废弃，使用上下文中的trace_id）
func generateTraceID() string {
	return random.Generate32BitRandomString()
}

// generateOprSeqFlag 生成操作序列标识
func generateOprSeqFlag() string {
	return random.Generate32BitRandomString()
}

// stringPtr 返回字符串指针
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// jsonMarshal 简化的JSON序列化
func jsonMarshal(v interface{}) (string, error) {
	return fmt.Sprintf("%+v", v), nil // 简化实现
} 