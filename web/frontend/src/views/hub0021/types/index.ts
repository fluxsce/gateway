// 路由管理相关类型定义（模块级兼容导出；组件内优先用各自 hooks/types）

export {
  AssertionOperator,
  AssertionType,
  type RouteAssertion,
} from '../components/assert-config/hooks/types'

import type {
  AssertionOperator,
  AssertionType,
} from '../components/assert-config/hooks/types'

/** 断言表单数据（遗留 hooks 使用） */
export interface RouteAssertionForm {
  assertionName: string // 断言名称
  assertionType: AssertionType | string // 断言类型(PATH,HEADER,QUERY,COOKIE,IP,BODY_CONTENT)
  assertionOperator: AssertionOperator | string // 断言操作符(EQUAL,NOT_EQUAL,CONTAINS,MATCHES等)
  fieldName?: string // 字段名称(HEADER/QUERY/COOKIE类型时使用)
  expectedValue?: string // 期望值(EQUAL/NOT_EQUAL等操作符时使用)
  patternValue?: string // 匹配模式(MATCHES/NOT_MATCHES操作符时使用,支持正则表达式)
  caseSensitive?: 'Y' | 'N' // 是否区分大小写(N否,Y是)
  assertionOrder?: number // 断言执行顺序(数值越小越先执行)
  isRequired?: 'Y' | 'N' // 是否必须匹配(N否,Y是)
  assertionDesc?: string // 断言描述
  activeFlag?: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
  routeAssertionId?: string // 路由断言ID
}

/** 匹配类型枚举 */
export enum MatchType {
  EXACT = 0, // 精确匹配
  PREFIX = 1, // 前缀匹配
  REGEX = 2, // 正则匹配
}

/** 过滤器执行时机(根据数据库设计调整) */
export enum FilterAction {
  PRE_ROUTING = 'pre-routing', // 路由前处理
  POST_ROUTING = 'post-routing', // 路由后处理
  PRE_RESPONSE = 'pre-response', // 响应前处理
}

/** 过滤器类型(根据数据库设计和filter.go逻辑) */
export enum FilterType {
  HEADER = 'header', // 请求头过滤器
  QUERY_PARAM = 'query-param', // 查询参数过滤器
  BODY = 'body', // 请求体过滤器
  URL = 'url', // URL过滤器
  METHOD = 'method', // HTTP方法过滤器
  COOKIE = 'cookie', // Cookie过滤器
  RESPONSE = 'response', // 响应过滤器
}

/** 过滤器执行模式 */
export enum FilterExecutionMode {
  SEQUENTIAL = 'SEQUENTIAL', // 顺序执行
  PARALLEL = 'PARALLEL', // 并行执行
}

/** 路由配置接口(根据数据库设计) */
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
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5
  extProperty?: Record<string, any> // 扩展属性
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(Y活动/启用,N非活动/禁用)
  noteText?: string // 备注信息
}

/** 过滤器配置接口 */
export interface FilterConfig {
  tenantId: string // 租户ID
  filterConfigId: string // 过滤器配置ID
  gatewayInstanceId?: string // 关联的网关实例ID
  routeConfigId?: string // 关联的路由配置ID
  filterName: string // 过滤器名称
  filterType: FilterType // 过滤器类型
  filterAction: FilterAction // 过滤器执行时机
  filterOrder: number // 过滤器执行顺序
  filterConfig: Record<string, any> // 过滤器配置内容
  filterDesc?: string // 过滤器描述
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5
  extProperty?: Record<string, any> // 扩展属性
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(Y活动/启用,N非活动/禁用)
  noteText?: string // 备注信息
}

/** 过滤器配置表单 */
export interface FilterConfigForm {
  filterName: string // 过滤器名称
  filterType: FilterType // 过滤器类型
  filterAction: FilterAction // 过滤器执行时机
  filterOrder: number // 过滤器执行顺序
  filterConfig: Record<string, any> // 过滤器配置内容
  filterDesc: string // 过滤器描述
  activeFlag: 'Y' | 'N' // 活动状态标记
  noteText: string // 备注信息
}

