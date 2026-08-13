import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { Ref } from 'vue'
import { computed, ref } from 'vue'
import type { Resource } from '../types'
import { useResourceService } from './useResourceService'

/**
 * 权限资源管理页面级 Hook
 * - 组合 useResourceService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useResourcePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0006')

  const service = useResourceService(searchFormRef)

  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditResource = ref<Resource | null>(null)

  const formDialogTitle = computed(() => {
    if (formDialogMode.value === 'create') return t('resource.dialog.createTitle')
    if (formDialogMode.value === 'edit') return t('resource.dialog.editTitle')
    return t('resource.dialog.viewTitle')
  })

  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditResource.value = null
    formDialogVisible.value = true
  }

  const openEditDialog = (resource: Resource) => {
    formDialogMode.value = 'edit'
    currentEditResource.value = resource
    formDialogVisible.value = true
  }

  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditResource.value = null
  }

  const openViewDialog = (resource: Resource) => {
    formDialogMode.value = 'view'
    currentEditResource.value = resource
    formDialogVisible.value = true
  }

  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    if (formDialogMode.value === 'view') {
      return
    }

    if (formDialogMode.value === 'create') {
      const success = await service.addResource(formData as Resource)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      if (!currentEditResource.value) return
      const updatedResource = {
        ...currentEditResource.value,
        ...formData,
      } as Resource
      const success = await service.editResource(updatedResource)
      if (success) {
        closeFormDialog()
      }
    }
  }

  const handleToolbarClick = async (key: string, _formData?: Record<string, any>) => {
    switch (key) {
      case 'view': {
        if (!gridRef?.value) {
          message.warning(t('resource.message.gridRefMissing'))
          return
        }
        const selectedRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!selectedRow) {
          message.warning(t('resource.message.selectToView'))
          return
        }
        openViewDialog(selectedRow as Resource)
        break
      }

      case 'add':
        openAddDialog()
        break

      case 'edit': {
        if (!gridRef?.value) {
          message.warning(t('resource.message.gridRefMissing'))
          return
        }
        const selectedRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!selectedRow) {
          message.warning(t('resource.message.selectToEdit'))
          return
        }
        openEditDialog(selectedRow as Resource)
        break
      }

      case 'delete': {
        if (!gridRef?.value) {
          message.warning(t('resource.message.gridRefMissing'))
          return
        }
        const selectedRow = gridRef.value.getSelectedOrCurrentRecord()
        if (!selectedRow) {
          message.warning(t('resource.message.selectToDelete'))
          return
        }
        await service.deleteResource(selectedRow as Resource)
        break
      }

      case 'search': {
        break
      }
    }
  }

  const handleMenuClick = async ({ key, row }: { key: string; row?: Resource }) => {
    if (!row) return

    switch (key) {
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
    service,
    formDialogVisible,
    formDialogMode,
    formDialogTitle,
    currentEditResource,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

export type ResourcePage = ReturnType<typeof useResourcePage>
