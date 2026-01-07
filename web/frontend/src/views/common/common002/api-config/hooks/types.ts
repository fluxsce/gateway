/**
 * API访问控制配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * API访问控制配置类型
 * 严格按照 HUB_GATEWAY_API_ACCESS_CONFIG 表结构定义
 */
export interface ApiAccessConfig {
  tenantId: string // 租户ID
  apiAccessConfigId: string // API访问配置ID
  securityConfigId: string // 关联的安全配置ID/实例id或者路由id
  configName: string // API访问配置名称
  defaultPolicy: 'allow' | 'deny' // 默认策略(allow允许,deny拒绝)
  whitelistPaths?: string[] // API路径白名单,JSON数组格式,支持通配符
  blacklistPaths?: string[] // API路径黑名单,JSON数组格式,支持通配符
  allowedMethods?: string[] // 允许的HTTP方法,数组格式
  blockedMethods?: string[] // 禁止的HTTP方法,数组格式

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
 * API访问控制配置列表模态框 Props
 */
export interface ApiAccessConfigListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模块ID（用于权限控制，必填） */
  moduleId: string
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
 * API访问控制配置列表模态框 Emits
 */
export interface ApiAccessConfigListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

