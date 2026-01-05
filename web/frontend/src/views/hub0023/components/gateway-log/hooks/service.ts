/**
 * Hub0023 网关日志管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as gatewayLogApi from '../../../api'
import type { GatewayLogListItem, GatewayLogQueryParams, GatewayLogResetParams } from '../../../types'
import { useGatewayLogModel } from './model'

/**
 * 网关日志服务 Hook（纯业务逻辑）
 * @param searchFormRef 搜索表单引用
 */
export function useGatewayLogService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useGatewayLogModel()

  const {
    loading,
    logList,
    pageInfo,
    setLogList,
    updatePagination,
    resetPagination
  } = model

  // ============= 数据加载 =============

  /**
   * 加载网关日志列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadGatewayLogs = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 处理时间范围字段（从 datetimerange 转换为 startTime 和 endTime）
      const processedParams: Partial<GatewayLogQueryParams> = {}
      if (finalSearchParams) {
        Object.keys(finalSearchParams).forEach(key => {
          if (key === 'timeRange' && Array.isArray(finalSearchParams[key]) && finalSearchParams[key].length === 2) {
            // 转换时间范围
            processedParams.startTime = formatDate(finalSearchParams[key][0], 'YYYY-MM-DDTHH:mm:ss')
            processedParams.endTime = formatDate(finalSearchParams[key][1], 'YYYY-MM-DDTHH:mm:ss')
          } else if (finalSearchParams[key] !== '' && finalSearchParams[key] !== null && finalSearchParams[key] !== undefined) {
            // 过滤掉空字符串、null 和 undefined 的查询条件
            ;(processedParams as Record<string, any>)[key] = finalSearchParams[key]
          }
        })
      }

      // 构建请求参数：合并查询条件和分页参数
      const params: GatewayLogQueryParams = {
        // 查询条件
        ...processedParams,
        // 分页参数
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        ),
        // 排序参数
        sortField: 'gatewayStartProcessingTime',
        sortOrder: 'DESC',
      } as GatewayLogQueryParams

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await gatewayLogApi.queryGatewayLogs(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const data = parseJsonData<GatewayLogListItem[]>(response, [])
        setLogList(data)

        // 解析分页信息
        try {
          const pageInfoData = parsePageInfo(response)
          if (pageInfoData && Object.keys(pageInfoData).length > 0) {
            updatePagination(pageInfoData)
          }
        } catch (error) {
          console.warn('分页信息解析失败:', error)
        }
      } else {
        const errorMsg = getApiMessage(response, '查询网关日志列表失败')
        message.error(errorMsg)
        setLogList([])
      }
    } catch (error) {
      console.error('加载网关日志列表失败:', error)
      message.error('加载网关日志列表失败')
      setLogList([])
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索网关日志
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    resetPagination()
    await loadGatewayLogs(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    resetPagination()
    await loadGatewayLogs()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadGatewayLogs()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadGatewayLogs()
  }

  // ============= 日志操作 =============

  /**
   * 重置网关日志（支持单个和批量）
   */
  const resetGatewayLogs = async (
    logs: GatewayLogListItem[],
    resetReason: string = '手动重置',
    operatorId: string = 'current_user',
  ): Promise<boolean> => {
    try {
      if (logs.length === 0) {
        message.warning('请选择要重置的日志')
        return false
      }

      // 过滤出可以重置的日志
      const resettableLogs = logs.filter((log) => log.resetFlag === 'N')
      if (resettableLogs.length === 0) {
        message.warning('所选日志都已重置，无需再次重置')
        return false
      }

      const params: GatewayLogResetParams = {
        traceIds: resettableLogs.map((log) => log.traceId),
        resetReason,
        operatorId,
      }

      const response = await gatewayLogApi.resetGatewayLogs(params)

      if (isApiSuccess(response)) {
        const count = resettableLogs.length
        message.success(`成功重置 ${count} 条网关日志`)
        await loadGatewayLogs()
        return true
      } else {
        const errorMsg = getApiMessage(response, '重置网关日志失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      console.error('重置网关日志失败:', error)
      message.error('重置网关日志失败')
      return false
    }
  }

  /**
   * 导出网关日志
   */
  const exportGatewayLogs = async (searchParams?: Record<string, any>) => {
    try {
      // 获取当前查询条件
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 处理时间范围字段
      const processedParams: Partial<GatewayLogQueryParams> = {}
      if (finalSearchParams) {
        Object.keys(finalSearchParams).forEach(key => {
          if (key === 'timeRange' && Array.isArray(finalSearchParams[key]) && finalSearchParams[key].length === 2) {
            processedParams.startTime = formatDate(finalSearchParams[key][0], 'YYYY-MM-DDTHH:mm:ss')
            processedParams.endTime = formatDate(finalSearchParams[key][1], 'YYYY-MM-DDTHH:mm:ss')
          } else if (finalSearchParams[key] !== '' && finalSearchParams[key] !== null && finalSearchParams[key] !== undefined) {
            ;(processedParams as Record<string, any>)[key] = finalSearchParams[key]
          }
        })
      }

      const params: GatewayLogQueryParams = {
        ...processedParams,
        ...createBackendPaginationParams(1, 10000), // 导出时使用较大的分页
        sortField: 'gatewayStartProcessingTime',
        sortOrder: 'DESC',
      } as GatewayLogQueryParams

      const response = await gatewayLogApi.exportGatewayLogs(params)

      if (isApiSuccess(response)) {
        const downloadUrl = parseJsonData<string>(response)
        // 创建下载链接
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = `网关日志_${new Date().toISOString().slice(0, 19).replace(/[:-]/g, '')}.xlsx`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        message.success('网关日志导出成功')
      } else {
        const errorMsg = getApiMessage(response, '导出网关日志失败')
        message.error(errorMsg)
      }
    } catch (error) {
      console.error('导出网关日志失败:', error)
      message.error('导出网关日志失败')
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadGatewayLogs,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,

    // 日志操作
    resetGatewayLogs,
    exportGatewayLogs,
  }
}

/**
 * 网关日志服务类型
 */
export type GatewayLogService = ReturnType<typeof useGatewayLogService>

