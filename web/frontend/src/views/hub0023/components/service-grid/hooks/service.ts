/**
 * 服务列表服务 Hook（纯业务逻辑）
 */

import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { queryAllServiceDefinitions } from '@/views/hub0021/api'
import type { ServiceDefinition } from '@/views/hub0022/components/service/types'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { useServiceListModel } from './model'

/**
 * 服务列表服务 Hook（纯业务逻辑）
 */
export function useServiceListService(gatewayInstanceId?: string, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useServiceListModel()

  const {
    loading,
    serviceList,
    pageInfo,
    setServiceList,
    updatePagination,
    clearServiceList,
    resetPagination,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载服务列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadServices = async (searchParams?: Record<string, any>) => {
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
      // 注意：使用 queryAllServiceDefinitions，不依赖代理配置ID
      const params = {
        // 查询条件（排除 gatewayInstanceId 和 proxyConfigId）
        ...Object.fromEntries(
          Object.entries(effectiveSearchParams).filter(
            ([key]) => key !== 'gatewayInstanceId' && key !== 'proxyConfigId'
          )
        ),
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      // 使用 queryAllServiceDefinitions，不依赖代理配置ID
      const response: JsonDataObj = await queryAllServiceDefinitions(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const services = parseJsonData<ServiceDefinition[]>(response, [])
        setServiceList(services)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        updatePagination(backendPageInfo)
      } else {
        message.error(getApiMessage(response, '查询服务列表失败'))
      }
    } catch (error) {
      message.error('加载服务列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索服务
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    // 重置分页到第一页
    resetPagination()
    await loadServices(formData)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    // 更新分页信息
    updatePagination({ pageIndex: currentPage, pageSize })
    // 重新加载数据
    await loadServices()
  }

  return {
    model,
    loadServices,
    handleSearch,
    handlePageChange,
  }
}

/**
 * Service 返回类型
 */
export type ServiceListService = ReturnType<typeof useServiceListService>

