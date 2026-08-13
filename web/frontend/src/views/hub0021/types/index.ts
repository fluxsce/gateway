// 路由管理相关类型定义

// 导入断言相关类型（已移动到 components/assert-config/hooks/types.ts）
import type {
  AssertionGroupTestResult,
  AssertionTestResult,
  RouteAssertion,
  RouteAssertionForm,
  RouteAssertionGroup,
  RouteAssertionGroupForm,
} from '../components/assert-config/hooks/types'
import {
  AssertionOperator,
  AssertionType,
} from '../components/assert-config/hooks/types'

// 为了向后兼容，重新导出断言相关类型
export {
  AssertionOperator, AssertionType, type AssertionGroupTestResult, type AssertionTestResult, type RouteAssertion,
  type RouteAssertionForm,
  type RouteAssertionGroup,
  type RouteAssertionGroupForm
}

// 匹配类型枚举
export enum MatchType {
  EXACT = 0, // 精确匹配
  PREFIX = 1, // 前缀匹配
  REGEX = 2, // 正则匹配
}

// 过滤器执行时机枚举 (根据数据库设计调整)
export enum FilterAction {
  PRE_ROUTING = 'pre-routing', // 路由前处理
  POST_ROUTING = 'post-routing', // 路由后处理
  PRE_RESPONSE = 'pre-response', // 响应前处理
}

// 过滤器类型枚举 (根据数据库设计和filter.go逻辑)
export enum FilterType {
  HEADER = 'header', // 请求头过滤器
  QUERY_PARAM = 'query-param', // 查询参数过滤器
  BODY = 'body', // 请求体过滤器
  URL = 'url', // URL过滤器
  METHOD = 'method', // HTTP方法过滤器
  COOKIE = 'cookie', // Cookie过滤器
  RESPONSE = 'response', // 响应过滤器
}

// 过滤器执行模式枚举
export enum FilterExecutionMode {
  SEQUENTIAL = 'SEQUENTIAL', // 顺序执行
  PARALLEL = 'PARALLEL', // 并行执行
}

// 路由配置接口 (根据数据库设计)
export interface RouteConfig {
  tenantId: string // 租户ID
  routeConfigId: string // 路由配置ID
  gatewayInstanceId: string // 关联的网关实例ID
  routeName: string // 路由名称
  routePath: string // 路由路径
  allowedMethods?: string[] | string // 允许的HTTP方法数组或JSON字符串
  allowedHosts?: string // 允许的域名(逗号分隔)
  matchType: MatchType // 匹配类型(0精确,1前缀,2正则)
  routePriority: number // 路由优先级(数值越小优先级越高)
  stripPathPrefix: 'Y' | 'N' // 是否剥离路径前缀
  rewritePath?: string // 重写路径
  enableWebsocket: 'Y' | 'N' // 是否支持WebSocket
  timeoutMs: number // 超时时间(毫秒)
  retryCount: number // 重试次数
  retryIntervalMs: number // 重试间隔(毫秒)

  // 关联配置
  serviceDefinitionId?: string // 关联的服务定义ID（单服务）或逗号分割的服务ID（多服务）
  logConfigId?: string // 关联的日志配置ID

