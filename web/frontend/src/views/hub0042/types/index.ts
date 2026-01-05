/**
 * Hub0042 JVM监控模块 - 类型定义文件
 * 
 * 本文件定义了JVM监控管理相关的所有TypeScript类型，包括：
 * - JVM资源信息：JvmResource相关类型
 * - 内存监控：JvmMemory、MemoryPool相关类型
 * - GC监控：JvmGc相关类型
 * - 线程监控：JvmThread、ThreadState、Deadlock、ThreadPool相关类型
 * - 类加载监控：JvmClass相关类型
 * - 查询参数、请求响应等辅助类型
 * 
 * @author 系统架构组
 * @version 1.0.0
 * @since 2024-01-01
 */

// ==================== 枚举类型定义 ====================

/**
 * 内存类型枚举
 */
export enum MemoryType {
  /** 堆内存 */
  HEAP = 'HEAP',
  /** 非堆内存 */
  NON_HEAP = 'NON_HEAP'
}

/**
 * 内存池类型枚举
 */
export enum PoolType {
  /** 堆内存 */
  HEAP = 'HEAP',
  /** 非堆内存 */
  NON_HEAP = 'NON_HEAP'
}

/**
 * 内存池分类枚举
 */
export enum PoolCategory {
  /** 年轻代 */
  YOUNG_GENERATION = '年轻代',
  /** 老年代 */
  OLD_GENERATION = '老年代',
  /** 元数据空间 */
  METASPACE = '元数据空间',
  /** 代码缓存 */
  CODE_CACHE = '代码缓存',
  /** 其他 */
  OTHER = '其他'
}

/**
 * 健康等级枚举
 */
export enum HealthGrade {
  /** 优秀 */
  EXCELLENT = 'EXCELLENT',
  /** 良好 */
  GOOD = 'GOOD',
  /** 一般 */
  FAIR = 'FAIR',
  /** 差 */
  POOR = 'POOR',
  /** 严重 */
  CRITICAL = 'CRITICAL'
}

/**
 * 线程状态枚举
 */
export enum ThreadState {
  /** 新建 */
  NEW = 'NEW',
  /** 可运行 */
  RUNNABLE = 'RUNNABLE',
  /** 阻塞 */
  BLOCKED = 'BLOCKED',
  /** 等待 */
  WAITING = 'WAITING',
  /** 限时等待 */
  TIMED_WAITING = 'TIMED_WAITING',
  /** 终止 */
  TERMINATED = 'TERMINATED'
}

/**
 * 严重程度枚举
 */
export enum SeverityLevel {
  /** 低 */
  LOW = 'LOW',
  /** 中等 */
  MEDIUM = 'MEDIUM',
  /** 高 */
  HIGH = 'HIGH',
  /** 严重 */
  CRITICAL = 'CRITICAL'
}

/**
 * 告警级别枚举
 */
export enum AlertLevel {
  /** 信息 */
  INFO = 'INFO',
  /** 警告 */
  WARNING = 'WARNING',
  /** 错误 */
  ERROR = 'ERROR',
  /** 严重 */
  CRITICAL = 'CRITICAL',
  /** 紧急 */
  EMERGENCY = 'EMERGENCY'
}

/**
 * 线程池类型枚举
 */
export enum ThreadPoolType {
  /** 固定线程池 */
  FIXED = 'FixedThreadPool',
  /** 缓存线程池 */
  CACHED = 'CachedThreadPool',
  /** 调度线程池 */
  SCHEDULED = 'ScheduledThreadPool',
  /** 自定义 */
  CUSTOM = 'Custom'
}

/**
 * 健康检查类型枚举
 */
export enum HealthCheckType {
  /** HTTP检查 */
  HTTP = 'HTTP',
  /** TCP检查 */
  TCP = 'TCP'
}

// ==================== 核心实体类型 ====================

/**
 * JVM资源信息接口
 * 对应后端 HUB_MONITOR_JVM_RESOURCE 表
 */
