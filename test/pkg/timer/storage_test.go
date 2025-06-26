package timer

import (
	"fmt"
	"testing"
	"time"

	"gohub/pkg/timer"
)

// TestMemoryTaskStorage 测试内存存储的完整功能
// 验证MemoryTaskStorage实现TaskStorage接口的正确性
func TestMemoryTaskStorage(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 测试初始状态
	t.Run("初始状态验证", func(t *testing.T) {
		// 验证初始任务数量为0
		if storage.GetTaskCount() != 0 {
			t.Errorf("初始任务数量 = %d, want 0", storage.GetTaskCount())
		}

		// 验证列出空任务配置
		configs, err := storage.ListTaskConfigs()
		if err != nil {
			t.Errorf("列出空任务配置失败: %v", err)
		}
		if len(configs) != 0 {
			t.Errorf("空任务配置列表长度 = %d, want 0", len(configs))
		}
	})
}

// TestTaskConfigStorage 测试任务配置的存储操作
// 验证任务配置的保存、加载、删除和列表功能
func TestTaskConfigStorage(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 创建测试任务配置
	config1 := CreateTestTaskConfig("storage-001", "存储测试任务1", timer.ScheduleTypeOnce)
	config2 := CreateTestTaskConfig("storage-002", "存储测试任务2", timer.ScheduleTypeInterval)

	// 测试保存任务配置
	t.Run("保存任务配置", func(t *testing.T) {
		err := storage.SaveTaskConfig(config1)
		if err != nil {
			t.Errorf("保存任务配置失败: %v", err)
		}

		// 验证任务数量增加
		if storage.GetTaskCount() != 1 {
			t.Errorf("保存后任务数量 = %d, want 1", storage.GetTaskCount())
		}

		// 保存第二个任务配置
		err = storage.SaveTaskConfig(config2)
		if err != nil {
			t.Errorf("保存第二个任务配置失败: %v", err)
		}

		if storage.GetTaskCount() != 2 {
			t.Errorf("保存两个任务后数量 = %d, want 2", storage.GetTaskCount())
		}
	})

	// 测试保存nil配置
	t.Run("保存nil配置", func(t *testing.T) {
		err := storage.SaveTaskConfig(nil)
		if err == nil {
			t.Error("保存nil配置应该返回错误")
		}
	})

	// 测试保存空ID配置
	t.Run("保存空ID配置", func(t *testing.T) {
		invalidConfig := &timer.TaskConfig{
			ID:           "",
			Name:         "无效配置",
			ScheduleType: timer.ScheduleTypeOnce,
		}
		err := storage.SaveTaskConfig(invalidConfig)
		if err == nil {
			t.Error("保存空ID配置应该返回错误")
		}
	})

	// 测试加载任务配置
	t.Run("加载任务配置", func(t *testing.T) {
		loadedConfig, err := storage.LoadTaskConfig("storage-001")
		if err != nil {
			t.Errorf("加载任务配置失败: %v", err)
		}

		if loadedConfig == nil {
			t.Fatal("加载的配置为nil")
		}

		// 验证配置内容
		if loadedConfig.ID != config1.ID {
			t.Errorf("加载的配置ID = %s, want %s", loadedConfig.ID, config1.ID)
		}
		if loadedConfig.Name != config1.Name {
			t.Errorf("加载的配置名称 = %s, want %s", loadedConfig.Name, config1.Name)
		}
		if loadedConfig.ScheduleType != config1.ScheduleType {
			t.Errorf("加载的配置调度类型 = %v, want %v", loadedConfig.ScheduleType, config1.ScheduleType)
		}
	})

	// 测试加载不存在的任务配置
	t.Run("加载不存在的任务配置", func(t *testing.T) {
		_, err := storage.LoadTaskConfig("non-existent")
		if err == nil {
			t.Error("加载不存在的任务配置应该返回错误")
		}
	})

	// 测试列出所有任务配置
	t.Run("列出所有任务配置", func(t *testing.T) {
		configs, err := storage.ListTaskConfigs()
		if err != nil {
			t.Errorf("列出任务配置失败: %v", err)
		}

		if len(configs) != 2 {
			t.Errorf("配置列表长度 = %d, want 2", len(configs))
		}

		// 验证配置ID存在
		configIDs := make(map[string]bool)
		for _, config := range configs {
			configIDs[config.ID] = true
		}

		if !configIDs["storage-001"] {
			t.Error("配置列表中缺少storage-001")
		}
		if !configIDs["storage-002"] {
			t.Error("配置列表中缺少storage-002")
		}
	})

	// 测试删除任务配置
	t.Run("删除任务配置", func(t *testing.T) {
		err := storage.DeleteTaskConfig("storage-001")
		if err != nil {
			t.Errorf("删除任务配置失败: %v", err)
		}

		// 验证任务数量减少
		if storage.GetTaskCount() != 1 {
			t.Errorf("删除后任务数量 = %d, want 1", storage.GetTaskCount())
		}

		// 验证已删除的配置无法加载
		_, err = storage.LoadTaskConfig("storage-001")
		if err == nil {
			t.Error("删除后仍能加载配置，删除操作失败")
		}
	})

	// 测试删除不存在的任务配置
	t.Run("删除不存在的任务配置", func(t *testing.T) {
		err := storage.DeleteTaskConfig("non-existent")
		if err != nil {
			t.Errorf("删除不存在的任务配置不应该返回错误: %v", err)
		}
	})
}

