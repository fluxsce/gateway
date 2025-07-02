# Gateway Log Writer Architecture

网关日志写入器架构，支持多种输出目标和异步批量写入。

## 架构概述

### 目录结构
```
internal/gateway/logwrite/
├── log_writer.go          # 主要接口和管理器
├── factory.go             # 工厂模式写入器创建
├── types/                 # 类型定义和配置
│   ├── access_log.go      # 访问日志结构
│   ├── log_config.go      # 日志配置结构
│   └── errors.go          # 错误定义
├── console/               # 控制台写入器
│   └── console_writer.go  # 支持彩色格式化输出
├── filewrite/             # 文件写入器
│   └── file_writer.go     # 支持日志轮转和压缩
├── dbwrite/               # 数据库写入器
│   └── db_writer.go       # 支持事务和批量写入
├── mongowrite/            # MongoDB写入器
│   └── mongo_writer.go    # 支持文档存储（待实现）
└── elasticsearch/         # Elasticsearch写入器
    └── es_writer.go       # 支持全文搜索（待实现）
```

## 核心接口

### LogWriter 接口
```go
type LogWriter interface {
    // Write 写入单条日志
    Write(ctx context.Context, log *types.AccessLog) error
    
    // BatchWrite 批量写入日志
    BatchWrite(ctx context.Context, logs []*types.AccessLog) error
    
    // Flush 刷新缓冲区
    Flush(ctx context.Context) error
    
    // Close 关闭写入器
    Close() error
}
```

### LogManager Map 缓存架构
LogManager 使用 `map[types.LogOutputTarget]LogWriter` 缓存具体的写入器实现：

```go
type LogManager struct {
    // 写入器实例映射 - 使用map缓存具体实现
    writers map[types.LogOutputTarget]LogWriter
    
    // 启用的输出目标列表
    enabledTargets []types.LogOutputTarget
    
    // ... 其他字段
}
```

**Map 缓存的优势:**
- ✅ **O(1) 快速查找**: 根据目标类型快速获取写入器
- ✅ **类型安全**: 强类型的目标类型作为 key
- ✅ **内存高效**: 只缓存实际启用的写入器
- ✅ **易于管理**: 支持运行时检查和统计

## 支持的输出目标

### 1. 控制台输出 (Console)
- **特性**: 彩色输出、多种格式、实时显示
- **适用场景**: 开发调试、实时监控
- **配置示例**:
```yaml
output_targets: "console"
log_format: "text"  # text, json, csv
```

### 2. 文件输出 (File)
- **特性**: 日志轮转、压缩存储、缓冲写入、自动清理
- **适用场景**: 生产环境日志存储
- **配置示例**:
```yaml
output_targets: "file"
file_config:
  path: "/var/log/gateway"
  prefix: "access"
  extension: ".log"
  max_size: 100      # MB
  max_age: 30        # 天
  max_backups: 10
  compress: true
  rotation_time: "daily"  # hourly, daily
```

### 3. 数据库输出 (Database)
- **特性**: 事务支持、批量写入、SQL优化、连接池
- **适用场景**: 结构化数据存储和查询
- **配置示例**:
```yaml
output_targets: "database"
database_config:
  driver: "mysql"
  dsn: "user:pass@tcp(localhost:3306)/logs"
  table_name: "gateway_access_logs"
  batch_size: 100
```

### 4. MongoDB输出 (MongoDB)
- **特性**: 文档存储、自动索引、分片支持、聚合查询
- **适用场景**: 非结构化数据存储和分析
- **配置示例**:
```yaml
output_targets: "mongodb"
mongodb_config:
  uri: "mongodb://localhost:27017"
  database: "gateway_logs"
  collection: "access_logs"
  batch_size: 100
```

### 5. Elasticsearch输出 (Elasticsearch)
- **特性**: 全文搜索、实时分析、索引管理、聚合统计
- **适用场景**: 日志搜索和分析
- **配置示例**:
```yaml
output_targets: "elasticsearch"
elasticsearch_config:
  addresses: ["http://localhost:9200"]
  index_name: "gateway-logs"
  index_pattern: "daily"  # daily, weekly, monthly
  batch_size: 100
```

## 使用方法

### 1. 初始化日志管理器
```go
import "gohub/internal/gateway/logwrite"

// 创建配置
config := &types.LogConfig{
    OutputTargets: "console,file",
    LogFormat: "json",
    EnableAsyncLogging: "Y",
    BatchSize: 100,
    AsyncQueueSize: 1000,
    BatchTimeoutMs: 5000,
}

// 初始化管理器
err := logwrite.InitLogManager(config)
if err != nil {
    log.Fatal("Failed to init log manager:", err)
}
```

