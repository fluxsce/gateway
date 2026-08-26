// Package clickhouse 注册 ClickHouse 驱动，单行 CRUD 走 sqlbase，批量写入按列式存储做优化。
package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"
	"gateway/pkg/database/sqlutils"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// ClickHouse ClickHouse 数据库实现。
// 在 sqlbase 上覆盖批量写入；单行 CRUD 仍走嵌入的 sqlbase.DB。
// UPDATE/DELETE 由方言生成 ALTER TABLE 语句。
//
// 核心特性:
//  1. 统一的数据库接口实现 - 符合 database.Database 接口规范
//  2. 多线程安全事务管理 - 支持多个 goroutine 并发开始和管理独立的事务（注意：ClickHouse 事务支持有限）
//  3. 自动连接池管理 - 配置最大连接数、空闲连接和连接生命周期
//  4. 智能日志记录 - 支持慢查询检测和 SQL 执行日志
//  5. 结构体映射 - 自动将 Go 结构体与数据库表映射
//  6. 上下文绑定事务 - 事务信息存储在 context 中，避免全局状态冲突
//  7. 列式存储优化 - 针对 ClickHouse 的列式存储特性进行优化
//  8. 批量操作优化 - 针对 ClickHouse 的批量插入性能进行优化
//
// 注意：ClickHouse 的事务支持有限，主要用于批量插入的原子性保证。
type ClickHouse struct {
	*sqlbase.DB
}

var _ database.Database = (*ClickHouse)(nil)

func init() {
	database.Register(database.DriverClickHouse, func() database.Database {
		return New()
	})
}

// New 创建未连接的 ClickHouse 实例。
// 连接池默认更大（MaxOpen 50 / MaxIdle 25），并校正 INSERT 的 RowsAffected（驱动常返回 0）。
func New() *ClickHouse {
	return &ClickHouse{
		DB: sqlbase.New(dialect.MustGet(database.DriverClickHouse), sqlbase.Hooks{
			DefaultMaxOpen: 50,
			DefaultMaxIdle: 25,
			FixRowsAffected: func(query string, n int64) int64 {
				if n == 0 && strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "INSERT") {
					return 1
				}
				return n
			},
		}),
	}
}

// BatchInsert 高性能批量插入记录。
// 将切片中的多个数据结构体批量插入到 ClickHouse 中。
//
// ClickHouse 优化的高性能批量插入策略：
//  1. 列式存储优化：一次性构建所有行数据，减少列式存储的重组开销
//  2. 自适应批量处理：根据数据量智能分批，避免内存溢出
//  3. 压缩传输优化：充分利用 ClickHouse 的压缩能力减少网络传输
//  4. 事务优化：针对 ClickHouse 的事务特性进行优化
//
// 高效的批量 INSERT 策略：
//  1. 智能分批：超过 5000 条自动分批处理，避免单次传输过大
//  2. 1-2000 条使用 Prepare 预编译循环
//  3. 2001-5000 条使用单条大 INSERT，利用列式写入
//  4. 在事务中执行（autoCommit=true 时自动创建，false 时使用当前事务）
//
// 性能优势：
//   - vs 逐条插入：性能提升明显
//   - vs 预编译循环：大批量时单条大 INSERT 更优
//   - 自动分批处理，无需业务层关心批次大小
//
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 要插入的数据切片，每个元素都是结构体
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 插入失败时返回错误信息
func (c *ClickHouse) BatchInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}
	totalLen := slice.Len()
	if totalLen == 0 {
		return 0, nil
	}

	const (
		maxBatchSize          = 5000
		prepareBatchThreshold = 2000
	)

	if totalLen <= maxBatchSize {
		if totalLen <= prepareBatchThreshold {
			return c.executeSingleBatchWithPrepare(ctx, table, dataSlice, autoCommit)
		}
		return c.executeSingleBatchWithBulkInsert(ctx, table, dataSlice, autoCommit)
	}

	var totalRowsAffected int64
	start := time.Now()
	batchCount := (totalLen + maxBatchSize - 1) / maxBatchSize

	for i := 0; i < totalLen; i += maxBatchSize {
		end := i + maxBatchSize
		if end > totalLen {
			end = totalLen
		}
		batchSlice := slice.Slice(i, end)
		batchData := batchSlice.Interface()
		batchStart := time.Now()
		var rowsAffected int64
		var err error
		if batchSlice.Len() <= prepareBatchThreshold {
			rowsAffected, err = c.executeSingleBatchWithPrepare(ctx, table, batchData, autoCommit)
		} else {
			rowsAffected, err = c.executeSingleBatchWithBulkInsert(ctx, table, batchData, autoCommit)
		}
		batchDuration := time.Since(batchStart)
		if err != nil {
			c.Logger().LogSQL(ctx, "SQL大批量插入失败", "", nil, err, time.Since(start), map[string]interface{}{
				"totalRecords":     totalLen,
				"processedRecords": totalRowsAffected,
				"failedBatchIndex": i/maxBatchSize + 1,
				"totalBatches":     batchCount,
				"failedBatchSize":  batchSlice.Len(),
			})
			return totalRowsAffected, fmt.Errorf("batch insert failed at batch %d/%d (records %d-%d): %w",
				i/maxBatchSize+1, batchCount, i+1, end, err)
		}
		totalRowsAffected += rowsAffected
		currentBatch := i/maxBatchSize + 1
		c.Logger().LogSQL(ctx, "SQL批量插入进度", "", nil, nil, batchDuration, map[string]interface{}{
			"batchIndex":   currentBatch,
			"totalBatches": batchCount,
			"batchSize":    batchSlice.Len(),
			"progress":     fmt.Sprintf("%.1f%%", float64(end)/float64(totalLen)*100),
			"rowsAffected": rowsAffected,
		})
	}

	c.Logger().LogSQL(ctx, "SQL智能批量插入完成", "", nil, nil, time.Since(start), map[string]interface{}{
		"totalRowsAffected": totalRowsAffected,
		"totalRecords":      totalLen,
		"maxBatchSize":      maxBatchSize,
		"totalBatches":      batchCount,
		"executionMode":     "smart_batch_insert",
	})
	return totalRowsAffected, nil
}

