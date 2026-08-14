/**
 * 隧道服务管理 - 数据模型定义
 * 定义服务表格配置、表单配置、选项等
 */

import type { RsDataFormField, RsDataFormRenderContext } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import type { TunnelService } from '../../../types'
import { TunnelClientSelector } from '../../tunnel-client-grid'

/**
 * 隧道服务表格配置（对齐 RsGrid Props 子集）。
 */
export interface TunnelServiceGridConfig {
  columns: RsGridColumn<TunnelService>[]
  selectable: boolean
  rowKey: string | ((row: TunnelService) => string)
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/** 服务类型选项 */
export const SERVICE_TYPE_OPTIONS = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'STCP', value: 'stcp' },
  { label: 'SUDP', value: 'sudp' },
  { label: 'XTCP', value: 'xtcp' },
]

/** 服务状态选项 */
export const SERVICE_STATUS_OPTIONS = [
  { label: '活动', value: 'active' },
  { label: '不活动', value: 'inactive' },
  { label: '错误', value: 'error' },
  { label: '离线', value: 'offline' },
]

/** 激活标识选项 */
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' },
]

/** 是否选项 */
export const YES_NO_OPTIONS = [
  { label: '是', value: 'Y' },
  { label: '否', value: 'N' },
]

/**
 * 获取服务类型标签变体
 */
function getServiceTypeVariant(type: string): RsTagVariant {
  switch (type) {
    case 'tcp':
    case 'http':
      return 'primary'
    case 'udp':
    case 'https':
      return 'success'
    case 'stcp':
    case 'xtcp':
      return 'info'
    default:
      return 'warning'
  }
}

/**
 * 获取服务状态标签变体
 */
function getServiceStatusVariant(status: string): RsTagVariant {
  switch (status) {
    case 'active':
      return 'success'
    case 'inactive':
      return 'warning'
    case 'error':
      return 'danger'
    case 'offline':
      return 'default'
    default:
      return 'default'
  }
}

/**
 * 隧道服务管理 Model
 */
