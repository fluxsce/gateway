# 指标数据类型转换

本包提供了将系统指标数据转换为数据库记录的功能。每个表结构都包含了相应的转换方法。

## 使用方法

### 1. CPU 指标转换

```go
// 单个CPU指标转换
cpuLog := types.NewCpuLogFromMetrics(
    cpuMetrics,     // *metricTypes.CPUMetrics
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)
```

### 2. 内存指标转换

```go
// 单个内存指标转换
memoryLog := types.NewMemoryLogFromMetrics(
    memoryMetrics,  // *metricTypes.MemoryMetrics
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)
```

### 3. 磁盘指标转换

```go
// 批量磁盘分区转换
partitionLogs := types.NewDiskPartitionLogsFromMetrics(
    diskMetrics,    // *metricTypes.DiskMetrics
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)

// 批量磁盘IO转换
ioLogs := types.NewDiskIoLogsFromMetrics(
    diskMetrics,    // *metricTypes.DiskMetrics
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)

// 单个磁盘分区转换
partitionLog := types.NewDiskPartitionLogFromMetrics(
    &partition,     // *metricTypes.DiskPartition
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
    index,          // int
)
```

### 4. 网络指标转换

```go
// 批量网络接口转换
networkLogs := types.NewNetworkLogsFromMetrics(
    networkMetrics, // *metricTypes.NetworkMetrics
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)

// 单个网络接口转换
networkLog := types.NewNetworkLogFromMetrics(
    &iface,         // *metricTypes.NetworkInterface
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
    index,          // int
)
```

### 5. 进程指标转换

```go
// 进程统计信息转换
statsLog := types.NewProcessStatsLogFromMetrics(
    systemProcesses, // *metricTypes.ProcessSystemStats
    tenantId,        // string
    serverId,        // string
    operator,        // string
    collectTime,     // time.Time
    oprSeqFlag,      // string
)

// 单个进程信息转换
processLog := types.NewProcessLogFromMetrics(
    processInfo,    // *metricTypes.ProcessInfo
    tenantId,       // string
    serverId,       // string
    operator,       // string
    collectTime,    // time.Time
    oprSeqFlag,     // string
)
```

## 设计优势

1. **封装性**: 转换逻辑封装在对应的类型定义中，便于维护
2. **可重用性**: 转换方法可以在不同的场景中复用
3. **类型安全**: 使用强类型参数，避免运行时错误
4. **一致性**: 所有转换方法都遵循相同的命名规范和参数模式
5. **批量处理**: 提供批量转换方法，提高处理效率
6. **并发安全**: 使用`random.Generate32BitRandomString()`生成唯一主键，确保并发环境下的ID唯一性

## 主键生成策略

### 并发安全的唯一ID生成

所有转换方法都使用`random.Generate32BitRandomString()`来生成主键，该方法具有以下特性：

- **高并发安全**: 使用时间戳+进程标识+原子计数器+强随机数组合
- **全局唯一**: 即使在多进程、多线程环境下也能保证唯一性
- **性能优化**: 避免了数据库查询来检查ID重复的开销
- **32位长度**: 符合数据库VARCHAR(32)字段要求

### ID生成算法

```
ID = 时间戳(13位) + 进程ID(4位) + 计数器(6位) + 随机数(9位) = 32位
```

- 时间戳：确保时间顺序性
- 进程ID：区分不同应用实例
- 计数器：原子递增，避免同进程内重复
- 随机数：增加随机性，防止ID预测

## 注意事项

- 所有转换方法都会自动生成唯一的32位随机ID
- 时间戳会自动设置为指定的采集时间
- JSON序列化会自动处理复杂数据结构（如IP地址列表、命令行参数等）
- 空值处理会自动进行适当的转换
- 主键生成完全并发安全，无需担心ID重复问题

## 性能特性

- **无锁设计**: 主键生成过程无需锁机制
- **高吞吐量**: 支持高并发场景下的大量数据转换
- **内存友好**: 转换过程中避免不必要的内存分配
- **批量优化**: 提供批量转换方法，减少函数调用开销 