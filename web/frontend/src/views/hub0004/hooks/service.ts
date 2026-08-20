/**
 * 审计日志服务层 Hook
 * 纯业务逻辑：列表查询、详情获取；导出由服务端出 CSV，本层只组筛选条件
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { Ref } from 'vue'
import { getAuditLog, queryAuditLogs } from '../api'
import type { AuthAuditLog } from '../types'
import { useAuditLogModel } from './model'

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
   * 当前筛选条件（不含分页），供服务端导出使用。
   */
  const buildExportParams = (): Record<string, any> => {
    let finalSearchParams: Record<string, any> | undefined
    if (searchFormRef?.value?.getFormData) {
      finalSearchParams = searchFormRef.value.getFormData() || {}
    }
    return buildFilterParams(finalSearchParams)
  }

  return {
    model,
    loadLogList,
    getLogDetail,
    buildExportParams,
  }
}
