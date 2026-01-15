package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"crypto/tls"
	"gateway/internal/gateway/constants"
	"gateway/internal/gateway/core"
	proxyutils "gateway/internal/gateway/handler/proxy/proxy-utils"
	"gateway/internal/gateway/handler/router"
	"gateway/internal/gateway/handler/service"
	"gateway/internal/gateway/logwrite"
	registryManager "gateway/internal/registry/manager"
)

// HTTPProxy HTTP代理实现
type HTTPProxy struct {
	*BaseProxyHandler
	client           *http.Client
	serviceManager   service.ServiceManager
	config           *HTTPProxyConfig
	wsUpgradeHandler *WebSocketUpgradeHandler // WebSocket升级处理器
}

// Handle 处理HTTP代理请求
func (h *HTTPProxy) Handle(ctx *core.Context) bool {
	if !h.IsEnabled() {
		return true
	}

	// 检查是否为WebSocket升级请求（类似nginx处理方式）
	if h.wsUpgradeHandler != nil && h.wsUpgradeHandler.IsWebSocketUpgrade(ctx.Request) {
		return h.handleWebSocketUpgrade(ctx)
	}

	// 获取服务ID数组
	serviceIDs := ctx.GetServiceIDs()
	if len(serviceIDs) == 0 {
		ctx.AddError(fmt.Errorf("服务ID不能为空"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "服务ID不能为空",
		})
		return false
	}

	// 判断是否为多服务转发：服务ID数量大于1，或存在多服务配置
	isMultiService := len(serviceIDs) > 1
	if !isMultiService {
		if _, exists := ctx.Get(constants.ContextKeyMultiServiceConfig); exists {
			isMultiService = true
		}
	}

	// 如果是多服务转发，直接使用多服务代理处理器
	if isMultiService {
		var multiServiceConfig *router.MultiServiceConfig
		if config, exists := ctx.Get(constants.ContextKeyMultiServiceConfig); exists {
			if cfg, ok := config.(*router.MultiServiceConfig); ok {
				multiServiceConfig = cfg
			}
		}
		multiServiceProxy := NewHTTPMultiServiceProxy(h)
		return multiServiceProxy.Handle(ctx, serviceIDs, multiServiceConfig)
	}

	// 单服务场景，使用原有的单服务处理逻辑（带重试）
	serviceID := serviceIDs[0]
	config := h.GetHTTPConfig()
	maxRetries := config.RetryCount
	if maxRetries < 0 {
		maxRetries = 0
	}

	retryTimeout := config.RetryTimeout
	if retryTimeout <= 0 {
		retryTimeout = 30 * time.Second // 默认30秒
	}

	var lastErr error
	var lastNode *service.NodeConfig
	// 累加所有重试的耗时
	var totalBackendDuration time.Duration

	// 执行请求和重试
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// 每次重试都重新选择节点（避免集群中某一台异常时一直重试同一台）
		serviceConfig, node, err := h.selectTargetNode(ctx, serviceID)
		if err != nil {
			// 选择节点失败，如果是重试，继续尝试；否则直接返回错误
			lastErr = fmt.Errorf("选择目标节点失败: %w", err)
			if attempt < maxRetries {
				ctx.AddError(fmt.Errorf("选择节点失败，准备重试 (第%d次): %w", attempt+1, err))
				select {
				case <-ctx.Request.Context().Done():
					return false
				case <-time.After(retryTimeout):
				}
				continue
			}
			ctx.AddError(lastErr)
			ctx.Abort(http.StatusServiceUnavailable, map[string]string{
				"error":   "服务不可用",
				"details": lastErr.Error(),
				"service": serviceID,
			})
			return false
		}

		// 执行代理请求（每次调用都会记录后端追踪日志）
		err, attemptDuration := h.proxyRequest(ctx, serviceConfig, node, attempt)

		// 累加本次请求的耗时
		totalBackendDuration += attemptDuration

		if err == nil {
			// 请求成功，清除重试过程中添加的错误信息
			// 避免成功响应中包含错误信息导致外层响应处理异常
			ctx.ClearErrors()
			// 设置累加后的总耗时
			ctx.SetMaxBackendDuration(totalBackendDuration)
			return true
		}

		// 请求失败，记录错误信息
		lastErr = err
		lastNode = node

		// 检查是否为SSE响应，SSE响应不需要重试
		if _, isSSE := ctx.Get(constants.ContextKeySSEResponse); isSSE {
			// SSE响应已开始流式传输，不进行重试
			return false
		}

		// 如果还有重试次数，继续重试
		if attempt < maxRetries {
			ctx.AddError(fmt.Errorf("请求失败，准备重试 (第%d次，节点: %s): %w", attempt+1, node.URL, err))
			select {
			case <-ctx.Request.Context().Done():
				return false
			case <-time.After(retryTimeout):
			}
			continue
		}
	}

	// 所有重试都失败，设置累加后的总耗时
	ctx.SetMaxBackendDuration(totalBackendDuration)

	// 所有重试都失败
	ctx.AddError(fmt.Errorf("代理请求失败（已重试%d次）: %w", maxRetries, lastErr))
	targetURL := ""
	if lastNode != nil {
		targetURL = lastNode.URL
	}
	ctx.Abort(http.StatusBadGateway, map[string]string{
		"error":      "代理请求失败",
		"details":    lastErr.Error(),
		"target_url": targetURL,
		"service":    serviceID,
	})
	ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
	return false
}

