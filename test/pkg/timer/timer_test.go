package timer

import (
	"context"
	"testing"
	"time"

	"gohub/pkg/timer"
)

// TestTaskStatus 测试任务状态枚举
// 验证TaskStatus枚举值的正确性和String方法的输出
func TestTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   timer.TaskStatus
		expected string
	}{
		{
			name:     "待执行状态",
			status:   timer.TaskStatusPending,
			expected: "PENDING",
		},
		{
			name:     "运行中状态",
			status:   timer.TaskStatusRunning,
			expected: "RUNNING",
		},
		{
			name:     "已完成状态",
			status:   timer.TaskStatusCompleted,
			expected: "COMPLETED",
		},
		{
			name:     "执行失败状态",
			status:   timer.TaskStatusFailed,
			expected: "FAILED",
		},
		{
			name:     "已取消状态",
			status:   timer.TaskStatusCancelled,
			expected: "CANCELLED",
		},
		{
			name:     "未知状态",
			status:   timer.TaskStatus(999),
			expected: "UNKNOWN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.status.String()
			if result != tt.expected {
				t.Errorf("TaskStatus.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestScheduleType 测试调度类型枚举
// 验证ScheduleType枚举值的定义和使用
func TestScheduleType(t *testing.T) {
	tests := []struct {
		name         string
		scheduleType timer.ScheduleType
		expected     int
	}{
		{
			name:         "一次性执行",
			scheduleType: timer.ScheduleTypeOnce,
			expected:     0,
		},
		{
			name:         "固定间隔执行",
			scheduleType: timer.ScheduleTypeInterval,
			expected:     1,
		},
		{
			name:         "Cron表达式执行",
			scheduleType: timer.ScheduleTypeCron,
			expected:     2,
		},
		{
			name:         "延迟执行",
			scheduleType: timer.ScheduleTypeDelay,
			expected:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.scheduleType) != tt.expected {
				t.Errorf("ScheduleType value = %v, want %v", int(tt.scheduleType), tt.expected)
			}
		})
	}
}

// TestTaskPriority 测试任务优先级枚举
// 验证TaskPriority枚举值的定义和使用
func TestTaskPriority(t *testing.T) {
	tests := []struct {
		name     string
		priority timer.TaskPriority
		expected int
	}{
		{
			name:     "低优先级",
			priority: timer.TaskPriorityLow,
			expected: 0,
		},
		{
			name:     "普通优先级",
			priority: timer.TaskPriorityNormal,
			expected: 1,
		},
		{
			name:     "高优先级",
			priority: timer.TaskPriorityHigh,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.priority) != tt.expected {
				t.Errorf("TaskPriority value = %v, want %v", int(tt.priority), tt.expected)
			}
		})
	}
}

// TestDefaultSchedulerConfig 测试默认调度器配置
// 验证DefaultSchedulerConfig函数返回的配置是否符合预期
func TestDefaultSchedulerConfig(t *testing.T) {
	config := timer.DefaultSchedulerConfig()

	// 验证配置不为空
	if config == nil {
		t.Fatal("DefaultSchedulerConfig() returned nil")
	}

	// 验证默认配置值
	expectedValues := map[string]interface{}{
		"Name":           "DefaultScheduler",
		"MaxWorkers":     5,
		"QueueSize":      100,
		"DefaultTimeout": time.Minute * 30,
		"DefaultRetries": 3,
	}

	// 检查调度器名称
	if config.Name != expectedValues["Name"] {
		t.Errorf("Default Name = %v, want %v", config.Name, expectedValues["Name"])
	}

	// 检查最大工作线程数
	if config.MaxWorkers != expectedValues["MaxWorkers"] {
		t.Errorf("Default MaxWorkers = %v, want %v", config.MaxWorkers, expectedValues["MaxWorkers"])
	}

	// 检查队列大小
	if config.QueueSize != expectedValues["QueueSize"] {
		t.Errorf("Default QueueSize = %v, want %v", config.QueueSize, expectedValues["QueueSize"])
	}

	// 检查默认超时时间
	if config.DefaultTimeout != expectedValues["DefaultTimeout"] {
		t.Errorf("Default DefaultTimeout = %v, want %v", config.DefaultTimeout, expectedValues["DefaultTimeout"])
	}

	// 检查默认重试次数
	if config.DefaultRetries != expectedValues["DefaultRetries"] {
		t.Errorf("Default DefaultRetries = %v, want %v", config.DefaultRetries, expectedValues["DefaultRetries"])
	}
}

