package models

import (
	"time"
)

// TaskConfig 定时任务配置模型，对应数据库HUB_TIMER_TASK_CONFIG表
type TaskConfig struct {
	TaskConfigId      string     `json:"taskConfigId" form:"taskConfigId" query:"taskConfigId" db:"taskConfigId"`                   // 任务配置ID，主键
	TenantId          string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                   // 租户ID，联合主键
	TaskId            string     `json:"taskId" form:"taskId" query:"taskId" db:"taskId" binding:"required"`                       // 任务唯一标识，业务ID
	TaskName          string     `json:"taskName" form:"taskName" query:"taskName" db:"taskName" binding:"required"`               // 任务名称
	TaskDescription   *string    `json:"taskDescription" form:"taskDescription" query:"taskDescription" db:"taskDescription"`     // 任务描述
	TaskPriority      int        `json:"taskPriority" form:"taskPriority" query:"taskPriority" db:"taskPriority" binding:"min=1,max=3"` // 任务优先级(1低优先级,2普通优先级,3高优先级)
	
	// 调度配置
	ScheduleType      int        `json:"scheduleType" form:"scheduleType" query:"scheduleType" db:"scheduleType" binding:"required,min=1,max=5"` // 调度类型(1一次性执行,2固定间隔,3Cron表达式,4延迟执行,5实时执行)
	CronExpression    *string    `json:"cronExpression" form:"cronExpression" query:"cronExpression" db:"cronExpression"`         // Cron表达式，scheduleType=3时必填
	IntervalSeconds   *int64     `json:"intervalSeconds" form:"intervalSeconds" query:"intervalSeconds" db:"intervalSeconds"`     // 执行间隔秒数，scheduleType=2时必填
	DelaySeconds      *int64     `json:"delaySeconds" form:"delaySeconds" query:"delaySeconds" db:"delaySeconds"`                 // 延迟秒数，scheduleType=4时必填
	StartTime         *time.Time `json:"startTime" form:"startTime" query:"startTime" db:"startTime"`                             // 任务开始时间
	EndTime           *time.Time `json:"endTime" form:"endTime" query:"endTime" db:"endTime"`                                     // 任务结束时间
	
	// 执行配置
	MaxRetries         int       `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries" binding:"min=0"`          // 最大重试次数
	RetryIntervalSeconds int64   `json:"retryIntervalSeconds" form:"retryIntervalSeconds" query:"retryIntervalSeconds" db:"retryIntervalSeconds" binding:"min=1"` // 重试间隔秒数
	TimeoutSeconds     int64     `json:"timeoutSeconds" form:"timeoutSeconds" query:"timeoutSeconds" db:"timeoutSeconds" binding:"min=1"` // 执行超时时间秒数
	
	// 任务参数
	TaskParams        *string    `json:"taskParams" form:"taskParams" query:"taskParams" db:"taskParams"`                         // 任务参数，JSON格式存储
	
	// 预留字段
	Reserved1         *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                             // 预留字段1
	Reserved2         *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                             // 预留字段2
	Reserved3         *string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                             // 预留字段3
	Reserved4         *string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                             // 预留字段4
	Reserved5         *string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                             // 预留字段5
	Reserved6         *string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                             // 预留字段6
	Reserved7         *string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                             // 预留字段7
	Reserved8         *string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                             // 预留字段8
	Reserved9         *string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                             // 预留字段9
	Reserved10        *string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                         // 预留字段10
	
	// 通用字段
	AddTime           time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                     // 创建时间
	AddWho            string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                         // 创建人ID
	EditTime          time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                 // 最后修改时间
	EditWho           string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                     // 最后修改人ID
	OprSeqFlag        string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                         // 操作序列标识
	CurrentVersion    int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`         // 当前版本号
	ActiveFlag        string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`    // 活动状态标记(N非活动,Y活动)
	NoteText          *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                 // 备注信息
}

// TableName 返回表名
func (TaskConfig) TableName() string {
	return "HUB_TIMER_TASK_CONFIG"
} 