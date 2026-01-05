/**
 * 服务定义选择器类型定义
 */

import type { ServiceDefinition } from '../types'

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 服务定义列表模态框 Props
 */
export interface ServiceDefinitionListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 网关实例ID（用于查询服务定义） */
  gatewayInstanceId?: string
}

/**
 * 服务定义列表模态框 Emits
 */
export interface ServiceDefinitionListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
  (e: 'select', services: ServiceDefinition[]): void
}

