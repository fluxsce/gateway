/**
 * 命名空间管理模块 Model
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
import type { ServiceGroup, ServiceGroupType } from '../types'
import {
    accessControlOptions,
    loadBalanceStrategyOptions,
    protocolTypeOptions,
    serviceGroupTypeOptions,
    statusOptions
} from '../types'

/**
 * 命名空间管理 Model
 */
export function useNamespaceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0040'
  /** 加载状态 */
  const loading = ref(false)

  /** 命名空间列表数据 */
  const namespaceList = ref<ServiceGroup[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'groupName',
        label: '名称',
        type: 'input',
        placeholder: '请输入英文命名空间名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'groupType',
        label: '类型',
        type: 'select',
        placeholder: '请选择类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...serviceGroupTypeOptions.map((opt: { label: string; value: ServiceGroupType }) => ({
            label: opt.label,
            value: opt.value
          }))
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
          ...statusOptions.map((opt: { label: string; value: 'Y' | 'N' }) => ({
            label: opt.label,
            value: opt.value
          }))
        ],
      },
      {
        field: 'accessControlEnabled',
        label: '访问控制',
        type: 'select',
        placeholder: '请选择',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          ...accessControlOptions.map((opt: { label: string; value: 'Y' | 'N' }) => ({
            label: opt.label,
            value: opt.value
          }))
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新建命名空间',
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建命名空间',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的命名空间',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的命名空间',
      }
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 数据编辑表单字段配置（供 GdataFormModal 使用） =============
  // 表单页签配置（基本信息 / 默认配置 / 权限配置 / 其他配置）
  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'config', label: '默认配置' },
    { key: 'permission', label: '权限配置' },
    { key: 'other', label: '其他配置' },
  ]

  const formFields: DataFormField[] = [
    // ============= 基本信息 Tab =============
    {
      field: 'groupName',
      label: '命名空间名称',
      type: 'input',
      placeholder: '请输入命名空间名称',
      span: 12,
      tabKey: 'basic',
      required: true,
      primary: true,
      props: {
        maxlength: 50,
        showCount: true,
      },
    },
    {
      field: 'groupType',
      label: '类型',
      type: 'select',
      placeholder: '请选择类型',
      span: 12,
      tabKey: 'basic',
      required: true,
      options: serviceGroupTypeOptions.map((opt: { label: string; value: ServiceGroupType }) => ({
        label: opt.label,
        value: opt.value
      })),
      defaultValue: 'BUSINESS',
    },
    {
      field: 'groupDescription',
      label: '描述',
      type: 'textarea',
      placeholder: '请输入描述信息',
      span: 24,
      tabKey: 'basic',
      props: {
        maxlength: 500,
        showCount: true,
        rows: 3,
      },
    },
    {
      field: 'accessControlEnabled',
      label: '访问控制',
      type: 'select',
      span: 12,
      tabKey: 'basic',
      defaultValue: 'N',
      options: accessControlOptions.map((opt: { label: string; value: 'Y' | 'N' }) => ({
        label: opt.label,
        value: opt.value
      })),
      tips: '启用后只有授权用户可以访问此命名空间',
    },

    // ============= 默认配置 Tab =============
    {
      field: 'defaultProtocolType',
      label: '默认协议',
      type: 'select',
      placeholder: '请选择默认协议',
      span: 12,
      tabKey: 'config',
      required: true,
      options: protocolTypeOptions.map((opt: { label: string; value: string }) => ({
        label: opt.label,
        value: opt.value
      })),
      defaultValue: 'HTTP',
    },
    {
      field: 'defaultLoadBalanceStrategy',
      label: '负载均衡策略',
      type: 'select',
      placeholder: '请选择负载均衡策略',
      span: 12,
      tabKey: 'config',
      required: true,
      options: loadBalanceStrategyOptions.map((opt: { label: string; value: string }) => ({
        label: opt.label,
        value: opt.value
      })),
      defaultValue: 'ROUND_ROBIN',
    },
    {
      field: 'defaultHealthCheckUrl',
      label: '健康检查URL',
      type: 'input',
      placeholder: '请输入健康检查URL',
      span: 12,
      tabKey: 'config',
      required: true,
      defaultValue: '/health',
    },
    {
      field: 'defaultHealthCheckIntervalSeconds',
      label: '检查间隔(秒)',
      type: 'number',
      placeholder: '请输入检查间隔',
      span: 12,
      tabKey: 'config',
      required: true,
      defaultValue: 30,
      props: {
        min: 5,
        max: 300,
        step: 5,
      },
    },

    // ============= 权限配置 Tab =============
    {
      field: 'adminUserIds',
      label: '管理员用户',
      type: 'select',
      placeholder: '请选择管理员用户',
      span: 24,
      tabKey: 'permission',
      props: {
        multiple: true,
        filterable: true,
        remote: true,
      },
      show: (formData: any) => formData?.accessControlEnabled === 'Y',
      tips: '管理员拥有命名空间的完全控制权限',
    },
    {
      field: 'readUserIds',
      label: '只读用户',
      type: 'select',
      placeholder: '请选择只读用户',
      span: 24,
      tabKey: 'permission',
      props: {
        multiple: true,
        filterable: true,
        remote: true,
      },
      show: (formData: any) => formData?.accessControlEnabled === 'Y',
      tips: '只读用户仅能查看命名空间信息，无法修改',
    },

    // ============= 其他配置 Tab =============
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
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'serviceGroupId',
        title: '分组ID',
        width: 180,
        showOverflow: true,
      },
      {
        field: 'groupName',
        title: '命名空间名称',
        width: 160,
        sortable: true,
        showOverflow: true,
      },
      {
        field: 'groupType',
        title: '类型',
        width: 100,
        align: 'center',
        formatter: ({ cellValue }: any) => {
          const typeConfig: Record<ServiceGroupType, string> = {
            BUSINESS: '业务',
            SYSTEM: '系统',
            TEST: '测试'
          }
          return typeConfig[cellValue as ServiceGroupType] || cellValue
        },
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => {
            const typeConfig: Record<ServiceGroupType, { type: string }> = {
              BUSINESS: { type: 'success' },
              SYSTEM: { type: 'warning' },
              TEST: { type: 'info' }
            }
            const groupType = row.groupType as ServiceGroupType
            const config = typeConfig[groupType] || { type: 'default' }
            return {
              type: config.type,
              content: typeConfig[groupType] ? 
                (groupType === 'BUSINESS' ? '业务' : groupType === 'SYSTEM' ? '系统' : '测试') : 
                groupType
            }
          },
        },
      },
      {
        field: 'groupDescription',
        title: '描述',
        width: 200,
        showOverflow: 'tooltip',
      },
      {
        field: 'ownerUserName',
        title: '所有者',
        width: 100,
        showOverflow: true,
        formatter: ({ row }: any) => row.ownerUserName || row.ownerUserId || '-',
      },
      {
        field: 'accessControlEnabled',
        title: '访问控制',
        width: 90,
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: row.accessControlEnabled === 'Y' ? 'success' : 'default',
            content: row.accessControlEnabled === 'Y' ? '已启用' : '未启用',
          }),
        },
      },
      {
        field: 'defaultProtocolType',
        title: '默认协议',
        width: 90,
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: 'info',
            content: row.defaultProtocolType,
          }),
        },
      },
      {
        field: 'defaultLoadBalanceStrategy',
        title: '负载均衡',
        width: 120,
        showOverflow: true,
        formatter: ({ cellValue }) => {
          const strategyMap: Record<string, string> = {
            ROUND_ROBIN: '轮询',
            WEIGHTED_ROUND_ROBIN: '加权轮询',
            LEAST_CONNECTIONS: '最少连接',
            RANDOM: '随机',
            IP_HASH: 'IP哈希'
          }
          return strategyMap[cellValue] || cellValue
        },
      },
      {
        field: 'defaultHealthCheckUrl',
        title: '健康检查',
        width: 130,
        showOverflow: true,
        formatter: ({ row }: any) => {
          return `${row.defaultHealthCheckUrl} (${row.defaultHealthCheckIntervalSeconds}s)`
        },
      },
      {
        field: 'serviceCount',
        title: '服务数',
        width: 80,
        align: 'center',
        formatter: ({ cellValue }) => cellValue?.toString() || '0',
      },
      {
        field: 'instanceCount',
        title: '实例数',
        width: 80,
        align: 'center',
        formatter: ({ cellValue }) => cellValue?.toString() || '0',
      },
      {
        field: 'activeFlag',
        title: '状态',
        width: 80,
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => ({
            type: row.activeFlag === 'Y' ? 'success' : 'error',
            content: row.activeFlag === 'Y' ? '活动' : '非活动',
          }),
        },
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
        field: 'addWhoName',
        title: '创建人',
        width: 90,
        showOverflow: true,
        formatter: ({ row }: any) => row.addWhoName || row.addWho || '-',
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
   * 设置命名空间列表
   */
  const setNamespaceList = (list: ServiceGroup[]) => {
    namespaceList.value = list
  }

  /**
   * 清空命名空间列表
   */
  const clearNamespaceList = () => {
    namespaceList.value = []
  }

  /**
   * 添加命名空间到列表
   */
  const addNamespaceToList = (namespace: ServiceGroup) => {
    namespaceList.value.unshift(namespace)
  }

  /**
   * 更新列表中的命名空间
   */
  const updateNamespaceInList = (
    serviceGroupId: string,
    tenantId: string,
    updatedNamespace: Partial<ServiceGroup>
  ) => {
    const index = namespaceList.value.findIndex(
      (n) => n.serviceGroupId === serviceGroupId && n.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(namespaceList.value[index], updatedNamespace)
    }
  }

  /**
   * 从列表中删除命名空间
   */
  const removeNamespaceFromList = (serviceGroupId: string, tenantId: string) => {
    const index = namespaceList.value.findIndex(
      (n) => n.serviceGroupId === serviceGroupId && n.tenantId === tenantId
    )
    if (index !== -1) {
      namespaceList.value.splice(index, 1)
    }
  }


  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    namespaceList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setNamespaceList,
    clearNamespaceList,
    addNamespaceToList,
    updateNamespaceInList,
    removeNamespaceFromList,
  }
}

/**
 * Model 返回类型
 */
export type NamespaceModel = ReturnType<typeof useNamespaceModel>

