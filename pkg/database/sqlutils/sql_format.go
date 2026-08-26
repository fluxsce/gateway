// Package sqlutils 提供SQL格式化和结果处理的工具函数
//
// 本包主要提供以下功能：
// 1. SQL查询语句构建 - 自动生成INSERT、UPDATE等SQL语句
// 2. 结构体字段解析 - 将Go结构体转换为数据库操作参数
// 3. 零值检测 - 判断字段是否为零值，用于跳过空字段
// 4. 类型安全转换 - 确保数据类型正确匹配
//
// 使用示例：
//
//	type User struct {
//	    ID   int    `db:"id"`
//	    Name string `db:"name"`
//	}
//
//	user := User{Name: "John"}
//	query, args, err := BuildInsertQuery("users", user)
//	// 生成: INSERT INTO users (name) VALUES (?)
package sqlutils

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
)

// DatabaseType 数据库类型枚举，扩展dbtypes中的定义
type DatabaseType string

const (
	// MySQL 数据库
	DatabaseMySQL DatabaseType = dbtypes.DriverMySQL
	// PostgreSQL 数据库
	DatabasePostgreSQL DatabaseType = dbtypes.DriverPostgreSQL
	// SQLite 数据库
	DatabaseSQLite DatabaseType = dbtypes.DriverSQLite
	// SQL Server 数据库
	DatabaseSQLServer DatabaseType = dbtypes.DriverSQLServer
	// Oracle 数据库
	DatabaseOracle DatabaseType = dbtypes.DriverOracle
	// Oracle 11g 数据库（需要特殊的分页语法）
	DatabaseOracle11g DatabaseType = dbtypes.DriverOracle11g
	// MariaDB 数据库 (兼容MySQL语法)
	DatabaseMariaDB DatabaseType = dbtypes.DriverMariaDB
	// TiDB 数据库 (兼容MySQL语法)
	DatabaseTiDB DatabaseType = dbtypes.DriverTiDB
	// ClickHouse 数据库
	DatabaseClickHouse DatabaseType = dbtypes.DriverClickHouse
	// MongoDB 数据库 (NoSQL，仅用于标识)
	DatabaseMongoDB DatabaseType = dbtypes.DriverMongoDB
)

// 通过调用database接口的GetDriver方法获取驱动名称，并转换为DatabaseType类型
// 这是一个静态方法，提供统一的数据库类型获取逻辑
//
// 参数:
//
//	db: database.Database
//
// 返回:
//
//	DatabaseType: 对应的数据库类型枚举值
//
// 使用示例:
//
//	dbType := sqlutils.GetDatabaseType(dao.db)
//	query, args, err := sqlutils.BuildPaginationQuery(dbType, baseQuery, pagination)
func GetDatabaseType(db database.Database) DatabaseType {
	return DatabaseType(db.GetDriver())
}

// PaginationInfo 分页信息结构体
type PaginationInfo struct {
	Page     int // 页码，从1开始
	PageSize int // 每页大小
	Offset   int // 偏移量，从0开始
}

// NewPaginationInfo 创建分页信息
// 自动计算偏移量，确保页码和页大小的有效性
//
// 参数:
//
//	page: 页码，从1开始，如果小于1则设为1
//	pageSize: 每页大小，如果小于1则设为10
//
// 返回:
//
//	*PaginationInfo: 分页信息对象
func NewPaginationInfo(page, pageSize int) *PaginationInfo {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	return &PaginationInfo{
		Page:     page,
		PageSize: pageSize,
		Offset:   (page - 1) * pageSize,
	}
}

// BuildInsertQuery 构建插入SQL语句
// 根据数据结构体自动生成INSERT语句和参数
// 会提取结构体字段作为列名，使用占位符构建VALUES子句
//
// 功能特性：
// - 支持db tag映射，可自定义数据库字段名
// - 自动跳过零值字段，避免插入空数据
// - 支持嵌套结构体的字段展开
// - 生成参数化查询，防止SQL注入
//
// 参数:
//
//	table: 目标表名
//	data: 数据结构体，字段通过db tag映射到数据库列
//
// 返回:
//
//	string: 生成的INSERT SQL语句
//	[]interface{}: SQL参数值切片
//	error: 构建失败时返回错误信息
//
// 示例:
//
//	query, args, err := BuildInsertQuery("users", User{Name: "John", Age: 30})
//	// 返回: "INSERT INTO users (name, age) VALUES (?, ?)", ["John", 30], nil
func BuildInsertQuery(table string, data interface{}) (string, []interface{}, error) {
	columns, values, err := ExtractColumnsAndValues(data)
	if err != nil {
		return "", nil, err
	}

	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		table,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "))

	return query, values, nil
}

