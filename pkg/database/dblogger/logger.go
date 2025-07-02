package dblogger

import (
	"context"
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
//   - ctx: 上下文，用于链路追踪
//   - operation: 操作类型描述
//   - query: 执行的SQL查询语句
//   - args: SQL查询的参数
//   - err: 执行过程中产生的错误
//   - duration: SQL执行耗时
//   - extra: 额外信息
func (l *DBLogger) LogSQL(ctx context.Context, operation string, query string, args []any, err error, duration time.Duration, extra map[string]interface{}) {
	if !l.Enabled {
		return
	}

	// 处理SQL查询，替换参数占位符以提高可读性
	formattedSQL := l.formatSQL(query, args)

	// 构建基础日志字段
	fields := []any{
		"sql", formattedSQL,
		"duration", duration.String(),
		"ms", duration.Milliseconds(),
	}

	// 添加额外信息
	if extra != nil {
		for k, v := range extra {
			fields = append(fields, k, v)
		}
	}

	// 如果有错误，记录错误日志
	if err != nil {
		// 记录错误
		logger.ErrorWithTrace(ctx, operation+"错误", append(fields, "error", err.Error())...)
		return
	}

	// 判断是否为慢查询
	if l.SlowThreshold > 0 && duration.Milliseconds() > int64(l.SlowThreshold) {
		// 记录慢查询日志
		logger.WarnWithTrace(ctx, "慢"+operation+" ["+fmt.Sprintf("%d", duration.Milliseconds())+"ms]", fields...)
	} else {
		// 记录普通SQL执行日志
		logger.DebugWithTrace(ctx, operation, fields...)
	}
}

// LogTx 记录事务操作日志
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - operation: 操作类型，如"开始事务"、"提交事务"、"回滚事务"
//   - err: 事务操作中产生的错误
func (l *DBLogger) LogTx(ctx context.Context, operation string, err error) {
	// 如果未启用日志或未启用事务日志，直接返回
	if !l.Enabled || !l.PrintTransaction {
		return
	}

	if err != nil {
		logger.ErrorWithTrace(ctx, "事务"+operation+"错误", err)
	} else {
		logger.InfoWithTrace(ctx, "事务"+operation)
	}
}

// LogError 记录数据库错误
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - operation: 操作类型描述
//   - err: 错误信息
func (l *DBLogger) LogError(ctx context.Context, operation string, err error) {
	// 如果未启用日志，直接返回
	if !l.Enabled {
		return
	}

	logger.ErrorWithTrace(ctx, operation+"错误", err)
}

// formatSQL 格式化SQL语句，替换占位符以提高可读性
// 支持标准?占位符和Oracle :1,:2格式占位符
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

	formattedSQL := query
	
	// 检查是否为Oracle格式的占位符 (:1, :2, :3...)
	if strings.Contains(query, ":1") || strings.Contains(query, ":2") {
		// Oracle格式占位符替换
		for i, arg := range args {
			placeholder := fmt.Sprintf(":%d", i+1)
			argStr := l.formatArgument(arg)
			formattedSQL = strings.ReplaceAll(formattedSQL, placeholder, argStr)
		}
	} else {
		// 标准?占位符替换
		for _, arg := range args {
			argStr := l.formatArgument(arg)
			// 替换第一个?
			pos := strings.Index(formattedSQL, "?")
			if pos >= 0 {
				formattedSQL = formattedSQL[:pos] + argStr + formattedSQL[pos+1:]
			}
		}
	}

	return formattedSQL
}

// formatArgument 格式化单个参数
func (l *DBLogger) formatArgument(arg any) string {
	switch v := arg.(type) {
	case string:
		// 为字符串添加引号
		return fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "\\'"))
	case []byte:
		// 二进制数据简化为[binary]
		return "[binary]"
	case time.Time:
		// 格式化时间
		return fmt.Sprintf("'%s'", v.Format(time.RFC3339))
	case nil:
		// NULL值
		return "NULL"
	default:
		// 其他类型直接转字符串
		return fmt.Sprintf("%v", v)
	}
}

// LogQueryResult 记录查询结果
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - operation: 操作类型描述
//   - count: 结果数量
//   - err: 查询过程中产生的错误
func (l *DBLogger) LogQueryResult(ctx context.Context, operation string, count int, err error) {
	// 如果未启用日志，直接返回
	if !l.Enabled {
		return
	}

	if err != nil {
		logger.ErrorWithTrace(ctx, operation+"结果错误", err)
	} else {
		logger.DebugWithTrace(ctx, operation+"结果", "count", count)
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
//   - ctx: 上下文，用于链路追踪
//   - driverName: 数据库驱动名称
//   - dsn: 数据源名称（连接字符串，会被掩码处理）
func (l *DBLogger) LogConnecting(ctx context.Context, driverName string, dsn string) {
	if !l.Enabled {
		return
	}

	// 掩码处理DSN，确保敏感信息不被记录
	maskedDSN := MaskDSN(dsn)

	logger.InfoWithTrace(ctx, "连接数据库",
		"driver", driverName,
		"dsn", maskedDSN,
	)
}

// LogConnected 记录数据库连接成功
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - driverName: 数据库驱动名称
//   - poolSettings: 连接池设置信息
func (l *DBLogger) LogConnected(ctx context.Context, driverName string, poolSettings map[string]any) {
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

	logger.InfoWithTrace(ctx, "数据库连接成功", args...)
}

// LogDisconnect 记录数据库断开连接
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - driverName: 数据库驱动名称
func (l *DBLogger) LogDisconnect(ctx context.Context, driverName string) {
	if !l.Enabled {
		return
	}

	logger.InfoWithTrace(ctx, "关闭数据库连接", "driver", driverName)
}

// LogPing 记录Ping操作
// 参数:
//   - ctx: 上下文，用于链路追踪
//   - err: Ping操作产生的错误，nil表示Ping成功
func (l *DBLogger) LogPing(ctx context.Context, err error) {
	if !l.Enabled {
		return
	}

	if err != nil {
		logger.ErrorWithTrace(ctx, "数据库Ping失败", err)
	} else {
		logger.DebugWithTrace(ctx, "数据库Ping成功")
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
