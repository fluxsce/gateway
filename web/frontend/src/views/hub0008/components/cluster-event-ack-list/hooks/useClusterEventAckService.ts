/**
 * 集群事件确认业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { getClusterEventAckDetail, queryClusterEventAcks } from '../../../api'
import type { ClusterEventAck } from '../../../types'
import { useClusterEventAckModel } from './model'

/**
 * 集群事件确认服务 Hook
 */
export function useClusterEventAckService(
  eventId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 初始化 Model
  const model = useClusterEventAckModel()

  const {
    loading,
    ackList,
    pageInfo,
    setAckList,
    updatePagination,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载集群事件确认列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadAcks = async (searchParams?: Record<string, any>) => {
    const finalEventId = eventId?.value || searchParams?.eventId
    if (!finalEventId) {
      setAckList([])
      pageInfo.value = undefined
      return
    }

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
      const params: Record<string, any> = {
        eventId: finalEventId,
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryClusterEventAcks(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const acks = Array.isArray(bizData) ? bizData : []
          setAckList(acks)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询事件处理节点列表失败')
      }
    } catch (error) {
      console.error('加载事件处理节点列表失败:', error)
      message.error('加载事件处理节点列表失败')
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

  // ============= 详情获取 =============

  /**
   * 获取集群事件确认详情
   * @param ackId 确认ID
   * @returns 集群事件确认详情
   */
  const getAckDetail = async (ackId: string): Promise<ClusterEventAck | null> => {
    if (!ackId) {
      message.warning('确认ID不能为空')
      return null
    }

    try {
      const response = await getClusterEventAckDetail(ackId)
      if (isApiSuccess(response)) {
        const ack = JSON.parse(response.bizData) as ClusterEventAck
        return ack
      } else {
        message.error(getApiMessage(response, '获取事件确认详情失败'))
        return null
      }
    } catch (error) {
      console.error('获取事件确认详情失败:', error)
      message.error('获取事件确认详情失败')
      return null
    }
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

    // 详情获取
    getAckDetail,
  }
}

/**
 * 服务返回类型
 */
export type ClusterEventAckService = ReturnType<typeof useClusterEventAckService>