// proxyRequest 代理请求到指定节点（内部方法）
// retryCount: 当前请求是第几次重试（0表示首次请求）
// 返回值:
// - error: 请求错误（如果有）
// - time.Duration: 本次请求的耗时，用于重试累加
func (h *HTTPProxy) proxyRequest(ctx *core.Context, serviceConfig *service.ServiceConfig, node *service.NodeConfig, retryCount int) (error, time.Duration) {
	// 解析目标URL
	target, err := url.Parse(node.URL)
	if err != nil {
		return fmt.Errorf("解析目标URL失败: %w", err), 0
	}

	// 智能处理路径拼接
	finalPath := h.buildProxyPath(ctx, target.Path)

	// 构建代理请求URL
	proxyURL := &url.URL{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Path:     finalPath,
		RawQuery: ctx.Request.URL.RawQuery,
	}

	// 设置目标URL
	ctx.SetTargetURL(proxyURL.String())

	// 创建代理请求
	var body io.Reader
	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %w", err), 0
		}
		body = bytes.NewReader(bodyBytes)
		// 重置原请求的Body
		ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		// 根据日志配置决定是否缓存请求体到上下文中，供日志记录使用
		if h.shouldRecordRequestBody(ctx) {
			ctx.Set("request_body", bodyBytes)
		}
	}

	proxyReq, err := http.NewRequestWithContext(
		context.Background(),
		ctx.Request.Method,
		proxyURL.String(),
		body,
	)
	if err != nil {
		return fmt.Errorf("创建代理请求失败: %w", err), 0
	}

	// 复制请求头
	for name, values := range ctx.Request.Header {
		// 跳过hop-by-hop头部 (RFC 7230 Section 6.1)
		// 这些头部仅对单个连接有效，不应该被转发
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

	// 设置必需的代理头部
	h.setProxyHeaders(ctx.Request, proxyReq, target.Host)

	// 移除 Accept-Encoding 头，让 Go 的 http.Client 自动处理压缩
	// 如果手动设置 Accept-Encoding，Go 的 http.Client 不会自动解压响应
	// 正确的做法是让 Go 标准库自动添加 Accept-Encoding 并自动解压响应
	proxyReq.Header.Del("Accept-Encoding")

	// 设置代理类型
	ctx.Set(constants.ContextKeyProxyType, h.GetType())

	// 记录请求开始时间（用于日志记录）
	requestStartTime := time.Now()

	// 记录请求相关信息（用于日志构建）
	// 从 proxyReq 中获取实际的请求方法和URL
	requestMethod := proxyReq.Method
	requestURL := proxyReq.URL.String()
	var requestSize int
	if bodyData, exists := ctx.Get("request_body"); exists {
		if bodyBytes, ok := bodyData.([]byte); ok {
			requestSize = len(bodyBytes)
		}
	}

	// 获取服务ID和服务名称（用于日志记录）
	serviceID := ""
	serviceName := ""
	if serviceConfig != nil {
		serviceID = serviceConfig.ID
		serviceName = serviceConfig.Name
	}

	// 用于在 defer 中捕获响应信息的变量
	var responseStatusCode int
	var responseHeaders map[string][]string
	var responseBody []byte
	var responseErr error
	var backendResponseTime time.Time // 后端请求结束时间（用于后端追踪日志）

	// 使用defer确保无论成功失败都能保存header参数，并写入后端追踪日志
	defer func() {
		// 在响应处理完成后复制header，避免影响核心时间统计
		headersCopy := make(http.Header)
		for k, v := range proxyReq.Header {
			headersCopy[k] = append([]string(nil), v...)
		}
		ctx.Set(constants.ContextKeyForwardHeaders, headersCopy)

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

		// 获取转发请求体
		var forwardBodyBytes []byte
		if bodyData, exists := ctx.Get("request_body"); exists {
			if bodyBytes, ok := bodyData.([]byte); ok {
				forwardBodyBytes = bodyBytes
			}
		}

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
			forwardBodyBytes,  // 转发请求体作为参数传入，避免并发覆盖
			responseErr,
			serviceName, // 服务名称，从 node 中获取
			retryCount,  // 重试次数
		)
	}()

	// 发送代理请求（异常直接抛出）
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		// 请求失败时记录错误和后端请求结束时间
		responseErr = err
		responseStatusCode = 0
		backendResponseTime = time.Now()
		// 计算本次请求的耗时（请求失败时，耗时从请求开始到失败的时间）
		attemptDuration := time.Since(requestStartTime)
		// 注意：不在这里设置 MaxBackendDuration，由重试循环累加后统一设置
		return err, attemptDuration // 直接返回错误和耗时，不包装
	}
	// defer resp.Body.Close() 位置正确：只有在成功获取响应时才设置 defer
	// 如果 err != nil，resp 为 nil，不会执行到这里，不需要关闭
	defer resp.Body.Close()

	// 保存响应状态码和响应头（用于日志记录）
	responseStatusCode = resp.StatusCode
	responseHeaders = make(map[string][]string)
	for name, values := range resp.Header {
		responseHeaders[name] = append([]string(nil), values...)
	}

	// 设置响应状态码到上下文（供日志记录使用）
	ctx.Set(constants.BackendStatusCode, resp.StatusCode)
	ctx.Set(constants.GatewayStatusCode, resp.StatusCode)

	// 检查是否为SSE响应，如果是则使用特殊处理逻辑
	if h.isSSEResponse(resp) {
		// 设置SSE响应标志位，SSE响应不需要重试
		ctx.Set(constants.ContextKeySSEResponse, true)
		// 对于SSE响应，复制除了已处理头部外的其他头部
		for name, values := range resp.Header {
			lowerName := strings.ToLower(name)
			// 跳过SSE特殊处理方法中已设置的头部
			if lowerName == "content-type" || lowerName == "cache-control" ||
				lowerName == "connection" || lowerName == "access-control-allow-origin" {
				continue
			}
			// 保留Transfer-Encoding用于分块传输
			if lowerName == "transfer-encoding" {
				for _, value := range values {
					ctx.Writer.Header().Add(name, value)
				}
				continue
			}
			for _, value := range values {
				ctx.Writer.Header().Add(name, value)
			}
		}

		// 使用专门的SSE处理方法
		// 注意：SSE响应处理完成后，响应信息会在 defer 中写入日志
		// responseTime 由网关流程结束时设置（gateway.go），不在代理处理中设置
		err = h.handleSSEResponse(ctx, resp)
		if err != nil {
			responseErr = err
		}
		// SSE 流式传输完成后，记录后端请求结束时间
		// 注意：对于SSE，后端请求结束时间是在流式传输完成后
		backendResponseTime = time.Now()
		// 计算本次请求的耗时（从请求开始到流式传输完成）
		attemptDuration := time.Since(requestStartTime)
		// 注意：不在这里设置 MaxBackendDuration，由重试循环累加后统一设置
		return err, attemptDuration
	}

	// 非SSE响应使用常规处理
	// 注意：常规响应处理完成后，响应信息会在 defer 中写入日志
	err = h.handleRegularResponse(ctx, resp)
	if err != nil {
		responseErr = err
	}

	// 如果配置了复制响应体，从上下文获取（已在 handleRegularResponse 中保存）
	if bodyData, exists := ctx.Get("response_body"); exists {
		if bodyBytes, ok := bodyData.([]byte); ok {
			responseBody = bodyBytes
		}
	}

	// 记录后端请求结束时间（响应体读取完成后，用于后端追踪日志）
	backendResponseTime = time.Now()
	// 计算本次请求的耗时（从请求开始到响应体读取完成）
	attemptDuration := time.Since(requestStartTime)
	// 注意：不在这里设置 MaxBackendDuration，由重试循环累加后统一设置

	// responseTime 由网关流程结束时设置（gateway.go），不在代理处理中设置
	return err, attemptDuration
}

