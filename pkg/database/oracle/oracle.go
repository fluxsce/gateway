//go:build !no_oracle
// +build !no_oracle

package oracle

import (
	"context"
	"database/sql"
	"fmt"
	"gohub/pkg/database"
	"gohub/pkg/database/dblogger"
	"gohub/pkg/database/dbtypes"
	"gohub/pkg/database/sqlutils"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/godror/godror"
)

// 注册Oracle驱动
func init() {
	database.Register(database.DriverOracle, func() database.Database {
		return &Oracle{}
	})
	
	// 注册Oracle 11g驱动（使用相同的实现但标识为不同的驱动类型，用于分页语法区分）
	database.Register(dbtypes.DriverOracle11g, func() database.Database {
		oracle := &Oracle{}
		oracle.isOracle11g = true
		return oracle
	})
}

// Oracle Oracle数据库实现
// 核心特性:
// 1. 统一的数据库接口实现 - 符合database.Database接口规范
// 2. 灵活的事务管理 - 支持不同隔离级别和可定制的事务选项
// 3. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
// 4. 智能日志记录 - 支持慢查询检测和SQL执行日志
// 5. 结构体映射 - 自动将Go结构体与数据库表映射
// 6. 带选项的操作 - 每个数据库操作都有带选项的版本，支持自定义执行行为
// 7. 活跃事务追踪 - 通过内部映射跟踪所有活跃事务
// 8. Oracle特性支持 - 支持Oracle特有的序列、过程调用等特性，自动转换占位符
type Oracle struct {
	db       *sql.DB
	config   *database.DbConfig
	logger   *dblogger.DBLogger
	mu       sync.RWMutex
	currentTx *sql.Tx // 当前活跃的事务
	isOracle11g bool
}

// convertPlaceholders 转换SQL占位符为Oracle格式
// 将标准?占位符转换为Oracle的:1,:2格式
func (o *Oracle) convertPlaceholders(qry string) string {
	n := strings.Count(qry, "?")
	if n == 0 {
		return qry
	}
	nLog10, x := 1, 10
	for n > x {
		nLog10++
		x *= 10
	}
	//fmt.Println("\n## n:", n, "x:", x, "nLog10:", nLog10)
	num := make([]byte, 0, nLog10)
	var buf strings.Builder
	buf.Grow(len(qry) + n*(nLog10))
	var idx int64
	for i := strings.IndexByte(qry, '?'); i >= 0; i = strings.IndexByte(qry, '?') {
		buf.WriteString(qry[:i])
		qry = qry[i+1:]
		buf.WriteByte(':')
		idx++
		num = strconv.AppendInt(num[:0], idx, 10)
		buf.Write(num)
	}
	buf.WriteString(qry)
	return buf.String()
}

// Connect 连接到Oracle数据库
// 建立Oracle数据库连接，配置连接池参数，并验证连接可用性
// 会根据配置设置最大连接数、空闲连接数、连接生命周期等参数
// 参数:
//   config: Oracle数据库配置，包含DSN、连接池设置、日志配置等
// 返回:
//   error: 连接建立失败时返回错误信息
func (o *Oracle) Connect(config *database.DbConfig) error {
	o.config = config
	o.logger = dblogger.NewDBLogger(config)

	// 使用背景上下文进行连接日志记录
	o.logger.LogConnecting(context.Background(), database.DriverOracle, config.DSN)

	// 打开数据库连接
	db, err := sql.Open("godror", config.DSN)
	if err != nil {
		o.logger.LogError(context.Background(), "打开Oracle连接", err)
		return fmt.Errorf("failed to open Oracle connection: %w", err)
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
		o.logger.LogPing(context.Background(), err)
		return fmt.Errorf("Oracle connection test failed: %w", err)
	}

	o.db = db
	o.logger.LogConnected(context.Background(), database.DriverOracle, map[string]any{
		"maxOpenConns":    maxOpenConns,
		"maxIdleConns":    maxIdleConns,
		"connMaxLifetime": connMaxLifetime.String(),
		"connMaxIdleTime": connMaxIdleTime.String(),
		"placeholders":    "Oracle :1,:2 format (auto-converted)",
	})

	return nil
}

