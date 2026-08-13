/**
 * API访问控制配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { RsDynamicTags, RsTag } from '@/ui'
import { h, ref } from 'vue'
import type { ApiAccessConfig } from './types'

/**
 * API访问控制配置表格配置（对齐 RsGrid Props 子集）。
 */
export interface ApiAccessConfigGridConfig {
  columns: RsGridColumn<ApiAccessConfig>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * API访问控制配置列表 Model
 * @param moduleId 模块ID（用于权限控制，由父级传入）
 */
export function useApiAccessConfigModel(moduleId: string) {
  // ============= 数据状态 =============
  /** 加载状态 */
  const loading = ref(false)

  /** API配置列表数据 */
  const configList = ref<ApiAccessConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'configName',
        label: '配置名称',
        type: 'input',
        placeholder: '请输入配置名称',
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
        label: '新建配置',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新建API访问控制配置',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: 'CreateOutline',
        tooltip: '编辑选中的配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: 'TrashOutline',
        type: 'error',
        tooltip: '删除选中的配置',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表单配置 =============

  // 创建路径列表渲染函数
  const createPathListRender = (field: 'whitelistPaths' | 'blacklistPaths', placeholder: string) => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(RsDynamicTags, {
        modelValue: value,
        'onUpdate:modelValue': (newValue: string[]) => {
          formData[field] = newValue
        },
        placeholder,
      })
    }
  }

  // 创建HTTP方法列表渲染函数
  const createMethodListRender = (field: 'allowedMethods' | 'blockedMethods', placeholder: string) => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(RsDynamicTags, {
        modelValue: value,
        'onUpdate:modelValue': (newValue: string[]) => {
          formData[field] = newValue
        },
        placeholder,
      })
    }
  }

  /** 表单字段配置 */
  const formFields: RsDataFormField[] = [
    // ============= 主键字段（隐藏，但必须存在用于更新） =============
    {
      field: 'apiAccessConfigId',
      label: 'API访问配置ID',
      type: 'input',
      span: 8,
      show: false,
    },
    {
      field: 'tenantId',
      label: '租户ID',
      type: 'input',
      span: 8,
      show: false,
    },
    {
      field: 'securityConfigId',
      label: '安全配置ID',
      type: 'input',
      span: 8,
      show: false,
    },
    // ============= 基础配置分组 =============
    {
      field: 'basicConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'configName',
          label: '配置名称',
          type: 'input',
          placeholder: '请输入配置名称',
          span: 12,
          required: true,
        },
        {
          field: 'defaultPolicy',
          label: '默认策略',
          type: 'select',
          span: 12,
          defaultValue: 'allow',
          tips: 'allow（允许）: 默认允许，黑名单中的API会被拒绝。deny（拒绝）: 默认拒绝，仅白名单允许。黑名单优先级高于白名单。',
          options: [
            { label: '允许（白名单模式）', value: 'allow' },
            { label: '拒绝（黑名单模式）', value: 'deny' },
          ],
        },
      ],
    },
    // ============= API路径配置分组 =============
    {
      field: 'pathConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'whitelistPaths',
          label: 'API路径白名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '支持通配符匹配，如：/api/v1/*',
          render: createPathListRender('whitelistPaths', '输入允许的API路径，如：/api/v1/*'),
        },
        {
          field: 'blacklistPaths',
          label: 'API路径黑名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '支持通配符匹配，优先级高于白名单',
          render: createPathListRender('blacklistPaths', '输入禁止的API路径，如：/api/admin/*'),
        },
      ],
    },
    // ============= HTTP方法配置分组 =============
    {
      field: 'methodConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'allowedMethods',
          label: '允许的HTTP方法',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '如：GET, POST, PUT, DELETE',
          render: createMethodListRender('allowedMethods', '输入允许的HTTP方法，如：GET, POST'),
        },
        {
          field: 'blockedMethods',
          label: '禁止的HTTP方法',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '优先级高于允许的方法',
          render: createMethodListRender('blockedMethods', '输入禁止的HTTP方法，如：DELETE'),
        },
      ],
    },
    // ============= 活动状态（放在最后） =============
    {
      field: 'activeFlag',
      label: '活动状态',
      type: 'switch',
      span: 12,
      defaultValue: 'Y',
      props: {
        checkedValue: 'Y',
        uncheckedValue: 'N',
      },
    },
  ]

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid 结构） */
  const gridConfig: ApiAccessConfigGridConfig = {
    columns: [
      {
        key: 'apiAccessConfigId',
        title: 'API访问配置ID',
        width: 200,
        visible: false,
      },
      {
        key: 'tenantId',
        title: '租户ID',
        width: 200,
        visible: false,
      },
      {
        key: 'securityConfigId',
        title: '安全配置ID',
        width: 200,
        visible: false,
      },
      {
        key: 'configName',
        title: '配置名称',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'defaultPolicy',
        title: '默认策略',
        width: 120,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.defaultPolicy === 'allow' ? 'success' : 'danger',
              size: 'sm',
            },
            () => (row.defaultPolicy === 'allow' ? '允许' : '拒绝'),
          ),
      },
      {
        key: 'activeFlag',
        title: '活动状态',
        width: 100,
        align: 'center',
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.activeFlag === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.activeFlag === 'Y' ? '活动' : '非活动'),
          ),
      },
      {
        key: 'whitelistPaths',
        title: '路径白名单',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'blacklistPaths',
        title: '路径黑名单',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'allowedMethods',
        title: '允许的方法',
        width: 150,
        ellipsis: true,
      },
      {
        key: 'blockedMethods',
        title: '禁止的方法',
        width: 150,
        ellipsis: true,
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
        width: 120,
      },
    ],
    selectable: true,
    rowKey: 'apiAccessConfigId',
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

  // ============= 工具方法 =============

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新分页信息
   */
  const updatePagination = (info: Partial<PageInfoObj>) => {
    if (!pageInfo.value) {
      pageInfo.value = info as PageInfoObj
    } else {
      Object.assign(pageInfo.value, info)
    }
  }

  /**
   * 设置配置列表
   */
  const setConfigList = (list: ApiAccessConfig[]) => {
    configList.value = list
  }

  /**
   * 添加配置到列表
   */
  const addConfigToList = (config: ApiAccessConfig) => {
    configList.value.push(config)
  }

  /**
   * 更新列表中的配置
   */
  const updateConfigInList = (apiAccessConfigId: string, tenantId: string, updatedConfig: Partial<ApiAccessConfig>) => {
    const index = configList.value.findIndex(
      (item) => item.apiAccessConfigId === apiAccessConfigId && item.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(configList.value[index], updatedConfig)
    }
  }

  /**
   * 从列表中移除配置
   */
  const removeConfigFromList = (apiAccessConfigId: string, tenantId: string) => {
    const index = configList.value.findIndex(
      (item) => item.apiAccessConfigId === apiAccessConfigId && item.tenantId === tenantId
    )
    if (index !== -1) {
      configList.value.splice(index, 1)
    }
  }

  return {
    moduleId,
    loading,
    configList,
    pageInfo,
    searchFormConfig,
    formFields,
    gridConfig,
    resetPagination,
    updatePagination,
    setConfigList,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
  }
}
