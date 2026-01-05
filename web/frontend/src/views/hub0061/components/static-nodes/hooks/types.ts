/**
 * 静态节点列表类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

import type {
    ActiveFlag,
    HealthCheckStatus,
    NodeStatus,
    ProxyType,
    TunnelStaticNode
} from '../../../types'

// 重新导出需要的类型
export type {
    ActiveFlag, HealthCheckStatus, NodeStatus,
    ProxyType, TunnelStaticNode
}

// ============= 常量定义 =============

/** 节点状态选项 */
export const NODE_STATUS_OPTIONS = [
  { label: '活跃', value: 'active' as NodeStatus, type: 'success' as const },
  { label: '非活跃', value: 'inactive' as NodeStatus, type: 'warning' as const },
  { label: '错误', value: 'error' as NodeStatus, type: 'error' as const },
]

/** 代理类型选项 */
export const PROXY_TYPE_OPTIONS = [
  { label: 'TCP', value: 'tcp' as ProxyType },
  { label: 'UDP', value: 'udp' as ProxyType },
]

/** 健康检查状态选项 */
export const HEALTH_CHECK_STATUS_OPTIONS = [
  { label: '健康', value: 'healthy' as HealthCheckStatus, type: 'success' as const },
  { label: '不健康', value: 'unhealthy' as HealthCheckStatus, type: 'error' as const },
  { label: '未知', value: 'unknown' as HealthCheckStatus, type: 'default' as const },
]

/** 活动标记选项 */
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' as ActiveFlag },
  { label: '禁用', value: 'N' as ActiveFlag },
]

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 静态节点列表模态框 Props
 */
export interface StaticNodeListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 静态服务器ID（必需，用于关联节点） */
  tunnelStaticServerId: string
  /** 静态服务器名称（用于显示） */
  serverName?: string
}

/**
 * 静态节点列表模态框 Emits
 */
export interface StaticNodeListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

