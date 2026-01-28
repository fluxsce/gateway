package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dblogger"
	"gateway/pkg/database/sqlutils"

	_ "github.com/mattn/go-sqlite3"
)

// 事务上下文键
const txContextKey = "gateway.sqlite.transaction"

// TxContext 事务上下文信息
type TxContext struct {
	tx      *sql.Tx
	id      string    // 事务ID
	created time.Time // 创建时间
	options *database.TxOptions
}

// setTxToContext 将事务信息设置到上下文中
func setTxToContext(ctx context.Context, txCtx *TxContext) context.Context {
	return context.WithValue(ctx, txContextKey, txCtx)
}

// getTxFromContext 从上下文中获取事务信息
func getTxFromContext(ctx context.Context) (*TxContext, bool) {
	txCtx, ok := ctx.Value(txContextKey).(*TxContext)
	return txCtx, ok
}

// generateTxID 生成事务ID
func generateTxID() string {
	return fmt.Sprintf("sqlite-tx-%d", time.Now().UnixNano())
}

// 注册SQLite驱动
func init() {
	database.Register(database.DriverSQLite, func() database.Database {
		return &SQLite{}
	})
}

// SQLite SQLite数据库实现
// 核心特性:
// 1. 统一的数据库接口实现 - 符合database.Database接口规范
// 2. 多线程安全事务管理 - 支持上下文绑定的事务，每个goroutine独立管理事务
// 3. 轻量级文件数据库 - 适合开发和小型应用场景
// 4. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
// 5. 智能日志记录 - 支持慢查询检测和SQL执行日志
// 6. 结构体映射 - 自动将Go结构体与数据库表映射
// 7. 上下文绑定事务 - 事务信息存储在context中，避免全局状态冲突
// 8. 并发安全 - SQLite在WAL模式下支持多读单写并发访问
// 9. Go底层优化 - 普通操作依赖Go database/sql的自动优化
type SQLite struct {
	db     *sql.DB
	config *database.DbConfig
	logger *dblogger.DBLogger
	mu     sync.RWMutex
}

// Connect 连接到SQLite数据库
// 建立SQLite数据库连接，配置连接池参数，并验证连接可用性
// SQLite特点：文件数据库，如果文件不存在会自动创建
// 参数:
//
//	config: SQLite数据库配置，包含DSN、连接池设置、日志配置等
//
// 返回:
//
//	error: 连接建立失败时返回错误信息
func (s *SQLite) Connect(config *database.DbConfig) error {
	s.config = config
	s.logger = dblogger.NewDBLogger(config)

	// 使用背景上下文进行连接日志记录
	s.logger.LogConnecting(context.Background(), database.DriverSQLite, config.DSN)

	// SQLite连接字符串通常是文件路径，支持特殊参数
	// 例如: "file:test.db?cache=shared&mode=rwc&_journal_mode=WAL"
	dsn := config.DSN
	if dsn == "" {
		dsn = ":memory:" // 默认使用内存数据库
	}

	// 打开数据库连接
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		s.logger.LogError(context.Background(), "打开SQLite连接", err)
		return fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// 设置连接池参数
	// SQLite推荐较小的连接池，因为它主要是单文件数据库
	maxOpenConns := 10
	if config.Pool.MaxOpenConns > 0 {
		maxOpenConns = config.Pool.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := 5
	if config.Pool.MaxIdleConns > 0 {
		maxIdleConns = config.Pool.MaxIdleConns
	}
	db.SetMaxIdleConns(maxIdleConns)

	connMaxLifetime := time.Hour
	if config.Pool.ConnMaxLifetime > 0 {
		connMaxLifetime = time.Duration(config.Pool.ConnMaxLifetime) * time.Second
	}
	db.SetConnMaxLifetime(connMaxLifetime)

	connMaxIdleTime := time.Hour
	if config.Pool.ConnMaxIdleTime > 0 {
		connMaxIdleTime = time.Duration(config.Pool.ConnMaxIdleTime) * time.Second
	}
	db.SetConnMaxIdleTime(connMaxIdleTime)

	// 检查连接是否正常
	if err := db.Ping(); err != nil {
		s.logger.LogPing(context.Background(), err)
		return fmt.Errorf("SQLite connection test failed: %w", err)
	}

	// 设置SQLite特定的PRAGMA语句来优化性能
	if err := s.configureDatabase(db); err != nil {
		return fmt.Errorf("SQLite database configuration failed: %w", err)
	}

	s.db = db
	s.logger.LogConnected(context.Background(), database.DriverSQLite, map[string]any{
		"maxOpenConns":    maxOpenConns,
		"maxIdleConns":    maxIdleConns,
		"connMaxLifetime": connMaxLifetime.String(),
		"connMaxIdleTime": connMaxIdleTime.String(),
		"dsn":             dsn,
	})

	return nil
}

// configureDatabase 配置SQLite数据库参数
// 设置WAL模式、同步模式等优化参数
func (s *SQLite) configureDatabase(db *sql.DB) error {
	// 设置WAL模式以支持并发读写
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// 设置同步模式为NORMAL以平衡性能和安全性
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}

	// 设置页面缓存大小（默认-2000表示2MB）
	if _, err := db.Exec("PRAGMA cache_size = -2000"); err != nil {
		return fmt.Errorf("failed to set cache size: %w", err)
	}

	// 设置忙等待超时 - 关键修复：解决"database table is locked"错误
	if _, err := db.Exec("PRAGMA busy_timeout = 30000"); err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}

