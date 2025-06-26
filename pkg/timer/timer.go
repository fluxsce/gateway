// Package timer 提供了一个功能完整的定时任务调度组件
// 支持多种调度策略：Cron表达式、固定间隔、延迟执行、一次性执行
package timer

import (
	"context"
	"time"
)

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending   TaskStatus = 1 // 待执行
	TaskStatusRunning   TaskStatus = 2 // 运行中
	TaskStatusCompleted TaskStatus = 3 // 已完成
	TaskStatusFailed    TaskStatus = 4 // 执行失败
	TaskStatusCancelled TaskStatus = 5 // 已取消
)

// String 返回任务状态的字符串表示
func (s TaskStatus) String() string {
	switch s {
	case TaskStatusPending:
		return "PENDING"
	case TaskStatusRunning:
		return "RUNNING"
	case TaskStatusCompleted:
		return "COMPLETED"
	case TaskStatusFailed:
		return "FAILED"
	case TaskStatusCancelled:
		return "CANCELLED"
	default:
		return "UNKNOWN"
	}
}

// ScheduleType 调度类型
type ScheduleType int

const (
	ScheduleTypeOnce     ScheduleType = iota // 一次性执行
	ScheduleTypeInterval                     // 固定间隔
	ScheduleTypeCron                         // Cron表达式
	ScheduleTypeDelay                        // 延迟执行
)

// TaskPriority 任务优先级
type TaskPriority int

const (
	TaskPriorityLow    TaskPriority = iota // 低优先级
	TaskPriorityNormal                     // 普通优先级
	TaskPriorityHigh                       // 高优先级
)

// TaskExecutor 任务执行器接口
// 所有定时任务都需要实现这个接口，用于定义具体的任务执行逻辑
type TaskExecutor interface {
	// Execute 执行任务的核心方法
	// 参数:
	//   ctx: 上下文，用于控制超时和取消
	//   params: 任务参数，可以是任意类型的数据
	// 返回:
	//   error: 执行错误，nil表示成功
	Execute(ctx context.Context, params interface{}) error
	
	// GetName 获取任务执行器的名称
	// 返回:
	//   string: 执行器的唯一标识名称，用于日志和监控
	GetName() string
}

// TaskConfig 任务配置
type TaskConfig struct {
	// 基本信息
	ID          string       `json:"id"`          // 任务唯一标识
	Name        string       `json:"name"`        // 任务名称
	Description string       `json:"description"` // 任务描述
	Priority    TaskPriority `json:"priority"`    // 任务优先级
	
	// 调度配置
	ScheduleType ScheduleType  `json:"scheduleType"` // 调度类型
	CronExpr     string        `json:"cronExpr"`     // Cron表达式
	Interval     time.Duration `json:"interval"`     // 执行间隔
	Delay        time.Duration `json:"delay"`        // 延迟时间
	StartTime    *time.Time    `json:"startTime"`    // 开始时间
	EndTime      *time.Time    `json:"endTime"`      // 结束时间
	
	// 执行配置
	MaxRetries    int           `json:"maxRetries"`    // 最大重试次数
	RetryInterval time.Duration `json:"retryInterval"` // 重试间隔
	Timeout       time.Duration `json:"timeout"`       // 执行超时时间
	
	// 任务参数
	Params interface{} `json:"params"` // 任务参数
	
	// 其他配置
	Enabled bool `json:"enabled"` // 是否启用
}

// TaskInfo 任务运行时信息
type TaskInfo struct {
	Config      *TaskConfig   `json:"config"`      // 任务配置
	Status      TaskStatus    `json:"status"`      // 当前状态
	NextRunTime *time.Time    `json:"nextRunTime"` // 下次执行时间
	LastRunTime *time.Time    `json:"lastRunTime"` // 上次执行时间
	LastResult  *TaskResult   `json:"lastResult"`  // 上次执行结果
	RunCount    int64         `json:"runCount"`    // 执行次数
	FailureCount int64        `json:"failureCount"` // 失败次数
	CreatedAt   time.Time     `json:"createdAt"`   // 创建时间
	UpdatedAt   time.Time     `json:"updatedAt"`   // 更新时间
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID     string        `json:"taskId"`     // 任务ID
	StartTime  time.Time     `json:"startTime"`  // 开始时间
	EndTime    time.Time     `json:"endTime"`    // 结束时间
	Duration   time.Duration `json:"duration"`   // 执行耗时
	Status     TaskStatus    `json:"status"`     // 执行状态
	Error      string        `json:"error"`      // 错误信息
	RetryCount int           `json:"retryCount"` // 重试次数
}

