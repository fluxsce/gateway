/**
 * 后端日志对话框页面级 Hook
 * - 处理后端日志的加载、状态管理和辅助方法
 */

import { formatDate, formatFileSize, getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { useMessage } from 'naive-ui'
import { computed, ref, watch } from 'vue'
import { getGatewayLog } from '../../api'
import type { GatewayLogInfo } from '../../types'

/**
 * 后端服务追踪日志接口
 * 对应表结构：HUB_GW_BACKEND_TRACE_LOG
 */
export interface BackendTraceLog {
  /** 租户ID */
  tenantId: string
  /** 链路追踪ID */
  traceId: string
  /** 后端服务追踪ID */
  backendTraceId: string
  /** 服务定义ID */
  serviceDefinitionId?: string
  /** 服务名称 */
  serviceName?: string
  /** 转发地址 */
  forwardAddress?: string
  /** 转发方法 */
  forwardMethod?: string
  /** 转发路径 */
  forwardPath?: string
  /** 转发查询参数 */
  forwardQuery?: string
  /** 转发头信息 */
  forwardHeaders?: string
  /** 转发体 */
  forwardBody?: string
  /** 请求大小 */
  requestSize?: number
  /** 负载均衡策略 */
  loadBalancerStrategy?: string
  /** 负载均衡决策 */
  loadBalancerDecision?: string
  /** 请求开始时间 */
  requestStartTime?: string
  /** 响应接收时间 */
  responseReceivedTime?: string
  /** 请求耗时(毫秒) */
  requestDurationMs?: number
  /** 状态码 */
  statusCode?: number
  /** 响应大小 */
  responseSize?: number
  /** 响应头信息 */
  responseHeaders?: string
  /** 响应体 */
  responseBody?: string
  /** 错误代码 */
  errorCode?: string
  /** 错误信息 */
  errorMessage?: string
  /** 成功标记 */
  successFlag?: string
  /** 追踪状态 */
  traceStatus?: string
  /** 重试次数 */
  retryCount?: number
  /** 扩展属性 */
  extProperty?: string
  /** 创建时间 */
  addTime?: string
  /** 创建人 */
  addWho?: string
  /** 活动状态 */
  activeFlag?: string
  /** 备注信息 */
  noteText?: string
}

/**
 * 网关日志详情接口（包含后端追踪日志）
 * 主信息（GatewayLogInfo的所有字段）+ 后端服务信息（BackendTraces数组）
 */
interface GatewayLogDetailWithBackendTraces extends GatewayLogInfo {
  backendTraces?: BackendTraceLog[]
}

/**
 * 后端日志对话框页面级 Hook
 */
export function useBackendLogsPage(
  props: { visible: boolean; traceId?: string },
  emit: (event: 'update:visible', value: boolean) => void
) {
  const message = useMessage()

  // 状态管理
  const loading = ref(false)
  const gatewayLogInfo = ref<GatewayLogInfo | null>(null)
  const backendTraces = ref<BackendTraceLog[]>([])
  const activeTab = ref<string>('basic')

  // 计算属性
  const dialogTitle = computed(() => {
    return props.traceId ? `后端日志详情 - ${props.traceId}` : '后端日志详情'
  })

  // 使用计算属性来管理模态框的显示状态
  const showModal = computed({
    get() {
      return props.visible
    },
    set(value: boolean) {
      emit('update:visible', value)
    },
  })

  // 监听visible变化
  watch(
    () => props.visible,
    (val) => {
      if (val && props.traceId) {
        loadBackendLogs()
      }
    },
    { immediate: true }
  )

  // 加载后端日志
  const loadBackendLogs = async () => {
    if (!props.traceId) {
      return
    }

    try {
      loading.value = true
      const response = await getGatewayLog({
        traceId: props.traceId,
      })

      if (isApiSuccess(response)) {
        const data = parseJsonData<GatewayLogDetailWithBackendTraces>(response)
        // 保存主表信息（外部请求信息）
        gatewayLogInfo.value = data
        // 保存后端追踪日志
        backendTraces.value = data.backendTraces || []
        // 默认显示基础信息 tab
        activeTab.value = 'basic'
      } else {
        const errorMsg = getApiMessage(response, '获取后端日志失败')
        message.error(errorMsg)
        gatewayLogInfo.value = null
        backendTraces.value = []
      }
    } catch (error) {
      console.error('获取后端日志失败:', error)
      message.error('获取后端日志失败')
      backendTraces.value = []
    } finally {
      loading.value = false
    }
  }

  // 获取服务标签名称
  const getServiceTabName = (trace: BackendTraceLog, index: number): string => {
    if (trace.serviceName) {
      return trace.serviceName
    }
    return `服务${index + 1}`
  }

  // 格式化JSON数据
  const formatJsonData = (data: string | object): string => {
    if (typeof data === 'string') {
      try {
        return JSON.stringify(JSON.parse(data), null, 2)
      } catch {
        return data
      }
    }
    return JSON.stringify(data, null, 2)
  }

  // 获取HTTP方法颜色
  const getMethodColor = (method?: string): string => {
    const colorMap: Record<string, string> = {
      GET: 'info',
      POST: 'success',
      PUT: 'warning',
      DELETE: 'error',
      PATCH: 'default',
      HEAD: 'default',
      OPTIONS: 'default',
    }
    return colorMap[method || ''] || 'default'
  }

  // 获取状态码类型
  const getStatusCodeType = (statusCode: number): string => {
    if (!statusCode || statusCode === 0) return 'default'
    if (statusCode >= 200 && statusCode < 300) return 'success'
    if (statusCode >= 300 && statusCode < 400) return 'warning'
    if (statusCode >= 400) return 'error'
    return 'default'
  }

  // 获取响应时间类型
  const getResponseTimeType = (responseTime: number): string => {
    if (responseTime === 0) return 'default'
    if (responseTime < 100) return 'success'
    if (responseTime < 500) return 'warning'
    return 'error'
  }

  // 获取追踪状态类型
  const getTraceStatusType = (status?: string): string => {
    const statusMap: Record<string, string> = {
      pending: 'warning',
      success: 'success',
      failed: 'error',
      timeout: 'error',
    }
    return statusMap[status || ''] || 'default'
  }

  // 获取追踪状态文本
  const getTraceStatusText = (status?: string): string => {
    const statusMap: Record<string, string> = {
      pending: '处理中',
      success: '成功',
      failed: '失败',
      timeout: '超时',
    }
    return statusMap[status || ''] || status || '未知'
  }

  // 关闭弹窗
  const handleCancel = () => {
    showModal.value = false
  }

  // 弹窗关闭后的处理
  const handleAfterLeave = () => {
    gatewayLogInfo.value = null
    backendTraces.value = []
    activeTab.value = 'basic'
  }

  // 获取日志级别类型
  const getLogLevelType = (level: string): string => {
    const typeMap: Record<string, string> = {
      ERROR: 'error',
      WARN: 'warning',
      INFO: 'info',
      DEBUG: 'success',
    }
    return typeMap[level] || 'default'
  }

  // 获取日志级别文本
  const getLogLevelText = (level: string): string => {
    const textMap: Record<string, string> = {
      ERROR: '错误',
      WARN: '警告',
      INFO: '信息',
      DEBUG: '调试',
    }
    return textMap[level] || level
  }

  // 获取日志类型颜色
  const getLogTypeColor = (type: string): string => {
    const colorMap: Record<string, string> = {
      ACCESS: 'info',
      ERROR: 'error',
      SYSTEM: 'warning',
    }
    return colorMap[type] || 'default'
  }

  // 获取日志类型文本
  const getLogTypeText = (type: string): string => {
    const textMap: Record<string, string> = {
      ACCESS: '访问日志',
      ERROR: '错误日志',
      SYSTEM: '系统日志',
    }
    return textMap[type] || type
  }

  // 获取代理类型颜色
  const getProxyTypeColor = (type?: string): string => {
    const colorMap: Record<string, string> = {
      http: 'info',
      websocket: 'warning',
      tcp: 'success',
      udp: 'error',
    }
    return colorMap[type || ''] || 'default'
  }

  // 获取代理类型文本
  const getProxyTypeText = (type?: string): string => {
    const textMap: Record<string, string> = {
      http: 'HTTP',
      websocket: 'WebSocket',
      tcp: 'TCP',
      udp: 'UDP',
    }
    return textMap[type || ''] || type || '未知'
  }

  return {
    // 状态
    loading,
    gatewayLogInfo,
    backendTraces,
    activeTab,
    showModal,

    // 计算属性
    dialogTitle,

    // 方法
    getServiceTabName,
    formatJsonData,
    getMethodColor,
    getStatusCodeType,
    getResponseTimeType,
    getTraceStatusType,
    getTraceStatusText,
    getLogLevelType,
    getLogLevelText,
    getLogTypeColor,
    getLogTypeText,
    getProxyTypeColor,
    getProxyTypeText,
    formatDate,
    formatFileSize,

    // 事件处理
    handleCancel,
    handleAfterLeave,
  }
}

export type BackendLogsPage = ReturnType<typeof useBackendLogsPage>

