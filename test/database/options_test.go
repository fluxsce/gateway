package database

import (
	"testing"

	"gateway/pkg/database"
	"gateway/pkg/database/dbtypes"
)

// 测试数据库配置选项
func TestDbConfig(t *testing.T) {
	// 测试默认值
	config := &database.DbConfig{
		Driver: dbtypes.DriverMySQL,
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
		Driver: dbtypes.DriverMySQL,
		DSN:    "user:pass@tcp(localhost:3306)/testdb",
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    20,
			MaxIdleConns:    10,
			ConnMaxLifetime: 7200, // 2小时，单位秒
			ConnMaxIdleTime: 2700, // 45分钟，单位秒
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
		Transaction: dbtypes.TransactionConfig{
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

// 测试事务选项
func TestTxOptions(t *testing.T) {
	// 测试默认选项
	txOpts := &database.TxOptions{}
	if txOpts.Isolation != 0 {
		t.Errorf("Isolation默认值应该是0，实际是%d", txOpts.Isolation)
	}

	if txOpts.ReadOnly {
		t.Errorf("ReadOnly默认值应该是false")
	}

	// 测试自定义选项
	txOpts = &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  true,
	}

	if txOpts.Isolation != database.IsolationReadCommitted {
		t.Errorf("Isolation应该是ReadCommitted，实际是%d", txOpts.Isolation)
	}

	if !txOpts.ReadOnly {
		t.Errorf("ReadOnly应该是true")
	}
}