/** Router配置接口(根据数据库设计和网关文档) */
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
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5
  extProperty?: Record<string, any> // 扩展属性
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(Y活动/启用,N非活动/禁用)
  noteText?: string // 备注信息
}

/** Router配置表单 */
export interface RouterConfigForm {
  routerName: string // Router名称
  routerDesc: string // Router描述
  defaultPriority: number // 默认路由优先级
  enableRouteCache: 'Y' | 'N' // 是否启用路由缓存
  routeCacheTtlSeconds: number // 路由缓存TTL(秒)
  maxRoutes: number // 最大路由数量限制
  routeMatchTimeout: number // 路由匹配超时时间(毫秒)
  enableStrictMode: 'Y' | 'N' // 是否启用严格模式
  enableMetrics: 'Y' | 'N' // 是否启用指标收集
  enableTracing: 'Y' | 'N' // 是否启用链路追踪
  caseSensitive: 'Y' | 'N' // 路径匹配是否区分大小写
  removeTrailingSlash: 'Y' | 'N' // 是否移除路径尾部斜杠
  enableGlobalFilters: 'Y' | 'N' // 是否启用全局过滤器
  filterExecutionMode: FilterExecutionMode // 过滤器执行模式
  maxFilterChainDepth: number // 最大过滤器链深度
  enableRoutePooling: 'Y' | 'N' // 是否启用路由对象池
  routePoolSize: number // 路由对象池大小
  enableAsyncProcessing: 'Y' | 'N' // 是否启用异步处理
  enableFallback: 'Y' | 'N' // 是否启用降级处理
  fallbackRoute: string // 降级路由路径
  notFoundStatusCode: number // 路由未找到时的状态码
  notFoundMessage: string // 路由未找到时的提示消息
  routerMetadata: Record<string, any> // Router元数据
  customConfig: Record<string, any> // 自定义配置
  activeFlag: 'Y' | 'N' // 活动状态标记
  noteText: string // 备注信息
}

/** 服务定义接口（遗留 modules 选择器使用） */
export interface ServiceDefinition {
  tenantId: string // 租户ID
  serviceDefinitionId: string // 服务定义ID
  serviceName: string // 服务名称
  serviceDesc?: string // 服务描述
  serviceType: number // 服务类型
  discoveryType?: string // 服务发现类型
  discoveryConfig?: Record<string, any> // 服务发现配置
  loadBalanceAlgorithm: string // 负载均衡算法
  healthCheckEnabled: 'Y' | 'N' // 是否启用健康检查
  healthCheckPath: string // 健康检查路径
  healthCheckIntervalSeconds: number // 健康检查间隔(秒)
  healthCheckTimeoutMs: number // 健康检查超时(毫秒)
  healthyThreshold: number // 健康阈值
  unhealthyThreshold: number // 不健康阈值
  serviceMetadata?: Record<string, any> // 服务元数据
  activeFlag: 'Y' | 'N' // 活动状态标记
}

/** 路由查询参数 */
export interface RouteQueryParams {
  tenantId?: string // 租户ID
  gatewayInstanceId?: string // 网关实例ID
  routeName?: string // 路由名称
  routePath?: string // 路由路径
  matchType?: MatchType // 匹配类型
  activeFlag?: 'Y' | 'N' // 活动状态标记
  pageIndex?: number // 页码
  pageSize?: number // 每页条数
}

/** Router查询参数 */
export interface RouterQueryParams {
  tenantId?: string // 租户ID
  gatewayInstanceId?: string // 网关实例ID
  routerName?: string // Router名称
  activeFlag?: 'Y' | 'N' // 活动状态标记
  pageIndex?: number // 页码
  pageSize?: number // 每页条数
}

/** 路由统计 */
export interface RouteStatistics {
  totalRoutes: number // 路由总数
  activeRoutes: number // 启用路由数
  inactiveRoutes: number // 禁用路由数
  exactMatchRoutes: number // 精确匹配路由数
  prefixMatchRoutes: number // 前缀匹配路由数
  regexMatchRoutes: number // 正则匹配路由数
}
