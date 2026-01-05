// 网关实例基础配置类型 - 严格按照HUB_GATEWAY_INSTANCE表结构定义
export interface GatewayInstance {
  // 基础标识字段
  tenantId: string // 租户ID
  gatewayInstanceId: string // 网关实例ID
  instanceName: string // 实例名称
  instanceDesc?: string // 实例描述
  bindAddress: string // 绑定地址，默认'0.0.0.0'

  // HTTP/HTTPS 端口配置
  httpPort?: number // HTTP监听端口
  httpsPort?: number // HTTPS监听端口
  tlsEnabled: 'Y' | 'N' // 是否启用TLS(N否,Y是)

  // 证书配置 - 支持文件路径和数据库存储
  certStorageType: 'FILE' | 'DATABASE' // 证书存储类型(FILE文件,DATABASE数据库)
  certFilePath?: string // 证书文件路径
  keyFilePath?: string // 私钥文件路径
  certContent?: string // 证书内容(PEM格式)
  keyContent?: string // 私钥内容(PEM格式)
  certChainContent?: string // 证书链内容(PEM格式)
  certPassword?: string // 证书密码(加密存储)

  // Go HTTP Server 核心配置
  maxConnections: number // 最大连接数，默认10000
  readTimeoutMs: number // 读取超时时间(毫秒)，默认30000
  writeTimeoutMs: number // 写入超时时间(毫秒)，默认30000
  idleTimeoutMs: number // 空闲连接超时时间(毫秒)，默认60000
  maxHeaderBytes: number // 最大请求头字节数(默认1MB)，默认1048576

  // 性能和并发配置
  maxWorkers: number // 最大工作协程数，默认1000
  keepAliveEnabled: 'Y' | 'N' // 是否启用Keep-Alive(N否,Y是)
  tcpKeepAliveEnabled: 'Y' | 'N' // 是否启用TCP Keep-Alive(N否,Y是)
  gracefulShutdownTimeoutMs: number // 优雅关闭超时时间(毫秒)，默认30000

  // TLS安全配置
  enableHttp2: 'Y' | 'N' // 是否启用HTTP/2(N否,Y是)
  tlsVersion: string // TLS协议版本，支持多选，逗号分隔，如'1.2,1.3'
  tlsCipherSuites?: string // TLS密码套件列表,逗号分隔
  disableGeneralOptionsHandler: 'Y' | 'N' // 是否禁用默认OPTIONS处理器(N否,Y是)

  // 关联配置
  logConfigId?: string // 关联的日志配置ID

  // 状态和元数据
  healthStatus: 'Y' | 'N' // 健康状态(N不健康,Y健康)
  lastHeartbeatTime?: string // 最后心跳时间
  instanceMetadata?: Record<string, any> // 实例元数据,JSON格式

  // 系统字段
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息

  // 预留字段
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5
  extProperty?: Record<string, any> // 扩展属性,JSON格式

  // 关联的日志配置
  logConfig?: LogConfig // 关联的日志配置
}


// 日志配置类型 - 严格按照HUB_GATEWAY_LOG_CONFIG表结构定义
export interface LogConfig {
  // 基础标识信息
  tenantId: string // 租户ID，联合主键
  logConfigId: string // 日志配置ID，联合主键
  configName: string // 配置名称
  configDesc: string // 配置描述

  // 日志内容控制
  logFormat: 'JSON' | 'TEXT' | 'CSV' // 日志格式
  recordRequestBody: 'Y' | 'N' // 是否记录请求体
  recordResponseBody: 'Y' | 'N' // 是否记录响应体
  recordHeaders: 'Y' | 'N' // 是否记录请求/响应头
  maxBodySizeBytes: number // 最大记录报文大小(字节)

  // 日志输出目标配置
  outputTargets: string // 输出目标,逗号分隔
  fileConfig?: string // 文件输出配置,JSON格式
  databaseConfig?: string // 数据库输出配置,JSON格式
  mongoConfig?: string // MongoDB输出配置,JSON格式
  elasticsearchConfig?: string // Elasticsearch输出配置,JSON格式
  clickhouseConfig?: string // ClickHouse输出配置,JSON格式

  // 异步和批量处理配置
  enableAsyncLogging: 'Y' | 'N' // 是否启用异步日志
  asyncQueueSize: number // 异步队列大小
  asyncFlushIntervalMs: number // 异步刷新间隔(毫秒)
  enableBatchProcessing: 'Y' | 'N' // 是否启用批量处理
  batchSize: number // 批处理大小
  batchTimeoutMs: number // 批处理超时时间(毫秒)

  // 日志保留和轮转配置
  logRetentionDays: number // 日志保留天数
  enableFileRotation: 'Y' | 'N' // 是否启用文件轮转
  maxFileSizeMB?: number // 最大文件大小(MB)
  maxFileCount?: number // 最大文件数量
  rotationPattern: 'HOURLY' | 'DAILY' | 'WEEKLY' | 'SIZE_BASED' // 轮转模式

  // 敏感数据处理
  enableSensitiveDataMasking: 'Y' | 'N' // 是否启用敏感数据脱敏
  sensitiveFields: string // 敏感字段列表,JSON数组格式
  maskingPattern: string // 脱敏替换模式

  // 性能优化配置
  bufferSize: number // 缓冲区大小(字节)
  flushThreshold: number // 刷新阈值(条目数)

  // 配置优先级
  configPriority: number // 配置优先级,数值越小优先级越高

  // 预留字段
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5

  // 扩展属性
  extProperty?: string // 扩展属性,JSON格式

  // 标准字段
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
}
