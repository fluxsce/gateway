/**
 * Hub0042 JVM监控模块 API 接口
 * 提供JVM资源监控、内存监控、GC监控、线程监控、线程池监控等功能的接口定义
 * 
 * 所有API请求都统一在 /gateway/hub0042 路径下
 * 使用标准HTTP POST方法
 * 
 * JVM资源管理API：
 * - POST /queryJvmResources - 查询JVM资源列表（支持分页、搜索和过滤）
 * - POST /getJvmResourceDetail - 获取JVM资源详情
 * - POST /getJvmOverview - 获取JVM监控概览
 * 
 * 内存监控API：
 * - POST /queryMemory - 查询内存使用情况
 * - POST /queryMemoryPools - 查询内存池列表
 * 
 * GC监控API：
 * - POST /queryGCSnapshots - 查询GC快照列表
 * - POST /getLatestGCSnapshot - 获取最新GC快照
 * 
 * 线程监控API：
 * - POST /queryThreads - 查询线程信息
 * - POST /queryDeadlocks - 查询死锁信息
 * 
 * 应用监控数据API：
 * - POST /queryAppMonitorData - 查询应用监控数据列表
 * - POST /getAppMonitorDataDetail - 获取应用监控数据详情
 * 
 * 线程池监控API：
 * - POST /queryThreadPools - 查询线程池列表
 * 
 * 类加载监控API：
 * - POST /queryClassLoading - 查询类加载信息
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { 
  JvmResource,
  JvmMemory,
  MemoryPool,
  JvmGc,
  JvmThread,
  JvmThreadState,
  JvmDeadlock,
  JvmThreadPool,
  JvmClass,
  JvmMonitorDetail,
  JvmResourceQueryRequest,
  MemoryQueryRequest,
  GcQueryRequest,
  ThreadQueryRequest,
  ThreadPoolQueryRequest,
  AppDataQueryRequest,
  GcTrendData,
  MemoryTrendData,
  ThreadPoolTrendData
} from '../types'

// 创建API实例 - 所有请求都在hub0042路径下
const jvmMonitorApi = createApi('/gateway/hub0042')

// ==================== JVM资源管理API ====================

/**
 * 查询JVM资源列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns JVM资源列表
 */
export const queryJvmResources = async (params: JvmResourceQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryJvmResources', params)
}

/**
 * 获取JVM资源详情
 * @param jvmResourceId JVM资源ID
 * @returns JVM资源详情
 */
export const getJvmResource = async (jvmResourceId: string): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getJvmResourceDetail', { jvmResourceId })
}

/**
 * 获取JVM监控概览
 * @param jvmResourceId JVM资源ID
 * @returns JVM监控概览信息
 */
export const getJvmOverview = async (jvmResourceId: string): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getJvmOverview', { jvmResourceId })
}

// ==================== 内存监控API ====================

/**
 * 查询内存使用情况
 * @param params 查询参数
 * @returns 内存使用情况
 */
export const queryMemoryUsage = async (params: MemoryQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryMemory', params)
}

/**
 * 查询内存池列表
 * @param params 查询参数
 * @returns 内存池列表
 */
export const queryMemoryPools = async (params: MemoryQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryMemoryPools', params)
}


// ==================== GC监控API ====================

/**
 * 查询GC快照列表
 * @param params 查询参数
 * @returns GC快照列表
 */
export const queryGcSnapshots = async (params: GcQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryGCSnapshots', params)
}

/**
 * 获取最新GC快照
 * @param jvmResourceId JVM资源ID
 * @returns 最新GC快照
 */
export const getLatestGcSnapshot = async (jvmResourceId: string): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getLatestGCSnapshot', { jvmResourceId })
}

/**
 * 获取GC趋势数据
 * @param params 查询参数
 * @returns GC趋势数据
 */
export const getGcTrend = async (params: GcQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getGcTrend', params)
}

// ==================== 线程监控API ====================

/**
 * 查询线程信息
 * @param params 查询参数
 * @returns 线程信息
 */
export const queryThreadInfo = async (params: ThreadQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryThreads', params)
}

/**
 * 获取线程状态统计
 * @param jvmThreadId 线程记录ID
 * @returns 线程状态统计信息
 */
export const getThreadState = async (jvmThreadId: string): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getThreadState', { jvmThreadId })
}

/**
 * 查询线程状态列表（支持时间范围）
 * @param params 查询参数
 * @returns 线程状态列表
 */
export const queryThreadStates = async (params: ThreadQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryThreadStates', params)
}

/**
 * 查询死锁信息
 * @param params 查询参数
 * @returns 死锁信息列表
 */
export const queryDeadlocks = async (params: ThreadQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryDeadlocks', params)
}

// ==================== 应用监控数据API ====================

/**
 * 查询应用监控数据列表（支持分页、搜索和过滤）
 * @param params 查询参数
 * @returns 应用监控数据列表
 */
export const queryAppMonitorData = async (params: AppDataQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryAppMonitorData', params)
}

/**
 * 获取应用监控数据详情
 * @param appDataId 应用监控数据ID
 * @returns 应用监控数据详情
 */
export const getAppMonitorDataDetail = async (appDataId: string): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/getAppMonitorDataDetail', { appDataId })
}

// ==================== 线程池监控API ====================

/**
 * 查询线程池列表
 * @param params 查询参数
 * @returns 线程池列表
 */
export const queryThreadPools = async (params: ThreadPoolQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryThreadPools', params)
}


// ==================== 类加载监控API ====================

/**
 * 查询类加载信息
 * @param params 查询参数
 * @returns 类加载信息
 */
export const queryClassLoading = async (params: ThreadQueryRequest): Promise<JsonDataObj> => {
  return jvmMonitorApi.post('/queryClassLoading', params)
}


// ==================== 导出统一API ====================

export default {
  // JVM资源管理
  queryJvmResources,
  getJvmResource,
  getJvmOverview,
  
  // 内存监控
  queryMemoryUsage,
  queryMemoryPools,
  
  // GC监控
  queryGcSnapshots,
  getLatestGcSnapshot,
  getGcTrend,
  
  // 线程监控
  queryThreadInfo,
  getThreadState,
  queryThreadStates,
  queryDeadlocks,
  
  // 应用监控数据
  queryAppMonitorData,
  getAppMonitorDataDetail,
  
  // 线程池监控
  queryThreadPools,
  
  // 类加载监控
  queryClassLoading
}

