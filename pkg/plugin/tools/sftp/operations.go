package sftp

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"gateway/pkg/plugin/tools/common"
	"gateway/pkg/plugin/tools/configs"
	"gateway/pkg/plugin/tools/types"
)

// ===== 目录传输方法 =====

// uploadDirectoryImpl 上传整个目录到远程服务器的内部实现
// 递归上传本地目录及其所有子目录和文件到远程服务器
// 参数:
//
//	ctx: 上下文，用于进度监控和取消操作
//	localDir: 本地目录路径
//	remoteDir: 远程目标目录路径
//	options: 上传选项配置
//
// 返回:
//
//	*common.BatchTransferResult: 批量传输结果信息
//	error: 上传失败时返回错误信息
func (c *sftpClient) uploadDirectoryImpl(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// TODO: 实现目录上传功能
	// 这是一个复杂的功能，需要递归遍历本地目录，创建远程目录结构，并上传所有文件
	return nil, fmt.Errorf("UploadDirectory方法尚未实现")
}

// downloadDirectoryImpl 从远程服务器下载整个目录的内部实现
// 递归下载远程目录及其所有子目录和文件到本地
// 参数:
//
//	ctx: 上下文，用于进度监控和取消操作
//	remoteDir: 远程目录路径
//	localDir: 本地目标目录路径
//	options: 下载选项配置
//
// 返回:
//
//	*common.BatchTransferResult: 批量传输结果信息
//	error: 下载失败时返回错误信息
func (c *sftpClient) downloadDirectoryImpl(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// TODO: 实现目录下载功能
	// 这是一个复杂的功能，需要递归遍历远程目录，创建本地目录结构，并下载所有文件
	return nil, fmt.Errorf("DownloadDirectory方法尚未实现")
}

// ===== 文件和目录操作方法 =====

// listDirectoryImpl 列出远程目录内容的内部实现
// 获取远程目录中的文件和子目录列表信息
// 参数:
//
//	ctx: 上下文
//	remotePath: 远程目录路径
//
// 返回:
//
//	[]*common.FileInfo: 文件信息列表
//	error: 操作失败时返回错误信息
func (c *sftpClient) listDirectoryImpl(ctx context.Context, remotePath string) ([]*types.FileInfo, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}

	// 列出目录内容
	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return nil, fmt.Errorf("列出目录失败 %s: %w", remotePath, err)
	}

	// 转换为通用文件信息格式
	var result []*types.FileInfo
	for _, file := range files {
		fileInfo := &types.FileInfo{
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

// createDirectoryImpl 在远程服务器创建目录的内部实现
// 在远程SFTP服务器上创建指定路径的目录，支持递归创建
// 参数:
//
//	ctx: 上下文
//	remotePath: 要创建的远程目录路径
//	recursive: 是否递归创建父目录
//
// 返回:
//
//	error: 创建失败时返回错误信息
func (c *sftpClient) createDirectoryImpl(ctx context.Context, remotePath string, recursive bool) error {
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
		return fmt.Errorf("创建目录失败 %s: %w", remotePath, err)
	}

	return nil
}

// removeFileImpl 删除远程文件的内部实现
// 删除远程SFTP服务器上的指定文件
// 参数:
//
//	ctx: 上下文
//	remotePath: 要删除的远程文件路径
//
// 返回:
//
//	error: 删除失败时返回错误信息
func (c *sftpClient) removeFileImpl(ctx context.Context, remotePath string) error {
	// 检查连接状态
	if !c.IsConnected() {
		return common.NewConnectionError("未连接到SFTP服务器", nil)
	}

	// 删除文件
	if err := c.sftpClient.Remove(remotePath); err != nil {
		return fmt.Errorf("删除文件失败 %s: %w", remotePath, err)
	}

	return nil
}

// removeDirectoryImpl 删除远程目录的内部实现
// 删除远程SFTP服务器上的指定目录，支持递归删除
// 参数:
//
//	ctx: 上下文
//	remotePath: 要删除的远程目录路径
//	recursive: 是否递归删除子目录和文件
//
// 返回:
//
//	error: 删除失败时返回错误信息
func (c *sftpClient) removeDirectoryImpl(ctx context.Context, remotePath string, recursive bool) error {
	// 检查连接状态
	if !c.IsConnected() {
		return common.NewConnectionError("未连接到SFTP服务器", nil)
	}

	if !recursive {
		// 删除空目录
		if err := c.sftpClient.RemoveDirectory(remotePath); err != nil {
			return fmt.Errorf("删除目录失败 %s: %w", remotePath, err)
		}
		return nil
	}

	// 递归删除目录
	return c.removeDirectoryRecursive(remotePath)
}

// getFileInfoImpl 获取远程文件信息的内部实现
// 获取远程文件或目录的详细信息
// 参数:
//
//	ctx: 上下文
//	remotePath: 远程文件或目录路径
//
// 返回:
//
//	*common.FileInfo: 文件信息
//	error: 获取失败时返回错误信息
func (c *sftpClient) getFileInfoImpl(ctx context.Context, remotePath string) (*types.FileInfo, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}

	// 获取文件信息
	info, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return nil, common.NewFileNotFoundError(remotePath, err)
	}

	return &types.FileInfo{
		Name:    info.Name(),
		Path:    remotePath,
		Size:    info.Size(),
		IsDir:   info.IsDir(),
		ModTime: info.ModTime(),
		Mode:    info.Mode(),
	}, nil
}

