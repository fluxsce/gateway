package logwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
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

	// GetLogConfig 获取日志配置
	GetLogConfig() *types.LogConfig
}

var (
	// 全局写入器缓存 - 按实例ID直接缓存LogWriter
	writerCache = make(map[string]LogWriter)
	// 保护写入器缓存的互斥锁
	cacheMutex sync.RWMutex
	// 本机IP缓存 - 程序启动时获取一次，后续直接使用
	localIPCache string
)

// init 包初始化函数，获取本机真实IP
func init() {
	localIPCache = getRealLocalIP()
	if localIPCache == "" {
		localIPCache = "127.0.0.1" // 备用默认值
		logger.Warn("Failed to get real local IP, using localhost as fallback")
	} else {
		logger.Info("Local IP cached", "ip", localIPCache)
	}
}

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

	// 从缓存中删除写入器
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
	accessLog := buildAccessLogFromContext(instanceID, gatewayCtx)

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
func buildAccessLogFromContext(instanceID string, gatewayCtx *core.Context) *types.AccessLog {
	// 通过 LogWriter 获取配置
	writer, err := GetLogWriter(instanceID)
	if err != nil {
		// 如果获取写入器失败，使用默认配置
		config := &types.LogConfig{}
		config.SetDefaults()
		logger.Warn("Failed to get writer for instance, using defaults", "instanceID", instanceID, "error", err)
		return buildAccessLogWithConfig(instanceID, gatewayCtx, config)
	}

	config := writer.GetLogConfig()
	if config == nil {
		// 如果配置为空，使用默认配置
		config = &types.LogConfig{}
		config.SetDefaults()
		logger.Warn("Writer returned nil config, using defaults", "instanceID", instanceID)
	}

	return buildAccessLogWithConfig(instanceID, gatewayCtx, config)
}

