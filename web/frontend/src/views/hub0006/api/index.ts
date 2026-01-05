import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { Resource } from '../types'

const resourceApi = createApi('/gateway/hub0006')

/**
 * 分页查询资源列表
 * @param params 查询参数
 * @returns 资源列表和分页信息
 */
export async function queryResources(params: any): Promise<JsonDataObj> {
  return resourceApi.post('/queryResources', params)
}

/**
 * 添加资源
 * @param data 资源创建数据
 * @returns 操作结果
 */
export async function addResource(data: Resource): Promise<JsonDataObj> {
  return resourceApi.post('/addResource', data)
}

/**
 * 编辑资源信息
 * @param data 资源更新数据
 * @returns 操作结果
 */
export async function editResource(data: Resource): Promise<JsonDataObj> {
  return resourceApi.post('/editResource', data)
}

/**
 * 删除资源
 * @param resourceId 资源ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteResource(resourceId: string, tenantId: string): Promise<JsonDataObj> {
  return resourceApi.post('/deleteResource', { resourceId, tenantId })
}

/**
 * 查询单个资源信息
 * @param resourceId 资源ID
 * @param tenantId 租户ID
 * @returns 资源信息
 */
export async function getResource(resourceId: string, tenantId: string): Promise<JsonDataObj> {
  return resourceApi.post('/getResource', { resourceId, tenantId })
}

/**
 * 更新资源状态
 * @param resourceId 资源ID
 * @param status 状态(Y/N)
 * @returns 操作结果
 */
export async function updateResourceStatus(resourceId: string, status: string): Promise<JsonDataObj> {
  return resourceApi.post('/updateResourceStatus', { resourceId, status })
}

