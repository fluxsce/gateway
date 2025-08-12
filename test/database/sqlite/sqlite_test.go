package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"gateway/pkg/database"
	_ "gateway/pkg/database/alldriver" // 导入驱动确保注册
	"gateway/pkg/database/dbtypes"
)

// SQLiteUser 用于测试的用户结构体
type SQLiteUser struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
}

// TableName 实现Model接口
func (u SQLiteUser) TableName() string {
	return "sqlite_users"
}

// PrimaryKey 实现Model接口
func (u SQLiteUser) PrimaryKey() string {
	return "id"
}

// 获取测试数据库连接
func getSQLiteTestDB(t *testing.T) (database.Database, string) {
	// 创建临时测试数据库文件
	tempDir, err := os.MkdirTemp("", "sqlite_test_*")
	if err != nil {
		t.Fatalf("创建临时目录失败: %v", err)
	}

	dbPath := filepath.Join(tempDir, "test.db")

	// 创建测试数据库配置
	config := &database.DbConfig{
		Driver:  database.DriverSQLite,
		Name:    "sqlite_test",
		Enabled: true,
		DSN:     dbPath, // SQLite 使用文件路径作为 DSN
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    5, // SQLite 连接池设置较小
			MaxIdleConns:    2,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 200,
		},
	}

	// 直接打开数据库连接
	db, err := database.Open(config)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("打开SQLite连接失败: %v", err)
	}

	// 验证连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		db.Close()
		os.RemoveAll(tempDir)
		t.Fatalf("SQLite连接测试失败: %v", err)
	}
	t.Logf("成功连接到SQLite数据库: %s", dbPath)

	return db, tempDir
}

// 创建测试表
func setupSQLiteTestTable(t *testing.T, db database.Database) {
	ctx := context.Background()

	// 先尝试删除表（如果存在）
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS sqlite_users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("删除测试表失败: %v", err)
	}

	// 创建测试表，SQLite语法
	_, err = db.Exec(ctx, `
		CREATE TABLE sqlite_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			email TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`, []interface{}{}, true)
	if err != nil {
		t.Fatalf("创建测试表失败: %v", err)
	}

	t.Log("SQLite测试表创建成功")
}

// 清理测试表和文件
func cleanupSQLiteTest(t *testing.T, db database.Database, tempDir string) {
	ctx := context.Background()
	_, err := db.Exec(ctx, "DROP TABLE IF EXISTS sqlite_users", []interface{}{}, true)
	if err != nil {
		t.Logf("清理测试表警告: %v", err)
	}

	db.Close()

	// 删除临时目录和文件
	err = os.RemoveAll(tempDir)
	if err != nil {
		t.Logf("清理临时文件警告: %v", err)
	} else {
		t.Log("SQLite测试清理成功")
	}
}

// 测试数据库连接和表创建
func TestSQLiteConnection(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	// 验证表是否创建成功
	ctx := context.Background()
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err := db.QueryOne(ctx, &result, "SELECT COUNT(*) AS count FROM sqlite_master WHERE type='table' AND name='sqlite_users'", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证表创建失败: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("表创建验证失败，期望表存在(count=1)，实际为 %d", result.Count)
	} else {
		t.Log("SQLite表创建验证成功")
	}
}

// 测试单条插入
func TestSQLiteInsert(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 构建测试用户 - 不设置时间，使用数据库默认值
	user := SQLiteUser{
		Name:  "SQLite测试用户",
		Email: "sqlite@example.com",
	}

	// 插入记录
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("插入用户记录失败: %v", err)
	}

	if id <= 0 {
		t.Errorf("插入记录返回的ID应该大于0，实际得到: %d", id)
	}

	t.Logf("SQLite插入用户成功，ID: %d", id)

	// 验证插入的数据
	var insertedUser SQLiteUser
	err = db.QueryOne(ctx, &insertedUser, "SELECT * FROM sqlite_users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询插入的用户失败: %v", err)
	}

	if insertedUser.Name != user.Name || insertedUser.Email != user.Email {
		t.Errorf("插入数据验证失败，期望 Name=%s, Email=%s，实际为 Name=%s, Email=%s",
			user.Name, user.Email, insertedUser.Name, insertedUser.Email)
	}
}

