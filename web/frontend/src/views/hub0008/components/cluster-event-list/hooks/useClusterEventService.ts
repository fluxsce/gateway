/**
 * 集群事件业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { queryClusterEvents } from '../../../api'
import { useClusterEventModel } from './model'

/**
 * 集群事件服务 Hook
 */
export function useClusterEventService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useClusterEventModel()

  const {
    loading,
    eventList,
    pageInfo,
    setEventList,
    updatePagination,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载集群事件列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadEvents = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryClusterEvents(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const events = Array.isArray(bizData) ? bizData : []
          setEventList(events)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询集群事件列表失败')
      }
    } catch (error) {
      console.error('加载集群事件列表失败:', error)
      message.error('加载集群事件列表失败')
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

