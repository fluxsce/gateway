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
		return &MySQL{
			txs: make(map[*sql.Tx]bool), // 确保每个实例都有自己独立的事务映射
		}
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
	db     *sql.DB
	config *database.DbConfig
	mu     sync.RWMutex
	txs    map[*sql.Tx]bool // 管理活跃的事务
	logger *dblogger.DBLogger
}

// Connect implements database.Database
// 连接到MySQL数据库
func (m *MySQL) Connect(config *database.DbConfig) error {
	m.config = config
	m.logger = dblogger.NewDBLogger(config)

	// 记录连接开始
	m.logger.LogConnecting(database.DriverMySQL, config.DSN)

	// 打开数据库连接
	db, err := sql.Open("mysql", config.DSN)
	if err != nil {
		m.logger.LogError("打开MySQL连接", err)
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

	// 记录连接池设置
	poolSettings := map[string]any{
		"maxOpenConns":    maxOpenConns,
		"maxIdleConns":    maxIdleConns,
		"connMaxLifetime": connMaxLifetime.String(),
		"connMaxIdleTime": connMaxIdleTime.String(),
	}

	// 检查连接是否正常
	if err := db.Ping(); err != nil {
		m.logger.LogPing(err)
		return fmt.Errorf("MySQL connection test failed: %w", err)
	}

	// 应用配置
	m.db = db
	m.txs = make(map[*sql.Tx]bool) // 初始化事务映射

	// 记录连接成功
	m.logger.LogConnected(database.DriverMySQL, poolSettings)

	return nil
}

// Close 关闭数据库连接
func (m *MySQL) Close() error {
	if m.db != nil {
		m.logger.LogDisconnect(database.DriverMySQL)
		return m.db.Close()
	}
	return nil
}

// DSN 返回数据库连接字符串
// 返回值会被处理以隐藏敏感信息
func (m *MySQL) DSN() string {
	if m.config == nil {
		return ""
	}
	// 导入数据库logger包只是为了访问这个函数
	return dblogger.MaskDSN(m.config.DSN)
}

// DB 返回底层的sql.DB实例
func (m *MySQL) DB() *sql.DB {
	return m.db
}

// DriverName 返回数据库驱动名称
func (m *MySQL) DriverName() string {
	return database.DriverMySQL
}

// GetDriver 获取数据库驱动类型
func (m *MySQL) GetDriver() string {
	return database.DriverMySQL
}

// GetName 获取数据库连接名称
func (m *MySQL) GetName() string {
	if m.config == nil {
		return ""
	}
	return m.config.Name
}

// Ping 测试数据库连接
func (m *MySQL) Ping(ctx context.Context) error {
	err := m.db.PingContext(ctx)
	m.logger.LogPing(err)
	return err
}

// BeginTx 开始事务
func (m *MySQL) BeginTx(ctx context.Context, options ...database.TxOption) (database.Transaction, error) {
	// 应用传入的事务选项
	txOpts := database.NewTxOptions(options...)

	// 创建SQL事务选项，用于传递给底层数据库
	sqlTxOpts := &sql.TxOptions{
		ReadOnly: txOpts.ReadOnly, // 设置事务是否为只读
	}

	// 根据隔离级别设置对应的SQL隔离级别
	switch txOpts.Isolation {
	case 1:
		sqlTxOpts.Isolation = sql.LevelReadUncommitted // 未提交读：可以读取未提交的数据
	case 2:
		sqlTxOpts.Isolation = sql.LevelReadCommitted // 提交读：只能读取已提交的数据
	case 3:
		sqlTxOpts.Isolation = sql.LevelRepeatableRead // 可重复读：确保在事务期间读取的数据不会改变
	case 4:
		sqlTxOpts.Isolation = sql.LevelSerializable // 可串行化：最高隔离级别，完全序列化事务执行
	default:
		sqlTxOpts.Isolation = sql.LevelDefault // 默认隔离级别
	}

	// 使用设置的选项开启底层SQL事务
	sqlTx, err := m.db.BeginTx(ctx, sqlTxOpts)
	if err != nil {
		m.logger.LogTx("开始", err)
		return nil, fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	// 记录开始事务日志
	m.logger.LogTx("开始", nil)

	// 注册事务，用于跟踪活跃事务
	m.mu.Lock()
	m.txs[sqlTx] = true
	m.mu.Unlock()

	// 创建事务封装对象，提供统一的接口
	tx := &MySQLTransaction{
		tx:     sqlTx,
		config: m.config,
		parent: m,
		logger: m.logger,
	}

	return tx, nil
}

// WithTx 使用已有事务执行函数
func (m *MySQL) WithTx(tx database.Transaction, fn func(tx database.Transaction) error) error {
	// 如果传入的事务为nil，则创建新事务
	var err error
	var myTx *MySQLTransaction
	var newTx bool

	if tx == nil {
		// 创建新事务
		newTx = true
		tx, err = m.BeginTx(context.Background())
		if err != nil {
			return err
		}

		// 尝试类型断言以获取内部事务对象
		var ok bool
		myTx, ok = tx.(*MySQLTransaction)
		if !ok {
			tx.Rollback() // 类型断言失败，回滚事务
			return fmt.Errorf("invalid transaction type")
		}
	} else {
		// 尝试类型断言确认传入的是MySQL事务
		var ok bool
		myTx, ok = tx.(*MySQLTransaction)
		if !ok {
			return fmt.Errorf("invalid transaction type")
		}

		// 验证事务状态 - 检查事务是否仍然在活跃事务映射中
		m.mu.RLock()
		_, exists := m.txs[myTx.tx]
		m.mu.RUnlock()

		if !exists && !newTx {
			return fmt.Errorf("%w: transaction is already ended or invalid", database.ErrTransaction)
		}
	}

	// 执行传入的事务函数
	if err := fn(tx); err != nil {
		// 如果函数执行失败，回滚事务
		// 验证事务仍然有效
		m.mu.RLock()
		_, exists := m.txs[myTx.tx]
		m.mu.RUnlock()

		if exists {
			_ = tx.Rollback()
		}
		return err
	}

	// 函数执行成功，提交事务
	// 先验证事务仍然有效
	m.mu.RLock()
	_, exists := m.txs[myTx.tx]
	m.mu.RUnlock()

	if !exists {
		return fmt.Errorf("%w: transaction is already ended or invalid", database.ErrTransaction)
	}

	return tx.Commit()
}

// Exec 直接执行SQL（无事务）
func (m *MySQL) Exec(ctx context.Context, query string, args []interface{}) (int64, error) {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 执行SQL语句
	result, err := m.db.ExecContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	m.logger.LogSQL("SQL执行", query, args, err, elapsed)

	if err != nil {
		return 0, err
	}

	// 返回受影响的行数
	return result.RowsAffected()
}

// ExecWithOptions 带选项执行SQL语句
func (m *MySQL) ExecWithOptions(ctx context.Context, query string, args []interface{}, options ...database.ExecOption) (int64, error) {
	// 应用传入的执行选项
	execOpts := database.NewExecOptions(m.config, options...)

	// 判断是否使用事务
	if *execOpts.UseTransaction {
		// 创建新事务
		tx, err := m.BeginTx(ctx)
		if err != nil {
			return 0, err
		}

		// 在事务中执行SQL
		result, err := tx.Exec(ctx, query, args)
		if err != nil {
			// 执行错误，回滚事务
			_ = tx.Rollback()
			return 0, err
		}

		// 执行成功，提交事务
		if err := tx.Commit(); err != nil {
			return 0, err
		}

		return result, nil
	}

	// 不使用事务，直接执行SQL
	return m.Exec(ctx, query, args)
}

// Query 查询多条记录（无事务）
func (m *MySQL) Query(ctx context.Context, dest interface{}, query string, args []interface{}) error {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 执行查询
	rows, err := m.db.QueryContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	m.logger.LogSQL("SQL查询", query, args, err, elapsed)

	if err != nil {
		return err
	}
	// 确保关闭结果集，防止资源泄漏
	defer rows.Close()

	// 扫描结果集到目标结构体切片
	err = scanRows(rows, dest)

	// 获取结果数量
	destVal := reflect.ValueOf(dest).Elem()
	count := 0
	if destVal.Kind() == reflect.Slice {
		count = destVal.Len()
	}

	// 记录查询结果
	m.logger.LogQueryResult("SQL查询", count, err)

	return err
}

// QueryWithOptions 带选项查询多条记录
func (m *MySQL) QueryWithOptions(ctx context.Context, dest interface{}, query string, args []interface{}, options ...database.QueryOption) error {
	// 应用传入的查询选项
	queryOpts := database.NewQueryOptions(m.config, options...)

	// 判断是否使用事务
	if *queryOpts.UseTransaction {
		// 创建新事务
		tx, err := m.BeginTx(ctx)
		if err != nil {
			return err
		}

		// 在事务中执行查询
		err = tx.Query(ctx, dest, query, args)
		if err != nil {
			// 查询失败，回滚事务
			_ = tx.Rollback()
			return err
		}

		// 查询成功，提交事务
		return tx.Commit()
	}

	// 不使用事务，直接查询
	return m.Query(ctx, dest, query, args)
}

// QueryOne 查询单条记录（无事务）
func (m *MySQL) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}) error {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 执行单行查询
	row := m.db.QueryRowContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	m.logger.LogSQL("SQL查询", query, args, nil, elapsed)

	// 扫描单行结果到目标结构体
	err := scanRow(row, dest)

	// 记录查询结果
	var count int
	if err != nil && err != database.ErrRecordNotFound {
		count = 0
	} else {
		count = 1
	}

	m.logger.LogQueryResult("SQL单行查询", count, err)

	return err
}

