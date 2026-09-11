package logwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// WriteBackendTraceLogSync 在代理 defer 中调用：本 goroutine 只抽字段并入队，不组对象、不写库。
// 组 BackendTraceLog、头 JSON、body 截断、acquire 写入器在提交 worker 中执行，避免挡住 ServeHTTP 返回。
// 入队前仍在请求栈上，可从 gatewayCtx.data 补 ID；worker 不再访问 ctx.Request / Writer。
//
// 记的是过滤之后真正转发出去的请求（调用方从 proxyReq 传入），不是进网关原文，也不是事后解引用 ctx.Request。
// 主表访问日志的方法/路径/头仍用过滤器运行前的 Original*，不要和本函数字段混用。
//
// 参数：
//   - instanceID: 网关实例ID（如果为空则从上下文获取）
//   - gatewayCtx: 网关上下文，仅用于补 traceID / instanceID / tenantID / 响应大小 / 转发头兜底
//   - serviceID: 服务ID（必填，用于多服务转发场景）
//   - traceID: 主请求的追踪ID（可选，如果为空则从上下文获取）
//   - requestMethod: 实际转发给上游的 HTTP 方法（GET, POST, PUT, DELETE等）
//   - requestURL: 实际转发的请求URL（完整URL，过滤/重写之后）
//   - requestSize: 向后端发送的请求大小（字节）
//   - requestStartTime: 向该后端发起请求的开始时间
//   - responseTime: 该后端响应结束时间（不是网关总响应时间）
//   - statusCode: 该后端返回的 HTTP 状态码
//   - responseHeaders: 该后端响应头（必填，不能从上下文获取，因为多服务转发时每个服务的响应头不同）
//   - responseBody: 该后端响应体（可选；是否落库由日志配置 RecordResponseBody 决定，并按 MaxBodySizeBytes 截断）
//   - forwardHeaders: 实际转发给该后端的请求头（必填，不能从上下文获取，因为多服务转发时每个服务的转发请求头可能不同）
//   - forwardBody: 实际转发给该后端的请求体（可选；是否落库由 RecordRequestBody 决定，并截断）
//   - err: 该次后端调用的错误（如果有）
//   - serviceName: 服务名称（可选，用于日志记录）
//   - retryCount: 重试次数（当前请求是第几次重试，0表示首次请求）
//
// 返回：
//   - error: 快照阶段失败只打日志并返回 nil，避免影响转发；入队后的写库错误也不回到调用方。
func WriteBackendTraceLogSync(
	instanceID string,
	gatewayCtx *core.Context,
	serviceID string,
	traceID string,
	requestMethod string,
	requestURL string,
	requestSize int,
	requestStartTime time.Time,
	responseTime time.Time,
	statusCode int,
	responseHeaders map[string][]string,
	responseBody []byte,
	forwardHeaders map[string][]string,
	forwardBody []byte,
	err error,
	serviceName string,
	retryCount int,
) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in backend trace log snapshot",
				"error", r,
				"traceId", traceID,
				"serviceId", serviceID,
				"instanceID", instanceID)
		}
	}()

	job := snapshotBackendTraceJob(
		instanceID,
		gatewayCtx,
		serviceID,
		traceID,
		requestMethod,
		requestURL,
		requestSize,
		requestStartTime,
		responseTime,
		statusCode,
		responseHeaders,
		responseBody,
		forwardHeaders,
		forwardBody,
		err,
		serviceName,
		retryCount,
	)
	if job == nil {
		return nil
	}
	submitBackendTraceJob(job)
	return nil
}

