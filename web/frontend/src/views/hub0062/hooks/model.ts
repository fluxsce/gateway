/**
 * 隧道客户端管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import type { ConnectionStatus, TunnelClient } from '../types'

/**
 * 隧道客户端表格配置（对齐 RsGrid Props 子集）。
 */
export interface TunnelClientGridConfig {
  columns: RsGridColumn<TunnelClient>[]
  selectable: boolean
  rowKey: string | ((row: TunnelClient) => string)
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/** 连接状态选项 */
export const CONNECTION_STATUS_OPTIONS = [
  { label: '已连接', value: 'connected' as ConnectionStatus, variant: 'success' as RsTagVariant },
  { label: '已断开', value: 'disconnected' as ConnectionStatus, variant: 'warning' as RsTagVariant },
  { label: '连接中', value: 'connecting' as ConnectionStatus, variant: 'info' as RsTagVariant },
  { label: '错误', value: 'error' as ConnectionStatus, variant: 'danger' as RsTagVariant },
]

/** 活动标记选项 */
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' },
]

/** TLS启用选项 */
export const TLS_ENABLE_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' },
]

/** 自动重连选项 */
export const AUTO_RECONNECT_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' },
]

/**
 * 获取连接状态展示文案
 */
function getConnectionStatusText(status: ConnectionStatus): string {
  const option = CONNECTION_STATUS_OPTIONS.find((opt) => opt.value === status)
  return option?.label || status
}

/**
 * 获取连接状态标签变体
 */
function getConnectionStatusVariant(status: ConnectionStatus): RsTagVariant {
  const option = CONNECTION_STATUS_OPTIONS.find((opt) => opt.value === status)
  return option?.variant || 'default'
}

/**
 * 隧道客户端管理 Model
 */
