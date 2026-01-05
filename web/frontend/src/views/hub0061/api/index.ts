/**
 * 静态端口映射管理模块API
 * 提供静态服务器和静态节点的增删改查等功能
 * 
 * API路径: /gateway/hub0061
 * 
 * 静态服务器管理API：
 * - POST /queryStaticServers - 查询静态服务器列表
 * - POST /getStaticServer - 获取静态服务器详情
 * - POST /createStaticServer - 创建静态服务器
 * - POST /updateStaticServer - 更新静态服务器
 * - POST /deleteStaticServer - 删除静态服务器
 * - POST /getStaticServerStats - 获取服务器统计信息
 * - POST /checkServerPortConflict - 检查端口冲突
 * - POST /startStaticServer - 启动服务器
 * - POST /stopStaticServer - 停止服务器
 * - POST /reloadStaticServer - 重载服务器配置
 * 
 * 静态节点管理API：
 * - POST /queryStaticNodes - 查询静态节点列表
 * - POST /getStaticNode - 获取静态节点详情
 * - POST /createStaticNode - 创建静态节点
 * - POST /updateStaticNode - 更新静态节点
 * - POST /deleteStaticNode - 删除静态节点
 * - POST /getStaticNodeStats - 获取节点统计信息
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
  PortConflictCheckParams,
  StaticNodeQueryParams,
  StaticServerQueryParams,
  TunnelStaticNode,
  TunnelStaticServer,
} from '../types'

// 创建API实例
const staticApi = createApi('/gateway/hub0061')

// ==================== 静态服务器管理API ====================

/**
 * 查询静态服务器列表
 * @param params 查询参数
 * @returns 静态服务器列表
 */
export const queryStaticServers = async (params: StaticServerQueryParams): Promise<JsonDataObj> => {
  return staticApi.post('/queryStaticServers', params)
}

/**
 * 获取静态服务器详情
 * @param tunnelStaticServerId 静态服务器ID
 * @returns 静态服务器详情
 */
export const getStaticServer = async (tunnelStaticServerId: string): Promise<JsonDataObj> => {
  return staticApi.post('/getStaticServer', { tunnelStaticServerId })
}

/**
 * 创建静态服务器
 * @param data 静态服务器数据
 * @returns 创建结果
 */
export const createStaticServer = async (data: Partial<TunnelStaticServer>): Promise<JsonDataObj> => {
  return staticApi.post('/createStaticServer', data)
}

/**
 * 更新静态服务器
 * @param data 静态服务器数据（包含tunnelStaticServerId）
 * @returns 更新结果
 */
export const updateStaticServer = async (data: Partial<TunnelStaticServer> & { tunnelStaticServerId: string }): Promise<JsonDataObj> => {
  return staticApi.post('/updateStaticServer', data)
}

/**
 * 删除静态服务器
 * @param tunnelStaticServerId 静态服务器ID
 * @returns 删除结果
 */
export const deleteStaticServer = async (tunnelStaticServerId: string): Promise<JsonDataObj> => {
  return staticApi.post('/deleteStaticServer', { tunnelStaticServerId })
}

/**
 * 获取静态服务器统计信息
 * @returns 统计信息
 */
export const getStaticServerStats = async (): Promise<JsonDataObj> => {
  return staticApi.post('/getStaticServerStats', {})
}

/**
 * 检查端口冲突
 * @param params 端口冲突检查参数
 * @returns 冲突检查结果
 */
export const checkServerPortConflict = async (params: PortConflictCheckParams): Promise<JsonDataObj> => {
  return staticApi.post('/checkServerPortConflict', params)
}

/**
 * 启动静态服务器
 * @param tunnelStaticServerId 静态服务器ID
 * @returns 启动结果（包含最新服务器信息）
 */
export const startStaticServer = async (tunnelStaticServerId: string): Promise<JsonDataObj> => {
  return staticApi.post('/startStaticServer', { tunnelStaticServerId })
}

/**
 * 停止静态服务器
 * @param tunnelStaticServerId 静态服务器ID
 * @returns 停止结果（包含最新服务器信息）
 */
export const stopStaticServer = async (tunnelStaticServerId: string): Promise<JsonDataObj> => {
  return staticApi.post('/stopStaticServer', { tunnelStaticServerId })
}

/**
 * 重载静态服务器配置
 * @param tunnelStaticServerId 静态服务器ID
 * @returns 重载结果（包含最新服务器信息）
 */
export const reloadStaticServer = async (tunnelStaticServerId: string): Promise<JsonDataObj> => {
  return staticApi.post('/reloadStaticServer', { tunnelStaticServerId })
}

// ==================== 静态节点管理API ====================

/**
 * 查询静态节点列表
 * @param params 查询参数
 * @returns 静态节点列表
 */
export const queryStaticNodes = async (params: StaticNodeQueryParams): Promise<JsonDataObj> => {
  return staticApi.post('/queryStaticNodes', params)
}

/**
 * 获取静态节点详情
 * @param tunnelStaticNodeId 静态节点ID
 * @returns 静态节点详情
 */
export const getStaticNode = async (tunnelStaticNodeId: string): Promise<JsonDataObj> => {
  return staticApi.post('/getStaticNode', { tunnelStaticNodeId })
}

/**
 * 创建静态节点
 * @param data 静态节点数据
 * @returns 创建结果
 */
export const createStaticNode = async (data: Partial<TunnelStaticNode>): Promise<JsonDataObj> => {
  return staticApi.post('/createStaticNode', data)
}

/**
 * 更新静态节点
 * @param data 静态节点数据（包含tunnelStaticNodeId）
 * @returns 更新结果
 */
export const updateStaticNode = async (data: Partial<TunnelStaticNode> & { tunnelStaticNodeId: string }): Promise<JsonDataObj> => {
  return staticApi.post('/updateStaticNode', data)
}

/**
 * 删除静态节点
 * @param tunnelStaticNodeId 静态节点ID
 * @returns 删除结果
 */
export const deleteStaticNode = async (tunnelStaticNodeId: string): Promise<JsonDataObj> => {
  return staticApi.post('/deleteStaticNode', { tunnelStaticNodeId })
}

/**
 * 获取静态节点统计信息
 * @param tunnelStaticServerId 可选的服务器ID，用于获取特定服务器的统计
 * @returns 统计信息
 */
export const getStaticNodeStats = async (tunnelStaticServerId?: string): Promise<JsonDataObj> => {
  return staticApi.post('/getStaticNodeStats', { tunnelStaticServerId })
}
