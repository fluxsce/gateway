// Package dsn 提供数据库连接字符串(DSN)生成功能
// 为不同的数据库类型提供统一的DSN生成接口
package dsn

import (
	"fmt"
	"gateway/pkg/database/dbtypes"
	huberrors "gateway/pkg/utils/huberrors"
	"net/url"
	"strings"
)

// Generate 根据数据库配置生成对应的DSN连接字符串
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: 生成的DSN连接字符串
//   - error: 错误信息
func Generate(config *dbtypes.DbConfig) (string, error) {
	// 如果配置中已经有DSN，直接返回
	if config.DSN != "" {
		return config.DSN, nil
	}

	// 根据驱动类型生成对应的DSN
	switch config.Driver {
	case dbtypes.DriverMySQL:
		return GenerateMySQL(config)
	case dbtypes.DriverPostgreSQL:
		return GeneratePostgreSQL(config)
	case dbtypes.DriverSQLite:
		return GenerateSQLite(config)
	case dbtypes.DriverOracle:
		// 如果配置指定使用SID而不是服务名，调用特殊的SID连接字符串生成函数
		if config.Connection.UseSID && config.Connection.SID != "" {
			return GenerateOracleWithSID(config, config.Connection.SID)
		}
		return GenerateOracle(config)
	case dbtypes.DriverClickHouse:
		return GenerateClickHouse(config)
	default:
		return "", huberrors.NewError("不支持的数据库驱动类型: %s", config.Driver)
	}
}

// GenerateMySQL 生成MySQL数据库的DSN连接字符串
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: MySQL格式的DSN
//   - error: 错误信息
func GenerateMySQL(config *dbtypes.DbConfig) (string, error) {
	// 构建MySQL DSN
	// 格式: username:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
	params := make(map[string]string)

	// 设置字符集
	if config.Connection.Charset != "" {
		params["charset"] = config.Connection.Charset
	} else {
		params["charset"] = "utf8mb4"
	}

	// 设置时间解析
	if config.Connection.ParseTime {
		params["parseTime"] = "True"
	} else {
		params["parseTime"] = "False"
	}

	// 设置时区
	if config.Connection.Loc != "" {
		params["loc"] = config.Connection.Loc
	} else {
		params["loc"] = "Local"
	}

	// 设置超时参数
	if config.Connection.MySQLConnectTimeout > 0 {
		params["timeout"] = fmt.Sprintf("%ds", config.Connection.MySQLConnectTimeout)
	}
	if config.Connection.MySQLReadTimeout > 0 {
		params["readTimeout"] = fmt.Sprintf("%ds", config.Connection.MySQLReadTimeout)
	}
	if config.Connection.MySQLWriteTimeout > 0 {
		params["writeTimeout"] = fmt.Sprintf("%ds", config.Connection.MySQLWriteTimeout)
	}

	// 构建参数字符串
	var paramStr string
	for k, v := range params {
		if paramStr != "" {
			paramStr += "&"
		}
		paramStr += k + "=" + v
	}

	// 获取端口，默认为3306
	port := 3306
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	// 组装完整DSN - 对用户名和密码进行URL编码以支持特殊字符
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		config.Connection.Database,
		paramStr,
	)

	return dsn, nil
}

