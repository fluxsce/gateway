import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { RouteAssertion } from '../components/assert-config/hooks/types'
import type { GatewayInstance, RouterConfig } from '../components/instance-tree/types'
import type { RouteConfig } from '../components/routes/types'

const routeApi = createApi('/gateway/hub0021')

/**
 * 查询所有网关实例
 * @param params 查询参数
 * @returns 网关实例列表
 */
export async function queryAllGatewayInstances(
  params: Partial<Pick<GatewayInstance, 'instanceName' | 'activeFlag'>> & {
    pageIndex?: number
    pageSize?: number
  } = {},
): Promise<JsonDataObj> {
  return routeApi.post('/queryAllGatewayInstances', params)
}

/**
 * 分页查询路由配置列表
 * @param params 查询参数
 * @returns 路由配置列表和分页信息
 */
export async function queryRouteConfigs(params: {
  gatewayInstanceId?: string
  routeName?: string
  routePath?: string
  matchType?: number
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryRouteConfigs', params)
}

/**
 * 查询路由配置详情
 * @param routeConfigId 路由配置ID
 * @returns 路由配置详情
 */
export async function getRouteConfig(
  routeConfigId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/getRouteConfig', { routeConfigId })
}

/**
 * 根据网关实例获取路由配置列表
 * @param gatewayInstanceId 网关实例ID
 * @returns 路由配置列表
 */
export async function getRouteConfigsByInstance(
  gatewayInstanceId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/routeConfigs/byInstance', { gatewayInstanceId })
}

/**
 * 添加路由配置
 * @param data 路由配置创建数据
 * @returns 操作结果
 */
export async function addRouteConfig(
  data: Partial<RouteConfig> & { gatewayInstanceId: string },
): Promise<JsonDataObj> {
  return routeApi.post('/addRouteConfig', data)
}

/**
 * 编辑路由配置
 * @param data 路由配置更新数据
 * @returns 操作结果
 */
export async function editRouteConfig(
  data: Partial<RouteConfig> & {
    routeConfigId: string
    gatewayInstanceId: string
  },
): Promise<JsonDataObj> {
  return routeApi.post('/editRouteConfig', data)
}

/**
 * 删除路由配置
 * @param routeConfigId 路由配置ID
 * @returns 操作结果
 */
export async function deleteRouteConfig(
  routeConfigId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/deleteRouteConfig', { routeConfigId })
}

/**
 * 添加路由断言
 * @param data 路由断言创建数据
 * @returns 操作结果
 */
export async function addRouteAssertion(
  data: Partial<RouteAssertion> & { routeConfigId: string },
): Promise<JsonDataObj> {
  return routeApi.post('/addRouteAssertion', data)
}

/**
 * 根据断言ID获取单个断言配置
 * @param routeAssertionId 路由断言ID
 * @returns 路由断言配置
 */
export async function getRouteAssertionById(
  routeAssertionId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/getRouteAssertionById', { routeAssertionId })
}

/**
 * 分页查询路由断言列表
 * @param params 查询参数（支持分页和多条件筛选）
 * @returns 路由断言列表和分页信息
 */
export async function queryRouteAssertions(params: {
  routeConfigId?: string
  assertionName?: string
  assertionType?: string
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryRouteAssertions', params)
}

/**
 * 编辑路由断言
 * @param data 路由断言更新数据
 * @returns 操作结果
 */
export async function editRouteAssertion(
  data: Partial<RouteAssertion> & {
    routeAssertionId: string
    routeConfigId: string
  },
): Promise<JsonDataObj> {
  return routeApi.post('/editRouteAssertion', data)
}

/**
 * 删除路由断言
 * @param routeAssertionId 路由断言ID
 * @returns 操作结果
 */
export async function deleteRouteAssertion(
  routeAssertionId: string,
): Promise<JsonDataObj> {
  return routeApi.delete('/deleteRouteAssertion', { routeAssertionId })
}

/**
 * 分页查询Router配置列表
 * @param params 查询参数
 * @returns Router配置列表和分页信息
 */
export async function queryRouterConfigs(params: {
  gatewayInstanceId?: string
  routerName?: string
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryRouterConfigs', params)
}

/**
 * 查询Router配置详情
 * @param routerConfigId Router配置ID
 * @returns Router配置详情
 */
export async function getRouterConfig(
  routerConfigId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/routerConfig', { routerConfigId })
}

/**
 * 根据网关实例获取Router配置列表
 * @param gatewayInstanceId 网关实例ID
 * @returns Router配置列表
 */
export async function getRouterConfigsByInstance(gatewayInstanceId: string): Promise<JsonDataObj> {
  return routeApi.post('/routerConfigs/byInstance', { gatewayInstanceId })
}

/**
 * 添加Router配置
 * @param data Router配置创建数据
 * @returns 操作结果
 */
export async function addRouterConfig(
  data: Partial<RouterConfig> & { gatewayInstanceId: string },
): Promise<JsonDataObj> {
  return routeApi.post('/addRouterConfig', data)
}

/**
 * 编辑Router配置
 * @param data Router配置更新数据
 * @returns 操作结果
 */
export async function editRouterConfig(
  data: Partial<RouterConfig> & {
    routerConfigId: string
    gatewayInstanceId: string
  },
): Promise<JsonDataObj> {
  return routeApi.post('/editRouterConfig', data)
}

/**
 * 删除Router配置
 * @param routerConfigId Router配置ID
 * @returns 操作结果
 */
export async function deleteRouterConfig(
  routerConfigId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/deleteRouterConfig', { routerConfigId })
}

/**
 * 根据服务定义ID获取服务定义详情
 * @param serviceDefinitionId 服务定义ID
 * @returns 服务定义详情
 */
export async function getServiceDefinitionById(
  serviceDefinitionId: string
): Promise<JsonDataObj> {
  return routeApi.post('/getServiceDefinitionById', { serviceDefinitionId })
}

/**
 * 分页查询服务定义列表
 * @param params 查询参数（包含分页、筛选条件等）
 * @returns 服务定义列表和分页信息
 */
export async function queryServiceDefinitions(params: {
  gatewayInstanceId?: string
  serviceName?: string
  serviceDefinitionId?: string
  serviceType?: number
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryServiceDefinitions', params)
}

/**
 * 查询所有服务定义（不依赖代理配置）
 * @param params 查询参数（包含分页、筛选条件等，不强制要求代理配置ID）
 * @returns 服务定义列表和分页信息
 */
export async function queryAllServiceDefinitions(params: {
  serviceName?: string
  serviceDefinitionId?: string
  serviceType?: number
  loadBalanceStrategy?: string
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryAllServiceDefinitions', params)
}

/**
 * 获取路由统计信息
 * @param params 查询参数
 * @returns 路由统计信息
 */
export async function queryRouteStatistics(params: {
  gatewayInstanceId?: string
  routeName?: string
  routePath?: string
  matchType?: number
  activeFlag?: string
}): Promise<JsonDataObj> {
  return routeApi.post('/routeStatistics', params)
}

/**
 * 分页查询过滤器配置列表（支持多参数）
 * @param params 查询参数
 * @returns 过滤器配置列表和分页信息
 */
export async function queryFilterConfigs(params: {
  gatewayInstanceId?: string
  routeConfigId?: string
  filterName?: string
  filterType?: string
  filterAction?: string
  activeFlag?: string
  pageIndex?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return routeApi.post('/queryFilterConfigs', params)
}

/**
 * 查询过滤器配置详情
 * @param filterConfigId 过滤器配置ID
 * @returns 过滤器配置详情
 */
export async function getFilterConfig(
  filterConfigId: string
): Promise<JsonDataObj> {
  return routeApi.post('/getFilterConfig', {
    filterConfigId
  })
}

/**
 * 添加过滤器配置
 * @param filterConfig 过滤器配置数据
 * @returns 操作结果
 */
export async function addFilterConfig(filterConfig: any): Promise<JsonDataObj> {
  return routeApi.post('/addFilterConfig', filterConfig)
}

/**
 * 编辑过滤器配置
 * @param filterConfig 过滤器配置数据
 * @returns 操作结果
 */
export async function editFilterConfig(filterConfig: any): Promise<JsonDataObj> {
  return routeApi.post('/editFilterConfig', filterConfig)
}

/**
 * 删除过滤器配置
 * @param filterConfigId 过滤器配置ID
 * @returns 操作结果
 */
export async function deleteFilterConfig(
  filterConfigId: string,
): Promise<JsonDataObj> {
  return routeApi.post('/deleteFilterConfig', {
    filterConfigId,
  })
}
