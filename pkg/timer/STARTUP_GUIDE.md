# Timer组件启动流程指南

本指南详细介绍了如何正确启动和使用Timer定时任务调度组件。

## 启动流程（9个步骤）

### 1. 导入必要的包

```go
import (
    "context"
    "time"
    
    "gohub/pkg/timer"
    "gohub/pkg/timer/executor"
    "gohub/pkg/timer/storage"
)
```

### 2. 创建存储后端

```go
// 使用内存存储（开发/测试环境）
store := storage.NewMemoryStorage()

// 可选：设置最大结果保存数量
store.SetMaxResults(200)
```

### 3. 创建调度器实例

```go
// 使用默认配置
scheduler := timer.NewStandardScheduler(nil, store)

// 或使用自定义配置
config := &timer.SchedulerConfig{
    Name:           "MyScheduler",
    MaxWorkers:     10,
    QueueSize:      200,
    DefaultTimeout: time.Minute * 15,
    DefaultRetries: 2,
}
scheduler := timer.NewStandardScheduler(config, store)
```

### 4. 创建任务执行器

```go
// 内置执行器
logExecutor := executor.NewLogExecutor()
httpExecutor := executor.NewHTTPExecutor()
cmdExecutor := executor.NewCommandExecutor()

// 自定义函数执行器
customExecutor := executor.NewFunctionExecutor("MyTask", func(ctx context.Context, params interface{}) error {
    // 实现具体的任务逻辑
    return nil
})
```

### 5. 配置任务

```go
// 间隔调度任务
intervalTask := &timer.TaskConfig{
    ID:           "interval-task-1",
    Name:         "间隔任务",
    Description:  "每30秒执行一次",
    Priority:     timer.TaskPriorityNormal,
    ScheduleType: timer.ScheduleTypeInterval,
    Interval:     30 * time.Second,
    Enabled:      true,
    MaxRetries:   3,
    Timeout:      time.Minute * 5,
    Params:       "Hello from interval task",
}

// Cron调度任务
cronTask := &timer.TaskConfig{
    ID:           "cron-task-1",
    Name:         "Cron任务",
    Description:  "工作日上午9点执行",
    ScheduleType: timer.ScheduleTypeCron,
    CronExpr:     "0 9 * * 1-5", // 工作日上午9点
    Enabled:      true,
    Params: map[string]interface{}{
        "action": "daily_report",
        "email":  "admin@example.com",
    },
}
```

### 6. 添加任务到调度器

```go
// 添加间隔任务
if err := scheduler.AddTask(intervalTask, logExecutor); err != nil {
    log.Fatalf("Failed to add interval task: %v", err)
}

// 添加Cron任务
if err := scheduler.AddTask(cronTask, customExecutor); err != nil {
    log.Fatalf("Failed to add cron task: %v", err)
}
```

### 7. 启动调度器

```go
// 启动调度器（这是关键步骤）
if err := scheduler.Start(); err != nil {
    log.Fatalf("Failed to start scheduler: %v", err)
}

log.Println("调度器启动成功")
```

### 8. 运行时管理（可选）

```go
// 查看所有任务
tasks, err := scheduler.ListTasks()
if err == nil {
    for _, task := range tasks {
        log.Printf("Task: %s, Status: %s", task.Config.Name, task.Status.String())
    }
}

// 手动触发任务
if err := scheduler.TriggerTask("interval-task-1", "手动触发"); err != nil {
    log.Printf("Failed to trigger task: %v", err)
}

// 暂停和恢复任务
scheduler.StopTask("interval-task-1")   // 暂停任务
scheduler.StartTask("interval-task-1")  // 恢复任务

// 查看任务执行历史
results, err := scheduler.GetTaskHistory("interval-task-1", 10)
if err == nil {
    log.Printf("Task executed %d times", len(results))
}
```

### 9. 优雅关闭

