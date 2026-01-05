/**
 * Hub0040 服务治理模块 - 类型定义和静态配置文件
 * 
 * 本文件定义了服务治理中命名空间管理相关的所有TypeScript类型和静态配置，包括：
 * - 类型别名：服务分组类型、负载均衡策略、协议类型
 * - 接口类型：服务分组、请求参数、响应数据等
 * - 静态选项配置：服务分组类型选项、协议类型选项、负载均衡策略选项等
 * 
 * @author 系统架构组
 * @version 1.0.0
 * @since 2024-01-01
 */

// ==================== 类型别名定义 ====================

/**
 * 服务分组类型
 * 用于区分不同用途的服务分组
 */
export type ServiceGroupType = 'BUSINESS' | 'SYSTEM' | 'TEST'

/**
 * 负载均衡策略类型
 * 定义服务实例间的负载分配策略
 */
export type LoadBalanceStrategy = 'ROUND_ROBIN' | 'WEIGHTED_ROUND_ROBIN' | 'LEAST_CONNECTIONS' | 'RANDOM' | 'IP_HASH'

/**
 * 协议类型
 * 定义服务支持的网络协议类型
 */
export type ProtocolType = 'HTTP' | 'HTTPS' | 'TCP' | 'UDP'

// ==================== 核心数据接口定义 ====================

/**
 * 服务分组（命名空间）接口
 * 对应数据库表：HUB_REGISTRY_SERVICE_GROUP
 */
export interface ServiceGroup {
  /** 服务分组ID - 主键，唯一标识符 */
  serviceGroupId: string
  /** 租户ID - 用于多租户数据隔离 */
  tenantId: string
  /** 分组名称 - 命名空间的显示名称，必须唯一 */
  groupName: string
  /** 分组描述 - 可选的详细描述信息 */
  groupDescription?: string
  /** 分组类型 - 业务类型、系统类型或测试类型 */
  groupType: ServiceGroupType
  /** 所有者用户ID - 拥有该命名空间的用户ID */
  ownerUserId: string
  /** 所有者用户名 - 所有者的显示名称（查询时填充） */
  ownerUserName?: string
  /** 管理员用户ID列表 - 拥有管理权限的用户ID数组 */
  adminUserIds?: string[]
  /** 管理员用户名列表 - 管理员的显示名称数组（查询时填充） */
  adminUserNames?: string[]
  /** 只读用户ID列表 - 拥有只读权限的用户ID数组 */
  readUserIds?: string[]
  /** 只读用户名列表 - 只读用户的显示名称数组（查询时填充） */
  readUserNames?: string[]
  /** 访问控制启用标识 - Y:启用访问控制, N:禁用访问控制 */
  accessControlEnabled: 'Y' | 'N'
  /** 默认协议类型 - 该命名空间下服务的默认网络协议 */
  defaultProtocolType: ProtocolType
  /** 默认负载均衡策略 - 该命名空间下服务的默认负载均衡算法 */
  defaultLoadBalanceStrategy: LoadBalanceStrategy
  /** 默认健康检查URL - 服务健康检查的默认路径 */
  defaultHealthCheckUrl: string
  /** 默认健康检查间隔秒数 - 健康检查的默认时间间隔 */
  defaultHealthCheckIntervalSeconds: number
  /** 创建时间 - 记录创建的时间戳 */
  addTime: string
  /** 创建人ID - 创建该记录的用户ID */
  addWho: string
  /** 创建人姓名 - 创建人的显示名称（查询时填充） */
  addWhoName?: string
  /** 最后修改时间 - 记录最后修改的时间戳 */
  editTime: string
  /** 最后修改人ID - 最后修改该记录的用户ID */
  editWho: string
  /** 最后修改人姓名 - 最后修改人的显示名称（查询时填充） */
  editWhoName?: string
  /** 操作序列标识 - 用于乐观锁控制并发修改 */
  oprSeqFlag: string
  /** 当前版本号 - 记录版本，用于版本控制 */
  currentVersion: number
  /** 活动状态标识 - Y:活动状态, N:非活动状态 */
  activeFlag: 'Y' | 'N'
  /** 备注信息 - 可选的备注文本 */
  noteText?: string
  /** 扩展属性 - JSON格式的扩展配置信息 */
  extProperty?: Record<string, any>
  
  // ==================== 统计信息字段 ====================
  /** 服务数量 - 该命名空间下的服务总数（统计字段） */
  serviceCount?: number
  /** 实例数量 - 该命名空间下的服务实例总数（统计字段） */
  instanceCount?: number
}

// ==================== 请求参数接口定义 ====================

/**
 * 创建服务分组请求参数
 * 用于新建命名空间时的数据传输
 */
