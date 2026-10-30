// Package vastbase 注册海量 Vastbase 驱动。
//
// 连接器是 openGauss 的 Go 驱动（gitcode.com/opengauss/openGauss-connector-go-pq）。
// Vastbase、openGauss、MogDB 共用这套协议，能处理默认的 sha256 口令。
// database/sql 的驱动名是 opengauss。连上后读取 sql_compatibility：
// PG 用 PostgreSQL 方言和 scripts/db/vastbase；A 用 Oracle 兼容方言和 scripts/db/vastbase_oracle。
package vastbase

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
	"gateway/pkg/database/sqlbase"
	"gateway/pkg/logger"

	_ "gitcode.com/opengauss/openGauss-connector-go-pq"
)

const compatProbeTimeout = 5 * time.Second

func init() {
	database.Register(dbtypes.DriverVastbase, func() database.Database {
		var inst *sqlbase.DB
		inst = sqlbase.New(dialect.MustGet(dbtypes.DriverVastbase), sqlbase.Hooks{
			AfterConnect: func(raw *sql.DB, _ *dbtypes.DbConfig) error {
				applySQLCompatibility(inst, raw)
				return nil
			},
		})
		return inst
	})
}

func applySQLCompatibility(inst *sqlbase.DB, raw *sql.DB) {
	mode, err := readSQLCompatibility(raw)
	if err != nil {
		logger.Warn("读取 Vastbase sql_compatibility 失败，按 PostgreSQL 兼容模式继续", "error", err)
		return
	}
	inst.SetSQLCompatibility(mode)
	logger.Info("Vastbase SQL 兼容模式", "sql_compatibility", mode)
	if dialect.VastbaseOracleMode(mode) {
		inst.UseDialect(dialect.MustGet(dbtypes.DriverVastbaseOracle))
		logger.Info("Vastbase 使用 Oracle 兼容方言", "scripts", "scripts/db/vastbase_oracle")
		return
	}
	if mode != "" && !strings.EqualFold(mode, "PG") {
		logger.Warn("Vastbase 兼容模式未单独支持，按 PostgreSQL 兼容脚本初始化", "sql_compatibility", mode)
	}
}

func readSQLCompatibility(raw *sql.DB) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), compatProbeTimeout)
	defer cancel()
	var mode string
	err := raw.QueryRowContext(ctx, `SELECT current_setting('sql_compatibility'::text)`).Scan(&mode)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(mode), nil
}
