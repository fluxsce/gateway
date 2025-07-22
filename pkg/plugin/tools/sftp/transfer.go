package sftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"

	"gohub/pkg/plugin/tools/common"
	"gohub/pkg/plugin/tools/configs"
	"gohub/pkg/plugin/tools/types"
)

// uploadFileImpl 上传单个文件或目录到远程服务器的内部实现
// 将本地文件或目录中的文件上传到远程SFTP服务器的指定路径
// 如果本地路径是目录，将上传目录中的所有文件（不递归子目录）
// 支持进度监控、断点续传、权限保持等功能
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   localPath: 本地文件或目录路径
//   remotePath: 远程目标路径
//   options: 上传选项配置
// 返回:
//   *common.TransferResult: 传输结果信息
//   error: 上传失败时返回错误信息
func (c *sftpClient) uploadFileImpl(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 使用默认选项（如果未提供）
	if options == nil {
		options = c.config.DefaultTransferOptions
	}
	
	// 检查本地路径是文件还是目录
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return nil, common.NewFileNotFoundError(localPath, err)
	}
	
	// 如果是目录，执行目录文件上传
	if localInfo.IsDir() {
		return c.uploadDirectoryFiles(ctx, localPath, remotePath, options)
	}
	
	// 如果是文件，执行单文件上传
	return c.uploadSingleFile(ctx, localPath, remotePath, options)
}

// uploadSingleFile 上传单个文件的实现
// 这是原来uploadFileImpl中的单文件上传逻辑
func (c *sftpClient) uploadSingleFile(ctx context.Context, localPath, remotePath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
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

// uploadDirectoryFiles 上传目录中的所有文件（不递归子目录）
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   localDir: 本地目录路径
//   remoteDir: 远程目标目录路径
//   options: 上传选项配置
// 返回:
//   *types.TransferResult: 汇总的传输结果信息
//   error: 上传失败时返回错误信息
func (c *sftpClient) uploadDirectoryFiles(ctx context.Context, localDir, remoteDir string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 读取本地目录内容
	entries, err := os.ReadDir(localDir)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("读取本地目录失败: %s", localDir), err)
	}
	
	// 确保远程目录存在（检查createTargetDir选项）
	if options.CreateTargetDir {
		if err := c.createRemoteDirectorySafe(remoteDir); err != nil {
			return nil, common.NewIOError(fmt.Sprintf("创建远程目录失败: %s", remoteDir), err)
		}
	}
	
	// 初始化汇总结果
	summaryResult := &types.TransferResult{
		OperationID:       fmt.Sprintf("upload-dir-%d", time.Now().UnixNano()),
		Type:              types.TransferTypeUpload,
		LocalPath:         localDir,
		RemotePath:        remoteDir,
		StartTime:         time.Now(),
		Success:           true,
		Metadata:          make(map[string]interface{}),
	}
	
	var totalBytes int64
	var fileCount int
	var errors []string
	var uploadedFiles []string
	
	// 遍历目录中的文件（不递归子目录）
	for _, entry := range entries {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			summaryResult.Success = false
			summaryResult.Error = fmt.Sprintf("操作被取消: %v", ctx.Err())
			break
		default:
		}
		
		// 跳过子目录
		if entry.IsDir() {
			continue
		}
		
		// 构造文件路径，处理特殊字符
		localFilePath := filepath.Join(localDir, entry.Name())
		// 远程路径使用正斜杠作为分隔符（SFTP标准），安全处理特殊字符
		remoteFilePath := c.buildSafeRemotePath(remoteDir, entry.Name())
		
		// 上传单个文件
		fileResult, err := c.uploadSingleFile(ctx, localFilePath, remoteFilePath, options)
		if err != nil {
			errors = append(errors, fmt.Sprintf("上传文件 %s 失败: %v", entry.Name(), err))
			summaryResult.Success = false
		} else if fileResult != nil {
			totalBytes += fileResult.BytesTransferred
			fileCount++
			uploadedFiles = append(uploadedFiles, entry.Name())
		}
	}
	
	// 完成汇总
	summaryResult.EndTime = time.Now()
	summaryResult.Duration = summaryResult.EndTime.Sub(summaryResult.StartTime)
	summaryResult.BytesTransferred = totalBytes
	
	// 计算平均速度
	if summaryResult.Duration > 0 {
		summaryResult.AverageSpeed = float64(totalBytes) / summaryResult.Duration.Seconds()
	}
	
	// 设置元数据
	summaryResult.Metadata["file_count"] = fileCount
	summaryResult.Metadata["total_files_processed"] = len(entries)
	summaryResult.Metadata["subdirectories_skipped"] = len(entries) - fileCount
	summaryResult.Metadata["delete_source_enabled"] = c.shouldDeleteSourceAfterTransfer(options)
	summaryResult.Metadata["uploaded_files"] = uploadedFiles
	summaryResult.Metadata["create_target_dir_enabled"] = options.CreateTargetDir
	
	// 如果有错误，将错误信息添加到结果中
	if len(errors) > 0 {
		summaryResult.Error = fmt.Sprintf("部分文件上传失败: %s", strings.Join(errors, "; "))
	}
	
	return summaryResult, nil
}

