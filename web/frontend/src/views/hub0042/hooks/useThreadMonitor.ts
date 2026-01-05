/**
 * 线程监控Hook
 * 
 * 提供线程信息、线程状态、死锁检测等监控功能
 * 
 * @example
 * ```typescript
 * const {
 *   loading,
 *   threadInfo,
 *   threadTrendData,
 *   queryThreadInfo,
 *   queryThreadStates
 * } = useThreadMonitor()
 * 
 * // 查询时间范围内的线程信息
 * await queryThreadInfo({
 *   jvmResourceId: 'xxx',
 *   startTime: '2025-10-12T00:00:00Z',
 *   endTime: '2025-10-12T23:59:59Z'
 * })
 * ```
 * 
 * @returns {Object} 线程监控相关的状态和方法
 */

import { ref, computed } from 'vue'
import { useMessage } from 'naive-ui'
import { isApiSuccess, getApiMessage, parseJsonData } from '@/utils/format'
import type { 
  JvmThread,
  JvmThreadState,
  JvmDeadlock,
  ThreadQueryRequest,
  ThreadTrendData,
  ThreadStateTrendData
} from '../types'
import * as api from '../api'

export function useThreadMonitor() {
  const message = useMessage()
  
  // ==================== 状态定义 ====================
  
  /** 加载状态 */
  const loading = ref(false)
  
  /** 单条线程信息（最新的） */
  const threadInfo = ref<JvmThread | null>(null)
  
  /** 线程信息列表（时间序列） */
  const threadInfoList = ref<JvmThread[]>([])
  
  /** 单条线程状态（最新的） */
  const threadState = ref<JvmThreadState | null>(null)
  
  /** 线程状态列表（时间序列） */
  const threadStateList = ref<JvmThreadState[]>([])
  
  /** 死锁检测记录列表 */
  const deadlocks = ref<JvmDeadlock[]>([])
  
  // ==================== API 方法 ====================
  
  /**
   * 查询线程信息
   * 
   * 支持时间范围查询，返回时间序列数据
   * 
   * @param params - 查询参数
   * @param params.jvmResourceId - JVM资源ID
   * @param params.startTime - 开始时间（可选，ISO格式）
   * @param params.endTime - 结束时间（可选，ISO格式）
   * 
   * @example
   * ```typescript
   * // 查询最新单条
   * await queryThreadInfo({ jvmResourceId: 'xxx' })
   * 
   * // 查询时间范围
   * await queryThreadInfo({
   *   jvmResourceId: 'xxx',
   *   startTime: '2025-10-12T00:00:00Z',
   *   endTime: '2025-10-12T23:59:59Z'
   * })
   * ```
   */
  const queryThreadInfo = async (params: ThreadQueryRequest) => {
    loading.value = true
    try {
      const result = await api.queryThreadInfo(params)
      
      if (isApiSuccess(result)) {
        const data = parseJsonData<JvmThread[]>(result, [])
        
        // 如果有时间范围，返回列表；否则返回单条
        if (params.startTime && params.endTime) {
          threadInfoList.value = data
        } else {
          threadInfo.value = Array.isArray(data) && data.length > 0 ? data[0] : (data as any)
        }
      } else {
        message.error(getApiMessage(result, '查询线程信息失败'))
      }
    } catch (error) {
      console.error('查询线程信息失败:', error)
      message.error('查询线程信息失败')
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 获取线程状态统计（单条）
   * 
   * 根据线程记录ID获取对应的线程状态详细信息
   * 
   * @param jvmThreadId - JVM线程记录ID
   * @returns 线程状态信息，失败返回null
   * 
   * @example
   * ```typescript
   * const state = await getThreadState('A25G81A99B10A39A09C94E98A45H08J7')
   * if (state) {
   *   console.log('RUNNABLE线程数:', state.runnableThreadCount)
   * }
   * ```
   */
  const getThreadState = async (jvmThreadId: string) => {
    loading.value = true
    try {
      const result = await api.getThreadState(jvmThreadId)
      
      if (isApiSuccess(result)) {
        const data = parseJsonData<JvmThreadState>(result, null as any)
        threadState.value = data
        return data
      } else {
        message.error(getApiMessage(result, '获取线程状态失败'))
        return null
      }
    } catch (error) {
      console.error('获取线程状态失败:', error)
      message.error('获取线程状态失败')
      return null
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 查询线程状态列表（批量查询）
   * 
   * 支持时间范围查询，一次性获取多条线程状态记录
   * 相比循环调用 getThreadState，性能提升显著
   * 
   * @param params - 查询参数
   * @param params.jvmResourceId - JVM资源ID
   * @param params.startTime - 开始时间（ISO格式）
   * @param params.endTime - 结束时间（ISO格式）
   * 
   * @example
   * ```typescript
   * // 批量查询2小时内的线程状态
   * await queryThreadStates({
   *   jvmResourceId: 'xxx',
   *   startTime: '2025-10-12T14:00:00Z',
   *   endTime: '2025-10-12T16:00:00Z'
   * })
   * 
   * // threadStateList.value 现在包含该时间段的所有状态记录
   * ```
   */
  const queryThreadStates = async (params: ThreadQueryRequest) => {
    loading.value = true
    try {
      const result = await api.queryThreadStates(params)
      
      if (isApiSuccess(result)) {
        threadStateList.value = parseJsonData<JvmThreadState[]>(result, [])
      } else {
        message.error(getApiMessage(result, '查询线程状态列表失败'))
      }
    } catch (error) {
      console.error('查询线程状态列表失败:', error)
      message.error('查询线程状态列表失败')
    } finally {
      loading.value = false
    }
  }
  
  /**
   * 查询死锁检测信息
   * 
   * 获取指定时间范围内的死锁检测记录
   * 
   * @param params - 查询参数
   * @param params.jvmResourceId - JVM资源ID
   * @param params.startTime - 开始时间（可选，ISO格式）
   * @param params.endTime - 结束时间（可选，ISO格式）
   * 
   * @example
   * ```typescript
   * await queryDeadlocks({
   *   jvmResourceId: 'xxx',
   *   startTime: '2025-10-12T00:00:00Z',
   *   endTime: '2025-10-12T23:59:59Z'
   * })
   * 
   * // 检查是否有死锁
   * if (hasDeadlock.value) {
   *   console.log('检测到死锁！')
   * }
   * ```
   */
  const queryDeadlocks = async (params: ThreadQueryRequest) => {
    loading.value = true
    try {
      const result = await api.queryDeadlocks(params)
      
      if (isApiSuccess(result)) {
        deadlocks.value = parseJsonData<JvmDeadlock[]>(result, [])
      } else {
        message.error(getApiMessage(result, '查询死锁信息失败'))
      }
    } catch (error) {
      console.error('查询死锁信息失败:', error)
      message.error('查询死锁信息失败')
    } finally {
      loading.value = false
    }
  }
  
  // ==================== 计算属性 ====================
  
  /**
   * 是否存在死锁
   * 
   * 检查死锁记录列表中是否有任何一条记录标记为存在死锁
   * 
   * @returns true表示存在死锁，false表示无死锁
   */
  const hasDeadlock = computed(() => 
    deadlocks.value.some(d => d.hasDeadlockFlag === 'Y')
  )
  
  /**
   * 活跃线程数
   * 
   * 返回当前处于RUNNABLE状态的线程数量
   * 
   * @returns RUNNABLE状态的线程数
   */
  const activeThreads = computed(() => 
    threadState.value?.runnableThreadCount || 0
  )
  
  /**
   * 阻塞线程数
   * 
   * 返回当前处于BLOCKED状态的线程数量
   * 
   * @returns BLOCKED状态的线程数
   */
  const blockedThreads = computed(() => 
    threadState.value?.blockedThreadCount || 0
  )
  
  /**
   * 等待线程数
   * 
   * 返回处于等待状态的线程总数（WAITING + TIMED_WAITING）
   * 
   * @returns 等待状态的线程总数
   */
  const waitingThreads = computed(() => 
    (threadState.value?.waitingThreadCount || 0) + (threadState.value?.timedWaitingThreadCount || 0)
  )
  
  /**
   * 线程状态分布
   * 
   * 返回当前各种状态的线程数量和对应的显示类型
   * 用于UI展示
   * 
   * @returns 线程状态分布数组
   */
  const threadStateDistribution = computed(() => {
    if (!threadState.value) return []
    
    return [
      { name: 'NEW', value: threadState.value.newThreadCount, type: 'info' as const },
      { name: 'RUNNABLE', value: threadState.value.runnableThreadCount, type: 'success' as const },
      { name: 'BLOCKED', value: threadState.value.blockedThreadCount, type: 'error' as const },
      { name: 'WAITING', value: threadState.value.waitingThreadCount, type: 'warning' as const },
      { name: 'TIMED_WAITING', value: threadState.value.timedWaitingThreadCount, type: 'warning' as const },
      { name: 'TERMINATED', value: threadState.value.terminatedThreadCount, type: 'default' as const }
    ]
  })
  
  /**
   * 线程趋势数据
   * 
   * 合并线程基础信息和状态信息，生成用于图表展示的时间序列数据
   * 
   * **注意**：此计算属性假设 threadInfoList 和 threadStateList 的索引是对应的
   * 即 threadInfoList[i] 和 threadStateList[i] 是同一时间点的数据
   * 
   * @returns 线程趋势数据数组，包含每个时间点的线程数量和比例信息
   * 
   * @example
   * ```typescript
   * // 用于线程数量趋势图
   * threadTrendData.value.forEach(data => {
   *   console.log(data.time, data.currentThreadCount)
   * })
   * ```
   */
  const threadTrendData = computed<ThreadTrendData[]>(() => {
    if (!threadInfoList.value || !threadStateList.value || 
        !Array.isArray(threadInfoList.value) || !Array.isArray(threadStateList.value) ||
        threadInfoList.value.length === 0 || threadStateList.value.length === 0) {
      return []
    }
    
    // 合并线程信息和状态数据
    // 假设两个数组的索引是对应的（同一时间点）
    return threadInfoList.value.map((info, index) => {
      const state = threadStateList.value[index]
      
      return {
        time: info?.collectionTime || '',
        currentThreadCount: info?.currentThreadCount || 0,
        peakThreadCount: info?.peakThreadCount || 0,
        daemonThreadCount: info?.daemonThreadCount || 0,
        userThreadCount: info?.userThreadCount || 0,
        activeThreadCount: state?.runnableThreadCount || 0,
        blockedThreadCount: state?.blockedThreadCount || 0,
        waitingThreadCount: (state?.waitingThreadCount || 0) + (state?.timedWaitingThreadCount || 0),
        activeThreadRatioPercent: state?.activeThreadRatioPercent || 0,
        blockedThreadRatioPercent: state?.blockedThreadRatioPercent || 0,
        waitingThreadRatioPercent: state?.waitingThreadRatioPercent || 0
      }
    })
  })
  
  /**
   * 线程状态趋势数据
   * 
   * 提取线程状态列表中的关键字段，生成用于状态分布图表的时间序列数据
   * 
   * **数据来源**：直接从后端返回的 JvmThreadState 列表映射而来
   * 
   * **TERMINATED 计算说明**：
   * - terminatedThreadCount 直接来自后端 HUB_MONITOR_JVM_THR_STATE 表
   * - 该值由后端通过 ThreadMXBean.getThreadInfo() 统计得出
   * - 表示在采集时刻处于 TERMINATED 状态的线程数
   * - 注意：TERMINATED 状态的线程通常很快被GC回收，所以这个值通常较小或为0
   * 
   * @returns 线程状态趋势数据数组，包含每个时间点的6种状态的线程数
   * 
   * @example
   * ```typescript
   * // 用于线程状态分布堆叠图
   * threadStateTrendData.value.forEach(data => {
   *   console.log(data.time)
   *   console.log('NEW:', data.newThreadCount)
   *   console.log('RUNNABLE:', data.runnableThreadCount)
   *   console.log('BLOCKED:', data.blockedThreadCount)
   *   console.log('WAITING:', data.waitingThreadCount)
   *   console.log('TIMED_WAITING:', data.timedWaitingThreadCount)
   *   console.log('TERMINATED:', data.terminatedThreadCount)
   *   console.log('总计:', data.totalThreadCount)
   * })
   * ```
   */
  const threadStateTrendData = computed<ThreadStateTrendData[]>(() => {
    if (!threadStateList.value || !Array.isArray(threadStateList.value)) {
      return []
    }
    
    return threadStateList.value.map(state => ({
      time: state?.collectionTime || '',
      newThreadCount: state?.newThreadCount || 0,
      runnableThreadCount: state?.runnableThreadCount || 0,
      blockedThreadCount: state?.blockedThreadCount || 0,
      waitingThreadCount: state?.waitingThreadCount || 0,
      timedWaitingThreadCount: state?.timedWaitingThreadCount || 0,
      terminatedThreadCount: state?.terminatedThreadCount || 0,
      totalThreadCount: state?.totalThreadCount || 0
    }))
  })
  
  return {
    // 状态
    loading,
    threadInfo,
    threadInfoList,
    threadState,
    threadStateList,
    deadlocks,
    
    // 计算属性
    hasDeadlock,
    activeThreads,
    blockedThreads,
    waitingThreads,
    threadStateDistribution,
    threadTrendData,
    threadStateTrendData,
    
    // 方法
    queryThreadInfo,
    getThreadState,
    queryThreadStates,
    queryDeadlocks
  }
}

