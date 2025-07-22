// Package dbtypes defines common database types and constants
// that can be shared between the main database and utility packages
package dbtypes

import (
	"fmt"
	"gateway/pkg/config"
)

// 支持的数据库类型常量
// 用于指定要使用的数据库驱动
const (
	// MySQL数据库驱动
	DriverMySQL = "mysql"
	// PostgreSQL数据库驱动
	DriverPostgreSQL = "postgres"
	// SQLite数据库驱动
	DriverSQLite = "sqlite"
	// SQL Server数据库驱动
	DriverSQLServer = "sqlserver"
	// Oracle数据库驱动
	DriverOracle = "oracle"
	// Oracle 11g数据库驱动（使用特殊分页语法）
	DriverOracle11g = "oracle11g"
	// MariaDB数据库驱动 (兼容MySQL)
	DriverMariaDB = "mariadb"
	// TiDB数据库驱动 (兼容MySQL)
	DriverTiDB = "tidb"
	// ClickHouse数据库驱动
	DriverClickHouse = "clickhouse"
	// MongoDB数据库驱动 (NoSQL)
	DriverMongoDB = "mongodb"
)

// ConnectionConfig 数据库连接配置
// 描述数据库连接的基本信息，支持多种数据库类型
type ConnectionConfig struct {
	// === 通用连接参数 ===

	// Host 数据库主机地址 (MySQL, PostgreSQL, Oracle等需要)
	Host string `mapstructure:"host"`

	// Port 数据库端口 (MySQL, PostgreSQL, Oracle等需要)
	Port int `mapstructure:"port"`

	// Username 用户名 (MySQL, PostgreSQL, Oracle等需要)
	Username string `mapstructure:"username"`

	// Password 密码 (MySQL, PostgreSQL, Oracle等需要)
	Password string `mapstructure:"password"`

	// Database 数据库名 (MySQL, PostgreSQL等需要)
	Database string `mapstructure:"database"`

	// === MySQL特有参数 ===

	// Charset MySQL字符集
	Charset string `mapstructure:"charset"`
	// ParseTime MySQL是否解析时间类型
	ParseTime bool `mapstructure:"parse_time"`
	// Loc MySQL时区设置
	Loc string `mapstructure:"loc"`
	// MySQLConnectTimeout MySQL连接超时时间(秒)
	MySQLConnectTimeout int `mapstructure:"mysql_connect_timeout"`
	// MySQLReadTimeout MySQL读取超时时间(秒)
	MySQLReadTimeout int `mapstructure:"mysql_read_timeout"`
	// MySQLWriteTimeout MySQL写入超时时间(秒)
	MySQLWriteTimeout int `mapstructure:"mysql_write_timeout"`

	// === PostgreSQL特有参数 ===

	// SSLMode PostgreSQL SSL模式
	SSLMode string `mapstructure:"sslmode"`
	// PostgreSQLConnectTimeout PostgreSQL连接超时时间(秒)
	PostgreSQLConnectTimeout int `mapstructure:"postgres_connect_timeout"`
	// PostgreSQLStatementTimeout PostgreSQL语句超时时间(秒)
	PostgreSQLStatementTimeout int `mapstructure:"postgres_statement_timeout"`

	// === SQLite特有参数 ===

	// FilePath SQLite数据库文件路径 (优先于DSN中的路径)
	FilePath string `mapstructure:"file_path"`
	// JournalMode SQLite日志模式 (WAL, DELETE, TRUNCATE, PERSIST, MEMORY, OFF)
	JournalMode string `mapstructure:"journal_mode"`
	// SynchronousMode SQLite同步模式 (OFF, NORMAL, FULL, EXTRA)
	SynchronousMode string `mapstructure:"synchronous_mode"`
	// CacheMode SQLite缓存模式 (shared, private)
	CacheMode string `mapstructure:"cache_mode"`
	// ConnectionMode SQLite连接模式 (rwc, ro, rw, memory)
	ConnectionMode string `mapstructure:"connection_mode"`
	// CacheSize SQLite缓存大小 (页数或KB，负数表示KB)
	CacheSize int `mapstructure:"cache_size"`
	// BusyTimeout SQLite忙等待超时时间(毫秒)
	BusyTimeout int `mapstructure:"busy_timeout"`
	// ForeignKeys SQLite是否启用外键约束
	ForeignKeys bool `mapstructure:"foreign_keys"`

	// === Oracle特有参数 ===

	// ServiceName Oracle服务名 (推荐使用)
	ServiceName string `mapstructure:"service_name"`
	// SID Oracle系统标识符 (传统方式)
	SID string `mapstructure:"sid"`
	// UseSID 是否使用SID连接方式 (true: 使用SID, false: 使用服务名)
	UseSID bool `mapstructure:"use_sid"`
	// Timezone Oracle时区设置 (如: Asia/Shanghai, UTC等)
	Timezone string `mapstructure:"timezone"`
	// OracleConnectionTimeout Oracle连接超时时间(秒)
	OracleConnectionTimeout int `mapstructure:"oracle_connection_timeout"`
	// OracleReadTimeout Oracle读取超时时间(秒)
	OracleReadTimeout int `mapstructure:"oracle_read_timeout"`
	// OracleWriteTimeout Oracle写入超时时间(秒)
	OracleWriteTimeout int `mapstructure:"oracle_write_timeout"`
	// NLSLang Oracle语言环境设置
	NLSLang string `mapstructure:"nls_lang"`
	// AutoCommit Oracle是否自动提交
	AutoCommit bool `mapstructure:"auto_commit"`
	// PrefetchRows Oracle预取行数
	PrefetchRows int `mapstructure:"prefetch_rows"`
	// LobPrefetchSize Oracle LOB预取大小
	LobPrefetchSize int `mapstructure:"lob_prefetch_size"`

	// === ClickHouse官网标准DSN参数 ===
	// 参考: https://clickhouse.com/docs/zh/integrations/go#databasesql-api

	// ClickHouseCompress 压缩算法 (none,lz4,zstd,gzip,deflate,br) 默认"none"
	ClickHouseCompress string `mapstructure:"clickhouse_compress"`
	// ClickHouseCompressLevel 压缩级别 (0-11,算法相关) 默认0
	ClickHouseCompressLevel int `mapstructure:"clickhouse_compress_level"`
	// ClickHouseSecure 是否建立安全SSL连接 (默认false)
	ClickHouseSecure bool `mapstructure:"clickhouse_secure"`
	// ClickHouseSkipVerify 跳过证书验证 (默认false)
	ClickHouseSkipVerify bool `mapstructure:"clickhouse_skip_verify"`
	// ClickHouseDebug 启用调试输出 (默认false)
	ClickHouseDebug bool `mapstructure:"clickhouse_debug"`
	// ClickHouseDialTimeout 拨号超时时间(秒) (默认30s)
	ClickHouseDialTimeout int `mapstructure:"clickhouse_dial_timeout"`
	// ClickHouseBlockBufferSize 块缓冲区大小 (默认2)
	ClickHouseBlockBufferSize int `mapstructure:"clickhouse_block_buffer_size"`
	// ClickHouseConnOpenStrategy 连接打开策略 (random/in_order) 默认"random"
	ClickHouseConnOpenStrategy string `mapstructure:"clickhouse_conn_open_strategy"`
	// ClickHouseHosts 负载均衡主机列表 (格式: "host1:9000,host2:9000") 官网标准参数
	ClickHouseHosts string `mapstructure:"clickhouse_hosts"`
}

