# 指标采集DAO操作说明

## 概述

该模块为之前设计的指标采集日志表提供完整的DAO操作层，包括类型定义和数据库操作。

## 目录结构

```
internal/metric_collect/
├── types/           # 类型定义
│   ├── constants.go           # 常量定义
│   ├── server_info.go         # 服务器信息类型
│   ├── cpu_log.go            # CPU日志类型
│   ├── memory_log.go         # 内存日志类型
│   ├── disk_partition_log.go  # 磁盘分区日志类型
│   ├── disk_io_log.go        # 磁盘IO日志类型
│   ├── network_log.go        # 网络日志类型
│   ├── process_log.go        # 进程日志类型
│   ├── process_stats_log.go  # 进程统计日志类型
│   └── temperature_log.go    # 温度日志类型
├── dao/             # DAO操作类
│   ├── server_info_dao.go         # 服务器信息DAO
│   ├── cpu_log_dao.go            # CPU日志DAO
│   ├── memory_log_dao.go         # 内存日志DAO
│   ├── disk_partition_log_dao.go  # 磁盘分区日志DAO
│   ├── disk_io_log_dao.go        # 磁盘IO日志DAO
│   ├── network_log_dao.go        # 网络日志DAO
│   ├── process_log_dao.go        # 进程日志DAO
│   ├── process_stats_log_dao.go  # 进程统计日志DAO
│   └── temperature_log_dao.go    # 温度日志DAO
└── README.md        # 说明文档
```

## 表结构设计

### 主表结构
- **HUB_METRIC_SERVER_INFO**: 服务器信息主表，存储服务器基本信息
- **HUB_METRIC_CPU_LOG**: CPU采集日志表，存储CPU使用情况
- **HUB_METRIC_MEMORY_LOG**: 内存采集日志表，存储内存使用情况
- **HUB_METRIC_DISK_PARTITION_LOG**: 磁盘分区日志表，存储磁盘分区信息
- **HUB_METRIC_DISK_IO_LOG**: 磁盘IO日志表，存储磁盘IO统计
- **HUB_METRIC_NETWORK_LOG**: 网络接口日志表，存储网络接口信息
- **HUB_METRIC_PROCESS_LOG**: 进程信息日志表，存储进程详细信息
- **HUB_METRIC_PROCESS_STATS_LOG**: 进程统计日志表，存储进程统计信息
- **HUB_METRIC_TEMPERATURE_LOG**: 温度信息日志表，存储温度传感器数据

### 表关系
- 服务器信息表为主表，其他表通过`metricServerId`字段关联
- 所有表都支持多租户，通过`tenantId`字段区分
- 所有表都包含完整的审计字段和扩展字段

## 使用方法

### 1. 导入包

```go
import (
    "gateway/internal/metric_collect/dao"
    "gateway/internal/metric_collect/types"
    "gateway/pkg/database"
)
```

### 2. 创建DAO实例

```go
// 假设已经有数据库连接
db := database.NewDatabase(...)

// 创建各种DAO实例
serverInfoDAO := dao.NewServerInfoDAO(db)
cpuLogDAO := dao.NewCpuLogDAO(db)
memoryLogDAO := dao.NewMemoryLogDAO(db)
// ... 其他DAO实例
```

### 3. 插入数据

