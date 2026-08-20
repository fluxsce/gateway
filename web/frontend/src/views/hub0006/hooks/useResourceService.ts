/**
 * 权限资源管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { rsConfirm } from '@/ui'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { JsonDataObj } from '@/types/api'
import { createBackendPaginationParams } from '@/utils/pagination'
import { WarningOutline } from '@vicons/ionicons5'
import type { Ref } from 'vue'
import * as resourceApi from '../api'
import type { Resource } from '../types/index'
import { useResourceModel } from './model'

/**
 * 资源服务 Hook（纯业务逻辑，不再依赖外部 options）
 */
export function useResourceService(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0006')

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
    removeResourcesFromList,
  } = model

  /**
   * 加载资源列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadResources = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined,
            ),
          )
        : {}

      const params = {
        ...effectiveSearchParams,
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize),
      }

      const response: JsonDataObj = await resourceApi.queryResources(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const resources = Array.isArray(bizData)
            ? bizData
            : Array.isArray(bizData?.data)
              ? bizData.data
              : []

          setResourceList(resources)
        }
      } else {
        message.error(response.errMsg || t('resource.message.queryFailed'))
      }
    } catch (error) {
      message.error(t('resource.message.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  const handleSearch = async (searchParams?: Record<string, any>) => {
    await loadResources(searchParams)
  }

  const handleReset = async () => {
    model.resetPagination()
    await loadResources()
  }

  const handlePageChange = async ({
    currentPage,
    pageSize,
  }: {
    currentPage: number
    pageSize: number
  }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadResources()
  }

  const handleRefresh = async () => {
    await loadResources()
  }

  const addResource = async (resourceData: Resource): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await resourceApi.addResource(resourceData)

      if (response.oK && response.state) {
        message.success(response.popMsg || t('resource.message.createSuccess'))

        if (response.bizData) {
          const newResource = JSON.parse(response.bizData)
          addResourceToList(newResource)
        } else {
          await loadResources()
        }

        return true
      } else {
        message.error(response.errMsg || response.popMsg || t('resource.message.createFailed'))
        return false
      }
    } catch (error) {
      message.error(t('resource.message.createFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const editResource = async (resourceData: Resource): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await resourceApi.editResource(resourceData)

      if (response.oK && response.state) {
        message.success(response.popMsg || t('resource.message.updateSuccess'))

        if (response.bizData) {
          const updatedResource = JSON.parse(response.bizData)
          updateResourceInList(
            updatedResource.resourceId,
            updatedResource.tenantId,
            updatedResource,
          )
        } else {
          updateResourceInList(resourceData.resourceId, resourceData.tenantId, resourceData)
        }

        return true
      } else {
        message.error(response.errMsg || response.popMsg || t('resource.message.updateFailed'))
        return false
      }
    } catch (error) {
      message.error(t('resource.message.updateFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteResource = async (resource: Resource): Promise<boolean> => {
    const confirmed = await rsConfirm.warning({
      title: t('resource.message.deleteConfirmTitle'),
      subtitle: t('resource.message.deleteConfirmSubtitle'),
      description: t('resource.message.deleteConfirm', { name: resource.resourceName }),
      icon: WarningOutline,
      confirmText: t('resource.message.deleteConfirmOk'),
      cancelText: t('resource.message.deleteConfirmCancel'),
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await resourceApi.deleteResource(
        resource.resourceId,
        resource.tenantId,
      )

      if (response.oK && response.state) {
        message.success(response.popMsg || t('resource.message.deleteSuccess'))
        removeResourceFromList(resource.resourceId, resource.tenantId)

        if (resourceList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadResources()
        }

        return true
      } else {
        message.error(response.errMsg || response.popMsg || t('resource.message.deleteFailed'))
        return false
      }
    } catch (error) {
      message.error(t('resource.message.deleteFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const batchDeleteResources = async (resources: Resource[]): Promise<boolean> => {
    const confirmed = await rsConfirm.warning({
      title: t('resource.message.batchDeleteConfirmTitle'),
      subtitle: t('resource.message.deleteConfirmSubtitle'),
      description: t('resource.message.batchDeleteConfirm', { count: resources.length }),
      icon: WarningOutline,
      confirmText: t('resource.message.deleteConfirmOk'),
      cancelText: t('resource.message.deleteConfirmCancel'),
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      let successCount = 0
      let failCount = 0

      for (const resource of resources) {
        try {
          const response: JsonDataObj = await resourceApi.deleteResource(
            resource.resourceId,
            resource.tenantId,
          )
          if (response.oK && response.state) {
            successCount++
          } else {
            failCount++
          }
        } catch {
          failCount++
        }
      }

      if (successCount > 0) {
        message.success(
          failCount > 0
            ? t('resource.message.batchDeleteResultWithFail', {
                success: successCount,
                fail: failCount,
              })
            : t('resource.message.batchDeleteResult', { success: successCount }),
        )
        removeResourcesFromList(resources.slice(0, successCount))
        await loadResources()
        return true
      } else {
        message.error(t('resource.message.batchDeleteAllFailed', { fail: failCount }))
        return false
      }
    } catch (error) {
      message.error(t('resource.message.batchDeleteFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    model,
    loadResources,
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
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
