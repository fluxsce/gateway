<template>
    <RsCard :title="t('process.title')" variant="outlined" :padding="false" class="monitor-card">
        <template #actions>
            <div class="card-extra">
                <MetricsDateTimeRange v-model="dateTimeRange" @change="handleTimeRangeChange" />
                <RsButton size="sm" :loading="loading" @click="refreshData">
                    <template #icon>
                        <GIcon icon="ReloadOutline" />
                    </template>
                    {{ t('common.refresh') }}
                </RsButton>
            </div>
        </template>

        <div class="chart-container">
            <div ref="chartRef" class="chart-element"></div>

            <div v-if="loading" class="chart-loading">
                <RsLoading size="lg" />
            </div>

            <div v-if="!loading && !data?.length" class="chart-empty">
                <RsEmpty :description="t('common.noData')" />
            </div>
        </div>
    </RsCard>
</template>

<script setup lang="ts">
import { GIcon } from '@/components/gicon'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate } from '@/utils/format'
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
import { RsButton, RsCard, RsEmpty, RsLoading } from '@/ui'
import { createAxisTooltipOptions } from '@/views/hub0000/components/metrics/echartsTooltip'
import MetricsDateTimeRange from '@/views/hub0000/components/metrics/MetricsDateTimeRange.vue'
import { onMounted, onUnmounted, ref, watch } from 'vue'
import type { ProcessInfo } from '../../types'

const { t } = useModuleI18n('hub0007')

function formatPercent(value: unknown, digits = 2): string {
    const n = Number(value)
    return Number.isFinite(n) ? n.toFixed(digits) : '-'
}

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
        type: Array as () => ProcessInfo[],
        default: () => []
    },
    loading: {
        type: Boolean,
        default: false
    },
    processDetailData: {
        type: Array as () => ProcessInfo[],
        default: () => []
    },
    cpuColor: {
        type: String,
        default: '#1890ff'
    },
    memoryColor: {
        type: String,
        default: '#52c41a'
    }
})

// 组件事件
const emit = defineEmits(['refresh', 'time-range-change'])

// 时间范围选择
const end = Date.now()
const start = end - 3600 * 1000 // 最近1小时
const dateTimeRange = ref<[number, number] | null>([start, end])

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

    if (!props.data || props.data.length === 0) {
        // 清空图表以确保不显示旧数据
        chart.clear()
        return
    }

    // 按时间分组处理数据
    const timeMap = new Map<string, ProcessInfo[]>()

    props.data.forEach(item => {
        if (!timeMap.has(item.collectTime)) {
            timeMap.set(item.collectTime, [])
        }
        timeMap.get(item.collectTime)?.push(item)
    })

    // 转换为图表数据
    const times: string[] = Array.from(timeMap.keys()).sort()

    // 计算每个时间点的平均CPU和内存使用率
    const cpuData: number[] = []
    const memoryData: number[] = []

    times.forEach(time => {
        const processes = timeMap.get(time) || []
        if (processes.length > 0) {
            const avgCpu = processes.reduce((sum, p) => sum + (Number(p.cpuPercent) || 0), 0) / processes.length
            const avgMemory = processes.reduce((sum, p) => sum + (Number(p.memoryPercent) || 0), 0) / processes.length
            cpuData.push(Number(avgCpu.toFixed(2)))
            memoryData.push(Number(avgMemory.toFixed(2)))
        } else {
            cpuData.push(0)
            memoryData.push(0)
        }
    })

    const option = {
        tooltip: createAxisTooltipOptions({
            appendToBody: true,
            confine: true,
            extraCssText: 'z-index: 9999;',
            formatter: (params: any) => {
                const index = params[0].dataIndex
                const processes = timeMap.get(times[index]) || []

                let result = formatDate(times[index], 'YYYY-MM-DD HH:mm:ss') + '<br/>'

                params.forEach((param: any) => {
                    const color = param.color
                    const marker = `<span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:${color};"></span>`
                    result += marker + param.seriesName + ': ' + formatPercent(param.value) + '%<br/>'
                })

                if (processes.length > 0) {
                    result += `<span style="color:#666;">${t('process.processCount')}: ${processes.length}</span><br/>`
                    result += `<br/><b>${t('process.detailTitle')}</b><br/>`

                    const topProcesses = [...processes]
                        .sort((a, b) => (Number(b.cpuPercent) || 0) - (Number(a.cpuPercent) || 0))
                        .slice(0, 5)

                    topProcesses.forEach((process, i) => {
                        result += `<div style="padding-left:10px;margin:2px 0;">`
                        result += `<b>${i + 1}. ${process.processName ?? '-'}</b> <span style="color:#666;">(PID: ${process.processId ?? '-'})</span><br/>`
                        result += `<span style="padding-left:15px;">CPU: ${formatPercent(process.cpuPercent)}%, `
                        result += `${t('process.memory')}: ${formatPercent(process.memoryPercent)}%, `
                        result += `${t('process.threads')}: ${process.threadCount ?? '-'}</span>`
                        result += `</div>`
                    })

                    if (processes.length > 5) {
                        result += `<div style="padding-left:10px;color:#999;">${t('process.moreProcesses', { count: processes.length - 5 })}</div>`
                    }
                }

                return result
            }
        }),
        legend: {
            data: [t('process.avgCpu'), t('process.avgMemory')],
            bottom: 0,
            padding: [10, 20]
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '15%',
            top: '10%',
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
            min: 0,
            max: 100,
            position: 'left',
            axisLabel: {
                formatter: '{value}%'
            }
        },
        series: [
            {
                name: t('process.avgCpu'),
                type: 'line',
                data: cpuData,
                smooth: true,
                showSymbol: false,
                lineStyle: {
                    width: 2,
                    type: 'solid'
                },
                itemStyle: {
                    color: props.cpuColor
                }
            },
            {
                name: t('process.avgMemory'),
                type: 'line',
                data: memoryData,
                smooth: true,
                showSymbol: false,
                lineStyle: {
                    width: 2,
                    type: 'dashed'
                },
                itemStyle: {
                    color: props.memoryColor
                }
            }
        ]
    }

    chart.setOption(option)
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

/* 响应式设计 */
@media (max-width: 768px) {
    .chart-container {
        height: 250px;
    }
}
</style>

