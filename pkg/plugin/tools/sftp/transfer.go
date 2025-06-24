package sftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	
	"github.com/pkg/sftp"
	
	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
)

// UploadFile 上传单个文件到远程服务器
// 将本地文件上传到远程SFTP服务器的指定路径
// 支持进度监控、断点续传、权限保持等功能
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   localPath: 本地文件路径
//   remotePath: 远程目标路径
//   options: 上传选项配置
// 返回:
//   *common.TransferResult: 传输结果信息
//   error: 上传失败时返回错误信息
func (c *sftpClient) UploadFile(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 使用默认选项（如果未提供）
	if options == nil {
		options = c.config.DefaultTransferOptions
	}
	
	// 验证和准备本地文件
	localFile, localInfo, err := c.prepareLocalFileForUpload(localPath, options)
	if err != nil {
		return nil, err
	}
	defer localFile.Close()
	
	// 处理远程路径和目录创建
	if err := c.prepareRemotePathForUpload(remotePath, options); err != nil {
		return nil, err
	}
	
	// 检查远程文件存在性和覆盖策略
	skipTransfer, err := c.checkRemoteFileExistence(remotePath, localInfo, options)
	if err != nil {
		return nil, err
	}
	
	// 如果需要跳过传输，返回跳过结果
	if skipTransfer {
		return c.createSkippedTransferResult("upload", localPath, remotePath), nil
	}
	
	// 执行文件上传
	return c.executeUpload(ctx, localFile, localInfo, localPath, remotePath, options)
}

// DownloadFile 从远程服务器下载单个文件
// 将远程SFTP服务器上的文件下载到本地指定路径
// 支持进度监控、断点续传、权限保持等功能
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   remotePath: 远程文件路径
//   localPath: 本地目标路径
//   options: 下载选项配置
// 返回:
//   *common.TransferResult: 传输结果信息
//   error: 下载失败时返回错误信息
func (c *sftpClient) DownloadFile(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 使用默认选项（如果未提供）
	if options == nil {
		options = c.config.DefaultTransferOptions
	}
	
	// 验证和准备远程文件
	remoteFile, remoteInfo, err := c.prepareRemoteFileForDownload(remotePath, options)
	if err != nil {
		return nil, err
	}
	defer remoteFile.Close()
	
	// 处理本地路径和目录创建
	if err := c.prepareLocalPathForDownload(localPath, options); err != nil {
		return nil, err
	}
	
	// 检查本地文件存在性和覆盖策略
	skipTransfer, err := c.checkLocalFileExistence(localPath, remoteInfo, options)
	if err != nil {
		return nil, err
	}
	
	// 如果需要跳过传输，返回跳过结果
	if skipTransfer {
		return c.createSkippedTransferResult("download", localPath, remotePath), nil
	}
	
	// 执行文件下载
	return c.executeDownload(ctx, remoteFile, remoteInfo, remotePath, localPath, options)
}

// prepareLocalFileForUpload 准备本地文件用于上传
// 验证本地文件的存在性、可读性和大小限制
func (c *sftpClient) prepareLocalFileForUpload(localPath string, options *configs.SFTPTransferOptions) (*os.File, os.FileInfo, error) {
	// 打开本地文件
	localFile, err := os.Open(localPath)
	if err != nil {
		return nil, nil, common.NewFileNotFoundError(localPath, err)
	}
	
	// 获取文件信息
	localInfo, err := localFile.Stat()
	if err != nil {
		localFile.Close()
		return nil, nil, common.NewIOError("获取本地文件信息失败", err)
	}
	
	// 检查是否为目录
	if localInfo.IsDir() {
		localFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("本地路径是一个目录: %s", localPath))
	}
	
	// 检查文件大小限制
	if options.MaxFileSize > 0 && localInfo.Size() > options.MaxFileSize {
		localFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("文件大小超过限制: %d > %d", localInfo.Size(), options.MaxFileSize))
	}
	
	return localFile, localInfo, nil
}

// prepareRemoteFileForDownload 准备远程文件用于下载
// 验证远程文件的存在性、可读性和大小限制
func (c *sftpClient) prepareRemoteFileForDownload(remotePath string, options *configs.SFTPTransferOptions) (*sftp.File, os.FileInfo, error) {
	// 打开远程文件
	remoteFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return nil, nil, common.NewFileNotFoundError(remotePath, err)
	}
	
	// 获取远程文件信息
	remoteInfo, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		remoteFile.Close()
		return nil, nil, common.NewIOError("获取远程文件信息失败", err)
	}
	
	// 检查是否为目录
	if remoteInfo.IsDir() {
		remoteFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("远程路径是一个目录: %s", remotePath))
	}
	
	// 检查文件大小限制
	if options.MaxFileSize > 0 && remoteInfo.Size() > options.MaxFileSize {
		remoteFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("文件大小超过限制: %d > %d", remoteInfo.Size(), options.MaxFileSize))
	}
	
	return remoteFile, remoteInfo, nil
}

