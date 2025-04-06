package database

import (
	"context"
	"errors"
	"fmt"
	"gohub/pkg/config"
	"gohub/pkg/database/dbtypes"
	"gohub/pkg/database/dsn"
	huberrors "gohub/pkg/utils/huberrors"
	"sync"
	"time"
)

// 定义通用数据库错误，方便上层业务进行错误处理
var (
	// ErrRecordNotFound 记录未找到错误
	// 当查询不返回任何结果时返回此错误
	ErrRecordNotFound error = errors.New("record not found")

	// ErrDuplicateKey 键重复错误
	// 当插入操作违反唯一性约束时返回此错误
	ErrDuplicateKey error = errors.New("duplicate key")

	// ErrConnection 连接错误
	// 当无法连接到数据库时返回此错误
	ErrConnection error = errors.New("database connection error")

	// ErrTransaction 事务错误
	// 当事务操作失败时返回此错误
	ErrTransaction error = errors.New("transaction error")

	// ErrInvalidQuery 无效查询错误
	// 当SQL查询语法错误或参数错误时返回此错误
	ErrInvalidQuery error = errors.New("invalid query")

	// ErrConfigNotFound 配置未找到错误
	// 当找不到指定的数据库配置时返回此错误
	ErrConfigNotFound error = errors.New("database config not found")
)

// 数据库工厂映射及缓存
var (
	// dbCreators 存储注册的数据库驱动创建函数
	dbCreators = make(map[string]DriverCreator)

	// dbConnections 缓存已创建的数据库连接实例
	// 键是连接名称，值是数据库连接实例
	dbConnections = make(map[string]Database)

	// connMutex 用于保护数据库连接缓存的互斥锁
	connMutex = sync.RWMutex{}
)

// 支持的数据库类型常量，使用dbtypes中的定义
const (
	// MySQL数据库驱动
	DriverMySQL = dbtypes.DriverMySQL
	// PostgreSQL数据库驱动
	DriverPostgreSQL = dbtypes.DriverPostgreSQL
	// SQLite数据库驱动
	DriverSQLite = dbtypes.DriverSQLite
)

// DbConfig 使用dbtypes包中的定义
type DbConfig = dbtypes.DbConfig

// DBExecutor 数据库执行器接口
// 定义了基本的数据库操作方法，无论是事务内还是事务外
// 这是最基础的接口，被Transaction和Database接口继承
type DBExecutor interface {
	// Exec 执行SQL语句
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	Exec(ctx context.Context, query string, args []interface{}) (int64, error)

	// Query 查询多条记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - dest: 目标结构体切片指针，查询结果将被映射到这里
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	// 返回:
	//   - error: 可能的错误
	Query(ctx context.Context, dest interface{}, query string, args []interface{}) error

	// QueryOne 查询单条记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - dest: 目标结构体指针，查询结果将被映射到这里
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	// 返回:
	//   - error: 可能的错误，如果没有记录则返回ErrRecordNotFound
	QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}) error

	// Insert 插入记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - data: 要插入的数据，通常是一个结构体
	// 返回:
	//   - int64: 插入的记录ID或影响的行数
	//   - error: 可能的错误
	Insert(ctx context.Context, table string, data interface{}) (int64, error)

	// Update 更新记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - data: 要更新的数据，通常是一个结构体
	//   - query: 条件查询，WHERE子句
	//   - args: 查询参数，用于替换查询中的占位符
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	Update(ctx context.Context, table string, data interface{}, query string, args []interface{}) (int64, error)

	// Delete 删除记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - query: 条件查询，WHERE子句
	//   - args: 查询参数，用于替换查询中的占位符
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	Delete(ctx context.Context, table string, query string, args []interface{}) (int64, error)
}

// Transaction 事务接口
// 在DBExecutor基础上添加事务控制方法
// 表示一个活跃的数据库事务
type Transaction interface {
	// 继承DBExecutor接口的所有方法
	DBExecutor

	// Commit 提交事务
	// 将事务中的所有更改持久化到数据库
	// 返回:
	//   - error: 提交事务时可能发生的错误
	Commit() error

	// Rollback 回滚事务
	// 撤销事务中的所有更改
	// 返回:
	//   - error: 回滚事务时可能发生的错误
	Rollback() error
}