// QueryOneWithOptions 带选项查询单条记录
func (m *MySQL) QueryOneWithOptions(ctx context.Context, dest interface{}, query string, args []interface{}, options ...database.QueryOption) error {
	// 应用传入的查询选项
	queryOpts := database.NewQueryOptions(m.config, options...)

	// 判断是否使用事务
	if *queryOpts.UseTransaction {
		// 创建新事务
		tx, err := m.BeginTx(ctx)
		if err != nil {
			return err
		}

		// 在事务中执行单行查询
		err = tx.QueryOne(ctx, dest, query, args)
		if err != nil {
			// 查询失败，回滚事务
			_ = tx.Rollback()
			return err
		}

		// 查询成功，提交事务
		return tx.Commit()
	}

	// 不使用事务，直接查询
	return m.QueryOne(ctx, dest, query, args)
}

// Insert 插入记录（无事务）
func (m *MySQL) Insert(ctx context.Context, table string, data interface{}) (int64, error) {
	// 构建INSERT语句和参数
	query, args, err := buildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 执行插入操作
	return m.Exec(ctx, query, args)
}

// InsertWithOptions 带选项插入记录
func (m *MySQL) InsertWithOptions(ctx context.Context, table string, data interface{}, options ...database.ExecOption) (int64, error) {
	// 构建INSERT语句和参数
	query, args, err := buildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 使用选项执行插入操作
	return m.ExecWithOptions(ctx, query, args, options...)
}

