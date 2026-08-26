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
		name:       dbtypes.DriverClickHouse,
		openDriver: "clickhouse",
		style:      Question,
		timeFn:     "now()",
		paginate: func(baseQuery string, page, pageSize, offset int) (string, []interface{}, error) {
			return paginateLimitOffset(baseQuery, pageSize, offset)
		},
		limitedDel: limitedDeleteLimit,
		generate:   generateClickHouse,
		validate: func(dsn string) error {
			if !strings.HasPrefix(dsn, "clickhouse://") {
				return huberrors.NewError("ClickHouse DSN格式不正确，应以clickhouse://开头")
			}
			return nil
		},
		updateFmt: "ALTER TABLE %s UPDATE %s",
		deleteFmt: "ALTER TABLE %s DELETE",
		inDelete: "ALTER TABLE %s DELETE WHERE %s IN (%s)",
	})
}

func generateClickHouse(config *dbtypes.DbConfig) (string, error) {
	if config.Connection.Host == "" {
		return "", huberrors.NewError("ClickHouse数据库需要host参数")
	}
	if config.Connection.Username == "" {
		return "", huberrors.NewError("ClickHouse数据库需要username参数")
	}
	if config.Connection.Database == "" {
		return "", huberrors.NewError("ClickHouse数据库需要database参数")
	}

	port := 9000
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	} else if config.Connection.ClickHouseSecure {
		port = 9440
	}

	hostList := fmt.Sprintf("%s:%d", config.Connection.Host, port)
	if config.Connection.ClickHouseHosts != "" {
		hostList += "," + config.Connection.ClickHouseHosts
	}

	dsn := fmt.Sprintf("clickhouse://%s", hostList)

	params := make([]string, 0)
	params = append(params, "database="+url.QueryEscape(config.Connection.Database))
	params = append(params, "username="+url.QueryEscape(config.Connection.Username))
	if config.Connection.Password != "" {
		params = append(params, "password="+url.QueryEscape(config.Connection.Password))
	}

	dialTimeout := 30
	if config.Connection.ClickHouseDialTimeout > 0 {
		dialTimeout = config.Connection.ClickHouseDialTimeout
	}
	params = append(params, fmt.Sprintf("dial_timeout=%ds", dialTimeout))

	compress := config.Connection.ClickHouseCompress
	if compress == "" {
		compress = "none"
	}
	validCompressAlgos := map[string]bool{
		"none": true, "lz4": true, "zstd": true,
		"gzip": true, "deflate": true, "br": true,
		"true": true, "false": true,
	}
	if !validCompressAlgos[compress] {
		return "", huberrors.NewError("不支持的压缩算法: %s，支持: none,lz4,zstd,gzip,deflate,br", compress)
	}
	if compress == "true" {
		compress = "lz4"
	} else if compress == "false" {
		compress = "none"
	}
	params = append(params, "compress="+compress)

	if compress != "none" && config.Connection.ClickHouseCompressLevel > 0 {
		params = append(params, fmt.Sprintf("compress_level=%d", config.Connection.ClickHouseCompressLevel))
	}
	if config.Connection.ClickHouseSecure {
		params = append(params, "secure=true")
	}
	if config.Connection.ClickHouseSkipVerify {
		params = append(params, "skip_verify=true")
	}
	if config.Connection.ClickHouseDebug {
		params = append(params, "debug=true")
	}
	if config.Connection.ClickHouseBlockBufferSize > 0 {
		params = append(params, fmt.Sprintf("block_buffer_size=%d", config.Connection.ClickHouseBlockBufferSize))
	}
	if config.Connection.ClickHouseConnOpenStrategy != "" {
		validStrategies := map[string]bool{"random": true, "in_order": true}
		if validStrategies[config.Connection.ClickHouseConnOpenStrategy] {
			params = append(params, "connection_open_strategy="+config.Connection.ClickHouseConnOpenStrategy)
		}
	}

	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}
	return dsn, nil
}
