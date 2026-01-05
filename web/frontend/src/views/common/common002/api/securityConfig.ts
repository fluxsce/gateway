import type { JsonDataObj } from '@/types/api'
import { createApi } from '@/api/request'

// 创建安全配置API实例，强制使用JSON格式
const securityApi = createApi('/gateway/hubcommon002')

/**
 * 分页查询安全配置列表
 * @param params 查询参数
 * @returns 安全配置列表和分页信息
 */
export async function querySecurityConfigs(params: any): Promise<JsonDataObj> {
  return securityApi.post('/querySecurityConfigs', params)
}

/**
 * 查询安全配置详情
 * @param securityConfigId 安全配置ID
 * @param tenantId 租户ID (可选，后端会话自动处理)
 * @returns 安全配置详情
 */
export async function getSecurityConfig(
  securityConfigId: string,
  tenantId?: string,
): Promise<JsonDataObj> {
  return securityApi.get('/getSecurityConfig', { securityConfigId, ...(tenantId && { tenantId }) })
}

/**
 * 添加安全配置
 * @param data 安全配置创建数据
 * @returns 操作结果
 */
export async function addSecurityConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/addSecurityConfig', data)
}

/**
 * 编辑安全配置
 * @param data 安全配置更新数据
 * @returns 操作结果
 */
export async function editSecurityConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/editSecurityConfig', data)
}

/**
 * 删除安全配置
 * @param securityConfigId 安全配置ID
 * @param tenantId 租户ID (可选，后端会话自动处理)
 * @returns 操作结果
 */
export async function deleteSecurityConfig(
  securityConfigId: string,
  tenantId?: string,
): Promise<JsonDataObj> {
  return securityApi.post('/deleteSecurityConfig', {
    securityConfigId,
    ...(tenantId && { tenantId }),
  })
}

/**
 * 根据网关实例查询安全配置
 * @param params 查询参数
 * @returns 安全配置列表
 */
