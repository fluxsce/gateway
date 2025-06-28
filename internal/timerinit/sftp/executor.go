package sftp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"gohub/pkg/logger"
	"gohub/pkg/plugin/tools/configs"
	"gohub/pkg/plugin/tools/interfaces"
	"gohub/pkg/timer"
)

// SFTPTaskExecutor SFTP任务执行器
// 实现timer.TaskExecutor接口，负责执行SFTP相关的定时任务
// 支持文件上传、下载、目录列表、文件删除、目录创建等操作
// 所有SFTP相关配置统一从configParameters获取，包括操作类型和操作参数
type SFTPTaskExecutor struct {
	taskId           string                 // 任务ID，用于标识唯一的定时任务
	taskName         string                 // 任务名称，用于日志记录和调试
	sftpClient       interfaces.SFTPTool    // SFTP客户端接口，负责实际的文件传输操作
	config           *configs.SFTPConfig    // SFTP连接配置，包含主机、端口、认证等信息
	configParameters map[string]interface{} // 配置参数映射，包含操作类型和所有操作相关参数
	                                        // 支持的参数包括：
	                                        // - operation/operationType: 操作类型(upload/download/list/delete/mkdir/sync)
	                                        // - localPath: 本地文件/目录路径
	                                        // - remotePath: 远程文件/目录路径
	                                        // - isDirectory: 是否为目录操作
	                                        // - recursive: 是否递归操作
	                                        // - disconnectAfterExecution: 执行完成后是否断开连接
}

// Execute 执行SFTP任务
// 实现timer.TaskExecutor接口的Execute方法
// 根据配置参数中的操作类型执行相应的SFTP操作
// 参数:
//   ctx: 上下文对象，用于控制任务执行生命周期和取消操作
//   params: 运行时参数，可以覆盖配置参数中的设置，支持map[string]interface{}或JSON字符串格式
// 返回:
//   *timer.ExecuteResult: 执行结果，包含成功状态、数据和消息
//   error: 执行过程中发生的错误
func (e *SFTPTaskExecutor) Execute(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 获取操作类型
	operation := e.getOperationType()
	
	logger.Info("开始执行SFTP任务",
		"taskId", e.taskId,
		"taskName", e.taskName,
		"operation", operation)

	// 确保SFTP客户端已连接
	if err := e.ensureConnected(ctx); err != nil {
		return e.createFailureResult("连接SFTP服务器失败", err), err
	}

	// 执行完成后检查是否需要断开连接（对于一次性任务）
	defer func() {
		// 如果配置了自动断开或者任务完成后断开，则关闭连接
		if e.shouldDisconnectAfterExecution() {
			if err := e.sftpClient.Disconnect(); err != nil {
				logger.Warn("断开SFTP连接失败", 
					"taskId", e.taskId, 
					"error", err)
			} else {
				logger.Debug("SFTP连接已断开", "taskId", e.taskId)
			}
		}
	}()

	// 根据操作类型执行相应的操作
	switch strings.ToLower(operation) {
	case "upload":
		return e.executeUpload(ctx, params)
	case "download":
		return e.executeDownload(ctx, params)
	case "sync":
		return e.executeSync(ctx, params)
	case "list":
		return e.executeList(ctx, params)
	case "delete":
		return e.executeDelete(ctx, params)
	case "mkdir":
		return e.executeMkdir(ctx, params)
	default:
		err := fmt.Errorf("不支持的操作类型: %s", operation)
		return e.createFailureResult("不支持的操作类型", err), err
	}
}

// GetName 获取任务执行器名称
// 实现timer.TaskExecutor接口的GetName方法
// 返回格式化的执行器名称，包含任务ID用于识别
// 返回:
//   string: 执行器名称，格式为"SFTPTaskExecutor-{taskId}"
func (e *SFTPTaskExecutor) GetName() string {
	return fmt.Sprintf("SFTPTaskExecutor-%s", e.taskId)
}

// GetSFTPClient 获取SFTP客户端实例
// 用于外部访问SFTP客户端，比如进行连接测试
// 返回:
//   interfaces.SFTPTool: SFTP客户端接口实例
func (e *SFTPTaskExecutor) GetSFTPClient() interfaces.SFTPTool {
	return e.sftpClient
}

