/**
 * 环境设置模块 API
 *
 * API路径: /gateway/hub0009
 * - POST /getEnvSettings - 读取归档策略、归档任务与 Web 超时
 * - POST /saveEnvSetting - 保存单个分组
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

/** 保存单个设置分组。 */
export const saveEnvSetting = async (payload: SaveEnvSettingPayload): Promise<JsonDataObj> => {
  return envSettingApi.post('/saveEnvSetting', payload)
}
