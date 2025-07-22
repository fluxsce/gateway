# ClickHouse数据库实现

## 概述

本模块提供了ClickHouse数据库的完整实现，参考MySQL实现的架构，符合统一的`database.Database`接口规范。ClickHouse是一个用于联机分析处理(OLAP)的列式数据库管理系统，特别适合大数据分析和实时查询场景。

## 快速导航

- [基本配置和使用](#使用示例) - 单机配置和基本操作
- [集群配置详解](#clickhouse集群配置详解) - **完整的集群配置支持**
- [配置参数详解](#clickhouse配置参数详解) - 所有参数说明
- [表引擎和性能优化](#clickhouse特有的考虑) - ClickHouse特有功能
- [故障处理和监控](#错误处理) - 生产环境最佳实践

## 特性

- ✅ **统一接口** - 完全符合`database.Database`接口规范
- ✅ **列式存储优化** - 针对ClickHouse列式存储特性进行优化
- ✅ **高性能批量插入** - 使用单条大INSERT语句优化批量操作
- ✅ **集群支持** - 完整的集群配置支持，包括负载均衡、故障转移、分片和副本
- ✅ **事务支持** - 有限的事务支持，主要用于批量操作原子性
- ✅ **连接池管理** - 针对分析型查询优化的连接池配置
- ✅ **日志记录** - 支持SQL执行日志和慢查询检测
- ✅ **结构体映射** - 自动映射Go结构体到数据库表
- ✅ **压缩传输** - 支持数据压缩，减少网络开销
- ⚠️ **有限UPDATE/DELETE** - ClickHouse的UPDATE/DELETE支持有限，建议使用其他方案

## 依赖

```bash
go get github.com/ClickHouse/clickhouse-go/v2
# 或者使用v1版本
go get github.com/ClickHouse/clickhouse-go
```

**注意：** 请根据你的ClickHouse服务器版本选择合适的驱动版本。

## 使用示例

### 基本配置

```go
import (
    "gateway/pkg/database"
    "gateway/pkg/database/dbtypes"
    _ "gateway/pkg/database/clickhouse" // 导入ClickHouse实现
)

// ClickHouse配置
config := &dbtypes.DbConfig{
    Name:    "clickhouse_main",
    Enabled: true,
    Driver:  dbtypes.DriverClickHouse,
    Connection: dbtypes.ConnectionConfig{
        Host:     "localhost",
        Port:     9000,           // ClickHouse原生协议端口
        Username: "default",
        Password: "",
        Database: "default",
        
        // ClickHouse特有参数
        ClickHouseCompress:             true,   // 启用压缩
        ClickHouseSecure:               false,  // 不使用SSL
        ClickHouseDebug:                false,  // 关闭调试
        ClickHouseConnectionTimeout:    10,     // 连接超时10秒
        ClickHouseReadTimeout:          30,     // 读取超时30秒
        ClickHouseWriteTimeout:         30,     // 写入超时30秒
        ClickHouseDialTimeout:          30,     // 拨号超时30秒
        ClickHouseMaxExecutionTime:     3600,   // 最大执行时间1小时
        ClickHouseClientName:           "gateway-app", // 客户端名称
        ClickHouseBlockBufferSize:      10,     // 块缓冲区大小
        ClickHouseMaxCompressionBuffer: 10240,  // 最大压缩缓冲区
        ClickHouseConnOpenStrategy:     "time", // 连接打开策略
        ClickHouseSettings:             "max_memory_usage=10000000000", // ClickHouse设置
    },
    Pool: dbtypes.PoolConfig{
        MaxOpenConns:    50,      // ClickHouse推荐较大连接池
        MaxIdleConns:    25,
        ConnMaxLifetime: 3600,    // 1小时
        ConnMaxIdleTime: 1800,    // 30分钟
    },
    Log: dbtypes.LogConfig{
        Enable:        true,
        SlowThreshold: 1000,      // ClickHouse查询通常较慢，设置为1秒
    },
}

// 连接数据库
db, err := database.Open(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### HTTP接口配置

```go
// 使用HTTP接口（端口8123）
config := &dbtypes.DbConfig{
    Name:    "clickhouse_http",
    Enabled: true,
    Driver:  dbtypes.DriverClickHouse,
    DSN:     "http://localhost:8123/default?username=default&password=",
    // 或者使用连接配置
    Connection: dbtypes.ConnectionConfig{
        Host:     "localhost",
        Port:     8123,           // HTTP接口端口
        Username: "default",
        Password: "",
        Database: "default",
    },
}
```

### 基本操作

```go
// 定义数据结构
type Event struct {
    ID        uint64    `db:"id"`
    UserID    uint64    `db:"user_id"`
    Action    string    `db:"action"`
    Timestamp time.Time `db:"timestamp"`
    Value     float64   `db:"value"`
}

// 插入单条记录
event := Event{
    UserID:    123,
    Action:    "click",
    Timestamp: time.Now(),
    Value:     1.0,
}

id, err := db.Insert(ctx, "events", event, true)
if err != nil {
    log.Fatal(err)
}

// 批量插入（推荐用法）
events := []Event{
    {UserID: 123, Action: "click", Timestamp: time.Now(), Value: 1.0},
    {UserID: 124, Action: "view", Timestamp: time.Now(), Value: 2.0},
    {UserID: 125, Action: "purchase", Timestamp: time.Now(), Value: 99.99},
}

rowsAffected, err := db.BatchInsert(ctx, "events", events, true)
if err != nil {
    log.Fatal(err)
}

// 查询多条记录
var results []Event
err = db.Query(ctx, &results, 
    "SELECT * FROM events WHERE user_id = ? AND timestamp >= ?", 
    []interface{}{123, time.Now().Add(-24*time.Hour)}, true)
if err != nil {
    log.Fatal(err)
}

// 查询单条记录
var event Event
err = db.QueryOne(ctx, &event, 
    "SELECT * FROM events WHERE id = ?", 
    []interface{}{1}, true)
if err != nil {
    if err == database.ErrRecordNotFound {
        log.Println("记录不存在")
    } else {
        log.Fatal(err)
    }
}

// 分析查询（ClickHouse擅长的领域）
type Analytics struct {
    Action string  `db:"action"`
    Count  uint64  `db:"count"`
    AvgValue float64 `db:"avg_value"`
}

var analytics []Analytics
err = db.Query(ctx, &analytics, `
    SELECT 
        action,
        COUNT(*) as count,
        AVG(value) as avg_value
    FROM events 
    WHERE timestamp >= ? 
    GROUP BY action 
    ORDER BY count DESC`,
    []interface{}{time.Now().Add(-7*24*time.Hour)}, true)
```

### 事务使用

```go
// 注意：ClickHouse的事务支持有限，主要用于批量操作
err = db.InTx(ctx, nil, func(txCtx context.Context) error {
    // 在事务中执行批量插入
    _, err := db.BatchInsert(txCtx, "events", events, false)
    if err != nil {
        return err
    }
    
    // 可以执行多个相关操作
    _, err = db.Insert(txCtx, "user_stats", userStat, false)
    return err
})

if err != nil {
    log.Fatal(err)
}
```

### 高级查询

```go
// 时间序列分析
type TimeSeriesData struct {
    Date  string  `db:"date"`
    Hour  uint8   `db:"hour"`
    Count uint64  `db:"count"`
}

var tsData []TimeSeriesData
err = db.Query(ctx, &tsData, `
    SELECT 
        toDate(timestamp) as date,
        toHour(timestamp) as hour,
        COUNT(*) as count
    FROM events 
    WHERE timestamp >= ? AND timestamp < ?
    GROUP BY date, hour
    ORDER BY date, hour`,
    []interface{}{startDate, endDate}, true)

// 使用ClickHouse特有的聚合函数
type UniqueAnalytics struct {
    Action      string `db:"action"`
    UniqueUsers uint64 `db:"unique_users"`
    TotalEvents uint64 `db:"total_events"`
}

var uniqueStats []UniqueAnalytics
err = db.Query(ctx, &uniqueStats, `
    SELECT 
        action,
        uniq(user_id) as unique_users,
        count() as total_events
    FROM events 
    WHERE timestamp >= ?
    GROUP BY action`,
    []interface{}{time.Now().Add(-24*time.Hour)}, true)
```

## 配置文件示例

### YAML配置

```yaml
database:
  default: "clickhouse_main"
  connections:
    clickhouse_main:
      enabled: true
      driver: "clickhouse"
      # 使用TCP原生协议（推荐）
      connection:
        host: "localhost"
        port: 9000
        username: "default"
        password: ""
        database: "analytics"
        
        # ClickHouse特有参数
        clickhouse_compress: true                    # 启用压缩（默认true）
        clickhouse_secure: false                     # SSL/TLS连接（默认false）
        clickhouse_debug: false                      # 调试模式（默认false）
        clickhouse_connection_timeout: 10            # 连接超时(秒)
        clickhouse_read_timeout: 30                  # 读取超时(秒)
        clickhouse_write_timeout: 30                 # 写入超时(秒)
        clickhouse_dial_timeout: 30                  # 拨号超时(秒)
        clickhouse_max_execution_time: 3600          # 最大执行时间(秒)
        clickhouse_client_name: "gateway-app"          # 客户端名称
        clickhouse_block_buffer_size: 10             # 块缓冲区大小
        clickhouse_max_compression_buffer: 10240     # 最大压缩缓冲区
        clickhouse_conn_open_strategy: "time"        # 连接打开策略(in_order/time/random)
        clickhouse_alt_hosts: "host2:9000,host3:9000" # 备用主机
        clickhouse_settings: "max_memory_usage=10000000000,send_logs_level=trace" # ClickHouse设置
      
      pool:
        max_open_conns: 50
        max_idle_conns: 25
        conn_max_lifetime: 3600
        conn_max_idle_time: 1800
        
      log:
        enable: true
        slow_threshold: 1000  # 1秒，ClickHouse查询通常较慢
        
    clickhouse_cluster:
      enabled: true
      driver: "clickhouse"
      # 集群配置示例
      connection:
        host: "clickhouse-node1"
        port: 9000
        username: "analytics_user"
        password: "secure_password"
        database: "analytics_cluster"
        
        # 集群相关配置
        clickhouse_compress: true
        clickhouse_secure: true                      # 生产环境使用SSL
        clickhouse_connection_timeout: 15
        clickhouse_read_timeout: 60                  # 分析查询可能较慢
        clickhouse_write_timeout: 60
        clickhouse_alt_hosts: "clickhouse-node2:9000,clickhouse-node3:9000"
        clickhouse_client_name: "gateway-analytics"
        clickhouse_settings: "max_memory_usage=20000000000,max_threads=8"
      
      pool:
        max_open_conns: 100                          # 集群可以支持更多连接
        max_idle_conns: 50
        conn_max_lifetime: 7200                      # 2小时
        conn_max_idle_time: 3600                     # 1小时
        
    clickhouse_http:
      enabled: true
      driver: "clickhouse"
      # 使用HTTP接口（简单配置）
      dsn: "http://localhost:8123/analytics?username=default&password=&compress=1"
      
      pool:
        max_open_conns: 30
        max_idle_conns: 15
        conn_max_lifetime: 3600
        conn_max_idle_time: 1800
```

### 环境变量配置

```bash
# ClickHouse基本配置
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_DATABASE=analytics

# ClickHouse特有参数
CLICKHOUSE_COMPRESS=true
CLICKHOUSE_SECURE=false
CLICKHOUSE_DEBUG=false
CLICKHOUSE_CONNECTION_TIMEOUT=10
CLICKHOUSE_READ_TIMEOUT=30
CLICKHOUSE_WRITE_TIMEOUT=30
CLICKHOUSE_DIAL_TIMEOUT=30
CLICKHOUSE_MAX_EXECUTION_TIME=3600
CLICKHOUSE_CLIENT_NAME=gateway-app
CLICKHOUSE_BLOCK_BUFFER_SIZE=10
CLICKHOUSE_MAX_COMPRESSION_BUFFER=10240
CLICKHOUSE_CONN_OPEN_STRATEGY=time
CLICKHOUSE_ALT_HOSTS=host2:9000,host3:9000
CLICKHOUSE_SETTINGS=max_memory_usage=10000000000,send_logs_level=info

# 或者使用完整DSN
CLICKHOUSE_DSN="tcp://localhost:9000?database=analytics&username=default&password=&compress=1&secure=false&debug=false&connection_timeout=10&read_timeout=30&write_timeout=30&client_name=gateway-app"
```

## ClickHouse配置参数详解

### ClickHouse特有参数及默认值

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `clickhouse_compress` | bool | true | 是否启用压缩传输，建议启用以减少网络开销 |
| `clickhouse_secure` | bool | false | 是否使用SSL/TLS连接，生产环境建议启用 |
| `clickhouse_debug` | bool | false | 是否启用调试模式，仅在开发时启用 |
| `clickhouse_connection_timeout` | int | 10 | 连接超时时间（秒） |
| `clickhouse_read_timeout` | int | 30 | 读取超时时间（秒） |
| `clickhouse_write_timeout` | int | 30 | 写入超时时间（秒） |
| `clickhouse_dial_timeout` | int | 0 | 拨号超时时间（秒），0表示使用系统默认 |
| `clickhouse_max_execution_time` | int | 0 | 最大执行时间（秒），0表示无限制 |
| `clickhouse_client_name` | string | "" | 客户端名称，用于连接标识 |
| `clickhouse_block_buffer_size` | int | 0 | 块缓冲区大小，0表示使用驱动默认值 |
| `clickhouse_max_compression_buffer` | int | 0 | 最大压缩缓冲区大小，0表示使用驱动默认值 |
| `clickhouse_conn_open_strategy` | string | "" | 连接打开策略：in_order、time、random |
| `clickhouse_alt_hosts` | string | "" | 备用主机列表，逗号分隔 |
| `clickhouse_settings` | string | "" | ClickHouse设置参数，格式：key1=value1,key2=value2 |

### 配置示例对比

**最简配置（使用默认值）：**
```yaml
clickhouse_simple:
  driver: "clickhouse"
  connection:
    host: "localhost"
    port: 9000
    username: "default"
    password: ""
    database: "test"
  # 其他参数将使用默认值
```

**生产环境配置（完整参数）：**
```yaml
clickhouse_production:
  driver: "clickhouse"
  connection:
    host: "clickhouse.example.com"
    port: 9440  # SSL端口
    username: "app_user"
    password: "secure_password"
    database: "production_db"
    
    clickhouse_compress: true
    clickhouse_secure: true
    clickhouse_connection_timeout: 15
    clickhouse_read_timeout: 60
    clickhouse_write_timeout: 60
    clickhouse_client_name: "production-app"
    clickhouse_alt_hosts: "clickhouse-backup.example.com:9440"
    clickhouse_settings: "max_memory_usage=8000000000,max_threads=4"
```

### 重要提示：压缩配置的特殊处理

⚠️ **注意**：由于Go语言bool类型的零值是`false`，但ClickHouse建议默认启用压缩，我们采用了以下处理策略：

1. **如果不设置`clickhouse_compress`字段**：默认启用压缩（compress=1）
2. **如果明确设置`clickhouse_compress: true`**：启用压缩（compress=1）
3. **如果明确设置`clickhouse_compress: false`**：仍然默认启用压缩，除非在`clickhouse_settings`中指定`compress=0`

**推荐做法**：
```yaml
# 推荐：明确设置压缩选项
connection:
  clickhouse_compress: true   # 明确启用压缩
  
# 或者在settings中控制
connection:
  clickhouse_settings: "compress=0"  # 明确禁用压缩
```

## ClickHouse集群配置详解

### 集群连接支持概述

✅ **完全支持ClickHouse集群配置**，包括：
- **多节点连接** - 支持主节点+备用节点配置
- **负载均衡策略** - 支持多种连接分发策略
- **故障转移** - 自动切换到可用节点
- **分片集群** - 支持分布式表查询
- **副本集群** - 支持高可用配置

### 1. 基础集群配置

#### 1.1 多节点配置
```yaml
clickhouse_cluster:
  driver: "clickhouse"
  connection:
    host: "clickhouse-node1.example.com"    # 主节点
    port: 9000
    username: "cluster_user"
    password: "secure_password"
    database: "cluster_db"
    
    # 集群关键配置
    clickhouse_alt_hosts: "clickhouse-node2.example.com:9000,clickhouse-node3.example.com:9000"
    clickhouse_conn_open_strategy: "time"    # 负载均衡策略
    clickhouse_connection_timeout: 15
    clickhouse_read_timeout: 60             # 集群查询可能较慢
    clickhouse_write_timeout: 60
```

#### 1.2 负载均衡策略

| 策略 | 说明 | 适用场景 |
|------|------|----------|
| `in_order` | 按顺序连接，优先使用主节点 | 主从架构，读写分离 |
| `time` | 基于响应时间选择最快节点 | 性能优先，智能负载均衡 |
| `random` | 随机选择节点连接 | 简单负载分散 |

#### 1.3 故障转移配置
```yaml
clickhouse_ha:
  driver: "clickhouse"
  connection:
    host: "clickhouse-primary.example.com"
    port: 9000
    username: "ha_user"
    password: "ha_password"
    database: "production_db"
    
    # 高可用配置
    clickhouse_alt_hosts: "clickhouse-secondary.example.com:9000,clickhouse-tertiary.example.com:9000"
    clickhouse_conn_open_strategy: "in_order"  # 优先主节点
    clickhouse_connection_timeout: 10          # 快速故障检测
    clickhouse_dial_timeout: 10
    clickhouse_settings: "connect_timeout_with_failover_ms=5000"  # 故障转移超时
```

### 2. 高级集群配置

#### 2.1 分片集群配置
```yaml
clickhouse_sharded:
  driver: "clickhouse"
  connection:
    host: "clickhouse-shard1.example.com"
    port: 9000
    username: "shard_user"
    password: "shard_password"
    database: "sharded_db"
    
    # 分片集群配置
    clickhouse_alt_hosts: "clickhouse-shard2.example.com:9000,clickhouse-shard3.example.com:9000"
    clickhouse_conn_open_strategy: "random"    # 分片间随机分布
    clickhouse_settings: "distributed_product_mode=global,insert_distributed_sync=1"
    
  pool:
    max_open_conns: 150    # 分片集群需要更多连接
    max_idle_conns: 75
```

#### 2.2 副本集群配置
```yaml
clickhouse_replicated:
  driver: "clickhouse"
  connection:
    host: "clickhouse-replica1.example.com"
    port: 9000
    username: "replica_user"
    password: "replica_password"
    database: "replicated_db"
    
    # 副本集群配置
    clickhouse_alt_hosts: "clickhouse-replica2.example.com:9000,clickhouse-replica3.example.com:9000"
    clickhouse_conn_open_strategy: "time"      # 性能优先选择
    clickhouse_settings: "replication_alter_partitions_sync=2,max_replica_delay_for_distributed_queries=300"
```

### 3. 完整的集群配置示例

#### 3.1 生产环境分布式集群
```yaml
database:
  default: "clickhouse_distributed"
  connections:
    clickhouse_distributed:
      enabled: true
      driver: "clickhouse"
      connection:
        host: "ch-cluster-lb.example.com"      # 负载均衡器
        port: 9000
        username: "distributed_user"
        password: "complex_password_123"
        database: "analytics_distributed"
        
        # 分布式集群配置
        clickhouse_compress: true
        clickhouse_secure: true                 # 生产环境必须启用SSL
        clickhouse_connection_timeout: 15
        clickhouse_read_timeout: 120            # 分布式查询可能很慢
        clickhouse_write_timeout: 120
        clickhouse_dial_timeout: 15
        clickhouse_max_execution_time: 7200     # 2小时超时
        clickhouse_client_name: "distributed-analytics"
        
        # 关键集群参数
        clickhouse_alt_hosts: "ch-node1.example.com:9000,ch-node2.example.com:9000,ch-node3.example.com:9000,ch-node4.example.com:9000"
        clickhouse_conn_open_strategy: "time"   # 智能负载均衡
        
        # 分布式查询优化设置
        clickhouse_settings: "distributed_product_mode=global,insert_distributed_sync=1,max_distributed_connections=8,distributed_directory_monitor_batch_inserts=1,distributed_directory_monitor_sleep_time_ms=50"
        
      pool:
        max_open_conns: 200                     # 大集群需要更多连接
        max_idle_conns: 100
        conn_max_lifetime: 7200                 # 2小时
        conn_max_idle_time: 3600                # 1小时
        
      log:
        enable: true
        slow_threshold: 5000                    # 分布式查询阈值5秒
```

#### 3.2 开发环境简单集群
```yaml
clickhouse_dev_cluster:
  enabled: true
  driver: "clickhouse"
  connection:
    host: "localhost"
    port: 9000
    username: "default"
    password: ""
    database: "dev_cluster"
    
    # 开发集群配置
    clickhouse_compress: true
    clickhouse_secure: false
    clickhouse_debug: true                     # 开发环境启用调试
    clickhouse_alt_hosts: "localhost:9001,localhost:9002"  # 本地多实例
    clickhouse_conn_open_strategy: "in_order"
    clickhouse_client_name: "dev-cluster"
```

### 4. 集群DSN配置方式

#### 4.1 通过DSN直接配置集群
```go
config := &dbtypes.DbConfig{
    Driver: dbtypes.DriverClickHouse,
    // 集群DSN配置
    DSN: "tcp://node1.example.com:9000?database=cluster_db&username=user&password=pass&alt_hosts=node2.example.com:9000,node3.example.com:9000&conn_open_strategy=time&compress=1",
}
```

#### 4.2 环境变量集群配置
```bash
# 集群环境变量配置
CLICKHOUSE_HOST=clickhouse-primary.example.com
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=cluster_user
CLICKHOUSE_PASSWORD=cluster_password
CLICKHOUSE_DATABASE=cluster_db
CLICKHOUSE_ALT_HOSTS=clickhouse-node2.example.com:9000,clickhouse-node3.example.com:9000
CLICKHOUSE_CONN_OPEN_STRATEGY=time
CLICKHOUSE_SETTINGS=distributed_product_mode=global,insert_distributed_sync=1
```

### 5. 集群使用示例

#### 5.1 连接到集群
```go
// 集群配置
config := &dbtypes.DbConfig{
    Name:    "clickhouse_cluster",
    Enabled: true,
    Driver:  dbtypes.DriverClickHouse,
    Connection: dbtypes.ConnectionConfig{
        Host:     "ch-node1.example.com",
        Port:     9000,
        Username: "cluster_user",
        Password: "cluster_pass",
        Database: "cluster_db",
        
        // 集群参数
        ClickHouseAltHosts:          "ch-node2.example.com:9000,ch-node3.example.com:9000",
        ClickHouseConnOpenStrategy:  "time",
        ClickHouseSettings:          "distributed_product_mode=global",
    },
}

// 连接集群
db, err := database.Open(config)
if err != nil {
    log.Fatal("集群连接失败:", err)
}
defer db.Close()
```

#### 5.2 分布式表查询
```go
// 查询分布式表
type DistributedStats struct {
    Shard    string `db:"shard"`
    Count    uint64 `db:"count"`
    AvgValue float64 `db:"avg_value"`
}

var stats []DistributedStats
err = db.Query(ctx, &stats, `
    SELECT 
        _shard_num as shard,
        COUNT(*) as count,
        AVG(value) as avg_value
    FROM distributed_events 
    WHERE timestamp >= ? 
    GROUP BY _shard_num
    ORDER BY shard`,
    []interface{}{time.Now().Add(-24*time.Hour)}, true)
```

#### 5.3 集群批量插入
```go
// 向分布式表批量插入数据
events := []Event{
    {UserID: 123, Action: "click", Timestamp: time.Now(), Value: 1.0},
    {UserID: 124, Action: "view", Timestamp: time.Now(), Value: 2.0},
    // ... 更多数据
}

// 批量插入到分布式表，会自动分布到各个分片
rowsAffected, err := db.BatchInsert(ctx, "distributed_events", events, true)
if err != nil {
    log.Fatal("集群批量插入失败:", err)
}
log.Printf("插入 %d 条记录到集群", rowsAffected)
```

### 6. 集群性能优化

#### 6.1 连接池优化
```yaml
pool:
  max_open_conns: 200        # 集群环境可以设置更大
  max_idle_conns: 100        # 保持更多空闲连接
  conn_max_lifetime: 7200    # 更长的连接生命周期
  conn_max_idle_time: 3600   # 集群环境连接复用更重要
```

#### 6.2 查询优化设置
```yaml
clickhouse_settings: "
  distributed_product_mode=global,
  insert_distributed_sync=1,
  max_distributed_connections=16,
  distributed_directory_monitor_batch_inserts=1,
  distributed_directory_monitor_sleep_time_ms=50,
  max_threads=8,
  max_memory_usage=20000000000
"
```

### 7. 集群监控和故障处理

#### 7.1 集群健康检查
```go
// 检查集群连接状态
func checkClusterHealth(db database.Database) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    // 测试集群连接
    if err := db.Ping(ctx); err != nil {
        return fmt.Errorf("集群连接失败: %w", err)
    }
    
    // 查询集群信息
    var clusterInfo []struct {
        Host     string `db:"host"`
        Port     uint16 `db:"port"`
        IsLocal  uint8  `db:"is_local"`
    }
    
    err := db.Query(ctx, &clusterInfo, 
        "SELECT host_name as host, port, is_local FROM system.clusters WHERE cluster = 'default'", 
        []interface{}{}, true)
    if err != nil {
        return fmt.Errorf("查询集群信息失败: %w", err)
    }
    
    log.Printf("集群节点数: %d", len(clusterInfo))
    return nil
}
```

#### 7.2 故障转移测试
```go
// 测试故障转移功能
func testFailover(db database.Database) {
    // 模拟网络故障后的查询
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    var count int64
    err := db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM system.tables", []interface{}{}, true)
    if err != nil {
        log.Printf("故障转移测试失败: %v", err)
    } else {
        log.Printf("故障转移成功，查询结果: %d", count)
    }
}
```

### 8. 集群配置最佳实践

1. **节点配置**：
   - 使用专用的集群用户和密码
   - 配置适当的超时时间（集群查询通常较慢）
   - 启用SSL/TLS（生产环境）

2. **负载均衡**：
   - 读多写少场景：使用 `time` 策略
   - 主从架构：使用 `in_order` 策略
   - 分片集群：使用 `random` 策略

3. **连接池**：
   - 集群环境建议更大的连接池
   - 更长的连接生命周期
   - 监控连接使用情况

4. **监控指标**：
   - 各节点响应时间
   - 故障转移频率
   - 分布式查询性能
   - 连接池使用率

## ClickHouse特有的考虑

### 1. 表引擎选择

ClickHouse支持多种表引擎，选择合适的引擎很重要：

```sql
-- MergeTree：最常用的引擎，适合时间序列数据
CREATE TABLE events (
    id UInt64,
    user_id UInt64,
    action String,
    timestamp DateTime,
    value Float64
) ENGINE = MergeTree()
PARTITION BY toDate(timestamp)
ORDER BY (user_id, timestamp);

-- ReplacingMergeTree：支持去重，类似UPDATE操作
CREATE TABLE user_profiles (
    user_id UInt64,
    name String,
    email String,
    updated_at DateTime
) ENGINE = ReplacingMergeTree(updated_at)
ORDER BY user_id;

-- SummingMergeTree：自动聚合数值字段
CREATE TABLE user_stats (
    user_id UInt64,
    date Date,
    clicks UInt64,
    views UInt64
) ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (user_id, date);
```

### 2. UPDATE和DELETE的替代方案

由于ClickHouse的UPDATE/DELETE支持有限，建议使用以下方案：

```go
// 方案1：使用ReplacingMergeTree + INSERT
// 表结构使用ReplacingMergeTree引擎
type UserProfile struct {
    UserID    uint64    `db:"user_id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    UpdatedAt time.Time `db:"updated_at"`
}

// "更新"实际上是插入新记录，引擎会自动去重
profile := UserProfile{
    UserID:    123,
    Name:      "New Name",
    Email:     "new@email.com",
    UpdatedAt: time.Now(),
}
db.Insert(ctx, "user_profiles", profile, true)

// 方案2：使用INSERT INTO ... SELECT
_, err = db.Exec(ctx, `
    INSERT INTO user_profiles_new 
    SELECT user_id, ?, email, ? 
    FROM user_profiles 
    WHERE user_id = ?`,
    []interface{}{"New Name", time.Now(), 123}, true)

// 方案3：对于"删除"，添加删除标记字段
type EventWithDelete struct {
    ID        uint64    `db:"id"`
    UserID    uint64    `db:"user_id"`
    Action    string    `db:"action"`
    Timestamp time.Time `db:"timestamp"`
    IsDeleted uint8     `db:"is_deleted"` // 0=正常, 1=已删除
}

// "删除"操作
_, err = db.Exec(ctx, 
    "INSERT INTO events SELECT id, user_id, action, timestamp, 1 FROM events WHERE id = ?",
    []interface{}{eventID}, true)
```

### 3. 数据类型映射

ClickHouse与Go类型的映射关系：

| ClickHouse类型 | Go类型 | 说明 |
|---------------|--------|------|
| UInt8, UInt16, UInt32, UInt64 | uint8, uint16, uint32, uint64 | 无符号整数 |
| Int8, Int16, Int32, Int64 | int8, int16, int32, int64 | 有符号整数 |
| Float32, Float64 | float32, float64 | 浮点数 |
| String | string | 字符串 |
| Date | time.Time | 日期 |
| DateTime | time.Time | 日期时间 |
| Array(T) | []T | 数组 |
| Nullable(T) | *T | 可空类型 |

### 4. 性能优化建议

1. **批量插入优化**：
   ```go
   // 推荐：使用批量插入
   db.BatchInsert(ctx, "events", largeEventSlice, true)
   
   // 不推荐：逐条插入
   for _, event := range events {
       db.Insert(ctx, "events", event, true)
   }
   ```

2. **查询优化**：
   ```go
   // 利用分区字段
   db.Query(ctx, &results, `
       SELECT * FROM events 
       WHERE timestamp >= ? AND timestamp < ?  -- 利用分区
       AND user_id = ?`,                       -- 利用排序键
       []interface{}{startTime, endTime, userID}, true)
   
   // 使用ClickHouse特有函数
   db.Query(ctx, &results, `
       SELECT 
           user_id,
           uniq(session_id) as unique_sessions,
           quantile(0.95)(response_time) as p95_response_time
       FROM events 
       WHERE toDate(timestamp) = today()
       GROUP BY user_id`,
       []interface{}{}, true)
   ```

3. **连接池设置**：
   ```yaml
   pool:
     max_open_conns: 50     # ClickHouse可以处理较多并发连接
     max_idle_conns: 25
     conn_max_lifetime: 3600
   ```

## 注意事项

### ClickHouse特有的限制

1. **事务支持有限**：ClickHouse主要设计用于OLAP场景，事务支持有限。

2. **UPDATE/DELETE性能较差**：建议使用ReplacingMergeTree等引擎的特性替代。

3. **实时性**：ClickHouse是最终一致性的，新插入的数据可能有短暂延迟才能查询到。

4. **JOIN性能**：复杂JOIN操作性能一般，建议预聚合或使用字典。

### 最佳实践

1. **表设计**：
   - 合理选择分区字段（通常是时间字段）
   - 选择合适的排序键（ORDER BY）
   - 使用合适的表引擎

2. **查询模式**：
   - 优先使用批量插入
   - 查询时充分利用分区和索引
   - 使用ClickHouse特有的聚合函数

3. **监控指标**：
   - 查询响应时间
   - 插入吞吐量
   - 分区大小和数量
   - 内存使用情况

## 错误处理

```go
import "gateway/pkg/database"

// 常见错误处理
_, err := db.Insert(ctx, "events", event, true)
if err != nil {
    switch {
    case errors.Is(err, database.ErrConnection):
        log.Println("连接错误，请检查ClickHouse服务状态")
    case errors.Is(err, database.ErrInvalidQuery):
        log.Println("查询语法错误")
    case strings.Contains(err.Error(), "timeout"):
        log.Println("查询超时，请优化查询或增加超时时间")
    default:
        log.Printf("未知错误: %v", err)
    }
}
```

## 测试

运行测试（需要ClickHouse数据库环境）：

```bash
# 启动ClickHouse容器
docker run -d --name clickhouse-server \
  -p 8123:8123 -p 9000:9000 \
  clickhouse/clickhouse-server

# 等待服务启动
sleep 10

# 运行测试
go test ./pkg/database/clickhouse/... -v

# 清理
docker stop clickhouse-server && docker rm clickhouse-server
```

## 开发和调试

```go
// 启用详细日志
config.Log.Enable = true
config.Log.SlowThreshold = 100 // 100ms

// 获取底层连接进行调试
if chDB, ok := db.(*clickhouse.ClickHouse); ok {
    sqlDB := chDB.DB()
    // 进行底层操作
}
```

## 相关资源

- [ClickHouse官方文档](https://clickhouse.com/docs/)
- [ClickHouse Go驱动](https://github.com/ClickHouse/clickhouse-go)
- [ClickHouse SQL语法](https://clickhouse.com/docs/en/sql-reference/)
- [表引擎选择指南](https://clickhouse.com/docs/en/engines/table-engines/)

## 版本兼容性

- ClickHouse Server: 20.3+
- Go: 1.18+
- ClickHouse Go Driver: v1.5+ 或 v2.0+

## 迁移指南

从其他数据库迁移到ClickHouse时的注意事项：

1. **从MySQL迁移**：主要调整批量插入策略和查询模式
2. **从PostgreSQL迁移**：注意数组类型和JSON处理的差异
3. **从MongoDB迁移**：需要重新设计表结构，充分利用列式存储优势 