// TestTaskInfoStorage 测试任务信息的存储操作
// 验证任务运行时信息的保存和加载功能
func TestTaskInfoStorage(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 创建测试任务配置和信息
	config := CreateTestTaskConfig("info-001", "信息存储测试", timer.ScheduleTypeInterval)
	now := time.Now()
	nextRun := now.Add(time.Hour)
	lastRun := now.Add(-time.Hour)

	taskInfo := &timer.TaskInfo{
		Config:       config,
		Status:       timer.TaskStatusRunning,
		NextRunTime:  &nextRun,
		LastRunTime:  &lastRun,
		RunCount:     5,
		FailureCount: 1,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// 测试保存任务信息
	t.Run("保存任务信息", func(t *testing.T) {
		err := storage.SaveTaskInfo(taskInfo)
		if err != nil {
			t.Errorf("保存任务信息失败: %v", err)
		}
	})

	// 测试保存nil任务信息
	t.Run("保存nil任务信息", func(t *testing.T) {
		err := storage.SaveTaskInfo(nil)
		if err == nil {
			t.Error("保存nil任务信息应该返回错误")
		}
	})

	// 测试保存无配置的任务信息
	t.Run("保存无配置的任务信息", func(t *testing.T) {
		invalidInfo := &timer.TaskInfo{
			Config: nil,
			Status: timer.TaskStatusPending,
		}
		err := storage.SaveTaskInfo(invalidInfo)
		if err == nil {
			t.Error("保存无配置的任务信息应该返回错误")
		}
	})

	// 测试保存空ID配置的任务信息
	t.Run("保存空ID配置的任务信息", func(t *testing.T) {
		invalidConfig := &timer.TaskConfig{ID: ""}
		invalidInfo := &timer.TaskInfo{
			Config: invalidConfig,
			Status: timer.TaskStatusPending,
		}
		err := storage.SaveTaskInfo(invalidInfo)
		if err == nil {
			t.Error("保存空ID配置的任务信息应该返回错误")
		}
	})

	// 测试加载任务信息
	t.Run("加载任务信息", func(t *testing.T) {
		loadedInfo, err := storage.LoadTaskInfo("info-001")
		if err != nil {
			t.Errorf("加载任务信息失败: %v", err)
		}

		if loadedInfo == nil {
			t.Fatal("加载的任务信息为nil")
		}

		// 验证任务信息内容
		if loadedInfo.Config.ID != taskInfo.Config.ID {
			t.Errorf("加载的任务配置ID = %s, want %s", loadedInfo.Config.ID, taskInfo.Config.ID)
		}
		if loadedInfo.Status != taskInfo.Status {
			t.Errorf("加载的任务状态 = %v, want %v", loadedInfo.Status, taskInfo.Status)
		}
		if loadedInfo.RunCount != taskInfo.RunCount {
			t.Errorf("加载的运行次数 = %d, want %d", loadedInfo.RunCount, taskInfo.RunCount)
		}
		if loadedInfo.FailureCount != taskInfo.FailureCount {
			t.Errorf("加载的失败次数 = %d, want %d", loadedInfo.FailureCount, taskInfo.FailureCount)
		}

		// 验证时间字段
		if loadedInfo.NextRunTime == nil || !loadedInfo.NextRunTime.Equal(*taskInfo.NextRunTime) {
			t.Errorf("加载的下次运行时间不匹配")
		}
		if loadedInfo.LastRunTime == nil || !loadedInfo.LastRunTime.Equal(*taskInfo.LastRunTime) {
			t.Errorf("加载的上次运行时间不匹配")
		}
	})

	// 测试加载不存在的任务信息
	t.Run("加载不存在的任务信息", func(t *testing.T) {
		_, err := storage.LoadTaskInfo("non-existent")
		if err == nil {
			t.Error("加载不存在的任务信息应该返回错误")
		}
	})

	// 测试更新任务信息
	t.Run("更新任务信息", func(t *testing.T) {
		// 修改任务信息
		taskInfo.Status = timer.TaskStatusCompleted
		taskInfo.RunCount = 10
		taskInfo.UpdatedAt = time.Now()

		err := storage.SaveTaskInfo(taskInfo)
		if err != nil {
			t.Errorf("更新任务信息失败: %v", err)
		}

		// 验证更新后的信息
		loadedInfo, err := storage.LoadTaskInfo("info-001")
		if err != nil {
			t.Errorf("加载更新后的任务信息失败: %v", err)
		}

		if loadedInfo.Status != timer.TaskStatusCompleted {
			t.Errorf("更新后的状态 = %v, want %v", loadedInfo.Status, timer.TaskStatusCompleted)
		}
		if loadedInfo.RunCount != 10 {
			t.Errorf("更新后的运行次数 = %d, want 10", loadedInfo.RunCount)
		}
	})
}

// TestTaskResultStorage 测试任务结果的存储操作
// 验证任务执行结果的保存和查询功能
func TestTaskResultStorage(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 创建测试任务结果
	taskID := "result-001"
	results := []*timer.TaskResult{
		{
			TaskID:     taskID,
			StartTime:  time.Now().Add(-time.Hour * 3),
			EndTime:    time.Now().Add(-time.Hour * 3).Add(time.Minute * 5),
			Duration:   time.Minute * 5,
			Status:     timer.TaskStatusCompleted,
			Error:      "",
			RetryCount: 0,
		},
		{
			TaskID:     taskID,
			StartTime:  time.Now().Add(-time.Hour * 2),
			EndTime:    time.Now().Add(-time.Hour * 2).Add(time.Minute * 3),
			Duration:   time.Minute * 3,
			Status:     timer.TaskStatusFailed,
			Error:      "模拟执行失败",
			RetryCount: 2,
		},
		{
			TaskID:     taskID,
			StartTime:  time.Now().Add(-time.Hour * 1),
			EndTime:    time.Now().Add(-time.Hour * 1).Add(time.Minute * 4),
			Duration:   time.Minute * 4,
			Status:     timer.TaskStatusCompleted,
			Error:      "",
			RetryCount: 1,
		},
	}

	// 测试保存任务结果
	t.Run("保存任务结果", func(t *testing.T) {
		for i, result := range results {
			err := storage.SaveTaskResult(result)
			if err != nil {
				t.Errorf("保存第%d个任务结果失败: %v", i+1, err)
			}

			// 验证结果数量
			if storage.GetResultCount(taskID) != i+1 {
				t.Errorf("保存第%d个结果后数量 = %d, want %d", i+1, storage.GetResultCount(taskID), i+1)
			}
		}
	})

	// 测试保存nil任务结果
	t.Run("保存nil任务结果", func(t *testing.T) {
		err := storage.SaveTaskResult(nil)
		if err == nil {
			t.Error("保存nil任务结果应该返回错误")
		}
	})

	// 测试保存空TaskID的任务结果
	t.Run("保存空TaskID的任务结果", func(t *testing.T) {
		invalidResult := &timer.TaskResult{
			TaskID:    "",
			StartTime: time.Now(),
			Status:    timer.TaskStatusCompleted,
		}
		err := storage.SaveTaskResult(invalidResult)
		if err == nil {
			t.Error("保存空TaskID的任务结果应该返回错误")
		}
	})

	// 测试加载任务结果（无限制）
	t.Run("加载任务结果无限制", func(t *testing.T) {
		loadedResults, err := storage.LoadTaskResults(taskID, 0)
		if err != nil {
			t.Errorf("加载任务结果失败: %v", err)
		}

		if len(loadedResults) != len(results) {
			t.Errorf("加载的结果数量 = %d, want %d", len(loadedResults), len(results))
		}

		// 验证结果按时间倒序排列（最新的在前面）
		for i := 1; i < len(loadedResults); i++ {
			if loadedResults[i-1].StartTime.Before(loadedResults[i].StartTime) {
				t.Error("结果应该按时间倒序排列")
			}
		}

		// 验证最新结果的内容
		latestResult := loadedResults[0]
		originalLatest := results[len(results)-1]
		if latestResult.TaskID != originalLatest.TaskID {
			t.Errorf("最新结果TaskID = %s, want %s", latestResult.TaskID, originalLatest.TaskID)
		}
		if latestResult.Status != originalLatest.Status {
			t.Errorf("最新结果状态 = %v, want %v", latestResult.Status, originalLatest.Status)
		}
	})

	// 测试加载任务结果（有限制）
	t.Run("加载任务结果有限制", func(t *testing.T) {
		limit := 2
		loadedResults, err := storage.LoadTaskResults(taskID, limit)
		if err != nil {
			t.Errorf("加载限制数量的任务结果失败: %v", err)
		}

		if len(loadedResults) != limit {
			t.Errorf("限制后的结果数量 = %d, want %d", len(loadedResults), limit)
		}

		// 验证返回的是最新的结果
		for i := 1; i < len(loadedResults); i++ {
			if loadedResults[i-1].StartTime.Before(loadedResults[i].StartTime) {
				t.Error("限制后的结果仍应该按时间倒序排列")
			}
		}
	})

	// 测试加载不存在任务的结果
	t.Run("加载不存在任务的结果", func(t *testing.T) {
		loadedResults, err := storage.LoadTaskResults("non-existent", 0)
		if err != nil {
			t.Errorf("加载不存在任务的结果失败: %v", err)
		}

		if len(loadedResults) != 0 {
			t.Errorf("不存在任务的结果数量 = %d, want 0", len(loadedResults))
		}
	})

	// 测试多任务结果存储
	t.Run("多任务结果存储", func(t *testing.T) {
		anotherTaskID := "result-002"
		anotherResult := &timer.TaskResult{
			TaskID:    anotherTaskID,
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Minute),
			Duration:  time.Minute,
			Status:    timer.TaskStatusCompleted,
			Error:     "",
		}

		err := storage.SaveTaskResult(anotherResult)
		if err != nil {
			t.Errorf("保存另一个任务的结果失败: %v", err)
		}

		// 验证不同任务的结果互不影响
		if storage.GetResultCount(taskID) != len(results) {
			t.Errorf("原任务结果数量 = %d, want %d", storage.GetResultCount(taskID), len(results))
		}
		if storage.GetResultCount(anotherTaskID) != 1 {
			t.Errorf("新任务结果数量 = %d, want 1", storage.GetResultCount(anotherTaskID))
		}
	})
}

