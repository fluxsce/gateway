package models

import (
	"time"
)

// CircuitBreakerConfig 熔断配置模型，对应 HUB_GW_CIRCUIT_BREAKER_CONFIG。
// 服务级配置通过 targetServiceId 关联服务定义；routeConfigId 预留给路由级。
type CircuitBreakerConfig struct {
	TenantId               string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	CircuitBreakerConfigId string     `json:"circuitBreakerConfigId" form:"circuitBreakerConfigId" query:"circuitBreakerConfigId" db:"circuitBreakerConfigId"`
	RouteConfigId          string     `json:"routeConfigId" form:"routeConfigId" query:"routeConfigId" db:"routeConfigId"`
	TargetServiceId        string     `json:"targetServiceId" form:"targetServiceId" query:"targetServiceId" db:"targetServiceId"`
	BreakerName            string     `json:"breakerName" form:"breakerName" query:"breakerName" db:"breakerName"`
	KeyStrategy            string     `json:"-" form:"-" query:"-" db:"keyStrategy"` // 库列 NOT NULL，运行时固定按节点摘除，不对外暴露策略
	ErrorRatePercent       int        `json:"errorRatePercent" form:"errorRatePercent" query:"errorRatePercent" db:"errorRatePercent"`
	MinimumRequests        int        `json:"minimumRequests" form:"minimumRequests" query:"minimumRequests" db:"minimumRequests"`
	HalfOpenMaxRequests    int        `json:"halfOpenMaxRequests" form:"halfOpenMaxRequests" query:"halfOpenMaxRequests" db:"halfOpenMaxRequests"`
	SlowCallThreshold      int64      `json:"slowCallThreshold" form:"slowCallThreshold" query:"slowCallThreshold" db:"slowCallThreshold"`
	SlowCallRatePercent    int        `json:"slowCallRatePercent" form:"slowCallRatePercent" query:"slowCallRatePercent" db:"slowCallRatePercent"`
	OpenTimeoutSeconds     int64      `json:"openTimeoutSeconds" form:"openTimeoutSeconds" query:"openTimeoutSeconds" db:"openTimeoutSeconds"`
	WindowSizeSeconds      int64      `json:"windowSizeSeconds" form:"windowSizeSeconds" query:"windowSizeSeconds" db:"windowSizeSeconds"`
	ErrorStatusCode        int        `json:"errorStatusCode" form:"errorStatusCode" query:"errorStatusCode" db:"errorStatusCode"`
	ErrorMessage           string     `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`
	StorageType            string     `json:"storageType" form:"storageType" query:"storageType" db:"storageType"`
	StorageConfig          string     `json:"storageConfig" form:"storageConfig" query:"storageConfig" db:"storageConfig"`
	ConfigPriority         int        `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"`
	Reserved1              string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`
	Reserved2              string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`
	Reserved3              *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`
	Reserved4              *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`
	Reserved5              *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`
	ExtProperty            string     `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`
	AddTime                time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho                 string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime               time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho                string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag             string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion         int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag             string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText               string     `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
}

// TableName 返回表名。
func (CircuitBreakerConfig) TableName() string {
	return "HUB_GW_CIRCUIT_BREAKER_CONFIG"
}
