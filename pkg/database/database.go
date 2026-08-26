// Package database 定义统一的数据库访问接口与连接工厂。
// 业务通过 Database（Conn + SQL + Records + TxManager）访问关系库；具体驱动在子包注册。
package database

import (
	"context"
	"errors"
	"fmt"
	"gateway/pkg/config"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dsn"
	"gateway/pkg/security"
	"sync"
)

// 定义通用数据库错误
var (
	// ErrRecordNotFound 记录未找到错误
	ErrRecordNotFound = errors.New("record not found")

	// ErrDuplicateKey 重复键错误
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrConnection 连接错误
	ErrConnection = errors.New("database connection error")

	// ErrTransaction 事务错误
	ErrTransaction = errors.New("transaction error")

	// ErrInvalidQuery 无效查询错误
	ErrInvalidQuery = errors.New("invalid query")

	// ErrConfigNotFound 配置未找到错误
	ErrConfigNotFound = errors.New("database config not found")
)

// IsRecordNotFound 是否为未找到记录。兼容历史代码的相等比较与包装后的 errors.Is。
func IsRecordNotFound(err error) bool {
	return errors.Is(err, ErrRecordNotFound)
}

// IsDuplicateKey 是否为唯一约束冲突。
func IsDuplicateKey(err error) bool {
	return errors.Is(err, ErrDuplicateKey)
}

// 数据库工厂映射及缓存
var (
	// dbCreators 存储注册的数据库驱动创建函数
	dbCreators = make(map[string]DriverCreator)

	// creatorMu 保护驱动注册表
	creatorMu = sync.RWMutex{}

	// dbConnections 缓存已创建的数据库连接实例
	dbConnections = make(map[string]Database)

	// connMutex 用于保护数据库连接缓存的互斥锁
	connMutex = sync.RWMutex{}
)

// 支持的数据库类型常量
const (
	DriverMySQL      = dbtypes.DriverMySQL
	DriverPostgreSQL = dbtypes.DriverPostgreSQL
	DriverSQLite     = dbtypes.DriverSQLite
	DriverOracle     = dbtypes.DriverOracle
	DriverSQLServer  = dbtypes.DriverSQLServer
	DriverClickHouse = dbtypes.DriverClickHouse
)

// DbConfig 数据库配置类型别名
type DbConfig = dbtypes.DbConfig

// IsolationLevel 事务隔离级别
// 定义数据库事务的隔离级别常量
// 不同的隔离级别提供不同程度的并发控制和数据一致性保证
type IsolationLevel int

const (
	// IsolationDefault 默认隔离级别
	// 使用数据库默认的隔离级别，通常为读已提交
	IsolationDefault IsolationLevel = 0

	// IsolationReadUncommitted 读未提交
	// 最低隔离级别，允许读取未提交的数据
	// 可能出现脏读、不可重复读、幻读问题
	IsolationReadUncommitted IsolationLevel = 1

	// IsolationReadCommitted 读已提交
	// 只允许读取已提交的数据，避免脏读
	// 可能出现不可重复读、幻读问题
	IsolationReadCommitted IsolationLevel = 2

	// IsolationRepeatableRead 可重复读
	// 保证在同一事务中多次读取同样记录的结果是一致的
	// 可能出现幻读问题
	IsolationRepeatableRead IsolationLevel = 3

	// IsolationSerializable 串行化
	// 最高隔离级别，完全避免脏读、不可重复读、幻读
	// 提供最强的数据一致性保证，但并发性能最低
	IsolationSerializable IsolationLevel = 4
)

// TxOptions 事务选项
// 定义事务的配置参数，包括隔离级别和访问模式
type TxOptions struct {
	// Isolation 事务隔离级别
	// 控制事务在并发环境下的数据可见性和一致性
	Isolation IsolationLevel

	// ReadOnly 是否只读事务
	// true: 只读事务，不允许修改数据，可以提高性能
	// false: 读写事务，允许查询和修改数据
	ReadOnly bool
}

