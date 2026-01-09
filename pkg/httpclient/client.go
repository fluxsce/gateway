// Package httpclient 提供统一的HTTP客户端接口和实现
// 支持多种HTTP方法、请求配置、响应处理等功能
package httpclient

import (
	"context"
	"net/http"
	"time"
)

// Client HTTP客户端接口
// 定义了所有HTTP客户端实现必须支持的基本操作
type Client interface {
	// 基本请求方法

	// Get 发送GET请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Get(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Post 发送POST请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - body: 请求体，可以是io.Reader、[]byte、string或可序列化为JSON的对象
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Post(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error)

	// Put 发送PUT请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - body: 请求体，可以是io.Reader、[]byte、string或可序列化为JSON的对象
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Put(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error)

	// Delete 发送DELETE请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Delete(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Patch 发送PATCH请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - body: 请求体，可以是io.Reader、[]byte、string或可序列化为JSON的对象
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Patch(ctx context.Context, url string, body interface{}, opts ...RequestOption) (*Response, error)

	// Head 发送HEAD请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Head(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Options 发送OPTIONS请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - url: 请求URL
	//   - opts: 请求选项，可设置请求头、查询参数等
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Options(ctx context.Context, url string, opts ...RequestOption) (*Response, error)

	// Do 发送自定义HTTP请求
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - req: HTTP请求对象
	// 返回:
	//   - *Response: HTTP响应对象
	//   - error: 可能的错误
	Do(ctx context.Context, req *http.Request) (*Response, error)

	// DoStream 发送自定义HTTP请求（流式响应）
	// 参数:
	//   - ctx: 上下文，用于控制请求超时和取消
	//   - req: HTTP请求对象
	// 返回:
	//   - *Response: HTTP响应对象，通过BodyReader访问响应体
	//   - error: 可能的错误
	// 注意：调用者负责关闭BodyReader
	DoStream(ctx context.Context, req *http.Request) (*Response, error)

	// 配置方法

	// SetDefaultHeader 设置默认请求头（线程安全）
	// 参数:
	//   - key: 请求头键
	//   - value: 请求头值
	SetDefaultHeader(key, value string)

	// SetTimeout 设置默认超时时间（线程安全）
	// 参数:
	//   - timeout: 超时时间
	SetTimeout(timeout time.Duration)

	// AddInterceptor 添加请求/响应拦截器（线程安全）
	// 参数:
	//   - interceptor: 拦截器函数
	AddInterceptor(interceptor Interceptor)

	// Close 关闭客户端（释放资源）
	Close() error
}
