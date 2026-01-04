package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/service"
	"gateway/internal/gateway/logwrite"
)

// ServiceResponse 服务响应信息
type ServiceResponse struct {
	ServiceID  string              // 服务ID
	NodeID     string              // 节点ID
	URL        string              // 请求的URL
	StatusCode int                 // HTTP状态码
	Headers    map[string][]string // 响应头
	Body       []byte              // 响应体
	Error      error               // 错误信息
	Duration   time.Duration       // 请求耗时
	StartTime  time.Time           // 请求开始时间（用于日志记录）
	Success    bool                // 是否成功
}

// HTTPMultiServiceProxy HTTP多服务代理处理器
// 负责处理一个请求转发到多个后端服务的场景
// 使用 HTTPProxy 的节点选择和路径构建方法，复用负载均衡逻辑
type HTTPMultiServiceProxy struct {
	httpProxy *HTTPProxy // HTTP代理实例，用于复用节点选择和路径构建方法
	client    *http.Client
	config    *HTTPProxyConfig
}

// NewHTTPMultiServiceProxy 创建HTTP多服务代理实例
func NewHTTPMultiServiceProxy(httpProxy *HTTPProxy) *HTTPMultiServiceProxy {
	return &HTTPMultiServiceProxy{
		httpProxy: httpProxy,
		client:    httpProxy.client,
		config:    httpProxy.config,
	}
}

// Handle 处理多服务并行转发
func (m *HTTPMultiServiceProxy) Handle(ctx *core.Context, serviceIDs []string, config *router.MultiServiceConfig) bool {
	if len(serviceIDs) == 0 {
		ctx.AddError(fmt.Errorf("服务ID列表不能为空"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "服务ID列表不能为空",
		})
		return false
	}

	// 使用默认配置
	if config == nil {
		config = &router.MultiServiceConfig{
			ResponseMergeStrategy: "first",
			MaxConcurrentRequests: 0,
			RequireAllSuccess:     false,
		}
	}

	// 预先读取请求体（因为多个goroutine需要共享）
	var requestBody []byte
	if ctx.Request.Body != nil {
		var err error
		requestBody, err = io.ReadAll(ctx.Request.Body)
		if err != nil {
			ctx.AddError(fmt.Errorf("读取请求体失败: %w", err))
			ctx.Abort(http.StatusBadRequest, map[string]string{
				"error": "读取请求体失败",
			})
			return false
		}
		// 重置请求体，以便后续可能的使用
		ctx.Request.Body = io.NopCloser(bytes.NewReader(requestBody))
	}

	// 控制并发数
	maxConcurrent := config.MaxConcurrentRequests
	if maxConcurrent <= 0 {
		maxConcurrent = len(serviceIDs) // 不限制
	}

	// 并行转发请求到多个服务
	responses := m.proxyToMultipleServices(ctx, serviceIDs, requestBody, maxConcurrent, config)

	// 设置多服务配置到上下文（供日志记录使用）
	ctx.Set(constants.ContextKeyMultiServiceConfig, config)

	// 根据策略合并响应
	return m.mergeServiceResponses(ctx, responses, config)
}

// proxyToMultipleServices 并行转发到多个服务
// 直接按照 serviceID 循环处理，复用 http_proxy.go 的逻辑
func (m *HTTPMultiServiceProxy) proxyToMultipleServices(
	ctx *core.Context,
	serviceIDs []string,
	requestBody []byte,
	maxConcurrent int,
	config *router.MultiServiceConfig,
) []*ServiceResponse {
	semaphore := make(chan struct{}, maxConcurrent)
	var wg sync.WaitGroup
	responses := make([]*ServiceResponse, len(serviceIDs))
	var mu sync.Mutex

	for i, serviceID := range serviceIDs {
		wg.Add(1)
		go func(index int, sid string) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// 选择目标节点（复用 http_proxy.go 的逻辑）
			serviceConfig, node, err := m.httpProxy.selectTargetNode(ctx, sid)
			if err != nil {
				ctx.AddError(fmt.Errorf("选择服务 %s 的目标节点失败: %w", sid, err))
				// 如果要求所有成功，直接返回错误
				if config.RequireAllSuccess {
					mu.Lock()
					responses[index] = &ServiceResponse{
						ServiceID: sid,
						Error:     err,
						Success:   false,
					}
					mu.Unlock()
					return
				}
				// 否则继续处理其他服务
				mu.Lock()
				responses[index] = &ServiceResponse{
					ServiceID: sid,
					Error:     err,
					Success:   false,
				}
				mu.Unlock()
				return
			}

			// 执行代理请求（复用 http_proxy.go 的逻辑，但不写入响应）
			response := m.proxyRequestToService(ctx, serviceConfig, node, requestBody)

			// 更新后端最大耗时（多服务场景下，记录所有服务中的最大耗时）
			mu.Lock()
			responses[index] = response
			ctx.SetMaxBackendDuration(response.Duration)
			mu.Unlock()

			// 日志写入已在 proxyRequestToService 的 defer 中处理，与 ProxyRequest 保持一致
		}(i, serviceID)
	}

	wg.Wait()
	return responses
}

