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
import { h, ref } from 'vue'
import { ServiceCenterInstanceNameSelector } from '../../hub0040/components'
import type { ServiceCenterInstance } from '../../hub0040/types'
import type { Namespace } from '../types/index'

/**
 * 命名空间管理 Model
 */
export function useNamespaceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0041'
  /** 加载状态 */
  const loading = ref(false)

  /** 命名空间列表数据 */
  const namespaceList = ref<Namespace[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'namespaceName',
        label: '命名空间名称',
        type: 'input',
        placeholder: '请输入命名空间名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'instanceName',
        label: '服务中心实例',
        type: 'custom',
        span: 6,
        render: (formData: Record<string, any>) => {
          return h(ServiceCenterInstanceNameSelector, {
            modelValue: formData.instanceName || '',
            'onUpdate:modelValue': (value: string) => {
              formData.instanceName = value
            },
          })
        },
      },
      {
        field: 'environment',
        label: '部署环境',
        type: 'select',
        placeholder: '请选择环境',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '开发环境', value: 'DEVELOPMENT' },
          { label: '预发布环境', value: 'STAGING' },
          { label: '生产环境', value: 'PRODUCTION' },
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

  // ============= 命名空间表单配置 =============
  const namespaceFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'quota', label: '配额配置' },
      { key: 'other', label: '其它' },
    ],
    fields: [
      // ============= 基本信息 Tab =============
      {
        field: 'namespaceId',
        label: '命名空间ID',
        type: 'input',
        placeholder: '自动生成或手动输入（32个字符：字母、数字、下划线）',
        span: 12,
        tabKey: 'basic',
        primary: true,
        required: true,
        tips: '命名空间唯一标识（主键），32个字符：字母、数字、下划线，如果不填写将自动生成',
        rules: [
          {
            pattern: /^[a-zA-Z0-9_]{1,32}$/,
            message: '命名空间ID只能包含字母、数字、下划线，且长度不超过32个字符',
            trigger: ['change', 'blur']
          }
        ],
        props: {
          maxlength: 32,
        },
      },
      {
        field: 'namespaceName',
        label: '命名空间名称',
        type: 'input',
        placeholder: '请输入命名空间名称（32个字符：字母、数字、下划线）',
        span: 12,
        tabKey: 'basic',
        required: true,
        rules: [
          {
            pattern: /^[a-zA-Z0-9_]{1,32}$/,
            message: '命名空间名称只能包含字母、数字、下划线，且长度不超过32个字符',
            trigger: ['change', 'blur']
          }
        ],
        props: {
          maxlength: 32,
        },
      },
      {
        field: 'instanceName',
        label: '服务中心实例',
        type: 'custom',
        span: 12,
        tabKey: 'basic',
        required: true,
        tips: '关联的服务中心实例名称',
        render: (formData: Record<string, any>) => {
          return h(ServiceCenterInstanceNameSelector, {
            modelValue: formData.instanceName || '',
            'onUpdate:modelValue': (value: string) => {
              formData.instanceName = value
            },
            onSelect: (instance: ServiceCenterInstance) => {
              // 选择实例后，自动填充环境
              if (instance && instance.environment) {
                formData.environment = instance.environment
              }
            },
          })
        },
      },
      {
        field: 'environment',
        label: '部署环境',
        type: 'select',
        placeholder: '请选择部署环境',
        span: 12,
        tabKey: 'basic',
        required: true,
        disabled: true,
        options: [
          { label: '开发环境', value: 'DEVELOPMENT' },
          { label: '预发布环境', value: 'STAGING' },
          { label: '生产环境', value: 'PRODUCTION' },
        ],
        tips: '部署环境由选择的服务中心实例自动确定',
      },
      {
        field: 'namespaceDescription',
        label: '命名空间描述',
        type: 'textarea',
        placeholder: '请输入命名空间描述',
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
      // ============= 配额配置 Tab =============
      {
        field: 'serviceQuotaLimit',
        label: '服务数量配额限制',
        type: 'number',
        placeholder: '200',
        span: 12,
        tabKey: 'quota',
        defaultValue: 200,
        tips: '该命名空间下允许的最大服务数量，0表示无限制',
        props: {
          min: 0,
        },
      },
      {
        field: 'configQuotaLimit',
        label: '配置数量配额限制',
        type: 'number',
        placeholder: '200',
        span: 12,
        tabKey: 'quota',
        defaultValue: 200,
        tips: '该命名空间下允许的最大配置数量，0表示无限制',
        props: {
          min: 0,
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
        field: 'namespaceName',
        title: '命名空间名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'instanceName',
        title: '服务中心实例',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'environment',
        title: '部署环境',
        sortable: true,
        align: 'center',
        showOverflow: true,
        formatter: ({ cellValue }) => {
          const envMap: Record<string, string> = {
            'DEVELOPMENT': '开发环境',
            'STAGING': '预发布环境',
            'PRODUCTION': '生产环境',
          }
          return envMap[cellValue] || cellValue
        },
      },
      {
        field: 'namespaceDescription',
        title: '描述',
        align: 'left',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'serviceQuotaLimit',
        title: '服务配额',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === 0 ? '无限制' : cellValue
        },
      },
      {
        field: 'configQuotaLimit',
        title: '配置配额',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === 0 ? '无限制' : cellValue
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
   * 设置命名空间列表
   */
  const setNamespaceList = (list: Namespace[]) => {
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
  const addNamespaceToList = (namespace: Namespace) => {
    namespaceList.value.unshift(namespace)
  }

  /**
   * 更新列表中的命名空间
   */
  const updateNamespaceInList = (
    namespaceId: string,
    tenantId: string,
    updatedNamespace: Partial<Namespace>
  ) => {
    const index = namespaceList.value.findIndex(
      (n) => n.namespaceId === namespaceId && n.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(namespaceList.value[index], updatedNamespace)
    }
  }

  /**
   * 从列表中删除命名空间
   */
  const removeNamespaceFromList = (
    namespaceId: string,
    tenantId: string
  ) => {
    const index = namespaceList.value.findIndex(
      (n) => n.namespaceId === namespaceId && n.tenantId === tenantId
    )
    if (index !== -1) {
      namespaceList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除命名空间
   */
  const removeNamespacesFromList = (namespaces: Namespace[]) => {
    namespaces.forEach((namespace) => {
      removeNamespaceFromList(namespace.namespaceId, namespace.tenantId)
    })
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
    namespaceFormConfig,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setNamespaceList,
    clearNamespaceList,
    addNamespaceToList,
    updateNamespaceInList,
    removeNamespaceFromList,
    removeNamespacesFromList,
  }
}

/**
 * Model 返回类型
 */
export type NamespaceModel = ReturnType<typeof useNamespaceModel>