// TaskScheduler 任务调度器接口
// 定义了完整的任务调度管理功能，包括任务管理、控制和监控
type TaskScheduler interface {
	// ===== 任务管理方法 =====
	
	// AddTask 添加新的定时任务到调度器
	// 参数:
	//   config: 任务配置信息，包含调度规则和基本信息
	//   executor: 任务执行器，定义具体的执行逻辑
	// 返回:
	//   error: 添加失败时返回错误信息
	AddTask(config *TaskConfig, executor TaskExecutor) error
	
	// RemoveTask 从调度器中移除指定任务
	// 参数:
	//   taskID: 要移除的任务ID
	// 返回:
	//   error: 移除失败时返回错误信息
	RemoveTask(taskID string) error
	
	// GetTask 获取指定任务的详细信息
	// 参数:
	//   taskID: 任务ID
	// 返回:
	//   *TaskInfo: 任务信息，包含配置、状态、执行历史等
	//   error: 获取失败时返回错误信息
	GetTask(taskID string) (*TaskInfo, error)
	
	// ListTasks 获取所有任务的列表
	// 返回:
	//   []*TaskInfo: 所有任务信息的切片
	//   error: 获取失败时返回错误信息
	ListTasks() ([]*TaskInfo, error)
	
	// ===== 任务控制方法 =====
	
	// StartTask 启动指定的任务调度
	// 参数:
	//   taskID: 要启动的任务ID
	// 返回:
	//   error: 启动失败时返回错误信息
	StartTask(taskID string) error
	
	// StopTask 停止指定的任务调度
	// 参数:
	//   taskID: 要停止的任务ID
	// 返回:
	//   error: 停止失败时返回错误信息
	StopTask(taskID string) error
	
	// TriggerTask 手动触发任务执行（不影响正常调度）
	// 参数:
	//   taskID: 要触发的任务ID
	//   params: 执行参数，会覆盖任务配置中的参数
	// 返回:
	//   error: 触发失败时返回错误信息
	TriggerTask(taskID string, params interface{}) error
	
	// ===== 调度器控制方法 =====
	
	// Start 启动整个调度器，开始任务调度
	// 返回:
	//   error: 启动失败时返回错误信息
	Start() error
	
	// Stop 停止整个调度器，停止所有任务调度
	// 返回:
	//   error: 停止失败时返回错误信息
	Stop() error
	
	// IsRunning 检查调度器是否正在运行
	// 返回:
	//   bool: true表示正在运行，false表示已停止
	IsRunning() bool
	
	// ===== 监控和查询方法 =====
	
	// GetTaskHistory 获取指定任务的执行历史记录
	// 参数:
	//   taskID: 任务ID
	//   limit: 返回记录的最大数量，0表示不限制
	// 返回:
	//   []*TaskResult: 执行结果列表，按时间倒序排列
	//   error: 获取失败时返回错误信息
	GetTaskHistory(taskID string, limit int) ([]*TaskResult, error)
	
	// GetRunningTasks 获取当前正在运行的任务列表
	// 返回:
	//   []*TaskInfo: 正在运行的任务信息列表
	//   error: 获取失败时返回错误信息
	GetRunningTasks() ([]*TaskInfo, error)
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	ID           string        `json:"id"`           // 调度器唯一标识
	Name         string        `json:"name"`         // 调度器名称
	MaxWorkers   int           `json:"maxWorkers"`   // 最大工作线程数
	QueueSize    int           `json:"queueSize"`    // 任务队列大小
	DefaultTimeout time.Duration `json:"defaultTimeout"` // 默认超时时间
	DefaultRetries int           `json:"defaultRetries"` // 默认重试次数
}

// DefaultSchedulerConfig 返回默认调度器配置
// 提供合理的默认值，适用于大多数场景
// 返回:
//   *SchedulerConfig: 包含默认配置的调度器配置对象
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		ID:             "default",          // 默认调度器ID
		Name:           "DefaultScheduler", // 调度器名称
		MaxWorkers:     5,                  // 最大工作线程数，控制并发执行的任务数量
		QueueSize:      100,                // 任务队列大小，超过此数量的任务会被阻塞
		DefaultTimeout: time.Minute * 30,   // 默认任务超时时间，30分钟
		DefaultRetries: 3,                  // 默认重试次数，失败后最多重试3次
	}
}

// TaskStorage 任务存储接口
// 定义了任务数据的持久化存储操作，支持任务配置、状态和执行结果的管理
type TaskStorage interface {
	// ===== 任务配置存储方法 =====
	
	// SaveTaskConfig 保存任务配置到存储中
	// 参数:
	//   config: 要保存的任务配置对象
	// 返回:
	//   error: 保存失败时返回错误信息
	SaveTaskConfig(config *TaskConfig) error
	
	// LoadTaskConfig 从存储中加载指定任务的配置
	// 参数:
	//   taskID: 任务ID
	// 返回:
	//   *TaskConfig: 任务配置对象
	//   error: 加载失败时返回错误信息
	LoadTaskConfig(taskID string) (*TaskConfig, error)
	
	// DeleteTaskConfig 从存储中删除指定任务的配置
	// 参数:
	//   taskID: 要删除的任务ID
	// 返回:
	//   error: 删除失败时返回错误信息
	DeleteTaskConfig(taskID string) error
	
	// ListTaskConfigs 获取所有任务配置的列表
	// 返回:
	//   []*TaskConfig: 所有任务配置的切片
	//   error: 获取失败时返回错误信息
	ListTaskConfigs() ([]*TaskConfig, error)
	
	// ===== 任务状态存储方法 =====
	
	// SaveTaskInfo 保存任务运行时信息到存储中
	// 参数:
	//   info: 要保存的任务信息对象
	// 返回:
	//   error: 保存失败时返回错误信息
	SaveTaskInfo(info *TaskInfo) error
	
	// LoadTaskInfo 从存储中加载指定任务的运行时信息
	// 参数:
	//   taskID: 任务ID
	// 返回:
	//   *TaskInfo: 任务信息对象
	//   error: 加载失败时返回错误信息
	LoadTaskInfo(taskID string) (*TaskInfo, error)
	
	// ===== 任务结果存储方法 =====
	
	// SaveTaskResult 保存任务执行结果到存储中
	// 参数:
	//   result: 要保存的任务执行结果
	// 返回:
	//   error: 保存失败时返回错误信息
	SaveTaskResult(result *TaskResult) error
	
	// LoadTaskResults 从存储中加载指定任务的执行历史记录
	// 参数:
	//   taskID: 任务ID
	//   limit: 返回记录的最大数量，0表示不限制
	// 返回:
	//   []*TaskResult: 执行结果列表，按时间倒序排列
	//   error: 加载失败时返回错误信息
	LoadTaskResults(taskID string, limit int) ([]*TaskResult, error)
} 