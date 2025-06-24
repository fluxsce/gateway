// Package sftp 提供SFTP文件传输功能
// 支持文件上传、下载、目录操作等完整的SFTP操作
package sftp

import (
	"context"
	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
)

// Client SFTP客户端接口
// 定义了完整的SFTP操作功能，支持文件传输、目录管理等操作
type Client interface {
	// ===== 连接管理 =====
	
	// Connect 连接到SFTP服务器
	// 建立与远程SFTP服务器的连接，支持密码和密钥认证
	// 参数:
	//   ctx: 上下文，用于超时控制和取消操作
	// 返回:
	//   error: 连接失败时返回错误信息
	Connect(ctx context.Context) error
	
	// Close 关闭SFTP连接
	// 释放所有相关资源，包括SSH连接和SFTP会话
	// 返回:
	//   error: 关闭失败时返回错误信息
	Close() error
	
	// IsConnected 检查连接状态
	// 返回当前SFTP连接是否处于活跃状态
	// 返回:
	//   bool: true表示已连接，false表示未连接
	IsConnected() bool
	
	// ===== 文件传输操作 =====
	
	// UploadFile 上传单个文件到远程服务器
	// 将本地文件上传到远程SFTP服务器的指定路径
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   localPath: 本地文件路径
	//   remotePath: 远程目标路径
	//   options: 上传选项配置
	// 返回:
	//   *common.TransferResult: 传输结果信息
	//   error: 上传失败时返回错误信息
	UploadFile(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error)
	
	// DownloadFile 从远程服务器下载单个文件
	// 将远程SFTP服务器上的文件下载到本地指定路径
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   remotePath: 远程文件路径
	//   localPath: 本地目标路径
	//   options: 下载选项配置
	// 返回:
	//   *common.TransferResult: 传输结果信息
	//   error: 下载失败时返回错误信息
	DownloadFile(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error)
	
	// UploadDirectory 上传整个目录到远程服务器
	// 递归上传本地目录及其所有子目录和文件到远程服务器
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   localDir: 本地目录路径
	//   remoteDir: 远程目标目录路径
	//   options: 上传选项配置
	// 返回:
	//   *common.BatchTransferResult: 批量传输结果信息
	//   error: 上传失败时返回错误信息
	UploadDirectory(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error)
	
	// DownloadDirectory 从远程服务器下载整个目录
	// 递归下载远程目录及其所有子目录和文件到本地
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   remoteDir: 远程目录路径
	//   localDir: 本地目标目录路径
	//   options: 下载选项配置
	// 返回:
	//   *common.BatchTransferResult: 批量传输结果信息
	//   error: 下载失败时返回错误信息
	DownloadDirectory(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error)
	
	// ===== 目录和文件操作 =====
	
	// ListDirectory 列出远程目录内容
	// 获取远程目录中的文件和子目录列表信息
	// 参数:
	//   ctx: 上下文
	//   remotePath: 远程目录路径
	// 返回:
	//   []*common.FileInfo: 文件信息列表
	//   error: 操作失败时返回错误信息
	ListDirectory(ctx context.Context, remotePath string) ([]*common.FileInfo, error)
	
	// CreateDirectory 在远程服务器创建目录
	// 在远程SFTP服务器上创建指定路径的目录，支持递归创建
	// 参数:
	//   ctx: 上下文
	//   remotePath: 要创建的远程目录路径
	//   recursive: 是否递归创建父目录
	// 返回:
	//   error: 创建失败时返回错误信息
	CreateDirectory(ctx context.Context, remotePath string, recursive bool) error
	
	// RemoveFile 删除远程文件
	// 删除远程SFTP服务器上的指定文件
	// 参数:
	//   ctx: 上下文
	//   remotePath: 要删除的远程文件路径
	// 返回:
	//   error: 删除失败时返回错误信息
	RemoveFile(ctx context.Context, remotePath string) error
	
	// RemoveDirectory 删除远程目录
	// 删除远程SFTP服务器上的指定目录，支持递归删除
	// 参数:
	//   ctx: 上下文
	//   remotePath: 要删除的远程目录路径
	//   recursive: 是否递归删除子目录和文件
	// 返回:
	//   error: 删除失败时返回错误信息
	RemoveDirectory(ctx context.Context, remotePath string, recursive bool) error
	
	// GetFileInfo 获取远程文件信息
	// 获取远程文件或目录的详细信息
	// 参数:
	//   ctx: 上下文
	//   remotePath: 远程文件或目录路径
	// 返回:
	//   *common.FileInfo: 文件信息
	//   error: 获取失败时返回错误信息
	GetFileInfo(ctx context.Context, remotePath string) (*common.FileInfo, error)
	
	// ===== 高级功能 =====
	
	// BatchTransfer 批量文件传输
	// 支持多个文件的批量上传或下载操作
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   operations: 传输操作列表
	//   options: 传输选项配置
	// 返回:
	//   *common.BatchTransferResult: 批量传输结果信息
	//   error: 传输失败时返回错误信息
	BatchTransfer(ctx context.Context, operations []*common.TransferOperation, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error)
	
	// SyncDirectory 目录同步
	// 将本地目录与远程目录进行同步，支持双向同步
	// 参数:
	//   ctx: 上下文
	//   localDir: 本地目录路径
	//   remoteDir: 远程目录路径
	//   syncMode: 同步模式（上传、下载、双向）
	//   options: 同步选项配置
	// 返回:
	//   *SyncResult: 同步结果信息
	//   error: 同步失败时返回错误信息
	SyncDirectory(ctx context.Context, localDir, remoteDir string, syncMode SyncMode, options *configs.SFTPSyncOptions) (*SyncResult, error)
	
	// GetConfig 获取客户端配置
	// 返回当前客户端使用的配置信息
	// 返回:
	//   *configs.SFTPConfig: 配置信息
	GetConfig() *configs.SFTPConfig
	
	// SetProgressCallback 设置进度回调函数
	// 设置用于监控传输进度的回调函数
	// 参数:
	//   callback: 进度回调函数
	SetProgressCallback(callback common.ProgressCallback)
	
	// SetErrorCallback 设置错误回调函数
	// 设置用于处理传输错误的回调函数
	// 参数:
	//   callback: 错误回调函数
	SetErrorCallback(callback common.ErrorCallback)
}

