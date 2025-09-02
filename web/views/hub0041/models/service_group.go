package models

import "time"

// ServiceGroup 服务分组（命名空间）信息模型
// 对应数据库表：HUB_REGISTRY_SERVICE_GROUP
// 用途：管理服务分组和命名空间信息
type ServiceGroup struct {
	// 主键信息
	ServiceGroupId string `json:"serviceGroupId" form:"serviceGroupId" query:"serviceGroupId" db:"serviceGroupId"` // 服务分组ID，主键
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID，用于多租户数据隔离

	// 分组基本信息
	GroupName        string `json:"groupName" form:"groupName" query:"groupName" db:"groupName"`                             // 分组名称
	GroupDescription string `json:"groupDescription" form:"groupDescription" query:"groupDescription" db:"groupDescription"` // 分组描述
	GroupType        string `json:"groupType" form:"groupType" query:"groupType" db:"groupType"`                             // 分组类型(BUSINESS,SYSTEM,TEST)

	// 授权信息
	OwnerUserId          string `json:"ownerUserId" form:"ownerUserId" query:"ownerUserId" db:"ownerUserId"`                                     // 分组所有者用户ID
	AdminUserIds         string `json:"adminUserIds,omitempty" form:"adminUserIds" query:"adminUserIds" db:"adminUserIds"`                       // 管理员用户ID列表，JSON格式
	ReadUserIds          string `json:"readUserIds,omitempty" form:"readUserIds" query:"readUserIds" db:"readUserIds"`                           // 只读用户ID列表，JSON格式
	AccessControlEnabled string `json:"accessControlEnabled" form:"accessControlEnabled" query:"accessControlEnabled" db:"accessControlEnabled"` // 是否启用访问控制(N否,Y是)

	// 配置信息
	DefaultProtocolType               string `json:"defaultProtocolType" form:"defaultProtocolType" query:"defaultProtocolType" db:"defaultProtocolType"`                                                         // 默认协议类型
	DefaultLoadBalanceStrategy        string `json:"defaultLoadBalanceStrategy" form:"defaultLoadBalanceStrategy" query:"defaultLoadBalanceStrategy" db:"defaultLoadBalanceStrategy"`                             // 默认负载均衡策略
	DefaultHealthCheckUrl             string `json:"defaultHealthCheckUrl" form:"defaultHealthCheckUrl" query:"defaultHealthCheckUrl" db:"defaultHealthCheckUrl"`                                                 // 默认健康检查URL
	DefaultHealthCheckIntervalSeconds int    `json:"defaultHealthCheckIntervalSeconds" form:"defaultHealthCheckIntervalSeconds" query:"defaultHealthCheckIntervalSeconds" db:"defaultHealthCheckIntervalSeconds"` // 默认健康检查间隔(秒)

	// 通用审计字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText,omitempty" form:"noteText" query:"noteText" db:"noteText"`               // 备注信息
	ExtProperty    string    `json:"extProperty,omitempty" form:"extProperty" query:"extProperty" db:"extProperty"`   // 扩展属性，JSON格式

	// 预留字段
	Reserved1  string `json:"reserved1,omitempty" form:"reserved1" query:"reserved1" db:"reserved1"`     // 预留字段1
	Reserved2  string `json:"reserved2,omitempty" form:"reserved2" query:"reserved2" db:"reserved2"`     // 预留字段2
	Reserved3  string `json:"reserved3,omitempty" form:"reserved3" query:"reserved3" db:"reserved3"`     // 预留字段3
	Reserved4  string `json:"reserved4,omitempty" form:"reserved4" query:"reserved4" db:"reserved4"`     // 预留字段4
	Reserved5  string `json:"reserved5,omitempty" form:"reserved5" query:"reserved5" db:"reserved5"`     // 预留字段5
	Reserved6  string `json:"reserved6,omitempty" form:"reserved6" query:"reserved6" db:"reserved6"`     // 预留字段6
	Reserved7  string `json:"reserved7,omitempty" form:"reserved7" query:"reserved7" db:"reserved7"`     // 预留字段7
	Reserved8  string `json:"reserved8,omitempty" form:"reserved8" query:"reserved8" db:"reserved8"`     // 预留字段8
	Reserved9  string `json:"reserved9,omitempty" form:"reserved9" query:"reserved9" db:"reserved9"`     // 预留字段9
	Reserved10 string `json:"reserved10,omitempty" form:"reserved10" query:"reserved10" db:"reserved10"` // 预留字段10
}
