/**
 * 系统节点监控 Hook
 * 管理选中节点的监控数据加载和状态
 */

import { formatDate } from '@/utils/format'
import { ref } from 'vue'
import * as serverNodeApi from '../api'

/**
 * 系统节点监控 Hook
 */
export function useServerNodeMonitor() {
  // ============= 状态 =============

  /** 选中的服务器ID */
  const selectedServerId = ref<string>('')

  /** 监控数据加载状态 */
  const cpuLoading = ref(false)
  const memoryLoading = ref(false)
  const diskLoading = ref(false)
  const diskIOLoading = ref(false)
  const networkLoading = ref(false)
  const processLoading = ref(false)

  /** 监控数据 */
  const cpuMetrics = ref<any[]>([])
  const memoryMetrics = ref<any[]>([])
  const diskMetrics = ref<any[]>([])
  const diskIOMetrics = ref<any[]>([])
  const networkMetrics = ref<any[]>([])
  const processMetrics = ref<any[]>([])

  /** 时间范围（默认最近1小时） */
  const timeRange = ref<[number, number]>([Date.now() - 3600 * 1000, Date.now()])

  // ============= 数据加载方法 =============

  /**
   * 加载CPU监控数据
   */
  const loadCPUMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    cpuLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryCPUMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        cpuMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载CPU监控数据失败:', error)
    } finally {
      cpuLoading.value = false
    }
  }

  /**
   * 加载内存监控数据
   */
  const loadMemoryMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    memoryLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryMemoryMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        memoryMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载内存监控数据失败:', error)
    } finally {
      memoryLoading.value = false
    }
  }

  /**
   * 加载磁盘监控数据
   */
  const loadDiskMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    diskLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryDiskMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        diskMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载磁盘监控数据失败:', error)
    } finally {
      diskLoading.value = false
    }
  }

  /**
   * 加载磁盘IO监控数据
   */
  const loadDiskIOMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    diskIOLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryDiskIOMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        diskIOMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载磁盘IO监控数据失败:', error)
    } finally {
      diskIOLoading.value = false
    }
  }

  /**
   * 加载网络监控数据
   */
  const loadNetworkMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    networkLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryNetworkMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        networkMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载网络监控数据失败:', error)
    } finally {
      networkLoading.value = false
    }
  }

  /**
   * 加载进程监控数据
   */
  const loadProcessMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    processLoading.value = true
    try {
      const [startTime, endTime] = timeRange.value
      const response = await serverNodeApi.queryProcessMetrics({
        metricServerId: targetServerId,
        startTime: formatDate(startTime, 'YYYY-MM-DD HH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DD HH:mm:ss')
      })

      if (response.oK && response.bizData) {
        processMetrics.value = JSON.parse(response.bizData)
      }
    } catch (error) {
      console.error('加载进程监控数据失败:', error)
    } finally {
      processLoading.value = false
    }
  }

  /**
   * 加载所有监控数据
   */
  const loadAllMetrics = async (serverId?: string) => {
    const targetServerId = serverId || selectedServerId.value
    if (!targetServerId) return

    await Promise.all([
      loadCPUMetrics(targetServerId),
      loadMemoryMetrics(targetServerId),
      loadDiskMetrics(targetServerId),
      loadDiskIOMetrics(targetServerId),
      loadNetworkMetrics(targetServerId),
      loadProcessMetrics(targetServerId)
    ])
  }

  /**
   * 清空所有监控数据
   */
  const clearAllMetrics = () => {
    cpuMetrics.value = []
    memoryMetrics.value = []
    diskMetrics.value = []
    diskIOMetrics.value = []
    networkMetrics.value = []
    processMetrics.value = []
  }

  /**
   * 设置选中的服务器并加载监控数据
   */
  const setSelectedServer = async (serverId: string) => {
    if (selectedServerId.value === serverId) return

    selectedServerId.value = serverId
    clearAllMetrics()

    if (serverId) {
      await loadAllMetrics(serverId)
    }
  }

  /**
   * 更新时间范围
   */
  const updateTimeRange = (range: [number, number]) => {
    timeRange.value = range
  }

  return {
    // 状态
    selectedServerId,
    cpuLoading,
    memoryLoading,
    diskLoading,
    diskIOLoading,
    networkLoading,
    processLoading,
    cpuMetrics,
    memoryMetrics,
    diskMetrics,
    diskIOMetrics,
    networkMetrics,
    processMetrics,
    timeRange,

    // 方法
    loadCPUMetrics,
    loadMemoryMetrics,
    loadDiskMetrics,
    loadDiskIOMetrics,
    loadNetworkMetrics,
    loadProcessMetrics,
    loadAllMetrics,
    clearAllMetrics,
    setSelectedServer,
    updateTimeRange
  }
}

/**
 * ServerNodeMonitor 类型定义
 */
export type ServerNodeMonitor = ReturnType<typeof useServerNodeMonitor>

