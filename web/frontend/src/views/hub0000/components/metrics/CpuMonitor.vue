<template>
    <n-card :title="t('hub0000.cpu.title')" :bordered="false" class="monitor-card">
        <template #header-extra>
            <div class="card-extra">
                <n-date-picker v-model:value="dateTimeRange" type="datetimerange" :shortcuts="timeRangeShortcuts"
                    :placeholder="t('hub0000.common.selectTimeRange')" @update:value="handleTimeRangeChange"
                    size="small" />
                <n-button size="small" @click="refreshData" :loading="loading">
                    <template #icon>
                        <n-icon>
                            <ReloadOutlined />
                        </n-icon>
                    </template>
                    {{ t('hub0000.common.refresh') }}
                </n-button>
            </div>
        </template>

        <div class="chart-container">
            <div ref="chartRef" class="chart-element"></div>

            <div v-if="loading" class="chart-loading">
                <n-spin size="large" />
            </div>

            <div v-if="!loading && (!data || !Array.isArray(data) || data.length === 0)" class="chart-empty">
                <n-empty :description="t('hub0000.common.noData')" />
            </div>
        </div>
    </n-card>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { NCard, NButton, NIcon, NSpin, NEmpty, NDatePicker } from 'naive-ui'
import { ReloadOutlined } from '@vicons/antd'
import * as echarts from 'echarts/core'
import { LineChart } from 'echarts/charts'
import {
    TitleComponent,
    TooltipComponent,
    GridComponent,
    LegendComponent,
    ToolboxComponent,
} from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { CPUMetrics } from '../../types'
import { formatDate } from '@/utils/format'

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

const { t } = useModuleI18n('hub0000')

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

// 时间范围快捷选项 - 改为计算属性
const timeRangeShortcuts = computed(() => {
    return {
        [t('hub0000.timeRangeShortcuts.lastHour')]: (): [number, number] => {
            const end = Date.now()
            const start = end - 3600 * 1000
            return [start, end]
        },
        [t('hub0000.timeRangeShortcuts.last6Hours')]: (): [number, number] => {
            const end = Date.now()
            const start = end - 6 * 3600 * 1000
            return [start, end]
        },
        [t('hub0000.timeRangeShortcuts.last12Hours')]: (): [number, number] => {
            const end = Date.now()
            const start = end - 12 * 3600 * 1000
            return [start, end]
        },
        [t('hub0000.timeRangeShortcuts.last24Hours')]: (): [number, number] => {
            const end = Date.now()
            const start = end - 24 * 3600 * 1000
            return [start, end]
        },
        [t('hub0000.timeRangeShortcuts.last7Days')]: (): [number, number] => {
            const end = Date.now()
            const start = end - 7 * 24 * 3600 * 1000
            return [start, end]
        }
    } as Record<string, () => [number, number]>
})

// 表格分页

// CPU表格列定义

// CPU表格数据

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
            name: t('hub0000.cpu.usage'),
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
            name: t('hub0000.cpu.userUsage'),
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
            name: t('hub0000.cpu.systemUsage'),
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
            formatter: (params: any) => {
                const detail = params[0].data.detail

                let result = ``

                if (detail) {
                    result += `${t('hub0000.cpu.detailTitle')}:<br/>`
                    result += `${t('hub0000.cpu.serverId')}: ${detail.metricServerId}<br/>`
                    result += `${t('hub0000.cpu.coreCount')}: ${detail.coreCount}<br/>`
                    result += `${t('hub0000.cpu.logicalCount')}: ${detail.logicalCount}<br/>`
                    result += `${t('hub0000.cpu.loadAvg1')}: ${detail.loadAvg1.toFixed(2)}<br/>`
                    result += `${t('hub0000.cpu.loadAvg5')}: ${detail.loadAvg5.toFixed(2)}<br/>`
                    result += `${t('hub0000.cpu.loadAvg15')}: ${detail.loadAvg15.toFixed(2)}<br/>`
                    result += `${t('hub0000.cpu.ioWaitUsage')}: ${detail.ioWaitPercent.toFixed(2)}%<br/>`
                    result += `${t('hub0000.cpu.irqUsage')}: ${detail.irqPercent.toFixed(2)}%<br/>`
                    result += `${t('hub0000.cpu.softIrqUsage')}: ${detail.softIrqPercent.toFixed(2)}%<br/>`
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
                t('hub0000.cpu.usage'),
                t('hub0000.cpu.userUsage'),
                t('hub0000.cpu.systemUsage')
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
.monitor-card {
    height: 100%;
}

.chart-container {
    position: relative;
    height: 300px;
    width: 100%;
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
