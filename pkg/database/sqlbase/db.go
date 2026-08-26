// Package sqlbase 实现基于 database/sql 的共享 Database。
// 各驱动包只负责注册、sql.Open 驱动名和少数钩子，CRUD/事务走本包。
package sqlbase

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dblogger"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
)

// Hooks 驱动在共享实现上的差异点。横向加库时优先填钩子，而不是再抄一份 CRUD。
type Hooks struct {
	// AfterConnect 在 Ping 成功后执行，例如 SQLite PRAGMA。
	AfterConnect func(db *sql.DB, cfg *dbtypes.DbConfig) error
	// ConvertArgs 在执行前改写参数，例如 SQLite 把 time.Time 转成文本。
	ConvertArgs func(args []interface{}) []interface{}
	// FixRowsAffected 校正 RowsAffected，例如 ClickHouse INSERT 常返回 0。
	FixRowsAffected func(query string, n int64) int64
	// Interceptors SQL 执行后的观测钩子，日志之后调用，可挂指标或 trace。
	Interceptors []Interceptor
	// DefaultMaxOpen 配置未设连接池时的最大打开连接。0 表示 25。
	DefaultMaxOpen int
	// DefaultMaxIdle 配置未设连接池时的最大空闲连接。0 表示 25。
	DefaultMaxIdle int
}

// DB 共享的 database.Database 实现，按 Dialect 处理占位符和变更语句。
// 各关系库驱动通过 sqlbase.New 注册本类型，仅在 Hooks 中表达连接后配置、参数转换等差异。
//
// 核心特性:
//  1. 统一的数据库接口实现 - 符合 database.Database 接口规范
//  2. 多线程安全事务管理 - 支持多个 goroutine 并发开始和管理独立的事务
//  3. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
//  4. 智能日志记录 - 支持慢查询检测和 SQL 执行日志
//  5. 结构体映射 - 自动将 Go 结构体与数据库表映射
//  6. 上下文绑定事务 - 事务信息存储在 context 中，避免全局状态冲突
//  7. Go 底层优化 - 普通操作依赖 Go database/sql 的自动优化
//  8. 智能预编译 - 仅在必要时（如批量操作）使用手动预编译
type DB struct {
	db      *sql.DB
	config  *database.DbConfig
	logger  *dblogger.DBLogger
	dialect dialect.Dialect
	hooks   Hooks
}

var _ database.Database = (*DB)(nil)

// New 按方言创建未连接实例，供 database.Register 的工厂使用。
// 调用 Connect 之后才能执行 SQL。
func New(d dialect.Dialect, hooks Hooks) *DB {
	return &DB{dialect: d, hooks: hooks}
}

// Connect 连接到数据库。
// 建立数据库连接，配置连接池参数，并验证连接可用性。
// 会根据配置设置最大连接数、空闲连接数、连接生命周期等参数。
// DSN 为空时使用方言 FallbackDSN；Ping 或 AfterConnect 失败会关闭已打开的 *sql.DB。
// 参数:
//
//	config: 数据库配置，包含 DSN、连接池设置、日志配置等
//
// 返回:
//
//	error: 连接建立失败时返回错误信息
func (d *DB) Connect(config *database.DbConfig) error {
	d.config = config
	d.logger = dblogger.NewDBLogger(config)

	if config.DSN == "" && d.dialect.FallbackDSN() != "" {
		config.DSN = d.dialect.FallbackDSN()
	}

	// 使用背景上下文进行连接日志记录
	d.logger.LogConnecting(context.Background(), d.GetDriver(), config.DSN)

	// 打开数据库连接
	db, err := sql.Open(d.dialect.OpenDriver(), config.DSN)
	if err != nil {
		d.logger.LogError(context.Background(), "打开数据库连接", err)
		return fmt.Errorf("failed to open %s connection: %w", d.GetDriver(), err)
	}

	// 设置连接池参数
	maxOpen := 25
	if d.hooks.DefaultMaxOpen > 0 {
		maxOpen = d.hooks.DefaultMaxOpen
	}
	if config.Pool.MaxOpenConns > 0 {
		maxOpen = config.Pool.MaxOpenConns
	}
	db.SetMaxOpenConns(maxOpen)

	maxIdle := 25
	if d.hooks.DefaultMaxIdle > 0 {
		maxIdle = d.hooks.DefaultMaxIdle
	}
	if config.Pool.MaxIdleConns > 0 {
		maxIdle = config.Pool.MaxIdleConns
	}
	db.SetMaxIdleConns(maxIdle)

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
		_ = db.Close()
		d.logger.LogPing(context.Background(), err)
		return fmt.Errorf("%s connection test failed: %w", d.GetDriver(), err)
	}

	if d.hooks.AfterConnect != nil {
		if err := d.hooks.AfterConnect(db, config); err != nil {
			_ = db.Close()
			return err
		}
	}

	d.db = db
	d.logger.LogConnected(context.Background(), d.GetDriver(), map[string]any{
		"maxOpenConns":    maxOpen,
		"maxIdleConns":    maxIdle,
		"connMaxLifetime": connMaxLifetime.String(),
		"connMaxIdleTime": connMaxIdleTime.String(),
	})
	return nil
}

