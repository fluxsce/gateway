package database

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/config"
	"gohub/pkg/database/dbtypes"
	"gohub/pkg/database/dsn"
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

// 数据库工厂映射及缓存
var (
	// dbCreators 存储注册的数据库驱动创建函数
	dbCreators = make(map[string]DriverCreator)

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

// Database 统一的数据库接口
// 通过autoCommit参数控制是否自动提交，简化事务处理
type Database interface {
	// === 连接管理 ===
	
	// Connect 连接数据库
	// 根据提供的配置建立数据库连接，包括连接池设置、日志配置等
	// 参数:
	//   config: 数据库配置，包含连接信息、池设置、日志等
	// 返回:
	//   error: 连接失败时返回错误信息
	Connect(config *DbConfig) error

	// Close 关闭数据库连接
	// 关闭当前数据库连接，释放相关资源
	// 如果有活跃的事务，会先回滚事务再关闭连接
	// 返回:
	//   error: 关闭失败时返回错误信息
	Close() error

	// Ping 测试数据库连接
	// 发送ping请求到数据库服务器，验证连接是否正常
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	// 返回:
	//   error: 连接异常时返回错误信息
	Ping(ctx context.Context) error

	// === 基本操作 ===

	// Exec 执行SQL语句
	// 执行INSERT、UPDATE、DELETE等不返回结果集的SQL语句
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   query: 要执行的SQL语句，可包含占位符
	//   args: SQL语句中占位符对应的参数值
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 执行失败时返回错误信息
	Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error)

	// Query 查询多条记录
	// 执行SELECT语句并将结果扫描到目标切片中
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   dest: 目标切片的指针，用于接收查询结果
	//   query: 要执行的SELECT语句，可包含占位符
	//   args: SQL语句中占位符对应的参数值
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   error: 查询失败或扫描失败时返回错误信息
	Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error

	// QueryOne 查询单条记录
	// 执行SELECT语句并将结果扫描到目标结构体中
	// 如果查询不到记录，返回ErrRecordNotFound错误
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   dest: 目标结构体的指针，用于接收查询结果
	//   query: 要执行的SELECT语句，可包含占位符
	//   args: SQL语句中占位符对应的参数值
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   error: 查询失败、扫描失败或记录不存在时返回错误信息
	QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error

	// Insert 插入记录
	// 根据提供的数据结构体自动构建INSERT语句并执行
	// 会自动提取结构体字段作为列名和值
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   data: 要插入的数据结构体，字段通过db tag映射到数据库列
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 插入记录的自增ID（如果有）
	//   error: 插入失败时返回错误信息
	Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error)

	// Update 更新记录
	// 根据提供的数据结构体和WHERE条件构建UPDATE语句并执行
	// 会自动提取结构体字段作为要更新的列和值
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   data: 包含更新数据的结构体，字段通过db tag映射到数据库列
	//   where: WHERE条件语句，可包含占位符
	//   args: WHERE条件中占位符对应的参数值
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 更新失败时返回错误信息
	Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool) (int64, error)

	// Delete 删除记录
	// 根据WHERE条件构建DELETE语句并执行
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   where: WHERE条件语句，可包含占位符
	//   args: WHERE条件中占位符对应的参数值
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 删除失败时返回错误信息
	Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error)

	// BatchInsert 批量插入记录
	// 将切片中的多个数据结构体批量插入到数据库中
	// 使用单条INSERT语句提高性能
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   dataSlice: 要插入的数据切片，每个元素都是结构体
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 插入失败时返回错误信息
	BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error)

	// BatchUpdate 批量更新记录
	// 将切片中的多个数据结构体批量更新到数据库中
	// 使用单条UPDATE语句提高性能，根据主键或指定字段进行更新
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   dataSlice: 要更新的数据切片，每个元素都是结构体
	//   keyFields: 用于匹配记录的关键字段列表（如主键字段）
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 更新失败时返回错误信息
	BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error)

	// BatchDelete 批量删除记录
	// 根据提供的数据切片批量删除记录，通过指定的关键字段匹配
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   dataSlice: 包含要删除记录信息的数据切片，每个元素都是结构体
	//   keyFields: 用于匹配记录的关键字段列表（如主键字段）
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 删除失败时返回错误信息
	BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error)

	// BatchDeleteByKeys 根据主键列表批量删除记录
	// 更高效的批量删除方式，直接提供主键值列表
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   table: 目标表名
	//   keyField: 主键字段名
	//   keys: 要删除的主键值列表
	//   autoCommit: true-自动提交, false-需要手动调用Commit/Rollback
	// 返回:
	//   int64: 受影响的行数
	//   error: 删除失败时返回错误信息
	BatchDeleteByKeys(ctx context.Context, table string, keyField string, keys []interface{}, autoCommit bool) (int64, error)

	// === 事务控制 ===

	// BeginTx 开始事务
	// 启动一个新的数据库事务，可以指定隔离级别和只读属性
	// 多线程安全：每个上下文可以独立管理事务
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   options: 事务选项，包含隔离级别和只读设置
	// 返回:
	//   context.Context: 包含事务信息的新上下文
	//   error: 开始事务失败时返回错误信息
	BeginTx(ctx context.Context, options *TxOptions) (context.Context, error)

	// Commit 提交事务
	// 提交上下文中的事务，使所有未提交的更改生效
	// 参数:
	//   ctx: 包含事务信息的上下文
	// 返回:
	//   error: 提交失败时返回错误信息
	Commit(ctx context.Context) error

	// Rollback 回滚事务
	// 回滚上下文中的事务，撤销所有未提交的更改
	// 参数:
	//   ctx: 包含事务信息的上下文
	// 返回:
	//   error: 回滚失败时返回错误信息
	Rollback(ctx context.Context) error

	// InTx 在事务中执行函数
	// 自动处理事务的开始、提交和回滚
	// 如果函数正常返回，自动提交事务
	// 如果函数返回错误或发生panic，自动回滚事务
	// 参数:
	//   ctx: 上下文，用于控制请求超时和取消
	//   options: 事务选项，包含隔离级别和只读设置
	//   fn: 在事务中执行的函数，接收包含事务的上下文，返回error表示是否成功
	// 返回:
	//   error: 事务执行失败时返回错误信息
	InTx(ctx context.Context, options *TxOptions, fn func(context.Context) error) error

	// === 工具方法 ===

	// GetDriver 获取数据库驱动类型
	// 返回当前数据库实例使用的驱动类型标识
	// 返回:
	//   string: 驱动类型（如"mysql", "postgres", "sqlite"）
	GetDriver() string

	// GetName 获取数据库连接名称
	// 返回当前数据库连接的名称标识
	// 返回:
	//   string: 连接名称
	GetName() string
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
//   Database: 数据库接口实例
type DriverCreator func() Database

