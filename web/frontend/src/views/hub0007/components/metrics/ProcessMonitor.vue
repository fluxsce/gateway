<template>
    <GCard title="进程监控" :show-title="true">
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

            <div v-if="!loading && !data?.length" class="chart-empty">
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
import type { ProcessInfo } from '../../types'

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

// 时间范围快捷选项
const timeRangeShortcuts: Record<string, () => [number, number]> = {
    '最近1小时': () => {
        const end = Date.now()
        const start = end - 3600 * 1000
        return [start, end]
    },
    '最近6小时': () => {
        const end = Date.now()
        const start = end - 6 * 3600 * 1000
        return [start, end]
    },
    '最近12小时': () => {
        const end = Date.now()
        const start = end - 12 * 3600 * 1000
        return [start, end]
    },
    '最近24小时': () => {
        const end = Date.now()
        const start = end - 24 * 3600 * 1000
        return [start, end]
    },
    '最近7天': () => {
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
            const avgCpu = processes.reduce((sum, p) => sum + p.cpuPercent, 0) / processes.length
            const avgMemory = processes.reduce((sum, p) => sum + p.memoryPercent, 0) / processes.length
            cpuData.push(Number(avgCpu.toFixed(2)))
            memoryData.push(Number(avgMemory.toFixed(2)))
        } else {
            cpuData.push(0)
            memoryData.push(0)
        }
    })

    const option = {
        tooltip: {
            trigger: 'axis',
            confine: false,
            appendToBody: true,
            formatter: (params: any) => {
                const index = params[0].dataIndex
                const processes = timeMap.get(times[index]) || []

                let result = formatDate(times[index], 'YYYY-MM-DD HH:mm:ss') + '<br/>'

                // 添加平均值信息
                params.forEach((param: any) => {
                    const color = param.color
                    const marker = `<span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:${color};"></span>`
                    result += marker + param.seriesName + ': ' + param.value + '%<br/>'
                })

                // 添加进程数量信息
                if (processes.length > 0) {
                    result += `<span style="color:#666;">进程数量: ${processes.length}</span><br/>`
                }

                // 添加前5个进程的CPU和内存使用率
                if (processes.length > 0) {
                    result += '<br/><b>进程详情</b><br/>'

                    // 按CPU使用率排序
                    const topProcesses = [...processes]
                        .sort((a, b) => b.cpuPercent - a.cpuPercent)
                        .slice(0, 5)

                    topProcesses.forEach((process, i) => {
                        result += `<div style="padding-left:10px;margin:2px 0;">`
                        result += `<b>${i + 1}. ${process.processName}</b> <span style="color:#666;">(PID: ${process.processId})</span><br/>`
                        result += `<span style="padding-left:15px;">CPU: ${process.cpuPercent.toFixed(2)}%, `
                        result += `内存: ${process.memoryPercent.toFixed(2)}%, `
                        result += `线程: ${process.threadCount}</span>`
                        result += `</div>`
                    })

                    if (processes.length > 5) {
                        result += `<div style="padding-left:10px;color:#999;">... 以及 ${processes.length - 5} 个其他进程</div>`
                    }
                }

                return result
            }
        },
        legend: {
            data: ['平均CPU使用率', '平均内存使用率'],
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
                name: '平均CPU使用率',
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
                name: '平均内存使用率',
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