// snapshotBackendTraceJob 在请求栈上抽一份可脱离 HTTP 对象的转发快照。
// 过滤器已改完；这里记下的方法/URL/头/体就是即将或已经发给上游的内容。
func snapshotBackendTraceJob(
	instanceID string,
	gatewayCtx *core.Context,
	serviceID string,
	traceID string,
	requestMethod string,
	requestURL string,
	requestSize int,
	requestStartTime time.Time,
	responseTime time.Time,
	statusCode int,
	responseHeaders map[string][]string,
	responseBody []byte,
	forwardHeaders map[string][]string,
	forwardBody []byte,
	err error,
	serviceName string,
	retryCount int,
) *backendTraceJob {
	if gatewayCtx == nil {
		logger.Error("gateway context is required for backend trace log",
			"serviceId", serviceID,
			"instanceID", instanceID)
		return nil
	}
	if traceID == "" {
		if id, exists := gatewayCtx.Get(constants.ContextKeyTraceID); exists {
			if idStr, ok := id.(string); ok {
				traceID = idStr
			}
		}
	}
	if instanceID == "" {
		if id, exists := gatewayCtx.GetString(constants.ContextKeyGatewayInstanceID); exists {
			instanceID = id
		}
	}
	if serviceID == "" {
		logger.Error("serviceID is required for backend trace log",
			"traceId", traceID,
			"instanceID", instanceID)
		return nil
	}
	if traceID == "" {
		logger.Error("traceID is required for backend trace log",
			"serviceId", serviceID,
			"instanceID", instanceID)
		return nil
	}
	if instanceID == "" {
		logger.Error("instanceID is required for backend trace log",
			"traceId", traceID,
			"serviceId", serviceID)
		return nil
	}

	tenantID := getTenantID(gatewayCtx)
	if tenantID == "" {
		logger.Error("tenantID is required for backend trace log",
			"traceId", traceID,
			"serviceId", serviceID,
			"instanceID", instanceID)
		return nil
	}

	if responseTime.IsZero() {
		responseTime = time.Now()
	}

	if len(forwardHeaders) == 0 {
		forwardHeaders = forwardHeadersFromContext(gatewayCtx)
	}

	// 此后不再读 ctx.Request。头/体已是过滤后转发快照，worker 只消费本结构。
	job := &backendTraceJob{
		instanceID:       instanceID,
		tenantID:         tenantID,
		traceID:          traceID,
		serviceID:        serviceID,
		serviceName:      serviceName,
		requestMethod:    requestMethod,
		requestURL:       requestURL,
		fallbackPath:     gatewayCtx.GetMatchedPath(),
		fallbackQuery:    getOriginalOrCurrentQuery(gatewayCtx),
		requestSize:      requestSize,
		responseSize:     resolveBackendResponseSize(gatewayCtx, responseBody, ""),
		requestStartTime: requestStartTime,
		responseTime:     responseTime,
		statusCode:       statusCode,
		responseHeaders:  responseHeaders,
		responseBody:     responseBody,
		forwardHeaders:   forwardHeaders,
		forwardBody:      forwardBody,
		retryCount:       retryCount,
	}
	if err != nil {
		job.hasErr = true
		job.errText = err.Error()
	}
	return job
}

