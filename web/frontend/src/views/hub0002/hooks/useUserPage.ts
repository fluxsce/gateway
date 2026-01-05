import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { User } from '../types'
import { useUserService } from './useUserService'

/**
 * 用户管理页面级 Hook
 * - 组合 useUserService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useUserPage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useUserService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditUser = ref<User | null>(null)

  /** 打开新增用户对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditUser.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑用户对话框 */
  const openEditDialog = (user: User) => {
    formDialogMode.value = 'edit'
    currentEditUser.value = user
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditUser.value = null
  }
  
  /** 打开查看详情对话框 */
  const openViewDialog = (user: User) => {
    formDialogMode.value = 'view'
    currentEditUser.value = user
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
      const success = await service.addUser(formData as User)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditUser.value) return
      // 合并当前用户ID和租户ID，确保更新的是正确的记录
      const updatedUser = {
        ...currentEditUser.value,
        ...formData
      } as User
      const success = await service.editUser(updatedUser)
      if (success) {
        closeFormDialog()
      }
    }
  }

  /**
   * 工具栏按钮点击处理
   * （基于原 useUserService.handleToolbarClick 的逻辑，但直接调用对话框等页面行为）
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
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
          message.warning('请先点击选择要编辑的用户')
          return
        }
        openEditDialog(currentRow as User)
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
          message.warning('请先点击选择要删除的用户')
          return
        }
        await service.deleteUser(currentRow as User)
        break
      }

      case 'resetPassword': {
        // 重置当前高亮行的密码
        if (!gridRef?.value) {
          message.warning('Grid 引用未设置')
          return
        }
        const currentRow = gridRef.value.getCurrentRecord()
        if (!currentRow) {
          message.warning('请先点击选择要重置密码的用户')
          return
        }
        await service.resetPassword(currentRow as User)
        break
      }

      case 'search': {
        // 如果传递了表单数据，直接使用它进行查询
        // formData 参数在 SearchForm 的 handleToolbarClick 中传递
        await service.handleSearch(formData)
        break
      }
    }
  }

  // 用户角色授权对话框状态
  const roleAuthDialogVisible = ref(false)
  const currentAuthUser = ref<User | null>(null)

  /** 打开用户角色授权对话框 */
  const openRoleAuthDialog = (user: User) => {
    currentAuthUser.value = user
    roleAuthDialogVisible.value = true
  }

  /** 关闭用户角色授权对话框 */
  const closeRoleAuthDialog = () => {
    roleAuthDialogVisible.value = false
    currentAuthUser.value = null
  }

  /**
   * 右键菜单点击处理
   * （基于原 useUserService.handleMenuClick 的逻辑）
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: User }) => {
    if (!row) return

    switch (code) {
      case 'view':
        openViewDialog(row)
        break

      case 'edit':
        openEditDialog(row)
        break

      case 'delete':
        await service.deleteUser(row)
        break

      case 'resetPassword':
        await service.resetPassword(row)
        break

      case 'roleAuth':
        openRoleAuthDialog(row)
        break
    }
  }

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditUser,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    // 用户角色授权对话框
    roleAuthDialogVisible,
    currentAuthUser,
    openRoleAuthDialog,
    closeRoleAuthDialog,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

export type UserPage = ReturnType<typeof useUserPage>


