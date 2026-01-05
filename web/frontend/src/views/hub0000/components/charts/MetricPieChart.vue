<template>
    <div class="metric-pie-chart">
        <div class="chart-header">
            <h3 class="chart-title">{{ title }}</h3>
            <div class="chart-controls">
                <n-button size="small" @click="handleRefresh" :loading="loading">
                    <template #icon>
                        <n-icon>
                            <ReloadOutlined />
                        </n-icon>
                    </template>
                </n-button>
            </div>
        </div>

        <div class="chart-content">
            <BaseChart :options="chartOptions" :height="height" @chart-ready="handleChartReady"
                @chart-click="handleChartClick" />
        </div>

        <div class="chart-legend" v-if="showLegend">
            <div class="legend-item" v-for="item in legendData" :key="item.name">
                <span class="legend-color" :style="{ backgroundColor: item.color }"></span>
                <span class="legend-name">{{ item.name }}</span>
                <span class="legend-value">{{ item.value }}</span>
                <span class="legend-percent">{{ item.percent }}%</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { ReloadOutlined } from '@vicons/antd'
import BaseChart from './BaseChart.vue'
import type { EChartsOption } from 'echarts'

interface PieDataItem {
    name: string
    value: number
    color?: string
}

interface Props {
    /** 图表标题 */
    title: string
    /** 图表数据 */
    data: PieDataItem[]
    /** 图表高度 */
    height?: number
    /** 是否加载中 */
    loading?: boolean
    /** 是否显示图例 */
    showLegend?: boolean
    /** 是否显示百分比 */
    showPercent?: boolean
    /** 是否显示环形图 */
    isDoughnut?: boolean
    /** 环形图内径比例 */
    innerRadius?: string
    /** 配色方案 */
    colorScheme?: string[]
}

const props = withDefaults(defineProps<Props>(), {
    height: 300,
    loading: false,
    showLegend: true,
    showPercent: true,
    isDoughnut: false,
    innerRadius: '40%',
    colorScheme: () => [
        '#1890ff',
        '#52c41a',
        '#faad14',
        '#ff4d4f',
        '#722ed1',
        '#fa8c16',
        '#13c2c2',
        '#eb2f96'
    ]
})

const emit = defineEmits<{
    refresh: []
    chartClick: [params: any]
}>()

// 计算图例数据
const legendData = computed(() => {
    const total = props.data.reduce((sum, item) => sum + item.value, 0)

    return props.data.map((item, index) => ({
        name: item.name,
        value: item.value,
        percent: total > 0 ? Math.round((item.value / total) * 100) : 0,
        color: item.color || props.colorScheme[index % props.colorScheme.length]
    }))
})

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
    const data = props.data.map((item, index) => ({
        name: item.name,
        value: item.value,
        itemStyle: {
            color: item.color || props.colorScheme[index % props.colorScheme.length]
        }
    }))

    return {
        tooltip: {
            trigger: 'item',
            formatter: (params: any) => {
                const percent = params.percent.toFixed(1)
                return `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 4px;">${params.name}</div>
            <div style="display: flex; align-items: center;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${params.color}; border-radius: 50%;"></span>
              <span>数量: ${params.value}</span>
            </div>
            <div style="margin-top: 2px;">
              <span>占比: ${percent}%</span>
            </div>
          </div>
        `
            }
        },
        legend: {
            show: false // 使用自定义图例
        },
        series: [
            {
                name: props.title,
                type: 'pie',
                radius: props.isDoughnut ? ['40%', '70%'] : '70%',
                center: ['50%', '50%'],
                data,
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowOffsetX: 0,
                        shadowColor: 'rgba(0, 0, 0, 0.5)'
                    }
                },
                label: {
                    show: props.showPercent,
                    position: 'outside',
                    formatter: '{d}%',
                    fontSize: 12,
                    fontWeight: 'bold'
                },
                labelLine: {
                    show: props.showPercent,
                    length: 15,
                    length2: 10,
                    smooth: true
                },
                animationType: 'scale',
                animationEasing: 'elasticOut',
                animationDelay: (idx: number) => idx * 100
            }
        ]
    }
})

// 处理刷新
const handleRefresh = () => {
    emit('refresh')
}

// 处理图表就绪
const handleChartReady = () => {
    // 可以在这里处理图表就绪后的逻辑
}

// 处理图表点击
const handleChartClick = (params: any) => {
    emit('chartClick', params)
}
</script>

<style scoped>
.metric-pie-chart {
    background: #fff;
    border-radius: 8px;
    padding: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
}

.chart-title {
    margin: 0;
    font-size: 16px;
    font-weight: 600;
    color: #262626;
}

.chart-controls {
    display: flex;
    gap: 8px;
    align-items: center;
}

.chart-content {
    min-height: 300px;
}

.chart-legend {
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f0f0f0;
}

.legend-item {
    display: flex;
    align-items: center;
    padding: 4px 0;
    font-size: 12px;
}

.legend-color {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    margin-right: 8px;
    flex-shrink: 0;
}

.legend-name {
    flex: 1;
    margin-right: 8px;
    color: #262626;
}

.legend-value {
    margin-right: 8px;
    color: #595959;
    font-weight: 500;
}

.legend-percent {
    color: #8c8c8c;
    font-size: 11px;
    min-width: 35px;
    text-align: right;
}

/* 暗色主题支持 */
.dark .metric-pie-chart {
    background: #1f1f1f;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.dark .chart-title {
    color: #e0e0e0;
}

.dark .chart-legend {
    border-top-color: #333;
}

.dark .legend-name {
    color: #e0e0e0;
}

.dark .legend-value {
    color: #bfbfbf;
}

.dark .legend-percent {
    color: #8c8c8c;
}
</style>