// executeSingleBatchWithPrepare 使用 Prepare 预编译执行单次批量插入。
// 使用 Prepare 预编译优化的批量插入逻辑，处理单个批次的数据。
//
// Prepare 预编译优化策略：
//  1. SQL 预编译：只编译一次 INSERT 语句，重复使用
//  2. 内存优化：逐行处理，避免大量参数堆积
//  3. 网络优化：减少 SQL 字符串传输，只传输参数
//  4. 解析优化：ClickHouse 只需解析一次 SQL 语句
//  5. 批量事务：在事务中批量执行，保证原子性
//  6. 资源清理：确保 prepared statement 和事务正确清理
//
// 性能优势：
//   - vs 大 SQL 拼接：减少内存占用
//   - vs 逐条插入：prepare 开销摊薄
func (c *ClickHouse) executeSingleBatchWithPrepare(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	batchSize := slice.Len()

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建预编译 INSERT 语句
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table, strings.Join(columns, ", "), strings.Join(placeholders, ", "))

	// 第三步：准备事务执行环境
	var needCommit bool
	var tx *sql.Tx
	var stmt *sql.Stmt
	defer func() {
		if stmt != nil {
			_ = stmt.Close()
		}
		if needCommit && tx != nil {
			_ = tx.Rollback()
		}
	}()

	if autoCommit {
		tx, err = c.StdDB().BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		needCommit = true
	} else {
		var ok bool
		tx, ok = c.Tx(ctx)
		if !ok {
			return 0, fmt.Errorf("no active transaction for batch insert")
		}
	}

	stmt, err = tx.PrepareContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to prepare insert statement: %w", err)
	}

	var totalRowsAffected int64
	batchStart := time.Now()
	for i := 0; i < batchSize; i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}
		result, err := stmt.ExecContext(ctx, values...)
		if err != nil {
			return 0, fmt.Errorf("failed to execute prepared statement for item %d: %w", i, err)
		}
		if rowsAffected, err := result.RowsAffected(); err == nil {
			totalRowsAffected += rowsAffected
		} else {
			totalRowsAffected++
		}
	}
	if totalRowsAffected == 0 && batchSize > 0 {
		totalRowsAffected = int64(batchSize)
	}

	_ = stmt.Close()
	stmt = nil
	if needCommit {
		if err := tx.Commit(); err != nil {
			return 0, fmt.Errorf("failed to commit batch insert transaction: %w", err)
		}
		needCommit = false
	}

	c.Logger().LogSQL(ctx, "SQL预编译批量插入", query, nil, nil, time.Since(batchStart), map[string]interface{}{
		"batchSize":     batchSize,
		"rowsAffected":  totalRowsAffected,
		"executionMode": "prepared_statement_batch",
	})
	return totalRowsAffected, nil
}

