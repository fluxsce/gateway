/**
 * 服务中心实例管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as serviceCenterApi from '../api'
import type { ServiceCenterInstance } from '../types'
import { useServiceCenterInstanceModel } from './model'

/**
 * 服务中心实例服务 Hook（纯业务逻辑）
 */
export function useServiceCenterInstanceService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useServiceCenterInstanceModel()

  const {
    loading,
    instanceList,
    pageInfo,
    setInstanceList,
    updatePagination,
    addInstanceToList,
    updateInstanceInList,
    removeInstanceFromList,
    removeInstancesFromList,
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
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await serviceCenterApi.queryServiceCenterInstances(params)

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
        }
      } else {
        message.error(response.errMsg || '查询实例列表失败')
      }
    } catch (error) {
      message.error('加载实例列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索实例
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadInstances 会自动从 searchFormRef 获取查询条件
    await loadInstances(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadInstances()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadInstances 会自动从 searchFormRef 获取查询条件
    await loadInstances()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadInstances()
  }

  // ============= 增删改 =============

  /**
   * 添加实例
   */
  const addInstance = async (instanceData: ServiceCenterInstance): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await serviceCenterApi.addServiceCenterInstance(instanceData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '服务中心实例创建成功')
        message.success(successMsg)
        
        // 如果返回了新增的实例数据，添加到列表
        if (response.bizData) {
          const newInstance = JSON.parse(response.bizData)
          addInstanceToList(newInstance)
        } else {
          // 否则重新加载列表
          await loadInstances()
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '新增实例失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('新增实例失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑实例
   */
  const editInstance = async (instanceData: Partial<ServiceCenterInstance> & {
    instanceName: string
    environment: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await serviceCenterApi.editServiceCenterInstance(instanceData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '服务中心实例更新成功')
        message.success(successMsg)
        
        // 更新列表中的实例数据
        if (response.bizData) {
          const updatedInstance = JSON.parse(response.bizData)
          updateInstanceInList(
            updatedInstance.instanceName,
            updatedInstance.environment,
            updatedInstance.tenantId,
            updatedInstance
          )
        } else {
          // 如果没有返回数据，使用提交的数据更新
          const existingInstance = instanceList.value.find(
            (i) => i.instanceName === instanceData.instanceName && i.environment === instanceData.environment
          )
          if (existingInstance) {
            updateInstanceInList(
              instanceData.instanceName,
              instanceData.environment,
              existingInstance.tenantId,
              instanceData
            )
          }
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '编辑实例失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('编辑实例失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除实例
   */
  const deleteInstance = async (instance: ServiceCenterInstance): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除实例 "${instance.instanceName}" (${instance.environment}) 吗？`,
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
      const response: JsonDataObj = await serviceCenterApi.deleteServiceCenterInstance(
        instance.instanceName,
        instance.environment
      )

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '删除实例成功')
        message.success(successMsg)
        removeInstanceFromList(instance.instanceName, instance.environment, instance.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (instanceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadInstances()
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '删除实例失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('删除实例失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 启动实例
   */
  const startInstance = async (instance: ServiceCenterInstance): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await serviceCenterApi.startServiceCenterInstance(
        instance.instanceName,
        instance.environment
      )

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '实例启动成功')
        message.success(successMsg)
        // 更新本地状态
        updateInstanceInList(
          instance.instanceName,
          instance.environment,
          instance.tenantId,
          { instanceStatus: 'RUNNING' }
        )
        // 重新加载列表以获取最新状态
        await loadInstances()
        return true
      } else {
        const errorMsg = getApiMessage(response, '启动失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('启动失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 停止实例
   */
  const stopInstance = async (instance: ServiceCenterInstance): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await serviceCenterApi.stopServiceCenterInstance(
        instance.instanceName,
        instance.environment
      )

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '实例停止成功')
        message.success(successMsg)
        // 更新本地状态
        updateInstanceInList(
          instance.instanceName,
          instance.environment,
          instance.tenantId,
          { instanceStatus: 'STOPPED' }
        )
        // 重新加载列表以获取最新状态
        await loadInstances()
        return true
      } else {
        const errorMsg = getApiMessage(response, '停止失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('停止失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 重新加载实例配置
   */
  const reloadInstance = async (instance: ServiceCenterInstance): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await serviceCenterApi.reloadServiceCenterInstance(
        instance.instanceName,
        instance.environment
      )

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '配置重载成功')
        message.success(successMsg)
        // 重新加载列表以获取最新状态
        await loadInstances()
        return true
      } else {
        const errorMsg = getApiMessage(response, '配置重载失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('配置重载失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取实例详情
   */
  const getInstanceDetail = async (
    instanceName: string,
    environment: string
  ): Promise<ServiceCenterInstance | null> => {
    try {
      const response: JsonDataObj = await serviceCenterApi.getServiceCenterInstance(
        instanceName,
        environment
      )
      if (isApiSuccess(response)) {
        const instance = parseJsonData<ServiceCenterInstance | null>(response, null)
        return instance
      } else {
        message.error(getApiMessage(response, '获取实例详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取实例详情失败')
      return null
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadInstances,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 实例操作
    addInstance,
    editInstance,
    deleteInstance,
    startInstance,
    stopInstance,
    reloadInstance,
    getInstanceDetail,
  }
}

/**
 * 服务中心实例服务类型
 */
export type ServiceCenterInstanceService = ReturnType<typeof useServiceCenterInstanceService>

