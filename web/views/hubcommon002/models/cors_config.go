package models

import (
	"time"
)

// CorsConfig CORS配置模型，对应数据库HUB_GATEWAY_CORS_CONFIG表
type CorsConfig struct {
	TenantId         string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                               // 租户ID，联合主键
	CorsConfigId     string     `json:"corsConfigId" form:"corsConfigId" query:"corsConfigId" db:"corsConfigId"`             // CORS配置ID，联合主键
	GatewayInstanceId *string   `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"` // 网关实例ID(实例级CORS)
	RouteConfigId    *string    `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`         // 路由配置ID(路由级CORS)
	ConfigName       string     `json:"configName" form:"configName" query:"configName" db:"configName"`                     // 配置名称
	AllowOrigins     *string    `json:"allowOrigins" form:"allowOrigins" query:"allowOrigins" db:"allowOrigins"`             // 允许的源，JSON数组格式
	AllowMethods     string     `json:"allowMethods" form:"allowMethods" query:"allowMethods" db:"allowMethods"`             // 允许的HTTP方法，逗号分隔
	AllowHeaders     *string    `json:"allowHeaders" form:"allowHeaders" query:"allowHeaders" db:"allowHeaders"`             // 允许的请求头，JSON数组格式
	ExposeHeaders    *string    `json:"exposeHeaders" form:"exposeHeaders" query:"exposeHeaders" db:"exposeHeaders"`         // 暴露的响应头，JSON数组格式
	AllowCredentials string     `json:"allowCredentials" form:"allowCredentials" query:"allowCredentials" db:"allowCredentials"` // 是否允许携带凭证(N否,Y是)
	MaxAgeSeconds    int        `json:"maxAgeSeconds" form:"maxAgeSeconds" query:"maxAgeSeconds" db:"maxAgeSeconds"`         // 预检请求缓存时间(秒)
	ConfigPriority   int        `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`     // 配置优先级，数值越小优先级越高
	Reserved1        *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                         // 预留字段1
	Reserved2        *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                         // 预留字段2
	Reserved3        *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                         // 预留字段3
	Reserved4        *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                         // 预留字段4
	Reserved5        *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                         // 预留字段5
	ExtProperty      *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                 // 扩展属性，JSON格式
	AddTime          time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                 // 创建时间
	AddWho           string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                     // 创建人ID
	EditTime         time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                             // 最后修改时间
	EditWho          string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                 // 最后修改人ID
	OprSeqFlag       string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                     // 操作序列标识
	CurrentVersion   int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`     // 当前版本号
	ActiveFlag       string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                     // 活动状态标记(N非活动,Y活动)
	NoteText         *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                             // 备注信息
}

// TableName 返回表名
func (CorsConfig) TableName() string {
	return "HUB_GATEWAY_CORS_CONFIG"
}
