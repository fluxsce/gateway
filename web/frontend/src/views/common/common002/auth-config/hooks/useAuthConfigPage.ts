/**
 * 认证配置页面级 Hook
 * - 组合 useAuthConfigService（纯业务逻辑）
 * - 处理表单对话框等页面交互
 */

import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref, watch } from 'vue'
import type { AuthConfig } from './types'
import { useAuthConfigService } from './useAuthConfigService'

/**
 * 认证配置页面级 Hook
 * @param moduleId 模块ID（用于权限控制，必填）
 * @param props 组件 props（包含 gatewayInstanceId、routeConfigId 的响应式 ref）
 */
export function useAuthConfigPage(
  moduleId: Ref<string>,
  props: {
    gatewayInstanceId?: Ref<string | undefined>
    routeConfigId?: Ref<string | undefined>
  }
) {
  const message = useMessage()

  // 业务服务（包含 model、增删改查等，传递 moduleId）
  const service = useAuthConfigService(moduleId.value)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditConfig = ref<AuthConfig | null>(null)

  /**
   * 将配置对象转换为表单数据格式（查询时：将 authConfig JSON 字符串解析为单独的 authConfig.xxx 字段）
   */
  const convertToFormData = (config: AuthConfig): any => {
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

    // 解析 authConfig JSON 字符串为对象
    const authConfigObj = parseJson(config.authConfig) || {}

    // 构建表单数据对象
    const formData: any = {
      ...config,
      // exemptPaths 在后端是 JSON 字符串，前端表单需要数组格式
      exemptPaths: parseJsonArray(config.exemptPaths),
      // exemptHeaders 在后端是 JSON 字符串，前端表单需要数组格式
      exemptHeaders: parseJsonArray(config.exemptHeaders),
    }

    // 将 authConfig 对象的嵌套属性展开为点号分隔的字段（用于表单字段绑定）
    // 这些字段名与 model.ts 中定义的字段名对应（如 authConfig.secret、authConfig.algorithm 等）
    if (authConfigObj && typeof authConfigObj === 'object') {
      Object.keys(authConfigObj).forEach((key) => {
        formData[`authConfig.${key}`] = authConfigObj[key]
      })
    }

    return formData
  }

  /**
   * 将表单数据转换为API数据格式（保存时：将 authConfig.xxx 字段合并为一个 authConfig JSON 字符串）
   */
  const convertToApiData = (formData: Record<string, any>): any => {
    // 从点号分隔字段重新构建 authConfig 对象
    const authConfigObj: Record<string, any> = {}
    
    // 遍历所有以 authConfig. 开头的字段
    Object.keys(formData).forEach((key) => {
      if (key.startsWith('authConfig.')) {
        const subKey = key.replace('authConfig.', '')
        authConfigObj[subKey] = formData[key]
      }
    })

    // 构建 API 数据对象，排除点号分隔字段
    const apiData: Record<string, any> = {}
    Object.keys(formData).forEach((key) => {
      // 排除以 authConfig. 开头的点号分隔字段（这些字段已经合并到 authConfig 对象中）
      if (!key.startsWith('authConfig.')) {
        apiData[key] = formData[key]
      }
    })

    return {
      ...apiData,
      // authConfig 前端是点号分隔字段，后端需要 JSON 字符串
      authConfig: Object.keys(authConfigObj).length > 0
        ? JSON.stringify(authConfigObj)
        : '{}',
      // exemptPaths 前端是数组，后端需要 JSON 字符串
      exemptPaths: Array.isArray(formData.exemptPaths)
        ? JSON.stringify(formData.exemptPaths)
        : formData.exemptPaths || '[]',
      // exemptHeaders 前端是数组，后端需要 JSON 字符串
      exemptHeaders: Array.isArray(formData.exemptHeaders)
        ? JSON.stringify(formData.exemptHeaders)
        : formData.exemptHeaders || '[]',
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

    // 将表单数据转换为API数据格式（合并点号分隔字段，转换为 JSON 字符串）
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
      // 注意：apiData 已经在组件层面完成了数据转换（点号分隔字段合并、JSON 转换等）
      const updatedConfig = {
        ...apiData,
        authConfigId: currentEditConfig.value.authConfigId,
        tenantId: currentEditConfig.value.tenantId,
      } as Partial<AuthConfig> & {
        authConfigId: string
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

