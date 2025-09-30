package permission

import "time"

// UserRole 用户角色信息
type UserRole struct {
	RoleId     string     `json:"roleId" db:"roleId"`         // 角色ID，主键
	RoleName   string     `json:"roleName" db:"roleName"`     // 角色名称
	RoleCode   string     `json:"roleCode" db:"roleCode"`     // 角色编码，用于程序判断
	RoleType   string     `json:"roleType" db:"roleType"`     // 角色类型(SYSTEM:系统角色,CUSTOM:自定义角色)
	RoleLevel  int        `json:"roleLevel" db:"roleLevel"`   // 角色级别，数字越小权限越高
	DataScope  string     `json:"dataScope" db:"dataScope"`   // 数据权限范围(ALL:全部,TENANT:租户,DEPT:部门,SELF:个人)
	ExpireTime *time.Time `json:"expireTime" db:"expireTime"` // 角色过期时间，NULL表示永不过期
}

// UserPermission 用户权限信息
type UserPermission struct {
	ResourceId     string     `json:"resourceId" db:"resourceId"`         // 资源ID，主键
	ResourceCode   string     `json:"resourceCode" db:"resourceCode"`     // 资源编码，用于程序判断
	ResourceName   string     `json:"resourceName" db:"resourceName"`     // 资源名称
	ResourceType   string     `json:"resourceType" db:"resourceType"`     // 资源类型(MODULE:模块,MENU:菜单,BUTTON:按钮,API:接口)
	ResourcePath   string     `json:"resourcePath" db:"resourcePath"`     // 资源路径(菜单路径或API路径)
	ResourceMethod string     `json:"resourceMethod" db:"resourceMethod"` // 请求方法(GET,POST,PUT,DELETE等)
	ModuleCode     string     `json:"moduleCode" db:"moduleCode"`         // 所属模块编码
	PermissionType string     `json:"permissionType" db:"permissionType"` // 权限类型(ALLOW:允许,DENY:拒绝)
	ExpireTime     *time.Time `json:"expireTime" db:"expireTime"`         // 权限过期时间，NULL表示永不过期
}

// DataPermission 数据权限信息
type DataPermission struct {
	DataPermissionId     string     `json:"dataPermissionId" db:"dataPermissionId"`         // 数据权限ID，主键
	UserId               string     `json:"userId" db:"userId"`                             // 用户ID，为空表示角色级权限
	RoleId               string     `json:"roleId" db:"roleId"`                             // 角色ID，为空表示用户级权限
	ResourceType         string     `json:"resourceType" db:"resourceType"`                 // 资源类型(TABLE:数据表,API:接口,MODULE:模块)
	ResourceCode         string     `json:"resourceCode" db:"resourceCode"`                 // 资源编码
	PermissionScope      string     `json:"permissionScope" db:"permissionScope"`           // 权限范围(ALL:全部,TENANT:租户,DEPT:部门,SELF:个人,CUSTOM:自定义)
	ScopeValue           string     `json:"scopeValue" db:"scopeValue"`                     // 权限范围值，JSON格式
	FilterCondition      string     `json:"filterCondition" db:"filterCondition"`           // 过滤条件，SQL WHERE条件
	ColumnPermissions    string     `json:"columnPermissions" db:"columnPermissions"`       // 字段权限，JSON格式
	OperationPermissions string     `json:"operationPermissions" db:"operationPermissions"` // 操作权限(read:只读,write:读写,delete:删除)
	ExpireTime           *time.Time `json:"expireTime" db:"expireTime"`                     // 过期时间
}

// PermissionCheckRequest 权限检查请求参数
type PermissionCheckRequest struct {
	UserId       string `json:"userId" binding:"required"`   // 用户ID，必填
	TenantId     string `json:"tenantId" binding:"required"` // 租户ID，必填
	ModuleCode   string `json:"moduleCode,omitempty"`        // 模块编码，可选
	ResourceCode string `json:"resourceCode,omitempty"`      // 资源编码，可选
	ButtonCode   string `json:"buttonCode,omitempty"`        // 按钮编码，可选
	ResourcePath string `json:"resourcePath,omitempty"`      // 资源路径，可选
	Method       string `json:"method,omitempty"`            // HTTP请求方法，可选
}