// Conn 管理连接生命周期。健康检查或只探测连通性时可以只依赖本接口。
type Conn interface {
	// Connect 根据配置建立数据库连接，包括连接池、日志等。
	// 失败时不应留下半开的底层连接。
	// 参数:
	//   config: 数据库配置，包含 DSN 或可生成 DSN 的连接信息、池设置、日志
	// 返回:
	//   error: 连接失败时返回错误信息
	Connect(config *DbConfig) error

	// Close 关闭数据库连接并释放资源。
	// 使用上下文绑定事务时，Close 不会自动回滚，调用方应先 Commit 或 Rollback。
	// 返回:
	//   error: 关闭失败时返回错误信息
	Close() error

	// Ping 向数据库发送探测请求，验证连接是否可用。
	// 参数:
	//   ctx: 用于超时和取消
	// 返回:
	//   error: 连接异常时返回错误信息
	Ping(ctx context.Context) error

	// GetDriver 返回逻辑驱动类型，如 mysql、sqlite、oracle、sqlserver、clickhouse。
	GetDriver() string

	// GetName 返回当前连接名称；配置为空时返回空字符串。
	GetName() string

	// SetName 设置连接名称，加载 YAML 配置时把连接名写回实例。
	SetName(name string)
}

// SQL 执行原始 SQL。分页、复杂查询、手写语句走这里。
// autoCommit 为 true 时直接用连接并自动提交；为 false 时使用 ctx 中的事务，需先 BeginTx 或 InTx。
type SQL interface {
	// Exec 执行 INSERT、UPDATE、DELETE 等不返回结果集的语句。
	// 占位符统一写 ?，执行前由方言改写成各库格式。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   query: SQL 语句，可含占位符
	//   args: 占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   int64: 受影响行数
	//   error: 失败时返回，唯一冲突等会包装为 ErrDuplicateKey
	Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error)

	// Query 执行 SELECT 并将多行扫描到 dest 切片。
	// dest 必须是切片指针，元素为结构体或结构体指针，字段通过 db tag 映射列名。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   dest: 目标切片指针
	//   query: SELECT 语句，可含占位符
	//   args: 占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   error: 查询或扫描失败时返回错误信息
	Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error

	// QueryOne 执行 SELECT 并将单行扫描到 dest 结构体。
	// 查不到记录时返回 ErrRecordNotFound，可用 err == ErrRecordNotFound 或 IsRecordNotFound 判断。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   dest: 目标结构体指针
	//   query: SELECT 语句，可含占位符
	//   args: 占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   error: 查询失败、扫描失败或记录不存在
	QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error

	// QueryEach 按游标逐行扫描到 dest 并回调，不把整结果集载入内存。
	// dest 必须是结构体指针，每行复用同一块内存：回调返回后即可再次扫描覆盖，
	// 调用方若要保留数据必须在回调内拷贝。回调返回 error 或 ctx 取消时停止。
	// 无论成功、失败还是中途退出，都会 Close 结果集并归还连接，不会泄漏游标。
	// 导出期间占用连接池中的一条连接，直到游标结束。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   dest: 每行扫描复用的结构体指针
	//   query: SELECT 语句，可含占位符
	//   args: 占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	//   fn: 每扫描完一行后调用，返回 error 则中止游标
	// 返回:
	//   error: 查询失败、扫描失败或回调失败
	QueryEach(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool, fn func() error) error
}

