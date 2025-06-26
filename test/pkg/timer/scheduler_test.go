package timer

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"gohub/pkg/timer"
)

// TestNewStandardScheduler 测试标准调度器的创建
// 验证NewStandardScheduler函数的基本功能和默认配置处理
func TestNewStandardScheduler(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 测试使用自定义配置创建调度器
	t.Run("使用自定义配置", func(t *testing.T) {
		config := &timer.SchedulerConfig{
			Name:           "TestScheduler",
			MaxWorkers:     3,
			QueueSize:      50,
			DefaultTimeout: time.Minute * 10,
			DefaultRetries: 2,
		}

		scheduler := timer.NewStandardScheduler(config, storage)
		if scheduler == nil {
			t.Fatal("NewStandardScheduler returned nil")
		}

		// 验证调度器状态
		if scheduler.IsRunning() {
			t.Error("新创建的调度器不应该处于运行状态")
		}
	})

	// 测试使用nil配置创建调度器（应使用默认配置）
	t.Run("使用默认配置", func(t *testing.T) {
		scheduler := timer.NewStandardScheduler(nil, storage)
		if scheduler == nil {
			t.Fatal("NewStandardScheduler with nil config returned nil")
		}

		// 验证调度器状态
		if scheduler.IsRunning() {
			t.Error("新创建的调度器不应该处于运行状态")
		}
	})
}

// TestSchedulerStartStop 测试调度器的启动和停止
// 验证调度器生命周期管理的正确性
func TestSchedulerStartStop(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 测试启动调度器
	t.Run("启动调度器", func(t *testing.T) {
		err := scheduler.Start()
		if err != nil {
			t.Fatalf("启动调度器失败: %v", err)
		}

		if !scheduler.IsRunning() {
			t.Error("调度器启动后应该处于运行状态")
		}

		// 清理：停止调度器
		defer func() {
			if err := scheduler.Stop(); err != nil {
				t.Errorf("停止调度器失败: %v", err)
			}
		}()
	})

	// 测试重复启动调度器（应该返回错误）
	t.Run("重复启动调度器", func(t *testing.T) {
		// 先启动调度器
		if err := scheduler.Start(); err != nil {
			t.Fatalf("首次启动调度器失败: %v", err)
		}

		// 尝试再次启动
		err := scheduler.Start()
		if err == nil {
			t.Error("重复启动调度器应该返回错误")
		}

		// 清理
		scheduler.Stop()
	})

	// 测试停止未运行的调度器
	t.Run("停止未运行的调度器", func(t *testing.T) {
		newScheduler := timer.NewStandardScheduler(nil, storage)
		err := newScheduler.Stop()
		if err == nil {
			t.Error("停止未运行的调度器应该返回错误")
		}
	})
}

// TestAddTask 测试添加任务功能
// 验证任务添加的各种场景和错误处理
func TestAddTask(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 测试添加有效任务
	t.Run("添加有效任务", func(t *testing.T) {
		config := CreateTestTaskConfig("test-add-001", "添加测试任务", timer.ScheduleTypeOnce)
		executor := NewTestTaskExecutor("test-executor", nil)

		err := scheduler.AddTask(config, executor)
		if err != nil {
			t.Errorf("添加有效任务失败: %v", err)
		}

		// 验证任务是否已保存
		if storage.GetTaskCount() != 1 {
			t.Errorf("任务数量 = %d, want 1", storage.GetTaskCount())
		}
	})

	// 测试添加重复ID的任务
	t.Run("添加重复ID的任务", func(t *testing.T) {
		config := CreateTestTaskConfig("test-add-001", "重复ID任务", timer.ScheduleTypeOnce)
		executor := NewTestTaskExecutor("duplicate-executor", nil)

		err := scheduler.AddTask(config, executor)
		if err == nil {
			t.Error("添加重复ID的任务应该返回错误")
		}
	})

	// 测试添加nil执行器的任务
	t.Run("添加nil执行器的任务", func(t *testing.T) {
		config := CreateTestTaskConfig("test-add-002", "nil执行器任务", timer.ScheduleTypeOnce)

		err := scheduler.AddTask(config, nil)
		if err == nil {
			t.Error("添加nil执行器的任务应该返回错误")
		}
	})

	// 测试添加无效配置的任务
	t.Run("添加无效配置的任务", func(t *testing.T) {
		// 创建无效配置（ID为空）
		config := &timer.TaskConfig{
			ID:           "", // 无效的空ID
			Name:         "无效任务",
			ScheduleType: timer.ScheduleTypeOnce,
			Enabled:      true,
		}
		executor := NewTestTaskExecutor("invalid-executor", nil)

		err := scheduler.AddTask(config, executor)
		if err == nil {
			t.Error("添加无效配置的任务应该返回错误")
		}
	})
}

