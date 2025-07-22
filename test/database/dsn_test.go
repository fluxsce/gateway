package database

import (
	"testing"

	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dsn"

	"github.com/stretchr/testify/assert"
)

// TestGenerateSQLite 测试SQLite DSN生成
func TestGenerateSQLite(t *testing.T) {
	tests := []struct {
		name     string
		config   *dbtypes.DbConfig
		expected string
		hasError bool
	}{
		{
			name: "内存数据库",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverSQLite,
				Connection: dbtypes.ConnectionConfig{
					Database: ":memory:",
				},
			},
			expected: ":memory:",
			hasError: false,
		},
		{
			name: "空数据库名默认内存",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverSQLite,
				Connection: dbtypes.ConnectionConfig{
					Database: "",
				},
			},
			expected: ":memory:",
			hasError: false,
		},
		{
			name: "简单数据库名自动添加扩展名",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverSQLite,
				Connection: dbtypes.ConnectionConfig{
					Database: "testdb",
				},
			},
			expected: "file:./testdb.db?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1&_busy_timeout=5000",
			hasError: false,
		},
		{
			name: "完整文件路径",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverSQLite,
				Connection: dbtypes.ConnectionConfig{
					Database: "./data/app.db",
				},
			},
			expected: "file:./data/app.db?cache=shared&mode=rwc&_journal_mode=WAL&_synchronous=NORMAL&_foreign_keys=1&_busy_timeout=5000",
			hasError: false,
		},
		{
			name: "已有DSN直接返回",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverSQLite,
				DSN:    "custom:memory:",
				Connection: dbtypes.ConnectionConfig{
					Database: "should_not_use",
				},
			},
			expected: "custom:memory:",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsn.Generate(tt.config)

			if tt.hasError && err == nil {
				t.Errorf("期望有错误，但没有错误")
				return
			}

			if !tt.hasError && err != nil {
				t.Errorf("不期望有错误，但得到错误: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("期望DSN: %s, 实际得到: %s", tt.expected, result)
			}
		})
	}

	// 测试自定义busy_timeout的SQLite DSN
	config := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Database:    "custom.db",
			BusyTimeout: 10000, // 10秒
		},
	}

	dsn, err := dsn.GenerateSQLite(config)
	assert.NoError(t, err)

	// 验证DSN包含自定义的busy_timeout
	assert.Contains(t, dsn, "_busy_timeout=10000")
	assert.Contains(t, dsn, "file:./custom.db")
	assert.Contains(t, dsn, "cache=shared")
	assert.Contains(t, dsn, "_journal_mode=WAL")
}

