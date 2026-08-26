package dialect

import (
	"fmt"
	"net/url"
	"strings"

	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"
)

func init() {
	oracle := oracleSpec(dbtypes.DriverOracle, func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
		return paginateOffsetFetch(baseQuery, "ORDER BY ROWID", pageSize, offset)
	})
	Register(oracle)

	o11 := oracleSpec(dbtypes.DriverOracle11g, func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
		return paginateOracle11g(baseQuery, pageSize, offset)
	})
	Register(o11)
}

func oracleSpec(name string, paginate PageFunc) *Spec {
	return &Spec{
		name:       name,
		openDriver: "godror",
		style:      ColonNum,
		timeFn:     "SYSDATE",
		paginate:   paginate,
		limitedDel: limitedDeleteOracle,
		generate:   generateOracle,
		validate: func(dsn string) error {
			if !strings.HasPrefix(dsn, "oracle://") {
				return huberrors.NewError("Oracle DSN格式不正确，应以oracle://开头")
			}
			return nil
		},
	}
}

func generateOracle(config *dbtypes.DbConfig) (string, error) {
	if config.Connection.UseSID && config.Connection.SID != "" {
		return GenerateOracleWithSID(config, config.Connection.SID)
	}

	if config.Connection.Host == "" {
		return "", huberrors.NewError("Oracle数据库需要host参数")
	}
	if config.Connection.Username == "" {
		return "", huberrors.NewError("Oracle数据库需要username参数")
	}
	if config.Connection.Password == "" {
		return "", huberrors.NewError("Oracle数据库需要password参数")
	}

	port := 1521
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	serviceName := config.Connection.Database
	if serviceName == "" {
		return "", huberrors.NewError("Oracle数据库需要database参数(作为服务名)")
	}

	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		serviceName)

	params := make([]string, 0)

	connectionTimeout := 30
	if config.Connection.OracleConnectionTimeout > 0 {
		connectionTimeout = config.Connection.OracleConnectionTimeout
	}
	params = append(params, fmt.Sprintf("CONNECTION_TIMEOUT=%ds", connectionTimeout))

	readTimeout := 30
	if config.Connection.OracleReadTimeout > 0 {
		readTimeout = config.Connection.OracleReadTimeout
	}
	params = append(params, fmt.Sprintf("READ_TIMEOUT=%ds", readTimeout))

	writeTimeout := 30
	if config.Connection.OracleWriteTimeout > 0 {
		writeTimeout = config.Connection.OracleWriteTimeout
	}
	params = append(params, fmt.Sprintf("WRITE_TIMEOUT=%ds", writeTimeout))

	timezone := "Asia/Shanghai"
	if config.Connection.Timezone != "" {
		timezone = config.Connection.Timezone
	}
	params = append(params, fmt.Sprintf("TIMEZONE=%s", timezone))

	nlsLang := "AMERICAN_AMERICA.UTF8"
	if config.Connection.NLSLang != "" {
		nlsLang = config.Connection.NLSLang
	}
	params = append(params, fmt.Sprintf("NLS_LANG=%s", nlsLang))
	params = append(params, "CHARSET=UTF8")
	params = append(params, "NLS_CHARACTERSET=AL32UTF8")

	prefetchRows := 500
	if config.Connection.PrefetchRows > 0 {
		prefetchRows = config.Connection.PrefetchRows
	}
	params = append(params, fmt.Sprintf("PREFETCH_ROWS=%d", prefetchRows))

	lobPrefetchSize := 4096
	if config.Connection.LobPrefetchSize > 0 {
		lobPrefetchSize = config.Connection.LobPrefetchSize
	}
	params = append(params, fmt.Sprintf("LOB_PREFETCH_SIZE=%d", lobPrefetchSize))

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}
	return dsn, nil
}

// GenerateOracleWithSID 生成使用 SID 的 Oracle DSN，供配置和测试直接调用。
func GenerateOracleWithSID(config *dbtypes.DbConfig, sid string) (string, error) {
	if config.Connection.Host == "" {
		return "", huberrors.NewError("Oracle数据库需要host参数")
	}
	if config.Connection.Username == "" {
		return "", huberrors.NewError("Oracle数据库需要username参数")
	}
	if config.Connection.Password == "" {
		return "", huberrors.NewError("Oracle数据库需要password参数")
	}
	if sid == "" {
		return "", huberrors.NewError("SID参数不能为空")
	}

	port := 1521
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	return fmt.Sprintf("oracle://%s:%s@%s:%d?sid=%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		sid), nil
}
