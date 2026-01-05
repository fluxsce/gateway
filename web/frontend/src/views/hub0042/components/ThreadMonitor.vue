<template>
  <div class="thread-monitor">
    <!-- 时间控制 -->
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
          {{ t('refresh') }}
        </n-button>
      </n-space>
    </n-card>

    <!-- 线程数量趋势图 -->
    <n-card :title="t('threadCountTrend')" :bordered="false" style="margin-top: 16px">
      <div ref="threadCountChartRef" style="width: 100%; height: 400px"></div>
    </n-card>

    <!-- 线程状态分布趋势图 -->
    <n-card :title="t('threadStateTrend')" :bordered="false" style="margin-top: 16px">
      <div ref="threadStateChartRef" style="width: 100%; height: 400px"></div>
    </n-card>

    <!-- 线程比例趋势图 -->
    <n-card :title="t('threadRatioTrend')" :bordered="false" style="margin-top: 16px">
      <div ref="threadRatioChartRef" style="width: 100%; height: 400px"></div>
    </n-card>

    <!-- 死锁检测趋势 -->
    <n-card :title="t('deadlockDetection')" :bordered="false" style="margin-top: 16px">
      <!-- 当前状态概览 -->
      <div v-if="deadlocks && deadlocks.length > 0" class="deadlock-overview">
        <div class="overview-grid">
          <div class="stat-card">
            <div class="stat-label">{{ t('totalDetections') }}</div>
            <div class="stat-value">{{ deadlocks.length }}</div>
            <div class="stat-suffix">{{ t('times') }}</div>
          </div>
          
          <div class="stat-card">
            <div class="stat-label">{{ t('deadlockOccurrences') }}</div>
            <div class="stat-value" :class="{ 'danger': deadlockOccurrences > 0, 'safe': deadlockOccurrences === 0 }">
              {{ deadlockOccurrences }}
            </div>
            <div class="stat-suffix">{{ t('times') }}</div>
          </div>
          
          <div class="status-card">
            <div class="status-label">{{ t('currentStatus') }}</div>
            <div class="status-badge" :class="{ 'error': hasDeadlock, 'success': !hasDeadlock }">
              <n-icon size="16" class="status-icon">
                <svg v-if="hasDeadlock" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10s10-4.48 10-10S17.52 2 12 2M13 17h-2v-6h2zm0-8h-2V7h2z"/>
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10s10-4.5 10-10S17.5 2 12 2m-2 15l-5-5l1.41-1.41L10 14.17l7.59-7.59L19 8z"/>
                </svg>
              </n-icon>
              <span class="status-text">
                {{ hasDeadlock ? t('deadlockDetected') : t('noDeadlock') }}
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 死锁趋势图表 -->
      <div v-if="deadlocks && deadlocks.length > 0" style="margin-top: 16px">
        <div ref="deadlockChartRef" style="width: 100%; height: 300px"></div>
      </div>
      
      <!-- 无数据时的成功状态 -->
      <n-result 
        v-else
        status="success" 
        :title="t('noDeadlock')" 
        :description="t('noDeadlockDesc')"
      >
        <template #icon>
          <n-icon size="60" color="#52c41a">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
              <path fill="currentColor" d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10s10-4.5 10-10S17.5 2 12 2m-2 15l-5-5l1.41-1.41L10 14.17l7.59-7.59L19 8z"/>
            </svg>
          </n-icon>
        </template>
      </n-result>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { NButton, NIcon, NDatePicker, NSpace } from 'naive-ui'
import { RefreshOutline } from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { formatDate } from '@/utils/format'
import { useThreadMonitor } from '../hooks'
import type { SeverityLevel, AlertLevel } from '../types'
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
  threadInfo,
  deadlocks,
  hasDeadlock,
  threadTrendData,
  threadStateTrendData,
  queryThreadInfo,
  getThreadState,
  queryThreadStates,
  queryDeadlocks
} = useThreadMonitor()

// 状态
const timeRange = ref<[number, number] | null>(null)