// Update 更新记录（无事务）
func (m *MySQL) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}) (int64, error) {
	// 构建基础UPDATE语句和参数
	query, updateArgs, err := buildUpdateQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 添加WHERE子句
	if where != "" {
		query = fmt.Sprintf("%s WHERE %s", query, where)
	}

	// 合并参数：先是更新字段的值，然后是WHERE条件的参数
	allArgs := append(updateArgs, args...)

	// 执行更新操作
	return m.Exec(ctx, query, allArgs)
}

// UpdateWithOptions 带选项更新记录
func (m *MySQL) UpdateWithOptions(ctx context.Context, table string, data interface{}, where string, args []interface{}, options ...database.ExecOption) (int64, error) {
	// 构建基础UPDATE语句和参数
	query, updateArgs, err := buildUpdateQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 添加WHERE子句
	if where != "" {
		query = fmt.Sprintf("%s WHERE %s", query, where)
	}

	// 合并参数：先是更新字段的值，然后是WHERE条件的参数
	allArgs := append(updateArgs, args...)

	// 使用选项执行更新操作
	return m.ExecWithOptions(ctx, query, allArgs, options...)
}

// Delete 删除记录（无事务）
func (m *MySQL) Delete(ctx context.Context, table string, where string, args []any) (int64, error) {
	// 构建基础DELETE语句
	query := "DELETE FROM " + table

	// 添加WHERE子句，避免误删除所有数据
	if where != "" {
		query = query + " WHERE " + where
	}

	// 执行删除操作
	return m.Exec(ctx, query, args)
}

