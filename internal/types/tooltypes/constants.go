package tooltypes

// 通用常量
const (
	// 活动状态标记
	ActiveFlagYes = "Y" // 活动
	ActiveFlagNo  = "N" // 非活动

	// 默认配置标记
	DefaultFlagYes = "Y" // 是默认配置
	DefaultFlagNo  = "N" // 非默认配置

	// 配置状态
	ConfigStatusEnabled  = "Y" // 启用
	ConfigStatusDisabled = "N" // 禁用
)

// 工具类型常量
const (
	ToolTypeTransfer  = "transfer"  // 传输类工具
	ToolTypeDatabase  = "database"  // 数据库类工具
	ToolTypeMonitor   = "monitor"   // 监控类工具
	ToolTypeAnalysis  = "analysis"  // 分析类工具
	ToolTypeSchedule  = "schedule"  // 调度类工具
	ToolTypeIntegration = "integration" // 集成类工具
)

// 认证类型常量
const (
	AuthTypePassword   = "password"   // 密码认证
	AuthTypePublicKey  = "publickey"  // 公钥认证
	AuthTypeOAuth      = "oauth"      // OAuth认证
	AuthTypeToken      = "token"      // Token认证
	AuthTypeCertificate = "certificate" // 证书认证
)

// 协议类型常量
const (
	ProtocolTypeTCP  = "TCP"  // TCP协议
	ProtocolTypeUDP  = "UDP"  // UDP协议
	ProtocolTypeHTTP = "HTTP" // HTTP协议
	ProtocolTypeHTTPS = "HTTPS" // HTTPS协议
	ProtocolTypeFTP  = "FTP"  // FTP协议
	ProtocolTypeSFTP = "SFTP" // SFTP协议
	ProtocolTypeSSH  = "SSH"  // SSH协议
)

// 加密类型常量
const (
	EncryptionTypeAES256 = "AES256" // AES-256加密
	EncryptionTypeRSA    = "RSA"    // RSA加密
	EncryptionTypeSM4    = "SM4"    // 国密SM4加密
)

// 访问级别常量
const (
	AccessLevelPrivate    = "private"    // 私有
	AccessLevelPublic     = "public"     // 公开
	AccessLevelRestricted = "restricted" // 受限
) 