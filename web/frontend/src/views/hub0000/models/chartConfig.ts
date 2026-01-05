/**
 * 图表配置模型
 * 用于定义各种监控图表的配置
 */

import type { MetricTrend, ChartConfig, TimeRangeOption } from '../types'

/**
 * 时间范围选项
 */
export const createTimeRangeOptions = (): TimeRangeOption[] => {
  return [
    {
      label: '最近2小时',
      value: '2h',
      duration: 2 * 60 * 60 * 1000, // 2小时
    },
    {
      label: '最近24小时',
      value: '24h',
      duration: 24 * 60 * 60 * 1000, // 24小时
    },
    {
      label: '最近7天',
      value: '7d',
      duration: 7 * 24 * 60 * 60 * 1000, // 7天
    },
    {
      label: '最近30天',
      value: '30d',
      duration: 30 * 24 * 60 * 60 * 1000, // 30天
    },
  ]
}

/**
 * 创建CPU使用率图表配置
 */
export const createCPUUsageChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: 'CPU使用率',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || 'CPU使用率',
    })),
    options: {
      smooth: true,
      color: '#1890ff',
      yAxis: {
        min: 0,
        max: 100,
        formatter: (value: number) => `${value}%`,
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          return `${time}<br/>${params.data.label}: ${params.data.value}%`
        },
      },
    },
  }
}

/**
 * 创建内存使用率图表配置
 */
export const createMemoryUsageChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: '内存使用率',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || '内存使用率',
    })),
    options: {
      smooth: true,
      color: '#52c41a',
      yAxis: {
        min: 0,
        max: 100,
        formatter: (value: number) => `${value}%`,
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          return `${time}<br/>${params.data.label}: ${params.data.value}%`
        },
      },
    },
  }
}

/**
 * 创建磁盘使用率图表配置
 */
export const createDiskUsageChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: '磁盘使用率',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || '磁盘使用率',
    })),
    options: {
      smooth: true,
      color: '#faad14',
      yAxis: {
        min: 0,
        max: 100,
        formatter: (value: number) => `${value}%`,
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          return `${time}<br/>${params.data.label}: ${params.data.value}%`
        },
      },
    },
  }
}

/**
 * 创建网络流量图表配置
 */
export const createNetworkTrafficChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: '网络流量',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || '网络流量',
    })),
    options: {
      smooth: true,
      color: '#722ed1',
      yAxis: {
        formatter: (value: number) => {
          if (value >= 1024 * 1024 * 1024) {
            return `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB/s`
          } else if (value >= 1024 * 1024) {
            return `${(value / (1024 * 1024)).toFixed(2)} MB/s`
          } else if (value >= 1024) {
            return `${(value / 1024).toFixed(2)} KB/s`
          } else {
            return `${value} B/s`
          }
        },
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          let valueStr = ''
          const value = params.data.value
          if (value >= 1024 * 1024 * 1024) {
            valueStr = `${(value / (1024 * 1024 * 1024)).toFixed(2)} GB/s`
          } else if (value >= 1024 * 1024) {
            valueStr = `${(value / (1024 * 1024)).toFixed(2)} MB/s`
          } else if (value >= 1024) {
            valueStr = `${(value / 1024).toFixed(2)} KB/s`
          } else {
            valueStr = `${value} B/s`
          }
          return `${time}<br/>${params.data.label}: ${valueStr}`
        },
      },
    },
  }
}

/**
 * 创建进程数量图表配置
 */
export const createProcessCountChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: '进程数量',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || '进程数量',
    })),
    options: {
      smooth: true,
      color: '#f5222d',
      yAxis: {
        min: 0,
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          return `${time}<br/>${params.data.label}: ${params.data.value}个`
        },
      },
    },
  }
}

/**
 * 创建温度图表配置
 */
export const createTemperatureChartConfig = (data: MetricTrend[]): ChartConfig => {
  return {
    type: 'line',
    title: '温度',
    data: data.map((item) => ({
      time: item.time,
      value: item.value,
      label: item.label || '温度',
    })),
    options: {
      smooth: true,
      color: '#fa8c16',
      yAxis: {
        min: 0,
        max: 100,
        formatter: (value: number) => `${value}°C`,
      },
      xAxis: {
        type: 'time',
      },
      tooltip: {
        formatter: (params: any) => {
          const time = new Date(params.data.time).toLocaleString()
          return `${time}<br/>${params.data.label}: ${params.data.value}°C`
        },
      },
    },
  }
}

/**
 * 创建仪表盘图表配置
 */
export const createGaugeChartConfig = (value: number, title: string, max = 100): ChartConfig => {
  return {
    type: 'gauge',
    title,
    data: [{ value, name: title }],
    options: {
      min: 0,
      max,
      splitNumber: 10,
      axisLine: {
        lineStyle: {
          width: 6,
          color: [
            [0.3, '#67e0e3'],
            [0.7, '#37a2da'],
            [1, '#fd666d'],
          ],
        },
      },
      pointer: {
        itemStyle: {
          color: 'auto',
        },
      },
      axisTick: {
        distance: -30,
        length: 8,
        lineStyle: {
          color: '#fff',
          width: 2,
        },
      },
      splitLine: {
        distance: -30,
        length: 30,
        lineStyle: {
          color: '#fff',
          width: 4,
        },
      },
      axisLabel: {
        color: 'auto',
        distance: 40,
        fontSize: 12,
        formatter: (value: number) => {
          if (max === 100) {
            return `${value}%`
          } else {
            return value.toString()
          }
        },
      },
      detail: {
        valueAnimation: true,
        formatter: (value: number) => {
          if (max === 100) {
            return `${value}%`
          } else {
            return value.toString()
          }
        },
        color: 'auto',
        fontSize: 30,
        offsetCenter: [0, '70%'],
      },
    },
  }
}

/**
 * 创建饼图配置
 */
export const createPieChartConfig = (
  data: Array<{ name: string; value: number }>,
  title: string,
): ChartConfig => {
  return {
    type: 'pie',
    title,
    data,
    options: {
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      label: {
        show: false,
        position: 'center',
      },
      emphasis: {
        label: {
          show: true,
          fontSize: '30',
          fontWeight: 'bold',
        },
      },
      labelLine: {
        show: false,
      },
      tooltip: {
        trigger: 'item',
        formatter: '{a} <br/>{b}: {c} ({d}%)',
      },
    },
  }
}

/**
 * 创建多指标对比图表配置
 */
export const createMultiMetricChartConfig = (
  data: Array<{ name: string; data: MetricTrend[] }>,
  title: string,
): ChartConfig => {
  return {
    type: 'line',
    title,
    data,
    options: {
      legend: {
        data: data.map((item) => item.name),
      },
      xAxis: {
        type: 'time',
      },
      yAxis: {
        type: 'value',
      },
      series: data.map((item, index) => ({
        name: item.name,
        type: 'line',
        smooth: true,
        data: item.data.map((d) => [d.time, d.value]),
      })),
      tooltip: {
        trigger: 'axis',
      },
    },
  }
}
