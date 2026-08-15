/**
 * 预警日志服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { Ref } from 'vue'
import {
  batchDeleteAlertLogs,
  deleteAlertLog,
  getAlertLog,
  queryAlertLogs,
} from '../api'
import type { AlertLog } from '../types'
import { useAlertLogModel } from './model'

/**
 * 从 RsDatePicker range（valueFormat=string）取出起止时间并转为接口格式。
 * 使用本地墙钟时间，不转 UTC。告警时间按本地 DATETIME 落库，转成 ISO
 * 后东八区当天 16:00 之后的记录会被当天查询滤掉。
 * @param timeRange - 表单 timeRange 字段
 * @returns 本地起止时间；无效时返回空对象
 */
function resolveTimeRangeBounds(timeRange: unknown): { start?: string; end?: string } {
  if (!timeRange || typeof timeRange !== 'object' || Array.isArray(timeRange)) return {}
  const { start, end } = timeRange as { start?: string; end?: string }
  if (!start || !end) return {}
  return {
    start: formatDate(start, 'YYYY-MM-DDTHH:mm:ss'),
    end: formatDate(end, 'YYYY-MM-DDTHH:mm:ss'),
  }
}

/**
 * 预警日志服务层 Hook
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAlertLogService(
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0082')
  const model = useAlertLogModel()

  /**
   * 加载日志列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
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
          } else if (finalSearchParams[key] !== '' && finalSearchParams[key] !== null && finalSearchParams[key] !== undefined) {
            processedParams[key] = finalSearchParams[key]
          }
        })
      }

      const queryParams: any = {
        ...processedParams,
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize,
        ),
      }

      const response = await queryAlertLogs(queryParams)

      if (isApiSuccess(response)) {
        const logs = parseJsonData<AlertLog[]>(response, []) || []
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
      console.error('查询预警日志失败:', error)
      message.error(error.message || t('message.queryFailed'))
      model.setLogList([])
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取日志详情
   * @param alertLogId 日志ID
   * @returns 日志详情，失败时返回 null
   */
  const getLogDetail = async (alertLogId: string): Promise<AlertLog | null> => {
    try {
      model.setLoading(true)
      const response = await getAlertLog(alertLogId)
      if (isApiSuccess(response)) {
        return parseJsonData<AlertLog>(response)
      } else {
        message.error(getApiMessage(response, t('message.detailFailed')))
        return null
      }
    } catch (error: any) {
      console.error('获取预警日志详情失败:', error)
      message.error(error.message || t('message.detailFailed'))
      return null
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除日志
   * @param alertLogId 日志ID
   * @returns 是否删除成功
   */
  const deleteLog = async (alertLogId: string): Promise<boolean> => {
    try {
      model.setLoading(true)
      const response = await deleteAlertLog(alertLogId)
      if (isApiSuccess(response)) {
        message.success(t('message.deleteSuccess'))
        await loadLogList()
        return true
      } else {
        message.error(getApiMessage(response, t('message.deleteFailed')))
        return false
      }
    } catch (error: any) {
      console.error('删除预警日志失败:', error)
      message.error(error.message || t('message.deleteFailed'))
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除日志
   * @param alertLogIds 日志ID数组
   * @returns 是否删除成功
   */
  const batchDeleteLogs = async (alertLogIds: string[]): Promise<boolean> => {
    if (alertLogIds.length === 0) {
      message.warning(t('message.selectToDelete'))
      return false
    }

    try {
      model.setLoading(true)
      const response = await batchDeleteAlertLogs(alertLogIds)
      if (isApiSuccess(response)) {
        message.success(t('message.batchDeleteSuccess', { count: alertLogIds.length }))
        await loadLogList()
        return true
      } else {
        message.error(getApiMessage(response, t('message.batchDeleteFailed')))
        return false
      }
    } catch (error: any) {
      console.error('批量删除预警日志失败:', error)
      message.error(error.message || t('message.batchDeleteFailed'))
      return false
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadLogList,
    getLogDetail,
    deleteLog,
    batchDeleteLogs,
  }
}
