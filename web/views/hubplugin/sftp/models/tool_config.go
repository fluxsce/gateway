package models

import (
	"time"
)

// ToolConfig 定义工具配置信息，对应数据库表 HUB_TOOL_CONFIG
type ToolConfig struct {
	// 主键信息
	ToolConfigId      string    `json:"toolConfigId" form:"toolConfigId" query:"toolConfigId" db:"toolConfigId"`             // 工具配置ID
	TenantId          string    `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                             // 租户ID
	
	// 工具基础信息
	ToolName          string    `json:"toolName" form:"toolName" query:"toolName" db:"toolName"`                             // 工具名称，如SFTP、SSH、FTP等
	ToolType          string    `json:"toolType" form:"toolType" query:"toolType" db:"toolType"`                             // 工具类型，如transfer、database、monitor等
	ToolVersion       *string   `json:"toolVersion" form:"toolVersion" query:"toolVersion" db:"toolVersion"`                 // 工具版本号
	ConfigName        string    `json:"configName" form:"configName" query:"configName" db:"configName"`                     // 配置名称，用于区分同一工具的不同配置
	ConfigDescription *string   `json:"configDescription" form:"configDescription" query:"configDescription" db:"configDescription"` // 配置描述信息
	
	// 分组信息
	ConfigGroupId     *string   `json:"configGroupId" form:"configGroupId" query:"configGroupId" db:"configGroupId"`         // 配置分组ID
	ConfigGroupName   *string   `json:"configGroupName" form:"configGroupName" query:"configGroupName" db:"configGroupName"` // 配置分组名称
	
	// 连接配置
	HostAddress       *string   `json:"hostAddress" form:"hostAddress" query:"hostAddress" db:"hostAddress"`                 // 主机地址或域名
	PortNumber        *int      `json:"portNumber" form:"portNumber" query:"portNumber" db:"portNumber"`                     // 端口号
	ProtocolType      *string   `json:"protocolType" form:"protocolType" query:"protocolType" db:"protocolType"`             // 协议类型，如TCP、UDP、HTTP等
	
	// 认证配置
	AuthType          *string   `json:"authType" form:"authType" query:"authType" db:"authType"`                             // 认证类型，如password、publickey、oauth等
	UserName          *string   `json:"userName" form:"userName" query:"userName" db:"userName"`                             // 用户名
	PasswordEncrypted *string   `json:"passwordEncrypted" form:"passwordEncrypted" query:"passwordEncrypted" db:"passwordEncrypted"` // 加密后的密码
	KeyFilePath       *string   `json:"keyFilePath" form:"keyFilePath" query:"keyFilePath" db:"keyFilePath"`                 // 密钥文件路径
	KeyFileContent    *string   `json:"keyFileContent" form:"keyFileContent" query:"keyFileContent" db:"keyFileContent"`     // 密钥文件内容，加密存储
	
	// 配置参数
	ConfigParameters   *string   `json:"configParameters" form:"configParameters" query:"configParameters" db:"configParameters"` // 配置参数，JSON格式存储
	EnvironmentVariables *string `json:"environmentVariables" form:"environmentVariables" query:"environmentVariables" db:"environmentVariables"` // 环境变量配置，JSON格式存储
	CustomSettings     *string   `json:"customSettings" form:"customSettings" query:"customSettings" db:"customSettings"`     // 自定义设置，JSON格式存储
	
	// 状态和控制
	ConfigStatus      string    `json:"configStatus" form:"configStatus" query:"configStatus" db:"configStatus"`             // 配置状态(N禁用,Y启用)
	DefaultFlag       string    `json:"defaultFlag" form:"defaultFlag" query:"defaultFlag" db:"defaultFlag"`                 // 是否为默认配置(N否,Y是)
	PriorityLevel     *int      `json:"priorityLevel" form:"priorityLevel" query:"priorityLevel" db:"priorityLevel"`         // 优先级，数值越小优先级越高
	
	// 安全和加密
	EncryptionType    *string   `json:"encryptionType" form:"encryptionType" query:"encryptionType" db:"encryptionType"`     // 加密类型，如AES256、RSA等
	EncryptionKey     *string   `json:"encryptionKey" form:"encryptionKey" query:"encryptionKey" db:"encryptionKey"`         // 加密密钥标识
	
	// 通用字段
	AddTime           time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                 // 创建时间
	AddWho            string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                     // 创建人ID
	EditTime          time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                             // 最后修改时间
	EditWho           string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                 // 最后修改人ID
	OprSeqFlag        string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                     // 操作序列标识
	CurrentVersion    int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`     // 当前版本号
	ActiveFlag        string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                     // 活动状态标记(N非活动,Y活动)
	NoteText          *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                             // 备注信息
	ExtProperty       *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                 // 扩展属性，JSON格式
	Reserved1         *string   `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                         // 预留字段1
	Reserved2         *string   `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                         // 预留字段2
	Reserved3         *string   `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                         // 预留字段3
	Reserved4         *string   `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                         // 预留字段4
	Reserved5         *string   `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                         // 预留字段5
	Reserved6         *string   `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                         // 预留字段6
	Reserved7         *string   `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                         // 预留字段7
	Reserved8         *string   `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                         // 预留字段8
	Reserved9         *string   `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                         // 预留字段9
	Reserved10        *string   `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                     // 预留字段10
}

// TableName 返回数据库表名
func (ToolConfig) TableName() string {
	return "HUB_TOOL_CONFIG"
}