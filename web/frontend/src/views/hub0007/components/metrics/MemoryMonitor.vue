<template>
    <GCard :title="displayTitle" :show-title="true">
        <template #header-extra>
            <div class="card-extra">
                <n-date-picker v-model:value="dateTimeRange" type="datetimerange" :shortcuts="timeRangeShortcuts"
                    placeholder="选择时间范围" @update:value="handleTimeRangeChange" size="small" />
                <n-button size="small" @click="refreshData" :loading="loading">
                    <template #icon>
                        <n-icon>
                            <ReloadOutlined />
                        </n-icon>
                    </template>
                    刷新
                </n-button>
            </div>
        </template>

        <div class="chart-container">
            <div ref="chartRef" class="chart-element"></div>

            <div v-if="loading" class="chart-loading">
                <n-spin size="large" />
            </div>

            <div v-if="!loading && !chartData.length" class="chart-empty">
                <n-empty description="暂无数据" />
            </div>
        </div>
    </GCard>
</template>

<script setup lang="ts">
import { GCard } from '@/components/gcard'
import { formatDate } from '@/utils/format'
import { ReloadOutlined } from '@vicons/antd'
import { LineChart } from 'echarts/charts'
import {
    GridComponent,
    LegendComponent,
    TitleComponent,
    ToolboxComponent,
    TooltipComponent,
} from 'echarts/components'
import * as echarts from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { NButton, NDatePicker, NEmpty, NIcon, NSpin } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { MemoryMetrics } from '../../types'

// 注册必要的 ECharts 组件
echarts.use([
    LineChart,
    TitleComponent,
    TooltipComponent,
    GridComponent,
    LegendComponent,
    ToolboxComponent,
    CanvasRenderer,
])

// 组件属性定义
const props = defineProps({
    title: {
        type: String,
        default: '内存使用率'
    },
    data: {
        type: Array as () => MemoryMetrics[],
        default: () => []
    },
    loading: {
        type: Boolean,
        default: false
    },
    warningThreshold: {
        type: Number,
        default: 80
    },
    dangerThreshold: {
        type: Number,
        default: 90
    },
    memoryDetailData: {
        type: Array as () => MemoryMetrics[],
        default: () => []
    }
})

// 计算属性 - 标题
const displayTitle = computed(() => props.title || '内存使用率')
const warningThreshold = computed(() => props.warningThreshold)
const dangerThreshold = computed(() => props.dangerThreshold)

// 事件定义
const emit = defineEmits<{
    (e: 'refresh'): void
    (e: 'time-range-change', range: [number, number] | null): void
}>()

// 图表引用
const chartRef = ref<HTMLElement | null>(null)
let chart: echarts.ECharts | null = null

// 时间范围选择
const end = Date.now()
const start = end - 3600 * 1000 // 最近1小时
const dateTimeRange = ref<[number, number] | null>([start, end])

// 时间范围快捷选项
const timeRangeShortcuts: Record<string, () => [number, number]> = {
    '最近1小时': (): [number, number] => {
        const end = Date.now()
        const start = end - 3600 * 1000
        return [start, end]
    },
    '最近6小时': (): [number, number] => {
        const end = Date.now()
        const start = end - 6 * 3600 * 1000
        return [start, end]
    },
    '最近12小时': (): [number, number] => {
        const end = Date.now()
        const start = end - 12 * 3600 * 1000
        return [start, end]
    },
    '最近24小时': (): [number, number] => {
        const end = Date.now()
        const start = end - 24 * 3600 * 1000
        return [start, end]
    },
    '最近7天': (): [number, number] => {
        const end = Date.now()
        const start = end - 7 * 24 * 3600 * 1000
        return [start, end]
    }
}

// 计算属性 - 图表数据
const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return []

    // 按时间分组并计算平均值
    const timeGroups = new Map<string, number[]>()

    props.data.forEach(metric => {
        const time = metric.collectTime
        if (!timeGroups.has(time)) {
            timeGroups.set(time, [])
        }
        timeGroups.get(time)!.push(metric.usagePercent)
    })

    // 后端数据是倒序，需要转为正序
    const result = Array.from(timeGroups.entries())
        .map(([time, values]) => {
            // 找到该时间点的完整内存指标数据
            const metric = props.data.find(m => m.collectTime === time)

            return {
                time,
                value: Math.round(values.reduce((sum, val) => sum + val, 0) / values.length * 100) / 100,
                detail: metric ? {
                    totalMemory: metric.totalMemory,
                    usedMemory: metric.usedMemory,
                    availableMemory: metric.availableMemory,
                    freeMemory: metric.freeMemory,
                    cachedMemory: metric.cachedMemory,
                    buffersMemory: metric.buffersMemory,
                    swapTotal: metric.swapTotal,
                    swapUsed: metric.swapUsed,
                    swapFree: metric.swapFree,
                    swapUsagePercent: metric.swapUsagePercent
                } : null
            }
        })
        .sort((a, b) => new Date(a.time).getTime() - new Date(b.time).getTime()) // 按时间正序排序

    return result
})

