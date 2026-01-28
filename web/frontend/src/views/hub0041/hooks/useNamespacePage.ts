/**
 * 命名空间管理页面级 Hook
 * - 组合 useNamespaceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { Namespace } from '../types'
import { useNamespaceService } from './useNamespaceService'

/**
 * 命名空间管理页面级 Hook
 */
export function useNamespacePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useNamespaceService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditNamespace = ref<Namespace | null>(null)
  const submitting = ref(false)

  /**
   * 打开新增命名空间对话框
   */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditNamespace.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑命名空间对话框
   */
  const openEditDialog = async (namespace: Namespace) => {
    try {
      const detailNamespace = await service.getNamespaceDetail(namespace.namespaceId)
      
      if (!detailNamespace) {
        message.error('获取命名空间详情失败')
        return
      }

      formDialogMode.value = 'edit'
      currentEditNamespace.value = detailNamespace
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取命名空间详情失败')
    }
  }

  /**
   * 打开查看命名空间对话框
   */
  const openViewDialog = async (namespace: Namespace) => {
    try {
      const detailNamespace = await service.getNamespaceDetail(namespace.namespaceId)
      
      if (!detailNamespace) {
        message.error('获取命名空间详情失败')
        return
      }

      formDialogMode.value = 'view'
      currentEditNamespace.value = detailNamespace
      formDialogVisible.value = true
    } catch (error) {
      message.error('获取命名空间详情失败')
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditNamespace.value = null
  }

  /**
   * 提交表单（新增/编辑）
   */
  const handleSubmit = async (formData: Namespace) => {
    submitting.value = true
    try {
      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addNamespace(formData)
      } else if (formDialogMode.value === 'edit') {
        success = await service.editNamespace({
          ...formData,
          namespaceId: currentEditNamespace.value!.namespaceId,
        })
      }

      if (success) {
        closeFormDialog()
        await service.handleRefresh()
      }
    } finally {
      submitting.value = false
    }
  }

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = (key: string) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'edit':
        const selectedRows = gridRef?.value?.getCheckboxRecords() || []
        if (selectedRows.length === 0) {
          message.warning('请先选择要编辑的命名空间')
          return
        }
        if (selectedRows.length > 1) {
          message.warning('只能编辑一个命名空间')
          return
        }
        openEditDialog(selectedRows[0])
        break
      case 'delete':
        const deleteRows = gridRef?.value?.getCheckboxRecords() || []
        if (deleteRows.length === 0) {
          message.warning('请先选择要删除的命名空间')
          return
        }
        handleBatchDelete(deleteRows)
        break
    }
  }

  /**
   * 批量删除命名空间
   */
  const handleBatchDelete = async (namespaces: Namespace[]) => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: `将删除 ${namespaces.length} 个命名空间`,
      content: '此操作不可恢复，请谨慎操作',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return
    }

    let successCount = 0
    for (const namespace of namespaces) {
      const success = await service.deleteNamespace(namespace)
      if (success) {
        successCount++
      }
    }

    if (successCount > 0) {
      message.success(`成功删除 ${successCount} 个命名空间`)
      await service.handleRefresh()
    }
  }

  /**
   * 处理表格右键菜单
   */
  const handleMenuClick = async (params: { code: string; row?: any }) => {
    if (!params.row) {
      return
    }
    const row = params.row as Namespace
    switch (params.code) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'delete':
        await service.deleteNamespace(row)
        await service.handleRefresh()
        break
    }
  }

  /**
   * 搜索处理
   */
  const handleSearch = () => {
    service.handleSearch()
  }

  /**
   * 表单提交处理（适配 GdataFormModal 的提交格式）
   */
  const handleFormSubmit = (formData?: Record<string, any>) => {
    if (formData) {
      handleSubmit(formData as any)
    }
  }

  return {
    // 服务（包含 model 和所有业务方法）
    service,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditNamespace,
    submitting,

    // 对话框方法
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleSubmit,
    handleFormSubmit,

    // 工具栏和菜单
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

export type NamespacePage = ReturnType<typeof useNamespacePage>

