package sqlbase

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"gateway/pkg/database/sqlutils"
)

// Exec 执行 SQL 语句。
// 执行 INSERT、UPDATE、DELETE 等不返回结果集的语句。
// 占位符统一写 ?，执行前由方言改写成各库格式。
// 使用 Go 底层自动优化，无需手动预编译。
// 支持事务和非事务模式执行。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	query: 要执行的 SQL 语句，可包含占位符
//	args: SQL 语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 执行失败时返回错误信息；唯一冲突等会包装为 ErrDuplicateKey
func (d *DB) Exec(ctx context.Context, query string, args []interface{}, autoCommit bool) (int64, error) {
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	// 直接执行，让 Go 底层自动优化
	result, err := executor.ExecContext(ctx, native, converted...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, err = result.RowsAffected()
		if err == nil {
			rowsAffected = d.fixRows(query, rowsAffected)
		}
	}

	err = d.report(ctx, "SQL执行", query, args, map[string]interface{}{
		"rowsAffected": rowsAffected,
	}, duration, err)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Query 查询多条记录。
// 执行 SELECT 语句并将结果扫描到目标切片中。
// dest 必须是切片指针，元素为结构体或结构体指针，字段通过 db tag 映射列名。
// 使用 Go 底层自动优化，无需手动预编译。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	dest: 目标切片的指针，用于接收查询结果
//	query: 要执行的 SELECT 语句，可包含占位符
//	args: SQL 语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败或扫描失败时返回错误信息
func (d *DB) Query(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	// 直接查询，让 Go 底层自动优化
	rows, err := executor.QueryContext(ctx, native, converted...)
	duration := time.Since(start)
	if err != nil {
		return d.report(ctx, "SQL查询", query, args, map[string]interface{}{"rowCount": 0}, duration, err)
	}
	defer rows.Close()

	err = sqlutils.ScanRows(rows, dest)
	rowCount := reflect.ValueOf(dest).Elem().Len()
	return d.report(ctx, "SQL查询", query, args, map[string]interface{}{
		"rowCount": rowCount,
	}, duration, err)
}

// QueryEach 按数据库游标逐行扫描到 dest 并回调，不把整结果集载入内存。
// dest 必须是结构体指针，每行复用同一块内存：回调返回后即可再次扫描覆盖，
// 调用方若要保留数据必须在回调内拷贝。回调返回 error 或 ctx 取消时停止。
// 无论成功、失败还是中途退出，都会 Close 结果集并归还连接，不会泄漏游标。
// 导出期间占用连接池中的一条连接，直到游标结束。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	dest: 每行扫描复用的结构体指针
//	query: 要执行的 SELECT 语句，可包含占位符
//	args: SQL 语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//	fn: 每扫描完一行后调用，返回 error 则中止游标
//
// 返回:
//
//	error: 查询失败、扫描失败或回调失败
func (d *DB) QueryEach(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool, fn func() error) error {
	if fn == nil {
		return fmt.Errorf("fn is required")
	}
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	rows, err := executor.QueryContext(ctx, native, converted...)
	duration := time.Since(start)
	if err != nil {
		return d.report(ctx, "SQL游标查询", query, args, map[string]interface{}{"rowCount": 0}, duration, err)
	}
	// 打开方保证关闭；ForEachRow 内再关一次是幂等的，避免中途 return 泄漏连接
	defer rows.Close()

	rowCount := 0
	err = sqlutils.ForEachRow(rows, dest, func() error {
		rowCount++
		return fn()
	})
	return d.report(ctx, "SQL游标查询", query, args, map[string]interface{}{
		"rowCount": rowCount,
	}, duration, err)
}

// QueryOne 查询单条记录。
// 执行 SELECT 语句并将结果扫描到目标结构体中。
// 如果查询不到记录，返回 ErrRecordNotFound 错误，可用 err == ErrRecordNotFound 或 IsRecordNotFound 判断。
// 使用智能字段映射，支持数据库列数与结构体字段数不匹配的情况。
// 使用 QueryContext 而不是 QueryRowContext，以便获取列信息进行智能映射。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	dest: 目标结构体的指针，用于接收查询结果
//	query: 要执行的 SELECT 语句，可包含占位符
//	args: SQL 语句中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	error: 查询失败、扫描失败或记录不存在时返回错误信息
func (d *DB) QueryOne(ctx context.Context, dest interface{}, query string, args []interface{}, autoCommit bool) error {
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	// 直接查询，让 Go 底层自动优化
	rows, err := executor.QueryContext(ctx, native, converted...)
	duration := time.Since(start)
	if err != nil {
		return d.report(ctx, "SQL单行查询", query, args, map[string]interface{}{"rowCount": 0}, duration, err)
	}

	// 使用智能扫描方式处理单行结果，支持字段数量不匹配
	err = sqlutils.ScanOneRow(rows, dest)
	rowCount := 0
	if err == nil {
		rowCount = 1
	}
	return d.report(ctx, "SQL单行查询", query, args, map[string]interface{}{
		"rowCount": rowCount,
	}, duration, err)
}
