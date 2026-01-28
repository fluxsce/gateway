/**
 * 服务监控业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as serviceApi from '../api'
import type { Service } from '../types'
import { useServiceModel } from './model'

/**
 * 服务监控服务 Hook（纯业务逻辑）
 */
export function useServiceService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useServiceModel()

  const {
    loading,
    serviceList,
    pageInfo,
    setServiceList,
    updatePagination,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList,
    removeServicesFromList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载服务列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   * @param requiredNamespaceId 必须的命名空间ID（如果提供，将强制使用此值）
   */
  const loadServices = async (searchParams?: Record<string, any>, requiredNamespaceId?: string) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 如果提供了必须的命名空间ID，强制使用它
      if (requiredNamespaceId) {
        finalSearchParams = {
          ...finalSearchParams,
          namespaceId: requiredNamespaceId,
        }
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
        ...effectiveSearchParams,
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API
      const response: JsonDataObj = await serviceApi.queryServices(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const services = Array.isArray(bizData) ? bizData : []
          setServiceList(services)
        }

        // 更新分页信息
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || getApiMessage(response))
        setServiceList([])
      }
    } catch (error: any) {
      message.error('加载服务列表失败: ' + (error.message || '未知错误'))
      setServiceList([])
    } finally {
      loading.value = false
    }
  }

  /**
   * 搜索服务
   * @param requiredNamespaceId 必须的命名空间ID（如果提供，将强制使用此值）
   */
  const handleSearch = async (requiredNamespaceId?: string) => {
    model.resetPagination()
    await loadServices(undefined, requiredNamespaceId)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    if (searchFormRef?.value?.resetForm) {
      searchFormRef.value.resetForm()
    }
    model.resetPagination()
    await loadServices({})
  }

  /**
   * 分页变化处理
   */
  const handlePageChange = async (pageIndex: number, pageSize: number) => {
    if (pageInfo.value) {
      pageInfo.value.pageIndex = pageIndex
      pageInfo.value.pageSize = pageSize
    }
    await loadServices()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadServices()
  }

  // ============= 服务增删改 =============

  /**
   * 添加服务
   */
  const addService = async (data: Service): Promise<boolean> => {
    try {
      const response: JsonDataObj = await serviceApi.addService(data)
      if (isApiSuccess(response)) {
        message.success(getApiMessage(response))
        return true
      } else {
        message.error(getApiMessage(response))
        return false
      }
    } catch (error: any) {
      message.error('添加服务失败: ' + (error.message || '未知错误'))
      return false
    }
  }

  /**
   * 编辑服务
   */
  const editService = async (data: Partial<Service> & { namespaceId: string; groupName: string; serviceName: string }): Promise<boolean> => {
    try {
      const response: JsonDataObj = await serviceApi.editService(data)
      if (isApiSuccess(response)) {
        message.success(getApiMessage(response))
        return true
      } else {
        message.error(getApiMessage(response))
        return false
      }
    } catch (error: any) {
      message.error('编辑服务失败: ' + (error.message || '未知错误'))
      return false
    }
  }

  /**
   * 删除服务
   */
  const deleteService = async (service: Service): Promise<boolean> => {
    try {
      const response: JsonDataObj = await serviceApi.deleteService(
        service.namespaceId,
        service.groupName,
        service.serviceName
      )
      if (isApiSuccess(response)) {
        message.success(getApiMessage(response))
        return true
      } else {
        message.error(getApiMessage(response))
        return false
      }
    } catch (error: any) {
      message.error('删除服务失败: ' + (error.message || '未知错误'))
      return false
    }
  }

  /**
   * 获取服务详情
   */
  const getServiceDetail = async (namespaceId: string, groupName: string, serviceName: string): Promise<Service | null> => {
    try {
      const response: JsonDataObj = await serviceApi.getService(namespaceId, groupName, serviceName)
      if (isApiSuccess(response)) {
        if (response.bizData) {
          return JSON.parse(response.bizData) as Service
        }
        return null
      } else {
        message.error(getApiMessage(response))
        return null
      }
    } catch (error: any) {
      message.error('获取服务详情失败: ' + (error.message || '未知错误'))
      return null
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据状态
    loading,
    serviceList,
    pageInfo,

    // 方法
    loadServices,
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    addService,
    editService,
    deleteService,
    getServiceDetail,
  }
}

export type ServiceService = ReturnType<typeof useServiceService>