// GetHTTPConfig 获取HTTP配置
func (h *HTTPProxy) GetHTTPConfig() HTTPProxyConfig {
	if h.config != nil {
		return *h.config
	}
	return DefaultHTTPProxyConfig
}

// Validate 验证HTTP代理配置
func (h *HTTPProxy) Validate() error {
	config := h.GetHTTPConfig()
	if config.Timeout <= 0 {
		return fmt.Errorf("超时时间必须大于0")
	}
	if config.MaxIdleConns < 0 {
		return fmt.Errorf("最大空闲连接数不能为负数")
	}
	if config.IdleConnTimeout < 0 {
		return fmt.Errorf("空闲连接超时不能为负数")
	}
	if config.BufferSize <= 0 {
		return fmt.Errorf("缓冲区大小必须大于0")
	}
	if config.MaxBufferSize <= 0 {
		return fmt.Errorf("最大缓冲区大小必须大于0")
	}
	if config.BufferSize > config.MaxBufferSize {
		return fmt.Errorf("缓冲区大小不能大于最大缓冲区大小")
	}
	if config.RetryCount < 0 {
		return fmt.Errorf("重试次数不能为负数")
	}
	if config.RetryTimeout < 0 {
		return fmt.Errorf("重试超时不能为负数")
	}

	return nil
}