// Close 关闭数据库连接
// 关闭Oracle数据库连接，释放相关资源
// 如果存在活跃事务，会先回滚事务再关闭连接
// 返回:
//   error: 关闭连接失败时返回错误信息
func (o *Oracle) Close() error {
	if o.currentTx != nil {
		o.currentTx.Rollback()
		o.currentTx = nil
	}
	if o.db != nil {
		o.logger.LogDisconnect(context.Background(), database.DriverOracle)
		return o.db.Close()
	}
	return nil
}

// DSN 返回数据库连接字符串
// 获取当前Oracle连接使用的数据源名称
// 返回值会被处理以隐藏敏感信息（如密码）
// 返回:
//   string: 处理后的DSN字符串，隐藏敏感信息
func (o *Oracle) DSN() string {
	if o.config == nil {
		return ""
	}
	return dblogger.MaskDSN(o.config.DSN)
}

// DB 返回底层的sql.DB实例
// 获取Oracle连接底层的标准库sql.DB实例
// 用于需要直接访问底层数据库连接的场景
// 返回:
//   *sql.DB: 底层的sql.DB实例
func (o *Oracle) DB() *sql.DB {
	return o.db
}

// DriverName 返回数据库驱动名称
// 获取当前数据库使用的驱动名称标识
// 返回:
//   string: 固定返回"oracle"或"oracle11g"
func (o *Oracle) DriverName() string {
	if o.isOracle11g {
		return dbtypes.DriverOracle11g
	}
	return database.DriverOracle
}

// GetDriver 获取数据库驱动类型
// 实现Database接口，返回Oracle驱动标识
// 返回:
//   string: Oracle驱动类型标识
func (o *Oracle) GetDriver() string {
	if o.isOracle11g {
		return dbtypes.DriverOracle11g
	}
	return database.DriverOracle
}

// GetName 获取数据库连接名称
// 实现Database接口，返回当前连接的名称
// 返回:
//   string: 数据库连接名称，如果配置为空则返回空字符串
func (o *Oracle) GetName() string {
	if o.config == nil {
		return ""
	}
	return o.config.Name
}

// SetName 设置数据库连接名称
// 用于在创建连接后设置连接名称标识
// 参数:
//   name: 连接名称
func (o *Oracle) SetName(name string) {
	if o.config != nil {
		o.config.Name = name
	}
}

// Ping 测试数据库连接
// 向Oracle服务器发送ping请求，验证连接状态
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
// 返回:
//   error: 连接异常时返回错误信息
func (o *Oracle) Ping(ctx context.Context) error {
	err := o.db.PingContext(ctx)
	o.logger.LogPing(ctx, err)
	return err
}

// BeginTx 开始事务
// 启动一个新的Oracle事务，支持指定隔离级别和只读属性
// 如果已有活跃事务，会返回错误
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   options: 事务选项，包含隔离级别和只读设置
// 返回:
//   error: 开始事务失败时返回错误信息
func (o *Oracle) BeginTx(ctx context.Context, options *database.TxOptions) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.currentTx != nil {
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

	tx, err := o.db.BeginTx(ctx, sqlTxOpts)
	if err != nil {
		o.logger.LogTx(ctx, "开始", err)
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	o.currentTx = tx
	o.logger.LogTx(ctx, "开始", nil)
	return nil
}

// Commit 提交事务
// 提交当前活跃的Oracle事务，使所有未提交的更改生效
// 如果没有活跃事务，会返回错误
// 返回:
//   error: 提交事务失败时返回错误信息
func (o *Oracle) Commit() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.currentTx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := o.currentTx.Commit()
	o.currentTx = nil
	o.logger.LogTx(context.Background(), "提交", err)
	
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// Rollback 回滚事务
// 回滚当前活跃的Oracle事务，撤销所有未提交的更改
// 如果没有活跃事务，会返回错误
// 返回:
//   error: 回滚事务失败时返回错误信息
func (o *Oracle) Rollback() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.currentTx == nil {
		return fmt.Errorf("no active transaction")
	}

	err := o.currentTx.Rollback()
	o.currentTx = nil
	o.logger.LogTx(context.Background(), "回滚", err)
	
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// InTx 在事务中执行函数
// 自动管理Oracle事务的生命周期
// 如果函数正常返回，自动提交事务
// 如果函数返回错误或发生panic，自动回滚事务
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   options: 事务选项，包含隔离级别和只读设置
//   fn: 在事务中执行的函数，返回error表示是否成功
// 返回:
//   error: 事务执行失败时返回错误信息
func (o *Oracle) InTx(ctx context.Context, options *database.TxOptions, fn func() error) error {
	if err := o.BeginTx(ctx, options); err != nil {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			o.Rollback()
			panic(r)
		}
	}()

	if err := fn(); err != nil {
		o.Rollback()
		return err
	}

	return o.Commit()
}

