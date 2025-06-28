package timer

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// ValidateTaskConfig 验证任务配置的有效性
// 检查任务配置中的必填字段和业务逻辑约束
// 参数:
//   config: 待验证的任务配置
// 返回:
//   error: 验证失败时返回具体的错误信息，成功时返回nil
func ValidateTaskConfig(config *TaskConfig) error {
	// 检查配置对象是否为空
	if config == nil {
		return errors.New("task config cannot be nil")
	}
	
	// 验证任务ID（必填，用于唯一标识任务）
	if config.ID == "" {
		return errors.New("task ID cannot be empty")
	}
	
	// 验证任务名称（必填，用于显示和日志）
	if config.Name == "" {
		return errors.New("task name cannot be empty")
	}
	
	// 根据调度类型验证相关配置
	switch config.ScheduleType {
	case ScheduleTypeCron:
		// Cron调度需要有效的Cron表达式
		if config.CronExpr == "" {
			return errors.New("cron expression is required for cron schedule type")
		}
	case ScheduleTypeInterval:
		// 间隔调度需要正数间隔时间
		if config.Interval <= 0 {
			return errors.New("interval must be greater than 0 for interval schedule type")
		}
	case ScheduleTypeDelay:
		// 延迟调度需要正数延迟时间
		if config.Delay <= 0 {
			return errors.New("delay must be greater than 0 for delay schedule type")
		}
	case ScheduleTypeOnce:
		// 一次性任务不需要额外验证
	default:
		return fmt.Errorf("unsupported schedule type: %v", config.ScheduleType)
	}
	
	// 验证重试次数不能为负数
	if config.MaxRetries < 0 {
		return errors.New("max retries cannot be negative")
	}
	
	return nil
}

// CopyTaskConfig 深拷贝任务配置
// 使用JSON序列化/反序列化实现深拷贝，确保配置对象的独立性
// 参数:
//   config: 源任务配置对象
// 返回:
//   *TaskConfig: 拷贝后的新配置对象
//   error: 拷贝过程中的错误
func CopyTaskConfig(config *TaskConfig) (*TaskConfig, error) {
	// 处理空配置的情况
	if config == nil {
		return nil, nil
	}
	
	// 将配置对象序列化为JSON
	data, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	
	// 将JSON反序列化为新的配置对象
	var copy TaskConfig
	if err := json.Unmarshal(data, &copy); err != nil {
		return nil, err
	}
	
	return &copy, nil
}

// NewTaskConfig 创建新的任务配置对象
// 基于给定的基本信息创建初始的任务配置，包含初始状态信息
// 参数:
//   id: 任务ID
//   name: 任务名称
//   scheduleType: 调度类型
// 返回:
//   *TaskConfig: 初始化的任务配置对象
func NewTaskConfig(id, name string, scheduleType ScheduleType) *TaskConfig {
	now := time.Now()
	return &TaskConfig{
		// 基本信息
		ID:           id,
		Name:         name,
		Priority:     TaskPriorityNormal,
		ScheduleType: scheduleType,
		Enabled:      true,
		CreatedAt:    now,
		
		// 运行时状态信息
		Status:       TaskStatusPending,  // 初始状态为待执行
		RunCount:     0,                  // 执行次数初始为0
		FailureCount: 0,                  // 失败次数初始为0
		UpdatedAt:    now,                // 更新时间
	}
}

// NewTaskResult 创建新的任务执行结果对象
// 用于记录任务的执行过程和结果，初始状态为运行中
// 参数:
//   taskID: 任务ID，用于关联到具体的任务
// 返回:
//   *TaskResult: 初始化的任务结果对象
func NewTaskResult(taskID string) *TaskResult {
	return &TaskResult{
		TaskID:     taskID,              // 关联的任务ID
		StartTime:  time.Now(),          // 任务开始执行时间
		Status:     TaskStatusRunning,   // 初始状态为运行中
		RetryCount: 0,                   // 重试次数初始为0
	}
}

// Complete 标记任务执行结果为成功完成
// 设置任务结束时间、计算执行耗时，并更新状态为已完成
func (r *TaskResult) Complete() {
	r.EndTime = time.Now()                          // 记录任务结束时间
	r.Duration = r.EndTime.Sub(r.StartTime)         // 计算任务执行耗时
	r.Status = TaskStatusCompleted                  // 更新状态为已完成
}

// Fail 标记任务执行结果为失败
// 设置任务结束时间、计算执行耗时、更新状态为失败，并记录错误信息
// 参数:
//   err: 导致任务失败的错误对象，可以为nil
func (r *TaskResult) Fail(err error) {
	r.EndTime = time.Now()                          // 记录任务结束时间
	r.Duration = r.EndTime.Sub(r.StartTime)         // 计算任务执行耗时
	r.Status = TaskStatusFailed                     // 更新状态为失败
	if err != nil {
		r.Error = err.Error()                       // 记录错误信息
	}
}

// IsExpired 检查任务是否已过期
// 根据任务配置中的结束时间判断任务是否已经过期
// 返回:
//   bool: true表示任务已过期，false表示任务未过期或无结束时间限制
func (c *TaskConfig) IsExpired() bool {
	// 如果没有设置结束时间，则认为任务永不过期
	if c.EndTime == nil {
		return false
	}
	// 检查当前时间是否已超过结束时间
	return time.Now().After(*c.EndTime)
}

// ShouldStart 检查任务是否应该开始执行
// 综合考虑任务的启用状态、过期状态和开始时间来判断是否应该启动任务
// 返回:
//   bool: true表示任务应该启动，false表示任务不应该启动
func (c *TaskConfig) ShouldStart() bool {
	// 检查任务是否已启用
	if !c.Enabled {
		return false
	}
	
	// 检查任务是否已过期
	if c.IsExpired() {
		return false
	}
	
	// 如果设置了开始时间，检查是否已到达开始时间
	if c.StartTime != nil && time.Now().Before(*c.StartTime) {
		return false
	}
	
	return true
} 