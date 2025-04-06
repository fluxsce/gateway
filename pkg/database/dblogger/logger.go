package dblogger

import (
	"fmt"
	"strings"
	"time"

	"gohub/pkg/database/dbtypes"
	"gohub/pkg/logger"
)

// DBLogger 数据库日志记录器
// 为数据库操作提供专门的日志记录功能
type DBLogger struct {
	// 是否启用日志
	Enabled bool
	// 慢查询阈值（毫秒）
	SlowThreshold int
	// 是否打印完整SQL
	PrintFullSQL bool
	// 是否打印执行时间
	PrintExecTime bool
	// 是否打印调用者信息
	PrintCaller bool
	// 是否记录事务操作
	PrintTransaction bool
}

// NewDBLogger 创建新的数据库日志记录器
// 参数:
//   - config: 数据库配置
//
// 返回:
//   - *DBLogger: 数据库日志记录器
func NewDBLogger(config *dbtypes.DbConfig) *DBLogger {
	return &DBLogger{
		Enabled:          config.Log.Enable,
		SlowThreshold:    config.Log.SlowThreshold,
		PrintFullSQL:     true, // 默认打印完整SQL
		PrintExecTime:    true, // 默认打印执行时间
		PrintCaller:      true, // 默认打印调用者信息
		PrintTransaction: true, // 默认记录事务操作
	}
}

// LogSQL 记录SQL执行日志
// 参数:
//   - operation: 操作类型描述
//   - query: 执行的SQL查询语句
//   - args: SQL查询的参数
//   - err: 执行过程中产生的错误
//   - elapsed: SQL执行耗时
func (l *DBLogger) LogSQL(operation string, query string, args []any, err error, elapsed time.Duration) {
	// 如果未启用日志，直接返回
	if !l.Enabled {
		return
	}

	// 处理SQL查询，替换参数占位符以提高可读性
	formattedSQL := l.formatSQL(query, args)

	// 如果有错误，记录错误日志
	if err != nil {
		// 记录错误
		logger.Error(operation+"错误", err)
		// 添加详细信息
		logger.Debug(operation+"详情",
			"sql", formattedSQL,
			"error", err.Error(),
		)
		return
	}

	// 构建日志字段
	fields := []any{
		"sql", formattedSQL,
	}

	// 添加执行时间
	if l.PrintExecTime {
		fields = append(fields, "elapsed", elapsed)
		fields = append(fields, "ms", elapsed.Milliseconds())
	}

	// 判断是否为慢查询
	if l.SlowThreshold > 0 && elapsed.Milliseconds() > int64(l.SlowThreshold) {
		// 记录慢查询日志
		logger.Warn("慢"+operation+" ["+fmt.Sprintf("%d", elapsed.Milliseconds())+"ms]", fields...)
	} else {
		// 记录普通SQL执行日志
		logger.Debug(operation, fields...)
	}
}

// LogTx 记录事务操作日志
// 参数:
//   - operation: 操作类型，如"开始事务"、"提交事务"、"回滚事务"
//   - err: 事务操作中产生的错误
func (l *DBLogger) LogTx(operation string, err error) {
	// 如果未启用日志或未启用事务日志，直接返回
	if !l.Enabled || !l.PrintTransaction {
		return
	}

	if err != nil {
		logger.Error("事务"+operation+"错误", err)
	} else {
		logger.Info("事务" + operation)
	}
}

// LogError 记录数据库错误
// 参数:
//   - operation: 操作类型描述
//   - err: 错误信息
func (l *DBLogger) LogError(operation string, err error) {
	// 如果未启用日志，直接返回
	if !l.Enabled {
		return
	}

	logger.Error(operation+"错误", err)
}

// formatSQL 格式化SQL语句，替换占位符以提高可读性
// 参数:
//   - query: SQL查询语句
//   - args: 查询参数
//
// 返回:
//   - string: 格式化后的SQL语句
func (l *DBLogger) formatSQL(query string, args []any) string {
	if !l.PrintFullSQL || len(args) == 0 {
		return query
	}

	// 简单实现：替换所有?为对应参数的字符串表示
	formattedSQL := query
	for _, arg := range args {
		var argStr string
		switch v := arg.(type) {
		case string:
			// 为字符串添加引号
			argStr = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "\\'"))
		case []byte:
			// 二进制数据简化为[binary]
			argStr = "[binary]"
		case time.Time:
			// 格式化时间
			argStr = fmt.Sprintf("'%s'", v.Format(time.RFC3339))
		case nil:
			// NULL值
			argStr = "NULL"
		default:
			// 其他类型直接转字符串
			argStr = fmt.Sprintf("%v", v)
		}

		// 替换第一个?
		pos := strings.Index(formattedSQL, "?")
		if pos >= 0 {
			formattedSQL = formattedSQL[:pos] + argStr + formattedSQL[pos+1:]
		}
	}

	return formattedSQL
}

