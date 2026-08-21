package models

import "time"

// SysSetting 环境设置分组行，对应表 HUB_SYS_SETTING。
// 每个租户每个 groupCode 一行，content 为该分组的 JSON。
type SysSetting struct {
	TenantId       string    `json:"tenantId" db:"tenantId"`
	GroupCode      string    `json:"groupCode" db:"groupCode"`
	Content        string    `json:"content" db:"content"`
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"`
	NoteText       string    `json:"noteText" db:"noteText"`
	ExtProperty    string    `json:"extProperty" db:"extProperty"`
}

// TableName 返回表名。
func (SysSetting) TableName() string {
	return "HUB_SYS_SETTING"
}

// SaveSettingRequest 保存单个分组。currentVersion 为读取时的版本，0 表示尚未落库。
type SaveSettingRequest struct {
	GroupCode             string `json:"groupCode" form:"groupCode"`
	CurrentVersion        int    `json:"currentVersion" form:"currentVersion"`
	AuditLogDays          int    `json:"auditLogDays" form:"auditLogDays"`
	TaskLogDays           int    `json:"taskLogDays" form:"taskLogDays"`
	AlertLogDays          int    `json:"alertLogDays" form:"alertLogDays"`
	ClusterEventDays      int    `json:"clusterEventDays" form:"clusterEventDays"`
	MetricsDays           int    `json:"metricsDays" form:"metricsDays"`
	GatewayLogDefaultDays int    `json:"gatewayLogDefaultDays" form:"gatewayLogDefaultDays"`
	Enabled               bool   `json:"enabled" form:"enabled"`
	IntervalMinutes       int    `json:"intervalMinutes" form:"intervalMinutes"`
	StartTime             string `json:"startTime" form:"startTime"`
	RequestTimeoutSeconds int    `json:"requestTimeoutSeconds" form:"requestTimeoutSeconds"`
	SessionExpireHours    int    `json:"sessionExpireHours" form:"sessionExpireHours"`
}

// EnvSettingsResponse 环境设置页一次拉取的全部分组。
type EnvSettingsResponse struct {
	Retention    RetentionView    `json:"retention"`
	RetentionJob RetentionJobView `json:"retentionJob"`
	WebTimeout   WebTimeoutView   `json:"webTimeout"`
}

// RetentionView 归档策略回显，带乐观锁版本。
type RetentionView struct {
	AuditLogDays          int `json:"auditLogDays"`
	TaskLogDays           int `json:"taskLogDays"`
	AlertLogDays          int `json:"alertLogDays"`
	ClusterEventDays      int `json:"clusterEventDays"`
	MetricsDays           int `json:"metricsDays"`
	GatewayLogDefaultDays int `json:"gatewayLogDefaultDays"`
	CurrentVersion        int `json:"currentVersion"`
}

// RetentionJobView 归档任务回显，带乐观锁版本。
type RetentionJobView struct {
	Enabled         bool   `json:"enabled"`
	IntervalMinutes int    `json:"intervalMinutes"`
	StartTime       string `json:"startTime"`
	CurrentVersion  int    `json:"currentVersion"`
}

// WebTimeoutView Web 超时回显，带乐观锁版本。
type WebTimeoutView struct {
	RequestTimeoutSeconds int `json:"requestTimeoutSeconds"`
	SessionExpireHours    int `json:"sessionExpireHours"`
	CurrentVersion        int `json:"currentVersion"`
}
