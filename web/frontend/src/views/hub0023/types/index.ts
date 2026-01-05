/**
 * Hub0023 网关日志管理模块 - 类型定义
 * 基于 HUB_GW_ACCESS_LOG 表结构定义
 */

/**
 * 网关访问日志完整信息接口
 * 严格按照 HUB_GW_ACCESS_LOG 表结构定义
 */
export interface GatewayLogInfo {
  /** 租户ID */
  tenantId: string
  /** 链路追踪ID(作为主键) */
  traceId: string
  /** 网关实例ID */
  gatewayInstanceId: string
  /** 网关实例名称(冗余字段,便于查询显示) */
  gatewayInstanceName?: string
  /** 网关节点IP地址 */
  gatewayNodeIp: string
  /** 路由配置ID */
  routeConfigId?: string
  /** 路由名称(冗余字段,便于查询显示) */
  routeName?: string
  /** 服务定义ID */
  serviceDefinitionId?: string
  /** 服务名称(冗余字段,便于查询显示) */
  serviceName?: string
  /** 代理类型(http,websocket,tcp,udp,可为空) */
  proxyType?: string
  /** 日志配置ID */
  logConfigId?: string

  // 请求基本信息
  /** 请求方法(GET,POST,PUT等) */
  requestMethod: string
  /** 请求路径 */
  requestPath: string
  /** 请求查询参数 */
  requestQuery?: string
  /** 请求大小(字节) */
  requestSize?: number
  /** 请求头信息,JSON格式 */
  requestHeaders?: string
  /** 请求体(可选,根据配置决定是否记录) */
  requestBody?: string

  // 客户端信息
  /** 客户端IP地址 */
  clientIpAddress: string
  /** 客户端端口 */
  clientPort?: number
  /** 用户代理信息 */
  userAgent?: string
  /** 来源页面 */
  referer?: string
  /** 用户标识(如有) */
  userIdentifier?: string

  // 关键时间点 (所有时间字段均为DATETIME类型，精确到毫秒)
  /** 网关开始处理时间(请求开始处理，必填) */
  gatewayStartProcessingTime: string
  /** 后端服务请求开始时间(可选) */
  backendRequestStartTime?: string
  /** 后端服务响应接收时间(可选) */
  backendResponseReceivedTime?: string
  /** 网关处理完成时间(可选，正在处理中或异常中断时为空) */
  gatewayFinishedProcessingTime?: string

  // 计算的时间指标 (所有时间指标均为毫秒)
  /** 总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL) */
  totalProcessingTimeMs?: number
  /** 网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL) */
  gatewayProcessingTimeMs?: number
  /** 后端服务响应时间(毫秒，可选) */
  backendResponseTimeMs?: number

  // 响应信息
  /** 网关响应状态码 */
  gatewayStatusCode: number
  /** 后端服务状态码 */
  backendStatusCode?: number
  /** 响应大小(字节) */
  responseSize?: number
  /** 响应头信息,JSON格式 */
  responseHeaders?: string
  /** 响应体(可选,根据配置决定是否记录) */
  responseBody?: string

  // 转发基本信息
  /** 匹配的路由路径 */
  matchedRoute?: string
  /** 转发地址 */
  forwardAddress?: string
  /** 转发方法 */
  forwardMethod?: string
  /** 转发参数,JSON格式 */
  forwardParams?: string
  /** 转发头信息,JSON格式 */
  forwardHeaders?: string
  /** 转发报文内容 */
  forwardBody?: string
  /** 负载均衡决策信息 */
  loadBalancerDecision?: string

  // 错误信息
  /** 错误信息(如有) */
  errorMessage?: string
  /** 错误代码(如有) */
  errorCode?: string

  // 追踪信息
  /** 父链路追踪ID */
  parentTraceId?: string

  // 日志重置标记和次数
  /** 日志重置标记(N否,Y是) */
  resetFlag: string
  /** 重试次数 */
  retryCount: number
  /** 重置次数 */
  resetCount: number

  // 标准数据库字段
  /** 日志级别 */
  logLevel: string
  /** 日志类型 */
  logType: string
  /** 预留字段1 */
  reserved1?: string
  /** 预留字段2 */
  reserved2?: string
  /** 预留字段3 */
  reserved3?: number
  /** 预留字段4 */
  reserved4?: number
  /** 预留字段5 */
  reserved5?: string
  /** 扩展属性,JSON格式 */
  extProperty?: string
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
  /** 活动状态标记(N非活动,Y活动) */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
}

/**
 * 网关访问日志列表项接口(不包含大字段以提高性能)
 * 基于 HUB_GW_ACCESS_LOG 表结构，去除大字段
 */
