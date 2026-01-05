/**
 * Hub0040 服务治理模块 - API接口层
 * 
 * 所有API请求都统一在 /gateway/hub0040 路径下，不进行跨模块处理
 * 遵循RESTful API设计规范，使用标准HTTP方法：
 * 
 * 服务分组管理API：
 * - POST /queryServiceGroups - 查询服务分组列表（支持分页、搜索和过滤）
 * - POST /getServiceGroup - 获取服务分组详情
 * - POST /createServiceGroup - 创建服务分组
 * - POST /updateServiceGroup - 更新服务分组
 * - POST /deleteServiceGroup - 删除服务分组
 * 
 * 注意：租户ID由后端自动从session中获取，前端无需传入
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
  ServiceGroupCreateRequest,
  ServiceGroupQueryRequest,
  ServiceGroupUpdateRequest
} from '../types'

// 创建API实例 - 所有请求都在hub0040路径下，不跨模块处理
const namespaceManagementApi = createApi('/gateway/hub0040')

/**
 * 查询服务分组列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 服务分组列表
 */
export const queryServiceGroups = async (params: ServiceGroupQueryRequest): Promise<JsonDataObj> => {
  return namespaceManagementApi.post('/queryServiceGroups', params)
}

/**
 * 获取服务分组详情
 * @param serviceGroupId 服务分组ID
 * @returns 服务分组详情
 */
export const getServiceGroupDetail = async (serviceGroupId: string): Promise<JsonDataObj> => {
  return namespaceManagementApi.post('/getServiceGroup', { serviceGroupId })
}

/**
 * 创建服务分组
 * @param data 创建数据
 * @returns 创建结果
 */
export const createServiceGroup = async (data: ServiceGroupCreateRequest): Promise<JsonDataObj> => {
  return namespaceManagementApi.post('/createServiceGroup', data)
}

/**
 * 更新服务分组
 * @param data 更新数据
 * @returns 更新结果
 */
export const updateServiceGroup = async (data: ServiceGroupUpdateRequest): Promise<JsonDataObj> => {
  return namespaceManagementApi.post('/updateServiceGroup', data)
}

/**
 * 删除服务分组
 * @param serviceGroupId 服务分组ID
 * @returns 删除结果
 */
export const deleteServiceGroup = async (serviceGroupId: string): Promise<JsonDataObj> => {
  return namespaceManagementApi.post('/deleteServiceGroup', { serviceGroupId })
}
