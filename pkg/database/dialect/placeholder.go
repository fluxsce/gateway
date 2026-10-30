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
// 字符串字面量、标识符引号、方括号标识符和注释中的 ? 保持原样，只给真正的占位符编号。
func RewriteQuery(query string, style PlaceholderStyle) string {
	if style == Question || !strings.Contains(query, "?") {
		return query
	}
	var buf strings.Builder
	buf.Grow(len(query) + 8)
	num := make([]byte, 0, 4)
	var idx int64
	for i := 0; i < len(query); {
		if n := literalSpan(query, i); n > 0 {
			buf.WriteString(query[i : i+n])
			i += n
			continue
		}
		if query[i] != '?' {
			buf.WriteByte(query[i])
			i++
			continue
		}
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
		i++
	}
	return buf.String()
}

// literalSpan 返回从 i 起应原样保留的字面量或注释长度。0 表示此处要继续扫描占位符。
func literalSpan(q string, i int) int {
	if i >= len(q) {
		return 0
	}
	switch q[i] {
	case '\'', '"', '`':
		return quotedSpan(q, i, q[i])
	case '[':
		return bracketIdentSpan(q, i)
	case '-':
		if i+1 < len(q) && q[i+1] == '-' {
			return lineCommentSpan(q, i)
		}
	case '/':
		if i+1 < len(q) && q[i+1] == '*' {
			return blockCommentSpan(q, i)
		}
	}
	return 0
}

// quotedSpan 扫描成对引号。连续两个引号是转义，不算结束。未闭合则保留到串尾。
func quotedSpan(q string, i int, quote byte) int {
	j := i + 1
	for j < len(q) {
		if q[j] != quote {
			j++
			continue
		}
		if j+1 < len(q) && q[j+1] == quote {
			j += 2
			continue
		}
		return j + 1 - i
	}
	return len(q) - i
}

// bracketIdentSpan 扫描 SQL Server 的 [标识符]。未闭合则保留到串尾。
func bracketIdentSpan(q string, i int) int {
	j := i + 1
	for j < len(q) {
		if q[j] == ']' {
			if j+1 < len(q) && q[j+1] == ']' {
				j += 2
				continue
			}
			return j + 1 - i
		}
		j++
	}
	return len(q) - i
}

// lineCommentSpan 扫描 -- 注释到换行，不含换行符。
func lineCommentSpan(q string, i int) int {
	j := i + 2
	for j < len(q) && q[j] != '\n' {
		j++
	}
	return j - i
}

// blockCommentSpan 扫描 /* */ 注释。未闭合则保留到串尾。
func blockCommentSpan(q string, i int) int {
	j := i + 2
	for j+1 < len(q) {
		if q[j] == '*' && q[j+1] == '/' {
			return j + 2 - i
		}
		j++
	}
	return len(q) - i
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
