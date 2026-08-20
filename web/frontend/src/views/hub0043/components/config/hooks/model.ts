/**
 * Hub0043 配置管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField, RsDataFormRenderContext } from '@/components/form/rs-data'
import type { RsSearchFormProps, RsSearchFormRenderContext } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import type { PageInfoObj } from '@/types/api'
import { RsRadio, RsRadioItem, type RsRadioValue } from '@/ui'
import { RsCodeEditor, type RsCodeEditorLanguage } from '@/ui/code-editor'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import { NamespaceNameSelector } from '../../../../hub0041/components'
import type { Config } from '../../../types/index'

/**
 * 配置表格配置（对齐 RsGrid Props 子集）。
 */
export interface ConfigGridConfig {
  columns: RsGridColumn<Config>[]
  selectable: boolean
  rowKey: string | ((row: Config) => string)
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 配置管理 Model
 */
export function useConfigModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0043'
  /** 加载状态 */
  const loading = ref(false)

  /** 配置列表数据 */
  const configList = ref<Config[]>([])

  /** 后端分页信息对象 */
  const pageInfo = ref<PageInfoObj | undefined>()

  // ============= 搜索表单配置 =============

  /** 搜索表单配置（符合 RsSearchFormProps 结构） */
  const searchFormConfig: Omit<RsSearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'namespaceId',
        label: '命名空间',
        type: 'custom',
        span: 6,
        required: true,
        render: (_formData: Record<string, any>, ctx: RsSearchFormRenderContext) => {
          return h(NamespaceNameSelector, {
            modelValue: (ctx.value as string) || '',
            'onUpdate:modelValue': (value: string) => ctx.onUpdate(value),
            onSelect: (namespace: { namespaceId?: string } | null) => {
              if (namespace?.namespaceId) {
                ctx.onUpdate(namespace.namespaceId)
              }
            },
          })
        },
      },
      {
        field: 'groupName',
        label: '分组名称',
        type: 'input',
        placeholder: '请输入分组名称，为空时查询全部',
        span: 6,
        clearable: true,
        defaultValue: '',
      },
      {
        field: 'configDataId',
        label: '配置ID',
        type: 'input',
        placeholder: '请输入配置ID（模糊查询）',
        span: 6,
        clearable: true,
      },
      {
        field: 'contentType',
        label: '内容类型',
        type: 'select',
        placeholder: '请选择内容类型',
        span: 6,
        clearable: true,
        options: [
          { label: '全部', value: '' },
          { label: '文本', value: 'text' },
          { label: 'JSON', value: 'json' },
          { label: 'XML', value: 'xml' },
          { label: 'YAML', value: 'yaml' },
          { label: 'Properties', value: 'properties' },
        ],
      },
    ],
    toolbarButtons: [
      {
        key: 'add',
        label: '新建配置',
        icon: 'AddOutline',
        type: 'primary',
        tooltip: '新建配置',
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

  // ============= 配置表单配置 =============
  const configFormConfig = {
    fields: [
      // ============= 基本信息 =============
      {
        field: 'namespaceId',
        label: '命名空间ID',
        type: 'custom',
        span: 12,
        required: true,
        primary: true,
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
          // 编辑/查看时锁定命名空间，新增时可从搜索条件带入后仍允许改选
          const locked = formData._mode === 'edit' || formData._mode === 'view'
          return h(NamespaceNameSelector, {
            modelValue: (ctx?.value as string) || formData.namespaceId || '',
            disabled: locked,
            'onUpdate:modelValue': (value: string) => ctx?.onUpdate(value),
            onSelect: (namespace: { namespaceId?: string } | null) => {
              if (namespace?.namespaceId) {
                ctx?.onUpdate(namespace.namespaceId)
              }
            },
          })
        },
      },
      {
        field: 'groupName',
        label: '分组名称',
        type: 'input',
        placeholder: 'DEFAULT_GROUP',
        span: 12,
        defaultValue: 'DEFAULT_GROUP',
        tips: '分组名称，默认为DEFAULT_GROUP',
      },
      {
        field: 'configDataId',
        label: '配置数据ID',
        type: 'input',
        placeholder: '请输入配置数据ID',
        span: 12,
        primary: true,
        required: true,
        tips: '配置的唯一标识（主键）',
      },
      {
        field: 'configDescription',
        label: '配置描述',
        type: 'textarea',
        placeholder: '请输入配置描述',
        span: 24,
        props: {
          rows: 3,
        },
      },
      {
        field: 'changeReason',
        label: '变更原因',
        type: 'textarea',
        placeholder: '请输入变更原因（可选）',
        span: 24,
        props: {
          rows: 2,
        },
      },
      // ============= 版本和MD5信息 =============
      {
        field: 'version',
        label: '版本号',
        type: 'number',
        span: 12,
        disabled: true,
      },
      {
        field: 'md5Value',
        label: 'MD5值',
        type: 'input',
        span: 12,
        disabled: true,
      },
      // ============= 内容类型（使用 radio，放在配置内容上面） =============
      {
        field: 'contentType',
        label: '内容类型',
        type: 'custom',
        span: 24,
        defaultValue: 'text',
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
          return h(
            RsRadio,
            {
              modelValue: (ctx?.value as string) || formData.contentType || 'text',
              'onUpdate:modelValue': (value?: RsRadioValue) => {
                if (typeof value === 'string') {
                  ctx?.onUpdate(value)
                }
              },
              size: 'sm',
            },
            {
              default: () => [
                h(RsRadioItem, { value: 'text' }, () => '文本'),
                h(RsRadioItem, { value: 'json' }, () => 'JSON'),
                h(RsRadioItem, { value: 'xml' }, () => 'XML'),
                h(RsRadioItem, { value: 'yaml' }, () => 'YAML'),
                h(RsRadioItem, { value: 'properties' }, () => 'Properties'),
              ],
            },
          )
        },
      },
      // ============= 配置内容（使用 RsCodeEditor） =============
      {
        field: 'configContent',
        label: '配置内容',
        type: 'custom',
        span: 24,
        required: true,
        render: (formData: Record<string, any>, ctx?: RsDataFormRenderContext) => {
          const contentTypeToLanguage: Record<string, RsCodeEditorLanguage> = {
            text: 'plaintext',
            json: 'json',
            xml: 'xml',
            yaml: 'yaml',
            properties: 'plaintext',
          }
          const language = contentTypeToLanguage[formData.contentType || 'text'] || 'plaintext'

          return h(RsCodeEditor, {
            modelValue: (ctx?.value as string) || formData.configContent || '',
            language,
            'onUpdate:modelValue': (value: string) => ctx?.onUpdate(value),
            height: '400px',
            showToolbar: false,
            placeholder: '请输入配置内容',
          })
        },
      },
    ] as RsDataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: ConfigGridConfig = {
    columns: [
      {
        key: 'configDataId',
        title: '配置ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'namespaceId',
        title: '命名空间',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'groupName',
        title: '分组名称',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 150,
      },
      {
        key: 'contentType',
        title: '内容类型',
        align: 'center',
        ellipsis: true,
        width: 120,
        formatter: (value) => {
          const typeMap: Record<string, string> = {
            text: '文本',
            json: 'JSON',
            xml: 'XML',
            yaml: 'YAML',
            properties: 'Properties',
          }
          const key = typeof value === 'string' ? value : ''
          return typeMap[key] || key
        },
      },
      {
        key: 'configDescription',
        title: '描述',
        align: 'left',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'version',
        title: '版本',
        align: 'center',
        width: 80,
      },
      {
        key: 'md5Value',
        title: 'MD5',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'addTime',
        title: '创建时间',
        align: 'center',
        ellipsis: true,
        width: 160,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '',
      },
      {
        key: 'addWho',
        title: '创建人',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
    ],
    selectable: true,
    rowKey: (row) => `${row.namespaceId}::${row.groupName}::${row.configDataId}`,
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
        { key: 'history', label: '历史版本', icon: 'clock' },
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
  const setConfigList = (list: Config[]) => {
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
  const addConfigToList = (config: Config) => {
    configList.value.unshift(config)
  }

  /**
   * 更新列表中的配置
   */
  const updateConfigInList = (
    namespaceId: string,
    groupName: string,
    configDataId: string,
    updatedConfig: Partial<Config>
  ) => {
    const index = configList.value.findIndex(
      (c) =>
        c.namespaceId === namespaceId &&
        c.groupName === groupName &&
        c.configDataId === configDataId
    )
    if (index !== -1) {
      Object.assign(configList.value[index], updatedConfig)
    }
  }

  /**
   * 从列表中删除配置
   */
  const removeConfigFromList = (
    namespaceId: string,
    groupName: string,
    configDataId: string
  ) => {
    const index = configList.value.findIndex(
      (c) =>
        c.namespaceId === namespaceId &&
        c.groupName === groupName &&
        c.configDataId === configDataId
    )
    if (index !== -1) {
      configList.value.splice(index, 1)
    }
  }

  /**
   * 批量删除配置
   */
  const removeConfigsFromList = (configs: Config[]) => {
    configs.forEach((config) => {
      removeConfigFromList(config.namespaceId, config.groupName, config.configDataId)
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
    configFormConfig,
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

/**
 * Model 返回类型
 */
export type ConfigModel = ReturnType<typeof useConfigModel>
