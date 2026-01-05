/**
 * 隧道客户端管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import {
  AddOutline,
  EyeOutline,
  LinkOutline,
  StopOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { h, ref } from 'vue'
import type { ConnectionStatus, TunnelClient } from '../../../types'

// ============= 常量定义 =============

/** 连接状态选项 */
export const CONNECTION_STATUS_OPTIONS = [
  { label: '已连接', value: 'connected' as ConnectionStatus, type: 'success' as const },
  { label: '已断开', value: 'disconnected' as ConnectionStatus, type: 'warning' as const },
  { label: '连接中', value: 'connecting' as ConnectionStatus, type: 'info' as const },
  { label: '错误', value: 'error' as ConnectionStatus, type: 'error' as const },
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
 * 隧道客户端管理 Model
 */
export function useTunnelClientModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0062:tunnel-client'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 客户端列表数据 */
  const clientList = ref<TunnelClient[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
          ...CONNECTION_STATUS_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
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
          ...ACTIVE_FLAG_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增客户端',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增隧道客户端',
      },
      {
        key: 'connect',
        label: '连接',
        icon: LinkOutline,
        type: 'success',
        tooltip: '连接选中的客户端',
      },
      {
        key: 'disconnect',
        label: '断开',
        icon: StopOutline,
        type: 'warning',
        tooltip: '断开选中的客户端',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的客户端',
      },
    ],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'tunnelClientId',
        title: '客户端ID',
        width: 200,
        showOverflow: 'tooltip',
      },
      {
        field: 'clientName',
        title: '客户端名称',
        width: 180,
        showOverflow: 'tooltip',
      },
      {
        field: 'serverAddress',
        title: '服务器地址',
        width: 200,
        showOverflow: 'tooltip',
        formatter: ({ row }) => `${row.serverAddress}:${row.serverPort}`,
      },
      {
        field: 'connectionStatus',
        title: '连接状态',
        width: 100,
        align: 'center',
        slots: { default: 'connectionStatus' },
      },
      {
        field: 'tlsEnable',
        title: 'TLS',
        width: 80,
        align: 'center',
        formatter: ({ row }) => row.tlsEnable === 'Y' ? '启用' : '禁用',
      },
      {
        field: 'autoReconnect',
        title: '自动重连',
        width: 100,
        align: 'center',
        formatter: ({ row }) => row.autoReconnect === 'Y' ? '启用' : '禁用',
      },
      {
        field: 'serviceCount',
        title: '服务数量',
        width: 100,
        align: 'center',
      },
      {
        field: 'reconnectCount',
        title: '重连次数',
        width: 100,
        align: 'center',
      },
      {
        field: 'lastConnectTime',
        title: '最后连接时间',
        width: 180,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) => cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        field: 'lastHeartbeat',
        title: '最后心跳',
        width: 180,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) => cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        field: 'activeFlag',
        title: '状态',
        width: 80,
        align: 'center',
        slots: { default: 'activeFlag' },
      },
      {
        field: 'addTime',
        title: '创建时间',
        width: 180,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) => cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
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
      showCopyCell: true,
      customMenus: [
        {
          code: 'view',
          name: '查看详情',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(EyeOutline) }),
        },
        {
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit',
        },
        {
          code: 'connect',
          name: '连接',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(LinkOutline) }),
        },
        {
          code: 'disconnect',
          name: '断开连接',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(StopOutline) }),
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
  }

  // ============= 表单字段配置 =============

  /** 表单字段配置（用于新增/编辑对话框） */
  const formFields: DataFormField[] = [
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

  /** 表单标签页配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'connection',
      label: '连接配置',
    },
    {
      key: 'advanced',
      label: '高级配置',
    },
  ]

  // ============= 辅助方法 =============

  /**
   * 获取连接状态标签
   */
  function getConnectionStatusLabel(status: ConnectionStatus): string {
    const option = CONNECTION_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.label || status
  }

  /**
   * 获取连接状态标签类型
   */
  function getConnectionStatusTagType(status: ConnectionStatus): 'success' | 'warning' | 'info' | 'error' | 'default' {
    const option = CONNECTION_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.type || 'default'
  }

  /**
   * 重置分页
   */
  function resetPagination() {
    if (pageInfo.value) {
      pageInfo.value.pageIndex = 1
    }
  }

  /**
   * 更新分页信息
   */
  function updatePagination(info: { pageIndex?: number; pageSize?: number }) {
    if (pageInfo.value) {
      if (info.pageIndex !== undefined) {
        pageInfo.value.pageIndex = info.pageIndex
      }
      if (info.pageSize !== undefined) {
        pageInfo.value.pageSize = info.pageSize
      }
    }
  }

  return {
    // 数据状态
    moduleId,
    loading,
    clientList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 方法
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