// TestGenerateOracle 测试Oracle DSN生成
func TestGenerateOracle(t *testing.T) {
	tests := []struct {
		name     string
		config   *dbtypes.DbConfig
		expected string
		hasError bool
	}{
		{
			name: "基本Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Port:     1521,
					Username: "hr",
					Password: "password",
					Database: "XEPDB1",
				},
			},
			expected: "hr/password@localhost:1521/XEPDB1?CONNECTION_TIMEOUT=30&READ_TIMEOUT=30&WRITE_TIMEOUT=30&NLS_LANG=AMERICAN_AMERICA.UTF8&PREFETCH_ROWS=500&LOB_PREFETCH_SIZE=4096",
			hasError: false,
		},
		{
			name: "默认端口Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "oracle-server",
					Username: "scott",
					Password: "tiger",
					Database: "XE",
				},
			},
			expected: "scott/tiger@oracle-server:1521/XE?CONNECTION_TIMEOUT=30&READ_TIMEOUT=30&WRITE_TIMEOUT=30&NLS_LANG=AMERICAN_AMERICA.UTF8&PREFETCH_ROWS=500&LOB_PREFETCH_SIZE=4096",
			hasError: false,
		},
		{
			name: "缺少Host的Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Username: "hr",
					Password: "password",
					Database: "XEPDB1",
				},
			},
			expected: "",
			hasError: true,
		},
		{
			name: "缺少Username的Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Password: "password",
					Database: "XEPDB1",
				},
			},
			expected: "",
			hasError: true,
		},
		{
			name: "缺少Password的Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Username: "hr",
					Database: "XEPDB1",
				},
			},
			expected: "",
			hasError: true,
		},
		{
			name: "缺少Database的Oracle配置",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Username: "hr",
					Password: "password",
				},
			},
			expected: "",
			hasError: true,
		},
		{
			name: "已有DSN直接返回",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				DSN:    "hr/password@custom:1521/CUSTOM",
				Connection: dbtypes.ConnectionConfig{
					Host:     "should_not_use",
					Database: "SHOULD_NOT_USE",
				},
			},
			expected: "hr/password@custom:1521/CUSTOM",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := dsn.Generate(tt.config)

			if tt.hasError && err == nil {
				t.Errorf("期望有错误，但没有错误")
				return
			}

			if !tt.hasError && err != nil {
				t.Errorf("不期望有错误，但得到错误: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("期望DSN: %s, 实际得到: %s", tt.expected, result)
			}
		})
	}

	// 测试带自定义超时配置的Oracle DSN
	config := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:                    "oracle.example.com",
			Port:                    1522,
			Username:                "oracleuser",
			Password:                "oraclepass",
			Database:                "ORCL",
			OracleConnectionTimeout: 60,
			OracleReadTimeout:       45,
			OracleWriteTimeout:      30,
		},
	}

	dsn, err := dsn.GenerateOracle(config)
	assert.NoError(t, err)

	// 验证DSN包含自定义超时参数
	assert.Contains(t, dsn, "CONNECTION_TIMEOUT=60")
	assert.Contains(t, dsn, "READ_TIMEOUT=45")
	assert.Contains(t, dsn, "WRITE_TIMEOUT=30")

	// 验证基本连接信息
	assert.Contains(t, dsn, "oracleuser/oraclepass@oracle.example.com:1522/ORCL")
}

// TestGenerateOracleWithSID 测试Oracle SID方式DSN生成
func TestGenerateOracleWithSID(t *testing.T) {
	config := &dbtypes.DbConfig{
		Driver: dbtypes.DriverOracle,
		Connection: dbtypes.ConnectionConfig{
			Host:     "oracle-server",
			Port:     1521,
			Username: "scott",
			Password: "tiger",
		},
	}

	expected := "scott/tiger@oracle-server:1521:XE"

	result, err := dsn.GenerateOracleWithSID(config, "XE")
	if err != nil {
		t.Fatalf("不期望有错误，但得到错误: %v", err)
	}

	if result != expected {
		t.Errorf("期望DSN: %s, 实际得到: %s", expected, result)
	}
}

// TestGenerateOracleWithSIDErrors 测试Oracle SID错误情况
func TestGenerateOracleWithSIDErrors(t *testing.T) {
	tests := []struct {
		name   string
		config *dbtypes.DbConfig
		sid    string
	}{
		{
			name: "缺少SID",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Username: "hr",
					Password: "password",
				},
			},
			sid: "",
		},
		{
			name: "缺少Host",
			config: &dbtypes.DbConfig{
				Driver: dbtypes.DriverOracle,
				Connection: dbtypes.ConnectionConfig{
					Username: "hr",
					Password: "password",
				},
			},
			sid: "XE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := dsn.GenerateOracleWithSID(tt.config, tt.sid)
			if err == nil {
				t.Errorf("期望有错误，但没有错误")
			}
		})
	}
}

