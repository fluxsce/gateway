package timer

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"gohub/pkg/timer"
)

// TestTimerIntegration 定时任务模块集成测试
// 测试完整的任务调度、执行和监控流程
func TestTimerIntegration(t *testing.T) {
	// 创建存储和调度器
	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "IntegrationTestScheduler",
		MaxWorkers:     2,
		QueueSize:      10,
		DefaultTimeout: time.Second * 5,
		DefaultRetries: 2,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	// 启动调度器
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	t.Run("完整工作流程测试", func(t *testing.T) {
		// 创建任务执行器
		var executionCount int
		var mu sync.Mutex
		executor := NewTestTaskExecutor("integration-executor", func(ctx context.Context, params interface{}) error {
			mu.Lock()
			executionCount++
			mu.Unlock()
			time.Sleep(time.Millisecond * 100) // 模拟任务执行时间
			return nil
		})

		// 创建任务配置
		taskConfig := CreateTestTaskConfig("integration-001", "集成测试任务", timer.ScheduleTypeInterval)
		taskConfig.Interval = time.Millisecond * 500 // 500毫秒间隔

		// 1. 添加任务
		err := scheduler.AddTask(taskConfig, executor)
		if err != nil {
			t.Errorf("添加任务失败: %v", err)
		}

		// 2. 验证任务已添加
		taskInfo, err := scheduler.GetTask("integration-001")
		if err != nil {
			t.Errorf("获取任务失败: %v", err)
		}
		if taskInfo == nil {
			t.Fatal("任务信息为nil")
		}
		if taskInfo.Config.Name != "集成测试任务" {
			t.Errorf("任务名称 = %s, want 集成测试任务", taskInfo.Config.Name)
		}

		// 3. 手动触发任务
		err = scheduler.TriggerTask("integration-001", "manual-trigger")
		if err != nil {
			t.Errorf("手动触发任务失败: %v", err)
		}

		// 4. 等待任务执行
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return executionCount > 0
		}, time.Second*3, time.Millisecond*100)

		if !success {
			t.Error("任务未在预期时间内执行")
		}

		// 5. 等待更多执行（由于间隔调度）
		time.Sleep(time.Second * 2)

		// 6. 检查执行次数
		mu.Lock()
		finalCount := executionCount
		mu.Unlock()

		if finalCount < 2 {
			t.Errorf("执行次数 = %d, want >= 2", finalCount)
		}

		// 7. 停止任务
		err = scheduler.StopTask("integration-001")
		if err != nil {
			t.Errorf("停止任务失败: %v", err)
		}

		// 8. 验证任务已停止
		taskInfo, err = scheduler.GetTask("integration-001")
		if err != nil {
			t.Errorf("获取停止后的任务失败: %v", err)
		}
		if taskInfo.Config.Enabled {
			t.Error("停止后任务仍处于启用状态")
		}

		// 9. 获取任务历史
		history, err := scheduler.GetTaskHistory("integration-001", 0)
		if err != nil {
			t.Errorf("获取任务历史失败: %v", err)
		}
		if len(history) == 0 {
			t.Error("任务历史为空")
		}

		// 10. 移除任务
		err = scheduler.RemoveTask("integration-001")
		if err != nil {
			t.Errorf("移除任务失败: %v", err)
		}

		// 11. 验证任务已移除
		_, err = scheduler.GetTask("integration-001")
		if err == nil {
			t.Error("移除后仍能获取任务")
		}
	})
}

// TestMultipleTasksIntegration 多任务集成测试
// 测试多个任务同时运行的场景
func TestMultipleTasksIntegration(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	t.Run("多任务并发执行", func(t *testing.T) {
		taskCount := 5
		var totalExecutions int
		var mu sync.Mutex

		// 创建多个任务
		for i := 0; i < taskCount; i++ {
			taskID := fmt.Sprintf("multi-task-%03d", i)
			executor := NewTestTaskExecutor(fmt.Sprintf("executor-%d", i), func(ctx context.Context, params interface{}) error {
				mu.Lock()
				totalExecutions++
				mu.Unlock()
				time.Sleep(time.Millisecond * 50)
				return nil
			})

			config := CreateTestTaskConfig(taskID, fmt.Sprintf("多任务测试%d", i), timer.ScheduleTypeOnce)
			err := scheduler.AddTask(config, executor)
			if err != nil {
				t.Errorf("添加任务%d失败: %v", i, err)
			}
		}

		// 同时触发所有任务
		for i := 0; i < taskCount; i++ {
			taskID := fmt.Sprintf("multi-task-%03d", i)
			err := scheduler.TriggerTask(taskID, fmt.Sprintf("params-%d", i))
			if err != nil {
				t.Errorf("触发任务%d失败: %v", i, err)
			}
		}

		// 等待所有任务完成
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return totalExecutions >= taskCount
		}, time.Second*5, time.Millisecond*100)

		if !success {
			mu.Lock()
			t.Errorf("多任务执行未完成，执行次数: %d, 期望: %d", totalExecutions, taskCount)
			mu.Unlock()
		}

		// 验证任务列表
		tasks, err := scheduler.ListTasks()
		if err != nil {
			t.Errorf("获取任务列表失败: %v", err)
		}
		if len(tasks) != taskCount {
			t.Errorf("任务列表长度 = %d, want %d", len(tasks), taskCount)
		}
	})
}

