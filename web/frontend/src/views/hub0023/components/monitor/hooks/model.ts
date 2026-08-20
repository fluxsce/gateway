/**
 * Hub0023 监控模块 Model
 * 统一管理搜索表单配置和数据状态
 */

import type { RsSearchFormExpose, RsSearchFormProps } from '@/components/form/rs-search'
import { formatDate } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import { queryGatewayInstances } from '@/views/hub0020/api'
import type { Ref } from 'vue'
import { h, nextTick, reactive, ref } from 'vue'
import { GatewayInstanceNameSelector } from '../../instance-grid'
import { RouteNameSelector } from '../../route-grid'
import type {
  GatewayMonitoringChartData,
  GatewayMonitoringOverview,
  TimeGranularity,
} from './types'

/** 监控查询允许的最大时间跨度（与表单校验一致：24 小时） */
const MONITORING_MAX_RANGE_MS = 24 * 60 * 60 * 1000

type DateTimeRangeValue = { start: string; end: string }

/** 将 RsDatePicker range（{ start, end }）转为毫秒，供跨度校验 */
function rangeToMs(value: unknown): [number, number] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const { start, end } = value as { start?: string; end?: string }
  if (!start || !end) return null
  const startMs = new Date(start.replace(' ', 'T')).getTime()
  const endMs = new Date(end.replace(' ', 'T')).getTime()
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) return null
  return [startMs, endMs]
}

/**
 * 默认时间范围：开始为当前时刻往前 1 小时，结束为当日 23:59:00。
 * 若两者跨度超过 24 小时（例如凌晨附近），则将开始时间推迟为「结束时间往前 24 小时」，以满足校验。
 */
function defaultMonitoringTimeRange(): DateTimeRangeValue {
  const now = Date.now()
  const startCandidate = now - 3600000
  const end = new Date()
  end.setHours(23, 59, 0, 0)
  const endMs = end.getTime()
  let startMs = startCandidate
  if (endMs - startMs > MONITORING_MAX_RANGE_MS) {
    startMs = endMs - MONITORING_MAX_RANGE_MS
  }
  return {
    start: formatDate(startMs, 'YYYY-MM-DD HH:mm:ss'),
    end: formatDate(endMs, 'YYYY-MM-DD HH:mm:ss'),
  }
}

/**
 * 监控模块 Model
 */
