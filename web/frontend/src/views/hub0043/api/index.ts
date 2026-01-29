import { createApi } from '@/api/request';
import type { JsonDataObj } from '@/types/api';
import type { Config, ConfigQuery, RollbackRequest } from '../types';

const configApi = createApi('/gateway/hub0043')

/**
 * 分页查询配置列表
 * @param params 查询参数
 * @returns 配置列表和分页信息
 */
export async function queryConfigs(params: ConfigQuery & { page?: number; pageSize?: number }): Promise<JsonDataObj> {
  return configApi.post('/queryConfigs', params)
}

/**
 * 查询配置详情
 * @param namespaceId 命名空间ID
 * @param groupName 分组名称
 * @param configDataId 配置数据ID
 * @returns 配置详情
 */
export async function getConfig(namespaceId: string, groupName: string, configDataId: string): Promise<JsonDataObj> {
  return configApi.post('/getConfig', { namespaceId, groupName, configDataId })
}

/**
 * 添加配置
 * @param data 配置创建数据
 * @returns 操作结果
 */
export async function addConfig(data: Partial<Config> & { namespaceId: string; configDataId: string; configContent: string }): Promise<JsonDataObj> {
  return configApi.post('/addConfig', data)
}

/**
 * 编辑配置
 * @param data 配置更新数据
 * @returns 操作结果
 */
export async function editConfig(
  data: Partial<Config> & {
    namespaceId: string
    groupName: string
    configDataId: string
    configContent: string
  },
): Promise<JsonDataObj> {
  return configApi.post('/editConfig', data)
}

/**
 * 删除配置
 * @param namespaceId 命名空间ID
 * @param groupName 分组名称
 * @param configDataId 配置数据ID
 * @returns 操作结果
 */
export async function deleteConfig(namespaceId: string, groupName: string, configDataId: string): Promise<JsonDataObj> {
  return configApi.post('/deleteConfig', { namespaceId, groupName, configDataId })
}

/**
 * 查询配置历史
 * @param params 查询参数
 * @returns 配置历史列表
 */
export async function queryConfigHistory(params: {
  namespaceId: string
  groupName: string
  configDataId: string
  limit?: number
}): Promise<JsonDataObj> {
  return configApi.post('/queryConfigHistory', { ...params, limit: params.limit || 50 })
}

/**
 * 根据历史配置ID获取详情
 * @param configHistoryId 配置历史ID
 * @returns 配置历史详情
 */
export async function getHistoryById(configHistoryId: string): Promise<JsonDataObj> {
  return configApi.post('/getHistoryById', { configHistoryId })
}

/**
 * 回滚配置
 * @param data 回滚请求数据
 * @returns 操作结果
 */
export async function rollbackConfig(data: RollbackRequest): Promise<JsonDataObj> {
  return configApi.post('/rollbackConfig', data)
}

