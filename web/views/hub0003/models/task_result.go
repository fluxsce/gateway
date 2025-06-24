package models

import (
	"time"
)

// TaskResult 定时任务执行结果模型，对应数据库HUB_TIMER_TASK_RESULT表
type TaskResult struct {
	TaskResultId            string     `json:"taskResultId" form:"taskResultId" query:"taskResultId" db:"taskResultId"`                             // 任务结果ID，主键
	TenantId                string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                             // 租户ID，联合主键
	TaskId                  string     `json:"taskId" form:"taskId" query:"taskId" db:"taskId" binding:"required"`                                 // 任务唯一标识
	TaskConfigId            string     `json:"taskConfigId" form:"taskConfigId" query:"taskConfigId" db:"taskConfigId" binding:"required"`         // 关联任务配置ID
	
	// 执行信息
	ExecutionStartTime      time.Time  `json:"executionStartTime" form:"executionStartTime" query:"executionStartTime" db:"executionStartTime"`   // 执行开始时间
	ExecutionEndTime        *time.Time `json:"executionEndTime" form:"executionEndTime" query:"executionEndTime" db:"executionEndTime"`           // 执行结束时间
	ExecutionDurationMs     *int64     `json:"executionDurationMs" form:"executionDurationMs" query:"executionDurationMs" db:"executionDurationMs"` // 执行耗时毫秒数
	ExecutionStatus         int        `json:"executionStatus" form:"executionStatus" query:"executionStatus" db:"executionStatus" binding:"min=1,max=5"` // 执行状态(1待执行,2运行中,3已完成,4执行失败,5已取消)
	
	// 结果信息
	ResultSuccess           string     `json:"resultSuccess" form:"resultSuccess" query:"resultSuccess" db:"resultSuccess" binding:"oneof=Y N"`   // 执行是否成功(N失败,Y成功)
	ErrorMessage            *string    `json:"errorMessage" form:"errorMessage" query:"errorMessage" db:"errorMessage"`                           // 错误信息
	ErrorStackTrace         *string    `json:"errorStackTrace" form:"errorStackTrace" query:"errorStackTrace" db:"errorStackTrace"`               // 错误堆栈信息
	
	// 重试信息
	RetryCount              int        `json:"retryCount" form:"retryCount" query:"retryCount" db:"retryCount"`                                   // 重试次数
	MaxRetryCount           int        `json:"maxRetryCount" form:"maxRetryCount" query:"maxRetryCount" db:"maxRetryCount"`                       // 最大重试次数
	
	// 执行参数和结果
	ExecutionParams         *string    `json:"executionParams" form:"executionParams" query:"executionParams" db:"executionParams"`               // 执行参数，JSON格式
	ExecutionResult         *string    `json:"executionResult" form:"executionResult" query:"executionResult" db:"executionResult"`               // 执行结果，JSON格式
	
	// 服务器信息
	ExecutorServerName      *string    `json:"executorServerName" form:"executorServerName" query:"executorServerName" db:"executorServerName"`   // 执行服务器名称
	ExecutorServerIp        *string    `json:"executorServerIp" form:"executorServerIp" query:"executorServerIp" db:"executorServerIp"`           // 执行服务器IP地址
	
	// 预留字段
	Reserved1               *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                       // 预留字段1
	Reserved2               *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                       // 预留字段2
	Reserved3               *string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                       // 预留字段3
	Reserved4               *string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                       // 预留字段4
	Reserved5               *string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                       // 预留字段5
	Reserved6               *string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                                       // 预留字段6
	Reserved7               *string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                                       // 预留字段7
	Reserved8               *string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                                       // 预留字段8
	Reserved9               *string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                                       // 预留字段9
	Reserved10              *string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                                   // 预留字段10
	
	// 通用字段
	AddTime                 time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                               // 创建时间
	AddWho                  string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                   // 创建人ID
	EditTime                time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                           // 最后修改时间
	EditWho                 string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                               // 最后修改人ID
	OprSeqFlag              string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                   // 操作序列标识
	CurrentVersion          int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                   // 当前版本号
	ActiveFlag              string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`              // 活动状态标记(N非活动,Y活动)
	NoteText                *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                           // 备注信息
}

// TableName 返回表名
func (TaskResult) TableName() string {
	return "HUB_TIMER_TASK_RESULT"
} 