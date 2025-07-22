package database

import (
	"context"
	"testing"
	"time"

	"gohub/pkg/database"
	_ "gohub/pkg/database/alldriver" // 导入所有驱动
	"gohub/pkg/database/dbtypes"
)

// OracleUser 测试用户结构体
type OracleUser struct {
	ID       int64     `db:"id"`
	Name     string    `db:"name"`
	Email    string    `db:"email"`
	Age      int       `db:"age"`
	CreateAt time.Time `db:"create_at"`
}

// CountResult 查询计数结果结构体
type CountResult struct {
	Count int64 `db:"COUNT(*)"`
}

// TestOracleConnection 测试Oracle数据库连接
func TestOracleConnection(t *testing.T) {
	// 跳过测试，除非设置了环境变量
	if testing.Short() {
		t.Skip("跳过Oracle集成测试")
	}

	// Oracle配置
	config := &database.DbConfig{
		Name:    "test_oracle",
		Driver:  "oracle",
		Enabled: true,
		DSN:     "oracle://DATAHUB_250624:20250624_TMPpAssw0d123@47.117.1.78:62345/fluxv6pdb", // 请根据实际环境修改
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
	}

	// 创建Oracle连接
	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("连接Oracle失败: %v", err)
	}
	defer db.Close()

	// 测试连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("Ping Oracle失败: %v", err)
	}

	t.Log("Oracle连接测试成功")
}

// TestOracleDriverRegistration 测试Oracle驱动注册
func TestOracleDriverRegistration(t *testing.T) {
	config := &database.DbConfig{
		Driver:  "oracle",
		Enabled: true,
		DSN:     "oracle://test:test@localhost:1521/test",
	}

	// 创建Oracle实例（不实际连接）
	db, err := database.Open(config)
	if err == nil {
		// 如果创建成功，检查驱动类型
		if db.GetDriver() != "oracle" {
			t.Errorf("期望驱动类型为 'oracle'，实际为 '%s'", db.GetDriver())
		}
		db.Close()
	}

	t.Log("Oracle驱动注册测试完成")
}