// GeneratePostgreSQL 生成PostgreSQL数据库的DSN连接字符串
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: PostgreSQL格式的DSN
//   - error: 错误信息
func GeneratePostgreSQL(config *dbtypes.DbConfig) (string, error) {
	// 构建PostgreSQL DSN
	// 格式: postgresql://username:password@host:port/database?sslmode=disable

	// 获取SSL模式，默认为disable
	sslmode := "disable"
	if config.Connection.SSLMode != "" {
		sslmode = config.Connection.SSLMode
	}

	// 获取端口，默认为5432
	port := 5432
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	// 构建PostgreSQL参数
	params := make([]string, 0)
	params = append(params, "sslmode="+sslmode)

	// 设置超时参数 - PostgreSQL需要时间单位
	if config.Connection.PostgreSQLConnectTimeout > 0 {
		params = append(params, fmt.Sprintf("connect_timeout=%ds", config.Connection.PostgreSQLConnectTimeout))
	}
	if config.Connection.PostgreSQLStatementTimeout > 0 {
		params = append(params, fmt.Sprintf("statement_timeout=%ds", config.Connection.PostgreSQLStatementTimeout))
	}

	// 组装完整DSN - 对用户名和密码进行URL编码以支持特殊字符
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

// GenerateSQLite 生成SQLite数据库的DSN连接字符串
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: SQLite格式的DSN
//   - error: 错误信息
func GenerateSQLite(config *dbtypes.DbConfig) (string, error) {
	// SQLite的数据库"名称"实际上是文件路径
	// 如果Database字段为空或者是特殊值，使用默认配置
	dbPath := config.Connection.Database

	// 处理特殊情况
	if dbPath == "" || dbPath == ":memory:" {
		return ":memory:", nil
	}

	// 如果Database字段看起来不像文件路径，使用默认路径
	if !strings.Contains(dbPath, ".") && !strings.Contains(dbPath, "/") && !strings.Contains(dbPath, "\\") {
		dbPath = fmt.Sprintf("./%s.db", dbPath)
	}

	// 构建SQLite参数
	params := make([]string, 0)

	// 默认参数
	params = append(params, "cache=shared")        // 共享缓存
	params = append(params, "mode=rwc")            // 读写创建模式
	params = append(params, "_journal_mode=WAL")   // WAL模式
	params = append(params, "_synchronous=NORMAL") // 正常同步
	params = append(params, "_foreign_keys=1")     // 启用外键

	// 设置busy_timeout，如果配置了则使用配置值，否则使用默认5秒
	busyTimeout := 5000 // 默认5秒
	if config.Connection.BusyTimeout > 0 {
		busyTimeout = config.Connection.BusyTimeout
	}
	params = append(params, fmt.Sprintf("_busy_timeout=%d", busyTimeout))

	// 如果有参数，使用file:前缀
	if len(params) > 0 {
		dsn := fmt.Sprintf("file:%s?%s", dbPath, strings.Join(params, "&"))
		return dsn, nil
	}

	// 简单文件路径
	return dbPath, nil
}

// GenerateOracle 生成Oracle数据库的DSN连接字符串
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: Oracle格式的DSN
//   - error: 错误信息
func GenerateOracle(config *dbtypes.DbConfig) (string, error) {
	// 验证必需参数
	if config.Connection.Host == "" {
		return "", huberrors.NewError("Oracle数据库需要host参数")
	}
	if config.Connection.Username == "" {
		return "", huberrors.NewError("Oracle数据库需要username参数")
	}
	if config.Connection.Password == "" {
		return "", huberrors.NewError("Oracle数据库需要password参数")
	}

	// 获取端口，默认为1521
	port := 1521
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	// 确定服务名或SID
	// 优先使用Database字段作为服务名，这是最常见的配置方式
	serviceName := config.Connection.Database
	if serviceName == "" {
		return "", huberrors.NewError("Oracle数据库需要database参数(作为服务名)")
	}

	// 构建基本连接字符串 - 对用户名和密码进行URL编码以支持特殊字符
	// Oracle DSN格式: oracle://username:password@host:port/service_name
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d/%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		serviceName)

	// 构建Oracle特有参数
	params := make([]string, 0)

	// 设置超时参数，如果配置了则使用配置值，否则使用默认30秒 - Oracle需要时间单位
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

	// 时区设置
	timezone := "Asia/Shanghai" // 默认使用中国时区
	if config.Connection.Timezone != "" {
		timezone = config.Connection.Timezone
	}
	params = append(params, fmt.Sprintf("TIMEZONE=%s", timezone))

	// 字符集设置
	nlsLang := "AMERICAN_AMERICA.UTF8"
	if config.Connection.NLSLang != "" {
		nlsLang = config.Connection.NLSLang
	}
	params = append(params, fmt.Sprintf("NLS_LANG=%s", nlsLang))

	// 添加Oracle字符集参数 - 解决中文字符编码问题
	params = append(params, "CHARSET=UTF8")              // 明确指定连接字符集
	params = append(params, "NLS_CHARACTERSET=AL32UTF8") // 数据库字符集

	// 性能优化参数
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

	// 如果有参数，添加到DSN
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn, nil
}

