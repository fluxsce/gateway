package database

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"gohub/pkg/database"
	"gohub/pkg/database/dbtypes"
	_ "gohub/pkg/database/sqlite" // 导入SQLite实现
)

// SqliteUser 测试用户结构体
type SqliteUser struct {
	ID       int64     `db:"id"`
	Name     string    `db:"name"`
	Email    string    `db:"email"`
	Age      int       `db:"age"`
	CreateAt time.Time `db:"create_at"`
}

// TestSQLiteConnect 测试SQLite连接
func TestSQLiteConnect(t *testing.T) {
	// 创建临时数据库文件
	dbFile := "test_sqlite.db"
	defer os.Remove(dbFile)

	config := &dbtypes.DbConfig{
		Name:    "test_sqlite",
		Enabled: true,
		Driver:  dbtypes.DriverSQLite,
		DSN:     dbFile,
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
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
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	// 测试ping
	if err := db.Ping(context.Background()); err != nil {
		t.Fatalf("Failed to ping SQLite: %v", err)
	}

	// 验证驱动类型
	if db.GetDriver() != dbtypes.DriverSQLite {
		t.Fatalf("Expected driver %s, got %s", dbtypes.DriverSQLite, db.GetDriver())
	}
}

// TestSQLiteMemoryDatabase 测试内存数据库
func TestSQLiteMemoryDatabase(t *testing.T) {
	config := &dbtypes.DbConfig{
		Name:    "test_memory",
		Enabled: true,
		Driver:  dbtypes.DriverSQLite,
		DSN:     ":memory:",
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    5,
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("Failed to connect to memory SQLite: %v", err)
	}
	defer db.Close()

	// 测试基本操作
	ctx := context.Background()

	// 创建表
	createSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT UNIQUE NOT NULL,
			age INTEGER,
			create_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err = db.Exec(ctx, createSQL, nil, true)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 插入测试数据
	user := SqliteUser{
		Name:     "张三",
		Email:    "zhangsan@example.com",
		Age:      25,
		CreateAt: time.Now(),
	}

	id, err := db.Insert(ctx, "users", user, true)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	if id == 0 {
		t.Fatal("Expected non-zero insert ID")
	}

	// 查询单条记录
	var resultUser SqliteUser
	err = db.QueryOne(ctx, &resultUser, "SELECT id, name, email, age, create_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if resultUser.Name != user.Name || resultUser.Email != user.Email {
		t.Fatalf("Query result mismatch: expected %+v, got %+v", user, resultUser)
	}

	// 更新记录
	updateData := SqliteUser{
		Name: "李四",
		Age:  30,
	}
	affected, err := db.Update(ctx, "users", updateData, "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	if affected != 1 {
		t.Fatalf("Expected 1 affected row, got %d", affected)
	}

	// 验证更新结果
	err = db.QueryOne(ctx, &resultUser, "SELECT id, name, email, age, create_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("Failed to query updated user: %v", err)
	}
	if resultUser.Name != "李四" || resultUser.Age != 30 {
		t.Fatalf("Update verification failed: got %+v", resultUser)
	}

	// 删除记录
	affected, err = db.Delete(ctx, "users", "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	if affected != 1 {
		t.Fatalf("Expected 1 affected row, got %d", affected)
	}

	// 验证删除结果
	err = db.QueryOne(ctx, &resultUser, "SELECT id, name, email, age, create_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != database.ErrRecordNotFound {
		t.Fatalf("Expected record not found error, got: %v", err)
	}
}

// TestSQLiteTransaction 测试事务
func TestSQLiteTransaction(t *testing.T) {
	config := &dbtypes.DbConfig{
		Name:    "test_transaction",
		Enabled: true,
		Driver:  dbtypes.DriverSQLite,
		DSN:     ":memory:",
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建表
	createSQL := `
		CREATE TABLE IF NOT EXISTS test_tx (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL
		)
	`
	_, err = db.Exec(ctx, createSQL, nil, true)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 测试事务提交
	err = db.InTx(ctx, nil, func() error {
		_, err := db.Exec(ctx, "INSERT INTO test_tx (name) VALUES (?)", []interface{}{"test1"}, false)
		if err != nil {
			return err
		}
		_, err = db.Exec(ctx, "INSERT INTO test_tx (name) VALUES (?)", []interface{}{"test2"}, false)
		return err
	})
	if err != nil {
		t.Fatalf("Transaction failed: %v", err)
	}

	// 验证数据已提交
	var count int
	err = db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM test_tx", nil, true)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 records, got %d", count)
	}

	// 测试事务回滚
	err = db.InTx(ctx, nil, func() error {
		_, err := db.Exec(ctx, "INSERT INTO test_tx (name) VALUES (?)", []interface{}{"test3"}, false)
		if err != nil {
			return err
		}
		// 模拟错误导致回滚
		return fmt.Errorf("simulated error")
	})
	if err == nil {
		t.Fatal("Expected transaction to fail")
	}

	// 验证数据已回滚
	err = db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM test_tx", nil, true)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 records after rollback, got %d", count)
	}
}

// TestSQLiteBatchInsert 测试批量插入
func TestSQLiteBatchInsert(t *testing.T) {
	config := &dbtypes.DbConfig{
		Name:    "test_batch",
		Enabled: true,
		Driver:  dbtypes.DriverSQLite,
		DSN:     ":memory:",
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("Failed to connect to SQLite: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 创建表
	createSQL := `
		CREATE TABLE IF NOT EXISTS batch_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL
		)
	`
	_, err = db.Exec(ctx, createSQL, nil, true)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// 准备批量数据
	users := []SqliteUser{
		{Name: "用户1", Email: "user1@example.com"},
		{Name: "用户2", Email: "user2@example.com"},
		{Name: "用户3", Email: "user3@example.com"},
	}

	// 批量插入
	affected, err := db.BatchInsert(ctx, "batch_users", users, true)
	if err != nil {
		t.Fatalf("Failed to batch insert: %v", err)
	}
	if affected != 3 {
		t.Fatalf("Expected 3 affected rows, got %d", affected)
	}

	// 验证插入结果
	var count int
	err = db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM batch_users", nil, true)
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}
	if count != 3 {
		t.Fatalf("Expected 3 records, got %d", count)
	}
} 