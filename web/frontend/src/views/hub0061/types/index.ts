/**
 * 静态端口映射管理模块类型定义
 * 与后端 internal/tunnel/types/tunnel_types.go 保持一致
 */

// ============================================================
// 枚举定义
// ============================================================

/** 服务器类型 */
export type ServerType = 'tcp' | 'udp'

/** 代理类型 */
export type ProxyType = 'tcp' | 'udp'

/** 服务器状态 */
export type ServerStatus = 'running' | 'stopped' | 'error'

/** 节点状态 */
export type NodeStatus = 'active' | 'inactive' | 'error'

/** 健康检查类型 */
export type HealthCheckType = 'tcp' | 'http' | 'https'

/** 健康检查状态 */
export type HealthCheckStatus = 'healthy' | 'unhealthy' | 'unknown'

/** 负载均衡类型 */
export type LoadBalanceType = 'roundrobin' | 'leastconn' | 'random'

/** 活动标记 */
export type ActiveFlag = 'Y' | 'N'

/** TLS启用标记 */
export type TlsEnable = 'Y' | 'N'

/** 压缩启用标记 */
export type CompressionEnable = 'Y' | 'N'

/** 加密启用标记 */
export type EncryptionEnable = 'Y' | 'N'

// ============================================================
// 静态服务器类型定义
// ============================================================

/** 静态隧道服务器配置 - 对应 TunnelStaticServer */
export interface TunnelStaticServer {
  tunnelStaticServerId: string           // 静态隧道服务器ID，主键
  tenantId: string                       // 租户ID
  serverName: string                     // 服务器名称
  serverDescription?: string | null      // 服务器描述
  listenAddress: string                  // 监听地址
  listenPort: number                     // 监听端口（公网端口）
  serverType: ServerType                 // 服务器类型(tcp,udp)
  maxConnections: number                 // 最大连接数
  connectionTimeout: number              // 连接超时时间(秒)
  readTimeout: number                    // 读取超时时间(秒)
  writeTimeout: number                   // 写入超时时间(秒)
  tlsEnable: TlsEnable                   // 启用TLS(N禁用,Y启用)
  tlsCertFile?: string | null            // TLS证书文件路径
  tlsKeyFile?: string | null             // TLS私钥文件路径
  tlsCaFile?: string | null              // TLS CA证书文件路径
  logLevel: string                       // 日志级别(debug,info,warn,error)
  logFile?: string | null                // 日志文件路径
  serverStatus: ServerStatus             // 服务器状态(running,stopped,error)
  startTime?: string | null              // 服务启动时间
  stopTime?: string | null               // 服务停止时间
  currentConnectionCount: number         // 当前连接数
  totalConnectionCount: number           // 总连接数
  totalBytesReceived: number             // 总接收字节数
  totalBytesSent: number                 // 总发送字节数
  healthCheckType?: HealthCheckType | null      // 健康检查类型(tcp,http,https)
  healthCheckUrl?: string | null                // 健康检查URL
  healthCheckInterval?: number | null           // 健康检查间隔(秒)
  healthCheckTimeout?: number | null            // 健康检查超时(秒)
  healthCheckMaxFailures?: number | null        // 健康检查最大失败次数
  loadBalanceType?: LoadBalanceType | null      // 负载均衡类型(roundrobin,leastconn,random)
  serverConfig?: string | null                  // 服务器配置，JSON格式
  nodes?: TunnelStaticNode[]                    // 后端节点列表

  // 通用字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: ActiveFlag
  noteText?: string | null
  extProperty?: string | null

  // 关联数据
  nodeCount?: number                     // 节点数量
}

/** 静态服务器查询请求参数 */
export interface StaticServerQueryParams {
  pageIndex: number
  pageSize: number
  serverName?: string                    // 服务器名称（模糊匹配）
  serverDescription?: string             // 服务器描述（模糊匹配）
  listenAddress?: string                 // 监听地址
  listenPort?: number                    // 监听端口
  serverStatus?: ServerStatus            // 服务器状态过滤
  serverType?: ServerType                // 服务器类型过滤
  activeFlag?: ActiveFlag                // 活动标记过滤
}

// ============================================================
// 静态节点类型定义
// ============================================================

/** 静态隧道节点配置 - 对应 TunnelStaticNode */
export interface TunnelStaticNode {
  tunnelStaticNodeId: string             // 静态隧道节点ID，主键
  tenantId: string                       // 租户ID
  tunnelStaticServerId: string           // 静态隧道服务器ID
  nodeName: string                       // 节点名称
  nodeDescription?: string | null        // 节点描述
  targetAddress: string                  // 目标地址（后端服务地址）
  targetPort: number                     // 目标端口（后端服务端口）
  proxyType: ProxyType                   // 代理类型(tcp,udp)
  maxConnections?: number | null         // 最大连接数
  connectionTimeout?: number | null      // 连接超时时间(秒)
  readTimeout?: number | null            // 读取超时时间(秒)
  writeTimeout?: number | null           // 写入超时时间(秒)
  retryCount?: number | null             // 重试次数
  retryInterval?: number | null          // 重试间隔(秒)
  compression: CompressionEnable         // 启用压缩(N禁用,Y启用)
  encryption: EncryptionEnable           // 启用加密(N禁用,Y启用)
  secretKey?: string | null              // 加密密钥
  customHeaders?: string | null          // 自定义HTTP头，JSON格式
  nodeStatus: NodeStatus                 // 节点状态(active,inactive,error)
  lastHealthCheck?: string | null        // 最后健康检查时间
  healthCheckStatus?: HealthCheckStatus | null  // 健康检查状态(healthy,unhealthy,unknown)
  currentConnectionCount: number         // 当前连接数
  totalConnectionCount: number           // 总连接数
  totalBytesReceived: number             // 总接收字节数
  totalBytesSent: number                 // 总发送字节数
  failureCount: number                   // 失败次数
  lastFailureTime?: string | null        // 最后失败时间
  nodeConfig?: string | null             // 节点配置，JSON格式

  // 通用字段
  addTime: string
  addWho: string
  editTime: string
  editWho: string
  oprSeqFlag: string
  currentVersion: number
  activeFlag: ActiveFlag
  noteText?: string | null
  extProperty?: string | null

  // 关联数据
  serverName?: string                    // 所属服务器名称
}

/** 静态节点查询请求参数 */
export interface StaticNodeQueryParams {
  pageIndex: number
  pageSize: number
  tunnelStaticServerId?: string          // 所属服务器ID
  nodeName?: string                      // 节点名称（模糊匹配）
  nodeDescription?: string               // 节点描述（模糊匹配）
  targetAddress?: string                 // 目标地址（模糊匹配）
  targetPort?: number                    // 目标端口
  nodeStatus?: NodeStatus                // 节点状态过滤
  proxyType?: ProxyType                  // 代理类型过滤
  healthCheckStatus?: HealthCheckStatus  // 健康检查状态过滤
  activeFlag?: ActiveFlag                // 活动标记过滤
}

/** 端口冲突检查参数 */
export interface PortConflictCheckParams {
  listenAddress: string
  listenPort: number
  serverType: ServerType
  excludeId?: string                     // 编辑时传入，用于排除自身
}
