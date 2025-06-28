package hub0003models

import (
	"time"
)

// TimerTask 定义任务模型，对应数据库表 HUB_TIMER_TASK
type TimerTask struct {
	// 主键信息
	TaskId            string `json:"taskId" form:"taskId" query:"taskId" db:"taskId"`
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	
	// 任务配置信息
	TaskName          string  `json:"taskName" form:"taskName" query:"taskName" db:"taskName"`
	TaskDescription   *string `json:"taskDescription" form:"taskDescription" query:"taskDescription" db:"taskDescription"`
	TaskPriority      int     `json:"taskPriority" form:"taskPriority" query:"taskPriority" db:"taskPriority"`
	SchedulerId       *string `json:"schedulerId" form:"schedulerId" query:"schedulerId" db:"schedulerId"`
	SchedulerName     *string `json:"schedulerName" form:"schedulerName" query:"schedulerName" db:"schedulerName"`
	
	// 调度配置
	ScheduleType      int     `json:"scheduleType" form:"scheduleType" query:"scheduleType" db:"scheduleType"`
	CronExpression    *string `json:"cronExpression" form:"cronExpression" query:"cronExpression" db:"cronExpression"`
	IntervalSeconds   *int64  `json:"intervalSeconds" form:"intervalSeconds" query:"intervalSeconds" db:"intervalSeconds"`
	DelaySeconds      *int64  `json:"delaySeconds" form:"delaySeconds" query:"delaySeconds" db:"delaySeconds"`
	StartTime         *time.Time `json:"startTime" form:"startTime" query:"startTime" db:"startTime"`
	EndTime           *time.Time `json:"endTime" form:"endTime" query:"endTime" db:"endTime"`
	
	// 执行配置
	MaxRetries        int     `json:"maxRetries" form:"maxRetries" query:"maxRetries" db:"maxRetries"`
	RetryIntervalSeconds int64 `json:"retryIntervalSeconds" form:"retryIntervalSeconds" query:"retryIntervalSeconds" db:"retryIntervalSeconds"`
	TimeoutSeconds    int64   `json:"timeoutSeconds" form:"timeoutSeconds" query:"timeoutSeconds" db:"timeoutSeconds"`
	TaskParams        *string `json:"taskParams" form:"taskParams" query:"taskParams" db:"taskParams"`
	
	// 任务执行器配置 - 关联到具体工具配置
	ExecutorType      *string `json:"executorType" form:"executorType" query:"executorType" db:"executorType"`
	ToolConfigId      *string `json:"toolConfigId" form:"toolConfigId" query:"toolConfigId" db:"toolConfigId"`
	ToolConfigName    *string `json:"toolConfigName" form:"toolConfigName" query:"toolConfigName" db:"toolConfigName"`
	OperationType     *string `json:"operationType" form:"operationType" query:"operationType" db:"operationType"`
	OperationConfig   *string `json:"operationConfig" form:"operationConfig" query:"operationConfig" db:"operationConfig"`
	
	// 运行时状态
	TaskStatus        int     `json:"taskStatus" form:"taskStatus" query:"taskStatus" db:"taskStatus"`
	NextRunTime       *time.Time `json:"nextRunTime" form:"nextRunTime" query:"nextRunTime" db:"nextRunTime"`
	LastRunTime       *time.Time `json:"lastRunTime" form:"lastRunTime" query:"lastRunTime" db:"lastRunTime"`
	RunCount          int64   `json:"runCount" form:"runCount" query:"runCount" db:"runCount"`
	SuccessCount      int64   `json:"successCount" form:"successCount" query:"successCount" db:"successCount"`
	FailureCount      int64   `json:"failureCount" form:"failureCount" query:"failureCount" db:"failureCount"`
	
	// 最后执行结果
	LastExecutionId   *string `json:"lastExecutionId" form:"lastExecutionId" query:"lastExecutionId" db:"lastExecutionId"`
	LastExecutionStartTime *time.Time `json:"lastExecutionStartTime" form:"lastExecutionStartTime" query:"lastExecutionStartTime" db:"lastExecutionStartTime"`
	LastExecutionEndTime *time.Time `json:"lastExecutionEndTime" form:"lastExecutionEndTime" query:"lastExecutionEndTime" db:"lastExecutionEndTime"`
	LastExecutionDurationMs *int64 `json:"lastExecutionDurationMs" form:"lastExecutionDurationMs" query:"lastExecutionDurationMs" db:"lastExecutionDurationMs"`
	LastExecutionStatus *int `json:"lastExecutionStatus" form:"lastExecutionStatus" query:"lastExecutionStatus" db:"lastExecutionStatus"`
	LastResultSuccess *string `json:"lastResultSuccess" form:"lastResultSuccess" query:"lastResultSuccess" db:"lastResultSuccess"`
	LastErrorMessage  *string `json:"lastErrorMessage" form:"lastErrorMessage" query:"lastErrorMessage" db:"lastErrorMessage"`
	LastRetryCount    *int    `json:"lastRetryCount" form:"lastRetryCount" query:"lastRetryCount" db:"lastRetryCount"`
	
	// 通用字段
	AddTime           time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`
	AddWho            string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`
	EditTime          time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`
	EditWho           string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`
	OprSeqFlag        string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion    int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`
	ActiveFlag        string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`
	NoteText          *string   `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`
	ExtProperty       *string   `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"`
	Reserved1         *string   `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`
	Reserved2         *string   `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`
	Reserved3         *string   `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`
	Reserved4         *string   `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`
	Reserved5         *string   `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`
	Reserved6         *string   `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`
	Reserved7         *string   `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`
	Reserved8         *string   `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`
	Reserved9         *string   `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`
	Reserved10        *string   `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`
}

// TableName 返回数据库表名
func (TimerTask) TableName() string {
	return "HUB_TIMER_TASK"
} 