// TestStorageClear 测试存储清理功能
// 验证存储数据的清空操作
func TestStorageClear(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 添加一些测试数据
	config := CreateTestTaskConfig("clear-001", "清理测试任务", timer.ScheduleTypeOnce)
	storage.SaveTaskConfig(config)

	taskInfo := &timer.TaskInfo{
		Config:    config,
		Status:    timer.TaskStatusPending,
		RunCount:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	storage.SaveTaskInfo(taskInfo)

	result := &timer.TaskResult{
		TaskID:    "clear-001",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Minute),
		Status:    timer.TaskStatusCompleted,
	}
	storage.SaveTaskResult(result)

	// 验证数据已存在
	if storage.GetTaskCount() == 0 {
		t.Error("清理前应该有任务数据")
	}
	if storage.GetResultCount("clear-001") == 0 {
		t.Error("清理前应该有结果数据")
	}

	// 执行清理
	t.Run("清理存储数据", func(t *testing.T) {
		storage.Clear()

		// 验证数据已清空
		if storage.GetTaskCount() != 0 {
			t.Errorf("清理后任务数量 = %d, want 0", storage.GetTaskCount())
		}
		if storage.GetResultCount("clear-001") != 0 {
			t.Errorf("清理后结果数量 = %d, want 0", storage.GetResultCount("clear-001"))
		}

		// 验证列表操作返回空结果
		configs, err := storage.ListTaskConfigs()
		if err != nil {
			t.Errorf("清理后列出配置失败: %v", err)
		}
		if len(configs) != 0 {
			t.Errorf("清理后配置列表长度 = %d, want 0", len(configs))
		}

		// 验证加载操作返回错误
		_, err = storage.LoadTaskInfo("clear-001")
		if err == nil {
			t.Error("清理后仍能加载任务信息")
		}

		results, err := storage.LoadTaskResults("clear-001", 0)
		if err != nil {
			t.Errorf("清理后加载结果失败: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("清理后结果列表长度 = %d, want 0", len(results))
		}
	})
}

// TestStorageConcurrency 测试存储的并发安全性
// 验证多线程环境下存储操作的正确性
func TestStorageConcurrency(t *testing.T) {
	storage := NewMemoryTaskStorage()

	// 并发保存任务配置
	t.Run("并发保存任务配置", func(t *testing.T) {
		const goroutineCount = 10
		const tasksPerGoroutine = 5

		done := make(chan bool, goroutineCount)

		for i := 0; i < goroutineCount; i++ {
			go func(goroutineID int) {
				defer func() { done <- true }()

				for j := 0; j < tasksPerGoroutine; j++ {
					taskID := fmt.Sprintf("concurrent-%d-%d", goroutineID, j)
					config := CreateTestTaskConfig(taskID, "并发测试任务", timer.ScheduleTypeOnce)

					err := storage.SaveTaskConfig(config)
					if err != nil {
						t.Errorf("并发保存任务配置失败: %v", err)
					}
				}
			}(i)
		}

		// 等待所有goroutine完成
		for i := 0; i < goroutineCount; i++ {
			<-done
		}

		// 验证所有任务都已保存
		expectedCount := goroutineCount * tasksPerGoroutine
		if storage.GetTaskCount() != expectedCount {
			t.Errorf("并发保存后任务数量 = %d, want %d", storage.GetTaskCount(), expectedCount)
		}
	})

	// 清理数据准备下一个测试
	storage.Clear()

	// 并发保存和加载
	t.Run("并发保存和加载", func(t *testing.T) {
		const operationCount = 20
		done := make(chan bool, operationCount*2)

		// 启动保存操作的goroutine
		for i := 0; i < operationCount; i++ {
			go func(id int) {
				defer func() { done <- true }()

				taskID := fmt.Sprintf("save-load-%d", id)
				config := CreateTestTaskConfig(taskID, "保存加载测试", timer.ScheduleTypeInterval)

				err := storage.SaveTaskConfig(config)
				if err != nil {
					t.Errorf("并发保存失败: %v", err)
				}
			}(i)
		}

		// 启动加载操作的goroutine
		for i := 0; i < operationCount; i++ {
			go func(id int) {
				defer func() { done <- true }()

				taskID := fmt.Sprintf("save-load-%d", id)
				// 尝试加载，可能成功也可能失败（取决于保存操作是否完成）
				_, err := storage.LoadTaskConfig(taskID)
				// 这里不检查错误，因为加载可能发生在保存之前
				_ = err
			}(i)
		}

		// 等待所有操作完成
		for i := 0; i < operationCount*2; i++ {
			<-done
		}

		// 验证最终状态
		if storage.GetTaskCount() != operationCount {
			t.Errorf("并发操作后任务数量 = %d, want %d", storage.GetTaskCount(), operationCount)
		}
	})
}

// BenchmarkTaskConfigSave 基准测试任务配置保存性能
func BenchmarkTaskConfigSave(b *testing.B) {
	storage := NewMemoryTaskStorage()
	config := CreateTestTaskConfig("bench-config", "基准测试配置", timer.ScheduleTypeOnce)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.ID = fmt.Sprintf("bench-config-%d", i)
		storage.SaveTaskConfig(config)
	}
}

// BenchmarkTaskConfigLoad 基准测试任务配置加载性能
func BenchmarkTaskConfigLoad(b *testing.B) {
	storage := NewMemoryTaskStorage()

	// 预先保存一些配置
	for i := 0; i < 1000; i++ {
		config := CreateTestTaskConfig(fmt.Sprintf("bench-load-%d", i), "基准加载测试", timer.ScheduleTypeOnce)
		storage.SaveTaskConfig(config)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		taskID := fmt.Sprintf("bench-load-%d", i%1000)
		storage.LoadTaskConfig(taskID)
	}
}

// BenchmarkTaskResultSave 基准测试任务结果保存性能
func BenchmarkTaskResultSave(b *testing.B) {
	storage := NewMemoryTaskStorage()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := &timer.TaskResult{
			TaskID:    fmt.Sprintf("bench-result-%d", i%100), // 使用100个不同的任务ID
			StartTime: time.Now(),
			EndTime:   time.Now().Add(time.Second),
			Duration:  time.Second,
			Status:    timer.TaskStatusCompleted,
		}
		storage.SaveTaskResult(result)
	}
} 