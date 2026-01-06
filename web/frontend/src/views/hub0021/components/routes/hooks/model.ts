/**
 * 路由配置列表管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField, DataFormTab } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, CheckmarkCircleOutline, GlobeOutline, TrashOutline } from '@vicons/ionicons5'
import { NCheckbox, NCheckboxGroup, NIcon, NSpace } from 'naive-ui'
import { h, ref } from 'vue'
import { ServiceDefinitionSelector } from '../../services'
import type { RouteConfig } from '../types'
import { MatchType } from '../types'

/**
 * 路由配置列表管理 Model
 */
export function useRouteConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0021'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 路由配置列表数据 */
  const routeList = ref<RouteConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'routeName',
        label: '路由名称',
        type: 'input',
        placeholder: '请输入路由名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'routePath',
        label: '路由路径',
        type: 'input',
        placeholder: '请输入路由路径',
        span: 6,
        clearable: true,
      },
      {
        field: 'matchType',
        label: '匹配类型',
        type: 'select',
        placeholder: '请选择匹配类型',
        span: 6,
        clearable: true,
        options: [
          { label: '精确匹配', value: MatchType.EXACT },
          { label: '前缀匹配', value: MatchType.PREFIX },
          { label: '正则匹配', value: MatchType.REGEX },
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: 'Y' },
          { label: '禁用', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增路由',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增路由配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的路由配置',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'routeConfigId',
        title: '路由配置ID',
        visible: false, // 隐藏主键字段，但保留在数据中以便编辑时使用
        width: 0,
      },
      {
        field: 'routeName',
        title: '路由名称',
        sortable: true,
        align: 'center',
        showOverflow: 'tooltip',
        slots: { default: 'routeName' },
        width: 200,
      },
      {
        field: 'routePath',
        title: '路由路径',
        align: 'center',
        showOverflow: 'tooltip',
        width: 250,
      },
      {
        field: 'matchType',
        title: '匹配类型',
        align: 'center',
        slots: { default: 'matchType' },
        width: 120,
      },
      {
        field: 'routePriority',
        title: '优先级',
        align: 'center',
        sortable: true,
        width: 100,
      },
      {
        field: 'allowedMethods',
        title: 'HTTP方法',
        align: 'center',
        slots: { default: 'allowedMethods' },
        width: 150,
      },
      {
        field: 'serviceName',
        title: '关联服务',
        align: 'center',
        showOverflow: 'tooltip',
        width: 180,
      },
      // 隐藏字段（字段存在但不显示）
      {
        field: 'timeoutMs',
        title: '超时时间(ms)',
        align: 'center',
        visible: false,
        width: 0,
      },
      {
        field: 'retryCount',
        title: '重试次数',
        align: 'center',
        visible: false,
        width: 0,
      },
      {
        field: 'enableWebsocket',
        title: 'WebSocket',
        align: 'center',
        visible: false,
        width: 0,
      },
      {
        field: 'routeMetadata',
        title: '路由元数据',
        align: 'center',
        visible: false,
        width: 0,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        slots: { default: 'activeFlag' },
        width: 100,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.addTime),
        width: 180,
      },
    ],
    showCheckbox: true,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      showCopyRow: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: 'vxe-icon-eye-fill',
        },
        {
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'routeConfig',
          name: '路由配置',
          prefixIcon: 'vxe-icon-setting',
          children: [
            {
              code: 'assertConfig',
              name: '路由断言配置',
              prefixIcon: () => h(NIcon, { size: 12 }, { default: () => h(CheckmarkCircleOutline) }),
            },
            {
              code: 'ipAccessControl',
              name: 'IP访问控制',
              prefixIcon: 'vxe-icon-lock',
            },
            {
              code: 'userAgentAccessControl',
              name: 'User-Agent访问控制',
              prefixIcon: 'vxe-icon-user',
            },
            {
              code: 'apiAccessControl',
              name: 'API访问控制',
              prefixIcon: 'vxe-icon-link',
            },
            {
              code: 'domainAccessControl',
              name: '域名访问控制',
              prefixIcon: () => h(NIcon, { size: 12 }, { default: () => h(GlobeOutline) }),
            },
            {
              code: 'corsConfig',
              name: '跨域配置',
              prefixIcon: 'vxe-icon-link',
            },
            {
              code: 'authConfig',
              name: '认证配置',
              prefixIcon: 'vxe-icon-setting',
            },
            {
              code: 'rateLimitConfig',
              name: '限流配置',
              prefixIcon: 'vxe-icon-setting',
            },
            {
              code: 'filters',
              name: '路由过滤器',
              prefixIcon: 'vxe-icon-setting',
            },
          ],
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 设置路由列表
   */
  function setRouteList(list: RouteConfig[]) {
    routeList.value = list
  }

  /**
   * 设置加载状态
   */
  function setLoading(value: boolean) {
    loading.value = value
  }

  /**
   * 重置分页信息
   */
  function resetPagination() {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  function updatePagination(newPageInfo: Partial<PageInfoObj>) {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 添加路由到列表
   */
  function addRouteToList(route: RouteConfig) {
    routeList.value.unshift(route)
  }

  /**
   * 更新列表中的路由
   */
  function updateRouteInList(route: RouteConfig) {
    const index = routeList.value.findIndex((r) => r.routeConfigId === route.routeConfigId)
    if (index >= 0) {
      routeList.value[index] = route
    }
  }

  /**
   * 从列表中移除路由
   */
  function removeRouteFromList(routeConfigId: string) {
    const index = routeList.value.findIndex((r) => r.routeConfigId === routeConfigId)
    if (index >= 0) {
      routeList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除路由
   */
  function removeRoutesFromList(routeConfigIds: string[]) {
    routeList.value = routeList.value.filter((r) => !routeConfigIds.includes(r.routeConfigId))
  }

  // ============= 路由表单配置 =============

  /** 路由表单配置（用于 GdataFormModal） */
  const routeFormConfig = {
    tabs: [
      {
        key: 'basic',
        label: '基本信息',
      },
      {
        key: 'metadata',
        label: '元数据配置',
      },
      {
        key: 'other',
        label: '其他',
      },
    ] as DataFormTab[],
    fields: [
      // ============= 主键字段（隐藏，但必须存在用于编辑） =============
      {
        field: 'routeConfigId',
        label: '路由配置ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
        primary: true,
        show: false,
      },
      {
        field: 'tenantId',
        label: '租户ID',
        type: 'input' as const,
        span: 12,
        tabKey: 'basic',
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
        field: 'routeName',
        label: '路由名称',
        type: 'input' as const,
        placeholder: '请输入路由名称',
        span: 12,
        tabKey: 'basic',
        required: true,
        tips: '路由的唯一标识名称，用于区分不同的路由规则',
        rules: [
          { required: true, message: '请输入路由名称', trigger: ['blur', 'input'] },
          { max: 100, message: '路由名称不能超过100个字符', trigger: ['blur', 'input'] },
        ],
      },
      {
        field: 'matchType',
        label: '匹配类型',
        type: 'select' as const,
        placeholder: '请选择匹配类型',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: MatchType.PREFIX,
        tips: '精确匹配：路径必须完全一致；前缀匹配：路径以指定前缀开头；正则匹配：使用正则表达式匹配路径',
        options: [
          { label: '精确匹配', value: MatchType.EXACT },
          { label: '前缀匹配', value: MatchType.PREFIX },
          { label: '正则匹配', value: MatchType.REGEX },
        ],
        rules: [
          {
            required: true,
            message: '请选择匹配类型',
            trigger: ['blur', 'change'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请选择匹配类型')
              }
              // 验证值是否为有效的匹配类型（0, 1, 2）
              const validValues = [MatchType.EXACT, MatchType.PREFIX, MatchType.REGEX]
              if (!validValues.includes(Number(value))) {
                return new Error('请选择有效的匹配类型')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'routePath',
        label: '路由路径',
        type: 'input' as const,
        placeholder: '请输入路由路径',
        span: 24,
        tabKey: 'basic',
        required: true,
        tips: (formData: Record<string, any>) => {
          const matchType = formData.matchType
          switch (matchType) {
            case MatchType.EXACT:
              return '精确匹配：请求路径必须完全匹配配置的路径\n示例: /api/users/123'
            case MatchType.PREFIX:
              return '前缀匹配：请求路径以配置的路径为前缀即可匹配\n示例: /api/users (匹配 /api/users/*)'
            case MatchType.REGEX:
              return '正则匹配：使用正则表达式匹配请求路径\n示例: ^/api/users/\\d+$'
            default:
              return '精确匹配示例：/api/users；前缀匹配示例：/api/；正则匹配示例：^/api/(users|orders)/?$'
          }
        },
        props: {
          onUpdateValue: (value: string, formData: Record<string, any>) => {
            // 自动补全 / 前缀（参考 useRouteForm.ts 的 handlePathInput 逻辑）
            if (value && !value.startsWith('/')) {
              formData.routePath = '/' + value
            } else {
              formData.routePath = value
            }
          },
        },
        rules: [
          { required: true, message: '请输入路由路径', trigger: ['blur', 'input'] },
          {
            pattern: /^\/.*/,
            message: '路由路径必须以 / 开头',
            trigger: ['blur', 'input'],
          },
          {
            validator: (rule: any, value: any) => {
              // 参考 useRouteForm.ts 的验证逻辑
              if (!value) {
                return true
              }
              // 基本格式验证：必须以 / 开头
              if (!value.startsWith('/')) {
                return new Error('路由路径必须以 / 开头')
              }
              // 正则匹配时验证正则表达式有效性
              // 注意：这里无法直接访问 formData，需要在提交时再次验证
              // 或者可以通过 rule 的 context 获取，但 Naive UI 的 validator 不直接提供 formData
              // 所以这里只做基本验证，正则表达式验证在提交时处理
              return true
            },
            trigger: ['blur', 'input'],
          },
        ],
      },
      {
        field: 'allowedMethods',
        label: 'HTTP方法',
        type: 'custom' as const,
        span: 24,
        tabKey: 'basic',
        tips: '选择允许的HTTP请求方法，未选择表示允许所有方法',
        render: (formData: Record<string, any>) => {
          const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']
          const currentValue = Array.isArray(formData.allowedMethods)
            ? formData.allowedMethods
            : typeof formData.allowedMethods === 'string' && formData.allowedMethods
              ? (() => {
                  try {
                    const parsed = JSON.parse(formData.allowedMethods)
                    return Array.isArray(parsed) ? parsed : []
                  } catch {
                    return formData.allowedMethods.split(',').map((m: string) => m.trim()).filter(Boolean)
                  }
                })()
              : []

          return h(NCheckboxGroup, {
            value: currentValue,
            'onUpdate:value': (value: (string | number)[]) => {
              formData.allowedMethods = value.map(v => String(v))
            },
          }, {
            default: () =>
              h(NSpace, {}, {
                default: () =>
                  methods.map((method) =>
                    h(NCheckbox, { value: method, label: method }, { default: () => method })
                  ),
              }),
          })
        },
      },
      {
        field: 'allowedHosts',
        label: '允许的主机',
        type: 'input' as const,
        placeholder: '留空表示允许所有主机，多个主机用逗号分隔',
        span: 12,
        tabKey: 'basic',
        tips: '限制允许的主机名，多个主机用逗号分隔，如：api.example.com,www.example.com',
      },
      {
        field: 'routePriority',
        label: '路由优先级',
        type: 'number' as const,
        placeholder: '数值越小优先级越高',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 100,
        tips: '路由匹配的优先级，数值越小优先级越高，建议范围：1-999',
        props: {
          min: 1,
          max: 999,
          style: { width: '100%' },
        },
        rules: [
          {
            required: true,
            message: '请输入路由优先级',
            trigger: ['blur', 'change', 'input'],
            validator: (_rule: any, value: any) => {
              if (value === null || value === undefined || value === '') {
                return new Error('请输入路由优先级')
              }
              // 转换为数字
              const num = typeof value === 'number' ? value : Number(value)
              if (isNaN(num)) {
                return new Error('路由优先级必须是数字')
              }
              if (num < 1 || num > 999) {
                return new Error('优先级必须在1-999之间')
              }
              return true
            },
          },
        ],
      },
      {
        field: 'serviceDefinitionId',
        label: '关联服务', // ServiceDefinitionSelector组件内部已有label，这里设为空避免重复
        type: 'custom' as const,
        span: 24,
        tabKey: 'basic',
        required: true,
        tips: '选择要关联的后端服务定义，多个服务使用逗号分割，如果没有可用选项，请先在服务管理中创建服务定义',
        render: (formData: Record<string, any>, context?: { gatewayInstanceId?: string }) => {
          const gatewayInstanceId = context?.gatewayInstanceId || ''
          
          return h(ServiceDefinitionSelector, {
            modelValue: formData.serviceDefinitionId,
            'onUpdate:modelValue': (value: string | null) => {
              formData.serviceDefinitionId = value || ''
            },
            gatewayInstanceId,
          })
        },
      },
      // ============= 多服务配置字段（使用点号分隔，直接存储在 routeMetadata 中） =============
      {
        field: 'routeMetadata.responseMergeStrategy',
        label: '响应合并策略',
        type: 'select' as const,
        span: 8,
        tabKey: 'basic',
        show: (formData: Record<string, any>) => {
          // 只有当选择了多个服务时才显示
          return formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
        },
        defaultValue: 'first',
        tips: 'first: 使用第一个成功的响应（默认）\nfirst_error: 使用第一个失败的响应\nall: 返回所有响应',
        options: [
          { label: '第一个成功响应', value: 'first' },
          { label: '第一个失败响应', value: 'first_error' },
          { label: '所有响应', value: 'all' },
        ],
      },
      {
        field: 'routeMetadata.maxConcurrentRequests',
        label: '最大并发请求数',
        type: 'number' as const,
        span: 8,
        tabKey: 'basic',
        show: (formData: Record<string, any>) => {
          // 只有当选择了多个服务时才显示
          return formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
        },
        defaultValue: 0,
        tips: '0表示不限制（使用所有服务），大于0时限制并发数',
        props: {
          min: 0,
          precision: 0,
        },
      },
      {
        field: 'routeMetadata.requireAllSuccess',
        label: '要求所有服务成功',
        type: 'switch' as const,
        span: 8,
        tabKey: 'basic',
        show: (formData: Record<string, any>) => {
          // 只有当选择了多个服务时才显示
          return formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
        },
        defaultValue: false,
        tips: '如果为true，任何一个服务失败都会返回错误；如果为false，使用第一个成功的响应',
        props: {
          checkedValue: true,
          uncheckedValue: false,
        },
      },
      {
        field: 'logConfigId',
        label: '日志配置ID',
        type: 'input' as const,
        placeholder: '请输入日志配置ID（可选）',
        span: 12,
        tabKey: 'basic',
        tips: '关联的日志配置ID，用于路由请求的日志记录',
      },
      {
        field: 'activeFlag',
        label: '启用状态',
        type: 'switch' as const,
        span: 12,
        tabKey: 'basic',
        defaultValue: 'Y',
        tips: '控制路由是否启用，禁用的路由不会参与匹配',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      // ============= 元数据配置 Tab =============
      {
        field: 'routeMetadata',
        label: '路由元数据',
        type: 'input' as const,
        placeholder: '请输入JSON格式的元数据（可选）',
        span: 24,
        tabKey: 'metadata',
        tips: '用于存储路由的自定义元数据信息，格式为JSON字符串，如：{"key1":"value1","key2":"value2"}',
        props: {
          type: 'textarea',
          rows: 4,
        },
      },
      {
        field: 'noteText',
        label: '备注信息',
        type: 'input' as const,
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'metadata',
        tips: '路由配置的备注说明信息',
        props: {
          type: 'textarea',
          rows: 3,
          maxlength: 500,
          showCount: true,
        },
      },
      // ============= 其他配置 Tab =============
      {
        field: 'addTime',
        label: '创建时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: '路由配置的创建时间',
      },
      {
        field: 'addWho',
        label: '创建人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: '路由配置的创建人',
      },
      {
        field: 'editTime',
        label: '修改时间',
        type: 'datetime' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: '路由配置的最后修改时间',
      },
      {
        field: 'editWho',
        label: '修改人',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: '路由配置的最后修改人',
      },
      {
        field: 'currentVersion',
        label: '版本号',
        type: 'number' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        tips: '路由配置的当前版本号',
      },
      {
        field: 'oprSeqFlag',
        label: '操作序列标识',
        type: 'input' as const,
        span: 12,
        tabKey: 'other',
        disabled: true,
        show: false, // 隐藏字段，通常不需要显示
        tips: '路由配置的操作序列标识',
      },
    ] as DataFormField[],
  }

  return {
    // 状态
    moduleId,
    loading,
    routeList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    routeFormConfig,

    // 方法
    setRouteList,
    setLoading,
    resetPagination,
    updatePagination,
    addRouteToList,
    updateRouteInList,
    removeRouteFromList,
    removeRoutesFromList,
  }
}

/**
 * 路由配置列表 Model 类型
 */
export type RouteConfigModel = ReturnType<typeof useRouteConfigModel>