// TestRemoveTask 测试移除任务功能
// 验证任务移除的正确性和相关资源清理
func TestRemoveTask(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 先添加一个任务
	config := CreateTestTaskConfig("test-remove-001", "移除测试任务", timer.ScheduleTypeInterval)
	executor := NewTestTaskExecutor("remove-executor", nil)
	scheduler.AddTask(config, executor)

	// 测试移除存在的任务
	t.Run("移除存在的任务", func(t *testing.T) {
		err := scheduler.RemoveTask("test-remove-001")
		if err != nil {
			t.Errorf("移除存在的任务失败: %v", err)
		}

		// 验证任务已被移除
		if storage.GetTaskCount() != 0 {
			t.Errorf("移除后任务数量 = %d, want 0", storage.GetTaskCount())
		}
	})

	// 测试移除不存在的任务
	t.Run("移除不存在的任务", func(t *testing.T) {
		err := scheduler.RemoveTask("non-existent-task")
		// 这里不应该返回错误，因为删除不存在的任务是安全的
		if err != nil {
			t.Errorf("移除不存在的任务不应该返回错误: %v", err)
		}
	})
}

// TestGetTask 测试获取任务功能
// 验证任务信息获取的正确性
func TestGetTask(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 先添加一个任务
	config := CreateTestTaskConfig("test-get-001", "获取测试任务", timer.ScheduleTypeInterval)
	executor := NewTestTaskExecutor("get-executor", nil)
	scheduler.AddTask(config, executor)

	// 测试获取存在的任务
	t.Run("获取存在的任务", func(t *testing.T) {
		taskInfo, err := scheduler.GetTask("test-get-001")
		if err != nil {
			t.Errorf("获取存在的任务失败: %v", err)
		}

		if taskInfo == nil {
			t.Fatal("获取的任务信息为nil")
		}

		if taskInfo.Config.ID != "test-get-001" {
			t.Errorf("任务ID = %s, want test-get-001", taskInfo.Config.ID)
		}

		if taskInfo.Config.Name != "获取测试任务" {
			t.Errorf("任务名称 = %s, want 获取测试任务", taskInfo.Config.Name)
		}
	})

	// 测试获取不存在的任务
	t.Run("获取不存在的任务", func(t *testing.T) {
		_, err := scheduler.GetTask("non-existent-task")
		if err == nil {
			t.Error("获取不存在的任务应该返回错误")
		}
	})
}

// TestListTasks 测试列出所有任务功能
// 验证任务列表获取的正确性
func TestListTasks(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 测试空任务列表
	t.Run("空任务列表", func(t *testing.T) {
		tasks, err := scheduler.ListTasks()
		if err != nil {
			t.Errorf("获取空任务列表失败: %v", err)
		}

		if len(tasks) != 0 {
			t.Errorf("空任务列表长度 = %d, want 0", len(tasks))
		}
	})

	// 添加多个任务
	configs := []*timer.TaskConfig{
		CreateTestTaskConfig("list-001", "列表任务1", timer.ScheduleTypeOnce),
		CreateTestTaskConfig("list-002", "列表任务2", timer.ScheduleTypeInterval),
		CreateTestTaskConfig("list-003", "列表任务3", timer.ScheduleTypeCron),
	}

	for i, config := range configs {
		executor := NewTestTaskExecutor(fmt.Sprintf("list-executor-%d", i+1), nil)
		scheduler.AddTask(config, executor)
	}

	// 测试获取多个任务的列表
	t.Run("获取多个任务的列表", func(t *testing.T) {
		tasks, err := scheduler.ListTasks()
		if err != nil {
			t.Errorf("获取任务列表失败: %v", err)
		}

		if len(tasks) != len(configs) {
			t.Errorf("任务列表长度 = %d, want %d", len(tasks), len(configs))
		}

		// 验证任务信息
		taskIDs := make(map[string]bool)
		for _, task := range tasks {
			taskIDs[task.Config.ID] = true
		}

		for _, config := range configs {
			if !taskIDs[config.ID] {
				t.Errorf("任务列表中缺少任务: %s", config.ID)
			}
		}
	})
}