// 格式化内存大小
const formatMemorySize = (bytes: number): string => {
    if (bytes === 0) return '0 B'

    const units = ['B', 'KB', 'MB', 'GB', 'TB', 'PB']
    const i = Math.floor(Math.log(bytes) / Math.log(1024))

    return `${(bytes / Math.pow(1024, i)).toFixed(2)} ${units[i]}`
}

// 刷新数据
const refreshData = () => {
    emit('refresh')
}

// 时间范围变化处理
const handleTimeRangeChange = (value: [number, number] | null) => {
    emit('time-range-change', value)
}

// 初始化图表
const initChart = () => {
    if (!chartRef.value) return

    chart = echarts.init(chartRef.value)

    // 监听窗口大小变化
    window.addEventListener('resize', handleResize)

    // 首次更新图表
    updateChart()
}

// 更新图表
const updateChart = () => {
    if (!chart) return

    const data = chartData.value
    if (data.length === 0) {
        // 清空图表以确保不显示旧数据
        chart.clear()
        return
    }

    const option = {
        title: {
            show: false
        },
        tooltip: {
            trigger: 'axis',
            confine: false,
            appendToBody: true,
            formatter: function (params: any) {
                const dataPoint = params[0]
                const detail = dataPoint.data.detail

                if (!detail) return `${dataPoint.marker}${dataPoint.seriesName}: ${dataPoint.data.value}%`

                let result = ``
                result += `${dataPoint.marker}内存使用率: <b>${dataPoint.data.value}%</b><br/>`
                result += `<br/>详细信息:<br/>`
                result += `总内存: ${formatMemorySize(detail.totalMemory)}<br/>`
                result += `已用内存: ${formatMemorySize(detail.usedMemory)}<br/>`
                result += `可用内存: ${formatMemorySize(detail.availableMemory)}<br/>`
                result += `空闲内存: ${formatMemorySize(detail.freeMemory)}<br/>`
                result += `缓存内存: ${formatMemorySize(detail.cachedMemory)}<br/>`
                result += `缓冲区内存: ${formatMemorySize(detail.buffersMemory)}<br/>`
                result += `<br/>交换分区:<br/>`
                result += `总交换空间: ${formatMemorySize(detail.swapTotal)}<br/>`
                result += `已用交换空间: ${formatMemorySize(detail.swapUsed)}<br/>`
                result += `空闲交换空间: ${formatMemorySize(detail.swapFree)}<br/>`
                result += `交换使用率: ${detail.swapUsagePercent.toFixed(1)}%`

                return result
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            data: data.map(item => formatDate(new Date(item.time), 'HH:mm:ss')),
            axisLabel: {
                rotate: 45
            }
        },
        yAxis: {
            type: 'value',
            min: 0,
            max: 100,
            splitLine: {
                show: true,
                lineStyle: {
                    type: 'dashed'
                }
            },
            axisLabel: {
                formatter: '{value}%'
            }
        },
        series: [
            {
                name: '内存使用率',
                type: 'line',
                data: data.map(item => ({
                    value: item.value,
                    detail: item.detail
                })),
                smooth: true,
                showSymbol: true,
                symbolSize: 6,
                lineStyle: {
                    width: 2,
                    color: '#52c41a'
                },
                itemStyle: {
                    color: function (params: any) {
                        const value = params.data.value
                        if (value >= dangerThreshold.value) {
                            return '#ff4d4f'
                        } else if (value >= warningThreshold.value) {
                            return '#faad14'
                        }
                        return '#52c41a'
                    }
                },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(82, 196, 26, 0.3)'
                        },
                        {
                            offset: 1,
                            color: 'rgba(82, 196, 26, 0.1)'
                        }
                    ])
                },
                markLine: {
                    silent: true,
                    lineStyle: {
                        color: '#faad14'
                    },
                    data: [
                        {
                            yAxis: warningThreshold.value,
                            label: {
                                formatter: `警告 ${warningThreshold.value}%`,
                                position: 'start'
                            }
                        },
                        {
                            yAxis: dangerThreshold.value,
                            lineStyle: {
                                color: '#ff4d4f'
                            },
                            label: {
                                formatter: `危险 ${dangerThreshold.value}%`,
                                position: 'start'
                            }
                        }
                    ]
                }
            }
        ]
    }

    chart.setOption(option)
}

// 处理窗口大小变化
const handleResize = () => {
    chart?.resize()
}

// 监听数据变化，更新图表
watch(() => props.data, () => {
    updateChart()
}, { deep: true })

// 生命周期钩子
onMounted(() => {
    initChart()
    // 初始化时触发时间范围变化事件
    emit('time-range-change', dateTimeRange.value)
})

onUnmounted(() => {
    if (chart) {
        chart.dispose()
        chart = null
    }
    window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.chart-container {
    position: relative;
    height: 300px;
    width: 100%;
    overflow: visible;
}

.chart-element {
    height: 100%;
    width: 100%;
}

.chart-loading,
.chart-empty {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    display: flex;
    align-items: center;
    justify-content: center;
}

.card-extra {
    display: flex;
    gap: 8px;
    align-items: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
    .chart-container {
        height: 250px;
    }
}
</style>

