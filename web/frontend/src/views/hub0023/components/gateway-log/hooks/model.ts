/**
 * Hub0023 网关日志管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate, formatFileSize } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import { queryGatewayInstances } from '@/views/hub0020/api'
import type { Ref } from 'vue'
import { h, nextTick, ref } from 'vue'
import type { GatewayLogListItem } from '../../../types'
import { GatewayInstanceNameSelector } from '../../instance-grid'
import { RouteNameSelector } from '../../route-grid'
import { ServiceNameSelector } from '../../service-grid'

/** 网关日志表格配置（对齐 RsGrid Props 子集）。 */
export interface GatewayLogGridConfig {
  columns: RsGridColumn<GatewayLogListItem>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/** 获取请求方法的标签类型 */
function getMethodTagType(method?: string): RsTagVariant {
  const methodColors: Record<string, RsTagVariant> = {
    GET: 'success',
    POST: 'info',
    PUT: 'warning',
    DELETE: 'danger',
    PATCH: 'default',
    HEAD: 'default',
    OPTIONS: 'default',
  }
  return methodColors[method || ''] || 'default'
}

/** 获取状态码的标签类型 */
function getStatusCodeTagType(statusCode?: number): RsTagVariant {
  if (!statusCode) return 'default'
  if (statusCode >= 200 && statusCode < 300) return 'success'
  if (statusCode >= 300 && statusCode < 400) return 'warning'
  if (statusCode >= 400) return 'danger'
  return 'default'
}

/** 获取耗时的标签类型 */
function getTimeTagType(
  time: number,
  errorThreshold: number,
  warningThreshold: number,
): RsTagVariant {
  if (time > errorThreshold) return 'danger'
  if (time > warningThreshold) return 'warning'
  return 'success'
}

/** 获取处理状态的标签类型 */
function getProcessingStatusTagType(row: GatewayLogListItem): RsTagVariant {
  if (row.errorMessage) return 'danger'
  if (row.gatewayFinishedProcessingTime) return 'success'
  return 'warning'
}

/** 获取处理状态文本 */
function getProcessingStatusText(row: GatewayLogListItem): string {
  if (row.errorMessage) return '异常'
  if (row.gatewayFinishedProcessingTime) return '已完成'
  return '处理中'
}

/** 获取代理类型的标签类型 */
function getProxyTypeTagType(proxyType?: string): RsTagVariant {
  const typeColors: Record<string, RsTagVariant> = {
    http: 'info',
    websocket: 'warning',
    tcp: 'success',
    udp: 'danger',
  }
  return typeColors[proxyType || ''] || 'default'
}

/** 获取日志级别的标签类型 */
function getLogLevelTagType(logLevel?: string): RsTagVariant {
  const levelColors: Record<string, RsTagVariant> = {
    DEBUG: 'default',
    INFO: 'info',
    WARN: 'warning',
    ERROR: 'danger',
  }
  return levelColors[logLevel || ''] || 'default'
}

/**
 * 网关日志管理 Model
 */
export function useGatewayLogModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023'
  /** 加载状态 */
  const loading = ref(false)