// TestValidateDSN 测试DSN验证功能
func TestValidateDSN(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		dsn      string
		hasError bool
	}{
		{
			name:     "有效的MySQL DSN",
			driver:   dbtypes.DriverMySQL,
			dsn:      "user:pass@tcp(localhost:3306)/db",
			hasError: false,
		},
		{
			name:     "无效的MySQL DSN",
			driver:   dbtypes.DriverMySQL,
			dsn:      "user:pass@localhost:3306/db",
			hasError: true,
		},
		{
			name:     "有效的PostgreSQL DSN",
			driver:   dbtypes.DriverPostgreSQL,
			dsn:      "postgresql://user:pass@localhost:5432/db",
			hasError: false,
		},
		{
			name:     "无效的PostgreSQL DSN",
			driver:   dbtypes.DriverPostgreSQL,
			dsn:      "user:pass@localhost:5432/db",
			hasError: true,
		},
		{
			name:     "有效的SQLite内存DSN",
			driver:   dbtypes.DriverSQLite,
			dsn:      ":memory:",
			hasError: false,
		},
		{
			name:     "有效的SQLite文件DSN",
			driver:   dbtypes.DriverSQLite,
			dsn:      "file:test.db",
			hasError: false,
		},
		{
			name:     "有效的SQLite简单DSN",
			driver:   dbtypes.DriverSQLite,
			dsn:      "./test.db",
			hasError: false,
		},
		{
			name:     "可能无效的SQLite DSN",
			driver:   dbtypes.DriverSQLite,
			dsn:      "invaliddsn",
			hasError: true,
		},
		{
			name:     "有效的Oracle DSN",
			driver:   dbtypes.DriverOracle,
			dsn:      "user/pass@localhost:1521/service",
			hasError: false,
		},
		{
			name:     "无效的Oracle DSN",
			driver:   dbtypes.DriverOracle,
			dsn:      "user:pass:localhost:1521:service",
			hasError: true,
		},
		{
			name:     "空DSN",
			driver:   dbtypes.DriverMySQL,
			dsn:      "",
			hasError: true,
		},
		{
			name:     "不支持的驱动",
			driver:   "unsupported",
			dsn:      "some_dsn",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dsn.ValidateDSN(tt.driver, tt.dsn)

			if tt.hasError && err == nil {
				t.Errorf("期望有错误，但没有错误")
			}

			if !tt.hasError && err != nil {
				t.Errorf("不期望有错误，但得到错误: %v", err)
			}
		})
	}
}

// TestMySQLCompatibility 测试MySQL DSN生成兼容性 (确保现有功能不受影响)
func TestMySQLCompatibility(t *testing.T) {
	config := &dbtypes.DbConfig{
		Driver: dbtypes.DriverMySQL,
		Connection: dbtypes.ConnectionConfig{
			Host:      "localhost",
			Port:      3306,
			Username:  "root",
			Password:  "password",
			Database:  "testdb",
			Charset:   "utf8mb4",
			ParseTime: true,
			Loc:       "Local",
		},
	}

	expected := "root:password@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"

	result, err := dsn.Generate(config)
	if err != nil {
		t.Fatalf("不期望有错误，但得到错误: %v", err)
	}

	if result != expected {
		t.Errorf("期望DSN: %s, 实际得到: %s", expected, result)
	}
}

// TestPostgreSQLCompatibility 测试PostgreSQL DSN生成兼容性 (确保现有功能不受影响)
func TestPostgreSQLCompatibility(t *testing.T) {
	config := &dbtypes.DbConfig{
		Driver: dbtypes.DriverPostgreSQL,
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "postgres",
			Password: "password",
			Database: "testdb",
			SSLMode:  "disable",
		},
	}

	expected := "postgresql://postgres:password@localhost:5432/testdb?sslmode=disable"

	result, err := dsn.Generate(config)
	if err != nil {
		t.Fatalf("不期望有错误，但得到错误: %v", err)
	}

	if result != expected {
		t.Errorf("期望DSN: %s, 实际得到: %s", expected, result)
	}
}

// TestMySQL 中添加超时测试
func TestGenerateMySQL(t *testing.T) {
	// 测试带超时配置的MySQL DSN
	config := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:                "127.0.0.1",
			Port:                3307,
			Username:            "testuser",
			Password:            "testpass",
			Database:            "testdb",
			Charset:             "utf8mb4",
			ParseTime:           true,
			Loc:                 "Asia/Shanghai",
			MySQLConnectTimeout: 10,
			MySQLReadTimeout:    20,
			MySQLWriteTimeout:   15,
		},
	}

	dsn, err := dsn.GenerateMySQL(config)
	assert.NoError(t, err)

	// 验证DSN包含超时参数
	assert.Contains(t, dsn, "timeout=10s")
	assert.Contains(t, dsn, "readTimeout=20s")
	assert.Contains(t, dsn, "writeTimeout=15s")
	assert.Contains(t, dsn, "charset=utf8mb4")
	assert.Contains(t, dsn, "parseTime=True")
	assert.Contains(t, dsn, "loc=Asia%2FShanghai") // URL编码

	// 验证完整格式
	expected := "testuser:testpass@tcp(127.0.0.1:3307)/testdb"
	assert.Contains(t, dsn, expected)
}

