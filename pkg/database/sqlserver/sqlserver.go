// Package sqlserver 注册 SQL Server 驱动，CRUD 走 sqlbase；占位符由方言改写成 @p1。
// ntlm 始终注册；Windows 另见 sqlserver_windows.go 的 winsspi，供 sqlserver_authenticator 使用。
package sqlserver

import (
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"

	_ "github.com/microsoft/go-mssqldb"
	_ "github.com/microsoft/go-mssqldb/integratedauth/ntlm"
)

func init() {
	register := func(name string) {
		database.Register(name, func() database.Database {
			return sqlbase.New(dialect.MustGet(name), sqlbase.Hooks{
				ConvertArgs: convertTimeArgs,
			})
		})
	}
	register(database.DriverSQLServer)
	register("mssql")
}

// convertTimeArgs 把 time.Time 写成无时区 ISO 墙钟，按 DATETIME2 比较。
// 驱动默认把 time.Time 编成 datetimeoffset；和 DATETIME2 比较时会把列当成 UTC，
// 东八区「最近一小时」对不上库里已有的本地时间（例如 23:50 落在 15:00–16:00 窗外）。
func convertTimeArgs(args []interface{}) []interface{} {
	if args == nil {
		return nil
	}
	converted := make([]interface{}, len(args))
	for i, arg := range args {
		switch t := arg.(type) {
		case time.Time:
			converted[i] = formatDateTime2(t)
		case *time.Time:
			if t == nil {
				converted[i] = nil
			} else {
				converted[i] = formatDateTime2(*t)
			}
		default:
			converted[i] = arg
		}
	}
	return converted
}

func formatDateTime2(t time.Time) string {
	if t.IsZero() {
		return "0001-01-01T00:00:00"
	}
	return t.Format("2006-01-02T15:04:05.9999999")
}
