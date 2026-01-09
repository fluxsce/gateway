// Package httpclient 提供HTTP客户端封装实现
//
// 本文件包含DefaultClient的核心实现，包括：
// - 客户端创建和配置
// - HTTP方法（GET、POST、PUT、DELETE等）
// - 请求执行和重试逻辑
// - 辅助方法（URL构建、请求体准备、请求头设置等）
//
// 设计说明：
// 所有方法保持在同一文件中，因为它们逻辑紧密相关，
// 且当前文件大小适中（约500行），便于维护和查阅。
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultClient 默认HTTP客户端实现
//
// 提供线程安全的HTTP客户端，支持以下功能：
// - 默认请求头管理
// - 请求/响应拦截器
// - 自动重试机制
// - 连接池管理
// - 流式响应处理
//
// 线程安全说明：
// - SetDefaultHeader、SetTimeout、AddInterceptor 方法是线程安全的
// - 所有请求方法（Get、Post等）可以并发调用
//
// 使用示例：
//
//	client, err := NewClient(nil) // 使用默认配置
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer client.Close()
//
//	resp, err := client.Get(ctx, "https://api.example.com/users")
type DefaultClient struct {
	client          *http.Client                                       // 底层HTTP客户端
	config          *ClientConfig                                      // 客户端配置
	defaultHeaders  map[string]string                                  // 默认请求头
	interceptors    []Interceptor                                      // 拦截器列表
	mutex           sync.RWMutex                                       // 读写锁，保护并发访问
	checkRedirect   func(req *http.Request, via []*http.Request) error // 重定向检查函数
	maxRedirects    int                                                // 最大重定向次数
	followRedirects bool                                               // 是否跟随重定向
}

// NewClient 创建新的HTTP客户端
//
// 参数：
//   - config: 客户端配置，如果为nil则使用默认配置
//
// 返回：
//   - *DefaultClient: 初始化的HTTP客户端实例
//   - error: 配置验证失败时返回错误
//
// 默认配置：
//   - Timeout: 30秒
//   - FollowRedirects: false（不跟随重定向）
//   - MaxRedirects: 10
//   - MaxIdleConns: 100
//   - MaxIdleConnsPerHost: 10
//   - IdleConnTimeout: 90秒
//   - Retry: nil（默认不重试，需用户显式配置）
//
// 使用示例：
//
//	// 使用默认配置
//	client, err := NewClient(nil)
//
//	// 使用自定义配置（启用重定向和重试）
//	client, err := NewClient(&ClientConfig{
//	    Timeout:         60 * time.Second,
//	    MaxRedirects:    5,
//	    FollowRedirects: true,  // 默认false，需显式启用
//	    Retry: &RetryConfig{    // 默认nil，需显式配置
//	        MaxRetries: 3,
//	        RetryDelay: time.Second,
//	    },
//	})
func NewClient(config *ClientConfig) (*DefaultClient, error) {
	if config == nil {
		config = &ClientConfig{
			Timeout:             30 * time.Second,
			FollowRedirects:     false, // 默认不跟随重定向
			MaxRedirects:        10,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
		}
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	// 创建HTTP客户端
	// 注意：不在http.Client上设置Timeout，而是通过context + Transport层超时控制
	// 原因：http.Client.Timeout包含读取响应体的时间，对流式响应不友好
	client := &http.Client{}

	// 设置Transport
	if config.Transport != nil {
		client.Transport = config.Transport
	} else {
		// 设置超时默认值
		dialTimeout := config.DialTimeout
		if dialTimeout <= 0 {
			dialTimeout = 10 * time.Second
		}
		tlsTimeout := config.TLSHandshakeTimeout
		if tlsTimeout <= 0 {
			tlsTimeout = 10 * time.Second
		}
		responseHeaderTimeout := config.ResponseHeaderTimeout
		if responseHeaderTimeout <= 0 {
			responseHeaderTimeout = 30 * time.Second
		}
		expectContinueTimeout := config.ExpectContinueTimeout
		if expectContinueTimeout <= 0 {
			expectContinueTimeout = 1 * time.Second
		}

		// 使用默认Transport并应用超时和连接池配置
		transport := &http.Transport{
			// 分层超时配置（对应不同阶段）
			DialContext: (&net.Dialer{
				Timeout:   dialTimeout,      // TCP连接超时
				KeepAlive: 30 * time.Second, // TCP KeepAlive
			}).DialContext,
			TLSHandshakeTimeout:   tlsTimeout,            // TLS握手超时
			ResponseHeaderTimeout: responseHeaderTimeout, // 等待响应头超时
			ExpectContinueTimeout: expectContinueTimeout, // 100-continue超时

			// 连接池配置
			MaxIdleConns:        config.MaxIdleConns,
			MaxIdleConnsPerHost: config.MaxIdleConnsPerHost,
			IdleConnTimeout:     config.IdleConnTimeout,
		}
		if config.MaxIdleConns == 0 {
			transport.MaxIdleConns = 100 // 默认值
		}
		if config.MaxIdleConnsPerHost == 0 {
			transport.MaxIdleConnsPerHost = 10 // 默认值
		}
		if config.IdleConnTimeout == 0 {
			transport.IdleConnTimeout = 90 * time.Second // 默认值
		}
		if config.MaxConnsPerHost > 0 {
			transport.MaxConnsPerHost = config.MaxConnsPerHost
		}
		client.Transport = transport
	}

	// 设置重定向处理
	maxRedirects := config.MaxRedirects
	if maxRedirects <= 0 {
		maxRedirects = 10
	}

	checkRedirect := config.CheckRedirect
	if checkRedirect == nil && config.FollowRedirects {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return ErrMaxRedirectsExceeded
			}
			return nil
		}
	} else if !config.FollowRedirects {
		checkRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	client.CheckRedirect = checkRedirect

	return &DefaultClient{
		client:          client,
		config:          config,
		defaultHeaders:  make(map[string]string),
		interceptors:    make([]Interceptor, 0),
		maxRedirects:    maxRedirects,
		followRedirects: config.FollowRedirects,
		checkRedirect:   checkRedirect,
	}, nil
}

