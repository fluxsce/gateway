package types

import "errors"

// 定义日志写入相关的错误
var (
	// ErrUnsupportedWriterType 不支持的日志写入器类型错误
	ErrUnsupportedWriterType = errors.New("unsupported writer type")
	
	// ErrWriterNotInitialized 日志写入器未初始化错误
	ErrWriterNotInitialized = errors.New("writer not initialized")
	
	// ErrInvalidConfig 无效配置错误
	ErrInvalidConfig = errors.New("invalid config")
	
	// ErrDatabaseConnectionFailed 数据库连接失败错误
	ErrDatabaseConnectionFailed = errors.New("database connection failed")
	
	// ErrTransactionFailed 事务操作失败错误
	ErrTransactionFailed = errors.New("transaction failed")
	
	// ErrWriteFailed 写入失败错误
	ErrWriteFailed = errors.New("write failed")
	
	// ErrFlushFailed 刷新失败错误
	ErrFlushFailed = errors.New("flush failed")
	
	// ErrCloseFailed 关闭失败错误
	ErrCloseFailed = errors.New("close failed")
	
	// ErrBufferFull 缓冲区已满错误
	ErrBufferFull = errors.New("buffer full")
	
	// ErrInvalidLogData 无效日志数据错误
	ErrInvalidLogData = errors.New("invalid log data")
	
	// ErrFileOperationFailed 文件操作失败错误
	ErrFileOperationFailed = errors.New("file operation failed")
	
	// ErrNetworkFailed 网络操作失败错误
	ErrNetworkFailed = errors.New("network operation failed")
	
	// ErrTimeoutExceeded 超时错误
	ErrTimeoutExceeded = errors.New("timeout exceeded")
	
	// ErrContextCanceled 上下文取消错误
	ErrContextCanceled = errors.New("context canceled")
) 