// PermissionCheckResponse 权限检查响应结果
type PermissionCheckResponse struct {
	HasPermission bool                   `json:"hasPermission"`         // 是否有权限
	Permissions   []string               `json:"permissions,omitempty"` // 权限列表
	DataScope     string                 `json:"dataScope,omitempty"`   // 数据权限范围
	Message       string                 `json:"message,omitempty"`     // 响应消息
	Details       map[string]interface{} `json:"details,omitempty"`     // 详细信息
}

// ModulePermission 模块权限信息
type ModulePermission struct {
	ModuleCode string     `json:"moduleCode"`           // 模块编码
	ModuleName string     `json:"moduleName"`           // 模块名称
	HasAccess  bool       `json:"hasAccess"`            // 是否有访问权限
	Buttons    []string   `json:"buttons,omitempty"`    // 按钮权限列表
	DataScope  string     `json:"dataScope,omitempty"`  // 数据权限范围
	ExpireTime *time.Time `json:"expireTime,omitempty"` // 过期时间
}

// ButtonPermission 按钮权限信息
type ButtonPermission struct {
	ButtonCode   string     `json:"buttonCode"`             // 按钮编码
	ButtonName   string     `json:"buttonName"`             // 按钮名称
	ResourcePath string     `json:"resourcePath,omitempty"` // 资源路径
	Method       string     `json:"method,omitempty"`       // HTTP请求方法
	HasAccess    bool       `json:"hasAccess"`              // 是否有访问权限
	ExpireTime   *time.Time `json:"expireTime,omitempty"`   // 过期时间
}

// UserPermissionSummary 用户权限汇总信息
type UserPermissionSummary struct {
	UserId         string             `json:"userId"`         // 用户ID
	TenantId       string             `json:"tenantId"`       // 租户ID
	Roles          []UserRole         `json:"roles"`          // 用户角色列表
	Modules        []ModulePermission `json:"modules"`        // 模块权限列表
	DataScope      string             `json:"dataScope"`      // 数据权限范围
	LastUpdateTime time.Time          `json:"lastUpdateTime"` // 最后更新时间
}

// PermissionFilter 权限过滤条件
type PermissionFilter struct {
	UserId       string   `json:"userId,omitempty"`       // 用户ID，可选
	TenantId     string   `json:"tenantId"`               // 租户ID，必填
	RoleCodes    []string `json:"roleCodes,omitempty"`    // 角色编码列表，可选
	ModuleCodes  []string `json:"moduleCodes,omitempty"`  // 模块编码列表，可选
	ResourceType string   `json:"resourceType,omitempty"` // 资源类型，可选
	OnlyActive   bool     `json:"onlyActive"`             // 只查询活跃数据
}

// DataScopeType 数据权限范围类型
type DataScopeType string

const (
	DataScopeAll    DataScopeType = "ALL"    // 全部数据
	DataScopeTenant DataScopeType = "TENANT" // 租户数据
	DataScopeDept   DataScopeType = "DEPT"   // 部门数据
	DataScopeSelf   DataScopeType = "SELF"   // 个人数据
	DataScopeCustom DataScopeType = "CUSTOM" // 自定义数据
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourceTypeModule ResourceType = "MODULE" // 模块
	ResourceTypeMenu   ResourceType = "MENU"   // 菜单
	ResourceTypeButton ResourceType = "BUTTON" // 按钮
	ResourceTypeAPI    ResourceType = "API"    // 接口
)

// PermissionType 权限类型
type PermissionType string

const (
	PermissionTypeAllow PermissionType = "ALLOW" // 允许
	PermissionTypeDeny  PermissionType = "DENY"  // 拒绝
)

// RoleType 角色类型
type RoleType string

const (
	RoleTypeSystem RoleType = "SYSTEM" // 系统角色
	RoleTypeCustom RoleType = "CUSTOM" // 自定义角色
)
