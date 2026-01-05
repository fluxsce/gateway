<template>
  <div class="threadpool-monitor">
    <!-- 控制栏 -->
    <n-space style="margin-bottom: 16px" align="center">
      <n-select
        v-model:value="selectedPool"
        :options="poolOptions"
        :placeholder="t('selectComponent')"
        style="width: 300px"
      />
      <n-button @click="loadData" :loading="loading">
        <template #icon>
          <n-icon :component="RefreshOutline" />
          </template>
        {{ t('refresh') }}
      </n-button>
      <n-radio-group v-model:value="timeRange" @update:value="loadData">
        <n-radio-button value="15m">{{ t('last15Minutes') }}</n-radio-button>
        <n-radio-button value="1h">{{ t('last1Hour') }}</n-radio-button>
        <n-radio-button value="6h">{{ t('last6Hours') }}</n-radio-button>
        <n-radio-button value="24h">{{ t('last24Hours') }}</n-radio-button>
      </n-radio-group>
    </n-space>

    <!-- 图表 -->
    <div ref="chartRef" style="width: 100%; height: 400px"></div>

    <!-- 空数据提示 -->
    <n-empty v-if="!loading && chartData.length === 0" :description="t('noData')" style="margin-top: 60px" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { RefreshOutline } from '@vicons/ionicons5'
import { queryAppMonitorData } from '../api'
import { isApiSuccess, parseJsonData, formatDate } from '@/utils/format'
import type { AppMonitorData } from '../types'
import * as echarts from 'echarts'
import type { ECharts } from 'echarts'

const props = defineProps<{
  jvmResourceId: string
}>()

const { t } = useModuleI18n('hub0042')

const chartRef = ref<HTMLDivElement>()
const chartInstance = ref<ECharts>()
const loading = ref(false)
const chartData = ref<AppMonitorData[]>([])
const allPoolNames = ref<string[]>([])
const selectedPool = ref<string>()
const timeRange = ref('1h')

// 线程池选项
const poolOptions = computed(() => {
  return allPoolNames.value.map(name => ({
    label: name,
    value: name
  }))
})

// 计算时间范围
const getTimeRange = () => {
  const now = Date.now()
  const ranges: Record<string, number> = {
    '15m': 15 * 60 * 1000,
    '1h': 60 * 60 * 1000,
    '6h': 6 * 60 * 60 * 1000,
    '24h': 24 * 60 * 60 * 1000
  }
  const startTime = now - ranges[timeRange.value]
  return {
    startTime: formatDate(startTime),
    endTime: formatDate(now)
  }
}

// 加载数据
const loadData = async () => {
  if (!props.jvmResourceId) return

  loading.value = true
  try {
    const { startTime, endTime } = getTimeRange()
    const result = await queryAppMonitorData({
      jvmResourceId: props.jvmResourceId,
      dataType: 'THREAD_POOL',
      dataName: selectedPool.value,
      startTime,
      endTime
    })

    if (isApiSuccess(result)) {
      const allData = parseJsonData<AppMonitorData[]>(result, [])
      chartData.value = allData.filter(d => d.dataType === 'THREAD_POOL')
      
      // 更新线程池名称列表（只在首次加载或没有选中时更新）
      if (!selectedPool.value) {
        const uniqueNames = [...new Set(chartData.value.map(d => d.dataName))]
        allPoolNames.value = uniqueNames
        if (uniqueNames.length > 0) {
          selectedPool.value = uniqueNames[0]
        }
      }

      await nextTick()
      renderChart()
    }
  } catch (error) {
    console.error('加载线程池数据失败:', error)
  } finally {
    loading.value = false
  }
}

