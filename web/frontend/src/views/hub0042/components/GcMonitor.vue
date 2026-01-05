<template>
  <div class="gc-monitor">
    <!-- GC统计卡片 -->
    <n-grid :cols="4" :x-gap="16" :y-gap="16">
      <n-gi>
        <n-card :bordered="false">
          <n-statistic :label="t('totalGcCount')" :value="totalGcCount">
            <template #suffix>
              <span style="font-size: 14px">{{ t('times') }}</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      
      <n-gi>
        <n-card :bordered="false">
          <n-statistic :label="t('youngGcCount')" :value="youngGcCount">
            <template #suffix>
              <span style="font-size: 14px">{{ t('times') }}</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      
      <n-gi>
        <n-card :bordered="false">
          <n-statistic :label="t('fullGcCount')" :value="fullGcCount">
            <template #suffix>
              <span style="font-size: 14px">{{ t('times') }}</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
      
      <n-gi>
        <n-card :bordered="false">
          <n-statistic :label="t('avgGcTime')" :value="avgGcTime.toFixed(2)">
            <template #suffix>
              <span style="font-size: 14px">ms</span>
            </template>
          </n-statistic>
        </n-card>
      </n-gi>
    </n-grid>

    <!-- GC趋势图表 -->
    <n-space vertical :size="16" style="margin-top: 16px">
      <!-- 时间选择和刷新控制 -->
      <n-card :bordered="false" size="small">
        <n-space :size="12">
          <n-date-picker
            v-model:value="timeRange"
            type="datetimerange"
            :shortcuts="timeRangeShortcuts"
            @update:value="handleTimeRangeChange"
            placeholder="选择时间范围"
            size="small"
            style="width: 280px"
            clearable
          />
          <n-button size="small" @click="handleRefresh" :loading="loading">
            <template #icon>
              <n-icon :component="RefreshOutline" />
            </template>
            刷新
          </n-button>
        </n-space>
      </n-card>

      <!-- GC次数和时间趋势 -->
      <n-card :title="t('gcCountTimeTrend')" :bordered="false">
        <div ref="gcTrendChartRef" style="width: 100%; height: 400px"></div>
      </n-card>
      
      <!-- 内存区域使用趋势 -->
      <n-card :title="t('memoryRegionTrend')" :bordered="false">
        <div ref="memoryTrendChartRef" style="width: 100%; height: 400px"></div>
      </n-card>
    </n-space>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { NButton, NIcon, NDatePicker, NSpace } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate } from '@/utils/format'
import { useGcMonitor } from '../hooks'
import * as echarts from 'echarts/core'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'
import { LineChart, BarChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'

echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  LineChart,
  BarChart,
  CanvasRenderer
])

const props = defineProps<{
  jvmResourceId: string
}>()

const { t } = useModuleI18n('hub0042')

const {
  loading,
  gcSnapshots,
  totalGcCount,
  youngGcCount,
  fullGcCount,
  avgGcTime,
  queryGcSnapshots
} = useGcMonitor()

// 状态
const timeRange = ref<[number, number] | null>(null)

// 图表实例
const gcTrendChartRef = ref<HTMLDivElement>()
const memoryTrendChartRef = ref<HTMLDivElement>()
let gcTrendChart: echarts.ECharts | null = null
let memoryTrendChart: echarts.ECharts | null = null

// 时间范围快捷选项
const timeRangeShortcuts = {
  '最近1小时': () => {
    const now = Date.now()
    return [now - 60 * 60 * 1000, now] as [number, number]
  },
  '最近2小时': () => {
    const now = Date.now()
    return [now - 2 * 60 * 60 * 1000, now] as [number, number]
  },
  '最近6小时': () => {
    const now = Date.now()
    return [now - 6 * 60 * 60 * 1000, now] as [number, number]
  },
  '最近12小时': () => {
    const now = Date.now()
    return [now - 12 * 60 * 60 * 1000, now] as [number, number]
  },
  '最近24小时': () => {
    const now = Date.now()
    return [now - 24 * 60 * 60 * 1000, now] as [number, number]
  }
}