// buildAccessLogWithConfig 根据配置构建访问日志
func buildAccessLogWithConfig(instanceID string, gatewayCtx *core.Context, config *types.LogConfig) *types.AccessLog {
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

	// 创建基础访问日志
	accessLog := &types.AccessLog{
		// 基础标识信息
		TenantID:          getTenantID(gatewayCtx), // 从上下文中获取租户ID
		TraceID:           traceID,                 // 使用上下文中的trace_id
		GatewayInstanceID: instanceID,
		GatewayNodeIP:     getLocalIP(),

		// 从上下文获取关联ID (转换为字符串)
		RouteConfigID:       gatewayCtx.GetRouteID(),
		ServiceDefinitionID: gatewayCtx.GetServiceID(),
		LogConfigID:         getLogConfigID(gatewayCtx),

		// 时间信息 - 使用上下文中的实际时间
		GatewayStartProcessingTime: gatewayCtx.GetStartTime(),
		// 不在初始化时设置完成时间，由后续逻辑根据实际情况设置

		// 标准数据库字段
		LogLevel:       "INFO",
		LogType:        types.LogTypeAccess,
		AddTime:        now,
		EditTime:       now,
		AddWho:         types.DefaultAddWho,
		EditWho:        types.DefaultEditWho,
		OprSeqFlag:     generateOprSeqFlag(),
		CurrentVersion: types.DefaultVersion,
		ActiveFlag:     types.DefaultActiveFlag,
		ResetFlag:      types.DefaultResetFlag,
		RetryCount:     0,
		ResetCount:     0,
	}

	// 设置请求信息 - 优先使用原始信息
	accessLog.SetRequestInfo(
		getOriginalOrCurrentMethod(req, gatewayCtx),
		getOriginalOrCurrentPath(req, gatewayCtx),
		getOriginalOrCurrentQuery(req, gatewayCtx),
		getOriginalOrCurrentHeaders(req, gatewayCtx, config),
		getRequestBodyWithConfig(req, gatewayCtx, config),
		getRequestSize(req),
	)

	// 设置客户端信息
	accessLog.SetClientInfo(
		getClientIP(req),
		getClientPortValue(req), // 将指针类型转换为值类型
		req.UserAgent(),
		req.Referer(),
		getUserIdentifier(gatewayCtx), // 从上下文获取用户标识
	)

	// 设置转发信息
	accessLog.SetForwardInfo(
		gatewayCtx.GetMatchedPath(),
		gatewayCtx.GetTargetURL(),
		req.Method, // 通常转发方法与请求方法相同
		getForwardParamsWithConfig(req, gatewayCtx, config),
		getForwardHeadersWithConfig(req, gatewayCtx, config),
		getForwardBodyWithConfig(req, gatewayCtx, config),
		getLoadBalancerDecision(gatewayCtx), // 负载均衡决策信息
	)

	// 从上下文获取冗余字段 - 这些字段用于提升查询性能，避免多表JOIN
	var gatewayInstanceName, routeName, serviceName, proxyType string

	// 获取网关实例名称（由网关启动时设置）
	if name, ok := gatewayCtx.GetString(constants.ContextKeyGatewayInstanceName); ok {
		gatewayInstanceName = name
	}

	// 获取路由名称（由路由处理器设置）
	if name, ok := gatewayCtx.GetString(constants.ContextKeyRouteConfigName); ok {
		routeName = name
	}

	// 获取服务名称（由服务发现处理器设置）
	if name, ok := gatewayCtx.GetString(constants.ContextKeyServiceDefinitionName); ok {
		serviceName = name
	}

	// 获取代理类型（由代理处理器设置，如http/websocket/tcp/udp）
	if pType, ok := gatewayCtx.GetString(constants.ContextKeyProxyType); ok {
		proxyType = pType
	}

	// 设置冗余字段到访问日志
	accessLog.SetRedundantFields(gatewayInstanceName, routeName, serviceName, proxyType)

	// 从上下文获取状态码
	// 网关状态码：优先从上下文获取，否则使用默认的 200 OK
	gatewayStatusCode := constants.GatewayStatusOK
	if statusCode, exists := gatewayCtx.GetInt(constants.GatewayStatusCode); exists {
		gatewayStatusCode = statusCode
	}

	// 后端状态码：可选，只有在调用后端服务时才会设置，0表示未设置
	var backendStatusCode int
	if statusCode, exists := gatewayCtx.GetInt(constants.BackendStatusCode); exists {
		backendStatusCode = statusCode
	}

	// 设置后端信息（如果有转发时间信息）
	if !gatewayCtx.GetForwardStartTime().IsZero() && !gatewayCtx.GetForwardResponseTime().IsZero() {
		forwardStartTime := gatewayCtx.GetForwardStartTime()
		forwardResponseTime := gatewayCtx.GetForwardResponseTime()
		accessLog.SetBackendInfo(
			backendStatusCode,
			forwardStartTime,
			forwardResponseTime,
		)
	} else if !gatewayCtx.GetForwardStartTime().IsZero() {
		// 如果只有转发开始时间，说明后端请求可能还在处理中或出现异常
		forwardStartTime := gatewayCtx.GetForwardStartTime()
		accessLog.SetBackendInfo(
			backendStatusCode,
			forwardStartTime,
			time.Time{}, // 后端响应时间为零时间
		)
	}

	// 如果 GetResponseTime() 为零值，则完成时间保持为零时间，表示处理中或异常中断

	// 设置响应信息（注意：SetResponseInfo 内部会重新设置完成时间，这里需要保护我们的设置）
	accessLog.SetResponseInfo(
		gatewayStatusCode,
		getResponseSize(gatewayCtx),                      // 从上下文获取响应大小
		getResponseHeadersWithConfig(gatewayCtx, config), // 从上下文获取响应头
		getResponseBodyWithConfig(gatewayCtx, config),    // 从上下文获取响应体
	)

	// 重新设置正确的完成时间并计算时间指标
	if !gatewayCtx.GetResponseTime().IsZero() {
		responseTime := gatewayCtx.GetResponseTime()
		accessLog.GatewayFinishedProcessingTime = responseTime
	} else {
		// 确保未完成的请求完成时间为零时间
		accessLog.GatewayFinishedProcessingTime = time.Time{}
	}

	// 计算处理时间指标
	accessLog.CalculateProcessingTime()

	// 如果有错误，设置错误信息并可能调整状态码
	if gatewayCtx.HasErrors() {
		errors := gatewayCtx.GetErrors()
		errorMessages := make([]string, len(errors))
		for i, err := range errors {
			errorMessages[i] = err.Error()
		}
		accessLog.SetErrorInfo(
			"GATEWAY_ERROR",
			fmt.Sprintf("Errors: %v", errorMessages),
		)

		// 如果没有设置状态码且有错误，使用内部服务器错误状态码
		if gatewayStatusCode == constants.GatewayStatusOK {
			accessLog.GatewayStatusCode = constants.GatewayStatusInternalError
		}
	}

	return accessLog
}

