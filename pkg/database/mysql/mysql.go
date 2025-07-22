package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"sync"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dblogger"
	"gateway/pkg/database/sqlutils"

	_ "github.com/go-sql-driver/mysql"
)

// 注册MySQL驱动
func init() {
	database.Register(database.DriverMySQL, func() database.Database {
		return &MySQL{}
	})
}

// MySQL MySQL数据库实现
// 核心特性:
// 1. 统一的数据库接口实现 - 符合database.Database接口规范
// 2. 多线程安全事务管理 - 支持多个goroutine并发开始和管理独立的事务
// 3. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
// 4. 智能日志记录 - 支持慢查询检测和SQL执行日志
// 5. 结构体映射 - 自动将Go结构体与数据库表映射
// 6. 上下文绑定事务 - 事务信息存储在context中，避免全局状态冲突
// 7. Go底层优化 - 普通操作依赖Go database/sql的自动优化
// 8. 智能预编译 - 仅在必要时（如批量操作）使用手动预编译
type MySQL struct {
	db     *sql.DB
	config *database.DbConfig
	logger *dblogger.DBLogger
	mu     sync.RWMutex
	// 移除全局单一事务字段，改为上下文绑定
	// currentTx *sql.Tx // 已删除 - 这是多线程问题的根源
}

// 事务上下文键，使用字符串常量更清晰
const txContextKey = "gateway.mysql.transaction"

// TxContext 事务上下文，包含事务和相关元数据
type TxContext struct {
	tx      *sql.Tx
	id      string              // 事务ID，用于日志跟踪
	created time.Time           // 事务创建时间
	options *database.TxOptions // 事务选项
}

// setTxToContext 将事务存储到上下文中
func setTxToContext(ctx context.Context, txCtx *TxContext) context.Context {
	return context.WithValue(ctx, txContextKey, txCtx)
}

// getTxFromContext 从上下文中获取事务
func getTxFromContext(ctx context.Context) (*TxContext, bool) {
	txCtx, ok := ctx.Value(txContextKey).(*TxContext)
	return txCtx, ok
}

// generateTxID 生成事务ID
func generateTxID() string {
	return fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), rand.Int63())
}

// Connect 连接到MySQL数据库
// 建立MySQL数据库连接，配置连接池参数，并验证连接可用性
// 会根据配置设置最大连接数、空闲连接数、连接生命周期等参数
// 参数:
//
//	config: MySQL数据库配置，包含DSN、连接池设置、日志配置等
//
// 返回:
//
//	error: 连接建立失败时返回错误信息
func (m *MySQL) Connect(config *database.DbConfig) error {
	m.config = config
	m.logger = dblogger.NewDBLogger(config)

	// 使用背景上下文进行连接日志记录
	m.logger.LogConnecting(context.Background(), database.DriverMySQL, config.DSN)

	// 打开数据库连接
	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		m.logger.LogError(context.Background(), "打开MySQL连接", err)
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// 设置连接池参数
	maxOpenConns := 25
	if config.Pool.MaxOpenConns > 0 {
		maxOpenConns = config.Pool.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpenConns)

	maxIdleConns := 25
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
		m.logger.LogPing(context.Background(), err)
		return fmt.Errorf("MySQL connection test failed: %w", err)
	}

	m.db = db
	m.logger.LogConnected(context.Background(), database.DriverMySQL, map[string]any{
		"maxOpenConns":    maxOpenConns,
		"maxIdleConns":    maxIdleConns,
		"connMaxLifetime": connMaxLifetime.String(),
		"connMaxIdleTime": connMaxIdleTime.String(),
	})

	return nil
}

// Close 关闭数据库连接
// 关闭MySQL数据库连接，释放相关资源
// 注意：使用上下文绑定事务的情况下，Close不会自动回滚事务
// 用户需要在关闭连接前手动处理事务
// 返回:
//
//	error: 关闭连接失败时返回错误信息
func (m *MySQL) Close() error {
	if m.db != nil {
		m.logger.LogDisconnect(context.Background(), database.DriverMySQL)
		return m.db.Close()
	}
	return nil
}

// DSN 返回数据库连接字符串
// 获取当前MySQL连接使用的数据源名称
// 返回值会被处理以隐藏敏感信息（如密码）
// 返回:
//
//	string: 处理后的DSN字符串，隐藏敏感信息
func (m *MySQL) DSN() string {
	if m.config == nil {
		return ""
	}
	// 导入数据库logger包只是为了访问这个函数
	return dblogger.MaskDSN(m.config.DSN)
}

