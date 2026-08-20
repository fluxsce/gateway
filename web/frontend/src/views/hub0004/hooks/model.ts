/**
 * 审计日志管理 Model
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
import type { AuditAction, AuditResult, AuthAuditLog } from '../types'

const MODULE_CODES = [
  'hub0000',
  'hub0001',
  'hub0002',
  'hub0003',
  'hub0004',
  'hub0005',
  'hub0006',
  'hub0007',
  'hub0008',
  'hub0020',
  'hub0021',
  'hub0022',
  'hub0023',
  'hub0040',
  'hub0041',
  'hub0042',
  'hub0043',
  'hub0060',
  'hub0061',
  'hub0062',
  'hub0080',
  'hub0081',
  'hub0082',
  'hubplugin',
] as const

const TARGET_TYPES = [
  'USER',
  'ROLE',
  'RESOURCE',
  'SECURITY_CONFIG',
  'INSTANCE',
  'ROUTE',
  'SERVICE',
  'ALERT_LOG',
  'HTTP',
  'API',
  'PROXY',
  'STATIC_HOST',
  'ROUTER_CONFIG',
  'ASSERTION',
  'FILTER',
  'CIRCUIT_BREAKER',
  'LOG_CONFIG',
  'ALERT_TEMPLATE',
  'ALERT_CHANNEL',
  'NAMESPACE',
  'CONFIG_DATA',
  'NODE',
  'TASK',
  'SCHEDULER',
  'TUNNEL_SERVER',
  'TUNNEL_CLIENT',
  'TUNNEL_SERVICE',
  'TUNNEL_STATIC',
  'SERVICE_CENTER',
] as const

/**
 * 审计日志表格配置（对齐 RsGrid Props 子集）。
 */
