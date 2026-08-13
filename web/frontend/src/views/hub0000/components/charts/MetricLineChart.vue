<template>
    <div class="metric-line-chart">
        <div class="chart-header">
            <h3 class="chart-title">{{ title }}</h3>
            <div class="chart-controls">
                <MetricsDateTimeRange v-model="timeRange" @change="handleTimeRangeChange" />
                <RsButton size="sm" :loading="loading" @click="handleRefresh">
                    <template #icon>
                        <GIcon icon="ReloadOutline" />
                    </template>
                </RsButton>
            </div>
        </div>

        <div class="chart-content">
            <BaseChart :options="chartOptions" :height="height" @chart-ready="handleChartReady"
                @chart-click="handleChartClick" />
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { GIcon } from '@/components/gicon'
import { RsButton } from '@/ui'
import MetricsDateTimeRange from '../metrics/MetricsDateTimeRange.vue'
import BaseChart from './BaseChart.vue'
import type { EChartsOption } from 'echarts'
import type { MetricTrend } from '../../types'

interface Props {
    /** 图表标题 */
    title: string
    /** 图表数据 */
    data: MetricTrend[]
    /** 图表高度 */
    height?: number
    /** 是否加载中 */
    loading?: boolean
    /** 数据单位 */
    unit?: string
    /** 是否显示面积 */
    showArea?: boolean
    /** 线条颜色 */
    color?: string
    /** 警告阈值 */
    warningThreshold?: number
    /** 危险阈值 */
    dangerThreshold?: number
    /** CPU详细数据（用于tooltip显示） */
    cpuDetailData?: any[]
}

const props = withDefaults(defineProps<Props>(), {
    height: 300,
    loading: false,
    unit: '%',
    showArea: true,
    color: '#1890ff',
    warningThreshold: 80,
    dangerThreshold: 90
})

const emit = defineEmits<{
    refresh: []
    chartClick: [params: any]
    timeRangeChange: [timeRange: [number, number] | null]
}>()

// 时间范围状态（默认最近 1 小时，与 MetricsDateTimeRange 快捷项对齐）
const timeRange = ref<[number, number] | null>(null)

// 初始化时间范围（默认最近 1 小时，本地时区）
const initTimeRange = () => {
    const now = Date.now()
    timeRange.value = [now - 3600 * 1000, now]
}

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
    const times = props.data.map(item => item.time)
    const values = props.data.map(item => item.value)

    return {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross'
            },
            formatter: (params: any) => {
                const data = params[0]
                const timePoint = data.name

                let tooltipContent = `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 4px;">${timePoint}</div>
            <div style="display: flex; align-items: center; margin-bottom: 4px;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${data.color}; border-radius: 50%;"></span>
              <span>${props.title}: ${data.value}${props.unit}</span>
            </div>
        `

                // 如果有CPU详细数据，查找对应时间点的CPU信息
                if (props.cpuDetailData && props.cpuDetailData.length > 0) {
                    const cpuData = props.cpuDetailData.find(item => item.collectTime === timePoint)
                    if (cpuData) {
                        tooltipContent += `
            <div style="border-top: 1px solid #e0e0e0; padding-top: 4px; margin-top: 4px;">
              <div style="font-size: 11px; color: #666; margin-bottom: 2px;">
                <span style="font-weight: 500;">CPU核心:</span> ${cpuData.coreCount}核/${cpuData.logicalCount}线程
              </div>
              <div style="font-size: 11px; color: #666;">
                <span style="font-weight: 500;">系统负载:</span> ${cpuData.loadAvg1} / ${cpuData.loadAvg5} / ${cpuData.loadAvg15}
              </div>
            </div>
          `
                    }
                }

                tooltipContent += `</div>`
                return tooltipContent
            }
        },
        legend: {
            data: [props.title],
            top: 5,
            textStyle: {
                fontSize: 12
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '10%',
            top: '15%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: times,
            axisLabel: {
                fontSize: 11,
                formatter: (value: string) => {
                    const date = new Date(value)
                    return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
                }
            },
            axisLine: {
                lineStyle: {
                    color: '#e0e0e0'
                }
            }
        },
        yAxis: {
            type: 'value',
            axisLabel: {
                fontSize: 11,
                formatter: `{value}${props.unit}`
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
                type: 'line',
                data: values,
                smooth: true,
                symbol: 'circle',
                symbolSize: 4,
                lineStyle: {
                    color: props.color,
                    width: 2
                },
                itemStyle: {
                    color: props.color
                },
                areaStyle: props.showArea ? {
                    color: {
                        type: 'linear',
                        x: 0,
                        y: 0,
                        x2: 0,
                        y2: 1,
                        colorStops: [
                            {
                                offset: 0,
                                color: props.color + '40'
                            },
                            {
                                offset: 1,
                                color: props.color + '10'
                            }
                        ]
                    }
                } : undefined,
                markLine: {
                    silent: true,
                    data: [
                        ...(props.warningThreshold ? [{
                            yAxis: props.warningThreshold,
                            lineStyle: {
                                color: '#faad14',
                                type: 'dashed' as const
                            },
                            label: {
                                formatter: `警告: ${props.warningThreshold}${props.unit}`
                            }
                        }] : []),
                        ...(props.dangerThreshold ? [{
                            yAxis: props.dangerThreshold,
                            lineStyle: {
                                color: '#ff4d4f',
                                type: 'dashed' as const
                            },
                            label: {
                                formatter: `危险: ${props.dangerThreshold}${props.unit}`
                            }
                        }] : [])
                    ]
                }
            }
        ]
    }
})

// 处理时间范围变化
const handleTimeRangeChange = (value: [number, number] | null) => {
    emit('timeRangeChange', value)
}

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

// 组件挂载时初始化时间范围
onMounted(() => {
    initTimeRange()
    // 触发初始时间范围变化
    emit('timeRangeChange', timeRange.value)
})
</script>

<style scoped>
.metric-line-chart {
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
    gap: 12px;
    align-items: center;
}

.chart-content {
    min-height: 300px;
}

/* 暗色主题支持 */
.dark .metric-line-chart {
    background: #1f1f1f;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.dark .chart-title {
    color: #e0e0e0;
}
</style>