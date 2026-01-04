package db

import (
	"context"
	"fmt"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/logger"
)

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
