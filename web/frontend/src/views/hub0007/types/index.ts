/**
 * 系统节点监控模块类型定义
 */

/**
 * 系统节点信息接口
 */
export interface ServerInfo {
  /** 节点ID，联合主键 */
  metricServerId: string
  /** 租户ID，联合主键 */
  tenantId: string
  /** 主机名 */
  hostname: string
  /** 操作系统类型 */
  osType: string
  /** 操作系统版本 */
  osVersion: string
  /** 内核版本 */
  kernelVersion?: string
  /** 系统架构 */
  architecture: string
  /** 系统启动时间 */
  bootTime: string
  /** 主IP地址 */
  ipAddress?: string
  /** 主MAC地址 */
  macAddress?: string
  /** 服务器位置 */
  serverLocation?: string
  /** 服务器类型(physical/virtual/unknown) */
  serverType: string
  /** 最后更新时间 */
  lastUpdateTime: string
  /** 网络信息详情，JSON格式 */
  networkInfo?: string
  /** 系统详细信息，JSON格式 */
  systemInfo?: string
  /** 硬件信息，JSON格式 */
  hardwareInfo?: string
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
  /** 活动状态标记(N非活动,Y活动) */
  activeFlag: string
  /** 备注信息 */
  noteText?: string
  /** 扩展属性，JSON格式 */
  extProperty?: string
  /** 预留字段1 */
  reserved1?: string
  /** 预留字段2 */
  reserved2?: string
  /** 预留字段3 */
  reserved3?: string
  /** 预留字段4 */
  reserved4?: string
  /** 预留字段5 */
  reserved5?: string
  /** 预留字段6 */
  reserved6?: string
  /** 预留字段7 */
  reserved7?: string
  /** 预留字段8 */
  reserved8?: string
  /** 预留字段9 */
  reserved9?: string
  /** 预留字段10 */
  reserved10?: string
}

/**
 * 系统节点查询条件接口
 */
export interface ServerInfoQuery {
  /** 主机名（模糊查询） */
  hostname?: string
  /** IP地址（模糊查询） */
  ipAddress?: string
  /** 操作系统类型 */
  osType?: string
  /** 服务器类型 */
  serverType?: string
  /** 服务器位置（模糊查询） */
  serverLocation?: string
  /** 活动标记 */
  activeFlag?: string
  /** 页码 */
  page?: number
  /** 每页数量 */
  pageSize?: number
}

/**
 * 服务器类型枚举
 */
export enum ServerType {
  /** 物理服务器 */
  PHYSICAL = 'physical',
  /** 虚拟服务器 */
  VIRTUAL = 'virtual',
  /** 未知类型 */
  UNKNOWN = 'unknown',
}

/**
 * 活动状态枚举
 */
export enum ActiveFlag {
  /** 活动 */
  ACTIVE = 'Y',
  /** 非活动 */
  INACTIVE = 'N',
}

/**
 * 操作系统类型枚举
 */
export enum OsType {
  /** Linux */
  LINUX = 'Linux',
  /** Windows */
  WINDOWS = 'Windows',
  /** MacOS */
  MACOS = 'MacOS',
  /** Unix */
  UNIX = 'Unix',
  /** 其他 */
  OTHER = 'Other',
}

/**
 * 基础字段接口
 */
interface BaseFields {
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
  /** 活动状态标记(N非活动,Y活动) */
  activeFlag?: string
  /** 备注信息 */
  noteText?: string
  /** 扩展属性，JSON格式 */
  extProperty?: string
}

/**
 * CPU使用率指标
 * 对应数据库表：HUB_METRIC_CPU_LOG
 */
