<template>
  <div class="rs-table-analytics-panel">
    <div
      v-for="metric in metrics"
      :key="metric.key"
      class="rs-table-analytics-panel__card"
    >
      <div class="rs-table-analytics-panel__label">{{ metric.label }}</div>
      <div class="rs-table-analytics-panel__value">{{ metric.text }}</div>
    </div>
    <div
      v-for="series in chartSeries ?? []"
      :key="series.id"
      class="rs-table-analytics-panel__series"
    >
      <div class="rs-table-analytics-panel__label">
        {{ series.id }} ({{ series.kind }})
      </div>
      <div class="rs-table-analytics-panel__series-meta">
        {{ series.categories.join(' / ') }}
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
/**
 * 表外指标卡展示层：渲染 metricValues；可选展示图表系列摘要（便于联调）。
 * 真实柱状/饼图由业务用 chartSeries 接 ECharts，不在此绑死库。
 */
import type {
  RsTableAnalyticsMetricValue,
  RsTableChartSeries,
} from '@/composables/useRsTableAnalytics'

defineOptions({ name: 'RsTableAnalyticsPanel' })

defineProps<{
  /** 已计算的指标列表（通常接 composable.metricValues） */
  metrics: RsTableAnalyticsMetricValue[]
  /** 可选：库无关图表系列，用于联调预览分类/数值 */
  chartSeries?: RsTableChartSeries[]
}>()

</script>

<style scoped>
.rs-table-analytics-panel {
  display: flex;
  flex-wrap: wrap;
  gap: var(--g-space-md, 12px);
  width: 100%;
}

.rs-table-analytics-panel__card {
  min-width: 120px;
  padding: var(--g-space-sm, 8px) var(--g-space-md, 12px);
  border: 1px solid var(--g-border-primary, var(--rs-border));
  border-radius: var(--g-radius-md, 8px);
  background: var(--g-bg-secondary, var(--rs-surface));
}

.rs-table-analytics-panel__label {
  font-size: var(--g-font-size-xs, 12px);
  color: var(--g-text-secondary, var(--rs-muted));
  margin-bottom: 4px;
}

.rs-table-analytics-panel__value {
  font-size: var(--g-font-size-lg, 18px);
  font-weight: 600;
  color: var(--g-text-primary, var(--rs-text));
}

.rs-table-analytics-panel__series {
  min-width: 160px;
  padding: var(--g-space-sm, 8px) var(--g-space-md, 12px);
  border: 1px dashed var(--g-border-primary, var(--rs-border));
  border-radius: var(--g-radius-md, 8px);
}

.rs-table-analytics-panel__series-meta {
  font-size: var(--g-font-size-xs, 12px);
  color: var(--g-text-secondary, var(--rs-muted));
  word-break: break-all;
}
</style>
