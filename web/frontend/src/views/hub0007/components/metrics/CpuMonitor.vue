<template>
    <GCard title="CPU使用率" :show-title="true">
        <template #header-extra>
            <div class="card-extra">
                <n-date-picker v-model:value="dateTimeRange" type="datetimerange" :shortcuts="timeRangeShortcuts"
                    placeholder="选择时间范围" @update:value="handleTimeRangeChange"
                    size="small" />
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

            <div v-if="!loading && (!data || !Array.isArray(data) || data.length === 0)" class="chart-empty">
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
import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { CPUMetrics } from '../../types'

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

// 图表实例
const chartRef = ref<HTMLElement | null>(null)
let chart: echarts.ECharts | null = null

// 组件属性定义
const props = defineProps({
    data: {
        type: Array as () => CPUMetrics[],
        default: () => []
    },
    loading: {
        type: Boolean,
        default: false
    },
    cpuDetailData: {
        type: Array as () => CPUMetrics[],
        default: () => []
    },
    userColor: {
        type: String,
        default: '#1890ff'
    },
    systemColor: {
        type: String,
        default: '#52c41a'
    },
    idleColor: {
        type: String,
        default: '#d9d9d9'
    }
})

// 组件事件
const emit = defineEmits(['refresh', 'time-range-change'])

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

// 初始化图表
const initChart = () => {
    if (!chartRef.value) return

    chart = echarts.init(chartRef.value)
    window.addEventListener('resize', handleResize)
    updateChart()
}

// 更新图表
const updateChart = () => {
    if (!chart) return

    if (!props.data || !Array.isArray(props.data) || props.data.length === 0) {
        // 清空图表以确保不显示旧数据
        chart.clear()
        return
    }

    // 转换为图表数据
    const times: string[] = props.data.map(item => item.collectTime).sort()
    const timeDataMap = new Map<string, CPUMetrics>()
    props.data.forEach(item => {
        timeDataMap.set(item.collectTime, item)
    })

    // 准备数据系列
    const series = [
        {
            name: 'CPU使用率',
            type: 'line',
            data: times.map(time => {
                const data = timeDataMap.get(time)
                return {
                    value: data ? data.usagePercent : 0,
                    detail: data // 存储完整的CPU指标数据
                }
            }),
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: 2
            },
            itemStyle: {
                color: props.userColor
            }
        },
        {
            name: '用户态使用率',
            type: 'line',
            data: times.map(time => {
                const data = timeDataMap.get(time)
                return {
                    value: data ? data.userPercent : 0,
                    detail: data
                }
            }),
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: 2
            },
            itemStyle: {
                color: props.systemColor
            }
        },
        {
            name: '系统态使用率',
            type: 'line',
            data: times.map(time => {
                const data = timeDataMap.get(time)
                return {
                    value: data ? data.systemPercent : 0,
                    detail: data
                }
            }),
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: 2
            },
            itemStyle: {
                color: props.idleColor
            }
        }
    ]

    const option = {
        tooltip: {
            trigger: 'axis',
            confine: false,
            appendToBody: true,
            formatter: (params: any) => {
                const detail = params[0].data.detail

                let result = ``

                if (detail) {
                    result += `CPU详细信息:<br/>`
                    result += `物理核心数: ${detail.coreCount}<br/>`
                    result += `逻辑核心数: ${detail.logicalCount}<br/>`
                    result += `1分钟负载: ${detail.loadAvg1.toFixed(2)}<br/>`
                    result += `5分钟负载: ${detail.loadAvg5.toFixed(2)}<br/>`
                    result += `15分钟负载: ${detail.loadAvg15.toFixed(2)}<br/>`
                    result += `IO等待率: ${detail.ioWaitPercent.toFixed(2)}%<br/>`
                    result += `硬中断率: ${detail.irqPercent.toFixed(2)}%<br/>`
                    result += `软中断率: ${detail.softIrqPercent.toFixed(2)}%<br/>`
                    result += '<br/>'
                }

                params.forEach((item: any) => {
                    const marker = item.marker
                    const seriesName = item.seriesName
                    const value = `${item.value.toFixed(2)}%`
                    result += `${marker} ${seriesName}: ${value}<br/>`
                })

                return result
            },
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985'
                }
            }
        },
        legend: {
            data: [
                'CPU使用率',
                '用户态使用率',
                '系统态使用率'
            ],
            type: 'scroll',
            bottom: 0
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '15%',
            top: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: times.map(time => formatDate(time, 'HH:mm:ss')),
            axisLabel: {
                interval: Math.floor(times.length / 10)
            }
        },
        yAxis: {
            type: 'value',
            axisLabel: {
                formatter: (value: number) => value.toFixed(2) + '%'
            },
            max: 100,
            min: 0
        },
        series
    }

    chart.setOption(option, true)
}

// 事件处理
const handleResize = () => {
    chart?.resize()
}

const refreshData = () => {
    emit('refresh')
}

const handleTimeRangeChange = (value: [number, number] | null) => {
    emit('time-range-change', value)
}

// 监听数据变化
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

.time-range-selector {
    margin: 8px 0 16px;
    display: flex;
    justify-content: flex-end;
}

.detail-section {
    margin-top: 16px;
}

/* 响应式设计 */
@media (max-width: 768px) {
    .chart-container {
        height: 250px;
    }
}
</style>

