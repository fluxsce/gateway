package sqlbase

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gateway/pkg/database/sqlutils"
)

// BatchInsert 批量插入记录。
// 将切片中的多个数据结构体批量插入到数据库中。
//
// 注意：这是保留手动预编译的方法，因为批量操作确实需要预编译优化。
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用 sql.PrepareContext() 预编译单条 INSERT 语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条插入数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 预编译循环执行流程：
//  1. 分析数据结构，提取列信息
//  2. 构建单条 INSERT 的预编译 SQL 语句
//  3. 开始事务（autoCommit=true 时自动创建，false 时使用当前事务）
//  4. 预编译单条 INSERT 语句（预编译一次，执行多次）
//  5. 循环执行：for _, item := range dataSlice { stmt.Exec(item...) }
//  6. 提交事务或在错误时回滚
//
// 优势对比：
//   - vs 大 SQL 拼接：内存使用稳定，不受批量大小影响
//   - vs 多次 Insert 调用：减少预编译开销，事务保证一致性
//   - vs Go 底层自动优化：批量操作时手动预编译性能更优
//
// 注意：
//   - BatchInsert 默认需要事务，确保批量操作的原子性
//   - 适合中小批量（≤1000 条），大批量建议业务层分批调用
//   - 任何单条记录插入失败都会回滚整个批次
//   - ClickHouse 会按数据量改用列式批量策略，见 clickhouse.BatchInsert
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	dataSlice: 要插入的数据切片，每个元素都是结构体
//	autoCommit: true-自动提交（本方法自开事务）, false-必须已有事务
//
// 返回:
//
//	int64: 受影响的行数
//	error: 插入失败时返回错误信息
func (d *DB) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice, err := asSlice(dataSlice)
	if err != nil || slice.Len() == 0 {
		return 0, err
	}

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建单条 INSERT 的预编译 SQL 语句
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))
	native := d.Rewrite(query)

	// 第三步：开始事务（BatchInsert 默认需要事务保证一致性）
	tx, needCommit, err := d.AcquireTx(ctx, autoCommit)
	if err != nil {
		return 0, err
	}

	// 第四步：预编译单条 INSERT 语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, native)
	if err != nil {
		if needCommit {
			_ = tx.Rollback()
		}
		return 0, d.WrapErr(fmt.Errorf("failed to prepare batch insert statement: %w", err))
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条插入数据
	var totalRowsAffected int64
	for i := 0; i < slice.Len(); i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}
		result, err := stmt.ExecContext(ctx, d.convertArgs(values)...)
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, d.WrapErr(fmt.Errorf("failed to insert item %d: %w", i, err))
		}
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 第六步：提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch insert transaction: %w", err)
		}
	}

	if err := d.report(ctx, "SQL批量插入", query, []interface{}{"[batch_data]"}, map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"columnsCount":  len(columns),
		"executionMode": "prepared_loop",
	}, duration, nil); err != nil {
		return 0, err
	}
	return totalRowsAffected, nil
}

// BatchUpdate 批量更新记录。
// 将切片中的多个数据结构体批量更新到数据库中。
// 使用预编译循环执行模式，根据指定的关键字段进行匹配更新。
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用 sql.PrepareContext() 预编译单条 UPDATE 语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条更新数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	dataSlice: 要更新的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交（本方法自开事务）, false-必须已有事务
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息；所有字段都是 key 时返回错误
func (d *DB) BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice, err := asSlice(dataSlice)
	if err != nil || slice.Len() == 0 {
		return 0, err
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

	// 第二步：构建 UPDATE 语句，分离 SET 子句和 WHERE 子句
	var setClauses []string
	var whereClause []string
	for _, col := range columns {
		if isKeyField(col, keyFields) {
			whereClause = append(whereClause, col+" = ?")
		} else {
			setClauses = append(setClauses, col+" = ?")
		}
	}
	if len(setClauses) == 0 {
		return 0, fmt.Errorf("no fields to update (all fields are key fields)")
	}

	query := d.dialect.UpdateSQL(table, strings.Join(setClauses, ", "), strings.Join(whereClause, " AND "))
	native := d.Rewrite(query)

	// 第三步：开始事务
	tx, needCommit, err := d.AcquireTx(ctx, autoCommit)
	if err != nil {
		if autoCommit {
			return 0, err
		}
		return 0, fmt.Errorf("no active transaction for batch update")
	}

	// 第四步：预编译 UPDATE 语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, native)
	if err != nil {
		if needCommit {
			_ = tx.Rollback()
		}
		return 0, d.WrapErr(fmt.Errorf("failed to prepare batch update statement: %w", err))
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条更新数据
	var totalRowsAffected int64
	for i := 0; i < slice.Len(); i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}

		// 重新排列参数：SET 子句参数 + WHERE 子句参数
		var args []interface{}
		for _, col := range columns {
			if !isKeyField(col, keyFields) {
				args = append(args, valueOf(columns, values, col))
			}
		}
		for _, keyField := range keyFields {
			args = append(args, valueOf(columns, values, keyField))
		}

		result, err := stmt.ExecContext(ctx, d.convertArgs(args)...)
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, d.WrapErr(fmt.Errorf("failed to update item %d: %w", i, err))
		}
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 第六步：提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch update transaction: %w", err)
		}
	}

	if err := d.report(ctx, "SQL批量更新", query, []interface{}{"[batch_data]"}, map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"keyFields":     keyFields,
		"executionMode": "prepared_loop",
	}, duration, nil); err != nil {
		return 0, err
	}
	return totalRowsAffected, nil
}