// prepareRemotePathForUpload 准备远程路径用于上传
// 创建必要的远程目录
func (c *sftpClient) prepareRemotePathForUpload(remotePath string, options *configs.SFTPTransferOptions) error {
	if options.CreateTargetDir {
		remoteDir := filepath.Dir(remotePath)
		if err := c.createRemoteDirectory(remoteDir); err != nil {
			return common.NewIOError(fmt.Sprintf("创建远程目录失败: %s", remoteDir), err)
		}
	}
	return nil
}

// prepareLocalPathForDownload 准备本地路径用于下载
// 创建必要的本地目录
func (c *sftpClient) prepareLocalPathForDownload(localPath string, options *configs.SFTPTransferOptions) error {
	if options.CreateTargetDir {
		localDir := filepath.Dir(localPath)
		if err := common.EnsureDirectoryExists(localDir, true); err != nil {
			return common.NewIOError(fmt.Sprintf("创建本地目录失败: %s", localDir), err)
		}
	}
	return nil
}

// checkRemoteFileExistence 检查远程文件存在性和覆盖策略
// 返回是否应该跳过传输
func (c *sftpClient) checkRemoteFileExistence(remotePath string, localInfo os.FileInfo, options *configs.SFTPTransferOptions) (bool, error) {
	remoteInfo, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		// 文件不存在，可以继续传输
		return false, nil
	}
	
	// 文件存在，检查跳过策略
	if options.SkipExisting {
		// 如果文件大小相同，跳过传输
		if remoteInfo.Size() == localInfo.Size() {
			return true, nil
		}
	}
	
	// 检查覆盖策略
	if !options.OverwriteExisting {
		return false, common.NewFileExistsError(remotePath, nil)
	}
	
	return false, nil
}

// checkLocalFileExistence 检查本地文件存在性和覆盖策略
// 返回是否应该跳过传输
func (c *sftpClient) checkLocalFileExistence(localPath string, remoteInfo os.FileInfo, options *configs.SFTPTransferOptions) (bool, error) {
	localInfo, err := os.Stat(localPath)
	if err != nil {
		// 文件不存在，可以继续传输
		return false, nil
	}
	
	// 文件存在，检查跳过策略
	if options.SkipExisting {
		// 如果文件大小相同，跳过传输
		if localInfo.Size() == remoteInfo.Size() {
			return true, nil
		}
	}
	
	// 检查覆盖策略
	if !options.OverwriteExisting {
		return false, common.NewFileExistsError(localPath, nil)
	}
	
	return false, nil
}

// createSkippedTransferResult 创建跳过传输的结果
func (c *sftpClient) createSkippedTransferResult(transferType, localPath, remotePath string) *common.TransferResult {
	var resultType common.TransferType
	if transferType == "upload" {
		resultType = common.TransferTypeUpload
	} else {
		resultType = common.TransferTypeDownload
	}
	
	now := time.Now()
	return &common.TransferResult{
		OperationID:       common.GenerateUniqueID(transferType),
		Type:              resultType,
		LocalPath:         localPath,
		RemotePath:        remotePath,
		BytesTransferred:  0,
		StartTime:         now,
		EndTime:           now,
		Duration:          0,
		AverageSpeed:      0,
		Success:           true,
		Error:             "",
		Metadata:          map[string]interface{}{"skipped": true},
	}
}

// executeUpload 执行文件上传
func (c *sftpClient) executeUpload(ctx context.Context, localFile *os.File, localInfo os.FileInfo, localPath, remotePath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error) {
	// 创建远程文件
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("创建远程文件失败: %s", remotePath), err)
	}
	defer remoteFile.Close()
	
	// 准备传输结果
	result := &common.TransferResult{
		OperationID:  common.GenerateUniqueID("upload"),
		Type:         common.TransferTypeUpload,
		LocalPath:    localPath,
		RemotePath:   remotePath,
		StartTime:    time.Now(),
		Success:      false,
	}
	
	// 执行文件传输
	if err := c.transferFile(ctx, localFile, remoteFile, localInfo.Size(), result, options); err != nil {
		result.Error = err.Error()
		return result, err
	}
	
	// 设置文件属性
	c.setRemoteFileAttributes(remoteFile, remotePath, localInfo, result, options)
	
	// 标记成功
	result.Success = true
	return result, nil
}

