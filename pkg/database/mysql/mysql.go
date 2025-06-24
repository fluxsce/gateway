package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"gohub/pkg/database"
	"gohub/pkg/database/dblogger"

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
// 2. 灵活的事务管理 - 支持不同隔离级别和可定制的事务选项
// 3. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
// 4. 智能日志记录 - 支持慢查询检测和SQL执行日志
// 5. 结构体映射 - 自动将Go结构体与数据库表映射
// 6. 带选项的操作 - 每个数据库操作都有带选项的版本，支持自定义执行行为
// 7. 活跃事务追踪 - 通过内部映射跟踪所有活跃事务
type MySQL struct {
	db       *sql.DB
	config   *database.DbConfig
	logger   *dblogger.DBLogger
	mu       sync.RWMutex
	currentTx *sql.Tx // 当前活跃的事务
}

// Connect 连接到MySQL数据库
// 建立MySQL数据库连接，配置连接池参数，并验证连接可用性
// 会根据配置设置最大连接数、空闲连接数、连接生命周期等参数
// 参数:
//   config: MySQL数据库配置，包含DSN、连接池设置、日志配置等
// 返回:
//   error: 连接建立失败时返回错误信息
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
// 如果存在活跃事务，会先回滚事务再关闭连接
// 返回:
//   error: 关闭连接失败时返回错误信息
func (m *MySQL) Close() error {
	if m.currentTx != nil {
		m.currentTx.Rollback()
		m.currentTx = nil
	}
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
//   string: 处理后的DSN字符串，隐藏敏感信息
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
//   *sql.DB: 底层的sql.DB实例
func (m *MySQL) DB() *sql.DB {
	return m.db
}

// DriverName 返回数据库驱动名称
// 获取当前数据库使用的驱动名称标识
// 返回:
//   string: 固定返回"mysql"
func (m *MySQL) DriverName() string {
	return database.DriverMySQL
}

// GetDriver 获取数据库驱动类型
// 实现Database接口，返回MySQL驱动标识
// 返回:
//   string: MySQL驱动类型标识
func (m *MySQL) GetDriver() string {
	return database.DriverMySQL
}

// GetName 获取数据库连接名称
// 实现Database接口，返回当前连接的名称
// 返回:
//   string: 数据库连接名称，如果配置为空则返回空字符串
func (m *MySQL) GetName() string {
	if m.config == nil {
		return ""
	}
	return m.config.Name
}

// SetName 设置数据库连接名称
// 用于在创建连接后设置连接名称标识
// 参数:
//   name: 连接名称
func (m *MySQL) SetName(name string) {
	if m.config != nil {
		m.config.Name = name
	}
}

// Ping 测试数据库连接
// 向MySQL服务器发送ping请求，验证连接状态
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
// 返回:
//   error: 连接异常时返回错误信息
func (m *MySQL) Ping(ctx context.Context) error {
	err := m.db.PingContext(ctx)
	m.logger.LogPing(ctx, err)
	return err
}

// BeginTx 开始事务
// 启动一个新的MySQL事务，支持指定隔离级别和只读属性
// 如果已有活跃事务，会返回错误
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   options: 事务选项，包含隔离级别和只读设置
// 返回:
//   error: 开始事务失败时返回错误信息
func (m *MySQL) BeginTx(ctx context.Context, options *database.TxOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentTx != nil {
		return fmt.Errorf("transaction already active")
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
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	m.currentTx = tx
	m.logger.LogTx(ctx, "开始", nil)
	return nil
}

// Commit 提交事务
// 提交当前活跃的MySQL事务，使所有未提交的更改生效
// 如果没有活跃事务，会返回错误
// 返回:
//   error: 提交事务失败时返回错误信息
func (m *MySQL) Commit() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentTx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := m.currentTx.Commit()
	m.currentTx = nil
	m.logger.LogTx(context.Background(), "提交", err)
	
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// Rollback 回滚事务
// 回滚当前活跃的MySQL事务，撤销所有未提交的更改
// 如果没有活跃事务，会返回错误
// 返回:
//   error: 回滚事务失败时返回错误信息
func (m *MySQL) Rollback() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentTx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := m.currentTx.Rollback()
	m.currentTx = nil
	m.logger.LogTx(context.Background(), "回滚", err)
	
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// InTx 在事务中执行函数
// 自动管理MySQL事务的生命周期
// 如果函数正常返回，自动提交事务
// 如果函数返回错误或发生panic，自动回滚事务
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   options: 事务选项，包含隔离级别和只读设置
//   fn: 在事务中执行的函数，返回error表示是否成功
// 返回:
//   error: 事务执行失败时返回错误信息
func (m *MySQL) InTx(ctx context.Context, options *database.TxOptions, fn func() error) error {
	if err := m.BeginTx(ctx, options); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			m.Rollback()
			panic(r)
		}
	}()

	if err := fn(); err != nil {
		m.Rollback()
		return err
	}

	return m.Commit()
}

