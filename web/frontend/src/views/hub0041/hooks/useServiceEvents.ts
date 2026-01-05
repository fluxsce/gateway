import { ref, h, computed } from 'vue'
import { 
  useMessage,
  NTag, NButton, NIcon, NTime, NTooltip, 
  NSpace, NEllipsis
} from 'naive-ui'
import { 
  TimeOutline, InformationCircleOutline,
  CloudOutline, ServerOutline, FolderOutline
} from '@vicons/ionicons5'
import { useModuleI18n } from '@/hooks/useModuleI18n'
import { usePagination } from '@/hooks/usePagination'
import { fetchServiceEventList, fetchServiceEventById } from '../api'
import { isApiSuccess, getApiMessage, parseJsonData, parsePageInfo } from '@/utils/format'
import type { ServiceEvent } from '../types'

/**
 * 服务事件相关的Hook
 * 用于获取和处理服务事件数据
 */
export function useServiceEvents() {
  // 国际化
  const { t } = useModuleI18n('hub0041')
  
  // 消息提示
  const message = useMessage()
  
  const loading = ref(false)
  const error = ref<Error | null>(null)
  
  // 事件列表数据
  const events = ref<ServiceEvent[]>([])
  const total = ref(0)

  /**
   * 获取事件类型图标
   */
  const getEventTypeIcon = (eventType: string) => {
    // 分组相关事件
    if (eventType.includes('GROUP_')) return FolderOutline
    
    // 服务相关事件
    if (eventType.includes('SERVICE_')) return CloudOutline
    
    // 实例相关事件
    if (eventType.includes('INSTANCE_')) return ServerOutline
    
    // 默认图标
    return InformationCircleOutline
  }

  /**
   * 获取事件类型颜色
   */
  const getEventTypeColor = (eventType: string) => {
    // 分组相关事件
    if (eventType === 'GROUP_CREATE') return 'success'
    if (eventType === 'GROUP_UPDATE') return 'info'
    if (eventType === 'GROUP_DELETE') return 'error'
    
    // 服务相关事件
    if (eventType === 'SERVICE_CREATE') return 'success'
    if (eventType === 'SERVICE_UPDATE') return 'info'
    if (eventType === 'SERVICE_DELETE') return 'error'
    
    // 实例相关事件
    if (eventType === 'INSTANCE_REGISTER') return 'success'
    if (eventType === 'INSTANCE_DEREGISTER') return 'error'
    if (eventType === 'INSTANCE_HEARTBEAT') return 'default'
    if (eventType === 'INSTANCE_HEALTH_CHANGE') return 'warning'
    if (eventType === 'INSTANCE_STATUS_CHANGE') return 'info'
    
    // 默认颜色
    return 'default'
  }

  // 分页回调函数引用，将在组件中设置
  let onPageChangeCallback: ((serviceId?: string, instanceId?: string) => void) | null = null
  
  // 使用分页Hook
  const {
    naiveConfig,
    queryParams,
    setTotal,
    resetPagination
  } = usePagination({
    initialPage: 1,
    initialPageSize: 10,
    onPageChange: (page: number) => {
      if (onPageChangeCallback) {
        onPageChangeCallback()
      }
    },
    onPageSizeChange: (page: number, pageSize: number) => {
      if (onPageChangeCallback) {
        onPageChangeCallback()
      }
    }
  })

  /**
   * 设置分页变化回调
   */
  const setPaginationCallback = (callback: (serviceId?: string, instanceId?: string) => void) => {
    onPageChangeCallback = callback
  }

  /**
   * 创建表格列定义
   */
  const createTableColumns = (onViewDetail: (event: ServiceEvent) => void) => [
    {
      title: t('columns.eventType'),
      key: 'eventType',
      width: 180,
      minWidth: 160,
      render(row: ServiceEvent) {
        return h(
          NTag,
          {
            type: getEventTypeColor(row.eventType),
            bordered: false,
            round: true,
            size: 'small'
          },
          {
            default: () => row.eventType,
            icon: () => h(NIcon, {}, { default: () => h(getEventTypeIcon(row.eventType)) })
          }
        )
      }
    },
    {
      title: t('columns.eventMessage'),
      key: 'eventMessage',
      minWidth: 200,
      ellipsis: {
        tooltip: {
          placement: 'bottom' as const,
          width: 'trigger' as const,
          style: { maxWidth: '400px' }
        }
      },
      render(row: ServiceEvent) {
        return row.eventMessage || '-'
      }
    },
    {
      title: t('columns.eventSource'),
      key: 'eventSource',
      width: 140,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('columns.serviceName'),
      key: 'serviceName',
      width: 120,
      ellipsis: {
        tooltip: true
      }
    },
    {
      title: t('columns.hostAddress'),
      key: 'hostAddress',
      width: 150,
      ellipsis: {
        tooltip: true
      },
      render(row: ServiceEvent) {
        return row.hostAddress ? `${row.hostAddress}:${row.portNumber || '-'}` : '-'
      }
    },
    {
      title: '服务运行节点IP',
      key: 'nodeIpAddress',
      width: 140,
      ellipsis: {
        tooltip: true
      },
      render(row: ServiceEvent) {
        return row.nodeIpAddress || '-'
      }
    },
    {
      title: t('columns.eventTime'),
      key: 'eventTime',
      width: 160,
      sorter: true,
      render(row: ServiceEvent) {
        return h(
          NTime,
          {
            time: new Date(row.eventTime),
            format: 'yyyy-MM-dd HH:mm:ss'
          }
        )
      }
    },
    {
      title: t('columns.actions'),
      key: 'actions',
      width: 80,
      fixed: 'right' as const,
      render(row: ServiceEvent) {
        return h(
          NSpace,
          { justify: 'center', size: 'small' },
          {
            default: () => [
              h(
                NTooltip,
                {},
                {
                  trigger: () => h(
                    NButton,
                    {
                      quaternary: true,
                      circle: true,
                      size: 'small',
                      onClick: () => onViewDetail(row)
                    },
                    { icon: () => h(NIcon, { size: 18 }, { default: () => h(InformationCircleOutline) }) }
                  ),
                  default: () => t('actions.view')
                }
              )
            ]
          }
        )
      }
    }
  ]

  /**
   * 查询服务事件列表
   * @param params 查询参数
   * @returns 事件列表和总数
   */
  const fetchServiceEvents = async (params: any) => {
    loading.value = true
    error.value = null
    
    try {
      const response = await fetchServiceEventList(params)
      
      if (isApiSuccess(response)) {
        // 解析业务数据
        const eventData = parseJsonData<ServiceEvent[]>(response, [])
        
        // 解析分页信息
        let totalCount = 0
        try {
          const pageInfo = parsePageInfo(response)
          totalCount = pageInfo.totalCount || 0
        } catch (parseError) {
          console.warn('解析分页信息失败:', parseError)
          // 如果分页信息解析失败，使用当前数据长度作为总数
          totalCount = eventData?.length || 0
        }
        
        return {
          data: eventData,
          totalCount: totalCount
        }
      } else {
        const errorMsg = getApiMessage(response, t('fetchServiceEventsFailed'))
        message.error(errorMsg)
        error.value = new Error(errorMsg)
        throw new Error(errorMsg)
      }
    } catch (err) {
      error.value = err as Error
      console.error('Failed to fetch service events:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  /**
   * 加载事件列表数据
   * @param serviceId 服务ID
   * @param instanceId 实例ID
   * @param customParams 自定义查询参数
   */
  const loadEvents = async (serviceId?: string, instanceId?: string, customParams?: Record<string, any>) => {
    try {
      // 查询参数构建
      const params = {
        pageIndex: queryParams.value.pageIndex,
        pageSize: queryParams.value.pageSize,
        serviceInstanceId: instanceId,
        serviceGroupId: serviceId ? serviceId : undefined,
        ...customParams // 合并自定义参数
      }

      // 调用API获取数据
      const { data, totalCount } = await fetchServiceEvents(params)
      events.value = data
      total.value = totalCount
      setTotal(totalCount)
    } catch (err) {
      message.error(t('messages.loadError'))
      console.error('加载服务事件列表失败:', err)
    }
  }

  /**
   * 获取服务事件详情
   * @param id 事件ID
   * @returns 事件详情
   */
  const getServiceEventById = async (id: string): Promise<ServiceEvent> => {
    loading.value = true
    error.value = null
    
    try {
      const response = await fetchServiceEventById(id)
      
      if (isApiSuccess(response)) {
        // 解析业务数据
        const eventData = parseJsonData<ServiceEvent>(response)
        return eventData
      } else {
        const errorMsg = getApiMessage(response, t('getServiceEventFailed'))
        message.error(errorMsg)
        error.value = new Error(errorMsg)
        throw new Error(errorMsg)
      }
    } catch (err) {
      error.value = err as Error
      console.error('Failed to get service event by id:', err)
      throw err
    } finally {
      loading.value = false
    }
  }

  return {
    // 响应式数据
    events,
    loading,
    error,
    total,
    
    // 分页相关
    naiveConfig,
    queryParams,
    setTotal,
    resetPagination,
    setPaginationCallback,
    
    // 方法
    fetchServiceEvents,
    getServiceEventById,
    loadEvents,
    createTableColumns,
    getEventTypeIcon,
    getEventTypeColor
  }
}