// TestTaskConfig 测试任务配置结构
// 验证TaskConfig结构体的字段设置和访问
func TestTaskConfig(t *testing.T) {
	// 创建测试时间
	startTime := time.Now().Add(time.Hour)
	endTime := time.Now().Add(time.Hour * 2)

	// 创建任务配置
	config := &timer.TaskConfig{
		ID:            "test-task-001",
		Name:          "测试任务",
		Description:   "这是一个测试任务",
		Priority:      timer.TaskPriorityHigh,
		ScheduleType:  timer.ScheduleTypeCron,
		CronExpr:      "0 0 * * *",
		Interval:      time.Hour,
		Delay:         time.Minute * 5,
		StartTime:     &startTime,
		EndTime:       &endTime,
		MaxRetries:    5,
		RetryInterval: time.Second * 30,
		Timeout:       time.Minute * 10,
		Params:        map[string]interface{}{"key": "value"},
		Enabled:       true,
	}

	// 验证基本字段
	if config.ID != "test-task-001" {
		t.Errorf("TaskConfig.ID = %v, want %v", config.ID, "test-task-001")
	}

	if config.Name != "测试任务" {
		t.Errorf("TaskConfig.Name = %v, want %v", config.Name, "测试任务")
	}

	if config.Priority != timer.TaskPriorityHigh {
		t.Errorf("TaskConfig.Priority = %v, want %v", config.Priority, timer.TaskPriorityHigh)
	}

	if config.ScheduleType != timer.ScheduleTypeCron {
		t.Errorf("TaskConfig.ScheduleType = %v, want %v", config.ScheduleType, timer.ScheduleTypeCron)
	}

	// 验证时间字段
	if config.StartTime == nil || !config.StartTime.Equal(startTime) {
		t.Errorf("TaskConfig.StartTime = %v, want %v", config.StartTime, &startTime)
	}

	if config.EndTime == nil || !config.EndTime.Equal(endTime) {
		t.Errorf("TaskConfig.EndTime = %v, want %v", config.EndTime, &endTime)
	}

	// 验证启用状态
	if !config.Enabled {
		t.Errorf("TaskConfig.Enabled = %v, want %v", config.Enabled, true)
	}
}

// TestTaskConfig 测试任务配置的状态管理
// 验证TaskConfig的状态管理方法
func TestTaskConfigStateMethods(t *testing.T) {
	// 创建任务配置
	config := CreateTestTaskConfig("test-state-001", "状态测试任务", timer.ScheduleTypeInterval)

	// 测试状态更新
	config.UpdateStatus(timer.TaskStatusRunning)
	if config.GetStatus() != timer.TaskStatusRunning {
		t.Errorf("UpdateStatus failed, got %v, want %v", config.GetStatus(), timer.TaskStatusRunning)
	}

	// 测试运行信息更新
	result := &timer.TaskResult{
		TaskID:   "test-state-001",
		Status:   timer.TaskStatusCompleted,
		Duration: time.Second * 10,
	}
	
	config.UpdateRunInfo(result)
	if config.GetRunCount() == 0 {
		t.Error("UpdateRunInfo should increment run count")
	}

	// 测试下次运行时间设置
	nextRun := time.Now().Add(time.Hour)
	config.SetNextRunTime(&nextRun)
	
	retrievedTime := config.GetNextRunTime()
	if retrievedTime == nil || !retrievedTime.Equal(nextRun) {
		t.Errorf("SetNextRunTime/GetNextRunTime failed, got %v, want %v", retrievedTime, &nextRun)
	}
}

// TestTaskResult 测试任务执行结果结构
// 验证TaskResult结构体的字段设置和访问
func TestTaskResult(t *testing.T) {
	// 创建任务执行结果
	startTime := time.Now()
	endTime := startTime.Add(time.Second * 30)
	duration := endTime.Sub(startTime)

	result := &timer.TaskResult{
		TaskID:     "test-result-001",
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   duration,
		Status:     timer.TaskStatusCompleted,
		Error:      "",
		RetryCount: 1,
	}

	// 验证字段值
	if result.TaskID != "test-result-001" {
		t.Errorf("TaskResult.TaskID = %v, want %v", result.TaskID, "test-result-001")
	}

	if result.Status != timer.TaskStatusCompleted {
		t.Errorf("TaskResult.Status = %v, want %v", result.Status, timer.TaskStatusCompleted)
	}

	if result.RetryCount != 1 {
		t.Errorf("TaskResult.RetryCount = %v, want %v", result.RetryCount, 1)
	}

	// 验证时间和持续时间
	if !result.StartTime.Equal(startTime) {
		t.Errorf("TaskResult.StartTime = %v, want %v", result.StartTime, startTime)
	}

	if !result.EndTime.Equal(endTime) {
		t.Errorf("TaskResult.EndTime = %v, want %v", result.EndTime, endTime)
	}

	if result.Duration != duration {
		t.Errorf("TaskResult.Duration = %v, want %v", result.Duration, duration)
	}

	// 验证错误信息为空（成功执行）
	if result.Error != "" {
		t.Errorf("TaskResult.Error = %v, want empty string", result.Error)
	}
}

