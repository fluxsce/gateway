import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { ServerInfoQuery } from '../types'

const serverNodeApi = createApi('/gateway/hub0007')

/**
 * 分页查询系统节点列表
 * @param params 查询参数
 * @returns 系统节点列表和分页信息
 */
export async function queryServerInfos(params: ServerInfoQuery): Promise<JsonDataObj> {
  return serverNodeApi.post('/queryServerInfos', params)
}

/**
 * 查询单个系统节点信息
 * @param metricServerId 节点ID
 * @returns 系统节点信息
 */
export async function getServerInfo(metricServerId: string): Promise<JsonDataObj> {
  return serverNodeApi.post('/getServerInfo', { metricServerId })
}

// ===================================================================
// 监控数据查询 API
// ===================================================================

/**
 * 监控数据查询参数
 */
export interface MetricQueryParams {
  metricServerId: string
  startTime?: string
  endTime?: string
}

/**
 * 查询CPU监控数据
 * @param params 查询参数
 * @returns CPU监控数据
 */
export async function queryCPUMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/cpu', params)
}

/**
 * 查询内存监控数据
 * @param params 查询参数
 * @returns 内存监控数据
 */
export async function queryMemoryMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/memory', params)
}

/**
 * 查询磁盘监控数据
 * @param params 查询参数
 * @returns 磁盘监控数据
 */
export async function queryDiskMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/disk', params)
}

/**
 * 查询磁盘IO监控数据
 * @param params 查询参数
 * @returns 磁盘IO监控数据
 */
export async function queryDiskIOMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/diskio', params)
}

/**
 * 查询网络监控数据
 * @param params 查询参数
 * @returns 网络监控数据
 */
export async function queryNetworkMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/network', params)
}

/**
 * 查询进程监控数据
 * @param params 查询参数
 * @returns 进程监控数据
 */
export async function queryProcessMetrics(params: MetricQueryParams): Promise<JsonDataObj> {
  return serverNodeApi.post('/metrics/process', params)
}

