/**
 * 过滤器配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, TrashOutline } from '@vicons/ionicons5'
import { NSwitch } from 'naive-ui'
import { h, ref } from 'vue'
import type { FilterConfig } from './types'
import {
  BODY_MODIFIER_OPTIONS,
  CONTENT_TYPES,
  COOKIE_OPERATION_OPTIONS,
  FILTER_ACTION_OPTIONS,
  FILTER_TYPE_OPTIONS,
  HEADER_MODIFIER_OPTIONS,
  HTTP_METHODS,
  METHOD_FILTER_MODE_OPTIONS,
  PATH_REWRITE_MODE_OPTIONS,
  QUERY_PARAM_MODIFIER_OPTIONS,
  RESPONSE_OPERATION_OPTIONS,
} from './types'

/**
 * 过滤器配置列表 Model
 */
export function useFilterConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0021-filter-config'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 过滤器配置列表数据 */
  const filterList = ref<FilterConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'filterName',
        label: '过滤器名称',
        type: 'input',
        placeholder: '请输入过滤器名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'filterType',
        label: '过滤器类型',
        type: 'select',
        placeholder: '请选择过滤器类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...FILTER_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'filterAction',
        label: '执行时机',
        type: 'select',
        placeholder: '请选择执行时机',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...FILTER_ACTION_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
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
        label: '新增过滤器',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增过滤器配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的过滤器配置',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 获取过滤器类型显示标签 */
  const getFilterTypeLabel = (filterType: string) => {
    const option = FILTER_TYPE_OPTIONS.find(opt => opt.value === filterType)
    return option?.label || filterType
  }

  /** 获取过滤器类型标签颜色 */
  const getFilterTypeTagType = (filterType: string): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const typeColorMap: Record<string, "default" | "success" | "error" | "warning" | "primary" | "info"> = {
      'header': 'primary',
      'query-param': 'info',
      'body': 'warning',
      'strip': 'success',
      'rewrite': 'success',
      'method': 'error',
      'cookie': 'default',
      'response': 'info'
    }
    return typeColorMap[filterType] || 'default'
  }

  /** 获取执行时机标签 */
  const getFilterActionLabel = (filterAction: string) => {
    const option = FILTER_ACTION_OPTIONS.find(opt => opt.value === filterAction)
    return option?.label || filterAction
  }

  /** 获取执行时机标签颜色 */
  const getFilterActionTagType = (filterAction: string): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const actionColorMap: Record<string, "default" | "success" | "error" | "warning" | "primary" | "info"> = {
      'pre-routing': 'success',
      'post-routing': 'info',
      'pre-response': 'warning'
    }
    return actionColorMap[filterAction] || 'default'
  }

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'filterConfigId',
        title: '过滤器配置ID',
        visible: false,
        width: 0,
      },
      {
        field: 'filterOrder',
        title: '执行顺序',
        align: 'center',
        width: 120,
        slots: { default: 'filterOrder' },
      },
      {
        field: 'filterName',
        title: '过滤器名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'filterType',
        title: '类型',
        align: 'center',
        width: 140,
        slots: { default: 'filterType' },
      },
      {
        field: 'filterAction',
        title: '执行时机',
        align: 'center',
        width: 120,
        slots: { default: 'filterAction' },
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'filterDesc',
        title: '描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 200,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.addTime),
        width: 180,
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        showOverflow: 'tooltip',
        width: 120,
      },
      {
        field: 'editTime',
        title: '修改时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.editTime),
        width: 180,
      },
      {
        field: 'editWho',
        title: '修改人',
        align: 'center',
        showOverflow: 'tooltip',
        width: 120,
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
          code: 'toggle-status',
          name: '切换状态',
          prefixIcon: 'vxe-icon-check',
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
   * 设置过滤器列表
   */
  function setFilterList(list: FilterConfig[]) {
    filterList.value = list
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
   * 添加过滤器到列表
   */
  function addFilterToList(filter: FilterConfig) {
    filterList.value.push(filter)
    // 按 filterOrder 排序
    filterList.value.sort((a, b) => (a.filterOrder || 0) - (b.filterOrder || 0))
  }

  /**
   * 更新列表中的过滤器
   * @param filterConfigId 过滤器配置ID
   * @param tenantId 租户ID（可选，用于精确匹配）
   * @param updatedFilter 更新的过滤器数据
   */
  function updateFilterInList(filterConfigId: string, tenantId: string | undefined, updatedFilter: Partial<FilterConfig>) {
    const index = filterList.value.findIndex(
      (f) => f.filterConfigId === filterConfigId && (!tenantId || f.tenantId === tenantId)
    )
    if (index !== -1) {
      // 使用 Object.assign 合并更新，保持响应式
      Object.assign(filterList.value[index], updatedFilter)
      // 按 filterOrder 排序（因为更新后顺序可能改变）
      filterList.value.sort((a, b) => (a.filterOrder || 0) - (b.filterOrder || 0))
    }
  }

  /**
   * 从列表中移除过滤器
   */
  function removeFilterFromList(filterConfigId: string) {
    const index = filterList.value.findIndex((f) => f.filterConfigId === filterConfigId)
    if (index >= 0) {
      filterList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除过滤器
   */
  function removeFiltersFromList(filterConfigIds: string[]) {
    filterList.value = filterList.value.filter((f) => !filterConfigIds.includes(f.filterConfigId))
  }

  // ============= 表单配置 =============

  /** 表单页签配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'other',
      label: '其他信息',
    },
  ]

  /** 过滤器表单配置（用于 GdataFormModal） */
  const formFields: DataFormField[] = [
    // 主键字段（隐藏，但必须存在用于编辑）
    {
      field: 'filterConfigId',
      label: '过滤器配置ID',
      type: 'input' as const,
      span: 12,
      primary: true,
      show: false,
    },
    {
      field: 'gatewayInstanceId',
      label: '网关实例ID',
      type: 'input' as const,
      span: 12,
      show: false,
    },
    {
      field: 'routeConfigId',
      label: '路由配置ID',
      type: 'input' as const,
      span: 12,
      show: false,
    },
    // 基本信息
    {
      field: 'filterName',
      label: '过滤器名称',
      type: 'input' as const,
      placeholder: '请输入过滤器名称',
      span: 12,
      required: true,
      rules: [
        { required: true, message: '请输入过滤器名称', trigger: ['blur', 'input'] },
        { max: 100, message: '过滤器名称不能超过100个字符', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'filterType',
      label: '过滤器类型',
      type: 'select' as const,
      placeholder: '请选择过滤器类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'header',
      options: FILTER_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择过滤器类型', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'filterAction',
      label: '执行时机',
      type: 'select' as const,
      placeholder: '请选择执行时机',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'pre-routing',
      show: false, // 隐藏执行时机字段，默认就是前置处理
      options: FILTER_ACTION_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择执行时机', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'filterOrder',
      label: '执行顺序',
      type: 'number' as const,
      placeholder: '数值越小优先级越高',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 1,
      props: {
        min: 1,
        precision: 0,
      },
      rules: [
        { 
          required: true, 
          type: 'number',
          message: '请输入执行顺序', 
          trigger: ['blur', 'change'],
          validator: (_rule: any, value: any) => {
            if (value === null || value === undefined || value === '') {
              return new Error('请输入执行顺序')
            }
            const num = typeof value === 'number' ? value : Number(value)
            if (isNaN(num)) {
              return new Error('执行顺序必须是数字')
            }
            if (num < 1) {
              return new Error('执行顺序必须大于等于1')
            }
            if (!Number.isInteger(num)) {
              return new Error('执行顺序必须是整数')
            }
            return true
          }
        },
      ],
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch' as const,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'filterDesc',
      label: '描述',
      type: 'input' as const,
      placeholder: '请输入过滤器描述',
      span: 24,
      tabKey: 'basic',
      props: {
        type: 'textarea',
        rows: 3,
        maxlength: 500,
        showCount: true,
      },
    },
    // ============= 过滤器配置（使用 fieldset 包裹） =============
    {
      field: 'filterConfigFieldset',
      label: '过滤器配置',
      type: 'fieldset' as const,
      span: 24,
      tabKey: 'basic',
      children: [
            // Header过滤器配置
        {
          field: 'config.headerConfig.modifierType',
          label: '修改类型',
          type: 'select' as const,
          placeholder: '请选择修改类型',
          span: 12,
          show: (formData: Record<string, any>) => formData.filterType === 'header',
          options: HEADER_MODIFIER_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        },
    {
      field: 'config.headerConfig.isRequestHeader',
      label: '作用范围',
      type: 'custom' as const,
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'header',
      defaultValue: true,
      render: (formData: Record<string, any>) => {
        return h(NSwitch, {
          value: formData['config.headerConfig.isRequestHeader'] ?? true,
          checkedValue: true,
          uncheckedValue: false,
          'onUpdate:value': (value: boolean) => {
            formData['config.headerConfig.isRequestHeader'] = value
          },
        }, {
          checked: () => '请求头',
          unchecked: () => '响应头',
        })
      },
    },
    {
      field: 'config.headerConfig.headerName',
      label: 'Header名称',
      type: 'input' as const,
      placeholder: '请输入Header名称',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'header',
    },
    {
      field: 'config.headerConfig.headerValue',
      label: 'Header值',
      type: 'input' as const,
      placeholder: '请输入Header值',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'header' && 
        formData['config.headerConfig.modifierType'] !== 'remove',
    },
    {
      field: 'config.headerConfig.targetHeaderName',
      label: '目标Header名称',
      type: 'input' as const,
      placeholder: '请输入重命名后的Header名称',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'header' && 
        formData['config.headerConfig.modifierType'] === 'rename',
    },
    
    // 查询参数过滤器配置
    {
      field: 'config.queryParamConfig.modifierType',
      label: '修改类型',
      type: 'select' as const,
      placeholder: '请选择修改类型',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'query-param',
      options: QUERY_PARAM_MODIFIER_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.queryParamConfig.paramName',
      label: '参数名称',
      type: 'input' as const,
      placeholder: '请输入参数名称',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'query-param',
    },
    {
      field: 'config.queryParamConfig.paramValue',
      label: '参数值',
      type: 'input' as const,
      placeholder: '请输入参数值',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'query-param' && 
        formData['config.queryParamConfig.modifierType'] !== 'remove',
    },
    {
      field: 'config.queryParamConfig.targetParamName',
      label: '目标参数名称',
      type: 'input' as const,
      placeholder: '请输入重命名后的参数名称',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'query-param' && 
        formData['config.queryParamConfig.modifierType'] === 'rename',
    },
    
    // 前缀剥离过滤器配置
    {
      field: 'config.stripConfig.prefix',
      label: '要剥离的前缀',
      type: 'input' as const,
      placeholder: '请输入要剥离的路径前缀，如：/api/v1',
      span: 24,
      show: (formData: Record<string, any>) => formData.filterType === 'strip',
    },
    
    // 路径重写过滤器配置
    {
      field: 'config.rewriteConfig.mode',
      label: '重写模式',
      type: 'select' as const,
      placeholder: '请选择重写模式',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'rewrite',
      options: PATH_REWRITE_MODE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.rewriteConfig.from',
      label: '查找内容',
      type: 'input' as const,
      placeholder: '请输入要替换的字符串或正则表达式',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'rewrite',
    },
    {
      field: 'config.rewriteConfig.to',
      label: '替换内容',
      type: 'input' as const,
      placeholder: '请输入替换后的内容',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'rewrite',
    },
    
    // 方法过滤器配置
    {
      field: 'config.methodConfig.mode',
      label: '过滤模式',
      type: 'select' as const,
      placeholder: '请选择过滤模式',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'method',
      options: METHOD_FILTER_MODE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.methodConfig.allowedMethods',
      label: '允许的方法',
      type: 'select' as const,
      placeholder: '请选择允许的HTTP方法',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'method' && 
        formData['config.methodConfig.mode'] === 'allow',
      props: {
        multiple: true,
      },
      options: HTTP_METHODS.map(method => ({ label: method, value: method })),
    },
    {
      field: 'config.methodConfig.deniedMethods',
      label: '拒绝的方法',
      type: 'select' as const,
      placeholder: '请选择拒绝的HTTP方法',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'method' && 
        formData['config.methodConfig.mode'] === 'deny',
      props: {
        multiple: true,
      },
      options: HTTP_METHODS.map(method => ({ label: method, value: method })),
    },
    {
      field: 'config.methodConfig.rejectStatusCode',
      label: '拒绝时状态码',
      type: 'number' as const,
      placeholder: '默认405',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'method',
      defaultValue: 405,
      props: {
        min: 400,
        max: 599,
      },
    },
    {
      field: 'config.methodConfig.rejectMessage',
      label: '拒绝时消息',
      type: 'input' as const,
      placeholder: '请输入拒绝时的错误消息',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'method',
      defaultValue: 'Method Not Allowed',
    },
    {
      field: 'config.methodConfig.caseSensitive',
      label: '区分大小写',
      type: 'switch' as const,
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'method',
      defaultValue: false,
      props: {
        checkedValue: true,
        uncheckedValue: false,
      },
    },
    
    // 请求体过滤器配置
    {
      field: 'config.bodyConfig.modifierType',
      label: '修改类型',
      type: 'select' as const,
      placeholder: '请选择修改类型',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'body',
      options: BODY_MODIFIER_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.bodyConfig.operation',
      label: '操作描述',
      type: 'input' as const,
      placeholder: '请输入操作描述',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'body',
    },
    {
      field: 'config.bodyConfig.allowedContentTypes',
      label: '允许的内容类型',
      type: 'select' as const,
      placeholder: '请选择允许的内容类型',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'body',
      props: {
        multiple: true,
      },
      options: CONTENT_TYPES.map(type => ({ label: type, value: type })),
    },
    {
      field: 'config.bodyConfig.maxBodySize',
      label: '最大请求体大小(字节)',
      type: 'number' as const,
      placeholder: '请输入最大请求体大小',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'body',
      props: {
        min: 0,
      },
    },
    {
      field: 'config.bodyConfig.filterConfigJson',
      label: '过滤器配置(JSON)',
      type: 'input' as const,
      placeholder: '请输入JSON格式的过滤器配置',
      span: 24,
      show: (formData: Record<string, any>) => formData.filterType === 'body',
      props: {
        type: 'textarea',
        rows: 4,
      },
    },
    
    // Cookie过滤器配置
    {
      field: 'config.cookieConfig.operation',
      label: '操作类型',
      type: 'select' as const,
      placeholder: '请选择操作类型',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'cookie',
      options: COOKIE_OPERATION_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.cookieConfig.applyToResponse',
      label: '应用到响应',
      type: 'switch' as const,
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'cookie',
      defaultValue: false,
      props: {
        checkedValue: true,
        uncheckedValue: false,
      },
    },
    {
      field: 'config.cookieConfig.cookieName',
      label: 'Cookie名称',
      type: 'input' as const,
      placeholder: '请输入Cookie名称',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'cookie',
    },
    {
      field: 'config.cookieConfig.cookieValue',
      label: 'Cookie值',
      type: 'input' as const,
      placeholder: '请输入Cookie值',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        formData['config.cookieConfig.operation'] !== 'remove',
    },
    {
      field: 'config.cookieConfig.cookieAttributes.domain',
      label: '域名',
      type: 'input' as const,
      placeholder: '如：.example.com',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
    },
    {
      field: 'config.cookieConfig.cookieAttributes.path',
      label: '路径',
      type: 'input' as const,
      placeholder: '如：/',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
    },
    {
      field: 'config.cookieConfig.cookieAttributes.maxAge',
      label: '最大年龄(秒)',
      type: 'number' as const,
      placeholder: '如：3600',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
      props: {
        min: 0,
      },
    },
    {
      field: 'config.cookieConfig.cookieAttributes.secure',
      label: '安全传输',
      type: 'switch' as const,
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
      defaultValue: false,
      props: {
        checkedValue: true,
        uncheckedValue: false,
      },
    },
    {
      field: 'config.cookieConfig.cookieAttributes.httpOnly',
      label: '仅HTTP',
      type: 'switch' as const,
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
      defaultValue: false,
      props: {
        checkedValue: true,
        uncheckedValue: false,
      },
    },
    {
      field: 'config.cookieConfig.cookieAttributes.sameSite',
      label: 'SameSite',
      type: 'select' as const,
      placeholder: '请选择SameSite属性',
      span: 12,
      show: (formData: Record<string, any>) => 
        formData.filterType === 'cookie' && 
        (formData['config.cookieConfig.operation'] === 'add' || 
         formData['config.cookieConfig.operation'] === 'modify'),
      options: [
        { label: 'Strict', value: 'Strict' },
        { label: 'Lax', value: 'Lax' },
        { label: 'None', value: 'None' },
      ],
    },
    
    // 响应过滤器配置
    {
      field: 'config.responseConfig.operation',
      label: '操作类型',
      type: 'select' as const,
      placeholder: '请选择操作类型',
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'response',
      options: RESPONSE_OPERATION_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'config.responseConfig.setInRequestPhase',
      label: '请求阶段设置',
      type: 'switch' as const,
      span: 12,
      show: (formData: Record<string, any>) => formData.filterType === 'response',
      defaultValue: false,
      props: {
        checkedValue: true,
        uncheckedValue: false,
      },
    },
    {
      field: 'config.responseConfig.filterConfigJson',
      label: '过滤器配置(JSON)',
      type: 'input' as const,
      placeholder: '请输入JSON格式的过滤器配置',
      span: 24,
      show: (formData: Record<string, any>) => formData.filterType === 'response',
      props: {
        type: 'textarea',
        rows: 4,
      },
    },
    {
      field: 'config.responseConfig.conditionsJson',
      label: '条件配置(JSON)',
      type: 'input' as const,
      placeholder: '请输入JSON格式的条件配置',
      span: 24,
      show: (formData: Record<string, any>) => formData.filterType === 'response',
      props: {
        type: 'textarea',
        rows: 3,
      },
    },
      ],
    },
    // ============= 其他信息（备注和时间） =============
    {
      field: 'noteText',
      label: '备注信息',
      type: 'input' as const,
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        type: 'textarea',
        rows: 3,
        maxlength: 500,
        showCount: true,
      },
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input' as const,
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
  ]

  return {
    // 状态
    moduleId,
    loading,
    filterList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 工具函数
    getFilterTypeLabel,
    getFilterTypeTagType,
    getFilterActionLabel,
    getFilterActionTagType,

    // 方法
    setFilterList,
    setLoading,
    resetPagination,
    updatePagination,
    addFilterToList,
    updateFilterInList,
    removeFilterFromList,
    removeFiltersFromList,
  }
}

/**
 * 过滤器配置列表 Model 类型
 */
export type FilterConfigModel = ReturnType<typeof useFilterConfigModel>


