// Package configs 提供工具配置定义
// 包含各种工具的配置结构和默认值
package configs

import (
	"time"
)

// SFTPConfig SFTP客户端配置
// 包含连接、认证、传输等所有配置选项
type SFTPConfig struct {
	// ===== 连接配置 =====
	
	// 服务器地址
	Host string `json:"host" yaml:"host"`
	
	// 服务器端口
	Port int `json:"port" yaml:"port"`
	
	// 用户名
	Username string `json:"username" yaml:"username"`
	
	// 连接超时时间
	ConnectTimeout time.Duration `json:"connect_timeout" yaml:"connect_timeout"`
	
	// 读取超时时间
	ReadTimeout time.Duration `json:"read_timeout" yaml:"read_timeout"`
	
	// 写入超时时间
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`
	
	// Keep-Alive间隔
	KeepAliveInterval time.Duration `json:"keep_alive_interval" yaml:"keep_alive_interval"`
	
	// 最大重连次数
	MaxReconnectAttempts int `json:"max_reconnect_attempts" yaml:"max_reconnect_attempts"`
	
	// 重连间隔
	ReconnectInterval time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`
	
	// 最大重连间隔（用于指数退避）
	MaxReconnectInterval time.Duration `json:"max_reconnect_interval" yaml:"max_reconnect_interval"`
	
	// 是否自动重连
	AutoReconnect bool `json:"auto_reconnect" yaml:"auto_reconnect"`
	
	// ===== 认证配置 =====
	
	// 认证方式
	AuthMethods []AuthMethod `json:"auth_methods" yaml:"auth_methods"`
	
	// 密码认证配置
	PasswordAuth *PasswordAuthConfig `json:"password_auth,omitempty" yaml:"password_auth,omitempty"`
	
	// 公钥认证配置
	PublicKeyAuth *PublicKeyAuthConfig `json:"public_key_auth,omitempty" yaml:"public_key_auth,omitempty"`
	
	// 键盘交互认证配置
	KeyboardInteractiveAuth *KeyboardInteractiveAuthConfig `json:"keyboard_interactive_auth,omitempty" yaml:"keyboard_interactive_auth,omitempty"`
	
	// 主机密钥验证配置
	HostKeyVerification *HostKeyVerificationConfig `json:"host_key_verification,omitempty" yaml:"host_key_verification,omitempty"`
	
	// ===== 传输配置 =====
	
	// 默认传输选项
	DefaultTransferOptions *SFTPTransferOptions `json:"default_transfer_options,omitempty" yaml:"default_transfer_options,omitempty"`
	
	// 默认同步选项
	DefaultSyncOptions *SFTPSyncOptions `json:"default_sync_options,omitempty" yaml:"default_sync_options,omitempty"`
	
	// ===== 性能配置 =====
	
	// 并发传输数量
	ConcurrentTransfers int `json:"concurrent_transfers" yaml:"concurrent_transfers"`
	
	// 缓冲区大小
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`
	
	// 最大包大小
	MaxPacketSize int `json:"max_packet_size" yaml:"max_packet_size"`
	
	// ===== 日志和监控配置 =====
	
	// 是否启用详细日志
	VerboseLogging bool `json:"verbose_logging" yaml:"verbose_logging"`
	
	// 是否启用进度监控
	EnableProgressMonitoring bool `json:"enable_progress_monitoring" yaml:"enable_progress_monitoring"`
	
	// 进度报告间隔
	ProgressReportInterval time.Duration `json:"progress_report_interval" yaml:"progress_report_interval"`
	
	// ===== 扩展配置 =====
	
	// 自定义属性
	CustomAttributes map[string]interface{} `json:"custom_attributes,omitempty" yaml:"custom_attributes,omitempty"`
	
	// 插件配置
	PluginConfigs map[string]interface{} `json:"plugin_configs,omitempty" yaml:"plugin_configs,omitempty"`
}

// AuthMethod 认证方式枚举
type AuthMethod int

const (
	// AuthMethodPassword 密码认证
	AuthMethodPassword AuthMethod = iota + 1
	
	// AuthMethodPublicKey 公钥认证
	AuthMethodPublicKey
	
	// AuthMethodKeyboardInteractive 键盘交互认证
	AuthMethodKeyboardInteractive
	
	// AuthMethodHostBased 基于主机的认证
	AuthMethodHostBased
)

// String 返回认证方式的字符串表示
func (a AuthMethod) String() string {
	switch a {
	case AuthMethodPassword:
		return "password"
	case AuthMethodPublicKey:
		return "publickey"
	case AuthMethodKeyboardInteractive:
		return "keyboard-interactive"
	case AuthMethodHostBased:
		return "hostbased"
	default:
		return "unknown"
	}
}

// PasswordAuthConfig 密码认证配置
type PasswordAuthConfig struct {
	// 密码
	Password string `json:"password" yaml:"password"`
	
	// 是否允许空密码
	AllowEmptyPassword bool `json:"allow_empty_password" yaml:"allow_empty_password"`
	
	// 密码重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
}

// PublicKeyAuthConfig 公钥认证配置
type PublicKeyAuthConfig struct {
	// 私钥文件路径
	PrivateKeyPath string `json:"private_key_path" yaml:"private_key_path"`
	
	// 私钥内容（PEM格式）
	PrivateKeyData []byte `json:"private_key_data,omitempty" yaml:"private_key_data,omitempty"`
	
	// 私钥密码
	Passphrase string `json:"passphrase,omitempty" yaml:"passphrase,omitempty"`
	
	// 公钥文件路径
	PublicKeyPath string `json:"public_key_path,omitempty" yaml:"public_key_path,omitempty"`
	
	// 公钥内容
	PublicKeyData []byte `json:"public_key_data,omitempty" yaml:"public_key_data,omitempty"`
}

// KeyboardInteractiveAuthConfig 键盘交互认证配置
type KeyboardInteractiveAuthConfig struct {
	// 是否启用键盘交互认证
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// 预设的回答列表
	Answers []string `json:"answers,omitempty" yaml:"answers,omitempty"`
	
	// 最大重试次数
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
}

// HostKeyVerificationConfig 主机密钥验证配置
type HostKeyVerificationConfig struct {
	// 是否启用主机密钥验证
	Enabled bool `json:"enabled" yaml:"enabled"`
	
	// 已知主机文件路径
	KnownHostsFile string `json:"known_hosts_file,omitempty" yaml:"known_hosts_file,omitempty"`
	
	// 严格主机密钥检查
	StrictHostKeyChecking bool `json:"strict_host_key_checking" yaml:"strict_host_key_checking"`
	
	// 受信任的主机密钥
	TrustedHostKeys [][]byte `json:"trusted_host_keys,omitempty" yaml:"trusted_host_keys,omitempty"`
	
	// 主机密钥算法
	HostKeyAlgorithms []string `json:"host_key_algorithms,omitempty" yaml:"host_key_algorithms,omitempty"`
}

// SFTPTransferOptions SFTP传输选项配置
// 控制文件传输的各种行为和性能参数
type SFTPTransferOptions struct {
	// ===== 传输行为 =====
	
	// 是否覆盖已存在的文件
	OverwriteExisting bool `json:"overwrite_existing" yaml:"overwrite_existing"`
	
	// 是否跳过已存在的文件
	SkipExisting bool `json:"skip_existing" yaml:"skip_existing"`
	
	// 是否创建目标目录
	CreateTargetDir bool `json:"create_target_dir" yaml:"create_target_dir"`
	
	// 是否保持文件权限
	PreservePermissions bool `json:"preserve_permissions" yaml:"preserve_permissions"`
	
	// 是否保持文件时间戳
	PreserveTimestamps bool `json:"preserve_timestamps" yaml:"preserve_timestamps"`
	
	// 是否验证传输完整性
	VerifyIntegrity bool `json:"verify_integrity" yaml:"verify_integrity"`
	
	// ===== 性能参数 =====
	
	// 缓冲区大小
	BufferSize int `json:"buffer_size" yaml:"buffer_size"`
	
	// 并发传输数量
	ConcurrentTransfers int `json:"concurrent_transfers" yaml:"concurrent_transfers"`
	
	// 传输超时时间
	TransferTimeout time.Duration `json:"transfer_timeout" yaml:"transfer_timeout"`
	
	// 重试次数
	RetryCount int `json:"retry_count" yaml:"retry_count"`
	
	// 重试间隔
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	
	// ===== 过滤和限制 =====
	
	// 文件大小限制（字节）
	MaxFileSize int64 `json:"max_file_size" yaml:"max_file_size"`
	
	// 最小文件大小（字节）
	MinFileSize int64 `json:"min_file_size" yaml:"min_file_size"`
	
	// 包含的文件模式
	IncludePatterns []string `json:"include_patterns,omitempty" yaml:"include_patterns,omitempty"`
	
	// 排除的文件模式
	ExcludePatterns []string `json:"exclude_patterns,omitempty" yaml:"exclude_patterns,omitempty"`
	
	// 包含的文件扩展名
	IncludeExtensions []string `json:"include_extensions,omitempty" yaml:"include_extensions,omitempty"`
	
	// 排除的文件扩展名
	ExcludeExtensions []string `json:"exclude_extensions,omitempty" yaml:"exclude_extensions,omitempty"`
	
	// ===== 进度监控 =====
	
	// 进度报告间隔
	ProgressReportInterval time.Duration `json:"progress_report_interval" yaml:"progress_report_interval"`
	
	// ===== 压缩选项 =====
	
	// 是否启用压缩
	EnableCompression bool `json:"enable_compression" yaml:"enable_compression"`
	
	// 压缩级别（1-9）
	CompressionLevel int `json:"compression_level" yaml:"compression_level"`
	
	// ===== 错误处理 =====
	
	// 错误时是否继续
	ContinueOnError bool `json:"continue_on_error" yaml:"continue_on_error"`
	
	// ===== 自定义选项 =====
	
	// 自定义属性
	CustomOptions map[string]interface{} `json:"custom_options,omitempty" yaml:"custom_options,omitempty"`
}

// SFTPSyncOptions SFTP同步选项配置
// 控制目录同步的各种行为参数
type SFTPSyncOptions struct {
	// ===== 基础同步选项 =====
	
	// 是否删除目标中不存在于源的文件
	DeleteExtraneous bool `json:"delete_extraneous" yaml:"delete_extraneous"`
	
	// 是否使用校验和比较文件
	UseChecksum bool `json:"use_checksum" yaml:"use_checksum"`
	
	// 是否只比较大小和修改时间
	UseSizeAndTime bool `json:"use_size_and_time" yaml:"use_size_and_time"`
	
	// 是否进行试运行（不实际执行操作）
	DryRun bool `json:"dry_run" yaml:"dry_run"`
	
	// ===== 冲突处理 =====
	
	// 冲突解决策略
	ConflictResolution ConflictResolution `json:"conflict_resolution" yaml:"conflict_resolution"`
	
	// 备份冲突文件
	BackupConflictedFiles bool `json:"backup_conflicted_files" yaml:"backup_conflicted_files"`
	
	// 备份文件后缀
	BackupSuffix string `json:"backup_suffix" yaml:"backup_suffix"`
	
	// ===== 性能选项 =====
	
	// 并发同步数量
	ConcurrentSyncs int `json:"concurrent_syncs" yaml:"concurrent_syncs"`
	
	// 批量操作大小
	BatchSize int `json:"batch_size" yaml:"batch_size"`
	
	// ===== 过滤选项 =====
	
	// 包含的文件模式
	IncludePatterns []string `json:"include_patterns,omitempty" yaml:"include_patterns,omitempty"`
	
	// 排除的文件模式
	ExcludePatterns []string `json:"exclude_patterns,omitempty" yaml:"exclude_patterns,omitempty"`
	
	// 是否同步隐藏文件
	SyncHiddenFiles bool `json:"sync_hidden_files" yaml:"sync_hidden_files"`
	
	// 是否同步空目录
	SyncEmptyDirectories bool `json:"sync_empty_directories" yaml:"sync_empty_directories"`
	
	// ===== 传输选项 =====
	
	// 嵌入的传输选项
	TransferOptions *SFTPTransferOptions `json:"transfer_options,omitempty" yaml:"transfer_options,omitempty"`
}

// ConflictResolution 冲突解决策略枚举
type ConflictResolution int

const (
	// ConflictResolutionSkip 跳过冲突文件
	ConflictResolutionSkip ConflictResolution = iota + 1
	
	// ConflictResolutionOverwrite 覆盖目标文件
	ConflictResolutionOverwrite
	
	// ConflictResolutionKeepBoth 保留两个文件
	ConflictResolutionKeepBoth
	
	// ConflictResolutionNewest 保留最新的文件
	ConflictResolutionNewest
	
	// ConflictResolutionLargest 保留最大的文件
	ConflictResolutionLargest
	
	// ConflictResolutionAsk 询问用户
	ConflictResolutionAsk
)

// String 返回冲突解决策略的字符串表示
func (c ConflictResolution) String() string {
	switch c {
	case ConflictResolutionSkip:
		return "skip"
	case ConflictResolutionOverwrite:
		return "overwrite"
	case ConflictResolutionKeepBoth:
		return "keep_both"
	case ConflictResolutionNewest:
		return "newest"
	case ConflictResolutionLargest:
		return "largest"
	case ConflictResolutionAsk:
		return "ask"
	default:
		return "unknown"
	}
}

// DefaultSFTPConfig 返回默认SFTP配置
func DefaultSFTPConfig() *SFTPConfig {
	return &SFTPConfig{
		Port:                     22,
		ConnectTimeout:           30 * time.Second,
		ReadTimeout:              30 * time.Second,
		WriteTimeout:             30 * time.Second,
		KeepAliveInterval:        30 * time.Second,
		MaxReconnectAttempts:     3,
		ReconnectInterval:        5 * time.Second,
		MaxReconnectInterval:     60 * time.Second,
		AutoReconnect:            true,
		AuthMethods:              []AuthMethod{AuthMethodPublicKey, AuthMethodPassword},
		ConcurrentTransfers:      3,
		BufferSize:               32 * 1024, // 32KB
		MaxPacketSize:            32 * 1024, // 32KB
		VerboseLogging:           false,
		EnableProgressMonitoring: true,
		ProgressReportInterval:   time.Second,
		DefaultTransferOptions:   DefaultSFTPTransferOptions(),
		DefaultSyncOptions:       DefaultSFTPSyncOptions(),
	}
}

// DefaultSFTPTransferOptions 返回默认SFTP传输选项
func DefaultSFTPTransferOptions() *SFTPTransferOptions {
	return &SFTPTransferOptions{
		OverwriteExisting:      false,
		SkipExisting:           true,
		CreateTargetDir:        true,
		PreservePermissions:    true,
		PreserveTimestamps:     true,
		VerifyIntegrity:        false,
		BufferSize:             32 * 1024, // 32KB
		ConcurrentTransfers:    3,
		TransferTimeout:        5 * time.Minute,
		RetryCount:             3,
		RetryInterval:          time.Second,
		MaxFileSize:            0, // 无限制
		MinFileSize:            0,
		ProgressReportInterval: time.Second,
		EnableCompression:      false,
		CompressionLevel:       6,
		ContinueOnError:        true,
	}
}

// DefaultSFTPSyncOptions 返回默认SFTP同步选项
func DefaultSFTPSyncOptions() *SFTPSyncOptions {
	return &SFTPSyncOptions{
		DeleteExtraneous:        false,
		UseChecksum:             false,
		UseSizeAndTime:          true,
		DryRun:                  false,
		ConflictResolution:      ConflictResolutionSkip,
		BackupConflictedFiles:   false,
		BackupSuffix:            ".bak",
		ConcurrentSyncs:         3,
		BatchSize:               100,
		SyncHiddenFiles:         false,
		SyncEmptyDirectories:    true,
		TransferOptions:         DefaultSFTPTransferOptions(),
	}
} 