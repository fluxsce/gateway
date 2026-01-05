<template>
  <div class="memory-pool-chart">
    <!-- 图表控制 -->
    <div class="chart-header">
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
    </div>

    <!-- 图表容器 -->
    <div class="chart-content">
      <div ref="chartRef" style="width: 100%; height: 400px"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { NButton, NIcon, NDatePicker, NSpace, useMessage } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import * as echarts from 'echarts/core'
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent
} from 'echarts/components'
import { LineChart } from 'echarts/charts'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'
import { isApiSuccess, getApiMessage, parseJsonData, formatDate } from '@/utils/format'
import * as api from '../api'
import type { MemoryPool } from '../types'

echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  LegendComponent,
  LineChart,
  CanvasRenderer
])

const props = defineProps<{
  jvmResourceId: string
}>()

const message = useMessage()

// 状态
const loading = ref(false)
const memoryPools = ref<MemoryPool[]>([])
const timeRange = ref<[number, number] | null>(null)

// 图表实例
const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null

// 时间范围快捷选项
const timeRangeShortcuts = {
  '最近1小时': () => {
    const now = Date.now()
    const start = now - 60 * 60 * 1000
    return [start, now] as [number, number]
  },
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
  '最近12小时': () => {
    const now = Date.now()
    const start = now - 12 * 60 * 60 * 1000
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

// 格式化内存大小
const formatMemorySize = (bytes: number): string => {
  if (bytes < 0) return 'N/A'
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
}

// 查询内存池数据
const queryMemoryPools = async () => {
  if (!props.jvmResourceId || !timeRange.value) return
  
  loading.value = true
  try {
    const [startTime, endTime] = timeRange.value
    const result = await api.queryMemoryPools({
      jvmResourceId: props.jvmResourceId,
      startTime: formatDate(startTime),
      endTime: formatDate(endTime)
    })
    
    if (isApiSuccess(result)) {
      const data = parseJsonData<MemoryPool[]>(result, [])
      memoryPools.value = (data && Array.isArray(data)) ? data : []
    } else {
      message.error(getApiMessage(result, '查询内存池列表失败'))
    }
  } catch (error) {
    console.error('查询内存池列表失败:', error)
    message.error('查询内存池列表失败')
  } finally {
    loading.value = false
  }
}

// 图表配置
const chartOptions = computed<EChartsOption>(() => {
  // 如果没有数据，返回最基本的空配置，避免显示"暂无数据"
  if (!memoryPools.value || memoryPools.value.length === 0) {
    return {
      xAxis: { type: 'category', data: [] },
      yAxis: { type: 'value' },
      series: []
    }
  }

  // 按内存池类型和时间分组
  // 结构: Map<poolType, Map<collectionTime, MemoryPool[]>>
  const poolTypeTimeMap = new Map<string, Map<string, MemoryPool[]>>()
  const allTimes = new Set<string>()
  
  memoryPools.value.forEach(pool => {
    let poolType = pool.poolName
    
    // 标准化内存池名称（相同类型的池会被聚合）
    if (poolType.toLowerCase().includes('eden')) {
      poolType = 'Eden Space'
    } else if (poolType.toLowerCase().includes('survivor')) {
      poolType = 'Survivor Space'
    } else if (poolType.toLowerCase().includes('old') && !poolType.includes('CodeHeap')) {
      poolType = 'Old Generation'
    } else if (poolType === 'Metaspace') {
      poolType = 'Metaspace'
    } else if (poolType === 'Compressed Class Space') {
      poolType = 'Compressed Class Space'
    } else if (poolType.includes('CodeHeap')) {
      if (poolType.includes('non-nmethods')) {
        poolType = 'CodeHeap (non-nmethods)'
      } else if (poolType.includes('profiled nmethods') || poolType.includes('profiled-nmethods')) {
        poolType = 'CodeHeap (profiled)'
      } else if (poolType.includes('non-profiled')) {
        poolType = 'CodeHeap (non-profiled)'
      } else {
        poolType = 'Code Cache'
      }
    }
    
    // 按类型和时间分组
    if (!poolTypeTimeMap.has(poolType)) {
      poolTypeTimeMap.set(poolType, new Map())
    }
    const timeMap = poolTypeTimeMap.get(poolType)!
    const collectionTime = pool.collectionTime
    
    if (!timeMap.has(collectionTime)) {
      timeMap.set(collectionTime, [])
    }
    timeMap.get(collectionTime)!.push(pool)
    allTimes.add(collectionTime)
  })

  // 如果没有分组，返回空配置
  if (poolTypeTimeMap.size === 0) {
    return {
      xAxis: { type: 'category', data: [] },
      yAxis: { type: 'value' },
      series: []
    }
  }

  // 排序时间并格式化
  const sortedTimes = Array.from(allTimes).sort()
  const formattedTimes = sortedTimes.map(time => {
    const date = new Date(time)
    return date.toLocaleString('zh-CN', { 
      month: '2-digit', 
      day: '2-digit', 
      hour: '2-digit', 
      minute: '2-digit',
      second: '2-digit'
    })
  })

  // 颜色方案
  const colors = [
    '#1890ff', // Eden
    '#52c41a', // Survivor
    '#faad14', // Old Gen
    '#722ed1', // Metaspace
    '#13c2c2', // Compressed Class
    '#f5222d', // CodeHeap 1
    '#eb2f96', // CodeHeap 2
    '#fa8c16', // CodeHeap 3
    '#2db7f5', // 其他
  ]

  // 为每种类型生成一条折线
  const series: any[] = []
  const poolTypeArray = Array.from(poolTypeTimeMap.entries())
  
  poolTypeArray.forEach(([typeName, timeMap], index) => {
    // 为每个时间点计算该类型的总使用量
    const data = sortedTimes.map(time => {
      const pools = timeMap.get(time) || []
      if (pools.length === 0) return null
      return pools.reduce((sum, p) => sum + p.currentUsedBytes, 0)
    })

    series.push({
      name: typeName,
      type: 'line' as const,
      data: data,
      smooth: true,
      symbol: 'circle',
      symbolSize: 6,
      connectNulls: true, // 连接空值
      lineStyle: {
        width: 2,
        color: colors[index % colors.length]
      },
      itemStyle: {
        color: colors[index % colors.length]
      },
      // 保存该类型在每个时间点的池数据
      poolsAtTime: Object.fromEntries(
        Array.from(timeMap.entries()).map(([time, pools]) => [time, pools])
      )
    })
  })

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross'
      },
      confine: false,
      enterable: true,
      appendToBody: true,
      position: function (point: any, params: any, dom: any, rect: any, size: any) {
        const chartWidth = size.viewSize[0]
        const chartHeight = size.viewSize[1]
        const tooltipWidth = size.contentSize[0]
        const tooltipHeight = size.contentSize[1]
        const x = point[0]
        const y = point[1]
        
        // 计算最佳位置，避免超出屏幕
        let posX = x + 15
        let posY = y - tooltipHeight / 2
        
        // 如果在右侧，tooltip 显示在左边
        if (x > chartWidth * 0.5) {
          posX = x - tooltipWidth - 20
        }
        
        // 确保 tooltip 不超出顶部和底部
        if (posY < 10) {
          posY = 10
        } else if (posY + tooltipHeight > chartHeight - 10) {
          posY = chartHeight - tooltipHeight - 10
        }
        
        return [posX, posY]
      },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        
        const timeIndex = params[0].dataIndex
        const time = sortedTimes[timeIndex]
        const formattedTime = params[0].name
        const formatSize = formatMemorySize
        
        let html = `
          <div style="max-height: 400px; max-width: 350px; overflow-y: auto; font-size: 12px; padding: 8px; background: white; border-radius: 4px; box-shadow: 0 2px 8px rgba(0,0,0,0.15);">
            <div style="font-weight: bold; margin-bottom: 10px; padding-bottom: 6px; border-bottom: 1px solid #eee; position: sticky; top: 0; background: white; z-index: 1;">${formattedTime}</div>
            <div style="display: grid; gap: 8px; padding-bottom: 4px;">
        `
        
        params.forEach((param: any) => {
          if (param.value === null) return
          
          const seriesData = series[param.seriesIndex]
          const poolsAtTime = seriesData.poolsAtTime[time] as MemoryPool[]
          if (!poolsAtTime || poolsAtTime.length === 0) return
          
          const usedBytes = param.value
          const totalCommittedBytes = poolsAtTime.reduce((sum, p) => sum + p.currentCommittedBytes, 0)
          const totalMaxBytes = poolsAtTime.reduce((sum, p) => sum + (p.currentMaxBytes > 0 ? p.currentMaxBytes : 0), 0)
          const usagePercent = totalCommittedBytes > 0 ? (usedBytes / totalCommittedBytes * 100) : 0
          
          html += `
            <div style="border-left: 4px solid ${param.color}; padding-left: 10px; background: #fafafa; padding: 8px; border-radius: 3px;">
              <div style="font-weight: 600; margin-bottom: 6px; color: #333;">${param.seriesName}</div>
              <div style="font-size: 11px; color: #666; line-height: 1.8;">
                <div><strong>使用量:</strong> ${formatSize(usedBytes)}</div>
                <div><strong>提交量:</strong> ${formatSize(totalCommittedBytes)}</div>
                ${totalMaxBytes > 0 ? `<div><strong>最大值:</strong> ${formatSize(totalMaxBytes)}</div>` : '<div><strong>最大值:</strong> 无限制</div>'}
                <div><strong>使用率:</strong> ${usagePercent.toFixed(1)}%</div>
              </div>
          `
          
          if (poolsAtTime.length > 1) {
            html += `<div style="margin-top: 6px; padding-top: 6px; border-top: 1px dashed #ddd; font-size: 10px; color: #999;">包含 ${poolsAtTime.length} 个内存池</div>`
          }
          
          html += '</div>'
        })
        
        html += `</div>
        </div>`
        return html
      }
    },
    legend: {
      type: 'scroll',
      bottom: 10,
      data: poolTypeArray.map(([name]) => name)
    },
    grid: {
      left: '10px',
      right: '30px',
      bottom: '80px',
      top: '30px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: formattedTimes,
      axisLabel: {
        fontSize: 12,
        rotate: 30,
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
      name: '内存使用量',
      nameTextStyle: {
        fontSize: 13,
        color: '#595959'
      },
      axisLabel: {
        fontSize: 13,
        color: '#595959',
        formatter: (value: number) => formatMemorySize(value)
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
      },
      min: 0
    },
    series: series
  }
})

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return
  
  // 检查 DOM 尺寸，如果为 0 则延迟初始化
  const { clientWidth, clientHeight } = chartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    // DOM 还未渲染完成，延迟初始化
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
    // 如果图表未初始化，尝试初始化
    initChart()
  } else {
    chartInstance.setOption(chartOptions.value, { notMerge: false })
  }
}

