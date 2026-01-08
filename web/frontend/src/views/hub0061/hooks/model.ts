/**
 * 静态服务管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatBytes, formatDate } from '@/utils/format'
import {
    AddOutline,
    PlayOutline,
    RefreshOutline,
    StopOutline,
    TrashOutline
} from '@vicons/ionicons5'
import { ref } from 'vue'
import type {
    HealthCheckType,
    LoadBalanceType,
    ServerStatus,
    ServerType,
    TunnelStaticServer
} from '../types'

// ============= 常量定义 =============

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
 * 静态服务管理 Model
 */
export function useStaticServerModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0061:static-server'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 静态服务列表数据 */
  const serverList = ref<TunnelStaticServer[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
          ...SERVER_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
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
          ...SERVER_STATUS_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增服务',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增静态服务',
      },
      {
        key: 'start',
        label: '启动',
        icon: PlayOutline,
        type: 'success',
        tooltip: '启动选中的服务',
      },
      {
        key: 'stop',
        label: '停止',
        icon: StopOutline,
        type: 'warning',
        tooltip: '停止选中的服务',
      },
      {
        key: 'reload',
        label: '重载',
        icon: RefreshOutline,
        type: 'info',
        tooltip: '重载选中的服务配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的服务',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 获取服务状态显示标签 */
  const getServerStatusLabel = (status: ServerStatus) => {
    const option = SERVER_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.label || status
  }

  /** 获取服务状态标签颜色 */
  const getServerStatusTagType = (status: ServerStatus): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const option = SERVER_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.type || 'default'
  }

  /** 获取服务类型显示标签 */
  const getServerTypeLabel = (serverType: ServerType) => {
    const option = SERVER_TYPE_OPTIONS.find(opt => opt.value === serverType)
    return option?.label || serverType
  }

  /** 获取负载均衡类型显示标签 */
  const getLoadBalanceLabel = (loadBalanceType: LoadBalanceType | null | undefined) => {
    if (!loadBalanceType) return '-'
    const option = LOAD_BALANCE_OPTIONS.find(opt => opt.value === loadBalanceType)
    return option?.label || loadBalanceType
  }

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'tunnelStaticServerId',
        title: '服务ID',
        visible: false,
        width: 0,
      },
      {
        field: 'serverName',
        title: '服务名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 180,
      },
      {
        field: 'listenAddress',
        title: '监听地址',
        align: 'center',
        width: 140,
        slots: { default: 'listenAddress' },
      },
      {
        field: 'serverType',
        title: '服务类型',
        align: 'center',
        width: 100,
        slots: { default: 'serverType' },
      },
      {
        field: 'serverStatus',
        title: '服务状态',
        align: 'center',
        width: 100,
        slots: { default: 'serverStatus' },
      },
      {
        field: 'nodeCount',
        title: '节点数',
        align: 'center',
        width: 80,
        slots: { default: 'nodeCount' },
      },
      {
        field: 'currentConnectionCount',
        title: '当前连接',
        align: 'center',
        width: 100,
      },
      {
        field: 'totalConnectionCount',
        title: '总连接数',
        align: 'center',
        width: 100,
      },
      {
        field: 'totalBytesReceived',
        title: '接收流量',
        align: 'center',
        width: 100,
        formatter: ({ row }) => formatBytes(row.totalBytesReceived),
      },
      {
        field: 'totalBytesSent',
        title: '发送流量',
        align: 'center',
        width: 100,
        formatter: ({ row }) => formatBytes(row.totalBytesSent),
      },
      {
        field: 'loadBalanceType',
        title: '负载均衡',
        align: 'center',
        width: 100,
        formatter: ({ row }) => getLoadBalanceLabel(row.loadBalanceType),
      },
      {
        field: 'maxConnections',
        title: '最大连接',
        align: 'center',
        width: 100,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 80,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'serverDescription',
        title: '描述',
        align: 'center',
        showOverflow: 'tooltip',
        width: 160,
      },
      {
        field: 'startTime',
        title: '启动时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.startTime),
        width: 160,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        formatter: ({ row }) => formatDate(row.addTime),
        width: 160,
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
          code: 'nodes',
          name: '管理节点',
          prefixIcon: 'vxe-icon-setting',
        },
        {
          code: 'start',
          name: '启动',
          prefixIcon: 'vxe-icon-caret-right',
        },
        {
          code: 'stop',
          name: '停止',
          prefixIcon: 'vxe-icon-square',
        },
        {
          code: 'reload',
          name: '重载配置',
          prefixIcon: 'vxe-icon-refresh',
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
  function updateServerInList(tunnelStaticServerId: string, tenantId: string | undefined, updatedServer: Partial<TunnelStaticServer>) {
    const index = serverList.value.findIndex(
      (s) => s.tunnelStaticServerId === tunnelStaticServerId && (!tenantId || s.tenantId === tenantId)
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
    serverList.value = serverList.value.filter((s) => !tunnelStaticServerIds.includes(s.tunnelStaticServerId))
  }

  // ============= 表单配置 =============

  /** 表单页签配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'network',
      label: '网络配置',
    },
    {
      key: 'loadbalance',
      label: '负载均衡',
    },
    {
      key: 'tls',
      label: 'TLS配置',
      show: false, // 后端暂未实现 TLS 功能
    },
    {
      key: 'other',
      label: '其他信息',
    },
  ]

  /** 服务表单配置（用于 GdataFormModal） */
  const formFields: DataFormField[] = [
    // 主键字段（隐藏，但必须存在用于编辑）
    {
      field: 'tunnelStaticServerId',
      label: '服务ID',
      type: 'input' as const,
      span: 12,
      primary: true,
      show: false,
    },
    // ==================== 基本信息 ====================
    {
      field: 'serverName',
      label: '服务名称',
      type: 'input' as const,
      placeholder: '请输入服务名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      tips: '用于标识此静态代理服务的唯一名称，建议使用有意义的命名便于管理',
      rules: [
        { required: true, message: '请输入服务名称', trigger: ['blur', 'input'] },
        { max: 100, message: '服务名称不能超过100个字符', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'serverType',
      label: '服务类型',
      type: 'select' as const,
      placeholder: '请选择服务类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'tcp',
      tips: 'TCP：适用于大多数场景如 SSH、数据库连接；UDP：适用于 DNS、游戏等场景',
      options: SERVER_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择服务类型', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'listenAddress',
      label: '监听地址',
      type: 'input' as const,
      placeholder: '请输入监听地址，如: 0.0.0.0',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: '0.0.0.0',
      tips: '代理服务监听的IP地址。0.0.0.0 表示监听所有网卡，127.0.0.1 仅本机访问',
      rules: [
        { required: true, message: '请输入监听地址', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'listenPort',
      label: '监听端口',
      type: 'number' as const,
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
      type: 'switch' as const,
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
      type: 'input' as const,
      placeholder: '请输入服务描述',
      span: 24,
      tabKey: 'basic',
      props: {
        type: 'textarea',
        rows: 2,
        maxlength: 500,
        showCount: true,
      },
    },
    // ==================== 网络配置 ====================
    {
      field: 'connectionTimeout',
      label: '连接超时(秒)',
      type: 'number' as const,
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
    // 以下字段暂未在后端使用，使用 show: false 隐藏
    {
      field: 'maxConnections',
      label: '最大连接数',
      type: 'number' as const,
      placeholder: '请输入最大连接数，0表示不限制',
      span: 12,
      tabKey: 'network',
      defaultValue: 0,
      show: false, // 后端暂未实现连接数限制
      tips: '服务允许的最大并发连接数。0 表示不限制，建议根据服务器性能设置',
      props: {
        min: 0,
        precision: 0,
      },
    },
    {
      field: 'readTimeout',
      label: '读取超时(秒)',
      type: 'number' as const,
      placeholder: '请输入读取超时时间',
      span: 12,
      tabKey: 'network',
      defaultValue: 60,
      show: false, // 后端使用 io.Copy 无超时控制
      tips: '从连接读取数据的超时时间。对于长连接场景（如 SSH）建议设置较大值',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'writeTimeout',
      label: '写入超时(秒)',
      type: 'number' as const,
      placeholder: '请输入写入超时时间',
      span: 12,
      tabKey: 'network',
      defaultValue: 60,
      show: false, // 后端使用 io.Copy 无超时控制
      tips: '向连接写入数据的超时时间。对于长连接场景（如 SSH）建议设置较大值',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'logLevel',
      label: '日志级别',
      type: 'select' as const,
      placeholder: '请选择日志级别',
      span: 12,
      tabKey: 'network',
      defaultValue: 'info',
      show: false, // 后端使用全局日志配置
      tips: '服务的日志记录级别。Debug 记录最详细，Error 只记录错误',
      options: [
        { label: 'Debug', value: 'debug' },
        { label: 'Info', value: 'info' },
        { label: 'Warn', value: 'warn' },
        { label: 'Error', value: 'error' },
      ],
    },
    // ==================== 负载均衡配置 ====================
    {
      field: 'loadBalanceType',
      label: '负载均衡类型',
      type: 'select' as const,
      placeholder: '请选择负载均衡类型',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 'roundrobin',
      tips: '轮询：依次选择节点，适合节点性能相近场景；最少连接：选择当前连接最少的节点，适合请求处理时间不均匀场景；随机：随机选择节点',
      options: LOAD_BALANCE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'healthCheckType',
      label: '健康检查类型',
      type: 'select' as const,
      placeholder: '请选择健康检查类型',
      span: 12,
      tabKey: 'loadbalance',
      clearable: true,
      tips: 'TCP：尝试建立 TCP 连接验证节点可达性；HTTP/HTTPS：发送 HTTP 请求检查响应状态码（2xx/3xx 为健康）',
      options: HEALTH_CHECK_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
    },
    {
      field: 'healthCheckUrl',
      label: '健康检查URL',
      type: 'input' as const,
      placeholder: '请输入健康检查URL，如: /health',
      span: 24,
      tabKey: 'loadbalance',
      tips: 'HTTP/HTTPS 健康检查的请求路径。留空时使用节点地址和端口的根路径。支持完整 URL 或相对路径',
      show: (formData: Record<string, any>) => formData.healthCheckType === 'http' || formData.healthCheckType === 'https',
    },
    {
      field: 'healthCheckInterval',
      label: '检查间隔(秒)',
      type: 'number' as const,
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
      type: 'number' as const,
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
      type: 'number' as const,
      placeholder: '请输入最大失败次数',
      span: 12,
      tabKey: 'loadbalance',
      defaultValue: 3,
      show: false, // 后端单次失败即标记不健康，暂未实现累计失败次数
      tips: '连续健康检查失败达到此次数后，节点将被标记为不健康并从负载均衡中移除',
      props: {
        min: 1,
        precision: 0,
      },
    },
    // ==================== TLS配置（暂未实现，使用 show: false 隐藏） ====================
    {
      field: 'tlsEnable',
      label: '启用TLS',
      type: 'switch' as const,
      span: 12,
      tabKey: 'tls',
      defaultValue: 'N',
      show: false, // 后端暂未实现 TLS 功能
      tips: '启用后代理服务将使用 TLS 加密传输。需要配置有效的证书和私钥',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'tlsCertFile',
      label: '证书文件路径',
      type: 'input' as const,
      placeholder: '请输入TLS证书文件路径',
      span: 24,
      tabKey: 'tls',
      show: false, // 后端暂未实现 TLS 功能
      tips: 'TLS 证书文件的绝对路径，支持 PEM 格式',
    },
    {
      field: 'tlsKeyFile',
      label: '私钥文件路径',
      type: 'input' as const,
      placeholder: '请输入TLS私钥文件路径',
      span: 24,
      tabKey: 'tls',
      show: false, // 后端暂未实现 TLS 功能
      tips: 'TLS 私钥文件的绝对路径，支持 PEM 格式',
    },
    {
      field: 'tlsCaFile',
      label: 'CA证书路径',
      type: 'input' as const,
      placeholder: '请输入CA证书文件路径（可选）',
      span: 24,
      tabKey: 'tls',
      show: false, // 后端暂未实现 TLS 功能
      tips: 'CA 证书文件路径，用于客户端证书验证（双向 TLS）。留空表示不验证客户端证书',
    },
    // ==================== 其他信息 ====================
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
    serverList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 工具函数
    getServerStatusLabel,
    getServerStatusTagType,
    getServerTypeLabel,
    getLoadBalanceLabel,

    // 方法
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
