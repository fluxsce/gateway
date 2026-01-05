import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { ProxyConfig } from '../components/instance-tree/types'
import type { ServiceDefinition } from '../components/service/types'
const proxyApi = createApi('/gateway/hub0022')

/**
 * 查询所有网关实例列表（hub0022模块）
 * @param params 查询参数
 * @returns 网关实例列表和分页信息
 */
export async function queryAllGatewayInstances(params: {
  instanceName?: string
  healthStatus?: 'Y' | 'N'
  activeFlag?: 'Y' | 'N'
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return proxyApi.post('/queryAllGatewayInstances', params)
}

/**
 * 根据网关实例ID获取代理配置（返回单条数据）
 * @param gatewayInstanceId 网关实例ID
 * @returns 代理配置对象或null
 */
export async function getProxyConfigByInstance(
  gatewayInstanceId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/getProxyConfigsByInstance', { gatewayInstanceId })
}

/**
 * 查询代理配置详情
 * @param proxyConfigId 代理配置ID
 * @param tenantId 租户ID
 * @returns 代理配置详情
 */
export async function getProxyConfig(
  proxyConfigId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/proxyConfig', { proxyConfigId, tenantId })
}

/**
 * 添加代理配置
 * @param data 代理配置创建数据
 * @returns 操作结果
 */
export async function addProxyConfig(
  data: ProxyConfig & { tenantId: string; gatewayInstanceId: string },
): Promise<JsonDataObj> {
  return proxyApi.post('/addProxyConfig', data)
}

/**
 * 编辑代理配置
 * @param data 代理配置更新数据
 * @returns 操作结果
 */
export async function editProxyConfig(
  data: Partial<ProxyConfig> & {
    tenantId: string
    proxyConfigId: string
    gatewayInstanceId: string
  },
): Promise<JsonDataObj> {
  return proxyApi.post('/editProxyConfig', data)
}

/**
 * 删除代理配置
 * @param proxyConfigId 代理配置ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteProxyConfig(
  proxyConfigId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/deleteProxyConfig', { proxyConfigId, tenantId })
}

/**
 * 代理连接测试
 * @param data 测试请求数据
 * @returns 测试结果
 */
export async function testProxyConnection(
  data: any & { tenantId: string },
): Promise<JsonDataObj> {
  return proxyApi.post('/testProxyConnection', data)
}

/**
 * 分页查询服务定义列表
 * @param params 查询参数
 * @returns 服务定义列表和分页信息
 */
export async function queryServiceDefinitions(params: any): Promise<JsonDataObj> {
  return proxyApi.post('/queryServiceDefinitions', params)
}

/**
 * 查询服务定义详情
 * @param serviceDefinitionId 服务定义ID
 * @param tenantId 租户ID
 * @returns 服务定义详情
 */
export async function getServiceDefinition(
  serviceDefinitionId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/getServiceDefinition', { serviceDefinitionId, tenantId })
}

/**
 * 添加服务定义
 * @param data 服务定义创建数据
 * @returns 操作结果
 */
export async function addServiceDefinition(
  data: Partial<ServiceDefinition> & { tenantId: string },
): Promise<JsonDataObj> {
  return proxyApi.post('/addServiceDefinition', data)
}

/**
 * 编辑服务定义
 * @param data 服务定义更新数据
 * @returns 操作结果
 */
export async function editServiceDefinition(
  data: Partial<ServiceDefinition> & {
    tenantId: string
    serviceDefinitionId: string
  },
): Promise<JsonDataObj> {
  return proxyApi.post('/editServiceDefinition', data)
}

/**
 * 删除服务定义
 * @param serviceDefinitionId 服务定义ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteServiceDefinition(
  serviceDefinitionId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/deleteServiceDefinition', { serviceDefinitionId, tenantId })
}

/**
 * 查询服务节点列表
 * @param params 查询参数
 * @returns 服务节点列表
 */
export async function queryServiceNodes(params: any): Promise<JsonDataObj> {
  return proxyApi.post('/queryServiceNodes', params)
}

/**
 * 获取服务节点详情
 * @param serviceNodeId 服务节点ID
 * @param tenantId 租户ID
 * @returns 服务节点详情
 */
export async function getServiceNode(
  serviceNodeId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/getServiceNode', { serviceNodeId, tenantId })
}

/**
 * 添加服务节点
 * @param data 服务节点数据（必须包含serviceDefinitionId字段，tenantId由后端从session获取）
 * @returns 操作结果
 */
export async function addServiceNode(data: {
  serviceDefinitionId: string // 关联的服务定义ID，必须提供
  nodeUrl: string
  nodeHost: string
  nodePort: number
  nodeProtocol: string
  nodeWeight: number
  healthStatus: string
  nodeEnabled: string
  nodeStatus: number
  activeFlag: string
  noteText?: string
  nodeMetadata?: any
  tenantId?: string // 可选，由后端从session获取
}): Promise<JsonDataObj> {
  return proxyApi.post('/addServiceNode', data)
}

/**
 * 编辑服务节点
 * @param data 服务节点数据（必须包含serviceDefinitionId和serviceNodeId字段，tenantId由后端从session获取）
 * @returns 操作结果
 */
export async function editServiceNode(data: {
  serviceNodeId: string // 服务节点ID，必须提供
  serviceDefinitionId: string // 关联的服务定义ID，必须提供
  nodeUrl?: string
  nodeHost?: string
  nodePort?: number
  nodeProtocol?: string
  nodeWeight?: number
  healthStatus?: string
  nodeEnabled?: string
  nodeStatus?: number
  activeFlag?: string
  noteText?: string
  nodeMetadata?: any
  tenantId?: string // 可选，由后端从session获取
}): Promise<JsonDataObj> {
  return proxyApi.post('/editServiceNode', data)
}

/**
 * 删除服务节点
 * @param serviceNodeId 服务节点ID
 * @param tenantId 租户ID
 * @returns 操作结果
 */
export async function deleteServiceNode(
  serviceNodeId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/deleteServiceNode', { serviceNodeId, tenantId })
}

/**
 * 检查服务健康状态
 * @param serviceDefinitionId 服务定义ID
 * @param tenantId 租户ID
 * @returns 健康检查结果
 */
export async function checkServiceHealth(
  serviceDefinitionId: string,
  tenantId: string,
): Promise<JsonDataObj> {
  return proxyApi.post('/checkServiceHealth', { serviceDefinitionId, tenantId })
}

/**
 * 更新节点健康状态
 * @param serviceNodeId 服务节点ID
 * @param tenantId 租户ID
 * @param healthStatus 健康状态
 * @returns 操作结果
 */
export async function updateNodeHealth(
  serviceNodeId: string,
  tenantId: string,
  healthStatus: 'Y' | 'N',
): Promise<JsonDataObj> {
  return proxyApi.post('/updateNodeHealth', { serviceNodeId, tenantId, healthStatus })
}

/**
 * 更新节点状态
 * @param data 包含节点ID和状态的对象
 * @returns 操作结果
 */
export async function updateServiceNodeStatus(data: {
  serviceNodeId: string
  activeFlag?: 'Y' | 'N'
  healthStatus?: 'Y' | 'N'
  nodeStatus?: number
  tenantId?: string
}): Promise<JsonDataObj> {
  const tenantId = data.tenantId || 'default'
  return proxyApi.post('/updateServiceNodeStatus', { ...data, tenantId })
}

/**
 * 批量删除服务节点
 * @param serviceNodeIds 服务节点ID数组
 * @returns 操作结果
 */
export async function batchDeleteServiceNodes(serviceNodeIds: string[]): Promise<JsonDataObj> {
  return proxyApi.post('/batchDeleteServiceNodes', { serviceNodeIds, tenantId: 'default' })
}

/**
 * 查询注册服务列表
 * @param params 查询参数
 * @returns 注册服务列表
 */
export async function registServiceQuery(params: {
  tenantId?: string
  serviceName?: string
  groupName?: string
  protocolType?: string
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return proxyApi.post('/registServiceQuery', params)
}