// DB 返回底层的sql.DB实例
// 获取MySQL连接底层的标准库sql.DB实例
// 用于需要直接访问底层数据库连接的场景
// 返回:
//
//	*sql.DB: 底层的sql.DB实例
func (m *MySQL) DB() *sql.DB {
	return m.db
}

// DriverName 返回数据库驱动名称
// 获取当前数据库使用的驱动名称标识
// 返回:
//
//	string: 固定返回"mysql"
func (m *MySQL) DriverName() string {
	return database.DriverMySQL
}

// GetDriver 获取数据库驱动类型
// 实现Database接口，返回MySQL驱动标识
// 返回:
//
//	string: MySQL驱动类型标识
func (m *MySQL) GetDriver() string {
	return database.DriverMySQL
}

// GetName 获取数据库连接名称
// 实现Database接口，返回当前连接的名称
// 返回:
//
//	string: 数据库连接名称，如果配置为空则返回空字符串
func (m *MySQL) GetName() string {
	if m.config == nil {
		return ""
	}
	return m.config.Name
}

// SetName 设置数据库连接名称
// 用于在创建连接后设置连接名称标识
// 参数:
//
//	name: 连接名称
func (m *MySQL) SetName(name string) {
	if m.config != nil {
		m.config.Name = name
	}
}

// Ping 测试数据库连接
// 向MySQL服务器发送ping请求，验证连接状态
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//
// 返回:
//
//	error: 连接异常时返回错误信息
func (m *MySQL) Ping(ctx context.Context) error {
	err := m.db.PingContext(ctx)
	m.logger.LogPing(ctx, err)
	return err
}

