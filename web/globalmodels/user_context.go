package globalmodels

import (
	"time"
)

// UserContext 用户上下文信息
// 
// 结构说明:
//   统一的用户上下文数据结构，用于JWT认证和Session认证的用户信息存储
//   包含用户基本信息、认证信息和会话状态等字段
//
// 使用场景:
//   - JWT中间件验证后的用户信息存储
//   - Session中间件验证后的用户信息存储
//   - 控制器中获取当前登录用户信息
//   - 权限验证和业务逻辑处理
//
// 注意事项:
//   - SessionId字段仅在Session认证时有值
//   - LoginTime在JWT认证时表示令牌签发时间，在Session认证时表示会话创建时间
//   - 所有字段都支持JSON序列化，便于日志记录和调试
type UserContext struct {
	// 基本用户信息
	UserId      string     `json:"userId"`      // 用户ID - 用户在系统中的唯一标识
	TenantId    string     `json:"tenantId"`    // 租户ID - 多租户系统中的租户标识
	UserName    string     `json:"userName"`    // 用户名 - 用于显示和日志记录
	RealName    string     `json:"realName"`    // 真实姓名 - 用户的真实姓名或显示名称
	DeptId      string     `json:"deptId"`      // 部门ID - 用户所属部门的标识
	
	// 扩展用户信息
	Email       string     `json:"email"`       // 邮箱 - 用户邮箱地址
	Mobile      string     `json:"mobile"`      // 手机号 - 用户手机号码
	Avatar      string     `json:"avatar"`      // 头像 - 用户头像URL或路径
	
	// 认证和会话信息
	SessionId    string     `json:"sessionId"`    // Session ID - 仅Session认证时有值
	LoginTime    *time.Time `json:"loginTime"`    // 登录时间 - 认证时间或会话创建时间
	LastActivity *time.Time `json:"lastActivity"` // 最后活动时间 - 最近一次验证session的时间
	ExpireAt     *time.Time `json:"expireAt"`     // 过期时间 - session失效的绝对时间
	
	// 环境信息
	ClientIP    string     `json:"clientIP"`    // 客户端IP - 用于安全验证和审计
	UserAgent   string     `json:"userAgent"`   // 用户代理 - 浏览器和操作系统信息
}
