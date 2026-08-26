// Package mysql 注册 MySQL 及兼容协议（MariaDB、TiDB）驱动，CRUD 走 sqlbase。
package mysql

import (
	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"

	_ "github.com/go-sql-driver/mysql"
)

func init() {
	// 注册 MySQL 及兼容协议驱动
	register := func(name string) {
		database.Register(name, func() database.Database {
			return sqlbase.New(dialect.MustGet(name), sqlbase.Hooks{})
		})
	}
	register(database.DriverMySQL)
	register(dbtypes.DriverMariaDB)
	register(dbtypes.DriverTiDB)
}
