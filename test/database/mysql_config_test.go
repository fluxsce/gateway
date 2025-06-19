package database

import (
	"context"
	"testing"
	"time"

	"gohub/pkg/database"
	_ "gohub/pkg/database/alldriver" // 导入驱动确保注册
)

// ConfigUser 用于测试的用户结构体
type ConfigUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName 实现Model接口
func (u ConfigUser) TableName() string {
	return "users"
}

// PrimaryKey 实现Model接口
func (u ConfigUser) PrimaryKey() string {
	return "id"
}

// 获取测试数据库连接
func getTestDB(t *testing.T) database.Database {
	// 创建测试数据库配置
	config := &database.DbConfig{
		Name:    "test",
		Enabled: true,
		Driver:  database.DriverMySQL,
		DSN:     "root:datahub@tcp(121.43.231.91:63306)/shangjian_test?charset=utf8mb3&parseTime=True&loc=Local",
	}

	// 打开数据库连接
	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("加载数据库连接失败: %v", err)
	}

	// 验证连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("数据库连接测试失败: %v", err)
	}
	t.Log("数据库连接成功")

	return db
}

// 创建测试表
func setupConfigTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()

	// 先尝试删除表（如果存在）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("删除测试表失败: %v", err)
	}

	// 创建测试表
	_, err = db.Exec(ctx, `
		CREATE TABLE users (
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
func cleanupConfigTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("清理测试表失败: %v", err)
	}
	t.Log("测试表清理成功")
}

// 测试表创建
func TestConfigTableSetup(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	// 验证表是否创建成功
	ctx := context.Background()

	// 使用结构体接收查询结果
	type TableCount struct {
		Count int `db:"count"`
	}
	var result TableCount
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证表创建失败: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("表创建验证失败，期望表存在(count=1)，实际为 %d", result.Count)
	} else {
		t.Log("表创建验证成功")
	}
}

// 测试插入操作
func TestConfigInsert(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 使用当前时间而不是零值
	now := time.Now()

	// 构建测试用户
	user := ConfigUser{
		Name:      "测试用户",
		Email:     "test@example.com",
		CreatedAt: now,
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

// 测试批量插入操作
func TestConfigBatchInsert(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 使用当前时间而不是零值
	now := time.Now()

	// 构建测试用户
	users := []ConfigUser{
		{Name: "张三", Email: "zhangsan@example.com", CreatedAt: now},
		{Name: "李四", Email: "lisi@example.com", CreatedAt: now},
		{Name: "王五", Email: "wangwu@example.com", CreatedAt: now},
	}

	for i, user := range users {
		// 插入记录
		id, err := db.Insert(ctx, user.TableName(), user, true)
		if err != nil {
			t.Fatalf("插入用户记录失败 (索引 %d): %v", i, err)
		}
		if id <= 0 {
			t.Errorf("插入记录返回的ID应该大于0，实际得到: %d", id)
		}
		t.Logf("插入用户 %s 成功，ID: %d", user.Name, id)
	}

	// 验证插入数量
	type CountResult struct {
		Total int `db:"total"`
	}
	var result CountResult
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) as total FROM users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证插入数量失败: %v", err)
	}

	if result.Total != len(users) {
		t.Errorf("批量插入验证失败，期望 %d 条记录，实际为 %d", len(users), result.Total)
	} else {
		t.Logf("批量插入验证成功，共插入 %d 条记录", result.Total)
	}
}

// 测试查询操作
func TestConfigQuery(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入一个用户
	now := time.Now()
	user := ConfigUser{
		Name:      "查询测试用户",
		Email:     "query@example.com",
		CreatedAt: now,
	}

	_, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 查询多条记录
	var users []ConfigUser
	err = db.Query(ctx, &users, "SELECT id, name, email, created_at FROM users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("查询用户列表失败: %v", err)
	}

	if len(users) == 0 {
		t.Errorf("查询结果为空，期望至少有一条记录")
	} else {
		t.Logf("查询成功，找到 %d 条记录", len(users))
		for _, u := range users {
			t.Logf("用户: ID=%d, Name=%s, Email=%s", u.ID, u.Name, u.Email)
		}
	}
}

// 测试单条查询操作
func TestConfigQueryOne(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入一个用户
	now := time.Now()
	user := ConfigUser{
		Name:      "单条查询测试用户",
		Email:     "queryone@example.com",
		CreatedAt: now,
	}

	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 查询单条记录
	var foundUser ConfigUser
	err = db.QueryOne(ctx, &foundUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询单条用户失败: %v", err)
	}

	if foundUser.ID != id {
		t.Errorf("查询的用户ID不匹配，期望 %d，实际 %d", id, foundUser.ID)
	}

	if foundUser.Name != user.Name {
		t.Errorf("查询的用户名不匹配，期望 %s，实际 %s", user.Name, foundUser.Name)
	}

	t.Logf("单条查询成功: ID=%d, Name=%s, Email=%s", foundUser.ID, foundUser.Name, foundUser.Email)
}

// 测试更新操作
func TestConfigUpdate(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入一个用户
	now := time.Now()
	user := ConfigUser{
		Name:      "更新测试用户",
		Email:     "update@example.com",
		CreatedAt: now,
	}

	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 更新用户信息
	user.ID = id
	user.Name = "更新后的用户"
	user.Email = "updated@example.com"

	affected, err := db.Update(ctx, user.TableName(), user, "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}

	if affected != 1 {
		t.Errorf("更新应该影响1行，实际影响 %d 行", affected)
	}

	// 验证更新结果
	var updatedUser ConfigUser
	err = db.QueryOne(ctx, &updatedUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}

	if updatedUser.Name != "更新后的用户" {
		t.Errorf("用户名更新失败，期望 '更新后的用户'，实际 '%s'", updatedUser.Name)
	}

	t.Logf("更新成功: ID=%d, Name=%s, Email=%s", updatedUser.ID, updatedUser.Name, updatedUser.Email)
}

// 测试删除操作
func TestConfigDelete(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入一个用户
	now := time.Now()
	user := ConfigUser{
		Name:      "删除测试用户",
		Email:     "delete@example.com",
		CreatedAt: now,
	}

	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入测试用户失败: %v", err)
	}

	// 删除用户
	affected, err := db.Delete(ctx, user.TableName(), "id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}

	if affected != 1 {
		t.Errorf("删除应该影响1行，实际影响 %d 行", affected)
	}

	// 验证删除结果
	var deletedUser ConfigUser
	err = db.QueryOne(ctx, &deletedUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err == nil {
		t.Errorf("删除后不应该能查到用户，但查到了: %+v", deletedUser)
	} else if err != database.ErrRecordNotFound {
		t.Fatalf("查询删除的用户应该返回 ErrRecordNotFound，实际返回: %v", err)
	}

	t.Log("删除操作验证成功")
}

// 测试事务提交
func TestConfigTransactionCommit(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入用户
	now := time.Now()
	user := ConfigUser{
		Name:      "事务测试用户",
		Email:     "transaction@example.com",
		CreatedAt: now,
	}

	id, err := db.Insert(ctx, user.TableName(), user, false) // autoCommit = false
	if err != nil {
		db.Rollback()
		t.Fatalf("在事务中插入用户失败: %v", err)
	}

	// 提交事务
	err = db.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证用户已插入
	var foundUser ConfigUser
	err = db.QueryOne(ctx, &foundUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("事务提交后查询用户失败: %v", err)
	}

	t.Logf("事务提交成功: ID=%d, Name=%s", foundUser.ID, foundUser.Name)
}

// 测试事务回滚
func TestConfigTransactionRollback(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入用户
	now := time.Now()
	user := ConfigUser{
		Name:      "回滚测试用户",
		Email:     "rollback@example.com",
		CreatedAt: now,
	}

	id, err := db.Insert(ctx, user.TableName(), user, false) // autoCommit = false
	if err != nil {
		db.Rollback()
		t.Fatalf("在事务中插入用户失败: %v", err)
	}

	// 回滚事务
	err = db.Rollback()
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证用户未插入
	var foundUser ConfigUser
	err = db.QueryOne(ctx, &foundUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{id}, true)
	if err == nil {
		t.Errorf("事务回滚后不应该能查到用户，但查到了: %+v", foundUser)
	} else if err != database.ErrRecordNotFound {
		t.Fatalf("查询回滚的用户应该返回 ErrRecordNotFound，实际返回: %v", err)
	}

	t.Log("事务回滚验证成功")
}

// 测试InTx自动管理事务
func TestConfigWithTx(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 使用InTx方法自动管理事务
	var insertedID int64
	err := db.InTx(ctx, &database.TxOptions{
		Isolation: database.IsolationReadCommitted,
		ReadOnly:  false,
	}, func() error {
		// 在事务中插入用户
		now := time.Now()
		user := ConfigUser{
			Name:      "InTx测试用户",
			Email:     "intx@example.com",
			CreatedAt: now,
		}

		id, err := db.Insert(ctx, user.TableName(), user, false) // autoCommit = false
		if err != nil {
			return err
		}
		insertedID = id

		t.Logf("在InTx中插入用户成功: ID=%d", id)
		return nil
	})

	if err != nil {
		t.Fatalf("InTx执行失败: %v", err)
	}

	// 验证用户已插入
	var foundUser ConfigUser
	err = db.QueryOne(ctx, &foundUser, "SELECT id, name, email, created_at FROM users WHERE id = ?", []interface{}{insertedID}, true)
	if err != nil {
		t.Fatalf("InTx事务提交后查询用户失败: %v", err)
	}

	t.Logf("InTx自动管理事务成功: ID=%d, Name=%s", foundUser.ID, foundUser.Name)
}
