package database

import (
	"context"
	"testing"
	"time"

	"gohub/pkg/database"
	_ "gohub/pkg/database/alldriver" // 导入驱动确保注册
)

// MySQLUser 用于测试的用户结构体
type MySQLUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName 实现Model接口
func (u MySQLUser) TableName() string {
	return "mysql_users"
}

// PrimaryKey 实现Model接口
func (u MySQLUser) PrimaryKey() string {
	return "id"
}

// 获取测试数据库连接
func getMySQLTestDB(t *testing.T) database.Database {
	// 创建测试数据库配置
	config := &database.DbConfig{
		Driver: database.DriverMySQL,
		Name:   "mysql_test",
		DSN:    "root:!@#wwe123@tcp(rm-bp1i81ckdj72o6vy2ao.mysql.rds.aliyuncs.com:3306)/wms_ftest?charset=utf8mb3&parseTime=True&loc=Local",
		Pool: database.PoolConfig{
			MaxOpenConns:    10,
			MaxIdleConns:    5,
			ConnMaxLifetime: 3600, // 1小时，单位秒
			ConnMaxIdleTime: 1800, // 30分钟，单位秒
		},
		Log: database.LogConfig{
			Enable:        true,
			SlowThreshold: 100, // 100毫秒
		},
	}

	// 直接打开数据库连接
	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("打开MySQL连接失败: %v", err)
	}

	// 验证连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("MySQL连接测试失败: %v", err)
	}
	t.Log("成功连接到MySQL数据库")

	return db
}

// 创建测试表
func setupMySQLTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()

	// 先尝试删除表（如果存在）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS mysql_users", nil)
	if err != nil {
		t.Fatalf("删除测试表失败: %v", err)
	}

	// 创建测试表
	_, err = db.Exec(ctx, `
		CREATE TABLE mysql_users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP NULL
		)
	`, nil)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	t.Log("测试表创建成功")
}

// 清理测试表
func cleanupMySQLTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS mysql_users", nil)
	if err != nil {
		t.Fatalf("清理测试表失败: %v", err)
	}
	t.Log("测试表清理成功")
}

// 测试表创建
func TestMySQLTableSetup(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	// 验证表是否创建成功
	ctx := context.Background()
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) AS count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'mysql_users'", nil)
	if err != nil {
		t.Fatalf("验证表创建失败: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("表创建验证失败，期望表存在(count=1)，实际为 %d", result.Count)
	} else {
		t.Log("表创建验证成功")
	}
}

// 测试单条插入
func TestMySQLInsert(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 使用当前时间而不是零值
	// now := time.Now()

	// 构建测试用户
	user := MySQLUser{
		Name:  "测试用户",
		Email: "test@example.com",
	}

	// 插入记录
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("插入用户记录失败: %v", err)
	}

	if id <= 0 {
		t.Errorf("插入记录返回的ID应该大于0，实际得到: %d", id)
	}

	t.Logf("插入用户成功，ID: %d", id)
}

// 测试批量插入
func TestMySQLBatchInsert(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 测试批量插入
	// 使用当前时间而不是零值
	now := time.Now()

	users := []MySQLUser{
		{Name: "用户A", Email: "usera@example.com", CreatedAt: now},
		{Name: "用户B", Email: "userb@example.com", CreatedAt: now},
		{Name: "用户C", Email: "userc@example.com"},
	}

	for i, user := range users {
		id, err := db.Insert(ctx, user.TableName(), user)
		if err != nil {
			t.Fatalf("插入用户记录失败 (索引 %d): %v", i, err)
		}
		t.Logf("成功插入用户: ID=%d, Name=%s", id, user.Name)
	}

	// 使用结构体接收查询结果
	type CountResult struct {
		Total int `db:"total"`
	}
	var result CountResult
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) AS total FROM mysql_users", nil)
	if err != nil {
		t.Fatalf("验证插入数量失败: %v", err)
	}

	if result.Total != len(users) {
		t.Errorf("批量插入验证失败，期望 %d 条记录，实际为 %d", len(users), result.Total)
	} else {
		t.Logf("批量插入验证成功，共插入 %d 条记录", result.Total)
	}
}