// Records 按结构体映射做增删改和批量操作，开箱即用，保留在公共 API 上。
// 字段通过 db tag 映射列名；占位符统一写 ?，执行前由方言改写。
type Records interface {
	// Insert 根据数据结构体构建 INSERT 并执行。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   data: 要插入的结构体，字段通过 db tag 映射到列
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   int64: 自增 ID（库不支持时为 0）
	//   error: 插入失败时返回，唯一冲突会包装为 ErrDuplicateKey
	Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error)

	// Update 根据数据结构体和 WHERE 构建 UPDATE 并执行。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   data: 含更新数据的结构体，字段通过 db tag 映射到列
	//   where: WHERE 条件，可含占位符；为空则不加 WHERE
	//   args: WHERE 中占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	//   skipZero: true 跳过零值字段；false 包含零值（用于清空字段）
	// 返回:
	//   int64: 受影响行数
	//   error: 更新失败时返回错误信息
	Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool, skipZero bool) (int64, error)

	// Delete 根据 WHERE 构建 DELETE 并执行。
	// ClickHouse 等库由方言改写成 ALTER TABLE ... DELETE。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   where: WHERE 条件，可含占位符；为空则不加 WHERE
	//   args: WHERE 中占位符对应的参数
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   int64: 受影响行数
	//   error: 删除失败时返回错误信息
	Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error)

	// BatchInsert 批量插入切片中的多条结构体。
	// 预编译一条 INSERT 后在事务中循环执行，任一条失败则回滚本批。
	// 适合中小批量；超大批量建议业务层分批调用。ClickHouse 会按数据量改用列式批量策略。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   dataSlice: 要插入的数据切片，元素为结构体
	//   autoCommit: true 时本方法自开事务并提交；false 时必须已有事务
	// 返回:
	//   int64: 受影响行数
	//   error: 插入失败时返回错误信息
	BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error)

	// BatchUpdate 按 keyFields 匹配，批量更新切片中的结构体。
	// 预编译一条 UPDATE 后在事务中循环执行。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   dataSlice: 要更新的数据切片，元素为结构体
	//   keyFields: 匹配记录的关键字段（如主键）
	//   autoCommit: true 时本方法自开事务并提交；false 时必须已有事务
	// 返回:
	//   int64: 受影响行数
	//   error: 更新失败时返回错误信息
	BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error)

	// BatchDelete 按 keyFields 从切片中取键值，批量删除匹配记录。
	// 预编译一条 DELETE 后在事务中循环执行。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   dataSlice: 含待删键值的数据切片
	//   keyFields: 匹配记录的关键字段
	//   autoCommit: true 时本方法自开事务并提交；false 时必须已有事务
	// 返回:
	//   int64: 受影响行数
	//   error: 删除失败时返回错误信息
	BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error)

	// BatchDeleteByKeys 按主键值列表用 IN 子句一次删除。
	// 比逐条 BatchDelete 更高效。ClickHouse 由方言改写成 ALTER TABLE ... DELETE。
	// 参数:
	//   ctx: 用于超时、取消，以及携带事务
	//   table: 目标表名
	//   keyField: 主键字段名
	//   keys: 要删除的主键值列表
	//   autoCommit: true 自动提交；false 在当前事务中执行
	// 返回:
	//   int64: 受影响行数
	//   error: 删除失败时返回错误信息
	BatchDeleteByKeys(ctx context.Context, table string, keyField string, keys []interface{}, autoCommit bool) (int64, error)
}

// TxManager 管理事务。autoCommit 为 false 时必须先 BeginTx 或 InTx，后续执行使用返回的 context。
// 事务绑定在 context 上，不同 goroutine 可各自持有独立事务。
type TxManager interface {
	// BeginTx 开始事务并写入新的 context。
	// 同一 context 中已有事务时返回错误。
	// 参数:
	//   ctx: 用于超时和取消
	//   options: 隔离级别和只读设置，可为 nil 表示库默认
	// 返回:
	//   context.Context: 携带事务的新上下文，后续 SQL 应使用它
	//   error: 失败时包装为 ErrTransaction
	BeginTx(ctx context.Context, options *TxOptions) (context.Context, error)

	// Commit 提交 ctx 中的事务，使未提交更改生效。
	// 参数:
	//   ctx: 由 BeginTx 或 InTx 返回的、含事务的上下文
	// 返回:
	//   error: 无活跃事务或提交失败
	Commit(ctx context.Context) error

	// Rollback 回滚 ctx 中的事务，撤销未提交更改。
	// 参数:
	//   ctx: 由 BeginTx 或 InTx 返回的、含事务的上下文
	// 返回:
	//   error: 无活跃事务或回滚失败
	Rollback(ctx context.Context) error

	// InTx 自动开始事务、执行 fn、成功则提交。
	// fn 返回错误或发生 panic 时回滚；panic 转为 error，避免进程崩溃。
	// 参数:
	//   ctx: 用于超时和取消
	//   options: 隔离级别和只读设置，可为 nil
	//   fn: 在事务中执行，接收含事务的 context
	// 返回:
	//   error: 开始、执行、提交失败，或从 panic 转换的错误
	InTx(ctx context.Context, options *TxOptions, fn func(context.Context) error) error
}

