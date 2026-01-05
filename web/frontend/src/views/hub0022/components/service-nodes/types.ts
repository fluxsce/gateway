/**
 * 服务节点管理类型定义
 * 统一管理业务类型，便于后续重构和维护
 */

// ============= 业务类型定义 =============

/**
 * 节点状态枚举
 * 表示服务节点的运行状态
 */
export enum NodeStatus {
  OFFLINE = 0, // 下线状态
  ONLINE = 1, // 在线状态
  MAINTENANCE = 2, // 维护状态
}

/**
 * 服务节点接口
 * 严格按照数据库表结构定义
 */
export interface ServiceNode {
  /** 租户ID */
  tenantId: string
  /** 服务节点ID */
  serviceNodeId: string
  /** 关联的网关实例ID */
  gatewayInstanceId: string
  /** 关联的服务定义ID */
  serviceDefinitionId: string
  /** 节点ID，来自NodeConfig.ID */
  nodeId: string
  /** 节点完整URL */
  nodeUrl: string
  /** 节点主机地址 */
  nodeHost: string
  /** 节点端口 */
  nodePort: number
  /** 节点协议（HTTP/HTTPS） */
  nodeProtocol: string
  /** 节点权重，用于负载均衡 */
  nodeWeight: number
  /** 健康状态，Y健康/N不健康 */
  healthStatus: 'Y' | 'N'
  /** 节点是否启用，Y启用/N禁用 */
  nodeEnabled: 'Y' | 'N'
  /** 节点元数据，JSON格式 */
  nodeMetadata?: string
  /** 节点运行状态，0下线/1在线/2维护 */
  nodeStatus: NodeStatus
  /** 最后健康检查时间 */
  lastHealthCheckTime?: string
  /** 健康检查结果详情 */
  healthCheckResult?: string
  /** 活动状态标记 */
  activeFlag: 'Y' | 'N'
  /** 备注信息 */
  noteText?: string
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
}

// ============= 组件 Props 和 Emits 类型 =============

/**
 * 服务节点列表模态框 Props
 */
export interface ServiceNodeListModalProps {
  /** 是否显示模态框 */
  visible: boolean
  /** 模态框标题 */
  title?: string
  /** 模态框宽度 */
  width?: number | string
  /** 挂载目标 */
  to?: string | HTMLElement | false
  /** 服务定义ID（必须提供，用于查询和新增） */
  serviceDefinitionId?: string
}

/**
 * 服务节点列表模态框 Emits
 */
export interface ServiceNodeListModalEmits {
  (e: 'update:visible', value: boolean): void
  (e: 'close'): void
  (e: 'refresh'): void
}