export interface JvmResource {
  /** JVM资源记录ID（由应用端生成的唯一标识），主键 */
  jvmResourceId: string
  /** 租户ID */
  tenantId: string
  
  // 应用标识信息
  /** 应用名称 */
  applicationName: string
  /** 分组名称 */
  groupName: string
  /** 服务分组ID */
  serviceGroupId: string
  /** 主机名 */
  hostName?: string
  /** 主机IP地址 */
  hostIpAddress?: string
  
  // 时间相关字段
  /** 数据采集时间 */
  collectionTime: string
  /** JVM启动时间 */
  jvmStartTime: string
  /** JVM运行时长（毫秒） */
  jvmUptimeMs: number
  
  // 健康状态字段
  /** JVM整体健康标记(Y健康,N异常) */
  healthyFlag: string
  /** JVM健康等级 */
  healthGrade?: HealthGrade
  /** 是否需要立即关注(Y是,N否) */
  requiresAttentionFlag: string
  /** 监控摘要信息 */
  summaryText?: string
  
  // 系统属性（JSON格式）
  /** JVM系统属性，JSON格式 */
  systemPropertiesJson?: string
  
  // 通用字段
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho?: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho?: string
  /** 操作序列标识 */
  oprSeqFlag?: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记(N非活动,Y活动) */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
}

/**
 * JVM内存信息接口
 * 对应后端 HUB_MONITOR_JVM_MEMORY 表
 */
export interface JvmMemory {
  /** JVM内存记录ID，主键 */
  jvmMemoryId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 内存类型
  /** 内存类型(HEAP/NON_HEAP) */
  memoryType: MemoryType
  
  // 内存使用情况（字节）
  /** 初始内存大小（字节） */
  initMemoryBytes: number
  /** 已使用内存大小（字节） */
  usedMemoryBytes: number
  /** 已提交内存大小（字节） */
  committedMemoryBytes: number
  /** 最大内存大小（字节），-1表示无限制 */
  maxMemoryBytes: number
  
  // 计算指标
  /** 内存使用率（百分比） */
  usagePercent: number
  /** 内存健康标记(Y健康,N异常) */
  healthyFlag: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 内存池信息接口
 * 对应后端 HUB_MONITOR_JVM_MEM_POOL 表
 */
export interface MemoryPool {
  /** 内存池记录ID，主键 */
  memoryPoolId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 内存池基本信息
  /** 内存池名称 */
  poolName: string
  /** 内存池类型(HEAP/NON_HEAP) */
  poolType: PoolType
  /** 内存池分类 */
  poolCategory?: string
  
  // 当前使用情况
  /** 当前初始内存（字节） */
  currentInitBytes: number
  /** 当前已使用内存（字节） */
  currentUsedBytes: number
  /** 当前已提交内存（字节） */
  currentCommittedBytes: number
  /** 当前最大内存（字节） */
  currentMaxBytes: number
  /** 当前使用率（百分比） */
  currentUsagePercent: number
  
  // 峰值使用情况
  /** 峰值初始内存（字节） */
  peakInitBytes?: number
  /** 峰值已使用内存（字节） */
  peakUsedBytes?: number
  /** 峰值已提交内存（字节） */
  peakCommittedBytes?: number
  /** 峰值最大内存（字节） */
  peakMaxBytes?: number
  /** 峰值使用率（百分比） */
  peakUsagePercent?: number
  
  // 阈值监控
  /** 是否支持使用阈值监控(Y是,N否) */
  usageThresholdSupported: string
  /** 使用阈值（字节） */
  usageThresholdBytes?: number
  /** 使用阈值超越次数 */
  usageThresholdCount?: number
  /** 是否支持收集使用量监控(Y是,N否) */
  collectionUsageSupported: string
  
