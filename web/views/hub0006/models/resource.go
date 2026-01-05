package models

import (
	"time"
)

// Resource 权限资源模型，对应数据库HUB_AUTH_RESOURCE表
type Resource struct {
	// 主键和租户信息
	ResourceId string `json:"resourceId" form:"resourceId" query:"resourceId" db:"resourceId"` // 资源ID，主键
	TenantId   string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`         // 租户ID，用于多租户数据隔离

	// 资源基本信息
	ResourceName   string `json:"resourceName" form:"resourceName" query:"resourceName" db:"resourceName"`         // 资源名称
	ResourceCode   string `json:"resourceCode" form:"resourceCode" query:"resourceCode" db:"resourceCode"`         // 资源编码，用于程序判断
	ResourceType   string `json:"resourceType" form:"resourceType" query:"resourceType" db:"resourceType"`         // 资源类型(MODULE:模块,MENU:菜单,BUTTON:按钮,API:接口)
	ResourcePath   string `json:"resourcePath" form:"resourcePath" query:"resourcePath" db:"resourcePath"`         // 资源路径(菜单路径或API路径)
	ResourceMethod string `json:"resourceMethod" form:"resourceMethod" query:"resourceMethod" db:"resourceMethod"` // 请求方法(GET,POST,PUT,DELETE等)

	// 层级关系
	ParentResourceId string `json:"parentResourceId" form:"parentResourceId" query:"parentResourceId" db:"parentResourceId"` // 父资源ID
	ResourceLevel    int    `json:"resourceLevel" form:"resourceLevel" query:"resourceLevel" db:"resourceLevel"`             // 资源层级
	SortOrder        int    `json:"sortOrder" form:"sortOrder" query:"sortOrder" db:"sortOrder"`                             // 排序顺序

	// 显示信息
	DisplayName string `json:"displayName" form:"displayName" query:"displayName" db:"displayName"` // 显示名称
	IconClass   string `json:"iconClass" form:"iconClass" query:"iconClass" db:"iconClass"`         // 图标样式类
	Description string `json:"description" form:"description" query:"description" db:"description"` // 资源描述
	Language    string `json:"language" form:"language" query:"language" db:"language"`             // 语言标识（如：zh-CN, en-US），用于多语言支持

	// 状态信息
	ResourceStatus string `json:"resourceStatus" form:"resourceStatus" query:"resourceStatus" db:"resourceStatus"` // 资源状态(Y:启用,N:禁用)
	BuiltInFlag    string `json:"builtInFlag" form:"builtInFlag" query:"builtInFlag" db:"builtInFlag"`             // 内置资源标记(Y:内置,N:自定义)

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
func (Resource) TableName() string {
	return "HUB_AUTH_RESOURCE"
}

// ResourceStatus 资源状态常量
const (
	ResourceStatusEnabled  = "Y" // 启用
	ResourceStatusDisabled = "N" // 禁用
)

// ResourceType 资源类型常量
const (
	ResourceTypeModule = "MODULE" // 模块
	ResourceTypeMenu   = "MENU"   // 菜单
	ResourceTypeButton = "BUTTON" // 按钮
	ResourceTypeAPI    = "API"    // 接口
)

// ResourceQuery 资源查询条件，对应前端 /queryResources 的查询参数
type ResourceQuery struct {
	ResourceName     string `json:"resourceName" form:"resourceName" query:"resourceName"`             // 资源名称（模糊查询）
	ResourceCode     string `json:"resourceCode" form:"resourceCode" query:"resourceCode"`             // 资源编码（模糊查询）
	ResourceType     string `json:"resourceType" form:"resourceType" query:"resourceType"`             // 资源类型：MODULE/MENU/BUTTON/API，空表示全部
	ResourceStatus   string `json:"resourceStatus" form:"resourceStatus" query:"resourceStatus"`       // 资源状态：Y/N，空表示全部
	BuiltInFlag      string `json:"builtInFlag" form:"builtInFlag" query:"builtInFlag"`                // 内置资源标记：Y/N，空表示全部
	ActiveFlag       string `json:"activeFlag" form:"activeFlag" query:"activeFlag"`                   // 活动标记：Y-活动，N-非活动，空表示全部
	ParentResourceId string `json:"parentResourceId" form:"parentResourceId" query:"parentResourceId"` // 父资源ID，空表示全部
}
