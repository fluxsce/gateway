/**
 * 限流配置业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import * as securityApi from '../../api/securityConfig'
import { useRateLimitConfigModel } from './model'
import type { RateLimitConfig } from './types'

/**
 * 限流配置服务 Hook（纯业务逻辑）
 * @param moduleId 模块ID（用于权限控制，必填）
 */
export function useRateLimitConfigService(moduleId: string) {
  const message = useMessage()

  // 初始化 Model（传递 moduleId）
  const model = useRateLimitConfigModel(moduleId)

  const { loading } = model

  // ============= 数据加载 =============

  /**
   * 获取限流配置详情
   * @param params 查询参数（gatewayInstanceId、routeConfigId）
   */
  const getConfigDetail = async (params: {
    gatewayInstanceId?: string
    routeConfigId?: string
  }): Promise<RateLimitConfig | null> => {
    try {
      const response: JsonDataObj = await securityApi.queryRateLimitConfigs(params)

      if (isApiSuccess(response)) {
        // queryRateLimitConfigs 返回单个配置或 null（没有配置时）
        return parseJsonData<RateLimitConfig | null>(response, null)
      } else {
        message.error(getApiMessage(response, '获取限流配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取限流配置详情失败')
      return null
    }
  }

  // ============= 增删改 =============

  /**
   * 添加限流配置
   */
  const addConfig = async (configData: Partial<RateLimitConfig> & {
    gatewayInstanceId?: string
    routeConfigId?: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.addRateLimitConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '新增限流配置成功'))
        return true
      } else {
        message.error(getApiMessage(response, '新增限流配置失败'))
        return false
      }
    } catch (error) {
      message.error('新增限流配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑限流配置
   */
  const editConfig = async (configData: Partial<RateLimitConfig> & {
    rateLimitConfigId: string
    tenantId: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.updateRateLimitConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '编辑限流配置成功'))
        return true
      } else {
        message.error(getApiMessage(response, '编辑限流配置失败'))
        return false
      }
    } catch (error) {
      message.error('编辑限流配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    getConfigDetail,

    // 配置操作
    addConfig,
    editConfig,
  }
}

