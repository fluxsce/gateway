package httpclient

import (
	"context"
	"io"
	"net/http"
	"time"
)

// RequestOption 请求选项函数类型
// 用于配置HTTP请求的各种选项
type RequestOption func(*RequestConfig)

// RequestConfig 请求配置
// 包含HTTP请求的所有配置选项
type RequestConfig struct {
	// Headers 请求头
	Headers map[string]string

	// QueryParams 查询参数
	QueryParams map[string]interface{}

	// Timeout 请求超时时间（覆盖客户端默认超时）
	Timeout time.Duration

	// Retry 重试配置
	Retry *RetryConfig

	// Cookies Cookie列表
	Cookies []*http.Cookie

	// BasicAuth 基础认证
	BasicAuth *BasicAuth

	// BearerToken Bearer Token认证
	BearerToken string

	// ContentType 内容类型（自动设置Content-Type头）
	ContentType string

	// FollowRedirects 是否跟随重定向（默认false）
	FollowRedirects bool

	// MaxRedirects 最大重定向次数（默认10）
	MaxRedirects int

	// StreamResponse 是否流式读取响应体（默认false）
	// 如果为true：
	//   - 响应体不会被完整读取到内存，需要通过BodyReader访问
	//   - 不会设置context超时（避免读取大文件时中断）
	//   - 超时保护：ResponseHeaderTimeout保护响应头阶段
	//   - 响应体读取超时：由用户传入带超时的ctx控制
	StreamResponse bool
}

// RetryConfig 重试配置
type RetryConfig struct {
	// MaxRetries 最大重试次数
	MaxRetries int

	// RetryDelay 重试延迟时间
	RetryDelay time.Duration

	// RetryOnStatusCodes 在哪些状态码下重试
	RetryOnStatusCodes []int

	// RetryOnErrors 在哪些错误下重试
	RetryOnErrors []error
}

// BasicAuth 基础认证配置
type BasicAuth struct {
	Username string
	Password string
}

// Response HTTP响应对象
// 封装了HTTP响应的基本信息，响应体解析由用户自行处理
//
// 资源管理：
//
//	普通响应（默认）：无需手动关闭，Body已读取完毕并自动关闭
//	流式响应：必须手动调用 resp.Close() 或 resp.BodyReader.Close()
//
// 示例：
//
//	// 普通响应 - 无需关闭
//	resp, _ := client.Get(ctx, url)
//	data := resp.Body  // 直接使用
//
//	// 流式响应 - 必须关闭
//	resp, _ := client.Get(ctx, url, WithStreamResponse(true))
//	defer resp.Close()  // 必须关闭！
//	io.Copy(file, resp.BodyReader)
type Response struct {
	// StatusCode 状态码
	StatusCode int

	// Status 状态文本
	Status string

	// Headers 响应头
	Headers http.Header

	// Body 响应体（字节数组）
	// 普通响应时包含完整响应体
	// 流式响应时为nil，需通过BodyReader读取
	Body []byte

	// BodyReader 响应体读取器（仅流式响应时有效）
	// 重要：使用完毕后必须调用Close()，否则会泄漏TCP连接！
	BodyReader io.ReadCloser

	// ContentLength 内容长度
	ContentLength int64

	// Request 原始请求对象
	Request *http.Request

	// Response 原始响应对象
	Response *http.Response
}

// IsSuccess 检查响应是否成功（状态码2xx）
func (r *Response) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// IsClientError 检查是否为客户端错误（状态码4xx）
func (r *Response) IsClientError() bool {
	return r.StatusCode >= 400 && r.StatusCode < 500
}

// IsServerError 检查是否为服务器错误（状态码5xx）
func (r *Response) IsServerError() bool {
	return r.StatusCode >= 500 && r.StatusCode < 600
}

// GetHeader 获取响应头值
func (r *Response) GetHeader(key string) string {
	return r.Headers.Get(key)
}

