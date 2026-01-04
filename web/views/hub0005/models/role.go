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

// RoleQuery 角色查询条件，对应前端 /queryRoles 的查询参数
type RoleQuery struct {
	RoleName        string `json:"roleName" form:"roleName" query:"roleName"`                      // 角色名称（模糊查询）
	RoleDescription string `json:"roleDescription" form:"roleDescription" query:"roleDescription"` // 角色描述（模糊查询）
	RoleStatus      string `json:"roleStatus" form:"roleStatus" query:"roleStatus"`                // 角色状态：Y/N，空表示全部
	BuiltInFlag     string `json:"builtInFlag" form:"builtInFlag" query:"builtInFlag"`             // 内置角色标记：Y/N，空表示全部
	ActiveFlag      string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`                // 活动标记：Y-活动，N-非活动，空表示全部
}

// RoleResource 角色权限关联模型，对应数据库HUB_AUTH_ROLE_RESOURCE表
type RoleResource struct {
	// 主键和租户信息
	RoleResourceId string `json:"roleResourceId" form:"roleResourceId" query:"roleResourceId" db:"roleResourceId"` // 角色资源关联ID，主键
	TenantId       string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID，用于多租户数据隔离

	// 关联信息
	RoleId     string `json:"roleId" form:"roleId" query:"roleId" db:"roleId"`                 // 角色ID
	ResourceId string `json:"resourceId" form:"resourceId" query:"resourceId" db:"resourceId"` // 资源ID

	// 权限控制
	PermissionType string     `json:"permissionType" form:"permissionType" query:"permissionType" db:"permissionType"` // 权限类型(ALLOW:允许,DENY:拒绝)
	GrantedBy      string     `json:"grantedBy" form:"grantedBy" query:"grantedBy" db:"grantedBy"`                     // 授权人ID
	GrantedTime    time.Time  `json:"grantedTime" form:"grantedTime" query:"grantedTime" db:"grantedTime"`             // 授权时间
	ExpireTime     *time.Time `json:"expireTime" form:"expireTime" query:"expireTime" db:"expireTime"`                 // 过期时间，NULL表示永不过期

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
func (RoleResource) TableName() string {
	return "HUB_AUTH_ROLE_RESOURCE"
}

// PermissionType 权限类型常量
const (
	PermissionTypeAllow = "ALLOW" // 允许
	PermissionTypeDeny  = "DENY"  // 拒绝
)
