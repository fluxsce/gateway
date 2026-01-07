/**
 * API访问控制配置列表业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { useGDialog } from '@/components/gdialog'
import { createBackendPaginationParams } from '@/components/gpage'
import type { JsonDataObj } from '@/types/api'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { WarningOutline } from '@vicons/ionicons5'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import * as securityApi from '../../api/securityConfig'
import { useApiAccessConfigModel } from './model'
import type { ApiAccessConfig } from './types'

/**
 * API访问控制配置服务 Hook（纯业务逻辑）
 * @param moduleId 模块ID（用于权限控制，必填）
 * @param securityConfigId 安全配置ID（可选，用于查询时过滤）
 */
export function useApiAccessConfigService(
  moduleId: string,
  securityConfigId?: Ref<string | undefined>,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 初始化 Model（传递 moduleId）
  const model = useApiAccessConfigModel(moduleId)

  const {
    loading,
    configList,
    pageInfo,
    setConfigList,
    updatePagination,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
    resetPagination,
  } = model

  // ============= 数据加载 =============

  /**
   * 加载API配置列表
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

      // 过滤掉空字符串、null 和 undefined 的查询条件
      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(
            Object.entries(finalSearchParams).filter(
              ([, value]) => value !== '' && value !== null && value !== undefined
            )
          )
        : {}

      // 构建请求参数：合并查询条件和分页参数
      const params: Record<string, any> = {
        // 查询条件
        ...effectiveSearchParams,
        // 分页参数（函数内部会自动使用配置常量作为默认值）
        ...createBackendPaginationParams(
          pageInfo.value?.pageIndex,
          pageInfo.value?.pageSize
        ),
      }

      // 必须携带 securityConfigId（避免关联错误）
      if (securityConfigId?.value) {
        params.securityConfigId = securityConfigId.value
      } else if (!effectiveSearchParams.securityConfigId) {
        // 如果没有传入 securityConfigId，且查询条件中也没有，则提示错误
        message.error('securityConfigId不能为空')
        return
      }

      // 调用 API（POST 请求，参数通过 body 传递）
      const response: JsonDataObj = await securityApi.queryApiAccessConfigs(params)

      if (isApiSuccess(response)) {
        // 解析业务数据
        const configs = parseJsonData<ApiAccessConfig[]>(response, [])
        setConfigList(configs)

        // 解析分页信息 - 直接使用后端返回的 PageInfoObj
        const backendPageInfo = parsePageInfo(response)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          updatePagination(backendPageInfo)
        }
      } else {
        message.error(getApiMessage(response, '查询API访问控制配置列表失败'))
      }
    } catch (error) {
      message.error('加载API访问控制配置列表失败')
    } finally {
      loading.value = false
    }
  }

  // ============= 搜索和分页 =============

  /**
   * 搜索API配置
   * @param searchParams 查询条件（可选，如果不传则从搜索表单获取）
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    resetPagination()
    // loadConfigs 会自动从 searchFormRef 获取查询条件
    await loadConfigs(searchParams)
  }

  /**
   * 重置搜索
   */
  const handleReset = async () => {
    resetPagination()
    await loadConfigs()
  }

  /**
   * 分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    updatePagination({ pageIndex: currentPage, pageSize })
    // loadConfigs 会自动从 searchFormRef 获取查询条件
    await loadConfigs()
  }

  /**
   * 刷新列表
   */
  const handleRefresh = async () => {
    await loadConfigs()
  }

  // ============= 增删改 =============

  /**
   * 添加API配置
   */
  const addConfig = async (configData: Partial<ApiAccessConfig> & { securityConfigId: string }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.addApiAccessConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '新增API访问控制配置成功'))
        
        // 如果返回了新增的配置数据，添加到列表
        const newConfig = parseJsonData<ApiAccessConfig | null>(response, null)
        if (newConfig) {
          addConfigToList(newConfig)
        } else {
          // 否则重新加载列表
          await loadConfigs()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '新增API访问控制配置失败'))
        return false
      }
    } catch (error) {
      message.error('新增API访问控制配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 编辑API配置
   */
  const editConfig = async (configData: Partial<ApiAccessConfig> & { securityConfigId: string; apiAccessConfigId: string; tenantId: string }): Promise<boolean> => {
    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.updateApiAccessConfig(configData)

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '编辑API访问控制配置成功'))
        
        // 更新列表中的配置数据：必须以返回的 response.bizData 为准
        const updatedConfig = parseJsonData<ApiAccessConfig | null>(response, null)
        if (updatedConfig) {
          updateConfigInList(updatedConfig.apiAccessConfigId, updatedConfig.tenantId, updatedConfig)
        } else {
          // 如果后端没有返回数据，重新加载列表以确保数据一致性
          await loadConfigs()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '编辑API访问控制配置失败'))
        return false
      }
    } catch (error) {
      message.error('编辑API访问控制配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 删除API配置
   */
  const deleteConfig = async (config: ApiAccessConfig): Promise<boolean> => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      subtitle: '此操作不可恢复，请谨慎操作',
      content: `确定要删除API访问控制配置 "${config.configName}" 吗？`,
      icon: WarningOutline,
      headerStyle: 'gradient',
      positiveText: '确定删除',
      negativeText: '取消',
      width: 500
    })

    if (!confirmed) {
      return false
    }

    loading.value = true
    try {
      const response: JsonDataObj = await securityApi.deleteApiAccessConfig({
        apiAccessConfigId: config.apiAccessConfigId,
        tenantId: config.tenantId
      })

      if (isApiSuccess(response)) {
        message.success(getApiMessage(response, '删除API访问控制配置成功'))
        removeConfigFromList(config.apiAccessConfigId, config.tenantId)
        
        // 如果当前页没有数据了，且不是第一页，则跳转到上一页
        if (configList.value.length === 0 && pageInfo.value && pageInfo.value.pageIndex > 1) {
          updatePagination({ pageIndex: pageInfo.value.pageIndex - 1 })
          await loadConfigs()
        }
        
        return true
      } else {
        message.error(getApiMessage(response, '删除API访问控制配置失败'))
        return false
      }
    } catch (error) {
      message.error('删除API访问控制配置失败')
      return false
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取配置详情（使用主键 apiAccessConfigId）
   */
  const getConfigDetail = async (apiAccessConfigId: string): Promise<ApiAccessConfig | null> => {
    try {
      const response: JsonDataObj = await securityApi.getApiAccessConfig({
        apiAccessConfigId
      })
      
      if (isApiSuccess(response)) {
        return parseJsonData<ApiAccessConfig>(response, {} as ApiAccessConfig)
      } else {
        message.error(getApiMessage(response, '获取API访问控制配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取API访问控制配置详情失败')
      return null
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

