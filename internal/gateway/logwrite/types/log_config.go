package types

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// LogLevel 日志级别
type LogLevel string

// 日志级别常量
const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
)

// LogFormat 日志格式
type LogFormat string

// 日志格式常量
const (
	LogFormatJSON LogFormat = "JSON"
	LogFormatText LogFormat = "TEXT"
	LogFormatCSV  LogFormat = "CSV"
)

// LogOutputTarget 日志输出目标
type LogOutputTarget string

// 日志输出目标常量
const (
	LogOutputConsole       LogOutputTarget = "CONSOLE"
	LogOutputFile          LogOutputTarget = "FILE"
	LogOutputDatabase      LogOutputTarget = "DATABASE"
	LogOutputMongoDB       LogOutputTarget = "MONGODB"
	LogOutputElasticsearch LogOutputTarget = "ELASTICSEARCH"
	LogOutputClickHouse    LogOutputTarget = "CLICKHOUSE" // 新增ClickHouse支持
)

// RotationPattern 日志轮转模式
type RotationPattern string

// 日志轮转模式常量
const (
	RotationHourly    RotationPattern = "HOURLY"
	RotationDaily     RotationPattern = "DAILY"
	RotationWeekly    RotationPattern = "WEEKLY"
	RotationSizeBased RotationPattern = "SIZE_BASED"
)

// 默认配置常量
const (
	DefaultAsyncQueueSize   = 10000
	DefaultBatchSize        = 100
	DefaultBatchTimeoutMs   = 5000
	DefaultMaxBodySizeBytes = 4096
	DefaultLogRetentionDays = 30
	DefaultMaxFileSizeMB    = 100
	DefaultMaxFileCount     = 10
	DefaultBufferSize       = 8192
	DefaultFlushThreshold   = 100
)