// SyncMode 同步模式枚举
type SyncMode int

const (
	// SyncModeUpload 仅上传（本地到远程）
	SyncModeUpload SyncMode = iota + 1
	
	// SyncModeDownload 仅下载（远程到本地）
	SyncModeDownload
	
	// SyncModeBidirectional 双向同步
	SyncModeBidirectional
)

// String 返回同步模式的字符串表示
func (s SyncMode) String() string {
	switch s {
	case SyncModeUpload:
		return "upload"
	case SyncModeDownload:
		return "download"
	case SyncModeBidirectional:
		return "bidirectional"
	default:
		return "unknown"
	}
}

// SyncResult 同步操作结果
// 包含目录同步操作的详细结果信息
type SyncResult struct {
	// 同步模式
	Mode SyncMode `json:"mode"`
	
	// 本地目录
	LocalDirectory string `json:"local_directory"`
	
	// 远程目录
	RemoteDirectory string `json:"remote_directory"`
	
	// 同步的文件数量
	FilesSynced int `json:"files_synced"`
	
	// 创建的目录数量
	DirectoriesCreated int `json:"directories_created"`
	
	// 删除的文件数量
	FilesDeleted int `json:"files_deleted"`
	
	// 删除的目录数量
	DirectoriesDeleted int `json:"directories_deleted"`
	
	// 冲突的文件数量
	ConflictedFiles int `json:"conflicted_files"`
	
	// 总传输字节数
	TotalBytesTransferred int64 `json:"total_bytes_transferred"`
	
	// 开始时间
	StartTime string `json:"start_time"`
	
	// 结束时间
	EndTime string `json:"end_time"`
	
	// 同步耗时
	Duration string `json:"duration"`
	
	// 是否成功
	Success bool `json:"success"`
	
	// 详细的同步操作列表
	Operations []*SyncOperation `json:"operations"`
	
	// 错误信息
	Errors []string `json:"errors,omitempty"`
}

// SyncOperation 同步操作详情
type SyncOperation struct {
	// 操作类型
	Action string `json:"action"` // create, update, delete, skip
	
	// 文件路径
	Path string `json:"path"`
	
	// 操作结果
	Success bool `json:"success"`
	
	// 错误信息
	Error string `json:"error,omitempty"`
	
	// 传输字节数
	BytesTransferred int64 `json:"bytes_transferred,omitempty"`
}

// NewClient 创建新的SFTP客户端
// 参数:
//   config: SFTP客户端配置
// 返回:
//   Client: SFTP客户端接口实现
//   error: 创建失败时返回错误信息
func NewClient(config *configs.SFTPConfig) (Client, error) {
	// 实际实现在client_impl.go中
	return NewSFTPClient(config)
} 