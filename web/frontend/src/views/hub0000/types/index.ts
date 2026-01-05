/**
 * Hub0000 系统指标采集模块类型定义
 * 基于数据库表结构设计，提供完整的系统监控数据类型
 */

/**
 * 数据库通用字段接口
 * 包含所有数据库表的公共字段，用于继承扩展
 */
interface BaseFields {
  /** 记录添加时间 */
  addTime: string
  /** 记录添加人员 */
  addWho: string
  /** 记录最后修改时间 */
  editTime: string
  /** 记录最后修改人员 */
  editWho: string
  /** 操作序列标识 */
  oprSeqFlag: string
  /** 当前版本号 */
  currentVersion: number
  /** 启用状态：Y-启用，N-停用 */
  activeFlag: 'Y' | 'N'
  /** 备注文本 */
  noteText?: string
  /** 扩展属性（JSON格式） */
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
 * 服务器基本信息
 * 对应数据库表：metric_server_info
 * 存储服务器的基本配置信息和硬件信息
 */
export interface ServerInfo extends BaseFields {
  /** 服务器唯一标识 */
  metricServerId: string
  /** 租户ID，用于多租户隔离 */
  tenantId: string
  /** 主机名 */
  hostname: string
  /** 操作系统类型（如：Linux、Windows） */
  osType: string
  /** 操作系统版本 */
  osVersion: string
  /** 内核版本 */
  kernelVersion?: string
  /** 系统架构（如：x64、arm64） */
  architecture: string
  /** 系统启动时间 */
  bootTime: string
  /** IP地址 */
  ipAddress?: string
  /** MAC地址 */
  macAddress?: string
  /** 服务器物理位置 */
  serverLocation?: string
  /** 服务器类型：物理机、虚拟机、未知 */
  serverType?: 'physical' | 'virtual' | 'unknown'
  /** 最后更新时间 */
  lastUpdateTime: string
  /** 网络信息（JSON格式） */
  networkInfo?: string
  /** 系统信息（JSON格式） */
  systemInfo?: string
  /** 硬件信息（JSON格式） */
  hardwareInfo?: string
}

/**
 * CPU使用率指标
 * 对应数据库表：metric_cpu_log
 * 记录CPU各项使用率指标和负载信息
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
 * 对应数据库表：metric_memory_log
 * 记录内存和交换分区的使用情况
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
 * 对应数据库表：metric_disk_partition_log
 * 记录各个磁盘分区的使用情况和inode信息
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
 * 对应数据库表：metric_disk_io_log
 * 记录磁盘读写性能统计信息
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
 * 对应数据库表：metric_network_log
 * 记录网络接口的流量和错误统计信息
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
 * 对应数据库表：metric_process_log
 * 记录系统中各个进程的详细运行状态
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

/**
 * 系统进程统计汇总
 * 对应数据库表：metric_process_stats_log
 * 记录系统进程数量的统计信息
 */
export interface ProcessSystemStats extends BaseFields {
  /** 进程统计记录ID */
  metricProcessStatsLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 正在运行的进程数 */
  runningCount: number
  /** 休眠状态的进程数 */
  sleepingCount: number
  /** 停止状态的进程数 */
  stoppedCount: number
  /** 僵尸进程数 */
  zombieCount: number
  /** 进程总数 */
  totalCount: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 系统温度监控信息
 * 对应数据库表：metric_temperature_log
 * 记录各个传感器的温度信息和阈值
 */
export interface TemperatureInfo extends BaseFields {
  /** 温度监控记录ID */
  metricTemperatureLogId: string
  /** 租户ID */
  tenantId: string
  /** 服务器ID */
  metricServerId: string
  /** 传感器名称（如：CPU、GPU、主板） */
  sensorName: string
  /** 温度值（摄氏度） */
  temperatureValue: number
  /** 高温阈值（摄氏度） */
  highThreshold?: number
  /** 临界温度阈值（摄氏度） */
  criticalThreshold?: number
  /** 数据采集时间 */
  collectTime: string
}

/**
 * 监控数据查询参数
 * 用于API查询时的参数过滤和分页
 */
export interface MetricQueryParams {
  /** 租户ID过滤 */
  tenantId?: string
  /** 服务器ID过滤 */
  metricServerId?: string
  /** 开始时间过滤 */
  startTime?: string
  /** 结束时间过滤 */
  endTime?: string
  /** 分页页码 */
  pageNum?: number
  /** 分页大小 */
  pageSize?: number
  /** 主机名过滤 */
  hostname?: string
  /** 操作系统类型过滤 */
  osType?: string
  /** 服务器类型过滤 */
  serverType?: string
  /** 启用状态过滤 */
  activeFlag?: 'Y' | 'N'
}

/**
 * 系统监控概览数据
 * 用于首页展示的汇总统计信息
 */
export interface SystemOverview {
  /** 服务器总数 */
  totalServers: number
  /** 在线服务器数 */
  onlineServers: number
  /** 离线服务器数 */
  offlineServers: number
  /** 平均CPU使用率（百分比） */
  avgCpuUsage: number
  /** 平均内存使用率（百分比） */
  avgMemoryUsage: number
  /** 平均磁盘使用率（百分比） */
  avgDiskUsage: number
  /** 系统总进程数 */
  totalProcesses: number
  /** 严重警告数量 */
  criticalAlerts: number
  /** 数据最后更新时间 */
  lastUpdateTime: string
}

/**
 * 实时监控数据
 * 用于实时监控页面展示的汇总数据
 */
export interface RealtimeMetrics {
  /** 服务器ID */
  serverId: string
  /** 主机名 */
  hostname: string
  /** CPU指标 */
  cpu: {
    /** CPU使用率（百分比） */
    usage: number
    /** 负载平均值数组 [1min, 5min, 15min] */
    loadAvg: number[]
  }
  /** 内存指标 */
  memory: {
    /** 内存使用量（字节） */
    usage: number
    /** 内存使用率（百分比） */
    usagePercent: number
    /** 总内存（字节） */
    total: number
    /** 可用内存（字节） */
    available: number
  }
  /** 磁盘指标 */
  disk: {
    /** 磁盘使用率（百分比） */
    usage: number
    /** 总空间（字节） */
    totalSpace: number
    /** 空闲空间（字节） */
    freeSpace: number
  }
  /** 网络指标 */
  network: {
    /** 接收速率（字节/秒） */
    receiveRate: number
    /** 发送速率（字节/秒） */
    sendRate: number
    /** 总流量（字节） */
    totalBytes: number
  }
  /** 进程指标 */
  processes: {
    /** 进程总数 */
    total: number
    /** 运行中进程数 */
    running: number
    /** 休眠进程数 */
    sleeping: number
    /** 僵尸进程数 */
    zombie: number
  }
  /** 温度指标（可选） */
  temperature?: {
    /** 温度值（摄氏度） */
    value: number
    /** 温度状态 */
    status: 'normal' | 'warning' | 'critical'
  }
  /** 数据时间戳 */
  timestamp: string
}

/**
 * 历史趋势数据点
 * 用于图表展示的时间序列数据
 */
export interface MetricTrend {
  /** 时间点 */
  time: string
  /** 数值 */
  value: number
  /** 标签（可选） */
  label?: string
}

/**
 * 警告级别枚举
 * 定义系统监控警告的严重程度
 */
export enum AlertLevel {
  /** 低级警告 */
  LOW = 'low',
  /** 中级警告 */
  MEDIUM = 'medium',
  /** 高级警告 */
  HIGH = 'high',
  /** 严重警告 */
  CRITICAL = 'critical',
}

/**
 * 监控警告信息
 * 用于警告管理和通知
 */
export interface MetricAlert {
  /** 警告唯一标识 */
  id: string
  /** 服务器ID */
  serverId: string
  /** 主机名 */
  hostname: string
  /** 警告类型 */
  type: 'cpu' | 'memory' | 'disk' | 'network' | 'process' | 'temperature'
  /** 警告级别 */
  level: AlertLevel
  /** 警告消息 */
  message: string
  /** 当前值 */
  value: number
  /** 阈值 */
  threshold: number
  /** 警告时间 */
  timestamp: string
  /** 是否已确认 */
  acknowledged: boolean
}

/**
 * 图表数据配置
 * 用于各种图表组件的配置
 */
export interface ChartConfig {
  /** 图表类型 */
  type: 'line' | 'bar' | 'pie' | 'gauge'
  /** 图表标题 */
  title: string
  /** 图表数据 */
  data: any[]
  /** 图表选项（可选） */
  options?: any
}

/**
 * 仪表盘面板配置
 * 用于自定义仪表盘布局
 */
export interface DashboardPanel {
  /** 面板唯一标识 */
  id: string
  /** 面板标题 */
  title: string
  /** 面板类型 */
  type: 'chart' | 'table' | 'stat' | 'alert'
  /** 面板位置和大小 */
  position: {
    /** X坐标 */
    x: number
    /** Y坐标 */
    y: number
    /** 宽度 */
    width: number
    /** 高度 */
    height: number
  }
  /** 面板配置 */
  config: ChartConfig | any
  /** 刷新间隔（毫秒，可选） */
  refreshInterval?: number
}

/**
 * 服务器状态枚举
 * 定义服务器的运行状态
 */
export enum ServerStatus {
  /** 在线状态 */
  ONLINE = 'online',
  /** 离线状态 */
  OFFLINE = 'offline',
  /** 警告状态 */
  WARNING = 'warning',
  /** 严重状态 */
  CRITICAL = 'critical',
}

/**
 * 指标类型枚举
 * 定义监控指标的分类
 */
export enum MetricType {
  /** CPU指标 */
  CPU = 'cpu',
  /** 内存指标 */
  MEMORY = 'memory',
  /** 磁盘指标 */
  DISK = 'disk',
  /** 网络指标 */
  NETWORK = 'network',
  /** 进程指标 */
  PROCESS = 'process',
  /** 温度指标 */
  TEMPERATURE = 'temperature',
}

/**
 * 时间范围选项
 * 用于时间选择器的配置
 */
export interface TimeRangeOption {
  /** 显示标签 */
  label: string
  /** 选项值 */
  value: string
  /** 时间长度（毫秒） */
  duration: number
}

/**
 * 导出BaseFields类型，供其他模块使用
 * 注意：BaseFields接口只能通过type导出，因为它没有直接export
 */
export type { BaseFields }
