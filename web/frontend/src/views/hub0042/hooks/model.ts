/**
 * 服务监控模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig, RsGridRowKey } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsIcon, RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import type { Service } from '../types/index'

/**
 * 服务表格配置（对齐 RsGrid Props 子集）。
 */
export interface ServiceGridConfig {
  columns: RsGridColumn<Service>[]
  selectable: boolean
  rowKey: RsGridRowKey<Service>
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

const groupColorTypes: RsTagVariant[] = ['primary', 'info', 'success', 'warning', 'danger']
const serviceColorTypes: RsTagVariant[] = ['success', 'info', 'primary', 'warning', 'danger']

/**
 * 根据字符串生成稳定哈希，用于标签配色。
 */
function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i)
    hash = hash & hash
  }
  return Math.abs(hash)
}

/**
 * 获取分组名称标签变体。
 */
function getGroupTagVariant(groupName: string): RsTagVariant {
  if (!groupName || groupName === 'DEFAULT_GROUP') return 'default'
  return groupColorTypes[hashString(groupName) % groupColorTypes.length]
}

/**
 * 获取服务名称标签变体。
 */
function getServiceTagVariant(serviceName: string): RsTagVariant {
  if (!serviceName) return 'default'
  return serviceColorTypes[hashString(serviceName) % serviceColorTypes.length]
}

/**
 * 服务联合主键（业务规则）。表格只消费函数结果，不改写行数据。
 */
export function buildServiceRowKey(
  service: Pick<Service, 'tenantId' | 'namespaceId' | 'groupName' | 'serviceName'>,
): string {
  return [service.tenantId, service.namespaceId, service.groupName, service.serviceName].join('::')
}

/**
 * 服务监控 Model
 */
