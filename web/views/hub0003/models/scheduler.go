package hub0003models

import (
	"time"
)

// TimerScheduler 定义调度器模型，对应数据库表 HUB_TIMER_SCHEDULER
type TimerScheduler struct {
	// 主键信息
	SchedulerId        string `json:"schedulerId" form:"schedulerId" query:"schedulerId" db:"schedulerId"`
	TenantId           string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	
	// 调度器基本信息
	SchedulerName      string `json:"schedulerName" form:"schedulerName" query:"schedulerName" db:"schedulerName"`
	SchedulerInstanceId string `json:"schedulerInstanceId" form:"schedulerInstanceId" query:"schedulerInstanceId" db:"schedulerInstanceId"`
	
	// 调度器配置
	MaxWorkers         int    `json:"maxWorkers" form:"maxWorkers" query:"maxWorkers" db:"maxWorkers"`
	QueueSize          int    `json:"queueSize" form:"queueSize" query:"queueSize" db:"queueSize"`
	DefaultTimeoutSeconds int64 `json:"defaultTimeoutSeconds" form:"defaultTimeoutSeconds" query:"defaultTimeoutSeconds" db:"defaultTimeoutSeconds"`
	DefaultRetries     int    `json:"defaultRetries" form:"defaultRetries" query:"defaultRetries" db:"defaultRetries"`
	
	// 调度器状态
	SchedulerStatus    int    `json:"schedulerStatus" form:"schedulerStatus" query:"schedulerStatus" db:"schedulerStatus"`
	LastStartTime      *time.Time `json:"lastStartTime" form:"lastStartTime" query:"lastStartTime" db:"lastStartTime"`
	LastStopTime       *time.Time `json:"lastStopTime" form:"lastStopTime" query:"lastStopTime" db:"lastStopTime"`
	
	// 服务器信息
	ServerName         *string `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`
	ServerIp           *string `json:"serverIp" form:"serverIp" query:"serverIp" db:"serverIp"`
	ServerPort         *int    `json:"serverPort" form:"serverPort" query:"serverPort" db:"serverPort"`
	
	// 监控信息
	TotalTaskCount     int    `json:"totalTaskCount" form:"totalTaskCount" query:"totalTaskCount" db:"totalTaskCount"`
	RunningTaskCount   int    `json:"runningTaskCount" form:"runningTaskCount" query:"runningTaskCount" db:"runningTaskCount"`
	LastHeartbeatTime  *time.Time `json:"lastHeartbeatTime" form:"lastHeartbeatTime" query:"lastHeartbeatTime" db:"lastHeartbeatTime"`
	
	// 通用字段
	AddTime            time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho             string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime           time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho            string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag         string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion     int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag         string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText           *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	Reserved1          *string   `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`
	Reserved2          *string   `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`
	Reserved3          *string   `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`
}

// TableName 返回数据库表名
func (TimerScheduler) TableName() string {
	return "HUB_TIMER_SCHEDULER"
} 