// Close 关闭数据库连接
// 关闭SQLite数据库连接，释放相关资源
// 返回:
//
//	error: 关闭连接失败时返回错误信息
func (s *SQLite) Close() error {
	if s.db != nil {
		s.logger.LogDisconnect(context.Background(), database.DriverSQLite)
		return s.db.Close()
	}
	return nil
}

// DSN 返回数据库连接字符串
// 获取当前SQLite连接使用的数据源名称
// 返回值会被处理以隐藏敏感信息（虽然SQLite通常不含敏感信息）
// 返回:
//
//	string: 处理后的DSN字符串
func (s *SQLite) DSN() string {
	if s.config == nil {
		return ""
	}
	return dblogger.MaskDSN(s.config.DSN)
}

// DB 返回底层的sql.DB实例
// 获取SQLite连接底层的标准库sql.DB实例
// 用于需要直接访问底层数据库连接的场景
// 返回:
//
//	*sql.DB: 底层的sql.DB实例
func (s *SQLite) DB() *sql.DB {
	return s.db
}

// DriverName 返回数据库驱动名称
// 获取当前数据库使用的驱动名称标识
// 返回:
//
//	string: 固定返回"sqlite"
func (s *SQLite) DriverName() string {
	return database.DriverSQLite
}

// GetDriver 获取数据库驱动类型
// 实现Database接口，返回SQLite驱动标识
// 返回:
//
//	string: SQLite驱动类型标识
func (s *SQLite) GetDriver() string {
	return database.DriverSQLite
}

// GetName 获取数据库连接名称
// 实现Database接口，返回当前连接的名称
// 返回:
//
//	string: 数据库连接名称，如果配置为空则返回空字符串
func (s *SQLite) GetName() string {
	if s.config == nil {
		return ""
	}
	return s.config.Name
}

// SetName 设置数据库连接名称
// 用于在创建连接后设置连接名称标识
// 参数:
//
//	name: 连接名称
func (s *SQLite) SetName(name string) {
	if s.config != nil {
		s.config.Name = name
	}
}

// Ping 测试数据库连接
// 向SQLite数据库发送ping请求，验证连接状态
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//
// 返回:
//
//	error: 连接异常时返回错误信息
func (s *SQLite) Ping(ctx context.Context) error {
	err := s.db.PingContext(ctx)
	s.logger.LogPing(ctx, err)
	return err
}