// getOperationType 从配置参数中获取操作类型
// 支持"operation"和"operationType"两种字段名，提供向后兼容性
// 返回:
//   string: 操作类型字符串，如"upload"、"download"等，如果未配置则返回空字符串
func (e *SFTPTaskExecutor) getOperationType() string {
	if e.configParameters == nil {
		return ""
	}
	
	// 优先使用"operation"字段
	if operation, ok := e.configParameters["defaultOperation"].(string); ok {
		return operation
	}
	
	// 兼容性：也检查operationType字段
	if operationType, ok := e.configParameters["operationType"].(string); ok {
		return operationType
	}
	
	return ""
}

// getConfigValue 从配置参数中获取指定键的值
// 提供统一的配置参数访问接口，支持类型安全的参数获取
// 参数:
//   key: 配置参数的键名
// 返回:
//   interface{}: 配置参数的值，如果不存在则返回nil
func (e *SFTPTaskExecutor) getConfigValue(key string) interface{} {
	if e.configParameters == nil {
		return nil
	}
	return e.configParameters[key]
}

// getConfigString 从配置参数中获取字符串值
// 提供类型安全的字符串参数获取方法
// 参数:
//   key: 配置参数的键名
// 返回:
//   string: 字符串值，如果不存在或类型不匹配则返回空字符串
func (e *SFTPTaskExecutor) getConfigString(key string) string {
	if value := e.getConfigValue(key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// getConfigBool 从配置参数中获取布尔值
// 提供类型安全的布尔参数获取方法
// 参数:
//   key: 配置参数的键名
// 返回:
//   bool: 布尔值，如果不存在或类型不匹配则返回false
func (e *SFTPTaskExecutor) getConfigBool(key string) bool {
	if value := e.getConfigValue(key); value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// ensureConnected 确保SFTP客户端已连接
// 检查连接状态，如果未连接则建立新连接
// 参数:
//   ctx: 上下文对象，用于控制连接超时
// 返回:
//   error: 连接过程中的错误，连接成功返回nil
func (e *SFTPTaskExecutor) ensureConnected(ctx context.Context) error {
	if e.sftpClient.IsConnected() {
		return nil
	}

	logger.Info("连接SFTP服务器", "taskId", e.taskId, "host", e.config.Host)
	return e.sftpClient.Connect(ctx)
}

// executeUpload 执行文件上传操作
// 支持单文件和目录上传，根据参数自动选择上传方式
// 参数:
//   ctx: 上下文对象，用于控制上传过程
//   params: 运行时参数，可以覆盖配置中的上传参数
// 返回:
//   *timer.ExecuteResult: 上传结果
//   error: 上传过程中的错误
func (e *SFTPTaskExecutor) executeUpload(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 解析上传参数
	uploadParams, err := e.parseUploadParams(params)
	if err != nil {
		return e.createFailureResult("解析上传参数失败", err), err
	}

	localPath := uploadParams.LocalPath
	remotePath := uploadParams.RemotePath
	options := uploadParams.Options

	logger.Info("执行文件上传",
		"taskId", e.taskId,
		"localPath", localPath,
		"remotePath", remotePath)

	// 检查是否为目录上传
	if uploadParams.IsDirectory {
		result, err := e.sftpClient.UploadDirectory(ctx, localPath, remotePath, options)
		if err != nil {
			return e.createFailureResult("上传目录失败", err), err
		}
		return e.createSuccessResult("目录上传成功", result), nil
	} else {
		result, err := e.sftpClient.UploadFile(ctx, localPath, remotePath, options)
		if err != nil {
			return e.createFailureResult("上传文件失败", err), err
		}
		return e.createSuccessResult("文件上传成功", result), nil
	}
}

// executeDownload 执行文件下载操作
// 支持单文件和目录下载，根据参数自动选择下载方式
// 参数:
//   ctx: 上下文对象，用于控制下载过程
//   params: 运行时参数，可以覆盖配置中的下载参数
// 返回:
//   *timer.ExecuteResult: 下载结果
//   error: 下载过程中的错误
func (e *SFTPTaskExecutor) executeDownload(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 解析下载参数
	downloadParams, err := e.parseDownloadParams(params)
	if err != nil {
		return e.createFailureResult("解析下载参数失败", err), err
	}

	remotePath := downloadParams.RemotePath
	localPath := downloadParams.LocalPath
	options := downloadParams.Options

	logger.Info("执行文件下载",
		"taskId", e.taskId,
		"remotePath", remotePath,
		"localPath", localPath)

	// 检查是否为目录下载
	if downloadParams.IsDirectory {
		result, err := e.sftpClient.DownloadDirectory(ctx, remotePath, localPath, options)
		if err != nil {
			return e.createFailureResult("下载目录失败", err), err
		}
		return e.createSuccessResult("目录下载成功", result), nil
	} else {
		result, err := e.sftpClient.DownloadFile(ctx, remotePath, localPath, options)
		if err != nil {
			return e.createFailureResult("下载文件失败", err), err
		}
		return e.createSuccessResult("文件下载成功", result), nil
	}
}

// executeSync 执行目录同步操作
// 目前尚未实现，预留接口用于未来扩展
// 参数:
//   ctx: 上下文对象
//   params: 同步参数
// 返回:
//   *timer.ExecuteResult: 同步结果
//   error: 同步过程中的错误
func (e *SFTPTaskExecutor) executeSync(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// TODO: 实现目录同步功能
	err := fmt.Errorf("目录同步功能尚未实现")
	return e.createFailureResult("目录同步功能尚未实现", err), err
}

// executeList 执行目录列表操作
// 列出指定远程目录的文件和子目录信息
// 参数:
//   ctx: 上下文对象，用于控制列表操作
//   params: 运行时参数，可以指定要列出的目录路径
// 返回:
//   *timer.ExecuteResult: 列表结果，包含文件信息数组
//   error: 列表过程中的错误
func (e *SFTPTaskExecutor) executeList(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 解析列表参数
	listParams, err := e.parseListParams(params)
	if err != nil {
		return e.createFailureResult("解析列表参数失败", err), err
	}

	remotePath := listParams.RemotePath

	logger.Info("执行目录列表",
		"taskId", e.taskId,
		"remotePath", remotePath)

	fileList, err := e.sftpClient.ListDirectory(ctx, remotePath)
	if err != nil {
		return e.createFailureResult("列出目录内容失败", err), err
	}

	return e.createSuccessResult("列出目录内容成功", map[string]interface{}{
		"path":  remotePath,
		"files": fileList,
		"count": len(fileList),
	}), nil
}

// executeDelete 执行文件删除操作
// 支持单文件删除和目录删除（可选递归删除）
// 参数:
//   ctx: 上下文对象，用于控制删除过程
//   params: 运行时参数，可以指定要删除的文件或目录路径
// 返回:
//   *timer.ExecuteResult: 删除结果
//   error: 删除过程中的错误
func (e *SFTPTaskExecutor) executeDelete(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 解析删除参数
	deleteParams, err := e.parseDeleteParams(params)
	if err != nil {
		return e.createFailureResult("解析删除参数失败", err), err
	}

	remotePath := deleteParams.RemotePath

	logger.Info("执行文件删除",
		"taskId", e.taskId,
		"remotePath", remotePath)

	// 检查是否为目录删除
	if deleteParams.IsDirectory {
		err := e.sftpClient.RemoveDirectory(ctx, remotePath, deleteParams.Recursive)
		if err != nil {
			return e.createFailureResult("删除目录失败", err), err
		}
		return e.createSuccessResult("目录删除成功", map[string]interface{}{
			"path": remotePath,
			"type": "directory",
		}), nil
	} else {
		err := e.sftpClient.RemoveFile(ctx, remotePath)
		if err != nil {
			return e.createFailureResult("删除文件失败", err), err
		}
		return e.createSuccessResult("文件删除成功", map[string]interface{}{
			"path": remotePath,
			"type": "file",
		}), nil
	}
}

// executeMkdir 执行目录创建操作
// 支持单级目录创建和递归目录创建
// 参数:
//   ctx: 上下文对象，用于控制创建过程
//   params: 运行时参数，可以指定要创建的目录路径和递归选项
// 返回:
//   *timer.ExecuteResult: 创建结果
//   error: 创建过程中的错误
func (e *SFTPTaskExecutor) executeMkdir(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	// 解析创建目录参数
	mkdirParams, err := e.parseMkdirParams(params)
	if err != nil {
		return e.createFailureResult("解析创建目录参数失败", err), err
	}

	remotePath := mkdirParams.RemotePath

	logger.Info("执行目录创建",
		"taskId", e.taskId,
		"remotePath", remotePath)

	err = e.sftpClient.CreateDirectory(ctx, remotePath, mkdirParams.Recursive)
	if err != nil {
		return e.createFailureResult("创建目录失败", err), err
	}

	return e.createSuccessResult("目录创建成功", map[string]interface{}{
		"path":      remotePath,
		"recursive": mkdirParams.Recursive,
	}), nil
}

// 参数结构定义

// UploadParams 上传参数结构
// 包含文件上传操作所需的所有参数
type UploadParams struct {
	LocalPath   string                        `json:"localPath"`   // 本地文件或目录路径
	RemotePath  string                        `json:"remotePath"`  // 远程目标路径
	IsDirectory bool                          `json:"isDirectory"` // 是否为目录上传
	Options     *configs.SFTPTransferOptions  `json:"options"`      // 传输选项配置
}

// DownloadParams 下载参数结构
// 包含文件下载操作所需的所有参数
type DownloadParams struct {
	RemotePath  string                        `json:"remotePath"`  // 远程文件或目录路径
	LocalPath   string                        `json:"localPath"`   // 本地目标路径
	IsDirectory bool                          `json:"isDirectory"` // 是否为目录下载
	Options     *configs.SFTPTransferOptions  `json:"options"`      // 传输选项配置
}

// ListParams 列表参数结构
// 包含目录列表操作所需的参数
type ListParams struct {
	RemotePath string `json:"remotePath"` // 要列出的远程目录路径
}

// DeleteParams 删除参数结构
// 包含文件/目录删除操作所需的参数
type DeleteParams struct {
	RemotePath  string `json:"remotePath"`  // 要删除的远程文件或目录路径
	IsDirectory bool   `json:"isDirectory"` // 是否为目录删除
	Recursive   bool   `json:"recursive"`    // 是否递归删除（仅对目录有效）
}

// MkdirParams 创建目录参数结构
// 包含目录创建操作所需的参数
type MkdirParams struct {
	RemotePath string `json:"remotePath"` // 要创建的远程目录路径
	Recursive  bool   `json:"recursive"`   // 是否递归创建父目录
}

// 参数解析方法

// parseUploadParams 解析上传参数
// 从配置参数和运行时参数中解析上传操作所需的参数
// 运行时参数优先级高于配置参数，可以覆盖配置中的设置
// 参数:
//   params: 运行时参数，支持map[string]interface{}或JSON字符串
// 返回:
//   *UploadParams: 解析后的上传参数
//   error: 解析过程中的错误或参数验证失败
func (e *SFTPTaskExecutor) parseUploadParams(params interface{}) (*UploadParams, error) {
	uploadParams := &UploadParams{
		Options: e.config.DefaultTransferOptions,
	}

	// 从配置参数中获取默认值
	uploadParams.LocalPath = e.getConfigString("localPath")
	uploadParams.RemotePath = e.getConfigString("remotePath")
	uploadParams.IsDirectory = e.getConfigBool("isDirectory")

	// 从运行时参数中获取值（优先级最高）
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if localPath, ok := paramsMap["localPath"].(string); ok && localPath != "" {
				uploadParams.LocalPath = localPath
			}
			if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
				uploadParams.RemotePath = remotePath
			}
			if isDir, ok := paramsMap["isDirectory"].(bool); ok {
				uploadParams.IsDirectory = isDir
			}
		} else if paramsStr, ok := params.(string); ok {
			// 尝试解析JSON字符串
			var paramsMap map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &paramsMap); err == nil {
				if localPath, ok := paramsMap["localPath"].(string); ok && localPath != "" {
					uploadParams.LocalPath = localPath
				}
				if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
					uploadParams.RemotePath = remotePath
				}
				if isDir, ok := paramsMap["isDirectory"].(bool); ok {
					uploadParams.IsDirectory = isDir
				}
			}
		}
	}

	// 验证必要参数
	if uploadParams.LocalPath == "" {
		return nil, fmt.Errorf("localPath参数不能为空")
	}
	if uploadParams.RemotePath == "" {
		return nil, fmt.Errorf("remotePath参数不能为空")
	}

	return uploadParams, nil
}