export async function querySecurityConfigsByGatewayInstance(params: {
  gatewayInstanceId: string
  tenantId?: string // 可选，后端会话自动处理
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/querySecurityConfigsByGatewayInstance', params)
}

/**
 * 根据路由配置查询安全配置
 * @param params 查询参数
 * @returns 安全配置列表
 */
export async function querySecurityConfigsByRouteConfig(params: {
  routeConfigId: string
  tenantId?: string // 可选，后端会话自动处理
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/querySecurityConfigsByRouteConfig', params)
}

// ===== IP访问控制配置模块 =====

/**
 * 添加IP访问控制配置
 * @param data IP访问控制配置数据
 * @returns 操作结果
 */
export async function addIpAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/ip-access/add', data)
}

/**
 * 获取IP访问控制配置详情
 * @param params 查询参数
 * @returns IP访问控制配置详情
 */
export async function getIpAccessConfig(params: {
  ipAccessConfigId?: string
  securityConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/ip-access/get', params)
}

/**
 * 更新IP访问控制配置
 * @param data IP访问控制配置更新数据
 * @returns 操作结果
 */
export async function updateIpAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/ip-access/update', data)
}

/**
 * 删除IP访问控制配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteIpAccessConfig(params: {
  ipAccessConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/ip-access/delete', params)
}

/**
 * 查询IP访问控制配置列表
 * @param params 查询参数
 * @returns IP访问控制配置列表
 */
export async function queryIpAccessConfigs(params: {
  securityConfigId?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/ip-access/query', params)
}

// ===== User-Agent访问控制配置模块 =====

/**
 * 添加User-Agent访问控制配置
 * @param data User-Agent访问控制配置数据
 * @returns 操作结果
 */
export async function addUseragentAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/useragent-access/add', data)
}

/**
 * 获取User-Agent访问控制配置详情
 * @param params 查询参数
 * @returns User-Agent访问控制配置详情
 */
export async function getUseragentAccessConfig(params: {
  useragentAccessConfigId?: string
  securityConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/useragent-access/get', params)
}

/**
 * 更新User-Agent访问控制配置
 * @param data User-Agent访问控制配置更新数据
 * @returns 操作结果
 */
export async function updateUseragentAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/useragent-access/update', data)
}

/**
 * 删除User-Agent访问控制配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteUseragentAccessConfig(params: {
  useragentAccessConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/useragent-access/delete', params)
}

/**
 * 查询User-Agent访问控制配置列表
 * @param params 查询参数
 * @returns User-Agent访问控制配置列表
 */
export async function queryUseragentAccessConfigs(params: {
  securityConfigId?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/useragent-access/query', params)
}

// ===== API访问控制配置模块 =====

/**
 * 添加API访问控制配置
 * @param data API访问控制配置数据
 * @returns 操作结果
 */
export async function addApiAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/api-access/add', data)
}

/**
 * 获取API访问控制配置详情
 * @param params 查询参数
 * @returns API访问控制配置详情
 */
export async function getApiAccessConfig(params: {
  apiAccessConfigId?: string
  securityConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/api-access/get', params)
}

/**
 * 更新API访问控制配置
 * @param data API访问控制配置更新数据
 * @returns 操作结果
 */
export async function updateApiAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/api-access/update', data)
}

/**
 * 删除API访问控制配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteApiAccessConfig(params: {
  apiAccessConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/api-access/delete', params)
}

/**
 * 查询API访问控制配置列表
 * @param params 查询参数
 * @returns API访问控制配置列表
 */
export async function queryApiAccessConfigs(params: {
  securityConfigId?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/api-access/query', params)
}

// ===== 域名访问控制配置模块 =====

/**
 * 添加域名访问控制配置
 * @param data 域名访问控制配置数据
 * @returns 操作结果
 */
export async function addDomainAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/domain-access/add', data)
}

/**
 * 获取域名访问控制配置详情
 * @param params 查询参数
 * @returns 域名访问控制配置详情
 */
export async function getDomainAccessConfig(params: {
  domainAccessConfigId?: string
  securityConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/domain-access/get', params)
}

/**
 * 更新域名访问控制配置
 * @param data 域名访问控制配置更新数据
 * @returns 操作结果
 */
export async function updateDomainAccessConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/domain-access/update', data)
}

/**
 * 删除域名访问控制配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteDomainAccessConfig(params: {
  domainAccessConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/domain-access/delete', params)
}

/**
 * 查询域名访问控制配置列表
 * @param params 查询参数
 * @returns 域名访问控制配置列表
 */
export async function queryDomainAccessConfigs(params: {
  securityConfigId?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/domain-access/query', params)
}

// ===== CORS配置模块 =====

/**
 * 添加CORS配置
 * @param data CORS配置数据
 * @returns 操作结果
 */
export async function addCorsConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/cors/add', data)
}

/**
 * 获取CORS配置详情
 * @param params 查询参数
 * @returns CORS配置详情
 */
export async function getCorsConfig(params: {
  corsConfigId?: string
  securityConfigId?: string
  gatewayInstanceId?: string
  routeConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/cors/get', params)
}

/**
 * 更新CORS配置
 * @param data CORS配置更新数据
 * @returns 操作结果
 */
export async function updateCorsConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/cors/update', data)
}

/**
 * 删除CORS配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteCorsConfig(params: {
  corsConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/cors/delete', params)
}

/**
 * 查询CORS配置列表
 * @param params 查询参数
 * @returns CORS配置列表
 */
export async function queryCorsConfigs(params: {
  securityConfigId?: string
  gatewayInstanceId?: string
  routeConfigId?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/cors/query', params)
}

// ===== 认证配置模块 =====

/**
 * 添加认证配置
 * @param data 认证配置数据
 * @returns 操作结果
 */
export async function addAuthConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/auth/add', data)
}

/**
 * 获取认证配置详情
 * @param params 查询参数
 * @returns 认证配置详情
 */
export async function getAuthConfig(params: {
  authConfigId?: string
  gatewayInstanceId?: string
  routeConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/auth/get', params)
}

/**
 * 更新认证配置
 * @param data 认证配置更新数据
 * @returns 操作结果
 */
export async function updateAuthConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/auth/update', data)
}

/**
 * 删除认证配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteAuthConfig(params: {
  authConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/auth/delete', params)
}

/**
 * 查询认证配置列表
 * @param params 查询参数
 * @returns 认证配置列表
 */
export async function queryAuthConfigs(params: {
  gatewayInstanceId?: string
  routeConfigId?: string
  authType?: string
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/auth/query', params)
}

// ===== 限流配置模块 =====

/**
 * 添加限流配置
 * @param data 限流配置数据
 * @returns 操作结果
 */
export async function addRateLimitConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/rate-limit/add', data)
}

/**
 * 获取限流配置详情
 * @param params 查询参数
 * @returns 限流配置详情
 */
export async function getRateLimitConfig(params: {
  rateLimitConfigId?: string
  gatewayInstanceId?: string
  routeConfigId?: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/rate-limit/get', params)
}

/**
 * 更新限流配置
 * @param data 限流配置更新数据
 * @returns 操作结果
 */
export async function updateRateLimitConfig(data: any): Promise<JsonDataObj> {
  return securityApi.post('/rate-limit/update', data)
}

/**
 * 删除限流配置
 * @param params 删除参数
 * @returns 操作结果
 */
export async function deleteRateLimitConfig(params: {
  rateLimitConfigId: string
  tenantId?: string
}): Promise<JsonDataObj> {
  return securityApi.post('/rate-limit/delete', params)
}

/**
 * 查询限流配置列表
 * @param params 查询参数
 * @returns 限流配置列表
 */
export async function queryRateLimitConfigs(params: {
  gatewayInstanceId?: string
  routeConfigId?: string
  limitType?: number
  tenantId?: string
  pageNo?: number
  pageSize?: number
}): Promise<JsonDataObj> {
  return securityApi.post('/rate-limit/query', params)
}