  // 健康状态
  /** 内存池健康标记(Y健康,N异常) */
  healthyFlag: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * GC快照接口
 * 对应后端 HUB_MONITOR_JVM_GC 表
 */
export interface JvmGc {
  /** GC快照记录ID，主键 */
  gcSnapshotId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // GC累积统计
  /** GC总次数（累积，所有GC收集器汇总） */
  collectionCount: number
  /** GC总耗时（毫秒，累积，所有GC收集器汇总） */
  collectionTimeMs: number
  
  // jstat -gc 风格的内存区域数据（单位：KB）
  /** Survivor 0 区容量（KB） */
  s0c?: number
  /** Survivor 1 区容量（KB） */
  s1c?: number
  /** Survivor 0 区使用量（KB） */
  s0u?: number
  /** Survivor 1 区使用量（KB） */
  s1u?: number
  /** Eden 区容量（KB） */
  ec?: number
  /** Eden 区使用量（KB） */
  eu?: number
  /** Old 区容量（KB） */
  oc?: number
  /** Old 区使用量（KB） */
  ou?: number
  /** Metaspace 容量（KB） */
  mc?: number
  /** Metaspace 使用量（KB） */
  mu?: number
  /** 压缩类空间容量（KB） */
  ccsc?: number
  /** 压缩类空间使用量（KB） */
  ccsu?: number
  
  // GC统计（jstat -gc 格式）
  /** 年轻代GC次数 */
  ygc?: number
  /** 年轻代GC总时间（秒） */
  ygct?: number
  /** Full GC次数 */
  fgc?: number
  /** Full GC总时间（秒） */
  fgct?: number
  /** 总GC时间（秒） */
  gct?: number
  
  // 时间戳信息
  /** 数据采集时间戳 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 线程信息接口
 * 对应后端 HUB_MONITOR_JVM_THREAD 表
 */
export interface JvmThread {
  /** JVM线程记录ID，主键 */
  jvmThreadId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 基础线程统计
  /** 当前线程数 */
  currentThreadCount: number
  /** 守护线程数 */
  daemonThreadCount: number
  /** 用户线程数 */
  userThreadCount: number
  /** 峰值线程数 */
  peakThreadCount: number
  /** 总启动线程数 */
  totalStartedThreadCount: number
  
  // 性能指标
  /** 线程增长率（百分比） */
  threadGrowthRatePercent?: number
  /** 守护线程比例（百分比） */
  daemonThreadRatioPercent?: number
  
  // 监控功能支持状态
  /** CPU时间监控是否支持(Y是,N否) */
  cpuTimeSupported: string
  /** CPU时间监控是否启用(Y是,N否) */
  cpuTimeEnabled: string
  /** 内存分配监控是否支持(Y是,N否) */
  memoryAllocSupported: string
  /** 内存分配监控是否启用(Y是,N否) */
  memoryAllocEnabled: string
  /** 争用监控是否支持(Y是,N否) */
  contentionSupported: string
  /** 争用监控是否启用(Y是,N否) */
  contentionEnabled: string
  
  // 健康状态
  /** 线程健康标记(Y健康,N异常) */
  healthyFlag: string
  /** 线程健康等级 */
  healthGrade?: HealthGrade
  /** 是否需要立即关注(Y是,N否) */
  requiresAttentionFlag: string
  /** 潜在问题列表，JSON格式 */
  potentialIssuesJson?: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 线程状态统计接口
 * 对应后端 HUB_MONITOR_JVM_THR_STATE 表
 */
export interface JvmThreadState {
  /** 线程状态记录ID，主键 */
  threadStateId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM线程记录ID */
  jvmThreadId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 线程状态分布
  /** NEW状态线程数 */
  newThreadCount: number
  /** RUNNABLE状态线程数 */
  runnableThreadCount: number
  /** BLOCKED状态线程数 */
  blockedThreadCount: number
  /** WAITING状态线程数 */
  waitingThreadCount: number
  /** TIMED_WAITING状态线程数 */
  timedWaitingThreadCount: number
  /** TERMINATED状态线程数 */
  terminatedThreadCount: number
  /** 总线程数 */
  totalThreadCount: number
  
