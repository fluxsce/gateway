/**
 * Hub0041 服务注册模块 - 类型定义文件
 * 
 * 本文件定义了服务注册管理相关的所有TypeScript类型，包括：
 * - 服务信息：Service相关类型
 * - 服务实例信息：ServiceInstance相关类型
 * - 外部注册中心配置：NacosConfig等类型
 * - 查询参数、请求响应等辅助类型
 * 
 * @author 系统架构组
 * @version 1.1.0
 * @since 2024-01-01
 */

// ==================== 枚举类型定义 ====================

/**
 * 实例状态枚举
 * 服务实例的运行状态
 */
export enum InstanceStatus {
  /** 正常运行 */
  UP = 'UP',
  /** 已停机 */
  DOWN = 'DOWN',
  /** 启动中 */
  STARTING = 'STARTING',
  /** 服务不可用 */
  OUT_OF_SERVICE = 'OUT_OF_SERVICE'
}

/**
 * 健康状态枚举
 * 服务实例的健康检查状态
 */
export enum HealthStatus {
  /** 健康 */
  HEALTHY = 'HEALTHY',
  /** 不健康 */
  UNHEALTHY = 'UNHEALTHY',
  /** 未知状态 */
  UNKNOWN = 'UNKNOWN'
}

/**
 * 客户端类型枚举
 * 注册服务的客户端类型
 */
export enum ClientType {
  /** Java客户端 */
  JAVA = 'JAVA',
  /** .NET客户端 */
  DOTNET = 'DOTNET',
  /** Node.js客户端 */
  NODEJS = 'NODEJS',
  /** Python客户端 */
  PYTHON = 'PYTHON',
  /** Go客户端 */
  GO = 'GO',
  /** 其他类型客户端 */
  OTHER = 'OTHER'
}

/**
 * 协议类型枚举
 * 服务支持的网络协议类型
 */
export enum ProtocolType {
  /** HTTP协议 */
  HTTP = 'HTTP',
  /** HTTPS协议 */
  HTTPS = 'HTTPS',
  /** TCP协议 */
  TCP = 'TCP',
  /** UDP协议 */
  UDP = 'UDP',
  /** GRPC协议 */
  GRPC = 'GRPC'
}

/**
 * 负载均衡策略枚举
 * 定义服务实例间的负载分配策略
 */
export enum LoadBalanceStrategy {
  /** 轮询 */
  ROUND_ROBIN = 'ROUND_ROBIN',
  /** 加权轮询 */
  WEIGHTED_ROUND_ROBIN = 'WEIGHTED_ROUND_ROBIN',
  /** 最少连接 */
  LEAST_CONNECTIONS = 'LEAST_CONNECTIONS',
  /** 随机 */
  RANDOM = 'RANDOM',
  /** IP哈希 */
  IP_HASH = 'IP_HASH'
}

/**
 * 健康检查类型枚举
 * 定义健康检查的协议类型
 */
export enum HealthCheckType {
  /** HTTP协议检查 */
  HTTP = 'HTTP',
  /** TCP协议检查 */
  TCP = 'TCP'
}

/**
 * 健康检查模式枚举
 * 定义健康检查的执行方式
 */
export enum HealthCheckMode {
  /** 主动探测模式 - 服务端主动发起检查 */
  ACTIVE = 'ACTIVE',
  /** 被动模式 - 依赖客户端上报健康状态 */
  PASSIVE = 'PASSIVE'
}

/**
 * 注册类型枚举
 * 定义服务注册的管理方式
 */
export enum RegistryType {
  /** 内部管理 */
  INTERNAL = 'INTERNAL',
  /** Nacos注册中心 */
  NACOS = 'NACOS',
  /** Consul注册中心 */
  CONSUL = 'CONSUL',
  /** Eureka注册中心 */
  EUREKA = 'EUREKA',
  /** ETCD注册中心 */
  ETCD = 'ETCD',
  /** ZooKeeper注册中心 */
  ZOOKEEPER = 'ZOOKEEPER'
}

/**
 * 事件类型枚举
 * 定义服务注册中心的各类事件类型
 * 
 * 包含以下几个类别:
 * 1. 分组相关事件 - 服务分组操作事件
 * 2. 服务相关事件 - 服务操作事件
 * 3. 实例相关事件 - 服务实例操作和状态变更事件
 */
export enum EventType {
  // ============ 分组相关事件 ============
  /** 服务组创建 */
  SERVICE_GROUP_CREATED = 'SERVICE_GROUP_CREATED',
  /** 服务组更新 */
  SERVICE_GROUP_UPDATED = 'SERVICE_GROUP_UPDATED',
  /** 服务组删除 */
  SERVICE_GROUP_DELETED = 'SERVICE_GROUP_DELETED',
  