// 图表实例
const threadCountChartRef = ref<HTMLDivElement>()
const threadStateChartRef = ref<HTMLDivElement>()
const threadRatioChartRef = ref<HTMLDivElement>()
const deadlockChartRef = ref<HTMLDivElement>()
let threadCountChart: echarts.ECharts | null = null
let threadStateChart: echarts.ECharts | null = null
let threadRatioChart: echarts.ECharts | null = null
let deadlockChart: echarts.ECharts | null = null

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

// 获取严重程度类型
const getSeverityType = (level?: SeverityLevel): 'default' | 'info' | 'warning' | 'error' => {
  if (!level) return 'default'
  const map: Record<SeverityLevel, 'default' | 'info' | 'warning' | 'error'> = {
    LOW: 'info',
    MEDIUM: 'warning',
    HIGH: 'error',
    CRITICAL: 'error'
  }
  return map[level] || 'default'
}

// 获取告警级别类型
const getAlertType = (level?: AlertLevel): 'default' | 'info' | 'warning' | 'error' => {
  if (!level) return 'default'
  const map: Record<AlertLevel, 'default' | 'info' | 'warning' | 'error'> = {
    INFO: 'info',
    WARNING: 'warning',
    ERROR: 'error',
    CRITICAL: 'error',
    EMERGENCY: 'error'
  }
  return map[level] || 'default'
}

// 格式化时间
const formatTime = (timeStr: string): string => {
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', { 
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit'
  })
}

// 格式化死锁检测时间（更详细）
const formatDeadlockTime = (timeStr?: string): string => {
  if (!timeStr) return '-'
  const date = new Date(timeStr)
  return date.toLocaleString('zh-CN', { 
    year: 'numeric',
    month: '2-digit', 
    day: '2-digit', 
    hour: '2-digit', 
    minute: '2-digit',
    second: '2-digit'
  })
}

// 计算死锁发生次数
const deadlockOccurrences = computed(() => {
  if (!deadlocks.value || !Array.isArray(deadlocks.value)) {
    return 0
  }
  return deadlocks.value.filter(d => d?.hasDeadlockFlag === 'Y').length
})