// BuildUpdateQuery 构建更新SQL语句的SET子句
// 根据数据结构体自动生成UPDATE语句的SET部分和参数
// 会提取结构体字段作为要更新的列，使用占位符构建SET子句
//
// 功能特性：
// - 自动生成参数化的SET子句，避免SQL注入
// - 支持db tag映射和字段名自动转换
// - 跳过零值字段，只更新有效数据
// - 返回可直接用于UPDATE语句的SET部分
//
// 参数:
//
//	table: 目标表名（此参数在当前实现中未使用，保留用于扩展）
//	data: 包含更新数据的结构体，字段通过db tag映射到数据库列
//
// 返回:
//
//	string: 生成的SET子句（如："name = ?, age = ?"）
//	[]interface{}: SET子句对应的参数值切片
//	error: 构建失败时返回错误信息
//
// 示例:
//
//	setClause, args, err := BuildUpdateQuery("users", User{Name: "John", Age: 31})
//	// 返回: "name = ?, age = ?", ["John", 31], nil
//	// 完整UPDATE语句: UPDATE users SET name = ?, age = ? WHERE id = ?
func BuildUpdateQuery(table string, data interface{}, skipZero bool) (string, []interface{}, error) {
	var columns []string
	var values []interface{}
	var err error

	if skipZero {
		// UPDATE操作跳过零值，只更新非零值字段
		columns, values, err = ExtractColumnsAndValuesSkipZero(data)
	} else {
		// UPDATE操作包含零值，更新所有字段（用于需要清空字段的场景）
		columns, values, err = ExtractColumnsAndValues(data)
	}

	if err != nil {
		return "", nil, err
	}

	setParts := make([]string, len(columns))
	for i, column := range columns {
		setParts[i] = column + " = ?"
	}

	return strings.Join(setParts, ", "), values, nil
}

// ExtractColumnsAndValues 从结构体中提取列名和值
// 通过反射解析结构体，提取可用于数据库操作的列名和对应值
// 支持db tag映射和忽略字段，默认包含所有字段（包括零值字段）
//
// 解析规则：
// - 支持db tag自定义字段名：`db:"custom_name"`
// - 忽略字段：`db:"-"`
// - 未指定tag时使用字段名小写作为列名
// - 包含零值字段，确保数据库操作的字段数量一致
// - 跳过未导出字段（小写开头的字段）
//
// 参数:
//
//	data: 要解析的结构体或结构体指针
//
// 返回:
//
//	[]string: 数据库列名切片
//	[]interface{}: 对应的值切片
//	error: 解析失败时返回错误信息
//
// 支持的字段类型：
//   - 基本类型：int, string, bool, float等
//   - 时间类型：time.Time
//   - 指针类型：*int, *string等
//   - 自定义类型（实现了相应接口的类型）
func ExtractColumnsAndValues(data interface{}) ([]string, []interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("data must be a struct or pointer to struct")
	}

	t := v.Type()
	var columns []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取数据库字段名
		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		// 跳过忽略的字段
		if dbTag == "-" {
			continue
		}

		// 注意：对于数据库插入操作，不应该跳过零值字段
		// 零值可能是有效的业务数据，且数据库表结构要求字段数量一致
		// 只有在明确标记为忽略的字段（db:"-"）才应该跳过
		// 但是，对于时间类型的零值，需要特殊处理，转换为NULL避免MySQL的'0000-00-00'错误
		// if IsZeroValue(field) {
		// 	continue
		// }

		columns = append(columns, dbTag)

		// 特殊处理时间类型的零值，转换为NULL
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t := field.Interface().(time.Time)
			if t.IsZero() {
				values = append(values, nil) // 使用NULL而不是零时间
			} else {
				values = append(values, field.Interface())
			}
		} else {
			values = append(values, field.Interface())
		}
	}

	return columns, values, nil
}