// TestTaskFailureRecovery 任务失败恢复测试
// 测试任务执行失败时的重试和恢复机制
func TestTaskFailureRecovery(t *testing.T) {
	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "FailureRecoveryScheduler",
		MaxWorkers:     1,
		QueueSize:      5,
		DefaultTimeout: time.Second * 2,
		DefaultRetries: 3,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	t.Run("任务失败重试机制", func(t *testing.T) {
		var attemptCount int
		var mu sync.Mutex

		// 创建会失败的执行器
		executor := NewTestTaskExecutor("failure-executor", func(ctx context.Context, params interface{}) error {
			mu.Lock()
			attemptCount++
			currentAttempt := attemptCount
			mu.Unlock()

			// 前3次失败，第4次成功
			if currentAttempt <= 3 {
				return fmt.Errorf("模拟第%d次失败", currentAttempt)
			}
			return nil
		})

		// 创建任务配置
		taskConfig := CreateTestTaskConfig("failure-001", "失败重试测试", timer.ScheduleTypeOnce)
		taskConfig.MaxRetries = 4 // 允许重试4次

		err := scheduler.AddTask(taskConfig, executor)
		if err != nil {
			t.Errorf("添加失败任务失败: %v", err)
		}

		// 触发任务
		err = scheduler.TriggerTask("failure-001", nil)
		if err != nil {
			t.Errorf("触发失败任务失败: %v", err)
		}

		// 等待任务完成（包括重试）
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return attemptCount >= 4
		}, time.Second*10, time.Millisecond*100)

		if !success {
			t.Error("任务重试未在预期时间内完成")
		}

		// 验证重试次数
		mu.Lock()
		if attemptCount != 4 {
			t.Errorf("重试次数 = %d, want 4", attemptCount)
		}
		mu.Unlock()

		// 等待任务处理完成
		time.Sleep(time.Second)

		// 检查任务历史
		history, err := scheduler.GetTaskHistory("failure-001", 1)
		if err != nil {
			t.Errorf("获取失败任务历史失败: %v", err)
		}

		if len(history) > 0 {
			result := history[0]
			if result.Status != timer.TaskStatusCompleted {
				t.Errorf("最终任务状态 = %v, want %v", result.Status, timer.TaskStatusCompleted)
			}
			if result.RetryCount != 3 { // 重试3次后成功
				t.Errorf("任务重试次数 = %d, want 3", result.RetryCount)
			}
		}
	})
}

// TestSchedulerLifecycle 调度器生命周期测试
// 测试调度器的启动、停止和重启流程
func TestSchedulerLifecycle(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	t.Run("调度器生命周期管理", func(t *testing.T) {
		// 1. 初始状态验证
		if scheduler.IsRunning() {
			t.Error("新创建的调度器不应该处于运行状态")
		}

		// 2. 启动调度器
		err := scheduler.Start()
		if err != nil {
			t.Errorf("启动调度器失败: %v", err)
		}
		if !scheduler.IsRunning() {
			t.Error("启动后调度器应该处于运行状态")
		}

		// 3. 在运行状态下添加任务
		executor := NewTestTaskExecutor("lifecycle-executor", nil)
		config := CreateTestTaskConfig("lifecycle-001", "生命周期测试", timer.ScheduleTypeOnce)
		err = scheduler.AddTask(config, executor)
		if err != nil {
			t.Errorf("在运行状态下添加任务失败: %v", err)
		}

		// 4. 停止调度器
		err = scheduler.Stop()
		if err != nil {
			t.Errorf("停止调度器失败: %v", err)
		}
		if scheduler.IsRunning() {
			t.Error("停止后调度器不应该处于运行状态")
		}

		// 5. 验证任务仍然存在
		taskInfo, err := scheduler.GetTask("lifecycle-001")
		if err != nil {
			t.Errorf("停止后获取任务失败: %v", err)
		}
		if taskInfo == nil {
			t.Error("停止后任务信息不应该为nil")
		}

		// 6. 重新启动调度器
		err = scheduler.Start()
		if err != nil {
			t.Errorf("重新启动调度器失败: %v", err)
		}
		if !scheduler.IsRunning() {
			t.Error("重新启动后调度器应该处于运行状态")
		}

		// 7. 验证任务仍然可用
		err = scheduler.TriggerTask("lifecycle-001", nil)
		if err != nil {
			t.Errorf("重启后触发任务失败: %v", err)
		}

		// 8. 最终清理
		scheduler.Stop()
	})
}

