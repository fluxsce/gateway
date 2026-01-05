/**
 * 系统监控业务逻辑管理
 * 负责处理系统监控相关的业务逻辑和API调用
 */

import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import { createLogger } from '@/utils/logger'
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

// 创建日志记录器
const logger = createLogger('SystemMonitorManagement')

/**
 * 系统监控业务逻辑管理
 */
export const useSystemMonitorManagement = (model: ReturnType<typeof useSystemMonitorModel>) => {
  // ===================================================================
  // 基础依赖
  // ===================================================================

  const message = useMessage()

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

      logger.info('开始加载服务器列表', queryParams)

      const response = await queryServerList(queryParams)

      if (isApiSuccess(response)) {
        const servers = parseJsonData<ServerInfo[]>(response, [])
        model.setServerList(servers)

        // 更新分页信息
        // TODO: 需要从response中解析分页信息并调用updatePagination

        message.success('服务器列表加载成功')
        logger.info('服务器列表加载成功', { count: servers.length })
        return servers
      } else {
        const errorMsg = getApiMessage(response, '加载服务器列表失败')
        message.error(errorMsg)
        logger.error('加载服务器列表失败', { error: errorMsg })
        return []
      }
    } catch (error) {
      const errorMsg = '加载服务器列表异常'
      message.error(errorMsg)
      logger.error(errorMsg, error)
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

      logger.info('开始获取服务器详情', { serverId })

      const response = await getServerDetail({ metricServerId: serverId })

      if (isApiSuccess(response)) {
        const server = parseJsonData<ServerInfo>(response)
        logger.info('获取服务器详情成功', { serverId })
        return server
      } else {
        const errorMsg = getApiMessage(response, '获取服务器详情失败')
        message.error(errorMsg)
        logger.error('获取服务器详情失败', { serverId, error: errorMsg })
        return null
      }
    } catch (error) {
      const errorMsg = '获取服务器详情异常'
      message.error(errorMsg)
      logger.error(errorMsg, { serverId, error })
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

      logger.info('开始加载CPU监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryCPUHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<CPUMetrics[]>(response, [])
        model.setCpuMetrics(metrics)

        logger.info('CPU监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载CPU监控数据失败')
        logger.error('加载CPU监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载CPU监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载内存监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryMemoryHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<MemoryMetrics[]>(response, [])
        model.setMemoryMetrics(metrics)

        logger.info('内存监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载内存监控数据失败')
        logger.error('加载内存监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载内存监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载磁盘监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryDiskHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<DiskPartition[]>(response, [])
        model.setDiskMetrics(metrics)

        logger.info('磁盘监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载磁盘监控数据失败')
        logger.error('加载磁盘监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载磁盘监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载网络监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryNetworkHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<NetworkInterface[]>(response, [])
        model.setNetworkMetrics(metrics)

        logger.info('网络监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载网络监控数据失败')
        logger.error('加载网络监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载网络监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载进程监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryProcessHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<ProcessInfo[]>(response, [])
        model.setProcessMetrics(metrics)

        logger.info('进程监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载进程监控数据失败')
        logger.error('加载进程监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载进程监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载温度监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryTemperatureHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<TemperatureInfo[]>(response, [])
        model.setTemperatureMetrics(metrics)

        logger.info('温度监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载温度监控数据失败')
        logger.error('加载温度监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载温度监控数据异常'
      logger.error(errorMsg, error)
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

      logger.info('开始加载磁盘IO监控数据', { serverId })

      const params: MetricQueryParams = {
        ...(serverId ? { metricServerId: serverId } : {}),
        ...model.queryParams,
      }

      const response = await queryDiskIOHistory(params)

      if (isApiSuccess(response)) {
        const metrics = parseJsonData<DiskIOStats[]>(response, [])
        model.setDiskIOMetrics(metrics)

        logger.info('磁盘IO监控数据加载成功', { count: metrics.length })
      } else {
        const errorMsg = getApiMessage(response, '加载磁盘IO监控数据失败')
        logger.error('加载磁盘IO监控数据失败', { error: errorMsg })
      }
    } catch (error) {
      const errorMsg = '加载磁盘IO监控数据异常'
      logger.error(errorMsg, error)
    } finally {
      model.diskIOLoading.value = false
    }
  }

  /**
   * 加载所有监控数据
   */
  const loadAllMetrics = async (serverId: string) => {
    if (!serverId) {
      logger.warn('未提供服务器ID，无法加载监控数据')
      return
    }

    try {
      operationLoading.value = true
      logger.info('开始加载所有监控数据', { serverId })

      await Promise.all([
        loadCPUMetrics(serverId),
        loadMemoryMetrics(serverId),
        loadDiskMetrics(serverId),
        loadDiskIOMetrics(serverId),
        loadNetworkMetrics(serverId),
        loadProcessMetrics(serverId),
        loadTemperatureMetrics(serverId),
      ])

      logger.info('所有监控数据加载成功')
    } catch (error) {
      const errorMsg = '加载所有监控数据异常'
      logger.error(errorMsg, error)
      throw error // 向上传播错误，让调用者处理
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

      logger.info('开始导出监控数据', params)

      const response = await exportMetricData(params)

      if (isApiSuccess(response)) {
        const downloadUrl = parseJsonData<string>(response)
        message.success('数据导出成功')
        logger.info('数据导出成功', { downloadUrl })
        return downloadUrl
      } else {
        const errorMsg = getApiMessage(response, '导出监控数据失败')
        message.error(errorMsg)
        logger.error('导出监控数据失败', { error: errorMsg })
        return null
      }
    } catch (error) {
      const errorMsg = '导出监控数据异常'
      message.error(errorMsg)
      logger.error(errorMsg, error)
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