// SetDefaultHeader 设置默认请求头
//
// 设置的请求头会自动添加到所有请求中。
// 可多次调用设置多个默认请求头，相同key会被覆盖。
// 此方法是线程安全的。
//
// 参数：
//   - key: 请求头名称
//   - value: 请求头值
//
// 使用示例：
//
//	client.SetDefaultHeader("X-API-Key", "your-api-key")
//	client.SetDefaultHeader("User-Agent", "MyApp/1.0")
func (c *DefaultClient) SetDefaultHeader(key, value string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.defaultHeaders == nil {
		c.defaultHeaders = make(map[string]string)
	}
	c.defaultHeaders[key] = value
}

// SetTimeout 设置默认超时时间
//
// 动态修改客户端的默认超时时间。
// 此方法是线程安全的。
//
// 超时机制说明：
// - 超时通过context控制，而非http.Client.Timeout
// - 这样对流式响应更友好（不会在读取大文件时超时）
// - 超时仅控制连接建立和获取响应头的时间
//
// 超时优先级：
// 1. 请求级别：WithTimeout(duration)
// 2. 客户端级别：SetTimeout(duration) 或 ClientConfig.Timeout
// 3. 用户传入的ctx：context.WithTimeout(ctx, duration)
//
// 参数：
//   - timeout: 超时时间，0表示不设置超时（使用用户ctx的超时）
func (c *DefaultClient) SetTimeout(timeout time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.config.Timeout = timeout
}

// AddInterceptor 添加请求/响应拦截器
//
// 拦截器按添加顺序执行，可用于：
// - 添加通用请求头（如认证信息）
// - 请求/响应日志记录
// - 请求修改或响应处理
// 此方法是线程安全的。
//
// 参数：
//   - interceptor: 拦截器函数
//
// 使用示例：
//
//	client.AddInterceptor(func(ctx *InterceptorContext) error {
//	    // 请求前：添加请求头
//	    if ctx.Response == nil {
//	        ctx.Request.Header.Set("X-Request-ID", uuid.New().String())
//	    }
//	    // 响应后：记录日志
//	    if ctx.Response != nil {
//	        log.Printf("Status: %d", ctx.Response.StatusCode)
//	    }
//	    return nil
//	})
func (c *DefaultClient) AddInterceptor(interceptor Interceptor) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.interceptors = append(c.interceptors, interceptor)
}