// parseDownloadParams 解析下载参数
// 从配置参数和运行时参数中解析下载操作所需的参数
// 运行时参数优先级高于配置参数，可以覆盖配置中的设置
// 参数:
//   params: 运行时参数，支持map[string]interface{}或JSON字符串
// 返回:
//   *DownloadParams: 解析后的下载参数
//   error: 解析过程中的错误或参数验证失败
func (e *SFTPTaskExecutor) parseDownloadParams(params interface{}) (*DownloadParams, error) {
	downloadParams := &DownloadParams{
		Options: e.config.DefaultTransferOptions,
	}

	// 从配置参数中获取默认值
	downloadParams.RemotePath = e.getConfigString("remotePath")
	downloadParams.LocalPath = e.getConfigString("localPath")
	downloadParams.IsDirectory = e.getConfigBool("isDirectory")

	// 从运行时参数中获取值
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
				downloadParams.RemotePath = remotePath
			}
			if localPath, ok := paramsMap["localPath"].(string); ok && localPath != "" {
				downloadParams.LocalPath = localPath
			}
			if isDir, ok := paramsMap["isDirectory"].(bool); ok {
				downloadParams.IsDirectory = isDir
			}
		} else if paramsStr, ok := params.(string); ok {
			var paramsMap map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &paramsMap); err == nil {
				if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
					downloadParams.RemotePath = remotePath
				}
				if localPath, ok := paramsMap["localPath"].(string); ok && localPath != "" {
					downloadParams.LocalPath = localPath
				}
				if isDir, ok := paramsMap["isDirectory"].(bool); ok {
					downloadParams.IsDirectory = isDir
				}
			}
		}
	}

	// 验证必要参数
	if downloadParams.RemotePath == "" {
		return nil, fmt.Errorf("remotePath参数不能为空")
	}
	if downloadParams.LocalPath == "" {
		return nil, fmt.Errorf("localPath参数不能为空")
	}

	return downloadParams, nil
}

