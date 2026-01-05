/**
 * IP访问控制配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * IP访问控制配置类型
 * 严格按照 HUB_GATEWAY_IP_ACCESS_CONFIG 表结构定义
 */
export interface IpAccessConfig {
  tenantId: string // 租户ID
  ipAccessConfigId: string // IP访问配置ID
  securityConfigId: string // 关联的安全配置ID/实例id或者路由id
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

// ============= 组件 Props 和 Emits 类型 =============

/**
 * IP访问控制配置列表模态框 Props
 */
export interface IpAccessConfigListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 安全配置ID（新增时使用） */
  securityConfigId?: string
}

/**
 * IP访问控制配置列表模态框 Emits
 */
export interface IpAccessConfigListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