// Get 发送GET请求
//
// 参数：
//   - ctx: 上下文，用于取消请求或设置截止时间
//   - requestURL: 请求URL（可以是完整URL或相对路径）
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象，包含状态码、响应头、响应体等
//   - error: 请求失败时返回错误
//
// 使用示例：
//
//	resp, err := client.Get(ctx, "/users",
//	    WithQueryParam("page", "1"),
//	    WithHeader("Accept", "application/json"),
//	)
func (c *DefaultClient) Get(ctx context.Context, requestURL string, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodGet, requestURL, nil, opts...)
}

// Post 发送POST请求
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - body: 请求体，支持以下类型：
//   - io.Reader: 直接使用
//   - []byte: 转换为bytes.Reader
//   - string: 转换为strings.Reader
//   - 其他类型: 自动序列化为JSON
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
//
// 使用示例：
//
//	// 发送JSON
//	resp, err := client.Post(ctx, "/users", map[string]string{"name": "John"})
//
//	// 发送表单
//	resp, err := client.Post(ctx, "/login", "username=admin&password=123",
//	    WithContentType("application/x-www-form-urlencoded"),
//	)
func (c *DefaultClient) Post(ctx context.Context, requestURL string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodPost, requestURL, body, opts...)
}

// Put 发送PUT请求
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - body: 请求体（同Post方法）
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) Put(ctx context.Context, requestURL string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodPut, requestURL, body, opts...)
}

// Delete 发送DELETE请求
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) Delete(ctx context.Context, requestURL string, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodDelete, requestURL, nil, opts...)
}

// Patch 发送PATCH请求
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - body: 请求体（同Post方法）
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) Patch(ctx context.Context, requestURL string, body interface{}, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodPatch, requestURL, body, opts...)
}

// Head 发送HEAD请求
//
// HEAD请求只返回响应头，不返回响应体。
// 常用于检查资源是否存在或获取资源元信息。
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象（Body为空）
//   - error: 请求失败时返回错误
func (c *DefaultClient) Head(ctx context.Context, requestURL string, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodHead, requestURL, nil, opts...)
}

// Options 发送OPTIONS请求
//
// OPTIONS请求用于获取服务器支持的HTTP方法。
// 常用于CORS预检请求。
//
// 参数：
//   - ctx: 上下文
//   - requestURL: 请求URL
//   - opts: 可选的请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) Options(ctx context.Context, requestURL string, opts ...RequestOption) (*Response, error) {
	return c.doRequest(ctx, http.MethodOptions, requestURL, nil, opts...)
}

// Do 发送自定义HTTP请求
//
// 用于发送自定义构建的http.Request，适用于需要完全控制请求的场景。
// 响应体会完整读取到内存中。
//
// 参数：
//   - ctx: 上下文
//   - req: 自定义的HTTP请求对象
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
//
// 注意：如需处理大响应体或流式响应，请使用DoStream方法
func (c *DefaultClient) Do(ctx context.Context, req *http.Request) (*Response, error) {
	return c.do(ctx, req, false)
}

// DoStream 发送自定义HTTP请求（流式响应）
//
// 响应体不会完整读取到内存，而是通过BodyReader流式访问。
// 适用于下载大文件或处理流式API响应。
//
// 参数：
//   - ctx: 上下文
//   - req: 自定义的HTTP请求对象
//
// 返回：
//   - *Response: 响应对象，通过BodyReader访问响应体
//   - error: 请求失败时返回错误
//
// 注意：调用者负责关闭BodyReader
//
// 使用示例：
//
//	resp, err := client.DoStream(ctx, req)
//	if err != nil {
//	    return err
//	}
//	defer resp.BodyReader.Close()
//	io.Copy(file, resp.BodyReader)
func (c *DefaultClient) DoStream(ctx context.Context, req *http.Request) (*Response, error) {
	return c.do(ctx, req, true)
}

