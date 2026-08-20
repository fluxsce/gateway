/**
 * 审计日志模块 API
 *
 * API路径: /gateway/hub0004
 * - POST /queryAuditLogs - 查询审计日志列表
 * - POST /getAuditLog - 获取审计日志详情
 * - POST /exportAuditLogs - 按筛选条件导出（最多 10000 条）
 */

import { createApi } from '@/api/request'
import { moduleApiPrefix } from '@/api/requestPath'
import type { JsonDataObj } from '@/types/api'
import type { AuthAuditLogQueryParams } from '../types'

const auditLogApi = createApi(moduleApiPrefix('hub0004'))

/**
 * 查询审计日志列表。
 * @param params - 分页与筛选条件
 */
export const queryAuditLogs = async (params: AuthAuditLogQueryParams): Promise<JsonDataObj> => {
  return auditLogApi.post('/queryAuditLogs', params)
}

/**
 * 获取审计日志详情。
 * @param auditId - 审计记录 ID
 */
export const getAuditLog = async (auditId: string): Promise<JsonDataObj> => {
  return auditLogApi.post('/getAuditLog', { auditId })
}

/**
 * 按筛选条件导出审计日志（最多 10000 条）。
 * @param params - 与列表相同的筛选条件，无需分页
 */
export const exportAuditLogs = async (
  params: Omit<AuthAuditLogQueryParams, 'pageIndex' | 'pageSize'>,
): Promise<JsonDataObj> => {
  return auditLogApi.post('/exportAuditLogs', params)
}
