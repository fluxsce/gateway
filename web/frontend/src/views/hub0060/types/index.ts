/**
 * 隧道服务器管理模块类型定义
 * 
 * 本文件定义了隧道服务器管理相关的所有TypeScript类型，包括：
 * - 枚举类型：服务器状态枚举
 * - 接口类型：隧道服务器、查询参数、表单数据、统计信息等
 * 
 * 对应后端模型：web/views/hub0060/models/tunnel_server.go
 * 对应数据库表：HUB_TUNNEL_SERVER
 * 
 * @author 系统架构组
 * @version 1.0.0
 */

// ==================== 枚举类型定义 ====================

/**
 * 隧道服务器状态枚举
 * 定义服务器的运行状态
 */
export enum TunnelServerStatus {
  /** 运行中 - 服务器正在运行 */
  RUNNING = 'running',
  /** 已停止 - 服务器已停止 */
  STOPPED = 'stopped',
  /** 错误 - 服务器运行出错 */
  ERROR = 'error'
}

// ==================== 核心数据接口定义 ====================

/**
 * 隧道服务器接口
 * 对应数据库表：HUB_TUNNEL_SERVER
 * 用途：管理隧道服务器信息
 */
export interface TunnelServer {
  // ==================== 主键信息 ====================
  /** 隧道服务器ID - 主键，唯一标识符 */
  tunnelServerId: string
  /** 租户ID - 用于多租户数据隔离 */
  tenantId: string
  /** 服务器名称 - 服务器的显示名称 */
  serverName: string
  /** 服务器描述 - 可选的详细描述信息 */
  serverDescription?: string

  // ==================== 服务器配置 ====================
  /** 控制端口监听地址 - 控制端口监听的IP地址，如: 0.0.0.0 */
  controlAddress: string
  /** 控制端口 - 接受客户端连接的控制端口 */
  controlPort: number
  /** 管理面板端口 - 管理面板的访问端口 */
  dashboardPort?: number
  /** 虚拟主机HTTP端口 - 虚拟主机HTTP服务的端口 */
  vhostHttpPort?: number
  /** 虚拟主机HTTPS端口 - 虚拟主机HTTPS服务的端口 */
  vhostHttpsPort?: number
  /** 最大客户端连接数 - 允许的最大客户端连接数 */
  maxClients: number
  /** Token认证启用状态 - Y:启用Token认证, N:禁用Token认证 */
  tokenAuth: 'Y' | 'N'
  /** 客户端认证Token - 客户端连接时使用的认证令牌 */
  authToken?: string
  /** TLS启用状态 - Y:启用TLS加密, N:禁用TLS加密 */
  tlsEnable: 'Y' | 'N'
  /** TLS证书文件路径 - TLS证书文件的完整路径 */
  tlsCertFile?: string
  /** TLS私钥文件路径 - TLS私钥文件的完整路径 */
  tlsKeyFile?: string
  /** 心跳间隔(秒) - 服务器与客户端之间的心跳检测间隔时间 */
  heartbeatInterval: number
  /** 心跳超时(秒) - 心跳检测的超时时间 */
  heartbeatTimeout: number
  /** 日志级别 - 日志记录的级别：debug, info, warn, error */
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  /** 每个客户端最大端口数 - 单个客户端可以使用的最大端口数 */
  maxPortsPerClient?: number
  /** 允许的端口范围 - JSON格式的端口范围配置，如: "10000-20000" */
  allowPorts?: string

  // ==================== 状态信息 ====================
  /** 服务器状态 - 服务器的运行状态：running(运行中), stopped(已停止), error(错误) */
  serverStatus: TunnelServerStatus
  /** 服务启动时间 - 服务器启动的时间戳 */
  startTime?: string
  /** 配置版本号 - 服务器配置的版本号，用于配置变更追踪 */
  configVersion?: string

  // ==================== 通用审计字段 ====================
  /** 创建时间 - 记录创建的时间戳 */
  addTime: string
  /** 创建人ID - 创建该记录的用户ID */
  addWho: string
  /** 最后修改时间 - 记录最后修改的时间戳 */
  editTime: string
  /** 最后修改人ID - 最后修改该记录的用户ID */
  editWho: string
  /** 操作序列标识 - 用于乐观锁控制并发修改 */
  oprSeqFlag: string
  /** 当前版本号 - 记录版本，用于版本控制 */
  currentVersion: number
  /** 活动状态标识 - Y:活动状态, N:非活动状态 */
  activeFlag: 'Y' | 'N'
  /** 备注信息 - 可选的备注文本 */
  noteText?: string
  /** 扩展属性 - JSON格式的扩展配置信息 */
  extProperty?: string
}

