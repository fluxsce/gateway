/**
 * 隧道服务管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { PlayCircleOutline, StopCircleOutline, WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as tunnelServiceApi from '../../../api'
import type { TunnelService } from '../../../types'
import { useTunnelServiceModel } from './model'

/**
 * 隧道服务服务 Hook（纯业务逻辑）
 */
export function useTunnelServiceService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useTunnelServiceModel()

  const {
    loading,
    serviceList,
    pageInfo,
    setServiceList,
    updatePagination,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList
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
      const paginationParams = createBackendPaginationParams(
        pageInfo.value?.pageIndex,
        pageInfo.value?.pageSize
      )
      
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（使用 pageIndex 与后端保持一致）
        pageIndex: paginationParams.pageIndex,
        pageSize: paginationParams.pageSize
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await tunnelServiceApi.queryTunnelServices(params)

      // 使用标准的 JsonDataObj 格式
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
        message.error(response.errMsg || '查询服务列表失败')
      }
    } catch (error) {
      console.error('加载服务列表失败:', error)
      message.error('加载服务列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索服务
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadServices 会自动从 searchFormRef 获取查询条件
    await loadServices(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadServices()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadServices 会自动从 searchFormRef 获取查询条件
    await loadServices()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadServices()
  }

  // ============= 服务操作 =============

  /**
   * 创建服务
   */
  const createService = async (serviceData: Partial<TunnelService>): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServiceApi.createTunnelService(serviceData as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '创建服务成功')
        
        // 如果返回了新增的服务数据，添加到列表
        if (response.bizData) {
          const newService = JSON.parse(response.bizData)
          addServiceToList(newService)
        } else {
          // 否则重新加载列表
          await loadServices()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '创建服务失败')
        return false
      }
    } catch (error) {
      console.error('创建服务失败:', error)
      message.error('创建服务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新服务
   */
  const updateService = async (serviceData: TunnelService): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServiceApi.updateTunnelService(serviceData as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '更新服务成功')
        
        // 更新列表中的服务数据
        if (response.bizData) {
          const updatedService = JSON.parse(response.bizData)
          updateServiceInList(updatedService.tunnelServiceId, updatedService.tenantId, updatedService)
        } else {
          updateServiceInList(serviceData.tunnelServiceId, serviceData.tenantId, serviceData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '更新服务失败')
        return false
      }
    } catch (error) {
      console.error('更新服务失败:', error)
      message.error('更新服务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除服务
   */
  const deleteService = async (service: TunnelService): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除服务 "${service.serviceName}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServiceApi.deleteTunnelService(service.tunnelServiceId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除服务成功')
        removeServiceFromList(service.tunnelServiceId, service.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (serviceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadServices()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除服务失败')
        return false
      }
    } catch (error) {
      console.error('删除服务失败:', error)
      message.error('删除服务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 注册服务
   */
  const registerService = async (service: TunnelService): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认注册',
      subtitle: '注册后服务将可在隧道服务器上使用',
      content: `确定要注册服务 "${service.serviceName}" 吗？`,
      icon: PlayCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定注册',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServiceApi.registerService(service.tunnelServiceId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '注册服务成功')
        await loadServices()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '注册服务失败')
        return false
      }
    } catch (error) {
      console.error('注册服务失败:', error)
      message.error('注册服务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 注销服务
   */
  const unregisterService = async (service: TunnelService): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认注销',
      subtitle: '注销后服务将从隧道服务器上移除',
      content: `确定要注销服务 "${service.serviceName}" 吗？`,
      icon: StopCircleOutline,
      headerStyle: 'gradient',
      positiveText: '确定注销',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await tunnelServiceApi.unregisterService(service.tunnelServiceId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '注销服务成功')
        await loadServices()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '注销服务失败')
        return false
      }
    } catch (error) {
      console.error('注销服务失败:', error)
      message.error('注销服务失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 查看服务详情
   */
  const viewService = async (service: TunnelService) => {
    try {
      const response: JsonDataObj = await tunnelServiceApi.getTunnelService(service.tunnelServiceId)
      
      if (response.oK) {
        const serviceInfo = JSON.parse(response.bizData)
        return serviceInfo
      } else {
        message.error(response.errMsg || '获取服务详情失败')
        return null
      }
    } catch (error) {
      console.error('获取服务详情失败:', error)
      message.error('获取服务详情失败')
      return null
    }
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,
    
    // 数据加载
    loadServices,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 服务操作
    createService,
    updateService,
    deleteService,
    registerService,
    unregisterService,
    viewService
  }
}

/**
 * 服务返回类型
 */
export type TunnelServiceService = ReturnType<typeof useTunnelServiceService>

