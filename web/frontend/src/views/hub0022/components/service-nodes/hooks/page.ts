/**
 * 服务节点管理列表页面级 Hook
 * - 组合 useServiceNodeService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import { getServiceNode } from '../../../api'
import type { ServiceNode } from '../types'
import { NodeStatus } from '../types'
import { useServiceNodeService } from './service'

/**
 * 服务节点管理列表页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param serviceDefinitionId 服务定义ID（必须提供，用于查询和新增）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useServiceNodePage(
  gridRef?: Ref<any> | any,
  serviceDefinitionId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useServiceNodeService(serviceDefinitionId, searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditNode = ref<ServiceNode | null>(null)

  /**
   * 将节点对象转换为表单数据格式
   */
  const convertToFormData = (node: ServiceNode): any => {
    // 安全解析 JSON 字符串
    const parseJson = (value: any): any => {
      if (typeof value === 'string' && value) {
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      }
      return value
    }

    const formData: any = {
      ...node,
    }

    // 解析 nodeMetadata（JSON 字符串 -> 格式化的 JSON 字符串，用于 textarea 显示）
    if (node.nodeMetadata) {
      const nodeMetadataObj = parseJson(node.nodeMetadata)
      if (nodeMetadataObj && typeof nodeMetadataObj === 'object') {
        try {
          formData.nodeMetadata = JSON.stringify(nodeMetadataObj, null, 2)
        } catch {
          formData.nodeMetadata = '{}'
        }
      } else {
        formData.nodeMetadata = node.nodeMetadata || '{}'
      }
    } else {
      formData.nodeMetadata = '{}'
    }

    return formData
  }

  /**
   * 将表单数据转换为API数据格式
   */
  const convertToApiData = (formData: Record<string, any>): any => {
    const apiData: Record<string, any> = {
      ...formData,
    }

    // 将 nodeMetadata 转换为 JSON 字符串
    if (formData.nodeMetadata) {
      if (typeof formData.nodeMetadata === 'string') {
        // 如果是字符串，验证是否为有效 JSON，然后直接使用
        try {
          JSON.parse(formData.nodeMetadata)
          apiData.nodeMetadata = formData.nodeMetadata
        } catch {
          // JSON 格式无效，使用空对象
          apiData.nodeMetadata = '{}'
        }
      } else if (typeof formData.nodeMetadata === 'object') {
        // 如果是对象，转换为 JSON 字符串
        apiData.nodeMetadata = JSON.stringify(formData.nodeMetadata)
      } else {
        apiData.nodeMetadata = '{}'
      }
    } else {
      apiData.nodeMetadata = '{}'
    }

    // 确保数值字段为数字类型
    if (formData.nodePort !== undefined) {
      apiData.nodePort = Number(formData.nodePort)
    }
    if (formData.nodeWeight !== undefined) {
      apiData.nodeWeight = Number(formData.nodeWeight)
    }
    if (formData.nodeStatus !== undefined) {
      apiData.nodeStatus = Number(formData.nodeStatus)
    }

    return apiData
  }

  /** 打开新增节点对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditNode.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑节点对话框 */
  const openEditDialog = async (node: ServiceNode) => {
    // 如果有 serviceNodeId，通过主键获取最新数据
    if (node.serviceNodeId) {
      try {
        const response = await getServiceNode(node.serviceNodeId, 'default')
        if (isApiSuccess(response)) {
          const latestNode = parseJsonData<ServiceNode | null>(response, null)
          if (latestNode) {
            formDialogMode.value = 'edit'
            currentEditNode.value = convertToFormData(latestNode)
            formDialogVisible.value = true
            return
          }
        }
      } catch (error) {
        // 获取失败时降级使用传入的节点数据
      }
    }
    // 降级：使用传入的节点数据
    formDialogMode.value = 'edit'
    currentEditNode.value = convertToFormData(node)
    formDialogVisible.value = true
  }

  /** 打开查看详情对话框 */
  const openViewDialog = async (node: ServiceNode) => {
    // 如果有 serviceNodeId，通过主键获取最新数据
    if (node.serviceNodeId) {
      try {
        const response = await getServiceNode(node.serviceNodeId, 'default')
        if (isApiSuccess(response)) {
          const latestNode = parseJsonData<ServiceNode | null>(response, null)
          if (latestNode) {
            formDialogMode.value = 'view'
            currentEditNode.value = convertToFormData(latestNode)
            formDialogVisible.value = true
            return
          }
        }
      } catch (error) {
        // 获取失败时降级使用传入的节点数据
      }
    }
    // 降级：使用传入的节点数据
    formDialogMode.value = 'view'
    currentEditNode.value = convertToFormData(node)
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditNode.value = null
  }

  /**
   * 获取表单初始数据
   */
  const getFormInitialData = (): any => {
    if (formDialogMode.value === 'create') {
      // 新增模式：使用默认值
      return {
        serviceDefinitionId: serviceDefinitionId?.value || '',
        nodeUrl: 'http://192.168.1.0:8080',
        nodeHost: '192.168.1.0',
        nodePort: 8080,
        nodeProtocol: 'HTTP',
        nodeWeight: 100,
        healthStatus: 'Y',
        activeFlag: 'Y',
        nodeStatus: NodeStatus.ONLINE,
        nodeMetadata: '{}',
        noteText: '',
      }
    } else if (currentEditNode.value) {
      // 编辑/查看模式：使用转换后的数据
      return currentEditNode.value
    }
    return undefined
  }

  /**
   * 提交表单（新增/编辑共用，由 GdataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return false

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return false
    }

    try {
      // 使用 convertToApiData 转换数据
      const apiData = convertToApiData(formData)

      // 确保 serviceDefinitionId 存在
      if (!apiData.serviceDefinitionId && serviceDefinitionId?.value) {
        apiData.serviceDefinitionId = serviceDefinitionId.value
      }

      if (!apiData.serviceDefinitionId) {
        message.error('服务定义ID不能为空')
        return false
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addNode(apiData as any)
      } else if (formDialogMode.value === 'edit' && currentEditNode.value?.serviceNodeId) {
        apiData.serviceNodeId = currentEditNode.value.serviceNodeId
        success = await service.editNode(apiData as any)
      }

      if (success) {
        closeFormDialog()
      }
      return success
    } catch (error) {
      message.error('操作失败')
      return false
    }
  }

  /**
   * 处理工具栏点击
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'edit': {
        // 编辑当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord?.()
        if (!currentRow) {
          message.warning('请先点击选择要编辑的服务节点')
          return
        }
        await openEditDialog(currentRow as ServiceNode)
        break
      }

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord?.()
        if (!currentRow) {
          message.warning('请先点击选择要删除的服务节点')
          return
        }
        await service.deleteNode((currentRow as ServiceNode).serviceNodeId)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        await handleSearch(formData)
        break
      }

      default:
        break
    }
  }

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: ServiceNode }) => {
    if (!row) return

    switch (code) {
      case 'edit':
        await openEditDialog(row)
        break
      case 'delete':
        await service.deleteNode(row.serviceNodeId)
        break
      default:
        break
    }
  }

  /**
   * 处理搜索
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  return {
    // Service
    service,

    // 表单对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditNode,
    getFormInitialData,

    // 方法
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

/**
 * 服务节点管理页面类型
 */
export type ServiceNodePage = ReturnType<typeof useServiceNodePage>