// 处理窗口大小变化
const handleResize = () => {
  if (chartInstance) {
    chartInstance.resize()
  }
}

// 处理时间范围变化
const handleTimeRangeChange = (range: [number, number] | null) => {
  timeRange.value = range
  if (range) {
    queryMemoryPools()
  }
}

// 处理刷新
const handleRefresh = () => {
  if (timeRange.value) {
    queryMemoryPools()
  }
}

// 监听数据变化
watch(() => memoryPools.value, () => {
  nextTick(() => {
    updateChart()
  })
}, { deep: true })

// 监听 jvmResourceId 变化
watch(() => props.jvmResourceId, (newId, oldId) => {
  if (newId && newId !== oldId) {
    // 重置状态
    memoryPools.value = []
    
    // 查询新的数据
    if (timeRange.value) {
      queryMemoryPools()
    }
  }
})

// 生命周期
onMounted(() => {
  nextTick(() => {
    initChart()
  })
  
  // 默认选择最近2小时
  const now = Date.now()
  timeRange.value = [now - 2 * 60 * 60 * 1000, now]
  
  // 初始加载数据
  if (props.jvmResourceId && timeRange.value) {
    queryMemoryPools()
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
.memory-pool-chart {
  .chart-header {
    display: flex;
    justify-content: flex-end;
    align-items: center;
    margin-bottom: 16px;
  }

  .chart-content {
    min-height: 400px;
  }
}
</style>