export interface GatewayLogListItem {
  /** 租户ID */
  tenantId: string
  /** 链路追踪ID */
  traceId: string
  /** 网关实例ID */
  gatewayInstanceId: string
  /** 网关实例名称(冗余字段,便于查询显示) */
  gatewayInstanceName?: string
  /** 网关节点IP地址 */
  gatewayNodeIp: string
  /** 路由配置ID */
  routeConfigId?: string
  /** 路由名称(冗余字段,便于查询显示) */
  routeName?: string
  /** 服务定义ID */
  serviceDefinitionId?: string
  /** 服务名称(冗余字段,便于查询显示) */
  serviceName?: string
  /** 代理类型(http,websocket,tcp,udp,可为空) */
  proxyType?: string

  // 请求基本信息(不包含大字段)
  /** 请求方法(GET,POST,PUT等) */
  requestMethod: string
  /** 请求路径 */
  requestPath: string
  /** 请求大小(字节) */
  requestSize?: number

  // 客户端信息
  /** 客户端IP地址 */
  clientIpAddress: string
  /** 客户端端口 */
  clientPort?: number
  /** 用户代理信息 */
  userAgent?: string
  /** 用户标识(如有) */
  userIdentifier?: string

  // 关键时间点
  /** 网关开始处理时间(请求开始处理，必填) */
  gatewayStartProcessingTime: string
  /** 网关处理完成时间(可选，正在处理中或异常中断时为空) */
  gatewayFinishedProcessingTime?: string

  // 计算的时间指标
  /** 总处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL) */
  totalProcessingTimeMs?: number
  /** 网关处理时间(毫秒，当gatewayFinishedProcessingTime为空时为NULL) */
  gatewayProcessingTimeMs?: number
  /** 后端服务响应时间(毫秒，可选) */
  backendResponseTimeMs?: number

  // 响应信息
  /** 网关响应状态码 */
  gatewayStatusCode: number
  /** 后端服务状态码 */
  backendStatusCode?: number
  /** 响应大小(字节) */
  responseSize?: number

  // 转发基本信息
  /** 匹配的路由路径 */
  matchedRoute?: string
  /** 转发地址 */
  forwardAddress?: string

  // 错误信息
  /** 错误信息(如有) */
  errorMessage?: string
  /** 错误代码(如有) */
  errorCode?: string

  // 日志重置标记和次数
  /** 日志重置标记(N否,Y是) */
  resetFlag: string
  /** 重试次数 */
  retryCount: number
  /** 重置次数 */
  resetCount: number

  // 标准数据库字段
  /** 日志级别 */
  logLevel: string
  /** 日志类型 */
  logType: string
  /** 创建时间 */
  addTime: string
  /** 活动状态标记(N非活动,Y活动) */
  activeFlag: string
}

/**
 * 网关访问日志查询参数接口
 * 基于 HUB_GW_ACCESS_LOG 表结构和索引设计
 */
export interface GatewayLogQueryParams {
  /** 分页参数 - 页码索引(从常量配置获取默认值) */
  pageIndex: number
  /** 每页大小(从常量配置获取默认值) */
  pageSize: number

  /** 链路追踪ID */
  traceId?: string
  /** 网关实例ID */
  gatewayInstanceId?: string
  /** 网关实例名称(利用冗余字段查询) */
  gatewayInstanceName?: string
  /** 路由配置ID */
  routeConfigId?: string
  /** 路由名称(利用冗余字段查询) */
  routeName?: string
  /** 服务定义ID */
  serviceDefinitionId?: string
  /** 服务名称(利用冗余字段查询) */
  serviceName?: string
  /** 代理类型 */
  proxyType?: string

  // 请求信息筛选
  /** 请求方法 */
  requestMethod?: string
  /** 请求路径(支持模糊匹配) */
  requestPath?: string
  /** 客户端IP地址 */
  clientIpAddress?: string
  /** 用户标识 */
  userIdentifier?: string

  // 时间范围筛选 - 基于网关开始处理时间
  /** 开始时间 */
  startTime?: string
  /** 结束时间 */
  endTime?: string

  // 状态筛选
  /** 网关响应状态码 */
  gatewayStatusCode?: number
  /** 后端服务状态码 */
  backendStatusCode?: number
  /** 日志级别 */
  logLevel?: string
  /** 日志类型 */
  logType?: string
  /** 重置标记 */
  resetFlag?: string

  // 性能筛选
  /** 最小处理时间(毫秒) */
  minProcessingTime?: number
  /** 最大处理时间(毫秒) */
  maxProcessingTime?: number

  // 错误筛选
  /** 是否只查询错误日志 */
  errorOnly?: boolean
  /** 错误代码 */
  errorCode?: string

  // 排序参数
  /** 排序字段 */
  sortField?: string
  /** 排序方向 */
  sortOrder?: 'ASC' | 'DESC'
}

/**
 * 网关访问日志获取详情参数接口
 */
export interface GatewayLogGetParams {
  /** 链路追踪ID */
  traceId: string
}

/**
 * 网关访问日志重置参数接口
 */
export interface GatewayLogResetParams {
  /** 链路追踪ID列表(支持批量) */
  traceIds: string[]
  /** 重置原因 */
  resetReason?: string
  /** 操作人ID */
  operatorId: string
}

