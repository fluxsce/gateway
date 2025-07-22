// Package interfaces 定义SFTP工具接口
// 提供SFTP文件传输相关的接口规范
package interfaces

import (
	"context"
	"gateway/pkg/plugin/tools/configs"
	"gateway/pkg/plugin/tools/types"
)

// SFTPTool SFTP工具接口
// 继承ConnectableTool接口，添加SFTP特有的文件传输方法
// 提供完整的SFTP操作功能，支持文件传输、目录管理等操作
type SFTPTool interface {
	ConnectableTool

	// ===== 文件传输操作 =====

	// UploadFile 上传单个文件到远程服务器
	// 将本地文件上传到远程SFTP服务器的指定路径
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   localPath: 本地文件路径
	//   remotePath: 远程目标路径
	//   options: 上传选项配置
	// 返回:
	//   *types.TransferResult: 传输结果信息
	//   error: 上传失败时返回错误信息
	UploadFile(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error)

	// DownloadFile 从远程服务器下载单个文件
	// 将远程SFTP服务器上的文件下载到本地指定路径
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   remotePath: 远程文件路径
	//   localPath: 本地目标路径
	//   options: 下载选项配置
	// 返回:
	//   *types.TransferResult: 传输结果信息
	//   error: 下载失败时返回错误信息
	DownloadFile(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error)

	// UploadDirectory 上传整个目录到远程服务器
	// 递归上传本地目录及其所有子目录和文件到远程服务器
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   localDir: 本地目录路径
	//   remoteDir: 远程目标目录路径
	//   options: 上传选项配置
	// 返回:
	//   *types.BatchTransferResult: 批量传输结果信息
	//   error: 上传失败时返回错误信息
	UploadDirectory(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error)

	// DownloadDirectory 从远程服务器下载整个目录
	// 递归下载远程目录及其所有子目录和文件到本地
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   remoteDir: 远程目录路径
	//   localDir: 本地目标目录路径
	//   options: 下载选项配置
	// 返回:
	//   *types.BatchTransferResult: 批量传输结果信息
	//   error: 下载失败时返回错误信息
	DownloadDirectory(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error)

	// ===== 目录和文件操作 =====

	// ListDirectory 列出远程目录内容
	// 获取远程目录中的文件和子目录列表信息
	// 参数:
	//   ctx: 上下文
	//   remotePath: 远程目录路径
	// 返回:
	//   []*types.FileInfo: 文件信息列表
	//   error: 操作失败时返回错误信息
	ListDirectory(ctx context.Context, remotePath string) ([]*types.FileInfo, error)

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
	//   *types.FileInfo: 文件信息
	//   error: 获取失败时返回错误信息
	GetFileInfo(ctx context.Context, remotePath string) (*types.FileInfo, error)

	// ===== 高级功能 =====

	// BatchTransfer 批量文件传输
	// 支持多个文件的批量上传或下载操作
	// 参数:
	//   ctx: 上下文，用于进度监控和取消操作
	//   operations: 传输操作列表
	//   options: 传输选项配置
	// 返回:
	//   *types.BatchTransferResult: 批量传输结果信息
	//   error: 传输失败时返回错误信息
	BatchTransfer(ctx context.Context, operations []*types.TransferOperation, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error)

	// SyncDirectory 目录同步
	// 将本地目录与远程目录进行同步，支持双向同步
	// 参数:
	//   ctx: 上下文
	//   localDir: 本地目录路径
	//   remoteDir: 远程目录路径
	//   syncMode: 同步模式（上传、下载、双向）
	//   options: 同步选项配置
	// 返回:
	//   *types.SyncResult: 同步结果信息
	//   error: 同步失败时返回错误信息
	SyncDirectory(ctx context.Context, localDir, remoteDir string, syncMode types.SyncMode, options *configs.SFTPSyncOptions) (*types.SyncResult, error)

	// GetConfig 获取客户端配置
	// 返回当前客户端使用的配置信息
	// 返回:
	//   *configs.SFTPConfig: 配置信息
	GetConfig() *configs.SFTPConfig

	// SetProgressCallback 设置进度回调函数
	// 设置用于监控传输进度的回调函数
	// 参数:
	//   callback: 进度回调函数
	SetProgressCallback(callback types.ProgressCallback)

	// SetErrorCallback 设置错误回调函数
	// 设置用于处理传输错误的回调函数
	// 参数:
	//   callback: 错误回调函数
	SetErrorCallback(callback types.ErrorCallback)
}
