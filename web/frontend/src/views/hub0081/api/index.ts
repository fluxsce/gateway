/**
 * 预警(告警)模板管理模块API
 *
 * API路径: /gateway/hub0081
 *
 * - POST /queryAlertTemplates - 查询模板列表
 * - POST /getAlertTemplate - 获取模板详情
 * - POST /createAlertTemplate - 创建模板
 * - POST /updateAlertTemplate - 更新模板
 * - POST /deleteAlertTemplate - 删除模板
 */

import { createApi } from '@/api/request'
import type { JsonDataObj } from '@/types/api'
import type { AlertTemplate, AlertTemplateQueryParams } from '../types'

const alertTemplateApi = createApi('/gateway/hub0081')

export const queryAlertTemplates = async (params: AlertTemplateQueryParams): Promise<JsonDataObj> => {
  return alertTemplateApi.post('/queryAlertTemplates', params)
}

export const getAlertTemplate = async (templateName: string): Promise<JsonDataObj> => {
  return alertTemplateApi.post('/getAlertTemplate', { templateName })
}

export const createAlertTemplate = async (data: Partial<AlertTemplate>): Promise<JsonDataObj> => {
  return alertTemplateApi.post('/createAlertTemplate', data)
}

export const updateAlertTemplate = async (data: Partial<AlertTemplate> & { templateName: string }): Promise<JsonDataObj> => {
  return alertTemplateApi.post('/updateAlertTemplate', data)
}

export const deleteAlertTemplate = async (templateName: string): Promise<JsonDataObj> => {
  return alertTemplateApi.post('/deleteAlertTemplate', { templateName })
}


