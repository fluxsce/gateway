package timertypes

// 通用常量
const (
	// 活动状态标记
	ActiveFlagYes = "Y" // 活动
	ActiveFlagNo  = "N" // 非活动

	// 成功标记
	ResultSuccess    = "Y" // 成功
	ResultFailure    = "N" // 失败
)

// 调度器状态常量
const (
	SchedulerStatusStopped  = 1 // 停止
	SchedulerStatusRunning  = 2 // 运行中
	SchedulerStatusPaused   = 3 // 暂停
)

// 任务优先级常量
const (
	TaskPriorityLow     = 1 // 低优先级
	TaskPriorityNormal  = 2 // 普通优先级
	TaskPriorityHigh    = 3 // 高优先级
)

// 调度类型常量
const (
	ScheduleTypeOneTime   = 1 // 一次性执行
	ScheduleTypeInterval  = 2 // 固定间隔
	ScheduleTypeCron      = 3 // Cron表达式
	ScheduleTypeDelay     = 4 // 延迟执行
	ScheduleTypeRealTime  = 5 // 实时执行
)

// 任务状态常量
const (
	TaskStatusPending   = 1 // 待执行
	TaskStatusRunning   = 2 // 运行中
	TaskStatusCompleted = 3 // 已完成
	TaskStatusFailed    = 4 // 执行失败
	TaskStatusCancelled = 5 // 已取消
)

// 执行状态常量
const (
	ExecutionStatusPending   = 1 // 待执行
	ExecutionStatusRunning   = 2 // 运行中
	ExecutionStatusCompleted = 3 // 已完成
	ExecutionStatusFailed    = 4 // 执行失败
	ExecutionStatusCancelled = 5 // 已取消
)

// 日志级别常量
const (
	LogLevelDebug = "DEBUG" // 调试
	LogLevelInfo  = "INFO"  // 信息
	LogLevelWarn  = "WARN"  // 警告
	LogLevelError = "ERROR" // 错误
)

// 执行阶段常量
const (
	ExecutionPhaseBefore  = "BEFORE_EXECUTE" // 执行前
	ExecutionPhaseRunning = "EXECUTING"      // 执行中
	ExecutionPhaseAfter   = "AFTER_EXECUTE"  // 执行后
	ExecutionPhaseRetry   = "RETRY"          // 重试
) 