// 线程数量趋势图表配置
const threadCountOptions = computed<EChartsOption>(() => {
  if (!threadTrendData.value || threadTrendData.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = threadTrendData.value.map(item => formatTime(item?.time || ''))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' }
    },
    legend: {
      data: ['当前线程数', '峰值线程数', '守护线程数', '用户线程数'],
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
    yAxis: {
      type: 'value',
      name: '线程数',
      axisLabel: {
        fontSize: 13,
        color: '#595959'
      }
    },
    series: [
      {
        name: '当前线程数',
        type: 'line',
        data: threadTrendData.value.map(item => item?.currentThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#1890ff', width: 2 },
        itemStyle: { color: '#1890ff' }
      },
      {
        name: '峰值线程数',
        type: 'line',
        data: threadTrendData.value.map(item => item?.peakThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#ff4d4f', width: 2, type: 'dashed' },
        itemStyle: { color: '#ff4d4f' }
      },
      {
        name: '守护线程数',
        type: 'line',
        data: threadTrendData.value.map(item => item?.daemonThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#52c41a', width: 2 },
        itemStyle: { color: '#52c41a' }
      },
      {
        name: '用户线程数',
        type: 'line',
        data: threadTrendData.value.map(item => item?.userThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#faad14', width: 2 },
        itemStyle: { color: '#faad14' }
      }
    ]
  }
})

// 线程状态分布趋势图表配置
const threadStateOptions = computed<EChartsOption>(() => {
  if (!threadStateTrendData.value || threadStateTrendData.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = threadStateTrendData.value.map(item => formatTime(item?.time || ''))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any) => {
        try {
          if (!Array.isArray(params) || params.length === 0) return ''
          const dataIndex = params[0].dataIndex
          
          // 安全检查数据是否存在
          if (!threadStateTrendData.value || dataIndex >= threadStateTrendData.value.length || dataIndex < 0) {
            return `<div style="font-size: 12px;">数据加载中...</div>`
          }
          
          const data = threadStateTrendData.value[dataIndex]
          if (!data) {
            return `<div style="font-size: 12px;">数据不可用</div>`
          }
          
          let html = `<div style="font-size: 12px;"><strong>${params[0].name}</strong><br/>`
          
          // 安全获取数据，提供默认值
          const totalThreadCount = data.totalThreadCount || 0
          const runnableThreadCount = data.runnableThreadCount || 0
          const timedWaitingThreadCount = data.timedWaitingThreadCount || 0
          const waitingThreadCount = data.waitingThreadCount || 0
          const blockedThreadCount = data.blockedThreadCount || 0
          const newThreadCount = data.newThreadCount || 0
          const terminatedThreadCount = data.terminatedThreadCount || 0
          
          // 按照重要性排序显示
          html += `<span style="color: #1890ff;">◆</span> 总线程数: <strong>${totalThreadCount}</strong><br/>`
          html += `<span style="color: #52c41a;">●</span> RUNNABLE: ${runnableThreadCount}<br/>`
          html += `<span style="color: #fa8c16;">●</span> TIMED_WAITING: ${timedWaitingThreadCount}<br/>`
          html += `<span style="color: #faad14;">●</span> WAITING: ${waitingThreadCount}<br/>`
          html += `<span style="color: #ff4d4f;">●</span> BLOCKED: ${blockedThreadCount}<br/>`
          html += `<span style="color: #d9d9d9;">●</span> NEW: ${newThreadCount}<br/>`
          html += `<span style="color: #8c8c8c;">●</span> TERMINATED: ${terminatedThreadCount}<br/>`
          
          // 验证数据完整性
          const sum = newThreadCount + runnableThreadCount + blockedThreadCount + 
                     waitingThreadCount + timedWaitingThreadCount + terminatedThreadCount
          if (sum !== totalThreadCount && totalThreadCount > 0) {
            html += `<span style="color: #ff4d4f; font-size: 11px;">⚠️ 数据不一致: ${sum} ≠ ${totalThreadCount}</span>`
          }
          
          html += '</div>'
          return html
        } catch (error) {
          console.error('Tooltip formatter error:', error)
          return `<div style="font-size: 12px;">显示错误</div>`
        }
      }
    },
    legend: {
      data: ['NEW', 'RUNNABLE', 'BLOCKED', 'WAITING', 'TIMED_WAITING', 'TERMINATED', '总线程数'],
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
    yAxis: {
      type: 'value',
      name: '线程数',
      axisLabel: {
        fontSize: 13,
        color: '#595959'
      }
    },
    series: [
      {
        name: 'NEW',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.newThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#d9d9d9', width: 2 },
        itemStyle: { color: '#d9d9d9' },
        symbol: 'circle',
        symbolSize: 4
      },
      {
        name: 'RUNNABLE',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.runnableThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#52c41a', width: 2 },
        itemStyle: { color: '#52c41a' },
        symbol: 'circle',
        symbolSize: 4
      },
      {
        name: 'BLOCKED',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.blockedThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#ff4d4f', width: 2 },
        itemStyle: { color: '#ff4d4f' },
        symbol: 'circle',
        symbolSize: 4
      },
      {
        name: 'WAITING',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.waitingThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#faad14', width: 2 },
        itemStyle: { color: '#faad14' },
        symbol: 'circle',
        symbolSize: 4
      },
      {
        name: 'TIMED_WAITING',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.timedWaitingThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#fa8c16', width: 2 },
        itemStyle: { color: '#fa8c16' },
        symbol: 'circle',
        symbolSize: 4
      },
      {
        name: 'TERMINATED',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.terminatedThreadCount || 0),
        smooth: true,
        lineStyle: { color: '#8c8c8c', width: 2 },
        itemStyle: { color: '#8c8c8c' },
        symbol: 'circle',
        symbolSize: 4
      },
      // 添加总线程数参考线
      {
        name: '总线程数',
        type: 'line',
        data: threadStateTrendData.value.map(item => item?.totalThreadCount || 0),
        smooth: true,
        lineStyle: { 
          color: '#1890ff', 
          width: 3,
          type: 'dashed'
        },
        itemStyle: { color: '#1890ff' },
        symbol: 'diamond',
        symbolSize: 6,
        z: 10
      }
    ]
  }
})

// 线程比例趋势图表配置
const threadRatioOptions = computed<EChartsOption>(() => {
  if (!threadTrendData.value || threadTrendData.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = threadTrendData.value.map(item => formatTime(item?.time || ''))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        let html = `<div style="font-size: 12px;"><strong>${params[0].name}</strong><br/>`
        params.forEach((param: any) => {
          html += `<span style="color: ${param.color};">●</span> ${param.seriesName}: ${param.value.toFixed(2)}%<br/>`
        })
        html += '</div>'
        return html
      }
    },
    legend: {
      data: ['活跃线程比例', '阻塞线程比例', '等待线程比例'],
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
    yAxis: {
      type: 'value',
      name: '比例(%)',
      max: 100,
      axisLabel: {
        fontSize: 13,
        color: '#595959',
        formatter: (value: number) => `${value}%`
      }
    },
    series: [
      {
        name: '活跃线程比例',
        type: 'line',
        data: threadTrendData.value.map(item => item?.activeThreadRatioPercent || 0),
        smooth: true,
        lineStyle: { color: '#52c41a', width: 2 },
        itemStyle: { color: '#52c41a' },
        areaStyle: { 
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#52c41a40' },
              { offset: 1, color: '#52c41a10' }
            ]
          }
        }
      },
      {
        name: '阻塞线程比例',
        type: 'line',
        data: threadTrendData.value.map(item => item?.blockedThreadRatioPercent || 0),
        smooth: true,
        lineStyle: { color: '#ff4d4f', width: 2 },
        itemStyle: { color: '#ff4d4f' },
        areaStyle: { 
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#ff4d4f40' },
              { offset: 1, color: '#ff4d4f10' }
            ]
          }
        }
      },
      {
        name: '等待线程比例',
        type: 'line',
        data: threadTrendData.value.map(item => item?.waitingThreadRatioPercent || 0),
        smooth: true,
        lineStyle: { color: '#faad14', width: 2 },
        itemStyle: { color: '#faad14' },
        areaStyle: { 
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#faad1440' },
              { offset: 1, color: '#faad1410' }
            ]
          }
        }
      }
    ]
  }
})

