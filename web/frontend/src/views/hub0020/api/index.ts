import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { GatewayInstance } from '../types'

const gatewayApi = createApi('/gateway/hub0020')

/**
 * 分页查询网关实例列表
 * @param params 查询参数
 * @returns 网关实例列表和分页信息
 */
export async function queryGatewayInstances(params: any): Promise<JsonDataObj> {
  return gatewayApi.post('/queryGatewayInstances', params)
}

/**
 * 查询网关实例详情
 * @param gatewayInstanceId 网关实例ID
 * @param tenantId 租户ID
 * @returns 网关实例详情
 */
export async function getGatewayInstance(
  gatewayInstanceId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/getGatewayInstance', { gatewayInstanceId, tenantId })
}

/**
 * 添加网关实例
 * @param data 网关实例创建数据
 * @returns 操作结果
 */
export async function addGatewayInstance(
  data: GatewayInstance,
): Promise<JsonDataObj> {
  return gatewayApi.post('/addGatewayInstance', data)
}

/**
 * 编辑网关实例
 * @param data 网关实例更新数据
 * @returns 操作结果
 */
export async function editGatewayInstance(
  data: Partial<GatewayInstance> & {
    gatewayInstanceId: string
  },
): Promise<JsonDataObj> {
  return gatewayApi.post('/editGatewayInstance', data)
}

/**
 * 删除网关实例
 * @param gatewayInstanceId 网关实例ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteGatewayInstance(
  gatewayInstanceId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/deleteGatewayInstance', { gatewayInstanceId, tenantId })
}

/**
 * 启动网关实例
 * @param gatewayInstanceId 网关实例ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function startGatewayInstance(
  gatewayInstanceId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/startGatewayInstance', { gatewayInstanceId, tenantId })
}

/**
 * 停止网关实例
 * @param gatewayInstanceId 网关实例ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function stopGatewayInstance(
  gatewayInstanceId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/stopGatewayInstance', { gatewayInstanceId, tenantId })
}

/**
 * 重新加载网关实例配置
 * @param gatewayInstanceId 网关实例ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function reloadGatewayInstance(
  gatewayInstanceId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/reloadGatewayInstance', { gatewayInstanceId, tenantId })
}

/**
 * 获取日志配置详情
 * @param logConfigId 日志配置ID
 * @returns 日志配置详情
 */
export async function getLogConfig(
  logConfigId: string,
): Promise<JsonDataObj> {
  return gatewayApi.post('/getLogConfig', { logConfigId })
}

/**
 * 编辑日志配置
 * @param data 日志配置数据
 * @returns 操作结果
 */
export async function editLogConfig(
  data: Record<string, any>,
): Promise<JsonDataObj> {
  return gatewayApi.post('/editLogConfig', data)
}