// parseListParams 解析列表参数
// 从配置参数和运行时参数中解析目录列表操作所需的参数
// 参数:
//   params: 运行时参数，支持map[string]interface{}或JSON字符串
// 返回:
//   *ListParams: 解析后的列表参数
//   error: 解析过程中的错误
func (e *SFTPTaskExecutor) parseListParams(params interface{}) (*ListParams, error) {
	listParams := &ListParams{
		RemotePath: "/", // 默认根目录
	}

	// 从配置参数中获取默认值
	if remotePath := e.getConfigString("remotePath"); remotePath != "" {
		listParams.RemotePath = remotePath
	}

	// 从运行时参数中获取值
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
				listParams.RemotePath = remotePath
			}
		} else if paramsStr, ok := params.(string); ok {
			var paramsMap map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &paramsMap); err == nil {
				if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
					listParams.RemotePath = remotePath
				}
			}
		}
	}

	return listParams, nil
}

// parseDeleteParams 解析删除参数
// 从配置参数和运行时参数中解析删除操作所需的参数
// 参数:
//   params: 运行时参数，支持map[string]interface{}或JSON字符串
// 返回:
//   *DeleteParams: 解析后的删除参数
//   error: 解析过程中的错误或参数验证失败
func (e *SFTPTaskExecutor) parseDeleteParams(params interface{}) (*DeleteParams, error) {
	deleteParams := &DeleteParams{
		Recursive: false,
	}

	// 从配置参数中获取默认值
	deleteParams.RemotePath = e.getConfigString("remotePath")
	deleteParams.IsDirectory = e.getConfigBool("isDirectory")
	deleteParams.Recursive = e.getConfigBool("recursive")

	// 从运行时参数中获取值
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
				deleteParams.RemotePath = remotePath
			}
			if isDir, ok := paramsMap["isDirectory"].(bool); ok {
				deleteParams.IsDirectory = isDir
			}
			if recursive, ok := paramsMap["recursive"].(bool); ok {
				deleteParams.Recursive = recursive
			}
		} else if paramsStr, ok := params.(string); ok {
			var paramsMap map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &paramsMap); err == nil {
				if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
					deleteParams.RemotePath = remotePath
				}
				if isDir, ok := paramsMap["isDirectory"].(bool); ok {
					deleteParams.IsDirectory = isDir
				}
				if recursive, ok := paramsMap["recursive"].(bool); ok {
					deleteParams.Recursive = recursive
				}
			}
		}
	}

	// 验证必要参数
	if deleteParams.RemotePath == "" {
		return nil, fmt.Errorf("remotePath参数不能为空")
	}

	return deleteParams, nil
}

