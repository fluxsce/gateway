/**
 * 预警(告警)日志管理模块API
 * 提供告警日志的查询、查看、删除、统计等功能
 * 
 * API路径: /gateway/hub0082
 * 
 * 预警日志管理API：
 * - POST /queryAlertLogs - 查询预警日志列表
 * - POST /getAlertLog - 获取预警日志详情
 * - POST /updateAlertLog - 更新预警日志（主要用于更新发送状态和结果）
 * - POST /deleteAlertLog - 删除预警日志
 * - POST /batchDeleteAlertLogs - 批量删除预警日志
 * - POST /getAlertLogStatistics - 获取预警日志统计信息
 * 
 * 注意：预警日志由系统自动创建，不提供手动创建接口
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
    AlertLog,
    AlertLogQueryParams,
} from '../types'

// 创建API实例
const alertLogApi = createApi('/gateway/hub0082')

// ==================== 预警日志管理API ====================

/**
 * 查询预警日志列表
 * @param params 查询参数
 * @returns 预警日志列表
 */
export const queryAlertLogs = async (params: AlertLogQueryParams): Promise<JsonDataObj> => {
  return alertLogApi.post('/queryAlertLogs', params)
}

/**
 * 获取预警日志详情
 * @param alertLogId 告警日志ID
 * @returns 预警日志详情
 */
export const getAlertLog = async (alertLogId: string): Promise<JsonDataObj> => {
  return alertLogApi.post('/getAlertLog', { alertLogId })
}

/**
 * 更新预警日志（主要用于更新发送状态和结果）
 * @param data 预警日志数据（包含alertLogId）
 * @returns 更新结果
 */
export const updateAlertLog = async (data: Partial<AlertLog> & { alertLogId: string }): Promise<JsonDataObj> => {
  return alertLogApi.post('/updateAlertLog', data)
}

/**
 * 删除预警日志
 * @param alertLogId 告警日志ID
 * @returns 删除结果
 */
export const deleteAlertLog = async (alertLogId: string): Promise<JsonDataObj> => {
  return alertLogApi.post('/deleteAlertLog', { alertLogId })
}

/**
 * 批量删除预警日志
 * @param alertLogIds 告警日志ID数组
 * @returns 删除结果
 */
export const batchDeleteAlertLogs = async (alertLogIds: string[]): Promise<JsonDataObj> => {
  return alertLogApi.post('/batchDeleteAlertLogs', { alertLogIds })
}

/**
 * 获取预警日志统计信息
 * @param startTime 开始时间（可选）
 * @param endTime 结束时间（可选）
 * @returns 统计信息
 */
export const getAlertLogStatistics = async (
  startTime?: string,
  endTime?: string
): Promise<JsonDataObj> => {
  const params: Record<string, any> = {}
  if (startTime) {
    params.startTime = startTime
  }
  if (endTime) {
    params.endTime = endTime
  }
  return alertLogApi.post('/getAlertLogStatistics', params)
}

