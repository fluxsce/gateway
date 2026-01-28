/**
 * 命名空间管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as namespaceApi from '../api'
import type { Namespace } from '../types'
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
    removeNamespaceFromList,
    removeNamespacesFromList,
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
      const params = {
        ...effectiveSearchParams,
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API
      const response: JsonDataObj = await namespaceApi.queryNamespaces(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const namespaces = Array.isArray(bizData) ? bizData : []
          setNamespaceList(namespaces)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询命名空间列表失败')
      }
    } catch (error) {
      message.error('加载命名空间列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  const handleSearch = async (searchParams?: Record<string, any>) => {
    await loadNamespaces(searchParams)
  }

  const handleReset = async () => {
    model.resetPagination()
    await loadNamespaces()
  }

  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadNamespaces()
  }

  const handleRefresh = async () => {
    await loadNamespaces()
  }

  // ============= 增删改 =============

  const addNamespace = async (namespaceData: Namespace): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await namespaceApi.addNamespace(namespaceData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '命名空间创建成功')
        message.success(successMsg)
        
        if (response.bizData) {
          const newNamespace = JSON.parse(response.bizData)
          addNamespaceToList(newNamespace)
        } else {
          await loadNamespaces()
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '新增命名空间失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('新增命名空间失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const editNamespace = async (namespaceData: Partial<Namespace> & {
    namespaceId: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await namespaceApi.editNamespace(namespaceData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '命名空间更新成功')
        message.success(successMsg)
        
        if (response.bizData) {
          const updatedNamespace = JSON.parse(response.bizData)
          updateNamespaceInList(
            updatedNamespace.namespaceId,
            updatedNamespace.tenantId,
            updatedNamespace
          )
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '编辑命名空间失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('编辑命名空间失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteNamespace = async (namespace: Namespace): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除命名空间 "${namespace.namespaceName}" (${namespace.namespaceId}) 吗？`,
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
      const response: JsonDataObj = await namespaceApi.deleteNamespace(namespace.namespaceId)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '命名空间删除成功')
        message.success(successMsg)
        removeNamespaceFromList(namespace.namespaceId, namespace.tenantId)
        return true
      } else {
        const errorMsg = getApiMessage(response, '删除命名空间失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('删除命名空间失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const getNamespaceDetail = async (namespaceId: string): Promise<Namespace | null> => {
    loading.value = true
    try {
      const response: JsonDataObj = await namespaceApi.getNamespace(namespaceId)

      if (isApiSuccess(response) && response.bizData) {
        return JSON.parse(response.bizData)
      } else {
        message.error(getApiMessage(response, '获取命名空间详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取命名空间详情失败')
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadNamespaces,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,

    // 命名空间操作
    addNamespace,
    editNamespace,
    deleteNamespace,
    getNamespaceDetail,
  }
}

export type NamespaceService = ReturnType<typeof useNamespaceService>

