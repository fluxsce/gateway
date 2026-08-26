package sqlbase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dialect"
)

// Call 一次 SQL 调用结束后的观测快照，传给 Interceptor。
type Call struct {
	// Op 操作标签，如 SQL执行、SQL插入。
	Op string
	// Query 业务侧 SQL（改写占位符之前），便于对照日志。
	Query string
	// Args 与 Query 对应的参数。
	Args []interface{}
	// Duration 从发起到驱动返回的耗时。
	Duration time.Duration
	// Extra 行数、批量大小等附加信息。
	Extra map[string]interface{}
	// Err 已经过 WrapErr 的错误，成功时为 nil。
	Err error
}

// Interceptor 在 SQL 完成并记日志之后调用，用于指标或链路。
// 不应再修改 SQL 或参数，也不应 panic。
type Interceptor func(ctx context.Context, call Call)

// WrapErr 把驱动错误映射为 database 包中的稳定错误。
// ErrRecordNotFound 与 sql.ErrNoRows 返回哨兵本身，兼容 err == ErrRecordNotFound。
// 唯一约束、断连、非法 SQL 分别包装为 ErrDuplicateKey、ErrConnection、ErrInvalidQuery。
func (d *DB) WrapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, database.ErrRecordNotFound) {
		return database.ErrRecordNotFound
	}
	if errors.Is(err, sql.ErrNoRows) {
		return database.ErrRecordNotFound
	}
	if errors.Is(err, database.ErrDuplicateKey) ||
		errors.Is(err, database.ErrConnection) ||
		errors.Is(err, database.ErrInvalidQuery) ||
		errors.Is(err, database.ErrTransaction) {
		return err
	}
	if d.dialect != nil {
		switch d.dialect.ClassifyError(err) {
		case dialect.ClassDuplicateKey:
			return fmt.Errorf("%w: %v", database.ErrDuplicateKey, err)
		case dialect.ClassConnection:
			return fmt.Errorf("%w: %v", database.ErrConnection, err)
		case dialect.ClassInvalidQuery:
			return fmt.Errorf("%w: %v", database.ErrInvalidQuery, err)
		}
	}
	return err
}

// report 归类错误、写 SQL 日志并调用 Interceptor。未找到记录在日志里当成功，返回值仍是哨兵。
func (d *DB) report(ctx context.Context, op, query string, args []interface{}, extra map[string]interface{}, duration time.Duration, err error) error {
	mapped := d.WrapErr(err)
	logErr := mapped
	if mapped != nil && errors.Is(mapped, database.ErrRecordNotFound) {
		logErr = nil
	}
	if d.logger != nil {
		d.logger.LogSQL(ctx, op, query, args, logErr, duration, extra)
	}
	for _, ic := range d.hooks.Interceptors {
		if ic != nil {
			ic(ctx, Call{
				Op:       op,
				Query:    query,
				Args:     args,
				Duration: duration,
				Extra:    extra,
				Err:      mapped,
			})
		}
	}
	return mapped
}
