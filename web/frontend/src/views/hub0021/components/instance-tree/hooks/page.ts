/**
 * 网关实例树组件 Page Hook
 * 处理页面交互、事件处理和渲染函数
 */

import { getApiMessage, isApiSuccess, parseJsonData } from '@/utils/format'
import { ServerOutline } from '@vicons/ionicons5'
import type { TreeOption } from 'naive-ui'
import { NIcon, NTag, useMessage } from 'naive-ui'
import { h, onBeforeUnmount, ref, watch } from 'vue'
import { addRouterConfig, editRouterConfig, getRouterConfigsByInstance } from '../../../api'
import type { GatewayInstance, InstanceTreeOption, RouterConfig } from '../types'
import { useGatewayInstanceTreeModel } from './model'
import { useGatewayInstanceTreeService } from './service'

/**
 * 网关实例树 Page Hook
 */
export function useGatewayInstanceTreePage() {
  const message = useMessage()

  // 初始化 Model
  const model = useGatewayInstanceTreeModel()

  // 初始化 Service
  const service = useGatewayInstanceTreeService(model)

  // ============= Router配置对话框状态 =============

  /** Router配置对话框显示状态 */
  const routerFormDialogVisible = ref(false)
  /** Router配置对话框模式 */
  const routerFormDialogMode = ref<'create' | 'edit' | 'view'>('create')
  /** 当前编辑的Router配置 */
  const currentEditRouter = ref<RouterConfig | null>(null)
  /** 提交状态 */
  const routerSubmitting = ref(false)
  /** 当前选中的网关实例ID（用于Router配置和全局过滤器配置） */
  const currentGatewayInstanceId = ref<string>('')

  // ============= 全局过滤器配置对话框状态 =============

  /** 全局过滤器配置对话框显示状态 */
  const globalFilterConfigDialogVisible = ref(false)

  // ============= 监听器 =============

  // 监听过滤关键词变化，重置到第一页并重新加载数据
  const stopFilterKeywordWatch = watch(model.filterKeyword, () => {
    model.resetPage()
    service.loadGatewayInstances()
  })

  // 组件卸载时清理监听器
  onBeforeUnmount(() => {
    stopFilterKeywordWatch()
  })

  // ============= 渲染函数 =============

  /**
   * 渲染节点前缀图标
   */
  function renderNodePrefix({ option }: { option: TreeOption }) {
    const instanceOption = option as InstanceTreeOption
    if (instanceOption.instance) {
      return h(NIcon, {
        size: 16,
        color: 'var(--g-primary)',
        style: { marginRight: '6px', flexShrink: 0 }
      }, {
        default: () => h(ServerOutline)
      })
    }
    return null
  }

  /**
   * 渲染节点标签（省略显示）
   */
  function renderNodeLabel({ option }: { option: TreeOption }) {
    return h('span', {
      class: 'tree-node-label',
      style: {
        display: 'inline-block',
        padding: '1px 4px',
        borderRadius: '4px',
        transition: 'background-color 0.2s'
      }
    }, option.label as string)
  }

  /**
   * 渲染节点后缀（健康状态标签）
   */
  function renderNodeSuffix({ option }: { option: TreeOption }) {
    const instanceOption = option as InstanceTreeOption
    if (instanceOption.instance) {
      return h(NTag, {
        type: instanceOption.instance.healthStatus === 'Y' ? 'success' : 'warning',
        size: 'small',
        style: { marginLeft: '8px', flexShrink: 0 }
      }, {
        default: () => instanceOption.instance!.healthStatus === 'Y' ? '健康' : '异常'
      })
    }
    return null
  }

  // ============= 事件处理 =============

  /**
   * 处理树节点选择
   * @param emit 组件 emit 函数
   */
  function handleTreeSelect(keys: string[], option: TreeOption, emit: (e: 'select', instanceId: string, instance: GatewayInstance) => void) {
    const instanceOption = option as InstanceTreeOption
    if (instanceOption.instance) {
      emit('select', instanceOption.instance.gatewayInstanceId, instanceOption.instance)
    }
  }

  /**
   * 处理分页变化（后端分页，需要重新加载数据）
   */
  function handlePageChange({ currentPage: page, pageSize: size }: { currentPage: number; pageSize: number }) {
    model.setCurrentPage(page)
    model.setPageSize(size)
    // 后端分页，需要重新加载数据
    service.loadGatewayInstances()
  }

  /**
   * 处理刷新
   */
  function handleRefresh() {
    service.loadGatewayInstances()
  }

  // ============= Router配置对话框操作 =============

  /**
   * 加载Router配置（根据 gatewayInstanceId 查询，返回单条数据）
   * 参考代理配置逻辑：如果配置存在则进入编辑模式，不存在则进入新增模式
   */
  const loadRouterConfig = async () => {
    const gatewayInstanceId = currentGatewayInstanceId.value

    if (!gatewayInstanceId) {
      return
    }

    try {
      const response = await getRouterConfigsByInstance(gatewayInstanceId)

      if (isApiSuccess(response)) {
        // 后端现在返回单个 RouterConfig 对象或 null
        const routerConfig = parseJsonData<RouterConfig | null>(response, null)
        
        if (routerConfig) {
          // 如果找到配置，进入编辑模式
          currentEditRouter.value = routerConfig
          routerFormDialogMode.value = 'edit'
        } else {
          // 如果没有配置，进入新增模式
          currentEditRouter.value = null
          routerFormDialogMode.value = 'create'
        }
      } else {
        // API 调用失败，默认进入新增模式
        currentEditRouter.value = null
        routerFormDialogMode.value = 'create'
      }
    } catch (error) {
      // 异常情况，默认进入新增模式
      currentEditRouter.value = null
      routerFormDialogMode.value = 'create'
    }
  }

  /**
   * 打开Router配置对话框
   */
  const openRouterDialog = async (instance: GatewayInstance) => {
    currentGatewayInstanceId.value = instance.gatewayInstanceId
    // 先加载配置数据，然后再打开对话框
    await loadRouterConfig()
    // 数据加载完成后再打开对话框
    routerFormDialogVisible.value = true
  }

  /**
   * 将Router配置转换为表单数据格式（解析 JSON 字符串字段）
   */
  const convertRouterToFormData = (router: RouterConfig): any => {
    // 安全解析 JSON 字符串
    const parseJson = (value: any): any => {
      if (typeof value === 'string' && value) {
        try {
          return JSON.parse(value)
        } catch {
          return {}
        }
      }
      return value || {}
    }

    return {
      ...router,
      routerMetadata: parseJson(router.routerMetadata),
      customConfig: parseJson(router.customConfig),
    }
  }

  /**
   * 获取Router配置表单初始数据
   */
  const getRouterFormInitialData = (): any => {
    if (routerFormDialogMode.value === 'create') {
      // 新增模式：使用默认值
      return {
        gatewayInstanceId: currentGatewayInstanceId.value,
        routerName: '默认Router',
        routerDesc: '',
        defaultPriority: 100,
        enableRouteCache: 'Y',
        routeCacheTtlSeconds: 300,
        maxRoutes: 1000,
        routeMatchTimeout: 5000,
        enableStrictMode: 'N',
        enableMetrics: 'Y',
        enableTracing: 'N',
        caseSensitive: 'N',
        removeTrailingSlash: 'Y',
        enableGlobalFilters: 'Y',
        filterExecutionMode: 'SEQUENTIAL',
        maxFilterChainDepth: 10,
        enableRoutePooling: 'N',
        routePoolSize: 100,
        enableAsyncProcessing: 'N',
        enableFallback: 'N',
        fallbackRoute: '',
        notFoundStatusCode: 404,
        notFoundMessage: 'Route not found',
        activeFlag: 'Y',
        noteText: '',
        routerMetadata: {},
        customConfig: {},
      }
    } else if (currentEditRouter.value) {
      // 编辑/查看模式：使用转换后的数据
      return convertRouterToFormData(currentEditRouter.value)
    }
    return undefined
  }

  /**
   * 提交Router配置表单
   */
  const handleRouterFormSubmit = async (formData?: Record<string, any>) => {
    if (!formData) return false

    // 查看模式下不执行提交
    if (routerFormDialogMode.value === 'view') {
      return false
    }

    try {
      routerSubmitting.value = true

      // 准备提交数据，将对象字段转换为 JSON 字符串
      const submitData: any = {
        ...formData,
        gatewayInstanceId: currentGatewayInstanceId.value || formData.gatewayInstanceId,
      }

      // 将 routerMetadata 和 customConfig 转换为 JSON 字符串
      if (formData.routerMetadata) {
        submitData.routerMetadata = typeof formData.routerMetadata === 'string'
          ? formData.routerMetadata
          : JSON.stringify(formData.routerMetadata)
      } else {
        submitData.routerMetadata = '{}'
      }

      if (formData.customConfig) {
        submitData.customConfig = typeof formData.customConfig === 'string'
          ? formData.customConfig
          : JSON.stringify(formData.customConfig)
      } else {
        submitData.customConfig = '{}'
      }

      let success = false
      if (routerFormDialogMode.value === 'create') {
        const response = await addRouterConfig(submitData)
        if (isApiSuccess(response)) {
          message.success(getApiMessage(response, '创建Router配置成功'))
          success = true
        } else {
          message.error(getApiMessage(response, '创建Router配置失败'))
        }
      } else if (routerFormDialogMode.value === 'edit' && currentEditRouter.value) {
        const routerConfigId = currentEditRouter.value.routerConfigId
        if (!routerConfigId) {
          message.error('Router配置ID不存在')
          return false
        }
        const response = await editRouterConfig({
          ...submitData,
          routerConfigId,
        })
        if (isApiSuccess(response)) {
          message.success(getApiMessage(response, '更新Router配置成功'))
          success = true
        } else {
          message.error(getApiMessage(response, '更新Router配置失败'))
        }
      }

      if (success) {
        // 保存成功后，重新加载配置数据（因为新增后应该进入编辑模式）
        await loadRouterConfig()
        // 保持对话框打开，让用户继续编辑
      }

      return success
    } catch (error) {
      message.error('操作失败')
      return false
    } finally {
      routerSubmitting.value = false
    }
  }

  /**
   * 打开全局过滤器配置对话框
   */
  function openGlobalFilterConfigDialog(instance: GatewayInstance) {
    currentGatewayInstanceId.value = instance.gatewayInstanceId
    globalFilterConfigDialogVisible.value = true
  }

  /**
   * 处理右键菜单点击
   */
  async function handleMenuClick({ code, node }: { code: string; node?: TreeOption }) {
    if (!node) return

    const instanceOption = node as InstanceTreeOption
    if (!instanceOption.instance) return

    const instance = instanceOption.instance

    switch (code) {
      case 'routerConfig':
        await openRouterDialog(instance)
        break
      case 'globalFilterConfig':
        openGlobalFilterConfigDialog(instance)
        break
      default:
        break
    }
  }

  return {
    // Model
    model,

    // Service
    service,

    // Router配置对话框状态
    routerFormDialogVisible,
    routerFormDialogMode,
    currentEditRouter,
    routerSubmitting,
    getRouterFormInitialData,
    currentGatewayInstanceId,

    // 全局过滤器配置对话框状态
    globalFilterConfigDialogVisible,

    // 渲染函数
    renderNodePrefix,
    renderNodeLabel,
    renderNodeSuffix,

    // 事件处理
    handleTreeSelect,
    handlePageChange,
    handleRefresh,
    handleMenuClick,
    handleRouterFormSubmit,
  }
}

/**
 * 网关实例树 Page 类型
 */
export type GatewayInstanceTreePage = ReturnType<typeof useGatewayInstanceTreePage>