export function useTunnelClientModel() {
  const moduleId = 'hub0062:tunnel-client'
  const loading = ref(false)
  const clientList = ref<TunnelClient[]>([])
  const pageInfo = ref<PageInfoObj | undefined>()

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'clientName',
        label: '客户端名称',
        type: 'input',
        placeholder: '请输入客户端名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serverAddress',
        label: '服务器地址',
        type: 'input',
        placeholder: '请输入服务器地址',
        span: 6,
        clearable: true,
      },
      {
        field: 'connectionStatus',
        label: '连接状态',
        type: 'select',
        placeholder: '请选择连接状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...CONNECTION_STATUS_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'activeFlag',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...ACTIVE_FLAG_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增客户端',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新增隧道客户端',
      },
      {
        key: 'connect',
        label: '连接',
        icon: 'LinkOutline',
        type: 'success',
        tooltip: '连接选中的客户端',
      },
      {
        key: 'disconnect',
        label: '断开',
        icon: 'StopOutline',
        type: 'warning',
        tooltip: '断开选中的客户端',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '删除选中的客户端',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'connection', label: '连接配置' },
    { key: 'advanced', label: '高级配置' },
  ]

  const formFields: RsDataFormField[] = [
    {
      field: 'clientName',
      label: '客户端名称',
      type: 'input',
      required: true,
      placeholder: '请输入客户端名称',
      span: 12,
      tabKey: 'basic',
      tips: '用于标识客户端的唯一名称，建议使用有意义的名称便于管理',
    },
    {
      field: 'clientDescription',
      label: '客户端描述',
      type: 'textarea',
      placeholder: '请输入客户端描述',
      span: 24,
      tabKey: 'basic',
      props: { rows: 3 },
      tips: '客户端的详细描述信息，可以包含用途、负责人等信息',
    },
    {
      field: 'activeFlag',
      label: '状态',
      type: 'select',
      required: true,
      options: ACTIVE_FLAG_OPTIONS,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      tips: '客户端的激活状态，禁用后客户端将无法连接到服务器',
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'basic',
      props: { rows: 3 },
      tips: '其他需要记录的信息，如维护记录、注意事项等',
    },
    {
      field: 'serverAddress',
      label: '服务器地址',
      type: 'input',
      required: true,
      placeholder: '请输入服务器地址',
      span: 12,
      tabKey: 'connection',
      tips: '隧道服务器的地址，可以是域名或IP地址（如：frps.example.com）',
    },
    {
      field: 'serverPort',
      label: '服务器端口',
      type: 'number',
      required: true,
      placeholder: '请输入服务器端口',
      span: 12,
      tabKey: 'connection',
      props: { min: 1, max: 65535 },
      tips: '隧道服务器监听的端口号，需与服务器配置一致',
    },
    {
      field: 'authToken',
      label: '认证令牌',
      type: 'input',
      required: true,
      placeholder: '请输入认证令牌',
      span: 24,
      tabKey: 'connection',
      tips: '用于客户端身份验证的令牌，由服务器提供，确保连接安全',
    },
    {
      field: 'tlsEnable',
      label: 'TLS加密',
      type: 'select',
      required: true,
      options: TLS_ENABLE_OPTIONS,
      span: 12,
      tabKey: 'connection',
      defaultValue: 'N',
      tips: '是否启用TLS加密传输，启用后数据传输更安全但会增加性能开销',
    },
    {
      field: 'autoReconnect',
      label: '自动重连',
      type: 'select',
      required: true,
      options: AUTO_RECONNECT_OPTIONS,
      span: 12,
      tabKey: 'advanced',
      defaultValue: 'Y',
      tips: '连接断开后是否自动重连，建议启用以保证服务稳定性',
    },
    {
      field: 'maxRetries',
      label: '最大重试次数',
      type: 'number',
      placeholder: '请输入最大重试次数',
      span: 12,
      tabKey: 'advanced',
      props: { min: 0, max: 100 },
      defaultValue: 3,
      tips: '自动重连时的最大尝试次数，0表示不限制重试次数',
    },
    {
      field: 'retryInterval',
      label: '重试间隔(秒)',
      type: 'number',
      placeholder: '请输入重试间隔',
      span: 12,
      tabKey: 'advanced',
      props: { min: 1, max: 300 },
      defaultValue: 30,
      tips: '两次重连尝试之间的等待时间，避免频繁重连造成服务器压力',
    },
    {
      field: 'heartbeatInterval',
      label: '心跳间隔(秒)',
      type: 'number',
      placeholder: '请输入心跳间隔',
      span: 12,
      tabKey: 'advanced',
      props: { min: 10, max: 300 },
      defaultValue: 30,
      tips: '客户端向服务器发送心跳包的时间间隔，用于保持连接活跃',
    },
    {
      field: 'heartbeatTimeout',
      label: '心跳超时(秒)',
      type: 'number',
      placeholder: '请输入心跳超时',
      span: 12,
      tabKey: 'advanced',
      props: { min: 10, max: 300 },
      defaultValue: 90,
      tips: '心跳响应的超时时间，超时后会触发重连机制',
    },
  ]

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: TunnelClientGridConfig = {
    columns: [
      {
        key: 'clientName',
        title: '客户端名称',
        width: 180,
        ellipsis: true,
      },
      {
        key: 'serverAddress',
        title: '服务器地址',
        width: 200,
        ellipsis: true,
        formatter: (_value, row) => `${row.serverAddress}:${row.serverPort}`,
      },
      {
        key: 'connectionStatus',
        title: '连接状态',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            { variant: getConnectionStatusVariant(row.connectionStatus), size: 'sm' },
            () => getConnectionStatusText(row.connectionStatus),
          ),
      },
      {
        key: 'tlsEnable',
        title: 'TLS',
        width: 80,
        align: 'center',
        formatter: (_value, row) => (row.tlsEnable === 'Y' ? '启用' : '禁用'),
      },
      {
        key: 'autoReconnect',
        title: '自动重连',
        width: 100,
        align: 'center',
        formatter: (_value, row) => (row.autoReconnect === 'Y' ? '启用' : '禁用'),
      },
      {
        key: 'serviceCount',
        title: '服务数量',
        width: 100,
        align: 'center',
      },
      {
        key: 'reconnectCount',
        title: '重连次数',
        width: 100,
        align: 'center',
      },
      {
        key: 'lastConnectTime',
        title: '最后连接时间',
        width: 180,
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
      },
      {
        key: 'lastHeartbeat',
        title: '最后心跳',
        width: 180,
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-'),
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
      {
        key: 'addTime',
        title: '创建时间',
        width: 180,
        sortable: true,
        ellipsis: true,
        formatter: (value) => (value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : ''),
      },
    ],
    selectable: true,
    rowKey: (row) => `${row.tunnelClientId}::${row.tenantId}`,
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
        { key: 'connect', label: '连接', icon: 'link' },
        { key: 'disconnect', label: '断开连接', icon: 'unplug' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
      ],
    },
    height: '100%',
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
   * 获取连接状态标签
   */
  function getConnectionStatusLabel(status: ConnectionStatus): string {
    return getConnectionStatusText(status)
  }

  /**
   * 获取连接状态标签类型
   */
  function getConnectionStatusTagType(status: ConnectionStatus): RsTagVariant {
    return getConnectionStatusVariant(status)
  }

  return {
    moduleId,
    loading,
    clientList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,
    getConnectionStatusLabel,
    getConnectionStatusTagType,
    resetPagination,
    updatePagination,
  }
}

/**
 * 隧道客户端管理 Model 类型
 */
export type TunnelClientModel = ReturnType<typeof useTunnelClientModel>
