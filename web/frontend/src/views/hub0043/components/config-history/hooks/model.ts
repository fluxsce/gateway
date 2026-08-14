/**
 * Hub0043 配置历史管理模块 Model
 * 统一管理搜索表单、表格配置和数据状态
 */

import type { RsDataFormField } from '@/components/form/rs-data'
import type { RsSearchFormProps, RsSearchFormRenderContext } from '@/components/form/rs-search'
import type { RsGridColumn, RsGridMenuConfig, RsGridPaginationConfig } from '@/components/rs-grid'
import { RsTag, type RsTagVariant } from '@/ui'
import { formatDate } from '@/utils/format'
import { h, ref } from 'vue'
import { NamespaceNameSelector } from '../../../../hub0041/components'
import type { ConfigHistory } from '../../../types'

/**
 * 配置历史表格配置（对齐 RsGrid Props 子集）。
 */
export interface ConfigHistoryGridConfig {
  columns: RsGridColumn<ConfigHistory>[]
  selectable: boolean
  rowKey: string
  height: string
  paginationConfig: RsGridPaginationConfig
  menuConfig: RsGridMenuConfig
}

/**
 * 配置历史管理 Model
 */
export function useConfigHistoryModel() {
  // ============= 数据状态 =============
  const moduleId = 'hub0043:history'
  /** 加载状态 */
  const loading = ref(false)

  /** 配置历史列表数据 */
  const historyList = ref<ConfigHistory[]>([])

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
        icon: 'ArrowBackOutline',
        tooltip: '返回配置列表',
      },
    ],
    showSearchButton: true,
    showResetButton: true,
  }

  // ============= 表格配置 =============

  /** 表格配置（符合 RsGrid Props 结构，排除响应式数据） */
  const gridConfig: ConfigHistoryGridConfig = {
    columns: [
      {
        key: 'configHistoryId',
        title: '历史ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 120,
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
        key: 'configDataId',
        title: '配置ID',
        sortable: true,
        align: 'center',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'changeType',
        title: '变更类型',
        align: 'center',
        width: 100,
        render: (row) => {
          const typeMap: Record<
            ConfigHistory['changeType'],
            { label: string; variant: RsTagVariant }
          > = {
            CREATE: { label: '创建', variant: 'success' },
            UPDATE: { label: '更新', variant: 'info' },
            DELETE: { label: '删除', variant: 'danger' },
            ROLLBACK: { label: '回滚', variant: 'warning' },
          }
          const typeInfo = typeMap[row.changeType] || {
            label: row.changeType || '-',
            variant: 'default' as const,
          }
          return h(RsTag, { variant: typeInfo.variant, size: 'sm' }, () => typeInfo.label)
        },
      },
      {
        key: 'oldVersion',
        title: '旧版本',
        align: 'center',
        width: 80,
      },
      {
        key: 'newVersion',
        title: '新版本',
        align: 'center',
        width: 80,
        formatter: (_value, row) => String(row.newVersion || row.configVersion || '-'),
      },
      {
        key: 'oldMd5Value',
        title: '旧MD5',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'newMd5Value',
        title: '新MD5',
        align: 'center',
        ellipsis: true,
        width: 120,
      },
      {
        key: 'changeReason',
        title: '变更原因',
        align: 'left',
        ellipsis: true,
        width: 200,
      },
      {
        key: 'changedBy',
        title: '变更人',
        align: 'center',
        width: 120,
      },
      {
        key: 'changedAt',
        title: '变更时间',
        align: 'center',
        width: 180,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        key: 'addTime',
        title: '创建时间',
        align: 'center',
        width: 180,
        formatter: (value) =>
          value ? formatDate(value as string, 'YYYY-MM-DD HH:mm:ss') : '-',
      },
      {
        key: 'addWho',
        title: '创建人',
        align: 'center',
        width: 120,
      },
    ],
    selectable: false,
    rowKey: 'configHistoryId',
    paginationConfig: {
      show: false,
    },
    menuConfig: {
      enabled: true,
      items: [
        { key: 'view', label: '查看详情', icon: 'eye' },
        { key: 'rollback', label: '回滚', icon: 'undo-2' },
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
    ] as RsDataFormField[],
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