// 渲染图表
const renderChart = () => {
  if (!chartRef.value || !selectedPool.value) return

  if (!chartInstance.value) {
    chartInstance.value = echarts.init(chartRef.value)
  }

  const poolData = chartData.value
    .filter(d => d.dataName === selectedPool.value)
    .sort((a, b) => new Date(a.collectionTime).getTime() - new Date(b.collectionTime).getTime())

  if (poolData.length === 0) {
    chartInstance.value.clear()
    return
  }

  const times = poolData.map(d => new Date(d.collectionTime).toLocaleTimeString())
  const rawData = poolData.map(d => {
    try {
      return JSON.parse(d.dataJson)
    } catch {
      return {}
    }
  })

  // 获取健康状态的颜色和文本
  const getHealthStatus = (healthyFlag: string) => {
    if (healthyFlag === 'Y') {
      return { color: '#18a058', text: '健康', icon: '●' }
    } else if (healthyFlag === 'N') {
      return { color: '#f5222d', text: '异常', icon: '●' }
    } else {
      return { color: '#faad14', text: '警告', icon: '●' }
    }
  }

  const option: echarts.EChartsOption = {
    tooltip: {
      trigger: 'item',
      formatter: (params: any) => {
        const dataIndex = params.dataIndex
        const data = rawData[dataIndex]
        const item = poolData[dataIndex]
        const seriesName = params.seriesName
        const value = params.value
        const healthStatus = getHealthStatus(item.healthyFlag)
        
        return `
          <div style="font-size: 12px;">
            <div style="font-weight: bold; margin-bottom: 6px;">
              ${params.name}
              <span style="margin-left: 8px; color: ${healthStatus.color};">
                ${healthStatus.icon} ${healthStatus.text}
              </span>
            </div>
            <div style="margin: 3px 0;">
              <span style="display: inline-block; margin-right: 4px; width: 10px; height: 10px; background-color: ${params.color}; border-radius: 50%;"></span>
              <span>${seriesName}: <span style="font-weight: bold;">${value}</span></span>
            </div>
            <div style="color: #666; margin-top: 8px; padding-top: 8px; border-top: 1px solid #e8e8e8;">
              <div style="margin: 3px 0;">${t('active')}: ${data.activeCount || 0}</div>
              <div style="margin: 3px 0;">${t('current')}: ${data.poolSize || 0}</div>
              <div style="margin: 3px 0;">队列大小: ${data.queueSize || 0}</div>
              <div style="margin: 3px 0;">核心线程数: ${data.corePoolSize || 0}</div>
              <div style="margin: 3px 0;">最大线程数: ${data.maximumPoolSize || 0}</div>
              <div style="margin: 3px 0;">队列容量: ${data.queueCapacity || 0}</div>
              <div style="margin: 3px 0;">已完成任务: ${data.completedTaskCount || 0}</div>
            </div>
          </div>
        `
      }
    },
    legend: {
      data: [t('active'), t('current'), '队列大小'],
      bottom: 0
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '60px',
      top: '10px',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: times,
      axisLabel: {
        rotate: 45,
        fontSize: 11
      }
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: t('active'),
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: (value: any, params: any) => {
          const item = poolData[params.dataIndex]
          return item.healthyFlag === 'N' ? 8 : 6
        },
        lineStyle: { width: 2, color: '#18a058' },
        itemStyle: {
          color: (params: any) => {
            const item = poolData[params.dataIndex]
            return item.healthyFlag === 'N' ? '#f5222d' : '#18a058'
          }
        },
        data: rawData.map(d => d.activeCount || 0)
      },
      {
        name: t('current'),
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: (value: any, params: any) => {
          const item = poolData[params.dataIndex]
          return item.healthyFlag === 'N' ? 8 : 6
        },
        lineStyle: { width: 2, color: '#2080f0' },
        itemStyle: {
          color: (params: any) => {
            const item = poolData[params.dataIndex]
            return item.healthyFlag === 'N' ? '#f5222d' : '#2080f0'
          }
        },
        data: rawData.map(d => d.poolSize || 0)
      },
      {
        name: '队列大小',
        type: 'line',
        smooth: true,
        symbol: 'circle',
        symbolSize: (value: any, params: any) => {
          const item = poolData[params.dataIndex]
          return item.healthyFlag === 'N' ? 8 : 6
        },
        lineStyle: { width: 2, color: '#faad14' },
        itemStyle: {
          color: (params: any) => {
            const item = poolData[params.dataIndex]
            return item.healthyFlag === 'N' ? '#f5222d' : '#faad14'
          }
        },
        data: rawData.map(d => d.queueSize || 0)
      }
    ]
  }

  chartInstance.value.setOption(option)
}

watch(selectedPool, loadData)
watch(() => props.jvmResourceId, loadData, { immediate: true })

const handleResize = () => chartInstance.value?.resize()

onMounted(() => window.addEventListener('resize', handleResize))
onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  chartInstance.value?.dispose()
})
</script>

<style scoped lang="scss">
.threadpool-monitor {
  width: 100%;
}
</style>
