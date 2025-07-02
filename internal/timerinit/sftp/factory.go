package sftp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gohub/internal/timerinit/common/dao"
	"gohub/internal/timerinit/common/taskinit"
	"gohub/internal/types/timertypes"
	"gohub/internal/types/tooltypes"
	"gohub/pkg/database"
	"gohub/pkg/logger"
	"gohub/pkg/plugin/tools/configs"
	sftpTool "gohub/pkg/plugin/tools/sftp"
	"gohub/pkg/timer"
	"gohub/pkg/utils/crypto"
)

// SFTPExecutorFactory SFTP任务执行器工厂
// 实现TaskExecutorFactory接口，专注于SFTP特定的逻辑
// 将SFTP配置与定时器任务配置分离，便于独立维护和管理
type SFTPExecutorFactory struct {
	daoManager *dao.DAOManager
}

// NewSFTPExecutorFactory 创建SFTP执行器工厂实例
// 参数:
//   daoManager: 数据访问对象管理器，用于查询工具配置
// 返回:
//   taskinit.TaskExecutorFactory: 任务执行器工厂接口实现
func NewSFTPExecutorFactory(daoManager *dao.DAOManager) taskinit.TaskExecutorFactory {
	return &SFTPExecutorFactory{
		daoManager: daoManager,
	}
}

// GetExecutorType 获取执行器类型
// 返回SFTP传输执行器的类型标识，用于任务分类和路由
// 返回:
//   string: 执行器类型标识
func (f *SFTPExecutorFactory) GetExecutorType() string {
	return "SFTP_TRANSFER"
}

// CreateExecutor 创建SFTP任务执行器
// 根据任务配置创建SFTP执行器实例，所有SFTP相关配置仅从toolConfig获取
// 这种设计将SFTP配置与定时器任务配置解耦，便于独立维护
// 参数:
//   ctx: 上下文对象，用于控制请求生命周期
//   task: 定时器任务配置，仅用于获取基本信息和工具配置ID
// 返回:
//   timer.TaskExecutor: 创建的SFTP任务执行器
//   error: 创建过程中的错误信息
func (f *SFTPExecutorFactory) CreateExecutor(ctx context.Context, task *timertypes.TimerTask) (timer.TaskExecutor, error) {
	// 验证必要的任务信息
	if task.ToolConfigId == "" {
		return nil, fmt.Errorf("任务 %s 缺少工具配置ID", task.TaskId)
	}

	// 获取工具配置 - SFTP相关配置的唯一来源
	toolConfig, err := f.getToolConfig(ctx, task.TenantId, task.ToolConfigId)
	if err != nil {
		return nil, fmt.Errorf("获取工具配置失败: %w", err)
	}

	// 验证工具配置类型
	if toolConfig.ToolType != "SFTP_TRANSFER" {
		return nil, fmt.Errorf("工具配置类型不匹配，期望SFTP，实际为: %s", toolConfig.ToolType)
	}

	// 创建SFTP配置 - 仅依赖toolConfig
	sftpConfig, err := f.createSFTPConfig(toolConfig)
	if err != nil {
		return nil, fmt.Errorf("创建SFTP配置失败: %w", err)
	}

	// 创建SFTP客户端
	sftpClient, err := sftpTool.NewClient(sftpConfig)
	if err != nil {
		return nil, fmt.Errorf("创建SFTP客户端失败: %w", err)
	}

	// 解析配置参数
	var configParameters map[string]interface{}
	if toolConfig.ConfigParameters != nil && *toolConfig.ConfigParameters != "" {
		if err := json.Unmarshal([]byte(*toolConfig.ConfigParameters), &configParameters); err != nil {
			return nil, fmt.Errorf("解析工具配置参数失败: %w", err)
		}
	}

	// 创建并返回SFTP任务执行器
	// 注意：这里只使用任务的基本标识信息，SFTP操作配置从toolConfig获取
	executor := &SFTPTaskExecutor{
		taskId:           task.TaskId,
		taskName:         task.TaskName,
		sftpClient:       sftpClient,
		config:           sftpConfig,
		configParameters: configParameters,
	}

	logger.Info("SFTP任务执行器创建成功", 
		"taskId", task.TaskId,
		"taskName", task.TaskName,
		"toolConfigId", task.ToolConfigId,
		"sftpHost", sftpConfig.Host,
		"sftpPort", sftpConfig.Port,
		"hasConfigParams", configParameters != nil)

	return executor, nil
}

