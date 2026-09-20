/**
 * 路由列表服务 Hook（纯业务逻辑）
 */

import { createBackendPaginationParams } from '@/utils/pagination'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useAppMessage } from '@/composables/useAppMessage'
import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { toValue, type MaybeRefOrGetter, type Ref } from 'vue'
import { queryRouteConfigs } from '@/views/hub0021/api'
import type { RouteConfig } from '@/views/hub0021/components/routes/types'
import { useRouteListModel } from './model'

function resolveGatewayInstanceId(
  gatewayInstanceId?: MaybeRefOrGetter<string | undefined>,
): string | undefined {
  const id = toValue(gatewayInstanceId)
  const trimmed = typeof id === 'string' ? id.trim() : ''
  return trimmed || undefined
}

/**
 * 路由列表服务 Hook（纯业务逻辑）
 */
export function useRouteListService(
  gatewayInstanceId?: MaybeRefOrGetter<string | undefined>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const message = useAppMessage()

  // 初始化 Model
  const model = useRouteListModel()

  const {
    loading,
    routeList,
    pageInfo,
    setRouteList,
    updatePagination,
    clearRouteList,
    resetPagination,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载路由列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadRoutes = async (searchParams?: Record<string, any>) => {
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

      const boundInstanceId = resolveGatewayInstanceId(gatewayInstanceId)

      // 构建请求参数：合并查询条件和分页参数
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 弹窗实例 ID 可能后变，每次加载再取当前值
        ...(effectiveSearchParams.gatewayInstanceId === undefined && boundInstanceId
          ? { gatewayInstanceId: boundInstanceId }
          : {}),
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryRouteConfigs(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const routes = parseJsonData<RouteConfig[]>(response, [])
        setRouteList(routes)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        updatePagination(backendPageInfo)
      } else {
        message.error(getApiMessage(response, '查询路由列表失败'))
      }
    } catch (error) {
      message.error('加载路由列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索路由
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    // 重置分页到第一页
    resetPagination()
    await loadRoutes(formData)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    // 更新分页信息
    updatePagination({ pageIndex: currentPage, pageSize })
    // 重新加载数据
    await loadRoutes()
  }

  return {
    model,
    loadRoutes,
    handleSearch,
    handlePageChange,
  }
}

/**
 * Service 返回类型
 */
export type RouteListService = ReturnType<typeof useRouteListService>

