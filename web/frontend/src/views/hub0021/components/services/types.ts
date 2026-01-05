/**
 * 服务定义选择器组件类型定义
 */

/**
 * 服务定义接口
 * 对应数据库表：HUB_GW_SERVICE_DEFINITION
 */
export interface ServiceDefinition {
  /** 租户ID */
  tenantId: string
  /** 服务定义ID */
  serviceDefinitionId: string
  /** 服务名称 */
  serviceName: string
  /** 服务描述 */
  serviceDesc?: string
  /** 服务类型：0-静态配置，1-服务发现 */
  serviceType: number
  /** 发现类型 */
  discoveryType?: string
  /** 发现配置 */
  discoveryConfig?: Record<string, any>
  /** 负载均衡策略 */
  loadBalanceStrategy: string
  /** 健康检查是否启用 */
  healthCheckEnabled: 'Y' | 'N'
  /** 健康检查路径 */
  healthCheckPath: string
  /** 健康检查间隔（秒） */
  healthCheckIntervalSeconds: number
  /** 健康检查超时（毫秒） */
  healthCheckTimeoutMs: number
  /** 健康阈值 */
  healthyThreshold: number
  /** 不健康阈值 */
  unhealthyThreshold: number
  /** 服务元数据 */
  serviceMetadata?: Record<string, any>
  /** 活动状态标记，Y启用/N禁用 */
  activeFlag: 'Y' | 'N'
  /** 创建时间 */
  addTime?: string
  /** 创建人 */
  addWho?: string
  /** 修改时间 */
  editTime?: string
  /** 修改人 */
  editWho?: string
}

