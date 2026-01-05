/**
 * Hub0000 系统指标采集模块API接口
 * 根据后端实际定义的路由进行调用
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { MetricQueryParams } from '../types'

// 创建API实例，使用正确的前缀
const hub0000Api = createApi('/gateway/hub0000')

// ===================================================================
// 服务器信息管理
// ===================================================================

/**
 * 查询服务器信息列表
 * @param params 查询参数
 * @returns 服务器信息列表
 */
export const queryServerList = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/server/query', params)
}

/**
 * 获取服务器信息详情
 * @param params 查询参数，包含服务器ID
 * @returns 服务器详情
 */
export const getServerDetail = (params: { metricServerId: string }): Promise<JsonDataObj> => {
  return hub0000Api.post('/server/detail', params)
}

// ===================================================================
// CPU 指标数据
// ===================================================================

/**
 * 查询CPU性能日志列表
 * @param params 查询参数
 * @returns CPU性能日志列表
 */
export const queryCPUHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/cpu/query', params)
}

// ===================================================================
// 内存指标数据
// ===================================================================

/**
 * 查询内存性能日志列表
 * @param params 查询参数
 * @returns 内存性能日志列表
 */
export const queryMemoryHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/memory/query', params)
}

// ===================================================================
// 磁盘指标数据
// ===================================================================

/**
 * 查询磁盘分区日志列表
 * @param params 查询参数
 * @returns 磁盘分区日志列表
 */
export const queryDiskHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/disk/partition/query', params)
}

/**
 * 查询磁盘IO日志列表
 * @param params 查询参数
 * @returns 磁盘IO日志列表
 */
export const queryDiskIOHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/disk/io/query', params)
}

// ===================================================================
// 网络指标数据
// ===================================================================

/**
 * 查询网络日志列表
 * @param params 查询参数
 * @returns 网络日志列表
 */
export const queryNetworkHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/network/query', params)
}

// ===================================================================
// 进程指标数据
// ===================================================================

/**
 * 查询进程日志列表
 * @param params 查询参数
 * @returns 进程日志列表
 */
export const queryProcessHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/process/query', params)
}

/**
 * 查询进程统计日志列表
 * @param params 查询参数
 * @returns 进程统计日志列表
 */
export const queryProcessStatsHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/process/stats/query', params)
}

// ===================================================================
// 温度指标数据
// ===================================================================

/**
 * 查询温度日志列表
 * @param params 查询参数
 * @returns 温度日志列表
 */
export const queryTemperatureHistory = (params: MetricQueryParams): Promise<JsonDataObj> => {
  return hub0000Api.post('/temperature/query', params)
}

// ===================================================================
// 导出功能（如果后端有提供的话，暂时保留接口定义）
// ===================================================================

/**
 * 导出监控数据
 * @param params 导出参数
 * @returns 导出结果
 */
export const exportMetricData = (
  params: MetricQueryParams & { format: 'csv' | 'xlsx' | 'json' },
): Promise<JsonDataObj> => {
  return hub0000Api.post('/export', params)
}
