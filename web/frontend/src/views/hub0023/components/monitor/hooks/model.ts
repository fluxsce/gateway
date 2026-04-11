/**
 * Hub0023 监控模块 Model
 * 统一管理搜索表单配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import { createBackendPaginationParams } from '@/components/gpage'
import { queryGatewayInstances } from '@/views/hub0020/api'
import { RefreshOutline } from '@vicons/ionicons5'
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

/**
 * 默认时间范围：开始为当前时刻往前 1 小时，结束为当日 23:59:00。
 * 若两者跨度超过 24 小时（例如凌晨附近），则将开始时间推迟为「结束时间往前 24 小时」，以满足校验。
 */
function defaultMonitoringTimeRange(): [number, number] {
  const now = Date.now()
  const startCandidate = now - 3600000
  const end = new Date()
  end.setHours(23, 59, 0, 0)
  const endMs = end.getTime()
  let startMs = startCandidate
  if (endMs - startMs > MONITORING_MAX_RANGE_MS) {
    startMs = endMs - MONITORING_MAX_RANGE_MS
  }
  return [startMs, endMs]
}

/**
 * 监控模块 Model
 */
export function useMonitoringModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0023-monitor'
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

  /** 时间范围 */
  const timeRange = ref<[number, number] | null>(null)

  /** 时间粒度 */
  const timeGranularity = ref<TimeGranularity>('MINUTE' as TimeGranularity)

  /** 最近一次查询使用的网关实例（表单 reset 后可能短暂为空，用于恢复条件） */
  const lastGatewayInstanceId = ref('')
  const lastGatewayInstanceName = ref('')

  // ============= 搜索表单配置 =============

  /** 默认时间范围（当前时刻前 1 小时 ～ 当日 23:59） */
  const initTimeRange = (): [number, number] => defaultMonitoringTimeRange()

  /** 时间范围快捷选项（限制在24小时内） */
  const timeRangeShortcuts = {
    '前1小时至今日23:59': () => defaultMonitoringTimeRange(),
    最近1小时: () => {
      const now = Date.now()
      return [now - 3600000, now] as [number, number]
    },
    最近6小时: () => {
      const now = Date.now()
      return [now - 21600000, now] as [number, number]
    },
    最近12小时: () => {
      const now = Date.now()
      return [now - 43200000, now] as [number, number]
    },
    最近24小时: () => {
      const now = Date.now()
      return [now - 86400000, now] as [number, number]
    },
  }

  /** 时间粒度选项 */
  const timeGranularityOptions = [
    { label: '按分钟', value: 'MINUTE' as TimeGranularity },
    { label: '按小时', value: 'HOUR' as TimeGranularity },
    { label: '按天', value: 'DAY' as TimeGranularity },
  ]

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
            validator: (_rule: any, value: any) => {
              if (!value || !Array.isArray(value) || value.length !== 2 || !value[0] || !value[1]) {
                return new Error('请选择时间范围')
              }
              const startTime = value[0]
              const endTime = value[1]
              const duration = endTime - startTime
              const maxDuration = 24 * 60 * 60 * 1000 // 24小时的毫秒数
              if (duration > maxDuration) {
                return new Error('时间范围不能超过24小时')
              }
              if (duration <= 0) {
                return new Error('结束时间必须大于开始时间')
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
            validator: (_rule: any, value: any) => {
              if (value === undefined || value === null || String(value).trim() === '') {
                return new Error('请选择或输入网关实例名称')
              }
              return true
            },
            trigger: ['change', 'blur', 'input'],
          },
        ],
        render: (formData: Record<string, any>) => {
          return h(GatewayInstanceNameSelector, {
            modelValue: formData.gatewayInstanceName || '',
            gatewayInstanceId: formData.gatewayInstanceId || '',
            'onUpdate:modelValue': (value: string) => {
              formData.gatewayInstanceName = value
            },
            'onUpdate:gatewayInstanceId': (value: string) => {
              formData.gatewayInstanceId = value
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
        icon: RefreshOutline,
        type: 'primary',
        tooltip: '刷新监控数据',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
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
  const bootstrapDefaultGatewayInstance = async (searchFormRef: Ref<any>) => {
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

