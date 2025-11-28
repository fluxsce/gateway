package models

import (
	"time"
)

// Role 角色模型，对应数据库HUB_AUTH_ROLE表
type Role struct {
	// 主键和租户信息
	RoleId   string `json:"roleId" form:"roleId" query:"roleId" db:"roleId"`         // 角色ID，主键
	TenantId string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"` // 租户ID，用于多租户数据隔离

	// 角色基本信息
	RoleName        string `json:"roleName" form:"roleName" query:"roleName" db:"roleName"`                             // 角色名称
	RoleDescription string `json:"roleDescription" form:"roleDescription" query:"roleDescription" db:"roleDescription"` // 角色描述

	// 角色状态
	RoleStatus  string `json:"roleStatus" form:"roleStatus" query:"roleStatus" db:"roleStatus"`     // 角色状态(Y:启用,N:禁用)
	BuiltInFlag string `json:"builtInFlag" form:"builtInFlag" query:"builtInFlag" db:"builtInFlag"` // 内置角色标记(Y:内置,N:自定义)

	// 数据权限范围
	DataScope string `json:"dataScope" form:"dataScope" query:"dataScope" db:"dataScope"` // 数据权限范围，TEXT类型，可存储复杂的权限配置

	// 通用字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
	ExtProperty    string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`             // 扩展属性，JSON格式
	Reserved1      string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                     // 预留字段1
	Reserved2      string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                     // 预留字段2
	Reserved3      string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                     // 预留字段3
	Reserved4      string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                     // 预留字段4
	Reserved5      string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                     // 预留字段5
	Reserved6      string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                     // 预留字段6
	Reserved7      string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                     // 预留字段7
	Reserved8      string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                     // 预留字段8
	Reserved9      string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                     // 预留字段9
	Reserved10     string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                 // 预留字段10
}

// TableName 返回表名
func (Role) TableName() string {
	return "HUB_AUTH_ROLE"
}

// RoleStatus 角色状态常量
const (
	RoleStatusEnabled  = "Y" // 启用
	RoleStatusDisabled = "N" // 禁用
)
