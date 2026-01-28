/**
 * Hub0043 配置历史管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { DataFormField } from '@/components/form/data/types'
import type { SearchFormProps } from '@/components/form/search/types'
import type { GridProps } from '@/components/grid'
import { formatDate } from '@/utils/format'
import { ArrowBackOutline } from '@vicons/ionicons5'
import { h, ref } from 'vue'
import { NamespaceNameSelector } from '../../../../hub0041/components'
import type { ConfigHistory } from '../../../types'

/**
 * 配置历史管理 Model
 */
export function useConfigHistoryModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0043-history'
  /** 加载状态 */
  const loading = ref(false)

  /** 配置历史列表数据 */
  const historyList = ref<ConfigHistory[]>([])

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
        placeholder: '请输入分组名称，默认DEFAULT_GROUP',
        span: 6,
        clearable: true,
        required: true,
      },
      {
        field: 'configDataId',
        label: '配置ID',
        type: 'input',
        placeholder: '请输入配置ID',
        span: 6,
        clearable: true,
        required: true,
      },
      {
        field: 'limit',
        label: '限制数量',
        type: 'number',
        placeholder: '限制数量，默认50',
        span: 6,
        defaultValue: 50,
      },
    ],
    toolbarButtons: [
      {
        key: 'back',
        label: '返回配置列表',
        icon: ArrowBackOutline,
        tooltip: '返回配置列表',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 GridProps 结构，排除响应式数据） */
  const gridConfig: Omit<GridProps, 'moduleId' | 'data' | 'loading'> = {
    columns: [
      {
        field: 'configHistoryId',
        title: '历史ID',
        sortable: true,
        align: 'center',
        width: 120,
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
        field: 'configDataId',
        title: '配置ID',
        sortable: true,
        align: 'center',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'changeType',
        title: '变更类型',
        align: 'center',
        width: 100,
        cellRender: {
          name: 'VxeRender',
          options: {
            render: ({ row }: { row: ConfigHistory }) => {
              const typeMap: Record<'CREATE' | 'UPDATE' | 'DELETE' | 'ROLLBACK', { label: string; type: 'success' | 'info' | 'error' | 'warning' }> = {
                CREATE: { label: '创建', type: 'success' },
                UPDATE: { label: '更新', type: 'info' },
                DELETE: { label: '删除', type: 'error' },
                ROLLBACK: { label: '回滚', type: 'warning' },
              }
              const changeType = row.changeType as keyof typeof typeMap
              const typeInfo = typeMap[changeType] || { label: changeType || '-', type: 'default' as const }
              return h('n-tag', { type: typeInfo.type, size: 'small' }, { default: () => typeInfo.label })
            },
          },
        },
      },
      {
        field: 'oldVersion',
        title: '旧版本',
        align: 'center',
        width: 80,
      },
      {
        field: 'newVersion',
        title: '新版本',
        align: 'center',
        width: 80,
        formatter: ({ row }: any) => {
          return row.newVersion || row.configVersion || '-'
        },
      },
      {
        field: 'oldMd5Value',
        title: '旧MD5',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
      {
        field: 'newMd5Value',
        title: '新MD5',
        align: 'center',
        showOverflow: true,
        width: 120,
      },
      {
        field: 'changeReason',
        title: '变更原因',
        align: 'left',
        showOverflow: true,
        width: 200,
      },
      {
        field: 'changedBy',
        title: '变更人',
        align: 'center',
        width: 120,
      },
      {
        field: 'changedAt',
        title: '变更时间',
        align: 'center',
        width: 180,
        formatter: ({ cellValue }) => {
          return cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : (cellValue || '-')
        },
      },
      {
        field: 'addTime',
        title: '创建时间',
        align: 'center',
        width: 180,
        formatter: ({ cellValue }) => {
          return cellValue ? formatDate(cellValue, 'YYYY-MM-DD HH:mm:ss') : (cellValue || '-')
        },
      },
      {
        field: 'addWho',
        title: '创建人',
        align: 'center',
        width: 120,
      },
    ],
    paginationConfig: {
      show: false, // 历史记录不使用分页
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
          code: 'rollback',
          name: '回滚',
          prefixIcon: 'vxe-icon-undo',
        },
      ],
    },
    height: '100%',
  }

  // ============= 详情表单配置 =============
  const detailFormConfig = {
    tabs: [
      { key: 'basic', label: '基本信息' },
      { key: 'content', label: '配置内容' },
    ],
    fields: [
      {
        field: 'configHistoryId',
        label: '历史ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'namespaceId',
        label: '命名空间ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'groupName',
        label: '分组名称',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'configDataId',
        label: '配置数据ID',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'changeType',
        label: '变更类型',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'oldVersion',
        label: '旧版本',
        type: 'number',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'newVersion',
        label: '新版本',
        type: 'number',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'oldMd5Value',
        label: '旧MD5值',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'newMd5Value',
        label: '新MD5值',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'changeReason',
        label: '变更原因',
        type: 'textarea',
        span: 24,
        tabKey: 'basic',
        disabled: true,
        props: {
          rows: 3,
        },
      },
      {
        field: 'changedBy',
        label: '变更人',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'changedAt',
        label: '变更时间',
        type: 'datetime',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'addTime',
        label: '创建时间',
        type: 'datetime',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'addWho',
        label: '创建人',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'editTime',
        label: '最后修改时间',
        type: 'datetime',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'editWho',
        label: '最后修改人',
        type: 'input',
        span: 12,
        tabKey: 'basic',
        disabled: true,
      },
      {
        field: 'oldContent',
        label: '旧配置内容',
        type: 'textarea',
        span: 24,
        tabKey: 'content',
        disabled: true,
        props: {
          rows: 15,
        },
      },
      {
        field: 'newContent',
        label: '新配置内容',
        type: 'textarea',
        span: 24,
        tabKey: 'content',
        disabled: true,
        props: {
          rows: 15,
        },
      },
    ] as DataFormField[],
  }

  // ============= 辅助方法 =============

  /**
   * 设置历史列表
   */
  const setHistoryList = (list: ConfigHistory[]) => {
    historyList.value = list
  }

  /**
   * 清空历史列表
   */
  const clearHistoryList = () => {
    historyList.value = []
  }

  return {
    // 基本信息
    moduleId,

    // 数据状态
    loading,
    historyList,

    // 配置
    searchFormConfig,
    gridConfig,
    detailFormConfig,

    // 方法
    setHistoryList,
    clearHistoryList,
  }
}

/**
 * Model 返回类型
 */
export type ConfigHistoryModel = ReturnType<typeof useConfigHistoryModel>

