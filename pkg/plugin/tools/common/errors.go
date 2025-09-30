package common

import (
	"fmt"
)

// 定义错误类型常量
const (
	// ErrTypeConnection 连接错误
	ErrTypeConnection = "connection_error"

	// ErrTypeAuthentication 认证错误
	ErrTypeAuthentication = "authentication_error"

	// ErrTypePermission 权限错误
	ErrTypePermission = "permission_error"

	// ErrTypeFileNotFound 文件未找到
	ErrTypeFileNotFound = "file_not_found"

	// ErrTypeFileExists 文件已存在
	ErrTypeFileExists = "file_exists"

	// ErrTypeIO IO错误
	ErrTypeIO = "io_error"

	// ErrTypeTimeout 超时错误
	ErrTypeTimeout = "timeout_error"

	// ErrTypeInvalidArgument 无效参数
	ErrTypeInvalidArgument = "invalid_argument"

	// ErrTypeUnsupported 不支持的操作
	ErrTypeUnsupported = "unsupported_operation"

	// ErrTypeInternal 内部错误
	ErrTypeInternal = "internal_error"

	// ErrTypeNotImplemented 功能未实现
	ErrTypeNotImplemented = "not_implemented"

	// ErrTypeNotConnected 未连接错误
	ErrTypeNotConnected = "not_connected"
)

// ToolError 工具错误接口
// 定义了工具错误的通用行为
type ToolError interface {
	error

	// Type 返回错误类型
	Type() string

	// IsRetryable 返回错误是否可重试
	IsRetryable() bool

	// OriginalError 返回原始错误
	OriginalError() error

	// WithContext 添加上下文信息并返回新错误
	WithContext(key string, value interface{}) ToolError
}

// BaseError 基础错误实现
// 实现了ToolError接口的基本功能
type BaseError struct {
	// 错误类型
	ErrorType string

	// 错误消息
	Message string

	// 原始错误
	Original error

	// 是否可重试
	Retryable bool

	// 上下文信息
	Context map[string]interface{}
}

// Error 实现error接口
func (e *BaseError) Error() string {
	if e.Original != nil {
		return fmt.Sprintf("%s: %s (caused by: %s)", e.ErrorType, e.Message, e.Original.Error())
	}
	return fmt.Sprintf("%s: %s", e.ErrorType, e.Message)
}

// Type 返回错误类型
func (e *BaseError) Type() string {
	return e.ErrorType
}

// IsRetryable 返回错误是否可重试
func (e *BaseError) IsRetryable() bool {
	return e.Retryable
}

// OriginalError 返回原始错误
func (e *BaseError) OriginalError() error {
	return e.Original
}

// WithContext 添加上下文信息并返回新错误
func (e *BaseError) WithContext(key string, value interface{}) ToolError {
	// 创建新的错误对象
	newErr := &BaseError{
		ErrorType: e.ErrorType,
		Message:   e.Message,
		Original:  e.Original,
		Retryable: e.Retryable,
		Context:   make(map[string]interface{}),
	}

	// 复制现有上下文
	for k, v := range e.Context {
		newErr.Context[k] = v
	}

	// 添加新上下文
	newErr.Context[key] = value

	return newErr
}

// NewError 创建新的工具错误
func NewError(errType string, message string, original error) ToolError {
	retryable := false

	// 根据错误类型确定是否可重试
	switch errType {
	case ErrTypeConnection, ErrTypeTimeout:
		retryable = true
	}

	return &BaseError{
		ErrorType: errType,
		Message:   message,
		Original:  original,
		Retryable: retryable,
		Context:   make(map[string]interface{}),
	}
}

// NewConnectionError 创建连接错误
func NewConnectionError(message string, original error) ToolError {
	return NewError(ErrTypeConnection, message, original)
}

// NewAuthenticationError 创建认证错误
func NewAuthenticationError(message string, original error) ToolError {
	return NewError(ErrTypeAuthentication, message, original)
}

// NewPermissionError 创建权限错误
func NewPermissionError(message string, original error) ToolError {
	return NewError(ErrTypePermission, message, original)
}

// NewFileNotFoundError 创建文件未找到错误
func NewFileNotFoundError(path string, original error) ToolError {
	err := NewError(ErrTypeFileNotFound, fmt.Sprintf("file not found: %s", path), original)
	return err.WithContext("path", path)
}

// NewFileExistsError 创建文件已存在错误
func NewFileExistsError(path string, original error) ToolError {
	err := NewError(ErrTypeFileExists, fmt.Sprintf("file already exists: %s", path), original)
	return err.WithContext("path", path)
}

// NewIOError 创建IO错误
func NewIOError(message string, original error) ToolError {
	return NewError(ErrTypeIO, message, original)
}

// NewTimeoutError 创建超时错误
func NewTimeoutError(message string, original error) ToolError {
	return NewError(ErrTypeTimeout, message, original)
}

// NewInvalidArgumentError 创建无效参数错误
func NewInvalidArgumentError(message string) ToolError {
	return NewError(ErrTypeInvalidArgument, message, nil)
}

// NewUnsupportedError 创建不支持的操作错误
func NewUnsupportedError(message string) ToolError {
	return NewError(ErrTypeUnsupported, message, nil)
}

// NewInternalError 创建内部错误
func NewInternalError(message string, original error) ToolError {
	return NewError(ErrTypeInternal, message, original)
}

// NewNotImplementedError 创建功能未实现错误
func NewNotImplementedError(message string) ToolError {
	return NewError(ErrTypeNotImplemented, message, nil)
}

// NewNotConnectedError 创建未连接错误
func NewNotConnectedError(message string) ToolError {
	return NewError(ErrTypeNotConnected, message, nil)
}