// Database 数据库接口
// 定义了数据库连接和高级操作的方法
// 主要服务于应用层，提供完整的数据库访问能力
type Database interface {
	// 继承DBExecutor接口的所有方法
	DBExecutor

	// Connect 连接数据库
	// 参数:
	//   - config: 数据库配置，包含连接信息和行为设置
	// 返回:
	//   - error: 连接数据库时可能发生的错误
	Connect(config *DbConfig) error

	// Close 关闭数据库连接
	// 释放数据库连接资源
	// 返回:
	//   - error: 关闭连接时可能发生的错误
	Close() error

	// Ping 测试数据库连接
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	// 返回:
	//   - error: 测试连接时可能发生的错误
	Ping(ctx context.Context) error

	// BeginTx 开始事务
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - options: 事务选项，如隔离级别和只读模式
	// 返回:
	//   - Transaction: 新创建的事务对象
	//   - error: 开始事务时可能发生的错误
	BeginTx(ctx context.Context, options ...TxOption) (Transaction, error)

	// WithTx 使用已有事务执行函数
	// 参数:
	//   - tx: 已存在的事务对象
	//   - fn: 在事务中执行的函数，接收事务对象作为参数
	// 返回:
	//   - error: 事务执行过程中可能发生的错误
	// 此方法负责提交或回滚事务，简化事务使用
	WithTx(tx Transaction, fn func(tx Transaction) error) error

	// ExecWithOptions 带选项执行SQL语句
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	//   - options: 执行选项，如是否使用事务
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	ExecWithOptions(ctx context.Context, query string, args []interface{}, options ...ExecOption) (int64, error)

	// QueryWithOptions 带选项查询多条记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - dest: 目标结构体切片指针，查询结果将被映射到这里
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	//   - options: 查询选项，如是否使用事务
	// 返回:
	//   - error: 可能的错误
	QueryWithOptions(ctx context.Context, dest interface{}, query string, args []interface{}, options ...QueryOption) error

	// QueryOneWithOptions 带选项查询单条记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - dest: 目标结构体指针，查询结果将被映射到这里
	//   - query: SQL查询语句
	//   - args: 查询参数，用于替换查询中的占位符
	//   - options: 查询选项，如是否使用事务
	// 返回:
	//   - error: 可能的错误，如果没有记录则返回ErrRecordNotFound
	QueryOneWithOptions(ctx context.Context, dest interface{}, query string, args []interface{}, options ...QueryOption) error

	// InsertWithOptions 带选项插入记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - data: 要插入的数据，通常是一个结构体
	//   - options: 执行选项，如是否使用事务
	// 返回:
	//   - int64: 插入的记录ID或影响的行数
	//   - error: 可能的错误
	InsertWithOptions(ctx context.Context, table string, data interface{}, options ...ExecOption) (int64, error)

	// UpdateWithOptions 带选项更新记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - data: 要更新的数据，通常是一个结构体
	//   - query: 条件查询，WHERE子句
	//   - args: 查询参数，用于替换查询中的占位符
	//   - options: 执行选项，如是否使用事务
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	UpdateWithOptions(ctx context.Context, table string, data interface{}, query string, args []interface{}, options ...ExecOption) (int64, error)

	// DeleteWithOptions 带选项删除记录
	// 参数:
	//   - ctx: 上下文，用于控制超时和取消
	//   - table: 表名
	//   - query: 条件查询，WHERE子句
	//   - args: 查询参数，用于替换查询中的占位符
	//   - options: 执行选项，如是否使用事务
	// 返回:
	//   - int64: 影响的行数
	//   - error: 可能的错误
	DeleteWithOptions(ctx context.Context, table string, query string, args []interface{}, options ...ExecOption) (int64, error)

	// GetDriver 获取数据库驱动类型
	// 返回当前数据库连接使用的驱动类型
	// 返回:
	//   - string: 驱动类型，如"mysql"、"postgres"等
	GetDriver() string

	// GetName 获取数据库连接名称
	// 返回当前数据库连接的名称
	// 返回:
	//   - string: 连接名称
	GetName() string
}

// Model 模型接口
// 定义了模型操作的通用方法
// 实现此接口的结构体可以与数据库操作无缝集成
type Model interface {
	// TableName 获取表名
	// 返回模型对应的数据库表名
	// 返回:
	//   - string: 表名
	TableName() string

	// PrimaryKey 获取主键
	// 返回模型的主键字段名
	// 返回:
	//   - string: 主键名
	PrimaryKey() string
}

// DriverCreator 工厂函数类型定义
// 用于创建特定数据库驱动的实例
type DriverCreator func() Database

// Register 注册数据库驱动
// 参数:
//   - driver: 驱动名称，如"mysql"
//   - creator: 创建函数，返回该驱动的数据库实例
//
// 驱动实现需要在init()中调用此函数注册自己
func Register(driver string, creator DriverCreator) {
	dbCreators[driver] = creator
}

