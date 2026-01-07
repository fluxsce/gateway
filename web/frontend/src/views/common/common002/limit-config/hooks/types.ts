/**
 * 限流配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * 限流配置类型
 * 严格按照 HUB_GW_RATE_LIMIT_CONFIG 表结构定义
 */
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
  customConfig: string | Record<string, any> // 自定义配置,JSON格式

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
 * 限流配置表单模态框 Props
 */
export interface RateLimitConfigFormModalProps {
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
  /** 网关实例ID（实例级限流） */
  gatewayInstanceId?: string
  /** 路由配置ID（路由级限流） */
  routeConfigId?: string
}

/**
 * 限流配置表单模态框 Emits
 */
export interface RateLimitConfigFormModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

