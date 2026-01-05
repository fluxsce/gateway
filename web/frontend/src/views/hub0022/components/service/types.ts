/**
 * 服务定义列表组件类型定义
 */

/**
 * 服务定义接口
 * 对应数据库表：HUB_GATEWAY_SERVICE_DEFINITION
 * 用于定义后端服务的完整配置信息
 */
export interface ServiceDefinition {
  /** 租户ID */
  tenantId: string
  /** 服务定义ID */
  serviceDefinitionId: string
  /** 关联的代理配置ID */
  proxyConfigId?: string
  /** 服务名称 */
  serviceName: string
  /** 服务描述 */
  serviceDesc?: string
  /** 服务类型，0静态配置/1服务发现 */
  serviceType: number
  /** 负载均衡策略 */
  loadBalanceStrategy: string
  /** 服务发现类型 */
  discoveryType?: string
  /** 服务发现配置，JSON格式 */
  discoveryConfig?: string
  /** 是否启用会话亲和性 */
  sessionAffinity: 'Y' | 'N'
  /** 是否启用粘性会话 */
  stickySession: 'Y' | 'N'
  /** 最大重试次数 */
  maxRetries: number
  /** 重试超时时间（毫秒） */
  retryTimeoutMs: number
  /** 是否启用熔断器 */
  enableCircuitBreaker: 'Y' | 'N'
  /** 是否启用健康检查 */
  healthCheckEnabled: 'Y' | 'N'
  /** 健康检查路径 */
  healthCheckPath: string
  /** 健康检查方法 */
  healthCheckMethod: string
  /** 健康检查间隔（秒） */
  healthCheckIntervalSeconds: number
  /** 健康检查超时（毫秒） */
  healthCheckTimeoutMs: number
  /** 健康阈值 */
  healthyThreshold: number
  /** 不健康阈值 */
  unhealthyThreshold: number
  /** 期望的状态码，逗号分隔 */
  expectedStatusCodes: string
  /** 健康检查请求头，JSON格式 */
  healthCheckHeaders?: string
  /** 负载均衡器配置，JSON格式 */
  loadBalancerConfig?: string
  /** 服务元数据，JSON格式 */
  serviceMetadata?: string
  /** 活动状态标记 */
  activeFlag: 'Y' | 'N'
  /** 备注信息 */
  noteText?: string
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
}


/**
 * 服务查询参数接口
 * 用于服务定义的分页查询
 */
export interface ServiceQueryParams {
  /** 服务名称（模糊查询） */
  serviceName?: string
  /** 服务类型 */
  serviceType?: number
  /** 负载均衡策略 */
  loadBalanceStrategy?: string
  /** 活动状态 */
  activeFlag?: 'Y' | 'N'
  /** 代理配置ID */
  proxyConfigId?: string
  /** 页码，从1开始 */
  pageIndex?: number
  /** 每页大小 */
  pageSize?: number
}

/**
 * 服务类型枚举
 * 0: 静态配置 - 手动配置服务节点
 * 1: 服务发现 - 通过注册中心自动发现服务
 */
export enum ServiceType {
  STATIC = 0,
  DISCOVERY = 1,
}

/**
 * 负载均衡策略枚举
 * 支持多种负载均衡算法
 */
export enum LoadBalanceStrategy {
  ROUND_ROBIN = 'round-robin', // 轮询算法
  RANDOM = 'random', // 随机算法
  IP_HASH = 'ip-hash', // IP哈希算法
  LEAST_CONN = 'least-conn', // 最少连接算法
  WEIGHTED_ROUND_ROBIN = 'weighted-round-robin', // 加权轮询算法
  CONSISTENT_HASH = 'consistent-hash', // 一致性哈希算法
}
