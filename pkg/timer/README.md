# Timer - 定时任务调度组件

一个功能完整、结构清晰的Go定时任务调度组件，支持多种调度策略和横向扩展。

## 目录结构

```
pkg/timer/
├── timer.go              # 核心接口和类型定义
├── task.go              # 任务相关工具函数
├── scheduler.go         # 标准调度器实现
├── cron/               # Cron表达式解析
│   └── cron.go
├── executor/           # 任务执行器
│   ├── executor.go     # 执行器接口和基础实现
│   └── builtin.go      # 内置执行器
├── storage/            # 存储实现
│   ├── storage.go      # 存储接口
│   └── memory.go       # 内存存储实现
└── examples/           # 使用示例
    └── basic.go
```

## 核心特性

- **多种调度策略**: 支持一次性执行、固定间隔、Cron表达式、延迟执行
- **完整的任务管理**: 任务的增删改查、启动停止、手动触发
- **内置执行器**: 日志输出、HTTP请求、系统命令等常用执行器
- **自定义执行器**: 支持函数执行器，可以轻松扩展
- **并发安全**: 使用工作线程池，支持并发执行
- **错误处理**: 支持重试机制和超时控制
- **存储持久化**: 内存存储，易于扩展Redis等其他存储

## 快速开始

> 📖 **详细启动指南**: 请查看 [STARTUP_GUIDE.md](./STARTUP_GUIDE.md) 获取完整的启动流程说明

### 1. 基本使用

```go
package main

import (
    "time"
    "github.com/gohub/pkg/timer"
    "github.com/gohub/pkg/timer/executor"
    "github.com/gohub/pkg/timer/storage"
)

func main() {
    // 创建存储和调度器
    store := storage.NewMemoryStorage()
    scheduler := timer.NewStandardScheduler(nil, store)
    
    // 创建任务配置
    taskConfig := &timer.TaskConfig{
        ID:           "my-task",
        Name:         "示例任务",
        ScheduleType: timer.ScheduleTypeInterval,
        Interval:     time.Minute,
        Enabled:      true,
        Params:       "Hello World!",
    }
    
    // 创建执行器
    executor := executor.NewLogExecutor()
    
    // 添加任务并启动调度器
    scheduler.AddTask(taskConfig, executor)
    scheduler.Start()
    
    // 等待任务执行
    time.Sleep(5 * time.Minute)
    
    // 停止调度器
    scheduler.Stop()
}
```

### 2. Cron表达式任务

```go
taskConfig := &timer.TaskConfig{
    ID:           "cron-task",
    Name:         "Cron任务",
    ScheduleType: timer.ScheduleTypeCron,
    CronExpr:     "0 9 * * 1-5", // 工作日上午9点
    Enabled:      true,
}
```

### 3. 自定义执行器

```go
customExecutor := executor.NewFunctionExecutor("MyTask", func(ctx context.Context, params interface{}) error {
    fmt.Printf("执行自定义任务: %v\n", params)
    return nil
})
```

### 4. HTTP健康检查任务

```go
httpExecutor := executor.NewHTTPExecutor()
taskConfig := &timer.TaskConfig{
    ID:           "health-check",
    Name:         "健康检查",
    ScheduleType: timer.ScheduleTypeInterval,
    Interval:     30 * time.Second,
    Enabled:      true,
    Params: map[string]interface{}{
        "url":    "https://example.com/health",
        "method": "GET",
    },
}
```

## API参考

### 调度器接口

- `AddTask(config, executor)` - 添加任务
- `RemoveTask(taskID)` - 移除任务
- `StartTask(taskID)` - 启动任务
- `StopTask(taskID)` - 停止任务
- `TriggerTask(taskID, params)` - 手动触发任务
- `Start()` - 启动调度器
- `Stop()` - 停止调度器
- `ListTasks()` - 列出所有任务
- `GetTaskHistory(taskID, limit)` - 获取任务执行历史

### 内置执行器

- `LogExecutor` - 日志输出执行器
- `HTTPExecutor` - HTTP请求执行器
- `CommandExecutor` - 系统命令执行器
- `FunctionExecutor` - 自定义函数执行器

### Cron表达式

支持标准5字段Cron表达式：`分钟 小时 日 月 周`

预定义表达式：
- `cron.EveryMinute` - 每分钟
- `cron.Hourly` - 每小时
- `cron.Daily` - 每天
- `cron.Weekly` - 每周
- `cron.Monthly` - 每月
- `cron.Yearly` - 每年

## 运行示例

```go
import "github.com/gohub/pkg/timer/examples"

// 运行基本示例
examples.BasicUsageExample()

// 运行所有示例
examples.RunAllExamples()
```

## 扩展

### 自定义存储

实现 `timer.TaskStorage` 接口即可添加新的存储后端（如Redis、MySQL等）。

### 自定义执行器

实现 `timer.TaskExecutor` 接口即可创建自定义执行器：

```go
type MyExecutor struct {
    name string
}

func (e *MyExecutor) Execute(ctx context.Context, params interface{}) error {
    // 执行逻辑
    return nil
}

func (e *MyExecutor) GetName() string {
    return e.name
}
```

## 最佳实践

