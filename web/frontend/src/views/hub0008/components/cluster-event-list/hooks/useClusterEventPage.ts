import type { Ref } from 'vue'
import { useClusterEventService } from './useClusterEventService'

/**
 * 集群事件页面级 Hook
 * - 组合 useClusterEventService（纯业务逻辑）
 * - 处理页面交互
 */
export function useClusterEventPage(searchFormRef?: Ref<any> | any) {
  // 业务服务（包含 model、增删改查等）
  const service = useClusterEventService(searchFormRef)

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 事件处理器
    handleSearch
  }
}

export type ClusterEventPage = ReturnType<typeof useClusterEventPage>

