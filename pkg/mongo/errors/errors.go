// Package errors 提供MongoDB操作的错误定义和处理功能
//
// 此包定义了MongoDB操作中可能遇到的所有错误类型，包括：
// - 预定义的标准错误类型
// - 自定义的MongoDB错误结构体
// - 错误代码常量定义
// - 便捷的错误创建和检查函数
//
// 错误代码范围：2001-3000
// - 连接相关错误（2001-2100）
// - 操作相关错误（2101-2200）
// - 数据相关错误（2201-2300）
// - 事务相关错误（2301-2400）
// - 索引相关错误（2401-2500）
// - 权限相关错误（2501-2600）
package errors

import (
	"errors"
	"fmt"
)

// === 预定义的标准错误类型 ===

var (
	// ErrDocumentNotFound 文档未找到错误
	// 当查询操作没有找到匹配的文档时返回此错误
	ErrDocumentNotFound = errors.New("document not found")

	// ErrDuplicateKey 重复键错误
	// 当插入或更新操作违反唯一约束时返回此错误
	ErrDuplicateKey = errors.New("duplicate key error")

	// ErrConnection MongoDB连接错误
	// 当无法建立或维持MongoDB连接时返回此错误
	ErrConnection = errors.New("mongodb connection error")

	// ErrInvalidObjectID 无效的ObjectID错误
	// 当提供的ObjectID格式不正确时返回此错误
	ErrInvalidObjectID = errors.New("invalid ObjectID")

	// ErrInvalidFilter 无效的过滤器错误
	// 当查询过滤器格式不正确时返回此错误
	ErrInvalidFilter = errors.New("invalid filter")

	// ErrInvalidDocument 无效的文档错误
	// 当文档结构不符合要求时返回此错误
	ErrInvalidDocument = errors.New("invalid document")

	// ErrTransactionFailed 事务失败错误
	// 当事务执行失败时返回此错误
	ErrTransactionFailed = errors.New("transaction failed")

	// ErrSessionNotFound 会话未找到错误
	// 当指定的会话不存在时返回此错误
	ErrSessionNotFound = errors.New("session not found")

	// ErrCollectionNotFound 集合未找到错误
	// 当指定的集合不存在时返回此错误
	ErrCollectionNotFound = errors.New("collection not found")

	// ErrIndexAlreadyExists 索引已存在错误
	// 当尝试创建已存在的索引时返回此错误
	ErrIndexAlreadyExists = errors.New("index already exists")

	// ErrInvalidAggregation 无效的聚合操作错误
	// 当聚合管道配置不正确时返回此错误
	ErrInvalidAggregation = errors.New("invalid aggregation pipeline")

	// ErrOperationTimeout 操作超时错误
	// 当操作执行时间超过设定的超时时间时返回此错误
	ErrOperationTimeout = errors.New("operation timeout")

	// ErrWriteConcernFailed 写关注失败错误
	// 当写操作无法满足指定的写关注级别时返回此错误
	ErrWriteConcernFailed = errors.New("write concern failed")

	// ErrReadConcernFailed 读关注失败错误
	// 当读操作无法满足指定的读关注级别时返回此错误
	ErrReadConcernFailed = errors.New("read concern failed")
)

// === MongoDB自定义错误类型 ===

// MongoError MongoDB自定义错误结构体
// 提供更详细的错误信息，包括操作类型、错误代码、错误消息和原始错误
type MongoError struct {
	Operation string // 操作名称，如 "insert", "update", "find" 等
	Code      int    // 错误代码，用于程序化错误处理
	Message   string // 错误消息，用于日志记录和用户提示
	Cause     error  // 原始错误，保留底层错误信息
}