// downloadFileImpl 从远程服务器下载单个文件或目录的内部实现
// 将远程SFTP服务器上的文件或目录中的文件下载到本地指定路径
// 如果远程路径是目录，将下载目录中的所有文件（不递归子目录）
// 支持进度监控、断点续传、权限保持等功能
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   remotePath: 远程文件或目录路径
//   localPath: 本地目标路径
//   options: 下载选项配置
// 返回:
//   *common.TransferResult: 传输结果信息
//   error: 下载失败时返回错误信息
func (c *sftpClient) downloadFileImpl(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 检查连接状态
	if !c.IsConnected() {
		return nil, common.NewConnectionError("未连接到SFTP服务器", nil)
	}
	
	// 使用默认选项（如果未提供）
	if options == nil {
		options = c.config.DefaultTransferOptions
	}
	
	// 检查远程路径是文件还是目录
	remoteInfo, err := c.sftpClient.Stat(remotePath)
	if err != nil {
		return nil, common.NewFileNotFoundError(remotePath, err)
	}
	
	// 如果是目录，执行目录文件下载
	if remoteInfo.IsDir() {
		return c.downloadDirectoryFiles(ctx, remotePath, localPath, options)
	}
	
	// 如果是文件，执行单文件下载
	return c.downloadSingleFile(ctx, remotePath, localPath, options)
}

