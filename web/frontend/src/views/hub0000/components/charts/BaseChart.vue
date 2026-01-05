<template>
    <div ref="chartRef" :style="{ width: '100%', height: height + 'px' }" class="base-chart" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch, nextTick } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface Props {
    /** 图表配置选项 */
    options: EChartsOption
    /** 图表高度 */
    height?: number
    /** 是否自动调整大小 */
    autoResize?: boolean
    /** 主题名称 */
    theme?: string
}

const props = withDefaults(defineProps<Props>(), {
    height: 400,
    autoResize: true,
    theme: 'default'
})

const emit = defineEmits<{
    chartReady: [chart: echarts.ECharts]
    chartClick: [params: any]
}>()

const chartRef = ref<HTMLDivElement>()
const chartInstance = ref<echarts.ECharts>()

// 初始化图表
const initChart = () => {
    if (!chartRef.value) return

    // 如果已有实例，先销毁
    if (chartInstance.value) {
        chartInstance.value.dispose()
    }

    // 创建新实例
    chartInstance.value = echarts.init(chartRef.value, props.theme)

    // 设置配置
    chartInstance.value.setOption(props.options)

    // 绑定事件
    chartInstance.value.on('click', (params) => {
        emit('chartClick', params)
    })

    // 触发就绪事件
    emit('chartReady', chartInstance.value)
}

// 更新图表配置
const updateChart = () => {
    if (chartInstance.value) {
        chartInstance.value.setOption(props.options, true)
    }
}

// 调整图表大小
const resizeChart = () => {
    if (chartInstance.value) {
        chartInstance.value.resize()
    }
}

// 监听配置变化
watch(
    () => props.options,
    () => {
        updateChart()
    },
    { deep: true }
)

// 监听主题变化
watch(
    () => props.theme,
    () => {
        initChart()
    }
)

// 窗口大小变化监听
const handleResize = () => {
    if (props.autoResize) {
        resizeChart()
    }
}

onMounted(() => {
    nextTick(() => {
        initChart()

        if (props.autoResize) {
            window.addEventListener('resize', handleResize)
        }
    })
})

onUnmounted(() => {
    if (chartInstance.value) {
        chartInstance.value.dispose()
    }

    if (props.autoResize) {
        window.removeEventListener('resize', handleResize)
    }
})

// 暴露方法
defineExpose({
    chartInstance,
    resizeChart,
    updateChart
})
</script>

<style scoped>
.base-chart {
    position: relative;
}
</style>