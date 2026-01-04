package db

import (
	"context"
	"strings"
	"time"

	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/logger"
	"gateway/pkg/utils/random"
)

// getStatementExecutionStatus 获取语句的执行状态
// 查询语句执行历史表，获取指定语句的最新执行状态
// 参数:
//   - ctx: 上下文对象，用于控制查询超时和取消
//   - conn: 数据库连接实例
//   - driver: 数据库驱动类型
//   - scriptName: 脚本文件名
//   - statementHash: 语句哈希值
//
// 返回:
//   - string: 执行状态（SUCCESS, FAILED, SKIPPED, 空字符串表示未执行过）
//   - error: 查询失败时返回错误信息
func getStatementExecutionStatus(ctx context.Context, conn database.Database, driver, scriptName, statementHash string) (string, error) {
	tenantId := config.GetString("database.tenant_id", "default")

	query := `SELECT executionStatus FROM HUB_STATEMENT_EXECUTION_HISTORY 
			  WHERE tenantId = ? AND scriptName = ? AND statementHash = ? 
			  AND databaseDriver = ? 
			  ORDER BY executionTime DESC LIMIT 1`

	// 定义结果结构体
	type StatusResult struct {
		ExecutionStatus string `db:"executionStatus"`
	}

	var result StatusResult
	err := conn.QueryOne(ctx, &result, query, []interface{}{tenantId, scriptName, statementHash, driver}, true)
	if err != nil {
		// 判断是否是预期的"记录不存在"或"表不存在"错误
		isRecordNotFound := strings.Contains(err.Error(), "no rows") ||
			strings.Contains(err.Error(), "record not found") ||
			strings.Contains(err.Error(), "not found")
		isTableNotExist := strings.Contains(err.Error(), "no such table") ||
			strings.Contains(err.Error(), "doesn't exist") ||
			(strings.Contains(err.Error(), "table") && strings.Contains(err.Error(), "not exist"))

		// 如果是预期的情况，认为语句未执行过
		if isRecordNotFound || isTableNotExist {
			logger.Debug("语句未执行过或历史表不存在",
				"table", "HUB_STATEMENT_EXECUTION_HISTORY",
				"script", scriptName,
				"statement_hash", statementHash)
			return "", nil
		}
		// 其他错误才返回错误信息
		return "", err
	}

	logger.Debug("查询语句执行状态完成",
		"script", scriptName,
		"statement_hash", statementHash,
		"driver", driver,
		"status", result.ExecutionStatus)

	return result.ExecutionStatus, nil
}

