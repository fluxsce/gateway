package hub0003models

import (
	"time"
)

// TimerExecutionLog 定义执行日志模型，对应数据库表 HUB_TIMER_EXECUTION_LOG
type TimerExecutionLog struct {
	// 主键信息
	ExecutionId       string `json:"executionId" form:"executionId" query:"executionId" db:"executionId"`
	TenantId          string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`
	TaskId            string `json:"taskId" form:"taskId" query:"taskId" db:"taskId"`
	
	// 任务信息（冗余）
	TaskName          *string `json:"taskName" form:"taskName" query:"taskName" db:"taskName"`
	SchedulerId       *string `json:"schedulerId" form:"schedulerId" query:"schedulerId" db:"schedulerId"`
	
	// 执行信息
	ExecutionStartTime time.Time `json:"executionStartTime" form:"executionStartTime" query:"executionStartTime" db:"executionStartTime"`
	ExecutionEndTime   *time.Time `json:"executionEndTime" form:"executionEndTime" query:"executionEndTime" db:"executionEndTime"`
	ExecutionDurationMs *int64 `json:"executionDurationMs" form:"executionDurationMs" query:"executionDurationMs" db:"executionDurationMs"`
	ExecutionStatus    int     `json:"executionStatus" form:"executionStatus" query:"executionStatus" db:"executionStatus"`
	ResultSuccess      string  `json:"resultSuccess" form:"resultSuccess" query:"resultSuccess" db:"resultSuccess"`
	
	// 错误信息
	ErrorMessage      *string `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`
	ErrorStackTrace   *string `json:"errorStackTrace" form:"errorStackTrace" query:"errorStackTrace" db:"errorStackTrace"`
	
	// 重试信息
	RetryCount        int     `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`
	MaxRetryCount     int     `json:"maxRetryCount" form:"maxRetryCount" query:"maxRetryCount" db:"maxRetryCount"`
	
	// 参数和结果
	ExecutionParams   *string `json:"executionParams" form:"executionParams" query:"executionParams" db:"executionParams"`
	ExecutionResult   *string `json:"executionResult" form:"executionResult" query:"executionResult" db:"executionResult"`
	
	// 执行环境
	ExecutorServerName *string `json:"executorServerName" form:"executorServerName" query:"executorServerName" db:"executorServerName"`
	ExecutorServerIp   *string `json:"executorServerIp" form:"executorServerIp" query:"executorServerIp" db:"executorServerIp"`
	
	// 日志信息
	LogLevel          *string `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel"`
	LogMessage        *string `json:"logMessage" form:"logMessage" query:"logMessage" db:"logMessage"`
	LogTimestamp      *time.Time `json:"logTimestamp" form:"logTimestamp" query:"logTimestamp" db:"logTimestamp"`
	
	// 执行上下文
	ExecutionPhase    *string `json:"executionPhase" form:"executionPhase" query:"executionPhase" db:"executionPhase"`
	ThreadName        *string `json:"threadName" form:"threadName" query:"threadName" db:"threadName"`
	ClassName         *string `json:"className" form:"className" query:"className" db:"className"`
	MethodName        *string `json:"methodName" form:"methodName" query:"methodName" db:"methodName"`
	
	// 异常信息
	ExceptionClass    *string `json:"exceptionClass" form:"exceptionClass" query:"exceptionClass" db:"exceptionClass"`
	ExceptionMessage  *string `json:"exceptionMessage" form:"exceptionMessage" query:"exceptionMessage" db:"exceptionMessage"`
	
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
func (TimerExecutionLog) TableName() string {
	return "HUB_TIMER_EXECUTION_LOG"
} 