// DeleteWithOptions 带选项删除记录
func (m *MySQL) DeleteWithOptions(ctx context.Context, table string, where string, args []any, options ...database.ExecOption) (int64, error) {
	// 构建基础DELETE语句
	query := "DELETE FROM " + table

	// 添加WHERE子句，避免误删除所有数据
	if where != "" {
		query = query + " WHERE " + where
	}

	// 使用选项执行删除操作
	return m.ExecWithOptions(ctx, query, args, options...)
}

// BatchInsert 批量插入多条记录
//
// 参数:
// - ctx: 上下文
// - table: 表名
// - dataSlice: 包含要插入数据的结构体切片
//
// 返回:
// - 受影响的行数
// - 错误信息
//
// 用法示例:
//
//	users := []User{
//	  {Name: "用户1", Age: 20},
//	  {Name: "用户2", Age: 30},
//	}
//	db.BatchInsert(ctx, "users", users)
func (m *MySQL) BatchInsert(ctx context.Context, table string, dataSlice interface{}) (int64, error) {
	// 获取反射值
	sliceVal := reflect.ValueOf(dataSlice)
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	// 空切片直接返回
	sliceLen := sliceVal.Len()
	if sliceLen == 0 {
		return 0, nil
	}

	// 获取第一个元素，用于提取字段信息
	firstElem := sliceVal.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}
	if firstElem.Kind() != reflect.Struct {
		return 0, fmt.Errorf("slice elements must be structs or pointers to structs")
	}

	// 提取字段信息
	typ := firstElem.Type()
	var columns []string
	fieldIndices := make([]int, 0)

	// 遍历结构体的所有字段，获取db标签
	for i := 0; i < firstElem.NumField(); i++ {
		field := typ.Field(i)

		// 获取db标签
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue // 跳过没有db标签或标记为忽略的字段
		}

		// 解析标签
		parts := strings.Split(tag, ",")
		column := parts[0] // 第一部分是列名

		columns = append(columns, column)
		fieldIndices = append(fieldIndices, i)
	}

	if len(columns) == 0 {
		return 0, fmt.Errorf("no valid columns found in the struct")
	}

	// 构建INSERT语句的前半部分
	placeholderGroup := "(" + strings.Repeat("?,", len(columns)-1) + "?)"
	allPlaceholders := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		allPlaceholders[i] = placeholderGroup
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(allPlaceholders, ","),
	)

	// 准备所有参数值
	values := make([]interface{}, 0, sliceLen*len(columns))
	for i := 0; i < sliceLen; i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		// 提取字段值
		for _, fieldIdx := range fieldIndices {
			fieldValue := elem.Field(fieldIdx).Interface()
			values = append(values, fieldValue)
		}
	}

	// 执行批量插入
	return m.Exec(ctx, query, values)
}

// MySQLTransaction MySQL事务实现
// 特点:
// 1. 完整事务管理 - 提供Commit和Rollback操作
// 2. 统一SQL执行接口 - 实现所有database.Transaction方法
// 3. 性能监控 - 记录所有SQL操作的执行时间
// 4. 错误追踪 - 详细记录事务中的SQL错误
// 5. 自动清理 - 事务结束时自动从MySQL实例的活跃事务映射中移除
type MySQLTransaction struct {
	tx     *sql.Tx            // 底层SQL事务对象
	config *database.DbConfig // 数据库配置
	parent *MySQL             // 父MySQL实例，用于事务管理
	logger *dblogger.DBLogger // 日志记录器
}

// Exec 执行SQL语句
func (t *MySQLTransaction) Exec(ctx context.Context, query string, args []interface{}) (int64, error) {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 在事务中执行SQL语句
	result, err := t.tx.ExecContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	t.logger.LogSQL("事务SQL执行", query, args, err, elapsed)

	if err != nil {
		return 0, err
	}

	// 返回受影响的行数
	return result.RowsAffected()
}

