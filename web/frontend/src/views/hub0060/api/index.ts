/**
 * 隧道服务器管理模块API
 * 提供隧道服务器的增删改查、状态管理、连接测试等功能的接口定义
 * 
 * 所有API请求都统一在 /gateway/hub0060 路径下，遵循RESTful API设计规范
 * 使用标准HTTP POST方法：
 * 
 * 隧道服务器管理API：
 * - POST /queryTunnelServers - 查询隧道服务器列表（支持分页、搜索和过滤）
 * - POST /getTunnelServer - 获取隧道服务器详情
 * - POST /createTunnelServer - 创建隧道服务器
 * - POST /updateTunnelServer - 更新隧道服务器信息
 * - POST /deleteTunnelServer - 删除隧道服务器
 * - POST /startTunnelServer - 启动隧道服务器
 * - POST /stopTunnelServer - 停止隧道服务器
 * - POST /restartTunnelServer - 重启隧道服务器
 * - POST /getTunnelServerStats - 获取隧道服务器统计信息
 * - POST /generateAuthToken - 生成认证令牌
 * - POST /getRegisteredClients - 获取已注册的客户端列表
 * - POST /testServerConnection - 测试服务器连接
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
  TunnelServerForm,
  TunnelServerQueryParams
} from '../types'

// 创建API实例 - 所有请求都在hub0060路径下
const tunnelServerApi = createApi('/gateway/hub0060')

// ==================== 隧道服务器管理API ====================

/**
 * 查询隧道服务器列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 隧道服务器列表
 */
export const queryTunnelServers = async (params: TunnelServerQueryParams): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/queryTunnelServers', params)
}

/**
 * 获取隧道服务器详情
 * @param tunnelServerId 隧道服务器ID
 * @returns 隧道服务器详情
 */
export const getTunnelServerDetail = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/getTunnelServer', { tunnelServerId })
}

/**
 * 创建隧道服务器
 * @param data 隧道服务器数据
 * @returns 创建结果
 */
export const createTunnelServer = async (data: TunnelServerForm): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/createTunnelServer', data)
}

/**
 * 更新隧道服务器信息
 * @param data 隧道服务器数据（包含tunnelServerId）
 * @returns 更新结果
 */
export const updateTunnelServer = async (data: TunnelServerForm & { tunnelServerId: string }): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/updateTunnelServer', data)
}

/**
 * 删除隧道服务器
 * @param tunnelServerId 隧道服务器ID
 * @returns 删除结果
 */
export const deleteTunnelServer = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/deleteTunnelServer', { tunnelServerId })
}

/**
 * 测试服务器连接
 * @param tunnelServerId 隧道服务器ID
 * @returns 连接测试结果
 */
export const testTunnelServerConnection = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/testServerConnection', { tunnelServerId })
}

/**
 * 获取隧道服务器统计信息
 * @returns 统计信息
 */
export const getTunnelServerStats = async (): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/getTunnelServerStats', {})
}

/**
 * 生成认证令牌
 * @returns 认证令牌
 */
export const generateAuthToken = async (): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/generateAuthToken', {})
}

/**
 * 获取已注册的客户端列表
 * @param tunnelServerId 隧道服务器ID (可选)
 * @returns 已注册的客户端列表
 */
export const getRegisteredClients = async (tunnelServerId?: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/getRegisteredClients', { tunnelServerId })
}

/**
 * 获取已注册的服务列表
 * @param tunnelServerId 隧道服务器ID (可选)
 * @returns 已注册的服务列表
 */
export const getRegisteredServices = async (tunnelServerId?: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/getRegisteredServices', { tunnelServerId })
}

// ==================== 服务器操作API ====================

/**
 * 启动隧道服务器
 * @param tunnelServerId 隧道服务器ID
 * @returns 启动结果
 */
export const startTunnelServer = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/startTunnelServer', { tunnelServerId })
}

/**
 * 停止隧道服务器
 * @param tunnelServerId 隧道服务器ID
 * @returns 停止结果
 */
export const stopTunnelServer = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/stopTunnelServer', { tunnelServerId })
}

/**
 * 重启隧道服务器
 * @param tunnelServerId 隧道服务器ID
 * @returns 重启结果
 */
export const restartTunnelServer = async (tunnelServerId: string): Promise<JsonDataObj> => {
  return tunnelServerApi.post('/restartTunnelServer', { tunnelServerId })
}

/**
 * 批量删除隧道服务器
 * @param tunnelServerIds 隧道服务器ID列表
 * @returns 删除结果
 */
export const batchDeleteTunnelServers = async (tunnelServerIds: string[]): Promise<JsonDataObj[]> => {
  return Promise.all(tunnelServerIds.map(id => deleteTunnelServer(id)))
}