// parseMkdirParams 解析创建目录参数
// 从配置参数和运行时参数中解析目录创建操作所需的参数
// 参数:
//   params: 运行时参数，支持map[string]interface{}或JSON字符串
// 返回:
//   *MkdirParams: 解析后的创建目录参数
//   error: 解析过程中的错误或参数验证失败
func (e *SFTPTaskExecutor) parseMkdirParams(params interface{}) (*MkdirParams, error) {
	mkdirParams := &MkdirParams{
		Recursive: true, // 默认递归创建
	}

	// 从配置参数中获取默认值
	mkdirParams.RemotePath = e.getConfigString("remotePath")
	if recursive := e.getConfigValue("recursive"); recursive != nil {
		if r, ok := recursive.(bool); ok {
			mkdirParams.Recursive = r
		}
	}

	// 从运行时参数中获取值
	if params != nil {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
				mkdirParams.RemotePath = remotePath
			}
			if recursive, ok := paramsMap["recursive"].(bool); ok {
				mkdirParams.Recursive = recursive
			}
		} else if paramsStr, ok := params.(string); ok {
			var paramsMap map[string]interface{}
			if err := json.Unmarshal([]byte(paramsStr), &paramsMap); err == nil {
				if remotePath, ok := paramsMap["remotePath"].(string); ok && remotePath != "" {
					mkdirParams.RemotePath = remotePath
				}
				if recursive, ok := paramsMap["recursive"].(bool); ok {
					mkdirParams.Recursive = recursive
				}
			}
		}
	}

	// 验证必要参数
	if mkdirParams.RemotePath == "" {
		return nil, fmt.Errorf("remotePath参数不能为空")
	}

	return mkdirParams, nil
}

