// 安全配置基础类型 - 严格按照HUB_GATEWAY_SECURITY_CONFIG表结构定义
export interface SecurityConfig {
  // 基础标识字段
  tenantId: string // 租户ID
  securityConfigId: string // 安全配置ID
  gatewayInstanceId?: string // 网关实例ID(实例级安全配置)
  routeConfigId?: string // 路由配置ID(路由级安全配置)
  configName: string // 安全配置名称
  configDesc?: string // 安全配置描述
  configPriority: number // 配置优先级,数值越小优先级越高
  customConfigJson?: Record<string, any> // 自定义配置参数,JSON格式

  // 预留字段
  reserved1?: string // 预留字段1
  reserved2?: string // 预留字段2
  reserved3?: number // 预留字段3
  reserved4?: number // 预留字段4
  reserved5?: string // 预留字段5
  extProperty?: Record<string, any> // 扩展属性,JSON格式

  // 系统字段
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
}

// IP访问控制配置类型 - 严格按照HUB_GATEWAY_IP_ACCESS_CONFIG表结构定义
export interface IpAccessConfig {
  tenantId: string // 租户ID
  ipAccessConfigId: string // IP访问配置ID
  securityConfigId: string // 关联的安全配置ID
  configName: string // IP访问配置名称
  defaultPolicy: 'allow' | 'deny' // 默认策略(allow允许,deny拒绝)
  whitelistIps?: string[] // IP白名单,JSON数组格式
  blacklistIps?: string[] // IP黑名单,JSON数组格式
  whitelistCidrs?: string[] // CIDR白名单,JSON数组格式
  blacklistCidrs?: string[] // CIDR黑名单,JSON数组格式
  trustXForwardedFor: 'Y' | 'N' // 是否信任X-Forwarded-For头(N否,Y是)
  trustXRealIp: 'Y' | 'N' // 是否信任X-Real-IP头(N否,Y是)

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// User-Agent访问控制配置类型 - 严格按照HUB_GATEWAY_USERAGENT_ACCESS_CONFIG表结构定义
export interface UserAgentAccessConfig {
  tenantId: string // 租户ID
  useragentAccessConfigId: string // User-Agent访问配置ID
  securityConfigId: string // 关联的安全配置ID
  configName: string // User-Agent访问配置名称
  defaultPolicy: 'allow' | 'deny' // 默认策略(allow允许,deny拒绝)
  whitelistPatterns?: string[] // User-Agent白名单模式,JSON数组格式,支持正则表达式
  blacklistPatterns?: string[] // User-Agent黑名单模式,JSON数组格式,支持正则表达式
  blockEmptyUserAgent: 'Y' | 'N' // 是否阻止空User-Agent(N否,Y是)

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// API访问控制配置类型 - 严格按照HUB_GATEWAY_API_ACCESS_CONFIG表结构定义
export interface ApiAccessConfig {
  tenantId: string // 租户ID
  apiAccessConfigId: string // API访问配置ID
  securityConfigId: string // 关联的安全配置ID
  configName: string // API访问配置名称
  defaultPolicy: 'allow' | 'deny' // 默认策略(allow允许,deny拒绝)
  whitelistPaths?: string[] // API路径白名单,JSON数组格式,支持通配符
  blacklistPaths?: string[] // API路径黑名单,JSON数组格式,支持通配符
  allowedMethods: string // 允许的HTTP方法,逗号分隔
  blockedMethods?: string // 禁止的HTTP方法,逗号分隔

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 域名访问控制配置类型 - 严格按照HUB_GATEWAY_DOMAIN_ACCESS_CONFIG表结构定义
export interface DomainAccessConfig {
  tenantId: string // 租户ID
  domainAccessConfigId: string // 域名访问配置ID
  securityConfigId: string // 关联的安全配置ID
  configName: string // 域名访问配置名称
  defaultPolicy: 'allow' | 'deny' // 默认策略(allow允许,deny拒绝)
  whitelistDomains?: string[] // 域名白名单,JSON数组格式
  blacklistDomains?: string[] // 域名黑名单,JSON数组格式
  allowSubdomains: 'Y' | 'N' // 是否允许子域名(N否,Y是)

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 安全配置基础表单类型 - 分离式架构，只包含基础配置
export interface SecurityConfigForm {
  configName: string // 安全配置名称
  configDesc: string // 安全配置描述
  configPriority: number // 配置优先级
  gatewayInstanceId?: string // 网关实例ID(可选)
  routeConfigId?: string // 路由配置ID(可选)
  customConfigJson?: Record<string, any> // 自定义配置参数
  noteText: string // 备注信息
}

// 子模块表单类型定义 - 用于各个子模块组件
export interface IpAccessConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记，替代enabled字段
  configName: string
  defaultPolicy: 'allow' | 'deny'
  whitelistIps: string[]
  blacklistIps: string[]
  whitelistCidrs: string[]
  blacklistCidrs: string[]
  trustXForwardedFor: 'Y' | 'N'
  trustXRealIp: 'Y' | 'N'
}

export interface UserAgentAccessConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记，替代enabled字段
  configName: string
  defaultPolicy: 'allow' | 'deny'
  whitelistPatterns: string[]
  blacklistPatterns: string[]
  blockEmptyUserAgent: 'Y' | 'N'
}

export interface ApiAccessConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记，替代enabled字段
  configName: string
  defaultPolicy: 'allow' | 'deny'
  whitelistPaths: string[]
  blacklistPaths: string[]
  allowedMethods: string[]
  blockedMethods: string[]
}

export interface DomainAccessConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记，替代enabled字段
  configName: string
  defaultPolicy: 'allow' | 'deny'
  whitelistDomains: string[]
  blacklistDomains: string[]
  allowSubdomains: 'Y' | 'N'
}

