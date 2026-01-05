/**
 * Hub0041 服务注册中心 API 接口
 * 提供服务注册、服务发现、实例管理等功能的接口定义
 * 
 * 所有API请求都统一在 /gateway/hub0041 路径下，遵循RESTful API设计规范
 * 使用标准HTTP POST方法：
 * 
 * 服务管理API：
 * - POST /queryServices - 查询服务列表（支持分页、搜索和过滤）
 * - POST /getService - 获取服务详情
 * - POST /createService - 创建服务
 * - POST /updateService - 更新服务信息
 * - POST /deleteService - 删除服务
 * 
 * 服务实例管理API：
 * - POST /queryServiceInstances - 查询服务实例列表
 * - POST /getServiceInstance - 获取服务实例详情
 * - POST /createServiceInstance - 创建服务实例
 * - POST /updateServiceInstance - 更新服务实例信息
 * - POST /deleteServiceInstance - 删除服务实例
 * - POST /updateInstanceHeartbeat - 更新服务实例心跳
 * - POST /updateInstanceHealthStatus - 更新服务实例健康状态
 * 
 * 配置和元数据API：
 * - POST /getServiceGroups - 获取服务分组列表
 * - POST /getServiceProtocolTypes - 获取协议类型选项
 * - POST /getLoadBalanceStrategies - 获取负载均衡策略选项
 * - POST /getInstanceStatusOptions - 获取实例状态选项
 * - POST /getHealthStatusOptions - 获取健康状态选项
 * - POST /getClientTypeOptions - 获取客户端类型选项
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { 
  Service, 
  ServiceInstance, 
  ServiceDetail,
  ServiceEvent,
  ServiceQueryRequest,
  ServiceInstanceQueryRequest
} from '../types'

// 创建API实例 - 所有请求都在hub0041路径下
const serviceRegistryApi = createApi('/gateway/hub0041')

// ==================== 服务管理API ====================

/**
 * 查询服务列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 服务列表
 */
export const queryServices = async (params: ServiceQueryRequest): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/queryServices', params)
}

/**
 * 获取服务详情
 * @param serviceName 服务名称
 * @returns 服务详情
 */
export const getService = async (serviceName: string): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getService', { serviceName })
}

/**
 * 创建服务
 * @param data 服务数据
 * @returns 创建结果
 */
export const createService = async (data: Partial<Service>): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/createService', data)
}

/**
 * 更新服务信息
 * @param data 服务数据
 * @returns 更新结果
 */
export const updateService = async (data: Partial<Service>): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/updateService', data)
}

/**
 * 删除服务
 * @param serviceName 服务名称
 * @returns 删除结果
 */
export const deleteService = async (serviceName: string): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/deleteService', { serviceName })
}

// ==================== 服务实例管理API ====================

/**
 * 查询服务实例列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 服务实例列表
 */
export const queryServiceInstances = async (params: ServiceInstanceQueryRequest): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/queryServiceInstances', params)
}

/**
 * 获取服务实例详情
 * @param serviceInstanceId 服务实例ID
 * @returns 服务实例详情
 */
export const getServiceInstance = async (serviceInstanceId: string): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getServiceInstance', { serviceInstanceId })
}

/**
 * 创建服务实例
 * @param data 服务实例数据
 * @returns 创建结果
 */
export const createServiceInstance = async (data: Partial<ServiceInstance>): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/createServiceInstance', data)
}

/**
 * 更新服务实例信息
 * @param data 服务实例数据
 * @returns 更新结果
 */
export const updateServiceInstance = async (data: Partial<ServiceInstance>): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/updateServiceInstance', data)
}

/**
 * 删除服务实例
 * @param serviceInstanceId 服务实例ID
 * @returns 删除结果
 */
export const deleteServiceInstance = async (serviceInstanceId: string): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/deleteServiceInstance', { serviceInstanceId })
}

/**
 * 更新服务实例心跳
 * @param serviceInstanceId 服务实例ID
 * @returns 更新结果
 */
export const updateInstanceHeartbeat = async (serviceInstanceId: string): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/updateInstanceHeartbeat', { serviceInstanceId })
}

/**
 * 更新服务实例健康状态
 * @param serviceInstanceId 服务实例ID
 * @param healthStatus 健康状态
 * @returns 更新结果
 */