  // 比例指标
  /** 活跃线程比例（百分比） */
  activeThreadRatioPercent?: number
  /** 阻塞线程比例（百分比） */
  blockedThreadRatioPercent?: number
  /** 等待状态线程比例（百分比） */
  waitingThreadRatioPercent?: number
  
  // 健康状态
  /** 线程状态健康标记(Y健康,N异常) */
  healthyFlag: string
  /** 健康等级 */
  healthGrade?: HealthGrade
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 死锁检测信息接口
 * 对应后端 HUB_MONITOR_JVM_DEADLOCK 表
 */
export interface JvmDeadlock {
  /** 死锁记录ID，主键 */
  deadlockId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM线程记录ID */
  jvmThreadId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 死锁基本信息
  /** 是否检测到死锁(Y是,N否) */
  hasDeadlockFlag: string
  /** 死锁线程数量 */
  deadlockThreadCount: number
  /** 死锁线程ID列表，逗号分隔 */
  deadlockThreadIds?: string
  /** 死锁线程名称列表，逗号分隔 */
  deadlockThreadNames?: string
  
  // 死锁严重程度
  /** 严重程度 */
  severityLevel?: SeverityLevel
  /** 严重程度描述 */
  severityDescription?: string
  /** 影响的线程组数量 */
  affectedThreadGroups?: number
  
  // 时间信息
  /** 死锁检测时间 */
  detectionTime?: string
  /** 死锁持续时间（毫秒） */
  deadlockDurationMs?: number
  /** 数据采集时间 */
  collectionTime: string
  
  // 诊断信息
  /** 死锁描述信息 */
  descriptionText?: string
  /** 建议的解决方案 */
  recommendedAction?: string
  /** 告警级别 */
  alertLevel?: AlertLevel
  /** 是否需要立即处理(Y是,N否) */
  requiresActionFlag: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

// ==================== 应用监控数据类型 ====================

/**
 * 应用监控数据类型枚举
 */
export enum AppDataType {
  /** 线程池 */
  THREAD_POOL = 'THREAD_POOL',
  /** HTTP连接池 */
  HTTP_CONNECTION_POOL = 'HTTP_CONNECTION_POOL',
  /** Druid数据源 */
  DRUID_DATASOURCE = 'DRUID_DATASOURCE',
  /** 连接池 */
  CONNECTION_POOL = 'CONNECTION_POOL',
  /** 自定义指标 */
  CUSTOM_METRIC = 'CUSTOM_METRIC',
  /** 缓存池 */
  CACHE_POOL = 'CACHE_POOL',
  /** 消息队列 */
  MESSAGE_QUEUE = 'MESSAGE_QUEUE'
}

/**
 * 应用监控数据分类枚举
 */
export enum AppDataCategory {
  /** 业务线程池 */
  BUSINESS = 'BUSINESS',
  /** IO线程池 */
  IO = 'IO',
  /** 第三方组件 */
  THIRD_PARTY = 'THIRD_PARTY',
  /** 业务指标 */
  BUSINESS_METRIC = 'BUSINESS_METRIC',
  /** 技术指标 */
  TECHNICAL_METRIC = 'TECHNICAL_METRIC'
}

/**
 * 应用监控数据接口
 * 对应后端 HUB_MONITOR_APP_DATA 表
 */
export interface AppMonitorData {
  /** 应用监控数据ID，主键 */
  appDataId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 数据分类标识
  /** 数据类型 */
  dataType: AppDataType | string
  /** 数据名称（如：线程池名称、指标名称等） */
  dataName: string
  /** 数据分类（如：业务线程池/IO线程池/业务指标/技术指标） */
  dataCategory?: AppDataCategory | string
  
  // 监控数据（JSON格式存储，支持不同类型的数据结构）
  /** 监控数据，JSON格式，包含具体的监控指标和值 */
  dataJson: string
  