// getToolConfig 获取工具配置
// 根据租户ID和工具配置ID从数据库查询工具配置信息
// 参数:
//   ctx: 上下文对象，用于控制请求生命周期
//   tenantId: 租户ID，用于多租户隔离
//   toolConfigId: 工具配置ID，唯一标识工具配置
// 返回:
//   *tooltypes.ToolConfig: 工具配置对象
//   error: 查询过程中的错误信息
func (f *SFTPExecutorFactory) getToolConfig(ctx context.Context, tenantId, toolConfigId string) (*tooltypes.ToolConfig, error) {
	if tenantId == "" {
		return nil, fmt.Errorf("租户ID不能为空")
	}
	if toolConfigId == "" {
		return nil, fmt.Errorf("工具配置ID不能为空")
	}

	toolConfig, err := f.daoManager.GetToolConfigDAO().GetConfigById(ctx, tenantId, toolConfigId)
	if err != nil {
		return nil, fmt.Errorf("查询工具配置失败: %w", err)
	}

	if toolConfig == nil {
		return nil, fmt.Errorf("未找到工具配置: tenantId=%s, toolConfigId=%s", tenantId, toolConfigId)
	}

	return toolConfig, nil
}

// createSFTPConfig 创建SFTP配置
// 将数据库中的工具配置转换为SFTP客户端所需的配置格式
// 所有SFTP相关配置都从toolConfig获取，确保配置的一致性和可维护性
// 参数:
//   toolConfig: 工具配置对象，包含SFTP连接和认证信息
// 返回:
//   *configs.SFTPConfig: SFTP客户端配置对象
//   error: 转换过程中的错误信息
func (f *SFTPExecutorFactory) createSFTPConfig(toolConfig *tooltypes.ToolConfig) (*configs.SFTPConfig, error) {
	// 创建默认SFTP配置
	sftpConfig := configs.DefaultSFTPConfig()
	
	// 设置基础信息
	sftpConfig.ID = toolConfig.ToolConfigId
	sftpConfig.Name = toolConfig.ConfigName
	if toolConfig.ConfigDescription != nil {
		sftpConfig.Description = *toolConfig.ConfigDescription
	}

	// 设置连接信息 - 必需字段验证
	if toolConfig.HostAddress == nil || *toolConfig.HostAddress == "" {
		return nil, fmt.Errorf("SFTP主机地址不能为空")
	}
	sftpConfig.Host = *toolConfig.HostAddress

	if toolConfig.PortNumber != nil {
		sftpConfig.Port = *toolConfig.PortNumber
	} else {
		sftpConfig.Port = 22 // SFTP默认端口
	}

	if toolConfig.UserName == nil || *toolConfig.UserName == "" {
		return nil, fmt.Errorf("SFTP用户名不能为空")
	}
	sftpConfig.Username = *toolConfig.UserName

	// 设置认证信息
	if err := f.configureSFTPAuthentication(sftpConfig, toolConfig); err != nil {
		return nil, fmt.Errorf("配置SFTP认证失败: %w", err)
	}

	// 解析自定义配置参数
	if toolConfig.ConfigParameters != nil && *toolConfig.ConfigParameters != "" {
		if err := f.applySFTPCustomConfig(sftpConfig, *toolConfig.ConfigParameters); err != nil {
			return nil, fmt.Errorf("应用自定义SFTP配置失败: %w", err)
		}
	}

	logger.Info("SFTP配置创建成功", 
		"configId", sftpConfig.ID,
		"host", sftpConfig.Host,
		"port", sftpConfig.Port,
		"username", sftpConfig.Username,
		"authType", f.getAuthTypeName(toolConfig))

	return sftpConfig, nil
}

