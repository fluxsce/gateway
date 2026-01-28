// 服务中心实例基础配置类型 - 严格按照HUB_SERVICE_CENTER_CONFIG表结构定义
export interface ServiceCenterInstance {
  // 主键和租户信息
  tenantId: string // 租户ID，联合主键
  instanceName: string // 实例名称，联合主键
  environment: 'DEVELOPMENT' | 'STAGING' | 'PRODUCTION' // 部署环境，联合主键

  // 服务器类型和监听配置
  serverType: 'GRPC' | 'HTTP' // 服务器类型
  listenAddress: string // 监听地址，默认'0.0.0.0'
  listenPort: number // 监听端口，默认12004

  // gRPC 消息大小配置
  maxRecvMsgSize: number // 最大接收消息大小（字节），默认16MB
  maxSendMsgSize: number // 最大发送消息大小（字节），默认16MB

  // gRPC Keep-Alive 配置
  keepAliveTime: number // Keep-alive 发送间隔（秒），默认30
  keepAliveTimeout: number // Keep-alive 超时时间（秒），默认10
  keepAliveMinTime: number // 客户端最小 Keep-alive 间隔（秒），默认15
  permitWithoutStream: 'Y' | 'N' // 是否允许无活跃流时发送 Keep-alive，默认Y

  // gRPC 连接管理配置
  maxConnectionIdle: number // 最大连接空闲时间（秒，0表示无限制），默认0
  maxConnectionAge: number // 最大连接存活时间（秒，0表示无限制），默认0
  maxConnectionAgeGrace: number // 连接关闭宽限期（秒），默认20

  // gRPC 功能开关
  enableReflection: 'Y' | 'N' // 是否启用 gRPC 反射，默认Y
  enableTLS: 'Y' | 'N' // 是否启用 TLS 加密，默认N

  // 证书配置 - 支持文件路径和数据库存储
  certStorageType: 'FILE' | 'DATABASE' // 证书存储类型，默认FILE
  certFilePath?: string // TLS 证书文件路径
  keyFilePath?: string // TLS 私钥文件路径
  certContent?: string // TLS 证书内容（PEM格式）
  keyContent?: string // TLS 私钥内容（PEM格式）
  certChainContent?: string // TLS 证书链内容（PEM格式）
  certPassword?: string // 证书密码（加密存储）
  enableMTLS: 'Y' | 'N' // 是否启用双向 TLS 认证，默认N

  // 性能调优配置
  maxConcurrentStreams: number // 最大并发流数量（0表示无限制），默认250
  readBufferSize: number // 读缓冲区大小（字节），默认32KB
  writeBufferSize: number // 写缓冲区大小（字节），默认32KB

  // 健康检查配置
  healthCheckInterval: number // 健康检查间隔（秒），0表示禁用，默认30
  healthCheckTimeout: number // 健康检查超时时间（秒），默认5

  // 实例状态管理
  instanceStatus: 'STOPPED' | 'STARTING' | 'RUNNING' | 'STOPPING' | 'ERROR' // 实例状态
  statusMessage?: string // 状态消息，记录启动、停止、异常等详细信息
  lastStatusTime?: string // 最后状态变更时间（启动/停止/异常）
  lastHealthCheckTime?: string // 最后健康检查时间

  // 访问控制配置
  enableAuth: 'Y' | 'N' // 是否启用认证，默认N
  ipWhitelist?: string // IP 白名单（JSON 数组格式）
  ipBlacklist?: string // IP 黑名单（JSON 数组格式）

  // 系统字段
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
  extProperty?: string // 扩展属性，JSON格式
}

