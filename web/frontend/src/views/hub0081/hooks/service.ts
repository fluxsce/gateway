/**
 * 预警模板管理服务层 Hook
 */

import { createBackendPaginationParams } from '@/components/gpage'
import { getApiMessage, isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { createAlertTemplate, deleteAlertTemplate, getAlertTemplate, queryAlertTemplates, updateAlertTemplate } from '../api'
import type { AlertTemplate } from '../types'
import { useAlertTemplateModel } from './model'

export function useAlertTemplateService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()
  const model = useAlertTemplateModel()

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

      const queryParams: any = {
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

  const getTemplateDetail = async (templateName: string): Promise<AlertTemplate | null> => {
    try {
      const resp = await getAlertTemplate(templateName)
      if (isApiSuccess(resp)) {
        return parseJsonData<AlertTemplate>(resp)
      }
      message.error(getApiMessage(resp, '获取模板详情失败'))
      return null
    } catch (e: any) {
      console.error('获取模板详情失败:', e)
      message.error(e.message || '获取模板详情失败')
      return null
    }
  }

  const addTemplate = async (data: Partial<AlertTemplate>): Promise<boolean> => {
    try {
      model.setLoading(true)
      const resp = await createAlertTemplate(data)
      if (isApiSuccess(resp)) {
        message.success('新增模板成功')
        await loadTemplateList()
        return true
      }
      message.error(getApiMessage(resp, '新增模板失败'))
      return false
    } catch (e: any) {
      console.error('新增模板失败:', e)
      message.error(e.message || '新增模板失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  const editTemplate = async (templateName: string, data: Partial<AlertTemplate>): Promise<boolean> => {
    try {
      model.setLoading(true)
      const resp = await updateAlertTemplate({ ...data, templateName })
      if (isApiSuccess(resp)) {
        message.success('更新模板成功')
        await loadTemplateList()
        return true
      }
      message.error(getApiMessage(resp, '更新模板失败'))
      return false
    } catch (e: any) {
      console.error('更新模板失败:', e)
      message.error(e.message || '更新模板失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  const removeTemplate = async (templateName: string): Promise<boolean> => {
    try {
      model.setLoading(true)
      const resp = await deleteAlertTemplate(templateName)
      if (isApiSuccess(resp)) {
        message.success('删除模板成功')
        await loadTemplateList()
        return true
      }
      message.error(getApiMessage(resp, '删除模板失败'))
      return false
    } catch (e: any) {
      console.error('删除模板失败:', e)
      message.error(e.message || '删除模板失败')
      return false
    } finally {
      model.setLoading(false)
    }
  }

  return {
    model,
    loadTemplateList,
    getTemplateDetail,
    addTemplate,
    editTemplate,
    removeTemplate,
  }
}


