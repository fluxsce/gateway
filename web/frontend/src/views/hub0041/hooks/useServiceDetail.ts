/**
 * 服务详情管理 Hook
 * 提供服务详情查询和实例管理功能
 */

import { ref } from 'vue'
import { useMessage } from 'naive-ui'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { 
  getServiceDetail, queryServiceInstances, updateInstanceHealthStatus, 
  getServiceInstance, updateServiceInstance 
} from '../api'
import { parseJsonData, safeParseJsonArray } from '@/utils/format'
import type { ServiceDetail } from '../types'

export const useServiceDetail = () => {
  // 国际化
  const { t } = useModuleI18n('hub0041')
  
  // 消息提示
  const message = useMessage()

  // 响应式数据
  const serviceDetail = ref<ServiceDetail | null>(null)
  const loading = ref(false)
  const instancesLoading = ref(false)

  /**
   * 获取服务详情
   */
  const fetchServiceDetail = async (serviceName: string) => {
    try {
      loading.value = true
      
      const response = await getServiceDetail(serviceName)
      
      if (response.oK) {
        // 使用format工具类解析服务详情数据
        try {
          serviceDetail.value = parseJsonData<ServiceDetail>(response)
        } catch (error) {
          console.error('Failed to parse service detail data:', error)
          serviceDetail.value = null
        }
      } else {
        message.error(t('fetchDetailFailed'))
        serviceDetail.value = null
      }
    } catch (error) {
      console.error('Failed to fetch service detail:', error)
      message.error(t('fetchDetailFailed'))
      serviceDetail.value = null
    } finally {
      loading.value = false
    }
  }

  /**
   * 刷新服务实例列表
   */
  const refreshServiceInstances = async (serviceName: string) => {
    if (!serviceDetail.value) return
    
    try {
      instancesLoading.value = true
      
      const response = await queryServiceInstances({ serviceName })
      
      if (response.oK) {
        try {
          // 使用format工具类解析实例数据
          const instances = safeParseJsonArray(response.bizData)
          
          // 更新实例列表
          serviceDetail.value.instances = instances
          
          // 更新统计信息
          updateInstanceStatistics()
          
          message.success(t('refreshInstancesSuccess'))
        } catch (error) {
          console.error('Failed to parse instances data:', error)
        }
      } else {
        message.error(t('refreshInstancesFailed'))
      }
    } catch (error) {
      console.error('Failed to refresh service instances:', error)
      message.error(t('refreshInstancesFailed'))
    } finally {
      instancesLoading.value = false
    }
  }

  /**
   * 更新实例统计信息
   */
  const updateInstanceStatistics = () => {
    if (!serviceDetail.value) return

    const instances = serviceDetail.value.instances
    const statistics = {
      totalInstances: instances.length,
      healthyInstances: instances.filter(i => i.healthStatus === 'HEALTHY').length,
      unhealthyInstances: instances.filter(i => i.healthStatus === 'UNHEALTHY').length,
      upInstances: instances.filter(i => i.instanceStatus === 'UP').length,
      downInstances: instances.filter(i => i.instanceStatus === 'DOWN').length
    }

    serviceDetail.value.statistics = statistics
  }

  /**
   * 执行健康检查
   */
  const performHealthCheck = async (serviceName: string, instanceId?: string) => {
    try {
      let response
      if (instanceId) {
        response = await updateInstanceHealthStatus(instanceId, 'HEALTHY')
      } else {
        // 如果没有指定实例ID，重新获取服务详情来刷新状态
        response = await getServiceDetail(serviceName)
        if (response.oK) {
          try {
            // 使用format工具类解析服务详情数据
            serviceDetail.value = parseJsonData<ServiceDetail>(response)
          } catch (error) {
            console.error('Failed to parse health check data:', error)
          }
        }
      }
      
      if (response.oK) {
        message.success(t('healthCheckSuccess'))
        // 如果是单个实例检查，刷新实例数据
        if (instanceId) {
          await refreshServiceInstances(serviceName)
        }
      } else {
        message.error(t('healthCheckFailed'))
      }
    } catch (error) {
      console.error('Failed to perform health check:', error)
      message.error(t('healthCheckFailed'))
    }
  }
  
  /**
   * 更新实例状态（上线/下线）
   */
  const updateInstanceStatus = async (serviceName: string, instance: any, status: 'UP' | 'DOWN') => {
    try {
      // 克隆实例对象以避免直接修改原对象
      const updatedInstance = { ...instance, instanceStatus: status }
      
      const response = await updateServiceInstance(updatedInstance)
      
      if (response.oK) {
        message.success(status === 'UP' ? t('bringUpSuccess') : t('takeDownSuccess'))
        // 刷新实例列表
        await refreshServiceInstances(serviceName)
      } else {
        message.error(status === 'UP' ? t('bringUpFailed') : t('takeDownFailed'))
      }
    } catch (error) {
      console.error(`Failed to ${status === 'UP' ? 'bring up' : 'take down'} instance:`, error)
      message.error(status === 'UP' ? t('bringUpFailed') : t('takeDownFailed'))
    }
  }

  /**
   * 获取实例元数据
   */
  const getInstanceMetadata = async (instanceId: string) => {
    try {
      const response = await getServiceInstance(instanceId)
      
      if (response.oK) {
        try {
          // 使用format工具类解析实例元数据
          return parseJsonData(response, null)
        } catch (error) {
          console.error('Failed to parse metadata:', error)
          return null
        }
      } else {
        message.error(t('fetchMetadataFailed'))
        return null
      }
    } catch (error) {
      console.error('Failed to fetch instance metadata:', error)
      message.error(t('fetchMetadataFailed'))
      return null
    }
  }

  /**
   * 清空服务详情
   */
  const clearServiceDetail = () => {
    serviceDetail.value = null
  }

  return {
    // 数据
    serviceDetail,
    loading,
    instancesLoading,
    
    // 方法
    fetchServiceDetail,
    refreshServiceInstances,
    performHealthCheck,
    updateInstanceStatus,
    getInstanceMetadata,
    clearServiceDetail
  }
}