// GetConnectionID 获取连接唯一标识
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: 连接唯一标识，即连接名称
func GetConnectionID(config *DbConfig) string {
	// 直接使用连接名称作为ID
	// 如果连接名称为空，则使用default
	name := config.Name
	if name == "" {
		name = "default"
	}
	return name
}

// Open 打开数据库连接
// 参数:
//   - config: 数据库配置，包含驱动类型和连接信息
//
// 返回:
//   - Database: 数据库接口实例
//   - error: 打开连接时可能发生的错误
//
// 使用示例:
//
//	db, err := database.Open(&database.DbConfig{Driver: "mysql", Name: "read", DSN: "..."})
func Open(config *DbConfig) (Database, error) {
	// 获取连接唯一标识，现在只是连接名称
	connID := GetConnectionID(config)

	// 先检查是否已有缓存的连接
	connMutex.RLock()
	db, exists := dbConnections[connID]
	connMutex.RUnlock()

	// 如果已存在且连接有效，直接返回
	if exists {
		// 简单测试连接是否有效
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := db.Ping(ctx)
		if err == nil {
			return db, nil
		}
		// 连接已失效，从缓存中删除
		connMutex.Lock()
		delete(dbConnections, connID)
		connMutex.Unlock()
	}

	// 获取指定驱动的创建函数
	creator, ok := dbCreators[config.Driver]
	if !ok {
		return nil, huberrors.NewError("unsupported database driver: %s", config.Driver)
	}

	// 为每个连接创建新的数据库实例
	db = creator()

	// 连接数据库
	if err := db.Connect(config); err != nil {
		return nil, huberrors.WrapError(err, "failed to connect to %s database", config.Driver)
	}

	// 缓存连接实例，使用连接名称作为键
	connMutex.Lock()
	dbConnections[connID] = db
	connMutex.Unlock()

	return db, nil
}

// GetConnection 获取指定名称的数据库连接
// 参数:
//   - connectionName: 连接名称，如为空则使用"default"
//
// 返回:
//   - Database: 数据库接口实例，如果连接不存在则返回nil
//
// 使用示例:
//
//	db := database.GetConnection("read") // 获取名为"read"的连接
//	defaultDB := database.GetConnection("") // 获取默认连接
func GetConnection(connectionName string) Database {
	if connectionName == "" {
		connectionName = "default"
	}

	connMutex.RLock()
	db, exists := dbConnections[connectionName]
	connMutex.RUnlock()

	if exists {
		return db
	}

	return nil
}

// CloseAllConnections 关闭所有数据库连接
// 返回关闭过程中遇到的第一个错误，如果没有错误则返回nil
// 使用优化的锁策略，避免在关闭连接时长时间持有全局锁
func CloseAllConnections() error {
	// 使用一个临时map来存储需要关闭的连接
	// 这样可以最小化持有锁的时间
	connections := make(map[string]Database)

	connMutex.Lock()
	// 复制连接映射表，而不是直接在持有锁的情况下操作它们
	for connID, db := range dbConnections {
		connections[connID] = db
	}
	// 清空全局连接映射表
	dbConnections = make(map[string]Database)
	connMutex.Unlock()

	// 在不持有锁的情况下关闭连接
	var firstErr error
	for connID, db := range connections {
		if err := db.Close(); err != nil && firstErr == nil {
			firstErr = huberrors.WrapError(err, "关闭数据库连接失败: %s", connID)
		}
	}

	return firstErr
}