// downloadSingleFile 下载单个文件的实现
// 这是原来downloadFileImpl中的单文件下载逻辑
func (c *sftpClient) downloadSingleFile(ctx context.Context, remotePath, localPath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
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

// downloadDirectoryFiles 下载远程目录中的所有文件（不递归子目录）
// 参数:
//   ctx: 上下文，用于进度监控和取消操作
//   remoteDir: 远程目录路径
//   localDir: 本地目标目录路径
//   options: 下载选项配置
// 返回:
//   *types.TransferResult: 汇总的传输结果信息
//   error: 下载失败时返回错误信息
func (c *sftpClient) downloadDirectoryFiles(ctx context.Context, remoteDir, localDir string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 读取远程目录内容
	files, err := c.sftpClient.ReadDir(remoteDir)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("读取远程目录失败: %s", remoteDir), err)
	}
	
	// 确保本地目录存在（检查createTargetDir选项）
	if options.CreateTargetDir {
		if err := common.EnsureDirectoryExists(localDir, true); err != nil {
			return nil, common.NewIOError(fmt.Sprintf("创建本地目录失败: %s", localDir), err)
		}
	}
	
	// 初始化汇总结果
	summaryResult := &types.TransferResult{
		OperationID:       fmt.Sprintf("download-dir-%d", time.Now().UnixNano()),
		Type:              types.TransferTypeDownload,
		LocalPath:         localDir,
		RemotePath:        remoteDir,
		StartTime:         time.Now(),
		Success:           true,
		Metadata:          make(map[string]interface{}),
	}
	
	var totalBytes int64
	var fileCount int
	var errors []string
	var downloadedFiles []string
	
	// 遍历目录中的文件（不递归子目录）
	for _, file := range files {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			summaryResult.Success = false
			summaryResult.Error = fmt.Sprintf("操作被取消: %v", ctx.Err())
			break
		default:
		}
		
		// 跳过子目录
		if file.IsDir() {
			continue
		}
		
		// 构造文件路径，处理特殊字符
		// 远程路径使用正斜杠作为分隔符（SFTP标准），安全处理特殊字符
		remoteFilePath := c.buildSafeRemotePath(remoteDir, file.Name())
		localFilePath := filepath.Join(localDir, file.Name())
		
		// 下载单个文件
		fileResult, err := c.downloadSingleFile(ctx, remoteFilePath, localFilePath, options)
		if err != nil {
			errors = append(errors, fmt.Sprintf("下载文件 %s 失败: %v", file.Name(), err))
			summaryResult.Success = false
		} else if fileResult != nil {
			totalBytes += fileResult.BytesTransferred
			fileCount++
			downloadedFiles = append(downloadedFiles, file.Name())
		}
	}
	
	// 完成汇总
	summaryResult.EndTime = time.Now()
	summaryResult.Duration = summaryResult.EndTime.Sub(summaryResult.StartTime)
	summaryResult.BytesTransferred = totalBytes
	
	// 计算平均速度
	if summaryResult.Duration > 0 {
		summaryResult.AverageSpeed = float64(totalBytes) / summaryResult.Duration.Seconds()
	}
	
	// 设置元数据
	summaryResult.Metadata["file_count"] = fileCount
	summaryResult.Metadata["total_files_processed"] = len(files)
	summaryResult.Metadata["subdirectories_skipped"] = len(files) - fileCount
	summaryResult.Metadata["delete_source_enabled"] = c.shouldDeleteSourceAfterTransfer(options)
	summaryResult.Metadata["downloaded_files"] = downloadedFiles
	summaryResult.Metadata["create_target_dir_enabled"] = options.CreateTargetDir
	
	// 如果有错误，将错误信息添加到结果中
	if len(errors) > 0 {
		summaryResult.Error = fmt.Sprintf("部分文件下载失败: %s", strings.Join(errors, "; "))
	}
	
	return summaryResult, nil
}