// BeginTx 开始事务
// 启动一个新的SQLite事务
// SQLite注意事项：
// 1. SQLite支持的隔离级别有限，主要通过锁机制实现
// 2. WAL模式下支持多读单写，但写事务仍然是串行的
// 3. 事务信息会绑定到返回的上下文中，避免全局状态冲突
// 4. 每个上下文可以独立管理事务，但SQLite层面写操作会串行化
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置
//
// 返回:
//
//	context.Context: 包含事务信息的新上下文
//	error: 开始事务失败时返回错误信息
func (s *SQLite) BeginTx(ctx context.Context, options *database.TxOptions) (context.Context, error) {
	// 检查是否已经有活跃事务
	if _, ok := getTxFromContext(ctx); ok {
		return ctx, fmt.Errorf("transaction already active in context")
	}
	var sqlTxOpts *sql.TxOptions
	if options != nil {
		sqlTxOpts = &sql.TxOptions{
			ReadOnly: options.ReadOnly,
		}

		// SQLite的隔离级别支持有限，主要通过不同的锁模式实现
		// 这里映射到标准的SQL隔离级别，但实际行为由SQLite内部控制
		switch options.Isolation {
		case database.IsolationReadUncommitted:
			sqlTxOpts.Isolation = sql.LevelReadUncommitted
		case database.IsolationReadCommitted:
			sqlTxOpts.Isolation = sql.LevelReadCommitted
		case database.IsolationRepeatableRead:
			sqlTxOpts.Isolation = sql.LevelRepeatableRead
		case database.IsolationSerializable:
			sqlTxOpts.Isolation = sql.LevelSerializable
		default:
			sqlTxOpts.Isolation = sql.LevelDefault
		}
	}

	tx, err := s.db.BeginTx(ctx, sqlTxOpts)
	if err != nil {
		s.logger.LogTx(ctx, "开始", err)
		return ctx, fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	// 创建事务上下文
	txCtx := &TxContext{
		tx:      tx,
		id:      generateTxID(),
		created: time.Now(),
		options: options,
	}

	// 将事务信息绑定到上下文
	newCtx := setTxToContext(ctx, txCtx)
	s.logger.LogTx(newCtx, "开始", nil)
	return newCtx, nil
}

// Commit 提交事务
// 提交指定上下文中的SQLite事务，使所有未提交的更改生效
// 参数:
//
//	ctx: 包含事务信息的上下文
//
// 返回:
//
//	error: 提交事务失败时返回错误信息
func (s *SQLite) Commit(ctx context.Context) error {
	txCtx, ok := getTxFromContext(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}

	err := txCtx.tx.Commit()
	txCtx.tx = nil // 清理事务指针
	s.logger.LogTx(ctx, "提交", err)

	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// Rollback 回滚事务
// 回滚指定上下文中的SQLite事务，撤销所有未提交的更改
// 参数:
//
//	ctx: 包含事务信息的上下文
//
// 返回:
//
//	error: 回滚事务失败时返回错误信息
func (s *SQLite) Rollback(ctx context.Context) error {
	txCtx, ok := getTxFromContext(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}

	err := txCtx.tx.Rollback()
	txCtx.tx = nil // 清理事务指针
	s.logger.LogTx(ctx, "回滚", err)

	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// InTx 在事务中执行函数
// 自动管理SQLite事务的生命周期，支持上下文绑定的事务
// 如果函数正常返回，自动提交事务
// 如果函数返回错误或发生panic，自动回滚事务并将panic转换为错误
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置
//	fn: 在事务中执行的函数，接收包含事务信息的上下文
//
// 返回:
//
//	error: 事务执行失败时返回错误信息，包括panic转换的错误
func (s *SQLite) InTx(ctx context.Context, options *database.TxOptions, fn func(context.Context) error) (err error) {
	txCtx, err := s.BeginTx(ctx, options)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			s.Rollback(txCtx)
			// 将panic转换为错误，避免程序崩溃
			err = fmt.Errorf("transaction panic recovered: %v", r)
		}
	}()

	if err := fn(txCtx); err != nil {
		s.Rollback(txCtx)
		return err
	}

	return s.Commit(txCtx)
}

// getExecutor 获取执行器（事务或连接）
// 根据autoCommit参数和上下文中的事务状态返回合适的执行器
// 如果autoCommit为false且上下文包含活跃事务，返回事务执行器
// 否则返回数据库连接执行器
// 参数:
//
//	ctx: 上下文，可能包含事务信息
//	autoCommit: 是否自动提交
//
// 返回:
//
//	interface: 执行器接口，可以是*sql.Tx或*sql.DB
func (s *SQLite) getExecutor(ctx context.Context, autoCommit bool) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
} {
	if !autoCommit {
		txCtx, ok := getTxFromContext(ctx)
		if ok && txCtx.tx != nil {
			return txCtx.tx
		}
	}
	return s.db
}

// Exec 执行SQL语句
// 执行INSERT、UPDATE、DELETE等不返回结果集的SQLite语句
// 支持事务和非事务模式执行
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	query: 要执行的SQL语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 执行失败时返回错误信息
func (s *SQLite) Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error) {
	// SQLite 需要将 time.Time 转换为字符串格式
	convertedArgs := s.convertTimeArgs(args)

	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	result, err := executor.ExecContext(ctx, query, convertedArgs...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	s.logger.LogSQL(ctx, "SQL执行", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Query 查询多条记录
// 执行SELECT语句并将结果扫描到目标切片中
// 自动处理结构体字段到数据库列的映射
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	dest: 目标切片的指针，用于接收查询结果
//	query: 要执行的SELECT语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败或扫描失败时返回错误信息
func (s *SQLite) Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	rows, err := executor.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		if err != sql.ErrNoRows {
			s.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, map[string]interface{}{
				"rowCount": 0,
			})
		}
		return err
	}
	defer rows.Close()

	err = sqlutils.ScanRows(rows, dest)
	rowCount := reflect.ValueOf(dest).Elem().Len()

	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		s.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": rowCount,
	}
	s.logger.LogSQL(ctx, "SQL查询", query, args, nil, duration, extra)

	return err
}

