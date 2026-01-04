package db

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"gateway/cmd/common/utils"
	mongoscript "gateway/internal/script/mongo"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/logger"
	mongofactory "gateway/pkg/mongo/factory"
)

// ScriptExecutionResult 脚本执行结果
// 记录单个脚本文件的执行状态和结果信息
type ScriptExecutionResult struct {
	// ScriptFile 脚本文件路径
	ScriptFile string

	// DatabaseName 目标数据库连接名称
	DatabaseName string

	// Driver 数据库驱动类型
	Driver string

	// Success 执行是否成功
	Success bool

	// Error 执行错误信息（如果有）
	Error error

	// Duration 执行耗时
	Duration time.Duration

	// StatementsExecuted 成功执行的SQL语句数量
	StatementsExecuted int

	// StatementsFailed 失败的SQL语句数量
	StatementsFailed int

	// StatementsSkipped 跳过的SQL语句数量
	StatementsSkipped int

	// ScriptVersion 脚本版本（基于内容的MD5哈希）
	ScriptVersion string

	// Skipped 是否因为已执行而跳过
	Skipped bool
}

// ScriptExecutionHistory 脚本执行历史记录（文件级别）
// 用于跟踪脚本文件的整体执行状态
type ScriptExecutionHistory struct {
	// ExecutionId 执行记录ID
	ExecutionId string `db:"executionId" json:"executionId"`

	// TenantId 租户ID
	TenantId string `db:"tenantId" json:"tenantId"`

	// ScriptName 脚本名称（文件名）
	ScriptName string `db:"scriptName" json:"scriptName"`

	// ScriptPath 脚本完整路径
	ScriptPath string `db:"scriptPath" json:"scriptPath"`

	// ScriptVersion 脚本版本（MD5哈希）
	ScriptVersion string `db:"scriptVersion" json:"scriptVersion"`

	// DatabaseDriver 数据库驱动类型
	DatabaseDriver string `db:"databaseDriver" json:"databaseDriver"`

	// ExecutionStatus 执行状态（SUCCESS, FAILED, SKIPPED）
	ExecutionStatus string `db:"executionStatus" json:"executionStatus"`

	// ExecutionTime 执行时间
	ExecutionTime time.Time `db:"executionTime" json:"executionTime"`

	// ExecutionDuration 执行耗时（毫秒）
	ExecutionDuration int64 `db:"executionDuration" json:"executionDuration"`

	// StatementsExecuted 执行的SQL语句数量
	StatementsExecuted int `db:"statementsExecuted" json:"statementsExecuted"`

	// ErrorMessage 错误信息（如果有）
	ErrorMessage string `db:"errorMessage" json:"errorMessage"`

	// CreatedAt 记录创建时间
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

// StatementExecutionHistory SQL语句执行历史记录（语句级别）
// 用于跟踪每个SQL语句的执行状态，支持增量执行
type StatementExecutionHistory struct {
	// StatementId 语句执行记录ID
	StatementId string `db:"statementId" json:"statementId"`

	// TenantId 租户ID
	TenantId string `db:"tenantId" json:"tenantId"`

	// ScriptName 所属脚本名称
	ScriptName string `db:"scriptName" json:"scriptName"`

	// StatementHash 语句内容哈希（用于唯一标识）
	StatementHash string `db:"statementHash" json:"statementHash"`

	// StatementType 语句类型（CREATE_TABLE, CREATE_INDEX等）
	StatementType string `db:"statementType" json:"statementType"`

	// StatementContent SQL语句内容
	StatementContent string `db:"statementContent" json:"statementContent"`

	// DatabaseDriver 数据库驱动类型
	DatabaseDriver string `db:"databaseDriver" json:"databaseDriver"`

	// ExecutionStatus 执行状态（SUCCESS, FAILED, SKIPPED）
	ExecutionStatus string `db:"executionStatus" json:"executionStatus"`

	// ExecutionTime 执行时间
	ExecutionTime time.Time `db:"executionTime" json:"executionTime"`

	// ExecutionDuration 执行耗时（毫秒）
	ExecutionDuration int64 `db:"executionDuration" json:"executionDuration"`

	// ErrorMessage 错误信息（如果有）
	ErrorMessage string `db:"errorMessage" json:"errorMessage"`

	// CreatedAt 记录创建时间
	CreatedAt time.Time `db:"createdAt" json:"createdAt"`
}

// InitializationSummary 初始化总结报告
// 包含所有数据库脚本初始化的汇总信息
type InitializationSummary struct {
	// TotalDatabases 总数据库连接数
	TotalDatabases int

	// SuccessfulDatabases 成功初始化的数据库数
	SuccessfulDatabases int

	// FailedDatabases 初始化失败的数据库数
	FailedDatabases int

	// TotalScripts 总脚本文件数
	TotalScripts int

	// SuccessfulScripts 成功执行的脚本数
	SuccessfulScripts int

	// TotalDuration 总执行时间
	TotalDuration time.Duration

	// Results 详细执行结果列表
	Results []ScriptExecutionResult
}

// InitializeDatabaseScripts 初始化数据库脚本
// 参考mongo_manager_init.go的模式，提供简洁的初始化接口
// 参数:
//   - ctx: 上下文对象，用于控制执行超时和取消
//   - db: 数据库连接实例（主数据库，用于记录执行历史）
//
// 返回:
//   - *InitializationSummary: 初始化结果汇总报告
//   - error: 初始化失败时返回错误信息
func InitializeDatabaseScripts(ctx context.Context, db database.Database) (*InitializationSummary, error) {
	logger.Info("开始初始化数据库脚本")

	// 检查脚本初始化配置
	enableScriptInit := config.GetBool("database.enable_script_initialization", true)
	if !enableScriptInit {
		logger.Info("数据库脚本初始化已禁用")
		return &InitializationSummary{
			TotalDatabases:      0,
			SuccessfulDatabases: 0,
			FailedDatabases:     0,
			TotalScripts:        0,
			SuccessfulScripts:   0,
			TotalDuration:       0,
			Results:             []ScriptExecutionResult{},
		}, nil
	}

	// 获取脚本目录路径，考虑服务启动模式下的路径解析
	scriptDirConfig := config.GetString("database.script_directory", "scripts/db")
	scriptDir := utils.ResolvePath(scriptDirConfig)

	// 获取主数据库驱动类型
	driver := db.GetDriver()
	if driver == "" {
		return nil, fmt.Errorf("无法确定数据库驱动类型")
	}

	logger.Info("开始执行数据库脚本初始化",
		"driver", driver,
		"script_dir", scriptDirConfig,
		"resolved_path", scriptDir,
		"service_mode", utils.IsServiceMode())

	startTime := time.Now()

	// 确保脚本执行历史表存在（在主数据库中）
	err := ensureScriptHistoryTable(ctx, db, driver)
	if err != nil {
		logger.Error("创建脚本执行历史表失败", "error", err)
		return nil, fmt.Errorf("创建脚本执行历史表失败: %w", err)
	}

	// 结果列表
	var results []ScriptExecutionResult

	// 1. 执行主数据库脚本初始化
	mainResult := executeScriptForDatabase(ctx, "default", db, db, driver, scriptDir)
	results = append(results, mainResult)

	// 2. 检查是否配置了 ClickHouse，如果有则也执行 ClickHouse 脚本
	clickhouseConn := database.GetConnection("clickhouse_main")
	if clickhouseConn != nil {
		logger.Info("检测到 ClickHouse 配置，开始执行 ClickHouse 脚本初始化")
		clickhouseDriver := clickhouseConn.GetDriver()

		// 执行 ClickHouse 脚本，但历史记录保存在主数据库中
		ckResult := executeScriptForDatabase(ctx, "clickhouse_main", db, clickhouseConn, clickhouseDriver, scriptDir)
		results = append(results, ckResult)
	}

	// 3. 检查是否配置了 MongoDB，如果有则也执行 MongoDB 脚本
	if IsMongoEnabled() {
		logger.Info("检测到 MongoDB 配置，开始执行 MongoDB 脚本初始化")

		// 执行 MongoDB 脚本
		mongoResult := executeMongoScriptForDatabase(ctx, "mongodb_default", scriptDir)
		results = append(results, mongoResult)
	}

	// 创建汇总报告
	summary := &InitializationSummary{
		TotalDatabases:      len(results),
		SuccessfulDatabases: 0,
		FailedDatabases:     0,
		TotalScripts:        len(results),
		SuccessfulScripts:   0,
		TotalDuration:       time.Since(startTime),
		Results:             results,
	}

	// 统计成功和失败数量
	for _, result := range results {
		if result.Success || result.Skipped {
			summary.SuccessfulDatabases++
			summary.SuccessfulScripts++

			if result.Skipped {
				logger.Info("数据库脚本已跳过（已执行过相同版本）",
					"database", result.DatabaseName,
					"driver", result.Driver,
					"script_version", result.ScriptVersion,
					"duration", result.Duration)
			} else {
				if result.StatementsFailed > 0 {
					logger.Warn("数据库脚本初始化完成（部分语句失败）",
						"database", result.DatabaseName,
						"driver", result.Driver,
						"executed", result.StatementsExecuted,
						"failed", result.StatementsFailed,
						"skipped", result.StatementsSkipped,
						"duration", result.Duration)
				} else {
					logger.Info("数据库脚本初始化成功",
						"database", result.DatabaseName,
						"driver", result.Driver,
						"executed", result.StatementsExecuted,
						"skipped", result.StatementsSkipped,
						"duration", result.Duration)
				}
			}
		} else {
			summary.FailedDatabases++
			logger.Error("数据库脚本初始化失败",
				"database", result.DatabaseName,
				"driver", result.Driver,
				"error", result.Error,
				"executed", result.StatementsExecuted,
				"failed", result.StatementsFailed,
				"duration", result.Duration)
		}
	}

	return summary, nil
}

// executeScriptForDatabase 为指定数据库执行初始化脚本
// 内部方法，负责查找并执行对应数据库类型的初始化脚本
// 支持目录结构，执行目录下的所有脚本文件
// 参数:
//   - ctx: 上下文对象
//   - databaseName: 数据库连接名称（用于日志和记录）
//   - historyConn: 用于记录执行历史的数据库连接（通常是主数据库）
//   - targetConn: 实际执行脚本的目标数据库连接
//   - driver: 数据库驱动类型
//   - scriptDir: 脚本目录路径
//
// 返回:
//   - ScriptExecutionResult: 脚本执行结果（汇总所有脚本文件的结果）
func executeScriptForDatabase(ctx context.Context, databaseName string, historyConn database.Database, targetConn database.Database, driver string, scriptDir string) ScriptExecutionResult {
	startTime := time.Now()

	result := ScriptExecutionResult{
		DatabaseName: databaseName,
		Driver:       driver,
		Success:      false,
		Skipped:      false,
	}

	// 查找对应的脚本文件目录下的所有脚本文件
	scriptFiles, err := findScriptFiles(driver, scriptDir)
	if err != nil {
		result.Error = fmt.Errorf("查找脚本文件失败: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	logger.Info("找到脚本文件",
		"database", databaseName,
		"driver", driver,
		"script_count", len(scriptFiles),
		"script_dir", scriptDir)

	// 汇总所有脚本文件的执行结果
	totalExecuted := 0
	totalFailed := 0
	totalSkipped := 0
	var firstError error
	allScriptFiles := []string{}

	// 循环执行每个脚本文件
	for i, scriptFile := range scriptFiles {
		allScriptFiles = append(allScriptFiles, scriptFile)
		scriptName := filepath.Base(scriptFile)

		logger.Info("开始执行脚本文件",
			"database", databaseName,
			"driver", driver,
			"script_index", i+1,
			"total_scripts", len(scriptFiles),
			"script", scriptName)

		// 读取脚本内容
		scriptContent, err := os.ReadFile(scriptFile)
		if err != nil {
			logger.Error("读取脚本文件失败",
				"database", databaseName,
				"script", scriptName,
				"error", err)
			if firstError == nil {
				firstError = fmt.Errorf("读取脚本文件 %s 失败: %w", scriptName, err)
			}
			totalFailed++
			continue
		}

		// 计算脚本版本（MD5哈希）
		scriptVersion := calculateScriptVersion(scriptContent)

		// 根据数据库类型执行脚本
		switch driver {
		case dbtypes.DriverMySQL, dbtypes.DriverSQLite, dbtypes.DriverOracle, dbtypes.DriverClickHouse:
			// SQL类型数据库 - 按语句级别执行
			// 注意：使用 historyConn 查询执行历史，使用 targetConn 执行SQL
			executedCount, failedCount, skippedCount, err := executeSQLScriptByStatements(ctx, historyConn, targetConn, driver, scriptName, string(scriptContent))

			totalExecuted += executedCount
			totalFailed += failedCount
			totalSkipped += skippedCount

			if err != nil {
				logger.Error("执行SQL脚本失败",
					"database", databaseName,
					"script", scriptName,
					"error", err)
				if firstError == nil {
					firstError = fmt.Errorf("执行脚本 %s 失败: %w", scriptName, err)
				}
				// 记录脚本整体执行失败的历史（记录到主数据库）
				recordScriptExecution(ctx, historyConn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
					time.Since(startTime), executedCount, err.Error())
			} else {
				status := "SUCCESS"
				errorMsg := ""

				// 如果有失败的语句，标记为部分成功
				if failedCount > 0 {
					status = "PARTIAL_SUCCESS"
					errorMsg = fmt.Sprintf("%d条语句执行失败", failedCount)
				}

				// 记录脚本整体执行历史（记录到主数据库）
				recordScriptExecution(ctx, historyConn, driver, scriptName, scriptFile, scriptVersion, status,
					time.Since(startTime), executedCount, errorMsg)

				logger.Info("脚本文件执行完成",
					"database", databaseName,
					"script", scriptName,
					"executed", executedCount,
					"failed", failedCount,
					"skipped", skippedCount)
			}

		case dbtypes.DriverMongoDB:
			// MongoDB JavaScript脚本
			err := fmt.Errorf("MongoDB脚本执行暂未实现")
			logger.Error("MongoDB脚本执行暂未实现",
				"database", databaseName,
				"script", scriptName)
			if firstError == nil {
				firstError = err
			}
			recordScriptExecution(ctx, historyConn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
				time.Since(startTime), 0, err.Error())

		default:
			err := fmt.Errorf("不支持的数据库驱动类型: %s", driver)
			logger.Error("不支持的数据库驱动类型",
				"database", databaseName,
				"driver", driver,
				"script", scriptName)
			if firstError == nil {
				firstError = err
			}
			recordScriptExecution(ctx, historyConn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
				time.Since(startTime), 0, err.Error())
		}
	}

	// 汇总结果
	result.ScriptFile = strings.Join(allScriptFiles, "; ")
	result.StatementsExecuted = totalExecuted
	result.StatementsFailed = totalFailed
	result.StatementsSkipped = totalSkipped
	result.Duration = time.Since(startTime)

	// 如果所有脚本都执行成功（即使有部分语句失败），则认为整体成功
	if firstError == nil || totalExecuted > 0 {
		result.Success = true
		if totalFailed > 0 {
			logger.Warn("数据库脚本执行完成（部分语句失败）",
				"database", databaseName,
				"total_scripts", len(scriptFiles),
				"executed", totalExecuted,
				"failed", totalFailed,
				"skipped", totalSkipped,
				"duration", result.Duration)
		} else {
			logger.Info("数据库脚本执行成功",
				"database", databaseName,
				"total_scripts", len(scriptFiles),
				"executed", totalExecuted,
				"skipped", totalSkipped,
				"duration", result.Duration)
		}
	} else {
		result.Error = firstError
		result.Success = false
		logger.Error("数据库脚本执行失败",
			"database", databaseName,
			"total_scripts", len(scriptFiles),
			"error", firstError,
			"executed", totalExecuted,
			"failed", totalFailed,
			"duration", result.Duration)
	}

	return result
}

// executeSQLScriptByStatements 按语句级别执行SQL脚本
// 解析SQL脚本，检查每个语句是否已执行，只执行未执行的语句，支持增量执行
// 单条语句失败不会中断整个初始化流程，会继续执行后续语句
// 参数:
//   - ctx: 上下文对象
//   - historyConn: 用于查询和记录执行历史的数据库连接（主数据库）
//   - targetConn: 实际执行SQL的目标数据库连接
//   - driver: 数据库驱动类型
//   - scriptName: 脚本名称
//   - scriptContent: 脚本内容
//
// 返回值: (成功执行数, 失败数, 跳过数, error)
func executeSQLScriptByStatements(ctx context.Context, historyConn database.Database, targetConn database.Database, driver, scriptName, scriptContent string) (int, int, int, error) {
	// 分割SQL语句
	statements := splitSQLStatements(scriptContent)

	logger.Info("SQL脚本分析完成", "script", scriptName, "total_statements", len(statements))

	executedCount := 0
	skippedCount := 0
	failedCount := 0
	var failedStatements []string

	// 逐条检查并执行SQL语句
	for i, stmt := range statements {
		// 跳过空语句和注释
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") || strings.HasPrefix(stmt, "/*") {
			continue
		}

		// 计算语句哈希值
		stmtHash := calculateStatementHash(stmt)
		stmtType := getSQLStatementType(stmt)

		// 检查语句执行状态（从主数据库查询历史）
		executionStatus, err := getStatementExecutionStatus(ctx, historyConn, driver, scriptName, stmtHash)
		if err != nil {
			logger.Warn("检查语句执行历史失败，继续执行",
				"statement_index", i+1,
				"statement_hash", stmtHash,
				"error", err)
		} else {
			switch executionStatus {
			case "SUCCESS":
				logger.Info("语句已成功执行过，跳过",
					"script", scriptName,
					"statement_index", i+1,
					"statement_type", stmtType,
					"statement_hash", stmtHash,
					"statement_preview", truncateString(stmt, 100))
				skippedCount++
				continue
			case "FAILED", "SKIPPED":
				logger.Info("语句之前执行失败或跳过，重新执行",
					"script", scriptName,
					"statement_index", i+1,
					"statement_type", stmtType,
					"statement_hash", stmtHash,
					"previous_status", executionStatus)
			case "":
				logger.Debug("语句未执行过，准备执行",
					"script", scriptName,
					"statement_index", i+1,
					"statement_type", stmtType,
					"statement_hash", stmtHash)
			}
		}

		// 记录即将执行的语句（用于调试）
		logger.Debug("准备执行SQL语句",
			"statement_index", i+1,
			"statement_type", stmtType,
			"statement_hash", stmtHash,
			"statement_preview", truncateString(stmt, 200))

		// 执行SQL语句（在目标数据库上执行）
		startTime := time.Now()
		_, err = targetConn.Exec(ctx, stmt, nil, false)
		duration := time.Since(startTime)

		if err != nil {
			// 记录执行失败的语句信息
			logger.Warn("SQL语句执行失败，继续执行后续语句",
				"statement_index", i+1,
				"statement_type", stmtType,
				"statement_hash", stmtHash,
				"statement_preview", truncateString(stmt, 200),
				"error", err)

			// 记录语句执行失败的历史（记录到主数据库）
			recordStatementExecution(ctx, historyConn, driver, scriptName, stmtHash, stmtType, stmt, "FAILED", duration, err.Error())

			// 记录失败的语句信息，但不中断执行
			failedCount++
			failedStatements = append(failedStatements, fmt.Sprintf("第%d条: %s (错误: %v)", i+1, truncateString(stmt, 100), err))

			// 继续执行下一条语句
			continue
		}

		// 记录语句执行成功的历史（记录到主数据库）
		recordStatementExecution(ctx, historyConn, driver, scriptName, stmtHash, stmtType, stmt, "SUCCESS", duration, "")

		executedCount++

		// 每执行10条语句记录一次进度（降低频率以便调试）
		if executedCount%10 == 0 {
			logger.Info("SQL脚本执行进度", "script", scriptName, "executed", executedCount, "skipped", skippedCount, "failed", failedCount, "total", len(statements))
		}
	}

	// 汇总执行结果
	logger.Info("SQL脚本执行完成",
		"script", scriptName,
		"total_executed", executedCount,
		"total_skipped", skippedCount,
		"total_failed", failedCount,
		"total_statements", len(statements))

	// 如果有失败的语句，记录详细信息但不返回错误
	if failedCount > 0 {
		logger.Warn("部分SQL语句执行失败",
			"script", scriptName,
			"failed_count", failedCount,
			"failed_statements", failedStatements)
	}

	return executedCount, failedCount, skippedCount, nil
}

// splitSQLStatements 分割SQL脚本为独立的语句
// 按分号分割SQL语句，处理多行语句和注释，确保正确的执行顺序
func splitSQLStatements(scriptContent string) []string {
	// 按分号分割，但需要处理字符串内的分号
	lines := strings.Split(scriptContent, "\n")
	var statements []string
	var currentStatement strings.Builder
	inString := false
	stringChar := byte(0)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "--") || strings.HasPrefix(line, "/*") {
			continue
		}

		// 处理每个字符
		for i := 0; i < len(line); i++ {
			char := line[i]

			// 处理字符串边界
			if !inString && (char == '\'' || char == '"') {
				inString = true
				stringChar = char
			} else if inString && char == stringChar {
				// 检查是否是转义字符
				if i == 0 || line[i-1] != '\\' {
					inString = false
					stringChar = 0
				}
			}

			currentStatement.WriteByte(char)

			// 如果遇到分号且不在字符串内，则结束当前语句
			if char == ';' && !inString {
				stmt := strings.TrimSpace(currentStatement.String())
				if stmt != "" && stmt != ";" {
					statements = append(statements, stmt)
				}
				currentStatement.Reset()
				break
			}
		}

		// 如果行没有以分号结尾，添加换行符
		if len(line) > 0 && line[len(line)-1] != ';' && currentStatement.Len() > 0 {
			currentStatement.WriteString("\n")
		}
	}

	// 处理最后一个语句（如果没有分号结尾）
	if currentStatement.Len() > 0 {
		stmt := strings.TrimSpace(currentStatement.String())
		if stmt != "" {
			statements = append(statements, stmt)
		}
	}

	return statements
}

// findScriptFiles 查找指定数据库驱动对应的脚本文件目录下的所有脚本文件
// 在脚本目录的子目录中查找匹配的初始化脚本文件，按文件名排序返回
func findScriptFiles(driver string, scriptDir string) ([]string, error) {
	// 支持的数据库驱动类型及其对应的目录名和文件扩展名
	driverToDir := map[string]string{
		dbtypes.DriverMySQL:      "mysql",
		dbtypes.DriverSQLite:     "sqlite",
		dbtypes.DriverOracle:     "oracle",
		dbtypes.DriverClickHouse: "clickhouse",
		dbtypes.DriverMongoDB:    "mongo",
	}

	driverToExt := map[string]string{
		dbtypes.DriverMySQL:      ".sql",
		dbtypes.DriverSQLite:     ".sql",
		dbtypes.DriverOracle:     ".sql",
		dbtypes.DriverClickHouse: ".sql",
		dbtypes.DriverMongoDB:    ".js",
	}

	// 获取对应的目录名和文件扩展名
	subDir, exists := driverToDir[driver]
	if !exists {
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	ext, exists := driverToExt[driver]
	if !exists {
		return nil, fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	// 构建脚本文件目录路径
	scriptSubDir := filepath.Join(scriptDir, subDir)

	// 检查目录是否存在
	if _, err := os.Stat(scriptSubDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("脚本目录不存在: %s", scriptSubDir)
	}

	// 读取目录下的所有文件（只读取当前目录，不递归）
	entries, err := os.ReadDir(scriptSubDir)
	if err != nil {
		return nil, fmt.Errorf("读取脚本目录失败: %w", err)
	}

	var scriptFiles []string
	for _, entry := range entries {
		// 只处理文件，跳过目录
		if entry.IsDir() {
			continue
		}

		// 检查文件扩展名
		if filepath.Ext(entry.Name()) == ext {
			// 排除 init.sql 文件（因为它只包含 source 命令，不适合程序执行）
			if entry.Name() != "init.sql" {
				scriptPath := filepath.Join(scriptSubDir, entry.Name())
				scriptFiles = append(scriptFiles, scriptPath)
			}
		}
	}

	// 如果没有找到任何脚本文件，返回错误
	if len(scriptFiles) == 0 {
		return nil, fmt.Errorf("未找到数据库 %s 的初始化脚本文件，查找路径: %s", driver, scriptSubDir)
	}

	// 按文件名排序，确保执行顺序一致
	sort.Strings(scriptFiles)

	return scriptFiles, nil
}

// truncateString 截断字符串到指定长度
// 辅助方法，用于在日志中显示SQL语句预览
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// getSQLStatementType 获取SQL语句类型
// 辅助方法，用于在日志中显示SQL语句类型
func getSQLStatementType(stmt string) string {
	stmt = strings.TrimSpace(strings.ToUpper(stmt))

	if strings.HasPrefix(stmt, "CREATE TABLE") {
		return "CREATE_TABLE"
	} else if strings.HasPrefix(stmt, "CREATE INDEX") {
		return "CREATE_INDEX"
	} else if strings.HasPrefix(stmt, "CREATE UNIQUE INDEX") {
		return "CREATE_UNIQUE_INDEX"
	} else if strings.HasPrefix(stmt, "INSERT INTO") {
		return "INSERT"
	} else if strings.HasPrefix(stmt, "PRAGMA") {
		return "PRAGMA"
	} else if strings.HasPrefix(stmt, "SELECT") {
		return "SELECT"
	} else if strings.HasPrefix(stmt, "ANALYZE") {
		return "ANALYZE"
	} else if strings.HasPrefix(stmt, "CREATE") {
		return "CREATE_OTHER"
	} else if strings.HasPrefix(stmt, "ALTER") {
		return "ALTER"
	} else if strings.HasPrefix(stmt, "DROP") {
		return "DROP"
	} else {
		return "UNKNOWN"
	}
}

// ListAvailableScripts 列出可用的数据库脚本文件
// 扫描脚本目录，返回所有可用的数据库初始化脚本
func ListAvailableScripts(scriptPath string) (map[string]string, error) {
	if scriptPath == "" {
		// 获取配置的脚本目录，并使用utils.ResolvePath处理路径
		scriptDirConfig := config.GetString("database.script_directory", "scripts/db")
		scriptPath = utils.ResolvePath(scriptDirConfig)
	}

	scripts := make(map[string]string)

	// 扫描脚本目录
	err := filepath.WalkDir(scriptPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// 检查文件扩展名和名称
		filename := d.Name()
		ext := filepath.Ext(filename)
		basename := strings.TrimSuffix(filename, ext)

		// 匹配支持的数据库类型
		supportedDrivers := map[string]string{
			"mysql":      dbtypes.DriverMySQL,
			"sqlite":     dbtypes.DriverSQLite,
			"oracle":     dbtypes.DriverOracle,
			"clickhouse": dbtypes.DriverClickHouse,
			"mongo":      dbtypes.DriverMongoDB,
		}

		for pattern, driver := range supportedDrivers {
			if strings.EqualFold(basename, pattern) {
				scripts[driver] = path
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("扫描脚本目录失败: %w", err)
	}

	return scripts, nil
}

// CheckScriptInitializationConfig 检查脚本初始化配置
// 验证配置文件中的脚本初始化相关配置是否正确
func CheckScriptInitializationConfig() (bool, bool, int, string) {
	enableScriptInit := config.GetBool("database.enable_script_initialization", true)
	allowPartialFailure := config.GetBool("database.allow_partial_failure", false)
	timeoutMinutes := config.GetInt("database.script_initialization_timeout", 30)
	scriptDir := config.GetString("database.script_directory", "scripts/db")

	return enableScriptInit, allowPartialFailure, timeoutMinutes, scriptDir
}

// calculateScriptVersion 计算脚本版本
// 使用MD5哈希计算脚本内容的版本标识，用于检测脚本内容是否发生变化
// 参数:
//   - scriptContent: 脚本文件的二进制内容
//
// 返回:
//   - string: 脚本内容的MD5哈希值（32位十六进制字符串）
func calculateScriptVersion(scriptContent []byte) string {
	hash := md5.Sum(scriptContent)
	return fmt.Sprintf("%x", hash)
}

// calculateStatementHash 计算SQL语句哈希值
// 使用MD5哈希计算SQL语句内容的唯一标识，用于跟踪语句执行状态
// 参数:
//   - statement: SQL语句内容字符串
//
// 返回:
//   - string: 语句内容的MD5哈希值（32位十六进制字符串）
func calculateStatementHash(statement string) string {
	// 先标准化语句：去除首尾空格，统一换行符
	normalized := strings.TrimSpace(strings.ReplaceAll(statement, "\r\n", "\n"))
	hash := md5.Sum([]byte(normalized))
	return fmt.Sprintf("%x", hash)
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
	tenantId := config.GetString("database.tenant_id", "default")

	var query string
	var args []interface{}

	if scriptName != "" {
		query = `SELECT executionId, tenantId, scriptName, scriptPath, scriptVersion, databaseDriver,
				  executionStatus, executionTime, executionDuration, statementsExecuted, errorMessage, createdAt
				  FROM HUB_SCRIPT_EXECUTION_HISTORY 
				  WHERE tenantId = ? AND scriptName = ?
				  ORDER BY executionTime DESC`
		args = []interface{}{tenantId, scriptName}
	} else {
		query = `SELECT executionId, tenantId, scriptName, scriptPath, scriptVersion, databaseDriver,
				  executionStatus, executionTime, executionDuration, statementsExecuted, errorMessage, createdAt
				  FROM HUB_SCRIPT_EXECUTION_HISTORY 
				  WHERE tenantId = ?
				  ORDER BY executionTime DESC`
		args = []interface{}{tenantId}
	}

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var histories []ScriptExecutionHistory
	err := conn.Query(ctx, &histories, query, args, true)
	if err != nil {
		return nil, fmt.Errorf("查询脚本执行历史失败: %w", err)
	}

	return histories, nil
}

// ForceExecuteScript 强制执行脚本
// 忽略版本检查，强制执行指定的脚本，主要用于手动维护和测试场景
// 参数:
//   - ctx: 上下文对象，用于控制执行超时和取消
//   - db: 数据库连接实例
//   - scriptPath: 脚本文件的完整路径
//
// 返回:
//   - *ScriptExecutionResult: 脚本执行结果
//   - error: 执行失败时返回错误信息
func ForceExecuteScript(ctx context.Context, db database.Database, scriptPath string) (*ScriptExecutionResult, error) {
	driver := db.GetDriver()
	if driver == "" {
		return nil, fmt.Errorf("无法确定数据库驱动类型")
	}

	// 确保脚本执行历史表存在
	err := ensureScriptHistoryTable(ctx, db, driver)
	if err != nil {
		return nil, fmt.Errorf("创建脚本执行历史表失败: %w", err)
	}

	scriptDir := filepath.Dir(scriptPath)
	// 对于手动执行，历史和执行都使用同一个数据库连接
	result := executeScriptForDatabase(ctx, "manual", db, db, driver, scriptDir)

	return &result, nil
}

// IsMongoEnabled 检查 MongoDB 是否已启用
// 返回:
//   - bool: true 表示 MongoDB 已配置且启用
func IsMongoEnabled() bool {
	_, err := mongofactory.GetDefaultConnection()
	return err == nil
}

// executeMongoScriptForDatabase 为 MongoDB 数据库执行初始化脚本
// 参数:
//   - ctx: 上下文对象
//   - databaseName: 数据库连接名称（用于日志）
//   - scriptDir: 脚本目录路径
//
// 返回:
//   - ScriptExecutionResult: 脚本执行结果
func executeMongoScriptForDatabase(ctx context.Context, databaseName string, scriptDir string) ScriptExecutionResult {
	startTime := time.Now()

	result := ScriptExecutionResult{
		DatabaseName: databaseName,
		Driver:       dbtypes.DriverMongoDB,
		Success:      false,
		Skipped:      false,
	}

	// 执行 MongoDB 脚本
	mongoResult, err := mongoscript.ExecuteMongoScript(ctx, scriptDir)
	if err != nil {
		result.Error = err
		result.Duration = time.Since(startTime)
		return result
	}

	// 转换结果
	result.ScriptFile = mongoResult.ScriptFile
	result.Success = mongoResult.Success
	result.StatementsExecuted = mongoResult.CommandsExecuted
	result.StatementsFailed = mongoResult.CommandsFailed
	result.Duration = mongoResult.Duration

	if mongoResult.Success {
		logger.Info("MongoDB 脚本执行成功",
			"database", databaseName,
			"executed", mongoResult.CommandsExecuted,
			"failed", mongoResult.CommandsFailed,
			"duration", mongoResult.Duration)
	} else {
		logger.Error("MongoDB 脚本执行失败",
			"database", databaseName,
			"error", mongoResult.Error,
			"executed", mongoResult.CommandsExecuted,
			"failed", mongoResult.CommandsFailed,
			"duration", mongoResult.Duration)
	}

	return result
}
