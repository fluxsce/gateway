import { useAppMessage } from '@/composables/useAppMessage'
import { computed, reactive, ref } from 'vue'
import * as routeApi from '../api'
import type { RouteConfig, RouteQueryParams, RouteStatistics } from '../types'

/**
 * 路由管理核心Hook
 */
export function useRouteManagement() {
  const message = useAppMessage()

  // 状态管理
  const loading = ref(false)
  const submitting = ref(false)
  const routes = ref<RouteConfig[]>([])
  const total = ref(0)
  const selectedRoutes = ref<string[]>([])

  // 统计信息
  const statistics = ref<RouteStatistics>({
    totalRoutes: 0,
    activeRoutes: 0,
    inactiveRoutes: 0,
    exactMatchRoutes: 0,
    prefixMatchRoutes: 0,
    regexMatchRoutes: 0,
  })

  // 查询参数
  const queryParams = reactive<RouteQueryParams>({
    tenantId: 'default', // TODO: 从用户信息获取
    gatewayInstanceId: '',
    routeName: '',
    routePath: '',
    matchType: undefined,
    activeFlag: undefined,
    pageIndex: 1,
    pageSize: 20,
  })

  // 选项配置
  const matchTypeOptions = [
    { label: '精确匹配', value: 0 },
    { label: '前缀匹配', value: 1 },
    { label: '正则匹配', value: 2 },
  ]

  const statusOptions = [
    { label: '启用', value: 'Y' },
    { label: '禁用', value: 'N' },
  ]

  // 计算属性
  const hasSelection = computed(() => selectedRoutes.value.length > 0)

  /**
   * 获取路由列表
   */
  const getRoutes = async () => {
    try {
      loading.value = true
      const response = await routeApi.queryRouteConfigs(queryParams)
      if (response.oK) {
        // 解析路由数据
        let routeList: RouteConfig[] = []
        if (response.bizData) {
          try {
            routeList = JSON.parse(response.bizData)
            // 确保是数组格式
            if (!Array.isArray(routeList)) {
              routeList = []
            }
          } catch (parseError) {
            console.error('解析路由数据失败:', parseError)
            routeList = []
          }
        }

        // 解析分页数据
        let pageInfo: any = {}
        if (response.pageQueryData) {
          try {
            pageInfo = JSON.parse(response.pageQueryData)
          } catch (parseError) {
            console.error('解析分页数据失败:', parseError)
            pageInfo = {}
          }
        }

        routes.value = routeList
        total.value = pageInfo.totalCount || routeList.length

        // 获取统计信息
        await getStatistics()

        console.log('路由列表加载成功:', {
          routeCount: routeList.length,
          totalCount: pageInfo.totalCount,
          pageIndex: pageInfo.pageIndex,
          pageSize: pageInfo.pageSize,
        })
      } else {
        console.warn('获取路由列表失败:', response.errMsg || response.popMsg)
        routes.value = []
        total.value = 0
        // 重置统计信息
        resetStatistics()
      }
    } catch (error) {
      console.error('获取路由列表失败:', error)
      message.error('获取路由列表失败')
      routes.value = []
      total.value = 0
      // 重置统计信息
      resetStatistics()
    } finally {
      loading.value = false
    }
  }

  /**
   * 获取统计信息
   */
  const getStatistics = async () => {
    try {
      const statisticsParams = {
        tenantId: queryParams.tenantId || 'default',
        gatewayInstanceId: queryParams.gatewayInstanceId,
        routeName: queryParams.routeName,
        routePath: queryParams.routePath,
        matchType: queryParams.matchType,
        activeFlag: queryParams.activeFlag,
      }

      const response = await routeApi.queryRouteStatistics(statisticsParams)
      if (response.oK && response.bizData) {
        try {
          const statisticsData = JSON.parse(response.bizData)
          statistics.value = {
            totalRoutes: statisticsData.totalRoutes || 0,
            activeRoutes: statisticsData.activeRoutes || 0,
            inactiveRoutes: statisticsData.inactiveRoutes || 0,
            exactMatchRoutes: statisticsData.exactMatchRoutes || 0,
            prefixMatchRoutes: statisticsData.prefixMatchRoutes || 0,
            regexMatchRoutes: statisticsData.regexMatchRoutes || 0,
          }
          console.log('统计信息获取成功:', statistics.value)
        } catch (parseError) {
          console.error('解析统计数据失败:', parseError)
          resetStatistics()
        }
      } else {
        console.warn('获取统计信息失败:', response.errMsg || response.popMsg)
        resetStatistics()
      }
    } catch (error) {
      console.error('获取统计信息失败:', error)
      resetStatistics()
    }
  }

  /**
   * 重置统计信息
   */
  const resetStatistics = () => {
    statistics.value = {
      totalRoutes: 0,
      activeRoutes: 0,
      inactiveRoutes: 0,
      exactMatchRoutes: 0,
      prefixMatchRoutes: 0,
      regexMatchRoutes: 0,
    }
  }

  /**
   * 删除路由
   */
  const deleteRoute = async (routeConfigId: string) => {
    try {
      submitting.value = true
      await routeApi.deleteRouteConfig(routeConfigId)
      message.success('删除成功')
      await getRoutes()
    } catch (error) {
      console.error('删除失败:', error)
      message.error('删除失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 批量删除路由
   */
  const batchDeleteRoutes = async (routeIds: string[]) => {
    try {
      submitting.value = true
      await Promise.all(
        routeIds.map((id) => routeApi.deleteRouteConfig(id)),
      )
      message.success(`成功删除 ${routeIds.length} 个路由`)
      selectedRoutes.value = []
      await getRoutes()
    } catch (error) {
      console.error('批量删除失败:', error)
      message.error('批量删除失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 切换路由状态
   */
  const toggleRoute = async (routeConfigId: string, activeFlag: 'Y' | 'N') => {
    try {
      submitting.value = true
      await routeApi.editRouteConfig({
        routeConfigId,
        gatewayInstanceId: queryParams.gatewayInstanceId || '',
        activeFlag,
      })
      message.success(`${activeFlag === 'Y' ? '启用' : '禁用'}成功`)
      await getRoutes()
    } catch (error) {
      console.error('切换状态失败:', error)
      message.error('切换状态失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 批量切换路由状态
   */
  const batchToggleRoutes = async (routeIds: string[], activeFlag: 'Y' | 'N') => {
    try {
      submitting.value = true
      await Promise.all(
        routeIds.map((id) =>
          routeApi.editRouteConfig({
            routeConfigId: id,
            gatewayInstanceId: queryParams.gatewayInstanceId || '',
            activeFlag,
          }),
        ),
      )
      message.success(`成功${activeFlag === 'Y' ? '启用' : '禁用'} ${routeIds.length} 个路由`)
      selectedRoutes.value = []
      await getRoutes()
    } catch (error) {
      console.error('批量切换状态失败:', error)
      message.error('批量切换状态失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 复制路由
   */
  const copyRoute = async (sourceRouteId: string, newRouteName: string) => {
    try {
      submitting.value = true
      // 先获取源路由信息
      const sourceRoute = await routeApi.getRouteConfig(
        sourceRouteId,
      )

      if (sourceRoute.oK && sourceRoute.bizData) {
        const routeData = JSON.parse(sourceRoute.bizData)
        // 复制路由配置
        await routeApi.addRouteConfig({
          ...routeData,
          routeName: newRouteName,
          routePath: `${routeData.routePath}_copy`,
          gatewayInstanceId: queryParams.gatewayInstanceId || '',
        })
        message.success('复制成功')
        await getRoutes()
      }
    } catch (error) {
      console.error('复制失败:', error)
      message.error('复制失败')
    } finally {
      submitting.value = false
    }
  }

  /**
   * 重置查询参数
   */
  const resetQuery = () => {
    queryParams.routeName = ''
    queryParams.routePath = ''
    queryParams.matchType = undefined
    queryParams.activeFlag = undefined
    queryParams.pageIndex = 1
  }

  /**
   * 处理选择变化
   */
  const handleSelectionChange = (keys: string[]) => {
    selectedRoutes.value = keys
  }

  /**
   * 处理页码变化
   */
  const handlePageChange = (page: number) => {
    queryParams.pageIndex = page
    getRoutes()
  }

  /**
   * 处理页面大小变化
   */
  const handlePageSizeChange = (pageSize: number) => {
    queryParams.pageSize = pageSize
    queryParams.pageIndex = 1
    getRoutes()
  }

  return {
    // 状态
    loading,
    submitting,
    routes,
    total,
    statistics,
    queryParams,
    selectedRoutes,

    // 选项
    matchTypeOptions,
    statusOptions,

    // 计算属性
    hasSelection,

    // 方法
    getRoutes,
    getStatistics,
    deleteRoute,
    batchDeleteRoutes,
    toggleRoute,
    batchToggleRoutes,
    copyRoute,
    resetQuery,
    handleSelectionChange,
    handlePageChange,
    handlePageSizeChange,
  }
}
