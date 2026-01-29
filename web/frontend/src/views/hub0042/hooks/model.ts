/**
 * 服务监控模块 Model
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
  TrashOutline
} from '@vicons/ionicons5'
import { ref } from 'vue'
import type { Service } from '../types/index'

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

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建服务',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的服务',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的服务',
      }
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
    ] as DataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'namespaceId',
        title: '命名空间ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'groupName',
        title: '分组名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
        slots: { default: 'groupName' },
      },
      {
        field: 'serviceName',
        title: '服务名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
        slots: { default: 'serviceName' },
      },
      {
        field: 'serviceType',
        title: '服务类型',
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) => {
          const typeMap: Record<string, string> = {
            'INTERNAL': '内部服务',
            'NACOS': 'Nacos',
            'CONSUL': 'Consul',
            'EUREKA': 'Eureka',
            'ETCD': 'ETCD',
            'ZOOKEEPER': 'ZooKeeper',
          }
          return typeMap[cellValue] || cellValue
        },
      },
      {
        field: 'serviceVersion',
        title: '服务版本',
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) => {
          return cellValue || '-'
        },
      },
      {
        field: 'serviceDescription',
        title: '服务描述',
        align: 'left',
        showOverflow: true,
        width: 200,
        formatter: ({ cellValue }) => {
          return cellValue || '-'
        },
      },
      {
        field: 'nodeCount',
        title: '节点数量',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue || 0
        },
      },
      {
        field: 'healthyNodeCount',
        title: '健康节点',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue || 0
        },
      },
      {
        field: 'unhealthyNodeCount',
        title: '不健康节点',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue || 0
        },
      },
      {
        field: 'activeFlag',
        title: '活动状态',
        align: 'center',
        slots: { default: 'activeFlag' },
      },
      {
        field: 'addTime',
        title: '创建时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'addWho',
        title: '创建人',
        showOverflow: true,
      },
      {
        field: 'editTime',
        title: '修改时间',
        sortable: true,
        showOverflow: true,
        formatter: ({ cellValue }) =>
          cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        field: 'editWho',
        title: '修改人',
        showOverflow: true,
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