// TestStartStopTask 测试启动和停止单个任务
// 验证单个任务的控制功能
func TestStartStopTask(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 添加一个禁用的任务
	config := CreateTestTaskConfig("test-control-001", "控制测试任务", timer.ScheduleTypeInterval)
	config.Enabled = false // 初始状态为禁用
	executor := NewTestTaskExecutor("control-executor", nil)
	scheduler.AddTask(config, executor)

	// 测试启动任务
	t.Run("启动任务", func(t *testing.T) {
		err := scheduler.StartTask("test-control-001")
		if err != nil {
			t.Errorf("启动任务失败: %v", err)
		}

		// 验证任务已启用
		taskInfo, _ := scheduler.GetTask("test-control-001")
		if taskInfo == nil || !taskInfo.Config.Enabled {
			t.Error("启动后任务应该处于启用状态")
		}
	})

	// 测试停止任务
	t.Run("停止任务", func(t *testing.T) {
		err := scheduler.StopTask("test-control-001")
		if err != nil {
			t.Errorf("停止任务失败: %v", err)
		}

		// 验证任务已禁用
		taskInfo, _ := scheduler.GetTask("test-control-001")
		if taskInfo == nil || taskInfo.Config.Enabled {
			t.Error("停止后任务应该处于禁用状态")
		}
	})

	// 测试启动不存在的任务
	t.Run("启动不存在的任务", func(t *testing.T) {
		err := scheduler.StartTask("non-existent-task")
		if err == nil {
			t.Error("启动不存在的任务应该返回错误")
		}
	})

	// 测试停止不存在的任务
	t.Run("停止不存在的任务", func(t *testing.T) {
		err := scheduler.StopTask("non-existent-task")
		if err == nil {
			t.Error("停止不存在的任务应该返回错误")
		}
	})
}

// TestTriggerTask 测试手动触发任务
// 验证任务手动执行功能
func TestTriggerTask(t *testing.T) {
	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "TriggerTestScheduler",
		MaxWorkers:     2,
		QueueSize:      10,
		DefaultTimeout: time.Second * 5,
		DefaultRetries: 1,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	// 启动调度器以便处理任务队列
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	// 创建可追踪执行的任务执行器
	var executionCount int
	var mu sync.Mutex
	executor := NewTestTaskExecutor("trigger-executor", func(ctx context.Context, params interface{}) error {
		mu.Lock()
		executionCount++
		mu.Unlock()
		return nil
	})

	// 添加任务
	taskConfig := CreateTestTaskConfig("test-trigger-001", "触发测试任务", timer.ScheduleTypeOnce)
	scheduler.AddTask(taskConfig, executor)

	// 测试手动触发任务
	t.Run("手动触发任务", func(t *testing.T) {
		err := scheduler.TriggerTask("test-trigger-001", "trigger-params")
		if err != nil {
			t.Errorf("手动触发任务失败: %v", err)
		}

		// 等待任务执行完成
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return executionCount > 0
		}, time.Second*3, time.Millisecond*100)

		if !success {
			t.Error("任务未在预期时间内执行")
		}

		// 验证执行次数
		mu.Lock()
		if executionCount != 1 {
			t.Errorf("任务执行次数 = %d, want 1", executionCount)
		}
		mu.Unlock()
	})

	// 测试触发不存在的任务
	t.Run("触发不存在的任务", func(t *testing.T) {
		err := scheduler.TriggerTask("non-existent-task", nil)
		if err == nil {
			t.Error("触发不存在的任务应该返回错误")
		}
	})

	// 测试使用nil参数触发任务
	t.Run("使用nil参数触发任务", func(t *testing.T) {
		mu.Lock()
		executionCount = 0 // 重置计数
		mu.Unlock()

		err := scheduler.TriggerTask("test-trigger-001", nil)
		if err != nil {
			t.Errorf("使用nil参数触发任务失败: %v", err)
		}

		// 等待任务执行完成
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return executionCount > 0
		}, time.Second*3, time.Millisecond*100)

		if !success {
			t.Error("使用nil参数的任务未在预期时间内执行")
		}
	})
}

