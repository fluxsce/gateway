package sqlbase

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"gateway/pkg/database"
)

type txKey struct{}

// TxContext 事务上下文，包含事务和相关元数据。
// 绑定在 context 上，供日志、隔离级别还原以及批量方法复用当前事务。
type TxContext struct {
	tx      *sql.Tx
	id      string              // 事务 ID，用于日志跟踪
	created time.Time           // 事务创建时间
	options *database.TxOptions // 事务选项
}

// setTx 将事务存储到上下文中。
func setTx(ctx context.Context, txCtx *TxContext) context.Context {
	return context.WithValue(ctx, txKey{}, txCtx)
}

// getTx 从上下文中获取事务。
func getTx(ctx context.Context) (*TxContext, bool) {
	txCtx, ok := ctx.Value(txKey{}).(*TxContext)
	return txCtx, ok
}

// generateTxID 生成事务 ID，用于日志跟踪。
func generateTxID() string {
	return fmt.Sprintf("tx_%d_%d", time.Now().UnixNano(), rand.Int63())
}

// toSQLIsolation 把本包隔离级别转成 database/sql 的 IsolationLevel。
func toSQLIsolation(level database.IsolationLevel) sql.IsolationLevel {
	switch level {
	case database.IsolationReadUncommitted:
		return sql.LevelReadUncommitted
	case database.IsolationReadCommitted:
		return sql.LevelReadCommitted
	case database.IsolationRepeatableRead:
		return sql.LevelRepeatableRead
	case database.IsolationSerializable:
		return sql.LevelSerializable
	default:
		return sql.LevelDefault
	}
}

type executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// getExecutor 获取执行器（事务或连接）。
// 根据 autoCommit 参数和上下文中的事务状态返回合适的执行器。
// 如果 autoCommit 为 false 且上下文中存在活跃事务，返回事务执行器；
// 否则返回数据库连接执行器（由库自动提交）。
// 参数:
//
//	ctx: 上下文，用于获取事务信息
//	autoCommit: 是否自动提交
//
// 返回:
//
//	executor: 执行器接口，可以是 *sql.Tx 或 *sql.DB
func (d *DB) getExecutor(ctx context.Context, autoCommit bool) executor {
	if !autoCommit {
		if txCtx, ok := getTx(ctx); ok && txCtx.tx != nil {
			return txCtx.tx
		}
	}
	return d.db
}

// Tx 取出 ctx 中的 *sql.Tx。
// 无事务时 ok 为 false。供覆盖批量方法时复用当前事务。
func (d *DB) Tx(ctx context.Context) (*sql.Tx, bool) {
	txCtx, ok := getTx(ctx)
	if !ok || txCtx == nil || txCtx.tx == nil {
		return nil, false
	}
	return txCtx.tx, true
}

// AcquireTx 按 autoCommit 取得事务。
// autoCommit 为 true 时新开事务，调用方必须提交或回滚（needCommit 为 true）；
// 为 false 时必须已有事务，否则返回错误。
// 批量操作默认需要事务，确保原子性。
func (d *DB) AcquireTx(ctx context.Context, autoCommit bool) (tx *sql.Tx, needCommit bool, err error) {
	if autoCommit {
		tx, err = d.db.BeginTx(ctx, nil)
		if err != nil {
			return nil, false, fmt.Errorf("failed to begin transaction: %w", err)
		}
		return tx, true, nil
	}
	tx, ok := d.Tx(ctx)
	if !ok {
		return nil, false, fmt.Errorf("no active transaction for batch operation")
	}
	return tx, false, nil
}

// BeginTx 开始事务。
// 启动一个新的事务，支持指定隔离级别和只读属性。
// 多线程安全：每个上下文可以独立管理事务。
// 同一 context 中已有事务时返回错误。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置，可为 nil 表示库默认
//
// 返回:
//
//	context.Context: 包含事务信息的新上下文，后续 SQL 应使用它
//	error: 开始事务失败时返回错误信息，包装为 ErrTransaction
func (d *DB) BeginTx(ctx context.Context, options *database.TxOptions) (context.Context, error) {
	// 检查是否已经有事务
	if _, ok := getTx(ctx); ok {
		return ctx, fmt.Errorf("transaction already active in context")
	}

	var sqlTxOpts *sql.TxOptions
	if options != nil {
		sqlTxOpts = &sql.TxOptions{
			ReadOnly:  options.ReadOnly,
			Isolation: toSQLIsolation(options.Isolation),
		}
	}

	tx, err := d.db.BeginTx(ctx, sqlTxOpts)
	if err != nil {
		d.logger.LogTx(ctx, "开始", err)
		return ctx, fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}

	// 将事务信息绑定到上下文
	newCtx := setTx(ctx, &TxContext{
		tx:      tx,
		id:      generateTxID(),
		created: time.Now(),
		options: options,
	})
	d.logger.LogTx(newCtx, "开始", nil)
	return newCtx, nil
}

// Commit 提交事务。
// 提交上下文中的事务，使所有未提交的更改生效。
// 参数:
//
//	ctx: 由 BeginTx 或 InTx 返回的、含事务的上下文
//
// 返回:
//
//	error: 无活跃事务或提交失败
func (d *DB) Commit(ctx context.Context) error {
	txCtx, ok := getTx(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}
	err := txCtx.tx.Commit()
	txCtx.tx = nil
	d.logger.LogTx(ctx, "提交", err)
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// Rollback 回滚事务。
// 回滚上下文中的事务，撤销所有未提交的更改。
// 参数:
//
//	ctx: 由 BeginTx 或 InTx 返回的、含事务的上下文
//
// 返回:
//
//	error: 无活跃事务或回滚失败
func (d *DB) Rollback(ctx context.Context) error {
	txCtx, ok := getTx(ctx)
	if !ok || txCtx.tx == nil {
		return fmt.Errorf("no active transaction in context")
	}
	err := txCtx.tx.Rollback()
	txCtx.tx = nil
	d.logger.LogTx(ctx, "回滚", err)
	if err != nil {
		return fmt.Errorf("%w: %v", database.ErrTransaction, err)
	}
	return nil
}

// InTx 在事务中执行函数。
// 自动管理事务的生命周期。
// 如果函数正常返回，自动提交事务。
// 如果函数返回错误或发生 panic，自动回滚事务并将 panic 转换为错误，避免进程崩溃。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	options: 事务选项，包含隔离级别和只读设置，可为 nil
//	fn: 在事务中执行的函数，接收包含事务的上下文，返回 error 表示是否成功
//
// 返回:
//
//	error: 事务执行失败时返回错误信息，包括 panic 转换的错误
func (d *DB) InTx(ctx context.Context, options *database.TxOptions, fn func(context.Context) error) (err error) {
	txCtx, err := d.BeginTx(ctx, options)
	if err != nil {
		return err
	}
	defer func() {
		if r := recover(); r != nil {
			_ = d.Rollback(txCtx)
			// 将 panic 转换为错误，避免程序崩溃
			err = fmt.Errorf("transaction panic recovered: %v", r)
		}
	}()
	if err := fn(txCtx); err != nil {
		_ = d.Rollback(txCtx)
		return err
	}
	return d.Commit(txCtx)
}
