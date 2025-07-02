//go:build !no_oracle
// +build !no_oracle

// Package alldriver 导入所有支持的数据库驱动
// 使用此包可以确保所有可能用到的数据库驱动都被正确导入和注册
// 当应用需要支持多种数据库时，可以在主程序中导入此包
package alldriver

import (
	// Oracle驱动需要Oracle客户端库和C编译器支持
	// 此驱动仅在未设置 no_oracle 构建标签时被导入
	_ "gohub/pkg/database/oracle"
	// 未来可能添加的其他驱动
	// _ "gohub/pkg/database/postgres"
	// _ "gohub/pkg/database/sqlite"
)

// 此包不包含实际代码，仅用于导入其他包
// 用法：在主程序中 import _ "gohub/pkg/database/alldriver"
// Package alldriver 导入所有支持的数据库驱动
// 使用此包可以确保所有可能用到的数据库驱动都被正确导入和注册
// 当应用需要支持多种数据库时，可以在主程序中导入此包