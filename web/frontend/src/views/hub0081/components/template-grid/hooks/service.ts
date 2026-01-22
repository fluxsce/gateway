/**
 * 模板列表查询服务层 Hook（仅查询功能）
 * 用于模板选择器组件
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { queryAlertTemplates } from '../../../api'
import type { AlertTemplate, AlertTemplateQueryParams } from '../../../types'
import { useAlertTemplateListModel } from './model'

export function useAlertTemplateListService(searchFormRef?: Ref<any> | any, channelType?: string) {
  const message = useMessage()
  const model = useAlertTemplateListModel(channelType)

  const loadTemplateList = async (searchParams?: Record<string, any>) => {
    try {
      model.setLoading(true)

      let finalSearchParams = searchParams
      if (!finalSearchParams && searchFormRef?.value?.getFormData) {
        finalSearchParams = searchFormRef.value.getFormData() || {}
      }

      const effectiveSearchParams = finalSearchParams
        ? Object.fromEntries(Object.entries(finalSearchParams).filter(([, v]) => v !== '' && v !== null && v !== undefined))
        : {}

      const queryParams: AlertTemplateQueryParams = {
        ...effectiveSearchParams,
        ...createBackendPaginationParams(model.pageInfo.value?.pageIndex, model.pageInfo.value?.pageSize),
      }

      const resp = await queryAlertTemplates(queryParams)
      if (isApiSuccess(resp)) {
        const rows = parseJsonData<AlertTemplate[]>(resp, []) || []
        model.setTemplateList(rows)

        const backendPageInfo = parsePageInfo(resp)
        if (backendPageInfo && Object.keys(backendPageInfo).length > 0) {
          model.updatePagination(backendPageInfo)
        } else {
          model.resetPagination()
        }
      } else {
        message.error(getApiMessage(resp, '加载模板列表失败'))
        model.setTemplateList([])
        model.resetPagination()
      }
    } catch (e: any) {
      console.error('加载模板列表失败:', e)
      message.error(e.message || '加载模板列表失败')
      model.setTemplateList([])
      model.resetPagination()
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadTemplateList,
  }
}

