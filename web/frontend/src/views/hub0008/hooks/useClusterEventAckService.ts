/**
 * 集群事件确认业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { createBackendPaginationParams } from '@/utils/pagination'
import type { JsonDataObj } from '@/types/api'
import type { Ref } from 'vue'
import { queryClusterEventAcks } from '../api'
import { useClusterEventAckModel } from './ackModel'

/**
 * 集群事件确认服务 Hook
 */
export function useClusterEventAckService(
  eventId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0008')

  // 初始化 Model
  const model = useClusterEventAckModel()

  const {
    loading,
    ackList,
    pageInfo,
    setAckList,
    updatePagination,
  } = model

  const loadAcks = async (searchParams?: Record<string, any>) => {
    const finalEventId = eventId?.value || searchParams?.eventId
    if (!finalEventId) {
      setAckList([])
      pageInfo.value = undefined
      return
    }

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

      const params: Record<string, any> = {
        eventId: finalEventId,
        ...effectiveSearchParams,
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      const response: JsonDataObj = await queryClusterEventAcks(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const acks = Array.isArray(bizData) ? bizData : []
          setAckList(acks)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || t('ack.message.queryFailed'))
      }
    } catch (error) {
      message.error(t('ack.message.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索集群事件确认
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    model.resetPagination()
    await loadAcks(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadAcks()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadAcks()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadAcks()
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,

    // 数据加载
    loadAcks,

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
export type ClusterEventAckService = ReturnType<typeof useClusterEventAckService>

