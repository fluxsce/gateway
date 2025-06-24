package sftp

import (
	"context"
	"fmt"
	"path/filepath"
	"time"
	
	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
)

// ===== 目录传输方法 =====

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
func (c *sftpClient) UploadDirectory(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error) {
	// TODO: 实现目录上传功能
	// 这是一个复杂的功能，需要递归遍历本地目录，创建远程目录结构，并上传所有文件
	return nil, common.NewNotImplementedError("UploadDirectory方法尚未实现")
}

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
func (c *sftpClient) DownloadDirectory(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error) {
	// TODO: 实现目录下载功能
	// 这是一个复杂的功能，需要递归遍历远程目录，创建本地目录结构，并下载所有文件
	return nil, common.NewNotImplementedError("DownloadDirectory方法尚未实现")
}

// ===== 文件和目录操作方法 =====

// ListDirectory 列出远程目录内容
// 获取远程目录中的文件和子目录列表信息
// 参数:
//   ctx: 上下文
//   remotePath: 远程目录路径
// 返回:
//   []*common.FileInfo: 文件信息列表
//   error: 操作失败时返回错误信息
func (c *sftpClient) ListDirectory(ctx context.Context, remotePath string) ([]*common.FileInfo, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 列出目录内容
	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("列出目录失败: %s", remotePath), err)
	}
	
	// 转换为通用文件信息格式
	var result []*common.FileInfo
	for _, file := range files {
		fileInfo := &common.FileInfo{
			Name:    file.Name(),
			Path:    filepath.Join(remotePath, file.Name()),
			Size:    file.Size(),
			IsDir:   file.IsDir(),
			ModTime: file.ModTime(),
			Mode:    file.Mode(),
		}
		result = append(result, fileInfo)
	}
	
	return result, nil
}

// CreateDirectory 在远程服务器创建目录
// 在远程SFTP服务器上创建指定路径的目录，支持递归创建
// 参数:
//   ctx: 上下文
//   remotePath: 要创建的远程目录路径
//   recursive: 是否递归创建父目录
// 返回:
//   error: 创建失败时返回错误信息
func (c *sftpClient) CreateDirectory(ctx context.Context, remotePath string, recursive bool) error {
	// 检查连接状态
	if !c.IsConnected() {
		return common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	if recursive {
		// 递归创建目录
		return c.createRemoteDirectory(remotePath)
	}
	
	// 创建单级目录
	if err := c.sftpClient.Mkdir(remotePath); err != nil {
		return common.NewIOError(fmt.Sprintf("创建目录失败: %s", remotePath), err)
	}
	
	return nil
}

// RemoveFile 删除远程文件
// 删除远程SFTP服务器上的指定文件
// 参数:
//   ctx: 上下文
//   remotePath: 要删除的远程文件路径
// 返回:
//   error: 删除失败时返回错误信息
func (c *sftpClient) RemoveFile(ctx context.Context, remotePath string) error {
	// 检查连接状态
	if !c.IsConnected() {
		return common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 删除文件
	if err := c.sftpClient.Remove(remotePath); err != nil {
		return common.NewIOError(fmt.Sprintf("删除文件失败: %s", remotePath), err)
	}
	
	return nil
}

// RemoveDirectory 删除远程目录
// 删除远程SFTP服务器上的指定目录，支持递归删除
// 参数:
//   ctx: 上下文
//   remotePath: 要删除的远程目录路径
//   recursive: 是否递归删除子目录和文件
// 返回:
//   error: 删除失败时返回错误信息
func (c *sftpClient) RemoveDirectory(ctx context.Context, remotePath string, recursive bool) error {
	// 检查连接状态
	if !c.IsConnected() {
		return common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	if !recursive {
		// 删除空目录
		if err := c.sftpClient.RemoveDirectory(remotePath); err != nil {
			return common.NewIOError(fmt.Sprintf("删除目录失败: %s", remotePath), err)
		}
		return nil
	}
	
	// 递归删除目录
	return c.removeDirectoryRecursive(remotePath)
}

// GetFileInfo 获取远程文件信息
// 获取远程文件或目录的详细信息
// 参数:
//   ctx: 上下文
//   remotePath: 远程文件或目录路径
// 返回:
//   *common.FileInfo: 文件信息
//   error: 获取失败时返回错误信息
func (c *sftpClient) GetFileInfo(ctx context.Context, remotePath string) (*common.FileInfo, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 获取文件信息
	info, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return nil, common.NewFileNotFoundError(remotePath, err)
	}
	
	return &common.FileInfo{
		Name:    info.Name(),
		Path:    remotePath,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
	}, nil
}

// ===== 高级功能方法 =====

// BatchTransfer 批量文件传输
// 支持多个文件的批量上传或下载操作
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   operations: 传输操作列表
//   options: 传输选项配置
// 返回:
//   *common.BatchTransferResult: 批量传输结果信息
//   error: 传输失败时返回错误信息
func (c *sftpClient) BatchTransfer(ctx context.Context, operations []*common.TransferOperation, options *configs.SFTPTransferOptions) (*common.BatchTransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 初始化批量传输结果
	result := &common.BatchTransferResult{
		TotalOperations:  len(operations),
		StartTime:        time.Now(),
		Results:          make([]*common.TransferResult, 0, len(operations)),
		Errors:           make([]string, 0),
	}
	
	// 执行每个传输操作
	for i, op := range operations {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			result.Errors = append(result.Errors, fmt.Sprintf("操作被取消: %v", ctx.Err()))
			break
		default:
		}
		
		// 执行单个传输操作
		transferResult, err := c.executeSingleTransferOperation(ctx, op, options)
		
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("操作 %d 失败: %v", i+1, err))
			result.FailedOperations++
		} else {
			result.SuccessfulOperations++
			if transferResult != nil {
				result.TotalBytesTransferred += transferResult.BytesTransferred
			}
		}
		
		// 添加传输结果到结果列表
		if transferResult != nil {
			result.Results = append(result.Results, transferResult)
		}
	}
	
	// 完成批量传输
	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)
	
	// 计算平均速度
	if result.TotalDuration > 0 {
		result.AverageSpeed = float64(result.TotalBytesTransferred) / result.TotalDuration.Seconds()
	}
	
	return result, nil
}

