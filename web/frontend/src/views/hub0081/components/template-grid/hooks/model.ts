/**
 * 模板列表查询 Model（仅查询功能）
 * 用于模板选择器组件
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { ref } from 'vue'
import type { AlertTemplate, ChannelType, DisplayFormat } from '../../../types'
import { ACTIVE_FLAG_OPTIONS, CHANNEL_TYPE_OPTIONS, DISPLAY_FORMAT_OPTIONS } from '../../../types'

export function useAlertTemplateListModel(channelType?: string) {
  const moduleId = `hub0081:alert-template-list${channelType ? `:${channelType}` : ''}`

  const loading = ref(false)
  const templateList = ref<AlertTemplate[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const getChannelTypeLabel = (channelType?: ChannelType | string | null) => {
    if (!channelType) return ''
    const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === channelType)
    return option?.label || String(channelType)
  }

  const getDisplayFormatLabel = (format?: DisplayFormat | string | null) => {
    if (!format) return ''
    const option = DISPLAY_FORMAT_OPTIONS.find(opt => opt.value === format)
    return option?.label || String(format)
  }

  // 搜索表单配置（简化版，只包含必要的搜索字段）
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'templateName',
        label: '模板名称',
        type: 'input',
        placeholder: '请输入模板名称',
        span: 8,
        clearable: true,
      },
      {
        field: 'channelType',
        label: '渠道类型',
        type: 'select',
        placeholder: '请选择渠道类型',
        span: 8,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...CHANNEL_TYPE_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
        defaultValue: channelType || '',
      },
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'select',
        placeholder: '请选择启用状态',
        span: 8,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...ACTIVE_FLAG_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
        defaultValue: 'Y', // 默认只显示启用的模板
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      { field: 'templateName', title: '模板名称', align: 'center', showOverflow: 'tooltip', width: 180 },
      { field: 'channelType', title: '渠道类型', align: 'center', width: 120, slots: { default: 'channelType' } },
      { field: 'displayFormat', title: '显示格式', align: 'center', width: 120, slots: { default: 'displayFormat' } },
      { field: 'activeFlag', title: '启用状态', align: 'center', width: 100, slots: { default: 'activeFlag' } },
      { field: 'templateDesc', title: '模板描述', align: 'center', showOverflow: 'tooltip', width: 240 },
    ],
    showCheckbox: false,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
  }

  function setTemplateList(list: AlertTemplate[]) {
    templateList.value = list
  }
  function setLoading(value: boolean) {
    loading.value = value
  }
  function resetPagination() {
    pageInfo.value = undefined
  }
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  return {
    moduleId,
    loading,
    templateList,
    pageInfo,

    searchFormConfig,
    gridConfig,

    getChannelTypeLabel,
    getDisplayFormatLabel,

    setTemplateList,
    setLoading,
    resetPagination,
    updatePagination,
  }
}

export type AlertTemplateListModel = ReturnType<typeof useAlertTemplateListModel>