// proxyRequestToService 向指定服务发送代理请求
// 复用 http_proxy.go 的 ProxyRequest 逻辑，但不写入响应，只返回响应信息
// 日志写入直接调用日志写入类，与 ProxyRequest 保持一致
func (m *HTTPMultiServiceProxy) proxyRequestToService(
	ctx *core.Context,
	serviceConfig *service.ServiceConfig,
	node *service.NodeConfig,
	requestBody []byte,
) *ServiceResponse {
	serviceID := ""
	nodeID := ""
	targetURL := ""
	if serviceConfig != nil {
		serviceID = serviceConfig.ID
	}
	if node != nil {
		nodeID = node.ID
		targetURL = node.URL
	}
	// 解析目标URL（复用 http_proxy.go 的逻辑）
	target, err := url.Parse(targetURL)
	if err != nil {
		return &ServiceResponse{
			ServiceID: serviceID,
			NodeID:    nodeID,
			Error:     fmt.Errorf("解析目标URL失败: %w", err),
			Success:   false,
		}
	}

	// 使用 HTTPProxy 的路径构建方法（复用）
	finalPath := m.httpProxy.buildProxyPath(ctx, target.Path)

	// 构建代理请求URL（复用 http_proxy.go 的逻辑）
	proxyURL := &url.URL{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Path:     finalPath,
		RawQuery: ctx.Request.URL.RawQuery,
	}

	// 创建代理请求（使用预先读取的请求体）
	var body io.Reader
	if len(requestBody) > 0 {
		body = bytes.NewReader(requestBody)
	}

	proxyReq, err := http.NewRequestWithContext(
		context.Background(),
		ctx.Request.Method,
		proxyURL.String(),
		body,
	)
	if err != nil {
		return &ServiceResponse{
			ServiceID: serviceID,
			NodeID:    nodeID,
			URL:       proxyURL.String(),
			Error:     fmt.Errorf("创建代理请求失败: %w", err),
			Success:   false,
		}
	}

	// 复制请求头（复用 http_proxy.go 的逻辑）
	for name, values := range ctx.Request.Header {
		// 跳过hop-by-hop头部 (RFC 7230 Section 6.1)
		if isHopByHopHeader(name) {
			continue
		}
		// 检查是否为Connection头部中列出的hop-by-hop头部
		if isConnectionHeader(name, ctx.Request.Header) {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 使用 HTTPProxy 的头部设置方法（复用）
	m.httpProxy.setProxyHeaders(ctx.Request, proxyReq, target.Host)

	// 记录请求开始时间（用于日志记录）
	requestStartTime := time.Now()

	// 记录请求相关信息（用于日志构建）
	requestMethod := proxyReq.Method
	requestURL := proxyURL.String()
	requestSize := len(requestBody)

	// 用于在 defer 中捕获响应信息的变量
	var responseStatusCode int
	var responseHeaders map[string][]string
	var responseBody []byte
	var responseErr error
	var backendResponseTime time.Time // 后端请求结束时间（用于后端追踪日志）

	// 获取服务名称（用于日志记录）
	serviceName := ""
	if serviceConfig != nil {
		serviceName = serviceConfig.Name
	}

	// 使用defer确保无论成功失败都能写入后端追踪日志（与 ProxyRequest 保持一致）
	defer func() {
		// 在响应处理完成后复制header，避免影响核心时间统计
		headersCopy := make(http.Header)
		for k, v := range proxyReq.Header {
			headersCopy[k] = append([]string(nil), v...)
		}

		// 后端请求结束时间（用于后端追踪日志，不等于网关响应时间）
		// 如果 backendResponseTime 为零，说明请求失败，使用当前时间
		if backendResponseTime.IsZero() {
			backendResponseTime = time.Now()
		}

		// 同步构建后端追踪日志对象并异步写入（避免上下文取消带来的异常）
		// 使用日志写入类的静态方法处理，响应信息和转发信息从局部变量获取（不从上下文获取，避免多服务转发混淆）
		// 将转发请求头转换为 map[string][]string 格式
		forwardHeadersMap := make(map[string][]string)
		for k, v := range headersCopy {
			forwardHeadersMap[k] = append([]string(nil), v...)
		}

		// 直接调用日志写入类（与 ProxyRequest 保持一致）
		_ = logwrite.WriteBackendTraceLogSync(
			"", // instanceID 从上下文获取
			ctx,
			serviceID, // 服务ID手动传入
			"",        // traceID 从上下文获取
			requestMethod,
			requestURL,
			requestSize,
			requestStartTime,
			backendResponseTime, // 后端请求结束时间（用于后端追踪日志）
			responseStatusCode,
			responseHeaders,
			responseBody,
			forwardHeadersMap, // 转发请求头作为参数传入，避免并发覆盖
			requestBody,       // 转发请求体作为参数传入，避免并发覆盖
			responseErr,
			serviceName, // 服务名称，从 node 中获取
		)
	}()

	// 发送代理请求
	resp, err := m.client.Do(proxyReq)
	if err != nil {
		// 请求失败时记录错误和后端请求结束时间
		responseErr = err
		responseStatusCode = 0
		backendResponseTime = time.Now()
		// 记录后端耗时（请求失败时，耗时从请求开始到失败的时间）
		ctx.SetMaxBackendDuration(time.Since(requestStartTime))
		return &ServiceResponse{
			ServiceID: serviceID,
			NodeID:    nodeID,
			URL:       requestURL,
			Error:     fmt.Errorf("发送代理请求失败: %w", err),
			Duration:  time.Since(requestStartTime),
			StartTime: requestStartTime,
			Success:   false,
		}
	}
	defer resp.Body.Close()

	// 保存响应状态码和响应头（用于日志记录）
	responseStatusCode = resp.StatusCode
	responseHeaders = make(map[string][]string)
	for name, values := range resp.Header {
		responseHeaders[name] = append([]string(nil), values...)
	}

	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		responseErr = fmt.Errorf("读取响应体失败: %w", err)
		backendResponseTime = time.Now()
		ctx.SetMaxBackendDuration(time.Since(requestStartTime))
		return &ServiceResponse{
			ServiceID:  serviceID,
			NodeID:     nodeID,
			URL:        requestURL,
			StatusCode: responseStatusCode,
			Headers:    responseHeaders,
			Error:      responseErr,
			Duration:   time.Since(requestStartTime),
			StartTime:  requestStartTime,
			Success:    false,
		}
	}
	responseBody = bodyBytes

	// 记录后端请求结束时间（响应体读取完成后，用于后端追踪日志）
	backendResponseTime = time.Now()
	// 记录后端耗时（从请求开始到响应体读取完成）
	ctx.SetMaxBackendDuration(time.Since(requestStartTime))

	return &ServiceResponse{
		ServiceID:  serviceID,
		NodeID:     nodeID,
		URL:        requestURL,
		StatusCode: responseStatusCode,
		Headers:    responseHeaders,
		Body:       responseBody,
		Duration:   time.Since(requestStartTime),
		StartTime:  requestStartTime,
		Success:    true,
	}
}

// mergeServiceResponses 合并多个服务响应
func (m *HTTPMultiServiceProxy) mergeServiceResponses(
	ctx *core.Context,
	responses []*ServiceResponse,
	config *router.MultiServiceConfig,
) bool {
	strategy := config.ResponseMergeStrategy
	if strategy == "" {
		strategy = "first"
	}

	// 分离成功和失败的响应
	successful := make([]*ServiceResponse, 0)
	failed := make([]*ServiceResponse, 0)
	for _, resp := range responses {
		if resp != nil {
			if resp.Success {
				successful = append(successful, resp)
			} else {
				failed = append(failed, resp)
			}
		}
	}

	// 如果要求所有成功，检查是否有失败
	if config.RequireAllSuccess && len(failed) > 0 {
		// 直接返回一个失败的响应（第一个失败的），包括其状态码、header 和 body
		return m.writeSingleResponse(ctx, failed[0])
	}

	// 根据策略处理响应
	switch strategy {
	case "first":
		// 使用第一个成功的响应
		if len(successful) > 0 {
			return m.writeSingleResponse(ctx, successful[0])
		}
		// 如果没有成功的，返回错误
		if len(failed) > 0 {
			ctx.Abort(http.StatusBadGateway, map[string]string{
				"error":   "所有服务转发失败",
				"details": failed[0].Error.Error(),
			})
			ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
			return false
		}
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "没有可用的响应",
		})
		return false

	case "first_error":
		// 使用第一个失败的响应
		if len(failed) > 0 {
			return m.writeSingleResponse(ctx, failed[0])
		}
		// 如果没有失败的，但有成功的，退回到第一个成功的响应
		if len(successful) > 0 {
			return m.writeSingleResponse(ctx, successful[0])
		}
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "没有可用的响应",
		})
		return false

	case "all":
		// 返回所有响应
		return m.writeAllResponses(ctx, responses)

	default:
		// 默认使用第一个成功的响应
		if len(successful) > 0 {
			return m.writeSingleResponse(ctx, successful[0])
		}
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "没有成功的响应",
		})
		return false
	}
}

