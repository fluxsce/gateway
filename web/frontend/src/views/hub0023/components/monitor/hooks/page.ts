/**
 * Hub0023 监控模块页面级 Hook
 * - 组合 useMonitoringService（纯业务逻辑）
 * - 处理工具栏、搜索等页面交互
 */

import type { Ref } from 'vue'
import { onBeforeUnmount, watch } from 'vue'
import { useMonitoringCharts } from './charts'
import { useMonitoringService } from './service'

/**
 * 监控页面级 Hook
 */
export function useMonitoringPage(searchFormRef?: Ref<any> | any) {
  // 业务服务（包含 model、查询等）
  const service = useMonitoringService(searchFormRef)

  // 图表管理
  const charts = useMonitoringCharts()

  /**
   * 处理搜索（接收 SearchForm 传递的表单数据）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    await service.handleSearch(formData)
  }

  /**
   * 处理重置
   */
  const handleReset = async () => {
    await service.handleReset()
  }

  /**
   * 工具栏按钮点击处理
   */
  const handleToolbarClick = async (key: string) => {
    switch (key) {
      case 'refresh':
        await service.refreshMonitoringData()
        break
      default:
        console.warn(`未知的工具栏按钮: ${key}`)
    }
  }

  /**
   * 初始化页面数据
   */
  const initPageData = async () => {
    // 初始化时间范围
    service.model.timeRange.value = service.model.initTimeRange()

    // 等待图表 ref 绑定后再初始化图表
    await nextTick()
    await charts.initCharts(service.model.overviewData, service.model.chartData)

    // 加载监控数据
    await service.loadMonitoringData()
  }

  // 监听数据变化，自动更新图表
  const stopWatch = watch(
    [() => service.model.overviewData, () => service.model.chartData],
    () => {
      charts.updateCharts(service.model.overviewData, service.model.chartData)
    },
    { deep: true }
  )

  // 组件卸载时清理资源
  onBeforeUnmount(() => {
    // 停止 watch 监听器
    stopWatch()
  })

  return {
    // Service
    service,

    // Charts
    charts,

    // 方法
    handleSearch,
    handleReset,
    handleToolbarClick,
    initPageData,
  }
}

/**
 * Page 返回类型
 */
export type MonitoringPage = ReturnType<typeof useMonitoringPage>