export function useServiceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0042'
  /** 加载状态 */
  const loading = ref(false)

  /** 服务列表数据 */
  const serviceList = ref<Service[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

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
        field: 'groupName',
        label: '分组名称',
        type: 'input',
        placeholder: '请输入分组名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '内部服务', value: 'INTERNAL' },
          { label: 'Nacos', value: 'NACOS' },
          { label: 'Consul', value: 'CONSUL' },
          { label: 'Eureka', value: 'EUREKA' },
          { label: 'ETCD', value: 'ETCD' },
          { label: 'ZooKeeper', value: 'ZOOKEEPER' },
        ],
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
        label: '新建服务',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新建服务',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: 'CreateOutline',
        tooltip: '编辑选中的服务',
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

  // ============= 服务表单配置 =============
  const serviceFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'config', label: '服务配置' },
      { key: 'other', label: '其它' },
    ],
    fields: [
      // ============= 基本信息 Tab =============
      {
        field: 'namespaceId',
        label: '命名空间ID',
        type: 'input',
        placeholder: '命名空间ID（自动填充）',
        span: 12,
        tabKey: 'basic',
        required: true,
        disabled: true, // 始终禁用，从选中的命名空间自动填充
        tips: '命名空间ID（主键），从上方命名空间列表自动获取',
      },
      {
        field: 'groupName',
        label: '分组名称',
        type: 'input',
        placeholder: '请输入分组名称，如：DEFAULT_GROUP',
        span: 12,
        tabKey: 'basic',
        required: true,
        primary: true,
        defaultValue: 'DEFAULT_GROUP',
        tips: '分组名称（主键），编辑模式下不允许修改',
      },
      {
        field: 'serviceName',
        label: '服务名称',
        type: 'input',
        placeholder: '请输入服务名称',
        span: 12,
        tabKey: 'basic',
        required: true,
        primary: true,
        tips: '服务名称（主键），编辑模式下不允许修改',
      },
      {
        field: 'serviceType',
        label: '服务类型',
        type: 'select',
        placeholder: '请选择服务类型',
        span: 12,
        tabKey: 'basic',
        required: true,
        defaultValue: 'INTERNAL',
        options: [
          { label: '内部服务', value: 'INTERNAL' },
          { label: 'Nacos', value: 'NACOS' },
          { label: 'Consul', value: 'CONSUL' },
          { label: 'Eureka', value: 'EUREKA' },
          { label: 'ETCD', value: 'ETCD' },
          { label: 'ZooKeeper', value: 'ZOOKEEPER' },
        ],
      },
      {
        field: 'serviceVersion',
        label: '服务版本',
        type: 'input',
        placeholder: '请输入服务版本号',
        span: 12,
        tabKey: 'basic',
      },
      {
        field: 'serviceDescription',
        label: '服务描述',
        type: 'textarea',
        placeholder: '请输入服务描述',
        span: 24,
        tabKey: 'basic',
        props: {
          rows: 3,
        },
      },
      {
        field: 'activeFlag',
        label: '活动状态',
        type: 'switch',
        span: 12,
        tabKey: 'basic',
        defaultValue: 'Y',
        props: {
          checkedValue: 'Y',
          uncheckedValue: 'N',
        },
      },
      // ============= 服务配置 Tab =============
      {
        field: 'protectThreshold',
        label: '保护阈值',
        type: 'number',
        placeholder: '0.00',
        span: 12,
        tabKey: 'config',
        defaultValue: 0.00,
        tips: '服务保护阈值，范围0.00-1.00，表示健康实例比例低于该值时触发保护',
        props: {
          min: 0,
          max: 1,
          step: 0.01,
          precision: 2,
        },
      },
      {
        field: 'externalServiceConfig',
        label: '外部服务配置',
        type: 'textarea',
        placeholder: '请输入外部服务配置（JSON格式）',
        span: 24,
        tabKey: 'config',
        tips: '外部服务配置，JSON格式，存储外部注册中心的连接配置等信息',
        props: {
          rows: 5,
        },
      },
      {
        field: 'metadataJson',
        label: '服务元数据',
        type: 'textarea',
        placeholder: '请输入服务元数据（JSON格式）',
        span: 24,
        tabKey: 'config',
        tips: '服务元数据，JSON格式，存储服务的扩展信息',
        props: {
          rows: 5,
        },
      },
      {
        field: 'tagsJson',
        label: '服务标签',
        type: 'textarea',
        placeholder: '请输入服务标签（JSON格式）',
        span: 24,
        tabKey: 'config',
        tips: '服务标签，JSON格式，用于服务分类和过滤',
        props: {
          rows: 3,
        },
      },
      {
        field: 'selectorJson',
        label: '服务选择器',
        type: 'textarea',
        placeholder: '请输入服务选择器（JSON格式）',
        span: 24,
        tabKey: 'config',
        tips: '服务选择器，JSON格式，用于服务路由规则',
        props: {
          rows: 5,
        },
      },
      // ============= 其它 Tab =============
      {
        field: 'noteText',
        label: '备注',
        type: 'textarea',
        placeholder: '请输入备注信息',
        span: 24,
        tabKey: 'other',
        props: {
          rows: 3,
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
    ] as RsDataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: ServiceGridConfig = {
    columns: [
      {
        key: 'namespaceId',
        title: '命名空间ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'groupName',
        title: '分组名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        render: (row) =>
          h(RsTag, { variant: getGroupTagVariant(row.groupName), size: 'sm' }, () => row.groupName),
      },
      {
        key: 'serviceName',
        title: '服务名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        render: (row) =>
          h(
            RsTag,
            { variant: getServiceTagVariant(row.serviceName), size: 'sm' },
            () => [h(RsIcon, { name: 'server', size: 12 }), ` ${row.serviceName}`],
          ),
      },
      {
        key: 'serviceType',
        title: '服务类型',
        align: 'center',
        ellipsis: true,
        formatter: (value) => {
          const typeMap: Record<string, string> = {
            INTERNAL: '内部服务',
            NACOS: 'Nacos',
            CONSUL: 'Consul',
            EUREKA: 'Eureka',
            ETCD: 'ETCD',
            ZOOKEEPER: 'ZooKeeper',
          }
          return typeMap[String(value || '')] || String(value || '')
        },
      },
      {
        key: 'serviceVersion',
        title: '服务版本',
        align: 'center',
        ellipsis: true,
        formatter: (value) => (value ? String(value) : '-'),
      },
      {
        key: 'serviceDescription',
        title: '服务描述',
        align: 'left',
        ellipsis: true,
        width: 200,
        formatter: (value) => (value ? String(value) : '-'),
      },
      {
        key: 'nodeCount',
        title: '节点数量',
        align: 'center',
        formatter: (value) => String(value || 0),
      },
      {
        key: 'healthyNodeCount',
        title: '健康节点',
        align: 'center',
        formatter: (value) => String(value || 0),
      },
      {
        key: 'unhealthyNodeCount',
        title: '不健康节点',
        align: 'center',
        formatter: (value) => String(value || 0),
      },
      {
        key: 'activeFlag',
        title: '活动状态',
        align: 'center',
        width: 100,
        render: (row) =>
          h(
            RsTag,
            { variant: row.activeFlag === 'Y' ? 'success' : 'default', size: 'sm' },
            () => (row.activeFlag === 'Y' ? '活动' : '非活动'),
          ),
      },
      {
        key: 'addTime',
        title: '创建时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'addWho',
        title: '创建人',
        ellipsis: true,
      },
      {
        key: 'editTime',
        title: '修改时间',
        sortable: true,
        ellipsis: true,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'editWho',
        title: '修改人',
        ellipsis: true,
      },
    ],
    selectable: true,
    rowKey: buildServiceRowKey,
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
        { key: 'delete', label: '删除', icon: 'trash-2', danger: true },
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
   * 设置服务列表
   */
  const setServiceList = (list: Service[]) => {
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
  const addServiceToList = (service: Service) => {
    serviceList.value.unshift(service)
  }

  /**
   * 更新列表中的服务
   */
  const updateServiceInList = (
    namespaceId: string,
    groupName: string,
    serviceName: string,
    tenantId: string,
    updatedService: Partial<Service>
  ) => {
    const index = serviceList.value.findIndex(
      (s) => s.namespaceId === namespaceId && s.groupName === groupName && s.serviceName === serviceName && s.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(serviceList.value[index], updatedService)
    }
  }

  /**
   * 从列表中删除服务
   */
  const removeServiceFromList = (
    namespaceId: string,
    groupName: string,
    serviceName: string,
    tenantId: string
  ) => {
    const index = serviceList.value.findIndex(
      (s) => s.namespaceId === namespaceId && s.groupName === groupName && s.serviceName === serviceName && s.tenantId === tenantId
    )
    if (index !== -1) {
      serviceList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除服务
   */
  const removeServicesFromList = (services: Service[]) => {
    services.forEach((service) => {
      removeServiceFromList(service.namespaceId, service.groupName, service.serviceName, service.tenantId)
    })
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
    serviceFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setServiceList,
    clearServiceList,
    addServiceToList,
    updateServiceInList,
    removeServiceFromList,
    removeServicesFromList,
  }
}

/**
 * Model 返回类型
 */
export type ServiceModel = ReturnType<typeof useServiceModel>

