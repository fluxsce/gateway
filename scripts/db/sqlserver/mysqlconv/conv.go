package mysqlconv

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	reCommentSQ   = regexp.MustCompile(`(?i)\s+COMMENT\s+'([^']|'')*'`)
	reCommentDQ   = regexp.MustCompile(`(?i)\s+COMMENT\s+"[^"]*"`)
	reAfter       = regexp.MustCompile(`(?i)\s+AFTER\s+[A-Za-z0-9_]+`)
	reOnUpdate    = regexp.MustCompile(`(?i)\s+ON UPDATE CURRENT_TIMESTAMP`)
	reNowInterval = regexp.MustCompile(`(?i)NOW\(\)\s*\+\s*INTERVAL\s+(\d+)\s+YEAR`)
	reNow         = regexp.MustCompile(`(?i)\bNOW\(\)`)
	reCurrentTS   = regexp.MustCompile(`(?i)\bCURRENT_TIMESTAMP\b`)
	reEngine      = regexp.MustCompile(`(?i)\s+ENGINE\s*=\s*\w+(\s+DEFAULT)?(\s+CHARSET\s*=\s*\w+)?(\s+COLLATE\s*=\s*\S+)?(\s+COMMENT\s*=\s*'[^']*')?`)
	reCreateIf    = regexp.MustCompile(`(?i)CREATE TABLE IF NOT EXISTS\s+`)
	reModifyCol   = regexp.MustCompile(`(?i)ALTER TABLE\s+([A-Za-z0-9_]+)\s+MODIFY COLUMN\s+`)
	reAddCol      = regexp.MustCompile(`(?i)ALTER TABLE\s+([A-Za-z0-9_]+)\s+ADD COLUMN\s+`)
	reUniqueKey   = regexp.MustCompile(`(?i)\bUNIQUE KEY\s+([A-Za-z0-9_]+)\s*\(`)
	reKey         = regexp.MustCompile(`(?i)\bKEY\s+([A-Za-z0-9_]+)\s*\(`)
	reLongText    = regexp.MustCompile(`(?i)\bLONGTEXT\b`)
	reMediumText  = regexp.MustCompile(`(?i)\bMEDIUMTEXT\b`)
	reTinyText    = regexp.MustCompile(`(?i)\bTINYTEXT\b`)
	reText        = regexp.MustCompile(`(?i)\bTEXT\b`)
	reJSONType    = regexp.MustCompile(`(?i)\bJSON\b`)
	reDateTime    = regexp.MustCompile(`(?i)\bDATETIME\b`)
	reDouble      = regexp.MustCompile(`(?i)\bDOUBLE(\s+PRECISION)?\b`)
	reBigUInt     = regexp.MustCompile(`(?i)\bBIGINT\(\d+\)\s+UNSIGNED`)
	reIntUDisp    = regexp.MustCompile(`(?i)\bINT\(\d+\)\s+UNSIGNED`)
	reIntU        = regexp.MustCompile(`(?i)\bINT\s+UNSIGNED`)
	reTinyInt     = regexp.MustCompile(`(?i)\bTINYINT\(\d+\)`)
	reIntDisp     = regexp.MustCompile(`(?i)\bINT\(\d+\)`)
	reBigIntDisp  = regexp.MustCompile(`(?i)\bBIGINT\(\d+\)`)
	reAutoInc     = regexp.MustCompile(`(?i)\s+AUTO_INCREMENT\b`)
	reVarchar     = regexp.MustCompile(`(?i)\bVARCHAR\s*\(\s*(\d+)\s*\)`)
	reChar        = regexp.MustCompile(`(?i)\bCHAR\s*\(\s*(\d+)\s*\)`)
	reCreateTable = regexp.MustCompile(`(?i)CREATE TABLE\s+([A-Za-z0-9_]+)\s*\(`)
	reAlterColDef = regexp.MustCompile(`(?i)(ALTER TABLE\s+[A-Za-z0-9_]+\s+ALTER COLUMN\s+[A-Za-z0-9_]+\s+N?VARCHAR\(\d+\))\s+DEFAULT NULL`)
	reAddBareCol  = regexp.MustCompile(`(?i)ALTER TABLE\s+([A-Za-z0-9_]+)\s+ADD\s+([A-Za-z0-9_]+)\b`)
	reCreateIndex = regexp.MustCompile(`(?i)CREATE INDEX\s+([A-Za-z0-9_]+)\s+ON\s+([A-Za-z0-9_]+)\s*\((.+?)\)\s*;`)
	reAddIndex    = regexp.MustCompile(`(?i)ALTER TABLE\s+([A-Za-z0-9_]+)\s+ADD INDEX\s+([A-Za-z0-9_]+)\s*\((.+?)\)\s*;`)
	reDropIndex   = regexp.MustCompile(`(?i)ALTER TABLE\s+([A-Za-z0-9_]+)\s+DROP INDEX\s+([A-Za-z0-9_]+)\s*;`)
	reInsertIgn   = regexp.MustCompile(`(?is)INSERT IGNORE INTO\s+([A-Za-z0-9_]+)\s*\((.*?)\)\s*`)
	reInsertInto  = regexp.MustCompile(`(?is)INSERT INTO\s+([A-Za-z0-9_]+)\s*\((.*?)\)\s*VALUES\s*`)
	reRoleLit     = regexp.MustCompile(`'(ROLE_(?:SUPER_ADMIN|VIEWER|TENANT_ADMIN|DEPT_ADMIN))'`)
	rePK          = regexp.MustCompile(`(?i)\bPRIMARY KEY\b`)
	reFK          = regexp.MustCompile(`(?i)\bFOREIGN KEY\b`)
	stripPrefix   = regexp.MustCompile(`([A-Za-z0-9_]+)\(\d+\)`)
)

