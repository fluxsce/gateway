/**
 * 表外分析编排：指标卡 + 图表系列（库无关）。
 *
 * 数据源约定与 niuma-ui 一致：有选中 → 选中集；否则 → viewRows。
 * 柱状图/饼图消费 `chartSeries`，再由业务映射到 ECharts 等。
 */

import { computed, ref, type ComputedRef, type Ref } from 'vue'
import {
  useRsTableChartBridge,
  useRsTableSelectionSource,
  type RsTableChartSeries,
  type RsTableChartSeriesDef,
} from '@/ui'

export type RsTableAnalyticsAggType = 'count' | 'sum' | 'avg' | 'min' | 'max'

export interface RsTableAnalyticsMetricDef {
  key: string
  label: string
  field?: string
  type: RsTableAnalyticsAggType
  formatter?: (value: number) => string
}

export interface RsTableAnalyticsMetricValue {
  key: string
  label: string
  value: number
  text: string
}

function readNumber(row: Record<string, unknown>, field: string): number | null {
  const raw = row[field]
  if (typeof raw === 'number' && Number.isFinite(raw)) return raw
  if (typeof raw === 'string' && raw.trim() !== '') {
    const n = Number(raw)
    return Number.isFinite(n) ? n : null
  }
  return null
}

function aggregateMetric(
  rows: Record<string, unknown>[],
  def: RsTableAnalyticsMetricDef,
): number {
  if (def.type === 'count') return rows.length
  if (!def.field) return 0
  const nums = rows
    .map((row) => readNumber(row, def.field!))
    .filter((n): n is number => n != null)
  if (!nums.length) return 0
  if (def.type === 'sum') return nums.reduce((a, b) => a + b, 0)
  if (def.type === 'avg') return nums.reduce((a, b) => a + b, 0) / nums.length
  if (def.type === 'min') return Math.min(...nums)
  if (def.type === 'max') return Math.max(...nums)
  return 0
}

export interface UseRsTableAnalyticsOptions {
  viewRows: Ref<Record<string, unknown>[]> | ComputedRef<Record<string, unknown>[]>
  selectedRows?: Ref<Record<string, unknown>[]> | ComputedRef<Record<string, unknown>[]>
  metrics?: RsTableAnalyticsMetricDef[]
  /**
   * 图表系列定义（bar/pie/line 意图）。
   * 后期接 ECharts：option.series[0].data = chartSeries[i].points
   */
  chartSeriesDefs?:
    | Ref<RsTableChartSeriesDef[]>
    | ComputedRef<RsTableChartSeriesDef[]>
    | RsTableChartSeriesDef[]
  mode?: Ref<'client' | 'server'> | ComputedRef<'client' | 'server'> | 'client' | 'server'
}

/**
 * 表格分析：指标 + 选中联动图表系列。
 *
 * @returns
 * - sourceRows / sourceMode：选中优先数据源
 * - metricValues：指标卡
 * - chartSeries：库无关柱状/饼图系列
 * - setRemoteValues：server 模式回填指标
 */
export function useRsTableAnalytics(options: UseRsTableAnalyticsOptions) {
  const remoteValues = ref<Record<string, number>>({})

  const selection = useRsTableSelectionSource({
    viewRows: options.viewRows,
    selectedRows: options.selectedRows,
  })

  const chartBridge = useRsTableChartBridge({
    viewRows: options.viewRows,
    selectedRows: options.selectedRows,
    seriesDefs: options.chartSeriesDefs ?? [],
  })

  const modeRef = computed(() => {
    const m = options.mode
    if (typeof m === 'string') return m
    return m?.value ?? 'client'
  })

  const metricValues = computed<RsTableAnalyticsMetricValue[]>(() => {
    const metrics = options.metrics ?? []
    return metrics.map((def) => {
      const value =
        modeRef.value === 'server'
          ? (remoteValues.value[def.key] ?? 0)
          : aggregateMetric(selection.sourceRows.value as Record<string, unknown>[], def)
      const text = def.formatter ? def.formatter(value) : String(value)
      return { key: def.key, label: def.label, value, text }
    })
  })

  function setRemoteValues(values: Record<string, number>) {
    remoteValues.value = { ...values }
  }

  /** 按图表 id 取系列，便于单图绑定 */
  function getChartSeriesById(id: string): RsTableChartSeries | undefined {
    return chartBridge.getSeriesById(id)
  }

  return {
    sourceRows: selection.sourceRows,
    sourceMode: selection.sourceMode,
    viewRows: selection.viewRows,
    selectedRows: selection.selectedRows,
    getSnapshot: selection.getSnapshot,
    subscribe: selection.subscribe,
    metricValues,
    remoteValues,
    setRemoteValues,
    mode: modeRef,
    chartSeries: chartBridge.series,
    getChartSeriesById,
  }
}

export type { RsTableChartSeries, RsTableChartSeriesDef }