1. **任务ID唯一性**: 确保任务ID在整个系统中唯一
2. **超时设置**: 为长时间运行的任务设置合适的超时时间
3. **错误处理**: 合理设置重试次数和重试间隔
4. **资源管理**: 根据系统负载调整工作线程数和队列大小
5. **监控日志**: 关注任务执行日志，及时发现问题

## 架构设计

### 核心组件关系

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   TaskConfig    │    │   TaskExecutor  │    │   TaskStorage   │
│   (任务配置)    │    │   (任务执行器)  │    │   (存储接口)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │ TaskScheduler   │
                    │ (任务调度器)    │
                    └─────────────────┘
                                 │
                    ┌─────────────────┐
                    │  工作线程池     │
                    │  (Worker Pool)  │
                    └─────────────────┘
```

### 组件职责

- **TaskConfig**: 定义任务的调度规则、重试策略、超时设置等配置信息
- **TaskExecutor**: 实现具体的任务执行逻辑，支持多种内置执行器和自定义扩展
- **TaskStorage**: 提供任务数据的持久化存储，支持内存存储和扩展存储后端
- **TaskScheduler**: 核心调度器，负责任务的生命周期管理和调度执行
- **Worker Pool**: 工作线程池，负责并发执行任务，控制系统资源使用

## 执行流程

### 1. 任务添加流程

```
用户调用 AddTask
    ↓
验证任务配置 (ValidateTaskConfig)
    ↓
检查任务ID唯一性
    ↓
保存任务配置到存储 (SaveTaskConfig)
    ↓
创建任务运行时信息 (NewTaskInfo)
    ↓
注册执行器到内存映射
    ↓
如果调度器运行中且任务已启用
    ↓
开始任务调度 (scheduleTask)
```

### 2. 任务调度流程

```
调度器启动 (Start)
    ↓
启动工作线程池
    ↓
加载所有任务配置
    ↓
为每个启用任务计算下次执行时间
    ↓
设置定时器 (time.Timer)
    ↓
定时器触发
    ↓
创建任务作业 (taskJob)
    ↓
提交到任务队列 (taskQueue)
    ↓
工作线程获取任务
    ↓
执行任务 (runTask)
    ↓
记录执行结果
    ↓
重新计算下次执行时间
    ↓
设置新的定时器
```

### 3. 任务执行流程

```
工作线程从队列获取任务
    ↓
创建执行结果对象 (NewTaskResult)
    ↓
更新任务状态为运行中
    ↓
设置执行超时控制
    ↓
调用执行器 Execute 方法
    ↓
处理执行结果：
    ├─ 成功：标记完成 (Complete)
    └─ 失败：
        ├─ 检查重试次数
        ├─ 如果可重试：延迟后重新执行
        └─ 否则：标记失败 (Fail)
    ↓
保存执行结果到存储
    ↓
更新任务统计信息
```

### 4. 调度策略

#### Cron表达式调度
- 使用标准5字段Cron表达式（分钟 小时 日 月 周）
- 支持通配符、范围、列表、步长等语法
- 计算下次执行时间基于当前时间和Cron规则

#### 固定间隔调度
- 基于配置的时间间隔重复执行
- 从任务完成时间开始计算下次执行时间
- 确保任务之间有固定的间隔时间

#### 一次性执行
- 在指定时间执行一次
- 执行完成后不再调度

#### 延迟执行
- 延迟指定时间后执行
- 可以结合其他调度策略使用

### 5. 错误处理机制

```
任务执行失败
    ↓
检查重试配置 (MaxRetries)
    ↓
如果重试次数未达上限：
    ├─ 等待重试间隔 (RetryInterval)
    ├─ 增加重试计数
    └─ 重新执行任务
    ↓
如果重试次数达上限：
    ├─ 标记任务失败
    ├─ 记录错误信息
    └─ 继续下次正常调度
```

### 6. 并发控制

- **工作线程池**: 限制同时执行的任务数量，防止系统资源耗尽
- **任务队列**: 缓冲待执行任务，平滑处理任务峰值
- **读写锁**: 保护调度器状态和任务映射的并发安全
- **上下文控制**: 支持任务超时和优雅取消

## 扩展机制

### 自定义执行器

```go
type MyExecutor struct {
    name string
}

func (e *MyExecutor) Execute(ctx context.Context, params interface{}) error {
    // 实现自定义执行逻辑
    return nil
}

func (e *MyExecutor) GetName() string {
    return e.name
}
```

### 自定义存储后端

```go
type MyStorage struct {
    // 存储实现
}

func (s *MyStorage) SaveTaskConfig(config *TaskConfig) error {
    // 实现配置保存逻辑
    return nil
}

// 实现 TaskStorage 接口的其他方法...
```

## 性能特性

- **内存高效**: 使用对象池和延迟加载减少内存分配
- **并发安全**: 采用读写锁和无锁数据结构优化并发性能
- **调度精确**: 基于系统定时器实现精确的任务调度
- **资源控制**: 通过工作线程池和队列大小控制系统资源使用

## 注意事项

- 调度器停止后，正在执行的任务会等待完成
- Cron表达式使用UTC时间，请注意时区问题
- 内存存储重启后数据会丢失，生产环境建议使用持久化存储
- 任务执行时间过长可能影响其他任务的调度精度
- 建议合理设置工作线程数和队列大小，避免系统资源耗尽 