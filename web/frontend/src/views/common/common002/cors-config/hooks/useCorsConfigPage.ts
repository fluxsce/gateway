/**
 * CORS配置页面级 Hook
 * - 组合 useCorsConfigService（纯业务逻辑）
 * - 处理表单对话框等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref, watch } from 'vue'
import type { CorsConfig } from './types'
import { useCorsConfigService } from './useCorsConfigService'

/**
 * CORS配置页面级 Hook
 * @param props 组件 props（包含 gatewayInstanceId、routeConfigId 的响应式 ref）
 */
export function useCorsConfigPage(props: {
  gatewayInstanceId?: Ref<string | undefined>
  routeConfigId?: Ref<string | undefined>
}) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等）
  const service = useCorsConfigService()

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<CorsConfig | null>(null)

  /**
   * 将配置对象转换为表单数据格式（处理数组字段）
   */
  const convertToFormData = (config: CorsConfig): any => {
    // 安全解析 JSON 字符串为数组
    const parseJsonArray = (value: any): string[] => {
      if (Array.isArray(value)) {
        return value
      }
      if (typeof value === 'string' && value) {
        try {
          return JSON.parse(value)
        } catch {
          return []
        }
      }
      return []
    }

    return {
      ...config,
      // allowOrigins 在后端是 JSON 字符串，前端表单需要数组格式
      allowOrigins: parseJsonArray(config.allowOrigins),
      // allowMethods 在后端是逗号分隔的字符串，前端表单需要数组格式
      allowMethods: typeof config.allowMethods === 'string' 
        ? config.allowMethods.split(',').map(m => m.trim()).filter(m => m)
        : (Array.isArray(config.allowMethods) ? config.allowMethods : []),
      // allowHeaders 在后端是 JSON 字符串，前端表单需要数组格式
      allowHeaders: parseJsonArray(config.allowHeaders),
      // exposeHeaders 在后端是 JSON 字符串，前端表单需要数组格式
      exposeHeaders: parseJsonArray(config.exposeHeaders),
    }
  }

  /**
   * 将表单数据转换为API数据格式（数组转换为后端需要的格式）
   */
  const convertToApiData = (formData: Record<string, any>): any => {
    return {
      ...formData,
      // allowOrigins 前端是数组，后端需要 JSON 字符串
      allowOrigins: Array.isArray(formData.allowOrigins) 
        ? JSON.stringify(formData.allowOrigins) 
        : formData.allowOrigins,
      // allowMethods 前端是数组，后端需要逗号分隔的字符串
      allowMethods: Array.isArray(formData.allowMethods) 
        ? formData.allowMethods.join(',') 
        : formData.allowMethods,
      // allowHeaders 前端是数组，后端需要 JSON 字符串
      allowHeaders: Array.isArray(formData.allowHeaders) 
        ? JSON.stringify(formData.allowHeaders) 
        : formData.allowHeaders,
      // exposeHeaders 前端是数组，后端需要 JSON 字符串
      exposeHeaders: Array.isArray(formData.exposeHeaders) 
        ? JSON.stringify(formData.exposeHeaders) 
        : formData.exposeHeaders,
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

    // 将表单数据转换为API数据格式（数组转换为后端需要的格式）
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
        ...convertToApiData(currentEditConfig.value),
        ...apiData,
        corsConfigId: currentEditConfig.value.corsConfigId,
        tenantId: currentEditConfig.value.tenantId,
      } as Partial<CorsConfig> & {
        corsConfigId: string
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

