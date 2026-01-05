package models

import (
	"time"
)

// UserRole 用户角色关联模型，对应数据库HUB_AUTH_USER_ROLE表
type UserRole struct {
	// 主键和租户信息
	UserRoleId string `json:"userRoleId" form:"userRoleId" query:"userRoleId" db:"userRoleId"` // 用户角色关联ID，主键
	TenantId   string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`         // 租户ID，用于多租户数据隔离

	// 关联信息
	UserId string `json:"userId" form:"userId" query:"userId" db:"userId"` // 用户ID
	RoleId string `json:"roleId" form:"roleId" query:"roleId" db:"roleId"` // 角色ID

	// 授权控制
	GrantedBy       string     `json:"grantedBy" form:"grantedBy" query:"grantedBy" db:"grantedBy"`                         // 授权人ID
	GrantedTime     time.Time  `json:"grantedTime" form:"grantedTime" query:"grantedTime" db:"grantedTime"`                 // 授权时间
	ExpireTime      *time.Time `json:"expireTime" form:"expireTime" query:"expireTime" db:"expireTime"`                     // 过期时间，NULL表示永不过期
	PrimaryRoleFlag string     `json:"primaryRoleFlag" form:"primaryRoleFlag" query:"primaryRoleFlag" db:"primaryRoleFlag"` // 主要角色标记(Y:主要角色,N:次要角色)

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
func (UserRole) TableName() string {
	return "HUB_AUTH_USER_ROLE"
}

// UserRoleRequest 用户角色授权请求
type UserRoleRequest struct {
	UserId     string  `json:"userId" form:"userId" query:"userId"`             // 用户ID
	RoleIds    string  `json:"roleIds" form:"roleIds" query:"roleIds"`          // 角色ID列表（逗号分割）
	ExpireTime *string `json:"expireTime" form:"expireTime" query:"expireTime"` // 过期时间（可选）
}

// UserRoleResponse 用户角色响应（包含角色信息）
type UserRoleResponse struct {
	UserRoleId      string     `json:"userRoleId"`      // 用户角色关联ID
	UserId          string     `json:"userId"`          // 用户ID
	RoleId          string     `json:"roleId"`          // 角色ID
	RoleName        string     `json:"roleName"`        // 角色名称
	RoleDescription string     `json:"roleDescription"` // 角色描述
	GrantedBy       string     `json:"grantedBy"`       // 授权人ID
	GrantedTime     time.Time  `json:"grantedTime"`     // 授权时间
	ExpireTime      *time.Time `json:"expireTime"`      // 过期时间
	PrimaryRoleFlag string     `json:"primaryRoleFlag"` // 主要角色标记
	ActiveFlag      string     `json:"activeFlag"`      // 活动状态标记
}

// PrimaryRoleFlag 主要角色标记常量
const (
	PrimaryRoleFlagYes = "Y" // 主要角色
	PrimaryRoleFlagNo  = "N" // 次要角色
)
