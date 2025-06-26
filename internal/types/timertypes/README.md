# 定时任务类型定义 (TimerTypes)

本目录包含了定时任务模块的所有数据库类型定义，对应优化后的数据库表结构。

## 📁 文件结构

```
timertypes/
├── README.md              # 本说明文件
├── types.go               # 类型汇总文件，统一导出
├── constants.go           # 常量定义
├── scheduler.go           # 调度器类型
├── task.go                # 任务类型
└── execution_log.go       # 执行日志类型
```

## 🔧 优化内容

### 1. 表结构简化
- **合并任务相关表**：将原有的任务配置、任务信息和任务结果表合并为一个任务表
- **执行日志整合**：将执行记录和日志信息整合到一个执行日志表中
- **减少关联查询**：通过冗余关键字段减少多表关联查询

### 2. 建立调度器关联
- **明确关联关系**：任务通过`schedulerId`字段与调度器建立关联
- **冗余关键字段**：在任务表中冗余`schedulerName`字段，方便查询显示

### 3. 增强功能支持
- **支持多种调度类型**：一次性执行、固定间隔、Cron表达式、延迟执行和实时执行
- **完整的执行状态跟踪**：记录任务的完整生命周期和执行历史

## 📊 表关系结构

```
HUB_TIMER_SCHEDULER (调度器)
    ↓ (1:N)
HUB_TIMER_TASK (任务)
    ↓ (1:N)
HUB_TIMER_EXECUTION_LOG (执行日志)
```

## 🚀 使用方法

### 基本导入

```go
import (
    "your-project/internal/types/timertypes"
)

// 使用类型
var scheduler timertypes.TimerScheduler
var task timertypes.TimerTask
var log timertypes.TimerExecutionLog

// 使用常量
status := timertypes.TimerTaskStatusRunning
priority := timertypes.TimerTaskPriorityHigh
```

### 创建调度器

```go
scheduler := &timertypes.TimerScheduler{
    SchedulerId:        "SCHEDULER_001",
    TenantId:           "TENANT_001",
    SchedulerName:      "主调度器",
    SchedulerInstanceId: "INSTANCE_001",
    MaxWorkers:         10,
    QueueSize:          100,
    SchedulerStatus:    timertypes.TimerSchedulerStatusStopped,
    ActiveFlag:         timertypes.TimerActiveYes,
}

// 启动调度器
if scheduler.CanStart() {
    scheduler.Start()
    fmt.Printf("调度器已启动，状态: %s", scheduler.GetStatusName())
}

// 更新心跳
scheduler.UpdateHeartbeat()
```

### 创建任务

```go
task := &timertypes.TimerTask{
    TaskId:            "TASK_001",
    TenantId:          "TENANT_001",
    TaskName:          "数据同步任务",
    TaskDescription:   timertypes.StringPtr("每日数据同步"),
    TaskPriority:      timertypes.TimerTaskPriorityHigh,
    SchedulerId:       timertypes.StringPtr("SCHEDULER_001"),
    SchedulerName:     timertypes.StringPtr("主调度器"),
    ScheduleType:      timertypes.TimerScheduleTypeCron,
    CronExpression:    timertypes.StringPtr("0 2 * * *"), // 每天凌晨2点
    MaxRetries:        3,
    TimeoutSeconds:    1800,
    ActiveFlag:        timertypes.TimerActiveYes,
}

// 验证配置
if err := task.Validate(); err != nil {
    log.Printf("任务配置验证失败: %v", err)
}

// 计算下次执行时间
if err := task.CalculateNextRunTime(); err != nil {
    log.Printf("计算下次执行时间失败: %v", err)
}
```

### 任务执行

```go
// 开始执行任务
task.StartExecution()

// 执行成功
task.CompleteExecution(true, nil)

// 或执行失败
errorMsg := "数据库连接失败"
task.CompleteExecution(false, &errorMsg)
```

### 创建执行日志

```go
// 创建执行日志
executionLog := &timertypes.TimerExecutionLog{
    ExecutionId:       "EXEC_001",
    TenantId:          "TENANT_001",
    TaskId:            "TASK_001",
    TaskName:          timertypes.StringPtr("数据同步任务"),
    ExecutionStartTime: time.Now(),
    ExecutionStatus:   timertypes.TimerExecutionStatusPending,
    ResultSuccess:     timertypes.TimerResultFailure, // 默认为失败，执行成功后更新
    ActiveFlag:        timertypes.TimerActiveYes,
}

// 开始执行
executionLog.StartExecution()

// 执行成功
result := "同步了1000条数据"
executionLog.CompleteExecution(true, &result)

// 或执行失败
executionLog.FailExecution("数据库连接失败", nil, nil)

// 添加日志
executionLog.AddLog(timertypes.TimerLogLevelInfo, "开始执行数据同步")
```

### 使用辅助函数

```go
// 创建信息日志
infoLog := timertypes.CreateInfoLog(
    "TENANT_001",
    "TASK_001", 
    "数据同步任务",
    "开始执行数据同步"
)

// 创建错误日志
errorLog := timertypes.CreateErrorLog(
    "TENANT_001",
    "TASK_001",
    "数据同步任务", 
    "数据同步失败",
    timertypes.StringPtr("DatabaseException"),
    timertypes.StringPtr("连接超时")
)
```

## 📋 字段说明

### 主要关联字段
- `schedulerId`: 关联调度器ID
- `taskId`: 任务ID  
- `executionId`: 执行日志ID

### 冗余字段（便于查询）
- `schedulerName`: 调度器名称
- `taskName`: 任务名称

### 状态字段
- `activeFlag`: 活动状态标记 (Y/N)
- `schedulerStatus`: 调度器状态 (1-3)
- `taskStatus`: 任务状态 (1-5)
- `executionStatus`: 执行状态 (1-5)
- `resultSuccess`: 执行结果 (Y/N)

## 🔗 相关文档

- [数据库表结构设计](../../../docs/database/mysql.sql)
- [定时任务模块文档](../../../pkg/timer/README.md)
- [数据库设计文档](../../../docs/database/timer_database_design.md)

## 📝 注意事项

1. **外键约束**: 生产环境中建议启用外键约束确保数据完整性
2. **索引优化**: 根据查询模式添加适当的索引
3. **数据清理**: 定期清理历史日志和结果数据
4. **并发安全**: 在并发环境中注意数据一致性
5. **性能监控**: 监控大表的查询性能

---

> 💡 如有问题或建议，请联系开发团队或提交Issue。 