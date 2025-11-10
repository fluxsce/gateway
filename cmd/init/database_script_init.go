package init

import (
	"context"
	"crypto/md5"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gateway/cmd/common/utils"
	"gateway/pkg/config"
	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/logger"
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
//   - db: 数据库连接实例
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

	// 获取数据库驱动类型
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

	// 确保脚本执行历史表存在
	err := ensureScriptHistoryTable(ctx, db, driver)
	if err != nil {
		logger.Error("创建脚本执行历史表失败", "error", err)
		return nil, fmt.Errorf("创建脚本执行历史表失败: %w", err)
	}

	// 执行脚本初始化
	result := executeScriptForDatabase(ctx, "default", db, driver, scriptDir)

	// 创建汇总报告
	summary := &InitializationSummary{
		TotalDatabases:      1,
		SuccessfulDatabases: 0,
		FailedDatabases:     0,
		TotalScripts:        1,
		SuccessfulScripts:   0,
		TotalDuration:       time.Since(startTime),
		Results:             []ScriptExecutionResult{result},
	}

	if result.Success || result.Skipped {
		summary.SuccessfulDatabases = 1
		summary.SuccessfulScripts = 1
		if result.Skipped {
			logger.Info("数据库脚本已跳过（已执行过相同版本）",
				"driver", driver,
				"script_version", result.ScriptVersion,
				"duration", result.Duration)
		} else {
			if result.StatementsFailed > 0 {
				logger.Warn("数据库脚本初始化完成（部分语句失败）",
					"driver", driver,
					"executed", result.StatementsExecuted,
					"failed", result.StatementsFailed,
					"skipped", result.StatementsSkipped,
					"duration", result.Duration)
			} else {
				logger.Info("数据库脚本初始化成功",
					"driver", driver,
					"executed", result.StatementsExecuted,
					"skipped", result.StatementsSkipped,
					"duration", result.Duration)
			}
		}
	} else {
		summary.FailedDatabases = 1
		logger.Error("数据库脚本初始化失败",
			"driver", driver,
			"error", result.Error,
			"executed", result.StatementsExecuted,
			"failed", result.StatementsFailed,
			"duration", result.Duration)
		return summary, nil
	}

	return summary, nil
}