// executeSingleBatchWithBulkInsert 使用批量 INSERT 执行单次批量插入。
// 使用大 SQL 拼接的批量插入逻辑，适合大批量数据利用 ClickHouse 列式存储优势。
//
// 批量 INSERT 优化策略：
//  1. 大 SQL 构建：一次性构建包含所有数据的 INSERT 语句
//  2. 列式存储优化：充分利用 ClickHouse 列式存储特性
//  3. 网络传输优化：一次传输大量数据，减少往返次数
//  4. 事务保证：在单个事务中完成所有插入
//  5. 内存管理：预分配切片，减少内存重分配和 GC 压力
//
// 适用场景：
//   - 大批量数据（2000+ 条）
//   - 需要充分利用 ClickHouse 列式存储优势
func (c *ClickHouse) executeSingleBatchWithBulkInsert(ctx context.Context, table string, dataSlice interface{}, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	batchSize := slice.Len()

	// 第一步：分析数据结构，提取列信息
	firstItem := slice.Index(0).Interface()
	columns, _, err := sqlutils.ExtractColumnsAndValues(firstItem)
	if err != nil {
		return 0, err
	}

	// 第二步：构建批量 INSERT 语句；预分配切片避免多次重分配
	columnsCount := len(columns)
	valuesClauses := make([]string, 0, batchSize)
	allArgs := make([]interface{}, 0, batchSize*columnsCount)
	placeholders := make([]string, columnsCount)
	for j := range placeholders {
		placeholders[j] = "?"
	}
	placeholderClause := "(" + strings.Join(placeholders, ", ") + ")"

	for i := 0; i < batchSize; i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}
		valuesClauses = append(valuesClauses, placeholderClause)
		allArgs = append(allArgs, values...)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		table, strings.Join(columns, ", "), strings.Join(valuesClauses, ", "))

	var needCommit bool
	var tx *sql.Tx
	if autoCommit {
		tx, err = c.StdDB().BeginTx(ctx, nil)
		if err != nil {
			return 0, fmt.Errorf("failed to begin transaction: %w", err)
		}
		needCommit = true
	} else {
		var ok bool
		tx, ok = c.Tx(ctx)
		if !ok {
			return 0, fmt.Errorf("no active transaction for batch insert")
		}
	}

	batchStart := time.Now()
	result, err := tx.ExecContext(ctx, query, allArgs...)
	var totalRowsAffected int64
	if err == nil {
		totalRowsAffected, _ = result.RowsAffected()
		if totalRowsAffected == 0 {
			totalRowsAffected = int64(batchSize)
		}
	}
	if err != nil {
		if needCommit {
			_ = tx.Rollback()
		}
		return 0, fmt.Errorf("failed to execute bulk insert: %w", err)
	}
	if needCommit {
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return 0, fmt.Errorf("failed to commit bulk insert transaction: %w", err)
		}
	}

	c.Logger().LogSQL(ctx, "SQL批量插入", query[:min(100, len(query))]+"...", nil, nil, time.Since(batchStart), map[string]interface{}{
		"batchSize":     batchSize,
		"rowsAffected":  totalRowsAffected,
		"executionMode": "bulk_insert_statement",
	})
	return totalRowsAffected, nil
}

// BatchUpdate 批量更新记录。
// 将切片中的多个数据结构体批量更新到 ClickHouse 中。
// 注意：ClickHouse 的 UPDATE 支持有限，这个方法主要用于兼容性。
// 建议使用 INSERT INTO ... SELECT 或 ReplacingMergeTree 引擎替代。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 要更新的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 更新失败时返回错误信息
func (c *ClickHouse) BatchUpdate(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}
	if slice.Len() == 0 {
		return 0, nil
	}
	if len(keyFields) == 0 {
		return 0, fmt.Errorf("keyFields cannot be empty")
	}

	// 注意：ClickHouse 的批量更新通常通过 INSERT INTO ... SELECT 实现，这里逐条走 ALTER UPDATE
	var totalRowsAffected int64
	for i := 0; i < slice.Len(); i++ {
		rowsAffected, err := c.Update(ctx, table, slice.Index(i).Interface(), "", nil, false, true)
		if err != nil {
			return totalRowsAffected, fmt.Errorf("failed to update item %d: %w", i, err)
		}
		totalRowsAffected += rowsAffected
	}
	return totalRowsAffected, nil
}

// BatchDelete 批量删除记录。
// 根据提供的数据切片批量删除记录，通过指定的关键字段匹配。
// 注意：ClickHouse 的 DELETE 支持有限，这个方法主要用于兼容性。
// 单主键时走 BatchDeleteByKeys；多主键字段暂不支持。
// 参数:
//
//	ctx: 上下文，用于控制请求超时和取消
//	table: 目标表名
//	dataSlice: 包含要删除记录信息的数据切片，每个元素都是结构体
//	keyFields: 用于匹配记录的关键字段列表（如主键字段）
//	autoCommit: true-自动提交, false-在当前事务中执行
//
// 返回:
//
//	int64: 受影响的行数
//	error: 删除失败时返回错误信息
func (c *ClickHouse) BatchDelete(ctx context.Context, table string, dataSlice interface{}, keyFields []string, autoCommit bool) (int64, error) {
	slice := reflect.ValueOf(dataSlice)
	if slice.Kind() != reflect.Slice {
		return 0, fmt.Errorf("dataSlice must be a slice")
	}
	if slice.Len() == 0 {
		return 0, nil
	}
	if len(keyFields) == 0 {
		return 0, fmt.Errorf("keyFields cannot be empty")
	}

	var keyValues []interface{}
	for i := 0; i < slice.Len(); i++ {
		_, values, err := sqlutils.ExtractColumnsAndValues(slice.Index(i).Interface())
		if err != nil {
			return 0, fmt.Errorf("failed to extract values from item %d: %w", i, err)
		}
		keyValues = append(keyValues, values...)
	}
	if len(keyFields) == 1 {
		return c.BatchDeleteByKeys(ctx, table, keyFields[0], keyValues, autoCommit)
	}
	return 0, fmt.Errorf("ClickHouse batch delete with multiple key fields is not fully supported")
}

// min 返回两个整数中的较小值，用于截断日志中的 SQL 片段。
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