// BeginTx 开始事务
// 启动一个新的MySQL事务，支持指定隔离级别和只读属性
// 多线程安全：每个上下文可以独立管理事务
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置
//
// 返回:
//
//	context.Context: 包含事务信息的新上下文
//	error: 开始事务失败时返回错误信息
func (m *MySQL) BeginTx(ctx context.Context, options *database.TxOptions) (context.Context, error) {
	// 检查是否已经有事务
	if _, ok := getTxFromContext(ctx); ok {
		return ctx, fmt.Errorf("transaction already active in context")
	}

	var sqlTxOpts *sql.TxOptions
	if options != nil {
		sqlTxOpts = &sql.TxOptions{
			ReadOnly: options.ReadOnly,
		}

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

	tx, err := m.db.BeginTx(ctx, sqlTxOpts)
	if err != nil {
		m.logger.LogTx(ctx, "开始", err)
		return ctx, fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	txCtx := &TxContext{
		tx:      tx,
		id:      generateTxID(),
		created: time.Now(),
		options: options,
	}

	// 将事务信息绑定到上下文
	newCtx := setTxToContext(ctx, txCtx)
	m.logger.LogTx(newCtx, "开始", nil)

	return newCtx, nil
}

// Commit 提交事务
// 提交上下文中的MySQL事务，使所有未提交的更改生效
// 参数:
//
//	ctx: 包含事务信息的上下文
//
// 返回:
//
//	error: 提交事务失败时返回错误信息
func (m *MySQL) Commit(ctx context.Context) error {
	txCtx, ok := getTxFromContext(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}

	err := txCtx.tx.Commit()
	txCtx.tx = nil
	m.logger.LogTx(ctx, "提交", err)

	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// Rollback 回滚事务
// 回滚上下文中的MySQL事务，撤销所有未提交的更改
// 参数:
//
//	ctx: 包含事务信息的上下文
//
// 返回:
//
//	error: 回滚事务失败时返回错误信息
func (m *MySQL) Rollback(ctx context.Context) error {
	txCtx, ok := getTxFromContext(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}

	err := txCtx.tx.Rollback()
	txCtx.tx = nil
	m.logger.LogTx(ctx, "回滚", err)

	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// InTx 在事务中执行函数
// 自动管理MySQL事务的生命周期
// 如果函数正常返回，自动提交事务
// 如果函数返回错误或发生panic，自动回滚事务并将panic转换为错误
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置
//	fn: 在事务中执行的函数，接收包含事务的上下文，返回error表示是否成功
//
// 返回:
//
//	error: 事务执行失败时返回错误信息，包括panic转换的错误
func (m *MySQL) InTx(ctx context.Context, options *database.TxOptions, fn func(context.Context) error) (err error) {
	txCtx, err := m.BeginTx(ctx, options)
	if err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			m.Rollback(txCtx)
			// 将panic转换为错误，避免程序崩溃
			err = fmt.Errorf("transaction panic recovered: %v", r)
		}
	}()

	if err := fn(txCtx); err != nil {
		m.Rollback(txCtx)
		return err
	}

	return m.Commit(txCtx)
}

// getExecutor 获取执行器（事务或连接）
// 根据autoCommit参数和上下文中的事务状态返回合适的执行器
// 如果autoCommit为false且上下文中存在活跃事务，返回事务执行器
// 否则返回数据库连接执行器
// 参数:
//
//	ctx: 上下文，用于获取事务信息
//	autoCommit: 是否自动提交
//
// 返回:
//
//	interface: 执行器接口，可以是*sql.Tx或*sql.DB
func (m *MySQL) getExecutor(ctx context.Context, autoCommit bool) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	if !autoCommit {
		txCtx, ok := getTxFromContext(ctx)
		if ok && txCtx.tx != nil {
			return txCtx.tx
		}
	}
	return m.db
}

// Exec 执行SQL语句
// 执行INSERT、UPDATE、DELETE等不返回结果集的MySQL语句
// 使用Go底层自动优化，无需手动预编译
// 支持事务和非事务模式执行
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	query: 要执行的SQL语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 执行失败时返回错误信息
func (m *MySQL) Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error) {
	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接执行，让Go底层自动优化
	result, err := executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	m.logger.LogSQL(ctx, "SQL执行", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Query 查询多条记录
// 执行SELECT语句并将结果扫描到目标切片中
// 使用Go底层自动优化，无需手动预编译
// 自动处理结构体字段到数据库列的映射
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	dest: 目标切片的指针，用于接收查询结果
//	query: 要执行的SELECT语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败或扫描失败时返回错误信息
func (m *MySQL) Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接查询，让Go底层自动优化
	rows, err := executor.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		if err != sql.ErrNoRows {
			m.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, map[string]interface{}{
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
		m.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": rowCount,
	}
	m.logger.LogSQL(ctx, "SQL查询", query, args, nil, duration, extra)

	return err
}

// QueryOne 查询单条记录
// 执行SELECT语句并将结果扫描到目标结构体中
// 如果查询不到记录，返回ErrRecordNotFound错误
// 使用智能字段映射，支持数据库列数与结构体字段数不匹配的情况
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	dest: 目标结构体的指针，用于接收查询结果
//	query: 要执行的SELECT语句，可包含占位符
//	args: SQL语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败、扫描失败或记录不存在时返回错误信息
func (m *MySQL) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接查询，让Go底层自动优化
	// 使用QueryContext而不是QueryRowContext，以便获取列信息进行智能映射
	rows, err := executor.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	if err != nil {
		m.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 使用智能扫描方式处理单行结果，支持字段数量不匹配
	err = sqlutils.ScanOneRow(rows, dest)

	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		m.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, map[string]interface{}{
			"rowCount": 0,
		})
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": map[bool]int{true: 1, false: 0}[err == nil],
	}
	m.logger.LogSQL(ctx, "SQL单行查询", query, args, nil, duration, extra)

	return err
}