// 死锁趋势图表配置
const deadlockOptions = computed<EChartsOption>(() => {
  if (!deadlocks.value || !Array.isArray(deadlocks.value) || deadlocks.value.length === 0) {
    return { xAxis: { type: 'category', data: [] }, yAxis: { type: 'value' }, series: [] }
  }

  const times = deadlocks.value.map(item => formatTime(item?.collectionTime || item?.detectionTime || ''))
  
  // 将死锁标记转换为数值：Y=1（有死锁），N=0（无死锁）
  const deadlockStatus = deadlocks.value.map(item => item?.hasDeadlockFlag === 'Y' ? 1 : 0)
  const deadlockThreadCounts = deadlocks.value.map(item => item?.deadlockThreadCount || 0)

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params: any) => {
        try {
          if (!Array.isArray(params) || params.length === 0) return ''
          const dataIndex = params[0].dataIndex
          
          // 安全检查数据是否存在
          if (!deadlocks.value || !Array.isArray(deadlocks.value) || dataIndex >= deadlocks.value.length || dataIndex < 0) {
            return `<div style="font-size: 12px;">数据加载中...</div>`
          }
          
          const deadlock = deadlocks.value[dataIndex]
          if (!deadlock) {
            return `<div style="font-size: 12px;">数据不可用</div>`
          }
          
          const time = params[0].name
          const hasDeadlock = deadlock.hasDeadlockFlag === 'Y'
          const status = hasDeadlock ? '存在死锁' : '无死锁'
          const threadCount = deadlock.deadlockThreadCount || 0
          
          let html = `<div style="font-size: 12px;"><strong>${time}</strong><br/>`
          html += `<span style="color: ${hasDeadlock ? '#ff4d4f' : '#52c41a'};">●</span> 状态: ${status}<br/>`
          if (threadCount > 0) {
            html += `<span style="color: #faad14;">●</span> 涉及线程: ${threadCount}个<br/>`
          }
          if (deadlock.severityLevel) {
            html += `<span>严重程度: ${deadlock.severityLevel}</span><br/>`
          }
          html += '</div>'
          return html
        } catch (error) {
          console.error('Deadlock tooltip formatter error:', error)
          return `<div style="font-size: 12px;">显示错误</div>`
        }
      }
    },
    legend: {
      data: ['死锁状态', '涉及线程数'],
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
        name: '死锁状态',
        position: 'left',
        min: 0,
        max: 1,
        interval: 1,
        axisLabel: {
          fontSize: 13,
          color: '#595959',
          formatter: (value: number) => value === 1 ? '有死锁' : '无死锁'
        }
      },
      {
        type: 'value',
        name: '涉及线程数',
        position: 'right',
        axisLabel: {
          fontSize: 13,
          color: '#595959'
        }
      }
    ],
    series: [
      {
        name: '死锁状态',
        type: 'line',
        data: deadlockStatus,
        smooth: false,
        step: 'end',
        lineStyle: { 
          color: '#ff4d4f', 
          width: 3
        },
        itemStyle: { 
          color: (params: any) => {
            return params.value === 1 ? '#ff4d4f' : '#52c41a'
          }
        },
        areaStyle: { 
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#ff4d4f40' },
              { offset: 1, color: '#ff4d4f10' }
            ]
          }
        },
        markPoint: {
          data: deadlocks.value.map((item, index) => {
            if (item?.hasDeadlockFlag === 'Y') {
              return {
                coord: [index, 1],
                symbol: 'pin',
                symbolSize: 50,
                itemStyle: { color: '#ff4d4f' }
              }
            }
            return null
          }).filter(Boolean) as any[]
        }
      },
      {
        name: '涉及线程数',
        type: 'bar',
        yAxisIndex: 1,
        data: deadlockThreadCounts,
        barWidth: '40%',
        itemStyle: { 
          color: '#faad14',
          borderRadius: [4, 4, 0, 0]
        }
      }
    ]
  }
})