// GC趋势图表配置
const gcTrendOptions = computed<EChartsOption>(() => {
  if (!gcSnapshots.value || gcSnapshots.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = gcSnapshots.value.map(item => {
    const date = new Date(item.collectionTime)
    return date.toLocaleString('zh-CN', { 
      month: '2-digit', 
      day: '2-digit', 
      hour: '2-digit', 
      minute: '2-digit'
    })
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['Young GC次数', 'Full GC次数', 'Young GC时间', 'Full GC时间'],
      bottom: 10
    },
    grid: {
      left: '10px',
      right: '10px',
      bottom: '60px',
      top: '30px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        fontSize: 12,
        rotate: 30,
        color: '#595959'
      }
    },
    yAxis: [
      {
        type: 'value',
        name: 'GC次数',
        position: 'left',
        axisLabel: {
          fontSize: 13,
          color: '#595959'
        }
      },
      {
        type: 'value',
        name: 'GC时间(s)',
        position: 'right',
        axisLabel: {
          fontSize: 13,
          color: '#595959',
          formatter: (value: number) => value.toFixed(2) + 's'
        }
      }
    ],
    series: [
      {
        name: 'Young GC次数',
        type: 'line',
        data: gcSnapshots.value.map(item => item.ygc || 0),
        smooth: true,
        lineStyle: { color: '#52c41a', width: 2 },
        itemStyle: { color: '#52c41a' }
      },
      {
        name: 'Full GC次数',
        type: 'line',
        data: gcSnapshots.value.map(item => item.fgc || 0),
        smooth: true,
        lineStyle: { color: '#ff4d4f', width: 2 },
        itemStyle: { color: '#ff4d4f' }
      },
      {
        name: 'Young GC时间',
        type: 'line',
        yAxisIndex: 1,
        data: gcSnapshots.value.map(item => item.ygct || 0),
        smooth: true,
        lineStyle: { color: '#1890ff', width: 2, type: 'dashed' },
        itemStyle: { color: '#1890ff' }
      },
      {
        name: 'Full GC时间',
        type: 'line',
        yAxisIndex: 1,
        data: gcSnapshots.value.map(item => item.fgct || 0),
        smooth: true,
        lineStyle: { color: '#faad14', width: 2, type: 'dashed' },
        itemStyle: { color: '#faad14' }
      }
    ]
  }
})

// 内存区域趋势图表配置
const memoryTrendOptions = computed<EChartsOption>(() => {
  if (!gcSnapshots.value || gcSnapshots.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = gcSnapshots.value.map(item => {
    const date = new Date(item.collectionTime)
    return date.toLocaleString('zh-CN', { 
      month: '2-digit', 
      day: '2-digit', 
      hour: '2-digit', 
      minute: '2-digit'
    })
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        let html = `<div style="font-size: 12px;"><strong>${params[0].name}</strong><br/>`
        params.forEach((param: any) => {
          html += `<span style="color: ${param.color};">●</span> ${param.seriesName}: ${param.value} KB<br/>`
        })
        html += '</div>'
        return html
      }
    },
    legend: {
      data: ['Eden', 'Survivor', 'Old Gen', 'Metaspace'],
      bottom: 10
    },
    grid: {
      left: '10px',
      right: '30px',
      bottom: '60px',
      top: '30px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        fontSize: 12,
        rotate: 30,
        color: '#595959'
      }
    },
    yAxis: {
      type: 'value',
      name: '内存使用量(KB)',
      axisLabel: {
        fontSize: 13,
        color: '#595959'
      }
    },
    series: [
      {
        name: 'Eden',
        type: 'line',
        data: gcSnapshots.value.map(item => item.eu || 0),
        smooth: true,
        areaStyle: { opacity: 0.3 },
        lineStyle: { color: '#52c41a', width: 2 },
        itemStyle: { color: '#52c41a' }
      },
      {
        name: 'Survivor',
        type: 'line',
        data: gcSnapshots.value.map(item => (item.s0u || 0) + (item.s1u || 0)),
        smooth: true,
        areaStyle: { opacity: 0.3 },
        lineStyle: { color: '#1890ff', width: 2 },
        itemStyle: { color: '#1890ff' }
      },
      {
        name: 'Old Gen',
        type: 'line',
        data: gcSnapshots.value.map(item => item.ou || 0),
        smooth: true,
        areaStyle: { opacity: 0.3 },
        lineStyle: { color: '#faad14', width: 2 },
        itemStyle: { color: '#faad14' }
      },
      {
        name: 'Metaspace',
        type: 'line',
        data: gcSnapshots.value.map(item => item.mu || 0),
        smooth: true,
        areaStyle: { opacity: 0.3 },
        lineStyle: { color: '#722ed1', width: 2 },
        itemStyle: { color: '#722ed1' }
      }
    ]
  }
})