// GenerateOracleWithSID 生成使用SID的Oracle数据库DSN连接字符串
// 这是一个辅助函数，用于生成传统的Oracle SID连接方式
// 参数:
//   - config: 数据库配置
//   - sid: Oracle SID
//
// 返回:
//   - string: Oracle格式的DSN (使用SID)
//   - error: 错误信息
func GenerateOracleWithSID(config *dbtypes.DbConfig, sid string) (string, error) {
	// 验证必需参数
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

	// 获取端口，默认为1521
	port := 1521
	if config.Connection.Port > 0 {
		port = config.Connection.Port
	}

	// 构建SID连接字符串 - 对用户名和密码进行URL编码以支持特殊字符
	// Oracle SID DSN格式: oracle://username:password@host:port?sid=SID
	dsn := fmt.Sprintf("oracle://%s:%s@%s:%d?sid=%s",
		url.QueryEscape(config.Connection.Username),
		url.QueryEscape(config.Connection.Password),
		config.Connection.Host,
		port,
		sid)

	return dsn, nil
}

// GenerateClickHouse 生成ClickHouse数据库的DSN连接字符串
// 按照官网标准实现: https://clickhouse.com/docs/zh/integrations/go#databasesql-api
//
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - string: ClickHouse格式的DSN
//   - error: 错误信息
func GenerateClickHouse(config *dbtypes.DbConfig) (string, error) {
	// 验证必需参数
	if config.Connection.Host == "" {
		return "", huberrors.NewError("ClickHouse数据库需要host参数")
	}
	if config.Connection.Username == "" {
		return "", huberrors.NewError("ClickHouse数据库需要username参数")
	}
	if config.Connection.Database == "" {
		return "", huberrors.NewError("ClickHouse数据库需要database参数")
	}

	// 获取端口 - 按官网标准：TLS为9440，非TLS为9000
	port := 9000 // 默认非TLS端口
	if config.Connection.Port > 0 {
		port = config.Connection.Port // 如果明确配置了端口，使用配置值
	} else if config.Connection.ClickHouseSecure {
		port = 9440 // TLS默认端口（仅在未明确配置端口时）
	}

	// 构建主机地址列表 - 官网标准：多个主机用逗号分隔
	// 格式: clickhouse://host1:port1,host2:port2/database
	hostList := fmt.Sprintf("%s:%d", config.Connection.Host, port)

	// 如果配置了额外的主机，追加到地址列表
	if config.Connection.ClickHouseHosts != "" {
		hostList += "," + config.Connection.ClickHouseHosts
	}

	// 构建基本连接字符串
	// ClickHouse官网标准DSN格式: clickhouse://host1:port1,host2:port2?database=dbname&username=user&password=pass
	dsn := fmt.Sprintf("clickhouse://%s", hostList)

	// 构建参数 - 对数据库名、用户名和密码进行URL编码以支持特殊字符
	params := make([]string, 0)
	params = append(params, "database="+url.QueryEscape(config.Connection.Database))
	params = append(params, "username="+url.QueryEscape(config.Connection.Username))
	if config.Connection.Password != "" {
		params = append(params, "password="+url.QueryEscape(config.Connection.Password))
	}

	// === ClickHouse官网标准DSN参数 ===

	// 拨号超时设置 - 官网标准格式
	dialTimeout := 30 // 默认30秒
	if config.Connection.ClickHouseDialTimeout > 0 {
		dialTimeout = config.Connection.ClickHouseDialTimeout
	}
	params = append(params, fmt.Sprintf("dial_timeout=%ds", dialTimeout))

	// 压缩设置 - 按官网标准支持算法名称
	// 支持: none, lz4, zstd, gzip, deflate, br
	compress := config.Connection.ClickHouseCompress
	if compress == "" {
		compress = "none" // 官网默认值
	}
	// 验证压缩算法是否支持
	validCompressAlgos := map[string]bool{
		"none": true, "lz4": true, "zstd": true,
		"gzip": true, "deflate": true, "br": true,
		"true": true, "false": true, // 向后兼容
	}
	if !validCompressAlgos[compress] {
		return "", huberrors.NewError("不支持的压缩算法: %s，支持: none,lz4,zstd,gzip,deflate,br", compress)
	}
	// true转换为lz4，false转换为none
	if compress == "true" {
		compress = "lz4"
	} else if compress == "false" {
		compress = "none"
	}
	params = append(params, "compress="+compress)

	// 压缩级别设置 - 仅在启用压缩时有效
	if compress != "none" && config.Connection.ClickHouseCompressLevel > 0 {
		params = append(params, fmt.Sprintf("compress_level=%d", config.Connection.ClickHouseCompressLevel))
	}

	// SSL/TLS设置 - 官网标准
	if config.Connection.ClickHouseSecure {
		params = append(params, "secure=true")
	}

	// 跳过证书验证 - 官网标准
	if config.Connection.ClickHouseSkipVerify {
		params = append(params, "skip_verify=true")
	}

	// 调试模式设置 - 官网标准
	if config.Connection.ClickHouseDebug {
		params = append(params, "debug=true")
	}

	// 块缓冲区大小 - 官网标准
	if config.Connection.ClickHouseBlockBufferSize > 0 {
		params = append(params, fmt.Sprintf("block_buffer_size=%d", config.Connection.ClickHouseBlockBufferSize))
	}

	// 连接打开策略 - 官网标准 (random/in_order)
	if config.Connection.ClickHouseConnOpenStrategy != "" {
		validStrategies := map[string]bool{"random": true, "in_order": true}
		if validStrategies[config.Connection.ClickHouseConnOpenStrategy] {
			params = append(params, "connection_open_strategy="+config.Connection.ClickHouseConnOpenStrategy)
		}
	}

	// === ClickHouse集群和高级参数 ===
	// 注意：多个主机已在地址部分处理，不需要hosts参数

	// 如果有参数，添加到DSN
	if len(params) > 0 {
		dsn += "?" + strings.Join(params, "&")
	}

	return dsn, nil
}