  // 核心指标（从JSON中提取的关键指标，便于查询和索引）
  /** 主要指标值（如：使用率、数量等） */
  primaryValue?: number
  /** 次要指标值（如：最大值、平均值等） */
  secondaryValue?: number
  /** 状态值（如：健康状态、连接状态等） */
  statusValue?: string
  
  // 健康状态
  /** 健康标记(Y健康,N异常) */
  healthyFlag: string
  /** 健康等级 */
  healthGrade?: HealthGrade
  /** 是否需要立即关注(Y是,N否) */
  requiresAttentionFlag: string
  
  // 标签和维度（便于分组查询）
  /** 标签信息，JSON格式 */
  tagsJson?: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 线程池数据JSON结构（存储在dataJson字段中）
 * 对应后端 DefaultThreadPoolMonitorProvider
 */
export interface ThreadPoolMonitorData {
  /** 活跃线程数 */
  activeCount: number
  /** 活跃线程比例 */
  activeThreadRatio: number
  /** 是否允许核心线程超时 */
  allowsCoreThreadTimeOut: boolean
  /** 数据采集时间戳 */
  collectTime: number
  /** 已完成任务数 */
  completedTaskCount: number
  /** 核心线程数 */
  corePoolSize: number
  /** 是否已关闭 */
  isShutdown: boolean
  /** 是否已终止 */
  isTerminated: boolean
  /** 是否正在终止 */
  isTerminating: boolean
  /** 线程空闲存活时间（秒） */
  keepAliveTime: number
  /** 历史最大线程数 */
  largestPoolSize: number
  /** 最大线程数 */
  maximumPoolSize: number
  /** 待处理任务数 */
  pendingTaskCount: number
  /** 当前线程数 */
  poolSize: number
  /** 队列总容量 */
  queueCapacity: number
  /** 队列剩余容量 */
  queueRemainingCapacity: number
  /** 队列当前大小 */
  queueSize: number
  /** 队列使用比例 */
  queueUsageRatio: number
  /** 总任务数（已提交） */
  taskCount: number
}

/**
 * HTTP连接池监控数据JSON结构
 * 对应后端 PoolingHttpClientConnectionManagerMonitorProvider
 */
export interface HttpConnectionPoolMonitorData {
  /** 活跃路由数量 */
  activeRouteCount: number
  /** 可用连接比例 */
  availableConnectionRatio: number
  /** 数据采集时间戳 */
  collectTime: number
  /** 连接使用比例 */
  connectionUsageRatio: number
  /** 每个路由的默认最大连接数 */
  defaultMaxPerRoute: number
  /** 最大总连接数 */
  maxTotal: number
  /** 待处理请求比例 */
  pendingRequestRatio: number
  /** 路由统计信息 */
  routeStats: {
    total: {
      available: number
      leased: number
      max: number
      pending: number
    }
  }
  /** 总可用连接数 */
  totalAvailable: number
  /** 总连接数 */
  totalConnections: number
  /** 总租用连接数 */
  totalLeased: number
  /** 总最大连接数 */
  totalMax: number
  /** 总待处理请求数 */
  totalPending: number
  /** 空闲后验证时间（毫秒） */
  validateAfterInactivity: number
}

/**
 * Druid数据源监控数据JSON结构
 * 对应后端 DruidDataSourceMonitorProvider
 */
export interface DruidDataSourceMonitorData {
  /** 活跃连接比例 */
  activeConnectionRatio: number
  /** 活跃连接数 */
  activeCount: number
  /** 关闭连接数 */
  closeCount: number
  /** 是否已关闭 */
  closed: boolean
  /** 数据采集时间戳 */
  collectTime: number
  /** 提交次数 */
  commitCount: number
  /** 连接次数 */
  connectCount: number
  /** 连接错误次数 */
  connectErrorCount: number
  /** 驱动类名 */
  driverClassName: string
  /** 是否启用 */
  enable: boolean
  /** 错误次数 */
  errorCount: number
  /** 错误率 */
  errorRate: number
  /** 批量执行次数 */
  executeBatchCount: number
  /** 执行次数 */
  executeCount: number
  /** 查询执行次数 */
  executeQueryCount: number
  /** 更新执行次数 */
  executeUpdateCount: number
  /** 是否已初始化 */
  inited: boolean
  /** 初始连接数 */
  initialSize: number
  /** 最大活跃连接数 */
  maxActive: number
  /** 最大等待时间（毫秒） */
  maxWait: number
  /** 最小空闲连接数 */
  minIdle: number
  /** 非空等待次数 */
  notEmptyWaitCount: number
  /** 非空等待时间（毫秒） */
  notEmptyWaitMillis: number
  /** 连接池使用比例 */
  poolUsageRatio: number
  /** 池化连接数 */
  poolingCount: number
  /** 回收次数 */
  recycleCount: number
  /** 移除废弃连接次数 */
  removeAbandonedCount: number
  /** 回滚次数 */
  rollbackCount: number
  /** 连接URL */
  url: string
  /** 用户名 */
  username: string
  /** 等待线程数 */
  waitThreadCount: number
}

/**
 * 线程池监控接口
 * 对应后端 HUB_MONITOR_JVM_THREADPOOL 表
 */
export interface JvmThreadPool {
  /** 线程池记录ID，主键 */
  threadPoolId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 线程池标识信息
  /** 线程池名称 */
  poolName: string
  /** 线程池类型 */
  poolType?: string
  
