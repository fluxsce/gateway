import { useMessage } from 'naive-ui'
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
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useTunnelServerService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditServer = ref<TunnelServer | null>(null)

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
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
    // 获取最新数据填充到编辑对话框
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
        // 打开新增对话框
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
          message.warning('请先点击选择要编辑的隧道服务器')
          return
        }
        await openEditDialog(currentRow as TunnelServer)
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
          message.warning('请先点击选择要删除的隧道服务器')
          return
        }
        await service.deleteTunnelServer(currentRow as TunnelServer)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        await service.handleSearch(formData)
        break
      }
    }
  }

  /**
   * 提交表单（新增/编辑共用，由 GdataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    if (formDialogMode.value === 'create') {
      // 新增模式
      const success = await service.createTunnelServer(formData as Partial<TunnelServer>)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditServer.value) return
      // 合并当前服务器ID和租户ID，确保更新的是正确的记录
      const updatedServer = {
        ...currentEditServer.value,
        ...formData
      } as TunnelServer
      const success = await service.updateTunnelServer(updatedServer)
      if (success) {
        closeFormDialog()
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: TunnelServer }) => {
    if (!row) return

    switch (code) {
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
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditServer,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

export type TunnelServerPage = ReturnType<typeof useTunnelServerPage>