  // ============ 服务相关事件 ============
  /** 服务注册 */
  SERVICE_REGISTERED = 'SERVICE_REGISTERED',
  /** 服务更新 */
  SERVICE_UPDATED = 'SERVICE_UPDATED',
  /** 服务注销 */
  SERVICE_DEREGISTERED = 'SERVICE_DEREGISTERED',
  
  // ============ 实例相关事件 ============
  /** 实例注册 */
  INSTANCE_REGISTERED = 'INSTANCE_REGISTERED',
  /** 实例注销 */
  INSTANCE_DEREGISTERED = 'INSTANCE_DEREGISTERED',
  /** 实例心跳更新 */
  INSTANCE_HEARTBEAT_UPDATED = 'INSTANCE_HEARTBEAT_UPDATED',
  /** 实例健康状态变更 */
  INSTANCE_HEALTH_CHANGE = 'INSTANCE_HEALTH_CHANGE',
  /** 实例状态变更 */
  INSTANCE_STATUS_CHANGE = 'INSTANCE_STATUS_CHANGE'
}

// ==================== 核心实体类型 ====================

/**
 * 服务事件信息接口
 * 对应后端 ServiceEvent 模型，HUB_REGISTRY_SERVICE_EVENT 表
 * 
 * 用于记录服务注册中心的各类事件，包括服务和实例的变更、状态变化等
 */
export interface ServiceEvent {
  /** 服务事件ID - 主键，事件的唯一标识 */
  serviceEventId: string
  /** 租户ID - 多租户环境下的租户标识 */
  tenantId: string

  // 关联信息 - 用于关联到具体服务和实例
  /** 服务分组ID - 关联服务分组 */
  serviceGroupId: string
  /** 服务实例ID - 关联具体的服务实例，某些事件可能为空 */
  serviceInstanceId: string

  // 事件基本信息（冗余字段，便于查询展示，无需关联表查询）
  /** 分组名称 - 冗余存储的分组名称 */
  groupName: string
  /** 服务名称 - 冗余存储的服务名称 */
  serviceName: string
  /** 主机地址 - 相关实例的主机地址 */
  hostAddress: string
  /** 端口号 - 相关实例的端口号 */
  portNumber: number

  /** 事件类型 - 使用EventType枚举定义的事件类型 */
  eventType: string
  /** 事件来源 - 事件的产生来源，如"用户操作"、"系统自动"等 */
  eventSource: string
  /** 事件产生节点的IP地址 */
  nodeIpAddress: string

  // 事件数据 - 事件的具体内容
  /** 事件数据 - JSON格式的详细事件数据 */
  eventDataJson: string
  /** 事件消息 - 事件的文字描述，便于理解 */
  eventMessage: string

  // 时间信息
  /** 事件发生时间 - 事件实际发生的时间点 */
  eventTime: string

  // 通用审计字段
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记 */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
  /** 扩展属性，JSON格式 */
  extProperty?: string

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: string
  reserved4?: string
  reserved5?: string
  reserved6?: string
  reserved7?: string
  reserved8?: string
  reserved9?: string
  reserved10?: string
}

/**
 * 服务实例信息接口
 * 对应后端 ServiceInstance 模型
 */
export interface ServiceInstance {
  /** 服务实例ID，主键 */
  serviceInstanceId: string
  /** 租户ID，用于多租户数据隔离 */
  tenantId: string

  // 关联服务和分组信息
  /** 服务分组ID，关联服务分组表主键 */
  serviceGroupId: string
  /** 服务名称，冗余字段便于查询 */
  serviceName: string
  /** 分组名称，冗余字段便于查询 */
  groupName: string

  // 网络连接信息
  /** 主机地址 */
  hostAddress: string
  /** 端口号 */
  portNumber: number
  /** 上下文路径 */
  contextPath: string

  // 实例状态信息
  /** 实例状态 */
  instanceStatus: InstanceStatus
  /** 健康状态 */
  healthStatus: HealthStatus

  // 负载均衡配置
  /** 权重值 */
  weightValue: number

  // 客户端信息
  /** 客户端ID */
  clientId?: string
  /** 客户端版本 */
  clientVersion?: string
  /** 客户端类型 */
  clientType: ClientType

  // 元数据和标签
  /** 实例元数据，JSON格式 */
  metadataJson?: string
  /** 实例标签，JSON格式 */
  tagsJson?: string

  // 时间戳信息
  /** 注册时间 */
  registerTime: string
  /** 最后心跳时间 */
  lastHeartbeatTime?: string
  /** 最后健康检查时间 */
  lastHealthCheckTime?: string

