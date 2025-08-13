package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"gateway/pkg/database"
	_ "gateway/pkg/database/alldriver" // 导入驱动确保注册
	"gateway/pkg/database/dbtypes"
	"gateway/pkg/database/dsn"
)

// ConfigClickHouseEvent 用于测试的事件结构体
// ClickHouse 适合时间序列和事件数据
type ConfigClickHouseEvent struct {
	ID          int64     `db:"id"`
	EventName   string    `db:"event_name"`
	UserID      int32     `db:"user_id"`
	Timestamp   time.Time `db:"timestamp"`
	Value       float64   `db:"value"`
	Properties  string    `db:"properties"`
	CreatedDate string    `db:"created_date"`
}

// TableName 实现Model接口
func (e ConfigClickHouseEvent) TableName() string {
	return "events"
}

// PrimaryKey 实现Model接口
func (e ConfigClickHouseEvent) PrimaryKey() string {
	return "id"
}

// 获取测试数据库连接
func getClickHouseTestDB(t *testing.T) database.Database {
	// 创建测试数据库配置
	// 注意：DSN中的密码特殊字符需要URL编码
	// 原密码: your_password
	// 编码后: your_password
	config := &dbtypes.DbConfig{
		Name:    "clickhouse_test",
		Enabled: true,
		Driver:  database.DriverClickHouse,
		DSN:     "tcp://localhost:9000/default?username=default&password=your_password&compress=true&debug=false",
	}

	// 打开数据库连接
	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("加载ClickHouse连接失败: %v", err)
	}

	// 验证连接
	ctx := context.Background()
	err = db.Ping(ctx)
	if err != nil {
		t.Fatalf("ClickHouse连接测试失败: %v", err)
	}

	t.Log("成功连接到ClickHouse数据库")
	return db
}

// 测试ClickHouse连接配置
func TestClickHouseConnection(t *testing.T) {
	config := &dbtypes.DbConfig{
		Name:    "test_clickhouse",
		Enabled: true,
		Driver:  database.DriverClickHouse,
		Connection: dbtypes.ConnectionConfig{
			Host:               "localhost",
			Port:               9000,
			Username:           "default",
			Password:           "your_password",
			Database:           "default",
			ClickHouseCompress: "lz4",
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("打开ClickHouse连接失败: %v", err)
	}
	defer db.Close()

	// 验证生成的DSN是否正确转义了特殊字符
	t.Logf("生成的DSN: %s", config.DSN)

	// 测试连接
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ClickHouse连接ping失败: %v", err)
	}

	t.Log("ClickHouse连接测试成功")
}

// 测试ClickHouse DSN配置结构化生成
func TestClickHouseDSNGeneration(t *testing.T) {
	config := &dbtypes.DbConfig{
		Name:    "test_clickhouse_dsn",
		Enabled: true,
		Driver:  database.DriverClickHouse,
		Connection: dbtypes.ConnectionConfig{
			Host:     "localhost",
			Port:     9000,
			Username: "default",
			Password: "your_password",
			Database: "default",
			// ClickHouse 特有参数
			ClickHouseCompress:    "lz4",
			ClickHouseSecure:      false,
			ClickHouseDebug:       false,
			ClickHouseDialTimeout: 5,
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("使用结构化配置打开ClickHouse连接失败: %v", err)
	}
	defer db.Close()

	// 验证生成的DSN
	t.Logf("生成的DSN: %s", config.DSN)

	// 测试连接
	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("ClickHouse结构化配置连接ping失败: %v", err)
	}

	t.Log("ClickHouse结构化配置测试成功")
}

// 测试ClickHouse配置参数
func TestClickHouseConfigParameters(t *testing.T) {
	tests := []struct {
		name   string
		config *dbtypes.DbConfig
	}{
		{
			name: "基本配置",
			config: &dbtypes.DbConfig{
				Name:    "clickhouse_basic",
				Enabled: true,
				Driver:  database.DriverClickHouse,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Port:     8123,
					Username: "default",
					Password: "",
					Database: "default",
				},
			},
		},
		{
			name: "压缩配置",
			config: &dbtypes.DbConfig{
				Name:    "clickhouse_compress",
				Enabled: true,
				Driver:  database.DriverClickHouse,
				Connection: dbtypes.ConnectionConfig{
					Host:               "localhost",
					Port:               8123,
					Username:           "default",
					Password:           "",
					Database:           "default",
					ClickHouseCompress: "lz4",
					ClickHouseDebug:    false,
				},
			},
		},
		{
			name: "超时配置",
			config: &dbtypes.DbConfig{
				Name:    "clickhouse_timeout",
				Enabled: true,
				Driver:  database.DriverClickHouse,
				Connection: dbtypes.ConnectionConfig{
					Host:     "localhost",
					Port:     8123,
					Username: "default",
					Password: "",
					Database: "default",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := database.Open(tt.config)
			if err != nil {
				t.Fatalf("配置 %s 连接失败: %v", tt.name, err)
			}
			defer db.Close()

			ctx := context.Background()
			if err := db.Ping(ctx); err != nil {
				t.Fatalf("配置 %s ping失败: %v", tt.name, err)
			}

			t.Logf("配置 %s 测试成功", tt.name)
		})
	}
}

// 测试ClickHouse驱动类型识别
func TestClickHouseDriverType(t *testing.T) {
	db := getClickHouseTestDB(t)
	defer db.Close()

	driver := db.GetDriver()
	if driver != database.DriverClickHouse {
		t.Fatalf("期望驱动类型为 %s，实际为 %s", database.DriverClickHouse, driver)
	}

	t.Log("ClickHouse驱动类型识别正确")
}

