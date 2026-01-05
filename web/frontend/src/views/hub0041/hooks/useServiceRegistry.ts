/**
 * 服务注册中心管理 Hook
 * 提供服务列表查询、搜索、分页等功能
 */

import { ref, reactive, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { usePagination } from '@/hooks/usePagination'
import { queryServices, getService, deleteService } from '../api'
import { isApiSuccess, getApiMessage, parseJsonData, parsePageInfo } from '@/utils/format'
import type { Service, ServiceQueryRequest } from '../types'

export const useServiceRegistry = () => {
  // 国际化
  const { t } = useModuleI18n('hub0041')
  
  // 消息提示
  const message = useMessage()

  // 响应式数据
  const services = ref<Service[]>([])
  const loading = ref(false)
  const total = ref(0)
  


  // 搜索参数
  const searchParams = reactive<Partial<ServiceQueryRequest>>({
    serviceName: '',
    groupName: '',
    protocolType: undefined,
    activeFlag: undefined
  })

  // 使用分页Hook
  const {
    naiveConfig,
    queryParams,
    setTotal,
    resetPagination
  } = usePagination({
    initialPage: 1,
    initialPageSize: 12,
    onPageChange: (page: number) => {
      fetchServices()
    },
    onPageSizeChange: (page: number, pageSize: number) => {
      fetchServices()
    }
  })

  // 计算属性
  const totalCount = computed(() => total.value)

  /**
   * 获取服务列表
   */
  const fetchServices = async () => {
    try {
      loading.value = true
      
      const requestParams: ServiceQueryRequest = {
        ...searchParams,
        pageIndex: queryParams.value.pageIndex,
        pageSize: queryParams.value.pageSize
      }

      const response = await queryServices(requestParams)
      
      if (isApiSuccess(response)) {
        // 解析业务数据 - bizData直接是Service数组
        const serviceList = parseJsonData<Service[]>(response, [])
        services.value = serviceList || []
        
        // 解析分页信息
        let totalCount = 0
        try {
          const pageInfo = parsePageInfo(response)
          totalCount = pageInfo.totalCount || 0
        } catch (error) {
          console.warn('解析分页信息失败:', error)
          // 如果分页信息解析失败，使用当前数据长度作为总数
          totalCount = serviceList?.length || 0
        }
        
        total.value = totalCount
        setTotal(totalCount)
      } else {
        message.error(getApiMessage(response, t('fetchServicesFailed')))
        services.value = []
        total.value = 0
        setTotal(0)
      }
    } catch (error) {
      console.error('Failed to fetch services:', error)
      message.error(t('fetchServicesFailed'))
      services.value = []
      total.value = 0
      setTotal(0)
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新单个服务状态
   */
  const refreshService = async (serviceName: string) => {
    try {
      // 通过重新获取服务信息来刷新状态
      const response = await getService(serviceName)
      if (isApiSuccess(response)) {
        message.success(t('refreshServiceSuccess'))
        // 重新获取服务列表以更新UI
        await fetchServices()
      } else {
        message.error(getApiMessage(response, t('refreshServiceFailed')))
      }
    } catch (error) {
      console.error('Failed to refresh service:', error)
      message.error(t('refreshServiceFailed'))
    }
  }





  /**
   * 搜索处理
   */
  const handleSearch = () => {
    // 重置到第一页
    resetPagination()
    fetchServices()
  }

  /**
   * 重置搜索条件
   */
  const handleReset = () => {
    Object.assign(searchParams, {
      serviceName: '',
      groupName: '',
      protocolType: undefined,
      activeFlag: undefined
    })
    resetPagination()
    fetchServices()
  }

  /**
   * 刷新数据
   */
  const handleRefresh = () => {
    fetchServices()
  }

  /**
   * 删除服务
   * @param serviceName 服务名称
   * @returns 删除结果的Promise
   */
  const handleDeleteService = async (serviceName: string) => {
    try {
      const response = await deleteService(serviceName)
      if (isApiSuccess(response)) {
        message.success(t('deleteServiceSuccess'))
        // 刷新服务列表
        await fetchServices()
        return { success: true, message: t('deleteServiceSuccess') }
      } else {
        const errorMsg = getApiMessage(response, t('deleteServiceFailed'))
        message.error(errorMsg)
        return { success: false, message: errorMsg }
      }
    } catch (error) {
      console.error('Failed to delete service:', error)
      const errorMsg = t('deleteServiceFailed')
      message.error(errorMsg)
      return { success: false, message: errorMsg }
    }
  }

  return {
    // 响应式数据
    services,
    loading,
    searchParams,
    naiveConfig,
    
    // 计算属性  
    totalCount,
    
    // 方法
    fetchServices,
    refreshService,
    handleSearch,
    handleReset,
    handleRefresh,
    handleDeleteService
  }
}
