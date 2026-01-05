/**
 * Hub0023 监控模块业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import { formatDate, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { getGatewayMonitoringChartData, getGatewayMonitoringOverview } from '../../../api'
import { useMonitoringModel } from './model'
import type { GatewayMonitoringQueryParams } from './types'

/**
 * 监控服务 Hook（纯业务逻辑）
 */
export function useMonitoringService(searchFormRef?: Ref<any> | any) {
  const message = useMessage()

  // 初始化 Model
  const model = useMonitoringModel()

  const {
    loading,
    overviewData,
    chartData,
    timeRange,
    timeGranularity,
    resetOverviewData,
    resetChartData,
  } = model

  /**
   * 验证时间范围
   */
  const validateTimeRange = (range: [number, number] | null): boolean => {
    if (!range || range.length !== 2) {
      message.error('请选择时间范围')
      return false
    }

    const startTime = range[0]
    const endTime = range[1]
    const duration = endTime - startTime
    const maxDuration = 24 * 60 * 60 * 1000 // 24小时的毫秒数

    if (duration > maxDuration) {
      message.error('时间范围不能超过24小时，请重新选择')
      return false
    }

    if (duration <= 0) {
      message.error('结束时间必须大于开始时间')
      return false
    }

    return true
  }

  /**
   * 加载监控数据
   */
  const loadMonitoringData = async (searchParams?: Record<string, any>) => {
    // 如果没有传入查询参数，从搜索表单获取
    let finalSearchParams = searchParams
    if (!finalSearchParams && searchFormRef?.value?.getFormData) {
      finalSearchParams = searchFormRef.value.getFormData() || {}
    }

    // 处理时间范围字段（从 datetimerange 转换为 startTime 和 endTime）
    const timeRangeValue = finalSearchParams?.timeRange || timeRange.value

    if (!validateTimeRange(timeRangeValue)) {
      return
    }

    loading.value = true
    try {
      const [startTime, endTime] = timeRangeValue!
      const queryParams: GatewayMonitoringQueryParams = {
        startTime: formatDate(startTime, 'YYYY-MM-DDTHH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DDTHH:mm:ss'),
        timeGranularity: (finalSearchParams?.timeGranularity || timeGranularity.value) as any,
        // 添加额外的查询条件
        ...(finalSearchParams?.routeName && { routeName: finalSearchParams.routeName }),
        ...(finalSearchParams?.requestPath && { requestPath: finalSearchParams.requestPath }),
      }

      // 并行请求概览数据和图表数据
      const [overviewResult, chartResult] = await Promise.all([
        getGatewayMonitoringOverview(queryParams),
        getGatewayMonitoringChartData(queryParams),
      ])

      // 更新概览数据
      if (isApiSuccess(overviewResult) && overviewResult.bizData) {
        const overview = parseJsonData<typeof overviewData>(overviewResult)
        Object.assign(overviewData, overview)
      } else {
        const errorMsg = getApiMessage(overviewResult, '获取概览数据失败')
        message.warning(errorMsg)
        resetOverviewData()
      }

      // 更新图表数据
      if (isApiSuccess(chartResult) && chartResult.bizData) {
        const charts = parseJsonData<typeof chartData>(chartResult)
        Object.assign(chartData, charts)
      } else {
        const errorMsg = getApiMessage(chartResult, '获取图表数据失败')
        message.warning(errorMsg)
        resetChartData()
      }

      message.success('监控数据加载成功')
    } catch (error) {
      console.error('加载监控数据失败:', error)
      message.error('加载监控数据失败，请检查网络连接或联系管理员')
      resetOverviewData()
      resetChartData()
    } finally {
      loading.value = false
    }
  }

  /**
   * 处理搜索
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    // 更新时间范围和时间粒度
    if (formData?.timeRange) {
      timeRange.value = formData.timeRange
    }
    if (formData?.timeGranularity) {
      timeGranularity.value = formData.timeGranularity
    }

    await loadMonitoringData(formData)
  }

  /**
   * 处理重置
   */
  const handleReset = async () => {
    // 重置为默认值
    timeRange.value = model.initTimeRange()
    timeGranularity.value = 'MINUTE' as any
    resetOverviewData()
    resetChartData()

    // 重新加载数据
    await loadMonitoringData({
      timeRange: timeRange.value,
      timeGranularity: timeGranularity.value,
    })
  }

  /**
   * 刷新监控数据
   */
  const refreshMonitoringData = async () => {
    await loadMonitoringData()
  }

  return {
    // Model
    model,

    // 方法
    validateTimeRange,
    loadMonitoringData,
    handleSearch,
    handleReset,
    refreshMonitoringData,
  }
}

/**
 * Service 返回类型
 */
export type MonitoringService = ReturnType<typeof useMonitoringService>