// getClientIP 获取客户端真实IP
// 尝试从各种头部获取真实IP
func getClientIP(req *http.Request) string {
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
	return localIPCache
}

// generateOprSeqFlag 生成操作序列标识
func generateOprSeqFlag() string {
	return random.Generate32BitRandomString()
}

// stringValue 从字符串指针获取字符串值，如果为nil则返回空字符串
func stringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// getRealLocalIP 获取本机真实IP
func getRealLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// getLogConfigID 从上下文获取日志配置ID
func getLogConfigID(gatewayCtx *core.Context) string {
	if logConfigID, ok := gatewayCtx.GetString(constants.ContextKeyLogConfigID); ok {
		return logConfigID
	}
	return ""
}

// getTenantID 从上下文获取租户ID
func getTenantID(gatewayCtx *core.Context) string {
	if tenantID, ok := gatewayCtx.GetString(constants.ContextKeyTenantID); ok {
		return tenantID
	}
	// 如果没有设置租户ID，返回默认值
	return "default"
}

// serializeRequestHeaders 序列化请求头为JSON字符串
func serializeRequestHeaders(req *http.Request) string {
	if req.Header == nil || len(req.Header) == 0 {
		return ""
	}

	// 简化的头部序列化，只保留关键信息
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			// 对于多值头部，只保留第一个值
			headers[key] = values[0]
		}
	}

	if len(headers) == 0 {
		return ""
	}

	if data, err := json.Marshal(headers); err == nil {
		return string(data)
	}

	return ""
}

// getRequestSize 计算请求大小
func getRequestSize(req *http.Request) int {
	size := 0

	// 计算请求行大小 (方法 + 路径 + 协议版本)
	size += len(req.Method) + len(req.URL.String()) + len(req.Proto) + 4 // 空格和CRLF

	// 计算请求头大小
	for key, values := range req.Header {
		for _, value := range values {
			size += len(key) + len(value) + 4 // ": " 和 CRLF
		}
	}
	size += 2 // 头部结束的 CRLF

	// 计算请求体大小
	if req.ContentLength > 0 {
		size += int(req.ContentLength)
	}

	return size
}

// getResponseSize 从上下文获取响应大小
func getResponseSize(gatewayCtx *core.Context) int {
	// 尝试从上下文中获取响应大小
	// 注意：需要在网关处理器中使用包装的ResponseWriter来记录写入的字节数
	if size, ok := gatewayCtx.GetInt("response_size"); ok {
		return size
	}

	// 如果没有记录响应大小，返回0（表示未知）
	return 0
}

// getResponseHeaders 从响应写入器获取响应头
func getResponseHeaders(gatewayCtx *core.Context) string {
	if gatewayCtx.Writer == nil {
		return ""
	}

	// 从ResponseWriter获取响应头
	responseHeaders := gatewayCtx.Writer.Header()
	if responseHeaders == nil || len(responseHeaders) == 0 {
		return ""
	}

	// 序列化响应头为JSON
	headers := make(map[string]string)
	for key, values := range responseHeaders {
		if len(values) > 0 {
			// 对于多值头部，只保留第一个值
			headers[key] = values[0]
		}
	}

	if len(headers) == 0 {
		return ""
	}

	if data, err := json.Marshal(headers); err == nil {
		return string(data)
	}

	return ""
}

// getClientPort 从请求中解析客户端端口
func getClientPort(req *http.Request) *int {
	// 从 RemoteAddr 解析端口
	if req.RemoteAddr != "" {
		if host, port, err := net.SplitHostPort(req.RemoteAddr); err == nil && host != "" {
			// 解析端口号
			if portNum, err := fmt.Sscanf(port, "%d", new(int)); err == nil && portNum == 1 {
				var p int
				if _, err := fmt.Sscanf(port, "%d", &p); err == nil {
					return &p
				}
			}
		}
	}
	return nil
}