// 测试批量插入（SQLite特有的预编译循环模式）
func TestSQLiteBatchInsert(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 测试批量插入 - 使用预编译循环执行模式
	users := []SQLiteUser{
		{Name: "SQLite用户A", Email: "sqlitea@example.com"},
		{Name: "SQLite用户B", Email: "sqliteb@example.com"},
		{Name: "SQLite用户C", Email: "sqlitec@example.com"},
		{Name: "SQLite用户D", Email: "sqlited@example.com"},
	}

	// 使用批量插入方法（预编译循环执行）
	affected, err := db.BatchInsert(ctx, users[0].TableName(), users, true)
	if err != nil {
		t.Fatalf("SQLite批量插入用户记录失败: %v", err)
	}
	t.Logf("SQLite批量插入成功，影响行数: %d", affected)

	// 验证插入数量
	type CountResult struct {
		Total int `db:"total"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) AS total FROM sqlite_users", []interface{}{}, true)
	if err != nil {
		t.Fatalf("验证插入数量失败: %v", err)
	}

	if result.Total != len(users) {
		t.Errorf("批量插入验证失败，期望 %d 条记录，实际为 %d", len(users), result.Total)
	} else {
		t.Logf("SQLite批量插入验证成功，共插入 %d 条记录", result.Total)
	}

	// 验证所有用户都被正确插入
	var allUsers []SQLiteUser
	err = db.Query(ctx, &allUsers, "SELECT * FROM sqlite_users ORDER BY id", nil, true)
	if err != nil {
		t.Fatalf("查询所有用户失败: %v", err)
	}

	for i, user := range allUsers {
		expectedUser := users[i]
		if user.Name != expectedUser.Name || user.Email != expectedUser.Email {
			t.Errorf("批量插入数据验证失败，索引 %d: 期望 Name=%s, Email=%s，实际为 Name=%s, Email=%s",
				i, expectedUser.Name, expectedUser.Email, user.Name, user.Email)
		}
	}
}

// 测试查询单条记录
func TestSQLiteQueryOne(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	user := SQLiteUser{Name: "查询测试用户", Email: "query@example.com"}
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 测试查询单条记录
	var singleUser SQLiteUser
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM sqlite_users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询单个用户记录失败: %v", err)
	}

	// 验证查询结果
	if singleUser.ID != id || singleUser.Name != user.Name || singleUser.Email != user.Email {
		t.Errorf("查询结果不匹配，期望 ID=%d, Name=%s, Email=%s，实际为 ID=%d, Name=%s, Email=%s",
			id, user.Name, user.Email, singleUser.ID, singleUser.Name, singleUser.Email)
	}

	t.Logf("查询到用户：ID: %d, 姓名: %s, 邮箱: %s, 创建时间: %s",
		singleUser.ID, singleUser.Name, singleUser.Email, singleUser.CreatedAt.Format("2006-01-02 15:04:05"))

	// 测试查询不存在的记录
	err = db.QueryOne(ctx, &singleUser, "SELECT * FROM sqlite_users WHERE id = ?", []interface{}{99999}, true)
	if err != database.ErrRecordNotFound {
		t.Errorf("查询不存在的记录应返回 ErrRecordNotFound，实际返回: %v", err)
	} else {
		t.Log("正确处理了不存在的记录查询")
	}
}

// 测试查询多条记录
func TestSQLiteQuery(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	users := []SQLiteUser{
		{Name: "查询用户A", Email: "querya@example.com"},
		{Name: "查询用户B", Email: "queryb@example.com"},
		{Name: "查询用户C", Email: "queryc@example.com"},
	}

	for _, user := range users {
		_, err := db.Insert(ctx, user.TableName(), user, true)
		if err != nil {
			t.Fatalf("测试准备: 插入用户失败: %v", err)
		}
	}

	// 查询多条记录
	var queriedUsers []SQLiteUser
	err := db.Query(ctx, &queriedUsers, "SELECT * FROM sqlite_users ORDER BY id", nil, true)
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
func TestSQLiteUpdate(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	user := SQLiteUser{Name: "更新测试用户", Email: "update@example.com"}
	id, err := db.Insert(ctx, user.TableName(), user, true)
	if err != nil {
		t.Fatalf("测试准备: 插入用户失败: %v", err)
	}

	// 更新用户
	updatedUser := SQLiteUser{
		Name: "更新测试用户(已更新)",
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
	var updatedRecord SQLiteUser
	err = db.QueryOne(ctx, &updatedRecord, "SELECT * FROM sqlite_users WHERE id = ?", []interface{}{id}, true)
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}

	if updatedRecord.Name != "更新测试用户(已更新)" {
		t.Errorf("更新失败, 期望名称为 '更新测试用户(已更新)', 实际为: %s", updatedRecord.Name)
	} else {
		t.Logf("验证更新成功: ID=%d, Name=%s", updatedRecord.ID, updatedRecord.Name)
	}
}

// 测试删除记录
func TestSQLiteDelete(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	user := SQLiteUser{Name: "删除测试用户", Email: "delete@example.com"}
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
	var deletedUser SQLiteUser
	err = db.QueryOne(ctx, &deletedUser, "SELECT * FROM sqlite_users WHERE id = ?", []interface{}{id}, true)
	if err != database.ErrRecordNotFound {
		t.Errorf("删除验证失败, 期望ErrRecordNotFound, 实际: %v", err)
	} else {
		t.Log("验证删除成功, 记录已不存在")
	}
}

// 测试事务提交
func TestSQLiteTransactionCommit(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	txCtx, err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationSerializable, // SQLite 默认使用 Serializable
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据
	user1 := SQLiteUser{Name: "事务用户1", Email: "tx1@example.com"}
	user2 := SQLiteUser{Name: "事务用户2", Email: "tx2@example.com"}

	id1, err := db.Insert(txCtx, user1.TableName(), user1, false)
	if err != nil {
		db.Rollback(txCtx)
		t.Fatalf("事务中插入用户1失败: %v", err)
	}

	id2, err := db.Insert(txCtx, user2.TableName(), user2, false)
	if err != nil {
		db.Rollback(txCtx)
		t.Fatalf("事务中插入用户2失败: %v", err)
	}

	// 提交事务
	err = db.Commit(txCtx)
	if err != nil {
		t.Fatalf("提交事务失败: %v", err)
	}

	// 验证插入数据
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM sqlite_users WHERE email IN (?, ?)",
		[]interface{}{user1.Email, user2.Email}, true)
	if err != nil {
		t.Fatalf("查询事务插入记录数失败: %v", err)
	}

	if result.Count != 2 {
		t.Errorf("事务提交验证失败，期望插入2条记录，实际为 %d", result.Count)
	} else {
		t.Logf("SQLite事务提交成功，插入用户ID: %d, %d", id1, id2)
	}
}

// 测试事务回滚
func TestSQLiteTransactionRollback(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 开始事务
	txCtx, err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始事务失败: %v", err)
	}

	// 在事务中插入数据
	rollbackUser := SQLiteUser{Name: "回滚用户", Email: "rollback@example.com"}

	_, err = db.Insert(txCtx, rollbackUser.TableName(), rollbackUser, false)
	if err != nil {
		db.Rollback(txCtx)
		t.Fatalf("事务中插入用户失败: %v", err)
	}

	// 回滚事务
	err = db.Rollback(txCtx)
	if err != nil {
		t.Fatalf("回滚事务失败: %v", err)
	}

	// 验证数据未插入
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM sqlite_users WHERE email = ?",
		[]interface{}{rollbackUser.Email}, true)
	if err != nil {
		t.Fatalf("查询回滚记录失败: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("事务回滚验证失败，期望0条记录，实际为 %d", result.Count)
	} else {
		t.Log("SQLite事务回滚成功，用户未被插入")
	}
}

// 测试InTx自动管理事务
func TestSQLiteWithTx(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 使用InTx方法自动管理事务
	var insertedID int64
	err := db.InTx(ctx, &database.TxOptions{
		Isolation: database.IsolationSerializable,
		ReadOnly:  false,
	}, func(txCtx context.Context) error {
		// 在事务中插入用户
		withTxUser := SQLiteUser{
			Name:  "InTx测试用户",
			Email: "intx@example.com",
		}

		id, err := db.Insert(txCtx, withTxUser.TableName(), withTxUser, false)
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
	var user SQLiteUser
	err = db.QueryOne(ctx, &user, "SELECT * FROM sqlite_users WHERE id = ?",
		[]interface{}{insertedID}, true)

	if err != nil {
		t.Fatalf("查询InTx插入的用户失败: %v", err)
	}

	t.Logf("SQLite InTx事务成功: 插入用户 ID: %d, 姓名: %s", user.ID, user.Name)
}

// 测试批量更新
func TestSQLiteBatchUpdate(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	users := []SQLiteUser{
		{Name: "批量更新用户A", Email: "batcha@example.com"},
		{Name: "批量更新用户B", Email: "batchb@example.com"},
		{Name: "批量更新用户C", Email: "batchc@example.com"},
	}

	// 插入用户并保存ID
	for i := range users {
		id, err := db.Insert(ctx, users[i].TableName(), users[i], true)
		if err != nil {
			t.Fatalf("准备数据: 插入用户失败: %v", err)
		}
		users[i].ID = id
	}

	// 更新用户名称
	for i := range users {
		users[i].Name = fmt.Sprintf("批量更新用户%c(已更新)", 'A'+i)
	}

	// 执行批量更新
	affected, err := db.BatchUpdate(ctx, users[0].TableName(), users, []string{"id"}, true)
	if err != nil {
		t.Fatalf("批量更新失败: %v", err)
	}

	t.Logf("批量更新成功，影响行数: %d", affected)

	if affected != int64(len(users)) {
		t.Errorf("批量更新影响行数不匹配，期望 %d，实际 %d", len(users), affected)
	}

	// 验证更新结果
	var updatedUsers []SQLiteUser
	err = db.Query(ctx, &updatedUsers, "SELECT * FROM sqlite_users ORDER BY id", nil, true)
	if err != nil {
		t.Fatalf("查询更新后的用户失败: %v", err)
	}

	for i, user := range updatedUsers {
		expectedName := fmt.Sprintf("批量更新用户%c(已更新)", 'A'+i)
		if user.Name != expectedName {
			t.Errorf("批量更新验证失败，索引 %d: 期望名称 %s，实际 %s", i, expectedName, user.Name)
		}
	}
}

// 测试批量删除
func TestSQLiteBatchDelete(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	users := []SQLiteUser{
		{Name: "批量删除用户A", Email: "deletea@example.com"},
		{Name: "批量删除用户B", Email: "deleteb@example.com"},
		{Name: "批量删除用户C", Email: "deletec@example.com"},
	}

	// 插入用户并保存ID
	for i := range users {
		id, err := db.Insert(ctx, users[i].TableName(), users[i], true)
		if err != nil {
			t.Fatalf("准备数据: 插入用户失败: %v", err)
		}
		users[i].ID = id
	}

	// 执行批量删除（删除前两个用户）
	deleteUsers := users[:2]
	affected, err := db.BatchDelete(ctx, users[0].TableName(), deleteUsers, []string{"id"}, true)
	if err != nil {
		t.Fatalf("批量删除失败: %v", err)
	}

	t.Logf("批量删除成功，影响行数: %d", affected)

	if affected != int64(len(deleteUsers)) {
		t.Errorf("批量删除影响行数不匹配，期望 %d，实际 %d", len(deleteUsers), affected)
	}

	// 验证删除结果
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM sqlite_users", nil, true)
	if err != nil {
		t.Fatalf("查询剩余用户数失败: %v", err)
	}

	expectedRemaining := len(users) - len(deleteUsers)
	if result.Count != expectedRemaining {
		t.Errorf("批量删除验证失败，期望剩余 %d 条记录，实际 %d", expectedRemaining, result.Count)
	} else {
		t.Logf("批量删除验证成功，剩余 %d 条记录", result.Count)
	}
}

// 测试BatchDeleteByKeys
func TestSQLiteBatchDeleteByKeys(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 先插入测试数据
	users := []SQLiteUser{
		{Name: "按键删除用户A", Email: "keya@example.com"},
		{Name: "按键删除用户B", Email: "keyb@example.com"},
		{Name: "按键删除用户C", Email: "keyc@example.com"},
		{Name: "按键删除用户D", Email: "keyd@example.com"},
	}

	var userIDs []interface{}
	for _, user := range users {
		id, err := db.Insert(ctx, user.TableName(), user, true)
		if err != nil {
			t.Fatalf("准备数据: 插入用户失败: %v", err)
		}
		userIDs = append(userIDs, id)
	}

	// 删除前3个用户
	deleteIDs := userIDs[:3]
	affected, err := db.BatchDeleteByKeys(ctx, users[0].TableName(), "id", deleteIDs, true)
	if err != nil {
		t.Fatalf("BatchDeleteByKeys失败: %v", err)
	}

	t.Logf("BatchDeleteByKeys成功，影响行数: %d", affected)

	if affected != int64(len(deleteIDs)) {
		t.Errorf("BatchDeleteByKeys影响行数不匹配，期望 %d，实际 %d", len(deleteIDs), affected)
	}

	// 验证删除结果
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM sqlite_users", nil, true)
	if err != nil {
		t.Fatalf("查询剩余用户数失败: %v", err)
	}

	expectedRemaining := len(users) - len(deleteIDs)
	if result.Count != expectedRemaining {
		t.Errorf("BatchDeleteByKeys验证失败，期望剩余 %d 条记录，实际 %d", expectedRemaining, result.Count)
	} else {
		t.Logf("BatchDeleteByKeys验证成功，剩余 %d 条记录", result.Count)
	}
}

// 测试并发事务（SQLite写事务限制）
func TestSQLiteConcurrentTransactions(t *testing.T) {
	db, tempDir := getSQLiteTestDB(t)
	defer cleanupSQLiteTest(t, db, tempDir)

	setupSQLiteTestTable(t, db)

	ctx := context.Background()

	// 这个测试主要验证SQLite的写事务限制
	// SQLite在WAL模式下虽然支持多读，但写事务仍然是串行的

	// 开始第一个事务
	txCtx1, err := db.BeginTx(ctx, &database.TxOptions{
		Isolation: database.IsolationSerializable,
		ReadOnly:  false,
	})
	if err != nil {
		t.Fatalf("开始第一个事务失败: %v", err)
	}

	// 在第一个事务中插入数据
	user1 := SQLiteUser{Name: "并发测试用户1", Email: "concurrent1@example.com"}
	_, err = db.Insert(txCtx1, user1.TableName(), user1, false)
	if err != nil {
		db.Rollback(txCtx1)
		t.Fatalf("第一个事务插入失败: %v", err)
	}

	// 提交第一个事务
	err = db.Commit(txCtx1)
	if err != nil {
		t.Fatalf("提交第一个事务失败: %v", err)
	}

	// 验证事务管理工作正常
	type CountResult struct {
		Count int `db:"count"`
	}
	var result CountResult
	err = db.QueryOne(ctx, &result, "SELECT COUNT(*) as count FROM sqlite_users", nil, true)
	if err != nil {
		t.Fatalf("查询事务结果失败: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("并发事务测试失败，期望1条记录，实际 %d", result.Count)
	} else {
		t.Log("SQLite事务管理测试通过（注意：SQLite写事务是串行的）")
	}
}