// do 执行HTTP请求的内部方法
//
// 这是所有请求执行的核心方法，负责：
// 1. 执行请求拦截器
// 2. 发送HTTP请求
// 3. 处理响应（流式或非流式）
// 4. 执行响应拦截器
//
// 参数：
//   - ctx: 上下文
//   - req: HTTP请求对象
//   - streamResponse: 是否使用流式响应
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) do(ctx context.Context, req *http.Request, streamResponse bool) (*Response, error) {
	// 执行请求拦截器
	interceptorCtx := &InterceptorContext{
		Request: req,
		Context: ctx,
	}

	// 线程安全：在锁内复制拦截器切片
	c.mutex.RLock()
	var interceptors []Interceptor
	if len(c.interceptors) > 0 {
		interceptors = make([]Interceptor, len(c.interceptors))
		copy(interceptors, c.interceptors)
	}
	c.mutex.RUnlock()

	for _, interceptor := range interceptors {
		if err := interceptor(interceptorCtx); err != nil {
			return nil, err
		}
	}

	// 执行请求
	resp, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		// 执行响应拦截器（即使出错）
		interceptorCtx.Error = err
		for _, interceptor := range interceptors {
			_ = interceptor(interceptorCtx)
		}
		return nil, NewHTTPError(ErrCodeRequestFailed, "request failed", err)
	}

	response := &Response{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
		Request:       req,
		Response:      resp,
	}

	if streamResponse {
		// 流式响应：不读取响应体，直接返回BodyReader
		response.BodyReader = resp.Body
		// 注意：调用者需要负责关闭BodyReader
	} else {
		// 非流式响应：完整读取响应体
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close() // 立即关闭，不使用defer避免后续重复关闭
		if err != nil {
			return nil, ErrResponseReadFailed
		}
		response.Body = body
	}

	// 执行响应拦截器
	interceptorCtx.Response = response
	for _, interceptor := range interceptors {
		if err := interceptor(interceptorCtx); err != nil {
			// 流式响应时，拦截器出错需要关闭BodyReader
			if streamResponse && response.BodyReader != nil {
				response.BodyReader.Close()
			}
			return response, err
		}
	}

	return response, nil
}