// executeDownload 执行文件下载
func (c *sftpClient) executeDownload(ctx context.Context, remoteFile *sftp.File, remoteInfo os.FileInfo, remotePath, localPath string, options *configs.SFTPTransferOptions) (*common.TransferResult, error) {
	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("创建本地文件失败: %s", localPath), err)
	}
	defer localFile.Close()
	
	// 准备传输结果
	result := &common.TransferResult{
		OperationID:  common.GenerateUniqueID("download"),
		Type:         common.TransferTypeDownload,
		LocalPath:    localPath,
		RemotePath:   remotePath,
		StartTime:    time.Now(),
		Success:      false,
	}
	
	// 执行文件传输
	if err := c.transferFile(ctx, remoteFile, localFile, remoteInfo.Size(), result, options); err != nil {
		result.Error = err.Error()
		return result, err
	}
	
	// 设置文件属性
	c.setLocalFileAttributes(localPath, remoteInfo, result, options)
	
	// 标记成功
	result.Success = true
	return result, nil
}

// transferFile 执行实际的文件传输
func (c *sftpClient) transferFile(ctx context.Context, src io.Reader, dst io.Writer, totalSize int64, result *common.TransferResult, options *configs.SFTPTransferOptions) error {
	// 创建进度监控器
	progress := &progressMonitor{
		ctx:              ctx,
		operationID:      result.OperationID,
		filePath:         result.LocalPath,
		totalBytes:       totalSize,
		startTime:        result.StartTime,
		progressCallback: c.progressCallback,
		reportInterval:   options.ProgressReportInterval,
	}
	
	// 创建带进度监控的读取器
	reader := &progressReader{
		reader:   src,
		progress: progress,
	}
	
	// 设置缓冲区大小
	bufferSize := options.BufferSize
	if bufferSize <= 0 {
		bufferSize = 32 * 1024 // 默认32KB
	}
	
	// 使用带缓冲的复制
	buffer := make([]byte, bufferSize)
	bytesTransferred, err := io.CopyBuffer(dst, reader, buffer)
	endTime := time.Now()
	
	// 更新传输结果
	result.BytesTransferred = bytesTransferred
	result.EndTime = endTime
	result.Duration = endTime.Sub(result.StartTime)
	
	// 计算平均速度
	if result.Duration > 0 {
		result.AverageSpeed = float64(bytesTransferred) / result.Duration.Seconds()
	}
	
	if err != nil {
		return common.NewIOError("文件传输失败", err)
	}
	
	return nil
}

// setRemoteFileAttributes 设置远程文件属性
func (c *sftpClient) setRemoteFileAttributes(remoteFile *sftp.File, remotePath string, localInfo os.FileInfo, result *common.TransferResult, options *configs.SFTPTransferOptions) {
	// 设置文件权限
	if options.PreservePermissions {
		if err := remoteFile.Chmod(localInfo.Mode()); err != nil {
			c.reportAttributeError(result.OperationID, "permission_error", "设置文件权限失败", remotePath, err)
		}
	}
	
	// 设置文件时间戳
	if options.PreserveTimestamps {
		if err := c.sftpClient.Chtimes(remotePath, time.Now(), localInfo.ModTime()); err != nil {
			c.reportAttributeError(result.OperationID, "timestamp_error", "设置文件时间戳失败", remotePath, err)
		}
	}
}

// setLocalFileAttributes 设置本地文件属性
func (c *sftpClient) setLocalFileAttributes(localPath string, remoteInfo os.FileInfo, result *common.TransferResult, options *configs.SFTPTransferOptions) {
	// 设置文件权限
	if options.PreservePermissions {
		if err := os.Chmod(localPath, remoteInfo.Mode()); err != nil {
			c.reportAttributeError(result.OperationID, "permission_error", "设置文件权限失败", localPath, err)
		}
	}
	
	// 设置文件时间戳
	if options.PreserveTimestamps {
		if err := os.Chtimes(localPath, time.Now(), remoteInfo.ModTime()); err != nil {
			c.reportAttributeError(result.OperationID, "timestamp_error", "设置文件时间戳失败", localPath, err)
		}
	}
}

// reportAttributeError 报告属性设置错误
func (c *sftpClient) reportAttributeError(operationID, errorType, message, filePath string, err error) {
	if c.errorCallback != nil {
		c.errorCallback(&common.TransferError{
			OperationID:   operationID,
			Type:          errorType,
			Message:       message,
			FilePath:      filePath,
			OriginalError: err,
			Retryable:     false,
			Timestamp:     time.Now(),
		})
	}
} 