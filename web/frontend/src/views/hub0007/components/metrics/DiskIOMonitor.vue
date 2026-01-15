<template>
    <GCard title="磁盘IO监控" :show-title="true">
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

            <div v-if="!loading && (!data || !Array.isArray(data) || data.length === 0)" class="chart-empty">
                <n-empty description="暂无数据" />
            </div>
        </div>

        <div v-if="diskIODetailData && diskIODetailData.length > 0" class="detail-section">
            <n-divider>磁盘IO详情</n-divider>
            <n-data-table :columns="diskIOColumns" :data="diskIOTableData" :pagination="tablePagination"
                :bordered="false" size="small" />
        </div>
    </GCard>
</template>

<script setup lang="ts">
import { GCard } from '@/components/gcard'
import { formatBytes, formatDate } from '@/utils/format'
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
import { NButton, NDataTable, NDatePicker, NDivider, NEmpty, NIcon, NSpin } from 'naive-ui'
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import type { DiskIOStats } from '../../types'

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
        type: Array as () => DiskIOStats[],
        default: () => []
    },
    loading: {
        type: Boolean,
        default: false
    },
    diskIODetailData: {
        type: Array as () => DiskIOStats[],
        default: () => []
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

// 表格分页
const tablePagination = {
    pageSize: 5,
    page: 1
}

// 磁盘IO表格列定义
const diskIOColumns = computed(() => [
    {
        title: '设备名称',
        key: 'deviceName'
    },
    {
        title: '读取速率',
        key: 'readRate',
        render: (row: DiskIOStats) => formatBytes(row.readRate) + '/s'
    },
    {
        title: '写入速率',
        key: 'writeRate',
        render: (row: DiskIOStats) => formatBytes(row.writeRate) + '/s'
    },
    {
        title: '读取次数',
        key: 'readCount'
    },
    {
        title: '写入次数',
        key: 'writeCount'
    },
    {
        title: '采集时间',
        key: 'collectTime',
        render: (row: DiskIOStats) => formatDate(row.collectTime)
    }
])

// 磁盘IO表格数据
const diskIOTableData = computed(() => {
    if (!props.diskIODetailData) return []

    // 按设备名称分组，获取每个设备的最新数据
    const latestDataByDevice = new Map<string, DiskIOStats>()

    props.diskIODetailData.forEach(item => {
        const existing = latestDataByDevice.get(item.deviceName)
        if (!existing || new Date(item.collectTime) > new Date(existing.collectTime)) {
            latestDataByDevice.set(item.deviceName, item)
        }
    })

    return Array.from(latestDataByDevice.values())
})

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

    // 按时间和设备名称分组处理数据
    const timeMap = new Map<string, Map<string, DiskIOStats>>()

    props.data.forEach(item => {
        if (!timeMap.has(item.collectTime)) {
            timeMap.set(item.collectTime, new Map())
        }
        timeMap.get(item.collectTime)?.set(item.deviceName, item)
    })

    // 转换为图表数据
    const times: string[] = Array.from(timeMap.keys()).sort()
    const devices = new Set<string>()

    props.data.forEach(item => devices.add(item.deviceName))

    // 定义不同设备的颜色方案
    const deviceColors = [
        { read: '#1890ff', write: '#36cfc9' }, // 蓝色和青色
        { read: '#722ed1', write: '#eb2f96' }, // 紫色和粉色
        { read: '#fa8c16', write: '#faad14' }, // 橙色和黄色
        { read: '#52c41a', write: '#a0d911' }, // 绿色和浅绿色
        { read: '#f5222d', write: '#ff7a45' }  // 红色和橙红色
    ]

    const series: any[] = []

    // 为每个设备创建读写系列
    Array.from(devices).forEach((device, index) => {
        // 根据设备索引获取颜色，如果设备数量超过颜色数组长度，则循环使用
        const colorIndex = index % deviceColors.length
        const deviceColor = deviceColors[colorIndex]

        const readData: number[] = []
        const writeData: number[] = []

        times.forEach(time => {
            const deviceData = timeMap.get(time)?.get(device)
            readData.push(deviceData ? deviceData.readRate : 0)
            writeData.push(deviceData ? deviceData.writeRate : 0)
        })

        series.push({
            name: `${device} 读取`,
            type: 'line',
            data: readData,
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: 2,
                type: 'solid'
            },
            itemStyle: {
                color: deviceColor.read
            }
        })

        series.push({
            name: `${device} 写入`,
            type: 'line',
            data: writeData,
            smooth: true,
            showSymbol: false,
            lineStyle: {
                width: 2,
                type: 'dashed'
            },
            itemStyle: {
                color: deviceColor.write
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
            formatter: (params: any) => {
                const firstParam = Array.isArray(params) ? params[0] : params
                const time = times[firstParam.dataIndex]
                let result = formatDate(time, 'YYYY-MM-DD HH:mm:ss') + '<br/>'

                // 按设备名称分组
                const deviceGroups = new Map<string, any[]>()

                params.forEach((param: any) => {
                    const seriesName = param.seriesName
                    const deviceName = seriesName.split(' ')[0]

                    if (!deviceGroups.has(deviceName)) {
                        deviceGroups.set(deviceName, [])
                    }
                    deviceGroups.get(deviceName)?.push(param)
                })

                // 按设备组织显示
                deviceGroups.forEach((params, deviceName) => {
                    result += `<div style="margin-top: 3px;"><b>${deviceName}</b></div>`

                    params.forEach((param: any) => {
                        const color = param.color
                        const isRead = param.seriesName.includes('读取')
                        const type = isRead ? '读取' : '写入'
                        const marker = `<span style="display:inline-block;margin-right:4px;border-radius:10px;width:10px;height:10px;background-color:${color};border:1px solid ${color}"></span>`
                        result += `<div style="padding-left: 12px;line-height:18px">${marker}${type}: ${formatBytes(Number(param.value))}/s</div>`
                    })
                })

                return result
            }
        },
        legend: {
            type: 'scroll',
            bottom: 0,
            padding: [10, 20],
            selected: series.reduce((acc: any, s: any) => {
                // 默认显示所有设备数据
                acc[s.name] = true;
                return acc;
            }, {}),
            formatter: (name: string) => {
                const deviceName = name.split(' ')[0];
                const type = name.includes('读取') ? '读取' : '写入';
                return `${deviceName} ${type}`;
            }
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
                formatter: (value: number) => formatBytes(value) + '/s'
            }
        },
        series
    }

    chart.setOption(option, true) // 添加 true 参数来完全替换之前的配置
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