// 结果创建方法

// createSuccessResult 创建成功结果
// 构造标准的成功执行结果，包含任务信息和执行数据
// 参数:
//   message: 成功消息描述
//   data: 执行结果数据，可以是任何类型
// 返回:
//   *timer.ExecuteResult: 成功结果对象
func (e *SFTPTaskExecutor) createSuccessResult(message string, data interface{}) *timer.ExecuteResult {
	return &timer.ExecuteResult{
		Success: true,
		Data:    data,
		Message: message,
		Extra: map[string]interface{}{
			"taskId":    e.taskId,
			"taskName":  e.taskName,
			"operation": e.getOperationType(),
		},
	}
}

// createFailureResult 创建失败结果
// 构造标准的失败执行结果，包含错误信息和任务上下文
// 参数:
//   message: 失败消息描述
//   err: 具体的错误对象，可以为nil
// 返回:
//   *timer.ExecuteResult: 失败结果对象
func (e *SFTPTaskExecutor) createFailureResult(message string, err error) *timer.ExecuteResult {
	errMsg := ""
	if err != nil {
		errMsg = err.Error()
	}

	return &timer.ExecuteResult{
		Success: false,
		Data:    nil,
		Message: fmt.Sprintf("%s: %s", message, errMsg),
		Extra: map[string]interface{}{
			"taskId":    e.taskId,
			"taskName":  e.taskName,
			"operation": e.getOperationType(),
			"error":     errMsg,
		},
	}
}

// shouldDisconnectAfterExecution 判断是否在执行完成后断开连接
// 根据配置决定是否在任务执行完成后断开SFTP连接
// 短连接模式适用于偶尔执行的任务，长连接模式适用于频繁执行的任务
// 返回:
//   bool: true表示需要断开连接，false表示保持连接
func (e *SFTPTaskExecutor) shouldDisconnectAfterExecution() bool {
	// 如果配置了不自动重连，说明是短连接模式，执行完应该断开
	if e.config != nil && !e.config.AutoReconnect {
		return true
	}
	
	// 检查配置参数中是否设置了disconnectAfterExecution
	if disconnect := e.getConfigBool("disconnectAfterExecution"); disconnect {
		return true
	}
	
	// 默认不断开连接，保持长连接以提高效率
	return false
}

// Close 关闭执行器并释放资源
// 实现timer.TaskExecutor接口的Close方法
// 用于清理SFTP连接等资源，通常在任务执行器生命周期结束时调用
// 返回:
//   error: 清理过程中的错误，成功返回nil
func (e *SFTPTaskExecutor) Close() error {
	if e.sftpClient != nil && e.sftpClient.IsConnected() {
		logger.Info("关闭SFTP执行器资源", "taskId", e.taskId)
		return e.sftpClient.Close()
	}
	return nil
} 