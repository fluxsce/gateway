<template>
    <div class="metric-gauge-chart">
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

        <div class="gauge-info">
            <div class="current-value">
                <span class="value">{{ currentValue }}</span>
                <span class="unit">{{ unit }}</span>
            </div>
            <div class="status-indicator" :class="statusClass">
                <span class="status-dot"></span>
                <span class="status-text">{{ statusText }}</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { ReloadOutlined } from '@vicons/antd'
import BaseChart from './BaseChart.vue'
import type { EChartsOption } from 'echarts'

interface Props {
    /** 图表标题 */
    title: string
    /** 当前值 */
    value: number
    /** 最小值 */
    min?: number
    /** 最大值 */
    max?: number
    /** 数据单位 */
    unit?: string
    /** 图表高度 */
    height?: number
    /** 是否加载中 */
    loading?: boolean
    /** 警告阈值 */
    warningThreshold?: number
    /** 危险阈值 */
    dangerThreshold?: number
    /** 仪表盘类型 */
    gaugeType?: 'arc' | 'semicircle' | 'circle'
}

const props = withDefaults(defineProps<Props>(), {
    min: 0,
    max: 100,
    unit: '%',
    height: 300,
    loading: false,
    warningThreshold: 70,
    dangerThreshold: 90,
    gaugeType: 'arc'
})

const emit = defineEmits<{
    refresh: []
    chartClick: [params: any]
}>()

// 当前值（保留小数点后1位）
const currentValue = computed(() => {
    return Number(props.value).toFixed(1)
})

// 状态计算
const status = computed(() => {
    if (props.value >= props.dangerThreshold) return 'danger'
    if (props.value >= props.warningThreshold) return 'warning'
    return 'normal'
})

const statusClass = computed(() => `status-${status.value}`)

const statusText = computed(() => {
    switch (status.value) {
        case 'danger':
            return '危险'
        case 'warning':
            return '警告'
        default:
            return '正常'
    }
})

// 颜色计算
const gaugeColor = computed(() => {
    const colorStops = [
        { offset: 0, color: '#52c41a' },
        { offset: props.warningThreshold / props.max, color: '#faad14' },
        { offset: props.dangerThreshold / props.max, color: '#ff4d4f' },
        { offset: 1, color: '#ff4d4f' }
    ]

    return {
        type: 'linear' as const,
        x: 0,
        y: 0,
        x2: 1,
        y2: 0,
        colorStops
    }
})

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
    const startAngle = props.gaugeType === 'semicircle' ? 180 : 225
    const endAngle = props.gaugeType === 'semicircle' ? 0 : -45

    return {
        tooltip: {
            formatter: `${props.title}: ${currentValue.value}${props.unit}`
        },
        series: [
            {
                type: 'gauge',
                startAngle,
                endAngle,
                radius: '80%',
                center: ['50%', props.gaugeType === 'semicircle' ? '70%' : '50%'],
                min: props.min,
                max: props.max,
                splitNumber: 10,
                progress: {
                    show: true,
                    width: 18,
                    itemStyle: {
                        color: gaugeColor.value
                    }
                },
                pointer: {
                    show: true,
                    length: '70%',
                    width: 8,
                    itemStyle: {
                        color: 'auto'
                    }
                },
                axisLine: {
                    lineStyle: {
                        width: 18,
                        color: [[1, '#e6f7ff']]
                    }
                },
                axisTick: {
                    show: true,
                    splitNumber: 2,
                    lineStyle: {
                        width: 2,
                        color: '#999'
                    }
                },
                splitLine: {
                    show: true,
                    length: 15,
                    lineStyle: {
                        width: 3,
                        color: '#999'
                    }
                },
                axisLabel: {
                    show: true,
                    distance: 25,
                    fontSize: 12,
                    color: '#999',
                    formatter: (value: number) => {
                        return value.toString()
                    }
                },
                title: {
                    show: false
                },
                detail: {
                    show: true,
                    valueAnimation: true,
                    width: '60%',
                    lineHeight: 40,
                    borderRadius: 8,
                    offsetCenter: [0, props.gaugeType === 'semicircle' ? '-20%' : '40%'],
                    fontSize: 24,
                    fontWeight: 'bold',
                    formatter: `{value}${props.unit}`,
                    color: 'inherit'
                },
                data: [
                    {
                        value: props.value,
                        name: props.title
                    }
                ]
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
.metric-gauge-chart {
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

.gauge-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f0f0f0;
}

.current-value {
    display: flex;
    align-items: baseline;
    gap: 4px;
}

.value {
    font-size: 24px;
    font-weight: bold;
    color: #262626;
}

.unit {
    font-size: 14px;
    color: #8c8c8c;
}

.status-indicator {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    font-weight: 500;
}

.status-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
}

.status-normal .status-dot {
    background-color: #52c41a;
}

.status-warning .status-dot {
    background-color: #faad14;
}

.status-danger .status-dot {
    background-color: #ff4d4f;
}

.status-normal .status-text {
    color: #52c41a;
}

.status-warning .status-text {
    color: #faad14;
}

.status-danger .status-text {
    color: #ff4d4f;
}

/* 暗色主题支持 */
.dark .metric-gauge-chart {
    background: #1f1f1f;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.dark .chart-title {
    color: #e0e0e0;
}

.dark .gauge-info {
    border-top-color: #333;
}

.dark .value {
    color: #e0e0e0;
}

.dark .unit {
    color: #8c8c8c;
}
</style>