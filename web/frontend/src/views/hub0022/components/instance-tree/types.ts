/**
 * 网关实例树组件类型定义
 */

import type { TreeOption } from 'naive-ui'

/**
 * 网关实例类型
 * 对应数据库表：HUB_GATEWAY_INSTANCE
 * 用于表示网关实例的完整配置信息
 */
export interface GatewayInstance {
  /** 网关实例ID，唯一标识 */
  gatewayInstanceId: string

  /** 实例名称 */
  instanceName: string

  /** 实例描述 */
  instanceDesc?: string

  /** 服务器主机地址 */
  serverHost: string

  /** 绑定地址，通常为 0.0.0.0 或具体IP */
  bindAddress: string

  /** 监听端口 */
  listenPort: number

  /** HTTP端口 */
  httpPort?: number

  /** HTTPS端口 */
  httpsPort?: number

  /** 是否启用TLS，Y启用/N禁用 */
  tlsEnabled: 'Y' | 'N'

  /** 证书文件路径 */
  certFilePath?: string

  /** 私钥文件路径 */
  keyFilePath?: string

  /** 最大连接数 */
  maxConnections: number

  /** 读取超时时间（毫秒） */
  readTimeoutMs: number

  /** 写入超时时间（毫秒） */
  writeTimeoutMs: number

  /** 空闲连接超时时间（毫秒） */
  idleTimeoutMs: number

  /** 健康状态，Y健康/N异常 */
  healthStatus: 'Y' | 'N'

  /** 最后心跳时间 */
  lastHeartbeatTime?: string

  /** 实例元数据，JSON格式字符串 */
  instanceMetadata?: string

  /** 活动状态标记，Y启用/N禁用 */
  activeFlag: 'Y' | 'N'

  /** 创建时间 */
  addTime?: string

  /** 最后修改时间 */
  editTime?: string

  /** 备注信息 */
  noteText?: string
}

/**
 * 代理类型枚举
 */
export enum ProxyType {
  HTTP = 'http',
  WEBSOCKET = 'websocket',
  TCP = 'tcp',
  UDP = 'udp',
}

/**
 * 代理配置类型
 * 对应数据库表：HUB_GW_PROXY_CONFIG
 * 对应后端模型：ProxyConfig
 */
export interface ProxyConfig {
  /** 租户ID，联合主键 */
  tenantId: string

  /** 代理配置ID，联合主键 */
  proxyConfigId: string

  /** 网关实例ID（代理配置仅支持实例级） */
  gatewayInstanceId: string

  /** 代理名称 */
  proxyName: string

  /** 代理类型（http, websocket, tcp, udp） */
  proxyType: 'http' | 'websocket' | 'tcp' | 'udp'

  /** 代理ID（来自ProxyConfig.ID） */
  proxyId: string

  /** 配置优先级，数值越小优先级越高 */
  configPriority: number

  /** 代理具体配置，JSON格式，根据proxyType存储对应配置 */
  proxyConfig: string

  /** 自定义配置，JSON格式 */
  customConfig: string

  /** 预留字段1 */
  reserved1?: string

  /** 预留字段2 */
  reserved2?: string

  /** 预留字段3 */
  reserved3?: number

  /** 预留字段4 */
  reserved4?: number

  /** 预留字段5 */
  reserved5?: string

  /** 扩展属性，JSON格式 */
  extProperty?: string

  /** 创建时间 */
  addTime?: string

  /** 创建人ID */
  addWho?: string

  /** 最后修改时间 */
  editTime?: string

  /** 最后修改人ID */
  editWho?: string

  /** 操作序列标识 */
  oprSeqFlag?: string

  /** 当前版本号 */
  currentVersion?: number

  /** 活动状态标记（N非活动/禁用，Y活动/启用） */
  activeFlag: 'Y' | 'N'

  /** 备注信息 */
  noteText?: string
}

/**
 * 实例树节点选项类型
 * 扩展了 naive-ui 的 TreeOption，添加了实例信息
 */
export interface InstanceTreeOption extends TreeOption {
  /** 关联的网关实例对象 */
  instance?: GatewayInstance
}

