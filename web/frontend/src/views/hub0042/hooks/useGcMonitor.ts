/**
 * GC监控Hook
 * 提供GC快照、GC趋势等监控功能
 */

import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { isApiSuccess, getApiMessage, parseJsonData } from '@/utils/format'
import type { 
  JvmGc, 
  GcQueryRequest,
  GcTrendData
} from '../types'
import * as api from '../api'

export function useGcMonitor() {
  const message = useMessage()
  
  // 状态
  const loading = ref(false)
  const gcSnapshots = ref<JvmGc[]>([])
  const latestGcSnapshot = ref<JvmGc | null>(null)
  const gcTrend = ref<GcTrendData[]>([])
  
  /**
   * 查询GC快照列表
   */
  const queryGcSnapshots = async (params: GcQueryRequest) => {
    loading.value = true
    try {
      const result = await api.queryGcSnapshots(params)
      
      if (isApiSuccess(result)) {
        gcSnapshots.value = parseJsonData<JvmGc[]>(result, [])
      } else {
        message.error(getApiMessage(result, '查询GC快照失败'))
      }
    } catch (error) {
      console.error('查询GC快照失败:', error)
      message.error('查询GC快照失败')
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 获取最新GC快照
   */
  const getLatestGcSnapshot = async (jvmResourceId: string) => {
    loading.value = true
    try {
      const result = await api.getLatestGcSnapshot(jvmResourceId)
      
      if (isApiSuccess(result)) {
        const data = parseJsonData<JvmGc>(result, null as any)
        latestGcSnapshot.value = data
        return data
      } else {
        message.error(getApiMessage(result, '获取最新GC快照失败'))
        return null
      }
    } catch (error) {
      console.error('获取最新GC快照失败:', error)
      message.error('获取最新GC快照失败')
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 获取GC趋势数据
   */
  const getGcTrend = async (params: GcQueryRequest) => {
    loading.value = true
    try {
      const result = await api.getGcTrend(params)
      
      if (isApiSuccess(result)) {
        gcTrend.value = parseJsonData<GcTrendData[]>(result, [])
      } else {
        message.error(getApiMessage(result, '获取GC趋势失败'))
      }
    } catch (error) {
      console.error('获取GC趋势失败:', error)
      message.error('获取GC趋势失败')
    } finally {
      loading.value = false
    }
  }
  
  // 计算属性
  const totalGcCount = computed(() => 
    latestGcSnapshot.value?.collectionCount || 0
  )
  
  const totalGcTime = computed(() => 
    latestGcSnapshot.value?.collectionTimeMs || 0
  )
  
  const youngGcCount = computed(() => 
    latestGcSnapshot.value?.ygc || 0
  )
  
  const fullGcCount = computed(() => 
    latestGcSnapshot.value?.fgc || 0
  )
  
  const avgGcTime = computed(() => {
    if (!latestGcSnapshot.value || latestGcSnapshot.value.collectionCount === 0) {
      return 0
    }
    return latestGcSnapshot.value.collectionTimeMs / latestGcSnapshot.value.collectionCount
  })
  
  /**
   * 格式化GC时间（毫秒转为秒）
   */
  const formatGcTime = (ms: number): string => {
    return (ms / 1000).toFixed(3) + 's'
  }
  
  /**
   * 计算内存使用率
   */
  const calculateUsagePercent = (used: number, capacity: number): number => {
    if (capacity === 0) return 0
    return Number(((used / capacity) * 100).toFixed(2))
  }
  
  return {
    // 状态
    loading,
    gcSnapshots,
    latestGcSnapshot,
    gcTrend,
    
    // 计算属性
    totalGcCount,
    totalGcTime,
    youngGcCount,
    fullGcCount,
    avgGcTime,
    
    // 方法
    queryGcSnapshots,
    getLatestGcSnapshot,
    getGcTrend,
    formatGcTime,
    calculateUsagePercent
  }
}