// Database 统一的数据库入口，由 Conn、SQL、Records、TxManager 组合而成。
// Insert/Batch 留在 Records 上给业务直接调用；测试可按需只 mock SQL 或 Conn。
//
// 横向扩展关系库：
//  1. 在 dialect 注册 Spec（占位符、分页、DSN、错误归类）
//  2. 新建 pkg/database/<驱动>，Register(sqlbase.New(..., Hooks{}))
//  3. 在 alldriver 空白导入该包
//  4. 仅当批量或连接有特例时写 Hooks，或嵌入 sqlbase.DB 覆盖方法
type Database interface {
	Conn
	SQL
	Records
	TxManager
}

// Model 模型接口
// 定义数据模型需要实现的基本方法
// 用于ORM映射和自动化数据库操作
type Model interface {
	// TableName 获取表名
	// 返回当前模型对应的数据库表名
	// 返回:
	//   string: 数据库表名
	TableName() string

	// PrimaryKey 获取主键
	// 返回当前模型的主键字段名
	// 返回:
	//   string: 主键字段名
	PrimaryKey() string
}

// DriverCreator 数据库驱动创建函数
// 用于创建特定数据库驱动实例的工厂函数类型
// 返回:
//
//	Database: 数据库接口实例
type DriverCreator func() Database

// Register 注册数据库驱动
// 将数据库驱动的创建函数注册到系统中
// 支持的驱动类型由常量定义（MySQL、PostgreSQL、SQLite等）
// 参数:
//
//	driver: 驱动类型标识符
//	creator: 驱动创建函数
func Register(driver string, creator DriverCreator) {
	creatorMu.Lock()
	defer creatorMu.Unlock()
	dbCreators[driver] = creator
}

// lookupCreator 按驱动名取已注册的工厂，读锁保护注册表。
func lookupCreator(driver string) (DriverCreator, bool) {
	creatorMu.RLock()
	defer creatorMu.RUnlock()
	creator, exists := dbCreators[driver]
	return creator, exists
}

// GetConnectionID 根据配置生成连接ID
// 为数据库连接生成唯一标识符，用于连接缓存
// 优先使用配置中的连接名称，否则根据连接参数生成
// 参数:
//
//	config: 数据库配置
//
// 返回:
//
//	string: 连接唯一标识符
func GetConnectionID(config *DbConfig) string {
	if config.Name != "" {
		return config.Name
	}
	return fmt.Sprintf("%s:%s:%d:%s", config.Driver, config.Connection.Host, config.Connection.Port, config.Connection.Database)
}

// Open 打开数据库连接
// 根据配置创建数据库连接，支持连接缓存和DSN自动生成
// 如果连接已存在则复用，否则创建新连接
// 参数:
//
//	config: 数据库配置，包含驱动类型、连接信息等
//
// 返回:
//
//	Database: 数据库接口实例
//	error: 连接失败时返回错误信息
func Open(config *DbConfig) (Database, error) {
	if config == nil {
		return nil, fmt.Errorf("%w: config is nil", ErrConfigNotFound)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("database connection %s is disabled", config.Name)
	}

	creator, exists := lookupCreator(config.Driver)
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// 解密密码（如果需要）
	decryptedPassword, err := security.DecryptWithDefaultKey(config.Connection.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database password: %w", err)
	}
	config.Connection.Password = decryptedPassword

	// 生成DSN
	if config.DSN == "" {
		dsnStr, err := dsn.Generate(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DSN for %s: %w", config.Driver, err)
		}
		config.DSN = dsnStr
	}

	connectionID := GetConnectionID(config)

	connMutex.Lock()
	defer connMutex.Unlock()

	// 检查是否已存在连接
	if conn, exists := dbConnections[connectionID]; exists {
		return conn, nil
	}

	// 创建新连接
	db := creator()
	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 缓存连接
	dbConnections[connectionID] = db

	return db, nil
}

