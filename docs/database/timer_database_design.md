# 定时任务模块数据库设计文档

## 1. 概述

本文档描述了定时任务模块（Timer Module）的数据库设计，严格遵循项目的数据库设计规范。定时任务模块支持多种调度策略，包括Cron表达式、固定间隔、延迟执行和一次性执行。

## 2. 设计原则

### 2.1 命名规范
- 所有表名以 `HUB_TIMER` 开头，符合项目模块前缀规范
- 字段命名使用驼峰命名法，语义明确
- 主键字段根据表名具体命名，如 `taskConfigId`、`taskInfoId` 等
- 严格禁止使用自增字段，主键使用VARCHAR(32)类型存储UUID
- **主键设计**：所有表使用复合主键 `(tenantId, tablePrimaryKey)`，确保多租户数据隔离

### 2.2 数据类型规范
- 字符串统一使用VARCHAR类型
- 数值类型统一使用INT类型
- 时间类型统一使用DATETIME
- 状态标记使用VARCHAR(1)，值为'Y'/'N'
- 大文本内容使用TEXT类型

### 2.3 通用字段
所有表都包含标准的通用字段：
- `tenantId`：租户ID，用于多租户数据隔离
- `addTime`、`addWho`、`editTime`、`editWho`：审计字段
- `oprSeqFlag`、`currentVersion`、`activeFlag`、`noteText`：版本控制和状态管理
- `reserved1` 到 `reserved10`：预留字段，用于业务扩展

## 3. 表结构设计

### 3.1 任务配置表 (HUB_TIMER_TASK_CONFIG)

**用途**: 存储定时任务的基本配置信息，包括调度规则、执行参数等。

**核心字段说明**:
- `taskConfigId`: 主键，任务配置的唯一标识
- `taskId`: 业务ID，任务的业务唯一标识，用于外部引用
- `taskName`: 任务名称，便于管理和识别
- `taskPriority`: 任务优先级 (1=低优先级, 2=普通优先级, 3=高优先级)
- `scheduleType`: 调度类型 (1=一次性, 2=固定间隔, 3=Cron表达式, 4=延迟执行)
- `cronExpression`: Cron表达式，当scheduleType=3时必填
- `intervalSeconds`: 执行间隔秒数，当scheduleType=2时必填
- `delaySeconds`: 延迟秒数，当scheduleType=4时必填
- `maxRetries`: 最大重试次数
- `timeoutSeconds`: 执行超时时间（秒）
- `taskParams`: 任务参数，JSON格式存储
- `enabledFlag`: 启用状态标记

**业务规则**:
1. `taskId` 必须全局唯一
2. 根据 `scheduleType` 的不同，相应的调度参数字段必须有值
3. `startTime` 和 `endTime` 定义任务的有效期
4. `enabledFlag='Y'` 的任务才会被调度执行

### 3.2 任务运行时信息表 (HUB_TIMER_TASK_INFO)

**用途**: 存储任务的运行状态和统计信息，记录任务的实时状态。

**核心字段说明**:
- `taskInfoId`: 主键，任务信息的唯一标识
- `taskConfigId`: 关联任务配置表的外键
- `taskStatus`: 任务状态 (1=待执行, 2=运行中, 3=已完成, 4=执行失败, 5=已取消)
- `nextRunTime`: 下次执行时间，调度器根据此字段安排任务执行
- `lastRunTime`: 上次执行时间
- `runCount`: 总执行次数
- `successCount`: 成功执行次数
- `failureCount`: 失败执行次数
- `lastResultId`: 最后一次执行结果的ID
- `lastExecutionDurationMs`: 最后一次执行耗时（毫秒）

**业务规则**:
1. 每个任务配置对应一条任务信息记录
2. `taskStatus` 反映任务的当前状态
3. 统计字段用于监控和分析任务执行情况
4. `nextRunTime` 用于调度器计算下次执行时间

### 3.3 任务执行结果表 (HUB_TIMER_TASK_RESULT)

**用途**: 记录每次任务执行的详细结果，用于审计和问题排查。

**核心字段说明**:
- `taskResultId`: 主键，执行结果的唯一标识
- `taskId`: 任务业务ID
- `taskConfigId`: 关联任务配置表
- `executionStartTime`: 执行开始时间
- `executionEndTime`: 执行结束时间
- `executionDurationMs`: 执行耗时（毫秒）
- `executionStatus`: 执行状态
- `resultSuccess`: 执行是否成功 (Y=成功, N=失败)
- `errorMessage`: 错误信息
- `retryCount`: 当次执行的重试次数
- `executionParams`: 执行时使用的参数（JSON格式）
- `executionResult`: 执行结果（JSON格式）
- `executorServerName`: 执行服务器名称
- `executorServerIp`: 执行服务器IP

