import { ref, reactive } from 'vue'
import { useAppMessage } from '@/composables/useAppMessage'
import * as routeApi from '../api'
import type { RouterConfig, RouterConfigForm, RouterQueryParams } from '../types'
import type { RouterConfig as ApiRouterConfig } from '../components/instance-tree/types'

/**
 * 将表单元数据对象转为接口所需的 JSON 字符串。
 */
function toJsonField(value: unknown): string {
  if (typeof value === 'string') return value
  return JSON.stringify(value ?? {})
}

/**
 * 组装提交给 add/editRouterConfig 的载荷。
 */
function toRouterConfigPayload(
  data: Partial<RouterConfigForm> & { tenantId: string; gatewayInstanceId: string },
): Partial<ApiRouterConfig> & { gatewayInstanceId: string } {
  return {
    ...data,
    routerMetadata: toJsonField(data.routerMetadata),
    customConfig: toJsonField(data.customConfig),
  }
}

export function useRouterConfigManager() {
  const message = useAppMessage()

  // 状态管理
  const loading = ref(false)
  const submitting = ref(false)
  const routerConfigs = ref<RouterConfig[]>([])
  const total = ref(0)
  const pageIndex = ref(1)
  const pageSize = ref(20)

  // 查询参数
  const queryParams = reactive<RouterQueryParams>({
    tenantId: 'default',
    gatewayInstanceId: '',
    pageIndex: 1,
    pageSize: 20,
  })

  /**
   * 获取Router配置列表
   */
  const getRouterConfigs = async (gatewayInstanceId: string, tenantId: string) => {
    try {
      loading.value = true
      queryParams.gatewayInstanceId = gatewayInstanceId
      queryParams.tenantId = tenantId
      queryParams.pageIndex = pageIndex.value
      queryParams.pageSize = pageSize.value

      const response = await routeApi.queryRouterConfigs(queryParams)
      if (response.oK && response.bizData) {
        const data = JSON.parse(response.bizData)
        routerConfigs.value = data.records || []
        total.value = data.total || 0
      } else {
        routerConfigs.value = []
        total.value = 0
      }
    } catch (error) {
      console.error('获取Router配置列表失败:', error)
      message.error('获取Router配置列表失败')
    } finally {
      loading.value = false
    }
  }

  /**
   * 创建Router配置
   */
  const createRouterConfig = async (
    data: RouterConfigForm & { tenantId: string; gatewayInstanceId: string },
  ) => {
    try {
      submitting.value = true
      await routeApi.addRouterConfig(toRouterConfigPayload(data))
      message.success('创建成功')
      return true
    } catch (error) {
      console.error('创建失败:', error)
      message.error('创建失败')
      return false
    } finally {
      submitting.value = false
    }
  }

  /**
   * 更新Router配置
   */
  const updateRouterConfig = async (
    routerConfigId: string,
    data: Partial<RouterConfigForm> & { tenantId: string; gatewayInstanceId: string },
  ) => {
    try {
      submitting.value = true
      await routeApi.editRouterConfig({
        ...toRouterConfigPayload(data),
        routerConfigId,
      })
      message.success('更新成功')
      return true
    } catch (error) {
      console.error('更新失败:', error)
      message.error('更新失败')
      return false
    } finally {
      submitting.value = false
    }
  }

  /**
   * 删除Router配置
   */
  const deleteRouterConfig = async (routerConfigId: string) => {
    try {
      submitting.value = true
      await routeApi.deleteRouterConfig(routerConfigId)
      message.success('删除成功')
      return true
    } catch (error) {
      console.error('删除失败:', error)
      message.error('删除失败')
      return false
    } finally {
      submitting.value = false
    }
  }

  return {
    // 状态
    routerConfigs,
    loading,
    submitting,
    total,
    pageIndex,
    pageSize,

    // 方法
    getRouterConfigs,
    createRouterConfig,
    updateRouterConfig,
    deleteRouterConfig,
  }
}
