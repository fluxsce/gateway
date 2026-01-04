package models

// LoginRequest 登录请求
type LoginRequest struct {
	UserId      string `json:"userId" form:"userId" query:"userId" binding:"required"`       // 用户ID
	Password    string `json:"password" form:"password" query:"password" binding:"required"` // 密码
	TenantId    string `json:"tenantId" form:"tenantId" query:"tenantId"`                    // 租户ID
	CaptchaId   string `json:"captchaId" form:"captchaId" query:"captchaId"`                 // 验证码ID（可选）
	CaptchaCode string `json:"captchaCode" form:"captchaCode" query:"captchaCode"`           // 验证码（可选）
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token        string `json:"token"`        // JWT令牌
	RefreshToken string `json:"refreshToken"` // 刷新令牌
	SessionId    string `json:"sessionId"`    // Session ID
	UserId       string `json:"userId"`       // 用户ID
	UserName     string `json:"userName"`     // 用户名
	RealName     string `json:"realName"`     // 真实姓名
	TenantId     string `json:"tenantId"`     // 租户ID
	DeptId       string `json:"deptId"`       // 部门ID
	Avatar       string `json:"avatar"`       // 头像
	ExpireAt     int64  `json:"expireAt"`     // Session过期时间戳
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

// CaptchaRequest 获取验证码请求
type CaptchaRequest struct {
	Type   string `json:"type" form:"type" query:"type"`       // 验证码类型：random(随机数)、sms(短信，扩展用)
	Mobile string `json:"mobile" form:"mobile" query:"mobile"` // 手机号（短信验证码时使用）
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaId string `json:"captchaId"` // 验证码ID，用于后续验证
	Code      string `json:"code"`      // 验证码内容（随机数时返回，短信时不返回）
	ExpireAt  int64  `json:"expireAt"`  // 过期时间戳
}

// VerifyCaptchaRequest 验证验证码请求
type VerifyCaptchaRequest struct {
	CaptchaId string `json:"captchaId" form:"captchaId" binding:"required"` // 验证码ID
	Code      string `json:"code" form:"code" binding:"required"`           // 用户输入的验证码
}

// ModulePermission 模块权限信息
type ModulePermission struct {
	ResourceId       string `json:"resourceId"`       // 资源ID
	ResourceCode     string `json:"resourceCode"`     // 资源编码
	ResourceName     string `json:"resourceName"`     // 资源名称
	DisplayName      string `json:"displayName"`      // 显示名称
	ResourcePath     string `json:"resourcePath"`     // 资源路径
	IconClass        string `json:"iconClass"`        // 图标样式类
	Description      string `json:"description"`      // 资源描述
	ResourceLevel    int    `json:"resourceLevel"`    // 资源层级
	SortOrder        int    `json:"sortOrder"`        // 排序顺序
	ParentResourceId string `json:"parentResourceId"` // 父资源ID
}

// ButtonPermission 按钮权限信息
type ButtonPermission struct {
	ResourceId       string `json:"resourceId"`       // 资源ID
	ResourceCode     string `json:"resourceCode"`     // 资源编码
	ResourceName     string `json:"resourceName"`     // 资源名称
	DisplayName      string `json:"displayName"`      // 显示名称
	ResourcePath     string `json:"resourcePath"`     // 资源路径
	ResourceMethod   string `json:"resourceMethod"`   // 请求方法
	ParentResourceId string `json:"parentResourceId"` // 父资源ID（所属模块）
	Description      string `json:"description"`      // 资源描述
}

// UserPermissionResponse 用户权限响应
type UserPermissionResponse struct {
	Modules []ModulePermission `json:"modules"` // 模块权限列表
	Buttons []ButtonPermission `json:"buttons"` // 按钮权限列表
}
