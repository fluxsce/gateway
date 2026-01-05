import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import type { Role } from '../types'
import { useRoleService } from './useRoleService'

/**
 * 角色管理页面级 Hook
 * - 组合 useRoleService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */
export function useRolePage(gridRef?: Ref<any> | any, searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useRoleService(searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditRole = ref<Role | null>(null)

  // 角色授权抽屉状态
  const roleAuthDrawerVisible = ref(false)
  const roleAuthRoleId = ref<string>('')
  const roleAuthRoleName = ref<string>('')

  /** 打开新增角色对话框 */
  const openAddDialog = () => {
    formDialogMode.value = 'create'
    currentEditRole.value = null
    formDialogVisible.value = true
  }

  /** 打开编辑角色对话框 */
  const openEditDialog = (role: Role) => {
    formDialogMode.value = 'edit'
    currentEditRole.value = role
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditRole.value = null
  }
  
  /** 打开查看详情对话框 */
  const openViewDialog = (role: Role) => {
    formDialogMode.value = 'view'
    currentEditRole.value = role
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
      const success = await service.addRole(formData as Role)
      if (success) {
        closeFormDialog()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditRole.value) return
      // 合并当前角色ID和租户ID，确保更新的是正确的记录
      const updatedRole = {
        ...currentEditRole.value,
        ...formData
      } as Role
      const success = await service.editRole(updatedRole)
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
          message.warning('请先点击选择要编辑的角色')
          return
        }
        openEditDialog(currentRow as Role)
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
          message.warning('请先点击选择要删除的角色')
          return
        }
        await service.deleteRole(currentRow as Role)
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

  /**
   * 打开角色授权抽屉
   */
  const openRoleAuthDrawer = (role: Role) => {
    roleAuthRoleId.value = role.roleId
    roleAuthRoleName.value = role.roleName
    roleAuthDrawerVisible.value = true
  }

  /**
   * 关闭角色授权抽屉
   */
  const closeRoleAuthDrawer = () => {
    roleAuthDrawerVisible.value = false
    roleAuthRoleId.value = ''
    roleAuthRoleName.value = ''
  }

  /**
   * 右键菜单点击处理
   */
  const handleMenuClick = async ({ code, row }: { code: string; row?: Role }) => {
    if (!row) return

    switch (code) {
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
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditRole,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    handleFormSubmit,

    // 角色授权抽屉
    roleAuthDrawerVisible,
    roleAuthRoleId,
    roleAuthRoleName,
    openRoleAuthDrawer,
    closeRoleAuthDrawer,

    // 事件处理器
    handleToolbarClick,
    handleMenuClick,
    handleSearch
  }
}

export type RolePage = ReturnType<typeof useRolePage>