  /** 网关日志列表数据 */
  const logList = ref<GatewayLogListItem[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

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

  /** 时间范围快捷项（可返回时间戳；DatePicker 会规范为 { start, end }） */
  const timeRangeShortcuts = [
    {
      label: '今天',
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
      label: '昨天',
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
      label: '最近1小时',
      value: () => {
        const now = Date.now()
        return {
          start: formatDate(now - 3600000, 'YYYY-MM-DD HH:mm:ss'),
          end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
        }
      },
    },
    {
      label: '最近6小时',
      value: () => {
        const now = Date.now()
        return {
          start: formatDate(now - 21600000, 'YYYY-MM-DD HH:mm:ss'),
          end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
        }
      },
    },
    {
      label: '最近24小时',
      value: () => {
        const now = Date.now()
        return {
          start: formatDate(now - 86400000, 'YYYY-MM-DD HH:mm:ss'),
          end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
        }
      },
    },
    {
      label: '最近7天',
      value: () => {
        const now = Date.now()
        return {
          start: formatDate(now - 604800000, 'YYYY-MM-DD HH:mm:ss'),
          end: formatDate(now, 'YYYY-MM-DD HH:mm:ss'),
        }
      },
    },
  ]

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    // 覆盖最长标签（如「仅非200状态」「报文体关键字」），固定列宽保证查询列对齐
    labelWidth: '8rem',
    fields: [
      {
        field: 'gatewayInstanceId',
        label: '',
        type: 'input',
        show: false,
        defaultValue: '',
      },
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
              return '请选择时间范围'
            },
            trigger: ['change', 'blur']
          }
        ],
        props: {
          shortcuts: timeRangeShortcuts,
          style: { width: '100%' }
        },
        defaultValue: initTodayTimeRange()
      },
      {
        field: 'gatewayInstanceName',
        label: '实例名称',
        type: 'custom',
        span: 8,
        required: true,
        rules: [
          {
            validator: (value: unknown) => {
              if (value === undefined || value === null || String(value).trim() === '') {
                return '请选择或输入网关实例名称'
              }
              return true
            },
            trigger: ['change', 'blur'],
          },
        ],
        render: (formData: Record<string, any>, ctx) => {
          return h(GatewayInstanceNameSelector, {
            modelValue: (ctx.value as string) || '',
            gatewayInstanceId: formData.gatewayInstanceId || '',
            'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
            'onUpdate:gatewayInstanceId': (value: string) => {
              ctx.setFieldValue('gatewayInstanceId', value)
            },
          })
        },
      },
      {
        field: 'routeName',
        label: '路由名称',
        type: 'custom',
        span: 8,
        render: (formData: Record<string, any>, ctx) => {
          return h(RouteNameSelector, {
            modelValue: (ctx.value as string) || '',
            'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
            gatewayInstanceId: formData.gatewayInstanceId || undefined,
          })
        },
      },
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'custom',
        span: 8,
        render: (formData: Record<string, any>, ctx) => {
          return h(ServiceNameSelector, {
            modelValue: (ctx.value as string) || '',
            'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
            gatewayInstanceId: formData.gatewayInstanceId || undefined,
          })
        },
      },
      {
        field: 'minProcessingTime',
        label: '网关耗时',
        type: 'number',
        placeholder: '最小耗时(毫秒)',
        span: 8,
        clearable: true,
        props: {
          min: 0,
          style: { width: '100%' }
        }
      },
      {
        field: 'requestQueryKeyword',
        label: '参数关键字',
        type: 'input',
        placeholder: '子串匹配请求参数，如 userId=1001',
        span: 8,
        clearable: true,
      },
      {
        field: 'requestBodyKeyword',
        label: '报文体关键字',
        type: 'input',
        placeholder: '子串匹配报文体，需已开启记录请求体',
        span: 8,
        clearable: true,
      },
    ],
    moreFields: [
      {
        field: 'traceId',
        label: '链路追踪ID',
        type: 'input',
        placeholder: '请输入链路追踪ID',
        span: 8,
        clearable: true,
      },
      {
        field: 'requestPath',
        label: '请求路径',
        type: 'input',
        placeholder: '请输入请求路径',
        span: 8,
        clearable: true,
      },
      {
        field: 'clientIpAddress',
        label: '客户端IP',
        type: 'input',
        placeholder: '请输入客户端IP',
        span: 8,
        clearable: true,
      },
      {
        field: 'requestMethod',
        label: '请求方法',
        type: 'select',
        placeholder: '请选择请求方法',
        span: 8,
        clearable: true,
        options: [
          { label: 'GET', value: 'GET' },
          { label: 'POST', value: 'POST' },
          { label: 'PUT', value: 'PUT' },
          { label: 'DELETE', value: 'DELETE' },
          { label: 'PATCH', value: 'PATCH' },
          { label: 'HEAD', value: 'HEAD' },
          { label: 'OPTIONS', value: 'OPTIONS' }
        ],
      },
      {
        field: 'proxyType',
        label: '代理类型',
        type: 'select',
        placeholder: '请选择代理类型',
        span: 8,
        clearable: true,
        options: [
          { label: 'HTTP', value: 'http' },
          { label: 'WebSocket', value: 'websocket' },
          { label: 'TCP', value: 'tcp' },
          { label: 'UDP', value: 'udp' }
        ],
      },
      {
        field: 'gatewayStatusCode',
        label: '状态码',
        type: 'number',
        placeholder: '请输入状态码',
        span: 8,
        clearable: true,
        props: {
          min: 100,
          max: 599,
          style: { width: '100%' }
        }
      },
      {
        field: 'backendStatusCode',
        label: '后端状态码',
        type: 'number',
        placeholder: '请输入后端状态码',
        span: 8,
        clearable: true,
        props: {
          min: 100,
          max: 599,
          style: { width: '100%' }
        }
      },
      {
        field: 'resetFlag',
        label: '重置状态',
        type: 'select',
        placeholder: '请选择重置状态',
        span: 8,
        clearable: true,
        options: [
          { label: '未重置', value: 'N' },
          { label: '已重置', value: 'Y' }
        ],
      },
      {
        field: 'userIdentifier',
        label: '用户标识',
        type: 'input',
        placeholder: '请输入用户标识',
        span: 8,
        clearable: true,
      },
      {
        field: 'errorOnly',
        label: '仅非200状态',
        type: 'switch',
        span: 8,
        defaultValue: false,
      },
    ],
    toolbarButtons: [
      {
        key: 'view',
        label: '查看详情',
        icon: 'EyeOutline',
        type: 'primary',
        tooltip: '查看选中日志的详情',
      },
      {
        key: 'batchReset',
        label: '批量重发',
        type: 'warning',
        icon: 'RefreshOutline',
        tooltip: '批量重发选中的日志',
      },
      {
        key: 'export',
        label: '导出日志',
        type: 'info',
        icon: 'DownloadOutline',
        tooltip: '导出日志数据',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: GatewayLogGridConfig = {
    columns: [
      {
        key: 'traceId',
        title: '链路追踪ID',
        width: 140,
        ellipsis: true,
      },
      {
        key: 'gatewayInstanceName',
        title: '网关实例',
        width: 120,
        ellipsis: true,
        formatter: (_v, row) => row.gatewayInstanceName || row.gatewayInstanceId || '-',
      },
      {
        key: 'routeName',
        title: '路由名称',
        width: 120,
        ellipsis: true,
        render: (row) =>
          h('span', { class: 'route-name-text' }, row.routeName || '-'),
      },
      {
        key: 'requestPath',
        title: '请求路径',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'requestMethod',
        title: '请求方法',
        width: 80,
        align: 'center',
        render: (row) =>
          h(RsTag, { variant: getMethodTagType(row.requestMethod), size: 'sm' }, () => row.requestMethod),
      },
      {
        key: 'gatewayStatusCode',
        title: '状态码',
        width: 80,
        align: 'center',
        sortable: true,
        render: (row) =>
          h(
            RsTag,
            { variant: getStatusCodeTagType(row.gatewayStatusCode), size: 'sm' },
            () => row.gatewayStatusCode,
          ),
      },
      {
        key: 'processingStatus',
        title: '处理状态',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: getProcessingStatusTagType(row), size: 'sm' },
            () => getProcessingStatusText(row),
          ),
      },
      {
        key: 'totalProcessingTimeMs',
        title: '总处理时间',
        width: 110,
        align: 'center',
        sortable: true,
        render: (row) =>
          row.totalProcessingTimeMs != null
            ? h(
                RsTag,
                {
                  variant: getTimeTagType(row.totalProcessingTimeMs, 5000, 1000),
                  size: 'sm',
                },
                () => `${row.totalProcessingTimeMs}ms`,
              )
            : '-',
      },
      {
        key: 'gatewayStartProcessingTime',
        title: '开始时间',
        width: 200,
        sortable: true,
        formatter: (v) => formatDate(v as string, 'YYYY-MM-DD HH:mm:ss.SSS'),
      },
      {
        key: 'gatewayFinishedProcessingTime',
        title: '完成时间',
        width: 200,
        sortable: true,
        formatter: (v) => (v ? formatDate(v as string, 'YYYY-MM-DD HH:mm:ss.SSS') : '-'),
      },
      {
        key: 'clientIpAddress',
        title: '客户端IP',
        width: 160,
      },
      {
        key: 'backendStatusCode',
        title: '后端状态码',
        width: 100,
        align: 'center',
        sortable: true,
        render: (row) =>
          row.backendStatusCode != null
            ? h(
                RsTag,
                { variant: getStatusCodeTagType(row.backendStatusCode), size: 'sm' },
                () => row.backendStatusCode,
              )
            : '-',
      },
      {
        key: 'gatewayProcessingTimeMs',
        title: '网关耗时',
        width: 100,
        align: 'center',
        sortable: true,
        render: (row) =>
          row.gatewayProcessingTimeMs != null
            ? h(
                RsTag,
                {
                  variant: getTimeTagType(row.gatewayProcessingTimeMs, 2000, 500),
                  size: 'sm',
                },
                () => `${row.gatewayProcessingTimeMs}ms`,
              )
            : '-',
      },
      {
        key: 'backendResponseTimeMs',
        title: '后端耗时',
        width: 100,
        align: 'center',
        sortable: true,
        render: (row) =>
          row.backendResponseTimeMs != null
            ? h(
                RsTag,
                {
                  variant: getTimeTagType(row.backendResponseTimeMs, 3000, 1000),
                  size: 'sm',
                },
                () => `${row.backendResponseTimeMs}ms`,
              )
            : '-',
      },
      {
        key: 'proxyType',
        title: '代理类型',
        width: 100,
        align: 'center',
        render: (row) =>
          row.proxyType
            ? h(
                RsTag,
                { variant: getProxyTypeTagType(row.proxyType), size: 'sm' },
                () => row.proxyType!.toUpperCase(),
              )
            : '-',
      },
      {
        key: 'resetFlag',
        title: '重置状态',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: row.resetFlag === 'Y' ? 'warning' : 'default', size: 'sm' },
            () => (row.resetFlag === 'Y' ? '已重置' : '未重置'),
          ),
      },
      {
        key: 'userIdentifier',
        title: '用户标识',
        width: 120,
        ellipsis: true,
        formatter: (v) => (v as string) || '-',
      },
      {
        key: 'requestSize',
        title: '请求大小',
        width: 100,
        formatter: (v) => formatFileSize(v as number),
      },
      {
        key: 'responseSize',
        title: '响应大小',
        width: 100,
        formatter: (v) => formatFileSize(v as number),
      },
      {
        key: 'resetCount',
        title: '重置次数',
        width: 100,
        formatter: (v) => (v != null ? String(v) : '-'),
      },
      {
        key: 'errorMessage',
        title: '错误信息',
        width: 150,
        ellipsis: true,
        formatter: (v) => (v as string) || '-',
      },
      {
        key: 'logLevel',
        title: '日志级别',
        width: 100,
        align: 'center',
        render: (row) =>
          row.logLevel
            ? h(
                RsTag,
                { variant: getLogLevelTagType(row.logLevel), size: 'sm' },
                () => row.logLevel,
              )
            : '-',
      },
    ],
    selectable: true,
    rowKey: 'traceId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'reset', label: '重发', icon: 'refresh-cw' },
      ],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 设置日志列表
   */
  const setLogList = (list: GatewayLogListItem[]) => {
    logList.value = list
  }

  /**
   * 清空日志列表
   */
  const clearLogList = () => {
    logList.value = []
  }

  /**
   * 拉取网关实例列表第一条并写入搜索表单（与后端列表默认排序一致：pageIndex=1）。
   * 不触发网关日志查询；由用户点击「查询」或其它入口再拉取日志。
   */
  const bootstrapDefaultGatewayInstance = async (searchFormRef: Ref<any>) => {
    const existing = searchFormRef.value?.getFormData?.() || {}
    if (String(existing.gatewayInstanceId || existing.gatewayInstanceName || '').trim()) {
      return
    }
    try {
      const res = await queryGatewayInstances({
        ...createBackendPaginationParams(1, 1),
      })
      if (!res?.oK || !res.bizData) return
      const list = JSON.parse(res.bizData) as Array<{ instanceName?: string; gatewayInstanceId?: string }>
      const first = Array.isArray(list) ? list[0] : undefined
      if (!first) return
      const instanceName = (first.instanceName || first.gatewayInstanceId || '').trim()
      const gatewayInstanceId = (first.gatewayInstanceId || '').trim()
      if (!instanceName && !gatewayInstanceId) return
      searchFormRef.value?.setFormData({
        gatewayInstanceName: instanceName || gatewayInstanceId,
        gatewayInstanceId,
      })
      await nextTick()
    } catch (e) {
      console.warn('[hub0023] 默认网关实例初始化失败', e)
    }
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    logList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setLogList,
    clearLogList,
    bootstrapDefaultGatewayInstance,

    // 表格渲染辅助（兼容旧调用）
    getMethodTagType,
    getStatusCodeTagType,
    getTimeTagType,
    getProcessingStatusTagType,
    getProcessingStatusText,
    getProxyTypeTagType,
    getLogLevelTagType,
  }
}

/**
 * 网关日志 Model 类型
 */
export type GatewayLogModel = ReturnType<typeof useGatewayLogModel>