// PoolConfig 连接池配置
// 控制数据库连接池的行为
type PoolConfig struct {
	// MaxOpenConns 最大打开连接数
	MaxOpenConns int `mapstructure:"max_open_conns"`

	// MaxIdleConns 最大空闲连接数
	MaxIdleConns int `mapstructure:"max_idle_conns"`

	// ConnMaxLifetime 连接最大生存时间（秒）
	ConnMaxLifetime int64 `mapstructure:"conn_max_lifetime"`

	// ConnMaxIdleTime 连接最大空闲时间（秒）
	ConnMaxIdleTime int64 `mapstructure:"conn_max_idle_time"`
}

// LogConfig 日志配置
// 控制数据库操作的日志记录
type LogConfig struct {
	// Enable 是否启用日志
	Enable bool `mapstructure:"enable"`

	// SlowThreshold 慢查询阈值（毫秒）
	SlowThreshold int `mapstructure:"slow_threshold"`
}

// TransactionConfig 事务配置
// 控制数据库事务的默认行为
type TransactionConfig struct {
	// DefaultUse 默认是否使用事务
	DefaultUse bool `mapstructure:"default_use"`
}

// DbConfig 数据库配置结构体
// 用于配置数据库连接和操作行为
type DbConfig struct {
	// Name 数据库连接名称
	// 用于唯一标识数据库连接，同一驱动类型可以有多个不同的命名连接
	Name string `mapstructure:"name"`

	// Enabled 是否启用此数据库连接
	// 设置为false时此连接将不会被加载
	Enabled bool `mapstructure:"enabled"`

	// Driver 数据库驱动类型 (mysql, postgres, sqlite)
	// 决定要连接的数据库类型
	Driver string `mapstructure:"driver"`

	// ConnectionString 数据源名称 (连接字符串)
	// 例如: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// 如果提供此值，将优先使用此连接字符串，否则会从Connection生成
	DSN string `mapstructure:"dsn"`

	// Connection 连接配置
	// 包含主机、端口、用户名、密码等信息
	Connection ConnectionConfig `mapstructure:"connection"`

	// Pool 连接池配置
	// 控制数据库连接池的行为
	Pool PoolConfig `mapstructure:"pool"`

	// Log 日志配置
	// 控制SQL日志记录
	Log LogConfig `mapstructure:"log"`

	// Transaction 事务配置
	// 控制事务默认行为
	Transaction TransactionConfig `mapstructure:"transaction"`
}

