/**
 * 模板列表查询 Model（仅查询功能）
 * 用于模板选择器组件
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag } from '@/ui'
import { h, ref } from 'vue'
import type { AlertTemplate, ChannelType, DisplayFormat } from '../../../types'
import { ACTIVE_FLAG_OPTIONS, CHANNEL_TYPE_OPTIONS, DISPLAY_FORMAT_OPTIONS } from '../../../types'

/**
 * 模板选择表格配置（对齐 RsGrid Props 子集）。
 */
export interface AlertTemplateListGridConfig {
  columns: RsGridColumn<AlertTemplate>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

export function useAlertTemplateListModel(channelType?: string) {
  const moduleId = 'hub0081'

  const loading = ref(false)
  const templateList = ref<AlertTemplate[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const getChannelTypeLabel = (type?: ChannelType | string | null) => {
    if (!type) return ''
    const option = CHANNEL_TYPE_OPTIONS.find(opt => opt.value === type)
    return option?.label || String(type)
  }

  const getDisplayFormatLabel = (format?: DisplayFormat | string | null) => {
    if (!format) return ''
    const option = DISPLAY_FORMAT_OPTIONS.find(opt => opt.value === format)
    return option?.label || String(format)
  }

  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
        defaultValue: 'Y',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const gridConfig: AlertTemplateListGridConfig = {
    columns: [
      { key: 'templateName', title: '模板名称', align: 'center', ellipsis: true, width: 180 },
      {
        key: 'channelType',
        title: '渠道类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: 'info', size: 'sm' },
            () => getChannelTypeLabel(row.channelType) || '通用',
          ),
      },
      {
        key: 'displayFormat',
        title: '显示格式',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            { variant: row.displayFormat === 'table' ? 'warning' : 'default', size: 'sm' },
            () => getDisplayFormatLabel(row.displayFormat),
          ),
      },
      {
        key: 'activeFlag',
        title: '启用状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      { key: 'templateDesc', title: '模板描述', align: 'center', ellipsis: true, width: 240 },
    ],
    selectable: false,
    rowKey: 'templateName',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: false,
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
