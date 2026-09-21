/**
 * 环境设置模块 API
 *
 * API路径: /gateway/hub0009
 * - POST /getEnvSettings - 读取归档策略、归档任务、Web 访问与全局环境变量
 * - POST /getSystemHealth - 探测平台库、ClickHouse、Mongo 与缓存
 * - POST /saveEnvSetting - 保存单个分组
 * - POST /saveEnvVar - 新增或更新一条全局环境变量
 * - POST /deleteEnvVar - 删除一条全局环境变量
 */

import { createApi } from '@/api/request'
import { moduleApiPrefix } from '@/api/requestPath'
import type { JsonDataObj } from '@/types/api'
import type { SaveEnvSettingPayload } from '../types'

const envSettingApi = createApi(moduleApiPrefix('hub0009'))

/** 读取当前租户环境设置。 */
export const getEnvSettings = async (): Promise<JsonDataObj> => {
  return envSettingApi.post('/getEnvSettings', {})
}

/** 探测本进程已装配依赖是否可达。 */
export const getSystemHealth = async (): Promise<JsonDataObj> => {
  return envSettingApi.post('/getSystemHealth', {})
}

/** 保存单个设置分组。 */
export const saveEnvSetting = async (payload: SaveEnvSettingPayload): Promise<JsonDataObj> => {
  return envSettingApi.post('/saveEnvSetting', payload)
}

export interface SaveEnvVarPayload {
  name: string
  originalName?: string
  value: string
  secret: boolean
  note: string
  currentVersion: number
}

export interface DeleteEnvVarPayload {
  name: string
  currentVersion: number
}

/** 新增或更新一条全局环境变量。 */
export const saveEnvVar = async (payload: SaveEnvVarPayload): Promise<JsonDataObj> => {
  return envSettingApi.post('/saveEnvVar', payload)
}

/** 删除一条全局环境变量。 */
export const deleteEnvVar = async (payload: DeleteEnvVarPayload): Promise<JsonDataObj> => {
  return envSettingApi.post('/deleteEnvVar', payload)
}
