<template>
  <div class="heap-memory-chart">
    <!-- 当前值概览 -->
    <div v-if="currentMemory" class="memory-overview">
      <n-space :size="24">
        <div class="memory-stat">
          <div class="stat-label">当前使用率</div>
          <div class="stat-value" :style="{ color: getStatusColor(currentMemory.usagePercent) }">
            {{ currentMemory.usagePercent.toFixed(2) }}%
          </div>
        </div>
        <n-divider vertical />
        <div class="memory-stat">
          <div class="stat-label">已使用</div>
          <div class="stat-value">{{ formatMemorySize(currentMemory.usedMemoryBytes) }}</div>
        </div>
        <n-divider vertical />
        <div class="memory-stat">
          <div class="stat-label">已提交</div>
          <div class="stat-value">{{ formatMemorySize(currentMemory.committedMemoryBytes) }}</div>
        </div>
        <n-divider vertical />
        <div class="memory-stat">
          <div class="stat-label">最大值</div>
          <div class="stat-value">
            {{ currentMemory.maxMemoryBytes > 0 
              ? formatMemorySize(currentMemory.maxMemoryBytes)
              : '无限制' 
            }}
          </div>
        </div>
      </n-space>
    </div>

    <!-- 图表控制 -->
    <div class="chart-header">
      <div class="chart-controls">
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
        </n-button>
      </div>
    </div>

    <!-- 图表内容 -->
    <div class="chart-content">
      <div ref="chartRef" style="width: 100%; height: 300px"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { NButton, NIcon, NDatePicker, NSpace, NDivider, useMessage } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { isApiSuccess, getApiMessage, parseJsonData, formatDate } from '@/utils/format'
import * as echarts from 'echarts/core'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent
} from 'echarts/components'
import { LineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'
import * as api from '../api'
import type { JvmMemory } from '../types'

// 注册 ECharts 组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  LineChart,
  CanvasRenderer
])

const props = defineProps<{
  jvmResourceId: string
}>()

const { t } = useModuleI18n('hub0042')
const message = useMessage()

// 状态
const loading = ref(false)
const trendData = ref<any[]>([])
const timeRange = ref<[number, number] | null>(null)

// 图表实例
const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null

// 当前内存信息（从趋势数据的最后一条记录获取）
const currentMemory = computed(() => {
  if (trendData.value.length === 0) return null
  const lastData = trendData.value[trendData.value.length - 1]
  return {
    usagePercent: lastData.value,
    usedMemoryBytes: lastData.usedMemoryBytes,
    committedMemoryBytes: lastData.committedMemoryBytes,
    maxMemoryBytes: lastData.maxMemoryBytes || -1,
    initMemoryBytes: lastData.initMemoryBytes || 0
  }
})

// 时间范围快捷选项
const timeRangeShortcuts = {
  '最近2小时': () => {
    const now = Date.now()
    const start = now - 2 * 60 * 60 * 1000
    return [start, now] as [number, number]
  },
  '最近6小时': () => {
    const now = Date.now()
    const start = now - 6 * 60 * 60 * 1000
    return [start, now] as [number, number]
  },
  '最近24小时': () => {
    const now = Date.now()
    const start = now - 24 * 60 * 60 * 1000
    return [start, now] as [number, number]
  },
  '最近7天': () => {
    const now = Date.now()
    const start = now - 7 * 24 * 60 * 60 * 1000
    return [start, now] as [number, number]
  }
}

// 查询内存趋势数据
const queryTrendData = async () => {
  if (!props.jvmResourceId || !timeRange.value) return
  
  loading.value = true
  try {
    const [startTime, endTime] = timeRange.value
    const result = await api.queryMemoryUsage({
      jvmResourceId: props.jvmResourceId,
      memoryType: 'HEAP' as any,
      startTime: formatDate(startTime),
      endTime: formatDate(endTime)
    })
    
    if (isApiSuccess(result)) {
      const data = parseJsonData<JvmMemory[]>(result, [])
      if (data && Array.isArray(data)) {
        trendData.value = data.map(item => ({
          time: item.collectionTime,
          value: item.usagePercent,
          usedMemoryBytes: item.usedMemoryBytes,
          committedMemoryBytes: item.committedMemoryBytes,
          maxMemoryBytes: item.maxMemoryBytes,
          initMemoryBytes: item.initMemoryBytes
        }))
      }
    } else {
      message.error(getApiMessage(result, '查询堆内存趋势数据失败'))
    }
  } catch (error) {
    console.error('查询堆内存趋势数据失败:', error)
    message.error('查询堆内存趋势数据失败')
  } finally {
    loading.value = false
  }
}