// getExecutor 获取执行器（事务或连接）
// 根据autoCommit参数和当前事务状态返回合适的执行器
// 如果autoCommit为false且存在活跃事务，返回事务执行器
// 否则返回数据库连接执行器
// 参数:
//   autoCommit: 是否自动提交
// 返回:
//   interface: 执行器接口，可以是*sql.Tx或*sql.DB
func (o *Oracle) getExecutor(autoCommit bool) interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
} {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if !autoCommit && o.currentTx != nil {
		return o.currentTx
	}
	return o.db
}

// Exec 执行SQL语句
// 执行INSERT、UPDATE、DELETE等不返回结果集的Oracle语句
// 支持事务和非事务模式执行
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   query: 要执行的SQL语句，使用标准的?占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 执行失败时返回错误信息
func (o *Oracle) Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error) {
	executor := o.getExecutor(autoCommit)
	
	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)
	
	start := time.Now()
	result, err := executor.ExecContext(ctx, convertedQuery, args...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	o.logger.LogSQL(ctx, "SQL执行", query, args, err, duration, extra)

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
//   query: 要执行的SELECT语句，使用标准的?占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   error: 查询失败或扫描失败时返回错误信息
func (o *Oracle) Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := o.getExecutor(autoCommit)
	
	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)
	
	start := time.Now()
	rows, err := executor.QueryContext(ctx, convertedQuery, args...)
	duration := time.Since(start)

	if err != nil {
		if err != sql.ErrNoRows {
			extra := map[string]interface{}{
				"rowCount": 0,
			}
			o.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, extra)
		}
		return err
	}
	defer rows.Close()

	err = sqlutils.ScanRows(rows, dest)
	rowCount := reflect.ValueOf(dest).Elem().Len()

	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		extra := map[string]interface{}{
			"rowCount": 0,
		}
		o.logger.LogSQL(ctx, "SQL查询", query, args, err, duration, extra)
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": rowCount,
	}
	o.logger.LogSQL(ctx, "SQL查询", query, args, nil, duration, extra)

	return err
}

// QueryOne 查询单条记录
// 执行SELECT语句并将结果扫描到目标结构体中
// 如果查询不到记录，返回ErrRecordNotFound错误
// 使用智能字段映射，支持数据库列数与结构体字段数不匹配的情况
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   dest: 目标结构体的指针，用于接收查询结果
//   query: 要执行的SELECT语句，使用标准的?占位符
//   args: SQL语句中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   error: 查询失败、扫描失败或记录不存在时返回错误信息
func (o *Oracle) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := o.getExecutor(autoCommit)

	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)

	start := time.Now()
	rows, err := executor.QueryContext(ctx, convertedQuery, args...)
	duration := time.Since(start)
	
	if err != nil {
		extra := map[string]interface{}{
			"rowCount": 0,
		}
		o.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, extra)
		return err
	}

	// 使用智能扫描方式处理单行结果，支持字段数量不匹配
	err = sqlutils.ScanOneRow(rows, dest)
	
	// 只有在有错误且不是未找到记录时才记录错误
	if err != nil && err != database.ErrRecordNotFound {
		extra := map[string]interface{}{
			"rowCount": 0,
		}
		o.logger.LogSQL(ctx, "SQL单行查询错误", query, args, err, duration, extra)
		return err
	}

	// 记录成功的查询及影响行数
	extra := map[string]interface{}{
		"rowCount": map[bool]int{true: 1, false: 0}[err == nil],
	}
	o.logger.LogSQL(ctx, "SQL单行查询", query, args, nil, duration, extra)

	return err
}

