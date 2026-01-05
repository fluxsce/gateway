/**
 * CORS配置类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * CORS配置类型
 * 严格按照 HUB_GATEWAY_CORS_CONFIG 表结构定义
 */
export interface CorsConfig {
  tenantId: string // 租户ID
  corsConfigId: string // CORS配置ID
  gatewayInstanceId?: string // 网关实例ID(实例级CORS)
  routeConfigId?: string // 路由配置ID(路由级CORS)
  configName: string // 配置名称
  allowOrigins: string[] // 允许的源,JSON数组格式
  allowMethods: string // 允许的HTTP方法,逗号分隔
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

// ============= 组件 Props 和 Emits 类型 =============

/**
 * CORS配置表单模态框 Props
 */
export interface CorsConfigFormModalProps {
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
  /** 网关实例ID（实例级CORS） */
  gatewayInstanceId?: string
  /** 路由配置ID（路由级CORS） */
  routeConfigId?: string
}

/**
 * CORS配置表单模态框 Emits
 */
export interface CorsConfigFormModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