// Close 关闭数据库连接，释放相关资源。
// 使用上下文绑定事务时，Close 不会自动回滚事务，调用方应先 Commit 或 Rollback。
// 返回:
//
//	error: 关闭连接失败时返回错误信息
func (d *DB) Close() error {
	if d.db != nil {
		d.logger.LogDisconnect(context.Background(), d.GetDriver())
		return d.db.Close()
	}
	return nil
}

// Ping 测试数据库连接。
// 向数据库发送 ping 请求，验证连接状态。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//
// 返回:
//
//	error: 连接异常时返回错误信息
func (d *DB) Ping(ctx context.Context) error {
	err := d.db.PingContext(ctx)
	d.logger.LogPing(ctx, err)
	return err
}

// GetDriver 获取数据库驱动类型。
// 实现 Database 接口；已 Connect 时优先用配置里的 driver，否则用方言名。
// 返回:
//
//	string: 驱动类型标识，如 mysql、sqlite、oracle、clickhouse
func (d *DB) GetDriver() string {
	if d.config != nil && d.config.Driver != "" {
		return d.config.Driver
	}
	if d.dialect != nil {
		return d.dialect.Name()
	}
	return ""
}

// GetName 获取数据库连接名称。
// 实现 Database 接口，返回当前连接的名称。
// 返回:
//
//	string: 数据库连接名称，如果配置为空则返回空字符串
func (d *DB) GetName() string {
	if d.config == nil {
		return ""
	}
	return d.config.Name
}

// SetName 设置数据库连接名称。
// 用于在创建连接后设置连接名称标识，LoadAllConnections 会把 YAML 键写回实例。
// 参数:
//
//	name: 连接名称
func (d *DB) SetName(name string) {
	if d.config != nil {
		d.config.Name = name
	}
}

// DriverName 返回数据库驱动名称。
// 兼容旧驱动上的同名方法，等价于 GetDriver。
// 返回:
//
//	string: 驱动名称标识
func (d *DB) DriverName() string {
	return d.GetDriver()
}

// DSN 返回数据库连接字符串。
// 获取当前连接使用的数据源名称，返回值会隐藏敏感信息（如密码）。
// 返回:
//
//	string: 处理后的 DSN 字符串，隐藏敏感信息
func (d *DB) DSN() string {
	if d.config == nil {
		return ""
	}
	return dblogger.MaskDSN(d.config.DSN)
}

// StdDB 返回底层的 sql.DB 实例。
// 用于需要直接访问底层数据库连接的场景，例如 ClickHouse 覆盖批量方法。
// 返回:
//
//	*sql.DB: 底层的 sql.DB 实例
func (d *DB) StdDB() *sql.DB {
	return d.db
}

// Logger 返回 SQL 日志器，供覆盖批量方法时记录执行情况。
// 返回:
//
//	*dblogger.DBLogger: 当前连接的日志器
func (d *DB) Logger() *dblogger.DBLogger {
	return d.logger
}

// Dialect 返回当前数据库方言，用于拼 SQL 和错误归类。
// 返回:
//
//	dialect.Dialect: 当前驱动对应的方言
func (d *DB) Dialect() dialect.Dialect {
	return d.dialect
}

// Rewrite 把 SQL 中的 ? 改写成当前方言的占位符。
// 无方言时原样返回。业务 SQL 统一写问号，执行前调用本方法。
// 参数:
//
//	query: 业务侧 SQL
//
// 返回:
//
//	string: 驱动可执行的 SQL
func (d *DB) Rewrite(query string) string {
	if d.dialect == nil {
		return query
	}
	return d.dialect.RewriteQuery(query)
}

// convertArgs 执行前改写参数。
// 例如 SQLite 把 time.Time 转成文本。未配置钩子时原样返回。
func (d *DB) convertArgs(args []interface{}) []interface{} {
	if d.hooks.ConvertArgs != nil {
		return d.hooks.ConvertArgs(args)
	}
	return args
}

// fixRows 校正 RowsAffected。
// 例如 ClickHouse INSERT 常返回 0，由 Hooks.FixRowsAffected 补齐。
func (d *DB) fixRows(query string, n int64) int64 {
	if d.hooks.FixRowsAffected != nil {
		return d.hooks.FixRowsAffected(query, n)
	}
	return n
}