// Register 注册数据库驱动
// 将数据库驱动的创建函数注册到系统中
// 支持的驱动类型由常量定义（MySQL、PostgreSQL、SQLite等）
// 参数:
//   driver: 驱动类型标识符
//   creator: 驱动创建函数
func Register(driver string, creator DriverCreator) {
	dbCreators[driver] = creator
}

// GetConnectionID 根据配置生成连接ID
// 为数据库连接生成唯一标识符，用于连接缓存
// 优先使用配置中的连接名称，否则根据连接参数生成
// 参数:
//   config: 数据库配置
// 返回:
//   string: 连接唯一标识符
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
//   config: 数据库配置，包含驱动类型、连接信息等
// 返回:
//   Database: 数据库接口实例
//   error: 连接失败时返回错误信息
func Open(config *DbConfig) (Database, error) {
	if config == nil {
		return nil, fmt.Errorf("%w: config is nil", ErrConfigNotFound)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("database connection %s is disabled", config.Name)
	}

	creator, exists := dbCreators[config.Driver]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

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
//   config: 数据库配置，包含驱动类型、连接信息等
// 返回:
//   Database: 数据库接口实例
//   error: 连接失败时返回错误信息
func openWithoutLock(config *DbConfig) (Database, error) {
	if config == nil {
		return nil, fmt.Errorf("%w: config is nil", ErrConfigNotFound)
	}

	if !config.Enabled {
		return nil, fmt.Errorf("database connection %s is disabled", config.Name)
	}

	creator, exists := dbCreators[config.Driver]
	if !exists {
		return nil, fmt.Errorf("unsupported database driver: %s", config.Driver)
	}

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
//   connectionName: 连接名称或连接ID
// 返回:
//   Database: 数据库接口实例，如果不存在则返回nil
func GetConnection(connectionName string) Database {
	connMutex.RLock()
	defer connMutex.RUnlock()

	return dbConnections[connectionName]
}

// GetDefaultConnection 获取默认数据库连接
// 默认连接名称为database.default
// 返回:
//   Database: 数据库接口实例
func GetDefaultConnection() Database {
	connMutex.RLock()
	defer connMutex.RUnlock()

	return dbConnections[config.GetString("database.default", "")]
}

// CloseAllConnections 关闭所有数据库连接
// 关闭系统中所有已建立的数据库连接，释放资源
// 通常在应用程序关闭时调用
// 返回:
//   error: 如果有连接关闭失败，返回包含所有错误的复合错误
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
//   configPath: 配置文件路径
// 返回:
//   map[string]Database: 连接名称到数据库实例的映射
//   error: 加载失败时返回错误信息
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

		// 设置连接名称
		if dbImpl, ok := db.(interface{ SetName(string) }); ok {
			dbImpl.SetName(name)
		}

		// 缓存连接
		connectionID := GetConnectionID(config)
		dbConnections[connectionID] = db
		connections[name] = db
	}

	return connections, nil
}
