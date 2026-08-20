/**
 * 集群事件业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { JsonDataObj } from '@/types/api'
import type { Ref } from 'vue'
import { queryClusterEvents } from '../api'
import { useClusterEventModel } from './model'

/**
 * 集群事件服务 Hook
 */
export function useClusterEventService(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0008')

  // 初始化 Model
  const model = useClusterEventModel()

  const {
    loading,
    eventList,
    pageInfo,
    setEventList,
    updatePagination,
  } = model

  const loadEvents = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      const params = {
        ...effectiveSearchParams,
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      const response: JsonDataObj = await queryClusterEvents(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const events = Array.isArray(bizData) ? bizData : []
          setEventList(events)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || t('event.message.queryFailed'))
      }
    } catch (error) {
      message.error(t('event.message.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索集群事件
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    model.resetPagination()
    await loadEvents(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadEvents()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadEvents()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadEvents()
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,

    // 数据加载
    loadEvents,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
  }
}

/**
 * 服务返回类型
 */
export type ClusterEventService = ReturnType<typeof useClusterEventService>