// Query 查询多条记录
func (t *MySQLTransaction) Query(ctx context.Context, dest interface{}, query string, args []interface{}) error {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 在事务中执行查询
	rows, err := t.tx.QueryContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	t.logger.LogSQL("事务SQL查询", query, args, err, elapsed)

	if err != nil {
		return err
	}
	// 确保关闭结果集，防止资源泄漏
	defer rows.Close()

	// 扫描结果集到目标结构体切片
	err = scanRows(rows, dest)

	// 获取结果数量
	destVal := reflect.ValueOf(dest).Elem()
	count := 0
	if destVal.Kind() == reflect.Slice {
		count = destVal.Len()
	}

	// 记录查询结果
	t.logger.LogQueryResult("事务SQL查询", count, err)

	return err
}

// QueryOne 查询单条记录
func (t *MySQLTransaction) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}) error {
	// 记录开始时间，用于计算SQL执行耗时
	start := time.Now()

	// 在事务中执行单行查询
	row := t.tx.QueryRowContext(ctx, query, args...)

	// 计算执行耗时
	elapsed := time.Since(start)

	// 记录SQL执行日志
	t.logger.LogSQL("事务SQL查询", query, args, nil, elapsed)

	// 扫描单行结果到目标结构体
	err := scanRow(row, dest)

	// 记录查询结果
	var count int
	if err != nil && err != database.ErrRecordNotFound {
		count = 0
	} else {
		count = 1
	}

	t.logger.LogQueryResult("事务SQL单行查询", count, err)

	return err
}

// Insert 插入记录
func (t *MySQLTransaction) Insert(ctx context.Context, table string, data interface{}) (int64, error) {
	// 构建INSERT语句和参数
	query, args, err := buildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 在事务中执行插入操作
	return t.Exec(ctx, query, args)
}

// Update 更新记录
func (t *MySQLTransaction) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}) (int64, error) {
	// 构建基础UPDATE语句和参数
	query, updateArgs, err := buildUpdateQuery(table, data)
	if err != nil {
		return 0, err
	}

	// 添加WHERE子句
	if where != "" {
		query = fmt.Sprintf("%s WHERE %s", query, where)
	}

	// 合并参数：先是更新字段的值，然后是WHERE条件的参数
	allArgs := append(updateArgs, args...)

	// 在事务中执行更新操作
	return t.Exec(ctx, query, allArgs)
}

// Delete 删除记录
func (t *MySQLTransaction) Delete(ctx context.Context, table string, where string, args []any) (int64, error) {
	// 构建基础DELETE语句
	query := "DELETE FROM " + table

	// 添加WHERE子句，避免误删除所有数据
	if where != "" {
		query = query + " WHERE " + where
	}

	// 在事务中执行删除操作
	return t.Exec(ctx, query, args)
}

// Commit 提交事务
func (t *MySQLTransaction) Commit() error {
	// 记录事务提交日志
	t.logger.LogTx("提交", nil)

	// 事务结束时从活跃事务映射中移除
	t.parent.mu.Lock()
	delete(t.parent.txs, t.tx)
	t.parent.mu.Unlock()

	// 提交底层SQL事务
	err := t.tx.Commit()
	if err != nil {
		t.logger.LogTx("提交", err)
	}
	return err
}

// Rollback 回滚事务
func (t *MySQLTransaction) Rollback() error {
	// 记录事务回滚日志
	t.logger.LogTx("回滚", nil)

	// 事务结束时从活跃事务映射中移除
	t.parent.mu.Lock()
	delete(t.parent.txs, t.tx)
	t.parent.mu.Unlock()

	// 回滚底层SQL事务
	err := t.tx.Rollback()
	if err != nil {
		t.logger.LogTx("回滚", err)
	}
	return err
}

