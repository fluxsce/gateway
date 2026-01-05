/**
 * JVM资源管理Hook
 * 提供JVM资源列表查询、详情获取等功能
 */

import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { isApiSuccess, getApiMessage, parseJsonData, parsePageInfo } from '@/utils/format'
import type { 
  JvmResource, 
  JvmResourceQueryRequest,
  JvmMonitorDetail
} from '../types'
import * as api from '../api'

export function useJvmResourceManagement() {
  const message = useMessage()
  
  // 状态
  const loading = ref(false)
  const jvmResources = ref<JvmResource[]>([])
  const currentResource = ref<JvmResource | null>(null)
  const monitorDetail = ref<JvmMonitorDetail | null>(null)
  const total = ref(0)
  
  // 查询参数
  const queryParams = ref<JvmResourceQueryRequest>({
    pageNum: 1,
    pageSize: 20
  })
  
  /**
   * 查询JVM资源列表
   */
  const queryJvmResources = async (params?: Partial<JvmResourceQueryRequest>) => {
    loading.value = true
    try {
      if (params) {
        queryParams.value = { ...queryParams.value, ...params }
      }
      
      const result = await api.queryJvmResources(queryParams.value)
      
      if (isApiSuccess(result)) {
        jvmResources.value = parseJsonData<JvmResource[]>(result, [])
        
        // 解析分页信息
        try {
          const pageInfo = parsePageInfo(result)
          total.value = pageInfo.totalCount || 0
        } catch (error) {
          console.warn('解析分页信息失败:', error)
          total.value = jvmResources.value.length
        }
      } else {
        message.error(getApiMessage(result, '查询JVM资源列表失败'))
      }
    } catch (error) {
      console.error('查询JVM资源列表失败:', error)
      message.error('查询JVM资源列表失败')
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 获取JVM资源详情
   */
  const getJvmResource = async (jvmResourceId: string) => {
    loading.value = true
    try {
      const result = await api.getJvmResource(jvmResourceId)
      
      if (isApiSuccess(result)) {
        const data = parseJsonData<JvmResource>(result, null as any)
        currentResource.value = data
        return data
      } else {
        message.error(getApiMessage(result, '获取JVM资源详情失败'))
        return null
      }
    } catch (error) {
      console.error('获取JVM资源详情失败:', error)
      message.error('获取JVM资源详情失败')
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 获取JVM完整监控详情
   */
  const getJvmMonitorDetail = async (jvmResourceId: string) => {
    loading.value = true
    try {
      const result = await api.getJvmOverview(jvmResourceId)
      
      if (isApiSuccess(result)) {
        const data = parseJsonData<JvmMonitorDetail>(result, null as any)
        monitorDetail.value = data
        return data
      } else {
        message.error(getApiMessage(result, '获取JVM监控详情失败'))
        return null
      }
    } catch (error) {
      console.error('获取JVM监控详情失败:', error)
      message.error('获取JVM监控详情失败')
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 刷新列表
   */
  const refresh = () => {
    queryJvmResources()
  }
  
  /**
   * 重置查询参数
   */
  const resetQuery = () => {
    queryParams.value = {
      pageNum: 1,
      pageSize: 20
    }
    queryJvmResources()
  }
  
  // 计算属性
  const hasData = computed(() => jvmResources.value.length > 0)
  const healthyCount = computed(() => 
    jvmResources.value.filter(r => r.healthyFlag === 'Y').length
  )
  const unhealthyCount = computed(() => 
    jvmResources.value.filter(r => r.healthyFlag === 'N').length
  )
  const attentionCount = computed(() => 
    jvmResources.value.filter(r => r.requiresAttentionFlag === 'Y').length
  )
  
  return {
    // 状态
    loading,
    jvmResources,
    currentResource,
    monitorDetail,
    total,
    queryParams,
    
    // 计算属性
    hasData,
    healthyCount,
    unhealthyCount,
    attentionCount,
    
    // 方法
    queryJvmResources,
    getJvmResource,
    getJvmMonitorDetail,
    refresh,
    resetQuery
  }
}

