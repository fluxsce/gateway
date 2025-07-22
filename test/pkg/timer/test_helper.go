package timer

import (
	"context"
	"time"

	"gateway/pkg/timer"
)

// CreateTestTaskConfig 创建用于测试的任务配置
// 参数:
//   - id: 任务ID
//   - name: 任务名称
//   - scheduleType: 调度类型
//
// 返回:
//   - *timer.TaskConfig: 任务配置对象
func CreateTestTaskConfig(id string, name string, scheduleType timer.ScheduleType) *timer.TaskConfig {
	config := &timer.TaskConfig{
		ID:            id,
		Name:          name,
		Description:   "测试任务",
		Priority:      timer.TaskPriorityNormal,
		ScheduleType:  scheduleType,
		Enabled:       true,
		MaxRetries:    2,
		RetryInterval: time.Second,
		Timeout:       time.Second * 30,
		CreatedAt:     time.Now(),
	}

	// 根据调度类型设置特定配置
	switch scheduleType {
	case timer.ScheduleTypeOnce:
		config.StartTime = &time.Time{} // 空时间表示立即执行
	case timer.ScheduleTypeInterval:
		config.Interval = time.Second * 2
	case timer.ScheduleTypeCron:
		config.CronExpr = "*/5 * * * *" // 每5分钟执行一次
	case timer.ScheduleTypeDelay:
		config.Delay = time.Second * 5
	}

	return config
}

// TestTaskExecutor 测试用的任务执行器
type TestTaskExecutor struct {
	name        string
	executeFunc func(ctx context.Context, params interface{}) error
}

// NewTestTaskExecutor 创建测试任务执行器
// 参数:
//   - name: 执行器名称
//   - executeFunc: 执行函数，如果为nil则使用默认的空实现
//
// 返回:
//   - *TestTaskExecutor: 测试执行器实例
func NewTestTaskExecutor(name string, executeFunc func(ctx context.Context, params interface{}) error) *TestTaskExecutor {
	if executeFunc == nil {
		executeFunc = func(ctx context.Context, params interface{}) error {
			// 默认实现：等待100ms模拟任务执行
			time.Sleep(time.Millisecond * 100)
			return nil
		}
	}

	return &TestTaskExecutor{
		name:        name,
		executeFunc: executeFunc,
	}
}

// Execute 实现TaskExecutor接口
func (e *TestTaskExecutor) Execute(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	err := e.executeFunc(ctx, params)
	return &timer.ExecuteResult{
		Success: err == nil,
		Message: "Test execution completed",
	}, err
}

// GetName 实现TaskExecutor接口
func (e *TestTaskExecutor) GetName() string {
	return e.name
}

// Close 实现TaskExecutor接口
func (e *TestTaskExecutor) Close() error {
	return nil
}

// WaitForCondition 等待条件满足或超时
// 参数:
//   - condition: 条件函数，返回true表示条件满足
//   - timeout: 最大等待时间
//   - interval: 检查间隔
//
// 返回:
//   - bool: true表示条件满足，false表示超时
func WaitForCondition(condition func() bool, timeout time.Duration, interval time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if condition() {
			return true
		}
		time.Sleep(interval)
	}
	return false
}
