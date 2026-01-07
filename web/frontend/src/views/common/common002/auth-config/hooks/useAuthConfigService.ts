/**
 * 认证配置业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import * as securityApi from '../../api/securityConfig'
import { useAuthConfigModel } from './model'
import type { AuthConfig } from './types'

/**
 * 认证配置服务 Hook（纯业务逻辑）
 * @param moduleId 模块ID（用于权限控制，必填）
 */
export function useAuthConfigService(moduleId: string) {
  const message = useMessage()

  // 初始化 Model（传递 moduleId）
  const model = useAuthConfigModel(moduleId)

  const { loading } = model

  // ============= 数据加载 =============

  /**
   * 获取认证配置详情
   * @param params 查询参数（gatewayInstanceId、routeConfigId）
   */
  const getConfigDetail = async (params: {
    gatewayInstanceId?: string
    routeConfigId?: string
  }): Promise<AuthConfig | null> => {
    try {
      const response: JsonDataObj = await securityApi.queryAuthConfigs(params)

      if (isApiSuccess(response)) {
        // queryAuthConfigs 返回单个配置或 null（没有配置时）
        // parseJsonData 会直接解析返回对象或 null
        return parseJsonData<AuthConfig | null>(response, null)
      } else {
        message.error(getApiMessage(response, '获取认证配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取认证配置详情失败')
      return null
    }
  }

  // ============= 增删改 =============

  /**
   * 添加认证配置
   */
  const addConfig = async (configData: Partial<AuthConfig> & {
    gatewayInstanceId?: string
    routeConfigId?: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.addAuthConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '新增认证配置成功'))
        return true
      } else {
        message.error(getApiMessage(response, '新增认证配置失败'))
        return false
      }
    } catch (error) {
      message.error('新增认证配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑认证配置
   */
  const editConfig = async (configData: Partial<AuthConfig> & {
    authConfigId: string
    tenantId: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.updateAuthConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '编辑认证配置成功'))
        return true
      } else {
        message.error(getApiMessage(response, '编辑认证配置失败'))
        return false
      }
    } catch (error) {
      message.error('编辑认证配置失败')
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