// writeBackendTraceFromJob 在提交 worker 中组从表对象并写入，字段含义与优化前同步路径一致。
func writeBackendTraceFromJob(job *backendTraceJob) {
	writer, release, buildErr := acquireLogWriter(job.instanceID)
	if buildErr != nil {
		logger.Error("Failed to get log writer",
			"error", buildErr,
			"traceId", job.traceID,
			"instanceID", job.instanceID)
		return
	}
	defer release()

	config := writer.GetLogConfig()
	if config == nil {
		config = &types.LogConfig{}
		config.SetDefaults()
	}

	now := time.Now()
	requestDuration := job.responseTime.Sub(job.requestStartTime)
	backendLog := types.NewBackendTraceLog(job.tenantID, job.traceID, random.Generate32BitRandomString())
	backendLog.AddTime = now
	backendLog.EditTime = now
	backendLog.AddWho = types.DefaultAddWho
	backendLog.EditWho = types.DefaultEditWho
	backendLog.OprSeqFlag = generateOprSeqFlag()
	backendLog.CurrentVersion = types.DefaultVersion
	backendLog.ActiveFlag = types.DefaultActiveFlag
	// 不包含节点信息，负载均衡选择是动态的。
	backendLog.SetServiceInfo(job.serviceID, job.serviceName)

	forwardAddress := job.requestURL
	forwardPath := ""
	forwardQuery := ""
	if parsedURL, parseErr := url.Parse(job.requestURL); parseErr == nil {
		forwardPath = parsedURL.Path
		forwardQuery = parsedURL.RawQuery
	} else {
		forwardPath = job.fallbackPath
		forwardQuery = job.fallbackQuery
	}

	forwardHeadersStr := marshalHeaderMap(job.forwardHeaders)
	forwardBodyStr := ""
	if config.IsRecordRequestBody() && len(job.forwardBody) > 0 {
		forwardBodyStr = stringValue(truncateAndReturnString(job.forwardBody, config.MaxBodySizeBytes))
	}
	backendLog.SetForwardInfo(forwardAddress, job.requestMethod, forwardPath, forwardQuery, forwardHeadersStr, forwardBodyStr, job.requestSize)

	responseReceivedTime := job.requestStartTime.Add(requestDuration)
	backendLog.SetTimeInfo(job.requestStartTime, responseReceivedTime)

	responseBodyStr := ""
	if config.IsRecordResponseBody() && len(job.responseBody) > 0 {
		responseBodyStr = stringValue(truncateAndReturnString(job.responseBody, config.MaxBodySizeBytes))
	}
	responseHeadersStr := marshalHeaderMap(job.responseHeaders)
	responseSize := job.responseSize
	if responseSize <= 0 {
		if len(job.responseBody) > 0 {
			responseSize = len(job.responseBody)
		} else {
			responseSize = len(responseBodyStr)
		}
	}
	backendLog.SetResponseInfo(job.statusCode, responseSize, responseHeadersStr, responseBodyStr)
	backendLog.SetRetryInfo(job.retryCount)

	if job.hasErr {
		errorCode := "BACKEND_ERROR"
		if job.statusCode >= 400 && job.statusCode < 500 {
			errorCode = "HTTP_CLIENT_ERROR"
		} else if job.statusCode >= 500 {
			errorCode = "HTTP_SERVER_ERROR"
		}
		backendLog.SetErrorInfo(errorCode, job.errText)
	} else if (job.statusCode >= 200 && job.statusCode < 300) || job.statusCode == 101 {
		// 101 Switching Protocols 视为 WebSocket 握手成功。
		backendLog.SetSuccess()
	} else if job.statusCode >= 400 {
		errorCode := "HTTP_ERROR"
		if job.statusCode >= 500 {
			errorCode = "HTTP_SERVER_ERROR"
		} else {
			errorCode = "HTTP_CLIENT_ERROR"
		}
		backendLog.SetErrorInfo(errorCode, fmt.Sprintf("HTTP状态码: %d", job.statusCode))
	}

	if writeErr := writer.WriteBackendTraceLog(context.Background(), backendLog); writeErr != nil {
		logger.Error("Failed to write backend trace log",
			"error", writeErr,
			"traceId", job.traceID,
			"serviceId", job.serviceID,
			"instanceID", job.instanceID)
	}
}

func marshalHeaderMap(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}
	headerBytes, err := json.Marshal(headers)
	if err != nil {
		return ""
	}
	return string(headerBytes)
}

func forwardHeadersFromContext(gatewayCtx *core.Context) map[string][]string {
	if gatewayCtx == nil {
		return nil
	}
	headersData, exists := gatewayCtx.Get(constants.ContextKeyForwardHeaders)
	if !exists {
		return nil
	}
	switch headers := headersData.(type) {
	case map[string][]string:
		return headers
	case http.Header:
		return headers
	default:
		return nil
	}
}

// resolveBackendResponseSize 解析后端追踪响应大小：优先流式累计字节，其次原始响应体，最后采样字符串长度。
func resolveBackendResponseSize(gatewayCtx *core.Context, responseBody []byte, responseBodyStr string) int {
	if gatewayCtx != nil {
		if size, ok := gatewayCtx.GetInt(constants.ContextKeyResponseSize); ok && size >= 0 {
			return size
		}
		if bytesVal, exists := gatewayCtx.Get(constants.ContextKeySSEBytesStreamed); exists {
			if n, ok := asInt64(bytesVal); ok {
				return clampInt64ToLogSize(n)
			}
		}
		if bytesVal, exists := gatewayCtx.Get(constants.ContextKeyWebSocketBytesSent); exists {
			if n, ok := asInt64(bytesVal); ok {
				return clampInt64ToLogSize(n)
			}
		}
	}
	if len(responseBody) > 0 {
		return len(responseBody)
	}
	return len(responseBodyStr)
}