export interface ServiceGroupCreateRequest {
  /** 分组名称 - 必填，命名空间的唯一标识名称 */
  groupName: string
  /** 分组描述 - 可选，命名空间的详细描述 */
  groupDescription?: string
  /** 分组类型 - 必填，指定命名空间的用途类型 */
  groupType: ServiceGroupType
  /** 管理员用户ID列表 - 可选，指定拥有管理权限的用户 */
  adminUserIds?: string[]
  /** 只读用户ID列表 - 可选，指定拥有只读权限的用户 */
  readUserIds?: string[]
  /** 访问控制启用标识 - 必填，是否启用访问控制 */
  accessControlEnabled: 'Y' | 'N'
  /** 默认协议类型 - 可选，默认为HTTP */
  defaultProtocolType?: ProtocolType
  /** 默认负载均衡策略 - 可选，默认为轮询 */
  defaultLoadBalanceStrategy?: LoadBalanceStrategy
  /** 默认健康检查URL - 可选，默认为/health */
  defaultHealthCheckUrl?: string
  /** 默认健康检查间隔秒数 - 可选，默认为30秒 */
  defaultHealthCheckIntervalSeconds?: number
  /** 备注信息 - 可选，额外的备注说明 */
  noteText?: string
  /** 扩展属性 - 可选，JSON格式的扩展配置 */
  extProperty?: Record<string, any>
}

/**
 * 更新服务分组请求参数
 * 继承创建请求的所有字段，但都为可选，并添加ID字段和活动状态
 */
export interface ServiceGroupUpdateRequest extends Partial<ServiceGroupCreateRequest> {
  /** 服务分组ID - 必填，要更新的命名空间ID */
  serviceGroupId: string
  /** 活动状态标识 - 可选，Y:活动状态, N:非活动状态 */
  activeFlag?: 'Y' | 'N'
}

/**
 * 查询服务分组请求参数
 * 用于分页查询和条件筛选
 * 注意：租户ID由后端自动从session中获取，前端无需传入
 */
export interface ServiceGroupQueryRequest {
  /** 分组类型 - 可选，按类型筛选 */
  groupType?: ServiceGroupType
  /** 所有者用户ID - 可选，按所有者筛选 */
  ownerUserId?: string
  /** 活动状态标识 - 可选，按状态筛选 */
  activeFlag?: 'Y' | 'N'
  /** 页码 - 必填，从1开始 */
  pageIndex: number
  /** 每页大小 - 必填，建议10-100之间 */
  pageSize: number
}

// ==================== 静态选项配置 ====================

/**
 * 服务分组类型选项
 */
export const serviceGroupTypeOptions: Array<{
  label: string
  value: ServiceGroupType
  description?: string
}> = [
  {
    label: '业务类型',
    value: 'BUSINESS',
    description: '用于业务相关的服务分组'
  },
  {
    label: '系统类型',
    value: 'SYSTEM',
    description: '用于系统基础设施相关的服务分组'
  },
  {
    label: '测试类型',
    value: 'TEST',
    description: '用于测试环境的服务分组'
  }
]

/**
 * 协议类型选项
 */
export const protocolTypeOptions: Array<{
  label: string
  value: ProtocolType
  description?: string
}> = [
  {
    label: 'HTTP',
    value: 'HTTP',
    description: 'HTTP协议'
  },
  {
    label: 'HTTPS',
    value: 'HTTPS',
    description: 'HTTPS协议'
  },
  {
    label: 'TCP',
    value: 'TCP',
    description: 'TCP协议'
  },
  {
    label: 'UDP',
    value: 'UDP',
    description: 'UDP协议'
  }
]

/**
 * 负载均衡策略选项
 */
export const loadBalanceStrategyOptions: Array<{
  label: string
  value: LoadBalanceStrategy
  description?: string
}> = [
  {
    label: '轮询',
    value: 'ROUND_ROBIN',
    description: '按顺序轮询所有可用实例'
  },
  {
    label: '加权轮询',
    value: 'WEIGHTED_ROUND_ROBIN',
    description: '根据权重进行轮询'
  },
  {
    label: '最少连接',
    value: 'LEAST_CONNECTIONS',
    description: '选择连接数最少的实例'
  },
  {
    label: '随机',
    value: 'RANDOM',
    description: '随机选择可用实例'
  },
  {
    label: 'IP哈希',
    value: 'IP_HASH',
    description: '根据客户端IP进行哈希选择'
  }
]

/**
 * 访问控制选项
 */
export const accessControlOptions: Array<{
  label: string
  value: 'Y' | 'N'
  description?: string
}> = [
  {
    label: '启用',
    value: 'Y',
    description: '启用访问控制，只有授权用户可以访问'
  },
  {
    label: '禁用',
    value: 'N',
    description: '禁用访问控制，所有用户都可以访问'
  }
]

/**
 * 状态选项
 */
export const statusOptions: Array<{
  label: string
  value: 'Y' | 'N'
  description?: string
}> = [
  {
    label: '活动',
    value: 'Y',
    description: '命名空间处于活动状态'
  },
  {
    label: '非活动',
    value: 'N',
    description: '命名空间处于非活动状态'
  }
]



// ==================== 说明注释 ====================

/**
 * API响应格式说明：
 * 本模块使用项目标准的 JsonDataObj 类型作为API响应格式，
 * 该类型定义在 @/types/api 中，包含以下字段：
 * - success: boolean - 请求是否成功
 * - code: number - 响应状态码
 * - message: string - 响应消息
 * - data: any - 响应数据
 * - timestamp?: number - 时间戳
 */