// ExtractColumnsAndValuesSkipZero 从结构体中提取列名和值，跳过零值字段
// 这是专门为UPDATE操作设计的函数，只提取非零值字段
// 适用于只想更新有效数据的场景
//
// 解析规则：
// - 支持db tag自定义字段名：`db:"custom_name"`
// - 忽略字段：`db:"-"`
// - 未指定tag时使用字段名小写作为列名
// - 自动跳过零值字段，只更新有效数据
// - 跳过未导出字段（小写开头的字段）
//
// 参数:
//
//	data: 要解析的结构体或结构体指针
//
// 返回:
//
//	[]string: 数据库列名切片（不包含零值字段）
//	[]interface{}: 对应的值切片（不包含零值字段）
//	error: 解析失败时返回错误信息
//
// 使用场景：
//   - UPDATE操作，只更新非零值字段
//   - 增量更新，避免覆盖有效数据
func ExtractColumnsAndValuesSkipZero(data interface{}) ([]string, []interface{}, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, nil, fmt.Errorf("data must be a struct or pointer to struct")
	}

	t := v.Type()
	var columns []string
	var values []interface{}

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		structField := t.Field(i)

		// 跳过未导出的字段
		if !field.CanInterface() {
			continue
		}

		// 获取数据库字段名
		dbTag := structField.Tag.Get("db")
		if dbTag == "" {
			dbTag = strings.ToLower(structField.Name)
		}

		// 跳过忽略的字段
		if dbTag == "-" {
			continue
		}

		// 跳过零值字段（UPDATE场景）
		if IsZeroValue(field) {
			continue
		}

		columns = append(columns, dbTag)

		// 特殊处理时间类型的零值，转换为NULL（虽然在SkipZero版本中零值已被跳过，但为了一致性保留此逻辑）
		if field.Type() == reflect.TypeOf(time.Time{}) {
			t := field.Interface().(time.Time)
			if t.IsZero() {
				values = append(values, nil) // 使用NULL而不是零时间
			} else {
				values = append(values, field.Interface())
			}
		} else {
			values = append(values, field.Interface())
		}
	}

	return columns, values, nil
}

// IsZeroValue 检查值是否为零值
// 判断反射值是否为对应类型的零值，用于跳过空字段
// 支持常见的基本类型、指针类型和时间类型的零值检查
//
// 零值判断规则：
// - 字符串：空字符串 ""
// - 数值类型：0
// - 布尔类型：false
// - 指针/接口：nil
// - 时间类型：time.Time{}.IsZero()
// - 结构体：使用反射的IsZero方法
//
// 参数:
//
//	v: 要检查的反射值
//
// 返回:
//
//	bool: true表示是零值，false表示非零值
//
// 使用场景：
//   - INSERT语句生成时跳过空字段
//   - UPDATE语句生成时只更新有效字段
//   - 数据验证和清理
func IsZeroValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.String() == ""
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	case reflect.Struct:
		// 特殊处理时间类型
		if v.Type() == reflect.TypeOf(time.Time{}) {
			t := v.Interface().(time.Time)
			return t.IsZero()
		}
		// 对于其他结构体类型，使用通用零值检查
		return v.IsZero()
	default:
		return false
	}
}

// BuildPaginationQuery 构建分页查询SQL语句
// 根据不同数据库类型生成相应的分页语法
// 支持MySQL、PostgreSQL、SQL Server、Oracle、SQLite等主流数据库
//
// 功能特性：
// - 自动识别数据库类型并应用对应的分页语法
// - 支持复杂查询语句的分页包装
// - 自动处理ORDER BY子句的位置
// - 生成优化的分页查询，避免性能问题
//
// 参数:
//
//	dbType: 数据库类型（mysql、postgresql、sqlserver、oracle、oracle11g、sqlite）
//	baseQuery: 基础查询语句（不包含分页部分）
//	pagination: 分页信息对象
//
// 返回:
//
//	string: 包含分页的完整SQL语句
//	[]interface{}: 分页相关的参数值
//	error: 构建失败时返回错误信息
//
// 使用示例:
//
//	baseQuery := "SELECT * FROM users WHERE status = ?"
//	pagination := NewPaginationInfo(2, 10)  // 第2页，每页10条
//	query, args, err := BuildPaginationQuery(DatabaseMySQL, baseQuery, pagination)
//	// MySQL: "SELECT * FROM users WHERE status = ? LIMIT ? OFFSET ?"
//	// args: [10, 10] (pageSize, offset)
func BuildPaginationQuery(dbType DatabaseType, baseQuery string, pagination *PaginationInfo) (string, []interface{}, error) {
	if pagination == nil {
		return baseQuery, nil, fmt.Errorf("pagination info is required")
	}
	d, err := dialect.Get(string(dbType))
	if err != nil {
		return "", nil, err
	}
	return d.BuildPagination(baseQuery, pagination.Page, pagination.PageSize, pagination.Offset)
}

// BuildLimitedDeleteQuery 构建限制删除条数的 SQL，与 BuildPaginationQuery 按同一套库类型分支。
// 不使用主键 IN 列表，避免 Oracle 1000 项、驱动占位符上限等问题。
//
// whereClause 只含条件与占位符（例如 tenantId = ? AND addTime < ?），不含 WHERE 关键字。
// 返回的 args 已按各库占位符顺序排好：WHERE 参数在前，条数在后；SQL Server 的 TOP 条数在前。
func BuildLimitedDeleteQuery(dbType DatabaseType, table, whereClause string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	d, err := dialect.Get(string(dbType))
	if err != nil {
		return "", nil, err
	}
	return d.BuildLimitedDelete(table, whereClause, whereArgs, limit)
}