// executeScriptForDatabase 为指定数据库执行初始化脚本
// 内部方法，负责查找并执行对应数据库类型的初始化脚本
func executeScriptForDatabase(ctx context.Context, databaseName string, conn database.Database, driver string, scriptDir string) ScriptExecutionResult {
	startTime := time.Now()

	result := ScriptExecutionResult{
		DatabaseName: databaseName,
		Driver:       driver,
		Success:      false,
		Skipped:      false,
	}

	// 查找对应的脚本文件
	scriptFile, err := findScriptFile(driver, scriptDir)
	if err != nil {
		result.Error = fmt.Errorf("查找脚本文件失败: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.ScriptFile = scriptFile

	// 读取脚本内容
	scriptContent, err := os.ReadFile(scriptFile)
	if err != nil {
		result.Error = fmt.Errorf("读取脚本文件失败: %w", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// 计算脚本版本（MD5哈希）
	scriptVersion := calculateScriptVersion(scriptContent)
	result.ScriptVersion = scriptVersion

	// 注意：现在我们不再检查整个脚本是否执行过，而是在语句级别进行检查
	// 这样允许增量执行新添加的语句，而不需要重新执行整个脚本

	logger.Info("开始执行数据库脚本",
		"database", databaseName,
		"driver", driver,
		"script", scriptFile,
		"version", scriptVersion)

	// 根据数据库类型执行脚本
	scriptName := filepath.Base(scriptFile)
	switch driver {
	case dbtypes.DriverMySQL, dbtypes.DriverSQLite, dbtypes.DriverOracle, dbtypes.DriverClickHouse:
		// SQL类型数据库 - 按语句级别执行
		executedCount, failedCount, skippedCount, err := executeSQLScriptByStatements(ctx, conn, driver, scriptName, string(scriptContent))
		result.StatementsExecuted = executedCount
		result.StatementsFailed = failedCount
		result.StatementsSkipped = skippedCount

		if err != nil {
			result.Error = fmt.Errorf("执行SQL脚本失败: %w", err)
			// 记录脚本整体执行失败的历史
			recordScriptExecution(ctx, conn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
				result.Duration, executedCount, err.Error())
		} else {
			// 只要没有致命错误，就认为执行成功（即使有部分语句失败）
			result.Success = true
			status := "SUCCESS"
			errorMsg := ""

			// 如果有失败的语句，标记为部分成功
			if failedCount > 0 {
				status = "PARTIAL_SUCCESS"
				errorMsg = fmt.Sprintf("%d条语句执行失败", failedCount)
			}

			// 记录脚本整体执行历史
			recordScriptExecution(ctx, conn, driver, scriptName, scriptFile, scriptVersion, status,
				result.Duration, executedCount, errorMsg)
		}

	case dbtypes.DriverMongoDB:
		// MongoDB JavaScript脚本
		result.Error = fmt.Errorf("MongoDB脚本执行暂未实现")
		recordScriptExecution(ctx, conn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
			result.Duration, 0, result.Error.Error())

	default:
		result.Error = fmt.Errorf("不支持的数据库驱动类型: %s", driver)
		recordScriptExecution(ctx, conn, driver, scriptName, scriptFile, scriptVersion, "FAILED",
			result.Duration, 0, result.Error.Error())
	}

	result.Duration = time.Since(startTime)

	if result.Success {
		if result.StatementsFailed > 0 {
			logger.Warn("数据库脚本执行完成（部分语句失败）",
				"database", databaseName,
				"executed", result.StatementsExecuted,
				"failed", result.StatementsFailed,
				"skipped", result.StatementsSkipped,
				"duration", result.Duration,
				"version", scriptVersion)
		} else {
			logger.Info("数据库脚本执行成功",
				"database", databaseName,
				"executed", result.StatementsExecuted,
				"skipped", result.StatementsSkipped,
				"duration", result.Duration,
				"version", scriptVersion)
		}
	} else {
		logger.Error("数据库脚本执行失败",
			"database", databaseName,
			"error", result.Error,
			"executed", result.StatementsExecuted,
			"failed", result.StatementsFailed,
			"duration", result.Duration,
			"version", scriptVersion)
	}

	return result
}

// executeSQLScriptByStatements 按语句级别执行SQL脚本
// 解析SQL脚本，检查每个语句是否已执行，只执行未执行的语句，支持增量执行
// 单条语句失败不会中断整个初始化流程，会继续执行后续语句
// 返回值: (成功执行数, 失败数, 跳过数, error)
func executeSQLScriptByStatements(ctx context.Context, conn database.Database, driver, scriptName, scriptContent string) (int, int, int, error) {
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

		// 检查语句执行状态
		executionStatus, err := getStatementExecutionStatus(ctx, conn, driver, scriptName, stmtHash)
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

		// 执行SQL语句
		startTime := time.Now()
		_, err = conn.Exec(ctx, stmt, nil, false)
		duration := time.Since(startTime)

		if err != nil {
			// 记录执行失败的语句信息
			logger.Warn("SQL语句执行失败，继续执行后续语句",
				"statement_index", i+1,
				"statement_type", stmtType,
				"statement_hash", stmtHash,
				"statement_preview", truncateString(stmt, 200),
				"error", err)

			// 记录语句执行失败的历史（自动判断插入或更新）
			recordStatementExecution(ctx, conn, driver, scriptName, stmtHash, stmtType, stmt, "FAILED", duration, err.Error())

			// 记录失败的语句信息，但不中断执行
			failedCount++
			failedStatements = append(failedStatements, fmt.Sprintf("第%d条: %s (错误: %v)", i+1, truncateString(stmt, 100), err))

			// 继续执行下一条语句
			continue
		}

		// 记录语句执行成功的历史（自动判断插入或更新）
		recordStatementExecution(ctx, conn, driver, scriptName, stmtHash, stmtType, stmt, "SUCCESS", duration, "")

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

// findScriptFile 查找指定数据库驱动对应的脚本文件
// 在脚本目录中查找匹配的初始化脚本文件
func findScriptFile(driver string, scriptDir string) (string, error) {
	// 支持的数据库驱动类型及其对应的脚本文件扩展名
	supportedDrivers := map[string]string{
		dbtypes.DriverMySQL:      ".sql",
		dbtypes.DriverSQLite:     ".sql",
		dbtypes.DriverOracle:     ".sql",
		dbtypes.DriverClickHouse: ".sql",
		dbtypes.DriverMongoDB:    ".js",
	}

	// 获取对应的文件扩展名
	ext, exists := supportedDrivers[driver]
	if !exists {
		return "", fmt.Errorf("不支持的数据库驱动: %s", driver)
	}

	// 构建可能的脚本文件名
	possibleNames := []string{
		driver + ext,
		strings.ToLower(driver) + ext,
		strings.ToUpper(driver) + ext,
	}

	// 在脚本目录中查找文件
	for _, name := range possibleNames {
		scriptPath := filepath.Join(scriptDir, name)
		if _, err := os.Stat(scriptPath); err == nil {
			return scriptPath, nil
		}
	}

	return "", fmt.Errorf("未找到数据库 %s 的初始化脚本文件，查找路径: %s", driver, scriptDir)
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

// ensureScriptHistoryTable 确保脚本执行历史表存在
// 创建用于跟踪脚本执行历史的表（包括文件级别和语句级别），支持多种数据库类型
// 该表用于记录每次脚本执行的详细信息，包括版本、状态、耗时等
// 参数:
//   - ctx: 上下文对象，用于控制创建表的超时和取消
//   - conn: 数据库连接实例
//   - driver: 数据库驱动类型（mysql, sqlite, oracle, clickhouse）
//
// 返回:
//   - error: 创建表失败时返回错误信息
func ensureScriptHistoryTable(ctx context.Context, conn database.Database, driver string) error {
	// 首先创建脚本执行历史表（文件级别）
	err := createScriptHistoryTable(ctx, conn, driver)
	if err != nil {
		return fmt.Errorf("创建脚本执行历史表失败: %w", err)
	}

	// 然后创建语句执行历史表（语句级别）
	err = createStatementHistoryTable(ctx, conn, driver)
	if err != nil {
		return fmt.Errorf("创建语句执行历史表失败: %w", err)
	}

	return nil
}

// createScriptHistoryTable 创建脚本执行历史表（文件级别）
func createScriptHistoryTable(ctx context.Context, conn database.Database, driver string) error {
	var createTableSQL string

	switch driver {
	case dbtypes.DriverMySQL:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_SCRIPT_EXECUTION_HISTORY (
    executionId VARCHAR(32) NOT NULL PRIMARY KEY,
    tenantId VARCHAR(32) NOT NULL DEFAULT 'default',
    scriptName VARCHAR(255) NOT NULL,
    scriptPath VARCHAR(500) NOT NULL,
    scriptVersion VARCHAR(32) NOT NULL,
    databaseDriver VARCHAR(50) NOT NULL,
    executionStatus VARCHAR(20) NOT NULL,
    executionTime DATETIME NOT NULL,
    executionDuration BIGINT NOT NULL DEFAULT 0,
    statementsExecuted INT NOT NULL DEFAULT 0,
    errorMessage TEXT,
    createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX IDX_SCRIPT_HIST_NAME (scriptName),
    INDEX IDX_SCRIPT_HIST_VERSION (scriptVersion),
    INDEX IDX_SCRIPT_HIST_STATUS (executionStatus),
    INDEX IDX_SCRIPT_HIST_DRIVER (databaseDriver),
    INDEX IDX_SCRIPT_HIST_TIME (executionTime),
    UNIQUE KEY UK_SCRIPT_VERSION (tenantId, scriptName, scriptVersion, databaseDriver)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='脚本执行历史表';`

	case dbtypes.DriverSQLite:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_SCRIPT_EXECUTION_HISTORY (
    executionId TEXT NOT NULL PRIMARY KEY,
    tenantId TEXT NOT NULL DEFAULT 'default',
    scriptName TEXT NOT NULL,
    scriptPath TEXT NOT NULL,
    scriptVersion TEXT NOT NULL,
    databaseDriver TEXT NOT NULL,
    executionStatus TEXT NOT NULL,
    executionTime DATETIME NOT NULL,
    executionDuration INTEGER NOT NULL DEFAULT 0,
    statementsExecuted INTEGER NOT NULL DEFAULT 0,
    errorMessage TEXT,
    createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS IDX_SCRIPT_HIST_NAME ON HUB_SCRIPT_EXECUTION_HISTORY(scriptName);
CREATE INDEX IF NOT EXISTS IDX_SCRIPT_HIST_VERSION ON HUB_SCRIPT_EXECUTION_HISTORY(scriptVersion);
CREATE INDEX IF NOT EXISTS IDX_SCRIPT_HIST_STATUS ON HUB_SCRIPT_EXECUTION_HISTORY(executionStatus);
CREATE INDEX IF NOT EXISTS IDX_SCRIPT_HIST_DRIVER ON HUB_SCRIPT_EXECUTION_HISTORY(databaseDriver);
CREATE INDEX IF NOT EXISTS IDX_SCRIPT_HIST_TIME ON HUB_SCRIPT_EXECUTION_HISTORY(executionTime);
CREATE UNIQUE INDEX IF NOT EXISTS UK_SCRIPT_VERSION ON HUB_SCRIPT_EXECUTION_HISTORY(tenantId, scriptName, scriptVersion, databaseDriver);`

	case dbtypes.DriverOracle:
		createTableSQL = `
BEGIN
    EXECUTE IMMEDIATE 'CREATE TABLE HUB_SCRIPT_EXECUTION_HISTORY (
        executionId VARCHAR2(32) NOT NULL PRIMARY KEY,
        tenantId VARCHAR2(32) DEFAULT ''default'' NOT NULL,
        scriptName VARCHAR2(255) NOT NULL,
        scriptPath VARCHAR2(500) NOT NULL,
        scriptVersion VARCHAR2(32) NOT NULL,
        databaseDriver VARCHAR2(50) NOT NULL,
        executionStatus VARCHAR2(20) NOT NULL,
        executionTime DATE NOT NULL,
        executionDuration NUMBER(19) DEFAULT 0 NOT NULL,
        statementsExecuted NUMBER(10) DEFAULT 0 NOT NULL,
        errorMessage CLOB,
        createdAt DATE DEFAULT SYSDATE NOT NULL
    )';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -955 THEN
            RAISE;
        END IF;
END;`

	case dbtypes.DriverClickHouse:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_SCRIPT_EXECUTION_HISTORY (
    executionId String,
    tenantId String DEFAULT 'default',
    scriptName String,
    scriptPath String,
    scriptVersion String,
    databaseDriver String,
    executionStatus String,
    executionTime DateTime,
    executionDuration Int64 DEFAULT 0,
    statementsExecuted Int32 DEFAULT 0,
    errorMessage String,
    createdAt DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (tenantId, scriptName, executionTime)
SETTINGS index_granularity = 8192;`

	default:
		return fmt.Errorf("不支持的数据库驱动类型: %s", driver)
	}

	logger.Debug("创建脚本执行历史表", "driver", driver)
	_, err := conn.Exec(ctx, createTableSQL, nil, false)
	if err != nil {
		return fmt.Errorf("创建脚本执行历史表失败: %w", err)
	}

	logger.Debug("脚本执行历史表创建成功", "driver", driver)
	return nil
}

// createStatementHistoryTable 创建语句执行历史表（语句级别）
func createStatementHistoryTable(ctx context.Context, conn database.Database, driver string) error {
	var createTableSQL string

	switch driver {
	case dbtypes.DriverMySQL:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_STATEMENT_EXECUTION_HISTORY (
    statementId VARCHAR(32) NOT NULL PRIMARY KEY,
    tenantId VARCHAR(32) NOT NULL DEFAULT 'default',
    scriptName VARCHAR(255) NOT NULL,
    statementHash VARCHAR(32) NOT NULL,
    statementType VARCHAR(50) NOT NULL,
    statementContent TEXT NOT NULL,
    databaseDriver VARCHAR(50) NOT NULL,
    executionStatus VARCHAR(20) NOT NULL,
    executionTime DATETIME NOT NULL,
    executionDuration BIGINT NOT NULL DEFAULT 0,
    errorMessage TEXT,
    createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX IDX_STMT_HIST_NAME (scriptName),
    INDEX IDX_STMT_HIST_HASH (statementHash),
    INDEX IDX_STMT_HIST_TYPE (statementType),
    INDEX IDX_STMT_HIST_STATUS (executionStatus),
    INDEX IDX_STMT_HIST_DRIVER (databaseDriver),
    INDEX IDX_STMT_HIST_TIME (executionTime),
    UNIQUE KEY UK_STMT_HASH (tenantId, scriptName, statementHash, databaseDriver)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='SQL语句执行历史表';`

	case dbtypes.DriverSQLite:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_STATEMENT_EXECUTION_HISTORY (
    statementId TEXT NOT NULL PRIMARY KEY,
    tenantId TEXT NOT NULL DEFAULT 'default',
    scriptName TEXT NOT NULL,
    statementHash TEXT NOT NULL,
    statementType TEXT NOT NULL,
    statementContent TEXT NOT NULL,
    databaseDriver TEXT NOT NULL,
    executionStatus TEXT NOT NULL,
    executionTime DATETIME NOT NULL,
    executionDuration INTEGER NOT NULL DEFAULT 0,
    errorMessage TEXT,
    createdAt DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_NAME ON HUB_STATEMENT_EXECUTION_HISTORY(scriptName);
CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_HASH ON HUB_STATEMENT_EXECUTION_HISTORY(statementHash);
CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_TYPE ON HUB_STATEMENT_EXECUTION_HISTORY(statementType);
CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_STATUS ON HUB_STATEMENT_EXECUTION_HISTORY(executionStatus);
CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_DRIVER ON HUB_STATEMENT_EXECUTION_HISTORY(databaseDriver);
CREATE INDEX IF NOT EXISTS IDX_STMT_HIST_TIME ON HUB_STATEMENT_EXECUTION_HISTORY(executionTime);
CREATE UNIQUE INDEX IF NOT EXISTS UK_STMT_HASH ON HUB_STATEMENT_EXECUTION_HISTORY(tenantId, scriptName, statementHash, databaseDriver);`

	case dbtypes.DriverOracle:
		createTableSQL = `
BEGIN
    EXECUTE IMMEDIATE 'CREATE TABLE HUB_STATEMENT_EXECUTION_HISTORY (
        statementId VARCHAR2(32) NOT NULL PRIMARY KEY,
        tenantId VARCHAR2(32) DEFAULT ''default'' NOT NULL,
        scriptName VARCHAR2(255) NOT NULL,
        statementHash VARCHAR2(32) NOT NULL,
        statementType VARCHAR2(50) NOT NULL,
        statementContent CLOB NOT NULL,
        databaseDriver VARCHAR2(50) NOT NULL,
        executionStatus VARCHAR2(20) NOT NULL,
        executionTime DATE NOT NULL,
        executionDuration NUMBER(19) DEFAULT 0 NOT NULL,
        errorMessage CLOB,
        createdAt DATE DEFAULT SYSDATE NOT NULL
    )';
EXCEPTION
    WHEN OTHERS THEN
        IF SQLCODE != -955 THEN
            RAISE;
        END IF;
END;`

	case dbtypes.DriverClickHouse:
		createTableSQL = `
CREATE TABLE IF NOT EXISTS HUB_STATEMENT_EXECUTION_HISTORY (
    statementId String,
    tenantId String DEFAULT 'default',
    scriptName String,
    statementHash String,
    statementType String,
    statementContent String,
    databaseDriver String,
    executionStatus String,
    executionTime DateTime,
    executionDuration Int64 DEFAULT 0,
    errorMessage String,
    createdAt DateTime DEFAULT now()
) ENGINE = MergeTree()
ORDER BY (tenantId, scriptName, statementHash)
SETTINGS index_granularity = 8192;`

	default:
		return fmt.Errorf("不支持的数据库驱动类型: %s", driver)
	}

	logger.Debug("创建语句执行历史表", "driver", driver)
	_, err := conn.Exec(ctx, createTableSQL, nil, false)
	if err != nil {
		return fmt.Errorf("创建语句执行历史表失败: %w", err)
	}

	logger.Debug("语句执行历史表创建成功", "driver", driver)
	return nil
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

// isStatementAlreadyExecuted 检查SQL语句是否已经执行过
// 查询语句执行历史表，检查指定哈希值的语句是否已成功执行
// 参数:
//   - ctx: 上下文对象，用于控制查询超时和取消
//   - conn: 数据库连接实例
//   - driver: 数据库驱动类型
//   - scriptName: 脚本文件名
//   - statementHash: 语句哈希值
//
// 返回:
//   - bool: true表示语句已执行过，false表示未执行过
//   - error: 查询失败时返回错误信息
//
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
		return "", fmt.Errorf("查询语句执行状态失败: %w", err)
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

	insertSQL := `INSERT INTO HUB_SCRIPT_EXECUTION_HISTORY 
				  (executionId, tenantId, scriptName, scriptPath, scriptVersion, databaseDriver, 
				   executionStatus, executionTime, executionDuration, statementsExecuted, errorMessage, createdAt)
				  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	durationMs := duration.Milliseconds()

	_, err := conn.Exec(ctx, insertSQL, []interface{}{
		executionId, tenantId, scriptName, scriptPath, scriptVersion, driver,
		status, now, durationMs, statementsExecuted, errorMessage, now,
	}, false)

	if err != nil {
		logger.Error("记录脚本执行历史失败",
			"error", err,
			"script", scriptName,
			"version", scriptVersion,
			"status", status)
	} else {
		logger.Debug("脚本执行历史记录成功",
			"script", scriptName,
			"version", scriptVersion,
			"status", status,
			"duration_ms", durationMs)
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
// 生成唯一的语句执行记录标识符，基于当前时间戳和纳秒数
// 返回:
//   - string: 格式为 "STMT_时间戳_纳秒" 的唯一标识符
func generateStatementId() string {
	now := time.Now()
	return fmt.Sprintf("STMT_%d_%d", now.Unix(), now.Nanosecond()%1000000)
}

// generateExecutionId 生成执行记录ID
// 生成唯一的执行记录标识符，基于当前时间戳和纳秒数
// 返回:
//   - string: 格式为 "EXEC_时间戳_纳秒" 的唯一标识符
func generateExecutionId() string {
	now := time.Now()
	return fmt.Sprintf("EXEC_%d_%d", now.Unix(), now.Nanosecond()%1000000)
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
	result := executeScriptForDatabase(ctx, "manual", db, driver, scriptDir)

	return &result, nil
}
