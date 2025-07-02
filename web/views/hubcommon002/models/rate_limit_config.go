package models

import (
	"time"
)

// RateLimitConfig 限流配置模型，对应数据库HUB_GW_RATE_LIMIT_CONFIG表
type RateLimitConfig struct {
	TenantId             string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                                     // 租户ID，联合主键
	RateLimitConfigId    string     `json:"rateLimitConfigId" form:"rateLimitConfigId" query:"rateLimitConfigId" db:"rateLimitConfigId"`             // 限流配置ID，联合主键
	GatewayInstanceId    *string    `json:"gatewayInstanceId" form:"gatewayInstanceId" query:"gatewayInstanceId" db:"gatewayInstanceId"`             // 网关实例ID(实例级限流)
	RouteConfigId        *string    `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`                             // 路由配置ID(路由级限流)
	LimitName            string     `json:"limitName" form:"limitName" query:"limitName" db:"limitName"`                                             // 限流规则名称
	Algorithm            string     `json:"algorithm" form:"algorithm" query:"algorithm" db:"algorithm" binding:"oneof=token-bucket leaky-bucket sliding-window fixed-window none"` // 限流算法
	KeyStrategy          string     `json:"keyStrategy" form:"keyStrategy" query:"keyStrategy" db:"keyStrategy" binding:"oneof=ip user path service route"` // 限流键策略
	LimitRate            int        `json:"limitRate" form:"limitRate" query:"limitRate" db:"limitRate" binding:"min=1"`                            // 限流速率(次/秒)
	BurstCapacity        int        `json:"burstCapacity" form:"burstCapacity" query:"burstCapacity" db:"burstCapacity" binding:"min=0"`            // 突发容量
	TimeWindowSeconds    int        `json:"timeWindowSeconds" form:"timeWindowSeconds" query:"timeWindowSeconds" db:"timeWindowSeconds" binding:"min=1"` // 时间窗口(秒)
	RejectionStatusCode  int        `json:"rejectionStatusCode" form:"rejectionStatusCode" query:"rejectionStatusCode" db:"rejectionStatusCode" binding:"min=100,max=599"` // 拒绝时的HTTP状态码
	RejectionMessage     string     `json:"rejectionMessage" form:"rejectionMessage" query:"rejectionMessage" db:"rejectionMessage"`               // 拒绝时的提示消息
	ConfigPriority       int        `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`                       // 配置优先级，数值越小优先级越高
	CustomConfig         string     `json:"customConfig" form:"customConfig" query:"customConfig" db:"customConfig"`                               // 自定义配置，JSON格式
	Reserved1            *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                             // 预留字段1
	Reserved2            *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                             // 预留字段2
	Reserved3            *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                             // 预留字段3
	Reserved4            *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                             // 预留字段4
	Reserved5            *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                             // 预留字段5
	ExtProperty          *string    `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`                                     // 扩展属性，JSON格式
	AddTime              time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                     // 创建时间
	AddWho               string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                         // 创建人ID
	EditTime             time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                                 // 最后修改时间
	EditWho              string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                     // 最后修改人ID
	OprSeqFlag           string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                         // 操作序列标识
	CurrentVersion       int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                         // 当前版本号
	ActiveFlag           string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`                    // 活动状态标记(N非活动,Y活动)
	NoteText             *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                                 // 备注信息
}

// TableName 返回表名
func (RateLimitConfig) TableName() string {
	return "HUB_GW_RATE_LIMIT_CONFIG"
}
