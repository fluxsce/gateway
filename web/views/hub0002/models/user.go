package models

import (
	"time"
)

// User 用户模型，对应数据库HUB_USER表
type User struct {
	UserId          string     `json:"userId" form:"userId" query:"userId" db:"userId"`                                                                 // 用户ID，联合主键
	TenantId        string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                         // 租户ID，联合主键
	UserName        string     `json:"userName" form:"userName" query:"userName" db:"userName"`                                                         // 用户名，登录账号
	Password        string     `json:"password" form:"password" query:"password" db:"password"`                                                         // 密码，加密存储
	RealName        string     `json:"realName" form:"realName" query:"realName" db:"realName"`                                                         // 真实姓名
	DeptId          string     `json:"deptId" form:"deptId" query:"deptId" db:"deptId"`                                                                 // 所属部门ID
	Email           string     `json:"email" form:"email" query:"email" db:"email"`                                                                     // 电子邮箱
	Mobile          string     `json:"mobile" form:"mobile" query:"mobile" db:"mobile"`                                                                 // 手机号码
	Avatar          string     `json:"avatar" form:"avatar" query:"avatar" db:"avatar"`                                                                 // 头像URL
	Gender          int        `json:"gender" form:"gender" query:"gender" db:"gender"`                                                                 // 性别：1-男，2-女，0-未知
	StatusFlag      string     `json:"statusFlag" form:"statusFlag" query:"statusFlag" db:"statusFlag"`                                                 // 状态：Y-启用，N-禁用
	DeptAdminFlag   string     `json:"deptAdminFlag" form:"deptAdminFlag" query:"deptAdminFlag" db:"deptAdminFlag"`                                     // 是否部门管理员：Y-是，N-否
	TenantAdminFlag string     `json:"tenantAdminFlag" form:"tenantAdminFlag" query:"tenantAdminFlag" db:"tenantAdminFlag"`                             // 是否租户管理员：Y-是，N-否
	UserExpireDate  time.Time  `json:"userExpireDate,time_format=2006-01-02 15:04:00" form:"userExpireDate" query:"userExpireDate" db:"userExpireDate"` // 用户过期时间
	LastLoginTime   *time.Time `json:"lastLoginTime" form:"lastLoginTime" query:"lastLoginTime" db:"lastLoginTime"`                                     // 最后登录时间
	LastLoginIp     string     `json:"lastLoginIp" form:"lastLoginIp" query:"lastLoginIp" db:"lastLoginIp"`                                             // 最后登录IP
	AddTime         time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                             // 创建时间
	AddWho          string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                                 // 创建人
	EditTime        time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                                         // 修改时间
	EditWho         string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                             // 修改人
	OprSeqFlag      string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                                 // 操作序列标识
	CurrentVersion  int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                                 // 当前版本号
	ActiveFlag      string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                                                 // 活动状态标记：Y-活动，N-非活动
	NoteText        string     `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                                         // 备注信息
}

// TableName 返回表名
func (User) TableName() string {
	return "HUB_USER"
}
