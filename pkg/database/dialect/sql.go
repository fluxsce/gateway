package dialect

import (
	"fmt"
	"strings"
)

// ValidIdent 判断表名/列名是否只含字母数字和下划线，防止拼进 SQL 的标识被注入。
func ValidIdent(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '_' && !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func appendWhereThenLimit(whereArgs []interface{}, limit int) []interface{} {
	args := make([]interface{}, 0, len(whereArgs)+1)
	args = append(args, whereArgs...)
	return append(args, limit)
}

func hasOrderBy(query string) bool {
	return strings.Contains(strings.ToUpper(query), "ORDER BY")
}

func paginateLimitOffset(baseQuery string, pageSize, offset int) (string, []interface{}, error) {
	return fmt.Sprintf("%s LIMIT ? OFFSET ?", baseQuery), []interface{}{pageSize, offset}, nil
}

func paginateOffsetFetch(baseQuery, defaultOrder string, pageSize, offset int) (string, []interface{}, error) {
	q := baseQuery
	if !hasOrderBy(baseQuery) {
		q = fmt.Sprintf("%s %s", baseQuery, defaultOrder)
	}
	return fmt.Sprintf("%s OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", q), []interface{}{offset, pageSize}, nil
}

func paginateOracle11g(baseQuery string, pageSize, offset int) (string, []interface{}, error) {
	upperQuery := strings.ToUpper(baseQuery)
	var orderByClause string
	if !strings.Contains(upperQuery, "ORDER BY") {
		orderByClause = "ORDER BY ROWID"
	} else {
		orderByPos := strings.LastIndex(upperQuery, "ORDER BY")
		orderByClause = baseQuery[orderByPos:]
		baseQuery = baseQuery[:orderByPos]
	}
	startRow := offset + 1
	endRow := offset + pageSize
	query := fmt.Sprintf(
		"SELECT * FROM (SELECT t.*, ROW_NUMBER() OVER(%s) AS rn FROM (%s) t) WHERE rn BETWEEN ? AND ?",
		orderByClause,
		baseQuery,
	)
	return query, []interface{}{startRow, endRow}, nil
}

func limitedDeleteLimit(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	query := "DELETE FROM " + table + " WHERE " + where + " LIMIT ?"
	return query, appendWhereThenLimit(whereArgs, limit), nil
}

func limitedDeleteSQLite(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	query := "WITH doomed AS (SELECT rowid AS rid FROM " + table +
		" WHERE " + where + " LIMIT ?) DELETE FROM " + table +
		" WHERE rowid IN (SELECT rid FROM doomed)"
	return query, appendWhereThenLimit(whereArgs, limit), nil
}

func limitedDeletePostgres(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	query := "DELETE FROM " + table + " WHERE ctid IN (SELECT ctid FROM " + table +
		" WHERE " + where + " LIMIT ?)"
	return query, appendWhereThenLimit(whereArgs, limit), nil
}

func limitedDeleteOracle(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	query := "DELETE FROM " + table + " WHERE (" + where + ") AND ROWNUM <= ?"
	return query, appendWhereThenLimit(whereArgs, limit), nil
}

func limitedDeleteSQLServer(table, where string, whereArgs []interface{}, limit int) (string, []interface{}, error) {
	query := "DELETE TOP (?) FROM " + table + " WHERE " + where
	args := make([]interface{}, 0, 1+len(whereArgs))
	args = append(args, limit)
	args = append(args, whereArgs...)
	return query, args, nil
}
