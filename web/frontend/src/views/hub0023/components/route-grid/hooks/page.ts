/**
 * 路由列表查询页面级 Hook（仅查询功能）
 * - 组合 useRouteListService（纯业务逻辑）
 * - 处理搜索等页面交互
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import type { RsGridExpose } from '@/components/rs-grid'
import type { Ref } from 'vue'
import { useRouteListService } from './service'

/**
 * 路由列表查询页面级 Hook
 */
export function useRouteListPage(
  gatewayInstanceId?: string,
  _gridRef?: Ref<RsGridExpose | null>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
) {
  const service = useRouteListService(gatewayInstanceId, searchFormRef)

  /**
   * 处理搜索（接收 RsSearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  return {
    service,
    handleSearch,
  }
}

export type RouteListPage = ReturnType<typeof useRouteListPage>