// Close 关闭HTTP代理
func (h *HTTPProxy) Close() error {
	var lastErr error

	// 优雅关闭WebSocket升级处理器
	if h.wsUpgradeHandler != nil {
		if err := h.wsUpgradeHandler.Shutdown(30 * time.Second); err != nil {
			lastErr = fmt.Errorf("关闭WebSocket升级处理器失败: %w", err)
		}
	}

	// 关闭HTTP客户端连接
	if h.client != nil {
		if transport, ok := h.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}

	// 关闭服务管理器
	// 服务管理器包含健康检查器等需要清理的资源
	if h.serviceManager != nil {
		if closer, ok := h.serviceManager.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				if lastErr == nil {
					lastErr = err
				}
			}
		}
	}

	return lastErr
}

// NewHTTPProxy 创建HTTP代理
func NewHTTPProxy(config ProxyConfig, serviceManager service.ServiceManager) (*HTTPProxy, error) {
	// 解析HTTP特定配置
	httpConfig := DefaultHTTPProxyConfig
	if config.Config != nil {
		parser := NewHTTPConfigParser()
		parser.ParseConfig(config.Config, &httpConfig)
	}

	// 创建WebSocket升级处理器（如果需要支持WebSocket升级）
	wsUpgradeHandler := NewWebSocketUpgradeHandler(serviceManager, nil)

	// 从HTTP配置中继承参数
	wsUpgradeHandler.InheritFromHTTPConfig(&httpConfig)

	// 创建HTTP代理实例
	httpProxy := &HTTPProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		serviceManager:   serviceManager,
		config:           &httpConfig,
		wsUpgradeHandler: wsUpgradeHandler,
	}

	// 使用配置创建HTTP客户端
	httpProxy.client = httpProxy.createHTTPClient(httpConfig)

	return httpProxy, nil
}

// NewHTTPProxyWithRegistry 创建带注册中心支持的HTTP代理（已弃用，使用NewHTTPProxy）
// Deprecated: 使用NewHTTPProxy代替，注册中心会自动获取
func NewHTTPProxyWithRegistry(config ProxyConfig, serviceManager service.ServiceManager, _ interface{}) (*HTTPProxy, error) {
	return NewHTTPProxy(config, serviceManager)
}

// handleWebSocketUpgrade 处理WebSocket协议升级
func (h *HTTPProxy) handleWebSocketUpgrade(ctx *core.Context) bool {
	err := h.wsUpgradeHandler.HandleWebSocketUpgrade(ctx, h.GetName(), h.GetType())
	if err != nil {
		ctx.AddError(fmt.Errorf("代理WebSocket升级请求失败: %w", err))
		// 如果连接已被 hijack（WebSocket 升级成功），不能再使用标准的 HTTP 响应方法
		// 此时应该直接返回，连接会在 WebSocket 升级处理器中关闭
		if ctx.IsResponded() {
			// 连接已被 hijack，不能调用 Abort，直接返回
			ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
			return false
		}
		// 如果升级失败（连接未被 hijack），可以正常返回错误响应
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "代理WebSocket升级请求失败",
		})
		ctx.Set(constants.GatewayStatusCode, constants.GatewayStatusBadGateway)
		return false
	}
	return true
}

// buildProxyPath 构建代理请求路径 - 简化的nginx proxy_pass处理方式
//
// 处理规则：
// 1. 目标路径为空或只有斜杠：使用请求地址
// 2. 前缀不一样：直接使用目标地址
// 3. 前缀一样：处理重复拼接问题
func (h *HTTPProxy) buildProxyPath(ctx *core.Context, targetPath string) string {
	requestPath := ctx.Request.URL.Path

	// 记住原始路径的斜杠状态
	originalTargetHasSlash := strings.HasSuffix(targetPath, "/")
	originalRequestHasSlash := strings.HasSuffix(requestPath, "/")

	// 清理路径
	targetPath = h.cleanPath(targetPath)
	requestPath = h.cleanPath(requestPath)

	// 1. 目标路径为空或只有斜杠：使用请求地址，但要保留原始请求路径的斜杠状态
	if targetPath == "" || targetPath == "/" {
		// 如果原始请求路径以斜杠结尾且清理后不是根路径，需要恢复斜杠
		if originalRequestHasSlash && requestPath != "/" {
			return requestPath + "/"
		}
		return requestPath
	}

	// 2. 前缀不一样：直接使用目标地址
	if !h.hasSamePrefix(targetPath, requestPath) {
		// 如果原始目标路径有斜杠，保留它
		if originalTargetHasSlash && !strings.HasSuffix(targetPath, "/") {
			return targetPath + "/"
		}
		return targetPath
	}

	// 3. 前缀一样：处理重复拼接问题
	// 特殊情况：如果路径完全相同，直接返回目标路径
	if targetPath == requestPath {
		if originalTargetHasSlash && !strings.HasSuffix(targetPath, "/") {
			return targetPath + "/"
		}
		return targetPath
	}

	// 如果请求路径以目标路径为前缀，直接返回请求路径避免重复
	if strings.HasPrefix(requestPath, targetPath) {
		return requestPath
	}

	// 否则根据是否有斜杠决定拼接方式
	if originalTargetHasSlash {
		// 目标路径原本有斜杠，直接拼接
		if requestPath == "/" {
			return targetPath + "/"
		}
		return targetPath + requestPath
	} else {
		// 目标路径不以/结尾，直接拼接
		return targetPath + requestPath
	}
}