  // 元数据和扩展
  routeMetadata?: Record<string, any> // 路由元数据
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 路由配置表单
export interface RouteConfigForm {
  routeName: string
  routePath: string
  allowedMethods: string[]
  allowedHosts: string
  matchType: MatchType
  routePriority: number
  stripPathPrefix: 'Y' | 'N'
  rewritePath: string
  enableWebsocket: 'Y' | 'N'
  timeoutMs: number
  retryCount: number
  retryIntervalMs: number
  serviceDefinitionId: string
  logConfigId: string
  routeMetadata: Record<string, any>
  activeFlag: 'Y' | 'N'
  noteText: string
}

// 路由断言相关类型定义已移动到 components/assert-config/hooks/types.ts
// 上面已通过重新导出保持向后兼容

// 过滤器配置接口
export interface FilterConfig {
  tenantId: string
  filterConfigId: string
  gatewayInstanceId?: string
  routeConfigId?: string
  filterName: string
  filterType: FilterType
  filterAction: FilterAction
  filterOrder: number
  filterConfig: Record<string, any>
  filterDesc?: string
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 过滤器配置表单
export interface FilterConfigForm {
  filterName: string
  filterType: FilterType
  filterAction: FilterAction
  filterOrder: number
  filterConfig: Record<string, any>
  filterDesc: string
  activeFlag: 'Y' | 'N'
  noteText: string
}

// Router配置接口 (根据数据库设计和网关文档)
export interface RouterConfig {
  tenantId: string // 租户ID
  routerConfigId: string // Router配置ID
  gatewayInstanceId: string // 关联的网关实例ID
  routerName: string // Router名称
  routerDesc?: string // Router描述

  // 基础配置
  defaultPriority: number // 默认路由优先级
  enableRouteCache: 'Y' | 'N' // 是否启用路由缓存
  routeCacheTtlSeconds: number // 路由缓存TTL(秒)
  maxRoutes?: number // 最大路由数量限制
  routeMatchTimeout?: number // 路由匹配超时时间(毫秒)

  // 高级配置
  enableStrictMode: 'Y' | 'N' // 是否启用严格模式
  enableMetrics: 'Y' | 'N' // 是否启用指标收集
  enableTracing: 'Y' | 'N' // 是否启用链路追踪
  caseSensitive: 'Y' | 'N' // 路径匹配是否区分大小写
  removeTrailingSlash: 'Y' | 'N' // 是否移除路径尾部斜杠

  // 过滤器配置
  enableGlobalFilters: 'Y' | 'N' // 是否启用全局过滤器
  filterExecutionMode: FilterExecutionMode // 过滤器执行模式
  maxFilterChainDepth?: number // 最大过滤器链深度

  // 性能优化配置
  enableRoutePooling: 'Y' | 'N' // 是否启用路由对象池
  routePoolSize?: number // 路由对象池大小
  enableAsyncProcessing: 'Y' | 'N' // 是否启用异步处理

  // 错误处理配置
  enableFallback: 'Y' | 'N' // 是否启用降级处理
  fallbackRoute?: string // 降级路由路径
  notFoundStatusCode: number // 路由未找到时的状态码
  notFoundMessage: string // 路由未找到时的提示消息