  // 基本线程信息
  /** 核心线程数 */
  corePoolSize: number
  /** 最大线程数 */
  maximumPoolSize: number
  /** 当前线程数 */
  poolSize: number
  /** 活跃线程数 */
  activeCount: number
  /** 历史最大线程数 */
  largestPoolSize: number
  
  // 任务统计信息
  /** 总任务数（已提交） */
  taskCount: number
  /** 已完成任务数 */
  completedTaskCount: number
  /** 待处理任务数 */
  pendingTaskCount: number
  
  // 队列信息
  /** 队列当前大小 */
  queueSize: number
  /** 队列剩余容量 */
  queueRemainingCapacity: number
  /** 队列总容量 */
  queueCapacity: number
  /** 队列类型 */
  queueType?: string
  
  // 使用率指标（百分比）
  /** 活跃线程比例（百分比） */
  activeThreadRatioPercent: number
  /** 队列使用比例（百分比） */
  queueUsageRatioPercent: number
  /** 线程池利用率（百分比） */
  poolUtilizationPercent: number
  
  // 线程池状态
  /** 是否已关闭(Y是,N否) */
  isShutdown: string
  /** 是否已终止(Y是,N否) */
  isTerminated: string
  /** 是否正在终止(Y是,N否) */
  isTerminating: string
  
  // 性能配置
  /** 线程空闲存活时间（毫秒） */
  keepAliveTimeMs?: number
  /** 是否允许核心线程超时(Y是,N否) */
  allowsCoreThreadTimeOut: string
  /** 拒绝策略 */
  rejectedExecutionHandler?: string
  
  // 性能指标
  /** 任务吞吐量（任务数/秒） */
  taskThroughputPerSecond?: number
  /** 平均任务完成率（百分比） */
  averageTaskCompletionRate?: number
  /** 任务拒绝次数 */
  taskRejectionCount?: number
  
  // 健康状态
  /** 线程池健康标记(Y健康,N异常) */
  healthyFlag: string
  /** 健康等级 */
  healthGrade?: HealthGrade
  /** 是否需要立即关注(Y是,N否) */
  requiresAttentionFlag: string
  
  // 健康评估阈值
  /** 活跃线程健康阈值（百分比） */
  healthyActiveThreadThreshold?: number
  /** 队列使用健康阈值（百分比） */
  healthyQueueThreshold?: number
  
