<template>
    <div class="network-traffic-chart">
        <div class="chart-header">
            <h3 class="chart-title">{{ title }}</h3>
            <div class="chart-controls">
                <n-date-picker v-model:value="timeRange" type="datetimerange" :shortcuts="timeRangeShortcuts"
                    @update:value="handleTimeRangeChange" placeholder="选择时间范围" size="small" style="width: 280px"
                    clearable />
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

        <div class="traffic-summary">
            <div class="summary-item">
                <span class="summary-label">
                    <span class="legend-dot upload"></span>
                    上传总量
                </span>
                <span class="summary-value">{{ formatBytes(totalUpload) }}</span>
            </div>
            <div class="summary-item">
                <span class="summary-label">
                    <span class="legend-dot download"></span>
                    下载总量
                </span>
                <span class="summary-value">{{ formatBytes(totalDownload) }}</span>
            </div>
            <div class="summary-item">
                <span class="summary-label">总流量</span>
                <span class="summary-value">{{ formatBytes(totalTraffic) }}</span>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NButton, NIcon, NDatePicker } from 'naive-ui'
import { ReloadOutlined } from '@vicons/antd'
import BaseChart from './BaseChart.vue'
import type { EChartsOption } from 'echarts'

interface NetworkTrafficData {
    time: string
    upload: number
    download: number
}

interface Props {
    /** 图表标题 */
    title: string
    /** 网络流量数据 */
    data: NetworkTrafficData[]
    /** 图表高度 */
    height?: number
    /** 是否加载中 */
    loading?: boolean
    /** 上传流量颜色 */
    uploadColor?: string
    /** 下载流量颜色 */
    downloadColor?: string
}

const props = withDefaults(defineProps<Props>(), {
    height: 300,
    loading: false,
    uploadColor: '#ff4d4f',
    downloadColor: '#52c41a'
})

const emit = defineEmits<{
    refresh: []
    chartClick: [params: any]
    timeRangeChange: [timeRange: [number, number] | null]
}>()

// 时间范围状态
const timeRange = ref<[number, number] | null>(null)

// 时间范围快捷选项（本地时区）
const timeRangeShortcuts = {
    '最近2小时': () => {
        const now = new Date()
        const start = now.getTime() - 2 * 60 * 60 * 1000
        return [start, now.getTime()] as [number, number]
    },
    '最近6小时': () => {
        const now = new Date()
        const start = now.getTime() - 6 * 60 * 60 * 1000
        return [start, now.getTime()] as [number, number]
    },
    '最近24小时': () => {
        const now = new Date()
        const start = now.getTime() - 24 * 60 * 60 * 1000
        return [start, now.getTime()] as [number, number]
    },
    '最近7天': () => {
        const now = new Date()
        const start = now.getTime() - 7 * 24 * 60 * 60 * 1000
        return [start, now.getTime()] as [number, number]
    },
    '最近30天': () => {
        const now = new Date()
        const start = now.getTime() - 30 * 24 * 60 * 60 * 1000
        return [start, now.getTime()] as [number, number]
    }
}

// 初始化时间范围（默认最近2小时，本地时区）
const initTimeRange = () => {
    const now = new Date()
    const start = now.getTime() - 2 * 60 * 60 * 1000
    timeRange.value = [start, now.getTime()]
}

// 计算流量总量
const totalUpload = computed(() => {
    return props.data.reduce((sum, item) => sum + item.upload, 0)
})

const totalDownload = computed(() => {
    return props.data.reduce((sum, item) => sum + item.download, 0)
})

const totalTraffic = computed(() => {
    return totalUpload.value + totalDownload.value
})

// 格式化字节数
const formatBytes = (bytes: number): string => {
    if (bytes === 0) return '0 B'

    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))

    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
    const times = props.data.map(item => item.time)
    const uploadData = props.data.map(item => item.upload)
    const downloadData = props.data.map(item => item.download)

    return {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'cross'
            },
            formatter: (params: any) => {
                const time = params[0].name
                const upload = params.find((p: any) => p.seriesName === '上传流量')?.value || 0
                const download = params.find((p: any) => p.seriesName === '下载流量')?.value || 0

                return `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 4px;">${time}</div>
            <div style="display: flex; align-items: center; margin-bottom: 2px;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${props.uploadColor}; border-radius: 50%;"></span>
              <span>上传: ${formatBytes(upload)}</span>
            </div>
            <div style="display: flex; align-items: center;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${props.downloadColor}; border-radius: 50%;"></span>
              <span>下载: ${formatBytes(download)}</span>
            </div>
          </div>
        `
            }
        },
        legend: {
            data: ['上传流量', '下载流量'],
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
                formatter: (value: number) => formatBytes(value)
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
                name: '上传流量',
                type: 'line',
                data: uploadData,
                smooth: true,
                symbol: 'circle',
                symbolSize: 4,
                lineStyle: {
                    color: props.uploadColor,
                    width: 2
                },
                itemStyle: {
                    color: props.uploadColor
                },
                areaStyle: {
                    color: {
                        type: 'linear' as const,
                        x: 0,
                        y: 0,
                        x2: 0,
                        y2: 1,
                        colorStops: [
                            {
                                offset: 0,
                                color: props.uploadColor + '40'
                            },
                            {
                                offset: 1,
                                color: props.uploadColor + '10'
                            }
                        ]
                    }
                }
            },
            {
                name: '下载流量',
                type: 'line',
                data: downloadData,
                smooth: true,
                symbol: 'circle',
                symbolSize: 4,
                lineStyle: {
                    color: props.downloadColor,
                    width: 2
                },
                itemStyle: {
                    color: props.downloadColor
                },
                areaStyle: {
                    color: {
                        type: 'linear' as const,
                        x: 0,
                        y: 0,
                        x2: 0,
                        y2: 1,
                        colorStops: [
                            {
                                offset: 0,
                                color: props.downloadColor + '40'
                            },
                            {
                                offset: 1,
                                color: props.downloadColor + '10'
                            }
                        ]
                    }
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
.network-traffic-chart {
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

.traffic-summary {
    display: flex;
    justify-content: space-around;
    margin-top: 16px;
    padding-top: 16px;
    border-top: 1px solid #f0f0f0;
}

.summary-item {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
}

.summary-label {
    display: flex;
    align-items: center;
    gap: 4px;
    font-size: 12px;
    color: #8c8c8c;
}

.legend-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
}

.legend-dot.upload {
    background-color: #ff4d4f;
}

.legend-dot.download {
    background-color: #52c41a;
}

.summary-value {
    font-size: 16px;
    font-weight: 600;
    color: #262626;
}

/* 暗色主题支持 */
.dark .network-traffic-chart {
    background: #1f1f1f;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
}

.dark .chart-title {
    color: #e0e0e0;
}

.dark .traffic-summary {
    border-top-color: #333;
}

.dark .summary-label {
    color: #8c8c8c;
}

.dark .summary-value {
    color: #e0e0e0;
}
</style>