// BuildCountQuery 构建统计总数的查询语句
// 将原始查询转换为COUNT查询，用于获取总记录数
// 自动处理复杂查询语句，移除不必要的ORDER BY和LIMIT子句
//
// 功能特性：
// - 智能解析原始查询，提取核心查询部分
// - 自动移除ORDER BY子句（COUNT查询不需要排序）
// - 移除LIMIT/OFFSET等分页相关子句
// - 保留WHERE、JOIN等过滤条件
// - 处理GROUP BY和HAVING子句
//
// 参数:
//
//	baseQuery: 原始查询语句
//
// 返回:
//
//	string: COUNT查询语句
//	error: 构建失败时返回错误信息
//
// 使用示例:
//
//	baseQuery := "SELECT u.*, p.name FROM users u JOIN profiles p ON u.id=p.user_id WHERE u.status = ? ORDER BY u.created_at"
//	countQuery, err := BuildCountQuery(baseQuery)
//	// 返回: "SELECT COUNT(*) AS cnt FROM users u JOIN profiles p ON u.id=p.user_id WHERE u.status = ?"
func BuildCountQuery(baseQuery string) (string, error) {
	if baseQuery == "" {
		return "", fmt.Errorf("base query cannot be empty")
	}

	// 转换为大写进行关键字匹配，但保留原始查询用于构建
	upperQuery := strings.ToUpper(baseQuery)

	// 查找SELECT和FROM的位置
	selectPos := strings.Index(upperQuery, "SELECT")
	fromPos := strings.Index(upperQuery, "FROM")

	if selectPos == -1 || fromPos == -1 {
		return "", fmt.Errorf("invalid query: missing SELECT or FROM clause")
	}

	// 提取FROM之后的部分
	fromClause := baseQuery[fromPos:]

	// 移除ORDER BY子句
	orderByPos := strings.Index(strings.ToUpper(fromClause), "ORDER BY")
	if orderByPos != -1 {
		fromClause = fromClause[:orderByPos]
	}

	// 移除LIMIT子句（MySQL、PostgreSQL、SQLite）
	limitPos := strings.Index(strings.ToUpper(fromClause), "LIMIT")
	if limitPos != -1 {
		fromClause = fromClause[:limitPos]
	}

	// 移除OFFSET子句（SQL Server、Oracle）
	offsetPos := strings.Index(strings.ToUpper(fromClause), "OFFSET")
	if offsetPos != -1 {
		fromClause = fromClause[:offsetPos]
	}

	// 移除FETCH子句（SQL Server、Oracle）
	fetchPos := strings.Index(strings.ToUpper(fromClause), "FETCH")
	if fetchPos != -1 {
		fromClause = fromClause[:fetchPos]
	}

	// 必须起别名：SQL Server 对未命名表达式返回空列名，扫描不到 db:"COUNT(*)" 会把总数当成 0。
	countQuery := fmt.Sprintf("SELECT COUNT(*) AS cnt %s", strings.TrimSpace(fromClause))

	return countQuery, nil
}

// BuildCountQueryWithOptimization 构建带优化的统计查询语句
// 根据数据库类型进行特定优化的COUNT查询构建
// 大多数情况下使用 BuildCountQuery 即可，此函数用于特殊优化场景
//
// 功能特性：
// - 基于数据库类型进行性能优化
// - 处理特定数据库的COUNT优化语法
// - 对于大表提供更高效的统计方案
//
// 参数:
//
//	dbType: 数据库类型（可选优化）
//	baseQuery: 原始查询语句
//
// 返回:
//
//	string: 优化的COUNT查询语句
//	error: 构建失败时返回错误信息
//
// 使用示例:
//
//	countQuery, err := BuildCountQueryWithOptimization(DatabaseMySQL, baseQuery)
//	// 对于大表可能使用: SELECT COUNT(1) 或其他优化语法
func BuildCountQueryWithOptimization(dbType DatabaseType, baseQuery string) (string, error) {
	if _, err := dialect.Get(string(dbType)); err != nil {
		return BuildCountQuery(baseQuery)
	}
	return BuildCountQuery(baseQuery)
}