// Convert 将一份 MySQL 脚本转为 SQL Server 方言。
func Convert(s, name string, srcSQLNames map[string]struct{}) string {
	header := "-- SQL Server 方言，由 scripts/db/mysql/" + name + " 转换。程序初始化会执行本目录除 init.sql 外的全部 .sql。\n"
	if name == "init.sql" {
		s = convertInitSQL(s, srcSQLNames)
	}
	s = applySQLServerDialect(s)
	s = injectUnicodeAlters(s)
	if name == "HUB_GW_ACCESS_LOG.sql" {
		s = strings.TrimRight(s, " \t\r\n") + "\n" + accessLogWidenSQL()
	}
	s = strings.TrimRight(s, " \t\r\n") + "\n"
	return header + s
}

func convertInitSQL(s string, srcSQLNames map[string]struct{}) string {
	s = strings.ReplaceAll(s, "source ", ":r ")
	s = strings.ReplaceAll(s, "MySQL", "SQL Server")
	s = strings.ReplaceAll(s, "mysql < init.sql", "sqlcmd -i init.sql")
	if srcSQLNames == nil {
		return s
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, ":r ") {
			file := strings.TrimSuffix(strings.TrimSpace(strings.TrimPrefix(trim, ":r ")), ";")
			if _, ok := srcSQLNames[file]; !ok {
				out = append(out, "-- 源目录无此文件，已跳过: "+file)
				continue
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

func applySQLServerDialect(s string) string {
	s = strings.ReplaceAll(s, "`", "")
	s = reCommentSQ.ReplaceAllString(s, "")
	s = reCommentDQ.ReplaceAllString(s, "")
	s = reAfter.ReplaceAllString(s, "")
	s = reOnUpdate.ReplaceAllString(s, "")
	s = reNowInterval.ReplaceAllString(s, "DATEADD(YEAR, $1, GETDATE())")
	s = mapCode(s, func(code string) string {
		code = reNow.ReplaceAllString(code, "GETDATE()")
		code = reCurrentTS.ReplaceAllString(code, "GETDATE()")
		return code
	})
	s = reEngine.ReplaceAllString(s, "")
	s = reCreateIf.ReplaceAllString(s, "CREATE TABLE ")

	s = convertSeedUpsert(s)
	s = convertInsertIgnore(s)

	s = reModifyCol.ReplaceAllString(s, "ALTER TABLE $1 ALTER COLUMN ")
	s = reAddCol.ReplaceAllString(s, "ALTER TABLE $1 ADD ")

	s = mapCode(s, func(code string) string {
		code = reUniqueKey.ReplaceAllString(code, "CONSTRAINT $1 UNIQUE (")
		code = rePK.ReplaceAllString(code, "<<<PK>>>")
		code = reFK.ReplaceAllString(code, "<<<FK>>>")
		code = reKey.ReplaceAllString(code, "INDEX $1 (")
		code = strings.ReplaceAll(code, "<<<PK>>>", "PRIMARY KEY")
		code = strings.ReplaceAll(code, "<<<FK>>>", "FOREIGN KEY")
		return code
	})

	s = mapCode(s, applyTypes)

	s = reCreateIndex.ReplaceAllStringFunc(s, func(m string) string {
		sub := reCreateIndex.FindStringSubmatch(m)
		cols := stripPrefix.ReplaceAllString(sub[3], "$1")
		return fmt.Sprintf("IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'%s' AND object_id = OBJECT_ID(N'dbo.%s'))\nCREATE INDEX %s ON %s (%s);", sub[1], sub[2], sub[1], sub[2], cols)
	})
	s = reAddIndex.ReplaceAllStringFunc(s, func(m string) string {
		sub := reAddIndex.FindStringSubmatch(m)
		cols := stripPrefix.ReplaceAllString(sub[3], "$1")
		return fmt.Sprintf("IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'%s' AND object_id = OBJECT_ID(N'dbo.%s'))\nCREATE INDEX %s ON %s (%s);", sub[2], sub[1], sub[2], sub[1], cols)
	})
	s = reDropIndex.ReplaceAllStringFunc(s, func(m string) string {
		sub := reDropIndex.FindStringSubmatch(m)
		table, idx := sub[1], sub[2]
		return fmt.Sprintf("IF EXISTS (SELECT 1 FROM sys.key_constraints WHERE name = N'%s' AND parent_object_id = OBJECT_ID(N'dbo.%s'))\nALTER TABLE %s DROP CONSTRAINT %s;\nIF EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'%s' AND object_id = OBJECT_ID(N'dbo.%s') AND is_primary_key = 0 AND is_unique_constraint = 0)\nDROP INDEX %s ON %s;", idx, table, table, idx, idx, table, idx, table)
	})

	s = reAlterColDef.ReplaceAllString(s, "$1 NULL")
	s = wrapAddColumnIfMissing(s)
	s = reCreateTable.ReplaceAllString(s, "IF OBJECT_ID(N'dbo.$1', N'U') IS NULL\nCREATE TABLE $1 (")

	// 逗号分隔的服务 ID 走 VARCHAR，避免 NVARCHAR(1000) 索引键超过 1700 字节。
	s = strings.ReplaceAll(s, "serviceDefinitionId NVARCHAR(", "serviceDefinitionId VARCHAR(")

	s = prefixNStringLiterals(s)
	s = strings.ReplaceAll(s, "可重复执行（INSERT IGNORE）", "可重复执行（IF NOT EXISTS）")
	return s
}

func applyTypes(code string) string {
	code = reLongText.ReplaceAllString(code, "NVARCHAR(MAX)")
	code = reMediumText.ReplaceAllString(code, "NVARCHAR(MAX)")
	code = reTinyText.ReplaceAllString(code, "NVARCHAR(MAX)")
	code = reText.ReplaceAllString(code, "NVARCHAR(MAX)")
	code = reJSONType.ReplaceAllString(code, "NVARCHAR(MAX)")
	code = reDateTime.ReplaceAllString(code, "DATETIME2")
	code = reDouble.ReplaceAllString(code, "FLOAT")
	code = reBigUInt.ReplaceAllString(code, "DECIMAL(20,0)")
	code = reIntUDisp.ReplaceAllString(code, "BIGINT")
	code = reIntU.ReplaceAllString(code, "BIGINT")
	code = reTinyInt.ReplaceAllString(code, "TINYINT")
	code = reIntDisp.ReplaceAllString(code, "INT")
	code = reBigIntDisp.ReplaceAllString(code, "BIGINT")
	code = reAutoInc.ReplaceAllString(code, " IDENTITY(1,1)")
	code = reVarchar.ReplaceAllString(code, "NVARCHAR($1)")
	code = reChar.ReplaceAllString(code, "NCHAR($1)")
	return code
}

func injectUnicodeAlters(s string) string {
	re := regexp.MustCompile(`(?i)CREATE TABLE\s+([A-Za-z0-9_]+)\s*\(`)
	loc := re.FindStringSubmatchIndex(s)
	if loc == nil {
		return s
	}
	table := s[loc[2]:loc[3]]
	open := loc[1] - 1
	close := matchingParen(s, open)
	if close < 0 {
		return s
	}
	end := close + 1
	for end < len(s) && (s[end] == ' ' || s[end] == '\t' || s[end] == '\r' || s[end] == '\n') {
		end++
	}
	if end < len(s) && s[end] == ';' {
		end++
	}
	body := s[open+1 : close]
	alters := unicodeAlterStatements(table, body)
	if alters == "" {
		return s
	}
	return s[:end] + "\n" + alters + s[end:]
}

func matchingParen(s string, open int) int {
	depth := 0
	inStr := false
	for i := open; i < len(s); i++ {
		c := s[i]
		if inStr {
			if c == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inStr = false
			}
			continue
		}
		switch c {
		case '\'':
			inStr = true
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func unicodeAlterStatements(table, body string) string {
	colRe := regexp.MustCompile(`(?i)^\s*([A-Za-z0-9_]+)\s+(NVARCHAR\(\d+\)|NCHAR\(\d+\))(\s+NOT NULL|\s+NULL)?`)
	var b strings.Builder
	for _, line := range strings.Split(body, "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "--") {
			continue
		}
		u := strings.ToUpper(trim)
		if strings.HasPrefix(u, "PRIMARY ") || strings.HasPrefix(u, "INDEX ") || strings.HasPrefix(u, "CONSTRAINT ") {
			continue
		}
		m := colRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		col, typ, nullability := m[1], strings.ToUpper(m[2]), strings.ToUpper(strings.TrimSpace(m[3]))
		if skipUnicodeAlterColumn(col) {
			continue
		}
		nullSQL := ""
		if nullability == "NOT NULL" {
			nullSQL = " NOT NULL"
		} else {
			nullSQL = " NULL"
		}
		b.WriteString("ALTER TABLE " + table + " ALTER COLUMN " + col + " " + typ + nullSQL + ";\n")
	}
	if b.Len() == 0 {
		return ""
	}
	return "-- 已有库：把原 VARCHAR 列升为 NVARCHAR，便于后续 N'...' 写入中文。\n" + b.String()
}

func skipUnicodeAlterColumn(col string) bool {
	c := strings.ToLower(col)
	if c == "tenantid" {
		return true
	}
	if strings.HasSuffix(c, "id") && c != "realname" {
		return true
	}
	return false
}

func wrapAddColumnIfMissing(s string) string {
	return mapCode(s, func(code string) string {
		return reAddBareCol.ReplaceAllStringFunc(code, func(m string) string {
			sub := reAddBareCol.FindStringSubmatch(m)
			if strings.EqualFold(sub[2], "INDEX") || strings.EqualFold(sub[2], "CONSTRAINT") {
				return m
			}
			return fmt.Sprintf("IF COL_LENGTH(N'dbo.%s', N'%s') IS NULL\nALTER TABLE %s ADD %s", sub[1], sub[2], sub[1], sub[2])
		})
	})
}

func accessLogWidenSQL() string {
	return `
-- 程序不执行 init.sql，多服务 ID/名称加长写在本文件，启动即可补上。
IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'idx_HUB_GW_ACCESS_LOG_time_service' AND object_id = OBJECT_ID(N'dbo.HUB_GW_ACCESS_LOG'))
DROP INDEX idx_HUB_GW_ACCESS_LOG_time_service ON HUB_GW_ACCESS_LOG;
IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'idx_HUB_GW_ACCESS_LOG_service_name' AND object_id = OBJECT_ID(N'dbo.HUB_GW_ACCESS_LOG'))
DROP INDEX idx_HUB_GW_ACCESS_LOG_service_name ON HUB_GW_ACCESS_LOG;
ALTER TABLE HUB_GW_ACCESS_LOG ALTER COLUMN serviceDefinitionId VARCHAR(1000) NULL;
ALTER TABLE HUB_GW_ACCESS_LOG ALTER COLUMN serviceName NVARCHAR(450) NULL;
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'idx_HUB_GW_ACCESS_LOG_time_service' AND object_id = OBJECT_ID(N'dbo.HUB_GW_ACCESS_LOG'))
CREATE INDEX idx_HUB_GW_ACCESS_LOG_time_service ON HUB_GW_ACCESS_LOG (gatewayStartProcessingTime, serviceDefinitionId);
IF NOT EXISTS (SELECT 1 FROM sys.indexes WHERE name = N'idx_HUB_GW_ACCESS_LOG_service_name' AND object_id = OBJECT_ID(N'dbo.HUB_GW_ACCESS_LOG'))
CREATE INDEX idx_HUB_GW_ACCESS_LOG_service_name ON HUB_GW_ACCESS_LOG (serviceName, gatewayStartProcessingTime);
`
}

func convertInsertIgnore(s string) string {
	var b strings.Builder
	rest := s
	for {
		loc := reInsertIgn.FindStringSubmatchIndex(rest)
		if loc == nil {
			b.WriteString(rest)
			break
		}
		b.WriteString(rest[:loc[0]])
		table := rest[loc[2]:loc[3]]
		colsRaw := rest[loc[4]:loc[5]]
		after := rest[loc[1]:]
		kind, body, consumed := readInsertBody(after)
		switch kind {
		case "values":
			b.WriteString(upsertFromValues(table, colsRaw, body, true))
		case "select":
			b.WriteString(insertSelectNotExists(table, colsRaw, body))
		default:
			b.WriteString(rest[loc[0]:loc[1]])
			rest = after
			continue
		}
		rest = after[consumed:]
	}
	return b.String()
}

func convertSeedUpsert(s string) string {
	seed := map[string]struct{}{
		"HUB_AUTH_RESOURCE": {},
		"HUB_AUTH_ROLE":     {},
		"HUB_USER":          {},
	}
	var b strings.Builder
	rest := s
	for {
		loc := reInsertInto.FindStringSubmatchIndex(rest)
		if loc == nil {
			b.WriteString(rest)
			break
		}
		table := rest[loc[2]:loc[3]]
		if _, ok := seed[strings.ToUpper(table)]; !ok {
			b.WriteString(rest[:loc[1]])
			rest = rest[loc[1]:]
			continue
		}
		b.WriteString(rest[:loc[0]])
		colsRaw := rest[loc[4]:loc[5]]
		after := rest[loc[1]:]
		kind, body, consumed := readInsertBody(after)
		if kind != "values" {
			b.WriteString(rest[loc[0]:loc[1]])
			rest = after
			continue
		}
		b.WriteString(upsertFromValues(table, colsRaw, body, false))
		rest = after[consumed:]
	}
	return b.String()
}

func readInsertBody(after string) (kind, body string, consumed int) {
	trim := strings.TrimLeft(after, " \t\r\n")
	skipped := len(after) - len(trim)
	if strings.HasPrefix(trim, "(") {
		end := scanUntilSemicolon(after, skipped)
		return "values", strings.TrimSpace(after[skipped:end]), endToConsumed(after, end)
	}
	u := strings.ToUpper(trim)
	if strings.HasPrefix(u, "VALUES") {
		start := skipped + len("VALUES")
		for start < len(after) && (after[start] == ' ' || after[start] == '\t' || after[start] == '\n' || after[start] == '\r') {
			start++
		}
		end := scanUntilSemicolon(after, start)
		return "values", strings.TrimSpace(after[start:end]), endToConsumed(after, end)
	}
	if strings.HasPrefix(u, "SELECT") {
		end := scanUntilSemicolon(after, skipped)
		return "select", strings.TrimSpace(after[skipped:end]), endToConsumed(after, end)
	}
	return "", "", 0
}

func endToConsumed(after string, end int) int {
	consumed := end
	if consumed < len(after) && after[consumed] == ';' {
		consumed++
	}
	return consumed
}

func scanUntilSemicolon(s string, from int) int {
	depth := 0
	inStr := false
	for i := from; i < len(s); i++ {
		c := s[i]
		if inStr {
			if c == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inStr = false
			}
			continue
		}
		if strings.HasPrefix(s[i:], "--") {
			if nl := strings.IndexByte(s[i:], '\n'); nl >= 0 {
				i += nl
				continue
			}
			return len(s)
		}
		switch c {
		case '\'':
			inStr = true
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ';':
			if depth == 0 {
				return i
			}
		}
	}
	return len(s)
}

func upsertFromValues(table, colsRaw, valuesBody string, _ bool) string {
	cols := splitCSV(stripSQLLineComments(colsRaw))
	rows := splitValueRows(stripSQLLineComments(valuesBody))
	var b strings.Builder
	for _, row := range rows {
		vals := splitCSV(strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(row), ")"), "("))
		if len(vals) != len(cols) {
			b.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;\n", table, strings.TrimSpace(colsRaw), row))
			continue
		}
		pred := pkPredicate(table, cols, vals)
		if pred == "" {
			b.WriteString(fmt.Sprintf("INSERT INTO %s (%s) VALUES %s;\n", table, strings.TrimSpace(colsRaw), row))
			continue
		}
		setParts := make([]string, 0, len(cols))
		for i, col := range cols {
			if isPKCol(table, col) || !shouldRepairCol(col) {
				continue
			}
			setParts = append(setParts, col+" = "+vals[i])
		}
		b.WriteString("IF NOT EXISTS (SELECT 1 FROM " + table + " WHERE " + pred + ")\n")
		b.WriteString("INSERT INTO " + table + " (" + strings.Join(cols, ", ") + ") VALUES (" + strings.Join(vals, ", ") + ")")
		if len(setParts) > 0 {
			b.WriteString("\nELSE\nUPDATE " + table + " SET " + strings.Join(setParts, ", ") + " WHERE " + pred)
		}
		b.WriteString(";\n")
	}
	return b.String()
}

func insertSelectNotExists(table, colsRaw, selectSQL string) string {
	fromTable := inferFromTable(selectSQL)
	role := ""
	if m := reRoleLit.FindStringSubmatch(selectSQL); m != nil {
		role = m[1]
	}
	extra := ""
	switch strings.ToUpper(table) {
	case "HUB_AUTH_ROLE_RESOURCE":
		src := fromTable
		if src == "" {
			src = "HUB_AUTH_RESOURCE"
		}
		if role == "" {
			role = "ROLE_SUPER_ADMIN"
		}
		extra = fmt.Sprintf("\n  AND NOT EXISTS (\n    SELECT 1 FROM HUB_AUTH_ROLE_RESOURCE rr\n    WHERE rr.tenantId = %s.tenantId\n      AND rr.roleId = '%s'\n      AND rr.resourceId = %s.resourceId\n  )", src, role, src)
	case "HUB_AUTH_RESOURCE":
		src := fromTable
		if src == "" {
			src = table
		}
		extra = fmt.Sprintf("\n  AND NOT EXISTS (\n    SELECT 1 FROM HUB_AUTH_RESOURCE r2\n    WHERE r2.tenantId = %s.tenantId AND r2.resourceId = %s.resourceId\n  )", src, src)
	default:
		extra = "\n  AND NOT EXISTS (SELECT 1 FROM " + table + " WHERE 1 = 0)"
	}
	return "INSERT INTO " + table + " (" + strings.TrimSpace(colsRaw) + ")\n" + strings.TrimSuffix(strings.TrimSpace(selectSQL), ";") + extra + ";\n"
}

func inferFromTable(selectSQL string) string {
	re := regexp.MustCompile(`(?i)\bFROM\s+([A-Za-z0-9_]+)`)
	m := re.FindStringSubmatch(selectSQL)
	if m == nil {
		return ""
	}
	return m[1]
}

func pkCols(table string) []string {
	switch strings.ToUpper(table) {
	case "HUB_AUTH_RESOURCE":
		return []string{"tenantId", "resourceId"}
	case "HUB_AUTH_ROLE_RESOURCE":
		return []string{"tenantId", "roleResourceId"}
	case "HUB_AUTH_ROLE":
		return []string{"tenantId", "roleId"}
	case "HUB_USER":
		return []string{"userId", "tenantId"}
	default:
		return nil
	}
}

func shouldRepairCol(col string) bool {
	c := strings.ToLower(col)
	switch {
	case strings.Contains(c, "name"), strings.Contains(c, "desc"), strings.HasSuffix(c, "text"), c == "description":
		return true
	default:
		return false
	}
}

func isPKCol(table, col string) bool {
	for _, c := range pkCols(table) {
		if strings.EqualFold(c, col) {
			return true
		}
	}
	return false
}

func pkPredicate(table string, cols, vals []string) string {
	idx := map[string]int{}
	for i, c := range cols {
		idx[c] = i
	}
	keys := pkCols(table)
	if len(keys) == 0 {
		return ""
	}
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		i, ok := idx[k]
		if !ok || i >= len(vals) {
			return ""
		}
		parts = append(parts, k+" = "+vals[i])
	}
	return strings.Join(parts, " AND ")
}

