package dialect

import (
	"errors"
	"fmt"
	"strings"

	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"

	mysqldriver "github.com/go-sql-driver/mysql"
)

func init() {
	Register(&Spec{
		name:       dbtypes.DriverMySQL,
		aliases:    []string{dbtypes.DriverMariaDB, dbtypes.DriverTiDB},
		openDriver: "mysql",
		style:      Question,
		timeFn:     "NOW()",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateLimitOffset(baseQuery, pageSize, offset)
		},
		limitedDel: limitedDeleteLimit,
		generate:   generateMySQL,
		validate: func(dsn string) error {
			if !strings.Contains(dsn, "@tcp(") {
				return huberrors.NewError("MySQL DSN格式不正确，缺少@tcp部分")
			}
			return nil
		},
		classify: classifyMySQL,
	})
}

func classifyMySQL(err error) ErrorClass {
	var me *mysqldriver.MySQLError
	if errors.As(err, &me) {
		switch me.Number {
		case 1062, 1169, 1022:
			return ClassDuplicateKey
		case 1040, 2002, 2003, 2006, 2013:
			return ClassConnection
		case 1064:
			return ClassInvalidQuery
		}
	}
	return ClassifyByMessage(err)
}

func generateMySQL(config *dbtypes.DbConfig) (string, error) {
	params := make(map[string]string)

	// 未配置 charset 时不写 DSN，沿用库/服务器默认，避免覆盖企业 utf8 / utf8_bin。
	if config.Connection.Charset != "" {
		params["charset"] = config.Connection.Charset
	}

	if config.Connection.ParseTime {
		params["parseTime"] = "True"
	} else {
		params["parseTime"] = "False"
	}

	if config.Connection.Loc != "" {
		params["loc"] = config.Connection.Loc
	} else {
		params["loc"] = "Local"
	}

	if config.Connection.MySQLConnectTimeout > 0 {
		params["timeout"] = fmt.Sprintf("%ds", config.Connection.MySQLConnectTimeout)
	}
	if config.Connection.MySQLReadTimeout > 0 {
		params["readTimeout"] = fmt.Sprintf("%ds", config.Connection.MySQLReadTimeout)
	}
	if config.Connection.MySQLWriteTimeout > 0 {
		params["writeTimeout"] = fmt.Sprintf("%ds", config.Connection.MySQLWriteTimeout)
	}

	var paramStr string
	for k, v := range params {
		if paramStr != "" {
			paramStr += "&"
		}
		paramStr += k + "=" + v
	}

	port := 3306
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		config.Connection.Username,
		config.Connection.Password,
		config.Connection.Host,
		port,
		config.Connection.Database,
	)
	if paramStr != "" {
		dsn += "?" + paramStr
	}
	return dsn, nil
}
