package timertypes

import (
	"time"
)

// TimerExecutionLog 定义任务执行日志，对应数据库表 HUB_TIMER_EXECUTION_LOG
type TimerExecutionLog struct {
	// 主键信息
	ExecutionId       string `json:"executionId" db:"executionId;primaryKey"`
	TenantId          string `json:"tenantId" db:"tenantId;primaryKey"`
	TaskId            string `json:"taskId" db:"taskId"`
	
	// 任务信息（冗余）
	TaskName          *string `json:"taskName" db:"taskName"`
	SchedulerId       *string `json:"schedulerId" db:"schedulerId"`
	
	// 执行信息
	ExecutionStartTime time.Time `json:"executionStartTime" db:"executionStartTime"`
	ExecutionEndTime   *time.Time `json:"executionEndTime" db:"executionEndTime"`
	ExecutionDurationMs *int64 `json:"executionDurationMs" db:"executionDurationMs"`
	ExecutionStatus    int    `json:"executionStatus" db:"executionStatus"`
	ResultSuccess      string `json:"resultSuccess" db:"resultSuccess"`
	
	// 错误信息
	ErrorMessage      *string `json:"errorMessage" db:"errorMessage"`
	ErrorStackTrace   *string `json:"errorStackTrace" db:"errorStackTrace"`
	
	// 重试信息
	RetryCount        int    `json:"retryCount" db:"retryCount"`
	MaxRetryCount     int    `json:"maxRetryCount" db:"maxRetryCount"`
	
	// 参数和结果
	ExecutionParams   *string `json:"executionParams" db:"executionParams"`
	ExecutionResult   *string `json:"executionResult" db:"executionResult"`
	
	// 执行环境
	ExecutorServerName *string `json:"executorServerName" db:"executorServerName"`
	ExecutorServerIp   *string `json:"executorServerIp" db:"executorServerIp"`
	
	// 日志信息
	LogLevel          *string `json:"logLevel" db:"logLevel"`
	LogMessage        *string `json:"logMessage" db:"logMessage"`
	LogTimestamp      *time.Time `json:"logTimestamp" db:"logTimestamp"`
	
	// 执行上下文
	ExecutionPhase    *string `json:"executionPhase" db:"executionPhase"`
	ThreadName        *string `json:"threadName" db:"threadName"`
	ClassName         *string `json:"className" db:"className"`
	MethodName        *string `json:"methodName" db:"methodName"`
	
	// 异常信息
	ExceptionClass    *string `json:"exceptionClass" db:"exceptionClass"`
	ExceptionMessage  *string `json:"exceptionMessage" db:"exceptionMessage"`
	
	// 通用字段
	AddTime           time.Time `json:"addTime" db:"addTime"`
	AddWho            string    `json:"addWho" db:"addWho"`
	EditTime          time.Time `json:"editTime" db:"editTime"`
	EditWho           string    `json:"editWho" db:"editWho"`
	OprSeqFlag        string    `json:"oprSeqFlag" db:"oprSeqFlag"`
	CurrentVersion    int       `json:"currentVersion" db:"currentVersion"`
	ActiveFlag        string    `json:"activeFlag" db:"activeFlag"`
	NoteText          *string   `json:"noteText" db:"noteText"`
	Reserved1         *string   `json:"reserved1" db:"reserved1"`
	Reserved2         *string   `json:"reserved2" db:"reserved2"`
	Reserved3         *string   `json:"reserved3" db:"reserved3"`
}

// TableName 返回数据库表名
func (TimerExecutionLog) TableName() string {
	return "HUB_TIMER_EXECUTION_LOG"
}