// hasSamePrefix 检查目标路径和请求路径是否有相同前缀
func (h *HTTPProxy) hasSamePrefix(targetPath, requestPath string) bool {
	// 获取目标路径的基础部分（去掉结尾斜杠）
	basePath := strings.TrimSuffix(targetPath, "/")

	// 特殊情况：如果目标路径是根路径，只有请求路径也是根路径才算相同前缀
	if basePath == "" {
		return requestPath == "/"
	}

	// 如果请求路径不以目标路径开头，前缀不同
	if !strings.HasPrefix(requestPath, basePath) {
		return false
	}

	// 检查路径边界：确保匹配的是完整的路径段
	// 例如："/ap" 不应该匹配 "/api/v1"
	if len(requestPath) > len(basePath) {
		nextChar := requestPath[len(basePath)]
		return nextChar == '/'
	}

	// 请求路径长度等于或小于目标路径，认为是相同前缀
	return true
}

// cleanPath 清理路径格式
func (h *HTTPProxy) cleanPath(p string) string {
	if p == "" {
		return "/"
	}

	// 确保以 / 开头
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}

	// 使用 path.Clean 清理路径
	return path.Clean(p)
}

// isSSEResponse 检查是否为SSE响应
func (h *HTTPProxy) isSSEResponse(resp *http.Response) bool {
	contentType := resp.Header.Get("Content-Type")
	return strings.HasPrefix(strings.ToLower(contentType), "text/event-stream")
}

// handleSSEResponse 处理SSE响应的特殊逻辑
// 类似nginx的proxy_buffering off和特殊头部处理
func (h *HTTPProxy) handleSSEResponse(ctx *core.Context, resp *http.Response) error {
	// 注意：响应头不再保存到上下文，因为多服务转发时每个服务的响应头不同
	// 响应头已在 ProxyRequest 的 defer 中从 resp 对象获取

	// 确保设置正确的SSE头部
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	//sse禁用缓存头部
	ctx.Writer.Header().Set("Cache-Control", "no-store, no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 设置响应状态码（已在 ProxyRequest 中设置到上下文）
	ctx.Writer.WriteHeader(resp.StatusCode)
	ctx.SetResponded()

	// 获取Flusher接口用于实时刷新
	flusher, ok := ctx.Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("响应写入器不支持刷新操作")
	}

	// 立即刷新响应头
	flusher.Flush()

	// 使用较小的缓冲区确保实时性（类似nginx proxy_buffering off）
	buffer := make([]byte, 1024) // 1KB缓冲区

	var clientClosed bool
	var lastError error

	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := ctx.Writer.Write(buffer[:n]); writeErr != nil {
				// 关键修复：记录客户端关闭连接的错误，但不返回错误
				// 返回错误会导致 chunked 编码不完整，浏览器会报 ERR_INCOMPLETE_CHUNKED_ENCODING
				// 客户端关闭连接是正常行为（用户关闭页面等），应该优雅处理
				// 记录错误用于监控和诊断，但不返回错误，让 Go 的 HTTP 服务器正确处理 chunked 编码的结束
				errMsg := writeErr.Error()
				if strings.Contains(errMsg, "broken pipe") ||
					strings.Contains(errMsg, "connection reset") ||
					strings.Contains(errMsg, "use of closed network connection") {
					// 客户端关闭连接，记录错误但不返回
					clientClosed = true
					lastError = writeErr
					ctx.AddError(fmt.Errorf("SSE client closed connection: %w", writeErr))
					break
				}
				// 其他写入错误，也记录但不返回
				lastError = writeErr
				ctx.AddError(fmt.Errorf("SSE write error: %w", writeErr))
				break
			}
			// 每次写入后立即刷新，确保数据实时到达客户端
			flusher.Flush()
		}
		if err != nil {
			if err == io.EOF {
				// 服务器端正常结束
				break
			}
			// 关键修复：记录读取错误，但不返回错误
			// 返回错误会导致 chunked 编码不完整
			errMsg := err.Error()
			if strings.Contains(errMsg, "broken pipe") ||
				strings.Contains(errMsg, "connection reset") ||
				strings.Contains(errMsg, "use of closed network connection") {
				// 客户端关闭连接，记录错误但不返回
				clientClosed = true
				lastError = err
				ctx.AddError(fmt.Errorf("SSE client closed connection during read: %w", err))
				break
			}
			// 其他读取错误，也记录但不返回
			lastError = err
			ctx.AddError(fmt.Errorf("SSE read error: %w", err))
			break
		}
	}

	// 如果客户端关闭连接，记录日志但不返回错误
	// 这样可以确保 chunked 编码正确结束，避免 ERR_INCOMPLETE_CHUNKED_ENCODING 错误
	if clientClosed && lastError != nil {
		// 错误已通过 ctx.AddError 记录，这里不返回错误
		// 让 Go 的 HTTP 服务器正确处理 chunked 编码的结束
	}

	// responseTime 由网关流程结束时设置（gateway.go），不在代理处理中设置
	return nil
}