// Error 实现error接口
// 返回格式化的错误信息字符串
func (e *MongoError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("mongo %s error [%d]: %s, caused by: %v", e.Operation, e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("mongo %s error [%d]: %s", e.Operation, e.Code, e.Message)
}

// Unwrap 返回原始错误
// 支持Go 1.13+的errors.Unwrap功能，用于错误链处理
func (e *MongoError) Unwrap() error {
	return e.Cause
}

// Is 支持errors.Is检查
// 允许使用errors.Is进行错误类型比较
func (e *MongoError) Is(target error) bool {
	if mongoErr, ok := target.(*MongoError); ok {
		return e.Code == mongoErr.Code
	}
	return false
}

// NewMongoError 创建新的MongoDB错误
//
// 参数：
//
//	operation: 操作名称
//	code: 错误代码
//	message: 错误消息
//	cause: 原始错误（可为nil）
func NewMongoError(operation string, code int, message string, cause error) *MongoError {
	return &MongoError{
		Operation: operation,
		Code:      code,
		Message:   message,
		Cause:     cause,
	}
}

// === MongoDB错误代码常量（范围：2001-3000） ===

const (
	// === 连接相关错误 (2001-2100) ===

	// ErrCodeConnection 通用连接错误
	ErrCodeConnection = 2001

	// ErrCodeAuthentication 认证失败错误
	ErrCodeAuthentication = 2002

	// ErrCodeNetworkTimeout 网络超时错误
	ErrCodeNetworkTimeout = 2003

	// ErrCodeDNSResolution DNS解析错误
	ErrCodeDNSResolution = 2004

	// === 操作相关错误 (2101-2200) ===

	// ErrCodeInvalidQuery 无效查询错误
	ErrCodeInvalidQuery = 2101

	// ErrCodeInvalidUpdate 无效更新错误
	ErrCodeInvalidUpdate = 2102

	// ErrCodeInvalidInsert 无效插入错误
	ErrCodeInvalidInsert = 2103

	// ErrCodeInvalidDelete 无效删除错误
	ErrCodeInvalidDelete = 2104

	// ErrCodeInvalidAggregation 无效聚合错误
	ErrCodeInvalidAggregation = 2105

	// === 数据相关错误 (2201-2300) ===

	// ErrCodeDocumentNotFound 文档未找到错误
	ErrCodeDocumentNotFound = 2201

	// ErrCodeDuplicateKey 重复键错误
	ErrCodeDuplicateKey = 2202

	// ErrCodeInvalidObjectID 无效ObjectID错误
	ErrCodeInvalidObjectID = 2203

	// ErrCodeValidationFailed 验证失败错误
	ErrCodeValidationFailed = 2204

	// ErrCodeSchemaViolation 模式违反错误
	ErrCodeSchemaViolation = 2205

	// === 事务相关错误 (2301-2400) ===

	// ErrCodeTransactionAborted 事务中止错误
	ErrCodeTransactionAborted = 2301

	// ErrCodeTransactionTimeout 事务超时错误
	ErrCodeTransactionTimeout = 2302

	// ErrCodeWriteConflict 写冲突错误
	ErrCodeWriteConflict = 2303

	// === 索引相关错误 (2401-2500) ===

	// ErrCodeIndexNotFound 索引未找到错误
	ErrCodeIndexNotFound = 2401

	// ErrCodeIndexCreationFailed 索引创建失败错误
	ErrCodeIndexCreationFailed = 2402

	// ErrCodeIndexDropFailed 索引删除失败错误
	ErrCodeIndexDropFailed = 2403

	// === 权限相关错误 (2501-2600) ===

	// ErrCodeUnauthorized 未授权错误
	ErrCodeUnauthorized = 2501

	// ErrCodeAccessDenied 访问拒绝错误
	ErrCodeAccessDenied = 2502

	// ErrCodeInsufficientPrivileges 权限不足错误
	ErrCodeInsufficientPrivileges = 2503
)

// === 便捷的错误创建函数 ===

// NewConnectionError 创建连接错误
// 用于连接相关的错误场景
func NewConnectionError(message string, cause error) *MongoError {
	return NewMongoError("connection", ErrCodeConnection, message, cause)
}

// NewQueryError 创建查询错误
// 用于查询操作相关的错误场景
func NewQueryError(message string, cause error) *MongoError {
	return NewMongoError("query", ErrCodeInvalidQuery, message, cause)
}

// NewUpdateError 创建更新错误
// 用于更新操作相关的错误场景
func NewUpdateError(message string, cause error) *MongoError {
	return NewMongoError("update", ErrCodeInvalidUpdate, message, cause)
}

// NewInsertError 创建插入错误
// 用于插入操作相关的错误场景
func NewInsertError(message string, cause error) *MongoError {
	return NewMongoError("insert", ErrCodeInvalidInsert, message, cause)
}

// NewDeleteError 创建删除错误
// 用于删除操作相关的错误场景
func NewDeleteError(message string, cause error) *MongoError {
	return NewMongoError("delete", ErrCodeInvalidDelete, message, cause)
}

// NewTransactionError 创建事务错误
// 用于事务操作相关的错误场景
func NewTransactionError(message string, cause error) *MongoError {
	return NewMongoError("transaction", ErrCodeTransactionAborted, message, cause)
}

// NewIndexError 创建索引错误
// 用于索引操作相关的错误场景
func NewIndexError(message string, cause error) *MongoError {
	return NewMongoError("index", ErrCodeIndexCreationFailed, message, cause)
}

// NewValidationError 创建验证错误
// 用于数据验证相关的错误场景
func NewValidationError(message string, cause error) *MongoError {
	return NewMongoError("validation", ErrCodeValidationFailed, message, cause)
}

// === 错误检查函数 ===

// IsDocumentNotFound 检查是否为文档未找到错误
// 支持检查自定义错误和标准错误
func IsDocumentNotFound(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeDocumentNotFound
	}
	return errors.Is(err, ErrDocumentNotFound)
}

