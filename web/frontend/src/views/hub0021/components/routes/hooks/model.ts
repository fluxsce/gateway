/**
 * 路由配置列表管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField, RsDataFormRenderContext, RsDataFormTab } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { RsCheckbox, RsTag, setByNamePath } from '@/ui'
import { h, ref } from 'vue'
import { ServiceDefinitionSelector } from '../../services'
import type { RouteConfig } from '../types'
import { BackendType, MatchType, normalizeRedirectStatus, resolveListBackend } from '../types'

/**
 * 路由配置表格配置（对齐 RsGrid Props 子集）。
 */
export interface RouteConfigGridConfig {
  columns: RsGridColumn<RouteConfig>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

const matchTypeLabelMap: Record<number, string> = {
  [MatchType.EXACT]: '精确匹配',
  [MatchType.PREFIX]: '前缀匹配',
  [MatchType.REGEX]: '正则匹配',
}

const matchTypeVariantMap: Record<number, 'success' | 'info' | 'warning' | 'default'> = {
  [MatchType.EXACT]: 'success',
  [MatchType.PREFIX]: 'info',
  [MatchType.REGEX]: 'warning',
}

const methodStyleMap: Record<string, Record<string, string>> = {
  GET: {
    backgroundColor: 'rgba(96, 165, 250, 0.1)',
    color: '#60a5fa',
    borderColor: 'rgba(96, 165, 250, 0.3)',
  },
  POST: {
    backgroundColor: 'rgba(52, 211, 153, 0.1)',
    color: '#34d399',
    borderColor: 'rgba(52, 211, 153, 0.3)',
  },
  PUT: {
    backgroundColor: 'rgba(251, 191, 36, 0.1)',
    color: '#fbbf24',
    borderColor: 'rgba(251, 191, 36, 0.3)',
  },
  DELETE: {
    backgroundColor: 'rgba(248, 113, 113, 0.1)',
    color: '#f87171',
    borderColor: 'rgba(248, 113, 113, 0.3)',
  },
  PATCH: {
    backgroundColor: 'rgba(129, 140, 248, 0.1)',
    color: '#818cf8',
    borderColor: 'rgba(129, 140, 248, 0.3)',
  },
  HEAD: {
    backgroundColor: '#f5f5f5',
    color: '#999',
    borderColor: '#d0d0d0',
  },
  OPTIONS: {
    backgroundColor: '#f5f5f5',
    color: '#999',
    borderColor: '#d0d0d0',
  },
}

const methodTagBaseStyle: Record<string, string> = {
  display: 'inline-block',
  padding: '2px 6px',
  borderRadius: '4px',
  fontSize: '11px',
  fontWeight: '500',
  lineHeight: '1.4',
  whiteSpace: 'nowrap',
  flexShrink: '0',
  border: '1px solid',
}

/** 解析允许的 HTTP 方法数组（数组、JSON 字符串或逗号分隔） */
function getAllowedMethods(allowedMethods?: string[] | string): string[] {
  if (!allowedMethods) return []
  if (Array.isArray(allowedMethods)) return allowedMethods
  try {
    const parsed = JSON.parse(allowedMethods)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return allowedMethods.split(',').map((method) => method.trim()).filter(Boolean)
  }
}

/** 判断是否为多个服务（serviceDefinitionId 含逗号） */
function isMultipleServices(serviceDefinitionId?: string): boolean {
  if (!serviceDefinitionId) return false
  return serviceDefinitionId.includes(',')
}

/** 从逗号分隔字符串解析服务 ID 列表 */
function getServiceIds(serviceDefinitionId?: string): string[] {
  if (!serviceDefinitionId) return []
  return serviceDefinitionId
    .split(',')
    .map((id) => id.trim())
    .filter((id) => id)
}

/** 优先从 routeMetadata.serviceNameMap 取服务显示名 */
function getServiceDisplayName(row: RouteConfig, serviceId: string): string {
  const metadata = row.routeMetadata as any
  if (!metadata) return serviceId
  const obj =
    typeof metadata === 'string'
      ? (() => {
          try {
            return JSON.parse(metadata)
          } catch {
            return {}
          }
        })()
      : metadata
  const map = obj?.serviceNameMap
  return map && typeof map === 'object' && map[serviceId] ? map[serviceId] : serviceId
}

/** 渲染 HTTP 方法芯片（最多展示 2 个，其余用 +N） */
function renderAllowedMethods(row: RouteConfig) {
  const methods = getAllowedMethods(row.allowedMethods)
  if (methods.length === 0) {
    return h(
      'span',
      {
        style: {
          ...methodTagBaseStyle,
          color: '#999',
          fontStyle: 'italic',
          border: 'none',
          backgroundColor: 'transparent',
        },
      },
      '全部',
    )
  }

  const display = methods.slice(0, 2)
  const remaining = Math.max(0, methods.length - 2)

  return h(
    'div',
    {
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: '4px',
        maxWidth: '100%',
        whiteSpace: 'nowrap',
        overflow: 'hidden',
      },
    },
    [
      ...display.map((method) => {
        const colorStyle = methodStyleMap[method.toUpperCase()] || {
          backgroundColor: '#f5f5f5',
          color: '#666',
          borderColor: '#e0e0e0',
        }
        return h(
          'span',
          {
            style: {
              ...methodTagBaseStyle,
              backgroundColor: colorStyle.backgroundColor,
              color: colorStyle.color,
              borderColor: colorStyle.borderColor,
            },
          },
          method,
        )
      }),
      remaining > 0
        ? h(
            'span',
            {
              style: {
                ...methodTagBaseStyle,
                backgroundColor: '#f5f5f5',
                color: '#666',
                borderColor: '#e0e0e0',
              },
            },
            `+${remaining}`,
          )
        : null,
    ],
  )
}