// Close 关闭客户端并释放资源
//
// 关闭所有空闲连接，释放连接池资源。
// 建议在应用退出前调用此方法。
//
// 返回：
//   - error: 始终返回nil（保留用于未来扩展）
//
// 使用示例：
//
//	client, _ := NewClient(nil)
//	defer client.Close()
func (c *DefaultClient) Close() error {
	if c.client != nil && c.client.Transport != nil {
		if transport, ok := c.client.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	return nil
}

// doRequest 执行HTTP请求的内部方法
//
// 这是所有HTTP方法（Get、Post等）的统一入口，负责：
// 1. 解析URL并添加查询参数
// 2. 应用请求选项
// 3. 准备请求体
// 4. 创建HTTP请求
// 5. 设置请求头和认证
// 6. 执行请求（带重试）
//
// 参数：
//   - ctx: 上下文
//   - method: HTTP方法（GET、POST等）
//   - requestURL: 完整的请求URL
//   - body: 请求体（可为nil）
//   - opts: 请求配置选项
//
// 返回：
//   - *Response: 响应对象
//   - error: 请求失败时返回错误
func (c *DefaultClient) doRequest(ctx context.Context, method, requestURL string, body interface{}, opts ...RequestOption) (*Response, error) {
	// 验证URL
	if requestURL == "" {
		return nil, ErrInvalidURL
	}

	// 解析URL
	parsedURL, err := url.Parse(requestURL)
	if err != nil {
		return nil, NewHTTPError(ErrCodeInvalidURL, "failed to parse URL", err)
	}

	// 构建请求配置
	reqConfig := &RequestConfig{
		FollowRedirects: c.followRedirects,
		MaxRedirects:    c.maxRedirects,
	}

	// 应用选项
	for _, opt := range opts {
		opt(reqConfig)
	}

	// 添加查询参数
	if len(reqConfig.QueryParams) > 0 {
		query := parsedURL.Query()
		for k, v := range reqConfig.QueryParams {
			query.Set(k, fmt.Sprintf("%v", v))
		}
		parsedURL.RawQuery = query.Encode()
	}

	// 准备请求体
	var bodyReader io.Reader
	var contentType string
	if body != nil {
		bodyReader, contentType, err = c.prepareBody(body, reqConfig.ContentType)
		if err != nil {
			return nil, err
		}
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, parsedURL.String(), bodyReader)
	if err != nil {
		return nil, NewHTTPError(ErrCodeCreateRequestFailed, "failed to create request", err)
	}

	// 设置请求头
	c.setHeaders(req, reqConfig, contentType)

	// 设置认证
	c.setAuth(req, reqConfig)

	// 设置Cookie
	for _, cookie := range reqConfig.Cookies {
		req.AddCookie(cookie)
	}

	// 设置超时
	// 优先级：请求级别超时 > 客户端级别超时 > 用户传入的ctx超时
	//
	// 注意：流式响应(StreamResponse=true)时，不设置context超时
	// 原因：context超时会中断响应体读取，导致大文件下载失败
	// 保护：ResponseHeaderTimeout已保护响应头阶段，响应体读取由用户控制
	timeout := reqConfig.Timeout
	if timeout <= 0 {
		timeout = c.config.Timeout
	}
	if timeout > 0 && !reqConfig.StreamResponse {
		// 非流式响应：设置context超时
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
		req = req.WithContext(ctx)
	}
	// 流式响应：不设置context超时，由ResponseHeaderTimeout保护响应头阶段
	// 响应体读取阶段由用户自行控制（可通过传入带超时的ctx实现）

	// 执行请求（带重试）
	return c.doRequestWithRetry(ctx, req, reqConfig)
}

// doRequestWithRetry 带重试的请求执行
//
// 根据重试配置自动重试失败的请求。
// 注意：默认不启用重试（Retry为nil时），需用户通过以下方式显式配置：
//   - 客户端级别：NewClient(&ClientConfig{Retry: &RetryConfig{...}})
//   - 请求级别：client.Get(ctx, url, WithRetry(&RetryConfig{...}))
//
// 重试策略：
//   - 如果配置了RetryOnStatusCodes，只对指定状态码重试
//   - 否则默认对5xx服务器错误重试
//   - 如果配置了RetryOnErrors，对指定错误类型重试
//
// 参数：
//   - ctx: 上下文
//   - req: HTTP请求对象
//   - config: 请求配置（包含重试配置）
//
// 返回：
//   - *Response: 响应对象
//   - error: 所有重试都失败时返回错误
//
// 重试限制：
//   - 流式响应不支持重试，会自动降级为非流式响应
//   - 带请求体的请求（POST/PUT/PATCH）：如果body是io.Reader，重试时无法重新读取
//     建议使用[]byte或string作为body，或确保body实现io.Seeker接口
func (c *DefaultClient) doRequestWithRetry(ctx context.Context, req *http.Request, config *RequestConfig) (*Response, error) {
	retryConfig := config.Retry
	if retryConfig == nil {
		retryConfig = c.config.Retry
	}

	// 流式响应不支持重试（因为响应体已打开）
	useStreaming := config.StreamResponse
	if retryConfig == nil || retryConfig.MaxRetries <= 0 {
		return c.do(ctx, req, useStreaming)
	}

	// 如果有重试配置，流式响应需要降级为非流式（因为重试需要重新读取响应体）
	if useStreaming {
		// 流式响应不支持重试，因为响应体已经打开
		// 降级为非流式响应进行重试
		useStreaming = false
	}

	// 保存原始请求体用于重试
	// 注意：只有GetBody被设置时才能安全重试带body的请求
	var getBody func() (io.ReadCloser, error)
	if req.Body != nil && req.GetBody != nil {
		getBody = req.GetBody
	} else if req.Body != nil {
		// 尝试读取并保存body内容（仅适用于小body）
		bodyBytes, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err == nil && len(bodyBytes) > 0 {
			getBody = func() (io.ReadCloser, error) {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
			// 恢复body供首次请求使用
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
			req.GetBody = getBody
		}
	}

	var lastErr error
	var lastResp *Response

	for attempt := 0; attempt <= retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			// 等待重试延迟
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(retryConfig.RetryDelay):
			}
			// 重试时需要重新设置请求体
			if getBody != nil {
				body, err := getBody()
				if err != nil {
					return nil, NewHTTPError(ErrCodeRequestFailed, "failed to reset request body for retry", err)
				}
				req.Body = body
			}
		}

		resp, err := c.do(ctx, req, useStreaming)
		if err == nil {
			// 检查状态码是否需要重试
			if retryConfig.RetryOnStatusCodes != nil {
				shouldRetry := false
				for _, code := range retryConfig.RetryOnStatusCodes {
					if resp.StatusCode == code {
						shouldRetry = true
						break
					}
				}
				if !shouldRetry {
					return resp, nil
				}
			} else {
				// 默认只对5xx错误重试
				if resp.IsServerError() {
					lastResp = resp
					continue
				}
				return resp, nil
			}
		} else {
			// 检查错误是否需要重试（使用errors.Is进行错误比较）
			if retryConfig.RetryOnErrors != nil {
				shouldRetry := false
				for _, retryErr := range retryConfig.RetryOnErrors {
					if errors.Is(err, retryErr) {
						shouldRetry = true
						break
					}
				}
				if !shouldRetry {
					return nil, err
				}
			}
			lastErr = err
		}

		lastResp = resp
	}

	if lastErr != nil {
		return nil, lastErr
	}

	if lastResp != nil {
		return lastResp, nil
	}

	return nil, NewHTTPError(ErrCodeRetryExhausted, "retry exhausted", nil)
}

