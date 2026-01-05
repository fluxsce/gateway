<template>
    <n-card :title="displayTitle" :bordered="false" class="monitor-card">
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

            <div v-if="!loading && !chartData.length" class="chart-empty">
                <n-empty :description="t('hub0000.common.noData')" />
            </div>
        </div>
    </n-card>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
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
import type { DiskPartition } from '../../types'
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

// 组件属性定义
const props = defineProps({
    title: {
        type: String,
        default: '磁盘使用率'
    },
    data: {
        type: Array as () => DiskPartition[],
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
    diskDetailData: {
        type: Array as () => DiskPartition[],
        default: () => []
    }
})

// 计算属性 - 标题
const displayTitle = computed(() => props.title || t('hub0000.disk.title'))

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

// 计算属性 - 图表数据
const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return []

    // 按时间分组并计算平均值
    const timeGroups = new Map<string, { values: number[], partitions: DiskPartition[] }>()

    props.data.forEach(metric => {
        const time = metric.collectTime
        if (!timeGroups.has(time)) {
            timeGroups.set(time, { values: [], partitions: [] })
        }
        const group = timeGroups.get(time)!
        group.values.push(metric.usagePercent)
        group.partitions.push(metric)
    })

    // 后端数据已按时间排序，无需再次排序
    return Array.from(timeGroups.entries())
        .map(([time, group]) => ({
            time,
            value: Math.round(group.values.reduce((sum, val) => sum + val, 0) / group.values.length * 100) / 100,
            detail: {
                partitions: group.partitions,
                partitionCount: group.partitions.length,
                totalSpace: group.partitions.reduce((sum, p) => sum + p.totalSpace, 0),
                usedSpace: group.partitions.reduce((sum, p) => sum + p.usedSpace, 0),
                freeSpace: group.partitions.reduce((sum, p) => sum + p.freeSpace, 0),
                avgUsagePercent: Math.round(group.values.reduce((sum, val) => sum + val, 0) / group.values.length * 100) / 100
            }
        }))
})

// 格式化磁盘大小
const formatDiskSize = (bytes: number): string => {
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

    const warningThresholdValue = computed(() => props.warningThreshold ?? 80)
    const dangerThresholdValue = computed(() => props.dangerThreshold ?? 90)

    const option = {
        title: {
            show: false
        },
        tooltip: {
            trigger: 'axis',
            formatter: function (params: any) {
                const dataPoint = params[0]
                const detail = dataPoint.data.detail

                let result = ``
                result += `${dataPoint.marker}${t('hub0000.disk.usage')}: <b>${dataPoint.data.value}%</b><br/>`
                result += `<br/>${t('hub0000.disk.details')}:<br/>`
                result += `${t('hub0000.disk.partitionCount')}: ${detail.partitionCount}<br/>`
                result += `${t('hub0000.disk.totalSpace')}: ${formatDiskSize(detail.totalSpace)}<br/>`
                result += `${t('hub0000.disk.usedSpace')}: ${formatDiskSize(detail.usedSpace)}<br/>`
                result += `${t('hub0000.disk.freeSpace')}: ${formatDiskSize(detail.freeSpace)}<br/>`

                // 添加分区详情
                if (detail.partitions.length > 0) {
                    result += `<br/>${t('hub0000.disk.partitionDetails')}:<br/>`

                    // 只显示前3个分区，避免tooltip过长
                    const displayPartitions = detail.partitions.slice(0, 3)
                    displayPartitions.forEach((partition: DiskPartition) => {
                        result += `${partition.mountPoint}: ${partition.usagePercent.toFixed(1)}% (${formatDiskSize(partition.usedSpace)}/${formatDiskSize(partition.totalSpace)})<br/>`
                    })

                    // 如果有更多分区，显示省略信息
                    if (detail.partitions.length > 3) {
                        result += `... ${t('hub0000.disk.morePartitions', { count: detail.partitions.length - 3 })}`
                    }
                }

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
                name: t('hub0000.disk.usage'),
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
                    color: '#faad14'
                },
                itemStyle: {
                    color: function (params: any) {
                        const value = params.data.value
                        if (value >= dangerThresholdValue.value) {
                            return '#ff4d4f'
                        } else if (value >= warningThresholdValue.value) {
                            return '#faad14'
                        }
                        return '#faad14'
                    }
                },
                areaStyle: {
                    color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                        {
                            offset: 0,
                            color: 'rgba(250, 173, 20, 0.3)'
                        },
                        {
                            offset: 1,
                            color: 'rgba(250, 173, 20, 0.1)'
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
                            yAxis: warningThresholdValue.value,
                            label: {
                                formatter: `${t('hub0000.disk.warning')} ${warningThresholdValue.value}%`,
                                position: 'start'
                            }
                        },
                        {
                            yAxis: dangerThresholdValue.value,
                            lineStyle: {
                                color: '#ff4d4f'
                            },
                            label: {
                                formatter: `${t('hub0000.disk.danger')} ${dangerThresholdValue.value}%`,
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

/* 响应式设计 */
@media (max-width: 768px) {
    .chart-container {
        height: 250px;
    }
}
</style>