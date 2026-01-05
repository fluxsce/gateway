/**
 * ECharts 图表组件统一导出
 * 基于 ECharts 5.x 的高性能图表组件库
 */

// 基础图表组件
export { default as BaseChart } from './BaseChart.vue'

// 时间序列图表组件
export { default as MetricLineChart } from './MetricLineChart.vue'

// 饼图组件
export { default as MetricPieChart } from './MetricPieChart.vue'

// 仪表盘组件
export { default as MetricGaugeChart } from './MetricGaugeChart.vue'

// 柱状图组件
export { default as MetricBarChart } from './MetricBarChart.vue'

// 网络流量图表组件
export { default as NetworkTrafficChart } from './NetworkTrafficChart.vue'

// 图表组件类型定义
export type { default as BaseChartProps } from './BaseChart.vue'
export type { default as MetricLineChartProps } from './MetricLineChart.vue'
export type { default as MetricPieChartProps } from './MetricPieChart.vue'
export type { default as MetricGaugeChartProps } from './MetricGaugeChart.vue'
export type { default as MetricBarChartProps } from './MetricBarChart.vue'
export type { default as NetworkTrafficChartProps } from './NetworkTrafficChart.vue'