// prepareLocalFileForUpload 准备本地文件用于上传
// 验证本地文件的存在性、可读性和大小限制
// 注意：此方法只处理文件，不处理目录
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
	
	// 检查是否为目录（在这个方法中目录应该被视为错误，因为目录处理已经在上层分离）
	if localInfo.IsDir() {
		localFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("此方法只处理文件，收到目录路径: %s", localPath))
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
// 注意：此方法只处理文件，不处理目录
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
	
	// 检查是否为目录（在这个方法中目录应该被视为错误，因为目录处理已经在上层分离）
	if remoteInfo.IsDir() {
		remoteFile.Close()
		return nil, nil, common.NewInvalidArgumentError(fmt.Sprintf("此方法只处理文件，收到目录路径: %s", remotePath))
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
		if err := c.createRemoteDirectorySafe(remoteDir); err != nil {
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
func (c *sftpClient) createSkippedTransferResult(transferType, localPath, remotePath string) *types.TransferResult {
	var resultType types.TransferType
	if transferType == "upload" {
		resultType = types.TransferTypeUpload
	} else {
		resultType = types.TransferTypeDownload
	}
	
	now := time.Now()
	return &types.TransferResult{
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

// backupLocalFile 备份本地文件到本地备份目录
func (c *sftpClient) backupLocalFile(sourcePath, backupDir string, result *types.TransferResult) error {
	if backupDir == "" {
		return nil
	}

	// 确保备份目录存在
	if err := common.EnsureDirectoryExists(backupDir, true); err != nil {
		return fmt.Errorf("创建备份目录失败 %s: %w", backupDir, err)
	}

	// 构建备份文件路径
	fileName := filepath.Base(sourcePath)
	backupPath := filepath.Join(backupDir, fileName)

	// 复制文件到备份目录
	bytesWritten, err := common.CopyFile(sourcePath, backupPath, 0644)
	if err != nil {
		return fmt.Errorf("备份文件失败 %s -> %s: %w", sourcePath, backupPath, err)
	}

	// 记录备份信息到结果元数据
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["backup_path"] = backupPath
	result.Metadata["backup_created"] = true
	result.Metadata["backup_size"] = bytesWritten

	return nil
}

// backupRemoteFile 备份远程文件到SFTP备份目录
func (c *sftpClient) backupRemoteFile(remotePath, backupDir string, result *types.TransferResult) error {
	if backupDir == "" {
		return nil
	}

	// 确保远程备份目录存在
	if err := c.createRemoteDirectorySafe(backupDir); err != nil {
		return fmt.Errorf("创建远程备份目录失败 %s: %w", backupDir, err)
	}

	// 构建备份文件路径
	fileName := filepath.Base(remotePath)
	backupPath := c.buildSafeRemotePath(backupDir, fileName)

	// 打开源文件
	sourceFile, err := c.sftpClient.Open(remotePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败 %s: %w", remotePath, err)
	}
	defer sourceFile.Close()

	// 创建备份文件
	backupFile, err := c.sftpClient.Create(backupPath)
	if err != nil {
		return fmt.Errorf("创建备份文件失败 %s: %w", backupPath, err)
	}
	defer backupFile.Close()

	// 复制文件内容
	if _, err := io.Copy(backupFile, sourceFile); err != nil {
		return fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 记录备份信息到结果元数据
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["remote_backup_path"] = backupPath
	result.Metadata["remote_backup_created"] = true

	return nil
}

// executeUpload 执行文件上传
func (c *sftpClient) executeUpload(ctx context.Context, localFile *os.File, localInfo os.FileInfo, localPath, remotePath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 创建远程文件
	remoteFile, err := c.sftpClient.Create(remotePath)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("创建远程文件失败: %s", remotePath), err)
	}
	defer remoteFile.Close()
	
	// 准备传输结果
	result := &types.TransferResult{
		OperationID:  common.GenerateUniqueID("upload"),
		Type:         types.TransferTypeUpload,
		LocalPath:    localPath,
		RemotePath:   remotePath,
		StartTime:    time.Now(),
		Success:      false,
		Metadata:     make(map[string]interface{}),
	}
	
	// 记录文件名和完整路径信息
	result.Metadata["local_filename"] = filepath.Base(localPath)
	result.Metadata["remote_filename"] = filepath.Base(remotePath)
	result.Metadata["full_local_path"] = localPath
	result.Metadata["full_remote_path"] = remotePath
	result.Metadata["operation_type"] = "single_file_upload"
	result.Metadata["file_size"] = localInfo.Size()
	
	// 执行文件传输
	if err := c.transferFile(ctx, localFile, remoteFile, localInfo.Size(), result, options); err != nil {
		result.Error = err.Error()
		return result, err
	}
	
	// 收集后续操作的错误信息
	var errors []string
	var warnings []string
	
	// 设置文件属性
	if err := c.setRemoteFileAttributes(remoteFile, remotePath, localInfo, result, options); err != nil {
		warnings = append(warnings, fmt.Sprintf("设置文件属性失败: %v", err))
	}
	
	// 检查是否需要备份本地文件（上传时在本地备份）
	if options.BackupDir != "" {
		if err := c.backupLocalFile(localPath, options.BackupDir, result); err != nil {
			// 备份失败视为警告，不影响整体成功状态
			warnings = append(warnings, fmt.Sprintf("备份本地文件失败: %v", err))
			c.reportAttributeError(result.OperationID, "backup_error", "备份本地文件失败", localPath, err)
		}
	}
	
	// 检查是否需要删除源文件
	if c.shouldDeleteSourceAfterTransfer(options) {
		// 关键修复：在删除文件前先关闭文件句柄（Windows系统要求）
		// 关闭传入的文件句柄，确保文件可以被删除
		if localFile != nil {
			if closeErr := localFile.Close(); closeErr != nil {
				// 记录关闭警告，但不影响删除操作
				result.Metadata["close_file_warning"] = fmt.Sprintf("关闭文件句柄时出现警告: %v", closeErr)
			}
		}
		
		if err := c.deleteLocalSourceFile(localPath, result); err != nil {
			// 删除源文件失败是严重错误，影响整体成功状态
			errors = append(errors, fmt.Sprintf("删除本地源文件失败: %v", err))
			c.reportAttributeError(result.OperationID, "source_delete_error", "删除本地源文件失败", localPath, err)
		}
	}
	
	// 构建最终结果
	if len(errors) > 0 {
		result.Success = false
		result.Error = fmt.Sprintf("上传完成但后续操作失败: %s", strings.Join(errors, "; "))
		if len(warnings) > 0 {
			result.Error += fmt.Sprintf(" 警告: %s", strings.Join(warnings, "; "))
		}
	} else {
		result.Success = true
		if len(warnings) > 0 {
			result.Error = fmt.Sprintf("上传成功但有警告: %s", strings.Join(warnings, "; "))
		}
	}
	
	// 在元数据中记录错误和警告信息
	if len(errors) > 0 {
		result.Metadata["post_transfer_errors"] = errors
	}
	if len(warnings) > 0 {
		result.Metadata["post_transfer_warnings"] = warnings
	}
	
	return result, nil
}

// executeDownload 执行文件下载
func (c *sftpClient) executeDownload(ctx context.Context, remoteFile *sftp.File, remoteInfo os.FileInfo, remotePath, localPath string, options *configs.SFTPTransferOptions) (*types.TransferResult, error) {
	// 创建本地文件
	localFile, err := os.Create(localPath)
	if err != nil {
		return nil, common.NewIOError(fmt.Sprintf("创建本地文件失败: %s", localPath), err)
	}
	defer localFile.Close()
	
	// 准备传输结果
	result := &types.TransferResult{
		OperationID:  common.GenerateUniqueID("download"),
		Type:         types.TransferTypeDownload,
		LocalPath:    localPath,
		RemotePath:   remotePath,
		StartTime:    time.Now(),
		Success:      false,
		Metadata:     make(map[string]interface{}),
	}
	
	// 记录文件名和完整路径信息
	result.Metadata["local_filename"] = filepath.Base(localPath)
	result.Metadata["remote_filename"] = filepath.Base(remotePath)
	result.Metadata["full_local_path"] = localPath
	result.Metadata["full_remote_path"] = remotePath
	result.Metadata["operation_type"] = "single_file_download"
	result.Metadata["file_size"] = remoteInfo.Size()
	
	// 执行文件传输
	if err := c.transferFile(ctx, remoteFile, localFile, remoteInfo.Size(), result, options); err != nil {
		result.Error = err.Error()
		return result, err
	}
	
	// 收集后续操作的错误信息
	var errors []string
	var warnings []string
	
	// 设置文件属性
	if err := c.setLocalFileAttributes(localPath, remoteInfo, result, options); err != nil {
		warnings = append(warnings, fmt.Sprintf("设置文件属性失败: %v", err))
	}
	
	// 检查是否需要备份远程文件（下载时在SFTP服务器备份）
	if options.BackupDir != "" {
		if err := c.backupRemoteFile(remotePath, options.BackupDir, result); err != nil {
			// 备份失败视为警告，不影响整体成功状态
			warnings = append(warnings, fmt.Sprintf("备份远程文件失败: %v", err))
			c.reportAttributeError(result.OperationID, "backup_error", "备份远程文件失败", remotePath, err)
		}
	}
	
	// 检查是否需要删除源文件
	if c.shouldDeleteSourceAfterTransfer(options) {
		// 关键修复：在删除远程文件前先关闭文件句柄
		// 关闭传入的远程文件句柄，确保文件可以被删除
		if remoteFile != nil {
			if closeErr := remoteFile.Close(); closeErr != nil {
				// 记录关闭警告，但不影响删除操作
				result.Metadata["close_remote_file_warning"] = fmt.Sprintf("关闭远程文件句柄时出现警告: %v", closeErr)
			}
		}
		
		if err := c.deleteRemoteSourceFile(remotePath, result); err != nil {
			// 删除源文件失败是严重错误，影响整体成功状态
			errors = append(errors, fmt.Sprintf("删除远程源文件失败: %v", err))
			c.reportAttributeError(result.OperationID, "source_delete_error", "删除远程源文件失败", remotePath, err)
		}
	}
	
	// 构建最终结果
	if len(errors) > 0 {
		result.Success = false
		result.Error = fmt.Sprintf("下载完成但后续操作失败: %s", strings.Join(errors, "; "))
		if len(warnings) > 0 {
			result.Error += fmt.Sprintf(" 警告: %s", strings.Join(warnings, "; "))
		}
	} else {
		result.Success = true
		if len(warnings) > 0 {
			result.Error = fmt.Sprintf("下载成功但有警告: %s", strings.Join(warnings, "; "))
		}
	}
	
	// 在元数据中记录错误和警告信息
	if len(errors) > 0 {
		result.Metadata["post_transfer_errors"] = errors
	}
	if len(warnings) > 0 {
		result.Metadata["post_transfer_warnings"] = warnings
	}
	
	return result, nil
}

// transferFile 执行实际的文件传输
func (c *sftpClient) transferFile(ctx context.Context, src io.Reader, dst io.Writer, totalSize int64, result *types.TransferResult, options *configs.SFTPTransferOptions) error {
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
func (c *sftpClient) setRemoteFileAttributes(remoteFile *sftp.File, remotePath string, localInfo os.FileInfo, result *types.TransferResult, options *configs.SFTPTransferOptions) error {
	var errors []string
	
	// 设置文件权限
	if options.PreservePermissions {
		if err := remoteFile.Chmod(localInfo.Mode()); err != nil {
			errors = append(errors, fmt.Sprintf("设置文件权限失败: %v", err))
			c.reportAttributeError(result.OperationID, "permission_error", "设置文件权限失败", remotePath, err)
		}
	}
	
	// 设置文件时间戳
	if options.PreserveTimestamps {
		if err := c.sftpClient.Chtimes(remotePath, time.Now(), localInfo.ModTime()); err != nil {
			errors = append(errors, fmt.Sprintf("设置文件时间戳失败: %v", err))
			c.reportAttributeError(result.OperationID, "timestamp_error", "设置文件时间戳失败", remotePath, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

// setLocalFileAttributes 设置本地文件属性
func (c *sftpClient) setLocalFileAttributes(localPath string, remoteInfo os.FileInfo, result *types.TransferResult, options *configs.SFTPTransferOptions) error {
	var errors []string
	
	// 设置文件权限
	if options.PreservePermissions {
		if err := os.Chmod(localPath, remoteInfo.Mode()); err != nil {
			errors = append(errors, fmt.Sprintf("设置文件权限失败: %v", err))
			c.reportAttributeError(result.OperationID, "permission_error", "设置文件权限失败", localPath, err)
		}
	}
	
	// 设置文件时间戳
	if options.PreserveTimestamps {
		if err := os.Chtimes(localPath, time.Now(), remoteInfo.ModTime()); err != nil {
			errors = append(errors, fmt.Sprintf("设置文件时间戳失败: %v", err))
			c.reportAttributeError(result.OperationID, "timestamp_error", "设置文件时间戳失败", localPath, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf(strings.Join(errors, "; "))
	}
	return nil
}

// reportAttributeError 报告属性设置错误
func (c *sftpClient) reportAttributeError(operationID, errorType, message, filePath string, err error) {
	if c.errorCallback != nil {
		c.errorCallback(&types.TransferError{
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

// shouldDeleteSourceAfterTransfer 检查是否应该在传输后删除源文件
// 优先使用传入的options，如果为nil或未设置，则使用client配置的默认选项
func (c *sftpClient) shouldDeleteSourceAfterTransfer(options *configs.SFTPTransferOptions) bool {
	// 如果传入的options为nil，使用客户端的默认配置
	if options == nil {
		if c.config != nil && c.config.DefaultTransferOptions != nil {
			return c.config.DefaultTransferOptions.DeleteSourceAfterTransfer
		}
		return false
	}
	
	// 使用传入的options中的设置
	return options.DeleteSourceAfterTransfer
}

// deleteLocalSourceFile 删除本地源文件
func (c *sftpClient) deleteLocalSourceFile(localPath string, result *types.TransferResult) error {
	err := os.Remove(localPath)
	if err != nil {
		return fmt.Errorf("删除本地源文件失败: %w", err)
	}
	
	// 在结果元数据中记录删除操作
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["source_file_deleted"] = true
	result.Metadata["deleted_file_path"] = localPath
	result.Metadata["delete_operation"] = "local_file"
	
	return nil
}

// deleteRemoteSourceFile 删除远程源文件
func (c *sftpClient) deleteRemoteSourceFile(remotePath string, result *types.TransferResult) error {
	err := c.sftpClient.Remove(remotePath)
	if err != nil {
		return fmt.Errorf("删除远程源文件失败: %w", err)
	}
	
	// 在结果元数据中记录删除操作
	if result.Metadata == nil {
		result.Metadata = make(map[string]interface{})
	}
	result.Metadata["source_file_deleted"] = true
	result.Metadata["deleted_file_path"] = remotePath
	result.Metadata["delete_operation"] = "remote_file"
	
	return nil
}

// createRemoteDirectorySafe 安全地创建远程目录
// 处理路径特殊字符，支持递归创建多层目录
func (c *sftpClient) createRemoteDirectorySafe(remotePath string) error {
	if remotePath == "" || remotePath == "/" || remotePath == "." {
		return nil // 根目录或空路径，无需创建
	}
	
	// 规范化路径，使用正斜杠
	cleanPath := filepath.ToSlash(filepath.Clean(remotePath))
	
	// 检查目录是否已存在
	if _, err := c.sftpClient.Stat(cleanPath); err == nil {
		return nil // 目录已存在
	}
	
	// 递归创建父目录
	parentDir := filepath.Dir(cleanPath)
	if parentDir != cleanPath && parentDir != "/" && parentDir != "." {
		if err := c.createRemoteDirectorySafe(parentDir); err != nil {
			return fmt.Errorf("创建父目录失败 %s: %w", parentDir, err)
		}
	}
	
	// 创建当前目录
	if err := c.sftpClient.Mkdir(cleanPath); err != nil {
		// 检查是否因为目录已存在而失败（并发创建情况）
		if _, statErr := c.sftpClient.Stat(cleanPath); statErr == nil {
			return nil // 目录已存在，忽略错误
		}
		return fmt.Errorf("创建目录失败 %s: %w", cleanPath, err)
	}
	
	return nil
}

// buildSafeRemotePath function has been moved to operations.go

 