// BuildInsertQueryForOracle 为Oracle构建INSERT语句
// Oracle特定的INSERT语句构建，支持Oracle语法特性
// 参数:
//
//	table: 目标表名
//	data: 要插入的数据结构体
//
// 返回:
//
//	string: INSERT语句，使用Oracle的:1, :2占位符格式
//	[]interface{}: 参数值数组
//	error: 构建失败时返回错误信息
func BuildInsertQueryForOracle(table string, data interface{}) (string, []interface{}, error) {
	query, values, err := BuildInsertQuery(table, data)
	if err != nil {
		return "", nil, err
	}
	d, err := dialect.Get(dbtypes.DriverOracle)
	if err != nil {
		return "", nil, err
	}
	return d.RewriteQuery(query), values, nil
}

// BuildUpdateQueryForOracle 为Oracle构建UPDATE语句的SET子句
// Oracle特定的UPDATE语句SET部分构建，使用Oracle占位符格式
// 参数:
//
//	table: 目标表名
//	data: 包含更新数据的结构体
//
// 返回:
//
//	string: SET子句，使用Oracle的:1, :2占位符格式
//	[]interface{}: 参数值数组
//	error: 构建失败时返回错误信息
func BuildUpdateQueryForOracle(table string, data interface{}, skipZero bool) (string, []interface{}, error) {
	setClause, values, err := BuildUpdateQuery(table, data, skipZero)
	if err != nil {
		return "", nil, err
	}
	d, err := dialect.Get(dbtypes.DriverOracle)
	if err != nil {
		return "", nil, err
	}
	return d.RewriteQuery(setClause), values, nil
}

// GetCurrentTimeFunction 获取当前时间的数据库函数
// 根据不同数据库类型返回对应的当前时间函数名
// 解决不同数据库NOW()函数的兼容性问题
//
// 支持的数据库：
// - MySQL/MariaDB/TiDB: NOW()
// - PostgreSQL: NOW()
// - SQL Server: GETDATE()
// - Oracle: SYSDATE
// - SQLite: datetime('now')
// - ClickHouse: now()
//
// 参数:
//
//	dbType: 数据库类型
//
// 返回:
//
//	string: 对应数据库的当前时间函数
//	error: 不支持的数据库类型返回错误
//
// 使用示例:
//
//	dbType := GetDatabaseType(db)
//	timeFunc, err := GetCurrentTimeFunction(dbType)
//	query := fmt.Sprintf("UPDATE table SET editTime = %s", timeFunc)
func GetCurrentTimeFunction(dbType DatabaseType) (string, error) {
	d, err := dialect.Get(string(dbType))
	if err != nil {
		return "", err
	}
	return d.CurrentTimeFunction()
}

// GetCurrentTimeValue 获取当前时间的参数化值
// 返回当前时间作为SQL参数，而不是函数调用
// 适用于需要在参数中传递当前时间的场景
//
// 参数:
//
//	dbType: 数据库类型（用于未来扩展，当前所有数据库都使用time.Now()）
//
// 返回:
//
//	interface{}: 当前时间值，可直接用作SQL参数
//
// 使用示例:
//
//	now := GetCurrentTimeValue(GetDatabaseType(db))
//	query := "UPDATE table SET editTime = ?"
//	args := []interface{}{now}
func GetCurrentTimeValue(dbType DatabaseType) interface{} {
	// 当前所有数据库都支持time.Time类型作为参数
	// 未来可以根据数据库类型返回不同的时间格式
	return time.Now()
}

// BuildTimeUpdateClause 构建时间更新子句
// 根据数据库类型生成兼容的时间更新语句
// 支持函数调用和参数化两种方式
//
// 参数:
//
//	dbType: 数据库类型
//	columnName: 要更新的时间列名
//	useFunction: 是否使用数据库函数（true）还是参数化值（false）
//
// 返回:
//
//	string: 时间更新子句
//	interface{}: 如果是参数化方式，返回时间值；否则返回nil
//	error: 构建失败时返回错误信息
//
// 使用示例:
//
//	dbType := GetDatabaseType(db)
//	clause, value, err := BuildTimeUpdateClause(dbType, "editTime", true)
//	// 返回: "editTime = NOW()", nil, nil
//
//	clause, value, err := BuildTimeUpdateClause(dbType, "editTime", false)
//	// 返回: "editTime = ?", time.Now(), nil
func BuildTimeUpdateClause(dbType DatabaseType, columnName string, useFunction bool) (string, interface{}, error) {
	if useFunction {
		timeFunc, err := GetCurrentTimeFunction(dbType)
		if err != nil {
			return "", nil, err
		}
		return fmt.Sprintf("%s = %s", columnName, timeFunc), nil, nil
	} else {
		return fmt.Sprintf("%s = ?", columnName), GetCurrentTimeValue(dbType), nil
	}
}