// TestGetTaskHistory 测试获取任务历史记录
// 验证任务执行历史的记录和查询功能
func TestGetTaskHistory(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 启动调度器
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	// 创建任务执行器
	executor := NewTestTaskExecutor("history-executor", func(ctx context.Context, params interface{}) error {
		time.Sleep(time.Millisecond * 10) // 模拟任务执行时间
		return nil
	})

	// 添加任务
	taskConfig := CreateTestTaskConfig("test-history-001", "历史测试任务", timer.ScheduleTypeOnce)
	scheduler.AddTask(taskConfig, executor)

	// 手动触发任务多次以生成历史记录
	for i := 0; i < 3; i++ {
		err := scheduler.TriggerTask("test-history-001", fmt.Sprintf("params-%d", i))
		if err != nil {
			t.Errorf("第%d次触发任务失败: %v", i+1, err)
		}
		time.Sleep(time.Millisecond * 50) // 等待任务执行完成
	}

	// 等待所有任务执行完成
	time.Sleep(time.Second)

	// 测试获取任务历史记录
	t.Run("获取任务历史记录", func(t *testing.T) {
		history, err := scheduler.GetTaskHistory("test-history-001", 0)
		if err != nil {
			t.Errorf("获取任务历史记录失败: %v", err)
		}

		if len(history) < 3 {
			t.Errorf("历史记录数量 = %d, want >= 3", len(history))
		}

		// 验证历史记录按时间倒序排列
		for i := 1; i < len(history); i++ {
			if history[i-1].StartTime.Before(history[i].StartTime) {
				t.Error("历史记录应该按时间倒序排列")
			}
		}
	})

	// 测试限制历史记录数量
	t.Run("限制历史记录数量", func(t *testing.T) {
		history, err := scheduler.GetTaskHistory("test-history-001", 2)
		if err != nil {
			t.Errorf("获取限制数量的历史记录失败: %v", err)
		}

		if len(history) > 2 {
			t.Errorf("限制后历史记录数量 = %d, want <= 2", len(history))
		}
	})

	// 测试获取不存在任务的历史记录
	t.Run("获取不存在任务的历史记录", func(t *testing.T) {
		history, err := scheduler.GetTaskHistory("non-existent-task", 0)
		if err != nil {
			t.Errorf("获取不存在任务的历史记录失败: %v", err)
		}

		if len(history) != 0 {
			t.Errorf("不存在任务的历史记录数量 = %d, want 0", len(history))
		}
	})
}

// TestGetRunningTasks 测试获取正在运行的任务
// 验证运行中任务的查询功能
func TestGetRunningTasks(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	// 测试无运行任务的情况
	t.Run("无运行任务", func(t *testing.T) {
		runningTasks, err := scheduler.GetRunningTasks()
		if err != nil {
			t.Errorf("获取运行任务失败: %v", err)
		}

		if len(runningTasks) != 0 {
			t.Errorf("运行任务数量 = %d, want 0", len(runningTasks))
		}
	})

	// 添加一些任务
	configs := []*timer.TaskConfig{
		CreateTestTaskConfig("running-001", "运行任务1", timer.ScheduleTypeOnce),
		CreateTestTaskConfig("running-002", "运行任务2", timer.ScheduleTypeInterval),
	}

	for i, config := range configs {
		executor := NewTestTaskExecutor(fmt.Sprintf("running-executor-%d", i+1), nil)
		scheduler.AddTask(config, executor)
	}

	// 手动设置某些任务为运行状态（模拟实际运行场景）
	taskInfo1, _ := scheduler.GetTask("running-001")
	if taskInfo1 != nil {
		taskInfo1.Status = timer.TaskStatusRunning
		storage.SaveTaskInfo(taskInfo1)
	}

	// 测试获取运行中的任务
	t.Run("获取运行中的任务", func(t *testing.T) {
		runningTasks, err := scheduler.GetRunningTasks()
		if err != nil {
			t.Errorf("获取运行任务失败: %v", err)
		}

		// 验证至少有一个运行中的任务
		foundRunning := false
		for _, task := range runningTasks {
			if task.Status == timer.TaskStatusRunning {
				foundRunning = true
				break
			}
		}

		if !foundRunning {
			t.Error("应该找到至少一个运行中的任务")
		}
	})
}