// LogQueryResult 记录查询结果
// 参数:
//   - operation: 操作类型描述
//   - count: 结果数量
//   - err: 查询过程中产生的错误
func (l *DBLogger) LogQueryResult(operation string, count int, err error) {
	// 如果未启用日志，直接返回
	if !l.Enabled {
		return
	}

	if err != nil {
		logger.Error(operation+"结果错误", err)
	} else {
		logger.Debug(operation+"结果", "count", count)
	}
}

// SetEnabled 设置是否启用日志
// 参数:
//   - enabled: 是否启用
func (l *DBLogger) SetEnabled(enabled bool) {
	l.Enabled = enabled
}

// SetSlowThreshold 设置慢查询阈值
// 参数:
//   - threshold: 阈值（毫秒）
func (l *DBLogger) SetSlowThreshold(threshold int) {
	l.SlowThreshold = threshold
}

// SetPrintFullSQL 设置是否打印完整SQL
// 参数:
//   - print: 是否打印
func (l *DBLogger) SetPrintFullSQL(print bool) {
	l.PrintFullSQL = print
}

// SetPrintExecTime 设置是否打印执行时间
// 参数:
//   - print: 是否打印
func (l *DBLogger) SetPrintExecTime(print bool) {
	l.PrintExecTime = print
}

// SetPrintCaller 设置是否打印调用者信息
// 参数:
//   - print: 是否打印
func (l *DBLogger) SetPrintCaller(print bool) {
	l.PrintCaller = print
}

// SetPrintTransaction 设置是否记录事务操作
// 参数:
//   - print: 是否记录
func (l *DBLogger) SetPrintTransaction(print bool) {
	l.PrintTransaction = print
}

// LogConnecting 记录数据库连接过程
// 参数:
//   - driverName: 数据库驱动名称
//   - dsn: 数据源名称（连接字符串，会被掩码处理）
func (l *DBLogger) LogConnecting(driverName string, dsn string) {
	if !l.Enabled {
		return
	}

	// 掩码处理DSN，确保敏感信息不被记录
	maskedDSN := MaskDSN(dsn)

	logger.Info("连接数据库",
		"driver", driverName,
		"dsn", maskedDSN,
	)
}

// LogConnected 记录数据库连接成功
// 参数:
//   - driverName: 数据库驱动名称
//   - poolSettings: 连接池设置信息
func (l *DBLogger) LogConnected(driverName string, poolSettings map[string]any) {
	if !l.Enabled {
		return
	}

	args := []any{
		"driver", driverName,
	}

	// 添加连接池设置信息
	for k, v := range poolSettings {
		args = append(args, k, v)
	}

	logger.Info("数据库连接成功", args...)
}

// LogDisconnect 记录数据库断开连接
// 参数:
//   - driverName: 数据库驱动名称
func (l *DBLogger) LogDisconnect(driverName string) {
	if !l.Enabled {
		return
	}

	logger.Info("关闭数据库连接", "driver", driverName)
}

// LogPing 记录Ping操作
// 参数:
//   - err: Ping操作产生的错误，nil表示Ping成功
func (l *DBLogger) LogPing(err error) {
	if !l.Enabled {
		return
	}

	if err != nil {
		logger.Error("数据库Ping失败", err)
	} else {
		logger.Debug("数据库Ping成功")
	}
}

// MaskDSN 掩码处理数据源名称，隐藏敏感信息
// 参数:
//   - dsn: 原始数据源名称
//
// 返回:
//   - string: 掩码处理后的数据源名称
func MaskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}

	// 处理MySQL风格的DSN: username:password@tcp(host:port)/dbname
	if strings.Contains(dsn, "@") {
		// 寻找密码部分并进行掩码处理
		parts := strings.Split(dsn, "@")
		if len(parts) < 2 {
			return dsn
		}

		credentials := strings.Split(parts[0], ":")
		if len(credentials) >= 2 {
			// 将密码替换为 ****
			credentials[1] = "****"
			parts[0] = strings.Join(credentials, ":")
			return strings.Join(parts, "@")
		}
	}

	// 其他类型的DSN，尝试通用处理方式
	// 这里简化处理，只返回一个通用掩码信息
	return "[masked_connection_string]"
}