// IsDuplicateKey 检查是否为重复键错误
// 支持检查自定义错误和标准错误
func IsDuplicateKey(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeDuplicateKey
	}
	return errors.Is(err, ErrDuplicateKey)
}

// IsConnectionError 检查是否为连接错误
// 包括连接、认证、网络超时和DNS解析错误
func IsConnectionError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeConnection ||
			mongoErr.Code == ErrCodeAuthentication ||
			mongoErr.Code == ErrCodeNetworkTimeout ||
			mongoErr.Code == ErrCodeDNSResolution
	}
	return errors.Is(err, ErrConnection)
}

// IsTransactionError 检查是否为事务错误
// 包括事务中止、超时和写冲突错误
func IsTransactionError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code >= ErrCodeTransactionAborted && mongoErr.Code <= ErrCodeWriteConflict
	}
	return errors.Is(err, ErrTransactionFailed)
}

// IsTimeoutError 检查是否为超时错误
// 包括网络超时和事务超时错误
func IsTimeoutError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeNetworkTimeout ||
			mongoErr.Code == ErrCodeTransactionTimeout
	}
	return errors.Is(err, ErrOperationTimeout)
}

// IsIndexError 检查是否为索引相关错误
// 包括索引未找到、创建失败和删除失败错误
func IsIndexError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code >= ErrCodeIndexNotFound && mongoErr.Code <= ErrCodeIndexDropFailed
	}
	return errors.Is(err, ErrIndexAlreadyExists)
}

// IsAuthenticationError 检查是否为认证错误
// 包括认证失败、未授权和权限不足错误
func IsAuthenticationError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeAuthentication ||
			mongoErr.Code == ErrCodeUnauthorized ||
			mongoErr.Code == ErrCodeAccessDenied ||
			mongoErr.Code == ErrCodeInsufficientPrivileges
	}
	return false
}

// IsValidationError 检查是否为验证错误
// 包括数据验证失败和模式违反错误
func IsValidationError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		return mongoErr.Code == ErrCodeValidationFailed ||
			mongoErr.Code == ErrCodeSchemaViolation
	}
	return false
}

// === 错误分类函数 ===

// GetErrorCategory 获取错误分类
// 根据错误代码返回错误所属的分类
func GetErrorCategory(err error) string {
	if mongoErr, ok := err.(*MongoError); ok {
		switch {
		case mongoErr.Code >= 2001 && mongoErr.Code <= 2100:
			return "connection"
		case mongoErr.Code >= 2101 && mongoErr.Code <= 2200:
			return "operation"
		case mongoErr.Code >= 2201 && mongoErr.Code <= 2300:
			return "data"
		case mongoErr.Code >= 2301 && mongoErr.Code <= 2400:
			return "transaction"
		case mongoErr.Code >= 2401 && mongoErr.Code <= 2500:
			return "index"
		case mongoErr.Code >= 2501 && mongoErr.Code <= 2600:
			return "permission"
		default:
			return "unknown"
		}
	}
	return "standard"
}

// IsRetryableError 检查错误是否可重试
// 某些类型的错误（如网络超时）可以通过重试解决
func IsRetryableError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		switch mongoErr.Code {
		case ErrCodeNetworkTimeout, ErrCodeConnection, ErrCodeWriteConflict:
			return true
		default:
			return false
		}
	}
	return false
}

// IsFatalError 检查错误是否为致命错误
// 致命错误通常需要人工干预，不应该自动重试
func IsFatalError(err error) bool {
	if mongoErr, ok := err.(*MongoError); ok {
		switch mongoErr.Code {
		case ErrCodeAuthentication, ErrCodeUnauthorized, ErrCodeAccessDenied,
			ErrCodeInsufficientPrivileges, ErrCodeValidationFailed, ErrCodeSchemaViolation:
			return true
		default:
			return false
		}
	}
	return false
}