// 初始化图表
const initThreadCountChart = () => {
  if (!threadCountChartRef.value) return
  
  const { clientWidth, clientHeight } = threadCountChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initThreadCountChart(), 100)
    return
  }
  
  if (threadCountChart) {
    threadCountChart.dispose()
  }
  
  threadCountChart = echarts.init(threadCountChartRef.value)
  threadCountChart.setOption(threadCountOptions.value)
}

const initThreadStateChart = () => {
  if (!threadStateChartRef.value) return
  
  const { clientWidth, clientHeight } = threadStateChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initThreadStateChart(), 100)
    return
  }
  
  if (threadStateChart) {
    threadStateChart.dispose()
  }
  
  threadStateChart = echarts.init(threadStateChartRef.value)
  threadStateChart.setOption(threadStateOptions.value)
}

const initThreadRatioChart = () => {
  if (!threadRatioChartRef.value) return
  
  const { clientWidth, clientHeight } = threadRatioChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initThreadRatioChart(), 100)
    return
  }
  
  if (threadRatioChart) {
    threadRatioChart.dispose()
  }
  
  threadRatioChart = echarts.init(threadRatioChartRef.value)
  threadRatioChart.setOption(threadRatioOptions.value)
}

const initDeadlockChart = () => {
  if (!deadlockChartRef.value) return
  
  const { clientWidth, clientHeight } = deadlockChartRef.value
  if (clientWidth === 0 || clientHeight === 0) {
    setTimeout(() => initDeadlockChart(), 100)
    return
  }
  
  if (deadlockChart) {
    deadlockChart.dispose()
  }
  
  deadlockChart = echarts.init(deadlockChartRef.value)
  deadlockChart.setOption(deadlockOptions.value)
}