// QueryOne 查询单条记录
// 执行SELECT语句并将结果扫描到目标结构体中
// 如果查询不到记录，返回ErrRecordNotFound错误
// 使用智能字段映射，支持数据库列数与结构体字段数不匹配的情况
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	dest: 目标结构体的指针，用于接收查询结果
//	query: 要执行的SELECT语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败、扫描失败或记录不存在时返回错误信息
func (s *SQLite) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	rows, err := executor.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		s.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 使用智能扫描方式处理单行结果，支持字段数量不匹配
	err = sqlutils.ScanOneRow(rows, dest)

	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		s.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": map[bool]int{true: 1, false: 0}[err == nil],
	}
	s.logger.LogSQL(ctx, "SQL单行查询", query, args, nil, duration, extra)

	return err
}

// Insert 插入记录
// 根据提供的数据结构体自动构建INSERT语句并执行
// 会自动提取结构体字段作为列名和值，支持db tag映射
// SQLite特点：支持ROWID自增主键
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	data: 要插入的数据结构体，字段通过db tag映射到数据库列
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 插入记录的自增ID（如果有）
//	error: 插入失败时返回错误信息
func (s *SQLite) Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error) {
	query, args, err := sqlutils.BuildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	// SQLite 需要将 time.Time 转换为字符串格式
	// 因为 SQLite 将日期时间存储为 TEXT 类型
	convertedArgs := s.convertTimeArgs(args)

	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	result, err := executor.ExecContext(ctx, query, convertedArgs...)
	duration := time.Since(start)

	var lastInsertId int64
	var rowsAffected int64
	if err == nil {
		lastInsertId, _ = result.LastInsertId()
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志（使用原始参数，便于调试）
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}
	s.logger.LogSQL(ctx, "SQL插入", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return lastInsertId, nil
}

// Update 更新记录
// 根据提供的数据结构体和WHERE条件构建UPDATE语句并执行
// 会自动提取结构体字段作为要更新的列和值
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	data: 包含更新数据的结构体，字段通过db tag映射到数据库列
//	where: WHERE条件语句，可包含占位符
//	args: WHERE条件中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息
func (s *SQLite) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool, skipZero bool) (int64, error) {
	setClause, setArgs, err := sqlutils.BuildUpdateQuery(table, data, skipZero)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("UPDATE %s SET %s", table, setClause)
	if where != "" {
		query += " WHERE " + where
		setArgs = append(setArgs, args...)
	}

	// SQLite 需要将 time.Time 转换为字符串格式
	convertedArgs := s.convertTimeArgs(setArgs)

	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	result, err := executor.ExecContext(ctx, query, convertedArgs...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	s.logger.LogSQL(ctx, "SQL更新", query, setArgs, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Delete 删除记录
// 根据WHERE条件构建DELETE语句并执行
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	where: WHERE条件语句，可包含占位符
//	args: WHERE条件中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (s *SQLite) Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}

	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	result, err := executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	s.logger.LogSQL(ctx, "SQL删除", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// BatchInsert 批量插入记录
// 将切片中的多个数据结构体批量插入到SQLite中
//
// SQLite批量插入特点：
// 1. 使用预编译循环执行模式：预编译一次，循环执行多次
// 2. 事务保证：默认在事务中执行，确保数据一致性
// 3. WAL模式优化：在WAL模式下批量插入性能很好
// 4. 错误处理：任何错误都会触发事务回滚，保证原子性
// 5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	dataSlice: 要插入的数据切片，每个元素都是结构体
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 插入失败时返回错误信息
func (s *SQLite) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	// 获取第一个元素来构建SQL结构
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 构建单条INSERT的预编译SQL语句（参考MySQL实现）
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 开始事务（BatchInsert默认需要事务保证一致性）
	var needCommit bool
	var tx *sql.Tx

	if autoCommit {
		// 自动提交模式：创建新事务
		tx, err = s.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		needCommit = true
	} else {
		// 手动事务模式：使用当前事务
		txCtx, ok := getTxFromContext(ctx)
		if !ok || txCtx.tx == nil {
			return 0, fmt.Errorf("no active transaction for batch insert")
		}
		tx = txCtx.tx
	}

	// 预编译单条INSERT语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		if needCommit {
			tx.Rollback()
		}
		return 0, fmt.Errorf("failed to prepare batch insert statement: %w", err)
	}
	defer stmt.Close()

	// 循环执行预编译语句，逐条插入数据
	var totalRowsAffected int64
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		_, values, err := sqlutils.ExtractColumnsAndValues(item)
		if err != nil {
			if needCommit {
				tx.Rollback()
			}
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}

		// 执行单条插入
		result, err := stmt.ExecContext(ctx, values...)
		if err != nil {
			if needCommit {
				tx.Rollback() // 出现错误时回滚事务
			}
			return 0, fmt.Errorf("failed to insert item %d: %w", i, err)
		}

		// 累计影响行数
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch insert transaction: %w", err)
		}
	}

	// 记录执行日志
	extra := map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"columnsCount":  len(columns),
		"executionMode": "prepared_loop",
	}
	s.logger.LogSQL(ctx, "SQL批量插入", query, []interface{}{"[batch_data]"}, nil, duration, extra)

	return totalRowsAffected, nil
}