  // 问题诊断
  /** 潜在问题列表，JSON格式 */
  potentialIssuesJson?: string
  /** 优化建议，JSON格式 */
  recommendationsJson?: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

/**
 * 类加载信息接口
 * 对应后端 HUB_MONITOR_JVM_CLASS 表
 */
export interface JvmClass {
  /** 类加载记录ID，主键 */
  classLoadingId: string
  /** 租户ID */
  tenantId: string
  /** 关联的JVM资源ID */
  jvmResourceId: string
  
  // 类加载统计
  /** 当前已加载类数量 */
  loadedClassCount: number
  /** 总加载类数量 */
  totalLoadedClassCount: number
  /** 已卸载类数量 */
  unloadedClassCount: number
  
  // 比例指标
  /** 类卸载率（百分比） */
  classUnloadRatePercent?: number
  /** 类保留率（百分比） */
  classRetentionRatePercent?: number
  
  // 配置状态
  /** 是否启用详细类加载输出(Y是,N否) */
  verboseClassLoading: string
  
  // 性能指标
  /** 每小时平均类加载数量 */
  loadingRatePerHour?: number
  /** 类加载效率 */
  loadingEfficiency?: number
  /** 内存使用效率评估 */
  memoryEfficiency?: string
  /** 类加载器健康状况 */
  loaderHealth?: string
  
  // 健康状态
  /** 类加载健康标记(Y健康,N异常) */
  healthyFlag: string
  /** 健康等级 */
  healthGrade?: HealthGrade
  /** 是否需要立即关注(Y是,N否) */
  requiresAttentionFlag: string
  /** 潜在问题列表，JSON格式 */
  potentialIssuesJson?: string
  
  // 时间字段
  /** 数据采集时间 */
  collectionTime: string
  
