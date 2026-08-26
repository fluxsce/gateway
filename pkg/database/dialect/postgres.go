package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"
)

func init() {
	Register(&Spec{
		name:       dbtypes.DriverPostgreSQL,
		openDriver: "postgres",
		style:      DollarNum,
		timeFn:     "NOW()",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateLimitOffset(baseQuery, pageSize, offset)
		},
		limitedDel: limitedDeletePostgres,
		generate:   generatePostgreSQL,
		validate: func(dsn string) error {
			if !strings.HasPrefix(dsn, "postgresql://") {
				return huberrors.NewError("PostgreSQL DSN格式不正确，应以postgresql://开头")
			}
			return nil
		},
	})
}

func generatePostgreSQL(config *dbtypes.DbConfig) (string, error) {
	sslmode := "disable"
	if config.Connection.SSLMode != "" {
		sslmode = config.Connection.SSLMode
	}

	port := 5432
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	params := make([]string, 0)
	params = append(params, "sslmode="+sslmode)

	if config.Connection.PostgreSQLConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%ds", config.Connection.PostgreSQLConnectTimeout))
	}
	if config.Connection.PostgreSQLStatementTimeout > 0 {
		params = append(params, fmt.Sprintf("statement_timeout=%ds", config.Connection.PostgreSQLStatementTimeout))
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		config.Connection.Database,
		strings.Join(params, "&"),
	)
	return dsn, nil
}