// BatchUpdate 批量更新记录
// 将切片中的多个数据结构体批量更新到SQLite中
// 使用预编译循环执行方式提高性能
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	dataSlice: 要更新的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息
func (s *SQLite) BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	// 获取第一个元素来构建SQL结构
	firstItem := slice.Index(0).Interface()
	allColumns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 分离SET字段和WHERE字段
	var setColumns []string
	var whereColumns []string

	keyFieldsMap := make(map[string]bool)
	for _, key := range keyFields {
		keyFieldsMap[key] = true
	}

	for _, col := range allColumns {
		if keyFieldsMap[col] {
			whereColumns = append(whereColumns, col)
		} else {
			setColumns = append(setColumns, col)
		}
	}

	if len(setColumns) == 0 {
		return 0, fmt.Errorf("no columns to update")
	}
	if len(whereColumns) == 0 {
		return 0, fmt.Errorf("no key fields for WHERE clause")
	}

	// 构建UPDATE语句
	var setParts []string
	for _, col := range setColumns {
		setParts = append(setParts, fmt.Sprintf("%s = ?", col))
	}

	var whereParts []string
	for _, col := range whereColumns {
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", col))
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table,
		strings.Join(setParts, ", "),
		strings.Join(whereParts, " AND "))

	executor := s.getExecutor(ctx, autoCommit)

	// 预编译语句
	start := time.Now()
	stmt, err := executor.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var totalAffected int64

	// 循环执行
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		_, values, err := sqlutils.ExtractColumnsAndValues(item)
		if err != nil {
			return totalAffected, err
		}

		// 按照SQL参数顺序重新排列：SET参数 + WHERE参数
		var args []interface{}
		valueMap := make(map[string]interface{})
		for j, col := range allColumns {
			valueMap[col] = values[j]
		}

		// 先添加SET参数
		for _, col := range setColumns {
			args = append(args, valueMap[col])
		}
		// 再添加WHERE参数
		for _, col := range whereColumns {
			args = append(args, valueMap[col])
		}

		result, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			return totalAffected, fmt.Errorf("failed to execute update for item %d: %w", i, err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return totalAffected, fmt.Errorf("failed to get rows affected for item %d: %w", i, err)
		}

		totalAffected += affected
	}

	duration := time.Since(start)

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": totalAffected,
		"itemCount":    slice.Len(),
	}
	s.logger.LogSQL(ctx, "SQL批量更新", query, []interface{}{fmt.Sprintf("batch of %d items", slice.Len())}, nil, duration, extra)

	return totalAffected, nil
}

