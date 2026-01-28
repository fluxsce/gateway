package types

import "time"

// ConfigData 配置数据实体
// 对应数据库表：HUB_SERVICE_CONFIG_DATA
type ConfigData struct {
	// 主键和租户信息
	ConfigDataId string `json:"configDataId" db:"configDataId" form:"configDataId" query:"configDataId"` // 配置数据ID，主键
	TenantId     string `json:"tenantId" db:"tenantId" form:"tenantId" query:"tenantId"`                 // 租户ID，用于多租户数据隔离

	// 关联命名空间和分组
	NamespaceId string `json:"namespaceId" db:"namespaceId" form:"namespaceId" query:"namespaceId"` // 命名空间ID，关联HUB_SERVICE_NAMESPACE表
	GroupName   string `json:"groupName" db:"groupName" form:"groupName" query:"groupName"`         // 分组名称，如DEFAULT_GROUP

	// 配置基本信息
	ConfigContent     string `json:"configContent" db:"configContent" form:"configContent"`             // 配置内容，支持大文本
	ContentType       string `json:"contentType" db:"contentType" form:"contentType"`                   // 内容类型(text:文本,json:JSON,xml:XML,yaml:YAML,properties:Properties)
	ConfigDescription string `json:"configDescription" db:"configDescription" form:"configDescription"` // 配置描述
	Encrypted         string `json:"encrypted" db:"encrypted" form:"encrypted"`                         // 是否加密存储(N否,Y是)

	// 版本信息
	Version int64 `json:"version" db:"version"` // 配置版本号（BIGINT），每次修改递增

	// MD5校验值（用于配置变更检测）
	Md5Value string `json:"md5Value" db:"md5Value"` // 配置内容的MD5值，用于快速比较配置是否变更

	// 通用字段（对应数据库 DATETIME/DATE 类型）
	AddTime        time.Time `json:"addTime" db:"addTime"`                                            // 创建时间（DATETIME/DATE NOT NULL）
	AddWho         string    `json:"addWho" db:"addWho" form:"addWho"`                                // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`                                          // 最后修改时间（DATETIME/DATE NOT NULL）
	EditWho        string    `json:"editWho" db:"editWho" form:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`                                      // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`                              // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag" form:"activeFlag" query:"activeFlag"` // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" db:"noteText" form:"noteText"`                          // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty" form:"extProperty"`                 // 扩展属性，JSON格式
}

// ConfigHistory 配置历史实体
// 对应数据库表：HUB_SERVICE_CONFIG_HISTORY
type ConfigHistory struct {
	// 主键和租户信息
	ConfigHistoryId string `json:"configHistoryId" db:"configHistoryId" query:"configHistoryId"` // 配置历史ID，主键
	TenantId        string `json:"tenantId" db:"tenantId" form:"tenantId" query:"tenantId"`      // 租户ID，用于多租户数据隔离

	// 关联配置数据
	ConfigDataId string `json:"configDataId" db:"configDataId" form:"configDataId" query:"configDataId"` // 配置数据ID，关联HUB_SERVICE_CONFIG_DATA表
	NamespaceId  string `json:"namespaceId" db:"namespaceId" form:"namespaceId" query:"namespaceId"`     // 命名空间ID，冗余字段便于查询
	GroupName    string `json:"groupName" db:"groupName" form:"groupName" query:"groupName"`             // 分组名称，冗余字段便于查询

	// 变更信息
	ChangeType  string `json:"changeType" db:"changeType" form:"changeType" query:"changeType"` // 变更类型(CREATE:创建,UPDATE:更新,DELETE:删除,ROLLBACK:回滚)
	OldContent  string `json:"oldContent" db:"oldContent"`                                      // 旧配置内容
	NewContent  string `json:"newContent" db:"newContent"`                                      // 新配置内容
	OldVersion  int64  `json:"oldVersion" db:"oldVersion"`                                      // 旧版本号（BIGINT）
	NewVersion  int64  `json:"newVersion" db:"newVersion"`                                      // 新版本号（BIGINT）
	OldMd5Value string `json:"oldMd5Value" db:"oldMd5Value"`                                    // 旧配置MD5值
	NewMd5Value string `json:"newMd5Value" db:"newMd5Value"`                                    // 新配置MD5值

	// 变更原因和操作人
	ChangeReason string    `json:"changeReason" db:"changeReason" form:"changeReason"` // 变更原因
	ChangedBy    string    `json:"changedBy" db:"changedBy" form:"changedBy"`          // 变更人ID
	ChangedAt    time.Time `json:"changedAt" db:"changedAt"`                           // 变更时间（DATETIME/DATE NOT NULL）

	// 通用字段（对应数据库 DATETIME/DATE 类型）
	AddTime        time.Time `json:"addTime" db:"addTime"`                                            // 创建时间（DATETIME/DATE NOT NULL）
	AddWho         string    `json:"addWho" db:"addWho" form:"addWho"`                                // 创建人ID
	EditTime       time.Time `json:"editTime" db:"editTime"`                                          // 最后修改时间（DATETIME/DATE NOT NULL）
	EditWho        string    `json:"editWho" db:"editWho" form:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`                                      // 操作序列标识
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`                              // 当前版本号
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag" form:"activeFlag" query:"activeFlag"` // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" db:"noteText" form:"noteText"`                          // 备注信息
	ExtProperty    string    `json:"extProperty" db:"extProperty" form:"extProperty"`                 // 扩展属性，JSON格式
}

// ChangeType 变更类型常量
const (
	ChangeTypeCreate   = "CREATE"   // 创建
	ChangeTypeUpdate   = "UPDATE"   // 更新
	ChangeTypeDelete   = "DELETE"   // 删除
	ChangeTypeRollback = "ROLLBACK" // 回滚
)

// ContentType 内容类型常量
const (
	ContentTypeText       = "text"       // 文本
	ContentTypeJSON       = "json"       // JSON
	ContentTypeXML        = "xml"        // XML
	ContentTypeYAML       = "yaml"       // YAML
	ContentTypeProperties = "properties" // Properties
)
