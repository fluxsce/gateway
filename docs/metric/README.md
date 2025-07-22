# 系统指标采集表结构设计

## 文档说明

本目录包含系统指标采集相关的数据库表结构设计文档。

### 文件列表

- `metric_tables_design.md` - 详细的表结构设计文档，包含设计原则、表结构定义、索引设计等
- `metric_tables_create.sql` - SQL创建脚本，可直接执行创建所有表结构
- `README.md` - 本文档，提供概览和使用指南

## 设计概览

### 核心设计思路

1. **主从表结构**：以服务器信息为主表，各指标类型分别对应独立的日志表
2. **遵循规范**：严格遵循项目的数据库设计规范
3. **无外键约束**：在应用层维护数据完整性，避免数据库层面的外键约束
4. **优化索引**：针对常见查询场景设计合理的索引结构
5. **复合主键设计**：所有表都使用复合主键（tenantId + 业务主键），确保多租户数据隔离

### 表结构概览

| 表名 | 说明 | 对应Go结构 | 复合主键 |
|------|------|-----------|----------|
| `HUB_METRIC_SERVER_INFO` | 服务器信息主表 | `SystemMetrics` | `(tenantId, metricServerId)` |
| `HUB_METRIC_CPU_LOG` | CPU采集日志表 | `CPUMetrics` | `(tenantId, metricCpuLogId)` |
| `HUB_METRIC_MEMORY_LOG` | 内存采集日志表 | `MemoryMetrics` | `(tenantId, metricMemoryLogId)` |
| `HUB_METRIC_DISK_PART_LOG` | 磁盘分区日志表 | `DiskPartition` | `(tenantId, metricDiskPartitionLogId)` |
| `HUB_METRIC_DISK_IO_LOG` | 磁盘IO日志表 | `DiskIOStats` | `(tenantId, metricDiskIoLogId)` |
| `HUB_METRIC_NETWORK_LOG` | 网络接口日志表 | `NetworkInterface` | `(tenantId, metricNetworkLogId)` |
| `HUB_METRIC_PROCESS_LOG` | 进程信息日志表 | `ProcessInfo` | `(tenantId, metricProcessLogId)` |
| `HUB_METRIC_PROCSTAT_LOG` | 进程统计日志表 | `ProcessSystemStats` | `(tenantId, metricProcessStatsLogId)` |
| `HUB_METRIC_TEMPERATURE_LOG` | 温度信息日志表 | `TemperatureInfo` | `(tenantId, metricTemperatureLogId)` |

## 主键设计规范

### 复合主键设计

根据数据库设计规范，所有表必须使用复合主键，包含租户ID字段：

- **主键格式**：`PRIMARY KEY (tenantId, 业务主键)`
- **设计目的**：确保多租户数据隔离，避免跨租户数据冲突
- **性能优势**：复合主键天然支持按租户查询，无需额外索引

### 多租户数据隔离

- 每个表的主键都包含 `tenantId` 字段
- 确保不同租户的数据完全隔离
- 查询时必须包含租户ID条件
- 避免跨租户数据访问的安全风险

## 使用指南

### 创建表结构

```bash
# 执行SQL脚本创建所有表
mysql -u username -p database_name < metric_tables_create.sql
```

### 数据关联

所有采集日志表通过 `metricServerId` 字段关联到服务器信息主表：

```sql
-- 查询特定服务器的CPU使用率
SELECT s.hostname, c.usagePercent, c.collectTime
FROM HUB_METRIC_SERVER_INFO s
JOIN HUB_METRIC_CPU_LOG c ON s.metricServerId = c.metricServerId
WHERE s.tenantId = 'TENANT001' AND c.tenantId = 'TENANT001'
  AND s.hostname = 'server01';
```

### 索引使用

每个表都设计了针对性的索引：

- **复合主键索引**：天然支持按租户查询
- **单字段索引**：服务器ID、时间等
- **复合索引**：服务器+时间、租户+时间等常见组合查询
- **业务索引**：使用率、状态等业务查询字段

### 数据清理

建议定期清理历史数据：

```sql
-- 删除3个月前的CPU日志
DELETE FROM HUB_METRIC_CPU_LOG 
WHERE tenantId = 'TENANT001'
  AND collectTime < DATE_SUB(NOW(), INTERVAL 3 MONTH);
```

## 规范合规性

### ✅ 已遵循的规范

1. **命名规范**
   - 表名以 `HUB_` 开头，使用大写字母和下划线
   - 字段名使用驼峰命名法
   - 主键字段体现业务含义

2. **字段规范**
   - 包含所有必须的通用字段
   - 时间字段使用 `DATETIME` 类型
   - 状态字段使用 `VARCHAR(1)` 类型，值为 'Y'/'N'
   - 禁止使用自增字段

3. **主键规范**
   - **复合主键设计**：所有表都使用复合主键（tenantId + 业务主键）
   - **多租户隔离**：确保不同租户的数据完全隔离
   - **性能优化**：复合主键天然支持按租户查询

4. **约束规范**
   - 不使用外键约束
   - 合理的索引设计
   - 详细的字段和表注释

## 性能优化

### 分区建议

对于数据量大的表，建议考虑按时间分区：

```sql
-- 按月分区示例
ALTER TABLE HUB_METRIC_CPU_LOG 
PARTITION BY RANGE (YEAR(collectTime)*100 + MONTH(collectTime));
```

### 查询优化

- 始终包含租户ID条件
- 使用时间范围查询
- 利用复合索引进行多维度查询
- 避免全表扫描

## 扩展性

### 新增指标类型

1. 在 `pkg/metric/types/metrics.go` 中定义新的数据结构
2. 按照相同的模式创建对应的数据库表
3. 确保包含所有必须的通用字段
4. 设计合理的索引结构
5. **必须使用复合主键**：`PRIMARY KEY (tenantId, 业务主键)`

### 字段扩展

- 使用 `extProperty` 字段存储JSON格式的扩展属性
- 使用 `reserved1` 到 `reserved10` 预留字段进行快速扩展

## 维护建议

1. **定期清理**：按照数据保留策略定期清理历史数据
2. **监控性能**：关注查询性能，必要时调整索引
3. **备份策略**：制定合理的备份和恢复策略
4. **容量规划**：根据数据增长情况进行容量规划
5. **租户隔离**：确保多租户环境下的数据安全隔离

## 相关文档

- [数据库设计规范](../database-naming-convention.md)
- [系统指标采集源码](../../pkg/metric/types/metrics.go)
- [详细设计文档](metric_tables_design.md) 