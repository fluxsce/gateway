package dialect

import (
	"fmt"
	"strconv"
	"strings"
)

// PlaceholderStyle SQL 参数占位风格。业务 SQL 统一写 ?，执行前按风格改写。
type PlaceholderStyle int

const (
	// Question 问号占位，MySQL、SQLite、ClickHouse 等。
	Question PlaceholderStyle = iota
	// ColonNum Oracle :1 :2。
	ColonNum
	// DollarNum PostgreSQL $1 $2。
	DollarNum
	// AtP SQL Server @p1 @p2。
	AtP
)

// Placeholder 按风格生成第 n 个占位符，n 从 1 起。
func Placeholder(style PlaceholderStyle, n int) string {
	switch style {
	case ColonNum:
		return fmt.Sprintf(":%d", n)
	case DollarNum:
		return fmt.Sprintf("$%d", n)
	case AtP:
		return fmt.Sprintf("@p%d", n)
	default:
		return "?"
	}
}

// RewriteQuery 把 query 中的 ? 按风格改写。问号风格或没有 ? 时原样返回。
// 与历史 Oracle convertPlaceholders 相同：不解析字符串字面量。
func RewriteQuery(query string, style PlaceholderStyle) string {
	if style == Question || !strings.Contains(query, "?") {
		return query
	}
	n := strings.Count(query, "?")
	nLog10, x := 1, 10
	for n > x {
		nLog10++
		x *= 10
	}
	num := make([]byte, 0, nLog10+2)
	var buf strings.Builder
	buf.Grow(len(query) + n*(nLog10+2))
	var idx int64
	rest := query
	for i := strings.IndexByte(rest, '?'); i >= 0; i = strings.IndexByte(rest, '?') {
		buf.WriteString(rest[:i])
		rest = rest[i+1:]
		idx++
		switch style {
		case ColonNum:
			buf.WriteByte(':')
			num = strconv.AppendInt(num[:0], idx, 10)
			buf.Write(num)
		case DollarNum:
			buf.WriteByte('$')
			num = strconv.AppendInt(num[:0], idx, 10)
			buf.Write(num)
		case AtP:
			buf.WriteString("@p")
			num = strconv.AppendInt(num[:0], idx, 10)
			buf.Write(num)
		default:
			buf.WriteByte('?')
		}
	}
	buf.WriteString(rest)
	return buf.String()
}

// Placeholders 生成 n 个占位符切片，从 1 编到 n。
func Placeholders(style PlaceholderStyle, n int) []string {
	if n <= 0 {
		return nil
	}
	out := make([]string, n)
	for i := 0; i < n; i++ {
		out[i] = Placeholder(style, i+1)
	}
	return out
}
