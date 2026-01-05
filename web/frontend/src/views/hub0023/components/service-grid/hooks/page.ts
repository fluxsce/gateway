/**
 * 服务列表查询页面级 Hook（仅查询功能）
 * - 组合 useServiceListService（纯业务逻辑）
 * - 处理搜索等页面交互
 */

import type { Ref } from 'vue'
import { useServiceListService } from './service'

/**
 * 服务列表查询页面级 Hook
 */
export function useServiceListPage(gatewayInstanceId?: string, gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  // 业务服务（包含 model、查询等）
  const service = useServiceListService(gatewayInstanceId, searchFormRef)

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  return {
    // 业务服务（包含 model 与查询）
    service,

    // 事件处理器
    handleSearch,
  }
}

export type ServiceListPage = ReturnType<typeof useServiceListPage>

