/**
 * 审计日志服务层 Hook
 * 纯业务逻辑：列表查询、详情获取与 CSV 导出
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { Ref } from 'vue'
import { exportAuditLogs, getAuditLog, queryAuditLogs } from '../api'
import type { AuthAuditLog } from '../types'
import { useAuditLogModel } from './model'

const MAX_EXPORT_ROWS = 10000

/**
 * 从 RsDatePicker range（valueFormat=string）取出起止时间。
 * 使用本地墙钟时间，不转 UTC。
 * @param timeRange - 表单 timeRange 字段
 */
function resolveTimeRangeBounds(timeRange: unknown): { start?: string; end?: string } {
  if (!timeRange || typeof timeRange !== 'object' || Array.isArray(timeRange)) return {}
  const { start, end } = timeRange as { start?: string; end?: string }
  if (!start || !end) return {}
  return {
    start: formatDate(start, 'YYYY-MM-DD HH:mm:ss'),
    end: formatDate(end, 'YYYY-MM-DD HH:mm:ss'),
  }
}

/**
 * 将搜索表单转为后端筛选条件（不含分页）。
 * @param searchParams - 查询条件
 */
function buildFilterParams(searchParams?: Record<string, any>): Record<string, any> {
  const processedParams: Record<string, any> = {}
  if (!searchParams) return processedParams
  Object.keys(searchParams).forEach((key) => {
    if (key === 'timeRange') {
      const bounds = resolveTimeRangeBounds(searchParams[key])
      if (bounds.start && bounds.end) {
        processedParams.startTime = bounds.start
        processedParams.endTime = bounds.end
      }
    } else if (
      searchParams[key] !== '' &&
      searchParams[key] !== null &&
      searchParams[key] !== undefined
    ) {
      processedParams[key] = searchParams[key]
    }
  })
  return processedParams
}

function csvEscape(value: unknown): string {
  let text = ''
  if (value == null) {
    text = ''
  } else if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
    text = String(value)
  } else {
    text = JSON.stringify(value)
  }
  if (/[",\n\r]/.test(text)) {
    return `"${text.replaceAll('"', '""')}"`
  }
  return text
}

function downloadCsv(filename: string, headers: string[], rows: string[][]) {
  const lines = [headers, ...rows].map((row) => row.map(csvEscape).join(','))
  const blob = new Blob(['\uFEFF' + lines.join('\n')], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  document.body.appendChild(link)
  link.click()
  link.remove()
  URL.revokeObjectURL(url)
}

/**
 * 审计日志服务层 Hook
 * @param searchFormRef - 搜索表单引用（可选）
 */
export function useAuditLogService(searchFormRef?: Ref<RsSearchFormExpose | null>) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0004')
  const model = useAuditLogModel()

  /**
   * 加载审计日志列表。
   * @param searchParams - 查询条件；不传则从表单读取
   */
  const loadLogList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const processedParams = buildFilterParams(finalSearchParams)

      const queryParams = {
        ...processedParams,
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize,
        ),
      }

      const response = await queryAuditLogs(queryParams)

      if (isApiSuccess(response)) {
        const logs = parseJsonData<AuthAuditLog[]>(response, []) || []
        model.setLogList(logs)

        const pageInfo = parsePageInfo(response)
        if (pageInfo) {
          model.updatePagination(pageInfo)
        }
      } else {
        message.error(getApiMessage(response, t('message.queryFailed')))
        model.setLogList([])
      }
    } catch (error: any) {
      message.error(error.message || t('message.queryFailed'))
      model.setLogList([])
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取审计日志详情。
   * @param auditId - 审计记录 ID
   */
  const getLogDetail = async (auditId: string): Promise<AuthAuditLog | null> => {
    try {
      const response = await getAuditLog(auditId)
      if (isApiSuccess(response)) {
        return parseJsonData<AuthAuditLog>(response)
      }
      message.error(getApiMessage(response, t('message.detailFailed')))
      return null
    } catch (error: any) {
      message.error(error.message || t('message.detailFailed'))
      return null
    }
  }

  /**
   * 按当前筛选条件导出 CSV。
   * @param searchParams - 查询条件；不传则从表单读取
   */
  const exportLogList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const response = await exportAuditLogs(buildFilterParams(finalSearchParams))
      if (!isApiSuccess(response)) {
        message.error(getApiMessage(response, t('message.exportFailed')))
        return
      }

      const logs = parseJsonData<AuthAuditLog[]>(response, []) || []
      if (logs.length === 0) {
        message.warning(t('message.exportEmpty'))
        return
      }

      const pageInfo = parsePageInfo(response)
      const total = pageInfo?.totalCount || logs.length
      if (total > logs.length) {
        message.warning(
          t('message.exportTruncated', { total, count: MAX_EXPORT_ROWS }),
        )
      }

      const headers = [
        t('columns.addTime'),
        t('columns.userName'),
        t('columns.userId'),
        t('columns.action'),
        t('columns.moduleCode'),
        t('columns.targetType'),
        t('columns.targetName'),
        t('columns.targetId'),
        t('columns.resourceCode'),
        t('columns.result'),
        t('columns.clientIP'),
        t('columns.requestMethod'),
        t('columns.requestPath'),
        t('columns.detail'),
      ]
      const rows = logs.map((row) => [
        row.addTime ? formatDate(row.addTime) : '',
        row.userName || '',
        row.userId || '',
        model.getActionLabel(row.action),
        model.getModuleLabel(row.moduleCode),
        model.getTargetTypeLabel(row.targetType),
        row.targetName || '',
        row.targetId || '',
        row.resourceCode || '',
        model.getResultLabel(row.result),
        row.clientIP || '',
        row.requestMethod || '',
        row.requestPath || '',
        row.detail || '',
      ])
      const stamp = formatDate(Date.now(), 'YYYYMMDDHHmmss')
      downloadCsv(`audit-log-${stamp}.csv`, headers, rows)
      message.success(t('message.exportSuccess', { count: logs.length }))
    } catch (error: any) {
      message.error(error.message || t('message.exportFailed'))
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadLogList,
    getLogDetail,
    exportLogList,
  }
}
