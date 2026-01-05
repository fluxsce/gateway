/**
 * 网关实例列表查询页面级 Hook（仅查询功能）
 * - 组合 useGatewayInstanceListService（纯业务逻辑）
 * - 处理查询等页面交互
 */

import type { Ref } from 'vue'
import { useGatewayInstanceListService } from './service'

/**
 * 网关实例列表查询页面级 Hook（仅查询功能）
 */
export function useGatewayInstanceListPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  // 业务服务（包含 model、查询等）
  const service = useGatewayInstanceListService(searchFormRef)

  // ============= 查询操作 =============

  /**
   * 处理搜索
   */
  const handleSearch = async () => {
    // 重置分页到第一页
    service.model.resetPagination()
    // 加载数据
    await service.loadInstances()
  }

  return {
    // 业务服务
    service,

    // 方法
    handleSearch,
  }
}

/**
 * Page 返回类型
 */
export type GatewayInstanceListPage = ReturnType<typeof useGatewayInstanceListPage>

