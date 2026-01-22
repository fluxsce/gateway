/**
 * 预警日志管理 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { TrashOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import type { AlertLevel, AlertLog, SendStatus } from '../types'
import { ALERT_LEVEL_OPTIONS, SEND_STATUS_OPTIONS } from '../types'

export function useAlertLogModel() {
  const moduleId = 'hub0082:alert-log'

  const loading = ref(false)
  const logList = ref<AlertLog[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  // 初始化当天时间范围
  const initTodayTimeRange = (): [number, number] => {
    const today = new Date()
    const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0)
    const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59)
    return [startOfDay.getTime(), endOfDay.getTime()]
  }

  // 时间范围快捷选项
  const timeRangeShortcuts = {
    '今天': () => {
      const today = new Date()
      const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0)
      const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59)
      return [startOfDay.getTime(), endOfDay.getTime()] as [number, number]
    },
    '昨天': () => {
      const yesterday = new Date()
      yesterday.setDate(yesterday.getDate() - 1)
      const startOfDay = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate(), 0, 0, 0)
      const endOfDay = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate(), 23, 59, 59)
      return [startOfDay.getTime(), endOfDay.getTime()] as [number, number]
    },
    '最近1小时': () => {
      const now = Date.now()
      return [now - 3600000, now] as [number, number]
    },
    '最近6小时': () => {
      const now = Date.now()
      return [now - 21600000, now] as [number, number]
    },
    '最近24小时': () => {
      const now = Date.now()
      return [now - 86400000, now] as [number, number]
    },
    '最近7天': () => {
      const now = Date.now()
      return [now - 604800000, now] as [number, number]
    },
  }

  // 搜索表单
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'timeRange',
        label: '时间范围',
        type: 'datetimerange',
        placeholder: '请选择时间范围',
        span: 8,
        clearable: true,
        required: true,
        rules: [
          {
            validator: (_rule: any, value: any) => {
              if (!value || !Array.isArray(value) || value.length !== 2 || !value[0] || !value[1]) {
                return new Error('请选择时间范围')
              }
              return true
            },
            trigger: ['change', 'blur'],
          },
        ],
        props: {
          shortcuts: timeRangeShortcuts,
          style: { width: '100%' },
        },
        defaultValue: initTodayTimeRange(),
      },
      {
        field: 'alertLogId',
        label: '日志ID',
        type: 'input',
        placeholder: '请输入日志ID',
        span: 6,
        clearable: true,
      },
      {
        field: 'alertLevel',
        label: '告警级别',
        type: 'select',
        placeholder: '请选择告警级别',
        span: 6,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...ALERT_LEVEL_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      },
      {
        field: 'alertType',
        label: '告警类型',
        type: 'input',
        placeholder: '请输入告警类型',
        span: 6,
        clearable: true,
      },
      {
        field: 'alertTitle',
        label: '告警标题',
        type: 'input',
        placeholder: '请输入告警标题',
        span: 6,
        clearable: true,
      },
      {
        field: 'channelName',
        label: '渠道名称',
        type: 'input',
        placeholder: '请输入渠道名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'sendStatus',
        label: '发送状态',
        type: 'select',
        placeholder: '请选择发送状态',
        span: 6,
        clearable: true,
        options: [{ label: '全部', value: '' }, ...SEND_STATUS_OPTIONS.map(o => ({ label: o.label, value: o.value }))],
      },
    ],
    toolbarButtons: [
      { key: 'delete', label: '删除', icon: TrashOutline, type: 'error', tooltip: '批量删除选中的日志' },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const getAlertLevelLabel = (level?: AlertLevel | string | null) => {
    if (!level) return ''
    const option = ALERT_LEVEL_OPTIONS.find(opt => opt.value === level)
    return option?.label || String(level)
  }

  const getAlertLevelTagType = (level?: AlertLevel | string | null): 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info' => {
    if (!level) return 'default'
    const levelMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info'> = {
      INFO: 'info',
      WARN: 'warning',
      ERROR: 'error',
      CRITICAL: 'error',
    }
    return levelMap[level] || 'default'
  }

  const getSendStatusLabel = (status?: SendStatus | string | null) => {
    if (!status) return ''
    const option = SEND_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.label || String(status)
  }

  const getSendStatusTagType = (status?: SendStatus | string | null): 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info' => {
    if (!status) return 'default'
    const statusMap: Record<string, 'default' | 'success' | 'error' | 'warning' | 'primary' | 'info'> = {
      PENDING: 'default',
      SENDING: 'info',
      SUCCESS: 'success',
      FAILED: 'error',
    }
    return statusMap[status] || 'default'
  }

  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      { field: 'alertLogId', title: '日志ID', align: 'center', showOverflow: 'tooltip', width: 200 },
      { field: 'alertLevel', title: '告警级别', align: 'center', width: 100, slots: { default: 'alertLevel' } },
      { field: 'alertType', title: '告警类型', align: 'center', showOverflow: 'tooltip', width: 120 },
      { field: 'alertTitle', title: '告警标题', align: 'center', showOverflow: 'tooltip', width: 200 },
      { field: 'alertContent', title: '告警内容', align: 'center', showOverflow: 'tooltip', width: 300 },
      { field: 'channelName', title: '渠道名称', align: 'center', showOverflow: 'tooltip', width: 150 },
      { field: 'sendStatus', title: '发送状态', align: 'center', width: 100, slots: { default: 'sendStatus' } },
      { field: 'alertTimestamp', title: '告警时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.alertTimestamp) },
      { field: 'sendTime', title: '发送时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.sendTime) },
      { field: 'sendErrorMessage', title: '错误信息', align: 'center', showOverflow: 'tooltip', width: 200 },
      { field: 'addTime', title: '创建时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.addTime) },
      { field: 'addWho', title: '创建人', align: 'center', width: 120, showOverflow: 'tooltip' },
      { field: 'editTime', title: '修改时间', align: 'center', width: 160, formatter: ({ row }) => formatDate(row.editTime) },
      { field: 'editWho', title: '修改人', align: 'center', width: 120, showOverflow: 'tooltip' },
    ],
    showCheckbox: true,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      customMenus: [
        { code: 'view', name: '查看详情', prefixIcon: 'vxe-icon-eye-fill' },
        { code: 'delete', name: '删除', prefixIcon: 'vxe-icon-delete' },
      ],
    },
  }

  function setLogList(list: AlertLog[]) {
    logList.value = list
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
    logList,
    pageInfo,

    searchFormConfig,
    gridConfig,

    getAlertLevelLabel,
    getAlertLevelTagType,
    getSendStatusLabel,
    getSendStatusTagType,

    setLogList,
    setLoading,
    resetPagination,
    updatePagination,
  }
}

export type AlertLogModel = ReturnType<typeof useAlertLogModel>

