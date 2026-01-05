/**
 * 网关实例树组件 Model
 * 统一管理数据状态和计算属性
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { ContextMenuConfig } from '@/components/gmenu'
import { FunnelOutline, SettingsOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { computed, h, ref } from 'vue'
import type { GatewayInstance, InstanceTreeOption } from '../types'
import { FilterExecutionMode } from '../types'

/**
 * 网关实例树 Model
 */
export function useGatewayInstanceTreeModel() {
  // ============= 数据状态 =============

  /** 模块ID */
  const moduleId = 'hub0021'

  /** 加载状态 */
  const loading = ref(false)

  /** 网关实例列表数据 */
  const instanceList = ref<GatewayInstance[]>([])

  /** 分页状态 */
  const currentPage = ref(1)
  const pageSize = ref(20) // 默认每页20条
  
  /** 数据总数（后端分页） */
  const totalCount = ref(0)

  /** 过滤关键词 */
  const filterKeyword = ref('')

  // ============= 计算属性 =============

  /**
   * 将实例列表转换为树形结构（后端分页，直接使用 instanceList）
   */
  const treeData = computed<InstanceTreeOption[]>(() => {
    return instanceList.value.map(instance => ({
      key: instance.gatewayInstanceId,
      label: getInstanceLabel(instance),
      instance: instance,
    }))
  })

  // ============= 辅助方法 =============

  /**
   * 获取实例标签
   */
  function getInstanceLabel(instance: GatewayInstance): string {
    const port = instance.tlsEnabled === 'Y' ? instance.httpsPort : instance.httpPort
    return `${instance.instanceName || '未命名'} (${instance.bindAddress || '-'}:${port || '-'})`
  }

  // ============= 状态更新方法 =============

  /**
   * 设置实例列表
   */
  function setInstanceList(list: GatewayInstance[]) {
    instanceList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 设置当前页
   */
  function setCurrentPage(page: number) {
    currentPage.value = page
  }

  /**
   * 设置每页大小
   */
  function setPageSize(size: number) {
    pageSize.value = size
  }

  /**
   * 设置过滤关键词
   */
  function setFilterKeyword(keyword: string) {
    filterKeyword.value = keyword
  }

  /**
   * 设置数据总数
   */
  function setTotalCount(count: number) {
    totalCount.value = count
  }

  /**
   * 重置分页到第一页
   */
  function resetPage() {
    currentPage.value = 1
  }

  /**
   * 清空搜索关键词
   */
  function clearFilter() {
    filterKeyword.value = ''
  }

  // ============= 右键菜单配置 =============

  /**
   * 树节点右键菜单配置
   */
  const treeMenuConfig: ContextMenuConfig = {
    enabled: true,
    showCopyNode: true,
    customMenus: [
      {
        code: 'routerConfig',
        name: 'Router配置',
        prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(SettingsOutline) }),
      },
      {
        code: 'globalFilterConfig',
        name: '全局过滤器配置',
        prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(FunnelOutline) }),
      },
    ],
  }

  // ============= Router配置表单配置 =============

  /**
   * Router配置表单字段配置
   * 用于 GdataFormModal 组件
   */
  const routerFormConfig = {
    tabs: [
      { 
        key: 'basic', 
        label: '基本信息',
      },
      { 
        key: 'cache', 
        label: '路由缓存配置',
      },
      { 
        key: 'filters', 
        label: '全局过滤器配置',
      },
      { 
        key: 'performance', 
        label: '性能优化配置',
      },
      { 
        key: 'error', 
        label: '错误处理配置',
      },
      { 
        key: 'other', 
        label: '其他配置',
      },
    ] as DataFormTab[],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于编辑） =============
      {
        field: 'routerConfigId',
        label: 'Router配置ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        primary: true,
        show: false,
      },
      {
        field: 'gatewayInstanceId',
        label: '网关实例ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        show: false,
      },
      // ============= 基本信息 Tab =============
      {
        field: 'routerName',
        label: 'Router名称',
        type: 'input' as const,
        placeholder: '请输入Router名称',
        span: 12,
        tabKey: 'basic',
        required: true,
        tips: 'Router配置的名称标识，用于区分不同的Router配置',
        rules: [
          { required: true, message: '请输入Router名称', trigger: ['blur', 'input'] },
          { max: 50, message: 'Router名称不能超过50个字符', trigger: ['blur', 'input'] },
        ],
      },
      {
        field: 'defaultPriority',
        label: '默认优先级',
        type: 'number' as const,
        placeholder: '请输入默认优先级',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 100,
        tips: '默认路由优先级，数值越小优先级越高',
        props: {
          min: 0,
          max: 9999,
        },
        rules: [
          { required: true, type: 'number', message: '请输入默认优先级', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'routerDesc',
        label: 'Router描述',
        type: 'textarea' as const,
        placeholder: '请输入Router描述信息',
        span: 24,
        tabKey: 'basic',
        tips: 'Router配置的描述说明，用于记录配置的用途、注意事项等信息',
        props: {
          rows: 2,
        },
      },
      // ============= 路由缓存配置 Tab =============
      {
        field: 'enableRouteCache',
        label: '启用路由缓存',
        type: 'switch' as const,
        span: 12,
        tabKey: 'cache',
        defaultValue: 'Y',
        tips: '是否启用路由缓存。启用后可以缓存路由匹配结果，提高路由匹配性能',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'routeCacheTtlSeconds',
        label: '缓存TTL(秒)',
        type: 'number' as const,
        placeholder: '请输入缓存TTL',
        span: 12,
        tabKey: 'cache',
        show: (formData: Record<string, any>) => formData.enableRouteCache === 'Y',
        required: true,
        defaultValue: 300,
        tips: '路由缓存的有效期，超过此时间的缓存将被清除',
        props: {
          min: 1,
          max: 86400,
        },
        rules: [
          { required: true, type: 'number', message: '请输入缓存TTL', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'maxRoutes',
        label: '最大路由数',
        type: 'number' as const,
        placeholder: '请输入最大路由数',
        span: 12,
        tabKey: 'cache',
        show: (formData: Record<string, any>) => formData.enableRouteCache === 'Y',
        defaultValue: 1000,
        tips: '路由缓存中最多可以缓存的路由数量',
        props: {
          min: 1,
          max: 10000,
        },
      },
      {
        field: 'routeMatchTimeout',
        label: '路由匹配超时(ms)',
        type: 'number' as const,
        placeholder: '请输入路由匹配超时时间',
        span: 12,
        tabKey: 'cache',
        show: (formData: Record<string, any>) => formData.enableRouteCache === 'Y',
        defaultValue: 5000,
        tips: '路由匹配的最大超时时间，超过此时间未匹配成功则返回未找到',
        props: {
          min: 100,
          max: 30000,
        },
      },
      // ============= 全局过滤器配置 Tab =============
      {
        field: 'enableGlobalFilters',
        label: '启用全局过滤器',
        type: 'switch' as const,
        span: 12,
        tabKey: 'filters',
        defaultValue: 'Y',
        tips: '是否启用全局过滤器。全局过滤器作用于所有路由，在路由匹配前执行',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'filterExecutionMode',
        label: '过滤器执行模式',
        type: 'select' as const,
        placeholder: '请选择执行模式',
        span: 12,
        tabKey: 'filters',
        show: (formData: Record<string, any>) => formData.enableGlobalFilters === 'Y',
        defaultValue: FilterExecutionMode.SEQUENTIAL,
        tips: '过滤器的执行模式：顺序执行或并行执行',
        options: [
          { label: '顺序执行', value: FilterExecutionMode.SEQUENTIAL },
          { label: '并行执行', value: FilterExecutionMode.PARALLEL },
        ],
      },
      {
        field: 'maxFilterChainDepth',
        label: '最大过滤器链深度',
        type: 'number' as const,
        placeholder: '请输入最大过滤器链深度',
        span: 12,
        tabKey: 'filters',
        show: (formData: Record<string, any>) => formData.enableGlobalFilters === 'Y',
        defaultValue: 10,
        tips: '过滤器链的最大深度，防止过滤器链过长导致性能问题',
        props: {
          min: 1,
          max: 100,
        },
      },
      {
        field: 'enableAsyncProcessing',
        label: '启用异步处理',
        type: 'switch' as const,
        span: 12,
        tabKey: 'filters',
        show: (formData: Record<string, any>) => formData.enableGlobalFilters === 'Y',
        defaultValue: 'N',
        tips: '是否启用过滤器的异步处理模式',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      // ============= 性能优化配置 Tab =============
      {
        field: 'enableStrictMode',
        label: '启用严格模式',
        type: 'switch' as const,
        span: 12,
        tabKey: 'performance',
        defaultValue: 'N',
        tips: '是否启用严格模式。严格模式下路由匹配更加严格，性能可能略有下降',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'enableRoutePooling',
        label: '启用路由池',
        type: 'switch' as const,
        span: 12,
        tabKey: 'performance',
        defaultValue: 'N',
        tips: '是否启用路由对象池。启用后可以复用路由对象，减少内存分配',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'routePoolSize',
        label: '路由池大小',
        type: 'number' as const,
        placeholder: '请输入路由池大小',
        span: 12,
        tabKey: 'performance',
        show: (formData: Record<string, any>) => formData.enableRoutePooling === 'Y',
        defaultValue: 100,
        tips: '路由对象池的大小，影响可以复用的路由对象数量',
        props: {
          min: 10,
          max: 1000,
        },
      },
      {
        field: 'caseSensitive',
        label: '大小写敏感',
        type: 'switch' as const,
        span: 12,
        tabKey: 'performance',
        defaultValue: 'N',
        tips: '路径匹配是否区分大小写',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'removeTrailingSlash',
        label: '移除尾部斜杠',
        type: 'switch' as const,
        span: 12,
        tabKey: 'performance',
        defaultValue: 'Y',
        tips: '是否自动移除路径尾部的斜杠进行匹配',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'enableMetrics',
        label: '启用监控指标',
        type: 'switch' as const,
        span: 12,
        tabKey: 'performance',
        defaultValue: 'Y',
        tips: '是否启用路由监控指标收集',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      // ============= 错误处理配置 Tab =============
      {
        field: 'enableFallback',
        label: '启用降级处理',
        type: 'switch' as const,
        span: 12,
        tabKey: 'error',
        defaultValue: 'N',
        tips: '是否启用降级处理。启用后当路由匹配失败时，可以降级到指定的路由',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'fallbackRoute',
        label: '降级路由',
        type: 'input' as const,
        placeholder: '请输入降级路由路径',
        span: 12,
        tabKey: 'error',
        show: (formData: Record<string, any>) => formData.enableFallback === 'Y',
        tips: '降级路由的路径，当路由匹配失败时使用此路由',
      },
      {
        field: 'notFoundStatusCode',
        label: '404状态码',
        type: 'number' as const,
        placeholder: '请输入404状态码',
        span: 12,
        tabKey: 'error',
        required: true,
        defaultValue: 404,
        tips: '路由未找到时返回的HTTP状态码',
        props: {
          min: 400,
          max: 599,
        },
        rules: [
          { required: true, type: 'number', message: '请输入404状态码', trigger: ['blur', 'change'] },
        ],
      },
      {
        field: 'notFoundMessage',
        label: '404消息',
        type: 'input' as const,
        placeholder: '请输入404消息',
        span: 12,
        tabKey: 'error',
        required: true,
        defaultValue: 'Route not found',
        tips: '路由未找到时返回的错误消息',
        rules: [
          { required: true, message: '请输入404消息', trigger: ['blur', 'input'] },
        ],
      },
      // ============= 其他配置 Tab =============
      {
        field: 'enableTracing',
        label: '启用链路追踪',
        type: 'switch' as const,
        span: 12,
        tabKey: 'other',
        defaultValue: 'N',
        tips: '是否启用链路追踪功能',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'activeFlag',
        label: '配置状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'other',
        defaultValue: 'Y',
        tips: 'Router配置的启用状态',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'noteText',
        label: '备注信息',
        type: 'textarea' as const,
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'other',
        tips: 'Router配置的备注说明',
        props: {
          rows: 3,
        },
      },
      {
        field: 'addTime',
        label: '创建时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: 'Router配置的创建时间',
      },
      {
        field: 'addWho',
        label: '创建人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: 'Router配置的创建人',
      },
      {
        field: 'editTime',
        label: '修改时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: 'Router配置的最后修改时间',
      },
      {
        field: 'editWho',
        label: '修改人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: 'Router配置的最后修改人',
      },
      {
        field: 'currentVersion',
        label: '版本号',
        type: 'number' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: 'Router配置的当前版本号',
      },
      {
        field: 'oprSeqFlag',
        label: '操作序列标识',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        show: false, // 隐藏字段，通常不需要显示
        tips: 'Router配置的操作序列标识',
      },
    ] as DataFormField[],
  }

  return {
    // 状态
    moduleId,
    loading,
    instanceList,
    currentPage,
    pageSize,
    totalCount,
    filterKeyword,

    // 计算属性
    treeData,

    // 配置
    treeMenuConfig,
    routerFormConfig,

    // 方法
    getInstanceLabel,
    setInstanceList,
    setLoading,
    setCurrentPage,
    setPageSize,
    setTotalCount,
    setFilterKeyword,
    resetPage,
    clearFilter,
  }
}

/**
 * 网关实例树 Model 类型
 */
export type GatewayInstanceTreeModel = ReturnType<typeof useGatewayInstanceTreeModel>

