# Oracle 数据库实现

## 概述

本包提供了 Oracle 数据库的完整实现，符合 `gateway/pkg/database` 包定义的统一数据库接口。支持 Oracle 11g、12c、18c、19c、21c 等版本。

## 功能特性

### 核心功能
- ✅ 统一的数据库接口实现 - 符合 `database.Database` 接口规范
- ✅ 灵活的事务管理 - 支持不同隔离级别和可定制的事务选项
- ✅ 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
- ✅ 智能日志记录 - 支持慢查询检测和 SQL 执行日志
- ✅ 结构体映射 - 自动将 Go 结构体与数据库表映射
- ✅ 带选项的操作 - 每个数据库操作都有带选项的版本，支持自定义执行行为
- ✅ 活跃事务追踪 - 通过内部映射跟踪所有活跃事务

### Oracle 特性支持
- ✅ Oracle 占位符格式 (`:1`, `:2`, `:3`...) 自动转换
- ✅ Oracle 批量插入 (`INSERT ALL` 语法)
- ✅ Oracle 分页查询支持 (`OFFSET ... ROWS FETCH NEXT ... ROWS ONLY`)
- ✅ Oracle 序列支持
- ✅ Oracle 过程调用支持
- ✅ Oracle 大对象 (LOB) 支持

## 安装和依赖

### 1. Oracle 客户端安装

在使用本包之前，需要安装 Oracle 客户端：

**Linux/macOS:**
```bash
# 下载 Oracle Instant Client
wget https://download.oracle.com/otn_software/linux/instantclient/instantclient-basic-linux.x64-21.1.0.0.0.zip
unzip instantclient-basic-linux.x64-21.1.0.0.0.zip
export LD_LIBRARY_PATH=/path/to/instantclient_21_1:$LD_LIBRARY_PATH
```

**Windows:**
```powershell
# 下载并安装 Oracle Instant Client
# 设置环境变量 PATH 包含 Oracle Instant Client 目录
```

### 2. Go 依赖

本包使用 `github.com/godror/godror` 作为 Oracle 驱动：

```bash
go get github.com/godror/godror@latest
```

## 快速开始

### 1. 导入包

```go
import (
    "gateway/pkg/database"
    _ "gateway/pkg/database/alldriver" // 导入所有驱动，包括 Oracle
)
```

### 2. 创建连接

```go
config := &database.DbConfig{
    Name:   "oracle_main",
    Driver: "oracle",
    DSN:    "oracle://scott:tiger@localhost:1521/XEPDB1",
    Pool: database.PoolConfig{
        MaxOpenConns:    50,
        MaxIdleConns:    25,
        ConnMaxLifetime: 3600, // 1小时
        ConnMaxIdleTime: 1800, // 30分钟
    },
    Log: database.LogConfig{
        Enable:        true,
        SlowThreshold: 200, // 200毫秒
    },
}

db, err := database.Open(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### 3. 基本操作

```go
ctx := context.Background()

// 插入记录
type User struct {
    ID   int64  `db:"id"`
    Name string `db:"name"`
    Age  int    `db:"age"`
}

user := User{Name: "张三", Age: 25}
id, err := db.Insert(ctx, "users", user, true)

// 查询记录
var users []User
err = db.Query(ctx, &users, "SELECT id, name, age FROM users WHERE age > :1", []interface{}{20}, true)

// 查询单条记录
var user User
err = db.QueryOne(ctx, &user, "SELECT id, name, age FROM users WHERE id = :1", []interface{}{1}, true)

// 更新记录
updateData := User{Name: "李四", Age: 30}
affected, err := db.Update(ctx, "users", updateData, "id = :1", []interface{}{1}, true)