export const updateInstanceHealthStatus = async (
  serviceInstanceId: string, 
  healthStatus: 'HEALTHY' | 'UNHEALTHY' | 'UNKNOWN'
): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/updateInstanceHealthStatus', { 
    serviceInstanceId, 
    healthStatus 
  })
}

// ==================== 配置和元数据API ====================

/**
 * 获取服务分组列表
 * @returns 服务分组列表
 */
export const getServiceGroups = async (): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getServiceGroups', {})
}

/**
 * 获取协议类型选项
 * @returns 协议类型列表
 */
export const getServiceProtocolTypes = async (): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getServiceProtocolTypes', {})
}

/**
 * 获取负载均衡策略选项
 * @returns 负载均衡策略列表
 */
export const getLoadBalanceStrategies = async (): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getLoadBalanceStrategies', {})
}

/**
 * 获取实例状态选项
 * @returns 实例状态列表
 */
export const getInstanceStatusOptions = async (): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getInstanceStatusOptions', {})
}

/**
 * 获取健康状态选项
 * @returns 健康状态列表
 */
export const getHealthStatusOptions = async (): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/getHealthStatusOptions', {})
}

// ==================== 服务事件API ====================

/**
 * 查询服务事件列表
 * @param params 查询参数 
 * @returns 服务事件列表
 */
export const fetchServiceEventList = async (params: any): Promise<JsonDataObj> => {
  return serviceRegistryApi.post('/queryServiceEvents', params)
}

/**
 * 获取服务事件详情
 * @param serviceEventId 服务事件ID
 * @returns 服务事件详情
 */
export const fetchServiceEventById = async (serviceEventId: string): Promise<JsonDataObj> => {
  return await serviceRegistryApi.post('/getServiceEvent', { serviceEventId })
}

// ==================== 便捷封装方法 ====================

/**
 * 获取服务详情（包含实例列表）
 * @param serviceName 服务名称
 * @returns 服务详情
 */
export const getServiceDetail = async (serviceName: string): Promise<JsonDataObj> => {
  // 获取服务基本信息
  const serviceResponse = await getService(serviceName)
  if (!serviceResponse.oK) {
    return serviceResponse
  }

  // 获取服务实例列表
  const instancesResponse = await queryServiceInstances({ 
    serviceName,
    pageIndex: 1,
    pageSize: 1000 
  })

  let serviceData: any = {}
  let instancesData: any[] = []

  try {
    serviceData = typeof serviceResponse.bizData === 'string' 
      ? JSON.parse(serviceResponse.bizData) 
      : serviceResponse.bizData
  } catch (error) {
    console.error('Failed to parse service data:', error)
  }

  try {
    if (instancesResponse.oK && instancesResponse.bizData) {
      const parsedInstancesData = typeof instancesResponse.bizData === 'string' 
        ? JSON.parse(instancesResponse.bizData) 
        : instancesResponse.bizData
      instancesData = parsedInstancesData?.items || []
      console.log('Raw instances response:', instancesResponse)
      console.log('Parsed instances data:', parsedInstancesData)
      console.log('Final instances data:', instancesData)
      if (instancesData.length > 0) {
        console.log('First instance sample:', instancesData[0])
        console.log('Instance keys:', Object.keys(instancesData[0]))
      }
    }
  } catch (error) {
    console.error('Failed to parse instances data:', error)
  }

  // 组装完整的服务详情
  const serviceDetail = {
    ...serviceData,
    instances: instancesData,
    statistics: calculateServiceStatistics(instancesData)
  }

  return {
    oK: true,
    state: true,
    bizData: JSON.stringify(serviceDetail),
    extObj: null,
    pageQueryData: '',
    messageId: '',
    errMsg: '',
    popMsg: '',
    extMsg: '',
    pkey1: '',
    pkey2: '',
    pkey3: '',
    pkey4: '',
    pkey5: '',
    pkey6: ''
  }
}

/**
 * 计算服务统计信息
 * @param instances 实例列表
 * @returns 统计信息
 */
const calculateServiceStatistics = (instances: any[]) => {
  return {
    totalInstances: instances.length,
    healthyInstances: instances.filter(i => i.healthStatus === 'HEALTHY').length,
    unhealthyInstances: instances.filter(i => i.healthStatus === 'UNHEALTHY').length,
    upInstances: instances.filter(i => i.instanceStatus === 'UP').length,
    downInstances: instances.filter(i => i.instanceStatus === 'DOWN').length
  }
}
