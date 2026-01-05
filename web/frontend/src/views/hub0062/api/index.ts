/**
 * 客户端管理模块API
 * 提供客户端和服务的增删改查、启停控制、关联数据查询等功能的接口定义
 * 
 * 所有API请求都统一在 /gateway/hub0062 路径下，遵循RESTful API设计规范
 * 使用标准HTTP POST方法：
 * 
 * 客户端基础CRUD操作：
 * - POST /queryTunnelClients - 查询客户端列表（支持分页、搜索和过滤）
 * - POST /getTunnelClient - 获取客户端详情
 * - POST /createTunnelClient - 创建客户端
 * - POST /updateTunnelClient - 更新客户端信息
 * - POST /deleteTunnelClient - 删除客户端
 * 
 * 客户端管理操作：
 * - POST /startClient - 启动客户端
 * - POST /stopClient - 停止客户端
 * - POST /restartClient - 重启客户端
 * 
 * 客户端统计和关联数据：
 * - POST /getClientStats - 获取客户端统计信息
 * - POST /getClientServices - 获取客户端注册的服务列表
 * 
 * 服务基础CRUD操作：
 * - POST /queryTunnelServices - 查询服务列表（支持分页、搜索和过滤）
 * - POST /getTunnelService - 获取服务详情
 * - POST /createTunnelService - 创建服务
 * - POST /updateTunnelService - 更新服务信息
 * - POST /deleteTunnelService - 删除服务
 * 
 * 服务管理操作：
 * - POST /registerService - 注册服务
 * - POST /unregisterService - 注销服务
 * 
 * 服务统计：
 * - POST /getServiceStats - 获取服务统计信息
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
  TunnelClient,
  TunnelClientQueryParams
} from '../types'

// 创建API实例 - 所有请求都在hub0062路径下
const tunnelClientApi = createApi('/gateway/hub0062')

// ==================== 基础CRUD操作 ====================

/**
 * 查询客户端列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 客户端列表
 */
export const queryTunnelClients = async (params: TunnelClientQueryParams): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/queryTunnelClients', params)
}

/**
 * 获取客户端详情
 * @param tunnelClientId 客户端ID
 * @returns 客户端详情
 */
export const getTunnelClient = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/getTunnelClient', { tunnelClientId })
}

/**
 * 创建客户端
 * @param data 客户端数据
 * @returns 创建结果
 */
export const createTunnelClient = async (data: TunnelClient): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/createTunnelClient', data)
}

/**
 * 更新客户端信息
 * @param data 客户端数据（包含tunnelClientId）
 * @returns 更新结果
 */
export const updateTunnelClient = async (data: TunnelClient & { tunnelClientId: string }): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/updateTunnelClient', data)
}

/**
 * 删除客户端
 * @param tunnelClientId 客户端ID
 * @returns 删除结果
 */
export const deleteTunnelClient = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/deleteTunnelClient', { tunnelClientId })
}

// ==================== 客户端管理操作 ====================

/**
 * 启动客户端
 * @param tunnelClientId 客户端ID
 * @returns 启动结果
 */
export const startClient = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/startClient', { tunnelClientId })
}

/**
 * 停止客户端
 * @param tunnelClientId 客户端ID
 * @returns 停止结果
 */
export const stopClient = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/stopClient', { tunnelClientId })
}

/**
 * 重启客户端
 * @param tunnelClientId 客户端ID
 * @returns 重启结果
 */
export const restartClient = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/restartClient', { tunnelClientId })
}

// ==================== 统计和关联数据 ====================

/**
 * 获取客户端统计信息
 * @returns 统计信息
 */
export const getClientStats = async (): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/getClientStats', {})
}

// ==================== 关联数据查询 ====================

/**
 * 获取客户端注册的服务列表
 * @param tunnelClientId 客户端ID
 * @returns 服务列表
 */
export const getClientServices = async (tunnelClientId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/getClientServices', { tunnelClientId })
}

// ==================== 服务基础CRUD操作 ====================

/**
 * 查询服务列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 服务列表
 */
export const queryTunnelServices = async (params: any): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/queryTunnelServices', params)
}

/**
 * 获取服务详情
 * @param tunnelServiceId 服务ID
 * @returns 服务详情
 */
export const getTunnelService = async (tunnelServiceId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/getTunnelService', { tunnelServiceId })
}

/**
 * 创建服务
 * @param data 服务数据
 * @returns 创建结果
 */
export const createTunnelService = async (data: any): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/createTunnelService', data)
}

/**
 * 更新服务信息
 * @param data 服务数据（包含tunnelServiceId）
 * @returns 更新结果
 */
export const updateTunnelService = async (data: any): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/updateTunnelService', data)
}

/**
 * 删除服务
 * @param tunnelServiceId 服务ID
 * @returns 删除结果
 */
export const deleteTunnelService = async (tunnelServiceId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/deleteTunnelService', { tunnelServiceId })
}

// ==================== 服务管理操作 ====================

/**
 * 注册服务
 * @param tunnelServiceId 服务ID
 * @returns 注册结果
 */
export const registerService = async (tunnelServiceId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/registerService', { tunnelServiceId })
}

/**
 * 注销服务
 * @param tunnelServiceId 服务ID
 * @returns 注销结果
 */
export const unregisterService = async (tunnelServiceId: string): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/unregisterService', { tunnelServiceId })
}

// ==================== 服务统计 ====================

/**
 * 获取服务统计信息
 * @returns 统计信息
 */
export const getServiceStats = async (): Promise<JsonDataObj> => {
  return tunnelClientApi.post('/getServiceStats', {})
}
