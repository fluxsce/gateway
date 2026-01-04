package init

// 数据库脚本初始化模块
//
// 本文件作为 internal/script/db 包的适配器，将调用委托给新的架构实现。
// 这样保持了向后兼容性，同时遵循了Go项目的最佳实践。
//
// 架构说明：
// - cmd/init: 提供初始化入口和向后兼容的API
// - internal/script/db: 实现具体的业务逻辑
//
// 使用建议：
// 新代码建议直接使用 internal/script/db 包：
//   import scriptdb "gateway/internal/script/db"
//   summary, err := scriptdb.InitializeDatabaseScripts(ctx, db)

import (
	"context"
	"time"

	"gateway/internal/script/db"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
)

// ScriptExecutionResult 脚本执行结果
// 记录单个脚本文件的执行状态和结果信息
type ScriptExecutionResult = db.ScriptExecutionResult

// ScriptExecutionHistory 脚本执行历史记录（文件级别）
// 用于跟踪脚本文件的整体执行状态
type ScriptExecutionHistory = db.ScriptExecutionHistory

// StatementExecutionHistory SQL语句执行历史记录（语句级别）
// 用于跟踪每个SQL语句的执行状态，支持增量执行
type StatementExecutionHistory = db.StatementExecutionHistory

// InitializationSummary 初始化总结报告
// 包含所有数据库脚本初始化的汇总信息
type InitializationSummary = db.InitializationSummary

// InitializeDatabaseScripts 初始化数据库脚本
// 参考mongo_manager_init.go的模式，提供简洁的初始化接口
// 参数:
//   - ctx: 上下文对象，用于控制执行超时和取消
//   - database: 数据库连接实例
//
// 返回:
//   - *InitializationSummary: 初始化结果汇总报告
//   - error: 初始化失败时返回错误信息
func InitializeDatabaseScripts(ctx context.Context, database database.Database) (*InitializationSummary, error) {
	return db.InitializeDatabaseScripts(ctx, database)
}

// ListAvailableScripts 列出可用的数据库脚本文件
// 扫描脚本目录，返回所有可用的数据库初始化脚本
func ListAvailableScripts(scriptPath string) (map[string]string, error) {
	return db.ListAvailableScripts(scriptPath)
}

// CheckScriptInitializationConfig 检查脚本初始化配置
// 验证配置文件中的脚本初始化相关配置是否正确
func CheckScriptInitializationConfig() (bool, bool, int, string) {
	return db.CheckScriptInitializationConfig()
}

// GetScriptExecutionHistory 获取脚本执行历史
// 查询指定脚本的执行历史记录
// 参数:
//   - ctx: 上下文对象，用于控制查询超时和取消
//   - conn: 数据库连接实例
//   - scriptName: 脚本文件名，如果为空则查询所有脚本
//   - limit: 限制返回记录数，0表示不限制
//
// 返回:
//   - []ScriptExecutionHistory: 脚本执行历史记录列表
//   - error: 查询失败时返回错误信息
func GetScriptExecutionHistory(ctx context.Context, conn database.Database, scriptName string, limit int) ([]ScriptExecutionHistory, error) {
	return db.GetScriptExecutionHistory(ctx, conn, scriptName, limit)
}

// ForceExecuteScript 强制执行脚本
// 忽略版本检查，强制执行指定的脚本，主要用于手动维护和测试场景
// 参数:
//   - ctx: 上下文对象，用于控制执行超时和取消
//   - database: 数据库连接实例
//   - scriptPath: 脚本文件的完整路径
//
// 返回:
//   - *ScriptExecutionResult: 脚本执行结果
//   - error: 执行失败时返回错误信息
func ForceExecuteScript(ctx context.Context, database database.Database, scriptPath string) (*ScriptExecutionResult, error) {
	return db.ForceExecuteScript(ctx, database, scriptPath)
}

// InitializeDatabaseScriptsWithConfig 带配置检查的数据库脚本初始化
// 整合了配置检查、超时控制和结果日志输出的完整初始化流程
// 参数:
//   - parentCtx: 父上下文对象
//   - database: 数据库连接实例
//
// 返回:
//   - error: 初始化失败时返回错误信息
func InitializeDatabaseScriptsWithConfig(parentCtx context.Context, database database.Database) error {
	// 检查是否启用脚本初始化
	enableScriptInit := config.GetBool("database.enable_script_initialization", true)
	if !enableScriptInit {
		logger.Info("数据库脚本初始化已禁用")
		return nil
	}

	// 获取超时配置
	timeoutMinutes := config.GetInt("database.script_initialization_timeout", 30)

	// 创建脚本初始化上下文
	initCtx, cancel := context.WithTimeout(parentCtx, time.Duration(timeoutMinutes)*time.Minute)
	defer cancel()

	// 执行数据库脚本初始化
	summary, err := db.InitializeDatabaseScripts(initCtx, database)
	if err != nil {
		return err
	}

	// 输出初始化结果
	logger.Info("数据库脚本初始化完成",
		"成功数据库", summary.SuccessfulDatabases,
		"失败数据库", summary.FailedDatabases,
		"执行时间", summary.TotalDuration,
		"SQL语句数", getTotalExecutedStatements(summary))

	return nil
}

// getTotalExecutedStatements 计算总执行语句数
// 辅助函数，用于统计数据库初始化过程中执行的SQL语句总数
func getTotalExecutedStatements(summary *InitializationSummary) int {
	total := 0
	for _, result := range summary.Results {
		total += result.StatementsExecuted
	}
	return total
}
