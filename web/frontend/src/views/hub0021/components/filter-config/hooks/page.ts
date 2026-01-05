/**
 * 过滤器配置列表页面级 Hook
 * - 组合 useFilterConfigService（纯业务逻辑）
 * - 处理新增对话框、工具栏、右键菜单等页面交互
 */

import { useGDialog } from '@/components/gdialog'
import { useMessage } from 'naive-ui'
import type { Ref } from 'vue'
import { ref } from 'vue'
import { useFilterConfigService } from './service'
import type { FilterConfig } from './types'

/**
 * 过滤器配置列表页面级 Hook
 * @param gridRef Grid 组件引用（可选）
 * @param gatewayInstanceId 网关实例ID（可选，用于全局过滤器）
 * @param routeConfigId 路由配置ID（可选，用于路由过滤器）
 * @param searchFormRef 搜索表单引用（可选）
 */
export function useFilterConfigPage(
  gridRef?: Ref<any> | any,
  gatewayInstanceId?: Ref<string | undefined> | string,
  routeConfigId?: Ref<string | undefined> | string,
  searchFormRef?: Ref<any> | any
) {
  const message = useMessage()
  const gDialog = useGDialog()

  // 业务服务（包含 model、增删改查等）
  const service = useFilterConfigService(gatewayInstanceId, routeConfigId, searchFormRef)

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
   * 将动态字段组装成 JSON 格式的 filterConfig
   */
  const buildFilterConfig = (formData: Record<string, any>): string => {
    const config: any = {}
    const filterType = formData.filterType

    // 根据 filterType 组装配置
    switch (filterType) {
      case 'header':
        config.headerConfig = {
          modifierType: formData['config.headerConfig.modifierType'],
          headerName: formData['config.headerConfig.headerName'],
          headerValue: formData['config.headerConfig.headerValue'],
          targetHeaderName: formData['config.headerConfig.targetHeaderName'],
          isRequestHeader: formData['config.headerConfig.isRequestHeader'] ?? true,
        }
        break
      case 'query-param':
        config.queryParamConfig = {
          modifierType: formData['config.queryParamConfig.modifierType'],
          paramName: formData['config.queryParamConfig.paramName'],
          paramValue: formData['config.queryParamConfig.paramValue'],
          targetParamName: formData['config.queryParamConfig.targetParamName'],
        }
        break
      case 'strip':
        config.stripConfig = {
          prefix: formData['config.stripConfig.prefix'],
        }
        break
      case 'rewrite':
        config.rewriteConfig = {
          mode: formData['config.rewriteConfig.mode'],
          from: formData['config.rewriteConfig.from'],
          to: formData['config.rewriteConfig.to'],
        }
        break
      case 'method':
        config.methodConfig = {
          mode: formData['config.methodConfig.mode'],
          allowedMethods: formData['config.methodConfig.allowedMethods'],
          deniedMethods: formData['config.methodConfig.deniedMethods'],
          rejectStatusCode: formData['config.methodConfig.rejectStatusCode'] ?? 405,
          rejectMessage: formData['config.methodConfig.rejectMessage'] ?? 'Method Not Allowed',
          caseSensitive: formData['config.methodConfig.caseSensitive'] ?? false,
        }
        break
      case 'body':
        config.bodyConfig = {
          modifierType: formData['config.bodyConfig.modifierType'],
          operation: formData['config.bodyConfig.operation'],
          allowedContentTypes: formData['config.bodyConfig.allowedContentTypes'],
          maxBodySize: formData['config.bodyConfig.maxBodySize'],
          filterConfig: formData['config.bodyConfig.filterConfigJson']
            ? JSON.parse(formData['config.bodyConfig.filterConfigJson'])
            : {},
        }
        break
      case 'cookie':
        config.cookieConfig = {
          operation: formData['config.cookieConfig.operation'],
          cookieName: formData['config.cookieConfig.cookieName'],
          cookieValue: formData['config.cookieConfig.cookieValue'],
          applyToResponse: formData['config.cookieConfig.applyToResponse'] ?? false,
          cookieAttributes: {
            domain: formData['config.cookieConfig.cookieAttributes.domain'],
            path: formData['config.cookieConfig.cookieAttributes.path'],
            maxAge: formData['config.cookieConfig.cookieAttributes.maxAge'],
            secure: formData['config.cookieConfig.cookieAttributes.secure'] ?? false,
            httpOnly: formData['config.cookieConfig.cookieAttributes.httpOnly'] ?? false,
            sameSite: formData['config.cookieConfig.cookieAttributes.sameSite'],
          },
        }
        break
      case 'response':
        config.responseConfig = {
          operation: formData['config.responseConfig.operation'],
          setInRequestPhase: formData['config.responseConfig.setInRequestPhase'] ?? false,
          filterConfig: formData['config.responseConfig.filterConfigJson']
            ? JSON.parse(formData['config.responseConfig.filterConfigJson'])
            : {},
          conditions: formData['config.responseConfig.conditionsJson']
            ? JSON.parse(formData['config.responseConfig.conditionsJson'])
            : {},
        }
        break
    }

    return JSON.stringify(config)
  }

  /**
   * 将 JSON 格式的 filterConfig 解析为动态字段
   */
  const parseFilterConfig = (filterConfig: string | undefined, filterType: string): Record<string, any> => {
    const result: Record<string, any> = {}
    
    if (!filterConfig) {
      return result
    }

    try {
      const config = JSON.parse(filterConfig)
      
      switch (filterType) {
        case 'header':
          if (config.headerConfig) {
            result['config.headerConfig.modifierType'] = config.headerConfig.modifierType
            result['config.headerConfig.headerName'] = config.headerConfig.headerName
            result['config.headerConfig.headerValue'] = config.headerConfig.headerValue
            result['config.headerConfig.targetHeaderName'] = config.headerConfig.targetHeaderName
            result['config.headerConfig.isRequestHeader'] = config.headerConfig.isRequestHeader ?? true
          }
          break
        case 'query-param':
          if (config.queryParamConfig) {
            result['config.queryParamConfig.modifierType'] = config.queryParamConfig.modifierType
            result['config.queryParamConfig.paramName'] = config.queryParamConfig.paramName
            result['config.queryParamConfig.paramValue'] = config.queryParamConfig.paramValue
            result['config.queryParamConfig.targetParamName'] = config.queryParamConfig.targetParamName
          }
          break
        case 'strip':
          if (config.stripConfig) {
            result['config.stripConfig.prefix'] = config.stripConfig.prefix
          }
          break
        case 'rewrite':
          if (config.rewriteConfig) {
            result['config.rewriteConfig.mode'] = config.rewriteConfig.mode
            result['config.rewriteConfig.from'] = config.rewriteConfig.from
            result['config.rewriteConfig.to'] = config.rewriteConfig.to
          }
          break
        case 'method':
          if (config.methodConfig) {
            result['config.methodConfig.mode'] = config.methodConfig.mode
            result['config.methodConfig.allowedMethods'] = config.methodConfig.allowedMethods
            result['config.methodConfig.deniedMethods'] = config.methodConfig.deniedMethods
            result['config.methodConfig.rejectStatusCode'] = config.methodConfig.rejectStatusCode ?? 405
            result['config.methodConfig.rejectMessage'] = config.methodConfig.rejectMessage ?? 'Method Not Allowed'
            result['config.methodConfig.caseSensitive'] = config.methodConfig.caseSensitive ?? false
          }
          break
        case 'body':
          if (config.bodyConfig) {
            result['config.bodyConfig.modifierType'] = config.bodyConfig.modifierType
            result['config.bodyConfig.operation'] = config.bodyConfig.operation
            result['config.bodyConfig.allowedContentTypes'] = config.bodyConfig.allowedContentTypes
            result['config.bodyConfig.maxBodySize'] = config.bodyConfig.maxBodySize
            result['config.bodyConfig.filterConfigJson'] = JSON.stringify(config.bodyConfig.filterConfig || {}, null, 2)
          }
          break
        case 'cookie':
          if (config.cookieConfig) {
            result['config.cookieConfig.operation'] = config.cookieConfig.operation
            result['config.cookieConfig.cookieName'] = config.cookieConfig.cookieName
            result['config.cookieConfig.cookieValue'] = config.cookieConfig.cookieValue
            result['config.cookieConfig.applyToResponse'] = config.cookieConfig.applyToResponse ?? false
            if (config.cookieConfig.cookieAttributes) {
              result['config.cookieConfig.cookieAttributes.domain'] = config.cookieConfig.cookieAttributes.domain
              result['config.cookieConfig.cookieAttributes.path'] = config.cookieConfig.cookieAttributes.path
              result['config.cookieConfig.cookieAttributes.maxAge'] = config.cookieConfig.cookieAttributes.maxAge
              result['config.cookieConfig.cookieAttributes.secure'] = config.cookieConfig.cookieAttributes.secure ?? false
              result['config.cookieConfig.cookieAttributes.httpOnly'] = config.cookieConfig.cookieAttributes.httpOnly ?? false
              result['config.cookieConfig.cookieAttributes.sameSite'] = config.cookieConfig.cookieAttributes.sameSite
            }
          }
          break
        case 'response':
          if (config.responseConfig) {
            result['config.responseConfig.operation'] = config.responseConfig.operation
            result['config.responseConfig.setInRequestPhase'] = config.responseConfig.setInRequestPhase ?? false
            result['config.responseConfig.filterConfigJson'] = JSON.stringify(config.responseConfig.filterConfig || {}, null, 2)
            result['config.responseConfig.conditionsJson'] = JSON.stringify(config.responseConfig.conditions || {}, null, 2)
          }
          break
      }
    } catch (error) {
      console.error('解析过滤器配置失败:', error)
    }

    return result
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
  const handleMenuClick = async ({ menu, row }: { menu: any; row: FilterConfig }) => {
    switch (menu.code) {
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
        console.warn('未知的菜单项:', menu.code)
    }
  }

  // ============= 单个操作处理 =============

  /**
   * 处理删除
   */
  const handleDelete = async (filter: FilterConfig) => {
    const confirmed = await gDialog.warning({
      title: '确认删除',
      content: `确定要删除过滤器"${filter.filterName}"吗？此操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
    })

    if (!confirmed) {
      return
    }

    // deleteFilter 内部已经会调用 loadFilterList，这里不需要重复调用
    await service.deleteFilter(filter.filterConfigId)
  }

  /**
   * 处理批量删除
   */
  const handleBatchDelete = async () => {
    if (!gridRef?.value?.getCheckboxRecords) {
      message.warning('无法获取选中的记录')
      return
    }

    const selectedRows = gridRef.value.getCheckboxRecords() as FilterConfig[]
    if (!selectedRows || selectedRows.length === 0) {
      message.warning('请先选择要删除的过滤器')
      return
    }

    const confirmed = await gDialog.warning({
      title: '确认批量删除',
      content: `确定要删除选中的 ${selectedRows.length} 个过滤器吗？此操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
    })

    if (!confirmed) {
      return
    }

    const filterConfigIds = selectedRows.map((row) => row.filterConfigId)
    // deleteFilters 内部已经会调用 loadFilterList，这里不需要重复调用
    await service.deleteFilters(filterConfigIds)
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