// getClientPortValue 从请求中解析客户端端口并返回值类型，0表示未设置
func getClientPortValue(req *http.Request) int {
	if port := getClientPort(req); port != nil {
		return *port
	}
	return 0 // 未设置时返回0
}

// getUserIdentifier 从上下文获取用户标识
func getUserIdentifier(gatewayCtx *core.Context) string {
	// 尝试从不同的上下文键获取用户标识
	if userID, ok := gatewayCtx.GetString("user_id"); ok {
		return userID
	}
	if userID, ok := gatewayCtx.GetString("user_identifier"); ok {
		return userID
	}
	if userID, ok := gatewayCtx.GetString("authenticated_user"); ok {
		return userID
	}
	return ""
}

// getRequestHeaders 根据配置获取请求头
func getRequestHeaders(req *http.Request, config *types.LogConfig) string {
	// 如果配置不记录请求头，返回空字符串
	if !config.IsRecordHeaders() {
		return ""
	}

	// 复用原有的序列化逻辑
	return serializeRequestHeaders(req)
}

// getRequestBodyWithConfig 根据配置获取请求体
func getRequestBodyWithConfig(req *http.Request, gatewayCtx *core.Context, config *types.LogConfig) string {
	// 如果配置不记录请求体，返回空字符串
	if !config.IsRecordRequestBody() {
		return ""
	}

	// 尝试从上下文获取缓存的请求体
	if bodyData, exists := gatewayCtx.Get("request_body"); exists {
		// 处理字节数据
		if bodyBytes, ok := bodyData.([]byte); ok {
			return stringValue(truncateAndReturnString(bodyBytes, config.MaxBodySizeBytes))
		}
		// 兼容字符串类型
		if bodyStr, ok := bodyData.(string); ok {
			return stringValue(truncateAndReturnString([]byte(bodyStr), config.MaxBodySizeBytes))
		}
	}

	// 如果上下文中没有缓存，返回空字符串
	// 注意：读取请求体可能会影响后续处理器，应该在中间件中缓存
	return ""
}

// getResponseHeadersWithConfig 根据配置获取响应头
func getResponseHeadersWithConfig(gatewayCtx *core.Context, config *types.LogConfig) string {
	// 如果配置不记录响应头，返回空字符串
	if !config.IsRecordHeaders() {
		return ""
	}

	// 复用原有的获取响应头逻辑
	return getResponseHeaders(gatewayCtx)
}

// getResponseBodyWithConfig 根据配置获取响应体
func getResponseBodyWithConfig(gatewayCtx *core.Context, config *types.LogConfig) string {
	// 如果配置不记录响应体，返回空字符串
	if !config.IsRecordResponseBody() {
		return ""
	}

	// 尝试从上下文中获取响应体（字节数据）
	if bodyData, exists := gatewayCtx.Get("response_body"); exists {
		// 处理字节数据
		if bodyBytes, ok := bodyData.([]byte); ok {
			return stringValue(truncateAndReturnString(bodyBytes, config.MaxBodySizeBytes))
		}
		// 兼容字符串类型
		if bodyStr, ok := bodyData.(string); ok {
			return stringValue(truncateAndReturnString([]byte(bodyStr), config.MaxBodySizeBytes))
		}
	}

	return ""
}

// truncateAndReturnString 根据最大长度截断字节数组并返回字符串指针
// 使用UTF-8安全的截断方式，避免截断多字节字符
func truncateAndReturnString(data []byte, maxSize int) *string {
	if len(data) == 0 {
		return nil
	}

	// 如果配置的最大大小为0，表示不限制大小
	if maxSize <= 0 || len(data) <= maxSize {
		result := string(data)
		return &result
	}

	// 需要截断，使用UTF-8安全的方式
	truncatedData := truncateUTF8Safe(data, maxSize-len("...[truncated]"))
	truncated := string(truncatedData) + "...[truncated]"
	return &truncated
}

// truncateUTF8Safe UTF-8安全的字节截断
// 确保不会在多字节字符中间截断
func truncateUTF8Safe(data []byte, maxBytes int) []byte {
	if len(data) <= maxBytes {
		return data
	}

	// 从maxBytes位置向前查找，找到一个完整的UTF-8字符边界
	for i := maxBytes; i > 0; i-- {
		// 检查第i个字节是否是UTF-8字符的开始
		if isUTF8Start(data[i]) {
			return data[:i]
		}
	}

	// 如果找不到合适的截断点，返回空字节数组
	return []byte{}
}