**业务规则**:
1. 每次任务执行都会产生一条执行结果记录
2. 支持分布式环境下的服务器信息记录
3. 错误信息和堆栈跟踪便于问题诊断
4. 执行参数和结果的JSON存储提供灵活性

### 3.4 调度器配置表 (HUB_TIMER_SCHEDULER_CONFIG)

**用途**: 存储调度器实例的配置信息，支持多实例部署和集群管理。

**核心字段说明**:
- `schedulerConfigId`: 主键，调度器配置的唯一标识
- `schedulerName`: 调度器名称
- `schedulerInstanceId`: 调度器实例ID，用于集群环境区分
- `maxWorkers`: 最大工作线程数
- `queueSize`: 任务队列大小
- `defaultTimeoutSeconds`: 默认超时时间
- `schedulerStatus`: 调度器状态 (1=停止, 2=运行中, 3=暂停)
- `totalTaskCount`: 总任务数
- `runningTaskCount`: 运行中任务数
- `lastHeartbeatTime`: 最后心跳时间

**业务规则**:
1. 支持多个调度器实例同时运行
2. `schedulerInstanceId` 在集群环境中必须唯一
3. 心跳机制用于监控调度器健康状态
4. 运行时统计信息便于监控和调优

### 3.5 任务执行日志表 (HUB_TIMER_TASK_LOG)

**用途**: 记录任务执行过程中的详细日志信息，用于调试和监控。

**核心字段说明**:
- `taskLogId`: 主键，日志记录的唯一标识
- `taskId`: 任务业务ID
- `taskResultId`: 关联任务执行结果
- `logLevel`: 日志级别 (DEBUG, INFO, WARN, ERROR)
- `logMessage`: 日志消息内容
- `logTimestamp`: 日志时间戳
- `executionPhase`: 执行阶段标识
- `threadName`: 执行线程名称
- `exceptionClass`: 异常类名
- `exceptionMessage`: 异常消息

**业务规则**:
1. 按执行阶段记录详细的执行日志
2. 支持不同日志级别的分类存储
3. 异常信息的详细记录便于问题诊断
4. 与任务执行结果关联，形成完整的执行链路

## 4. 索引设计

### 4.1 主键索引
所有表都有复合主键索引 `(tenantId, tablePrimaryKey)`，确保记录的唯一性和多租户数据隔离。

### 4.2 普通索引
基于常用查询条件创建索引：
- 业务字段索引：如 `taskId`、`schedulerName` 等，便于按业务标识查询
- 状态字段索引：便于按状态查询
- 时间字段索引：便于按时间范围查询
- 关联字段索引：便于表间关联查询

### 4.3 索引设计原则
- **不使用UNIQUE KEY索引**：避免跨租户的唯一性约束问题
- **租户隔离优先**：所有查询都必须包含租户ID条件
- **性能优化**：为常用查询字段创建适当的索引
- **索引维护**：定期评估索引使用情况，优化查询性能

## 5. 数据关系

### 5.1 表间关系
```
HUB_TIMER_TASK_CONFIG (1) ←→ (1) HUB_TIMER_TASK_INFO
HUB_TIMER_TASK_CONFIG (1) ←→ (N) HUB_TIMER_TASK_RESULT  
HUB_TIMER_TASK_RESULT (1) ←→ (N) HUB_TIMER_TASK_LOG
```

### 5.2 外键约束
考虑到性能和灵活性，外键约束设计为可选：
- 在应用层维护数据一致性
- 数据库层不强制外键约束
- 通过索引保证查询性能
- **复合主键外键**：如果启用外键约束，需要包含 `tenantId` 字段

## 6. 数据生命周期管理

### 6.1 历史数据清理
- **任务执行结果表**: 建议保留最近3-6个月的数据
- **任务执行日志表**: 建议保留最近1-3个月的数据
- **任务配置表**: 长期保留，定期归档非活动任务
- **任务信息表**: 与配置表同步管理

### 6.2 数据归档策略
1. 按时间分区存储历史数据
2. 定期将历史数据迁移到归档库
3. 保留必要的统计汇总数据

## 7. 性能优化建议

### 7.1 查询优化
- 合理使用索引，避免全表扫描
- 对于大数据量的历史查询，考虑分页处理
- 统计查询使用适当的聚合索引

### 7.2 存储优化
- 定期清理过期的执行结果和日志数据
- 对于JSON字段，考虑使用压缩存储
- 合理设置表的存储引擎参数