// BatchInsert 在事务中批量插入多条记录
func (t *MySQLTransaction) BatchInsert(ctx context.Context, table string, dataSlice interface{}) (int64, error) {
	// 获取反射值
	sliceVal := reflect.ValueOf(dataSlice)
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	// 空切片直接返回
	sliceLen := sliceVal.Len()
	if sliceLen == 0 {
		return 0, nil
	}

	// 获取第一个元素，用于提取字段信息
	firstElem := sliceVal.Index(0)
	if firstElem.Kind() == reflect.Ptr {
		firstElem = firstElem.Elem()
	}
	if firstElem.Kind() != reflect.Struct {
		return 0, fmt.Errorf("slice elements must be structs or pointers to structs")
	}

	// 提取字段信息
	typ := firstElem.Type()
	var columns []string
	fieldIndices := make([]int, 0)

	// 遍历结构体的所有字段，获取db标签
	for i := 0; i < firstElem.NumField(); i++ {
		field := typ.Field(i)

		// 获取db标签
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue // 跳过没有db标签或标记为忽略的字段
		}

		// 解析标签
		parts := strings.Split(tag, ",")
		column := parts[0] // 第一部分是列名

		columns = append(columns, column)
		fieldIndices = append(fieldIndices, i)
	}

	if len(columns) == 0 {
		return 0, fmt.Errorf("no valid columns found in the struct")
	}

	// 构建INSERT语句的前半部分
	placeholderGroup := "(" + strings.Repeat("?,", len(columns)-1) + "?)"
	allPlaceholders := make([]string, sliceLen)
	for i := 0; i < sliceLen; i++ {
		allPlaceholders[i] = placeholderGroup
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		table,
		strings.Join(columns, ", "),
		strings.Join(allPlaceholders, ","),
	)

	// 准备所有参数值
	values := make([]interface{}, 0, sliceLen*len(columns))
	for i := 0; i < sliceLen; i++ {
		elem := sliceVal.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}

		// 提取字段值
		for _, fieldIdx := range fieldIndices {
			fieldValue := elem.Field(fieldIdx).Interface()
			values = append(values, fieldValue)
		}
	}

	// 在事务中执行批量插入
	return t.Exec(ctx, query, values)
}

// BatchInsertWithChunk 分批批量插入，处理大量数据
//
// 参数:
// - ctx: 上下文
// - table: 表名
// - dataSlice: 包含要插入数据的结构体切片
// - chunkSize: 每批处理的记录数，推荐值为500-1000
//
// 返回:
// - 受影响的总行数
// - 错误信息
//
// 用法示例:
//
//	users := []User{...} // 大量用户数据
//	db.BatchInsertWithChunk(ctx, "users", users, 1000)
func (m *MySQL) BatchInsertWithChunk(ctx context.Context, table string, dataSlice interface{}, chunkSize int) (int64, error) {
	if chunkSize <= 0 {
		chunkSize = 1000 // 默认块大小
	}

	// 获取反射值
	sliceVal := reflect.ValueOf(dataSlice)
	if sliceVal.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}

	// 空切片直接返回
	sliceLen := sliceVal.Len()
	if sliceLen == 0 {
		return 0, nil
	}

	// 如果数据量小于chunkSize，直接使用BatchInsert
	if sliceLen <= chunkSize {
		return m.BatchInsert(ctx, table, dataSlice)
	}

	// 使用事务进行批量插入
	var totalAffected int64 = 0
	tx, err := m.BeginTx(ctx)
	if err != nil {
		return 0, err
	}

	// 确保在函数返回前处理事务
	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback() // 发生panic时回滚
			panic(r)          // 重新抛出panic
		} else if err != nil {
			_ = tx.Rollback() // 发生错误时回滚
		}
	}()

	// 分批处理
	for i := 0; i < sliceLen; i += chunkSize {
		end := i + chunkSize
		if end > sliceLen {
			end = sliceLen
		}

		// 创建当前批次的子切片
		batchSize := end - i
		batchSlice := reflect.MakeSlice(reflect.SliceOf(sliceVal.Type().Elem()), batchSize, batchSize)

		// 填充子切片
		for j := 0; j < batchSize; j++ {
			batchSlice.Index(j).Set(sliceVal.Index(i + j))
		}

		// 对当前批次执行批量插入
		affected, err := tx.(*MySQLTransaction).BatchInsert(ctx, table, batchSlice.Interface())
		if err != nil {
			// 错误处理在defer中
			return 0, err
		}
		totalAffected += affected
	}

	// 提交事务
	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return totalAffected, nil
}

