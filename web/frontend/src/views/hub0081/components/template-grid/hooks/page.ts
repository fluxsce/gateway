/**
 * 模板列表查询页面级 Hook（仅查询功能）
 * 用于模板选择器组件
 */

import type { Ref } from 'vue'
import { useAlertTemplateListService } from './service'

/**
 * 模板列表查询页面级 Hook（仅查询功能）
 */
export function useAlertTemplateListPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any, channelType?: string) {
  // 业务服务（包含 model、查询等）
  const service = useAlertTemplateListService(searchFormRef, channelType)

  // ============= 查询操作 =============

  /**
   * 处理搜索
   */
  const handleSearch = async () => {
    // 重置分页到第一页
    service.model.resetPagination()
    // 加载数据
    await service.loadTemplateList()
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadTemplateList()
  }

  return {
    // 业务服务
    service,

    // 方法
    handleSearch,
    handlePageChange,
  }
}

/**
 * Page 返回类型
 */
export type AlertTemplateListPage = ReturnType<typeof useAlertTemplateListPage>

