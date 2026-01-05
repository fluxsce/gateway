/**
 * 服务定义选择器 Service
 * 处理数据加载和API调用
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import { queryServiceDefinitions } from '../../../api'
import type { ServiceDefinition } from '../types'
import type { ServiceDefinitionSelectorModel } from './model'

/**
 * 服务定义选择器 Service
 */
export function useServiceDefinitionSelectorService(
  model: ServiceDefinitionSelectorModel,
  gatewayInstanceId?: string,
  searchFormRef?: any
) {
  const message = useMessage()

  const {
    loading,
    setServiceList,
    setLoading,
    updatePagination,
  } = model

  /**
   * 加载服务定义列表
   */
  async function loadServiceDefinitions(customParams: any = {}) {
    if (!gatewayInstanceId) {
      setServiceList([])
      return
    }

    try {
      setLoading(true)
      
      // 获取搜索表单数据
      let searchParams = customParams
      if (!searchParams && searchFormRef?.value?.getFormData) {
        searchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤空值
      const effectiveSearchParams = Object.fromEntries(
        Object.entries(searchParams).filter(([, value]) => value !== '' && value !== null && value !== undefined)
      )

      // 构建请求参数，包含分页和搜索条件
      const params = {
        gatewayInstanceId, // 作为筛选条件
        ...effectiveSearchParams,
        ...createBackendPaginationParams(model.pageInfo.value?.pageIndex, model.pageInfo.value?.pageSize),
      }

      const response = await queryServiceDefinitions(params as any)

      if (isApiSuccess(response)) {
        const serviceList = parseJsonData<ServiceDefinition[]>(response, [])
        setServiceList(Array.isArray(serviceList) ? serviceList : [])
        // 解析并更新分页信息
        updatePagination(parsePageInfo(response))
      } else {
        message.error(response.errMsg || '获取服务定义列表失败')
        setServiceList([])
        updatePagination({})
      }
    } catch (error: any) {
      console.error('加载服务定义列表失败:', error)
      message.error('加载服务定义列表失败: ' + (error?.message || '未知错误'))
      setServiceList([])
      updatePagination({})
    } finally {
      setLoading(false)
    }
  }

  return {
    loadServiceDefinitions,
  }
}