// 查询GC快照列表
const loadGcSnapshots = async () => {
  if (!props.jvmResourceId || !timeRange.value) return
  
  const [startTime, endTime] = timeRange.value
  await queryGcSnapshots({
    jvmResourceId: props.jvmResourceId,
    startTime: formatDate(startTime),
    endTime: formatDate(endTime)
  })
}

// 处理时间范围变化
const handleTimeRangeChange = (range: [number, number] | null) => {
  timeRange.value = range
  if (range) {
    loadGcSnapshots()
  }
}

// 处理刷新
const handleRefresh = () => {
  if (timeRange.value) {
    loadGcSnapshots()
  }
}

// 初始化GC趋势图表
const initGcTrendChart = () => {
  if (!gcTrendChartRef.value) return
  
  const { clientWidth, clientHeight } = gcTrendChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initGcTrendChart(), 100)
    return
  }
  
  if (gcTrendChart) {
    gcTrendChart.dispose()
  }
  
  gcTrendChart = echarts.init(gcTrendChartRef.value)
  gcTrendChart.setOption(gcTrendOptions.value)
  window.addEventListener('resize', handleResize)
}

// 初始化内存趋势图表
const initMemoryTrendChart = () => {
  if (!memoryTrendChartRef.value) return
  
  const { clientWidth, clientHeight } = memoryTrendChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initMemoryTrendChart(), 100)
    return
  }
  
  if (memoryTrendChart) {
    memoryTrendChart.dispose()
  }
  
  memoryTrendChart = echarts.init(memoryTrendChartRef.value)
  memoryTrendChart.setOption(memoryTrendOptions.value)
}

// 更新图表
const updateCharts = () => {
  if (gcTrendChart) {
    gcTrendChart.setOption(gcTrendOptions.value)
  }
  if (memoryTrendChart) {
    memoryTrendChart.setOption(memoryTrendOptions.value)
  }
}

// 处理窗口大小变化
const handleResize = () => {
  if (gcTrendChart) {
    gcTrendChart.resize()
  }
  if (memoryTrendChart) {
    memoryTrendChart.resize()
  }
}

// 监听数据变化
watch(() => gcSnapshots.value, () => {
  nextTick(() => {
    updateCharts()
  })
}, { deep: true })

// 监听资源ID变化
watch(() => props.jvmResourceId, () => {
  if (timeRange.value) {
    loadGcSnapshots()
  }
})

onMounted(() => {
  // 两个图表都可见，同时初始化
  nextTick(() => {
    initGcTrendChart()
    initMemoryTrendChart()
  })
  
  // 默认选择最近2小时
  const now = Date.now()
  timeRange.value = [now - 2 * 60 * 60 * 1000, now]
  
  // 加载数据
  if (props.jvmResourceId && timeRange.value) {
    loadGcSnapshots()
  }
})

onBeforeUnmount(() => {
  if (gcTrendChart) {
    gcTrendChart.dispose()
    gcTrendChart = null
  }
  if (memoryTrendChart) {
    memoryTrendChart.dispose()
    memoryTrendChart = null
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
// 样式文件
</style>

