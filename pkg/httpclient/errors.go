// Package errors 提供HTTP客户端操作的错误定义和处理功能
//
// 错误代码范围：1001-2000
// - 配置相关错误 (1001-1100)
// - 请求相关错误 (1101-1200)
// - 响应相关错误 (1201-1300)
// - 网络相关错误 (1301-1400)
package httpclient

import "fmt"

// === HTTP错误代码常量（仅保留实际使用的，范围：1001-2000） ===

const (
	// === 配置相关错误 (1001-1100) ===

	// ErrCodeInvalidURL 无效URL错误代码
	ErrCodeInvalidURL = 1001

	// ErrCodeInvalidTimeout 无效超时时间错误代码
	ErrCodeInvalidTimeout = 1002

	// ErrCodeInvalidMaxRedirects 无效最大重定向次数错误代码
	ErrCodeInvalidMaxRedirects = 1003

	// ErrCodeInvalidMaxIdleConns 无效最大空闲连接数错误代码
	ErrCodeInvalidMaxIdleConns = 1004

	// ErrCodeInvalidMaxIdleConnsPerHost 无效每个主机最大空闲连接数错误代码
	ErrCodeInvalidMaxIdleConnsPerHost = 1005

	// ErrCodeInvalidMaxConnsPerHost 无效每个主机最大连接数错误代码
	ErrCodeInvalidMaxConnsPerHost = 1006

	// === 请求相关错误 (1101-1200) ===

	// ErrCodeRequestFailed 请求失败错误代码
	ErrCodeRequestFailed = 1101

	// ErrCodeRequestTimeout 请求超时错误代码
	ErrCodeRequestTimeout = 1102

	// ErrCodeCreateRequestFailed 创建请求失败错误代码
	ErrCodeCreateRequestFailed = 1103

	// ErrCodeMarshalBodyFailed 序列化请求体失败错误代码
	ErrCodeMarshalBodyFailed = 1104

	// ErrCodeMaxRedirectsExceeded 超过最大重定向次数错误代码
	ErrCodeMaxRedirectsExceeded = 1105

	// ErrCodeRetryExhausted 重试耗尽错误代码
	ErrCodeRetryExhausted = 1106

	// === 响应相关错误 (1201-1300) ===

	// ErrCodeResponseReadFailed 读取响应失败错误代码
	ErrCodeResponseReadFailed = 1201

	// === 网络相关错误 (1301-1400) ===

	// ErrCodeNetworkError 网络错误代码
	ErrCodeNetworkError = 1301
)

// === 预定义的标准错误类型（仅保留实际使用的） ===

var (
	// ErrInvalidURL URL无效错误
	ErrInvalidURL = &HTTPError{
		Code:    ErrCodeInvalidURL,
		Message: "invalid URL",
	}

	// ErrInvalidTimeout 超时时间无效错误
	ErrInvalidTimeout = &HTTPError{
		Code:    ErrCodeInvalidTimeout,
		Message: "invalid timeout duration",
	}

	// ErrInvalidMaxRedirects 最大重定向次数无效错误
	ErrInvalidMaxRedirects = &HTTPError{
		Code:    ErrCodeInvalidMaxRedirects,
		Message: "invalid max redirects",
	}

	// ErrMaxRedirectsExceeded 超过最大重定向次数错误
	ErrMaxRedirectsExceeded = &HTTPError{
		Code:    ErrCodeMaxRedirectsExceeded,
		Message: "max redirects exceeded",
	}

	// ErrResponseReadFailed 读取响应失败错误
	ErrResponseReadFailed = &HTTPError{
		Code:    ErrCodeResponseReadFailed,
		Message: "failed to read response body",
	}

	// ErrInvalidMaxIdleConns 最大空闲连接数无效错误
	ErrInvalidMaxIdleConns = &HTTPError{
		Code:    ErrCodeInvalidMaxIdleConns,
		Message: "invalid max idle connections",
	}

	// ErrInvalidMaxIdleConnsPerHost 每个主机的最大空闲连接数无效错误
	ErrInvalidMaxIdleConnsPerHost = &HTTPError{
		Code:    ErrCodeInvalidMaxIdleConnsPerHost,
		Message: "invalid max idle connections per host",
	}

	// ErrInvalidMaxConnsPerHost 每个主机的最大连接数无效错误
	ErrInvalidMaxConnsPerHost = &HTTPError{
		Code:    ErrCodeInvalidMaxConnsPerHost,
		Message: "invalid max connections per host",
	}
)

// HTTPError HTTP客户端错误结构体
// 提供更详细的错误信息，包括错误代码、错误消息、HTTP状态码、URL和原始错误
type HTTPError struct {
	// Code 错误代码（使用ErrCodeXXX常量）
	// 用于程序化错误处理和错误分类
	Code int

	// Message 错误消息
	// 用于日志记录和用户提示
	Message string

	// StatusCode HTTP状态码（如果有）
	// 当错误来自HTTP响应时，包含响应的状态码
	StatusCode int

	// URL 请求URL（如果有）
	// 当错误与特定请求相关时，包含请求的URL
	URL string

	// Err 原始错误
	// 保留底层错误信息，支持错误链处理
	Err error
}

// Error 实现error接口
// 返回格式化的错误信息字符串
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("http client error [%d]: %s: %v", e.Code, e.Message, e.Err)
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("http client error [%d]: %s (status: %d, url: %s)", e.Code, e.Message, e.StatusCode, e.URL)
	}
	if e.URL != "" {
		return fmt.Sprintf("http client error [%d]: %s (url: %s)", e.Code, e.Message, e.URL)
	}
	return fmt.Sprintf("http client error [%d]: %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError 创建新的HTTP错误
//
// 参数:
//   - code: 错误代码（使用ErrCodeXXX常量）
//   - message: 错误消息
//   - err: 原始错误（可为nil）
func NewHTTPError(code int, message string, err error) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

// NewHTTPErrorWithStatus 创建带状态码的HTTP错误
//
// 参数:
//   - code: 错误代码（使用ErrCodeXXX常量）
//   - message: 错误消息
//   - statusCode: HTTP状态码
//   - url: 请求URL
func NewHTTPErrorWithStatus(code int, message string, statusCode int, url string) *HTTPError {
	return &HTTPError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		URL:        url,
	}
}

// IsHTTPError 检查错误是否为HTTPError
// 支持检查自定义错误和标准错误
func IsHTTPError(err error) bool {
	_, ok := err.(*HTTPError)
	return ok
}

// AsHTTPError 将错误转换为HTTPError
// 如果错误是HTTPError类型，返回该错误和true；否则返回nil和false
func AsHTTPError(err error) (*HTTPError, bool) {
	httpErr, ok := err.(*HTTPError)
	return httpErr, ok
}

// === 错误检查函数 ===

// IsTimeoutError 检查是否为超时错误
func IsTimeoutError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code == ErrCodeRequestTimeout
	}
	return false
}

// IsNetworkError 检查是否为网络错误
func IsNetworkError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		return httpErr.Code == ErrCodeNetworkError
	}
	return false
}

// IsRetryableError 检查错误是否可重试
// 某些类型的错误（如网络超时、5xx状态码）可以通过重试解决
func IsRetryableError(err error) bool {
	if httpErr, ok := err.(*HTTPError); ok {
		if httpErr.Code == ErrCodeRequestTimeout || httpErr.Code == ErrCodeNetworkError ||
			httpErr.Code == ErrCodeResponseReadFailed {
			return true
		}
		// 5xx状态码通常可重试
		if httpErr.StatusCode >= 500 && httpErr.StatusCode < 600 {
			return true
		}
		return false
	}
	return false
}
