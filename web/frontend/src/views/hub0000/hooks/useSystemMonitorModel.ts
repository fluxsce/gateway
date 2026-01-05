/**
 * 系统监控数据模型管理
 * 重新架构：明确区分静态服务器信息和时间序列监控数据
 *
 * 数据架构说明：
 * - ServerInfo: 静态服务器信息，每个服务器部署时唯一
 * - 其他所有指标: 按采集时间存放的时间序列数据
 */

import { ref, computed, reactive } from 'vue'
import { createLogger } from '@/utils/logger'
import { usePagination } from '@/hooks/usePagination'
import type {
  ServerInfo,
  CPUMetrics,
  MemoryMetrics,
  DiskPartition,
  NetworkInterface,
  ProcessInfo,
  TemperatureInfo,
  MetricQueryParams,
  MetricTrend,
  DiskIOStats,
} from '../types'
import { ServerStatus } from '../types'

// 创建日志记录器
const logger = createLogger('SystemMonitorModel')

/**
 * 系统监控数据模型
 */
export const useSystemMonitorModel = () => {
  // ===================================================================
  // 静态服务器信息 (ServerInfo - 每个服务器部署时唯一)
  // ===================================================================

  // 服务器基本信息列表
  const serverList = ref<ServerInfo[]>([])
  const selectedServerIds = ref<string[]>([])
  const serverListLoading = ref(false)
  const currentServer = ref<ServerInfo | null>(null)

  // ===================================================================
  // 时间序列监控数据 (按采集时间存放)
  // ===================================================================

  // CPU 监控数据
  const cpuMetrics = ref<CPUMetrics[]>([])
  const cpuTrendData = ref<MetricTrend[]>([])
  const cpuLoading = ref(false)

  // 内存监控数据
  const memoryMetrics = ref<MemoryMetrics[]>([])
  const memoryTrendData = ref<MetricTrend[]>([])
  const memoryLoading = ref(false)

  // 磁盘监控数据
  const diskMetrics = ref<DiskPartition[]>([])
  const diskTrendData = ref<MetricTrend[]>([])
  const diskLoading = ref(false)

  // 磁盘IO监控数据
  const diskIOMetrics = ref<DiskIOStats[]>([])
  const diskIOLoading = ref(false)

  // 网络监控数据
  const networkMetrics = ref<NetworkInterface[]>([])
  const networkTrendData = ref<{ time: string; upload: number; download: number }[]>([])
  const networkLoading = ref(false)

  // 进程监控数据
  const processMetrics = ref<ProcessInfo[]>([])
  const processLoading = ref(false)

  // 温度监控数据
  const temperatureMetrics = ref<TemperatureInfo[]>([])
  const temperatureLoading = ref(false)

  // ===================================================================
  // 查询参数和状态
  // ===================================================================

  // 查询参数
  const queryParams = reactive<MetricQueryParams>({
    pageNum: 1,
    pageSize: 200,
    activeFlag: 'Y',
  })

  // 时间范围选择（已移除固定时间范围，现在使用具体的开始时间和结束时间）

  // 分页管理
  const pagination = usePagination()

  // ===================================================================
  // 计算属性 - 实时状态统计
  // ===================================================================

  // 在线服务器数量（基于最新监控数据）
  const onlineServerCount = computed(() => {
    const now = new Date()
    const onlineThreshold = 5 * 60 * 1000 // 5分钟

    return serverList.value.filter((server) => {
      // 检查各项指标的最新采集时间
      const latestCpu = cpuMetrics.value
        .filter((m) => m.metricServerId === server.metricServerId)
        .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

      if (!latestCpu) return false

      const lastUpdate = new Date(latestCpu.collectTime)
      return now.getTime() - lastUpdate.getTime() <= onlineThreshold
    }).length
  })

  // 离线服务器数量
  const offlineServerCount = computed(() => {
    return serverList.value.length - onlineServerCount.value
  })

  // 当前CPU平均使用率
  const currentCpuUsage = computed(() => {
    if (cpuMetrics.value.length === 0) return 0

    // 获取每个服务器的最新CPU数据
    const latestCpuData = new Map<string, CPUMetrics>()

    cpuMetrics.value.forEach((metric) => {
      const existing = latestCpuData.get(metric.metricServerId)
      if (!existing || new Date(metric.collectTime) > new Date(existing.collectTime)) {
        latestCpuData.set(metric.metricServerId, metric)
      }
    })

    const totalUsage = Array.from(latestCpuData.values()).reduce(
      (sum, metric) => sum + metric.usagePercent,
      0,
    )
    return Math.round((totalUsage / latestCpuData.size) * 100) / 100
  })

  // 当前内存平均使用率
  const currentMemoryUsage = computed(() => {
    if (memoryMetrics.value.length === 0) return 0

    // 获取每个服务器的最新内存数据
    const latestMemoryData = new Map<string, MemoryMetrics>()

    memoryMetrics.value.forEach((metric) => {
      const existing = latestMemoryData.get(metric.metricServerId)
      if (!existing || new Date(metric.collectTime) > new Date(existing.collectTime)) {
        latestMemoryData.set(metric.metricServerId, metric)
      }
    })

    const totalUsage = Array.from(latestMemoryData.values()).reduce(
      (sum, metric) => sum + metric.usagePercent,
      0,
    )
    return Math.round((totalUsage / latestMemoryData.size) * 100) / 100
  })

  // 当前磁盘平均使用率
  const currentDiskUsage = computed(() => {
    if (diskMetrics.value.length === 0) return 0

    // 获取每个服务器的最新磁盘数据
    const latestDiskData = new Map<string, DiskPartition>()

    diskMetrics.value.forEach((metric) => {
      const existing = latestDiskData.get(metric.metricServerId)
      if (!existing || new Date(metric.collectTime) > new Date(existing.collectTime)) {
        latestDiskData.set(metric.metricServerId, metric)
      }
    })

    const totalUsage = Array.from(latestDiskData.values()).reduce(
      (sum, metric) => sum + metric.usagePercent,
      0,
    )
    return Math.round((totalUsage / latestDiskData.size) * 100) / 100
  })

  // 服务器状态分布
  const serverStatusDistribution = computed(() => {
    const distribution = { online: 0, offline: 0, warning: 0, critical: 0 }
    const now = new Date()
    const offlineThreshold = 5 * 60 * 1000 // 5分钟
    const warningThreshold = 10 * 60 * 1000 // 10分钟

    serverList.value.forEach((server) => {
      // 获取服务器最新的监控数据
      const latestCpu = cpuMetrics.value
        .filter((m) => m.metricServerId === server.metricServerId)
        .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

      const latestMemory = memoryMetrics.value
        .filter((m) => m.metricServerId === server.metricServerId)
        .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

      if (!latestCpu || !latestMemory) {
        distribution.offline++
        return
      }

      const lastUpdate = new Date(latestCpu.collectTime)
      const timeDiff = now.getTime() - lastUpdate.getTime()

      if (timeDiff > warningThreshold) {
        distribution.offline++
      } else if (timeDiff > offlineThreshold) {
        distribution.warning++
      } else if (latestCpu.usagePercent > 90 || latestMemory.usagePercent > 90) {
        distribution.critical++
      } else {
        distribution.online++
      }
    })

    return distribution
  })

  // ===================================================================
  // 数据操作方法 - 静态数据
  // ===================================================================

  /**
   * 设置服务器列表数据
   */
  const setServerList = (data: ServerInfo[]) => {
    serverList.value = data
    logger.info('设置服务器列表数据', { count: data.length })
  }

  /**
   * 设置当前选中服务器
   */
  const setCurrentServer = (server: ServerInfo | null) => {
    currentServer.value = server
    logger.info('设置当前服务器', { serverId: server?.metricServerId })
  }

  /**
   * 添加或更新服务器数据
   */
  const upsertServerData = (server: ServerInfo) => {
    const index = serverList.value.findIndex((s) => s.metricServerId === server.metricServerId)
    if (index >= 0) {
      serverList.value[index] = server
    } else {
      serverList.value.push(server)
    }
    logger.info('更新服务器数据', { serverId: server.metricServerId })
  }

  /**
   * 删除服务器数据
   */
  const removeServerData = (serverId: string) => {
    const index = serverList.value.findIndex((s) => s.metricServerId === serverId)
    if (index >= 0) {
      serverList.value.splice(index, 1)
    }
    logger.info('删除服务器数据', { serverId })
  }

  // ===================================================================
  // 数据操作方法 - 时间序列数据
  // ===================================================================

  /**
   * 设置CPU监控数据
   */
  const setCpuMetrics = (data: CPUMetrics[]) => {
    cpuMetrics.value = data
    logger.info('设置CPU监控数据', { count: data.length })
  }

  /**
   * 设置CPU趋势数据
   */
  const setCpuTrendData = (data: MetricTrend[]) => {
    cpuTrendData.value = data
    logger.info('设置CPU趋势数据', { count: data.length })
  }

  /**
   * 设置内存监控数据
   */
  const setMemoryMetrics = (data: MemoryMetrics[]) => {
    memoryMetrics.value = data
    logger.info('设置内存监控数据', { count: data.length })
  }

  /**
   * 设置内存趋势数据
   */
  const setMemoryTrendData = (data: MetricTrend[]) => {
    memoryTrendData.value = data
    logger.info('设置内存趋势数据', { count: data.length })
  }

  /**
   * 设置磁盘监控数据
   */
  const setDiskMetrics = (data: DiskPartition[]) => {
    diskMetrics.value = data
    logger.info('设置磁盘监控数据', { count: data.length })
  }

  /**
   * 设置磁盘趋势数据
   */
  const setDiskTrendData = (data: MetricTrend[]) => {
    diskTrendData.value = data
    logger.info('设置磁盘趋势数据', { count: data.length })
  }

  /**
   * 设置磁盘IO指标数据
   */
  const setDiskIOMetrics = (data: DiskIOStats[]) => {
    diskIOMetrics.value = data
    logger.info('设置磁盘IO指标数据', { count: data.length })
  }

  /**
   * 设置网络监控数据
   */
  const setNetworkMetrics = (data: NetworkInterface[]) => {
    networkMetrics.value = data
    logger.info('设置网络监控数据', { count: data.length })
  }

  /**
   * 设置网络趋势数据
   */
  const setNetworkTrendData = (data: { time: string; upload: number; download: number }[]) => {
    networkTrendData.value = data
    logger.info('设置网络趋势数据', { count: data.length })
  }

  /**
   * 设置进程监控数据
   */
  const setProcessMetrics = (data: ProcessInfo[]) => {
    processMetrics.value = data
    logger.info('设置进程监控数据', { count: data.length })
  }

  /**
   * 设置温度监控数据
   */
  const setTemperatureMetrics = (data: TemperatureInfo[]) => {
    temperatureMetrics.value = data
    logger.info('设置温度监控数据', { count: data.length })
  }

  /**
   * 清空所有监控数据
   */
  const clearAllMetrics = () => {
    cpuMetrics.value = []
    cpuTrendData.value = []
    memoryMetrics.value = []
    memoryTrendData.value = []
    diskMetrics.value = []
    diskTrendData.value = []
    diskIOMetrics.value = []
    networkMetrics.value = []
    networkTrendData.value = []
    processMetrics.value = []
    temperatureMetrics.value = []
    logger.info('已清空所有监控数据')
  }

  // ===================================================================
  // 查询参数管理
  // ===================================================================

  /**
   * 更新查询参数
   */
  const updateQueryParams = (params: Partial<MetricQueryParams>) => {
    Object.assign(queryParams, params)
    logger.info('更新查询参数', params)
  }

  /**
   * 重置查询参数
   */
  const resetQueryParams = () => {
    Object.assign(queryParams, {
      pageNum: 1,
      pageSize: 20,
      activeFlag: 'Y',
    })
    logger.info('重置查询参数')
  }

  // 移除了setTimeRange方法，现在使用具体的开始时间和结束时间

  // ===================================================================
  // 工具方法
  // ===================================================================

  /**
   * 获取服务器状态
   */
  const getServerStatus = (serverId: string): ServerStatus => {
    const now = new Date()
    const offlineThreshold = 5 * 60 * 1000 // 5分钟

    // 获取服务器最新的CPU数据
    const latestCpu = cpuMetrics.value
      .filter((m) => m.metricServerId === serverId)
      .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

    if (!latestCpu) return ServerStatus.OFFLINE

    const lastUpdate = new Date(latestCpu.collectTime)
    const timeDiff = now.getTime() - lastUpdate.getTime()

    if (timeDiff > offlineThreshold) return ServerStatus.OFFLINE
    if (latestCpu.usagePercent > 90) return ServerStatus.CRITICAL
    if (latestCpu.usagePercent > 80) return ServerStatus.WARNING
    return ServerStatus.ONLINE
  }

  /**
   * 获取服务器最新指标数据
   */
  const getServerLatestMetrics = (serverId: string) => {
    const latestCpu = cpuMetrics.value
      .filter((m) => m.metricServerId === serverId)
      .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

    const latestMemory = memoryMetrics.value
      .filter((m) => m.metricServerId === serverId)
      .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

    const latestDisk = diskMetrics.value
      .filter((m) => m.metricServerId === serverId)
      .sort((a, b) => new Date(b.collectTime).getTime() - new Date(a.collectTime).getTime())[0]

    return {
      cpu: latestCpu,
      memory: latestMemory,
      disk: latestDisk,
    }
  }

  return {
    // 静态数据
    serverList,
    selectedServerIds,
    serverListLoading,
    currentServer,

    // 时间序列数据
    cpuMetrics,
    cpuTrendData,
    cpuLoading,
    memoryMetrics,
    memoryTrendData,
    memoryLoading,
    diskMetrics,
    diskTrendData,
    diskLoading,
    diskIOMetrics,
    diskIOLoading,
    networkMetrics,
    networkTrendData,
    networkLoading,
    processMetrics,
    processLoading,
    temperatureMetrics,
    temperatureLoading,

    // 查询参数
    queryParams,
    pagination,

    // 计算属性
    onlineServerCount,
    offlineServerCount,
    currentCpuUsage,
    currentMemoryUsage,
    currentDiskUsage,
    serverStatusDistribution,

    // 静态数据方法
    setServerList,
    setCurrentServer,
    upsertServerData,
    removeServerData,

    // 时间序列数据方法
    setCpuMetrics,
    setCpuTrendData,
    setMemoryMetrics,
    setMemoryTrendData,
    setDiskMetrics,
    setDiskTrendData,
    setDiskIOMetrics,
    setNetworkMetrics,
    setNetworkTrendData,
    setProcessMetrics,
    setTemperatureMetrics,
    clearAllMetrics,

    // 查询参数方法
    updateQueryParams,
    resetQueryParams,

    // 工具方法
    getServerStatus,
    getServerLatestMetrics,
  }
}