// getExecutor 获取执行器（事务或连接）
// 根据autoCommit参数和当前事务状态返回合适的执行器
// 如果autoCommit为false且存在活跃事务，返回事务执行器
// 否则返回数据库连接执行器
// 参数:
//   autoCommit: 是否自动提交
// 返回:
//   interface: 执行器接口，可以是*sql.Tx或*sql.DB
func (m *MySQL) getExecutor(autoCommit bool) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if !autoCommit && m.currentTx != nil {
		return m.currentTx
	}
	return m.db
}

// Exec 执行SQL语句
// 执行INSERT、UPDATE、DELETE等不返回结果集的MySQL语句
// 支持事务和非事务模式执行
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   query: 要执行的SQL语句，可包含占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 执行失败时返回错误信息
func (m *MySQL) Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error) {
	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
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
// 自动处理结构体字段到数据库列的映射
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   dest: 目标切片的指针，用于接收查询结果
//   query: 要执行的SELECT语句，可包含占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   error: 查询失败或扫描失败时返回错误信息
func (m *MySQL) Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
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

	err = scanRows(rows, dest)
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
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   dest: 目标结构体的指针，用于接收查询结果
//   query: 要执行的SELECT语句，可包含占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   error: 查询失败、扫描失败或记录不存在时返回错误信息
func (m *MySQL) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
	row := executor.QueryRowContext(ctx, query, args...)
	duration := time.Since(start)

	err := scanRow(row, dest)
	
	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		m.logger.LogSQL(ctx, "SQL单行查询", query, args, err, duration, map[string]interface{}{
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
// 会自动提取结构体字段作为列名和值，支持db tag映射
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   data: 要插入的数据结构体，字段通过db tag映射到数据库列
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 插入记录的自增ID（如果有）
//   error: 插入失败时返回错误信息
func (m *MySQL) Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error) {
	query, args, err := buildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
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
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   data: 包含更新数据的结构体，字段通过db tag映射到数据库列
//   where: WHERE条件语句，可包含占位符
//   args: WHERE条件中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 更新失败时返回错误信息
func (m *MySQL) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool) (int64, error) {
	setClause, setArgs, err := buildUpdateQuery(table, data)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("UPDATE %s SET %s", table, setClause)
	if where != "" {
		query += " WHERE " + where
		setArgs = append(setArgs, args...)
	}

	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
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
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   where: WHERE条件语句，可包含占位符
//   args: WHERE条件中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 删除失败时返回错误信息
func (m *MySQL) Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}

	executor := m.getExecutor(autoCommit)
	
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
	m.logger.LogSQL(ctx, "SQL删除", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// BatchInsert 批量插入记录
// 将切片中的多个数据结构体批量插入到MySQL中
// 使用单条INSERT语句提高插入性能
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   dataSlice: 要插入的数据切片，每个元素都是结构体
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 插入失败时返回错误信息
func (m *MySQL) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	// 获取第一个元素来构建SQL结构
	firstItem := slice.Index(0).Interface()
	columns, _, err := extractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 构建批量插入SQL
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	placeholder := "(" + strings.Join(placeholders, ", ") + ")"

	var allPlaceholders []string
	var allArgs []interface{}

	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		_, values, err := extractColumnsAndValues(item)
		if err != nil {
			return 0, err
		}
		allPlaceholders = append(allPlaceholders, placeholder)
		allArgs = append(allArgs, values...)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(allPlaceholders, ", "))

	executor := m.getExecutor(autoCommit)
	
	start := time.Now()
	result, err := executor.ExecContext(ctx, query, allArgs...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	m.logger.LogSQL(ctx, "SQL批量插入", query, allArgs, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// 工具函数

// buildInsertQuery 构建插入SQL语句
// 根据数据结构体自动生成INSERT语句和参数
// 会提取结构体字段作为列名，使用占位符构建VALUES子句
// 参数:
//   table: 目标表名
//   data: 数据结构体，字段通过db tag映射到数据库列
// 返回:
//   string: 生成的INSERT SQL语句
//   []interface{}: SQL参数值切片
//   error: 构建失败时返回错误信息
func buildInsertQuery(table string, data interface{}) (string, []interface{}, error) {
	columns, values, err := extractColumnsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, values, nil
}

// buildUpdateQuery 构建更新SQL语句的SET子句
// 根据数据结构体自动生成UPDATE语句的SET部分和参数
// 会提取结构体字段作为要更新的列，使用占位符构建SET子句
// 参数:
//   table: 目标表名（此参数在当前实现中未使用，保留用于扩展）
//   data: 包含更新数据的结构体，字段通过db tag映射到数据库列
// 返回:
//   string: 生成的SET子句（如："name = ?, age = ?"）
//   []interface{}: SET子句对应的参数值切片
//   error: 构建失败时返回错误信息
func buildUpdateQuery(table string, data interface{}) (string, []interface{}, error) {
	columns, values, err := extractColumnsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	setParts := make([]string, len(columns))
	for i, column := range columns {
		setParts[i] = column + " = ?"
	}

	return strings.Join(setParts, ", "), values, nil
}

// extractColumnsAndValues 从结构体中提取列名和值
// 通过反射解析结构体，提取可用于数据库操作的列名和对应值
// 支持db tag映射，跳过零值字段和忽略字段
// 参数:
//   data: 要解析的结构体或结构体指针
// 返回:
//   []string: 数据库列名切片
//   []interface{}: 对应的值切片
//   error: 解析失败时返回错误信息
func extractColumnsAndValues(data interface{}) ([]string, []interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("data must be a struct or pointer to struct")
	}

	t := v.Type()
	var columns []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取数据库字段名
		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		// 跳过忽略的字段
		if dbTag == "-" {
			continue
		}

		// 跳过零值字段（可选）
		if isZeroValue(field) {
			continue
		}

		columns = append(columns, dbTag)
		values = append(values, field.Interface())
	}

	return columns, values, nil
}

// isZeroValue 检查值是否为零值
// 判断反射值是否为对应类型的零值，用于跳过空字段
// 支持常见的基本类型、指针类型和时间类型的零值检查
// 参数:
//   v: 要检查的反射值
// 返回:
//   bool: true表示是零值，false表示非零值
func isZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理时间类型
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		// 对于其他结构体类型，使用通用零值检查
		return v.IsZero()
	default:
		return false
	}
}