### 7.3 并发优化
- 使用乐观锁机制（oprSeqFlag字段）
- 避免长事务锁定
- 合理设计事务边界

## 8. 多租户支持

### 8.1 租户隔离
- 所有表都包含 `tenantId` 字段，确保数据完全隔离
- 查询时必须加上租户ID条件，防止跨租户数据访问
- 建议在应用层统一管理租户上下文
- 为 `tenantId` 字段创建索引，提高查询性能

### 8.2 租户管理
- 租户ID应在用户登录时确定并保存在会话中
- 所有数据操作都应自动注入当前租户ID
- 支持租户级别的数据备份和恢复
- 可以按租户进行数据统计和监控

## 9. 预留字段使用

### 9.1 预留字段规范
- 每个表包含 `reserved1` 到 `reserved10` 共10个预留字段
- 统一类型为 `VARCHAR(500) DEFAULT NULL`
- 用于快速业务扩展，避免频繁修改表结构
- 使用时应在代码注释中说明具体用途

### 9.2 预留字段最佳实践
- 按功能模块规划预留字段的使用
- 建议制定预留字段使用约定文档
- 定期评估预留字段使用情况
- 必要时可以将预留字段正式化为业务字段

## 10. 监控和维护

### 10.1 监控指标
- 任务执行成功率（按租户统计）
- 任务执行耗时分布
- 调度器负载情况
- 数据库表大小增长趋势
- 租户级别的资源使用情况

### 10.2 维护建议
- 定期检查和优化索引
- 监控慢查询并进行优化
- 定期备份重要配置数据
- 建立数据质量检查机制
- 按租户进行数据归档和清理

## 11. 使用示例

### 11.1 创建定时任务
```sql
-- 1. 插入任务配置
INSERT INTO HUB_TIMER_TASK_CONFIG (
    taskConfigId, tenantId, taskId, taskName, taskDescription,
    scheduleType, cronExpression, maxRetries, timeoutSeconds,
    enabledFlag, addWho, editWho, oprSeqFlag
) VALUES (
    'TASK_CONFIG_001', 'TENANT_001', 'DAILY_REPORT_TASK', '日报生成任务', '每日凌晨生成业务报表',
    3, '0 0 2 * * ?', 3, 3600,
    'Y', 'ADMIN_001', 'ADMIN_001', 'SEQ_001'
);

-- 2. 初始化任务信息
INSERT INTO HUB_TIMER_TASK_INFO (
    taskInfoId, tenantId, taskConfigId, taskId, taskStatus,
    addWho, editWho, oprSeqFlag
) VALUES (
    'TASK_INFO_001', 'TENANT_001', 'TASK_CONFIG_001', 'DAILY_REPORT_TASK', 1,
    'ADMIN_001', 'ADMIN_001', 'SEQ_001'
);
```

### 11.2 查询任务状态
```sql
-- 查询指定租户的所有活动任务状态
SELECT 
    tc.taskId,
    tc.taskName,
    ti.taskStatus,
    ti.nextRunTime,
    ti.lastRunTime,
    ti.runCount,
    ti.successCount,
    ti.failureCount
FROM HUB_TIMER_TASK_CONFIG tc
INNER JOIN HUB_TIMER_TASK_INFO ti ON tc.tenantId = ti.tenantId AND tc.taskConfigId = ti.taskConfigId
WHERE tc.tenantId = 'TENANT_001' 
  AND tc.activeFlag = 'Y' 
  AND tc.enabledFlag = 'Y';
```

### 11.3 查询执行历史
```sql
-- 查询指定租户和任务的最近执行历史
SELECT 
    taskId,
    executionStartTime,
    executionEndTime,
    executionDurationMs,
    resultSuccess,
    errorMessage
FROM HUB_TIMER_TASK_RESULT 
WHERE tenantId = 'TENANT_001' 
  AND taskId = 'DAILY_REPORT_TASK'
ORDER BY executionStartTime DESC
LIMIT 10;
```

## 12. 版本历史

| 版本 | 日期 | 修改内容 | 修改人 |
|------|------|----------|---------|
| 1.0 | 2024-12-28 | 初始版本，完成基础表结构设计 | System |
| 1.1 | 2024-12-28 | 添加租户ID和预留字段支持，更新设计规范 | System |
| 1.2 | 2024-12-28 | 调整主键结构为复合主键(tenantId, tablePrimaryKey)，移除UNIQUE KEY索引 | System |

---

*本文档遵循项目数据库设计规范，如有疑问请参考 `database-naming-convention.md`* 