// BodyString 获取响应体的字符串形式
//
// 行为说明：
//   - 如果 Body 不为空，直接返回 Body 的字符串形式
//   - 如果 Body 为空但 BodyReader 不为空（流式响应），从 BodyReader 读取并关闭
//   - 如果两者都为空，返回空字符串
//
// 注意：
//   - 从 BodyReader 读取后会自动关闭流，后续无法再次读取
//   - 如果需要多次读取流式响应，请直接使用 BodyReader 并手动管理
func (r *Response) BodyString() string {
	// 优先使用 Body 字段
	if r.Body != nil && len(r.Body) > 0 {
		return string(r.Body)
	}

	// 如果 Body 为空，尝试从 BodyReader 读取（流式响应）
	if r.BodyReader != nil {
		bodyBytes, err := io.ReadAll(r.BodyReader)
		// 读取后立即关闭，避免资源泄漏
		_ = r.BodyReader.Close()
		r.BodyReader = nil // 标记为已关闭

		if err != nil {
			// 读取失败时返回错误信息（虽然方法签名是 string，但可以包含错误提示）
			return ""
		}

		// 将读取的内容保存到 Body 字段，以便后续使用
		r.Body = bodyBytes
		return string(bodyBytes)
	}

	// 两者都为空
	return ""
}

// Close 关闭响应体（流式响应时需要调用）
func (r *Response) Close() error {
	if r.BodyReader != nil {
		return r.BodyReader.Close()
	}
	return nil
}

// Interceptor 拦截器函数类型
// 在请求发送前和响应返回后执行
type Interceptor func(ctx *InterceptorContext) error

// InterceptorContext 拦截器上下文
type InterceptorContext struct {
	// Request 请求对象
	Request *http.Request

	// Response 响应对象（响应拦截器时可用）
	Response *Response

	// Error 错误（响应拦截器时可用）
	Error error

	// Context 上下文
	Context context.Context
}

// ClientConfig 客户端配置
type ClientConfig struct {
	// Timeout 请求总超时时间（默认30秒，0表示不设置）
	// 通过context控制，可被请求级别的WithTimeout覆盖
	Timeout time.Duration

	// DialTimeout TCP连接超时（默认10秒）
	// 仅控制建立TCP连接的时间
	DialTimeout time.Duration

	// TLSHandshakeTimeout TLS握手超时（默认10秒）
	TLSHandshakeTimeout time.Duration

	// ResponseHeaderTimeout 等待响应头超时（默认30秒）
	// 从发送完请求到收到响应头的时间
	// 对流式响应友好：不包含读取响应体的时间
	ResponseHeaderTimeout time.Duration

	// ExpectContinueTimeout 100-continue超时（默认1秒，通常不需要修改）
	// 作用：上传大文件时，先发送请求头询问服务器是否接受
	//   - 服务器返回 100 Continue → 继续发送文件
	//   - 服务器返回 4xx/5xx → 不发送文件，节省带宽
	//   - 超时 → 假设服务器不支持，直接发送
	ExpectContinueTimeout time.Duration

	// DefaultHeaders 默认请求头
	DefaultHeaders map[string]string

	// Transport HTTP传输配置
	// 如果设置，上面的超时配置将被忽略
	// 如果为nil，将使用默认Transport并应用超时和连接池配置
	Transport *http.Transport

	// CheckRedirect 重定向检查函数
	CheckRedirect func(req *http.Request, via []*http.Request) error

	// MaxRedirects 最大重定向次数（默认10）
	MaxRedirects int

	// FollowRedirects 是否跟随重定向（默认false，不跟随）
	FollowRedirects bool

	// Retry 默认重试配置
	Retry *RetryConfig

	// 连接池配置（仅在Transport为nil时生效）

	// MaxIdleConns 最大空闲连接数（默认100）
	MaxIdleConns int

	// MaxIdleConnsPerHost 每个主机的最大空闲连接数（默认10）
	MaxIdleConnsPerHost int

	// IdleConnTimeout 空闲连接超时时间（默认90秒）
	IdleConnTimeout time.Duration

	// MaxConnsPerHost 每个主机的最大连接数（0表示无限制）
	MaxConnsPerHost int
}

