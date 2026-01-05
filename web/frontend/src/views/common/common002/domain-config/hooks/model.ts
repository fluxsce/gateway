/**
 * 域名访问控制配置列表 Model
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
import { NDynamicTags, NSelect, NTooltip } from 'naive-ui'
import { h, ref } from 'vue'
import type { DomainAccessConfig } from './types'

/**
 * 域名访问控制配置列表 Model
 */
export function useDomainAccessConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hubcommon002-domain-access'
  /** 加载状态 */
  const loading = ref(false)

  /** 域名配置列表数据 */
  const configList = ref<DomainAccessConfig[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
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
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建域名访问控制配置',
      },
      {
        key: 'edit',
        label: '编辑',
        icon: CreateOutline,
        tooltip: '编辑选中的配置',
      },
      {
        key: 'delete',
        label: '删除',
        icon: TrashOutline,
        type: 'error',
        tooltip: '删除选中的配置',
      }
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表单配置 =============

  // 创建域名列表渲染函数
  const createDomainListRender = (field: 'whitelistDomains' | 'blacklistDomains', placeholder: string) => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(NDynamicTags, {
        value,
        'onUpdate:value': (newValue: string[]) => {
          formData[field] = newValue
        },
        inputProps: {
          placeholder,
        },
      })
    }
  }

  /** 表单字段配置 */
  const formFields: DataFormField[] = [
    // ============= 主键字段（隐藏，但必须存在用于更新） =============
    {
      field: 'domainAccessConfigId',
      label: '域名访问配置ID',
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
          type: 'custom',
          span: 12,
          defaultValue: 'allow',
          render: (formData: Record<string, any>) => {
            return h(NTooltip, {
              trigger: 'hover',
              placement: 'top',
            }, {
              trigger: () => h('div', { style: 'width: 100%;' }, [
                h(NSelect, {
                  value: formData.defaultPolicy,
                  'onUpdate:value': (value: string) => {
                    formData.defaultPolicy = value
                  },
                  placeholder: '请选择默认策略',
                  options: [
                    { label: '允许（白名单模式）', value: 'allow' },
                    { label: '拒绝（黑名单模式）', value: 'deny' },
                  ],
                }),
              ]),
              default: () => h('div', { style: 'max-width: 320px; line-height: 1.5;' }, [
                h('p', { style: 'margin: 0 0 8px 0;' }, [
                  h('strong', '默认策略说明：'),
                ]),
                h('p', { style: 'margin: 0 0 8px 0;' }, [
                  h('strong', '• allow（允许）:'),
                  ' 默认允许访问，只有在黑名单中的域名会被拒绝',
                ]),
                h('p', { style: 'margin: 0 0 8px 0;' }, [
                  h('strong', '• deny（拒绝）:'),
                  ' 默认拒绝访问，只有在白名单中的域名才被允许',
                ]),
                h('p', { style: 'margin: 0; color: #f0a020;' }, [
                  h('strong', '⚠️ 重要：'),
                  '黑名单优先级高于白名单，无论默认策略如何，黑名单中的域名都会被拒绝',
                ]),
              ]),
            })
          },
        },
        {
          field: 'allowSubdomains',
          label: '允许子域名',
          type: 'switch',
          span: 12,
          defaultValue: 'Y',
          tips: '是否允许匹配子域名（如：*.example.com）',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    // ============= 域名白名单分组 =============
    {
      field: 'whitelistConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'whitelistDomains',
          label: '域名白名单',
          type: 'custom',
          span: 24,
          defaultValue: [],
          tips: '如：example.com, api.example.com，白名单中的域名将被允许访问',
          render: createDomainListRender('whitelistDomains', '输入允许的域名，如：example.com'),
        },
      ],
    },
    // ============= 域名黑名单分组 =============
    {
      field: 'blacklistConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'blacklistDomains',
          label: '域名黑名单',
          type: 'custom',
          span: 24,
          defaultValue: [],
          tips: '优先级高于白名单，直接拒绝访问',
          render: createDomainListRender('blacklistDomains', '输入禁止的域名，如：malicious.com'),
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

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      // ============= 主键字段（隐藏，但必须存在用于数据操作） =============
      {
        field: 'domainAccessConfigId',
        title: '域名访问配置ID',
        visible: false,
      },
      {
        field: 'tenantId',
        title: '租户ID',
        visible: false,
      },
      {
        field: 'securityConfigId',
        title: '安全配置ID',
        visible: false,
      },
      // ============= 业务字段 =============
      {
        field: 'configName',
        title: '配置名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
      },
      {
        field: 'defaultPolicy',
        title: '默认策略',
        align: 'center',
        width: 120,
        slots: { default: 'defaultPolicy' },
      },
      {
        field: 'activeFlag',
        title: '状态',
        align: 'center',
        width: 100,
        slots: { default: 'activeFlag' },
      },
      {
        field: 'allowSubdomains',
        title: '允许子域名',
        align: 'center',
        width: 120,
        slots: { default: 'allowSubdomains' },
      },
      {
        field: 'whitelistDomains',
        title: '域名白名单',
        align: 'left',
        width: 300,
        showOverflow: true,
      },
      {
        field: 'blacklistDomains',
        title: '域名黑名单',
        align: 'left',
        width: 300,
        showOverflow: true,
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

  // ============= 数据操作方法 =============

  /**
   * 设置配置列表数据
   */
  const setConfigList = (list: DomainAccessConfig[]) => {
    configList.value = list
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
   * 添加配置到列表
   */
  const addConfigToList = (config: DomainAccessConfig) => {
    configList.value.unshift(config)
  }

  /**
   * 重置分页
   */
  const resetPagination = () => {
    pageInfo.value = undefined
  }

  /**
   * 更新列表中的配置
   */
  const updateConfigInList = (domainAccessConfigId: string, tenantId: string, updatedConfig: Partial<DomainAccessConfig>) => {
    const index = configList.value.findIndex(
      (item) => item.domainAccessConfigId === domainAccessConfigId && item.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(configList.value[index], updatedConfig)
    }
  }

  /**
   * 从列表中移除配置
   */
  const removeConfigFromList = (domainAccessConfigId: string, tenantId: string) => {
    const index = configList.value.findIndex(
      (item) => item.domainAccessConfigId === domainAccessConfigId && item.tenantId === tenantId
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
    setConfigList,
    updatePagination,
    resetPagination,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
  }
}

