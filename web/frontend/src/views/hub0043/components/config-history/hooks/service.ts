/**
 * Hub0043 配置历史管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { queryConfigHistory, getHistoryById, rollbackConfig } from '../../../api'
import type { ConfigHistory, ConfigHistoryRequest, RollbackRequest } from '../../../types'
import { useConfigHistoryModel } from './model'

/**
 * 配置历史服务 Hook（纯业务逻辑）
 * @param searchFormRef 搜索表单引用
 */
export function useConfigHistoryService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useConfigHistoryModel()

  const {
    loading,
    historyList,
    setHistoryList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载配置历史
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadHistory = async (searchParams?: Record<string, any>) => {
    loading.value = true
    try {
      // 如果没有传入查询参数，从搜索表单获取
      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      // 验证必填字段
      if (!finalSearchParams?.namespaceId) {
        message.warning('请先选择命名空间')
        loading.value = false
        return
      }
      if (!finalSearchParams?.groupName) {
        message.warning('请输入分组名称')
        loading.value = false
        return
      }
      if (!finalSearchParams?.configDataId) {
        message.warning('请输入配置ID')
        loading.value = false
        return
      }

      const params: ConfigHistoryRequest = {
        namespaceId: finalSearchParams.namespaceId,
        groupName: finalSearchParams.groupName || 'DEFAULT_GROUP',
        configDataId: finalSearchParams.configDataId,
        limit: finalSearchParams.limit || 50,
      }

      const response: JsonDataObj = await queryConfigHistory(params)

      if (isApiSuccess(response) && response.bizData) {
        const data = JSON.parse(response.bizData)
        setHistoryList(Array.isArray(data) ? data : [])
      } else {
        message.error(getApiMessage(response, '获取配置历史失败'))
      }
    } catch (error) {
      message.error('获取配置历史失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取历史详情
   */
  const getHistoryDetail = async (configHistoryId: string): Promise<ConfigHistory | null> => {
    loading.value = true
    try {
      const response: JsonDataObj = await getHistoryById(configHistoryId)

      if (isApiSuccess(response) && response.bizData) {
        return JSON.parse(response.bizData)
      } else {
        message.error(getApiMessage(response, '获取历史详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取历史详情失败')
      return null
    } finally {
      loading.value = false
    }
  }

  /**
   * 回滚配置
   */
  const rollback = async (rollbackData: RollbackRequest): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await rollbackConfig(rollbackData)

      if (isApiSuccess(response)) {
        message.success('配置回滚成功')
        return true
      } else {
        message.error(getApiMessage(response, '配置回滚失败'))
        return false
      }
    } catch (error) {
      message.error('配置回滚失败')
      return false
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadHistory,
    getHistoryDetail,
    rollback,
  }
}

export type ConfigHistoryService = ReturnType<typeof useConfigHistoryService>

