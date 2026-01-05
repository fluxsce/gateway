/**
 * 限流配置页面级 Hook
 * - 组合 useRateLimitConfigService（纯业务逻辑）
 * - 处理表单对话框等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref, watch } from 'vue'
import type { RateLimitConfig } from './types'
import { useRateLimitConfigService } from './useRateLimitConfigService'

/**
 * 限流配置页面级 Hook
 * @param props 组件 props（包含 gatewayInstanceId、routeConfigId 的响应式 ref）
 */
export function useRateLimitConfigPage(props: {
  gatewayInstanceId?: Ref<string | undefined>
  routeConfigId?: Ref<string | undefined>
}) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useRateLimitConfigService()

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<RateLimitConfig | null>(null)

  /**
   * 将配置对象转换为表单数据格式（处理 JSON 字符串字段）
   */
  const convertToFormData = (config: RateLimitConfig): any => {
    // 安全解析 JSON 字符串
    const parseJson = (value: any): any => {
      if (typeof value === 'string' && value) {
        try {
          return JSON.parse(value)
        } catch {
          return value
        }
      }
      return value
    }

    return {
      ...config,
      // customConfig 在后端是 JSON 字符串，前端表单需要对象格式
      customConfig: parseJson(config.customConfig) || {},
    }
  }

  /**
   * 将表单数据转换为API数据格式（对象转换为后端需要的格式）
   */
  const convertToApiData = (formData: Record<string, any>): any => {
    return {
      ...formData,
      // customConfig 前端是对象，后端需要 JSON 字符串
      customConfig: typeof formData.customConfig === 'object'
        ? JSON.stringify(formData.customConfig)
        : formData.customConfig || '{}',
    }
  }

  /**
   * 加载配置详情（根据 props 中的 ID）
   */
  const loadConfig = async () => {
    const gatewayInstanceId = props.gatewayInstanceId?.value
    const routeConfigId = props.routeConfigId?.value

    if (!gatewayInstanceId && !routeConfigId) {
      return
    }

    const config = await service.getConfigDetail({
      gatewayInstanceId,
      routeConfigId,
    })

    if (config) {
      currentEditConfig.value = convertToFormData(config)
      formDialogMode.value = 'edit'
    } else {
      // 如果没有配置，则进入新增模式
      currentEditConfig.value = null
      formDialogMode.value = 'create'
    }
  }

  /** 打开表单对话框 */
  const openDialog = async () => {
    // 先加载配置数据，然后再打开对话框
    await loadConfig()
    // 数据加载完成后再打开对话框
    formDialogVisible.value = true
  }

  /** 关闭表单对话框 */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditConfig.value = null
  }

  /** 提交表单（新增/编辑共用，由 GDataFormModal 收集表单数据后回调） */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      return
    }

    // 将表单数据转换为API数据格式（JSON 转换等）
    const apiData = convertToApiData(formData)

    // 添加关联ID
    if (props.gatewayInstanceId?.value) {
      apiData.gatewayInstanceId = props.gatewayInstanceId.value
    }
    if (props.routeConfigId?.value) {
      apiData.routeConfigId = props.routeConfigId.value
    }

    if (formDialogMode.value === 'create') {
      // 新增模式
      const success = await service.addConfig(apiData)
      if (success) {
        // 保存成功后，重新加载配置数据（因为新增后应该进入编辑模式）
        await loadConfig()
      }
    } else if (formDialogMode.value === 'edit') {
      // 编辑模式
      if (!currentEditConfig.value) return
      
      // 合并当前配置ID和租户ID，确保更新的是正确的记录
      const updatedConfig = {
        ...apiData,
        rateLimitConfigId: currentEditConfig.value.rateLimitConfigId,
        tenantId: currentEditConfig.value.tenantId,
      } as Partial<RateLimitConfig> & {
        rateLimitConfigId: string
        tenantId: string
      }
      
      const success = await service.editConfig(updatedConfig)
      if (success) {
        // 保存成功后，重新加载配置数据（获取最新数据）
        await loadConfig()
      }
    }
  }

  // 监听 props 变化，重新加载配置
  watch(
    () => [props.gatewayInstanceId?.value, props.routeConfigId?.value],
    () => {
      if (formDialogVisible.value) {
        loadConfig()
      }
    }
  )

  return {
    // 业务服务（包含 model 与增删改查）
    service,

    // 表单对话框（新增/编辑/查看共用）
    formDialogVisible,
    formDialogMode,
    currentEditConfig,
    openDialog,
    closeFormDialog,
    handleFormSubmit,
    loadConfig,
  }
}