// 查询参数类型
export interface SecurityConfigQueryParams {
  tenantId?: string // 可选，后端会话自动处理
  gatewayInstanceId?: string
  routeConfigId?: string
  configName?: string
  activeFlag?: 'Y' | 'N'
  pageNo?: number
  pageSize?: number
}

// 统计信息类型
export interface SecurityConfigStatistics {
  totalConfigs: number
  activeConfigs: number
  inactiveConfigs: number
  instanceLevelConfigs: number
  routeLevelConfigs: number
}

// 选择选项类型
export interface SelectOption {
  label: string
  value: string | number
}

// 配置状态类型
export interface ConfigStatus {
  total: number
  active: number
  inactive: number
}

// CORS配置类型 - 严格按照HUB_GATEWAY_CORS_CONFIG表结构定义
export interface CorsConfig {
  tenantId: string // 租户ID
  corsConfigId: string // CORS配置ID
  gatewayInstanceId?: string // 网关实例ID(实例级CORS)
  routeConfigId?: string // 路由配置ID(路由级CORS)
  configName: string // 配置名称
  allowOrigins: string[] // 允许的源,JSON数组格式
  allowMethods: string // 允许的HTTP方法
  allowHeaders?: string[] // 允许的请求头,JSON数组格式
  exposeHeaders?: string[] // 暴露的响应头,JSON数组格式
  allowCredentials: 'Y' | 'N' // 是否允许携带凭证(N否,Y是)
  maxAgeSeconds: number // 预检请求缓存时间(秒)
  configPriority: number // 配置优先级,数值越小优先级越高

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// CORS配置表单类型
export interface CorsConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记
  configName: string // 配置名称
  allowOrigins: string[] // 允许的源
  allowMethods: string[] // 允许的HTTP方法
  allowHeaders: string[] // 允许的请求头
  exposeHeaders: string[] // 暴露的响应头
  allowCredentials: 'Y' | 'N' // 是否允许携带凭证
  maxAgeSeconds: number // 预检请求缓存时间(秒)
  configPriority: number // 配置优先级
}