// ==================== 请求参数接口定义 ====================

/**
 * 隧道服务器查询参数接口
 * 用于分页查询和条件筛选
 * 注意：租户ID由后端自动从session中获取，前端无需传入
 */
export interface TunnelServerQueryParams {
  /** 租户ID - 可选，按租户筛选（通常由后端自动处理） */
  tenantId?: string
  /** 服务器名称 - 可选，按服务器名称模糊查询 */
  serverName?: string
  /** 服务器状态 - 可选，按服务器状态筛选 */
  serverStatus?: TunnelServerStatus
  /** 控制地址 - 可选，按控制地址筛选 */
  controlAddress?: string
  /** 控制端口 - 可选，按控制端口筛选 */
  controlPort?: number
  /** 活动状态标识 - 可选，按活动状态筛选：Y(活动), N(非活动) */
  activeFlag?: 'Y' | 'N'
  /** 页码 - 必填，从1开始 */
  pageIndex: number
  /** 每页大小 - 必填，建议10-100之间 */
  pageSize: number
}

/**
 * 隧道服务器创建/编辑表单接口
 * 用于新建和编辑隧道服务器时的数据传输
 */
export interface TunnelServerForm {
  /** 隧道服务器ID - 编辑时必填，新建时不需要 */
  tunnelServerId?: string
  /** 服务器名称 - 必填，服务器的显示名称 */
  serverName: string
  /** 服务器描述 - 可选，服务器的详细描述 */
  serverDescription?: string
  /** 控制端口监听地址 - 必填，如: 0.0.0.0 */
  controlAddress: string
  /** 控制端口 - 必填，接受客户端连接的控制端口 */
  controlPort: number
  /** 管理面板端口 - 可选，管理面板的访问端口 */
  dashboardPort?: number
  /** 虚拟主机HTTP端口 - 可选，虚拟主机HTTP服务的端口 */
  vhostHttpPort?: number
  /** 虚拟主机HTTPS端口 - 可选，虚拟主机HTTPS服务的端口 */
  vhostHttpsPort?: number
  /** 最大客户端连接数 - 必填，允许的最大客户端连接数 */
  maxClients: number
  /** Token认证启用状态 - 必填，Y:启用, N:禁用 */
  tokenAuth: 'Y' | 'N'
  /** 客户端认证Token - 可选，启用Token认证时必填 */
  authToken?: string
  /** TLS启用状态 - 必填，Y:启用, N:禁用 */
  tlsEnable: 'Y' | 'N'
  /** TLS证书文件路径 - 可选，启用TLS时必填 */
  tlsCertFile?: string
  /** TLS私钥文件路径 - 可选，启用TLS时必填 */
  tlsKeyFile?: string
  /** 心跳间隔(秒) - 必填，服务器与客户端之间的心跳检测间隔时间 */
  heartbeatInterval: number
  /** 心跳超时(秒) - 必填，心跳检测的超时时间 */
  heartbeatTimeout: number
  /** 日志级别 - 必填，日志记录的级别：debug, info, warn, error */
  logLevel: 'debug' | 'info' | 'warn' | 'error'
  /** 每个客户端最大端口数 - 可选，单个客户端可以使用的最大端口数 */
  maxPortsPerClient?: number
  /** 允许的端口范围 - 可选，JSON格式的端口范围配置 */
  allowPorts?: string
  /** 备注信息 - 可选，额外的备注说明 */
  noteText?: string
}

// ==================== 统计信息接口定义 ====================

/**
 * 隧道服务器统计信息接口
 * 用于展示隧道服务器的整体统计信息
 */
export interface TunnelServerStats {
  /** 总服务器数 - 所有隧道服务器的总数 */
  totalServers: number
  /** 运行中服务器数 - 状态为运行中的服务器数量 */
  runningServers: number
  /** 已停止服务器数 - 状态为已停止的服务器数量 */
  stoppedServers: number
  /** 错误服务器数 - 状态为错误的服务器数量 */
  errorServers: number
  /** 总客户端数 - 连接到所有服务器的客户端总数 */
  totalClients: number
  /** 总连接数 - 所有服务器的总连接数 */
  totalConnections: number
}

// ==================== 说明注释 ====================

/**
 * API响应格式说明：
 * 本模块使用项目标准的 JsonDataObj 类型作为API响应格式，
 * 该类型定义在 @/types/api 中，包含以下字段：
 * - oK: boolean - 请求是否成功
 * - state: boolean - 业务状态
 * - errMsg: string - 错误消息
 * - popMsg: string - 提示消息
 * - bizData: string - 业务数据（JSON字符串）
 * - pageQueryData: string - 分页信息（JSON字符串）
 */