// handleRegularResponse 处理常规HTTP响应
func (h *HTTPProxy) handleRegularResponse(ctx *core.Context, resp *http.Response) error {
	// 复制响应头
	// 注意：响应头不再保存到上下文，因为多服务转发时每个服务的响应头不同
	// 响应头已在 ProxyRequest 的 defer 中从 resp 对象获取
	for name, values := range resp.Header {
		for _, value := range values {
			ctx.Writer.Header().Add(name, value)
		}
	}

	// 设置响应状态码（已在 ProxyRequest 中设置）
	ctx.Writer.WriteHeader(resp.StatusCode)
	// 标记为已响应（responseTime 由网关流程结束时设置，不在代理处理中设置）
	ctx.SetResponded()
	// 重置 responseTime，由网关流程结束时统一设置
	ctx.SetResponseTime(time.Time{})

	// 复制响应体
	config := h.GetHTTPConfig()
	// 根据HTTP配置和日志配置决定是否需要复制响应体
	shouldCopyBody := config.CopyResponseBody || h.shouldRecordResponseBody(ctx)
	if shouldCopyBody {
		// 如果需要复制响应体到上下文中（用于日志记录或HTTP配置要求）
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应体失败: %w", err)
		}
		// 只有在需要记录响应体时才保存到上下文
		if h.shouldRecordResponseBody(ctx) {
			ctx.Set("response_body", bodyBytes)
		}
		_, err = ctx.Writer.Write(bodyBytes)
		if err != nil {
			return fmt.Errorf("写入响应体失败: %w", err)
		}
	} else {
		// 直接流式复制
		_, err := io.Copy(ctx.Writer, resp.Body)
		if err != nil {
			return fmt.Errorf("复制响应体失败: %w", err)
		}
	}

	// responseTime 由网关流程结束时设置（gateway.go），不在代理处理中设置
	return nil
}

// isHopByHopHeader 检查是否为hop-by-hop头部
// 根据RFC 7230 Section 6.1，这些头部不应该被代理转发
func isHopByHopHeader(name string) bool {
	// 标准的hop-by-hop头部（不区分大小写）
	switch strings.ToLower(name) {
	case "connection",
		"keep-alive",
		"proxy-authenticate",
		"proxy-authorization",
		"te",
		"trailers",
		"upgrade":
		return true
	case "host":
		// Host头部需要特殊处理 - 代理需要设置正确的目标Host
		return true
	case "content-length":
		// Content-Length在有Transfer-Encoding时会被覆盖
		// Go的HTTP客户端会自动计算正确的Content-Length
		return true
	case "transfer-encoding":
		// ⚠️ 对于SSE，我们需要保留Transfer-Encoding: chunked
		// 这里应该根据响应类型决定是否移除
		// 暂时保留原有逻辑，后续在复制头部时特殊处理
		return true
	default:
		return false
	}
}

// isConnectionHeader 检查头部是否在Connection头部中列出
// Connection头部可以列出额外的hop-by-hop头部
func isConnectionHeader(name string, headers http.Header) bool {
	connectionHeaders := headers.Get("Connection")
	if connectionHeaders == "" {
		return false
	}

	// 解析Connection头部中的token
	for _, token := range strings.Split(connectionHeaders, ",") {
		token = strings.TrimSpace(token)
		if strings.EqualFold(name, token) {
			return true
		}
	}
	return false
}

