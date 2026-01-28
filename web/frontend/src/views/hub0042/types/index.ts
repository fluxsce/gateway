// 服务基础类型 - 严格按照HUB_SERVICE表结构定义
export interface Service {
  // 主键和租户信息
  tenantId: string // 租户ID，用于多租户数据隔离
  namespaceId: string // 命名空间ID，关联HUB_SERVICE_NAMESPACE表
  groupName: string // 分组名称，如DEFAULT_GROUP
  serviceName: string // 服务名称，全局唯一标识

  // 服务类型
  serviceType: 'INTERNAL' | 'NACOS' | 'CONSUL' | 'EUREKA' | 'ETCD' | 'ZOOKEEPER' // 服务类型

  // 服务基本信息
  serviceVersion?: string // 服务版本号
  serviceDescription?: string // 服务描述
  externalServiceConfig?: string // 外部服务配置，JSON格式

  // 服务元数据
  metadataJson?: string // 服务元数据，JSON格式
  tagsJson?: string // 服务标签，JSON格式

  // 服务保护阈值
  protectThreshold?: number // 服务保护阈值，范围0.00-1.00

  // 服务选择器
  selectorJson?: string // 服务选择器，JSON格式

  // 系统字段
  addTime: string // 创建时间
  addWho: string // 创建人ID
  editTime: string // 最后修改时间
  editWho: string // 最后修改人ID
  oprSeqFlag: string // 操作序列标识
  currentVersion: number // 当前版本号
  activeFlag: 'Y' | 'N' // 活动状态标记(N非活动,Y活动)
  noteText?: string // 备注信息
  extProperty?: string // 扩展属性，JSON格式

  // 扩展字段（从缓存获取）
  nodeCount?: number // 节点数量
  healthyNodeCount?: number // 健康节点数量
  unhealthyNodeCount?: number // 不健康节点数量
  nodes?: ServiceNode[] // 节点列表
}

// 服务节点类型
export interface ServiceNode {
  nodeId: string // 节点ID
  ipAddress: string // IP地址
  portNumber: number // 端口号
  instanceStatus: 'UP' | 'DOWN' | 'STARTING' | 'OUT_OF_SERVICE' // 节点状态
  healthyStatus: 'HEALTHY' | 'UNHEALTHY' | 'UNKNOWN' // 健康状态
  ephemeral: 'Y' | 'N' // 是否临时节点
  weight: number // 权重值
  metadataJson?: string // 节点元数据，JSON格式
  registerTime?: string // 注册时间
  lastBeatTime?: string // 最后心跳时间
  lastCheckTime?: string // 最后健康检查时间
  activeFlag: 'Y' | 'N' // 活动状态标记
}

