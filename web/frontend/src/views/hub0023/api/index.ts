/**
 * Hub0023 网关日志管理模块 - API接口层
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { GatewayMonitoringQueryParams } from '../components/monitor/hooks/types'
import type {
  GatewayLogGetParams,
  GatewayLogQueryParams,
  GatewayLogResetParams,
} from '../types'

const gatewayLogApi = createApi('/gateway/hub0023')

/**
 * 网关日志查询API - 列表查询，不返回大字段以提高性能
 * @param params 查询参数
 * @returns 网关日志列表
 */
export const queryGatewayLogs = async (params: GatewayLogQueryParams): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/query', params)
}

/**
 * 网关日志详情获取API - 返回完整字段信息，包括大字段
 * @param params 获取参数
 * @returns 网关日志详情
 */
export const getGatewayLog = async (params: GatewayLogGetParams): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/get', params)
}

/**
 * 网关日志重置API（支持批量）
 * @param params 重置参数
 * @returns 重置结果
 */
export const resetGatewayLogs = async (params: GatewayLogResetParams): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/reset', params)
}

/**
 * 网关日志导出API
 * @param params 查询参数
 * @returns 导出结果
 */
export const exportGatewayLogs = async (params: GatewayLogQueryParams): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/export', params)
}

/**
 * 网关日志清理API - 根据条件清理历史日志
 * @param params 清理参数
 * @returns 清理结果
 */
export const cleanupGatewayLogs = async (params: {
  beforeTime: string
  keepDays?: number
}): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/cleanup', params)
}

/**
 * 获取网关监控概览数据API
 * @param params 查询参数
 * @returns 监控概览数据
 */
export const getGatewayMonitoringOverview = async (
  params: GatewayMonitoringQueryParams,
): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/monitoring/overview', params)
}

/**
 * 获取网关监控图表数据API
 * @param params 查询参数
 * @returns 监控图表数据
 */
export const getGatewayMonitoringChartData = async (
  params: GatewayMonitoringQueryParams,
): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/monitoring/chart-data', params)
}

/**
 * 获取网关实时监控数据API
 * @param params 查询参数
 * @returns 实时监控数据
 */
export const getGatewayRealtimeMonitoringData = async (
  params: GatewayMonitoringQueryParams,
): Promise<JsonDataObj> => {
  return gatewayLogApi.post('/gateway-log/monitoring/realtime', params)
}