export function useMonitoringModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023'
  /** 加载状态 */
  const loading = ref(false)

  /** 监控概览数据 */
  const overviewData = reactive<GatewayMonitoringOverview>({
    totalRequests: 0,
    successRequests: 0,
    failedRequests: 0,
    requestsPerSecond: 0,
    avgResponseTimeMs: 0,
    minResponseTimeMs: 0,
    maxResponseTimeMs: 0,
  })

  /** 图表数据 */
  const chartData = reactive<GatewayMonitoringChartData>({
    requestTrend: [],
    responseTimeTrend: [],
    statusCodeDistribution: [],
    hotRoutes: [],
  })

  /** 时间范围（RsDatePicker valueFormat=string） */
  const timeRange = ref<DateTimeRangeValue | null>(null)

  /** 时间粒度 */
  const timeGranularity = ref<TimeGranularity>('MINUTE' as TimeGranularity)

  /** 最近一次查询使用的网关实例（表单 reset 后可能短暂为空，用于恢复条件） */
  const lastGatewayInstanceId = ref('')
  const lastGatewayInstanceName = ref('')

  // ============= 搜索表单配置 =============

  /** 默认时间范围 */
  const initTimeRange = (): DateTimeRangeValue => defaultMonitoringTimeRange()

  /** 时间范围快捷选项（限制在24小时内） */
  const timeRangeShortcuts = [
    {
      label: '前1小时至今日23:59',
      value: () => defaultMonitoringTimeRange(),
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
      label: '最近12小时',
      value: () => {
        const now = Date.now()
        return {
          start: formatDate(now - 43200000, 'YYYY-MM-DD HH:mm:ss'),
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
  ]

  /** 时间粒度选项 */
  const timeGranularityOptions = [
    { label: '按分钟', value: 'MINUTE' as TimeGranularity },
    { label: '按小时', value: 'HOUR' as TimeGranularity },
    { label: '按天', value: 'DAY' as TimeGranularity },
  ]

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
        placeholder: '请选择时间范围（必填，不超过24小时）',
        span: 8,
        clearable: true,
        required: true,
        rules: [
          {
            validator: (value: unknown) => {
              const bounds = rangeToMs(value)
              if (!bounds) return '请选择时间范围'
              const duration = bounds[1] - bounds[0]
              if (duration > MONITORING_MAX_RANGE_MS) return '时间范围不能超过24小时'
              if (duration <= 0) return '结束时间必须大于开始时间'
              return true
            },
            trigger: ['change', 'blur'],
          },
        ],
        props: {
          shortcuts: timeRangeShortcuts,
          style: { width: '100%' },
        },
        defaultValue: initTimeRange(),
      },
      {
        field: 'timeGranularity',
        label: '时间粒度',
        type: 'select',
        placeholder: '请选择时间粒度',
        span: 8,
        clearable: false,
        options: timeGranularityOptions,
        defaultValue: 'MINUTE' as TimeGranularity,
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
        field: 'requestPath',
        label: '请求路径',
        type: 'input',
        placeholder: '请输入请求路径（可选）',
        span: 8,
        clearable: true,
      },
    ],
    toolbarButtons: [
      {
        key: 'refresh',
        label: '刷新数据',
        icon: 'RefreshOutline',
        type: 'primary',
        tooltip: '刷新监控数据',
        skipPermission: true,
      },
    ],
    showSearchButton: true,
    showResetButton: true,
    resetButtonKey: 'resetQuery',
  }

  /**
   * 获取时间粒度标签
   */
  const getTimeGranularityLabel = (): string => {
    switch (timeGranularity.value) {
      case 'MINUTE':
        return '按分钟'
      case 'HOUR':
        return '按小时'
      case 'DAY':
        return '按天'
      default:
        return '按分钟'
    }
  }

  /**
   * 重置概览数据
   */
  const resetOverviewData = () => {
    Object.assign(overviewData, {
      totalRequests: 0,
      successRequests: 0,
      failedRequests: 0,
      requestsPerSecond: 0,
      avgResponseTimeMs: 0,
      minResponseTimeMs: 0,
      maxResponseTimeMs: 0,
    })
  }

  /**
   * 重置图表数据
   */
  const resetChartData = () => {
    Object.assign(chartData, {
      requestTrend: [],
      responseTimeTrend: [],
      statusCodeDistribution: [],
      hotRoutes: [],
    })
  }

  /**
   * 拉取网关实例列表第一条并写入搜索表单（与网关日志一致）。
   * 监控页由 initPageData 统一拉数，此处不触发 submit。
   */
  const bootstrapDefaultGatewayInstance = async (searchFormRef: Ref<RsSearchFormExpose | null>) => {
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
      lastGatewayInstanceId.value = gatewayInstanceId
      lastGatewayInstanceName.value = instanceName || gatewayInstanceId
      await nextTick()
    } catch (e) {
      console.warn('[hub0023-monitor] 默认网关实例初始化失败', e)
    }
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    overviewData,
    chartData,
    timeRange,
    timeGranularity,
    lastGatewayInstanceId,
    lastGatewayInstanceName,

    // 配置
    searchFormConfig,
    timeRangeShortcuts,
    timeGranularityOptions,

    // 方法
    initTimeRange,
    getTimeGranularityLabel,
    resetOverviewData,
    resetChartData,
    bootstrapDefaultGatewayInstance,
  }
}

/**
 * Model 返回类型
 */
export type MonitoringModel = ReturnType<typeof useMonitoringModel>