// prepareBody 准备请求体
//
// 将各种类型的body转换为io.Reader，并确定Content-Type。
//
// 支持的body类型（对应Postman传输类型）：
// - nil: 返回nil reader
// - io.Reader: 直接使用（binary模式）
// - *FormData: multipart/form-data（form-data模式，支持文件上传）
// - *FormURLEncoded: application/x-www-form-urlencoded
// - url.Values: application/x-www-form-urlencoded
// - BodyReader接口: 自定义请求体
// - []byte: 原始字节（raw模式）
// - string: 原始字符串（raw模式）
// - 其他类型: 自动序列化为JSON（raw/json模式）
//
// 参数：
//   - body: 请求体数据
//   - contentType: 指定的Content-Type（可为空）
//
// 返回：
//   - io.Reader: 请求体reader
//   - string: Content-Type
//   - error: 序列化失败时返回错误
func (c *DefaultClient) prepareBody(body interface{}, contentType string) (io.Reader, string, error) {
	if body == nil {
		return nil, "", nil
	}

	// 如果已经是io.Reader（binary模式）
	if reader, ok := body.(io.Reader); ok {
		return reader, contentType, nil
	}

	// 如果是FormData（multipart/form-data，支持大文件）
	if formData, ok := body.(*FormData); ok {
		return c.prepareFormData(formData)
	}

	// 如果是FormURLEncoded（application/x-www-form-urlencoded）
	if formURLEncoded, ok := body.(*FormURLEncoded); ok {
		return c.prepareFormURLEncoded(formURLEncoded)
	}

	// 如果是url.Values（application/x-www-form-urlencoded）
	if urlValues, ok := body.(url.Values); ok {
		encoded := urlValues.Encode()
		return strings.NewReader(encoded), "application/x-www-form-urlencoded", nil
	}

	// 如果是BodyReader接口
	if bodyReader, ok := body.(BodyReader); ok {
		reader, err := bodyReader.Reader()
		if err != nil {
			return nil, "", err
		}
		ct := bodyReader.ContentType()
		if ct != "" {
			contentType = ct
		}
		return reader, contentType, nil
	}

	// 如果是[]byte（raw模式）
	if bodyBytes, ok := body.([]byte); ok {
		return bytes.NewReader(bodyBytes), contentType, nil
	}

	// 如果是string（raw模式）
	if str, ok := body.(string); ok {
		return strings.NewReader(str), contentType, nil
	}

	// 尝试序列化为JSON（raw/json模式，包括GraphQLRequest）
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, "", NewHTTPError(ErrCodeMarshalBodyFailed, "failed to marshal body", err)
	}

	if contentType == "" {
		contentType = "application/json"
	}

	return bytes.NewReader(jsonBytes), contentType, nil
}

// prepareFormURLEncoded 准备application/x-www-form-urlencoded请求体
func (c *DefaultClient) prepareFormURLEncoded(form *FormURLEncoded) (io.Reader, string, error) {
	values := url.Values{}
	for key, value := range form.Fields {
		values.Set(key, value)
	}
	return strings.NewReader(values.Encode()), "application/x-www-form-urlencoded", nil
}

