<template>
    <div class="metric-bar-chart">
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
    </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { ReloadOutlined } from '@vicons/antd'
import BaseChart from './BaseChart.vue'
import type { EChartsOption } from 'echarts'

interface BarDataItem {
    name: string
    value: number
    color?: string
}

interface Props {
    /** 图表标题 */
    title: string
    /** 图表数据 */
    data: BarDataItem[]
    /** 图表高度 */
    height?: number
    /** 是否加载中 */
    loading?: boolean
    /** 数据单位 */
    unit?: string
    /** 是否水平显示 */
    horizontal?: boolean
    /** 是否显示数值标签 */
    showLabel?: boolean
    /** 配色方案 */
    colorScheme?: string[]
}

const props = withDefaults(defineProps<Props>(), {
    height: 300,
    loading: false,
    unit: '',
    horizontal: false,
    showLabel: true,
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

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
    const data = props.data.map((item, index) => ({
        name: item.name,
        value: item.value,
        itemStyle: {
            color: item.color || props.colorScheme[index % props.colorScheme.length]
        }
    }))

    const names = data.map(item => item.name)
    const values = data.map(item => item.value)

    return {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow'
            },
            formatter: (params: any) => {
                const data = params[0]
                return `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 4px;">${data.name}</div>
            <div style="display: flex; align-items: center;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${data.color}; border-radius: 2px;"></span>
              <span>${props.title}: ${data.value}${props.unit}</span>
            </div>
          </div>
        `
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            top: '10%',
            containLabel: true
        },
        xAxis: {
            type: props.horizontal ? 'value' : 'category',
            data: props.horizontal ? undefined : names,
            axisLabel: {
                fontSize: 11,
                formatter: props.horizontal ? `{value}${props.unit}` : undefined
            },
            axisLine: {
                lineStyle: {
                    color: '#e0e0e0'
                }
            }
        },
        yAxis: {
            type: props.horizontal ? 'category' : 'value',
            data: props.horizontal ? names : undefined,
            axisLabel: {
                fontSize: 11,
                formatter: props.horizontal ? undefined : `{value}${props.unit}`
            },
            axisLine: {
                lineStyle: {
                    color: '#e0e0e0'
                }
            },
            splitLine: {
                lineStyle: {
                    color: '#f0f0f0'
                }
            }
        },
        series: [
            {
                name: props.title,
                type: 'bar',
                data: values,
                label: {
                    show: props.showLabel,
                    position: props.horizontal ? 'right' : 'top',
                    formatter: `{c}${props.unit}`,
                    fontSize: 11,
                    color: '#666'
                },
                itemStyle: {
                    borderRadius: props.horizontal ? [0, 4, 4, 0] : [4, 4, 0, 0]
                },
                emphasis: {
                    itemStyle: {
                        shadowBlur: 10,
                        shadowColor: 'rgba(0, 0, 0, 0.3)'
                    }
                },
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
.metric-bar-chart {
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

/* 暗色主题支持 */
.dark .metric-bar-chart {
    background: #1f1f1f;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.dark .chart-title {
    color: #e0e0e0;
}
</style>