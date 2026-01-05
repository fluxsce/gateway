/**
 * Hub0023 监控模块 Model
 * 统一管理搜索表单配置和数据状态
 */

import type { SearchFormProps } from '@/components/form/search/types'
import { RefreshOutline } from '@vicons/ionicons5'
import { reactive, ref } from 'vue'
import type {
  GatewayMonitoringChartData,
  GatewayMonitoringOverview,
  TimeGranularity,
} from './types'

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

  // ============= 搜索表单配置 =============

  /** 初始化最近1小时时间范围 */
  const initTimeRange = (): [number, number] => {
    const now = Date.now()
    return [now - 3600000, now]
  }

  /** 时间范围快捷选项（限制在24小时内） */
  const timeRangeShortcuts = {
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
        field: 'timeRange',
        label: '时间范围',
        type: 'datetimerange',
        placeholder: '请选择时间范围（必填，不超过24小时）',
        span: 8,
        clearable: true,
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
        field: 'routeName',
        label: '路由名称',
        type: 'input',
        placeholder: '请输入路由名称（可选）',
        span: 8,
        clearable: true,
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

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    overviewData,
    chartData,
    timeRange,
    timeGranularity,

    // 配置
    searchFormConfig,
    timeRangeShortcuts,
    timeGranularityOptions,

    // 方法
    initTimeRange,
    getTimeGranularityLabel,
    resetOverviewData,
    resetChartData,
  }
}

/**
 * Model 返回类型
 */
export type MonitoringModel = ReturnType<typeof useMonitoringModel>

