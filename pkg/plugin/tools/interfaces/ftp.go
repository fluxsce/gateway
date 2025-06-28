// Package interfaces 定义FTP工具接口
// 提供FTP文件传输相关的接口规范
package interfaces

import (
	"context"
	
	"gohub/pkg/plugin/tools/types"
)

// FTPTool FTP工具接口
// 继承ConnectableTool接口，添加FTP特有的文件传输方法
type FTPTool interface {
	ConnectableTool
	
	// ===== 文件传输操作 =====
	
	// UploadFile 上传文件
	// 将本地文件上传到远程FTP服务器
	UploadFile(ctx context.Context, localPath, remotePath string) error
	
	// DownloadFile 下载文件
	// 从远程FTP服务器下载文件到本地
	DownloadFile(ctx context.Context, remotePath, localPath string) error
	
	// UploadDirectory 上传目录
	// 递归上传整个目录到远程FTP服务器
	UploadDirectory(ctx context.Context, localDir, remoteDir string) error
	
	// DownloadDirectory 下载目录
	// 递归下载整个目录到本地
	DownloadDirectory(ctx context.Context, remoteDir, localDir string) error
	
	// ===== 目录和文件操作 =====
	
	// ListDirectory 列出目录内容
	// 获取远程目录中的文件和子目录列表
	ListDirectory(ctx context.Context, remotePath string) ([]string, error)
	
	// ListDirectoryDetails 列出目录详细信息
	// 获取远程目录中文件和子目录的详细信息
	ListDirectoryDetails(ctx context.Context, remotePath string) ([]*FTPFileInfo, error)
	
	// CreateDirectory 创建目录
	// 在远程FTP服务器上创建目录
	CreateDirectory(ctx context.Context, remotePath string) error
	
	// RemoveFile 删除文件
	// 删除远程FTP服务器上的文件
	RemoveFile(ctx context.Context, remotePath string) error
	
	// RemoveDirectory 删除目录
	// 删除远程FTP服务器上的目录
	RemoveDirectory(ctx context.Context, remotePath string) error
	
	// RenameFile 重命名文件
	// 重命名远程FTP服务器上的文件或目录
	RenameFile(ctx context.Context, oldPath, newPath string) error
	
	// ===== 文件信息操作 =====
	
	// GetFileInfo 获取文件信息
	// 获取远程文件或目录的详细信息
	GetFileInfo(ctx context.Context, remotePath string) (*FTPFileInfo, error)
	
	// FileExists 检查文件是否存在
	// 检查远程路径上的文件或目录是否存在
	FileExists(ctx context.Context, remotePath string) (bool, error)
	
	// GetFileSize 获取文件大小
	// 获取远程文件的大小（字节）
	GetFileSize(ctx context.Context, remotePath string) (int64, error)
	
	// GetCurrentDirectory 获取当前工作目录
	// 获取FTP服务器上的当前工作目录
	GetCurrentDirectory(ctx context.Context) (string, error)
	
	// ChangeDirectory 切换工作目录
	// 切换FTP服务器上的工作目录
	ChangeDirectory(ctx context.Context, remotePath string) error
	
	// ===== 传输模式操作 =====
	
	// SetTransferMode 设置传输模式
	// 设置FTP传输模式（ASCII或Binary）
	SetTransferMode(ctx context.Context, mode FTPTransferMode) error
	
	// GetTransferMode 获取传输模式
	// 获取当前的FTP传输模式
	GetTransferMode(ctx context.Context) (FTPTransferMode, error)
	
	// SetPassiveMode 设置被动模式
	// 启用或禁用FTP被动模式
	SetPassiveMode(ctx context.Context, passive bool) error
	
	// IsPassiveMode 检查是否为被动模式
	// 检查当前是否启用了被动模式
	IsPassiveMode(ctx context.Context) (bool, error)
	
	// ===== 高级操作 =====
	
	// BatchTransfer 批量传输
	// 批量执行多个文件传输操作
	BatchTransfer(ctx context.Context, operations []FTPOperation) error
	
	// SyncDirectory 目录同步
	// 同步本地和远程目录
	SyncDirectory(ctx context.Context, localDir, remoteDir string, mode types.SyncMode) error
	
	// TestConnection 测试连接
	// 测试FTP连接是否正常
	TestConnection(ctx context.Context) error
	
	// SendCommand 发送原始FTP命令
	// 发送原始的FTP协议命令
	SendCommand(ctx context.Context, command string) (*FTPResponse, error)
}

// FTPFileInfo FTP文件信息
type FTPFileInfo struct {
	// 文件名
	Name string `json:"name"`
	
	// 文件大小（字节）
	Size int64 `json:"size"`
	
	// 是否为目录
	IsDir bool `json:"is_dir"`
	
	// 文件权限
	Mode string `json:"mode"`
	
	// 修改时间
	ModTime string `json:"mod_time"`
	
	// 所有者
	Owner string `json:"owner"`
	
	// 组
	Group string `json:"group"`
	
	// 文件路径
	Path string `json:"path"`
	
	// 链接数
	Links int `json:"links"`
}

// FTPTransferMode FTP传输模式
type FTPTransferMode int

const (
	// FTPModeASCII ASCII模式（文本文件）
	FTPModeASCII FTPTransferMode = iota + 1
	
	// FTPModeBinary 二进制模式（所有文件类型）
	FTPModeBinary
)

// String 返回传输模式的字符串表示
func (m FTPTransferMode) String() string {
	switch m {
	case FTPModeASCII:
		return "ascii"
	case FTPModeBinary:
		return "binary"
	default:
		return "unknown"
	}
}

// FTPOperation FTP操作
type FTPOperation struct {
	// 操作类型
	Type FTPOperationType `json:"type"`
	
	// 本地路径
	LocalPath string `json:"local_path"`
	
	// 远程路径
	RemotePath string `json:"remote_path"`
	
	// 操作参数
	Options map[string]interface{} `json:"options,omitempty"`
}

// FTPOperationType FTP操作类型
type FTPOperationType int

const (
	// FTPOpUploadFile 上传文件
	FTPOpUploadFile FTPOperationType = iota + 1
	
	// FTPOpDownloadFile 下载文件
	FTPOpDownloadFile
	
	// FTPOpUploadDir 上传目录
	FTPOpUploadDir
	
	// FTPOpDownloadDir 下载目录
	FTPOpDownloadDir
	
	// FTPOpCreateDir 创建目录
	FTPOpCreateDir
	
	// FTPOpRemoveFile 删除文件
	FTPOpRemoveFile
	
	// FTPOpRemoveDir 删除目录
	FTPOpRemoveDir
	
	// FTPOpRename 重命名
	FTPOpRename
)

// String 返回操作类型的字符串表示
func (t FTPOperationType) String() string {
	switch t {
	case FTPOpUploadFile:
		return "upload_file"
	case FTPOpDownloadFile:
		return "download_file"
	case FTPOpUploadDir:
		return "upload_dir"
	case FTPOpDownloadDir:
		return "download_dir"
	case FTPOpCreateDir:
		return "create_dir"
	case FTPOpRemoveFile:
		return "remove_file"
	case FTPOpRemoveDir:
		return "remove_dir"
	case FTPOpRename:
		return "rename"
	default:
		return "unknown"
	}
}

// FTPResponse FTP响应
type FTPResponse struct {
	// 响应码
	Code int `json:"code"`
	
	// 响应消息
	Message string `json:"message"`
	
	// 是否成功
	Success bool `json:"success"`
	
	// 原始响应
	Raw string `json:"raw"`
} 