// ===== 高级功能方法 =====

// batchTransferImpl 批量文件传输的内部实现
// 支持多个文件的批量上传或下载操作
// 参数:
//
//	ctx: 上下文，用于进度监控和取消操作
//	operations: 传输操作列表
//	options: 传输选项配置
//
// 返回:
//
//	*common.BatchTransferResult: 批量传输结果信息
//	error: 传输失败时返回错误信息
func (c *sftpClient) batchTransferImpl(ctx context.Context, operations []*types.TransferOperation, options *configs.SFTPTransferOptions) (*types.BatchTransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}

	// 初始化批量传输结果
	result := &types.BatchTransferResult{
		TotalOperations: len(operations),
		StartTime:       time.Now(),
		Results:         make([]*types.TransferResult, 0, len(operations)),
		Errors:          make([]string, 0),
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
func (c *sftpClient) executeSingleTransferOperation(ctx context.Context, op *types.TransferOperation, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	switch op.Type {
	case types.TransferTypeUpload:
		return c.uploadFileImpl(ctx, op.LocalPath, op.RemotePath, options)
	case types.TransferTypeDownload:
		return c.downloadFileImpl(ctx, op.RemotePath, op.LocalPath, options)
	default:
		return nil, fmt.Errorf("不支持的传输类型: %v", op.Type)
	}
}

// ===== 辅助方法 =====

// buildSafeRemotePath 构建安全的远程路径
// 正确处理路径分隔符和特殊字符
func (c *sftpClient) buildSafeRemotePath(remoteDir, fileName string) string {
	// 确保目录路径以正斜杠结尾
	if !strings.HasSuffix(remoteDir, "/") {
		remoteDir += "/"
	}

	// 组合路径
	fullPath := remoteDir + fileName

	// 规范化路径，确保使用正斜杠且没有双斜杠
	cleanPath := filepath.ToSlash(filepath.Clean(fullPath))

	// 处理以斜杠开头的情况
	if strings.HasPrefix(cleanPath, "./") {
		cleanPath = cleanPath[2:]
	}

	return cleanPath
}

// createRemoteDirectory 创建远程目录
// 递归创建远程目录，类似于 mkdir -p
func (c *sftpClient) createRemoteDirectory(path string) error {
	return c.createRemoteDirectorySafe(path)
}

// removeDirectoryRecursive 递归删除远程目录
// 深度优先遍历目录树，删除所有文件和子目录
func (c *sftpClient) removeDirectoryRecursive(remotePath string) error {
	// 列出目录内容
	files, err := c.sftpClient.ReadDir(remotePath)
	if err != nil {
		return fmt.Errorf("读取目录失败 %s: %w", remotePath, err)
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
				return fmt.Errorf("删除文件失败 %s: %w", childPath, err)
			}
		}
	}

	// 删除空目录
	if err := c.sftpClient.RemoveDirectory(remotePath); err != nil {
		return fmt.Errorf("删除目录失败 %s: %w", remotePath, err)
	}

	return nil
}