// BatchDelete 批量删除记录。
// 根据提供的数据切片批量删除记录，通过指定的关键字段匹配。
// 使用预编译循环执行模式提高性能。
//
// 高效的预编译循环执行模式：
//  1. 预编译一次：使用 sql.PrepareContext() 预编译单条 DELETE 语句
//  2. 事务保证：默认在事务中执行，确保数据一致性
//  3. 循环执行：在事务中循环执行预编译语句，逐条删除数据
//  4. 错误处理：任何错误都会触发事务回滚，保证原子性
//  5. 资源管理：自动关闭预编译语句和管理事务生命周期
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	dataSlice: 包含要删除记录信息的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交（本方法自开事务）, false-必须已有事务
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (d *DB) BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice, err := asSlice(dataSlice)
	if err != nil || slice.Len() == 0 {
		return 0, err
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

	// 第二步：构建 DELETE 语句的 WHERE 子句
	whereClause := make([]string, 0, len(keyFields))
	for _, keyField := range keyFields {
		whereClause = append(whereClause, keyField+" = ?")
	}
	query := d.dialect.DeleteSQL(table, strings.Join(whereClause, " AND "))
	native := d.Rewrite(query)

	// 第三步：开始事务
	tx, needCommit, err := d.AcquireTx(ctx, autoCommit)
	if err != nil {
		if autoCommit {
			return 0, err
		}
		return 0, fmt.Errorf("no active transaction for batch delete")
	}

	// 第四步：预编译 DELETE 语句
	start := time.Now()
	stmt, err := tx.PrepareContext(ctx, native)
	if err != nil {
		if needCommit {
			_ = tx.Rollback()
		}
		return 0, d.WrapErr(fmt.Errorf("failed to prepare batch delete statement: %w", err))
	}
	defer stmt.Close()

	// 第五步：循环执行预编译语句，逐条删除数据
	var totalRowsAffected int64
	for i := 0; i < slice.Len(); i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}
		args := make([]interface{}, 0, len(keyFields))
		for _, keyField := range keyFields {
			args = append(args, valueOf(columns, values, keyField))
		}
		result, err := stmt.ExecContext(ctx, d.convertArgs(args)...)
		if err != nil {
			if needCommit {
				_ = tx.Rollback()
			}
			return 0, d.WrapErr(fmt.Errorf("failed to delete item %d: %w", i, err))
		}
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		}
	}
	duration := time.Since(start)

	// 第六步：提交事务（如果是自动提交模式）
	if needCommit {
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to commit batch delete transaction: %w", err)
		}
	}

	if err := d.report(ctx, "SQL批量删除", query, []interface{}{"[batch_data]"}, map[string]interface{}{
		"rowsAffected":  totalRowsAffected,
		"batchSize":     slice.Len(),
		"keyFields":     keyFields,
		"executionMode": "prepared_loop",
	}, duration, nil); err != nil {
		return 0, err
	}
	return totalRowsAffected, nil
}

// BatchDeleteByKeys 根据主键列表批量删除记录。
// 更高效的批量删除方式，直接提供主键值列表。
// 使用 IN 子句进行批量删除，比逐条删除更高效。
// SQL 由方言生成，ClickHouse 会改写成 ALTER TABLE ... DELETE。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消，以及携带事务
//	table: 目标表名
//	keyField: 主键字段名
//	keys: 要删除的主键值列表
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (d *DB) BatchDeleteByKeys(ctx context.Context, table string, keyField string, keys []interface{}, autoCommit bool) (int64, error) {
	if len(keys) == 0 {
		return 0, nil
	}
	if keyField == "" {
		return 0, fmt.Errorf("keyField cannot be empty")
	}

	query := d.dialect.BatchDeleteByKeysSQL(table, keyField, len(keys))
	executor := d.getExecutor(ctx, autoCommit)
	native := d.Rewrite(query)
	converted := d.convertArgs(keys)

	start := time.Now()
	// 直接执行，使用 IN 子句批量删除
	result, err := executor.ExecContext(ctx, native, converted...)
	duration := time.Since(start)

	var rowsAffected int64
	if err == nil {
		rowsAffected, _ = result.RowsAffected()
	}
	err = d.report(ctx, "SQL批量删除(主键)", query, keys, map[string]interface{}{
		"rowsAffected":  rowsAffected,
		"batchSize":     len(keys),
		"keyField":      keyField,
		"executionMode": "in_clause",
	}, duration, err)
	if err != nil {
		return 0, err
	}
	return rowsAffected, nil
}

// asSlice 校验 dataSlice 为切片并返回其反射值。
func asSlice(dataSlice interface{}) (reflect.Value, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return reflect.Value{}, fmt.Errorf("dataSlice must be a slice")
	}
	return slice, nil
}

// isKeyField 判断列名是否属于匹配用的关键字段。
func isKeyField(col string, keyFields []string) bool {
	for _, keyField := range keyFields {
		if col == keyField {
			return true
		}
	}
	return false
}

// valueOf 按列名从 ExtractColumnsAndValues 的结果中取对应值。
func valueOf(columns []string, values []interface{}, name string) interface{} {
	for j, column := range columns {
		if column == name {
			return values[j]
		}
	}
	return nil
}
