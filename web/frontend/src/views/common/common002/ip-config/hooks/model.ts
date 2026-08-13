/**
 * IP访问控制配置列表 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { formatDate } from '@/utils/format'
import { validateCIDRList, validateIPList } from '@/utils/validate'
import { RsDynamicTags, RsTag } from '@/ui'
import { h, ref } from 'vue'
import type { IpAccessConfig } from './types'

/**
 * IP访问控制配置表格配置（对齐 RsGrid Props 子集）。
 */
export interface IpAccessConfigGridConfig {
  columns: RsGridColumn<IpAccessConfig>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * IP访问控制配置列表 Model
 * @param moduleId 模块ID（用于权限控制，由父级传入）
 */
export function useIpAccessConfigModel(moduleId: string) {
  // ============= 数据状态 =============
  /** 加载状态 */
  const loading = ref(false)

  /** IP配置列表数据 */
  const configList = ref<IpAccessConfig[]>([])

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
        tooltip: '新建IP访问控制配置',
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

  // 创建 IP 列表渲染函数（name 对齐 RsForm.rules，校验错误由 RsDynamicTags 展示）
  const createIpListRender = (field: 'whitelistIps' | 'blacklistIps') => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(RsDynamicTags, {
        name: field,
        modelValue: value,
        'onUpdate:modelValue': (newValue: string[]) => {
          formData[field] = newValue
        },
        placeholder: '输入IP地址，如：192.168.1.100',
      })
    }
  }

  // 创建 CIDR 列表渲染函数（name 对齐 RsForm.rules，校验错误由 RsDynamicTags 展示）
  const createCidrListRender = (field: 'whitelistCidrs' | 'blacklistCidrs') => {
    return (formData: Record<string, any>) => {
      const value = formData[field] || []

      return h(RsDynamicTags, {
        name: field,
        modelValue: value,
        'onUpdate:modelValue': (newValue: string[]) => {
          formData[field] = newValue
        },
        placeholder: '输入CIDR网段，如：192.168.1.0/24',
      })
    }
  }

  /** 表单字段配置 */
  const formFields: RsDataFormField[] = [
    // ============= 主键字段（隐藏，但必须存在用于更新） =============
    {
      field: 'ipAccessConfigId',
      label: 'IP访问配置ID',
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
          tips: 'allow（允许）: 默认允许，黑名单中的IP会被拒绝。deny（拒绝）: 默认拒绝，仅白名单允许。黑名单优先级高于白名单。',
          options: [
            { label: '允许（白名单模式）', value: 'allow' },
            { label: '拒绝（黑名单模式）', value: 'deny' },
          ],
        },
        {
          field: 'trustXForwardedFor',
          label: '信任X-Forwarded-For',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
        {
          field: 'trustXRealIp',
          label: '信任X-Real-IP',
          type: 'switch',
          span: 12,
          defaultValue: 'N',
          props: {
            checkedValue: 'Y',
            uncheckedValue: 'N',
          },
        },
      ],
    },
    // ============= IP白名单分组 =============
    {
      field: 'whitelistConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'whitelistIps',
          label: 'IP白名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '精确匹配单个IP地址，白名单中的IP将被允许访问',
          render: createIpListRender('whitelistIps'),
          rules: [
            {
              validator: (value: unknown) => {
                if (!Array.isArray(value) || value.length === 0) return true
                const result = validateIPList(value as string[])
                if (!result.valid) {
                  return `无效的IP地址：${result.invalidIps.join(', ')}`
                }
                return true
              },
              trigger: ['change', 'blur'],
            },
          ],
        },
        {
          field: 'whitelistCidrs',
          label: 'CIDR白名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '匹配网段范围，如：192.168.1.0/24（192.168.1.1-254）',
          render: createCidrListRender('whitelistCidrs'),
          rules: [
            {
              validator: (value: unknown) => {
                if (!Array.isArray(value) || value.length === 0) return true
                const result = validateCIDRList(value as string[])
                if (!result.valid) {
                  return `无效的CIDR网段：${result.invalidCidrs.join(', ')}`
                }
                return true
              },
              trigger: ['change', 'blur'],
            },
          ],
        },
      ],
    },
    // ============= IP黑名单分组 =============
    {
      field: 'blacklistConfig',
      label: '',
      type: 'fieldset',
      span: 24,
      children: [
        {
          field: 'blacklistIps',
          label: 'IP黑名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '精确匹配单个IP地址，黑名单优先级最高，直接拒绝访问',
          render: createIpListRender('blacklistIps'),
          rules: [
            {
              validator: (value: unknown) => {
                if (!Array.isArray(value) || value.length === 0) return true
                const result = validateIPList(value as string[])
                if (!result.valid) {
                  return `无效的IP地址：${result.invalidIps.join(', ')}`
                }
                return true
              },
              trigger: ['change', 'blur'],
            },
          ],
        },
        {
          field: 'blacklistCidrs',
          label: 'CIDR黑名单',
          type: 'custom',
          span: 12,
          defaultValue: [],
          tips: '匹配网段范围，黑名单优先级最高，直接拒绝访问',
          render: createCidrListRender('blacklistCidrs'),
          rules: [
            {
              validator: (value: unknown) => {
                if (!Array.isArray(value) || value.length === 0) return true
                const result = validateCIDRList(value as string[])
                if (!result.valid) {
                  return `无效的CIDR网段：${result.invalidCidrs.join(', ')}`
                }
                return true
              },
              trigger: ['change', 'blur'],
            },
          ],
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
  const gridConfig: IpAccessConfigGridConfig = {
    columns: [
      {
        key: 'ipAccessConfigId',
        title: 'IP访问配置ID',
        visible: false,
      },
      {
        key: 'tenantId',
        title: '租户ID',
        visible: false,
      },
      {
        key: 'securityConfigId',
        title: '安全配置ID',
        visible: false,
      },
      {
        key: 'configName',
        title: '配置名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
      },
      {
        key: 'defaultPolicy',
        title: '默认策略',
        align: 'center',
        width: 120,
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
        title: '状态',
        align: 'center',
        width: 100,
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
        key: 'whitelistIps',
        title: 'IP白名单',
        align: 'left',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'whitelistCidrs',
        title: 'CIDR白名单',
        align: 'left',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'blacklistIps',
        title: 'IP黑名单',
        align: 'left',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'blacklistCidrs',
        title: 'CIDR黑名单',
        align: 'left',
        width: 200,
        ellipsis: true,
      },
      {
        key: 'trustXForwardedFor',
        title: '信任X-Forwarded-For',
        align: 'center',
        width: 150,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.trustXForwardedFor === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.trustXForwardedFor === 'Y' ? '是' : '否'),
          ),
      },
      {
        key: 'trustXRealIp',
        title: '信任X-Real-IP',
        align: 'center',
        width: 130,
        render: (row) =>
          h(
            RsTag,
            {
              variant: row.trustXRealIp === 'Y' ? 'success' : 'default',
              size: 'sm',
            },
            () => (row.trustXRealIp === 'Y' ? '是' : '否'),
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
    rowKey: 'ipAccessConfigId',
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
   * 设置配置列表
   */
  const setConfigList = (list: IpAccessConfig[]) => {
    configList.value = list
  }

  /**
   * 清空配置列表
   */
  const clearConfigList = () => {
    configList.value = []
  }

  /**
   * 添加配置到列表
   */
  const addConfigToList = (config: IpAccessConfig) => {
    configList.value.unshift(config)
  }

  /**
   * 更新列表中的配置
   */
  const updateConfigInList = (ipAccessConfigId: string, tenantId: string, updatedConfig: Partial<IpAccessConfig>) => {
    const index = configList.value.findIndex(
      (c) => c.ipAccessConfigId === ipAccessConfigId && c.tenantId === tenantId
    )
    if (index !== -1) {
      Object.assign(configList.value[index], updatedConfig)
    }
  }

  /**
   * 从列表中删除配置
   */
  const removeConfigFromList = (ipAccessConfigId: string, tenantId: string) => {
    const index = configList.value.findIndex(
      (c) => c.ipAccessConfigId === ipAccessConfigId && c.tenantId === tenantId
    )
    if (index !== -1) {
      configList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除配置
   */
  const removeConfigsFromList = (configs: IpAccessConfig[]) => {
    configs.forEach((config) => {
      removeConfigFromList(config.ipAccessConfigId, config.tenantId)
    })
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    configList,
    pageInfo,

    // 配置
    searchFormConfig,
    formFields,
    gridConfig,

    // 方法
    resetPagination,
    updatePagination,
    setConfigList,
    clearConfigList,
    addConfigToList,
    updateConfigInList,
    removeConfigFromList,
    removeConfigsFromList,
  }
}

// 类型定义已移至 types.ts
