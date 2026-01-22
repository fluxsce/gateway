/**
 * 预警日志服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
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
 * 预警日志服务层 Hook
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAlertLogService(
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 使用 model
  const model = useAlertLogModel()

  /**
   * 加载日志列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadLogList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 处理时间范围字段（从 datetimerange 转换为 startTime 和 endTime）
      const processedParams: Record<string, any> = {}
      if (finalSearchParams) {
        Object.keys(finalSearchParams).forEach(key => {
          if (key === 'timeRange' && Array.isArray(finalSearchParams[key]) && finalSearchParams[key].length === 2) {
            // 转换时间范围：将时间戳数组转换为 ISO 格式字符串
            processedParams.startTime = new Date(finalSearchParams[key][0]).toISOString()
            processedParams.endTime = new Date(finalSearchParams[key][1]).toISOString()
          } else if (finalSearchParams[key] !== '' && finalSearchParams[key] !== null && finalSearchParams[key] !== undefined) {
            // 过滤掉空字符串、null 和 undefined 的查询条件
            processedParams[key] = finalSearchParams[key]
          }
        })
      }

      // 构建请求参数：合并查询条件和分页参数
      const queryParams: any = {
        // 查询条件
        ...processedParams,
        // 分页参数
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize
        ),
      }

      const response = await queryAlertLogs(queryParams)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const logs = parseJsonData<AlertLog[]>(response, []) || []
        model.setLogList(logs)

        // 解析分页信息
        const pageInfo = parsePageInfo(response)
        if (pageInfo) {
          model.updatePagination(pageInfo)
        }
      } else {
        message.error(getApiMessage(response, '查询预警日志失败'))
        model.setLogList([])
      }
    } catch (error: any) {
      console.error('查询预警日志失败:', error)
      message.error(error.message || '查询预警日志失败')
      model.setLogList([])
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取日志详情
   * @param alertLogId 日志ID
   */
  const getLogDetail = async (alertLogId: string): Promise<AlertLog | null> => {
    try {
      model.setLoading(true)
      const response = await getAlertLog(alertLogId)
      if (isApiSuccess(response)) {
        return parseJsonData<AlertLog>(response)
      } else {
        message.error(getApiMessage(response, '获取预警日志详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取预警日志详情失败:', error)
      message.error(error.message || '获取预警日志详情失败')
      return null
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 删除日志
   * @param alertLogId 日志ID
   */
  const deleteLog = async (alertLogId: string): Promise<boolean> => {
    try {
      model.setLoading(true)
      const response = await deleteAlertLog(alertLogId)
      if (isApiSuccess(response)) {
        message.success('删除预警日志成功')
        await loadLogList() // 刷新列表
        return true
      } else {
        message.error(getApiMessage(response, '删除预警日志失败'))
        return false
      }
    } catch (error: any) {
      console.error('删除预警日志失败:', error)
      message.error(error.message || '删除预警日志失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 批量删除日志
   * @param alertLogIds 日志ID数组
   */
  const batchDeleteLogs = async (alertLogIds: string[]): Promise<boolean> => {
    if (alertLogIds.length === 0) {
      message.warning('请选择要删除的日志')
      return false
    }

    try {
      model.setLoading(true)
      const response = await batchDeleteAlertLogs(alertLogIds)
      if (isApiSuccess(response)) {
        message.success(`成功删除 ${alertLogIds.length} 条预警日志`)
        await loadLogList() // 刷新列表
        return true
      } else {
        message.error(getApiMessage(response, '批量删除预警日志失败'))
        return false
      }
    } catch (error: any) {
      console.error('批量删除预警日志失败:', error)
      message.error(error.message || '批量删除预警日志失败')
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