// TestSchedulerWithFailingTasks 测试调度器处理失败任务的能力
// 验证错误处理和重试机制
func TestSchedulerWithFailingTasks(t *testing.T) {
	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "FailTestScheduler",
		MaxWorkers:     1,
		QueueSize:      5,
		DefaultTimeout: time.Second * 2,
		DefaultRetries: 2,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	// 启动调度器
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	// 创建会失败的任务执行器
	var attemptCount int
	var mu sync.Mutex
	executor := NewTestTaskExecutor("failing-executor", func(ctx context.Context, params interface{}) error {
		mu.Lock()
		attemptCount++
		currentAttempt := attemptCount
		mu.Unlock()

		// 前两次尝试失败，第三次成功
		if currentAttempt < 3 {
			return errors.New("模拟任务执行失败")
		}
		return nil
	})

	// 添加会失败的任务
	taskConfig := CreateTestTaskConfig("test-fail-001", "失败测试任务", timer.ScheduleTypeOnce)
	taskConfig.MaxRetries = 3
	scheduler.AddTask(taskConfig, executor)

	// 触发任务
	err = scheduler.TriggerTask("test-fail-001", nil)
	if err != nil {
		t.Errorf("触发失败任务失败: %v", err)
	}

	// 等待任务执行完成（包括重试）
	success := WaitForCondition(func() bool {
		mu.Lock()
		defer mu.Unlock()
		return attemptCount >= 3
	}, time.Second*10, time.Millisecond*100)

	if !success {
		t.Error("任务未在预期时间内完成重试")
	}

	// 验证重试次数
	mu.Lock()
	if attemptCount < 3 {
		t.Errorf("任务重试次数 = %d, want >= 3", attemptCount)
	}
	mu.Unlock()

	// 等待任务完全处理完成
	time.Sleep(time.Second)

	// 检查任务历史记录
	history, err := scheduler.GetTaskHistory("test-fail-001", 1)
	if err != nil {
		t.Errorf("获取失败任务历史记录失败: %v", err)
	}

	if len(history) > 0 {
		result := history[0]
		if result.RetryCount != 2 { // 重试2次后成功
			t.Errorf("任务结果重试次数 = %d, want 2", result.RetryCount)
		}
		if result.Status != timer.TaskStatusCompleted {
			t.Errorf("最终任务状态 = %v, want %v", result.Status, timer.TaskStatusCompleted)
		}
	}
}

// TestSchedulerConcurrency 测试调度器的并发处理能力
// 验证多任务并发执行的正确性
func TestSchedulerConcurrency(t *testing.T) {
	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "ConcurrencyTestScheduler",
		MaxWorkers:     3,
		QueueSize:      20,
		DefaultTimeout: time.Second * 5,
		DefaultRetries: 1,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	// 启动调度器
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	// 创建多个任务
	taskCount := 10
	var completedCount int64
	var mu sync.Mutex

	for i := 0; i < taskCount; i++ {
		taskID := fmt.Sprintf("concurrent-%03d", i)
		executor := NewTestTaskExecutor(fmt.Sprintf("concurrent-executor-%d", i), func(ctx context.Context, params interface{}) error {
			// 模拟任务执行时间
			time.Sleep(time.Millisecond * 100)
			
			mu.Lock()
			completedCount++
			mu.Unlock()
			return nil
		})

		taskConfig := CreateTestTaskConfig(taskID, fmt.Sprintf("并发任务%d", i), timer.ScheduleTypeOnce)
		err := scheduler.AddTask(taskConfig, executor)
		if err != nil {
			t.Errorf("添加并发任务%d失败: %v", i, err)
		}
	}

	// 同时触发所有任务
	for i := 0; i < taskCount; i++ {
		taskID := fmt.Sprintf("concurrent-%03d", i)
		err := scheduler.TriggerTask(taskID, fmt.Sprintf("params-%d", i))
		if err != nil {
			t.Errorf("触发并发任务%d失败: %v", i, err)
		}
	}

	// 等待所有任务完成
	success := WaitForCondition(func() bool {
		mu.Lock()
		defer mu.Unlock()
		return completedCount == int64(taskCount)
	}, time.Second*10, time.Millisecond*100)

	if !success {
		mu.Lock()
		t.Errorf("并发任务未全部完成，完成数量: %d, 期望: %d", completedCount, taskCount)
		mu.Unlock()
	}
}

// BenchmarkSchedulerAddTask 基准测试任务添加性能
func BenchmarkSchedulerAddTask(b *testing.B) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)
	executor := NewTestTaskExecutor("bench-executor", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskID := fmt.Sprintf("bench-task-%d", i)
		config := CreateTestTaskConfig(taskID, "基准测试任务", timer.ScheduleTypeOnce)
		scheduler.AddTask(config, executor)
	}
}

// BenchmarkSchedulerTriggerTask 基准测试任务触发性能
func BenchmarkSchedulerTriggerTask(b *testing.B) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)
	executor := NewTestTaskExecutor("bench-trigger-executor", nil)

	// 预先添加任务
	config := CreateTestTaskConfig("bench-trigger-task", "基准触发任务", timer.ScheduleTypeOnce)
	scheduler.AddTask(config, executor)
	scheduler.Start()
	defer scheduler.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler.TriggerTask("bench-trigger-task", fmt.Sprintf("params-%d", i))
	}
} 