// TestCronTaskIntegration Cron任务集成测试
// 测试Cron表达式调度的实际工作情况
func TestCronTaskIntegration(t *testing.T) {
	storage := NewMemoryTaskStorage()
	scheduler := timer.NewStandardScheduler(nil, storage)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	t.Run("Cron任务调度", func(t *testing.T) {
		var executionCount int
		var mu sync.Mutex

		executor := NewTestTaskExecutor("cron-executor", func(ctx context.Context, params interface{}) error {
			mu.Lock()
			executionCount++
			mu.Unlock()
			return nil
		})

		// 创建每秒执行的Cron任务
		config := CreateTestTaskConfig("cron-001", "Cron测试任务", timer.ScheduleTypeCron)
		config.CronExpr = "*/1 * * * * *" // 每秒执行一次

		err := scheduler.AddTask(config, executor)
		if err != nil {
			t.Errorf("添加Cron任务失败: %v", err)
		}

		// 等待任务执行几次
		time.Sleep(time.Second * 3)

		// 验证执行次数
		mu.Lock()
		if executionCount < 2 {
			t.Errorf("Cron任务执行次数 = %d, want >= 2", executionCount)
		}
		mu.Unlock()

		// 停止任务
		err = scheduler.StopTask("cron-001")
		if err != nil {
			t.Errorf("停止Cron任务失败: %v", err)
		}

		// 记录停止时的执行次数
		mu.Lock()
		stopCount := executionCount
		mu.Unlock()

		// 等待一段时间，验证任务确实停止了
		time.Sleep(time.Second * 2)

		mu.Lock()
		finalCount := executionCount
		mu.Unlock()

		if finalCount > stopCount {
			t.Errorf("停止后任务仍在执行，停止时: %d, 最终: %d", stopCount, finalCount)
		}
	})
}

// TestPerformanceUnderLoad 负载性能测试
// 测试调度器在高负载下的性能表现
func TestPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试（使用 -short 标志）")
	}

	storage := NewMemoryTaskStorage()
	config := &timer.SchedulerConfig{
		Name:           "PerformanceTestScheduler",
		MaxWorkers:     5,
		QueueSize:      100,
		DefaultTimeout: time.Second * 10,
		DefaultRetries: 1,
	}
	scheduler := timer.NewStandardScheduler(config, storage)

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("启动调度器失败: %v", err)
	}
	defer scheduler.Stop()

	t.Run("高负载性能测试", func(t *testing.T) {
		taskCount := 50
		var completedCount int64
		var mu sync.Mutex

		startTime := time.Now()

		// 创建大量任务
		for i := 0; i < taskCount; i++ {
			taskID := fmt.Sprintf("perf-task-%03d", i)
			executor := NewTestTaskExecutor(fmt.Sprintf("perf-executor-%d", i), func(ctx context.Context, params interface{}) error {
				// 模拟一些工作
				time.Sleep(time.Millisecond * 10)
				mu.Lock()
				completedCount++
				mu.Unlock()
				return nil
			})

			config := CreateTestTaskConfig(taskID, fmt.Sprintf("性能测试任务%d", i), timer.ScheduleTypeOnce)
			err := scheduler.AddTask(config, executor)
			if err != nil {
				t.Errorf("添加性能测试任务%d失败: %v", i, err)
			}
		}

		// 同时触发所有任务
		for i := 0; i < taskCount; i++ {
			taskID := fmt.Sprintf("perf-task-%03d", i)
			err := scheduler.TriggerTask(taskID, nil)
			if err != nil {
				t.Errorf("触发性能测试任务%d失败: %v", i, err)
			}
		}

		// 等待所有任务完成
		success := WaitForCondition(func() bool {
			mu.Lock()
			defer mu.Unlock()
			return completedCount == int64(taskCount)
		}, time.Second*30, time.Millisecond*100)

		duration := time.Since(startTime)

		if !success {
			mu.Lock()
			t.Errorf("性能测试未完成，完成数量: %d, 期望: %d", completedCount, taskCount)
			mu.Unlock()
		} else {
			t.Logf("性能测试完成：%d个任务在%v内执行完毕", taskCount, duration)
			t.Logf("平均每个任务耗时：%v", duration/time.Duration(taskCount))
		}

		// 验证所有任务都有历史记录
		totalResults := 0
		for i := 0; i < taskCount; i++ {
			taskID := fmt.Sprintf("perf-task-%03d", i)
			history, err := scheduler.GetTaskHistory(taskID, 0)
			if err != nil {
				t.Errorf("获取任务%s历史失败: %v", taskID, err)
			}
			totalResults += len(history)
		}

		if totalResults < taskCount {
			t.Errorf("历史记录数量 = %d, want >= %d", totalResults, taskCount)
		}
	})
} 