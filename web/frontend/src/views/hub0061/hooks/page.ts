/**
 * 静态服务管理页面级 Hook
 * - 组合 useStaticServerService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { TunnelStaticServer } from '../types'
import { useStaticServerService } from './service'

/**
 * 静态服务管理页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useStaticServerPage(
  gridRef?: Ref<any> | any,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useStaticServerService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditServer = ref<TunnelStaticServer | null>(null)

  // 节点管理对话框状态
  const nodeDialogVisible = ref(false)
  const currentNodeServer = ref<TunnelStaticServer | null>(null)

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadServerList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadServerList()
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

      case 'start': {
        // 启动当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要启动的服务')
          return
        }
        await handleStart(currentRow as TunnelStaticServer)
        break
      }

      case 'stop': {
        // 停止当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要停止的服务')
          return
        }
        await handleStop(currentRow as TunnelStaticServer)
        break
      }

      case 'reload': {
        // 重载当前高亮的行
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要重载的服务')
          return
        }
        await handleReload(currentRow as TunnelStaticServer)
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
          message.warning('请先点击选择要删除的服务')
          return
        }
        await handleDelete(currentRow as TunnelStaticServer)
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
    currentEditServer.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (server: TunnelStaticServer) => {
    if (!server.tunnelStaticServerId) {
      message.warning('服务ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestServer = await service.getServerDetail(server.tunnelStaticServerId)
      if (latestServer) {
        formDialogMode.value = 'edit'
        currentEditServer.value = latestServer
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'edit'
        currentEditServer.value = server
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'edit'
      currentEditServer.value = server
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (server: TunnelStaticServer) => {
    if (!server.tunnelStaticServerId) {
      message.warning('服务ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestServer = await service.getServerDetail(server.tunnelStaticServerId)
      if (latestServer) {
        formDialogMode.value = 'view'
        currentEditServer.value = latestServer
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'view'
        currentEditServer.value = server
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'view'
      currentEditServer.value = server
      formDialogVisible.value = true
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditServer.value = null
  }

  /**
   * 打开节点管理对话框
   */
  const openNodeDialog = (server: TunnelStaticServer) => {
    currentNodeServer.value = server
    nodeDialogVisible.value = true
  }

  /**
   * 关闭节点管理对话框
   */
  const closeNodeDialog = () => {
    nodeDialogVisible.value = false
    currentNodeServer.value = null
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
      const submitData: Partial<TunnelStaticServer> = {
        tunnelStaticServerId: formData.tunnelStaticServerId,
        serverName: formData.serverName,
        serverDescription: formData.serverDescription,
        listenAddress: formData.listenAddress,
        listenPort: formData.listenPort,
        serverType: formData.serverType,
        maxConnections: formData.maxConnections,
        connectionTimeout: formData.connectionTimeout,
        readTimeout: formData.readTimeout,
        writeTimeout: formData.writeTimeout,
        tlsEnable: formData.tlsEnable,
        tlsCertFile: formData.tlsCertFile,
        tlsKeyFile: formData.tlsKeyFile,
        tlsCaFile: formData.tlsCaFile,
        logLevel: formData.logLevel,
        healthCheckType: formData.healthCheckType,
        healthCheckUrl: formData.healthCheckUrl,
        healthCheckInterval: formData.healthCheckInterval,
        healthCheckTimeout: formData.healthCheckTimeout,
        healthCheckMaxFailures: formData.healthCheckMaxFailures,
        loadBalanceType: formData.loadBalanceType,
        activeFlag: formData.activeFlag,
        noteText: formData.noteText,
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addServer(submitData)
      } else if (formDialogMode.value === 'edit' && currentEditServer.value) {
        success = await service.editServer(
          currentEditServer.value.tunnelStaticServerId,
          submitData
        )
      }

      if (success) {
        closeFormDialog()
      }
    } catch (error: any) {
      console.error('提交服务配置失败:', error)
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ menu, row }: { menu: any; row: TunnelStaticServer }) => {
    const code = menu?.code || menu
    switch (code) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'nodes':
        openNodeDialog(row)
        break
      case 'start':
        await handleStart(row)
        break
      case 'stop':
        await handleStop(row)
        break
      case 'reload':
        await handleReload(row)
        break
      case 'toggle-status':
        await handleToggleStatus(row)
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
   * 处理启动
   */
  const handleStart = async (server: TunnelStaticServer) => {
    await service.startServer(server)
  }

  /**
   * 处理停止
   */
  const handleStop = async (server: TunnelStaticServer) => {
    await service.stopServer(server)
  }

  /**
   * 处理重载
   */
  const handleReload = async (server: TunnelStaticServer) => {
    await service.reloadServer(server)
  }

  /**
   * 处理删除
   */
  const handleDelete = async (server: TunnelStaticServer) => {
    await service.deleteServer(server)
  }

  /**
   * 处理切换状态
   */
  const handleToggleStatus = async (server: TunnelStaticServer) => {
    await service.toggleServerStatus(server)
  }


  return {
    // 服务
    service,

    // 表单对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditServer,

    // 节点管理对话框状态
    nodeDialogVisible,
    currentNodeServer,

    // 方法
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    openNodeDialog,
    closeNodeDialog,
    handleFormSubmit,
    handleStart,
    handleStop,
    handleReload,
    handleDelete,
    handleToggleStatus,
  }
}

/**
 * 静态服务管理 Page 类型
 */
export type StaticServerPage = ReturnType<typeof useStaticServerPage>
