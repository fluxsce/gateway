package logwrite

import (
	"time"
)

// ExecutionLogLevel 日志级别
type ExecutionLogLevel string

const (
	LogLevelDebug ExecutionLogLevel = "DEBUG"
	LogLevelInfo  ExecutionLogLevel = "INFO"
	LogLevelWarn  ExecutionLogLevel = "WARN"
	LogLevelError ExecutionLogLevel = "ERROR"
)

// ExecutionStatus 执行状态
type ExecutionStatus int

const (
	StatusPending    ExecutionStatus = 1 // 待执行
	StatusRunning    ExecutionStatus = 2 // 运行中
	StatusCompleted  ExecutionStatus = 3 // 已完成
	StatusFailed     ExecutionStatus = 4 // 执行失败
	StatusCancelled  ExecutionStatus = 5 // 已取消
)

// ExecutionPhase 执行阶段
type ExecutionPhase string

const (
	PhaseBeforeExecute ExecutionPhase = "BEFORE_EXECUTE"
	PhaseExecuting     ExecutionPhase = "EXECUTING"
	PhaseAfterExecute  ExecutionPhase = "AFTER_EXECUTE"
	PhaseRetry         ExecutionPhase = "RETRY"
)

// TimerExecutionLog 任务执行日志结构体，对应 HUB_TIMER_EXECUTION_LOG 表
type TimerExecutionLog struct {
	// 主键信息
	ExecutionId string `json:"executionId" db:"executionId"`
	TenantId    string `json:"tenantId" db:"tenantId"`
	TaskId      string `json:"taskId" db:"taskId"`

	// 任务信息（冗余）
	TaskName    *string `json:"taskName,omitempty" db:"taskName"`
	SchedulerId *string `json:"schedulerId,omitempty" db:"schedulerId"`

	// 执行信息
	ExecutionStartTime  time.Time        `json:"executionStartTime" db:"executionStartTime"`
	ExecutionEndTime    *time.Time       `json:"executionEndTime,omitempty" db:"executionEndTime"`
	ExecutionDurationMs *int64           `json:"executionDurationMs,omitempty" db:"executionDurationMs"`
	ExecutionStatus     ExecutionStatus  `json:"executionStatus" db:"executionStatus"`
	ResultSuccess       string           `json:"resultSuccess" db:"resultSuccess"` // Y/N

	// 错误信息
	ErrorMessage    *string `json:"errorMessage,omitempty" db:"errorMessage"`
	ErrorStackTrace *string `json:"errorStackTrace,omitempty" db:"errorStackTrace"`

	// 重试信息
	RetryCount    int `json:"retryCount" db:"retryCount"`
	MaxRetryCount int `json:"maxRetryCount" db:"maxRetryCount"`

	// 参数和结果
	ExecutionParams *string `json:"executionParams,omitempty" db:"executionParams"`
	ExecutionResult *string `json:"executionResult,omitempty" db:"executionResult"`

	// 执行环境
	ExecutorServerName *string `json:"executorServerName,omitempty" db:"executorServerName"`
	ExecutorServerIp   *string `json:"executorServerIp,omitempty" db:"executorServerIp"`

	// 日志信息
	LogLevel     *ExecutionLogLevel `json:"logLevel,omitempty" db:"logLevel"`
	LogMessage   *string            `json:"logMessage,omitempty" db:"logMessage"`
	LogTimestamp *time.Time         `json:"logTimestamp,omitempty" db:"logTimestamp"`

	// 执行上下文
	ExecutionPhase *ExecutionPhase `json:"executionPhase,omitempty" db:"executionPhase"`
	ThreadName     *string         `json:"threadName,omitempty" db:"threadName"`
	ClassName      *string         `json:"className,omitempty" db:"className"`
	MethodName     *string         `json:"methodName,omitempty" db:"methodName"`

	// 异常信息
	ExceptionClass   *string `json:"exceptionClass,omitempty" db:"exceptionClass"`
	ExceptionMessage *string `json:"exceptionMessage,omitempty" db:"exceptionMessage"`

	// 通用字段
	AddTime        time.Time `json:"addTime" db:"addTime"`
	AddWho         string    `json:"addWho" db:"addWho"`
	EditTime       time.Time `json:"editTime" db:"editTime"`
	EditWho        string    `json:"editWho" db:"editWho"`
	OprSeqFlag     string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag     string    `json:"activeFlag" db:"activeFlag"` // Y/N
	NoteText       *string   `json:"noteText,omitempty" db:"noteText"`
	ExtProperty    *string   `json:"extProperty,omitempty" db:"extProperty"`

	// 预留字段
	Reserved1  *string `json:"reserved1,omitempty" db:"reserved1"`
	Reserved2  *string `json:"reserved2,omitempty" db:"reserved2"`
	Reserved3  *string `json:"reserved3,omitempty" db:"reserved3"`
	Reserved4  *string `json:"reserved4,omitempty" db:"reserved4"`
	Reserved5  *string `json:"reserved5,omitempty" db:"reserved5"`
	Reserved6  *string `json:"reserved6,omitempty" db:"reserved6"`
	Reserved7  *string `json:"reserved7,omitempty" db:"reserved7"`
	Reserved8  *string `json:"reserved8,omitempty" db:"reserved8"`
	Reserved9  *string `json:"reserved9,omitempty" db:"reserved9"`
	Reserved10 *string `json:"reserved10,omitempty" db:"reserved10"`
}

// TableName 实现 Model 接口
func (t *TimerExecutionLog) TableName() string {
	return "HUB_TIMER_EXECUTION_LOG"
}

// PrimaryKey 实现 Model 接口
func (t *TimerExecutionLog) PrimaryKey() string {
	return "executionId"
} 