package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"gohub/internal/gateway/core"
	"gohub/internal/gateway/handler/service"
)

// HTTPProxy HTTP代理实现
type HTTPProxy struct {
	*BaseProxyHandler
	client         *http.Client
	serviceManager service.ServiceManager
	config         *HTTPProxyConfig
}

// Handle 处理HTTP代理请求
func (h *HTTPProxy) Handle(ctx *core.Context) bool {
	if !h.IsEnabled() {
		return true
	}

	// 获取服务ID
	serviceID := ctx.GetServiceID()
	if serviceID == "" {
		ctx.AddError(fmt.Errorf("服务ID不能为空"))
		ctx.Abort(http.StatusBadRequest, map[string]string{
			"error": "服务ID不能为空",
		})
		return false
	}

	// 从负载均衡器获取目标节点
	node, err := h.serviceManager.SelectNode(serviceID, ctx)
	if err != nil {
		ctx.AddError(fmt.Errorf("选择目标节点失败: %w", err))
		ctx.Abort(http.StatusServiceUnavailable, map[string]string{
			"error": "服务不可用",
		})
		return false
	}

	// 代理请求
	err = h.ProxyRequest(ctx, node.URL)
	if err != nil {
		ctx.AddError(fmt.Errorf("代理请求失败: %w", err))
		ctx.Abort(http.StatusBadGateway, map[string]string{
			"error": "代理请求失败",
		})
		return false
	}

	return true
}

// ProxyRequest 代理请求到指定URL
func (h *HTTPProxy) ProxyRequest(ctx *core.Context, targetURL string) error {
	// 解析目标URL
	target, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("解析目标URL失败: %w", err)
	}

	// 构建代理请求URL
	proxyURL := &url.URL{
		Scheme:   target.Scheme,
		Host:     target.Host,
		Path:     target.Path + ctx.Request.URL.Path,
		RawQuery: ctx.Request.URL.RawQuery,
	}

	// 创建代理请求
	var body io.Reader
	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			return fmt.Errorf("读取请求体失败: %w", err)
		}
		body = bytes.NewReader(bodyBytes)
		// 重置原请求的Body
		ctx.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	proxyReq, err := http.NewRequestWithContext(
		context.Background(),
		ctx.Request.Method,
		proxyURL.String(),
		body,
	)
	if err != nil {
		return fmt.Errorf("创建代理请求失败: %w", err)
	}

	// 复制请求头
	for name, values := range ctx.Request.Header {
		// 跳过一些不应该转发的头部
		if name == "Host" || name == "Connection" || name == "Content-Length" {
			continue
		}
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	// 设置Host头部
	proxyReq.Host = target.Host

	// 发送代理请求
	resp, err := h.client.Do(proxyReq)
	if err != nil {
		return fmt.Errorf("发送代理请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 复制响应头
	for name, values := range resp.Header {
		for _, value := range values {
			ctx.Writer.Header().Add(name, value)
		}
	}

	// 设置响应状态码
	ctx.Writer.WriteHeader(resp.StatusCode)
	ctx.SetResponded()
	// 复制响应体
	if h.GetHTTPConfig().CopyResponseBody {
		// 如果需要复制响应体到上下文中
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("读取响应体失败: %w", err)
		}
		ctx.Set("response_body", bodyBytes)
		_, err = ctx.Writer.Write(bodyBytes)
		if err != nil {
			return fmt.Errorf("写入响应体失败: %w", err)
		}
	} else {
		// 直接流式复制
		_, err = io.Copy(ctx.Writer, resp.Body)
		if err != nil {
			return fmt.Errorf("复制响应体失败: %w", err)
		}
	}

	return nil
}

// GetHTTPConfig 获取HTTP代理配置
func (h *HTTPProxy) GetHTTPConfig() *HTTPProxyConfig {
	return h.config
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
	if h.client != nil {
		if transport, ok := h.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// NewHTTPProxy 创建HTTP代理
func NewHTTPProxy(config ProxyConfig, serviceManager service.ServiceManager) (*HTTPProxy, error) {
	// 解析HTTP特定配置
	httpConfig := DefaultHTTPProxyConfig
	if config.Config != nil {
		parseHTTPConfig(config.Config, &httpConfig)
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		MaxIdleConns:        httpConfig.MaxIdleConns,
		IdleConnTimeout:     httpConfig.IdleConnTimeout,
		TLSHandshakeTimeout: 10 * time.Second,
		DisableCompression:  false,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   httpConfig.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !httpConfig.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= 10 {
				return fmt.Errorf("太多重定向")
			}
			return nil
		},
	}

	return &HTTPProxy{
		BaseProxyHandler: NewBaseProxyHandler(config.Type, config.Enabled, config.Name),
		client:           client,
		serviceManager:   serviceManager,
		config:           &httpConfig,
	}, nil
}

// parseHTTPConfig 解析HTTP配置
func parseHTTPConfig(configMap map[string]interface{}, httpConfig *HTTPProxyConfig) {
	if timeout, ok := configMap["timeout"]; ok {
		if timeoutStr, ok := timeout.(string); ok {
			if d, err := time.ParseDuration(timeoutStr); err == nil {
				httpConfig.Timeout = d
			}
		}
	}

	if followRedirects, ok := configMap["follow_redirects"]; ok {
		if b, ok := followRedirects.(bool); ok {
			httpConfig.FollowRedirects = b
		}
	}

	if keepAlive, ok := configMap["keep_alive"]; ok {
		if b, ok := keepAlive.(bool); ok {
			httpConfig.KeepAlive = b
		}
	}

	if maxIdleConns, ok := configMap["max_idle_conns"]; ok {
		if i, ok := maxIdleConns.(int); ok {
			httpConfig.MaxIdleConns = i
		}
	}

	if retryCount, ok := configMap["retry_count"]; ok {
		if i, ok := retryCount.(int); ok {
			httpConfig.RetryCount = i
		}
	}
}
