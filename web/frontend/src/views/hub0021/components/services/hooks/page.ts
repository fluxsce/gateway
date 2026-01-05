/**
 * 服务定义列表页面级 Hook
 * - 组合 useServiceDefinitionSelectorService（纯业务逻辑）
 * - 处理搜索、分页等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import type { ServiceDefinitionSelectorModel } from './model'
import { useServiceDefinitionSelectorService } from './service'

/**
 * 服务定义列表页面级 Hook
 * @param model Model hook
 * @param gatewayInstanceId 网关实例ID（可选）
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useServiceDefinitionListPage(
  model: ServiceDefinitionSelectorModel,
  gatewayInstanceId?: Ref<string | undefined> | string,
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 获取 gatewayInstanceId 的值（可能是 Ref 或普通值）
  const getGatewayInstanceId = () => {
    return typeof gatewayInstanceId === 'string' ? gatewayInstanceId : gatewayInstanceId?.value
  }

  // 业务服务
  const service = useServiceDefinitionSelectorService(model, getGatewayInstanceId(), searchFormRef)

  // 移除自动加载逻辑
  // gatewayInstanceId 变化时不清空数据，也不自动加载
  // 数据加载应该由用户主动搜索触发

  /**
   * 处理搜索
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    const instanceId = getGatewayInstanceId()
    if (!instanceId) {
      message.warning('请先选择网关实例')
      return
    }

    // 重置分页后加载服务定义列表
    model.resetPagination()
    await service.loadServiceDefinitions(formData)
  }

  /**
   * 处理重置
   */
  const handleReset = async () => {
    if (searchFormRef?.value?.resetFields) {
      searchFormRef.value.resetFields()
    }
    model.resetPagination()
    await handleSearch()
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    model.updatePagination({
      pageIndex: currentPage,
      pageSize: pageSize,
    })
    await service.loadServiceDefinitions()
  }

  /**
   * 工具栏按钮点击处理（当前不需要，但保留接口）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'search':
        await handleSearch(formData)
        break
      case 'reset':
        await handleReset()
        break
      default:
        break
    }
  }

  return {
    service,
    handleSearch,
    handleReset,
    handlePageChange,
    handleToolbarClick,
  }
}