// Insert 插入记录
// 根据提供的数据结构体自动构建INSERT语句并执行
// 使用Go底层自动优化，无需手动预编译
// 会自动提取结构体字段作为列名和值，支持db tag映射
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	data: 要插入的数据结构体，字段通过db tag映射到数据库列
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 插入记录的自增ID（如果有）
//	error: 插入失败时返回错误信息
func (m *MySQL) Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error) {
	query, args, err := sqlutils.BuildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接执行，让Go底层自动优化
	result, err := executor.ExecContext(ctx, query, args...)
	duration := time.Since(start)

	var lastInsertId int64
	var rowsAffected int64
	if err == nil {
		lastInsertId, _ = result.LastInsertId()
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}
	m.logger.LogSQL(ctx, "SQL插入", query, args, err, duration, extra)

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
//	ctx: 上下文，用于控制请求超时和取消
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
func (m *MySQL) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool) (int64, error) {
	setClause, setArgs, err := sqlutils.BuildUpdateQuery(table, data)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("UPDATE %s SET %s", table, setClause)
	if where != "" {
		query += " WHERE " + where
		setArgs = append(setArgs, args...)
	}

	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接执行，让Go底层自动优化
	result, err := executor.ExecContext(ctx, query, setArgs...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	m.logger.LogSQL(ctx, "SQL更新", query, setArgs, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// Delete 删除记录
// 根据WHERE条件构建DELETE语句并执行
// 使用Go底层自动优化，无需手动预编译
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	where: WHERE条件语句，可包含占位符
//	args: WHERE条件中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (m *MySQL) Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}

	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接执行，让Go底层自动优化
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
	m.logger.LogSQL(ctx, "SQL删除", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// BatchInsert 批量插入记录
// 将切片中的多个数据结构体批量插入到MySQL中
//
// 注意：这是唯一保留手动预编译的方法，因为批量操作确实需要预编译优化
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用sql.PrepareContext()预编译单条INSERT语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条插入数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 预编译循环执行流程：
//  1. 分析数据结构，提取列信息
//  2. 构建单条INSERT的预编译SQL语句
//  3. 开始事务（autoCommit=true时自动创建，false时使用当前事务）
//  4. 预编译单条INSERT语句（预编译一次，执行多次）
//  5. 循环执行：for _, item := range dataSlice { stmt.Exec(item...) }
//  6. 提交事务或在错误时回滚
//
// 优势对比：
//   - vs 大SQL拼接：内存使用稳定，不受批量大小影响
//   - vs 多次Insert调用：减少预编译开销，事务保证一致性
//   - vs Go底层自动优化：批量操作时手动预编译性能更优
//
// 注意：
//   - BatchInsert默认需要事务，确保批量操作的原子性
//   - 适合中小批量（≤1000条），大批量建议业务层分批调用
//   - 任何单条记录插入失败都会回滚整个批次
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 要插入的数据切片，每个元素都是结构体
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 插入失败时返回错误信息
func (m *MySQL) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建单条INSERT的预编译SQL语句
	// 这是最高效的方式：预编译一次，循环执行多次
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	// 第三步：开始事务（BatchInsert默认需要事务保证一致性）
	var needCommit bool
	var tx *sql.Tx

	if autoCommit {
		// 自动提交模式：创建新事务
		tx, err = m.db.BeginTx(ctx, nil)
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

	// 第四步：预编译单条INSERT语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		if needCommit {
			tx.Rollback()
		}
		return 0, fmt.Errorf("failed to prepare batch insert statement: %w", err)
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条插入数据
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

	// 第六步：提交事务（如果是自动提交模式）
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
	m.logger.LogSQL(ctx, "SQL批量插入", query, []interface{}{"[batch_data]"}, nil, duration, extra)

	return totalRowsAffected, nil
}

// BatchUpdate 批量更新记录
// 将切片中的多个数据结构体批量更新到MySQL中
// 使用预编译循环执行模式，根据指定的关键字段进行匹配更新
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用sql.PrepareContext()预编译单条UPDATE语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条更新数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 要更新的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息
func (m *MySQL) BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
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

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建UPDATE语句，分离SET子句和WHERE子句
	var setClauses []string
	var whereClause []string

	for _, col := range columns {
		isKeyField := false
		for _, keyField := range keyFields {
			if col == keyField {
				isKeyField = true
				break
			}
		}

		if isKeyField {
			whereClause = append(whereClause, col+" = ?")
		} else {
			setClauses = append(setClauses, col+" = ?")
		}
	}

	if len(setClauses) == 0 {
		return 0, fmt.Errorf("no fields to update (all fields are key fields)")
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		table,
		strings.Join(setClauses, ", "),
		strings.Join(whereClause, " AND "))

	// 第三步：开始事务
	var needCommit bool
	var tx *sql.Tx

	if autoCommit {
		tx, err = m.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		needCommit = true
	} else {
		txCtx, ok := getTxFromContext(ctx)
		if !ok || txCtx.tx == nil {
			return 0, fmt.Errorf("no active transaction for batch update")
		}
		tx = txCtx.tx
	}

	// 第四步：预编译UPDATE语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		if needCommit {
			tx.Rollback()
		}
		return 0, fmt.Errorf("failed to prepare batch update statement: %w", err)
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条更新数据
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

		// 重新排列参数：SET子句参数 + WHERE子句参数
		var args []interface{}
		for _, col := range columns {
			isKeyField := false
			for _, keyField := range keyFields {
				if col == keyField {
					isKeyField = true
					break
				}
			}

			if !isKeyField {
				// 找到对应的值
				for j, column := range columns {
					if column == col {
						args = append(args, values[j])
						break
					}
				}
			}
		}

		// 添加WHERE条件的参数
		for _, keyField := range keyFields {
			for j, column := range columns {
				if column == keyField {
					args = append(args, values[j])
					break
				}
			}
		}

		// 执行单条更新
		result, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			if needCommit {
				tx.Rollback()
			}
			return 0, fmt.Errorf("failed to update item %d: %w", i, err)
		}

		// 累计影响行数
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 第六步：提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch update transaction: %w", err)
		}
	}

	// 记录执行日志
	extra := map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"keyFields":     keyFields,
		"executionMode": "prepared_loop",
	}
	m.logger.LogSQL(ctx, "SQL批量更新", query, []interface{}{"[batch_data]"}, nil, duration, extra)

	return totalRowsAffected, nil
}