// 测试查询单条记录
func TestMySQLQueryOne(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := MySQLUser{Name: "用户A", Email: "usera@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 查询单条记录
	var queriedUser MySQLUser
	err = db.QueryOne(ctx, &queriedUser, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id})
	if err != nil {
		t.Fatalf("查询单个用户失败: %v", err)
	}

	t.Logf("查询到用户: ID=%d, Name=%s, Email=%s", queriedUser.ID, queriedUser.Name, queriedUser.Email)

	// 验证查询结果
	if queriedUser.ID != id || queriedUser.Name != user.Name || queriedUser.Email != user.Email {
		t.Errorf("查询结果不匹配，期望 Name=%s, Email=%s，实际为 Name=%s, Email=%s",
			user.Name, user.Email, queriedUser.Name, queriedUser.Email)
	}

	// 测试记录不存在的情况
	err = db.QueryOne(ctx, &queriedUser, "SELECT * FROM mysql_users WHERE name = ?", []interface{}{"不存在的用户"})
	if err != database.ErrRecordNotFound {
		t.Errorf("查询不存在的记录应返回ErrRecordNotFound, 实际返回: %v", err)
	} else {
		t.Log("正确处理了不存在的记录")
	}
}

// 测试查询多条记录
func TestMySQLQuery(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	users := []MySQLUser{
		{Name: "用户A", Email: "usera@example.com", CreatedAt: now},
		{Name: "用户B", Email: "userb@example.com", CreatedAt: now},
		{Name: "用户C", Email: "userc@example.com", CreatedAt: now},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 查询多条记录
	var queriedUsers []MySQLUser
	err := db.Query(ctx, &queriedUsers, "SELECT * FROM mysql_users ORDER BY id", nil)
	if err != nil {
		t.Fatalf("查询多条记录失败: %v", err)
	}

	t.Logf("查询到 %d 条记录", len(queriedUsers))
	for _, u := range queriedUsers {
		t.Logf("用户: ID=%d, Name=%s, Email=%s", u.ID, u.Name, u.Email)
	}

	// 验证查询结果数量
	if len(queriedUsers) != len(users) {
		t.Errorf("查询结果数量不匹配，期望 %d 条记录，实际为 %d", len(users), len(queriedUsers))
	}
}

// 测试更新记录
func TestMySQLUpdate(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := MySQLUser{Name: "用户B", Email: "userb@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 更新用户
	updatedUser := MySQLUser{
		Name: "用户B(已更新)",
	}

	affected, err := db.Update(ctx, user.TableName(), updatedUser, "id = ?", []interface{}{id})
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	t.Logf("更新成功, 影响行数: %d", affected)

	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证更新
	var updatedRecord MySQLUser
	err = db.QueryOne(ctx, &updatedRecord, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id})
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}

	if updatedRecord.Name != "用户B(已更新)" {
		t.Errorf("更新失败, 期望名称为 '用户B(已更新)', 实际为: %s", updatedRecord.Name)
	} else {
		t.Logf("验证更新成功: ID=%d, Name=%s", updatedRecord.ID, updatedRecord.Name)
	}
}

// 测试删除记录
func TestMySQLDelete(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := MySQLUser{Name: "用户C", Email: "userc@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 删除用户
	affected, err := db.Delete(ctx, user.TableName(), "id = ?", []interface{}{id})
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	t.Logf("删除成功, 影响行数: %d", affected)

	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证删除
	var deletedUser MySQLUser
	err = db.QueryOne(ctx, &deletedUser, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id})
	if err != database.ErrRecordNotFound {
		t.Errorf("删除验证失败, 期望ErrRecordNotFound, 实际: %v", err)
	} else {
		t.Log("验证删除成功, 记录已不存在")
	}
}

