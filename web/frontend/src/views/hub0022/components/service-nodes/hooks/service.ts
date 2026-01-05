/**
 * 服务节点管理列表业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
  addServiceNode,
  batchDeleteServiceNodes,
  deleteServiceNode,
  editServiceNode,
  queryServiceNodes,
} from '../../../api'
import type { ServiceNode } from '../types'
import { useServiceNodeModel } from './model'

/**
 * 服务节点管理服务 Hook（纯业务逻辑）
 * @param serviceDefinitionId 服务定义ID（必须提供，用于查询和新增）
 */
export function useServiceNodeService(
  serviceDefinitionId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useServiceNodeModel()

  const {
    loading,
    nodeList,
    pageInfo,
    setNodeList,
    updatePagination,
    addNodeToList,
    updateNodeInList,
    removeNodeFromList,
    removeNodesFromList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载服务节点列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadNodeList = async (searchParams?: Record<string, any>) => {
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

      // 必须携带 serviceDefinitionId
      const finalServiceDefinitionId = serviceDefinitionId?.value || effectiveSearchParams.serviceDefinitionId
      if (!finalServiceDefinitionId) {
        message.error('serviceDefinitionId不能为空')
        return
      }

      // 构建请求参数：合并查询条件和分页参数
      const params: Record<string, any> = {
        // 查询条件
        serviceDefinitionId: finalServiceDefinitionId,
        ...effectiveSearchParams,
        // 分页参数（直接使用 pageIndex/pageSize）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        ),
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await queryServiceNodes(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const nodes = parseJsonData<ServiceNode[]>(response, [])
        setNodeList(Array.isArray(nodes) ? nodes : [])

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(getApiMessage(response, '查询服务节点列表失败'))
      }
    } catch (error) {
      message.error('加载服务节点列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索服务节点
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    model.resetPagination()
    await loadNodeList(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadNodeList()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadNodeList()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadNodeList()
  }

  // ============= 增删改 =============

  /**
   * 添加服务节点
   */
  const addNode = async (nodeData: Partial<ServiceNode> & { serviceDefinitionId: string }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await addServiceNode(nodeData as any)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '新增服务节点成功'))
        await loadNodeList()
        return true
      } else {
        message.error(getApiMessage(response, '新增服务节点失败'))
        return false
      }
    } catch (error) {
      message.error('新增服务节点失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑服务节点
   */
  const editNode = async (
    nodeData: Partial<ServiceNode> & { serviceNodeId: string; serviceDefinitionId: string }
  ): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await editServiceNode(nodeData as any)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '编辑服务节点成功'))
        await loadNodeList()
        return true
      } else {
        message.error(getApiMessage(response, '编辑服务节点失败'))
        return false
      }
    } catch (error) {
      message.error('编辑服务节点失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除服务节点
   */
  const deleteNode = async (serviceNodeId: string): Promise<boolean> => {
    // 查找要删除的节点信息，用于确认对话框
    const nodeToDelete = nodeList.value.find((node) => node.serviceNodeId === serviceNodeId)
    if (!nodeToDelete) {
      message.error('未找到要删除的服务节点')
      return false
    }

    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除服务节点 "${nodeToDelete.nodeUrl || nodeToDelete.nodeHost || serviceNodeId}" 吗？`,
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
      const response: JsonDataObj = await deleteServiceNode(serviceNodeId, 'default')

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '删除服务节点成功'))
        removeNodeFromList(serviceNodeId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (nodeList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadNodeList()
        } else {
          await loadNodeList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除服务节点失败'))
        return false
      }
    } catch (error) {
      message.error('删除服务节点失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除服务节点
   */
  const batchDeleteNodes = async (serviceNodeIds: string[]): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${serviceNodeIds.length} 个服务节点吗？`,
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
      const response: JsonDataObj = await batchDeleteServiceNodes(serviceNodeIds)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '批量删除服务节点成功'))
        removeNodesFromList(serviceNodeIds)
        await loadNodeList()
        return true
      } else {
        message.error(getApiMessage(response, '批量删除服务节点失败'))
        return false
      }
    } catch (error) {
      message.error('批量删除服务节点失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model
    model,

    // 数据加载
    loadNodeList,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,

    // 增删改
    addNode,
    editNode,
    deleteNode,
    batchDeleteNodes,
  }
}

/**
 * 服务节点管理服务类型
 */
export type ServiceNodeService = ReturnType<typeof useServiceNodeService>