// ValidateDSN 验证生成的DSN是否符合格式要求
// 参数:
//   - driver: 数据库驱动类型
//   - dsn: 要验证的DSN字符串
//
// 返回:
//   - error: 验证失败时返回错误信息
func ValidateDSN(driver string, dsn string) error {
	if dsn == "" {
		return huberrors.NewError("DSN不能为空")
	}

	switch driver {
	case dbtypes.DriverMySQL:
		if !strings.Contains(dsn, "@tcp(") {
			return huberrors.NewError("MySQL DSN格式不正确，缺少@tcp部分")
		}
	case dbtypes.DriverPostgreSQL:
		if !strings.HasPrefix(dsn, "postgresql://") {
			return huberrors.NewError("PostgreSQL DSN格式不正确，应以postgresql://开头")
		}
	case dbtypes.DriverSQLite:
		// SQLite DSN比较灵活，基本不需要特殊验证
		if dsn != ":memory:" && !strings.Contains(dsn, ".") && !strings.HasPrefix(dsn, "file:") {
			return huberrors.NewError("SQLite DSN格式可能不正确")
		}
	case dbtypes.DriverOracle:
		if !strings.HasPrefix(dsn, "oracle://") {
			return huberrors.NewError("Oracle DSN格式不正确，应以oracle://开头")
		}
	case dbtypes.DriverClickHouse:
		if !strings.HasPrefix(dsn, "clickhouse://") {
			return huberrors.NewError("ClickHouse DSN格式不正确，应以clickhouse://开头")
		}
	default:
		return huberrors.NewError("不支持的数据库驱动类型: %s", driver)
	}

	return nil
}