// buildInsertQuery 构建插入SQL语句和参数
//
// 参数:
// - table: 要插入数据的表名
// - data: 包含要插入数据的结构体，字段通过db标签与数据库列映射
//
// 返回:
// - 生成的INSERT SQL语句
// - SQL参数值切片
// - 发生的错误，如果成功则为nil
func buildInsertQuery(table string, data interface{}) (string, []interface{}, error) {
	// 提取结构体的列和值
	columns, values, err := extractColumnsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	// 确保有字段可插入
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no fields to insert")
	}

	// 创建问号占位符
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	// 构建INSERT语句
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),      // 列名用逗号分隔
		strings.Join(placeholders, ", "), // 占位符用逗号分隔
	)

	return query, values, nil
}

// buildUpdateQuery 构建更新SQL语句和参数
//
// 参数:
// - table: 要更新数据的表名
// - data: 包含要更新数据的结构体，字段通过db标签与数据库列映射
//
// 返回:
// - 生成的UPDATE SQL语句（不包含WHERE部分）
// - SQL参数值切片
// - 发生的错误，如果成功则为nil
func buildUpdateQuery(table string, data any) (string, []any, error) {
	// 提取结构体的列和值
	columns, values, err := extractColumnsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	// 确保有字段可更新
	if len(columns) == 0 {
		return "", nil, fmt.Errorf("no fields to update")
	}

	// 构建SET部分
	setParts := make([]string, len(columns))
	for i, column := range columns {
		setParts[i] = column + " = ?" // 每个字段使用 列名=? 的形式
	}

	// 构建UPDATE语句
	query := fmt.Sprintf(
		"UPDATE %s SET %s",
		table,
		strings.Join(setParts, ", "), // SET部分用逗号分隔
	)

	return query, values, nil
}

// extractColumnsAndValues 从结构体提取数据库列名和对应的值
//
// 参数:
// - data: 需要提取信息的结构体或结构体指针
//
// 返回:
// - 从db标签提取的列名切片
// - 对应列的值切片
// - 发生的错误，如果成功则为nil
func extractColumnsAndValues(data interface{}) ([]string, []interface{}, error) {
	// 获取反射值，解引用指针
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// 确保数据是结构体类型
	if val.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("data must be a struct or pointer to struct")
	}

	// 获取结构体类型信息
	typ := val.Type()
	columns := make([]string, 0, val.NumField()) // 预分配合适的容量
	values := make([]interface{}, 0, val.NumField())

	// 遍历结构体字段
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := val.Field(i)

		// 获取db标签，用于确定数据库列名
		tag := field.Tag.Get("db")
		if tag == "" || tag == "-" {
			continue // 跳过没有db标签或标记为忽略的字段
		}

		// 如果设置了跳过零值，且字段值为零值，则跳过该字段
		if isZeroValue(fieldValue) {
			continue
		}

		// 解析标签，可能包含其他选项
		parts := strings.Split(tag, ",")
		column := parts[0] // 第一部分是列名

		// 获取字段值并保存
		columns = append(columns, column)
		values = append(values, fieldValue.Interface())
	}

	return columns, values, nil
}

// isZeroValue 检查字段是否为零值
func isZeroValue(v reflect.Value) bool {
	// 检查是否可比较，如果不可比较，使用反射的方式判断
	if !v.Type().Comparable() {
		return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
	}

	// 对于字符串类型，空字符串视为零值
	if v.Kind() == reflect.String {
		return v.String() == ""
	}

	// 直接比较零值
	zero := reflect.Zero(v.Type()).Interface()
	return reflect.DeepEqual(v.Interface(), zero)
}

