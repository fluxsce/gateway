import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'

const clusterEventApi = createApi('/gateway/hub0008')

/**
 * 分页查询集群事件列表
 * @param params 查询参数
 * @returns 集群事件列表和分页信息
 */
export async function queryClusterEvents(params: any): Promise<JsonDataObj> {
  return clusterEventApi.post('/queryClusterEvents', params)
}

/**
 * 获取集群事件详情
 * @param eventId 事件ID
 * @returns 集群事件详情
 */
export async function getClusterEventDetail(eventId: string): Promise<JsonDataObj> {
  return clusterEventApi.post('/getClusterEventDetail', { eventId })
}

/**
 * 分页查询集群事件处理节点列表
 * @param params 查询参数
 * @returns 集群事件处理节点列表和分页信息
 */
export async function queryClusterEventAcks(params: any): Promise<JsonDataObj> {
  return clusterEventApi.post('/queryClusterEventAcks', params)
}

/**
 * 获取集群事件确认详情
 * @param ackId 确认ID
 * @returns 集群事件确认详情
 */
export async function getClusterEventAckDetail(ackId: string): Promise<JsonDataObj> {
  return clusterEventApi.post('/getClusterEventAckDetail', { ackId })
}

