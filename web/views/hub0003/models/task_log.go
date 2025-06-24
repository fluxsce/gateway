package models

import (
	"time"
)

// TaskLog 定时任务日志模型，对应数据库HUB_TIMER_TASK_LOG表
type TaskLog struct {
	TaskLogId         string     `json:"taskLogId" form:"taskLogId" query:"taskLogId" db:"taskLogId"`                           // 任务日志ID，主键
	TenantId          string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                               // 租户ID，联合主键
	TaskId            string     `json:"taskId" form:"taskId" query:"taskId" db:"taskId" binding:"required"`                   // 任务唯一标识
	TaskResultId      *string    `json:"taskResultId" form:"taskResultId" query:"taskResultId" db:"taskResultId"`               // 关联任务结果ID
	
	// 日志信息
	LogLevel          string     `json:"logLevel" form:"logLevel" query:"logLevel" db:"logLevel" binding:"required,oneof=DEBUG INFO WARN ERROR"` // 日志级别(DEBUG,INFO,WARN,ERROR)
	LogMessage        string     `json:"logMessage" form:"logMessage" query:"logMessage" db:"logMessage" binding:"required"`   // 日志消息内容
	LogTimestamp      time.Time  `json:"logTimestamp" form:"logTimestamp" query:"logTimestamp" db:"logTimestamp"`             // 日志时间戳
	
	// 执行上下文
	ExecutionPhase    *string    `json:"executionPhase" form:"executionPhase" query:"executionPhase" db:"executionPhase"`       // 执行阶段(BEFORE_EXECUTE,EXECUTING,AFTER_EXECUTE,RETRY)
	ThreadName        *string    `json:"threadName" form:"threadName" query:"threadName" db:"threadName"`                       // 执行线程名称
	ClassName         *string    `json:"className" form:"className" query:"className" db:"className"`                           // 执行类名
	MethodName        *string    `json:"methodName" form:"methodName" query:"methodName" db:"methodName"`                       // 执行方法名
	
	// 异常信息
	ExceptionClass    *string    `json:"exceptionClass" form:"exceptionClass" query:"exceptionClass" db:"exceptionClass"`       // 异常类名
	ExceptionMessage  *string    `json:"exceptionMessage" form:"exceptionMessage" query:"exceptionMessage" db:"exceptionMessage"` // 异常消息
	
	// 预留字段
	Reserved1         *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                           // 预留字段1
	Reserved2         *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                           // 预留字段2
	Reserved3         *string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                           // 预留字段3
	Reserved4         *string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                           // 预留字段4
	Reserved5         *string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                           // 预留字段5
	Reserved6         *string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                           // 预留字段6
	Reserved7         *string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                           // 预留字段7
	Reserved8         *string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                           // 预留字段8
	Reserved9         *string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                           // 预留字段9
	Reserved10        *string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                       // 预留字段10
	
	// 通用字段
	AddTime           time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                   // 创建时间
	AddWho            string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                       // 创建人ID
	EditTime          time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                               // 最后修改时间
	EditWho           string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                   // 最后修改人ID
	OprSeqFlag        string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                       // 操作序列标识
	CurrentVersion    int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`       // 当前版本号
	ActiveFlag        string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`  // 活动状态标记(N非活动,Y活动)
	NoteText          *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                               // 备注信息
}

// TableName 返回表名
func (TaskLog) TableName() string {
	return "HUB_TIMER_TASK_LOG"
} 