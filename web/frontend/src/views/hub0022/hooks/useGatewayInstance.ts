import { useAppMessage } from '@/composables/useAppMessage'
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
      const res = await queryAllGatewayInstances({
        activeFlag: 'Y', // 默认只加载启用的实例
        pageIndex: 1,
        pageSize: 100,
      })

      if (res.oK) {
        // 解析bizData字段，这是一个JSON字符串
        try {
          const instanceData = JSON.parse(res.bizData || '[]')
          instanceList.value = instanceData || []

          // 如果有可用实例，默认选择第一个
          if (instanceList.value.length > 0 && !selectedInstanceId.value) {
            handleInstanceChange(instanceList.value[0].gatewayInstanceId)
          }
        } catch (parseError) {
          console.error('解析网关实例数据失败:', parseError)
          message.error('解析网关实例数据失败')
        }
      } else {
        message.error(res.errMsg || '获取网关实例列表失败')
      }
    } catch (error) {
      console.error('加载网关实例失败:', error)
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
