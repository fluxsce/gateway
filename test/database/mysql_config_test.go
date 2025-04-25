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
	// 从配置文件加载数据库连接
	db, err := database.OpenWithConfigFile("../../configs/database.yaml", "mysql")
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
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS users", nil)
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
	`, nil)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	t.Log("测试表创建成功")
}

// 清理测试表
func cleanupConfigTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS users", nil)
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
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'", nil)
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
	id, err := db.Insert(ctx, user.TableName(), user)
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
		id, err := db.Insert(ctx, user.TableName(), user)
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
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) as total FROM users", nil)
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

	// 先插入测试数据
	now := time.Now()
	users := []ConfigUser{
		{Name: "张三", Email: "zhangsan@example.com", CreatedAt: now},
		{Name: "李四", Email: "lisi@example.com", CreatedAt: now},
		{Name: "王五", Email: "wangwu@example.com", CreatedAt: now},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 查询所有用户
	var queryUsers []ConfigUser
	err := db.Query(ctx, &queryUsers, "SELECT * FROM users ORDER BY id", nil)
	if err != nil {
		t.Fatalf("查询所有用户记录失败: %v", err)
	}

	// 验证结果
	if len(queryUsers) != len(users) {
		t.Errorf("查询结果数量不匹配，期望 %d 条记录，实际为 %d", len(users), len(queryUsers))
	}

	t.Logf("查询到 %d 个用户", len(queryUsers))
	for _, user := range queryUsers {
		t.Logf("用户ID: %d, 姓名: %s, 邮箱: %s", user.ID, user.Name, user.Email)
	}
}

// 测试查询单条记录
func TestConfigQueryOne(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := ConfigUser{Name: "张三", Email: "zhangsan@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 测试查询单条记录
	var singleUser ConfigUser
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM users WHERE id = ?", []interface{}{id})
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
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM users WHERE id = ?", []interface{}{99999})
	if err != database.ErrRecordNotFound {
		t.Errorf("查询不存在的记录应返回 ErrRecordNotFound，实际返回: %v", err)
	} else {
		t.Log("正确处理了不存在的记录查询")
	}
}

// 测试更新操作
func TestConfigUpdate(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := ConfigUser{Name: "张三", Email: "zhangsan@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}
	t.Logf("插入用户成功，ID: %d", id)

	// 准备更新数据 - 确保包含有效的created_at值
	updatedUser := ConfigUser{
		Name:      "张三(已更新)",
		CreatedAt: now, // 使用有效的时间值而不是零值
	}

	// 执行更新 - 使用WHERE条件而不是主键
	affected, err := db.Update(ctx, user.TableName(), updatedUser, "email = ?", []interface{}{user.Email})
	if err != nil {
		t.Fatalf("更新用户失败: %v", err)
	}
	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证更新结果 - 使用更新后的名字查询
	var updatedRecord ConfigUser
	err = db.QueryOne(ctx, &updatedRecord, "SELECT * FROM users WHERE name = ?", []interface{}{"张三(已更新)"})
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}

	if updatedRecord.Name != "张三(已更新)" {
		t.Errorf("更新失败，期望姓名为'张三(已更新)'，实际为: %s", updatedRecord.Name)
	}

	t.Logf("更新用户成功: ID: %d, 新姓名: %s", updatedRecord.ID, updatedRecord.Name)
}

// 测试删除操作
func TestConfigDelete(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	now := time.Now()
	user := ConfigUser{Name: "张三", Email: "zhangsan@example.com", CreatedAt: now}
	id, err := db.Insert(ctx, user.TableName(), user)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}
	t.Logf("插入用户成功，ID: %d", id)

	// 执行删除 - 使用WHERE条件而不是主键
	affected, err := db.Delete(ctx, user.TableName(), "email = ?", []interface{}{user.Email})
	if err != nil {
		t.Fatalf("删除用户失败: %v", err)
	}
	if affected != 1 {
		t.Errorf("期望影响1行记录，实际影响: %d", affected)
	}

	// 验证删除结果
	var deletedUser ConfigUser
	err = db.QueryOne(ctx, &deletedUser, "SELECT * FROM users WHERE email = ?", []interface{}{user.Email})
	if err == nil {
		t.Errorf("删除用户后仍能查询到记录: %+v", deletedUser)
	} else if err != database.ErrRecordNotFound {
		t.Errorf("期望错误为'record not found'，实际为: %v", err)
	}

	t.Logf("删除用户成功: 邮箱: %s", user.Email)
}

// 测试事务提交
func TestConfigTransactionCommit(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据
	now := time.Now()
	txUser := ConfigUser{
		Name:      "事务用户",
		Email:     "transaction@example.com",
		CreatedAt: now,
	}

	id, err := tx.Insert(ctx, txUser.TableName(), txUser)
	if err != nil {
		tx.Rollback()
		t.Fatalf("事务中插入用户失败: %v", err)
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证插入结果
	var insertedUser ConfigUser
	err = db.QueryOne(ctx, &insertedUser, "SELECT * FROM users WHERE id = ?", []interface{}{id})
	if err != nil {
		t.Fatalf("查询事务插入的用户失败: %v", err)
	}

	t.Logf("事务提交成功: 插入用户 ID: %d, 姓名: %s", insertedUser.ID, insertedUser.Name)
}

// 测试事务回滚
func TestConfigTransactionRollback(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	tx, err := db.BeginTx(ctx)
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据
	now := time.Now()
	rollbackUser := ConfigUser{
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
	var insertedUser ConfigUser
	err = db.QueryOne(ctx, &insertedUser, "SELECT * FROM users WHERE email = ?",
		[]interface{}{rollbackUser.Email})

	if err == nil {
		t.Errorf("事务回滚后仍能查询到记录: %+v", insertedUser)
	} else if err != database.ErrRecordNotFound {
		t.Errorf("期望错误为'record not found'，实际为: %v", err)
	}

	t.Log("事务回滚成功: 用户未被插入")
}

// 测试WithTx函数
func TestConfigWithTx(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	now := time.Now()
	withTxUser := ConfigUser{
		Name:      "WithTx用户",
		Email:     "withtx@example.com",
		CreatedAt: now,
	}

	var insertedID int64
	err := db.WithTx(nil, func(tx database.Transaction) error {
		// 在事务函数中执行插入
		id, err := tx.Insert(ctx, withTxUser.TableName(), withTxUser)
		if err != nil {
			return err
		}
		insertedID = id
		return nil
	})

	if err != nil {
		t.Fatalf("WithTx事务执行失败: %v", err)
	}

	// 验证插入结果
	var user ConfigUser
	err = db.QueryOne(ctx, &user, "SELECT * FROM users WHERE id = ?", []interface{}{insertedID})
	if err != nil {
		t.Fatalf("查询WithTx插入的用户失败: %v", err)
	}

	t.Logf("WithTx事务成功: 插入用户 ID: %d, 姓名: %s", user.ID, user.Name)
}

// 验证插入数量
func TestConfigCountInsert(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()

	setupConfigTestTable(t, db)
	//defer cleanupConfigTestTable(t, db)

	ctx := context.Background()

	// 先插入一些测试数据
	now := time.Now()
	users := []ConfigUser{
		{Name: "用户1", Email: "user1@example.com", CreatedAt: now},
		{Name: "用户2", Email: "user2@example.com", CreatedAt: now},
		{Name: "用户3", Email: "user3@example.com", CreatedAt: now},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 查询记录总数 - 使用结构体接收
	type CountResult struct {
		Total int `db:"total"`
	}
	var result CountResult
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) as total FROM users", nil)
	if err != nil {
		t.Fatalf("查询记录总数失败: %v", err)
	}

	if result.Total != len(users) {
		t.Errorf("记录数量不匹配，期望 %d 条记录，实际为 %d", len(users), result.Total)
	} else {
		t.Logf("记录数量验证成功，共 %d 条记录", result.Total)
	}
}
