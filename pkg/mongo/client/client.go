// Package client 提供MongoDB客户端的核心实现
//
// 此包实现了MongoDB操作的核心功能，包括：
// - 客户端连接管理
// - 数据库和集合操作
// - CRUD操作的具体实现
// - 游标和结果处理
// - 索引管理
//
// 设计原则：
// - 接口实现：实现types包中定义的接口
// - 错误处理：统一的错误处理和转换
// - 资源管理：正确的连接和资源生命周期管理
// - 线程安全：支持并发访问
//
// 文件结构：
// - client_impl.go: Client结构体和客户端相关方法
// - database.go: Database结构体和数据库相关方法
// - collection.go: Collection结构体和基本CRUD方法
// - collection_advanced.go: Collection的高级操作方法
// - cursor.go: Cursor和Result结构体和相关方法
// - utils.go: 工具函数和辅助方法
package client