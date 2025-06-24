package models

import (
	"time"
)

// TaskInfo 定时任务运行时信息模型，对应数据库HUB_TIMER_TASK_INFO表
type TaskInfo struct {
	TaskInfoId                 string     `json:"taskInfoId" form:"taskInfoId" query:"taskInfoId" db:"taskInfoId"`                                       // 任务信息ID，主键
	TenantId                   string     `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                                               // 租户ID，联合主键
	TaskConfigId               string     `json:"taskConfigId" form:"taskConfigId" query:"taskConfigId" db:"taskConfigId" binding:"required"`           // 关联任务配置ID
	TaskId                     string     `json:"taskId" form:"taskId" query:"taskId" db:"taskId" binding:"required"`                                   // 任务唯一标识
	
	// 状态信息
	TaskStatus                 int        `json:"taskStatus" form:"taskStatus" query:"taskStatus" db:"taskStatus" binding:"min=1,max=5"`               // 任务状态(1待执行,2运行中,3已完成,4执行失败,5已取消)
	NextRunTime               *time.Time `json:"nextRunTime" form:"nextRunTime" query:"nextRunTime" db:"nextRunTime"`                                 // 下次执行时间
	LastRunTime               *time.Time `json:"lastRunTime" form:"lastRunTime" query:"lastRunTime" db:"lastRunTime"`                                 // 上次执行时间
	
	// 统计信息
	RunCount                   int64      `json:"runCount" form:"runCount" query:"runCount" db:"runCount"`                                             // 执行次数
	SuccessCount               int64      `json:"successCount" form:"successCount" query:"successCount" db:"successCount"`                             // 成功次数
	FailureCount               int64      `json:"failureCount" form:"failureCount" query:"failureCount" db:"failureCount"`                             // 失败次数
	
	// 最后执行结果
	LastResultId              *string    `json:"lastResultId" form:"lastResultId" query:"lastResultId" db:"lastResultId"`                             // 最后执行结果ID
	LastExecutionDurationMs   *int64     `json:"lastExecutionDurationMs" form:"lastExecutionDurationMs" query:"lastExecutionDurationMs" db:"lastExecutionDurationMs"` // 最后执行耗时毫秒数
	LastErrorMessage          *string    `json:"lastErrorMessage" form:"lastErrorMessage" query:"lastErrorMessage" db:"lastErrorMessage"`             // 最后错误信息
	
	// 预留字段
	Reserved1                 *string    `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"`                                         // 预留字段1
	Reserved2                 *string    `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"`                                         // 预留字段2
	Reserved3                 *string    `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"`                                         // 预留字段3
	Reserved4                 *string    `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"`                                         // 预留字段4
	Reserved5                 *string    `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"`                                         // 预留字段5
	Reserved6                 *string    `json:"reserved6" form:"reserved6" query:"reserved6" db:"reserved6"`                                         // 预留字段6
	Reserved7                 *string    `json:"reserved7" form:"reserved7" query:"reserved7" db:"reserved7"`                                         // 预留字段7
	Reserved8                 *string    `json:"reserved8" form:"reserved8" query:"reserved8" db:"reserved8"`                                         // 预留字段8
	Reserved9                 *string    `json:"reserved9" form:"reserved9" query:"reserved9" db:"reserved9"`                                         // 预留字段9
	Reserved10                *string    `json:"reserved10" form:"reserved10" query:"reserved10" db:"reserved10"`                                     // 预留字段10
	
	// 通用字段
	AddTime                   time.Time  `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                                                 // 创建时间
	AddWho                    string     `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                                     // 创建人ID
	EditTime                  time.Time  `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                                             // 最后修改时间
	EditWho                   string     `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                                                 // 最后修改人ID
	OprSeqFlag                string     `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                                     // 操作序列标识
	CurrentVersion            int        `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"`                     // 当前版本号
	ActiveFlag                string     `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag" binding:"oneof=Y N"`                // 活动状态标记(N非活动,Y活动)
	NoteText                  *string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                                             // 备注信息
}

// TableName 返回表名
func (TaskInfo) TableName() string {
	return "HUB_TIMER_TASK_INFO"
} 