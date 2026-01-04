package ck

import (
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// GetClickHouseConnection 获取 ClickHouse 数据源连接
// ClickHouse 使用独立的数据源配置 clickhouse_main
// 返回:
//   - database.Database: ClickHouse 数据库连接实例，如果未配置则返回 nil
func GetClickHouseConnection() database.Database {
	// 从全局连接池获取 clickhouse_main 连接
	clickhouseDB := database.GetConnection("clickhouse_main")

	if clickhouseDB == nil {
		logger.Debug("未配置 clickhouse_main 数据源")
		return nil
	}

	logger.Debug("成功获取 ClickHouse 数据源", "connection", "clickhouse_main")
	return clickhouseDB
}

// IsClickHouseEnabled 检查 ClickHouse 数据源是否已启用
// 返回:
//   - bool: true 表示 ClickHouse 已配置且启用
func IsClickHouseEnabled() bool {
	return GetClickHouseConnection() != nil
}

// GetClickHouseConnectionName 获取 ClickHouse 连接名称
// 返回:
//   - string: ClickHouse 连接名称
func GetClickHouseConnectionName() string {
	return "clickhouse_main"
}
