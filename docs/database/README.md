# GoHub 数据库组件

GoHub 数据库组件提供了一个统一的数据库操作接口，支持事务管理，并提供了灵活的配置选项。目前实现了 MySQL 驱动。

## 特性

- 统一的数据库操作接口
- 灵活的事务管理
- 可配置的连接池
- 丰富的查询和执行选项
- 类型安全的模型操作

## 架构

数据库组件的架构主要包括以下几个部分：

### 接口层

- `DBExecutor`: 基础数据库操作接口，包含 `Exec`、`Query`、`QueryOne`、`Insert`、`Update` 和 `Delete` 等方法
- `Transaction`: 继承于 `DBExecutor`，增加了 `Commit` 和 `Rollback` 方法
- `Database`: 继承于 `DBExecutor`，增加了连接管理和事务控制方法，以及带选项的操作方法

### 实现层

- MySQL 实现: 提供 MySQL 数据库的具体实现

### 配置和选项

- `DbConfig`: 数据库配置，包括驱动类型、连接信息、连接池配置等
- `ExecOptions`: 执行选项，控制执行操作的行为
- `QueryOptions`: 查询选项，控制查询操作的行为
- `TxOptions`: 事务选项，控制事务的隔离级别和只读状态

## 使用示例

### 初始化数据库连接

```go
config := &database.DbConfig{
    Driver:             database.DriverMySQL,
    DSN:                "user:pass@tcp(localhost:3306)/database?charset=utf8mb4&parseTime=True",
    MaxOpenConns:       100,
    MaxIdleConns:       10,
    ConnMaxLifetime:    time.Hour,
    DefaultLogSQL:      true,
    DefaultUseTransaction: false,
}

db, err := database.Open(config)
if err != nil {
    panic(err)
}
defer db.Close()
```

### 基本查询操作

```go
// 定义模型
type User struct {
    ID   int    `db:"id"`
    Name string `db:"name"`
}

func (u User) TableName() string {
    return "users"
}

func (u User) PrimaryKey() string {
    return "id"
}

// 查询多条记录
var users []User
err := db.Query(context.Background(), &users, "SELECT * FROM users WHERE age > ?", 18)

// 查询单条记录
var user User
err := db.QueryOne(context.Background(), &user, "SELECT * FROM users WHERE id = ?", 1)
```

### 执行 SQL

```go
// 执行 SQL 语句
affected, err := db.Exec(context.Background(), "UPDATE users SET name = ? WHERE id = ?", "新名字", 1)
```

### 使用事务

```go
// 显式事务
tx, err := db.BeginTx(context.Background())
if err != nil {
    return err
}

// 使用 defer 确保事务结束
defer func() {
    if p := recover(); p != nil {
        tx.Rollback()
        panic(p)
    }
}()

// 执行事务操作
_, err = tx.Insert(context.Background(), "users", &User{Name: "用户1"})
if err != nil {
    tx.Rollback()
    return err
}

_, err = tx.Insert(context.Background(), "users", &User{Name: "用户2"})
if err != nil {
    tx.Rollback()
    return err
}

// 提交事务
return tx.Commit()
```

### 使用 WithTx 辅助函数

```go
err := db.WithTx(nil, func(tx database.Transaction) error {
    // 在事务中执行操作
    _, err := tx.Insert(context.Background(), "users", &User{Name: "用户1"})
    if err != nil {
        return err // 自动回滚
    }
    
    _, err = tx.Insert(context.Background(), "users", &User{Name: "用户2"})
    if err != nil {
        return err // 自动回滚
    }
    
    return nil // 自动提交
})
```

### 使用选项

```go
// 带事务选项的执行
affected, err := db.ExecWithOptions(
    context.Background(),
    "UPDATE users SET name = ? WHERE id = ?",
    []interface{}{"新名字", 1},
    database.WithTransaction(true),
)

// 带隔离级别的事务
tx, err := db.BeginTx(
    context.Background(),
    database.WithIsolation(database.IsolationRepeatableRead),
    database.WithReadOnly(true),
)
```

## 测试

数据库组件提供了完整的测试套件：

- `mysql_test.go`: MySQL 实现的功能测试
- `options_test.go`: 配置和选项的单元测试
- `interface_test.go`: 接口行为的单元测试

通过以下命令运行测试：

```bash
go test -v ./test/database/...
```

## 扩展

要添加新的数据库驱动，只需实现 `Database` 接口并在初始化时通过 `database.Register` 函数注册即可。 