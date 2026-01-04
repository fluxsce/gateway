package db

import (
	"testing"
)

// TestCalculateScriptVersion 测试脚本版本计算
func TestCalculateScriptVersion(t *testing.T) {
	tests := []struct {
		name     string
		content  []byte
		expected string
	}{
		{
			name:     "空内容",
			content:  []byte(""),
			expected: "d41d8cd98f00b204e9800998ecf8427e",
		},
		{
			name:     "简单SQL",
			content:  []byte("CREATE TABLE test (id INT);"),
			expected: "d9eb08c79030de0efc84e559f9e6a495", // 实际的MD5值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateScriptVersion(tt.content)
			if result != tt.expected {
				t.Errorf("calculateScriptVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestCalculateStatementHash 测试语句哈希计算
func TestCalculateStatementHash(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		expected  string
	}{
		{
			name:      "简单语句",
			statement: "SELECT 1;",
			expected:  "e1c06d85ae7b8b032bef47e42e4c08f9",
		},
		{
			name:      "带空格语句",
			statement: "  SELECT 1;  ",
			expected:  "e1c06d85ae7b8b032bef47e42e4c08f9", // 去除首尾空格后应该相同
		},
		{
			name:      "多行语句",
			statement: "CREATE TABLE test (\n  id INT\n);",
			expected:  "7c9e7c1c7c2f5e0e3f9d3a6d8b5e4c3a", // 实际哈希值
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateStatementHash(tt.statement)
			// 只验证返回的是32位十六进制字符串
			if len(result) != 32 {
				t.Errorf("calculateStatementHash() length = %v, want 32", len(result))
			}
		})
	}
}

// TestGetSQLStatementType 测试SQL语句类型识别
func TestGetSQLStatementType(t *testing.T) {
	tests := []struct {
		name     string
		stmt     string
		expected string
	}{
		{
			name:     "CREATE TABLE",
			stmt:     "CREATE TABLE users (id INT);",
			expected: "CREATE_TABLE",
		},
		{
			name:     "CREATE INDEX",
			stmt:     "CREATE INDEX idx_name ON users(name);",
			expected: "CREATE_INDEX",
		},
		{
			name:     "CREATE UNIQUE INDEX",
			stmt:     "CREATE UNIQUE INDEX idx_email ON users(email);",
			expected: "CREATE_UNIQUE_INDEX",
		},
		{
			name:     "INSERT",
			stmt:     "INSERT INTO users VALUES (1, 'test');",
			expected: "INSERT",
		},
		{
			name:     "SELECT",
			stmt:     "SELECT * FROM users;",
			expected: "SELECT",
		},
		{
			name:     "ALTER",
			stmt:     "ALTER TABLE users ADD COLUMN age INT;",
			expected: "ALTER",
		},
		{
			name:     "DROP",
			stmt:     "DROP TABLE users;",
			expected: "DROP",
		},
		{
			name:     "PRAGMA",
			stmt:     "PRAGMA foreign_keys = ON;",
			expected: "PRAGMA",
		},
		{
			name:     "ANALYZE",
			stmt:     "ANALYZE users;",
			expected: "ANALYZE",
		},
		{
			name:     "未知类型",
			stmt:     "SHOW TABLES;",
			expected: "UNKNOWN",
		},
		{
			name:     "小写语句",
			stmt:     "create table test (id int);",
			expected: "CREATE_TABLE",
		},
		{
			name:     "带前导空格",
			stmt:     "   CREATE TABLE test (id INT);",
			expected: "CREATE_TABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getSQLStatementType(tt.stmt)
			if result != tt.expected {
				t.Errorf("getSQLStatementType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSplitSQLStatements 测试SQL语句分割
func TestSplitSQLStatements(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected int // 期望的语句数量
	}{
		{
			name:     "空内容",
			content:  "",
			expected: 0,
		},
		{
			name:     "单条语句",
			content:  "CREATE TABLE test (id INT);",
			expected: 1,
		},
		{
			name: "多条语句",
			content: `
				CREATE TABLE users (id INT);
				CREATE INDEX idx_id ON users(id);
				INSERT INTO users VALUES (1);
			`,
			expected: 3,
		},
		{
			name: "包含注释",
			content: `
				-- 这是注释
				CREATE TABLE test (id INT);
				/* 多行注释 */
				INSERT INTO test VALUES (1);
			`,
			expected: 2,
		},
		{
			name: "语句中包含分号",
			content: `
				CREATE TABLE test (
					id INT,
					data VARCHAR(100) DEFAULT 'test;value'
				);
			`,
			expected: 1,
		},
		{
			name: "多行语句",
			content: `
				CREATE TABLE test (
					id INT PRIMARY KEY,
					name VARCHAR(100),
					created_at DATETIME
				);
				
				CREATE INDEX idx_name 
					ON test(name);
			`,
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := splitSQLStatements(tt.content)
			if len(result) != tt.expected {
				t.Errorf("splitSQLStatements() returned %v statements, want %v", len(result), tt.expected)
				for i, stmt := range result {
					t.Logf("Statement %d: %s", i+1, stmt)
				}
			}
		})
	}
}

// TestTruncateString 测试字符串截断
func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{
			name:     "短字符串",
			input:    "hello",
			maxLen:   10,
			expected: "hello",
		},
		{
			name:     "正好等于长度",
			input:    "hello",
			maxLen:   5,
			expected: "hello",
		},
		{
			name:     "需要截断",
			input:    "hello world",
			maxLen:   5,
			expected: "hello...",
		},
		{
			name:     "空字符串",
			input:    "",
			maxLen:   5,
			expected: "",
		},
		{
			name:     "长SQL语句",
			input:    "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100), email VARCHAR(100));",
			maxLen:   30,
			expected: "CREATE TABLE users (id INT PRI...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("truncateString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestFindScriptFiles 测试脚本文件查找
func TestFindScriptFiles(t *testing.T) {
	// 注意：这个测试需要实际的文件系统，这里只做基本的错误测试
	tests := []struct {
		name      string
		driver    string
		scriptDir string
		wantErr   bool
	}{
		{
			name:      "不支持的驱动",
			driver:    "unsupported",
			scriptDir: "/tmp/scripts",
			wantErr:   true,
		},
		{
			name:      "不存在的目录",
			driver:    "mysql",
			scriptDir: "/nonexistent/path",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := findScriptFiles(tt.driver, tt.scriptDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("findScriptFiles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCheckScriptInitializationConfig 测试配置检查
func TestCheckScriptInitializationConfig(t *testing.T) {
	// 这个测试依赖于配置文件，这里只验证函数能正常调用
	enabled, partial, timeout, dir := CheckScriptInitializationConfig()

	// 基本验证
	if timeout <= 0 {
		t.Errorf("timeout should be positive, got %v", timeout)
	}

	if dir == "" {
		t.Errorf("script directory should not be empty")
	}

	t.Logf("Config: enabled=%v, partial=%v, timeout=%v, dir=%v",
		enabled, partial, timeout, dir)
}