export function useTunnelServiceModel() {
  const moduleId = 'hub0062:service'
  const loading = ref(false)
  const serviceList = ref<TunnelService[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 6,
        clearable: true,
        options: SERVICE_TYPE_OPTIONS,
      },
      {
        field: 'serviceStatus',
        label: '服务状态',
        type: 'select',
        placeholder: '请选择服务状态',
        span: 6,
        clearable: true,
        options: SERVICE_STATUS_OPTIONS,
      },
      {
        field: 'keyword',
        label: '关键词',
        type: 'input',
        placeholder: '服务名称/本地地址/子域名',
        span: 6,
        clearable: true,
      },
    ],
    toolbarButtons: [
      { key: 'create', label: '新增服务', type: 'primary', icon: 'AddOutline', tooltip: '新增隧道服务' },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const formTabs = [
    { key: 'basic', label: '基础信息' },
    { key: 'address', label: '地址配置' },
    { key: 'advanced', label: '高级配置' },
  ]

  const formFields: RsDataFormField[] = [
    {
      field: 'serviceName',
      label: '服务名称',
      type: 'input',
      required: true,
      tabKey: 'basic',
      span: 12,
      placeholder: '请输入服务名称',
      props: { maxlength: 100 },
      tips: '服务的唯一标识名称',
    },
    {
      field: 'serviceDescription',
      label: '服务描述',
      type: 'textarea',
      tabKey: 'basic',
      span: 24,
      placeholder: '请输入服务描述',
      props: { rows: 3, maxlength: 500 },
      tips: '服务的详细说明',
    },
    {
      field: 'tunnelClientId',
      label: '客户端ID',
      type: 'custom',
      required: true,
      tabKey: 'basic',
      span: 12,
      render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
        return h(TunnelClientSelector, {
          modelValue: (ctx?.value as string) || formData.tunnelClientId || '',
          'onUpdate:modelValue': (value: string) => {
            ctx?.onUpdate(value)
          },
        })
      },
      tips: '服务所属的隧道客户端ID',
    },
    {
      field: 'serviceType',
      label: '服务类型',
      type: 'select',
      required: true,
      tabKey: 'basic',
      span: 12,
      placeholder: '请选择服务类型',
      options: SERVICE_TYPE_OPTIONS,
      tips: '服务的协议类型',
    },
    {
      field: 'localAddress',
      label: '本地地址',
      type: 'input',
      required: true,
      tabKey: 'address',
      span: 12,
      placeholder: '例如: 127.0.0.1',
      props: { maxlength: 100 },
      tips: '本地服务的IP地址',
    },
    {
      field: 'localPort',
      label: '本地端口',
      type: 'number',
      required: true,
      tabKey: 'address',
      span: 12,
      placeholder: '请输入本地端口',
      props: { min: 1, max: 65535 },
      tips: '本地服务的端口号 (1-65535)',
    },
    {
      field: 'remotePort',
      label: '远程端口',
      type: 'number',
      tabKey: 'address',
      span: 12,
      placeholder: '服务器分配的远程端口',
      props: { min: 1, max: 65535 },
      tips: '服务器端暴露的端口号，留空则由服务器自动分配',
    },
    {
      field: 'subDomain',
      label: '子域名',
      type: 'input',
      tabKey: 'address',
      span: 12,
      placeholder: '例如: myapp',
      props: { maxlength: 100 },
      tips: 'HTTP/HTTPS服务的子域名',
    },
    {
      field: 'customDomains',
      label: '自定义域名',
      type: 'textarea',
      tabKey: 'address',
      span: 24,
      placeholder: '多个域名用逗号分隔',
      props: { rows: 2 },
      tips: 'HTTP/HTTPS服务的自定义域名列表',
    },
    {
      field: 'useEncryption',
      label: '启用加密',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      options: YES_NO_OPTIONS,
      tips: '是否对传输数据进行加密',
    },
    {
      field: 'useCompression',
      label: '启用压缩',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      options: YES_NO_OPTIONS,
      tips: '是否对传输数据进行压缩',
    },
    {
      field: 'secretKey',
      label: '加密密钥',
      type: 'input',
      tabKey: 'advanced',
      span: 24,
      placeholder: '请输入加密密钥',
      props: { type: 'password', maxlength: 100 },
      tips: '用于加密的密钥，启用加密时必填',
    },
    {
      field: 'maxConnections',
      label: '最大连接数',
      type: 'number',
      tabKey: 'advanced',
      span: 12,
      placeholder: '0表示不限制',
      props: { min: 0 },
      tips: '服务允许的最大并发连接数',
    },
    {
      field: 'bandwidthLimit',
      label: '带宽限制',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      placeholder: '例如: 1MB, 100KB',
      props: { maxlength: 50 },
      tips: '服务的带宽限制，如: 1MB, 100KB',
    },
    {
      field: 'httpUser',
      label: 'HTTP用户名',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      placeholder: 'HTTP基础认证用户名',
      props: { maxlength: 100 },
      tips: 'HTTP/HTTPS服务的基础认证用户名',
    },
    {
      field: 'httpPassword',
      label: 'HTTP密码',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      placeholder: 'HTTP基础认证密码',
      props: { type: 'password', maxlength: 100 },
      tips: 'HTTP/HTTPS服务的基础认证密码',
    },
    {
      field: 'hostHeaderRewrite',
      label: 'Host头重写',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      placeholder: '例如: example.com',
      props: { maxlength: 200 },
      tips: 'HTTP/HTTPS服务的Host头重写',
    },
    {
      field: 'healthCheckType',
      label: '健康检查类型',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      placeholder: '请选择健康检查类型',
      options: [
        { label: 'TCP', value: 'tcp' },
        { label: 'HTTP', value: 'http' },
      ],
      tips: '服务的健康检查方式',
    },
    {
      field: 'healthCheckUrl',
      label: '健康检查URL',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      placeholder: '例如: /health',
      props: { maxlength: 200 },
      tips: 'HTTP健康检查的URL路径',
    },
    {
      field: 'activeFlag',
      label: '激活状态',
      type: 'select',
      tabKey: 'basic',
      span: 12,
      defaultValue: 'Y',
      options: ACTIVE_FLAG_OPTIONS,
      tips: '服务是否激活',
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      tabKey: 'basic',
      span: 24,
      placeholder: '请输入备注信息',
      props: { rows: 3, maxlength: 500 },
      tips: '服务的备注说明',
    },
  ]

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: TunnelServiceGridConfig = {
    columns: [
      { key: 'tunnelServiceId', title: '服务ID', width: 200, ellipsis: true },
      { key: 'serviceName', title: '服务名称', width: 150, ellipsis: true, sortable: true },
      { key: 'tunnelClientId', title: '客户端ID', width: 200, ellipsis: true },
      {
        key: 'serviceType',
        title: '服务类型',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: getServiceTypeVariant(row.serviceType), size: 'sm' },
            () => getServiceTypeLabel(row.serviceType),
          ),
      },
      {
        key: 'localAddress',
        title: '本地地址',
        width: 180,
        ellipsis: true,
        formatter: (_value, row) => `${row.localAddress}:${row.localPort}`,
      },
      { key: 'remotePort', title: '远程端口', width: 100 },
      { key: 'subDomain', title: '子域名', width: 150, ellipsis: true },
      {
        key: 'serviceStatus',
        title: '服务状态',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: getServiceStatusVariant(row.serviceStatus), size: 'sm' },
            () => getServiceStatusLabel(row.serviceStatus),
          ),
      },
      { key: 'connectionCount', title: '当前连接', width: 100, align: 'right' },
      { key: 'totalConnections', title: '总连接数', width: 100, align: 'right' },
      {
        key: 'registeredTime',
        title: '注册时间',
        width: 160,
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'lastActiveTime',
        title: '最后活动',
        width: 160,
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
      {
        key: 'activeFlag',
        title: '状态',
        width: 80,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
            () => (row.activeFlag === 'Y' ? '启用' : '禁用'),
          ),
      },
    ],
    selectable: true,
    rowKey: (row) => `${row.tunnelServiceId}::${row.tenantId}`,
    paginationConfig: {
      show: true,
      pageInfo: pageInfo as any,
      align: 'right',
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'register', label: '注册服务', icon: 'upload' },
        { key: 'unregister', label: '注销服务', icon: 'circle-stop' },
        { key: 'edit', label: '编辑', icon: 'pencil' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
      ],
    },
    height: '100%',
  }

  /**
   * 获取服务类型标签
   */
  const getServiceTypeLabel = (type: string): string => {
    const option = SERVICE_TYPE_OPTIONS.find((opt) => opt.value === type)
    return option?.label || type.toUpperCase()
  }

  /**
   * 获取服务状态标签
   */
  const getServiceStatusLabel = (status: string): string => {
    const option = SERVICE_STATUS_OPTIONS.find((opt) => opt.value === status)
    return option?.label || status
  }

  /**
   * 获取服务状态标签类型
   */
  const getServiceStatusTagType = (status: string): RsTagVariant => {
    return getServiceStatusVariant(status)
  }

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息（接收后端 PageInfoObj）
   */
  const updatePagination = (newPageInfo: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = newPageInfo as PageInfoObj
    } else {
      Object.assign(pageInfo.value, newPageInfo)
    }
  }

  /**
   * 设置服务列表
   */
  const setServiceList = (list: TunnelService[]) => {
    serviceList.value = list
  }

  /**
   * 清空服务列表
   */
  const clearServiceList = () => {
    serviceList.value = []
  }

  /**
   * 添加服务到列表
   */
  const addServiceToList = (service: TunnelService) => {
    serviceList.value.unshift(service)
  }

  /**
   * 更新列表中的服务
   */
  const updateServiceInList = (
    tunnelServiceId: string,
    tenantId: string,
    updatedService: Partial<TunnelService>,
  ) => {
    const index = serviceList.value.findIndex(
      (s) => s.tunnelServiceId === tunnelServiceId && s.tenantId === tenantId,
    )
    if (index !== -1) {
      Object.assign(serviceList.value[index], updatedService)
    }
  }

  /**
   * 从列表中删除服务
   */
  const removeServiceFromList = (tunnelServiceId: string, tenantId: string) => {
    const index = serviceList.value.findIndex(
      (s) => s.tunnelServiceId === tunnelServiceId && s.tenantId === tenantId,
    )
    if (index !== -1) {
      serviceList.value.splice(index, 1)
    }
  }

  return {
    moduleId,
    loading,
    serviceList,
    pageInfo,
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,
    resetPagination,
    updatePagination,
    setServiceList,
    clearServiceList,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList,
    getServiceTypeLabel,
    getServiceStatusLabel,
    getServiceStatusTagType,
  }
}

/**
 * Model 返回类型
 */
export type TunnelServiceModel = ReturnType<typeof useTunnelServiceModel>
