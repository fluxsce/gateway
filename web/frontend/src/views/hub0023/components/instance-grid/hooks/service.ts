/**
 * 网关实例列表查询业务逻辑层（仅查询功能）
 * 处理所有与后端交互的业务逻辑
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import * as gatewayApi from '@/views/hub0020/api'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { useGatewayInstanceListModel } from './model'

/**
 * 网关实例列表查询服务 Hook（纯业务逻辑，仅查询功能）
 */
export function useGatewayInstanceListService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useGatewayInstanceListModel()

  const {
    loading,
    instanceList,
    pageInfo,
    setInstanceList,
    updatePagination,
    clearInstanceList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载实例列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadInstances = async (searchParams?: Record<string, any>) => {
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
        ),
      }

      // 调用后端接口
      const response: JsonDataObj = await gatewayApi.queryGatewayInstances(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const instances = Array.isArray(bizData) ? bizData : []
          setInstanceList(instances)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        } else {
          // 如果没有分页信息，清空分页状态
          model.resetPagination()
        }

        return { success: true }
      } else {
        const errorMsg = response.errMsg || '查询网关实例列表失败'
        message.error(errorMsg)
        clearInstanceList()
        model.resetPagination()
        return { success: false, error: errorMsg }
      }
    } catch (error: any) {
      console.error('查询网关实例列表失败:', error)
      message.error(error?.message || '查询网关实例列表失败')
      clearInstanceList()
      model.resetPagination()
      return { success: false, error: error?.message }
    } finally {
      loading.value = false
    }
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    // 更新分页信息
    updatePagination({ pageIndex: currentPage, pageSize })
    // 重新加载数据
    await loadInstances()
  }

  return {
    // Model
    model,

    // 数据状态（从 model 中暴露）
    loading,
    instanceList,
    pageInfo,

    // 方法
    loadInstances,
    handlePageChange,
  }
}

/**
 * Service 返回类型
 */
export type GatewayInstanceListService = ReturnType<typeof useGatewayInstanceListService>

