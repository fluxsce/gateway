package models

import (
	"time"
)

// SchedulerConfig 调度器配置模型，对应数据库HUB_TIMER_SCHEDULER_CONFIG表
type SchedulerConfig struct {
	SchedulerConfigId       string     `json:"schedulerConfigId" form:"schedulerConfigId" query:"schedulerConfigId" db:"schedulerConfigId"`     // 调度器配置ID，主键
	TenantId                string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                         // 租户ID，联合主键
	SchedulerName           string     `json:"schedulerName" form:"schedulerName" query:"schedulerName" db:"schedulerName" binding:"required"`   // 调度器名称
	SchedulerInstanceId     string     `json:"schedulerInstanceId" form:"schedulerInstanceId" query:"schedulerInstanceId" db:"schedulerInstanceId" binding:"required"` // 调度器实例ID，用于集群环境区分
	
	// 调度器配置
	MaxWorkers              int        `json:"maxWorkers" form:"maxWorkers" query:"maxWorkers" db:"maxWorkers" binding:"min=1"`                 // 最大工作线程数
	QueueSize               int        `json:"queueSize" form:"queueSize" query:"queueSize" db:"queueSize" binding:"min=1"`                   // 任务队列大小
	DefaultTimeoutSeconds   int64      `json:"defaultTimeoutSeconds" form:"defaultTimeoutSeconds" query:"defaultTimeoutSeconds" db:"defaultTimeoutSeconds" binding:"min=1"` // 默认超时时间秒数
	DefaultRetries          int        `json:"defaultRetries" form:"defaultRetries" query:"defaultRetries" db:"defaultRetries" binding:"min=0"` // 默认重试次数
	
	// 调度器状态
	SchedulerStatus         int        `json:"schedulerStatus" form:"schedulerStatus" query:"schedulerStatus" db:"schedulerStatus" binding:"min=1,max=3"` // 调度器状态(1停止,2运行中,3暂停)
	LastStartTime           *time.Time `json:"lastStartTime" form:"lastStartTime" query:"lastStartTime" db:"lastStartTime"`                   // 最后启动时间
	LastStopTime            *time.Time `json:"lastStopTime" form:"lastStopTime" query:"lastStopTime" db:"lastStopTime"`                       // 最后停止时间
	
	// 服务器信息
	ServerName              *string    `json:"serverName" form:"serverName" query:"serverName" db:"serverName"`                               // 服务器名称
	ServerIp                *string    `json:"serverIp" form:"serverIp" query:"serverIp" db:"serverIp"`                                       // 服务器IP地址
	ServerPort              *int       `json:"serverPort" form:"serverPort" query:"serverPort" db:"serverPort"`                               // 服务器端口
	
	// 监控信息
	TotalTaskCount          int        `json:"totalTaskCount" form:"totalTaskCount" query:"totalTaskCount" db:"totalTaskCount"`               // 总任务数
	RunningTaskCount        int        `json:"runningTaskCount" form:"runningTaskCount" query:"runningTaskCount" db:"runningTaskCount"`       // 运行中任务数
	LastHeartbeatTime       *time.Time `json:"lastHeartbeatTime" form:"lastHeartbeatTime" query:"lastHeartbeatTime" db:"lastHeartbeatTime"`   // 最后心跳时间
	
	// 预留字段
	Reserved1               *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                   // 预留字段1
	Reserved2               *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                   // 预留字段2
	Reserved3               *string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                   // 预留字段3
	Reserved4               *string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                   // 预留字段4
	Reserved5               *string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                   // 预留字段5
	Reserved6               *string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                                   // 预留字段6
	Reserved7               *string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                                   // 预留字段7
	Reserved8               *string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                                   // 预留字段8
	Reserved9               *string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                                   // 预留字段9
	Reserved10              *string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                               // 预留字段10
	
	// 通用字段
	AddTime                 time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                           // 创建时间
	AddWho                  string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                               // 创建人ID
	EditTime                time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                       // 最后修改时间
	EditWho                 string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                           // 最后修改人ID
	OprSeqFlag              string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                               // 操作序列标识
	CurrentVersion          int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`               // 当前版本号
	ActiveFlag              string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`          // 活动状态标记(N非活动,Y活动)
	NoteText                *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                       // 备注信息
}

// TableName 返回表名
func (SchedulerConfig) TableName() string {
	return "HUB_TIMER_SCHEDULER_CONFIG"
} 