package dialect

import (
	"fmt"
	"strings"

	"gateway/pkg/database/dbtypes"
)

// PageFunc 按页码拼分页 SQL。
type PageFunc func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error)

// LimitedDeleteFunc 拼限制条数的删除 SQL。
type LimitedDeleteFunc func(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error)

// DSNFunc 按配置生成 DSN。
type DSNFunc func(cfg *dbtypes.DbConfig) (string, error)

// Spec 用字段描述一种方言，横向加库时优先填一份 Spec 再 Register。
type Spec struct {
	name       string
	aliases    []string
	openDriver string
	fallback   string
	style      PlaceholderStyle
	timeFn     string
	paginate   PageFunc
	limitedDel LimitedDeleteFunc
	generate   DSNFunc
	validate   func(dsn string) error
	updateFmt  string
	deleteFmt  string
	inDelete   string
	classify   func(error) ErrorClass
}

// Name 返回规范驱动名。
func (s *Spec) Name() string { return s.name }

// Aliases 返回驱动别名。
func (s *Spec) Aliases() []string { return s.aliases }

// OpenDriver 返回 sql.Open 驱动名。
func (s *Spec) OpenDriver() string { return s.openDriver }

// FallbackDSN 返回备用 DSN。
func (s *Spec) FallbackDSN() string { return s.fallback }

// RewriteQuery 按占位风格改写 ?。
func (s *Spec) RewriteQuery(query string) string {
	return RewriteQuery(query, s.style)
}

// Placeholder 返回第 index 个占位符。
func (s *Spec) Placeholder(index int) string {
	return Placeholder(s.style, index)
}

// BuildPagination 拼分页。
func (s *Spec) BuildPagination(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
	if s.paginate == nil {
		return "", nil, fmt.Errorf("unsupported database type: %s", s.name)
	}
	return s.paginate(baseQuery, page, pageSize, offset)
}

// BuildLimitedDelete 拼限制条数删除。
func (s *Spec) BuildLimitedDelete(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	if !ValidIdent(table) {
		return "", nil, fmt.Errorf("invalid table name")
	}
	if strings.TrimSpace(where) == "" {
		return "", nil, fmt.Errorf("where clause is required")
	}
	if limit < 1 {
		limit = 10
	}
	if s.limitedDel == nil {
		return "", nil, fmt.Errorf("unsupported database type: %s", s.name)
	}
	return s.limitedDel(table, where, whereArgs, limit)
}

// CurrentTimeFunction 返回当前时间函数。
func (s *Spec) CurrentTimeFunction() (string, error) {
	if s.timeFn == "" {
		return "", fmt.Errorf("unsupported database type: %s", s.name)
	}
	return s.timeFn, nil
}

// ClassifyError 归类驱动错误。
func (s *Spec) ClassifyError(err error) ErrorClass {
	if s.classify != nil {
		return UnwrapClass(err, s.classify)
	}
	return ClassifyByMessage(err)
}

// GenerateDSN 生成连接串。
func (s *Spec) GenerateDSN(cfg *dbtypes.DbConfig) (string, error) {
	if s.generate == nil {
		return "", fmt.Errorf("不支持的数据库驱动类型: %s", s.name)
	}
	return s.generate(cfg)
}

// ValidateDSN 检查 DSN 外观。
func (s *Spec) ValidateDSN(dsn string) error {
	if dsn == "" {
		return fmt.Errorf("DSN不能为空")
	}
	if s.validate != nil {
		return s.validate(dsn)
	}
	return nil
}

// UpdateSQL 拼更新语句。
func (s *Spec) UpdateSQL(table, setClause, where string) string {
	fmtStr := s.updateFmt
	if fmtStr == "" {
		fmtStr = "UPDATE %s SET %s"
	}
	q := fmt.Sprintf(fmtStr, table, setClause)
	if where != "" {
		q += " WHERE " + where
	}
	return q
}

// DeleteSQL 拼删除语句。
func (s *Spec) DeleteSQL(table, where string) string {
	fmtStr := s.deleteFmt
	if fmtStr == "" {
		fmtStr = "DELETE FROM %s"
	}
	q := fmt.Sprintf(fmtStr, table)
	if where != "" {
		q += " WHERE " + where
	}
	return q
}

// BatchDeleteByKeysSQL 按主键 IN 删除。
func (s *Spec) BatchDeleteByKeysSQL(table, keyField string, keyCount int) string {
	ph := make([]string, keyCount)
	for i := range ph {
		ph[i] = "?"
	}
	fmtStr := s.inDelete
	if fmtStr == "" {
		fmtStr = "DELETE FROM %s WHERE %s IN (%s)"
	}
	return fmt.Sprintf(fmtStr, table, keyField, strings.Join(ph, ", "))
}