// Validate 验证客户端配置
func (c *ClientConfig) Validate() error {
	if c.Timeout < 0 {
		return ErrInvalidTimeout
	}
	if c.MaxRedirects < 0 {
		return ErrInvalidMaxRedirects
	}
	if c.MaxIdleConns < 0 {
		return ErrInvalidMaxIdleConns
	}
	if c.MaxIdleConnsPerHost < 0 {
		return ErrInvalidMaxIdleConnsPerHost
	}
	if c.MaxConnsPerHost < 0 {
		return ErrInvalidMaxConnsPerHost
	}
	return nil
}

// RequestOption 辅助函数

// WithHeader 设置请求头
func WithHeader(key, value string) RequestOption {
	return func(cfg *RequestConfig) {
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string)
		}
		cfg.Headers[key] = value
	}
}

// WithHeaders 批量设置请求头
func WithHeaders(headers map[string]string) RequestOption {
	return func(cfg *RequestConfig) {
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string)
		}
		for k, v := range headers {
			cfg.Headers[k] = v
		}
	}
}

// WithQueryParam 设置查询参数
func WithQueryParam(key string, value interface{}) RequestOption {
	return func(cfg *RequestConfig) {
		if cfg.QueryParams == nil {
			cfg.QueryParams = make(map[string]interface{})
		}
		cfg.QueryParams[key] = value
	}
}

// WithQueryParams 批量设置查询参数
func WithQueryParams(params map[string]interface{}) RequestOption {
	return func(cfg *RequestConfig) {
		if cfg.QueryParams == nil {
			cfg.QueryParams = make(map[string]interface{})
		}
		for k, v := range params {
			cfg.QueryParams[k] = v
		}
	}
}

// WithTimeout 设置请求超时
func WithTimeout(timeout time.Duration) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Timeout = timeout
	}
}

// WithRetry 设置重试配置
func WithRetry(retry *RetryConfig) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Retry = retry
	}
}

// WithCookie 添加Cookie
func WithCookie(cookie *http.Cookie) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Cookies = append(cfg.Cookies, cookie)
	}
}

// WithCookies 批量添加Cookie
func WithCookies(cookies []*http.Cookie) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.Cookies = append(cfg.Cookies, cookies...)
	}
}

// WithBasicAuth 设置基础认证
func WithBasicAuth(username, password string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.BasicAuth = &BasicAuth{
			Username: username,
			Password: password,
		}
	}
}

// WithBearerToken 设置Bearer Token
func WithBearerToken(token string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.BearerToken = token
	}
}

// WithContentType 设置内容类型
func WithContentType(contentType string) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.ContentType = contentType
	}
}

// WithFollowRedirects 设置是否跟随重定向
func WithFollowRedirects(follow bool) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.FollowRedirects = follow
	}
}

// WithMaxRedirects 设置最大重定向次数
func WithMaxRedirects(max int) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.MaxRedirects = max
	}
}

// WithStreamResponse 设置是否流式读取响应体
// 适用于大文件下载等场景，避免将整个响应体加载到内存
func WithStreamResponse(stream bool) RequestOption {
	return func(cfg *RequestConfig) {
		cfg.StreamResponse = stream
	}
}

// BodyReader 请求体读取器接口
// 用于将不同类型的请求体转换为io.Reader
type BodyReader interface {
	Reader() (io.Reader, error)
	ContentType() string
}

