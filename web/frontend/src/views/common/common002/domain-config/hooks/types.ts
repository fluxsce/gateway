/**
 * 域名访问控制配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * 域名访问控制配置类型
 * 严格按照 HUB_GATEWAY_DOMAIN_ACCESS_CONFIG 表结构定义
 */
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

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 域名访问控制配置列表模态框 Props
 */
export interface DomainAccessConfigListModalProps {
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
 * 域名访问控制配置列表模态框 Emits
 */
export interface DomainAccessConfigListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