// 测试事务提交
func TestMySQLTransactionCommit(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中执行多个操作
	now := time.Now()
	user1 := MySQLUser{Name: "事务用户1", Email: "tx1@example.com", CreatedAt: now}
	user2 := MySQLUser{Name: "事务用户2", Email: "tx2@example.com", CreatedAt: now}

	id1, err := tx.Insert(ctx, user1.TableName(), user1)
	if err != nil {
		tx.Rollback()
		t.Fatalf("事务中插入用户1失败: %v", err)
	}

	id2, err := tx.Insert(ctx, user2.TableName(), user2)
	if err != nil {
		tx.Rollback()
		t.Fatalf("事务中插入用户2失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	t.Logf("事务成功提交, 插入用户ID: %d, %d", id1, id2)

	// 验证两条记录都插入成功
	var count int
	err = db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM mysql_users WHERE email IN (?, ?)",
		[]interface{}{user1.Email, user2.Email})
	if err != nil {
		t.Fatalf("查询事务插入记录数失败: %v", err)
	}

	if count != 2 {
		t.Errorf("事务提交验证失败, 期望记录数2, 实际: %d", count)
	} else {
		t.Log("事务提交验证成功, 两条记录都已插入")
	}
}

// 测试事务回滚
func TestMySQLTransactionRollback(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据
	now := time.Now()
	rollbackUser := MySQLUser{
		Name:      "回滚用户",
		Email:     "rollback@example.com",
		CreatedAt: now,
	}

	_, err = tx.Insert(ctx, rollbackUser.TableName(), rollbackUser)
	if err != nil {
		tx.Rollback()
		t.Fatalf("事务中插入用户失败: %v", err)
	}

	// 回滚事务
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证数据未插入
	var count int
	err = db.QueryOne(ctx, &count, "SELECT COUNT(*) FROM mysql_users WHERE email = ?",
		[]interface{}{rollbackUser.Email})
	if err != nil {
		t.Fatalf("查询回滚记录失败: %v", err)
	}

	if count != 0 {
		t.Errorf("事务回滚验证失败, 期望记录数0, 实际: %d", count)
	} else {
		t.Log("事务回滚验证成功, 记录未被插入")
	}
}

// 测试带选项的查询
func TestMySQLQueryWithOptions(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	users := []MySQLUser{
		{Name: "事务用户1", Email: "tx1@example.com", CreatedAt: now},
		{Name: "事务用户2", Email: "tx2@example.com", CreatedAt: now},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 测试带选项的查询
	var queriedUsers []MySQLUser
	err := db.QueryWithOptions(
		ctx,
		&queriedUsers,
		"SELECT * FROM mysql_users WHERE name LIKE ?",
		[]interface{}{"%事务用户%"},
		database.WithQueryTransaction(true),
	)

	if err != nil {
		t.Fatalf("带选项查询失败: %v", err)
	}

	t.Logf("带选项查询成功, 返回 %d 条记录", len(queriedUsers))
	for _, u := range queriedUsers {
		t.Logf("用户: ID=%d, Name=%s", u.ID, u.Name)
	}

	// 验证查询结果数量
	if len(queriedUsers) != len(users) {
		t.Errorf("查询结果数量不匹配，期望 %d 条记录，实际为 %d", len(users), len(queriedUsers))
	}
}

// 测试WithTx功能
func TestMySQLWithTx(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 使用WithTx函数管理事务
	now := time.Now()
	withTxUser := MySQLUser{
		Name:      "WithTx用户",
		Email:     "withtx@example.com",
		CreatedAt: now,
	}

	var insertedID int64
	err := db.WithTx(nil, func(tx database.Transaction) error {
		// 在事务中执行操作
		id, err := tx.Insert(ctx, withTxUser.TableName(), withTxUser)
		if err != nil {
			return err
		}
		insertedID = id
		return nil // 返回nil表示成功, 事务会自动提交
	})

	if err != nil {
		t.Fatalf("WithTx执行失败: %v", err)
	}

	// 验证记录已插入
	var user MySQLUser
	err = db.QueryOne(ctx, &user, "SELECT * FROM mysql_users WHERE id = ?",
		[]interface{}{insertedID})

	if err != nil {
		t.Fatalf("验证WithTx插入记录失败: %v", err)
	}

	t.Logf("WithTx功能验证成功, 插入用户: ID=%d, Name=%s", user.ID, user.Name)
}
