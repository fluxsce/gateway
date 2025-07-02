package models

import (
	"time"
)

// AuthConfig 认证配置模型
type AuthConfig struct {
	TenantId          string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                                     // 租户ID
	AuthConfigId      string     `json:"authConfigId" form:"authConfigId" query:"authConfigId" db:"authConfigId"`                                                   // 认证配置ID
	GatewayInstanceId *string    `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"`                               // 网关实例ID(实例级认证)
	RouteConfigId     *string    `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                                               // 路由配置ID(路由级认证)
	AuthName          string     `json:"authName" form:"authName" query:"authName" db:"authName"`                                                                     // 认证配置名称
	AuthType          string     `json:"authType" form:"authType" query:"authType" db:"authType" binding:"oneof=JWT API_KEY OAUTH2 BASIC"`                         // 认证类型(JWT,API_KEY,OAUTH2,BASIC)
	AuthStrategy      string     `json:"authStrategy" form:"authStrategy" query:"authStrategy" db:"authStrategy" binding:"oneof=REQUIRED OPTIONAL DISABLED"`       // 认证策略(REQUIRED,OPTIONAL,DISABLED)
	AuthConfig        string     `json:"authConfig" form:"authConfig" query:"authConfig" db:"authConfig"`                                                           // 认证参数配置,JSON格式
	ExemptPaths       *string    `json:"exemptPaths" form:"exemptPaths" query:"exemptPaths" db:"exemptPaths"`                                                       // 豁免路径列表,JSON数组格式
	ExemptHeaders     *string    `json:"exemptHeaders" form:"exemptHeaders" query:"exemptHeaders" db:"exemptHeaders"`                                               // 豁免请求头列表,JSON数组格式
	FailureStatusCode int        `json:"failureStatusCode" form:"failureStatusCode" query:"failureStatusCode" db:"failureStatusCode" binding:"min=100,max=599"`    // 认证失败状态码
	FailureMessage    string     `json:"failureMessage" form:"failureMessage" query:"failureMessage" db:"failureMessage"`                                         // 认证失败提示消息
	ConfigPriority    int        `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`                                         // 配置优先级,数值越小优先级越高
	Reserved1         *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                                               // 预留字段1
	Reserved2         *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                                               // 预留字段2
	Reserved3         *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                                               // 预留字段3
	Reserved4         *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                                               // 预留字段4
	Reserved5         *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                                               // 预留字段5
	ExtProperty       *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                                                       // 扩展属性,JSON格式
	AddTime           time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                                       // 创建时间
	AddWho            string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                                           // 创建人ID
	EditTime          time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                                                   // 最后修改时间
	EditWho           string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                                       // 最后修改人ID
	OprSeqFlag        string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                                           // 操作序列标识
	CurrentVersion    int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                                           // 当前版本号
	ActiveFlag        string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`                                      // 活动状态标记(N非活动,Y活动)
	NoteText          *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                                                   // 备注信息
}

// TableName 返回表名
func (AuthConfig) TableName() string {
	return "HUB_GW_AUTH_CONFIG"
}