```go
// 程序退出前停止调度器
if err := scheduler.Stop(); err != nil {
    log.Printf("Failed to stop scheduler: %v", err)
} else {
    log.Println("调度器已停止")
}
```

## 完整示例代码

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "gohub/pkg/timer"
    "gohub/pkg/timer/executor"
    "gohub/pkg/timer/storage"
)

func main() {
    // 1. 创建存储
    store := storage.NewMemoryStorage()
    
    // 2. 创建调度器
    scheduler := timer.NewStandardScheduler(nil, store)
    
    // 3. 创建执行器
    logExecutor := executor.NewLogExecutor()
    
    // 4. 配置任务
    taskConfig := &timer.TaskConfig{
        ID:           "demo-task",
        Name:         "演示任务",
        Description:  "每分钟执行的演示任务",
        ScheduleType: timer.ScheduleTypeInterval,
        Interval:     time.Minute,
        Enabled:      true,
        Params:       "Hello from Timer!",
    }
    
    // 5. 添加任务
    if err := scheduler.AddTask(taskConfig, logExecutor); err != nil {
        log.Fatalf("添加任务失败: %v", err)
    }
    
    // 6. 启动调度器
    if err := scheduler.Start(); err != nil {
        log.Fatalf("启动调度器失败: %v", err)
    }
    log.Println("调度器启动成功")
    
    // 7. 等待中断信号
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    // 8. 优雅关闭
    log.Println("正在关闭调度器...")
    if err := scheduler.Stop(); err != nil {
        log.Printf("关闭调度器失败: %v", err)
    } else {
        log.Println("调度器已安全关闭")
    }
}
```

## 重要注意事项

### 启动顺序
1. **存储 → 调度器 → 执行器 → 任务配置 → 添加任务 → 启动调度器**
2. 必须在添加完所有任务后再调用 `Start()`
3. 调度器启动后仍可以动态添加/移除任务

### 错误处理
- 每个步骤都应该检查错误并适当处理
- 任务添加失败不会影响已添加的任务
- 调度器启动失败时应该检查配置和存储连接

### 性能考虑
- `MaxWorkers` 控制并发执行的任务数量，设置过大可能导致资源耗尽
- `QueueSize` 控制任务队列大小，设置过小可能导致任务丢失
- 长时间运行的任务应该设置合理的 `Timeout`

### 生产环境建议
1. **使用持久化存储**: 生产环境建议实现 Redis 或数据库存储后端
2. **监控和日志**: 集成应用监控系统，关注任务执行状态
3. **优雅关闭**: 确保程序退出时正确停止调度器
4. **资源限制**: 根据系统资源合理设置工作线程数和队列大小
5. **错误处理**: 实现完善的错误处理和重试机制

## 故障排除

### 常见问题

**Q: 任务没有执行**
- 检查任务是否已启用 (`Enabled: true`)
- 检查调度器是否已启动 (`scheduler.Start()`)
- 检查任务配置是否有效 (`ValidateTaskConfig`)

**Q: Cron任务时间不准确**
- Cron表达式使用UTC时间，注意时区转换
- 检查Cron表达式语法是否正确

**Q: 任务执行失败**
- 查看任务执行历史 (`GetTaskHistory`)
- 检查执行器实现是否正确
- 确认任务参数格式是否匹配

**Q: 内存使用过高**
- 减少 `MaxResults` 设置
- 检查任务是否存在内存泄漏
- 适当调整工作线程数

### 调试技巧
1. 使用 `LogExecutor` 测试任务调度是否正常
2. 通过 `GetRunningTasks()` 查看当前运行状态
3. 使用 `TriggerTask()` 手动测试任务执行
4. 查看任务执行历史分析问题模式

## 最佳实践

1. **任务幂等性**: 确保任务可以安全地重复执行
2. **错误恢复**: 实现适当的错误处理和重试逻辑
3. **资源清理**: 任务执行完成后及时清理资源
4. **配置管理**: 将任务配置外部化，便于管理和修改
5. **监控告警**: 对关键任务设置监控和告警机制 