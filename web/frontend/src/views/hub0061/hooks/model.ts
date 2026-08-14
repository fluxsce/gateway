/**
 * 静态服务管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态（RsSearchForm / RsGrid / RsDataForm）
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type {
  RsGridColumn,
  RsGridMenuConfig,
  RsGridPaginationConfig,
} from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatBytes, formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import type {
  HealthCheckType,
  LoadBalanceType,
  ServerStatus,
  ServerType,
  TunnelStaticServer,
} from '../types'

/** 服务器类型选项 */
export const SERVER_TYPE_OPTIONS = [
  { label: 'TCP', value: 'tcp' as ServerType },
  { label: 'UDP', value: 'udp' as ServerType },
]

/** 服务器状态选项 */
export const SERVER_STATUS_OPTIONS = [
  { label: '运行中', value: 'running' as ServerStatus, type: 'success' as const },
  { label: '已停止', value: 'stopped' as ServerStatus, type: 'warning' as const },
  { label: '错误', value: 'error' as ServerStatus, type: 'error' as const },
]

/** 负载均衡类型选项 */
export const LOAD_BALANCE_OPTIONS = [
  { label: '轮询', value: 'roundrobin' as LoadBalanceType },
  { label: '最少连接', value: 'leastconn' as LoadBalanceType },
  { label: '随机', value: 'random' as LoadBalanceType },
]

/** 健康检查类型选项 */
export const HEALTH_CHECK_TYPE_OPTIONS = [
  { label: 'TCP', value: 'tcp' as HealthCheckType },
  { label: 'HTTP', value: 'http' as HealthCheckType },
  { label: 'HTTPS', value: 'https' as HealthCheckType },
]

/** 活动标记选项 */
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' },
]

/**
 * 静态服务表格配置（对齐 RsGrid Props 子集）。
 */