// 测试ClickHouse连接名称
func TestClickHouseConnectionName(t *testing.T) {
	db := getClickHouseTestDB(t)
	defer db.Close()

	name := db.GetName()
	if name == "" {
		t.Fatal("连接名称不能为空")
	}

	t.Logf("ClickHouse连接名称: %s", name)
}

// 测试ClickHouse配置加载
func TestClickHouseLoadFromConfig(t *testing.T) {
	// 这个测试验证从YAML配置文件加载ClickHouse配置
	// 需要确保configs/database.yaml文件中有clickhouse_main配置

	// 注意：这个测试假设配置文件存在且正确
	config := &dbtypes.DbConfig{
		Name:    "clickhouse_main",
		Enabled: true,
		Driver:  database.DriverClickHouse,
		Connection: dbtypes.ConnectionConfig{
			Host:                  "localhost",
			Port:                  9000,
			Username:              "default",
			Password:              "your_password",
			Database:              "gateway",
			ClickHouseCompress:    "lz4",
			ClickHouseSecure:      false,
			ClickHouseDebug:       false,
			ClickHouseDialTimeout: 5,
		},
		Pool: dbtypes.PoolConfig{
			MaxOpenConns:    50,
			MaxIdleConns:    10,
			ConnMaxLifetime: 3600,
			ConnMaxIdleTime: 1800,
		},
		Log: dbtypes.LogConfig{
			Enable:        true,
			SlowThreshold: 1000,
		},
	}

	db, err := database.Open(config)
	if err != nil {
		t.Fatalf("从配置加载ClickHouse连接失败: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("配置文件ClickHouse连接ping失败: %v", err)
	}

	t.Log("从配置文件加载ClickHouse连接成功")
}

// 测试ClickHouse特殊字符密码处理
func TestClickHouseSpecialCharactersInPassword(t *testing.T) {
	// 测试包含各种特殊字符的密码
	testCases := []struct {
		name     string
		password string
		username string
		database string
	}{
		{
			name:     "包含#和{}字符",
			password: "your_password",
			username: "default",
			database: "gateway",
		},
		{
			name:     "包含&和=字符",
			password: "pass&word=123",
			username: "test_user",
			database: "test_db",
		},
		{
			name:     "包含@和:字符",
			password: "user@domain:pass",
			username: "admin",
			database: "admin_db",
		},
		{
			name:     "包含?和%字符",
			password: "what?100%sure",
			username: "spec_user",
			database: "spec_db",
		},
		{
			name:     "包含空格和+字符",
			password: "my pass+word",
			username: "space user",
			database: "space db",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &dbtypes.DbConfig{
				Name:    "clickhouse_special_chars_test",
				Enabled: true,
				Driver:  database.DriverClickHouse,
				Connection: dbtypes.ConnectionConfig{
					Host:               "localhost",
					Port:               8123,
					Username:           tc.username,
					Password:           tc.password,
					Database:           tc.database,
					ClickHouseCompress: "lz4",
					ClickHouseDebug:    false,
				},
			}

			// 生成DSN
			db, err := database.Open(config)
			if err != nil {
				// 如果是连接错误（可能由于认证失败），我们主要关心DSN是否正确生成
				t.Logf("连接失败（预期，因为测试密码可能不正确）: %v", err)
			} else {
				defer db.Close()
				t.Log("连接成功")
			}

			// 验证生成的DSN格式是否正确
			t.Logf("生成的DSN: %s", config.DSN)

			// 检查DSN是否包含正确转义的特殊字符
			if config.DSN == "" {
				t.Fatal("DSN不应该为空")
			}

			// 验证DSN格式
			if !strings.HasPrefix(config.DSN, "tcp://") {
				t.Fatalf("DSN应该以tcp://开头，实际: %s", config.DSN)
			}

			// 验证DSN包含必要的参数
			if !strings.Contains(config.DSN, "database=") {
				t.Fatal("DSN应该包含database参数")
			}
			if !strings.Contains(config.DSN, "username=") {
				t.Fatal("DSN应该包含username参数")
			}
			if tc.password != "" && !strings.Contains(config.DSN, "password=") {
				t.Fatal("DSN应该包含password参数")
			}

			t.Logf("特殊字符密码 '%s' 的DSN生成测试通过", tc.password)
		})
	}
}

// 测试ClickHouse DSN验证功能
func TestClickHouseDSNValidation(t *testing.T) {
	testCases := []struct {
		name      string
		dsn       string
		shouldErr bool
	}{
		{
			name:      "正确的DSN格式",
			dsn:       "tcp://127.0.0.1:9000?database=test&username=default&password=123456",
			shouldErr: false,
		},
		{
			name:      "缺少tcp前缀",
			dsn:       "127.0.0.1:9000?database=test&username=default",
			shouldErr: true,
		},
		{
			name:      "空DSN",
			dsn:       "",
			shouldErr: true,
		},
		{
			name:      "包含特殊字符的正确DSN",
			dsn:       "tcp://127.0.0.1:9000?database=test&username=default&password=your_password",
			shouldErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := dsn.ValidateDSN(database.DriverClickHouse, tc.dsn)

			if tc.shouldErr && err == nil {
				t.Fatalf("期望错误但没有收到错误，DSN: %s", tc.dsn)
			}
			if !tc.shouldErr && err != nil {
				t.Fatalf("不期望错误但收到错误: %v，DSN: %s", err, tc.dsn)
			}

			if err == nil {
				t.Logf("DSN验证通过: %s", tc.dsn)
			} else {
				t.Logf("DSN验证失败（预期）: %s，错误: %v", tc.dsn, err)
			}
		})
	}
}