// 格式化内存大小
const formatMemorySize = (bytes: number): string => {
  if (bytes < 0) return 'N/A'
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

// 获取状态颜色
const getStatusColor = (percent: number): string => {
  if (percent >= 90) return '#ff4d4f' // 危险-红色
  if (percent >= 75) return '#faad14' // 警告-橙色
  return '#52c41a' // 正常-绿色
}

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
  const times = trendData.value.map(item => {
    const date = new Date(item.time)
    return date.toLocaleString('zh-CN', { 
      month: '2-digit', 
      day: '2-digit', 
      hour: '2-digit', 
      minute: '2-digit' 
    })
  })
  
  const values = trendData.value.map(item => item.value)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        
        const dataIndex = params[0].dataIndex
        const originalData = trendData.value[dataIndex]
        const time = params[0].name
        const usagePercent = originalData.value
        const usedMemory = originalData.usedMemoryBytes || 0
        const committedMemory = originalData.committedMemoryBytes || 0
        
        return `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 6px;">${time}</div>
            <div style="margin-bottom: 3px;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: #1890ff; border-radius: 50%;"></span>
              <span>使用率: ${usagePercent.toFixed(2)}%</span>
            </div>
            <div style="margin-bottom: 3px;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: #1890ff; border-radius: 50%; opacity: 0.6;"></span>
              <span>使用量: ${formatMemorySize(usedMemory)}</span>
            </div>
            <div>
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: #faad14; border-radius: 50%;"></span>
              <span>提交量: ${formatMemorySize(committedMemory)}</span>
            </div>
          </div>
        `
      }
    },
    grid: {
      left: '10px',
      right: '30px',
      bottom: '50px',
      top: '30px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        fontSize: 13,
        rotate: 0,
        color: '#595959'
      },
      axisLine: {
        lineStyle: {
          color: '#d9d9d9'
        }
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        fontSize: 13,
        color: '#595959',
        formatter: (value: number) => `${value}%`
      },
      axisLine: {
        lineStyle: {
          color: '#d9d9d9'
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
        name: t('heapMemory'),
        type: 'line',
        data: values,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: {
          color: '#1890ff',
          width: 2
        },
        itemStyle: {
          color: '#1890ff'
        },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              {
                offset: 0,
                color: '#1890ff40'
              },
              {
                offset: 1,
                color: '#1890ff10'
              }
            ]
          }
        },
        markLine: {
          silent: true,
          symbol: 'none',
          data: [
            {
              yAxis: 75,
              lineStyle: {
                color: '#faad14',
                type: 'dashed' as const,
                width: 1
              },
              label: {
                formatter: '警告: 75%',
                fontSize: 10
              }
            },
            {
              yAxis: 90,
              lineStyle: {
                color: '#ff4d4f',
                type: 'dashed' as const,
                width: 1
              },
              label: {
                formatter: '危险: 90%',
                fontSize: 10
              }
            }
          ]
        }
      }
    ]
  }
})

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return
  
  // 检查 DOM 尺寸，如果为 0 则延迟初始化
  const { clientWidth, clientHeight } = chartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initChart(), 100)
    return
  }
  
  // 如果已经初始化，先销毁
  if (chartInstance) {
    chartInstance.dispose()
  }
  
  chartInstance = echarts.init(chartRef.value)
  chartInstance.setOption(chartOptions.value)
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
}

// 更新图表
const updateChart = () => {
  if (!chartInstance) {
    initChart()
  } else {
    chartInstance.setOption(chartOptions.value)
  }
}

// 处理窗口大小变化
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 刷新
const handleRefresh = async () => {
  if (timeRange.value) {
    await queryTrendData()
  }
}

// 时间范围变化
const handleTimeRangeChange = (range: [number, number] | null) => {
  timeRange.value = range
  if (range) {
    queryTrendData()
  }
}

// 监听 jvmResourceId 变化
watch(() => props.jvmResourceId, async (newId, oldId) => {
  if (newId && newId !== oldId) {
    // 重置状态
    trendData.value = []
    
    // 查询新的数据
    if (timeRange.value) {
      await queryTrendData()
    }
  }
})

// 监听 trendData 变化
watch(() => trendData.value, () => {
  nextTick(() => {
    updateChart()
  })
}, { deep: true })

onMounted(() => {
  nextTick(() => {
    initChart()
  })
  
  // 默认选择最近2小时
  const now = Date.now()
  timeRange.value = [now - 2 * 60 * 60 * 1000, now]
  
  // 初始加载数据
  if (props.jvmResourceId && timeRange.value) {
    queryTrendData()
  }
})

onBeforeUnmount(() => {
  if (chartInstance) {
    chartInstance.dispose()
    chartInstance = null
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.heap-memory-chart {
  .memory-overview {
    padding: 16px;
    background: #fafafa;
    border-radius: 4px;
    margin-bottom: 16px;

    .memory-stat {
      text-align: center;

      .stat-label {
        font-size: 12px;
        color: #999;
        margin-bottom: 4px;
      }

      .stat-value {
        font-size: 18px;
        font-weight: 600;
        color: #262626;
      }
    }
  }

  .chart-header {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    margin-bottom: 16px;
  }

  .chart-controls {
    display: flex;
    gap: 12px;
    align-items: center;
  }

  .chart-content {
    min-height: 300px;
  }
}
</style>

