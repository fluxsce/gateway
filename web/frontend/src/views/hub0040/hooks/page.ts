import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { ServiceGroup } from '../types'
import { useNamespaceService } from './service'

/**
 * 命名空间管理页面级 Hook
 * - 组合 useNamespaceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useNamespacePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useNamespaceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditNamespace = ref<ServiceGroup | null>(null)

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /** 打开新增命名空间对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditNamespace.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑命名空间对话框 */
  const openEditDialog = (namespace: ServiceGroup) => {
    formDialogMode.value = 'edit'
    currentEditNamespace.value = namespace
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditNamespace.value = null
  }
  
  /** 打开查看详情对话框 */
  const openViewDialog = (namespace: ServiceGroup) => {
    formDialogMode.value = 'view'
    currentEditNamespace.value = namespace
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
          message.warning('请先点击选择要编辑的命名空间')
          return
        }
        openEditDialog(currentRow as ServiceGroup)
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
          message.warning('请先点击选择要删除的命名空间')
          return
        }
        await service.deleteNamespace(currentRow as ServiceGroup)
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
      const success = await service.createNamespace(formData as Partial<ServiceGroup>)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditNamespace.value) return
      // 合并当前命名空间ID和租户ID，确保更新的是正确的记录
      const updatedNamespace = {
        ...currentEditNamespace.value,
        ...formData
      } as ServiceGroup
      const success = await service.updateNamespace(updatedNamespace)
      if (success) {
        closeFormDialog()
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: ServiceGroup }) => {
    if (!row) return

    switch (code) {
      case 'view':
        openViewDialog(row)
        break

      case 'edit':
        openEditDialog(row)
        break

      case 'delete':
        await service.deleteNamespace(row)
        break
    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditNamespace,
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

export type NamespacePage = ReturnType<typeof useNamespacePage>