// 认证配置类型 - 严格按照HUB_GATEWAY_AUTH_CONFIG表结构定义
export interface AuthConfig {
  tenantId: string // 租户ID
  authConfigId: string // 认证配置ID
  gatewayInstanceId?: string // 网关实例ID(实例级认证)
  routeConfigId?: string // 路由配置ID(路由级认证)
  authName: string // 认证配置名称
  authType: 'JWT' | 'API_KEY' | 'OAUTH2' | 'BASIC' // 认证类型
  authStrategy: 'REQUIRED' | 'OPTIONAL' | 'DISABLED' // 认证策略
  authConfig: Record<string, any> // 认证参数配置,JSON格式
  exemptPaths?: string[] // 豁免路径列表,JSON数组格式
  exemptHeaders?: string[] // 豁免请求头列表,JSON数组格式
  failureStatusCode: number // 认证失败状态码
  failureMessage: string // 认证失败提示消息
  configPriority: number // 配置优先级,数值越小优先级越高

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 认证配置表单类型
export interface AuthConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记
  authName: string // 认证配置名称
  authType: 'JWT' | 'API_KEY' | 'OAUTH2' | 'BASIC' // 认证类型
  authStrategy: 'REQUIRED' | 'OPTIONAL' | 'DISABLED' // 认证策略
  authConfig: Record<string, any> // 认证参数配置，支持动态字段
  exemptPaths: string[] // 豁免路径列表
  exemptHeaders: string[] // 豁免请求头列表
  failureStatusCode: number // 认证失败状态码
  failureMessage: string // 认证失败提示消息
  configPriority: number // 配置优先级
}

// JWT认证配置
export interface JWTAuthConfig {
  secret: string // JWT密钥
  issuer?: string // 签发者
  expiration: number // 过期时间（秒）
  algorithm: string // 签名算法：HS256, HS384, HS512, RS256
  verifyExpiration: boolean // 是否验证过期时间
  verifyIssuer: boolean // 是否验证签发者
  refreshWindow: number // 强制刷新时间窗口（秒）
  includeInResponse: boolean // 是否在响应中包含token信息
  responseHeaderName: string // token在响应中的头部名称
}

// API Key认证配置
export interface APIKeyAuthConfig {
  keyLocation: 'header' | 'query' // API Key位置
  keyName: string // 参数名称
  validKeys: string[] // 有效的API Keys
}

// OAuth2认证配置
export interface OAuth2AuthConfig {
  tokenEndpoint: string // Token端点
  clientID: string // 客户端ID
  clientSecret: string // 客户端密钥
  scope: string // 授权范围
  introspectEndpoint: string // 内省端点
}

// Basic认证配置
export interface BasicAuthConfig {
  username: string // 用户名
  password: string // 密码
}

// 限流配置类型 - 严格按照HUB_GATEWAY_RATE_LIMIT_CONFIG表结构定义
export interface RateLimitConfig {
  tenantId: string // 租户ID
  rateLimitConfigId: string // 限流配置ID
  gatewayInstanceId?: string // 网关实例ID(实例级限流)
  routeConfigId?: string // 路由配置ID(路由级限流)
  limitName: string // 限流规则名称
  algorithm: 'token-bucket' | 'leaky-bucket' | 'sliding-window' | 'fixed-window' | 'none' // 限流算法
  keyStrategy: 'ip' | 'user' | 'path' | 'service' | 'route' // 限流键策略
  limitRate: number // 限流速率(次/秒)
  burstCapacity: number // 突发容量
  timeWindowSeconds: number // 时间窗口(秒)
  rejectionStatusCode: number // 拒绝时的HTTP状态码
  rejectionMessage: string // 拒绝时的提示消息
  configPriority: number // 配置优先级,数值越小优先级越高
  customConfig: Record<string, any> // 自定义配置,JSON格式

  // 预留字段
  reserved1?: string
  reserved2?: string
  reserved3?: number
  reserved4?: number
  reserved5?: string // 预留字段5 (DATETIME类型在API中以字符串形式传输)
  extProperty?: Record<string, any>

  // 系统字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: 'Y' | 'N'
  noteText?: string
}

// 限流配置表单类型
export interface RateLimitConfigForm {
  activeFlag: 'Y' | 'N' // 活动状态标记
  limitName: string // 限流规则名称
  algorithm: 'token-bucket' | 'leaky-bucket' | 'sliding-window' | 'fixed-window' | 'none' // 限流算法
  keyStrategy: 'ip' | 'user' | 'path' | 'service' | 'route' // 限流键策略
  limitRate: number // 限流速率(次/秒)
  burstCapacity: number // 突发容量
  timeWindowSeconds: number // 时间窗口(秒)
  rejectionStatusCode: number // 拒绝时的HTTP状态码
  rejectionMessage: string // 拒绝时的提示消息
  configPriority: number // 配置优先级
  customConfig: Record<string, any> // 自定义配置,JSON格式
}
