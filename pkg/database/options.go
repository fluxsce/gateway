package database

import (
	"gohub/pkg/database/dbtypes"
)

// 将类型重定向到dbtypes包
type (
	// ConnectionConfig 数据库连接配置
	ConnectionConfig = dbtypes.ConnectionConfig
	// PoolConfig 连接池配置
	PoolConfig = dbtypes.PoolConfig
	// LogConfig 日志配置
	LogConfig = dbtypes.LogConfig
	// TransactionConfig 事务配置
	TransactionConfig = dbtypes.TransactionConfig
)

// ExecOptions 执行选项
// 用于控制SQL执行操作的行为
type ExecOptions struct {
	// UseTransaction 是否使用事务
	// 如果为空，则使用DefaultUseTransaction
	// 值为true时，SQL操作将在事务中执行
	UseTransaction *bool
}

// ExecOption 执行选项函数
// 函数签名，用于修改ExecOptions结构
type ExecOption func(*ExecOptions)

// WithTransaction 设置是否使用事务
// 参数:
//   - useTransaction: 是否在事务中执行SQL操作
//
// 返回:
//   - ExecOption: 执行选项函数
//
// 用法示例:
//
//	db.ExecWithOptions(ctx, query, args, WithTransaction(true))
func WithTransaction(useTransaction bool) ExecOption {
	return func(o *ExecOptions) {
		o.UseTransaction = &useTransaction
	}
}

// QueryOptions 查询选项
// 用于控制SQL查询操作的行为
type QueryOptions struct {
	// UseTransaction 是否使用事务
	// 如果为空，则使用DefaultUseTransaction
	// 值为true时，查询操作将在事务中执行
	UseTransaction *bool
}

// QueryOption 查询选项函数
// 函数签名，用于修改QueryOptions结构
type QueryOption func(*QueryOptions)

// WithQueryTransaction 设置查询是否使用事务
// 参数:
//   - useTransaction: 是否在事务中执行查询操作
//
// 返回:
//   - QueryOption: 查询选项函数
//
// 用法示例:
//
//	db.QueryWithOptions(ctx, &users, query, args, WithQueryTransaction(true))
func WithQueryTransaction(useTransaction bool) QueryOption {
	return func(o *QueryOptions) {
		o.UseTransaction = &useTransaction
	}
}

// TxOptions 事务选项
// 用于控制事务的行为特性
type TxOptions struct {
	// Isolation 事务隔离级别
	// 0: 默认, 1: 读未提交, 2: 读已提交, 3: 可重复读, 4: 串行化
	// 控制事务的隔离级别，影响并发操作的可见性和一致性
	Isolation int

	// ReadOnly 是否只读事务
	// 设置为true时，事务将不允许修改数据
	ReadOnly bool
}

// TxOption 事务选项函数
// 函数签名，用于修改TxOptions结构
type TxOption func(*TxOptions)

// WithIsolation 设置事务隔离级别
// 参数:
//   - isolation: 隔离级别 (0-4)
//
// 返回:
//   - TxOption: 事务选项函数
//
// 用法示例:
//
//	db.BeginTx(ctx, WithIsolation(3)) // 设置为可重复读隔离级别
func WithIsolation(isolation int) TxOption {
	return func(o *TxOptions) {
		o.Isolation = isolation
	}
}

// WithReadOnly 设置是否只读事务
// 参数:
//   - readOnly: 是否为只读事务
//
// 返回:
//   - TxOption: 事务选项函数
//
// 用法示例:
//
//	db.BeginTx(ctx, WithReadOnly(true)) // 创建只读事务
func WithReadOnly(readOnly bool) TxOption {
	return func(o *TxOptions) {
		o.ReadOnly = readOnly
	}
}

// NewExecOptions 创建新的执行选项
// 参数:
//   - config: 数据库配置
//   - options: 执行选项变参
//
// 返回:
//   - *ExecOptions: 配置好的执行选项对象
//
// 内部使用，设置默认值并应用用户提供的选项
func NewExecOptions(config *dbtypes.DbConfig, options ...ExecOption) *ExecOptions {
	// 获取默认是否使用事务
	defaultUseTransaction := config.Transaction.DefaultUse

	// 初始化选项，使用配置的默认值
	opts := &ExecOptions{
		UseTransaction: &defaultUseTransaction,
	}

	// 应用用户提供的选项
	for _, option := range options {
		option(opts)
	}

	return opts
}

// NewQueryOptions 创建新的查询选项
// 参数:
//   - config: 数据库配置
//   - options: 查询选项变参
//
// 返回:
//   - *QueryOptions: 配置好的查询选项对象
//
// 内部使用，设置默认值并应用用户提供的选项
func NewQueryOptions(config *dbtypes.DbConfig, options ...QueryOption) *QueryOptions {
	// 获取默认是否使用事务
	defaultUseTransaction := config.Transaction.DefaultUse

	// 初始化选项，使用配置的默认值
	opts := &QueryOptions{
		UseTransaction: &defaultUseTransaction,
	}

	// 应用用户提供的选项
	for _, option := range options {
		option(opts)
	}

	return opts
}

// NewTxOptions 创建新的事务选项
// 参数:
//   - options: 事务选项变参
//
// 返回:
//   - *TxOptions: 配置好的事务选项对象
//
// 内部使用，设置默认值并应用用户提供的选项
func NewTxOptions(options ...TxOption) *TxOptions {
	// 初始化选项，使用默认值
	opts := &TxOptions{
		Isolation: 0, // 默认隔离级别
		ReadOnly:  false,
	}

	// 应用用户提供的选项
	for _, option := range options {
		option(opts)
	}

	return opts
}