// ========================================
// 请求体类型说明（对应Postman的传输类型）
// ========================================
//
// 1. form-data (multipart/form-data)
//    使用 *FormData 类型，支持文件上传和表单字段混合
//    示例：client.Post(ctx, url, &FormData{...})
//
// 2. x-www-form-urlencoded
//    使用 url.Values 或 *FormURLEncoded 类型
//    示例：client.Post(ctx, url, &FormURLEncoded{...})
//
// 3. raw (JSON/XML/Text/HTML)
//    - JSON: 传入struct/map，自动序列化
//    - XML/Text/HTML: 传入string或[]byte，配合WithContentType
//    示例：client.Post(ctx, url, map[string]string{"key": "value"})
//
// 4. binary
//    传入io.Reader（如*os.File）
//    示例：client.Post(ctx, url, file, WithContentType("application/octet-stream"))
//
// 5. GraphQL
//    传入GraphQL查询结构体，自动序列化为JSON
//    示例：client.Post(ctx, url, &GraphQLRequest{...})
//
// ========================================
// 资源管理说明（重要）
// ========================================
//
// 需要手动关闭资源的场景：
//
// 1. 流式响应（WithStreamResponse(true)或使用DoStream）
//    必须关闭：resp.Close() 或 resp.BodyReader.Close()
//    示例：
//      resp, _ := client.Get(ctx, url, WithStreamResponse(true))
//      defer resp.Close()  // 必须关闭！
//
// 2. FormData使用文件路径时
//    FormData.Close() 会关闭通过路径打开的文件
//    示例：
//      formData := &FormData{}
//      formData.AddFile("file", "/path/to/file")
//      defer formData.Close()  // 关闭打开的文件
//      client.Post(ctx, url, formData)
//
// 不需要手动关闭的场景：
//
// 1. 普通响应（默认非流式）
//    响应体已自动读取并关闭，直接使用resp.Body即可
//
// 2. FormData使用io.Reader时
//    由调用者管理Reader的生命周期
//
// ========================================

// FormFile 表单文件
type FormFile struct {
	// FieldName 表单字段名
	FieldName string
	// FileName 文件名
	FileName string
	// Reader 文件内容读取器（支持大文件流式读取）
	Reader io.Reader
	// FilePath 文件路径（与Reader二选一，优先使用Reader）
	FilePath string
}

// FormData multipart/form-data 表单数据
// 支持大文件流式上传
//
// 资源管理：如果使用文件路径添加文件，请在请求完成后调用Close()关闭打开的文件
type FormData struct {
	// Fields 表单字段（键值对）
	Fields map[string]string
	// Files 表单文件列表
	Files []FormFile
	// openedFiles 内部打开的文件，需要关闭
	openedFiles []io.Closer
}

// AddField 添加表单字段
func (f *FormData) AddField(key, value string) *FormData {
	if f.Fields == nil {
		f.Fields = make(map[string]string)
	}
	f.Fields[key] = value
	return f
}

// AddFile 通过文件路径添加文件
// 注意：使用此方法后，必须在请求完成后调用FormData.Close()释放资源
func (f *FormData) AddFile(fieldName, filePath string) *FormData {
	f.Files = append(f.Files, FormFile{
		FieldName: fieldName,
		FilePath:  filePath,
	})
	return f
}

// AddFileReader 通过Reader添加文件
// 注意：Reader的生命周期由调用者管理
func (f *FormData) AddFileReader(fieldName, fileName string, reader io.Reader) *FormData {
	f.Files = append(f.Files, FormFile{
		FieldName: fieldName,
		FileName:  fileName,
		Reader:    reader,
	})
	return f
}

// Close 关闭FormData打开的所有文件资源
// 仅当使用AddFile(通过路径添加文件)时需要调用
func (f *FormData) Close() error {
	var lastErr error
	for _, closer := range f.openedFiles {
		if err := closer.Close(); err != nil {
			lastErr = err
		}
	}
	f.openedFiles = nil
	return lastErr
}

// FormURLEncoded application/x-www-form-urlencoded 表单数据
// 用于简单表单提交，不支持文件上传
type FormURLEncoded struct {
	// Fields 表单字段
	Fields map[string]string
}

// Add 添加字段
func (f *FormURLEncoded) Add(key, value string) *FormURLEncoded {
	if f.Fields == nil {
		f.Fields = make(map[string]string)
	}
	f.Fields[key] = value
	return f
}

// GraphQLRequest GraphQL请求体
type GraphQLRequest struct {
	// Query GraphQL查询语句
	Query string `json:"query"`
	// Variables 查询变量
	Variables map[string]interface{} `json:"variables,omitempty"`
	// OperationName 操作名称
	OperationName string `json:"operationName,omitempty"`
}
