/**
 * 服务定义列表业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { addServiceDefinition, deleteServiceDefinition, editServiceDefinition, queryServiceDefinitions } from '../../../api'
import type { ServiceDefinition } from '../types'
import { useServiceDefinitionModel } from './model'

/**
 * 服务定义列表服务 Hook（纯业务逻辑）
 * @param gatewayInstanceId 网关实例ID（作为proxyConfigId使用）
 */
export function useServiceDefinitionService(gatewayInstanceId?: string, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model（暂时不传递 pageHook，在 page.ts 中动态更新）
  const model = useServiceDefinitionModel()

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
   * 加载服务定义列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadServiceList = async (searchParams?: Record<string, any>) => {
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
        // 如果 searchParams 中没有 proxyConfigId，且构造函数参数有，则使用构造函数参数的（gatewayInstanceId作为proxyConfigId）
        ...(effectiveSearchParams.proxyConfigId === undefined && gatewayInstanceId ? { proxyConfigId: gatewayInstanceId } : {}),
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize),
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryServiceDefinitions(params as any)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const services = Array.isArray(bizData) ? bizData : []
          setServiceList(services)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询服务定义列表失败')
      }
    } catch (error) {
      message.error('加载服务定义列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索服务定义
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadServiceList 会自动从 searchFormRef 获取查询条件
    await loadServiceList(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadServiceList()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadServiceList()
  }

  // ============= 数据操作 =============

  /**
   * 删除服务定义
   */
  const deleteService = async (serviceDefinitionId: string): Promise<boolean> => {
    // 查找要删除的服务定义信息，用于确认对话框
    const serviceToDelete = serviceList.value.find((service) => service.serviceDefinitionId === serviceDefinitionId)
    if (!serviceToDelete) {
      message.error('未找到要删除的服务定义')
      return false
    }

    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除服务定义 "${serviceToDelete.serviceName || serviceToDelete.serviceDefinitionId}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response = await deleteServiceDefinition(serviceDefinitionId, 'default')
      if (response.oK) {
        message.success(getApiMessage(response, '删除服务定义成功'))
        removeServiceFromList(serviceDefinitionId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (serviceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadServiceList()
        } else {
          await loadServiceList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除服务定义失败'))
        return false
      }
    } catch (error) {
      message.error('删除服务定义失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除服务定义
   */
  const batchDeleteServices = async (serviceDefinitionIds: string[]) => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${serviceDefinitionIds.length} 个服务定义吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return
    }

    loading.value = true
    try {
      let successCount = 0
      let failCount = 0

      // 逐个删除
      for (const id of serviceDefinitionIds) {
        try {
          const response = await deleteServiceDefinition(id, 'default')
          if (response.oK) {
            successCount++
          } else {
            failCount++
          }
        } catch {
          failCount++
        }
      }

      // 显示结果
      if (successCount > 0) {
        message.success(`成功删除 ${successCount} 个服务定义${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
        removeServicesFromList(serviceDefinitionIds.slice(0, successCount))
        
        // 重新加载列表
        await loadServiceList()
      } else {
        message.error(`删除失败，共 ${failCount} 个`)
      }
    } catch (error) {
      message.error('批量删除失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 切换服务定义状态
   */
  const toggleServiceStatus = async (service: ServiceDefinition) => {
    const newStatus = service.activeFlag === 'Y' ? 'N' : 'Y'
    const action = newStatus === 'Y' ? '启用' : '禁用'

    try {
      const response = await editServiceDefinition({
        serviceDefinitionId: service.serviceDefinitionId,
        tenantId: 'default',
        activeFlag: newStatus,
      })

      if (response.oK) {
        message.success(`${action}成功`)
        updateServiceInList({ ...service, activeFlag: newStatus })
        await loadServiceList()
      } else {
        message.error(response.errMsg || `${action}失败`)
      }
    } catch (error) {
      message.error(`${action}失败`)
    }
  }

  /**
   * 批量切换服务定义状态
   */
  const batchToggleServiceStatus = async (serviceDefinitionIds: string[], status: 'Y' | 'N') => {
    const action = status === 'Y' ? '启用' : '禁用'

    try {
      const updatePromises = serviceDefinitionIds.map((id) =>
        editServiceDefinition({
          serviceDefinitionId: id,
          tenantId: 'default',
          activeFlag: status,
        })
      )
      const responses = await Promise.all(updatePromises)

      const successCount = responses.filter((response: any) => response.oK).length
      if (successCount === serviceDefinitionIds.length) {
        message.success(`成功${action} ${successCount} 个服务定义`)
      } else {
        message.warning(`${action}了 ${successCount}/${serviceDefinitionIds.length} 个服务定义`)
      }

      await loadServiceList()
    } catch (error) {
      message.error(`批量${action}失败`)
    }
  }

  /**
   * 添加服务定义
   */
  const addService = async (serviceData: Partial<ServiceDefinition> & { tenantId: string }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await addServiceDefinition(serviceData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '服务定义创建成功')
        message.success(successMsg)

        // 如果返回了新增的服务数据，添加到列表
        if (response.bizData) {
          const newService = JSON.parse(response.bizData)
          addServiceToList(newService)
        } else {
          // 否则重新加载列表
          await loadServiceList()
        }

        return true
      } else {
        const errorMsg = getApiMessage(response, '新增服务定义失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('新增服务定义失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑服务定义
   */
  const editService = async (
    serviceDefinitionId: string,
    serviceData: Partial<ServiceDefinition> & { tenantId: string }
  ): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await editServiceDefinition({
        ...serviceData,
        serviceDefinitionId,
      })

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '服务定义更新成功')
        message.success(successMsg)

        // 如果返回了更新的服务数据，更新列表中的对应项
        if (response.bizData) {
          const updatedService = JSON.parse(response.bizData)
          updateServiceInList(updatedService)
        } else {
          // 否则重新加载列表
          await loadServiceList()
        }

        return true
      } else {
        const errorMsg = getApiMessage(response, '更新服务定义失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('更新服务定义失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadServiceList,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,

    // 服务定义操作
    addService,
    editService,
    deleteService,
    batchDeleteServices,
    toggleServiceStatus,
    batchToggleServiceStatus,
  }
}

/**
 * 服务定义列表服务类型
 */
export type ServiceDefinitionService = ReturnType<typeof useServiceDefinitionService>