// scanRows 将SQL查询结果扫描到结构体切片
//
// 参数:
// - rows: SQL查询结果集
// - dest: 目标结构体切片的指针，例如 *[]User
//
// 返回:
// - 扫描过程中发生的错误，如果成功则为nil
func scanRows(rows *sql.Rows, dest interface{}) error {
	// 确保dest是非nil指针
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer to a slice")
	}

	// 获取切片的反射值
	sliceValue := value.Elem()
	if sliceValue.Kind() != reflect.Slice {
		return fmt.Errorf("dest must be a pointer to a slice")
	}

	// 确定切片元素类型
	elemType := sliceValue.Type().Elem()
	isPtr := elemType.Kind() == reflect.Ptr // 检查切片元素是否为指针
	if isPtr {
		elemType = elemType.Elem() // 如果是指针，获取其指向的类型
	}

	// 确保元素类型是结构体
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("slice elements must be structs or pointers to structs")
	}

	// 获取查询返回的列名
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 扫描每一行数据
	for rows.Next() {
		// 创建新的结构体实例
		elemValue := reflect.New(elemType)

		// 准备扫描目标
		scanTargets, err := prepareScanTargets(elemValue.Elem(), columns)
		if err != nil {
			return err
		}

		// 扫描行数据到目标变量
		if err := rows.Scan(scanTargets...); err != nil {
			return err
		}

		// 根据切片元素类型添加到结果集
		if isPtr {
			sliceValue.Set(reflect.Append(sliceValue, elemValue)) // 添加指针
		} else {
			sliceValue.Set(reflect.Append(sliceValue, elemValue.Elem())) // 添加值
		}
	}

	// 检查行迭代过程中是否有错误
	return rows.Err()
}

// scanRow 将SQL查询单行结果扫描到结构体
//
// 参数:
// - row: SQL查询的单行结果
// - dest: 目标结构体的指针，例如 *User
//
// 返回:
// - 扫描过程中发生的错误，如果成功则为nil
// - 当未找到记录时返回 database.ErrRecordNotFound
func scanRow(row *sql.Row, dest interface{}) error {
	// 确保dest是非nil指针
	value := reflect.ValueOf(dest)
	if value.Kind() != reflect.Ptr || value.IsNil() {
		return fmt.Errorf("dest must be a non-nil pointer to a struct")
	}

	// 获取结构体的反射值
	structValue := value.Elem()
	if structValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to a struct")
	}

	// 通过反射获取字段信息，构建扫描目标切片
	fieldValues := make([]interface{}, structValue.NumField())
	for i := 0; i < structValue.NumField(); i++ {
		fieldValues[i] = structValue.Field(i).Addr().Interface() // 获取字段地址
	}

	// 扫描行数据到字段
	if err := row.Scan(fieldValues...); err != nil {
		// 处理"无记录"错误，转换为自定义错误类型
		if err == sql.ErrNoRows {
			return database.ErrRecordNotFound
		}
		return err
	}

	return nil
}

// prepareScanTargets 准备SQL结果扫描的目标变量
//
// 参数:
// - structValue: 目标结构体的反射值
// - columns: 查询结果的列名
//
// 返回:
// - 准备好的扫描目标切片，与columns一一对应
// - 准备过程中发生的错误，如果成功则为nil
func prepareScanTargets(structValue reflect.Value, columns []string) ([]interface{}, error) {
	// 为每个列创建一个扫描目标
	targets := make([]interface{}, len(columns))
	for i, column := range columns {
		// 查找与列名匹配的字段
		field, ok := findFieldByColumn(structValue, column)
		if !ok {
			// 未找到匹配字段，使用占位符接收值但不存储
			var placeholder interface{}
			targets[i] = &placeholder
			continue
		}
		// 使用找到的字段的地址作为扫描目标
		targets[i] = field.Addr().Interface()
	}
	return targets, nil
}

// findFieldByColumn 根据列名查找结构体中的匹配字段
//
// 参数:
// - structValue: 结构体的反射值
// - column: 数据库列名
//
// 返回:
// - 找到的字段的反射值
// - 是否找到匹配字段的布尔值
func findFieldByColumn(structValue reflect.Value, column string) (reflect.Value, bool) {
	// 获取结构体类型信息
	typ := structValue.Type()
	// 遍历所有字段
	for i := 0; i < structValue.NumField(); i++ {
		field := typ.Field(i)

		// 获取db标签
		tag := field.Tag.Get("db")
		if tag == "" {
			continue // 跳过没有db标签的字段
		}

		// 解析标签，检查是否匹配列名
		parts := strings.Split(tag, ",")
		if parts[0] == column {
			return structValue.Field(i), true // 返回匹配的字段
		}
	}
	return reflect.Value{}, false // 未找到匹配的字段
}
