import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { Resource } from '../types'
import { useResourceService } from './useResourceService'

/**
 * 权限资源管理页面级 Hook
 * - 组合 useResourceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useResourcePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useResourceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditResource = ref<Resource | null>(null)

  /** 打开新增资源对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditResource.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑资源对话框 */
  const openEditDialog = (resource: Resource) => {
    formDialogMode.value = 'edit'
    currentEditResource.value = resource
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditResource.value = null
  }
  
  /** 打开查看详情对话框 */
  const openViewDialog = (resource: Resource) => {
    formDialogMode.value = 'view'
    currentEditResource.value = resource
    formDialogVisible.value = true
  }

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /** 提交表单（新增/编辑共用，由 GdataFormModal 收集表单数据后回调） */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    if (formDialogMode.value === 'create') {
      // 新增模式
      const success = await service.addResource(formData as Resource)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditResource.value) return
      // 合并当前资源ID和租户ID，确保更新的是正确的记录
      const updatedResource = {
        ...currentEditResource.value,
        ...formData
      } as Resource
      const success = await service.editResource(updatedResource)
      if (success) {
        closeFormDialog()
      }
    }
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'view': {
        // 查看当前高亮的行（点击选中的行）
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要查看的资源')
          return
        }
        openViewDialog(currentRow as Resource)
        break
      }

      case 'add':
        // 直接打开新增对话框
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
          message.warning('请先点击选择要编辑的资源')
          return
        }
        openEditDialog(currentRow as Resource)
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
          message.warning('请先点击选择要删除的资源')
          return
        }
        await service.deleteResource(currentRow as Resource)
        break
      }

      case 'search': {
        // search 操作由 @search 事件处理，这里不需要重复处理
        // 避免重复请求（SearchForm 会同时触发 @search 和 @toolbar-click 事件）
        break
      }
    }
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: Resource }) => {
    if (!row) return

    switch (code) {
      case 'view':
        openViewDialog(row)
        break

      case 'edit':
        openEditDialog(row)
        break

      case 'delete':
        await service.deleteResource(row)
        break
    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditResource,
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

export type ResourcePage = ReturnType<typeof useResourcePage>

