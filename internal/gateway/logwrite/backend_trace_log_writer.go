package logwrite

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/logwrite/types"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// WriteBackendTraceLogSync 静态方法：同步构建后端追踪日志对象并写入（在 defer 中调用）
// 该方法在 defer 中同步构建日志对象，避免上下文取消带来的异常
// writer 内部支持异步缓存，直接调用即可
//
// 参数：
//   - instanceID: 网关实例ID（如果为空则从上下文获取）
//   - gatewayCtx: 网关上下文
//   - serviceID: 服务ID（必填，用于多服务转发场景）
//   - traceID: 主请求的追踪ID（可选，如果为空则从上下文获取）
//   - requestMethod: HTTP请求方法（GET, POST, PUT, DELETE等）
//   - requestURL: 实际转发的请求URL（完整URL）
//   - requestSize: 请求大小（字节）
//   - requestStartTime: 请求开始时间
//   - responseTime: 响应时间
//   - statusCode: HTTP状态码
//   - responseHeaders: 响应头（必填，不能从上下文获取，因为多服务转发时每个服务的响应头不同）
//   - responseBody: 响应体（可选）
//   - forwardHeaders: 转发请求头（必填，不能从上下文获取，因为多服务转发时每个服务的转发请求头可能不同）
//   - forwardBody: 转发请求体（可选，如果与主请求不同则记录）
//   - err: 错误信息（如果有）
//   - serviceName: 服务名称（可选，用于日志记录）
//
// 返回：
//   - error: 构建失败时返回错误信息（注意：写入是异步的，错误可能不会立即返回）
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
) error {
	// 使用 defer + recover 确保整个方法（包括构建和写入）的异常不会导致服务崩溃
	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic in backend trace log sync",
				"error", r,
				"traceId", traceID,
				"serviceId", serviceID,
				"instanceID", instanceID)
		}
	}()

	// 如果没有提供 traceID，从上下文获取
	if traceID == "" {
		if id, exists := gatewayCtx.Get(constants.ContextKeyTraceID); exists {
			if idStr, ok := id.(string); ok {
				traceID = idStr
			}
		}
		if traceID == "" {
			logger.Error("traceID is required for backend trace log",
				"serviceId", serviceID,
				"instanceID", instanceID)
			return nil // 不返回错误，避免影响服务
		}
	}

	// 如果没有提供 instanceID，尝试从上下文获取
	if instanceID == "" {
		if id, exists := gatewayCtx.GetString(constants.ContextKeyGatewayInstanceID); exists {
			instanceID = id
		}
		if instanceID == "" {
			logger.Error("instanceID is required for backend trace log",
				"traceId", traceID,
				"serviceId", serviceID)
			return nil // 不返回错误，避免影响服务
		}
	}

	// 服务ID必须提供
	if serviceID == "" {
		logger.Error("serviceID is required for backend trace log",
			"traceId", traceID,
			"instanceID", instanceID)
		return nil // 不返回错误，避免影响服务
	}

	// 获取租户ID（从上下文中获取，使用log_writer.go中提供的getTenantID方法）
	tenantID := getTenantID(gatewayCtx)
	if tenantID == "" {
		logger.Error("tenantID is required for backend trace log",
			"traceId", traceID,
			"serviceId", serviceID,
			"instanceID", instanceID)
		return nil // 不返回错误，避免影响服务
	}

	// 响应信息已作为参数传入，不需要从上下文获取
	// 这样可以避免多服务转发时响应头混淆的问题

	// 获取日志写入器
	writer, buildErr := GetLogWriter(instanceID)
	if buildErr != nil {
		logger.Error("Failed to get log writer",
			"error", buildErr,
			"traceId", traceID,
			"instanceID", instanceID)
		return nil // 不返回错误，避免影响服务
	}

	// 获取日志配置
	config := writer.GetLogConfig()
	if config == nil {
		config = &types.LogConfig{}
		config.SetDefaults()
	}

	// 同步构建日志对象（避免上下文取消带来的异常）
	now := time.Now()
	requestDuration := responseTime.Sub(requestStartTime)

	// 生成后端追踪ID
	backendTraceID := random.Generate32BitRandomString()

	// 创建后端追踪日志实例（传入租户ID、追踪ID、后端追踪ID）
	backendLog := types.NewBackendTraceLog(tenantID, traceID, backendTraceID)

	// 设置标准数据库字段
	backendLog.AddTime = now
	backendLog.EditTime = now
	backendLog.AddWho = types.DefaultAddWho
	backendLog.EditWho = types.DefaultEditWho
	backendLog.OprSeqFlag = generateOprSeqFlag()
	backendLog.CurrentVersion = types.DefaultVersion
	backendLog.ActiveFlag = types.DefaultActiveFlag

	// 设置服务信息（不包含节点信息，因为负载均衡选择是动态的）
	backendLog.SetServiceInfo(serviceID, serviceName)

	// 设置转发信息
	// 从 requestURL 解析出路径和查询参数（这是实际转发的URL）
	forwardAddress := requestURL
	forwardPath := ""
	forwardQuery := ""
	if parsedURL, parseErr := url.Parse(requestURL); parseErr == nil {
		forwardPath = parsedURL.Path
		forwardQuery = parsedURL.RawQuery
	} else {
		// 如果解析失败，使用默认值
		forwardPath = gatewayCtx.GetMatchedPath()
		forwardQuery = getOriginalOrCurrentQuery(gatewayCtx)
	}

	// 将转发请求头转换为JSON字符串
	forwardHeadersStr := ""
	if len(forwardHeaders) > 0 {
		if headerBytes, marshalErr := json.Marshal(forwardHeaders); marshalErr == nil {
			forwardHeadersStr = string(headerBytes)
		}
	} else {
		// 如果没有提供转发请求头，尝试从上下文获取（兼容性处理，单服务转发场景）
		if headersData, exists := gatewayCtx.Get(constants.ContextKeyForwardHeaders); exists {
			if headers, ok := headersData.(map[string][]string); ok {
				if headerBytes, marshalErr := json.Marshal(headers); marshalErr == nil {
					forwardHeadersStr = string(headerBytes)
				}
			}
		}
		// 如果上下文中也没有，尝试从配置获取（兼容性处理）
		if forwardHeadersStr == "" {
			forwardHeadersStr = getForwardHeadersWithConfig(gatewayCtx, config)
		}
	}

	// 将转发请求体转换为字符串（根据日志配置决定是否记录）
	forwardBodyStr := ""
	if config.IsRecordRequestBody() {
		if len(forwardBody) > 0 {
			// 根据最大长度截断请求体
			forwardBodyStr = stringValue(truncateAndReturnString(forwardBody, config.MaxBodySizeBytes))
		} else {
			// 如果没有提供转发请求体，尝试从上下文获取（兼容性处理，单服务转发场景）
			forwardBodyStr = getForwardBodyWithConfig(gatewayCtx, config)
		}
	}

	backendLog.SetForwardInfo(forwardAddress, requestMethod, forwardPath, forwardQuery, forwardHeadersStr, forwardBodyStr, requestSize)

	// 设置时间信息
	responseReceivedTime := requestStartTime.Add(requestDuration)
	backendLog.SetTimeInfo(requestStartTime, responseReceivedTime)

	// 设置响应信息（根据日志配置决定是否记录响应体）
	responseBodyStr := ""
	if config.IsRecordResponseBody() {
		if len(responseBody) > 0 {
			// 根据最大长度截断响应体
			responseBodyStr = stringValue(truncateAndReturnString(responseBody, config.MaxBodySizeBytes))
		}
	}
	responseHeadersStr := ""
	if len(responseHeaders) > 0 {
		if headerBytes, marshalErr := json.Marshal(responseHeaders); marshalErr == nil {
			responseHeadersStr = string(headerBytes)
		}
	}
	responseSize := len(responseBodyStr)
	backendLog.SetResponseInfo(statusCode, responseSize, responseHeadersStr, responseBodyStr)

	// 设置错误信息和成功状态
	if err != nil {
		errorCode := "BACKEND_ERROR"
		if statusCode >= 400 && statusCode < 500 {
			errorCode = "HTTP_CLIENT_ERROR"
		} else if statusCode >= 500 {
			errorCode = "HTTP_SERVER_ERROR"
		}
		backendLog.SetErrorInfo(errorCode, err.Error())
	} else if statusCode >= 200 && statusCode < 300 {
		backendLog.SetSuccess()
	} else if statusCode >= 400 {
		errorCode := "HTTP_ERROR"
		if statusCode >= 500 {
			errorCode = "HTTP_SERVER_ERROR"
		} else {
			errorCode = "HTTP_CLIENT_ERROR"
		}
		backendLog.SetErrorInfo(errorCode, fmt.Sprintf("HTTP状态码: %d", statusCode))
	}

	// 直接调用 writer 写入，writer 内部支持异步缓存
	if writeErr := writer.WriteBackendTraceLog(context.Background(), backendLog); writeErr != nil {
		logger.Error("Failed to write backend trace log",
			"error", writeErr,
			"traceId", traceID,
			"serviceId", serviceID,
			"instanceID", instanceID)
		// 日志写入失败不影响服务，只记录错误，不返回错误
		return nil
	}

	return nil
}