### 2. 写入日志
```go
// 获取管理器实例
manager := logwrite.GetLogManager()

// 创建访问日志
accessLog := &types.AccessLog{
    TraceID: "trace-123",
    RequestMethod: "GET",
    RequestPath: "/api/users",
    GatewayStatusCode: 200,
    GatewayReceivedTime: time.Now(),
    // ... 其他字段
}

// 写入单条日志到所有启用的写入器
err := manager.Write(ctx, accessLog)

// 写入单条日志到特定目标
err := manager.WriteToTarget(ctx, types.LogOutputFile, accessLog)

// 批量写入到所有写入器
logs := []*types.AccessLog{accessLog1, accessLog2}
err := manager.BatchWrite(ctx, logs)

// 批量写入到特定目标
err := manager.BatchWriteToTarget(ctx, types.LogOutputDatabase, logs)

// 刷新所有缓冲区
err := manager.Flush(ctx)

// 刷新特定目标的缓冲区
err := manager.FlushTarget(ctx, types.LogOutputFile)
```

### 3. Map 缓存功能使用
```go
// 检查写入器是否存在
if manager.HasWriter(types.LogOutputConsole) {
    // 获取特定写入器
    writer := manager.GetWriter(types.LogOutputConsole)
    // 直接使用写入器
    err := writer.Write(ctx, accessLog)
}

// 获取启用的目标列表
targets := manager.GetEnabledTargets()
fmt.Printf("启用的目标: %v\n", targets)

// 获取写入器统计信息
stats := manager.GetWriterStats()
fmt.Printf("总写入器数量: %d\n", stats["total_writers"])
```

### 4. 多输出目标配置
```yaml
# 同时输出到控制台和文件
output_targets: "console,file"

# 同时输出到数据库和Elasticsearch
output_targets: "database,elasticsearch"
```

## 异步写入

### 配置异步写入
```yaml
enable_async_logging: "Y"
async_queue_size: 1000      # 异步队列大小
batch_size: 100             # 批量写入大小
batch_timeout_ms: 5000      # 批量写入超时
async_flush_interval_ms: 1000  # 异步刷新间隔
```

### 异步写入流程
1. 日志写入异步队列
2. 后台协程批量处理日志
3. 定时刷新缓冲区
4. 优雅关闭时处理剩余日志

## 工厂模式

### 使用工厂创建写入器
```go
// 创建工厂
factory := logwrite.NewWriterFactory()

// 验证配置
err := factory.ValidateConfig(config)

// 创建写入器
writers, err := factory.CreateWriters(config)

// 获取支持的目标
targets := factory.GetSupportedTargets()

// 获取写入器信息
infos := factory.GetWriterInfos()
```

## 性能优化

### 1. 缓冲写入
- 所有写入器都支持缓冲写入
- 可配置缓冲区大小和刷新间隔
- 自动批量提交提高性能

### 2. 异步处理
- 支持异步日志写入
- 避免阻塞主业务流程
- 可配置队列大小和处理协程数

### 3. 连接池
- 数据库写入器使用连接池
- 减少连接创建开销
- 支持连接重用和超时管理

## 错误处理

### 1. 写入器级别错误
- 单个写入器失败不影响其他写入器
- 记录错误日志便于排查
- 支持重试机制

### 2. 优雅降级
- 写入器创建失败时继续创建其他写入器
- 至少保证一个写入器可用
- 提供默认控制台输出

### 3. 资源清理
- 支持优雅关闭
- 自动清理资源和连接
- 处理剩余缓冲区数据

## 扩展开发

### 添加新的写入器
1. 在对应子目录创建写入器实现
2. 实现 `LogWriter` 接口
3. 在工厂中添加创建方法
4. 更新配置类型定义
5. 添加相应的测试用例

### 示例：添加Redis写入器
```go
// 1. 创建 redis/redis_writer.go
type RedisWriter struct {
    config *types.LogConfig
    client redis.Client
}

func (w *RedisWriter) Write(ctx context.Context, log *types.AccessLog) error {
    // 实现Redis写入逻辑
}

// 2. 在factory.go中添加
case types.LogOutputRedis:
    return f.createRedisWriter(config)
```

## 监控和指标

### 建议添加的监控指标
- 写入延迟
- 写入成功率
- 队列长度
- 缓冲区使用率
- 错误计数

### 健康检查
- 写入器连接状态
- 队列是否阻塞
- 磁盘空间检查
- 网络连接状态

## 最佳实践

1. **配置管理**: 使用配置文件管理不同环境的日志配置
2. **性能测试**: 在生产环境部署前进行性能测试
3. **监控告警**: 设置日志写入的监控和告警
4. **容量规划**: 根据日志量规划存储容量
5. **安全考虑**: 敏感信息脱敏处理
6. **备份策略**: 重要日志的备份和恢复策略 