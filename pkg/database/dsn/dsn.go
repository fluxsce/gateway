// Package dsn 提供数据库连接字符串(DSN)生成功能
// 为不同的数据库类型提供统一的DSN生成接口
package dsn

import (
	"fmt"
	"gohub/pkg/database/dbtypes"
	huberrors "gohub/pkg/utils/huberrors"
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

	// 组装完整DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		config.Connection.Username,
		config.Connection.Password,
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

	// 组装完整DSN
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		config.Connection.Username,
		config.Connection.Password,
		config.Connection.Host,
		port,
		config.Connection.Database,
		sslmode,
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
	// 构建SQLite DSN
	// 格式: file:database?mode=rw
	dsn := fmt.Sprintf("file:%s?mode=rw", config.Connection.Database)
	return dsn, nil
}