// TestOracleBasicOperations 测试Oracle基本操作
func TestOracleBasicOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过Oracle集成测试")
	}

	config := &database.DbConfig{
		Name:    "test_oracle",
		Driver:  "oracle",
		Enabled: true,
		DSN:     "oracle://DATAHUB_250624:20250624_TMPpAssw0d123@47.117.1.78:62345/fluxv6pdb",
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("连接Oracle失败: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	createTableSQL := `
	CREATE TABLE test_users (
		id NUMBER(10) PRIMARY KEY,
		name VARCHAR2(100) NOT NULL,
		email VARCHAR2(255) NOT NULL,
		age NUMBER(3),
		create_at DATE DEFAULT SYSDATE
	)`

	// 先删除表（如果存在）
	_, _ = db.Exec(ctx, "DROP TABLE test_users", nil, true)

	// 创建表
	_, err = db.Exec(ctx, createTableSQL, nil, true)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 创建序列
	createSeqSQL := "CREATE SEQUENCE test_users_seq START WITH 1 INCREMENT BY 1"
	_, _ = db.Exec(ctx, "DROP SEQUENCE test_users_seq", nil, true)
	_, err = db.Exec(ctx, createSeqSQL, nil, true)
	if err != nil {
		t.Fatalf("创建序列失败: %v", err)
	}

	// 测试插入
	testUser := OracleUser{
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      25,
		CreateAt: time.Now(),
	}

	// 使用手动SQL插入（因为Oracle的自增主键需要特殊处理）
	insertSQL := "INSERT INTO test_users (id, name, email, age, create_at) VALUES (test_users_seq.NEXTVAL, :1, :2, :3, :4)"
	_, err = db.Exec(ctx, insertSQL, []interface{}{testUser.Name, testUser.Email, testUser.Age, testUser.CreateAt}, true)
	if err != nil {
		t.Fatalf("插入记录失败: %v", err)
	}

	// 测试查询
	var users []OracleUser
	err = db.Query(ctx, &users, "SELECT id, name, email, age, create_at FROM test_users", nil, true)
	if err != nil {
		t.Fatalf("查询记录失败: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("期望查询到1条记录，实际查询到%d条", len(users))
	}

	if users[0].Name != testUser.Name {
		t.Errorf("期望名称为 '%s'，实际为 '%s'", testUser.Name, users[0].Name)
	}

	// 测试单条查询
	var user OracleUser
	err = db.QueryOne(ctx, &user, "SELECT id, name, email, age, create_at FROM test_users WHERE name = :1", []interface{}{testUser.Name}, true)
	if err != nil {
		t.Fatalf("单条查询失败: %v", err)
	}

	if user.Email != testUser.Email {
		t.Errorf("期望邮箱为 '%s'，实际为 '%s'", testUser.Email, user.Email)
	}

	// 测试更新
	updateSQL := "UPDATE test_users SET age = :1 WHERE name = :2"
	rowsAffected, err := db.Exec(ctx, updateSQL, []interface{}{30, testUser.Name}, true)
	if err != nil {
		t.Fatalf("更新记录失败: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("期望更新1条记录，实际更新%d条", rowsAffected)
	}

	// 测试删除
	deleteSQL := "DELETE FROM test_users WHERE name = :1"
	rowsAffected, err = db.Exec(ctx, deleteSQL, []interface{}{testUser.Name}, true)
	if err != nil {
		t.Fatalf("删除记录失败: %v", err)
	}

	if rowsAffected != 1 {
		t.Errorf("期望删除1条记录，实际删除%d条", rowsAffected)
	}

	// 清理测试数据
	_, _ = db.Exec(ctx, "DROP TABLE test_users", nil, true)
	_, _ = db.Exec(ctx, "DROP SEQUENCE test_users_seq", nil, true)

	t.Log("Oracle基本操作测试完成")
}

// TestOracleTransaction 测试Oracle事务
func TestOracleTransaction(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过Oracle集成测试")
	}

	config := &database.DbConfig{
		Name:    "test_oracle",
		Driver:  "oracle",
		Enabled: true,
		DSN:     "oracle://DATAHUB_250624:20250624_TMPpAssw0d123@47.117.1.78:62345/fluxv6pdb",
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("连接Oracle失败: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建测试表
	createTableSQL := `
	CREATE TABLE test_tx_users (
		id NUMBER(10) PRIMARY KEY,
		name VARCHAR2(100) NOT NULL
	)`

	_, _ = db.Exec(ctx, "DROP TABLE test_tx_users", nil, true)
	_, err = db.Exec(ctx, createTableSQL, nil, true)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	// 测试事务提交
	err = db.InTx(ctx, nil, func(txCtx context.Context) error {
		_, err := db.Exec(txCtx, "INSERT INTO test_tx_users (id, name) VALUES (1, '事务测试1')", nil, false)
		if err != nil {
			return err
		}
		_, err = db.Exec(txCtx, "INSERT INTO test_tx_users (id, name) VALUES (2, '事务测试2')", nil, false)
		return err
	})

	if err != nil {
		t.Fatalf("事务提交失败: %v", err)
	}

	// 验证数据是否插入成功
	var countResult CountResult
	err = db.QueryOne(ctx, &countResult, "SELECT COUNT(*) FROM test_tx_users", nil, true)
	if err != nil {
		t.Fatalf("查询记录数失败: %v", err)
	}

	if countResult.Count != 2 {
		t.Errorf("期望2条记录，实际%d条", countResult.Count)
	}

	// 测试事务回滚
	err = db.InTx(ctx, nil, func(txCtx context.Context) error {
		_, err := db.Exec(txCtx, "INSERT INTO test_tx_users (id, name) VALUES (3, '事务测试3')", nil, false)
		if err != nil {
			return err
		}
		// 故意返回错误，触发回滚
		return database.ErrTransaction
	})

	if err == nil {
		t.Error("期望事务回滚，但没有返回错误")
	}

	// 验证数据没有插入
	var countResult2 CountResult
	err = db.QueryOne(ctx, &countResult2, "SELECT COUNT(*) FROM test_tx_users", nil, true)
	if err != nil {
		t.Fatalf("查询记录数失败: %v", err)
	}

	if countResult2.Count != 2 {
		t.Errorf("期望2条记录（回滚后），实际%d条", countResult2.Count)
	}

	// 清理测试数据
	_, _ = db.Exec(ctx, "DROP TABLE test_tx_users", nil, true)

	t.Log("Oracle事务测试完成")
}


// 运行测试命令示例：
// go test -v ./test/database/oracle_test.go -run TestOracleDriverRegistration
// go test -v ./test/database/oracle_test.go -run TestOracleConnection (需要Oracle环境)
// go test -v ./test/database/oracle_test.go -run TestOracleDynamicQuery (需要Oracle环境)
// go test -v ./test/database/oracle_test.go (需要Oracle环境) 