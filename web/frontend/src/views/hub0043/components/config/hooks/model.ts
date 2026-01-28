/**
 * Hub0043 配置管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import { GCodeMirror } from '@/components'
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
import { NRadio, NRadioGroup } from 'naive-ui'
import { h, ref } from 'vue'
import { NamespaceNameSelector } from '../../../../hub0041/components'
import type { Config } from '../../../types/index'

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

  /** 搜索表单配置（符合 SearchFormProps 结构） */
  const searchFormConfig: Omit<SearchFormProps, 'moduleId'> = {
    fields: [
      {
        field: 'namespaceId',
        label: '命名空间',
        type: 'custom',
        span: 6,
        required: true,
        render: (formData: Record<string, any>) => {
          return h(NamespaceNameSelector, {
            modelValue: formData.namespaceId || '',
            'onUpdate:modelValue': (value: string) => {
              formData.namespaceId = value
            },
            onSelect: (namespace: any) => {
              // 选择命名空间后，可以在这里处理额外逻辑
              if (namespace) {
                formData.namespaceId = namespace.namespaceId
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
        icon: AddOutline,
        type: 'primary',
        tooltip: '新建配置',
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
        disabled: true,
        render: (formData: Record<string, any>) => {
          // 编辑模式下禁用命名空间ID（通过检查是否有configDataId判断是否为编辑模式）
          const isEditMode = true
          return h(NamespaceNameSelector, {
            modelValue: formData.namespaceId || '',
            disabled: isEditMode,
            'onUpdate:modelValue': (value: string) => {
              formData.namespaceId = value
            },
            onSelect: (namespace: any) => {
              // 选择命名空间后，可以在这里处理额外逻辑
              if (namespace) {
                formData.namespaceId = namespace.namespaceId
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
      // ============= 内容类型（使用radio，放在配置内容上面） =============
      {
        field: 'contentType',
        label: '内容类型',
        type: 'custom',
        span: 24,
        defaultValue: 'text',
        render: (formData: Record<string, any>) => {
          return h(NRadioGroup, {
            value: formData.contentType || 'text',
            'onUpdate:value': (value: string) => {
              formData.contentType = value
            },
          }, {
            default: () => [
              h(NRadio, { label: '文本', value: 'text' }),
              h(NRadio, { label: 'JSON', value: 'json' }),
              h(NRadio, { label: 'XML', value: 'xml' }),
              h(NRadio, { label: 'YAML', value: 'yaml' }),
              h(NRadio, { label: 'Properties', value: 'properties' }),
            ]
          })
        },
      },
      // ============= 配置内容（使用GCodeMirror） =============
      {
        field: 'configContent',
        label: '配置内容',
        type: 'custom',
        span: 24,
        required: true,
        render: (formData: Record<string, any>) => {
          // 根据contentType动态设置language
          const contentTypeToLanguage: Record<string, string> = {
            'text': 'plaintext',
            'json': 'json',
            'xml': 'xml',
            'yaml': 'yaml',
            'properties': 'properties', // 使用 legacy-modes 的 properties 模式
          }
          const language = contentTypeToLanguage[formData.contentType || 'text'] || 'plaintext'
          
          return h(GCodeMirror, {
            modelValue: formData.configContent || '',
            language: language as any,
            'onUpdate:modelValue': (value: string) => {
              formData.configContent = value
            },
            height: '400px',
            placeholder: '请输入配置内容',
          })
        },
      },
    ] as DataFormField[],
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'configDataId',
        title: '配置ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'namespaceId',
        title: '命名空间',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 150,
      },
      {
        field: 'groupName',
        title: '分组名称',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 150,
      },
      {
        field: 'contentType',
        title: '内容类型',
        align: 'center',
        showOverflow: true,
        width: 120,
        formatter: ({ cellValue }) => {
          const typeMap: Record<string, string> = {
            'text': '文本',
            'json': 'JSON',
            'xml': 'XML',
            'yaml': 'YAML',
            'properties': 'Properties',
          }
          return typeMap[cellValue] || cellValue
        },
      },
      {
        field: 'configDescription',
        title: '描述',
        align: 'left',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'version',
        title: '版本',
        align: 'center',
        width: 80,
      },
      {
        field: 'md5Value',
        title: 'MD5',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        showOverflow: true,
        width: 160,
        formatter: ({ cellValue }) => {
          if (!cellValue) return ''
          return formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss')
        },
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        showOverflow: true,
        width: 120,
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
          code: 'history',
          name: '历史版本',
          prefixIcon: 'vxe-icon-time',
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

