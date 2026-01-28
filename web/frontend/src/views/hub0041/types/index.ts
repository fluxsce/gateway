// 命名空间基础类型 - 严格按照HUB_SERVICE_NAMESPACE表结构定义
export interface Namespace {
  // 主键和租户信息
  namespaceId: string // 命名空间ID，主键
  tenantId: string // 租户ID，用于多租户数据隔离

  // 关联服务中心实例
  instanceName: string // 服务中心实例名称，关联 HUB_SERVICE_INSTANCE
  environment: 'DEVELOPMENT' | 'STAGING' | 'PRODUCTION' // 部署环境

  // 命名空间基本信息
  namespaceName: string // 命名空间名称
  namespaceDescription?: string // 命名空间描述

  // 命名空间配置
  serviceQuotaLimit: number // 服务数量配额限制，0表示无限制，默认200
  configQuotaLimit: number // 配置数量配额限制，0表示无限制，默认200

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
}