// prepareFormData 准备multipart/form-data请求体
//
// 根据是否有文件自动选择处理方式：
// - 有文件：使用io.Pipe流式处理，避免大文件占用内存
// - 无文件：使用bytes.Buffer同步处理，更简单高效
//
// 资源管理：
// - 通过FilePath打开的文件会记录在FormData.openedFiles中
// - 请求完成后应调用FormData.Close()关闭这些文件
//
// 参数：
//   - formData: 表单数据
//
// 返回：
//   - io.Reader: 请求体reader
//   - string: Content-Type（包含boundary）
//   - error: 构建失败时返回错误
func (c *DefaultClient) prepareFormData(formData *FormData) (io.Reader, string, error) {
	// 预处理：打开需要通过路径读取的文件
	for i := range formData.Files {
		file := &formData.Files[i]
		if file.Reader == nil && file.FilePath != "" {
			f, err := os.Open(file.FilePath)
			if err != nil {
				// 关闭已打开的文件
				formData.Close()
				return nil, "", NewHTTPError(ErrCodeRequestFailed, "failed to open file: "+file.FilePath, err)
			}
			file.Reader = f
			// 如果没有指定文件名，使用路径中的文件名
			if file.FileName == "" {
				file.FileName = filepath.Base(file.FilePath)
			}
			// 记录打开的文件，以便后续关闭
			formData.openedFiles = append(formData.openedFiles, f)
		}
	}

	// 无文件时使用同步方式（更简单高效）
	if len(formData.Files) == 0 {
		return c.prepareFormDataSync(formData)
	}

	// 有文件时使用流式处理（避免大文件占用内存）
	return c.prepareFormDataStream(formData)
}

// prepareFormDataSync 同步方式构建表单（无文件时使用）
func (c *DefaultClient) prepareFormDataSync(formData *FormData) (io.Reader, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 写入普通字段
	for key, value := range formData.Fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, "", NewHTTPError(ErrCodeRequestFailed, "failed to write form field", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, "", NewHTTPError(ErrCodeRequestFailed, "failed to close multipart writer", err)
	}

	return &buf, writer.FormDataContentType(), nil
}

// prepareFormDataStream 流式方式构建表单（有文件时使用）
// 使用io.Pipe + goroutine实现流式上传，避免将大文件完整加载到内存
func (c *DefaultClient) prepareFormDataStream(formData *FormData) (io.Reader, string, error) {
	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer writer.Close()

		// 写入普通字段
		for key, value := range formData.Fields {
			if err := writer.WriteField(key, value); err != nil {
				pw.CloseWithError(err)
				return
			}
		}

		// 写入文件（流式）
		for _, file := range formData.Files {
			if file.Reader == nil {
				continue
			}
			part, err := writer.CreateFormFile(file.FieldName, file.FileName)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			if _, err := io.Copy(part, file.Reader); err != nil {
				pw.CloseWithError(err)
				return
			}
		}
	}()

	return pr, writer.FormDataContentType(), nil
}

// setHeaders 设置请求头
//
// 按以下优先级设置请求头（后设置的会覆盖先设置的）：
// 1. 客户端默认请求头（defaultHeaders）
// 2. 请求配置中的请求头（config.Headers）
// 3. Content-Type（优先使用参数中的contentType，其次使用config.ContentType）
//
// 参数：
//   - req: HTTP请求对象
//   - config: 请求配置
//   - contentType: 从请求体推断的Content-Type
func (c *DefaultClient) setHeaders(req *http.Request, config *RequestConfig, contentType string) {
	// 设置默认请求头（直接访问，减少复制）
	c.mutex.RLock()
	for k, v := range c.defaultHeaders {
		req.Header.Set(k, v)
	}
	c.mutex.RUnlock()

	// 设置配置中的请求头
	if config.Headers != nil {
		for k, v := range config.Headers {
			req.Header.Set(k, v)
		}
	}

	// 设置Content-Type
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else if config.ContentType != "" {
		req.Header.Set("Content-Type", config.ContentType)
	}
}

// setAuth 设置认证信息
//
// 支持的认证方式：
// - Basic认证：设置Authorization: Basic <base64(username:password)>
// - Bearer Token认证：设置Authorization: Bearer <token>
//
// 参数：
//   - req: HTTP请求对象
//   - config: 请求配置（包含认证信息）
//
// 注意：如果同时设置了BasicAuth和BearerToken，BearerToken会覆盖BasicAuth
func (c *DefaultClient) setAuth(req *http.Request, config *RequestConfig) {
	if config.BasicAuth != nil {
		req.SetBasicAuth(config.BasicAuth.Username, config.BasicAuth.Password)
	}

	if config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+config.BearerToken)
	}
}
