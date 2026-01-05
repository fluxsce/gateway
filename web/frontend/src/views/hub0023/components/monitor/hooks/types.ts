/**
 * Hub0023 监控模块类型定义
 * 内聚在监控组件内部
 */

/**
 * 时间粒度枚举
 */
export enum TimeGranularity {
  /** 分钟 */
  MINUTE = 'MINUTE',
  /** 小时 */
  HOUR = 'HOUR',
  /** 天 */
  DAY = 'DAY',
}

/**
 * 网关监控概览数据接口
 */
export interface GatewayMonitoringOverview {
  /** 总请求数 */
  totalRequests: number
  /** 成功请求数 */
  successRequests: number
  /** 失败请求数 */
  failedRequests: number
  /** 每秒请求数(QPS) */
  requestsPerSecond: number
  /** 平均响应时间(毫秒) */
  avgResponseTimeMs: number
  /** 最小响应时间(毫秒) */
  minResponseTimeMs: number
  /** 最大响应时间(毫秒) */
  maxResponseTimeMs: number
}

/**
 * 响应时间详细指标接口
 */
export interface ResponseTimeMetrics {
  /** 时间戳 */
  timestamp: number
  /** 平均响应时间(毫秒) */
  avgResponseTimeMs: number
  /** 最小响应时间(毫秒) */
  minResponseTimeMs: number
  /** 最大响应时间(毫秒) */
  maxResponseTimeMs: number
  /** 50%响应时间(毫秒) */
  p50ResponseTimeMs: number
  /** 90%响应时间(毫秒) */
  p90ResponseTimeMs: number
  /** 99%响应时间(毫秒) */
  p99ResponseTimeMs: number
  /** 该时间点的总请求数 */
  requestCount: number
}

/**
 * 请求指标数据接口
 */
export interface RequestMetrics {
  /** 时间戳 */
  timestamp: number
  /** 总请求数 */
  totalRequests: number
  /** 成功请求数 */
  successRequests: number
  /** 失败请求数 */
  failedRequests: number
  /** 每秒请求数(QPS) */
  requestsPerSecond: number
}

/**
 * 网关监控状态码分布数据接口
 */
export interface GatewayMonitoringStatusCodeData {
  /** 状态码 */
  statusCode: string
  /** 数量 */
  count: number
  /** 百分比 */
  percentage: number
  /** 状态码分类（2xx成功、4xx客户端错误、5xx服务端错误等） */
  category?: string
  /** 状态码描述 */
  description?: string
}

/**
 * 网关监控热点路由数据接口
 */
export interface GatewayMonitoringHotRouteData {
  /** 路由路径 */
  routePath: string
  /** 请求数量 */
  requestCount: number
  /** 平均响应时间(毫秒) */
  avgResponseTimeMs: number
  /** 错误率(%) */
  errorRate: number
  /** QPS */
  qps: number
  /** 路由配置ID */
  routeConfigId?: string
  /** 路由名称 */
  routeName?: string
  /** 服务名称 */
  serviceName?: string
  /** 最大响应时间(毫秒) */
  maxResponseTimeMs?: number
  /** 最小响应时间(毫秒) */
  minResponseTimeMs?: number
}

/**
 * 网关监控图表数据接口
 */
export interface GatewayMonitoringChartData {
  /** 请求量趋势(按分钟) */
  requestTrend: RequestMetrics[]
  /** 响应时间趋势(按分钟) */
  responseTimeTrend: ResponseTimeMetrics[]
  /** 状态码分布 */
  statusCodeDistribution: GatewayMonitoringStatusCodeData[]
  /** 热点路由TOP10 */
  hotRoutes: GatewayMonitoringHotRouteData[]
}

/**
 * 网关监控数据查询参数接口
 */
export interface GatewayMonitoringQueryParams {
  /** 网关实例ID */
  gatewayInstanceId?: string
  /** 开始时间 */
  startTime: string
  /** 结束时间 */
  endTime: string
  /** 时间粒度(MINUTE,HOUR,DAY) */
  timeGranularity: TimeGranularity

  // 基础筛选参数
  /** 路由配置ID */
  routeConfigId?: string
  /** 路由名称(利用冗余字段查询) */
  routeName?: string
  /** 服务定义ID */
  serviceDefinitionId?: string
  /** 服务名称(利用冗余字段查询) */
  serviceName?: string

  // 请求筛选参数
  /** 请求路径(支持模糊匹配) */
  requestPath?: string
}