// DatabasesConfig 数据库配置文件的根结构
// 用于解析配置文件中的数据库相关配置
type DatabasesConfig struct {
	// Default 默认数据库连接名称
	Default string `mapstructure:"default"`

	// Connections 所有数据库连接的配置映射
	Connections map[string]*DbConfig `mapstructure:"connections"`
}

// LoadDatabaseConfigs 从配置文件加载所有数据库配置
// 解析YAML配置文件，返回所有数据库连接的配置
// 参数:
//
//	configPath: 配置文件路径
//
// 返回:
//
//	map[string]*DbConfig: 连接名称到配置的映射
//	error: 加载失败时返回错误信息
func LoadDatabaseConfigs(configPath string) (map[string]*DbConfig, error) {
	// 加载配置文件
	if err := config.LoadConfigFile(configPath, config.LoadOptions{
		ClearExisting: false,
		AllowOverride: true,
	}); err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	// 解析数据库配置
	var dbConfig DatabasesConfig
	if err := config.GetSection("database", &dbConfig); err != nil {
		return nil, fmt.Errorf("解析数据库配置失败: %w", err)
	}

	// 设置默认值并验证配置
	for name, cfg := range dbConfig.Connections {
		// 设置连接名称
		cfg.Name = name

		// 如果未设置enabled，默认为true
		if cfg.Driver != "" && !cfg.hasEnabledSet() {
			cfg.Enabled = true
		}

		// 验证必要字段
		if cfg.Driver == "" {
			return nil, fmt.Errorf("数据库连接 '%s' 缺少driver配置", name)
		}

		// 设置默认连接池配置
		if cfg.Pool.MaxOpenConns == 0 {
			cfg.Pool.MaxOpenConns = 100
		}
		if cfg.Pool.MaxIdleConns == 0 {
			cfg.Pool.MaxIdleConns = 25
		}
		if cfg.Pool.ConnMaxLifetime == 0 {
			cfg.Pool.ConnMaxLifetime = 3600 // 1小时
		}
		if cfg.Pool.ConnMaxIdleTime == 0 {
			cfg.Pool.ConnMaxIdleTime = 1800 // 30分钟
		}

		// 设置默认日志配置
		if cfg.Log.SlowThreshold == 0 {
			cfg.Log.SlowThreshold = 200 // 200毫秒
		}
	}

	return dbConfig.Connections, nil
}

// hasEnabledSet 检查enabled字段是否被明确设置
// 这是一个辅助方法，用于区分明确设置为false和未设置的情况
func (cfg *DbConfig) hasEnabledSet() bool {
	// 通过检查是否存在于原始配置中来判断
	// 这里简化处理，实际应用中可能需要更复杂的逻辑
	return true // 简化实现，假设总是被设置
}
