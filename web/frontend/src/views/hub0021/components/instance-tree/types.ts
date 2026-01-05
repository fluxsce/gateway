/**
 * 网关实例树组件类型定义
 */

import type { TreeOption } from 'naive-ui'

// ============= 枚举定义 =============

/**
 * 过滤器执行模式枚举
 */
export enum FilterExecutionMode {
  /** 顺序执行 */
  SEQUENTIAL = 'SEQUENTIAL',
  /** 并行执行 */
  PARALLEL = 'PARALLEL',
}

/**
 * 网关实例类型
 * 对应数据库表：HUB_GATEWAY_INSTANCE
 * 用于表示网关实例的完整配置信息
 */
export interface GatewayInstance {
  /** 网关实例ID，唯一标识 */
  gatewayInstanceId: string

  /** 实例名称 */
  instanceName: string

  /** 实例描述 */
  instanceDesc?: string

  /** 服务器主机地址 */
  serverHost: string

  /** 绑定地址，通常为 0.0.0.0 或具体IP */
  bindAddress: string

  /** 监听端口 */
  listenPort: number

  /** HTTP端口 */
  httpPort?: number

  /** HTTPS端口 */
  httpsPort?: number

  /** 是否启用TLS，Y启用/N禁用 */
  tlsEnabled: 'Y' | 'N'

  /** 证书文件路径 */
  certFilePath?: string

  /** 私钥文件路径 */
  keyFilePath?: string

  /** 最大连接数 */
  maxConnections: number

  /** 读取超时时间（毫秒） */
  readTimeoutMs: number

  /** 写入超时时间（毫秒） */
  writeTimeoutMs: number

  /** 空闲连接超时时间（毫秒） */
  idleTimeoutMs: number

  /** 健康状态，Y健康/N异常 */
  healthStatus: 'Y' | 'N'

  /** 最后心跳时间 */
  lastHeartbeatTime?: string

  /** 实例元数据，JSON格式字符串 */
  instanceMetadata?: string

  /** 活动状态标记，Y启用/N禁用 */
  activeFlag: 'Y' | 'N'

  /** 创建时间 */
  addTime?: string

  /** 最后修改时间 */
  editTime?: string

  /** 备注信息 */
  noteText?: string
}

/**
 * 实例树节点选项类型
 * 扩展了 naive-ui 的 TreeOption，添加了实例信息
 */
export interface InstanceTreeOption extends TreeOption {
  /** 关联的网关实例对象 */
  instance?: GatewayInstance
}

/**
 * Router配置类型
 * 对应数据库表：HUB_GW_ROUTER_CONFIG
 * 用于表示Router的完整配置信息
 */
export interface RouterConfig {
  /** 租户ID，联合主键 */
  tenantId: string

  /** Router配置ID，联合主键 */
  routerConfigId: string

  /** 关联的网关实例ID */
  gatewayInstanceId: string

  /** Router名称 */
  routerName: string

  /** Router描述 */
  routerDesc?: string

  // Router基础配置

  /** 默认路由优先级 */
  defaultPriority: number

  /** 是否启用路由缓存，Y启用/N禁用 */
  enableRouteCache: 'Y' | 'N'

  /** 路由缓存TTL(秒) */
  routeCacheTtlSeconds: number

  /** 最大路由数量限制 */
  maxRoutes?: number | null

  /** 路由匹配超时时间(毫秒) */
  routeMatchTimeout?: number | null

  // Router高级配置

  /** 是否启用严格模式，Y启用/N禁用 */
  enableStrictMode: 'Y' | 'N'

  /** 是否启用路由指标收集，Y启用/N禁用 */
  enableMetrics: 'Y' | 'N'

  /** 是否启用链路追踪，Y启用/N禁用 */
  enableTracing: 'Y' | 'N'

  /** 路径匹配是否区分大小写，Y是/N否 */
  caseSensitive: 'Y' | 'N'

  /** 是否移除路径尾部斜杠，Y是/N否 */
  removeTrailingSlash: 'Y' | 'N'

  // 路由处理配置

  /** 是否启用全局过滤器，Y启用/N禁用 */
  enableGlobalFilters: 'Y' | 'N'

  /** 过滤器执行模式：SEQUENTIAL顺序，PARALLEL并行 */
  filterExecutionMode: 'SEQUENTIAL' | 'PARALLEL'

  /** 最大过滤器链深度 */
  maxFilterChainDepth?: number | null

  // 性能优化配置

  /** 是否启用路由对象池，Y启用/N禁用 */
  enableRoutePooling: 'Y' | 'N'

  /** 路由对象池大小 */
  routePoolSize?: number | null

  /** 是否启用异步处理，Y启用/N禁用 */
  enableAsyncProcessing: 'Y' | 'N'

  // 错误处理配置

  /** 是否启用降级处理，Y启用/N禁用 */
  enableFallback: 'Y' | 'N'

  /** 降级路由路径 */
  fallbackRoute?: string

  /** 路由未找到时的状态码 */
  notFoundStatusCode: number

  /** 路由未找到时的提示消息 */
  notFoundMessage?: string

  // 自定义配置

  /** Router元数据，JSON格式 */
  routerMetadata?: string

  /** 自定义配置，JSON格式 */
  customConfig?: string

  // 预留字段

  /** 预留字段1 */
  reserved1?: string

  /** 预留字段2 */
  reserved2?: string

  /** 预留字段3 */
  reserved3?: number | null

  /** 预留字段4 */
  reserved4?: number | null

  /** 预留字段5 */
  reserved5?: string | null

  // 扩展属性

  /** 扩展属性，JSON格式 */
  extProperty?: string

  // 标准字段

  /** 创建时间 */
  addTime?: string

  /** 创建人ID */
  addWho?: string

  /** 最后修改时间 */
  editTime?: string

  /** 最后修改人ID */
  editWho?: string

  /** 操作序列标识 */
  oprSeqFlag?: string

  /** 当前版本号 */
  currentVersion?: number

  /** 活动状态标记，Y活动/启用，N非活动/禁用 */
  activeFlag: 'Y' | 'N'

  /** 备注信息 */
  noteText?: string
}