// 删除记录
affected, err = db.Delete(ctx, "users", "id = :1", []interface{}{1}, true)
```

### 4. 事务操作

```go
err = db.InTx(ctx, nil, func() error {
    // 在事务中执行多个操作
    _, err := db.Insert(ctx, "users", user1, false)
    if err != nil {
        return err
    }
    
    _, err = db.Insert(ctx, "users", user2, false)
    if err != nil {
        return err
    }
    
    return nil // 提交事务
})
```

## 配置说明

### DSN 格式

支持多种 DSN 格式：

```yaml
# 标准格式
dsn: "oracle://username:password@host:port/service_name"

# 简化格式
dsn: "username/password@host:port/service_name"

# 完整格式
dsn: "username/password@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=host)(PORT=port))(CONNECT_DATA=(SERVICE_NAME=service_name)))"

# 使用 SID
dsn: "oracle://username:password@host:port/XE?sid=true"
```

### 完整配置示例

```yaml
database:
  default: "oracle_main"
  connections:
    oracle_main:
      enabled: true
      driver: "oracle"
      dsn: "oracle://scott:tiger@localhost:1521/XEPDB1"
      
      pool:
        max_open_conns: 50
        max_idle_conns: 25
        conn_max_lifetime: 3600
        conn_max_idle_time: 1800
        
      log:
        enable: true
        slow_threshold: 200
        
      transaction:
        default_use: false
```

## 注意事项

### Oracle 特有的考虑

1. **占位符格式**: Oracle 使用 `:1`, `:2`, `:3`... 格式的占位符，本实现会自动从 `?` 格式转换。

2. **自增主键**: Oracle 使用序列实现自增，需要手动创建序列：
   ```sql
   CREATE SEQUENCE users_seq START WITH 1 INCREMENT BY 1;
   ```

3. **批量插入**: 使用 Oracle 的 `INSERT ALL` 语法进行批量插入。

4. **大小写敏感**: Oracle 默认将未加引号的标识符转换为大写。

5. **数据类型映射**:
   - `VARCHAR2` -> `string`
   - `NUMBER` -> `int64`, `float64`
   - `DATE` -> `time.Time`
   - `CLOB` -> `string`
   - `BLOB` -> `[]byte`

### 性能建议

1. **连接池设置**:
   - 生产环境建议 `max_open_conns` 设置为 50-100
   - `max_idle_conns` 设置为 `max_open_conns` 的 50%
   - `conn_max_lifetime` 建议 1-4 小时
   - `conn_max_idle_time` 建议 30 分钟-2 小时

2. **SQL 优化**:
   - 使用绑定变量避免硬解析
   - 合理使用索引
   - 批量操作时使用 `BatchInsert`

3. **事务管理**:
   - 保持事务尽可能短
   - 避免长时间运行的事务
   - 合理设置隔离级别

## 测试

运行测试（需要 Oracle 数据库环境）：

```bash
# 运行驱动注册测试（无需实际数据库）
go test -v ./test/database -run TestOracleDriverRegistration

# 运行连接测试（需要 Oracle 数据库）
go test -v ./test/database -run TestOracleConnection

# 运行所有测试（需要 Oracle 数据库）
go test -v ./test/database -run TestOracle
```

## 故障排除

### 常见问题

1. **连接失败**:
   - 检查 Oracle 服务器是否运行
   - 验证防火墙设置（端口 1521）
   - 确认 SERVICE_NAME 或 SID 正确

2. **客户端库问题**:
   - 确保 Oracle Instant Client 已正确安装
   - 检查环境变量 `LD_LIBRARY_PATH`（Linux/macOS）或 `PATH`（Windows）

3. **编码问题**:
   - 确保数据库字符集设置正确
   - 使用 UTF-8 编码

4. **性能问题**:
   - 调整连接池参数
   - 启用 SQL 日志分析慢查询
   - 检查数据库统计信息和执行计划

### 日志分析

启用详细日志来诊断问题：

```yaml
log:
  enable: true
  slow_threshold: 100  # 降低阈值以捕获更多慢查询
```

## 许可证

本实现遵循项目的整体许可证。Oracle 是 Oracle Corporation 的注册商标。 