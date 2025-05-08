package auth

import (
	"time"
)

// UserContext 用户上下文信息
type UserContext struct {
	UserId      string     `json:"userId"`      // 用户ID
	TenantId    string     `json:"tenantId"`    // 租户ID
	UserName    string     `json:"userName"`    // 用户名
	RealName    string     `json:"realName"`    // 真实姓名
	DeptId      string     `json:"deptId"`      // 部门ID
	Roles       []string   `json:"roles"`       // 角色列表
	Permissions []string   `json:"permissions"` // 权限列表
	LoginTime   *time.Time `json:"loginTime"`   // 登录时间
}
