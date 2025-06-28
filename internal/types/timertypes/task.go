package timertypes

import (
	"errors"
	"fmt"
	"time"
)

// TimerTask 定义任务配置和运行时信息，对应数据库表 HUB_TIMER_TASK
type TimerTask struct {
	// 主键信息
	TaskId            string `json:"taskId" db:"taskId"`
	TenantId          string `json:"tenantId" db:"tenantId"`
	
	// 任务配置信息
	TaskName          string  `json:"taskName" db:"taskName"`
	TaskDescription   *string `json:"taskDescription" db:"taskDescription"`
	TaskPriority      int     `json:"taskPriority" db:"taskPriority"`
	SchedulerId       *string `json:"schedulerId" db:"schedulerId"`
	SchedulerName     *string `json:"schedulerName" db:"schedulerName"`
	
	// 调度配置
	ScheduleType      int     `json:"scheduleType" db:"scheduleType"`
	CronExpression    *string `json:"cronExpression" db:"cronExpression"`
	IntervalSeconds   *int64  `json:"intervalSeconds" db:"intervalSeconds"`
	DelaySeconds      *int64  `json:"delaySeconds" db:"delaySeconds"`
	StartTime         *time.Time `json:"startTime" db:"startTime"`
	EndTime           *time.Time `json:"endTime" db:"endTime"`
	
	// 执行配置
	MaxRetries        int     `json:"maxRetries" db:"maxRetries"`
	RetryIntervalSeconds int64 `json:"retryIntervalSeconds" db:"retryIntervalSeconds"`
	TimeoutSeconds    int64   `json:"timeoutSeconds" db:"timeoutSeconds"`
	TaskParams        *string `json:"taskParams" db:"taskParams"`
	// -- 任务执行器配置 - 关联到具体工具配置
	ExecutorType      string  `json:"executorType" db:"executorType"`
	ToolConfigId      string  `json:"toolConfigId" db:"toolConfigId"`
	ToolConfigName    string  `json:"toolConfigName" db:"toolConfigName"`
	OperationType     string  `json:"operationType" db:"operationType"`
	OperationConfig   *string `json:"operationConfig" db:"operationConfig"`
	
	// 运行时状态
	TaskStatus        int     `json:"taskStatus" db:"taskStatus"`
	NextRunTime       *time.Time `json:"nextRunTime" db:"nextRunTime"`
	LastRunTime       *time.Time `json:"lastRunTime" db:"lastRunTime"`
	RunCount          int64   `json:"runCount" db:"runCount"`
	SuccessCount      int64   `json:"successCount" db:"successCount"`
	FailureCount      int64   `json:"failureCount" db:"failureCount"`
	
	// 最后执行结果
	LastExecutionId   *string `json:"lastExecutionId" db:"lastExecutionId"`
	LastExecutionStartTime *time.Time `json:"lastExecutionStartTime" db:"lastExecutionStartTime"`
	LastExecutionEndTime *time.Time `json:"lastExecutionEndTime" db:"lastExecutionEndTime"`
	LastExecutionDurationMs *int64 `json:"lastExecutionDurationMs" db:"lastExecutionDurationMs"`
	LastExecutionStatus *int `json:"lastExecutionStatus" db:"lastExecutionStatus"`
	LastResultSuccess *string `json:"lastResultSuccess" db:"lastResultSuccess"`
	LastErrorMessage  *string `json:"lastErrorMessage" db:"lastErrorMessage"`
	LastRetryCount    *int    `json:"lastRetryCount" db:"lastRetryCount"`
	
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
func (TimerTask) TableName() string {
	return "HUB_TIMER_TASK"
}

// Validate 验证任务配置是否有效
func (t *TimerTask) Validate() error {
	// 验证必填字段
	if t.TaskName == "" {
		return errors.New("任务名称不能为空")
	}
	
	// 验证调度类型相关字段
	switch t.ScheduleType {
	case ScheduleTypeCron:
		if t.CronExpression == nil || *t.CronExpression == "" {
			return errors.New("Cron表达式不能为空")
		}
	case ScheduleTypeInterval:
		if t.IntervalSeconds == nil || *t.IntervalSeconds <= 0 {
			return errors.New("执行间隔必须大于0")
		}
	case ScheduleTypeDelay:
		if t.DelaySeconds == nil || *t.DelaySeconds <= 0 {
			return errors.New("延迟时间必须大于0")
		}
	case ScheduleTypeOneTime:
		if t.StartTime == nil {
			return errors.New("一次性任务必须设置开始时间")
		}
	}
	
	// 验证时间范围
	if t.StartTime != nil && t.EndTime != nil && t.StartTime.After(*t.EndTime) {
		return errors.New("开始时间不能晚于结束时间")
	}
	
	return nil
}

// GetStatusName 获取任务状态名称
func (t *TimerTask) GetStatusName() string {
	switch t.TaskStatus {
	case TaskStatusPending:
		return "待执行"
	case TaskStatusRunning:
		return "运行中"
	case TaskStatusCompleted:
		return "已完成"
	case TaskStatusFailed:
		return "执行失败"
	case TaskStatusCancelled:
		return "已取消"
	default:
		return "未知状态"
	}
}

// GetPriorityName 获取任务优先级名称
func (t *TimerTask) GetPriorityName() string {
	switch t.TaskPriority {
	case TaskPriorityLow:
		return "低"
	case TaskPriorityNormal:
		return "中"
	case TaskPriorityHigh:
		return "高"
	default:
		return "未知"
	}
}

// GetScheduleTypeName 获取调度类型名称
func (t *TimerTask) GetScheduleTypeName() string {
	switch t.ScheduleType {
	case ScheduleTypeOneTime:
		return "一次性执行"
	case ScheduleTypeInterval:
		return "固定间隔"
	case ScheduleTypeCron:
		return "Cron表达式"
	case ScheduleTypeDelay:
		return "延迟执行"
	case ScheduleTypeRealTime:
		return "实时执行"
	default:
		return "未知类型"
	}
}

// IsActive 检查任务是否处于活动状态
func (t *TimerTask) IsActive() bool {
	return t.ActiveFlag == ActiveFlagYes
}

// CanExecute 检查任务是否可以执行
func (t *TimerTask) CanExecute() bool {
	// 检查基本条件
	if !t.IsActive() || t.TaskStatus == TaskStatusRunning {
		return false
	}
	
	// 检查时间范围
	now := time.Now()
	if t.StartTime != nil && now.Before(*t.StartTime) {
		return false
	}
	if t.EndTime != nil && now.After(*t.EndTime) {
		return false
	}
	
	return true
}

// UpdateRunCount 更新执行次数统计
func (t *TimerTask) UpdateRunCount(success bool) {
	t.RunCount++
	if success {
		t.SuccessCount++
	} else {
		t.FailureCount++
	}
}

// GetSuccessRate 获取成功率
func (t *TimerTask) GetSuccessRate() float64 {
	if t.RunCount == 0 {
		return 0
	}
	return float64(t.SuccessCount) / float64(t.RunCount) * 100
}

// UpdateLastExecution 更新最后执行信息
func (t *TimerTask) UpdateLastExecution(executionId string, status int, success bool, errorMsg *string, retryCount int) {
	t.LastExecutionId = &executionId
	t.LastExecutionStatus = &status
	
	if success {
		resultSuccess := ResultSuccess
		t.LastResultSuccess = &resultSuccess
	} else {
		resultFailure := ResultFailure
		t.LastResultSuccess = &resultFailure
	}
	
	t.LastErrorMessage = errorMsg
	t.LastRetryCount = &retryCount
	
	// 更新任务状态
	if status == ExecutionStatusCompleted {
		if success {
			t.TaskStatus = TaskStatusCompleted
		} else {
			t.TaskStatus = TaskStatusFailed
		}
	} else if status == ExecutionStatusCancelled {
		t.TaskStatus = TaskStatusCancelled
	}
}

// CalculateNextRunTime 计算下次执行时间
func (t *TimerTask) CalculateNextRunTime() error {
	now := time.Now()
	
	switch t.ScheduleType {
	case ScheduleTypeOneTime:
		// 一次性任务没有下次执行时间
		t.NextRunTime = nil
		return nil
		
	case ScheduleTypeInterval:
		if t.IntervalSeconds == nil {
			return errors.New("执行间隔不能为空")
		}
		
		// 如果有上次执行时间，基于上次执行时间计算
		if t.LastRunTime != nil {
			next := t.LastRunTime.Add(time.Duration(*t.IntervalSeconds) * time.Second)
			t.NextRunTime = &next
		} else if t.StartTime != nil {
			// 如果有开始时间但还未执行过，使用开始时间
			t.NextRunTime = t.StartTime
		} else {
			// 否则从当前时间开始计算
			next := now.Add(time.Duration(*t.IntervalSeconds) * time.Second)
			t.NextRunTime = &next
		}
		
	case ScheduleTypeDelay:
		if t.DelaySeconds == nil {
			return errors.New("延迟时间不能为空")
		}
		
		// 延迟任务只在第一次计算时有效
		if t.LastRunTime == nil {
			next := now.Add(time.Duration(*t.DelaySeconds) * time.Second)
			t.NextRunTime = &next
		} else {
			// 延迟任务执行一次后不再执行
			t.NextRunTime = nil
		}
		
	case ScheduleTypeCron:
		// 需要使用Cron表达式库计算下次执行时间
		// 这里简化处理，实际应使用cron解析库
		return fmt.Errorf("Cron表达式计算需要使用专门的解析库")
		
	case ScheduleTypeRealTime:
		// 实时执行任务没有固定的下次执行时间
		t.NextRunTime = nil
		
	default:
		return fmt.Errorf("不支持的调度类型: %d", t.ScheduleType)
	}
	
	// 检查是否超过结束时间
	if t.EndTime != nil && t.NextRunTime != nil && t.NextRunTime.After(*t.EndTime) {
		t.NextRunTime = nil // 已超过结束时间，不再执行
	}
	
	return nil
}

// StartExecution 开始执行任务
func (t *TimerTask) StartExecution() {
	t.TaskStatus = TaskStatusRunning
	now := time.Now()
	t.LastRunTime = &now
	t.LastExecutionStartTime = &now
}

// CompleteExecution 完成任务执行
func (t *TimerTask) CompleteExecution(success bool, errorMsg *string) {
	now := time.Now()
	t.LastExecutionEndTime = &now
	
	// 计算执行时长
	if t.LastExecutionStartTime != nil {
		duration := now.Sub(*t.LastExecutionStartTime).Milliseconds()
		t.LastExecutionDurationMs = &duration
	}
	
	// 更新执行状态
	if success {
		t.TaskStatus = TaskStatusCompleted
		resultSuccess := ResultSuccess
		t.LastResultSuccess = &resultSuccess
		executionStatus := ExecutionStatusCompleted
		t.LastExecutionStatus = &executionStatus
		t.SuccessCount++
	} else {
		t.TaskStatus = TaskStatusFailed
		resultFailure := ResultFailure
		t.LastResultSuccess = &resultFailure
		executionStatus := ExecutionStatusFailed
		t.LastExecutionStatus = &executionStatus
		t.LastErrorMessage = errorMsg
		t.FailureCount++
	}
	
	t.RunCount++
	
	// 计算下次执行时间
	_ = t.CalculateNextRunTime()
} 