/**
 * RsTable 表外分析壳：指标面板 + 编排 composable 再导出。
 *
 * 图表：业务用 useRsTableAnalytics().chartSeries 接 ECharts；
 * 本包不内置图表渲染，避免绑死可视化库。
 */

export { default as RsTableAnalyticsPanel } from './RsTableAnalyticsPanel.vue'
export {
  useRsTableAnalytics,
  type RsTableAnalyticsAggType,
  type RsTableAnalyticsMetricDef,
  type RsTableAnalyticsMetricValue,
  type UseRsTableAnalyticsOptions,
  type RsTableChartSeries,
  type RsTableChartSeriesDef,
} from '@/composables/useRsTableAnalytics'
