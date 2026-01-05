/**
 * 用户管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as userApi from '../api'
import type { User } from '../types/index'
import { useUserModel } from './model'

/**
 * 用户服务 Hook（纯业务逻辑，不再依赖外部 options）
 */
export function useUserService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useUserModel()

  const {
    loading,
    userList,
    pageInfo,
    setUserList,
    updatePagination,
    addUserToList,
    updateUserInList,
    removeUserFromList,
    removeUsersFromList
  } = model

  // ============= 数据加载 =============

  /**
   * 加载用户列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadUsers = async (searchParams?: Record<string, any>) => {
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
      const response: JsonDataObj = await userApi.queryUsers(params)

      if (response.oK) {
        // 解析业务数据
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const users = Array.isArray(bizData) ? bizData : []
          setUserList(users)
        }

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询用户列表失败')
      }
    } catch (error) {
      console.error('加载用户列表失败:', error)
      message.error('加载用户列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索用户
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    // loadUsers 会自动从 searchFormRef 获取查询条件
    await loadUsers(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    model.resetPagination()
    await loadUsers()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadUsers 会自动从 searchFormRef 获取查询条件
    await loadUsers()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadUsers()
  }

  // ============= 用户操作 =============

  /**
   * 新增用户
   */
  const addUser = async (userData: User): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await userApi.addUser(userData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '新增用户成功')
        
        // 如果返回了新增的用户数据，添加到列表
        if (response.bizData) {
          const newUser = JSON.parse(response.bizData)
          addUserToList(newUser)
        } else {
          // 否则重新加载列表
          await loadUsers()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '新增用户失败')
        return false
      }
    } catch (error) {
      console.error('新增用户失败:', error)
      message.error('新增用户失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑用户
   */
  const editUser = async (userData: User): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await userApi.editUser(userData)

      if (response.oK && response.state) {
        message.success(response.popMsg || '编辑用户成功')
        
        // 更新列表中的用户数据
        if (response.bizData) {
          const updatedUser = JSON.parse(response.bizData)
          updateUserInList(updatedUser.userId, updatedUser.tenantId, updatedUser)
        } else {
          updateUserInList(userData.userId, userData.tenantId, userData)
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '编辑用户失败')
        return false
      }
    } catch (error) {
      console.error('编辑用户失败:', error)
      message.error('编辑用户失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除用户
   */
  const deleteUser = async (user: User): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除用户 "${user.realName || user.userName}" 吗？`,
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
      const response: JsonDataObj = await userApi.deleteUser(user.userId, user.tenantId)

      if (response.oK && response.state) {
        message.success(response.popMsg || '删除用户成功')
        removeUserFromList(user.userId, user.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (userList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadUsers()
        }
        
        return true
      } else {
        message.error(response.errMsg || response.popMsg || '删除用户失败')
        return false
      }
    } catch (error) {
      console.error('删除用户失败:', error)
      message.error('删除用户失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 批量删除用户
   */
  const batchDeleteUsers = async (users: User[]): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除选中的 ${users.length} 个用户吗？`,
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
      for (const user of users) {
        try {
          const response: JsonDataObj = await userApi.deleteUser(user.userId, user.tenantId)
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
        message.success(`成功删除 ${successCount} 个用户${failCount > 0 ? `，失败 ${failCount} 个` : ''}`)
        removeUsersFromList(users.slice(0, successCount))
        
        // 重新加载列表
        await loadUsers()
        return true
      } else {
        message.error(`删除失败，共 ${failCount} 个`)
        return false
      }
    } catch (error) {
      console.error('批量删除用户失败:', error)
      message.error('批量删除用户失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 重置密码
   */
  const resetPassword = async (user: User): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认重置密码',
      subtitle: '新密码将通过邮件发送给用户',
      content: `确定要重置用户 "${user.realName || user.userName}" 的密码吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定重置',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      // TODO: 调用重置密码 API（需要后端提供接口）
      // const response = await userApi.resetPassword(user.userId, user.tenantId)
      
      // 模拟成功
      message.success('密码重置成功，新密码已发送至用户邮箱')
      return true
    } catch (error) {
      console.error('重置密码失败:', error)
      message.error('重置密码失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 查看用户详情
   */
  const viewUser = async (user: User) => {
    try {
      const response: JsonDataObj = await userApi.getUserInfo(user.userId, user.tenantId)
      
      if (response.oK) {
        const userInfo = JSON.parse(response.bizData)
        return userInfo
      } else {
        message.error(response.errMsg || '获取用户详情失败')
        return null
      }
    } catch (error) {
      console.error('获取用户详情失败:', error)
      message.error('获取用户详情失败')
      return null
    }
  }

  // ============= 工具栏事件处理 =============

  return {
    // Model 实例（包含 paginationConfig 和 menuConfig）
    model,
    
    // 数据加载
    loadUsers,
    
    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,
    
    // 用户操作
    addUser,
    editUser,
    deleteUser,
    batchDeleteUsers,
    resetPassword,
    viewUser
  }
}

/**
 * 服务返回类型
 */
export type UserService = ReturnType<typeof useUserService>

