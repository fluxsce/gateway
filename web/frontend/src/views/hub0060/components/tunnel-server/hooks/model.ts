/**
 * 隧道服务器管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import {
  AddOutline,
  CreateOutline,
  PlayOutline,
  ReloadOutline,
  StopCircleOutline,
  TrashOutline
} from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { h, ref } from 'vue'
import type { TunnelServer } from '../../../types'

/**
 * 隧道服务器管理 Model
 */
export function useTunnelServerModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0060'
  /** 加载状态 */
  const loading = ref(false)

  /** 隧道服务器列表数据 */
  const tunnelServerList = ref<TunnelServer[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serverName',
        label: '服务器名称',
        type: 'input',
        placeholder: '请输入服务器名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serverStatus',
        label: '服务器状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '运行中', value: 'running' },
          { label: '已停止', value: 'stopped' },
          { label: '错误', value: 'error' },
        ],
      },
      {
        field: 'controlAddress',
        label: '控制地址',
        type: 'input',
        placeholder: '请输入控制地址',
        span: 6,
        clearable: true,
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '活动', value: 'Y' },
          { label: '非活动', value: 'N' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新增服务器',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新增隧道服务器',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的服务器',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的服务器',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 数据编辑表单字段配置（供 GdataFormModal 使用） =============
  const formTabs = [
    { key: 'basic', label: '基础配置' },
    { key: 'network', label: '网络配置' },
    { key: 'security', label: '安全配置' },
    { key: 'advanced', label: '高级配置' },
    { key: 'other', label: '其他信息' },
  ]

  const formFields: DataFormField[] = [
    // ============= 基础配置 Tab =============
    {
      field: 'serverName',
      label: '服务器名称',
      type: 'input',
      placeholder: '请输入服务器名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      primary: true,
      props: {
        maxlength: 100,
        showCount: true,
      },
      tips: '隧道路由服务器的唯一标识名称，用于区分不同的服务器实例',
    },
    {
      field: 'serverDescription',
      label: '服务器描述',
      type: 'textarea',
      placeholder: '请输入服务器描述',
      span: 24,
      tabKey: 'basic',
      props: {
        maxlength: 500,
        showCount: true,
        rows: 3,
      },
      tips: '服务器的详细描述信息，用于说明服务器的用途、部署位置等',
    },
    {
      field: 'controlAddress',
      label: '控制地址',
      type: 'input',
      placeholder: '请输入控制地址，如: 0.0.0.0',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: '0.0.0.0',
      tips: '控制端口监听的IP地址。0.0.0.0 表示监听所有网络接口，127.0.0.1 表示仅本地访问',
    },
    {
      field: 'controlPort',
      label: '控制端口',
      type: 'number',
      placeholder: '请输入控制端口',
      span: 12,
      tabKey: 'basic',
      required: true,
      defaultValue: 7000,
      props: {
        min: 1,
        max: 65535,
      },
      tips: '接受客户端连接的控制端口。客户端通过此端口与服务器建立控制连接，用于认证、心跳和服务注册',
    },
    {
      field: 'dashboardPort',
      label: '管理面板端口',
      type: 'number',
      placeholder: '请输入管理面板端口',
      span: 12,
      tabKey: 'basic',
      defaultValue: 7500,
      props: {
        min: 1,
        max: 65535,
      },
      tips: '管理面板的访问端口，用于查看服务器状态、连接统计和性能指标',
    },

    // ============= 网络配置 Tab =============
    {
      field: 'vhostHttpPort',
      label: 'HTTP端口',
      type: 'number',
      placeholder: '虚拟主机HTTP端口',
      span: 12,
      tabKey: 'network',
      defaultValue: 80,
      props: {
        min: 1,
        max: 65535,
      },
      tips: '虚拟主机HTTP服务的监听端口。用于处理基于域名的HTTP隧道请求',
    },
    {
      field: 'vhostHttpsPort',
      label: 'HTTPS端口',
      type: 'number',
      placeholder: '虚拟主机HTTPS端口',
      span: 12,
      tabKey: 'network',
      defaultValue: 443,
      props: {
        min: 1,
        max: 65535,
      },
      tips: '虚拟主机HTTPS服务的监听端口。用于处理基于域名的HTTPS隧道请求，需要配置TLS证书',
    },
    {
      field: 'maxClients',
      label: '最大客户端数',
      type: 'number',
      placeholder: '请输入最大客户端数',
      span: 12,
      tabKey: 'network',
      required: true,
      defaultValue: 1000,
      props: {
        min: 1,
        max: 10000,
      },
      tips: '允许同时连接到服务器的最大客户端数量。超过此限制的新连接将被拒绝',
    },
    {
      field: 'maxPortsPerClient',
      label: '每客户端最大端口数',
      type: 'number',
      placeholder: '请输入每客户端最大端口数',
      span: 12,
      tabKey: 'network',
      defaultValue: 10,
      props: {
        min: 1,
        max: 1000,
      },
      tips: '单个客户端可以使用的最大端口数量。用于限制每个客户端可注册的服务数量',
    },
    {
      field: 'allowPorts',
      label: '允许的端口范围',
      type: 'input',
      placeholder: '例如: 10000-20000',
      span: 24,
      tabKey: 'network',
      tips: '格式: 10000-20000 或 10000,20000,30000',
    },

    // ============= 安全配置 Tab =============
    {
      field: 'tokenAuth',
      label: 'Token认证',
      type: 'select',
      span: 12,
      tabKey: 'security',
      defaultValue: 'Y',
      options: [
        { label: '启用', value: 'Y' },
        { label: '禁用', value: 'N' },
      ],
      tips: '启用Token认证后，客户端连接时必须提供正确的认证令牌，提高安全性',
    },
    {
      field: 'authToken',
      label: '认证Token',
      type: 'input',
      placeholder: '请输入认证Token',
      span: 12,
      tabKey: 'security',
      show: (formData: any) => formData?.tokenAuth === 'Y',
      props: {
        type: 'password',
        showPasswordOn: 'click',
      },
      tips: '客户端连接时使用的认证令牌。如果为空，系统将自动生成一个32位随机字符串',
    },
    {
      field: 'tlsEnable',
      label: 'TLS加密',
      type: 'select',
      span: 12,
      tabKey: 'security',
      defaultValue: 'N',
      options: [
        { label: '启用', value: 'Y' },
        { label: '禁用', value: 'N' },
      ],
      tips: '启用TLS加密后，客户端与服务器之间的通信将使用TLS协议加密，提高数据传输安全性',
    },
    {
      field: 'tlsCertFile',
      label: 'TLS证书文件',
      type: 'input',
      placeholder: '请输入TLS证书文件路径',
      span: 12,
      tabKey: 'security',
      show: (formData: any) => formData?.tlsEnable === 'Y',
      tips: 'TLS证书文件的完整路径。证书文件必须是服务器可访问的，通常为 .crt 或 .pem 格式',
    },
    {
      field: 'tlsKeyFile',
      label: 'TLS密钥文件',
      type: 'input',
      placeholder: '请输入TLS密钥文件路径',
      span: 12,
      tabKey: 'security',
      show: (formData: any) => formData?.tlsEnable === 'Y',
      tips: 'TLS私钥文件的完整路径。密钥文件必须是服务器可访问的，通常为 .key 格式，需要与证书文件匹配',
    },

    // ============= 高级配置 Tab =============
    {
      field: 'heartbeatInterval',
      label: '心跳间隔(秒)',
      type: 'number',
      placeholder: '请输入心跳间隔',
      span: 12,
      tabKey: 'advanced',
      required: true,
      defaultValue: 30,
      props: {
        min: 5,
        max: 300,
      },
      tips: '服务器与客户端之间的心跳检测间隔时间。客户端会定期发送心跳消息以保持连接活跃',
    },
    {
      field: 'heartbeatTimeout',
      label: '心跳超时(秒)',
      type: 'number',
      placeholder: '请输入心跳超时',
      span: 12,
      tabKey: 'advanced',
      required: true,
      defaultValue: 90,
      props: {
        min: 10,
        max: 600,
      },
      tips: '心跳检测的超时时间。如果在此时间内未收到客户端心跳，服务器将认为客户端已断开连接',
    },
    {
      field: 'logLevel',
      label: '日志级别',
      type: 'select',
      placeholder: '请选择日志级别',
      span: 12,
      tabKey: 'advanced',
      defaultValue: 'info',
      options: [
        { label: 'Debug', value: 'debug' },
        { label: 'Info', value: 'info' },
        { label: 'Warn', value: 'warn' },
        { label: 'Error', value: 'error' },
      ],
      tips: '服务器日志记录的级别。Debug包含最详细的调试信息，Error仅记录错误信息',
    },

    // ============= 其他信息 Tab =============
    {
      field: 'noteText',
      label: '备注信息',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'other',
      props: {
        maxlength: 500,
        showCount: true,
        rows: 4,
      },
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 12,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 12,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 12,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 12,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'activeFlag',
      label: '活动状态',
      type: 'select',
      span: 12,
      tabKey: 'other',
      defaultValue: 'Y',
      options: [
        { label: '活动', value: 'Y' },
        { label: '非活动', value: 'N' },
      ],
    },
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'serverName',
        title: '服务器名称',
        width: 180,
        sortable: true,
        showOverflow: true,
        slots: { default: 'serverName' },
      },
      {
        field: 'controlAddress',
        title: '控制地址',
        width: 140,
        showOverflow: true,
        formatter: ({ row }: any) => `${row.controlAddress}:${row.controlPort}`,
      },
      {
        field: 'dashboardPort',
        title: '管理端口',
        width: 90,
        align: 'center',
        formatter: ({ cellValue }) => cellValue || '-',
      },
      {
        field: 'serverStatus',
        title: '状态',
        width: 90,
        align: 'center',
        slots: { default: 'serverStatus' },
      },
      {
        field: 'vhostHttpPort',
        title: 'HTTP端口',
        width: 90,
        align: 'center',
        formatter: ({ cellValue }) => cellValue || '-',
      },
      {
        field: 'vhostHttpsPort',
        title: 'HTTPS端口',
        width: 100,
        align: 'center',
        formatter: ({ cellValue }) => cellValue || '-',
      },
      {
        field: 'maxClients',
        title: '最大客户端',
        width: 100,
        align: 'center',
        slots: { default: 'maxClients' },
      },
      {
        field: 'tokenAuth',
        title: 'Token认证',
        width: 90,
        align: 'center',
        slots: { default: 'tokenAuth' },
      },
      {
        field: 'tlsEnable',
        title: 'TLS',
        width: 70,
        align: 'center',
        slots: { default: 'tlsEnable' },
      },
      {
        field: 'heartbeatInterval',
        title: '心跳间隔(秒)',
        width: 110,
        align: 'center',
        formatter: ({ cellValue }) => cellValue ? `${cellValue}s` : '-',
      },
      {
        field: 'heartbeatTimeout',
        title: '心跳超时(秒)',
        width: 110,
        align: 'center',
        formatter: ({ cellValue }) => cellValue ? `${cellValue}s` : '-',
      },
      {
        field: 'startTime',
        title: '启动时间',
        width: 140,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm') : '-',
      },
      {
        field: 'addTime',
        title: '创建时间',
        width: 140,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm') : '',
      },
      {
        field: 'addWho',
        title: '创建人',
        width: 100,
        showOverflow: true,
      },
      {
        field: 'activeFlag',
        title: '活动状态',
        width: 100,
        align: 'center',
        formatter: ({ row }: any) => (row.activeFlag === 'Y' ? '活动' : '非活动'),
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
        {
          code: 'start',
          name: '启动服务器',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(PlayOutline) }),
        },
        {
          code: 'stop',
          name: '停止服务器',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(StopCircleOutline) }),
        },
        {
          code: 'restart',
          name: '重启服务器',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(ReloadOutline) }),
        },
      ],
    },
    height: '100%',
  }

  // ============= 辅助方法 =============

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
   * 设置隧道服务器列表
   */
  const setTunnelServerList = (list: TunnelServer[]) => {
    tunnelServerList.value = list
  }

  /**
   * 清空隧道服务器列表
   */
  const clearTunnelServerList = () => {
    tunnelServerList.value = []
  }

  /**
   * 添加隧道服务器到列表
   */
  const addTunnelServerToList = (server: TunnelServer) => {
    tunnelServerList.value.unshift(server)
  }

  /**
   * 更新列表中的隧道服务器
   */
  const updateTunnelServerInList = (
    tunnelServerId: string,
    tenantId: string,
    updatedServer: Partial<TunnelServer>
  ) => {
    const index = tunnelServerList.value.findIndex(
      (s) => s.tunnelServerId === tunnelServerId && s.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(tunnelServerList.value[index], updatedServer)
    }
  }

  /**
   * 从列表中删除隧道服务器
   */
  const removeTunnelServerFromList = (tunnelServerId: string, tenantId: string) => {
    const index = tunnelServerList.value.findIndex(
      (s) => s.tunnelServerId === tunnelServerId && s.tenantId === tenantId
    )
    if (index !== -1) {
      tunnelServerList.value.splice(index, 1)
    }
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    tunnelServerList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setTunnelServerList,
    clearTunnelServerList,
    addTunnelServerToList,
    updateTunnelServerInList,
    removeTunnelServerFromList,
  }
}

/**
 * Model 返回类型
 */
export type TunnelServerModel = ReturnType<typeof useTunnelServerModel>

