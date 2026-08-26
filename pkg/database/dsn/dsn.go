// Package dsn 提供数据库连接字符串(DSN)生成功能。
// 具体生成规则在 dialect 中按驱动注册，本包保持原有导出函数供测试和调用方使用。
package dsn

import (
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dialect"
	huberrors "gateway/pkg/utils/huberrors"
)

func generateByDriver(driver string, config *dbtypes.DbConfig) (string, error) {
	d, err := dialect.Get(driver)
	if err != nil {
		return "", huberrors.NewError("不支持的数据库驱动类型: %s", driver)
	}
	return d.GenerateDSN(config)
}

// Generate 根据数据库配置生成对应的DSN连接字符串。
func Generate(config *dbtypes.DbConfig) (string, error) {
	if config.DSN != "" {
		return config.DSN, nil
	}
	return generateByDriver(config.Driver, config)
}

// GenerateMySQL 生成MySQL数据库的DSN连接字符串。
func GenerateMySQL(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverMySQL, config)
}

// GeneratePostgreSQL 生成PostgreSQL数据库的DSN连接字符串。
func GeneratePostgreSQL(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverPostgreSQL, config)
}

// GenerateSQLite 生成SQLite数据库的DSN连接字符串。
func GenerateSQLite(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverSQLite, config)
}

// GenerateOracle 生成Oracle数据库的DSN连接字符串。
func GenerateOracle(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverOracle, config)
}

// GenerateOracleWithSID 生成使用SID的Oracle数据库DSN连接字符串。
func GenerateOracleWithSID(config *dbtypes.DbConfig, sid string) (string, error) {
	return dialect.GenerateOracleWithSID(config, sid)
}

// GenerateClickHouse 生成ClickHouse数据库的DSN连接字符串。
func GenerateClickHouse(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverClickHouse, config)
}

// GenerateSQLServer 生成 SQL Server 数据库的 DSN 连接字符串。
func GenerateSQLServer(config *dbtypes.DbConfig) (string, error) {
	return generateByDriver(dbtypes.DriverSQLServer, config)
}

// ValidateDSN 验证生成的DSN是否符合格式要求。
func ValidateDSN(driver string, dsn string) error {
	if dsn == "" {
		return huberrors.NewError("DSN不能为空")
	}
	d, err := dialect.Get(driver)
	if err != nil {
		return huberrors.NewError("不支持的数据库驱动类型: %s", driver)
	}
	return d.ValidateDSN(dsn)
}
