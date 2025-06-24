// Package examples 提供定时任务组件的使用示例
package examples

import (
	"context"
	"fmt"
	"log"
	"time"
	
	"gohub/pkg/timer"
	"gohub/pkg/timer/cron"
	"gohub/pkg/timer/executor"
	"gohub/pkg/timer/storage"
)

// BasicUsageExample 基本使用示例
func BasicUsageExample() {
	fmt.Println("=== 基本使用示例 ===")
	
	// 1. 创建存储和调度器
	store := storage.NewMemoryStorage()
	scheduler := timer.NewStandardScheduler(nil, store)
	
	// 2. 创建一个简单的日志任务
	logExecutor := executor.NewLogExecutor()
	
	taskConfig := &timer.TaskConfig{
		ID:           "log-task-1",
		Name:         "每分钟日志任务",
		Description:  "每分钟输出一条日志",
		Priority:     timer.TaskPriorityNormal,
		ScheduleType: timer.ScheduleTypeInterval,
		Interval:     time.Minute,
		Enabled:      true,
		Params:       "Hello from scheduled task!",
	}
	
	// 3. 添加任务
	if err := scheduler.AddTask(taskConfig, logExecutor); err != nil {
		log.Fatalf("Failed to add task: %v", err)
	}
	
	// 4. 启动调度器
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	
	// 5. 等待一段时间观察任务执行
	time.Sleep(3 * time.Minute)
	
	// 6. 停止调度器
	if err := scheduler.Stop(); err != nil {
		log.Printf("Failed to stop scheduler: %v", err)
	}
	
	// 7. 查看任务执行历史
	results, err := scheduler.GetTaskHistory("log-task-1", 10)
	if err != nil {
		log.Printf("Failed to get task history: %v", err)
	} else {
		fmt.Printf("Task executed %d times\n", len(results))
		for _, result := range results {
			fmt.Printf("- %s: %s (Duration: %v)\n", 
				result.StartTime.Format("15:04:05"), 
				result.Status.String(), 
				result.Duration)
		}
	}
}

// CronTaskExample Cron任务示例
func CronTaskExample() {
	fmt.Println("=== Cron任务示例 ===")
	
	store := storage.NewMemoryStorage()
	scheduler := timer.NewStandardScheduler(nil, store)
	
	// 创建一个每5分钟执行的任务
	taskConfig := &timer.TaskConfig{
		ID:           "cron-task-1",
		Name:         "每5分钟执行的任务",
		Description:  "使用Cron表达式每5分钟执行一次",
		ScheduleType: timer.ScheduleTypeCron,
		CronExpr:     "*/5 * * * *", // 每5分钟
		Enabled:      true,
		Params:       "Cron task executed!",
	}
	
	logExecutor := executor.NewLogExecutor()
	
	if err := scheduler.AddTask(taskConfig, logExecutor); err != nil {
		log.Fatalf("Failed to add cron task: %v", err)
	}
	
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	
	// 演示Cron表达式解析
	fmt.Println("Cron表达式示例:")
	expressions := []string{
		cron.EveryMinute,
		cron.Hourly,
		cron.Daily,
		cron.Weekly,
		"0 9 * * 1-5", // 工作日上午9点
	}
	
	for _, expr := range expressions {
		schedule, err := cron.ParseCron(expr)
		if err != nil {
			fmt.Printf("- %s: 解析失败 - %v\n", expr, err)
			continue
		}
		
		next := schedule.Next(time.Now())
		fmt.Printf("- %s: 下次执行时间 %s\n", expr, next.Format("2006-01-02 15:04:05"))
	}
	
	time.Sleep(time.Minute)
	scheduler.Stop()
}

