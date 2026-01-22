/**
 * 预警模板管理页面级 Hook
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { AlertTemplate } from '../types'
import { useAlertTemplateService } from './service'

export function useAlertTemplatePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  const service = useAlertTemplateService(searchFormRef)

  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditTemplate = ref<AlertTemplate | null>(null)

  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadTemplateList(searchParams)
  }

  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadTemplateList()
  }

  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditTemplate.value = null
    formDialogVisible.value = true
  }

  const openEditDialog = async (row: AlertTemplate) => {
    formDialogMode.value = 'edit'
    // 尽量用后端详情，保证字段完整
    const detail = await service.getTemplateDetail(row.templateName)
    currentEditTemplate.value = detail || row
    formDialogVisible.value = true
  }

  const openViewDialog = async (row: AlertTemplate) => {
    formDialogMode.value = 'view'
    const detail = await service.getTemplateDetail(row.templateName)
    currentEditTemplate.value = detail || row
    formDialogVisible.value = true
  }

  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditTemplate.value = null
  }

  // 保持原始内容：不做 trim / JSON 强校验，避免影响模板内容（空格/换行）编辑体验
  // 同时移除仅用于 UI 展示的字段，避免提交到后端
  const normalizeFormData = (formData: Record<string, any>): Partial<AlertTemplate> => {
    const data: Record<string, any> = { ...formData }
    delete data.__templateHelp
    return data as Partial<AlertTemplate>
  }

  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    if (formDialogMode.value === 'view') {
      closeFormDialog()
      return
    }

    const submitData = normalizeFormData(formData)

    if (formDialogMode.value === 'create') {
      const ok = await service.addTemplate(submitData)
      if (ok) closeFormDialog()
    } else if (formDialogMode.value === 'edit') {
      const name = String(formData.templateName || '').trim()
      if (!name) {
        message.warning('模板名称不能为空')
        return
      }
      const ok = await service.editTemplate(name, submitData)
      if (ok) closeFormDialog()
    }
  }

  const handleDelete = async (row: AlertTemplate) => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      content: `确定要删除模板"${row.templateName}"吗？此操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
    })
    if (!confirmed) return
    await service.removeTemplate(row.templateName)
  }

  const handleToolbarClick = async (key: string) => {
    if (key === 'add') {
      openAddDialog()
      return
    }
    if (key === 'delete') {
      // 批量删除：依赖 gridRef 的选中行
      const rows: AlertTemplate[] = gridRef?.value?.getSelectedRows?.() || []
      if (!rows || rows.length === 0) {
        message.warning('请先选择要删除的模板')
        return
      }
      const confirmed = await gDialog.warning({
        title: '确认删除',
        content: `确定要删除选中的 ${rows.length} 个模板吗？此操作不可恢复。`,
        positiveText: '删除',
        negativeText: '取消',
      })
      if (!confirmed) return
      for (const r of rows) {
        // 串行删除即可，避免并发压测后端
        // eslint-disable-next-line no-await-in-loop
        await service.removeTemplate(r.templateName)
      }
      return
    }
  }

  const handleMenuClick = async ({ menu, row }: { menu: { code: string }; row: AlertTemplate }) => {
    const code = menu?.code
    if (code === 'view') return openViewDialog(row)
    if (code === 'edit') return openEditDialog(row)
    if (code === 'delete') return handleDelete(row)
  }

  return {
    service,
    formDialogVisible,
    formDialogMode,
    currentEditTemplate,
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,
  }
}


