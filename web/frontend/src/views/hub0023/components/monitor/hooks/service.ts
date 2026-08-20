/**
 * Hub0023 监控模块业务逻辑层
 * 处理所有与后端交互的业务逻辑
 */

import type { RsSearchFormExpose } from '@/components/form/rs-search'
import { useAppMessage } from '@/composables/useAppMessage'
import { formatDate, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import type { Ref } from 'vue'
import { getGatewayMonitoringChartData, getGatewayMonitoringOverview } from '../../../api'
import { useMonitoringModel } from './model'
import type { GatewayMonitoringQueryParams } from './types'

/** 从 RsDatePicker range（{ start, end }）解析毫秒起止 */
function rangeToMs(value: unknown): [number, number] | null {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return null
  const { start, end } = value as { start?: string; end?: string }
  if (!start || !end) return null
  const startMs = new Date(start.replace(' ', 'T')).getTime()
  const endMs = new Date(end.replace(' ', 'T')).getTime()
  if (Number.isNaN(startMs) || Number.isNaN(endMs)) return null
  return [startMs, endMs]
}

/**
 * 监控服务 Hook（纯业务逻辑）
 */
export function useMonitoringService(searchFormRef?: Ref<RsSearchFormExpose | null>) {
  const message = useAppMessage()

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
  const validateTimeRange = (range: unknown): boolean => {
    const bounds = rangeToMs(range)
    if (!bounds) {
      message.error('请选择时间范围')
      return false
    }

    const duration = bounds[1] - bounds[0]
    const maxDuration = 24 * 60 * 60 * 1000

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

    const gatewayInstanceId = String(finalSearchParams?.gatewayInstanceId ?? '').trim()
    if (!gatewayInstanceId) {
      message.error('请选择网关实例')
      return
    }

    model.lastGatewayInstanceId.value = gatewayInstanceId
    const gname = String(finalSearchParams?.gatewayInstanceName ?? '').trim()
    if (gname) {
      model.lastGatewayInstanceName.value = gname
    }

    loading.value = true
    try {
      const bounds = rangeToMs(timeRangeValue)!
      const [startTime, endTime] = bounds
      const queryParams: GatewayMonitoringQueryParams = {
        gatewayInstanceId,
        startTime: formatDate(startTime, 'YYYY-MM-DDTHH:mm:ss'),
        endTime: formatDate(endTime, 'YYYY-MM-DDTHH:mm:ss'),
        timeGranularity: (finalSearchParams?.timeGranularity || timeGranularity.value) as any,
        ...(finalSearchParams?.routeName && { routeName: finalSearchParams.routeName }),
        ...(finalSearchParams?.requestPath && { requestPath: finalSearchParams.requestPath }),
      }

      // 并行请求概览与图表
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
    if (formData?.timeRange && typeof formData.timeRange === 'object' && !Array.isArray(formData.timeRange)) {
      const range = formData.timeRange as { start?: string; end?: string }
      timeRange.value =
        range.start && range.end ? { start: range.start, end: range.end } : null
    }
    if (formData?.timeGranularity) {
      timeGranularity.value = formData.timeGranularity
    }
    if (formData?.gatewayInstanceId != null && String(formData.gatewayInstanceId).trim() !== '') {
      model.lastGatewayInstanceId.value = String(formData.gatewayInstanceId).trim()
    }
    if (formData?.gatewayInstanceName != null && String(formData.gatewayInstanceName).trim() !== '') {
      model.lastGatewayInstanceName.value = String(formData.gatewayInstanceName).trim()
    }

    await loadMonitoringData(formData)
  }

  /**
   * 处理重置
   */
  const handleReset = async () => {
    timeRange.value = model.initTimeRange()
    timeGranularity.value = 'MINUTE' as any
    resetOverviewData()
    resetChartData()

    const formSnapshot = searchFormRef?.value?.getFormData?.() || {}
    const gid =
      String(formSnapshot.gatewayInstanceId || model.lastGatewayInstanceId.value || '').trim()
    const gname = String(
      formSnapshot.gatewayInstanceName || model.lastGatewayInstanceName.value || '',
    ).trim()
    await loadMonitoringData({
      gatewayInstanceId: gid,
      gatewayInstanceName: gname || gid,
      timeRange: timeRange.value,
      timeGranularity: timeGranularity.value,
      routeName: formSnapshot.routeName,
      requestPath: formSnapshot.requestPath,
    })
    if (gid && searchFormRef?.value?.setFormData) {
      searchFormRef.value.setFormData({
        ...formSnapshot,
        gatewayInstanceId: gid,
        gatewayInstanceName: gname || gid,
        timeRange: timeRange.value,
        timeGranularity: timeGranularity.value,
      })
    }
  }

  /**
   * 刷新监控数据
   */
  const refreshMonitoringData = async () => {
    await loadMonitoringData()
  }

  return {
    model,
    validateTimeRange,
    loadMonitoringData,
    handleSearch,
    handleReset,
    refreshMonitoringData,
  }
}

export type MonitoringService = ReturnType<typeof useMonitoringService>