  // 通用审计字段
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记 */
  activeFlag: string
  /** 临时实例标记(Y是临时实例,N否) */
  tempInstanceFlag: string
  /** 备注信息 */
  noteText?: string
  /** 扩展属性，JSON格式 */
  extProperty?: string

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: string
  reserved4?: string
  reserved5?: string
  reserved6?: string
  reserved7?: string
  reserved8?: string
  reserved9?: string
  reserved10?: string
}

/**
 * 服务信息接口
 * 对应后端 Service 模型
 */
export interface Service {
  /** 租户ID，用于多租户数据隔离 */
  tenantId: string
  /** 服务名称，主键 */
  serviceName: string

  // 关联分组信息
  /** 服务分组ID，关联服务分组表主键 */
  serviceGroupId: string
  /** 分组名称，冗余字段便于查询 */
  groupName: string

  // 服务基本信息
  /** 服务描述 */
  serviceDescription: string

  // 服务配置
  /** 协议类型 */
  protocolType: ProtocolType
  /** 上下文路径 */
  contextPath: string
  /** 负载均衡策略 */
  loadBalanceStrategy: LoadBalanceStrategy

  // 健康检查配置
  /** 健康检查URL */
  healthCheckUrl: string
  /** 健康检查间隔(秒) */
  healthCheckIntervalSeconds: number
  /** 健康检查超时(秒) */
  healthCheckTimeoutSeconds: number
  /** 健康检查类型(HTTP,TCP) */
  healthCheckType: HealthCheckType
  /** 健康检查模式(ACTIVE:主动探测,PASSIVE:客户端上报) */
  healthCheckMode: HealthCheckMode

  // 注册管理配置
  /** 注册类型(INTERNAL:内部管理,NACOS:Nacos注册中心,CONSUL:Consul,EUREKA:Eureka,ETCD:ETCD,ZOOKEEPER:ZooKeeper) */
  registryType: RegistryType
  /** 外部注册中心配置，JSON格式，仅当registryType非INTERNAL时使用 */
  externalRegistryConfig?: string

  // 服务实例列表
  /** 服务实例列表 */
  instances?: ServiceInstance[]

  // 元数据和标签
  /** 服务元数据，JSON格式 */
  metadataJson?: string
  /** 服务标签，JSON格式 */
  tagsJson?: string

  // 通用审计字段
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记 */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
  /** 扩展属性，JSON格式 */
  extProperty?: string

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: string
  reserved4?: string
  reserved5?: string
  reserved6?: string
  reserved7?: string
  reserved8?: string
  reserved9?: string
  reserved10?: string

  // 统计信息（查询时可能返回）
  /** 服务实例数量 */
  instanceCount?: number
  /** 健康实例数量 */
  healthyInstanceCount?: number
}

// ==================== 查询参数类型 ====================

/**
 * 服务查询请求参数
 * 用于分页查询和条件筛选服务列表
 */
export interface ServiceQueryRequest {
  /** 服务名称 - 可选，支持模糊查询 */
  serviceName?: string
  /** 服务分组ID - 可选，按分组筛选 */
  serviceGroupId?: string
  /** 分组名称 - 可选，按分组名称筛选 */
  groupName?: string
  /** 协议类型 - 可选，按协议类型筛选 */
  protocolType?: ProtocolType
  /** 活动状态 - 可选，Y/N */
  activeFlag?: string
  /** 分页页码，从1开始 */
  pageIndex?: number
  /** 每页大小，默认20 */
  pageSize?: number
}

/**
 * 服务实例查询请求参数
 * 用于分页查询和条件筛选服务实例列表
 */
export interface ServiceInstanceQueryRequest {
  /** 服务名称 - 可选，按服务名称筛选 */
  serviceName?: string
  /** 服务分组ID - 可选，按分组筛选 */
  serviceGroupId?: string
  /** 主机地址 - 可选，支持模糊查询 */
  hostAddress?: string
  /** 实例状态 - 可选，按状态筛选 */
  instanceStatus?: InstanceStatus
  /** 健康状态 - 可选，按健康状态筛选 */
  healthStatus?: HealthStatus
  /** 客户端类型 - 可选，按客户端类型筛选 */
  clientType?: ClientType
  /** 活动状态 - 可选，Y/N */
  activeFlag?: string
  /** 分页页码，从1开始 */
  pageIndex?: number
  /** 每页大小，默认20 */
  pageSize?: number
}

// ==================== 表格操作类型 ====================

/**
 * 表格操作类型枚举
 * 定义表格行操作的类型
 */
export type TableAction = 
  | 'view'           // 查看详情
  | 'refresh'        // 刷新状态
  | 'health-check'   // 健康检查
  | 'metadata'       // 查看元数据

