/**
 * 路由配置列表组件类型定义
 */

/**
 * 匹配类型枚举
 * 0: 精确匹配 - 路径必须完全匹配
 * 1: 前缀匹配 - 路径以指定前缀开头即可匹配
 * 2: 正则匹配 - 使用正则表达式匹配路径
 */
export enum MatchType {
  EXACT = 0, // 精确匹配
  PREFIX = 1, // 前缀匹配
  REGEX = 2, // 正则匹配
}

/**
 * 路由配置接口
 * 对应数据库表：HUB_GW_ROUTE_CONFIG
 * 用于定义路由的完整配置信息
 */
export interface RouteConfig {
  /** 租户ID */
  tenantId: string
  /** 路由配置ID */
  routeConfigId: string
  /** 关联的网关实例ID */
  gatewayInstanceId: string
  /** 路由名称 */
  routeName: string
  /** 路由路径 */
  routePath: string
  /** 允许的HTTP方法数组或JSON字符串 */
  allowedMethods?: string[] | string
  /** 允许的域名(逗号分隔) */
  allowedHosts?: string
  /** 匹配类型(0精确,1前缀,2正则) */
  matchType: MatchType
  /** 路由优先级(数值越小优先级越高) */
  routePriority: number
  /** 是否剥离已匹配路由路径前缀后再转发 */
  stripPathPrefix: 'Y' | 'N'
  /** 重写路径；非空时整段替换，空则使用原有路径拼接 */
  rewritePath?: string
  /**
   * WebSocket 路由标记。
   * N 仍允许 Upgrade（兼容历史默认）；Y 表示明确标识 WebSocket 用途。
   */
  enableWebsocket: 'Y' | 'N'
  /**
   * 路由请求总超时（毫秒）。
   * 仅当 routeMetadata.overrideProxyTimeout=Y 且值大于0时覆盖代理总超时。
   */
  timeoutMs: number
  /** 路由级重试次数；须开启覆盖且与 retryIntervalMs 同时大于0才覆盖代理 */
  retryCount: number
  /** 路由级重试间隔（毫秒）；须开启覆盖且与 retryCount 同时大于0才覆盖代理 */
  retryIntervalMs: number

  // 关联配置
  /** 关联的服务定义ID */
  serviceDefinitionId?: string
  /** 关联的日志配置ID */
  logConfigId?: string
  /** 关联的服务名称（用于显示） */
  serviceName?: string

  // 元数据和扩展
  /** 路由元数据 */
  routeMetadata?: Record<string, any>
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 标准字段
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
  /** 活动状态标记，Y活动/启用，N非活动/禁用 */
  activeFlag: 'Y' | 'N'
  /** 备注信息 */
  noteText?: string
}
