/**
 * 静态节点列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatBytes, formatDate } from '@/utils/format'
import { AddOutline, TrashOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import type { HealthCheckStatus, NodeStatus, ProxyType, TunnelStaticNode } from './types'
import {
  HEALTH_CHECK_STATUS_OPTIONS,
  NODE_STATUS_OPTIONS,
  PROXY_TYPE_OPTIONS
} from './types'

/**
 * 静态节点列表 Model
 */
export function useStaticNodeModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0061:static-nodes'
  
  /** 加载状态 */
  const loading = ref(false)

  /** 静态节点列表数据 */
  const nodeList = ref<TunnelStaticNode[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'nodeName',
        label: '节点名称',
        type: 'input',
        placeholder: '请输入节点名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'targetAddress',
        label: '目标地址',
        type: 'input',
        placeholder: '请输入目标地址',
        span: 6,
        clearable: true,
      },
      {
        field: 'nodeStatus',
        label: '节点状态',
        type: 'select',
        placeholder: '请选择节点状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...NODE_STATUS_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
      {
        field: 'healthCheckStatus',
        label: '健康状态',
        type: 'select',
        placeholder: '请选择健康状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...HEALTH_CHECK_STATUS_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增节点',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增静态节点',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '批量删除选中的节点',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 获取节点状态显示标签 */
  const getNodeStatusLabel = (nodeStatus: NodeStatus) => {
    const option = NODE_STATUS_OPTIONS.find(opt => opt.value === nodeStatus)
    return option?.label || nodeStatus
  }

  /** 获取节点状态标签颜色 */
  const getNodeStatusTagType = (nodeStatus: NodeStatus): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    const option = NODE_STATUS_OPTIONS.find(opt => opt.value === nodeStatus)
    return option?.type || 'default'
  }

  /** 获取代理类型显示标签 */
  const getProxyTypeLabel = (proxyType: ProxyType) => {
    const option = PROXY_TYPE_OPTIONS.find(opt => opt.value === proxyType)
    return option?.label || proxyType
  }

  /** 获取健康检查状态显示标签 */
  const getHealthCheckStatusLabel = (status: HealthCheckStatus | null | undefined) => {
    if (!status) return '未知'
    const option = HEALTH_CHECK_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.label || status
  }

  /** 获取健康检查状态标签颜色 */
  const getHealthCheckStatusTagType = (status: HealthCheckStatus | null | undefined): "default" | "success" | "error" | "warning" | "primary" | "info" => {
    if (!status) return 'default'
    const option = HEALTH_CHECK_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.type || 'default'
  }

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'tunnelStaticNodeId',
        title: '节点ID',
        visible: false,
        width: 0,
      },
      {
        field: 'nodeName',
        title: '节点名称',
        align: 'center',
        showOverflow: 'tooltip',
        width: 160,
      },
      {
        field: 'targetAddress',
        title: '目标地址',
        align: 'center',
        showOverflow: 'tooltip',
        width: 140,
      },
      {
        field: 'targetPort',
        title: '目标端口',
        align: 'center',
        width: 100,
      },
      {
        field: 'proxyType',
        title: '代理类型',
        align: 'center',
        width: 100,
        slots: { default: 'proxyType' },
      },
      {
        field: 'nodeStatus',
        title: '节点状态',
        align: 'center',
        width: 100,
        slots: { default: 'nodeStatus' },
      },
      {
        field: 'healthCheckStatus',
        title: '健康状态',
        align: 'center',
        width: 100,
        slots: { default: 'healthCheckStatus' },
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
        field: 'failureCount',
        title: '失败次数',
        align: 'center',
        width: 100,
      },
      {
        field: 'lastHealthCheck',
        title: '最后检查',
        align: 'center',
        formatter: ({ row }) => formatDate(row.lastHealthCheck),
        width: 160,
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 80,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'nodeDescription',
        title: '描述',
        align: 'center',
        showOverflow: 'tooltip',
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
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete',
        },
      ],
    },
  }

  // ============= 状态更新方法 =============

  /**
   * 设置节点列表
   */
  function setNodeList(list: TunnelStaticNode[]) {
    nodeList.value = list
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
   * 添加节点到列表
   */
  function addNodeToList(node: TunnelStaticNode) {
    nodeList.value.push(node)
  }

  /**
   * 更新列表中的节点
   * @param tunnelStaticNodeId 节点ID
   * @param tenantId 租户ID（可选，用于精确匹配）
   * @param updatedNode 更新的节点数据
   */
  function updateNodeInList(tunnelStaticNodeId: string, tenantId: string | undefined, updatedNode: Partial<TunnelStaticNode>) {
    const index = nodeList.value.findIndex(
      (n) => n.tunnelStaticNodeId === tunnelStaticNodeId && (!tenantId || n.tenantId === tenantId)
    )
    if (index !== -1) {
      // 使用 Object.assign 合并更新，保持响应式
      Object.assign(nodeList.value[index], updatedNode)
    }
  }

  /**
   * 从列表中移除节点
   */
  function removeNodeFromList(tunnelStaticNodeId: string) {
    const index = nodeList.value.findIndex((n) => n.tunnelStaticNodeId === tunnelStaticNodeId)
    if (index >= 0) {
      nodeList.value.splice(index, 1)
    }
  }

  /**
   * 从列表中批量移除节点
   */
  function removeNodesFromList(tunnelStaticNodeIds: string[]) {
    nodeList.value = nodeList.value.filter((n) => !tunnelStaticNodeIds.includes(n.tunnelStaticNodeId))
  }

  // ============= 表单配置 =============

  /** 表单页签配置 */
  const formTabs = [
    {
      key: 'basic',
      label: '基本信息',
    },
    {
      key: 'advanced',
      label: '高级配置',
      show: false, // 后端暂未实现节点级别的高级配置
    },
    {
      key: 'other',
      label: '其他信息',
    },
  ]

  /** 节点表单配置（用于 GdataFormModal） */
  const formFields: DataFormField[] = [
    // 主键字段（隐藏，但必须存在用于编辑）
    {
      field: 'tunnelStaticNodeId',
      label: '节点ID',
      type: 'input' as const,
      span: 12,
      primary: true,
      show: false,
    },
    {
      field: 'tunnelStaticServerId',
      label: '服务器ID',
      type: 'input' as const,
      span: 12,
      show: false,
    },
    // 基本信息
    {
      field: 'nodeName',
      label: '节点名称',
      type: 'input' as const,
      placeholder: '请输入节点名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      tips: '用于标识此后端节点的唯一名称，建议使用有意义的命名便于管理',
      rules: [
        { required: true, message: '请输入节点名称', trigger: ['blur', 'input'] },
        { max: 100, message: '节点名称不能超过100个字符', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'targetAddress',
      label: '目标地址',
      type: 'input' as const,
      placeholder: '请输入后端服务地址（IP或域名）',
      span: 12,
      tabKey: 'basic',
      required: true,
      tips: '后端服务的 IP 地址或域名。支持内网地址如 192.168.1.100 或域名如 backend.local',
      rules: [
        { required: true, message: '请输入目标地址', trigger: ['blur', 'input'] },
      ],
    },
    {
      field: 'targetPort',
      label: '目标端口',
      type: 'number' as const,
      placeholder: '请输入后端服务端口',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 80,
      tips: '后端服务监听的端口号（1-65535）。如 SSH 服务通常是 22，HTTP 是 80',
      props: {
        min: 1,
        max: 65535,
        precision: 0,
      },
      rules: [
        { 
          required: true, 
          type: 'number',
          message: '请输入目标端口', 
          trigger: ['blur', 'change'],
        },
      ],
    },
    {
      field: 'proxyType',
      label: '代理类型',
      type: 'select' as const,
      placeholder: '请选择代理类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 'tcp',
      tips: 'TCP：适用于大多数场景如 SSH、数据库连接；UDP：适用于 DNS、游戏等场景。需与服务器类型一致',
      options: PROXY_TYPE_OPTIONS.map(opt => ({ label: opt.label, value: opt.value })),
      rules: [
        { required: true, message: '请选择代理类型', trigger: ['blur', 'change'] },
      ],
    },
    {
      field: 'activeFlag',
      label: '启用状态',
      type: 'switch' as const,
      span: 12,
      tabKey: 'basic',
      defaultValue: 'Y',
      tips: '禁用后此节点将不会被负载均衡选中，不会接收新的连接请求',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'nodeDescription',
      label: '节点描述',
      type: 'input' as const,
      placeholder: '请输入节点描述',
      span: 24,
      tabKey: 'basic',
      props: {
        type: 'textarea',
        rows: 2,
        maxlength: 500,
        showCount: true,
      },
    },
    // 高级配置（后端暂未实现节点级别的高级配置，使用服务器级别配置）
    {
      field: 'maxConnections',
      label: '最大连接数',
      type: 'number' as const,
      placeholder: '请输入最大连接数，0表示不限制',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 0,
      show: false, // 后端使用服务器级别的连接数限制
      tips: '此节点允许的最大并发连接数。0 表示不限制',
      props: {
        min: 0,
        precision: 0,
      },
    },
    {
      field: 'connectionTimeout',
      label: '连接超时(秒)',
      type: 'number' as const,
      placeholder: '请输入连接超时时间',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 30,
      show: false, // 后端使用服务器的 ConnectionTimeout
      tips: '连接到此节点的超时时间',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'readTimeout',
      label: '读取超时(秒)',
      type: 'number' as const,
      placeholder: '请输入读取超时时间',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 60,
      show: false, // 后端使用 io.Copy 无超时控制
      tips: '从此节点读取数据的超时时间',
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
      tabKey: 'advanced',
      defaultValue: 60,
      show: false, // 后端使用 io.Copy 无超时控制
      tips: '向此节点写入数据的超时时间',
      props: {
        min: 1,
        precision: 0,
      },
    },
    {
      field: 'retryCount',
      label: '重试次数',
      type: 'number' as const,
      placeholder: '请输入重试次数',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 3,
      show: false, // 后端暂未实现节点级别的重试逻辑
      tips: '连接失败时的重试次数',
      props: {
        min: 0,
        precision: 0,
      },
    },
    {
      field: 'retryInterval',
      label: '重试间隔(秒)',
      type: 'number' as const,
      placeholder: '请输入重试间隔',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 1,
      show: false, // 后端暂未实现节点级别的重试逻辑
      tips: '重试之间的等待时间',
      props: {
        min: 0,
        precision: 0,
      },
    },
    {
      field: 'compression',
      label: '启用压缩',
      type: 'switch' as const,
      span: 12,
      tabKey: 'advanced',
      defaultValue: 'N',
      show: false, // 后端暂未实现压缩功能
      tips: '启用后将压缩传输数据，可减少带宽占用但增加 CPU 开销',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'encryption',
      label: '启用加密',
      type: 'switch' as const,
      span: 12,
      tabKey: 'advanced',
      defaultValue: 'N',
      show: false, // 后端暂未实现加密功能
      tips: '启用后将加密传输数据，提高安全性',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
    {
      field: 'secretKey',
      label: '加密密钥',
      type: 'input' as const,
      placeholder: '请输入加密密钥',
      span: 24,
      tabKey: 'advanced',
      show: false, // 后端暂未实现加密功能
      tips: '用于数据加密的密钥',
      props: {
        type: 'password',
        showPasswordOn: 'click',
      },
    },
    // 其他信息
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
    nodeList,
    pageInfo,

    // 配置
    searchFormConfig,
    gridConfig,
    formFields,
    formTabs,

    // 工具函数
    getNodeStatusLabel,
    getNodeStatusTagType,
    getProxyTypeLabel,
    getHealthCheckStatusLabel,
    getHealthCheckStatusTagType,

    // 方法
    setNodeList,
    setLoading,
    resetPagination,
    updatePagination,
    addNodeToList,
    updateNodeInList,
    removeNodeFromList,
    removeNodesFromList,
  }
}

/**
 * 静态节点列表 Model 类型
 */
export type StaticNodeModel = ReturnType<typeof useStaticNodeModel>