// scanRows 扫描多行结果到目标切片
// 将SQL查询返回的多行结果扫描到Go切片中
// 自动处理结构体字段到数据库列的映射
// 参数:
//   rows: SQL查询返回的行结果集
//   dest: 目标切片的指针，元素类型应为结构体或结构体指针
// 返回:
//   error: 扫描失败时返回错误信息
func scanRows(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	sliceValue := destValue.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to slice")
	}

	elementType := sliceValue.Type().Elem()
	isPtr := elementType.Kind() == reflect.Ptr
	if isPtr {
		elementType = elementType.Elem()
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		// 创建新的结构体实例
		newElement := reflect.New(elementType)
		
		// 准备扫描目标（包含NULL值安全处理）
		scanTargets, fields := prepareScanTargetsWithFields(newElement.Elem(), columns)
		if len(scanTargets) == 0 {
			return fmt.Errorf("no valid scan targets prepared")
		}

		// 扫描行数据
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		// 处理扫描后的值转换
		if err := processScannedValues(scanTargets, fields); err != nil {
			return err
		}

		// 添加到切片
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, newElement))
		} else {
			sliceValue.Set(reflect.Append(sliceValue, newElement.Elem()))
		}
	}

	return rows.Err()
}

// scanRow 扫描单行结果到目标结构体
// 将SQL查询返回的单行结果扫描到Go结构体中
// 自动按字段顺序进行扫描，支持NULL值安全处理
// 参数:
//   row: SQL查询返回的单行结果
//   dest: 目标结构体的指针
// 返回:
//   error: 扫描失败时返回错误信息
func scanRow(row *sql.Row, dest interface{}) error {
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	structValue := destValue.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	// 这里简化处理，实际应该获取列名
	// 由于sql.Row没有Columns方法，这里需要特殊处理
	// 简化实现：直接按字段顺序扫描
	structType := structValue.Type()
	var scanTargets []interface{}
	var fields []reflect.Value

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		if !field.CanSet() {
			continue
		}
		
		structField := structType.Field(i)
		dbTag := structField.Tag.Get("db")
		if dbTag == "-" {
			continue
		}

		// 创建NULL值安全的扫描目标
		scanTarget := createNullSafeScanTarget(field)
		scanTargets = append(scanTargets, scanTarget)
		fields = append(fields, field)
	}

	if err := row.Scan(scanTargets...); err != nil {
		if err == sql.ErrNoRows {
			return database.ErrRecordNotFound
		}
		return err
	}

	// 处理扫描后的值转换
	return processScannedValues(scanTargets, fields)
}

