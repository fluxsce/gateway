/**
 * Hub0043 配置管理业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as configApi from '../../../api'
import type { Config } from '../../../types'
import { useConfigModel } from './model'

/**
 * 配置服务 Hook（纯业务逻辑）
 * @param searchFormRef 搜索表单引用
 */
export function useConfigService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model
  const model = useConfigModel()

  const {
    loading,
    configList,
    pageInfo,
    setConfigList,
    updatePagination,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
    removeConfigsFromList,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载配置列表
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const loadConfigs = async (searchParams?: Record<string, any>) => {
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

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const paginationParams = createBackendPaginationParams(
        pageInfo.value?.pageIndex,
        pageInfo.value?.pageSize
      )
      const params: any = {
        ...effectiveSearchParams,
      }
      if (paginationParams.pageIndex !== undefined) {
        params.page = paginationParams.pageIndex
      }
      if (paginationParams.pageSize !== undefined) {
        params.pageSize = paginationParams.pageSize
      }

      // 调用 API
      const response: JsonDataObj = await configApi.queryConfigs(params as any)

      if (response.oK) {
        if (response.bizData) {
          const bizData = JSON.parse(response.bizData)
          const configs = Array.isArray(bizData) ? bizData : []
          setConfigList(configs)
        }

        if (response.pageQueryData) {
          const backendPageInfo = JSON.parse(response.pageQueryData)
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(response.errMsg || '查询配置列表失败')
      }
    } catch (error) {
      message.error('加载配置列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  const handleSearch = async (searchParams?: Record<string, any>) => {
    await loadConfigs(searchParams)
  }

  const handleReset = async () => {
    model.resetPagination()
    await loadConfigs()
  }

  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    await loadConfigs()
  }

  const handleRefresh = async () => {
    await loadConfigs()
  }

  // ============= 增删改 =============

  const addConfig = async (configData: Partial<Config> & {
    namespaceId: string
    configDataId: string
    configContent: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await configApi.addConfig(configData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '配置创建成功')
        message.success(successMsg)
        
        if (response.bizData) {
          const newConfig = JSON.parse(response.bizData)
          addConfigToList(newConfig as any)
        } else {
          await loadConfigs()
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '新增配置失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('新增配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const editConfig = async (configData: Partial<Config> & {
    namespaceId: string
    groupName: string
    configDataId: string
    configContent: string
  }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await configApi.editConfig(configData)

      if (isApiSuccess(response)) {
        const successMsg = getApiMessage(response, '配置更新成功')
        message.success(successMsg)
        
        if (response.bizData) {
          const updatedConfig = JSON.parse(response.bizData)
          updateConfigInList(
            updatedConfig.namespaceId,
            updatedConfig.groupName,
            updatedConfig.configDataId,
            updatedConfig
          )
        }
        
        return true
      } else {
        const errorMsg = getApiMessage(response, '编辑配置失败')
        message.error(errorMsg)
        return false
      }
    } catch (error) {
      message.error('编辑配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const deleteConfig = async (config: Config, skipConfirm = false): Promise<boolean> => {
    // 如果不需要确认，直接执行删除；否则显示确认对话框
    if (!skipConfirm) {
      const confirmed = await gDialog.warning({
        title: '确认删除',
        subtitle: '此操作不可恢复，请谨慎操作',
        content: `确定要删除配置吗？\n\n配置ID: ${config.configDataId}\n命名空间: ${config.namespaceId}\n分组: ${config.groupName || 'DEFAULT_GROUP'}`,
        icon: WarningOutline,
        headerStyle: 'gradient',
        positiveText: '确定删除',
        negativeText: '取消',
        width: 500
      })

      if (!confirmed) {
        return false
      }
    }

    loading.value = true
    try {
      const response: JsonDataObj = await configApi.deleteConfig(
        config.namespaceId,
        config.groupName || 'DEFAULT_GROUP',
        config.configDataId
      )

      if (isApiSuccess(response)) {
        // 只有在非批量删除时才显示成功消息（批量删除会在 handleBatchDelete 中统一显示）
        if (!skipConfirm) {
          const successMsg = getApiMessage(response, '配置删除成功')
          message.success(successMsg)
        }
        removeConfigFromList(
          config.namespaceId,
          config.groupName || 'DEFAULT_GROUP',
          config.configDataId
        )
        return true
      } else {
        // 只有在非批量删除时才显示错误消息（批量删除会在 handleBatchDelete 中统一显示）
        if (!skipConfirm) {
          const errorMsg = getApiMessage(response, '删除配置失败')
          message.error(errorMsg)
        }
        return false
      }
    } catch (error) {
      message.error('删除配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  const getConfigDetail = async (
    namespaceId: string,
    groupName: string,
    configDataId: string
  ): Promise<Config | null> => {
    loading.value = true
    try {
      const response: JsonDataObj = await configApi.getConfig(namespaceId, groupName, configDataId)

      if (isApiSuccess(response) && response.bizData) {
        return JSON.parse(response.bizData)
      } else {
        message.error(getApiMessage(response, '获取配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取配置详情失败')
      return null
    } finally {
      loading.value = false
    }
  }

  return {
    // Model 实例（包含所有状态和配置）
    model,

    // 数据加载
    loadConfigs,

    // 搜索和分页
    handleSearch,
    handleReset,
    handlePageChange,
    handleRefresh,

    // 配置操作
    addConfig,
    editConfig,
    deleteConfig,
    getConfigDetail,
  }
}

export type ConfigService = ReturnType<typeof useConfigService>

