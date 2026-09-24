import { useAppMessage } from '@/composables/useAppMessage'
import { isApiSuccess, parseJsonData, parsePageInfo } from '@/utils/format'
import { createBackendPaginationParams } from '@/utils/pagination'
import { computed, ref } from 'vue'
import { queryAllGatewayInstances } from '../api'
import type { GatewayInstance, ProxyType } from '../components/instance-tree'

/**
 * 网关实例管理Hook
 * 提供网关实例的加载、选择等功能
 */
export function useGatewayInstance() {
  const message = useAppMessage()

  // 网关实例状态
  const loadingInstances = ref(false)
  const instanceList = ref<GatewayInstance[]>([])
  const instanceTotal = ref(0)
  const selectedInstanceId = ref<string>('')
  const selectedInstance = ref<GatewayInstance | null>(null)
  const instanceDetailsVisible = ref(false) // 默认收起实例详情
  const instanceProxyType = ref<ProxyType | null>(null)

  // 网关实例ID - 从选择的实例获取
  const gatewayInstanceId = computed(() => selectedInstanceId.value || '')

  // 计算属性：检查是否已选择网关实例
  const hasSelectedInstance = computed(() => Boolean(selectedInstanceId.value))

  // 网关实例选项
  const instanceOptions = computed(() => {
    return instanceList.value.map((instance) => {
      // 根据TLS状态选择显示的端口
      const port = instance.tlsEnabled === 'Y' ? instance.httpsPort : instance.httpPort

      return {
        label: `${instance.instanceName || '未命名'} (${instance.bindAddress || '-'}:${port || '-'})`,
        value: instance.gatewayInstanceId,
        disabled: instance.activeFlag !== 'Y',
      }
    })
  })

  // 切换实例详情显示
  function toggleInstanceDetails() {
    instanceDetailsVisible.value = !instanceDetailsVisible.value
  }

  // 加载网关实例列表
  async function loadGatewayInstances() {
    try {
      loadingInstances.value = true
      const page = createBackendPaginationParams(1, undefined)
      const res = await queryAllGatewayInstances({
        activeFlag: 'Y',
        pageIndex: page.pageIndex,
        pageSize: page.pageSize,
      })

      if (isApiSuccess(res)) {
        const rows = parseJsonData<GatewayInstance[]>(res, [])
        instanceList.value = Array.isArray(rows) ? rows : []
        instanceTotal.value = parsePageInfo(res).totalCount || instanceList.value.length
        if (instanceList.value.length > 0 && !selectedInstanceId.value) {
          handleInstanceChange(instanceList.value[0].gatewayInstanceId)
        }
      } else {
        message.error(res.errMsg || '获取网关实例列表失败')
      }
    } catch (error) {
      message.error('加载网关实例失败')
    } finally {
      loadingInstances.value = false
    }
  }

  // 处理实例选择变更
  function handleInstanceChange(instanceId: string) {
    selectedInstanceId.value = instanceId

    // 更新选中的实例信息
    selectedInstance.value =
      instanceList.value.find((item) => item.gatewayInstanceId === instanceId) || null
  }

  // 设置实例的代理类型
  function setInstanceProxyType(type: ProxyType | null) {
    instanceProxyType.value = type
  }

  return {
    // 状态
    loadingInstances,
    instanceList,
    instanceTotal,
    selectedInstanceId,
    selectedInstance,
    instanceDetailsVisible,
    instanceProxyType,
    gatewayInstanceId,
    hasSelectedInstance,
    instanceOptions,

    // 方法
    toggleInstanceDetails,
    loadGatewayInstances,
    handleInstanceChange,
    setInstanceProxyType,
  }
}