// 更新图表
const updateCharts = () => {
  if (threadCountChart) {
    threadCountChart.setOption(threadCountOptions.value)
  }
  if (threadStateChart) {
    threadStateChart.setOption(threadStateOptions.value)
  }
  if (threadRatioChart) {
    threadRatioChart.setOption(threadRatioOptions.value)
  }
  if (deadlockChart) {
    deadlockChart.setOption(deadlockOptions.value)
  }
}

// 处理窗口大小变化
const handleResize = () => {
  if (threadCountChart) {
    threadCountChart.resize()
  }
  if (threadStateChart) {
    threadStateChart.resize()
  }
  if (threadRatioChart) {
    threadRatioChart.resize()
  }
  if (deadlockChart) {
    deadlockChart.resize()
  }
}

// 加载数据
const loadData = async () => {
  if (!props.jvmResourceId) return
  
  const params: any = { jvmResourceId: props.jvmResourceId }
  
  // 如果有时间范围，添加到参数中
  if (timeRange.value) {
    const [startTime, endTime] = timeRange.value
    params.startTime = formatDate(startTime)
    params.endTime = formatDate(endTime)
  }
  
  // 并行查询：线程信息、线程状态列表、死锁信息
  if (timeRange.value) {
    // 有时间范围：批量查询
    await Promise.all([
      queryThreadInfo(params),
      queryThreadStates(params),
      queryDeadlocks(params)
    ])
  } else {
    // 无时间范围：查询最新的单条记录
    await queryThreadInfo(params)
    
    // 如果有线程信息，再查询对应的状态
    if (threadInfo.value?.jvmThreadId) {
      await Promise.all([
        getThreadState(threadInfo.value.jvmThreadId),
        queryDeadlocks(params)
      ])
    }
  }
}

// 处理时间范围变化
const handleTimeRangeChange = (range: [number, number] | null) => {
  timeRange.value = range
  if (range) {
    loadData()
  }
}

// 处理刷新
const handleRefresh = () => {
  if (timeRange.value) {
    loadData()
  }
}

// 监听资源ID变化
watch(() => props.jvmResourceId, () => {
  if (timeRange.value) {
    loadData()
  }
}, { immediate: false })

// 监听趋势数据变化
watch(() => threadTrendData.value, () => {
  nextTick(() => {
    updateCharts()
  })
}, { deep: true })

watch(() => threadStateTrendData.value, () => {
  nextTick(() => {
    updateCharts()
  })
}, { deep: true })

// 监听死锁数据变化
watch(() => deadlocks.value, () => {
  nextTick(() => {
    if (deadlocks.value && deadlocks.value.length > 0 && !deadlockChart) {
      initDeadlockChart()
    } else if (deadlockChart) {
      deadlockChart.setOption(deadlockOptions.value)
    }
  })
}, { deep: true })

