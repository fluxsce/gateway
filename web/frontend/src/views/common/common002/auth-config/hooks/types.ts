/**
 * 认证配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * 认证配置类型
 * 严格按照 HUB_GATEWAY_AUTH_CONFIG 表结构定义
 */
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

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 认证配置表单模态框 Props
 */
export interface AuthConfigFormModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 模块ID（用于挂载） */
  moduleId?: string
  /** 网关实例ID（实例级认证） */
  gatewayInstanceId?: string
  /** 路由配置ID（路由级认证） */
  routeConfigId?: string
}

/**
 * 认证配置表单模态框 Emits
 */
export interface AuthConfigFormModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

