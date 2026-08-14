/**
 * 模板列表查询页面级 Hook（仅查询功能）
 * 用于模板选择器组件
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import type { RsGridExpose } from '@/components/rs-grid'
import type { Ref } from 'vue'
import { useAlertTemplateListService } from './service'

/**
 * 模板列表查询页面级 Hook（仅查询功能）
 */
export function useAlertTemplateListPage(
  gridRef?: Ref<RsGridExpose | null>,
  searchFormRef?: Ref<RsSearchFormExpose | null>,
  channelType?: string,
) {
  void gridRef
  const service = useAlertTemplateListService(searchFormRef, channelType)

  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadTemplateList(searchParams)
  }

  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadTemplateList()
  }

  return {
    service,
    handleSearch,
    handlePageChange,
  }
}

/**
 * Page 返回类型
 */
export type AlertTemplateListPage = ReturnType<typeof useAlertTemplateListPage>
