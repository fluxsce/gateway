//go:build !no_oracle
// +build !no_oracle

// Package oracle 注册 Oracle / Oracle 11g 驱动，CRUD 走 sqlbase；占位符由方言改写成 :1。
package oracle

import (
	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"

	_ "github.com/godror/godror"
)

func init() {
	database.Register(database.DriverOracle, func() database.Database {
		return sqlbase.New(dialect.MustGet(database.DriverOracle), sqlbase.Hooks{})
	})
	database.Register(dbtypes.DriverOracle11g, func() database.Database {
		return sqlbase.New(dialect.MustGet(dbtypes.DriverOracle11g), sqlbase.Hooks{})
	})
}
