/**
 * Hub0023 网关日志管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate, formatFileSize } from '@/utils/format'
import { DownloadOutline, EyeOutline, RefreshOutline } from '@vicons/ionicons5'
import { h, ref } from 'vue'
import type { GatewayLogListItem } from '../../../types'
import { GatewayInstanceNameSelector } from '../../instance-grid'
import { RouteNameSelector } from '../../route-grid'
import { ServiceNameSelector } from '../../service-grid'

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

  /** 初始化当天时间范围 */
  const initTodayTimeRange = (): [number, number] => {
    const today = new Date()
    const startOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 0, 0, 0)
    const endOfDay = new Date(today.getFullYear(), today.getMonth(), today.getDate(), 23, 59, 59)
    return [startOfDay.getTime(), endOfDay.getTime()]
  }

  /** 时间范围快捷选项 */
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
    }
  }

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'timeRange',
        label: '时间范围',
        type: 'datetimerange',
        placeholder: '请选择时间范围',
        span: 8,
        clearable: true,
        rules: [
          {
            validator: (_rule: any, value: any) => {
              if (!value || !Array.isArray(value) || value.length !== 2 || !value[0] || !value[1]) {
                return new Error('请选择时间范围')
              }
              return true
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
        render: (formData: Record<string, any>) => {
          return h(GatewayInstanceNameSelector, {
            modelValue: formData.gatewayInstanceName || '',
            'onUpdate:modelValue': (value: string) => {
              formData.gatewayInstanceName = value
            },
          })
        },
      },
      {
        field: 'routeName',
        label: '路由名称',
        type: 'custom',
        span: 8,
        render: (formData: Record<string, any>) => {
          return h(RouteNameSelector, {
            modelValue: formData.routeName || '',
            'onUpdate:modelValue': (value: string) => {
              formData.routeName = value
            },
            gatewayInstanceId: formData.gatewayInstanceName ? undefined : undefined, // 可以根据需要传递实例ID
          })
        },
      },
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'custom',
        span: 8,
        render: (formData: Record<string, any>) => {
          return h(ServiceNameSelector, {
            modelValue: formData.serviceName || '',
            'onUpdate:modelValue': (value: string) => {
              formData.serviceName = value
            },
            gatewayInstanceId: formData.gatewayInstanceName ? undefined : undefined, // 可以根据需要传递实例ID
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
        label: '只显示错误',
        type: 'switch',
        span: 8,
        defaultValue: false,
      },
    ],
    toolbarButtons: [
      {
        key: 'view',
        label: '查看详情',
        icon: EyeOutline,
        type: 'primary',
        tooltip: '查看选中日志的详情',
      },
      {
        key: 'batchReset',
        label: '批量重发',
        type: 'warning',
        icon: RefreshOutline,
        tooltip: '批量重发选中的日志',
      },
      {
        key: 'export',
        label: '导出日志',
        type: 'info',
        icon: DownloadOutline,
        tooltip: '导出日志数据',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    rowId: 'traceId',
    columns: [
      {
        field: 'traceId',
        title: '链路追踪ID',
        width: 140,
        showOverflow: 'tooltip',
      },
      {
        field: 'gatewayInstanceName',
        title: '网关实例',
        width: 120,
        showOverflow: 'tooltip',
        formatter: ({ row }: any) => row.gatewayInstanceName || row.gatewayInstanceId || '-',
      },
      {
        field: 'routeName',
        title: '路由名称',
        width: 120,
        showOverflow: 'tooltip',
        slots: { default: 'routeName' },
      },
      {
        field: 'requestPath',
        title: '请求路径',
        width: 200,
        showOverflow: 'tooltip',
      },
      {
        field: 'requestMethod',
        title: '请求方法',
        width: 80,
        align: 'center',
        slots: { default: 'requestMethod' },
      },
      {
        field: 'gatewayStatusCode',
        title: '状态码',
        width: 80,
        align: 'center',
        sortable: true,
        slots: { default: 'gatewayStatusCode' },
      },
      {
        field: 'processingStatus',
        title: '处理状态',
        width: 100,
        align: 'center',
        slots: { default: 'processingStatus' },
      },
      {
        field: 'totalProcessingTimeMs',
        title: '总处理时间',
        width: 110,
        align: 'center',
        sortable: true,
        slots: { default: 'totalProcessingTimeMs' },
      },
      {
        field: 'gatewayStartProcessingTime',
        title: '开始时间',
        width: 200,
        sortable: true,
        formatter: ({ cellValue }: any) => formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss.SSS'),
      },
      {
        field: 'gatewayFinishedProcessingTime',
        title: '完成时间',
        width: 200,
        sortable: true,
        formatter: ({ cellValue }: any) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss.SSS') : '-',
      },
      {
        field: 'clientIpAddress',
        title: '客户端IP',
        width: 160,
      },
      {
        field: 'backendStatusCode',
        title: '后端状态码',
        width: 100,
        align: 'center',
        sortable: true,
        slots: { default: 'backendStatusCode' },
      },
      {
        field: 'gatewayProcessingTimeMs',
        title: '网关耗时',
        width: 100,
        align: 'center',
        sortable: true,
        slots: { default: 'gatewayProcessingTimeMs' },
      },
      {
        field: 'backendResponseTimeMs',
        title: '后端耗时',
        width: 100,
        align: 'center',
        sortable: true,
        slots: { default: 'backendResponseTimeMs' },
      },
      {
        field: 'proxyType',
        title: '代理类型',
        width: 100,
        align: 'center',
        slots: { default: 'proxyType' },
      },
      {
        field: 'resetFlag',
        title: '重置状态',
        width: 100,
        align: 'center',
        slots: { default: 'resetFlag' },
      },
      {
        field: 'userIdentifier',
        title: '用户标识',
        width: 120,
        showOverflow: 'tooltip',
        formatter: ({ cellValue }: any) => cellValue || '-',
      },
      {
        field: 'requestSize',
        title: '请求大小',
        width: 100,
        formatter: ({ cellValue }: any) => cellValue != null ? formatFileSize(cellValue) : '-',
      },
      {
        field: 'responseSize',
        title: '响应大小',
        width: 100,
        formatter: ({ cellValue }: any) => cellValue != null ? formatFileSize(cellValue) : '-',
      },
      {
        field: 'resetCount',
        title: '重置次数',
        width: 100,
        formatter: ({ cellValue }: any) => cellValue != null ? cellValue : '-',
      },
      {
        field: 'errorMessage',
        title: '错误信息',
        width: 150,
        showOverflow: 'tooltip',
        formatter: ({ cellValue }: any) => cellValue || '-',
      },
      {
        field: 'logLevel',
        title: '日志级别',
        width: 100,
        align: 'center',
        slots: { default: 'logLevel' },
      },
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
      showCopyCell: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
        {
          code: 'reset',
          name: '重发',
          prefixIcon: 'vxe-icon-refresh',
        },
      ],
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

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

  // ============= 表格渲染辅助方法 =============

  /**
   * 获取请求方法的标签类型
   */
  const getMethodTagType = (method?: string): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    const methodColors: Record<string, 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary'> = {
      GET: 'success',
      POST: 'info',
      PUT: 'warning',
      DELETE: 'error',
      PATCH: 'default',
      HEAD: 'default',
      OPTIONS: 'default',
    }
    return methodColors[method || ''] || 'default'
  }

  /**
   * 获取状态码的标签类型
   */
  const getStatusCodeTagType = (statusCode?: number): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    if (!statusCode) return 'default'
    if (statusCode >= 200 && statusCode < 300) {
      return 'success'
    } else if (statusCode >= 300 && statusCode < 400) {
      return 'warning'
    } else if (statusCode >= 400) {
      return 'error'
    }
    return 'default'
  }

  /**
   * 获取耗时的标签类型
   */
  const getTimeTagType = (
    time: number,
    errorThreshold: number,
    warningThreshold: number
  ): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    if (time > errorThreshold) {
      return 'error'
    } else if (time > warningThreshold) {
      return 'warning'
    } else {
      return 'success'
    }
  }

  /**
   * 获取处理状态的标签类型
   */
  const getProcessingStatusTagType = (row: GatewayLogListItem): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    const isFinished = !!row.gatewayFinishedProcessingTime
    const hasError = !!row.errorMessage

    if (hasError) {
      return 'error'
    } else if (isFinished) {
      return 'success'
    } else {
      return 'warning'
    }
  }

  /**
   * 获取处理状态文本
   */
  const getProcessingStatusText = (row: GatewayLogListItem): string => {
    const isFinished = !!row.gatewayFinishedProcessingTime
    const hasError = !!row.errorMessage

    if (hasError) {
      return '异常'
    } else if (isFinished) {
      return '已完成'
    } else {
      return '处理中'
    }
  }

  /**
   * 获取代理类型的标签类型
   */
  const getProxyTypeTagType = (proxyType?: string): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    const typeColors: Record<string, 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary'> = {
      http: 'info',
      websocket: 'warning',
      tcp: 'success',
      udp: 'error',
    }
    return typeColors[proxyType || ''] || 'default'
  }

  /**
   * 获取日志级别的标签类型
   */
  const getLogLevelTagType = (logLevel?: string): 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary' => {
    const levelColors: Record<string, 'default' | 'error' | 'success' | 'info' | 'warning' | 'primary'> = {
      DEBUG: 'default',
      INFO: 'info',
      WARN: 'warning',
      ERROR: 'error',
    }
    return levelColors[logLevel || ''] || 'default'
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

    // 表格渲染辅助方法
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
 * Model 返回类型
 */
export type GatewayLogModel = ReturnType<typeof useGatewayLogModel>

