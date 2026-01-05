/**
 * 隧道客户端与服务管理模块类型定义
 */

// 连接状态枚举
export enum ConnectionStatus {
  CONNECTED = 'connected',
  DISCONNECTED = 'disconnected',
  CONNECTING = 'connecting',
  ERROR = 'error'
}

// 服务状态枚举
export enum ServiceStatus {
  ACTIVE = 'active',
  INACTIVE = 'inactive',
  ERROR = 'error',
  OFFLINE = 'offline'
}

/**
 * 隧道客户端基础信息
 * 表: tuc_tunnel_client
 * 
 * 隧道客户端用于连接到远程隧道服务器，建立安全的网络隧道。
 * 客户端和服务器完全隔离，通过手动配置服务器地址、端口和认证令牌来建立连接。
 */
export interface TunnelClient {
  /** 客户端唯一标识 (主键) */
  tunnelClientId: string
  
  /** 租户ID */
  tenantId: string
  
  /** 用户ID */
  userId: string
  
  /** 客户端名称 (必填) */
  clientName: string
  
  /** 客户端描述 */
  clientDescription?: string
  
  /** 客户端版本号 */
  clientVersion?: string
  
  /** 操作系统类型 (如: Windows, Linux, macOS) */
  operatingSystem?: string
  
  /** 客户端IP地址 */
  clientIpAddress?: string
  
  /** 客户端MAC地址 */
  clientMacAddress?: string
  
  /** 服务器地址 (必填，手动输入) */
  serverAddress: string
  
  /** 服务器端口 (必填，手动输入，范围: 1-65535) */
  serverPort: number
  
  /** 认证令牌 (必填，手动输入，用于服务器身份验证) */
  authToken: string
  
  /** 是否启用TLS加密 ('Y': 启用, 'N': 不启用) */
  tlsEnable: 'Y' | 'N'
  
  /** 是否自动重连 ('Y': 启用, 'N': 不启用) */
  autoReconnect: 'Y' | 'N'
  
  /** 最大重试次数 (自动重连时使用) */
  maxRetries: number
  
  /** 重试间隔 (秒，自动重连时使用) */
  retryInterval: number
  
  /** 心跳间隔 (秒) */
  heartbeatInterval: number
  
  /** 心跳超时时间 (秒) */
  heartbeatTimeout: number
  
  /** 连接状态 (connected/disconnected/connecting/error) */
  connectionStatus: ConnectionStatus
  
  /** 最后连接时间 */
  lastConnectTime?: string
  
  /** 最后断开时间 */
  lastDisconnectTime?: string
  
  /** 总连接时长 (秒) */
  totalConnectTime: number
  
  /** 重连次数统计 */
  reconnectCount: number
  
  /** 服务数量统计 */
  serviceCount: number
  
  /** 最后心跳时间 */
  lastHeartbeat?: string
  
  /** 客户端配置 (JSON格式字符串) */
  clientConfig?: string
  
  /** 创建时间 */
  addTime: string
  
  /** 创建人 */
  addWho: string
  
  /** 修改时间 */
  editTime: string
  
  /** 修改人 */
  editWho: string
  
  /** 操作序列标识 */
  oprSeqFlag: string
  
  /** 当前版本号 (乐观锁) */
  currentVersion: number
  
  /** 激活标识 ('Y': 激活, 'N': 未激活) */
  activeFlag: 'Y' | 'N'
  
  /** 备注信息 */
  noteText?: string
  
  /** 扩展属性 (JSON格式) */
  extProperty?: string
}

// 客户端查询参数
export interface TunnelClientQueryParams {
  clientName?: string
  connectionStatus?: ConnectionStatus
  operatingSystem?: string
  serverAddress?: string
  userId?: string
  keyword?: string
  activeFlag?: 'Y' | 'N'
  pageIndex: number
  pageSize: number
}

// 客户端统计信息
export interface TunnelClientStats {
  totalClients: number
  connectedClients: number
  disconnectedClients: number
  connectingClients: number
  errorClients: number
  totalServices: number
}

/**
 * 隧道服务基础信息
 * 表: hub_tunnel_service
 * 
 * 隧道服务是客户端注册到服务器的具体服务配置（如SSH、HTTP等）
 */
export interface TunnelService {
  /** 服务唯一标识 (主键) */
  tunnelServiceId: string
  
  /** 租户ID */
  tenantId: string
  
  /** 客户端ID (外键) */
  tunnelClientId: string
  
  /** 用户ID */
  userId: string
  
  /** 服务名称 (必填) */
  serviceName: string
  
  /** 服务描述 */
  serviceDescription?: string
  
  /** 服务类型 (tcp/udp/http/https/stcp/sudp/xtcp) */
  serviceType: string
  
  /** 本地地址 (必填) */
  localAddress: string
  
  /** 本地端口 (必填，范围: 1-65535) */
  localPort: number
  
  /** 远程端口 (服务器分配的端口) */
  remotePort?: number
  
  /** 自定义域名 (JSON格式字符串) */
  customDomains?: string
  
  /** 子域名 */
  subDomain?: string
  
  /** HTTP认证用户名 */
  httpUser?: string
  
  /** HTTP认证密码 */
  httpPassword?: string
  
  /** Host头重写 */
  hostHeaderRewrite?: string
  
  /** 自定义HTTP头 (JSON格式) */
  headers?: string
  
  /** 路径映射 (JSON格式) */
  locations?: string
  
  /** 是否启用加密 ('Y': 启用, 'N': 不启用) */
  useEncryption: 'Y' | 'N'
  
  /** 是否启用压缩 ('Y': 启用, 'N': 不启用) */
  useCompression: 'Y' | 'N'
  
  /** 加密密钥 */
  secretKey?: string
  
  /** 带宽限制 */
  bandwidthLimit?: string
  
  /** 最大连接数 */
  maxConnections?: number
  
  /** 健康检查类型 */
  healthCheckType?: string
  
  /** 健康检查URL */
  healthCheckUrl?: string
  
  /** 服务状态 (active/inactive/error/offline) */
  serviceStatus: ServiceStatus
  
  /** 注册时间 */
  registeredTime: string
  
  /** 最后活动时间 */
  lastActiveTime?: string
  
  /** 当前连接数 */
  connectionCount: number
  
  /** 总连接数 */
  totalConnections: number
  
  /** 总流量 (字节) */
  totalTraffic: number
  
  /** 服务配置 (JSON格式字符串) */
  serviceConfig?: string
  
  /** 创建时间 */
  addTime: string
  
  /** 创建人 */
  addWho: string
  
  /** 修改时间 */
  editTime: string
  
  /** 修改人 */
  editWho: string
  
  /** 操作序列标识 */
  oprSeqFlag: string
  
  /** 当前版本号 (乐观锁) */
  currentVersion: number
  
  /** 激活标识 ('Y': 激活, 'N': 未激活) */
  activeFlag: 'Y' | 'N'
  
  /** 备注信息 */
  noteText?: string
  
  /** 扩展属性 (JSON格式) */
  extProperty?: string
}

// 服务查询参数
export interface TunnelServiceQueryParams {
  serviceName?: string
  tunnelClientId?: string
  serviceType?: string
  serviceStatus?: ServiceStatus
  userId?: string
  keyword?: string
  activeFlag?: 'Y' | 'N'
  pageIndex: number
  pageSize: number
}

// 服务统计信息
export interface TunnelServiceStats {
  totalServices: number
  activeServices: number
  inactiveServices: number
  errorServices: number
  offlineServices: number
  totalConnections: number
  totalTraffic: number
}