// writeSingleResponse 写入单个响应
func (m *HTTPMultiServiceProxy) writeSingleResponse(ctx *core.Context, response *ServiceResponse) bool {
	// 如果后端没有实际请求成功（例如选择节点失败、请求发送失败等），没有可用的后端响应
	// 这种情况下，直接返回一个网关错误响应，而不是写入空的 header/body
	if response == nil || !response.Success || response.StatusCode <= 0 {
		errMsg := "后端服务请求失败"
		if response != nil && response.Error != nil {
			errMsg = response.Error.Error()
		}
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": errMsg,
		})
		ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
		return false
	}

	// 复制响应头
	for name, values := range response.Headers {
		for _, value := range values {
			ctx.Writer.Header().Add(name, value)
		}
	}

	// 设置响应状态码
	ctx.Writer.WriteHeader(response.StatusCode)
	ctx.Set(constants.BackendStatusCode, response.StatusCode)
	ctx.Set(constants.GatewayStatusCode, response.StatusCode)
	ctx.SetResponded()

	// 写入响应体
	_, err := ctx.Writer.Write(response.Body)
	if err != nil {
		ctx.AddError(fmt.Errorf("写入响应体失败: %w", err))
		return false
	}

	return true
}

// writeAllResponses 写入所有响应
func (m *HTTPMultiServiceProxy) writeAllResponses(ctx *core.Context, responses []*ServiceResponse) bool {
	if len(responses) == 0 {
		return false
	}

	// 合并所有响应头（简单合并，不做去重或冲突处理）
	for _, resp := range responses {
		if resp == nil {
			continue
		}
		for name, values := range resp.Headers {
			for _, value := range values {
				ctx.Writer.Header().Add(name, value)
			}
		}
	}

	// 选择一个状态码：优先使用第一个有状态码的响应，否则使用 200
	statusCode := http.StatusOK
	for _, resp := range responses {
		if resp != nil && resp.StatusCode > 0 {
			statusCode = resp.StatusCode
			break
		}
	}

	// 设置响应状态码
	ctx.Writer.WriteHeader(statusCode)
	ctx.Set(constants.BackendStatusCode, statusCode)
	ctx.Set(constants.GatewayStatusCode, statusCode)
	ctx.SetResponded()

	// 直接合并所有 body 顺序写入，不做特殊处理
	var mergedBody bytes.Buffer
	for _, resp := range responses {
		if resp == nil || len(resp.Body) == 0 {
			continue
		}
		if _, err := mergedBody.Write(resp.Body); err != nil {
			ctx.AddError(fmt.Errorf("合并响应体失败: %w", err))
			return false
		}
	}

	_, err := ctx.Writer.Write(mergedBody.Bytes())
	if err != nil {
		ctx.AddError(fmt.Errorf("写入响应体失败: %w", err))
		return false
	}

	return true
}
