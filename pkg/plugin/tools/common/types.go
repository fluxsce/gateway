// Package common 提供工具包的公共类型和函数
// 定义了各种工具共享的数据结构和接口
package common

import (
	"os"
	"time"
)

// FileInfo 文件信息结构
// 包含文件或目录的详细信息
type FileInfo struct {
	// 文件名
	Name string `json:"name"`
	
	// 完整路径
	Path string `json:"path"`
	
	// 文件大小（字节）
	Size int64 `json:"size"`
	
	// 是否为目录
	IsDir bool `json:"is_dir"`
	
	// 修改时间
	ModTime time.Time `json:"mod_time"`
	
	// 文件权限
	Mode os.FileMode `json:"mode"`
	
	// 文件所有者
	Owner string `json:"owner,omitempty"`
	
	// 文件组
	Group string `json:"group,omitempty"`
	
	// 额外属性
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// TransferType 传输类型枚举
type TransferType int

const (
	// TransferTypeUpload 上传操作
	TransferTypeUpload TransferType = iota + 1
	
	// TransferTypeDownload 下载操作
	TransferTypeDownload
)

// String 返回传输类型的字符串表示
func (t TransferType) String() string {
	switch t {
	case TransferTypeUpload:
		return "upload"
	case TransferTypeDownload:
		return "download"
	default:
		return "unknown"
	}
}

// TransferProgress 传输进度信息
type TransferProgress struct {
	// 操作ID
	OperationID string `json:"operation_id"`
	
	// 文件路径
	FilePath string `json:"file_path"`
	
	// 总字节数
	TotalBytes int64 `json:"total_bytes"`
	
	// 已传输字节数
	TransferredBytes int64 `json:"transferred_bytes"`
	
	// 传输进度百分比
	Percentage float64 `json:"percentage"`
	
	// 传输速度（字节/秒）
	Speed float64 `json:"speed"`
	
	// 剩余时间估计
	ETA time.Duration `json:"eta"`
	
	// 开始时间
	StartTime time.Time `json:"start_time"`
	
	// 当前时间
	CurrentTime time.Time `json:"current_time"`
}

// TransferResult 单个传输操作结果
// 包含传输操作的详细结果信息
type TransferResult struct {
	// 操作ID
	OperationID string `json:"operation_id"`
	
	// 传输类型
	Type TransferType `json:"type"`
	
	// 本地路径
	LocalPath string `json:"local_path"`
	
	// 远程路径
	RemotePath string `json:"remote_path"`
	
	// 传输的字节数
	BytesTransferred int64 `json:"bytes_transferred"`
	
	// 传输开始时间
	StartTime time.Time `json:"start_time"`
	
	// 传输结束时间
	EndTime time.Time `json:"end_time"`
	
	// 传输耗时
	Duration time.Duration `json:"duration"`
	
	// 平均传输速度（字节/秒）
	AverageSpeed float64 `json:"average_speed"`
	
	// 是否成功
	Success bool `json:"success"`
	
	// 错误信息（如果失败）
	Error string `json:"error,omitempty"`
	
	// 额外信息
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// BatchTransferResult 批量传输结果
// 包含多个传输操作的汇总结果信息
type BatchTransferResult struct {
	// 总操作数
	TotalOperations int `json:"total_operations"`
	
	// 成功操作数
	SuccessfulOperations int `json:"successful_operations"`
	
	// 失败操作数
	FailedOperations int `json:"failed_operations"`
	
	// 跳过操作数
	SkippedOperations int `json:"skipped_operations"`
	
	// 总传输字节数
	TotalBytesTransferred int64 `json:"total_bytes_transferred"`
	
	// 总耗时
	TotalDuration time.Duration `json:"total_duration"`
	
	// 平均传输速度
	AverageSpeed float64 `json:"average_speed"`
	
	// 各个操作的详细结果
	Results []*TransferResult `json:"results"`
	
	// 开始时间
	StartTime time.Time `json:"start_time"`
	
	// 结束时间
	EndTime time.Time `json:"end_time"`
	
	// 错误汇总
	Errors []string `json:"errors,omitempty"`
}

// TransferOperation 传输操作定义
// 定义单个文件传输操作的详细信息
type TransferOperation struct {
	// 操作类型：上传或下载
	Type TransferType `json:"type"`
	
	// 本地文件路径
	LocalPath string `json:"local_path"`
	
	// 远程文件路径
	RemotePath string `json:"remote_path"`
	
	// 操作优先级（数值越小优先级越高）
	Priority int `json:"priority"`
	
	// 操作标识符，用于跟踪和管理
	ID string `json:"id"`
	
	// 操作元数据
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ProgressCallback 传输进度回调函数
// 用于监控单个文件传输的进度
type ProgressCallback func(progress *TransferProgress)

// ErrorCallback 错误回调函数
// 用于处理传输过程中的错误
type ErrorCallback func(err *TransferError)

// TransferError 传输错误信息
type TransferError struct {
	// 操作ID
	OperationID string `json:"operation_id"`
	
	// 错误类型
	Type string `json:"type"`
	
	// 错误消息
	Message string `json:"message"`
	
	// 文件路径
	FilePath string `json:"file_path"`
	
	// 原始错误
	OriginalError error `json:"-"`
	
	// 是否可重试
	Retryable bool `json:"retryable"`
	
	// 发生时间
	Timestamp time.Time `json:"timestamp"`
} 