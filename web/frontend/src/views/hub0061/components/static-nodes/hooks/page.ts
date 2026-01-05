/**
 * 静态节点列表页面级 Hook
 * - 组合 useStaticNodeService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import { useStaticNodeService } from './service'
import type { TunnelStaticNode } from './types'

/**
 * 静态节点列表页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param tunnelStaticServerId 静态服务器ID（必需，用于关联节点）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useStaticNodePage(
  gridRef?: Ref<any> | any,
  tunnelStaticServerId?: Ref<string> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useStaticNodeService(tunnelStaticServerId || '', searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditNode = ref<TunnelStaticNode | null>(null)

  // ============= 统一校验 =============

  /**
   * 校验是否已选择服务器
   */
  const validateContext = (showMessage = true): boolean => {
    const serverId = typeof tunnelStaticServerId === 'string' ? tunnelStaticServerId : tunnelStaticServerId?.value

    if (!serverId) {
      if (showMessage) {
        message.warning('请先选择静态服务器')
      }
      return false
    }
    return true
  }

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadNodeList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadNodeList()
  }

  // ============= 工具栏按钮处理 =============

  /**
   * 处理工具栏按钮点击
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要删除的节点')
          return
        }
        await handleDelete(currentRow as TunnelStaticNode)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        await handleSearch(formData)
        break
      }

      default:
        console.warn('未知的工具栏按钮:', key)
    }
  }

  // ============= 对话框处理 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    if (!validateContext()) {
      return
    }
    formDialogMode.value = 'create'
    currentEditNode.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (node: TunnelStaticNode) => {
    if (!node.tunnelStaticNodeId) {
      message.warning('节点ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestNode = await service.getNodeDetail(node.tunnelStaticNodeId)
      if (latestNode) {
        formDialogMode.value = 'edit'
        currentEditNode.value = latestNode
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'edit'
        currentEditNode.value = node
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'edit'
      currentEditNode.value = node
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (node: TunnelStaticNode) => {
    if (!node.tunnelStaticNodeId) {
      message.warning('节点ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestNode = await service.getNodeDetail(node.tunnelStaticNodeId)
      if (latestNode) {
        formDialogMode.value = 'view'
        currentEditNode.value = latestNode
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'view'
        currentEditNode.value = node
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'view'
      currentEditNode.value = node
      formDialogVisible.value = true
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditNode.value = null
  }

  /**
   * 处理表单提交
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      closeFormDialog()
      return
    }

    // 校验是否已选择服务器
    if (!validateContext()) {
      return false
    }

    try {
      const submitData: Partial<TunnelStaticNode> = {
        tunnelStaticNodeId: formData.tunnelStaticNodeId,
        tunnelStaticServerId: formData.tunnelStaticServerId || (typeof tunnelStaticServerId === 'string' ? tunnelStaticServerId : tunnelStaticServerId?.value),
        nodeName: formData.nodeName,
        nodeDescription: formData.nodeDescription,
        targetAddress: formData.targetAddress,
        targetPort: formData.targetPort,
        proxyType: formData.proxyType,
        maxConnections: formData.maxConnections,
        connectionTimeout: formData.connectionTimeout,
        readTimeout: formData.readTimeout,
        writeTimeout: formData.writeTimeout,
        retryCount: formData.retryCount,
        retryInterval: formData.retryInterval,
        compression: formData.compression,
        encryption: formData.encryption,
        secretKey: formData.secretKey,
        activeFlag: formData.activeFlag,
        noteText: formData.noteText,
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addNode(submitData)
      } else if (formDialogMode.value === 'edit' && currentEditNode.value) {
        success = await service.editNode(
          currentEditNode.value.tunnelStaticNodeId,
          submitData
        )
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error: any) {
      console.error('提交节点配置失败:', error)
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ menu, row }: { menu: any; row: TunnelStaticNode }) => {
    switch (menu.code) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'toggle-status':
        await handleToggleStatus(row)
        break
      case 'delete':
        await handleDelete(row)
        break
      default:
        console.warn('未知的菜单项:', menu.code)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (node: TunnelStaticNode) => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      content: `确定要删除节点"${node.nodeName}"吗？此操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
    })

    if (!confirmed) {
      return
    }

    await service.deleteNode(node.tunnelStaticNodeId)
  }

  /**
   * 处理切换状态
   */
  const handleToggleStatus = async (node: TunnelStaticNode) => {
    await service.toggleNodeStatus(node)
  }

  return {
    // 服务
    service,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditNode,

    // 方法
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,
    handleDelete,
    handleToggleStatus,
  }
}

