/**
 * 预警(告警)配置管理模块API
 * 提供告警渠道配置的增删改查等功能
 * 
 * API路径: /gateway/hub0080
 * 
 * 告警渠道配置管理API：
 * - POST /queryAlertConfigs - 查询告警渠道配置列表
 * - POST /getAlertConfig - 获取告警渠道配置详情
 * - POST /createAlertConfig - 创建告警渠道配置
 * - POST /updateAlertConfig - 更新告警渠道配置
 * - POST /setDefaultChannel - 设置默认告警渠道
 * - POST /reloadAlertChannel - 重载告警渠道配置
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type {
    AlertConfig,
    AlertConfigQueryParams,
} from '../types'

// 创建API实例
const alertConfigApi = createApi('/gateway/hub0080')

// ==================== 告警渠道配置管理API ====================

/**
 * 查询告警渠道配置列表
 * @param params 查询参数
 * @returns 告警渠道配置列表
 */
export const queryAlertConfigs = async (params: AlertConfigQueryParams): Promise<JsonDataObj> => {
  return alertConfigApi.post('/queryAlertConfigs', params)
}

/**
 * 获取告警渠道配置详情
 * @param channelName 渠道名称
 * @returns 告警渠道配置详情
 */
export const getAlertConfig = async (channelName: string): Promise<JsonDataObj> => {
  return alertConfigApi.post('/getAlertConfig', { channelName })
}

/**
 * 创建告警渠道配置
 * @param data 告警渠道配置数据
 * @returns 创建结果
 */
export const createAlertConfig = async (data: Partial<AlertConfig>): Promise<JsonDataObj> => {
  return alertConfigApi.post('/createAlertConfig', data)
}

/**
 * 更新告警渠道配置
 * @param data 告警渠道配置数据（包含channelName）
 * @returns 更新结果
 */
export const updateAlertConfig = async (data: Partial<AlertConfig> & { channelName: string }): Promise<JsonDataObj> => {
  return alertConfigApi.post('/updateAlertConfig', data)
}

/**
 * 设置默认告警渠道
 * @param channelName 渠道名称
 * @returns 设置结果
 */
export const setDefaultChannel = async (channelName: string): Promise<JsonDataObj> => {
  return alertConfigApi.post('/setDefaultChannel', { channelName })
}

/**
 * 测试告警渠道
 * @param channelName 渠道名称
 * @param title 测试消息主题（可选）
 * @param content 测试消息内容（可选）
 * @returns 测试结果
 */
export const testAlertChannel = async (
  channelName: string,
  title?: string,
  content?: string
): Promise<JsonDataObj> => {
  const params: Record<string, any> = { channelName }
  if (title) {
    params.title = title
  }
  if (content) {
    params.content = content
  }
  return alertConfigApi.post('/testAlertChannel', params)
}

/**
 * 重载告警渠道配置
 * @param channelName 渠道名称
 * @returns 重载结果
 */
export const reloadAlertChannel = async (channelName: string): Promise<JsonDataObj> => {
  return alertConfigApi.post('/reloadAlertChannel', { channelName })
}