// recordScriptExecution 记录脚本执行历史
// 将脚本执行结果保存到历史表中，用于跟踪脚本执行状态和防止重复执行
// 参数:
//   - ctx: 上下文对象，用于控制插入超时和取消
//   - conn: 数据库连接实例
//   - driver: 数据库驱动类型
//   - scriptName: 脚本文件名
//   - scriptPath: 脚本完整路径
//   - scriptVersion: 脚本版本（MD5哈希值）
//   - status: 执行状态（SUCCESS, FAILED, SKIPPED）
//   - duration: 执行耗时
//   - statementsExecuted: 执行的SQL语句数量
//   - errorMessage: 错误信息（如果有）
func recordScriptExecution(ctx context.Context, conn database.Database, driver, scriptName, scriptPath, scriptVersion, status string, duration time.Duration, statementsExecuted int, errorMessage string) {
	tenantId := config.GetString("database.tenant_id", "default")
	executionId := generateExecutionId()
	now := time.Now()
	durationMs := duration.Milliseconds()

	// 统一先查询校验，存在更新不存在插入
	// 唯一键：UK_SCRIPT_VERSION (tenantId, scriptName, scriptVersion, databaseDriver)
	checkQuery := `SELECT executionId FROM HUB_SCRIPT_EXECUTION_HISTORY 
				   WHERE tenantId = ? AND scriptName = ? AND scriptVersion = ? AND databaseDriver = ?`

	type ExistingRecord struct {
		ExecutionId string `db:"executionId"`
	}

	var existing ExistingRecord
	err := conn.QueryOne(ctx, &existing, checkQuery, []interface{}{tenantId, scriptName, scriptVersion, driver}, true)

	// 判断是否是真正的错误（排除"记录不存在"的情况）
	isRecordNotFound := err != nil && (strings.Contains(err.Error(), "no rows") ||
		strings.Contains(err.Error(), "record not found") ||
		strings.Contains(err.Error(), "not found"))
	isTableNotExist := err != nil && (strings.Contains(err.Error(), "no such table") ||
		strings.Contains(err.Error(), "doesn't exist") ||
		(strings.Contains(err.Error(), "table") && strings.Contains(err.Error(), "not exist")))

	// 只有在非预期错误时才记录错误日志并返回
	if err != nil && !isRecordNotFound && !isTableNotExist {
		logger.Error("检查脚本执行历史记录失败",
			"error", err,
			"script", scriptName,
			"version", scriptVersion)
		return
	}

	if existing.ExecutionId != "" {
		// 记录存在，执行更新
		updateSQL := `UPDATE HUB_SCRIPT_EXECUTION_HISTORY 
					  SET executionId = ?, scriptPath = ?, executionStatus = ?, 
					      executionTime = ?, executionDuration = ?, statementsExecuted = ?, 
					      errorMessage = ?
					  WHERE tenantId = ? AND scriptName = ? AND scriptVersion = ? AND databaseDriver = ?`

		_, err = conn.Exec(ctx, updateSQL, []interface{}{
			executionId, scriptPath, status, now, durationMs, statementsExecuted, errorMessage,
			tenantId, scriptName, scriptVersion, driver,
		}, false)

		if err != nil {
			logger.Error("更新脚本执行历史失败",
				"error", err,
				"script", scriptName,
				"version", scriptVersion,
				"execution_id", existing.ExecutionId,
				"status", status)
		} else {
			logger.Debug("脚本执行历史更新成功",
				"script", scriptName,
				"version", scriptVersion,
				"execution_id", existing.ExecutionId,
				"status", status,
				"duration_ms", durationMs)
		}
	} else {
		// 记录不存在，执行插入
		insertSQL := `INSERT INTO HUB_SCRIPT_EXECUTION_HISTORY 
					  (executionId, tenantId, scriptName, scriptPath, scriptVersion, databaseDriver, 
					   executionStatus, executionTime, executionDuration, statementsExecuted, errorMessage, createdAt)
					  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = conn.Exec(ctx, insertSQL, []interface{}{
			executionId, tenantId, scriptName, scriptPath, scriptVersion, driver,
			status, now, durationMs, statementsExecuted, errorMessage, now,
		}, false)

		if err != nil {
			logger.Error("插入脚本执行历史失败",
				"error", err,
				"script", scriptName,
				"version", scriptVersion,
				"execution_id", executionId,
				"status", status)
		} else {
			logger.Debug("脚本执行历史插入成功",
				"script", scriptName,
				"version", scriptVersion,
				"execution_id", executionId,
				"status", status,
				"duration_ms", durationMs)
		}
	}
}

// recordStatementExecution 记录SQL语句执行历史
// 将SQL语句执行结果保存到语句历史表中，如果记录已存在则更新，不存在则插入
// 参数:
//   - ctx: 上下文对象，用于控制操作超时和取消
//   - conn: 数据库连接实例
//   - driver: 数据库驱动类型
//   - scriptName: 脚本文件名
//   - statementHash: 语句哈希值
//   - statementType: 语句类型
//   - statementContent: 语句内容
//   - status: 执行状态（SUCCESS, FAILED, SKIPPED）
//   - duration: 执行耗时
//   - errorMessage: 错误信息（如果有）
func recordStatementExecution(ctx context.Context, conn database.Database, driver, scriptName, statementHash, statementType, statementContent, status string, duration time.Duration, errorMessage string) {
	tenantId := config.GetString("database.tenant_id", "default")
	now := time.Now()
	durationMs := duration.Milliseconds()

	// 首先检查记录是否存在
	checkQuery := `SELECT statementId FROM HUB_STATEMENT_EXECUTION_HISTORY 
				   WHERE tenantId = ? AND scriptName = ? AND statementHash = ? AND databaseDriver = ?`

	type ExistingRecord struct {
		StatementId string `db:"statementId"`
	}

	var existing ExistingRecord
	err := conn.QueryOne(ctx, &existing, checkQuery, []interface{}{tenantId, scriptName, statementHash, driver}, true)

	// 判断是否是真正的错误（排除"记录不存在"的情况）
	isRecordNotFound := err != nil && (strings.Contains(err.Error(), "no rows") ||
		strings.Contains(err.Error(), "record not found") ||
		strings.Contains(err.Error(), "not found"))
	isTableNotExist := err != nil && (strings.Contains(err.Error(), "no such table") ||
		strings.Contains(err.Error(), "doesn't exist") ||
		strings.Contains(err.Error(), "table") && strings.Contains(err.Error(), "not exist"))

	// 只有在非预期错误时才记录错误日志
	if err != nil && !isRecordNotFound && !isTableNotExist {
		logger.Error("检查语句执行历史记录失败",
			"error", err,
			"script", scriptName,
			"statement_hash", statementHash)
		return
	}

	if existing.StatementId != "" {
		// 记录存在，执行更新
		updateSQL := `UPDATE HUB_STATEMENT_EXECUTION_HISTORY 
					  SET executionStatus = ?, executionTime = ?, executionDuration = ?, 
					      errorMessage = ?, createdAt = ?
					  WHERE statementId = ?`

		_, err = conn.Exec(ctx, updateSQL, []interface{}{
			status, now, durationMs, errorMessage, now, existing.StatementId,
		}, false)

		if err != nil {
			logger.Error("更新语句执行历史失败",
				"error", err,
				"script", scriptName,
				"statement_hash", statementHash,
				"statement_id", existing.StatementId,
				"status", status)
		} else {
			logger.Debug("语句执行历史更新成功",
				"script", scriptName,
				"statement_hash", statementHash,
				"statement_id", existing.StatementId,
				"status", status,
				"duration_ms", durationMs)
		}
	} else {
		// 记录不存在，执行插入
		statementId := generateStatementId()
		insertSQL := `INSERT INTO HUB_STATEMENT_EXECUTION_HISTORY 
					  (statementId, tenantId, scriptName, statementHash, statementType, statementContent, 
					   databaseDriver, executionStatus, executionTime, executionDuration, errorMessage, createdAt)
					  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

		_, err = conn.Exec(ctx, insertSQL, []interface{}{
			statementId, tenantId, scriptName, statementHash, statementType, statementContent,
			driver, status, now, durationMs, errorMessage, now,
		}, false)

		if err != nil {
			logger.Error("插入语句执行历史失败",
				"error", err,
				"script", scriptName,
				"statement_hash", statementHash,
				"statement_id", statementId,
				"status", status)
		} else {
			logger.Debug("语句执行历史插入成功",
				"script", scriptName,
				"statement_hash", statementHash,
				"statement_id", statementId,
				"status", status,
				"duration_ms", durationMs)
		}
	}
}

// generateStatementId 生成语句执行记录ID
// 生成唯一的语句执行记录标识符，使用 random 包生成32位唯一字符串
// 返回:
//   - string: 32位唯一标识符，符合数据库VARCHAR(32)字段要求
func generateStatementId() string {
	return random.Generate32BitRandomString()
}

// generateExecutionId 生成执行记录ID
// 生成唯一的执行记录标识符，使用 random 包生成32位唯一字符串
// 返回:
//   - string: 32位唯一标识符，符合数据库VARCHAR(32)字段要求
func generateExecutionId() string {
	return random.Generate32BitRandomString()
}
