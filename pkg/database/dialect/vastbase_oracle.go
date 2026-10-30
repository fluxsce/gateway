package dialect

import (
	"strings"

	"gateway/pkg/database/dbtypes"
)

func init() {
	Register(&Spec{
		name:       dbtypes.DriverVastbaseOracle,
		openDriver: "opengauss",
		style:      DollarNum,
		timeFn:     "SYSDATE",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateOffsetFetch(baseQuery, "ORDER BY ROWID", pageSize, offset)
		},
		limitedDel: limitedDeleteOracle,
		generate:   generatePostgreSQL,
		validate: func(dsn string) error {
			if !strings.HasPrefix(dsn, "postgresql://") && !strings.HasPrefix(dsn, "postgres://") {
				return errVastbaseDSN
			}
			return nil
		},
	})
}

// VastbaseOracleMode 判断 Vastbase 的 sql_compatibility 是否为 Oracle 兼容（A）。
func VastbaseOracleMode(mode string) bool {
	return strings.EqualFold(strings.TrimSpace(mode), "A")
}