```go
// 创建服务器信息
serverInfo := &types.ServerInfo{
    MetricServerId: "server_001",
    TenantId:       "tenant_001",
    Hostname:       "web-server-01",
    OsType:         "Linux",
    OsVersion:      "Ubuntu 20.04",
    Architecture:   "x86_64",
    BootTime:       time.Now().Add(-24 * time.Hour),
    AddWho:         "system",
    EditWho:        "system",
    OprSeqFlag:     "SEQ001",
}

// 插入服务器信息
err := serverInfoDAO.InsertServerInfo(ctx, serverInfo)
if err != nil {
    log.Printf("插入服务器信息失败: %v", err)
}

// 创建CPU日志
cpuLog := &types.CpuLog{
    MetricCpuLogId: "cpu_log_001",
    TenantId:       "tenant_001",
    MetricServerId: "server_001",
    UsagePercent:   75.5,
    UserPercent:    45.2,
    SystemPercent:  20.3,
    IdlePercent:    24.5,
    CoreCount:      8,
    LogicalCount:   16,
    LoadAvg1:       2.5,
    LoadAvg5:       2.3,
    LoadAvg15:      2.1,
    CollectTime:    time.Now(),
    AddWho:         "system",
    EditWho:        "system",
    OprSeqFlag:     "SEQ002",
}

// 插入CPU日志
err = cpuLogDAO.InsertCpuLog(ctx, cpuLog)
if err != nil {
    log.Printf("插入CPU日志失败: %v", err)
}
```

### 4. 批量插入数据

```go
// 批量插入CPU日志
cpuLogs := []*types.CpuLog{
    // ... 多个CPU日志对象
}

err := cpuLogDAO.BatchInsertCpuLog(ctx, cpuLogs)
if err != nil {
    log.Printf("批量插入CPU日志失败: %v", err)
}
```

### 5. 删除数据

```go
// 软删除：将activeFlag设置为'N'
err := cpuLogDAO.DeleteCpuLog(ctx, "tenant_001", "cpu_log_001")
if err != nil {
    log.Printf("删除CPU日志失败: %v", err)
}

// 物理删除：根据时间范围删除
beforeTime := time.Now().Add(-30 * 24 * time.Hour) // 30天前
err = cpuLogDAO.DeleteCpuLogByTime(ctx, "tenant_001", beforeTime)
if err != nil {
    log.Printf("按时间删除CPU日志失败: %v", err)
}

// 物理删除：根据服务器ID删除
err = cpuLogDAO.DeleteCpuLogByServer(ctx, "tenant_001", "server_001")
if err != nil {
    log.Printf("按服务器删除CPU日志失败: %v", err)
}
```

## DAO操作说明

### 通用操作

每个DAO类都提供以下操作：

1. **Insert[TableName]**: 插入单条记录
2. **BatchInsert[TableName]**: 批量插入记录
3. **Delete[TableName]**: 软删除记录（设置activeFlag='N'）
4. **Delete[TableName]ByTime**: 根据时间范围物理删除记录
5. **Delete[TableName]ByServer**: 根据服务器ID物理删除记录

### 字段自动处理

- **添加时间(addTime)**: 自动设置为当前时间
- **修改时间(editTime)**: 自动设置为当前时间
- **版本号(currentVersion)**: 自动设置为1
- **活动标记(activeFlag)**: 自动设置为'Y'

### 必填字段验证

所有DAO操作都会验证以下必填字段：
- **日志ID**: 各种日志表的主键ID
- **租户ID(tenantId)**: 多租户标识
- **服务器ID(metricServerId)**: 关联的服务器ID
- **创建人ID(addWho)**: 记录创建者
- **修改人ID(editWho)**: 记录修改者
- **操作序列标识(oprSeqFlag)**: 用于追踪操作序列

## 数据库依赖

该模块依赖于项目中的数据库抽象层：
- `gateway/pkg/database`: 提供数据库操作接口
- 支持MySQL、Oracle、SQLite等多种数据库
- 支持事务操作和批量操作

## 使用注意事项

1. **并发安全**: 所有DAO操作都是线程安全的
2. **事务支持**: 可以在事务中调用DAO操作
3. **错误处理**: 所有操作都返回详细的错误信息
4. **性能优化**: 提供批量操作接口，适合大数据量场景
5. **数据完整性**: 通过必填字段验证确保数据完整性

## 扩展说明

如需扩展功能，可以：
1. 在types包中添加新的结构体字段
2. 在dao包中添加新的查询方法
3. 保持与现有接口的兼容性 