package models

import (
	"time"
)

// ToolConfigGroup 定义工具配置分组信息，对应数据库表 HUB_TOOL_CONFIG_GROUP
type ToolConfigGroup struct {
	// 主键信息
	ConfigGroupId     string    `json:"configGroupId" form:"configGroupId" query:"configGroupId" db:"configGroupId"`         // 配置分组ID
	TenantId          string    `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                             // 租户ID
	
	// 分组信息
	GroupName         string    `json:"groupName" form:"groupName" query:"groupName" db:"groupName"`                         // 分组名称
	GroupDescription  *string   `json:"groupDescription" form:"groupDescription" query:"groupDescription" db:"groupDescription"` // 分组描述
	ParentGroupId     *string   `json:"parentGroupId" form:"parentGroupId" query:"parentGroupId" db:"parentGroupId"`         // 父分组ID，支持层级结构
	GroupLevel        *int      `json:"groupLevel" form:"groupLevel" query:"groupLevel" db:"groupLevel"`                     // 分组层级，从1开始
	GroupPath         *string   `json:"groupPath" form:"groupPath" query:"groupPath" db:"groupPath"`                         // 分组路径，如/root/parent/child
	
	// 分组属性
	GroupType         *string   `json:"groupType" form:"groupType" query:"groupType" db:"groupType"`                         // 分组类型，如environment、project、department
	SortOrder         *int      `json:"sortOrder" form:"sortOrder" query:"sortOrder" db:"sortOrder"`                         // 排序顺序，数值越小越靠前
	GroupIcon         *string   `json:"groupIcon" form:"groupIcon" query:"groupIcon" db:"groupIcon"`                         // 分组图标
	GroupColor        *string   `json:"groupColor" form:"groupColor" query:"groupColor" db:"groupColor"`                     // 分组颜色代码
	
	// 权限控制
	AccessLevel       *string   `json:"accessLevel" form:"accessLevel" query:"accessLevel" db:"accessLevel"`                 // 访问级别，如private、public、restricted
	AllowedUsers      *string   `json:"allowedUsers" form:"allowedUsers" query:"allowedUsers" db:"allowedUsers"`             // 允许访问的用户列表，JSON格式
	AllowedRoles      *string   `json:"allowedRoles" form:"allowedRoles" query:"allowedRoles" db:"allowedRoles"`             // 允许访问的角色列表，JSON格式
	
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
func (ToolConfigGroup) TableName() string {
	return "HUB_TOOL_CONFIG_GROUP"
}

	