// prepareScanTargetsWithFields 准备扫描目标并返回对应的字段
// 为scanRows函数提供的增强版本，同时返回扫描目标和对应字段
// 参数:
//   structValue: 目标结构体的反射值
//   columns: 数据库列名切片
// 返回:
//   []interface{}: 扫描目标切片，每个元素对应一个数据库列
//   []reflect.Value: 对应的结构体字段切片
func prepareScanTargetsWithFields(structValue reflect.Value, columns []string) ([]interface{}, []reflect.Value) {
	var scanTargets []interface{}
	var fields []reflect.Value

	for _, column := range columns {
		field, found := findFieldByColumn(structValue, column)
		if !found {
			// 如果找不到对应字段，使用一个丢弃变量
			var discard interface{}
			scanTargets = append(scanTargets, &discard)
			fields = append(fields, reflect.Value{}) // 空值占位
			continue
		}

		if !field.CanSet() {
			// 字段不可设置，使用丢弃变量
			var discard interface{}
			scanTargets = append(scanTargets, &discard)
			fields = append(fields, reflect.Value{}) // 空值占位
			continue
		}

		// 创建NULL值安全的扫描目标
		scanTarget := createNullSafeScanTarget(field)
		scanTargets = append(scanTargets, scanTarget)
		fields = append(fields, field)
	}

	return scanTargets, fields
}

// findFieldByColumn 根据列名查找对应的结构体字段
// 通过db tag或字段名（转小写）匹配数据库列名
// 支持db tag映射，优先使用tag定义的名称
// 参数:
//   structValue: 要搜索的结构体反射值
//   column: 要匹配的数据库列名
// 返回:
//   reflect.Value: 找到的字段反射值
//   bool: 是否找到匹配的字段
func findFieldByColumn(structValue reflect.Value, column string) (reflect.Value, bool) {
	structType := structValue.Type()

	for i := 0; i < structValue.NumField(); i++ {
		field := structValue.Field(i)
		structField := structType.Field(i)

		// 获取数据库字段名
		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		if dbTag == column {
			return field, true
		}
	}

	return reflect.Value{}, false
}

// createNullSafeScanTarget 创建NULL值安全的扫描目标
// 根据字段类型创建相应的sql.NullXXX类型，用于安全扫描可能为NULL的数据库值
// 参数:
//   field: 目标字段的反射值
// 返回:
//   interface{}: 扫描目标，可以是sql.NullString、sql.NullInt64等
func createNullSafeScanTarget(field reflect.Value) interface{} {
	fieldType := field.Type()
	
	switch fieldType.Kind() {
	case reflect.String:
		return &sql.NullString{}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return &sql.NullInt64{}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &sql.NullInt64{} // 使用Int64处理无符号整数
	case reflect.Float32, reflect.Float64:
		return &sql.NullFloat64{}
	case reflect.Bool:
		return &sql.NullBool{}
	case reflect.Ptr:
		// 如果是指针类型，创建对应基础类型的NULL扫描目标
		elemType := fieldType.Elem()
		switch elemType.Kind() {
		case reflect.String:
			return &sql.NullString{}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return &sql.NullInt64{}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return &sql.NullInt64{}
		case reflect.Float32, reflect.Float64:
			return &sql.NullFloat64{}
		case reflect.Bool:
			return &sql.NullBool{}
		default:
			if elemType == reflect.TypeOf(time.Time{}) {
				return &sql.NullTime{}
			}
		}
	case reflect.Struct:
		// 特殊处理时间类型
		if fieldType == reflect.TypeOf(time.Time{}) {
			return &sql.NullTime{}
		}
	}
	
	// 如果无法确定类型，返回通用接口
	var discard interface{}
	return &discard
}

