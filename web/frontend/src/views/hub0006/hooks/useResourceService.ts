/**
 * 权限资源管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as resourceApi from '../api'
import type { Resource } from '../types/index'
import { useResourceModel } from './model'

/**
 * 资源服务 Hook（纯业务逻辑，不再依赖外部 options）
 */
export function useResourceService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useResourceModel()

  const {
    loading,
    resourceList,
    pageInfo,
    setResourceList,
    updatePagination,
    addResourceToList,
    updateResourceInList,
    removeResourceFromList,
    removeResourcesFromList
  } = model

  // ============= 数据加载 =============

  /**
   * 加载资源列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadResources = async (searchParams?: Record<string, any>) => {
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
      const response: JsonDataObj = await resourceApi.queryResources(params)

      if (response.oK) {
        // 解析业务数据 - 树形结构数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          // 如果返回的是数组，直接使用；如果是对象且包含数组，提取数组
          const resources = Array.isArray(bizData) 
            ? bizData 
            : (Array.isArray(bizData?.data) ? bizData.data : [])
          
          setResourceList(resources)
        }
      } else {
        message.error(response.errMsg || '查询资源列表失败')
      }
    } catch (error) {
      console.error('加载资源列表失败:', error)
      message.error('加载资源列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索资源
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadResources 会自动从 searchFormRef 获取查询条件
    await loadResources(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadResources()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadResources 会自动从 searchFormRef 获取查询条件
    await loadResources()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadResources()
  }

  // ============= 增删改 =============

  /**
   * 添加资源
   */
  const addResource = async (resourceData: Resource): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await resourceApi.addResource(resourceData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '新增资源成功')
        
        // 如果返回了新增的资源数据，添加到列表
        if (response.bizData) {
          const newResource = JSON.parse(response.bizData)
          addResourceToList(newResource)
        } else {
          // 否则重新加载列表
          await loadResources()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '新增资源失败')
        return false
      }
    } catch (error) {
      console.error('新增资源失败:', error)
      message.error('新增资源失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑资源
   */
  const editResource = async (resourceData: Resource): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await resourceApi.editResource(resourceData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '编辑资源成功')
        
        // 更新列表中的资源数据
        if (response.bizData) {
          const updatedResource = JSON.parse(response.bizData)
          updateResourceInList(updatedResource.resourceId, updatedResource.tenantId, updatedResource)
        } else {
          updateResourceInList(resourceData.resourceId, resourceData.tenantId, resourceData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '编辑资源失败')
        return false
      }
    } catch (error) {
      console.error('编辑资源失败:', error)
      message.error('编辑资源失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除资源
   */
  const deleteResource = async (resource: Resource): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除资源 "${resource.resourceName}" 吗？`,
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
      const response: JsonDataObj = await resourceApi.deleteResource(resource.resourceId, resource.tenantId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除资源成功')
        removeResourceFromList(resource.resourceId, resource.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (resourceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadResources()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除资源失败')
        return false
      }
    } catch (error) {
      console.error('删除资源失败:', error)
      message.error('删除资源失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除资源
   */
  const batchDeleteResources = async (resources: Resource[]): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${resources.length} 个资源吗？`,
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
      let successCount = 0
      let failCount = 0

      // 逐个删除
      for (const resource of resources) {
        try {
          const response: JsonDataObj = await resourceApi.deleteResource(resource.resourceId, resource.tenantId)
          if (response.oK && response.state) {
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
        message.success(`成功删除 ${successCount} 个资源${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
        removeResourcesFromList(resources.slice(0, successCount))
        
        // 重新加载列表
        await loadResources()
        return true
      } else {
        message.error(`删除失败，共 ${failCount} 个`)
        return false
      }
    } catch (error) {
      console.error('批量删除资源失败:', error)
      message.error('批量删除资源失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadResources,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 资源操作
    addResource,
    editResource,
    deleteResource,
    batchDeleteResources,
  }
}

/**
 * 资源服务类型
 */
export type ResourceService = ReturnType<typeof useResourceService>

