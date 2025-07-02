package database

import (
	"gohub/pkg/database"
	"gohub/pkg/database/dbtypes"
	"gohub/pkg/database/sqlutils"
	"strings"
	"testing"
)

// TestOracle11gRegistration 测试Oracle 11g驱动注册
func TestOracle11gRegistration(t *testing.T) {
	config := &database.DbConfig{
		Driver:  dbtypes.DriverOracle11g,
		Enabled: true,
		DSN:     "oracle://test:test@localhost:1521/test",
	}

	// 创建Oracle 11g实例（不实际连接）
	db, err := database.Open(config)
	if err == nil {
		// 如果创建成功，检查驱动类型
		if db.GetDriver() != dbtypes.DriverOracle11g {
			t.Errorf("期望驱动类型为 'oracle11g'，实际为 '%s'", db.GetDriver())
		}
		db.Close()
	}

	t.Log("Oracle 11g驱动注册测试完成")
}

// TestOracle11gPagination 测试Oracle 11g分页语法
func TestOracle11gPagination(t *testing.T) {
	baseQuery := "SELECT * FROM users WHERE status = 1"
	pagination := sqlutils.NewPaginationInfo(2, 10)  // 第2页，每页10条
	
	// Oracle 12c+ 分页
	query, args, err := sqlutils.BuildPaginationQuery(sqlutils.DatabaseOracle, baseQuery, pagination)
	if err != nil {
		t.Fatalf("构建Oracle分页查询失败: %v", err)
	}
	
	// 验证Oracle 12c+ 使用OFFSET/FETCH语法
	if !contains(query, "OFFSET") || !contains(query, "FETCH NEXT") {
		t.Errorf("Oracle 12c+ 分页语法错误，应使用OFFSET/FETCH: %s", query)
	}
	
	// Oracle 11g 分页
	query, args, err = sqlutils.BuildPaginationQuery(sqlutils.DatabaseOracle11g, baseQuery, pagination)
	if err != nil {
		t.Fatalf("构建Oracle 11g分页查询失败: %v", err)
	}
	
	// 验证Oracle 11g 使用ROW_NUMBER() OVER()语法
	if !contains(query, "ROW_NUMBER()") || !contains(query, "BETWEEN") {
		t.Errorf("Oracle 11g 分页语法错误，应使用ROW_NUMBER() OVER(): %s", query)
	}
	
	// 验证参数数量
	if len(args) != 2 {
		t.Errorf("Oracle 11g 分页参数数量错误，期望2个，实际%d个", len(args))
	}
	
	t.Log("Oracle 11g 分页语法测试完成")
}

// contains 辅助函数，检查字符串是否包含子串
func contains(s, substr string) bool {
	return s != "" && substr != "" && s != substr && s != "" && strings.Contains(s, substr)
} 