// setProxyHeaders 设置代理头部
func (h *HTTPProxy) setProxyHeaders(req *http.Request, proxyReq *http.Request, targetHost string) {
	// 获取配置，如果没有配置则使用默认值
	config := h.GetHTTPConfig()

	// 1. 设置标准代理头部
	if config.SetHeaders != nil {
		for key, value := range config.SetHeaders {
			proxyReq.Header.Set(key, value)
		}
	}

	// 2. 设置默认的代理头部（如果没有在配置中设置）
	if proxyReq.Header.Get("User-Agent") == "" {
		proxyReq.Header.Set("User-Agent", "Gateway-Gateway/1.0")
	}

	// 3. 设置X-Forwarded-* 头部
	if config.AddXForwardedFor {
		if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
			proxyReq.Header.Set("X-Forwarded-For", xff+", "+h.getClientIP(req))
		} else {
			proxyReq.Header.Set("X-Forwarded-For", h.getClientIP(req))
		}
	}

	if config.AddXRealIP {
		proxyReq.Header.Set("X-Real-IP", h.getClientIP(req))
	}

	if config.AddXForwardedProto {
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		proxyReq.Header.Set("X-Forwarded-Proto", scheme)
	}

	if config.AddXForwardedFor { // 如果启用了X-Forwarded-For，也启用X-Forwarded-Host
		proxyReq.Header.Set("X-Forwarded-Host", req.Host)
	}

	// 4. 处理Host头部
	if config.PreserveHost {
		// 保留原始Host头部 - 使用req.Host而不是header中的Host
		if req.Host != "" {
			proxyReq.Host = req.Host
			proxyReq.Header.Set("Host", req.Host)
		}
	} else {
		// 设置为目标主机（默认行为）
		proxyReq.Header.Set("Host", targetHost)
	}

	// 5. 处理需要隐藏的头部
	if config.HideHeaders != nil {
		for _, headerName := range config.HideHeaders {
			proxyReq.Header.Del(headerName)
		}
	}

	// 6. 处理需要明确传递的头部
	passHeaders := config.PassHeaders

	if passHeaders != nil && len(passHeaders) > 0 {
		// 创建一个新的头部map，只包含允许的头部
		allowedHeaders := make(map[string]bool)
		for _, headerName := range passHeaders {
			allowedHeaders[strings.ToLower(headerName)] = true
		}

		// 过滤头部
		for name := range proxyReq.Header {
			if !allowedHeaders[strings.ToLower(name)] && !isSystemHeader(name) {
				proxyReq.Header.Del(name)
			}
		}
	}

	// 7. 设置Connection头部 - 根据HTTP版本和KeepAlive配置
	if config.KeepAlive && config.HTTPVersion == "1.1" {
		proxyReq.Header.Set("Connection", "")
	} else {
		proxyReq.Header.Set("Connection", "close")
	}
}

// 系统头部不应该被proxy_pass_header过滤
func isSystemHeader(name string) bool {
	systemHeaders := map[string]bool{
		"host":              true,
		"x-forwarded-for":   true,
		"x-real-ip":         true,
		"x-forwarded-proto": true,
		"x-forwarded-host":  true,
		"user-agent":        true,
	}
	return systemHeaders[strings.ToLower(name)]
}