  // 通用字段
  addTime: string
  addWho?: string
  editTime: string
  editWho?: string
  oprSeqFlag?: string
  currentVersion: number
  activeFlag: string
  noteText?: string
}

// ==================== 查询参数类型 ====================

/**
 * JVM资源查询请求参数
 */
export interface JvmResourceQueryRequest {
  /** 租户ID */
  tenantId?: string
  /** 应用名称（模糊查询） */
  applicationName?: string
  /** 分组名称（模糊查询） */
  groupName?: string
  /** 服务分组ID */
  serviceGroupId?: string
  /** 主机IP地址 */
  hostIpAddress?: string
  /** 健康标记 */
  healthyFlag?: string
  /** 是否需要关注 */
  requiresAttentionFlag?: string
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
  /** 页码 */
  pageNum?: number
  /** 每页大小 */
  pageSize?: number
}

/**
 * 内存监控查询请求参数
 */
export interface MemoryQueryRequest {
  /** JVM资源ID */
  jvmResourceId: string
  /** 租户ID */
  tenantId?: string
  /** 内存类型 */
  memoryType?: MemoryType
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
}

/**
 * GC监控查询请求参数
 */
export interface GcQueryRequest {
  /** JVM资源ID */
  jvmResourceId: string
  /** 租户ID */
  tenantId?: string
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
  /** 限制条数 */
  limit?: number
}

/**
 * 线程监控查询请求参数
 */
export interface ThreadQueryRequest {
  /** JVM资源ID */
  jvmResourceId: string
  /** 租户ID */
  tenantId?: string
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
}

/**
 * 应用监控数据查询请求参数
 */
export interface AppDataQueryRequest {
  /** JVM资源ID */
  jvmResourceId: string
  /** 租户ID */
  tenantId?: string
  /** 数据类型 */
  dataType?: AppDataType | string
  /** 数据名称（模糊查询） */
  dataName?: string
  /** 数据分类 */
  dataCategory?: AppDataCategory | string
  /** 健康标记 */
  healthyFlag?: string
  /** 是否需要关注 */
  requiresAttentionFlag?: string
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
}

/**
 * 线程池查询请求参数
 */
export interface ThreadPoolQueryRequest {
  /** JVM资源ID */
  jvmResourceId: string
  /** 租户ID */
  tenantId?: string
  /** 线程池名称 */
  poolName?: string
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string
}

// ==================== 响应数据类型 ====================

/**
 * JVM监控详情数据（整合了所有监控信息）
 */
export interface JvmMonitorDetail {
  /** JVM资源基本信息 */
  resource: JvmResource
  /** 内存信息（堆内存+非堆内存） */
  memory: JvmMemory[]
  /** 内存池信息 */
  memoryPools: MemoryPool[]
  /** GC快照 */
  gc: JvmGc
  /** 线程信息 */
  thread: JvmThread
  /** 线程状态统计 */
  threadState: JvmThreadState
  /** 死锁检测信息 */
  deadlock: JvmDeadlock
  /** 线程池列表 */
  threadPools: JvmThreadPool[]
  /** 类加载信息 */
  classLoading: JvmClass
}

/**
 * GC趋势数据点
 */
export interface GcTrendData {
  /** 时间 */
  time: string
  /** GC次数增量 */
  gcCountIncrease: number
  /** GC耗时增量（毫秒） */
  gcTimeIncrease: number
  /** 年轻代GC次数增量 */
  ygcIncrease: number
  /** Full GC次数增量 */
  fgcIncrease: number
  /** 时间间隔（秒） */
  intervalSeconds: number
}

/**
 * 内存趋势数据点
 */
export interface MemoryTrendData {
  /** 时间 */
  time: string
  /** 内存类型 */
  memoryType: MemoryType
  /** 使用量（字节） */
  usedMemoryBytes: number
  /** 使用率（百分比） */
  usagePercent: number
}

/**
 * 线程池趋势数据
 */
export interface ThreadPoolTrendData {
  /** 线程池名称 */
  poolName: string
  /** 小时 */
  hour: string
  /** 平均活跃线程比例 */
  avgActiveRatio: number
  /** 平均队列使用率 */
  avgQueueUsage: number
  /** 最大活跃线程比例 */
  maxActiveRatio: number
  /** 最大队列使用率 */
  maxQueueUsage: number
}

/**
 * 线程趋势数据点
 */
export interface ThreadTrendData {
  /** 时间 */
  time: string
  /** 当前线程数 */
  currentThreadCount: number
  /** 峰值线程数 */
  peakThreadCount: number
  /** 守护线程数 */
  daemonThreadCount: number
  /** 用户线程数 */
  userThreadCount: number
  /** 活跃线程数（RUNNABLE） */
  activeThreadCount: number
  /** 阻塞线程数 */
  blockedThreadCount: number
  /** 等待线程数（WAITING + TIMED_WAITING） */
  waitingThreadCount: number
  /** 活跃线程比例 */
  activeThreadRatioPercent: number
  /** 阻塞线程比例 */
  blockedThreadRatioPercent: number
  /** 等待线程比例 */
  waitingThreadRatioPercent: number
}

/**
 * 线程状态趋势数据点
 */
export interface ThreadStateTrendData {
  /** 时间 */
  time: string
  /** NEW状态线程数 */
  newThreadCount: number
  /** RUNNABLE状态线程数 */
  runnableThreadCount: number
  /** BLOCKED状态线程数 */
  blockedThreadCount: number
  /** WAITING状态线程数 */
  waitingThreadCount: number
  /** TIMED_WAITING状态线程数 */
  timedWaitingThreadCount: number
  /** TERMINATED状态线程数 */
  terminatedThreadCount: number
  /** 总线程数 */
  totalThreadCount: number
}

/**
 * 统计卡片数据
 */
export interface MonitorStatCard {
  /** 标题 */
  title: string
  /** 值 */
  value: string | number
  /** 单位 */
  unit?: string
  /** 趋势（up/down/stable） */
  trend?: 'up' | 'down' | 'stable'
  /** 趋势值 */
  trendValue?: string
  /** 健康状态 */
  status?: 'success' | 'warning' | 'error' | 'info'
  /** 图标 */
  icon?: string
}