export interface StaticServerGridConfig {
  columns: RsGridColumn<TunnelStaticServer>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 将旧 naive-ui tag type 映射为 RsTag variant。
 */
function toTagVariant(type?: string): RsTagVariant {
  if (type === 'error') return 'danger'
  if (
    type === 'success' ||
    type === 'warning' ||
    type === 'info' ||
    type === 'primary' ||
    type === 'default'
  ) {
    return type
  }
  return 'default'
}

/**
 * 静态服务管理 Model
 */
export function useStaticServerModel() {
  const moduleId = 'hub0061:static-server'

  /** 加载状态 */
  const loading = ref(false)

  /** 静态服务列表数据 */
  const serverList = ref<TunnelStaticServer[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serverName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'listenAddress',
        label: '监听地址',
        type: 'input',
        placeholder: '请输入监听地址',
        span: 6,
        clearable: true,
      },
      {
        field: 'serverType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...SERVER_TYPE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'serverStatus',
        label: '服务状态',
        type: 'select',
        placeholder: '请选择服务状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...SERVER_STATUS_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增服务',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新增静态服务',
      },
      {
        key: 'start',
        label: '启动',
        icon: 'PlayOutline',
        type: 'success',
        tooltip: '启动选中的服务',
      },
      {
        key: 'stop',
        label: '停止',
        icon: 'StopOutline',
        type: 'warning',
        tooltip: '停止选中的服务',
      },
      {
        key: 'reload',
        label: '重载',
        icon: 'RefreshOutline',
        type: 'info',
        tooltip: '重载选中的服务配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '删除选中的服务',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  /** 获取服务状态显示标签 */
  const getServerStatusLabel = (status: ServerStatus) => {
    const option = SERVER_STATUS_OPTIONS.find((opt) => opt.value === status)
    return option?.label || status
  }

  /** 获取服务状态标签颜色 */
  const getServerStatusTagType = (status: ServerStatus): RsTagVariant => {
    const option = SERVER_STATUS_OPTIONS.find((opt) => opt.value === status)
    return toTagVariant(option?.type)
  }

  /** 获取服务类型显示标签 */
  const getServerTypeLabel = (serverType: ServerType) => {
    const option = SERVER_TYPE_OPTIONS.find((opt) => opt.value === serverType)
    return option?.label || serverType
  }

  /** 获取负载均衡类型显示标签 */
  const getLoadBalanceLabel = (loadBalanceType: LoadBalanceType | null | undefined) => {
    if (!loadBalanceType) return '-'
    const option = LOAD_BALANCE_OPTIONS.find((opt) => opt.value === loadBalanceType)
    return option?.label || loadBalanceType
  }

  /** 表格配置（符合 RsGrid Props 结构） */
  const gridConfig: StaticServerGridConfig = {
    columns: [
      {
        key: 'tunnelStaticServerId',
        title: '服务ID',
        visible: false,
        width: 0,
      },
      {
        key: 'serverName',
        title: '服务名称',
        align: 'center',
        ellipsis: true,
        width: 180,
      },
      {
        key: 'listenAddress',
        title: '监听地址',
        align: 'center',
        width: 160,
        render: (row) =>
          h(
            'span',
            {
              style: {
                color: 'var(--rs-primary)',
                fontFamily: "Consolas, Monaco, 'Courier New', monospace",
              },
            },
            `${row.listenAddress}:${row.listenPort}`,
          ),
      },
      {
        key: 'serverType',
        title: '服务类型',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: row.serverType === 'tcp' ? 'primary' : 'info', size: 'sm' },
            () => getServerTypeLabel(row.serverType),
          ),
      },
      {
        key: 'serverStatus',
        title: '服务状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: getServerStatusTagType(row.serverStatus), size: 'sm' },
            () => getServerStatusLabel(row.serverStatus),
          ),
      },
      {
        key: 'nodeCount',
        title: '节点数',
        align: 'center',
        width: 100,
      },
      {
        key: 'currentConnectionCount',
        title: '当前连接',
        align: 'center',
        width: 100,
      },
      {
        key: 'totalConnectionCount',
        title: '总连接数',
        align: 'center',
        width: 100,
      },
      {
        key: 'totalBytesReceived',
        title: '接收流量',
        align: 'center',
        width: 100,
        formatter: (value) => formatBytes(value as number),
      },
      {
        key: 'totalBytesSent',
        title: '发送流量',
        align: 'center',
        width: 100,
        formatter: (value) => formatBytes(value as number),
      },
      {
        key: 'loadBalanceType',
        title: '负载均衡',
        align: 'center',
        width: 100,
        formatter: (value) => getLoadBalanceLabel(value as LoadBalanceType),
      },
      {
        key: 'maxConnections',
        title: '最大连接',
        align: 'center',
        width: 100,
      },
      {
        key: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 80,
      },
      {
        key: 'serverDescription',
        title: '描述',
        align: 'center',
        ellipsis: true,
        width: 160,
      },
      {
        key: 'startTime',
        title: '启动时间',
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
      {
        key: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: (value) => (value ? formatDate(value as string) : ''),
        width: 160,
      },
    ],
    selectable: true,
    rowKey: 'tunnelStaticServerId',
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
        { key: 'nodes', label: '管理节点', icon: 'settings' },
        { key: 'start', label: '启动', icon: 'play' },
        { key: 'stop', label: '停止', icon: 'square' },
        { key: 'reload', label: '重载配置', icon: 'refresh-cw' },
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
      ],
    },
  }

  /**
   * 设置服务列表
   */
  function setServerList(list: TunnelStaticServer[]) {
    serverList.value = list
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
   * 添加服务到列表
   */
  function addServerToList(server: TunnelStaticServer) {
    serverList.value.push(server)
  }

  /**
   * 更新列表中的服务
   */
  function updateServerInList(
    tunnelStaticServerId: string,
    tenantId: string | undefined,
    updatedServer: Partial<TunnelStaticServer>,
  ) {
    const index = serverList.value.findIndex(
      (s) =>
        s.tunnelStaticServerId === tunnelStaticServerId && (!tenantId || s.tenantId === tenantId),
    )
    if (index !== -1) {
      Object.assign(serverList.value[index], updatedServer)
    }
  }

  /**
   * 从列表中移除服务
   */
  function removeServerFromList(tunnelStaticServerId: string) {
    const index = serverList.value.findIndex((s) => s.tunnelStaticServerId === tunnelStaticServerId)
    if (index >= 0) {
      serverList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除服务
   */
  function removeServersFromList(tunnelStaticServerIds: string[]) {
    serverList.value = serverList.value.filter(
      (s) => !tunnelStaticServerIds.includes(s.tunnelStaticServerId),
    )
  }

  /** 表单页签配置 */
  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'network', label: '网络配置' },
    { key: 'loadbalance', label: '负载均衡' },
    { key: 'tls', label: 'TLS配置', show: false },
    { key: 'other', label: '其他信息' },
  ]

  /** 服务表单配置（用于 RsDataFormModal） */
  const formFields: RsDataFormField[] = [
    {
      field: 'tunnelStaticServerId',
      label: '服务ID',
      type: 'input',
      span: 12,
      primary: true,
      show: false,
    },
    {
      field: 'serverName',
      label: '服务名称',
      type: 'input',
      placeholder: '请输入服务名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      tips: '用于标识此静态代理服务的唯一名称，建议使用有意义的命名便于管理',
      rules: [
        { required: true, message: '请输入服务名称', trigger: ['blur', 'change'] },
        { max: 100, message: '服务名称不能超过100个字符', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'serverType',
      label: '服务类型',
      type: 'select',
      placeholder: '请选择服务类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'tcp',
      tips: 'TCP：适用于大多数场景如 SSH、数据库连接；UDP：适用于 DNS、游戏等场景',
      options: SERVER_TYPE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
      rules: [{ required: true, message: '请选择服务类型', trigger: ['blur', 'change'] }],
    },
    {
      field: 'listenAddress',
      label: '监听地址',
      type: 'input',
      placeholder: '请输入监听地址，如: 0.0.0.0',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: '0.0.0.0',
      tips: '代理服务监听的IP地址。0.0.0.0 表示监听所有网卡，127.0.0.1 仅本机访问',
      rules: [{ required: true, message: '请输入监听地址', trigger: ['blur', 'change'] }],
    },
    {
      field: 'listenPort',
      label: '监听端口',
      type: 'number',
      placeholder: '请输入监听端口',
      span: 12,
      tabKey: 'basic',
      required: true,
      tips: '代理服务监听的端口号（1-65535）。请确保端口未被占用，修改后需要重载配置',
      props: {
        min: 1,
        max: 65535,
        precision: 0,
      },
      rules: [
        { required: true, type: 'number', message: '请输入监听端口', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch',
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      tips: '禁用后服务将不会接受新的连接请求',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'serverDescription',
      label: '服务描述',
      type: 'textarea',
      placeholder: '请输入服务描述',
      span: 24,
      tabKey: 'basic',
      props: {
        rows: 2,
        maxlength: 500,
      },
    },
    {
      field: 'connectionTimeout',
      label: '连接超时(秒)',
      type: 'number',
      placeholder: '请输入连接超时时间',
      span: 12,
      tabKey: 'network',
      defaultValue: 30,
      tips: '建立到后端节点的连接超时时间。超时后将尝试下一个节点或返回错误',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'maxConnections',
      label: '最大连接数',
      type: 'number',
      placeholder: '请输入最大连接数，0表示不限制',
      span: 12,
      tabKey: 'network',
      defaultValue: 0,
      show: false,
      tips: '服务允许的最大并发连接数。0 表示不限制，建议根据服务器性能设置',
      props: {
        min: 0,
        precision: 0,
      },
    },
    {
      field: 'readTimeout',
      label: '读取超时(秒)',
      type: 'number',
      placeholder: '请输入读取超时时间',
      span: 12,
      tabKey: 'network',
      defaultValue: 60,
      show: false,
      tips: '从连接读取数据的超时时间。对于长连接场景（如 SSH）建议设置较大值',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'writeTimeout',
      label: '写入超时(秒)',
      type: 'number',
      placeholder: '请输入写入超时时间',
      span: 12,
      tabKey: 'network',
      defaultValue: 60,
      show: false,
      tips: '向连接写入数据的超时时间。对于长连接场景（如 SSH）建议设置较大值',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'logLevel',
      label: '日志级别',
      type: 'select',
      placeholder: '请选择日志级别',
      span: 12,
      tabKey: 'network',
      defaultValue: 'info',
      show: false,
      tips: '服务的日志记录级别。Debug 记录最详细，Error 只记录错误',
      options: [
        { label: 'Debug', value: 'debug' },
        { label: 'Info', value: 'info' },
        { label: 'Warn', value: 'warn' },
        { label: 'Error', value: 'error' },
      ],
    },
    {
      field: 'loadBalanceType',
      label: '负载均衡类型',
      type: 'select',
      placeholder: '请选择负载均衡类型',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 'roundrobin',
      tips: '轮询：依次选择节点，适合节点性能相近场景；最少连接：选择当前连接最少的节点，适合请求处理时间不均匀场景；随机：随机选择节点',
      options: LOAD_BALANCE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'healthCheckType',
      label: '健康检查类型',
      type: 'select',
      placeholder: '请选择健康检查类型',
      span: 12,
      tabKey: 'loadbalance',
      clearable: true,
      tips: 'TCP：尝试建立 TCP 连接验证节点可达性；HTTP/HTTPS：发送 HTTP 请求检查响应状态码（2xx/3xx 为健康）',
      options: HEALTH_CHECK_TYPE_OPTIONS.map((opt) => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'healthCheckUrl',
      label: '健康检查URL',
      type: 'input',
      placeholder: '请输入健康检查URL，如: /health',
      span: 24,
      tabKey: 'loadbalance',
      tips: 'HTTP/HTTPS 健康检查的请求路径。留空时使用节点地址和端口的根路径。支持完整 URL 或相对路径',
      show: (formData: Record<string, any>) =>
        formData.healthCheckType === 'http' || formData.healthCheckType === 'https',
    },
    {
      field: 'healthCheckInterval',
      label: '检查间隔(秒)',
      type: 'number',
      placeholder: '请输入健康检查间隔',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 30,
      tips: '健康检查的执行间隔。建议不低于 5 秒，避免对后端节点造成过大压力',
      show: (formData: Record<string, any>) => !!formData.healthCheckType,
      props: {
        min: 5,
        precision: 0,
      },
    },
    {
      field: 'healthCheckTimeout',
      label: '检查超时(秒)',
      type: 'number',
      placeholder: '请输入健康检查超时',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 5,
      tips: '单次健康检查的超时时间。超时后节点将被标记为不健康',
      show: (formData: Record<string, any>) => !!formData.healthCheckType,
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'healthCheckMaxFailures',
      label: '最大失败次数',
      type: 'number',
      placeholder: '请输入最大失败次数',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 3,
      show: false,
      tips: '连续健康检查失败达到此次数后，节点将被标记为不健康并从负载均衡中移除',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'tlsEnable',
      label: '启用TLS',
      type: 'switch',
      span: 12,
      tabKey: 'tls',
      defaultValue: 'N',
      show: false,
      tips: '启用后代理服务将使用 TLS 加密传输。需要配置有效的证书和私钥',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'tlsCertFile',
      label: '证书文件路径',
      type: 'input',
      placeholder: '请输入TLS证书文件路径',
      span: 24,
      tabKey: 'tls',
      show: false,
      tips: 'TLS 证书文件的绝对路径，支持 PEM 格式',
    },
    {
      field: 'tlsKeyFile',
      label: '私钥文件路径',
      type: 'input',
      placeholder: '请输入TLS私钥文件路径',
      span: 24,
      tabKey: 'tls',
      show: false,
      tips: 'TLS 私钥文件的绝对路径，支持 PEM 格式',
    },
    {
      field: 'tlsCaFile',
      label: 'CA证书路径',
      type: 'input',
      placeholder: '请输入CA证书文件路径（可选）',
      span: 24,
      tabKey: 'tls',
      show: false,
      tips: 'CA 证书文件路径，用于客户端证书验证（双向 TLS）。留空表示不验证客户端证书',
    },
    {
      field: 'noteText',
      label: '备注信息',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        rows: 3,
        maxlength: 500,
      },
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 12,
      tabKey: 'other',
      disabled: true,
    },
  ]

  return {
    moduleId,
    loading,
    serverList,
    pageInfo,
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,
    getServerStatusLabel,
    getServerStatusTagType,
    getServerTypeLabel,
    getLoadBalanceLabel,
    setServerList,
    setLoading,
    resetPagination,
    updatePagination,
    addServerToList,
    updateServerInList,
    removeServerFromList,
    removeServersFromList,
  }
}

/**
 * 静态服务管理 Model 类型
 */
export type StaticServerModel = ReturnType<typeof useStaticServerModel>