export interface AuditLogGridConfig {
  columns: RsGridColumn<AuthAuditLog>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 审计日志管理 Model
 */
export function useAuditLogModel() {
  const { t, locale } = useModuleI18n('hub0004')
  const moduleId = 'hub0004'

  const loading = ref(false)
  const logList = ref<AuthAuditLog[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  const searchFormConfig = reactive<Omit<RsSearchFormProps, 'moduleId'>>({
    fields: [],
    moreFields: [],
    toolbarButtons: [],
    showSearchButton: true,
    showResetButton: true,
  })

  const gridConfig = reactive<AuditLogGridConfig>({
    columns: [],
    selectable: false,
    rowKey: 'auditId',
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

  /** 默认最近 7 天（RsDatePicker valueFormat=string 的 range 形态） */
  const initLast7DaysRange = (): { start: string; end: string } => {
    const now = Date.now()
    return {
      start: formatDate(now - 604800000, 'YYYY-MM-DD HH:mm:ss'),
      end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
    }
  }

  /**
   * 获取动作显示文案。
   * @param action - 审计动作
   */
  const getActionLabel = (action?: AuditAction | string | null) => {
    if (!action) return ''
    const map: Record<string, string> = {
      CREATE: t('action.create'),
      UPDATE: t('action.update'),
      DELETE: t('action.delete'),
      ROLLBACK: t('action.rollback'),
      GRANT: t('action.grant'),
      EXPORT: t('action.export'),
      LOGIN: t('action.login'),
      LOGIN_FAIL: t('action.loginFail'),
      KICK: t('action.kick'),
    }
    return map[action] || String(action)
  }

  /**
   * 将动作映射为 RsTag variant。
   * @param action - 审计动作
   */
  const getActionTagType = (action?: AuditAction | string | null): RsTagVariant => {
    if (!action) return 'default'
    const map: Record<string, RsTagVariant> = {
      CREATE: 'success',
      UPDATE: 'info',
      DELETE: 'danger',
      ROLLBACK: 'warning',
      GRANT: 'warning',
      EXPORT: 'info',
      LOGIN: 'success',
      LOGIN_FAIL: 'danger',
      KICK: 'warning',
    }
    return map[action] || 'default'
  }

  /**
   * 获取结果显示文案。
   * @param result - 审计结果
   */
  const getResultLabel = (result?: AuditResult | string | null) => {
    if (!result) return ''
    if (result === 'Y') return t('result.success')
    if (result === 'N') return t('result.fail')
    return String(result)
  }

  /**
   * 将结果映射为 RsTag variant。
   * @param result - 审计结果
   */
  const getResultTagType = (result?: AuditResult | string | null): RsTagVariant => {
    if (result === 'Y') return 'success'
    if (result === 'N') return 'danger'
    return 'default'
  }

  /**
   * 获取模块显示文案。
   * @param code - 模块编码
   */
  const getModuleLabel = (code?: string | null) => {
    if (!code) return ''
    const key = `module.${code}`
    const label = t(key)
    if (label && label !== key) {
      return `${code} ${label}`
    }
    return String(code)
  }

  /**
   * 获取目标类型显示文案。
   * @param type - 目标类型
   */
  const getTargetTypeLabel = (type?: string | null) => {
    if (!type) return ''
    const key = `targetType.${type}`
    const label = t(key)
    if (label && label !== key) return label
    return String(type)
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

    searchFormConfig.toolbarButtons = [
      {
        key: 'export',
        label: t('common.export'),
        icon: 'DownloadOutline',
        tooltip: t('common.export'),
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
        props: {
          shortcuts: timeRangeShortcuts,
          style: { width: '100%' },
        },
        defaultValue: initLast7DaysRange(),
      },
      {
        field: 'action',
        label: t('search.action'),
        type: 'select',
        placeholder: t('search.actionPlaceholder'),
        span: 5,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('action.create'), value: 'CREATE' },
          { label: t('action.update'), value: 'UPDATE' },
          { label: t('action.delete'), value: 'DELETE' },
          { label: t('action.rollback'), value: 'ROLLBACK' },
          { label: t('action.grant'), value: 'GRANT' },
          { label: t('action.export'), value: 'EXPORT' },
          { label: t('action.login'), value: 'LOGIN' },
          { label: t('action.loginFail'), value: 'LOGIN_FAIL' },
          { label: t('action.kick'), value: 'KICK' },
        ],
      },
      {
        field: 'moduleCode',
        label: t('search.moduleCode'),
        type: 'select',
        placeholder: t('search.moduleCodePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          ...MODULE_CODES.map((code) => ({
            label: getModuleLabel(code),
            value: code,
          })),
        ],
      },
      {
        field: 'userName',
        label: t('search.userName'),
        type: 'input',
        placeholder: t('search.userNamePlaceholder'),
        span: 5,
        clearable: true,
      },
    ]
    searchFormConfig.moreFields = [
      {
        field: 'result',
        label: t('search.result'),
        type: 'select',
        placeholder: t('search.resultPlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          { label: t('result.success'), value: 'Y' },
          { label: t('result.fail'), value: 'N' },
        ],
      },
      {
        field: 'targetType',
        label: t('search.targetType'),
        type: 'select',
        placeholder: t('search.targetTypePlaceholder'),
        span: 6,
        clearable: true,
        options: [
          { label: t('common.all'), value: '' },
          ...TARGET_TYPES.map((type) => ({
            label: getTargetTypeLabel(type),
            value: type,
          })),
        ],
      },
      {
        field: 'targetName',
        label: t('search.targetName'),
        type: 'input',
        placeholder: t('search.targetNamePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'targetId',
        label: t('search.targetId'),
        type: 'input',
        placeholder: t('search.targetIdPlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'resourceCode',
        label: t('search.resourceCode'),
        type: 'input',
        placeholder: t('search.resourceCodePlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'userId',
        label: t('search.userId'),
        type: 'input',
        placeholder: t('search.userIdPlaceholder'),
        span: 6,
        clearable: true,
      },
      {
        field: 'clientIP',
        label: t('search.clientIP'),
        type: 'input',
        placeholder: t('search.clientIPPlaceholder'),
        span: 6,
        clearable: true,
      },
    ]

    gridConfig.columns = [
      {
        key: 'addTime',
        title: t('columns.addTime'),
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 170,
      },
      {
        key: 'userName',
        title: t('columns.userName'),
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'action',
        title: t('columns.action'),
        align: 'center',
        width: 90,
        render: (row) =>
          h(
            RsTag,
            { variant: getActionTagType(row.action), size: 'sm' },
            () => getActionLabel(row.action),
          ),
      },
      {
        key: 'moduleCode',
        title: t('columns.moduleCode'),
        align: 'center',
        ellipsis: true,
        width: 180,
        formatter: (_value, row) => getModuleLabel(row.moduleCode),
      },
      {
        key: 'targetType',
        title: t('columns.targetType'),
        align: 'center',
        width: 110,
        formatter: (_value, row) => getTargetTypeLabel(row.targetType) || '-',
      },
      {
        key: 'targetName',
        title: t('columns.targetName'),
        align: 'center',
        ellipsis: true,
        width: 160,
      },
      {
        key: 'targetId',
        title: t('columns.targetId'),
        align: 'center',
        ellipsis: true,
        width: 180,
      },
      {
        key: 'resourceCode',
        title: t('columns.resourceCode'),
        align: 'center',
        ellipsis: true,
        width: 160,
      },
      {
        key: 'result',
        title: t('columns.result'),
        align: 'center',
        width: 80,
        render: (row) =>
          h(
            RsTag,
            { variant: getResultTagType(row.result), size: 'sm' },
            () => getResultLabel(row.result),
          ),
      },
      {
        key: 'clientIP',
        title: t('columns.clientIP'),
        align: 'center',
        ellipsis: true,
        width: 130,
      },
      {
        key: 'requestPath',
        title: t('columns.requestPath'),
        align: 'left',
        ellipsis: true,
        minWidth: 220,
      },
    ]
    gridConfig.menuConfig = {
      enabled: true,
      items: [{ key: 'view', label: t('common.viewDetail'), icon: 'eye' }],
    }
  }

  watch(locale, applyI18n, { immediate: true })

  /**
   * 设置日志列表数据。
   * @param list - 审计日志数组
   */
  function setLogList(list: AuthAuditLog[]) {
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

    getActionLabel,
    getActionTagType,
    getResultLabel,
    getResultTagType,
    getModuleLabel,
    getTargetTypeLabel,

    setLogList,
    setLoading,
    resetPagination,
    updatePagination,
  }
}

export type AuditLogModel = ReturnType<typeof useAuditLogModel>
