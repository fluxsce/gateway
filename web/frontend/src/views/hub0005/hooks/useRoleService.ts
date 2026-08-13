/**
 * 角色管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { rsConfirm } from '@/ui'
import { useAppMessage } from '@/composables/useAppMessage'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { JsonDataObj } from '@/types/api'
import { createBackendPaginationParams } from '@/utils/pagination'
import { WarningOutline } from '@vicons/ionicons5'
import type { Ref } from 'vue'
import * as roleApi from '../api'
import type { Role } from '../types/index'
import { useRoleModel } from './model'

/**
 * 角色服务 Hook（纯业务逻辑，不再依赖外部 options）
 */
export function useRoleService(searchFormRef?: Ref<any> | any) {
  const message = useAppMessage()
  const { t } = useModuleI18n('hub0005')

  const model = useRoleModel()

  const {
    loading,
    roleList,
    pageInfo,
    setRoleList,
    updatePagination,
    addRoleToList,
    updateRoleInList,
    removeRoleFromList,
    removeRolesFromList,
  } = model

  /**
   * 加载角色列表
   */
  const loadRoles = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined,
            ),
          )
        : {}

      const params = {
        ...effectiveSearchParams,
        ...createBackendPaginationParams(pageInfo.value?.pageIndex, pageInfo.value?.pageSize),
      }

      const response: JsonDataObj = await roleApi.queryRoles(params)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const roles = Array.isArray(bizData) ? bizData : []
          setRoleList(roles)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || t('role.message.queryFailed'))
      }
    } catch (error) {
      console.error('加载角色列表失败:', error)
      message.error(t('role.message.loadFailed'))
    } finally {
      loading.value = false
    }
  }

  const handleSearch = async (searchParams?: Record<string, any>) => {
    await loadRoles(searchParams)
  }

  const handleReset = async () => {
    model.resetPagination()
    await loadRoles()
  }

  const handlePageChange = async ({
    currentPage,
    pageSize,
  }: {
    currentPage: number
    pageSize: number
  }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadRoles()
  }

  const handleRefresh = async () => {
    await loadRoles()
  }

  const addRole = async (roleData: Role): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.addRole(roleData)

      if (response.oK && response.state) {
        message.success(response.popMsg || t('role.message.createSuccess'))

        if (response.bizData) {
          const newRole = JSON.parse(response.bizData)
          addRoleToList(newRole)
        } else {
          await loadRoles()
        }

        return true
      }

      message.error(response.errMsg || response.popMsg || t('role.message.createFailed'))
      return false
    } catch (error) {
      console.error('新增角色失败:', error)
      message.error(t('role.message.createFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const editRole = async (roleData: Role): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.editRole(roleData)

      if (response.oK && response.state) {
        message.success(response.popMsg || t('role.message.updateSuccess'))

        if (response.bizData) {
          const updatedRole = JSON.parse(response.bizData)
          updateRoleInList(updatedRole.roleId, updatedRole.tenantId, updatedRole)
        } else {
          updateRoleInList(roleData.roleId, roleData.tenantId, roleData)
        }

        return true
      }

      message.error(response.errMsg || response.popMsg || t('role.message.updateFailed'))
      return false
    } catch (error) {
      console.error('编辑角色失败:', error)
      message.error(t('role.message.updateFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteRole = async (role: Role): Promise<boolean> => {
    const confirmed = await rsConfirm.warning({
      title: t('role.message.deleteConfirmTitle'),
      subtitle: t('role.message.deleteConfirmSubtitle'),
      description: t('role.message.deleteConfirm', { name: role.roleName }),
      icon: WarningOutline,
      confirmText: t('role.message.deleteConfirmOk'),
      cancelText: t('role.message.deleteConfirmCancel'),
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.deleteRole(role.roleId, role.tenantId)

      if (response.oK && response.state) {
        message.success(response.popMsg || t('role.message.deleteSuccess'))
        removeRoleFromList(role.roleId, role.tenantId)

        if (roleList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadRoles()
        }

        return true
      }

      message.error(response.errMsg || response.popMsg || t('role.message.deleteFailed'))
      return false
    } catch (error) {
      console.error('删除角色失败:', error)
      message.error(t('role.message.deleteFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  const batchDeleteRoles = async (roles: Role[]): Promise<boolean> => {
    const confirmed = await rsConfirm.warning({
      title: t('role.message.batchDeleteConfirmTitle'),
      subtitle: t('role.message.deleteConfirmSubtitle'),
      description: t('role.message.batchDeleteConfirm', { count: roles.length }),
      icon: WarningOutline,
      confirmText: t('role.message.deleteConfirmOk'),
      cancelText: t('role.message.deleteConfirmCancel'),
      width: 500,
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      let successCount = 0
      let failCount = 0

      for (const role of roles) {
        try {
          const response: JsonDataObj = await roleApi.deleteRole(role.roleId, role.tenantId)
          if (response.oK && response.state) {
            successCount++
          } else {
            failCount++
          }
        } catch {
          failCount++
        }
      }

      if (successCount > 0) {
        message.success(
          failCount > 0
            ? t('role.message.batchDeleteResultWithFail', {
                success: successCount,
                fail: failCount,
              })
            : t('role.message.batchDeleteResult', { success: successCount }),
        )
        removeRolesFromList(roles.slice(0, successCount))
        await loadRoles()
        return true
      }

      message.error(t('role.message.batchDeleteAllFailed', { fail: failCount }))
      return false
    } catch (error) {
      console.error('批量删除角色失败:', error)
      message.error(t('role.message.batchDeleteFailed'))
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    model,
    loadRoles,
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    addRole,
    editRole,
    deleteRole,
    batchDeleteRoles,
  }
}

/**
 * 角色服务类型
 */
export type RoleService = ReturnType<typeof useRoleService>