function renderBackendType(row: RouteConfig) {
  const backend = resolveListBackend(row)
  if (backend === BackendType.STATIC) {
    return h(RsTag, { size: 'sm', variant: 'warning' }, () => '静态资源')
  }
  if (backend === BackendType.REDIRECT) {
    return h(RsTag, { size: 'sm', variant: 'success' }, () => '重定向')
  }
  if (backend === BackendType.PROXY) {
    return h(RsTag, { size: 'sm', variant: 'info' }, () => '服务代理')
  }
  return h(RsTag, { size: 'sm', variant: 'danger' }, () => '未配置')
}

function isProxyBackend(formData: Record<string, any>): boolean {
  return formData.backendType === BackendType.PROXY || !formData.backendType
}

function isStaticBackend(formData: Record<string, any>): boolean {
  return formData.backendType === BackendType.STATIC
}

function isRedirectBackend(formData: Record<string, any>): boolean {
  return formData.backendType === BackendType.REDIRECT
}

/** 渲染关联服务标签（单服务 / 多服务） */
function renderServiceName(row: RouteConfig) {
  if (resolveListBackend(row) === BackendType.STATIC) {
    const root = String(row.staticRootDirectory || '').trim()
    return h(RsTag, { size: 'sm', variant: 'warning' }, () => root || '本机目录')
  }
  if (resolveListBackend(row) === BackendType.REDIRECT) {
    const status = normalizeRedirectStatus(row.redirectStatus)
    const target = String(row.redirectLocation || '').trim() || '未填写目标'
    return h(RsTag, { size: 'sm', variant: 'success' }, () => `${status} ${target}`)
  }

  if (row.serviceName) {
    return h(RsTag, { size: 'sm', variant: 'success' }, () => row.serviceName)
  }

  if (row.serviceDefinitionId) {
    const serviceDefinitionId = row.serviceDefinitionId
    if (isMultipleServices(serviceDefinitionId)) {
      const ids = getServiceIds(serviceDefinitionId)
      return h(
        'div',
        {
          style: {
            display: 'flex',
            flexWrap: 'wrap',
            alignItems: 'center',
            gap: '2px',
          },
        },
        [
          ...ids.map((serviceId) =>
            h(
              RsTag,
              {
                size: 'sm',
                variant: 'info',
                style: { marginRight: '4px', marginBottom: '2px' },
              },
              () => getServiceDisplayName(row, serviceId),
            ),
          ),
          h(
            RsTag,
            { size: 'sm', variant: 'default', style: { marginLeft: '4px' } },
            () => `${ids.length}个服务`,
          ),
        ],
      )
    }

    return h(RsTag, { size: 'sm', variant: 'info' }, () =>
      getServiceDisplayName(row, serviceDefinitionId),
    )
  }

  return h(RsTag, { size: 'sm', variant: 'default' }, () => '未关联')
}

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

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
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
        field: 'backendType',
        label: '后端',
        type: 'select',
        placeholder: '请选择后端',
        span: 6,
        clearable: true,
        options: [
          { label: '服务代理', value: BackendType.PROXY },
          { label: '静态资源', value: BackendType.STATIC },
          { label: '重定向', value: BackendType.REDIRECT },
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
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新增路由配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '批量删除选中的路由配置',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: RouteConfigGridConfig = {
    columns: [
      {
        key: 'routeConfigId',
        title: '路由配置ID',
        visible: false,
      },
      {
        key: 'routeName',
        title: '路由名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 200,
        render: (row) =>
          h(
            'span',
            { style: { color: 'var(--g-primary, #7c3aed)' } },
            row.routeName,
          ),
      },
      {
        key: 'routePath',
        title: '路由路径',
        align: 'center',
        ellipsis: true,
        width: 250,
      },
      {
        key: 'matchType',
        title: '匹配类型',
        align: 'center',
        width: 120,
        render: (row) =>
          h(
            RsTag,
            {
              variant: matchTypeVariantMap[row.matchType] || 'default',
              size: 'sm',
            },
            () => matchTypeLabelMap[row.matchType] || '未知',
          ),
      },
      {
        key: 'backendType',
        title: '后端',
        align: 'center',
        width: 110,
        render: (row) => renderBackendType(row),
      },
      {
        key: 'routePriority',
        title: '优先级',
        align: 'center',
        sortable: true,
        width: 100,
      },
      {
        key: 'allowedMethods',
        title: 'HTTP方法',
        align: 'center',
        width: 150,
        render: (row) => renderAllowedMethods(row),
      },
      {
        key: 'serviceName',
        title: '目标',
        align: 'center',
        ellipsis: true,
        width: 200,
        render: (row) => renderServiceName(row),
      },
      {
        key: 'timeoutMs',
        title: '总超时(ms)',
        align: 'center',
        width: 110,
      },
      {
        key: 'stripPathPrefix',
        title: '剥前缀',
        align: 'center',
        width: 80,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.stripPathPrefix === 'Y' ? 'warning' : 'default',
              size: 'sm',
            },
            () => (row.stripPathPrefix === 'Y' ? '剥除' : '保留'),
          ),
      },
      {
        key: 'enableWebsocket',
        title: 'WebSocket',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.enableWebsocket === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.enableWebsocket === 'Y' ? '已标记' : '兼容'),
          ),
      },
      {
        key: 'retryCount',
        title: '重试次数',
        align: 'center',
        visible: false,
      },
      {
        key: 'routeMetadata',
        title: '路由元数据',
        align: 'center',
        visible: false,
      },
      {
        key: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
      {
        key: 'addTime',
        title: '创建时间',
        align: 'center',
        width: 180,
        formatter: (_v, row) => formatDate(row.addTime),
      },
    ],
    selectable: true,
    rowKey: 'routeConfigId',
    height: '100%',
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'edit', label: '编辑', icon: 'pencil' },
        {
          key: 'routeConfig',
          label: '路由配置',
          icon: 'settings',
          children: [
            { key: 'assertConfig', label: '路由断言配置', icon: 'circle-check' },
            { key: 'ipAccessControl', label: 'IP访问控制', icon: 'lock' },
            { key: 'userAgentAccessControl', label: 'User-Agent访问控制', icon: 'user' },
            { key: 'apiAccessControl', label: 'API访问控制', icon: 'link' },
            { key: 'domainAccessControl', label: '域名访问控制', icon: 'globe' },
            { key: 'corsConfig', label: '跨域配置', icon: 'link' },
            { key: 'authConfig', label: '认证配置', icon: 'settings' },
            { key: 'rateLimitConfig', label: '限流配置', icon: 'settings' },
            { key: 'filters', label: '路由过滤器', icon: 'settings' },
          ],
        },
        { key: 'staticHostConfig', label: '静态资源', icon: 'folder' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
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

  /** 路由表单配置（用于 RsDataFormModal） */
  const routeFormConfig = {
    tabs: [
      {
        key: 'basic',
        label: '基本信息',
      },
      {
        key: 'forward',
        label: '转发策略',
        show: (formData: Record<string, any>) => isProxyBackend(formData),
      },
      {
        key: 'metadata',
        label: '元数据配置',
      },
      {
        key: 'other',
        label: '其他',
      },
    ] as RsDataFormTab[],
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
        field: 'backendType',
        label: '后端',
        type: 'select' as const,
        placeholder: '请选择后端',
        span: 24,
        tabKey: 'basic',
        required: true,
        defaultValue: BackendType.PROXY,
        tips: '请求命中后从哪里出内容。服务代理：转到已选的后端服务。静态资源：从本机目录出文件，保存后填写目录。重定向：回 301/302/307/308，同页填写目标地址。',
        options: [
          { label: '服务代理', value: BackendType.PROXY },
          { label: '静态资源', value: BackendType.STATIC },
          { label: '重定向', value: BackendType.REDIRECT },
        ],
        rules: [
          {
            required: true,
            message: '请选择后端',
            trigger: ['blur', 'change'],
            validator: (value: unknown) => {
              if (
                value !== BackendType.PROXY &&
                value !== BackendType.STATIC &&
                value !== BackendType.REDIRECT
              ) {
                return '请选择后端'
              }
              return true
            },
          },
        ],
      },
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
            validator: (value: unknown) => {
              if (value === null || value === undefined || value === '') {
                return '请选择匹配类型'
              }
              const validValues = [MatchType.EXACT, MatchType.PREFIX, MatchType.REGEX]
              if (!validValues.includes(Number(value))) {
                return '请选择有效的匹配类型'
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
            validator: (value: unknown) => {
              if (typeof value !== 'string' || !value) {
                return true
              }
              if (!value.startsWith('/')) {
                return '路由路径必须以 / 开头'
              }
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
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
          const methods = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'HEAD', 'OPTIONS']
          const currentValue = getAllowedMethods(ctx ? ctx.value : formData.allowedMethods)

          return h(
            'div',
            { style: { display: 'flex', flexWrap: 'wrap', gap: '8px' } },
            methods.map((method) =>
              h(
                RsCheckbox,
                {
                  modelValue: currentValue.includes(method),
                  'onUpdate:modelValue': (checked: boolean) => {
                    const next = new Set(currentValue.map((v: string) => String(v)))
                    if (checked) next.add(method)
                    else next.delete(method)
                    const list = Array.from(next)
                    if (ctx?.onUpdate) ctx.onUpdate(list)
                    else formData.allowedMethods = list
                  },
                },
                { default: () => method },
              ),
            ),
          )
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
            validator: (value: unknown) => {
              if (value === null || value === undefined || value === '') {
                return '请输入路由优先级'
              }
              const num = typeof value === 'number' ? value : Number(value)
              if (isNaN(num)) {
                return '路由优先级必须是数字'
              }
              if (num < 1 || num > 999) {
                return '优先级必须在1-999之间'
              }
              return true
            },
          },
        ],
      },
      {
        field: 'serviceDefinitionId',
        label: '关联服务',
        type: 'custom' as const,
        span: 24,
        tabKey: 'basic',
        required: false,
        show: (formData: Record<string, any>) => isProxyBackend(formData),
        tips: '服务代理必须选择当前网关实例下已启用的服务定义。可多选以并行转发。',
        rules: [],
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext & { gatewayInstanceId?: string }) => {
          const gatewayInstanceId = ctx?.gatewayInstanceId || ''
          const rawId = ctx ? ctx.value : formData.serviceDefinitionId
          let currentId = ''
          if (typeof rawId === 'string') currentId = rawId
          else if (rawId != null) currentId = String(rawId)

          return h(ServiceDefinitionSelector, {
            modelValue: currentId,
            'onUpdate:modelValue': (value: string | null) => {
              if (ctx?.onUpdate) ctx.onUpdate(value ?? '')
              else formData.serviceDefinitionId = value ?? ''
            },
            'onUpdate:serviceNameMap': (map: Record<string, string>) => {
              if (ctx?.setFieldValue) ctx.setFieldValue('routeMetadata.serviceNameMap', map)
              else setByNamePath(formData, 'routeMetadata.serviceNameMap', map)
            },
            gatewayInstanceId,
          })
        },
      },
      {
        field: '_staticBackendHint',
        label: '静态资源',
        type: 'custom' as const,
        span: 24,
        tabKey: 'basic',
        show: (formData: Record<string, any>) => isStaticBackend(formData),
        tips: '保存后会打开静态资源配置，填写网站文件目录。登录、限流、跨域仍按本路由生效。',
        render: () =>
          h(
            'div',
            {
              style: {
                lineHeight: '1.6',
                color: 'var(--g-text-secondary, #64748b)',
              },
            },
            '不转到后端服务。保存后填写本机文件夹（如 Vue/React 的 dist），访问者看到的是该目录里的网页和文件。',
          ),
      },
      {
        field: 'redirectStatus',
        label: '重定向状态码',
        type: 'select' as const,
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 301,
        show: (formData: Record<string, any>) => isRedirectBackend(formData),
        tips: '301/308 永久跳转，可被缓存；302/307 临时跳转，不缓存。301/302 可能把 POST 改成 GET，要保留原方法用 307/308。旧登录入口一般用 301。',
        options: [
          { label: '301 永久重定向', value: 301 },
          { label: '302 临时重定向', value: 302 },
          { label: '307 临时重定向（保留方法）', value: 307 },
          { label: '308 永久重定向（保留方法）', value: 308 },
        ],
        rules: [
          {
            required: true,
            message: '请选择重定向状态码',
            trigger: ['blur', 'change'],
            validator: (value: unknown) => {
              if (normalizeRedirectStatus(value) !== Number(value)) {
                return '请选择 301、302、307 或 308'
              }
              return true
            },
          },
        ],
      },
      {
        field: 'redirectLocation',
        label: '重定向目标',
        type: 'input' as const,
        span: 24,
        tabKey: 'basic',
        required: true,
        placeholder: '例如 /#/datahublogin 或 https://www.example.com/#/datahublogin',
        show: (formData: Record<string, any>) => isRedirectBackend(formData),
        tips: '浏览器 Location。优先写站点内绝对路径 /#/datahublogin。绝对地址可用 https://... 或 {scheme}://{host}/...；scheme 只认 TLS 或单一合法的 X-Forwarded-Proto，host 只认请求 Host。不要写 //host。',
        rules: [
          {
            required: true,
            message: '请填写重定向目标',
            trigger: ['blur', 'input'],
            validator: (value: unknown) => {
              const loc = typeof value === 'string' ? value.trim() : ''
              if (!loc) {
                return '请填写重定向目标'
              }
              if (loc.length > 500) {
                return '重定向目标不能超过500个字符'
              }
              if (/[\r\n]/.test(loc)) {
                return '重定向目标不能包含换行'
              }
              if (loc.startsWith('//')) {
                return '不能使用协议相对地址'
              }
              return true
            },
          },
        ],
      },
      // ============= 多服务配置字段（NamePath，写在 routeMetadata 对象上） =============
      {
        field: 'routeMetadata.responseMergeStrategy',
        label: '响应合并策略',
        type: 'select' as const,
        span: 8,
        tabKey: 'basic',
        show: (formData: Record<string, any>) => {
          return isProxyBackend(formData) && formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
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
          return isProxyBackend(formData) && formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
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
          return isProxyBackend(formData) && formData.serviceDefinitionId && formData.serviceDefinitionId.includes(',')
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
      // ============= 转发策略 Tab =============
      {
        field: 'routeMetadata.overrideProxyTimeout',
        label: '覆盖代理超时/重试',
        type: 'switch' as const,
        span: 24,
        tabKey: 'forward',
        defaultValue: 'N',
        tips: '仅允许 Y/N。默认 N：沿用代理超时与重试，兼容历史路由。仅当为 Y 时，才使用下方路由超时/重试覆盖代理',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'timeoutMs',
        label: '请求总超时（毫秒）',
        type: 'number' as const,
        placeholder: '0表示仍沿用代理总超时',
        span: 12,
        tabKey: 'forward',
        required: true,
        defaultValue: 0,
        show: (formData: Record<string, any>) => formData.routeMetadata?.overrideProxyTimeout === 'Y',
        tips: '仅在开启「覆盖代理超时/重试」后生效。大于0时覆盖代理总超时；0表示该项仍用代理。SSE收到事件流响应头后也会停止绝对总超时',
        props: {
          min: 0,
          max: 3600000,
          style: { width: '100%' },
        },
        rules: [
          {
            required: true,
            message: '请输入请求总超时',
            trigger: ['blur', 'change'],
            type: 'number',
            validator: (value: unknown) => {
              if (value === null || value === undefined || value === '') {
                return '请输入请求总超时'
              }
              const num = Number(value)
              if (Number.isNaN(num) || num < 0) {
                return '请求总超时不能为负数'
              }
              if (num > 3600000) {
                return '请求总超时不能超过3600000毫秒'
              }
              return true
            },
          },
        ],
      },
      {
        field: 'retryCount',
        label: '重试次数',
        type: 'number' as const,
        placeholder: '请输入重试次数',
        span: 12,
        tabKey: 'forward',
        defaultValue: 0,
        show: (formData: Record<string, any>) => formData.routeMetadata?.overrideProxyTimeout === 'Y',
        tips: '仅在开启覆盖后生效。须与「重试间隔」同时大于0才覆盖代理重试；任一为0则重试仍用代理',
        props: {
          min: 0,
          max: 20,
          style: { width: '100%' },
        },
      },
      {
        field: 'retryIntervalMs',
        label: '重试间隔（毫秒）',
        type: 'number' as const,
        placeholder: '0表示沿用代理重试间隔',
        span: 12,
        tabKey: 'forward',
        defaultValue: 0,
        show: (formData: Record<string, any>) => formData.routeMetadata?.overrideProxyTimeout === 'Y',
        tips: '仅在开启覆盖后生效。须与「重试次数」同时大于0才覆盖代理重试。表示失败后到下次尝试前的等待，不是单次执行超时',
        props: {
          min: 0,
          max: 300000,
          style: { width: '100%' },
        },
      },
      {
        field: 'stripPathPrefix',
        label: '剥离路径前缀',
        type: 'switch' as const,
        span: 12,
        tabKey: 'forward',
        defaultValue: 'N',
        show: (formData: Record<string, any>) => isProxyBackend(formData),
        tips: '只影响反向代理拼上游路径。Y：去掉已匹配路由前缀后再拼到节点路径；N：保持历史 nginx 拼接。静态资源路由不显示本项。',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'enableWebsocket',
        label: 'WebSocket标记',
        type: 'switch' as const,
        span: 12,
        tabKey: 'forward',
        defaultValue: 'N',
        tips: '路由级标记。N仍允许HTTP Upgrade/专项WebSocket（兼容历史默认N）；Y表示明确标识本路由面向WebSocket。真正准入看握手请求与代理类型',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      {
        field: 'rewritePath',
        label: '重写路径',
        type: 'input' as const,
        placeholder: '留空表示不重写，如 /stream/events',
        span: 24,
        tabKey: 'forward',
        show: (formData: Record<string, any>) => isProxyBackend(formData),
        tips: '只影响反向代理。非空时整段替换为该路径，不再拼接客户端剩余路径。静态资源路由不显示本项。',
        rules: [
          {
            max: 200,
            message: '重写路径不能超过200个字符',
            trigger: ['blur', 'input'],
          },
        ],
      },
      // ============= 元数据配置 Tab =============
      {
        field: '_routeMetadataPreview',
        label: '路由元数据',
        type: 'custom' as const,
        span: 24,
        tabKey: 'metadata',
        tips: '只读预览当前 routeMetadata 对象。结构化字段在其它页签编辑；接口中的其它键会随对象一并保存。',
        render: (formData: Record<string, any>) => {
          const meta = formData.routeMetadata
          const text =
            meta && typeof meta === 'object' && !Array.isArray(meta)
              ? JSON.stringify(meta, null, 2)
              : '{}'
          return h('textarea', {
            value: text,
            readOnly: true,
            rows: 8,
            class: 'rs-data-form__textarea',
            style:
              'width:100%;box-sizing:border-box;padding:8px 12px;border:1px solid var(--rs-border-color, #d9d9d9);border-radius:4px;font:inherit;resize:vertical;background:var(--rs-fill-color, #fafafa);',
          })
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
        show: false,
        tips: '路由配置的操作序列标识',
      },
    ] as RsDataFormField[],
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