/**
 * 服务详情扩展信息
 * 包含实例列表和统计信息
 */
export interface ServiceDetail extends Service {
  /** 服务实例列表 */
  instances: ServiceInstance[]
  /** 统计信息 */
  statistics: {
    /** 总实例数 */
    totalInstances: number
    /** 健康实例数 */
    healthyInstances: number
    /** 不健康实例数 */
    unhealthyInstances: number
    /** UP状态实例数 */
    upInstances: number
    /** DOWN状态实例数 */
    downInstances: number
  }
}

// ==================== 外部注册中心配置类型 ====================

/**
 * Nacos服务器配置接口
 * 对应后端 ServerConfig 模型
 * 
 * 用于配置单个Nacos服务器的连接参数
 */
export interface ServerConfig {
  /** 服务器地址（必填） */
  host: string
  /** 服务器端口（可选，默认8848） */
  port?: number
  /** GRPC端口（可选，默认为HTTP端口+1000） */
  grpcPort?: number
  /** 上下文路径（可选，默认"/nacos"） */
  contextPath?: string
  /** 协议（可选，默认"http"） */
  scheme?: 'http' | 'https'
}

/**
 * Nacos配置接口
 * 对应后端 NacosConfig 模型
 * 
 * 用于配置Nacos注册中心的连接参数
 */
export interface NacosConfig {
  // === 服务器配置 ===
  /** 服务器列表（必填）- 单机部署时包含一个服务器，集群部署时包含多个服务器 */
  servers: ServerConfig[]
  
  // === 命名空间和分组 ===
  /** 命名空间（可选，默认"public"） - 用于环境隔离 */
  namespace?: string
  /** 默认分组（可选，默认"DEFAULT_GROUP"） */
  group?: string
  
  // === 认证配置 ===
  /** 用户名（可选，认证时需要） */
  username?: string
  /** 密码（可选，认证时需要） */
  password?: string
  /** 访问密钥（可选，用于阿里云MSE等云服务） */
  accessKey?: string
  /** 密钥（可选，用于阿里云MSE等云服务） */
  secretKey?: string
  
  // === 网络配置 ===
  /** 超时时间（秒）（可选，默认5秒） */
  timeout?: number
  /** 心跳间隔（秒）（可选，默认5秒） */
  beatInterval?: number
  
  // === 缓存配置 ===
  /** 本地缓存目录（可选，默认"/tmp/nacos/cache"） */
  cacheDir?: string
  /** 启动时不加载缓存（可选，默认true） */
  notLoadCacheAtStart?: boolean
  /** 禁用快照缓存（可选，默认false） */
  disableUseSnapShot?: boolean
  /** 空结果时更新缓存（可选，默认false） */
  updateCacheWhenEmpty?: boolean
  
  // === 日志配置 ===
  /** 日志目录（可选，默认"/tmp/nacos/log"） */
  logDir?: string
  /** 日志级别（可选，默认"info"） */
  logLevel?: 'debug' | 'info' | 'warn' | 'error'
  /** 输出到控制台（可选，默认false） */
  appendToStdout?: boolean
  
  // === 性能配置 ===
  /** 更新线程数（可选，默认20） */
  updateThreadNum?: number
  
  // === TLS配置 ===
  /** 启用TLS（可选，默认false） */
  enableTLS?: boolean
  /** 信任所有证书（可选，仅用于测试环境） */
  trustAll?: boolean
  /** CA证书文件路径（可选） */
  caFile?: string
  /** 客户端证书文件路径（可选） */
  certFile?: string
  /** 客户端私钥文件路径（可选） */
  keyFile?: string
  
  // === 应用信息 ===
  /** 应用名称（可选） */
  appName?: string
  /** 应用标识（可选） */
  appKey?: string
  
  // === 高级配置 ===
  /** 启用KMS加密（可选，用于阿里云等云服务） */
  openKMS?: boolean
  /** 地域ID（可选，用于阿里云等云服务） */
  regionId?: string
}

/**
 * 外部注册中心配置接口
 * 根据不同的注册中心类型，包含不同的配置内容
 */
export interface ExternalRegistryConfig {
  /** Nacos注册中心配置 */
  nacos?: NacosConfig
  // 未来可以添加其他注册中心的配置
  // consul?: ConsulConfig
  // eureka?: EurekaConfig
  // etcd?: EtcdConfig
  // zookeeper?: ZookeeperConfig
}

// ==================== 用户信息类型（复用） ====================

/**
 * 用户信息接口
 * 用于用户选择和显示
 */
export interface UserInfo {
  /** 用户ID */
  userId: string
  /** 用户名 */
  userName: string
  /** 真实姓名 */
  realName: string
  /** 邮箱地址 */
  email?: string
}