// BatchDelete 批量删除记录
// 根据提供的数据切片批量删除记录，通过指定的关键字段匹配
// 使用预编译循环执行方式提高性能
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	dataSlice: 包含要删除记录信息的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (s *SQLite) BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	if len(keyFields) == 0 {
		return 0, fmt.Errorf("keyFields cannot be empty")
	}

	// 构建DELETE语句
	var whereParts []string
	for _, field := range keyFields {
		whereParts = append(whereParts, fmt.Sprintf("%s = ?", field))
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		table,
		strings.Join(whereParts, " AND "))

	executor := s.getExecutor(ctx, autoCommit)

	// 预编译语句
	start := time.Now()
	stmt, err := executor.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	var totalAffected int64

	// 循环执行
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		columns, values, err := sqlutils.ExtractColumnsAndValues(item)
		if err != nil {
			return totalAffected, err
		}

		// 提取关键字段对应的值
		valueMap := make(map[string]interface{})
		for j, col := range columns {
			valueMap[col] = values[j]
		}

		var args []interface{}
		for _, field := range keyFields {
			if val, exists := valueMap[field]; exists {
				args = append(args, val)
			} else {
				return totalAffected, fmt.Errorf("key field %s not found in item %d", field, i)
			}
		}

		result, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			return totalAffected, fmt.Errorf("failed to execute delete for item %d: %w", i, err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return totalAffected, fmt.Errorf("failed to get rows affected for item %d: %w", i, err)
		}

		totalAffected += affected
	}

	duration := time.Since(start)

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": totalAffected,
		"itemCount":    slice.Len(),
	}
	s.logger.LogSQL(ctx, "SQL批量删除", query, []interface{}{fmt.Sprintf("batch of %d items", slice.Len())}, nil, duration, extra)

	return totalAffected, nil
}

// BatchDeleteByKeys 根据主键列表批量删除记录
// 使用IN子句的高效批量删除方式
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，可能包含事务信息
//	table: 目标表名
//	keyField: 主键字段名
//	keys: 要删除的主键值列表
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (s *SQLite) BatchDeleteByKeys(ctx context.Context, table string, keyField string, keys []interface{}, autoCommit bool) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	// 构建IN子句
	placeholders := make([]string, len(keys))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
		table, keyField, strings.Join(placeholders, ", "))

	executor := s.getExecutor(ctx, autoCommit)

	start := time.Now()
	result, err := executor.ExecContext(ctx, query, keys...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
		"keyCount":     len(keys),
	}
	s.logger.LogSQL(ctx, "SQL批量删除(按键)", query, keys, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// 重构说明：
// 1. SQL格式化功能已移动到 pkg/database/sqlutils/sql_format.go
//    - BuildInsertQuery: 构建INSERT语句
//    - BuildUpdateQuery: 构建UPDATE语句的SET子句
//    - ExtractColumnsAndValues: 从结构体提取列名和值
//    - IsZeroValue: 检查值是否为零值
//
// 2. 结果处理功能已移动到 pkg/database/sqlutils/result_format.go
//    - ScanRows: 扫描多行结果到切片
//    - ScanOneRow: 扫描单行结果到结构体
//    - PrepareScanTargetsWithFields: 准备扫描目标
//    - FindFieldByColumn: 根据列名查找字段
//    - CreateNullSafeScanTarget: 创建NULL值安全的扫描目标
//    - ProcessScannedValues: 处理扫描后的值转换
//    - ScanRowsTraditional: 传统方式扫描多行结果
//    - ScanRowsWithInterfaceSlice: 接口切片方式扫描多行结果

// SQLite特殊说明：
// 1. 连接池配置建议较小的值，因为SQLite主要是单文件数据库
// 2. 自动配置WAL模式以支持并发读写
// 3. 启用外键约束以保证数据完整性
// 4. 事务隔离级别映射有限，主要由SQLite内部锁机制控制
// 5. 支持:memory:内存数据库，适合测试场景
// 6. 文件数据库如果不存在会自动创建
//
// 多线程事务限制：
// 1. SQLite支持多读单写：多个读事务可以并发，但写事务是串行的
// 2. WAL模式优化：虽然启用了WAL模式，但写操作仍然需要获取独占锁
// 3. 上下文绑定：事务信息绑定到上下文，支持多个 goroutine 独立管理事务
// 4. 性能考虑：对于高并发写入场景，建议考虑使用MySQL等关系型数据库
// 5. 适用场景：适合轻量级应用、开发测试、嵌入式系统等场景

// convertTimeArgs 将参数中的 time.Time 转换为字符串格式
// SQLite 将日期时间存储为 TEXT 类型，需要字符串格式
// 支持的格式：2006-01-02 15:04:05（SQLite 标准日期时间格式）
func (s *SQLite) convertTimeArgs(args []interface{}) []interface{} {
	if args == nil {
		return nil
	}

	converted := make([]interface{}, len(args))
	for i, arg := range args {
		if t, ok := arg.(time.Time); ok {
			// 转换为 SQLite 标准日期时间格式
			converted[i] = t.Format("2006-01-02 15:04:05")
		} else {
			converted[i] = arg
		}
	}
	return converted
}
