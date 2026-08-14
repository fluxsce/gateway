import { useAppMessage } from '@/composables/useAppMessage'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { TunnelServer } from '../../../types'
import { useTunnelServerService } from './service'

/**
 * 隧道服务器管理页面级 Hook
 * - 组合 useTunnelServerService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useTunnelServerPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useTunnelServerService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditServer = ref<TunnelServer | null>(null)
  const submitting = ref(false)

  /**
   * 处理搜索（接收 RsSearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /** 打开新增隧道服务器对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditServer.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑隧道服务器对话框 */
  const openEditDialog = async (server: TunnelServer) => {
    const latestServer = await service.viewTunnelServer(server)
    if (latestServer) {
      formDialogMode.value = 'edit'
      currentEditServer.value = latestServer
      formDialogVisible.value = true
    }
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditServer.value = null
  }

  /** 打开查看详情对话框 */
  const openViewDialog = (server: TunnelServer) => {
    formDialogMode.value = 'view'
    currentEditServer.value = server
    formDialogVisible.value = true
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'edit': {
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRow = gridRef.value.getActiveRow?.()
        if (!selectedRow) {
          message.warning('请先选择或点击要编辑的隧道服务器')
          return
        }
        await openEditDialog(selectedRow as TunnelServer)
        break
      }

      case 'delete': {
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const selectedRow = gridRef.value.getActiveRow?.()
        if (!selectedRow) {
          message.warning('请先选择或点击要删除的隧道服务器')
          return
        }
        await service.deleteTunnelServer(selectedRow as TunnelServer)
        break
      }

      case 'search': {
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 提交表单（新增/编辑共用，由 RsDataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    if (formDialogMode.value === 'view') {
      return
    }

    submitting.value = true
    try {
      if (formDialogMode.value === 'create') {
        const success = await service.createTunnelServer(formData as Partial<TunnelServer>)
        if (success) {
          closeFormDialog()
        }
      } else if (formDialogMode.value === 'edit') {
        if (!currentEditServer.value) return
        const updatedServer = {
          ...currentEditServer.value,
          ...formData,
        } as TunnelServer
        const success = await service.updateTunnelServer(updatedServer)
        if (success) {
          closeFormDialog()
        }
      }
    } finally {
      submitting.value = false
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: TunnelServer }) => {
    if (!row) return

    switch (key) {
      case 'view':
        openViewDialog(row)
        break

      case 'edit':
        await openEditDialog(row)
        break

      case 'delete':
        await service.deleteTunnelServer(row)
        break

      case 'start':
        await service.startTunnelServer(row)
        break

      case 'stop':
        await service.stopTunnelServer(row)
        break

      case 'restart':
        await service.restartTunnelServer(row)
        break
    }
  }

  return {
    service,

    formDialogVisible,
    formDialogMode,
    currentEditServer,
    submitting,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

export type TunnelServerPage = ReturnType<typeof useTunnelServerPage>