  // 自定义配置
  routerMetadata?: Record<string, any> // Router元数据
  customConfig: Record<string, any> // 自定义配置
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// Router配置表单
export interface RouterConfigForm {
  routerName: string
  routerDesc: string
  defaultPriority: number
  enableRouteCache: 'Y' | 'N'
  routeCacheTtlSeconds: number
  maxRoutes: number
  routeMatchTimeout: number
  enableStrictMode: 'Y' | 'N'
  enableMetrics: 'Y' | 'N'
  enableTracing: 'Y' | 'N'
  caseSensitive: 'Y' | 'N'
  removeTrailingSlash: 'Y' | 'N'
  enableGlobalFilters: 'Y' | 'N'
  filterExecutionMode: FilterExecutionMode
  maxFilterChainDepth: number
  enableRoutePooling: 'Y' | 'N'
  routePoolSize: number
  enableAsyncProcessing: 'Y' | 'N'
  enableFallback: 'Y' | 'N'
  fallbackRoute: string
  notFoundStatusCode: number
  notFoundMessage: string
  routerMetadata: Record<string, any>
  customConfig: Record<string, any>
  activeFlag: 'Y' | 'N'
  noteText: string
}

// 服务定义接口（用于关联）
export interface ServiceDefinition {
  tenantId: string
  serviceDefinitionId: string
  serviceName: string
  serviceDesc?: string
  serviceType: number
  discoveryType?: string
  discoveryConfig?: Record<string, any>
  loadBalanceAlgorithm: string
  healthCheckEnabled: 'Y' | 'N'
  healthCheckPath: string
  healthCheckIntervalSeconds: number
  healthCheckTimeoutMs: number
  healthyThreshold: number
  unhealthyThreshold: number
  serviceMetadata?: Record<string, any>
  activeFlag: 'Y' | 'N'
}

// 日志配置接口（用于关联）
export interface LogConfig {
  tenantId: string
  logConfigId: string
  configName: string
  logLevel: string
  logFormat: string
  outputTargets: string
  enableAccessLog: 'Y' | 'N'
  enableErrorLog: 'Y' | 'N'
  enableAuditLog: 'Y' | 'N'
  activeFlag: 'Y' | 'N'
}

// 查询参数接口
export interface RouteQueryParams {
  tenantId?: string
  gatewayInstanceId?: string
  routeName?: string
  routePath?: string
  matchType?: MatchType
  activeFlag?: 'Y' | 'N'
  pageIndex?: number
  pageSize?: number
}

export interface RouterQueryParams {
  tenantId?: string
  gatewayInstanceId?: string
  routerName?: string
  activeFlag?: 'Y' | 'N'
  pageIndex?: number
  pageSize?: number
}

export interface FilterQueryParams {
  tenantId?: string
  gatewayInstanceId?: string
  routeConfigId?: string
  filterType?: FilterType
  filterAction?: FilterAction
  activeFlag?: 'Y' | 'N'
  pageIndex?: number
  pageSize?: number
}

// 统计信息接口
export interface RouteStatistics {
  totalRoutes: number
  activeRoutes: number
  inactiveRoutes: number
  exactMatchRoutes: number
  prefixMatchRoutes: number
  regexMatchRoutes: number
}

// 选择选项接口
export interface SelectOption {
  label: string
  value: string | number
  disabled?: boolean
}

// 路由测试接口
export interface RouteTestRequest {
  method: string
  path: string
  headers?: Record<string, string>
  queryParams?: Record<string, string>
  host?: string
}

export interface RouteTestResult {
  matched: boolean
  routeConfigId?: string
  routeName?: string
  matchType?: MatchType
  priority?: number
  targetService?: string
  executionTime: number
  assertionResult?: AssertionGroupTestResult
  conflictRoutes?: {
    routeConfigId: string
    routeName: string
    routePath: string
    matchType: MatchType
    priority: number
  }[]
}

// 网关实例接口
export interface GatewayInstance {
  tenantId: string
  gatewayInstanceId: string
  instanceName: string
  instanceDesc?: string
  bindAddress: string
  httpPort?: number
  httpsPort?: number
  tlsEnabled: 'Y' | 'N'
  certStorageType: 'FILE' | 'DATABASE'
  certFilePath?: string
  keyFilePath?: string
  certContent?: string
  keyContent?: string
  certChainContent?: string
  certPassword?: string
  maxConnections: number
  readTimeoutMs: number
  writeTimeoutMs: number
  idleTimeoutMs: number
  maxHeaderBytes: number
  maxWorkers: number
  keepAliveEnabled: 'Y' | 'N'
  tcpKeepAliveEnabled: 'Y' | 'N'
  gracefulShutdownTimeoutMs: number
  enableHttp2: 'Y' | 'N'
  tlsVersion?: string
  tlsCipherSuites?: string
  disableGeneralOptionsHandler: 'Y' | 'N'
  logConfigId?: string
  healthStatus: 'Y' | 'N'
  lastHeartbeatTime?: string
  instanceMetadata?: string
  activeFlag: 'Y' | 'N'
  noteText?: string
  addTime: string
  addWho: string
  editTime: string
  editWho: string
}

// 网关实例选项
export interface GatewayInstanceOption {
  label: string
  value: string
  disabled?: boolean
  healthStatus?: 'Y' | 'N'
}

// 网关实例查询参数
export interface GatewayInstanceQueryParams {
  tenantId?: string
  instanceName?: string
  healthStatus?: 'Y' | 'N'
  activeFlag?: 'Y' | 'N'
  pageIndex?: number
  pageSize?: number
}
