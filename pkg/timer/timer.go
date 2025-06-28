// Package timer 提供了一个功能完整的定时任务调度组件
// 支持多种调度策略：Cron表达式、固定间隔、延迟执行、一次性执行
package timer

import (
	"context"
	"sync"
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

// ExecuteResult 任务执行结果详情
type ExecuteResult struct {
	Success bool        `json:"success"`     // 是否执行成功
	Data    interface{} `json:"data"`        // 执行返回的数据
	Message string      `json:"message"`     // 执行的消息（成功或失败的详细信息）
	Extra   interface{} `json:"extra"`       // 额外的扩展信息
}

// TaskExecutor 任务执行器接口
// 所有定时任务都需要实现这个接口，用于定义具体的任务执行逻辑
type TaskExecutor interface {
	// Execute 执行任务的核心方法
	// 参数:
	//   ctx: 上下文，用于控制超时和取消
	//   params: 任务参数，可以是任意类型的数据
	// 返回:
	//   *ExecuteResult: 执行结果，包含执行状态、数据和消息
	//   error: 执行过程中的错误，nil表示执行过程成功（注意：这不等同于业务是否成功）
	Execute(ctx context.Context, params interface{}) (*ExecuteResult, error)
	
	// GetName 获取任务执行器的名称
	// 返回:
	//   string: 执行器的唯一标识名称，用于日志和监控
	GetName() string
	
	// Close 关闭执行器并释放资源
	// 用于清理执行器占用的资源，如网络连接、文件句柄等
	// 返回:
	//   error: 关闭过程中的错误，nil表示成功关闭
	Close() error
}

// TaskConfig 任务配置（包含运行时状态信息，但不维护timer）
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
	Enabled   bool      `json:"enabled"`   // 是否启用
	CreatedAt time.Time `json:"createdAt"` // 创建时间
	
	// 运行时状态信息
	Status       TaskStatus    `json:"status"`       // 当前状态
	NextRunTime  *time.Time    `json:"nextRunTime"`  // 下次执行时间
	LastRunTime  *time.Time    `json:"lastRunTime"`  // 上次执行时间
	LastResult   *TaskResult   `json:"lastResult"`   // 上次执行结果
	RunCount     int64         `json:"runCount"`     // 执行次数
	FailureCount int64         `json:"failureCount"` // 失败次数
	UpdatedAt    time.Time     `json:"updatedAt"`    // 更新时间
	
	// 并发控制
	mu sync.RWMutex `json:"-"` // 读写锁，用于并发安全
}

// UpdateStatus 线程安全地更新任务状态
func (tc *TaskConfig) UpdateStatus(status TaskStatus) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Status = status
	tc.UpdatedAt = time.Now()
}

// UpdateRunInfo 线程安全地更新运行信息
func (tc *TaskConfig) UpdateRunInfo(result *TaskResult) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.LastResult = result
	tc.LastRunTime = &result.EndTime
	tc.RunCount++
	if result.Status == TaskStatusFailed {
		tc.FailureCount++
	}
	tc.UpdatedAt = time.Now()
	
	// 关键修复：根据执行结果重置任务状态，确保间隔任务能够继续调度
	// 对于重复执行的任务（如间隔任务、cron任务），执行完成后应该重置为待执行状态
	switch tc.ScheduleType {
	case ScheduleTypeInterval, ScheduleTypeCron:
		// 间隔任务和cron任务执行完成后，重置为待执行状态，以便继续调度
		tc.Status = TaskStatusPending
	case ScheduleTypeOnce, ScheduleTypeDelay:
		// 一次性任务和延迟任务根据执行结果设置最终状态  
		if result.Status == TaskStatusFailed {
			tc.Status = TaskStatusFailed
		} else {
			tc.Status = TaskStatusCompleted
		}
	default:
		// 其他类型任务根据执行结果设置状态
		if result.Status == TaskStatusFailed {
			tc.Status = TaskStatusFailed
		} else {
			tc.Status = TaskStatusCompleted
		}
	}
}

// GetStatus 线程安全地获取任务状态
func (tc *TaskConfig) GetStatus() TaskStatus {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.Status
}

// GetRunCount 线程安全地获取运行次数
func (tc *TaskConfig) GetRunCount() int64 {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.RunCount
}