// LogConfig 日志配置结构体，对应数据库表 HUB_GW_LOG_CONFIG
//
// 设计说明：
// 1. 支持多种日志输出目标：控制台、文件、数据库、MongoDB、Elasticsearch、ClickHouse
// 2. 提供异步和批量处理能力，优化性能
// 3. 支持日志轮转和保留策略
// 4. 内置敏感数据脱敏功能
// 5. 可配置的日志内容控制
// 6. 支持实例级和路由级配置
type LogConfig struct {
	// 基础标识信息
	TenantID    string `json:"tenantId" db:"tenantId"`       // 租户ID，多租户环境标识
	LogConfigID string `json:"logConfigId" db:"logConfigId"` // 日志配置ID，全局唯一标识
	ConfigName  string `json:"configName" db:"configName"`   // 配置名称，便于管理和识别
	ConfigDesc  string `json:"configDesc" db:"configDesc"`   // 配置描述，详细说明配置用途

	// 日志内容控制 - 控制日志记录的详细程度和格式
	LogFormat          string `json:"logFormat" db:"logFormat"`                   // 日志格式(JSON,TEXT,CSV)
	RecordRequestBody  string `json:"recordRequestBody" db:"recordRequestBody"`   // 是否记录请求体(N否,Y是)
	RecordResponseBody string `json:"recordResponseBody" db:"recordResponseBody"` // 是否记录响应体(N否,Y是)
	RecordHeaders      string `json:"recordHeaders" db:"recordHeaders"`           // 是否记录请求/响应头(N否,Y是)
	MaxBodySizeBytes   int    `json:"maxBodySizeBytes" db:"maxBodySizeBytes"`     // 最大记录的报文大小，超过则截断

	// 日志输出目标配置 - 支持多种存储后端
	OutputTargets       string `json:"outputTargets" db:"outputTargets"`             // 输出目标列表，逗号分隔(CONSOLE,FILE,DATABASE,MONGODB,ELASTICSEARCH,CLICKHOUSE)
	FileConfig          string `json:"fileConfig" db:"fileConfig"`                   // 文件输出配置，JSON格式
	DatabaseConfig      string `json:"databaseConfig" db:"databaseConfig"`           // 数据库输出配置，JSON格式
	MongoConfig         string `json:"mongoConfig" db:"mongoConfig"`                 // MongoDB输出配置，JSON格式
	ElasticsearchConfig string `json:"elasticsearchConfig" db:"elasticsearchConfig"` // Elasticsearch输出配置，JSON格式
	ClickHouseConfig    string `json:"clickhouseConfig" db:"clickhouseConfig"`       // ClickHouse输出配置，JSON格式，支持高性能列式存储

	// 异步和批量处理配置 - 性能优化相关
	EnableAsyncLogging    string `json:"enableAsyncLogging" db:"enableAsyncLogging"`       // 是否启用异步日志(N否,Y是)
	AsyncQueueSize        int    `json:"asyncQueueSize" db:"asyncQueueSize"`               // 异步队列大小(100-1000000)
	AsyncFlushIntervalMs  int    `json:"asyncFlushIntervalMs" db:"asyncFlushIntervalMs"`   // 异步刷新间隔(毫秒)
	EnableBatchProcessing string `json:"enableBatchProcessing" db:"enableBatchProcessing"` // 是否启用批量处理(N否,Y是)
	BatchSize             int    `json:"batchSize" db:"batchSize"`                         // 批量处理大小(1-10000)
	BatchTimeoutMs        int    `json:"batchTimeoutMs" db:"batchTimeoutMs"`               // 批量处理超时时间(毫秒)

	// 日志保留和轮转配置 - 磁盘空间管理
	LogRetentionDays   int    `json:"logRetentionDays" db:"logRetentionDays"`     // 日志保留天数(1-3650)
	EnableFileRotation string `json:"enableFileRotation" db:"enableFileRotation"` // 是否启用文件轮转(N否,Y是)
	MaxFileSizeMB      int    `json:"maxFileSizeMB" db:"maxFileSizeMB"`           // 单个文件最大大小(MB)
	MaxFileCount       int    `json:"maxFileCount" db:"maxFileCount"`             // 最大文件数量
	RotationPattern    string `json:"rotationPattern" db:"rotationPattern"`       // 轮转模式(HOURLY,DAILY,WEEKLY,SIZE_BASED)

	// 敏感数据处理 - 数据安全保护
	EnableSensitiveDataMasking string `json:"enableSensitiveDataMasking" db:"enableSensitiveDataMasking"` // 是否启用敏感数据脱敏(N否,Y是)
	SensitiveFields            string `json:"sensitiveFields" db:"sensitiveFields"`                       // 敏感字段列表，JSON数组
	MaskingPattern             string `json:"maskingPattern" db:"maskingPattern"`                         // 脱敏替换模式(如***)

	// 性能优化配置
	BufferSize     int `json:"bufferSize" db:"bufferSize"`         // 缓冲区大小(字节)
	FlushThreshold int `json:"flushThreshold" db:"flushThreshold"` // 刷新阈值(条数)

	// 配置管理
	ConfigPriority int    `json:"configPriority" db:"configPriority"` // 配置优先级，数值越小优先级越高
	ActiveFlag     string `json:"activeFlag" db:"activeFlag"`         // 活动状态标记(N非活动,Y活动)

	// 扩展属性（JSON字符串）
	ExtProperty string `json:"extProperty" db:"extProperty"` // 扩展属性,JSON格式（用于告警等扩展配置）

	// 解析后的告警配置（构建时预解析，避免重复解析JSON）
	alertConfig *AlertConfig // 私有字段，通过 GetAlertConfig() 访问
}

// SetAlertConfig 设置告警配置（供构建时使用）
func (c *LogConfig) SetAlertConfig(cfg *AlertConfig) {
	c.alertConfig = cfg
}

