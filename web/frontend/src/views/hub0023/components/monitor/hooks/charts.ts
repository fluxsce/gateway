/**
 * Hub0023 监控模块图表管理 Hook
 * 内聚在监控组件内部
 */

import { formatDate } from '@/utils/format'
import * as echarts from 'echarts'
import { nextTick, onUnmounted, ref } from 'vue'
import type {
  GatewayMonitoringChartData,
  GatewayMonitoringOverview,
} from './types'

/**
 * 监控图表管理
 */
export function useMonitoringCharts() {
  // 图表DOM引用
  const requestTrendChartRef = ref<HTMLDivElement>()
  const responseTimeChartRef = ref<HTMLDivElement>()
  const requestMetricsChartRef = ref<HTMLDivElement>()
  const statusCodeChartRef = ref<HTMLDivElement>()
  const hotRoutesChartRef = ref<HTMLDivElement>()

  // 图表实例
  let requestTrendChart: echarts.ECharts | null = null
  let responseTimeChart: echarts.ECharts | null = null
  let requestMetricsChart: echarts.ECharts | null = null
  let statusCodeChart: echarts.ECharts | null = null
  let hotRoutesChart: echarts.ECharts | null = null

  /**
   * 获取状态码颜色
   * 为不同状态码提供不同颜色，避免重复
   */
  const getStatusCodeColor = (statusCode: string) => {
    // 预定义颜色调色板
    const colorPalettes = {
      // 2xx 成功状态码 - 绿色系
      success: ['#52c41a', '#73d13d', '#95de64', '#b7eb8f', '#d9f7be', '#389e0d', '#237804'],
      // 3xx 重定向状态码 - 橙色系
      redirect: ['#fa8c16', '#ffa940', '#ffc069', '#ffd591', '#ffe7ba', '#d46b08', '#ad4e00'],
      // 4xx 客户端错误 - 黄色系
      clientError: ['#faad14', '#ffc53d', '#ffd666', '#ffe58f', '#fff1b8', '#d48806', '#ad6800'],
      // 5xx 服务端错误 - 红色系
      serverError: ['#ff4d4f', '#ff7875', '#ffa39e', '#ffccc7', '#ffe1e1', '#f5222d', '#cf1322'],
      // 其他状态码 - 灰色系
      other: ['#d9d9d9', '#bfbfbf', '#8c8c8c', '#595959', '#434343', '#262626', '#1f1f1f'],
    }

    // 具体状态码映射
    const specificColors: Record<string, string> = {
      // 2xx 成功
      '200': colorPalettes.success[0], // OK - 主绿色
      '201': colorPalettes.success[1], // Created - 亮绿色
      '202': colorPalettes.success[2], // Accepted - 中绿色
      '204': colorPalettes.success[3], // No Content - 浅绿色
      '206': colorPalettes.success[4], // Partial Content - 很浅绿色

      // 3xx 重定向
      '301': colorPalettes.redirect[0], // Moved Permanently - 主橙色
      '302': colorPalettes.redirect[1], // Found - 亮橙色
      '304': colorPalettes.redirect[2], // Not Modified - 中橙色
      '307': colorPalettes.redirect[3], // Temporary Redirect - 浅橙色
      '308': colorPalettes.redirect[4], // Permanent Redirect - 很浅橙色

      // 4xx 客户端错误
      '400': colorPalettes.clientError[0], // Bad Request - 主黄色
      '401': colorPalettes.clientError[1], // Unauthorized - 亮黄色
      '403': colorPalettes.clientError[2], // Forbidden - 中黄色
      '404': colorPalettes.clientError[3], // Not Found - 浅黄色
      '405': colorPalettes.clientError[4], // Method Not Allowed - 很浅黄色
      '409': colorPalettes.clientError[5], // Conflict - 深黄色
      '429': colorPalettes.clientError[6], // Too Many Requests - 最深黄色

      // 5xx 服务端错误
      '500': colorPalettes.serverError[0], // Internal Server Error - 主红色
      '501': colorPalettes.serverError[1], // Not Implemented - 亮红色
      '502': colorPalettes.serverError[2], // Bad Gateway - 中红色
      '503': colorPalettes.serverError[3], // Service Unavailable - 浅红色
      '504': colorPalettes.serverError[4], // Gateway Timeout - 很浅红色
      '505': colorPalettes.serverError[5], // HTTP Version Not Supported - 深红色
    }

    // 优先使用具体状态码映射
    if (specificColors[statusCode]) {
      return specificColors[statusCode]
    }

    // 如果没有具体映射，则按范围分配颜色，并根据状态码值计算偏移
    const code = parseInt(statusCode)
    let palette: string[]
    let baseIndex = 0

    if (code >= 200 && code < 300) {
      palette = colorPalettes.success
      baseIndex = (code - 200) % palette.length
    } else if (code >= 300 && code < 400) {
      palette = colorPalettes.redirect
      baseIndex = (code - 300) % palette.length
    } else if (code >= 400 && code < 500) {
      palette = colorPalettes.clientError
      baseIndex = (code - 400) % palette.length
    } else if (code >= 500 && code < 600) {
      palette = colorPalettes.serverError
      baseIndex = (code - 500) % palette.length
    } else {
      palette = colorPalettes.other
      baseIndex = code % palette.length
    }

    return palette[baseIndex]
  }

  /**
   * 获取热点路由颜色（基于错误率）
   */
  const getHotRouteColor = (errorRate: number) => {
    if (errorRate <= 1) {
      return '#52c41a' // 低错误率 - 绿色
    } else if (errorRate <= 5) {
      return '#faad14' // 中等错误率 - 黄色
    } else if (errorRate <= 10) {
      return '#fa8c16' // 较高错误率 - 橙色
    } else {
      return '#ff4d4f' // 高错误率 - 红色
    }
  }

  /**
   * 初始化请求趋势图
   */
  const initRequestTrendChart = (chartData: GatewayMonitoringChartData) => {
    if (requestTrendChartRef.value && !requestTrendChart) {
      requestTrendChart = echarts.init(requestTrendChartRef.value)
    }
    if (requestTrendChart) {
      requestTrendChart.setOption({
        tooltip: { trigger: 'axis' },
        xAxis: {
          type: 'time',
          axisLabel: {
            formatter: function (value: number) {
              return formatDate(value, 'HH:mm')
            },
          },
        },
        yAxis: { type: 'value', name: '请求数' },
        grid: { top: 40, bottom: 60, left: 80, right: 40 },
        series: [
          {
            name: '总请求数',
            type: 'line',
            data: chartData.requestTrend.length > 0 
              ? chartData.requestTrend.map((item) => [item.timestamp, item.totalRequests])
              : [],
            smooth: true,
            areaStyle: { opacity: 0.3 },
            itemStyle: { color: '#18a058' },
          },
        ],
      }, true)
    }
  }

  /**
   * 初始化响应时间图
   */
  const initResponseTimeChart = (chartData: GatewayMonitoringChartData) => {
    if (responseTimeChartRef.value && !responseTimeChart) {
      responseTimeChart = echarts.init(responseTimeChartRef.value)
    }
    if (responseTimeChart) {
      responseTimeChart.setOption({
        tooltip: {
          trigger: 'axis',
          formatter: function (params: any) {
            let tooltip = `<div style="padding: 8px;"><strong>${formatDate(params[0].axisValue, 'MM-DD HH:mm')}</strong><br/>`

            // 从responseTimeTrend数据中找到对应的请求数量
            const dataPoint = chartData.responseTimeTrend.find(
              (item) => item.timestamp === params[0].axisValue,
            )
            if (dataPoint) {
              tooltip += `请求数量: ${dataPoint.requestCount}<br/>`
            }

            params.forEach((param: any) => {
              tooltip += `${param.seriesName}: ${param.value[1]}ms<br/>`
            })
            tooltip += '</div>'
            return tooltip
          },
        },
        xAxis: {
          type: 'time',
          axisLabel: {
            formatter: function (value: number) {
              return formatDate(value, 'HH:mm')
            },
          },
        },
        yAxis: { type: 'value', name: '毫秒' },
        legend: {
          top: 0,
          type: 'scroll',
          orient: 'horizontal',
          itemGap: 20,
        },
        grid: { top: 60, bottom: 60, left: 80, right: 40 },
        series: [
          {
            name: '平均响应时间',
            type: 'line',
            data: chartData.responseTimeTrend.map((item) => [
              item.timestamp,
              item.avgResponseTimeMs,
            ]),
            smooth: true,
            itemStyle: { color: '#5470c6' },
          },
          {
            name: 'P50响应时间',
            type: 'line',
            data: chartData.responseTimeTrend.map((item) => [
              item.timestamp,
              item.p50ResponseTimeMs,
            ]),
            smooth: true,
            itemStyle: { color: '#91cc75' },
          },
          {
            name: 'P90响应时间',
            type: 'line',
            data: chartData.responseTimeTrend.map((item) => [
              item.timestamp,
              item.p90ResponseTimeMs,
            ]),
            smooth: true,
            itemStyle: { color: '#fac858' },
          },
          {
            name: 'P99响应时间',
            type: 'line',
            data: chartData.responseTimeTrend.map((item) => [
              item.timestamp,
              item.p99ResponseTimeMs,
            ]),
            smooth: true,
            itemStyle: { color: '#ee6666' },
          },
          {
            name: '最大响应时间',
            type: 'line',
            data: chartData.responseTimeTrend.map((item) => [
              item.timestamp,
              item.maxResponseTimeMs,
            ]),
            smooth: true,
            itemStyle: { color: '#73c0de' },
            lineStyle: { type: 'dashed' },
          },
        ],
      }, true)
    }
  }

  /**
   * 初始化请求指标饼图
   */
  const initRequestMetricsChart = (overviewData: GatewayMonitoringOverview) => {
    if (requestMetricsChartRef.value && !requestMetricsChart) {
      requestMetricsChart = echarts.init(requestMetricsChartRef.value)
    }
    if (requestMetricsChart) {
      requestMetricsChart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: '{a} <br/>{b}: {c} ({d}%)',
        },
        legend: {
          orient: 'vertical',
          left: 'left',
          top: 'center',
        },
        series: [
          {
            name: '请求指标',
            type: 'pie',
            data: [
              {
                name: '成功请求',
                value: overviewData.successRequests,
                itemStyle: { color: '#52c41a' },
              },
              {
                name: '失败请求',
                value: overviewData.failedRequests,
                itemStyle: { color: '#ff4d4f' },
              },
            ],
            radius: ['40%', '70%'],
            center: ['65%', '50%'],
            emphasis: {
              itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' },
            },
            label: {
              show: true,
              formatter: '{b}: {d}%',
            },
          },
        ],
      }, true)
    }
  }

  /**
   * 初始化状态码分布图
   */
  const initStatusCodeChart = (chartData: GatewayMonitoringChartData) => {
    if (statusCodeChartRef.value && !statusCodeChart) {
      statusCodeChart = echarts.init(statusCodeChartRef.value)
    }
    if (statusCodeChart) {
      statusCodeChart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: function (params: any) {
            const data = chartData.statusCodeDistribution.find(
              (item) => item.statusCode === params.name,
            )
            if (data) {
              return `
                                <div style="padding: 8px;">
                                    <div><strong>${data.statusCode}</strong></div>
                                    <div>分类: ${data.category || '未分类'}</div>
                                    <div>描述: ${data.description || '无描述'}</div>
                                    <div>数量: ${data.count}</div>
                                    <div>百分比: ${data.percentage.toFixed(2)}%</div>
                                </div>
                            `
            }
            return `${params.name}: ${params.value} (${params.percent}%)`
          },
        },
        legend: {
          orient: 'vertical',
          left: 'left',
          top: 'center',
        },
        series: [
          {
            name: '状态码',
            type: 'pie',
            data: chartData.statusCodeDistribution.map((item) => ({
              name: item.statusCode,
              value: item.count,
              itemStyle: {
                color: getStatusCodeColor(item.statusCode),
              },
            })),
            radius: ['40%', '70%'],
            center: ['65%', '50%'],
            emphasis: {
              itemStyle: {
                shadowBlur: 10,
                shadowOffsetX: 0,
                shadowColor: 'rgba(0, 0, 0, 0.5)',
              },
            },
            label: {
              show: true,
              formatter: '{b}: {d}%',
            },
          },
        ],
      }, true)
    }
  }

  /**
   * 初始化热点路由图
   */
  const initHotRoutesChart = (chartData: GatewayMonitoringChartData) => {
    if (hotRoutesChartRef.value && !hotRoutesChart) {
      hotRoutesChart = echarts.init(hotRoutesChartRef.value)
    }
    if (hotRoutesChart) {
      hotRoutesChart.setOption({
        tooltip: {
          trigger: 'axis',
          confine: false, // 允许 tooltip 超出图表容器，避免被裁剪
          // 注意：ECharts 5.x 中 tooltip 会自动附加到 body，但需要确保 z-index 足够高
          formatter: function (params: any) {
            const data = chartData.hotRoutes.find((item) => item.routePath === params[0].name)
            if (data) {
              return `
                                <div style="padding: 8px;">
                                    <div><strong>${data.routePath}</strong></div>
                                    <div>请求数量: ${data.requestCount}</div>
                                    <div>最大响应时间: ${data.maxResponseTimeMs || 'N/A'}ms</div>
                                    <div>最小响应时间: ${data.minResponseTimeMs || 'N/A'}ms</div>
                                    <div>错误率: ${data.errorRate}%</div>
                                    <div>QPS: ${data.qps}</div>
                                    <div>服务名称: ${data.serviceName || 'N/A'}</div>
                                </div>
                            `
            }
            return ''
          },
        },
        xAxis: { type: 'value', name: '请求数' },
        yAxis: {
          type: 'category',
          data: chartData.hotRoutes.map((item) => item.routePath),
          axisLabel: {
            formatter: function (value: string) {
              return value.length > 20 ? value.substring(0, 20) + '...' : value
            },
          },
        },
        grid: { top: 40, bottom: 60, left: 140, right: 80 },
        series: [
          {
            name: '请求量',
            type: 'bar',
            data: chartData.hotRoutes.map((item) => item.requestCount),
            itemStyle: { color: '#5470c6' },
            barMaxWidth: 40,
          },
        ],
      }, true)
    }
  }

  /**
   * 初始化所有图表
   */
  const initCharts = async (
    overviewData: GatewayMonitoringOverview,
    chartData: GatewayMonitoringChartData,
  ) => {
    await nextTick()

    // 初始化所有图表（即使没有数据也创建实例，后续数据更新时会自动渲染）
    if (requestTrendChartRef.value && !requestTrendChart) {
      initRequestTrendChart(chartData)
    }

    if (responseTimeChartRef.value && !responseTimeChart) {
      initResponseTimeChart(chartData)
    }

    if (requestMetricsChartRef.value && !requestMetricsChart) {
      initRequestMetricsChart(overviewData)
    }

    if (statusCodeChartRef.value && !statusCodeChart) {
      initStatusCodeChart(chartData)
    }

    if (hotRoutesChartRef.value && !hotRoutesChart) {
      initHotRoutesChart(chartData)
    }
  }

  /**
   * 更新图表数据
   */
  const updateCharts = (
    overviewData: GatewayMonitoringOverview,
    chartData: GatewayMonitoringChartData,
  ) => {
    // 请求趋势图
    if (chartData.requestTrend.length > 0) {
      if (!requestTrendChart) {
        initRequestTrendChart(chartData)
      } else {
        requestTrendChart.setOption({
          xAxis: {
            type: 'time',
            axisLabel: {
              formatter: function (value: number) {
                return formatDate(value, 'HH:mm')
              },
            },
          },
          series: [
            {
              data: chartData.requestTrend.map((item) => [item.timestamp, item.totalRequests]),
              itemStyle: { color: '#18a058' },
            },
          ],
        })
      }
    } else if (requestTrendChart) {
      requestTrendChart.dispose()
      requestTrendChart = null
    }

    // 响应时间趋势图
    if (chartData.responseTimeTrend.length > 0) {
      if (!responseTimeChart) {
        initResponseTimeChart(chartData)
      } else {
        responseTimeChart.setOption({
          xAxis: {
            type: 'time',
            axisLabel: {
              formatter: function (value: number) {
                return formatDate(value, 'HH:mm')
              },
            },
          },
          legend: {
            top: 0,
            type: 'scroll',
            orient: 'horizontal',
            itemGap: 20,
          },
          series: [
            {
              data: chartData.responseTimeTrend.map((item) => [
                item.timestamp,
                item.avgResponseTimeMs,
              ]),
            },
            {
              data: chartData.responseTimeTrend.map((item) => [
                item.timestamp,
                item.p50ResponseTimeMs,
              ]),
            },
            {
              data: chartData.responseTimeTrend.map((item) => [
                item.timestamp,
                item.p90ResponseTimeMs,
              ]),
            },
            {
              data: chartData.responseTimeTrend.map((item) => [
                item.timestamp,
                item.p99ResponseTimeMs,
              ]),
            },
            {
              data: chartData.responseTimeTrend.map((item) => [
                item.timestamp,
                item.maxResponseTimeMs,
              ]),
            },
          ],
        })
      }
    } else if (responseTimeChart) {
      responseTimeChart.dispose()
      responseTimeChart = null
    }

    // 请求指标饼图
    if (overviewData.totalRequests > 0) {
      if (!requestMetricsChart) {
        initRequestMetricsChart(overviewData)
      } else {
        requestMetricsChart.setOption({
          legend: {
            orient: 'vertical',
            left: 'left',
            top: 'center',
          },
          series: [
            {
              data: [
                {
                  name: '成功请求',
                  value: overviewData.successRequests,
                  itemStyle: { color: '#52c41a' },
                },
                {
                  name: '失败请求',
                  value: overviewData.failedRequests,
                  itemStyle: { color: '#ff4d4f' },
                },
              ],
              center: ['65%', '50%'],
            },
          ],
        })
      }
    } else if (requestMetricsChart) {
      requestMetricsChart.dispose()
      requestMetricsChart = null
    }

    // 状态码分布图
    if (chartData.statusCodeDistribution.length > 0) {
      if (!statusCodeChart) {
        initStatusCodeChart(chartData)
      } else {
        statusCodeChart.setOption({
          legend: {
            orient: 'vertical',
            left: 'left',
            top: 'center',
          },
          series: [
            {
              data: chartData.statusCodeDistribution.map((item) => ({
                name: item.statusCode,
                value: item.count,
                itemStyle: { color: getStatusCodeColor(item.statusCode) },
              })),
              center: ['65%', '50%'],
            },
          ],
        })
      }
    } else if (statusCodeChart) {
      statusCodeChart.dispose()
      statusCodeChart = null
    }

    // 热点路由图
    if (chartData.hotRoutes.length > 0) {
      if (!hotRoutesChart) {
        initHotRoutesChart(chartData)
      } else {
        hotRoutesChart.setOption({
          yAxis: {
            data: chartData.hotRoutes.map((item) => item.routePath),
            axisLabel: {
              formatter: function (value: string) {
                return value.length > 20 ? value.substring(0, 20) + '...' : value
              },
            },
          },
          series: [
            {
              data: chartData.hotRoutes.map((item) => ({
                value: item.requestCount,
                itemStyle: { color: getHotRouteColor(item.errorRate) },
              })),
            },
          ],
        })
      }
    } else if (hotRoutesChart) {
      hotRoutesChart.dispose()
      hotRoutesChart = null
    }
  }

  /**
   * 销毁图表
   */
  const destroyCharts = () => {
    if (requestTrendChart) {
      requestTrendChart.dispose()
      requestTrendChart = null
    }
    if (responseTimeChart) {
      responseTimeChart.dispose()
      responseTimeChart = null
    }
    if (requestMetricsChart) {
      requestMetricsChart.dispose()
      requestMetricsChart = null
    }
    if (statusCodeChart) {
      statusCodeChart.dispose()
      statusCodeChart = null
    }
    if (hotRoutesChart) {
      hotRoutesChart.dispose()
      hotRoutesChart = null
    }
  }

  // 组件卸载时销毁图表
  onUnmounted(() => {
    destroyCharts()
  })

  return {
    // DOM引用
    requestTrendChartRef,
    responseTimeChartRef,
    requestMetricsChartRef,
    statusCodeChartRef,
    hotRoutesChartRef,

    // 方法
    initCharts,
    updateCharts,
    destroyCharts,
  }
}

/**
 * Charts 返回类型
 */
export type MonitoringCharts = ReturnType<typeof useMonitoringCharts>

