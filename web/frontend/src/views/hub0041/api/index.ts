import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { Namespace } from '../types'

const namespaceApi = createApi('/serviceCenter/hub0041')

/**
 * 分页查询命名空间列表
 * @param params 查询参数
 * @returns 命名空间列表和分页信息
 */
export async function queryNamespaces(params: any): Promise<JsonDataObj> {
  return namespaceApi.post('/queryNamespaces', params)
}

/**
 * 查询命名空间详情
 * @param namespaceId 命名空间ID
 * @returns 命名空间详情
 */
export async function getNamespace(namespaceId: string): Promise<JsonDataObj> {
  return namespaceApi.post('/getNamespace', { namespaceId })
}

/**
 * 添加命名空间
 * @param data 命名空间创建数据
 * @returns 操作结果
 */
export async function addNamespace(data: Namespace): Promise<JsonDataObj> {
  return namespaceApi.post('/addNamespace', data)
}

/**
 * 编辑命名空间
 * @param data 命名空间更新数据
 * @returns 操作结果
 */
export async function editNamespace(
  data: Partial<Namespace> & {
    namespaceId: string
  },
): Promise<JsonDataObj> {
  return namespaceApi.post('/editNamespace', data)
}

/**
 * 删除命名空间
 * @param namespaceId 命名空间ID
 * @returns 操作结果
 */
export async function deleteNamespace(namespaceId: string): Promise<JsonDataObj> {
  return namespaceApi.post('/deleteNamespace', { namespaceId })
}

