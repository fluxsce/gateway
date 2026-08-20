/**
 * 预警日志管理 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, reactive, ref, watch } from 'vue'
import type { AlertLevel, AlertLog, SendStatus } from '../types'

/**
 * 预警日志表格配置（对齐 RsGrid Props 子集）。
 */
export interface AlertLogGridConfig {
  columns: RsGridColumn<AlertLog>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 预警日志管理 Model
 */
export function useAlertLogModel() {
  const { t, locale } = useModuleI18n('hub0082')
  const moduleId = 'hub0082'

  const loading = ref(false)
  const logList = ref<AlertLog[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const gridConfig = reactive<AlertLogGridConfig>({
    columns: [],
    selectable: true,
    rowKey: 'alertLogId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [],
    },
  })

  /** 初始化当天时间范围（RsDatePicker valueFormat=string 的 range 形态） */
  const initTodayTimeRange = (): { start: string; end: string } => {
    const today = new Date()
    const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0)
    const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59)
    return {
      start: formatDate(startOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
      end: formatDate(endOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
    }
  }

  /**
   * 获取告警级别显示标签。
   * @param level - 告警级别
   * @returns 对应文案，未知值原样返回
   */
  const getAlertLevelLabel = (level?: AlertLevel | string | null) => {
    if (!level) return ''
    const levelMap: Record<string, string> = {
      INFO: t('level.info'),
      WARN: t('level.warn'),
      ERROR: t('level.error'),
      CRITICAL: t('level.critical'),
    }
    return levelMap[level] || String(level)
  }

  /**
   * 将告警级别映射为 RsTag variant。
   * @param level - 告警级别
   */
  const getAlertLevelTagType = (level?: AlertLevel | string | null): RsTagVariant => {
    if (!level) return 'default'
    const levelMap: Record<string, RsTagVariant> = {
      INFO: 'info',
      WARN: 'warning',
      ERROR: 'danger',
      CRITICAL: 'danger',
    }
    return levelMap[level] || 'default'
  }

  /**
   * 获取发送状态显示标签。
   * @param status - 发送状态
   */
  const getSendStatusLabel = (status?: SendStatus | string | null) => {
    if (!status) return ''
    const statusMap: Record<string, string> = {
      PENDING: t('sendStatus.pending'),
      SENDING: t('sendStatus.sending'),
      SUCCESS: t('sendStatus.success'),
      FAILED: t('sendStatus.failed'),
    }
    return statusMap[status] || String(status)
  }

  /**
   * 将发送状态映射为 RsTag variant。
   * @param status - 发送状态
   */
  const getSendStatusTagType = (status?: SendStatus | string | null): RsTagVariant => {
    if (!status) return 'default'
    const statusMap: Record<string, RsTagVariant> = {
      PENDING: 'default',
      SENDING: 'info',
      SUCCESS: 'success',
      FAILED: 'danger',
    }
    return statusMap[status] || 'default'
  }

  /** 按当前语言刷新搜索表单 / 表格文案 */
  function applyI18n() {
    const timeRangeShortcuts = [
      {
        label: t('shortcuts.today'),
        value: () => {
          const today = new Date()
          const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0)
          const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59)
          return {
            start: formatDate(startOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(endOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
      {
        label: t('shortcuts.yesterday'),
        value: () => {
          const yesterday = new Date()
          yesterday.setDate(yesterday.getDate() - 1)
          const startOfDay = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate(), 0, 0, 0)
          const endOfDay = new Date(yesterday.getFullYear(), yesterday.getMonth(), yesterday.getDate(), 23, 59, 59)
          return {
            start: formatDate(startOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(endOfDay.getTime(), 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
      {
        label: t('shortcuts.lastHour'),
        value: () => {
          const now = Date.now()
          return {
            start: formatDate(now - 3600000, 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
      {
        label: t('shortcuts.last6Hours'),
        value: () => {
          const now = Date.now()
          return {
            start: formatDate(now - 21600000, 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
      {
        label: t('shortcuts.last24Hours'),
        value: () => {
          const now = Date.now()
          return {
            start: formatDate(now - 86400000, 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
      {
        label: t('shortcuts.last7Days'),
        value: () => {
          const now = Date.now()
          return {
            start: formatDate(now - 604800000, 'YYYY-MM-DD HH:mm:ss'),
            end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
          }
        },
      },
    ]

    searchFormConfig.fields = [
      {
        field: 'timeRange',
        label: t('search.timeRange'),
        type: 'datetimerange',
        placeholder: t('search.timeRangePlaceholder'),
        span: 8,
        clearable: true,
        required: true,
        rules: [
          {
            validator: (value: unknown) => {
              if (
                value &&
                typeof value === 'object' &&
                !Array.isArray(value) &&
                (value as { start?: unknown; end?: unknown }).start &&
                (value as { start?: unknown; end?: unknown }).end
              ) {
                return true
              }
              return t('search.timeRangeRequired')
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
        label: t('search.alertLogId'),
        type: 'input',
        placeholder: t('search.alertLogIdPlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'alertLevel',
        label: t('search.alertLevel'),
        type: 'select',
        placeholder: t('search.alertLevelPlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('level.info'), value: 'INFO' },
          { label: t('level.warn'), value: 'WARN' },
          { label: t('level.error'), value: 'ERROR' },
          { label: t('level.critical'), value: 'CRITICAL' },
        ],
      },
      {
        field: 'alertType',
        label: t('search.alertType'),
        type: 'input',
        placeholder: t('search.alertTypePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'alertTitle',
        label: t('search.alertTitle'),
        type: 'input',
        placeholder: t('search.alertTitlePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'channelName',
        label: t('search.channelName'),
        type: 'input',
        placeholder: t('search.channelNamePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'sendStatus',
        label: t('search.sendStatus'),
        type: 'select',
        placeholder: t('search.sendStatusPlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('sendStatus.pending'), value: 'PENDING' },
          { label: t('sendStatus.sending'), value: 'SENDING' },
          { label: t('sendStatus.success'), value: 'SUCCESS' },
          { label: t('sendStatus.failed'), value: 'FAILED' },
        ],
      },
    ]
    searchFormConfig.toolbarButtons = [
      {
        key: 'delete',
        label: t('toolbar.delete'),
        icon: 'TrashOutline',
        type: 'error',
        tooltip: t('toolbar.deleteTooltip'),
      },
    ]

    gridConfig.columns = [
      {
        key: 'alertLogId',
        title: t('columns.alertLogId'),
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'alertLevel',
        title: t('columns.alertLevel'),
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: getAlertLevelTagType(row.alertLevel), size: 'sm' },
            () => getAlertLevelLabel(row.alertLevel),
          ),
      },
      {
        key: 'alertType',
        title: t('columns.alertType'),
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'alertTitle',
        title: t('columns.alertTitle'),
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'alertContent',
        title: t('columns.alertContent'),
        align: 'center',
        ellipsis: true,
        width: 300,
      },
      {
        key: 'channelName',
        title: t('columns.channelName'),
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'sendStatus',
        title: t('columns.sendStatus'),
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: getSendStatusTagType(row.sendStatus), size: 'sm' },
            () => getSendStatusLabel(row.sendStatus),
          ),
      },
      {
        key: 'alertTimestamp',
        title: t('columns.alertTimestamp'),
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
      {
        key: 'sendTime',
        title: t('columns.sendTime'),
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
      {
        key: 'sendErrorMessage',
        title: t('columns.sendErrorMessage'),
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'addTime',
        title: t('columns.addTime'),
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
      {
        key: 'addWho',
        title: t('columns.addWho'),
        align: 'center',
        width: 120,
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: t('columns.editTime'),
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
      {
        key: 'editWho',
        title: t('columns.editWho'),
        align: 'center',
        width: 120,
        ellipsis: true,
      },
    ]
    gridConfig.menuConfig = {
      enabled: true,
      items: [
        { key: 'view', label: t('common.viewDetail'), icon: 'eye' },
        { key: 'delete', label: t('common.delete'), icon: 'trash-2', danger: true },
      ],
    }
  }

  watch(locale, applyI18n, { immediate: true })

  /**
   * 设置日志列表数据。
   * @param list - 预警日志数组
   */
  function setLogList(list: AlertLog[]) {
    logList.value = list
  }

  /**
   * 设置加载状态。
   * @param value - 是否加载中
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /** 重置分页信息，下次查询从第一页开始 */
  function resetPagination() {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）。
   * @param newPageInfo - 部分分页字段
   */
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
