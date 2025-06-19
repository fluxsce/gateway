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
		Driver:  database.DriverMySQL,
		Name:    "mysql_test",
		Enabled: true,
		DSN:     "root:datahub@tcp(121.43.231.91:63306)/shangjian_test?charset=utf8mb3&parseTime=True&loc=Local",
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
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS mysql_users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("删除测试表失败: %v", err)
	}

	// 创建测试表，使用DEFAULT CURRENT_TIMESTAMP
	_, err = db.Exec(ctx, `
		CREATE TABLE mysql_users (
			id BIGINT AUTO_INCREMENT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) NOT NULL UNIQUE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	t.Log("测试表创建成功")
}

// 清理测试表
func cleanupMySQLTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS mysql_users", []interface{}{}, true)
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
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) AS count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'mysql_users'", []interface{}{}, true)
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

	// 构建测试用户 - 不设置时间，使用数据库默认值
	user := MySQLUser{
		Name:  "测试用户",
		Email: "test@example.com",
	}

	// 插入记录
	id, err := db.Insert(ctx, user.TableName(), user, true)
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

	// 测试真正的批量插入 - 不设置时间，使用数据库默认值
	users := []MySQLUser{
		{Name: "用户A", Email: "usera@example.com"},
		{Name: "用户B", Email: "userb@example.com"},
		{Name: "用户C", Email: "userc@example.com"},
	}

	// 使用真正的批量插入方法
	affected, err := db.BatchInsert(ctx, users[0].TableName(), users, true)
	if err != nil {
		t.Fatalf("批量插入用户记录失败: %v", err)
	}
	t.Logf("批量插入成功，影响行数: %d", affected)

	// 使用结构体接收查询结果
	type CountResult struct {
		Total int `db:"total"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) AS total FROM mysql_users", []interface{}{}, true)
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
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据 - 不设置时间，使用数据库默认值
	user := MySQLUser{Name: "测试用户", Email: "test@example.com"}
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 测试查询单条记录
	var singleUser MySQLUser
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询单个用户记录失败: %v", err)
	}

	// 验证查询结果
	if singleUser.ID != id || singleUser.Name != user.Name || singleUser.Email != user.Email {
		t.Errorf("查询结果不匹配，期望 ID=%d, Name=%s, Email=%s，实际为 ID=%d, Name=%s, Email=%s",
			id, user.Name, user.Email, singleUser.ID, singleUser.Name, singleUser.Email)
	}

	t.Logf("查询到用户：ID: %d, 姓名: %s, 邮箱: %s", singleUser.ID, singleUser.Name, singleUser.Email)

	// 测试查询不存在的记录
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{99999}, true)
	if err != database.ErrRecordNotFound {
		t.Errorf("查询不存在的记录应返回 ErrRecordNotFound，实际返回: %v", err)
	} else {
		t.Log("正确处理了不存在的记录查询")
	}
}

// 测试查询多条记录
func TestMySQLQuery(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据 - 不设置时间，使用数据库默认值
	users := []MySQLUser{
		{Name: "用户A", Email: "usera@example.com"},
		{Name: "用户B", Email: "userb@example.com"},
		{Name: "用户C", Email: "userc@example.com"},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user, true)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 查询多条记录
	var queriedUsers []MySQLUser
	err := db.Query(ctx, &queriedUsers, "SELECT * FROM mysql_users ORDER BY id", nil, true)
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

	// 先插入测试数据 - 不设置时间，使用数据库默认值
	user := MySQLUser{Name: "用户B", Email: "userb@example.com"}
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 更新用户
	updatedUser := MySQLUser{
		Name: "用户B(已更新)",
	}

	affected, err := db.Update(ctx, user.TableName(), updatedUser, "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	t.Logf("更新成功, 影响行数: %d", affected)

	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证更新
	var updatedRecord MySQLUser
	err = db.QueryOne(ctx, &updatedRecord, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id}, true)
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

	// 先插入测试数据 - 不设置时间，使用数据库默认值
	user := MySQLUser{Name: "用户C", Email: "userc@example.com"}
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 删除用户
	affected, err := db.Delete(ctx, user.TableName(), "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	t.Logf("删除成功, 影响行数: %d", affected)

	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证删除
	var deletedUser MySQLUser
	err = db.QueryOne(ctx, &deletedUser, "SELECT * FROM mysql_users WHERE id = ?", []interface{}{id}, true)
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
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据 - 不设置时间，使用数据库默认值
	user1 := MySQLUser{Name: "事务用户1", Email: "tx1@example.com"}
	user2 := MySQLUser{Name: "事务用户2", Email: "tx2@example.com"}

	id1, err := db.Insert(ctx, user1.TableName(), user1, false)
	if err != nil {
		db.Rollback()
		t.Fatalf("事务中插入用户1失败: %v", err)
	}

	id2, err := db.Insert(ctx, user2.TableName(), user2, false)
	if err != nil {
		db.Rollback()
		t.Fatalf("事务中插入用户2失败: %v", err)
	}

	// 提交事务
	err = db.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证插入数据
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM mysql_users WHERE email IN (?, ?)",
		[]interface{}{user1.Email, user2.Email}, true)
	if err != nil {
		t.Fatalf("查询事务插入记录数失败: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("事务提交验证失败，期望插入2条记录，实际为 %d", result.Count)
	} else {
		t.Logf("事务提交成功，插入用户ID: %d, %d", id1, id2)
	}
}

// 测试事务回滚
func TestMySQLTransactionRollback(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据 - 不设置时间，使用数据库默认值
	rollbackUser := MySQLUser{Name: "回滚用户", Email: "rollback@example.com"}

	_, err = db.Insert(ctx, rollbackUser.TableName(), rollbackUser, false)
	if err != nil {
		db.Rollback()
		t.Fatalf("事务中插入用户失败: %v", err)
	}

	// 回滚事务
	err = db.Rollback()
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证数据未插入
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM mysql_users WHERE email = ?",
		[]interface{}{rollbackUser.Email}, true)
	if err != nil {
		t.Fatalf("查询回滚记录失败: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("事务回滚验证失败，期望0条记录，实际为 %d", result.Count)
	} else {
		t.Log("事务回滚成功，用户未被插入")
	}
}

// 测试InTx自动管理事务
func TestMySQLWithTx(t *testing.T) {
	db := getMySQLTestDB(t)
	defer db.Close()

	setupMySQLTestTable(t, db)
	//defer cleanupMySQLTestTable(t, db)

	ctx := context.Background()

	// 使用InTx方法自动管理事务
	var insertedID int64
	err := db.InTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	}, func() error {
		// 在事务中插入用户 - 不设置时间，使用数据库默认值
		withTxUser := MySQLUser{
			Name:  "InTx测试用户",
			Email: "intx@example.com",
		}

		id, err := db.Insert(ctx, withTxUser.TableName(), withTxUser, false)
		if err != nil {
			return err
		}
		insertedID = id
		return nil
	})

	if err != nil {
		t.Fatalf("InTx执行失败: %v", err)
	}

	// 验证用户已插入
	var user MySQLUser
	err = db.QueryOne(ctx, &user, "SELECT * FROM mysql_users WHERE id = ?",
		[]interface{}{insertedID}, true)

	if err != nil {
		t.Fatalf("查询InTx插入的用户失败: %v", err)
	}

	t.Logf("InTx事务成功: 插入用户 ID: %d, 姓名: %s", user.ID, user.Name)
}