// TestPostgreSQL 中添加超时测试
func TestGeneratePostgreSQL(t *testing.T) {
	// 测试带超时配置的PostgreSQL DSN
	config := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:                       "localhost",
			Port:                       5433,
			Username:                   "postgres",
			Password:                   "password",
			Database:                   "testdb",
			SSLMode:                    "require",
			PostgreSQLConnectTimeout:   15,
			PostgreSQLStatementTimeout: 60,
		},
	}

	dsn, err := dsn.GeneratePostgreSQL(config)
	assert.NoError(t, err)

	// 验证DSN包含超时参数
	assert.Contains(t, dsn, "connect_timeout=15")
	assert.Contains(t, dsn, "statement_timeout=60s")
	assert.Contains(t, dsn, "sslmode=require")

	// 验证完整格式
	expected := "postgresql://postgres:password@localhost:5433/testdb"
	assert.Contains(t, dsn, expected)
}

// TestTimeoutDefaults 测试默认超时配置
func TestTimeoutDefaults(t *testing.T) {
	// MySQL默认超时
	mysqlConfig := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Username: "user",
			Password: "pass",
			Database: "db",
		},
	}

	mysqlDSN, err := dsn.GenerateMySQL(mysqlConfig)
	assert.NoError(t, err)
	// 没有配置超时参数时，DSN中不应包含timeout参数
	assert.NotContains(t, mysqlDSN, "timeout=")
	assert.NotContains(t, mysqlDSN, "readTimeout=")
	assert.NotContains(t, mysqlDSN, "writeTimeout=")

	// Oracle默认超时
	oracleConfig := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Username: "user",
			Password: "pass",
			Database: "ORCL",
		},
	}

	oracleDSN, err := dsn.GenerateOracle(oracleConfig)
	assert.NoError(t, err)
	// Oracle有默认超时配置
	assert.Contains(t, oracleDSN, "CONNECTION_TIMEOUT=30")
	assert.Contains(t, oracleDSN, "READ_TIMEOUT=30")
	assert.Contains(t, oracleDSN, "WRITE_TIMEOUT=30")

	// SQLite默认超时
	sqliteConfig := &dbtypes.DbConfig{
		Connection: dbtypes.ConnectionConfig{
			Database: "test.db",
		},
	}

	sqliteDSN, err := dsn.GenerateSQLite(sqliteConfig)
	assert.NoError(t, err)
	// SQLite有默认的busy_timeout
	assert.Contains(t, sqliteDSN, "_busy_timeout=5000")
}

func TestOracleDSNGeneration(t *testing.T) {
	// 测试服务名方式
	config := &dbtypes.DbConfig{
		Driver: dbtypes.DriverOracle,
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Port:     1521,
			Username: "system",
			Password: "oracle",
			Database: "ORCLPDB", // 作为服务名使用
		},
	}

	// 生成DSN
	dsnStr, err := dsn.Generate(config)
	if err != nil {
		t.Fatalf("生成Oracle DSN失败: %v", err)
	}

	// 验证DSN格式
	expectedPrefix := "oracle://system:oracle@localhost:1521/ORCLPDB"
	if dsnStr[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("期望DSN前缀为 '%s'，实际为 '%s'", expectedPrefix, dsnStr[:len(expectedPrefix)])
	}

	// 测试SID方式
	config = &dbtypes.DbConfig{
		Driver: dbtypes.DriverOracle,
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Port:     1521,
			Username: "system",
			Password: "oracle",
			SID:      "ORCL",
			UseSID:   true,
		},
	}

	// 生成DSN
	dsnStr, err = dsn.Generate(config)
	if err != nil {
		t.Fatalf("生成Oracle SID DSN失败: %v", err)
	}

	// 验证DSN格式
	expectedSIDPrefix := "oracle://system:oracle@localhost:1521?sid=ORCL"
	if dsnStr[:len(expectedSIDPrefix)] != expectedSIDPrefix {
		t.Errorf("期望SID DSN前缀为 '%s'，实际为 '%s'", expectedSIDPrefix, dsnStr[:len(expectedSIDPrefix)])
	}
}