// configureSFTPAuthentication 配置SFTP认证信息
// 根据工具配置中的认证类型设置相应的认证方式
// 参数:
//   sftpConfig: SFTP配置对象
//   toolConfig: 工具配置对象，包含认证信息
// 返回:
//   error: 配置过程中的错误信息
func (f *SFTPExecutorFactory) configureSFTPAuthentication(sftpConfig *configs.SFTPConfig, toolConfig *tooltypes.ToolConfig) error {
	if toolConfig.AuthType == nil {
		return fmt.Errorf("认证类型不能为空")
	}

	switch *toolConfig.AuthType {
	case tooltypes.AuthTypePassword:
		// 密码认证
		if toolConfig.PasswordEncrypted == nil || *toolConfig.PasswordEncrypted == "" {
			return fmt.Errorf("密码认证要求提供密码")
		}
		
		// 解密密码
		decryptedPassword := *toolConfig.PasswordEncrypted
		if crypto.IsEncryptedString(*toolConfig.PasswordEncrypted) {
			var err error
			decryptedPassword, err = crypto.DecryptString(*toolConfig.PasswordEncrypted)
			if err != nil {
				logger.Error("SFTP密码解密失败", "error", err, "configId", toolConfig.ToolConfigId)
				return fmt.Errorf("密码解密失败: %w", err)
			}
			logger.Debug("SFTP密码解密成功", "configId", toolConfig.ToolConfigId)
		} else {
			logger.Debug("SFTP密码未加密，直接使用", "configId", toolConfig.ToolConfigId)
		}
		
		sftpConfig.PasswordAuth = &configs.PasswordAuthConfig{
			Password: decryptedPassword,
		}
		
	case tooltypes.AuthTypePublicKey:
		// 公钥认证
		if toolConfig.KeyFileContent == nil && toolConfig.KeyFilePath == nil {
			return fmt.Errorf("公钥认证要求提供私钥文件内容或路径")
		}
		sftpConfig.PublicKeyAuth = &configs.PublicKeyAuthConfig{}
		if toolConfig.KeyFilePath != nil {
			sftpConfig.PublicKeyAuth.PrivateKeyPath = *toolConfig.KeyFilePath
		}
		if toolConfig.KeyFileContent != nil {
			// 解密私钥内容
			decryptedKeyContent := *toolConfig.KeyFileContent
			if crypto.IsEncryptedString(*toolConfig.KeyFileContent) {
				var err error
				decryptedKeyContent, err = crypto.DecryptString(*toolConfig.KeyFileContent)
				if err != nil {
					logger.Error("SFTP私钥内容解密失败", "error", err, "configId", toolConfig.ToolConfigId)
					return fmt.Errorf("私钥内容解密失败: %w", err)
				}
				logger.Debug("SFTP私钥内容解密成功", "configId", toolConfig.ToolConfigId)
			} else {
				logger.Debug("SFTP私钥内容未加密，直接使用", "configId", toolConfig.ToolConfigId)
			}
			sftpConfig.PublicKeyAuth.PrivateKeyData = []byte(decryptedKeyContent)
		}
		
	default:
		return fmt.Errorf("不支持的认证类型: %v", *toolConfig.AuthType)
	}

	return nil
}

// applySFTPCustomConfig 应用自定义SFTP配置
// 直接JSON映射，简洁高效
// 参数:
//   sftpConfig: SFTP配置对象
//   configParams: 自定义配置参数JSON字符串
// 返回:
//   error: 应用过程中的错误信息
func (f *SFTPExecutorFactory) applySFTPCustomConfig(sftpConfig *configs.SFTPConfig, configParams string) error {
	// 解析为通用map
	var customParams map[string]interface{}
	if err := json.Unmarshal([]byte(configParams), &customParams); err != nil {
		return fmt.Errorf("解析自定义配置参数失败: %w", err)
	}

	// 确保默认传输选项存在
	if sftpConfig.DefaultTransferOptions == nil {
		sftpConfig.DefaultTransferOptions = configs.DefaultSFTPTransferOptions()
	}

	// 直接映射 - 使用反射自动处理所有字段
	f.mapConfigFields(sftpConfig, customParams)
	f.mapTransferOptions(sftpConfig.DefaultTransferOptions, customParams)

	logger.Debug("应用自定义SFTP配置完成", 
		"configId", sftpConfig.ID,
		"customParamsCount", len(customParams))

	return nil
}

