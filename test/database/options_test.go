package database

import (
	"testing"

	"gohub/pkg/database"
)

// 测试数据库配置选项
func TestDbConfig(t *testing.T) {
	// 测试默认值
	config := &database.DbConfig{
		Driver: "mysql",
		DSN:    "user:pass@tcp(localhost:3306)/testdb",
	}

	// 确保默认值合理
	if config.Pool.MaxOpenConns == 0 {
		t.Log("Pool.MaxOpenConns默认值为0，可能需要设置一个合理的值")
	}

	if config.Pool.MaxIdleConns == 0 {
		t.Log("Pool.MaxIdleConns默认值为0，可能需要设置一个合理的值")
	}

	// 创建带有完整配置的对象
	fullConfig := &database.DbConfig{
		Driver: "mysql",
		DSN:    "user:pass@tcp(localhost:3306)/testdb",
		Pool: database.PoolConfig{
			MaxOpenConns:    20,
			MaxIdleConns:    10,
			ConnMaxLifetime: 7200, // 2小时，单位秒
			ConnMaxIdleTime: 2700, // 45分钟，单位秒
		},
		Log: database.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
		Transaction: database.TransactionConfig{
			DefaultUse: true,
		},
	}

	// 验证值已正确设置
	if fullConfig.Pool.MaxOpenConns != 20 {
		t.Errorf("Pool.MaxOpenConns应该是20，实际是%d", fullConfig.Pool.MaxOpenConns)
	}

	if fullConfig.Pool.MaxIdleConns != 10 {
		t.Errorf("Pool.MaxIdleConns应该是10，实际是%d", fullConfig.Pool.MaxIdleConns)
	}

	if fullConfig.Pool.ConnMaxLifetime != 7200 {
		t.Errorf("Pool.ConnMaxLifetime应该是7200秒，实际是%v", fullConfig.Pool.ConnMaxLifetime)
	}

	if fullConfig.Pool.ConnMaxIdleTime != 2700 {
		t.Errorf("Pool.ConnMaxIdleTime应该是2700秒，实际是%v", fullConfig.Pool.ConnMaxIdleTime)
	}

	if !fullConfig.Log.Enable {
		t.Errorf("Log.Enable应该是true")
	}

	if fullConfig.Log.SlowThreshold != 200 {
		t.Errorf("Log.SlowThreshold应该是200，实际是%d", fullConfig.Log.SlowThreshold)
	}

	if !fullConfig.Transaction.DefaultUse {
		t.Errorf("Transaction.DefaultUse应该是true")
	}
}

// 测试执行选项
func TestExecOptions(t *testing.T) {
	config := &database.DbConfig{
		Transaction: database.TransactionConfig{
			DefaultUse: true,
		},
	}

	// 测试默认选项
	execOpts := database.NewExecOptions(config)
	if !*execOpts.UseTransaction {
		t.Errorf("UseTransaction默认值应该继承自配置的Transaction.DefaultUse")
	}

	// 测试自定义选项
	execOpts = database.NewExecOptions(config, database.WithTransaction(false))
	if *execOpts.UseTransaction {
		t.Errorf("WithTransaction(false)应该将UseTransaction设为false")
	}

	// 测试多个选项
	useTransaction := true
	execOpts = database.NewExecOptions(config,
		database.WithTransaction(false),
		database.WithTransaction(useTransaction))

	if !*execOpts.UseTransaction {
		t.Errorf("后面的选项应该覆盖前面的选项")
	}
}

// 测试查询选项
func TestQueryOptions(t *testing.T) {
	config := &database.DbConfig{
		Transaction: database.TransactionConfig{
			DefaultUse: false,
		},
	}

	// 测试默认选项
	queryOpts := database.NewQueryOptions(config)
	if *queryOpts.UseTransaction {
		t.Errorf("UseTransaction默认值应该继承自配置的Transaction.DefaultUse")
	}

	// 测试自定义选项
	queryOpts = database.NewQueryOptions(config, database.WithQueryTransaction(true))
	if !*queryOpts.UseTransaction {
		t.Errorf("WithQueryTransaction(true)应该将UseTransaction设为true")
	}

	// 测试多个选项
	useTransaction := false
	queryOpts = database.NewQueryOptions(config,
		database.WithQueryTransaction(true),
		database.WithQueryTransaction(useTransaction))

	if *queryOpts.UseTransaction {
		t.Errorf("后面的选项应该覆盖前面的选项")
	}
}

// 测试事务选项
func TestTxOptions(t *testing.T) {
	// 测试默认选项
	txOpts := database.NewTxOptions()
	if txOpts.Isolation != 0 {
		t.Errorf("Isolation默认值应该是0，实际是%d", txOpts.Isolation)
	}

	if txOpts.ReadOnly {
		t.Errorf("ReadOnly默认值应该是false")
	}

	// 测试隔离级别选项
	txOpts = database.NewTxOptions(database.WithIsolation(2))
	if txOpts.Isolation != 2 {
		t.Errorf("WithIsolation(2)应该将Isolation设为2，实际是%d", txOpts.Isolation)
	}

	// 测试只读选项
	txOpts = database.NewTxOptions(database.WithReadOnly(true))
	if !txOpts.ReadOnly {
		t.Errorf("WithReadOnly(true)应该将ReadOnly设为true")
	}

	// 测试多个选项
	txOpts = database.NewTxOptions(
		database.WithIsolation(3),
		database.WithReadOnly(true))

	if txOpts.Isolation != 3 {
		t.Errorf("Isolation应该是3，实际是%d", txOpts.Isolation)
	}

	if !txOpts.ReadOnly {
		t.Errorf("ReadOnly应该是true")
	}

	// 测试选项覆盖
	txOpts = database.NewTxOptions(
		database.WithIsolation(1),
		database.WithIsolation(4),
		database.WithReadOnly(false),
		database.WithReadOnly(true))

	if txOpts.Isolation != 4 {
		t.Errorf("后面的隔离级别选项应该覆盖前面的，应该是4，实际是%d", txOpts.Isolation)
	}

	if !txOpts.ReadOnly {
		t.Errorf("后面的只读选项应该覆盖前面的，应该是true")
	}
}
