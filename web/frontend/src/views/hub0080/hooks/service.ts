/**
 * 告警渠道配置列表服务层 Hook
 * 纯业务逻辑：数据获取、增删改查等操作
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import {
  createAlertConfig,
  getAlertConfig,
  queryAlertConfigs,
  reloadAlertChannel,
  setDefaultChannel,
  testAlertChannel,
  updateAlertConfig,
} from '../api'
import type { AlertConfig } from '../types'
import { useAlertConfigModel } from './model'

/**
 * 告警渠道配置列表服务层 Hook
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useAlertConfigService(
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()

  // 使用 model
  const model = useAlertConfigModel()
  
  // 解构 model 中的列表操作方法
  const {
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
    removeConfigsFromList,
  } = model

  /**
   * 加载配置列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadConfigList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const queryParams: any = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数
        ...createBackendPaginationParams(
          model.pageInfo.value?.pageIndex,
          model.pageInfo.value?.pageSize
        ),
      }

      const response = await queryAlertConfigs(queryParams)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const configs = parseJsonData<AlertConfig[]>(response, []) || []
        model.setConfigList(configs)

        // 解析分页信息
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          model.updatePagination(backendPageInfo)
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(response, '加载配置列表失败'))
        model.setConfigList([])
        model.resetPagination()
      }
    } catch (error: any) {
      console.error('加载配置列表失败:', error)
      message.error(error.message || '加载配置列表失败')
      model.setConfigList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 获取配置详情
   */
  const getConfigDetail = async (channelName: string): Promise<AlertConfig | null> => {
    try {
      const response = await getAlertConfig(channelName)
      if (isApiSuccess(response)) {
        return parseJsonData<AlertConfig>(response)
      } else {
        message.error(getApiMessage(response, '获取配置详情失败'))
        return null
      }
    } catch (error: any) {
      console.error('获取配置详情失败:', error)
      message.error(error.message || '获取配置详情失败')
      return null
    }
  }

  /**
   * 新增配置
   */
  const addConfig = async (configData: Partial<AlertConfig>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await createAlertConfig(configData)
      if (isApiSuccess(response)) {
        message.success('新增配置成功')
        
        // 如果返回了新增的配置数据，添加到列表
        const newConfig = parseJsonData<AlertConfig | null>(response, null)
        if (newConfig) {
          addConfigToList(newConfig)
        } else {
          // 否则重新加载列表
          await loadConfigList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '新增配置失败'))
        return false
      }
    } catch (error: any) {
      console.error('新增配置失败:', error)
      message.error(error.message || '新增配置失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 编辑配置
   */
  const editConfig = async (channelName: string, configData: Partial<AlertConfig>): Promise<boolean> => {
    try {
      model.setLoading(true)

      const submitData = {
        ...configData,
        channelName,
      }

      const response = await updateAlertConfig(submitData)
      if (isApiSuccess(response)) {
        message.success('编辑配置成功')
        
        // 更新列表中的配置数据：必须以返回的 response.bizData 为准
        const updatedConfig = parseJsonData<AlertConfig | null>(response, null)
        if (updatedConfig) {
          updateConfigInList(updatedConfig.channelName, updatedConfig.tenantId, updatedConfig)
        } else {
          // 如果后端没有返回数据，重新加载列表以确保数据一致性
          await loadConfigList()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '编辑配置失败'))
        return false
      }
    } catch (error: any) {
      console.error('编辑配置失败:', error)
      message.error(error.message || '编辑配置失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 设置默认渠道
   */
  const setDefault = async (channelName: string): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await setDefaultChannel(channelName)
      if (isApiSuccess(response)) {
        message.success('设置默认渠道成功')
        
        // 重新加载列表以确保数据一致性
        await loadConfigList()
        
        return true
      } else {
        message.error(getApiMessage(response, '设置默认渠道失败'))
        return false
      }
    } catch (error: any) {
      console.error('设置默认渠道失败:', error)
      message.error(error.message || '设置默认渠道失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 切换配置状态
   */
  const toggleConfigStatus = async (config: AlertConfig): Promise<boolean> => {
    const newStatus = config.activeFlag === 'Y' ? 'N' : 'Y'
    return await editConfig(config.channelName, {
      ...config,
      activeFlag: newStatus,
    })
  }

  /**
   * 测试告警渠道
   */
  const testChannel = async (channelName: string): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await testAlertChannel(channelName)
      if (isApiSuccess(response)) {
        const result = parseJsonData<any>(response)
        const successMsg = result?.message || '告警渠道测试成功'
        message.success(successMsg)
        return true
      } else {
        message.error(getApiMessage(response, '告警渠道测试失败'))
        return false
      }
    } catch (error: any) {
      console.error('测试告警渠道失败:', error)
      message.error(error.message || '测试告警渠道失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  /**
   * 重载告警渠道配置（使配置立即生效）
   */
  const reloadChannel = async (channelName: string): Promise<boolean> => {
    try {
      model.setLoading(true)

      const response = await reloadAlertChannel(channelName)
      if (isApiSuccess(response)) {
        const result = parseJsonData<any>(response)
        const msg = result?.reloaded ? '配置重载成功' : (result?.message || '配置重载成功')
        message.success(msg)
        return true
      } else {
        message.error(getApiMessage(response, '配置重载失败'))
        return false
      }
    } catch (error: any) {
      console.error('配置重载失败:', error)
      message.error(error.message || '配置重载失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadConfigList,
    getConfigDetail,
    addConfig,
    editConfig,
    setDefault,
    toggleConfigStatus,
    testChannel,
    reloadChannel,
  }
}