// TestSchedulerConfig 测试调度器配置结构
// 验证SchedulerConfig结构体的字段设置和访问
func TestSchedulerConfig(t *testing.T) {
	config := &timer.SchedulerConfig{
		Name:           "TestScheduler",
		MaxWorkers:     10,
		QueueSize:      200,
		DefaultTimeout: time.Minute * 15,
		DefaultRetries: 5,
	}

	// 验证字段值
	if config.Name != "TestScheduler" {
		t.Errorf("SchedulerConfig.Name = %v, want %v", config.Name, "TestScheduler")
	}

	if config.MaxWorkers != 10 {
		t.Errorf("SchedulerConfig.MaxWorkers = %v, want %v", config.MaxWorkers, 10)
	}

	if config.QueueSize != 200 {
		t.Errorf("SchedulerConfig.QueueSize = %v, want %v", config.QueueSize, 200)
	}

	if config.DefaultTimeout != time.Minute*15 {
		t.Errorf("SchedulerConfig.DefaultTimeout = %v, want %v", config.DefaultTimeout, time.Minute*15)
	}

	if config.DefaultRetries != 5 {
		t.Errorf("SchedulerConfig.DefaultRetries = %v, want %v", config.DefaultRetries, 5)
	}
}

// MockTaskExecutor 测试用的任务执行器实现
// 用于测试TaskExecutor接口的基本功能
type MockTaskExecutor struct {
	name        string
	executeFunc func(ctx context.Context, params interface{}) error
}

// NewMockTaskExecutor 创建模拟任务执行器
func NewMockTaskExecutor(name string, executeFunc func(ctx context.Context, params interface{}) error) *MockTaskExecutor {
	return &MockTaskExecutor{
		name:        name,
		executeFunc: executeFunc,
	}
}

// Execute 实现TaskExecutor接口
func (m *MockTaskExecutor) Execute(ctx context.Context, params interface{}) (*timer.ExecuteResult, error) {
	var err error
	if m.executeFunc != nil {
		err = m.executeFunc(ctx, params)
	}
	return &timer.ExecuteResult{
		Success: err == nil,
		Message: "Mock execution completed",
	}, err
}

// GetName 实现TaskExecutor接口
func (m *MockTaskExecutor) GetName() string {
	return m.name
}

// Close 实现TaskExecutor接口
func (m *MockTaskExecutor) Close() error {
	return nil
}

// TestTaskExecutorInterface 测试TaskExecutor接口
// 验证TaskExecutor接口的实现和使用
func TestTaskExecutorInterface(t *testing.T) {
	// 测试成功执行的情况
	t.Run("成功执行", func(t *testing.T) {
		executor := NewMockTaskExecutor("test-executor", func(ctx context.Context, params interface{}) error {
			return nil
		})

		// 验证名称
		if executor.GetName() != "test-executor" {
			t.Errorf("GetName() = %v, want %v", executor.GetName(), "test-executor")
		}

		// 验证执行
		ctx := context.Background()
		result, err := executor.Execute(ctx, "test-params")
		if err != nil {
			t.Errorf("Execute() error = %v, want nil", err)
		}
		if !result.Success {
			t.Error("Execute() result.Success = false, want true")
		}
	})

	// 测试执行失败的情况
	t.Run("执行失败", func(t *testing.T) {
		executor := NewMockTaskExecutor("failing-executor", func(ctx context.Context, params interface{}) error {
			return context.DeadlineExceeded
		})

		ctx := context.Background()
		result, err := executor.Execute(ctx, nil)
		if err == nil {
			t.Error("Execute() error = nil, want error")
		}
		if result.Success {
			t.Error("Execute() result.Success = true, want false")
		}
	})

	// 测试上下文取消
	t.Run("上下文取消", func(t *testing.T) {
		executor := NewMockTaskExecutor("context-executor", func(ctx context.Context, params interface{}) error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
				return nil
			}
		})

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // 立即取消上下文

		result, err := executor.Execute(ctx, nil)
		if err != context.Canceled {
			t.Errorf("Execute() error = %v, want %v", err, context.Canceled)
		}
		if result.Success {
			t.Error("Execute() result.Success = true, want false")
		}
	})
}

// BenchmarkTaskStatusString 基准测试TaskStatus.String()方法
// 测试String方法的性能表现
func BenchmarkTaskStatusString(b *testing.B) {
	status := timer.TaskStatusRunning
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = status.String()
	}
}

// BenchmarkDefaultSchedulerConfig 基准测试DefaultSchedulerConfig函数
// 测试默认配置创建的性能表现
func BenchmarkDefaultSchedulerConfig(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = timer.DefaultSchedulerConfig()
	}
} 