func stripSQLLineComments(s string) string {
	var b strings.Builder
	inStr := false
	for i := 0; i < len(s); i++ {
		if inStr {
			b.WriteByte(s[i])
			if s[i] == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					b.WriteByte(s[i+1])
					i++
					continue
				}
				inStr = false
			}
			continue
		}
		if strings.HasPrefix(s[i:], "--") {
			nl := strings.IndexByte(s[i:], '\n')
			if nl < 0 {
				break
			}
			b.WriteByte('\n')
			i += nl
			continue
		}
		if s[i] == '\'' {
			inStr = true
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func splitCSV(s string) []string {
	var out []string
	depth := 0
	inStr := false
	start := 0
	s = strings.TrimSpace(s)
	for i := 0; i < len(s); i++ {
		c := s[i]
		if inStr {
			if c == '\'' {
				if i+1 < len(s) && s[i+1] == '\'' {
					i++
					continue
				}
				inStr = false
			}
			continue
		}
		switch c {
		case '\'':
			inStr = true
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case ',':
			if depth == 0 {
				out = append(out, strings.TrimSpace(s[start:i]))
				start = i + 1
			}
		}
	}
	if start <= len(s) {
		tail := strings.TrimSpace(s[start:])
		if tail != "" {
			out = append(out, tail)
		}
	}
	return out
}

func splitValueRows(body string) []string {
	body = strings.TrimSpace(body)
	var rows []string
	depth := 0
	inStr := false
	start := -1
	for i := 0; i < len(body); i++ {
		c := body[i]
		if inStr {
			if c == '\'' {
				if i+1 < len(body) && body[i+1] == '\'' {
					i++
					continue
				}
				inStr = false
			}
			continue
		}
		switch c {
		case '\'':
			inStr = true
		case '(':
			if depth == 0 {
				start = i
			}
			depth++
		case ')':
			depth--
			if depth == 0 && start >= 0 {
				rows = append(rows, strings.TrimSpace(body[start:i+1]))
				start = -1
			}
		}
	}
	return rows
}

// mapCode 只变换字符串和注释之外的 SQL。
func mapCode(s string, fn func(string) string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "--") {
			end := strings.IndexByte(s[i:], '\n')
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i : i+end+1])
			i += end + 1
			continue
		}
		if strings.HasPrefix(s[i:], "/*") {
			end := strings.Index(s[i:], "*/")
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i : i+end+2])
			i += end + 2
			continue
		}
		if s[i] == '\'' {
			j := i + 1
			for j < len(s) {
				if s[j] == '\'' {
					if j+1 < len(s) && s[j+1] == '\'' {
						j += 2
						continue
					}
					j++
					break
				}
				j++
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		j := i + 1
		for j < len(s) {
			if s[j] == '\'' || strings.HasPrefix(s[j:], "--") || strings.HasPrefix(s[j:], "/*") {
				break
			}
			j++
		}
		b.WriteString(fn(s[i:j]))
		i = j
	}
	return b.String()
}

func prefixNStringLiterals(s string) string {
	var b strings.Builder
	i := 0
	for i < len(s) {
		if strings.HasPrefix(s[i:], "--") {
			end := strings.IndexByte(s[i:], '\n')
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i : i+end+1])
			i += end + 1
			continue
		}
		if strings.HasPrefix(s[i:], "/*") {
			end := strings.Index(s[i:], "*/")
			if end < 0 {
				b.WriteString(s[i:])
				break
			}
			b.WriteString(s[i : i+end+2])
			i += end + 2
			continue
		}
		if s[i] == '\'' {
			alreadyN := i > 0 && (s[i-1] == 'N' || s[i-1] == 'n')
			j := i + 1
			for j < len(s) {
				if s[j] == '\'' {
					if j+1 < len(s) && s[j+1] == '\'' {
						j += 2
						continue
					}
					j++
					break
				}
				j++
			}
			if !alreadyN {
				b.WriteByte('N')
			}
			b.WriteString(s[i:j])
			i = j
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// ListSQLNames 列出目录中的 .sql 文件名。
func ListSQLNames(dir string) (map[string]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".sql") {
			continue
		}
		out[e.Name()] = struct{}{}
	}
	return out, nil
}