// AlertConfig 告警配置（从 extProperty 解析）
type AlertConfig struct {
	AlertEnabled       bool
	ChannelName        string
	AlertStatusCodes   []int
	AlertOnTimeout     bool
	TimeoutThresholdMs int
}

// GetAlertConfig 获取告警配置（如果未解析则解析，已解析则直接返回）
func (c *LogConfig) GetAlertConfig() *AlertConfig {
	if c.alertConfig != nil {
		return c.alertConfig
	}
	// 如果未解析，则解析一次（延迟解析，兼容旧代码）
	c.alertConfig = ParseAlertConfigFromExtProperty(c.ExtProperty)
	return c.alertConfig
}

// ParseAlertConfigFromExtProperty 从 extProperty JSON 字符串解析告警配置（导出函数，供其他包使用）
// 按照前端实际保存的格式解析：
// - alertEnabled: 'Y'/'N' 字符串
// - channelName: string
// - alertStatusCodes: string[] (字符串数组，需要转换为 int[])
// - alertOnTimeout: 'Y'/'N' 字符串
// - timeoutThresholdMs: number
func ParseAlertConfigFromExtProperty(extProperty string) *AlertConfig {
	cfg := &AlertConfig{
		AlertEnabled:       false,
		ChannelName:        "",
		AlertStatusCodes:   []int{502}, // 默认502
		AlertOnTimeout:     true,       // 默认开启超时告警
		TimeoutThresholdMs: 120000,     // 默认120秒
	}

	if strings.TrimSpace(extProperty) == "" {
		return cfg
	}

	var m map[string]interface{}
	if err := json.Unmarshal([]byte(extProperty), &m); err != nil {
		return cfg
	}

	// alertEnabled: 'Y'/'N' 字符串
	if v, ok := m["alertEnabled"].(string); ok {
		cfg.AlertEnabled = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// channelName: string
	if v, ok := m["channelName"].(string); ok {
		cfg.ChannelName = v
	}

	// alertStatusCodes: string[] (前端保存的是字符串数组)
	if v, ok := m["alertStatusCodes"].([]interface{}); ok {
		codes := make([]int, 0, len(v))
		for _, item := range v {
			var code int
			switch t := item.(type) {
			case string:
				if c, err := strconv.Atoi(strings.TrimSpace(t)); err == nil && c > 0 {
					code = c
				}
			case float64:
				code = int(t)
			case int:
				code = t
			}
			if code > 0 {
				codes = append(codes, code)
			}
		}
		if len(codes) > 0 {
			cfg.AlertStatusCodes = codes
		}
	}

	// alertOnTimeout: 'Y'/'N' 字符串（默认开启）
	if v, ok := m["alertOnTimeout"].(string); ok {
		cfg.AlertOnTimeout = strings.TrimSpace(strings.ToUpper(v)) == "Y"
	}

	// timeoutThresholdMs: number
	if v, ok := m["timeoutThresholdMs"]; ok {
		switch t := v.(type) {
		case float64:
			cfg.TimeoutThresholdMs = int(t)
		case int:
			cfg.TimeoutThresholdMs = t
		}
	}

	return cfg
}

// FileOutputConfig 文件输出配置
type FileOutputConfig struct {
	Path         string `json:"path"`          // 日志文件路径
	Prefix       string `json:"prefix"`        // 文件名前缀
	Extension    string `json:"extension"`     // 文件扩展名
	Compress     bool   `json:"compress"`      // 是否压缩
	MaxSize      int    `json:"max_size"`      // 单个文件最大大小(MB)
	MaxAge       int    `json:"max_age"`       // 文件保留最大天数
	MaxBackups   int    `json:"max_backups"`   // 最大备份文件数
	LocalTime    bool   `json:"local_time"`    // 是否使用本地时间
	RotationTime string `json:"rotation_time"` // 轮转时间(daily, hourly)
}

// DatabaseOutputConfig 数据库输出配置
type DatabaseOutputConfig struct {
	ConnectionName string `json:"connection_name"` // 数据库连接名
	TableName      string `json:"table_name"`      // 表名
	BatchSize      int    `json:"batch_size"`      // 批量插入大小
	AsyncInsert    bool   `json:"async_insert"`    // 是否异步插入
}

// MongoDBOutputConfig MongoDB输出配置
type MongoDBOutputConfig struct {
	URI            string `json:"uri"`             // MongoDB连接URI
	Database       string `json:"database"`        // 数据库名
	Collection     string `json:"collection"`      // 集合名
	ConnectTimeout int    `json:"connect_timeout"` // 连接超时(毫秒)
	BatchSize      int    `json:"batch_size"`      // 批量插入大小
	AsyncInsert    bool   `json:"async_insert"`    // 是否异步插入
}

// ElasticsearchOutputConfig Elasticsearch输出配置
type ElasticsearchOutputConfig struct {
	Addresses    []string `json:"addresses"`     // ES地址列表
	IndexName    string   `json:"index_name"`    // 索引名称
	IndexPattern string   `json:"index_pattern"` // 索引模式(如daily)
	Username     string   `json:"username"`      // 用户名
	Password     string   `json:"password"`      // 密码
	BatchSize    int      `json:"batch_size"`    // 批量插入大小
	AsyncInsert  bool     `json:"async_insert"`  // 是否异步插入
}

// ClickHouseOutputConfig ClickHouse输出配置 - 新增高性能列式存储支持
type ClickHouseOutputConfig struct {
	DSN            string `json:"dsn"`             // ClickHouse连接DSN
	Database       string `json:"database"`        // 数据库名
	Table          string `json:"table"`           // 表名
	BatchSize      int    `json:"batch_size"`      // 批量插入大小
	FlushInterval  int    `json:"flush_interval"`  // 刷新间隔(秒)
	AsyncInsert    bool   `json:"async_insert"`    // 是否异步插入
	ConnectTimeout int    `json:"connect_timeout"` // 连接超时(毫秒)
	MaxOpenConns   int    `json:"max_open_conns"`  // 最大连接数
	Compress       bool   `json:"compress"`        // 是否启用压缩
}

// GetOutputTargets 获取输出目标列表
func (c *LogConfig) GetOutputTargets() []LogOutputTarget {
	if c.OutputTargets == "" {
		return []LogOutputTarget{LogOutputConsole}
	}

	targets := strings.Split(c.OutputTargets, ",")
	result := make([]LogOutputTarget, 0, len(targets))

	for _, target := range targets {
		target = strings.TrimSpace(target)
		if target != "" {
			result = append(result, LogOutputTarget(target))
		}
	}

	return result
}

// GetFileConfig 解析文件输出配置
func (c *LogConfig) GetFileConfig() (*FileOutputConfig, error) {
	if c.FileConfig == "" {
		return &FileOutputConfig{
			Path:         "./logs",
			Prefix:       "gateway-access",
			Extension:    ".log",
			Compress:     true,
			MaxSize:      100,
			MaxAge:       7,
			MaxBackups:   10,
			LocalTime:    true,
			RotationTime: "daily",
		}, nil
	}

	var config FileOutputConfig
	err := json.Unmarshal([]byte(c.FileConfig), &config)
	return &config, err
}

// GetDatabaseConfig 解析数据库输出配置
func (c *LogConfig) GetDatabaseConfig() (*DatabaseOutputConfig, error) {
	if c.DatabaseConfig == "" {
		return &DatabaseOutputConfig{
			ConnectionName: "default",
			TableName:      "HUB_GW_ACCESS_LOG",
			BatchSize:      100,
			AsyncInsert:    true,
		}, nil
	}

	var config DatabaseOutputConfig
	err := json.Unmarshal([]byte(c.DatabaseConfig), &config)
	return &config, err
}

// GetMongoConfig 解析MongoDB输出配置
func (c *LogConfig) GetMongoConfig() (*MongoDBOutputConfig, error) {
	if c.MongoConfig == "" {
		return &MongoDBOutputConfig{
			URI:            "mongodb://localhost:27017",
			Database:       "gateway_logs",
			Collection:     "access_logs",
			ConnectTimeout: 5000,
			BatchSize:      100,
			AsyncInsert:    true,
		}, nil
	}

	var config MongoDBOutputConfig
	err := json.Unmarshal([]byte(c.MongoConfig), &config)
	return &config, err
}

// GetElasticsearchConfig 解析Elasticsearch输出配置
func (c *LogConfig) GetElasticsearchConfig() (*ElasticsearchOutputConfig, error) {
	if c.ElasticsearchConfig == "" {
		return &ElasticsearchOutputConfig{
			Addresses:    []string{"http://localhost:9200"},
			IndexName:    "gateway-logs",
			IndexPattern: "daily",
			BatchSize:    100,
			AsyncInsert:  true,
		}, nil
	}

	var config ElasticsearchOutputConfig
	err := json.Unmarshal([]byte(c.ElasticsearchConfig), &config)
	return &config, err
}

// GetClickHouseConfig 解析ClickHouse输出配置 - 新增方法
func (c *LogConfig) GetClickHouseConfig() (*ClickHouseOutputConfig, error) {
	if c.ClickHouseConfig == "" {
		return &ClickHouseOutputConfig{
			DSN:            "tcp://localhost:9000/gateway_logs",
			Database:       "gateway_logs",
			Table:          "access_logs",
			BatchSize:      1000,
			FlushInterval:  10,
			AsyncInsert:    true,
			ConnectTimeout: 10000,
			MaxOpenConns:   10,
			Compress:       true,
		}, nil
	}

	var config ClickHouseOutputConfig
	err := json.Unmarshal([]byte(c.ClickHouseConfig), &config)
	return &config, err
}

// IsRecordRequestBody 是否记录请求体
func (c *LogConfig) IsRecordRequestBody() bool {
	return c.RecordRequestBody == "Y"
}

// IsRecordResponseBody 是否记录响应体
func (c *LogConfig) IsRecordResponseBody() bool {
	return c.RecordResponseBody == "Y"
}

// IsRecordHeaders 是否记录请求/响应头
func (c *LogConfig) IsRecordHeaders() bool {
	return c.RecordHeaders == "Y"
}

// IsAsyncLogging 是否启用异步日志
func (c *LogConfig) IsAsyncLogging() bool {
	return c.EnableAsyncLogging == "Y"
}

// IsBatchProcessing 是否启用批量处理
func (c *LogConfig) IsBatchProcessing() bool {
	return c.EnableBatchProcessing == "Y"
}

// IsSensitiveDataMasking 是否启用敏感数据脱敏
func (c *LogConfig) IsSensitiveDataMasking() bool {
	return c.EnableSensitiveDataMasking == "Y"
}

// Validate 验证日志配置的有效性
func (c *LogConfig) Validate() error {
	if c.TenantID == "" {
		return fmt.Errorf("租户ID不能为空")
	}

	if c.LogConfigID == "" {
		return fmt.Errorf("日志配置ID不能为空")
	}

	if c.ConfigName == "" {
		return fmt.Errorf("配置名称不能为空")
	}

	// 验证日志格式
	validFormats := []string{string(LogFormatJSON), string(LogFormatText), string(LogFormatCSV)}
	if !contains(validFormats, c.LogFormat) {
		return fmt.Errorf("无效的日志格式: %s", c.LogFormat)
	}

	// 验证输出目标
	targets := c.GetOutputTargets()
	if len(targets) == 0 {
		return fmt.Errorf("至少需要配置一个输出目标")
	}

	validTargets := []string{
		string(LogOutputConsole), string(LogOutputFile), string(LogOutputDatabase),
		string(LogOutputMongoDB), string(LogOutputElasticsearch), string(LogOutputClickHouse),
	}
	for _, target := range targets {
		if !contains(validTargets, string(target)) {
			return fmt.Errorf("无效的输出目标: %s", target)
		}
	}

	// 验证数值范围
	if c.AsyncQueueSize < 100 || c.AsyncQueueSize > 1000000 {
		return fmt.Errorf("异步队列大小必须在100-1000000之间")
	}

	if c.BatchSize < 1 || c.BatchSize > 10000 {
		return fmt.Errorf("批处理大小必须在1-10000之间")
	}

	if c.MaxBodySizeBytes < 0 || c.MaxBodySizeBytes > 10485760 { // 10MB
		return fmt.Errorf("最大报文大小必须在0-10MB之间")
	}

	if c.LogRetentionDays < 1 || c.LogRetentionDays > 3650 { // 10年
		return fmt.Errorf("日志保留天数必须在1-3650之间")
	}

	// 验证轮转模式
	if c.EnableFileRotation == "Y" {
		validPatterns := []string{
			string(RotationHourly), string(RotationDaily),
			string(RotationWeekly), string(RotationSizeBased),
		}
		if !contains(validPatterns, c.RotationPattern) {
			return fmt.Errorf("无效的轮转模式: %s", c.RotationPattern)
		}
	}

	return nil
}

// SetDefaults 设置默认值
func (c *LogConfig) SetDefaults() {
	if c.LogFormat == "" {
		c.LogFormat = string(LogFormatJSON)
	}

	if c.RecordRequestBody == "" {
		c.RecordRequestBody = "N"
	}

	if c.RecordResponseBody == "" {
		c.RecordResponseBody = "N"
	}

	if c.RecordHeaders == "" {
		c.RecordHeaders = "Y"
	}

	if c.MaxBodySizeBytes == 0 {
		c.MaxBodySizeBytes = DefaultMaxBodySizeBytes
	}

	if c.OutputTargets == "" {
		c.OutputTargets = string(LogOutputConsole)
	}

	if c.EnableAsyncLogging == "" {
		c.EnableAsyncLogging = "Y"
	}

	if c.AsyncQueueSize == 0 {
		c.AsyncQueueSize = DefaultAsyncQueueSize
	}

	if c.AsyncFlushIntervalMs == 0 {
		c.AsyncFlushIntervalMs = 10000
	}

	if c.EnableBatchProcessing == "" {
		c.EnableBatchProcessing = "Y"
	}

	if c.BatchSize == 0 {
		c.BatchSize = DefaultBatchSize
	}

	if c.BatchTimeoutMs == 0 {
		c.BatchTimeoutMs = DefaultBatchTimeoutMs
	}

	if c.LogRetentionDays == 0 {
		c.LogRetentionDays = DefaultLogRetentionDays
	}

	if c.EnableFileRotation == "" {
		c.EnableFileRotation = "Y"
	}

	if c.MaxFileSizeMB == 0 {
		c.MaxFileSizeMB = DefaultMaxFileSizeMB
	}

	if c.MaxFileCount == 0 {
		c.MaxFileCount = DefaultMaxFileCount
	}

	if c.RotationPattern == "" {
		c.RotationPattern = string(RotationDaily)
	}

	if c.EnableSensitiveDataMasking == "" {
		c.EnableSensitiveDataMasking = "Y"
	}

	if c.MaskingPattern == "" {
		c.MaskingPattern = "***"
	}

	if c.BufferSize == 0 {
		c.BufferSize = DefaultBufferSize
	}

	if c.FlushThreshold == 0 {
		c.FlushThreshold = DefaultFlushThreshold
	}

	if c.ActiveFlag == "" {
		c.ActiveFlag = "Y"
	}
}

// contains 辅助函数：检查字符串是否在切片中
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
