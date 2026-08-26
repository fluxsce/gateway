package sqlbase

import (
	"context"
	"time"

	"gateway/pkg/database/sqlutils"
)

// Insert 插入记录。
// 根据提供的数据结构体自动构建 INSERT 语句并执行。
// 使用 Go 底层自动优化，无需手动预编译。
// 会自动提取结构体字段作为列名和值，支持 db tag 映射。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	data: 要插入的数据结构体，字段通过 db tag 映射到数据库列
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 插入记录的自增 ID（如果有；库不支持时为 0）
//	error: 插入失败时返回错误信息；唯一冲突会包装为 ErrDuplicateKey
func (d *DB) Insert(ctx context.Context, table string, data interface{}, autoCommit bool) (int64, error) {
	query, args, err := sqlutils.BuildInsertQuery(table, data)
	if err != nil {
		return 0, err
	}

	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	// 直接执行，让 Go 底层自动优化
	result, err := executor.ExecContext(ctx, native, converted...)
	duration := time.Since(start)

	var lastInsertId int64
	var rowsAffected int64
	if err == nil {
		lastInsertId, _ = result.LastInsertId()
		rowsAffected, _ = result.RowsAffected()
		rowsAffected = d.fixRows(query, rowsAffected)
	}

	err = d.report(ctx, "SQL插入", query, args, map[string]interface{}{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}, duration, err)
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

// Update 更新记录。
// 根据提供的数据结构体和 WHERE 条件构建 UPDATE 语句并执行。
// 会自动提取结构体字段作为要更新的列和值。具体语法由方言生成，例如 ClickHouse 使用 ALTER TABLE ... UPDATE。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	data: 包含更新数据的结构体，字段通过 db tag 映射到数据库列
//	where: WHERE 条件语句，可包含占位符；为空则不加 WHERE
//	args: WHERE 条件中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//	skipZero: true-跳过零值字段；false-包含零值（用于清空字段）
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息
func (d *DB) Update(ctx context.Context, table string, data interface{}, where string, args []interface{}, autoCommit bool, skipZero bool) (int64, error) {
	setClause, setArgs, err := sqlutils.BuildUpdateQuery(table, data, skipZero)
	if err != nil {
		return 0, err
	}
	if where != "" {
		setArgs = append(setArgs, args...)
	}
	query := d.dialect.UpdateSQL(table, setClause, where)

	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(setArgs)

	start := time.Now()
	// 直接执行，让 Go 底层自动优化
	result, err := executor.ExecContext(ctx, native, converted...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	err = d.report(ctx, "SQL更新", query, setArgs, map[string]interface{}{
		"rowsAffected": rowsAffected,
	}, duration, err)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// Delete 删除记录。
// 根据 WHERE 条件构建 DELETE 语句并执行。
// 使用 Go 底层自动优化，无需手动预编译。
// 具体语法由方言生成，例如 ClickHouse 使用 ALTER TABLE ... DELETE。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	where: WHERE 条件语句，可包含占位符；为空则不加 WHERE
//	args: WHERE 条件中占位符对应的参数值
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (d *DB) Delete(ctx context.Context, table string, where string, args []interface{}, autoCommit bool) (int64, error) {
	query := d.dialect.DeleteSQL(table, where)
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(args)

	start := time.Now()
	// 直接执行，让 Go 底层自动优化
	result, err := executor.ExecContext(ctx, native, converted...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}

	err = d.report(ctx, "SQL删除", query, args, map[string]interface{}{
		"rowsAffected": rowsAffected,
	}, duration, err)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}
