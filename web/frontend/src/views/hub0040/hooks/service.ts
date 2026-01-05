/**
 * 命名空间管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as namespaceApi from '../api'
import type { ServiceGroup } from '../types'
import { useNamespaceModel } from './model'

/**
 * 命名空间服务 Hook（纯业务逻辑）
 */
export function useNamespaceService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useNamespaceModel()

  const {
    loading,
    namespaceList,
    pageInfo,
    setNamespaceList,
    updatePagination,
    addNamespaceToList,
    updateNamespaceInList,
    removeNamespaceFromList
  } = model

  // ============= 数据加载 =============

  /**
   * 加载命名空间列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadNamespaces = async (searchParams?: Record<string, any>) => {
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
      const response: JsonDataObj = await namespaceApi.queryServiceGroups(params)

      // 使用标准的 JsonDataObj 格式（与 hub0002 保持一致）
      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const namespaces = Array.isArray(bizData) ? bizData : []
          setNamespaceList(namespaces)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询命名空间列表失败')
      }
    } catch (error) {
      console.error('加载命名空间列表失败:', error)
      message.error('加载命名空间列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索命名空间
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadNamespaces 会自动从 searchFormRef 获取查询条件
    await loadNamespaces(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadNamespaces()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadNamespaces 会自动从 searchFormRef 获取查询条件
    await loadNamespaces()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadNamespaces()
  }

  // ============= 命名空间操作 =============

  /**
   * 启用命名空间
   */
  const enableNamespace = async (serviceGroupId: string): Promise<boolean> => {
    loading.value = true
    try {
      // 先获取命名空间详情
      const detailResponse: JsonDataObj = await namespaceApi.getServiceGroupDetail(serviceGroupId)
      if (!detailResponse.oK || !detailResponse.state) {
        message.error(detailResponse.errMsg || detailResponse.popMsg || '获取命名空间详情失败')
        return false
      }

      const namespace = detailResponse.bizData as unknown as ServiceGroup
      
      // 通过更新 activeFlag 来启用
      const response: JsonDataObj = await namespaceApi.updateServiceGroup({
        serviceGroupId: namespace.serviceGroupId,
        activeFlag: 'Y'
      })

      if (response.oK && response.state) {
        message.success(response.popMsg || '启用成功')
        await loadNamespaces()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '启用失败')
        return false
      }
    } catch (error) {
      console.error('启用命名空间失败:', error)
      message.error('启用失败，请稍后重试')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 禁用命名空间
   */
  const disableNamespace = async (serviceGroupId: string): Promise<boolean> => {
    loading.value = true
    try {
      // 先获取命名空间详情
      const detailResponse: JsonDataObj = await namespaceApi.getServiceGroupDetail(serviceGroupId)
      if (!detailResponse.oK || !detailResponse.state) {
        message.error(detailResponse.errMsg || detailResponse.popMsg || '获取命名空间详情失败')
        return false
      }

      const namespace = detailResponse.bizData as unknown as ServiceGroup
      
      // 通过更新 activeFlag 来禁用
      const response: JsonDataObj = await namespaceApi.updateServiceGroup({
        serviceGroupId: namespace.serviceGroupId,
        activeFlag: 'N'
      })

      if (response.oK && response.state) {
        message.success(response.popMsg || '禁用成功')
        await loadNamespaces()
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '禁用失败')
        return false
      }
    } catch (error) {
      console.error('禁用命名空间失败:', error)
      message.error('禁用失败，请稍后重试')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除命名空间
   */
  const deleteNamespace = async (namespace: ServiceGroup): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除命名空间 "${namespace.groupName}" 吗？`,
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
      const response: JsonDataObj = await namespaceApi.deleteServiceGroup(namespace.serviceGroupId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除成功')
        removeNamespaceFromList(namespace.serviceGroupId, namespace.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (namespaceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadNamespaces()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除失败')
        return false
      }
    } catch (error) {
      console.error('删除命名空间失败:', error)
      message.error('删除失败，请稍后重试')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建命名空间
   */
  const createNamespace = async (namespaceData: Partial<ServiceGroup>): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await namespaceApi.createServiceGroup(namespaceData as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '创建命名空间成功')
        
        // 如果返回了新增的命名空间数据，添加到列表
        if (response.bizData) {
          const newNamespace = JSON.parse(response.bizData)
          addNamespaceToList(newNamespace)
        } else {
          // 否则重新加载列表
          await loadNamespaces()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '创建命名空间失败')
        return false
      }
    } catch (error) {
      console.error('创建命名空间失败:', error)
      message.error('创建命名空间失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 更新命名空间
   */
  const updateNamespace = async (namespaceData: ServiceGroup): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await namespaceApi.updateServiceGroup(namespaceData as any)

      if (response.oK && response.state) {
        message.success(response.popMsg || '更新命名空间成功')
        
        // 更新列表中的命名空间数据
        if (response.bizData) {
          const updatedNamespace = JSON.parse(response.bizData)
          updateNamespaceInList(updatedNamespace.serviceGroupId, updatedNamespace.tenantId, updatedNamespace)
        } else {
          updateNamespaceInList(namespaceData.serviceGroupId, namespaceData.tenantId, namespaceData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '更新命名空间失败')
        return false
      }
    } catch (error) {
      console.error('更新命名空间失败:', error)
      message.error('更新命名空间失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 查看命名空间详情
   */
  const viewNamespace = async (namespace: ServiceGroup) => {
    try {
      const response: JsonDataObj = await namespaceApi.getServiceGroupDetail(namespace.serviceGroupId)
      
      if (response.oK) {
        const namespaceInfo = JSON.parse(response.bizData)
        return namespaceInfo
      } else {
        message.error(response.errMsg || '获取命名空间详情失败')
        return null
      }
    } catch (error) {
      console.error('获取命名空间详情失败:', error)
      message.error('获取命名空间详情失败')
      return null
    }
  }

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,
    
    // 数据加载
    loadNamespaces,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 命名空间操作
    createNamespace,
    updateNamespace,
    enableNamespace,
    disableNamespace,
    deleteNamespace,
    viewNamespace
  }
}

/**
 * 服务返回类型
 */
export type NamespaceService = ReturnType<typeof useNamespaceService>

