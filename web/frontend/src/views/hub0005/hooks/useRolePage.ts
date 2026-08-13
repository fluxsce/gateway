import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { Ref } from 'vue'
import { computed, ref } from 'vue'
import type { Role } from '../types'
import { useRoleService } from './useRoleService'

/**
 * 角色管理页面级 Hook
 * - 组合 useRoleService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useRolePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0005')

  const service = useRoleService(searchFormRef)

  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditRole = ref<Role | null>(null)

  const roleAuthDrawerVisible = ref(false)
  const roleAuthRoleId = ref<string>('')
  const roleAuthRoleName = ref<string>('')

  const formDialogTitle = computed(() => {
    if (formDialogMode.value === 'create') return t('role.dialog.createTitle')
    if (formDialogMode.value === 'edit') return t('role.dialog.editTitle')
    return t('role.dialog.viewTitle')
  })

  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditRole.value = null
    formDialogVisible.value = true
  }

  const openEditDialog = (role: Role) => {
    formDialogMode.value = 'edit'
    currentEditRole.value = role
    formDialogVisible.value = true
  }

  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditRole.value = null
  }

  const openViewDialog = (role: Role) => {
    formDialogMode.value = 'view'
    currentEditRole.value = role
    formDialogVisible.value = true
  }

  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return
    if (formDialogMode.value === 'view') return

    if (formDialogMode.value === 'create') {
      const success = await service.addRole(formData as Role)
      if (success) closeFormDialog()
      return
    }

    if (formDialogMode.value === 'edit') {
      if (!currentEditRole.value) return
      const updatedRole = {
        ...currentEditRole.value,
        ...formData,
      } as Role
      const success = await service.editRole(updatedRole)
      if (success) closeFormDialog()
    }
  }

  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break

      case 'edit': {
        if (!gridRef?.value) return
        const selectedRow = gridRef.value.getActiveRow()
        if (!selectedRow) {
          message.warning(t('role.message.selectToEdit'))
          return
        }
        openEditDialog(selectedRow as Role)
        break
      }

      case 'delete': {
        if (!gridRef?.value) return
        const selectedRow = gridRef.value.getActiveRow()
        if (!selectedRow) {
          message.warning(t('role.message.selectToDelete'))
          return
        }
        await service.deleteRole(selectedRow as Role)
        break
      }

      case 'search':
        await service.handleSearch(formData)
        break
    }
  }

  const openRoleAuthDrawer = (role: Role) => {
    roleAuthRoleId.value = role.roleId
    roleAuthRoleName.value = role.roleName
    roleAuthDrawerVisible.value = true
  }

  const closeRoleAuthDrawer = () => {
    roleAuthDrawerVisible.value = false
    roleAuthRoleId.value = ''
    roleAuthRoleName.value = ''
  }

  const handleMenuClick = async ({ key, row }: { key: string; row?: Role }) => {
    if (!row) return

    switch (key) {
      case 'view':
        openViewDialog(row)
        break
      case 'edit':
        openEditDialog(row)
        break
      case 'delete':
        await service.deleteRole(row)
        break
      case 'roleAuth':
        openRoleAuthDrawer(row)
        break
    }
  }

  return {
    service,
    formDialogVisible,
    formDialogMode,
    formDialogTitle,
    currentEditRole,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,
    roleAuthDrawerVisible,
    roleAuthRoleId,
    roleAuthRoleName,
    openRoleAuthDrawer,
    closeRoleAuthDrawer,
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
  }
}

export type RolePage = ReturnType<typeof useRolePage>
