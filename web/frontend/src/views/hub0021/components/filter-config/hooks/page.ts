/**
 * 过滤器配置列表页面级 Hook
 * - 组合 useFilterConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useAppMessage } from '@/composables/useAppMessage'
import { rsConfirm } from '@/ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import { useFilterConfigService } from './service'
import type { FilterConfig } from './types'

/**
 * 过滤器配置列表页面级 Hook
 * @param moduleId 模块ID（用于权限控制，必填）
 * @param gridRef Grid 组件引用（可选）
 * @param gatewayInstanceId 网关实例ID（可选，用于全局过滤器）
 * @param routeConfigId 路由配置ID（可选，用于路由过滤器）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useFilterConfigPage(
  moduleId: Ref<string>,
  gridRef?: Ref<any> | any,
  gatewayInstanceId?: Ref<string | undefined> | string,
  routeConfigId?: Ref<string | undefined> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useAppMessage()
// 业务服务（包含 model、增删改查等，传递模块ID）
  const service = useFilterConfigService(moduleId.value, gatewayInstanceId, routeConfigId, searchFormRef)

  // 表单对话框状态（新增/编辑/查看共用）
  const formDialogVisible = ref(false)
  const formDialogMode = ref<'create' | 'edit' | 'view'>('create')
  const currentEditFilter = ref<FilterConfig | null>(null)

  // ============= 统一校验 =============

  /**
   * 校验是否已选择网关实例或路由
   */
  const validateContext = (showMessage = true): boolean => {
    const instanceId = typeof gatewayInstanceId === 'string' ? gatewayInstanceId : gatewayInstanceId?.value
    const routeId = typeof routeConfigId === 'string' ? routeConfigId : routeConfigId?.value

    if (!instanceId && !routeId) {
      if (showMessage) {
        message.warning('请先选择网关实例或路由')
      }
      return false
    }
    return true
  }

  // ============= 搜索和分页 =============

  /**
   * 处理搜索
   */
  const handleSearch = async (searchParams?: Record<string, any>) => {
    service.model.resetPagination()
    await service.loadFilterList(searchParams)
  }

  /**
   * 处理分页变化
   */
  const handlePageChange = async ({ currentPage, pageSize }: { currentPage: number; pageSize: number }) => {
    service.model.updatePagination({ pageIndex: currentPage, pageSize })
    await service.loadFilterList()
  }

  // ============= 工具栏按钮处理 =============

  /**
   * 处理工具栏按钮点击
   */
  const handleToolbarClick = async (key: string) => {
    switch (key) {
      case 'add':
        openAddDialog()
        break
      case 'delete':
        await handleBatchDelete()
        break
      default:
        console.warn('未知的工具栏按钮:', key)
    }
  }

  // ============= 对话框处理 =============

  /**
   * 打开新增对话框
   */
  const openAddDialog = () => {
    if (!validateContext()) {
      return
    }
    formDialogMode.value = 'create'
    currentEditFilter.value = null
    formDialogVisible.value = true
  }

  /**
   * 打开编辑对话框
   */
  const openEditDialog = async (filter: FilterConfig) => {
    if (!filter.filterConfigId) {
      message.warning('过滤器配置ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestFilter = await service.getFilterDetail(filter.filterConfigId)
      if (latestFilter) {
        formDialogMode.value = 'edit'
        // 解析 filterConfig 为动态字段
        const parsedConfig = parseFilterConfig(latestFilter.filterConfig, latestFilter.filterType)
        currentEditFilter.value = {
          ...latestFilter,
          ...parsedConfig,
        }
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'edit'
        const parsedConfig = parseFilterConfig(filter.filterConfig, filter.filterType)
        currentEditFilter.value = {
          ...filter,
          ...parsedConfig,
        }
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'edit'
      const parsedConfig = parseFilterConfig(filter.filterConfig, filter.filterType)
      currentEditFilter.value = {
        ...filter,
        ...parsedConfig,
      }
      formDialogVisible.value = true
    }
  }

  /**
   * 打开查看对话框
   */
  const openViewDialog = async (filter: FilterConfig) => {
    if (!filter.filterConfigId) {
      message.warning('过滤器配置ID不能为空')
      return
    }

    try {
      // 获取最新数据
      const latestFilter = await service.getFilterDetail(filter.filterConfigId)
      if (latestFilter) {
        formDialogMode.value = 'view'
        // 解析 filterConfig 为动态字段
        const parsedConfig = parseFilterConfig(latestFilter.filterConfig, latestFilter.filterType)
        currentEditFilter.value = {
          ...latestFilter,
          ...parsedConfig,
        }
        formDialogVisible.value = true
      } else {
        // 降级：使用传入的数据
        formDialogMode.value = 'view'
        const parsedConfig = parseFilterConfig(filter.filterConfig, filter.filterType)
        currentEditFilter.value = {
          ...filter,
          ...parsedConfig,
        }
        formDialogVisible.value = true
      }
    } catch (error) {
      // 降级：使用传入的数据
      formDialogMode.value = 'view'
      const parsedConfig = parseFilterConfig(filter.filterConfig, filter.filterType)
      currentEditFilter.value = {
        ...filter,
        ...parsedConfig,
      }
      formDialogVisible.value = true
    }
  }

  /**
   * 关闭表单对话框
   */
  const closeFormDialog = () => {
    formDialogVisible.value = false
    currentEditFilter.value = null
  }

  /**
   * 提交时取 form.config 子树，只序列化当前 filterType 对应的一支。
   */
  const buildFilterConfig = (formData: Record<string, any>): string => {
    const tree = formData.config ?? {}
    const filterType = formData.filterType
    const config: Record<string, any> = {}

    const parseJsonField = (raw: unknown) => {
      if (raw && typeof raw === 'object') return raw
      if (typeof raw === 'string' && raw.trim()) {
        try {
          return JSON.parse(raw)
        } catch {
          return {}
        }
      }
      return {}
    }

    switch (filterType) {
      case 'header':
        config.headerConfig = {
          isRequestHeader: true,
          ...tree.headerConfig,
        }
        break
      case 'query-param':
        config.queryParamConfig = tree.queryParamConfig ?? {}
        break
      case 'strip':
        config.stripConfig = tree.stripConfig ?? {}
        break
      case 'rewrite':
        config.rewriteConfig = tree.rewriteConfig ?? {}
        break
      case 'method':
        config.methodConfig = {
          rejectStatusCode: 405,
          rejectMessage: 'Method Not Allowed',
          caseSensitive: false,
          ...tree.methodConfig,
        }
        break
      case 'body': {
        const body = { ...tree.bodyConfig }
        body.filterConfig = parseJsonField(body.filterConfigJson ?? body.filterConfig)
        delete body.filterConfigJson
        config.bodyConfig = body
        break
      }
      case 'cookie':
        config.cookieConfig = {
          applyToResponse: false,
          ...tree.cookieConfig,
          cookieAttributes: {
            secure: false,
            httpOnly: false,
            ...tree.cookieConfig?.cookieAttributes,
          },
        }
        break
      case 'response': {
        const response = { ...tree.responseConfig }
        response.filterConfig = parseJsonField(response.filterConfigJson ?? response.filterConfig)
        response.conditions = parseJsonField(response.conditionsJson ?? response.conditions)
        delete response.filterConfigJson
        delete response.conditionsJson
        config.responseConfig = {
          setInRequestPhase: false,
          ...response,
        }
        break
      }
    }

    return JSON.stringify(config)
  }

  /**
   * 读接口：JSON 字符串 → form.config 对象树（不再摊成 config.xxx 扁平 key）。
   * textarea 用的 filterConfigJson / conditionsJson 是树上的展示投影。
   */
  const parseFilterConfig = (filterConfig: string | undefined, _filterType?: string): Record<string, any> => {
    if (!filterConfig) return {}

    try {
      const config = JSON.parse(filterConfig)
      if (!config || typeof config !== 'object' || Array.isArray(config)) return {}

      if (config.bodyConfig?.filterConfig && typeof config.bodyConfig.filterConfig !== 'string') {
        config.bodyConfig.filterConfigJson = JSON.stringify(config.bodyConfig.filterConfig, null, 2)
      }
      if (config.responseConfig) {
        if (config.responseConfig.filterConfig && typeof config.responseConfig.filterConfig !== 'string') {
          config.responseConfig.filterConfigJson = JSON.stringify(
            config.responseConfig.filterConfig,
            null,
            2,
          )
        }
        if (config.responseConfig.conditions && typeof config.responseConfig.conditions !== 'string') {
          config.responseConfig.conditionsJson = JSON.stringify(
            config.responseConfig.conditions,
            null,
            2,
          )
        }
      }

      return { config }
    } catch (error) {
      console.error('解析过滤器配置失败:', error)
      return {}
    }
  }

  /**
   * 处理表单提交
   */
  const handleFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return

    // 查看模式下不执行提交
    if (formDialogMode.value === 'view') {
      closeFormDialog()
      return
    }

    // 校验是否已选择实例或路由
    if (!validateContext()) {
      return false
    }

    try {
      // 组装 filterConfig
      const filterConfig = buildFilterConfig(formData)
      
      // 移除动态配置字段，只保留基础字段
      const submitData: Partial<FilterConfig> = {
        filterConfigId: formData.filterConfigId,
        gatewayInstanceId: formData.gatewayInstanceId || (typeof gatewayInstanceId === 'string' ? gatewayInstanceId : gatewayInstanceId?.value),
        routeConfigId: formData.routeConfigId || (typeof routeConfigId === 'string' ? routeConfigId : routeConfigId?.value),
        filterName: formData.filterName,
        filterType: formData.filterType,
        filterAction: formData.filterAction,
        filterOrder: formData.filterOrder,
        filterDesc: formData.filterDesc,
        activeFlag: formData.activeFlag,
        noteText: formData.noteText,
        filterConfig,
      }

      let success = false
      if (formDialogMode.value === 'create') {
        success = await service.addFilter(submitData)
      } else if (formDialogMode.value === 'edit' && currentEditFilter.value) {
        success = await service.editFilter(
          currentEditFilter.value.filterConfigId,
          submitData
        )
      }

      if (success) {
        closeFormDialog()
        // addFilter 和 editFilter 内部已经处理了列表更新，这里不需要重复刷新
      }
    } catch (error: any) {
      console.error('提交过滤器配置失败:', error)
      message.error(error.message || '提交失败，请重试')
    }
  }

  // ============= 右键菜单处理 =============

  /**
   * 处理右键菜单点击
   */
  const handleMenuClick = async ({ key, row }: { key: string; row?: FilterConfig }) => {
    if (!row) return
    switch (key) {
      case 'view':
        await openViewDialog(row)
        break
      case 'edit':
        await openEditDialog(row)
        break
      case 'toggle-status':
        await handleToggleStatus(row)
        break
      case 'delete':
        await handleDelete(row)
        break
      default:
        console.warn('未知的菜单项:', key)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (filter: FilterConfig) => {
    const confirmed = await rsConfirm.warning({
      title: '确认删除',
      description: `确定要删除过滤器"${filter.filterName}"吗？此操作不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    // deleteFilter 内部已经会调用 loadFilterList，这里不需要重复调用
    await service.deleteFilter(filter.filterConfigId)
  }

  /**
   * 处理批量删除
   * 使用 getActiveRows：优先勾选行，无勾选时回退当前高亮行
   */
  const handleBatchDelete = async () => {
    if (!gridRef?.value) {
      message.warning('无法获取表格引用')
      return
    }

    const selectedRows = (gridRef.value.getActiveRows?.() || []) as FilterConfig[]
    if (selectedRows.length === 0) {
      message.warning('请先选择要删除的过滤器')
      return
    }

    const confirmed = await rsConfirm.warning({
      title: selectedRows.length > 1 ? '确认批量删除' : '确认删除',
      description:
        selectedRows.length > 1
          ? `确定要删除选中的 ${selectedRows.length} 个过滤器吗？此操作不可恢复。`
          : `确定要删除过滤器"${selectedRows[0].filterName}"吗？此操作不可恢复。`,
      confirmText: '删除',
      cancelText: '取消',
    })

    if (!confirmed) {
      return
    }

    const filterConfigIds = selectedRows
      .map((row) => row.filterConfigId)
      .filter((id): id is string => Boolean(id))

    if (filterConfigIds.length === 0) {
      message.warning('过滤器配置ID无效')
      return
    }

    const success =
      filterConfigIds.length === 1
        ? await service.deleteFilter(filterConfigIds[0])
        : await service.deleteFilters(filterConfigIds)

    if (success) {
      gridRef.value.clearSelection?.()
    }
  }

  /**
   * 处理切换状态
   */
  const handleToggleStatus = async (filter: FilterConfig) => {
    // toggleFilterStatus 内部已经会调用 loadFilterList，这里不需要重复调用
    await service.toggleFilterStatus(filter)
  }

  return {
    // 服务
    service,

    // 对话框状态
    formDialogVisible,
    formDialogMode,
    currentEditFilter,

    // 方法
    handleSearch,
    handlePageChange,
    handleToolbarClick,
    handleMenuClick,
    openAddDialog,
    openEditDialog,
    openViewDialog,
    closeFormDialog,
    handleFormSubmit,
    handleDelete,
    handleBatchDelete,
    handleToggleStatus,
  }
}

