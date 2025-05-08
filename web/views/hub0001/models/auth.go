package models

// LoginRequest 登录请求
type LoginRequest struct {
	UserId   string `json:"userId" form:"userId" query:"userId" binding:"required"`       // 用户ID
	Password string `json:"password" form:"password" query:"password" binding:"required"` // 密码
	TenantId string `json:"tenantId" form:"tenantId" query:"tenantId" binding:"required"` // 租户ID
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`        // JWT令牌
	RefreshToken string `json:"refreshToken"` // 刷新令牌
	UserId       string `json:"userId"`       // 用户ID
	UserName     string `json:"userName"`     // 用户名
	RealName     string `json:"realName"`     // 真实姓名
	TenantId     string `json:"tenantId"`     // 租户ID
	DeptId       string `json:"deptId"`       // 部门ID
	Avatar       string `json:"avatar"`       // 头像
}

// TokenData JWT令牌数据
type TokenData struct {
	UserId   string `json:"userId"`   // 用户ID
	TenantId string `json:"tenantId"` // 租户ID
	UserName string `json:"userName"` // 用户名
	RealName string `json:"realName"` // 真实姓名
	DeptId   string `json:"deptId"`   // 部门ID
}

// PasswordChangeRequest 密码修改请求
type PasswordChangeRequest struct {
	UserId      string `json:"userId" form:"userId" binding:"required"`           // 用户ID
	OldPassword string `json:"oldPassword" form:"oldPassword" binding:"required"` // 旧密码
	NewPassword string `json:"newPassword" form:"newPassword" binding:"required"` // 新密码
}

// RefreshTokenRequest 刷新令牌请求
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" form:"refreshToken" binding:"required"` // 刷新令牌
}

// RefreshTokenResponse 刷新令牌响应
type RefreshTokenResponse struct {
	Token        string `json:"token"`        // 访问令牌
	RefreshToken string `json:"refreshToken"` // 刷新令牌
}