// mapConfigFields 自动映射配置字段
func (f *SFTPExecutorFactory) mapConfigFields(config *configs.SFTPConfig, params map[string]interface{}) {
	// 连接超时相关（需要特殊处理的Duration类型）
	if val, ok := params["connectTimeout"]; ok {
		if timeout, ok := val.(float64); ok {
			config.ConnectTimeout = time.Duration(timeout) * time.Second
		}
	}
	if val, ok := params["readTimeout"]; ok {
		if timeout, ok := val.(float64); ok {
			config.ReadTimeout = time.Duration(timeout) * time.Second
		}
	}
	if val, ok := params["writeTimeout"]; ok {
		if timeout, ok := val.(float64); ok {
			config.WriteTimeout = time.Duration(timeout) * time.Second
		}
	}
	if val, ok := params["keepAliveInterval"]; ok {
		if interval, ok := val.(float64); ok {
			config.KeepAliveInterval = time.Duration(interval) * time.Second
		}
	}
	if val, ok := params["reconnectInterval"]; ok {
		if interval, ok := val.(float64); ok {
			config.ReconnectInterval = time.Duration(interval) * time.Second
		}
	}
	if val, ok := params["maxReconnectInterval"]; ok {
		if interval, ok := val.(float64); ok {
			config.MaxReconnectInterval = time.Duration(interval) * time.Second
		}
	}
	if val, ok := params["progressReportInterval"]; ok {
		if interval, ok := val.(float64); ok {
			config.ProgressReportInterval = time.Duration(interval) * time.Second
		}
	}

	// 直接映射的字段（类型匹配）
	if val, ok := params["maxReconnectAttempts"].(float64); ok {
		config.MaxReconnectAttempts = int(val)
	}
	if val, ok := params["autoReconnect"].(bool); ok {
		config.AutoReconnect = val
	}
	if val, ok := params["concurrentTransfers"].(float64); ok {
		config.ConcurrentTransfers = int(val)
	}
	if val, ok := params["bufferSize"].(float64); ok {
		config.BufferSize = int(val)
	}
	if val, ok := params["maxPacketSize"].(float64); ok {
		config.MaxPacketSize = int(val)
	}
	if val, ok := params["verboseLogging"].(bool); ok {
		config.VerboseLogging = val
	}
	if val, ok := params["enableProgressMonitoring"].(bool); ok {
		config.EnableProgressMonitoring = val
	}
}