// executeSingleTransferOperation 执行单个传输操作
// 根据操作类型调用相应的传输方法
func (c *sftpClient) executeSingleTransferOperation(ctx context.Context, op *common.TransferOperation, options *configs.SFTPTransferOptions) (*common.TransferResult, error) {
	switch op.Type {
	case common.TransferTypeUpload:
		return c.UploadFile(ctx, op.LocalPath, op.RemotePath, options)
	case common.TransferTypeDownload:
		return c.DownloadFile(ctx, op.RemotePath, op.LocalPath, options)
	default:
		return nil, common.NewInvalidArgumentError(fmt.Sprintf("不支持的传输类型: %v", op.Type))
	}
}

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
func (c *sftpClient) SyncDirectory(ctx context.Context, localDir, remoteDir string, syncMode SyncMode, options *configs.SFTPSyncOptions) (*SyncResult, error) {
	// TODO: 实现目录同步功能
	// 这是一个非常复杂的功能，需要比较本地和远程目录的差异，并执行相应的同步操作
	return nil, common.NewNotImplementedError("SyncDirectory方法尚未实现")
}

// ===== 辅助方法 =====

// createRemoteDirectory 创建远程目录
// 递归创建远程目录，类似于 mkdir -p
func (c *sftpClient) createRemoteDirectory(path string) error {
	// 检查目录是否已存在
	info, err := c.sftpClient.Stat(path)
	if err == nil {
		if info.IsDir() {
			return nil // 目录已存在
		}
		return fmt.Errorf("路径已存在但不是目录: %s", path)
	}
	
	// 使用SFTP客户端的MkdirAll方法递归创建目录
	return c.sftpClient.MkdirAll(path)
}

// removeDirectoryRecursive 递归删除远程目录
// 深度优先遍历目录树，删除所有文件和子目录
func (c *sftpClient) removeDirectoryRecursive(remotePath string) error {
	// 列出目录内容
	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return common.NewIOError(fmt.Sprintf("读取目录失败: %s", remotePath), err)
	}
	
	// 递归删除子项
	for _, file := range files {
		childPath := filepath.Join(remotePath, file.Name())
		
		if file.IsDir() {
			// 递归删除子目录
			if err := c.removeDirectoryRecursive(childPath); err != nil {
				return err
			}
		} else {
			// 删除文件
			if err := c.sftpClient.Remove(childPath); err != nil {
				return common.NewIOError(fmt.Sprintf("删除文件失败: %s", childPath), err)
			}
		}
	}
	
	// 删除空目录
	if err := c.sftpClient.RemoveDirectory(remotePath); err != nil {
		return common.NewIOError(fmt.Sprintf("删除目录失败: %s", remotePath), err)
	}
	
	return nil
} 