// processScannedValues 处理扫描后的值转换
// 将sql.NullXXX类型的值转换为目标字段类型，处理NULL值
// 参数:
//   scanTargets: 扫描目标切片
//   fields: 目标字段切片
// 返回:
//   error: 转换失败时返回错误信息
func processScannedValues(scanTargets []interface{}, fields []reflect.Value) error {
	for i, scanTarget := range scanTargets {
		if i >= len(fields) {
			continue
		}
		
		field := fields[i]
		if !field.IsValid() || !field.CanSet() {
			continue
		}
		
		// 根据扫描目标类型处理值转换
		switch v := scanTarget.(type) {
		case *sql.NullString:
			if field.Kind() == reflect.Ptr {
				// 处理指针类型字段
				if v.Valid {
					strValue := v.String
					field.Set(reflect.ValueOf(&strValue))
				} else {
					field.Set(reflect.Zero(field.Type()))
				}
			} else {
				// 处理非指针类型字段
				if v.Valid {
					field.SetString(v.String)
				} else {
					field.SetString("")
				}
			}
		case *sql.NullInt64:
			if field.Kind() == reflect.Ptr {
				// 处理指针类型字段
				if v.Valid {
					elemType := field.Type().Elem()
					switch elemType.Kind() {
					case reflect.Int:
						intValue := int(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int8:
						intValue := int8(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int16:
						intValue := int16(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int32:
						intValue := int32(v.Int64)
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Int64:
						intValue := v.Int64
						field.Set(reflect.ValueOf(&intValue))
					case reflect.Uint:
						if v.Int64 >= 0 {
							uintValue := uint(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint8:
						if v.Int64 >= 0 {
							uintValue := uint8(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint16:
						if v.Int64 >= 0 {
							uintValue := uint16(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint32:
						if v.Int64 >= 0 {
							uintValue := uint32(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					case reflect.Uint64:
						if v.Int64 >= 0 {
							uintValue := uint64(v.Int64)
							field.Set(reflect.ValueOf(&uintValue))
						} else {
							field.Set(reflect.Zero(field.Type()))
						}
					}
				} else {
					field.Set(reflect.Zero(field.Type()))
				}
			} else {
				// 处理非指针类型字段
				if v.Valid {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(v.Int64)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						if v.Int64 >= 0 {
							field.SetUint(uint64(v.Int64))
						} else {
							field.SetUint(0)
						}
					}
				} else {
					switch field.Kind() {
					case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
						field.SetInt(0)
					case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
						field.SetUint(0)
					}
				}
			}
		case *sql.NullFloat64:
			if v.Valid {
				field.SetFloat(v.Float64)
			} else {
				field.SetFloat(0)
			}
		case *sql.NullBool:
			if v.Valid {
				field.SetBool(v.Bool)
			} else {
				field.SetBool(false)
			}
		case *sql.NullTime:
			if v.Valid {
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(reflect.ValueOf(v.Time))
				} else if field.Type() == reflect.TypeOf(&time.Time{}) {
					field.Set(reflect.ValueOf(&v.Time))
				}
			} else {
				if field.Type() == reflect.TypeOf(time.Time{}) {
					field.Set(reflect.ValueOf(time.Time{}))
				} else if field.Type() == reflect.TypeOf(&time.Time{}) {
					field.Set(reflect.ValueOf((*time.Time)(nil)))
				}
			}
		default:
			// 对于其他类型，不做处理
		}
	}
	
	return nil
}