// mapTransferOptions 映射传输选项
func (f *SFTPExecutorFactory) mapTransferOptions(options *configs.SFTPTransferOptions, params map[string]interface{}) {
	// 直接布尔映射
	if val, ok := params["overwriteExisting"].(bool); ok {
		options.OverwriteExisting = val
	}
	if val, ok := params["skipExisting"].(bool); ok {
		options.SkipExisting = val
	}
	if val, ok := params["createTargetDir"].(bool); ok {
		options.CreateTargetDir = val
	}
	if val, ok := params["deleteSourceAfterTransfer"].(bool); ok {
		options.DeleteSourceAfterTransfer = val
	}
	if val, ok := params["preservePermissions"].(bool); ok {
		options.PreservePermissions = val
	}
	if val, ok := params["preserveTimestamps"].(bool); ok {
		options.PreserveTimestamps = val
	}
	if val, ok := params["verifyIntegrity"].(bool); ok {
		options.VerifyIntegrity = val
	}
	if val, ok := params["continueOnError"].(bool); ok {
		options.ContinueOnError = val
	}
	if val, ok := params["enableCompression"].(bool); ok {
		options.EnableCompression = val
	}

	// 数值映射
	if val, ok := params["bufferSize"].(float64); ok {
		options.BufferSize = int(val)
	}
	if val, ok := params["concurrentTransfers"].(float64); ok {
		options.ConcurrentTransfers = int(val)
	}
	if val, ok := params["retryCount"].(float64); ok {
		options.RetryCount = int(val)
	}
	if val, ok := params["compressionLevel"].(float64); ok {
		options.CompressionLevel = int(val)
	}
	if val, ok := params["maxFileSize"].(float64); ok {
		options.MaxFileSize = int64(val)
	}
	if val, ok := params["minFileSize"].(float64); ok {
		options.MinFileSize = int64(val)
	}

	// Duration映射
	if val, ok := params["transferTimeout"].(float64); ok {
		options.TransferTimeout = time.Duration(val) * time.Second
	}
	if val, ok := params["retryInterval"].(float64); ok {
		options.RetryInterval = time.Duration(val) * time.Second
	}
	if val, ok := params["progressReportInterval"].(float64); ok {
		options.ProgressReportInterval = time.Duration(val) * time.Second
	}

	// 字符串数组映射
	if val, ok := params["includePatterns"].([]interface{}); ok {
		options.IncludePatterns = f.convertToStringSlice(val)
	}
	if val, ok := params["excludePatterns"].([]interface{}); ok {
		options.ExcludePatterns = f.convertToStringSlice(val)
	}
	if val, ok := params["includeExtensions"].([]interface{}); ok {
		options.IncludeExtensions = f.convertToStringSlice(val)
	}
	if val, ok := params["excludeExtensions"].([]interface{}); ok {
		options.ExcludeExtensions = f.convertToStringSlice(val)
	}
}

// convertToStringSlice 将interface{}切片转换为字符串切片
func (f *SFTPExecutorFactory) convertToStringSlice(src []interface{}) []string {
	result := make([]string, 0, len(src))
	for _, item := range src {
		if str, ok := item.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// getAuthTypeName 获取认证类型名称
// 用于日志记录，将认证类型枚举转换为可读的字符串
// 参数:
//   toolConfig: 工具配置对象
// 返回:
//   string: 认证类型名称
func (f *SFTPExecutorFactory) getAuthTypeName(toolConfig *tooltypes.ToolConfig) string {
	if toolConfig.AuthType == nil {
		return "未知"
	}
	switch *toolConfig.AuthType {
	case tooltypes.AuthTypePassword:
		return "密码认证"
	case tooltypes.AuthTypePublicKey:
		return "公钥认证"
	default:
		return "未知类型"
	}
}

// CreateSFTPExecutorStatic 静态方法创建SFTP执行器
// 提供一个便捷的静态方法，直接创建SFTP任务执行器
// 参数:
//   ctx: 上下文对象
//   db: 数据库连接
//   tenantId: 租户ID
//   toolConfigId: 工具配置ID
// 返回:
//   *SFTPTaskExecutor: SFTP任务执行器
//   error: 创建过程中的错误信息
func CreateSFTPExecutorStatic(ctx context.Context, db database.Database, tenantId, toolConfigId string) (*SFTPTaskExecutor, error) {
	// 创建临时任务对象
	task := &timertypes.TimerTask{
		TaskId:       fmt.Sprintf("static_%s_%d", toolConfigId, time.Now().Unix()),
		TaskName:     "静态创建SFTP任务",
		TenantId:     tenantId,
		ToolConfigId: toolConfigId,
		ExecutorType: "SFTP_TRANSFER",
		TaskStatus:   timertypes.TaskStatusPending,
		ActiveFlag:   "Y",
	}

	// 创建工厂实例
	daoManager := dao.NewDAOManager(db)
	factory := NewSFTPExecutorFactory(daoManager)
	
	executor, err := factory.CreateExecutor(ctx, task)
	if err != nil {
		return nil, err
	}
	
	// 类型断言转换为SFTP执行器
	sftpExecutor, ok := executor.(*SFTPTaskExecutor)
	if !ok {
		return nil, fmt.Errorf("创建的执行器类型不匹配，期望*SFTPTaskExecutor")
	}
	
	return sftpExecutor, nil
}