// LoadConfig 从配置文件加载数据库配置
// 参数:
//   - configPath: 配置文件路径，如果为空则使用默认路径 configs/database.yaml
//   - connection: 数据库连接名称，如果为空则使用default
//
// 返回:
//   - *DbConfig: 加载的配置
//   - error: 错误信息
//
// 用法示例:
//
//	config, err := database.LoadConfig("", "mysql")
func LoadConfig(configPath string, connection string) (*DbConfig, error) {
	if configPath == "" {
		configPath = "configs/database.yaml"
	}

	// 尝试从全局配置获取数据库配置
	// 如果全局配置中没有数据库配置，则尝试加载
	if !config.IsExist("database.connections") && configPath != "" {
		err := config.LoadConfigFile(configPath, config.LoadOptions{
			AllowOverride: true,
			ClearExisting: false,
		})
		if err != nil {
			return nil, huberrors.WrapError(err, "加载数据库配置文件失败: %s", configPath)
		}
	}

	// 如果未指定连接名称，使用默认连接
	if connection == "" {
		connection = config.GetString("database.default", "default")
	}

	// 读取特定连接的配置
	key := fmt.Sprintf("database.connections.%s", connection)
	var dbConfig DbConfig
	if err := config.GetSection(key, &dbConfig); err != nil {
		return nil, huberrors.WrapError(err, "解析数据库配置失败: %s", key)
	}

	// 验证配置有效性
	if dbConfig.Driver == "" {
		return nil, huberrors.WrapError(ErrConfigNotFound, "数据库驱动类型未指定: %s", key)
	}

	// 设置连接名称
	dbConfig.Name = connection

	// 如果未指定启用状态，默认为启用
	if !dbConfig.Enabled {
		enabledKey := fmt.Sprintf("database.connections.%s.enabled", connection)
		dbConfig.Enabled = config.GetBool(enabledKey, true)
	}

	// 如果DSN为空，则根据结构化配置生成连接字符串
	if dbConfig.DSN == "" {
		dsnStr, err := dsn.Generate(&dbConfig)
		if err != nil {
			return nil, huberrors.WrapError(err, "生成数据库连接字符串失败")
		}
		dbConfig.DSN = dsnStr
	}

	return &dbConfig, nil
}

// OpenWithConfigFile 使用配置文件打开数据库连接
// 参数:
//   - configPath: 配置文件路径，如果为空则使用默认路径 configs/database.yaml
//   - connection: 数据库连接名称，如果为空则使用默认连接
//
// 返回:
//   - Database: 数据库实例
//   - error: 错误信息
//
// 用法示例:
//
//	db, err := database.OpenWithConfigFile("", "")              // 使用默认连接
//	db1, err := database.OpenWithConfigFile("", "mysql")        // 使用名为mysql的连接
//	db2, err := database.OpenWithConfigFile("", "mysql_second") // 使用同一种driver(mysql)的另一个连接
func OpenWithConfigFile(configPath string, connection string) (Database, error) {
	// 加载数据库配置
	config, err := LoadConfig(configPath, connection)
	if err != nil {
		return nil, huberrors.WrapError(err, "无法加载数据库配置")
	}

	// 使用配置打开数据库连接
	return Open(config)
}

// LoadAllConnections 从配置文件加载所有启用的数据库连接
// 参数:
//   - configPath: 配置文件路径，如果为空则使用默认路径 configs/database.yaml
//
// 返回:
//   - map[string]Database: 连接名称到数据库实例的映射
//   - error: 错误信息
//
// 用法示例:
//
//	connections, err := database.LoadAllConnections("")
//	if err != nil {
//		// 处理错误
//	}
//	// 使用特定连接
//	mysqlDB := connections["mysql"]  // 这里的"mysql"是连接名称
//	readDB := connections["read"]    // 获取名为"read"的连接
func LoadAllConnections(configPath string) (map[string]Database, error) {
	// 尝试使用全局配置中已有的连接配置
	connectionsMap := config.Get("database.connections", nil)

	// 如果全局配置中没有连接配置，且指定了配置文件路径，则尝试加载
	if connectionsMap == nil && configPath != "" {
		// 以不清除现有配置的方式加载
		err := config.LoadConfigFile(configPath, config.LoadOptions{
			AllowOverride: true,
			ClearExisting: false,
		})
		if err != nil {
			return nil, huberrors.WrapError(err, "加载数据库配置文件失败: %s", configPath)
		}

		// 再次尝试获取连接配置
		connectionsMap = config.Get("database.connections", nil)
	}

	if connectionsMap == nil {
		return nil, huberrors.NewError("配置文件中未找到数据库连接配置")
	}

	connections := make(map[string]Database)
	cm, ok := connectionsMap.(map[string]interface{})
	if !ok {
		return nil, huberrors.NewError("数据库连接配置格式错误")
	}

	// 遍历所有连接配置
	for connName := range cm {
		// 检查连接是否启用
		enabledKey := fmt.Sprintf("database.connections.%s.enabled", connName)
		enabled := config.GetBool(enabledKey, true) // 默认启用
		if !enabled {
			continue // 跳过禁用的连接
		}

		// 加载并连接数据库
		db, err := OpenWithConfigFile(configPath, connName)
		if err != nil {
			return connections, huberrors.WrapError(err, "加载数据库连接失败: %s", connName)
		}

		// 添加到连接映射，直接使用连接名称作为键
		connections[connName] = db
	}

	// 如果没有成功加载任何连接，返回错误
	if len(connections) == 0 {
		return nil, huberrors.NewError("未能加载任何数据库连接，请检查配置文件")
	}

	return connections, nil
}