// CustomExecutorExample 自定义执行器示例
func CustomExecutorExample() {
	fmt.Println("=== 自定义执行器示例 ===")
	
	store := storage.NewMemoryStorage()
	scheduler := timer.NewStandardScheduler(nil, store)
	
	// 创建自定义函数执行器
	customExecutor := executor.NewFunctionExecutor("CustomTask", func(ctx context.Context, params interface{}) error {
		fmt.Printf("自定义任务执行: %v\n", params)
		
		// 模拟一些工作
		select {
		case <-time.After(time.Second):
			fmt.Println("任务完成")
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	
	taskConfig := &timer.TaskConfig{
		ID:           "custom-task-1",
		Name:         "自定义任务",
		Description:  "使用自定义执行器的任务",
		ScheduleType: timer.ScheduleTypeInterval,
		Interval:     30 * time.Second,
		Enabled:      true,
		Params:       map[string]interface{}{
			"message": "Hello from custom executor!",
			"count":   42,
		},
	}
	
	if err := scheduler.AddTask(taskConfig, customExecutor); err != nil {
		log.Fatalf("Failed to add custom task: %v", err)
	}
	
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	
	// 手动触发任务
	fmt.Println("手动触发任务...")
	if err := scheduler.TriggerTask("custom-task-1", "手动触发的参数"); err != nil {
		log.Printf("Failed to trigger task: %v", err)
	}
	
	time.Sleep(2 * time.Minute)
	scheduler.Stop()
}

// HTTPTaskExample HTTP任务示例
func HTTPTaskExample() {
	fmt.Println("=== HTTP任务示例 ===")
	
	store := storage.NewMemoryStorage()
	scheduler := timer.NewStandardScheduler(nil, store)
	
	// 创建HTTP健康检查任务
	httpExecutor := executor.NewHTTPExecutor()
	
	taskConfig := &timer.TaskConfig{
		ID:           "health-check-1",
		Name:         "健康检查任务",
		Description:  "定期检查服务健康状态",
		ScheduleType: timer.ScheduleTypeInterval,
		Interval:     30 * time.Second,
		Enabled:      true,
		Timeout:      10 * time.Second,
		MaxRetries:   2,
		Params: map[string]interface{}{
			"url":    "https://httpbin.org/status/200",
			"method": "GET",
		},
	}
	
	if err := scheduler.AddTask(taskConfig, httpExecutor); err != nil {
		log.Fatalf("Failed to add HTTP task: %v", err)
	}
	
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	
	time.Sleep(2 * time.Minute)
	scheduler.Stop()
	
	// 查看执行结果
	results, _ := scheduler.GetTaskHistory("health-check-1", 5)
	fmt.Printf("HTTP任务执行了 %d 次\n", len(results))
}

// TaskManagementExample 任务管理示例
func TaskManagementExample() {
	fmt.Println("=== 任务管理示例 ===")
	
	store := storage.NewMemoryStorage()
	scheduler := timer.NewStandardScheduler(nil, store)
	
	logExecutor := executor.NewLogExecutor()
	
	// 添加多个任务
	tasks := []*timer.TaskConfig{
		{
			ID:           "task-1",
			Name:         "任务1",
			ScheduleType: timer.ScheduleTypeInterval,
			Interval:     30 * time.Second,
			Enabled:      true,
			Params:       "Task 1 message",
		},
		{
			ID:           "task-2",
			Name:         "任务2",
			ScheduleType: timer.ScheduleTypeInterval,
			Interval:     45 * time.Second,
			Enabled:      false, // 初始禁用
			Params:       "Task 2 message",
		},
	}
	
	for _, task := range tasks {
		if err := scheduler.AddTask(task, logExecutor); err != nil {
			log.Printf("Failed to add task %s: %v", task.ID, err)
		}
	}
	
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Failed to start scheduler: %v", err)
	}
	
	// 等待一段时间
	time.Sleep(time.Minute)
	
	// 启动任务2
	fmt.Println("启动任务2...")
	if err := scheduler.StartTask("task-2"); err != nil {
		log.Printf("Failed to start task-2: %v", err)
	}
	
	// 等待更长时间
	time.Sleep(2 * time.Minute)
	
	// 停止任务1
	fmt.Println("停止任务1...")
	if err := scheduler.StopTask("task-1"); err != nil {
		log.Printf("Failed to stop task-1: %v", err)
	}
	
	time.Sleep(time.Minute)
	
	// 列出所有任务
	allTasks, err := scheduler.ListTasks()
	if err != nil {
		log.Printf("Failed to list tasks: %v", err)
	} else {
		fmt.Printf("总共有 %d 个任务:\n", len(allTasks))
		for _, task := range allTasks {
			fmt.Printf("- %s: %s (状态: %s, 执行次数: %d)\n",
				task.Config.ID,
				task.Config.Name,
				task.Status.String(),
				task.RunCount)
		}
	}
	
	scheduler.Stop()
}

// RunAllExamples 运行所有示例
func RunAllExamples() {
	fmt.Println("开始运行定时任务组件示例...")
	
	// 注意：这些示例会运行较长时间，实际使用时请根据需要选择
	
	BasicUsageExample()
	fmt.Println()
	
	CronTaskExample()
	fmt.Println()
	
	CustomExecutorExample()
	fmt.Println()
	
	HTTPTaskExample()
	fmt.Println()
	
	TaskManagementExample()
	fmt.Println()
	
	fmt.Println("所有示例运行完成!")
} 