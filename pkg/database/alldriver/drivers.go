package alldriver

// Package alldriver 导入所有支持的数据库驱动
// 使用此包可以确保所有可能用到的数据库驱动都被正确导入和注册
// 当应用需要支持多种数据库时，可以在主程序中导入此包
import (
	// 导入MySQL驱动包，确保其init()函数被调用
	_ "gateway/pkg/database/mysql"
	// 导入ClickHouse驱动包，确保其init()函数被调用
	_ "gateway/pkg/database/clickhouse"
	// Oracle驱动需要Oracle客户端库和C编译器支持
	// 如需使用Oracle，请安装C编译器后取消注释以下行：
	// _ "gateway/pkg/database/oracle"
	// 未来可能添加的其他驱动
	// _ "gateway/pkg/database/postgres"
	// _ "gateway/pkg/database/sqlite"
)

// 此包不包含实际代码，仅用于导入其他包
// 用法：在主程序中 import _ "gateway/pkg/database/alldriver"
