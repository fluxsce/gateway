// Package sqlite 注册 SQLite 驱动：连接后设 PRAGMA，执行前把 time.Time 转成文本。
package sqlite

import (
	"database/sql"
	"fmt"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"

	_ "github.com/mattn/go-sqlite3"
)

func init() {
	database.Register(database.DriverSQLite, func() database.Database {
		return sqlbase.New(dialect.MustGet(database.DriverSQLite), sqlbase.Hooks{
			AfterConnect:   configureDatabase,
			ConvertArgs:    convertTimeArgs,
			DefaultMaxOpen: 10,
			DefaultMaxIdle: 5,
		})
	})
}

// configureDatabase 配置 SQLite 数据库参数。
// 设置 WAL 模式、同步模式等优化参数：
//   - WAL 模式以支持并发读写
//   - synchronous=NORMAL 以平衡性能和安全性
//   - cache_size=-2000 表示 2MB 页面缓存
//   - busy_timeout=30000 避免 database table is locked
//   - 启用外键约束以保证数据完整性
func configureDatabase(db *sql.DB, _ *dbtypes.DbConfig) error {
	// 设置 WAL 模式以支持并发读写
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}
	// 设置同步模式为 NORMAL 以平衡性能和安全性
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		return fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	// 设置页面缓存大小（默认 -2000 表示 2MB）
	if _, err := db.Exec("PRAGMA cache_size = -2000"); err != nil {
		return fmt.Errorf("failed to set cache size: %w", err)
	}
	// 设置忙等待超时，减少 database table is locked
	if _, err := db.Exec("PRAGMA busy_timeout = 30000"); err != nil {
		return fmt.Errorf("failed to set busy timeout: %w", err)
	}
	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	return nil
}

// convertTimeArgs 将参数中的 time.Time 转换为字符串格式。
// SQLite 将日期时间存储为 TEXT 类型，需要字符串格式。
// 支持的格式：2006-01-02 15:04:05（SQLite 标准日期时间格式）。
func convertTimeArgs(args []interface{}) []interface{} {
	if args == nil {
		return nil
	}
	converted := make([]interface{}, len(args))
	for i, arg := range args {
		if t, ok := arg.(time.Time); ok {
			// 转换为 SQLite 标准日期时间格式
			converted[i] = t.Format("2006-01-02 15:04:05")
		} else {
			converted[i] = arg
		}
	}
	return converted
}