onMounted(() => {
  // 初始化图表
  nextTick(() => {
    initThreadCountChart()
    initThreadStateChart()
    initThreadRatioChart()
    // 死锁图表会在有数据时才初始化
  })
  
  // 默认选择最近2小时
  const now = Date.now()
  timeRange.value = [now - 2 * 60 * 60 * 1000, now]
  
  // 加载数据
  if (props.jvmResourceId && timeRange.value) {
    loadData()
  }
  
  // 监听窗口大小变化
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  if (threadCountChart) {
    threadCountChart.dispose()
    threadCountChart = null
  }
  if (threadStateChart) {
    threadStateChart.dispose()
    threadStateChart = null
  }
  if (threadRatioChart) {
    threadRatioChart.dispose()
    threadRatioChart = null
  }
  if (deadlockChart) {
    deadlockChart.dispose()
    deadlockChart = null
  }
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped lang="scss">
.thread-monitor {
  .deadlock-overview {
    margin-bottom: 16px;
    
    .overview-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 16px;
      align-items: stretch;
    }
    
    .stat-card {
      background: linear-gradient(135deg, #f8faff 0%, #f0f4ff 100%);
      border: 1px solid #e8f0fe;
      border-radius: 12px;
      padding: 20px;
      text-align: center;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      
      &::before {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        height: 3px;
        background: linear-gradient(90deg, #1890ff, #40a9ff);
      }
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px rgba(24, 144, 255, 0.12);
        border-color: #40a9ff;
      }
      
      .stat-label {
        font-size: 13px;
        color: #666;
        font-weight: 500;
        margin-bottom: 8px;
        letter-spacing: 0.5px;
      }
      
      .stat-value {
        font-size: 28px;
        font-weight: 700;
        color: #1890ff;
        margin-bottom: 4px;
        line-height: 1;
        
        &.danger {
          color: #ff4d4f;
        }
        
        &.safe {
          color: #52c41a;
        }
      }
      
      .stat-suffix {
        font-size: 12px;
        color: #999;
        font-weight: 400;
      }
    }
    
    .status-card {
      background: linear-gradient(135deg, #fafafa 0%, #f5f5f5 100%);
      border: 1px solid #e8e8e8;
      border-radius: 12px;
      padding: 20px;
      text-align: center;
      transition: all 0.3s ease;
      position: relative;
      overflow: hidden;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px rgba(0, 0, 0, 0.08);
      }
      
      .status-label {
        font-size: 13px;
        color: #666;
        font-weight: 500;
        margin-bottom: 12px;
        letter-spacing: 0.5px;
      }
      
      .status-badge {
        display: inline-flex;
        align-items: center;
        gap: 8px;
        padding: 8px 16px;
        border-radius: 20px;
        font-weight: 600;
        font-size: 14px;
        transition: all 0.3s ease;
        
        &.success {
          background: linear-gradient(135deg, #f6ffed 0%, #d9f7be 100%);
          color: #389e0d;
          border: 1px solid #b7eb8f;
          
          .status-icon {
            color: #52c41a;
          }
        }
        
        &.error {
          background: linear-gradient(135deg, #fff2f0 0%, #ffccc7 100%);
          color: #cf1322;
          border: 1px solid #ffadd2;
          
          .status-icon {
            color: #ff4d4f;
          }
        }
        
        .status-icon {
          flex-shrink: 0;
          transition: transform 0.3s ease;
        }
        
        .status-text {
          white-space: nowrap;
        }
        
        &:hover {
          transform: scale(1.05);
          
          .status-icon {
            transform: rotate(10deg);
          }
        }
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .thread-monitor {
    .deadlock-overview {
      .overview-grid {
        grid-template-columns: 1fr;
        gap: 12px;
      }
      
      .stat-card,
      .status-card {
        padding: 16px;
        
        .stat-value {
          font-size: 24px;
        }
      }
    }
  }
}

// 暗色主题适配
[data-theme='dark'] {
  .thread-monitor {
    .deadlock-overview {
      .stat-card {
        background: linear-gradient(135deg, #1a1a1a 0%, #262626 100%);
        border-color: #404040;
        
        &::before {
          background: linear-gradient(90deg, #177ddc, #40a9ff);
        }
        
        .stat-label {
          color: #bfbfbf;
        }
        
        .stat-value {
          color: #40a9ff;
          
          &.danger {
            color: #ff7875;
          }
          
          &.safe {
            color: #73d13d;
          }
        }
        
        .stat-suffix {
          color: #8c8c8c;
        }
      }
      
      .status-card {
        background: linear-gradient(135deg, #1a1a1a 0%, #262626 100%);
        border-color: #404040;
        
        .status-label {
          color: #bfbfbf;
        }
        
        .status-badge {
          &.success {
            background: linear-gradient(135deg, #162312 0%, #274916 100%);
            color: #73d13d;
            border-color: #389e0d;
          }
          
          &.error {
            background: linear-gradient(135deg, #2a1215 0%, #431418 100%);
            color: #ff7875;
            border-color: #cf1322;
          }
        }
      }
    }
  }
}
</style>
