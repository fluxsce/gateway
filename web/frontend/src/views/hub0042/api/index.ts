import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { Service, ServiceNode } from '../types'

const serviceApi = createApi('/gateway/hub0042')

/**
 * 分页查询服务列表
 * @param params 查询参数
 * @returns 服务列表和分页信息
 */
export async function queryServices(params: any): Promise<JsonDataObj> {
  return serviceApi.post('/queryServices', params)
}

/**
 * 查询服务详情
 * @param namespaceId 命名空间ID
 * @param groupName 分组名称
 * @param serviceName 服务名称
 * @returns 服务详情
 */
export async function getService(namespaceId: string, groupName: string, serviceName: string): Promise<JsonDataObj> {
  return serviceApi.post('/getService', { namespaceId, groupName, serviceName })
}

/**
 * 添加服务
 * @param data 服务创建数据
 * @returns 操作结果
 */
export async function addService(data: Service): Promise<JsonDataObj> {
  return serviceApi.post('/addService', data)
}

/**
 * 编辑服务
 * @param data 服务更新数据
 * @returns 操作结果
 */
export async function editService(
  data: Partial<Service> & {
    namespaceId: string
    groupName: string
    serviceName: string
  },
): Promise<JsonDataObj> {
  return serviceApi.post('/editService', data)
}

/**
 * 删除服务
 * @param namespaceId 命名空间ID
 * @param groupName 分组名称
 * @param serviceName 服务名称
 * @returns 操作结果
 */
export async function deleteService(namespaceId: string, groupName: string, serviceName: string): Promise<JsonDataObj> {
  return serviceApi.post('/deleteService', { namespaceId, groupName, serviceName })
}

/**
 * 编辑节点
 * @param data 节点更新数据
 * @returns 操作结果
 */
export async function editNode(
  data: Partial<ServiceNode> & {
    nodeId: string
  },
): Promise<JsonDataObj> {
  return serviceApi.post('/editNode', data)
}

/**
 * 下线节点
 * @param nodeId 节点ID
 * @returns 操作结果
 */
export async function offlineNode(nodeId: string): Promise<JsonDataObj> {
  return serviceApi.post('/offlineNode', { nodeId })
}

/**
 * 上线节点（通过 editNode 实现）
 * @param nodeId 节点ID
 * @returns 操作结果
 */
export async function onlineNode(nodeId: string): Promise<JsonDataObj> {
  // 上线节点通过 editNode 实现，将 instanceStatus 设置为 UP
  return editNode({
    nodeId,
    instanceStatus: 'UP',
    healthyStatus: 'HEALTHY', // 上线时同时设置健康状态
  })
}

