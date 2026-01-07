/**
 * 权限资源管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { EyeOutline } from '@vicons/ionicons5'
import { ref } from 'vue'
import type { Resource } from '../types/index'
import { BuiltInFlag, FlagEnum, ResourceStatus, ResourceType } from '../types/index'

/**
 * 资源管理 Model
 */
export function useResourceModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0006'
  /** 加载状态 */
  const loading = ref(false)

  /** 资源列表数据 */
  const resourceList = ref<Resource[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'resourceName',
        label: '资源名称',
        type: 'input',
        placeholder: '请输入资源名称',
        span: 6,
        clearable: true,
      },
      {
        field: 'resourceCode',
        label: '资源编码',
        type: 'input',
        placeholder: '请输入资源编码',
        span: 6,
        clearable: true,
      },
      {
        field: 'resourceType',
        label: '资源类型',
        type: 'select',
        placeholder: '请选择资源类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '模块', value: ResourceType.MODULE },
          { label: '菜单', value: ResourceType.MENU },
          { label: '按钮', value: ResourceType.BUTTON },
          { label: '接口', value: ResourceType.API },
        ],
      },
      {
        field: 'resourceStatus',
        label: '状态',
        type: 'select',
        placeholder: '请选择状态',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '启用', value: ResourceStatus.ENABLED },
          { label: '禁用', value: ResourceStatus.DISABLED },
        ],
      },
      {
        field: 'builtInFlag',
        label: '类型',
        type: 'select',
        placeholder: '请选择类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '内置', value: BuiltInFlag.BUILT_IN },
          { label: '自定义', value: BuiltInFlag.CUSTOM },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'view',
        label: '查看详情',
        icon: EyeOutline,
        tooltip: '查看选中资源的详情',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 数据编辑表单字段配置（供 GdataFormModal 使用） =============
  const formTabs = [
    { key: 'basic', label: '基本信息' },
    { key: 'hierarchy', label: '层级关系' },
    { key: 'other', label: '其他信息' },
  ]

  const formFields: DataFormField[] = [
    {
      field: 'resourceId',
      label: '资源ID',
      type: 'input',
      placeholder: '请输入资源ID',
      span: 8,
      tabKey: 'basic',
      required: true,
      primary: true,
    },
    {
      field: 'resourceName',
      label: '资源名称',
      type: 'input',
      placeholder: '请输入资源名称',
      span: 8,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'resourceCode',
      label: '资源编码',
      type: 'input',
      placeholder: '请输入资源编码',
      span: 8,
      tabKey: 'basic',
      required: true,
    },
    {
      field: 'resourceType',
      label: '资源类型',
      type: 'select',
      placeholder: '请选择资源类型',
      span: 8,
      tabKey: 'basic',
      required: true,
      options: [
        { label: '模块', value: ResourceType.MODULE },
        { label: '菜单', value: ResourceType.MENU },
        { label: '按钮', value: ResourceType.BUTTON },
        { label: '接口', value: ResourceType.API },
      ],
    },
    {
      field: 'resourcePath',
      label: '资源路径',
      type: 'input',
      placeholder: '请输入资源路径（菜单路径或API路径）',
      span: 12,
      tabKey: 'basic',
    },
    {
      field: 'resourceMethod',
      label: '请求方法',
      type: 'select',
      placeholder: '请选择请求方法',
      span: 6,
      tabKey: 'basic',
      options: [
        { label: 'GET', value: 'GET' },
        { label: 'POST', value: 'POST' },
        { label: 'PUT', value: 'PUT' },
        { label: 'DELETE', value: 'DELETE' },
        { label: 'PATCH', value: 'PATCH' },
      ],
    },
    {
      field: 'displayName',
      label: '显示名称',
      type: 'input',
      placeholder: '请输入显示名称',
      span: 8,
      tabKey: 'basic',
    },
    {
      field: 'iconClass',
      label: '图标样式类',
      type: 'input',
      placeholder: '请输入图标样式类',
      span: 8,
      tabKey: 'basic',
    },
    {
      field: 'description',
      label: '资源描述',
      type: 'textarea',
      placeholder: '请输入资源描述',
      span: 24,
      tabKey: 'basic',
    },
    {
      field: 'resourceStatus',
      label: '资源状态',
      type: 'select',
      placeholder: '请选择状态',
      span: 8,
      tabKey: 'basic',
      defaultValue: ResourceStatus.ENABLED,
      options: [
        { label: '启用', value: ResourceStatus.ENABLED },
        { label: '禁用', value: ResourceStatus.DISABLED },
      ],
    },
    {
      field: 'builtInFlag',
      label: '内置标记',
      type: 'select',
      placeholder: '请选择类型',
      span: 8,
      tabKey: 'basic',
      defaultValue: BuiltInFlag.CUSTOM,
      options: [
        { label: '内置', value: BuiltInFlag.BUILT_IN },
        { label: '自定义', value: BuiltInFlag.CUSTOM },
      ],
    },
    {
      field: 'parentResourceId',
      label: '父资源ID',
      type: 'input',
      placeholder: '请输入父资源ID',
      span: 8,
      tabKey: 'hierarchy',
    },
    {
      field: 'resourceLevel',
      label: '资源层级',
      type: 'number',
      placeholder: '请输入资源层级',
      span: 8,
      tabKey: 'hierarchy',
      defaultValue: 1,
    },
    {
      field: 'sortOrder',
      label: '排序顺序',
      type: 'number',
      placeholder: '请输入排序顺序',
      span: 8,
      tabKey: 'hierarchy',
      defaultValue: 0,
    },
    {
      field: 'addTime',
      label: '创建时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'addWho',
      label: '创建人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editTime',
      label: '修改时间',
      type: 'datetime',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'editWho',
      label: '修改人',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'oprSeqFlag',
      label: '操作序列标识',
      type: 'input',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'currentVersion',
      label: '当前版本号',
      type: 'number',
      span: 8,
      disabled: true,
      tabKey: 'other',
    },
    {
      field: 'activeFlag',
      label: '活动标记',
      type: 'select',
      span: 8,
      tabKey: 'basic',
      defaultValue: FlagEnum.YES,
      options: [
        { label: '活动', value: FlagEnum.YES },
        { label: '非活动', value: FlagEnum.NO },
      ],
    },
    {
      field: 'noteText',
      label: '备注',
      type: 'textarea',
      placeholder: '请输入备注信息',
      span: 24,
      tabKey: 'basic',
    },
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    rowId: 'resourceId', // 设置行唯一标识，用于树形结构
    columns: [
      {
        field: 'resourceId',
        title: '资源ID',
        showOverflow: true,
        width:200,
        treeNode: true, // 树形表格需要在列上设置 treeNode 属性
      },
      {
        field: 'resourceName',
        title: '资源名称',
        sortable: true,
        showOverflow: true,
      },
      {
        field: 'resourceCode',
        title: '资源编码',
        sortable: true,
        width:200,
        showOverflow: true,
      },
      {
        field: 'resourceType',
        title: '资源类型',
        align: 'center',
        cellRender: {
          name: 'VxeTag',
          props: ({ row }: any) => {
            const typeMap: Record<string, { type: string; content: string }> = {
              [ResourceType.MODULE]: { type: 'info', content: '模块' },
              [ResourceType.MENU]: { type: 'success', content: '菜单' },
              [ResourceType.BUTTON]: { type: 'warning', content: '按钮' },
              [ResourceType.API]: { type: 'primary', content: '接口' },
            }
            return typeMap[row.resourceType] || { type: 'default', content: row.resourceType }
          },
        },
      },
      {
        field: 'resourcePath',
        title: '资源路径',
        showOverflow: 'tooltip',
      },
      {
        field: 'resourceMethod',
        title: '请求方法',
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'resourceLevel',
        title: '层级',
        align: 'center',
        sortable: true,
      },
      {
        field: 'sortOrder',
        title: '排序',
        align: 'center',
        sortable: true,
      },
      {
        field: 'resourceStatus',
        title: '状态',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === ResourceStatus.ENABLED ? '启用' : '禁用'
        },
      },
      {
        field: 'builtInFlag',
        title: '类型',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === BuiltInFlag.BUILT_IN ? '内置' : '自定义'
        },
      },
      {
        field: 'activeFlag',
        title: '活动标记',
        align: 'center',
        formatter: ({ cellValue }) => {
          return cellValue === FlagEnum.YES ? '活动' : '非活动'
        },
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
      {
        field: 'description',
        title: '描述',
        showOverflow: 'tooltip',
      },
    ],
    showCheckbox: false, // 权限资源模块只允许查看，不需要复选框
    paginationConfig: {
      show: false, // 树形结构不需要分页
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
        // 权限资源模块只允许查看，不允许编辑和删除
      ],
    },
    // 树形配置：资源有层级关系，需要树形展示
    // 后端已返回树形结构（包含children），所以不需要transform
    // 注意：使用 type=expand 列时，不能使用 showLine
    treeConfig: {
      transform: false, // 数据已经是树形结构，不需要转换
      childrenField: 'children', // 子节点字段名
      indent: 20, // 每一级的缩进距离
      iconOpen: 'vxe-icon-arrow-down', // 展开图标
      iconClose: 'vxe-icon-arrow-right', // 折叠图标
      expandAll: false, // 默认不展开所有节点
      accordion: false, // 是否每次只展开一个同级树节点
      trigger: 'default', // 触发方式：default（点击箭头）、cell（点击单元格）、row（点击行）
    },
    // 展开配置：当使用 tree-config.transform=false 时，需要设置 expand-config.mode=fixed
    expandConfig: {
      mode: 'fixed', // 固定模式，避免与树形结构冲突
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
   * 设置资源列表
   */
  const setResourceList = (list: Resource[]) => {
    resourceList.value = list
  }

  /**
   * 清空资源列表
   */
  const clearResourceList = () => {
    resourceList.value = []
  }

  /**
   * 添加资源到列表
   */
  const addResourceToList = (resource: Resource) => {
    resourceList.value.unshift(resource)
  }

  /**
   * 更新列表中的资源
   */
  const updateResourceInList = (resourceId: string, tenantId: string, updatedResource: Partial<Resource>) => {
    const index = resourceList.value.findIndex((r) => r.resourceId === resourceId && r.tenantId === tenantId)
    if (index !== -1) {
      Object.assign(resourceList.value[index], updatedResource)
    }
  }

  /**
   * 从列表中删除资源
   */
  const removeResourceFromList = (resourceId: string, tenantId: string) => {
    const index = resourceList.value.findIndex((r) => r.resourceId === resourceId && r.tenantId === tenantId)
    if (index !== -1) {
      resourceList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除资源
   */
  const removeResourcesFromList = (resources: Resource[]) => {
    resources.forEach((resource) => {
      removeResourceFromList(resource.resourceId, resource.tenantId)
    })
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    resourceList,
    pageInfo,

    // 配置
    searchFormConfig,
    formTabs,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setResourceList,
    clearResourceList,
    addResourceToList,
    updateResourceInList,
    removeResourceFromList,
    removeResourcesFromList,
  }
}

/**
 * Model 返回类型
 */
export type ResourceModel = ReturnType<typeof useResourceModel>

