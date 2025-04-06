// Package alldriver 导入所有支持的数据库驱动
// 使用此包可以确保所有可能用到的数据库驱动都被正确导入和注册
// 当应用需要支持多种数据库时，可以在主程序中导入此包
package alldriver

import (
	// 导入MySQL驱动包，确保其init()函数被调用
	_ "gohub/pkg/database/mysql"
	// 未来可能添加的其他驱动
	// _ "gohub/pkg/database/postgres"
	// _ "gohub/pkg/database/sqlite"
)

// 此包不包含实际代码，仅用于导入其他包
// 用法：在主程序中 import _ "gohub/pkg/database/alldriver"
