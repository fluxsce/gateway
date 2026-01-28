package types

import "time"

// Namespace 命名空间实体
// 对应数据库表：HUB_SERVICE_NAMESPACE
type Namespace struct {
	// 主键和租户信息
	NamespaceId string `json:"namespaceId" db:"namespaceId" form:"namespaceId" query:"namespaceId"` // 命名空间ID，主键
	TenantId    string `json:"tenantId" db:"tenantId" form:"tenantId" query:"tenantId"`             // 租户ID，用于多租户数据隔离

	// 关联服务中心实例
	InstanceName string `json:"instanceName" db:"instanceName" form:"instanceName" query:"instanceName"` // 服务中心实例名称，关联 HUB_SERVICE_CENTER_CONFIG
	Environment  string `json:"environment" db:"environment" form:"environment" query:"environment"`     // 部署环境（DEVELOPMENT开发,STAGING预发布,PRODUCTION生产）

	// 命名空间基本信息
	NamespaceName     string `json:"namespaceName" db:"namespaceName" form:"namespaceName" query:"namespaceName"` // 命名空间名称
	NamespaceDesc     string `json:"namespaceDescription" db:"namespaceDescription" form:"namespaceDescription"`  // 命名空间描述
	ServiceQuotaLimit int    `json:"serviceQuotaLimit" db:"serviceQuotaLimit" form:"serviceQuotaLimit"`           // 服务数量配额限制，0表示无限制
	ConfigQuotaLimit  int    `json:"configQuotaLimit" db:"configQuotaLimit" form:"configQuotaLimit"`              // 配置数量配额限制，0表示无限制

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
