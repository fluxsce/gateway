package dialect

import (
	"strings"

	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"
)

// errVastbaseDSN 是 Vastbase 连接串外观不对时的错误。PostgreSQL 兼容与 Oracle 兼容共用这条。
var errVastbaseDSN = huberrors.NewError("Vastbase DSN格式不正确，应以 postgresql:// 或 postgres:// 开头")

func init() {
	Register(&Spec{
		name:       dbtypes.DriverVastbase,
		openDriver: "opengauss",
		style:      DollarNum,
		timeFn:     "NOW()",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateLimitOffset(baseQuery, pageSize, offset)
		},
		limitedDel: limitedDeletePostgres,
		generate:   generatePostgreSQL,
		validate: func(dsn string) error {
			if !strings.HasPrefix(dsn, "postgresql://") && !strings.HasPrefix(dsn, "postgres://") {
				return errVastbaseDSN
			}
			return nil
		},
	})
}
