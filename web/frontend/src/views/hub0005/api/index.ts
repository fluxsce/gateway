import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { Role } from '../types'

const roleApi = createApi('/gateway/hub0005')

/**
 * 分页查询角色列表
 * @param params 查询参数
 * @returns 角色列表和分页信息
 */
export async function queryRoles(params: any): Promise<JsonDataObj> {
  return roleApi.post('/queryRoles', params)
}

/**
 * 添加角色
 * @param data 角色创建数据
 * @returns 操作结果
 */
export async function addRole(data: Role): Promise<JsonDataObj> {
  return roleApi.post('/addRole', data)
}

/**
 * 编辑角色信息
 * @param data 角色更新数据
 * @returns 操作结果
 */
export async function editRole(data: Role): Promise<JsonDataObj> {
  return roleApi.post('/editRole', data)
}

/**
 * 删除角色
 * @param roleId 角色ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteRole(roleId: string, tenantId: string): Promise<JsonDataObj> {
  return roleApi.post('/deleteRole', { roleId, tenantId })
}

/**
 * 查询单个角色信息
 * @param roleId 角色ID
 * @param tenantId 租户ID
 * @returns 角色信息
 */
export async function getRole(roleId: string, tenantId: string): Promise<JsonDataObj> {
  return roleApi.post('/getRole', { roleId, tenantId })
}

/**
 * 获取角色授权的资源列表（树形结构）
 * @param roleId 角色ID
 * @returns 资源树形列表，每个资源包含 checked 字段表示是否已授权
 */
export async function getRoleResources(roleId: string): Promise<JsonDataObj> {
  return roleApi.post('/getRoleResources', { roleId })
}

/**
 * 保存角色授权
 * @param data 授权数据
 * @param data.roleId 角色ID
 * @param data.resourceIds 资源ID列表（逗号分割的字符串，如 "id1,id2,id3"）
 * @param data.permissionType 权限类型（ALLOW/DENY），默认为 ALLOW
 * @param data.expireTime 过期时间（可选）
 * @returns 操作结果
 */
export async function saveRoleResources(data: {
  roleId: string
  resourceIds: string
  permissionType?: string
  expireTime?: string | null
}): Promise<JsonDataObj> {
  return roleApi.post('/saveRoleResources', data)
}

