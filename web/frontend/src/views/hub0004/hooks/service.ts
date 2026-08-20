/**
 * 审计日志服务层 Hook
 * 纯业务逻辑：列表查询与详情获取
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

      const processedParams: Record<string, any> = {}
      if (finalSearchParams) {
        Object.keys(finalSearchParams).forEach((key) => {
          if (key === 'timeRange') {
            const bounds = resolveTimeRangeBounds(finalSearchParams[key])
            if (bounds.start && bounds.end) {
              processedParams.startTime = bounds.start
              processedParams.endTime = bounds.end
            }
          } else if (
            finalSearchParams[key] !== '' &&
            finalSearchParams[key] !== null &&
            finalSearchParams[key] !== undefined
          ) {
            processedParams[key] = finalSearchParams[key]
          }
        })
      }

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

  return {
    model,
    loadLogList,
    getLogDetail,
  }
}