// GetLastRunTime 线程安全地获取上次执行时间
func (tc *TaskConfig) GetLastRunTime() *time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.LastRunTime
}

// GetNextRunTime 线程安全地获取下次执行时间
func (tc *TaskConfig) GetNextRunTime() *time.Time {
	tc.mu.RLock()
	defer tc.mu.RUnlock()
	return tc.NextRunTime
}

// SetNextRunTime 线程安全地设置下次执行时间
func (tc *TaskConfig) SetNextRunTime(nextTime *time.Time) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.NextRunTime = nextTime
	tc.UpdatedAt = time.Now()
}

// TaskResult 任务执行结果（用于外部结果处理）
type TaskResult struct {
	TaskID     string        `json:"taskId"`     // 任务ID
	StartTime  time.Time     `json:"startTime"`  // 开始时间
	EndTime    time.Time     `json:"endTime"`    // 结束时间
	Duration   time.Duration `json:"duration"`   // 执行耗时
	Status     TaskStatus    `json:"status"`     // 执行状态
	Error      string        `json:"error"`      // 错误信息（执行过程错误）
	RetryCount int           `json:"retryCount"` // 重试次数
	Result     *ExecuteResult `json:"result"`    // 执行结果详情
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
	//   *TaskConfig: 任务配置信息
	//   error: 获取失败时返回错误信息
	GetTask(taskID string) (*TaskConfig, error)
	
	// ListTasks 获取所有任务的列表
	// 返回:
	//   []*TaskConfig: 所有任务配置信息的切片
	//   error: 获取失败时返回错误信息
	ListTasks() ([]*TaskConfig, error)
	
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
	
	// GetRunningTasks 获取当前正在运行的任务列表
	// 返回:
	//   []*TaskConfig: 正在运行的任务配置列表
	//   error: 获取失败时返回错误信息
	GetRunningTasks() ([]*TaskConfig, error)
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	ID               string                    `json:"id"`               // 调度器唯一标识
	Name             string                    `json:"name"`             // 调度器名称
	TenantId         string                    `json:"tenantId"`         // 租户ID标识
	MaxWorkers       int                       `json:"maxWorkers"`       // 最大工作线程数
	QueueSize        int                       `json:"queueSize"`        // 任务队列大小
	DefaultTimeout   time.Duration             `json:"defaultTimeout"`   // 默认超时时间
	DefaultRetries   int                       `json:"defaultRetries"`   // 默认重试次数
	ScheduleInterval time.Duration             `json:"scheduleInterval"` // 调度检查间隔
	
	// 任务配置映射
	Tasks map[string]*TaskConfig `json:"tasks"` // 任务ID到任务配置的映射
	
	// 并发控制
	mu sync.RWMutex `json:"-"` // 读写锁，用于并发安全
}

// DefaultSchedulerConfig 返回默认的调度器配置
func DefaultSchedulerConfig() *SchedulerConfig {
	return &SchedulerConfig{
		ID:               "default-scheduler",
		Name:             "DefaultScheduler",
		TenantId:         "default",
		MaxWorkers:       5,
		QueueSize:        100,
		DefaultTimeout:   time.Minute * 30,
		DefaultRetries:   3,
		ScheduleInterval: time.Second,
		Tasks:            make(map[string]*TaskConfig),
	}
}

// AddTask 线程安全地添加任务配置
func (sc *SchedulerConfig) AddTask(config *TaskConfig) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.Tasks[config.ID] = config
}

// RemoveTask 线程安全地移除任务配置
func (sc *SchedulerConfig) RemoveTask(taskID string) {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	delete(sc.Tasks, taskID)
}

// GetTask 线程安全地获取任务配置
func (sc *SchedulerConfig) GetTask(taskID string) (*TaskConfig, bool) {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	config, exists := sc.Tasks[taskID]
	return config, exists
}

// ListTasks 线程安全地获取所有任务配置
func (sc *SchedulerConfig) ListTasks() []*TaskConfig {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	
	tasks := make([]*TaskConfig, 0, len(sc.Tasks))
	for _, config := range sc.Tasks {
		tasks = append(tasks, config)
	}
	return tasks
} 