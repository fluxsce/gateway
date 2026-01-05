/**
 * 隧道客户端管理页面级 Hook
 * - 组合 useTunnelClientService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { TunnelClient } from '../../../types'
import { useTunnelClientService } from './service'

/**
 * 隧道客户端管理页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useTunnelClientPage(
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

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
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadClientList()
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

      case 'edit': {
        // 编辑当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要编辑的客户端')
          return
        }
        await openEditDialog(currentRow as TunnelClient)
        break
      }

      case 'connect': {
        // 连接当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要连接的客户端')
          return
        }
        await handleConnect(currentRow as TunnelClient)
        break
      }

      case 'disconnect': {
        // 断开当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要断开的客户端')
          return
        }
        await handleDisconnect(currentRow as TunnelClient)
        break
      }

      case 'delete': {
        // 删除当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要删除的客户端')
          return
        }
        await handleDelete(currentRow as TunnelClient)
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
        success = await service.editClient(
          currentEditClient.value.tunnelClientId,
          submitData
        )
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error: any) {
      console.error('提交客户端配置失败:', error)
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ menu, row }: { menu: any; row: TunnelClient }) => {
    const code = menu?.code || menu
    switch (code) {
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
        console.warn('未知的菜单项:', code)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理连接
   */
  const handleConnect = async (client: TunnelClient) => {
    await service.connectClient(client)
  }

  /**
   * 处理断开连接
   */
  const handleDisconnect = async (client: TunnelClient) => {
    await service.disconnectClient(client)
  }

  /**
   * 处理删除
   */
  const handleDelete = async (client: TunnelClient) => {
    await service.deleteClient(client)
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
  }
}

/**
 * 隧道客户端管理 Page 类型
 */
export type TunnelClientPage = ReturnType<typeof useTunnelClientPage>

