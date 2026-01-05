import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { User } from '../types'

const userApi = createApi('/gateway/hub0002')

/**
 * 分页查询用户列表
 * @param params 查询参数
 * @returns 用户列表和分页信息
 */
export async function queryUsers(params: any): Promise<JsonDataObj> {
  return userApi.post('/queryUsers', params)
}
/**
 * 查询部门树
 * @returns 部门树
 */
export async function queryDeptsTree(): Promise<JsonDataObj> {
  return userApi.post('/queryDeptsTree')
}

/**
 * 删除用户
 * @param userId 用户ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteUser(userId: string, tenantId: string): Promise<JsonDataObj> {
  return userApi.post('/deleteUser', { userId, tenantId })
}

/**
 * 编辑用户信息
 * @param data 用户更新数据
 * @returns 操作结果
 */
export async function editUser(data: User): Promise<JsonDataObj> {
  return userApi.post('/editUser', data)
}

/**
 * 添加用户
 * @param data 用户创建数据
 * @returns 操作结果
 */
export async function addUser(data: User): Promise<JsonDataObj> {
  return userApi.post('/addUser', data)
}

/**
 * 查询单个用户信息
 * @param userId 用户ID
 * @param tenantId 租户ID
 * @returns 用户信息
 */
export async function getUserInfo(userId: string, tenantId: string): Promise<JsonDataObj> {
  return userApi.post('/getUser', { userId, tenantId })
}

/**
 * 修改密码
 * @param data 密码修改数据
 * @returns 操作结果
 */
export async function changePassword(data: {
  userId: string
  tenantId: string
  oldPassword: string
  newPassword: string
}): Promise<JsonDataObj> {
  return userApi.post('/changePassword', data)
}

/**
 * 获取用户角色列表
 * @param userId 用户ID
 * @returns 用户角色列表
 */
export async function getUserRoles(userId: string): Promise<JsonDataObj> {
  return userApi.post('/getUserRoles', { userId })
}

/**
 * 为用户分配角色
 * @param data 用户角色分配数据
 * @returns 操作结果
 */
export async function assignUserRoles(data: {
  userId: string
  roleIds: string[] | string
  expireTime?: string
}): Promise<JsonDataObj> {
  // 将 roleIds 数组转换为逗号分割的字符串
  const requestData = {
    userId: data.userId,
    roleIds: Array.isArray(data.roleIds) ? data.roleIds.join(',') : data.roleIds,
    expireTime: data.expireTime
  }
  return userApi.post('/assignUserRoles', requestData)
}