export interface CPUMetrics extends BaseFields {
  /** CPU监控记录ID */
  metricCpuLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** CPU总使用率（百分比） */
  usagePercent: number
  /** 用户态使用率（百分比） */
  userPercent: number
  /** 系统态使用率（百分比） */
  systemPercent: number
  /** 空闲率（百分比） */
  idlePercent: number
  /** IO等待率（百分比） */
  ioWaitPercent: number
  /** 硬中断率（百分比） */
  irqPercent: number
  /** 软中断率（百分比） */
  softIrqPercent: number
  /** 物理核心数 */
  coreCount: number
  /** 逻辑核心数 */
  logicalCount: number
  /** 1分钟负载平均值 */
  loadAvg1: number
  /** 5分钟负载平均值 */
  loadAvg5: number
  /** 15分钟负载平均值 */
  loadAvg15: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 内存使用指标
 * 对应数据库表：HUB_METRIC_MEMORY_LOG
 */
export interface MemoryMetrics extends BaseFields {
  /** 内存监控记录ID */
  metricMemoryLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 总内存大小（字节） */
  totalMemory: number
  /** 可用内存大小（字节） */
  availableMemory: number
  /** 已用内存大小（字节） */
  usedMemory: number
  /** 内存使用率（百分比） */
  usagePercent: number
  /** 空闲内存大小（字节） */
  freeMemory: number
  /** 缓存内存大小（字节） */
  cachedMemory: number
  /** 缓冲区内存大小（字节） */
  buffersMemory: number
  /** 共享内存大小（字节） */
  sharedMemory: number
  /** 交换分区总大小（字节） */
  swapTotal: number
  /** 交换分区已用大小（字节） */
  swapUsed: number
  /** 交换分区空闲大小（字节） */
  swapFree: number
  /** 交换分区使用率（百分比） */
  swapUsagePercent: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 磁盘分区使用指标
 * 对应数据库表：HUB_METRIC_DISK_PARTITION_LOG
 */
export interface DiskPartition extends BaseFields {
  /** 磁盘分区监控记录ID */
  metricDiskPartitionLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 设备名称（如：/dev/sda1） */
  deviceName: string
  /** 挂载点（如：/、/home） */
  mountPoint: string
  /** 文件系统类型（如：ext4、xfs） */
  fileSystem: string
  /** 总空间大小（字节） */
  totalSpace: number
  /** 已用空间大小（字节） */
  usedSpace: number
  /** 空闲空间大小（字节） */
  freeSpace: number
  /** 空间使用率（百分比） */
  usagePercent: number
  /** 总inode数量 */
  inodesTotal: number
  /** 已用inode数量 */
  inodesUsed: number
  /** 空闲inode数量 */
  inodesFree: number
  /** inode使用率（百分比） */
  inodesUsagePercent: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 磁盘IO性能指标
 * 对应数据库表：HUB_METRIC_DISK_IO_LOG
 */
export interface DiskIOStats extends BaseFields {
  /** 磁盘IO监控记录ID */
  metricDiskIoLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 设备名称（如：sda、nvme0n1） */
  deviceName: string
  /** 读取次数 */
  readCount: number
  /** 写入次数 */
  writeCount: number
  /** 读取字节数 */
  readBytes: number
  /** 写入字节数 */
  writeBytes: number
  /** 读取耗时（毫秒） */
  readTime: number
  /** 写入耗时（毫秒） */
  writeTime: number
  /** 正在进行的IO操作数 */
  ioInProgress: number
  /** IO总耗时（毫秒） */
  ioTime: number
  /** 读取速率（字节/秒） */
  readRate: number
  /** 写入速率（字节/秒） */
  writeRate: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 网络接口统计指标
 * 对应数据库表：HUB_METRIC_NETWORK_LOG
 */
export interface NetworkInterface extends BaseFields {
  /** 网络监控记录ID */
  metricNetworkLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 网络接口名称（如：eth0、wlan0） */
  interfaceName: string
  /** 硬件地址（MAC地址） */
  hardwareAddr?: string
  /** IP地址列表（JSON格式） */
  ipAddresses?: string
  /** 接口状态（如：up、down） */
  interfaceStatus: string
  /** 接口类型（如：ethernet、wifi） */
  interfaceType?: string
  /** 接收字节数 */
  bytesReceived: number
  /** 发送字节数 */
  bytesSent: number
  /** 接收包数 */
  packetsReceived: number
  /** 发送包数 */
  packetsSent: number
  /** 接收错误数 */
  errorsReceived: number
  /** 发送错误数 */
  errorsSent: number
  /** 接收丢包数 */
  droppedReceived: number
  /** 发送丢包数 */
  droppedSent: number
  /** 接收速率（字节/秒） */
  receiveRate: number
  /** 发送速率（字节/秒） */
  sendRate: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 进程详细信息
 * 对应数据库表：HUB_METRIC_PROCESS_LOG
 */
export interface ProcessInfo extends BaseFields {
  /** 进程监控记录ID */
  metricProcessLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 进程ID */
  processId: number
  /** 父进程ID */
  parentProcessId?: number
  /** 进程名称 */
  processName: string
  /** 进程状态（如：running、sleeping、zombie） */
  processStatus: string
  /** 进程创建时间 */
  createTime: string
  /** 进程运行时间（秒） */
  runTime: number
  /** 进程内存使用量（字节） */
  memoryUsage: number
  /** 进程内存使用率（百分比） */
  memoryPercent: number
  /** 进程CPU使用率（百分比） */
  cpuPercent: number
  /** 线程数量 */
  threadCount: number
  /** 文件描述符数量 */
  fileDescriptorCount: number
  /** 命令行参数（JSON格式） */
  commandLine?: string
  /** 可执行文件路径 */
  executablePath?: string
  /** 工作目录 */
  workingDirectory?: string
  /** 数据采集时间 */
  collectTime: string
}

