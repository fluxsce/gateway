/**
 * 集群事件类型定义
 */

/**
 * 集群事件
 */
export interface ClusterEvent {
  /** 事件ID，主键 */
  eventId: string
  /** 租户ID */
  tenantId: string
  /** 发布节点ID(hostname:port) */
  sourceNodeId: string
  /** 发布节点IP */
  sourceNodeIp: string
  /** 事件类型 */
  eventType: string
  /** 事件动作(CREATE/UPDATE/DELETE/REFRESH/INVALIDATE) */
  eventAction: string
  /** 事件负载(JSON字符串) */
  eventPayload: string
  /** 事件时间 */
  eventTime: string
  /** 过期时间 */
  expireTime?: string | null
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记：Y-活动，N-非活动 */
  activeFlag: string
  /** 备注信息 */
  noteText: string
  /** 扩展属性 */
  extProperty: string
  /** 预留字段1 */
  reserved1: string
  /** 预留字段2 */
  reserved2: string
  /** 预留字段3 */
  reserved3: string
  /** 预留字段4 */
  reserved4: string
  /** 预留字段5 */
  reserved5: string
}

/**
 * 集群事件确认（处理节点）
 */
export interface ClusterEventAck {
  /** 确认ID，主键 */
  ackId: string
  /** 租户ID */
  tenantId: string
  /** 事件ID */
  eventId: string
  /** 处理节点ID(hostname:port) */
  nodeId: string
  /** 处理节点IP */
  nodeIp: string
  /** 确认状态(PENDING/SUCCESS/FAILED/SKIPPED) */
  ackStatus: string
  /** 处理时间 */
  processTime?: string | null
  /** 结果信息 */
  resultMessage: string
  /** 重试次数 */
  retryCount: number
  /** 创建时间 */
  addTime: string
  /** 创建人ID */
  addWho: string
  /** 最后修改时间 */
  editTime: string
  /** 最后修改人ID */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 活动状态标记：Y-活动，N-非活动 */
  activeFlag: string
  /** 备注信息 */
  noteText: string
  /** 扩展属性 */
  extProperty: string
  /** 预留字段1 */
  reserved1: string
  /** 预留字段2 */
  reserved2: string
  /** 预留字段3 */
  reserved3: string
}
