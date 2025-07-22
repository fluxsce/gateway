package models

import (
	"time"
)

// LogConfig 日志配置模型，对应数据库表 HUB_GW_LOG_CONFIG
type LogConfig struct {
	// 基础标识信息
	TenantId        string `json:"tenantId" form:"tenantId" query:"tenantId" db:"tenantId"`                         // 租户ID，联合主键
	LogConfigId     string `json:"logConfigId" form:"logConfigId" query:"logConfigId" db:"logConfigId"`             // 日志配置ID，联合主键
	ConfigName      string `json:"configName" form:"configName" query:"configName" db:"configName"`                 // 配置名称
	ConfigDesc      string `json:"configDesc" form:"configDesc" query:"configDesc" db:"configDesc"`                 // 配置描述
	
	// 日志内容控制
	LogFormat         string `json:"logFormat" form:"logFormat" query:"logFormat" db:"logFormat"`                             // 日志格式(JSON,TEXT,CSV)
	RecordRequestBody string `json:"recordRequestBody" form:"recordRequestBody" query:"recordRequestBody" db:"recordRequestBody"` // 是否记录请求体(N否,Y是)
	RecordResponseBody string `json:"recordResponseBody" form:"recordResponseBody" query:"recordResponseBody" db:"recordResponseBody"` // 是否记录响应体(N否,Y是)
	RecordHeaders     string `json:"recordHeaders" form:"recordHeaders" query:"recordHeaders" db:"recordHeaders"`             // 是否记录请求/响应头(N否,Y是)
	MaxBodySizeBytes  int    `json:"maxBodySizeBytes" form:"maxBodySizeBytes" query:"maxBodySizeBytes" db:"maxBodySizeBytes"` // 最大记录报文大小(字节)
	
	// 日志输出目标配置
	OutputTargets       string `json:"outputTargets" form:"outputTargets" query:"outputTargets" db:"outputTargets"`                   // 输出目标,逗号分隔
	FileConfig          string `json:"fileConfig" form:"fileConfig" query:"fileConfig" db:"fileConfig"`                               // 文件输出配置,JSON格式
	DatabaseConfig      string `json:"databaseConfig" form:"databaseConfig" query:"databaseConfig" db:"databaseConfig"`               // 数据库输出配置,JSON格式
	MongoConfig         string `json:"mongoConfig" form:"mongoConfig" query:"mongoConfig" db:"mongoConfig"`                           // MongoDB输出配置,JSON格式
	ElasticsearchConfig string `json:"elasticsearchConfig" form:"elasticsearchConfig" query:"elasticsearchConfig" db:"elasticsearchConfig"` // Elasticsearch输出配置,JSON格式
	ClickhouseConfig    string `json:"clickhouseConfig" form:"clickhouseConfig" query:"clickhouseConfig" db:"clickhouseConfig"`       // ClickHouse输出配置,JSON格式
	
	// 异步和批量处理配置
	EnableAsyncLogging    string `json:"enableAsyncLogging" form:"enableAsyncLogging" query:"enableAsyncLogging" db:"enableAsyncLogging"`       // 是否启用异步日志(N否,Y是)
	AsyncQueueSize        int    `json:"asyncQueueSize" form:"asyncQueueSize" query:"asyncQueueSize" db:"asyncQueueSize"`                       // 异步队列大小
	AsyncFlushIntervalMs  int    `json:"asyncFlushIntervalMs" form:"asyncFlushIntervalMs" query:"asyncFlushIntervalMs" db:"asyncFlushIntervalMs"` // 异步刷新间隔(毫秒)
	EnableBatchProcessing string `json:"enableBatchProcessing" form:"enableBatchProcessing" query:"enableBatchProcessing" db:"enableBatchProcessing"` // 是否启用批量处理(N否,Y是)
	BatchSize             int    `json:"batchSize" form:"batchSize" query:"batchSize" db:"batchSize"`                                             // 批处理大小
	BatchTimeoutMs        int    `json:"batchTimeoutMs" form:"batchTimeoutMs" query:"batchTimeoutMs" db:"batchTimeoutMs"`                       // 批处理超时时间(毫秒)
	
	// 日志保留和轮转配置
	LogRetentionDays   int    `json:"logRetentionDays" form:"logRetentionDays" query:"logRetentionDays" db:"logRetentionDays"`           // 日志保留天数
	EnableFileRotation string `json:"enableFileRotation" form:"enableFileRotation" query:"enableFileRotation" db:"enableFileRotation"` // 是否启用文件轮转(N否,Y是)
	MaxFileSizeMB      *int   `json:"maxFileSizeMB" form:"maxFileSizeMB" query:"maxFileSizeMB" db:"maxFileSizeMB"`                       // 最大文件大小(MB)
	MaxFileCount       *int   `json:"maxFileCount" form:"maxFileCount" query:"maxFileCount" db:"maxFileCount"`                           // 最大文件数量
	RotationPattern    string `json:"rotationPattern" form:"rotationPattern" query:"rotationPattern" db:"rotationPattern"`               // 轮转模式(HOURLY,DAILY,WEEKLY,SIZE_BASED)
	
	// 敏感数据处理
	EnableSensitiveDataMasking string `json:"enableSensitiveDataMasking" form:"enableSensitiveDataMasking" query:"enableSensitiveDataMasking" db:"enableSensitiveDataMasking"` // 是否启用敏感数据脱敏(N否,Y是)
	SensitiveFields            string `json:"sensitiveFields" form:"sensitiveFields" query:"sensitiveFields" db:"sensitiveFields"`                                                 // 敏感字段列表,JSON数组格式
	MaskingPattern             string `json:"maskingPattern" form:"maskingPattern" query:"maskingPattern" db:"maskingPattern"`                                                   // 脱敏替换模式
	
	// 性能优化配置
	BufferSize     int `json:"bufferSize" form:"bufferSize" query:"bufferSize" db:"bufferSize"`             // 缓冲区大小(字节)
	FlushThreshold int `json:"flushThreshold" form:"flushThreshold" query:"flushThreshold" db:"flushThreshold"` // 刷新阈值(条目数)
	
	// 配置优先级
	ConfigPriority int `json:"configPriority" form:"configPriority" query:"configPriority" db:"configPriority"` // 配置优先级,数值越小优先级越高
	
	// 预留字段
	Reserved1 string     `json:"reserved1" form:"reserved1" query:"reserved1" db:"reserved1"` // 预留字段1
	Reserved2 string     `json:"reserved2" form:"reserved2" query:"reserved2" db:"reserved2"` // 预留字段2
	Reserved3 *int       `json:"reserved3" form:"reserved3" query:"reserved3" db:"reserved3"` // 预留字段3
	Reserved4 *int       `json:"reserved4" form:"reserved4" query:"reserved4" db:"reserved4"` // 预留字段4
	Reserved5 *time.Time `json:"reserved5" form:"reserved5" query:"reserved5" db:"reserved5"` // 预留字段5
	
	// 扩展属性
	ExtProperty string `json:"extProperty" form:"extProperty" query:"extProperty" db:"extProperty"` // 扩展属性,JSON格式
	
	// 标准字段
	AddTime        time.Time `json:"addTime" form:"addTime" query:"addTime" db:"addTime"`                             // 创建时间
	AddWho         string    `json:"addWho" form:"addWho" query:"addWho" db:"addWho"`                                 // 创建人ID
	EditTime       time.Time `json:"editTime" form:"editTime" query:"editTime" db:"editTime"`                         // 最后修改时间
	EditWho        string    `json:"editWho" form:"editWho" query:"editWho" db:"editWho"`                             // 最后修改人ID
	OprSeqFlag     string    `json:"oprSeqFlag" form:"oprSeqFlag" query:"oprSeqFlag" db:"oprSeqFlag"`                 // 操作序列标识
	CurrentVersion int       `json:"currentVersion" form:"currentVersion" query:"currentVersion" db:"currentVersion"` // 当前版本号
	ActiveFlag     string    `json:"activeFlag" form:"activeFlag" query:"activeFlag" db:"activeFlag"`                 // 活动状态标记(N非活动,Y活动)
	NoteText       string    `json:"noteText" form:"noteText" query:"noteText" db:"noteText"`                         // 备注信息
}

// TableName 返回表名
func (LogConfig) TableName() string {
	return "HUB_GW_LOG_CONFIG"
}

 