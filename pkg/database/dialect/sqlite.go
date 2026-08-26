package dialect

import (
	"fmt"
	"strings"

	appconfig "gateway/pkg/config"
	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"
)

func init() {
	Register(&Spec{
		name:       dbtypes.DriverSQLite,
		openDriver: "sqlite3",
		fallback:   ":memory:",
		style:      Question,
		timeFn:     "datetime('now')",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateLimitOffset(baseQuery, pageSize, offset)
		},
		limitedDel: limitedDeleteSQLite,
		generate:   generateSQLite,
		validate: func(dsn string) error {
			if dsn != ":memory:" && !strings.Contains(dsn, ".") && !strings.HasPrefix(dsn, "file:") {
				return huberrors.NewError("SQLite DSN格式可能不正确")
			}
			return nil
		},
	})
}

func generateSQLite(config *dbtypes.DbConfig) (string, error) {
	if err := validateSQLiteConfigSafety(config); err != nil {
		return "", err
	}

	dbPath := config.Connection.FilePath
	if dbPath == "" {
		dbPath = config.Connection.Database
	}

	if dbPath == "" || dbPath == ":memory:" {
		return ":memory:", nil
	}

	if !strings.Contains(dbPath, ".") && !strings.Contains(dbPath, "/") && !strings.Contains(dbPath, "\\") {
		dbPath = fmt.Sprintf("./%s.db", dbPath)
	}

	dbPath = appconfig.ResolvePath(dbPath)

	params := make([]string, 0)

	cacheMode := config.Connection.CacheMode
	if cacheMode == "" {
		cacheMode = "private"
	}
	params = append(params, "cache="+cacheMode)

	connectionMode := config.Connection.ConnectionMode
	if connectionMode == "" {
		connectionMode = "rwc"
	}
	params = append(params, "mode="+connectionMode)

	journalMode := config.Connection.JournalMode
	if journalMode == "" {
		journalMode = "WAL"
	}
	params = append(params, "_journal_mode="+journalMode)

	synchronousMode := config.Connection.SynchronousMode
	if synchronousMode == "" {
		synchronousMode = "FULL"
	}
	if synchronousMode == "OFF" {
		synchronousMode = "NORMAL"
	}
	params = append(params, "_synchronous="+synchronousMode)

	if config.Connection.ForeignKeys {
		params = append(params, "_foreign_keys=1")
	} else {
		params = append(params, "_foreign_keys=0")
	}

	busyTimeout := config.Connection.BusyTimeout
	if busyTimeout == 0 {
		busyTimeout = 5000
	}
	params = append(params, fmt.Sprintf("_busy_timeout=%d", busyTimeout))

	if config.Connection.CacheSize != 0 {
		params = append(params, fmt.Sprintf("_cache_size=%d", config.Connection.CacheSize))
	}
	if config.Connection.AutoVacuum != "" {
		params = append(params, "_auto_vacuum="+config.Connection.AutoVacuum)
	}
	if config.Connection.TempStore != "" {
		params = append(params, "_temp_store="+config.Connection.TempStore)
	}
	if config.Connection.PageSize > 0 {
		params = append(params, fmt.Sprintf("_page_size=%d", config.Connection.PageSize))
	}
	if config.Connection.MaxPageCount > 0 {
		params = append(params, fmt.Sprintf("_max_page_count=%d", config.Connection.MaxPageCount))
	}
	if config.Connection.LockingMode != "" {
		params = append(params, "_locking_mode="+config.Connection.LockingMode)
	}

	secureDelete := config.Connection.SecureDelete
	if secureDelete == "" {
		secureDelete = "true"
	}
	params = append(params, "_secure_delete="+secureDelete)

	walAutocheckpoint := config.Connection.WALAutocheckpoint
	if walAutocheckpoint == 0 && journalMode == "WAL" {
		walAutocheckpoint = 1000
	}
	if walAutocheckpoint > 0 {
		params = append(params, fmt.Sprintf("_wal_autocheckpoint=%d", walAutocheckpoint))
	}
	if config.Connection.QueryOnly {
		params = append(params, "_query_only=1")
	}

	if len(params) > 0 {
		return fmt.Sprintf("file:%s?%s", dbPath, strings.Join(params, "&")), nil
	}
	return dbPath, nil
}

func validateSQLiteConfigSafety(config *dbtypes.DbConfig) error {
	if config == nil || config.Connection == (dbtypes.ConnectionConfig{}) {
		return nil
	}

	conn := &config.Connection

	if conn.SynchronousMode == "OFF" && conn.JournalMode != "WAL" {
		return huberrors.NewError("数据安全风险：synchronous=OFF + journal_mode=%s 可能导致数据丢失，建议使用WAL模式", conn.JournalMode)
	}
	if conn.JournalMode == "OFF" {
		return huberrors.NewError("数据安全风险：journal_mode=OFF 会禁用事务日志，可能导致数据库损坏")
	}
	if conn.JournalMode == "MEMORY" && conn.SynchronousMode == "OFF" {
		return huberrors.NewError("数据安全风险：journal_mode=MEMORY + synchronous=OFF 组合在系统崩溃时会丢失所有数据")
	}
	if conn.ConnectionMode == "memory" && conn.JournalMode != "" && conn.JournalMode != "MEMORY" {
		return huberrors.NewError("配置冲突：内存数据库(mode=memory)应该使用journal_mode=MEMORY")
	}

	if conn.PageSize > 0 {
		validPageSizes := []int{512, 1024, 2048, 4096, 8192, 16384, 32768, 65536}
		isValid := false
		for _, size := range validPageSizes {
			if conn.PageSize == size {
				isValid = true
				break
			}
		}
		if !isValid {
			return huberrors.NewError("无效的页面大小：%d，有效值为：512, 1024, 2048, 4096, 8192, 16384, 32768, 65536", conn.PageSize)
		}
	}

	if conn.CacheSize < -2000000 {
		return huberrors.NewError("缓存大小过大：%d KB，可能导致内存不足", -conn.CacheSize)
	}
	if conn.BusyTimeout < 0 {
		return huberrors.NewError("无效的忙等待超时：%d，必须为非负数", conn.BusyTimeout)
	}
	if conn.JournalMode == "WAL" && conn.LockingMode == "EXCLUSIVE" {
		return huberrors.NewError("配置冲突：WAL模式不支持EXCLUSIVE锁定模式")
	}
	return nil
}
