<template>
    <GCard :title="title || defaultProps.title" :show-title="true">
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
import type { NetworkInterface } from '../../types'

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
const props = defineProps<{
    title?: string
    data: NetworkInterface[] | null
    loading: boolean
    uploadColor?: string
    downloadColor?: string
}>()

// 默认属性值
const defaultProps = {
    title: '网络流量',
    uploadColor: '#ff4d4f',
    downloadColor: '#52c41a',
}

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

// 计算属性 - 图表数据
const chartData = computed(() => {
    if (!props.data || props.data.length === 0) return []

    // 按时间分组
    const timeGroups = new Map<string, {
        interfaces: NetworkInterface[]
    }>()

    props.data.forEach(metric => {
        const time = metric.collectTime
        if (!timeGroups.has(time)) {
            timeGroups.set(time, { interfaces: [] })
        }
        const group = timeGroups.get(time)!
        group.interfaces.push(metric)
    })

    // 转换为图表数据并按时间正序排序
    return Array.from(timeGroups.entries())
        .sort((a, b) => new Date(a[0]).getTime() - new Date(b[0]).getTime())  // 按时间正序排序
        .map(([time, group]) => {
            return {
                time,
                detail: {
                    interfaces: group.interfaces,
                    interfaceCount: group.interfaces.length
                }
            }
        })
})

