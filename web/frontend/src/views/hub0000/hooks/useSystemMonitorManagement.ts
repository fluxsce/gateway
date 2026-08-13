/**
 * 系统监控业务逻辑管理
 * 负责处理系统监控相关的业务逻辑和API调用
 */

import { ref } from 'vue'
import { useAppMessage } from '@/composables/useAppMessage'
import { parseJsonData, isApiSuccess, getApiMessage } from '@/utils/format'
import type { useSystemMonitorModel } from './useSystemMonitorModel'
import type {
  ServerInfo,
  CPUMetrics,
  MemoryMetrics,
  DiskPartition,
  DiskIOStats,
  NetworkInterface,
  ProcessInfo,
  TemperatureInfo,
  MetricQueryParams,
} from '../types'
import {
  queryServerList,
  getServerDetail,
  queryCPUHistory,
  queryMemoryHistory,
  queryDiskHistory,
  queryDiskIOHistory,
  queryNetworkHistory,
  queryProcessHistory,
  queryTemperatureHistory,
  exportMetricData,
} from '../api'

/**
 * 系统监控业务逻辑管理
 */
export const useSystemMonitorManagement = (model: ReturnType<typeof useSystemMonitorModel>) => {
  // ===================================================================
  // 基础依赖
  // ===================================================================

  const message = useAppMessage()

  // 操作状态
  const operationLoading = ref(false)

  // ===================================================================
  // 服务器管理
  // ===================================================================

  /**
   * 加载服务器列表
   */
  const loadServerList = async (params?: Partial<MetricQueryParams>) => {
    try {
      model.serverListLoading.value = true

      // 合并查询参数
      const queryParams = { ...model.queryParams, ...params }

      const response = await queryServerList(queryParams)

      if (isApiSuccess(response)) {
        const servers = parseJsonData<ServerInfo[]>(response, [])
        model.setServerList(servers)

        // 更新分页信息
        // TODO: 需要从response中解析分页信息并调用updatePagination

        message.success('服务器列表加载成功')
        return servers
      } else {
        message.error(getApiMessage(response, '加载服务器列表失败'))
        return []
      }
    } catch {
      message.error('加载服务器列表异常')
      return []
    } finally {
      model.serverListLoading.value = false
    }
  }

  /**
   * 获取服务器详情
   */
  const getServerInfo = async (serverId: string): Promise<ServerInfo | null> => {
    try {
      operationLoading.value = true

      const response = await getServerDetail({ metricServerId: serverId })

      if (isApiSuccess(response)) {
        return parseJsonData<ServerInfo>(response)
      } else {
        message.error(getApiMessage(response, '获取服务器详情失败'))
        return null
      }
    } catch {
      message.error('获取服务器详情异常')
      return null
    } finally {
      operationLoading.value = false
    }
  }

  /**
   * 获取默认选中的服务器ID
   */
  const getDefaultServerId = (): string => {
    return model.serverList.value.length > 0 ? model.serverList.value[0].metricServerId : ''
  }

  // ===================================================================
  // 监控数据管理
  // ===================================================================

  /**
   * 加载CPU监控数据
   */
  const loadCPUMetrics = async (serverId?: string) => {
    try {
      model.cpuLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryCPUHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<CPUMetrics[]>(response, [])
        model.setCpuMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.cpuLoading.value = false
    }
  }

  /**
   * 加载内存监控数据
   */
  const loadMemoryMetrics = async (serverId?: string) => {
    try {
      model.memoryLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryMemoryHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<MemoryMetrics[]>(response, [])
        model.setMemoryMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.memoryLoading.value = false
    }
  }

  /**
   * 加载磁盘监控数据
   */
  const loadDiskMetrics = async (serverId?: string) => {
    try {
      model.diskLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryDiskHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<DiskPartition[]>(response, [])
        model.setDiskMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.diskLoading.value = false
    }
  }

  /**
   * 加载网络监控数据
   */
  const loadNetworkMetrics = async (serverId?: string) => {
    try {
      model.networkLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryNetworkHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<NetworkInterface[]>(response, [])
        model.setNetworkMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.networkLoading.value = false
    }
  }

  /**
   * 加载进程监控数据
   */
  const loadProcessMetrics = async (serverId?: string) => {
    try {
      model.processLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryProcessHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<ProcessInfo[]>(response, [])
        model.setProcessMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.processLoading.value = false
    }
  }

  /**
   * 加载温度监控数据
   */
  const loadTemperatureMetrics = async (serverId?: string) => {
    try {
      model.temperatureLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryTemperatureHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<TemperatureInfo[]>(response, [])
        model.setTemperatureMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.temperatureLoading.value = false
    }
  }

  /**
   * 加载磁盘IO监控数据
   */
  const loadDiskIOMetrics = async (serverId?: string) => {
    try {
      model.diskIOLoading.value = true

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryDiskIOHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<DiskIOStats[]>(response, [])
        model.setDiskIOMetrics(metrics)
      }
    } catch {
      // 静默失败，由页面层处理提示
    } finally {
      model.diskIOLoading.value = false
    }
  }

  /**
   * 加载所有监控数据
   */
  const loadAllMetrics = async (serverId: string) => {
    if (!serverId) {
      return
    }

    try {
      operationLoading.value = true

      await Promise.all([
        loadCPUMetrics(serverId),
        loadMemoryMetrics(serverId),
        loadDiskMetrics(serverId),
        loadDiskIOMetrics(serverId),
        loadNetworkMetrics(serverId),
        loadProcessMetrics(serverId),
        loadTemperatureMetrics(serverId),
      ])
    } catch (error) {
      // 向上传播错误，让调用者处理
      throw error
    } finally {
      operationLoading.value = false
    }
  }

  // ===================================================================
  // 数据导出
  // ===================================================================

  /**
   * 导出监控数据
   */
  const exportData = async (
    params: MetricQueryParams & { format: 'csv' | 'xlsx' | 'json' },
  ): Promise<string | null> => {
    try {
      operationLoading.value = true

      const response = await exportMetricData(params)

      if (isApiSuccess(response)) {
        const downloadUrl = parseJsonData<string>(response)
        message.success('数据导出成功')
        return downloadUrl
      } else {
        message.error(getApiMessage(response, '导出监控数据失败'))
        return null
      }
    } catch {
      message.error('导出监控数据异常')
      return null
    } finally {
      operationLoading.value = false
    }
  }

  // ===================================================================
  // 返回接口
  // ===================================================================

  return {
    // 状态
    operationLoading,

    // 服务器管理
    loadServerList,
    getServerInfo,
    getDefaultServerId,

    // 监控数据
    loadCPUMetrics,
    loadMemoryMetrics,
    loadDiskMetrics,
    loadDiskIOMetrics,
    loadNetworkMetrics,
    loadProcessMetrics,
    loadTemperatureMetrics,
    loadAllMetrics,

    // 数据导出
    exportData,
  }
}
