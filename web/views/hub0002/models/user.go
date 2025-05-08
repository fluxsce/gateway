package models

import (
	"time"
)

// User 用户模型，对应数据库HUB_USER表
type User struct {
	UserId          string     `json:"userId" db:"userId"`                   // 用户ID，联合主键
	TenantId        string     `json:"tenantId" db:"tenantId"`               // 租户ID，联合主键
	UserName        string     `json:"userName" db:"userName"`               // 用户名，登录账号
	Password        string     `json:"password" db:"password"`               // 密码，加密存储
	RealName        string     `json:"realName" db:"realName"`               // 真实姓名
	DeptId          string     `json:"deptId" db:"deptId"`                   // 所属部门ID
	Email           string     `json:"email" db:"email"`                     // 电子邮箱
	Mobile          string     `json:"mobile" db:"mobile"`                   // 手机号码
	Avatar          string     `json:"avatar" db:"avatar"`                   // 头像URL
	Gender          int        `json:"gender" db:"gender"`                   // 性别：1-男，2-女，0-未知
	StatusFlag      string     `json:"statusFlag" db:"statusFlag"`           // 状态：Y-启用，N-禁用
	DeptAdminFlag   string     `json:"deptAdminFlag" db:"deptAdminFlag"`     // 是否部门管理员：Y-是，N-否
	TenantAdminFlag string     `json:"tenantAdminFlag" db:"tenantAdminFlag"` // 是否租户管理员：Y-是，N-否
	UserExpireDate  time.Time  `json:"userExpireDate" db:"userExpireDate"`   // 用户过期时间
	LastLoginTime   *time.Time `json:"lastLoginTime" db:"lastLoginTime"`     // 最后登录时间
	LastLoginIp     string     `json:"lastLoginIp" db:"lastLoginIp"`         // 最后登录IP
	AddTime         time.Time  `json:"addTime" db:"addTime"`                 // 创建时间
	AddWho          string     `json:"addWho" db:"addWho"`                   // 创建人
	EditTime        time.Time  `json:"editTime" db:"editTime"`               // 修改时间
	EditWho         string     `json:"editWho" db:"editWho"`                 // 修改人
	OprSeqFlag      string     `json:"oprSeqFlag" db:"oprSeqFlag"`           // 操作序列标识
	CurrentVersion  int        `json:"currentVersion" db:"currentVersion"`   // 当前版本号
	ActiveFlag      string     `json:"activeFlag" db:"activeFlag"`           // 活动状态标记：Y-活动，N-非活动
	NoteText        string     `json:"noteText" db:"noteText"`               // 备注信息
}

// TableName 返回表名
func (User) TableName() string {
	return "HUB_USER"
}