// GetStatusName 获取执行状态名称
func (l *TimerExecutionLog) GetStatusName() string {
	switch l.ExecutionStatus {
	case ExecutionStatusPending:
		return "待执行"
	case ExecutionStatusRunning:
		return "运行中"
	case ExecutionStatusCompleted:
		return "已完成"
	case ExecutionStatusFailed:
		return "执行失败"
	case ExecutionStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// IsSuccess 检查执行是否成功
func (l *TimerExecutionLog) IsSuccess() bool {
	return l.ResultSuccess == ResultSuccess
}

// StartExecution 开始执行
func (l *TimerExecutionLog) StartExecution() {
	l.ExecutionStatus = ExecutionStatusRunning
	l.ExecutionStartTime = time.Now()
	if l.ExecutionPhase != nil {
		*l.ExecutionPhase = ExecutionPhaseRunning
	}
}

// CompleteExecution 完成执行
func (l *TimerExecutionLog) CompleteExecution(success bool, result *string) {
	now := time.Now()
	l.ExecutionEndTime = &now
	
	// 计算执行时长
	duration := now.Sub(l.ExecutionStartTime).Milliseconds()
	l.ExecutionDurationMs = &duration
	
	// 更新执行状态
	if success {
		l.ExecutionStatus = ExecutionStatusCompleted
		l.ResultSuccess = ResultSuccess
		l.ExecutionResult = result
		if l.ExecutionPhase != nil {
			*l.ExecutionPhase = ExecutionPhaseAfter
		}
	} else {
		l.ExecutionStatus = ExecutionStatusFailed
		l.ResultSuccess = ResultFailure
		if l.ExecutionPhase != nil {
			*l.ExecutionPhase = ExecutionPhaseAfter
		}
	}
}

// FailExecution 执行失败
func (l *TimerExecutionLog) FailExecution(errorMsg string, stackTrace *string, exceptionClass *string) {
	now := time.Now()
	l.ExecutionEndTime = &now
	
	// 计算执行时长
	duration := now.Sub(l.ExecutionStartTime).Milliseconds()
	l.ExecutionDurationMs = &duration
	
	// 更新执行状态
	l.ExecutionStatus = ExecutionStatusFailed
	l.ResultSuccess = ResultFailure
	l.ErrorMessage = &errorMsg
	l.ErrorStackTrace = stackTrace
	l.ExceptionClass = exceptionClass
	
	if l.ExecutionPhase != nil {
		*l.ExecutionPhase = ExecutionPhaseAfter
	}
}

// CancelExecution 取消执行
func (l *TimerExecutionLog) CancelExecution(reason string) {
	now := time.Now()
	l.ExecutionEndTime = &now
	
	// 计算执行时长
	duration := now.Sub(l.ExecutionStartTime).Milliseconds()
	l.ExecutionDurationMs = &duration
	
	// 更新执行状态
	l.ExecutionStatus = ExecutionStatusCancelled
	l.ResultSuccess = ResultFailure
	l.ErrorMessage = &reason
	
	if l.ExecutionPhase != nil {
		*l.ExecutionPhase = ExecutionPhaseAfter
	}
}

// IncrementRetry 增加重试次数
func (l *TimerExecutionLog) IncrementRetry() {
	l.RetryCount++
	if l.ExecutionPhase != nil {
		*l.ExecutionPhase = ExecutionPhaseRetry
	}
}

// AddLog 添加日志
func (l *TimerExecutionLog) AddLog(level string, message string) {
	now := time.Now()
	l.LogLevel = &level
	l.LogMessage = &message
	l.LogTimestamp = &now
}

// CreateInfoLog 创建信息级别日志
func CreateInfoLog(tenantId string, taskId string, taskName string, message string) *TimerExecutionLog {
	now := time.Now()
	phase := ExecutionPhaseBefore
	level := LogLevelInfo
	
	return &TimerExecutionLog{
		ExecutionId:       generateUUID(),
		TenantId:          tenantId,
		TaskId:            taskId,
		TaskName:          &taskName,
		ExecutionStartTime: now,
		ExecutionStatus:   ExecutionStatusPending,
		ResultSuccess:     ResultFailure, // 默认为失败，执行成功后再更新
		LogLevel:          &level,
		LogMessage:        &message,
		LogTimestamp:      &now,
		ExecutionPhase:    &phase,
		AddTime:           now,
		EditTime:          now,
		ActiveFlag:        ActiveFlagYes,
	}
}

// CreateErrorLog 创建错误级别日志
func CreateErrorLog(tenantId string, taskId string, taskName string, message string, exceptionClass *string, exceptionMsg *string) *TimerExecutionLog {
	now := time.Now()
	phase := ExecutionPhaseBefore
	level := LogLevelError
	
	return &TimerExecutionLog{
		ExecutionId:       generateUUID(),
		TenantId:          tenantId,
		TaskId:            taskId,
		TaskName:          &taskName,
		ExecutionStartTime: now,
		ExecutionStatus:   ExecutionStatusFailed,
		ResultSuccess:     ResultFailure,
		ErrorMessage:      &message,
		ExceptionClass:    exceptionClass,
		ExceptionMessage:  exceptionMsg,
		LogLevel:          &level,
		LogMessage:        &message,
		LogTimestamp:      &now,
		ExecutionPhase:    &phase,
		AddTime:           now,
		EditTime:          now,
		ActiveFlag:        ActiveFlagYes,
	}
}

// 生成UUID的辅助函数（实际应使用UUID库）
func generateUUID() string {
	return "EXEC_" + time.Now().Format("20060102150405") + "_" + randomString(4)
}

// 生成随机字符串的辅助函数
func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().Nanosecond()%len(letters)]
	}
	return string(b)
} 