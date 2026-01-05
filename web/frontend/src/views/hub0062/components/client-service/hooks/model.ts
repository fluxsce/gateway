/**
 * 隧道服务管理 - 数据模型定义
 * 定义服务表格配置、表单配置、选项等
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid/types'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { AddOutline, CloudUploadOutline, EyeOutline, StopOutline } from '@vicons/ionicons5'
import { NIcon } from 'naive-ui'
import { h, ref } from 'vue'
import type { TunnelService } from '../../../types'
import { TunnelClientSelector } from '../../tunnel-client-grid'

// 服务类型选项
export const SERVICE_TYPE_OPTIONS = [
  { label: 'TCP', value: 'tcp' },
  { label: 'UDP', value: 'udp' },
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
  { label: 'STCP', value: 'stcp' },
  { label: 'SUDP', value: 'sudp' },
  { label: 'XTCP', value: 'xtcp' }
]

// 服务状态选项
export const SERVICE_STATUS_OPTIONS = [
  { label: '活动', value: 'active' },
  { label: '不活动', value: 'inactive' },
  { label: '错误', value: 'error' },
  { label: '离线', value: 'offline' }
]

// 激活标识选项
export const ACTIVE_FLAG_OPTIONS = [
  { label: '启用', value: 'Y' },
  { label: '禁用', value: 'N' }
]

// 是否选项
export const YES_NO_OPTIONS = [
  { label: '是', value: 'Y' },
  { label: '否', value: 'N' }
]

export function useTunnelServiceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0062:service'
  /** 加载状态 */
  const loading = ref(false)

  /** 服务列表数据 */
  const serviceList = ref<TunnelService[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // 搜索表单配置
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        props: {
          placeholder: '请输入服务名称'
        }
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        props: {
          placeholder: '请选择服务类型',
          options: SERVICE_TYPE_OPTIONS
        }
      },
      {
        field: 'serviceStatus',
        label: '服务状态',
        type: 'select',
        props: {
          placeholder: '请选择服务状态',
          options: SERVICE_STATUS_OPTIONS
        }
      },
      {
        field: 'keyword',
        label: '关键词',
        type: 'input',
        props: {
          placeholder: '服务名称/本地地址/子域名'
        }
      }
    ],
    toolbarButtons: [
      { key: 'create', label: '新增服务', type: 'primary', icon: AddOutline }
    ]
  }

  // 表格配置
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      { field: 'tunnelServiceId', title: '服务ID', width: 200, showOverflow: 'tooltip' },
      { field: 'serviceName', title: '服务名称', width: 150, showOverflow: 'tooltip', sortable: true },
      { field: 'tunnelClientId', title: '客户端ID', width: 200, showOverflow: 'tooltip' },
      { field: 'serviceType', title: '服务类型', width: 100, slots: { default: 'serviceType' } },
      { 
        field: 'localAddress', 
        title: '本地地址', 
        width: 180,
        showOverflow: 'tooltip',
        formatter: ({ row }) => `${row.localAddress}:${row.localPort}`
      },
      { field: 'remotePort', title: '远程端口', width: 100 },
      { field: 'subDomain', title: '子域名', width: 150, showOverflow: 'tooltip' },
      { field: 'serviceStatus', title: '服务状态', width: 100, slots: { default: 'serviceStatus' } },
      { 
        field: 'connectionCount', 
        title: '当前连接', 
        width: 100,
        align: 'right'
      },
      { 
        field: 'totalConnections', 
        title: '总连接数', 
        width: 100,
        align: 'right'
      },
      { 
        field: 'registeredTime', 
        title: '注册时间', 
        width: 160,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) => formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss')
      },
      { 
        field: 'lastActiveTime', 
        title: '最后活动', 
        width: 160,
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) => formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss')
      },
      { field: 'activeFlag', title: '状态', width: 80, slots: { default: 'activeFlag' } }
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
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(EyeOutline) })
        },
        {
          code: 'register',
          name: '注册服务',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(CloudUploadOutline) })
        },
        {
          code: 'unregister',
          name: '注销服务',
          prefixIcon: () => h(NIcon, { size: 14 }, { default: () => h(StopOutline) })
        },
        {
          code: 'edit',
          name: '编辑',
          prefixIcon: 'vxe-icon-edit'
        },
        {
          code: 'delete',
          name: '删除',
          prefixIcon: 'vxe-icon-delete'
        }
      ]
    }
  }

  // 表单字段配置
  const formFields: DataFormField[] = [
    // 基础信息
    {
      field: 'serviceName',
      label: '服务名称',
      type: 'input',
      required: true,
      tabKey: 'basic',
      span: 12,
      props: {
        placeholder: '请输入服务名称',
        maxlength: 100
      },
      tips: '服务的唯一标识名称'
    },
    {
      field: 'serviceDescription',
      label: '服务描述',
      type: 'textarea',
      tabKey: 'basic',
      span: 24,
      props: {
        placeholder: '请输入服务描述',
        rows: 3,
        maxlength: 500
      },
      tips: '服务的详细说明'
    },
    {
      field: 'tunnelClientId',
      label: '客户端ID',
      type: 'custom',
      required: true,
      tabKey: 'basic',
      span: 12,
      render: (formData: Record<string, any>) => {
        return h(TunnelClientSelector, {
          modelValue: formData.tunnelClientId || '',
          'onUpdate:modelValue': (value: string) => {
            formData.tunnelClientId = value
          }
        })
      },
      tips: '服务所属的隧道客户端ID'
    },
    {
      field: 'serviceType',
      label: '服务类型',
      type: 'select',
      required: true,
      tabKey: 'basic',
      span: 12,
      props: {
        placeholder: '请选择服务类型',
        options: SERVICE_TYPE_OPTIONS
      },
      tips: '服务的协议类型'
    },
    
    // 地址配置
    {
      field: 'localAddress',
      label: '本地地址',
      type: 'input',
      required: true,
      tabKey: 'address',
      span: 12,
      props: {
        placeholder: '例如: 127.0.0.1',
        maxlength: 100
      },
      tips: '本地服务的IP地址'
    },
    {
      field: 'localPort',
      label: '本地端口',
      type: 'number',
      required: true,
      tabKey: 'address',
      span: 12,
      props: {
        placeholder: '请输入本地端口',
        min: 1,
        max: 65535
      },
      tips: '本地服务的端口号 (1-65535)'
    },
    {
      field: 'remotePort',
      label: '远程端口',
      type: 'number',
      tabKey: 'address',
      span: 12,
      props: {
        placeholder: '服务器分配的远程端口',
        min: 1,
        max: 65535
      },
      tips: '服务器端暴露的端口号，留空则由服务器自动分配'
    },
    {
      field: 'subDomain',
      label: '子域名',
      type: 'input',
      tabKey: 'address',
      span: 12,
      props: {
        placeholder: '例如: myapp',
        maxlength: 100
      },
      tips: 'HTTP/HTTPS服务的子域名'
    },
    {
      field: 'customDomains',
      label: '自定义域名',
      type: 'textarea',
      tabKey: 'address',
      span: 24,
      props: {
        placeholder: '多个域名用逗号分隔',
        rows: 2
      },
      tips: 'HTTP/HTTPS服务的自定义域名列表'
    },
    
    // 高级配置
    {
      field: 'useEncryption',
      label: '启用加密',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      props: {
        options: YES_NO_OPTIONS
      },
      tips: '是否对传输数据进行加密'
    },
    {
      field: 'useCompression',
      label: '启用压缩',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      props: {
        options: YES_NO_OPTIONS
      },
      tips: '是否对传输数据进行压缩'
    },
    {
      field: 'secretKey',
      label: '加密密钥',
      type: 'input',
      tabKey: 'advanced',
      span: 24,
      props: {
        placeholder: '请输入加密密钥',
        type: 'password',
        maxlength: 100
      },
      tips: '用于加密的密钥，启用加密时必填'
    },
    {
      field: 'maxConnections',
      label: '最大连接数',
      type: 'number',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: '0表示不限制',
        min: 0
      },
      tips: '服务允许的最大并发连接数'
    },
    {
      field: 'bandwidthLimit',
      label: '带宽限制',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: '例如: 1MB, 100KB',
        maxlength: 50
      },
      tips: '服务的带宽限制，如: 1MB, 100KB'
    },
    {
      field: 'httpUser',
      label: 'HTTP用户名',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: 'HTTP基础认证用户名',
        maxlength: 100
      },
      tips: 'HTTP/HTTPS服务的基础认证用户名'
    },
    {
      field: 'httpPassword',
      label: 'HTTP密码',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: 'HTTP基础认证密码',
        type: 'password',
        maxlength: 100
      },
      tips: 'HTTP/HTTPS服务的基础认证密码'
    },
    {
      field: 'hostHeaderRewrite',
      label: 'Host头重写',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: '例如: example.com',
        maxlength: 200
      },
      tips: 'HTTP/HTTPS服务的Host头重写'
    },
    {
      field: 'healthCheckType',
      label: '健康检查类型',
      type: 'select',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: '请选择健康检查类型',
        options: [
          { label: 'TCP', value: 'tcp' },
          { label: 'HTTP', value: 'http' }
        ]
      },
      tips: '服务的健康检查方式'
    },
    {
      field: 'healthCheckUrl',
      label: '健康检查URL',
      type: 'input',
      tabKey: 'advanced',
      span: 12,
      props: {
        placeholder: '例如: /health',
        maxlength: 200
      },
      tips: 'HTTP健康检查的URL路径'
    },
    {
      field: 'activeFlag',
      label: '激活状态',
      type: 'select',
      tabKey: 'basic',
      span: 12,
      defaultValue: 'Y',
      props: {
        options: ACTIVE_FLAG_OPTIONS
      },
      tips: '服务是否激活'
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      tabKey: 'basic',
      span: 24,
      props: {
        placeholder: '请输入备注信息',
        rows: 3,
        maxlength: 500
      },
      tips: '服务的备注说明'
    }
  ]

  // 表单标签页
  const formTabs = [
    { key: 'basic', label: '基础信息' },
    { key: 'address', label: '地址配置' },
    { key: 'advanced', label: '高级配置' }
  ]

  // 获取服务类型标签
  const getServiceTypeLabel = (type: string): string => {
    const option = SERVICE_TYPE_OPTIONS.find(opt => opt.value === type)
    return option?.label || type.toUpperCase()
  }

  // 获取服务状态标签
  const getServiceStatusLabel = (status: string): string => {
    const option = SERVICE_STATUS_OPTIONS.find(opt => opt.value === status)
    return option?.label || status
  }

  // 获取服务状态标签类型
  const getServiceStatusTagType = (status: string): 'success' | 'warning' | 'error' | 'default' => {
    switch (status) {
      case 'active':
        return 'success'
      case 'inactive':
        return 'warning'
      case 'error':
        return 'error'
      case 'offline':
        return 'default'
      default:
        return 'default'
    }
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
    updatedService: Partial<TunnelService>
  ) => {
    const index = serviceList.value.findIndex(
      (s) => s.tunnelServiceId === tunnelServiceId && s.tenantId === tenantId
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
      (s) => s.tunnelServiceId === tunnelServiceId && s.tenantId === tenantId
    )
    if (index !== -1) {
      serviceList.value.splice(index, 1)
    }
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    serviceList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setServiceList,
    clearServiceList,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList,
    getServiceTypeLabel,
    getServiceStatusLabel,
    getServiceStatusTagType
  }
}

/**
 * Model 返回类型
 */
export type TunnelServiceModel = ReturnType<typeof useTunnelServiceModel>

