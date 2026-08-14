/**
 * 路由配置列表页面级 Hook
 * - 组合 useRouteConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { getApiMessage, isApiSuccess } from '@/utils/format'
import { useAppMessage } from '@/composables/useAppMessage'
import { takeNamedObject, type RsDataFormRenderContext } from '@/components/form/rs-data'
import { getByNamePath, setByNamePath } from '@/ui'
import type { Ref } from 'vue'
import { onMounted, ref } from 'vue'
import { getRouteConfig } from '../../../api'
import type { RouteConfig } from '../types'
import { MatchType } from '../types'
import { useRouteConfigService } from './service'

/**
 * 路由配置列表页面级 Hook
 */
export function useRouteConfigPage(
  gatewayInstanceId?: string,
  searchFormRef?: Ref<any> | any,
  gridRef?: Ref<any> | any
) {
  const message = useAppMessage()

  // 业务服务（包含 model、增删改查等）
  const serviceResult = useRouteConfigService(gatewayInstanceId, searchFormRef)

  // 对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditRoute = ref<RouteConfig | null>(null)

  // 路由配置对话框状态（所有配置对话框共用同一个路由配置ID）
  const currentRouteConfigId = ref<string>('')
  
  // 断言配置对话框状态
  const assertConfigDialogVisible = ref(false)

  // 路由级配置对话框状态
  const ipAccessControlDialogVisible = ref(false)
  const userAgentAccessControlDialogVisible = ref(false)
  const apiAccessControlDialogVisible = ref(false)
  const domainAccessControlDialogVisible = ref(false)
  const corsConfigDialogVisible = ref(false)
  const authConfigDialogVisible = ref(false)
  const rateLimitConfigDialogVisible = ref(false)
  const filterConfigDialogVisible = ref(false)

  // ============= 统一校验 =============

  /**
   * 校验是否已选择网关实例
   * @param showMessage 是否显示提示消息
   * @returns 是否已选择实例
   */
  const validateInstanceSelected = (showMessage = true): boolean => {
    if (!gatewayInstanceId) {
      if (showMessage) {
        message.warning('请先选择网关实例')
      }
      return false
    }
    return true
  }

  // ============= 对话框操作 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    if (!validateInstanceSelected()) {
      return
    }
    formDialogMode.value = 'create'
    formDialogVisible.value = true
    currentEditRoute.value = null
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (route: RouteConfig): Promise<RouteConfig | null> => {
    if (!validateInstanceSelected()) {
      return null
    }
    try {
      // 获取完整详情
      const response = await getRouteConfig(route.routeConfigId)
      
      if (isApiSuccess(response)) {
        const detailRoute = JSON.parse(response.bizData) as RouteConfig
        
        formDialogMode.value = 'edit'
        formDialogVisible.value = true
        currentEditRoute.value = detailRoute
        return detailRoute
      } else {
        message.error(getApiMessage(response, '获取路由配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取路由配置详情失败')
      return null
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (route: RouteConfig): Promise<RouteConfig | null> => {
    if (!validateInstanceSelected()) {
      return null
    }
    try {
      // 获取完整详情
      const response = await getRouteConfig(route.routeConfigId)
      
      if (isApiSuccess(response)) {
        const detailRoute = JSON.parse(response.bizData) as RouteConfig
        
        formDialogMode.value = 'view'
        formDialogVisible.value = true
        currentEditRoute.value = detailRoute
        return detailRoute
      } else {
        message.error(getApiMessage(response, '获取路由配置详情失败'))
        return null
      }
    } catch (error) {
      message.error('获取路由配置详情失败')
      return null
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditRoute.value = null
  }

  /**
   * 获取表单初始数据（处理 JSON 字段转换）
   */
  const getRouteFormInitialData = (): Record<string, any> | undefined => {
    if (!currentEditRoute.value) {
      return undefined
    }

    const route = currentEditRoute.value
    const formData: Record<string, any> = {
      ...route,
    }

    // 处理 allowedMethods（字符串转数组）
    if (typeof route.allowedMethods === 'string' && route.allowedMethods) {
      try {
        formData.allowedMethods = JSON.parse(route.allowedMethods)
      } catch {
        // 如果不是 JSON，尝试按逗号分割
        formData.allowedMethods = route.allowedMethods.split(',').map((m) => m.trim()).filter(Boolean)
      }
    } else if (Array.isArray(route.allowedMethods)) {
      formData.allowedMethods = route.allowedMethods
    } else {
      formData.allowedMethods = []
    }

    // 解析 routeMetadata 为嵌套对象（NamePath）
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

    // 解析 routeMetadata 为嵌套对象，结构化字段走 NamePath；其它键随对象保留
    const routeMetadataObj = parseJson(route.routeMetadata)
    if (routeMetadataObj && typeof routeMetadataObj === 'object' && !Array.isArray(routeMetadataObj)) {
      formData.routeMetadata = { ...routeMetadataObj }
    } else {
      formData.routeMetadata = {}
    }
    // 历史路由未配置或非 Y 时默认关闭；仅 "Y" 开启覆盖
    const overrideFlag = getByNamePath(formData, 'routeMetadata.overrideProxyTimeout')
    setByNamePath(formData, 'routeMetadata.overrideProxyTimeout', overrideFlag === 'Y' ? 'Y' : 'N')

    return formData
  }

  /**
   * 提交表单（新增/编辑共用，由 RsDataFormModal 收集表单数据后回调）
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      closeFormDialog()
      return
    }

    // 校验是否已选择实例
    if (!validateInstanceSelected()) {
      return false
    }

    // 验证路由路径（参考 useRouteForm.ts 的验证逻辑）
    if (formData.routePath) {
      // 基本格式验证：必须以 / 开头
      if (!formData.routePath.startsWith('/')) {
        message.warning('路由路径必须以 / 开头')
        return false
      }
      // 正则匹配时验证正则表达式有效性
      if (formData.matchType === MatchType.REGEX) {
        try {
          new RegExp(formData.routePath)
        } catch {
          message.warning('请输入有效的正则表达式')
          return false
        }
      }
    }

    try {
      // 准备提交数据
      const processedData: Partial<RouteConfig> = {
        ...formData,
        gatewayInstanceId: gatewayInstanceId || formData.gatewayInstanceId,
      }

      // 处理 routeMetadata：嵌套对象打包为 JSON 字符串
      const routeMetadataObj = takeNamedObject(formData, 'routeMetadata')

      const finalData: Partial<RouteConfig> = {
        ...processedData,
      }

      Object.keys(finalData).forEach((key) => {
        if (key.startsWith('routeMetadata.')) {
          delete finalData[key as keyof typeof finalData]
        }
      })

      // 处理 allowedMethods（如果是数组，转换为 JSON 字符串）
      if (Array.isArray(finalData.allowedMethods)) {
        finalData.allowedMethods = JSON.stringify(finalData.allowedMethods) as any
      }

      // 将 routeMetadata 转换为 JSON 字符串（后端存储为 JSON 字符串）
      finalData.routeMetadata = JSON.stringify(routeMetadataObj) as any

      let success = false
      if (formDialogMode.value === 'create') {
        success = await serviceResult.addRoute(finalData as any)
      } else if (formDialogMode.value === 'edit' && currentEditRoute.value) {
        success = await serviceResult.editRoute(
          currentEditRoute.value.routeConfigId,
          finalData as any
        )
      }

      if (success) {
        closeFormDialog()
        // 刷新列表
        serviceResult.loadRouteList()
      }

      return success
    } catch (error) {
      message.error('操作失败')
      return false
    }
  }

  /**
   * 处理删除
   */
  const handleDelete = async (route: RouteConfig) => {
    if (!validateInstanceSelected()) {
      return
    }
    await serviceResult.deleteRoute(route.routeConfigId)
  }

  /**
   * 处理批量删除
   * 使用 getActiveRows：优先勾选行，无勾选时回退当前高亮行
   */
  const handleBatchDelete = async () => {
    if (!validateInstanceSelected()) {
      return
    }
    if (!gridRef?.value) {
      message.warning('Grid 引用未设置')
      return
    }
    const selectedRecords = gridRef.value.getActiveRows?.() || []
    if (selectedRecords.length === 0) {
      message.warning('请选择或点击要删除的路由配置')
      return
    }
    const routeConfigIds = selectedRecords.map((record: RouteConfig) => record.routeConfigId)
    await serviceResult.batchDeleteRoutes(routeConfigIds)
  }

  /**
   * 处理搜索（合并 gatewayInstanceId）
   */
  const handleSearch = async (formData?: Record<string, any>) => {
    if (!validateInstanceSelected()) {
      return
    }
    // 合并 gatewayInstanceId 到查询参数中（validateInstanceSelected 已保证实例 ID 存在）
    const searchParams = {
      ...formData,
      gatewayInstanceId,
    }
    await serviceResult.handleSearch(searchParams)
  }

  /**
   * 工具栏按钮点击处理
   * @param key 按钮 key
   * @param formData 表单数据（可选，search 操作时会传递）
   */
  const handleToolbarClick = async (key: string, formData?: Record<string, any>) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'delete':
        await handleBatchDelete()
        break
      case 'search':
        await handleSearch(formData)
        break
      case 'reset':
        await serviceResult.handleReset()
        break
      default:
        break
    }
  }

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async (params: { key: string; row?: any; column?: any }) => {
    const { key, row } = params
    if (!row) {
      return
    }
    const route = row as RouteConfig
    
    switch (key) {
      case 'view':
        await openViewDialog(route)
        break
      case 'edit':
        await openEditDialog(route)
        break
      case 'assertConfig':
        // 打开路由断言配置对话框
        currentRouteConfigId.value = route.routeConfigId
        assertConfigDialogVisible.value = true
        break
      case 'filters':
        // 打开路由过滤器配置对话框
        currentRouteConfigId.value = route.routeConfigId
        filterConfigDialogVisible.value = true
        break
      case 'ipAccessControl':
        // 打开IP访问控制对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        ipAccessControlDialogVisible.value = true
        break
      case 'userAgentAccessControl':
        // 打开User-Agent访问控制对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        userAgentAccessControlDialogVisible.value = true
        break
      case 'apiAccessControl':
        // 打开API访问控制对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        apiAccessControlDialogVisible.value = true
        break
      case 'domainAccessControl':
        // 打开域名访问控制对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        domainAccessControlDialogVisible.value = true
        break
      case 'corsConfig':
        // 打开跨域配置对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        corsConfigDialogVisible.value = true
        break
      case 'authConfig':
        // 打开认证配置对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        authConfigDialogVisible.value = true
        break
      case 'rateLimitConfig':
        // 打开限流配置对话框（路由级配置）
        currentRouteConfigId.value = route.routeConfigId
        rateLimitConfigDialogVisible.value = true
        break
      case 'delete':
        await handleDelete(route)
        break
      default:
        break
    }
  }

  // ============= 初始化 =============

  // 组件挂载时加载数据
  onMounted(() => {
    if (gatewayInstanceId) {
      serviceResult.loadRouteList()
    }
  })

  // 更新 routeFormConfig，注入 gatewayInstanceId 到自定义字段的 context
  const updatedFields = serviceResult.model.routeFormConfig.fields.map((field: any) => {
    if (field.field === 'serviceDefinitionId' && field.render) {
      const originalRender = field.render
      return {
        ...field,
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
          return originalRender(formData, {
            ...ctx,
            gatewayInstanceId: gatewayInstanceId || '',
          })
        },
      }
    }
    return field
  })

  const routeFormConfigWithContext = {
    ...serviceResult.model.routeFormConfig,
    fields: updatedFields,
  }

  return {
    // Service 实例（包含所有业务逻辑和状态）
    service: serviceResult,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditRoute,

    // 当前路由配置ID（所有配置对话框共用）
    currentRouteConfigId,

    // 路由级配置对话框状态
    assertConfigDialogVisible,
    ipAccessControlDialogVisible,
    userAgentAccessControlDialogVisible,
    apiAccessControlDialogVisible,
    domainAccessControlDialogVisible,
    corsConfigDialogVisible,
    authConfigDialogVisible,
    rateLimitConfigDialogVisible,
    filterConfigDialogVisible,

    // 表单配置（包含注入的 context）
    routeFormConfig: routeFormConfigWithContext,

    // 方法
    handleFormSubmit,
    getRouteFormInitialData,
    handleToolbarClick,
    handleMenuClick,
    handleSearch,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
  }
}

/**
 * 路由配置列表页面类型
 */
export type RouteConfigPage = ReturnType<typeof useRouteConfigPage>