// Insert 插入记录
// 根据提供的数据结构体自动构建INSERT语句并执行
// 会自动提取结构体字段作为列名和值，支持db tag映射
// 对于Oracle，会自动处理RETURNING子句获取自增ID（通过序列）
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   data: 要插入的数据结构体，字段通过db tag映射到数据库列
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 插入记录的自增ID（如果有）
//   error: 插入失败时返回错误信息
func (o *Oracle) Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error) {
	query, args, err := sqlutils.BuildInsertQueryForOracle(table, data)
	if err != nil {
		return 0, err
	}

	executor := o.getExecutor(autoCommit)
	
	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)
	
	start := time.Now()
	result, err := executor.ExecContext(ctx, convertedQuery, args...)
	duration := time.Since(start)

	var lastInsertId int64
	var rowsAffected int64
	if err == nil {
		// Oracle不直接支持LastInsertId，通常需要使用RETURNING子句或序列
		// 这里先尝试获取，如果不支持会返回0
		lastInsertId, _ = result.LastInsertId()
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}
	o.logger.LogSQL(ctx, "SQL插入", query, args, err, duration, extra)

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
//   where: WHERE条件语句，使用标准的?占位符
//   args: WHERE条件中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 更新失败时返回错误信息
func (o *Oracle) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool) (int64, error) {
	setClause, setArgs, err := sqlutils.BuildUpdateQueryForOracle(table, data)
	if err != nil {
		return 0, err
	}

	query := fmt.Sprintf("UPDATE %s SET %s", table, setClause)
	finalArgs := setArgs // 先添加SET子句的参数
	
	if where != "" {
		query += " WHERE " + where
		finalArgs = append(finalArgs, args...) // 再添加WHERE子句的参数
	}

	executor := o.getExecutor(autoCommit)
	
	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)
	
	start := time.Now()
	result, err := executor.ExecContext(ctx, convertedQuery, finalArgs...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	o.logger.LogSQL(ctx, "SQL更新", query, finalArgs, err, duration, extra)

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
//   where: WHERE条件语句，使用标准的?占位符
//   args: WHERE条件中占位符对应的参数值
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 删除失败时返回错误信息
func (o *Oracle) Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error) {
	query := fmt.Sprintf("DELETE FROM %s", table)
	if where != "" {
		query += " WHERE " + where
	}

	executor := o.getExecutor(autoCommit)
	
	// 转换占位符为Oracle格式
	convertedQuery := o.convertPlaceholders(query)
	
	start := time.Now()
	result, err := executor.ExecContext(ctx, convertedQuery, args...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": rowsAffected,
	}
	o.logger.LogSQL(ctx, "SQL删除", query, args, err, duration, extra)

	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

// BatchInsert 批量插入记录
// 将切片中的多个数据结构体批量插入到Oracle中
// 使用多个单独的INSERT语句进行批量插入（更稳定的方式）
// 参数:
//   ctx: 上下文，用于控制请求超时和取消
//   table: 目标表名
//   dataSlice: 要插入的数据切片，每个元素都是结构体
//   autoCommit: true-自动提交, false-在当前事务中执行
// 返回:
//   int64: 受影响的行数
//   error: 插入失败时返回错误信息
func (o *Oracle) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	if slice.Len() == 0 {
		return 0, nil
	}

	executor := o.getExecutor(autoCommit)
	var totalRowsAffected int64
	
	start := time.Now()
	
	// 使用简单的逐条插入方式，更加稳定
	for i := 0; i < slice.Len(); i++ {
		item := slice.Index(i).Interface()
		query, args, err := sqlutils.BuildInsertQueryForOracle(table, item)
		if err != nil {
			return 0, err
		}
		
		// 转换占位符为Oracle格式
		convertedQuery := o.convertPlaceholders(query)
		
		result, err := executor.ExecContext(ctx, convertedQuery, args...)
		if err != nil {
			o.logger.LogSQL(ctx, "SQL批量插入失败", query, args, err, time.Since(start), map[string]interface{}{
				"rowsAffected": totalRowsAffected,
				"failedIndex":  i,
			})
			return totalRowsAffected, err
		}
		
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	
	duration := time.Since(start)

	// 记录日志
	extra := map[string]interface{}{
		"rowsAffected": totalRowsAffected,
		"batchSize":    slice.Len(),
	}
	o.logger.LogSQL(ctx, "SQL批量插入", fmt.Sprintf("批量插入到 %s", table), nil, nil, duration, extra)

	return totalRowsAffected, nil
} 