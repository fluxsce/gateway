/**
 * 角色管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as roleApi from '../api'
import type { Role } from '../types/index'
import { useRoleModel } from './model'

/**
 * 角色服务 Hook（纯业务逻辑，不再依赖外部 options）
 */
export function useRoleService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
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
    removeRolesFromList
  } = model

  // ============= 数据加载 =============

  /**
   * 加载角色列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadRoles = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const params = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        )
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await roleApi.queryRoles(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const roles = Array.isArray(bizData) ? bizData : []
          setRoleList(roles)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询角色列表失败')
      }
    } catch (error) {
      console.error('加载角色列表失败:', error)
      message.error('加载角色列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索角色
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadRoles 会自动从 searchFormRef 获取查询条件
    await loadRoles(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadRoles()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadRoles 会自动从 searchFormRef 获取查询条件
    await loadRoles()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadRoles()
  }

  // ============= 增删改 =============

  /**
   * 添加角色
   */
  const addRole = async (roleData: Role): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.addRole(roleData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '新增角色成功')
        
        // 如果返回了新增的角色数据，添加到列表
        if (response.bizData) {
          const newRole = JSON.parse(response.bizData)
          addRoleToList(newRole)
        } else {
          // 否则重新加载列表
          await loadRoles()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '新增角色失败')
        return false
      }
    } catch (error) {
      console.error('新增角色失败:', error)
      message.error('新增角色失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑角色
   */
  const editRole = async (roleData: Role): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.editRole(roleData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '编辑角色成功')
        
        // 更新列表中的角色数据
        if (response.bizData) {
          const updatedRole = JSON.parse(response.bizData)
          updateRoleInList(updatedRole.roleId, updatedRole.tenantId, updatedRole)
        } else {
          updateRoleInList(roleData.roleId, roleData.tenantId, roleData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '编辑角色失败')
        return false
      }
    } catch (error) {
      console.error('编辑角色失败:', error)
      message.error('编辑角色失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除角色
   */
  const deleteRole = async (role: Role): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除角色 "${role.roleName}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await roleApi.deleteRole(role.roleId, role.tenantId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除角色成功')
        removeRoleFromList(role.roleId, role.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (roleList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadRoles()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除角色失败')
        return false
      }
    } catch (error) {
      console.error('删除角色失败:', error)
      message.error('删除角色失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除角色
   */
  const batchDeleteRoles = async (roles: Role[]): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${roles.length} 个角色吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      let successCount = 0
      let failCount = 0

      // 逐个删除
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

      // 显示结果
      if (successCount > 0) {
        message.success(`成功删除 ${successCount} 个角色${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
        removeRolesFromList(roles.slice(0, successCount))
        
        // 重新加载列表
        await loadRoles()
        return true
      } else {
        message.error(`删除失败，共 ${failCount} 个`)
        return false
      }
    } catch (error) {
      console.error('批量删除角色失败:', error)
      message.error('批量删除角色失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadRoles,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 角色操作
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

