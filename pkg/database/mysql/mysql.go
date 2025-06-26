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
	"gohub/pkg/database/sqlutils"

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
//   ctx: 上下文，用于控制请求超时和取消
//   dest: 目标结构体的指针，用于接收查询结果
//   query: 要执行的SELECT语句，可包含占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   error: 查询失败、扫描失败或记录不存在时返回错误信息
func (m *MySQL) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := m.getExecutor(autoCommit)
	/** 
	 * 方法使用的是 QueryRowContext，它返回的是 *sql.Row，而 sql.Row 没有 Columns() 方法。
	 * 我们需要修改 QueryOne 方法使用 QueryContext 替代，这样就能获取列信息并处理字段不匹配的问题
	*/

	start := time.Now()
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
	query, args, err := sqlutils.BuildInsertQuery(table, data)
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
	setClause, setArgs, err := sqlutils.BuildUpdateQuery(table, data)
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
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
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
		_, values, err := sqlutils.ExtractColumnsAndValues(item)
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
// 注意：SQL格式化和结果处理相关的工具函数已移动到 sqlutils 包中
// 请使用 sqlutils.BuildInsertQuery, sqlutils.ScanRows 等函数

// 重构说明：
// 1. SQL格式化功能已移动到 pkg/database/sqlutils/sql_format.go
//    - BuildInsertQuery: 构建INSERT语句
//    - BuildUpdateQuery: 构建UPDATE语句的SET子句
//    - ExtractColumnsAndValues: 从结构体提取列名和值
//    - IsZeroValue: 检查值是否为零值
//
// 2. 结果处理功能已移动到 pkg/database/sqlutils/result_format.go
//    - ScanRows: 扫描多行结果到切片
//    - ScanRow: 扫描单行结果到结构体
//    - PrepareScanTargetsWithFields: 准备扫描目标
//    - FindFieldByColumn: 根据列名查找字段
//    - CreateNullSafeScanTarget: 创建NULL值安全的扫描目标
//    - ProcessScannedValues: 处理扫描后的值转换
//    - ScanRowsTraditional: 传统方式扫描多行结果
//    - ScanRowsWithInterfaceSlice: 接口切片方式扫描多行结果