// 格式化网络速率
const formatNetworkSpeed = (bytesPerSecond: number): string => {
    if (bytesPerSecond === 0) return '0 B/s'

    if (bytesPerSecond < 1024) {
        return `${bytesPerSecond.toFixed(2)} B/s`
    } else if (bytesPerSecond < 1024 * 1024) {
        return `${(bytesPerSecond / 1024).toFixed(2)} KB/s`
    } else if (bytesPerSecond < 1024 * 1024 * 1024) {
        return `${(bytesPerSecond / (1024 * 1024)).toFixed(2)} MB/s`
    } else {
        return `${(bytesPerSecond / (1024 * 1024 * 1024)).toFixed(2)} GB/s`
    }
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


    // 提取所有时间点和网络接口
    const times = data.map(item => formatDate(new Date(item.time), 'HH:mm:ss'))
    const allInterfaces = new Set<string>()

    // 收集所有接口，并优先排序IPv4接口
    const ipv4Interfaces: string[] = []
    const otherInterfaces: string[] = []

    data.forEach(item => {
        item.detail.interfaces.forEach(iface => {
            if (!allInterfaces.has(iface.interfaceName)) {
                // 判断是否为IPv4接口（通常IPv4地址包含点号分隔）
                // 检查ipAddresses是否包含IPv4地址格式
                const hasIpv4 = iface.ipAddresses &&
                    (iface.ipAddresses.includes('.') && !iface.ipAddresses.includes(':')) ||
                    iface.interfaceName.match(/eth\d+|en\d+|bond\d+|wlan\d+|ens\d+|enp\d+|lo\b|vlan\d+/i);

                if (hasIpv4) {
                    ipv4Interfaces.push(iface.interfaceName)
                } else {
                    otherInterfaces.push(iface.interfaceName)
                }
                allInterfaces.add(iface.interfaceName)
            }
        })
    })

    // 合并接口列表，IPv4接口优先
    const sortedInterfaces = [...ipv4Interfaces, ...otherInterfaces]

    // 生成接口颜色
    const interfaceColors = [
        { send: '#fa541c', receive: '#73d13d' },
        { send: '#722ed1', receive: '#13c2c2' },
        { send: '#f5222d', receive: '#52c41a' },
        { send: '#fa8c16', receive: '#1890ff' },
        { send: '#eb2f96', receive: '#fadb14' }
    ]

    // 为每个接口创建系列数据
    const series: any[] = []

    // 添加每个接口的数据
    sortedInterfaces.forEach((ifaceName, index) => {
        const colorIndex = index % interfaceColors.length
        const colors = interfaceColors[colorIndex]

        // 为接口创建上传数据系列
        const sendData = data.map(item => {
            const iface = item.detail.interfaces.find(i => i.interfaceName === ifaceName)
            // 转换为MB/s
            return iface ? Math.round(iface.sendRate / (1024 * 1024) * 100) / 100 : 0
        })

        // 为接口创建下载数据系列
        const receiveData = data.map(item => {
            const iface = item.detail.interfaces.find(i => i.interfaceName === ifaceName)
            // 转换为MB/s
            return iface ? Math.round(iface.receiveRate / (1024 * 1024) * 100) / 100 : 0
        })

        // IPv4接口使用粗线条，非IPv4接口使用细线条
        const lineWidth = ipv4Interfaces.includes(ifaceName) ? 2 : 1

        series.push({
            name: `${ifaceName} 上传`,
            type: 'line',
            data: sendData,
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: lineWidth,
                type: 'solid'
            },
            itemStyle: {
                color: colors.send
            },
            emphasis: {
                focus: 'series'
            },
            tooltip: {
                valueFormatter: (value: number) => `${value} MB/s`
            }
        })

        series.push({
            name: `${ifaceName} 下载`,
            type: 'line',
            data: receiveData,
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: lineWidth,
                type: 'dashed'
            },
            itemStyle: {
                color: colors.receive
            },
            emphasis: {
                focus: 'series'
            },
            tooltip: {
                valueFormatter: (value: number) => `${value} MB/s`
            }
        })
    })

    const option = {
        title: {
            show: false
        },
        tooltip: {
            trigger: 'axis',
            confine: false,
            appendToBody: true,
            axisPointer: {
                type: 'cross',
                label: {
                    backgroundColor: '#6a7985'
                }
            },
            formatter: function (params: any) {
                const time = params[0].axisValue
                const timeIndex = times.indexOf(time)
                const detail = data[timeIndex]?.detail

                let result = `${time}<br/>`

                // 按接口分组显示
                const interfaceGroups = new Map<string, any[]>()

                params.forEach((param: any) => {
                    const seriesName = param.seriesName
                    const ifaceName = seriesName.split(' ')[0]

                    if (!interfaceGroups.has(ifaceName)) {
                        interfaceGroups.set(ifaceName, [])
                    }
                    interfaceGroups.get(ifaceName)!.push(param)
                })

                if (detail) {
                    // 确保显示所有接口，不仅是当前选中的数据系列
                    const allInterfaces = new Set<string>()

                    // 先添加所有选中的系列对应的接口
                    interfaceGroups.forEach((_, ifaceName) => {
                        allInterfaces.add(ifaceName)
                    })

                    // 然后添加detail中的所有接口
                    detail.interfaces.forEach(iface => {
                        allInterfaces.add(iface.interfaceName)
                    })

                    // 显示接口总数
                    result += `接口总数: ${detail.interfaceCount}<br/>`

                    // 先显示IPv4接口，再显示其他接口
                    const ipv4Interfaces: string[] = []
                    const otherInterfaces: string[] = []

                    Array.from(allInterfaces).forEach(ifaceName => {
                        const iface = detail.interfaces.find(i => i.interfaceName === ifaceName)
                        if (!iface) return

                        // 判断是否为IPv4接口
                        const hasIpv4 = iface.ipAddresses &&
                            (iface.ipAddresses.includes('.') && !iface.ipAddresses.includes(':')) ||
                            iface.interfaceName.match(/eth\d+|en\d+|bond\d+|wlan\d+|ens\d+|enp\d+|lo\b|vlan\d+/i);

                        if (hasIpv4) {
                            ipv4Interfaces.push(ifaceName)
                        } else {
                            otherInterfaces.push(ifaceName)
                        }
                    })

                    // 显示所有接口详情
                    const allInterfacesList = [...ipv4Interfaces, ...otherInterfaces]

                    if (allInterfacesList.length > 0) {
                        result += `<br/>接口详情:<br/>`

                        allInterfacesList.forEach(ifaceName => {
                            const iface = detail.interfaces.find(i => i.interfaceName === ifaceName)
                            if (!iface) return

                            result += `<div style="padding:3px 0;border-bottom:1px solid #eee;">`
                            result += `<b>${ifaceName}</b>: `

                            // 是否为IPv4接口
                            const isIpv4 = ipv4Interfaces.includes(ifaceName)
                            if (isIpv4) {
                                result += `<span style="color:#1890ff;">[IPv4]</span> `
                            }

                            result += `<br/><span style="padding-left:10px;color:#ff4d4f;">↑${formatNetworkSpeed(iface.sendRate)}</span> `
                            result += `<span style="padding-left:5px;color:#52c41a;">↓${formatNetworkSpeed(iface.receiveRate)}</span>`

                            if (iface.errorsReceived > 0 || iface.errorsSent > 0) {
                                result += ` <span style="color:#ff7875;">(错误: ${iface.errorsReceived + iface.errorsSent})</span>`
                            }

                            // 显示IP地址信息
                            if (iface.ipAddresses) {
                                result += `<br/><span style="padding-left:10px;color:#999;font-size:12px;">IP: ${iface.ipAddresses}</span>`
                            }

                            result += `</div>`
                        })
                    }
                }

                return result
            }
        },
        legend: {
            type: 'scroll',
            data: series.map(s => s.name),
            bottom: 0,
            padding: [5, 10],
            selected: series.reduce((acc: any, s: any) => {
                // 默认显示IPv4接口
                const ifaceName = s.name.split(' ')[0]
                acc[s.name] = ipv4Interfaces.includes(ifaceName);
                return acc;
            }, {})
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '40px',
            top: '10px',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            boundaryGap: false,
            data: times,
            axisLabel: {
                rotate: 45
            }
        },
        yAxis: {
            type: 'value',
            splitLine: {
                show: true,
                lineStyle: {
                    type: 'dashed'
                }
            },
            axisLabel: {
                formatter: (value: number) => value.toFixed(2) + ' MB/s'
            }
        },
        series: series
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