// openWithoutLock 内部方法：打开数据库连接（不加锁）
// 此方法假设调用者已经持有connMutex锁
// 参数:
//
//	config: 数据库配置，包含驱动类型、连接信息等
//
// 返回:
//
//	Database: 数据库接口实例
//	error: 连接失败时返回错误信息
func openWithoutLock(config *DbConfig) (Database, error) {
	if config == nil {
		return nil, fmt.Errorf("%w: config is nil", ErrConfigNotFound)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("database connection %s is disabled", config.Name)
	}

	creator, exists := lookupCreator(config.Driver)
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

	// 解密密码（如果需要）
	decryptedPassword, err := security.DecryptWithDefaultKey(config.Connection.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt database password: %w", err)
	}
	config.Connection.Password = decryptedPassword

	// 生成DSN
	if config.DSN == "" {
		dsnStr, err := dsn.Generate(config)
		if err != nil {
			return nil, fmt.Errorf("failed to generate DSN for %s: %w", config.Driver, err)
		}
		config.DSN = dsnStr
	}

	connectionID := GetConnectionID(config)

	// 检查是否已存在连接（此时调用者已持有锁）
	if conn, exists := dbConnections[connectionID]; exists {
		return conn, nil
	}

	// 创建新连接
	db := creator()
	if err := db.Connect(config); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 缓存连接
	dbConnections[connectionID] = db

	return db, nil
}

// GetConnection 获取已缓存的数据库连接
// 从连接池中获取指定名称的数据库连接
// 参数:
//
//	connectionName: 连接名称或连接ID
//
// 返回:
//
//	Database: 数据库接口实例，如果不存在则返回nil
func GetConnection(connectionName string) Database {
	connMutex.RLock()
	defer connMutex.RUnlock()

	return dbConnections[connectionName]
}

// GetDefaultConnection 获取默认数据库连接
// 默认连接名称为database.default
// 返回:
//
//	Database: 数据库接口实例
func GetDefaultConnection() Database {
	connMutex.RLock()
	defer connMutex.RUnlock()

	return dbConnections[config.GetString("database.default", "")]
}

// CloseAllConnections 关闭所有数据库连接
// 关闭系统中所有已建立的数据库连接，释放资源
// 通常在应用程序关闭时调用
// 返回:
//
//	error: 如果有连接关闭失败，返回包含所有错误的复合错误
func CloseAllConnections() error {
	connMutex.Lock()
	defer connMutex.Unlock()

	var errs []error
	for name, conn := range dbConnections {
		if err := conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection %s: %w", name, err))
		}
	}

	// 清空连接缓存
	dbConnections = make(map[string]Database)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	return nil
}

// LoadAllConnections 从配置文件加载所有数据库连接
// 解析配置文件中的所有数据库连接配置，创建并缓存连接实例
// 只有enabled为true的连接才会被创建
// 参数:
//
//	configPath: 配置文件路径
//
// 返回:
//
//	map[string]Database: 连接名称到数据库实例的映射
//	error: 加载失败时返回错误信息
func LoadAllConnections(configPath string) (map[string]Database, error) {
	connMutex.Lock()
	defer connMutex.Unlock()

	// 加载配置文件
	configs, err := dbtypes.LoadDatabaseConfigs(configPath)
	if err != nil {
		return nil, fmt.Errorf("加载数据库配置失败: %w", err)
	}

	connections := make(map[string]Database)

	// 遍历所有配置，创建启用的连接
	for name, config := range configs {
		// 检查连接是否启用
		if !config.Enabled {
			continue
		}

		// 创建数据库连接
		db, err := openWithoutLock(config)
		if err != nil {
			return nil, fmt.Errorf("创建数据库连接 '%s' 失败: %w", name, err)
		}

		db.SetName(name)

		// 缓存连接
		connectionID := GetConnectionID(config)
		dbConnections[connectionID] = db
		connections[name] = db
	}

	return connections, nil
}