// BatchDelete 批量删除记录
// 根据提供的数据切片批量删除记录，通过指定的关键字段匹配
// 使用预编译循环执行模式提高性能
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用sql.PrepareContext()预编译单条DELETE语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条删除数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 包含要删除记录信息的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (m *MySQL) BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
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

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建DELETE语句的WHERE子句
	var whereClause []string
	for _, keyField := range keyFields {
		whereClause = append(whereClause, keyField+" = ?")
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s",
		table,
		strings.Join(whereClause, " AND "))

	// 第三步：开始事务
	var needCommit bool
	var tx *sql.Tx

	if autoCommit {
		tx, err = m.db.BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		needCommit = true
	} else {
		txCtx, ok := getTxFromContext(ctx)
		if !ok || txCtx.tx == nil {
			return 0, fmt.Errorf("no active transaction for batch delete")
		}
		tx = txCtx.tx
	}

	// 第四步：预编译DELETE语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		if needCommit {
			tx.Rollback()
		}
		return 0, fmt.Errorf("failed to prepare batch delete statement: %w", err)
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条删除数据
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

		// 提取WHERE条件的参数值
		var args []interface{}
		for _, keyField := range keyFields {
			for j, column := range columns {
				if column == keyField {
					args = append(args, values[j])
					break
				}
			}
		}

		// 执行单条删除
		result, err := stmt.ExecContext(ctx, args...)
		if err != nil {
			if needCommit {
				tx.Rollback()
			}
			return 0, fmt.Errorf("failed to delete item %d: %w", i, err)
		}

		// 累计影响行数
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 第六步：提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch delete transaction: %w", err)
		}
	}

	// 记录执行日志
	extra := map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"keyFields":     keyFields,
		"executionMode": "prepared_loop",
	}
	m.logger.LogSQL(ctx, "SQL批量删除", query, []interface{}{"[batch_data]"}, nil, duration, extra)

	return totalRowsAffected, nil
}

// BatchDeleteByKeys 根据主键列表批量删除记录
// 更高效的批量删除方式，直接提供主键值列表
// 使用IN子句进行批量删除，比逐条删除更高效
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	keyField: 主键字段名
//	keys: 要删除的主键值列表
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (m *MySQL) BatchDeleteByKeys(ctx context.Context, table string, keyField string, keys []interface{}, autoCommit bool) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}

	if keyField == "" {
		return 0, fmt.Errorf("keyField cannot be empty")
	}

	// 构建IN子句的占位符
	placeholders := make([]string, len(keys))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE %s IN (%s)",
		table,
		keyField,
		strings.Join(placeholders, ", "))

	executor := m.getExecutor(ctx, autoCommit)

	start := time.Now()

	// 直接执行，使用IN子句批量删除
	result, err := executor.ExecContext(ctx, query, keys...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected":  rowsAffected,
		"batchSize":     len(keys),
		"keyField":      keyField,
		"executionMode": "in_clause",
	}
	m.logger.LogSQL(ctx, "SQL批量删除(主键)", query, keys, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// 实现说明
//
// 1. 普通操作优化：
//    - Exec、Query、QueryOne、Insert、Update、Delete等单次操作
//    - 直接使用Go database/sql的ExecContext、QueryContext等方法
//    - 依赖Go底层的自动优化和驱动层优化
//    - 简化代码，减少预编译语句管理的复杂度
//
// 2. 批量操作优化：
//    - BatchInsert等批量操作仍使用手动预编译
//    - 一次预编译，多次执行，显著提升批量操作性能
//    - 在事务中执行，保证数据一致性
//
// 3. 工具函数依赖：
//    - SQL格式化：sqlutils.BuildInsertQuery, BuildUpdateQuery等
//    - 结果扫描：sqlutils.ScanRows, ScanOneRow等
//    - 详细功能请参考 pkg/database/sqlutils/ 包
