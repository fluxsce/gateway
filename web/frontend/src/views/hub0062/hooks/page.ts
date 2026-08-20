/**
 * 隧道客户端管理页面级 Hook
 * - 组合 useTunnelClientService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { rsConfirm } from '@/ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { TunnelClient } from '../types'
import { useTunnelClientService } from './service'

/**
 * 隧道客户端管理页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useTunnelClientPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  // 业务服务（包含 model、增删改查等）
  const service = useTunnelClientService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditClient = ref<TunnelClient | null>(null)

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadClientList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({
    currentPage,
    pageSize,
  }: {
    currentPage: number
    pageSize: number
  }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadClientList()
  }

  // ============= 工具栏按钮处理 =============

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = async (key: string) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'connect':
        await handleBatchConnect()
        break
      case 'disconnect':
        await handleBatchDisconnect()
        break
      case 'delete':
        await handleBatchDelete()
        break
      default:
        break
    }
  }

  // ============= 对话框处理 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditClient.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (client: TunnelClient) => {
    if (!client.tunnelClientId) {
      message.warning('客户端ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestClient = await service.getClientDetail(client.tunnelClientId)
      if (latestClient) {
        formDialogMode.value = 'edit'
        currentEditClient.value = latestClient
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'edit'
        currentEditClient.value = client
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'edit'
      currentEditClient.value = client
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (client: TunnelClient) => {
    if (!client.tunnelClientId) {
      message.warning('客户端ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestClient = await service.getClientDetail(client.tunnelClientId)
      if (latestClient) {
        formDialogMode.value = 'view'
        currentEditClient.value = latestClient
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'view'
        currentEditClient.value = client
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'view'
      currentEditClient.value = client
      formDialogVisible.value = true
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditClient.value = null
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

    try {
      const submitData: Partial<TunnelClient> = {
        tunnelClientId: formData.tunnelClientId,
        clientName: formData.clientName,
        clientDescription: formData.clientDescription,
        serverAddress: formData.serverAddress,
        serverPort: formData.serverPort,
        authToken: formData.authToken,
        tlsEnable: formData.tlsEnable,
        autoReconnect: formData.autoReconnect,
        maxRetries: formData.maxRetries,
        retryInterval: formData.retryInterval,
        heartbeatInterval: formData.heartbeatInterval,
        heartbeatTimeout: formData.heartbeatTimeout,
        activeFlag: formData.activeFlag,
        noteText: formData.noteText,
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addClient(submitData)
      } else if (formDialogMode.value === 'edit' && currentEditClient.value) {
        success = await service.editClient(currentEditClient.value.tunnelClientId, submitData)
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error: any) {
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: TunnelClient }) => {
    if (!row) return
    switch (key) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'connect':
        await handleConnect(row)
        break
      case 'disconnect':
        await handleDisconnect(row)
        break
      case 'delete':
        await handleDelete(row)
        break
      default:
        break
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理连接
   */
  const handleConnect = async (client: TunnelClient) => {
    const confirmed = await rsConfirm.confirm({
      title: '确认连接',
      description: `确定要连接客户端"${client.clientName}"吗？`,
      confirmText: '连接',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    await service.connectClient(client.tunnelClientId)
  }

  /**
   * 处理断开连接
   */
  const handleDisconnect = async (client: TunnelClient) => {
    const confirmed = await rsConfirm.warning({
      title: '确认断开',
      description: `确定要断开客户端"${client.clientName}"的连接吗？`,
      confirmText: '断开',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    await service.disconnectClient(client.tunnelClientId)
  }

  /**
   * 处理删除
   */
  const handleDelete = async (client: TunnelClient) => {
    const confirmed = await rsConfirm.warning({
      title: '确认删除',
      description: `确定要删除客户端"${client.clientName}"吗？此操作不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    await service.deleteClient(client.tunnelClientId)
  }

  // ============= 批量操作处理 =============

  /**
   * 获取选中的记录
   */
  const getSelectedRecords = (): TunnelClient[] => {
    if (!gridRef?.value?.getActiveRows) {
      return []
    }
    return (gridRef.value.getActiveRows() || []) as TunnelClient[]
  }

  /**
   * 处理批量连接
   */
  const handleBatchConnect = async () => {
    const selectedRows = getSelectedRecords()
    if (!selectedRows || selectedRows.length === 0) {
      message.warning('请先选择要连接的客户端')
      return
    }

    const confirmed = await rsConfirm.confirm({
      title: '确认批量连接',
      description: `确定要连接选中的 ${selectedRows.length} 个客户端吗？`,
      confirmText: '连接',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    // 逐个连接
    for (const client of selectedRows) {
      await service.connectClient(client.tunnelClientId)
    }
  }

  /**
   * 处理批量断开
   */
  const handleBatchDisconnect = async () => {
    const selectedRows = getSelectedRecords()
    if (!selectedRows || selectedRows.length === 0) {
      message.warning('请先选择要断开的客户端')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: '确认批量断开',
      description: `确定要断开选中的 ${selectedRows.length} 个客户端的连接吗？`,
      confirmText: '断开',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    // 逐个断开
    for (const client of selectedRows) {
      await service.disconnectClient(client.tunnelClientId)
    }
  }

  /**
   * 处理批量删除
   */
  const handleBatchDelete = async () => {
    const selectedRows = getSelectedRecords()
    if (!selectedRows || selectedRows.length === 0) {
      message.warning('请先选择要删除的客户端')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: '确认批量删除',
      description: `确定要删除选中的 ${selectedRows.length} 个客户端吗？此操作不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    const tunnelClientIds = selectedRows.map((row) => row.tunnelClientId)
    await service.deleteClients(tunnelClientIds)
  }

  return {
    // 服务
    service,

    // 表单对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditClient,

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
    handleConnect,
    handleDisconnect,
    handleDelete,
    handleBatchConnect,
    handleBatchDisconnect,
    handleBatchDelete,
  }
}

/**
 * 隧道客户端管理 Page 类型
 */
export type TunnelClientPage = ReturnType<typeof useTunnelClientPage>
