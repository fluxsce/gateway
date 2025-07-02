# SQLite数据库实现

## 概述

本模块提供了SQLite数据库的完整实现，参考MySQL实现的架构，符合统一的`database.Database`接口规范。

## 特性

- ✅ **统一接口** - 完全符合`database.Database`接口规范
- ✅ **轻量级** - 适合开发环境和小型应用
- ✅ **零配置** - 支持内存数据库和文件数据库
- ✅ **事务支持** - 完整的ACID事务支持
- ✅ **连接池** - 自动管理连接池
- ✅ **日志记录** - 支持SQL执行日志和慢查询检测
- ✅ **结构体映射** - 自动映射Go结构体到数据库表
- ✅ **批量操作** - 支持批量插入等高效操作
- ✅ **并发安全** - WAL模式支持多读单写

## 依赖

```bash
go get github.com/mattn/go-sqlite3
```

**注意：** SQLite驱动需要CGO支持，请确保：
- Windows: 安装GCC (推荐使用TDM-GCC或MinGW)
- Linux/MacOS: 通常已内置GCC支持

## 使用示例

### 基本配置

```go
import (
    "gohub/pkg/database"
    "gohub/pkg/database/dbtypes"
    _ "gohub/pkg/database/sqlite" // 导入SQLite实现
)

// 内存数据库配置
config := &dbtypes.DbConfig{
    Name:    "sqlite_memory",
    Enabled: true,
    Driver:  dbtypes.DriverSQLite,
    DSN:     ":memory:",
    Pool: dbtypes.PoolConfig{
        MaxOpenConns:    10,
        MaxIdleConns:    5,
        ConnMaxLifetime: 3600,
        ConnMaxIdleTime: 1800,
    },
}

// 文件数据库配置
config := &dbtypes.DbConfig{
    Name:    "sqlite_file",
    Enabled: true,
    Driver:  dbtypes.DriverSQLite,
    DSN:     "data.db",
    // 或使用完整路径和参数
    // DSN: "file:data.db?cache=shared&mode=rwc&_journal_mode=WAL",
}

// 连接数据库
db, err := database.Open(config)
if err != nil {
    log.Fatal(err)
}
defer db.Close()
```

### 基本操作

```go
type User struct {
    ID       int64     `db:"id"`
    Name     string    `db:"name"`
    Email    string    `db:"email"`
    Age      int       `db:"age"`
    CreateAt time.Time `db:"create_at"`
}

ctx := context.Background()

// 插入
user := User{Name: "张三", Email: "zhangsan@example.com", Age: 25}
id, err := db.Insert(ctx, "users", user, true)

// 查询单条
var user User
err = db.QueryOne(ctx, &user, "SELECT * FROM users WHERE id = ?", []interface{}{id}, true)

// 查询多条
var users []User
err = db.Query(ctx, &users, "SELECT * FROM users WHERE age > ?", []interface{}{20}, true)

// 更新
updateData := User{Name: "李四", Age: 30}
affected, err := db.Update(ctx, "users", updateData, "id = ?", []interface{}{id}, true)

// 删除
affected, err := db.Delete(ctx, "users", "id = ?", []interface{}{id}, true)

// 批量插入
users := []User{
    {Name: "用户1", Email: "user1@example.com", Age: 25},
    {Name: "用户2", Email: "user2@example.com", Age: 26},
}
affected, err := db.BatchInsert(ctx, "users", users, true)
```

### 事务操作

```go
// 自动事务管理
err = db.InTx(ctx, nil, func() error {
    _, err := db.Insert(ctx, "users", user1, false)
    if err != nil {
        return err
    }
    _, err = db.Insert(ctx, "users", user2, false)
    return err
})

// 手动事务管理
err = db.BeginTx(ctx, nil)
if err != nil {
    return err
}

_, err = db.Insert(ctx, "users", user, false)
if err != nil {
    db.Rollback()
    return err
}

err = db.Commit()
```

## 配置说明

### DSN格式

```
# 内存数据库
:memory:

# 文件数据库
data.db
./data/app.db
/path/to/database.db

# 带参数的文件数据库
file:data.db?cache=shared&mode=rwc&_journal_mode=WAL&_busy_timeout=5000

# 常用参数
cache=shared        # 共享缓存
mode=rwc           # 读写创建模式
_journal_mode=WAL  # WAL日志模式
_busy_timeout=5000 # 忙等待超时(毫秒)
_foreign_keys=1    # 启用外键约束
```

### 连接池配置

```go
Pool: dbtypes.PoolConfig{
    MaxOpenConns:    10,   // 最大连接数 (SQLite建议较小值)
    MaxIdleConns:    5,    // 最大空闲连接数
    ConnMaxLifetime: 3600, // 连接最大生命周期(秒)
    ConnMaxIdleTime: 1800, // 连接最大空闲时间(秒)
}
```

### 日志配置

```go
Log: dbtypes.LogConfig{
    Enable:        true, // 启用日志
    SlowThreshold: 200,  // 慢查询阈值(毫秒)
}
```

## 性能优化

### 自动优化配置

SQLite实现会自动设置以下优化参数：

- `PRAGMA journal_mode = WAL` - 启用WAL模式支持并发
- `PRAGMA synchronous = NORMAL` - 平衡性能和安全性
- `PRAGMA cache_size = -2000` - 设置2MB页面缓存
- `PRAGMA foreign_keys = ON` - 启用外键约束

### 最佳实践

1. **连接池**: SQLite建议使用较小的连接池 (MaxOpenConns: 10)
2. **事务**: 批量操作时使用事务提高性能
3. **索引**: 为经常查询的字段创建索引
4. **WAL模式**: 已自动启用，支持并发读写
5. **内存数据库**: 适合测试和临时数据存储

## 错误处理

常见错误及解决方法：

```go
// 记录不存在
if err == database.ErrRecordNotFound {
    // 处理记录不存在的情况
}

// 重复键
if err == database.ErrDuplicateKey {
    // 处理重复键错误
}

// 连接错误
if err == database.ErrConnection {
    // 处理连接错误
}
```

## 注意事项

1. **CGO依赖**: SQLite驱动需要CGO支持，确保编译环境正确配置
2. **文件权限**: 确保应用有足够权限访问数据库文件
3. **并发**: WAL模式支持多读单写，但仍需注意并发控制
4. **内存数据库**: 进程结束后数据会丢失，仅适合临时存储
5. **文件锁**: 同一文件数据库只能被一个进程独占访问

## 测试

创建测试文件验证功能：

```go
package main

import (
    "context"
    "log"
    "gohub/pkg/database"
    "gohub/pkg/database/dbtypes"
    _ "gohub/pkg/database/sqlite"
)

func main() {
    config := &dbtypes.DbConfig{
        Driver: dbtypes.DriverSQLite,
        DSN:    ":memory:",
    }
    
    db, err := database.Open(config)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // 测试基本功能
    ctx := context.Background()
    _, err = db.Exec(ctx, "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)", nil, true)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("SQLite实现测试通过!")
}
```

编译运行：
```bash
# 确保启用CGO
export CGO_ENABLED=1  # Linux/MacOS
$env:CGO_ENABLED="1"  # Windows PowerShell

go run test.go
```

## 相关文档

- [SQLite官方文档](https://www.sqlite.org/docs.html)
- [go-sqlite3驱动文档](https://github.com/mattn/go-sqlite3)
- [数据库接口规范](../database.go) 