// getClientIP 获取客户端真实IP
func (h *HTTPProxy) getClientIP(req *http.Request) string {
	// 优先级：X-Forwarded-For > X-Real-IP > RemoteAddr

	// 1. 检查X-Forwarded-For头部
	if xff := req.Header.Get("X-Forwarded-For"); xff != "" {
		// 取第一个IP（原始客户端IP）
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// 2. 检查X-Real-IP头部
	if xri := req.Header.Get("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. 使用RemoteAddr
	if ip, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		return ip
	}

	return req.RemoteAddr
}

// parseTLSVersion 解析TLS版本字符串为crypto/tls常量
func parseTLSVersion(version string) uint16 {
	switch version {
	case "1.0":
		return tls.VersionTLS10
	case "1.1":
		return tls.VersionTLS11
	case "1.2":
		return tls.VersionTLS12
	case "1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12 // 默认使用TLS 1.2
	}
}

// createHTTPClient 创建HTTP客户端
func (h *HTTPProxy) createHTTPClient(config HTTPProxyConfig) *http.Client {
	// 设置超时配置
	connectTimeout := config.ConnectTimeout
	if connectTimeout == 0 {
		connectTimeout = config.Timeout
	}
	if connectTimeout == 0 {
		connectTimeout = 30 * time.Second
	}

	readTimeout := config.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}

	// 创建TLS配置
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.TLSInsecureSkipVerify,
		MinVersion:         parseTLSVersion(config.TLSMinVersion),
		ServerName:         config.TLSServerName,
	}

	// 设置最大TLS版本（如果配置了）
	if config.TLSMaxVersion != "" {
		tlsConfig.MaxVersion = parseTLSVersion(config.TLSMaxVersion)
	}

	// 根据是否启用代理缓冲来调整缓冲区大小
	readBufferSize := config.BufferSize
	writeBufferSize := config.BufferSize

	// 如果禁用代理缓冲（通常用于SSE等实时流），使用更小的缓冲区
	if !config.ProxyBuffering {
		readBufferSize = 1024  // 1KB，更适合实时流
		writeBufferSize = 1024 // 1KB，更适合实时流
	}

	// 创建传输层配置
	transport := &http.Transport{
		// 连接池配置
		MaxIdleConns:        config.MaxIdleConns,     // 全局最大空闲连接数
		MaxIdleConnsPerHost: config.MaxIdleConns / 4, // 每个主机的最大空闲连接数
		MaxConnsPerHost:     config.MaxIdleConns * 2, // 每个主机的最大连接数
		IdleConnTimeout:     config.IdleConnTimeout,  // 空闲连接超时

		// 超时配置
		TLSHandshakeTimeout:   10 * time.Second, // TLS握手超时
		ResponseHeaderTimeout: readTimeout,      // 响应头超时
		ExpectContinueTimeout: 1 * time.Second,  // 100-continue超时

		// Keep-Alive配置
		DisableKeepAlives: !config.KeepAlive, // 根据配置决定是否禁用Keep-Alive

		// 缓冲区配置 - 根据代理缓冲设置动态调整
		ReadBufferSize:  readBufferSize,  // 读缓冲区大小
		WriteBufferSize: writeBufferSize, // 写缓冲区大小

		// 连接拨号配置
		DialContext: (&net.Dialer{
			Timeout:   connectTimeout,   // 连接超时
			KeepAlive: 30 * time.Second, // TCP Keep-Alive间隔
		}).DialContext,

		// 使用配置的TLS设置
		TLSClientConfig: tlsConfig,
	}

	// 创建客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   config.Timeout, // 总超时时间
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !config.FollowRedirects {
				return http.ErrUseLastResponse
			}
			// 限制重定向次数避免无限循环
			if len(via) >= 5 {
				return fmt.Errorf("太多重定向")
			}
			return nil
		},
	}

	return client
}

// selectTargetNode 选择目标节点，支持服务注册中心发现
// 返回服务配置和节点配置
func (h *HTTPProxy) selectTargetNode(ctx *core.Context, serviceID string) (*service.ServiceConfig, *service.NodeConfig, error) {
	// 首先尝试从服务管理器获取服务配置
	serviceConfig, exists := h.serviceManager.GetService(serviceID)
	if !exists {
		return nil, nil, fmt.Errorf("服务 %s 不存在", serviceID)
	}

	// 检查是否为注册中心服务
	if proxyutils.IsRegistryService(serviceConfig.ServiceMetadata) {
		// 使用注册中心服务发现
		node, err := h.selectNodeFromRegistry(ctx, serviceConfig)
		if err != nil {
			return nil, nil, err
		}
		return serviceConfig, node, nil
	}

	// 使用传统的负载均衡选择节点
	node, err := h.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		return nil, nil, err
	}
	return serviceConfig, node, nil
}

// selectNodeFromRegistry 从注册中心选择节点
func (h *HTTPProxy) selectNodeFromRegistry(ctx *core.Context, serviceConfig *service.ServiceConfig) (*service.NodeConfig, error) {
	// 使用静态方法从注册中心创建节点配置（内部会自动获取注册中心管理器实例）
	node, err := proxyutils.CreateNodeFromRegistry(ctx, serviceConfig)
	if err != nil {
		return nil, fmt.Errorf("从注册中心创建节点配置失败: %w", err)
	}

	// 记录服务发现结果到上下文
	ctx.Set("discovery_type", "REGISTRY")
	ctx.Set("discovered_instance", map[string]interface{}{
		"instanceId":     node.ID,
		"url":            node.URL,
		"tenantId":       node.Metadata["tenantId"],
		"serviceGroupId": node.Metadata["serviceGroupId"],
		"serviceName":    node.Metadata["serviceName"],
		"healthStatus":   node.Metadata["healthStatus"],
	})

	return node, nil
}

// GetRegistryManager 获取注册中心管理器
func (h *HTTPProxy) GetRegistryManager() *registryManager.RegistryManager {
	return registryManager.GetInstance()
}

// shouldRecordRequestBody 检查是否应该记录请求体（根据日志配置）
func (h *HTTPProxy) shouldRecordRequestBody(ctx *core.Context) bool {
	// 直接从上下文获取日志配置，避免重复获取
	config := ctx.GetLogConfig()
	if config == nil {
		return false
	}
	return config.IsRecordRequestBody()
}

// shouldRecordResponseBody 检查是否应该记录响应体（根据日志配置）
func (h *HTTPProxy) shouldRecordResponseBody(ctx *core.Context) bool {
	// 直接从上下文获取日志配置，避免重复获取
	config := ctx.GetLogConfig()
	if config == nil {
		return false
	}
	return config.IsRecordResponseBody()
}
