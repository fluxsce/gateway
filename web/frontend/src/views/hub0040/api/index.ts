import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { ServiceCenterInstance } from '../types'

const serviceCenterApi = createApi('/serviceCenter/hub0040')

/**
 * 分页查询服务中心实例列表
 * @param params 查询参数
 * @returns 服务中心实例列表和分页信息
 */
export async function queryServiceCenterInstances(params: any): Promise<JsonDataObj> {
  return serviceCenterApi.post('/queryServiceCenterInstances', params)
}

/**
 * 查询服务中心实例详情
 * @param instanceName 实例名称
 * @param environment 部署环境
 * @returns 服务中心实例详情
 */
export async function getServiceCenterInstance(
  instanceName: string,
  environment: string,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/getServiceCenterInstance', { instanceName, environment })
}

/**
 * 添加服务中心实例
 * @param data 服务中心实例创建数据
 * @returns 操作结果
 */
export async function addServiceCenterInstance(
  data: ServiceCenterInstance,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/addServiceCenterInstance', data)
}

/**
 * 编辑服务中心实例
 * @param data 服务中心实例更新数据
 * @returns 操作结果
 */
export async function editServiceCenterInstance(
  data: Partial<ServiceCenterInstance> & {
    instanceName: string
    environment: string
  },
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/editServiceCenterInstance', data)
}

/**
 * 删除服务中心实例
 * @param instanceName 实例名称
 * @param environment 部署环境
 * @returns 操作结果
 */
export async function deleteServiceCenterInstance(
  instanceName: string,
  environment: string,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/deleteServiceCenterInstance', { instanceName, environment })
}

/**
 * 启动服务中心实例
 * @param instanceName 实例名称
 * @param environment 部署环境
 * @returns 操作结果
 */
export async function startServiceCenterInstance(
  instanceName: string,
  environment: string,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/startServiceCenterInstance', { instanceName, environment })
}

/**
 * 停止服务中心实例
 * @param instanceName 实例名称
 * @param environment 部署环境
 * @returns 操作结果
 */
export async function stopServiceCenterInstance(
  instanceName: string,
  environment: string,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/stopServiceCenterInstance', { instanceName, environment })
}

/**
 * 重新加载服务中心实例配置
 * @param instanceName 实例名称
 * @param environment 部署环境
 * @returns 操作结果
 */
export async function reloadServiceCenterInstance(
  instanceName: string,
  environment: string,
): Promise<JsonDataObj> {
  return serviceCenterApi.post('/reloadServiceCenterInstance', { instanceName, environment })
}