// isUTF8Start 检查字节是否是UTF-8字符的开始字节
func isUTF8Start(b byte) bool {
	// UTF-8字符的开始字节模式：
	// 0xxxxxxx (ASCII, 0-127)
	// 110xxxxx (2字节字符的开始)
	// 1110xxxx (3字节字符的开始)
	// 11110xxx (4字节字符的开始)
	// 不是 10xxxxxx (continuation字节)
	return (b&0x80) == 0 || (b&0xC0) == 0xC0
}

// getForwardParamsWithConfig 根据配置获取转发参数
func getForwardParamsWithConfig(req *http.Request, gatewayCtx *core.Context, config *types.LogConfig) string {
	return ""
}

// getForwardHeadersWithConfig 根据配置获取转发头部
func getForwardHeadersWithConfig(req *http.Request, gatewayCtx *core.Context, config *types.LogConfig) string {
	// 如果配置不记录头部，返回空字符串
	if !config.IsRecordHeaders() {
		return ""
	}

	// 从上下文获取转发头部（http.Header类型）
	if forwardHeaders, exists := gatewayCtx.Get(constants.ContextKeyForwardHeaders); exists {
		if headers, ok := forwardHeaders.(http.Header); ok {
			// 将http.Header序列化为JSON字符串
			headersMap := make(map[string]string)
			for key, values := range headers {
				if len(values) > 0 {
					// 对于多值头部，只保留第一个值
					headersMap[key] = values[0]
				}
			}

			if len(headersMap) > 0 {
				if data, err := json.Marshal(headersMap); err == nil {
					return string(data)
				}
			}
		}
	}

	return ""
}

// getForwardBodyWithConfig 根据配置获取转发请求体
func getForwardBodyWithConfig(req *http.Request, gatewayCtx *core.Context, config *types.LogConfig) string {
	return ""
}

// getLoadBalancerDecision 获取负载均衡决策信息
func getLoadBalancerDecision(gatewayCtx *core.Context) string {
	// 从上下文获取负载均衡决策信息
	return ""
}

// getOriginalOrCurrentMethod 获取原始请求方法或当前请求方法
func getOriginalOrCurrentMethod(req *http.Request, gatewayCtx *core.Context) string {
	if originalMethod, ok := gatewayCtx.GetString(constants.ContextKeyOriginalMethod); ok {
		return originalMethod
	}
	return req.Method
}

// getOriginalOrCurrentPath 获取原始请求路径或当前请求路径
func getOriginalOrCurrentPath(req *http.Request, gatewayCtx *core.Context) string {
	if originalPath, ok := gatewayCtx.GetString(constants.ContextKeyOriginalURLPath); ok {
		return originalPath
	}
	return req.URL.Path
}

// getOriginalOrCurrentQuery 获取原始请求查询参数或当前请求查询参数
func getOriginalOrCurrentQuery(req *http.Request, gatewayCtx *core.Context) string {
	if originalQuery, ok := gatewayCtx.GetString(constants.ContextKeyOriginalQueryString); ok {
		return originalQuery
	}
	return req.URL.RawQuery
}

// getOriginalOrCurrentHeaders 获取原始请求头或当前请求头
func getOriginalOrCurrentHeaders(req *http.Request, gatewayCtx *core.Context, config *types.LogConfig) string {
	// 如果配置不记录请求头，返回空字符串
	if !config.IsRecordHeaders() {
		return ""
	}

	// 尝试获取原始请求头
	if originalHeaders, exists := gatewayCtx.Get(constants.ContextKeyOriginalHeaders); exists {
		if headers, ok := originalHeaders.(map[string][]string); ok {
			// 将原始头部序列化为JSON字符串
			headersMap := make(map[string]string)
			for key, values := range headers {
				if len(values) > 0 {
					// 对于多值头部，只保留第一个值
					headersMap[key] = values[0]
				}
			}

			if len(headersMap) > 0 {
				if data, err := json.Marshal(headersMap); err == nil {
					return string(data)
				}
			}
		}
	}

	// 如果没有原始头部，使用当前请求头
	return serializeRequestHeaders(req)
}
